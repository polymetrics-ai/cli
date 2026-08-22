package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"polymetrics.ai/internal/connectors/certify"
	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

type githubReleasedReadFixtureRequest struct {
	Method string
	Path   string
}

type githubReleasedReadFixtureObserved struct {
	sync.Mutex
	requests []githubReleasedReadFixtureRequest
}

type githubReleasedReadFixtureJob struct {
	command engine.CLICommand
	args    []string
	index   int
}

type githubReleasedReadFixtureResult struct {
	job    githubReleasedReadFixtureJob
	output string
	err    error
}

type githubReleasedReadFixtureRoute struct {
	method          string
	path            string
	contentType     string
	bodyMode        string
	literalSegments int
}

// TestPMBinaryExecutesGitHubDirectReadCandidatesAgainstFixture is a local
// behavioral proof for the declaration-generated GitHub read surface. It
// drives a freshly built pm binary through the real certification runner,
// command parser, direct-read engine, and GitHub HTTP transport. The fixture
// deliberately returns a concrete JSON object for every declared candidate so
// the runner's output assertions execute rather than merely inspecting the
// candidate ledger.
func TestPMBinaryExecutesGitHubDirectReadCandidatesAgainstFixture(t *testing.T) {
	const token = "github-direct-read-fixture-token"
	const tokenEnv = "PM_GITHUB_DIRECT_READ_FIXTURE_TOKEN"
	t.Setenv(tokenEnv, token)

	var observed struct {
		sync.Mutex
		paths []string
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			http.Error(w, "direct-read fixture received a non-GET request", http.StatusMethodNotAllowed)
			return
		}
		if strings.Contains(request.URL.EscapedPath(), "{") || strings.Contains(request.URL.EscapedPath(), "}") {
			http.Error(w, "direct-read fixture received an unresolved path parameter", http.StatusBadRequest)
			return
		}
		if request.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, "direct-read fixture received no declared bearer authentication", http.StatusUnauthorized)
			return
		}
		observed.Lock()
		observed.paths = append(observed.paths, request.URL.EscapedPath())
		observed.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"fixture":"github-direct-read","request_path":"`+request.URL.EscapedPath()+`"}`)
	}))
	t.Cleanup(server.Close)

	binary := buildTransportPM(t)
	root := filepath.Join(t.TempDir(), "project")
	generatedStages := githubGeneratedDirectReadStageSelection(t)
	coordinator := startFixtureSharedRateLimitCoordinator(t)
	output, diagnostics, err := runFixtureTransportPMJSONWithEnvironment(binary, "", map[string]string{
		"POLYMETRICS_DRAGONFLY_ADDR": coordinator.address,
	},
		"connectors", "certify", "github",
		"--direct-read-only",
		"--config", "base_url="+server.URL,
		"--config", "auth_type=token",
		"--config", "rate_limit_account=fixture-account",
		"--config", "owner=Polymetrics-Cert",
		"--config", "repo=pm-cert-3993-20260810-wz0fru",
		"--config", "certification_stages="+generatedStages,
		"--from-env", "token="+tokenEnv,
		"--root", root,
		"--json",
	)
	var envelope struct {
		Kind   string         `json:"kind"`
		Report certify.Report `json:"report"`
	}
	if err := json.Unmarshal([]byte(output), &envelope); err != nil {
		t.Fatalf("decode GitHub direct-read fixture certification stdout: %v\nstdout:\n%s\nstderr:\n%s", err, redactTransportFailureOutput(output, token), redactTransportFailureOutput(diagnostics, token))
	}
	if envelope.Kind != "ConnectorCertification" {
		t.Fatalf("GitHub direct-read fixture certification kind = %q, want ConnectorCertification", envelope.Kind)
	}
	if envelope.Report.Capabilities.DirectRead == nil || envelope.Report.Capabilities.DirectRead.Result != "pass" || envelope.Report.Capabilities.DirectRead.StagesChecked != 97 {
		t.Fatalf("GitHub direct-read fixture capability = %+v (command error=%v), want 97 passing generated candidate executions", envelope.Report.Capabilities.DirectRead, err)
	}
	generatedStageCount := 0
	for _, stage := range envelope.Report.Stages {
		if strings.HasPrefix(stage.Name, "generated_direct_read_") {
			if !stage.Passed || stage.CLI.Kind != "ConnectorCommandDirectRead" || stage.CLI.ExitCode != 0 {
				t.Fatalf("generated direct-read fixture stage = %+v, want successful command output", stage)
			}
			generatedStageCount++
		}
	}
	if generatedStageCount != 97 {
		t.Fatalf("generated direct-read fixture stages = %d, want 97 including all 77 restored commands", generatedStageCount)
	}
	observed.Lock()
	requestCount := len(observed.paths)
	observed.Unlock()
	if requestCount < 98 { // 97 direct-read stages plus credential validation.
		t.Fatalf("GitHub direct-read fixture requests = %d, want at least 98 real connector HTTP requests", requestCount)
	}
	if keys := coordinator.keyCount(t); keys == 0 {
		t.Fatal("GitHub direct-read fixture made provider requests without storing any declared shared-admission state")
	}
}

// TestPMBinaryExecutesGitHubDisputedPartialVerdictsAgainstFixture adjudicates
// the exact commands whose hard-coded partial expectations fail the release
// gate. Each command has the same implemented declaration it had in v0.2.1;
// this test proves the real binary emits its declared request and accepts the
// declared JSON result before a verdict test may classify it as non-executable.
func TestPMBinaryExecutesGitHubDisputedPartialVerdictsAgainstFixture(t *testing.T) {
	const token = "github-disputed-partial-fixture-token"
	const tokenEnv = "PM_GITHUB_DISPUTED_PARTIAL_FIXTURE_TOKEN"
	t.Setenv(tokenEnv, token)

	paths := []string{
		"agent-task list",
		"cache list",
		"codespace list",
		"gist list",
		"gpg-key list",
		"issue status",
		"org list",
		"pr checks",
		"repo gitignore list",
		"repo license list",
		"ruleset check",
		"search prs",
		"secret list",
		"ssh-key list",
		"variable list",
	}
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("load GitHub disputed partial verdict commands: %v", err)
	}
	if bundle.CLISurface == nil {
		t.Fatal("GitHub bundle has no CLI surface")
	}
	commands := make([]engine.CLICommand, 0, len(paths))
	byPath := make(map[string]engine.CLICommand, len(bundle.CLISurface.Commands))
	for _, command := range bundle.CLISurface.Commands {
		byPath[command.Path] = command
	}
	for _, path := range paths {
		command, found := byPath[path]
		if !found {
			t.Fatalf("GitHub disputed command %q is absent", path)
		}
		if command.Availability != "implemented" || command.Intent != "direct_read" || len(command.APISurface) != 1 || command.OutputPolicy != "json_redacted" {
			t.Fatalf("GitHub disputed command %q declaration = %+v, want one implemented JSON direct-read route", path, command)
		}
		commands = append(commands, command)
	}

	var observed githubReleasedReadFixtureObserved
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodGet {
			http.Error(w, "disputed-partial fixture received a non-GET request", http.StatusMethodNotAllowed)
			return
		}
		if request.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, "disputed-partial fixture received no declared bearer authentication", http.StatusUnauthorized)
			return
		}
		if strings.Contains(request.URL.EscapedPath(), "{") || strings.Contains(request.URL.EscapedPath(), "}") {
			http.Error(w, "disputed-partial fixture received an unresolved path parameter", http.StatusBadRequest)
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
	mustRunReleasedReadPM(t, binary, token, "init", "--root", root, "--json")
	mustRunReleasedReadPM(t, binary, token,
		"credentials", "add", "github-disputed-partial",
		"--connector", "github",
		"--config", "base_url="+server.URL,
		"--config", "auth_type=token",
		"--config", "rate_limit_account=disputed-partial-fixture",
		"--config", "owner=Polymetrics-Cert",
		"--config", "repo=pm-cert-3993-20260810-wz0fru",
		"--from-env", "token="+tokenEnv,
		"--root", root,
		"--json",
	)
	observed.Lock()
	observed.requests = nil
	observed.Unlock()

	for index, command := range commands {
		t.Run(command.Path, func(t *testing.T) {
			args := append([]string{"github"}, strings.Fields(command.Path)...)
			args = append(args, "--credential", "github-disputed-partial")
			for _, flag := range command.Flags {
				if flag.Required {
					args = append(args, "--"+flag.Name, githubReleasedReadFixtureFlagValue(t, command, flag))
				}
			}
			args = append(args, "--root", root, "--json")

			before := githubReleasedReadFixtureRequestCount(&observed)
			output, diagnostics, err := runFixtureTransportPMJSON(binary, "", args...)
			if err != nil {
				t.Fatalf("pm %s failed: %v\nstderr:\n%s", command.Path, err, redactTransportFailureOutput(diagnostics, token))
			}
			githubReleasedReadFixtureAssertOutput(t, command, output, root, index)
			var envelope struct {
				Method string `json:"method"`
				Path   string `json:"path"`
			}
			if err := json.Unmarshal([]byte(output), &envelope); err != nil {
				t.Fatalf("decode %s emitted command output: %v\n%s", command.Path, err, output)
			}
			observed.Lock()
			if len(observed.requests) != before+1 {
				observed.Unlock()
				t.Fatalf("%s fixture requests = %d, want exactly one new declared provider request", command.Path, len(observed.requests)-before)
			}
			request := observed.requests[before]
			observed.Unlock()
			wantMethod := strings.ToUpper(command.APISurface[0].Method)
			if request.Method != wantMethod || !githubReleasedReadFixturePathMatches(command.APISurface[0].Path, request.Path) {
				t.Fatalf("%s provider request = %+v, want declared %s %s with all path parameters resolved", command.Path, request, wantMethod, command.APISurface[0].Path)
			}
			if request.Method != envelope.Method || request.Path != envelope.Path {
				t.Fatalf("%s provider request = %+v, emitted output = %+v; want exact declared request/output agreement", command.Path, request, envelope)
			}
		})
	}
}

// runFixtureTransportPMJSON keeps a fixture child's machine-readable stdout
// separate from diagnostics and removes the runtime coordinator aliases that
// an unrelated test may export. Certification fixtures that need a shared
// coordinator pass one explicitly through runFixtureTransportPMJSONWithEnvironment.
func runFixtureTransportPMJSON(binary, stdin string, args ...string) (stdout, stderr string, err error) {
	return runFixtureTransportPMJSONWithEnvironment(binary, stdin, nil, args...)
}

// runFixtureTransportPMJSONWithEnvironment starts the real pm binary with a
// self-contained environment. It deliberately replaces both documented
// coordinator aliases rather than inheriting a process-global address: a
// certification tier that declares require_shared must use the fixture's
// actual local coordinator, never silently fall back to process-local limits.
func runFixtureTransportPMJSONWithEnvironment(binary, stdin string, additions map[string]string, args ...string) (stdout, stderr string, err error) {
	command := exec.Command(binary, args...)
	command.Env = transportFixtureEnvironment(os.Environ(), additions)
	if stdin != "" {
		command.Stdin = strings.NewReader(stdin)
	}
	var stdoutBuffer, stderrBuffer bytes.Buffer
	command.Stdout = &stdoutBuffer
	command.Stderr = &stderrBuffer
	err = command.Run()
	return stdoutBuffer.String(), stderrBuffer.String(), err
}

func transportFixtureEnvironment(environment []string, additions map[string]string) []string {
	const (
		primaryCoordinatorAddress = "POLYMETRICS_DRAGONFLY_ADDR"
		aliasCoordinatorAddress   = "PM_DRAGONFLY_ADDR"
	)
	filtered := make([]string, 0, len(environment))
	for _, entry := range environment {
		name, _, found := strings.Cut(entry, "=")
		if found && (name == primaryCoordinatorAddress || name == aliasCoordinatorAddress) {
			continue
		}
		filtered = append(filtered, entry)
	}
	for name, value := range additions {
		filtered = append(filtered, name+"="+value)
	}
	return filtered
}

// fixtureSharedRateLimitCoordinator is an isolated, disposable
// Redis-compatible coordinator for the fresh pm process. The production
// registry is explicitly Redis-compatible, and the build-tagged coordination
// integration tests cover the same shared admission scripts against Dragonfly.
// This fixture keeps its own Redis process so ordinary CI does not need a
// shared daemon, while certification stages still execute their declared
// require_shared admission path end-to-end.
type fixtureSharedRateLimitCoordinator struct {
	address string
}

func startFixtureSharedRateLimitCoordinator(t *testing.T) fixtureSharedRateLimitCoordinator {
	t.Helper()
	redisServer, err := exec.LookPath("redis-server")
	if err != nil {
		t.Fatalf("GitHub certification fixture requires the Redis-compatible redis-server executable: %v", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve local shared-coordinator address: %v", err)
	}
	address := listener.Addr().String()
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatalf("release local shared-coordinator address: %v", err)
	}

	process := exec.Command(redisServer,
		"--bind", "127.0.0.1",
		"--port", strconv.Itoa(port),
		"--save", "",
		"--appendonly", "no",
		"--protected-mode", "no",
		"--loglevel", "warning",
	)
	process.Stdout = io.Discard
	process.Stderr = io.Discard
	if err := process.Start(); err != nil {
		t.Fatalf("start local shared coordinator: %v", err)
	}
	t.Cleanup(func() {
		if process.Process == nil {
			return
		}
		_ = process.Process.Signal(os.Interrupt)
		done := make(chan error, 1)
		go func() { done <- process.Wait() }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			_ = process.Process.Kill()
			<-done
		}
	})
	if err := waitForFixtureSharedRateLimitCoordinator(address); err != nil {
		t.Fatalf("wait for local shared coordinator at %s: %v", address, err)
	}
	return fixtureSharedRateLimitCoordinator{address: address}
}

func waitForFixtureSharedRateLimitCoordinator(address string) error {
	deadline := time.Now().Add(5 * time.Second)
	var lastErr error
	for time.Now().Before(deadline) {
		if err := fixtureSharedRateLimitCoordinatorCommand(address, "PING", func(response string) error {
			if response != "+PONG" {
				return fmt.Errorf("PING response = %q, want +PONG", response)
			}
			return nil
		}); err == nil {
			return nil
		} else {
			lastErr = err
		}
		time.Sleep(25 * time.Millisecond)
	}
	return lastErr
}

func (c fixtureSharedRateLimitCoordinator) keyCount(t *testing.T) int {
	t.Helper()
	var count int
	if err := fixtureSharedRateLimitCoordinatorCommand(c.address, "DBSIZE", func(response string) error {
		if !strings.HasPrefix(response, ":") {
			return fmt.Errorf("DBSIZE response = %q, want integer", response)
		}
		value, err := strconv.Atoi(strings.TrimPrefix(response, ":"))
		if err != nil {
			return fmt.Errorf("decode DBSIZE response %q: %w", response, err)
		}
		count = value
		return nil
	}); err != nil {
		t.Fatalf("inspect local shared coordinator: %v", err)
	}
	return count
}

func fixtureSharedRateLimitCoordinatorCommand(address, command string, check func(string) error) error {
	connection, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		return err
	}
	defer connection.Close()
	if err := connection.SetDeadline(time.Now().Add(time.Second)); err != nil {
		return err
	}
	request := "*1\r\n$" + strconv.Itoa(len(command)) + "\r\n" + command + "\r\n"
	if _, err := io.WriteString(connection, request); err != nil {
		return err
	}
	response, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil {
		return err
	}
	return check(strings.TrimSpace(response))
}

func githubGeneratedDirectReadStageSelection(t *testing.T) string {
	t.Helper()
	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("load GitHub certification candidates: %v", err)
	}
	if bundle.Certification == nil {
		t.Fatal("GitHub certification definition is absent")
	}
	stages := make([]string, 0, len(bundle.Certification.DirectReadCandidates))
	for _, candidate := range bundle.Certification.DirectReadCandidates {
		if candidate.Generated {
			stages = append(stages, candidate.StageName)
		}
	}
	sort.Strings(stages)
	return strings.Join(stages, ",")
}

// TestPMBinaryExecutesGitHubReleasedReadSurfaceAgainstFixture is the release
// guard for source-projection reachability. It executes every API-surface
// direct-read target through a freshly built pm binary, asserting each
// declaration's HTTP method, an interpolated request path, and the returned
// fixture output. The 633 targets are a strict superset of the 370 routes
// restored from the over-blocking projection; this is local fixture evidence,
// not a live-provider certification claim.
func TestPMBinaryExecutesGitHubReleasedReadSurfaceAgainstFixture(t *testing.T) {
	const token = "github-released-read-fixture-token"
	const tokenEnv = "PM_GITHUB_RELEASED_READ_FIXTURE_TOKEN"
	t.Setenv(tokenEnv, token)

	bundle, err := engine.Load(defs.FS, "github")
	if err != nil {
		t.Fatalf("load GitHub released read surface: %v", err)
	}
	commands := githubReleasedReadSurfaceCommands(t, bundle)
	if len(commands) != 633 {
		t.Fatalf("GitHub released read commands = %d, want 633 API-surface direct-read targets", len(commands))
	}
	routes := githubReleasedReadFixtureRoutes(t, bundle, commands)
	if _, mode := githubReleasedReadFixtureResponse(routes, http.MethodGet, "/gists/fixture/star"); mode != "empty" {
		t.Fatalf("GitHub gists star fixture mode = %q, want declaration-owned empty status response", mode)
	}

	var observed githubReleasedReadFixtureObserved
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, "released-read fixture received no declared bearer authentication", http.StatusUnauthorized)
			return
		}
		if strings.Contains(request.URL.EscapedPath(), "{") || strings.Contains(request.URL.EscapedPath(), "}") {
			http.Error(w, "released-read fixture received an unresolved path parameter", http.StatusBadRequest)
			return
		}
		observed.Lock()
		observed.requests = append(observed.requests, githubReleasedReadFixtureRequest{Method: request.Method, Path: request.URL.EscapedPath()})
		observed.Unlock()
		contentType, bodyMode := githubReleasedReadFixtureResponse(routes, request.Method, request.URL.EscapedPath())
		w.Header().Set("Content-Type", contentType)
		switch bodyMode {
		case "empty":
		case "text":
			_, _ = io.WriteString(w, "github-released-read "+request.Method+" "+request.URL.EscapedPath())
		case "repository_file":
			_, _ = io.WriteString(w, `{"type":"file","fixture":"github-released-read"}`)
		case "repository_directory":
			_, _ = io.WriteString(w, `[{"type":"file","fixture":"github-released-read"}]`)
		default:
			_, _ = io.WriteString(w, `{"fixture":"github-released-read","method":"`+request.Method+`","path":"`+request.URL.EscapedPath()+`"}`)
		}
	}))
	t.Cleanup(server.Close)

	binary := buildTransportPM(t)
	root := filepath.Join(t.TempDir(), "project")
	mustRunReleasedReadPM(t, binary, token, "init", "--root", root, "--json")
	mustRunReleasedReadPM(t, binary, token,
		"credentials", "add", "github-released-read",
		"--connector", "github",
		"--config", "base_url="+server.URL,
		"--config", "auth_type=token",
		"--config", "rate_limit_account=released-read-fixture",
		"--config", "owner=Polymetrics-Cert",
		"--config", "repo=pm-cert-3993-20260810-wz0fru",
		"--from-env", "token="+tokenEnv,
		"--root", root,
		"--json",
	)

	jobs := make([]githubReleasedReadFixtureJob, 0, len(commands))
	for index, command := range commands {
		args := append([]string{"github"}, strings.Fields(command.Path)...)
		args = append(args, "--credential", "github-released-read")
		for _, flag := range command.Flags {
			if !flag.Required {
				continue
			}
			args = append(args, "--"+flag.Name, githubReleasedReadFixtureFlagValue(t, command, flag))
		}
		if command.Intent == "binary_download" || command.Intent == "text_export" {
			destRoot := filepath.Join(root, "downloads", fmt.Sprintf("%03d", index))
			if err := os.MkdirAll(destRoot, 0o755); err != nil {
				t.Fatalf("create fixture destination for %s: %v", command.Path, err)
			}
			args = append(args, "--dest-root", destRoot)
		}
		args = append(args, "--root", root, "--json")
		jobs = append(jobs, githubReleasedReadFixtureJob{command: command, args: args, index: index})
	}

	// Connector command reads are side-effect free against the fixture. Bound
	// real-binary concurrency keeps the exhaustive release-surface test a
	// practical gate while each job still owns a separate pm invocation and,
	// for binary downloads, a separate destination root.
	const workers = 12
	queue := make(chan githubReleasedReadFixtureJob)
	results := make(chan githubReleasedReadFixtureResult, len(jobs))
	var group sync.WaitGroup
	for range workers {
		group.Add(1)
		go func() {
			defer group.Done()
			for job := range queue {
				output, err := runTransportPM(binary, "", job.args...)
				results <- githubReleasedReadFixtureResult{job: job, output: output, err: err}
			}
		}()
	}
	for _, job := range jobs {
		queue <- job
	}
	close(queue)
	group.Wait()
	close(results)

	for result := range results {
		if result.err != nil {
			t.Fatalf("pm %s failed: %v\n%s", result.job.command.Path, result.err, redactTransportFailureOutput(result.output, token))
		}
		githubReleasedReadFixtureAssertOutput(t, result.job.command, result.output, root, result.job.index)
	}
	if got := githubReleasedReadFixtureRequestCount(&observed); got != len(commands) {
		t.Fatalf("GitHub released-read fixture requests = %d, want %d one-request command executions", got, len(commands))
	}
}

func githubReleasedReadSurfaceCommands(t *testing.T, bundle engine.Bundle) []engine.CLICommand {
	t.Helper()
	if bundle.CLISurface == nil {
		t.Fatal("GitHub bundle lacks CLI surface")
	}
	// api_surface.json is intentionally a build-time artifact and is excluded
	// from defs.FS. Load it from this exact test worktree so the fixture proves
	// the released source-projection contract rather than an embedded subset.
	rawSurface, err := os.ReadFile(filepath.Join(transportRepositoryRoot(t), "internal", "connectors", "defs", "github", "api_surface.json"))
	if err != nil {
		t.Fatalf("read GitHub API surface: %v", err)
	}
	surface, err := engine.ParseAPISurface(rawSurface)
	if err != nil {
		t.Fatalf("parse GitHub API surface: %v", err)
	}
	targets := map[string]bool{}
	for _, endpoint := range surface.Endpoints {
		if endpoint.CoveredBy == nil {
			continue
		}
		if endpoint.CoveredBy.DirectRead != "" {
			targets[endpoint.CoveredBy.DirectRead] = true
		}
		for _, command := range endpoint.CoveredBy.DirectReads {
			targets[command] = true
		}
	}
	commands := make([]engine.CLICommand, 0, len(targets))
	for _, command := range bundle.CLISurface.Commands {
		if !targets[command.Path] {
			continue
		}
		if command.Availability != "implemented" || len(command.APISurface) != 1 {
			t.Fatalf("API-surface target %q is not an executable one-endpoint command: %+v", command.Path, command)
		}
		switch command.Intent {
		case "direct_read", "binary_download", "text_export", "status_check":
		default:
			t.Fatalf("API-surface target %q intent = %q, want executable read intent", command.Path, command.Intent)
		}
		commands = append(commands, command)
	}
	if len(commands) != len(targets) {
		t.Fatalf("GitHub API-surface targets = %d, resolved commands = %d", len(targets), len(commands))
	}
	sort.Slice(commands, func(i, j int) bool { return commands[i].Path < commands[j].Path })
	return commands
}

func githubReleasedReadFixtureRoutes(t *testing.T, bundle engine.Bundle, commands []engine.CLICommand) []githubReleasedReadFixtureRoute {
	t.Helper()
	operations := make(map[string]engine.OperationSpec, len(bundle.Operations))
	for _, operation := range bundle.Operations {
		operations[operation.ID] = operation
	}
	routes := make([]githubReleasedReadFixtureRoute, 0)
	for _, command := range commands {
		route := githubReleasedReadFixtureRoute{
			method:          strings.ToUpper(command.APISurface[0].Method),
			path:            command.APISurface[0].Path,
			contentType:     "application/json",
			bodyMode:        "json",
			literalSegments: githubReleasedReadFixtureLiteralSegments(command.APISurface[0].Path),
			// A direct_read command may still have the closed status-only
			// output policy (for example GitHub's membership checks), so the
			// declaration-owned policy, not intent alone, controls the body.
		}
		switch command.OutputPolicy {
		case "none":
			route.bodyMode = "empty"
		case "text":
			route.contentType = "text/plain; charset=utf-8"
			route.bodyMode = "text"
		case "repository_contents_file_metadata":
			route.bodyMode = "repository_file"
		case "repository_contents_directory":
			route.bodyMode = "repository_directory"
		}
		if command.Intent == "binary_download" || command.Intent == "text_export" {
			operation, ok := operations[command.Operation]
			if !ok || operation.Binary == nil || len(operation.Binary.ContentTypes) == 0 {
				t.Fatalf("binary API-surface command %q lacks a declared binary content type", command.Path)
			}
			route.contentType = githubReleasedReadFixtureConcreteContentType(operation.Binary.ContentTypes[0], operation.Binary.Charset)
		}
		routes = append(routes, route)
	}
	return routes
}

func githubReleasedReadFixtureResponse(routes []githubReleasedReadFixtureRoute, method, path string) (contentType, bodyMode string) {
	// GitHub uses the same contents URL template for a file and a directory;
	// the command's declared output policy distinguishes them. The fixture uses
	// the two disposable path values below to exercise both contracts.
	if strings.HasSuffix(path, "/contents/fixture-directory") {
		return "application/json", "repository_directory"
	}
	if strings.HasSuffix(path, "/contents/fixture-file") {
		return "application/json", "repository_file"
	}
	best := -1
	var selected githubReleasedReadFixtureRoute
	for _, route := range routes {
		if route.method == method && route.literalSegments > best && githubReleasedReadFixturePathMatches(route.path, path) {
			best = route.literalSegments
			selected = route
		}
	}
	if best >= 0 {
		return selected.contentType, selected.bodyMode
	}
	return "application/json", "json"
}

func githubReleasedReadFixtureLiteralSegments(path string) int {
	count := 0
	for _, part := range strings.Split(strings.Trim(path, "/"), "/") {
		if !strings.HasPrefix(part, "{") || !strings.HasSuffix(part, "}") {
			count++
		}
	}
	return count
}

func githubReleasedReadFixtureConcreteContentType(declared, charset string) string {
	contentType := strings.TrimSpace(declared)
	switch contentType {
	case "application/*":
		contentType = "application/octet-stream"
	case "text/*":
		contentType = "text/plain"
	}
	if charset != "" && !strings.Contains(strings.ToLower(contentType), "charset=") {
		contentType += "; charset=" + charset
	}
	return contentType
}

func githubReleasedReadFixturePathMatches(template, path string) bool {
	templateParts := strings.Split(strings.Trim(template, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(templateParts) != len(pathParts) {
		return false
	}
	for index, templatePart := range templateParts {
		if strings.HasPrefix(templatePart, "{") && strings.HasSuffix(templatePart, "}") {
			continue
		}
		if templatePart != pathParts[index] {
			return false
		}
	}
	return true
}

func githubReleasedReadFixtureFlagValue(t *testing.T, command engine.CLICommand, flag engine.CLIFlag) string {
	t.Helper()
	if flag.MapsTo == "path.path" {
		switch command.OutputPolicy {
		case "repository_contents_directory":
			return "fixture-directory"
		case "repository_contents_file_metadata":
			return "fixture-file"
		}
	}
	if flag.Type == "enum" {
		if len(flag.Values) == 0 {
			t.Fatalf("required enum --%s has no declared values", flag.Name)
		}
		return flag.Values[0]
	}
	switch flag.Type {
	case "integer", "number":
		return "7"
	case "boolean":
		return "true"
	case "string_array":
		return "sha256:fixture"
	case "", "string":
		return "fixture"
	default:
		t.Fatalf("required --%s uses unsupported fixture type %q", flag.Name, flag.Type)
		return ""
	}
}

func githubReleasedReadFixtureRequestCount(observed *githubReleasedReadFixtureObserved) int {
	observed.Lock()
	defer observed.Unlock()
	return len(observed.requests)
}

func githubReleasedReadFixtureAssertOutput(t *testing.T, command engine.CLICommand, output, root string, index int) {
	t.Helper()
	var envelope struct {
		Kind     string          `json:"kind"`
		Command  string          `json:"command"`
		Method   string          `json:"method"`
		Path     string          `json:"path"`
		Status   int             `json:"status"`
		Response json.RawMessage `json:"response"`
	}
	if err := json.Unmarshal([]byte(output), &envelope); err != nil {
		t.Fatalf("decode %s fixture command output: %v\n%s", command.Path, err, output)
	}
	wantKind := "ConnectorCommandDirectRead"
	if command.Intent == "binary_download" || command.Intent == "text_export" {
		wantKind = "ConnectorCommandBinaryDownload"
	}
	wantMethod := strings.ToUpper(command.APISurface[0].Method)
	if envelope.Kind != wantKind || envelope.Command != command.Path || envelope.Method != wantMethod || strings.TrimSpace(envelope.Path) == "" || envelope.Status != http.StatusOK {
		t.Fatalf("%s output = %+v, want kind=%s command=%q method=%s nonempty path status=%d", command.Path, envelope, wantKind, command.Path, wantMethod, http.StatusOK)
	}
	if wantKind == "ConnectorCommandBinaryDownload" {
		destRoot := filepath.Join(root, "downloads", fmt.Sprintf("%03d", index))
		entries, err := os.ReadDir(destRoot)
		if err != nil || len(entries) != 1 || entries[0].IsDir() {
			t.Fatalf("%s binary fixture output at %s = entries=%v err=%v, want one downloaded file", command.Path, destRoot, entries, err)
		}
		body, err := os.ReadFile(filepath.Join(destRoot, entries[0].Name()))
		if err != nil || !strings.Contains(string(body), `"fixture":"github-released-read"`) {
			t.Fatalf("%s downloaded fixture body = %q err=%v, want declared fixture response", command.Path, body, err)
		}
		return
	}
	if command.OutputPolicy == "none" {
		if response := strings.TrimSpace(string(envelope.Response)); response != "" && response != "null" {
			t.Fatalf("%s status-only response = %s, want no response body", command.Path, response)
		}
		return
	}
	if command.OutputPolicy == "text" {
		var response string
		if err := json.Unmarshal(envelope.Response, &response); err != nil || response != "github-released-read "+wantMethod+" "+envelope.Path {
			t.Fatalf("%s text response = %q err=%v, want fixture text matching emitted request", command.Path, response, err)
		}
		return
	}
	if command.OutputPolicy == "repository_contents_directory" {
		var response []map[string]any
		if err := json.Unmarshal(envelope.Response, &response); err != nil || len(response) != 1 || response[0]["fixture"] != "github-released-read" {
			t.Fatalf("%s directory response = %#v err=%v, want declared fixture listing matching emitted request", command.Path, response, err)
		}
		return
	}
	var response map[string]any
	if err := json.Unmarshal(envelope.Response, &response); err != nil {
		t.Fatalf("decode %s fixture direct-read response: %v\n%s", command.Path, err, output)
	}
	if command.OutputPolicy == "repository_contents_file_metadata" {
		if response["fixture"] != "github-released-read" {
			t.Fatalf("%s file response = %#v, want declared fixture metadata", command.Path, response)
		}
		return
	}
	if response["fixture"] != "github-released-read" || response["method"] != wantMethod || response["path"] != envelope.Path {
		t.Fatalf("%s direct-read response = %#v, want the fixture response matching emitted request", command.Path, response)
	}
}

func mustRunReleasedReadPM(t *testing.T, binary, token string, args ...string) {
	t.Helper()
	output, err := runTransportPM(binary, "", args...)
	if err != nil {
		t.Fatalf("pm %s failed: %v\n%s", transportCommandName(args), err, redactTransportFailureOutput(output, token))
	}
}
