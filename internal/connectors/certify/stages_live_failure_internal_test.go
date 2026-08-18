package certify

import (
	"os"
	"strings"
	"testing"
)

func TestEffectiveCredentialConfigAddsGitHubBaseURL(t *testing.T) {
	got := effectiveCredentialConfig("github", map[string]string{"owner": "octo", "repo": "hello"})
	if got["base_url"] != "https://api.github.com" {
		t.Fatalf("base_url = %q, want GitHub default", got["base_url"])
	}
	if got["tier"] != "certification" {
		t.Fatalf("tier = %q, want definition-owned certification selector", got["tier"])
	}
	got["owner"] = "mutated"
	orig := map[string]string{"owner": "octo", "base_url": "https://github.example/api", "tier": "normal"}
	got = effectiveCredentialConfig("github", orig)
	if got["base_url"] != "https://github.example/api" || got["tier"] != "certification" || orig["owner"] != "octo" || orig["tier"] != "normal" {
		t.Fatalf("effective config = %+v orig=%+v", got, orig)
	}
	unknown := effectiveCredentialConfig("unknown-certification-connector", nil)
	if unknown["base_url"] != "" {
		t.Fatalf("unknown connector base_url = %q, want no invented default", unknown["base_url"])
	}
}

func TestLiveStreamUnavailableClassifiesGitHubUnavailableErrors(t *testing.T) {
	rc := &runContext{opts: Options{Connector: "github"}}
	cases := []CLIResult{
		{Kind: "Error", Stdout: `{"error":{"message":"github stream=code_scanning_alerts page=0: http 403 for https://api.github.com/repos/o/r/code-scanning/alerts: [redacted]"},"kind":"Error"}`},
		{Kind: "Error", Stderr: `GitHub stream=dependabot_alerts page=0: HTTP 404 for https://api.github.com/repos/o/r/dependabot/alerts`},
		{Kind: "Error", Stderr: `github stream=secret_scanning_alerts page=0: request failed with Status 403`},
		{Kind: "Error", Stdout: `{"error":{"message":"graphql errors: Your token has not been granted the required scopes to execute this query"},"kind":"Error"}`},
		{Kind: "Error", Stdout: `{"error":{"message":"resolve graphql variable \"number\": interpolate: unresolved key \"number\" in query"},"kind":"Error"}`},
	}
	for _, res := range cases {
		if !liveStreamUnavailable(rc, res) {
			t.Fatalf("liveStreamUnavailable(%q) = false, want true", res.Stdout)
		}
	}
	unknown := &runContext{opts: Options{Connector: "unknown-certification-connector"}}
	if liveStreamUnavailable(unknown, cases[0]) {
		t.Fatalf("liveStreamUnavailable unknown connector = true, want safe false")
	}
}

func TestSafeCLIErrorEnvelopeDiagnosticFingerprintsSecret(t *testing.T) {
	const token = "cert-canary-etl-diagnostic-3989"
	diagnostic, err := safeCLIErrorEnvelopeDiagnostic(CLIResult{Envelope: map[string]any{
		"error": map[string]any{
			"category": "internal",
			"code":     "provider_error",
			"message":  "GitHub rejected credential " + token + " because the fixture stream is unavailable",
		},
	}}, []string{token})
	if err != nil {
		t.Fatalf("safe CLI error diagnostic: %v", err)
	}
	if len(ScanForSecrets(diagnostic, []string{token})) != 0 {
		t.Fatal("safe CLI error diagnostic retained a planted credential")
	}
	for _, want := range []string{"category=\"internal\"", "code=\"provider_error\"", "fixture stream is unavailable", "{{pmcertfp:v1:"} {
		if !strings.Contains(diagnostic, want) {
			t.Fatalf("safe CLI error diagnostic omitted %q", want)
		}
	}
}

func TestAssertKindFingerprintsErrorEnvelopeDiagnostic(t *testing.T) {
	const token = "cert-canary-assert-kind-diagnostic-3989"
	rc := &runContext{harness: NewHarness(t.TempDir(), WithSecrets(token))}
	passed, diagnostic := assertKind(rc, "resume", CLIResult{
		ExitCode: 1,
		Kind:     "Error",
		Envelope: map[string]any{"error": map[string]any{
			"category": "internal",
			"code":     "provider_error",
			"message":  "resume rejected credential " + token + " while replaying the source stream",
		}},
	}, "ETLRun", 0)
	if passed {
		t.Fatal("assertKind accepted an Error envelope")
	}
	if len(ScanForSecrets(diagnostic, []string{token})) != 0 {
		t.Fatal("assertKind diagnostic retained a planted credential")
	}
	for _, want := range []string{"kind got=\"Error\"", "cli_error", "provider_error", "replaying the source stream", "{{pmcertfp:v1:"} {
		if !strings.Contains(diagnostic, want) {
			t.Fatalf("assertKind diagnostic omitted %q", want)
		}
	}
}

func TestFullRefreshOverwriteDedupedSkipsWithoutCursor(t *testing.T) {
	rc := &runContext{
		opts:               Options{Connector: "sample"},
		capturePath:        "capture.jsonl",
		catalogStreamSpecs: []streamSpec{{Name: "branches", PrimaryKey: "name"}},
		currentStream:      "branches",
	}
	rep := Report{}
	if err := stageFullRefreshOverwriteDeduped(rc, &rep); err != nil {
		t.Fatalf("stageFullRefreshOverwriteDeduped: %v", err)
	}
	if len(rep.Stages) != 2 {
		t.Fatalf("len(stages) = %d, want 2", len(rep.Stages))
	}
	for _, stage := range rep.Stages {
		if stage.Passed || !stringsHasPrefix(stage.Error, "skipped: stream has no cursor field") {
			t.Fatalf("stage = %+v, want cursor skip", stage)
		}
	}
}

func TestResumeSkipsWhenIncrementalDidNotProduceCursor(t *testing.T) {
	rc := &runContext{opts: Options{Connector: "sample"}}
	rep := Report{}
	if err := stageResume(rc, &rep); err != nil {
		t.Fatalf("stageResume: %v", err)
	}
	stage := rep.Stages[len(rep.Stages)-1]
	if stage.Name != "resume" {
		t.Fatalf("stage name = %q, want resume", stage.Name)
	}
	if stage.Passed || !stringsHasPrefix(stage.Error, "skipped:") {
		t.Fatalf("resume stage = %+v, want documented skip", stage)
	}
}

func TestQueryContractSkipsWhenLiveReadProducedNoCapture(t *testing.T) {
	rc := &runContext{opts: Options{Connector: "github"}, currentStream: "code_scanning_alerts"}
	rep := Report{}
	if err := stageQueryContract(rc, &rep); err != nil {
		t.Fatalf("stageQueryContract: %v", err)
	}
	stage := rep.Stages[len(rep.Stages)-1]
	if stage.Name != "query_contract" || stage.Passed || !stringsHasPrefix(stage.Error, "skipped: no live capture") {
		t.Fatalf("query stage = %+v, want documented no-capture skip", stage)
	}
}

func TestFlowRoundtripSkipsWhenLiveCaptureIsEmpty(t *testing.T) {
	capturePath := t.TempDir() + "/empty.jsonl"
	if err := os.WriteFile(capturePath, nil, 0o600); err != nil {
		t.Fatalf("write empty capture: %v", err)
	}
	rc := &runContext{opts: Options{Connector: "github"}, capturePath: capturePath, currentStream: "pull_requests"}
	rep := Report{}
	if err := stageFlowRoundtrip(rc, &rep); err != nil {
		t.Fatalf("stageFlowRoundtrip: %v", err)
	}
	stage := rep.Stages[len(rep.Stages)-1]
	if stage.Name != "flow_roundtrip" || stage.Passed || !stringsHasPrefix(stage.Error, "skipped: live capture is empty") {
		t.Fatalf("flow stage = %+v, want documented empty-capture skip", stage)
	}
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
