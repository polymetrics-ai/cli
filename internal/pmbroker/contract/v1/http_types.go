package contractv1

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"
)

const (
	// HeaderIdempotencyKey carries the mutation idempotency key over HTTP.
	HeaderIdempotencyKey = "Idempotency-Key"
	// HeaderCorrelationID carries the client-generated safe request correlation ID.
	HeaderCorrelationID = "X-Correlation-ID"
	// HeaderExecutionPlanDigest carries the immutable execution-plan digest in transport metadata.
	HeaderExecutionPlanDigest = "PM-Broker-Execution-Plan-Digest"
)

const (
	// DefaultPageLimit is the default bounded list limit for /v1 collection reads.
	DefaultPageLimit = 50
	// MaxPageLimit is the largest page size a PM Broker /v1 client may request.
	MaxPageLimit = 100
)

const (
	// ErrorCodeBadRequest reports a typed request that failed schema or semantic checks.
	ErrorCodeBadRequest ErrorCode = "bad_request"
	// ErrorCodeNotFound reports a missing typed PM Broker resource.
	ErrorCodeNotFound ErrorCode = "not_found"
	// ErrorCodeUnsafeRequest reports unsafe Host, Origin, or correlation metadata.
	ErrorCodeUnsafeRequest ErrorCode = "unsafe_request"
	// ErrorCodeAuthenticationRequired reports a typed request without explicit authorization.
	ErrorCodeAuthenticationRequired ErrorCode = "authentication_required"
	// ErrorCodeRateLimited reports a request refused by broker rate limiting.
	ErrorCodeRateLimited ErrorCode = "rate_limited"
)

var (
	// ErrInvalidEndpoint means the broker endpoint is not an HTTP(S) base URL safe for /v1.
	ErrInvalidEndpoint = errors.New("contractv1: invalid broker endpoint")
	// ErrAuthenticationRequired means a typed request needs explicit PM Broker authorization.
	ErrAuthenticationRequired = errors.New("contractv1: authentication required")
	// ErrAuthenticationFailed means the explicit PM Broker authorization seam failed safely.
	ErrAuthenticationFailed = errors.New("contractv1: authentication failed")
	// ErrInvalidPagination means a list request exceeded the bounded pagination contract.
	ErrInvalidPagination = errors.New("contractv1: invalid pagination")
	// ErrInvalidCorrelationID means a request correlation ID is absent or unsafe.
	ErrInvalidCorrelationID = errors.New("contractv1: invalid correlation id")
)

var (
	errorCodePattern           = regexp.MustCompile(`^[a-z][a-z0-9_]{2,63}$`)
	authorizationSchemePattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9._~-]{0,31}$`)
	pageCursorPattern          = regexp.MustCompile(`^cur_[A-Za-z0-9_-]{8,128}$`)
)

// Authorization is an opaque HTTP Authorization header value provided by an explicit auth seam.
// The credential value is write-only: the client can send it, but diagnostics never expose it.
type Authorization struct {
	scheme     string
	credential string
}

// NewAuthorization validates an explicit PM Broker Authorization header value.
func NewAuthorization(scheme string, credential string) (Authorization, error) {
	if !authorizationSchemePattern.MatchString(scheme) || credential == "" || hasUnsafeHeaderValue(credential) {
		return Authorization{}, ErrAuthenticationFailed
	}
	return Authorization{scheme: scheme, credential: credential}, nil
}

func (authorization Authorization) headerValue() (string, bool) {
	if authorization.scheme == "" || authorization.credential == "" {
		return "", false
	}
	if !authorizationSchemePattern.MatchString(authorization.scheme) || hasUnsafeHeaderValue(authorization.credential) {
		return "", false
	}
	return authorization.scheme + " " + authorization.credential, true
}

// Authenticator provides explicit request authorization without exposing generic headers.
type Authenticator interface {
	PMBrokerAuthorization(context.Context) (Authorization, error)
}

// AuthenticatorFunc adapts a function into an Authenticator.
type AuthenticatorFunc func(context.Context) (Authorization, error)

// PMBrokerAuthorization returns the authorization value for one PM Broker request.
func (fn AuthenticatorFunc) PMBrokerAuthorization(ctx context.Context) (Authorization, error) {
	if fn == nil {
		return Authorization{}, ErrAuthenticationFailed
	}
	return fn(ctx)
}

// CorrelationIDProvider provides safe request correlation IDs.
type CorrelationIDProvider interface {
	PMBrokerCorrelationID(context.Context) (CorrelationID, error)
}

// CorrelationIDProviderFunc adapts a function into a CorrelationIDProvider.
type CorrelationIDProviderFunc func(context.Context) (CorrelationID, error)

// PMBrokerCorrelationID returns the correlation ID for one PM Broker request.
func (fn CorrelationIDProviderFunc) PMBrokerCorrelationID(ctx context.Context) (CorrelationID, error) {
	if fn == nil {
		return "", ErrInvalidCorrelationID
	}
	return fn(ctx)
}

// Pagination is the bounded request shape for PM Broker /v1 list operations.
type Pagination struct {
	Limit  int        `json:"limit"`
	Cursor PageCursor `json:"cursor,omitempty"`
}

// PageCursor is an opaque, non-secret collection cursor.
type PageCursor string

// IsValid reports whether the cursor is empty or safely opaque.
func (cursor PageCursor) IsValid() bool {
	return cursor == "" || pageCursorPattern.MatchString(string(cursor))
}

// Validate reports whether pagination stays inside the bounded /v1 contract.
func (pagination Pagination) Validate() error {
	if pagination.Limit <= 0 || pagination.Limit > MaxPageLimit || !pagination.Cursor.IsValid() {
		return ErrInvalidPagination
	}
	return nil
}

func (pagination Pagination) normalized() Pagination {
	if pagination.Limit == 0 {
		pagination.Limit = DefaultPageLimit
	}
	return pagination
}

// ConnectorConnectionPage is the typed list response for connector connections.
type ConnectorConnectionPage struct {
	ConnectorConnections []ConnectorConnection `json:"connector_connections"`
	NextCursor           PageCursor            `json:"next_cursor,omitempty"`
}

// Validate reports whether the page preserves bounded, typed connector connection data.
func (page ConnectorConnectionPage) Validate() error {
	if len(page.ConnectorConnections) > MaxPageLimit || !page.NextCursor.IsValid() {
		return ErrInvalidPagination
	}
	for _, connection := range page.ConnectorConnections {
		if err := connection.Validate(); err != nil {
			return err
		}
	}
	return nil
}

// Validate reports whether compatibility discovery is safe and supports this client.
func (compatibility Compatibility) Validate() error {
	if compatibility.CurrentVersion != ContractVersion1 || compatibility.MinimumClientVersion != ContractVersion1 ||
		len(compatibility.SupportedVersions) == 0 {
		return ErrInvalidErrorResponse
	}
	compatible := false
	for _, version := range compatibility.SupportedVersions {
		if version == ContractVersion1 {
			compatible = true
		}
		if value, ok := version.headerValue(); !ok || value != string(version) {
			return ErrInvalidErrorResponse
		}
	}
	if !compatible {
		return ErrInvalidErrorResponse
	}
	return compatibility.IncompatibleVersionRefusal.ValidateIncompatibleVersion()
}

// ErrorResponse is the safe structured error envelope used by non-426 broker errors.
type ErrorResponse struct {
	Error SafeError `json:"error"`
}

// Validate reports whether the error response is safe to expose in CLI diagnostics.
func (response ErrorResponse) Validate() error {
	return response.Error.Validate()
}

// Validate reports whether the safe error uses stable low-cardinality fields.
func (safeError SafeError) Validate() error {
	if !errorCodePattern.MatchString(string(safeError.Code)) || !isSafeErrorMessage(safeError.Message) ||
		!safeError.CorrelationID.IsSafe() {
		return ErrInvalidErrorResponse
	}
	for _, version := range safeError.SupportedVersions {
		if value, ok := version.headerValue(); !ok || value != string(version) {
			return ErrInvalidErrorResponse
		}
	}
	return nil
}

func isSafeErrorMessage(message string) bool {
	if message == "" || hasUnsafeHeaderValue(message) {
		return false
	}
	lowerMessage := strings.ToLower(message)
	for _, marker := range unsafeDisplayHintMarkers {
		if strings.Contains(lowerMessage, marker) {
			return false
		}
	}
	return true
}

// RateLimit captures safe HTTP 429 retry metadata.
type RateLimit struct {
	RetryAfter time.Duration `json:"retry_after"`
	Limit      int           `json:"limit,omitempty"`
	Remaining  int           `json:"remaining,omitempty"`
}

// ClientDiagnostics is a redacted, non-secret snapshot of the configured transport.
type ClientDiagnostics struct {
	Endpoint        string          `json:"endpoint"`
	ContractVersion ContractVersion `json:"contract_version"`
	AuthConfigured  bool            `json:"auth_configured"`
	Transport       string          `json:"transport"`
}
