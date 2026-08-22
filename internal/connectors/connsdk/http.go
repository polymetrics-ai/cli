package connsdk

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors/transportpolicy"
	"polymetrics.ai/internal/safety"
)

// maxErrorBody bounds how much of an error response body is captured in HTTPError.
const maxErrorBody = 8 << 10 // 8 KiB
const DefaultMaxResponseBody = 64 << 20
const maxRedirects = 10

// Response is a captured HTTP response with its body already read.
type Response struct {
	Status         int
	Header         http.Header
	Body           []byte
	requestURL     string
	rateLimitRoute RateLimitRoute
}

type StatusRange struct {
	Min int
	Max int
}

type UnexpectedStatusError struct {
	Status int
}

// ProviderResponseError is the printable-safe typed identity of a non-success
// provider response. The private HTTPError remains inside the engine for rate
// and authentication classification; callers receive status without raw URL,
// headers, or body.
type ProviderResponseError struct {
	Status int
}

func (e *ProviderResponseError) Error() string {
	return fmt.Sprintf("provider response status %d", e.Status)
}

func (e *UnexpectedStatusError) Error() string {
	return fmt.Sprintf("successful response status %d is not declared", e.Status)
}

// HTTPError is returned when a request completes with a 4xx/5xx status after
// exhausting retries. Its body and headers are never assumed to be secret-free
// by callers, but connsdk itself never logs them.
type HTTPError struct {
	Status int
	URL    string
	Header http.Header
	Body   string
	// RawBody preserves captured response bytes for typed receipt consumers.
	// Error() never renders it.
	RawBody []byte
}

func (e *HTTPError) Error() string {
	msg := strings.TrimSpace(e.Body)
	if msg == "" {
		msg = http.StatusText(e.Status)
	}
	if len(msg) > 512 {
		msg = msg[:512] + "..."
	}
	return safety.RedactErrorText(fmt.Sprintf("http %d for %s: %s", e.Status, e.URL, msg))
}

// CredentialRejectedError is the safe identity of a provider-verified
// authentication rejection. It carries neither the provider URL nor response
// body, either of which may contain credential material. Response formatters
// may preserve this type without exposing the raw HTTPError that proved it.
type CredentialRejectedError struct{}

func (*CredentialRejectedError) Error() string {
	return "provider rejected the credential"
}

// Requester performs JSON HTTP requests with auth, retry, and rate-limit handling.
// The zero value is usable once Client/BaseURL are set; sensible defaults are
// applied for the rest on first use.
type MultipartForm struct {
	Fields   map[string]string
	Files    []MultipartFile
	MaxBytes int64
}

type MultipartFile struct {
	FieldName string
	Path      string
	// Root, when set, confines every access to this file beneath it, and
	// RelPath is the root-relative path opened through it. Containment lives at
	// the open rather than in a check performed once beforehand: Root is
	// consulted on every Stat and Open, so a path swapped for an escaping
	// symlink after validation is refused instead of followed. Path is then
	// only a display value and is never opened.
	Root    *os.Root
	RelPath string

	FileName string
	// ContentType is the part header the bundle declares. When AllowedMediaTypes
	// bounds the part, the sent header is replaced by the type the bytes actually
	// sniffed as, so ContentType is then the authoring intent rather than the
	// wire value.
	ContentType string
	// AllowedMediaTypes, when non-empty, bounds what the file's own bytes may
	// sniff as, and is the only restriction on the part's type — a single-entry
	// list is how a bundle demands exactly one type. It is enforced before any
	// request is made, and the sniffed type it admits becomes the header we send,
	// so the claim made to the provider is one we have verified rather than one
	// we merely declared.
	AllowedMediaTypes []string
	MaxBytes          int64
	ExpectedSHA256    string
}

// sourceName is the path used in messages and as the default upload filename.
func (f MultipartFile) sourceName() string {
	if f.Root != nil {
		return f.RelPath
	}
	return f.Path
}

// stat resolves the file's metadata under Root when one is set.
func (f MultipartFile) stat() (os.FileInfo, error) {
	if f.Root != nil {
		return f.Root.Stat(f.RelPath)
	}
	return os.Stat(f.Path)
}

// open opens the file under Root when one is set. os.Root refuses any path that
// escapes the root, including via a symlink swapped in after an earlier check.
func (f MultipartFile) open() (*os.File, error) {
	if f.Root != nil {
		return f.Root.Open(f.RelPath)
	}
	return os.Open(f.Path)
}

// needsSnapshot reports whether the file must be copied to a bounded temp file
// before it is sent: either its content is bound to an approved digest, or its
// media type is bounded and has to be checked against the actual bytes.
func (f MultipartFile) needsSnapshot() bool {
	return f.ExpectedSHA256 != "" || len(f.AllowedMediaTypes) > 0
}

type requestBody struct {
	Reader      io.Reader
	ContentType string
	Cleanup     func() error
}

type Requester struct {
	// Client is the HTTP client. Defaults to a client with a 60s timeout.
	Client *http.Client
	// BaseURL is prepended to relative paths. A path beginning with http:// or
	// https:// is treated as absolute and used as-is (e.g. Link-header next URLs).
	BaseURL string
	// Auth, when set, is applied to every request before it is sent.
	Auth Authenticator
	// UserAgent and DefaultHeaders are applied to every request.
	UserAgent           string
	DefaultHeaders      map[string]string
	DefaultHeaderValues http.Header
	// Accept overrides the Accept header (defaults to application/json).
	Accept           string
	AcceptedStatuses []StatusRange
	RedirectPolicy   *RedirectPolicy

	// MaxRetries is the number of additional attempts after the first (default 4).
	MaxRetries int
	// DisableRetries disables Requester-managed transient-status, transport-error,
	// and 401 reauthentication retries. Strict non-idempotent writes set it. With
	// the standard net/http transport, those writes use a one-use HTTP/1
	// connection so the transport cannot replay them after a reused-connection
	// failure. Safe replayable reads can still be replayed inside net/http.
	DisableRetries bool
	// BaseBackoff and MaxBackoff bound exponential backoff (defaults 500ms / 30s).
	BaseBackoff time.Duration
	MaxBackoff  time.Duration
	// RetryStatuses are HTTP statuses that trigger a retry. Defaults to
	// 429, 500, 502, 503, 504.
	RetryStatuses map[int]bool
	// Sleep waits for d or until ctx is cancelled. Injectable for tests.
	Sleep func(ctx context.Context, d time.Duration) error
	// Now supplies the clock used to interpret provider reset headers.
	// Injectable tests can therefore prove the exact reset timestamp.
	Now func() time.Time
	// Jitter supplies full jitter for fallback exponential retries. It is never
	// called for a valid provider Retry-After reset. Returned values are clamped
	// to the fallback cap so an implementation cannot lengthen a retry.
	Jitter func(cap time.Duration) time.Duration
	// Admission runs immediately before each logical Requester send and permitted
	// redirect hop. It must honor the request context; an error prevents that
	// attempt from reaching the provider. A safe replayable read can be replayed
	// inside net/http without another admission. The engine attaches a
	// declaration-aware implementation where a policy matches.
	Admission RateLimitAdmission
	// Observer receives parsed response rate-limit facts synchronously. It is
	// deliberately not an output hook; #3755 owns operator-visible events.
	Observer RateLimitObserver
	// RateLimitCostHeader is the one provider header named by the selected
	// declaration that reports a request's actual point cost. It is an HTTP
	// field name validated by the bundle loader, never a caller-provided value.
	// The parsed scalar is delivered through RateLimitObservation; raw headers
	// are never retained.
	RateLimitCostHeader string
	RouteRateLimits     RateLimitRouteResolver
	// RateLimitEvents records bounded admission/observation transitions for a
	// caller that needs an audit trail (for example, certification). It never
	// receives raw provider data and cannot influence request control flow.
	RateLimitEvents RateLimitEventSink
	// RateLimitAdmissionTimeout bounds one admission wait without shortening
	// the surrounding request or certification run. A deadline failure is
	// emitted as a not_sent event and the provider is never contacted.
	RateLimitAdmissionTimeout time.Duration
}

func (r *Requester) client() *http.Client {
	if r.Client != nil {
		return r.Client
	}
	return &http.Client{Timeout: 60 * time.Second}
}

func (r *Requester) clientFor(ctx context.Context) *http.Client {
	return transportpolicy.HTTPClientRetainingRedirectResponse(ctx, r.client())
}

func isSafeReplayableRead(method string) bool {
	switch strings.ToUpper(strings.TrimSpace(method)) {
	case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		return true
	default:
		return false
	}
}

func noReplayClient(client *http.Client) *http.Client {
	clone := *client
	clone.CheckRedirect = func(*http.Request, []*http.Request) error {
		return transportpolicy.ErrRedirectRefused
	}
	transport, ok := client.Transport.(*http.Transport)
	if client.Transport == nil {
		transport, ok = http.DefaultTransport.(*http.Transport)
	}
	if !ok {
		return &clone
	}
	strictTransport := transport.Clone()
	strictTransport.DisableKeepAlives = true
	clone.Transport = strictTransport
	return &clone
}

func noReplayResponseClient(client *http.Client) *http.Client {
	strict := noReplayClient(client)
	strict.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return strict
}

func noRedirectResponseClient(client *http.Client) *http.Client {
	clone := *client
	clone.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &clone
}

func disableTransportReplay(req *http.Request, strictWrite bool) {
	req.GetBody = nil
	req.Header.Del("Idempotency-Key")
	req.Header.Del("X-Idempotency-Key")
	if strictWrite {
		req.Close = true
	}
}

func (r *Requester) maxRetries() int {
	if r.DisableRetries {
		return 0
	}
	if r.MaxRetries > 0 {
		return r.MaxRetries
	}
	return 4
}

func (r *Requester) baseBackoff() time.Duration {
	if r.BaseBackoff > 0 {
		return r.BaseBackoff
	}
	return 500 * time.Millisecond
}

func (r *Requester) maxBackoff() time.Duration {
	if r.MaxBackoff > 0 {
		return r.MaxBackoff
	}
	return 30 * time.Second
}

func (r *Requester) shouldRetry(status int) bool {
	if r.RetryStatuses != nil {
		return r.RetryStatuses[status]
	}
	switch status {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	default:
		return false
	}
}

func (r *Requester) sleep(ctx context.Context, d time.Duration) error {
	if r.Sleep != nil {
		return r.Sleep(ctx, d)
	}
	return ctxSleep(ctx, d)
}

func (r *Requester) now() time.Time {
	if r.Now != nil {
		return r.Now()
	}
	return time.Now()
}

func (r *Requester) admit(ctx context.Context, route RateLimitRoute) (string, error) {
	costHeader := r.RateLimitCostHeader
	if r.Admission != nil {
		if err := r.Admission.Admit(ctx, RateLimitRequest{Method: route.Method, Attempt: route.Attempt}); err != nil {
			return "", err
		}
	}
	if r.RouteRateLimits == nil {
		return costHeader, nil
	}
	routeCostHeader, err := r.RouteRateLimits.AdmitRoute(ctx, route)
	if err != nil {
		return "", err
	}
	if routeCostHeader != "" {
		if costHeader != "" && !strings.EqualFold(costHeader, routeCostHeader) {
			return "", fmt.Errorf("matching rate-limit policies use different actual-cost headers")
		}
		costHeader = routeCostHeader
	}
	return costHeader, nil
}

func (r *Requester) observeRateLimit(ctx context.Context, route RateLimitRoute, status int, header http.Header, costHeader string) RateLimitObservation {
	observation, ok := rateLimitObservation(status, header, route.Attempt, r.now(), costHeader)
	if ok {
		r.emitRateLimitReset(route.Method, observation)
		if r.Observer != nil {
			r.Observer.Observe(ctx, observation)
		}
		if r.RouteRateLimits != nil {
			r.RouteRateLimits.ObserveRoute(ctx, route, observation)
		}
	}
	return observation
}

// ObserveRateLimit records a bounded, declaration-approved response fact
// discovered after the requester's standard header observation. It preserves
// the actual admitted route and attempt from Response, so a caller cannot
// attach an observation to an unsent request or invent a different route.
//
// This seam is intentionally narrow: fixed GraphQL operation parsing may pass
// its rateLimit selection, but arbitrary response-body extraction remains
// unavailable to connector callers.
func (r *Requester) ObserveRateLimit(ctx context.Context, response *Response, observation RateLimitObservation) {
	if r == nil || response == nil || response.rateLimitRoute.Attempt <= 0 {
		return
	}
	observation.Status = response.Status
	observation.Attempt = response.rateLimitRoute.Attempt
	observation.Attempted = true
	if observation.ObservedAt.IsZero() {
		observation.ObservedAt = r.now()
	}
	r.emitRateLimitReset(response.rateLimitRoute.Method, observation)
	if r.Observer != nil {
		r.Observer.Observe(ctx, observation)
	}
	if r.RouteRateLimits != nil {
		r.RouteRateLimits.ObserveRoute(ctx, response.rateLimitRoute, observation)
	}
}

func (r *Requester) emitRateLimitEvent(event RateLimitEvent) {
	if r == nil || r.RateLimitEvents == nil {
		return
	}
	r.RateLimitEvents.RecordRateLimitEvent(event)
}

func (r *Requester) emitRateLimitReset(method string, observation RateLimitObservation) {
	if !observation.HasReset {
		return
	}
	r.emitRateLimitEvent(RateLimitEvent{
		Type:    RateLimitEventReset,
		Method:  method,
		Attempt: observation.Attempt,
		ResetAt: observation.ResetAt,
	})
}

type rateLimitAdmissionError struct {
	err error
}

func (e *rateLimitAdmissionError) Error() string {
	return fmt.Sprintf("rate-limit admission: %v", e.err)
}

func (e *rateLimitAdmissionError) Unwrap() error {
	return e.err
}

// requestAdmissionError marks a durable credential fence separately from a
// rate-limit admission refusal. It is terminal for the current operation: a
// retry would only poll the same fenced epoch and must never buy another send.
type requestAdmissionError struct {
	err error
}

func (e *requestAdmissionError) Error() string {
	return fmt.Sprintf("request admission: %v", e.err)
}

func (e *requestAdmissionError) Unwrap() error {
	return e.err
}

const minimumObservableRateLimitWait = time.Millisecond

func (r *Requester) admitRequesterSend(ctx context.Context, req *http.Request, requesterAttempt *int, route *RateLimitRoute, costHeader *string) error {
	if err := CheckRequestAdmission(ctx); err != nil {
		return &requestAdmissionError{err: err}
	}
	nextAttempt := *requesterAttempt + 1
	nextRoute := RateLimitRoute{Method: req.Method, Path: r.rateLimitRoutePath(req.URL), Attempt: nextAttempt}
	admissionCtx, cancelAdmission := r.rateLimitAdmissionContext(ctx)
	defer cancelAdmission()
	started := time.Now()
	header, err := r.admit(admissionCtx, nextRoute)
	if elapsed := time.Since(started); elapsed >= minimumObservableRateLimitWait {
		r.emitRateLimitEvent(RateLimitEvent{
			Type:       RateLimitEventWait,
			Method:     nextRoute.Method,
			Attempt:    nextRoute.Attempt,
			DurationMS: elapsed.Milliseconds(),
		})
	}
	if err != nil {
		reason := "admission_refused"
		if errors.Is(err, context.DeadlineExceeded) {
			reason = "deadline_cutoff"
		}
		r.emitRateLimitEvent(RateLimitEvent{
			Type:    RateLimitEventNotSent,
			Method:  nextRoute.Method,
			Attempt: nextRoute.Attempt,
			Reason:  reason,
		})
		return &rateLimitAdmissionError{err: err}
	}
	*requesterAttempt = nextAttempt
	*route = nextRoute
	*costHeader = header
	r.emitRateLimitEvent(RateLimitEvent{
		Type:    RateLimitEventAttempt,
		Method:  nextRoute.Method,
		Attempt: nextRoute.Attempt,
	})
	return nil
}

func (r *Requester) rateLimitAdmissionContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if r == nil || r.RateLimitAdmissionTimeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, r.RateLimitAdmissionTimeout)
}

func (r *Requester) rateLimitRoutePath(requestURL *url.URL) string {
	if requestURL == nil {
		return ""
	}
	path := requestURL.EscapedPath()
	if path == "" {
		path = "/"
	}
	base, err := url.Parse(r.BaseURL)
	if err != nil {
		return path
	}
	basePath := strings.TrimRight(base.EscapedPath(), "/")
	if basePath == "" {
		return path
	}
	if path == basePath {
		return "/"
	}
	if strings.HasPrefix(path, basePath+"/") {
		return strings.TrimPrefix(path, basePath)
	}
	return path
}

func (r *Requester) clientWithRateLimitAdmission(client *http.Client, requesterAttempt *int, route *RateLimitRoute, costHeader *string) *http.Client {
	clone := *client
	checkRedirect := clone.CheckRedirect
	clone.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if checkRedirect != nil {
			if err := checkRedirect(req, via); err != nil {
				return err
			}
		} else if len(via) >= maxRedirects {
			return fmt.Errorf("stopped after %d redirects", maxRedirects)
		}
		return r.admitRequesterSend(req.Context(), req, requesterAttempt, route, costHeader)
	}
	return &clone
}

func isRateLimitAdmissionError(err error) bool {
	var admissionErr *rateLimitAdmissionError
	return errors.As(err, &admissionErr)
}

func isRequestAdmissionError(err error) bool {
	var admissionErr *requestAdmissionError
	return errors.As(err, &admissionErr)
}

// ctxSleep waits for d or returns early if ctx is cancelled.
func ctxSleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// resolveURL builds the absolute request URL from path + query.
func (r *Requester) resolveURL(path string, query url.Values) (string, error) {
	raw := path
	if !strings.HasPrefix(path, "http://") && !strings.HasPrefix(path, "https://") {
		base := strings.TrimRight(r.BaseURL, "/")
		raw = base + "/" + strings.TrimLeft(path, "/")
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse url %q: %w", raw, err)
	}
	if len(query) > 0 {
		existing := u.Query()
		for k, vs := range query {
			existing.Del(k)
			for _, v := range vs {
				existing.Add(k, v)
			}
		}
		u.RawQuery = existing.Encode()
	}
	return u.String(), nil
}

// Do performs an HTTP request with an optional JSON body, retrying on transient
// failures, and returns the captured response. A 4xx/5xx after retries is
// returned as *HTTPError.
func (r *Requester) Do(ctx context.Context, method, path string, query url.Values, body any) (*Response, error) {
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
	}
	return r.do(ctx, method, path, query, payload, "application/json", DefaultMaxResponseBody)
}

// DoLimited performs Do while bounding the captured successful response body to
// maxBodyBytes+1. Callers can reject len(resp.Body) > maxBodyBytes without ever
// buffering the default 64 MiB response cap.
func (r *Requester) DoLimited(ctx context.Context, method, path string, query url.Values, body any, maxBodyBytes int) (*Response, error) {
	return r.DoJSONLimited(ctx, method, path, query, body, "application/json", maxBodyBytes)
}

// DoStatusCheck executes a bodyless HEAD request and returns the final response
// metadata even when its status is non-2xx. Transport, request setup, and
// admission failures still return errors.
func (r *Requester) DoStatusCheck(ctx context.Context, path string, query url.Values, maxBodyBytes int) (*Response, error) {
	return r.doWithBodyPolicy(ctx, http.MethodHead, path, query, maxBodyBytes+1, true, func() (*requestBody, error) {
		return nil, nil
	})
}

// DoJSONLimited is the narrow JSON-family counterpart to DoLimited. The
// operation engine admits its media type from a closed declaration set; this
// method does not expose arbitrary raw request media types to callers.
func (r *Requester) DoJSONLimited(ctx context.Context, method, path string, query url.Values, body any, contentType string, maxBodyBytes int) (*Response, error) {
	var payload []byte
	if body != nil {
		var err error
		payload, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request body: %w", err)
		}
	}
	return r.do(ctx, method, path, query, payload, contentType, maxBodyBytes+1)
}

// DoPreparedJSONLimited sends the exact JSON bytes a closed operation preview
// already bound. It is deliberately JSON-only: callers cannot select a raw
// media type, headers, route, or transport policy through this helper. The
// engine uses it after schema validation and bounded preparation so execution
// cannot re-marshal caller-owned maps while approval is pending.
func (r *Requester) DoPreparedJSONLimited(ctx context.Context, method, path string, query url.Values, payload []byte, contentType string, maxBodyBytes int) (*Response, error) {
	if len(payload) != 0 && !json.Valid(payload) {
		return nil, errors.New("prepared JSON body is invalid")
	}
	return r.do(ctx, method, path, query, append([]byte(nil), payload...), contentType, maxBodyBytes+1)
}

// DoTextLimited performs a bounded request with one literal text/plain body.
// It deliberately has no caller-selected media type: operation executors use
// it only after admitting an explicit text/plain declaration, and it shares
// DoLimited's request core (including auth, retry, admission, and observation)
// rather than offering a generic raw HTTP escape hatch.
func (r *Requester) DoTextLimited(ctx context.Context, method, path string, query url.Values, body string, maxBodyBytes int) (*Response, error) {
	return r.do(ctx, method, path, query, []byte(body), "text/plain", maxBodyBytes+1)
}

// DoBinaryLimited sends one declaration-owned application/octet-stream body.
// The media type is fixed here rather than accepted from the caller, keeping
// this a binary capability instead of a generic raw-request escape hatch.
func (r *Requester) DoBinaryLimited(ctx context.Context, method, path string, query url.Values, body []byte, maxBodyBytes int) (*Response, error) {
	if body == nil {
		body = []byte{}
	}
	return r.do(ctx, method, path, query, body, "application/octet-stream", maxBodyBytes+1)
}

// DoForm performs an HTTP request with an application/x-www-form-urlencoded body,
// reusing the same auth, retry, and rate-limit handling as Do. It is the form
// counterpart used by APIs (e.g. Stripe) whose write endpoints take form bodies.
// A nil/empty form sends no body.
func (r *Requester) DoForm(ctx context.Context, method, path string, query, form url.Values) (*Response, error) {
	var payload []byte
	contentType := ""
	if len(form) > 0 {
		payload = []byte(form.Encode())
		contentType = "application/x-www-form-urlencoded"
	}
	return r.do(ctx, method, path, query, payload, contentType, DefaultMaxResponseBody)
}

// DoFormLimited performs DoForm while bounding the captured successful
// response body to maxBodyBytes+1. It is the form counterpart to DoLimited so
// a declared form write does not trade typed request shaping for an unbounded
// response buffer.
func (r *Requester) DoFormLimited(ctx context.Context, method, path string, query, form url.Values, maxBodyBytes int) (*Response, error) {
	var payload []byte
	contentType := ""
	if len(form) > 0 {
		payload = []byte(form.Encode())
		contentType = "application/x-www-form-urlencoded"
	}
	return r.do(ctx, method, path, query, payload, contentType, maxBodyBytes+1)
}

// DoPreparedFormLimited sends the exact canonical form body a closed
// operation preview already bound. Form parsing and re-encoding must agree so
// this remains a typed form capability rather than a raw-body escape hatch.
func (r *Requester) DoPreparedFormLimited(ctx context.Context, method, path string, query url.Values, payload []byte, maxBodyBytes int) (*Response, error) {
	if len(payload) == 0 {
		return r.do(ctx, method, path, query, nil, "", maxBodyBytes+1)
	}
	form, err := url.ParseQuery(string(payload))
	if err != nil || form.Encode() != string(payload) {
		return nil, errors.New("prepared form body is not canonical")
	}
	return r.do(ctx, method, path, query, append([]byte(nil), payload...), "application/x-www-form-urlencoded", maxBodyBytes+1)
}

// DoMultipart performs an HTTP request with a multipart/form-data body. File
// parts are opened for each retry attempt, so callers may use it with the same
// retry policy as JSON/form requests without reusing a consumed reader.
func (r *Requester) DoMultipart(ctx context.Context, method, path string, query url.Values, form MultipartForm) (*Response, error) {
	return r.doMultipart(ctx, method, path, query, form, DefaultMaxResponseBody)
}

// DoMultipartLimited performs DoMultipart while bounding a successful response
// to maxBodyBytes+1. It is the multipart counterpart to DoLimited and
// DoFormLimited: operation-level rest.max_bytes constrains capture itself,
// rather than allowing a multipart response to fill the default 64 MiB buffer
// before the caller can reject it.
func (r *Requester) DoMultipartLimited(ctx context.Context, method, path string, query url.Values, form MultipartForm, maxBodyBytes int) (*Response, error) {
	return r.doMultipart(ctx, method, path, query, form, maxBodyBytes+1)
}

func (r *Requester) doMultipart(ctx context.Context, method, path string, query url.Values, form MultipartForm, maxResponseBytes int) (*Response, error) {
	if err := validateMultipartForm(form); err != nil {
		return nil, err
	}
	prepared, cleanup, err := snapshotApprovedMultipartFiles(ctx, form)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	boundary, err := multipartBoundary(prepared)
	if err != nil {
		return nil, err
	}
	return r.doWithBody(ctx, method, path, query, maxResponseBytes, func() (*requestBody, error) {
		return multipartBody(prepared, boundary)
	})
}

func validateMultipartForm(form MultipartForm) error {
	var total int64
	for i, file := range form.Files {
		if strings.TrimSpace(file.FieldName) == "" {
			return fmt.Errorf("multipart file %d field name is required", i)
		}
		if strings.TrimSpace(file.sourceName()) == "" {
			return fmt.Errorf("multipart file %q path is required", file.FieldName)
		}
		if file.ExpectedSHA256 != "" {
			digest, err := hex.DecodeString(file.ExpectedSHA256)
			if err != nil || len(digest) != sha256.Size {
				return fmt.Errorf("multipart file %q expected SHA-256 is invalid", file.FieldName)
			}
		}
		for _, allowed := range file.AllowedMediaTypes {
			if _, _, err := mime.ParseMediaType(allowed); err != nil {
				return fmt.Errorf("multipart file %q allowed media type %q is invalid: %w", file.FieldName, allowed, err)
			}
		}
		info, err := file.stat()
		if err != nil {
			return fmt.Errorf("multipart file %q: %w", file.FieldName, err)
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("multipart file %q must be a regular file", file.FieldName)
		}
		if file.MaxBytes > 0 && info.Size() > file.MaxBytes {
			return fmt.Errorf("multipart file %q too large: %d bytes exceeds limit %d", file.FieldName, info.Size(), file.MaxBytes)
		}
		total += info.Size()
		if form.MaxBytes > 0 && total > form.MaxBytes {
			return fmt.Errorf("multipart files too large: %d bytes exceeds limit %d", total, form.MaxBytes)
		}
	}
	return nil
}

func snapshotApprovedMultipartFiles(ctx context.Context, form MultipartForm) (MultipartForm, func(), error) {
	prepared := form
	prepared.Files = append([]MultipartFile(nil), form.Files...)
	tempPaths := make([]string, 0, len(form.Files))
	cleanup := func() {
		for _, path := range tempPaths {
			_ = os.Remove(path)
		}
	}
	var total int64
	for i, file := range form.Files {
		if !file.needsSnapshot() {
			info, err := file.stat()
			if err != nil {
				cleanup()
				return MultipartForm{}, func() {}, fmt.Errorf("multipart file %q: %w", file.FieldName, err)
			}
			total += info.Size()
			continue
		}
		limit := file.MaxBytes
		if limit <= 0 {
			limit = -1
		}
		if form.MaxBytes > 0 {
			remaining := form.MaxBytes - total
			if remaining < 0 {
				cleanup()
				return MultipartForm{}, func() {}, fmt.Errorf("multipart files too large: exceeds limit %d", form.MaxBytes)
			}
			if limit < 0 || remaining < limit {
				limit = remaining
			}
		}
		tempPath, size, digest, sniffed, err := snapshotMultipartFile(ctx, file, limit)
		if err != nil {
			cleanup()
			return MultipartForm{}, func() {}, err
		}
		tempPaths = append(tempPaths, tempPath)
		if file.ExpectedSHA256 != "" {
			expected, _ := hex.DecodeString(file.ExpectedSHA256)
			if !bytes.Equal(digest, expected) {
				cleanup()
				return MultipartForm{}, func() {}, fmt.Errorf("multipart file %q changed since approval", file.FieldName)
			}
		}
		if err := checkAllowedMediaType(file, sniffed); err != nil {
			cleanup()
			return MultipartForm{}, func() {}, err
		}
		if prepared.Files[i].FileName == "" {
			prepared.Files[i].FileName = filepath.Base(file.sourceName())
		}
		// The snapshot lives outside the root by design, so the root handle is
		// dropped: subsequent opens must target the verified copy, which is what
		// makes the digest and media-type checks binding.
		prepared.Files[i].Path = tempPath
		prepared.Files[i].Root = nil
		prepared.Files[i].RelPath = ""
		// The part header now describes the bytes we actually send, not the type
		// the bundle hoped for. checkAllowedMediaType has already confirmed the
		// sniffed type is one the bundle declared acceptable, so this is both
		// truthful and within the declared bound. Only set when an allowlist made
		// it binding: without one the sniff is unverified against any declaration,
		// and http.DetectContentType is coarse enough (every CSV is text/plain)
		// that overriding a deliberate content_type would lose information.
		if len(file.AllowedMediaTypes) > 0 {
			prepared.Files[i].ContentType = sniffed
		}
		total += size
	}
	return prepared, cleanup, nil
}

// checkAllowedMediaType rejects a file whose actual bytes do not sniff as one of
// the media types the bundle declared. Upload fails closed here, deliberately
// unlike the download direction: on download the provider makes the Content-Type
// claim and providers misreport it, so a mismatch is recorded and surfaced; on
// upload we are the party making the claim to the provider, so an unsatisfiable
// claim is our bug and must not reach the wire.
func checkAllowedMediaType(file MultipartFile, sniffed string) error {
	if len(file.AllowedMediaTypes) == 0 {
		return nil
	}
	got, _, err := mime.ParseMediaType(sniffed)
	if err != nil {
		return fmt.Errorf("multipart file %q content type %q is not a valid media type: %w", file.FieldName, sniffed, err)
	}
	for _, allowed := range file.AllowedMediaTypes {
		want, _, err := mime.ParseMediaType(allowed)
		if err != nil {
			return fmt.Errorf("multipart file %q allowed media type %q is invalid: %w", file.FieldName, allowed, err)
		}
		if strings.EqualFold(got, want) {
			return nil
		}
	}
	if got == "application/octet-stream" {
		return fmt.Errorf("multipart file %q content could not be classified (sniffed %s); allowed media types are %s", file.FieldName, got, strings.Join(file.AllowedMediaTypes, ", "))
	}
	return fmt.Errorf("multipart file %q content sniffed as %s, which is not among the allowed media types %s", file.FieldName, got, strings.Join(file.AllowedMediaTypes, ", "))
}

// sniffLimit is the number of leading bytes http.DetectContentType inspects.
const sniffLimit = 512

// snapshotMultipartFile copies the source into a bounded temp file, computing
// the SHA-256 and sniffing the media type in the same pass, so neither costs an
// extra read and neither can race a separate one.
func snapshotMultipartFile(ctx context.Context, file MultipartFile, maxBytes int64) (string, int64, []byte, string, error) {
	source, err := file.open()
	if err != nil {
		return "", 0, nil, "", fmt.Errorf("multipart file %q: %w", file.FieldName, err)
	}
	defer func() { _ = source.Close() }()
	temp, err := os.CreateTemp("", "polymetrics-upload-*")
	if err != nil {
		return "", 0, nil, "", fmt.Errorf("snapshot multipart file %q: %w", file.FieldName, err)
	}
	tempPath := temp.Name()
	removeTemp := true
	defer func() {
		_ = temp.Close()
		if removeTemp {
			_ = os.Remove(tempPath)
		}
	}()

	hash := sha256.New()
	prefix := &prefixWriter{limit: sniffLimit}
	reader := io.Reader(&contextReader{ctx: ctx, reader: source})
	if maxBytes >= 0 {
		reader = io.LimitReader(reader, maxBytes)
	}
	written, err := io.Copy(io.MultiWriter(temp, hash, prefix), reader)
	if err != nil {
		return "", written, nil, "", fmt.Errorf("snapshot multipart file %q: %w", file.FieldName, err)
	}
	if maxBytes >= 0 && written == maxBytes {
		var extra [1]byte
		n, readErr := (&contextReader{ctx: ctx, reader: source}).Read(extra[:])
		if n > 0 {
			return "", written, nil, "", fmt.Errorf("multipart file %q too large: exceeds limit %d", file.FieldName, maxBytes)
		}
		if readErr != nil && readErr != io.EOF {
			return "", written, nil, "", fmt.Errorf("snapshot multipart file %q: %w", file.FieldName, readErr)
		}
	}
	if err := temp.Close(); err != nil {
		return "", written, nil, "", fmt.Errorf("snapshot multipart file %q: %w", file.FieldName, err)
	}
	removeTemp = false
	return tempPath, written, hash.Sum(nil), http.DetectContentType(prefix.head), nil
}

// prefixWriter retains the first limit bytes written through it and discards the
// rest, so the sniff sample rides the snapshot copy instead of a second read.
type prefixWriter struct {
	head  []byte
	limit int
}

func (w *prefixWriter) Write(p []byte) (int, error) {
	if remaining := w.limit - len(w.head); remaining > 0 {
		if len(p) < remaining {
			remaining = len(p)
		}
		w.head = append(w.head, p[:remaining]...)
	}
	return len(p), nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r *contextReader) Read(p []byte) (int, error) {
	select {
	case <-r.ctx.Done():
		return 0, r.ctx.Err()
	default:
		return r.reader.Read(p)
	}
}

type multipartCountingWriter struct {
	size int64
}

func (w *multipartCountingWriter) Write(p []byte) (int, error) {
	if err := w.add(int64(len(p))); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (w *multipartCountingWriter) add(n int64) error {
	if n < 0 || w.size > (1<<63-1)-n {
		return fmt.Errorf("multipart payload size overflow")
	}
	w.size += n
	return nil
}

type multipartLimitWriter struct {
	writer    io.Writer
	remaining int64
}

func (w *multipartLimitWriter) Write(p []byte) (int, error) {
	if int64(len(p)) > w.remaining {
		return 0, fmt.Errorf("multipart payload too large: exceeds limit")
	}
	n, err := w.writer.Write(p)
	w.remaining -= int64(n)
	return n, err
}

func multipartBoundary(form MultipartForm) (string, error) {
	counter := &multipartCountingWriter{}
	writer := multipart.NewWriter(counter)
	keys := make([]string, 0, len(form.Fields))
	for key := range form.Fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := writer.WriteField(key, form.Fields[key]); err != nil {
			return "", err
		}
	}
	for _, file := range form.Files {
		if _, err := writer.CreatePart(multipartFileHeader(file)); err != nil {
			return "", err
		}
		info, err := file.stat()
		if err != nil {
			return "", fmt.Errorf("multipart file %q: %w", file.FieldName, err)
		}
		if err := counter.add(info.Size()); err != nil {
			return "", err
		}
	}
	if err := writer.Close(); err != nil {
		return "", err
	}
	if form.MaxBytes > 0 && counter.size > form.MaxBytes {
		return "", fmt.Errorf("multipart payload too large: %d bytes exceeds limit %d", counter.size, form.MaxBytes)
	}
	return writer.Boundary(), nil
}

func multipartBody(form MultipartForm, boundary string) (*requestBody, error) {
	pr, pw := io.Pipe()
	var sink io.Writer = pw
	if form.MaxBytes > 0 {
		sink = &multipartLimitWriter{writer: pw, remaining: form.MaxBytes}
	}
	mw := multipart.NewWriter(sink)
	if err := mw.SetBoundary(boundary); err != nil {
		_ = pr.Close()
		_ = pw.Close()
		return nil, err
	}
	done := make(chan error, 1)
	go func() {
		err := writeMultipartForm(mw, form)
		if closeErr := mw.Close(); err == nil {
			err = closeErr
		}
		if err != nil {
			_ = pw.CloseWithError(err)
		} else {
			_ = pw.Close()
		}
		done <- err
	}()
	return &requestBody{
		Reader:      pr,
		ContentType: mw.FormDataContentType(),
		Cleanup: func() error {
			_ = pr.Close()
			err := <-done
			// A server can send a final response before it consumes the
			// complete streamed upload. Closing the reader then unblocks the
			// producer with io.ErrClosedPipe; that expected cleanup result must
			// not hide the provider response.
			if errors.Is(err, io.ErrClosedPipe) {
				return nil
			}
			return err
		},
	}, nil
}

func writeMultipartForm(mw *multipart.Writer, form MultipartForm) error {
	keys := make([]string, 0, len(form.Fields))
	for key := range form.Fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if err := mw.WriteField(key, form.Fields[key]); err != nil {
			return err
		}
	}
	var total int64
	for _, file := range form.Files {
		limit := file.MaxBytes
		if limit <= 0 {
			limit = -1
		}
		if form.MaxBytes > 0 {
			remaining := form.MaxBytes - total
			if limit < 0 || remaining < limit {
				limit = remaining
			}
		}
		written, err := writeMultipartFile(mw, file, limit)
		if err != nil {
			return err
		}
		total += written
	}
	return nil
}

func writeMultipartFile(mw *multipart.Writer, file MultipartFile, maxBytes int64) (int64, error) {
	part, err := mw.CreatePart(multipartFileHeader(file))
	if err != nil {
		return 0, err
	}
	f, err := file.open()
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()
	if maxBytes < 0 {
		written, err := io.Copy(part, f)
		return written, err
	}
	written, err := io.CopyN(part, f, maxBytes)
	if err != nil && err != io.EOF {
		return written, err
	}
	if written < maxBytes {
		return written, nil
	}
	var extra [1]byte
	n, readErr := f.Read(extra[:])
	if n > 0 {
		return written, fmt.Errorf("multipart file %q too large: exceeds limit %d", file.FieldName, maxBytes)
	}
	if readErr != nil && readErr != io.EOF {
		return written, readErr
	}
	return written, nil
}

func multipartFileHeader(file MultipartFile) textproto.MIMEHeader {
	name := file.FileName
	if strings.TrimSpace(name) == "" {
		name = filepath.Base(file.sourceName())
	}
	header := make(textproto.MIMEHeader)
	header.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename=%q`, file.FieldName, name))
	if file.ContentType != "" {
		header.Set("Content-Type", file.ContentType)
	}
	return header
}

// do is the shared request core for Do/DoForm. payload is the already-encoded
// body (nil for none) and contentType is the Content-Type to set when a body is
// present.
func (r *Requester) do(ctx context.Context, method, path string, query url.Values, payload []byte, contentType string, maxBodyBytes int) (*Response, error) {
	return r.doWithBody(ctx, method, path, query, maxBodyBytes, func() (*requestBody, error) {
		if payload == nil {
			return nil, nil
		}
		return &requestBody{Reader: bytes.NewReader(payload), ContentType: contentType}, nil
	})
}

func (r *Requester) doWithBody(ctx context.Context, method, path string, query url.Values, maxBodyBytes int, bodyFactory func() (*requestBody, error)) (*Response, error) {
	return r.doWithBodyPolicy(ctx, method, path, query, maxBodyBytes, false, bodyFactory)
}

func (r *Requester) doWithBodyPolicy(ctx context.Context, method, path string, query url.Values, maxBodyBytes int, returnFinalStatus bool, bodyFactory func() (*requestBody, error)) (*Response, error) {
	fullURL, err := r.resolveURL(path, query)
	if err != nil {
		return nil, err
	}
	if maxBodyBytes <= 0 {
		maxBodyBytes = DefaultMaxResponseBody
	}
	baseURL, err := url.Parse(fullURL)
	if err != nil {
		return nil, fmt.Errorf("parse url %q: %w", fullURL, err)
	}

	attempts := r.maxRetries() + 1
	if r.DisableRetries {
		attempts = 1
	}
	var lastErr error
	// A retry terminal cause (cancellation, a later dial failure, or a refused
	// redirect) must not erase the most recent provider response. Keep both the
	// bounded receipt and its typed error independently of the latest transport
	// failure so callers can publish one complete terminal receipt.
	var lastProviderResponse *Response
	var lastProviderErr error
	// reauthAttempted bounds the 401-refresh path to ONCE per request. It is
	// set before the refresh is attempted (see below), so a provider that keeps
	// returning 401 terminates with that 401 instead of being hammered.
	reauthAttempted := false
	requesterAttempt := 0
	route := RateLimitRoute{}
	costHeader := ""
	// Mutation authority is independent of retry eligibility. A provider-
	// scoped idempotency key may permit another attempt against this exact
	// URL, but it never authorizes following a redirect or discarding the last
	// provider response.
	strictWrite := !isSafeReplayableRead(method)
	var credKeys []string
	baseClient := r.redirectClient(ctx, baseURL, r.RedirectPolicy, &credKeys)
	if strictWrite {
		if r.DisableRetries {
			baseClient = noReplayResponseClient(baseClient)
		} else {
			baseClient = noRedirectResponseClient(baseClient)
		}
	}
	client := r.clientWithRateLimitAdmission(baseClient, &requesterAttempt, &route, &costHeader)
	for attempt := 0; attempt < attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return lastProviderResponse, errors.Join(lastProviderErr, err)
		}

		body, err := bodyFactory()
		if err != nil {
			return nil, err
		}
		var reader io.Reader
		var contentType string
		if body != nil {
			reader = body.Reader
			contentType = body.ContentType
		}
		req, err := http.NewRequestWithContext(ctx, method, fullURL, reader)
		if err != nil {
			_ = cleanupRequestBody(body)
			return nil, fmt.Errorf("build request: %w", err)
		}
		r.applyHeaders(req, body != nil, contentType)
		before := req.Header.Clone()
		if r.Auth != nil {
			if err := r.Auth.Apply(ctx, req); err != nil {
				_ = cleanupRequestBody(body)
				return nil, fmt.Errorf("apply auth: %w", err)
			}
		}
		credKeys = credentialHeaderKeys(before, req.Header, r.DefaultHeaders, r.DefaultHeaderValues)
		if r.DisableRetries {
			disableTransportReplay(req, strictWrite)
		}
		if err := r.admitRequesterSend(ctx, req, &requesterAttempt, &route, &costHeader); err != nil {
			_ = cleanupRequestBody(body)
			return lastProviderResponse, errors.Join(lastProviderErr, err)
		}

		resp, err := client.Do(req)
		if err != nil {
			bodyErr := cleanupRequestBody(body)
			terminal := captureTerminalResponse(resp, maxBodyBytes, fullURL, route)
			lastErr = fmt.Errorf("send request: %w", err)
			if bodyErr != nil {
				lastErr = errors.Join(lastErr, fmt.Errorf("send request body: %w", bodyErr))
			}
			if terminal != nil {
				lastProviderResponse = terminal
				lastProviderErr = responseHTTPError(terminal.Status, fullURL, terminal.Header, terminal.Body, RateLimitObservation{})
			}
			if errors.Is(err, transportpolicy.ErrRedirectRefused) {
				return lastProviderResponse, errors.Join(lastProviderErr, lastErr)
			}
			if isRateLimitAdmissionError(err) || isRequestAdmissionError(err) {
				return lastProviderResponse, errors.Join(lastProviderErr, lastErr)
			}
			if attempt < attempts-1 {
				if werr := r.sleep(ctx, r.backoff(attempt, RateLimitObservation{})); werr != nil {
					return lastProviderResponse, errors.Join(lastProviderErr, lastErr, werr)
				}
				continue
			}
			return lastProviderResponse, errors.Join(lastProviderErr, lastErr)
		}
		observation := r.observeRateLimit(ctx, route, resp.StatusCode, resp.Header, costHeader)
		bodyErr := cleanupRequestBody(body)
		if bodyErr != nil {
			return captureTerminalResponse(resp, maxBodyBytes, fullURL, route), fmt.Errorf("send request body: %w", bodyErr)
		}

		respBody, readErr := io.ReadAll(io.LimitReader(resp.Body, int64(maxBodyBytes)))
		_ = resp.Body.Close()
		terminal := &Response{Status: resp.StatusCode, Header: resp.Header.Clone(), Body: respBody, requestURL: fullURL, rateLimitRoute: route}
		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
			lastProviderResponse = terminal
		}
		if resp.StatusCode >= 200 && resp.StatusCode < 300 && !r.acceptsSuccessfulStatus(resp.StatusCode) {
			statusErr := &UnexpectedStatusError{Status: resp.StatusCode}
			if readErr != nil {
				return terminal, errors.Join(statusErr, fmt.Errorf("read response body: %w", readErr))
			}
			return terminal, statusErr
		}
		if readErr != nil {
			readBodyErr := fmt.Errorf("read response body: %w", readErr)
			if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
				return terminal, errors.Join(requesterResponseStatusError(ctx, strictWrite, resp.StatusCode, fullURL, resp.Header, respBody, observation), readBodyErr)
			}
			return terminal, readBodyErr
		}

		// A 401 can mean the credential was invalidated out of band (revoked
		// grant, password change, scope change) rather than that it was never
		// valid — something an expiry clock cannot see. An Authenticator that
		// knows how to renew itself gets exactly one chance to do so.
		//
		// Authenticators that do not implement AuthRefresher — every mode that
		// predates the refresh-token grant — never enter this branch, so their
		// 401 behaviour is byte-for-byte unchanged.
		if !r.DisableRetries && resp.StatusCode == http.StatusUnauthorized && !reauthAttempted {
			if refresher, ok := r.Auth.(AuthRefresher); ok {
				// Set before the attempt: a refresh that itself errors must not
				// buy a second one.
				reauthAttempted = true
				if err := refresher.RefreshAuth(ctx, req); err == nil {
					lastErr = &HTTPError{Status: resp.StatusCode, URL: fullURL, Header: resp.Header.Clone(), Body: string(respBody), RawBody: append([]byte(nil), respBody...)}
					lastProviderErr = lastErr
					// The reauth retry does not spend the transient-failure
					// budget, so a MaxRetries:0 requester still gets its one
					// post-refresh attempt. Bounded by reauthAttempted, which
					// is now true, so this can decrement at most once.
					attempt--
					continue
				}
				// The refresh failed; fall through and report the 401 itself,
				// which is the more useful of the two errors.
			}
		}

		if !r.DisableRetries && r.shouldRetry(resp.StatusCode) && attempt < attempts-1 {
			lastErr = responseHTTPError(resp.StatusCode, fullURL, resp.Header, respBody, observation)
			lastProviderErr = lastErr
			if werr := r.sleep(ctx, r.backoff(attempt, observation)); werr != nil {
				return lastProviderResponse, errors.Join(lastProviderErr, werr)
			}
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			if returnFinalStatus {
				return terminal, nil
			}
			statusErr := requesterResponseStatusError(ctx, strictWrite, resp.StatusCode, fullURL, resp.Header, respBody, observation)
			// Status-only checks intentionally retain their final non-2xx
			// response. Destructive declared writes additionally need the typed
			// provider response for result preservation. Ordinary binary/text GET
			// callers keep their established nil-response error semantics.
			if strictWrite || transportpolicy.IsDestructive(ctx) {
				return terminal, statusErr
			}
			return nil, statusErr
		}

		return terminal, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("request to %s failed after %d attempts", fullURL, attempts)
	}
	return lastProviderResponse, errors.Join(lastProviderErr, lastErr)
}

func (r *Requester) acceptsSuccessfulStatus(status int) bool {
	if len(r.AcceptedStatuses) == 0 {
		return true
	}
	for _, allowed := range r.AcceptedStatuses {
		if status >= allowed.Min && status <= allowed.Max {
			return true
		}
	}
	return false
}

func captureTerminalResponse(resp *http.Response, maxBodyBytes int, fullURL string, route RateLimitRoute) *Response {
	if resp == nil {
		return nil
	}
	var body []byte
	if resp.Body != nil {
		body, _ = io.ReadAll(io.LimitReader(resp.Body, int64(maxBodyBytes)))
		_ = resp.Body.Close()
	}
	return &Response{Status: resp.StatusCode, Header: resp.Header, Body: body, requestURL: fullURL, rateLimitRoute: route}
}

func cleanupRequestBody(body *requestBody) error {
	if body == nil || body.Cleanup == nil {
		return nil
	}
	return body.Cleanup()
}

// DoJSON performs a request and decodes a successful response into out (which may
// be nil to discard the body). Numbers are decoded with json.Number to preserve
// integer fidelity, matching the rest of the codebase.
func (r *Requester) DoJSON(ctx context.Context, method, path string, query url.Values, body, out any) error {
	resp, err := r.Do(ctx, method, path, query, body)
	if err != nil {
		return err
	}
	if out == nil || len(resp.Body) == 0 {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(resp.Body))
	dec.UseNumber()
	if err := dec.Decode(out); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}
	return nil
}

func (r *Requester) applyHeaders(req *http.Request, hasBody bool, contentType string) {
	accept := r.Accept
	if accept == "" {
		accept = "application/json"
	}
	req.Header.Set("Accept", accept)
	if r.UserAgent != "" {
		req.Header.Set("User-Agent", r.UserAgent)
	}
	if hasBody {
		if contentType == "" {
			contentType = "application/json"
		}
		req.Header.Set("Content-Type", contentType)
	}
	for k, v := range r.DefaultHeaders {
		req.Header.Set(k, v)
	}
	for k, values := range r.DefaultHeaderValues {
		for _, value := range values {
			req.Header.Add(k, value)
		}
	}
}

// backoff computes the wait before the next attempt. A valid provider
// Retry-After is deterministic and is honored exactly, even when it exceeds
// MaxBackoff. Only unhinted fallback retries use bounded full jitter.
func (r *Requester) backoff(attempt int, observation RateLimitObservation) time.Duration {
	if observation.HasRetryAfter {
		if observation.HasReset {
			remaining := observation.ResetAt.Sub(r.now())
			if remaining <= 0 {
				return 0
			}
			return remaining
		}
		return observation.RetryAfter
	}
	return r.fullJitter(r.fallbackBackoff(attempt))
}

func (r *Requester) fallbackBackoff(attempt int) time.Duration {
	delay := r.baseBackoff()
	maximum := r.maxBackoff()
	if delay >= maximum {
		return maximum
	}
	for i := 0; i < attempt; i++ {
		if delay > maximum/2 {
			return maximum
		}
		delay *= 2
	}
	if delay > maximum {
		return maximum
	}
	return delay
}

func (r *Requester) fullJitter(cap time.Duration) time.Duration {
	if cap <= 0 {
		return 0
	}
	if r.Jitter == nil {
		return time.Duration(rand.Int64N(int64(cap)))
	}
	delay := r.Jitter(cap)
	if delay < 0 {
		return 0
	}
	if delay > cap {
		return cap
	}
	return delay
}

func responseHTTPError(status int, requestURL string, header http.Header, body []byte, observation RateLimitObservation) error {
	httpErr := &HTTPError{Status: status, URL: requestURL, Header: header.Clone(), Body: string(body), RawBody: append([]byte(nil), body...)}
	if status != http.StatusTooManyRequests {
		return httpErr
	}
	return &RateLimitError{
		HTTPError:       httpErr,
		Source:          observation.Source,
		RetryAfter:      observation.RetryAfter,
		HasRetryAfter:   observation.HasRetryAfter,
		ResetAt:         observation.ResetAt,
		HasReset:        observation.HasReset,
		ResetAtAbsolute: observation.ResetAtAbsolute,
	}
}
func requesterResponseStatusError(ctx context.Context, strictWrite bool, status int, requestURL string, header http.Header, body []byte, observation RateLimitObservation) error {
	responseErr := responseHTTPError(status, requestURL, header, body, observation)
	if (strictWrite || transportpolicy.IsDestructive(ctx)) && status >= http.StatusMultipleChoices && status < http.StatusBadRequest {
		return fmt.Errorf("%w: %w", transportpolicy.ErrRedirectRefused, responseErr)
	}
	return responseErr
}

func truncate(body []byte) string {
	if len(body) > maxErrorBody {
		return string(body[:maxErrorBody])
	}
	return string(body)
}
