package kyc

import (
	"strings"

	"github.com/justinush/maestro-consumer/internal/model"
)

type ApplicantRecord = model.ApplicantRecord
type Profile = model.Profile
type Document = model.Document

// StartRequest selects which workflow to start (host routing).
type StartRequest struct {
	Entity string `json:"entity"`
	Flow   string `json:"flow"`
}

func (r StartRequest) Validate() error {
	if strings.TrimSpace(r.Entity) == "" || strings.TrimSpace(r.Flow) == "" {
		return ErrInvalid
	}
	return nil
}

// StatusResponse is the API view of a run plus demo applicant data.
type StatusResponse struct {
	RunID           string     `json:"runId"`
	ApplicantID     string     `json:"applicantId"`
	WorkflowID      string     `json:"workflowId,omitempty"`
	WorkflowVersion string     `json:"workflowVersion,omitempty"`
	Entity          string     `json:"entity,omitempty"`
	Flow            string     `json:"flow,omitempty"`
	ExternalRef     string     `json:"externalRef,omitempty"`
	Status          string     `json:"status"`
	Step            string     `json:"step"`
	Terminal        bool       `json:"terminal"`
	Profile         *Profile   `json:"profile,omitempty"`
	Documents       []Document `json:"documents,omitempty"`
}

type EventsResponse struct {
	RunID  string   `json:"runId"`
	Events []string `json:"events"`
}
