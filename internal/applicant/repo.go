package applicant

import "github.com/justinush/maestro-consumer/internal/model"

type Repository interface {
	Create(applicantID, runID string) error
	GetByRunID(runID string) (*model.ApplicantRecord, error)
	SaveProfile(runID string, p model.Profile) error
	AddDocument(runID string, d model.Document) error
}
