package vendor

import "context"

// Session is the host-side mapping from vendor externalRef to a Maestro run.
type Session struct {
	ExternalRef  string
	RunID        string
	ExpectedStep string
}

// Store persists vendor session and webhook idempotency data.
type Store interface {
	EnsureSession(ctx context.Context, runID, stepID string) (Session, error)
	LookupByRunID(ctx context.Context, runID string) (Session, error)
	LookupByExternalRef(ctx context.Context, externalRef string) (Session, error)
	TryRecordWebhookEvent(ctx context.Context, eventID, runID string) (recorded bool, err error)
}
