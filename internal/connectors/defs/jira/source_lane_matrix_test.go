package jira

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"
)

const (
	jiraSourceLaneMatrixPath = "sources/jira-source-lane-matrix.json"
	jiraSourceLockPath       = "sources/jira-operation-source-lock.json"
	jiraCrosswalkPath        = "sources/jira-operation-crosswalk.json"
	jiraStreamsPath          = "streams.json"
)

var jiraSourceLaneNames = []string{
	"direct_read",
	"direct_write",
	"binary_download",
	"binary_upload",
	"etl",
	"reverse_etl",
	"sync_transport",
}

var jiraExpectedLaneCounts = map[string]map[string]int{
	"direct_read":     {"mapped_unproven": 309, "not_applicable": 308},
	"direct_write":    {"mapped_unproven": 308, "not_applicable": 309},
	"binary_download": {"mapped_unproven": 3, "not_applicable": 614},
	"binary_upload":   {"mapped_unproven": 4, "not_applicable": 613},
	"etl":             {"mapped_unproven": 103, "not_applicable": 514},
	"reverse_etl":     {"mapped_unproven": 308, "not_applicable": 309},
	"sync_transport":  {"missing_foundation": 1, "not_applicable": 616},
}

// jiraDocumentedReadSemantics is an intentionally small, source-cited
// exception inventory for Jira's POST operations that evaluate, validate,
// search, preview, or bulk-fetch provider state without mutating it. It is not
// a keyword classifier: each row binds an exact retained source operation to
// the source text that establishes the read behavior. The ordinary REST method
// remains a provenance fact, but cannot by itself turn these reads into writes.
var jiraDocumentedReadSemantics = map[string]jiraSourceTextEvidence{
	"jira.rest.getCustomFieldsConfigurations":                  {Contains: "Returns a [paginated](#pagination) list of configurations"},
	"jira.rest.getBulkChangelogs":                              {Contains: "Bulk fetch changelogs for multiple issues"},
	"jira.rest.getCommentsByIds":                               {Contains: "Returns a [paginated](#pagination) list of comments"},
	"jira.rest.analyseExpression":                              {Contains: "Analyses and validates Jira expressions."},
	"jira.rest.evaluateJiraExpression":                         {Contains: "Evaluates a Jira expression and returns its value."},
	"jira.rest.evaluateJSISJiraExpression":                     {Contains: "Evaluates a Jira expression and returns its value."},
	"jira.rest.getCustomFieldContextsForProjectsAndIssueTypes": {Contains: "Returns a [paginated](#pagination) list of project and issue type mappings"},
	"jira.rest.bulkFetchIssues":                                {Contains: "Returns the details for a set of requested issues."},
	"jira.rest.getIsWatchingIssueBulk":                         {Contains: "Returns, for the user, details of the watched status"},
	"jira.rest.getChangeLogsByIds":                             {Contains: "Returns changelogs for an issue specified by a list of changelog IDs."},
	"jira.rest.getAutoCompletePost":                            {Contains: "Returns reference data for JQL searches."},
	"jira.rest.getPrecomputationsByID":                         {Contains: "Returns function precomputations by IDs"},
	"jira.rest.matchIssues":                                    {Contains: "Checks whether one or more issues would be returned"},
	"jira.rest.parseJqlQueries":                                {Contains: "Parses and validates JQL queries."},
	"jira.rest.migrateQueries":                                 {Contains: "Converts one or more JQL queries"},
	"jira.rest.sanitiseJqlQueries":                             {Contains: "Sanitizes one or more JQL queries"},
	"jira.rest.getBulkPermissions":                             {Contains: "Returns:"},
	"jira.rest.getPermittedProjects":                           {Contains: "Returns all the projects where the user is granted"},
	"jira.rest.suggestedPrioritiesForMappings":                 {Contains: "Returns a [paginated](#pagination) list of priorities"},
	"jira.rest.searchForIssuesUsingJqlPost":                    {Contains: "Searches for issues using [JQL]"},
	"jira.rest.countIssues":                                    {Contains: "Provide an estimated count of the issues"},
	"jira.rest.searchAndReconsileIssuesUsingJqlPost":           {Contains: "Searches for issues using [JQL]"},
	"jira.rest.readWorkflowFromHistory":                        {Contains: "Returns a workflow and related statuses"},
	"jira.rest.listWorkflowHistory":                            {Contains: "Returns a list of workflow history entries"},
	"jira.rest.readWorkflows":                                  {Contains: "Returns a list of workflows and related statuses"},
	"jira.rest.validateCreateWorkflows":                        {Contains: "Validate the payload for bulk create workflows."},
	"jira.rest.readWorkflowPreviews":                           {Contains: "Returns a requested workflow within a given project."},
	"jira.rest.readWorkflowSchemes":                            {Contains: "Returns a list of workflow schemes"},
	"jira.rest.validateUpdateWorkflows":                        {Contains: "Validate the payload for bulk update workflows."},
	"jira.rest.getRequiredWorkflowSchemeMappings":              {Contains: "Gets the required status mappings"},
	"jira.rest.getWorklogsForIds":                              {Contains: "Returns worklog details for a list of worklog IDs."},
	"jira.rest.MigrationResource.workflowRuleSearch_post":      {Contains: "Returns configurations for workflow transition rules"},
	"jira.rest.getWorklogsByIssueIdAndWorklogId":               {Contains: "Returns worklog details for a list of issue ID and worklog ID pairs."},
}

// jiraBodyPaginationEvidence covers source operations whose retained request
// example, response example, and summary together document offset pagination
// in a JSON body. They are deliberately named source facts rather than an
// unbounded scan for field names in arbitrary request bodies.
var jiraBodyPaginationEvidence = map[string]jiraSourceTextEvidence{
	"jira.rest.searchForIssuesUsingJqlPost":    {Contains: "\"maxResults\"", ContinuationKind: "body_offset"},
	"jira.rest.suggestedPrioritiesForMappings": {Contains: "\"maxResults\"", ContinuationKind: "body_offset"},
	"jira.rest.evaluateJSISJiraExpression":     {Contains: "\"nextPageToken\"", ContinuationKind: "body_cursor"},
}

type jiraSourceTextEvidence struct {
	Contains         string
	ContinuationKind string
}

// These are source-text and media predicates only. They identify candidate
// rows for Track A accounting, never an admitted executable binary action.
var jiraBinaryDownloadEvidence = map[string]jiraBinaryEvidence{
	"jira.rest.getAvatarImageByType": {
		SuccessMedia: []string{"image/png", "image/svg+xml"},
		Signals:      []string{"success_response_media:image/png", "success_response_media:image/svg+xml"},
	},
	"jira.rest.getAvatarImageByID": {
		SuccessMedia: []string{"image/png", "image/svg+xml"},
		Signals:      []string{"success_response_media:image/png", "success_response_media:image/svg+xml"},
	},
	"jira.rest.getAvatarImageByOwner": {
		SuccessMedia: []string{"image/png", "image/svg+xml"},
		Signals:      []string{"success_response_media:image/png", "success_response_media:image/svg+xml"},
	},
}

var jiraBinaryUploadEvidence = map[string]jiraBinaryEvidence{
	"jira.rest.addAttachment": {
		RequestMedia: []string{"multipart/form-data"},
		Signals:      []string{"request_body_media:multipart/form-data"},
	},
	"jira.rest.createIssueTypeAvatar": {
		RequestMedia:       []string{"*/*"},
		SourceTextContains: "image isn't included",
		Signals:            []string{"request_body_media:*/*", "source_text:image isn't included"},
	},
	"jira.rest.createProjectAvatar": {
		RequestMedia:       []string{"*/*"},
		SourceTextContains: "image isn't included",
		Signals:            []string{"request_body_media:*/*", "source_text:image isn't included"},
	},
	"jira.rest.storeAvatar": {
		RequestMedia:       []string{"*/*"},
		SourceTextContains: "image isn't included",
		Signals:            []string{"request_body_media:*/*", "source_text:image isn't included"},
	},
}

var jiraLegacyETLStreams = map[string]jiraLegacyETLStream{
	"jira.rest.searchForIssuesUsingJql": {
		SourceID: "jira.rest.searchForIssuesUsingJql",
		Stream:   "issues",
		Path:     "/rest/api/3/search",
	},
	"jira.rest.searchProjects": {
		SourceID: "jira.rest.searchProjects",
		Stream:   "projects",
		Path:     "/rest/api/3/project/search",
	},
	"jira.rest.getAllUsers": {
		SourceID: "jira.rest.getAllUsers",
		Stream:   "users",
		Path:     "/rest/api/3/users/search",
	},
}

var jiraEventCursorStates = map[string]string{
	"jira.rest.registerDynamicWebhooks":  "webhook_registration",
	"jira.rest.deleteWebhookById":        "webhook_lifecycle_control",
	"jira.rest.refreshWebhooks":          "webhook_lifecycle_control",
	"jira.rest.getDynamicWebhooksForApp": "webhook_registration_catalog",
	"jira.rest.getFailedWebhooks":        "webhook_failed_delivery_cursor",
}

type jiraBinaryEvidence struct {
	RequestMedia       []string
	SuccessMedia       []string
	SourceTextContains string
	Signals            []string
}

type jiraSourceLaneMatrix struct {
	SchemaVersion                int                              `json:"schema_version"`
	Connector                    string                           `json:"connector"`
	Lanes                        []string                         `json:"lanes"`
	SourceLock                   jiraSourceLaneMatrixSourceLock   `json:"source_lock"`
	SourceBoundaryReconciliation jiraSourceBoundaryReconciliation `json:"source_boundary_reconciliation"`
	LegacyETLReconciliation      jiraLegacyETLReconciliation      `json:"legacy_etl_reconciliation"`
	SourceOperations             []jiraSourceLaneMatrixRow        `json:"source_operations"`
	Summary                      jiraSourceLaneMatrixSummary      `json:"summary"`
}

type jiraSourceLaneMatrixSourceLock struct {
	Path           string `json:"path"`
	SchemaVersion  int    `json:"schema_version"`
	Connector      string `json:"connector"`
	SourceDocument struct {
		SourceURL      string `json:"source_url"`
		SHA256         string `json:"sha256"`
		Bytes          int    `json:"bytes"`
		OperationCount int    `json:"operation_count"`
	} `json:"source_document"`
}

type jiraSourceBoundaryReconciliation struct {
	Identity          string `json:"identity"`
	SourceLockRows    int    `json:"source_lock_rows"`
	CrosswalkRows     int    `json:"crosswalk_rows"`
	CrosswalkOnlyRows int    `json:"crosswalk_only_rows"`
	LockOnlyRows      int    `json:"lock_only_rows"`
}

type jiraLegacyETLReconciliation struct {
	SourcePagingCandidateCriterion string                `json:"source_paging_candidate_criterion"`
	SourcePagingCandidates         int                   `json:"source_paging_candidates"`
	LegacyStreamBacklinks          []jiraLegacyETLStream `json:"legacy_stream_backlinks"`
	RemainingPagingCandidates      int                   `json:"remaining_paging_candidates"`
}

type jiraLegacyETLStream struct {
	SourceID string `json:"source_id"`
	Stream   string `json:"stream"`
	Path     string `json:"path"`
}

type jiraSourceLaneMatrixSummary struct {
	SourceRows             int                       `json:"source_rows"`
	SourceRowsWithAllLanes int                       `json:"source_rows_with_all_lanes"`
	LaneCounts             map[string]map[string]int `json:"lane_counts"`
}

type jiraSourceLaneMatrixRow struct {
	SourceID    string                              `json:"source_id"`
	SourceFacts jiraSourceLaneMatrixSourceFacts     `json:"source_facts"`
	Lanes       map[string]jiraSourceLaneMatrixCell `json:"lanes"`
}

type jiraSourceLaneMatrixSourceFacts struct {
	Protocol    string `json:"protocol"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	OperationID string `json:"operation_id"`
	Deprecated  *bool  `json:"deprecated"`
	Citation    struct {
		SourceLocation string `json:"source_location"`
	} `json:"citation"`
	ScopeAndFanout struct {
		PathVariables   []string `json:"path_variables"`
		QueryParameters []string `json:"query_parameters"`
		Fanout          struct {
			State string `json:"state"`
		} `json:"fanout"`
	} `json:"scope_and_fanout"`
	Media struct {
		RequestMediaTypes         []string `json:"request_media_types"`
		SuccessResponseMediaTypes []string `json:"success_response_media_types"`
		BinarySignals             []string `json:"binary_signals"`
	} `json:"media"`
	Pagination struct {
		State                 string   `json:"state"`
		PagingQueryParameters []string `json:"paging_query_parameters"`
	} `json:"pagination"`
	EventCursor struct {
		State string `json:"state"`
	} `json:"event_cursor"`
	OperationSemantics struct {
		State string `json:"state"`
	} `json:"operation_semantics"`
}

type jiraSourceLaneMatrixCell struct {
	Applicability string          `json:"applicability"`
	Disposition   string          `json:"disposition"`
	Reason        string          `json:"reason"`
	Mapping       json.RawMessage `json:"mapping,omitempty"`
}

type jiraSourceLaneLock struct {
	SchemaVersion int    `json:"schema_version"`
	Connector     string `json:"connector"`
	Counts        struct {
		Total int `json:"total"`
	} `json:"counts"`
	REST struct {
		SourceURL  string                      `json:"source_url"`
		SHA256     string                      `json:"sha256"`
		Bytes      int                         `json:"bytes"`
		Operations []jiraLockedSourceOperation `json:"operations"`
	} `json:"rest"`
}

type jiraLockedSourceOperation struct {
	ID              string                      `json:"id"`
	Protocol        string                      `json:"protocol"`
	Method          string                      `json:"method"`
	Path            string                      `json:"path"`
	OperationID     string                      `json:"operation_id"`
	Deprecated      *bool                       `json:"deprecated"`
	SourceLocation  string                      `json:"source_location"`
	SourceOperation jiraLockedProviderOperation `json:"source_operation"`
}

type jiraLockedProviderOperation struct {
	Summary     string                        `json:"summary"`
	Description string                        `json:"description"`
	Parameters  []jiraLockedSourceParameter   `json:"parameters"`
	RequestBody *jiraLockedRequestBody        `json:"requestBody"`
	Responses   map[string]jiraLockedResponse `json:"responses"`
}

type jiraLockedSourceParameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description"`
}

type jiraLockedRequestBody struct {
	Content map[string]json.RawMessage `json:"content"`
}

type jiraLockedResponse struct {
	Description string                     `json:"description"`
	Content     map[string]json.RawMessage `json:"content"`
}

type jiraSourceLaneCrosswalk struct {
	Connector  string `json:"connector"`
	SourceLock string `json:"source_lock"`
	Accounting struct {
		SourceOperations           int `json:"source_operations"`
		SourceUniqueMethodPath     int `json:"source_unique_method_path"`
		APISurfaceEndpoints        int `json:"api_surface_endpoints"`
		APISurfaceUniqueMethodPath int `json:"api_surface_unique_method_path"`
		ExactSourceToSurface       int `json:"exact_source_to_surface"`
		SourceOnly                 int `json:"source_only"`
		SurfaceOnly                int `json:"surface_only"`
	} `json:"accounting"`
	SourceOperations []jiraCrosswalkSourceOperation `json:"source_operations"`
}

type jiraCrosswalkSourceOperation struct {
	SourceID       string `json:"source_id"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	SourceLocation string `json:"source_location"`
}

type jiraStreamsDefinition struct {
	Streams []jiraLegacyETLStreamDefinition `json:"streams"`
}

type jiraLegacyETLStreamDefinition struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func TestJiraSourceLaneMatrixRetainsEveryLockedOperationAndLane(t *testing.T) {
	matrix := loadJiraSourceLaneMatrix(t)
	lock := loadJiraSourceLaneLock(t)
	crosswalk := loadJiraSourceLaneCrosswalk(t)
	streams := loadJiraStreamsDefinition(t)
	if err := validateJiraSourceLaneMatrix(matrix, lock, crosswalk, streams); err != nil {
		t.Fatalf("validate Jira source lane matrix: %v", err)
	}

	t.Run("rejects missing lane cell", func(t *testing.T) {
		broken := cloneJiraSourceLaneMatrix(t, matrix)
		delete(broken.SourceOperations[0].Lanes, "sync_transport")
		if err := validateJiraSourceLaneMatrix(broken, lock, crosswalk, streams); err == nil || !strings.Contains(err.Error(), "missing lane cell") {
			t.Fatalf("missing-cell validation error = %v, want missing lane cell", err)
		}
	})

	t.Run("rejects hidden source row", func(t *testing.T) {
		broken := cloneJiraSourceLaneMatrix(t, matrix)
		broken.SourceOperations = broken.SourceOperations[1:]
		if err := validateJiraSourceLaneMatrix(broken, lock, crosswalk, streams); err == nil || !strings.Contains(err.Error(), "source rows matrix=") {
			t.Fatalf("hidden-row validation error = %v, want source rows matrix", err)
		}
	})

	t.Run("rejects legacy ETL backlink drift", func(t *testing.T) {
		broken := cloneJiraSourceLaneMatrix(t, matrix)
		row := jiraMatrixRow(t, &broken, "jira.rest.searchForIssuesUsingJql")
		var mapping map[string]any
		if err := json.Unmarshal(row.Lanes["etl"].Mapping, &mapping); err != nil {
			t.Fatal(err)
		}
		mapping["definition_backlink"].(map[string]any)["path"] = "wrong-streams.json"
		row.Lanes["etl"] = jiraSourceLaneMatrixCell{
			Applicability: row.Lanes["etl"].Applicability,
			Disposition:   row.Lanes["etl"].Disposition,
			Reason:        row.Lanes["etl"].Reason,
			Mapping:       mustJiraJSON(t, mapping),
		}
		if err := validateJiraSourceLaneMatrix(broken, lock, crosswalk, streams); err == nil || !strings.Contains(err.Error(), "legacy stream backlink") {
			t.Fatalf("backlink validation error = %v, want legacy stream backlink", err)
		}
	})

	t.Run("rejects invalid executable disposition", func(t *testing.T) {
		broken := cloneJiraSourceLaneMatrix(t, matrix)
		row := jiraMatrixRow(t, &broken, "jira.rest.getBanner")
		cell := row.Lanes["direct_read"]
		cell.Disposition = "implemented"
		row.Lanes["direct_read"] = cell
		if err := validateJiraSourceLaneMatrix(broken, lock, crosswalk, streams); err == nil || !strings.Contains(err.Error(), "lane direct_read") {
			t.Fatalf("invalid-disposition validation error = %v, want lane direct_read", err)
		}
	})
}

func TestJiraSourceLaneMatrixUsesDocumentedSemanticsAndContinuationFacts(t *testing.T) {
	matrix := loadJiraSourceLaneMatrix(t)
	lock := loadJiraSourceLaneLock(t)
	locked := make(map[string]jiraLockedSourceOperation, len(lock.REST.Operations))
	for _, operation := range lock.REST.Operations {
		locked[operation.ID] = operation
	}

	type laneExpectation struct {
		sourceID string
		semantic string
		direct   string
		write    string
		reverse  string
		etl      string
		sync     string
	}
	cases := []laneExpectation{
		// Happy paths: source-documented read operations are not promoted to
		// provider writes merely because the request body uses POST.
		{"jira.rest.getWorklogsForIds", "documented_read_semantics", "applicable", "not_applicable", "not_applicable", "not_applicable", "not_applicable"},
		{"jira.rest.getCustomFieldsConfigurations", "documented_read_semantics", "applicable", "not_applicable", "not_applicable", "applicable", "not_applicable"},
		{"jira.rest.getCustomFieldContextsForProjectsAndIssueTypes", "documented_read_semantics", "applicable", "not_applicable", "not_applicable", "applicable", "not_applicable"},
		{"jira.rest.evaluateJSISJiraExpression", "documented_read_semantics", "applicable", "not_applicable", "not_applicable", "applicable", "not_applicable"},
		// Ordinary source mutations, including deletes, remain independently
		// visible in both bounded direct-write and warehouse reverse-ETL lanes.
		{"jira.rest.bulkSetIssuesPropertiesList", "mutation_verb_candidate", "not_applicable", "applicable", "applicable", "not_applicable", "not_applicable"},
		{"jira.rest.deleteWorklog", "mutation_verb_candidate", "not_applicable", "applicable", "applicable", "not_applicable", "not_applicable"},
		// The continuation variants below deliberately cover cursor, singular
		// maxResult offset, response nextPage, and source-backed POST body offset.
		{"jira.rest.getBulkEditableFields", "read_verb_candidate", "applicable", "not_applicable", "not_applicable", "applicable", "not_applicable"},
		{"jira.rest.getAvailableTransitions", "read_verb_candidate", "applicable", "not_applicable", "not_applicable", "applicable", "not_applicable"},
		{"jira.rest.getBulkScreenTabs", "read_verb_candidate", "applicable", "not_applicable", "not_applicable", "applicable", "not_applicable"},
		{"jira.rest.findUserKeysByQuery", "read_verb_candidate", "applicable", "not_applicable", "not_applicable", "applicable", "not_applicable"},
		{"jira.rest.getIdsOfWorklogsDeletedSince", "read_verb_candidate", "applicable", "not_applicable", "not_applicable", "applicable", "not_applicable"},
		{"jira.rest.getIdsOfWorklogsModifiedSince", "read_verb_candidate", "applicable", "not_applicable", "not_applicable", "applicable", "not_applicable"},
		{"jira.rest.searchForIssuesUsingJqlPost", "documented_read_semantics", "applicable", "not_applicable", "not_applicable", "applicable", "not_applicable"},
		{"jira.rest.suggestedPrioritiesForMappings", "documented_read_semantics", "applicable", "not_applicable", "not_applicable", "applicable", "not_applicable"},
		// Bad/edge cases: a suggestion-list limit or an ordinary array response
		// is not a continuation contract, and list pagination is not sync transport.
		{"jira.rest.findUsersForPicker", "read_verb_candidate", "applicable", "not_applicable", "not_applicable", "not_applicable", "not_applicable"},
		{"jira.rest.getProjectVersions", "read_verb_candidate", "applicable", "not_applicable", "not_applicable", "not_applicable", "not_applicable"},
		{"jira.rest.getDynamicWebhooksForApp", "read_verb_candidate", "applicable", "not_applicable", "not_applicable", "applicable", "not_applicable"},
		// The one source-cited inbound webhook registration remains the only
		// sync gap. It is intentionally distinct from list pagination.
		{"jira.rest.registerDynamicWebhooks", "mutation_verb_candidate", "not_applicable", "applicable", "applicable", "not_applicable", "applicable"},
	}

	for _, tc := range cases {
		t.Run(tc.sourceID, func(t *testing.T) {
			operation, ok := locked[tc.sourceID]
			if !ok {
				t.Fatalf("source operation %q is absent from lock", tc.sourceID)
			}
			if got := jiraOperationSemantics(operation); got != tc.semantic {
				t.Fatalf("semantic classification=%q, want %q", got, tc.semantic)
			}
			row := jiraMatrixRow(t, &matrix, tc.sourceID)
			for lane, want := range map[string]string{
				"direct_read":    tc.direct,
				"direct_write":   tc.write,
				"reverse_etl":    tc.reverse,
				"etl":            tc.etl,
				"sync_transport": tc.sync,
			} {
				if got := row.Lanes[lane].Applicability; got != want {
					t.Errorf("%s applicability=%q, want %q", lane, got, want)
				}
			}
		})
	}

	broken := locked["jira.rest.getWorklogsForIds"]
	broken.SourceOperation.Description = "provider contract text removed"
	if err := validateJiraCandidateEvidence(broken); err == nil || !strings.Contains(err.Error(), "documented POST read semantics") {
		t.Fatalf("missing semantic-source evidence error=%v, want documented POST read semantics", err)
	}
}

func loadJiraSourceLaneMatrix(t *testing.T) jiraSourceLaneMatrix {
	t.Helper()
	return loadJiraJSON[jiraSourceLaneMatrix](t, jiraSourceLaneMatrixPath)
}

func loadJiraSourceLaneLock(t *testing.T) jiraSourceLaneLock {
	t.Helper()
	return loadJiraJSON[jiraSourceLaneLock](t, jiraSourceLockPath)
}

func loadJiraSourceLaneCrosswalk(t *testing.T) jiraSourceLaneCrosswalk {
	t.Helper()
	return loadJiraJSON[jiraSourceLaneCrosswalk](t, jiraCrosswalkPath)
}

func loadJiraStreamsDefinition(t *testing.T) jiraStreamsDefinition {
	t.Helper()
	return loadJiraJSON[jiraStreamsDefinition](t, jiraStreamsPath)
}

func loadJiraJSON[T any](t *testing.T, path string) T {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var decoded T
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return decoded
}

func validateJiraSourceLaneMatrix(matrix jiraSourceLaneMatrix, lock jiraSourceLaneLock, crosswalk jiraSourceLaneCrosswalk, streams jiraStreamsDefinition) error {
	if matrix.SchemaVersion != 1 || matrix.Connector != "jira" {
		return fmt.Errorf("matrix identity schema=%d connector=%q, want schema=1 connector=jira", matrix.SchemaVersion, matrix.Connector)
	}
	if !slices.Equal(matrix.Lanes, jiraSourceLaneNames) {
		return fmt.Errorf("lane order = %v, want %v", matrix.Lanes, jiraSourceLaneNames)
	}
	if err := validateJiraMatrixLockBinding(matrix.SourceLock, lock); err != nil {
		return err
	}

	locked := make(map[string]jiraLockedSourceOperation, len(lock.REST.Operations))
	lockedMethodPaths := make(map[string]struct{}, len(lock.REST.Operations))
	for _, operation := range lock.REST.Operations {
		if _, exists := locked[operation.ID]; exists {
			return fmt.Errorf("duplicate source lock ID %q", operation.ID)
		}
		locked[operation.ID] = operation
		lockedMethodPaths[jiraMethodPathKey(operation.Method, operation.Path)] = struct{}{}
	}
	if lock.Counts.Total != 617 || len(locked) != 617 {
		return fmt.Errorf("source lock denominator=%d unique=%d, want 617", lock.Counts.Total, len(locked))
	}
	if len(matrix.SourceOperations) != len(locked) {
		return fmt.Errorf("source rows matrix=%d lock=%d, want 617", len(matrix.SourceOperations), len(locked))
	}
	if err := validateJiraCrosswalkBoundary(matrix.SourceBoundaryReconciliation, crosswalk, lockedMethodPaths); err != nil {
		return err
	}
	if err := validateJiraLegacyETLReconciliation(matrix.LegacyETLReconciliation, locked, streams); err != nil {
		return err
	}

	counts := make(map[string]map[string]int, len(jiraSourceLaneNames))
	seen := make(map[string]struct{}, len(matrix.SourceOperations))
	for _, row := range matrix.SourceOperations {
		if _, exists := seen[row.SourceID]; exists {
			return fmt.Errorf("duplicate matrix source ID %q", row.SourceID)
		}
		seen[row.SourceID] = struct{}{}
		operation, ok := locked[row.SourceID]
		if !ok {
			return fmt.Errorf("matrix source ID %q is absent from the source lock", row.SourceID)
		}
		if err := validateJiraSourceFacts(row.SourceFacts, operation); err != nil {
			return fmt.Errorf("source facts %q: %w", row.SourceID, err)
		}
		for _, lane := range jiraSourceLaneNames {
			cell, ok := row.Lanes[lane]
			if !ok {
				return fmt.Errorf("missing lane cell: %s %s", row.SourceID, lane)
			}
			if err := validateJiraLaneCell(row.SourceID, lane, cell, operation); err != nil {
				return err
			}
			if counts[lane] == nil {
				counts[lane] = make(map[string]int)
			}
			counts[lane][cell.Disposition]++
		}
		if len(row.Lanes) != len(jiraSourceLaneNames) {
			return fmt.Errorf("lane cell count %s=%d, want %d", row.SourceID, len(row.Lanes), len(jiraSourceLaneNames))
		}
	}
	if len(seen) != len(locked) {
		return fmt.Errorf("matrix does not retain every locked source ID: matrix=%d lock=%d", len(seen), len(locked))
	}
	if matrix.Summary.SourceRows != len(locked) || matrix.Summary.SourceRowsWithAllLanes != len(locked) {
		return fmt.Errorf("summary source rows=%d rows_with_all_lanes=%d, want 617", matrix.Summary.SourceRows, matrix.Summary.SourceRowsWithAllLanes)
	}
	for _, lane := range jiraSourceLaneNames {
		if !equalJiraLaneCounts(counts[lane], jiraExpectedLaneCounts[lane]) {
			return fmt.Errorf("expected %s counts=%v, computed=%v", lane, jiraExpectedLaneCounts[lane], counts[lane])
		}
		if !equalJiraLaneCounts(matrix.Summary.LaneCounts[lane], counts[lane]) {
			return fmt.Errorf("summary %s counts=%v, computed=%v", lane, matrix.Summary.LaneCounts[lane], counts[lane])
		}
	}
	return nil
}

func validateJiraMatrixLockBinding(binding jiraSourceLaneMatrixSourceLock, lock jiraSourceLaneLock) error {
	if binding.Path != jiraSourceLockPath || binding.SchemaVersion != lock.SchemaVersion || binding.Connector != lock.Connector {
		return fmt.Errorf("source lock binding path=%q schema=%d connector=%q", binding.Path, binding.SchemaVersion, binding.Connector)
	}
	if binding.SourceDocument.SourceURL != lock.REST.SourceURL || binding.SourceDocument.SHA256 != lock.REST.SHA256 || binding.SourceDocument.Bytes != lock.REST.Bytes || binding.SourceDocument.OperationCount != lock.Counts.Total {
		return fmt.Errorf("source lock document binding drift")
	}
	return nil
}

func validateJiraCrosswalkBoundary(boundary jiraSourceBoundaryReconciliation, crosswalk jiraSourceLaneCrosswalk, lockedMethodPaths map[string]struct{}) error {
	if crosswalk.Connector != "jira" || crosswalk.SourceLock != jiraSourceLockPath || len(crosswalk.SourceOperations) != 617 {
		return fmt.Errorf("crosswalk identity or denominator drift")
	}
	if boundary.Identity != "method + path" || boundary.SourceLockRows != 617 || boundary.CrosswalkRows != 617 || boundary.CrosswalkOnlyRows != 0 || boundary.LockOnlyRows != 0 {
		return fmt.Errorf("crosswalk boundary counts identity=%q lock=%d crosswalk=%d crosswalk_only=%d lock_only=%d", boundary.Identity, boundary.SourceLockRows, boundary.CrosswalkRows, boundary.CrosswalkOnlyRows, boundary.LockOnlyRows)
	}
	if crosswalk.Accounting.SourceOperations != 617 || crosswalk.Accounting.SourceUniqueMethodPath != 617 || crosswalk.Accounting.APISurfaceEndpoints != 617 || crosswalk.Accounting.APISurfaceUniqueMethodPath != 617 || crosswalk.Accounting.ExactSourceToSurface != 617 || crosswalk.Accounting.SourceOnly != 0 || crosswalk.Accounting.SurfaceOnly != 0 {
		return fmt.Errorf("crosswalk accounting drift")
	}
	crosswalkMethodPaths := make(map[string]jiraCrosswalkSourceOperation, len(crosswalk.SourceOperations))
	for _, operation := range crosswalk.SourceOperations {
		key := jiraMethodPathKey(operation.Method, operation.Path)
		if _, exists := crosswalkMethodPaths[key]; exists {
			return fmt.Errorf("duplicate crosswalk method/path identity %q", key)
		}
		crosswalkMethodPaths[key] = operation
	}
	for key := range lockedMethodPaths {
		if _, present := crosswalkMethodPaths[key]; !present {
			return fmt.Errorf("lock-minus-crosswalk identity %q", key)
		}
	}
	for key := range crosswalkMethodPaths {
		if _, present := lockedMethodPaths[key]; !present {
			return fmt.Errorf("crosswalk-minus-lock identity %q", key)
		}
	}
	return nil
}

func validateJiraLegacyETLReconciliation(reconciliation jiraLegacyETLReconciliation, locked map[string]jiraLockedSourceOperation, streams jiraStreamsDefinition) error {
	if reconciliation.SourcePagingCandidateCriterion != "source-read operation with retained continuation evidence" || reconciliation.SourcePagingCandidates != 103 || reconciliation.RemainingPagingCandidates != 100 {
		return fmt.Errorf("legacy ETL paging reconciliation counts or criterion drift")
	}
	if len(reconciliation.LegacyStreamBacklinks) != len(jiraLegacyETLStreams) {
		return fmt.Errorf("legacy stream backlink count=%d, want %d", len(reconciliation.LegacyStreamBacklinks), len(jiraLegacyETLStreams))
	}
	streamPaths := make(map[string]string, len(streams.Streams))
	for _, stream := range streams.Streams {
		streamPaths[stream.Name] = stream.Path
	}
	seen := make(map[string]struct{}, len(reconciliation.LegacyStreamBacklinks))
	for _, backlink := range reconciliation.LegacyStreamBacklinks {
		expected, ok := jiraLegacyETLStreams[backlink.SourceID]
		if !ok {
			return fmt.Errorf("unexpected legacy stream backlink source ID %q", backlink.SourceID)
		}
		if _, duplicate := seen[backlink.SourceID]; duplicate {
			return fmt.Errorf("duplicate legacy stream backlink source ID %q", backlink.SourceID)
		}
		seen[backlink.SourceID] = struct{}{}
		operation := locked[backlink.SourceID]
		if operation.Path != expected.Path || !jiraIsETLCandidate(operation) {
			return fmt.Errorf("legacy stream backlink source facts drift for %q", backlink.SourceID)
		}
		if backlink.Stream != expected.Stream || backlink.Path != expected.Path || streamPaths[backlink.Stream] != backlink.Path {
			return fmt.Errorf("legacy stream backlink drift for %q", backlink.SourceID)
		}
	}
	if len(seen) != len(jiraLegacyETLStreams) {
		return fmt.Errorf("legacy stream backlink coverage=%d, want %d", len(seen), len(jiraLegacyETLStreams))
	}
	actualCandidates := 0
	for _, operation := range locked {
		if jiraIsETLCandidate(operation) {
			actualCandidates++
		}
	}
	if actualCandidates != reconciliation.SourcePagingCandidates || actualCandidates-len(seen) != reconciliation.RemainingPagingCandidates {
		return fmt.Errorf("legacy ETL source candidate accounting=%d legacy=%d remaining=%d", actualCandidates, len(seen), actualCandidates-len(seen))
	}
	return nil
}

func validateJiraSourceFacts(facts jiraSourceLaneMatrixSourceFacts, operation jiraLockedSourceOperation) error {
	if facts.Protocol != operation.Protocol || facts.Method != operation.Method || facts.Path != operation.Path || facts.OperationID != operation.OperationID || facts.Citation.SourceLocation != operation.SourceLocation || !equalJiraOptionalBool(facts.Deprecated, operation.Deprecated) {
		return fmt.Errorf("identity or citation drift")
	}
	if !slices.Equal(facts.ScopeAndFanout.PathVariables, jiraPathVariables(operation)) || !slices.Equal(facts.ScopeAndFanout.QueryParameters, jiraQueryParameters(operation)) || facts.ScopeAndFanout.Fanout.State != "not_declared" {
		return fmt.Errorf("scope or fanout facts drift")
	}
	requestMedia, responseMedia := jiraMediaTypes(operation)
	if !slices.Equal(facts.Media.RequestMediaTypes, requestMedia) || !slices.Equal(facts.Media.SuccessResponseMediaTypes, responseMedia) || !slices.Equal(facts.Media.BinarySignals, jiraBinarySignals(operation)) {
		return fmt.Errorf("media facts drift")
	}
	wantPaginationState := jiraPaginationState(operation)
	if facts.Pagination.State != wantPaginationState || !slices.Equal(facts.Pagination.PagingQueryParameters, jiraPagingQueryParameters(operation)) {
		return fmt.Errorf("pagination facts drift")
	}
	if facts.EventCursor.State != jiraEventCursorState(operation) || facts.OperationSemantics.State != jiraOperationSemantics(operation) {
		return fmt.Errorf("event/cursor or operation semantics facts drift")
	}
	if err := validateJiraCandidateEvidence(operation); err != nil {
		return err
	}
	return nil
}

func validateJiraCandidateEvidence(operation jiraLockedSourceOperation) error {
	if evidence, ok := jiraDocumentedReadSemantics[operation.ID]; ok {
		if operation.Method != "POST" || !strings.Contains(jiraSourceText(operation), evidence.Contains) {
			return fmt.Errorf("documented POST read semantics evidence drift")
		}
	}
	if evidence, ok := jiraBodyPaginationEvidence[operation.ID]; ok {
		sourceText := jiraSourceText(operation)
		if !jiraIsDirectReadCandidate(operation) || !jiraHasBodyContinuationEvidence(sourceText, evidence) {
			return fmt.Errorf("documented body pagination evidence drift")
		}
	}
	if evidence, ok := jiraBinaryDownloadEvidence[operation.ID]; ok && !slices.Equal(jiraSortedUnique(evidence.SuccessMedia), jiraSuccessMediaIntersection(operation, evidence.SuccessMedia)) {
		return fmt.Errorf("binary download media evidence drift")
	}
	if evidence, ok := jiraBinaryUploadEvidence[operation.ID]; ok {
		if !slices.Equal(jiraSortedUnique(evidence.RequestMedia), jiraRequestMediaIntersection(operation, evidence.RequestMedia)) {
			return fmt.Errorf("binary upload media evidence drift")
		}
		if evidence.SourceTextContains != "" && !strings.Contains(jiraSourceText(operation), evidence.SourceTextContains) {
			return fmt.Errorf("binary upload source text %q is absent", evidence.SourceTextContains)
		}
	}
	if operation.ID == "jira.rest.registerDynamicWebhooks" {
		sourceText := jiraSourceText(operation)
		if !strings.Contains(sourceText, "\"url\"") || !strings.Contains(sourceText, "\"events\"") {
			return fmt.Errorf("webhook registration source evidence is absent")
		}
	}
	return nil
}

func validateJiraLaneCell(sourceID, lane string, cell jiraSourceLaneMatrixCell, operation jiraLockedSourceOperation) error {
	wantApplicable, wantDisposition := jiraExpectedLane(operation, lane)
	if cell.Applicability != wantApplicable || cell.Disposition != wantDisposition {
		return fmt.Errorf("lane %s %s applicability=%q disposition=%q, want applicability=%q disposition=%q", lane, sourceID, cell.Applicability, cell.Disposition, wantApplicable, wantDisposition)
	}
	if strings.TrimSpace(cell.Reason) == "" {
		return fmt.Errorf("lane %s %s lacks a reason", lane, sourceID)
	}
	if cell.Applicability == "not_applicable" {
		if cell.Disposition != "not_applicable" || len(cell.Mapping) != 0 {
			return fmt.Errorf("not-applicable lane promoted or mapped: %s %s", lane, sourceID)
		}
		return nil
	}
	if cell.Applicability != "applicable" || (cell.Disposition != "mapped_unproven" && cell.Disposition != "missing_foundation") {
		return fmt.Errorf("invalid applicable lane state: %s %s", lane, sourceID)
	}
	if len(cell.Mapping) == 0 || !json.Valid(cell.Mapping) {
		return fmt.Errorf("applicable lane lacks valid mapping evidence: %s %s", lane, sourceID)
	}
	if cell.Disposition == "implemented" {
		return fmt.Errorf("Track A must not claim implemented lane: %s %s", lane, sourceID)
	}
	if cell.Disposition == "missing_foundation" && lane != "sync_transport" {
		return fmt.Errorf("non-sync missing-foundation lane: %s %s", lane, sourceID)
	}
	if err := validateJiraLaneMapping(lane, cell.Mapping, operation); err != nil {
		return fmt.Errorf("mapping evidence %s %s: %w", lane, sourceID, err)
	}
	return nil
}

func validateJiraLaneMapping(lane string, raw json.RawMessage, operation jiraLockedSourceOperation) error {
	switch lane {
	case "direct_read":
		var mapping struct {
			SourceFact struct {
				Method         string `json:"method"`
				Classification string `json:"classification"`
			} `json:"source_fact"`
			RuntimeClaim string `json:"runtime_claim"`
		}
		if err := json.Unmarshal(raw, &mapping); err != nil {
			return err
		}
		if !jiraIsDirectReadCandidate(operation) || mapping.SourceFact.Method != operation.Method || mapping.SourceFact.Classification != jiraOperationSemantics(operation) || strings.TrimSpace(mapping.RuntimeClaim) == "" {
			return fmt.Errorf("direct-read source mapping drift")
		}
	case "direct_write":
		return validateJiraMutationLaneMapping(raw, operation, false)
	case "binary_download":
		return validateJiraBinaryLaneMapping(raw, jiraBinaryDownloadEvidence[operation.ID].Signals)
	case "binary_upload":
		return validateJiraBinaryLaneMapping(raw, jiraBinaryUploadEvidence[operation.ID].Signals)
	case "etl":
		return validateJiraETLLaneMapping(raw, operation)
	case "reverse_etl":
		return validateJiraMutationLaneMapping(raw, operation, true)
	case "sync_transport":
		var mapping struct {
			FoundationID string `json:"foundation_id"`
			AtlasLookup  struct {
				ConsultedID    string   `json:"consulted_id"`
				Classification string   `json:"classification"`
				OwnerSymbols   []string `json:"owner_symbols"`
				Insufficiency  string   `json:"insufficiency"`
			} `json:"atlas_lookup"`
			SourceEventEvidence struct {
				URLField    string `json:"url_field"`
				EventsField string `json:"events_field"`
			} `json:"source_event_evidence"`
			RuntimeClaim string `json:"runtime_claim"`
		}
		if err := json.Unmarshal(raw, &mapping); err != nil {
			return err
		}
		if mapping.FoundationID != "cli-webhook-event-surface-foundation-r1" || mapping.AtlasLookup.ConsultedID != "transport.sync-contract.v1" || mapping.AtlasLookup.Classification != "actual_gap" || !slices.Equal(mapping.AtlasLookup.OwnerSymbols, []string{"internal/connectors/sync_transport.go#SyncTransportDescriptor", "internal/synctransport/orchestrator.go#(*Orchestrator).Run"}) || strings.TrimSpace(mapping.AtlasLookup.Insufficiency) == "" || mapping.SourceEventEvidence.URLField != "url" || mapping.SourceEventEvidence.EventsField != "webhooks[].events" || strings.TrimSpace(mapping.RuntimeClaim) == "" {
			return fmt.Errorf("missing-foundation mapping drift")
		}
	default:
		return fmt.Errorf("unknown lane %q", lane)
	}
	return nil
}

func validateJiraMutationLaneMapping(raw json.RawMessage, operation jiraLockedSourceOperation, reverseETL bool) error {
	var mapping struct {
		SourceFact struct {
			Method         string `json:"method"`
			Classification string `json:"classification"`
		} `json:"source_fact"`
		RequiredFlow string `json:"required_flow"`
		RuntimeClaim string `json:"runtime_claim"`
	}
	if err := json.Unmarshal(raw, &mapping); err != nil {
		return err
	}
	if !jiraIsMutationCandidate(operation) || mapping.SourceFact.Method != operation.Method || mapping.SourceFact.Classification != jiraOperationSemantics(operation) || strings.TrimSpace(mapping.RuntimeClaim) == "" {
		return fmt.Errorf("mutation source mapping drift")
	}
	if reverseETL && strings.TrimSpace(mapping.RequiredFlow) == "" {
		return fmt.Errorf("reverse-ETL required flow is absent")
	}
	return nil
}

func validateJiraBinaryLaneMapping(raw json.RawMessage, wantSignals []string) error {
	var mapping struct {
		SourceBinarySignals []string `json:"source_binary_signals"`
		RuntimeClaim        string   `json:"runtime_claim"`
	}
	if err := json.Unmarshal(raw, &mapping); err != nil {
		return err
	}
	if !slices.Equal(mapping.SourceBinarySignals, wantSignals) || strings.TrimSpace(mapping.RuntimeClaim) == "" {
		return fmt.Errorf("binary source mapping drift")
	}
	return nil
}

func validateJiraETLLaneMapping(raw json.RawMessage, operation jiraLockedSourceOperation) error {
	var mapping struct {
		SourceFact struct {
			Method         string `json:"method"`
			Criterion      string `json:"criterion"`
			QueryParameter string `json:"query_parameter"`
			Classification string `json:"classification"`
		} `json:"source_fact"`
		DefinitionBacklink struct {
			Kind       string `json:"kind"`
			Path       string `json:"path"`
			Stream     string `json:"stream"`
			StreamPath string `json:"stream_path"`
			SourceID   string `json:"source_id"`
		} `json:"definition_backlink"`
		RuntimeClaim string `json:"runtime_claim"`
	}
	if err := json.Unmarshal(raw, &mapping); err != nil {
		return err
	}
	evidence, ok := jiraContinuationEvidence(operation)
	if !ok || mapping.SourceFact.Method != operation.Method || strings.TrimSpace(mapping.RuntimeClaim) == "" {
		return fmt.Errorf("ETL source mapping drift")
	}
	if mapping.SourceFact.Criterion == "retained query parameter maxResults" {
		if mapping.SourceFact.QueryParameter != "maxResults" || !slices.Contains(jiraQueryParameters(operation), "maxResults") {
			return fmt.Errorf("legacy maxResults ETL source mapping drift")
		}
	} else if mapping.SourceFact.Criterion != "documented source continuation" || mapping.SourceFact.QueryParameter != "" || mapping.SourceFact.Classification != evidence.Kind {
		return fmt.Errorf("documented continuation ETL source mapping drift")
	}
	if legacy, ok := jiraLegacyETLStreams[operation.ID]; ok {
		if mapping.DefinitionBacklink.Kind != "legacy_stream_backlink" || mapping.DefinitionBacklink.Path != jiraStreamsPath || mapping.DefinitionBacklink.Stream != legacy.Stream || mapping.DefinitionBacklink.StreamPath != legacy.Path || mapping.DefinitionBacklink.SourceID != operation.ID {
			return fmt.Errorf("legacy stream backlink drift")
		}
		return nil
	}
	if mapping.DefinitionBacklink.Kind != "future_stream_projection_required" || mapping.DefinitionBacklink.Path != jiraStreamsPath || mapping.DefinitionBacklink.SourceID != operation.ID || mapping.DefinitionBacklink.Stream != "" || mapping.DefinitionBacklink.StreamPath != "" {
		return fmt.Errorf("future stream projection backlink drift")
	}
	return nil
}

func jiraExpectedLane(operation jiraLockedSourceOperation, lane string) (string, string) {
	applicable := false
	switch lane {
	case "direct_read":
		applicable = jiraIsDirectReadCandidate(operation)
	case "direct_write", "reverse_etl":
		applicable = jiraIsMutationCandidate(operation)
	case "binary_download":
		_, applicable = jiraBinaryDownloadEvidence[operation.ID]
	case "binary_upload":
		_, applicable = jiraBinaryUploadEvidence[operation.ID]
	case "etl":
		applicable = jiraIsETLCandidate(operation)
	case "sync_transport":
		applicable = operation.ID == "jira.rest.registerDynamicWebhooks"
	}
	if !applicable {
		return "not_applicable", "not_applicable"
	}
	if lane == "sync_transport" {
		return "applicable", "missing_foundation"
	}
	return "applicable", "mapped_unproven"
}

func jiraIsETLCandidate(operation jiraLockedSourceOperation) bool {
	_, ok := jiraContinuationEvidence(operation)
	return ok
}

type jiraContinuationFact struct {
	Kind string
}

func jiraIsDirectReadCandidate(operation jiraLockedSourceOperation) bool {
	return jiraOperationSemantics(operation) != "mutation_verb_candidate"
}

func jiraIsMutationCandidate(operation jiraLockedSourceOperation) bool {
	return !jiraIsDirectReadCandidate(operation)
}

func jiraContinuationEvidence(operation jiraLockedSourceOperation) (jiraContinuationFact, bool) {
	if !jiraIsDirectReadCandidate(operation) {
		return jiraContinuationFact{}, false
	}
	if evidence, ok := jiraBodyPaginationEvidence[operation.ID]; ok {
		sourceText := jiraSourceText(operation)
		if jiraHasBodyContinuationEvidence(sourceText, evidence) {
			return jiraContinuationFact{Kind: evidence.ContinuationKind}, true
		}
	}
	for _, parameter := range operation.SourceOperation.Parameters {
		if parameter.In != "query" {
			continue
		}
		description := strings.ToLower(parameter.Description)
		switch {
		case strings.Contains(description, "cursor"):
			return jiraContinuationFact{Kind: "query_cursor"}, true
		case strings.Contains(description, "page offset"),
			strings.Contains(description, "first item to return in a page"),
			strings.Contains(description, "starting index of the returned"),
			strings.Contains(description, "index of the first item to return"):
			return jiraContinuationFact{Kind: "query_offset"}, true
		}
	}
	sourceText := strings.ToLower(jiraSourceText(operation))
	switch {
	case strings.Contains(sourceText, "cursor-based pagination") && strings.Contains(sourceText, "next"):
		return jiraContinuationFact{Kind: "response_link_cursor"}, true
	case strings.Contains(sourceText, "nextpage") && strings.Contains(sourceText, "next page"):
		return jiraContinuationFact{Kind: "response_next_page"}, true
	default:
		return jiraContinuationFact{}, false
	}
}

func jiraHasBodyContinuationEvidence(sourceText string, evidence jiraSourceTextEvidence) bool {
	if !strings.Contains(sourceText, evidence.Contains) {
		return false
	}
	switch evidence.ContinuationKind {
	case "body_offset":
		return strings.Contains(sourceText, "\"startAt\"")
	case "body_cursor":
		return strings.Contains(strings.ToLower(sourceText), "scrolling view")
	default:
		return false
	}
}

func jiraPaginationState(operation jiraLockedSourceOperation) string {
	if slices.Contains(jiraQueryParameters(operation), "maxResults") {
		if _, ok := jiraContinuationEvidence(operation); !ok {
			return "max_results_limit_only"
		}
		return "max_results_query_candidate"
	}
	if evidence, ok := jiraContinuationEvidence(operation); ok {
		return "documented_" + evidence.Kind + "_candidate"
	}
	return "not_max_results_candidate"
}

func jiraPathVariables(operation jiraLockedSourceOperation) []string {
	values := make([]string, 0, len(operation.SourceOperation.Parameters))
	for _, parameter := range operation.SourceOperation.Parameters {
		if parameter.In == "path" && parameter.Name != "" {
			values = append(values, parameter.Name)
		}
	}
	return jiraSortedUnique(values)
}

func jiraQueryParameters(operation jiraLockedSourceOperation) []string {
	values := make([]string, 0, len(operation.SourceOperation.Parameters))
	for _, parameter := range operation.SourceOperation.Parameters {
		if parameter.In == "query" && parameter.Name != "" {
			values = append(values, parameter.Name)
		}
	}
	return jiraSortedUnique(values)
}

func jiraPagingQueryParameters(operation jiraLockedSourceOperation) []string {
	known := map[string]struct{}{
		"after": {}, "endingBefore": {}, "failedAfter": {}, "maxResult": {}, "maxResults": {}, "nextPageToken": {}, "since": {}, "startAt": {}, "startingAfter": {},
	}
	values := make([]string, 0)
	for _, parameter := range jiraQueryParameters(operation) {
		if _, ok := known[parameter]; ok {
			values = append(values, parameter)
		}
	}
	return jiraSortedUnique(values)
}

func jiraMediaTypes(operation jiraLockedSourceOperation) ([]string, []string) {
	request := make([]string, 0)
	if operation.SourceOperation.RequestBody != nil {
		for mediaType := range operation.SourceOperation.RequestBody.Content {
			request = append(request, mediaType)
		}
	}
	response := make([]string, 0)
	for status, result := range operation.SourceOperation.Responses {
		if !strings.HasPrefix(status, "2") {
			continue
		}
		for mediaType := range result.Content {
			response = append(response, mediaType)
		}
	}
	return jiraSortedUnique(request), jiraSortedUnique(response)
}

func jiraBinarySignals(operation jiraLockedSourceOperation) []string {
	if evidence, ok := jiraBinaryDownloadEvidence[operation.ID]; ok {
		return append([]string(nil), evidence.Signals...)
	}
	if evidence, ok := jiraBinaryUploadEvidence[operation.ID]; ok {
		return append([]string(nil), evidence.Signals...)
	}
	return []string{}
}

func jiraSuccessMediaIntersection(operation jiraLockedSourceOperation, want []string) []string {
	_, actual := jiraMediaTypes(operation)
	return jiraMediaIntersection(actual, want)
}

func jiraRequestMediaIntersection(operation jiraLockedSourceOperation, want []string) []string {
	actual, _ := jiraMediaTypes(operation)
	return jiraMediaIntersection(actual, want)
}

func jiraMediaIntersection(actual, want []string) []string {
	available := make(map[string]struct{}, len(actual))
	for _, value := range actual {
		available[value] = struct{}{}
	}
	intersection := make([]string, 0, len(want))
	for _, value := range want {
		if _, ok := available[value]; ok {
			intersection = append(intersection, value)
		}
	}
	return jiraSortedUnique(intersection)
}

func jiraEventCursorState(operation jiraLockedSourceOperation) string {
	if state, ok := jiraEventCursorStates[operation.ID]; ok {
		return state
	}
	return "not_declared"
}

func jiraOperationSemantics(operation jiraLockedSourceOperation) string {
	if _, ok := jiraDocumentedReadSemantics[operation.ID]; ok {
		return "documented_read_semantics"
	}
	if operation.Method == "GET" {
		return "read_verb_candidate"
	}
	return "mutation_verb_candidate"
}

func jiraSourceText(operation jiraLockedSourceOperation) string {
	raw, err := json.Marshal(operation.SourceOperation)
	if err != nil {
		return operation.SourceOperation.Summary + "\n" + operation.SourceOperation.Description
	}
	return string(raw)
}

func jiraMethodPathKey(method, path string) string {
	return method + " " + path
}

func jiraSortedUnique(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	copyValues := append([]string(nil), values...)
	sort.Strings(copyValues)
	return slices.Compact(copyValues)
}

func equalJiraOptionalBool(left, right *bool) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func equalJiraLaneCounts(got, want map[string]int) bool {
	if len(got) != len(want) {
		return false
	}
	for disposition, count := range want {
		if got[disposition] != count {
			return false
		}
	}
	return true
}

func cloneJiraSourceLaneMatrix(t *testing.T, matrix jiraSourceLaneMatrix) jiraSourceLaneMatrix {
	t.Helper()
	return loadJiraJSONFromBytes[jiraSourceLaneMatrix](t, mustJiraJSON(t, matrix))
}

func loadJiraJSONFromBytes[T any](t *testing.T, raw []byte) T {
	t.Helper()
	var decoded T
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("decode clone: %v", err)
	}
	return decoded
}

func mustJiraJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal JSON: %v", err)
	}
	return raw
}

func jiraMatrixRow(t *testing.T, matrix *jiraSourceLaneMatrix, sourceID string) *jiraSourceLaneMatrixRow {
	t.Helper()
	for index := range matrix.SourceOperations {
		if matrix.SourceOperations[index].SourceID == sourceID {
			return &matrix.SourceOperations[index]
		}
	}
	t.Fatalf("matrix row %q not found", sourceID)
	return nil
}
