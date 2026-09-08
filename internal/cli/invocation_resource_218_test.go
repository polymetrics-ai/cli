package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"polymetrics.ai/internal/app"
	"runtime/pprof"
	"strings"
	"testing"
	"time"
)

func TestCLIInvocationResources218(t *testing.T) {
	if os.Getenv("PM_TEST_RESOURCE_CHILD_218") == "1" {
		stacks := func() int {
			var b bytes.Buffer
			if err := pprof.Lookup("goroutine").WriteTo(&b, 2); err != nil {
				t.Fatal(err)
			}
			return strings.Count(b.String(), ".(*CircuitBreakerManager).cleanupLoop(")
		}
		before := stacks()
		for range 3 {
			if code := Run([]string{"--root", t.TempDir(), "version"}, io.Discard, io.Discard); code != 0 {
				t.Fatal("public Run failed")
			}
		}
		root := t.TempDir()
		if err := app.InitProject(root); err != nil {
			t.Fatal(err)
		}
		for _, tc := range []struct {
			args    []string
			healthy bool
		}{{[]string{"help"}, true}, {[]string{"connectors", "inspect", "github", "--json"}, true}, {[]string{"connectors", "inspect", "asana", "--sources", "--json"}, true}, {[]string{"credentials", "test", "absent"}, false}} {
			code := Run(append([]string{"--root", root}, tc.args...), io.Discard, io.Discard)
			if (code == 0) != tc.healthy {
				t.Fatal("local/help/source/missing-credential control changed outcome")
			}
		}
		deadline := time.Now().Add(2 * time.Second)
		for stacks() != before && time.Now().Before(deadline) {
			time.Sleep(10 * time.Millisecond)
		}
		after := stacks()
		t.Logf("public Run version_cycles=3 additional_local_controls=4 maintenance_loops_before=%d after=%d", before, after)
		if after != before {
			t.Fatal("ordinary public invocations retained Redis maintenance loops")
		}
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCLIInvocationResources218$", "-test.v", "-test.count=1")
	cmd.Env = append(os.Environ(), "PM_TEST_RESOURCE_CHILD_218=1")
	out, err := cmd.CombinedOutput()
	t.Logf("fresh child output:\n%s", out)
	if err != nil {
		t.Fatalf("resource child: %v", err)
	}
}
