package sentry

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"unicode"
)

const (
	sentryMatrixPath     = "sources/sentry-source-lane-matrix.json"
	sentrySourceLockPath = "sources/sentry-operation-source-lock.json"
)

var sentryLaneOrder = []string{
	"direct_read",
	"direct_write",
	"binary_download",
	"binary_upload",
	"etl",
	"reverse_etl",
	"sync_transport",
}

type sentrySourceInfo struct {
	ID             string
	Protocol       string
	Method         string
	Path           string
	OperationID    string
	SourceURL      string
	SourceLocation string
	Operation      map[string]any
}

type sentryActionFact struct {
	Kind            string
	EvidenceField   string
	Headline        string
	SuccessStatuses []string
}

type sentryContinuationFact struct {
	Kind        string
	Parameter   string
	Description string
}

type sentryEventTransportFact struct {
	Kind                string
	CallbackField       string
	EventSelectorField  string
	Headline            string
	RequiredFields      []string
	DocumentedCallbacks int
}

type sentryExtractabilityFact struct {
	Kind             string
	ResponseStatuses []string
}

type sentryArtifactRecords struct {
	API             map[string]struct{}
	Streams         map[string]struct{}
	Operations      map[string]map[string]any
	Commands        map[string]map[string]any
	SurfaceByStream map[string]string
}

func TestSentrySourceLaneMatrixContract(t *testing.T) {
	matrix := readSentryJSONObject(t, sentryMatrixPath)
	lock := readSentryJSONObject(t, sentrySourceLockPath)

	if err := validateSentrySourceLaneMatrix(matrix, lock, readSentryArtifactRecords(t)); err != nil {
		t.Fatal(err)
	}
}

func TestSentrySourceLaneMatrixRejectsHiddenSourceRow(t *testing.T) {
	matrix, lock := readSentryMatrixAndLock(t)
	operations := sentryArrayField(t, matrix, "operations")
	matrix["operations"] = operations[:len(operations)-1]

	assertSentryMatrixValidationError(t, matrix, lock, "source row absent from matrix")
}

func TestSentrySourceLaneMatrixRejectsDuplicateSourceID(t *testing.T) {
	matrix, lock := readSentryMatrixAndLock(t)
	operations := sentryArrayField(t, matrix, "operations")
	matrix["operations"] = append(operations, operations[0])

	assertSentryMatrixValidationError(t, matrix, lock, "duplicate matrix source id")
}

func TestSentrySourceLaneMatrixRejectsInvalidArtifactBacklink(t *testing.T) {
	matrix, lock := readSentryMatrixAndLock(t)
	artifacts := sentryArrayField(t, matrix, "artifacts")
	apiSurface := sentryObjectValue(t, artifacts[0], "api surface artifact")
	links := sentryArrayField(t, apiSurface, "links")
	link := sentryObjectValue(t, links[0], "api surface backlink")
	link["source_id"] = "sentry.rest.NoSuchOperation"

	assertSentryMatrixValidationError(t, matrix, lock, "artifact api_surface.json link")
}

func TestSentrySourceLaneMatrixRejectsMissingSourceFact(t *testing.T) {
	matrix, lock := readSentryMatrixAndLock(t)
	operation := sentryObjectValue(t, sentryArrayField(t, matrix, "operations")[0], "matrix operation")
	facts := sentryObjectField(t, operation, "facts")
	sentryObjectField(t, facts, "scope")["source_location"] = ""

	assertSentryMatrixValidationError(t, matrix, lock, "does not preserve exact scope/path evidence")
}

func TestSentrySourceLaneMatrixRejectsMissingSemanticActionFact(t *testing.T) {
	matrix, lock := readSentryMatrixAndLock(t)
	operation := sentryObjectValue(t, sentryArrayField(t, matrix, "operations")[0], "matrix operation")
	delete(sentryObjectField(t, operation, "facts"), "action")

	assertSentryMatrixValidationError(t, matrix, lock, "semantic action/response evidence")
}

func TestSentrySourceLaneMatrixRejectsMissingDocumentedContinuationETLDisposition(t *testing.T) {
	matrix, lock := readSentryMatrixAndLock(t)
	sources := sentrySourceOperationsByID(t, lock)
	for _, rawOperation := range sentryArrayField(t, matrix, "operations") {
		operation := sentryObjectValue(t, rawOperation, "matrix operation")
		source := sources[sentryStringField(t, operation, "source_id")]
		if !sentryRequiresETL(source) {
			continue
		}
		sentryMatrixCellByLane(t, operation, "etl")["state"] = "not_applicable"
		assertSentryMatrixValidationError(t, matrix, lock, "documented continuation source operation")
		return
	}
	t.Fatal("matrix has no documented continuation source operation")
}

func TestSentrySourceLaneMatrixRejectsMissingMutationReverseETLDisposition(t *testing.T) {
	matrix, lock := readSentryMatrixAndLock(t)
	sources := sentrySourceOperationsByID(t, lock)
	for _, rawOperation := range sentryArrayField(t, matrix, "operations") {
		operation := sentryObjectValue(t, rawOperation, "matrix operation")
		source := sources[sentryStringField(t, operation, "source_id")]
		if sentrySourceAction(source).Kind != "mutation" {
			continue
		}
		sentryMatrixCellByLane(t, operation, "reverse_etl")["state"] = "not_applicable"
		assertSentryMatrixValidationError(t, matrix, lock, "semantic mutation source operation")
		return
	}
	t.Fatal("matrix has no semantic mutation source operation")
}

func TestSentrySourceLaneMatrixRejectsArrayOrPageSizeOnlyETLDisposition(t *testing.T) {
	matrix, lock := readSentryMatrixAndLock(t)
	sources := sentrySourceOperationsByID(t, lock)
	for _, rawOperation := range sentryArrayField(t, matrix, "operations") {
		operation := sentryObjectValue(t, rawOperation, "matrix operation")
		source := sources[sentryStringField(t, operation, "source_id")]
		if sentrySourceAction(source).Kind != "read" || sentryRequiresETL(source) ||
			(sentrySourceExtractability(source).Kind != "json_array_response" && !sentrySourceHasPageSizeOnlyControl(source)) {
			continue
		}
		sentryMatrixCellByLane(t, operation, "etl")["state"] = "mapped_unproven"
		assertSentryMatrixValidationError(t, matrix, lock, "etl cell")
		return
	}
	t.Fatal("matrix has no array-only or page-size-only read")
}

func TestSentrySourceLaneMatrixRejectsPageSizeOnlyETLDisposition(t *testing.T) {
	matrix, lock := readSentryMatrixAndLock(t)
	sources := sentrySourceOperationsByID(t, lock)
	for _, rawOperation := range sentryArrayField(t, matrix, "operations") {
		operation := sentryObjectValue(t, rawOperation, "matrix operation")
		source := sources[sentryStringField(t, operation, "source_id")]
		if sentrySourceAction(source).Kind != "read" || sentryRequiresETL(source) || !sentrySourceHasPageSizeOnlyControl(source) {
			continue
		}
		sentryMatrixCellByLane(t, operation, "etl")["state"] = "mapped_unproven"
		assertSentryMatrixValidationError(t, matrix, lock, "etl cell")
		return
	}
	t.Fatal("matrix has no page-size-only read")
}

func TestSentrySourceLaneMatrixRejectsPaginationAsSyncTransport(t *testing.T) {
	matrix, lock := readSentryMatrixAndLock(t)
	sources := sentrySourceOperationsByID(t, lock)
	for _, rawOperation := range sentryArrayField(t, matrix, "operations") {
		operation := sentryObjectValue(t, rawOperation, "matrix operation")
		source := sources[sentryStringField(t, operation, "source_id")]
		if !sentryRequiresETL(source) || sentryRequiresSyncTransport(source) {
			continue
		}
		sentryMatrixCellByLane(t, operation, "sync_transport")["state"] = "mapped_unproven"
		assertSentryMatrixValidationError(t, matrix, lock, "sync_transport cell")
		return
	}
	t.Fatal("matrix has no continuation read without a webhook registration contract")
}

func TestSentrySourceLaneMatrixRejectsMissingWebhookRegistrationSyncTransport(t *testing.T) {
	matrix, lock := readSentryMatrixAndLock(t)
	sources := sentrySourceOperationsByID(t, lock)
	for _, rawOperation := range sentryArrayField(t, matrix, "operations") {
		operation := sentryObjectValue(t, rawOperation, "matrix operation")
		source := sources[sentryStringField(t, operation, "source_id")]
		if !sentryRequiresSyncTransport(source) {
			continue
		}
		sentryMatrixCellByLane(t, operation, "sync_transport")["state"] = "not_applicable"
		assertSentryMatrixValidationError(t, matrix, lock, "webhook registration source operation")
		return
	}
	t.Fatal("matrix has no source-backed webhook registration contract")
}

func TestSentrySourceLaneMatrixRejectsSourceCountMismatch(t *testing.T) {
	matrix, lock := readSentryMatrixAndLock(t)
	sentryObjectField(t, matrix, "source_lock")["source_operation_count"] = float64(222)

	assertSentryMatrixValidationError(t, matrix, lock, "source lock operation count")
}

func TestSentrySourceLaneMatrixPreservesPagingBinaryAndDeleteSurface(t *testing.T) {
	matrix, lock := readSentryMatrixAndLock(t)
	if err := validateSentrySourceLaneMatrix(matrix, lock, readSentryArtifactRecords(t)); err != nil {
		t.Fatal(err)
	}

	counts := sentryMatrixCellCounts(matrix)
	if got := counts["etl"]["mapped_unproven"]; got != 45 {
		t.Fatalf("etl mapped_unproven cells = %d, want 45", got)
	}
	if got := counts["sync_transport"]["mapped_unproven"]; got != 1 {
		t.Fatalf("sync_transport mapped_unproven cells = %d, want 1", got)
	}
	if got := counts["binary_upload"]["mapped_unproven"]; got != 2 {
		t.Fatalf("binary_upload mapped_unproven cells = %d, want 2", got)
	}
	if got := counts["direct_write"]["mapped_unproven"]; got != 103 {
		t.Fatalf("direct_write mapped_unproven cells = %d, want 103", got)
	}
	if got := counts["reverse_etl"]["mapped_unproven"]; got != 103 {
		t.Fatalf("reverse_etl mapped_unproven cells = %d, want 103", got)
	}
}

func TestSentrySemanticLanePredicatesUseSourceContracts(t *testing.T) {
	tests := []struct {
		name   string
		source sentrySourceInfo
		want   map[string][2]string
	}{
		{
			name: "semantic read is not selected by HTTP method",
			source: sentryTestSource("POST", "Query records", map[string]any{
				"responses": map[string]any{"200": sentryTestJSONResponse("object")},
			}),
			want: map[string][2]string{
				"direct_read":  {"mapped_unproven", "sentry.source.direct_read.semantic_read_success_response.v2"},
				"direct_write": {"not_applicable", "sentry.source.direct_write.semantic_read_not_applicable.v2"},
				"reverse_etl":  {"not_applicable", "sentry.source.reverse_etl.semantic_read_not_applicable.v2"},
			},
		},
		{
			name: "semantic mutation is not selected by HTTP method",
			source: sentryTestSource("GET", "Delete record", map[string]any{
				"responses": map[string]any{"204": map[string]any{"description": "Deleted"}},
			}),
			want: map[string][2]string{
				"direct_read":  {"not_applicable", "sentry.source.direct_read.semantic_mutation_not_applicable.v2"},
				"direct_write": {"mapped_unproven", "sentry.source.direct_write.semantic_mutation_success_response.v2"},
				"reverse_etl":  {"mapped_unproven", "sentry.source.reverse_etl.semantic_mutation_success_response.v2"},
			},
		},
		{
			name: "ETL requires source-described continuation independent of parameter or schema names",
			source: sentryTestSource("POST", "Query records", map[string]any{
				"parameters": []any{map[string]any{
					"name": "page_token", "in": "query", "description": "Use this token to retrieve the next page of results.", "schema": map[string]any{"type": "string"},
				}},
				"responses": map[string]any{"200": sentryTestJSONResponse("object")},
			}),
			want: map[string][2]string{
				"etl": {"mapped_unproven", "sentry.source.etl.documented_continuation_read.v2"},
			},
		},
		{
			name: "JSON array and page size alone are not continuation",
			source: sentryTestSource("GET", "List records", map[string]any{
				"parameters": []any{map[string]any{
					"name": "limit", "in": "query", "description": "Maximum records per page.", "schema": map[string]any{"type": "integer"},
				}},
				"responses": map[string]any{"200": sentryTestJSONResponse("array")},
			}),
			want: map[string][2]string{
				"etl": {"not_applicable", "sentry.source.etl.no_documented_continuation_read.v2"},
			},
		},
		{
			name: "webhook registration is sync but a collection list is not",
			source: sentryTestSource("POST", "Register event webhook", map[string]any{
				"requestBody": map[string]any{"content": map[string]any{"application/json": map[string]any{"schema": map[string]any{
					"type":     "object",
					"required": []any{"callback_endpoint", "event_types"},
					"properties": map[string]any{
						"callback_endpoint": map[string]any{"type": "string", "description": "Webhook callback URL."},
						"event_types":       map[string]any{"type": "array", "description": "Events to subscribe to."},
					},
				}}}},
				"responses": map[string]any{"201": sentryTestJSONResponse("object")},
			}),
			want: map[string][2]string{
				"sync_transport": {"mapped_unproven", "sentry.source.sync_transport.webhook_registration_contract.v2"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := sentryExpectedCells(test.source)
			if err != nil {
				t.Fatalf("sentryExpectedCells() error = %v", err)
			}
			for lane, want := range test.want {
				if got[lane] != want {
					t.Fatalf("%s = %v, want %v", lane, got[lane], want)
				}
			}
		})
	}
}

func TestSentryWebhookRegistrationRequiresRequiredCallbackAndEventFields(t *testing.T) {
	source := sentryTestSource("POST", "Register event webhook", map[string]any{
		"requestBody": map[string]any{"content": map[string]any{"application/json": map[string]any{"schema": map[string]any{
			"type": "object",
			"properties": map[string]any{
				"callback_endpoint": map[string]any{"type": "string", "description": "Webhook callback URL."},
				"event_types":       map[string]any{"type": "array", "description": "Events to subscribe to."},
			},
		}}}},
		"responses": map[string]any{"201": sentryTestJSONResponse("object")},
	})

	got, err := sentryExpectedCells(source)
	if err != nil {
		t.Fatalf("sentryExpectedCells() error = %v", err)
	}
	if want := [2]string{"not_applicable", "sentry.source.sync_transport.no_webhook_registration_contract.v2"}; got["sync_transport"] != want {
		t.Fatalf("sync_transport = %v, want %v", got["sync_transport"], want)
	}
}

func sentryTestSource(method, headline string, operation map[string]any) sentrySourceInfo {
	operation["summary"] = headline
	return sentrySourceInfo{
		ID:             "sentry.rest.test",
		Method:         method,
		OperationID:    headline,
		SourceLocation: "paths.test",
		SourceURL:      "https://example.invalid/sentry",
		Operation:      operation,
	}
}

func sentryTestJSONResponse(schemaType string) map[string]any {
	return map[string]any{
		"description": "Success",
		"content": map[string]any{
			"application/json": map[string]any{"schema": map[string]any{"type": schemaType}},
		},
	}
}

func readSentryMatrixAndLock(t *testing.T) (map[string]any, map[string]any) {
	t.Helper()
	return readSentryJSONObject(t, sentryMatrixPath), readSentryJSONObject(t, sentrySourceLockPath)
}

func assertSentryMatrixValidationError(t *testing.T, matrix, lock map[string]any, want string) {
	t.Helper()
	err := validateSentrySourceLaneMatrix(matrix, lock, readSentryArtifactRecords(t))
	if err == nil {
		t.Fatalf("matrix validation unexpectedly passed, want error containing %q", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("matrix validation error = %q, want substring %q", err, want)
	}
}

func validateSentrySourceLaneMatrix(matrix, lock map[string]any, artifacts sentryArtifactRecords) error {
	if sentryNumberField(matrix, "schema_version") != 1 {
		return fmt.Errorf("matrix schema_version = %d, want 1", sentryNumberField(matrix, "schema_version"))
	}
	if sentryStringField(nil, matrix, "connector") != "sentry" {
		return fmt.Errorf("matrix connector = %q, want sentry", sentryStringField(nil, matrix, "connector"))
	}
	if !sentryEqualStrings(sentryStringArrayField(nil, matrix, "lanes"), sentryLaneOrder) {
		return fmt.Errorf("matrix lanes = %v, want %v", sentryStringArrayField(nil, matrix, "lanes"), sentryLaneOrder)
	}

	sources, err := sentrySourceOperationsByIDNoTest(lock)
	if err != nil {
		return err
	}
	if err := sentryValidateSourceLockReference(sentryObjectField(nil, matrix, "source_lock"), lock, len(sources)); err != nil {
		return err
	}

	matrixByID := make(map[string]map[string]any, len(sources))
	for _, rawOperation := range sentryArrayField(nil, matrix, "operations") {
		operation := sentryObjectValue(nil, rawOperation, "matrix operation")
		sourceID := sentryStringField(nil, operation, "source_id")
		if _, exists := matrixByID[sourceID]; exists {
			return fmt.Errorf("duplicate matrix source id %q", sourceID)
		}
		matrixByID[sourceID] = operation
	}
	for _, sourceID := range sentrySortedMapKeys(sources) {
		source := sources[sourceID]
		operation, exists := matrixByID[sourceID]
		if !exists {
			return fmt.Errorf("source row absent from matrix: %s", sourceID)
		}
		if err := sentryValidateOperation(operation, source); err != nil {
			return err
		}
	}
	for sourceID := range matrixByID {
		if _, exists := sources[sourceID]; !exists {
			return fmt.Errorf("matrix source id %q is not in the source lock", sourceID)
		}
	}
	if len(matrixByID) != len(sources) {
		return fmt.Errorf("matrix operation count = %d, source lock count = %d", len(matrixByID), len(sources))
	}
	if err := sentryValidateCountReconciliation(matrix, lock, sources); err != nil {
		return err
	}
	return sentryValidateArtifactLinks(matrix, artifacts, matrixByID, sources)
}

func sentryValidateSourceLockReference(reference, lock map[string]any, sourceCount int) error {
	rest := sentryObjectField(nil, lock, "rest")
	if sentryStringField(nil, reference, "path") != sentrySourceLockPath {
		return fmt.Errorf("matrix source lock path = %q, want %q", sentryStringField(nil, reference, "path"), sentrySourceLockPath)
	}
	if sentryStringField(nil, reference, "source_url") != sentryStringField(nil, rest, "source_url") ||
		sentryStringField(nil, reference, "sha256") != sentryStringField(nil, rest, "sha256") ||
		sentryStringField(nil, reference, "captured_at") != sentryStringField(nil, lock, "captured_at") {
		return errors.New("matrix source lock identity does not match the pinned Sentry source")
	}
	counts := sentryObjectField(nil, lock, "counts")
	if count := sentryNumberField(reference, "source_operation_count"); count != sourceCount ||
		count != sentryNumberField(counts, "rest") || count != sentryNumberField(counts, "total") {
		return fmt.Errorf("source lock operation count = %d, want %d", count, sourceCount)
	}
	return nil
}

func sentryValidateOperation(operation map[string]any, source sentrySourceInfo) error {
	if sentryStringField(nil, operation, "protocol") != source.Protocol ||
		sentryStringField(nil, operation, "method") != source.Method ||
		sentryStringField(nil, operation, "path") != source.Path ||
		sentryStringField(nil, operation, "operation_id") != source.OperationID ||
		sentryStringField(nil, operation, "source_url") != source.SourceURL {
		return fmt.Errorf("source operation %s does not preserve the locked operation identity", source.ID)
	}
	if location := sentryStringField(nil, operation, "source_location"); location == "" || location != source.SourceLocation {
		return fmt.Errorf("source operation %s does not preserve its cited source location", source.ID)
	}
	if err := sentryValidateFacts(sentryObjectField(nil, operation, "facts"), source); err != nil {
		return err
	}

	expected, err := sentryExpectedCells(source)
	if err != nil {
		return err
	}
	cells := sentryArrayField(nil, operation, "cells")
	if len(cells) != len(sentryLaneOrder) {
		return fmt.Errorf("source operation %s has %d cells, want %d", source.ID, len(cells), len(sentryLaneOrder))
	}
	seen := make(map[string]struct{}, len(cells))
	for _, rawCell := range cells {
		cell := sentryObjectValue(nil, rawCell, "matrix cell")
		lane := sentryStringField(nil, cell, "lane")
		if _, duplicate := seen[lane]; duplicate {
			return fmt.Errorf("source operation %s has duplicate %s cell", source.ID, lane)
		}
		seen[lane] = struct{}{}
		want, exists := expected[lane]
		if !exists {
			return fmt.Errorf("source operation %s has unknown lane %q", source.ID, lane)
		}
		state, reason := sentryStringField(nil, cell, "state"), sentryStringField(nil, cell, "reason")
		if state != want[0] || reason != want[1] {
			if lane == "etl" && sentryRequiresETL(source) {
				return fmt.Errorf("documented continuation source operation %s has %s cell = %s/%s, want %s/%s", source.ID, lane, state, reason, want[0], want[1])
			}
			if lane == "reverse_etl" && sentrySourceAction(source).Kind == "mutation" {
				return fmt.Errorf("semantic mutation source operation %s has %s cell = %s/%s, want %s/%s", source.ID, lane, state, reason, want[0], want[1])
			}
			if lane == "sync_transport" && sentryRequiresSyncTransport(source) {
				return fmt.Errorf("webhook registration source operation %s has %s cell = %s/%s, want %s/%s", source.ID, lane, state, reason, want[0], want[1])
			}
			return fmt.Errorf("source operation %s has %s cell = %s/%s, want %s/%s", source.ID, lane, state, reason, want[0], want[1])
		}
		if sentryStringField(nil, cell, "source_location") != source.SourceLocation {
			return fmt.Errorf("source operation %s has uncited %s cell", source.ID, lane)
		}
		if sentryStringField(nil, cell, "source_url") != source.SourceURL {
			return fmt.Errorf("source operation %s has an uncited %s cell source URL", source.ID, lane)
		}
		if state != "mapped_unproven" && state != "not_applicable" {
			return fmt.Errorf("source operation %s has unsupported mapping-only state %q", source.ID, state)
		}
	}
	for _, lane := range sentryLaneOrder {
		if _, exists := seen[lane]; !exists {
			return fmt.Errorf("source operation %s has no %s cell", source.ID, lane)
		}
	}
	return nil
}

func sentryValidateFacts(facts map[string]any, source sentrySourceInfo) error {
	action := sentrySourceAction(source)
	if action.Kind == "unknown" {
		return fmt.Errorf("source operation %s has no summary/action evidence", source.ID)
	}
	gotAction := sentryObjectFieldOrEmpty(facts, "action")
	if sentryStringFieldOrEmpty(gotAction, "kind") != action.Kind ||
		sentryStringFieldOrEmpty(gotAction, "evidence_field") != action.EvidenceField ||
		sentryStringFieldOrEmpty(gotAction, "headline") != action.Headline ||
		!sentryEqualStrings(sentryStringArrayFieldOrEmpty(gotAction, "success_statuses"), action.SuccessStatuses) ||
		sentryStringFieldOrEmpty(gotAction, "source_location") != source.SourceLocation {
		return fmt.Errorf("source operation %s does not preserve exact semantic action/response evidence", source.ID)
	}

	continuation := sentrySourceContinuation(source)
	gotContinuation := sentryObjectFieldOrEmpty(facts, "continuation")
	if sentryStringFieldOrEmpty(gotContinuation, "kind") != continuation.Kind ||
		sentryStringFieldOrEmpty(gotContinuation, "parameter") != continuation.Parameter ||
		sentryStringFieldOrEmpty(gotContinuation, "description") != continuation.Description ||
		sentryStringFieldOrEmpty(gotContinuation, "source_location") != source.SourceLocation {
		return fmt.Errorf("source operation %s does not preserve exact continuation evidence", source.ID)
	}

	pathParameters, requiredPath, queryParameters, requiredQuery := sentrySourceScope(source)
	scope := sentryObjectField(nil, facts, "scope")
	if !sentryEqualStrings(sentryStringArrayField(nil, scope, "path_parameters"), pathParameters) ||
		!sentryEqualStrings(sentryStringArrayField(nil, scope, "required_path_parameters"), requiredPath) ||
		!sentryEqualStrings(sentryStringArrayField(nil, scope, "query_parameters"), queryParameters) ||
		!sentryEqualStrings(sentryStringArrayField(nil, scope, "required_query_parameters"), requiredQuery) ||
		sentryStringField(nil, scope, "source_location") != source.SourceLocation {
		return fmt.Errorf("source operation %s does not preserve exact scope/path evidence", source.ID)
	}

	requestMedia, responseMedia := sentrySourceMedia(source)
	media := sentryObjectField(nil, facts, "media")
	if !sentryEqualStrings(sentryStringArrayField(nil, media, "request"), requestMedia) ||
		!sentryEqualStrings(sentryStringArrayField(nil, media, "response"), responseMedia) ||
		sentryStringField(nil, media, "source_location") != source.SourceLocation {
		return fmt.Errorf("source operation %s does not preserve exact media evidence", source.ID)
	}

	eventTransport := sentrySourceEventTransport(source)
	gotEventTransport := sentryObjectFieldOrEmpty(facts, "event_transport")
	if sentryStringFieldOrEmpty(gotEventTransport, "kind") != eventTransport.Kind ||
		sentryStringFieldOrEmpty(gotEventTransport, "callback_field") != eventTransport.CallbackField ||
		sentryStringFieldOrEmpty(gotEventTransport, "event_selector_field") != eventTransport.EventSelectorField ||
		sentryStringFieldOrEmpty(gotEventTransport, "headline") != eventTransport.Headline ||
		!sentryEqualStrings(sentryStringArrayFieldOrEmpty(gotEventTransport, "required_fields"), eventTransport.RequiredFields) ||
		sentryNumberFieldOrZero(gotEventTransport, "documented_callbacks") != eventTransport.DocumentedCallbacks ||
		sentryStringFieldOrEmpty(gotEventTransport, "source_location") != source.SourceLocation {
		return fmt.Errorf("source operation %s does not preserve its webhook/event-registration evidence", source.ID)
	}

	extractability := sentrySourceExtractability(source)
	gotExtractability := sentryObjectField(nil, facts, "extractability")
	if sentryStringField(nil, gotExtractability, "kind") != extractability.Kind ||
		!sentryEqualStrings(sentryStringArrayField(nil, gotExtractability, "response_statuses"), extractability.ResponseStatuses) ||
		sentryStringField(nil, gotExtractability, "source_location") != source.SourceLocation {
		return fmt.Errorf("source operation %s does not preserve exact collection-response evidence", source.ID)
	}
	return nil
}

func sentryExpectedCells(source sentrySourceInfo) (map[string][2]string, error) {
	action := sentrySourceAction(source)
	if action.Kind == "unknown" {
		return nil, fmt.Errorf("source operation %s has no summary/action evidence", source.ID)
	}
	hasSuccessResponse := len(action.SuccessStatuses) > 0
	isBinaryResponse := sentrySourceHasBinaryResponse(source)
	isMultipart := sentrySourceHasRequestMedia(source, "multipart/form-data")
	requiresETL := sentryRequiresETL(source)
	requiresSyncTransport := sentryRequiresSyncTransport(source)

	expected := map[string][2]string{}
	if action.Kind == "read" {
		if hasSuccessResponse {
			expected["direct_read"] = [2]string{"mapped_unproven", "sentry.source.direct_read.semantic_read_success_response.v2"}
		} else {
			expected["direct_read"] = [2]string{"not_applicable", "sentry.source.direct_read.no_documented_success_response.v2"}
		}
		expected["direct_write"] = [2]string{"not_applicable", "sentry.source.direct_write.semantic_read_not_applicable.v2"}
		expected["reverse_etl"] = [2]string{"not_applicable", "sentry.source.reverse_etl.semantic_read_not_applicable.v2"}
	} else {
		expected["direct_read"] = [2]string{"not_applicable", "sentry.source.direct_read.semantic_mutation_not_applicable.v2"}
		if hasSuccessResponse {
			expected["direct_write"] = [2]string{"mapped_unproven", "sentry.source.direct_write.semantic_mutation_success_response.v2"}
			expected["reverse_etl"] = [2]string{"mapped_unproven", "sentry.source.reverse_etl.semantic_mutation_success_response.v2"}
		} else {
			expected["direct_write"] = [2]string{"not_applicable", "sentry.source.direct_write.no_documented_success_response.v2"}
			expected["reverse_etl"] = [2]string{"not_applicable", "sentry.source.reverse_etl.no_documented_success_response.v2"}
		}
	}
	if isBinaryResponse {
		expected["binary_download"] = [2]string{"mapped_unproven", "sentry.source.binary_download.binary_response_media.v1"}
	} else {
		expected["binary_download"] = [2]string{"not_applicable", "sentry.source.binary_download.no_binary_response_media.v1"}
	}
	if isMultipart {
		expected["binary_upload"] = [2]string{"mapped_unproven", "sentry.source.binary_upload.multipart_request_media.v1"}
	} else {
		expected["binary_upload"] = [2]string{"not_applicable", "sentry.source.binary_upload.no_multipart_request_media.v1"}
	}
	if requiresETL {
		expected["etl"] = [2]string{"mapped_unproven", "sentry.source.etl.documented_continuation_read.v2"}
	} else {
		expected["etl"] = [2]string{"not_applicable", "sentry.source.etl.no_documented_continuation_read.v2"}
	}
	if requiresSyncTransport {
		expected["sync_transport"] = [2]string{"mapped_unproven", "sentry.source.sync_transport.webhook_registration_contract.v2"}
	} else {
		expected["sync_transport"] = [2]string{"not_applicable", "sentry.source.sync_transport.no_webhook_registration_contract.v2"}
	}
	return expected, nil
}

func sentryValidateCountReconciliation(matrix, lock map[string]any, sources map[string]sentrySourceInfo) error {
	if len(sentryArrayField(nil, matrix, "operations")) != 223 || len(sources) != 223 {
		return fmt.Errorf("retained source count reconciliation = matrix:%d lock:%d, want 223", len(sentryArrayField(nil, matrix, "operations")), len(sources))
	}

	var semanticReadCount, mutationCount, deleteCount, continuationCount, pageSizeOnlyCount, collectionReadCount, etlCount, syncTransportCount, multipartCount, documentedCallbackCount int
	multipartIDs := make(map[string]struct{})
	for _, source := range sources {
		action := sentrySourceAction(source)
		if action.Kind == "read" {
			semanticReadCount++
		}
		if action.Kind == "mutation" {
			mutationCount++
		}
		if source.Method == "DELETE" {
			deleteCount++
		}
		if sentrySourceContinuation(source).Kind != "not_documented" {
			continuationCount++
		}
		if sentrySourceHasPageSizeOnlyControl(source) {
			pageSizeOnlyCount++
		}
		if action.Kind == "read" && sentrySourceExtractability(source).Kind == "json_array_response" {
			collectionReadCount++
		}
		if sentryRequiresETL(source) {
			etlCount++
		}
		if sentryRequiresSyncTransport(source) {
			syncTransportCount++
		}
		documentedCallbackCount += sentrySourceEventTransport(source).DocumentedCallbacks
		if sentrySourceHasRequestMedia(source, "multipart/form-data") {
			multipartCount++
			multipartIDs[source.ID] = struct{}{}
		}
	}
	if semanticReadCount != 120 || mutationCount != 103 || deleteCount != 35 || continuationCount != 45 || pageSizeOnlyCount != 1 || collectionReadCount != 54 || etlCount != 45 || syncTransportCount != 1 || multipartCount != 2 || documentedCallbackCount != 0 {
		return fmt.Errorf("source fact counts = semantic_read:%d mutation:%d delete:%d continuation:%d page_size_only:%d collection_read:%d etl:%d sync_transport:%d multipart:%d documented_callbacks:%d", semanticReadCount, mutationCount, deleteCount, continuationCount, pageSizeOnlyCount, collectionReadCount, etlCount, syncTransportCount, multipartCount, documentedCallbackCount)
	}
	if !sentryEqualStringSet(multipartIDs, map[string]struct{}{
		"sentry.rest.uploadOrganizationReleaseFile": {},
		"sentry.rest.uploadProjectReleaseFile":      {},
	}) {
		return errors.New("Sentry source facts no longer match the cited multipart operation set")
	}

	cells := sentryMatrixCellCounts(matrix)
	wantMapped := map[string]int{
		"direct_read":     120,
		"direct_write":    103,
		"binary_download": 0,
		"binary_upload":   2,
		"etl":             45,
		"reverse_etl":     103,
		"sync_transport":  1,
	}
	mappedTotal, notApplicableTotal := 0, 0
	for _, lane := range sentryLaneOrder {
		if cells[lane]["mapped_unproven"] != wantMapped[lane] {
			return fmt.Errorf("%s mapped_unproven count = %d, want %d", lane, cells[lane]["mapped_unproven"], wantMapped[lane])
		}
		if cells[lane]["not_applicable"] != 223-wantMapped[lane] {
			return fmt.Errorf("%s not_applicable count = %d, want %d", lane, cells[lane]["not_applicable"], 223-wantMapped[lane])
		}
		if cells[lane]["implemented"] != 0 || cells[lane]["missing_foundation"] != 0 {
			return fmt.Errorf("%s contains an execution or foundation state", lane)
		}
		mappedTotal += cells[lane]["mapped_unproven"]
		notApplicableTotal += cells[lane]["not_applicable"]
	}
	if mappedTotal != 374 || notApplicableTotal != 1187 || mappedTotal+notApplicableTotal != 1561 {
		return fmt.Errorf("matrix cell totals = mapped:%d not_applicable:%d total:%d, want 374/1187/1561", mappedTotal, notApplicableTotal, mappedTotal+notApplicableTotal)
	}
	return nil
}

func sentryValidateArtifactLinks(matrix map[string]any, records sentryArtifactRecords, matrixByID map[string]map[string]any, sources map[string]sentrySourceInfo) error {
	byPath := make(map[string]map[string]any)
	for _, rawArtifact := range sentryArrayField(nil, matrix, "artifacts") {
		artifact := sentryObjectValue(nil, rawArtifact, "matrix artifact")
		path := sentryStringField(nil, artifact, "path")
		if _, exists := byPath[path]; exists {
			return fmt.Errorf("duplicate artifact matrix %q", path)
		}
		byPath[path] = artifact
	}
	if len(byPath) != 4 {
		return fmt.Errorf("matrix artifact count = %d, want 4", len(byPath))
	}
	if err := sentryValidateAPISurfaceLinks(byPath["api_surface.json"], records, matrixByID, sources); err != nil {
		return err
	}
	if err := sentryValidateStreamLinks(byPath["streams.json"], records, matrixByID); err != nil {
		return err
	}
	if err := sentryValidateOperationLinks(byPath["operations.json"], records, matrixByID, sources); err != nil {
		return err
	}
	if err := sentryValidateCommandLinks(byPath["cli_surface.json"], records, matrixByID); err != nil {
		return err
	}
	return sentryValidateBacklinkGaps(matrix, records, sources)
}

func sentryValidateAPISurfaceLinks(artifact map[string]any, records sentryArtifactRecords, matrixByID map[string]map[string]any, sources map[string]sentrySourceInfo) error {
	if artifact == nil {
		return errors.New("matrix is missing api_surface.json artifact")
	}
	links := sentryArrayField(nil, artifact, "links")
	unlinked := sentryStringReasonMap(artifact, "unlinked_records")
	if sentryNumberField(artifact, "record_count") != len(records.API) || len(links)+len(unlinked) != len(records.API) {
		return fmt.Errorf("artifact api_surface.json link count = records:%d links:%d unlinked:%d, want %d", sentryNumberField(artifact, "record_count"), len(links), len(unlinked), len(records.API))
	}
	seen := make(map[string]struct{}, len(links))
	for _, rawLink := range links {
		link := sentryObjectValue(nil, rawLink, "api surface backlink")
		record := sentryStringField(nil, link, "record")
		if _, exists := records.API[record]; !exists {
			return fmt.Errorf("artifact api_surface.json link references nonexistent record %q", record)
		}
		if _, duplicate := seen[record]; duplicate {
			return fmt.Errorf("artifact api_surface.json link duplicates record %q", record)
		}
		seen[record] = struct{}{}
		sourceID := sentryStringField(nil, link, "source_id")
		operation, exists := matrixByID[sourceID]
		if !exists {
			return fmt.Errorf("artifact api_surface.json link %q references nonexistent source cell owner %q", record, sourceID)
		}
		if record != sentryStringField(nil, operation, "method")+" "+sentryStringField(nil, operation, "path") {
			return fmt.Errorf("artifact api_surface.json link %q does not preserve source route", record)
		}
		source, exists := sources[sourceID]
		if !exists {
			return fmt.Errorf("artifact api_surface.json link %q references nonexistent source %q", record, sourceID)
		}
		lane := "direct_write"
		if sentrySourceAction(source).Kind == "read" {
			lane = "direct_read"
		}
		if !sentryEqualStrings(sentryStringArrayField(nil, link, "lanes"), []string{lane}) || sentryMatrixCellByLaneNoTest(operation, lane) == nil {
			return fmt.Errorf("artifact api_surface.json link %q references nonexistent cell %q", record, lane)
		}
	}
	for record, reason := range unlinked {
		if _, exists := records.API[record]; !exists || reason != "sentry.source.backlink.no_exact_source_lock_operation.v1" {
			return fmt.Errorf("artifact api_surface.json unlinked record %q is not an explicit source-information gap", record)
		}
		if _, linked := seen[record]; linked {
			return fmt.Errorf("artifact api_surface.json record %q is both linked and unlinked", record)
		}
	}
	return nil
}

func sentryValidateStreamLinks(artifact map[string]any, records sentryArtifactRecords, matrixByID map[string]map[string]any) error {
	expected := map[string]string{
		"issues":   "sentry.rest.List a Project's Issues",
		"events":   "sentry.rest.listProjectEvents",
		"releases": "sentry.rest.listOrganizationReleases",
	}
	if artifact == nil {
		return errors.New("matrix is missing streams.json artifact")
	}
	links := sentryArrayField(nil, artifact, "links")
	unlinked := sentryStringReasonMap(artifact, "unlinked_records")
	if sentryNumberField(artifact, "record_count") != len(records.Streams) || len(links) != len(expected) || len(links)+len(unlinked) != len(records.Streams) {
		return errors.New("artifact streams.json link count does not reconcile")
	}
	seen := make(map[string]struct{}, len(links))
	for _, rawLink := range links {
		link := sentryObjectValue(nil, rawLink, "stream backlink")
		record := sentryStringField(nil, link, "record")
		if _, exists := records.Streams[record]; !exists {
			return fmt.Errorf("artifact streams.json link references nonexistent record %q", record)
		}
		if _, duplicate := seen[record]; duplicate {
			return fmt.Errorf("artifact streams.json link duplicates record %q", record)
		}
		seen[record] = struct{}{}
		if sentryStringField(nil, link, "source_id") != expected[record] {
			return fmt.Errorf("artifact streams.json link %q has incorrect source id", record)
		}
		operation := matrixByID[expected[record]]
		if operation == nil {
			return fmt.Errorf("artifact streams.json link %q references nonexistent source cell owner", record)
		}
		if records.SurfaceByStream[record] != "GET "+sentryStringField(nil, operation, "path") {
			return fmt.Errorf("artifact streams.json link %q does not retain the exact api-surface route", record)
		}
		for _, lane := range []string{"direct_read", "etl"} {
			if sentryMatrixCellByLaneNoTest(operation, lane) == nil {
				return fmt.Errorf("artifact streams.json link %q references nonexistent cell %q", record, lane)
			}
		}
		if !sentryEqualStrings(sentryStringArrayField(nil, link, "lanes"), []string{"direct_read", "etl"}) {
			return fmt.Errorf("artifact streams.json link %q has incorrect lanes", record)
		}
	}
	if reason, exists := unlinked["projects"]; !exists || reason != "sentry.source.backlink.no_exact_source_lock_operation.v1" || len(unlinked) != 1 {
		return errors.New("artifact streams.json must retain projects as an explicit source-information gap")
	}
	return nil
}

func sentryValidateOperationLinks(artifact map[string]any, records sentryArtifactRecords, matrixByID map[string]map[string]any, sources map[string]sentrySourceInfo) error {
	const record = "sentry.seer_models_list"
	const sourceID = "sentry.rest.listSeerModels"
	if artifact == nil {
		return errors.New("matrix is missing operations.json artifact")
	}
	if sentryNumberField(artifact, "record_count") != len(records.Operations) || len(sentryArrayField(nil, artifact, "links")) != 1 {
		return errors.New("artifact operations.json link count does not reconcile")
	}
	link := sentryObjectValue(nil, sentryArrayField(nil, artifact, "links")[0], "operation backlink")
	if _, exists := records.Operations[record]; !exists || sentryStringField(nil, link, "record") != record || sentryStringField(nil, link, "source_id") != sourceID {
		return errors.New("artifact operations.json link does not retain the Seer Models source binding")
	}
	operation := matrixByID[sourceID]
	if operation == nil || !sentryEqualStrings(sentryStringArrayField(nil, link, "lanes"), []string{"direct_read"}) || sentryMatrixCellByLaneNoTest(operation, "direct_read") == nil {
		return errors.New("artifact operations.json link references nonexistent source cell")
	}
	actual := records.Operations[record]
	sourceOperation := sentryObjectField(nil, actual, "source_operation")
	source := sources[sourceID]
	if sentryStringField(nil, sourceOperation, "id") != sourceID || sentryStringField(nil, sourceOperation, "method") != source.Method || sentryStringField(nil, sourceOperation, "path") != source.Path {
		return errors.New("artifact operations.json source binding does not preserve the locked Seer Models route")
	}
	return sentryValidateClosedRoutePrecedent(link)
}

func sentryValidateCommandLinks(artifact map[string]any, records sentryArtifactRecords, matrixByID map[string]map[string]any) error {
	const record = "seer list-models"
	const sourceID = "sentry.rest.listSeerModels"
	if artifact == nil {
		return errors.New("matrix is missing cli_surface.json artifact")
	}
	if sentryNumberField(artifact, "record_count") != len(records.Commands) || len(sentryArrayField(nil, artifact, "links")) != 1 {
		return errors.New("artifact cli_surface.json link count does not reconcile")
	}
	link := sentryObjectValue(nil, sentryArrayField(nil, artifact, "links")[0], "command backlink")
	actual, exists := records.Commands[record]
	if !exists || sentryStringField(nil, link, "record") != record || sentryStringField(nil, link, "source_id") != sourceID {
		return errors.New("artifact cli_surface.json link does not retain the Seer Models source binding")
	}
	operation := matrixByID[sourceID]
	if operation == nil || !sentryEqualStrings(sentryStringArrayField(nil, link, "lanes"), []string{"direct_read"}) || sentryMatrixCellByLaneNoTest(operation, "direct_read") == nil {
		return errors.New("artifact cli_surface.json link references nonexistent source cell")
	}
	if sentryStringField(nil, actual, "operation") != "sentry.seer_models_list" || sentryStringField(nil, actual, "source_operation") != sourceID {
		return errors.New("artifact cli_surface.json command does not preserve the source operation binding")
	}
	return sentryValidateClosedRoutePrecedent(link)
}

func sentryValidateClosedRoutePrecedent(link map[string]any) error {
	precedent := sentryObjectField(nil, link, "precedent")
	if sentryNumberField(precedent, "issue") != 4365 ||
		sentryStringField(nil, precedent, "status") != "closed" ||
		sentryStringField(nil, precedent, "relationship") != "existing_artifact_route_binding" ||
		sentryStringField(nil, precedent, "source_membership") != "source_lock_only" {
		return errors.New("closed #4365 precedent is not constrained to its existing artifact route binding")
	}
	return nil
}

func sentryValidateBacklinkGaps(matrix map[string]any, records sentryArtifactRecords, sources map[string]sentrySourceInfo) error {
	gaps := sentryArrayField(nil, matrix, "backlink_gaps")
	expected := make(map[string]sentrySourceInfo)
	for sourceID, source := range sources {
		if _, exists := records.API[source.Method+" "+source.Path]; !exists {
			expected[sourceID] = source
		}
	}
	if len(gaps) != len(expected) || len(expected) != 3 {
		return fmt.Errorf("matrix backlink gap count = %d, want %d", len(gaps), len(expected))
	}
	seen := make(map[string]struct{}, len(gaps))
	for _, rawGap := range gaps {
		gap := sentryObjectValue(nil, rawGap, "backlink gap")
		sourceID := sentryStringField(nil, gap, "source_id")
		source, exists := expected[sourceID]
		if !exists || sentryStringField(nil, gap, "artifact_path") != "api_surface.json" ||
			sentryStringField(nil, gap, "reason") != "sentry.source.backlink.no_exact_api_surface_record.v1" ||
			sentryStringField(nil, gap, "source_location") != source.SourceLocation {
			return fmt.Errorf("matrix backlink gap for %q is not source-cited", sourceID)
		}
		if _, duplicate := seen[sourceID]; duplicate {
			return fmt.Errorf("matrix backlink gap duplicates source operation %q", sourceID)
		}
		seen[sourceID] = struct{}{}
	}
	for sourceID := range expected {
		if _, exists := seen[sourceID]; !exists {
			return fmt.Errorf("matrix is missing explicit api-surface backlink gap for %q", sourceID)
		}
	}
	return nil
}

func sentrySourceOperationsByID(t *testing.T, lock map[string]any) map[string]sentrySourceInfo {
	t.Helper()
	operations, err := sentrySourceOperationsByIDNoTest(lock)
	if err != nil {
		t.Fatal(err)
	}
	return operations
}

func sentrySourceOperationsByIDNoTest(lock map[string]any) (map[string]sentrySourceInfo, error) {
	rest := sentryObjectField(nil, lock, "rest")
	sourceURL := sentryStringField(nil, rest, "source_url")
	result := make(map[string]sentrySourceInfo)
	for _, rawOperation := range sentryArrayField(nil, rest, "operations") {
		record := sentryObjectValue(nil, rawOperation, "source lock operation")
		source := sentrySourceInfo{
			ID:             sentryStringField(nil, record, "id"),
			Protocol:       sentryStringField(nil, record, "protocol"),
			Method:         sentryStringField(nil, record, "method"),
			Path:           sentryStringField(nil, record, "path"),
			OperationID:    sentryStringField(nil, record, "operation_id"),
			SourceURL:      sourceURL,
			SourceLocation: sentryStringField(nil, record, "source_location"),
			Operation:      sentryObjectField(nil, record, "source_operation"),
		}
		if source.ID == "" || source.SourceLocation == "" {
			return nil, errors.New("source lock has an uncited operation")
		}
		if _, exists := result[source.ID]; exists {
			return nil, fmt.Errorf("duplicate source lock operation id %q", source.ID)
		}
		result[source.ID] = source
	}
	return result, nil
}

func sentrySourceScope(source sentrySourceInfo) ([]string, []string, []string, []string) {
	path, requiredPath := map[string]struct{}{}, map[string]struct{}{}
	query, requiredQuery := map[string]struct{}{}, map[string]struct{}{}
	for _, rawParameter := range sentryArrayFieldOrEmpty(source.Operation, "parameters") {
		parameter := sentryObjectValue(nil, rawParameter, "source parameter")
		name := sentryStringField(nil, parameter, "name")
		switch sentryStringField(nil, parameter, "in") {
		case "path":
			path[name] = struct{}{}
			if sentryBoolFieldOrFalse(parameter, "required") {
				requiredPath[name] = struct{}{}
			}
		case "query":
			query[name] = struct{}{}
			if sentryBoolFieldOrFalse(parameter, "required") {
				requiredQuery[name] = struct{}{}
			}
		}
	}
	return sentrySortedSet(path), sentrySortedSet(requiredPath), sentrySortedSet(query), sentrySortedSet(requiredQuery)
}

func sentrySourceAction(source sentrySourceInfo) sentryActionFact {
	evidenceField, headline := sentrySourceActionHeadline(source)
	tokens := sentrySemanticTokens(headline)
	for _, token := range tokens {
		if sentryContainsString(sentryMutationActionTerms, token) {
			return sentryActionFact{
				Kind:            "mutation",
				EvidenceField:   evidenceField,
				Headline:        headline,
				SuccessStatuses: sentrySourceSuccessStatuses(source),
			}
		}
		if sentryContainsString(sentryReadActionTerms, token) {
			return sentryActionFact{
				Kind:            "read",
				EvidenceField:   evidenceField,
				Headline:        headline,
				SuccessStatuses: sentrySourceSuccessStatuses(source),
			}
		}
	}
	return sentryActionFact{
		Kind:            "unknown",
		EvidenceField:   evidenceField,
		Headline:        headline,
		SuccessStatuses: sentrySourceSuccessStatuses(source),
	}
}

var sentryReadActionTerms = []string{
	"fetch",
	"get",
	"list",
	"query",
	"resolve",
	"retrieve",
}

var sentryMutationActionTerms = []string{
	"add",
	"create",
	"delete",
	"disable",
	"edit",
	"enable",
	"link",
	"mutate",
	"provision",
	"register",
	"remove",
	"start",
	"submit",
	"unlink",
	"update",
	"upload",
}

func sentrySourceActionHeadline(source sentrySourceInfo) (string, string) {
	if summary := strings.TrimSpace(sentryStringFieldOrEmpty(source.Operation, "summary")); summary != "" {
		return "summary", summary
	}
	return "operation_id", strings.TrimSpace(source.OperationID)
}

func sentrySemanticTokens(value string) []string {
	return strings.FieldsFunc(strings.ToLower(value), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func sentrySourceSuccessStatuses(source sentrySourceInfo) []string {
	statuses := make(map[string]struct{})
	for status := range sentryObjectFieldOrEmpty(source.Operation, "responses") {
		if strings.HasPrefix(status, "2") {
			statuses[status] = struct{}{}
		}
	}
	return sentrySortedSet(statuses)
}

func sentrySourceContinuation(source sentrySourceInfo) sentryContinuationFact {
	for _, rawParameter := range sentryArrayFieldOrEmpty(source.Operation, "parameters") {
		parameter := sentryObjectValue(nil, rawParameter, "source parameter")
		if sentryStringField(nil, parameter, "in") != "query" {
			continue
		}
		description := sentryStringFieldOrEmpty(parameter, "description")
		if !sentryDescribesContinuation(description) {
			continue
		}
		return sentryContinuationFact{
			Kind:        "query_parameter",
			Parameter:   sentryStringField(nil, parameter, "name"),
			Description: description,
		}
	}
	return sentryContinuationFact{Kind: "not_documented"}
}

func sentryDescribesContinuation(description string) bool {
	text := strings.ToLower(description)
	if strings.Contains(text, "paginate") || strings.Contains(text, "pagination") {
		return true
	}
	return strings.Contains(text, "next") && (strings.Contains(text, "page") || strings.Contains(text, "result"))
}

func sentrySourceHasPageSizeOnlyControl(source sentrySourceInfo) bool {
	if sentrySourceContinuation(source).Kind != "not_documented" {
		return false
	}
	for _, rawParameter := range sentryArrayFieldOrEmpty(source.Operation, "parameters") {
		parameter := sentryObjectValue(nil, rawParameter, "source parameter")
		if sentryStringField(nil, parameter, "in") != "query" {
			continue
		}
		name, text := strings.ToLower(sentryStringFieldOrEmpty(parameter, "name")), strings.ToLower(sentryStringFieldOrEmpty(parameter, "description"))
		if name == "per_page" || strings.Contains(text, "per page") || strings.Contains(text, "page size") {
			return true
		}
	}
	return false
}

func sentrySourceEventTransport(source sentrySourceInfo) sentryEventTransportFact {
	action := sentrySourceAction(source)
	callbackCount := len(sentryObjectFieldOrEmpty(source.Operation, "callbacks"))
	if action.Kind != "mutation" || !sentryContainsString(sentrySemanticTokens(action.Headline), "register") {
		return sentryEventTransportFact{Kind: "not_documented", RequiredFields: []string{}, DocumentedCallbacks: callbackCount}
	}
	for _, rawMedia := range sentryObjectFieldOrEmpty(sentryObjectFieldOrEmpty(source.Operation, "requestBody"), "content") {
		media := sentryObjectValue(nil, rawMedia, "request media")
		schema := sentryObjectFieldOrEmpty(media, "schema")
		properties := sentryObjectFieldOrEmpty(schema, "properties")
		requiredFields := sentryStringArrayFieldOrEmpty(schema, "required")
		sort.Strings(requiredFields)
		callbackField, eventSelectorField := "", ""
		for name, rawProperty := range properties {
			property := sentryObjectValue(nil, rawProperty, "request property")
			description := strings.ToLower(sentryStringFieldOrEmpty(property, "description"))
			switch {
			case sentryStringFieldOrEmpty(property, "type") == "string" && (strings.Contains(description, "webhook") || strings.Contains(description, "callback")):
				callbackField = name
			case sentryStringFieldOrEmpty(property, "type") == "array" && strings.Contains(description, "event") && (strings.Contains(description, "subscribe") || strings.Contains(description, "select")):
				eventSelectorField = name
			}
		}
		if callbackField != "" && eventSelectorField != "" && sentryContains(requiredFields, callbackField) && sentryContains(requiredFields, eventSelectorField) {
			return sentryEventTransportFact{
				Kind:                "webhook_registration",
				CallbackField:       callbackField,
				EventSelectorField:  eventSelectorField,
				Headline:            action.Headline,
				RequiredFields:      sentrySortedSet(map[string]struct{}{callbackField: {}, eventSelectorField: {}}),
				DocumentedCallbacks: callbackCount,
			}
		}
	}
	return sentryEventTransportFact{Kind: "not_documented", RequiredFields: []string{}, DocumentedCallbacks: callbackCount}
}

func sentrySourceExtractability(source sentrySourceInfo) sentryExtractabilityFact {
	statuses := make(map[string]struct{})
	for status, rawResponse := range sentryObjectFieldOrEmpty(source.Operation, "responses") {
		response := sentryObjectValue(nil, rawResponse, "source response")
		content := sentryObjectFieldOrEmpty(response, "content")
		jsonMedia, exists := content["application/json"]
		if !exists {
			continue
		}
		schema := sentryObjectFieldOrEmpty(sentryObjectValue(nil, jsonMedia, "JSON media"), "schema")
		if sentryStringFieldOrEmpty(schema, "type") == "array" {
			statuses[status] = struct{}{}
		}
	}
	if len(statuses) == 0 {
		return sentryExtractabilityFact{Kind: "not_documented", ResponseStatuses: []string{}}
	}
	return sentryExtractabilityFact{Kind: "json_array_response", ResponseStatuses: sentrySortedSet(statuses)}
}

func sentrySourceMedia(source sentrySourceInfo) ([]string, []string) {
	request := make(map[string]struct{})
	if requestBody := sentryObjectFieldOrEmpty(source.Operation, "requestBody"); requestBody != nil {
		for mediaType := range sentryObjectFieldOrEmpty(requestBody, "content") {
			request[mediaType] = struct{}{}
		}
	}
	response := make(map[string]struct{})
	for _, rawResponse := range sentryObjectFieldOrEmpty(source.Operation, "responses") {
		responseObject := sentryObjectValue(nil, rawResponse, "source response")
		for mediaType := range sentryObjectFieldOrEmpty(responseObject, "content") {
			response[mediaType] = struct{}{}
		}
	}
	return sentrySortedSet(request), sentrySortedSet(response)
}

func sentrySourceHasRequestMedia(source sentrySourceInfo, want string) bool {
	request, _ := sentrySourceMedia(source)
	return sentryContains(request, want)
}

func sentrySourceHasBinaryResponse(source sentrySourceInfo) bool {
	_, response := sentrySourceMedia(source)
	for _, mediaType := range response {
		if mediaType == "application/octet-stream" || mediaType == "application/pdf" || strings.HasPrefix(mediaType, "image/") || strings.HasPrefix(mediaType, "audio/") || strings.HasPrefix(mediaType, "video/") {
			return true
		}
	}
	return false
}

func sentryRequiresETL(source sentrySourceInfo) bool {
	return sentrySourceAction(source).Kind == "read" && sentrySourceContinuation(source).Kind != "not_documented"
}

func sentryRequiresSyncTransport(source sentrySourceInfo) bool {
	return sentrySourceEventTransport(source).Kind == "webhook_registration"
}

func sentryMatrixCellByLane(t *testing.T, operation map[string]any, lane string) map[string]any {
	t.Helper()
	cell := sentryMatrixCellByLaneNoTest(operation, lane)
	if cell == nil {
		t.Fatalf("source operation %s has no %s cell", sentryStringField(nil, operation, "source_id"), lane)
	}
	return cell
}

func sentryMatrixCellByLaneNoTest(operation map[string]any, lane string) map[string]any {
	for _, rawCell := range sentryArrayField(nil, operation, "cells") {
		cell := sentryObjectValue(nil, rawCell, "matrix cell")
		if sentryStringField(nil, cell, "lane") == lane {
			return cell
		}
	}
	return nil
}

func sentryMatrixCellCounts(matrix map[string]any) map[string]map[string]int {
	counts := make(map[string]map[string]int, len(sentryLaneOrder))
	for _, lane := range sentryLaneOrder {
		counts[lane] = map[string]int{}
	}
	for _, rawOperation := range sentryArrayField(nil, matrix, "operations") {
		operation := sentryObjectValue(nil, rawOperation, "matrix operation")
		for _, rawCell := range sentryArrayField(nil, operation, "cells") {
			cell := sentryObjectValue(nil, rawCell, "matrix cell")
			lane, state := sentryStringField(nil, cell, "lane"), sentryStringField(nil, cell, "state")
			if _, exists := counts[lane]; !exists {
				counts[lane] = map[string]int{}
			}
			counts[lane][state]++
		}
	}
	return counts
}

func readSentryArtifactRecords(t *testing.T) sentryArtifactRecords {
	t.Helper()
	surface := readSentryJSONObject(t, "api_surface.json")
	streams := readSentryJSONObject(t, "streams.json")
	operations := readSentryJSONObject(t, "operations.json")
	cliSurface := readSentryJSONObject(t, "cli_surface.json")
	records := sentryArtifactRecords{
		API:             map[string]struct{}{},
		Streams:         map[string]struct{}{},
		Operations:      map[string]map[string]any{},
		Commands:        map[string]map[string]any{},
		SurfaceByStream: map[string]string{},
	}
	for _, rawEndpoint := range sentryArrayField(t, surface, "endpoints") {
		endpoint := sentryObjectValue(t, rawEndpoint, "api surface endpoint")
		record := sentryStringField(t, endpoint, "method") + " " + sentryStringField(t, endpoint, "path")
		records.API[record] = struct{}{}
		if coveredBy := sentryObjectFieldOrEmpty(endpoint, "covered_by"); coveredBy != nil {
			if stream := sentryStringFieldOrEmpty(coveredBy, "stream"); stream != "" {
				records.SurfaceByStream[stream] = record
			}
		}
	}
	for _, rawStream := range sentryArrayField(t, streams, "streams") {
		stream := sentryObjectValue(t, rawStream, "stream")
		records.Streams[sentryStringField(t, stream, "name")] = struct{}{}
	}
	for _, rawOperation := range sentryArrayField(t, operations, "operations") {
		operation := sentryObjectValue(t, rawOperation, "operation")
		records.Operations[sentryStringField(t, operation, "id")] = operation
	}
	for _, rawCommand := range sentryArrayField(t, cliSurface, "commands") {
		command := sentryObjectValue(t, rawCommand, "command")
		records.Commands[sentryStringField(t, command, "path")] = command
	}
	return records
}

func readSentryJSONObject(t *testing.T, path string) map[string]any {
	t.Helper()
	bytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var value map[string]any
	if err := json.Unmarshal(bytes, &value); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return value
}

func sentryStringReasonMap(object map[string]any, name string) map[string]string {
	result := make(map[string]string)
	for _, rawRecord := range sentryArrayField(nil, object, name) {
		record := sentryObjectValue(nil, rawRecord, "unlinked artifact record")
		result[sentryStringField(nil, record, "record")] = sentryStringField(nil, record, "reason")
	}
	return result
}

func sentryObjectField(t *testing.T, object map[string]any, name string) map[string]any {
	sentryMarkHelper(t)
	value, exists := object[name]
	if !exists {
		if t != nil {
			t.Fatalf("JSON object is missing %q", name)
		}
		panic(fmt.Sprintf("JSON object is missing %q", name))
	}
	return sentryObjectValue(t, value, name)
}

func sentryObjectFieldOrEmpty(object map[string]any, name string) map[string]any {
	value, exists := object[name]
	if !exists || value == nil {
		return map[string]any{}
	}
	result, ok := value.(map[string]any)
	if !ok {
		return map[string]any{}
	}
	return result
}

func sentryObjectValue(t *testing.T, value any, context string) map[string]any {
	sentryMarkHelper(t)
	result, ok := value.(map[string]any)
	if !ok {
		if t != nil {
			t.Fatalf("%s is %T, want JSON object", context, value)
		}
		panic(fmt.Sprintf("%s is %T, want JSON object", context, value))
	}
	return result
}

func sentryArrayField(t *testing.T, object map[string]any, name string) []any {
	sentryMarkHelper(t)
	value, exists := object[name]
	if !exists {
		if t != nil {
			t.Fatalf("JSON object is missing %q", name)
		}
		panic(fmt.Sprintf("JSON object is missing %q", name))
	}
	result, ok := value.([]any)
	if !ok {
		if t != nil {
			t.Fatalf("JSON field %q is %T, want array", name, value)
		}
		panic(fmt.Sprintf("JSON field %q is %T, want array", name, value))
	}
	return result
}

func sentryArrayFieldOrEmpty(object map[string]any, name string) []any {
	value, exists := object[name]
	if !exists || value == nil {
		return []any{}
	}
	result, ok := value.([]any)
	if !ok {
		return []any{}
	}
	return result
}

func sentryStringArrayField(t *testing.T, object map[string]any, name string) []string {
	sentryMarkHelper(t)
	values := sentryArrayField(t, object, name)
	result := make([]string, 0, len(values))
	for _, value := range values {
		text, ok := value.(string)
		if !ok {
			if t != nil {
				t.Fatalf("JSON field %q has non-string entry %T", name, value)
			}
			panic(fmt.Sprintf("JSON field %q has non-string entry %T", name, value))
		}
		result = append(result, text)
	}
	return result
}

func sentryStringArrayFieldOrEmpty(object map[string]any, name string) []string {
	values := sentryArrayFieldOrEmpty(object, name)
	result := make([]string, 0, len(values))
	for _, value := range values {
		text, ok := value.(string)
		if !ok {
			return []string{}
		}
		result = append(result, text)
	}
	return result
}

func sentryStringField(t *testing.T, object map[string]any, name string) string {
	sentryMarkHelper(t)
	value, exists := object[name]
	if !exists {
		if t != nil {
			t.Fatalf("JSON object is missing %q", name)
		}
		panic(fmt.Sprintf("JSON object is missing %q", name))
	}
	result, ok := value.(string)
	if !ok {
		if t != nil {
			t.Fatalf("JSON field %q is %T, want string", name, value)
		}
		panic(fmt.Sprintf("JSON field %q is %T, want string", name, value))
	}
	return result
}

func sentryStringFieldOrEmpty(object map[string]any, name string) string {
	value, exists := object[name]
	if !exists || value == nil {
		return ""
	}
	result, ok := value.(string)
	if !ok {
		return ""
	}
	return result
}

func sentryBoolFieldOrFalse(object map[string]any, name string) bool {
	value, exists := object[name]
	if !exists || value == nil {
		return false
	}
	result, ok := value.(bool)
	return ok && result
}

func sentryNumberField(object map[string]any, name string) int {
	value, exists := object[name]
	if !exists {
		panic(fmt.Sprintf("JSON object is missing %q", name))
	}
	number, ok := value.(float64)
	if !ok || number != float64(int(number)) {
		panic(fmt.Sprintf("JSON field %q is %T, want integral number", name, value))
	}
	return int(number)
}

func sentryNumberFieldOrZero(object map[string]any, name string) int {
	value, exists := object[name]
	if !exists || value == nil {
		return 0
	}
	number, ok := value.(float64)
	if !ok || number != float64(int(number)) {
		return 0
	}
	return int(number)
}

func sentrySortedSet(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func sentrySortedMapKeys(values map[string]sentrySourceInfo) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func sentryContains(values []string, want string) bool {
	index := sort.SearchStrings(values, want)
	return index < len(values) && values[index] == want
}

func sentryContainsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func sentryEqualStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range got {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}

func sentryEqualStringSet(got, want map[string]struct{}) bool {
	if len(got) != len(want) {
		return false
	}
	for value := range got {
		if _, exists := want[value]; !exists {
			return false
		}
	}
	return true
}

func sentryMarkHelper(t *testing.T) {
	if t != nil {
		t.Helper()
	}
}
