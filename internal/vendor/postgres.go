package vendor

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func idempotencyKey(runID, stepID string) string {
	return runID + ":" + stepID
}

func externalRefFromKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return "vs_" + hex.EncodeToString(sum[:8])
}

func (s *Postgres) EnsureSession(ctx context.Context, runID, stepID string) (Session, error) {
	key := idempotencyKey(runID, stepID)
	ref := externalRefFromKey(key)
	const wantStep = "wait-vendor-result"

	_, err := s.pool.Exec(ctx, `
		INSERT INTO vendor_sessions (external_ref, run_id, expected_step, idempotency_key)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (idempotency_key) DO NOTHING
	`, ref, runID, wantStep, key)
	if err != nil {
		return Session{}, fmt.Errorf("vendor ensure session: %w", err)
	}

	var sess Session
	err = s.pool.QueryRow(ctx, `
		SELECT external_ref, run_id, expected_step
		FROM vendor_sessions
		WHERE idempotency_key = $1
	`, key).Scan(&sess.ExternalRef, &sess.RunID, &sess.ExpectedStep)
	if err != nil {
		return Session{}, fmt.Errorf("vendor load session: %w", err)
	}
	return sess, nil
}

func (s *Postgres) LookupByRunID(ctx context.Context, runID string) (Session, error) {
	var sess Session
	err := s.pool.QueryRow(ctx, `
		SELECT external_ref, run_id, expected_step
		FROM vendor_sessions
		WHERE run_id = $1
	`, runID).Scan(&sess.ExternalRef, &sess.RunID, &sess.ExpectedStep)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, fmt.Errorf("%w: run %q", ErrNotFound, runID)
	}
	if err != nil {
		return Session{}, fmt.Errorf("vendor lookup by run: %w", err)
	}
	return sess, nil
}

func (s *Postgres) LookupByExternalRef(ctx context.Context, externalRef string) (Session, error) {
	var sess Session
	err := s.pool.QueryRow(ctx, `
		SELECT external_ref, run_id, expected_step
		FROM vendor_sessions
		WHERE external_ref = $1
	`, externalRef).Scan(&sess.ExternalRef, &sess.RunID, &sess.ExpectedStep)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, fmt.Errorf("%w: externalRef %q", ErrNotFound, externalRef)
	}
	if err != nil {
		return Session{}, fmt.Errorf("vendor lookup: %w", err)
	}
	return sess, nil
}

func (s *Postgres) TryRecordWebhookEvent(ctx context.Context, eventID, runID string) (bool, error) {
	tag, err := s.pool.Exec(ctx, `
		INSERT INTO webhook_events (event_id, run_id)
		VALUES ($1, $2)
		ON CONFLICT (event_id) DO NOTHING
	`, eventID, runID)
	if err != nil {
		return false, fmt.Errorf("vendor record webhook event: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}
