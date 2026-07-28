package contractv1

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
)

const syntheticBrokerBaseURL = "http://pm-broker.synthetic.invalid"

// SyntheticBroker is a deterministic, network-free fake PM Broker.
type SyntheticBroker struct {
	mu                    sync.Mutex
	fixtures              SyntheticFixtures
	executionPlanRequests []ExecutionPlanRequest
}

// NewSyntheticBroker returns a fake broker loaded with the accepted /v1 fixtures.
func NewSyntheticBroker() *SyntheticBroker {
	return &SyntheticBroker{fixtures: AcceptedSyntheticFixtures()}
}

// NewClient returns a typed client wired to this in-memory fake broker.
func (broker *SyntheticBroker) NewClient(opts ...ClientOption) *Client {
	config := clientConfig{contractVersion: ContractVersion1}
	for _, opt := range opts {
		opt(&config)
	}
	return &Client{
		contractVersion: config.contractVersion,
		httpClient: &http.Client{
			Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
				select {
				case <-request.Context().Done():
					return nil, request.Context().Err()
				default:
				}

				recorder := httptest.NewRecorder()
				broker.serveHTTP(recorder, request)
				return recorder.Result(), nil
			}),
		},
	}
}

// ExecutionPlanRequestCount returns how many execution-plan requests reached execution.
func (broker *SyntheticBroker) ExecutionPlanRequestCount() int {
	broker.mu.Lock()
	defer broker.mu.Unlock()
	return len(broker.executionPlanRequests)
}

// Client is the typed synthetic PM Broker /v1 client.
type Client struct {
	contractVersion ContractVersion
	httpClient      *http.Client
}

// ClientOption configures the synthetic client without exposing arbitrary transport fields.
type ClientOption func(*clientConfig)

type clientConfig struct {
	contractVersion ContractVersion
}

// WithClientContractVersion sets the version sent on typed /v1 operations.
func WithClientContractVersion(version ContractVersion) ClientOption {
	return func(config *clientConfig) {
		config.contractVersion = version
	}
}

// Compatibility returns /v1 compatibility information. This method does not send
// PM-Broker-API-Version because the accepted contract exempts compatibility discovery.
func (client *Client) Compatibility(ctx context.Context) (Compatibility, error) {
	return doJSON[Compatibility](ctx, client, http.MethodGet, "/v1/compatibility", nil, false)
}

// ConnectorConnection returns the deterministic synthetic connector connection fixture.
func (client *Client) ConnectorConnection(ctx context.Context, id ConnectorConnectionID) (ConnectorConnection, error) {
	if !id.IsValid() {
		return ConnectorConnection{}, ErrInvalidIdentityBoundary
	}
	return doJSON[ConnectorConnection](ctx, client, http.MethodGet, "/v1/connector-connections/"+string(id), nil, true)
}

// CreateExecutionPlan sends a typed execution-plan request to the synthetic broker.
func (client *Client) CreateExecutionPlan(ctx context.Context, request ExecutionPlanRequest) (ExecutionPlan, error) {
	if err := request.Validate(); err != nil {
		return ExecutionPlan{}, err
	}
	return doJSON[ExecutionPlan](ctx, client, http.MethodPost, "/v1/execution-plans", request, true)
}

func doJSON[T any](ctx context.Context, client *Client, method string, target string, requestBody any, sendVersion bool) (T, error) {
	var zero T
	body, err := encodeRequestBody(requestBody)
	if err != nil {
		return zero, err
	}
	request, err := http.NewRequestWithContext(ctx, method, syntheticBrokerBaseURL+target, body)
	if err != nil {
		return zero, fmt.Errorf("build synthetic broker request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	if requestBody != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if sendVersion {
		if version, ok := client.contractVersion.headerValue(); ok {
			request.Header.Set(HeaderAPIVersion, version)
		}
	}

	response, err := client.httpClient.Do(request)
	if err != nil {
		return zero, fmt.Errorf("send synthetic broker request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusUpgradeRequired {
		var refusal IncompatibleContractVersionErrorResponse
		if err := decodeResponse(response.Body, &refusal); err != nil {
			return zero, err
		}
		if err := refusal.Validate(); err != nil {
			return zero, err
		}
		return zero, &IncompatibleContractVersionError{
			StatusCode: response.StatusCode,
			Response:   refusal,
		}
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return zero, fmt.Errorf("%w: status %d", ErrUnexpectedResponse, response.StatusCode)
	}
	if err := decodeResponse(response.Body, &zero); err != nil {
		return zero, err
	}
	return zero, nil
}

func encodeRequestBody(requestBody any) (io.Reader, error) {
	if requestBody == nil {
		return nil, nil
	}
	payload, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("encode synthetic broker request: %w", err)
	}
	return bytes.NewReader(payload), nil
}

func decodeResponse(body io.Reader, target any) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode synthetic broker response: %w", err)
	}
	var trailing struct{}
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("%w: trailing JSON values", ErrUnexpectedResponse)
		}
		return fmt.Errorf("decode synthetic broker response trailer: %w", err)
	}
	return nil
}

// IncompatibleContractVersionError reports the exact HTTP 426 PM Broker refusal.
type IncompatibleContractVersionError struct {
	StatusCode int
	Response   IncompatibleContractVersionErrorResponse
}

// Error returns a client-safe version-refusal message.
func (err *IncompatibleContractVersionError) Error() string {
	if err == nil {
		return "contractv1: incompatible contract version"
	}
	return fmt.Sprintf("contractv1: %s", err.Response.Error.Code)
}

func (broker *SyntheticBroker) serveHTTP(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")

	if request.Method == http.MethodGet && request.URL.Path == "/v1/compatibility" {
		writeJSON(response, http.StatusOK, broker.fixtures.Compatibility)
		return
	}

	if request.Header.Get(HeaderAPIVersion) != string(ContractVersion1) {
		writeJSON(response, http.StatusUpgradeRequired, broker.fixtures.IncompatibleVersionError)
		return
	}

	switch {
	case request.Method == http.MethodGet && strings.HasPrefix(request.URL.Path, "/v1/connector-connections/"):
		broker.serveConnectorConnection(response, request)
	case request.Method == http.MethodPost && request.URL.Path == "/v1/execution-plans":
		broker.serveCreateExecutionPlan(response, request)
	default:
		writeJSON(response, http.StatusNotFound, IncompatibleContractVersionErrorResponse{
			Error: SafeError{
				Code:          ErrorCode("not_found"),
				Message:       "requested synthetic broker route was not found",
				CorrelationID: "corr_0123456789abcdef",
			},
		})
	}
}

func (broker *SyntheticBroker) serveConnectorConnection(response http.ResponseWriter, request *http.Request) {
	id := strings.TrimPrefix(request.URL.Path, "/v1/connector-connections/")
	if id != string(broker.fixtures.ConnectorConnection.ConnectorConnectionID) {
		writeJSON(response, http.StatusNotFound, IncompatibleContractVersionErrorResponse{
			Error: SafeError{
				Code:          ErrorCode("not_found"),
				Message:       "requested synthetic connector connection was not found",
				CorrelationID: "corr_0123456789abcdef",
			},
		})
		return
	}
	writeJSON(response, http.StatusOK, broker.fixtures.ConnectorConnection)
}

func (broker *SyntheticBroker) serveCreateExecutionPlan(response http.ResponseWriter, request *http.Request) {
	var planRequest ExecutionPlanRequest
	if err := decodeResponse(request.Body, &planRequest); err != nil {
		writeJSON(response, http.StatusBadRequest, IncompatibleContractVersionErrorResponse{
			Error: SafeError{
				Code:          ErrorCode("bad_request"),
				Message:       "execution plan request did not match the synthetic contract",
				CorrelationID: "corr_0123456789abcdef",
			},
		})
		return
	}
	if err := planRequest.Validate(); err != nil || !reflect.DeepEqual(planRequest, broker.fixtures.ExecutionPlanRequest) {
		writeJSON(response, http.StatusBadRequest, IncompatibleContractVersionErrorResponse{
			Error: SafeError{
				Code:          ErrorCode("bad_request"),
				Message:       "execution plan request did not match the synthetic contract",
				CorrelationID: "corr_0123456789abcdef",
			},
		})
		return
	}

	broker.mu.Lock()
	broker.executionPlanRequests = append(broker.executionPlanRequests, planRequest)
	broker.mu.Unlock()

	writeJSON(response, http.StatusOK, broker.fixtures.ExecutionPlan)
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.WriteHeader(status)
	if err := json.NewEncoder(response).Encode(value); err != nil {
		panic(fmt.Errorf("encode synthetic broker response: %w", err))
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}
