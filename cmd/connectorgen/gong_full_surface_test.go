package main

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func TestGongFullSurfaceCommandAndOperationCoverage(t *testing.T) {
	api := loadGongJSON[struct {
		Endpoints []struct {
			Method    string         `json:"method"`
			Path      string         `json:"path"`
			CoveredBy map[string]any `json:"covered_by"`
			Operation map[string]any `json:"operation"`
		} `json:"endpoints"`
	}](t, "../../internal/connectors/defs/gong/api_surface.json")
	cli := loadGongJSON[struct {
		GlobalFlags []struct {
			Name string `json:"name"`
			Type string `json:"type"`
		} `json:"global_flags"`
		Commands []struct {
			Path         string   `json:"path"`
			Intent       string   `json:"intent"`
			Availability string   `json:"availability"`
			Stream       string   `json:"stream"`
			Write        string   `json:"write"`
			Operation    string   `json:"operation"`
			OutputPolicy string   `json:"output_policy"`
			Examples     []string `json:"examples"`
			Flags        []struct {
				Name       string `json:"name"`
				MapsTo     string `json:"maps_to"`
				Type       string `json:"type"`
				Format     string `json:"format"`
				AllowEmpty *bool  `json:"allow_empty"`
			} `json:"flags"`
			Constraints []struct {
				Kind    string `json:"kind"`
				Left    string `json:"left"`
				Message string `json:"message"`
			} `json:"constraints"`
			RedactFields []string `json:"redact_fields"`
		} `json:"commands"`
	}](t, "../../internal/connectors/defs/gong/cli_surface.json")
	writes := loadGongJSON[struct {
		Actions []struct {
			Name         string          `json:"name"`
			Kind         string          `json:"kind"`
			Method       string          `json:"method"`
			Path         string          `json:"path"`
			Risk         string          `json:"risk"`
			Confirm      string          `json:"confirm"`
			RedactFields []string        `json:"redact_fields"`
			RecordSchema json.RawMessage `json:"record_schema"`
		} `json:"actions"`
	}](t, "../../internal/connectors/defs/gong/writes.json")
	ops := loadGongJSON[struct {
		Operations []struct {
			ID              string          `json:"id"`
			Kind            string          `json:"kind"`
			Risk            string          `json:"risk"`
			Approval        string          `json:"approval"`
			OutputPolicy    string          `json:"output_policy"`
			MutationClass   string          `json:"mutation_class"`
			SecretSensitive bool            `json:"secret_sensitive"`
			REST            json.RawMessage `json:"rest"`
			SensitivePolicy json.RawMessage `json:"sensitive_policy"`
		} `json:"operations"`
	}](t, "../../internal/connectors/defs/gong/operations.json")

	if got, want := len(writes.Actions), 27; got != want {
		t.Fatalf("write actions = %d, want %d", got, want)
	}
	if got, want := len(ops.Operations), 69; got != want {
		t.Fatalf("operations = %d, want %d", got, want)
	}
	for _, flag := range cli.GlobalFlags {
		if flag.Name == "approve" && flag.Type != "string" {
			t.Fatalf("global --approve type = %q, want string approval token", flag.Type)
		}
	}

	coverage := map[string]int{}
	for _, ep := range api.Endpoints {
		if ep.CoveredBy != nil {
			for key := range ep.CoveredBy {
				coverage[key]++
			}
		}
		if ep.Operation != nil {
			coverage["operation"]++
		}
	}
	wantCoverage := map[string]int{"stream": 12, "direct_read": 30, "write": 27}
	for key, want := range wantCoverage {
		if got := coverage[key]; got != want {
			t.Fatalf("coverage[%s] = %d, want %d (all coverage: %+v)", key, got, want, coverage)
		}
	}

	commandsByPath := map[string]struct {
		intent, availability, stream, write, operation, outputPolicy string
	}{}
	flagsByPath := map[string]map[string]gongCommandFlag{}
	examplesByPath := map[string][]string{}
	constraintsByPath := map[string]map[string]string{}
	redactFieldsByPath := map[string][]string{}
	for _, cmd := range cli.Commands {
		commandsByPath[cmd.Path] = struct {
			intent, availability, stream, write, operation, outputPolicy string
		}{cmd.Intent, cmd.Availability, cmd.Stream, cmd.Write, cmd.Operation, cmd.OutputPolicy}
		flagsByPath[cmd.Path] = map[string]gongCommandFlag{}
		examplesByPath[cmd.Path] = cmd.Examples
		constraintsByPath[cmd.Path] = map[string]string{}
		redactFieldsByPath[cmd.Path] = append([]string(nil), cmd.RedactFields...)
		for _, flag := range cmd.Flags {
			flagsByPath[cmd.Path][flag.Name] = gongCommandFlag{mapsTo: flag.MapsTo, typeName: flag.Type, format: flag.Format, allowEmpty: flag.AllowEmpty}
		}
		for _, constraint := range cmd.Constraints {
			if constraint.Kind == "required" {
				constraintsByPath[cmd.Path][constraint.Left] = constraint.Message
			}
		}
		if cmd.Intent == "direct_read" && cmd.Availability != "implemented" {
			t.Fatalf("direct read command %q availability = %q, want implemented", cmd.Path, cmd.Availability)
		}
	}
	for _, tc := range []struct {
		path, intent, availability, target string
	}{
		{path: "calls list", intent: "etl", availability: "implemented", target: "calls"},
		{path: "workspaces list", intent: "etl", availability: "implemented", target: "workspaces"},
		{path: "calls get", intent: "direct_read", availability: "implemented", target: "json_redacted"},
		{path: "users get", intent: "direct_read", availability: "implemented", target: "json_redacted"},
		{path: "calls create", intent: "reverse_etl", availability: "partial", target: "add_call"},
		{path: "privacy erase-phone", intent: "reverse_etl", availability: "partial", target: "purge_phone_number"},
		{path: "calls extensive", intent: "direct_read", availability: "implemented", target: "json_redacted"},
		{path: "calls transcript", intent: "direct_read", availability: "implemented", target: "json_redacted"},
		{path: "meetings integration-status", intent: "direct_read", availability: "implemented", target: "json_redacted"},
		{path: "crm upload-entities", intent: "reverse_etl", availability: "implemented", target: "upload_crm_entities"},
		{path: "targets list", intent: "direct_read", availability: "implemented", target: "json_redacted"},
		{path: "targets upload-assignments", intent: "reverse_etl", availability: "implemented", target: "upload_target_assignments"},
	} {
		cmd, ok := commandsByPath[tc.path]
		if !ok {
			t.Fatalf("missing cli command %q", tc.path)
		}
		if cmd.intent != tc.intent || cmd.availability != tc.availability {
			t.Fatalf("command %q intent/availability = %s/%s, want %s/%s", tc.path, cmd.intent, cmd.availability, tc.intent, tc.availability)
		}
		switch tc.intent {
		case "etl":
			if cmd.stream != tc.target {
				t.Fatalf("command %q stream = %q, want %q", tc.path, cmd.stream, tc.target)
			}
		case "direct_read":
			if tc.availability == "implemented" && cmd.outputPolicy != tc.target {
				t.Fatalf("command %q output_policy = %q, want %q", tc.path, cmd.outputPolicy, tc.target)
			}
			if tc.availability != "implemented" && cmd.operation != tc.target {
				t.Fatalf("command %q operation = %q, want %q", tc.path, cmd.operation, tc.target)
			}
		case "reverse_etl":
			if (tc.availability == "partial" || tc.availability == "implemented") && cmd.write != tc.target {
				t.Fatalf("command %q write = %q, want %q", tc.path, cmd.write, tc.target)
			}
			if tc.availability == "planned" && cmd.operation != tc.target {
				t.Fatalf("command %q operation = %q, want %q", tc.path, cmd.operation, tc.target)
			}
		}
	}

	writesByName := map[string]struct {
		kind, method, path, risk, confirm string
		redactFields                      []string
		recordSchema                      json.RawMessage
	}{}
	for _, action := range writes.Actions {
		writesByName[action.Name] = struct {
			kind, method, path, risk, confirm string
			redactFields                      []string
			recordSchema                      json.RawMessage
		}{action.Kind, action.Method, action.Path, action.Risk, action.Confirm, append([]string(nil), action.RedactFields...), action.RecordSchema}
	}
	if got := flagsByPath["calls transcript"]["call-id"].mapsTo; got != "body.filter.callIds" {
		t.Fatalf("calls transcript --call-id maps_to = %q, want body.filter.callIds", got)
	}
	if _, exists := flagsByPath["calls transcript"]["body"]; exists {
		t.Fatal("calls transcript must not expose a raw body flag")
	}

	minimumExampleFlags := map[string][]string{
		"calls list":                         {"from", "to"},
		"calls get":                          {"id"},
		"permissions profiles list":          {"workspaceId"},
		"permissions profile get":            {"profileId"},
		"permissions profile-users list":     {"profileId"},
		"crm entity-schema get":              {"integrationId", "objectType"},
		"crm entities get":                   {"integrationId", "objectType", "objectsCrmIds"},
		"crm request-status get":             {"integrationId", "clientRequestId"},
		"users get":                          {"id"},
		"users settings-history":             {"id"},
		"flows list":                         {"flowOwnerEmail"},
		"flows folders list":                 {"flowFolderOwnerEmail"},
		"flows bulk-assignment get":          {"id"},
		"entities get-brief":                 {"workspaceId", "briefName", "crmEntityType", "crmEntityId", "timePeriod"},
		"entities ask":                       {"workspaceId", "crmEntityType", "crmEntityId", "timePeriod", "question"},
		"privacy find-phone":                 {"phoneNumber"},
		"privacy find-email":                 {"emailAddress"},
		"logs list":                          {"logType", "fromDateTime"},
		"coaching list":                      {"workspace-id", "manager-id", "from", "to"},
		"targets list":                       {"workspaceId"},
		"calls extensive":                    {"call-id"},
		"calls users-access get":             {"call-id"},
		"tasks list":                         {"status", "task-action", "task-type", "user-id"},
		"stats interaction":                  {"from-date", "to-date"},
		"stats activity-scorecards":          {"scorecard-id"},
		"stats activity-day-by-day":          {"from-date", "to-date"},
		"stats activity-aggregate":           {"from-date", "to-date"},
		"stats activity-aggregate-by-period": {"from-date", "to-date", "aggregation-period"},
		"calls transcript":                   {"call-id"},
	}
	for path, requiredFlags := range minimumExampleFlags {
		examples := examplesByPath[path]
		if len(examples) == 0 {
			t.Fatalf("command %q has no executable example", path)
		}
		for _, flag := range requiredFlags {
			if !strings.Contains(examples[0], "--"+flag+" ") {
				t.Fatalf("command %q example %q omits required --%s", path, examples[0], flag)
			}
		}
	}

	for _, name := range []string{"add_call", "update_permission_profile", "delete_meeting", "integration_settings", "purge_phone_number", "update_task", "upload_call_media", "upload_crm_entities", "upload_crm_entity_schema", "upload_target_assignments"} {
		if _, ok := writesByName[name]; !ok {
			t.Fatalf("missing write action %q", name)
		}
	}
	if writesByName["delete_meeting"].confirm != "destructive" || writesByName["purge_phone_number"].confirm != "destructive" {
		t.Fatalf("destructive Gong writes must require destructive confirmation: %+v %+v", writesByName["delete_meeting"], writesByName["purge_phone_number"])
	}
	for name, action := range writesByName {
		if strings.HasPrefix(action.path, "/v2/") {
			t.Fatalf("write action %q path = %q, want connector-relative path under base_url /v2", name, action.path)
		}
		if strings.Contains(action.path, "{") && !strings.Contains(action.path, "{{") {
			t.Fatalf("write action %q path = %q, want engine template interpolation with {{ record.<field> }}", name, action.path)
		}
	}
	assertGongWriteRequiredFields(t, "add_calls_users_access", writesByName["add_calls_users_access"].recordSchema, "callAccessList")
	assertGongWriteRequiredFields(t, "delete_calls_users_access", writesByName["delete_calls_users_access"].recordSchema, "callAccessList")
	assertGongWriteRequiredFields(t, "upload_crm_entity_schema", writesByName["upload_crm_entity_schema"].recordSchema, "integrationId", "objectType", "selected_fields")
	if writesByName["upload_crm_entity_schema"].path != "/crm/entity-schema?integrationId={{ record.integrationId }}&objectType={{ record.objectType }}" {
		t.Fatalf("upload_crm_entity_schema path = %q, want required Gong query parameters", writesByName["upload_crm_entity_schema"].path)
	}
	assertGongWriteRequiredFields(t, "upload_target_assignments", writesByName["upload_target_assignments"].recordSchema, "targetId", "workspaceId", "assignments_file_path")
	assertStringSliceContains(t, writesByName["upload_target_assignments"].redactFields, "assignments_file_path")
	assertStringSliceContains(t, writesByName["upload_target_assignments"].redactFields, "assignments_file_content")
	assertStringSliceContains(t, redactFieldsByPath["targets upload-assignments"], "assignments_file_path")
	assertStringSliceContains(t, redactFieldsByPath["targets upload-assignments"], "assignments_file_content")
	if _, err := os.Stat("../../internal/connectors/defs/gong/fixtures/writes/upload_target_assignments.json"); err != nil {
		t.Fatalf("missing upload_target_assignments write fixture: %v", err)
	}
	assertGongFlagFormat(t, flagsByPath, "calls list", "from", "date-time")
	assertGongFlagFormat(t, flagsByPath, "calls list", "to", "date-time")
	assertGongRequiredConstraint(t, constraintsByPath, "calls list", "query.fromDateTime")
	assertGongRequiredConstraint(t, constraintsByPath, "calls list", "query.toDateTime")
	assertGongFlag(t, flagsByPath, "permissions profiles list", "workspaceId", "query.workspaceId", "string", boolPtr(false))
	assertGongRequiredConstraint(t, constraintsByPath, "permissions profiles list", "query.workspaceId")
	assertGongFlag(t, flagsByPath, "permissions profile get", "profileId", "query.profileId", "string", boolPtr(false))
	assertGongRequiredConstraint(t, constraintsByPath, "permissions profile get", "query.profileId")
	assertGongFlag(t, flagsByPath, "permissions profile-users list", "profileId", "query.profileId", "string", boolPtr(false))
	assertGongRequiredConstraint(t, constraintsByPath, "permissions profile-users list", "query.profileId")
	assertGongFlag(t, flagsByPath, "crm entity-schema get", "integrationId", "query.integrationId", "integer", nil)
	assertGongFlag(t, flagsByPath, "crm entity-schema get", "objectType", "query.objectType", "enum", boolPtr(false))
	assertGongRequiredConstraint(t, constraintsByPath, "crm entity-schema get", "query.integrationId")
	assertGongRequiredConstraint(t, constraintsByPath, "crm entity-schema get", "query.objectType")
	assertGongFlag(t, flagsByPath, "crm entities get", "integrationId", "query.integrationId", "integer", nil)
	assertGongFlag(t, flagsByPath, "crm entities get", "objectType", "query.objectType", "enum", boolPtr(false))
	assertGongFlag(t, flagsByPath, "crm entities get", "objectsCrmIds", "query.objectsCrmIds", "string", boolPtr(false))
	assertGongRequiredConstraint(t, constraintsByPath, "crm entities get", "query.integrationId")
	assertGongRequiredConstraint(t, constraintsByPath, "crm entities get", "query.objectType")
	assertGongRequiredConstraint(t, constraintsByPath, "crm entities get", "query.objectsCrmIds")
	assertGongFlag(t, flagsByPath, "logs list", "logType", "query.logType", "enum", boolPtr(false))
	assertGongFlagFormat(t, flagsByPath, "logs list", "fromDateTime", "date-time")
	assertGongRequiredConstraint(t, constraintsByPath, "logs list", "query.logType")
	assertGongRequiredConstraint(t, constraintsByPath, "logs list", "query.fromDateTime")
	assertGongFlag(t, flagsByPath, "entities get-brief", "workspaceId", "query.workspaceId", "integer", nil)
	assertGongFlag(t, flagsByPath, "entities get-brief", "crmEntityType", "query.crmEntityType", "enum", boolPtr(false))
	assertGongFlag(t, flagsByPath, "entities get-brief", "timePeriod", "query.timePeriod", "enum", boolPtr(false))
	assertGongFlagFormat(t, flagsByPath, "entities get-brief", "fromDateTime", "date-time")
	for _, target := range []string{"query.workspaceId", "query.briefName", "query.crmEntityType", "query.crmEntityId", "query.timePeriod"} {
		assertGongRequiredConstraint(t, constraintsByPath, "entities get-brief", target)
	}
	assertGongFlag(t, flagsByPath, "entities ask", "workspaceId", "query.workspaceId", "integer", nil)
	assertGongFlag(t, flagsByPath, "entities ask", "crmEntityType", "query.crmEntityType", "enum", boolPtr(false))
	assertGongFlag(t, flagsByPath, "entities ask", "timePeriod", "query.timePeriod", "enum", boolPtr(false))
	assertGongFlagFormat(t, flagsByPath, "entities ask", "fromDateTime", "date-time")
	for _, target := range []string{"query.workspaceId", "query.crmEntityType", "query.crmEntityId", "query.timePeriod", "query.question"} {
		assertGongRequiredConstraint(t, constraintsByPath, "entities ask", target)
	}
	assertGongFlag(t, flagsByPath, "privacy find-phone", "phoneNumber", "query.phoneNumber", "string", boolPtr(false))
	assertGongRequiredConstraint(t, constraintsByPath, "privacy find-phone", "query.phoneNumber")
	assertGongFlag(t, flagsByPath, "privacy find-email", "emailAddress", "query.emailAddress", "string", boolPtr(false))
	assertGongFlagFormat(t, flagsByPath, "privacy find-email", "emailAddress", "email")
	assertGongRequiredConstraint(t, constraintsByPath, "privacy find-email", "query.emailAddress")
	assertGongFlag(t, flagsByPath, "crm request-status get", "integrationId", "query.integrationId", "integer", nil)
	assertGongRequiredConstraint(t, constraintsByPath, "crm request-status get", "query.integrationId")
	assertGongRequiredConstraint(t, constraintsByPath, "crm request-status get", "query.clientRequestId")
	assertGongFlag(t, flagsByPath, "coaching list", "workspace-id", "query.workspace-id", "integer", nil)
	assertGongFlag(t, flagsByPath, "coaching list", "manager-id", "query.manager-id", "integer", nil)
	assertGongFlagFormat(t, flagsByPath, "coaching list", "from", "date-time")
	assertGongFlagFormat(t, flagsByPath, "coaching list", "to", "date-time")
	for _, target := range []string{"query.workspace-id", "query.manager-id", "query.from", "query.to"} {
		assertGongRequiredConstraint(t, constraintsByPath, "coaching list", target)
	}
	assertGongFlag(t, flagsByPath, "flows list", "flowOwnerEmail", "query.flowOwnerEmail", "string", boolPtr(false))
	assertGongFlagFormat(t, flagsByPath, "flows list", "flowOwnerEmail", "email")
	assertGongRequiredConstraint(t, constraintsByPath, "flows list", "query.flowOwnerEmail")
	assertGongFlag(t, flagsByPath, "flows folders list", "flowFolderOwnerEmail", "query.flowFolderOwnerEmail", "string", boolPtr(false))
	assertGongFlagFormat(t, flagsByPath, "flows folders list", "flowFolderOwnerEmail", "email")
	assertGongRequiredConstraint(t, constraintsByPath, "flows folders list", "query.flowFolderOwnerEmail")
	assertGongFlag(t, flagsByPath, "targets list", "workspaceId", "query.workspaceId", "integer", nil)
	assertGongRequiredConstraint(t, constraintsByPath, "targets list", "query.workspaceId")
	assertGongFlag(t, flagsByPath, "crm upload-entity-schema", "integration-id", "record.integrationId", "string", boolPtr(false))
	assertGongFlag(t, flagsByPath, "crm upload-entity-schema", "object-type", "record.objectType", "enum", nil)

	opsByID := map[string]struct {
		kind, risk, approval, outputPolicy, mutationClass string
		secretSensitive                                   bool
		rest, sensitivePolicy                             json.RawMessage
	}{}
	for _, op := range ops.Operations {
		opsByID[op.ID] = struct {
			kind, risk, approval, outputPolicy, mutationClass string
			secretSensitive                                   bool
			rest, sensitivePolicy                             json.RawMessage
		}{op.Kind, op.Risk, op.Approval, op.OutputPolicy, op.MutationClass, op.SecretSensitive, op.REST, op.SensitivePolicy}
	}
	for _, id := range []string{"gong.calls_extensive", "gong.stats_interaction", "gong.calls_transcript", "gong.calls_media_upload", "gong.crm_upload_entities", "gong.list_target_definitions", "gong.upload_assignments"} {
		if _, ok := opsByID[id]; !ok {
			t.Fatalf("missing operation %q", id)
		}
	}
	if opsByID["gong.calls_extensive"].kind != "rest_read" || !json.Valid(opsByID["gong.calls_extensive"].rest) {
		t.Fatalf("calls extensive operation = %+v, want typed rest_read", opsByID["gong.calls_extensive"])
	}
	if opsByID["gong.calls_media_upload"].mutationClass == "" || len(opsByID["gong.calls_media_upload"].sensitivePolicy) == 0 {
		t.Fatalf("media upload operation missing mutation class or sensitive policy: %+v", opsByID["gong.calls_media_upload"])
	}
}

type gongCommandFlag struct {
	mapsTo     string
	typeName   string
	format     string
	allowEmpty *bool
}

func assertGongFlag(t *testing.T, flagsByPath map[string]map[string]gongCommandFlag, commandPath, flagName, mapsTo, typeName string, allowEmpty *bool) {
	t.Helper()
	flags, ok := flagsByPath[commandPath]
	if !ok {
		t.Fatalf("missing command %q", commandPath)
	}
	flag, ok := flags[flagName]
	if !ok {
		t.Fatalf("command %q missing flag --%s", commandPath, flagName)
	}
	if flag.mapsTo != mapsTo || flag.typeName != typeName {
		t.Fatalf("command %q flag --%s = maps_to:%q type:%q, want maps_to:%q type:%q", commandPath, flagName, flag.mapsTo, flag.typeName, mapsTo, typeName)
	}
	if allowEmpty == nil {
		if flag.allowEmpty != nil {
			t.Fatalf("command %q flag --%s allow_empty = %v, want absent", commandPath, flagName, flag.allowEmpty)
		}
		return
	}
	if flag.allowEmpty == nil || *flag.allowEmpty != *allowEmpty {
		t.Fatalf("command %q flag --%s allow_empty = %v, want %t", commandPath, flagName, flag.allowEmpty, *allowEmpty)
	}
}

func assertGongFlagFormat(t *testing.T, flagsByPath map[string]map[string]gongCommandFlag, commandPath, flagName, format string) {
	t.Helper()
	flag, ok := flagsByPath[commandPath][flagName]
	if !ok {
		t.Fatalf("command %q missing flag --%s", commandPath, flagName)
	}
	if flag.format != format {
		t.Fatalf("command %q flag --%s format = %q, want %q", commandPath, flagName, flag.format, format)
	}
}

func assertGongRequiredConstraint(t *testing.T, constraintsByPath map[string]map[string]string, commandPath, target string) {
	t.Helper()
	if constraintsByPath[commandPath][target] == "" {
		t.Fatalf("command %q missing required constraint for %s", commandPath, target)
	}
}

func assertStringSliceContains(t *testing.T, values []string, want string) {
	t.Helper()
	for _, value := range values {
		if value == want {
			return
		}
	}
	t.Fatalf("values %v missing %q", values, want)
}

func boolPtr(v bool) *bool { return &v }

func assertGongWriteRequiredFields(t *testing.T, action string, raw json.RawMessage, fields ...string) {
	t.Helper()
	var schema struct {
		Required []string `json:"required"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("unmarshal %s record_schema: %v", action, err)
	}
	seen := map[string]bool{}
	for _, field := range schema.Required {
		seen[field] = true
	}
	for _, field := range fields {
		if !seen[field] {
			t.Fatalf("write action %q record_schema.required = %v, missing %q", action, schema.Required, field)
		}
	}
}

func TestGongMetadataEnablesWriteCapability(t *testing.T) {
	metadata := loadGongJSON[struct {
		Capabilities struct {
			Read  bool `json:"read"`
			Write bool `json:"write"`
		} `json:"capabilities"`
	}](t, "../../internal/connectors/defs/gong/metadata.json")
	if !metadata.Capabilities.Read || !metadata.Capabilities.Write {
		t.Fatalf("Gong capabilities read/write = %t/%t, want true/true", metadata.Capabilities.Read, metadata.Capabilities.Write)
	}
}

func loadGongJSON[T any](t *testing.T, path string) T {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
	return out
}
