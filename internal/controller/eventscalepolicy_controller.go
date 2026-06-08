package controller

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	autoscalingv1 "github.com/gtmong0077/predictive-scale-operator/api/v1"
)

const (
	PhaseIdle           = "IDLE"
	PhasePreScaling     = "PRE_SCALING"
	PhaseReadyBuffer    = "READY_BUFFER"
	PhaseEventActive    = "EVENT_ACTIVE"
	PhaseScaleDownGuard = "SCALE_DOWN_GUARD"
	PhaseCompleted      = "COMPLETED"

	reconcileInterval = 30 * time.Second
	readyThreshold    = 0.95
)

type EventScalePolicyReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=autoscaling.eventscale.com,resources=eventscalepolicies,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling.eventscale.com,resources=eventscalepolicies/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=autoscaling.eventscale.com,resources=eventscalepolicies/finalizers,verbs=update
// +kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;update;patch
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch
// +kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch

func (r *EventScalePolicyReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var policy autoscalingv1.EventScalePolicy
	if err := r.Get(ctx, req.NamespacedName, &policy); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if policy.Status.Phase == PhaseCompleted {
		return ctrl.Result{}, nil
	}

	now := time.Now()
	openTime := policy.Spec.OpenTime.Time
	preScaleSeconds := ComputePreScaleSeconds(policy.Spec.StartupProfile, policy.Spec.WarmupSeconds)
	preScaleStart := openTime.Add(-time.Duration(preScaleSeconds) * time.Second)
	guardStart := openTime.Add(time.Duration(policy.Spec.MaxOverprovisionMinutes) * time.Minute)
	completedStart := guardStart.Add(time.Duration(policy.Spec.ScaleDownGuardMinutes) * time.Minute)

	finalReplicas := ComputeTargetReplicas(
		policy.Spec.ExpectedPeakRPS,
		policy.Spec.PodStableRPS,
		policy.Spec.TargetUtilization,
		policy.Spec.MinReplicas,
		policy.Spec.MaxReplicas,
	)

	readyReplicas, _, err := CountReadyPods(ctx, r.Client, policy.Namespace, policy.Spec.TargetDeployment)
	if err != nil {
		logger.Error(err, "Could not count ready Pods", "deployment", policy.Spec.TargetDeployment)
		return ctrl.Result{RequeueAfter: reconcileInterval}, err
	}

	targetPhase, desiredMin, currentStep, requeueAfter := r.resolvePhase(
		now, openTime, preScaleStart, guardStart, completedStart,
		finalReplicas, policy.Spec.MinReplicas, policy.Spec.StepPolicy,
		readyReplicas,
	)

	if err := r.syncStatus(ctx, &policy, targetPhase, finalReplicas, readyReplicas, desiredMin, currentStep, preScaleStart); err != nil {
		return ctrl.Result{}, err
	}

	if err := PatchHPAMinMax(ctx, r.Client, r.Scheme, &policy, desiredMin, policy.Spec.MaxReplicas); err != nil {
		logger.Error(err, "Could not patch HPA", "deployment", policy.Spec.TargetDeployment)
		return ctrl.Result{RequeueAfter: reconcileInterval}, err
	}

	if targetPhase == PhaseCompleted {
		return ctrl.Result{}, nil
	}

	if requeueAfter <= 0 {
		requeueAfter = reconcileInterval
	}
	return ctrl.Result{RequeueAfter: requeueAfter}, nil
}

func (r *EventScalePolicyReconciler) resolvePhase(
	now, openTime, preScaleStart, guardStart, completedStart time.Time,
	finalReplicas, minReplicas int32,
	steps []autoscalingv1.StepPolicy,
	readyReplicas int32,
) (phase string, desiredMin int32, currentStep int32, requeueAfter time.Duration) {
	switch {
	case now.Before(preScaleStart):
		return PhaseIdle, minReplicas, -1, preScaleStart.Sub(now)

	case now.Before(openTime):
		secondsUntilOpen := int64(openTime.Sub(now).Seconds())
		stepIndex, desired := resolvePreScaleStep(finalReplicas, minReplicas, steps, readyReplicas, secondsUntilOpen)

		readyTarget := int32(float64(finalReplicas) * readyThreshold)
		if readyReplicas >= readyTarget && secondsUntilOpen <= 30 {
			return PhaseReadyBuffer, finalReplicas, stepIndex, openTime.Sub(now)
		}

		return PhasePreScaling, desired, stepIndex, minDuration(openTime.Sub(now), reconcileInterval)

	case now.Before(guardStart):
		return PhaseEventActive, finalReplicas, int32(len(steps) - 1), guardStart.Sub(now)

	case now.Before(completedStart):
		return PhaseScaleDownGuard, minReplicas, -1, completedStart.Sub(now)

	default:
		return PhaseCompleted, minReplicas, -1, 0
	}
}

func (r *EventScalePolicyReconciler) syncStatus(
	ctx context.Context,
	policy *autoscalingv1.EventScalePolicy,
	phase string,
	finalReplicas, readyReplicas, desiredMin, currentStep int32,
	preScaleStart time.Time,
) error {
	now := metav1Now()
	changed := false

	if policy.Status.Phase != phase {
		policy.Status.Phase = phase
		policy.Status.LastTransitionTime = now
		changed = true
	}
	if policy.Status.ComputedTargetReplicas != finalReplicas {
		policy.Status.ComputedTargetReplicas = finalReplicas
		changed = true
	}
	if policy.Status.ReadyReplicas != readyReplicas {
		policy.Status.ReadyReplicas = readyReplicas
		changed = true
	}
	if policy.Status.DesiredMinReplicas != desiredMin {
		policy.Status.DesiredMinReplicas = desiredMin
		changed = true
	}
	if policy.Status.CurrentStep != currentStep {
		policy.Status.CurrentStep = currentStep
		changed = true
	}

	preScaleStartMeta := metav1Time(preScaleStart)
	if policy.Status.PreScaleStartTime == nil || !policy.Status.PreScaleStartTime.Equal(&preScaleStartMeta) {
		policy.Status.PreScaleStartTime = &preScaleStartMeta
		changed = true
	}

	if !changed {
		return nil
	}

	if err := r.Status().Update(ctx, policy); err != nil {
		return err
	}

	log.FromContext(ctx).Info("Updated EventScalePolicy status",
		"phase", phase,
		"targetReplicas", finalReplicas,
		"readyReplicas", readyReplicas,
		"desiredMinReplicas", desiredMin,
	)
	return nil
}

func (r *EventScalePolicyReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&autoscalingv1.EventScalePolicy{}).
		Owns(&autoscalingv2.HorizontalPodAutoscaler{}).
		Watches(
			&appsv1.Deployment{},
			handler.EnqueueRequestsFromMapFunc(r.findPoliciesForDeployment),
		).
		Watches(
			&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(r.findPoliciesForPod),
		).
		Watches(
			&autoscalingv2.HorizontalPodAutoscaler{},
			handler.EnqueueRequestsFromMapFunc(r.findPoliciesForHPA),
		).
		Complete(r)
}

func (r *EventScalePolicyReconciler) findPoliciesForPod(ctx context.Context, obj client.Object) []reconcile.Request {
	pod, ok := obj.(*corev1.Pod)
	if !ok {
		return nil
	}

	deploymentName := deploymentNameFromPod(ctx, r.Client, pod)
	if deploymentName == "" {
		return nil
	}
	return r.listPolicyRequests(ctx, pod.Namespace, deploymentName)
}

func deploymentNameFromPod(ctx context.Context, c client.Client, pod *corev1.Pod) string {
	for _, ref := range pod.OwnerReferences {
		if ref.Kind != "ReplicaSet" {
			continue
		}

		var rs appsv1.ReplicaSet
		if err := c.Get(ctx, types.NamespacedName{
			Name:      ref.Name,
			Namespace: pod.Namespace,
		}, &rs); err != nil {
			continue
		}

		for _, rsRef := range rs.OwnerReferences {
			if rsRef.Kind == "Deployment" {
				return rsRef.Name
			}
		}
	}
	return ""
}

func (r *EventScalePolicyReconciler) findPoliciesForDeployment(ctx context.Context, obj client.Object) []reconcile.Request {
	deployment, ok := obj.(*appsv1.Deployment)
	if !ok {
		return nil
	}
	return r.listPolicyRequests(ctx, deployment.Namespace, deployment.Name)
}

func (r *EventScalePolicyReconciler) findPoliciesForHPA(ctx context.Context, obj client.Object) []reconcile.Request {
	hpa, ok := obj.(*autoscalingv2.HorizontalPodAutoscaler)
	if !ok {
		return nil
	}
	if hpa.Spec.ScaleTargetRef.Kind != "Deployment" {
		return nil
	}
	return r.listPolicyRequests(ctx, hpa.Namespace, hpa.Spec.ScaleTargetRef.Name)
}

func (r *EventScalePolicyReconciler) listPolicyRequests(ctx context.Context, namespace, deploymentName string) []reconcile.Request {
	var policies autoscalingv1.EventScalePolicyList
	if err := r.List(ctx, &policies, client.InNamespace(namespace)); err != nil {
		return nil
	}

	requests := make([]reconcile.Request, 0)
	for i := range policies.Items {
		if policies.Items[i].Spec.TargetDeployment == deploymentName {
			requests = append(requests, reconcile.Request{
				NamespacedName: client.ObjectKeyFromObject(&policies.Items[i]),
			})
		}
	}
	return requests
}

func minDuration(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}

func metav1Now() *metav1.Time {
	now := metav1.Now()
	return &now
}

func metav1Time(t time.Time) metav1.Time {
	return metav1.NewTime(t)
}
