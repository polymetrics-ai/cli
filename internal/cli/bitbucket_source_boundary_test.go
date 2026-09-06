package cli_test

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync/atomic"
	"testing"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/cli"
	"polymetrics.ai/internal/connectors/engine"
)

const bitbucketSourceBoundaryDefsRoot = "../connectors/defs"

// TestBitbucketPublicCommandsStayCredentialBound proves that every implemented
// Bitbucket command reaches the ordinary credential boundary without provider
// I/O when no credential is supplied.
func TestBitbucketPublicCommandsStayCredentialBound(t *testing.T) {
	bundle, err := engine.Load(os.DirFS(bitbucketSourceBoundaryDefsRoot), "bitbucket")
	if err != nil {
		t.Fatalf("load Bitbucket declaration bundle: %v", err)
	}
	if bundle.CLISurface == nil {
		t.Fatal("Bitbucket CLI surface is absent")
	}

	public := make([]engine.CLICommand, 0, len(bundle.CLISurface.Commands))
	for _, command := range bundle.CLISurface.Commands {
		if command.Availability != "implemented" {
			continue
		}
		if command.Intent != "etl" && command.Intent != "reverse_etl" {
			t.Fatalf("Bitbucket implemented command %q intent=%q, want ETL or reverse ETL", command.Path, command.Intent)
		}
		public = append(public, command)
	}
	sort.Slice(public, func(i, j int) bool { return public[i].Path < public[j].Path })
	if got, want := len(public), 3; got != want {
		t.Fatalf("Bitbucket implemented public commands=%d, want %d execution commands", got, want)
	}
	wantCommands := []string{"repositories create", "repositories delete", "repositories list"}
	for index, command := range public {
		if command.Path != wantCommands[index] {
			t.Fatalf("Bitbucket public command[%d]=%q, want %q", index, command.Path, wantCommands[index])
		}
	}

	root := t.TempDir()
	if err := app.InitProject(root); err != nil {
		t.Fatalf("init Bitbucket command-boundary project: %v", err)
	}
	spy := &bitbucketNoNetworkTransportSpy{}
	previous := http.DefaultTransport
	http.DefaultTransport = spy
	t.Cleanup(func() { http.DefaultTransport = previous })

	for _, command := range public {
		command := command
		t.Run(strings.ReplaceAll(command.Path, " ", "_"), func(t *testing.T) {
			args := append([]string{"bitbucket"}, strings.Fields(command.Path)...)
			args = append(args, "--root", root, "--config", "base_url=https://invalid.example")
			var stdout, stderr bytes.Buffer
			if code := cli.Run(args, &stdout, &stderr); code == 0 {
				t.Fatalf("Run(%v) unexpectedly succeeded; stdout=%s stderr=%s", args, stdout.String(), stderr.String())
			}
			output := strings.TrimSpace(stdout.String() + stderr.String())
			if output != "error: missing --credential" {
				t.Fatalf("Run(%v)=%q, want ordinary credential-bound refusal", args, output)
			}
			lower := strings.ToLower(output)
			for _, forbidden := range []string{"source-bound", "source descriptor", "source operation", "preflight"} {
				if strings.Contains(lower, forbidden) {
					t.Fatalf("Run(%v) reached descriptor-specific refusal %q: %s", args, forbidden, output)
				}
			}
		})
	}
	if got := spy.requests.Load(); got != 0 {
		t.Fatalf("Bitbucket public command provider requests=%d, want zero", got)
	}
}

type bitbucketNoNetworkTransportSpy struct {
	requests atomic.Int64
}

func (spy *bitbucketNoNetworkTransportSpy) RoundTrip(*http.Request) (*http.Response, error) {
	spy.requests.Add(1)
	return nil, fmt.Errorf("unexpected Bitbucket provider I/O")
}
