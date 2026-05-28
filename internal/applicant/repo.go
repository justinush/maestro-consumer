package applicant

import (
	"context"

	"github.com/justinush/maestro-consumer/internal/model"
)

type Repository interface {
	Create(ctx context.Context, applicantID, runID string) error
	GetByRunID(ctx context.Context, runID string) (*model.ApplicantRecord, error)
	SaveProfile(ctx context.Context, runID string, p model.Profile) error
	AddDocument(ctx context.Context, runID string, d model.Document) error
}
