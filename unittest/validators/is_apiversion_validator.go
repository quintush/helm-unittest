package validators

import "github.com/lrills/helm-unittest/unittest/common"

// IsAPIVersionValidator validate apiVersion of manifest is Of
type IsAPIVersionValidator struct {
	Of string
}

func (v IsAPIVersionValidator) failInfo(actual interface{}, not bool) []string {
	var notAnnotation string
	if not {
		notAnnotation = " NOT to be"
	}
	isAPIVersionFailFormat := "Expected" + notAnnotation + " apiVersion:%s"
	if not {
		return splitInfof(isAPIVersionFailFormat, v.Of)
	}
	return splitInfof(isAPIVersionFailFormat+"\nActual:%s", v.Of, common.TrustedMarshalYAML(actual))
}

// Validate implement Validatable
func (v IsAPIVersionValidator) Validate(context *ValidateContext) (bool, []string) {
	manifests, err := context.getManifests()
	if err != nil {
		return false, splitInfof(errorFormat, err.Error())
	}

	validateSuccess := true
	validateErrors := make([]string, 0)

	for _, manifest := range manifests {
		if kind, ok := manifest["apiVersion"].(string); (ok && kind == v.Of) == context.Negative {
			validateSuccess = validateSuccess && false
			errorMessage := v.failInfo(manifest["apiVersion"], context.Negative)
			validateErrors = append(validateErrors, errorMessage...)
			continue
		}

		validateSuccess = validateSuccess && true
	}

	return validateSuccess, validateErrors
}
