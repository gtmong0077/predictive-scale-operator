package controller

import (
	"context"
	"fmt"

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	autoscalingv1 "github.com/gtmong0077/predictive-scale-operator/api/v1"
)

func hpaNameForDeployment(deploymentName string) string {
	return fmt.Sprintf("%s-hpa", deploymentName)
}

func FindHPA(ctx context.Context, c client.Client, namespace, targetDeployment string) (*autoscalingv2.HorizontalPodAutoscaler, error) {
	var hpa autoscalingv2.HorizontalPodAutoscaler
	if err := c.Get(ctx, types.NamespacedName{
		Name:      hpaNameForDeployment(targetDeployment),
		Namespace: namespace,
	}, &hpa); err != nil {
		return nil, err
	}
	return &hpa, nil
}

func PatchHPAMinMax(
	ctx context.Context,
	c client.Client,
	scheme *runtime.Scheme,
	policy *autoscalingv1.EventScalePolicy,
	minReplicas, maxReplicas int32,
) error {
	logger := log.FromContext(ctx)

	hpa, err := FindHPA(ctx, c, policy.Namespace, policy.Spec.TargetDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("HPA not found, skipping patch", "HPA", hpaNameForDeployment(policy.Spec.TargetDeployment))
			return nil
		}
		return err
	}

	minChanged := hpa.Spec.MinReplicas == nil || *hpa.Spec.MinReplicas != minReplicas
	maxChanged := hpa.Spec.MaxReplicas != maxReplicas
	ownerMissing := len(hpa.OwnerReferences) == 0
	if !minChanged && !maxChanged && !ownerMissing {
		return nil
	}

	original := hpa.DeepCopy()
	hpa.Spec.MinReplicas = &minReplicas
	hpa.Spec.MaxReplicas = maxReplicas

	if ownerMissing {
		if err := ctrl.SetControllerReference(policy, hpa, scheme); err != nil {
			return err
		}
	}

	logger.Info("Patching HPA replicas",
		"HPA", hpa.Name,
		"minReplicas", minReplicas,
		"maxReplicas", maxReplicas,
	)
	return c.Patch(ctx, hpa, client.MergeFrom(original))
}
