package contractv1_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	contractv1 "polymetrics.ai/internal/pmbroker/contract/v1"
)

func TestHTTPClientLoopbackAndRemoteShareTypedSemantics(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	fixtures := contractv1.AcceptedSyntheticFixtures()

	endpoints := []struct {
		name     string
		endpoint string
	}{
		{name: "loopback", endpoint: "http://127.0.0.1:18080"},
		{name: "container service", endpoint: "http://pm-broker:8080"},
		{name: "container internal dns", endpoint: "http://pm-broker.internal:8080"},
	}

	for _, tt := range endpoints {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			broker := contractv1.NewSyntheticBroker()
			client, err := broker.NewHTTPClient(tt.endpoint, syntheticAuthorization{},
				contractv1.WithClientCorrelationIDProvider(staticCorrelationID("corr_0123456789abcdea")),
			)
			if err != nil {
				t.Fatalf("NewHTTPClient() error = %v", err)
			}

			compatibility, err := client.NegotiateCompatibility(ctx)
			if err != nil {
				t.Fatalf("NegotiateCompatibility() error = %v", err)
			}
			if !reflect.DeepEqual(compatibility, fixtures.Compatibility) {
				t.Fatalf("NegotiateCompatibility() = %#v, want %#v", compatibility, fixtures.Compatibility)
			}

			page, err := client.ListConnectorConnections(ctx, contractv1.Pagination{Limit: 1})
			if err != nil {
				t.Fatalf("ListConnectorConnections() error = %v", err)
			}
			if len(page.ConnectorConnections) != 1 || !reflect.DeepEqual(page.ConnectorConnections[0], fixtures.ConnectorConnection) {
				t.Fatalf("ListConnectorConnections() = %#v, want fixture connection", page)
			}

			connection, err := client.ConnectorConnection(ctx, fixtures.ConnectorConnection.ConnectorConnectionID)
			if err != nil {
				t.Fatalf("ConnectorConnection() error = %v", err)
			}
			if !reflect.DeepEqual(connection, fixtures.ConnectorConnection) {
				t.Fatalf("ConnectorConnection() = %#v, want %#v", connection, fixtures.ConnectorConnection)
			}

			plan, err := client.CreateExecutionPlan(ctx, fixtures.ExecutionPlanRequest)
			if err != nil {
				t.Fatalf("CreateExecutionPlan() error = %v", err)
			}
			if !reflect.DeepEqual(plan, fixtures.ExecutionPlan) {
				t.Fatalf("CreateExecutionPlan() = %#v, want %#v", plan, fixtures.ExecutionPlan)
			}
		})
	}
}

func TestHTTPClientRejectsConnectorConnectionPageAboveRequestedLimit(t *testing.T) {
	t.Parallel()

	fixtures := contractv1.AcceptedSyntheticFixtures()
	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.URL.Query().Get("limit") != "1" {
			t.Errorf("limit query = %q, want 1", request.URL.Query().Get("limit"))
			http.Error(response, "bad limit", http.StatusBadRequest)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		page := contractv1.ConnectorConnectionPage{
			ConnectorConnections: []contractv1.ConnectorConnection{
				fixtures.ConnectorConnection,
				fixtures.ConnectorConnection,
			},
		}
		if err := json.NewEncoder(response).Encode(page); err != nil {
			t.Errorf("encode connector connection page: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client, err := contractv1.NewHTTPClient(server.URL, syntheticAuthorization{})
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}

	_, err = client.ListConnectorConnections(context.Background(), contractv1.Pagination{Limit: 1})
	if !errors.Is(err, contractv1.ErrInvalidPagination) {
		t.Fatalf("ListConnectorConnections() error = %v, want %v", err, contractv1.ErrInvalidPagination)
	}
}

func TestHTTPClientRejectsOversizedBrokerResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		response.Header().Set("Content-Type", "application/json")
		if _, err := response.Write([]byte(`{"connector_connections":[]}`)); err != nil {
			t.Errorf("write connector connection page: %v", err)
			return
		}
		if _, err := response.Write(bytes.Repeat([]byte(" "), 2<<20)); err != nil {
			t.Errorf("write oversized response padding: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client, err := contractv1.NewHTTPClient(server.URL, syntheticAuthorization{})
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}

	_, err = client.ListConnectorConnections(context.Background(), contractv1.Pagination{Limit: 1})
	if !errors.Is(err, contractv1.ErrUnexpectedResponse) {
		t.Fatalf("ListConnectorConnections() error = %v, want %v", err, contractv1.ErrUnexpectedResponse)
	}
}

func TestHTTPClientAuthenticationCorrelationIdempotencyAndDigestTransport(t *testing.T) {
	t.Parallel()

	fixtures := contractv1.AcceptedSyntheticFixtures()
	broker := contractv1.NewSyntheticBroker()
	authenticator := &countingAuthorization{}

	client, err := broker.NewHTTPClient("http://127.0.0.1:18080", authenticator,
		contractv1.WithClientCorrelationIDProvider(staticCorrelationID("corr_0123456789abcdeb")),
	)
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}

	plan, err := client.CreateExecutionPlan(context.Background(), fixtures.ExecutionPlanRequest)
	if err != nil {
		t.Fatalf("CreateExecutionPlan() error = %v", err)
	}
	if plan.Digest != fixtures.ExecutionPlan.Digest {
		t.Fatalf("Digest = %q, want %q", plan.Digest, fixtures.ExecutionPlan.Digest)
	}
	if authenticator.calls != 1 {
		t.Fatalf("authenticator calls = %d, want 1", authenticator.calls)
	}

	request := lastObservedRequest(t, broker)
	if request.APIVersion != contractv1.ContractVersion1 {
		t.Fatalf("%s = %q, want %q", contractv1.HeaderAPIVersion, request.APIVersion, contractv1.ContractVersion1)
	}
	if request.IdempotencyKey != fixtures.ExecutionPlanRequest.IdempotencyKey {
		t.Fatalf("%s = %q, want %q", contractv1.HeaderIdempotencyKey, request.IdempotencyKey, fixtures.ExecutionPlanRequest.IdempotencyKey)
	}
	if request.CorrelationID != "corr_0123456789abcdeb" {
		t.Fatalf("%s = %q, want deterministic correlation ID", contractv1.HeaderCorrelationID, request.CorrelationID)
	}
	if !request.ExplicitAuthorization {
		t.Fatal("explicit authorization was false; typed request did not use the explicit auth seam")
	}
	if request.CookiePresent {
		t.Fatal("CookiePresent = true, want false")
	}
}

func TestHTTPClientRejectsUnsafeEndpointHostOriginAndAmbientCookies(t *testing.T) {
	t.Parallel()

	for _, endpoint := range []string{
		"https://userinfo@pm-broker.example",
		"https://pm-broker.example/v1",
		"https://pm-broker.example/v1?workspace=synthetic",
		"https://pm-broker.example/#fragment",
		"http://pm-broker.example",
		"http://anything",
		"http://evil.internal",
		"http://pm-broker.evil.internal",
		"grpc://pm-broker.example",
		"unix:///tmp/pm-broker.sock",
		"http://",
	} {
		t.Run(endpoint, func(t *testing.T) {
			t.Parallel()

			_, err := contractv1.NewHTTPClient(endpoint, syntheticAuthorization{})
			if !errors.Is(err, contractv1.ErrInvalidEndpoint) {
				t.Fatalf("NewHTTPClient(%q) error = %v, want %v", endpoint, err, contractv1.ErrInvalidEndpoint)
			}
			if err != nil && strings.Contains(err.Error(), endpoint) {
				t.Fatalf("endpoint error leaked raw URL %q in %q", endpoint, err.Error())
			}
		})
	}

	broker := contractv1.NewSyntheticBroker()
	unsafeHost := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/v1/compatibility", nil)
	unsafeHost.Host = "127.0.0.1:18080@evil.invalid"
	unsafeHostRecorder := httptest.NewRecorder()
	broker.ServeHTTP(unsafeHostRecorder, unsafeHost)
	if unsafeHostRecorder.Code != http.StatusBadRequest {
		t.Fatalf("unsafe Host status = %d, want %d", unsafeHostRecorder.Code, http.StatusBadRequest)
	}

	if err := broker.AllowEndpoint("http://127.0.0.1:18080"); err != nil {
		t.Fatalf("AllowEndpoint() error = %v", err)
	}

	unsafeOrigin := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/v1/compatibility", nil)
	unsafeOrigin.Host = "127.0.0.1:18080"
	unsafeOrigin.Header.Set("Origin", "http://evil.invalid")
	unsafeOriginRecorder := httptest.NewRecorder()
	broker.ServeHTTP(unsafeOriginRecorder, unsafeOrigin)
	if unsafeOriginRecorder.Code != http.StatusForbidden {
		t.Fatalf("unsafe Origin status = %d, want %d", unsafeOriginRecorder.Code, http.StatusForbidden)
	}

	fixtures := contractv1.AcceptedSyntheticFixtures()
	payload, err := json.Marshal(fixtures.ExecutionPlanRequest)
	if err != nil {
		t.Fatalf("marshal execution plan request: %v", err)
	}
	cookieOnly := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/v1/execution-plans", bytes.NewReader(payload))
	cookieOnly.Host = "127.0.0.1:18080"
	cookieOnly.Header.Set(contractv1.HeaderAPIVersion, string(contractv1.ContractVersion1))
	cookieOnly.Header.Set(contractv1.HeaderIdempotencyKey, string(fixtures.ExecutionPlanRequest.IdempotencyKey))
	cookieOnly.Header.Set(contractv1.HeaderCorrelationID, "corr_0123456789abcdec")
	cookieOnly.Header.Set("Content-Type", "application/json")
	cookieOnly.Header.Set("Cookie", "session=ambient")
	cookieRecorder := httptest.NewRecorder()
	broker.ServeHTTP(cookieRecorder, cookieOnly)
	if cookieRecorder.Code != http.StatusBadRequest {
		t.Fatalf("cookie-only mutation status = %d, want %d", cookieRecorder.Code, http.StatusBadRequest)
	}
	if got := broker.ExecutionPlanRequestCount(); got != 0 {
		t.Fatalf("ExecutionPlanRequestCount() = %d, want 0 after cookie-only mutation", got)
	}
}

func TestHTTPClientStructuredErrorsRateLimitsAndCompatibilityNegotiation(t *testing.T) {
	t.Parallel()

	fixtures := contractv1.AcceptedSyntheticFixtures()

	t.Run("exact incompatible contract version", func(t *testing.T) {
		t.Parallel()

		broker := contractv1.NewSyntheticBroker()
		client, err := broker.NewHTTPClient("http://127.0.0.1:18080", syntheticAuthorization{},
			contractv1.WithClientContractVersion(contractv1.ContractVersion("0.9")),
		)
		if err != nil {
			t.Fatalf("NewHTTPClient() error = %v", err)
		}

		_, err = client.CreateExecutionPlan(context.Background(), fixtures.ExecutionPlanRequest)
		if err == nil {
			t.Fatal("CreateExecutionPlan() error = nil, want version refusal")
		}
		var refusal *contractv1.IncompatibleContractVersionError
		if !errors.As(err, &refusal) {
			t.Fatalf("CreateExecutionPlan() error type = %T, want IncompatibleContractVersionError", err)
		}
		if refusal.StatusCode != http.StatusUpgradeRequired {
			t.Fatalf("StatusCode = %d, want %d", refusal.StatusCode, http.StatusUpgradeRequired)
		}
		if refusal.Response.Error.Code != contractv1.ErrorCodeIncompatibleContractVersion {
			t.Fatalf("error code = %q, want %q", refusal.Response.Error.Code, contractv1.ErrorCodeIncompatibleContractVersion)
		}
	})

	t.Run("negotiate rejects configured unsupported version", func(t *testing.T) {
		t.Parallel()

		broker := contractv1.NewSyntheticBroker()
		client, err := broker.NewHTTPClient("http://127.0.0.1:18080", syntheticAuthorization{},
			contractv1.WithClientContractVersion(contractv1.ContractVersion("0.9")),
		)
		if err != nil {
			t.Fatalf("NewHTTPClient() error = %v", err)
		}

		_, err = client.NegotiateCompatibility(context.Background())
		if err == nil {
			t.Fatal("NegotiateCompatibility() error = nil, want version refusal")
		}
		var refusal *contractv1.IncompatibleContractVersionError
		if !errors.As(err, &refusal) {
			t.Fatalf("NegotiateCompatibility() error type = %T, want IncompatibleContractVersionError", err)
		}
		if refusal.StatusCode != http.StatusUpgradeRequired {
			t.Fatalf("StatusCode = %d, want %d", refusal.StatusCode, http.StatusUpgradeRequired)
		}
		if refusal.Response.Error.Code != contractv1.ErrorCodeIncompatibleContractVersion {
			t.Fatalf("error code = %q, want %q", refusal.Response.Error.Code, contractv1.ErrorCodeIncompatibleContractVersion)
		}
	})

	t.Run("rate limit metadata", func(t *testing.T) {
		t.Parallel()

		broker := contractv1.NewSyntheticBroker()
		broker.RateLimitNextMutation(3 * time.Second)
		client, err := broker.NewHTTPClient("http://127.0.0.1:18080", syntheticAuthorization{})
		if err != nil {
			t.Fatalf("NewHTTPClient() error = %v", err)
		}

		_, err = client.CreateExecutionPlan(context.Background(), fixtures.ExecutionPlanRequest)
		if err == nil {
			t.Fatal("CreateExecutionPlan() error = nil, want rate-limit error")
		}
		var brokerErr *contractv1.BrokerError
		if !errors.As(err, &brokerErr) {
			t.Fatalf("CreateExecutionPlan() error type = %T, want BrokerError", err)
		}
		if brokerErr.StatusCode != http.StatusTooManyRequests {
			t.Fatalf("StatusCode = %d, want %d", brokerErr.StatusCode, http.StatusTooManyRequests)
		}
		if brokerErr.Response.Error.Code != contractv1.ErrorCodeRateLimited {
			t.Fatalf("error code = %q, want %q", brokerErr.Response.Error.Code, contractv1.ErrorCodeRateLimited)
		}
		if brokerErr.RateLimit == nil || brokerErr.RateLimit.RetryAfter != 3*time.Second {
			t.Fatalf("RateLimit = %#v, want retry-after 3s", brokerErr.RateLimit)
		}
	})

	t.Run("digest header mismatch", func(t *testing.T) {
		t.Parallel()

		broker := contractv1.NewSyntheticBroker()
		broker.CorruptNextExecutionPlanDigestHeader()
		client, err := broker.NewHTTPClient("http://127.0.0.1:18080", syntheticAuthorization{})
		if err != nil {
			t.Fatalf("NewHTTPClient() error = %v", err)
		}

		_, err = client.CreateExecutionPlan(context.Background(), fixtures.ExecutionPlanRequest)
		if !errors.Is(err, contractv1.ErrInvalidExecutionPlan) {
			t.Fatalf("CreateExecutionPlan() error = %v, want %v", err, contractv1.ErrInvalidExecutionPlan)
		}
	})

	t.Run("safe structured not found error", func(t *testing.T) {
		t.Parallel()

		broker := contractv1.NewSyntheticBroker()
		client, err := broker.NewHTTPClient("http://pm-broker.internal:8080", syntheticAuthorization{})
		if err != nil {
			t.Fatalf("NewHTTPClient() error = %v", err)
		}

		_, err = client.ConnectorConnection(context.Background(), contractv1.ConnectorConnectionID("ccn_aaaaaaaaaaaaaaaa"))
		if err == nil {
			t.Fatal("ConnectorConnection() error = nil, want structured not-found error")
		}
		var brokerErr *contractv1.BrokerError
		if !errors.As(err, &brokerErr) {
			t.Fatalf("ConnectorConnection() error type = %T, want BrokerError", err)
		}
		if brokerErr.Response.Error.Code != contractv1.ErrorCodeNotFound {
			t.Fatalf("error code = %q, want %q", brokerErr.Response.Error.Code, contractv1.ErrorCodeNotFound)
		}
		if !brokerErr.Response.Error.CorrelationID.IsSafe() {
			t.Fatalf("correlation ID %q was not safe", brokerErr.Response.Error.CorrelationID)
		}
	})
}

func TestHTTPClientAcceptsEnumConnectorConnectionResponses(t *testing.T) {
	t.Parallel()

	fixtures := contractv1.AcceptedSyntheticFixtures()
	connection := fixtures.ConnectorConnection
	connection.ConnectorKind = "gcp"
	connection.Status = "not_ready"
	connection.WriteMode = "plan_only"

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			t.Errorf("method = %s, want %s", request.Method, http.MethodGet)
			http.Error(response, "bad method", http.StatusBadRequest)
			return
		}
		if request.URL.Path != "/v1/connector-connections/"+string(connection.ConnectorConnectionID) {
			t.Errorf("path = %s, want connector connection path", request.URL.Path)
			http.Error(response, "bad path", http.StatusBadRequest)
			return
		}
		response.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(response).Encode(connection); err != nil {
			t.Errorf("encode connector connection response: %v", err)
		}
	}))
	t.Cleanup(server.Close)

	client, err := contractv1.NewHTTPClient(server.URL, syntheticAuthorization{})
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}

	got, err := client.ConnectorConnection(context.Background(), connection.ConnectorConnectionID)
	if err != nil {
		t.Fatalf("ConnectorConnection() error = %v", err)
	}
	if !reflect.DeepEqual(got, connection) {
		t.Fatalf("ConnectorConnection() = %#v, want %#v", got, connection)
	}
}

func TestHTTPClientRejectsExecutionPlanResponseRequestMismatch(t *testing.T) {
	t.Parallel()

	fixtures := contractv1.AcceptedSyntheticFixtures()
	tests := []struct {
		name   string
		mutate func(*contractv1.ExecutionPlan)
	}{
		{
			name: "organization",
			mutate: func(plan *contractv1.ExecutionPlan) {
				plan.OrganizationID = contractv1.OrganizationID("org_aaaaaaaaaaaaaaaa")
			},
		},
		{
			name: "workspace",
			mutate: func(plan *contractv1.ExecutionPlan) {
				plan.WorkspaceID = contractv1.WorkspaceID("wks_aaaaaaaaaaaaaaaa")
			},
		},
		{
			name: "environment",
			mutate: func(plan *contractv1.ExecutionPlan) {
				plan.EnvironmentID = contractv1.EnvironmentID("env_aaaaaaaaaaaaaaaa")
			},
		},
		{
			name: "broker profile",
			mutate: func(plan *contractv1.ExecutionPlan) {
				plan.BrokerProfileID = contractv1.BrokerProfileID("bpf_aaaaaaaaaaaaaaaa")
			},
		},
		{
			name: "connector connection",
			mutate: func(plan *contractv1.ExecutionPlan) {
				plan.ConnectorConnectionID = contractv1.ConnectorConnectionID("ccn_aaaaaaaaaaaaaaaa")
				plan.Intent.ValidateConnectorConnection.ConnectorConnectionID = contractv1.ConnectorConnectionID("ccn_aaaaaaaaaaaaaaaa")
			},
		},
		{
			name: "idempotency key",
			mutate: func(plan *contractv1.ExecutionPlan) {
				plan.IdempotencyKey = contractv1.IdempotencyKey("idem_aaaaaaaaaaaaaaaaaaaa")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			plan := cloneExecutionPlan(fixtures.ExecutionPlan)
			tt.mutate(&plan)
			server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
				if request.Method != http.MethodPost {
					t.Errorf("method = %s, want %s", request.Method, http.MethodPost)
					http.Error(response, "bad method", http.StatusBadRequest)
					return
				}
				if request.URL.Path != "/v1/execution-plans" {
					t.Errorf("path = %s, want /v1/execution-plans", request.URL.Path)
					http.Error(response, "bad path", http.StatusBadRequest)
					return
				}
				response.Header().Set("Content-Type", "application/json")
				response.Header().Set(contractv1.HeaderExecutionPlanDigest, string(plan.Digest))
				if err := json.NewEncoder(response).Encode(plan); err != nil {
					t.Errorf("encode execution plan response: %v", err)
				}
			}))
			t.Cleanup(server.Close)

			client, err := contractv1.NewHTTPClient(server.URL, syntheticAuthorization{})
			if err != nil {
				t.Fatalf("NewHTTPClient() error = %v", err)
			}

			_, err = client.CreateExecutionPlan(context.Background(), fixtures.ExecutionPlanRequest)
			if !errors.Is(err, contractv1.ErrInvalidExecutionPlan) {
				t.Fatalf("CreateExecutionPlan() error = %v, want %v", err, contractv1.ErrInvalidExecutionPlan)
			}
		})
	}
}

func cloneExecutionPlan(plan contractv1.ExecutionPlan) contractv1.ExecutionPlan {
	if plan.Intent.ValidateConnectorConnection != nil {
		intent := *plan.Intent.ValidateConnectorConnection
		plan.Intent.ValidateConnectorConnection = &intent
	}
	return plan
}

func TestHTTPClientSafetySurfaceNoCredentialsNoGRPCNoGenericEscape(t *testing.T) {
	t.Parallel()

	client, err := contractv1.NewHTTPClient("https://pm-broker.example", syntheticAuthorization{})
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}

	diagnostics := client.Diagnostics()
	diagnosticPayload, err := json.Marshal(diagnostics)
	if err != nil {
		t.Fatalf("marshal diagnostics: %v", err)
	}
	unsafeVersionClient, err := contractv1.NewHTTPClient("https://pm-broker.example", syntheticAuthorization{},
		contractv1.WithClientContractVersion(contractv1.ContractVersion("1.0\nX-Bad: value")),
	)
	if err != nil {
		t.Fatalf("NewHTTPClient() with unsafe diagnostic version error = %v", err)
	}
	unsafeVersionDiagnostics, err := json.Marshal(unsafeVersionClient.Diagnostics())
	if err != nil {
		t.Fatalf("marshal unsafe version diagnostics: %v", err)
	}
	if strings.Contains(string(unsafeVersionDiagnostics), "X-Bad") || !strings.Contains(string(unsafeVersionDiagnostics), "invalid") {
		t.Fatalf("unsafe version diagnostics = %s, want redacted invalid version", unsafeVersionDiagnostics)
	}
	for _, forbidden := range []string{
		"userinfo",
		"Authorization",
		"grpc://",
		"unix://",
		strings.Repeat("x", 24),
	} {
		if strings.Contains(string(diagnosticPayload), forbidden) {
			t.Fatalf("diagnostics leaked forbidden marker %q in %s", forbidden, diagnosticPayload)
		}
	}

	clientType := reflect.TypeOf(client)
	for _, forbiddenMethod := range []string{
		"Do",
		"RoundTrip",
		"Request",
		"RawJSON",
		"WithHeader",
		"WithEndpoint",
		"WithURL",
		"WithBody",
		"GRPC",
		"gRPC",
		"Socket",
		"Unix",
	} {
		if _, ok := clientType.MethodByName(forbiddenMethod); ok {
			t.Fatalf("Client exposes forbidden escape-hatch method %q", forbiddenMethod)
		}
	}

	for _, endpoint := range []string{"grpc://pm-broker.example", "http+unix://pm-broker.sock"} {
		_, err := contractv1.NewHTTPClient(endpoint, syntheticAuthorization{})
		if !errors.Is(err, contractv1.ErrInvalidEndpoint) {
			t.Fatalf("NewHTTPClient(%q) error = %v, want %v", endpoint, err, contractv1.ErrInvalidEndpoint)
		}
	}
}

func TestHTTPClientTypedNilAdaptersFailSafely(t *testing.T) {
	t.Parallel()

	fixtures := contractv1.AcceptedSyntheticFixtures()
	broker := contractv1.NewSyntheticBroker()
	var nilAuth contractv1.AuthenticatorFunc
	authClient, err := broker.NewHTTPClient("http://127.0.0.1:18080", nilAuth)
	if err != nil {
		t.Fatalf("NewHTTPClient() error = %v", err)
	}
	_, err = authClient.ConnectorConnection(context.Background(), fixtures.ConnectorConnection.ConnectorConnectionID)
	if !errors.Is(err, contractv1.ErrAuthenticationFailed) {
		t.Fatalf("ConnectorConnection() error = %v, want %v", err, contractv1.ErrAuthenticationFailed)
	}

	var nilCorrelation contractv1.CorrelationIDProviderFunc
	correlationClient, err := broker.NewHTTPClient("http://127.0.0.1:18080", syntheticAuthorization{},
		contractv1.WithClientCorrelationIDProvider(nilCorrelation),
	)
	if err != nil {
		t.Fatalf("NewHTTPClient() with nil correlation provider error = %v", err)
	}
	_, err = correlationClient.Compatibility(context.Background())
	if !errors.Is(err, contractv1.ErrInvalidCorrelationID) {
		t.Fatalf("Compatibility() error = %v, want %v", err, contractv1.ErrInvalidCorrelationID)
	}
}

func lastObservedRequest(t *testing.T, broker *contractv1.SyntheticBroker) contractv1.ObservedRequest {
	t.Helper()
	requests := broker.ObservedRequests()
	if len(requests) == 0 {
		t.Fatal("no requests observed")
	}
	return requests[len(requests)-1]
}

type syntheticAuthorization struct{}

func (syntheticAuthorization) PMBrokerAuthorization(context.Context) (contractv1.Authorization, error) {
	return contractv1.NewAuthorization("PMBroker", strings.Repeat("x", 24))
}

type countingAuthorization struct {
	calls int
}

func (auth *countingAuthorization) PMBrokerAuthorization(context.Context) (contractv1.Authorization, error) {
	auth.calls++
	return contractv1.NewAuthorization("PMBroker", strings.Repeat("x", 24))
}

type staticCorrelationID string

func (id staticCorrelationID) PMBrokerCorrelationID(context.Context) (contractv1.CorrelationID, error) {
	return contractv1.CorrelationID(id), nil
}
