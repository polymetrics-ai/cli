package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunDispatchesCLI(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := run([]string{"help", "connectors"}, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run(help connectors) code = %d, stderr = %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "pm connectors") {
		t.Fatalf("help output missing connectors manual:\n%s", stdout.String())
	}
}
