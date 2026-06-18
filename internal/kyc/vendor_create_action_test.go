package kyc

import (
	"context"
	"testing"

	"github.com/justinush/maestro-consumer/internal/vendor"
	"github.com/justinush/maestro/pkg/engine"
)

func TestVendorCreateSessionRunner_SetsExternalRef(t *testing.T) {
	t.Parallel()

	store := vendor.NewMemory()
	runner := NewVendorCreateSessionRunner(store)

	vars := map[string]any{}
	err := runner.Run(engine.ActionContext{
		RunID:     "run-1",
		StepID:    "create-verification",
		ListName:  "onEnter",
		Variables: vars,
	})
	if err != nil {
		t.Fatal(err)
	}
	ref, ok := vars["externalRef"].(string)
	if !ok || ref == "" {
		t.Fatalf("variables.externalRef: got %#v", vars["externalRef"])
	}

	sess, err := store.LookupByRunID(context.Background(), "run-1")
	if err != nil {
		t.Fatal(err)
	}
	if sess.ExternalRef != ref {
		t.Fatalf("store ref %q != variables %q", sess.ExternalRef, ref)
	}
}
