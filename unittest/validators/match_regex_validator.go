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

func (v MatchRegexValidator) failInfo(actual string, index int, not bool) []string {
	var notAnnotation = ""
	if not {
		notAnnotation = " NOT"
	}
	regexFailFormat := `
Path:%s
Expected` + notAnnotation + ` to match:%s
Actual:%s
`
	return splitInfof(regexFailFormat, index, v.Path, v.Pattern, actual)
}

// Validate implement Validatable
func (v MatchRegexValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, -1, err.Error())
	}

	validateSuccess := true
	validateErrors := make([]string, 0)

	for idx, manifest := range manifests {
		actual, err := valueutils.GetValueOfSetPath(manifest, v.Path)
		if err != nil {
			validateSuccess = validateSuccess && false
			errorMessage := splitInfof(errorFormat, idx, err.Error())
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		p, err := regexp.Compile(v.Pattern)
		if err != nil {
			validateSuccess = validateSuccess && false
			errorMessage := splitInfof(errorFormat, -1, err.Error())
			validateErrors = append(validateErrors, errorMessage...)
			break
		}

		if s, ok := actual.(string); ok {
			if p.MatchString(s) == context.Negative {
				validateSuccess = validateSuccess && false
				errorMessage := v.failInfo(s, idx, context.Negative)
				validateErrors = append(validateErrors, errorMessage...)
				continue
			}

			validateSuccess = validateSuccess && true
			continue
		}

		validateSuccess = validateSuccess && false
		errorMessage := splitInfof(errorFormat, idx, fmt.Sprintf(
			"expect '%s' to be a string, got:\n%s",
			v.Path,
			common.TrustedMarshalYAML(actual),
		))
		validateErrors = append(validateErrors, errorMessage...)
	}

	return validateSuccess, validateErrors
}
