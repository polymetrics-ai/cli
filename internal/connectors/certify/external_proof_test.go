package certify_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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
		BuildSHA256:    strings.Repeat("b", 64),
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
	imported, err := certify.ReadExternalProof(root, proofPath)
	if err != nil {
		t.Fatalf("read external proof: %v", err)
	}
	if got, want := imported.PMBuildSHA256, input.BuildSHA256; got != want {
		t.Fatalf("proof build provenance = %q, want %q", got, want)
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

func TestReadExternalProofRefusesAnUnfingerprintedResponseRegression(t *testing.T) {
	const canary = "cert-read-external-proof-canary"
	root := t.TempDir()
	proofPath, err := certify.WriteExternalProof(root, certify.ExternalProofInput{
		Connector: "sample", RunID: "external-proof-read-3989", BinarySHA256: strings.Repeat("a", 64),
		Command: []string{"pm", "connectors", "certify", "sample", "--from-env", "token=PM_CERT_TOKEN"},
		Stdout:  "completed", ExitCode: 0, Passed: true, FullParity: true, PreparedValues: []string{canary},
		HTTPExchanges: []certify.ObservedHTTPExchange{{
			Request: certify.ObservedHTTPRequest{Method: http.MethodGet, Target: "https://api.example.test/items?token=" + canary,
				Headers: http.Header{"Authorization": {"Bearer " + canary}}, Body: certify.ObservedBody{Complete: true}},
			Response: certify.ObservedHTTPResponse{Status: http.StatusOK,
				Body: certify.ObservedBody{Bytes: []byte(`{"account":"` + canary + `"}`), OriginalBytes: len(`{"account":"` + canary + `"}`), Complete: true}},
		}},
		FlowRoundTripReferences: []string{"flow_plan", "flow_preview", "flow_run", "flow_status"},
	})
	if err != nil {
		t.Fatalf("WriteExternalProof() = %v", err)
	}
	if _, err := certify.ReadExternalProof(root, proofPath); err != nil {
		t.Fatalf("ReadExternalProof() = %v", err)
	}
	raw, err := os.ReadFile(proofPath)
	if err != nil {
		t.Fatal(err)
	}
	var artifact map[string]any
	if err := json.Unmarshal(raw, &artifact); err != nil {
		t.Fatal(err)
	}
	exchanges := artifact["http_exchanges"].([]any)
	response := exchanges[0].(map[string]any)["response"].(map[string]any)
	response["body"].(map[string]any)["value"] = map[string]any{"account": canary}
	corrupted, err := json.Marshal(artifact)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(proofPath, corrupted, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := certify.ReadExternalProof(root, proofPath); err == nil || !strings.Contains(err.Error(), "unredacted") {
		t.Fatalf("ReadExternalProof() error = %v, want an unredacted response rejection", err)
	}
}

func TestWriteExternalProofPublishesBoundedObservedOperations(t *testing.T) {
	const credential = "cert-canary-bounded-scope-3989"
	root := t.TempDir()

	proofPath, err := certify.WriteExternalProof(root, certify.ExternalProofInput{
		Connector:      "sample",
		RunID:          "bounded-observed-operations-3989",
		BinarySHA256:   strings.Repeat("b", 64),
		Command:        []string{"pm", "connectors", "certify", "sample", "--from-env", "token=PM_CERT_TOKEN"},
		Stdout:         "completed provider observation before certification failure\n",
		ExitCode:       1,
		Passed:         false,
		FullParity:     false,
		PreparedValues: []string{credential},
		HTTPExchanges: []certify.ObservedHTTPExchange{{
			Request: certify.ObservedHTTPRequest{
				Method:  http.MethodGet,
				Target:  "https://provider.example.test/records",
				Headers: map[string][]string{"Authorization": {"Bearer " + credential}},
				Body:    certify.ObservedBody{Complete: true},
			},
			Response: certify.ObservedHTTPResponse{
				Status: http.StatusOK,
				Body:   certify.ObservedBody{Bytes: []byte(`{"result":"authenticated"}`), OriginalBytes: len(`{"result":"authenticated"}`), Complete: true},
			},
		}},
	})
	if err != nil {
		t.Fatalf("write bounded external proof: %v", err)
	}
	raw, err := os.ReadFile(proofPath)
	if err != nil {
		t.Fatalf("read bounded external proof: %v", err)
	}
	if bytes.Contains(raw, []byte(credential)) {
		t.Fatal("bounded external proof retained a planted credential")
	}
	var proof struct {
		Version                 int      `json:"version"`
		CredentialScope         string   `json:"credential_scope"`
		CredentialScopeProof    string   `json:"credential_scope_proof"`
		FlowRoundTripReferences []string `json:"flow_round_trip_references"`
		Process                 struct {
			ExitCode int `json:"exit_code"`
		} `json:"process"`
	}
	if err := json.Unmarshal(raw, &proof); err != nil {
		t.Fatalf("parse bounded external proof: %v", err)
	}
	if proof.Version != 2 || proof.CredentialScope != "observed_operations" || proof.CredentialScopeProof != "protocol_exchanges" {
		t.Fatalf("bounded external proof claim = version:%d scope:%q proof:%q, want schema-v2 observed operations", proof.Version, proof.CredentialScope, proof.CredentialScopeProof)
	}
	if proof.Process.ExitCode != 1 {
		t.Fatalf("bounded external proof exit code = %d, want the completed certification failure", proof.Process.ExitCode)
	}
	if len(proof.FlowRoundTripReferences) != 0 {
		t.Fatalf("bounded observed-operations proof retained full-parity flow references: %v", proof.FlowRoundTripReferences)
	}
}

func TestWriteExternalProofRefusesBoundedClaimWithoutSuccessfulProviderResponse(t *testing.T) {
	const credential = "cert-canary-no-provider-success-3989"
	root := t.TempDir()

	_, err := certify.WriteExternalProof(root, certify.ExternalProofInput{
		Connector:      "sample",
		RunID:          "no-provider-success-3989",
		BinarySHA256:   strings.Repeat("c", 64),
		Command:        []string{"pm", "connectors", "certify", "sample", "--from-env", "token=PM_CERT_TOKEN"},
		Stdout:         "provider returned an authorization failure\n",
		ExitCode:       1,
		PreparedValues: []string{credential},
		HTTPExchanges: []certify.ObservedHTTPExchange{{
			Request: certify.ObservedHTTPRequest{
				Method:  http.MethodGet,
				Target:  "https://provider.example.test/records",
				Headers: map[string][]string{"Authorization": {"Bearer " + credential}},
				Body:    certify.ObservedBody{Complete: true},
			},
			Response: certify.ObservedHTTPResponse{
				Status: http.StatusUnauthorized,
				Body:   certify.ObservedBody{Bytes: []byte(`{"error":"unauthorized"}`), OriginalBytes: len(`{"error":"unauthorized"}`), Complete: true},
			},
		}},
	})
	if err == nil || !strings.Contains(err.Error(), "observed successful provider response") {
		t.Fatalf("write bounded proof error = %v, want missing successful provider response", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, ".polymetrics", "certifications")); !os.IsNotExist(statErr) {
		t.Fatalf("refused bounded proof wrote certification material: stat error = %v", statErr)
	}
}

func TestWriteExternalProofFingerprintsOpaqueBodiesAndSeparatesCredentials(t *testing.T) {
	const credentialA = "cert-canary-opaque-a-3989"
	const credentialB = "cert-canary-opaque-b-3989"
	root := t.TempDir()

	proofPath, err := certify.WriteExternalProof(root, certify.ExternalProofInput{
		Connector:      "sample",
		RunID:          "opaque-fingerprint-semantics-3989",
		BinarySHA256:   strings.Repeat("f", 64),
		Command:        []string{"pm", "connectors", "certify", "sample", "--from-env", "token=PM_CERT_TOKEN"},
		ExitCode:       0,
		Passed:         true,
		FullParity:     true,
		PreparedValues: []string{credentialA, credentialB},
		HTTPExchanges: []certify.ObservedHTTPExchange{{
			Request: certify.ObservedHTTPRequest{
				Method: http.MethodPost,
				Target: "https://proof.invalid/opaque",
				Body: certify.ObservedBody{
					Bytes:         []byte("opaque-request=" + credentialA + "&alternate=" + credentialB + "&repeat=" + credentialA),
					OriginalBytes: len("opaque-request=" + credentialA + "&alternate=" + credentialB + "&repeat=" + credentialA),
					Complete:      true,
				},
			},
			Response: certify.ObservedHTTPResponse{
				Status: http.StatusOK,
				Body: certify.ObservedBody{
					Bytes:         []byte("opaque-response=" + credentialB + "&echo=" + credentialA),
					OriginalBytes: len("opaque-response=" + credentialB + "&echo=" + credentialA),
					Complete:      true,
				},
			},
		}},
		FlowRoundTripReferences: []string{"flow_plan", "flow_preview", "flow_run", "flow_status"},
	})
	if err != nil {
		t.Fatalf("write opaque proof: %v", err)
	}
	raw, err := os.ReadFile(proofPath)
	if err != nil {
		t.Fatalf("read opaque proof: %v", err)
	}
	salt, err := os.ReadFile(filepath.Join(root, ".polymetrics", "certifications", ".external-proof-salt"))
	if err != nil {
		t.Fatalf("read root proof salt: %v", err)
	}

	assertOpaqueExternalProofFingerprintSemantics(t, raw, credentialA, credentialB, salt)
}

func assertOpaqueExternalProofFingerprintSemantics(t *testing.T, raw []byte, credentialA, credentialB string, salt []byte) {
	t.Helper()
	for _, forbidden := range [][]byte{[]byte(credentialA), []byte(credentialB), salt} {
		if bytes.Contains(raw, forbidden) {
			t.Fatal("opaque proof retained raw credential material or its repository salt")
		}
	}

	var proof struct {
		CredentialFingerprints []string `json:"credential_fingerprints"`
		HTTPExchanges          []struct {
			Request struct {
				Body struct {
					Encoding string `json:"encoding"`
					Value    string `json:"value"`
				} `json:"body"`
			} `json:"request"`
			Response struct {
				Body struct {
					Encoding string `json:"encoding"`
					Value    string `json:"value"`
				} `json:"body"`
			} `json:"response"`
		} `json:"http_exchanges"`
	}
	if err := json.Unmarshal(raw, &proof); err != nil {
		t.Fatalf("parse opaque proof: %v", err)
	}
	if len(proof.HTTPExchanges) != 1 {
		t.Fatalf("opaque proof exchanges = %d, want 1", len(proof.HTTPExchanges))
	}

	fingerprintA := externalProofTestFingerprint(salt, credentialA)
	fingerprintB := externalProofTestFingerprint(salt, credentialB)
	if fingerprintA == fingerprintB {
		t.Fatal("distinct credential values produced one fingerprint")
	}
	fingerprints := make(map[string]bool, len(proof.CredentialFingerprints))
	for _, fingerprint := range proof.CredentialFingerprints {
		fingerprints[fingerprint] = true
	}
	if len(fingerprints) != 2 || !fingerprints[fingerprintA] || !fingerprints[fingerprintB] {
		t.Fatalf("credential fingerprints do not retain distinct A and B markers: %v", proof.CredentialFingerprints)
	}

	exchange := proof.HTTPExchanges[0]
	if exchange.Request.Body.Encoding != "opaque" || exchange.Response.Body.Encoding != "opaque" {
		t.Fatalf("opaque body encodings = request:%q response:%q, want opaque/opaque", exchange.Request.Body.Encoding, exchange.Response.Body.Encoding)
	}
	if got := strings.Count(exchange.Request.Body.Value, fingerprintA); got != 2 {
		t.Fatalf("opaque request repeated credential A markers = %d, want 2", got)
	}
	if got := strings.Count(exchange.Request.Body.Value, fingerprintB); got != 1 {
		t.Fatalf("opaque request credential B markers = %d, want 1", got)
	}
	if got := strings.Count(exchange.Response.Body.Value, fingerprintA); got != 1 {
		t.Fatalf("opaque response credential A markers = %d, want 1", got)
	}
	if got := strings.Count(exchange.Response.Body.Value, fingerprintB); got != 1 {
		t.Fatalf("opaque response credential B markers = %d, want 1", got)
	}
}

func externalProofTestFingerprint(salt []byte, value string) string {
	mac := hmac.New(sha256.New, salt)
	_, _ = mac.Write([]byte(value))
	return "{{pmcertfp:v1:" + hex.EncodeToString(mac.Sum(nil)) + "}}"
}

func TestRedactExternalProofDiagnosticFingerprintsSecrets(t *testing.T) {
	const credential = "cert-canary-diagnostic-3989"
	diagnostic, err := certify.RedactExternalProofDiagnostic(
		"github provider rejected credential="+credential+
			" encoded="+base64.StdEncoding.EncodeToString([]byte(credential))+
			" query="+url.QueryEscape(credential)+
			" because repository visibility is restricted",
		[]string{credential},
	)
	if err != nil {
		t.Fatalf("redact external-proof diagnostic: %v", err)
	}
	if len(certify.ScanForSecrets(diagnostic, []string{credential})) != 0 {
		t.Fatal("fingerprinted diagnostic retained a planted secret form")
	}
	if !strings.Contains(diagnostic, "github provider rejected") || !strings.Contains(diagnostic, "repository visibility is restricted") {
		t.Fatal("fingerprinted diagnostic did not retain its readable provider reason")
	}
	if strings.Count(diagnostic, "{{pmcertfp:v1:") != 3 {
		t.Fatal("fingerprinted diagnostic did not replace every planted secret form")
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

func TestWriteExternalProofAcceptsObservedOperationsWithoutFullParity(t *testing.T) {
	root := t.TempDir()
	proofPath, err := certify.WriteExternalProof(root, certify.ExternalProofInput{
		Connector:      "sample",
		RunID:          "observed-operations-4215",
		BinarySHA256:   strings.Repeat("f", 64),
		Command:        []string{"pm", "connectors", "certify", "sample", "--direct-read-only"},
		ExitCode:       0,
		Passed:         true,
		FullParity:     false,
		PreparedValues: []string{"cert-canary"},
		HTTPExchanges: []certify.ObservedHTTPExchange{{
			Request:  certify.ObservedHTTPRequest{Method: http.MethodGet, Target: "https://proof.invalid/observed", Body: certify.ObservedBody{Complete: true}},
			Response: certify.ObservedHTTPResponse{Status: http.StatusOK, Body: certify.ObservedBody{Complete: true}},
		}},
	})
	if err != nil {
		t.Fatalf("write observed-operations proof: %v", err)
	}
	if _, err := certify.ReadExternalProof(root, proofPath); err != nil {
		t.Fatalf("read observed-operations proof: %v", err)
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
