// Package greenhouse contains Greenhouse-specific write hooks.
package greenhouse

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

func init() {
	engine.RegisterHooks("greenhouse", func() engine.Hooks { return New() })
}

// Hooks is the Greenhouse hook set.
type Hooks struct{}

// New returns a Greenhouse hook set.
func New() *Hooks { return &Hooks{} }

func (h *Hooks) ConnectorName() string { return "greenhouse" }

var (
	_ engine.Hooks     = (*Hooks)(nil)
	_ engine.WriteHook = (*Hooks)(nil)
)

var hiringTeamMemberFields = []string{
	"hiring_managers",
	"recruiters",
	"coordinators",
	"sourcers",
}

func (h *Hooks) ExecuteWrite(ctx context.Context, action engine.WriteAction, rec connectors.Record, rt *engine.Runtime) (bool, error) {
	switch action.Name {
	case "destroy_openings":
		return true, destroyOpenings(ctx, action, rec, rt)
	case "replace_hiring_team", "add_hiring_team_members", "remove_hiring_team_member":
		return false, validateHiringTeamMembers(action.Name, rec)
	default:
		return false, nil
	}
}

func destroyOpenings(ctx context.Context, action engine.WriteAction, rec connectors.Record, rt *engine.Runtime) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if rt == nil || rt.Requester == nil {
		return fmt.Errorf("greenhouse %s requires an initialized runtime", action.Name)
	}
	if n, ok := arrayLen(rec["ids"]); !ok || n == 0 {
		return fmt.Errorf("greenhouse %s requires at least one opening id", action.Name)
	}
	path, err := engine.InterpolatePath(action.Path, engine.Vars{
		Config:  rt.Config.Config,
		Secrets: rt.Config.Secrets,
		Record:  map[string]any(rec),
	})
	if err != nil {
		return fmt.Errorf("greenhouse %s resolve path: %w", action.Name, err)
	}
	body := bodyFields(rec, action.BodyFields)
	resp, err := rt.Requester.Do(ctx, methodOrDefault(action.Method), path, nil, body)
	if err != nil {
		return err
	}
	var parsed struct {
		NotDeleted []any `json:"not_deleted"`
	}
	if len(resp.Body) > 0 {
		if err := json.Unmarshal(resp.Body, &parsed); err != nil {
			return fmt.Errorf("greenhouse %s decode response: %w", action.Name, err)
		}
	}
	if len(parsed.NotDeleted) > 0 {
		return fmt.Errorf("greenhouse %s response included %d not_deleted opening id(s)", action.Name, len(parsed.NotDeleted))
	}
	return nil
}

func validateHiringTeamMembers(actionName string, rec connectors.Record) error {
	for _, field := range hiringTeamMemberFields {
		if n, ok := arrayLen(rec[field]); ok && n > 0 {
			return nil
		}
	}
	return fmt.Errorf("greenhouse %s requires at least one non-empty hiring-team member list", actionName)
}

func bodyFields(rec connectors.Record, fields []string) map[string]any {
	out := make(map[string]any, len(fields))
	for _, field := range fields {
		if value, ok := rec[field]; ok {
			out[field] = value
		}
	}
	return out
}

func arrayLen(value any) (int, bool) {
	if value == nil {
		return 0, false
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Array && rv.Kind() != reflect.Slice {
		return 0, false
	}
	return rv.Len(), true
}

func methodOrDefault(method string) string {
	if strings.TrimSpace(method) == "" {
		return "GET"
	}
	return method
}
