package cli

import (
	"bytes"
	"testing"
)

func TestAgentImage_MissingSubcommand(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", dir, "agent", "image"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("want non-zero exit for missing subcommand")
	}
}

func TestAgentImage_PodmanAbsent(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)
	// Point podman at a binary that does not exist → clear, actionable error.
	t.Setenv("POLYMETRICS_PODMAN_BIN", "definitely-not-a-real-binary-xyz")
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", dir, "agent", "image", "build"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("want non-zero exit when podman is absent")
	}
	if !bytes.Contains(stderr.Bytes(), []byte("podman")) && !bytes.Contains(stdout.Bytes(), []byte("podman")) {
		t.Errorf("error should mention podman; got stderr=%s", stderr.String())
	}
}

func TestAgentImage_UnknownSubcommand(t *testing.T) {
	dir := t.TempDir()
	initProject(t, dir)
	t.Setenv("POLYMETRICS_PODMAN_BIN", "sh") // exists, so we reach the switch
	var stdout, stderr bytes.Buffer
	code := Run([]string{"--root", dir, "agent", "image", "frobnicate"}, &stdout, &stderr)
	if code == 0 {
		t.Fatal("want non-zero exit for unknown subcommand")
	}
}
