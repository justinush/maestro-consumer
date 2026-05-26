package persist

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/justinush/maestro/pkg/run"
)

// RunStore implements run.Store with Postgres (state as JSONB).
type RunStore struct {
	pool *pgxpool.Pool
}

func NewRunStore(pool *pgxpool.Pool) *RunStore {
	return &RunStore{pool: pool}
}

func (s *RunStore) Create(ctx context.Context, rec *run.RunRecord) error {
	if rec == nil || rec.RunID == "" {
		return fmt.Errorf("persist: invalid record")
	}
	stored, err := cloneRecord(rec)
	if err != nil {
		return err
	}
	stored.Revision = 1

	stateJSON, err := json.Marshal(stored.State)
	if err != nil {
		return fmt.Errorf("persist: marshal state: %w", err)
	}

	tag, err := s.pool.Exec(ctx, `
		INSERT INTO workflow_runs (run_id, workflow_id, workflow_version, revision, state)
		VALUES ($1, $2, $3, $4, $5::jsonb)
	`, stored.RunID, stored.WorkflowID, stored.WorkflowVersion, stored.Revision, stateJSON)
	if err != nil {
		return fmt.Errorf("persist: insert: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("persist: insert: no rows")
	}
	return nil
}

func (s *RunStore) Get(ctx context.Context, runID string) (*run.RunRecord, error) {
	if runID == "" {
		return nil, run.ErrNotFound
	}
	var (
		rec       run.RunRecord
		stateJSON []byte
	)
	err := s.pool.QueryRow(ctx, `
		SELECT run_id, workflow_id, workflow_version, revision, state
		FROM workflow_runs
		WHERE run_id = $1
	`, runID).Scan(&rec.RunID, &rec.WorkflowID, &rec.WorkflowVersion, &rec.Revision, &stateJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, run.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("persist: get: %w", err)
	}
	if err := json.Unmarshal(stateJSON, &rec.State); err != nil {
		return nil, fmt.Errorf("persist: unmarshal state: %w", err)
	}
	return cloneRecord(&rec)
}

func (s *RunStore) Save(ctx context.Context, rec *run.RunRecord) error {
	if rec == nil || rec.RunID == "" {
		return run.ErrNotFound
	}
	stored, err := cloneRecord(rec)
	if err != nil {
		return err
	}

	stateJSON, err := json.Marshal(stored.State)
	if err != nil {
		return fmt.Errorf("persist: marshal state: %w", err)
	}

	tag, err := s.pool.Exec(ctx, `
		UPDATE workflow_runs
		SET workflow_id = $2,
		    workflow_version = $3,
		    revision = revision + 1,
		    state = $4::jsonb,
		    updated_at = now()
		WHERE run_id = $1 AND revision = $5
	`, stored.RunID, stored.WorkflowID, stored.WorkflowVersion, stateJSON, stored.Revision)
	if err != nil {
		return fmt.Errorf("persist: save: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return nil
	}

	var exists bool
	err = s.pool.QueryRow(ctx, `SELECT true FROM workflow_runs WHERE run_id = $1`, stored.RunID).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return run.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("persist: save check: %w", err)
	}
	return run.ErrRevisionConflict
}

func cloneRecord(r *run.RunRecord) (*run.RunRecord, error) {
	if r == nil {
		return nil, nil
	}
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	var out run.RunRecord
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
