package adaptationengine

import (
	"fmt"
	sascomv1 "github.com/yashvardhan-kukreja/consulkv-commander/api/v1"
	"github.com/yashvardhan-kukreja/consulkv-commander/internal/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func (c Client) selfHeal(item *sascomv1.KVGroup, invalidationsOutput utils.InvalidationsOutput, configMapPayloadUntilNow map[string]string, raisePager bool) (map[string]string, error) {
	consulKvClient := utils.NewConsulKV(item.Spec.ConsulUrl)

	failedDeletions := utils.InvalidationsOutput{}
	for _, inv := range invalidationsOutput {
		inv := inv
		slashedPath := strings.ReplaceAll(inv.Path, ".", "/")
		if err := consulKvClient.DeletePath(slashedPath); err != nil {
			fmt.Printf("failed to DELETE the key at the path %s: %s\n", inv.Path, err.Error())
			failedDeletions = append(failedDeletions, inv)
		}
	}

	var urgencyLevel UrgencyLevel
	var pagerBody string

	switch item.Spec.QoS {
	case sascomv1.Critical:
		pagerBody = fmt.Sprintf("A KV group (%s) was found to leak some sensitive data"+
			"\nDetails:"+
			"\n%s", client.ObjectKeyFromObject(item).String(), invalidationsOutput)

		if len(failedDeletions) != 0 {
			raisePager = true
			urgencyLevel = HighUrgencyLevel
			pagerBody += fmt.Sprintf("\nSome keys failed to get deleted"+
				"\nDetails:"+
				"\n%s", failedDeletions)
		} else {
			urgencyLevel = HighUrgencyLevel
			pagerBody = "[ALREADY SAFELY TAKEN CARE OF BY GETTING RID OF THE KEYS]\n" + pagerBody
		}
	case sascomv1.Medium:
		urgencyLevel = LowUrgencyLevel
		pagerBody = fmt.Sprintf("A KV group (%s) was found to leak some sensitive data"+
			"\nDetails:"+
			"\n%s", client.ObjectKeyFromObject(item).String(), invalidationsOutput)

		if len(failedDeletions) != 0 {
			pagerBody += fmt.Sprintf("\nSome keys failed to get deleted"+
				"\nDetails:"+
				"\n%s", failedDeletions)
		} else {
			pagerBody = "[ALREADY SAFELY TAKEN CARE OF BY GETTING RID OF THE KEYS]\n" + pagerBody
		}
	default: // including Relaxed mode
		raisePager = false
	}

	if len(failedDeletions) != 0 {
		_ = c.adaptSheet(item, failedDeletions)
	}

	defer func() {
		c.invalidationsTrackingContext.SetInvalidationsOutput(client.ObjectKeyFromObject(item).String(), failedDeletions, string(sascomv1.SelfHealing))
	}()

	if raisePager {
		_ = c.RaisePager(urgencyLevel, pagerBody)
	}

	sanitizedConfigMapPayload := utils.RemoveKeysFromMap(configMapPayloadUntilNow, invalidationsOutput.Paths())
	return sanitizedConfigMapPayload, nil
}
