package main

import (
	"os"
	"path/filepath"
	"testing"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

func TestMondayCLISurfaceValidationAndSafety(t *testing.T) {
	findings, _ := validateBundleDir(os.DirFS(filepath.FromSlash("../../internal/connectors/defs")), "monday")
	if len(findings) != 0 {
		t.Fatalf("monday validate findings = %+v, want none", findings)
	}

	b, err := engine.Load(defs.FS, "monday")
	if err != nil {
		t.Fatalf("Load(defs.FS, monday): %v", err)
	}
	if b.CLISurface == nil {
		t.Fatal("monday cli_surface.json is not loaded")
	}

	implemented := map[string]string{}
	for _, cmd := range b.CLISurface.Commands {
		if cmd.Intent == "raw_api" || cmd.Intent == "direct_write" {
			t.Fatalf("command %q exposes forbidden intent %q", cmd.Path, cmd.Intent)
		}
		if cmd.Availability != "implemented" || cmd.Intent != "etl" {
			continue
		}
		implemented[cmd.Path] = cmd.Stream
	}

	want := map[string]string{
		"board list": "boards",
		"item list":  "items",
		"user list":  "users",
		"team list":  "teams",
		"tag list":   "tags",
	}
	for path, stream := range want {
		if implemented[path] != stream {
			t.Fatalf("implemented[%q] = %q, want %q (all implemented: %+v)", path, implemented[path], stream, implemented)
		}
	}
	if len(implemented) != len(want) {
		t.Fatalf("implemented commands = %+v, want exactly %+v", implemented, want)
	}
}
