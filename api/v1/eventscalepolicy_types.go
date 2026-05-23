package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TargetRef는 조작할 HPA 리소스를 가리킵니다.
type TargetRef struct {
	Name string `json:"name"`
}

// EventScalePolicySpec은 친구분과 정의한 핵심 필드들을 포함합니다.
type EventScalePolicySpec struct {
	TargetRef TargetRef   `json:"targetRef"`
	OpenTime  metav1.Time `json:"openTime"`

	// +kubebuilder:validation:Minimum=1
	TargetReplicas int32 `json:"targetReplicas"`

	// +kubebuilder:validation:Minimum=0
	PreScaleSeconds int32 `json:"preScaleSeconds"`

	// +kubebuilder:validation:Minimum=0
	RestoreAfterSeconds int32 `json:"restoreAfterSeconds"`

	// +kubebuilder:validation:Minimum=1
	MinReplicasAfterEvent int32 `json:"minReplicasAfterEvent"`
}

// EventScalePolicyStatus는 오퍼레이터가 관리하는 현재 상태를 기록합니다.
type EventScalePolicyStatus struct {
	Phase string `json:"phase,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="TargetReplicas",type=integer,JSONPath=`.spec.targetReplicas`

type EventScalePolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   EventScalePolicySpec   `json:"spec,omitempty"`
	Status EventScalePolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
type EventScalePolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []EventScalePolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&EventScalePolicy{}, &EventScalePolicyList{})
}
