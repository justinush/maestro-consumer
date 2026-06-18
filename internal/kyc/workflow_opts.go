package kyc

import "github.com/justinush/maestro/pkg/validate"

func WorkflowValidateOptions() validate.Options {
	return validate.Options{
		AllowedActionTypes: []string{
			ActionTypeVendorCreateSession,
		},
	}
}
