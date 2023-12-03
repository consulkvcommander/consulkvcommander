package secretengine

import (
	"fmt"
	k8sStrings "k8s.io/utils/strings"
	"regexp"
	"strings"
)

func isPathWhitelisted(path string, whitelist []string) bool {
	for _, whitelistedPath := range whitelist {
		whitelistedPathIsDirectory := string(whitelistedPath[len(whitelistedPath)-1]) == "/"
		if !whitelistedPathIsDirectory {
			return whitelistedPath == path
		}
		// every key under this whitelisted directory would be considered whitelisted
		path = strings.ReplaceAll(path, ".", "/")
		if k8sStrings.ShortenString(path, len(whitelistedPath)) == whitelistedPath {
			return true
		}
	}
	return false
}

func validate(value string, validationRegexes []string) (matchesRegex bool, matchingRegexp string, outputErr error) {
	for _, regex := range validationRegexes {
		r, err := regexp.Compile(regex)
		if err != nil {
			matchesRegex = true
			matchingRegexp = regex
			outputErr = fmt.Errorf("validation regex '%s' failed to get compiled: %w", regex, err)
			return
		}
		if r.MatchString(value) {
			matchesRegex = true
			matchingRegexp = regex
			return
		}
	}
	matchesRegex = false
	return
}
