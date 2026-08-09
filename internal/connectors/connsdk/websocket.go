package connsdk

import (
	"context"
	cryptorand "crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"

	"polymetrics.ai/internal/safety"
)

const websocketAcceptGUID = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"

// WebSocketUpgrade is the authenticated, protocol-verified byte stream that
// remains after one declaration-owned HTTP WebSocket upgrade. It is not a
// generic URL client: callers supply a connector-relative path and exactly one
// declared subprotocol through OpenWebSocket.
type WebSocketUpgrade struct {
	Conn   io.ReadWriteCloser
	Header http.Header
}

// OpenWebSocket opens one authenticated RFC 6455 upgrade using the Requester's
// existing configured origin, headers, auth, and rate-limit admission. It
// intentionally accepts neither an absolute URL, redirect policy, arbitrary
// request headers, nor a caller-controlled HTTP method. The engine owns the
// higher-level frame contract and must close Conn synchronously.
func (r *Requester) OpenWebSocket(ctx context.Context, path, subprotocol string) (*WebSocketUpgrade, error) {
	if r == nil {
		return nil, fmt.Errorf("websocket requester is unavailable")
	}
	if err := validateWebSocketUpgradePath(path); err != nil {
		return nil, err
	}
	if err := validateWebSocketSubprotocol(subprotocol); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	fullURL, err := r.resolveURL(path, nil)
	if err != nil {
		return nil, fmt.Errorf("resolve websocket upgrade URL: %w", err)
	}
	key, err := newWebSocketKey()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build websocket upgrade request: %w", err)
	}
	r.applyHeaders(req, false, "")
	if r.Auth != nil {
		if err := r.Auth.Apply(ctx, req); err != nil {
			return nil, fmt.Errorf("apply websocket upgrade auth: %w", err)
		}
	}
	// Apply fixed handshake headers after default/auth headers so a connector
	// declaration cannot weaken the RFC 6455 protocol boundary through a
	// duplicated generic HTTP header.
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", key)
	req.Header.Set("Sec-WebSocket-Protocol", subprotocol)

	requesterAttempt := 0
	clientBase := *r.client()
	clientBase.Timeout = 0 // the caller context owns an upgraded session lifetime
	clientBase.CheckRedirect = func(*http.Request, []*http.Request) error {
		// WebSocket redirects can silently move authentication to an unrelated
		// host. A websocket_session declaration has no redirect field, so every
		// redirect is a terminal non-101 response rather than a second request.
		return http.ErrUseLastResponse
	}
	client := r.clientWithRateLimitAdmission(&clientBase, &requesterAttempt)
	if err := r.admitRequesterSend(ctx, http.MethodGet, &requesterAttempt); err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		if isRateLimitAdmissionError(err) {
			return nil, fmt.Errorf("websocket upgrade admission: %w", err)
		}
		// req.URL can carry an API-key query authenticator. Do not return the
		// transport's raw request URL/error text to a terminal caller.
		return nil, fmt.Errorf("send websocket upgrade request: %s", safetyRedactError(err))
	}
	observation := r.observeRateLimit(ctx, resp.StatusCode, resp.Header, requesterAttempt)
	if resp.StatusCode != http.StatusSwitchingProtocols {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		_ = resp.Body.Close()
		if resp.StatusCode >= 400 {
			return nil, responseHTTPError(resp.StatusCode, fullURL, body, observation)
		}
		return nil, fmt.Errorf("websocket upgrade expected HTTP 101, got %d", resp.StatusCode)
	}
	if !hasHTTPToken(resp.Header.Get("Connection"), "upgrade") || !strings.EqualFold(strings.TrimSpace(resp.Header.Get("Upgrade")), "websocket") {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("websocket upgrade response is missing required Upgrade headers")
	}
	if got := strings.TrimSpace(resp.Header.Get("Sec-WebSocket-Accept")); got != websocketAcceptValue(key) {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("websocket upgrade response has an invalid Sec-WebSocket-Accept")
	}
	if values := resp.Header.Values("Sec-WebSocket-Protocol"); len(values) != 1 || strings.TrimSpace(values[0]) != subprotocol {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("websocket upgrade response did not select declared subprotocol")
	}
	conn, ok := resp.Body.(io.ReadWriteCloser)
	if !ok {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("websocket upgrade response does not expose a read-write connection")
	}
	return &WebSocketUpgrade{Conn: conn, Header: resp.Header.Clone()}, nil
}

func validateWebSocketUpgradePath(path string) error {
	trimmed := strings.TrimSpace(path)
	if trimmed != path || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.ContainsAny(path, "\r\n?#") {
		return fmt.Errorf("websocket upgrade path must be a rooted connector-relative path")
	}
	return nil
}

func validateWebSocketSubprotocol(value string) error {
	if strings.TrimSpace(value) != value || value == "" {
		return fmt.Errorf("websocket upgrade requires one valid subprotocol")
	}
	for _, char := range value {
		if !isWebSocketTokenChar(char) {
			return fmt.Errorf("websocket upgrade requires one valid subprotocol")
		}
	}
	return nil
}

func isWebSocketTokenChar(char rune) bool {
	if char >= 'a' && char <= 'z' || char >= 'A' && char <= 'Z' || char >= '0' && char <= '9' {
		return true
	}
	switch char {
	case '!', '#', '$', '%', '&', '\'', '*', '+', '-', '.', '^', '_', '`', '|', '~':
		return true
	default:
		return false
	}
}

func newWebSocketKey() (string, error) {
	var nonce [16]byte
	if _, err := cryptorand.Read(nonce[:]); err != nil {
		return "", fmt.Errorf("generate websocket handshake key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(nonce[:]), nil
}

func websocketAcceptValue(key string) string {
	sum := sha1.Sum([]byte(key + websocketAcceptGUID))
	return base64.StdEncoding.EncodeToString(sum[:])
}

func hasHTTPToken(value, want string) bool {
	for _, candidate := range strings.Split(value, ",") {
		if strings.EqualFold(strings.TrimSpace(candidate), want) {
			return true
		}
	}
	return false
}

func safetyRedactError(err error) string {
	if err == nil {
		return "websocket transport error"
	}
	return safety.RedactErrorText(err.Error())
}
