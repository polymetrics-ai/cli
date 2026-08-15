package certify_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/certify"
)

func TestWriteExternalProofFingerprintsObservedExternalTranscript(t *testing.T) {
	const credentialCanary = "cert-canary-credential-3989"
	const responseCanary = "cert-canary-response-3989"

	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer "+credentialCanary {
			t.Fatalf("authorization = %q, want exact prepared credential", got)
		}
		w.Header().Set("X-Reply", responseCanary)
		_, _ = io.WriteString(w, `{"result":"`+responseCanary+`"}`)
	}))
	defer server.Close()

	transport := certify.NewObservedTransport(server.Client().Transport, 1024)
	request, err := http.NewRequest(http.MethodPost, server.URL+"/proof?token="+credentialCanary, strings.NewReader(`{"token":"`+credentialCanary+`"}`))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	request.Header.Set("Authorization", "Bearer "+credentialCanary)
	response, err := (&http.Client{Transport: transport}).Do(request)
	if err != nil {
		t.Fatalf("observed request: %v", err)
	}
	if _, err := io.ReadAll(response.Body); err != nil {
		t.Fatalf("read response: %v", err)
	}
	if err := response.Body.Close(); err != nil {
		t.Fatalf("close response: %v", err)
	}

	root := t.TempDir()
	input := certify.ExternalProofInput{
		Connector:      "sample",
		RunID:          "external-proof-3989",
		BinarySHA256:   strings.Repeat("a", 64),
		Command:        []string{"pm", "connectors", "certify", "sample", "--from-env", "token=PM_CERT_TOKEN"},
		Stdout:         "external child completed\n",
		Stderr:         "",
		ExitCode:       0,
		Passed:         true,
		PreparedValues: []string{credentialCanary},
		HTTPExchanges:  transport.Exchanges(),
		FullParity:     true,
		FlowRoundTripReferences: []string{
			"flow_plan", "flow_preview", "flow_run", "flow_status",
		},
	}
	proofPath, err := certify.WriteExternalProof(root, input)
	if err != nil {
		t.Fatalf("write external proof: %v", err)
	}
	raw, err := os.ReadFile(proofPath)
	if err != nil {
		t.Fatalf("read proof: %v", err)
	}
	for _, forbidden := range []string{credentialCanary, responseCanary, "Bearer " + credentialCanary} {
		if bytes.Contains(raw, []byte(forbidden)) {
			t.Fatalf("proof contains raw value %q: %s", forbidden, raw)
		}
	}
	if !bytes.Contains(raw, []byte("pmcertfp:v1:")) {
		t.Fatalf("proof has no fingerprint markers: %s", raw)
	}
	if !bytes.Contains(raw, []byte(`"command": [`)) || !bytes.Contains(raw, []byte(`"certify"`)) {
		t.Fatalf("proof did not retain exact safe process argv: %s", raw)
	}
	for _, reference := range []string{"flow_plan", "flow_preview", "flow_run", "flow_status"} {
		if !bytes.Contains(raw, []byte(`"`+reference+`"`)) {
			t.Fatalf("proof omitted flow read-back reference %q: %s", reference, raw)
		}
	}
	matched, err := certify.VerifyExternalProofTranscript(root, proofPath, "external child completed\n", "")
	if err != nil {
		t.Fatalf("verify observed external transcript: %v", err)
	}
	if !matched {
		t.Fatal("external proof does not match the exact observed process output")
	}
	matched, err = certify.VerifyExternalProofTranscript(root, proofPath, "different process output\n", "")
	if err != nil {
		t.Fatalf("verify mismatched external transcript: %v", err)
	}
	if matched {
		t.Fatal("different process output verified against captured transcript")
	}

	input.RunID = "external-proof-replay-3989"
	replayPath, err := certify.WriteExternalProof(root, input)
	if err != nil {
		t.Fatalf("write replay proof: %v", err)
	}
	replayRaw, err := os.ReadFile(replayPath)
	if err != nil {
		t.Fatalf("read replay proof: %v", err)
	}
	if got, want := credentialFingerprint(t, replayRaw), credentialFingerprint(t, raw); got != want {
		t.Fatalf("same-root credential fingerprint changed across replay: got %q want %q", got, want)
	}

	otherRoot := t.TempDir()
	input.RunID = "external-proof-other-salt-3989"
	otherPath, err := certify.WriteExternalProof(otherRoot, input)
	if err != nil {
		t.Fatalf("write distinct-root proof: %v", err)
	}
	otherRaw, err := os.ReadFile(otherPath)
	if err != nil {
		t.Fatalf("read distinct-root proof: %v", err)
	}
	if got, previous := credentialFingerprint(t, otherRaw), credentialFingerprint(t, raw); got == previous {
		t.Fatalf("distinct roots reused a credential fingerprint: %q", got)
	}
}

func TestWriteExternalProofRefusesTruncatedBodyWithoutArtifactWrites(t *testing.T) {
	transport := certify.NewObservedTransport(roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("opaque-response-body")),
		}, nil
	}), 4)
	request, err := http.NewRequest(http.MethodGet, "https://proof.invalid/opaque", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	response, err := (&http.Client{Transport: transport}).Do(request)
	if err != nil {
		t.Fatalf("observed request: %v", err)
	}
	if _, err := io.ReadAll(response.Body); err != nil {
		t.Fatalf("read response: %v", err)
	}
	_ = response.Body.Close()

	root := t.TempDir()
	_, err = certify.WriteExternalProof(root, certify.ExternalProofInput{
		Connector:      "sample",
		RunID:          "truncated-proof-3989",
		BinarySHA256:   strings.Repeat("b", 64),
		Command:        []string{"pm", "connectors", "certify", "sample"},
		ExitCode:       0,
		Passed:         true,
		PreparedValues: []string{"cert-canary"},
		HTTPExchanges:  transport.Exchanges(),
		FullParity:     true,
		FlowRoundTripReferences: []string{
			"flow_plan", "flow_preview", "flow_run", "flow_status",
		},
	})
	if err == nil || !strings.Contains(err.Error(), "truncated") {
		t.Fatalf("write truncated proof error = %v, want bounded-body refusal", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, ".polymetrics", "certifications")); !os.IsNotExist(statErr) {
		t.Fatalf("proof refusal created artifact material: stat error = %v, want not exist", statErr)
	}
}

func TestWriteExternalProofRefusesMissingFlowReferencesWithoutArtifactWrites(t *testing.T) {
	root := t.TempDir()
	_, err := certify.WriteExternalProof(root, certify.ExternalProofInput{
		Connector:      "sample",
		RunID:          "missing-flow-3989",
		BinarySHA256:   strings.Repeat("c", 64),
		Command:        []string{"pm", "connectors", "certify", "sample"},
		ExitCode:       0,
		Passed:         true,
		FullParity:     true,
		PreparedValues: []string{"cert-canary"},
		HTTPExchanges: []certify.ObservedHTTPExchange{{
			Request: certify.ObservedHTTPRequest{
				Method: http.MethodGet,
				Target: "https://proof.invalid/flow",
				Body:   certify.ObservedBody{Complete: true},
			},
			Response: certify.ObservedHTTPResponse{
				Status: http.StatusOK,
				Body:   certify.ObservedBody{Complete: true},
			},
		}},
		FlowRoundTripReferences: []string{"flow_plan", "flow_preview"},
	})
	if err == nil || !strings.Contains(err.Error(), "flow round-trip reference") {
		t.Fatalf("write proof without flow references error = %v, want refusal", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, ".polymetrics", "certifications")); !os.IsNotExist(statErr) {
		t.Fatalf("missing-flow refusal created artifact material: stat error = %v, want not exist", statErr)
	}
}

func TestWriteExternalProofRefusesExchangesBeyondRedirectRetryBoundWithoutWrites(t *testing.T) {
	root := t.TempDir()
	exchanges := make([]certify.ObservedHTTPExchange, 17)
	for index := range exchanges {
		exchanges[index] = certify.ObservedHTTPExchange{
			Request: certify.ObservedHTTPRequest{
				Method: http.MethodGet,
				Target: "https://proof.invalid/retry",
				Body:   certify.ObservedBody{Complete: true},
			},
			Response: certify.ObservedHTTPResponse{
				Status: http.StatusServiceUnavailable,
				Body:   certify.ObservedBody{Complete: true},
			},
		}
	}
	_, err := certify.WriteExternalProof(root, certify.ExternalProofInput{
		Connector:               "sample",
		RunID:                   "too-many-exchanges-3989",
		BinarySHA256:            strings.Repeat("d", 64),
		Command:                 []string{"pm", "connectors", "certify", "sample"},
		ExitCode:                0,
		Passed:                  true,
		FullParity:              true,
		PreparedValues:          []string{"cert-canary"},
		HTTPExchanges:           exchanges,
		FlowRoundTripReferences: []string{"flow_plan", "flow_preview", "flow_run", "flow_status"},
	})
	if err == nil || !strings.Contains(err.Error(), "exceeding bounded redirect/retry limit") {
		t.Fatalf("write proof above exchange bound error = %v, want bounded redirect/retry refusal", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, ".polymetrics", "certifications")); !os.IsNotExist(statErr) {
		t.Fatalf("exchange-bound refusal created artifact material: stat error = %v, want not exist", statErr)
	}
}

func TestWriteExternalProofAcceptsWholeSurfaceAbovePerRequestRetryBound(t *testing.T) {
	root := t.TempDir()
	exchanges := make([]certify.ObservedHTTPExchange, 17)
	for index := range exchanges {
		exchanges[index] = certify.ObservedHTTPExchange{
			Request: certify.ObservedHTTPRequest{
				Method: http.MethodGet,
				Target: fmt.Sprintf("https://proof.invalid/surface/%d", index),
				Body:   certify.ObservedBody{Complete: true},
			},
			Response: certify.ObservedHTTPResponse{
				Status: http.StatusOK,
				Body:   certify.ObservedBody{Complete: true},
			},
		}
	}
	proofPath, err := certify.WriteExternalProof(root, certify.ExternalProofInput{
		Connector:               "sample",
		RunID:                   "whole-surface-3989",
		BinarySHA256:            strings.Repeat("e", 64),
		Command:                 []string{"pm", "connectors", "certify", "sample"},
		ExitCode:                0,
		Passed:                  true,
		FullParity:              true,
		PreparedValues:          []string{"cert-canary"},
		HTTPExchanges:           exchanges,
		FlowRoundTripReferences: []string{"flow_plan", "flow_preview", "flow_run", "flow_status"},
	})
	if err != nil {
		t.Fatalf("write whole-surface proof: %v", err)
	}
	if _, err := os.Stat(proofPath); err != nil {
		t.Fatalf("stat whole-surface proof: %v", err)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (fn roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func credentialFingerprint(t *testing.T, raw []byte) string {
	t.Helper()
	var proof struct {
		CredentialFingerprints []string `json:"credential_fingerprints"`
	}
	if err := json.Unmarshal(raw, &proof); err != nil {
		t.Fatalf("parse proof fingerprints: %v", err)
	}
	if len(proof.CredentialFingerprints) != 1 {
		t.Fatalf("credential fingerprints = %#v, want one", proof.CredentialFingerprints)
	}
	return proof.CredentialFingerprints[0]
}
