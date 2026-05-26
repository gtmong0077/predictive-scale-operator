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
	restoreTime := openTime.Add(time.Duration(policy.Spec.RestoreAfterSeconds) * time.Second)

	// 계단식 확장 시점 (오픈 시간 기준 -130초, -90초, -50초)
	step1Time := openTime.Add(-130 * time.Second)
	step2Time := openTime.Add(-90 * time.Second)
	step3Time := openTime.Add(-50 * time.Second)

	var targetPhase string
	if now.After(restoreTime) {
		targetPhase = PhaseCoolDown
	} else if now.After(openTime) || now.Equal(openTime) {
		targetPhase = PhaseActive
	} else if now.After(step1Time) || now.Equal(step1Time) {
		targetPhase = PhasePreScaling
	} else {
		targetPhase = PhaseIdle
	}

	// 3. 상태별 행동(Action) 및 다음 알람(Requeue) 설정
	switch targetPhase {
	case PhaseIdle:
		// 130초 전(Step 1)까지 꿀잠 자기
		logger.Info("이벤트 대기 중...", "Step1 시작시간", step1Time)
		return r.updatePhaseAndRequeue(ctx, &policy, PhaseIdle, step1Time.Sub(now))

	case PhasePreScaling:
		var currentStepTarget int32
		var nextWakeUpTime time.Time
		targetTotal := policy.Spec.TargetReplicas

		// 현재 시간에 따라 몇 %를 띄울지 결정 (Level-triggered 방식의 장점!)
		if now.After(step3Time) || now.Equal(step3Time) {
			// 50초 전 ~ 오픈 전: 100% 투입
			currentStepTarget = targetTotal
			nextWakeUpTime = openTime // 다음 기상은 티켓팅 오픈 시간!
			logger.Info("사전 확장 3단계 투입 [100%]", "목표파드수", currentStepTarget)
		} else if now.After(step2Time) || now.Equal(step2Time) {
			// 90초 전 ~ 50초 전: 60% 투입
			currentStepTarget = (targetTotal * 60) / 100
			if currentStepTarget == 0 {
				currentStepTarget = 1
			} // 최소 1개 보장
			nextWakeUpTime = step3Time // 다음 기상은 50초 전(Step 3)
			logger.Info("사전 확장 2단계 투입 [60%]", "목표파드수", currentStepTarget)
		} else {
			// 130초 전 ~ 90초 전: 30% 투입
			currentStepTarget = (targetTotal * 30) / 100
			if currentStepTarget == 0 {
				currentStepTarget = 1
			} // 최소 1개 보장
			nextWakeUpTime = step2Time // 다음 기상은 90초 전(Step 2)
			logger.Info("사전 확장 1단계 투입 [30%]", "목표파드수", currentStepTarget)
		}

		// 계산된 타겟(30%, 60%, 100%)으로 HPA 업데이트
		if err := r.updateHPAMinReplicas(ctx, &policy, currentStepTarget); err != nil {
			logger.Error(err, "HPA minReplicas 업데이트 실패")
			return ctrl.Result{}, err
		}

		// 목표 스텝 달성 완료, 다음 스텝 시간까지 다시 대기 모드
		return r.updatePhaseAndRequeue(ctx, &policy, PhasePreScaling, nextWakeUpTime.Sub(now))

	case PhaseActive:
		// (기존과 동일) 100% 유지 확인 및 복구 시간까지 대기
		logger.Info("이벤트 진행 중! 100% 상태 유지")
		if err := r.updateHPAMinReplicas(ctx, &policy, policy.Spec.TargetReplicas); err != nil {
			return ctrl.Result{}, err
		}
		return r.updatePhaseAndRequeue(ctx, &policy, PhaseActive, restoreTime.Sub(now))

	case PhaseCoolDown:
		// (기존과 동일) 이벤트 종료, 파드 축소
		logger.Info("이벤트 종료. 기존 스케일로 복구")
		if err := r.updateHPAMinReplicas(ctx, &policy, policy.Spec.MinReplicasAfterEvent); err != nil {
			return ctrl.Result{}, err
		}
		// 작업 완료! 더 이상 깨워달라고 하지 않음 (Requeue 안 함)
		return r.updatePhase(ctx, &policy, PhaseCoolDown)
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

// 에러 2 해결: 작업이 완전히 끝났을 때 상태만 변경하고 Requeue는 하지 않는 헬퍼 함수
func (r *EventScalePolicyReconciler) updatePhase(ctx context.Context, policy *autoscalingv1.EventScalePolicy, phase string) (ctrl.Result, error) {
	if policy.Status.Phase != phase {
		policy.Status.Phase = phase
		if err := r.Status().Update(ctx, policy); err != nil {
			return ctrl.Result{}, err
		}
		log.FromContext(ctx).Info("Phase transitioned to final state", "Phase", phase)
	}
	return ctrl.Result{}, nil
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
