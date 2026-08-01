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

var anonymizeCandidateFields = map[string]bool{
	"full_name":                        true,
	"current_company":                  true,
	"current_title":                    true,
	"tags":                             true,
	"phone_numbers":                    true,
	"emails":                           true,
	"social_media_links":               true,
	"websites":                         true,
	"addresses":                        true,
	"location":                         true,
	"custom_candidate_fields":          true,
	"source":                           true,
	"recruiter":                        true,
	"coordinator":                      true,
	"attachments":                      true,
	"application_questions":            true,
	"referral_questions":               true,
	"notes":                            true,
	"rejection_notes":                  true,
	"email_addresses":                  true,
	"activity_items":                   true,
	"innotes":                          true,
	"inmails":                          true,
	"rejection_reason":                 true,
	"scorecards_and_interviews":        true,
	"offers":                           true,
	"credited_to":                      true,
	"headline":                         true,
	"all_offer_versions":               true,
	"follow_up_reminders":              true,
	"custom_application_fields":        true,
	"education":                        true,
	"employment":                       true,
	"candidate_stage_data":             true,
	"prospect_owner":                   true,
	"custom_rejection_question_fields": true,
	"touchpoints":                      true,
	"prospect_pool_and_stage":          true,
	"prospect_jobs":                    true,
	"prospect_offices":                 true,
	"prospect_offices_and_departments": true,
	"match_score_reasoning":            true,
	"identity_verification":            true,
	"third_party_integrations":         true,
}

func (h *Hooks) ExecuteWrite(ctx context.Context, action engine.WriteAction, rec connectors.Record, rt *engine.Runtime) (bool, error) {
	if action.Name == "destroy_openings" {
		if err := ValidateWriteRecord(action.Name, rec); err != nil {
			return true, err
		}
		return true, destroyOpenings(ctx, action, rec, rt)
	}
	if err := ValidateWriteRecord(action.Name, rec); err != nil {
		return false, err
	}
	return false, nil
}

func ValidateWriteRecord(actionName string, rec connectors.Record) error {
	switch actionName {
	case "anonymize_candidate":
		return validateAnonymizeCandidateFields(actionName, rec)
	case "replace_hiring_team", "add_hiring_team_members", "remove_hiring_team_member":
		return validateHiringTeamMembers(actionName, rec)
	case "destroy_openings":
		return validateRequiredArray(actionName, rec, "ids", "opening id")
	default:
		return nil
	}
}

func destroyOpenings(ctx context.Context, action engine.WriteAction, rec connectors.Record, rt *engine.Runtime) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if rt == nil || rt.Requester == nil {
		return fmt.Errorf("greenhouse %s requires an initialized runtime", action.Name)
	}
	if err := validateRequiredArray(action.Name, rec, "ids", "opening id"); err != nil {
		return err
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
	sawList := false
	for _, field := range hiringTeamMemberFields {
		value, exists := rec[field]
		if !exists {
			continue
		}
		items, ok := arrayValues(value)
		if !ok {
			return fmt.Errorf("greenhouse %s field %s must be an array", actionName, field)
		}
		if len(items) == 0 {
			return fmt.Errorf("greenhouse %s field %s must contain at least one member id", actionName, field)
		}
		sawList = true
	}
	if !sawList {
		return fmt.Errorf("greenhouse %s requires at least one non-empty hiring-team member list", actionName)
	}
	return nil
}

func validateAnonymizeCandidateFields(actionName string, rec connectors.Record) error {
	items, ok := arrayValues(rec["field_names"])
	if !ok || len(items) == 0 {
		return fmt.Errorf("greenhouse %s requires at least one anonymize field name", actionName)
	}
	if len(items) > len(anonymizeCandidateFields) {
		return fmt.Errorf("greenhouse %s field_names contains too many field names", actionName)
	}
	seen := map[string]bool{}
	for _, item := range items {
		field, ok := item.(string)
		if !ok || strings.TrimSpace(field) == "" {
			return fmt.Errorf("greenhouse %s field_names must contain documented field names", actionName)
		}
		if !anonymizeCandidateFields[field] {
			return fmt.Errorf("greenhouse %s field_names contains an unsupported field name", actionName)
		}
		if seen[field] {
			return fmt.Errorf("greenhouse %s field_names contains a duplicate field name", actionName)
		}
		seen[field] = true
	}
	return nil
}

func validateRequiredArray(actionName string, rec connectors.Record, field, label string) error {
	items, ok := arrayValues(rec[field])
	if !ok || len(items) == 0 {
		return fmt.Errorf("greenhouse %s requires at least one %s", actionName, label)
	}
	return nil
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

func arrayValues(value any) ([]any, bool) {
	if value == nil {
		return nil, false
	}
	rv := reflect.ValueOf(value)
	if rv.Kind() != reflect.Array && rv.Kind() != reflect.Slice {
		return nil, false
	}
	items := make([]any, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		items[i] = rv.Index(i).Interface()
	}
	return items, true
}

func methodOrDefault(method string) string {
	if strings.TrimSpace(method) == "" {
		return "GET"
	}
	return method
}
