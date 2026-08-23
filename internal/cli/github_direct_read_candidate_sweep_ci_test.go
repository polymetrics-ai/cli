//go:build github_fixture_sweep

package cli

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// TestPMBinaryExecutesGitHubGeneratedDirectReadCandidatesAgainstFixture is
// the exhaustive executable-surface proof for all 97 generated candidates.
// It has a dedicated CI job because it starts a fresh binary for every
// candidate; keeping it out of the shared internal/cli deadline retains every
// behavioral assertion without weakening the default package gate.
func TestPMBinaryExecutesGitHubGeneratedDirectReadCandidatesAgainstFixture(t *testing.T) {
	const token = "github-generated-direct-read-fixture-token"
	const tokenEnv = "PM_GITHUB_GENERATED_DIRECT_READ_FIXTURE_TOKEN"
	t.Setenv(tokenEnv, token)

	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("load GitHub generated direct-read candidates: %v", err)
	}
	candidates := githubGeneratedDirectReadCandidates(t, bundle)
	if len(candidates) != 97 {
		t.Fatalf("GitHub generated direct-read candidates = %d, want 97", len(candidates))
	}
	if bundle.CLISurface == nil {
		t.Fatal("GitHub bundle has no CLI surface")
	}
	commands := make(map[string]engine.CLICommand, len(bundle.CLISurface.Commands))
	for _, command := range bundle.CLISurface.Commands {
		commands[command.Path] = command
	}

	var observed githubReleasedReadFixtureObserved
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			http.Error(w, "generated direct-read fixture received a non-GET request", http.StatusMethodNotAllowed)
			return
		}
		if strings.Contains(request.URL.EscapedPath(), "{") || strings.Contains(request.URL.EscapedPath(), "}") {
			http.Error(w, "generated direct-read fixture received an unresolved path parameter", http.StatusBadRequest)
			return
		}
		if request.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, "generated direct-read fixture received no declared bearer authentication", http.StatusUnauthorized)
			return
		}
		observed.Lock()
		observed.requests = append(observed.requests, githubReleasedReadFixtureRequest{Method: request.Method, Path: request.URL.EscapedPath()})
		observed.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"fixture":"github-released-read","method":"`+request.Method+`","path":"`+request.URL.EscapedPath()+`"}`)
	}))
	t.Cleanup(server.Close)

	binary := buildTransportPM(t)
	root := filepath.Join(t.TempDir(), "project")
	mustRunFixtureTransportPM(t, binary, token, "init", "--root", root, "--json")
	mustRunFixtureTransportPM(t, binary, token,
		"credentials", "add", "github-generated-direct-read",
		"--connector", "github",
		"--config", "base_url="+server.URL,
		"--config", "auth_type=token",
		"--config", "rate_limit_account=generated-direct-read-fixture",
		"--config", "owner=Polymetrics-Cert",
		"--config", "repo=pm-cert-3993-20260810-wz0fru",
		"--from-env", "token="+tokenEnv,
		"--root", root,
		"--json",
	)

	jobs := make([]githubGeneratedDirectReadFixtureJob, 0, len(candidates))
	for index, candidate := range candidates {
		command, found := commands[candidate.Command]
		if !found {
			t.Fatalf("generated candidate %q names absent CLI command %q", candidate.StageName, candidate.Command)
		}
		if command.Availability != "implemented" || command.Intent != "direct_read" || len(command.APISurface) != 1 {
			t.Fatalf("generated candidate %q command = %+v, want one implemented direct-read API surface", candidate.StageName, command)
		}
		args := githubGeneratedDirectReadCandidateArgs(t, candidate, "github-generated-direct-read")
		args = append(args, "--root", root)
		jobs = append(jobs, githubGeneratedDirectReadFixtureJob{candidate: candidate, command: command, args: args, index: index})
	}

	const workers = 12
	queue := make(chan githubGeneratedDirectReadFixtureJob)
	results := make(chan githubGeneratedDirectReadFixtureResult, len(jobs))
	var group sync.WaitGroup
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for job := range queue {
				output, diagnostics, err := runFixtureTransportPMJSON(binary, "", job.args...)
				results <- githubGeneratedDirectReadFixtureResult{job: job, output: output, diagnostics: diagnostics, err: err}
			}
		}()
	}
	for _, job := range jobs {
		queue <- job
	}
	close(queue)
	group.Wait()
	close(results)

	emittedRequests := make(map[githubReleasedReadFixtureRequest]int, len(jobs))
	for result := range results {
		if result.err != nil {
			t.Fatalf("pm %s (%s) failed: %v\nstdout:\n%s\nstderr:\n%s", result.job.candidate.Command, result.job.candidate.StageName, result.err, redactTransportFailureOutput(result.output, token), redactTransportFailureOutput(result.diagnostics, token))
		}
		githubReleasedReadFixtureAssertOutput(t, result.job.command, result.output, root, result.job.index)
		githubGeneratedDirectReadFixtureAssertOutputAssertions(t, result.job.candidate, result.output)

		var envelope struct {
			Method string `json:"method"`
			Path   string `json:"path"`
		}
		if err := json.Unmarshal([]byte(result.output), &envelope); err != nil {
			t.Fatalf("decode %s emitted command output: %v\n%s", result.job.candidate.Command, err, result.output)
		}
		wantMethod := strings.ToUpper(result.job.command.APISurface[0].Method)
		if envelope.Method != wantMethod || !githubReleasedReadFixturePathMatches(result.job.command.APISurface[0].Path, envelope.Path) {
			t.Fatalf("%s emitted request = %s %s, want declared %s %s with all path parameters resolved", result.job.candidate.Command, envelope.Method, envelope.Path, wantMethod, result.job.command.APISurface[0].Path)
		}
		emittedRequests[githubReleasedReadFixtureRequest{Method: envelope.Method, Path: envelope.Path}]++
	}
	if got := githubReleasedReadFixtureRequestCount(&observed); got != len(candidates) {
		t.Fatalf("GitHub generated direct-read fixture requests = %d, want %d one-request candidate executions", got, len(candidates))
	}
	observed.Lock()
	fixtureRequests := make(map[githubReleasedReadFixtureRequest]int, len(observed.requests))
	for _, request := range observed.requests {
		fixtureRequests[request]++
	}
	observed.Unlock()
	if !reflect.DeepEqual(fixtureRequests, emittedRequests) {
		t.Fatalf("GitHub generated direct-read fixture requests = %#v, want emitted declared requests %#v", fixtureRequests, emittedRequests)
	}
}
