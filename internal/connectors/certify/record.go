// Tier-1 record/replay support (certification design §E): a tiny transport
// seam wrapping any http.RoundTripper so a live HTTP call can be captured to
// a sanitized "golden cassette" JSON file at record time, then served back
// deterministically at replay time (replay.go) with zero network access and
// zero secret-leak risk in CI. Cassettes are organized
// <dir>/<stage>/<seq>.json, matched by (method, path, per-stage sequence).
package certify

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// sensitiveHeaderNames lists header names whose VALUES are always scrubbed
// from a recorded cassette regardless of whether they were passed as a
// planted secret (certification design §E: "Authorization/Cookie/X-Api-Key-
// class headers"). Matching is case-insensitive (http.Header already
// canonicalizes keys, but callers may pass either case).
var sensitiveHeaderNames = map[string]bool{
	"Authorization": true,
	"Cookie":        true,
	"Set-Cookie":    true,
	"X-Api-Key":     true,
	"X-Auth-Token":  true,
}

const redactedPlaceholder = "***REDACTED***"

// cassetteRequest is the recorded shape of an HTTP request (certification
// design §E matcher key: method + path).
type cassetteRequest struct {
	Method string      `json:"method"`
	Path   string      `json:"path"`
	Query  string      `json:"query,omitempty"`
	Header http.Header `json:"header,omitempty"`
	Body   string      `json:"body,omitempty"`
}

// cassetteResponse is the recorded shape of an HTTP response.
type cassetteResponse struct {
	Status int         `json:"status"`
	Header http.Header `json:"header,omitempty"`
	Body   string      `json:"body"`
}

// cassetteEntry is one <dir>/<stage>/<seq>.json file's contents.
type cassetteEntry struct {
	Request  cassetteRequest  `json:"request"`
	Response cassetteResponse `json:"response"`
}

// RecordingTransport wraps an underlying http.RoundTripper, forwarding every
// request unmodified to the network and appending a sanitized cassette
// entry for later replay. Safe for concurrent use; Flush persists all
// entries recorded so far (call it once per stage, or at the end of a run).
type RecordingTransport struct {
	underlying http.RoundTripper
	dir        string
	stage      string
	secrets    []string

	mu      sync.Mutex
	entries []cassetteEntry
	seq     int
}

// RecordOption configures a RecordingTransport.
type RecordOption func(*RecordingTransport)

// WithRecordSecrets registers planted secret values (in addition to the
// fixed sensitive-header scrub) that must never appear verbatim in a
// persisted cassette entry (certification design §E: "secret values (exact,
// base64, URL-encoded)").
func WithRecordSecrets(secrets ...string) RecordOption {
	return func(rt *RecordingTransport) {
		rt.secrets = append(rt.secrets, secrets...)
	}
}

// NewRecordingTransport builds a RecordingTransport that forwards requests
// to underlying and stages sanitized cassette entries for stage under dir.
// underlying defaults to http.DefaultTransport when nil.
func NewRecordingTransport(underlying http.RoundTripper, dir, stage string, opts ...RecordOption) *RecordingTransport {
	if underlying == nil {
		underlying = http.DefaultTransport
	}
	rt := &RecordingTransport{underlying: underlying, dir: dir, stage: stage}
	for _, opt := range opts {
		opt(rt)
	}
	return rt
}

// RoundTrip forwards req to the underlying transport, returning its response
// to the caller UNMODIFIED (the caller must see the real live body), while
// separately staging a sanitized copy for Flush to persist.
func (rt *RecordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var reqBodyCopy []byte
	if req.Body != nil {
		var err error
		reqBodyCopy, err = io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("certify: read request body for recording: %w", err)
		}
		req.Body = io.NopCloser(strings.NewReader(string(reqBodyCopy)))
	}

	resp, err := rt.underlying.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	respBodyCopy, readErr := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if readErr != nil {
		return nil, fmt.Errorf("certify: read response body for recording: %w", readErr)
	}
	// Restore an unmutated copy of the body for the CALLER (sanitization
	// below must never affect what the real connector code sees).
	resp.Body = io.NopCloser(strings.NewReader(string(respBodyCopy)))

	entry := cassetteEntry{
		Request: cassetteRequest{
			Method: req.Method,
			Path:   req.URL.Path,
			Query:  req.URL.RawQuery,
			Header: sanitizeHeader(req.Header, rt.secrets),
			Body:   sanitizeText(string(reqBodyCopy), rt.secrets),
		},
		Response: cassetteResponse{
			Status: resp.StatusCode,
			Header: sanitizeHeader(resp.Header, rt.secrets),
			Body:   sanitizeText(string(respBodyCopy), rt.secrets),
		},
	}

	rt.mu.Lock()
	rt.entries = append(rt.entries, entry)
	rt.mu.Unlock()

	return resp, nil
}

// Flush persists every cassette entry recorded so far to
// <dir>/<stage>/<seq>.json (seq is a 1-based, zero-padded 4-digit counter),
// per certification design §E on-disk layout. Safe to call multiple times;
// subsequent calls only write entries recorded since the last Flush.
func (rt *RecordingTransport) Flush() error {
	rt.mu.Lock()
	pending := rt.entries
	rt.entries = nil
	startSeq := rt.seq
	rt.seq += len(pending)
	rt.mu.Unlock()

	if len(pending) == 0 {
		return nil
	}

	stageDir := filepath.Join(rt.dir, rt.stage)
	if err := os.MkdirAll(stageDir, 0o755); err != nil {
		return fmt.Errorf("certify: create cassette stage dir %s: %w", stageDir, err)
	}

	for i, entry := range pending {
		seq := startSeq + i + 1
		raw, err := json.MarshalIndent(entry, "", "  ")
		if err != nil {
			return fmt.Errorf("certify: marshal cassette entry: %w", err)
		}
		path := filepath.Join(stageDir, cassetteFileName(seq))
		if err := os.WriteFile(path, raw, 0o644); err != nil {
			return fmt.Errorf("certify: write cassette %s: %w", path, err)
		}
	}
	return nil
}

// cassetteFileName renders seq as the fixed "NNNN.json" format used by both
// RecordingTransport.Flush and ReplayTransport's loader.
func cassetteFileName(seq int) string {
	return fmt.Sprintf("%04d.json", seq)
}

// sanitizeHeader returns a copy of h with sensitive header values redacted
// and any planted secret value scrubbed from the remaining values.
func sanitizeHeader(h http.Header, secrets []string) http.Header {
	if len(h) == 0 {
		return nil
	}
	out := make(http.Header, len(h))
	for name, values := range h {
		if sensitiveHeaderNames[http.CanonicalHeaderKey(name)] {
			out[name] = []string{redactedPlaceholder}
			continue
		}
		cleaned := make([]string, len(values))
		for i, v := range values {
			cleaned[i] = sanitizeText(v, secrets)
		}
		out[name] = cleaned
	}
	return out
}

// sanitizeText scrubs every planted secret value's exact, base64, and
// URL-encoded forms from text (certification design §E), mirroring
// ScanForSecrets' detection surface so nothing recorded can leak what
// ScanForSecrets is later relied on to catch.
func sanitizeText(text string, secrets []string) string {
	for _, secret := range secrets {
		trimmed := strings.TrimSpace(secret)
		if trimmed == "" {
			continue
		}
		text = strings.ReplaceAll(text, secret, redactedPlaceholder)
		text = strings.ReplaceAll(text, base64.StdEncoding.EncodeToString([]byte(secret)), redactedPlaceholder)
		text = strings.ReplaceAll(text, base64.RawStdEncoding.EncodeToString([]byte(secret)), redactedPlaceholder)
		text = strings.ReplaceAll(text, url.QueryEscape(secret), redactedPlaceholder)
	}
	return text
}

// sortedCassetteSeqs is a small helper used by replay.go to list a stage
// directory's cassette files in sequence order.
func sortedCassetteSeqs(names []string) []string {
	sorted := append([]string(nil), names...)
	sort.Slice(sorted, func(i, j int) bool {
		return cassetteSeqNum(sorted[i]) < cassetteSeqNum(sorted[j])
	})
	return sorted
}

func cassetteSeqNum(name string) int {
	base := strings.TrimSuffix(filepath.Base(name), filepath.Ext(name))
	n, _ := strconv.Atoi(base)
	return n
}
