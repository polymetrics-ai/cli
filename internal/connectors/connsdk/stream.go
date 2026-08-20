package connsdk

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// maxRedirects bounds redirect hops. Go's default client stops at 10;
// installing a CheckRedirect replaces that policy, so the cap is re-stated
// here rather than silently lost.
// StreamOptions is the per-request redirect/credential policy for DoStream.
//
// The zero value is the safe default: same-origin only, credentials attached.
// Every widening is explicit and per-operation.
type StreamOptions struct {
	// Accept is a declaration-owned representation selector. It is never
	// populated from caller input; binary operations use it for endpoints that
	// require a fixed response media type.
	Accept string
	// AllowCrossHost permits a hop to ANY other origin. Credentials are
	// stripped on such a hop regardless.
	AllowCrossHost bool
	// AllowedHosts permits hops to exactly these hosts (host or host:port).
	// Credentials are stripped on such a hop regardless.
	AllowedHosts   []string
	RedirectPolicy *RedirectPolicy
}

type RedirectPolicy struct {
	MaxHops         int
	AllowSameOrigin bool
	AllowedHosts    []string
}

// StreamResponse is a response whose body is still OPEN. The caller owns it
// and MUST Close it.
type StreamResponse struct {
	Status int
	Header http.Header
	Body   io.ReadCloser
	URL    string
}

// DoStream performs a request and returns the response body as an open
// io.ReadCloser instead of buffering it.
//
// It exists because every other Requester method funnels through doWithBody,
// which reads the whole body into Response.Body []byte under a 64 MiB cap — so
// a declared 100 MiB max_bytes cannot be satisfied through Do/DoLimited, and a
// large download would be buffered entirely in memory even when it fits.
//
// Two properties are load-bearing:
//
//   - Retry safety. The body is NEVER read inside the retry loop. On a
//     transport error or a retryable status the response body is closed and
//     discarded before the next attempt, so partial bytes from a failed
//     attempt can never concatenate with a later one. The body is handed to
//     the caller only on the terminal successful attempt.
//
//   - Redirect credential safety. Go strips only Authorization,
//     WWW-Authenticate and Cookie across cross-domain redirects, but 71
//     connector definitions authenticate with a custom header (X-API-Key,
//     Circle-Token, DOLAPIKEY, Ocp-Apim-Subscription-Key, ...) which Go
//     FORWARDS. Download endpoints redirect to CDNs constantly. DoStream
//     installs an explicit CheckRedirect that re-checks the origin on every
//     hop and strips every credential-bearing header before a cross-origin
//     hop proceeds.
//
// The redirect policy is installed on a CLONE of the client. Mutating the
// shared Requester.Client would silently apply this policy to every other
// request the connector makes.
func (r *Requester) DoStream(ctx context.Context, method, path string, query url.Values, opts StreamOptions) (*StreamResponse, error) {
	fullURL, err := r.resolveURL(path, query)
	if err != nil {
		return nil, err
	}
	base, err := url.Parse(fullURL)
	if err != nil {
		return nil, fmt.Errorf("parse url %q: %w", fullURL, err)
	}

	// credKeys names the header keys carrying credentials for the in-flight
	// attempt. It is captured by the redirect policy rather than smuggled
	// through a header, so nothing extra ever reaches the wire. Each DoStream
	// call has its own client clone and its own variable, and redirects run
	// synchronously inside client.Do, so there is no sharing across calls.
	var credKeys []string
	requesterAttempt := 0
	route := RateLimitRoute{}
	costHeader := ""
	strictWrite := r.DisableRetries && !isSafeReplayableRead(method)
	baseClient := r.streamClient(base, opts, &credKeys)
	if strictWrite {
		baseClient = noReplayClient(baseClient)
	}
	client := r.clientWithRateLimitAdmission(baseClient, &requesterAttempt, &route, &costHeader)

	attempts := r.maxRetries() + 1
	var lastErr error
	for attempt := 0; attempt < attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		req, err := http.NewRequestWithContext(ctx, method, fullURL, nil)
		if err != nil {
			return nil, fmt.Errorf("build request: %w", err)
		}
		r.applyHeaders(req, false, "")
		if opts.Accept != "" {
			if strings.ContainsAny(opts.Accept, "\r\n") {
				return nil, fmt.Errorf("invalid Accept header value")
			}
			req.Header.Set("Accept", opts.Accept)
		}
		// Snapshot the header keys auth contributes, so the redirect policy
		// can strip exactly those on a cross-origin hop without having to
		// know which scheme any given connector uses.
		before := headerKeySet(req.Header)
		if r.Auth != nil {
			if err := r.Auth.Apply(ctx, req); err != nil {
				return nil, fmt.Errorf("apply auth: %w", err)
			}
		}
		credKeys = credentialHeaderKeys(before, req.Header, r.DefaultHeaders, r.DefaultHeaderValues)
		if r.DisableRetries {
			disableTransportReplay(req, strictWrite)
		}
		if err := r.admitRequesterSend(ctx, req, &requesterAttempt, &route, &costHeader); err != nil {
			return nil, err
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("send request: %w", err)
			if isRateLimitAdmissionError(err) {
				return nil, lastErr
			}
			if attempt < attempts-1 && !isRedirectPolicyError(err) {
				if werr := r.sleep(ctx, r.backoff(attempt, RateLimitObservation{})); werr != nil {
					return nil, werr
				}
				continue
			}
			return nil, lastErr
		}
		observation := r.observeRateLimit(ctx, route, resp.StatusCode, resp.Header, costHeader)

		if r.shouldRetry(resp.StatusCode) && attempt < attempts-1 {
			// Discard this attempt's body entirely; nothing from a failed
			// attempt may reach the caller.
			body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
			_ = resp.Body.Close()
			lastErr = responseHTTPError(resp.StatusCode, fullURL, body, observation)
			if werr := r.sleep(ctx, r.backoff(attempt, observation)); werr != nil {
				return nil, werr
			}
			continue
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
			_ = resp.Body.Close()
			return nil, responseHTTPError(resp.StatusCode, fullURL, body, observation)
		}
		if !r.acceptsSuccessfulStatus(resp.StatusCode) {
			_ = resp.Body.Close()
			return nil, &UnexpectedStatusError{Status: resp.StatusCode}
		}

		return &StreamResponse{
			Status: resp.StatusCode,
			Header: resp.Header,
			Body:   resp.Body,
			URL:    resp.Request.URL.String(),
		}, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("request to %s failed after %d attempts", fullURL, attempts)
	}
	return nil, lastErr
}

// streamClient returns a shallow copy of the configured client carrying the
// redirect policy. The shared client is never mutated.
func (r *Requester) streamClient(base *url.URL, opts StreamOptions, credKeys *[]string) *http.Client {
	clone := *r.client()
	clone.Timeout = 0
	clone.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if opts.RedirectPolicy != nil {
			return checkRedirectPolicy(req, via, base, *opts.RedirectPolicy, credKeys)
		}
		if len(via) >= maxRedirects {
			return fmt.Errorf("stopped after %d redirects", maxRedirects)
		}
		sameOrigin := req.URL.Host == base.Host && req.URL.Scheme == base.Scheme
		if sameOrigin {
			return nil
		}
		if err := allowCrossOrigin(req.URL, base, opts); err != nil {
			return err
		}
		// Permitted, but the credential never travels off-origin.
		stripCredentialHeaders(req, *credKeys)
		return nil
	}
	return &clone
}

func (r *Requester) redirectClient(ctx context.Context, base *url.URL, policy *RedirectPolicy, credKeys *[]string) *http.Client {
	if policy == nil {
		return r.clientFor(ctx)
	}
	clone := *r.clientFor(ctx)
	prior := clone.CheckRedirect
	clone.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if err := checkRedirectPolicy(req, via, base, *policy, credKeys); err != nil {
			return err
		}
		if prior != nil {
			return prior(req, via)
		}
		return nil
	}
	return &clone
}

func checkRedirectPolicy(req *http.Request, via []*http.Request, base *url.URL, policy RedirectPolicy, credKeys *[]string) error {
	if policy.MaxHops <= 0 || len(via) > policy.MaxHops {
		return fmt.Errorf("redirect refused by declared policy")
	}
	if strings.EqualFold(base.Scheme, "https") && !strings.EqualFold(req.URL.Scheme, "https") {
		return fmt.Errorf("scheme downgrade to %q blocked (base origin %s://%s)", req.URL.Redacted(), base.Scheme, base.Host)
	}
	sameOrigin := req.URL.Host == base.Host && req.URL.Scheme == base.Scheme
	if sameOrigin {
		if !policy.AllowSameOrigin {
			return fmt.Errorf("redirect refused by declared policy")
		}
		return nil
	}
	if err := allowCrossOrigin(req.URL, base, StreamOptions{AllowedHosts: policy.AllowedHosts}); err != nil {
		return err
	}
	stripCredentialHeaders(req, *credKeys)
	return nil
}

// allowCrossOrigin fails closed: an unparseable or host-less target, a
// differing host, or a scheme downgrade is refused unless explicitly widened.
// The "cross-host" wording matches the engine's existing checkOrigin guard so
// callers and tests can match on the same substring.
func allowCrossOrigin(next, base *url.URL, opts StreamOptions) error {
	if next.Host == "" {
		return fmt.Errorf("cross-host redirect to %q has no host; rejecting (fail closed)", next.Redacted())
	}
	if opts.AllowCrossHost {
		return nil
	}
	for _, h := range opts.AllowedHosts {
		if strings.EqualFold(strings.TrimSpace(h), next.Host) {
			return nil
		}
	}
	if next.Host == base.Host && next.Scheme != base.Scheme {
		return fmt.Errorf("scheme downgrade to %q blocked (base origin %s://%s)", next.Redacted(), base.Scheme, base.Host)
	}
	return fmt.Errorf("cross-host redirect to %q blocked (base host %q); declare allow_cross_host or allowed_hosts to permit", next.Redacted(), base.Host)
}

// alwaysSensitiveHeaders are stripped cross-origin regardless of how they were
// set. Go already strips these three itself; re-stating them costs nothing and
// removes any dependence on that behavior staying true.
var alwaysSensitiveHeaders = []string{"Authorization", "Www-Authenticate", "Cookie", "Proxy-Authorization"}

func headerKeySet(h http.Header) map[string]bool {
	out := make(map[string]bool, len(h))
	for k := range h {
		out[k] = true
	}
	return out
}

// credentialHeaderKeys returns every header key that must not cross an origin
// boundary: those the Authenticator added, plus every configured default
// header (a connector may carry its API key there), plus the always-sensitive
// set.
func credentialHeaderKeys(before map[string]bool, after http.Header, defaults map[string]string, defaultValues http.Header) []string {
	seen := map[string]bool{}
	add := func(k string) {
		canonical := http.CanonicalHeaderKey(k)
		if canonical != "" {
			seen[canonical] = true
		}
	}
	for k := range after {
		if !before[k] {
			add(k)
		}
	}
	for k := range defaults {
		add(k)
	}
	for k := range defaultValues {
		add(k)
	}
	for _, k := range alwaysSensitiveHeaders {
		add(k)
	}
	out := make([]string, 0, len(seen))
	for k := range seen {
		out = append(out, k)
	}
	return out
}

// stripCredentialHeaders removes every credential-bearing header from a
// permitted cross-origin hop. Go copies the original request's headers onto
// the redirect request, so without this the connector credential would follow
// the user to a third-party host.
func stripCredentialHeaders(req *http.Request, keys []string) {
	for _, k := range keys {
		if k = strings.TrimSpace(k); k != "" {
			req.Header.Del(k)
		}
	}
	for _, k := range alwaysSensitiveHeaders {
		req.Header.Del(k)
	}
}

// isRedirectPolicyError reports whether err came from our CheckRedirect. Such
// a refusal is a policy decision, not a transient failure, so retrying it
// would only repeat the same rejection.
func isRedirectPolicyError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "cross-host redirect") ||
		strings.Contains(msg, "scheme downgrade") ||
		strings.Contains(msg, "stopped after") ||
		strings.Contains(msg, "redirect refused")
}
