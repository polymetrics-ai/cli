package contractv1

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

const syntheticBrokerBaseURL = "http://pm-broker.synthetic.invalid"

// SyntheticBroker is a deterministic, network-free fake PM Broker.
type SyntheticBroker struct {
	mu                                   sync.Mutex
	fixtures                             SyntheticFixtures
	executionPlanRequests                []ExecutionPlanRequest
	observedRequests                     []ObservedRequest
	allowedHosts                         map[string]struct{}
	rateLimitNextMutation                bool
	rateLimitRetryAfter                  time.Duration
	corruptNextExecutionPlanDigestHeader bool
}

// ObservedRequest is a redacted summary of a synthetic broker request.
type ObservedRequest struct {
	Method                string
	Path                  string
	APIVersion            ContractVersion
	IdempotencyKey        IdempotencyKey
	CorrelationID         CorrelationID
	ExplicitAuthorization bool
	CookiePresent         bool
}

// NewSyntheticBroker returns a fake broker loaded with the accepted /v1 fixtures.
func NewSyntheticBroker() *SyntheticBroker {
	broker := &SyntheticBroker{
		fixtures:     AcceptedSyntheticFixtures(),
		allowedHosts: map[string]struct{}{},
	}
	if err := broker.AllowEndpoint(syntheticBrokerBaseURL); err != nil {
		panic(fmt.Errorf("allow synthetic broker endpoint: %w", err))
	}
	return broker
}

// NewClient returns a typed client wired to this in-memory fake broker.
func (broker *SyntheticBroker) NewClient(opts ...ClientOption) *Client {
	client, err := broker.NewHTTPClient(syntheticBrokerBaseURL, syntheticAuthorization{}, opts...)
	if err != nil {
		panic(fmt.Errorf("build synthetic broker client: %w", err))
	}
	return client
}

// NewHTTPClient returns a typed client using this fake broker's in-process transport.
func (broker *SyntheticBroker) NewHTTPClient(endpoint string, authenticator Authenticator, opts ...ClientOption) (*Client, error) {
	if err := broker.AllowEndpoint(endpoint); err != nil {
		return nil, err
	}
	options := make([]ClientOption, 0, len(opts)+1)
	options = append(options, withClientRoundTripper(broker.roundTripper()))
	options = append(options, opts...)
	return NewHTTPClient(endpoint, authenticator, options...)
}

func (broker *SyntheticBroker) roundTripper() http.RoundTripper {
	return roundTripFunc(func(request *http.Request) (*http.Response, error) {
		select {
		case <-request.Context().Done():
			return nil, request.Context().Err()
		default:
		}

		recorder := httptest.NewRecorder()
		broker.ServeHTTP(recorder, request)
		return recorder.Result(), nil
	})
}

// AllowEndpoint pins an endpoint host that this fake broker will accept.
func (broker *SyntheticBroker) AllowEndpoint(endpoint string) error {
	parsedEndpoint, err := parseBrokerEndpoint(endpoint)
	if err != nil {
		return err
	}
	broker.mu.Lock()
	defer broker.mu.Unlock()
	if broker.allowedHosts == nil {
		broker.allowedHosts = map[string]struct{}{}
	}
	broker.allowedHosts[parsedEndpoint.Host] = struct{}{}
	return nil
}

// ObservedRequests returns redacted request summaries captured by this fake broker.
func (broker *SyntheticBroker) ObservedRequests() []ObservedRequest {
	broker.mu.Lock()
	defer broker.mu.Unlock()
	return append([]ObservedRequest(nil), broker.observedRequests...)
}

func (broker *SyntheticBroker) isAllowedHost(host string) bool {
	broker.mu.Lock()
	defer broker.mu.Unlock()
	_, ok := broker.allowedHosts[host]
	return ok
}

func (broker *SyntheticBroker) recordObservedRequest(request *http.Request) {
	observed := ObservedRequest{
		Method:                request.Method,
		Path:                  request.URL.Path,
		APIVersion:            ContractVersion(request.Header.Get(HeaderAPIVersion)),
		IdempotencyKey:        IdempotencyKey(request.Header.Get(HeaderIdempotencyKey)),
		CorrelationID:         CorrelationID(request.Header.Get(HeaderCorrelationID)),
		ExplicitAuthorization: hasExplicitAuthorization(request),
		CookiePresent:         request.Header.Get("Cookie") != "",
	}
	broker.mu.Lock()
	defer broker.mu.Unlock()
	broker.observedRequests = append(broker.observedRequests, observed)
}

// ServeHTTP serves the synthetic PM Broker /v1 HTTP contract without network access.
func (broker *SyntheticBroker) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Content-Type", "application/json")
	broker.recordObservedRequest(request)

	if !isSafeHost(request.Host) || !broker.isAllowedHost(request.Host) {
		writeError(response, http.StatusBadRequest, ErrorCodeUnsafeRequest, "request host was not accepted", fallbackCorrelationID)
		return
	}
	if !isSafeOrigin(request.Header.Get("Origin"), request.Host) {
		writeError(response, http.StatusForbidden, ErrorCodeUnsafeRequest, "request origin was not accepted", fallbackCorrelationID)
		return
	}
	if correlationID := request.Header.Get(HeaderCorrelationID); correlationID != "" && !CorrelationID(correlationID).IsSafe() {
		writeError(response, http.StatusBadRequest, ErrorCodeUnsafeRequest, "request correlation id was not accepted", fallbackCorrelationID)
		return
	}
	if request.Header.Get("Cookie") != "" {
		writeError(response, http.StatusBadRequest, ErrorCodeUnsafeRequest, "ambient cookies are not accepted", requestCorrelationID(request))
		return
	}

	if request.Method == http.MethodGet && request.URL.Path == "/v1/compatibility" {
		writeJSON(response, http.StatusOK, broker.fixtures.Compatibility)
		return
	}

	if request.Header.Get(HeaderAPIVersion) != string(ContractVersion1) {
		writeJSON(response, http.StatusUpgradeRequired, broker.fixtures.IncompatibleVersionError)
		return
	}
	if !hasExplicitAuthorization(request) {
		writeError(response, http.StatusUnauthorized, ErrorCodeAuthenticationRequired, "explicit authorization is required", requestCorrelationID(request))
		return
	}
	response.Header().Set(HeaderCorrelationID, string(requestCorrelationID(request)))
	switch {
	case request.Method == http.MethodGet && request.URL.Path == "/v1/connector-connections":
		broker.serveListConnectorConnections(response, request)
	case request.Method == http.MethodGet && strings.HasPrefix(request.URL.Path, "/v1/connector-connections/"):
		broker.serveConnectorConnection(response, request)
	case request.Method == http.MethodPost && request.URL.Path == "/v1/execution-plans":
		broker.serveCreateExecutionPlan(response, request)
	default:
		writeError(response, http.StatusNotFound, ErrorCodeNotFound, "requested synthetic broker route was not found", requestCorrelationID(request))
	}
}

// ExecutionPlanRequestCount returns how many execution-plan requests were accepted.
func (broker *SyntheticBroker) ExecutionPlanRequestCount() int {
	broker.mu.Lock()
	defer broker.mu.Unlock()
	return len(broker.executionPlanRequests)
}

// RateLimitNextMutation makes the next synthetic mutation return HTTP 429.
func (broker *SyntheticBroker) RateLimitNextMutation(retryAfter time.Duration) {
	if retryAfter <= 0 {
		retryAfter = time.Second
	}
	broker.mu.Lock()
	defer broker.mu.Unlock()
	broker.rateLimitNextMutation = true
	broker.rateLimitRetryAfter = retryAfter
}

func (broker *SyntheticBroker) takeMutationRateLimit() (time.Duration, bool) {
	broker.mu.Lock()
	defer broker.mu.Unlock()
	if !broker.rateLimitNextMutation {
		return 0, false
	}
	broker.rateLimitNextMutation = false
	return broker.rateLimitRetryAfter, true
}

// CorruptNextExecutionPlanDigestHeader makes the next plan response carry a mismatched digest header.
func (broker *SyntheticBroker) CorruptNextExecutionPlanDigestHeader() {
	broker.mu.Lock()
	defer broker.mu.Unlock()
	broker.corruptNextExecutionPlanDigestHeader = true
}

func (broker *SyntheticBroker) executionPlanDigestHeader() string {
	broker.mu.Lock()
	defer broker.mu.Unlock()
	if !broker.corruptNextExecutionPlanDigestHeader {
		return string(broker.fixtures.ExecutionPlan.Digest)
	}
	broker.corruptNextExecutionPlanDigestHeader = false
	return "sha256:" + strings.Repeat("0", 64)
}

// Client is the typed PM Broker /v1 HTTP/JSON client.
type Client struct {
	contractVersion       ContractVersion
	endpoint              *url.URL
	httpClient            *http.Client
	authenticator         Authenticator
	correlationIDProvider CorrelationIDProvider
}

// ClientOption configures the typed client without exposing arbitrary request fields.
type ClientOption func(*clientConfig)

type clientConfig struct {
	contractVersion       ContractVersion
	roundTripper          http.RoundTripper
	correlationIDProvider CorrelationIDProvider
}

// NewHTTPClient builds a typed PM Broker /v1 HTTP/JSON client for loopback or remote/container endpoints.
func NewHTTPClient(endpoint string, authenticator Authenticator, opts ...ClientOption) (*Client, error) {
	parsedEndpoint, err := parseBrokerEndpoint(endpoint)
	if err != nil {
		return nil, err
	}

	config := clientConfig{
		contractVersion:       ContractVersion1,
		roundTripper:          http.DefaultTransport,
		correlationIDProvider: CorrelationIDProviderFunc(newCorrelationID),
	}
	for _, opt := range opts {
		opt(&config)
	}
	if config.roundTripper == nil || config.correlationIDProvider == nil {
		return nil, ErrInvalidEndpoint
	}

	return &Client{
		contractVersion: config.contractVersion,
		endpoint:        parsedEndpoint,
		httpClient: &http.Client{
			Transport: config.roundTripper,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		authenticator:         authenticator,
		correlationIDProvider: config.correlationIDProvider,
	}, nil
}

// WithClientContractVersion sets the version sent on typed /v1 operations.
func WithClientContractVersion(version ContractVersion) ClientOption {
	return func(config *clientConfig) {
		config.contractVersion = version
	}
}

func withClientRoundTripper(roundTripper http.RoundTripper) ClientOption {
	return func(config *clientConfig) {
		config.roundTripper = roundTripper
	}
}

// WithClientCorrelationIDProvider sets the request correlation ID provider.
func WithClientCorrelationIDProvider(provider CorrelationIDProvider) ClientOption {
	return func(config *clientConfig) {
		config.correlationIDProvider = provider
	}
}

// Diagnostics returns redacted client transport diagnostics.
func (client *Client) Diagnostics() ClientDiagnostics {
	if client == nil {
		return ClientDiagnostics{Transport: "http-json-v1"}
	}
	endpoint := ""
	if client.endpoint != nil {
		endpoint = redactedEndpoint(client.endpoint)
	}
	contractVersion := client.contractVersion
	if value, ok := contractVersion.headerValue(); !ok || value != string(contractVersion) {
		contractVersion = ContractVersion("invalid")
	}
	return ClientDiagnostics{
		Endpoint:        endpoint,
		ContractVersion: contractVersion,
		AuthConfigured:  client.authenticator != nil,
		Transport:       "http-json-v1",
	}
}

// Compatibility returns /v1 compatibility information. This method does not send
// PM-Broker-API-Version because the accepted contract exempts compatibility discovery.
func (client *Client) Compatibility(ctx context.Context) (Compatibility, error) {
	return doJSON[Compatibility](ctx, client, http.MethodGet, "/v1/compatibility", nil, nil, requestOptions{})
}

// NegotiateCompatibility returns compatibility information after validating this client's version.
func (client *Client) NegotiateCompatibility(ctx context.Context) (Compatibility, error) {
	compatibility, err := client.Compatibility(ctx)
	if err != nil {
		return Compatibility{}, err
	}
	if err := compatibility.Validate(); err != nil {
		return Compatibility{}, err
	}
	return compatibility, nil
}

// ListConnectorConnections returns a bounded page of connector connections.
func (client *Client) ListConnectorConnections(ctx context.Context, pagination Pagination) (ConnectorConnectionPage, error) {
	pagination = pagination.normalized()
	if err := pagination.Validate(); err != nil {
		return ConnectorConnectionPage{}, err
	}
	query := url.Values{}
	query.Set("limit", strconv.Itoa(pagination.Limit))
	if pagination.Cursor != "" {
		query.Set("cursor", string(pagination.Cursor))
	}
	page, err := doJSON[ConnectorConnectionPage](ctx, client, http.MethodGet, "/v1/connector-connections", query, nil, requestOptions{
		sendVersion: true,
		requireAuth: true,
	})
	if err != nil {
		return ConnectorConnectionPage{}, err
	}
	if err := page.Validate(); err != nil {
		return ConnectorConnectionPage{}, err
	}
	return page, nil
}

// ConnectorConnection returns a typed connector connection by ID.
func (client *Client) ConnectorConnection(ctx context.Context, id ConnectorConnectionID) (ConnectorConnection, error) {
	if !id.IsValid() {
		return ConnectorConnection{}, ErrInvalidIdentityBoundary
	}
	connection, err := doJSON[ConnectorConnection](ctx, client, http.MethodGet, "/v1/connector-connections/"+string(id), nil, nil, requestOptions{
		sendVersion: true,
		requireAuth: true,
	})
	if err != nil {
		return ConnectorConnection{}, err
	}
	if err := connection.Validate(); err != nil {
		return ConnectorConnection{}, err
	}
	return connection, nil
}

// CreateExecutionPlan sends a typed execution-plan request to the configured broker endpoint.
func (client *Client) CreateExecutionPlan(ctx context.Context, request ExecutionPlanRequest) (ExecutionPlan, error) {
	if err := request.Validate(); err != nil {
		return ExecutionPlan{}, err
	}
	plan, err := doJSON[ExecutionPlan](ctx, client, http.MethodPost, "/v1/execution-plans", nil, request, requestOptions{
		sendVersion:               true,
		requireAuth:               true,
		mutation:                  true,
		idempotencyKey:            request.IdempotencyKey,
		expectExecutionPlanDigest: true,
	})
	if err != nil {
		return ExecutionPlan{}, err
	}
	if err := plan.Validate(); err != nil {
		return ExecutionPlan{}, err
	}
	return plan, nil
}

type requestOptions struct {
	sendVersion               bool
	requireAuth               bool
	mutation                  bool
	idempotencyKey            IdempotencyKey
	expectExecutionPlanDigest bool
}

func doJSON[T any](ctx context.Context, client *Client, method string, target string, query url.Values, requestBody any, options requestOptions) (T, error) {
	var zero T
	body, err := encodeRequestBody(requestBody)
	if err != nil {
		return zero, err
	}
	targetURL := client.resolve(target, query)
	request, err := http.NewRequestWithContext(ctx, method, targetURL, body)
	if err != nil {
		return zero, fmt.Errorf("build broker request: %w", err)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Del("Cookie")
	if requestBody != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if options.sendVersion {
		if version, ok := client.contractVersion.headerValue(); ok {
			request.Header.Set(HeaderAPIVersion, version)
		}
	}
	correlationID, err := client.correlationID(ctx)
	if err != nil {
		return zero, err
	}
	request.Header.Set(HeaderCorrelationID, string(correlationID))
	if options.requireAuth {
		if err := client.applyAuthorization(ctx, request); err != nil {
			return zero, err
		}
	}
	if options.mutation {
		request.Header.Del("Cookie")
		if !options.idempotencyKey.IsValid() {
			return zero, ErrInvalidExecutionPlan
		}
		request.Header.Set(HeaderIdempotencyKey, string(options.idempotencyKey))
	}

	response, err := client.httpClient.Do(request)
	if err != nil {
		return zero, fmt.Errorf("send broker request: %w", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

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
		brokerError, err := decodeBrokerError(response)
		if err != nil {
			return zero, err
		}
		return zero, brokerError
	}

	var value T
	if err := decodeResponse(response.Body, &value); err != nil {
		return zero, err
	}
	if options.expectExecutionPlanDigest {
		if err := validateExecutionPlanDigestHeader(response.Header, value); err != nil {
			return zero, err
		}
	}
	return value, nil
}

func (client *Client) resolve(target string, query url.Values) string {
	resolved := *client.endpoint
	basePath := strings.TrimRight(resolved.Path, "/")
	if basePath == "" {
		resolved.Path = target
	} else {
		resolved.Path = basePath + target
	}
	resolved.RawQuery = query.Encode()
	resolved.Fragment = ""
	resolved.User = nil
	return resolved.String()
}

func (client *Client) correlationID(ctx context.Context) (CorrelationID, error) {
	if client == nil || client.correlationIDProvider == nil {
		return "", ErrInvalidCorrelationID
	}
	correlationID, err := client.correlationIDProvider.PMBrokerCorrelationID(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "", err
		}
		return "", ErrInvalidCorrelationID
	}
	if !correlationID.IsSafe() {
		return "", ErrInvalidCorrelationID
	}
	return correlationID, nil
}

func (client *Client) applyAuthorization(ctx context.Context, request *http.Request) error {
	if client == nil || client.authenticator == nil {
		return ErrAuthenticationRequired
	}
	authorization, err := client.authenticator.PMBrokerAuthorization(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return err
		}
		return ErrAuthenticationFailed
	}
	value, ok := authorization.headerValue()
	if !ok {
		return ErrAuthenticationFailed
	}
	request.Header.Set("Authorization", value)
	return nil
}

func encodeRequestBody(requestBody any) (io.Reader, error) {
	if requestBody == nil {
		return nil, nil
	}
	payload, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("encode broker request: %w", err)
	}
	return bytes.NewReader(payload), nil
}

func decodeResponse(body io.Reader, target any) error {
	decoder := json.NewDecoder(body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decode broker response: %w", err)
	}
	var trailing struct{}
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return fmt.Errorf("%w: trailing JSON values", ErrUnexpectedResponse)
		}
		return fmt.Errorf("decode broker response trailer: %w", err)
	}
	return nil
}

func decodeBrokerError(response *http.Response) (*BrokerError, error) {
	var errorResponse ErrorResponse
	if err := decodeResponse(response.Body, &errorResponse); err != nil {
		return nil, err
	}
	if err := errorResponse.Validate(); err != nil {
		return nil, err
	}
	brokerError := &BrokerError{
		StatusCode: response.StatusCode,
		Response:   errorResponse,
	}
	if response.StatusCode == http.StatusTooManyRequests {
		brokerError.RateLimit = parseRateLimit(response.Header)
	}
	return brokerError, nil
}

func validateExecutionPlanDigestHeader(header http.Header, value any) error {
	plan, ok := any(value).(ExecutionPlan)
	if !ok {
		return ErrInvalidExecutionPlan
	}
	if header.Get(HeaderExecutionPlanDigest) != string(plan.Digest) {
		return ErrInvalidExecutionPlan
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

// BrokerError reports a safe structured non-426 PM Broker error.
type BrokerError struct {
	StatusCode int
	Response   ErrorResponse
	RateLimit  *RateLimit
}

// Error returns a low-cardinality, redacted broker error string.
func (err *BrokerError) Error() string {
	if err == nil {
		return "contractv1: broker error"
	}
	return fmt.Sprintf("contractv1: broker error %s status %d correlation %s", err.Response.Error.Code, err.StatusCode, err.Response.Error.CorrelationID)
}

func (broker *SyntheticBroker) serveListConnectorConnections(response http.ResponseWriter, request *http.Request) {
	pagination, err := paginationFromQuery(request.URL.Query())
	if err != nil {
		writeError(response, http.StatusBadRequest, ErrorCodeBadRequest, "connector connection pagination was not accepted", requestCorrelationID(request))
		return
	}
	connections := []ConnectorConnection{broker.fixtures.ConnectorConnection}
	if pagination.Limit < len(connections) {
		connections = connections[:pagination.Limit]
	}
	writeJSON(response, http.StatusOK, ConnectorConnectionPage{ConnectorConnections: connections})
}

func (broker *SyntheticBroker) serveConnectorConnection(response http.ResponseWriter, request *http.Request) {
	id := strings.TrimPrefix(request.URL.Path, "/v1/connector-connections/")
	if id != string(broker.fixtures.ConnectorConnection.ConnectorConnectionID) {
		writeError(response, http.StatusNotFound, ErrorCodeNotFound, "requested synthetic connector connection was not found", requestCorrelationID(request))
		return
	}
	writeJSON(response, http.StatusOK, broker.fixtures.ConnectorConnection)
}

func (broker *SyntheticBroker) serveCreateExecutionPlan(response http.ResponseWriter, request *http.Request) {
	var planRequest ExecutionPlanRequest
	if err := decodeResponse(request.Body, &planRequest); err != nil {
		writeError(response, http.StatusBadRequest, ErrorCodeBadRequest, "execution plan request did not match the synthetic contract", requestCorrelationID(request))
		return
	}
	if err := planRequest.Validate(); err != nil || !reflect.DeepEqual(planRequest, broker.fixtures.ExecutionPlanRequest) {
		writeError(response, http.StatusBadRequest, ErrorCodeBadRequest, "execution plan request did not match the synthetic contract", requestCorrelationID(request))
		return
	}
	if request.Header.Get(HeaderIdempotencyKey) != string(planRequest.IdempotencyKey) {
		writeError(response, http.StatusBadRequest, ErrorCodeBadRequest, "execution plan idempotency key was not accepted", requestCorrelationID(request))
		return
	}
	if retryAfter, ok := broker.takeMutationRateLimit(); ok {
		seconds := int(retryAfter.Round(time.Second).Seconds())
		if seconds <= 0 {
			seconds = 1
		}
		response.Header().Set("Retry-After", strconv.Itoa(seconds))
		writeError(response, http.StatusTooManyRequests, ErrorCodeRateLimited, "request rate limit exceeded", requestCorrelationID(request))
		return
	}

	broker.mu.Lock()
	broker.executionPlanRequests = append(broker.executionPlanRequests, planRequest)
	broker.mu.Unlock()

	response.Header().Set(HeaderExecutionPlanDigest, broker.executionPlanDigestHeader())
	writeJSON(response, http.StatusOK, broker.fixtures.ExecutionPlan)
}

func writeJSON(response http.ResponseWriter, status int, value any) {
	response.WriteHeader(status)
	if err := json.NewEncoder(response).Encode(value); err != nil {
		panic(fmt.Errorf("encode synthetic broker response: %w", err))
	}
}

func writeError(response http.ResponseWriter, status int, code ErrorCode, message string, correlationID CorrelationID) {
	if !correlationID.IsSafe() {
		correlationID = fallbackCorrelationID
	}
	writeJSON(response, status, ErrorResponse{
		Error: SafeError{
			Code:          code,
			Message:       message,
			CorrelationID: correlationID,
		},
	})
}

func paginationFromQuery(query url.Values) (Pagination, error) {
	for key := range query {
		if key != "limit" && key != "cursor" {
			return Pagination{}, ErrInvalidPagination
		}
	}
	limit := DefaultPageLimit
	if rawLimit := query.Get("limit"); rawLimit != "" {
		parsedLimit, err := strconv.Atoi(rawLimit)
		if err != nil {
			return Pagination{}, ErrInvalidPagination
		}
		limit = parsedLimit
	}
	pagination := Pagination{Limit: limit, Cursor: PageCursor(query.Get("cursor"))}
	if err := pagination.Validate(); err != nil {
		return Pagination{}, err
	}
	return pagination, nil
}

func parseBrokerEndpoint(endpoint string) (*url.URL, error) {
	if endpoint == "" || strings.TrimSpace(endpoint) != endpoint || hasUnsafeHeaderValue(endpoint) {
		return nil, ErrInvalidEndpoint
	}
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return nil, ErrInvalidEndpoint
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return nil, ErrInvalidEndpoint
	}
	if parsed.Host == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || !isSafeHost(parsed.Host) {
		return nil, ErrInvalidEndpoint
	}
	if parsed.Scheme == "http" && !isAllowedPlainHTTPHost(parsed.Host) {
		return nil, ErrInvalidEndpoint
	}
	if parsed.RawPath != "" || strings.Contains(parsed.Path, "\\") || hasUnsafeHeaderValue(parsed.Path) {
		return nil, ErrInvalidEndpoint
	}
	if parsed.Path != "" && parsed.Path != "/" && parsed.Path != "." {
		return nil, ErrInvalidEndpoint
	}
	parsed.Path = ""
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed, nil
}

func redactedEndpoint(endpoint *url.URL) string {
	if endpoint == nil {
		return ""
	}
	redacted := *endpoint
	redacted.User = nil
	redacted.RawQuery = ""
	redacted.Fragment = ""
	return redacted.String()
}

func parseRateLimit(header http.Header) *RateLimit {
	rateLimit := &RateLimit{}
	if retryAfter := header.Get("Retry-After"); retryAfter != "" {
		if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds > 0 {
			rateLimit.RetryAfter = time.Duration(seconds) * time.Second
		}
	}
	if limit := header.Get("RateLimit-Limit"); limit != "" {
		if parsed, err := strconv.Atoi(limit); err == nil && parsed >= 0 {
			rateLimit.Limit = parsed
		}
	}
	if remaining := header.Get("RateLimit-Remaining"); remaining != "" {
		if parsed, err := strconv.Atoi(remaining); err == nil && parsed >= 0 {
			rateLimit.Remaining = parsed
		}
	}
	return rateLimit
}

func newCorrelationID(context.Context) (CorrelationID, error) {
	var randomBytes [16]byte
	if _, err := rand.Read(randomBytes[:]); err != nil {
		return "", ErrInvalidCorrelationID
	}
	return CorrelationID("corr_" + hex.EncodeToString(randomBytes[:])), nil
}

const fallbackCorrelationID CorrelationID = "corr_0123456789abcdef"

func requestCorrelationID(request *http.Request) CorrelationID {
	if request == nil {
		return fallbackCorrelationID
	}
	correlationID := CorrelationID(request.Header.Get(HeaderCorrelationID))
	if !correlationID.IsSafe() {
		return fallbackCorrelationID
	}
	return correlationID
}

func hasExplicitAuthorization(request *http.Request) bool {
	if request == nil {
		return false
	}
	value := request.Header.Get("Authorization")
	return value != "" && !hasUnsafeHeaderValue(value)
}

func isSafeHost(host string) bool {
	if host == "" || hasUnsafeHeaderValue(host) || strings.ContainsAny(host, "/\\@ ") {
		return false
	}
	return true
}

func isAllowedPlainHTTPHost(host string) bool {
	hostname := hostnameForPolicy(host)
	return hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" ||
		hostname == "pm-broker.synthetic.invalid" || strings.HasSuffix(hostname, ".internal") ||
		!strings.Contains(hostname, ".")
}

func hostnameForPolicy(host string) string {
	if hostname, _, err := net.SplitHostPort(host); err == nil {
		host = hostname
	}
	return strings.Trim(strings.ToLower(host), "[]")
}

func isSafeOrigin(origin string, requestHost string) bool {
	if origin == "" {
		return true
	}
	if hasUnsafeHeaderValue(origin) || !isSafeHost(requestHost) {
		return false
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return false
	}
	if parsed.User != nil || parsed.Host == "" || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" || !isSafeHost(parsed.Host) {
		return false
	}
	return strings.EqualFold(parsed.Host, requestHost)
}

func hasUnsafeHeaderValue(value string) bool {
	for _, char := range value {
		if char < 0x20 || char == 0x7f {
			return true
		}
	}
	return false
}

type syntheticAuthorization struct{}

func (syntheticAuthorization) PMBrokerAuthorization(context.Context) (Authorization, error) {
	return NewAuthorization("PMBroker", strings.Repeat("x", 24))
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}
