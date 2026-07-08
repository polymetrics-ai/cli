package certify

import "testing"

func TestEffectiveCredentialConfigAddsGitHubBaseURL(t *testing.T) {
	got := effectiveCredentialConfig("github", map[string]string{"owner": "octo", "repo": "hello"})
	if got["base_url"] != "https://api.github.com" {
		t.Fatalf("base_url = %q, want GitHub default", got["base_url"])
	}
	got["owner"] = "mutated"
	orig := map[string]string{"owner": "octo", "base_url": "https://github.example/api"}
	got = effectiveCredentialConfig("github", orig)
	if got["base_url"] != "https://github.example/api" || orig["owner"] != "octo" {
		t.Fatalf("effective config = %+v orig=%+v", got, orig)
	}
}

func TestLiveStreamUnavailableClassifiesGitHubUnavailableErrors(t *testing.T) {
	rc := &runContext{opts: Options{Connector: "github"}}
	cases := []CLIResult{
		{Kind: "Error", Stdout: `{"error":{"message":"github stream=code_scanning_alerts page=0: http 403 for https://api.github.com/repos/o/r/code-scanning/alerts: [redacted]"},"kind":"Error"}`},
		{Kind: "Error", Stdout: `{"error":{"message":"graphql errors: Your token has not been granted the required scopes to execute this query"},"kind":"Error"}`},
		{Kind: "Error", Stdout: `{"error":{"message":"resolve graphql variable \"number\": interpolate: unresolved key \"number\" in query"},"kind":"Error"}`},
	}
	for _, res := range cases {
		if !liveStreamUnavailable(rc, res) {
			t.Fatalf("liveStreamUnavailable(%q) = false, want true", res.Stdout)
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

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
