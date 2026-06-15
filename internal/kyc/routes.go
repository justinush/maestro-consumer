package kyc

import (
	"fmt"
	"strings"

	"github.com/justinush/maestro/pkg/workflow"
)

type RouteKey struct {
	Entity string
	Flow   string
}

var routes = map[RouteKey]workflow.Key{
	{"SG", "MAIN"}:    {ID: "kyc.sg.main", Version: "1.0.0"},
	{"SG", "REFRESH"}: {ID: "kyc.sg.refresh", Version: "1.0.0"},
	{"SG", "VENDOR"}:  {ID: "kyc.sg.vendor", Version: "1.0.0"},
	{"ID", "MAIN"}:    {ID: "kyc.id.main", Version: "1.0.0"},
}

// LookupRoute resolves entity+flow to a workflow.Key (normalized to upper case).
func LookupRoute(entity, flow string) (workflow.Key, error) {
	rk := RouteKey{
		Entity: strings.ToUpper(strings.TrimSpace(entity)),
		Flow:   strings.ToUpper(strings.TrimSpace(flow)),
	}
	if rk.Entity == "" || rk.Flow == "" {
		return workflow.Key{}, fmt.Errorf("%w: entity and flow are required", ErrInvalid)
	}
	key, ok := routes[rk]
	if !ok {
		return workflow.Key{}, fmt.Errorf("%w: unknown route %s/%s", ErrUnknownRoute, rk.Entity, rk.Flow)
	}
	return key, nil
}
