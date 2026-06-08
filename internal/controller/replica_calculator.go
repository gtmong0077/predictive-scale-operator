package controller

import (
	"math"
	"sort"

	autoscalingv1 "github.com/gtmong0077/predictive-scale-operator/api/v1"
)

func ComputeTargetReplicas(expectedPeakRPS, podStableRPS int32, util float64, min, max int32) int32 {
	if expectedPeakRPS <= 0 || podStableRPS <= 0 || util <= 0 {
		return min
	}

	denominator := float64(podStableRPS) * util
	required := int32(math.Ceil(float64(expectedPeakRPS) / denominator))

	if required < min {
		return min
	}
	if required > max {
		return max
	}
	return required
}

func ComputePreScaleSeconds(profile autoscalingv1.StartupProfile, warmupSeconds int32) int32 {
	total := profile.DetectDecideSeconds +
		profile.NodeSeconds +
		profile.PodReadySeconds +
		warmupSeconds +
		profile.SafetyMarginSeconds
	if total < 0 {
		return 0
	}
	return total
}

func StepTarget(final int32, ratio float64) int32 {
	if final <= 0 || ratio <= 0 {
		return 1
	}

	target := int32(math.Ceil(float64(final) * ratio))
	if target < 1 {
		return 1
	}
	return target
}

func sortStepsByOffsetDesc(steps []autoscalingv1.StepPolicy) []autoscalingv1.StepPolicy {
	sorted := append([]autoscalingv1.StepPolicy{}, steps...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].OffsetSeconds > sorted[j].OffsetSeconds
	})
	return sorted
}

// resolvePreScaleStep returns the active step index and desired minReplicas during PRE_SCALING.
// Steps advance by time; readiness gates progression to the next step unless openTime is near.
func resolvePreScaleStep(
	final int32,
	minReplicas int32,
	steps []autoscalingv1.StepPolicy,
	readyReplicas int32,
	secondsUntilOpen int64,
) (stepIndex int32, desiredMin int32) {
	if len(steps) == 0 {
		return -1, final
	}

	sorted := sortStepsByOffsetDesc(steps)
	activeStep := int32(-1)
	desiredMin = minReplicas

	for i, step := range sorted {
		if secondsUntilOpen > int64(step.OffsetSeconds) {
			break
		}

		if i > 0 {
			prevTarget := StepTarget(final, sorted[i-1].Ratio)
			if readyReplicas < prevTarget && secondsUntilOpen > 30 {
				break
			}
		}

		activeStep = int32(i)
		desiredMin = StepTarget(final, step.Ratio)
	}

	if activeStep < 0 {
		return -1, minReplicas
	}
	return activeStep, desiredMin
}
