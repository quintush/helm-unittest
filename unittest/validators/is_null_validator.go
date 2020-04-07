package validators

import (
	"github.com/lrills/helm-unittest/unittest/common"
	"github.com/lrills/helm-unittest/unittest/valueutils"
)

// IsNullValidator validate value of Path id kind
type IsNullValidator struct {
	Path string
}

func (v IsNullValidator) failInfo(actual interface{}, not bool) []string {
	var notAnnotation string
	if not {
		notAnnotation = " NOT"
	}

	isNullFailFormat := `
Path:%s
Expected` + notAnnotation + ` to be null, got:
%s
`
	return splitInfof(isNullFailFormat, v.Path, common.TrustedMarshalYAML(actual))
}

// Validate implement Validatable
func (v IsNullValidator) Validate(context *ValidateContext) (bool, []string) {
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

		if actual == nil == context.Negative {
			validateSuccess = validateSuccess && false
			errorMessage := v.failInfo(actual, context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		validateSuccess = validateSuccess && true
	}

	return validateSuccess, validateErrors
}
