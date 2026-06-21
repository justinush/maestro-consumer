package kyc

import "github.com/justinush/maestro/pkg/validate"

// WorkflowValidateOptions returns validate.Options for workflow.LoadDir.
// Keep AllowedActionTypes in sync with NewActionRegistry registrations.
func WorkflowValidateOptions() validate.Options {
	return validate.Options{
		AllowedActionTypes: []string{
			ActionTypeVendorCreateSession,
		},
	}
}
