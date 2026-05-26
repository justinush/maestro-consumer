package applicant

import "github.com/justinush/maestro-consumer/internal/kyc"

type Repository interface {
	Create(applicantID, runID string) error
	GetByRunID(runID string) (*kyc.ApplicantRecord, error)
	SaveProfile(runID string, p kyc.Profile) error
	AddDocument(runID string, d kyc.Document) error
}
