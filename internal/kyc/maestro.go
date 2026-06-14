package kyc

import (
	"context"
	"fmt"

	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/maestro"
	"github.com/justinush/maestro/pkg/run"
	"github.com/justinush/maestro/pkg/workflow"
)

func PersistNewRun(ctx context.Context, store run.Store, in *engine.Instance, def *definition.WorkflowDefinition) error {
	rec := run.RecordFromInstance(in, def, 0)
	if rec == nil {
		return fmt.Errorf("record from instance: nil")
	}
	if err := store.Create(ctx, rec); err != nil {
		return fmt.Errorf("store create: %w", err)
	}
	return nil
}

func SaveRun(ctx context.Context, store run.Store, runID string, in *engine.Instance, def *definition.WorkflowDefinition) error {
	loaded, err := store.Get(ctx, runID)
	if err != nil {
		return fmt.Errorf("store get: %w", err)
	}
	rec := run.RecordFromInstance(in, def, loaded.Revision)
	if rec == nil {
		return fmt.Errorf("record from instance: nil")
	}
	if err := store.Save(ctx, rec); err != nil {
		return fmt.Errorf("store save: %w", err)
	}
	return nil
}

// RestoreRun loads the run record and restores via workflow.Registry (lookup by rec workflow id/version).
func RestoreRun(
	ctx context.Context,
	reg *workflow.Registry,
	store run.Store,
	runID string,
	opts maestro.InstanceOptions,
) (*engine.Instance, *definition.WorkflowDefinition, error) {
	rec, err := store.Get(ctx, runID)
	if err != nil {
		return nil, nil, fmt.Errorf("store get: %w", err)
	}
	in, err := reg.RestoreInstance(rec, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("restore: %w", err)
	}
	rt, err := reg.Lookup(workflow.Key{ID: rec.WorkflowID, Version: rec.WorkflowVersion})
	if err != nil {
		return nil, nil, fmt.Errorf("lookup: %w", err)
	}
	return in, rt.Definition(), nil
}

func DriveUntilBlocked(in *engine.Instance) error {
	res := in.RunUntilBlocked()
	switch res.Status {
	case engine.RunBlocked, engine.RunCompleted:
		return nil
	case engine.RunFailed:
		return fmt.Errorf("run failed: %w", res.Err)
	default:
		return fmt.Errorf("run: unexpected status %v", res.Status)
	}
}
