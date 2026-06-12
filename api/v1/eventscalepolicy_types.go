package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StepPolicy defines a pre-scale step relative to openTime.
type StepPolicy struct {
	OffsetSeconds int32   `json:"offsetSeconds"`
	Ratio         float64 `json:"ratio"`
}

// StartupProfile holds measured startup delays for preScaleSeconds calculation.
// preScaleSeconds = detectDecide + node + podReady + warmup + safetyMargin
type StartupProfile struct {
	DetectDecideSeconds int32 `json:"detectDecideSeconds,omitempty"`
	NodeSeconds         int32 `json:"nodeSeconds,omitempty"`
	PodReadySeconds     int32 `json:"podReadySeconds,omitempty"`
	SafetyMarginSeconds int32 `json:"safetyMarginSeconds,omitempty"`
}

// EventScalePolicySpec defines the desired state of EventScalePolicy.
type EventScalePolicySpec struct {
	TargetDeployment string      `json:"targetDeployment"`
	OpenTime         metav1.Time `json:"openTime"`

	ExpectedPeakRPS   int32   `json:"expectedPeakRPS"`
	PodStableRPS      int32   `json:"podStableRPS"`
	TargetUtilization float64 `json:"targetUtilization"`
	MinReplicas       int32   `json:"minReplicas"`
	MaxReplicas       int32   `json:"maxReplicas"`

	ReadinessPercentile string         `json:"readinessPercentile,omitempty"`
	WarmupSeconds       int32          `json:"warmupSeconds"`
	StartupProfile      StartupProfile `json:"startupProfile"`

	StepPolicy []StepPolicy `json:"stepPolicy"`

	ScaleDownGuardMinutes   int32 `json:"scaleDownGuardMinutes"`
	MaxOverprovisionMinutes int32 `json:"maxOverprovisionMinutes"`

	MetricPolicy []string `json:"metricPolicy,omitempty"`
}

// EventScalePolicyStatus defines the observed state of EventScalePolicy.
type EventScalePolicyStatus struct {
	Phase string `json:"phase,omitempty"`

	ComputedTargetReplicas int32 `json:"computedTargetReplicas,omitempty"`
	CurrentStep            int32 `json:"currentStep,omitempty"`
	ReadyReplicas          int32 `json:"readyReplicas,omitempty"`
	DesiredMinReplicas     int32 `json:"desiredMinReplicas,omitempty"`

	PreScaleStartTime  *metav1.Time `json:"preScaleStartTime,omitempty"`
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Target",type=integer,JSONPath=`.status.computedTargetReplicas`
// +kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="OpenTime",type=date,JSONPath=`.spec.openTime`

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
