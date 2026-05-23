package controller

import (
	"context"
	"time"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	autoscalingv1 "github.com/gtmong0077/predictive-scale-operator/api/v1" // 모듈 이름에 맞게 수정하세요
)

// 4단계 상태 정의
const (
	PhaseIdle       = "IDLE"         // 시작 전 대기
	PhasePreScaling = "PRE_SCALING"  // 오픈 전 파드 미리 확장
	PhaseActive     = "EVENT_ACTIVE" // 티켓팅 진행 중 (HPA 위임 구간)
	PhaseCoolDown   = "COOL_DOWN"    // 이벤트 종료 및 자원 회수 시점
	PhaseCompleted  = "COMPLETED"    // 완전히 종료됨
)

type EventScalePolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=autoscaling.yourdomain.com,resources=eventscalepolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling.yourdomain.com,resources=eventscalepolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;update;patch

func (r *EventScalePolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// 1. 정책 CRD 조회
	var policy autoscalingv1.EventScalePolicy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 이미 처리가 완료된 이벤트면 더 이상 Reconcile하지 않음
	if policy.Status.Phase == PhaseCompleted {
		return ctrl.Result{}, nil
	}

	// 2. 시간(타이밍) 계산
	now := time.Now()
	openTime := policy.Spec.OpenTime.Time
	preScaleTime := openTime.Add(-time.Duration(policy.Spec.PreScaleSeconds) * time.Second)
	restoreTime := openTime.Add(time.Duration(policy.Spec.RestoreAfterSeconds) * time.Second)

	// 3. 현재 시간에 따른 '목표 상태(Target Phase)' 판단
	targetPhase := PhaseIdle
	if now.After(restoreTime) || now.Equal(restoreTime) {
		targetPhase = PhaseCoolDown
	} else if now.After(openTime) || now.Equal(openTime) {
		targetPhase = PhaseActive
	} else if now.After(preScaleTime) || now.Equal(preScaleTime) {
		targetPhase = PhasePreScaling
	}

	// 4. 상태에 따른 스위치문 분기 (4단계 액션)
	switch targetPhase {
	case PhaseIdle:
		// [1단계] 대기: 사전 확장 시간까지 얌전히 기다립니다.
		return r.updatePhaseAndRequeue(ctx, &policy, PhaseIdle, preScaleTime.Sub(now))

	case PhasePreScaling:
		// [2단계] 사전 확장: HPA의 minReplicas를 TargetReplicas로 올립니다.
		if err := r.updateHPAMinReplicas(ctx, &policy, policy.Spec.TargetReplicas); err != nil {
			logger.Error(err, "Failed to scale up HPA during PRE_SCALING")
			return ctrl.Result{}, err
		}
		// 스케일업을 완료했으면 정각(openTime)까지 대기합니다.
		return r.updatePhaseAndRequeue(ctx, &policy, PhasePreScaling, openTime.Sub(now))

	case PhaseActive:
		// [3단계] 이벤트 진행 중: 타겟 HPA를 건드리지 않고 K8s Native HPA에게 위임합니다.
		// (만약 여기에 모니터링 알람이나 로깅을 쏘고 싶다면 이 블록에 추가하면 됩니다)
		logger.Info("Event is currently ACTIVE. Delegating autoscaling to HPA.")
		return r.updatePhaseAndRequeue(ctx, &policy, PhaseActive, restoreTime.Sub(now))

	case PhaseCoolDown:
		// [4단계] 복구 및 종료: 이벤트가 끝났으므로 HPA를 원래 상태로 롤백합니다.
		if err := r.updateHPAMinReplicas(ctx, &policy, policy.Spec.MinReplicasAfterEvent); err != nil {
			logger.Error(err, "Failed to restore HPA during COOL_DOWN")
			return ctrl.Result{}, err
		}

		// 롤백이 성공적으로 끝났다면 상태를 COMPLETED로 변경하고 Requeue를 중단합니다.
		policy.Status.Phase = PhaseCompleted
		if err := r.Status().Update(ctx, &policy); err != nil {
			return ctrl.Result{}, err
		}
		logger.Info("EventScalePolicy successfully restored and COMPLETED")
		return ctrl.Result{}, nil
	}

	return ctrl.Result{}, nil
}

// 상태(Phase)를 업데이트하고 다음 이벤트 시점까지 Requeue(예약)하는 헬퍼 함수
func (r *EventScalePolicyReconciler) updatePhaseAndRequeue(ctx context.Context, policy *autoscalingv1.EventScalePolicy, phase string, waitTime time.Duration) (ctrl.Result, error) {
	if policy.Status.Phase != phase {
		policy.Status.Phase = phase
		if err := r.Status().Update(ctx, policy); err != nil {
			return ctrl.Result{}, err
		}
		log.FromContext(ctx).Info("Phase transitioned", "NewPhase", phase)
	}

	// 방어 코드: 시간이 음수면 즉시 실행(0)
	if waitTime < 0 {
		waitTime = 0
	}
	return ctrl.Result{RequeueAfter: waitTime}, nil
}

// 타겟 HPA 리소스를 찾아 minReplicas를 안전하게 패치(Patch)하는 헬퍼 함수
func (r *EventScalePolicyReconciler) updateHPAMinReplicas(ctx context.Context, policy *autoscalingv1.EventScalePolicy, targetMinReplicas int32) error {
	hpaName := types.NamespacedName{
		Name:      policy.Spec.TargetRef.Name,
		Namespace: policy.Namespace,
	}

	var hpa autoscalingv2.HorizontalPodAutoscaler
	if err := r.Get(ctx, hpaName, &hpa); err != nil {
		if errors.IsNotFound(err) {
			log.FromContext(ctx).Info("HPA not found, skipping patch", "HPA", hpaName.Name)
			return nil
		}
		return err
	}

	// 멱등성 보장: 이미 값이 목표치와 같다면 API 서버에 패치 요청을 보내지 않음
	if hpa.Spec.MinReplicas != nil && *hpa.Spec.MinReplicas == targetMinReplicas {
		return nil
	}

	// 기존 객체 복사 후 Patch 적용 (Conflict 에러 방지)
	originalHPA := hpa.DeepCopy()
	hpa.Spec.MinReplicas = &targetMinReplicas

	log.FromContext(ctx).Info("Patching HPA minReplicas", "HPA", hpa.Name, "NewMinReplicas", targetMinReplicas)
	return r.Patch(ctx, &hpa, client.MergeFrom(originalHPA))
}

func (r *EventScalePolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&autoscalingv1.EventScalePolicy{}).
		Complete(r)
}
