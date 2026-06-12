package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func CountReadyPods(ctx context.Context, c client.Client, namespace, deploymentName string) (ready int32, total int32, err error) {
	var deployment appsv1.Deployment
	if err := c.Get(ctx, types.NamespacedName{
		Name:      deploymentName,
		Namespace: namespace,
	}, &deployment); err != nil {
		if errors.IsNotFound(err) {
			return 0, 0, nil
		}
		return 0, 0, err
	}

	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		return 0, 0, err
	}

	var podList corev1.PodList
	if err := c.List(ctx, &podList, &client.ListOptions{
		Namespace:     namespace,
		LabelSelector: selector,
	}); err != nil {
		return 0, 0, err
	}

	for i := range podList.Items {
		pod := &podList.Items[i]
		total++
		if isPodReady(pod) {
			ready++
		}
	}

	return ready, total, nil
}

func isPodReady(pod *corev1.Pod) bool {
	if pod.DeletionTimestamp != nil {
		return false
	}
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}
