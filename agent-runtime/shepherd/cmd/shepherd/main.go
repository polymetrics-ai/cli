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
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/authority"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/contract"
	decisionlog "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/decision"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/domain"
	shepherdgit "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/git"
	shepherdgithub "github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/github"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/gsd"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/store"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/supervisor"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/telemetry"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/validation"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/workspace"
)

const version = "0.1.0"

const defaultMaxEventBytes = 8 * 1024 * 1024

const defaultWriteScopePollInterval = 500 * time.Millisecond

var independentValidatorFactory = func(_ *gsd.Runner, config fileConfig) validation.Validator {
	return validation.GSDValidator{
		Command: config.PiCommand, GSDHome: config.GSDHome, StateDir: config.StateDir,
		Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
	}
}

type fileConfig struct {
	GSDCommand          []string `json:"gsd_command"`
	PiCommand           []string `json:"pi_command"`
	WorkDir             string   `json:"work_dir"`
	GSDHome             string   `json:"gsd_home"`
	StateDir            string   `json:"state_dir"`
	CoordinatorModel    string   `json:"coordinator_model"`
	ImplementationModel string   `json:"implementation_model"`
	GSDVersion          string   `json:"gsd_version"`
	TimeoutSeconds      int      `json:"timeout_seconds"`
	HeartbeatSeconds    int      `json:"heartbeat_seconds"`
	MaxEventBytes       int      `json:"max_event_bytes"`
	MaxUnitAttempts     int      `json:"max_unit_attempts"`
	Repository          string   `json:"repository"`
	PullRequest         int      `json:"pull_request"`
	Runtime             string   `json:"runtime"`
	ContainerImage      string   `json:"container_image"`
	AuthFile            string   `json:"auth_file"`
	ContainerNetwork    string   `json:"container_network"`
	PolicyDir           string   `json:"policy_dir"`
	GitCommonDir        string   `json:"git_common_dir"`
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
		return errors.New("usage: shepherd <version|query|eval|repair|resume|run|start|supervise> [flags]")
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
	publish := flags.Bool("publish", false, "synchronize the decision ledger to the bound pull request")
	recordDecision := flags.Bool("record", false, "append an attributed operational decision and publish it")
	decisionUnit := flags.String("decision-unit", "", "canonical unit for a recorded operational decision")
	decisionQuestion := flags.String("decision-question", "", "bounded decision being made")
	decisionAnswer := flags.String("decision-answer", "", "bounded selected action")
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
		if *recordDecision {
			if !*publish {
				return errors.New("--record requires --publish so the PR remains the visible decision summary")
			}
			store, err := decisionlog.Open(filepath.Join(config.StateDir, "decisions"))
			if err != nil {
				return err
			}
			defer store.Close()
			return recordOperationalDecision(ctx, store, shepherdgithub.NewCLIClient(), shepherdgithub.Target{
				Repository: config.Repository, PullRequest: config.PullRequest, DeliveryID: deliveryID(*issue),
			}, deliveryID(*issue), *decisionUnit, *decisionQuestion, *decisionAnswer, *decisionActor, *decisionBasis)
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
		summary := decisionlog.Markdown(filtered)
		if *publish {
			if err := publishDecisions(ctx, config, deliveryID(*issue), summary); err != nil {
				return err
			}
		}
		fmt.Print(summary)
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
		if err := gsd.ApplyPinnedPromptToolPatch(config.GSDCommand, config.GSDHome, config.GSDVersion); err != nil {
			return fmt.Errorf("runtime compatibility: %w", err)
		}
	}
	if err := gsd.NormalizeRuntimeSettings(config.GSDHome, config.CoordinatorModel, config.ImplementationModel, "high"); err != nil {
		return fmt.Errorf("runtime normalization: %w", err)
	}
	if err := gsd.ValidateRuntimeSettings(config.GSDHome, config.WorkDir, config.CoordinatorModel, "high"); err != nil {
		return fmt.Errorf("runtime admission: %w", err)
	}
	if err := gsd.ValidateModelPreferences(config.GSDHome, config.WorkDir, config.CoordinatorModel, config.ImplementationModel, "high"); err != nil {
		return fmt.Errorf("runtime admission: %w", err)
	}
	registry := gsd.BuiltinUnitRegistry()
	if config.Runtime == "host" {
		loadedRegistry, err := gsd.LoadPinnedUnitRegistry(config.GSDCommand, config.GSDHome, config.GSDVersion)
		if err != nil {
			return fmt.Errorf("runtime admission: %w", err)
		}
		registry = loadedRegistry
	}
	selectedModel, err := launchModelForCommand(config, registry, *command)
	if err != nil {
		return err
	}
	runner, err := gsd.NewRunner(gsd.Config{
		Command: config.GSDCommand, WorkDir: config.WorkDir, GSDHome: config.GSDHome, StateDir: config.StateDir,
		Model: selectedModel, Thinking: "high",
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
		return runHeadless(ctx, runner, config, registry, deliveryID(*issue), *issue, "", *command, nil, *confirmDepth, *continueUnit, *decisionActor, *decisionBasis)
	case "supervise":
		if *issue <= 0 {
			return errors.New("--issue is required")
		}
		if *contextPath == "" {
			return errors.New("--context is required")
		}
		return runSupervise(ctx, runner, config, registry, *issue, *contextPath, *confirmDepth, *decisionActor, *decisionBasis)
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
		if err := materializeProtectedIssueContext(config.StateDir, *issue, raw); err != nil {
			return fmt.Errorf("materialize protected issue context: %w", err)
		}
		if err := materializeContainerContext(config, path, raw); err != nil {
			return fmt.Errorf("materialize protected context: %w", err)
		}
		hash := sha256.Sum256(raw)
		contextHash := "sha256:" + hex.EncodeToString(hash[:])
		if *adoptExisting {
			return adoptExistingDelivery(ctx, runner, config, deliveryID(*issue), *issue, contextHash)
		}
		return runHeadless(ctx, runner, config, registry, deliveryID(*issue), *issue, contextHash, "new-milestone", []string{"--context", path}, *confirmDepth, false, *decisionActor, *decisionBasis)
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func runSupervise(ctx context.Context, runner *gsd.Runner, config fileConfig, registry gsd.UnitRegistry, issue int, contextPath string, confirmDepth bool, decisionActor, decisionBasis string) error {
	path, err := governedPath(config.WorkDir, contextPath)
	if err != nil {
		return fmt.Errorf("context: %w", err)
	}
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open context: %w", err)
	}
	_, raw, decodeErr := contract.DecodeIssueContext(file, issue)
	closeErr := file.Close()
	if decodeErr != nil {
		return decodeErr
	}
	if closeErr != nil {
		return closeErr
	}
	if err := materializeProtectedIssueContext(config.StateDir, issue, raw); err != nil {
		return fmt.Errorf("materialize protected issue context: %w", err)
	}
	if err := materializeContainerContext(config, path, raw); err != nil {
		return fmt.Errorf("materialize protected context: %w", err)
	}
	hash := sha256.Sum256(raw)
	contextHash := "sha256:" + hex.EncodeToString(hash[:])
	authority, err := store.Open(ctx, filepath.Join(config.StateDir, "authority.db"))
	if err != nil {
		return err
	}
	defer authority.Close()
	if _, _, err := ensureIssueDelivery(ctx, authority, config, issue, contextHash); err != nil {
		return err
	}
	deliveryID := deliveryID(issue)
	if err := consumeGitHubDecisionReplies(ctx, authority, shepherdgithub.NewCLIClient(), config, deliveryID, issue); err != nil {
		return err
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		snapshot, err := runner.Query(ctx)
		if err != nil {
			return fmt.Errorf("supervise query: %w", err)
		}
		decision, err := supervisor.DecideWithRegistry(snapshot, registry)
		if err != nil {
			return fmt.Errorf("supervise policy: %w", err)
		}
		switch decision.Kind {
		case supervisor.DecisionFinalGate:
			return emitSuperviseStatus(deliveryID, decision)
		case supervisor.DecisionBlocked:
			if emitErr := emitSuperviseStatus(deliveryID, decision); emitErr != nil {
				return emitErr
			}
			return commandExitError{code: 10, err: errors.New(decision.Reason)}
		case supervisor.DecisionDispatch:
			unitRunner := runner
			expectedModel, modelErr := modelForUnitType(config, registry, snapshot.Next.UnitType)
			if modelErr != nil {
				return modelErr
			}
			unitRunner, modelErr = runner.WithModel(expectedModel)
			if modelErr != nil {
				return modelErr
			}
			err := runHeadless(ctx, unitRunner, config, registry, deliveryID, issue, "", decision.Command, nil, confirmDepth, false, decisionActor, decisionBasis)
			if err == nil {
				continue
			}
			if isAutomaticallyRetryable(err) {
				continue
			}
			blocked := decision
			blocked.Kind = supervisor.DecisionBlocked
			blocked.Reason = classifyUnitFailure(gsd.Result{Terminal: gsd.TerminalError, Err: err})
			if emitErr := emitSuperviseStatus(deliveryID, blocked); emitErr != nil {
				return emitErr
			}
			return commandExitError{code: 10, err: err}
		default:
			return fmt.Errorf("unknown supervise decision %q", decision.Kind)
		}
	}
}

func requireMatchingAttemptDispatch(canonical, attempt gsd.WorkflowSnapshot) error {
	if canonical.MilestoneID != attempt.MilestoneID || canonical.Phase != attempt.Phase || canonical.Next.Action != attempt.Next.Action || canonical.Next.UnitType != attempt.Next.UnitType || canonical.Next.UnitID != attempt.Next.UnitID {
		return fmt.Errorf("attempt GSD state does not match canonical dispatch: canonical=%s/%s %s %s/%s attempt=%s/%s %s %s/%s", canonical.MilestoneID, canonical.Phase, canonical.Next.Action, canonical.Next.UnitType, canonical.Next.UnitID, attempt.MilestoneID, attempt.Phase, attempt.Next.Action, attempt.Next.UnitType, attempt.Next.UnitID)
	}
	return nil
}

func emitSuperviseStatus(deliveryID string, decision supervisor.Decision) error {
	encoded, err := json.Marshal(struct {
		DeliveryID string                  `json:"delivery_id"`
		Status     supervisor.DecisionKind `json:"status"`
		Reason     string                  `json:"reason"`
		Phase      string                  `json:"phase"`
		NextAction string                  `json:"next_action"`
		Unit       string                  `json:"unit,omitempty"`
	}{
		DeliveryID: deliveryID,
		Status:     decision.Kind,
		Reason:     decision.Reason,
		Phase:      decision.Snapshot.Phase,
		NextAction: decision.Snapshot.Next.Action,
		Unit:       decision.Unit,
	})
	if err != nil {
		return err
	}
	fmt.Println(string(encoded))
	return nil
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
	if _, _, err := ensureIssueDelivery(ctx, authority, config, issue, contextHash); err != nil {
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
	if len(config.PiCommand) != 1 || !filepath.IsAbs(config.PiCommand[0]) || strings.ContainsAny(config.PiCommand[0], "\r\n\x00") {
		return fileConfig{}, errors.New("pi_command must contain one allowlisted absolute Pi executable")
	}
	piExecutable, err := filepath.EvalSymlinks(config.PiCommand[0])
	if err != nil {
		return fileConfig{}, fmt.Errorf("resolve pi_command: %w", err)
	}
	piInfo, err := os.Stat(piExecutable)
	if err != nil || !piInfo.Mode().IsRegular() || piInfo.Mode()&0o111 == 0 {
		return fileConfig{}, errors.New("pi_command must resolve to an executable regular file")
	}
	config.PiCommand[0] = piExecutable
	if config.CoordinatorModel == "" {
		config.CoordinatorModel = "openai-codex/gpt-5.6-sol"
	}
	if config.ImplementationModel == "" {
		config.ImplementationModel = "openai-codex/gpt-5.5"
	}
	if config.CoordinatorModel != "openai-codex/gpt-5.6-sol" || config.ImplementationModel != "openai-codex/gpt-5.5" {
		return fileConfig{}, errors.New("coordinator_model must be gpt-5.6-sol and implementation_model must be gpt-5.5")
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
		config.MaxEventBytes = defaultMaxEventBytes
	}
	if config.MaxUnitAttempts <= 0 {
		config.MaxUnitAttempts = 3
	}
	if config.MaxUnitAttempts > 20 {
		return fileConfig{}, errors.New("max_unit_attempts must be between 1 and 20")
	}
	if config.Runtime == "" {
		config.Runtime = "host"
	}
	if config.Runtime != "host" && config.Runtime != "podman" {
		return fileConfig{}, errors.New("runtime must be host or podman")
	}
	if config.Repository == "" || config.PullRequest <= 0 {
		return fileConfig{}, errors.New("repository and pull_request are required for decision publication")
	}
	if err := shepherdgithub.ValidateTarget(shepherdgithub.Target{Repository: config.Repository, PullRequest: config.PullRequest, DeliveryID: "issue-1"}); err != nil {
		return fileConfig{}, fmt.Errorf("decision publication target: %w", err)
	}
	if config.Runtime == "podman" && (!filepath.IsAbs(config.AuthFile) || !filepath.IsAbs(config.PolicyDir) || !filepath.IsAbs(config.GitCommonDir) || config.ContainerImage == "") {
		return fileConfig{}, errors.New("podman runtime requires absolute auth_file, policy_dir, and git_common_dir plus container_image")
	}
	if within, err := pathWithin(config.WorkDir, config.StateDir); err != nil || within {
		return fileConfig{}, errors.New("state_dir must be outside the worker-controlled work directory")
	}
	return config, nil
}

func runHeadless(ctx context.Context, runner *gsd.Runner, config fileConfig, registry gsd.UnitRegistry, deliveryID string, issue int, contextHash, command string, args []string, confirmDepth, continueUnit bool, decisionActor, decisionBasis string) error {
	if err := os.MkdirAll(config.StateDir, 0o700); err != nil {
		return err
	}
	authority, err := store.Open(ctx, filepath.Join(config.StateDir, "authority.db"))
	if err != nil {
		return err
	}
	defer authority.Close()
	executionID := fmt.Sprintf("execution-%d", time.Now().UTC().UnixNano())
	var delivery store.Delivery
	if contextHash != "" {
		if _, _, err := ensureIssueDelivery(ctx, authority, config, issue, contextHash); err != nil {
			return err
		}
		if command == "new-milestone" {
			if err := authority.RetryFailedIntake(ctx, deliveryID); err != nil {
				return err
			}
		}
	}
	delivery, err = authority.GetDelivery(ctx, deliveryID)
	if err != nil {
		return errors.New("delivery has not been initialized with a validated issue context")
	}
	if delivery.Issue != issue || delivery.WorkDir != config.WorkDir {
		return errors.New("delivery does not match requested issue and work directory")
	}
	issueContext, err := loadProtectedIssueContext(config.StateDir, issue, delivery.ContextHash)
	if err != nil {
		return fmt.Errorf("protected issue context: %w", err)
	}
	if err := validateDeliveryInvocation(delivery, config, issueContext); err != nil {
		return err
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
	if gsd.IsCanonicalUnitCommand(command) && before.Next.UnitType != command {
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
	expectedModel, err := expectedModelForObservedUnit(config, registry, command, before.Next.UnitType)
	if err != nil {
		return err
	}
	if modelRunner, err := runner.WithModel(expectedModel); err != nil {
		return err
	} else {
		runner = modelRunner
	}
	attempt, err := authority.BeginAttempt(ctx, deliveryID, executionID)
	if err != nil {
		return err
	}
	finishedAttempt := false
	defer func() {
		if !finishedAttempt {
			_ = authority.FinishAttempt(context.Background(), deliveryID, executionID, domain.RunFailed)
		}
	}()
	unitAttemptKey := store.UnitAttemptKey{DeliveryID: deliveryID, Generation: attempt.Generation,
		UnitID: trustedUnit, HeadSHA: startSnapshot.HeadSHA}
	unitAttempt, err := authority.BeginUnitAttempt(ctx, unitAttemptKey, int64(config.MaxUnitAttempts))
	if err != nil {
		target := domain.RunFailed
		if errors.Is(err, store.ErrRetryBudgetExhausted) {
			target = domain.RunAwaitingDecision
		}
		if finishErr := authority.FinishAttempt(ctx, deliveryID, executionID, target); finishErr != nil {
			return errors.Join(err, finishErr)
		}
		finishedAttempt = true
		if errors.Is(err, store.ErrRetryBudgetExhausted) {
			return commandExitError{code: 10, err: err}
		}
		return err
	}
	unitAttemptFinished := false
	defer func() {
		if !unitAttemptFinished {
			_ = authority.FinishUnitAttempt(context.Background(), unitAttemptKey, "controller_interrupted")
		}
	}()
	executionWorkDir := config.WorkDir
	executionRunner := runner
	var attemptWorktree *workspace.AttemptWorktree
	var attemptManager *workspace.Manager
	if gsd.IsCanonicalUnitCommand(command) && config.Runtime != "host" {
		return errors.New("canonical unit supervision requires host runtime disposable attempt worktrees")
	}
	if gsd.IsCanonicalUnitCommand(command) && config.Runtime == "host" {
		manager, managerErr := workspace.NewManager(config.WorkDir, filepath.Join(config.StateDir, "attempt-worktrees"))
		if managerErr != nil {
			return managerErr
		}
		attemptTree, createErr := manager.Create(ctx, workspace.AttemptIdentity{DeliveryID: deliveryID, Generation: attempt.Generation, UnitID: trustedUnit, Attempt: unitAttempt.Attempts, BaseHead: startSnapshot.HeadSHA})
		if createErr != nil {
			return createErr
		}
		if prepareErr := manager.PrepareGSDState(ctx, attemptTree); prepareErr != nil {
			return fmt.Errorf("prepare attempt GSD state: %w", prepareErr)
		}
		attemptWorktree = &attemptTree
		attemptManager = manager
		executionWorkDir = attemptTree.Root
		executionRunner, err = runner.WithWorkDir(executionWorkDir)
		if err != nil {
			return err
		}
		attemptSnapshot, queryErr := executionRunner.Query(ctx)
		if queryErr != nil {
			return fmt.Errorf("pre-dispatch attempt query failed: %w", queryErr)
		}
		if err := requireMatchingAttemptDispatch(before, attemptSnapshot); err != nil {
			return err
		}
	}
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
	decisionPublisher := shepherdgithub.NewCLIClient()
	decisionTarget := shepherdgithub.Target{Repository: config.Repository, PullRequest: config.PullRequest, DeliveryID: deliveryID}
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
			if event.Kind == gsd.EventModelSelect && event.Model != "" && observedModel == "" {
				observedModel = event.Model
			}
			if event.Thinking != "" && observedThinking == "" {
				observedThinking = event.Thinking
			}
			kind := mapEventKind(event.Kind)
			appendActivity(kind, event.Status, trustedUnit, event.Model, event.Tool, event.At)
			fmt.Fprintf(os.Stderr, "%s kind=%s unit=%s tool=%s model=%s\n", event.At.Format(time.RFC3339), event.Kind, trustedUnit, event.Tool, event.Model)
		},
		Heartbeat: func(heartbeat gsd.Heartbeat) {
			progress := fmt.Sprintf("%s children=%d turns=%d", heartbeat.InFlightTool, heartbeat.ChildCount, heartbeat.ChildTurns)
			appendActivity("heartbeat", heartbeat.ChildStatus, "", "", strings.TrimSpace(progress), heartbeat.At)
			fmt.Fprintf(os.Stderr, "%s heartbeat alive=%t in_flight_tool=%s child_status=%s children=%d turns=%d\n",
				heartbeat.At.Format(time.RFC3339), heartbeat.ProcessAlive, heartbeat.InFlightTool,
				heartbeat.ChildStatus, heartbeat.ChildCount, heartbeat.ChildTurns)
		},
		Question: func(questionCtx context.Context, question gsd.Question) (gsd.UIResponse, error) {
			var response gsd.UIResponse
			var responseErr error
			if confirmDepth {
				if response, approved := approveDepthQuestion(question); approved {
					fmt.Fprintf(os.Stderr, "DECISION GATE [select] actor=%s planning depth approved by explicit --confirm-depth operator flag\n", decisionActor)
					if err := appendAndPublishDecision(questionCtx, decisions, decisionPublisher, decisionTarget, deliveryID, executionID, trustedUnit, question, response, decisionActor, decisionBasis); err != nil {
						return gsd.UIResponse{Cancelled: true}, err
					}
					appendActivity("decision", decisionActor, trustedUnit, "", question.ID, time.Now().UTC())
					return response, nil
				}
			}
			response, responseErr = terminalQuestion(questionCtx, question, decisionActor)
			if responseErr != nil || response.Cancelled {
				return response, responseErr
			}
			if err := appendAndPublishDecision(questionCtx, decisions, decisionPublisher, decisionTarget, deliveryID, executionID, trustedUnit, question, response, decisionActor, decisionBasis); err != nil {
				return gsd.UIResponse{Cancelled: true}, err
			}
			appendActivity("decision", decisionActor, trustedUnit, "", question.ID, time.Now().UTC())
			return response, nil
		},
	}
	appendActivity("run.started", "running", trustedUnit, "", "", time.Now().UTC())
	preReconcile, err := gsd.ReconcileOrphanedSubagents(config.GSDHome, executionWorkDir, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("reconcile orphaned subagents: %w", err)
	}
	if preReconcile.InterruptedRuns > 0 {
		appendActivity("transition", "interrupted_orphaned_subagents", trustedUnit, "", "", time.Now().UTC())
	}
	var scopeMu sync.Mutex
	var scopeViolation error
	var scopeMonitorDone chan struct{}
	var stopScopeMonitor context.CancelFunc
	if gsd.IsCanonicalUnitCommand(command) {
		scopeMonitorDone = make(chan struct{})
		var scopeCtx context.Context
		scopeCtx, stopScopeMonitor = context.WithCancel(runCtx)
		results := monitorWriteScope(scopeCtx, executionWorkDir, issueContext.WriteScope, defaultWriteScopePollInterval)
		go func() {
			defer close(scopeMonitorDone)
			if monitorErr := <-results; monitorErr != nil {
				scopeMu.Lock()
				scopeViolation = monitorErr
				scopeMu.Unlock()
				appendActivity("policy", "write_scope_violation", trustedUnit, "", "", time.Now().UTC())
				cancel()
			}
		}()
	}
	result := executionRunner.Run(runCtx, command, args, observer)
	postReconcile, reconcileErr := gsd.ReconcileOrphanedSubagents(config.GSDHome, executionWorkDir, time.Now().UTC())
	if reconcileErr != nil {
		result.Terminal = gsd.TerminalError
		result.Err = joinTerminalFailure(fmt.Errorf("reconcile terminated subagents: %w", reconcileErr), result.Err)
	} else if postReconcile.InterruptedRuns > 0 {
		appendActivity("transition", "interrupted_orphaned_subagents", trustedUnit, "", "", time.Now().UTC())
		if result.Terminal == gsd.TerminalSuccess {
			result.Terminal = gsd.TerminalError
			result.Err = errors.New("GSD process exited while nested subagents were still running")
		}
	}
	if restoreErr := gsd.NormalizeRuntimeSettings(config.GSDHome, config.CoordinatorModel, config.ImplementationModel, "high"); restoreErr != nil {
		result.Terminal = gsd.TerminalError
		result.Err = joinTerminalFailure(fmt.Errorf("restore coordinator runtime identity: %w", restoreErr), result.Err)
	}
	if stopScopeMonitor != nil {
		stopScopeMonitor()
	}
	if scopeMonitorDone != nil {
		<-scopeMonitorDone
	}
	scopeMu.Lock()
	monitorErr := scopeViolation
	scopeMu.Unlock()
	if monitorErr != nil {
		result.Terminal = gsd.TerminalError
		result.Err = joinTerminalFailure(monitorErr, result.Err)
	}
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
		model, thinking, identityErr := gsd.ReadSessionIdentity(sessionsDir, executionWorkDir)
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
		snapshot, queryErr := executionRunner.Query(ctx)
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
	if result.Terminal == gsd.TerminalSuccess && (observedModel != expectedModel || observedThinking != "high") {
		result.Terminal = gsd.TerminalError
		result.Err = fmt.Errorf("effective runtime identity was not observed as %s/high", expectedModel)
	}
	if err := authority.CheckLease(ctx, lease, time.Now().UTC()); err != nil {
		result.Terminal = gsd.TerminalError
		result.Err = err
	}
	candidateWorkDir := config.WorkDir
	candidateHead := ""
	if result.Terminal == gsd.TerminalSuccess && gsd.IsCanonicalUnitCommand(command) {
		message := "chore(gsd): checkpoint " + strings.ReplaceAll(before.Next.UnitID, "/", " ")
		if attemptWorktree != nil {
			if attemptManager == nil {
				result.Terminal = gsd.TerminalError
				result.Err = errors.New("attempt manager missing for candidate unit")
			} else if head, checkpointErr := attemptManager.CheckpointCandidate(ctx, *attemptWorktree, issueContext.WriteScope, message); checkpointErr != nil {
				result.Terminal = gsd.TerminalError
				result.Err = fmt.Errorf("checkpoint candidate canonical unit: %w", checkpointErr)
			} else {
				candidateHead = head
				candidateWorkDir = attemptWorktree.Root
				appendActivity("transition", "checkpointed_attempt_candidate", trustedUnit, "", "", time.Now().UTC())
			}
		} else if head, checkpointErr := shepherdgit.CheckpointWithinScopes(ctx, config.WorkDir, issueContext.WriteScope, message); checkpointErr != nil {
			result.Terminal = gsd.TerminalError
			result.Err = fmt.Errorf("checkpoint successful canonical unit: %w", checkpointErr)
		} else {
			candidateHead = head
			appendActivity("transition", "checkpointed_scoped_unit", trustedUnit, "", "", time.Now().UTC())
		}
	}
	endSnapshot, snapshotErr := shepherdgit.Inspect(ctx, config.WorkDir)
	if snapshotErr != nil {
		result.Terminal = gsd.TerminalError
		result.Err = joinTerminalFailure(result.Err, errors.New("inspect worktree after governed unit"))
	} else if cleanErr := shepherdgit.RequireClean(endSnapshot); cleanErr != nil {
		result.Terminal = gsd.TerminalError
		result.Err = joinTerminalFailure(result.Err, errors.New("worktree must be clean after a governed unit"))
	} else if result.Terminal == gsd.TerminalSuccess && gsd.IsCanonicalUnitCommand(command) {
		if candidateHead == "" {
			result.Terminal = gsd.TerminalError
			result.Err = errors.New("candidate head is required before independent validation")
		} else if endSnapshot.HeadSHA != startSnapshot.HeadSHA {
			result.Terminal = gsd.TerminalError
			result.Err = errors.New("canonical head changed before ratified promotion")
		} else {
			validator := independentValidatorFactory(runner, config)
			stateVersion, versionErr := authority.GovernanceStateVersion(ctx)
			if versionErr != nil {
				result.Terminal = gsd.TerminalError
				result.Err = versionErr
			} else if err := persistSuccessProof(ctx, authority, validator, config, registry, issueContext.WriteScope, candidateWorkDir, deliveryID, trustedUnit, before.Next.UnitType, issueContext.PRBase, attempt.Generation, unitAttempt.Attempts, stateVersion, startSnapshot.HeadSHA, candidateHead); err != nil {
				result.Terminal = gsd.TerminalError
				result.Err = err
			} else if attemptWorktree != nil {
				if err := attemptManager.PromoteCandidate(ctx, *attemptWorktree, candidateHead); err != nil {
					result.Terminal = gsd.TerminalError
					result.Err = fmt.Errorf("promote ratified candidate: %w", err)
				} else if adoptErr := attemptManager.AdoptGSDState(ctx, *attemptWorktree); adoptErr != nil {
					result.Terminal = gsd.TerminalError
					result.Err = fmt.Errorf("adopt ratified GSD state: %w", adoptErr)
				} else {
					appendActivity("transition", "promoted_ratified_attempt", trustedUnit, "", "", time.Now().UTC())
					_ = attemptManager.Discard(context.Background(), *attemptWorktree)
				}
			}
		}
	}
	endSnapshot, snapshotErr = shepherdgit.Inspect(ctx, config.WorkDir)
	if snapshotErr != nil {
		result.Terminal = gsd.TerminalError
		result.Err = joinTerminalFailure(result.Err, errors.New("inspect worktree after validation and promotion"))
	} else if cleanErr := shepherdgit.RequireClean(endSnapshot); cleanErr != nil {
		result.Terminal = gsd.TerminalError
		result.Err = joinTerminalFailure(result.Err, errors.New("worktree must be clean after validation and promotion"))
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
	unitOutcome := string(result.Terminal)
	if result.Err != nil {
		unitOutcome = classifyUnitFailure(result)
	}
	if err := authority.FinishUnitAttempt(ctx, unitAttemptKey, unitOutcome); err != nil {
		result.Terminal = gsd.TerminalError
		result.Err = joinTerminalFailure(result.Err, err)
	} else {
		unitAttemptFinished = true
	}
	remainingAttempts := unitAttempt.Remaining
	if result.Terminal != gsd.TerminalSuccess && result.Err != nil && isAutomaticallyRetryable(result.Err) {
		class := classifyUnitFailure(result)
		budget, budgetErr := authority.BeginRecoveryAttempt(ctx, store.RecoveryBudgetKey{DeliveryID: deliveryID, Generation: attempt.Generation, UnitID: trustedUnit, HeadSHA: startSnapshot.HeadSHA, FailureClass: class}, int64(config.MaxUnitAttempts), time.Second, result.Err.Error(), "Sol/high recovery planning required before next retry", time.Now().UTC())
		if errors.Is(budgetErr, store.ErrRetryBudgetExhausted) {
			remainingAttempts = 0
		} else if budgetErr != nil {
			result.Err = joinTerminalFailure(result.Err, budgetErr)
			remainingAttempts = 0
		} else {
			remainingAttempts = budget.MaxAttempts - budget.Attempts
		}
	}
	targetState := finalUnitRunState(&result, postPhase, remainingAttempts)
	if targetState == domain.RunAwaitingDecision {
		class := classifyUnitFailure(result)
		if err := persistAndPublishDecisionRequest(ctx, authority, lease, decisionPublisher, decisionTarget, deliveryID, issue, trustedUnit, attempt.Generation, startSnapshot.HeadSHA, class, result.Err); err != nil {
			result.Terminal = gsd.TerminalError
			result.Err = joinTerminalFailure(result.Err, err)
			targetState = domain.RunFailed
		}
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

func monitorWriteScope(ctx context.Context, root string, scopes []string, interval time.Duration) <-chan error {
	results := make(chan error, 1)
	go func() {
		defer close(results)
		if interval <= 0 {
			results <- errors.New("write scope monitor interval must be positive")
			return
		}
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				results <- nil
				return
			case <-ticker.C:
				paths, err := shepherdgit.ChangedPathsOutsideScopes(ctx, root, scopes)
				if err != nil {
					if ctx.Err() != nil {
						results <- nil
						return
					}
					results <- fmt.Errorf("inspect live write scope: %w", err)
					return
				}
				if len(paths) > 0 {
					results <- fmt.Errorf("live write-scope breach: changed path %q is outside the issue write scope", paths[0])
					return
				}
			}
		}
	}()
	return results
}

func modelForUnitType(config fileConfig, registry gsd.UnitRegistry, unitType string) (string, error) {
	role, err := registry.ModelRoleForUnit(unitType)
	if err != nil {
		return "", err
	}
	if role == gsd.ModelRoleImplementation {
		return config.ImplementationModel, nil
	}
	return config.CoordinatorModel, nil
}

func expectedModelForObservedUnit(config fileConfig, registry gsd.UnitRegistry, command, canonicalUnitType string) (string, error) {
	if canonicalUnitType != "" {
		return modelForUnitType(config, registry, canonicalUnitType)
	}
	if command == "new-milestone" || command == "query" || command == "status" || command == "next" || command == "auto" {
		return config.CoordinatorModel, nil
	}
	if command == "discuss" {
		return modelForUnitType(config, registry, "discuss-milestone")
	}
	return modelForUnitType(config, registry, command)
}

func launchModelForCommand(config fileConfig, registry gsd.UnitRegistry, command string) (string, error) {
	return expectedModelForObservedUnit(config, registry, command, "")
}

const (
	unitFailureRuntimeContractMismatch = "runtime_contract_mismatch"
	unitFailureArtifactMissing         = "artifact_missing"
	unitFailureFalseGreen              = "false_green"
	unitFailureInterrupted             = "interrupted"
	unitFailureOrphanedSubagent        = "orphaned_subagent"
	unitFailureStaleHead               = "stale_head"
	unitFailureScopeBreach             = "scope_breach"
	unitFailureModelDrift              = "model_drift"
	unitFailureRetryExhausted          = "retry_exhausted"
	unitFailureRuntimeFailure          = "runtime_failure"
	unitFailureUnsafe                  = "unsafe_failure"
)

func classifyUnitFailure(result gsd.Result) string {
	if result.Err == nil {
		return string(result.Terminal)
	}
	if errors.Is(result.Err, store.ErrRetryBudgetExhausted) {
		return unitFailureRetryExhausted
	}
	if errors.Is(result.Err, gsd.ErrRuntimeContractMismatch) {
		return unitFailureRuntimeContractMismatch
	}
	message := strings.ToLower(result.Err.Error())
	switch {
	case strings.Contains(message, "runtime_contract_mismatch") || strings.Contains(message, "runtime contract"):
		return unitFailureRuntimeContractMismatch
	case strings.Contains(message, "artifact") && (strings.Contains(message, "missing") || strings.Contains(message, "not found") || strings.Contains(message, "regular file")):
		return unitFailureArtifactMissing
	case strings.Contains(message, "canonical") && (strings.Contains(message, "did not advance") || strings.Contains(message, "unchanged")):
		return unitFailureFalseGreen
	case result.Terminal == gsd.TerminalCancelled || result.Terminal == gsd.TerminalTimeout || strings.Contains(message, "interrupted") || strings.Contains(message, "cancelled") || strings.Contains(message, "deadline exceeded") || strings.Contains(message, "timeout"):
		return unitFailureInterrupted
	case strings.Contains(message, "subagent") && (strings.Contains(message, "orphan") || strings.Contains(message, "still running") || strings.Contains(message, "unreconciled")):
		return unitFailureOrphanedSubagent
	case strings.Contains(message, "stale") && strings.Contains(message, "head"):
		return unitFailureStaleHead
	case strings.Contains(message, "head continuity") || strings.Contains(message, "head changed"):
		return unitFailureStaleHead
	case strings.Contains(message, "scope breach") || strings.Contains(message, "write-scope") || strings.Contains(message, "outside the issue write scope"):
		return unitFailureScopeBreach
	case strings.Contains(message, "model") || strings.Contains(message, "thinking") || strings.Contains(message, "runtime identity"):
		return unitFailureModelDrift
	case strings.Contains(message, "runtime:") || strings.Contains(message, "headless") || strings.Contains(message, "process"):
		return unitFailureRuntimeFailure
	case result.Terminal == gsd.TerminalRejected || result.Terminal == gsd.TerminalBlocked:
		return unitFailureUnsafe
	default:
		return unitFailureUnsafe
	}
}

func isAutomaticallyRetryable(err error) bool {
	if err == nil {
		return false
	}
	class := classifyUnitFailure(gsd.Result{Terminal: gsd.TerminalError, Err: err})
	switch class {
	case unitFailureArtifactMissing, unitFailureFalseGreen, unitFailureInterrupted, unitFailureRuntimeFailure:
		return true
	default:
		return false
	}
}

func finalUnitRunState(result *gsd.Result, postPhase string, remainingAttempts int64) domain.RunState {
	if result.Terminal != gsd.TerminalSuccess && result.Terminal != gsd.TerminalBlocked && isAutomaticallyRetryable(result.Err) {
		if remainingAttempts > 0 {
			return domain.RunReady
		}
		result.Err = joinTerminalFailure(result.Err, store.ErrRetryBudgetExhausted)
		return domain.RunAwaitingDecision
	}
	return targetRunState(result.Terminal, result.Err, postPhase)
}

func targetRunState(terminal gsd.Terminal, terminalErr error, postPhase string) domain.RunState {
	if terminal == gsd.TerminalBlocked {
		if errors.Is(terminalErr, gsd.ErrMutatingSkip) {
			return domain.RunReady
		}
		return domain.RunBlocked
	}
	if terminal == gsd.TerminalSuccess {
		if postPhase == "complete" {
			return domain.RunHumanGate
		}
		return domain.RunReady
	}
	return domain.RunFailed
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

func joinTerminalFailure(primary, secondary error) error {
	if primary == nil {
		return secondary
	}
	return errors.Join(primary, secondary)
}

type commandExitError struct {
	code int
	err  error
}

func (e commandExitError) Error() string { return e.err.Error() }
func (e commandExitError) Unwrap() error { return e.err }
func (e commandExitError) ExitCode() int { return e.code }

func deliveryID(issue int) string { return "issue-" + strconv.Itoa(issue) }

func ensureIssueDelivery(ctx context.Context, authority *store.Store, config fileConfig, issue int, contextHash string) (store.Delivery, contract.IssueContext, error) {
	issueContext, err := loadProtectedIssueContext(config.StateDir, issue, contextHash)
	if err != nil {
		return store.Delivery{}, contract.IssueContext{}, fmt.Errorf("protected issue context: %w", err)
	}
	snapshot, err := shepherdgit.Inspect(ctx, config.WorkDir)
	if err != nil {
		return store.Delivery{}, contract.IssueContext{}, err
	}
	if snapshot.Branch != issueContext.Branch {
		return store.Delivery{}, contract.IssueContext{}, fmt.Errorf("issue context branch %s does not match attached worktree branch %s", issueContext.Branch, snapshot.Branch)
	}
	initialHead := snapshot.HeadSHA
	if existing, getErr := authority.GetDelivery(ctx, deliveryID(issue)); getErr == nil && existing.InitialHead != "" {
		initialHead = existing.InitialHead
	}
	delivery := store.Delivery{
		ID: deliveryID(issue), Issue: issue, ParentIssue: issueContext.ParentIssue,
		WorkDir: config.WorkDir, ContextHash: contextHash, Branch: issueContext.Branch,
		BaseBranch: issueContext.PRBase, GSDProjectRoot: config.WorkDir,
		InitialHead: initialHead, GSDVersion: config.GSDVersion,
	}
	if err := authority.EnsureDelivery(ctx, delivery); err != nil {
		return store.Delivery{}, contract.IssueContext{}, err
	}
	identity := gsd.IssueProjectIdentity{
		DeliveryID: delivery.ID, Issue: delivery.Issue, ParentIssue: delivery.ParentIssue,
		Branch: delivery.Branch, BaseBranch: delivery.BaseBranch, ProjectRoot: delivery.GSDProjectRoot,
		InitialHead: delivery.InitialHead, ContextHash: delivery.ContextHash, GSDVersion: delivery.GSDVersion,
	}
	if err := gsd.BootstrapIssueProject(config.WorkDir, identity, issueContext); err != nil {
		return store.Delivery{}, contract.IssueContext{}, fmt.Errorf("bootstrap issue GSD project: %w", err)
	}
	return delivery, issueContext, nil
}

func validateDeliveryInvocation(delivery store.Delivery, config fileConfig, issueContext contract.IssueContext) error {
	if delivery.Issue != issueContext.Issue || delivery.ParentIssue != issueContext.ParentIssue ||
		delivery.WorkDir != config.WorkDir || delivery.GSDProjectRoot != config.WorkDir ||
		delivery.Branch != issueContext.Branch || delivery.BaseBranch != issueContext.PRBase ||
		delivery.GSDVersion != config.GSDVersion {
		return errors.New("delivery invocation does not match the canonical issue GSD identity")
	}
	return nil
}

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

type decisionPublisher interface {
	SyncDecisionComment(context.Context, shepherdgithub.Target, string) error
}

func appendAndPublishDecision(ctx context.Context, store *decisionlog.Store, publisher decisionPublisher, target shepherdgithub.Target, deliveryID, executionID, unitID string, question gsd.Question, response gsd.UIResponse, actorValue, basis string) error {
	if err := appendDecision(store, deliveryID, executionID, unitID, question, response, actorValue, basis); err != nil {
		return err
	}
	records, err := store.Records()
	if err != nil {
		return fmt.Errorf("read durable decisions: %w", err)
	}
	filtered := records[:0]
	for _, record := range records {
		if record.DeliveryID == deliveryID {
			filtered = append(filtered, record)
		}
	}
	if err := publisher.SyncDecisionComment(ctx, target, decisionlog.Markdown(filtered)); err != nil {
		return fmt.Errorf("publish durable decision: %w", err)
	}
	return nil
}

func recordOperationalDecision(ctx context.Context, store *decisionlog.Store, publisher decisionPublisher, target shepherdgithub.Target, deliveryID, unitID, question, answer, actor, basis string) error {
	if strings.TrimSpace(unitID) == "" || strings.TrimSpace(question) == "" || strings.TrimSpace(answer) == "" {
		return errors.New("decision unit, question, and answer are required")
	}
	executionID := fmt.Sprintf("decision-%d", time.Now().UTC().UnixNano())
	return appendAndPublishDecision(ctx, store, publisher, target, deliveryID, executionID, unitID,
		gsd.Question{ID: "operational-decision", Title: question}, gsd.UIResponse{Value: answer}, actor, basis)
}

func publishDecisions(ctx context.Context, config fileConfig, deliveryID, summary string) error {
	client := shepherdgithub.NewCLIClient()
	return client.SyncDecisionComment(ctx, shepherdgithub.Target{
		Repository: config.Repository, PullRequest: config.PullRequest, DeliveryID: deliveryID,
	}, summary)
}

func gsdStateArtifacts(workDir string) ([]shepherdgit.Artifact, error) {
	root := filepath.Join(workDir, ".gsd")
	if info, err := os.Lstat(root); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	} else if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
		return nil, errors.New("GSD state root must be a real directory")
	}
	var artifacts []shepherdgit.Artifact
	for _, relRoot := range []string{"STATE.md", "state-manifest.json", "phases"} {
		path := filepath.Join(root, relRoot)
		if _, err := os.Lstat(path); os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, err
		}
		if err := filepath.WalkDir(path, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			info, err := entry.Info()
			if err != nil {
				return err
			}
			if info.Mode()&os.ModeSymlink != 0 {
				return errors.New("GSD state artifact must not be a symlink")
			}
			if entry.IsDir() || !info.Mode().IsRegular() {
				return nil
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			rel, err := filepath.Rel(workDir, path)
			if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
				return errors.New("GSD state artifact escapes worktree")
			}
			hash := sha256.Sum256(raw)
			artifacts = append(artifacts, shepherdgit.Artifact{Path: filepath.ToSlash(rel), Hash: "sha256:" + hex.EncodeToString(hash[:])})
			return nil
		}); err != nil {
			return nil, err
		}
	}
	sort.Slice(artifacts, func(i, j int) bool { return artifacts[i].Path < artifacts[j].Path })
	return artifacts, nil
}

func persistSuccessProof(ctx context.Context, authorityStore *store.Store, validator validation.Validator, config fileConfig, registry gsd.UnitRegistry, scopes []string, workDir, deliveryID, unitID, unitType, baseBranch string, generation, attempt, stateVersion int64, startHead, candidateHead string) error {
	metadata, ok := registry.Lookup(unitType)
	if !ok {
		return fmt.Errorf("missing official unit metadata for %s", unitType)
	}
	artifacts, err := shepherdgit.ArtifactManifest(ctx, workDir, startHead, candidateHead, scopes)
	if err != nil {
		return err
	}
	stateArtifacts, err := gsdStateArtifacts(workDir)
	if err != nil {
		return err
	}
	artifacts = append(artifacts, stateArtifacts...)
	if len(artifacts) == 0 {
		return errors.New("successful canonical unit produced no artifact manifest")
	}
	sort.Slice(artifacts, func(i, j int) bool { return artifacts[i].Path < artifacts[j].Path })
	contract := struct {
		UnitType              string   `json:"unit_type"`
		PhaseChain            []string `json:"phase_chain"`
		RequiredWorkflowTools []string `json:"required_workflow_tools"`
	}{UnitType: unitType, PhaseChain: metadata.PhaseChain, RequiredWorkflowTools: metadata.RequiredWorkflowTools}
	contractRaw, err := json.Marshal(contract)
	if err != nil {
		return err
	}
	contractHash := sha256.Sum256(contractRaw)
	manifest := struct {
		UnitType              string                 `json:"unit_type"`
		PhaseChain            []string               `json:"phase_chain"`
		RequiredWorkflowTools []string               `json:"required_workflow_tools"`
		Artifacts             []shepherdgit.Artifact `json:"artifacts"`
	}{UnitType: unitType, PhaseChain: metadata.PhaseChain, RequiredWorkflowTools: metadata.RequiredWorkflowTools, Artifacts: artifacts}
	raw, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	evidenceHash := sha256.Sum256(raw)
	artifactHashes := make([]validation.ArtifactHash, 0, len(artifacts))
	for _, artifact := range artifacts {
		artifactHashes = append(artifactHashes, validation.ArtifactHash{Path: artifact.Path, Hash: artifact.Hash})
	}
	evidenceHashValue := "sha256:" + hex.EncodeToString(evidenceHash[:])
	contractHashValue := "sha256:" + hex.EncodeToString(contractHash[:])
	requiredGates := requiredGatesForUnit(metadata)
	validationResult, err := validator.Validate(ctx, validation.Request{
		Repository: config.Repository, PullRequest: config.PullRequest, BaseBranch: baseBranch,
		DeliveryID: deliveryID, Generation: generation, UnitID: unitID, UnitType: unitType, Attempt: attempt,
		StateVersion: stateVersion, WorkDir: workDir, GSDHome: config.GSDHome, StateDir: config.StateDir,
		BaseHead: startHead, CandidateHead: candidateHead,
		ContractHash: contractHashValue,
		EvidenceHash: evidenceHashValue, ArtifactHashes: artifactHashes,
		RequireGates: requiredGates,
	})
	if err != nil {
		return err
	}
	if validationResult.SessionID == "" {
		return errors.New("validator session identity is required")
	}
	if validationResult.EvidenceHash != evidenceHashValue {
		return errors.New("validator evidence hash does not match candidate artifacts")
	}
	issuedAt := validationResult.IssuedAt
	if issuedAt.IsZero() {
		return errors.New("validator issue time is required")
	}
	attestation, err := authority.Ratify(authority.RatificationRequest{
		Repository: config.Repository, PR: config.PullRequest, BaseBranch: baseBranch, BaseSHA: startHead,
		CandidateHead: candidateHead, ObservedHead: validationResult.ObservedHead,
		RunID: deliveryID, Generation: generation, UnitID: unitID, Attempt: attempt, StateVersion: stateVersion,
		ContractHash: contractHashValue, EvidenceHash: validationResult.EvidenceHash,
		Validator: validationResult.ObservedModel, Thinking: validationResult.Thinking, ValidatorSessionID: validationResult.SessionID, Verdict: validationResult.Verdict,
		LocalGates: validationResult.LocalGates, UAT: validationResult.UAT, MilestoneValid: validationResult.MilestoneValid,
		RequiredLocalGates: requiredGates.LocalGates, RequiredUAT: requiredGates.UAT, RequiredMilestoneValid: requiredGates.MilestoneValid,
		IssuedAt: validationResult.IssuedAt, ExpiresAt: validationResult.ExpiresAt,
	}, time.Now().UTC())
	if err != nil {
		return err
	}
	proofIDHash := sha256.Sum256([]byte(deliveryID + ":" + unitID + ":" + candidateHead + ":" + hex.EncodeToString(evidenceHash[:])))
	if err := authorityStore.PutArtifactProof(ctx, store.ArtifactProof{
		ProofID: hex.EncodeToString(proofIDHash[:8]), DeliveryID: deliveryID, Generation: generation, UnitID: unitID, Attempt: attempt,
		StartHead: startHead, CandidateHead: candidateHead, ValidatedHead: attestation.HeadSHA,
		ExpectedArtifact: string(raw), ArtifactHash: "sha256:" + hex.EncodeToString(evidenceHash[:]),
		Validator: attestation.Validator, Thinking: attestation.Thinking,
		Ratified: attestation.HeadSHA == candidateHead && attestation.Validator == validationResult.ObservedModel,
	}); err != nil {
		return err
	}
	return authorityStore.PutAttestation(ctx, store.AttestationRecord{
		Repository: attestation.Repository, PR: attestation.PR, BaseBranch: attestation.BaseBranch,
		BaseHead: attestation.BaseSHA, CandidateHead: attestation.HeadSHA, ObservedHead: validationResult.ObservedHead,
		RunID: deliveryID, Generation: generation, UnitID: unitID, Attempt: attempt, StateVersion: stateVersion,
		ContractHash: attestation.ContractHash, EvidenceHash: attestation.EvidenceHash,
		ValidatorSessionID: attestation.ValidatorSessionID, HeadSHA: attestation.HeadSHA,
		Validator: attestation.Validator, Thinking: attestation.Thinking, Verdict: attestation.Verdict,
		LocalGates: attestation.LocalGates, UAT: attestation.UAT, MilestoneValid: attestation.MilestoneValid,
		CreatedAt: attestation.IssuedAt, ExpiresAt: attestation.ExpiresAt,
	})
}

func requiredGatesForUnit(metadata gsd.UnitMetadata) validation.GateRequirements {
	gates := validation.GateRequirements{LocalGates: true}
	for _, phase := range metadata.PhaseChain {
		switch phase {
		case "uat":
			gates.UAT = true
		case "validation", "completion":
			gates.MilestoneValid = true
		}
	}
	if metadata.ScopeClass == "section-close" {
		gates.MilestoneValid = true
	}
	return gates
}

func consumeGitHubDecisionReplies(ctx context.Context, authority *store.Store, client *shepherdgithub.Client, config fileConfig, deliveryID string, issue int) error {
	requests, err := authority.ListOpenDecisionRequests(ctx, deliveryID)
	if err != nil {
		return err
	}
	for _, request := range requests {
		replies, err := client.PollDecisionReplies(ctx, shepherdgithub.QuestionRequest{
			RequestID: request.RequestID, Repository: config.Repository, Issue: issue, PullRequest: config.PullRequest,
			DeliveryID: deliveryID, UnitID: request.UnitID, Generation: request.Generation, HeadSHA: request.HeadSHA,
			Evidence: request.Evidence, Options: request.Options, RecommendedOption: request.RecommendedOption,
			SafeDefault: request.SafeDefault, ExpiresAt: request.ExpiresAt, Mention: "karthik-sivadas",
		}, "karthik-sivadas")
		if err != nil {
			return err
		}
		if len(replies) == 0 {
			continue
		}
		reply := replies[0]
		if err := authority.AcceptDecisionRequestAnswer(ctx, request.RequestID, reply.Option, reply.Author, request.Generation, request.HeadSHA, time.Now().UTC()); err != nil {
			return err
		}
		if _, err := authority.ConsumeDecisionRequest(ctx, request.RequestID); err != nil {
			return err
		}
		if reply.Option == "retry" || reply.Option == "continue" {
			return authority.ResumeDelivery(ctx, domain.HumanDecision{RunID: deliveryID, Generation: request.Generation, ActorKind: domain.ActorHuman, Approved: true})
		}
		return authority.BlockAwaitingDecision(ctx, deliveryID, request.Generation)
	}
	return nil
}

func persistAndPublishDecisionRequest(ctx context.Context, authority *store.Store, lease store.Lease, publisher *shepherdgithub.Client, target shepherdgithub.Target, deliveryID string, issue int, unitID string, generation int64, headSHA, failureClass string, cause error) error {
	if publisher == nil {
		return errors.New("decision question publisher is required")
	}
	basis := failureClass
	if cause != nil {
		basis = failureClass + ": redacted local diagnostic available in activity log"
	}
	hash := sha256.Sum256([]byte(deliveryID + ":" + strconv.FormatInt(generation, 10) + ":" + unitID + ":" + headSHA + ":" + failureClass))
	requestID := "decision-" + hex.EncodeToString(hash[:8])
	expiresAt := time.Now().UTC().Add(24 * time.Hour)
	request := store.DecisionRequest{
		RequestID: requestID, DeliveryID: deliveryID, Issue: issue, PullRequest: target.PullRequest,
		UnitID: unitID, Generation: generation, HeadSHA: headSHA, Kind: failureClass,
		Evidence: boundedEvidence(basis), Options: []string{"retry", "stop"}, RecommendedOption: "retry",
		SafeDefault: "stop", ExpiresAt: expiresAt, Status: store.DecisionRequestOpen,
	}
	stored, err := authority.UpsertDecisionRequest(ctx, request)
	if err != nil {
		return err
	}
	grant, err := domain.NewGrant(deliveryID, target.Repository, issue, domain.CapabilityPRUpdate, lease.Epoch)
	if err != nil {
		return err
	}
	if err := authority.PutGrant(ctx, grant); err != nil {
		return err
	}
	payloadHash := sha256.Sum256([]byte(stored.RequestID + ":" + stored.Evidence))
	effectKey := "github-question:" + stored.RequestID
	if _, err := authority.Enqueue(ctx, lease, store.Effect{Key: effectKey, RunID: deliveryID, Repository: target.Repository, Issue: issue, Capability: domain.CapabilityPRUpdate, Target: "pr:" + strconv.Itoa(target.PullRequest), PayloadHash: "sha256:" + hex.EncodeToString(payloadHash[:]), Epoch: lease.Epoch}, time.Now().UTC()); err != nil {
		return err
	}
	if _, err := authority.ClaimEffect(ctx, lease, effectKey, time.Now().UTC()); err != nil {
		return err
	}
	commentID, err := publisher.SyncQuestionComment(ctx, shepherdgithub.QuestionRequest{
		RequestID: stored.RequestID, Repository: target.Repository, Issue: issue, PullRequest: target.PullRequest,
		DeliveryID: deliveryID, UnitID: unitID, Generation: generation, HeadSHA: headSHA,
		Evidence: stored.Evidence, Options: stored.Options, RecommendedOption: stored.RecommendedOption,
		SafeDefault: stored.SafeDefault, ExpiresAt: stored.ExpiresAt, Mention: "karthik-sivadas",
	})
	if err != nil {
		_ = authority.MarkEffectFailed(ctx, lease, effectKey, err, time.Now().UTC())
		return err
	}
	if err := authority.MarkEffectSent(ctx, lease, effectKey, time.Now().UTC()); err != nil {
		return err
	}
	return authority.MarkDecisionRequestPublished(ctx, stored.RequestID, commentID)
}

func boundedEvidence(value string) string {
	lower := strings.ToLower(value)
	if strings.Contains(lower, "token") || strings.Contains(lower, "secret") || strings.Contains(lower, "credential") || strings.Contains(lower, "password") || strings.Contains(lower, "authorization") {
		return "redacted sensitive diagnostic; see local typed failure class"
	}
	value = strings.TrimSpace(strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, value))
	if len(value) > 1000 {
		return value[:1000]
	}
	if value == "" {
		return "retry budget requires human decision"
	}
	return value
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

func terminalQuestion(ctx context.Context, question gsd.Question, actor string) (gsd.UIResponse, error) {
	label := "HUMAN GATE"
	if strings.TrimSpace(actor) != string(decisionlog.ActorHuman) {
		label = "SHEPHERD GATE actor=" + strings.TrimSpace(actor)
	}
	fmt.Fprintf(os.Stderr, "\n%s [%s] %s\n", label, question.Method, question.Title)
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
	canonicalRoot, err := canonicalPathAllowMissing(root)
	if err != nil {
		return false, err
	}
	canonicalPath, err := canonicalPathAllowMissing(path)
	if err != nil {
		return false, err
	}
	relative, err := filepath.Rel(canonicalRoot, canonicalPath)
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

func materializeProtectedIssueContext(stateDir string, issue int, raw []byte) error {
	if !filepath.IsAbs(stateDir) || issue <= 0 || len(raw) == 0 || len(raw) > contract.MaxIssueContextBytes {
		return errors.New("valid protected context target and bounded contents are required")
	}
	directory := filepath.Join(stateDir, "context")
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return err
	}
	path := filepath.Join(directory, fmt.Sprintf("issue-%d.json", issue))
	if err := compareProtectedContext(path, raw); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	temporary, err := os.CreateTemp(directory, ".issue-context-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer os.Remove(temporaryPath)
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(raw); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Link(temporaryPath, path); err != nil {
		if compareErr := compareProtectedContext(path, raw); compareErr == nil {
			return nil
		}
		return errors.New("protected issue context could not be atomically bound")
	}
	directoryHandle, err := os.Open(directory)
	if err != nil {
		return err
	}
	syncErr := directoryHandle.Sync()
	closeErr := directoryHandle.Close()
	return errors.Join(syncErr, closeErr)
}

func loadProtectedIssueContext(stateDir string, issue int, expectedHash string) (contract.IssueContext, error) {
	path := filepath.Join(stateDir, "context", fmt.Sprintf("issue-%d.json", issue))
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() {
		return contract.IssueContext{}, errors.New("protected issue context must be a regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return contract.IssueContext{}, err
	}
	context, raw, decodeErr := contract.DecodeIssueContext(file, issue)
	closeErr := file.Close()
	if decodeErr != nil {
		return contract.IssueContext{}, decodeErr
	}
	if closeErr != nil {
		return contract.IssueContext{}, closeErr
	}
	hash := sha256.Sum256(raw)
	if expectedHash != "sha256:"+hex.EncodeToString(hash[:]) {
		return contract.IssueContext{}, errors.New("protected issue context hash does not match authority")
	}
	return context, nil
}

func compareProtectedContext(path string, expected []byte) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || info.Size() > contract.MaxIssueContextBytes {
		return errors.New("protected issue context is not a bounded regular file")
	}
	existing, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if !strings.EqualFold(hex.EncodeToString(sha256Sum(existing)), hex.EncodeToString(sha256Sum(expected))) {
		return errors.New("protected issue context is already bound to different contents")
	}
	return nil
}

func sha256Sum(raw []byte) []byte {
	hash := sha256.Sum256(raw)
	return hash[:]
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
