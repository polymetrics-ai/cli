package certify

import (
	"testing"
)

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

func TestBinaryUploadCandidateForGitHub(t *testing.T) {
	candidate, ok := binaryUploadCandidateFor("github", map[string]string{
		"binary_upload_file_path":  "fixture.bin",
		"binary_upload_name":       "fixture.bin",
		"binary_upload_release_id": "42",
	})
	if !ok {
		t.Fatal("binaryUploadCandidateFor(github) ok = false, want true")
	}
	if candidate.StageName != "binary_upload_sweep_releases_assets_upload" || candidate.Command != "releases assets upload" {
		t.Fatalf("candidate = %+v, want GitHub release-asset upload candidate", candidate)
	}
	wantPrefix := []string{"github", "releases", "assets", "upload", "--credential", sourceCredentialName}
	if len(candidate.Args) < len(wantPrefix) {
		t.Fatalf("Args = %v, want prefix %v", candidate.Args, wantPrefix)
	}
	for i := range wantPrefix {
		if candidate.Args[i] != wantPrefix[i] {
			t.Fatalf("Args[%d] = %q, want %q; args=%v", i, candidate.Args[i], wantPrefix[i], candidate.Args)
		}
	}
	for _, want := range []string{"--file-path", "fixture.bin", "--name", "--release-id", "42"} {
		found := false
		for _, arg := range candidate.Args {
			if arg == want {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Args = %v, want configured candidate argument %q", candidate.Args, want)
		}
	}
}

// A command rejection is valuable safety evidence, but it cannot be recorded
// as an upload pass: no transfer, provider response, read-back, or cleanup has
// happened. This drives the real in-process CLI harness and pins the report
// roll-up as non-passing.
func TestStageBinaryUploadSweepRefusalIsNeverPass(t *testing.T) {
	rc := &runContext{
		harness: NewHarness(t.TempDir()),
		opts:    Options{Connector: "github", Full: true},
	}
	report := Report{}
	if err := stageBinaryUploadSweep(rc, &report); err != nil {
		t.Fatalf("stageBinaryUploadSweep: %v", err)
	}
	if report.Capabilities.BinaryUpload == nil || report.Capabilities.BinaryUpload.Result != "blocked" {
		t.Fatalf("binary upload capability = %+v, want blocked", report.Capabilities.BinaryUpload)
	}
	if len(report.Stages) != 1 || report.Stages[0].Passed || report.Stages[0].Status != "blocked" {
		t.Fatalf("binary upload stage = %+v, want one non-passing blocked stage", report.Stages)
	}
}
