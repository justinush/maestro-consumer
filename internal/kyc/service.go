package kyc

import (
	"context"
	"fmt"

	"github.com/justinush/maestro-consumer/internal/applicant"
	"github.com/justinush/maestro-consumer/internal/vendor"
	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/maestro"
	"github.com/justinush/maestro/pkg/run"
	"github.com/justinush/maestro/pkg/workflow"
)

type Service struct {
	reg       *workflow.Registry
	runs      run.Store
	apps      applicant.Repository
	vendor    vendor.Store
	actionReg *engine.Registry
}

func NewService(
	reg *workflow.Registry,
	runs run.Store,
	apps applicant.Repository,
	vendorStore vendor.Store,
	actionReg *engine.Registry,
) *Service {
	return &Service{
		reg:       reg,
		runs:      runs,
		apps:      apps,
		vendor:    vendorStore,
		actionReg: actionReg,
	}
}

func (s *Service) instanceOpts(runID string, initial map[string]any) maestro.InstanceOptions {
	opts := maestro.InstanceOptions{ActionRegistry: s.actionReg}
	if runID != "" {
		opts.RunID = runID
	}
	if initial != nil {
		opts.InitialVariables = initial
	}
	return opts
}

type afterSuccess func() error

func (s *Service) Start(ctx context.Context, req StartRequest) (StatusResponse, error) {
	if err := req.Validate(); err != nil {
		return StatusResponse{}, err
	}
	wfKey, err := LookupRoute(req.Entity, req.Flow)
	if err != nil {
		return StatusResponse{}, err
	}

	applicantID := newID("app")
	runID := newID("run")

	in, err := s.reg.NewInstance(wfKey, s.instanceOpts(runID, map[string]any{
		"applicantId": applicantID,
	}))
	if err != nil {
		return StatusResponse{}, err
	}
	if err := DriveUntilBlocked(in); err != nil {
		return StatusResponse{}, err
	}

	rt, err := s.reg.Lookup(wfKey)
	if err != nil {
		return StatusResponse{}, err
	}
	def := rt.Definition()

	if err := PersistNewRun(ctx, s.runs, in, def); err != nil {
		return StatusResponse{}, err
	}
	if err := s.apps.Create(ctx, applicantID, runID); err != nil {
		return StatusResponse{}, err
	}

	app, err := s.apps.GetByRunID(ctx, runID)
	if err != nil {
		return StatusResponse{}, err
	}
	resp := BuildStatus(app, in, false)
	resp = s.enrichExternalRef(ctx, resp, runID)
	resp = withWorkflowMeta(resp, def)
	resp = withRouteMeta(resp, req.Entity, req.Flow)
	return resp, nil
}

func (s *Service) Get(ctx context.Context, runID string) (StatusResponse, error) {
	app, err := s.apps.GetByRunID(ctx, runID)
	if err != nil {
		return StatusResponse{}, err
	}
	in, def, err := s.restore(ctx, runID)
	if err != nil {
		return StatusResponse{}, err
	}
	resp := BuildStatus(app, in, in.IsTerminal())
	resp = s.enrichExternalRef(ctx, resp, runID)
	return withWorkflowMeta(resp, def), nil
}

func (s *Service) GetEvents(ctx context.Context, runID string) (EventsResponse, error) {
	if _, err := s.apps.GetByRunID(ctx, runID); err != nil {
		return EventsResponse{}, err
	}
	in, _, err := s.restore(ctx, runID)
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
		return s.apps.SaveProfile(ctx, runID, p)
	})
}

func (s *Service) SubmitDocument(ctx context.Context, runID string, d Document) (StatusResponse, error) {
	if err := d.Validate(); err != nil {
		return StatusResponse{}, err
	}

	in, def, err := s.restore(ctx, runID)
	if err != nil {
		return StatusResponse{}, err
	}
	const want = "document-upload"
	if in.CurrentStepID() != want {
		return StatusResponse{}, fmt.Errorf("%w: want %q, at %q", ErrWrongStep, want, in.CurrentStepID())
	}

	app, err := s.apps.GetByRunID(ctx, runID)
	if err != nil {
		return StatusResponse{}, err
	}
	if err := FakeVendorCheckLiveness(app.ApplicantID); err != nil {
		return StatusResponse{}, err
	}

	needsReview := d.Type == "passport"
	return s.submitOnExistingInstance(ctx, runID, want, in, def, map[string]any{
		"documentType": d.Type,
		"documentRef":  d.Ref,
		"review":       map[string]any{"required": needsReview},
	}, func() error {
		return s.apps.AddDocument(ctx, runID, d)
	})
}

func (s *Service) SubmitReview(ctx context.Context, runID string, approved bool) (StatusResponse, error) {
	return s.submitOnStep(ctx, runID, "manual-review", map[string]any{
		"approved": approved,
		"review":   map[string]any{"approved": approved},
	}, nil)
}

func (s *Service) restore(ctx context.Context, runID string) (*engine.Instance, *definition.WorkflowDefinition, error) {
	return RestoreRun(ctx, s.reg, s.runs, runID, s.instanceOpts("", nil))
}

func (s *Service) submitOnStep(
	ctx context.Context,
	runID, wantStep string,
	input map[string]any,
	after afterSuccess,
) (StatusResponse, error) {
	in, def, err := s.restore(ctx, runID)
	if err != nil {
		return StatusResponse{}, err
	}
	return s.submitOnExistingInstance(ctx, runID, wantStep, in, def, input, after)
}

func (s *Service) submitOnExistingInstance(
	ctx context.Context,
	runID, wantStep string,
	in *engine.Instance,
	def *definition.WorkflowDefinition,
	input map[string]any,
	after afterSuccess,
) (StatusResponse, error) {
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
	if err := SaveRun(ctx, s.runs, runID, in, def); err != nil {
		return StatusResponse{}, err
	}
	if after != nil {
		if err := after(); err != nil {
			return StatusResponse{}, err
		}
	}

	app, err := s.apps.GetByRunID(ctx, runID)
	if err != nil {
		return StatusResponse{}, err
	}
	resp := BuildStatus(app, in, in.IsTerminal())
	resp = s.enrichExternalRef(ctx, resp, runID)
	return withWorkflowMeta(resp, def), nil
}
