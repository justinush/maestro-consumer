package kyc

import (
	"context"

	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/workflow"
)

const (
	workflowIDVendor       = "kyc.sg.vendor"
	stepCreateVerification = "create-verification"
	stepWaitVendorResult   = "wait-vendor-result"
)

func isVendorWorkflow(key workflow.Key) bool {
	return key.ID == workflowIDVendor
}

// registerVendorSession runs after the create action step, before persisting the run.
func (s *Service) registerVendorSession(ctx context.Context, key workflow.Key, runID string, in *engine.Instance) error {
	if !isVendorWorkflow(key) {
		return nil
	}
	if in.CurrentStepID() != stepWaitVendorResult {
		return nil
	}
	_, err := s.vendor.EnsureSession(ctx, runID, stepCreateVerification)
	return err
}

func (s *Service) enrichExternalRef(ctx context.Context, resp StatusResponse, runID string) StatusResponse {
	sess, err := s.vendor.LookupByRunID(ctx, runID)
	if err != nil {
		return resp
	}
	resp.ExternalRef = sess.ExternalRef
	return resp
}
