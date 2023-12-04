package secretengine

import (
	"fmt"
	sascomv1 "github.com/yashvardhan-kukreja/consulkv-commander/api/v1"
	"github.com/yashvardhan-kukreja/consulkv-commander/internal/adaptationengine"
	"github.com/yashvardhan-kukreja/consulkv-commander/internal/knowledgebase"
	"github.com/yashvardhan-kukreja/consulkv-commander/internal/utils"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	predefinedRegexAliases = map[string]string{
		"email":           "^[\\w-\\.]+@([\\w-]+\\.)+[\\w-]{2,4}$",
		"contact":         "^\\s*(?:\\+?(\\d{1,3}))?([-. (]*(\\d{3})[-. )]*)?((\\d{3})[-. ]*(\\d{2,4})(?:[-.x ]*(\\d+))?)\\s*$",
		"bitcoin-address": "([13][a-km-zA-HJ-NP-Z0-9]{26,33})",
		"badwords":        "\\b(?:(?:ass+(?:\\s+)?|i+(?:\\s+)?|butt+(?:\\s+)?|mo(?:(?:m|t|d)h?(?:e|a)?r?)(?:\\s+)?)?f(?:(?:\\s+)?u+)?(?:(?:\\s+)?c+)?(?:(?:\\s+)?k+)?(?:(?:e|a)(?:r+)?|i(?:n(?:g)?)?)?(?:s+)?(?:\\s+)?(?:hole|head|(?:yo?)?u?)?)+\\b",
		"html-tags":       "</?\\w+((\\s+\\w+(\\s*=\\s*(?:\".*?\"|'.*?'|[^'\">\\s]+))?)+\\s*|\\s*)/?>",
		"sql-injection":   "\"((SELECT|DELETE|UPDATE|INSERT INTO) (\\*|[A-Z0-9_]+) (FROM) ([A-Z0-9_]+))( (WHERE) ([A-Z0-9_]+) (=|<|>|>=|<=|==|!=) (\\?|\\$[A-Z]{1}[A-Z_]+)( (AND) ([A-Z0-9_]+) (=|<|>|>=|<=|==|!=) (\\?))?)?\"",
	}
)

type Client struct {
	invalidationsTrackingContext *knowledgebase.KnowledgeBaseContext
	advisoryLock                 *AdvisoryLock
	adaptationEngineClient       adaptationengine.Client
}

func NewClient(invalidationsTrackingContext *knowledgebase.KnowledgeBaseContext, adaptationEngineClient adaptationengine.Client) Client {
	return Client{
		invalidationsTrackingContext,
		NewAdvisoryLock(),
		adaptationEngineClient,
	}
}

func (s Client) Run(item *sascomv1.ConsulKV, configMapPayloadUntilNow map[string]string, pathToWeights map[string]int) (map[string]string, error) {
	consulKvKey := client.ObjectKeyFromObject(item).String()

	s.advisoryLock.Init(consulKvKey)

	s.advisoryLock.Lock(consulKvKey)
	defer s.advisoryLock.Unlock(consulKvKey)

	invalidationsOutput := getInvalidations(item, configMapPayloadUntilNow)

	sanitizedConfigMapPayload, err := s.adaptationEngineClient.Adapt(item, invalidationsOutput, configMapPayloadUntilNow, pathToWeights)
	if err != nil {
		return nil, fmt.Errorf("failed to adapt the system: %w", err)
	}

	return sanitizedConfigMapPayload, nil
}

func getInvalidations(consulKv *sascomv1.ConsulKV, configMapPayload map[string]string) utils.InvalidationsOutput {
	invalidationsOutput := []utils.Invalidation{}
	if len(consulKv.Spec.GuardAgainst) == 0 {
		return invalidationsOutput
	}

	validationRegexes := []string{}
	for _, guard := range consulKv.Spec.GuardAgainst {
		validationRegex := guard
		regexAlias, found := predefinedRegexAliases[guard]
		if found {
			validationRegex = regexAlias
		}
		validationRegexes = append(validationRegexes, validationRegex)
	}

	for pathToValidate, valueToValidate := range configMapPayload {
		// skip validation is pathToValidate was ultimately found to be whitelisted
		if isPathWhitelisted(pathToValidate, consulKv.Spec.WhitelistedPaths) {
			continue
		}

		matchesRegex, matchingRegexp, err := validate(valueToValidate, validationRegexes)
		if matchesRegex {

			invalidation := utils.Invalidation{
				Path:         pathToValidate,
				Value:        valueToValidate,
				FailingRegex: matchingRegexp,
			}
			if err != nil {
				invalidation.AnyError = err.Error()
			}
			invalidationsOutput = append(invalidationsOutput, invalidation)
		}
	}
	return invalidationsOutput
}
