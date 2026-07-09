package certify

import "testing"

func TestBinaryDownloadCandidateForGitHub(t *testing.T) {
	candidate, ok := binaryDownloadCandidateFor("github")
	if !ok {
		t.Fatal("binaryDownloadCandidateFor(github) ok = false, want true")
	}
	if candidate.StageName != "binary_download_sweep_release_download" {
		t.Fatalf("StageName = %q", candidate.StageName)
	}
	wantArgs := []string{"github", "release", "download", "--credential", sourceCredentialName, "--json"}
	if len(candidate.Args) != len(wantArgs) {
		t.Fatalf("Args len = %d, want %d: %v", len(candidate.Args), len(wantArgs), candidate.Args)
	}
	for i := range wantArgs {
		if candidate.Args[i] != wantArgs[i] {
			t.Fatalf("Args[%d] = %q, want %q; args=%v", i, candidate.Args[i], wantArgs[i], candidate.Args)
		}
	}
}

func TestBinaryDownloadCandidateForUnknownConnector(t *testing.T) {
	if candidate, ok := binaryDownloadCandidateFor("sample"); ok {
		t.Fatalf("binaryDownloadCandidateFor(sample) = %+v, true; want no candidate", candidate)
	}
}
