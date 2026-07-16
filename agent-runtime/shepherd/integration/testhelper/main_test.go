//go:build integration

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFakeGitHubRejectsNonCommentTargetsAndMalformedPayloads(t *testing.T) {
	state := filepath.Join(t.TempDir(), "comments.json")
	t.Setenv("SHEPHERD_INTEGRATION_FAKE_GITHUB_STATE", state)
	if err := runGitHub([]string{"api", "repos/polymetrics-ai/cli/pulls/391/merge", "--method", "PUT"}); err == nil {
		t.Fatal("merge endpoint was accepted")
	}
	for _, payload := range []string{
		`{"body":"one","body":"two"}`,
		`{"body":"one","extra":true}`,
		`{"body":"one"} trailing`,
	} {
		withStdin(t, payload, func() {
			if _, err := readCommentPayload(); err == nil {
				t.Fatalf("malformed payload was accepted: %s", payload)
			}
		})
	}
	comment := newBotComment(101, "existing")
	if err := writeGitHubComments(state, []githubComment{comment}); err != nil {
		t.Fatal(err)
	}
	withStdin(t, `{"body":"replacement"}`, func() {
		err := runGitHub([]string{"api", "repos/polymetrics-ai/cli/issues/comments/102",
			"--method", "PATCH", "--input", "-"})
		if err == nil {
			t.Fatal("wrong comment ID was accepted")
		}
	})
}

func withStdin(t *testing.T, content string, run func()) {
	t.Helper()
	file, err := os.CreateTemp(t.TempDir(), "stdin-*")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(content); err != nil {
		t.Fatal(err)
	}
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatal(err)
	}
	original := os.Stdin
	os.Stdin = file
	defer func() {
		os.Stdin = original
		_ = file.Close()
	}()
	run()
}
