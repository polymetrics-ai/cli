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
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	requiredModel    = "openai-codex/gpt-5.6-sol"
	defaultHeartbeat = 15 * time.Second
	defaultMaxEvent  = 256 * 1024
)

var supportedCommands = map[string]struct{}{
	"auto": {}, "next": {}, "status": {}, "new-milestone": {}, "query": {},
}

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
	Command           []string
	WorkDir           string
	GSDHome           string
	Model             string
	Thinking          string
	Timeout           time.Duration
	HeartbeatInterval time.Duration
	MaxEventBytes     int
	Environment       []string
	Container         *ContainerConfig
}

type Heartbeat struct {
	At           time.Time
	LastEventAt  time.Time
	InFlightTool string
	ProcessAlive bool
}

type Observer struct {
	Event     func(Event)
	Heartbeat func(Heartbeat)
	Question  func(context.Context, Question) (UIResponse, error)
}

type Question struct {
	ID      string
	Method  string
	Title   string
	Options []string
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
	if !filepath.IsAbs(config.WorkDir) || !filepath.IsAbs(config.GSDHome) {
		return nil, errors.New("absolute work directory and controlled GSD home are required")
	}
	if config.Container == nil {
		if len(config.Command) == 0 || strings.TrimSpace(config.Command[0]) == "" {
			return nil, errors.New("GSD command is required")
		}
	} else {
		if err := config.Container.Validate(config.WorkDir); err != nil {
			return nil, err
		}
		for _, dir := range []string{config.Container.GSDStateDir, config.Container.PlanningDir} {
			if err := os.MkdirAll(dir, 0o700); err != nil {
				return nil, fmt.Errorf("create isolated container state: %w", err)
			}
		}
		if err := provisionContainerPolicy(config.Container.PolicyDir, config.Container.GSDStateDir); err != nil {
			return nil, err
		}
	}
	if config.Model != requiredModel || config.Thinking != "high" {
		return nil, fmt.Errorf("model must be %s with high thinking", requiredModel)
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
	if config.MaxEventBytes <= 0 {
		config.MaxEventBytes = defaultMaxEvent
	}
	return &Runner{config: config}, nil
}

func (r *Runner) Run(parent context.Context, command string, args []string, observer Observer) Result {
	started := time.Now().UTC()
	if _, ok := supportedCommands[command]; !ok {
		return Result{Terminal: TerminalRejected, ExitCode: -1, Err: fmt.Errorf("unsupported headless command %q", command), Started: started, Ended: time.Now().UTC()}
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

	commandArgs := r.runtimeArgs()
	responseTimeout := r.config.Timeout + time.Minute
	commandArgs = append(commandArgs,
		"headless", "--json", "--supervised", "--model", r.config.Model,
		"--response-timeout", strconv.FormatInt(responseTimeout.Milliseconds(), 10),
		"--max-restarts", "0",
		"--events", "agent_start,turn_start,tool_execution_start,model_select,thinking_level_select,extension_ui_request",
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
	inFlightTool := ""
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
				inFlightTool = scanned.event.Tool
			case EventToolEnd:
				inFlightTool = ""
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
			}{Type: "extension_ui_response", ID: pendingQuestion.ID, UIResponse: answered.response}
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
			if observer.Heartbeat != nil {
				observer.Heartbeat(Heartbeat{At: at.UTC(), LastEventAt: lastEventAt, InFlightTool: inFlightTool, ProcessAlive: true})
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
	cmd.Env = governedEnvironment(r.config.GSDHome, r.config.Environment)
}

func governedEnvironment(gsdHome string, extra []string) []string {
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
		"GH_CONFIG_DIR="+filepath.Join(gsdHome, "gh-disabled"),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_ASKPASS=",
		"GIT_CONFIG_COUNT=2",
		"GIT_CONFIG_KEY_0=credential.helper",
		"GIT_CONFIG_VALUE_0=",
		"GIT_CONFIG_KEY_1=remote.origin.pushurl",
		"GIT_CONFIG_VALUE_1=file:///dev/null/shepherd-disabled",
	)
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
	for scanner.Scan() {
		var header struct {
			Type string `json:"type"`
		}
		if err := json.Unmarshal(scanner.Bytes(), &header); err != nil {
			send(scanResult{err: fmt.Errorf("decode event header: %w", err)})
			return
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
			if !send(scanResult{question: &Question{ID: question.ID, Method: question.Method, Title: question.Title, Options: question.Options}}) {
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

func classifyResult(ctx context.Context, started time.Time, err error, stderr string) Result {
	result := Result{ExitCode: exitCode(err), Err: err, Stderr: stderr, Started: started, Ended: time.Now().UTC()}
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		result.Terminal = TerminalTimeout
	case errors.Is(ctx.Err(), context.Canceled):
		result.Terminal = TerminalCancelled
	case err == nil:
		result.Terminal = TerminalSuccess
	case result.ExitCode == 10:
		result.Terminal = TerminalBlocked
	case result.ExitCode == 11:
		result.Terminal = TerminalCancelled
	default:
		result.Terminal = TerminalError
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
	mu    sync.Mutex
	data  bytes.Buffer
	limit int
}

func (b *boundedBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	written := len(p)
	remaining := b.limit - b.data.Len()
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
