package stripe

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

const (
	stripeMatrixPath     = "sources/stripe-source-lane-matrix.json"
	stripeSourceLockPath = "sources/stripe-operation-source-lock.json"
)

var (
	stripeLaneOrder = []string{
		"direct_read",
		"direct_write",
		"binary_download",
		"binary_upload",
		"etl",
		"reverse_etl",
		"sync_transport",
	}
	stripePathParameter = regexp.MustCompile("[{]([^{}]+)[}]")
)

type stripeSourceInfo struct {
	ID             string
	Protocol       string
	Method         string
	Path           string
	OperationID    string
	SourceLocation string
	Operation      map[string]any
}

type stripePaginationFact struct {
	Kind            string
	QueryParameters []string
	ResponseFields  []string
}

type stripeWebhookRegistrationFact struct {
	Kind                  string
	RequiredRequestFields []string
}

type stripeArtifactRecords struct {
	API     map[string]struct{}
	Streams map[string]struct{}
	Writes  map[string]struct{}
}

func TestStripeSourceLaneMatrixContract(t *testing.T) {
	matrix := readStripeJSONObject(t, stripeMatrixPath)
	lock := readStripeJSONObject(t, stripeSourceLockPath)

	if err := validateStripeSourceLaneMatrix(matrix, lock, readStripeArtifactRecords(t)); err != nil {
		t.Fatal(err)
	}
}

func TestStripeSourceLaneMatrixRejectsHiddenSourceRow(t *testing.T) {
	matrix, lock := readStripeMatrixAndLock(t)
	operations := stripeArrayField(t, matrix, "operations")
	matrix["operations"] = operations[:len(operations)-1]

	assertStripeMatrixValidationError(t, matrix, lock, "source row absent from matrix")
}

func TestStripeSourceLaneMatrixRejectsDuplicateSourceID(t *testing.T) {
	matrix, lock := readStripeMatrixAndLock(t)
	operations := stripeArrayField(t, matrix, "operations")
	matrix["operations"] = append(operations, operations[0])

	assertStripeMatrixValidationError(t, matrix, lock, "duplicate matrix source id")
}

func TestStripeSourceLaneMatrixRejectsInvalidArtifactBacklink(t *testing.T) {
	matrix, lock := readStripeMatrixAndLock(t)
	artifacts := stripeArrayField(t, matrix, "artifacts")
	apiSurface := stripeObjectValue(t, artifacts[0], "matrix artifact")
	links := stripeArrayField(t, apiSurface, "links")
	link := stripeObjectValue(t, links[0], "api surface backlink")
	link["source_id"] = "stripe.rest.NoSuchOperation"

	assertStripeMatrixValidationError(t, matrix, lock, "artifact api_surface.json link")
}

func TestStripeSourceLaneMatrixRejectsMissingPagingDisposition(t *testing.T) {
	matrix, lock := readStripeMatrixAndLock(t)
	sourceByID := stripeSourceOperationsByID(t, lock)
	for _, rawOperation := range stripeArrayField(t, matrix, "operations") {
		operation := stripeObjectValue(t, rawOperation, "matrix operation")
		source := sourceByID[stripeStringField(t, operation, "source_id")]
		pagination := stripeSourcePagination(source)
		if pagination.Kind != "id_cursor" && pagination.Kind != "page_token" {
			continue
		}
		stripeMatrixCellByLane(t, operation, "etl")["state"] = "not_applicable"
		assertStripeMatrixValidationError(t, matrix, lock, "paging source operation")
		return
	}
	t.Fatal("matrix has no documented paging operation")
}

func TestStripeSourceLaneMatrixRejectsMissingMutationReverseETLDisposition(t *testing.T) {
	matrix, lock := readStripeMatrixAndLock(t)
	for _, rawOperation := range stripeArrayField(t, matrix, "operations") {
		operation := stripeObjectValue(t, rawOperation, "matrix operation")
		if !stripeIsMutation(stripeStringField(t, operation, "method")) {
			continue
		}
		stripeMatrixCellByLane(t, operation, "reverse_etl")["state"] = "not_applicable"
		assertStripeMatrixValidationError(t, matrix, lock, "mutation source operation")
		return
	}
	t.Fatal("matrix has no mutation operation")
}

func TestStripeSourceLaneMatrixRejectsSourceCountMismatch(t *testing.T) {
	matrix, lock := readStripeMatrixAndLock(t)
	stripeObjectField(t, matrix, "source_lock")["source_operation_count"] = float64(588)

	assertStripeMatrixValidationError(t, matrix, lock, "source lock operation count")
}

func TestStripeSourceLaneMatrixPreservesBinaryAndDeleteSurface(t *testing.T) {
	matrix, lock := readStripeMatrixAndLock(t)
	if err := validateStripeSourceLaneMatrix(matrix, lock, readStripeArtifactRecords(t)); err != nil {
		t.Fatal(err)
	}

	counts := stripeMatrixCellCounts(matrix)
	if got := counts["binary_download"]["mapped_unproven"]; got != 1 {
		t.Fatalf("binary_download mapped_unproven cells = %d, want 1", got)
	}
	if got := counts["binary_upload"]["mapped_unproven"]; got != 1 {
		t.Fatalf("binary_upload mapped_unproven cells = %d, want 1", got)
	}
	if got := counts["direct_write"]["mapped_unproven"]; got != 326 {
		t.Fatalf("direct_write mapped_unproven cells = %d, want 326", got)
	}
	if got := counts["reverse_etl"]["mapped_unproven"]; got != 326 {
		t.Fatalf("reverse_etl mapped_unproven cells = %d, want 326", got)
	}
	if got := counts["sync_transport"]["mapped_unproven"]; got != 0 {
		t.Fatalf("sync_transport mapped_unproven cells = %d, want 0", got)
	}
	if got := counts["sync_transport"]["missing_foundation"]; got != 1 {
		t.Fatalf("sync_transport missing_foundation cells = %d, want 1", got)
	}
}

func TestStripeSourceLaneMatrixSeparatesPageableETLFromWebhookSync(t *testing.T) {
	matrix, lock := readStripeMatrixAndLock(t)
	sourceByID := stripeSourceOperationsByID(t, lock)

	pageableRead := sourceByID["stripe.rest.GetCustomers"]
	if pagination := stripeSourcePagination(pageableRead); pagination.Kind != "id_cursor" {
		t.Fatalf("customers pagination = %#v, want id_cursor", pagination)
	}
	if registration := stripeSourceWebhookRegistration(pageableRead); registration.Kind != "not_documented" {
		t.Fatalf("customers webhook registration = %#v, want not_documented", registration)
	}

	webhookRegistration := sourceByID["stripe.rest.PostWebhookEndpoints"]
	if registration := stripeSourceWebhookRegistration(webhookRegistration); registration.Kind != "webhook_registration" ||
		!stripeEqualStrings(registration.RequiredRequestFields, []string{"enabled_events", "url"}) {
		t.Fatalf("webhook registration facts = %#v, want webhook_registration with enabled_events/url", registration)
	}

	byID := stripeMatrixOperationsByID(t, matrix)
	if got := stripeStringField(t, stripeMatrixCellByLane(t, byID[pageableRead.ID], "etl"), "state"); got != "mapped_unproven" {
		t.Fatalf("pageable read ETL state = %q, want mapped_unproven", got)
	}
	if got := stripeStringField(t, stripeMatrixCellByLane(t, byID[pageableRead.ID], "sync_transport"), "state"); got != "not_applicable" {
		t.Fatalf("pageable read sync state = %q, want not_applicable", got)
	}
	if got := stripeStringField(t, stripeMatrixCellByLane(t, byID[webhookRegistration.ID], "sync_transport"), "state"); got != "missing_foundation" {
		t.Fatalf("webhook registration sync state = %q, want missing_foundation", got)
	}

	stripeMatrixCellByLane(t, byID[pageableRead.ID], "sync_transport")["state"] = "mapped_unproven"
	stripeMatrixCellByLane(t, byID[pageableRead.ID], "sync_transport")["reason"] = "stripe.source.sync_transport.documented_pagination.v1"
	assertStripeMatrixValidationError(t, matrix, lock, "source operation "+pageableRead.ID)
}

func TestStripeWebhookRegistrationRequiresProviderSemanticContract(t *testing.T) {
	tests := []struct {
		name       string
		source     stripeSourceInfo
		want       string
		wantFields []string
	}{
		{
			name: "documented webhook registration",
			source: stripeSourceInfo{
				Method:    "POST",
				Operation: stripeWebhookRequestOperation("Create a webhook endpoint", "A webhook endpoint must have a url and a list of enabled_events.", "url", "enabled_events"),
			},
			want:       "webhook_registration",
			wantFields: []string{"enabled_events", "url"},
		},
		{
			name: "pageable read is not a webhook registration",
			source: stripeSourceInfo{
				Method:    "GET",
				Operation: stripeWebhookRequestOperation("List all customers", "Returns a list of customers.", "url", "enabled_events"),
			},
			want: "not_documented",
		},
		{
			name: "URL without event selector is not a webhook registration",
			source: stripeSourceInfo{
				Method:    "POST",
				Operation: stripeWebhookRequestOperation("Create a webhook endpoint", "A webhook endpoint must have a url.", "url"),
			},
			want: "not_documented",
		},
		{
			name: "event fields without webhook semantics are not a registration",
			source: stripeSourceInfo{
				Method:    "POST",
				Operation: stripeWebhookRequestOperation("Create an outbound request", "Create a request with a URL and selected events.", "url", "enabled_events"),
			},
			want: "not_documented",
		},
		{
			name: "fields split across media schemas are not one request contract",
			source: stripeSourceInfo{
				Method: "POST",
				Operation: map[string]any{
					"summary":     "Create a webhook endpoint",
					"description": "A webhook endpoint can receive events.",
					"requestBody": map[string]any{
						"content": map[string]any{
							"application/json":                  map[string]any{"schema": map[string]any{"required": []any{"url"}}},
							"application/x-www-form-urlencoded": map[string]any{"schema": map[string]any{"required": []any{"enabled_events"}}},
						},
					},
				},
			},
			want: "not_documented",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := stripeSourceWebhookRegistration(test.source)
			if got.Kind != test.want || !stripeEqualStrings(got.RequiredRequestFields, test.wantFields) {
				t.Fatalf("webhook registration = %#v, want kind=%q fields=%v", got, test.want, test.wantFields)
			}
		})
	}
}

func TestStripeSourcePaginationRequiresProviderContinuation(t *testing.T) {
	tests := []struct {
		name   string
		source stripeSourceInfo
		want   string
	}{
		{
			name: "cursor response plus starting-after is extractable",
			source: stripeSourceInfo{
				Method:    "GET",
				Operation: stripeListResponseOperation([]string{"starting_after"}, false),
			},
			want: "id_cursor",
		},
		{
			name: "page token response plus page is extractable",
			source: stripeSourceInfo{
				Method:    "GET",
				Operation: stripeListResponseOperation([]string{"page"}, true),
			},
			want: "page_token",
		},
		{
			name: "list shape without request continuation is not ETL",
			source: stripeSourceInfo{
				Method:    "GET",
				Operation: stripeListResponseOperation(nil, false),
			},
			want: "response_list_without_operation_continuation",
		},
		{
			name: "POST list shape is not an ETL read",
			source: stripeSourceInfo{
				Method:    "POST",
				Operation: stripeListResponseOperation([]string{"starting_after"}, false),
			},
			want: "not_documented",
		},
		{
			name: "string has-more marker is not a continuation contract",
			source: stripeSourceInfo{
				Method: "GET",
				Operation: map[string]any{
					"parameters": []any{map[string]any{"in": "query", "name": "starting_after"}},
					"responses": map[string]any{
						"200": map[string]any{
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"properties": map[string]any{
											"data":     map[string]any{"type": "array"},
											"has_more": map[string]any{"type": "string"},
										},
									},
								},
							},
						},
					},
				},
			},
			want: "not_documented",
		},
		{
			name: "malformed provider marker is not a continuation contract",
			source: stripeSourceInfo{
				Method: "GET",
				Operation: map[string]any{
					"parameters": []any{map[string]any{"in": "query", "name": "starting_after"}},
					"responses": map[string]any{
						"200": map[string]any{
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"properties": map[string]any{
											"data":     map[string]any{"type": "array"},
											"has_more": "unknown",
										},
									},
								},
							},
						},
					},
				},
			},
			want: "not_documented",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := stripeSourcePagination(test.source); got.Kind != test.want {
				t.Fatalf("pagination = %#v, want kind=%q", got, test.want)
			}
		})
	}
}

func stripeWebhookRequestOperation(summary, description string, required ...string) map[string]any {
	fields := make([]any, len(required))
	for index, field := range required {
		fields[index] = field
	}
	return map[string]any{
		"summary":     summary,
		"description": description,
		"requestBody": map[string]any{
			"content": map[string]any{
				"application/x-www-form-urlencoded": map[string]any{
					"schema": map[string]any{
						"required": fields,
					},
				},
			},
		},
	}
}

func stripeListResponseOperation(queryParameters []string, nextPage bool) map[string]any {
	parameters := make([]any, len(queryParameters))
	for index, parameter := range queryParameters {
		parameters[index] = map[string]any{"in": "query", "name": parameter}
	}
	properties := map[string]any{
		"data":     map[string]any{"type": "array"},
		"has_more": map[string]any{"type": "boolean"},
	}
	if nextPage {
		properties["next_page"] = map[string]any{"type": "string"}
	}
	return map[string]any{
		"parameters": parameters,
		"responses": map[string]any{
			"200": map[string]any{
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{
							"properties": properties,
						},
					},
				},
			},
		},
	}
}

func TestStripeSourceLaneMatrixDirectAndBinaryCellsFollowLockedFacts(t *testing.T) {
	matrix, lock := readStripeMatrixAndLock(t)
	byID := stripeMatrixOperationsByID(t, matrix)
	for _, test := range []struct {
		name   string
		source string
		lanes  map[string]string
	}{
		{
			name:   "bounded GET read",
			source: "stripe.rest.GetBalance",
			lanes:  map[string]string{"direct_read": "mapped_unproven", "direct_write": "not_applicable", "reverse_etl": "not_applicable"},
		},
		{
			name:   "provider mutation is direct-write and reverse-ETL",
			source: "stripe.rest.DeleteCustomersCustomer",
			lanes:  map[string]string{"direct_read": "not_applicable", "direct_write": "mapped_unproven", "reverse_etl": "mapped_unproven"},
		},
		{
			name:   "PDF response is binary download",
			source: "stripe.rest.GetQuotesQuotePdf",
			lanes:  map[string]string{"binary_download": "mapped_unproven", "binary_upload": "not_applicable"},
		},
		{
			name:   "multipart request is binary upload",
			source: "stripe.rest.PostFiles",
			lanes:  map[string]string{"binary_download": "not_applicable", "binary_upload": "mapped_unproven", "direct_write": "mapped_unproven", "reverse_etl": "mapped_unproven"},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			operation := byID[test.source]
			if operation == nil {
				t.Fatalf("matrix operation %q is absent", test.source)
			}
			for lane, want := range test.lanes {
				if got := stripeStringField(t, stripeMatrixCellByLane(t, operation, lane), "state"); got != want {
					t.Errorf("%s state = %q, want %q", lane, got, want)
				}
			}
		})
	}

	if err := validateStripeSourceLaneMatrix(matrix, lock, readStripeArtifactRecords(t)); err != nil {
		t.Fatal(err)
	}
}

func readStripeMatrixAndLock(t *testing.T) (map[string]any, map[string]any) {
	t.Helper()
	return readStripeJSONObject(t, stripeMatrixPath), readStripeJSONObject(t, stripeSourceLockPath)
}

func assertStripeMatrixValidationError(t *testing.T, matrix, lock map[string]any, want string) {
	t.Helper()
	err := validateStripeSourceLaneMatrix(matrix, lock, readStripeArtifactRecords(t))
	if err == nil {
		t.Fatalf("matrix validation unexpectedly passed, want error containing %q", want)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("matrix validation error = %q, want substring %q", err, want)
	}
}

func validateStripeSourceLaneMatrix(matrix, lock map[string]any, artifacts stripeArtifactRecords) error {
	if stripeNumberField(matrix, "schema_version") != 1 {
		return fmt.Errorf("matrix schema_version = %d, want 1", stripeNumberField(matrix, "schema_version"))
	}
	if stripeStringField(nil, matrix, "connector") != "stripe" {
		return fmt.Errorf("matrix connector = %q, want stripe", stripeStringField(nil, matrix, "connector"))
	}
	if !stripeEqualStrings(stripeStringArrayField(nil, matrix, "lanes"), stripeLaneOrder) {
		return fmt.Errorf("matrix lanes = %v, want %v", stripeStringArrayField(nil, matrix, "lanes"), stripeLaneOrder)
	}

	sourceByID, err := stripeSourceOperationsByIDNoTest(lock)
	if err != nil {
		return err
	}
	if err := stripeValidateSourceLockReference(stripeObjectField(nil, matrix, "source_lock"), lock, len(sourceByID)); err != nil {
		return err
	}

	matrixByID := make(map[string]map[string]any, len(sourceByID))
	for _, rawOperation := range stripeArrayField(nil, matrix, "operations") {
		operation := stripeObjectValue(nil, rawOperation, "matrix operation")
		sourceID := stripeStringField(nil, operation, "source_id")
		if _, exists := matrixByID[sourceID]; exists {
			return fmt.Errorf("duplicate matrix source id %q", sourceID)
		}
		matrixByID[sourceID] = operation
	}
	for sourceID, source := range sourceByID {
		operation, exists := matrixByID[sourceID]
		if !exists {
			return fmt.Errorf("source row absent from matrix: %s", sourceID)
		}
		if err := stripeValidateOperation(operation, source); err != nil {
			return err
		}
	}
	for sourceID := range matrixByID {
		if _, exists := sourceByID[sourceID]; !exists {
			return fmt.Errorf("matrix source id %q is not in the source lock", sourceID)
		}
	}
	if len(matrixByID) != len(sourceByID) {
		return fmt.Errorf("matrix operation count = %d, source lock count = %d", len(matrixByID), len(sourceByID))
	}
	if err := stripeValidateCountReconciliation(matrix, lock, sourceByID); err != nil {
		return err
	}
	return stripeValidateArtifactLinks(matrix, artifacts, matrixByID)
}

func stripeValidateSourceLockReference(reference, lock map[string]any, sourceCount int) error {
	rest := stripeObjectField(nil, lock, "rest")
	if stripeStringField(nil, reference, "path") != stripeSourceLockPath {
		return fmt.Errorf("matrix source lock path = %q, want %q", stripeStringField(nil, reference, "path"), stripeSourceLockPath)
	}
	if stripeStringField(nil, reference, "source_url") != stripeStringField(nil, rest, "source_url") ||
		stripeStringField(nil, reference, "sha256") != stripeStringField(nil, rest, "sha256") ||
		stripeStringField(nil, reference, "captured_at") != stripeStringField(nil, lock, "captured_at") {
		return errors.New("matrix source lock identity does not match the pinned Stripe source")
	}
	counts := stripeObjectField(nil, lock, "counts")
	if count := stripeNumberField(reference, "source_operation_count"); count != sourceCount ||
		count != stripeNumberField(counts, "rest") || count != stripeNumberField(counts, "total") {
		return fmt.Errorf("source lock operation count = %d, want %d", count, sourceCount)
	}
	return nil
}

func stripeValidateOperation(operation map[string]any, source stripeSourceInfo) error {
	if stripeStringField(nil, operation, "protocol") != source.Protocol ||
		stripeStringField(nil, operation, "method") != source.Method ||
		stripeStringField(nil, operation, "path") != source.Path ||
		stripeStringField(nil, operation, "operation_id") != source.OperationID {
		return fmt.Errorf("source operation %s does not preserve the locked operation identity", source.ID)
	}
	if location := stripeStringField(nil, operation, "source_location"); location == "" || location != source.SourceLocation {
		return fmt.Errorf("source operation %s does not preserve its cited source location", source.ID)
	}
	if err := stripeValidateFacts(stripeObjectField(nil, operation, "facts"), source); err != nil {
		return err
	}

	expected := stripeExpectedCells(source)
	cells := stripeArrayField(nil, operation, "cells")
	if len(cells) != len(stripeLaneOrder) {
		return fmt.Errorf("source operation %s has %d cells, want %d", source.ID, len(cells), len(stripeLaneOrder))
	}
	seen := make(map[string]struct{}, len(cells))
	for _, rawCell := range cells {
		cell := stripeObjectValue(nil, rawCell, "matrix cell")
		lane := stripeStringField(nil, cell, "lane")
		if _, duplicate := seen[lane]; duplicate {
			return fmt.Errorf("source operation %s has duplicate %s cell", source.ID, lane)
		}
		seen[lane] = struct{}{}
		want, exists := expected[lane]
		if !exists {
			return fmt.Errorf("source operation %s has unknown lane %q", source.ID, lane)
		}
		if state, reason := stripeStringField(nil, cell, "state"), stripeStringField(nil, cell, "reason"); state != want[0] || reason != want[1] {
			pagination := stripeSourcePagination(source)
			if pagination.Kind == "id_cursor" || pagination.Kind == "page_token" {
				return fmt.Errorf("paging source operation %s has %s cell = %s/%s, want %s/%s", source.ID, lane, state, reason, want[0], want[1])
			}
			if stripeIsMutation(source.Method) {
				return fmt.Errorf("mutation source operation %s has %s cell = %s/%s, want %s/%s", source.ID, lane, state, reason, want[0], want[1])
			}
			return fmt.Errorf("source operation %s has %s cell = %s/%s, want %s/%s", source.ID, lane, state, reason, want[0], want[1])
		}
		if stripeStringField(nil, cell, "source_location") != source.SourceLocation {
			return fmt.Errorf("source operation %s has uncited %s cell", source.ID, lane)
		}
		state := stripeStringField(nil, cell, "state")
		if state != "mapped_unproven" && state != "missing_foundation" && state != "not_applicable" {
			return fmt.Errorf("source operation %s has unsupported mapping-only state %q", source.ID, state)
		}
		if state == "missing_foundation" && lane != "sync_transport" {
			return fmt.Errorf("source operation %s has non-sync missing-foundation cell %q", source.ID, lane)
		}
	}
	for _, lane := range stripeLaneOrder {
		if _, exists := seen[lane]; !exists {
			return fmt.Errorf("source operation %s has no %s cell", source.ID, lane)
		}
	}
	return nil
}

func stripeValidateFacts(facts map[string]any, source stripeSourceInfo) error {
	classification := "mutation"
	if source.Method == "GET" {
		classification = "read"
	} else if !stripeIsMutation(source.Method) {
		return fmt.Errorf("source operation %s uses unsupported retained method %q", source.ID, source.Method)
	}
	if stripeStringField(nil, facts, "classification") != classification {
		return fmt.Errorf("source operation %s classification = %q, want %q", source.ID, stripeStringField(nil, facts, "classification"), classification)
	}

	pagination := stripeSourcePagination(source)
	gotPagination := stripeObjectField(nil, facts, "pagination")
	if stripeStringField(nil, gotPagination, "kind") != pagination.Kind ||
		!stripeEqualStrings(stripeStringArrayField(nil, gotPagination, "query_parameters"), pagination.QueryParameters) ||
		!stripeEqualStrings(stripeStringArrayField(nil, gotPagination, "response_fields"), pagination.ResponseFields) ||
		stripeStringField(nil, gotPagination, "source_location") != source.SourceLocation {
		return fmt.Errorf("source operation %s does not preserve exact pagination evidence", source.ID)
	}

	pathParameters, queryParameters := stripeSourceScope(source)
	scope := stripeObjectField(nil, facts, "scope")
	if !stripeEqualStrings(stripeStringArrayField(nil, scope, "path_parameters"), pathParameters) ||
		!stripeEqualStrings(stripeStringArrayField(nil, scope, "query_parameters"), queryParameters) ||
		stripeStringField(nil, scope, "source_location") != source.SourceLocation {
		return fmt.Errorf("source operation %s does not preserve exact scope/path evidence", source.ID)
	}

	requestMedia, responseMedia := stripeSourceMedia(source)
	media := stripeObjectField(nil, facts, "media")
	if !stripeEqualStrings(stripeStringArrayField(nil, media, "request"), requestMedia) ||
		!stripeEqualStrings(stripeStringArrayField(nil, media, "response"), responseMedia) ||
		stripeStringField(nil, media, "source_location") != source.SourceLocation {
		return fmt.Errorf("source operation %s does not preserve exact media evidence", source.ID)
	}

	eventCursor := stripeObjectField(nil, facts, "event_cursor")
	callbacks := stripeObjectFieldOrEmpty(source.Operation, "callbacks")
	registration := stripeSourceWebhookRegistration(source)
	if len(callbacks) != 0 || stripeStringField(nil, eventCursor, "kind") != registration.Kind ||
		!stripeEqualStrings(stripeStringArrayFieldOrEmpty(eventCursor, "required_request_fields"), registration.RequiredRequestFields) ||
		stripeNumberField(eventCursor, "documented_callbacks") != 0 ||
		stripeStringField(nil, eventCursor, "source_location") != source.SourceLocation {
		return fmt.Errorf("source operation %s does not preserve its event/cursor evidence", source.ID)
	}
	return nil
}

func stripeExpectedCells(source stripeSourceInfo) map[string][2]string {
	pagination := stripeSourcePagination(source)
	isPaging := pagination.Kind == "id_cursor" || pagination.Kind == "page_token"
	webhookRegistration := stripeSourceWebhookRegistration(source)
	isPDF := stripeSourceHasResponseMedia(source, "application/pdf")
	isMultipart := stripeSourceHasRequestMedia(source, "multipart/form-data")
	isMutation := stripeIsMutation(source.Method)

	expected := map[string][2]string{}
	if source.Method == "GET" {
		expected["direct_read"] = [2]string{"mapped_unproven", "stripe.source.direct_read.documented_get_response.v1"}
		expected["direct_write"] = [2]string{"not_applicable", "stripe.source.direct_write.get_not_applicable.v1"}
		expected["reverse_etl"] = [2]string{"not_applicable", "stripe.source.reverse_etl.get_not_applicable.v1"}
	} else if isMutation {
		expected["direct_read"] = [2]string{"not_applicable", "stripe.source.direct_read.mutation_verb_not_applicable.v1"}
		expected["direct_write"] = [2]string{"mapped_unproven", "stripe.source.direct_write.mutation_verb.v1"}
		expected["reverse_etl"] = [2]string{"mapped_unproven", "stripe.source.reverse_etl.mutation_verb.v1"}
	}
	if isPDF {
		expected["binary_download"] = [2]string{"mapped_unproven", "stripe.source.binary_download.pdf_response.v1"}
	} else {
		expected["binary_download"] = [2]string{"not_applicable", "stripe.source.binary_download.no_pdf_response.v1"}
	}
	if isMultipart {
		expected["binary_upload"] = [2]string{"mapped_unproven", "stripe.source.binary_upload.multipart_request.v1"}
	} else {
		expected["binary_upload"] = [2]string{"not_applicable", "stripe.source.binary_upload.no_multipart_request.v1"}
	}
	if isPaging {
		expected["etl"] = [2]string{"mapped_unproven", "stripe.source.etl.documented_pagination.v1"}
	} else if pagination.Kind == "response_list_without_operation_continuation" {
		expected["etl"] = [2]string{"not_applicable", "stripe.source.etl.response_list_without_operation_continuation.v1"}
	} else {
		expected["etl"] = [2]string{"not_applicable", "stripe.source.etl.no_documented_pagination.v1"}
	}
	if webhookRegistration.Kind == "webhook_registration" {
		expected["sync_transport"] = [2]string{"missing_foundation", "stripe.foundation.inbound_webhook_receiver.v1"}
	} else {
		expected["sync_transport"] = [2]string{"not_applicable", "stripe.source.sync_transport.no_documented_webhook_registration.v1"}
	}
	return expected
}

func stripeValidateCountReconciliation(matrix, lock map[string]any, sourceByID map[string]stripeSourceInfo) error {
	if len(stripeArrayField(nil, matrix, "operations")) != 589 || len(sourceByID) != 589 {
		return fmt.Errorf("retained source count reconciliation = matrix:%d lock:%d, want 589", len(stripeArrayField(nil, matrix, "operations")), len(sourceByID))
	}

	var getCount, mutationCount, deleteCount, pagingCount, cursorCount, pageTokenCount, listWithoutContinuationCount, pdfCount, multipartCount, webhookRegistrationCount int
	listWithoutContinuation := make(map[string]struct{})
	pdfIDs := make(map[string]struct{})
	multipartIDs := make(map[string]struct{})
	webhookRegistrationIDs := make(map[string]struct{})
	for _, source := range sourceByID {
		if source.Method == "GET" {
			getCount++
		}
		if stripeIsMutation(source.Method) {
			mutationCount++
		}
		if source.Method == "DELETE" {
			deleteCount++
		}
		switch stripeSourcePagination(source).Kind {
		case "id_cursor":
			pagingCount++
			cursorCount++
		case "page_token":
			pagingCount++
			pageTokenCount++
		case "response_list_without_operation_continuation":
			listWithoutContinuationCount++
			listWithoutContinuation[source.ID] = struct{}{}
		}
		if stripeSourceHasResponseMedia(source, "application/pdf") {
			pdfCount++
			pdfIDs[source.ID] = struct{}{}
		}
		if stripeSourceHasRequestMedia(source, "multipart/form-data") {
			multipartCount++
			multipartIDs[source.ID] = struct{}{}
		}
		if stripeSourceWebhookRegistration(source).Kind == "webhook_registration" {
			webhookRegistrationCount++
			webhookRegistrationIDs[source.ID] = struct{}{}
		}
	}
	if getCount != 263 || mutationCount != 326 || deleteCount != 32 || pagingCount != 128 || cursorCount != 121 || pageTokenCount != 7 || listWithoutContinuationCount != 2 || pdfCount != 1 || multipartCount != 1 || webhookRegistrationCount != 1 {
		return fmt.Errorf("source fact counts = GET:%d mutation:%d delete:%d paging:%d cursor:%d page_token:%d list_without_continuation:%d pdf:%d multipart:%d webhook_registration:%d", getCount, mutationCount, deleteCount, pagingCount, cursorCount, pageTokenCount, listWithoutContinuationCount, pdfCount, multipartCount, webhookRegistrationCount)
	}
	if !stripeEqualStringSet(listWithoutContinuation, map[string]struct{}{
		"stripe.rest.GetAccountsAccountCapabilities": {},
		"stripe.rest.GetReportingReportTypes":        {},
	}) || !stripeEqualStringSet(pdfIDs, map[string]struct{}{"stripe.rest.GetQuotesQuotePdf": {}}) ||
		!stripeEqualStringSet(multipartIDs, map[string]struct{}{"stripe.rest.PostFiles": {}}) ||
		!stripeEqualStringSet(webhookRegistrationIDs, map[string]struct{}{"stripe.rest.PostWebhookEndpoints": {}}) {
		return errors.New("Stripe source facts no longer match the cited paging, binary, or webhook registration operation set")
	}

	cells := stripeMatrixCellCounts(matrix)
	wantMapped := map[string]int{
		"direct_read":     263,
		"direct_write":    326,
		"binary_download": 1,
		"binary_upload":   1,
		"etl":             128,
		"reverse_etl":     326,
		"sync_transport":  0,
	}
	wantMissingFoundation := map[string]int{
		"direct_read":     0,
		"direct_write":    0,
		"binary_download": 0,
		"binary_upload":   0,
		"etl":             0,
		"reverse_etl":     0,
		"sync_transport":  1,
	}
	mappedTotal, missingFoundationTotal, notApplicableTotal := 0, 0, 0
	for _, lane := range stripeLaneOrder {
		if cells[lane]["mapped_unproven"] != wantMapped[lane] {
			return fmt.Errorf("%s mapped_unproven count = %d, want %d", lane, cells[lane]["mapped_unproven"], wantMapped[lane])
		}
		if cells[lane]["missing_foundation"] != wantMissingFoundation[lane] {
			return fmt.Errorf("%s missing_foundation count = %d, want %d", lane, cells[lane]["missing_foundation"], wantMissingFoundation[lane])
		}
		if cells[lane]["not_applicable"] != 589-wantMapped[lane]-wantMissingFoundation[lane] {
			return fmt.Errorf("%s not_applicable count = %d, want %d", lane, cells[lane]["not_applicable"], 589-wantMapped[lane]-wantMissingFoundation[lane])
		}
		if cells[lane]["implemented"] != 0 {
			return fmt.Errorf("%s contains an execution state", lane)
		}
		mappedTotal += cells[lane]["mapped_unproven"]
		missingFoundationTotal += cells[lane]["missing_foundation"]
		notApplicableTotal += cells[lane]["not_applicable"]
	}
	if mappedTotal != 1045 || missingFoundationTotal != 1 || notApplicableTotal != 3077 || mappedTotal+missingFoundationTotal+notApplicableTotal != 4123 {
		return fmt.Errorf("matrix cell totals = mapped:%d missing_foundation:%d not_applicable:%d total:%d, want 1045/1/3077/4123", mappedTotal, missingFoundationTotal, notApplicableTotal, mappedTotal+missingFoundationTotal+notApplicableTotal)
	}
	return nil
}

func stripeValidateArtifactLinks(matrix map[string]any, records stripeArtifactRecords, matrixByID map[string]map[string]any) error {
	byPath := make(map[string]map[string]any)
	for _, rawArtifact := range stripeArrayField(nil, matrix, "artifacts") {
		artifact := stripeObjectValue(nil, rawArtifact, "matrix artifact")
		path := stripeStringField(nil, artifact, "path")
		if _, exists := byPath[path]; exists {
			return fmt.Errorf("duplicate artifact matrix %q", path)
		}
		byPath[path] = artifact
	}
	if len(byPath) != 3 {
		return fmt.Errorf("matrix artifact count = %d, want 3", len(byPath))
	}
	if err := stripeValidateAPISurfaceLinks(byPath["api_surface.json"], records.API, matrixByID); err != nil {
		return err
	}
	if err := stripeValidateNamedArtifactLinks("streams.json", byPath["streams.json"], records.Streams, matrixByID, map[string]string{
		"customers":     "stripe.rest.GetCustomers",
		"charges":       "stripe.rest.GetCharges",
		"invoices":      "stripe.rest.GetInvoices",
		"subscriptions": "stripe.rest.GetSubscriptions",
		"products":      "stripe.rest.GetProducts",
	}, []string{"direct_read", "etl"}); err != nil {
		return err
	}
	return stripeValidateNamedArtifactLinks("writes.json", byPath["writes.json"], records.Writes, matrixByID, map[string]string{
		"create_customer": "stripe.rest.PostCustomers",
		"update_customer": "stripe.rest.PostCustomersCustomer",
		"delete_customer": "stripe.rest.DeleteCustomersCustomer",
	}, []string{"direct_write", "reverse_etl"})
}

func stripeValidateAPISurfaceLinks(artifact map[string]any, records map[string]struct{}, matrixByID map[string]map[string]any) error {
	if artifact == nil {
		return errors.New("matrix is missing api_surface.json artifact")
	}
	links := stripeArrayField(nil, artifact, "links")
	if stripeNumberField(artifact, "record_count") != len(records) || len(links) != len(records) {
		return fmt.Errorf("artifact api_surface.json link count = records:%d links:%d, want %d", stripeNumberField(artifact, "record_count"), len(links), len(records))
	}
	seen := make(map[string]struct{}, len(links))
	for _, rawLink := range links {
		link := stripeObjectValue(nil, rawLink, "api surface backlink")
		record := stripeStringField(nil, link, "record")
		if _, exists := records[record]; !exists {
			return fmt.Errorf("artifact api_surface.json link references nonexistent record %q", record)
		}
		if _, duplicate := seen[record]; duplicate {
			return fmt.Errorf("artifact api_surface.json link duplicates record %q", record)
		}
		seen[record] = struct{}{}
		sourceID := stripeStringField(nil, link, "source_id")
		operation, exists := matrixByID[sourceID]
		if !exists {
			return fmt.Errorf("artifact api_surface.json link %q references nonexistent source cell owner %q", record, sourceID)
		}
		if record != stripeStringField(nil, operation, "method")+" "+stripeStringField(nil, operation, "path") {
			return fmt.Errorf("artifact api_surface.json link %q does not preserve source route", record)
		}
		lane := "direct_write"
		if stripeStringField(nil, operation, "method") == "GET" {
			lane = "direct_read"
		}
		if !stripeEqualStrings(stripeStringArrayField(nil, link, "lanes"), []string{lane}) || stripeMatrixCellByLaneNoTest(operation, lane) == nil {
			return fmt.Errorf("artifact api_surface.json link %q references nonexistent cell %q", record, lane)
		}
	}
	return nil
}

func stripeValidateNamedArtifactLinks(path string, artifact map[string]any, records map[string]struct{}, matrixByID map[string]map[string]any, expectedSourceID map[string]string, expectedLanes []string) error {
	if artifact == nil {
		return fmt.Errorf("matrix is missing %s artifact", path)
	}
	links := stripeArrayField(nil, artifact, "links")
	if stripeNumberField(artifact, "record_count") != len(records) || len(links) != len(records) || len(records) != len(expectedSourceID) {
		return fmt.Errorf("artifact %s link count does not reconcile", path)
	}
	seen := make(map[string]struct{}, len(links))
	for _, rawLink := range links {
		link := stripeObjectValue(nil, rawLink, "artifact backlink")
		record := stripeStringField(nil, link, "record")
		if _, exists := records[record]; !exists {
			return fmt.Errorf("artifact %s link references nonexistent record %q", path, record)
		}
		if _, duplicate := seen[record]; duplicate {
			return fmt.Errorf("artifact %s link duplicates record %q", path, record)
		}
		seen[record] = struct{}{}
		sourceID := stripeStringField(nil, link, "source_id")
		if sourceID != expectedSourceID[record] {
			return fmt.Errorf("artifact %s link %q source id = %q, want %q", path, record, sourceID, expectedSourceID[record])
		}
		operation, exists := matrixByID[sourceID]
		if !exists {
			return fmt.Errorf("artifact %s link %q references nonexistent source cell owner %q", path, record, sourceID)
		}
		if !stripeEqualStrings(stripeStringArrayField(nil, link, "lanes"), expectedLanes) {
			return fmt.Errorf("artifact %s link %q lanes = %v, want %v", path, record, stripeStringArrayField(nil, link, "lanes"), expectedLanes)
		}
		for _, lane := range expectedLanes {
			if stripeMatrixCellByLaneNoTest(operation, lane) == nil {
				return fmt.Errorf("artifact %s link %q references nonexistent cell %q", path, record, lane)
			}
		}
	}
	return nil
}

func stripeSourceOperationsByID(t *testing.T, lock map[string]any) map[string]stripeSourceInfo {
	t.Helper()
	operations, err := stripeSourceOperationsByIDNoTest(lock)
	if err != nil {
		t.Fatal(err)
	}
	return operations
}

func stripeMatrixOperationsByID(t *testing.T, matrix map[string]any) map[string]map[string]any {
	t.Helper()
	operations := make(map[string]map[string]any, len(stripeArrayField(t, matrix, "operations")))
	for _, rawOperation := range stripeArrayField(t, matrix, "operations") {
		operation := stripeObjectValue(t, rawOperation, "matrix operation")
		sourceID := stripeStringField(t, operation, "source_id")
		if _, duplicate := operations[sourceID]; duplicate {
			t.Fatalf("duplicate matrix source id %q", sourceID)
		}
		operations[sourceID] = operation
	}
	return operations
}

func stripeSourceOperationsByIDNoTest(lock map[string]any) (map[string]stripeSourceInfo, error) {
	rest := stripeObjectField(nil, lock, "rest")
	result := make(map[string]stripeSourceInfo)
	for _, rawOperation := range stripeArrayField(nil, rest, "operations") {
		record := stripeObjectValue(nil, rawOperation, "source lock operation")
		operation := stripeObjectField(nil, record, "source_operation")
		source := stripeSourceInfo{
			ID:             stripeStringField(nil, record, "id"),
			Protocol:       stripeStringField(nil, record, "protocol"),
			Method:         stripeStringField(nil, record, "method"),
			Path:           stripeStringField(nil, record, "path"),
			OperationID:    stripeStringField(nil, record, "operation_id"),
			SourceLocation: stripeStringField(nil, record, "source_location"),
			Operation:      operation,
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

func stripeSourceScope(source stripeSourceInfo) ([]string, []string) {
	path := make(map[string]struct{})
	for _, match := range stripePathParameter.FindAllStringSubmatch(source.Path, -1) {
		path[match[1]] = struct{}{}
	}
	query := make(map[string]struct{})
	for _, rawParameter := range stripeArrayFieldOrEmpty(source.Operation, "parameters") {
		parameter := stripeObjectValue(nil, rawParameter, "source parameter")
		switch stripeStringField(nil, parameter, "in") {
		case "path":
			path[stripeStringField(nil, parameter, "name")] = struct{}{}
		case "query":
			query[stripeStringField(nil, parameter, "name")] = struct{}{}
		}
	}
	return stripeSortedSet(path), stripeSortedSet(query)
}

func stripeSourceMedia(source stripeSourceInfo) ([]string, []string) {
	request := make(map[string]struct{})
	if requestBody := stripeObjectFieldOrEmpty(source.Operation, "requestBody"); requestBody != nil {
		for mediaType := range stripeObjectFieldOrEmpty(requestBody, "content") {
			request[mediaType] = struct{}{}
		}
	}
	response := make(map[string]struct{})
	for _, rawResponse := range stripeObjectFieldOrEmpty(source.Operation, "responses") {
		responseObject := stripeObjectValue(nil, rawResponse, "source response")
		for mediaType := range stripeObjectFieldOrEmpty(responseObject, "content") {
			response[mediaType] = struct{}{}
		}
	}
	return stripeSortedSet(request), stripeSortedSet(response)
}

func stripeSourcePagination(source stripeSourceInfo) stripePaginationFact {
	dataArray, hasMore, nextPage := stripeSourceListResponseFacts(source)
	_, queryParameters := stripeSourceScope(source)
	if source.Method == "GET" && dataArray && hasMore && stripeContains(queryParameters, "starting_after") {
		return stripePaginationFact{Kind: "id_cursor", QueryParameters: []string{"starting_after"}, ResponseFields: []string{"data", "has_more"}}
	}
	if source.Method == "GET" && dataArray && hasMore && stripeContains(queryParameters, "page") && nextPage {
		return stripePaginationFact{Kind: "page_token", QueryParameters: []string{"page"}, ResponseFields: []string{"data", "has_more", "next_page"}}
	}
	if source.Method == "GET" && dataArray && hasMore {
		return stripePaginationFact{Kind: "response_list_without_operation_continuation", QueryParameters: []string{}, ResponseFields: []string{"data", "has_more"}}
	}
	return stripePaginationFact{Kind: "not_documented", QueryParameters: []string{}, ResponseFields: []string{}}
}

func stripeSourceWebhookRegistration(source stripeSourceInfo) stripeWebhookRegistrationFact {
	if source.Method != "POST" {
		return stripeWebhookRegistrationFact{Kind: "not_documented", RequiredRequestFields: []string{}}
	}
	if !strings.Contains(stripeSourceText(source), "webhook endpoint") ||
		!stripeSourceRequestSchemaRequiresFields(source, "enabled_events", "url") {
		return stripeWebhookRegistrationFact{Kind: "not_documented", RequiredRequestFields: []string{}}
	}
	return stripeWebhookRegistrationFact{Kind: "webhook_registration", RequiredRequestFields: []string{"enabled_events", "url"}}
}

func stripeSourceRequestSchemaRequiresFields(source stripeSourceInfo, want ...string) bool {
	requestBody := stripeObjectFieldOrEmpty(source.Operation, "requestBody")
	for _, rawMedia := range stripeObjectFieldOrEmpty(requestBody, "content") {
		media := stripeObjectValue(nil, rawMedia, "source request media")
		schema := stripeObjectFieldOrEmpty(media, "schema")
		fields := make(map[string]struct{})
		for _, rawField := range stripeArrayFieldOrEmpty(schema, "required") {
			field, ok := rawField.(string)
			if ok && field != "" {
				fields[field] = struct{}{}
			}
		}
		matches := true
		for _, field := range want {
			if _, ok := fields[field]; !ok {
				matches = false
				break
			}
		}
		if matches {
			return true
		}
	}
	return false
}

func stripeSourceText(source stripeSourceInfo) string {
	var parts []string
	for _, field := range []string{"summary", "description"} {
		if value, ok := source.Operation[field].(string); ok {
			parts = append(parts, value)
		}
	}
	return strings.ToLower(strings.Join(parts, "\n"))
}

func stripeSourceListResponseFacts(source stripeSourceInfo) (bool, bool, bool) {
	var dataArray, hasMore, nextPage bool
	for _, rawResponse := range stripeObjectFieldOrEmpty(source.Operation, "responses") {
		response := stripeObjectValue(nil, rawResponse, "source response")
		content := stripeObjectFieldOrEmpty(response, "content")
		jsonMedia, exists := content["application/json"]
		if !exists {
			continue
		}
		schema := stripeObjectFieldOrEmpty(stripeObjectValue(nil, jsonMedia, "JSON media"), "schema")
		properties := stripeObjectFieldOrEmpty(schema, "properties")
		data, exists := properties["data"]
		if exists && stripeSchemaType(data) == "array" {
			dataArray = true
		}
		if hasMoreField, exists := properties["has_more"]; exists &&
			stripeSchemaType(hasMoreField) == "boolean" {
			hasMore = true
		}
		if nextPageField, exists := properties["next_page"]; exists &&
			stripeSchemaType(nextPageField) == "string" {
			nextPage = true
		}
	}
	return dataArray, hasMore, nextPage
}

func stripeSchemaType(value any) string {
	schema, ok := value.(map[string]any)
	if !ok {
		return ""
	}
	kind, _ := schema["type"].(string)
	return kind
}

func stripeSourceHasResponseMedia(source stripeSourceInfo, want string) bool {
	_, response := stripeSourceMedia(source)
	return stripeContains(response, want)
}

func stripeSourceHasRequestMedia(source stripeSourceInfo, want string) bool {
	request, _ := stripeSourceMedia(source)
	return stripeContains(request, want)
}

func stripeIsMutation(method string) bool {
	return method == "POST" || method == "DELETE"
}

func stripeMatrixCellByLane(t *testing.T, operation map[string]any, lane string) map[string]any {
	t.Helper()
	cell := stripeMatrixCellByLaneNoTest(operation, lane)
	if cell == nil {
		t.Fatalf("source operation %s has no %s cell", stripeStringField(nil, operation, "source_id"), lane)
	}
	return cell
}

func stripeMatrixCellByLaneNoTest(operation map[string]any, lane string) map[string]any {
	for _, rawCell := range stripeArrayField(nil, operation, "cells") {
		cell := stripeObjectValue(nil, rawCell, "matrix cell")
		if stripeStringField(nil, cell, "lane") == lane {
			return cell
		}
	}
	return nil
}

func stripeMatrixCellCounts(matrix map[string]any) map[string]map[string]int {
	counts := make(map[string]map[string]int, len(stripeLaneOrder))
	for _, lane := range stripeLaneOrder {
		counts[lane] = map[string]int{}
	}
	for _, rawOperation := range stripeArrayField(nil, matrix, "operations") {
		operation := stripeObjectValue(nil, rawOperation, "matrix operation")
		for _, rawCell := range stripeArrayField(nil, operation, "cells") {
			cell := stripeObjectValue(nil, rawCell, "matrix cell")
			lane, state := stripeStringField(nil, cell, "lane"), stripeStringField(nil, cell, "state")
			if _, exists := counts[lane]; !exists {
				counts[lane] = map[string]int{}
			}
			counts[lane][state]++
		}
	}
	return counts
}

func readStripeArtifactRecords(t *testing.T) stripeArtifactRecords {
	t.Helper()
	surface := readStripeJSONObject(t, "api_surface.json")
	streams := readStripeJSONObject(t, "streams.json")
	writes := readStripeJSONObject(t, "writes.json")
	records := stripeArtifactRecords{
		API:     map[string]struct{}{},
		Streams: map[string]struct{}{},
		Writes:  map[string]struct{}{},
	}
	for _, rawEndpoint := range stripeArrayField(t, surface, "endpoints") {
		endpoint := stripeObjectValue(t, rawEndpoint, "api surface endpoint")
		records.API[stripeStringField(t, endpoint, "method")+" "+stripeStringField(t, endpoint, "path")] = struct{}{}
	}
	for _, rawStream := range stripeArrayField(t, streams, "streams") {
		stream := stripeObjectValue(t, rawStream, "stream")
		records.Streams[stripeStringField(t, stream, "name")] = struct{}{}
	}
	for _, rawAction := range stripeArrayField(t, writes, "actions") {
		action := stripeObjectValue(t, rawAction, "write action")
		records.Writes[stripeStringField(t, action, "name")] = struct{}{}
	}
	return records
}

func readStripeJSONObject(t *testing.T, path string) map[string]any {
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

func stripeObjectField(t *testing.T, object map[string]any, name string) map[string]any {
	stripeMarkHelper(t)
	value, exists := object[name]
	if !exists {
		if t != nil {
			t.Fatalf("JSON object is missing %q", name)
		}
		panic(fmt.Sprintf("JSON object is missing %q", name))
	}
	return stripeObjectValue(t, value, name)
}

func stripeObjectFieldOrEmpty(object map[string]any, name string) map[string]any {
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

func stripeObjectValue(t *testing.T, value any, context string) map[string]any {
	stripeMarkHelper(t)
	result, ok := value.(map[string]any)
	if !ok {
		if t != nil {
			t.Fatalf("%s is %T, want JSON object", context, value)
		}
		panic(fmt.Sprintf("%s is %T, want JSON object", context, value))
	}
	return result
}

func stripeArrayField(t *testing.T, object map[string]any, name string) []any {
	stripeMarkHelper(t)
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

func stripeArrayFieldOrEmpty(object map[string]any, name string) []any {
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

func stripeStringArrayField(t *testing.T, object map[string]any, name string) []string {
	stripeMarkHelper(t)
	values := stripeArrayField(t, object, name)
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

func stripeStringArrayFieldOrEmpty(object map[string]any, name string) []string {
	values := stripeArrayFieldOrEmpty(object, name)
	result := make([]string, 0, len(values))
	for _, value := range values {
		text, ok := value.(string)
		if !ok {
			return nil
		}
		result = append(result, text)
	}
	return result
}

func stripeStringField(t *testing.T, object map[string]any, name string) string {
	stripeMarkHelper(t)
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

func stripeNumberField(object map[string]any, name string) int {
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

func stripeSortedSet(values map[string]struct{}) []string {
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func stripeContains(values []string, want string) bool {
	index := sort.SearchStrings(values, want)
	return index < len(values) && values[index] == want
}

func stripeEqualStrings(got, want []string) bool {
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

func stripeEqualStringSet(got, want map[string]struct{}) bool {
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

func stripeMarkHelper(t *testing.T) {
	if t != nil {
		t.Helper()
	}
}
