package gsd

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultHeartbeat = 15 * time.Second
	defaultMaxEvent  = 256 * 1024
)

var (
	ErrSilentTool = errors.New("silent_tool")
	ErrDeadWorker = errors.New("dead_worker")
)

var governedModels = map[string]struct{}{
	"openai-codex/gpt-5.6-sol": {},
	"openai-codex/gpt-5.5":     {},
}

var baseSupportedCommands = map[string]struct{}{
	"auto": {}, "next": {}, "status": {}, "new-milestone": {}, "query": {}, "discuss": {},
}

var milestoneTargetPattern = regexp.MustCompile(`^M[0-9]{3,}(?:-[a-z0-9]+)?$`)

var fireAndForgetUI = map[string]struct{}{
	"notify": {}, "setStatus": {}, "setWidget": {}, "setTitle": {}, "set_editor_text": {},
}

type Terminal string

const (
	TerminalSuccess   Terminal = "success"
	TerminalError     Terminal = "error"
	TerminalBlocked   Terminal = "blocked"
	TerminalCancelled Terminal = "cancelled"
	TerminalTimeout   Terminal = "timeout"
	TerminalRejected  Terminal = "rejected"
)

type Config struct {
	Command               []string
	WorkDir               string
	GSDHome               string
	StateDir              string
	Model                 string
	Thinking              string
	Timeout               time.Duration
	HeartbeatInterval     time.Duration
	StartupNoEventTimeout time.Duration
	MaxEventBytes         int
	Environment           []string
	Container             *ContainerConfig
	Registry              UnitRegistry
	RuntimeGuard          *HostRuntimeGuard
	ContractUnit          string
}

type Heartbeat struct {
	At           time.Time
	LastEventAt  time.Time
	InFlightTool string
	ProcessAlive bool
	ChildStatus  string
	ChildCount   int
	ChildTurns   int64
}

type Observer struct {
	Event     func(Event)
	Heartbeat func(Heartbeat)
	Question  func(context.Context, Question) (UIResponse, error)
}

type Question struct {
	// ID is the stable semantic ID authored by ask_user_questions. It is safe to
	// use for narrowly scoped policy decisions and durable audit records.
	ID string
	// RequestID is the ephemeral extension RPC envelope ID. Responses must use
	// this value so the GSD runtime can correlate them with the pending request.
	RequestID string
	Method    string
	Title     string
	Options   []string
}

type UIResponse struct {
	Value     string `json:"value,omitempty"`
	Confirmed *bool  `json:"confirmed,omitempty"`
	Cancelled bool   `json:"cancelled,omitempty"`
}

type Result struct {
	Terminal Terminal
	ExitCode int
	Err      error
	Stderr   string
	Started  time.Time
	Ended    time.Time
}

type scanResult struct {
	event    Event
	question *Question
	err      error
}

type questionResult struct {
	response UIResponse
	err      error
}

type Runner struct{ config Config }

func NewRunner(config Config) (*Runner, error) {
	if len(config.Registry.Units) == 0 {
		return nil, fmt.Errorf("%w: authoritative unit registry is required", ErrRuntimeContractMismatch)
	}
	if !filepath.IsAbs(config.WorkDir) || !filepath.IsAbs(config.GSDHome) || !filepath.IsAbs(config.StateDir) {
		return nil, errors.New("absolute work directory, controlled GSD home, and delivery state directory are required")
	}
	if within, err := pathWithin(config.WorkDir, config.StateDir); err != nil || within {
		return nil, errors.New("delivery state directory must be outside the worker-controlled project")
	}
	for _, dir := range []string{config.GSDHome, filepath.Join(config.StateDir, "runtime", "gsd-state")} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			return nil, fmt.Errorf("create local GSD runtime directory: %w", err)
		}
	}
	if config.Container == nil {
		if err := ensureLocalSessionBridge(config.GSDHome); err != nil {
			return nil, err
		}
	}
	if config.Container == nil {
		if len(config.Command) == 0 || strings.TrimSpace(config.Command[0]) == "" {
			return nil, errors.New("GSD command is required")
		}
		if len(config.Command) == 2 && filepath.Base(config.Command[1]) == "loader.js" && filepath.Base(filepath.Dir(config.Command[1])) == "dist" {
			if config.RuntimeGuard == nil {
				return nil, fmt.Errorf("%w: official host runtime guard is required", ErrRuntimeContractMismatch)
			}
			if err := config.RuntimeGuard.ValidateForWorkDir(context.Background(), config.WorkDir); err != nil {
				return nil, err
			}
		}
	} else {
		if err := config.Container.Validate(config.WorkDir); err != nil {
			return nil, err
		}
		for _, dir := range []string{config.Container.GSDStateDir, config.Container.PlanningDir, config.Container.SessionsDir, config.Container.BackgroundDir, config.Container.BackupDir} {
			if err := os.MkdirAll(dir, 0o700); err != nil {
				return nil, fmt.Errorf("create isolated container state: %w", err)
			}
		}
		if err := provisionContainerPolicy(config.Container.PolicyDir, config.Container.GSDStateDir); err != nil {
			return nil, err
		}
	}
	if _, ok := governedModels[config.Model]; !ok || config.Thinking != "high" {
		return nil, errors.New("model must be a governed Shepherd or implementation model with high thinking")
	}
	if config.Timeout <= 0 {
		config.Timeout = 5 * time.Minute
	}
	if config.HeartbeatInterval <= 0 {
		config.HeartbeatInterval = defaultHeartbeat
	}
	if config.HeartbeatInterval > defaultHeartbeat {
		return nil, fmt.Errorf("heartbeat interval must not exceed %s", defaultHeartbeat)
	}
	if config.StartupNoEventTimeout <= 0 {
		config.StartupNoEventTimeout = 2 * time.Minute
	}
	if config.StartupNoEventTimeout < 2*config.HeartbeatInterval {
		config.StartupNoEventTimeout = 2 * config.HeartbeatInterval
	}
	if config.MaxEventBytes <= 0 {
		config.MaxEventBytes = defaultMaxEvent
	}
	return &Runner{config: config}, nil
}

func (r *Runner) RequiresSessionBinding() bool {
	return r.config.RuntimeGuard != nil || r.config.Container != nil
}

func (r *Runner) WithModel(model string) (*Runner, error) {
	if _, ok := governedModels[model]; !ok {
		return nil, errors.New("model must be a governed Shepherd or implementation model with high thinking")
	}
	config := r.config
	config.Model = model
	return &Runner{config: config}, nil
}

func (r *Runner) WithUnitContract(unitType string) (*Runner, error) {
	if _, ok := r.config.Registry.Lookup(unitType); !ok {
		return nil, fmt.Errorf("%w: unknown explicit unit contract %q", ErrRuntimeContractMismatch, unitType)
	}
	config := r.config
	config.ContractUnit = unitType
	return &Runner{config: config}, nil
}

func (r *Runner) WithWorkDir(workDir string) (*Runner, error) {
	if !filepath.IsAbs(workDir) {
		return nil, errors.New("absolute work directory is required")
	}
	config := r.config
	config.WorkDir = workDir
	return NewRunner(config)
}

func ensureLocalSessionBridge(gsdHome string) error {
	piSessions := filepath.Join(gsdHome, "agent", "sessions")
	if err := os.MkdirAll(piSessions, 0o700); err != nil {
		return fmt.Errorf("create local Pi sessions directory: %w", err)
	}
	resumeSessions := filepath.Join(gsdHome, "sessions")
	info, err := os.Lstat(resumeSessions)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.Symlink(filepath.Join("agent", "sessions"), resumeSessions); err != nil {
			return fmt.Errorf("create official headless session bridge: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("inspect official headless session bridge: %w", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return errors.New("official headless sessions path conflicts with the governed Pi sessions directory")
	}
	target, err := os.Readlink(resumeSessions)
	if err != nil {
		return fmt.Errorf("read official headless session bridge: %w", err)
	}
	if target != filepath.Join("agent", "sessions") {
		return errors.New("official headless session bridge has an unexpected target")
	}
	return nil
}

func (r *Runner) Run(parent context.Context, command string, args []string, observer Observer) Result {
	started := time.Now().UTC()
	if r.config.RuntimeGuard != nil {
		if err := r.config.RuntimeGuard.ValidateForWorkDir(parent, r.config.WorkDir); err != nil {
			return failedResult(started, err)
		}
	}
	if !r.supportsCommand(command) {
		return Result{Terminal: TerminalRejected, ExitCode: -1, Err: fmt.Errorf("unsupported headless command %q", command), Started: started, Ended: time.Now().UTC()}
	}
	validDiscussArgs := len(args) == 1 && milestoneTargetPattern.MatchString(args[0])
	if len(args) == 3 && milestoneTargetPattern.MatchString(args[0]) && args[1] == "--resume" && sessionIDPattern.MatchString(args[2]) {
		validDiscussArgs = true
	}
	if command == "discuss" && !validDiscussArgs {
		return Result{Terminal: TerminalRejected, ExitCode: -1, Err: errors.New("discuss requires one canonical milestone target"), Started: started, Ended: time.Now().UTC()}
	}
	for i, arg := range args {
		if strings.HasPrefix(arg, "--answers") || strings.HasPrefix(arg, "--context-text") || strings.ContainsAny(arg, "\r\n\x00") {
			return Result{Terminal: TerminalRejected, ExitCode: -1, Err: errors.New("answer files, inline context, and control characters are forbidden"), Started: started, Ended: time.Now().UTC()}
		}
		if arg == "--context" {
			if i+1 >= len(args) || !isWithin(r.config.WorkDir, args[i+1]) {
				return Result{Terminal: TerminalRejected, ExitCode: -1, Err: errors.New("context must be an existing file inside the work directory"), Started: started, Ended: time.Now().UTC()}
			}
		}
	}
	ctx, cancel := context.WithTimeout(parent, r.config.Timeout)
	defer cancel()
	contractUnit := r.config.ContractUnit
	if contractUnit == "" {
		contractUnit = command
		if command == "discuss" {
			contractUnit = "discuss-milestone"
		}
	}
	commandArgs := r.runtimeArgs()
	responseTimeout := r.config.Timeout + time.Minute
	commandArgs = append(commandArgs,
		"headless", "--json", "--supervised", "--model", r.config.Model,
		"--response-timeout", strconv.FormatInt(responseTimeout.Milliseconds(), 10),
		"--max-restarts", "0",
		"--events", "agent_start,agent_end,turn_start,tool_execution_start,tool_execution_end,model_select,thinking_level_select,extension_ui_request",
		"--timeout", strconv.FormatInt(r.config.Timeout.Milliseconds(), 10), command,
	)
	commandArgs = append(commandArgs, args...)
	cmd := r.execCommand(ctx, commandArgs)
	configureProcessTree(cmd)
	cmd.Dir = r.config.WorkDir
	r.configureEnvironment(cmd)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return failedResult(started, err)
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return failedResult(started, err)
	}
	var stderr boundedBuffer
	stderr.limit = 64 * 1024
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return failedResult(started, err)
	}

	events := make(chan scanResult)
	scanDone := make(chan struct{})
	go func() {
		defer close(scanDone)
		scanEvents(ctx, stdout, r.config.MaxEventBytes, events)
	}()
	waited := make(chan error, 1)
	go func() {
		// os/exec documents that Wait must not race a StdoutPipe reader. A fast child can otherwise
		// close the descriptor before Scanner observes EOF and turn a valid exit into a scan error.
		<-scanDone
		waited <- cmd.Wait()
	}()

	ticker := time.NewTicker(r.config.HeartbeatInterval)
	defer ticker.Stop()
	lastEventAt := started
	inFlightTools := make(map[string]string)
	eventsOpen := true
	var questionResults <-chan questionResult
	var pendingQuestion *Question
	for {
		select {
		case scanned, ok := <-events:
			if !ok {
				eventsOpen = false
				events = nil
				continue
			}
			if scanned.err != nil {
				cancel()
				waitErr := <-waited
				if waitErr != nil && !errors.Is(waitErr, context.Canceled) {
					scanned.err = errors.Join(scanned.err, waitErr)
				}
				return Result{Terminal: TerminalError, ExitCode: exitCode(waitErr), Err: scanned.err, Stderr: stderr.String(), Started: started, Ended: time.Now().UTC()}
			}
			if scanned.question != nil {
				if pendingQuestion != nil {
					cancel()
					waitErr := <-waited
					return Result{Terminal: TerminalBlocked, ExitCode: exitCode(waitErr), Err: errors.New("multiple simultaneous human gates are forbidden"), Stderr: stderr.String(), Started: started, Ended: time.Now().UTC()}
				}
				pendingQuestion = scanned.question
				results := make(chan questionResult, 1)
				questionResults = results
				if observer.Question == nil {
					results <- questionResult{response: UIResponse{Cancelled: true}}
				} else {
					question := *scanned.question
					go func() {
						response, questionErr := observer.Question(ctx, question)
						results <- questionResult{response: response, err: questionErr}
					}()
				}
				continue
			}
			lastEventAt = scanned.event.At
			switch scanned.event.Kind {
			case EventToolStart:
				if contractErr := r.config.Registry.ValidateObservedTool(contractUnit, scanned.event.Tool); contractErr != nil {
					cancel()
					waitErr := <-waited
					return Result{Terminal: TerminalError, ExitCode: exitCode(waitErr), Err: contractErr, Stderr: stderr.String(), Started: started, Ended: time.Now().UTC()}
				}
				if _, exists := inFlightTools[scanned.event.ToolCallID]; exists {
					cancel()
					waitErr := <-waited
					return Result{Terminal: TerminalError, ExitCode: exitCode(waitErr), Err: fmt.Errorf("duplicate active tool call %q", scanned.event.ToolCallID), Stderr: stderr.String(), Started: started, Ended: time.Now().UTC()}
				}
				inFlightTools[scanned.event.ToolCallID] = scanned.event.Tool
			case EventToolEnd:
				if _, exists := inFlightTools[scanned.event.ToolCallID]; !exists {
					cancel()
					waitErr := <-waited
					return Result{Terminal: TerminalError, ExitCode: exitCode(waitErr), Err: fmt.Errorf("tool end has no active start for %q", scanned.event.ToolCallID), Stderr: stderr.String(), Started: started, Ended: time.Now().UTC()}
				}
				delete(inFlightTools, scanned.event.ToolCallID)
			}
			if observer.Event != nil {
				observer.Event(scanned.event)
			}
		case answered := <-questionResults:
			questionResults = nil
			if answered.err != nil {
				cancel()
				waitErr := <-waited
				if ctx.Err() != nil {
					return classifyResult(ctx, started, waitErr, stderr.String())
				}
				return Result{Terminal: TerminalBlocked, ExitCode: exitCode(waitErr), Err: answered.err, Stderr: stderr.String(), Started: started, Ended: time.Now().UTC()}
			}
			payload := struct {
				Type string `json:"type"`
				ID   string `json:"id"`
				UIResponse
			}{Type: "extension_ui_response", ID: pendingQuestion.RequestID, UIResponse: answered.response}
			raw, marshalErr := json.Marshal(payload)
			if marshalErr != nil {
				cancel()
				waitErr := <-waited
				return Result{Terminal: TerminalBlocked, ExitCode: exitCode(waitErr), Err: fmt.Errorf("encode supervised response: %w", marshalErr), Stderr: stderr.String(), Started: started, Ended: time.Now().UTC()}
			}
			if _, writeErr := stdin.Write(append(raw, '\n')); writeErr != nil {
				cancel()
				waitErr := <-waited
				return Result{Terminal: TerminalBlocked, ExitCode: exitCode(waitErr), Err: fmt.Errorf("write supervised response: %w", writeErr), Stderr: stderr.String(), Started: started, Ended: time.Now().UTC()}
			}
			pendingQuestion = nil
		case at := <-ticker.C:
			progress, _ := ReadSubagentProgress(r.config.GSDHome, r.config.WorkDir)
			if observer.Heartbeat != nil {
				observer.Heartbeat(Heartbeat{At: at.UTC(), LastEventAt: lastEventAt,
					InFlightTool: summarizeInFlightTools(inFlightTools), ProcessAlive: true,
					ChildStatus: progress.Status, ChildCount: progress.RunningChildren, ChildTurns: progress.Turns})
			}
			if lastEventAt.Equal(started) && len(inFlightTools) == 0 && progress.RunningChildren == 0 && progress.Turns == 0 && at.Sub(started) >= r.config.StartupNoEventTimeout {
				cancel()
				waitErr := <-waited
				return Result{Terminal: TerminalError, ExitCode: exitCode(waitErr), Err: fmt.Errorf("%w: no model, tool, or child activity observed before startup deadline", ErrSilentTool), Stderr: stderr.String(), Started: started, Ended: time.Now().UTC()}
			}
		case waitErr := <-waited:
			if eventsOpen {
				for scanned := range events {
					if scanned.err != nil {
						return Result{Terminal: TerminalError, ExitCode: exitCode(waitErr), Err: scanned.err, Stderr: stderr.String(), Started: started, Ended: time.Now().UTC()}
					}
					if observer.Event != nil {
						observer.Event(scanned.event)
					}
				}
			}
			return classifyResult(ctx, started, waitErr, stderr.String())
		case <-ctx.Done():
			waitErr := <-waited
			return classifyResult(ctx, started, waitErr, stderr.String())
		}
	}
}

func (r *Runner) supportsCommand(command string) bool {
	if _, ok := baseSupportedCommands[command]; ok {
		return true
	}
	return r.config.Registry.IsCanonicalCommand(command)
}

func summarizeInFlightTools(active map[string]string) string {
	tools := make([]string, 0, len(active))
	for _, tool := range active {
		tools = append(tools, tool)
	}
	sort.Strings(tools)
	return strings.Join(tools, ", ")
}

func isWithin(root, path string) bool {
	if path == "-" || !filepath.IsAbs(path) {
		return false
	}
	relative, err := filepath.Rel(root, filepath.Clean(path))
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return false
	}
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

func (r *Runner) Query(parent context.Context) (WorkflowSnapshot, error) {
	if r.config.RuntimeGuard != nil {
		if err := r.config.RuntimeGuard.ValidateForWorkDir(parent, r.config.WorkDir); err != nil {
			return WorkflowSnapshot{}, err
		}
	}
	ctx, cancel := context.WithTimeout(parent, min(r.config.Timeout, 30*time.Second))
	defer cancel()
	args := r.runtimeArgs()
	args = append(args, "headless", "query")
	cmd := r.execCommand(ctx, args)
	configureProcessTree(cmd)
	cmd.Dir = r.config.WorkDir
	r.configureEnvironment(cmd)
	var stdout boundedBuffer
	stdout.limit = 1024 * 1024
	var stderr boundedBuffer
	stderr.limit = 64 * 1024
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return WorkflowSnapshot{}, fmt.Errorf("headless query failed: %w: %s", err, stderr.String())
	}
	return DecodeQuery([]byte(stdout.String()))
}

func (r *Runner) RebuildMarkdown(parent context.Context) error {
	if r.config.RuntimeGuard != nil {
		if err := r.config.RuntimeGuard.ValidateForWorkDir(parent, r.config.WorkDir); err != nil {
			return err
		}
	}
	ctx, cancel := context.WithTimeout(parent, min(r.config.Timeout, 60*time.Second))
	defer cancel()
	args := r.runtimeArgs()
	args = append(args, "--no-session", "--print", "/gsd rebuild markdown")
	cmd := r.execCommand(ctx, args)
	configureProcessTree(cmd)
	cmd.Dir = r.config.WorkDir
	r.configureEnvironment(cmd)
	var stdout, stderr boundedBuffer
	stdout.limit, stderr.limit = 64*1024, 64*1024
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GSD markdown rebuild failed: %w", err)
	}
	notifications := filepath.Join(r.config.WorkDir, ".gsd", "notifications.jsonl")
	if r.config.Container != nil {
		notifications = filepath.Join(r.config.Container.GSDStateDir, "notifications.jsonl")
	}
	return validateRebuildNotification(notifications)
}

func validateRebuildNotification(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return errors.New("GSD rebuild did not emit a durable result")
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	if info.Size() > 1024*1024 {
		if _, err := file.Seek(info.Size()-1024*1024, io.SeekStart); err != nil {
			return err
		}
	}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 256*1024)
	latest := ""
	for scanner.Scan() {
		var entry struct {
			Message string `json:"message"`
		}
		if json.Unmarshal(scanner.Bytes(), &entry) == nil && strings.Contains(entry.Message, "gsd rebuild markdown: rebuilt") {
			latest = entry.Message
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if latest == "" {
		return errors.New("GSD rebuild result is missing")
	}
	if strings.Contains(latest, "Errors:") {
		return errors.New("GSD rebuild completed with projection errors")
	}
	return nil
}

func (r *Runner) execCommand(ctx context.Context, args []string) *exec.Cmd {
	if r.config.Container != nil {
		return exec.CommandContext(ctx, r.config.Container.Engine, r.config.Container.commandArgs(r.config.WorkDir, args)...)
	}
	return exec.CommandContext(ctx, r.config.Command[0], args...)
}

func (r *Runner) runtimeArgs() []string {
	if r.config.Container != nil {
		return nil
	}
	return append([]string{}, r.config.Command[1:]...)
}

func (r *Runner) configureEnvironment(cmd *exec.Cmd) {
	if r.config.Container != nil {
		cmd.Env = os.Environ()
		return
	}
	cmd.Env = governedEnvironment(r.config.GSDHome, r.config.StateDir, r.config.WorkDir, r.config.Environment)
}

func governedEnvironment(gsdHome, stateDir, workDir string, extra []string) []string {
	combined := append(append([]string{}, os.Environ()...), extra...)
	environment := make([]string, 0, 16)
	allowed := map[string]struct{}{
		"PATH": {}, "TMPDIR": {}, "LANG": {}, "LC_ALL": {}, "TERM": {}, "COLORTERM": {}, "NO_COLOR": {},
		"GO_WANT_RUNNER_HELPER": {}, "RUNNER_HELPER_MODE": {},
	}
	for _, entry := range combined {
		name, _, ok := strings.Cut(entry, "=")
		upper := strings.ToUpper(name)
		if !ok {
			continue
		}
		if _, keep := allowed[upper]; keep {
			environment = append(environment, entry)
		}
	}
	return append(environment,
		"HOME="+gsdHome,
		"GSD_HOME="+gsdHome,
		"GSD_PROJECT_ROOT="+workDir,
		"GSD_STATE_DIR="+filepath.Join(stateDir, "runtime", "gsd-state"),
		"GH_CONFIG_DIR="+filepath.Join(gsdHome, "gh-disabled"),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS=",
		"GIT_CONFIG_COUNT=5",
		"GIT_CONFIG_KEY_0=credential.helper",
		"GIT_CONFIG_VALUE_0=",
		"GIT_CONFIG_KEY_1=remote.origin.pushurl",
		"GIT_CONFIG_VALUE_1=file:///dev/null/shepherd-disabled",
		"GIT_CONFIG_KEY_2=safe.directory",
		"GIT_CONFIG_VALUE_2="+workDir,
		"GIT_CONFIG_KEY_3=user.name",
		"GIT_CONFIG_VALUE_3=Polymetrics Shepherd",
		"GIT_CONFIG_KEY_4=user.email",
		"GIT_CONFIG_VALUE_4=shepherd@localhost.invalid",
	)
}

func pathWithin(root, candidate string) (bool, error) {
	canonicalRoot, err := canonicalPathAllowMissing(root)
	if err != nil {
		return false, err
	}
	canonicalCandidate, err := canonicalPathAllowMissing(candidate)
	if err != nil {
		return false, err
	}
	relative, err := filepath.Rel(canonicalRoot, canonicalCandidate)
	if err != nil {
		return false, err
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)), nil
}

func canonicalPathAllowMissing(path string) (string, error) {
	if path == "" || !filepath.IsAbs(path) {
		return "", errors.New("absolute path is required")
	}
	clean := filepath.Clean(path)
	if info, err := os.Lstat(clean); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			return "", errors.New("path boundary must not be a symlink")
		}
		return filepath.EvalSymlinks(clean)
	} else if !os.IsNotExist(err) {
		return "", err
	}
	ancestor := filepath.Dir(clean)
	missing := []string{filepath.Base(clean)}
	for {
		if _, err := os.Lstat(ancestor); err == nil {
			resolved, err := filepath.EvalSymlinks(ancestor)
			if err != nil {
				return "", err
			}
			for i := len(missing) - 1; i >= 0; i-- {
				resolved = filepath.Join(resolved, missing[i])
			}
			return resolved, nil
		} else if !os.IsNotExist(err) {
			return "", err
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			return "", errors.New("path boundary has no existing ancestor")
		}
		missing = append(missing, filepath.Base(ancestor))
		ancestor = parent
	}
}

func scanEvents(ctx context.Context, reader io.Reader, maxBytes int, output chan<- scanResult) {
	defer close(output)
	send := func(result scanResult) bool {
		select {
		case output <- result:
			return true
		case <-ctx.Done():
			return false
		}
	}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 4096), maxBytes)
	type questionMeta struct {
		ID      string
		Options []string
	}
	questionMetaByTitle := make(map[string]questionMeta)
	for scanner.Scan() {
		var header struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &header); err != nil {
			send(scanResult{err: fmt.Errorf("decode event header: %w", err)})
			return
		}
		if header.Type == string(EventToolStart) {
			var toolStart struct {
				ToolName string `json:"toolName"`
				Input    struct {
					Questions []struct {
						ID       string `json:"id"`
						Header   string `json:"header"`
						Question string `json:"question"`
						Options  []struct {
							Label string `json:"label"`
						} `json:"options"`
					} `json:"questions"`
				} `json:"input"`
				Args struct {
					Questions []struct {
						ID       string `json:"id"`
						Header   string `json:"header"`
						Question string `json:"question"`
						Options  []struct {
							Label string `json:"label"`
						} `json:"options"`
					} `json:"questions"`
				} `json:"args"`
			}
			if err := json.Unmarshal(scanner.Bytes(), &toolStart); err != nil {
				send(scanResult{err: fmt.Errorf("decode tool metadata: %w", err)})
				return
			}
			if isAskUserQuestions(toolStart.ToolName) {
				questions := toolStart.Input.Questions
				if len(questions) == 0 {
					questions = toolStart.Args.Questions
				}
				for _, question := range questions {
					if question.ID == "" || question.Header == "" || question.Question == "" {
						continue
					}
					options := make([]string, 0, len(question.Options))
					for _, option := range question.Options {
						options = append(options, option.Label)
					}
					questionMetaByTitle[question.Header+": "+question.Question] = questionMeta{ID: question.ID, Options: options}
				}
			}
		}
		if header.Type == "extension_ui_request" {
			var question struct {
				ID      string   `json:"id"`
				Method  string   `json:"method"`
				Title   string   `json:"title"`
				Options []string `json:"options"`
			}
			if err := json.Unmarshal(scanner.Bytes(), &question); err != nil || question.ID == "" || question.Method == "" {
				send(scanResult{err: errors.New("invalid supervised UI request")})
				return
			}
			if _, ok := fireAndForgetUI[question.Method]; ok {
				continue
			}
			meta := questionMetaByTitle[question.Title]
			if !send(scanResult{question: &Question{ID: meta.ID, RequestID: question.ID, Method: question.Method, Title: question.Title, Options: question.Options}}) {
				return
			}
			continue
		}
		event, err := ProjectEvent(scanner.Bytes(), maxBytes)
		if !send(scanResult{event: event, err: err}) {
			return
		}
		if err != nil {
			return
		}
	}
	if err := scanner.Err(); err != nil {
		send(scanResult{err: fmt.Errorf("scan GSD event stream: %w", err)})
	}
}

func isAskUserQuestions(toolName string) bool {
	if toolName == "ask_user_questions" {
		return true
	}
	parts := strings.Split(toolName, "__")
	return len(parts) == 3 && parts[0] == "mcp" && parts[1] != "" && parts[2] == "ask_user_questions"
}

func classifyResult(ctx context.Context, started time.Time, err error, stderr string) Result {
	result := Result{ExitCode: exitCode(err), Err: err, Stderr: stderr, Started: started, Ended: time.Now().UTC()}
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		result.Terminal = TerminalTimeout
		result.Err = errors.Join(context.DeadlineExceeded, err)
	case errors.Is(ctx.Err(), context.Canceled):
		result.Terminal = TerminalCancelled
		result.Err = errors.Join(context.Canceled, err)
	case err == nil:
		result.Terminal = TerminalSuccess
	case result.ExitCode == 10:
		result.Terminal = TerminalBlocked
	case result.ExitCode == 11:
		result.Terminal = TerminalCancelled
	default:
		result.Terminal = TerminalError
		if err != nil {
			result.Err = errors.Join(ErrDeadWorker, err)
		}
	}
	return result
}

func failedResult(started time.Time, err error) Result {
	return Result{Terminal: TerminalError, ExitCode: -1, Err: err, Started: started, Ended: time.Now().UTC()}
}

func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

type boundedBuffer struct {
	mu        sync.Mutex
	data      bytes.Buffer
	limit     int
	truncated bool
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	written := len(p)
	remaining := b.limit - b.data.Len()
	if len(p) > remaining {
		b.truncated = true
	}
	if remaining > 0 {
		if len(p) > remaining {
			p = p[:remaining]
		}
		_, _ = b.data.Write(p)
	}
	return written, nil
}

func (b *boundedBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.data.String()
}

func (b *boundedBuffer) Truncated() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.truncated
}
