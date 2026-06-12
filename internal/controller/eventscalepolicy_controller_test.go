/*
Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	autoscalingv1 "github.com/gtmong0077/predictive-scale-operator/api/v1"
)

func samplePolicySpec(openTime time.Time) autoscalingv1.EventScalePolicySpec {
	return autoscalingv1.EventScalePolicySpec{
		TargetDeployment:        "ticket-api",
		OpenTime:                metav1.NewTime(openTime),
		ExpectedPeakRPS:         5000,
		PodStableRPS:            120,
		TargetUtilization:       0.7,
		MinReplicas:             3,
		MaxReplicas:             80,
		WarmupSeconds:           30,
		ScaleDownGuardMinutes:   10,
		MaxOverprovisionMinutes: 15,
		StartupProfile: autoscalingv1.StartupProfile{
			DetectDecideSeconds: 20,
			PodReadySeconds:     90,
			SafetyMarginSeconds: 30,
		},
		StepPolicy: []autoscalingv1.StepPolicy{
			{OffsetSeconds: 180, Ratio: 0.4},
			{OffsetSeconds: 90, Ratio: 0.75},
			{OffsetSeconds: 30, Ratio: 1.0},
		},
	}
}

var _ = Describe("EventScalePolicy Controller", func() {
	Context("When reconciling a resource", func() {
		const resourceName = "test-resource"

		ctx := context.Background()
		typeNamespacedName := types.NamespacedName{
			Name:      resourceName,
			Namespace: "default",
		}

		BeforeEach(func() {
			minReplicas := int32(3)
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ticket-api",
					Namespace: "default",
					Labels: map[string]string{
						"app": "ticket-api",
					},
				},
				Spec: appsv1.DeploymentSpec{
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{"app": "ticket-api"},
					},
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{"app": "ticket-api"},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "ticket-api",
									Image: "nginx:alpine",
									Ports: []corev1.ContainerPort{{ContainerPort: 8080}},
									ReadinessProbe: &corev1.Probe{
										ProbeHandler: corev1.ProbeHandler{
											HTTPGet: &corev1.HTTPGetAction{
												Path: "/",
												Port: intstr.FromInt32(8080),
											},
										},
									},
								},
							},
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, deployment)).To(Succeed())

			hpa := &autoscalingv2.HorizontalPodAutoscaler{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "ticket-api-hpa",
					Namespace: "default",
				},
				Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
					ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
						APIVersion: "apps/v1",
						Kind:       "Deployment",
						Name:       "ticket-api",
					},
					MinReplicas: &minReplicas,
					MaxReplicas: 80,
				},
			}
			Expect(k8sClient.Create(ctx, hpa)).To(Succeed())

			policy := &autoscalingv1.EventScalePolicy{}
			err := k8sClient.Get(ctx, typeNamespacedName, policy)
			if err != nil && errors.IsNotFound(err) {
				resource := &autoscalingv1.EventScalePolicy{
					ObjectMeta: metav1.ObjectMeta{
						Name:      resourceName,
						Namespace: "default",
					},
					Spec: samplePolicySpec(time.Now().Add(10 * time.Minute)),
				}
				Expect(k8sClient.Create(ctx, resource)).To(Succeed())
			}
		})

		AfterEach(func() {
			_ = k8sClient.Delete(ctx, &autoscalingv1.EventScalePolicy{
				ObjectMeta: metav1.ObjectMeta{Name: resourceName, Namespace: "default"},
			})
			_ = k8sClient.Delete(ctx, &autoscalingv2.HorizontalPodAutoscaler{
				ObjectMeta: metav1.ObjectMeta{Name: "ticket-api-hpa", Namespace: "default"},
			})
			_ = k8sClient.Delete(ctx, &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{Name: "ticket-api", Namespace: "default"},
			})
		})

		It("should reconcile into IDLE and patch HPA bounds", func() {
			controllerReconciler := &EventScalePolicyReconciler{
				Client: k8sClient,
				Scheme: k8sClient.Scheme(),
			}

			_, err := controllerReconciler.Reconcile(ctx, reconcile.Request{
				NamespacedName: typeNamespacedName,
			})
			Expect(err).NotTo(HaveOccurred())

			updated := &autoscalingv1.EventScalePolicy{}
			Expect(k8sClient.Get(ctx, typeNamespacedName, updated)).To(Succeed())
			Expect(updated.Status.Phase).To(Equal(PhaseIdle))
			Expect(updated.Status.ComputedTargetReplicas).To(Equal(int32(60)))
			Expect(updated.Status.DesiredMinReplicas).To(Equal(int32(3)))
			Expect(updated.Status.PreScaleStartTime).NotTo(BeNil())

			hpa := &autoscalingv2.HorizontalPodAutoscaler{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "ticket-api-hpa", Namespace: "default"}, hpa)).To(Succeed())
			Expect(hpa.Spec.MinReplicas).NotTo(BeNil())
			Expect(*hpa.Spec.MinReplicas).To(Equal(int32(3)))
			Expect(hpa.Spec.MaxReplicas).To(Equal(int32(80)))
			Expect(hpa.OwnerReferences).NotTo(BeEmpty())
		})
	})
})
