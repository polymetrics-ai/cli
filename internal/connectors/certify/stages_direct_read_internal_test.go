package certify

import "testing"

func TestDirectReadCandidatesForGitHub(t *testing.T) {
	candidates := directReadCandidatesFor("github", map[string]string{
		"direct_read_path":     "docs/index.md",
		"direct_read_dir_path": "",
		"direct_read_ref":      "main",
	})
	if len(candidates) != 2 {
		t.Fatalf("len(candidates) = %d, want 2: %+v", len(candidates), candidates)
	}

	want := map[string][]string{
		"repo read-file": {"github", "repo", "read-file", "--credential", sourceCredentialName, "--path", "docs/index.md", "--ref", "main", "--max-bytes", "1048576", "--json"},
		"repo read-dir":  {"github", "repo", "read-dir", "--credential", sourceCredentialName, "--path", "", "--ref", "main", "--max-bytes", "1048576", "--json"},
	}
	for _, candidate := range candidates {
		wantArgs, ok := want[candidate.Command]
		if !ok {
			t.Fatalf("unexpected candidate command %q", candidate.Command)
		}
		if len(candidate.Args) != len(wantArgs) {
			t.Fatalf("%s Args len = %d, want %d: %v", candidate.Command, len(candidate.Args), len(wantArgs), candidate.Args)
		}
		for i := range wantArgs {
			if candidate.Args[i] != wantArgs[i] {
				t.Fatalf("%s Args[%d] = %q, want %q; args=%v", candidate.Command, i, candidate.Args[i], wantArgs[i], candidate.Args)
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
