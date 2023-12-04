package adaptationengine

import (
	"fmt"
	sascomv1 "github.com/yashvardhan-kukreja/consulkv-commander/api/v1"
	"github.com/yashvardhan-kukreja/consulkv-commander/internal/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (c Client) selfProtect(item *sascomv1.ConsulKV, invalidationsOutput utils.InvalidationsOutput, configMapPayloadUntilNow map[string]string, raisePager bool) (map[string]string, error) {
	defer func() {
		c.invalidationsTrackingContext.SetInvalidationsOutput(client.ObjectKeyFromObject(item).String(), invalidationsOutput, string(sascomv1.SelfProtecting))
	}()
	var urgencyLevel UrgencyLevel
	var pagerBody string

	switch item.Spec.QoS {
	case sascomv1.Critical:
		urgencyLevel = HighUrgencyLevel
		pagerBody = fmt.Sprintf("A KV group (%s) was found to leak some sensitive data"+
			"\nDetails:"+
			"\n%s", client.ObjectKeyFromObject(item).String(), invalidationsOutput)
	case sascomv1.Medium:
		urgencyLevel = LowUrgencyLevel
		pagerBody = fmt.Sprintf("A KV group (%s) was found to leak some sensitive data"+
			"\nDetails:"+
			"\n%s", client.ObjectKeyFromObject(item).String(), invalidationsOutput)
	default: // including Relaxed mode
		raisePager = false
	}

	_ = c.adaptSheet(item, invalidationsOutput)
	if raisePager {
		pagerBody = pagerBody + "[NOTE]" +
			"\nMitigation: The configmap in the cluster is rendered to ignore all the aforementioned sensitive keys."
		_ = c.RaisePager(urgencyLevel, pagerBody)
	}

	sanitizedConfigMapPayload := utils.RemoveKeysFromMap(configMapPayloadUntilNow, invalidationsOutput.Paths())
	return sanitizedConfigMapPayload, nil
}
