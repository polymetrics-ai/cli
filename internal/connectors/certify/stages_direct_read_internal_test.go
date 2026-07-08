package certify

import "testing"

func TestDirectReadCandidateForGitHub(t *testing.T) {
	candidate, ok := directReadCandidateFor("github", map[string]string{"direct_read_path": "docs/index.md", "direct_read_ref": "main"})
	if !ok {
		t.Fatal("directReadCandidateFor(github) ok = false, want true")
	}
	if candidate.StageName != "direct_read_sweep_repo_read_file" {
		t.Fatalf("StageName = %q", candidate.StageName)
	}
	wantArgs := []string{"github", "repo", "read-file", "--credential", sourceCredentialName, "--path", "docs/index.md", "--ref", "main", "--max-bytes", "1048576", "--json"}
	if len(candidate.Args) != len(wantArgs) {
		t.Fatalf("Args len = %d, want %d: %v", len(candidate.Args), len(wantArgs), candidate.Args)
	}
	for i := range wantArgs {
		if candidate.Args[i] != wantArgs[i] {
			t.Fatalf("Args[%d] = %q, want %q; args=%v", i, candidate.Args[i], wantArgs[i], candidate.Args)
		}
	}
}

func TestDirectReadCandidateForUnknownConnector(t *testing.T) {
	if candidate, ok := directReadCandidateFor("sample", nil); ok {
		t.Fatalf("directReadCandidateFor(sample) = %+v, true; want no candidate", candidate)
	}
}
