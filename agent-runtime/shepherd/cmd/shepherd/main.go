package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/contract"
	decisionlog "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/decision"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/domain"
	shepherdgit "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/git"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/gsd"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/store"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/telemetry"
)

const version = "0.1.0"

type fileConfig struct {
	GSDCommand       []string `json:"gsd_command"`
	WorkDir          string   `json:"work_dir"`
	GSDHome          string   `json:"gsd_home"`
	StateDir         string   `json:"state_dir"`
	CoordinatorModel string   `json:"coordinator_model"`
	GSDVersion       string   `json:"gsd_version"`
	TimeoutSeconds   int      `json:"timeout_seconds"`
	HeartbeatSeconds int      `json:"heartbeat_seconds"`
	MaxEventBytes    int      `json:"max_event_bytes"`
	Runtime          string   `json:"runtime"`
	ContainerImage   string   `json:"container_image"`
	AuthFile         string   `json:"auth_file"`
	ContainerNetwork string   `json:"container_network"`
	PolicyDir        string   `json:"policy_dir"`
	GitCommonDir     string   `json:"git_common_dir"`
}

type decisionInput struct {
	DeliveryID string `json:"delivery_id"`
	Generation int64  `json:"generation"`
	Actor      string `json:"actor"`
	Approved   bool   `json:"approved"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "shepherd:", err)
		var coded interface{ ExitCode() int }
		if errors.As(err, &coded) {
			os.Exit(coded.ExitCode())
		}
		os.Exit(1)
	}
}

func run(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("usage: shepherd <version|query|eval|repair|resume|run|start> [flags]")
	}
	if args[0] == "version" {
		fmt.Println(version)
		return nil
	}
	flags := flag.NewFlagSet(args[0], flag.ContinueOnError)
	configPath := flags.String("config", "", "path to governed Shepherd JSON config")
	command := flags.String("command", "next", "supported GSD headless command")
	issue := flags.Int("issue", 0, "GitHub issue number")
	contextPath := flags.String("context", "", "validated milestone context file")
	auto := flags.Bool("auto", false, "continue into GSD auto-mode after milestone intake")
	adoptExisting := flags.Bool("adopt-existing", false, "bind a previously created active GSD milestone")
	decisionPath := flags.String("decision", "", "path to protected explicit human decision JSON")
	action := flags.String("action", "", "governed maintenance action")
	confirmDepth := flags.Bool("confirm-depth", false, "apply explicit operator approval only to the GSD planning-depth gate")
	continueUnit := flags.Bool("continue-unit", false, "resume the latest local Pi session bound to the same canonical unit")
	decisionActor := flags.String("decision-actor", "human", "provenance for interactive decisions: human, shepherd, or contract")
	decisionBasis := flags.String("decision-basis", "interactive terminal response", "concise provenance basis for interactive decisions")
	if err := flags.Parse(args[1:]); err != nil {
		return err
	}
	config, err := loadConfig(*configPath)
	if err != nil {
		return err
	}
	if args[0] == "eval" {
		activities, err := telemetry.ReadDirectory(filepath.Join(config.StateDir, "activity", "segments"))
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(telemetry.Evaluate(activities))
	}
	if args[0] == "decisions" {
		if *issue <= 0 {
			return errors.New("--issue is required")
		}
		records, err := decisionlog.Read(filepath.Join(config.StateDir, "decisions"))
		if err != nil {
			return err
		}
		filtered := records[:0]
		for _, record := range records {
			if record.DeliveryID == deliveryID(*issue) {
				filtered = append(filtered, record)
			}
		}
		fmt.Print(decisionlog.Markdown(filtered))
		return nil
	}
	if args[0] == "resume" {
		if *issue <= 0 {
			return errors.New("--issue is required")
		}
		path, err := governedPath(config.StateDir, *decisionPath)
		if err != nil {
			return fmt.Errorf("decision: %w", err)
		}
		raw, err := os.ReadFile(path)
		if err != nil || len(raw) > 64*1024 {
			return errors.New("decision is unreadable or oversized")
		}
		var input decisionInput
		decoder := json.NewDecoder(strings.NewReader(string(raw)))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&input); err != nil {
			return fmt.Errorf("decode decision: %w", err)
		}
		if input.DeliveryID != deliveryID(*issue) || input.Actor != "human" || !input.Approved {
			return errors.New("decision must be an explicit human approval for this delivery")
		}
		authority, err := store.Open(ctx, filepath.Join(config.StateDir, "authority.db"))
		if err != nil {
			return err
		}
		defer authority.Close()
		return authority.ResumeDelivery(ctx, domain.HumanDecision{RunID: input.DeliveryID, Generation: input.Generation, ActorKind: domain.ActorHuman, Approved: true})
	}
	var container *gsd.ContainerConfig
	if config.Runtime == "podman" {
		container = &gsd.ContainerConfig{Engine: "podman", Image: config.ContainerImage,
			GSDStateDir: filepath.Join(config.StateDir, "runtime", "gsd"), PlanningDir: filepath.Join(config.StateDir, "runtime", "planning"),
			AuthFile: config.AuthFile, SettingsFile: filepath.Join(config.GSDHome, "agent", "settings.json"), Network: config.ContainerNetwork,
			PolicyDir: config.PolicyDir, GitCommonDir: config.GitCommonDir,
			SessionsDir: filepath.Join(config.StateDir, "runtime", "sessions"), BackgroundDir: filepath.Join(config.StateDir, "runtime", "bg-shell"),
			BackupDir: filepath.Join(config.StateDir, "runtime", "gsd-backups")}
		if err := gsd.ValidatePinnedContainer(ctx, *container, config.GSDVersion); err != nil {
			return fmt.Errorf("runtime admission: %w", err)
		}
	} else {
		if err := gsd.ValidatePinnedCommand(config.GSDCommand, config.GSDVersion); err != nil {
			return fmt.Errorf("runtime admission: %w", err)
		}
		if err := gsd.ApplyPinnedHeadlessToolPatch(config.GSDCommand, config.GSDVersion); err != nil {
			return fmt.Errorf("runtime compatibility: %w", err)
		}
	}
	if err := gsd.ValidateRuntimeSettings(config.GSDHome, config.WorkDir, config.CoordinatorModel, "high"); err != nil {
		return fmt.Errorf("runtime admission: %w", err)
	}
	runner, err := gsd.NewRunner(gsd.Config{
		Command: config.GSDCommand, WorkDir: config.WorkDir, GSDHome: config.GSDHome, StateDir: config.StateDir,
		Model: config.CoordinatorModel, Thinking: "high",
		Timeout:           time.Duration(config.TimeoutSeconds) * time.Second,
		HeartbeatInterval: time.Duration(config.HeartbeatSeconds) * time.Second,
		MaxEventBytes:     config.MaxEventBytes,
		Container:         container,
	})
	if err != nil {
		return err
	}
	switch args[0] {
	case "query":
		if *issue <= 0 {
			return errors.New("--issue is required because GSD query may reconcile state")
		}
		return runFencedQuery(ctx, runner, config, deliveryID(*issue), *issue)
	case "repair":
		if *issue <= 0 || (*action != "rebuild-markdown" && *action != "finalize-milestone-worktree") {
			return errors.New("repair requires --issue and a supported typed --action")
		}
		return runFencedRepair(ctx, runner, config, deliveryID(*issue), *issue, *action)
	case "run":
		if *issue <= 0 {
			return errors.New("--issue is required")
		}
		if *command == "auto" || *command == "recover" || *command == "new-milestone" {
			return errors.New("generic run permits only one fenced unit; use --command next, discuss, or status")
		}
		return runHeadless(ctx, runner, config, deliveryID(*issue), *issue, "", *command, nil, *confirmDepth, *continueUnit, *decisionActor, *decisionBasis)
	case "start":
		if *issue <= 0 {
			return errors.New("--issue is required")
		}
		path, err := governedPath(config.WorkDir, *contextPath)
		if err != nil {
			return fmt.Errorf("context: %w", err)
		}
		if *auto {
			return errors.New("start --auto is disabled; run one fenced GSD unit at a time")
		}
		file, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("open context: %w", err)
		}
		_, raw, decodeErr := contract.DecodeIssueContext(file, *issue)
		closeErr := file.Close()
		if decodeErr != nil {
			return decodeErr
		}
		if closeErr != nil {
			return closeErr
		}
		if err := materializeContainerContext(config, path, raw); err != nil {
			return fmt.Errorf("materialize protected context: %w", err)
		}
		hash := sha256.Sum256(raw)
		contextHash := "sha256:" + hex.EncodeToString(hash[:])
		if *adoptExisting {
			return adoptExistingDelivery(ctx, runner, config, deliveryID(*issue), *issue, contextHash)
		}
		return runHeadless(ctx, runner, config, deliveryID(*issue), *issue, contextHash, "new-milestone", []string{"--context", path}, *confirmDepth, false, *decisionActor, *decisionBasis)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runFencedRepair(ctx context.Context, runner *gsd.Runner, config fileConfig, deliveryID string, issue int, action string) error {
	if err := os.MkdirAll(config.StateDir, 0o700); err != nil {
		return err
	}
	authority, err := store.Open(ctx, filepath.Join(config.StateDir, "authority.db"))
	if err != nil {
		return err
	}
	defer authority.Close()
	delivery, err := authority.GetDelivery(ctx, deliveryID)
	if err != nil || delivery.Issue != issue || delivery.WorkDir != config.WorkDir {
		return errors.New("repair delivery does not match issue/worktree")
	}
	owner := fmt.Sprintf("repair-%d", time.Now().UTC().UnixNano())
	lease, err := authority.AcquireLease(ctx, deliveryID, owner, time.Now().UTC(), 90*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = authority.ReleaseLease(context.Background(), lease) }()
	switch action {
	case "rebuild-markdown":
		if err := runner.RebuildMarkdown(ctx); err != nil {
			return err
		}
	case "finalize-milestone-worktree":
		if delivery.MilestoneID == "" {
			return errors.New("delivery has no bound milestone worktree")
		}
		snapshot, inspectErr := shepherdgit.Inspect(ctx, config.WorkDir)
		if inspectErr != nil {
			return inspectErr
		}
		if err := shepherdgit.FinalizeInitializingMilestoneWorktree(ctx, config.WorkDir, delivery.MilestoneID, snapshot.HeadSHA); err != nil {
			return err
		}
	}
	if err := authority.CheckLease(ctx, lease, time.Now().UTC()); err != nil {
		return err
	}
	activity, err := telemetry.Open(ctx, filepath.Join(config.StateDir, "activity", "segments"))
	if err != nil {
		return err
	}
	defer activity.Close()
	hash := sha256.Sum256([]byte(owner + ":" + action))
	_, err = activity.Append(ctx, telemetry.Activity{ID: hex.EncodeToString(hash[:]), RunID: deliveryID, Kind: "transition", Status: strings.ReplaceAll(action, "-", "_"), At: time.Now().UTC()})
	return err
}

func adoptExistingDelivery(ctx context.Context, runner *gsd.Runner, config fileConfig, deliveryID string, issue int, contextHash string) error {
	if err := os.MkdirAll(config.StateDir, 0o700); err != nil {
		return err
	}
	authority, err := store.Open(ctx, filepath.Join(config.StateDir, "authority.db"))
	if err != nil {
		return err
	}
	defer authority.Close()
	if err := authority.EnsureDelivery(ctx, store.Delivery{ID: deliveryID, Issue: issue, WorkDir: config.WorkDir, ContextHash: contextHash}); err != nil {
		return err
	}
	owner := fmt.Sprintf("adopt-%d", time.Now().UTC().UnixNano())
	lease, err := authority.AcquireLease(ctx, deliveryID, owner, time.Now().UTC(), 90*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = authority.ReleaseLease(context.Background(), lease) }()
	snapshot, err := runner.Query(ctx)
	if err != nil {
		return err
	}
	if snapshot.MilestoneID == "" {
		return errors.New("no active GSD milestone exists to adopt")
	}
	if err := authority.BindMilestone(ctx, deliveryID, snapshot.MilestoneID); err != nil {
		return err
	}
	if err := authority.PrepareAdoptedDelivery(ctx, deliveryID, snapshot.MilestoneID); err != nil {
		return err
	}
	if err := authority.CheckLease(ctx, lease, time.Now().UTC()); err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(struct {
		DeliveryID  string `json:"delivery_id"`
		MilestoneID string `json:"milestone_id"`
		State       string `json:"state"`
	}{DeliveryID: deliveryID, MilestoneID: snapshot.MilestoneID, State: "adopted"})
}

func runFencedQuery(ctx context.Context, runner *gsd.Runner, config fileConfig, deliveryID string, issue int) error {
	if err := os.MkdirAll(config.StateDir, 0o700); err != nil {
		return err
	}
	authority, err := store.Open(ctx, filepath.Join(config.StateDir, "authority.db"))
	if err != nil {
		return err
	}
	defer authority.Close()
	delivery, err := authority.GetDelivery(ctx, deliveryID)
	if err != nil || delivery.Issue != issue || delivery.WorkDir != config.WorkDir {
		return errors.New("query delivery is not initialized or does not match this issue/worktree")
	}
	owner := fmt.Sprintf("query-%d", time.Now().UTC().UnixNano())
	lease, err := authority.AcquireLease(ctx, deliveryID, owner, time.Now().UTC(), 90*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = authority.ReleaseLease(context.Background(), lease) }()
	snapshot, err := runner.Query(ctx)
	if err != nil {
		return err
	}
	if snapshot.MilestoneID != "" {
		if err := authority.BindMilestone(ctx, deliveryID, snapshot.MilestoneID); err != nil {
			return err
		}
	}
	if err := authority.CheckLease(ctx, lease, time.Now().UTC()); err != nil {
		return err
	}
	return json.NewEncoder(os.Stdout).Encode(snapshot)
}

func loadConfig(path string) (fileConfig, error) {
	if path == "" {
		return fileConfig{}, errors.New("--config is required")
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return fileConfig{}, err
	}
	var config fileConfig
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&config); err != nil {
		return fileConfig{}, fmt.Errorf("decode config: %w", err)
	}
	if !filepath.IsAbs(config.WorkDir) || !filepath.IsAbs(config.GSDHome) || !filepath.IsAbs(config.StateDir) {
		return fileConfig{}, errors.New("work_dir, gsd_home, and state_dir must be absolute")
	}
	if config.CoordinatorModel == "" {
		config.CoordinatorModel = "openai-codex/gpt-5.6-sol"
	}
	if config.GSDVersion == "" {
		config.GSDVersion = "1.11.0"
	}
	if config.TimeoutSeconds <= 0 {
		config.TimeoutSeconds = 3600
	}
	if config.HeartbeatSeconds <= 0 {
		config.HeartbeatSeconds = 15
	}
	if config.MaxEventBytes <= 0 {
		config.MaxEventBytes = 256 * 1024
	}
	if config.Runtime == "" {
		config.Runtime = "host"
	}
	if config.Runtime != "host" && config.Runtime != "podman" {
		return fileConfig{}, errors.New("runtime must be host or podman")
	}
	if config.Runtime == "podman" && (!filepath.IsAbs(config.AuthFile) || !filepath.IsAbs(config.PolicyDir) || !filepath.IsAbs(config.GitCommonDir) || config.ContainerImage == "") {
		return fileConfig{}, errors.New("podman runtime requires absolute auth_file, policy_dir, and git_common_dir plus container_image")
	}
	if within, err := pathWithin(config.WorkDir, config.StateDir); err != nil || within {
		return fileConfig{}, errors.New("state_dir must be outside the worker-controlled work directory")
	}
	return config, nil
}

func runHeadless(ctx context.Context, runner *gsd.Runner, config fileConfig, deliveryID string, issue int, contextHash, command string, args []string, confirmDepth, continueUnit bool, decisionActor, decisionBasis string) error {
	if err := os.MkdirAll(config.StateDir, 0o700); err != nil {
		return err
	}
	authority, err := store.Open(ctx, filepath.Join(config.StateDir, "authority.db"))
	if err != nil {
		return err
	}
	defer authority.Close()
	executionID := fmt.Sprintf("execution-%d", time.Now().UTC().UnixNano())
	if contextHash != "" {
		if err := authority.EnsureDelivery(ctx, store.Delivery{ID: deliveryID, Issue: issue, WorkDir: config.WorkDir, ContextHash: contextHash}); err != nil {
			return err
		}
		if command == "new-milestone" {
			if err := authority.RetryFailedIntake(ctx, deliveryID); err != nil {
				return err
			}
		}
	} else {
		delivery, err := authority.GetDelivery(ctx, deliveryID)
		if err != nil {
			return errors.New("delivery has not been initialized with a validated issue context")
		}
		if delivery.Issue != issue || delivery.WorkDir != config.WorkDir {
			return errors.New("delivery does not match requested issue and work directory")
		}
	}
	lease, err := authority.AcquireLease(ctx, deliveryID, executionID, time.Now().UTC(), time.Duration(config.TimeoutSeconds+60)*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = authority.ReleaseLease(context.Background(), lease) }()
	startSnapshot, err := shepherdgit.Inspect(ctx, config.WorkDir)
	if err != nil {
		return err
	}
	if err := shepherdgit.RequireClean(startSnapshot); err != nil {
		return err
	}
	trustedUnit := command
	before, queryErr := runner.Query(ctx)
	if queryErr != nil {
		return fmt.Errorf("pre-run fenced query failed: %w", queryErr)
	}
	if before.Next.UnitID != "" {
		trustedUnit = before.Next.UnitType + "/" + before.Next.UnitID
	}
	if command == "discuss" {
		if before.Next.Action != "dispatch" || before.Next.UnitType != "discuss-milestone" || before.MilestoneID == "" {
			return errors.New("targeted discuss is allowed only when the canonical next unit is discuss-milestone")
		}
		args = []string{before.MilestoneID}
	}
	if command == "research-milestone" && before.Next.UnitType != command {
		return fmt.Errorf("%s is allowed only when it is the canonical next unit", command)
	}
	if continueUnit {
		if command != "discuss" || config.Runtime == "podman" {
			return errors.New("--continue-unit supports only a local canonical discuss-milestone")
		}
		sessionID, sessionErr := gsd.LatestSessionID(filepath.Join(config.GSDHome, "agent", "sessions"), config.WorkDir)
		if sessionErr != nil {
			return fmt.Errorf("resolve local continuation session: %w", sessionErr)
		}
		args = append(args, "--resume", sessionID)
	}
	if command == "new-milestone" && before.MilestoneID != "" {
		return errors.New("an active GSD milestone already exists; use start --adopt-existing after verifying its issue context")
	}
	if before.MilestoneID != "" {
		if err := authority.BindMilestone(ctx, deliveryID, before.MilestoneID); err != nil {
			return err
		}
	}
	attempt, err := authority.BeginAttempt(ctx, deliveryID, executionID)
	if err != nil {
		return err
	}
	_ = attempt
	finishedAttempt := false
	defer func() {
		if !finishedAttempt {
			_ = authority.FinishAttempt(context.Background(), deliveryID, executionID, domain.RunFailed)
		}
	}()
	activity, err := telemetry.Open(ctx, filepath.Join(config.StateDir, "activity", "segments"))
	if err != nil {
		return err
	}
	defer activity.Close()
	decisions, err := decisionlog.Open(filepath.Join(config.StateDir, "decisions"))
	if err != nil {
		return err
	}
	defer decisions.Close()
	var sequence atomic.Uint64
	var activityMu sync.Mutex
	var activityErr error
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	appendActivity := func(kind, status, unit, model, tool string, at time.Time) {
		activityMu.Lock()
		defer activityMu.Unlock()
		if activityErr != nil {
			return
		}
		seq := sequence.Add(1)
		idInput := executionID + ":" + strconv.FormatUint(seq, 10) + ":" + kind
		hash := sha256.Sum256([]byte(idInput))
		_, activityErr = activity.Append(runCtx, telemetry.Activity{
			ID: hex.EncodeToString(hash[:]), RunID: deliveryID, UnitID: unit, Kind: kind,
			Status: status, Model: model, Tool: tool, At: at,
		})
		if activityErr != nil {
			cancel()
		}
	}
	var observedModel, observedThinking string
	observer := gsd.Observer{
		Event: func(event gsd.Event) {
			if event.Model != "" {
				observedModel = event.Model
			}
			if event.Thinking != "" {
				observedThinking = event.Thinking
			}
			kind := mapEventKind(event.Kind)
			appendActivity(kind, event.Status, trustedUnit, event.Model, event.Tool, event.At)
			fmt.Fprintf(os.Stderr, "%s kind=%s unit=%s tool=%s model=%s\n", event.At.Format(time.RFC3339), event.Kind, trustedUnit, event.Tool, event.Model)
		},
		Heartbeat: func(heartbeat gsd.Heartbeat) {
			appendActivity("heartbeat", "alive", "", "", heartbeat.InFlightTool, heartbeat.At)
			fmt.Fprintf(os.Stderr, "%s heartbeat alive=%t in_flight_tool=%s\n", heartbeat.At.Format(time.RFC3339), heartbeat.ProcessAlive, heartbeat.InFlightTool)
		},
		Question: func(questionCtx context.Context, question gsd.Question) (gsd.UIResponse, error) {
			var response gsd.UIResponse
			var responseErr error
			if confirmDepth {
				if response, approved := approveDepthQuestion(question); approved {
					fmt.Fprintln(os.Stderr, "HUMAN GATE [select] planning depth approved by explicit --confirm-depth operator flag")
					if err := appendDecision(decisions, deliveryID, executionID, trustedUnit, question, response, decisionActor, decisionBasis); err != nil {
						return gsd.UIResponse{Cancelled: true}, err
					}
					appendActivity("decision", decisionActor, trustedUnit, "", question.ID, time.Now().UTC())
					return response, nil
				}
			}
			response, responseErr = terminalQuestion(questionCtx, question)
			if responseErr != nil || response.Cancelled {
				return response, responseErr
			}
			if err := appendDecision(decisions, deliveryID, executionID, trustedUnit, question, response, decisionActor, decisionBasis); err != nil {
				return gsd.UIResponse{Cancelled: true}, err
			}
			appendActivity("decision", decisionActor, trustedUnit, "", question.ID, time.Now().UTC())
			return response, nil
		},
	}
	appendActivity("run.started", "running", trustedUnit, "", "", time.Now().UTC())
	result := runner.Run(runCtx, command, args, observer)
	if result.Terminal == gsd.TerminalSuccess && command == "new-milestone" && config.Runtime == "podman" {
		workflow, queryErr := runner.Query(ctx)
		if queryErr != nil || workflow.MilestoneID == "" {
			result.Terminal = gsd.TerminalError
			result.Err = errors.New("completed milestone intake did not expose a canonical milestone")
		} else if repairErr := shepherdgit.FinalizeInitializingMilestoneWorktree(ctx, config.WorkDir, workflow.MilestoneID, startSnapshot.HeadSHA); repairErr != nil {
			result.Terminal = gsd.TerminalError
			result.Err = fmt.Errorf("finalize milestone worktree: %w", repairErr)
		} else {
			appendActivity("transition", "finalized_milestone_worktree", trustedUnit, "", "", time.Now().UTC())
		}
	}
	if restoreErr := shepherdgit.RestoreIndex(ctx, config.WorkDir); restoreErr != nil {
		result.Terminal = gsd.TerminalError
		result.Err = restoreErr
	}
	if observedModel == "" || observedThinking == "" {
		sessionsDir := filepath.Join(config.GSDHome, "agent", "sessions")
		if config.Runtime == "podman" {
			sessionsDir = filepath.Join(config.StateDir, "runtime", "sessions")
		}
		model, thinking, identityErr := gsd.ReadSessionIdentity(sessionsDir)
		if identityErr == nil {
			observedModel, observedThinking = model, thinking
			appendActivity("model.activity", "session_metadata", trustedUnit, model, "", time.Now().UTC())
		}
	}
	postPhase := ""
	activityMu.Lock()
	spoolErr := activityErr
	activityMu.Unlock()
	if spoolErr != nil {
		result.Terminal = gsd.TerminalError
		result.Err = fmt.Errorf("durable activity append failed: %w", spoolErr)
	}
	if result.Terminal != gsd.TerminalRejected {
		snapshot, queryErr := runner.Query(ctx)
		if queryErr != nil {
			result.Terminal = gsd.TerminalError
			result.Err = fmt.Errorf("post-run query reconciliation failed: %w", queryErr)
		} else {
			postPhase = snapshot.Phase
			if snapshot.MilestoneID != "" {
				if bindErr := authority.BindMilestone(ctx, deliveryID, snapshot.MilestoneID); bindErr != nil {
					result.Terminal = gsd.TerminalError
					result.Err = bindErr
				}
			}
			terminal, reconcileErr := gsd.Reconcile(command, result, before, snapshot)
			result.Terminal = terminal
			if reconcileErr != nil {
				result.Err = reconcileErr
			}
		}
	}
	if result.Terminal == gsd.TerminalSuccess && (observedModel != config.CoordinatorModel || observedThinking != "high") {
		result.Terminal = gsd.TerminalError
		result.Err = fmt.Errorf("effective runtime identity was not observed as %s/high", config.CoordinatorModel)
	}
	if err := authority.CheckLease(ctx, lease, time.Now().UTC()); err != nil {
		result.Terminal = gsd.TerminalError
		result.Err = err
	}
	endSnapshot, snapshotErr := shepherdgit.Inspect(ctx, config.WorkDir)
	if snapshotErr != nil || shepherdgit.RequireClean(endSnapshot) != nil {
		result.Terminal = gsd.TerminalError
		result.Err = errors.New("worktree must be clean after a governed unit")
	} else if err := authority.RecordAttemptHeads(ctx, deliveryID, executionID, startSnapshot.HeadSHA, endSnapshot.HeadSHA); err != nil {
		result.Terminal = gsd.TerminalError
		result.Err = err
	}
	appendActivity("run.terminal", string(result.Terminal), "", "", "", result.Ended)
	activityMu.Lock()
	terminalSpoolErr := activityErr
	activityMu.Unlock()
	if terminalSpoolErr != nil {
		result.Terminal = gsd.TerminalError
		result.Err = fmt.Errorf("durable terminal append failed: %w", terminalSpoolErr)
	}
	targetState := domain.RunFailed
	if result.Terminal == gsd.TerminalBlocked {
		targetState = domain.RunBlocked
	}
	if result.Terminal == gsd.TerminalSuccess {
		targetState = domain.RunReady
	}
	if result.Terminal == gsd.TerminalSuccess && postPhase == "complete" {
		targetState = domain.RunHumanGate
	}
	if err := authority.FinishAttempt(ctx, deliveryID, executionID, targetState); err != nil {
		result.Terminal = gsd.TerminalError
		result.Err = err
	} else {
		finishedAttempt = true
	}
	encoded, marshalErr := json.Marshal(struct {
		DeliveryID  string       `json:"delivery_id"`
		ExecutionID string       `json:"execution_id"`
		Terminal    gsd.Terminal `json:"terminal"`
		ExitCode    int          `json:"exit_code"`
	}{DeliveryID: deliveryID, ExecutionID: executionID, Terminal: result.Terminal, ExitCode: result.ExitCode})
	if marshalErr != nil {
		return marshalErr
	}
	fmt.Println(string(encoded))
	terminalErr := result.Err
	if terminalErr == nil {
		terminalErr = errors.New("runtime did not provide an error")
	}
	if diagnostic := terminalDiagnostic(result.Stderr); diagnostic != "" {
		terminalErr = fmt.Errorf("%w; runtime: %s", terminalErr, diagnostic)
	}
	if result.Terminal == gsd.TerminalBlocked {
		return commandExitError{code: 10, err: fmt.Errorf("GSD terminal=%s exit=%d: %w", result.Terminal, result.ExitCode, terminalErr)}
	}
	if result.Terminal != gsd.TerminalSuccess {
		return fmt.Errorf("GSD terminal=%s exit=%d: %w", result.Terminal, result.ExitCode, terminalErr)
	}
	return nil
}

func terminalDiagnostic(stderr string) string {
	lines := strings.Split(stderr, "\n")
	for index := len(lines) - 1; index >= 0; index-- {
		line := strings.TrimSpace(strings.Map(func(r rune) rune {
			if r < 0x20 || r == 0x7f {
				return -1
			}
			return r
		}, lines[index]))
		if line == "" {
			continue
		}
		if len(line) > 512 {
			line = line[:512]
		}
		return line
	}
	return ""
}

type commandExitError struct {
	code int
	err  error
}

func (e commandExitError) Error() string { return e.err.Error() }
func (e commandExitError) Unwrap() error { return e.err }
func (e commandExitError) ExitCode() int { return e.code }

func deliveryID(issue int) string { return "issue-" + strconv.Itoa(issue) }

func appendDecision(store *decisionlog.Store, deliveryID, executionID, unitID string, question gsd.Question, response gsd.UIResponse, actorValue, basis string) error {
	actor := decisionlog.Actor(strings.TrimSpace(actorValue))
	if actor != decisionlog.ActorHuman && actor != decisionlog.ActorShepherd && actor != decisionlog.ActorContract {
		return errors.New("decision actor must be human, shepherd, or contract")
	}
	answer := strings.TrimSpace(response.Value)
	if answer == "" && response.Confirmed != nil {
		answer = strconv.FormatBool(*response.Confirmed)
	}
	questionID := strings.TrimSpace(question.ID)
	if questionID == "" {
		questionID = "request-" + strings.TrimSpace(question.RequestID)
	}
	idHash := sha256.Sum256([]byte(executionID + ":" + questionID))
	return store.Append(context.Background(), decisionlog.Record{
		ID: hex.EncodeToString(idHash[:]), DeliveryID: deliveryID, ExecutionID: executionID, UnitID: unitID,
		QuestionID: questionID, Question: strings.TrimSpace(question.Title), Answer: answer,
		Actor: actor, Basis: strings.TrimSpace(basis), At: time.Now().UTC(),
	})
}

func approveDepthQuestion(question gsd.Question) (gsd.UIResponse, bool) {
	cancelled := gsd.UIResponse{Cancelled: true}
	const prefix = "depth_verification_M"
	const suffix = "_confirm"
	id := strings.TrimSpace(question.ID)
	milestone := strings.TrimSuffix(strings.TrimPrefix(id, prefix), suffix)
	if question.Method != "select" || !strings.HasPrefix(id, prefix) || !strings.HasSuffix(id, suffix) ||
		len(milestone) < 3 || strings.Trim(milestone[:3], "0123456789") != "" ||
		(len(milestone) > 3 && (milestone[3] != '-' || len(milestone) == 4 || strings.IndexFunc(milestone[4:], func(r rune) bool {
			return (r < 'a' || r > 'z') && (r < '0' || r > '9')
		}) >= 0)) {
		return cancelled, false
	}
	for _, option := range question.Options {
		normalized := strings.ToLower(strings.TrimSpace(option))
		if strings.HasPrefix(normalized, "confirm") || strings.HasPrefix(normalized, "depth verified") {
			return gsd.UIResponse{Value: option}, true
		}
	}
	return cancelled, false
}

func terminalQuestion(ctx context.Context, question gsd.Question) (gsd.UIResponse, error) {
	fmt.Fprintf(os.Stderr, "\nHUMAN GATE [%s] %s\n", question.Method, question.Title)
	for i, option := range question.Options {
		fmt.Fprintf(os.Stderr, "  %d. %s\n", i+1, option)
	}
	fmt.Fprint(os.Stderr, "Response (number, yes/no, text, or cancel): ")
	type inputResult struct {
		line string
		err  error
	}
	input := make(chan inputResult, 1)
	go func() {
		line, err := bufio.NewReader(os.Stdin).ReadString('\n')
		input <- inputResult{line: line, err: err}
	}()
	var line string
	select {
	case <-ctx.Done():
		return gsd.UIResponse{Cancelled: true}, ctx.Err()
	case read := <-input:
		if read.err != nil {
			return gsd.UIResponse{Cancelled: true}, errors.New("explicit human response unavailable")
		}
		line = read.line
	}
	answer := strings.TrimSpace(line)
	if answer == "" || strings.EqualFold(answer, "cancel") {
		return gsd.UIResponse{Cancelled: true}, nil
	}
	if question.Method == "confirm" {
		confirmed := strings.EqualFold(answer, "yes") || strings.EqualFold(answer, "y")
		return gsd.UIResponse{Confirmed: &confirmed}, nil
	}
	if question.Method == "select" {
		choice, err := strconv.Atoi(answer)
		if err != nil || choice < 1 || choice > len(question.Options) {
			return gsd.UIResponse{Cancelled: true}, errors.New("selection must be a listed number")
		}
		return gsd.UIResponse{Value: question.Options[choice-1]}, nil
	}
	return gsd.UIResponse{Value: answer}, nil
}

func governedPath(root, path string) (string, error) {
	if path == "" {
		return "", errors.New("path is required")
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	clean, err := filepath.EvalSymlinks(filepath.Clean(path))
	if err != nil {
		return "", errors.New("path cannot be resolved")
	}
	canonicalRoot, err := filepath.EvalSymlinks(filepath.Clean(root))
	if err != nil {
		return "", errors.New("work directory cannot be resolved")
	}
	relative, err := filepath.Rel(canonicalRoot, clean)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", errors.New("path escapes work directory")
	}
	info, err := os.Stat(clean)
	if err != nil || !info.Mode().IsRegular() {
		return "", errors.New("path must be an existing regular file")
	}
	return clean, nil
}

func pathWithin(root, path string) (bool, error) {
	relative, err := filepath.Rel(filepath.Clean(root), filepath.Clean(path))
	if err != nil {
		return false, err
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)), nil
}

func materializeContainerContext(config fileConfig, contextPath string, raw []byte) error {
	if config.Runtime != "podman" {
		return nil
	}
	planningRoot := filepath.Join(config.WorkDir, ".planning")
	within, err := pathWithin(planningRoot, contextPath)
	if err != nil || !within {
		return err
	}
	relative, err := filepath.Rel(planningRoot, contextPath)
	if err != nil || relative == "." || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return errors.New("context has an unsafe planning-relative path")
	}
	target := filepath.Join(config.StateDir, "runtime", "planning", relative)
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil {
		return err
	}
	return os.WriteFile(target, raw, 0o600)
}

func mapEventKind(kind gsd.EventKind) string {
	switch kind {
	case gsd.EventAgentStart:
		return "transition"
	case gsd.EventTurnStart:
		return "unit.started"
	case gsd.EventToolStart:
		return "tool.started"
	case gsd.EventToolEnd:
		return "tool.terminal"
	case gsd.EventAgentEnd:
		return "unit.terminal"
	case gsd.EventModelSelect, gsd.EventThinkingSelect:
		return "model.activity"
	default:
		return "transition"
	}
}
