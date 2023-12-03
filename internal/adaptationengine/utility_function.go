package adaptationengine

import (
	sascomv1 "github.com/yashvardhan-kukreja/consulkv-commander/api/v1"
	"github.com/yashvardhan-kukreja/consulkv-commander/internal/utils"
)

func (c Client) utilityFunction(invalidationsOutput utils.InvalidationsOutput, pathToWeights map[string]int) (float32, sascomv1.AdaptationMode, bool) {
	var numerator, denominator int

	for _, weight := range pathToWeights {
		denominator += weight
	}

	for _, inv := range invalidationsOutput {
		weight := pathToWeights[inv.Path]
		numerator += weight
	}

	if denominator == 0 {
		return -1, sascomv1.NonAdaptive, false
	}

	utilityValue := 1 - (float32(numerator) / float32(denominator))

	var adaptationMode sascomv1.AdaptationMode
	if utilityValue >= 0 && utilityValue <= 0.3 {
		adaptationMode = sascomv1.SelfHealing
	} else if utilityValue <= 0.8 {
		adaptationMode = sascomv1.SelfProtecting
	} else {
		adaptationMode = sascomv1.NonAdaptive
	}

	raisePager := false
	if utilityValue <= 0.95 {
		raisePager = true
	}

	return utilityValue, adaptationMode, raisePager
}
