package connectors

import "testing"

func TestWave1ParentSyncRuntimeConfigPreservesBothPolicies(t *testing.T) {
	cfg := RuntimeConfig{
		ProjectDir: "/tmp/project",
		LocalWritePolicy: &LocalWritePolicy{
			ProjectRoot:   "/tmp/project",
			AllowExternal: false,
		},
		ApprovedPayloadSHA256: map[string]string{
			PayloadApprovalKey(2, "media_file_path"): "approved-digest",
		},
	}

	if cfg.LocalWritePolicy == nil || cfg.LocalWritePolicy.ProjectRoot != cfg.ProjectDir {
		t.Fatalf("local write policy = %+v, project dir = %q", cfg.LocalWritePolicy, cfg.ProjectDir)
	}
	if got := cfg.ApprovedPayloadSHA256[PayloadApprovalKey(2, "media_file_path")]; got != "approved-digest" {
		t.Fatalf("approved payload digest = %q, want approved-digest", got)
	}
}
