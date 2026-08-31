package gitlab

import (
	"strings"
	"testing"
)

// gitLabBodylessPOSTReadSources is deliberately an exact, source-operation
// cohort rather than a verb rule. Each retained operation has a documented
// successful response and explicitly no source request body; the closed
// rest.no_request_body contract is what permits its bounded POST read.
var gitLabBodylessPOSTReadSources = map[string]struct {
	path string
}{
	"postApiV4AiThirdPartyAgentsDirectAccess": {
		path: "/ai/third_party_agents/direct_access",
	},
	"postApiV4CodeSuggestionsConnectionDetails": {
		path: "/code_suggestions/connection_details",
	},
	"postApiV4GeoNodeProxyIdGraphql": {
		path: "/geo/node_proxy/{id}/graphql",
	},
	"postApiV4IntegrationsSlackOptions": {
		path: "/integrations/slack/options",
	},
	"postApiV4PackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelPackagesConanPackageReferenceUploadUrls": {
		path: "/packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/packages/{conan_package_reference}/upload_urls",
	},
	"postApiV4PackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelUploadUrls": {
		path: "/packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/upload_urls",
	},
	"postApiV4ProjectsIdPackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelPackagesConanPackageReferenceUploadUrls": {
		path: "/projects/{id}/packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/packages/{conan_package_reference}/upload_urls",
	},
	"postApiV4ProjectsIdPackagesConanV1ConansPackageNamePackageVersionPackageUsernamePackageChannelUploadUrls": {
		path: "/projects/{id}/packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/upload_urls",
	},
}

// gitLabTypedAliasReadSources keeps the first materialized alias cohort
// small and source-backed. The selected alias in each operation is scalar and
// has the engine's closed one-bracket-segment spelling; unsupported arrays and
// objects remain absent rather than gaining a lossy CLI encoding.
var gitLabTypedAliasReadSources = map[string]struct {
	path     string
	aliasKey string
}{
	"getApiV4AnalyticsCodeReview":                            {path: "/analytics/code_review", aliasKey: "not[milestone_title]"},
	"getApiV4GroupsIdDashEpics":                              {path: "/groups/{id}/-/epics", aliasKey: "not[author_id]"},
	"getApiV4GroupsIdEpics":                                  {path: "/groups/{id}/epics", aliasKey: "not[author_id]"},
	"getApiV4GroupsIdIssues":                                 {path: "/groups/{id}/issues", aliasKey: "not[milestone]"},
	"getApiV4GroupsIdIssuesStatistics":                       {path: "/groups/{id}/issues_statistics", aliasKey: "not[milestone]"},
	"getApiV4GroupsIdMergeRequests":                          {path: "/groups/{id}/merge_requests", aliasKey: "not[author_id]"},
	"getApiV4IssuesStatistics":                               {path: "/issues_statistics", aliasKey: "not[milestone]"},
	"getApiV4MergeRequests":                                  {path: "/merge_requests", aliasKey: "not[author_id]"},
	"getApiV4ProjectsIdDeploymentsDeploymentIdMergeRequests": {path: "/projects/{id}/deployments/{deployment_id}/merge_requests", aliasKey: "not[author_id]"},
	"getApiV4ProjectsIdIssues":                               {path: "/projects/{id}/issues", aliasKey: "not[milestone]"},
	"getApiV4ProjectsIdIssuesStatistics":                     {path: "/projects/{id}/issues_statistics", aliasKey: "not[milestone]"},
	"getApiV4ProjectsIdMergeRequests":                        {path: "/projects/{id}/merge_requests", aliasKey: "not[author_id]"},
	"getApiV4ProjectsIdRepositoryFilesFilePathBlame":         {path: "/projects/{id}/repository/files/{file_path}/blame", aliasKey: "range[start]"},
	"getApiV4ProjectsIdVariablesKey":                         {path: "/projects/{id}/variables/{key}", aliasKey: "filter[environment_scope]"},
}

func TestGitLabSourceBoundMaterializationCohort(t *testing.T) {
	matrix := loadGitLabObject(t, gitLabSourceLaneMatrixPath)
	operations := loadGitLabObject(t, "operations.json")
	cli := loadGitLabObject(t, "cli_surface.json")
	api := loadGitLabObject(t, "api_surface.json")

	for sourceID, want := range gitLabTypedAliasReadSources {
		gitLabAssertMaterializedDirectRead(t, matrix, operations, cli, api, sourceID, "GET", want.path, want.aliasKey, false)
	}
	for sourceID, want := range gitLabBodylessPOSTReadSources {
		gitLabAssertMaterializedDirectRead(t, matrix, operations, cli, api, sourceID, "POST", want.path, "", true)
	}

	const statusSourceID = "headApiV4ProjectsIdRepositoryBranchesBranch"
	const statusOperationID = "source_status_48454144202f70726f6a656374732f7b69647d2f7265706f7369746f72792f6272616e636865732f7b6272616e63687d"
	status := gitLabOperationByID(t, operations, statusOperationID)
	if stringAt(status, "kind") != "rest_status" || stringAt(objectAt(status, "rest"), "method") != "HEAD" || stringAt(status, "output_policy") != "status" {
		t.Fatalf("GitLab HEAD source %q operation = %#v, want a closed rest_status declaration", statusSourceID, status)
	}
	statusCommand := gitLabCommandByOperationID(t, cli, statusOperationID)
	if stringAt(statusCommand, "intent") != "status_check" || stringAt(statusCommand, "availability") != "implemented" || stringAt(statusCommand, "operation") != stringAt(status, "id") {
		t.Fatalf("GitLab HEAD source %q command = %#v, want an implemented status-check binding", statusSourceID, statusCommand)
	}
	if !strings.Contains(stringAt(statusCommand, "notes"), "source_operation="+statusSourceID) {
		t.Fatalf("GitLab HEAD source %q command notes = %q, want exact retained source ID", statusSourceID, stringAt(statusCommand, "notes"))
	}
	gitLabAssertMaterializedMatrixCell(t, matrix, "gitlab.rest."+statusSourceID, "direct_read", "status_check", stringAt(statusCommand, "path"))

	const etlSourceID = "getApiV4ProjectsIdMlMlflowApi20MlflowMetricsGetHistory"
	etlCommand := gitLabCommandBySourceIDAndIntent(t, cli, etlSourceID, "etl")
	if stringAt(etlCommand, "intent") != "etl" || stringAt(etlCommand, "availability") != "implemented" || stringAt(etlCommand, "stream") != "mlflow_metric_history" {
		t.Fatalf("GitLab MLflow ETL command = %#v, want the exact declared full-refresh stream", etlCommand)
	}
	gitLabAssertMaterializedMatrixCell(t, matrix, "gitlab.rest."+etlSourceID, "etl", "stream", "mlflow_metric_history")
	gitLabAssertMLflowSourceTransportBinding(t, etlSourceID)

	// The five source-semantic POST reads with a documented request body must
	// remain mapped-unproven. This is the negative boundary for the new closed
	// bodyless POST form, not a generic POST promotion.
	for _, sourceID := range []string{
		"postApiV4CodeSuggestionsDirectAccess",
		"postApiV4Glql",
		"postApiV4ProjectsIdMlMlflowApi20MlflowExperimentsSearch",
		"postApiV4ProjectsIdMlMlflowApi20MlflowRunsSearch",
		"postApiV4ProjectsIdRepositoryBlobsBatch",
	} {
		row := gitLabMatrixRow(t, matrix, "gitlab.rest."+sourceID)
		if got := stringAt(objectAt(objectAt(row, "lanes"), "direct_read"), "disposition"); got != "mapped_unproven" {
			t.Fatalf("GitLab body-bearing POST %q direct-read disposition = %q, want mapped_unproven", sourceID, got)
		}
	}
}

// gitLabAssertMLflowSourceTransportBinding proves the already-registered
// declarative source transport names this exact documented full-refresh
// stream. It is not a webhook/event sync claim and it does not admit any
// body-bearing POST pagination route.
func gitLabAssertMLflowSourceTransportBinding(t *testing.T, sourceID string) {
	t.Helper()
	transport := loadGitLabObject(t, "sync_transport.json")
	sourceTransport := objectAt(transport, "source_transport")
	if !containsGitLabString(stringSlice(sourceTransport["eligible_streams"]), "mlflow_metric_history") {
		t.Fatalf("GitLab source transport streams=%#v, want source-backed mlflow_metric_history", sourceTransport["eligible_streams"])
	}

	contract := loadGitLabObject(t, "enabled_connector_contract.json")
	for _, raw := range mustGitLabArray(t, contract["lanes"]) {
		lane := mustGitLabObject(t, raw)
		if stringAt(lane, "name") != "sync_transport" {
			continue
		}
		source := objectAt(lane, "source")
		if source["partition"] != false || numberAt(source, "expected") != 5 || numberAt(source, "implemented") != 5 || !containsGitLabString(stringSlice(source["operation_ids"]), sourceID) {
			t.Fatalf("GitLab sync enabled-contract source=%#v, want five exact stream bindings including %q", source, sourceID)
		}
		for _, rawStream := range mustGitLabArray(t, objectAt(lane, "transport")["streams"]) {
			stream := mustGitLabObject(t, rawStream)
			if stringAt(stream, "stream") != "mlflow_metric_history" {
				continue
			}
			if stringAt(stream, "source_operation") != sourceID || stringAt(stream, "cursor_evidence") != "source_cited" || stringAt(stream, "delete_evidence") != "not_declared" || stringAt(stream, "order_evidence") != "not_declared" {
				t.Fatalf("GitLab MLflow source transport evidence=%#v, want exact continuation and no event/delete/order claim", stream)
			}
			return
		}
		t.Fatalf("GitLab sync enabled contract has no mlflow_metric_history stream")
	}
	t.Fatal("GitLab enabled contract has no sync_transport lane")
}

func gitLabAssertMaterializedDirectRead(t *testing.T, matrix, operations, cli, api map[string]any, sourceID, method, path, aliasKey string, bodylessPOST bool) {
	t.Helper()
	op := gitLabOperationBySourceID(t, operations, sourceID)
	rest := objectAt(op, "rest")
	if stringAt(op, "kind") != "rest_read" || stringAt(rest, "method") != method || stringAt(rest, "path") != path {
		t.Fatalf("GitLab source %q operation = %#v, want %s rest_read %s", sourceID, op, method, path)
	}
	if bodylessPOST && rest["no_request_body"] != true {
		t.Fatalf("GitLab source %q rest contract = %#v, want no_request_body", sourceID, rest)
	}
	if !bodylessPOST && rest["no_request_body"] != nil {
		t.Fatalf("GitLab source %q unexpectedly declares no_request_body", sourceID)
	}
	command := gitLabCommandBySourceIDAndIntent(t, cli, sourceID, "direct_read")
	if stringAt(command, "intent") != "direct_read" || stringAt(command, "availability") != "implemented" || stringAt(command, "operation") != stringAt(op, "id") {
		t.Fatalf("GitLab source %q command = %#v, want an implemented direct-read binding", sourceID, command)
	}
	if aliasKey != "" {
		wantMapsTo := "query." + aliasKey
		found := false
		for _, raw := range mustGitLabArray(t, command["flags"]) {
			flag := mustGitLabObject(t, raw)
			if stringAt(flag, "maps_to") == wantMapsTo {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("GitLab source %q flags = %#v, want exact provider wire key %q", sourceID, command["flags"], wantMapsTo)
		}
	}
	gitLabAssertMaterializedMatrixCell(t, matrix, "gitlab.rest."+sourceID, "direct_read", "direct_read", stringAt(command, "path"))
	endpoint := gitLabAPISurfaceEndpoint(t, api, method, path)
	coverage := objectAt(endpoint, "covered_by")
	if stringAt(coverage, "direct_read") != stringAt(command, "path") {
		t.Fatalf("GitLab source %q api coverage = %#v, want direct_read %q", sourceID, coverage, stringAt(command, "path"))
	}
}

func gitLabOperationBySourceID(t *testing.T, operations map[string]any, sourceID string) map[string]any {
	t.Helper()
	for _, raw := range mustGitLabArray(t, operations["operations"]) {
		op := mustGitLabObject(t, raw)
		binding, ok := op["source_operation"].(map[string]any)
		if ok && stringAt(binding, "id") == sourceID {
			return op
		}
	}
	t.Fatalf("GitLab operations.json has no source-bound operation %q", sourceID)
	return nil
}

func gitLabOperationByID(t *testing.T, operations map[string]any, id string) map[string]any {
	t.Helper()
	for _, raw := range mustGitLabArray(t, operations["operations"]) {
		op := mustGitLabObject(t, raw)
		if stringAt(op, "id") == id {
			return op
		}
	}
	t.Fatalf("GitLab operations.json has no operation %q", id)
	return nil
}

func gitLabCommandBySourceIDAndIntent(t *testing.T, cli map[string]any, sourceID, intent string) map[string]any {
	t.Helper()
	for _, raw := range mustGitLabArray(t, cli["commands"]) {
		command := mustGitLabObject(t, raw)
		if stringAt(command, "source_operation") == sourceID && stringAt(command, "intent") == intent {
			return command
		}
	}
	t.Fatalf("GitLab cli_surface.json has no source-bound %s command %q", intent, sourceID)
	return nil
}

func gitLabCommandByOperationID(t *testing.T, cli map[string]any, operationID string) map[string]any {
	t.Helper()
	for _, raw := range mustGitLabArray(t, cli["commands"]) {
		command := mustGitLabObject(t, raw)
		if stringAt(command, "operation") == operationID {
			return command
		}
	}
	t.Fatalf("GitLab cli_surface.json has no command for operation %q", operationID)
	return nil
}

func gitLabAPISurfaceEndpoint(t *testing.T, api map[string]any, method, path string) map[string]any {
	t.Helper()
	for _, raw := range mustGitLabArray(t, api["endpoints"]) {
		endpoint := mustGitLabObject(t, raw)
		if strings.EqualFold(stringAt(endpoint, "method"), method) && stringAt(endpoint, "path") == path {
			return endpoint
		}
	}
	t.Fatalf("GitLab api_surface has no %s %s endpoint", method, path)
	return nil
}

func gitLabAssertMaterializedMatrixCell(t *testing.T, matrix map[string]any, sourceID, lane, kind, target string) {
	t.Helper()
	row := gitLabMatrixRow(t, matrix, sourceID)
	cell := objectAt(objectAt(row, "lanes"), lane)
	if stringAt(cell, "applicability") != "source_candidate" || stringAt(cell, "disposition") != "implemented" {
		t.Fatalf("GitLab matrix %s %s = %#v, want implemented source candidate", sourceID, lane, cell)
	}
	backlink := objectAt(objectAt(cell, "mapping"), "definition_backlink")
	if stringAt(backlink, "kind") != kind || stringAt(backlink, "target") != target {
		t.Fatalf("GitLab matrix %s %s backlink = %#v, want %s %q", sourceID, lane, backlink, kind, target)
	}
}
