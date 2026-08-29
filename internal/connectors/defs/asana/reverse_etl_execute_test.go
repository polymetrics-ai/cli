// Package asana holds the Asana bundle's connector-local executable-parity
// test. The bundle itself is pure JSON; this file exists only so the promoted
// read and write-action surfaces are proven by execution rather than inspection.
package asana

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/warehouse"
)

const (
	bundleName               = "asana"
	asanaSourceOpenAPISHA256 = "cb3b90f4e0af56035eab0c648974f625b942a28a7144aa6c2326e38ca0bb3d56"

	// blockedReverseETLReason is the api_surface.json reason this lane exists
	// to retire. Every endpoint still carrying it is an operation the audit
	// counted as supported-but-unreachable.
	blockedReverseETLReason = "planned reverse-ETL write action; blocked until a named action has a bounded record schema, redaction, sanitized fixture, and plan -> preview -> explicit approval -> execute evidence."

	// noRequestContractReason and noBoundedShapeReason are where the rows this
	// lane could NOT promote went. They are counted, not hand-waved: a bounded
	// record schema has to come from the pinned OpenAPI source the bundle
	// already cites, and inventing one would reproduce exactly the
	// implemented-but-unreachable hole this repository is closing.
	noRequestContractReason = "Blocked by default until the pinned Asana OpenAPI source declares a request body for this operation. A typed reverse-ETL action requires a bounded record schema and none can be derived without inventing the payload shape."

	noBoundedShapeReason = "Blocked by default until this operation's request body has a bounded, flag-representable shape in the pinned Asana OpenAPI source."

	// promotedReverseETLOperations plus the two deferred counts must equal the
	// ledger rows that originally carried blockedReverseETLReason, so the
	// shortfall can never be hidden by rewording a reason string.
	promotedReverseETLOperations = 60
	deferredNoRequestContract    = 0
	deferredNoBoundedShape       = 0
	originalBlockedReverseETL    = 60

	// reverseETLBoundEndpoints is the measured number of api_surface.json rows
	// bound to a real write action. It includes the source-complete no-body
	// DELETE/POST promotions, so their source coverage cannot drift back to a
	// blocked ledger row while the command remains executable.
	reverseETLBoundEndpoints = 128

	// blockedDestructiveOperations must remain zero now that every source-locked
	// destructive mutation is bound to a typed action. The generic /batch
	// wrapper is unsupported_api, not an unbound destructive operation.
	blockedDestructiveOperations = 0
)

type asanaOperationSourceLock struct {
	SchemaVersion int    `json:"schema_version"`
	Connector     string `json:"connector"`
	REST          struct {
		SHA256     string                 `json:"sha256"`
		Operations []asanaLockedOperation `json:"operations"`
	} `json:"rest"`
}

type asanaLockedOperation struct {
	ID     string `json:"id"`
	Method string `json:"method"`
	Path   string `json:"path"`
	Source struct {
		RequestBody *struct {
			Required bool `json:"required"`
		} `json:"requestBody"`
	} `json:"source_operation"`
}

func loadBundle(t *testing.T) engine.Bundle {
	t.Helper()
	b, err := engine.Load(os.DirFS(".."), bundleName)
	if err != nil {
		t.Fatalf("load %s bundle: %v", bundleName, err)
	}
	return b
}

func runtimeConfig(baseURL string) connectors.RuntimeConfig {
	return connectors.RuntimeConfig{
		Config:              map[string]string{"base_url": baseURL},
		Secrets:             map[string]string{"access_token": "synthetic-test-token"},
		CredentialRevision:  "asana-fixture-credential-revision",
		ConfigurationDigest: "asana-fixture-configuration-digest",
		WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
	}
}

func loadOperationSourceLock(t *testing.T) asanaOperationSourceLock {
	t.Helper()
	raw, err := os.ReadFile("sources/asana-operation-source-lock.json")
	if err != nil {
		t.Fatalf("read Asana operation source lock: %v", err)
	}
	var lock asanaOperationSourceLock
	if err := json.Unmarshal(raw, &lock); err != nil {
		t.Fatalf("decode Asana operation source lock: %v", err)
	}
	return lock
}

func TestSourceBoundReadControlsMaterializeEveryCompleteReadLane(t *testing.T) {
	bundle := loadBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	commands := map[string]connectors.CommandSurfaceCommand{}
	for _, command := range connector.CommandSurface().Commands {
		commands[command.Path] = command
	}
	for _, test := range []struct {
		path       string
		wantSource string
		wantStream string
	}{
		{path: "custom-fields list", wantSource: "asana.rest.getCustomFieldsForWorkspace", wantStream: "custom_fields"},
		{path: "project-statuses list", wantSource: "asana.rest.getProjectStatusesForProject", wantStream: "project_statuses"},
		{path: "projects list", wantSource: "asana.rest.getProjects", wantStream: "projects"},
		{path: "sections list", wantSource: "asana.rest.getSectionsForProject", wantStream: "sections"},
		{path: "stories list", wantSource: "asana.rest.getStoriesForTask", wantStream: "stories"},
		{path: "tags list", wantSource: "asana.rest.getTags", wantStream: "tags"},
		{path: "tasks list", wantSource: "asana.rest.getTasks", wantStream: "tasks"},
		{path: "team-memberships list", wantSource: "asana.rest.getTeamMemberships", wantStream: "team_memberships"},
		{path: "teams list", wantSource: "asana.rest.getTeamsForWorkspace", wantStream: "teams"},
		{path: "users list", wantSource: "asana.rest.getUsers", wantStream: "users"},
		{path: "workspace-memberships list", wantSource: "asana.rest.getWorkspaceMembershipsForWorkspace", wantStream: "workspace_memberships"},
		{path: "workspaces list", wantSource: "asana.rest.getWorkspaces", wantStream: "workspaces"},
	} {
		t.Run(test.path, func(t *testing.T) {
			command, found := commands[test.path]
			if !found {
				t.Fatalf("command %q is absent", test.path)
			}
			if command.Intent != "direct_read" || command.Availability != "implemented" || command.SourceOperation != test.wantSource || command.Operation != "" || command.Stream != test.wantStream {
				t.Fatalf("command = %+v, want implemented source-bound stream %q", command, test.wantStream)
			}
			if err := commandrunner.Preflight(connector, strings.Fields(command.Path)); err != nil {
				t.Fatalf("Preflight(%q) = %v, want source-bound credential boundary", command.Path, err)
			}
		})
	}

	for _, test := range []struct {
		path      string
		source    string
		operation string
	}{
		{path: "access-requests get-access-requests", source: "asana.rest.getAccessRequests", operation: "get_access_requests"},
		{path: "agents get-agents-for-workspace", source: "asana.rest.getAgentsForWorkspace", operation: "get_agents_for_workspace"},
		{path: "agents get-agent", source: "asana.rest.getAgent", operation: "get_agent"},
		{path: "memberships get-membership", source: "asana.rest.getMembership", operation: "get_membership"},
	} {
		command := commands[test.path]
		if command.Intent != "direct_read" || command.Availability != "implemented" || command.SourceOperation != test.source || command.Operation != test.operation {
			t.Fatalf("%q = %+v, want implemented bounded direct read", test.path, command)
		}
		if err := commandrunner.Preflight(connector, strings.Fields(command.Path)); err != nil {
			t.Fatalf("Preflight(%q) = %v, want source-bound credential boundary", command.Path, err)
		}
	}
}

func TestExecutableCommandsAreAccountedForByPinnedSourceLock(t *testing.T) {
	bundle := loadBundle(t)
	lock := loadOperationSourceLock(t)
	if lock.SchemaVersion != 2 || lock.Connector != bundleName || lock.REST.SHA256 != asanaSourceOpenAPISHA256 {
		t.Fatalf("source lock identity = schema %d connector %q REST sha256 %q, want pinned Asana v2 lock %q",
			lock.SchemaVersion, lock.Connector, lock.REST.SHA256, asanaSourceOpenAPISHA256)
	}
	if len(lock.REST.Operations) != 249 {
		t.Fatalf("source lock REST operations = %d, want 249", len(lock.REST.Operations))
	}

	lockedByID := make(map[string]asanaLockedOperation, len(lock.REST.Operations))
	lockedEndpoints := make(map[string]struct{}, len(lock.REST.Operations))
	for _, operation := range lock.REST.Operations {
		if _, duplicate := lockedByID[operation.ID]; duplicate {
			t.Fatalf("source lock repeats operation id %q", operation.ID)
		}
		lockedByID[operation.ID] = operation
		lockedEndpoints[strings.ToUpper(operation.Method)+" "+operation.Path] = struct{}{}
	}

	directReads := 0
	directWrites := 0
	writeEndpoints := make(map[string]struct{})
	for _, command := range bundle.CLISurface.Commands {
		if command.Availability != "implemented" {
			continue
		}
		switch command.Intent {
		case "direct_read":
			directReads++
			if command.SourceOperation == "" || len(command.APISurface) != 1 {
				t.Fatalf("implemented direct read %q lacks one exact source-lock binding", command.Path)
			}
			locked, ok := lockedByID[command.SourceOperation]
			if !ok {
				t.Fatalf("implemented direct read %q references absent source operation %q", command.Path, command.SourceOperation)
			}
			endpoint := command.APISurface[0]
			if !strings.EqualFold(locked.Method, endpoint.Method) || locked.Path != endpoint.Path {
				t.Fatalf("implemented direct read %q maps %s %s, source lock %q maps %s %s",
					command.Path, endpoint.Method, endpoint.Path, command.SourceOperation, locked.Method, locked.Path)
			}
		case "direct_write":
			directWrites++
			if command.Write == "" || len(command.APISurface) != 1 {
				t.Fatalf("implemented direct write %q lacks one exact action/endpoint binding", command.Path)
			}
			endpoint := strings.ToUpper(command.APISurface[0].Method) + " " + command.APISurface[0].Path
			if _, ok := lockedEndpoints[endpoint]; !ok {
				t.Fatalf("implemented direct write %q endpoint %s is absent from the pinned source lock", command.Path, endpoint)
			}
			writeEndpoints[endpoint] = struct{}{}
		}
	}
	if directReads != 119 {
		t.Fatalf("source-locked implemented direct reads = %d, want 119", directReads)
	}
	if directWrites != 130 || len(writeEndpoints) != 129 {
		t.Fatalf("source-accounted implemented direct writes = %d across %d provider endpoints, want 130 across 129",
			directWrites, len(writeEndpoints))
	}
}

// TestSourceLockedDeferredOperationsAreMaterialized pins the final 34 real
// provider mutations that were previously represented only by planned CLI
// rows. The legacy attachment command aliases are deliberately absent from
// this table: they are additional names for createAttachmentForObject, not
// provider operations that may inflate the immutable 249-operation census.
func TestSourceLockedDeferredOperationsAreMaterialized(t *testing.T) {
	type mutationBinding struct {
		sourceID string
		method   string
		path     string
		command  string
		action   string
	}
	mutations := []mutationBinding{
		{sourceID: "asana.rest.createAccessRequest", method: "POST", path: "/access_requests", command: "access-requests create-access-request", action: "create_access_request"},
		{sourceID: "asana.rest.updateAllocation", method: "PUT", path: "/allocations/{allocation_gid}", command: "allocations update-allocation", action: "update_allocation"},
		{sourceID: "asana.rest.createAllocation", method: "POST", path: "/allocations", command: "allocations create-allocation", action: "create_allocation"},
		{sourceID: "asana.rest.createBudget", method: "POST", path: "/budgets", command: "budgets create-budget", action: "create_budget"},
		{sourceID: "asana.rest.updateBudget", method: "PUT", path: "/budgets/{budget_gid}", command: "budgets update-budget", action: "update_budget"},
		{sourceID: "asana.rest.removeSupportingRelationship", method: "POST", path: "/goals/{goal_gid}/removeSupportingRelationship", command: "goal-relationships remove-supporting-relationship", action: "remove_supporting_relationship"},
		{sourceID: "asana.rest.removeFollowers", method: "POST", path: "/goals/{goal_gid}/removeFollowers", command: "goals remove-followers", action: "remove_followers"},
		{sourceID: "asana.rest.removeCustomFieldSettingForGoal", method: "POST", path: "/goals/{goal_gid}/removeCustomFieldSetting", command: "goals remove-custom-field-setting-for-goal", action: "remove_custom_field_setting_for_goal"},
		{sourceID: "asana.rest.createMembership", method: "POST", path: "/memberships", command: "memberships create-membership", action: "create_membership"},
		{sourceID: "asana.rest.updateMembership", method: "PUT", path: "/memberships/{membership_gid}", command: "memberships update-membership", action: "update_membership"},
		{sourceID: "asana.rest.createOrganizationExport", method: "POST", path: "/organization_exports", command: "organization-exports create-organization-export", action: "create_organization_export"},
		{sourceID: "asana.rest.removeItemForPortfolio", method: "POST", path: "/portfolios/{portfolio_gid}/removeItem", command: "portfolios remove-item-for-portfolio", action: "remove_item_for_portfolio"},
		{sourceID: "asana.rest.removeCustomFieldSettingForPortfolio", method: "POST", path: "/portfolios/{portfolio_gid}/removeCustomFieldSetting", command: "portfolios remove-custom-field-setting-for-portfolio", action: "remove_custom_field_setting_for_portfolio"},
		{sourceID: "asana.rest.removeMembersForPortfolio", method: "POST", path: "/portfolios/{portfolio_gid}/removeMembers", command: "portfolios remove-members-for-portfolio", action: "remove_members_for_portfolio"},
		{sourceID: "asana.rest.removeCustomFieldSettingForProject", method: "POST", path: "/projects/{project_gid}/removeCustomFieldSetting", command: "projects remove-custom-field-setting-for-project", action: "remove_custom_field_setting_for_project"},
		{sourceID: "asana.rest.removeMembersForProject", method: "POST", path: "/projects/{project_gid}/removeMembers", command: "projects remove-members-for-project", action: "remove_members_for_project"},
		{sourceID: "asana.rest.removeFollowersForProject", method: "POST", path: "/projects/{project_gid}/removeFollowers", command: "projects remove-followers-for-project", action: "remove_followers_for_project"},
		{sourceID: "asana.rest.createRate", method: "POST", path: "/rates", command: "rates create-rate", action: "create_rate"},
		{sourceID: "asana.rest.updateRate", method: "PUT", path: "/rates/{rate_gid}", command: "rates update-rate", action: "update_rate"},
		{sourceID: "asana.rest.createRole", method: "POST", path: "/roles", command: "roles create-role", action: "create_role"},
		{sourceID: "asana.rest.updateRole", method: "PUT", path: "/roles/{role_gid}", command: "roles update-role", action: "update_role"},
		{sourceID: "asana.rest.removeDependenciesForTask", method: "POST", path: "/tasks/{task_gid}/removeDependencies", command: "tasks remove-dependencies-for-task", action: "remove_dependencies_for_task"},
		{sourceID: "asana.rest.removeDependentsForTask", method: "POST", path: "/tasks/{task_gid}/removeDependents", command: "tasks remove-dependents-for-task", action: "remove_dependents_for_task"},
		{sourceID: "asana.rest.removeProjectForTask", method: "POST", path: "/tasks/{task_gid}/removeProject", command: "tasks remove-project-for-task", action: "remove_project_for_task"},
		{sourceID: "asana.rest.removeTagForTask", method: "POST", path: "/tasks/{task_gid}/removeTag", command: "tasks remove-tag-for-task", action: "remove_tag_for_task"},
		{sourceID: "asana.rest.removeFollowerForTask", method: "POST", path: "/tasks/{task_gid}/removeFollowers", command: "tasks remove-follower-for-task", action: "remove_follower_for_task"},
		{sourceID: "asana.rest.removeUserForTeam", method: "POST", path: "/teams/{team_gid}/removeUser", command: "teams remove-user-for-team", action: "remove_user_for_team"},
		{sourceID: "asana.rest.updateTimeTrackingCategory", method: "PUT", path: "/time_tracking_categories/{time_tracking_category_gid}", command: "time-tracking-categories update-time-tracking-category", action: "update_time_tracking_category"},
		{sourceID: "asana.rest.createTimeTrackingCategory", method: "POST", path: "/time_tracking_categories", command: "time-tracking-categories create-time-tracking-category", action: "create_time_tracking_category"},
		{sourceID: "asana.rest.updateTimesheetApprovalStatus", method: "PUT", path: "/timesheet_approval_statuses/{timesheet_approval_status_gid}", command: "timesheet-approval-statuses update-timesheet-approval-status", action: "update_timesheet_approval_status"},
		{sourceID: "asana.rest.createTimesheetApprovalStatus", method: "POST", path: "/timesheet_approval_statuses", command: "timesheet-approval-statuses create-timesheet-approval-status", action: "create_timesheet_approval_status"},
		{sourceID: "asana.rest.createWebhook", method: "POST", path: "/webhooks", command: "webhooks create-webhook", action: "create_webhook"},
		{sourceID: "asana.rest.updateWebhook", method: "PUT", path: "/webhooks/{webhook_gid}", command: "webhooks update-webhook", action: "update_webhook"},
		{sourceID: "asana.rest.removeUserForWorkspace", method: "POST", path: "/workspaces/{workspace_gid}/removeUser", command: "workspaces remove-user-for-workspace", action: "remove_user_for_workspace"},
	}

	bundle := loadBundle(t)
	lock := loadOperationSourceLock(t)
	lockedByID := make(map[string]asanaLockedOperation, len(lock.REST.Operations))
	for _, operation := range lock.REST.Operations {
		lockedByID[operation.ID] = operation
	}
	commandsByPath := make(map[string]engine.CLICommand, len(bundle.CLISurface.Commands))
	for _, command := range bundle.CLISurface.Commands {
		commandsByPath[command.Path] = command
	}
	actionsByName := make(map[string]engine.WriteAction, len(bundle.Writes))
	for _, action := range bundle.Writes {
		actionsByName[action.Name] = action
	}
	apiByEndpoint := make(map[string]engine.SurfaceEndpoint, len(bundle.Surface.Endpoints))
	for _, endpoint := range bundle.Surface.Endpoints {
		apiByEndpoint[strings.ToUpper(endpoint.Method)+" "+endpoint.Path] = endpoint
	}

	for _, binding := range mutations {
		t.Run(binding.action, func(t *testing.T) {
			locked, ok := lockedByID[binding.sourceID]
			if !ok || !strings.EqualFold(locked.Method, binding.method) || locked.Path != binding.path {
				t.Fatalf("source lock %q = %+v, want %s %s", binding.sourceID, locked, binding.method, binding.path)
			}
			command, ok := commandsByPath[binding.command]
			if !ok || command.Availability != "implemented" || command.Intent != "direct_write" || command.Write != binding.action || command.SourceOperation != binding.sourceID {
				t.Fatalf("command %q = %+v, want implemented source-bound direct_write %q", binding.command, command, binding.action)
			}
			if len(command.APISurface) != 1 || !strings.EqualFold(command.APISurface[0].Method, binding.method) || command.APISurface[0].Path != binding.path {
				t.Fatalf("command %q api surface = %+v, want %s %s", binding.command, command.APISurface, binding.method, binding.path)
			}
			bodyRequired := binding.sourceID != "asana.rest.createMembership"
			if locked.Source.RequestBody == nil || locked.Source.RequestBody.Required != bodyRequired {
				t.Fatalf("source lock %q request body = %+v, want required=%t", binding.sourceID, locked.Source.RequestBody, bodyRequired)
			}
			var dataFlag *engine.CLIFlag
			for index := range command.Flags {
				if command.Flags[index].MapsTo == "record.data" {
					dataFlag = &command.Flags[index]
					break
				}
			}
			if dataFlag == nil || dataFlag.Type != "json" || dataFlag.Required != bodyRequired {
				t.Fatalf("command %q data flag = %+v, want source-required typed JSON=%t", binding.command, dataFlag, bodyRequired)
			}
			action, ok := actionsByName[binding.action]
			if !ok || !strings.EqualFold(action.Method, binding.method) {
				t.Fatalf("write action %q = %+v, want method %s", binding.action, action, binding.method)
			}
			var recordSchema struct {
				Required []string `json:"required"`
			}
			if err := json.Unmarshal(action.RecordSchema, &recordSchema); err != nil {
				t.Fatalf("decode write action %q record schema: %v", binding.action, err)
			}
			if got := slices.Contains(recordSchema.Required, "data"); got != bodyRequired {
				t.Fatalf("write action %q required data = %t, want source request body required=%t", binding.action, got, bodyRequired)
			}
			if action.BodyType != "json" || !reflect.DeepEqual(action.BodyFields, []string{"data"}) {
				t.Fatalf("write action %q body = type %q fields %v, want source JSON data envelope", binding.action, action.BodyType, action.BodyFields)
			}
			endpoint := apiByEndpoint[binding.method+" "+binding.path]
			if endpoint.CoveredBy == nil || endpoint.CoveredBy.Write != binding.action {
				t.Fatalf("api surface %s %s = %+v, want write %q", binding.method, binding.path, endpoint, binding.action)
			}
			if _, err := os.Stat(filepath.Join("fixtures", "writes", binding.action+".json")); err != nil {
				t.Fatalf("sanitized fixture for %q: %v", binding.action, err)
			}
		})
	}

}

func TestGetMembershipExecutesItsSourceLockedPathBinding(t *testing.T) {
	bundle := loadBundle(t)
	capture := newCaptureServer()
	defer capture.Close()
	// Keep source-bound configuration closed: point the trusted test bundle at
	// the local capture server rather than using the user-overridable base_url
	// field, which the real source-origin preflight correctly refuses.
	bundle.HTTP.URL = capture.URL
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))

	emitted := 0
	result, err := commandrunner.Run(context.Background(), connector, commandrunner.Request{
		Path: []string{"memberships", "get-membership"},
		Flags: map[string][]string{
			"membership-gid": {"fixture-membership"},
			"opt-pretty":     {"true"},
		},
		Config: connectors.RuntimeConfig{
			Secrets:            map[string]string{"access_token": "synthetic-test-token"},
			CredentialRevision: "asana-fixture-credential-revision",
		},
	}, func(connectors.Record) error {
		emitted++
		return nil
	})
	if err != nil {
		t.Fatalf("run getMembership: %v", err)
	}
	if result.DirectRead == nil || emitted != 0 || result.DirectRead.Status != http.StatusOK {
		t.Fatalf("getMembership result = %+v emitted=%d, want one bounded direct-read response returned in metadata", result.DirectRead, emitted)
	}
	request := capture.Last()
	if request == nil || request.Method != http.MethodGet || request.Path != "/memberships/fixture-membership" || !strings.Contains(request.Query, "opt_pretty=true") {
		t.Fatalf("getMembership request = %+v, want GET /memberships/fixture-membership?opt_pretty=true", request)
	}
}

// TestEveryLockedOperationHasOneAccountedCommandLane is the executable form of
// the 249-row source-lock -> API surface -> CLI lane matrix. Provider operations
// are counted once even when one operation deliberately exposes several command
// variants (the two attachment writes and their planned aliases).
func TestEveryLockedOperationHasOneAccountedCommandLane(t *testing.T) {
	bundle := loadBundle(t)
	lock := loadOperationSourceLock(t)

	endpointKey := func(method, path string) string {
		return strings.ToUpper(method) + " " + path
	}
	apiByEndpoint := make(map[string]engine.SurfaceEndpoint, len(bundle.Surface.Endpoints))
	for _, endpoint := range bundle.Surface.Endpoints {
		key := endpointKey(endpoint.Method, endpoint.Path)
		if _, duplicate := apiByEndpoint[key]; duplicate {
			t.Fatalf("api_surface repeats provider endpoint %s", key)
		}
		apiByEndpoint[key] = endpoint
	}
	commandsByEndpoint := make(map[string][]engine.CLICommand)
	for _, command := range bundle.CLISurface.Commands {
		for _, endpoint := range command.APISurface {
			key := endpointKey(endpoint.Method, endpoint.Path)
			commandsByEndpoint[key] = append(commandsByEndpoint[key], command)
		}
	}

	states := map[string]int{}
	commandRows := 0
	for _, operation := range lock.REST.Operations {
		key := endpointKey(operation.Method, operation.Path)
		endpoint, ok := apiByEndpoint[key]
		if !ok {
			t.Errorf("source operation %q has no api_surface row for %s", operation.ID, key)
			continue
		}
		commands := commandsByEndpoint[key]
		if len(commands) == 0 {
			t.Errorf("source operation %q has no command lane for %s", operation.ID, key)
			continue
		}
		commandRows += len(commands)

		hasCommand := func(intent, availability string) bool {
			for _, command := range commands {
				if command.Intent == intent && command.Availability == availability {
					return true
				}
			}
			return false
		}
		switch {
		case endpoint.CoveredBy != nil && endpoint.CoveredBy.DirectRead != "":
			states["implemented_direct_read"]++
			if !hasCommand("direct_read", "implemented") {
				t.Errorf("source operation %q is direct-read covered but has no implemented direct_read command", operation.ID)
			}
		case endpoint.CoveredBy != nil && endpoint.CoveredBy.Stream != "":
			states["implemented_stream_read_etl"]++
			if !hasCommand("direct_read", "implemented") {
				t.Errorf("source operation %q is stream covered but has no implemented direct_read command", operation.ID)
			}
		case endpoint.CoveredBy != nil && (endpoint.CoveredBy.Write != "" || len(endpoint.CoveredBy.Writes) > 0):
			states["implemented_direct_write_reverse_etl"]++
			if !hasCommand("direct_write", "implemented") {
				t.Errorf("source operation %q is write covered but has no implemented direct_write command", operation.ID)
			}
		case hasCommand("direct_read", "planned"):
			states["planned_direct_read"]++
		case hasCommand("direct_write", "planned"):
			states["planned_direct_write"]++
		case hasCommand("direct_write", "unsupported_api"):
			states["unsupported_api"]++
		default:
			t.Errorf("source operation %q has commands but no accounted lane: %+v", operation.ID, commands)
		}
	}

	wantStates := map[string]int{
		"implemented_direct_read":              107,
		"implemented_stream_read_etl":          12,
		"implemented_direct_write_reverse_etl": 129,
		"unsupported_api":                      1,
	}
	if !reflect.DeepEqual(states, wantStates) {
		t.Fatalf("249-row source-lock lane matrix = %v, want %v", states, wantStates)
	}
	if commandRows != 252 {
		t.Fatalf("source-lock matrix command rows = %d, want 252 (249 operations plus three attachment aliases)", commandRows)
	}
}

func TestAsanaSourceTransportClaimsOnlyProvenProjectTaskEventScope(t *testing.T) {
	bundle := loadBundle(t)
	if bundle.SyncTransport == nil || bundle.SyncTransport.Source == nil {
		t.Fatal("Asana bundle has no declared source transport")
	}
	if bundle.SyncTransport.Destination != nil {
		t.Fatalf("Asana destination transport = %+v, want no unproven destination role", bundle.SyncTransport.Destination)
	}
	source := bundle.SyncTransport.Source
	if source.Executor.Family != "declarative_api" || source.Executor.ID != "asana_event_token_source" {
		t.Fatalf("Asana source executor = %+v, want registered event-token executor", source.Executor)
	}
	if !reflect.DeepEqual(source.EligibleStreams, []string{"tasks"}) {
		t.Fatalf("Asana event-token streams = %v, want only the source-proven project task scope", source.EligibleStreams)
	}
	wantModes := []string{"incremental_append", "incremental_upsert", "incremental_dedupe"}
	if got := fmt.Sprint(source.Modes); got != fmt.Sprint(wantModes) {
		t.Fatalf("Asana event-token modes = %v, want %v", source.Modes, wantModes)
	}
	if source.Delivery.Idempotency != "keyed" || source.Delivery.Ordering != "window_coalesced" || source.Delivery.Deletes != "tombstone" {
		t.Fatalf("Asana event-token delivery = %+v, want keyed complete-window coalesce with tombstones", source.Delivery)
	}
	if source.Conformance.Suite != "asana_event_token_source" || source.Conformance.RunID != "source_lock_contract_v1" {
		t.Fatalf("Asana event-token conformance = %+v", source.Conformance)
	}
	for _, action := range bundle.Writes {
		if action.IdempotencyKeyHeader != "" {
			t.Fatalf("Asana write action %q fabricates provider idempotency header %q", action.Name, action.IdempotencyKeyHeader)
		}
	}
}

func TestSourceBoundReadsRejectRawProviderPagingControls(t *testing.T) {
	bundle := loadBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))
	for _, command := range connector.CommandSurface().Commands {
		if command.Intent != "direct_read" || command.SourceOperation == "" {
			continue
		}
		for _, flag := range command.Flags {
			if flag.Name == "offset" || flag.Name == "limit" || flag.MapsTo == "query.offset" || flag.MapsTo == "query.limit" {
				t.Fatalf("source-bound direct read %q exposes raw provider paging flag %+v; use the declared --page/--page-cursor contract", command.Path, flag)
			}
		}
	}

	for _, rawPagingFlag := range []string{"offset", "limit"} {
		t.Run(rawPagingFlag, func(t *testing.T) {
			_, err := commandrunner.Run(context.Background(), connector, commandrunner.Request{
				Path:  []string{"agents", "get-agents-for-workspace"},
				Flags: map[string][]string{"workspace-gid": {"fixture-workspace"}, rawPagingFlag: {"1"}},
			}, func(connectors.Record) error {
				t.Fatal("raw paging control reached direct-read execution")
				return nil
			})
			if err == nil || !strings.Contains(err.Error(), "unknown flag --"+rawPagingFlag) {
				t.Fatalf("raw --%s error = %v, want command-level closed-paging refusal", rawPagingFlag, err)
			}
		})
	}
}

// TestReverseETLLedgerReconciles accounts for every ledger row that originally
// carried the typed reverse-ETL blocked reason. A promoted row is bound to a
// write action; a row that could not be promoted must carry one of the two
// precise, cited reasons and be counted here. Two independent things are
// pinned: the MEASURED number of endpoints bound to a write action, and the
// arithmetic that the three buckets still add up to originalBlockedReverseETL.
// Neither alone is enough - the count alone would not notice a reworded
// blocked reason, and the arithmetic alone would not notice a covered_by block
// being dropped and re-blocked under a third reason string.
func TestReverseETLLedgerReconciles(t *testing.T) {
	b := loadBundle(t)

	var stillGeneric []string
	promoted, noContract, noShape := 0, 0, 0
	for _, ep := range b.Surface.Endpoints {
		if ep.CoveredBy != nil && ep.CoveredBy.Write != "" {
			promoted++
			continue
		}
		if ep.Operation == nil {
			continue
		}
		switch ep.Operation.Reason {
		case blockedReverseETLReason:
			stillGeneric = append(stillGeneric, fmt.Sprintf("%s %s", ep.Method, ep.Path))
		case noRequestContractReason:
			noContract++
		case noBoundedShapeReason:
			noShape++
		}
	}

	if len(stillGeneric) > 0 {
		sort.Strings(stillGeneric)
		t.Fatalf("%d reverse-ETL operations still carry the generic blocked reason instead of being promoted or given a cited blocker:\n  %s",
			len(stillGeneric), strings.Join(stillGeneric, "\n  "))
	}
	if noContract != deferredNoRequestContract {
		t.Errorf("rows blocked on a missing request contract = %d, want %d", noContract, deferredNoRequestContract)
	}
	if noShape != deferredNoBoundedShape {
		t.Errorf("rows blocked on an unbounded request shape = %d, want %d", noShape, deferredNoBoundedShape)
	}
	if promoted != reverseETLBoundEndpoints {
		t.Errorf("api_surface rows bound to a write action = %d, want %d", promoted, reverseETLBoundEndpoints)
	}
	if got := promotedReverseETLOperations + noContract + noShape; got != originalBlockedReverseETL {
		t.Fatalf("promoted(%d) + deferred(%d+%d) = %d, want %d ledger rows accounted for",
			promotedReverseETLOperations, noContract, noShape, got, originalBlockedReverseETL)
	}
}

// TestDestructiveOperationsStayBlocked pins only the remaining source-incomplete
// destructive rows, and requires every bound DELETE to carry a typed
// confirmation challenge through the shared gate.
func TestDestructiveOperationsStayBlocked(t *testing.T) {
	b := loadBundle(t)
	actions := map[string]engine.WriteAction{}
	for _, a := range b.Writes {
		actions[a.Name] = a
	}

	unbound := 0
	for _, ep := range b.Surface.Endpoints {
		if ep.Operation != nil && ep.Operation.Model == "destructive_action" {
			unbound++
		}
		if ep.CoveredBy == nil || ep.CoveredBy.Write == "" {
			continue
		}
		if !strings.EqualFold(ep.Method, http.MethodDelete) {
			continue
		}
		action, ok := actions[ep.CoveredBy.Write]
		if !ok {
			t.Errorf("endpoint %s %s is covered by unknown write %q", ep.Method, ep.Path, ep.CoveredBy.Write)
			continue
		}
		if strings.TrimSpace(action.Confirm) == "" {
			t.Errorf("bound DELETE endpoint %s %s uses write %q with no typed confirmation challenge",
				ep.Method, ep.Path, action.Name)
		}
	}
	if unbound != blockedDestructiveOperations {
		t.Fatalf("source-incomplete destructive_action rows still unbound = %d, want %d",
			unbound, blockedDestructiveOperations)
	}
}

// TestDirectWriteActionsExecute reconciles every declared write action with
// the operator surface, then drives every sanitized JSON/no-body fixture
// through the real commandrunner and engine. The two multipart attachment
// variants have source-derived focused plan coverage below; this broad harness
// does not invent provider responses or JSON fixtures for them.
func TestDirectWriteActionsExecute(t *testing.T) {
	b := loadBundle(t)
	if len(b.Writes) == 0 {
		t.Fatalf("bundle declares no write actions")
	}

	capture := newCaptureServer()
	defer capture.Close()

	cfg := runtimeConfig(capture.URL)
	replay := b
	conn := engine.New(replay, engine.HooksFor(b.Name))
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() = %v", err)
	}
	application, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() = %v", err)
	}
	if _, err := application.AddCredential(context.Background(), app.AddCredentialRequest{
		Name: b.Name + "-fixture", Connector: b.Name,
		Config: map[string]string{"base_url": capture.URL}, Secrets: map[string]string{"access_token": "synthetic-test-token"},
	}); err != nil {
		t.Fatalf("AddCredential(%s) = %v", b.Name, err)
	}

	commandsByWrite := map[string]connectors.CommandSurfaceCommand{}
	for _, cmd := range conn.CommandSurface().Commands {
		if cmd.Intent != "direct_write" || cmd.Write == "" {
			continue
		}
		if prior, ok := commandsByWrite[cmd.Write]; ok {
			t.Fatalf("write %q is referenced by two commands (%q and %q)", cmd.Write, prior.Path, cmd.Path)
		}
		commandsByWrite[cmd.Write] = cmd
	}

	for _, action := range b.Writes {
		t.Run(action.Name, func(t *testing.T) {
			cmd, ok := commandsByWrite[action.Name]
			if !ok {
				t.Fatalf("write action %q has no cli_surface command", action.Name)
			}
			if cmd.Intent != "direct_write" {
				t.Fatalf("write action %q command intent = %q, want direct_write", action.Name, cmd.Intent)
			}
			path := strings.Fields(cmd.Path)
			if action.Name == "upload_attachment_file" || action.Name == "create_external_attachment" {
				if err := commandrunner.Preflight(conn, path); err != nil {
					t.Fatalf("attachment Preflight(%q) = %v, want focused plan test to remain reachable", cmd.Path, err)
				}
				return
			}
			fixture := loadWriteFixture(t, action)

			// An "implemented" command is a promise the runtime keeps: it must
			// resolve, build a record from its own flags, and stage a preview
			// for approval. A command still marked "partial" must be refused,
			// so the availability field can never overstate reachability.
			if cmd.Availability == "implemented" {
				if err := commandrunner.Preflight(conn, path); err != nil {
					t.Fatalf("Preflight(%q) = %v, want nil", cmd.Path, err)
				}
				built, err := commandrunner.BuildWriteCommand(context.Background(), conn, commandrunner.Request{
					Path:    path,
					Flags:   flagsFromRecord(t, cmd, fixture.Record),
					Config:  cfg,
					Preview: true,
				})
				if err != nil {
					t.Fatalf("BuildWriteCommand(%q) = %v, want nil", cmd.Path, err)
				}
				if !built.ApprovalRequired {
					t.Fatalf("command %q does not require approval", cmd.Path)
				}
				if built.Preview == nil || built.Preview.RecordsStaged != 1 {
					t.Fatalf("command %q preview = %+v, want 1 staged record", cmd.Path, built.Preview)
				}
				assertRedactFieldsLoadCompatible(t, action, cmd, fixture.Record)
				assertCommandRecordPreserved(t, built.Record, built.RedactedRecord)
			} else if err := commandrunner.Preflight(conn, path); err == nil {
				t.Fatalf("Preflight(%q) = nil for a command marked %q, want it refused", cmd.Path, cmd.Availability)
			}

			// execute: the real engine issues the request.
			capture.Reset()
			records := []connectors.Record{connectors.Record(fixture.Record)}
			hooks := engine.HooksFor(b.Name)
			written, failed := 0, 0
			if engine.DestructiveTargetForWrite(b.Name, action).RequiresApproval() {
				plan, preview, err := application.PlanConnectorCommand(context.Background(), app.PlanConnectorCommandRequest{
					Name: action.Name, Connector: b.Name, Credential: b.Name + "-fixture", Path: path,
					Flags: flagsFromRecord(t, cmd, fixture.Record), Preview: true,
				})
				if err != nil {
					t.Fatalf("PlanConnectorCommand(%q) = %v", action.Name, err)
				}
				if preview == nil || preview.Digest == "" || plan.ApprovalToken == "" {
					t.Fatalf("PlanConnectorCommand(%q) did not produce genuine preview approval", action.Name)
				}
				run, err := application.RunReverseETL(context.Background(), app.RunReverseETLRequest{
					PlanID: plan.ID, ApprovalToken: plan.ApprovalToken,
					Confirmation:  connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive},
					WithheldFlags: flagsFromRecord(t, cmd, fixture.Record),
				})
				if err != nil {
					t.Fatalf("RunReverseETL(%q) = %v", action.Name, err)
				}
				written, failed = run.RecordsSucceeded, run.RecordsFailed
			} else {
				result, err := engine.Write(context.Background(), replay, connectors.WriteRequest{Action: action.Name, Config: cfg}, records, hooks)
				if err != nil {
					t.Fatalf("engine.Write(%q) = %v, want nil", action.Name, err)
				}
				written, failed = result.RecordsWritten, result.RecordsFailed
			}
			if written != 1 || failed != 0 {
				t.Fatalf("write(%q) result = %d written %d failed, want 1 written 0 failed", action.Name, written, failed)
			}
			got := capture.Last()
			if got == nil {
				t.Fatalf("engine.Write(%q) sent no HTTP request", action.Name)
			}
			if !strings.EqualFold(got.Method, fixture.Expect.Method) {
				t.Fatalf("method = %q, want %q", got.Method, fixture.Expect.Method)
			}
			if got.Path != fixture.Expect.Path {
				t.Fatalf("path = %q, want %q", got.Path, fixture.Expect.Path)
			}
			for key, want := range fixture.Expect.Body {
				val, ok := got.Body[key]
				if !ok {
					t.Fatalf("request body is missing key %q", key)
				}
				if fmt.Sprint(val) != fmt.Sprint(want) {
					t.Fatalf("request body[%q] = %v, want %v", key, val, want)
				}
			}
		})
	}
}

// TestBulkReverseETLActionsPlanPreviewApproveAndRun proves that every Asana
// writes.json action is usable from the saved reverse-ETL route, not only from
// its one-record command. Each subtest materializes one source row in the local
// warehouse, dry-runs the exact provider request, consumes the bounded approval,
// and then observes exactly one request at the capture server.
func TestBulkReverseETLActionsPlanPreviewApproveAndRun(t *testing.T) {
	bundle := loadBundle(t)
	capture := newCaptureServer()
	defer capture.Close()

	ctx := context.Background()
	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("InitProject() = %v", err)
	}
	application, err := app.Open(root)
	if err != nil {
		t.Fatalf("Open() = %v", err)
	}
	credentialName := bundle.Name + "-bulk-fixture"
	if _, err := application.AddCredential(ctx, app.AddCredentialRequest{
		Name: credentialName, Connector: bundle.Name,
		Config: map[string]string{"base_url": capture.URL}, Secrets: map[string]string{"access_token": "synthetic-test-token"},
	}); err != nil {
		t.Fatalf("AddCredential(%s) = %v", bundle.Name, err)
	}

	// Bulk reverse-ETL resolves multipart payloads relative to the app-owned
	// runtime project directory. Keep the bounded fixture there so the real
	// payload-identity and root-confinement checks participate in all three
	// plan/preview/run reads of the record.
	if err := os.WriteFile(filepath.Join(application.ProjectDir(), "asana-bulk-attachment.txt"), []byte("bounded Asana bulk attachment fixture"), 0o600); err != nil {
		t.Fatalf("write bounded bulk attachment fixture: %v", err)
	}

	for _, action := range bundle.Writes {
		t.Run(action.Name, func(t *testing.T) {
			fixture := bulkWriteFixture(t, action)
			table := "asana_bulk_" + action.Name
			tablePath := filepath.Join(application.ProjectDir(), "warehouse", table+warehouse.TableFileExt)
			if err := warehouse.WriteTable(ctx, tablePath, []warehouse.Row{warehouse.Row(fixture.Record)}); err != nil {
				t.Fatalf("materialize source table: %v", err)
			}

			plan, err := application.PlanReverseETL(ctx, app.PlanReverseETLRequest{
				Name:                  "asana_bulk_" + action.Name,
				SourceTable:           table,
				DestinationConnector:  bundle.Name,
				DestinationCredential: credentialName,
				Action:                action.Name,
				Mappings:              identityMappings(fixture.Record),
				Limit:                 1,
			})
			if err != nil {
				t.Fatalf("PlanReverseETL() = %v", err)
			}
			if plan.RecordCount != 1 {
				t.Fatalf("planned records = %d, want 1", plan.RecordCount)
			}
			approvalToken := plan.ApprovalToken

			previewed, preview, err := application.PreviewReversePlan(ctx, plan.ID, nil)
			if err != nil {
				t.Fatalf("PreviewReversePlan() = %v", err)
			}
			if preview.RecordsStaged != 1 || preview.Action != action.Name || preview.Digest == "" {
				t.Fatalf("preview = %+v, want one staged %q action with a digest", preview, action.Name)
			}
			if previewed.ApprovalToken != "" {
				approvalToken = previewed.ApprovalToken
			}
			if approvalToken == "" {
				t.Fatal("plan/preview lifecycle produced no bounded approval token")
			}

			confirmation := connectors.WriteConfirmation{}
			if engine.DestructiveTargetForWrite(bundle.Name, action).RequiresApproval() {
				confirmation = connectors.WriteConfirmation{Kind: connectors.ConfirmationKindDestructive}
			}
			capture.Reset()
			run, err := application.RunReverseETL(ctx, app.RunReverseETLRequest{
				PlanID:        plan.ID,
				ApprovalToken: approvalToken,
				Confirmation:  confirmation,
			})
			if err != nil {
				t.Fatalf("RunReverseETL() = %v", err)
			}
			if run.Status != "completed" || run.RecordsStaged != 1 || run.RecordsSucceeded != 1 || run.RecordsFailed != 0 {
				t.Fatalf("run = status %q, staged/succeeded/failed %d/%d/%d, want completed 1/1/0",
					run.Status, run.RecordsStaged, run.RecordsSucceeded, run.RecordsFailed)
			}
			if capture.Count() != 1 {
				t.Fatalf("provider requests = %d, want exactly 1", capture.Count())
			}
			got := capture.Last()
			if got == nil {
				t.Fatal("bulk reverse ETL sent no provider request")
			}
			if !strings.EqualFold(got.Method, fixture.Expect.Method) {
				t.Fatalf("method = %q, want %q", got.Method, fixture.Expect.Method)
			}
			if got.Path != fixture.Expect.Path {
				t.Fatalf("path = %q, want %q", got.Path, fixture.Expect.Path)
			}
		})
	}
}

func TestAttachmentDirectWritesBuildSourceDerivedOneRecordPlansWithoutProviderIO(t *testing.T) {
	bundle := loadBundle(t)
	capture := newCaptureServer()
	defer capture.Close()
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))

	projectRoot := t.TempDir()
	payloadPath := filepath.Join(projectRoot, "attachment.txt")
	if err := os.WriteFile(payloadPath, []byte("bounded Asana attachment fixture"), 0o600); err != nil {
		t.Fatalf("write bounded attachment fixture: %v", err)
	}
	uploadRecord := connectors.Record{"parent": "fixture-task", "file_path": "attachment.txt"}
	uploadConfig := runtimeConfig(capture.URL)
	uploadConfig.ProjectDir = projectRoot
	approved, err := engine.ApprovedMultipartPayloadSHA256ForWrite(context.Background(), bundle, connectors.WriteRequest{
		Action: "upload_attachment_file", Config: uploadConfig,
	}, []connectors.Record{uploadRecord}, engine.HooksFor(bundle.Name))
	if err != nil {
		t.Fatalf("derive bounded upload payload approval: %v", err)
	}
	if digest := approved[connectors.PayloadApprovalKey(0, "file_path")]; len(digest) != 64 {
		t.Fatalf("approved upload payload digest = %q, want SHA-256", digest)
	}
	uploadConfig.ApprovedPayloadSHA256 = approved

	tests := []struct {
		name          string
		path          []string
		flags         map[string][]string
		config        connectors.RuntimeConfig
		wantAction    string
		wantRecord    connectors.Record
		requiredFlags []string
		wantPayloads  int
	}{
		{
			name: "bounded local file", path: []string{"attachments", "upload-attachment-file"},
			flags:  map[string][]string{"parent": {"fixture-task"}, "file-path": {"attachment.txt"}},
			config: uploadConfig, wantAction: "upload_attachment_file", wantRecord: uploadRecord,
			requiredFlags: []string{"parent", "file-path"}, wantPayloads: 1,
		},
		{
			name: "external URL", path: []string{"attachments", "create-external-attachment"},
			flags: map[string][]string{
				"parent": {"fixture-task"}, "resource-subtype": {"external"},
				"url": {"https://example.test/source"}, "name": {"Source reference"},
			},
			config: runtimeConfig(capture.URL), wantAction: "create_external_attachment",
			wantRecord: connectors.Record{
				"parent": "fixture-task", "resource_subtype": "external",
				"url": "https://example.test/source", "name": "Source reference",
			},
			requiredFlags: []string{"parent", "resource-subtype", "url", "name"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			built, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{
				Path: test.path, Flags: test.flags, Config: test.config, Preview: true,
			})
			if err != nil {
				t.Fatalf("BuildWriteCommand(%q): %v", strings.Join(test.path, " "), err)
			}
			if built.Intent != "direct_write" || built.Write != test.wantAction || !built.ApprovalRequired {
				t.Fatalf("attachment plan identity = %+v, want approval-gated direct_write %q", built, test.wantAction)
			}
			if !reflect.DeepEqual(built.Record, test.wantRecord) || built.Preview == nil || built.Preview.RecordsStaged != 1 {
				t.Fatalf("attachment one-record plan = record %#v preview %+v, want %#v and one staged record",
					built.Record, built.Preview, test.wantRecord)
			}
			for _, required := range test.requiredFlags {
				missing := make(map[string][]string, len(test.flags)-1)
				for name, value := range test.flags {
					if name != required {
						missing[name] = value
					}
				}
				if _, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{
					Path: test.path, Flags: missing, Config: test.config,
				}); err == nil || !strings.Contains(err.Error(), "missing required flag --"+required) {
					t.Fatalf("attachment plan without --%s error = %v, want required-field refusal", required, err)
				}
			}
		})
	}

	if err := app.InitProject(projectRoot); err != nil {
		t.Fatalf("InitProject attachment plan: %v", err)
	}
	application, err := app.Open(projectRoot)
	if err != nil {
		t.Fatalf("Open attachment plan project: %v", err)
	}
	credentialName := bundle.Name + "-attachment-fixture"
	if _, err := application.AddCredential(context.Background(), app.AddCredentialRequest{
		Name: credentialName, Connector: bundle.Name,
		Config: map[string]string{"base_url": capture.URL}, Secrets: map[string]string{"access_token": "synthetic-test-token"},
	}); err != nil {
		t.Fatalf("AddCredential attachment plan: %v", err)
	}
	for _, test := range tests {
		t.Run(test.name+" app plan", func(t *testing.T) {
			plan, preview, err := application.PlanConnectorCommand(context.Background(), app.PlanConnectorCommandRequest{
				Name: "attachment-" + strings.ReplaceAll(test.name, " ", "-"), Connector: bundle.Name,
				Credential: credentialName, Path: test.path, Flags: test.flags, Preview: true,
			})
			if err != nil {
				t.Fatalf("PlanConnectorCommand(%q): %v", strings.Join(test.path, " "), err)
			}
			if plan.ConnectorCommandIntent != "direct_write" || plan.Action != test.wantAction || plan.RecordCount != 1 {
				t.Fatalf("persisted attachment plan = %+v, want one-record direct_write %q", plan, test.wantAction)
			}
			if preview == nil || preview.RecordsStaged != 1 || preview.Digest == "" || plan.ApprovalToken == "" {
				t.Fatalf("persisted attachment preview = %+v token=%t, want one staged approved preview",
					preview, plan.ApprovalToken != "")
			}
			if len(plan.PayloadIdentity) != test.wantPayloads {
				t.Fatalf("attachment payload identities = %+v, want %d", plan.PayloadIdentity, test.wantPayloads)
			}
			if test.wantPayloads == 1 {
				identity := plan.PayloadIdentity[0]
				if identity.Field != "file_path" || identity.SizeBytes != int64(len("bounded Asana attachment fixture")) || len(identity.ContentSHA256) != 64 {
					t.Fatalf("bounded attachment payload identity = %+v", identity)
				}
			}
		})
	}
	if got := capture.Last(); got != nil {
		t.Fatalf("attachment planning performed provider I/O: %+v", got)
	}
}

func TestAttachmentBinaryUploadAliasUsesExplicitProviderUnrestrictedMediaPolicy(t *testing.T) {
	bundle := loadBundle(t)
	connector := engine.New(bundle, engine.HooksFor(bundle.Name))

	var upload engine.WriteAction
	for _, action := range bundle.Writes {
		if action.Name == "upload_attachment_file" {
			upload = action
			break
		}
	}
	if upload.Name == "" || upload.Multipart == nil || len(upload.Multipart.Parts) != 2 {
		t.Fatalf("upload_attachment_file = %+v, want declared two-part multipart action", upload)
	}
	filePart := upload.Multipart.Parts[1]
	if filePart.Name != "file" || filePart.MediaPolicy != connectors.BinaryUploadMediaPolicyProviderUnrestricted || len(filePart.AllowedMediaTypes) != 0 {
		t.Fatalf("attachment file media contract = %+v, want explicit provider-unrestricted policy and no fabricated allow-list", filePart)
	}

	var binaryCommand connectors.CommandSurfaceCommand
	for _, command := range connector.CommandSurface().Commands {
		if command.Path == "attachments binary-upload-attachment" {
			binaryCommand = command
			break
		}
	}
	if binaryCommand.Intent != "binary_upload" || binaryCommand.Availability != "implemented" || binaryCommand.Write != upload.Name {
		t.Fatalf("binary upload command = %+v, want implemented alias bound to upload_attachment_file", binaryCommand)
	}
	if err := commandrunner.Preflight(connector, strings.Fields(binaryCommand.Path)); err != nil {
		t.Fatalf("binary upload preflight: %v", err)
	}
	plan, err := commandrunner.BuildWriteCommand(context.Background(), connector, commandrunner.Request{
		Path: strings.Fields(binaryCommand.Path),
		Flags: map[string][]string{
			"parent":    {"fixture-task"},
			"file-path": {"opaque.fixture"},
		},
	})
	if err != nil {
		t.Fatalf("binary upload plan: %v", err)
	}
	if plan.Intent != "binary_upload" || plan.Write != upload.Name || plan.Record["file_path"] != "opaque.fixture" || !plan.ApprovalRequired {
		t.Fatalf("binary upload plan = %+v, want one approval-bound declared attachment file", plan)
	}
}

func assertRedactFieldsLoadCompatible(t *testing.T, action engine.WriteAction, cmd connectors.CommandSurfaceCommand, record map[string]any) {
	t.Helper()
	for _, field := range cmd.RedactFields {
		if strings.TrimSpace(field) == "" {
			t.Errorf("command %q loaded an empty redact field", cmd.Path)
		}
	}
	for _, field := range action.RedactFields {
		target := strings.TrimPrefix(strings.TrimSpace(field), "record.")
		if _, ok := lookupMapPath(record, strings.Split(target, ".")); !ok {
			t.Errorf("write action %q loaded redact field %q that cannot resolve in the sanitized record", action.Name, field)
		}
	}
}

func assertCommandRecordPreserved(t *testing.T, original, preserved connectors.Record) {
	t.Helper()
	if !reflect.DeepEqual(original, preserved) {
		t.Errorf("command record = %#v, want complete record %#v", preserved, original)
	}
}

// lookupMapPath mirrors the engine's record-path resolver exactly: it descends
// maps only and has no array-index support, so an array-nested declaration
// fails here for the same reason it is inert in the engine.
func lookupMapPath(record map[string]any, parts []string) (any, bool) {
	var current any = record
	for _, part := range parts {
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

type writeFixture struct {
	Record map[string]any `json:"record"`
	Expect struct {
		Method string         `json:"method"`
		Path   string         `json:"path"`
		Body   map[string]any `json:"body,omitempty"`
	} `json:"expect"`
}

func loadWriteFixture(t *testing.T, action engine.WriteAction) writeFixture {
	t.Helper()
	raw, err := os.ReadFile("fixtures/writes/" + action.Name + ".json")
	if os.IsNotExist(err) && action.BodyType == "none" && action.Confirm == "destructive" {
		// A source-complete no-body mutation has no payload fixture to hand
		// author. Its declared path fields are the complete bounded record
		// contract, so derive synthetic safe values and still execute the real
		// plan -> preview -> approval -> HTTP request route below. Other actions
		// retain their explicit sanitized fixtures.
		record := make(map[string]any, len(action.PathFields))
		for _, field := range action.PathFields {
			record[field] = "fixture-" + field
		}
		path, interpolateErr := engine.InterpolatePath(action.Path, engine.Vars{Record: record})
		if interpolateErr != nil {
			t.Fatalf("derive no-body fixture path for %q: %v", action.Name, interpolateErr)
		}
		return writeFixture{Record: record, Expect: struct {
			Method string         `json:"method"`
			Path   string         `json:"path"`
			Body   map[string]any `json:"body,omitempty"`
		}{Method: action.Method, Path: path}}
	}
	if err != nil {
		t.Fatalf("read sanitized fixture for %q: %v", action.Name, err)
	}
	var fx writeFixture
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&fx); err != nil {
		t.Fatalf("parse fixture for %q: %v", action.Name, err)
	}
	return fx
}

func bulkWriteFixture(t *testing.T, action engine.WriteAction) writeFixture {
	t.Helper()
	switch action.Name {
	case "upload_attachment_file":
		return writeFixture{
			Record: map[string]any{"parent": "fixture-task", "file_path": "asana-bulk-attachment.txt"},
			Expect: struct {
				Method string         `json:"method"`
				Path   string         `json:"path"`
				Body   map[string]any `json:"body,omitempty"`
			}{Method: http.MethodPost, Path: "/attachments"},
		}
	case "create_external_attachment":
		return writeFixture{
			Record: map[string]any{
				"parent": "fixture-task", "resource_subtype": "external",
				"url": "https://example.test/source", "name": "Source reference",
			},
			Expect: struct {
				Method string         `json:"method"`
				Path   string         `json:"path"`
				Body   map[string]any `json:"body,omitempty"`
			}{Method: http.MethodPost, Path: "/attachments"},
		}
	default:
		return loadWriteFixture(t, action)
	}
}

func identityMappings(record map[string]any) map[string]string {
	mappings := make(map[string]string, len(record))
	for field := range record {
		mappings[field] = field
	}
	return mappings
}

// flagsFromRecord reverse-maps a fixture record onto the command's declared
// flags, so BuildWriteCommand exercises the same flag plumbing the CLI uses.
func flagsFromRecord(t *testing.T, cmd connectors.CommandSurfaceCommand, record map[string]any) map[string][]string {
	t.Helper()
	flags := map[string][]string{}
	for _, flag := range cmd.Flags {
		target, ok := strings.CutPrefix(flag.MapsTo, "record.")
		if !ok || target == "" {
			continue
		}
		value, found := lookupRecordPath(record, strings.Split(target, "."))
		if !found || value == nil {
			continue
		}
		if flag.Type == "json" {
			encoded, err := json.Marshal(value)
			if err != nil {
				t.Fatalf("marshal fixture field %q for command %q: %v", target, cmd.Path, err)
			}
			flags[flag.Name] = []string{string(encoded)}
			continue
		}
		switch typed := value.(type) {
		case []any:
			var parts []string
			for _, item := range typed {
				if _, isObject := item.(map[string]any); isObject {
					parts = nil
					break
				}
				parts = append(parts, fmt.Sprint(item))
			}
			if len(parts) > 0 {
				flags[flag.Name] = []string{strings.Join(parts, ",")}
			}
		case map[string]any:
			t.Fatalf("command %q maps object field %q through non-JSON flag %q", cmd.Path, target, flag.Name)
		default:
			flags[flag.Name] = []string{fmt.Sprint(typed)}
		}
	}
	return flags
}

// lookupRecordPath walks a dotted record path, descending through array
// indices the same way commandrunner's record-path builder does, so a flag
// mapped to record.users.0.email resolves against the fixture.
func lookupRecordPath(record map[string]any, parts []string) (any, bool) {
	var current any = record
	for _, part := range parts {
		if index, err := strconv.Atoi(part); err == nil {
			items, ok := current.([]any)
			if !ok || index < 0 || index >= len(items) {
				return nil, false
			}
			current = items[index]
			continue
		}
		object, ok := current.(map[string]any)
		if !ok {
			return nil, false
		}
		current, ok = object[part]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

type capturedRequest struct {
	Method string
	Path   string
	Query  string
	Body   map[string]any
}

type captureServer struct {
	*httptest.Server
	mu    sync.Mutex
	last  *capturedRequest
	count int
}

func newCaptureServer() *captureServer {
	c := &captureServer{}
	c.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		captured := &capturedRequest{Method: r.Method, Path: r.URL.Path, Query: r.URL.RawQuery}
		if len(raw) > 0 {
			var body map[string]any
			dec := json.NewDecoder(bytes.NewReader(raw))
			dec.UseNumber()
			if err := dec.Decode(&body); err == nil {
				captured.Body = body
			}
		}
		c.mu.Lock()
		c.last = captured
		c.count++
		c.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":{}}`))
	}))
	return c
}

func (c *captureServer) Reset() {
	c.mu.Lock()
	c.last = nil
	c.count = 0
	c.mu.Unlock()
}

func (c *captureServer) Last() *capturedRequest {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.last
}

func (c *captureServer) Count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.count
}
