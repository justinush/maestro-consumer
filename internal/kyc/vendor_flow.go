package kyc

import (
	"context"
)

const (
	workflowIDVendor     = "kyc.sg.vendor"
	stepWaitVendorResult = "wait-vendor-result"
)

func (s *Service) enrichExternalRef(ctx context.Context, resp StatusResponse, runID string) StatusResponse {
	sess, err := s.vendor.LookupByRunID(ctx, runID)
	if err != nil {
		return resp
	}
	resp.ExternalRef = sess.ExternalRef
	return resp
}
