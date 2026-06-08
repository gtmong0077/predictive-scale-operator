package controller

import (
	"testing"

	autoscalingv1 "github.com/gtmong0077/predictive-scale-operator/api/v1"
)

func TestComputeTargetReplicas(t *testing.T) {
	tests := []struct {
		name     string
		peakRPS  int32
		stable   int32
		util     float64
		min      int32
		max      int32
		expected int32
	}{
		{
			name:     "concert example",
			peakRPS:  5000,
			stable:   120,
			util:     0.7,
			min:      3,
			max:      80,
			expected: 60,
		},
		{
			name:     "clamped to min",
			peakRPS:  10,
			stable:   120,
			util:     0.7,
			min:      3,
			max:      80,
			expected: 3,
		},
		{
			name:     "clamped to max",
			peakRPS:  100000,
			stable:   120,
			util:     0.7,
			min:      3,
			max:      80,
			expected: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeTargetReplicas(tt.peakRPS, tt.stable, tt.util, tt.min, tt.max)
			if got != tt.expected {
				t.Fatalf("ComputeTargetReplicas() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestComputePreScaleSeconds(t *testing.T) {
	profile := autoscalingv1.StartupProfile{
		DetectDecideSeconds: 20,
		NodeSeconds:         0,
		PodReadySeconds:     90,
		SafetyMarginSeconds: 30,
	}

	got := ComputePreScaleSeconds(profile, 30)
	want := int32(170)
	if got != want {
		t.Fatalf("ComputePreScaleSeconds() = %d, want %d", got, want)
	}
}

func TestStepTarget(t *testing.T) {
	tests := []struct {
		final    int32
		ratio    float64
		expected int32
	}{
		{final: 60, ratio: 0.4, expected: 24},
		{final: 60, ratio: 0.75, expected: 45},
		{final: 60, ratio: 1.0, expected: 60},
	}

	for _, tt := range tests {
		got := StepTarget(tt.final, tt.ratio)
		if got != tt.expected {
			t.Fatalf("StepTarget(%d, %v) = %d, want %d", tt.final, tt.ratio, got, tt.expected)
		}
	}
}

func TestResolvePreScaleStep(t *testing.T) {
	steps := []autoscalingv1.StepPolicy{
		{OffsetSeconds: 180, Ratio: 0.4},
		{OffsetSeconds: 90, Ratio: 0.75},
		{OffsetSeconds: 30, Ratio: 1.0},
	}

	stepIndex, desired := resolvePreScaleStep(60, 3, steps, 0, 120)
	if stepIndex != 0 {
		t.Fatalf("stepIndex = %d, want 0", stepIndex)
	}
	if desired != 24 {
		t.Fatalf("desiredMin = %d, want 24", desired)
	}

	stepIndex, desired = resolvePreScaleStep(60, 3, steps, 24, 60)
	if stepIndex != 1 {
		t.Fatalf("stepIndex = %d, want 1", stepIndex)
	}
	if desired != 45 {
		t.Fatalf("desiredMin = %d, want 45", desired)
	}
}
