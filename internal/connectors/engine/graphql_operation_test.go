package engine

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/coordination"
)

// graphQLOperationBundle is deliberately an in-memory, loopback-only bundle.
// Its fixed document and closed variables schema make the tests exercise the
// operation executor rather than a generic GraphQL transport.
func graphQLOperationBundle(baseURL, kind string) Bundle {
	operationName := "GetWidget"
	document := "query GetWidget($id: ID!, $first: Int!, $after: String, $last: Int, $before: String, $filter: WidgetFilterInput) { widget(id: $id, filter: $filter) { id items(first: $first, after: $after, last: $last, before: $before) { nodes { id name secret } pageInfo { hasNextPage hasPreviousPage startCursor endCursor } } } rateLimit { limit cost remaining resetAt } }"
	mutationClass := ""
	approval := "none"
	risk := "low"
	var confirmation *ConfirmationSpec
	if kind == "graphql_mutation" {
		operationName = "DeleteWidget"
		document = "mutation DeleteWidget($id: ID!) { deleteWidget(input: { id: $id }) { deletedId } rateLimit { limit cost remaining resetAt } }"
		mutationClass = "delete"
		approval = "plan-preview-confirm-execute"
		risk = "critical"
		confirmation = &ConfirmationSpec{Kind: connectors.ConfirmationKindDestructive}
	}
	var pagination *GraphQLOperationPaginationSpec
	variablesSchema := json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["id"],
		"properties":{
			"id":{"type":"string"}
		}
	}`)
	if kind == "graphql_query" {
		variablesSchema = json.RawMessage(`{
			"type":"object",
			"additionalProperties":false,
			"required":["id"],
			"properties":{
				"id":{"type":"string"},
				"first":{"type":"integer"},
				"after":{"type":["string","null"]},
				"last":{"type":"integer"},
				"before":{"type":["string","null"]},
				"filter":{
					"type":"object",
					"additionalProperties":false,
					"properties":{
						"status":{"type":"string"},
						"ids":{"type":"array","maxItems":10,"items":{"type":"string"}}
					}
				}
			}
		}`)
		pagination = &GraphQLOperationPaginationSpec{
			ConnectionPath:           "widget.items",
			CursorVariable:           "after",
			PageSizeVariable:         "first",
			BackwardCursorVariable:   "before",
			BackwardPageSizeVariable: "last",
			MaxPageSize:              50,
		}
	}

	return Bundle{
		Name: "acme",
		HTTP: HTTPBase{URL: baseURL},
		Operations: []OperationSpec{{
			ID:            "acme.widgets." + strings.TrimPrefix(kind, "graphql_"),
			Kind:          kind,
			Summary:       "Fixed test GraphQL operation",
			Risk:          risk,
			Approval:      approval,
			OutputPolicy:  "json_redacted",
			MutationClass: mutationClass,
			Destructive:   kind == "graphql_mutation",
			Confirmation:  confirmation,
			GraphQL: &GraphQLOperationSpec{
				Document:        document,
				OperationName:   operationName,
				Path:            "/graphql",
				MaxBytes:        4096,
				VariablesSchema: variablesSchema,
				Pagination:      pagination,
			},
		}},
		Surface: &APISurface{Endpoints: []SurfaceEndpoint{{
			Method: http.MethodPost,
			Path:   "/graphql",
			CoveredBy: &SurfaceCoverage{
				Operations: []string{"acme.widgets." + strings.TrimPrefix(kind, "graphql_")},
			},
			Operation: &SurfaceOperation{
				Model:            "direct_read",
				Status:           "blocked",
				Risk:             "low",
				BlockedByDefault: true,
				Reason:           "fixed GraphQL operation fixture",
			},
		}}},
	}
}

func graphQLMutationWithSensitiveInput(baseURL string) Bundle {
	bundle := graphQLOperationBundle(baseURL, "graphql_mutation")
	op := &bundle.Operations[0]
	op.GraphQL.Document = "mutation DeleteWidget($input: DeleteWidgetInput!) { deleteWidget(input: $input) { deletedId } }"
	op.GraphQL.VariablesSchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["input"],
		"properties":{
			"input":{
				"type":"object",
				"additionalProperties":false,
				"required":["githubPat"],
				"properties":{"githubPat":{"type":"string"}}
			}
		}
	}`)
	op.SensitivePolicy = &SensitivePolicySpec{InputMode: "env", RedactFields: []string{"body.input"}, Transform: "none", ApprovalMode: "typed_confirmation"}
	op.SecretSensitive = true
	op.MutationClass = "secret"
	return bundle
}

type rejectingGraphQLRateLimitClock struct {
	now   time.Time
	waits []time.Duration
}

func (c *rejectingGraphQLRateLimitClock) Now() time.Time { return c.now }

func (c *rejectingGraphQLRateLimitClock) Sleep(_ context.Context, wait time.Duration) error {
	c.waits = append(c.waits, wait)
	return context.DeadlineExceeded
}

func graphQLRateLimitPolicy(id string, limit, windowSeconds int, responseBody bool) connsdk.RateLimitPolicy {
	defaultCost := 1.0
	cost := &connsdk.RateLimitCost{DefaultCost: &defaultCost}
	if responseBody {
		cost.ResponseBody = string(connsdk.RateLimitCostSourceGraphQLRateLimit)
	}
	return connsdk.RateLimitPolicy{
		ID: id,
		Selector: connsdk.RateLimitSelector{
			AuthTypes: []string{"oauth"},
			Endpoints: []connsdk.RateLimitEndpointSelector{{Method: http.MethodPost, Path: "/graphql"}},
		},
		Scope: connsdk.RateLimitScope{SubjectKind: connsdk.RateLimitScopeAccount, SubjectConfig: "account_id"},
		Budgets: []connsdk.RateLimitBudget{{
			Model:         connsdk.RateLimitBudgetFixedWindow,
			Dimension:     connsdk.RateLimitBudgetSustained,
			Unit:          connsdk.RateLimitBudgetPoints,
			Limit:         intPtr(limit),
			WindowSeconds: intPtr(windowSeconds),
			Cost:          cost,
		}},
	}
}

func graphQLRateLimitTestConfig(t *testing.T) connectors.RuntimeConfig {
	t.Helper()
	identity, err := connectors.NewCoordinationIdentity([]byte("graphql-rate-limit-test-salt"), connectors.CredentialBinding{
		BindingID:      "graphql-rate-limit-test-binding",
		ProviderFamily: "github",
		AuthProfile:    "oauth",
	})
	if err != nil {
		t.Fatalf("NewCoordinationIdentity: %v", err)
	}
	return connectors.RuntimeConfig{
		Config:               map[string]string{"account_id": "account-test-001", "auth_type": "oauth"},
		CoordinationIdentity: identity,
	}
}

func graphQLRateLimitResponse(cost, remaining int, resetAt string) string {
	return `{"data":{"widget":{"id":"widget-1","items":{"nodes":[],"pageInfo":{"hasNextPage":false}}},"rateLimit":{"cost":` + strconv.Itoa(cost) + `,"remaining":` + strconv.Itoa(remaining) + `,"resetAt":"` + resetAt + `"}}}`
}

func TestOperationGraphQLRateLimitResponseTightensNextAdmission(t *testing.T) {
	now := time.Date(2026, time.August, 14, 12, 0, 0, 0, time.UTC)
	for _, tt := range []struct {
		name          string
		primaryLimit  int
		primaryWindow int
		cost          int
		remaining     int
		resetAt       string
		wantWait      time.Duration
	}{
		{
			name:          "cost",
			primaryLimit:  4,
			primaryWindow: 3600,
			cost:          4,
			remaining:     4,
			wantWait:      time.Hour,
		},
		{
			name:          "remaining and resetAt",
			primaryLimit:  100,
			primaryWindow: 1,
			cost:          1,
			remaining:     0,
			resetAt:       now.Add(time.Hour).Format(time.RFC3339),
			wantWait:      time.Hour,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls++
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(graphQLRateLimitResponse(tt.cost, tt.remaining, tt.resetAt)))
			}))
			t.Cleanup(server.Close)

			bundle := graphQLOperationBundle(server.URL, "graphql_query")
			bundle.RateLimits = &connsdk.RateLimits{
				State: connsdk.RateLimitStateDeclared,
				Policies: []connsdk.RateLimitPolicy{
					graphQLRateLimitPolicy("graphql-primary", tt.primaryLimit, tt.primaryWindow, true),
				},
			}
			clock := &rejectingGraphQLRateLimitClock{now: now}
			restore := replaceRateLimitRegistryForTest(coordination.NewRateLimitRegistry(clock))
			t.Cleanup(restore)
			request := connectors.OperationDirectReadRequest{
				Operation: "acme.widgets.query",
				Config:    graphQLRateLimitTestConfig(t),
				Body:      map[string]any{"id": "widget-1", "first": 1},
				MaxBytes:  4096,
			}
			if _, err := OperationDirectRead(context.Background(), bundle, request, nil); err != nil {
				t.Fatalf("first GraphQL request: %v", err)
			}
			if _, err := OperationDirectRead(context.Background(), bundle, request, nil); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("second GraphQL request error = %v, want admission deadline", err)
			}
			if got, want := calls, 1; got != want {
				t.Fatalf("blocked GraphQL request reached provider: calls = %d, want %d", got, want)
			}
			if len(clock.waits) != 1 || clock.waits[0] != tt.wantWait {
				t.Fatalf("GraphQL observation wait = %v, want [%s]", clock.waits, tt.wantWait)
			}
		})
	}
}

func TestOperationGraphQLRateLimitCostDoesNotDoubleChargeSecondaryBudget(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(graphQLRateLimitResponse(4, 100, "")))
	}))
	t.Cleanup(server.Close)

	bundle := graphQLOperationBundle(server.URL, "graphql_query")
	secondary := graphQLRateLimitPolicy("graphql-secondary", 10, 60, false)
	secondary.Budgets[0].Cost.DefaultCost = func() *float64 { value := 5.0; return &value }()
	bundle.RateLimits = &connsdk.RateLimits{
		State: connsdk.RateLimitStateDeclared,
		Policies: []connsdk.RateLimitPolicy{
			graphQLRateLimitPolicy("graphql-primary", 100, 3600, true),
			secondary,
		},
	}
	clock := &rejectingGraphQLRateLimitClock{now: time.Date(2026, time.August, 14, 12, 0, 0, 0, time.UTC)}
	restore := replaceRateLimitRegistryForTest(coordination.NewRateLimitRegistry(clock))
	t.Cleanup(restore)
	request := connectors.OperationDirectReadRequest{
		Operation: "acme.widgets.query",
		Config:    graphQLRateLimitTestConfig(t),
		Body:      map[string]any{"id": "widget-1", "first": 1},
		MaxBytes:  4096,
	}
	for attempt := 1; attempt <= 2; attempt++ {
		if _, err := OperationDirectRead(context.Background(), bundle, request, nil); err != nil {
			t.Fatalf("GraphQL request %d: %v", attempt, err)
		}
	}
	if got, want := calls, 2; got != want {
		t.Fatalf("GraphQL secondary budget was charged by primary cost: calls = %d, want %d", got, want)
	}
	if len(clock.waits) != 0 {
		t.Fatalf("GraphQL secondary budget waited after two 5-point admissions: %v", clock.waits)
	}
}

func TestOperationGraphQLRateLimitPrimaryRemainingDoesNotBlockSecondaryBudget(t *testing.T) {
	now := time.Date(2026, time.August, 14, 12, 0, 0, 0, time.UTC)
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(graphQLRateLimitResponse(1, 0, now.Add(time.Hour).Format(time.RFC3339))))
	}))
	t.Cleanup(server.Close)

	primary := graphQLRateLimitPolicy("graphql-primary", 100, 3600, true)
	secondary := graphQLRateLimitPolicy("graphql-secondary", 10, 60, false)
	bundle := graphQLOperationBundle(server.URL, "graphql_query")
	bundle.RateLimits = &connsdk.RateLimits{
		State:    connsdk.RateLimitStateDeclared,
		Policies: []connsdk.RateLimitPolicy{primary, secondary},
	}
	clock := &rejectingGraphQLRateLimitClock{now: now}
	restore := replaceRateLimitRegistryForTest(coordination.NewRateLimitRegistry(clock))
	t.Cleanup(restore)
	request := connectors.OperationDirectReadRequest{
		Operation: "acme.widgets.query",
		Config:    graphQLRateLimitTestConfig(t),
		Body:      map[string]any{"id": "widget-1", "first": 1},
		MaxBytes:  4096,
	}
	if _, err := OperationDirectRead(context.Background(), bundle, request, nil); err != nil {
		t.Fatalf("primary GraphQL request: %v", err)
	}

	// Exercise the same declared secondary policy in a fresh runtime after the
	// primary response reported zero remaining. It must still reach the
	// provider: that body field belongs only to the GraphQL primary family.
	bundle.RateLimits.Policies = []connsdk.RateLimitPolicy{secondary}
	if _, err := OperationDirectRead(context.Background(), bundle, request, nil); err != nil {
		t.Fatalf("secondary GraphQL request: %v", err)
	}
	if got, want := calls, 2; got != want {
		t.Fatalf("primary remaining/reset blocked secondary provider request: calls = %d, want %d", got, want)
	}
	if len(clock.waits) != 0 {
		t.Fatalf("secondary GraphQL budget waited for primary reset: %v", clock.waits)
	}
}

// A structured CLI value is safe for a fixed GraphQL operation only when the
// declaration's own closed variables schema admits that exact top-level
// variable as an object or array.  This is intentionally a narrower question
// than "can this operation accept a JSON body": the latter would recreate a
// generic GraphQL transport through command flags.
func TestValidateGraphQLOperationStructuredJSONVariableRequiresClosedTopLevelContainer(t *testing.T) {
	bundle := graphQLOperationBundle("http://127.0.0.1", "graphql_query")
	op := bundle.Operations[0]

	if err := ValidateGraphQLOperationStructuredJSONVariable(op, "filter"); err != nil {
		t.Fatalf("closed object variable: %v", err)
	}
	for variable, want := range map[string]string{
		"id":            "object or array",
		"missing":       "not declared",
		"filter.status": "top-level",
	} {
		t.Run(variable, func(t *testing.T) {
			if err := ValidateGraphQLOperationStructuredJSONVariable(op, variable); err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("ValidateGraphQLOperationStructuredJSONVariable(%q) = %v, want %q", variable, err, want)
			}
		})
	}

	mutation := graphQLOperationBundle("http://127.0.0.1", "graphql_mutation").Operations[0]
	if err := ValidateGraphQLOperationStructuredJSONVariable(mutation, "id"); err == nil || !strings.Contains(err.Error(), "object or array") {
		t.Fatalf("scalar mutation variable = %v, want closed container rejection", err)
	}
}

func TestGraphQLOperationVariablesRejectsMixedPaginationDirections(t *testing.T) {
	op := graphQLOperationBundle("http://127.0.0.1", "graphql_query").Operations[0]

	for name, variables := range map[string]map[string]any{
		"page sizes": {"id": "widget-1", "first": 10, "last": 10},
		"cursors":    {"id": "widget-1", "after": "next", "before": "previous"},
		"crossed":    {"id": "widget-1", "first": 10, "before": "previous"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := graphQLOperationVariables(op, variables, 0, "")
			if err == nil || !strings.Contains(err.Error(), "mixes forward and backward pagination") {
				t.Fatalf("graphQLOperationVariables() error = %v, want mixed-direction rejection", err)
			}
		})
	}

	if _, err := graphQLOperationVariables(op, map[string]any{"id": "widget-1", "last": 10}, 0, ""); err != nil {
		t.Fatalf("backward-only pagination: %v", err)
	}
}

func TestGraphQLOperationVariablesRequiresExactlyOnePaginationDirection(t *testing.T) {
	op := graphQLOperationBundle("https://example.test", "graphql_query").Operations[0]

	tests := []struct {
		name       string
		variables  map[string]any
		pageCursor string
		wantErr    bool
	}{
		{name: "neither direction", variables: map[string]any{"id": "widget-1"}, wantErr: true},
		{name: "both directions", variables: map[string]any{"id": "widget-1", "first": 10, "last": 10}, wantErr: true},
		{name: "cursor without direction", variables: map[string]any{"id": "widget-1"}, pageCursor: "opaque-next", wantErr: true},
		{name: "forward", variables: map[string]any{"id": "widget-1", "first": 10}},
		{name: "backward", variables: map[string]any{"id": "widget-1", "last": 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := graphQLOperationVariables(op, tt.variables, 0, tt.pageCursor)
			if tt.wantErr {
				if err == nil || !strings.Contains(err.Error(), "exactly one pagination direction") {
					t.Fatalf("graphQLOperationVariables() error = %v, want exact-direction preflight refusal", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("graphQLOperationVariables() error = %v", err)
			}
		})
	}
}

func TestOperationDirectReadBackwardGraphQLPaginationUsesPreviousPageInfo(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPost || r.URL.Path != "/graphql" {
			t.Fatalf("request = %s %s, want POST /graphql", r.Method, r.URL.Path)
		}
		var payload struct {
			Variables map[string]any `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		last := int(payload.Variables["last"].(float64))
		pageInfo := `{"hasPreviousPage":true,"startCursor":"previous-cursor","hasNextPage":false,"endCursor":"unrelated-forward-cursor"}`
		switch last {
		case 3:
			pageInfo = `{"hasPreviousPage":false,"hasNextPage":true,"endCursor":"unrelated-forward-cursor"}`
		case 4:
			pageInfo = `{"hasPreviousPage":true,"hasNextPage":false,"endCursor":"unrelated-forward-cursor"}`
		case 5:
			pageInfo = `{"hasPreviousPage":"true","hasNextPage":false,"endCursor":"unrelated-forward-cursor"}`
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"widget":{"items":{"nodes":[{"id":"item-1"}],"pageInfo":` + pageInfo + `}}}}`))
	}))
	defer server.Close()

	bundle := graphQLOperationBundle(server.URL, "graphql_query")
	for _, tt := range []struct {
		name       string
		last       int
		wantError  string
		wantMore   bool
		wantCursor string
	}{
		{name: "previous page", last: 2, wantMore: true, wantCursor: "previous-cursor"},
		{name: "terminal page", last: 3},
		{name: "missing previous cursor", last: 4, wantError: "startCursor"},
		{name: "malformed previous state", last: 5, wantError: "hasPreviousPage"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			result, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
				Operation: "acme.widgets.query",
				Body:      map[string]any{"id": "widget-1", "last": tt.last},
				MaxBytes:  4096,
			}, nil)
			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Fatalf("OperationDirectRead() error = %v, want %q", err, tt.wantError)
				}
				return
			}
			if err != nil {
				t.Fatalf("OperationDirectRead: %v", err)
			}
			if result.Page.HasMore != tt.wantMore || result.Page.NextCursor != tt.wantCursor || result.Page.Complete != !tt.wantMore {
				t.Fatalf("backward page = %+v, want has_more=%t cursor=%q", result.Page, tt.wantMore, tt.wantCursor)
			}
		})
	}
	if calls != 4 {
		t.Fatalf("query calls = %d, want 4", calls)
	}
}

func TestGraphQLIntUsesSigned32BitDomain(t *testing.T) {
	op := graphQLOperationBundle("https://example.test", "graphql_query").Operations[0]
	op.GraphQL.Pagination = nil
	op.GraphQL.Document = "query IntFixture($count: Int!, $input: Input!, $counts: [Int!]!) { widget(count: $count, input: $input, counts: $counts) { id } }"
	op.GraphQL.VariablesSchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"required":["count","input","counts"],
		"properties":{
			"count":{"type":"integer","minimum":-2147483648,"maximum":2147483647},
			"input":{"type":"object","additionalProperties":false,"required":["count"],"properties":{"count":{"type":"integer","minimum":-2147483648,"maximum":2147483647}}},
			"counts":{"type":"array","maxItems":3,"items":{"type":"integer","minimum":-2147483648,"maximum":2147483647}}
		}
	}`)

	for name, variables := range map[string]map[string]any{
		"signed bounds": {
			"count":  -2147483648,
			"input":  map[string]any{"count": 2147483647},
			"counts": []any{-2147483648, 2147483647},
		},
		"root above maximum": {
			"count":  2147483648,
			"input":  map[string]any{"count": 0},
			"counts": []any{0},
		},
		"nested below minimum": {
			"count":  0,
			"input":  map[string]any{"count": -2147483649},
			"counts": []any{0},
		},
		"list above maximum": {
			"count":  0,
			"input":  map[string]any{"count": 0},
			"counts": []any{2147483648},
		},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := graphQLOperationVariables(op, variables, 0, "")
			if name == "signed bounds" {
				if err != nil {
					t.Fatalf("graphQLOperationVariables() error = %v, want signed bounds accepted", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), "maximum") && !strings.Contains(err.Error(), "minimum") {
				t.Fatalf("graphQLOperationVariables() error = %v, want signed-32-bit preflight refusal", err)
			}
		})
	}
}

func TestOperationDirectReadExecutesFixedGraphQLQueryAndPreservesPartialData(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPost || r.URL.Path != "/graphql" {
			t.Fatalf("request = %s %s, want POST /graphql", r.Method, r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if got, want := payload["operationName"], "GetWidget"; got != want {
			t.Fatalf("operationName = %#v, want %q", got, want)
		}
		query, _ := payload["query"].(string)
		if !strings.HasPrefix(query, "query GetWidget") || strings.Contains(query, "callerSelection") {
			t.Fatalf("query = %q, want only fixed bundle document", query)
		}
		vars, _ := payload["variables"].(map[string]any)
		if got, want := vars["id"], "widget-1"; got != want {
			t.Fatalf("variables.id = %#v, want %q", got, want)
		}
		if got, want := vars["first"], float64(2); got != want {
			t.Fatalf("variables.first = %#v, want %v", got, want)
		}
		if _, found := vars["callerSelection"]; found {
			t.Fatalf("caller-controlled variable reached the request: %#v", vars)
		}
		_, _ = io.Copy(io.Discard, r.Body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"data": {
				"widget": {"id":"widget-1","items":{"nodes":[{"id":"item-1","name":"visible","secret":"response-secret"}],"pageInfo":{"hasNextPage":true,"endCursor":"next-cursor"}}},
				"rateLimit": {"limit":5000,"cost":1,"remaining":4999,"resetAt":"2026-08-09T00:00:00Z"}
			},
			"errors": [{"message":"field warning","path":["widget","items",0],"locations":[{"line":3,"column":7}],"extensions":{"code":"PARTIAL","occurrence_id":"graphql-occurrence-9007199254740993"}}]
		}`))
	}))
	defer server.Close()

	bundle := graphQLOperationBundle(server.URL, "graphql_query")
	result, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
		Operation: "acme.widgets.query",
		Body:      map[string]any{"id": "widget-1", "first": 2},
		MaxBytes:  4096,
	}, nil)
	if err != nil {
		t.Fatalf("OperationDirectRead: %v", err)
	}
	if calls != 1 {
		t.Fatalf("query calls = %d, want 1", calls)
	}
	if result.Method != http.MethodPost || result.Path != "/graphql" || result.Status != http.StatusOK {
		t.Fatalf("result transport = %+v, want POST /graphql 200", result)
	}
	if result.GraphQL == nil || !result.GraphQL.PartialData || len(result.GraphQL.Errors) != 1 {
		t.Fatalf("GraphQL metadata = %+v, want bounded partial-data error metadata", result.GraphQL)
	}
	if got, want := result.GraphQL.Errors[0].Message, "field warning"; got != want {
		t.Fatalf("GraphQL error message = %q, want %q", got, want)
	}
	if result.Operation != "acme.widgets.query" || result.Receipt == nil || !strings.Contains(result.Receipt.BodyRaw, "graphql-occurrence-9007199254740993") {
		t.Fatalf("GraphQL complete receipt = %#v, want full bounded envelope", result.Receipt)
	}
	if result.GraphQL.RateLimit == nil || result.GraphQL.RateLimit.Remaining != 4999 || result.GraphQL.RateLimit.Cost != 1 {
		t.Fatalf("GraphQL rate limit = %+v, want reported bounded data", result.GraphQL.RateLimit)
	}
	if result.Page.Strategy != "graphql_cursor" || !result.Page.HasMore || result.Page.NextCursor != "next-cursor" || result.Page.Complete {
		t.Fatalf("page = %+v, want GraphQL cursor continuation", result.Page)
	}
	body, ok := result.Body.(map[string]any)
	if !ok {
		t.Fatalf("body type = %T, want data object", result.Body)
	}
	widget, _ := body["widget"].(map[string]any)
	items, _ := widget["items"].(map[string]any)
	nodes, _ := items["nodes"].([]any)
	if len(nodes) != 1 {
		t.Fatalf("nodes = %#v, want one fixed-selection item", nodes)
	}
	node, _ := nodes[0].(map[string]any)
	if node["secret"] != "response-secret" {
		t.Fatalf("node = %#v, want ordinary unclassified provider field preserved", node)
	}
}

func TestOperationDirectReadRejectsUntypedGraphQLInputsBeforeNetwork(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls++ }))
	defer server.Close()
	bundle := graphQLOperationBundle(server.URL, "graphql_query")

	for name, request := range map[string]connectors.OperationDirectReadRequest{
		"unknown variable": {
			Operation: "acme.widgets.query",
			Body:      map[string]any{"id": "widget-1", "callerSelection": "everything"},
		},
		"unknown nested variable": {
			Operation: "acme.widgets.query",
			Body: map[string]any{
				"id":     "widget-1",
				"filter": map[string]any{"callerSelection": "everything"},
			},
		},
		"unbounded page number": {
			Operation: "acme.widgets.query",
			Body:      map[string]any{"id": "widget-1"},
			Page:      2,
		},
		"undeclared query override": {
			Operation: "acme.widgets.query",
			Body:      map[string]any{"id": "widget-1"},
			Query:     map[string]string{"raw": "not-a-graphql-variable"},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := OperationDirectRead(context.Background(), bundle, request, nil); err == nil {
				t.Fatal("OperationDirectRead error = nil, want pre-network GraphQL boundary rejection")
			}
		})
	}
	if calls != 0 {
		t.Fatalf("rejected GraphQL inputs reached the network; calls = %d", calls)
	}
}

func TestOperationGraphQLRejectsUnboundVariableSchemaBeforeNetwork(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { calls++ }))
	defer server.Close()
	bundle := graphQLOperationBundle(server.URL, "graphql_query")
	bundle.Operations[0].GraphQL.VariablesSchema = json.RawMessage(`{
		"type":"object",
		"additionalProperties":false,
		"properties":{"callerSelection":{"type":"string"}}
	}`)

	_, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
		Operation: "acme.widgets.query",
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "not referenced by its fixed GraphQL document") {
		t.Fatalf("OperationDirectRead error = %v, want unbound variables-schema rejection", err)
	}
	if calls != 0 {
		t.Fatalf("unbound variable schema reached the network; calls = %d", calls)
	}
}

func TestOperationDirectReadRedactsGraphQLHTTPErrorBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"access_token=fake-test-value"}`))
	}))
	defer server.Close()

	bundle := graphQLOperationBundle(server.URL, "graphql_query")
	_, err := OperationDirectRead(context.Background(), bundle, connectors.OperationDirectReadRequest{
		Operation: "acme.widgets.query",
		Body:      map[string]any{"id": "widget-1"},
		MaxBytes:  4096,
	}, nil)
	if err == nil {
		t.Fatal("OperationDirectRead error = nil, want redacted GraphQL HTTP error")
	}
	if strings.Contains(err.Error(), "fake-test-value") {
		t.Fatalf("GraphQL HTTP error leaked provider body value: %v", err)
	}
	if !strings.Contains(err.Error(), "http 400") {
		t.Fatalf("GraphQL HTTP error = %v, want body-free status diagnostic", err)
	}
}

func TestOperationDirectWriteUsesSharedApprovalForFixedGraphQLMutation(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodPost || r.URL.Path != "/graphql" {
			t.Fatalf("request = %s %s, want POST /graphql", r.Method, r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode mutation payload: %v", err)
		}
		if got, want := payload["operationName"], "DeleteWidget"; got != want {
			t.Fatalf("operationName = %#v, want %q", got, want)
		}
		query, _ := payload["query"].(string)
		if !strings.HasPrefix(query, "mutation DeleteWidget") {
			t.Fatalf("query = %q, want fixed mutation document", query)
		}
		vars, _ := payload["variables"].(map[string]any)
		if got, want := vars["id"], "widget-1"; got != want {
			t.Fatalf("variables.id = %#v, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"deleteWidget":{"deletedId":"widget-1"},"rateLimit":{"limit":4,"cost":4,"remaining":4,"resetAt":""}},"extensions":{"trace_id":"provider-trace"},"provider_fact":"retained"}`))
	}))
	defer server.Close()

	bundle := graphQLOperationBundle(server.URL, "graphql_mutation")
	bundle.RateLimits = &connsdk.RateLimits{
		State: connsdk.RateLimitStateDeclared,
		Policies: []connsdk.RateLimitPolicy{
			graphQLRateLimitPolicy("graphql-mutation-primary", 4, 3600, true),
		},
	}
	clock := &rejectingGraphQLRateLimitClock{now: time.Date(2026, time.August, 14, 12, 0, 0, 0, time.UTC)}
	restore := replaceRateLimitRegistryForTest(coordination.NewRateLimitRegistry(clock))
	t.Cleanup(restore)
	rateConfig := graphQLRateLimitTestConfig(t)
	rateConfig.CredentialRevision = "fixture-credential-revision"
	rateConfig.ConfigurationDigest = "fixture-configuration-digest"
	rateConfig.WriteApprovalScope = connectors.WriteApprovalScopeFixture
	req := connectors.OperationDirectWriteRequest{
		Operation: "acme.widgets.mutation",
		Config:    rateConfig,
		Body:      map[string]any{"id": "widget-1"},
	}
	untyped := req
	untyped.Body = map[string]any{"id": "widget-1", "callerSelection": "everything"}
	if _, err := PreviewOperationDirectWrite(context.Background(), bundle, untyped, nil); err == nil {
		t.Fatal("PreviewOperationDirectWrite accepted an undeclared GraphQL mutation variable")
	}
	if calls != 0 {
		t.Fatalf("undeclared mutation variable reached the network; calls = %d", calls)
	}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	if calls != 0 {
		t.Fatalf("preview reached the network; calls = %d", calls)
	}
	if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err == nil {
		t.Fatal("unapproved GraphQL mutation succeeded")
	}
	if calls != 0 {
		t.Fatalf("unapproved mutation reached the network; calls = %d", calls)
	}

	req.PreviewDigest = preview.Digest
	req.Approval = approvedEvidenceForPreview(t, preview)
	result, err := OperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("approved OperationDirectWrite: %v", err)
	}
	if calls != 1 {
		t.Fatalf("approved mutation calls = %d, want 1", calls)
	}
	if result.GraphQL == nil || result.GraphQL.RateLimit == nil || result.GraphQL.RateLimit.Remaining != 4 {
		t.Fatalf("mutation GraphQL metadata = %+v, want rate-limit data", result.GraphQL)
	}
	body, ok := result.Body.(map[string]any)
	if !ok {
		t.Fatalf("mutation body type = %T, want GraphQL envelope", result.Body)
	}
	data, ok := body["data"].(map[string]any)
	if !ok || data["deleteWidget"] == nil {
		t.Fatalf("mutation body = %#v, want GraphQL data envelope", body)
	}
	if extensions, ok := body["extensions"].(map[string]any); !ok || extensions["trace_id"] != "provider-trace" || body["provider_fact"] != "retained" {
		t.Fatalf("mutation body = %#v, want complete provider envelope", body)
	}

	if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err == nil {
		t.Fatal("replayed GraphQL approval succeeded")
	}
	if calls != 1 {
		t.Fatalf("replayed mutation reached network; calls = %d", calls)
	}

	nextPreview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("next PreviewOperationDirectWrite: %v", err)
	}
	next := req
	next.PreviewDigest = nextPreview.Digest
	next.Approval = approvedEvidenceForPreview(t, nextPreview)
	if _, err := OperationDirectWrite(context.Background(), bundle, next, nil); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("second approved GraphQL mutation error = %v, want rate-limit admission deadline", err)
	}
	if got, want := calls, 1; got != want {
		t.Fatalf("rate-limited GraphQL mutation reached provider: calls = %d, want %d", got, want)
	}
	if len(clock.waits) != 1 || clock.waits[0] != time.Hour {
		t.Fatalf("GraphQL mutation cost observation wait = %v, want [1h]", clock.waits)
	}
}

func TestOperationDirectWritePreservesExplicitNonJSONGraphQLMutationResponse(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		contentType string
		body        []byte
		encoding    string
		raw         string
	}{
		{
			name:        "text",
			contentType: "text/plain",
			body:        []byte(`{"errors":[{"message":"provider text is not a GraphQL envelope"}]}`),
			encoding:    "text",
			raw:         `{"errors":[{"message":"provider text is not a GraphQL envelope"}]}`,
		},
		{
			name:        "binary",
			contentType: "application/octet-stream",
			body:        []byte{0xff, 0x00, 0x01, 0x7f},
			encoding:    "base64",
			raw:         base64.StdEncoding.EncodeToString([]byte{0xff, 0x00, 0x01, 0x7f}),
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls++
				w.Header().Set("Content-Type", testCase.contentType)
				_, _ = w.Write(testCase.body)
			}))
			defer server.Close()

			bundle := graphQLOperationBundle(server.URL, "graphql_mutation")
			req := connectors.OperationDirectWriteRequest{
				Operation: "acme.widgets.mutation",
				Config: connectors.RuntimeConfig{
					CredentialRevision:  "fixture-credential-revision",
					ConfigurationDigest: "fixture-configuration-digest",
					WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
				},
				Body: map[string]any{"id": "widget-1"},
			}
			preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
			if err != nil {
				t.Fatalf("PreviewOperationDirectWrite: %v", err)
			}
			req.PreviewDigest = preview.Digest
			req.Approval = approvedEvidenceForPreview(t, preview)
			result, err := OperationDirectWrite(context.Background(), bundle, req, nil)
			if err != nil {
				t.Fatalf("OperationDirectWrite() = %v", err)
			}
			if calls != 1 || !result.ResponseReceived || !result.BodyPresent || result.BodyBytes != len(testCase.body) || result.BodyRaw != testCase.raw || result.BodyRawEncoding != testCase.encoding {
				t.Fatalf("non-JSON GraphQL mutation result = %#v calls=%d, want exact provider bytes", result, calls)
			}
			if result.Body != nil || result.GraphQL != nil {
				t.Fatalf("non-JSON GraphQL mutation parsed a provider body: %#v", result)
			}
		})
	}
}

func TestOperationDirectWriteFailsClosedOnGraphQLErrors(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"deleteWidget":null},"errors":[{"message":"mutation denied"}]}`))
	}))
	defer server.Close()

	bundle := graphQLOperationBundle(server.URL, "graphql_mutation")
	req := connectors.OperationDirectWriteRequest{
		Operation: "acme.widgets.mutation",
		Config: connectors.RuntimeConfig{
			CredentialRevision:  "fixture-credential-revision",
			ConfigurationDigest: "fixture-configuration-digest",
			WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
		},
		Body: map[string]any{"id": "widget-1"},
	}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	req.PreviewDigest = preview.Digest
	req.Approval = approvedEvidenceForPreview(t, preview)
	result, err := OperationDirectWrite(context.Background(), bundle, req, nil)
	if err == nil || !strings.Contains(strings.ToLower(err.Error()), "graphql") {
		t.Fatalf("OperationDirectWrite error = %v, want GraphQL error rejection", err)
	}
	if strings.Contains(err.Error(), "mutation denied") {
		t.Fatal("OperationDirectWrite error leaked a provider GraphQL detail")
	}
	body, ok := result.Body.(map[string]any)
	if !result.ResponseReceived || !result.BodyPresent || !ok || body["errors"] == nil || result.GraphQL == nil || len(result.GraphQL.Errors) != 1 {
		t.Fatal("GraphQL failed-write result did not retain the received envelope and metadata")
	}
	if got := result.GraphQL.Errors[0].Message; got != "mutation denied" {
		t.Fatalf("GraphQL provider error metadata = %q, want exact provider value", got)
	}
	if calls != 1 {
		t.Fatalf("GraphQL error test calls = %d, want exactly one approved request", calls)
	}
}

func TestOperationDirectWriteRetainsGraphQLApplicationErrorResponse(t *testing.T) {
	const secret = "graphql-provider-canary"
	const responseBody = `{"data":{"deleteWidget":null},"errors":[{"message":"` + secret + `"}]}`
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("X-Provider-Trace", "trace-123")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(responseBody))
	}))
	defer server.Close()

	bundle := graphQLOperationBundle(server.URL, "graphql_mutation")
	bundle.Operations[0].SecretSensitive = true
	req := connectors.OperationDirectWriteRequest{
		Operation: "acme.widgets.mutation",
		Config: connectors.RuntimeConfig{
			CredentialRevision:  "fixture-credential-revision",
			ConfigurationDigest: "fixture-configuration-digest",
			WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
		},
		Body: map[string]any{"id": "widget-1"},
	}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	req.PreviewDigest = preview.Digest
	req.Approval = approvedEvidenceForPreview(t, preview)
	_, err = OperationDirectWrite(context.Background(), bundle, req, nil)
	if err == nil {
		t.Fatal("OperationDirectWrite error = nil, want GraphQL application error")
	}
	if strings.Contains(err.Error(), secret) {
		t.Fatalf("GraphQL application error leaked provider secret: %v", err)
	}
	var providerErr *connsdk.HTTPError
	if !errors.As(err, &providerErr) {
		t.Fatalf("GraphQL application error cause = %T, want provider response error", err)
	}
	if providerErr.Status != http.StatusOK || providerErr.Header.Get("X-Provider-Trace") != "trace-123" || providerErr.Body != responseBody {
		t.Fatal("GraphQL application provider response did not retain exact status, headers, and body")
	}
	if calls != 1 {
		t.Fatalf("GraphQL application error calls = %d, want exactly one approved request", calls)
	}
}

func TestOperationDirectWriteBindsGraphQLQueryAuth(t *testing.T) {
	const token = "graphql-query-key"
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if got := r.URL.Query().Get("api_key"); got != token {
			t.Error("api_key did not match the preview-bound value")
		}
		_, _ = w.Write([]byte(`{"data":{"deleteWidget":{"deletedId":"widget-1"}}}`))
	}))
	defer server.Close()

	bundle := graphQLOperationBundle(server.URL, "graphql_mutation")
	bundle.HTTP.Auth = []AuthSpec{{Mode: "api_key_query", Param: "api_key", Value: "{{ secrets.api_key }}"}}
	req := connectors.OperationDirectWriteRequest{
		Operation: "acme.widgets.mutation",
		Config: connectors.RuntimeConfig{
			CredentialRevision:  "fixture-credential-revision",
			ConfigurationDigest: "fixture-configuration-digest",
			WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
			Secrets:             map[string]string{"api_key": token},
		},
		Body: map[string]any{"id": "widget-1"},
	}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	prepared, err := prepareOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("prepareOperationDirectWrite: %v", err)
	}
	if got := prepared.query.Get("api_key"); got != token {
		t.Fatal("prepared api_key did not retain the exact bound value")
	}
	if got := prepared.prepared.Requests[0].Query; got != "api_key="+token {
		t.Fatal("prepared query did not retain the exact bound value")
	}
	req.PreviewDigest = preview.Digest
	req.Approval = approvedEvidenceForPreview(t, preview)
	if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err != nil {
		t.Fatalf("OperationDirectWrite: %v", err)
	}
	if calls != 1 {
		t.Fatalf("GraphQL query-auth calls = %d, want exactly one approved request", calls)
	}
}

func TestOperationDirectWriteRetainsTerminalGraphQLHTTPResponse(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"access_token=fake-test-value"}`))
	}))
	defer server.Close()

	bundle := graphQLOperationBundle(server.URL, "graphql_mutation")
	req := connectors.OperationDirectWriteRequest{
		Operation: "acme.widgets.mutation",
		Config: connectors.RuntimeConfig{
			CredentialRevision:  "fixture-credential-revision",
			ConfigurationDigest: "fixture-configuration-digest",
			WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
		},
		Body: map[string]any{"id": "widget-1"},
	}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	req.PreviewDigest = preview.Digest
	req.Approval = approvedEvidenceForPreview(t, preview)
	result, err := OperationDirectWrite(context.Background(), bundle, req, nil)
	if err == nil {
		t.Fatal("OperationDirectWrite error = nil, want GraphQL HTTP error")
	}
	if strings.Contains(err.Error(), "fake-test-value") {
		t.Fatal("GraphQL HTTP error leaked a provider body value")
	}
	if !result.ResponseReceived || result.Status != http.StatusBadRequest || result.BodyRaw != `{"message":"access_token=fake-test-value"}` {
		t.Fatal("GraphQL terminal response did not retain the complete provider response")
	}
	if calls != 1 {
		t.Fatalf("GraphQL HTTP error calls = %d, want exactly one approved request", calls)
	}
}

func TestOperationDirectWriteSecretOperationRetainsProviderResponses(t *testing.T) {
	const sensitiveValue = "provider-echoed-github-pat"
	for _, tt := range []struct {
		name       string
		statusCode int
		body       string
	}{
		{
			name:       "graphql errors",
			statusCode: http.StatusOK,
			body:       `{"data":null,"errors":[{"message":"invalid token=` + sensitiveValue + `"}]}`,
		},
		{
			name:       "non JSON HTTP error",
			statusCode: http.StatusBadRequest,
			body:       "invalid githubPat " + sensitiveValue,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			calls := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				calls++
				if tt.name == "graphql errors" {
					w.Header().Set("Content-Type", "application/json")
				}
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			bundle := graphQLMutationWithSensitiveInput(server.URL)
			req := connectors.OperationDirectWriteRequest{
				Operation: "acme.widgets.mutation",
				Config: connectors.RuntimeConfig{
					CredentialRevision:  "fixture-credential-revision",
					ConfigurationDigest: "fixture-configuration-digest",
					WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
				},
				Body: map[string]any{"input": map[string]any{"githubPat": sensitiveValue}},
			}
			preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
			if err != nil {
				t.Fatalf("PreviewOperationDirectWrite: %v", err)
			}
			req.PreviewDigest = preview.Digest
			req.Approval = approvedEvidenceForPreview(t, preview)
			result, err := OperationDirectWrite(context.Background(), bundle, req, nil)
			if err == nil {
				t.Fatal("OperationDirectWrite error = nil, want provider failure")
			}
			if strings.Contains(err.Error(), sensitiveValue) {
				t.Fatal("OperationDirectWrite leaked a provider response through its synthetic error")
			}
			if !result.ResponseReceived || result.BodyRaw != tt.body {
				t.Fatal("OperationDirectWrite result did not retain the exact provider response")
			}
			if calls != 1 {
				t.Fatalf("provider error calls = %d, want exactly one approved request", calls)
			}
		})
	}
}

func TestOperationDirectWriteFailsClosedOnMissingGraphQLData(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls++
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":null}`))
	}))
	defer server.Close()

	bundle := graphQLOperationBundle(server.URL, "graphql_mutation")
	req := connectors.OperationDirectWriteRequest{
		Operation: "acme.widgets.mutation",
		Config: connectors.RuntimeConfig{
			CredentialRevision:  "fixture-credential-revision",
			ConfigurationDigest: "fixture-configuration-digest",
			WriteApprovalScope:  connectors.WriteApprovalScopeFixture,
		},
		Body: map[string]any{"id": "widget-1"},
	}
	preview, err := PreviewOperationDirectWrite(context.Background(), bundle, req, nil)
	if err != nil {
		t.Fatalf("PreviewOperationDirectWrite: %v", err)
	}
	req.PreviewDigest = preview.Digest
	req.Approval = approvedEvidenceForPreview(t, preview)
	if _, err := OperationDirectWrite(context.Background(), bundle, req, nil); err == nil || !strings.Contains(err.Error(), "no data") {
		t.Fatalf("OperationDirectWrite error = %v, want missing GraphQL data rejection", err)
	}
	if calls != 1 {
		t.Fatalf("missing-data mutation calls = %d, want exactly one approved request", calls)
	}
}

func TestPreflightOperationDirectWriteRequiresExactFixedGraphQLBinding(t *testing.T) {
	bundle := graphQLOperationBundle("http://127.0.0.1", "graphql_mutation")
	if err := PreflightOperationDirectWrite(bundle, "acme.widgets.mutation", http.MethodPost, "/graphql", "json_redacted"); err != nil {
		t.Fatalf("PreflightOperationDirectWrite fixed GraphQL binding: %v", err)
	}
	for name, binding := range map[string]struct {
		method string
		path   string
		policy string
	}{
		"wrong method": {method: http.MethodDelete, path: "/graphql", policy: "json_redacted"},
		"wrong path":   {method: http.MethodPost, path: "/graphql/raw", policy: "json_redacted"},
		"wrong policy": {method: http.MethodPost, path: "/graphql", policy: "json"},
	} {
		t.Run(name, func(t *testing.T) {
			if err := PreflightOperationDirectWrite(bundle, "acme.widgets.mutation", binding.method, binding.path, binding.policy); err == nil {
				t.Fatal("PreflightOperationDirectWrite error = nil, want fixed-contract boundary rejection")
			}
		})
	}
	unsafePath := graphQLOperationBundle("http://127.0.0.1", "graphql_mutation")
	unsafePath.Operations[0].GraphQL.Path = "/graphql/../raw"
	if err := PreflightOperationDirectWrite(unsafePath, "acme.widgets.mutation", http.MethodPost, "/graphql/../raw", "json_redacted"); err == nil || !strings.Contains(err.Error(), "traversal") {
		t.Fatalf("PreflightOperationDirectWrite unsafe GraphQL path = %v, want declaration-path rejection", err)
	}
}

func TestValidateOperationDirectWriteMappingsAcceptsDeclaredGraphQLVariables(t *testing.T) {
	bundle := graphQLOperationBundle("https://example.invalid", "graphql_mutation")
	op, err := findOperation(bundle, "acme.widgets.mutation")
	if err != nil {
		t.Fatalf("findOperation: %v", err)
	}
	for _, tc := range []struct {
		name       string
		pathFields []string
		bodyFields []string
		wantErr    string
	}{
		{name: "declared variable", bodyFields: []string{"id"}},
		{name: "undeclared variable", bodyFields: []string{"override"}, wantErr: "is not declared"},
		{name: "nested variable path", bodyFields: []string{"id.owner"}, wantErr: "top-level GraphQL variable"},
		{name: "path override", pathFields: []string{"owner"}, wantErr: "does not accept path fields"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateOperationDirectWriteMappings(op, tc.pathFields, tc.bodyFields)
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("ValidateOperationDirectWriteMappings: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("ValidateOperationDirectWriteMappings error = %v, want %q", err, tc.wantErr)
			}
		})
	}
}

func TestPreflightOperationDirectWriteRequiresNamedGraphQLTransportCoverage(t *testing.T) {
	bundle := graphQLOperationBundle("http://127.0.0.1", "graphql_mutation")
	bundle.Surface.Endpoints[0].CoveredBy.Operations = nil

	err := PreflightOperationDirectWrite(bundle, "acme.widgets.mutation", http.MethodPost, "/graphql", "json_redacted")
	if err == nil || !strings.Contains(err.Error(), "does not cover GraphQL operation") {
		t.Fatalf("PreflightOperationDirectWrite missing named GraphQL transport coverage = %v, want fail-closed coverage rejection", err)
	}
}

func TestPreflightOperationDirectReadRejectsMixedFixedGraphQLDocument(t *testing.T) {
	bundle := graphQLOperationBundle("http://127.0.0.1", "graphql_query")
	// A document that starts as a query but selects a separately named mutation
	// through operationName would turn a direct_read command into a write.
	bundle.Operations[0].GraphQL.Document += " mutation DeleteWidget($id: ID!) { deleteWidget(input: { id: $id }) { deletedId } }"
	bundle.Operations[0].GraphQL.OperationName = "DeleteWidget"

	err := PreflightOperationDirectRead(bundle, "acme.widgets.query", http.MethodPost, "/graphql", 4096, "json_redacted")
	if err == nil || !strings.Contains(err.Error(), "exactly one named query") {
		t.Fatalf("PreflightOperationDirectRead mixed GraphQL document = %v, want fixed query-only rejection", err)
	}
}
