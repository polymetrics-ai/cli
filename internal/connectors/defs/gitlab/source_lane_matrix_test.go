package gitlab

import (
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

const (
	gitLabSourceLaneMatrixPath = "sources/gitlab-source-lane-matrix.json"
	gitLabSourceLockPath       = "sources/gitlab-operation-source-lock.json"
	gitLabBinaryLockPath       = "sources/gitlab-binary-operation-source-lock.json"
	gitLabCrosswalkPath        = "sources/gitlab-operation-crosswalk.json"
	gitLabDescriptorPath       = "sources/gitlab-operation-descriptor.json"
	gitLabRetainedArtifacts    = "sources/gitlab-retained-artifacts.json"
	gitLabStreamsPath          = "streams.json"
	gitLabSnapshotCommit       = "dc481bac1a8b78d60ac0b4a2b0dfd1a9068ce8db"
	gitLabRuntimeClaim         = "Source-backed mapping only; no runtime execution, certification, or availability proof is claimed."
)

var gitLabLanes = []string{
	"direct_read",
	"direct_write",
	"binary_download",
	"binary_upload",
	"etl",
	"reverse_etl",
	"sync_transport",
}

var gitLabExpectedCounts = map[string]map[string]int{
	"direct_read":     {"mapped_unproven": 763, "not_applicable": 991},
	"direct_write":    {"mapped_unproven": 991, "not_applicable": 763},
	"binary_download": {"mapped_unproven": 1, "not_applicable": 1753},
	"binary_upload":   {"mapped_unproven": 46, "not_applicable": 1708},
	"etl":             {"implemented": 1, "mapped_unproven": 1, "not_applicable": 1752},
	"reverse_etl":     {"mapped_unproven": 991, "not_applicable": 763},
	"sync_transport":  {"missing_foundation": 3, "not_applicable": 1751},
}

var gitLabLegacyStreams = map[string]map[string]string{
	"gitlab.rest.getApiV4Projects": {"stream": "projects", "path": "/projects"},
	"gitlab.rest.getApiV4Groups":   {"stream": "groups", "path": "/groups"},
	"gitlab.rest.getApiV4Users":    {"stream": "users", "path": "/users"},
	"gitlab.rest.getApiV4Issues":   {"stream": "issues", "path": "/issues"},
}

// gitLabMaterializedETLSpecs is a deliberately tiny source-ID allowlist, not
// a method/name heuristic. A row enters it only after retained source facts,
// a field-complete declaration, and a provider-to-DuckDB witness agree.
// Keeping the connector-local path/config mapping here prevents a future
// paginator from becoming executable merely because it looks like a list.
var gitLabMaterializedETLSpecs = map[string]map[string]any{
	"gitlab.rest.getApiV4ProjectsIdMlMlflowApi20MlflowMetricsGetHistory": {
		"source_operation": "getApiV4ProjectsIdMlMlflowApi20MlflowMetricsGetHistory",
		"method":           "GET",
		"path":             "/projects/{id}/ml/mlflow/api/2.0/mlflow/metrics/get-history",
		"operation":        "get_mlflow_metrics_history",
		"stream":           "mlflow_metrics_history",
		"schema":           "schemas/mlflow_metrics_history.json",
		"records_path":     "metrics",
		"mode":             "full_refresh",
		"pagination": map[string]any{
			"type":         "cursor",
			"cursor_param": "page_token",
			"token_path":   "next_page_token",
		},
		"config_bindings": []any{
			map[string]any{"config_key": "project_id", "source_parameter": map[string]any{"in": "path", "name": "id"}, "template": "{{ config.project_id }}"},
			map[string]any{"config_key": "mlflow_run_id", "source_parameter": map[string]any{"in": "query", "name": "run_id"}, "template": "{{ config.mlflow_run_id }}"},
			map[string]any{"config_key": "mlflow_metric_key", "source_parameter": map[string]any{"in": "query", "name": "metric_key"}, "template": "{{ config.mlflow_metric_key }}"},
		},
	},
}

var gitLabPathRestrictionSpecs = []map[string]string{
	{
		"record_id":              "gitlab.required-path-parameter.postApiV4GroupsIdDashEpicsEpicIidIssuesIssueId",
		"source_id":              "gitlab.rest.postApiV4GroupsIdDashEpicsEpicIidIssuesIssueId",
		"descriptor_location":    "paths[\"/api/v4/groups/{id}/(-/)epics/{epic_iid}/issues/{epic_issue_id}\"].post.request.path",
		"missing_placeholder":    "epic_issue_id",
		"source_location_suffix": ".post",
	},
	{
		"record_id":              "gitlab.required-path-parameter.getApiV4JobsIdSbomScansSbomScanId",
		"source_id":              "gitlab.rest.getApiV4JobsIdSbomScansSbomScanId",
		"descriptor_location":    "paths[\"/api/v4/jobs/{id}/sbom_scans/{sbom_digest}\"].get.request.path",
		"missing_placeholder":    "sbom_digest",
		"source_location_suffix": ".get",
	},
}

func TestGitLabSourceLaneMatrixRetainsEveryLockedOperationAndLane(t *testing.T) {
	matrix := loadGitLabObject(t, gitLabSourceLaneMatrixPath)
	lock := loadGitLabObject(t, gitLabSourceLockPath)
	binaryLock := loadGitLabObject(t, gitLabBinaryLockPath)
	crosswalk := loadGitLabObject(t, gitLabCrosswalkPath)
	descriptor := loadGitLabObject(t, gitLabDescriptorPath)
	retainedArtifacts := loadGitLabObject(t, gitLabRetainedArtifacts)
	streams := loadGitLabObject(t, gitLabStreamsPath)
	if err := validateGitLabSourceLaneMatrix(matrix, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams); err != nil {
		t.Fatalf("validate GitLab source lane matrix: %v", err)
	}

	t.Run("rejects hidden source row", func(t *testing.T) {
		broken := cloneGitLabObject(t, matrix)
		rows := mustGitLabArray(t, broken["source_operations"])
		broken["source_operations"] = rows[1:]
		if err := validateGitLabSourceLaneMatrix(broken, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams); err == nil || !strings.Contains(err.Error(), "source rows matrix=") {
			t.Fatalf("hidden-row validation error = %v, want source rows matrix", err)
		}
	})

	t.Run("rejects missing lane cell", func(t *testing.T) {
		broken := cloneGitLabObject(t, matrix)
		rows := mustGitLabArray(t, broken["source_operations"])
		delete(mustGitLabObject(t, rows[0])["lanes"].(map[string]any), "etl")
		if err := validateGitLabSourceLaneMatrix(broken, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams); err == nil || !strings.Contains(err.Error(), "lane cells") {
			t.Fatalf("missing-cell validation error = %v, want lane cells", err)
		}
	})

	t.Run("rejects invalid mutation disposition", func(t *testing.T) {
		broken := cloneGitLabObject(t, matrix)
		row := gitLabMatrixRow(t, broken, "gitlab.rest.deleteApiV4AdminActiveContextDeadQueue")
		mustGitLabObject(t, mustGitLabObject(t, row["lanes"])["direct_write"])["disposition"] = "not_applicable"
		if err := validateGitLabSourceLaneMatrix(broken, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams); err == nil || !strings.Contains(err.Error(), "lane cells") {
			t.Fatalf("mutation-disposition validation error = %v, want lane cells", err)
		}
	})

	t.Run("rejects crosswalk boundary drop", func(t *testing.T) {
		broken := cloneGitLabObject(t, matrix)
		boundary := mustGitLabObject(t, broken["source_boundary_reconciliation"])
		records := mustGitLabArray(t, boundary["crosswalk_only_source_identities"])
		boundary["crosswalk_only_source_identities"] = records[1:]
		if err := validateGitLabSourceLaneMatrix(broken, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams); err == nil || !strings.Contains(err.Error(), "crosswalk boundary") {
			t.Fatalf("boundary validation error = %v, want crosswalk boundary", err)
		}
	})

	t.Run("rejects executable promotion", func(t *testing.T) {
		broken := cloneGitLabObject(t, matrix)
		row := gitLabMatrixRow(t, broken, "gitlab.rest.getApiV4Projects")
		mustGitLabObject(t, mustGitLabObject(t, row["lanes"])["direct_read"])["disposition"] = "implemented"
		if err := validateGitLabSourceLaneMatrix(broken, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams); err == nil || !strings.Contains(err.Error(), "lane cells") {
			t.Fatalf("executable-promotion validation error = %v, want lane cells", err)
		}
	})

	t.Run("retains source-semantic HEAD and POST direct reads", func(t *testing.T) {
		var headRows, postRows int
		for _, raw := range mustGitLabArray(t, matrix["source_operations"]) {
			row := mustGitLabObject(t, raw)
			semantics := gitLabFactsOperationSemantics(objectAt(row, "source_facts"))
			state := stringAt(semantics, "state")
			if state != "source_semantic_head_read" && state != "source_semantic_post_read" {
				continue
			}
			cell := objectAt(objectAt(row, "lanes"), "direct_read")
			if stringAt(cell, "applicability") != "source_candidate" || stringAt(cell, "disposition") != "mapped_unproven" {
				t.Fatalf("semantic read %s direct-read cell=%#v, want source-backed mapped_unproven", stringAt(row, "source_id"), cell)
			}
			backlink := objectAt(objectAt(cell, "mapping"), "definition_backlink")
			if stringAt(backlink, "kind") != "source_lock" || stringAt(backlink, "path") != gitLabSourceLockPath || stringAt(backlink, "source_id") != stringAt(row, "source_id") {
				t.Fatalf("semantic read %s source backlink=%#v, want exact retained source-lock binding", stringAt(row, "source_id"), backlink)
			}
			classification := stringAt(objectAt(objectAt(cell, "mapping"), "source_fact"), "classification")
			switch state {
			case "source_semantic_head_read":
				headRows++
				if classification != "source_semantic_head_read_candidate" {
					t.Fatalf("HEAD semantic read %s classification=%q", stringAt(row, "source_id"), classification)
				}
			case "source_semantic_post_read":
				postRows++
				if classification != "source_semantic_post_read_candidate" && classification != "source_semantic_post_read_requires_untyped_json_body" {
					t.Fatalf("POST semantic read %s classification=%q, want ordinary or untyped-JSON-body source-backed direct-read mapping", stringAt(row, "source_id"), classification)
				}
			}
		}
		if headRows == 0 || postRows == 0 {
			t.Fatalf("semantic read coverage head=%d post=%d, want both source-backed kinds", headRows, postRows)
		}
	})

	t.Run("retains Conan JSON-body direct reads as mapped-unproven", func(t *testing.T) {
		const bodyRequirement = "the request must include a json object with the name and size of the individual files."
		want := map[string]struct{}{
			"gitlab.rest.postApiV4PackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelPackagesConanPackageReferenceUploadUrls":           {},
			"gitlab.rest.postApiV4PackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelUploadUrls":                                        {},
			"gitlab.rest.postApiV4ProjectsIdPackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelPackagesConanPackageReferenceUploadUrls": {},
			"gitlab.rest.postApiV4ProjectsIdPackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelUploadUrls":                              {},
		}
		found := make(map[string]struct{}, len(want))
		for _, raw := range mustMapSlice(objectAt(lock, "rest")["operations"]) {
			source := objectAt(raw, "source_operation")
			if !strings.Contains(strings.ToLower(stringAt(source, "description")), bodyRequirement) {
				continue
			}
			sourceID := "gitlab.rest." + stringAt(raw, "id")
			if _, expected := want[sourceID]; !expected {
				t.Fatalf("unexpected retained JSON-body-without-schema source %q", sourceID)
			}
			if _, hasBody := source["requestBody"]; hasBody || len(sourceRequestMediaTypes(raw)) != 0 {
				t.Fatalf("Conan source %q unexpectedly has a typed request body/media contract", sourceID)
			}
			row := gitLabMatrixRow(t, matrix, sourceID)
			directRead := objectAt(objectAt(row, "lanes"), "direct_read")
			if stringAt(directRead, "applicability") != "source_candidate" || stringAt(directRead, "disposition") != "mapped_unproven" {
				t.Fatalf("Conan source %q direct-read cell=%#v, want source-backed mapped-unproven", sourceID, directRead)
			}
			if stringAt(objectAt(objectAt(directRead, "mapping"), "source_fact"), "classification") != "source_semantic_post_read_requires_untyped_json_body" {
				t.Fatalf("Conan source %q classification=%#v, want untyped JSON-body direct-read mapping", sourceID, objectAt(objectAt(directRead, "mapping"), "source_fact"))
			}
			reason := stringAt(directRead, "reason")
			if !strings.Contains(reason, "requires a JSON object with the name and size") || !strings.Contains(reason, "no requestBody media type or closed schema") || !strings.Contains(reason, "bodyless or raw-body") {
				t.Fatalf("Conan source %q reason=%q, want exact typed-body/no-shortcut boundary", sourceID, reason)
			}
			cellBytes, err := json.Marshal(directRead)
			if err != nil {
				t.Fatalf("marshal Conan source %q direct-read cell: %v", sourceID, err)
			}
			if strings.Contains(string(cellBytes), "no_request_body") || strings.Contains(string(cellBytes), "raw_body") {
				t.Fatalf("Conan source %q direct-read cell=%s, must not claim bodyless/raw-body execution", sourceID, cellBytes)
			}
			found[sourceID] = struct{}{}
		}
		if !reflect.DeepEqual(found, want) {
			t.Fatalf("retained Conan JSON-body source IDs=%v, want %v", sortedGitLabSet(found), sortedGitLabSet(want))
		}
		conanRawOperationIDs := gitLabRawOperationIDsForRows(t, matrix, want)
		cli := loadGitLabObject(t, "cli_surface.json")
		operations := loadGitLabObject(t, "operations.json")
		if err := gitLabNoImplementedDirectReadArtifactError(cli, operations, conanRawOperationIDs); err != nil {
			t.Fatalf("baseline Conan direct-read artifact guard: %v", err)
		}
		if got := gitLabMatchingCLIArtifactCount(cli, conanRawOperationIDs); got == 0 {
			t.Fatal("Conan artifact guard selected no legacy CLI rows; canonical matrix IDs must normalize to raw source operation IDs")
		}

		broken := cloneGitLabObject(t, matrix)
		row := gitLabMatrixRow(t, broken, "gitlab.rest.postApiV4PackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelUploadUrls")
		cell := objectAt(objectAt(row, "lanes"), "direct_read")
		cell["disposition"] = "implemented"
		if err := validateGitLabSourceLaneMatrix(broken, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams); err == nil || !strings.Contains(err.Error(), "lane cells") {
			t.Fatalf("Conan implemented promotion error = %v, want deterministic lane-cell rejection", err)
		}

		for _, forbidden := range []string{"no_request_body", "raw_body"} {
			broken := cloneGitLabObject(t, matrix)
			row := gitLabMatrixRow(t, broken, "gitlab.rest.postApiV4PackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelUploadUrls")
			mapping := objectAt(objectAt(objectAt(row, "lanes"), "direct_read"), "mapping")
			mapping[forbidden] = true
			if err := validateGitLabSourceLaneMatrix(broken, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams); err == nil || !strings.Contains(err.Error(), "lane cells") {
				t.Fatalf("Conan %s promotion error = %v, want deterministic lane-cell rejection", forbidden, err)
			}
		}

		brokenCLI := cloneGitLabObject(t, cli)
		commands := mustGitLabArray(t, brokenCLI["commands"])
		brokenCLI["commands"] = append(commands, map[string]any{
			"source_operation": "postApiV4PackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelUploadUrls",
			"intent":           "direct_read",
			"availability":     "implemented",
		})
		if err := gitLabNoImplementedDirectReadArtifactError(brokenCLI, operations, conanRawOperationIDs); err == nil || !strings.Contains(err.Error(), "implemented CLI direct-read") {
			t.Fatalf("Conan CLI direct-read binding error = %v, want deterministic direct-read binding rejection", err)
		}

		brokenOperations := cloneGitLabObject(t, operations)
		declared := mustGitLabArray(t, brokenOperations["operations"])
		brokenOperations["operations"] = append(declared, map[string]any{
			"kind": "rest_read",
			"source_operation": map[string]any{
				"id": "postApiV4PackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelUploadUrls",
			},
		})
		if err := gitLabNoImplementedDirectReadArtifactError(cli, brokenOperations, conanRawOperationIDs); err == nil || !strings.Contains(err.Error(), "declared rest-read") {
			t.Fatalf("Conan rest-read binding error = %v, want deterministic direct-read binding rejection", err)
		}
	})

	t.Run("keeps status-only semantic POST reads unmapped from direct execution", func(t *testing.T) {
		statusSourceIDs := map[string]struct{}{
			"gitlab.rest.postApiV4AiThirdPartyAgentsDirectAccess":   {},
			"gitlab.rest.postApiV4CodeSuggestionsConnectionDetails": {},
			"gitlab.rest.postApiV4GeoNodeProxyIdGraphql":            {},
			"gitlab.rest.postApiV4IntegrationsSlackOptions":         {},
		}
		for sourceID := range statusSourceIDs {
			row := gitLabMatrixRow(t, matrix, sourceID)
			facts := objectAt(row, "source_facts")
			if stringAt(gitLabFactsOperationSemantics(facts), "state") != "source_semantic_post_read" || len(stringSlice(facts["success_response_media_types"])) != 0 {
				t.Fatalf("status source %q facts=%#v, want semantic POST with no declared response media", sourceID, facts)
			}
			cell := objectAt(objectAt(row, "lanes"), "direct_read")
			if stringAt(cell, "disposition") != "mapped_unproven" {
				t.Fatalf("status source %q direct-read=%#v, want mapped-unproven", sourceID, cell)
			}
		}

		statusRawOperationIDs := gitLabRawOperationIDsForRows(t, matrix, statusSourceIDs)
		cli := loadGitLabObject(t, "cli_surface.json")
		operations := loadGitLabObject(t, "operations.json")
		if err := gitLabNoImplementedDirectReadArtifactError(cli, operations, statusRawOperationIDs); err != nil {
			t.Fatalf("status-only direct-read artifact guard: %v", err)
		}
		if got := gitLabMatchingCLIArtifactCount(cli, statusRawOperationIDs); got == 0 {
			t.Fatal("status-only artifact guard selected no CLI rows; canonical matrix IDs must normalize to raw source operation IDs")
		}

		broken := cloneGitLabObject(t, matrix)
		row := gitLabMatrixRow(t, broken, "gitlab.rest.postApiV4CodeSuggestionsConnectionDetails")
		mapping := objectAt(objectAt(objectAt(row, "lanes"), "direct_read"), "mapping")
		mapping["output_policy"] = "json_redacted"
		if err := validateGitLabSourceLaneMatrix(broken, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams); err == nil || !strings.Contains(err.Error(), "lane cells") {
			t.Fatalf("status-only JSON output promotion error = %v, want deterministic lane-cell rejection", err)
		}
	})

	t.Run("rejects a mutation POST misclassified as a direct read", func(t *testing.T) {
		broken := cloneGitLabObject(t, matrix)
		var selected map[string]any
		for _, raw := range mustGitLabArray(t, broken["source_operations"]) {
			row := mustGitLabObject(t, raw)
			facts := objectAt(row, "source_facts")
			if stringAt(facts, "method") == "POST" && stringAt(gitLabFactsOperationSemantics(facts), "state") != "source_semantic_post_read" {
				selected = row
				break
			}
		}
		if selected == nil {
			t.Fatal("no retained mutation POST source row")
		}
		lanes := objectAt(selected, "lanes")
		lanes["direct_read"] = gitLabMappedCell("incorrect read promotion", "source_semantic_post_read_candidate", map[string]any{"kind": "source_lock", "path": gitLabSourceLockPath, "source_id": stringAt(selected, "source_id")})
		if err := validateGitLabSourceLaneMatrix(broken, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams); err == nil || !strings.Contains(err.Error(), "lane cells") {
			t.Fatalf("mutation POST direct-read promotion error = %v, want lane cells", err)
		}
	})

	t.Run("rejects a semantic POST read misclassified as a mutation", func(t *testing.T) {
		broken := cloneGitLabObject(t, matrix)
		var selected map[string]any
		for _, raw := range mustGitLabArray(t, broken["source_operations"]) {
			row := mustGitLabObject(t, raw)
			if stringAt(gitLabFactsOperationSemantics(objectAt(row, "source_facts")), "state") == "source_semantic_post_read" {
				selected = row
				break
			}
		}
		if selected == nil {
			t.Fatal("no source-semantic POST read retained")
		}
		lanes := objectAt(selected, "lanes")
		lanes["direct_write"] = gitLabMappedCell("incorrect mutation promotion", "mutation_verb_candidate", map[string]any{"kind": "source_lock", "path": gitLabSourceLockPath, "source_id": stringAt(selected, "source_id")})
		lanes["reverse_etl"] = cloneGitLabMap(objectAt(lanes, "direct_write"))
		if err := validateGitLabSourceLaneMatrix(broken, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams); err == nil || !strings.Contains(err.Error(), "lane cells") {
			t.Fatalf("semantic POST mutation promotion error = %v, want lane cells", err)
		}
	})

	t.Run("requires a retained request-to-response continuation pair for ETL", func(t *testing.T) {
		var paired, unpaired, pagePerPageWithoutContinuation int
		for _, raw := range mustGitLabArray(t, matrix["source_operations"]) {
			row := mustGitLabObject(t, raw)
			pagination := objectAt(objectAt(row, "source_facts"), "pagination")
			state := gitLabPaginationState(pagination)
			etl := objectAt(objectAt(row, "lanes"), "etl")
			switch state {
			case "request_response_continuation_candidate":
				paired++
				if _, materialized := gitLabMaterializedETLSpecs[stringAt(row, "source_id")]; materialized {
					if stringAt(etl, "disposition") != "implemented" {
						t.Fatalf("materialized continuation %s ETL=%#v, want implemented", stringAt(row, "source_id"), etl)
					}
				} else if stringAt(etl, "disposition") != "mapped_unproven" {
					t.Fatalf("unmaterialized continuation %s ETL=%#v, want mapped_unproven", stringAt(row, "source_id"), etl)
				}
			case "request_controls_without_response_continuation":
				unpaired++
				if stringAt(etl, "disposition") != "not_applicable" || !strings.Contains(stringAt(etl, "reason"), "no retained response continuation") {
					t.Fatalf("unpaired continuation %s ETL=%#v, want explicit non-candidate", stringAt(row, "source_id"), etl)
				}
				controls := stringSlice(pagination["request_controls"])
				if containsGitLabString(controls, "page") && containsGitLabString(controls, "per_page") {
					pagePerPageWithoutContinuation++
				}
			}
		}
		if paired == 0 || unpaired == 0 || pagePerPageWithoutContinuation == 0 {
			t.Fatalf("continuation coverage paired=%d unpaired=%d page_per_page_without_response=%d, want all source states", paired, unpaired, pagePerPageWithoutContinuation)
		}
	})
}

// TestGitLabMLflowMetricsHistoryETLDeclarationIsSourceBound is deliberately
// source-first: the connector's project_id spelling is an explicit declared
// config-to-path binding for the retained {id} parameter, never an inferred
// alias. It stays in the GitLab definition test package because it proves one
// connector declaration against its frozen source facts, not a new runtime
// dialect.
func TestGitLabMLflowMetricsHistoryETLDeclarationIsSourceBound(t *testing.T) {
	const (
		sourceID       = "gitlab.rest.getApiV4ProjectsIdMlMlflowApi20MlflowMetricsGetHistory"
		rawOperationID = "getApiV4ProjectsIdMlMlflowApi20MlflowMetricsGetHistory"
		streamName     = "mlflow_metrics_history"
		lockedPath     = "/api/v4/projects/{id}/ml/mlflow/api/2.0/mlflow/metrics/get-history"
		mappingPath    = "/projects/{id}/ml/mlflow/api/2.0/mlflow/metrics/get-history"
	)

	matrix := loadGitLabObject(t, gitLabSourceLaneMatrixPath)
	lock := loadGitLabObject(t, gitLabSourceLockPath)
	row := gitLabMatrixRow(t, matrix, sourceID)
	facts := objectAt(row, "source_facts")
	if stringAt(facts, "operation_id") != rawOperationID || stringAt(facts, "method") != "GET" || stringAt(facts, "path") != lockedPath || stringAt(facts, "mapping_path") != mappingPath {
		t.Fatalf("MLflow source facts = %#v, want exact retained identity/method/paths", facts)
	}
	continuation := objectAt(objectAt(facts, "pagination"), "continuation")
	if stringAt(continuation, "request") != "page_token" || stringAt(continuation, "response") != "next_page_token" {
		t.Fatalf("MLflow source continuation = %#v, want page_token -> next_page_token", continuation)
	}

	var locked map[string]any
	for _, raw := range mustMapSlice(objectAt(lock, "rest")["operations"]) {
		if stringAt(raw, "id") == rawOperationID {
			locked = raw
			break
		}
	}
	if locked == nil {
		t.Fatalf("retained GitLab source lock omits %q", rawOperationID)
	}
	parameters := mustMapSlice(objectAt(locked, "source_operation")["parameters"])
	for _, want := range []struct {
		name     string
		location string
		required bool
	}{
		{name: "id", location: "path", required: true},
		{name: "run_id", location: "query", required: true},
		{name: "metric_key", location: "query", required: true},
		{name: "page_token", location: "query", required: false},
	} {
		found := false
		for _, parameter := range parameters {
			if stringAt(parameter, "name") == want.name && stringAt(parameter, "in") == want.location && parameter["required"] == want.required {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("retained MLflow source omits required parameter fact %#v", want)
		}
	}

	// RED on the frozen base: the exact source-backed ETL lane is still M-U.
	etl := objectAt(objectAt(row, "lanes"), "etl")
	if stringAt(etl, "disposition") != "implemented" {
		t.Fatalf("RED: MLflow ETL %q disposition=%q, want implemented only after a field-complete source-bound stream is declared", sourceID, stringAt(etl, "disposition"))
	}

	mapping := objectAt(etl, "mapping")
	if stringAt(mapping, "source_id") != sourceID || stringAt(mapping, "stream") != streamName || stringAt(mapping, "schema") != "schemas/mlflow_metrics_history.json" || stringAt(mapping, "mode") != "full_refresh" {
		t.Fatalf("MLflow ETL mapping=%#v, want exact source ID, stream, schema, and full-refresh mode", mapping)
	}
	bindings := mustMapSlice(mapping["config_bindings"])
	wantBindings := []map[string]any{
		{"config_key": "project_id", "source_parameter": map[string]any{"in": "path", "name": "id"}, "template": "{{ config.project_id }}"},
		{"config_key": "mlflow_run_id", "source_parameter": map[string]any{"in": "query", "name": "run_id"}, "template": "{{ config.mlflow_run_id }}"},
		{"config_key": "mlflow_metric_key", "source_parameter": map[string]any{"in": "query", "name": "metric_key"}, "template": "{{ config.mlflow_metric_key }}"},
	}
	if len(bindings) != len(wantBindings) {
		t.Fatalf("MLflow config bindings=%#v, want the three explicit source-backed bindings", bindings)
	}
	for index := range wantBindings {
		if !sameGitLabJSON(bindings[index], wantBindings[index]) {
			t.Fatalf("MLflow config binding %d=%#v, want %#v", index, bindings[index], wantBindings[index])
		}
	}

	spec := loadGitLabObject(t, "spec.json")
	for _, key := range []string{"project_id", "mlflow_run_id", "mlflow_metric_key"} {
		if _, declared := objectAt(spec, "properties")[key]; !declared {
			t.Fatalf("MLflow stream config key %q is not declared in spec.json", key)
		}
	}

	bundle, err := engine.Load(os.DirFS(".."), "gitlab")
	if err != nil {
		t.Fatalf("load GitLab MLflow bundle: %v", err)
	}
	streamIndex := -1
	for index := range bundle.Streams {
		if bundle.Streams[index].Name == streamName {
			streamIndex = index
			break
		}
	}
	if streamIndex < 0 {
		t.Fatalf("MLflow source-bound ETL stream %q is not declared", streamName)
	}
	stream := bundle.Streams[streamIndex]
	if stream.Path != "/projects/{{ config.project_id }}/ml/mlflow/api/2.0/mlflow/metrics/get-history" || stream.Records.Path != "metrics" || stream.SchemaRef != "schemas/mlflow_metrics_history.json" || stream.Pagination == nil || stream.Pagination.Type != "cursor" || stream.Pagination.CursorParam != "page_token" || stream.Pagination.TokenPath != "next_page_token" {
		t.Fatalf("MLflow stream=%+v, want exact config-bound path, records, schema, and source cursor", stream)
	}
	if stream.Query["run_id"].Template != "{{ config.mlflow_run_id }}" || stream.Query["metric_key"].Template != "{{ config.mlflow_metric_key }}" || stream.Query["max_results"].Template != "1000" {
		t.Fatalf("MLflow stream query=%+v, want explicit required source config and source default max_results", stream.Query)
	}

	if err := engine.PreflightSourceBoundStreamRead(bundle, streamName, rawOperationID, "GET", mappingPath); err != nil {
		t.Fatalf("valid MLflow source-bound stream preflight: %v", err)
	}
	for _, testCase := range []struct {
		name            string
		sourceOperation string
		method          string
		path            string
		mutate          func(*engine.Bundle)
		want            string
	}{
		{name: "source operation", sourceOperation: "not-a-retained-operation", method: "GET", path: mappingPath, want: "does not match"},
		{name: "method", sourceOperation: rawOperationID, method: "POST", path: mappingPath, want: "does not match"},
		{name: "path", sourceOperation: rawOperationID, method: "GET", path: "/projects/{id}/other", want: "does not match"},
		{name: "schema", sourceOperation: rawOperationID, method: "GET", path: mappingPath, mutate: func(b *engine.Bundle) { b.Streams[streamIndex].SchemaRef = "" }, want: "record semantics"},
		{name: "pagination", sourceOperation: rawOperationID, method: "GET", path: mappingPath, mutate: func(b *engine.Bundle) { b.Streams[streamIndex].Pagination = nil; b.HTTP.Pagination = nil }, want: "pagination semantics"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			broken := bundle
			broken.Streams = append([]engine.StreamSpec(nil), bundle.Streams...)
			if testCase.mutate != nil {
				testCase.mutate(&broken)
			}
			err := engine.PreflightSourceBoundStreamRead(broken, streamName, testCase.sourceOperation, testCase.method, testCase.path)
			if err == nil || !strings.Contains(err.Error(), testCase.want) {
				t.Fatalf("MLflow %s preflight error=%v, want %q before credential or provider I/O", testCase.name, err, testCase.want)
			}
		})
	}

	for _, testCase := range []struct {
		name   string
		mutate func(map[string]any)
	}{
		{name: "disposition", mutate: func(row map[string]any) {
			objectAt(row, "lanes")["etl"] = gitLabMappedCell("regression", "source_request_response_continuation_candidate", map[string]any{"kind": "source_lock", "path": gitLabSourceLockPath, "source_id": sourceID})
		}},
		{name: "source ID", mutate: func(row map[string]any) {
			objectAt(objectAt(objectAt(row, "lanes"), "etl"), "mapping")["source_id"] = "gitlab.rest.not-the-locked-mlflow-operation"
		}},
		{name: "stream", mutate: func(row map[string]any) {
			objectAt(objectAt(objectAt(row, "lanes"), "etl"), "mapping")["stream"] = "issues"
		}},
		{name: "schema", mutate: func(row map[string]any) {
			objectAt(objectAt(objectAt(row, "lanes"), "etl"), "mapping")["schema"] = "schemas/issues.json"
		}},
		{name: "pagination", mutate: func(row map[string]any) {
			objectAt(objectAt(objectAt(row, "lanes"), "etl"), "mapping")["pagination"] = map[string]any{"type": "cursor", "cursor_param": "page", "token_path": "next"}
		}},
	} {
		t.Run("matrix rejects "+testCase.name+" mismatch", func(t *testing.T) {
			broken := cloneGitLabObject(t, matrix)
			testCase.mutate(gitLabMatrixRow(t, broken, sourceID))
			if err := validateGitLabSourceLaneMatrix(broken, lock, loadGitLabObject(t, gitLabBinaryLockPath), loadGitLabObject(t, gitLabCrosswalkPath), loadGitLabObject(t, gitLabDescriptorPath), loadGitLabObject(t, gitLabRetainedArtifacts), loadGitLabObject(t, gitLabStreamsPath)); err == nil || !strings.Contains(err.Error(), "lane cells") {
				t.Fatalf("MLflow matrix %s mismatch error=%v, want deterministic lane-cell rejection", testCase.name, err)
			}
		})
	}
}

func validateGitLabSourceLaneMatrix(matrix, lock, binaryLock, crosswalk, descriptor, retainedArtifacts, streams map[string]any) error {
	if numberAt(matrix, "schema_version") != 1 || stringAt(matrix, "connector") != "gitlab" {
		return fmt.Errorf("matrix identity drift")
	}
	if !reflect.DeepEqual(stringSlice(matrix["lanes"]), gitLabLanes) {
		return fmt.Errorf("lane order drift")
	}
	if err := validateGitLabSnapshotBinding(matrix, lock, binaryLock, crosswalk, descriptor, retainedArtifacts); err != nil {
		return err
	}

	primary := mustMapSlice(objectAt(lock, "rest")["operations"])
	descriptors := mustMapSlice(descriptor["operations"])
	descriptorByID := make(map[string]map[string]any, len(descriptors))
	for _, operation := range descriptors {
		id := stringAt(operation, "operation_id")
		if _, exists := descriptorByID[id]; exists {
			return fmt.Errorf("duplicate descriptor operation ID %q", id)
		}
		descriptorByID[id] = operation
	}

	primaryByID := make(map[string]map[string]any, len(primary))
	for _, operation := range primary {
		id := stringAt(operation, "id")
		if _, exists := primaryByID[id]; exists {
			return fmt.Errorf("duplicate primary source lock ID %q", id)
		}
		if descriptorByID[id] == nil {
			return fmt.Errorf("primary source lock ID %q has no descriptor", id)
		}
		primaryByID[id] = operation
	}
	if numberAt(objectAt(lock, "counts"), "total") != 1752 || len(primaryByID) != 1752 || len(descriptorByID) != 1752 {
		return fmt.Errorf("primary source lock or descriptor denominator drift")
	}
	if err := validateGitLabPaginationReconciliation(matrix, primaryByID, descriptorByID); err != nil {
		return err
	}

	supplemental := gitLabSupplementalOperations(binaryLock)
	if numberAt(objectAt(binaryLock, "counts"), "total") != 2 || len(supplemental) != 2 {
		return fmt.Errorf("supplemental binary source lock denominator drift")
	}

	if err := validateGitLabCrosswalkBoundary(matrix, crosswalk, primaryByID, supplemental); err != nil {
		return err
	}
	if err := validateGitLabMappingRestrictions(matrix, lock, descriptor, primaryByID); err != nil {
		return err
	}
	if err := validateGitLabFoundationAtlas(matrix); err != nil {
		return err
	}
	if err := validateGitLabLegacyStreamBacklinks(matrix, primaryByID, descriptorByID, streams); err != nil {
		return err
	}

	rows := mustMapSlice(matrix["source_operations"])
	if len(rows) != 1754 {
		return fmt.Errorf("source rows matrix=%d, want 1754", len(rows))
	}
	counts := make(map[string]map[string]int, len(gitLabLanes))
	seen := make(map[string]struct{}, len(rows))
	rest := objectAt(lock, "rest")
	for _, row := range rows {
		id := stringAt(row, "source_id")
		if _, duplicate := seen[id]; duplicate {
			return fmt.Errorf("duplicate matrix source ID %q", id)
		}
		seen[id] = struct{}{}

		var expectedFacts map[string]any
		var sourceOperation map[string]any
		expectedOrigin := "rendered_reference_supplement"
		if strings.HasPrefix(id, "gitlab.rest.") {
			operationID := strings.TrimPrefix(id, "gitlab.rest.")
			operation := primaryByID[operationID]
			if operation == nil {
				return fmt.Errorf("matrix primary source ID %q is absent from source lock", id)
			}
			expectedFacts = expectedGitLabPrimaryFacts(operation, descriptorByID[operationID], rest)
			sourceOperation = objectAt(operation, "source_operation")
			expectedOrigin = "primary_openapi"
		} else {
			operation := supplemental[id]
			if operation == nil {
				return fmt.Errorf("matrix supplemental source ID %q is absent from binary source lock", id)
			}
			expectedFacts = expectedGitLabSupplementalFacts(operation)
		}
		if stringAt(row, "source_origin") != expectedOrigin {
			return fmt.Errorf("source origin %q drift", id)
		}
		if !sameGitLabJSON(objectAt(row, "source_facts"), expectedFacts) {
			return fmt.Errorf("source facts %q drift", id)
		}

		lanes := objectAt(row, "lanes")
		if len(lanes) != len(gitLabLanes) {
			return fmt.Errorf("lane cells %s=%d, want %d", id, len(lanes), len(gitLabLanes))
		}
		wantLanes := expectedGitLabLanes(id, expectedFacts, sourceOperation)
		if !sameGitLabJSON(lanes, wantLanes) {
			return fmt.Errorf("lane cells %q drift", id)
		}
		for _, lane := range gitLabLanes {
			cell := objectAt(lanes, lane)
			disposition := stringAt(cell, "disposition")
			if counts[lane] == nil {
				counts[lane] = make(map[string]int)
			}
			counts[lane][disposition]++
		}
	}
	if len(seen) != 1754 {
		return fmt.Errorf("matrix source rows retained=%d, want 1754", len(seen))
	}
	for id := range primaryByID {
		if _, exists := seen["gitlab.rest."+id]; !exists {
			return fmt.Errorf("primary source ID hidden from matrix %q", id)
		}
	}
	for id := range supplemental {
		if _, exists := seen[id]; !exists {
			return fmt.Errorf("supplemental source ID hidden from matrix %q", id)
		}
	}

	summary := objectAt(matrix, "summary")
	if numberAt(summary, "source_rows") != 1754 || numberAt(summary, "source_rows_with_all_lanes") != 1754 || numberAt(summary, "total_lane_cells") != 12278 {
		return fmt.Errorf("matrix summary source accounting drift")
	}
	for _, lane := range gitLabLanes {
		if !equalGitLabCounts(counts[lane], gitLabExpectedCounts[lane]) {
			return fmt.Errorf("computed %s counts=%v, want %v", lane, counts[lane], gitLabExpectedCounts[lane])
		}
		if !equalGitLabCounts(numberMap(objectAt(summary, "lane_counts")[lane]), counts[lane]) {
			return fmt.Errorf("summary %s lane counts drift", lane)
		}
	}
	return nil
}

func validateGitLabSnapshotBinding(matrix, lock, binaryLock, crosswalk, descriptor, retainedArtifacts map[string]any) error {
	snapshot := objectAt(matrix, "source_snapshot")
	if stringAt(snapshot, "source_snapshot_ref") != "fm/cli-top100-declaration-batch-r1" || stringAt(snapshot, "source_snapshot_commit") != gitLabSnapshotCommit || stringAt(snapshot, "materialization") != "source_lock_backed_descriptor_correction" || stringAt(snapshot, "base_materialization") != "git_archive_byte_identical" {
		return fmt.Errorf("source snapshot identity drift")
	}
	provenanceIdentity := objectAt(snapshot, "descriptor_correction_provenance")
	provenancePath := stringAt(provenanceIdentity, "path")
	provenance, err := readGitLabObject(provenancePath)
	if err != nil {
		return fmt.Errorf("read descriptor correction provenance %q: %w", provenancePath, err)
	}
	provenanceContents, err := os.ReadFile(provenancePath)
	if err != nil {
		return fmt.Errorf("read descriptor correction provenance bytes %q: %w", provenancePath, err)
	}
	if len(provenanceContents) != numberAt(provenanceIdentity, "bytes") || gitLabBlobSHA1(provenanceContents) != stringAt(provenanceIdentity, "git_blob_sha1") {
		return fmt.Errorf("descriptor correction provenance byte/blob identity drift")
	}
	if stringAt(provenance, "connector") != "gitlab" || stringAt(provenance, "kind") != "source_lock_backed_descriptor_correction" {
		return fmt.Errorf("descriptor correction provenance identity drift")
	}
	baseSnapshot := objectAt(provenance, "base_snapshot")
	if stringAt(baseSnapshot, "ref") != stringAt(snapshot, "source_snapshot_ref") || stringAt(baseSnapshot, "commit") != stringAt(snapshot, "source_snapshot_commit") || stringAt(baseSnapshot, "materialization") != stringAt(snapshot, "base_materialization") {
		return fmt.Errorf("descriptor correction base snapshot drift")
	}
	target := objectAt(provenance, "target")
	targetPath := stringAt(target, "path")
	baseFiles, err := gitLabSourceFileIdentityMap(mustMapSlice(snapshot["base_retained_files"]))
	if err != nil {
		return fmt.Errorf("base retained source files: %w", err)
	}
	files, err := gitLabSourceFileIdentityMap(mustMapSlice(snapshot["retained_files"]))
	if err != nil {
		return fmt.Errorf("derived retained source files: %w", err)
	}
	if len(files) == 0 || len(files) != len(baseFiles) {
		return fmt.Errorf("retained source file inventories differ")
	}
	original := objectAt(target, "original")
	derived := objectAt(target, "derived")
	for filePath, base := range baseFiles {
		got, found := files[filePath]
		if !found {
			return fmt.Errorf("derived retained source file inventory omits %q", filePath)
		}
		if filePath == targetPath {
			if !sameGitLabJSON(base, original) || !sameGitLabJSON(got, derived) {
				return fmt.Errorf("descriptor correction retained identity drift")
			}
		} else if !sameGitLabJSON(base, got) {
			return fmt.Errorf("uncorrected retained source file identity changed %q", filePath)
		}
		contents, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("read retained source file %q: %w", filePath, err)
		}
		if len(contents) != numberAt(got, "bytes") || gitLabBlobSHA1(contents) != stringAt(got, "git_blob_sha1") {
			return fmt.Errorf("retained source file byte/blob identity drift %q", filePath)
		}
	}
	if _, found := baseFiles[targetPath]; !found {
		return fmt.Errorf("descriptor correction target %q is not retained", targetPath)
	}
	sourceLock := objectAt(provenance, "source_lock")
	rest := objectAt(lock, "rest")
	if stringAt(sourceLock, "path") != gitLabSourceLockPath || stringAt(sourceLock, "source_url") != stringAt(rest, "source_url") || stringAt(sourceLock, "sha256") != stringAt(rest, "sha256") || numberAt(sourceLock, "bytes") != numberAt(rest, "bytes") {
		return fmt.Errorf("descriptor correction source-lock evidence drift")
	}
	if stringAt(lock, "connector") != "gitlab" || stringAt(binaryLock, "connector") != "gitlab" || stringAt(crosswalk, "connector") != "gitlab" || numberAt(descriptor, "schema_version") < 1 || stringAt(retainedArtifacts, "connector") != "gitlab" {
		return fmt.Errorf("retained source artifact identity drift")
	}
	for _, artifact := range mustMapSlice(retainedArtifacts["artifacts"]) {
		sha := stringAt(artifact, "sha256")
		contents, err := os.ReadFile("sources/artifacts/" + sha + ".artifact")
		if err != nil {
			return fmt.Errorf("read retained artifact %q: %w", sha, err)
		}
		digest := sha256.Sum256(contents)
		if hex.EncodeToString(digest[:]) != sha || len(contents) != numberAt(artifact, "bytes") {
			return fmt.Errorf("retained artifact digest/bytes drift %q", sha)
		}
	}
	return nil
}

func gitLabSourceFileIdentityMap(files []map[string]any) (map[string]map[string]any, error) {
	identities := make(map[string]map[string]any, len(files))
	for _, file := range files {
		filePath := stringAt(file, "path")
		if filePath == "" || stringAt(file, "git_blob_sha1") == "" || numberAt(file, "bytes") <= 0 {
			return nil, fmt.Errorf("retained source file identity is incomplete")
		}
		if _, duplicate := identities[filePath]; duplicate {
			return nil, fmt.Errorf("retained source file identity is duplicated %q", filePath)
		}
		identities[filePath] = file
	}
	return identities, nil
}

func validateGitLabCrosswalkBoundary(matrix, crosswalk map[string]any, primary map[string]map[string]any, supplemental map[string]map[string]any) error {
	if stringAt(crosswalk, "source_lock") != gitLabSourceLockPath || len(mustMapSlice(crosswalk["source_operations"])) != 1755 {
		return fmt.Errorf("crosswalk identity or denominator drift")
	}
	boundary := objectAt(matrix, "source_boundary_reconciliation")
	if stringAt(boundary, "identity") != "source_id" || numberAt(boundary, "primary_source_rows") != 1752 || numberAt(boundary, "supplemental_binary_source_rows") != 2 || numberAt(boundary, "retained_source_rows") != 1754 || numberAt(boundary, "crosswalk_rows") != 1755 || numberAt(boundary, "crosswalk_only_rows") != 3 || numberAt(boundary, "supplemental_not_in_primary_crosswalk_rows") != 2 {
		return fmt.Errorf("crosswalk boundary accounting drift")
	}
	primaryIDs := make(map[string]struct{}, len(primary))
	for id := range primary {
		primaryIDs["gitlab.rest."+id] = struct{}{}
	}
	for id := range supplemental {
		if _, inPrimary := primaryIDs[id]; inPrimary {
			return fmt.Errorf("supplemental source ID %q unexpectedly appears in primary source namespace", id)
		}
	}

	want := make([]any, 0, 3)
	for _, entry := range mustMapSlice(crosswalk["source_operations"]) {
		id := stringAt(entry, "source_id")
		if _, present := primaryIDs[id]; present {
			continue
		}
		want = append(want, map[string]any{
			"source_id":       id,
			"operation_id":    stringAt(entry, "operation_id"),
			"method":          stringAt(entry, "method"),
			"path":            stringAt(entry, "path"),
			"source_location": stringAt(entry, "source_location"),
			"disposition":     "not_source_row",
			"reason":          "Present in the retained GitLab crosswalk but absent from the immutable primary and supplemental source locks; this is boundary evidence, not a matrix source row.",
			"crosswalk_entry": entry,
		})
	}
	if len(want) != 3 {
		return fmt.Errorf("crosswalk-minus-primary source accounting=%d, want 3", len(want))
	}
	if !sameGitLabJSON(boundary["crosswalk_only_source_identities"], want) {
		return fmt.Errorf("crosswalk boundary identities drift")
	}
	return nil
}

func validateGitLabMappingRestrictions(matrix, lock, descriptor map[string]any, primary map[string]map[string]any) error {
	want := expectedGitLabMappingRestrictions(lock, descriptor, primary)
	if !sameGitLabJSON(matrix["mapping_restrictions"], want) {
		return fmt.Errorf("source-visible mapping restrictions drift")
	}
	return nil
}

func validateGitLabFoundationAtlas(matrix map[string]any) error {
	atlas := objectAt(matrix, "foundation_atlas")
	if stringAt(atlas, "consulted_snapshot_ref") != "fm/cli-top100-declaration-batch-r1" || stringAt(atlas, "consulted_snapshot_commit") != gitLabSnapshotCommit || stringAt(atlas, "catalog_path") != "docs/connector-canon/foundations/catalog.json" || stringAt(atlas, "usage") != "authoring_only_not_a_runtime_loader" {
		return fmt.Errorf("Foundation Atlas provenance drift")
	}
	wantReuse := []any{"source.retention-import.v1", "source.projection-admission.v1", "runtime.direct-execution.v1", "warehouse.stage-etl.v1", "warehouse.reverse-etl.v1", "transport.sync-contract.v1"}
	if !sameGitLabJSON(atlas["consulted_capabilities"], wantReuse) {
		return fmt.Errorf("Foundation Atlas capability lookup drift")
	}
	gap := objectAt(atlas, "sync_actual_gap")
	if stringAt(gap, "gap_id") != "gitlab-inbound-webhook-source-executor-r1" || stringAt(gap, "lane") != "sync_transport" || stringAt(gap, "consulted_atlas_id") != "transport.sync-contract.v1" || stringAt(gap, "status") != "recorded_only_requires_captain_approval_before_implementation" || !sameGitLabJSON(gap["source_ids"], []any{"gitlab.rest.postApiV4GroupsIdHooks", "gitlab.rest.postApiV4Hooks", "gitlab.rest.postApiV4ProjectsIdHooks"}) || strings.TrimSpace(stringAt(gap, "missing_capability")) == "" || strings.TrimSpace(stringAt(gap, "why_existing_capability_is_insufficient")) == "" || strings.TrimSpace(stringAt(gap, "proof_test_idea")) == "" {
		return fmt.Errorf("Foundation Atlas sync-gap record drift")
	}
	return nil
}

func validateGitLabPaginationReconciliation(matrix map[string]any, primary, descriptors map[string]map[string]any) error {
	states := make(map[string]int)
	for operationID, operation := range primary {
		states[gitLabPaginationState(gitLabPaginationFacts(operation, descriptors[operationID]))]++
	}
	if states["request_response_continuation_candidate"] == 0 || states["request_controls_without_response_continuation"] == 0 {
		return fmt.Errorf("source pagination evidence accounting=%v, want both retained continuation and retained incomplete-control states", states)
	}
	reconciliation := objectAt(matrix, "source_paging_reconciliation")
	if stringAt(reconciliation, "criterion") != "Retained request continuation control paired with retained successful-response continuation evidence; method and operation name alone never establish ETL." ||
		numberAt(reconciliation, "request_response_continuation_candidates") != states["request_response_continuation_candidate"] ||
		numberAt(reconciliation, "request_controls_without_response_continuation") != states["request_controls_without_response_continuation"] ||
		numberAt(reconciliation, "total_retained_pagination_facts") != states["request_response_continuation_candidate"]+states["request_controls_without_response_continuation"] {
		return fmt.Errorf("matrix source pagination reconciliation drift: computed states=%v", states)
	}
	return nil
}

func validateGitLabLegacyStreamBacklinks(matrix map[string]any, primary, descriptors map[string]map[string]any, streams map[string]any) error {
	streamPaths := make(map[string]string)
	for _, stream := range mustMapSlice(streams["streams"]) {
		streamPaths[stringAt(stream, "name")] = stringAt(stream, "path")
	}
	for id, want := range gitLabLegacyStreams {
		op := primary[strings.TrimPrefix(id, "gitlab.rest.")]
		desc := descriptors[strings.TrimPrefix(id, "gitlab.rest.")]
		if op == nil || desc == nil || gitLabPaginationState(gitLabPaginationFacts(op, desc)) == "not_documented_by_locked_operation" {
			return fmt.Errorf("legacy stream source lookup drift %q", id)
		}
		if stringAt(op, "method") != "GET" || streamPaths[want["stream"]] != want["path"] {
			return fmt.Errorf("legacy stream source/definition drift %q", id)
		}
	}
	return nil
}

func expectedGitLabPrimaryFacts(operation, descriptor, rest map[string]any) map[string]any {
	id := stringAt(operation, "id")
	params := mapSliceOrEmpty(objectAt(operation, "source_operation")["parameters"])
	facts := map[string]any{
		"source_kind":    "primary_openapi",
		"source_lock":    gitLabSourceLockPath,
		"source_id":      "gitlab.rest." + id,
		"operation_id":   id,
		"protocol":       stringAt(operation, "protocol"),
		"method":         stringAt(operation, "method"),
		"path":           stringAt(operation, "path"),
		"mapping_path":   stringAt(descriptor, "mapping_path"),
		"deprecated":     operation["deprecated"],
		"source_summary": stringAt(objectAt(operation, "source_operation"), "summary"),
		"citation": map[string]any{
			"url":      stringAt(rest, "source_url"),
			"sha256":   stringAt(rest, "sha256"),
			"bytes":    rest["bytes"],
			"location": stringAt(operation, "source_location"),
		},
		"scope_fanout": map[string]any{
			"path_parameters":  sourceParameterNames(params, "path"),
			"query_parameters": sourceParameterNames(params, "query"),
		},
		"request_media_types":          sourceRequestMediaTypes(operation),
		"success_response_media_types": sourceSuccessResponseMediaTypes(operation),
		"pagination":                   gitLabPaginationFacts(operation, descriptor),
		"binary": map[string]any{
			"request_binary_fields": gitLabBinaryRequestFields(descriptor),
			"download_state":        "not_declared_by_primary_source",
		},
		"event_cursor":                   gitLabEventFacts(descriptor),
		"mapping_restriction_record_ids": gitLabRestrictionIDs("gitlab.rest." + id),
		"crosswalk_state":                "primary_crosswalk_exact",
	}
	// The source facts already retain method, summary, citation, and response
	// media. Store an explicit semantic record only where it changes the
	// historical verb-based classification, avoiding noise for ordinary GETs.
	semantics := gitLabOperationSemantics(operation)
	if state := stringAt(semantics, "state"); state == "source_semantic_head_read" || state == "source_semantic_post_read" {
		facts["operation_semantics"] = semantics
	}
	return facts
}

func expectedGitLabSupplementalFacts(operation map[string]any) map[string]any {
	id := stringAt(operation, "id")
	role := "generic_package_file_upload"
	if id == "gitlab.docs.repository_files.raw_download" {
		role = "repository_file_raw_download"
	}
	return map[string]any{
		"source_kind": "rendered_reference_supplement",
		"source_lock": gitLabBinaryLockPath,
		"source_id":   id,
		"protocol":    stringAt(operation, "protocol"),
		"method":      stringAt(operation, "method"),
		"path":        stringAt(operation, "path"),
		"citation": map[string]any{
			"url":      stringAt(operation, "citation_url"),
			"sha256":   operation["document_sha256"],
			"bytes":    operation["document_bytes"],
			"location": stringAt(operation, "source_location"),
		},
		"scope_fanout": map[string]any{
			"path_parameters":  pathTemplateParameters(stringAt(operation, "path")),
			"query_parameters": []any{},
		},
		"request_media_types":          []any{},
		"success_response_media_types": []any{},
		"pagination":                   map[string]any{"kind": "not_documented", "controls": []any{}},
		"binary": map[string]any{
			"rendered_reference_role": role,
		},
		"event_cursor":    map[string]any{"state": "not_documented"},
		"crosswalk_state": "supplemental_source_not_in_primary_crosswalk",
	}
}

func expectedGitLabLanes(id string, facts, sourceOperation map[string]any) map[string]any {
	method := strings.ToUpper(stringAt(facts, "method"))
	lockPath := stringAt(facts, "source_lock")
	backlink := map[string]any{"kind": "source_lock", "path": lockPath, "source_id": id}
	lanes := make(map[string]any, len(gitLabLanes))
	semantics := gitLabFactsOperationSemantics(facts)
	switch stringAt(semantics, "state") {
	case "source_safe_get_read":
		lanes["direct_read"] = gitLabMappedCell("Locked GET source row; source-backed direct-read candidate only.", "read_verb_candidate", backlink)
	case "source_semantic_head_read":
		lanes["direct_read"] = gitLabMappedCell("Locked HEAD source row has a retained successful metadata response; source-backed bounded direct-read candidate only.", "source_semantic_head_read_candidate", backlink)
	case "source_semantic_post_read":
		reason := "Locked POST source summary and retained successful response document a bounded query/lookup read; source-backed direct-read candidate only."
		classification := "source_semantic_post_read_candidate"
		if gitLabSemanticPostReadRequiresUntypedJSONBody(sourceOperation) {
			reason = "Locked POST source summary and retained successful response document a bounded query/lookup read, but its retained description requires a JSON object with the name and size of the individual files and provides no requestBody media type or closed schema; keep the source-backed direct-read cell mapped_unproven and do not materialize a bodyless or raw-body request."
			classification = "source_semantic_post_read_requires_untyped_json_body"
		}
		lanes["direct_read"] = gitLabMappedCell(reason, classification, backlink)
	default:
		lanes["direct_read"] = gitLabNotApplicableCell("Locked method " + method + " is not a GET direct-read candidate.")
	}
	if gitLabMutationCandidate(method, semantics) {
		cell := gitLabMappedCell("Locked "+method+" source row; source-backed mutation candidate only.", "mutation_verb_candidate", backlink)
		lanes["direct_write"] = cell
		lanes["reverse_etl"] = cloneGitLabMap(cell)
	} else {
		lanes["direct_write"] = gitLabNotApplicableCell("Locked method " + method + " is not a provider-mutation candidate.")
		lanes["reverse_etl"] = gitLabNotApplicableCell("Locked method " + method + " is not a provider-mutation candidate.")
	}

	binary := objectAt(facts, "binary")
	if stringAt(binary, "rendered_reference_role") == "repository_file_raw_download" {
		lanes["binary_download"] = gitLabMappedCell("Retained rendered GitLab reference explicitly documents repository-file raw download.", "binary_download_candidate", backlink)
	} else {
		lanes["binary_download"] = gitLabNotApplicableCell("No retained source evidence declares a binary-download candidate.")
	}
	if fields := stringSliceOrEmpty(binary["request_binary_fields"]); len(fields) > 0 {
		lanes["binary_upload"] = gitLabMappedCell("Resolved source request schema declares binary field(s): "+strings.Join(fields, ", ")+".", "binary_upload_candidate", backlink)
	} else if stringAt(binary, "rendered_reference_role") == "generic_package_file_upload" {
		lanes["binary_upload"] = gitLabMappedCell("Retained rendered GitLab reference explicitly documents generic-package file upload.", "binary_upload_candidate", backlink)
	} else {
		lanes["binary_upload"] = gitLabNotApplicableCell("No retained source evidence declares a binary-upload candidate.")
	}

	pagination := objectAt(facts, "pagination")
	if gitLabPaginationState(pagination) == "request_response_continuation_candidate" && gitLabDirectReadCandidate(semantics) {
		if materialized, ok := gitLabMaterializedETLSpecs[id]; ok {
			lanes["etl"] = gitLabImplementedETLCell(id, materialized)
			return gitLabFinalizeExpectedLanes(lanes, facts, id, backlink)
		}
		mapping := backlink
		if stream, ok := gitLabLegacyStreams[id]; ok {
			mapping = map[string]any{"kind": "existing_stream", "path": gitLabStreamsPath, "stream": stream["stream"], "stream_path": stream["path"]}
		}
		lanes["etl"] = gitLabMappedCell("Source retains a request-to-response continuation pair: "+strings.Join(stringSlice(pagination["request_controls"]), ", ")+" -> "+strings.Join(stringSlice(pagination["response_controls"]), ", ")+".", "source_request_response_continuation_candidate", mapping)
	} else if gitLabPaginationState(pagination) == "request_controls_without_response_continuation" {
		lanes["etl"] = gitLabNotApplicableCell("Source retains pagination-shaped request controls but no retained response continuation; do not claim an ETL candidate.")
	} else {
		lanes["etl"] = gitLabNotApplicableCell("No explicit source pagination controls match the Track A extraction criterion.")
	}

	return gitLabFinalizeExpectedLanes(lanes, facts, id, backlink)
}

func gitLabFinalizeExpectedLanes(lanes map[string]any, facts map[string]any, id string, backlink map[string]any) map[string]any {
	event := objectAt(facts, "event_cursor")
	if stringAt(event, "state") == "webhook_registration" {
		lanes["sync_transport"] = map[string]any{
			"applicability": "source_candidate",
			"disposition":   "missing_foundation",
			"reason":        "Source documents webhook registration with a required URL and event selectors; the consulted Atlas has no closed inbound GitLab webhook source executor.",
			"mapping": map[string]any{
				"source_fact":         map[string]any{"classification": "webhook_registration_candidate", "source_id": id},
				"definition_backlink": backlink,
				"foundation_gap_id":   "gitlab-inbound-webhook-source-executor-r1",
				"consulted_atlas_id":  "transport.sync-contract.v1",
				"runtime_claim":       "No inbound GitLab webhook receiver or selected source executor is claimed.",
			},
		}
	} else {
		lanes["sync_transport"] = gitLabNotApplicableCell("No retained source fact documents a webhook registration with required URL and event selectors.")
	}
	return lanes
}

func gitLabImplementedETLCell(sourceID string, spec map[string]any) map[string]any {
	return map[string]any{
		"applicability": "source_candidate",
		"disposition":   "implemented",
		"reason":        "Retained source documents page_token -> next_page_token continuation; the exact source-bound stream, schema, and declared full-refresh DuckDB route are materialized.",
		"mapping": map[string]any{
			"source_id":       sourceID,
			"stream":          spec["stream"],
			"schema":          spec["schema"],
			"mode":            spec["mode"],
			"records_path":    spec["records_path"],
			"pagination":      spec["pagination"],
			"config_bindings": spec["config_bindings"],
			"source_fact": map[string]any{
				"classification": "source_request_response_continuation_candidate",
				"continuation":   map[string]any{"request": "page_token", "response": "next_page_token"},
			},
			"definition_backlink": map[string]any{
				"kind":         "stream_etl",
				"path":         "operations.json",
				"operation":    spec["operation"],
				"streams_path": gitLabStreamsPath,
				"stream":       spec["stream"],
			},
			"runtime_claim": "Declared source-bound stream_etl is exercised through the connector's existing declarative provider-to-DuckDB full-refresh route.",
		},
	}
}

func gitLabSemanticPostReadRequiresUntypedJSONBody(sourceOperation map[string]any) bool {
	const bodyRequirement = "the request must include a json object with the name and size of the individual files."
	if _, hasRequestBody := sourceOperation["requestBody"]; hasRequestBody {
		return false
	}
	return strings.Contains(strings.ToLower(stringAt(sourceOperation, "description")), bodyRequirement)
}

func gitLabMappedCell(reason, classification string, backlink map[string]any) map[string]any {
	return map[string]any{
		"applicability": "source_candidate",
		"disposition":   "mapped_unproven",
		"reason":        reason,
		"mapping": map[string]any{
			"source_fact":         map[string]any{"classification": classification},
			"definition_backlink": backlink,
			"runtime_claim":       gitLabRuntimeClaim,
		},
	}
}

func gitLabNotApplicableCell(reason string) map[string]any {
	return map[string]any{"applicability": "not_applicable", "disposition": "not_applicable", "reason": reason}
}

func expectedGitLabMappingRestrictions(lock, descriptor map[string]any, primary map[string]map[string]any) []any {
	rest := objectAt(lock, "rest")
	gaps := mustMapSlice(descriptor["gaps"])
	result := make([]any, 0, len(gitLabPathRestrictionSpecs))
	for _, spec := range gitLabPathRestrictionSpecs {
		operationID := strings.TrimPrefix(spec["source_id"], "gitlab.rest.")
		op := primary[operationID]
		if op == nil || !strings.HasSuffix(stringAt(op, "source_location"), spec["source_location_suffix"]) {
			panic("invalid static GitLab path restriction source ID")
		}
		var descriptorGap map[string]any
		for _, gap := range gaps {
			if stringAt(gap, "foundation") == "cli-malformed-path-parameter-foundation-r1" && stringAt(gap, "location") == spec["descriptor_location"] {
				descriptorGap = gap
				break
			}
		}
		if descriptorGap == nil {
			panic("missing expected static GitLab descriptor gap")
		}
		result = append(result, map[string]any{
			"record_id":             spec["record_id"],
			"state":                 "mapping_restriction",
			"source_id":             spec["source_id"],
			"source_location":       stringAt(op, "source_location"),
			"descriptor_gap":        descriptorGap,
			"missing_placeholder":   spec["missing_placeholder"],
			"rest_path_bridge":      rest["path_bridge"],
			"atlas_lookup":          "source.projection-admission.v1",
			"needed_mapping_repair": "Projection/import mapping must preserve the source path and record the missing required placeholder visibly; it must not synthesize a binding or erase the source row.",
			"status":                "recorded_only_no_shared_code_changed",
		})
	}
	return result
}

// gitLabOperationSemantics derives lane eligibility from the retained provider
// operation, rather than assuming every POST mutates or every GET is the only
// possible read. It intentionally uses source text and successful-response
// evidence, never an operation-ID allow-list.
func gitLabOperationSemantics(operation map[string]any) map[string]any {
	sourceOperation := objectAt(operation, "source_operation")
	method := strings.ToUpper(stringAt(operation, "method"))
	summary := stringAt(sourceOperation, "summary")
	description := stringAt(sourceOperation, "description")
	successStatuses := gitLabSuccessResponseStatuses(sourceOperation)
	state := "not_a_documented_bounded_read"
	switch {
	case method == "GET" && len(successStatuses) > 0:
		state = "source_safe_get_read"
	case method == "HEAD" && len(successStatuses) > 0:
		state = "source_semantic_head_read"
	case method == "POST" && len(successStatuses) > 0 && gitLabSemanticPostReadSummary(summary, description):
		state = "source_semantic_post_read"
	}
	return map[string]any{
		"state":                     state,
		"source_summary":            summary,
		"success_response_statuses": stringsToAny(successStatuses),
	}
}

func gitLabSemanticPostReadSummary(summary, description string) bool {
	words := strings.Fields(strings.ToLower(strings.TrimSpace(summary)))
	if len(words) == 0 {
		return false
	}
	switch words[0] {
	case "get", "retrieve", "list", "search", "searches", "query":
		return true
	case "execute":
		text := strings.ToLower(summary + " " + description)
		return strings.Contains(text, "query") || strings.Contains(text, "graphql")
	default:
		return false
	}
}

func gitLabSuccessResponseStatuses(sourceOperation map[string]any) []string {
	responses, _ := sourceOperation["responses"].(map[string]any)
	statuses := make([]string, 0, len(responses))
	for status := range responses {
		if strings.HasPrefix(status, "2") || strings.HasPrefix(status, "3") {
			statuses = append(statuses, status)
		}
	}
	sort.Strings(statuses)
	return statuses
}

// gitLabPaginationFacts records the request and response continuation facts
// separately. A pagination-shaped request is not enough to claim ETL: a
// retained source must also tell us how the next request is obtained.
func gitLabPaginationFacts(operation, descriptor map[string]any) map[string]any {
	sourceOperation := objectAt(operation, "source_operation")
	requestControls := gitLabPaginationRequestControls(sourceOperation, descriptor)
	responseControls := gitLabPaginationResponseControls(sourceOperation, descriptor)
	if len(requestControls) == 0 {
		return map[string]any{"kind": "not_documented", "controls": []any{}}
	}
	facts := map[string]any{
		"state":             "not_documented_by_locked_operation",
		"request_controls":  stringsToAny(requestControls),
		"response_controls": stringsToAny(responseControls),
	}
	if continuation, ok := gitLabContinuationPair(requestControls, responseControls); ok {
		facts["state"] = "request_response_continuation_candidate"
		facts["continuation"] = continuation
		return facts
	}
	if len(requestControls) > 0 {
		facts["state"] = "request_controls_without_response_continuation"
	}
	return facts
}

func gitLabPaginationState(facts map[string]any) string {
	if state := stringAt(facts, "state"); state != "" {
		return state
	}
	if stringAt(facts, "kind") == "not_documented" {
		return "not_documented_by_locked_operation"
	}
	return "invalid_pagination_fact"
}

func gitLabPaginationRequestControls(sourceOperation, descriptor map[string]any) []string {
	controls := make(map[string]string)
	for _, parameter := range mapSliceOrEmpty(sourceOperation["parameters"]) {
		if stringAt(parameter, "in") != "query" {
			continue
		}
		gitLabAddPaginationInput(controls, stringAt(parameter, "name"), stringAt(parameter, "description"))
	}
	if body, ok := sourceOperation["requestBody"].(map[string]any); ok {
		content, _ := body["content"].(map[string]any)
		for _, content := range content {
			contentMap, ok := content.(map[string]any)
			if !ok {
				continue
			}
			if schema, ok := contentMap["schema"].(map[string]any); ok {
				gitLabCollectPaginationInputs(schema, controls)
			}
		}
	}
	request := objectAt(descriptor, "request")
	for _, parameter := range mapSliceOrEmpty(request["query"]) {
		gitLabAddPaginationInput(controls, stringAt(parameter, "name"), stringAt(parameter, "description"))
	}
	if body, ok := request["body"].(map[string]any); ok {
		if schema, ok := body["schema"].(map[string]any); ok {
			gitLabCollectPaginationInputs(schema, controls)
		}
	}
	return gitLabSortedPaginationControlNames(controls)
}

func gitLabCollectPaginationInputs(schema map[string]any, controls map[string]string) {
	properties, _ := schema["properties"].(map[string]any)
	for name, raw := range properties {
		property, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		gitLabAddPaginationInput(controls, name, stringAt(property, "description"))
		gitLabCollectPaginationInputs(property, controls)
	}
	if items, ok := schema["items"].(map[string]any); ok {
		gitLabCollectPaginationInputs(items, controls)
	}
	for _, key := range []string{"allOf", "anyOf", "oneOf"} {
		for _, raw := range mapSliceOrEmpty(schema[key]) {
			gitLabCollectPaginationInputs(raw, controls)
		}
	}
}

func gitLabAddPaginationInput(controls map[string]string, name, _ string) {
	canonical := gitLabPaginationCanonicalName(name)
	if canonical == "" {
		return
	}
	// A retained provider parameter named page, cursor, after, or offset is
	// itself request-side pagination evidence. Descriptions are frequently
	// omitted or inconsistent across OpenAPI sources, so they must not hide a
	// candidate; a matching successful-response continuation is still required
	// before the operation receives an ETL lane.
	controls[canonical] = name
}

func gitLabPaginationResponseControls(sourceOperation, descriptor map[string]any) []string {
	controls := make(map[string]string)
	for status, raw := range objectAt(sourceOperation, "responses") {
		if !strings.HasPrefix(status, "2") {
			continue
		}
		response, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if headers, ok := response["headers"].(map[string]any); ok {
			for name, value := range headers {
				header, _ := value.(map[string]any)
				gitLabAddPaginationResponse(controls, name, stringAt(header, "description"))
			}
		}
		content, _ := response["content"].(map[string]any)
		for _, rawContent := range content {
			contentMap, ok := rawContent.(map[string]any)
			if !ok {
				continue
			}
			if schema, ok := contentMap["schema"].(map[string]any); ok {
				gitLabCollectPaginationResponses(schema, controls)
			}
		}
	}
	for _, response := range mapSliceOrEmpty(descriptor["responses"]) {
		if !strings.HasPrefix(stringAt(response, "status"), "2") {
			continue
		}
		declaration, _ := response["declaration"].(map[string]any)
		content, _ := declaration["content"].(map[string]any)
		for _, rawContent := range content {
			contentMap, ok := rawContent.(map[string]any)
			if !ok {
				continue
			}
			if schema, ok := contentMap["schema"].(map[string]any); ok {
				gitLabCollectPaginationResponses(schema, controls)
			}
		}
	}
	return gitLabSortedPaginationControlNames(controls)
}

func gitLabCollectPaginationResponses(schema map[string]any, controls map[string]string) {
	properties, _ := schema["properties"].(map[string]any)
	for name, raw := range properties {
		property, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		gitLabAddPaginationResponse(controls, name, stringAt(property, "description"))
		gitLabCollectPaginationResponses(property, controls)
	}
	if items, ok := schema["items"].(map[string]any); ok {
		gitLabCollectPaginationResponses(items, controls)
	}
	for _, key := range []string{"allOf", "anyOf", "oneOf"} {
		for _, raw := range mapSliceOrEmpty(schema[key]) {
			gitLabCollectPaginationResponses(raw, controls)
		}
	}
}

func gitLabAddPaginationResponse(controls map[string]string, name, description string) {
	canonical := gitLabPaginationCanonicalName(name)
	if canonical == "" {
		return
	}
	text := strings.ToLower(name + " " + description)
	switch canonical {
	case "nextpagetoken", "nextpage", "nextcursor", "endcursor", "hasnextpage", "hasmore", "nextoffset":
		controls[canonical] = name
	default:
		if strings.Contains(text, "next") || strings.Contains(text, "cursor") || strings.Contains(text, "pagination") {
			controls[canonical] = name
		}
	}
}

func gitLabPaginationCanonicalName(name string) string {
	normalized := strings.NewReplacer("_", "", "-", "", " ", "").Replace(strings.ToLower(name))
	switch normalized {
	case "page", "perpage", "pagetoken", "after", "cursor", "startcursor", "offset", "nextpagetoken", "nextpage", "nextcursor", "endcursor", "hasnextpage", "hasmore", "nextoffset":
		return normalized
	default:
		return ""
	}
}

func gitLabSortedPaginationControlNames(controls map[string]string) []string {
	values := make([]string, 0, len(controls))
	for _, name := range controls {
		values = append(values, name)
	}
	sort.Strings(values)
	return values
}

func gitLabContinuationPair(requestControls, responseControls []string) (map[string]any, bool) {
	request := gitLabPaginationControlIndex(requestControls)
	response := gitLabPaginationControlIndex(responseControls)
	switch {
	case request["after"] != "" && response["endcursor"] != "" && response["hasnextpage"] != "":
		return map[string]any{"request": request["after"], "response": response["endcursor"], "has_more": response["hasnextpage"]}, true
	case request["pagetoken"] != "" && response["nextpagetoken"] != "":
		return map[string]any{"request": request["pagetoken"], "response": response["nextpagetoken"]}, true
	case request["cursor"] != "" && response["nextcursor"] != "":
		return map[string]any{"request": request["cursor"], "response": response["nextcursor"]}, true
	case request["startcursor"] != "" && response["nextcursor"] != "":
		return map[string]any{"request": request["startcursor"], "response": response["nextcursor"]}, true
	case request["page"] != "" && request["perpage"] != "" && response["nextpage"] != "":
		return map[string]any{"request": request["page"], "response": response["nextpage"], "page_size": request["perpage"]}, true
	case request["offset"] != "" && response["nextoffset"] != "":
		return map[string]any{"request": request["offset"], "response": response["nextoffset"]}, true
	default:
		return nil, false
	}
}

func gitLabPaginationControlIndex(controls []string) map[string]string {
	index := make(map[string]string, len(controls))
	for _, control := range controls {
		index[gitLabPaginationCanonicalName(control)] = control
	}
	return index
}

func gitLabDirectReadCandidate(semantics map[string]any) bool {
	switch stringAt(semantics, "state") {
	case "source_safe_get_read", "source_semantic_head_read", "source_semantic_post_read":
		return true
	default:
		return false
	}
}

func gitLabFactsOperationSemantics(facts map[string]any) map[string]any {
	if semantics, ok := facts["operation_semantics"].(map[string]any); ok {
		return semantics
	}
	if strings.ToUpper(stringAt(facts, "method")) == "GET" {
		return map[string]any{"state": "source_safe_get_read"}
	}
	return map[string]any{"state": "not_a_documented_bounded_read"}
}

func gitLabBinaryRequestFields(descriptor map[string]any) []any {
	request := objectAt(descriptor, "request")
	body, exists := request["body"]
	if !exists {
		return []any{}
	}
	fields := make([]string, 0)
	gitLabFindBinaryFields(objectAt(mustGitLabObjectNoTest(body), "schema"), "body", &fields)
	sort.Strings(fields)
	return stringsToAny(fields)
}

func gitLabFindBinaryFields(schema map[string]any, path string, fields *[]string) {
	if stringAt(schema, "format") == "binary" {
		*fields = append(*fields, path)
	}
	if properties, ok := schema["properties"].(map[string]any); ok {
		keys := make([]string, 0, len(properties))
		for name := range properties {
			keys = append(keys, name)
		}
		sort.Strings(keys)
		for _, name := range keys {
			child, ok := properties[name].(map[string]any)
			if ok {
				gitLabFindBinaryFields(child, path+"."+name, fields)
			}
		}
	}
	if items, ok := schema["items"].(map[string]any); ok {
		gitLabFindBinaryFields(items, path+"[]", fields)
	}
	for _, key := range []string{"oneOf", "anyOf", "allOf"} {
		if values, ok := schema[key].([]any); ok {
			for _, value := range values {
				if child, ok := value.(map[string]any); ok {
					gitLabFindBinaryFields(child, path, fields)
				}
			}
		}
	}
}

func gitLabEventFacts(descriptor map[string]any) map[string]any {
	if !strings.HasSuffix(stringAt(descriptor, "path"), "/hooks") {
		return map[string]any{"state": "not_documented"}
	}
	request := objectAt(descriptor, "request")
	body, exists := request["body"]
	if !exists {
		return map[string]any{"state": "not_documented"}
	}
	bodyMap := mustGitLabObjectNoTest(body)
	schema := objectAt(bodyMap, "schema")
	required := make(map[string]struct{})
	for _, name := range stringSlice(schema["required"]) {
		required[name] = struct{}{}
	}
	if _, urlRequired := required["url"]; !urlRequired {
		return map[string]any{"state": "not_documented"}
	}
	eventFields := make([]string, 0)
	for name := range objectAt(schema, "properties") {
		if strings.HasSuffix(name, "_events") {
			eventFields = append(eventFields, name)
		}
	}
	if len(eventFields) == 0 {
		return map[string]any{"state": "not_documented"}
	}
	sort.Strings(eventFields)
	return map[string]any{"state": "webhook_registration", "url_required": true, "event_selectors": stringsToAny(eventFields)}
}

func gitLabSupplementalOperations(binaryLock map[string]any) map[string]map[string]any {
	result := make(map[string]map[string]any, 2)
	for _, document := range mustMapSlice(objectAt(binaryLock, "rest")["source_documents"]) {
		artifact := objectAt(document, "artifact")
		for _, rawOperation := range mustMapSlice(document["operations"]) {
			operation := cloneGitLabMap(rawOperation)
			operation["document_sha256"] = artifact["sha256"]
			operation["document_bytes"] = artifact["bytes"]
			id := stringAt(operation, "id")
			if _, exists := result[id]; exists {
				panic("duplicate supplemental GitLab source ID")
			}
			result[id] = operation
		}
	}
	return result
}

func sourceParameterNames(parameters []map[string]any, location string) []any {
	names := make([]string, 0)
	for _, parameter := range parameters {
		if stringAt(parameter, "in") == location {
			names = append(names, stringAt(parameter, "name"))
		}
	}
	sort.Strings(names)
	return stringsToAny(names)
}

func sourceRequestMediaTypes(operation map[string]any) []any {
	body, exists := objectAt(operation, "source_operation")["requestBody"]
	if !exists {
		return []any{}
	}
	return sortedObjectKeys(objectAt(mustGitLabObjectNoTest(body), "content"))
}

func sourceSuccessResponseMediaTypes(operation map[string]any) []any {
	media := make(map[string]struct{})
	responses, _ := objectAt(operation, "source_operation")["responses"].(map[string]any)
	for status, rawResponse := range responses {
		if !strings.HasPrefix(status, "2") {
			continue
		}
		response := mustGitLabObjectNoTest(rawResponse)
		content, _ := response["content"].(map[string]any)
		for _, mediaType := range sortedObjectKeys(content) {
			media[mediaType.(string)] = struct{}{}
		}
	}
	keys := make([]string, 0, len(media))
	for mediaType := range media {
		keys = append(keys, mediaType)
	}
	sort.Strings(keys)
	return stringsToAny(keys)
}

func gitLabRestrictionIDs(sourceID string) []any {
	for _, spec := range gitLabPathRestrictionSpecs {
		if spec["source_id"] == sourceID {
			return []any{spec["record_id"]}
		}
	}
	return []any{}
}

func pathTemplateParameters(path string) []any {
	params := make([]string, 0)
	for remaining := path; ; {
		open := strings.IndexByte(remaining, '{')
		if open < 0 {
			break
		}
		remaining = remaining[open+1:]
		close := strings.IndexByte(remaining, '}')
		if close < 0 {
			break
		}
		params = append(params, remaining[:close])
		remaining = remaining[close+1:]
	}
	return stringsToAny(params)
}

func gitLabIsMutation(method string) bool {
	return method == "POST" || method == "PUT" || method == "PATCH" || method == "DELETE"
}

func gitLabMutationCandidate(method string, semantics map[string]any) bool {
	return gitLabIsMutation(method) && stringAt(semantics, "state") != "source_semantic_post_read"
}

func gitLabMatrixRow(t *testing.T, matrix map[string]any, sourceID string) map[string]any {
	t.Helper()
	for _, row := range mustGitLabArray(t, matrix["source_operations"]) {
		object := mustGitLabObject(t, row)
		if stringAt(object, "source_id") == sourceID {
			return object
		}
	}
	t.Fatalf("matrix source row %q not found", sourceID)
	return nil
}

func gitLabRawOperationIDsForRows(t *testing.T, matrix map[string]any, sourceIDs map[string]struct{}) map[string]struct{} {
	t.Helper()
	rawOperationIDs := make(map[string]struct{}, len(sourceIDs))
	for sourceID := range sourceIDs {
		facts := objectAt(gitLabMatrixRow(t, matrix, sourceID), "source_facts")
		rawOperationID := stringAt(facts, "operation_id")
		if rawOperationID == "" {
			t.Fatalf("matrix source %q has no raw source_facts.operation_id", sourceID)
		}
		rawOperationIDs[rawOperationID] = struct{}{}
	}
	if len(rawOperationIDs) != len(sourceIDs) {
		t.Fatalf("matrix source IDs=%v normalized to raw IDs=%v, want one raw source_facts.operation_id per source", sortedGitLabSet(sourceIDs), sortedGitLabSet(rawOperationIDs))
	}
	return rawOperationIDs
}

func gitLabNoImplementedDirectReadArtifactError(cli, operations map[string]any, rawOperationIDs map[string]struct{}) error {
	for _, raw := range mustMapSlice(cli["commands"]) {
		sourceOperationID := stringAt(raw, "source_operation")
		if _, selected := rawOperationIDs[sourceOperationID]; selected && stringAt(raw, "intent") == "direct_read" && stringAt(raw, "availability") == "implemented" {
			return fmt.Errorf("implemented CLI direct-read for source operation %q", sourceOperationID)
		}
	}
	for _, raw := range mustMapSlice(operations["operations"]) {
		sourceOperationID := stringAt(objectAt(raw, "source_operation"), "id")
		if _, selected := rawOperationIDs[sourceOperationID]; selected && stringAt(raw, "kind") == "rest_read" {
			return fmt.Errorf("declared rest-read for source operation %q", sourceOperationID)
		}
	}
	return nil
}

func gitLabMatchingCLIArtifactCount(cli map[string]any, rawOperationIDs map[string]struct{}) int {
	count := 0
	for _, raw := range mustMapSlice(cli["commands"]) {
		if _, selected := rawOperationIDs[stringAt(raw, "source_operation")]; selected {
			count++
		}
	}
	return count
}

func loadGitLabObject(t *testing.T, path string) map[string]any {
	t.Helper()
	value, err := readGitLabObject(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return value
}

func readGitLabObject(path string) (map[string]any, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var value map[string]any
	if err := json.Unmarshal(contents, &value); err != nil {
		return nil, err
	}
	return value, nil
}

func cloneGitLabObject(t *testing.T, value map[string]any) map[string]any {
	t.Helper()
	contents, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal clone: %v", err)
	}
	var clone map[string]any
	if err := json.Unmarshal(contents, &clone); err != nil {
		t.Fatalf("unmarshal clone: %v", err)
	}
	return clone
}

func cloneGitLabMap(value map[string]any) map[string]any {
	contents, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	var clone map[string]any
	if err := json.Unmarshal(contents, &clone); err != nil {
		panic(err)
	}
	return clone
}

func mustGitLabArray(t *testing.T, value any) []any {
	t.Helper()
	array, ok := value.([]any)
	if !ok {
		t.Fatalf("array type %T", value)
	}
	return array
}

func mustMapSlice(value any) []map[string]any {
	array, ok := value.([]any)
	if !ok {
		panic(fmt.Sprintf("array type %T", value))
	}
	result := make([]map[string]any, len(array))
	for i, item := range array {
		object, ok := item.(map[string]any)
		if !ok {
			panic(fmt.Sprintf("array object type %T", item))
		}
		result[i] = object
	}
	return result
}

func mapSliceOrEmpty(value any) []map[string]any {
	if value == nil {
		return []map[string]any{}
	}
	return mustMapSlice(value)
}

func mustGitLabObject(t *testing.T, value any) map[string]any {
	t.Helper()
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("object type %T", value)
	}
	return object
}

func mustGitLabObjectNoTest(value any) map[string]any {
	object, ok := value.(map[string]any)
	if !ok {
		panic(fmt.Sprintf("object type %T", value))
	}
	return object
}

func objectAt(value map[string]any, key string) map[string]any {
	object, ok := value[key].(map[string]any)
	if !ok {
		panic(fmt.Sprintf("object %q type %T", key, value[key]))
	}
	return object
}

func stringAt(value map[string]any, key string) string {
	if value[key] == nil {
		return ""
	}
	stringValue, ok := value[key].(string)
	if !ok {
		panic(fmt.Sprintf("string %q type %T", key, value[key]))
	}
	return stringValue
}

func numberAt(value map[string]any, key string) int {
	number, ok := value[key].(float64)
	if !ok {
		panic(fmt.Sprintf("number %q type %T", key, value[key]))
	}
	return int(number)
}

func stringSlice(value any) []string {
	array, ok := value.([]any)
	if !ok {
		panic(fmt.Sprintf("string array type %T", value))
	}
	result := make([]string, len(array))
	for i, item := range array {
		stringValue, ok := item.(string)
		if !ok {
			panic(fmt.Sprintf("string array item type %T", item))
		}
		result[i] = stringValue
	}
	return result
}

func stringSliceOrEmpty(value any) []string {
	if value == nil {
		return []string{}
	}
	return stringSlice(value)
}

func containsGitLabString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func sortedGitLabSet(values map[string]struct{}) []string {
	keys := make([]string, 0, len(values))
	for value := range values {
		keys = append(keys, value)
	}
	sort.Strings(keys)
	return keys
}

func stringsToAny(values []string) []any {
	result := make([]any, len(values))
	for i, value := range values {
		result[i] = value
	}
	return result
}

func sortedObjectKeys(value map[string]any) []any {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return stringsToAny(keys)
}

func numberMap(value any) map[string]int {
	object := mustGitLabObjectNoTest(value)
	result := make(map[string]int, len(object))
	for key, raw := range object {
		number, ok := raw.(float64)
		if !ok {
			panic(fmt.Sprintf("number map %q type %T", key, raw))
		}
		result[key] = int(number)
	}
	return result
}

func equalGitLabCounts(got, want map[string]int) bool {
	return reflect.DeepEqual(got, want)
}

func sameGitLabJSON(got, want any) bool {
	return reflect.DeepEqual(got, want)
}

func gitLabBlobSHA1(contents []byte) string {
	prefix := []byte(fmt.Sprintf("blob %d\x00", len(contents)))
	digest := sha1.Sum(append(prefix, contents...))
	return hex.EncodeToString(digest[:])
}
