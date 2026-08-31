package jira

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"slices"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
)

const jiraSemanticPostReadCorrectionsPath = "sources/jira-semantic-post-read-corrections.json"

type jiraSemanticPostReadCorrections struct {
	SchemaVersion int    `json:"schema_version"`
	Connector     string `json:"connector"`
	SourceLock    struct {
		Path           string `json:"path"`
		SchemaVersion  int    `json:"schema_version"`
		SourceURL      string `json:"source_url"`
		SHA256         string `json:"sha256"`
		OperationCount int    `json:"operation_count"`
	} `json:"source_lock"`
	Corrections []jiraSemanticPostReadCorrection `json:"corrections"`
}

type jiraSemanticPostReadCorrection struct {
	SourceID                string `json:"source_id"`
	SourceLocation          string `json:"source_location"`
	Method                  string `json:"method"`
	Path                    string `json:"path"`
	SourceSemanticsContains string `json:"source_semantics_contains"`
	RequestSchemaPointer    string `json:"request_schema_pointer"`
	ConnectorOperationID    string `json:"connector_operation_id"`
	CLIPath                 string `json:"cli_path"`
	FormerWriteAction       string `json:"former_write_action"`
}

func TestJiraSemanticPOSTReadCorrectionsRemainTypedDirectReads(t *testing.T) {
	corrections := loadJiraSemanticPostReadCorrections(t)
	lock := loadJiraSourceLaneLock(t)
	matrix := loadJiraSourceLaneMatrix(t)
	bundle, err := engine.Load(os.DirFS(".."), "jira")
	if err != nil {
		t.Fatalf("load Jira bundle: %v", err)
	}
	connector := engine.New(bundle, nil)

	if corrections.SchemaVersion != 1 || corrections.Connector != "jira" || corrections.SourceLock.Path != jiraSourceLockPath || corrections.SourceLock.SchemaVersion != lock.SchemaVersion || corrections.SourceLock.SourceURL != lock.REST.SourceURL || corrections.SourceLock.SHA256 != lock.REST.SHA256 || corrections.SourceLock.OperationCount != lock.Counts.Total {
		t.Fatalf("correction ledger source-lock binding = %+v, want frozen Jira source lock %+v", corrections.SourceLock, lock.REST)
	}
	if len(corrections.Corrections) == 0 {
		t.Fatal("semantic POST-read correction ledger is empty")
	}

	operations := make(map[string]engine.OperationSpec, len(bundle.Operations))
	for _, operation := range bundle.Operations {
		operations[operation.ID] = operation
	}
	writes := make(map[string]struct{}, len(bundle.Writes))
	for _, action := range bundle.Writes {
		writes[action.Name] = struct{}{}
	}
	commands := make(map[string]engine.CLICommand, len(bundle.CLISurface.Commands))
	for _, command := range bundle.CLISurface.Commands {
		commands[command.Path] = command
	}

	locked := make(map[string]jiraLockedSourceOperation, len(lock.REST.Operations))
	for _, operation := range lock.REST.Operations {
		locked[operation.ID] = operation
	}
	seenSourceIDs := make(map[string]struct{}, len(corrections.Corrections))
	seenOperationIDs := make(map[string]struct{}, len(corrections.Corrections))
	seenCLIPaths := make(map[string]struct{}, len(corrections.Corrections))
	for _, correction := range corrections.Corrections {
		t.Run(correction.SourceID, func(t *testing.T) {
			if correction.SourceID == "" || correction.ConnectorOperationID == "" || correction.CLIPath == "" || correction.FormerWriteAction == "" {
				t.Fatalf("incomplete correction row: %+v", correction)
			}
			if _, duplicate := seenSourceIDs[correction.SourceID]; duplicate {
				t.Fatalf("duplicate correction source ID %q", correction.SourceID)
			}
			seenSourceIDs[correction.SourceID] = struct{}{}
			if _, duplicate := seenOperationIDs[correction.ConnectorOperationID]; duplicate {
				t.Fatalf("duplicate correction operation ID %q", correction.ConnectorOperationID)
			}
			seenOperationIDs[correction.ConnectorOperationID] = struct{}{}
			if _, duplicate := seenCLIPaths[correction.CLIPath]; duplicate {
				t.Fatalf("duplicate correction CLI path %q", correction.CLIPath)
			}
			seenCLIPaths[correction.CLIPath] = struct{}{}

			source, ok := locked[correction.SourceID]
			if !ok {
				t.Fatalf("source operation %q absent from frozen lock", correction.SourceID)
			}
			if source.Method != correction.Method || correction.Method != http.MethodPost || source.Path != correction.Path || source.SourceLocation != correction.SourceLocation {
				t.Fatalf("source identity = %s %s at %s, correction = %s %s at %s", source.Method, source.Path, source.SourceLocation, correction.Method, correction.Path, correction.SourceLocation)
			}
			evidence, ok := jiraDocumentedReadSemantics[correction.SourceID]
			if !ok || evidence.Contains != correction.SourceSemanticsContains || !strings.Contains(jiraSourceText(source), correction.SourceSemanticsContains) {
				t.Fatalf("source read semantics drift for %q: evidence=%+v source=%q", correction.SourceID, evidence, jiraSourceText(source))
			}
			if correction.RequestSchemaPointer != "source_operation.requestBody.content.application/json.schema" || source.SourceOperation.RequestBody == nil || len(source.SourceOperation.RequestBody.Content["application/json"]) == 0 {
				t.Fatalf("source request-schema evidence is absent for %q", correction.SourceID)
			}

			row := jiraMatrixRow(t, &matrix, correction.SourceID)
			if row.Lanes["direct_read"].Applicability != "applicable" || row.Lanes["direct_write"].Applicability != "not_applicable" || row.Lanes["reverse_etl"].Applicability != "not_applicable" {
				t.Fatalf("source lane semantics for %q = direct_read:%+v direct_write:%+v reverse_etl:%+v, want read-only source classification", correction.SourceID, row.Lanes["direct_read"], row.Lanes["direct_write"], row.Lanes["reverse_etl"])
			}
			if _, stale := writes[correction.FormerWriteAction]; stale {
				t.Fatalf("documented read %q remains a write action %q", correction.SourceID, correction.FormerWriteAction)
			}

			operation, ok := operations[correction.ConnectorOperationID]
			if !ok || operation.Kind != "rest_read" || operation.REST == nil || operation.REST.Method != http.MethodPost || operation.REST.Path != correction.Path || operation.REST.ContentType != "application/json" || operation.REST.MaxBytes <= 0 || len(operation.REST.BodySchema) == 0 {
				t.Fatalf("typed direct-read operation %q = %+v, want source-backed POST JSON read", correction.ConnectorOperationID, operation)
			}
			command, ok := commands[correction.CLIPath]
			if !ok || command.Intent != "direct_read" || command.Availability != "implemented" || command.Operation != correction.ConnectorOperationID || command.Write != "" || command.SourceOperation != "" || command.OutputPolicy != "json_redacted" || len(command.APISurface) != 1 || command.APISurface[0].Method != http.MethodPost || command.APISurface[0].Path != correction.Path {
				t.Fatalf("CLI direct-read correction %q = %+v", correction.CLIPath, command)
			}
			for _, flag := range command.Flags {
				if !strings.HasPrefix(flag.MapsTo, "path.") && !strings.HasPrefix(flag.MapsTo, "body.") {
					t.Fatalf("CLI direct-read correction %q leaves non-read flag binding %q", correction.CLIPath, flag.MapsTo)
				}
			}
			if !jiraAPISurfaceDirectReadCovers(bundle, correction.CLIPath, correction.Method, correction.Path) {
				t.Fatalf("API surface does not bind direct read %q to %s %s", correction.CLIPath, correction.Method, correction.Path)
			}
			if err := commandrunner.Preflight(connector, strings.Fields(correction.CLIPath)); err != nil {
				t.Fatalf("typed direct-read runtime preflight %q: %v", correction.CLIPath, err)
			}
		})
	}
}

func TestJiraSemanticPOSTReadCorrectionExecutesClosedTypedRequest(t *testing.T) {
	corrections := loadJiraSemanticPostReadCorrections(t)
	var correction *jiraSemanticPostReadCorrection
	for index := range corrections.Corrections {
		candidate := &corrections.Corrections[index]
		if candidate.CLIPath == "field get-custom-field-contexts-for-projects-and-issue-types" {
			correction = candidate
			break
		}
	}
	if correction == nil {
		t.Fatal("field context mapping correction is absent")
	}

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/rest/api/3/field/field-123/context/mapping" {
			t.Fatalf("provider request = %s %s, want POST corrected field-context path", request.Method, request.URL.Path)
		}
		var body struct {
			FieldID  string `json:"fieldId"`
			Mappings []struct {
				IssueTypeID string `json:"issueTypeId"`
				ProjectID   string `json:"projectId"`
			} `json:"mappings"`
		}
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatalf("decode corrected POST body: %v", err)
		}
		if body.FieldID != "" || len(body.Mappings) != 1 || body.Mappings[0].IssueTypeID != "issue-type-123" || body.Mappings[0].ProjectID != "project-123" {
			t.Fatalf("corrected typed POST body = %+v, want path-only fieldId and declared mapping body", body)
		}
		response.Header().Set("Content-Type", "application/json")
		_, _ = response.Write([]byte(`{"values":[]}`))
	}))
	t.Cleanup(server.Close)

	bundle, err := engine.Load(os.DirFS(".."), "jira")
	if err != nil {
		t.Fatalf("load Jira bundle: %v", err)
	}
	result, err := commandrunner.Run(context.Background(), engine.New(bundle, nil), commandrunner.Request{
		Path: strings.Fields(correction.CLIPath),
		Flags: map[string][]string{
			"field-id":                 {"field-123"},
			"mappings-0-issue-type-id": {"issue-type-123"},
			"mappings-0-project-id":    {"project-123"},
		},
		Config: connectors.RuntimeConfig{
			Config:  map[string]string{"base_url": server.URL, "email": "fixture@example.test"},
			Secrets: map[string]string{"api_token": "fixture-token"},
		},
	}, nil)
	if err != nil {
		t.Fatalf("run corrected typed direct read: %v", err)
	}
	if result.DirectRead == nil || result.DirectRead.Operation != correction.ConnectorOperationID || result.DirectRead.Method != http.MethodPost || result.DirectRead.Status != http.StatusOK {
		t.Fatalf("corrected direct-read result = %+v", result)
	}
}

func TestJiraSemanticPOSTReadCorrectionRejectsEvidenceDrift(t *testing.T) {
	corrections := loadJiraSemanticPostReadCorrections(t)
	if len(corrections.Corrections) == 0 {
		t.Fatal("semantic POST-read correction ledger is empty")
	}
	broken := corrections
	broken.Corrections = append([]jiraSemanticPostReadCorrection(nil), corrections.Corrections...)
	broken.Corrections[0].Method = http.MethodGet
	if err := validateJiraSemanticPostReadCorrectionLedger(broken, loadJiraSourceLaneLock(t)); err == nil || !strings.Contains(err.Error(), "POST") {
		t.Fatalf("method-drift validation error = %v, want POST source identity rejection", err)
	}
	broken = corrections
	broken.Corrections = append([]jiraSemanticPostReadCorrection(nil), corrections.Corrections...)
	broken.Corrections[1].SourceID = broken.Corrections[0].SourceID
	if err := validateJiraSemanticPostReadCorrectionLedger(broken, loadJiraSourceLaneLock(t)); err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Fatalf("duplicate-source validation error = %v, want duplicate source rejection", err)
	}
}

func loadJiraSemanticPostReadCorrections(t *testing.T) jiraSemanticPostReadCorrections {
	t.Helper()
	raw, err := os.ReadFile(jiraSemanticPostReadCorrectionsPath)
	if err != nil {
		t.Fatalf("read semantic POST-read correction ledger: %v", err)
	}
	var corrections jiraSemanticPostReadCorrections
	if err := json.Unmarshal(raw, &corrections); err != nil {
		t.Fatalf("decode semantic POST-read correction ledger: %v", err)
	}
	if err := validateJiraSemanticPostReadCorrectionLedger(corrections, loadJiraSourceLaneLock(t)); err != nil {
		t.Fatalf("validate semantic POST-read correction ledger: %v", err)
	}
	return corrections
}

func validateJiraSemanticPostReadCorrectionLedger(corrections jiraSemanticPostReadCorrections, lock jiraSourceLaneLock) error {
	if corrections.SchemaVersion != 1 || corrections.Connector != "jira" || corrections.SourceLock.Path != jiraSourceLockPath || corrections.SourceLock.SchemaVersion != lock.SchemaVersion || corrections.SourceLock.SourceURL != lock.REST.SourceURL || corrections.SourceLock.SHA256 != lock.REST.SHA256 || corrections.SourceLock.OperationCount != lock.Counts.Total {
		return &jiraSemanticPostReadCorrectionError{message: "source-lock binding drift"}
	}
	locked := make(map[string]jiraLockedSourceOperation, len(lock.REST.Operations))
	for _, operation := range lock.REST.Operations {
		locked[operation.ID] = operation
	}
	seen := make(map[string]struct{}, len(corrections.Corrections))
	for _, correction := range corrections.Corrections {
		if _, duplicate := seen[correction.SourceID]; duplicate {
			return &jiraSemanticPostReadCorrectionError{message: "duplicate source ID " + correction.SourceID}
		}
		seen[correction.SourceID] = struct{}{}
		source, ok := locked[correction.SourceID]
		if !ok || correction.Method != http.MethodPost || source.Method != correction.Method || source.Path != correction.Path || source.SourceLocation != correction.SourceLocation {
			return &jiraSemanticPostReadCorrectionError{message: "POST source identity drift for " + correction.SourceID}
		}
		if correction.RequestSchemaPointer != "source_operation.requestBody.content.application/json.schema" || source.SourceOperation.RequestBody == nil || len(source.SourceOperation.RequestBody.Content["application/json"]) == 0 {
			return &jiraSemanticPostReadCorrectionError{message: "request-schema evidence drift for " + correction.SourceID}
		}
		evidence, ok := jiraDocumentedReadSemantics[correction.SourceID]
		if !ok || evidence.Contains != correction.SourceSemanticsContains || !strings.Contains(jiraSourceText(source), correction.SourceSemanticsContains) {
			return &jiraSemanticPostReadCorrectionError{message: "source semantic evidence drift for " + correction.SourceID}
		}
	}
	return nil
}

type jiraSemanticPostReadCorrectionError struct {
	message string
}

func (err *jiraSemanticPostReadCorrectionError) Error() string {
	return err.message
}

func jiraAPISurfaceDirectReadCovers(bundle engine.Bundle, cliPath, method, path string) bool {
	if bundle.Surface == nil {
		return false
	}
	for _, endpoint := range bundle.Surface.Endpoints {
		if endpoint.Method == method && endpoint.Path == path && endpoint.CoveredBy != nil && (endpoint.CoveredBy.DirectRead == cliPath || slices.Contains(endpoint.CoveredBy.DirectReads, cliPath)) {
			return true
		}
	}
	return false
}
