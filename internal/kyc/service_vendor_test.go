package kyc_test

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/justinush/maestro-consumer/internal/applicant"
	"github.com/justinush/maestro-consumer/internal/kyc"
	"github.com/justinush/maestro-consumer/internal/vendor"
	"github.com/justinush/maestro/pkg/run"
	"github.com/justinush/maestro/pkg/workflow"
)

func TestVendorWebhook_BridgeHappyPath(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	root := filepath.Join("..", "..")
	wfDir := filepath.Join(root, "workflows")
	reg, err := workflow.LoadDir(wfDir, kyc.WorkflowValidateOptions())
	if err != nil {
		t.Fatal(err)
	}

	vendorStore := vendor.NewMemory()
	actionReg := kyc.NewActionRegistry(vendorStore)
	runs := run.NewMemoryStore()
	apps := applicant.NewMemory()

	svc := kyc.NewService(reg, runs, apps, vendorStore, actionReg)

	start, err := svc.Start(ctx, kyc.StartRequest{Entity: "SG", Flow: "VENDOR"})
	if err != nil {
		t.Fatal(err)
	}
	if start.Status != "awaiting_vendor_callback" {
		t.Fatalf("status: got %q", start.Status)
	}
	if start.Step != "wait-vendor-result" {
		t.Fatalf("step: got %q", start.Step)
	}
	if start.ExternalRef == "" {
		t.Fatal("expected externalRef on start response")
	}

	done, err := svc.HandleVendorWebhook(ctx, kyc.VendorWebhookRequest{
		ExternalRef: start.ExternalRef,
		Status:      "approved",
		EventID:     "evt-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !done.Terminal || done.Status != "approved" {
		t.Fatalf("after webhook: status=%q terminal=%v", done.Status, done.Terminal)
	}

	again, err := svc.HandleVendorWebhook(ctx, kyc.VendorWebhookRequest{
		ExternalRef: start.ExternalRef,
		Status:      "approved",
		EventID:     "evt-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !again.Terminal {
		t.Fatal("duplicate webhook should return same terminal state")
	}
}
