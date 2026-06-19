package kyc

import (
	"context"

	"github.com/justinush/maestro-consumer/internal/vendor"
	"github.com/justinush/maestro/pkg/engine"
)

// vendorCreateSessionRunner runs the vendor-create-session onEnter action.
type vendorCreateSessionRunner struct {
	store vendor.Store
}

// NewVendorCreateSessionRunner returns an ActionRunner that idempotently creates
// a vendor session and mirrors externalRef into workflow variables.
func NewVendorCreateSessionRunner(store vendor.Store) engine.ActionRunner {
	return &vendorCreateSessionRunner{store: store}
}

func (r *vendorCreateSessionRunner) Run(ctx engine.ActionContext) error {
	sess, err := r.store.EnsureSession(context.Background(), ctx.RunID, ctx.StepID)
	if err != nil {
		return err
	}
	ctx.Variables["externalRef"] = sess.ExternalRef
	return nil
}
