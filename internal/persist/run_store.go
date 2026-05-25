package persist

import (
	"context"
	"encoding/json"
	"fmt"

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
