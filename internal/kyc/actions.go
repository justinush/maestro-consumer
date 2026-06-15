package kyc

import (
	"github.com/justinush/maestro/pkg/engine"
)

// NewActionRegistry returns the action runners used by workflow YAML.
func NewActionRegistry() *engine.Registry {
	reg := engine.NewRegistry()
	reg.MustRegister("stub", engine.NewStubRunner())
	return reg
}
