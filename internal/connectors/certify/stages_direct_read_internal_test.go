package certify

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors/engine"
)

func TestDirectReadCandidatesForGitHub(t *testing.T) {
	candidates := directReadCandidatesFor("github", map[string]string{
		"direct_read_path":     "docs/index.md",
		"direct_read_dir_path": "",
		"direct_read_ref":      "main",
	})
	if len(candidates) != 120 {
		t.Fatalf("len(candidates) = %d, want 120 (23 manual plus 97 generated): %+v", len(candidates), candidates)
	}

	want := map[string][]string{
		"repo read-file": {"github", "repo", "read-file", "--credential", sourceCredentialName, "--path", "docs/index.md", "--ref", "main", "--max-bytes", "1048576", "--json"},
		"branches view":  {"github", "branches", "view", "--credential", sourceCredentialName, "--branch", "main", "--max-bytes", "1048576", "--json"},
		"git ref view":   {"github", "git", "ref", "view", "--credential", sourceCredentialName, "--ref", "heads/main", "--max-bytes", "1048576", "--json"},
		"commits view":   {"github", "commits", "view", "--credential", sourceCredentialName, "--ref", "main", "--max-bytes", "1048576", "--json"},
	}
	seen := make(map[string]bool, len(want))
	for _, candidate := range candidates {
		wantArgs, ok := want[candidate.Command]
		if len(candidate.OutputAssertions) == 0 {
			t.Fatalf("%s has no produced-value assertion", candidate.Command)
		}
		if !ok {
			continue
		}
		seen[candidate.Command] = true
		if len(candidate.Args) != len(wantArgs) {
			t.Fatalf("%s Args len = %d, want %d: %v", candidate.Command, len(candidate.Args), len(wantArgs), candidate.Args)
		}
		for i := range wantArgs {
			if candidate.Args[i] != wantArgs[i] {
				t.Fatalf("%s Args[%d] = %q, want %q; args=%v", candidate.Command, i, candidate.Args[i], wantArgs[i], candidate.Args)
			}
		}
	}
	for command := range want {
		if !seen[command] {
			t.Fatalf("did not find expected candidate %q", command)
		}
	}
}

func TestDirectReadCandidatesForGitHubTrialCohorts(t *testing.T) {
	wantCounts := map[string]int{
		"trial_advanced_security": 31,
		"trial_codespaces":        22,
		"trial_copilot":           23,
		"trial_enterprise":        21,
	}
	for cohort, want := range wantCounts {
		candidates := directReadCandidatesFor("github", map[string]string{"certification_cohort": cohort})
		if len(candidates) != want {
			t.Fatalf("%s candidates = %d, want %d", cohort, len(candidates), want)
		}
		for _, candidate := range candidates {
			if !strings.HasPrefix(candidate.StageName, "generated_direct_read_") {
				t.Fatalf("%s includes non-generated candidate %+v", cohort, candidate)
			}
			if len(candidate.OutputAssertions) != 1 || candidate.OutputAssertions[0].ValueType != "object_or_array" {
				t.Fatalf("%s candidate %q assertion = %#v, want generated object-or-array assertion", cohort, candidate.Command, candidate.OutputAssertions)
			}
		}
	}
}

func TestDirectReadCandidateForGitHub(t *testing.T) {
	candidate, ok := directReadCandidateFor("github", map[string]string{"direct_read_path": "docs/index.md", "direct_read_ref": "main"})
	if !ok {
		t.Fatal("directReadCandidateFor(github) ok = false, want true")
	}
	if candidate.Command != "repo read-file" {
		t.Fatalf("Command = %q, want repo read-file", candidate.Command)
	}
}

func TestDirectReadCandidateForUnknownConnector(t *testing.T) {
	if candidate, ok := directReadCandidateFor("sample", nil); ok {
		t.Fatalf("directReadCandidateFor(sample) = %+v, true; want no candidate", candidate)
	}
}

func TestDirectReadOutputAssertions(t *testing.T) {
	res := CLIResult{Envelope: map[string]any{
		"kind":     "ConnectorCommandDirectRead",
		"response": map[string]any{"name": "fixture-readme", "type": "file"},
	}}

	t.Run("matching produced value passes", func(t *testing.T) {
		passed, reason := assertDirectReadOutputAssertions("fixture_read", res, []engine.CertificationOutputAssertion{{
			JSONPointer: "/response/name",
			Equals:      "fixture-readme",
		}})
		if !passed || reason != "" {
			t.Fatalf("assertDirectReadOutputAssertions = %t, %q; want pass", passed, reason)
		}
	})

	t.Run("produced response type passes", func(t *testing.T) {
		passed, reason := assertDirectReadOutputAssertions("fixture_read", res, []engine.CertificationOutputAssertion{{
			JSONPointer: "/response",
			ValueType:   "object",
		}})
		if !passed || reason != "" {
			t.Fatalf("assertDirectReadOutputAssertions = %t, %q; want object response pass", passed, reason)
		}
	})

	t.Run("generated structural assertion accepts an object response", func(t *testing.T) {
		passed, reason := assertDirectReadOutputAssertions("generated_fixture_read", res, []engine.CertificationOutputAssertion{{
			JSONPointer: "/response",
			ValueType:   "object_or_array",
		}})
		if !passed || reason != "" {
			t.Fatalf("assertDirectReadOutputAssertions = %t, %q; want generated structural assertion pass", passed, reason)
		}
	})

	t.Run("generated structural assertion rejects scalar response", func(t *testing.T) {
		scalar := CLIResult{Envelope: map[string]any{
			"kind":     "ConnectorCommandDirectRead",
			"response": "not-a-produced-record-or-collection",
		}}
		passed, reason := assertDirectReadOutputAssertions("generated_fixture_read", scalar, []engine.CertificationOutputAssertion{{
			JSONPointer: "/response",
			ValueType:   "object_or_array",
		}})
		if passed || !strings.Contains(reason, "wrong type") {
			t.Fatalf("assertDirectReadOutputAssertions = %t, %q; want scalar generated assertion failure", passed, reason)
		}
	})

	t.Run("post-schema mismatch turns the direct-read assertion red", func(t *testing.T) {
		passed, reason := assertDirectReadOutputAssertions("fixture_read", res, []engine.CertificationOutputAssertion{{
			JSONPointer: "/response/name",
			Equals:      "impossible-after-schema-compile",
		}})
		if passed {
			t.Fatal("assertDirectReadOutputAssertions passed an impossible declared value")
		}
		if !strings.Contains(reason, "fixture_read") || !strings.Contains(reason, "/response/name") || !strings.Contains(reason, "does not match") {
			t.Fatalf("failure reason = %q, want stage and pointer without rendered values", reason)
		}
		if strings.Contains(reason, "impossible-after-schema-compile") || strings.Contains(reason, "fixture-readme") {
			t.Fatalf("failure reason rendered assertion values: %q", reason)
		}
	})
}
