package kyc

import (
	"github.com/justinush/maestro-consumer/internal/vendor"
	"github.com/justinush/maestro/pkg/engine"
)

const (
	// ActionTypeVendorCreateSession is the custom onEnter type for vendor outbound create.
	ActionTypeVendorCreateSession = "vendor-create-session"
)

// NewActionRegistry returns the action runners used by workflow YAML.
func NewActionRegistry(vendorStore vendor.Store) *engine.Registry {
	reg := engine.NewRegistry()
	reg.MustRegister("stub", engine.NewStubRunner())
	reg.MustRegister(ActionTypeVendorCreateSession, NewVendorCreateSessionRunner(vendorStore))
	return reg
}
