package kyc

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/justinush/maestro/pkg/validate"
	"github.com/justinush/maestro/pkg/workflow"
)

func TestWorkflowLoad_RejectsCustomActionWithoutAllowlist(t *testing.T) {
	t.Parallel()

	wfDir := filepath.Join("..", "..", "workflows")
	_, err := workflow.LoadDir(wfDir, validate.Options{})
	if err == nil {
		t.Fatal("expected error without AllowedActionTypes")
	}
	if !strings.Contains(err.Error(), "schema validation") {
		t.Fatalf("error %q should mention schema validation", err.Error())
	}
}

func TestWorkflowLoad_AcceptsCustomActionWithAllowlist(t *testing.T) {
	t.Parallel()

	wfDir := filepath.Join("..", "..", "workflows")
	reg, err := workflow.LoadDir(wfDir, WorkflowValidateOptions())
	if err != nil {
		t.Fatal(err)
	}
	if !reg.Contains(workflow.Key{ID: "kyc.sg.vendor", Version: "1.0.0"}) {
		t.Fatal("missing kyc.sg.vendor workflow")
	}
}
