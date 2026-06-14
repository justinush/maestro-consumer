package kyc

import (
	"context"
	"fmt"
	"strings"
)

// VendorWebhookRequest is the demo vendor callback body.
type VendorWebhookRequest struct {
	ExternalRef string `json:"externalRef"`
	Status      string `json:"status"`
	EventID     string `json:"eventId"`
}

func (r VendorWebhookRequest) Validate() error {
	if strings.TrimSpace(r.ExternalRef) == "" {
		return ErrInvalid
	}
	if strings.TrimSpace(r.Status) == "" {
		return ErrInvalid
	}
	if strings.TrimSpace(r.EventID) == "" {
		return ErrInvalid
	}
	return nil
}

func (s *Service) HandleVendorWebhook(ctx context.Context, req VendorWebhookRequest) (StatusResponse, error) {
	if err := req.Validate(); err != nil {
		return StatusResponse{}, err
	}

	sess, err := s.vendor.LookupByExternalRef(ctx, req.ExternalRef)
	if err != nil {
		return StatusResponse{}, fmt.Errorf("%w: %w", ErrNotFound, err)
	}

	recorded, err := s.vendor.TryRecordWebhookEvent(ctx, req.EventID, sess.RunID)
	if err != nil {
		return StatusResponse{}, err
	}
	if !recorded {
		// Callback idempotency - vendor redelivered same eventId.
		return s.Get(ctx, sess.RunID)
	}

	return s.submitOnStep(ctx, sess.RunID, sess.ExpectedStep, map[string]any{
		"status":      req.Status,
		"externalRef": req.ExternalRef,
	}, nil)
}
