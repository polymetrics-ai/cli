package engine

import (
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
)

// TestRedditResidualWritesAreExecutable pins the final nine-row completion
// contract at the loaded-bundle boundary. The two S3 lease actions are still
// ordinary typed write actions here; the Reddit WriteHook supplies only their
// provider-mandated second request.
func TestRedditResidualWritesAreExecutable(t *testing.T) {
	bundle := loadRedditBundle(t)

	wantPaths := map[string]string{
		"emoji_asset_upload_s3":  "/api/v1/{{ config.subreddit }}/emoji_asset_upload_s3.json",
		"upload_sr_img":          "/r/{{ config.subreddit }}/api/upload_sr_img",
		"widget_image_upload_s3": "/r/{{ config.subreddit }}/api/widget_image_upload_s3",
		"vote":                   "/api/vote",
	}
	actions := make(map[string]WriteAction, len(bundle.Writes))
	for _, action := range bundle.Writes {
		actions[action.Name] = action
	}
	for name, wantPath := range wantPaths {
		action, ok := actions[name]
		if !ok {
			t.Errorf("missing residual Reddit write action %q", name)
			continue
		}
		if action.Path != wantPath {
			t.Errorf("action %q path = %q, want %q", name, action.Path, wantPath)
		}
		if action.IsBatchable() {
			t.Errorf("action %q is batchable; residual interactive/file actions must stay one-record-only", name)
		}
	}

	vote, ok := actions["vote"]
	if !ok {
		return
	}
	if vote.Confirmation == nil || vote.Confirmation.Kind != connectors.ConfirmationKindDestructive {
		t.Errorf("vote confirmation = %#v, want typed destructive per-invocation confirmation", vote.Confirmation)
	}
	if !strings.Contains(vote.Risk, "votes must be cast by humans") {
		t.Errorf("vote risk must carry Reddit's human-only rule, got %q", vote.Risk)
	}

	for _, name := range []string{"subreddit_emoji_json", "subreddit_emoji_emoji_name", "subreddit_emoji_custom_size"} {
		action, ok := actions[name]
		if !ok {
			t.Errorf("missing existing emoji action %q", name)
			continue
		}
		if strings.Contains(action.Path, "record.subreddit") || !strings.Contains(action.Path, "config.subreddit") {
			t.Errorf("emoji action %q path = %q, want config-scoped subreddit", name, action.Path)
		}
	}

	if bundle.CLISurface == nil {
		t.Fatal("Reddit cli surface is missing")
	}
	for _, command := range bundle.CLISurface.Commands {
		if command.Path != "vote" {
			continue
		}
		if command.Intent != "reverse_etl" || command.Availability != "implemented" || command.Write != "vote" {
			t.Errorf("vote command = %+v, want implemented individual reverse_etl command for vote", command)
		}
		return
	}
	t.Error("missing direct human command path \"vote\"")
}

func TestRedditEnforcesItsDocumentedRateLimit(t *testing.T) {
	bundle := loadRedditBundle(t)
	if bundle.HTTP.RateLimit == nil || bundle.HTTP.RateLimit.RequestsPerMinute != 100 {
		t.Errorf("Reddit stream rate limit = %+v, want 100 requests per minute", bundle.HTTP.RateLimit)
	}
	if bundle.Metadata.RateLimit.RequestsPerMinute != 100 {
		t.Errorf("Reddit inspected rate limit = %+v, want 100 requests per minute", bundle.Metadata.RateLimit)
	}
}
