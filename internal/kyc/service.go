package kyc

import (
	"context"
	"fmt"

	"github.com/justinush/maestro-consumer/internal/applicant"
	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/maestro"
	"github.com/justinush/maestro/pkg/run"
)

type Service struct {
	rt   *maestro.Runtime
	def  *definition.WorkflowDefinition
	runs run.Store
	apps applicant.Repository
}

func NewService(rt *maestro.Runtime, runs run.Store, apps applicant.Repository) *Service {
	return &Service{
		rt:   rt,
		def:  rt.Definition(),
		runs: runs,
		apps: apps,
	}
}

type afterSuccess func() error

func (s *Service) Start(ctx context.Context) (StatusResponse, error) {
	applicantID := newID("app")
	runID := newID("run")

	in, err := s.rt.NewInstance(maestro.InstanceOptions{
		RunID: runID,
		InitialVariables: map[string]any{
			"applicantId": applicantID,
		},
	})
	if err != nil {
		return StatusResponse{}, err
	}
	if err := DriveUntilBlocked(in); err != nil {
		return StatusResponse{}, err
	}
	if err := PersistNewRun(ctx, s.runs, in, s.def); err != nil {
		return StatusResponse{}, err
	}
	if err := s.apps.Create(applicantID, runID); err != nil {
		return StatusResponse{}, err
	}
	app, err := s.apps.GetByRunID(runID)
	if err != nil {
		return StatusResponse{}, err
	}
	return BuildStatus(app, in, false), nil
}

func (s *Service) Get(ctx context.Context, runID string) (StatusResponse, error) {
	app, err := s.apps.GetByRunID(runID)
	if err != nil {
		return StatusResponse{}, err
	}
	in, err := RestoreRun(ctx, s.rt, s.runs, runID)
	if err != nil {
		return StatusResponse{}, err
	}
	return BuildStatus(app, in, in.IsTerminal()), nil
}

func (s *Service) GetEvents(ctx context.Context, runID string) (EventsResponse, error) {
	if _, err := s.apps.GetByRunID(runID); err != nil {
		return EventsResponse{}, err
	}
	in, err := RestoreRun(ctx, s.rt, s.runs, runID)
	if err != nil {
		return EventsResponse{}, err
	}
	events := in.Events()
	lines := make([]string, len(events))
	for i := range events {
		lines[i] = events[i].String()
	}
	return EventsResponse{RunID: runID, Events: lines}, nil
}

func (s *Service) SubmitProfile(ctx context.Context, runID string, p Profile) (StatusResponse, error) {
	if err := p.Validate(); err != nil {
		return StatusResponse{}, err
	}
	return s.submitOnStep(ctx, runID, "collect-profile", map[string]any{
		"fullName": p.FullName,
		"email":    p.Email,
	}, func() error {
		return s.apps.SaveProfile(runID, p)
	})
}

func (s *Service) SubmitDocument(ctx context.Context, runID string, d Document) (StatusResponse, error) {
	if err := d.Validate(); err != nil {
		return StatusResponse{}, err
	}
	in, err := RestoreRun(ctx, s.rt, s.runs, runID)
	if err != nil {
		return StatusResponse{}, err
	}
	const want = "document-upload"
	if in.CurrentStepID() != want {
		return StatusResponse{}, fmt.Errorf("%w: want %q, at %q", ErrWrongStep, want, in.CurrentStepID())
	}
	app, err := s.apps.GetByRunID(runID)
	if err != nil {
		return StatusResponse{}, err
	}
	if err := FakeVendorCheckLiveness(app.ApplicantID); err != nil {
		return StatusResponse{}, err
	}
	needsReview := d.Type == "passport"
	return s.submitOnStep(ctx, runID, want, map[string]any{
		"documentType": d.Type,
		"documentRef":  d.Ref,
		"review":       map[string]any{"required": needsReview},
	}, func() error {
		return s.apps.AddDocument(runID, d)
	})
}

func (s *Service) SubmitReview(ctx context.Context, runID string, approved bool) (StatusResponse, error) {
	return s.submitOnStep(ctx, runID, "manual-review", map[string]any{
		"approved": approved,
		"review":   map[string]any{"approved": approved},
	}, nil)
}

func (s *Service) submitOnStep(
	ctx context.Context,
	runID, wantStep string,
	input map[string]any,
	after afterSuccess,
) (StatusResponse, error) {
	in, err := RestoreRun(ctx, s.rt, s.runs, runID)
	if err != nil {
		return StatusResponse{}, err
	}
	if in.CurrentStepID() != wantStep {
		return StatusResponse{}, fmt.Errorf("%w: want %q, at %q", ErrWrongStep, wantStep, in.CurrentStepID())
	}
	sub := in.SubmitInput(input)
	switch sub.Status {
	case engine.SubmitAdvanced, engine.SubmitStayOnStep:
	case engine.SubmitFailed:
		return StatusResponse{}, fmt.Errorf("submit: %w", sub.Err)
	default:
		return StatusResponse{}, fmt.Errorf("submit: unexpected status %v", sub.Status)
	}
	if err := DriveUntilBlocked(in); err != nil {
		return StatusResponse{}, err
	}
	if err := SaveRun(ctx, s.runs, runID, in, s.def); err != nil {
		return StatusResponse{}, err
	}
	if after != nil {
		if err := after(); err != nil {
			return StatusResponse{}, err
		}
	}
	app, err := s.apps.GetByRunID(runID)
	if err != nil {
		return StatusResponse{}, err
	}
	return BuildStatus(app, in, in.IsTerminal()), nil
}
