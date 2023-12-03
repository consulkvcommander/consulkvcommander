package adaptationengine

import (
	"fmt"
	"github.com/PagerDuty/go-pagerduty"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/google/go-cmp/cmp"
	sascomv1 "github.com/yashvardhan-kukreja/consulkv-commander/api/v1"
	"github.com/yashvardhan-kukreja/consulkv-commander/internal/customcontext"
	"github.com/yashvardhan-kukreja/consulkv-commander/internal/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

type Client struct {
	k8sClient                    client.Client
	pagerDutyClient              *pagerduty.Client
	pagerDutySender              string
	s3Session                    *session.Session
	sheetLink                    string
	invalidationsTrackingContext *customcontext.CustomContext
}

func NewClient(k8sClient client.Client, pdClient *pagerduty.Client, pdSender string, invalidationsTrackingContext *customcontext.CustomContext, sheetLink string, awsConfig *aws.Config) (Client, error) {
	s3Session, err := session.NewSession(awsConfig)
	if err != nil {
		return Client{}, fmt.Errorf("error occurred while setting up the S3 session for the secret engine client: %w", err)
	}
	return Client{
		k8sClient,
		pdClient,
		pdSender,
		s3Session,
		sheetLink,
		invalidationsTrackingContext,
	}, nil
}

func (c Client) Adapt(item *sascomv1.KVGroup, invalidationsOutput utils.InvalidationsOutput, configMapPayloadUntilNow map[string]string, pathToWeights map[string]int) (map[string]string, error) {
	utilityValue, adaptationMode, raisePager := c.utilityFunction(invalidationsOutput, pathToWeights)

	item.Status.UtilityFunctionValue = fmt.Sprintf("%v", utilityValue)
	item.Status.AdaptationMode = adaptationMode

	if canIgnorePagingInvalidationsOutput(c.invalidationsTrackingContext, client.ObjectKeyFromObject(item).String(), invalidationsOutput, string(adaptationMode)) {
		raisePager = false
	}

	switch adaptationMode {
	case sascomv1.SelfHealing:
		return c.selfHeal(item, invalidationsOutput, configMapPayloadUntilNow, raisePager)
	case sascomv1.SelfProtecting:
		return c.selfProtect(item, invalidationsOutput, configMapPayloadUntilNow, raisePager)
	default:
		return configMapPayloadUntilNow, nil
	}
}

func canIgnorePagingInvalidationsOutput(ctx *customcontext.CustomContext, kvGroupKey string, newInvalidationsOutput utils.InvalidationsOutput, adaptationMode string) bool {
	if len(newInvalidationsOutput) == 0 {
		return true
	}
	previousInvalidationOutput := ctx.GetLastInvalidationsOutput(kvGroupKey, adaptationMode)
	if !cmp.Equal(newInvalidationsOutput, previousInvalidationOutput) {
		return false
	}
	// if the last set of invalidations are same as the new ones, but they happened more than 30 minutes ago, re-consider them
	previousInvalidationsTime := ctx.GetLastInvalidationsTime(kvGroupKey, adaptationMode)
	if (time.Now().UnixMilli() - previousInvalidationsTime) > (1000 * 60 * 30) {
		return false
	}
	return true // new invalidations are same as previous invalidations and that too happened even under 30 minutes back
}
