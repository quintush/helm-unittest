package validators

import (
	"fmt"
	"regexp"

	"github.com/lrills/helm-unittest/unittest/common"
	"github.com/lrills/helm-unittest/unittest/valueutils"
)

// MatchRegexValidator validate value of Path match Pattern
type MatchRegexValidator struct {
	Path    string
	Pattern string
}

func (v MatchRegexValidator) failInfo(actual string, not bool) []string {
	var notAnnotation = ""
	if not {
		notAnnotation = " NOT"
	}
	regexFailFormat := `
Path:%s
Expected` + notAnnotation + ` to match:%s
Actual:%s
`
	return splitInfof(regexFailFormat, v.Path, v.Pattern, actual)
}

// Validate implement Validatable
func (v MatchRegexValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	validateSuccess := true
	validateErrors := make([]string, 0)

	for _, manifest := range manifests {
		actual, err := valueutils.GetValueOfSetPath(manifest, v.Path)
		if err != nil {
			validateSuccess = validateSuccess && false
			errorMessage := splitInfof(errorFormat, err.Error())
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		p, err := regexp.Compile(v.Pattern)
		if err != nil {
			validateSuccess = validateSuccess && false
			errorMessage := splitInfof(errorFormat, err.Error())
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		if s, ok := actual.(string); ok {
			if p.MatchString(s) == context.Negative {
				validateSuccess = validateSuccess && false
				errorMessage := v.failInfo(s, context.Negative)
				validateErrors = append(validateErrors, errorMessage...)
				continue
			}

			validateSuccess = validateSuccess && true
			continue
		}

		validateSuccess = validateSuccess && false
		errorMessage := splitInfof(errorFormat, fmt.Sprintf(
			"expect '%s' to be a string, got:\n%s",
			v.Path,
			common.TrustedMarshalYAML(actual),
		))
		validateErrors = append(validateErrors, errorMessage...)
	}

	return validateSuccess, validateErrors
}
