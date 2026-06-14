package kyc

import (
	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
)

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

func withWorkflowMeta(resp StatusResponse, def *definition.WorkflowDefinition) StatusResponse {
	if def != nil {
		resp.WorkflowID = def.ID
		resp.WorkflowVersion = def.Version
	}
	return resp
}

func withRouteMeta(resp StatusResponse, entity, flow string) StatusResponse {
	resp.Entity = entity
	resp.Flow = flow
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
	case "wait-vendor-result":
		return "awaiting_vendor_callback"
	default:
		return "in_progress"
	}
}
