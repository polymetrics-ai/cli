package connsdk

import (
	"context"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// RateLimitRequest is the safe, per-attempt context given to a requester
// admission implementation. It intentionally carries no URL, headers, body,
// credentials, or scope identity. #3753 selects a declared policy and #3754
// constructs any opaque scope key from a credential binding, policy, and
// declared non-secret subject; neither may derive a key from this request.
// Logical sends include outer Client.Do calls and permitted redirect hops. A
// safe replayable read can be replayed inside net/http without another
// admission and does not increment Attempt.
type RateLimitRequest struct {
	Method  string
	Attempt int // one-based logical Requester send
}

type RateLimitRoute struct {
	Method  string
	Path    string
	Attempt int
}

type RateLimitRouteResolver interface {
	AdmitRoute(context.Context, RateLimitRoute) (string, error)
	ObserveRoute(context.Context, RateLimitRoute, RateLimitObservation)
}

// RateLimitAdmission gates a logical Requester send. Implementations must honor
// ctx so a rate-limit wait cannot outlive the caller. An admission error
// prevents the requester from sending that attempt. A successful call permits
// one logical send, not every physical transport write.
type RateLimitAdmission interface {
	Admit(ctx context.Context, request RateLimitRequest) error
}

// RateLimitObserver receives parsed, secret-free rate-limit facts from a
// provider response. It is called synchronously before a retry is scheduled
// so attached policies can tighten their next admissions. It is not an operator
// output hook; #3755 owns human and JSON event rendering.
type RateLimitObserver interface {
	Observe(ctx context.Context, observation RateLimitObservation)
}

// RateLimitObservationSource identifies the provider signal that made an
// observation relevant. A response can carry several parsed fields; Source
// names the most specific timing/status signal without retaining raw headers.
type RateLimitObservationSource string

const (
	RateLimitObservationSourceRetryAfter RateLimitObservationSource = "retry_after"
	RateLimitObservationSourceHeaders    RateLimitObservationSource = "response_headers"
	RateLimitObservationSourceHTTP429    RateLimitObservationSource = "http_429"
)

// RateLimitObservation is a typed subset of provider rate-limit response
// metadata. The booleans distinguish an absent header from a valid zero value
// (for example, remaining=0 or Retry-After: 0). It deliberately retains no
// raw response headers, body, URL, or credential-derived data.
type RateLimitObservation struct {
	Source     RateLimitObservationSource
	Status     int
	Attempt    int // same logical-send count as RateLimitRequest.Attempt
	Attempted  bool
	ObservedAt time.Time

	RetryAfter      time.Duration
	HasRetryAfter   bool
	ResetAt         time.Time
	HasReset        bool
	ResetAtAbsolute bool

	Limit        int64
	HasLimit     bool
	Remaining    int64
	HasRemaining bool

	// Cost is the provider-reported cost for a request whose selected policy
	// declared RateLimitCost.ResponseHeader. It is deliberately scalar and
	// typed: the requester never retains a response header map or body in the
	// observation.
	Cost    float64
	HasCost bool
}

// RateLimitError reports a terminal HTTP 429. It preserves the existing
// *HTTPError in its error chain while exposing typed provider reset timing for
// callers that need to make a safe, contextual decision. A zero ResetAt means
// the provider did not send a parseable reset signal.
type RateLimitError struct {
	HTTPError       *HTTPError
	Source          RateLimitObservationSource
	RetryAfter      time.Duration
	HasRetryAfter   bool
	ResetAt         time.Time
	HasReset        bool
	ResetAtAbsolute bool
}

func (e *RateLimitError) Error() string {
	if e == nil || e.HTTPError == nil {
		return "rate limited"
	}
	return e.HTTPError.Error()
}

// Unwrap preserves errors.As(err, *HTTPError) compatibility for callers that
// already classify HTTP failures by status.
func (e *RateLimitError) Unwrap() error {
	if e == nil || e.HTTPError == nil {
		return nil
	}
	return e.HTTPError
}

// rateLimitObservation parses only known numeric rate-limit headers. Unknown
// or provider-specific headers are intentionally not copied through this
// foundation. A declared actual-cost header is parsed into one typed scalar
// without turning arbitrary headers into event payloads.
func rateLimitObservation(status int, header http.Header, attempt int, now time.Time, costHeader string) (RateLimitObservation, bool) {
	observation := RateLimitObservation{
		Status:     status,
		Attempt:    attempt,
		Attempted:  true,
		ObservedAt: now,
	}

	if delay, resetAt, absolute, ok := parseRetryAfterAtWithAbsolute(header.Get("Retry-After"), now); ok {
		observation.Source = RateLimitObservationSourceRetryAfter
		observation.RetryAfter = delay
		observation.HasRetryAfter = true
		observation.ResetAt = resetAt
		observation.HasReset = true
		observation.ResetAtAbsolute = absolute
	}
	if limit, ok := parseNonNegativeRateLimitHeader(header, "RateLimit-Limit", "X-RateLimit-Limit", "X-Rate-Limit-Limit"); ok {
		observation.Limit = limit
		observation.HasLimit = true
	}
	if remaining, ok := parseNonNegativeRateLimitHeader(header, "RateLimit-Remaining", "X-RateLimit-Remaining", "X-Rate-Limit-Remaining"); ok {
		observation.Remaining = remaining
		observation.HasRemaining = true
	}
	if !observation.HasReset {
		if resetAt, absolute, ok := parseRateLimitResetWithAbsolute(header, now); ok {
			observation.ResetAt = resetAt
			observation.HasReset = true
			observation.ResetAtAbsolute = absolute
		}
	}
	if cost, ok := parsePositiveRateLimitCost(header.Get(costHeader)); ok {
		observation.Cost = cost
		observation.HasCost = true
	}

	switch {
	case observation.Source == RateLimitObservationSourceRetryAfter:
		return observation, true
	case observation.HasLimit || observation.HasRemaining || observation.HasReset || observation.HasCost:
		observation.Source = RateLimitObservationSourceHeaders
		return observation, true
	case status == http.StatusTooManyRequests:
		observation.Source = RateLimitObservationSourceHTTP429
		return observation, true
	default:
		return RateLimitObservation{}, false
	}
}

func parsePositiveRateLimitCost(value string) (float64, bool) {
	if strings.TrimSpace(value) == "" {
		return 0, false
	}
	cost, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || cost < 0 || math.IsNaN(cost) || math.IsInf(cost, 0) {
		return 0, false
	}
	return cost, true
}

func parseNonNegativeRateLimitHeader(header http.Header, names ...string) (int64, bool) {
	for _, name := range names {
		value := strings.TrimSpace(header.Get(name))
		if value == "" {
			continue
		}
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err == nil && parsed >= 0 {
			return parsed, true
		}
	}
	return 0, false
}

// parseRateLimitReset supports RFC 9333's RateLimit-Reset delta-seconds and
// the widely used X-RateLimit-Reset Unix-seconds convention. Retry-After is
// handled separately and wins whenever it is present because it is explicit
// about the immediate retry.
func parseRateLimitReset(header http.Header, now time.Time) (time.Time, bool) {
	resetAt, _, ok := parseRateLimitResetWithAbsolute(header, now)
	return resetAt, ok
}

func parseRateLimitResetWithAbsolute(header http.Header, now time.Time) (time.Time, bool, bool) {
	if seconds, ok := parseNonNegativeRateLimitHeader(header, "RateLimit-Reset"); ok {
		if delay, valid := durationFromSeconds(seconds); valid {
			return now.Add(delay), false, true
		}
	}
	if epoch, ok := parseNonNegativeRateLimitHeader(header, "X-RateLimit-Reset", "X-Rate-Limit-Reset"); ok {
		return time.Unix(epoch, 0).UTC(), true, true
	}
	return time.Time{}, false, false
}

// parseRetryAfterAt parses Retry-After as either delay-seconds or an HTTP
// date relative to now. It returns the provider reset time alongside the wait
// duration so callers do not have to parse the header more than once.
func parseRetryAfterAt(value string, now time.Time) (time.Duration, time.Time, bool) {
	delay, resetAt, _, ok := parseRetryAfterAtWithAbsolute(value, now)
	return delay, resetAt, ok
}

func parseRetryAfterAtWithAbsolute(value string, now time.Time) (time.Duration, time.Time, bool, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, time.Time{}, false, false
	}
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		delay, ok := durationFromSeconds(seconds)
		if !ok {
			return 0, time.Time{}, false, false
		}
		return delay, now.Add(delay), false, true
	}
	if resetAt, err := http.ParseTime(value); err == nil {
		delay := resetAt.Sub(now)
		if delay < 0 {
			delay = 0
		}
		return delay, resetAt, true, true
	}
	return 0, time.Time{}, false, false
}

func durationFromSeconds(seconds int64) (time.Duration, bool) {
	const maxDurationSeconds = int64((1<<63 - 1) / int64(time.Second))
	if seconds < 0 || seconds > maxDurationSeconds {
		return 0, false
	}
	return time.Duration(seconds) * time.Second, true
}

// parseRetryAfter remains the compatibility helper for package callers and
// tests that only need a duration. Requester paths use parseRetryAfterAt with
// their injectable clock so retry timing and typed observation agree exactly.
func parseRetryAfter(value string) (time.Duration, bool) {
	delay, _, ok := parseRetryAfterAt(value, time.Now())
	return delay, ok
}
