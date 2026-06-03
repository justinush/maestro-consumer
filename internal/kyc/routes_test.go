package kyc

import (
	"errors"
	"testing"

	"github.com/justinush/maestro/pkg/workflow"
)

func TestLookupRoute_ok(t *testing.T) {
	t.Parallel()
	key, err := LookupRoute("sg", "main")
	if err != nil {
		t.Fatal(err)
	}
	want := workflow.Key{ID: "kyc.sg.main", Version: "1.0.0"}
	if key != want {
		t.Fatalf("got %#v, want %#v", key, want)
	}
}

func TestLookupRoute_unknown(t *testing.T) {
	t.Parallel()
	_, err := LookupRoute("XX", "MAIN")
	if !errors.Is(err, ErrUnknownRoute) {
		t.Fatalf("want ErrUnknownRoute, got %v", err)
	}
}

func TestLookupRoute_empty(t *testing.T) {
	t.Parallel()
	_, err := LookupRoute("", "MAIN")
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("want ErrInvalid, got %v", err)
	}
}
