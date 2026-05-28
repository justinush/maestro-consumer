package applicant

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/justinush/maestro-consumer/internal/model"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func NewPostgres(pool *pgxpool.Pool) *Postgres {
	return &Postgres{pool: pool}
}

func (r *Postgres) Create(ctx context.Context, applicantID, runID string) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO applicants (applicant_id, run_id, documents)
		VALUES ($1, $2, '[]'::jsonb)
	`, applicantID, runID)
	if err != nil {
		return fmt.Errorf("applicant create: %w", err)
	}
	return nil
}

func (r *Postgres) GetByRunID(ctx context.Context, runID string) (*model.ApplicantRecord, error) {
	var (
		rec         model.ApplicantRecord
		profileJSON []byte
		docsJSON    []byte
	)

	err := r.pool.QueryRow(ctx, `
		SELECT applicant_id, run_id, COALESCE(profile, 'null'::jsonb), documents
		FROM applicants
		WHERE run_id = $1
	`, runID).Scan(&rec.ApplicantID, &rec.RunID, &profileJSON, &docsJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("%w: run %q", model.ErrNotFound, runID)
	}
	if err != nil {
		return nil, fmt.Errorf("applicant get: %w", err)
	}

	if string(profileJSON) != "null" {
		if err := json.Unmarshal(profileJSON, &rec.Profile); err != nil {
			return nil, fmt.Errorf("applicant profile json: %w", err)
		}
	}
	if len(docsJSON) > 0 {
		if err := json.Unmarshal(docsJSON, &rec.Documents); err != nil {
			return nil, fmt.Errorf("applicant documents json: %w", err)
		}
	}

	return &rec, nil
}

func (r *Postgres) SaveProfile(ctx context.Context, runID string, p model.Profile) error {
	profileJSON, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("profile json: %w", err)
	}

	tag, err := r.pool.Exec(ctx, `
		UPDATE applicants
		SET profile = $2::jsonb
		WHERE run_id = $1
	`, runID, profileJSON)
	if err != nil {
		return fmt.Errorf("save profile: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: run %q", model.ErrNotFound, runID)
	}
	return nil
}

func (r *Postgres) AddDocument(ctx context.Context, runID string, d model.Document) error {
	wrap, err := json.Marshal([]model.Document{d})
	if err != nil {
		return fmt.Errorf("document json: %w", err)
	}

	tag, err := r.pool.Exec(ctx, `
		UPDATE applicants
		SET documents = documents || $2::jsonb
		WHERE run_id = $1
	`, runID, wrap)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("%w: run %q", model.ErrNotFound, runID)
	}
	return nil
}
