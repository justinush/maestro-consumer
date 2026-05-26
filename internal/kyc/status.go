package kyc

import "github.com/justinush/maestro/pkg/engine"

type StatusResponse struct {
	RunID       string     `json:"runId"`
	ApplicantID string     `json:"applicantId"`
	Status      string     `json:"status"`
	Step        string     `json:"step"`
	Terminal    bool       `json:"terminal"`
	Profile     *Profile   `json:"profile,omitempty"`
	Documents   []Document `json:"documents,omitempty"`
}

type EventsResponse struct {
	RunID  string   `json:"runId"`
	Events []string `json:"events"`
}

func BuildStatus(app *ApplicantRecord, in *engine.Instance, completed bool) StatusResponse {
	step := in.CurrentStepID()
	resp := StatusResponse{
		RunID:       app.RunID,
		ApplicantID: app.ApplicantID,
		Step:        step,
		Terminal:    completed || in.IsTerminal(),
		Status:      mapStepToStatus(step, completed || in.IsTerminal()),
	}
	if app.Profile.FullName != "" || app.Profile.Email != "" {
		p := app.Profile
		resp.Profile = &p
	}
	if len(app.Documents) > 0 {
		resp.Documents = app.Documents
	}
	return resp
}

func mapStepToStatus(stepID string, terminal bool) string {
	if terminal && stepID == "approved" {
		return "approved"
	}
	switch stepID {
	case "collect-profile":
		return "awaiting_profile"
	case "document-upload":
		return "awaiting_document"
	case "run-liveness-check":
		return "processing_liveness"
	case "manual-review":
		return "awaiting_review"
	default:
		return "in_progress"
	}
}
