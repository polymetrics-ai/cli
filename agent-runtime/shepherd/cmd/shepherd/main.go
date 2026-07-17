package main

import (
	"bufio"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
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
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/outbox"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/recovery"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/store"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/supervisor"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/telemetry"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/validation"
	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/workspace"
)

const version = "0.1.0"

const defaultMaxEventBytes = 8 * 1024 * 1024

const defaultWriteScopePollInterval = 500 * time.Millisecond

var errPreDispatchAttemptQuery = errors.New("pre-dispatch attempt query failed")

var exitAwaitingDecision = integrationExitAwaitingDecision

var independentValidatorFactory = func(_ *gsd.Runner, config fileConfig) validation.Validator {
	return validation.GSDValidator{
		Command: config.PiCommand, GSDHome: config.GSDHome, StateDir: config.StateDir,
		Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
	}
}

var recoveryPlannerFactory = func(config fileConfig) (recovery.Planner, error) {
	return recovery.NewPiPlanner(recovery.PiPlannerConfig{
		Command: config.PiCommand, GSDHome: config.GSDHome, StateDir: config.StateDir,
		Timeout: time.Duration(config.TimeoutSeconds) * time.Second,
	})
}

var decisionReplyPollerFactory = func() decisionReplyPoller {
	return shepherdgithub.NewCLIReplyPoller()
}

var prepareAttemptGSDState = func(ctx context.Context, manager *workspace.Manager, attempt workspace.AttemptWorktree) error {
	return manager.PrepareGSDState(ctx, attempt)
}

var reconcileOwnedAttempt = func(ctx context.Context, manager *workspace.Manager, owned workspace.OwnedAttempt) (workspace.ReconcileResult, error) {
	return manager.ReconcileOwnedAttempt(ctx, owned)
}

type fileConfig struct {
	GSDCommand          []string `json:"gsd_command"`
	PiCommand           []string `json:"pi_command"`
	WorkDir             string   `json:"work_dir"`
	GSDHome             string   `json:"gsd_home"`
	StateDir            string   `json:"state_dir"`
	AttemptRoot         string   `json:"attempt_root"`
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
	continueUnit := flags.Bool("continue-unit", false, "fail closed: disposable governed-unit continuation is not yet qualified")
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
	resolvedWorkDir, err := filepath.EvalSymlinks(config.WorkDir)
	if err != nil {
		return fmt.Errorf("resolve governed work_dir: %w", err)
	}
	config.WorkDir = resolvedWorkDir
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
			decisionStore, err := decisionlog.Open(filepath.Join(config.StateDir, "decisions"))
			if err != nil {
				return err
			}
			defer decisionStore.Close()
			return withStandaloneExternalEffectController(ctx, config, *issue, func(effects *externalEffectController) error {
				return recordOperationalDecision(ctx, decisionStore, effects, deliveryID(*issue), *decisionUnit,
					*decisionQuestion, *decisionAnswer, *decisionActor, *decisionBasis)
			})
		}
		records, err := decisionlog.Read(filepath.Join(config.StateDir, "decisions"))
		if err != nil {
			return err
		}
		snapshot, err := decisionlog.SnapshotRecords(records, deliveryID(*issue))
		if err != nil {
			return err
		}
		if *publish {
			if err := withStandaloneExternalEffectController(ctx, config, *issue, func(effects *externalEffectController) error {
				_, err := effects.RequestSummary(ctx, snapshot)
				return err
			}); err != nil {
				return err
			}
		}
		fmt.Print(snapshot.Summary)
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
		if err := withStandaloneExternalEffectController(ctx, config, *issue,
			func(controller *externalEffectController) error {
				if err := controller.RecoverClaims(ctx); err != nil {
					return err
				}
				if err := controller.RecoverUncertain(ctx); err != nil {
					return err
				}
				if err := resolveHumanClearedAttempts(ctx, controller.authority, config,
					input.DeliveryID, controller.workspaceManager, &controller.lease); err != nil {
					return err
				}
				activeClaims, err := controller.store.ListDelivery(ctx, input.DeliveryID, outbox.StateClaimed)
				if err != nil {
					return err
				}
				if len(activeClaims) > 0 {
					return errors.New("active external effect claim must expire or settle before resume")
				}
				return controller.authority.ResumeDeliveryFenced(ctx, controller.lease, domain.HumanDecision{
					RunID: input.DeliveryID, Generation: input.Generation,
					ActorKind: domain.ActorHuman, Approved: true,
				}, time.Now().UTC())
			}); err != nil {
			return fmt.Errorf("fenced human resume: %w", err)
		}
		return nil
	}
	var container *gsd.ContainerConfig
	var runtimeGuard *gsd.HostRuntimeGuard
	var registry gsd.UnitRegistry
	forceSessionBinding := false
	if config.Runtime == "podman" {
		container = &gsd.ContainerConfig{Engine: "podman", Image: config.ContainerImage,
			GSDStateDir: filepath.Join(config.StateDir, "runtime", "gsd"), PlanningDir: filepath.Join(config.StateDir, "runtime", "planning"),
			AuthFile: config.AuthFile, SettingsFile: filepath.Join(config.GSDHome, "agent", "settings.json"), Network: config.ContainerNetwork,
			PolicyDir: config.PolicyDir, GitCommonDir: config.GitCommonDir,
			SessionsDir: filepath.Join(config.StateDir, "runtime", "sessions"), BackgroundDir: filepath.Join(config.StateDir, "runtime", "bg-shell"),
			BackupDir: filepath.Join(config.StateDir, "runtime", "gsd-backups")}
		resolved, err := gsd.ResolvePinnedContainerImage(ctx, *container)
		if err != nil {
			return fmt.Errorf("runtime admission: %w", err)
		}
		container = &resolved
		registry, err = gsd.LoadPinnedContainerUnitRegistry(ctx, *container, config.GSDVersion)
		if err != nil {
			return fmt.Errorf("runtime admission: %w", err)
		}
	} else {
		integrationExecutor, integrationErr := integrationGSDExecutor()
		if integrationErr != nil {
			return fmt.Errorf("runtime admission: %w", integrationErr)
		}
		if integrationExecutor != "" {
			// The integration binary imports the exact official registry but
			// replaces only external execution. Session evidence remains mandatory.
			registry, err = gsd.LoadPinnedUnitRegistry(ctx, config.GSDCommand, config.GSDHome, config.GSDVersion)
			if err != nil {
				return fmt.Errorf("runtime admission: %w", err)
			}
			config.GSDCommand = []string{integrationExecutor}
			forceSessionBinding = true
		} else {
			config.GSDCommand, err = gsd.PreparePinnedHostRuntime(ctx, config.GSDCommand, config.GSDHome, config.GSDVersion, config.WorkDir, config.AttemptRoot)
			if err != nil {
				return fmt.Errorf("runtime admission: %w", err)
			}
			runtimeGuard, err = gsd.NewPinnedHostRuntimeGuard(config.GSDCommand)
			if err != nil {
				return fmt.Errorf("runtime admission: %w", err)
			}
			registry, err = gsd.LoadPinnedUnitRegistry(ctx, config.GSDCommand, config.GSDHome, config.GSDVersion)
			if err != nil {
				return fmt.Errorf("runtime admission: %w", err)
			}
			if err := gsd.ApplyPinnedHeadlessToolPatch(config.GSDCommand, config.GSDVersion); err != nil {
				return fmt.Errorf("runtime compatibility: %w", err)
			}
			if err := gsd.ApplyPinnedPromptToolPatch(config.GSDCommand, config.GSDHome, config.GSDVersion); err != nil {
				return fmt.Errorf("runtime compatibility: %w", err)
			}
			if err := runtimeGuard.BindPromptRuntime(config.GSDCommand, config.GSDHome, registry); err != nil {
				return fmt.Errorf("runtime admission: %w", err)
			}
		}
	}
	if err := gsd.NormalizeRuntimeSettings(config.GSDHome, config.CoordinatorModel, config.ImplementationModel, "high"); err != nil {
		return fmt.Errorf("runtime normalization: %w", err)
	}
	if err := gsd.ValidateRuntimeSettings(config.GSDHome, config.WorkDir, config.CoordinatorModel, "high"); err != nil {
		return fmt.Errorf("runtime admission: %w", err)
	}
	if err := gsd.ValidateModelPreferences(config.GSDHome, config.WorkDir, registry, config.CoordinatorModel, config.ImplementationModel, "high"); err != nil {
		return fmt.Errorf("runtime admission: %w", err)
	}
	if runtimeGuard != nil {
		if err := runtimeGuard.BindModelPolicy(config.WorkDir, config.CoordinatorModel, config.ImplementationModel, "high"); err != nil {
			return fmt.Errorf("runtime admission: %w", err)
		}
	}
	selectedModel, err := launchModelForCommand(config, registry, *command)
	if err != nil {
		return err
	}
	runner, err := gsd.NewRunner(gsd.Config{
		Command: config.GSDCommand, WorkDir: config.WorkDir, GSDHome: config.GSDHome, StateDir: config.StateDir,
		Model: selectedModel, Thinking: "high",
		Timeout:               time.Duration(config.TimeoutSeconds) * time.Second,
		HeartbeatInterval:     time.Duration(config.HeartbeatSeconds) * time.Second,
		MaxEventBytes:         config.MaxEventBytes,
		Container:             container,
		Registry:              registry,
		RuntimeGuard:          runtimeGuard,
		RequireSessionBinding: forceSessionBinding,
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
		return runHeadless(ctx, runner, config, registry, deliveryID(*issue), *issue, "", *command, nil, nil, *confirmDepth, *continueUnit, *decisionActor, *decisionBasis)
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
		return runHeadless(ctx, runner, config, registry, deliveryID(*issue), *issue, contextHash, "new-milestone", []string{"--context", path}, nil, *confirmDepth, false, *decisionActor, *decisionBasis)
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
	manager, repositoryLock, lockErr := lockDeliveryWorkspace(config)
	if lockErr != nil {
		return fmt.Errorf("startup delivery lock: %w", lockErr)
	}
	deliveryID := deliveryID(issue)
	var startupErr error
	var reconcileLease store.Lease
	var reconciliationLeaseHeld bool
	if existingDelivery, getErr := authority.GetDelivery(ctx, deliveryID); getErr == nil {
		reconcileOwner := fmt.Sprintf("reconcile-%d", time.Now().UTC().UnixNano())
		reconcileLease, startupErr = authority.AcquireReconciliationLease(ctx, deliveryID, reconcileOwner, time.Now().UTC(), 90*time.Second)
		if startupErr == nil {
			reconciliationLeaseHeld = true
			preRecoveryRun, runErr := authority.GetDeliveryRun(ctx, deliveryID)
			preRecoveryHead, inspectErr := shepherdgit.Inspect(ctx, config.WorkDir)
			startupErr = errors.Join(runErr, inspectErr)
			if startupErr == nil {
				startupErr = recoverPostGitPromotionJournals(ctx, authority, manager, repositoryLock,
					existingDelivery, reconcileLease)
			}
			if startupErr == nil {
				blocked, blockedErr := authority.HasBlockedPromotionJournal(ctx, deliveryID)
				startupErr = blockedErr
				if blockedErr == nil && blocked {
					startupErr = errors.New("blocked promotion journal requires journal-specific human recovery")
				}
			}
			if startupErr == nil {
				startupErr = validatePreGitPromotionJournals(ctx, authority, manager, existingDelivery,
					reconcileLease)
			}
			if startupErr == nil && config.Runtime == "host" && preRecoveryRun.State == domain.RunRunning {
				startupErr = reconcileInterruptedHostDelivery(ctx, authority, manager, repositoryLock,
					config, deliveryID, reconcileLease)
				if startupErr == nil {
					preRecoveryRun, startupErr = authority.GetDeliveryRun(ctx, deliveryID)
				}
			}
			if startupErr == nil {
				preRecoveryHead, startupErr = shepherdgit.Inspect(ctx, config.WorkDir)
			}
			if startupErr == nil {
				preRecoveryEffects, effectErr := openExternalEffectController(ctx, authority, reconcileLease,
					config, deliveryID, issue, preRecoveryRun.Generation, preRecoveryHead.HeadSHA)
				if effectErr == nil {
					effectErr = reconcileExternalEffects(ctx, authority, preRecoveryEffects, config.StateDir, deliveryID)
					effectErr = errors.Join(effectErr, preRecoveryEffects.Close())
				}
				startupErr = effectErr
			}
			pollCutoff := time.Now().UTC()
			if startupErr == nil {
				_, startupErr = consumeGitHubDecisionReplies(ctx, authority, reconcileLease,
					decisionReplyPollerFactory(), deliveryID, preRecoveryRun.Generation, preRecoveryHead.HeadSHA,
					decisionApplicationFence(ctx, repositoryLock, config.WorkDir, existingDelivery.Branch,
						preRecoveryHead.HeadSHA))
			}
			if startupErr == nil {
				startupErr = expireUnansweredDecisionRequests(ctx, authority, reconcileLease, deliveryID,
					pollCutoff, decisionApplicationFence(ctx, repositoryLock, config.WorkDir,
						existingDelivery.Branch, preRecoveryHead.HeadSHA))
			}
			if startupErr == nil {
				postDecisionRun, decisionErr := authority.GetDeliveryRun(ctx, deliveryID)
				blocking, blockingErr := authority.HasBlockingDecisionDisposition(ctx, deliveryID,
					postDecisionRun.Generation)
				startupErr = errors.Join(decisionErr, blockingErr)
				if startupErr == nil && blocking {
					startupErr = errors.New("human decision blocks recovered promotion")
				}
				if startupErr == nil {
					startupErr = authority.CancelUnansweredPromotionDecisions(ctx, reconcileLease,
						time.Now().UTC())
				}
				if startupErr == nil {
					openRequests, openErr := authority.ListOpenDecisionRequests(ctx, deliveryID)
					startupErr = openErr
					if openErr == nil && len(openRequests) > 0 {
						startupErr = errors.New("unrelated human decision defers recovered promotion")
					}
				}
			}
			if startupErr == nil {
				startupErr = recoverPromotionJournals(ctx, authority, manager, repositoryLock, existingDelivery, reconcileLease)
			}
		}
	} else if !errors.Is(getErr, sql.ErrNoRows) {
		startupErr = getErr
	}
	if startupErr != nil {
		if reconciliationLeaseHeld {
			startupErr = errors.Join(startupErr, authority.ReleaseLease(context.Background(), reconcileLease))
		}
		startupErr = errors.Join(startupErr, repositoryLock.Close())
		return fmt.Errorf("startup promotion recovery: %w", startupErr)
	}
	_, _, ensureErr := ensureIssueDelivery(ctx, authority, config, issue, contextHash)
	if ensureErr != nil {
		if reconciliationLeaseHeld {
			ensureErr = errors.Join(ensureErr, authority.ReleaseLease(context.Background(), reconcileLease))
		}
		_ = repositoryLock.Close()
		return ensureErr
	}
	if config.Runtime == "host" {
		if !reconciliationLeaseHeld {
			reconcileOwner := fmt.Sprintf("reconcile-%d", time.Now().UTC().UnixNano())
			reconcileLease, startupErr = authority.AcquireReconciliationLease(ctx, deliveryID, reconcileOwner, time.Now().UTC(), 90*time.Second)
			reconciliationLeaseHeld = startupErr == nil
		}
		if startupErr == nil {
			startupErr = reconcileInterruptedHostDelivery(ctx, authority, manager, repositoryLock,
				config, deliveryID, reconcileLease)
		}
	}
	if startupErr == nil && !reconciliationLeaseHeld {
		reconcileOwner := fmt.Sprintf("reconcile-%d", time.Now().UTC().UnixNano())
		reconcileLease, startupErr = authority.AcquireReconciliationLease(ctx, deliveryID, reconcileOwner,
			time.Now().UTC(), 90*time.Second)
		reconciliationLeaseHeld = startupErr == nil
	}
	if startupErr == nil {
		runState, runErr := authority.GetDeliveryRun(ctx, deliveryID)
		canonical, inspectErr := shepherdgit.Inspect(ctx, config.WorkDir)
		delivery, deliveryErr := authority.GetDelivery(ctx, deliveryID)
		startupErr = errors.Join(runErr, inspectErr, deliveryErr)
		if startupErr == nil && (canonical.Branch != delivery.Branch || canonical.HeadSHA == "") {
			startupErr = errors.New("external effect reconciliation requires the authorized canonical branch and head")
		}
		if startupErr == nil {
			_, startupErr = consumeGitHubDecisionReplies(ctx, authority, reconcileLease,
				decisionReplyPollerFactory(), deliveryID, runState.Generation, canonical.HeadSHA,
				decisionApplicationFence(ctx, repositoryLock, config.WorkDir, delivery.Branch, canonical.HeadSHA))
		}
		if startupErr == nil {
			runState, runErr = authority.GetDeliveryRun(ctx, deliveryID)
			startupErr = runErr
		}
		if startupErr == nil {
			effects, effectErr := openExternalEffectController(ctx, authority, reconcileLease, config,
				deliveryID, issue, runState.Generation, canonical.HeadSHA)
			if effectErr == nil {
				effectErr = reconcileExternalEffects(ctx, authority, effects, config.StateDir, deliveryID)
				effectErr = errors.Join(effectErr, effects.Close())
			}
			startupErr = effectErr
		}
	}
	if startupErr == nil && reconciliationLeaseHeld {
		runState, runErr := authority.GetDeliveryRun(ctx, deliveryID)
		canonical, inspectErr := shepherdgit.Inspect(ctx, config.WorkDir)
		delivery, deliveryErr := authority.GetDelivery(ctx, deliveryID)
		startupErr = errors.Join(runErr, inspectErr, deliveryErr)
		pollCutoff := time.Now().UTC()
		if startupErr == nil {
			_, startupErr = consumeGitHubDecisionReplies(ctx, authority, reconcileLease,
				decisionReplyPollerFactory(), deliveryID, runState.Generation, canonical.HeadSHA,
				decisionApplicationFence(ctx, repositoryLock, config.WorkDir, delivery.Branch, canonical.HeadSHA))
		}
		if startupErr == nil {
			startupErr = expireUnansweredDecisionRequests(ctx, authority, reconcileLease, deliveryID,
				pollCutoff, decisionApplicationFence(ctx, repositoryLock, config.WorkDir,
					delivery.Branch, canonical.HeadSHA))
		}
		if startupErr == nil {
			startupErr = authority.ResolveRecoveredPromotionDecisions(ctx, reconcileLease, time.Now().UTC())
		}
	}
	if reconciliationLeaseHeld {
		startupErr = errors.Join(startupErr, authority.ReleaseLease(context.Background(), reconcileLease))
	}
	startupErr = errors.Join(startupErr, repositoryLock.Close())
	if startupErr != nil {
		return fmt.Errorf("startup attempt reconciliation: %w", startupErr)
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		snapshot, err := superviseQuery(ctx, authority, runner, config, deliveryID)
		if err != nil {
			return fmt.Errorf("supervise query: %w", err)
		}
		decision, err := supervisor.DecideWithRegistry(snapshot, registry)
		if err != nil {
			return fmt.Errorf("supervise policy: %w", err)
		}
		switch decision.Kind {
		case supervisor.DecisionFinalGate:
			_, finalRepositoryLock, gateErr := lockDeliveryWorkspace(config)
			if gateErr != nil {
				return gateErr
			}
			runState, runErr := authority.GetDeliveryRun(ctx, deliveryID)
			delivery, deliveryErr := authority.GetDelivery(ctx, deliveryID)
			if runErr != nil || deliveryErr != nil {
				_ = finalRepositoryLock.Close()
				return errors.Join(runErr, deliveryErr)
			}
			now := time.Now().UTC()
			gateLease, gateErr := authority.AcquireLease(ctx, deliveryID,
				fmt.Sprintf("final-gate-%d", now.UnixNano()), now, 90*time.Second)
			if gateErr != nil {
				_ = finalRepositoryLock.Close()
				return gateErr
			}
			canonical, inspectErr := shepherdgit.Inspect(ctx, config.WorkDir)
			if inspectErr == nil && canonical.Branch != delivery.Branch {
				inspectErr = errors.New("final human gate requires the immutable delivery branch")
			}
			if inspectErr == nil {
				inspectErr = shepherdgit.RequireClean(canonical)
			}
			if inspectErr == nil {
				inspectErr = validateFinalGateGSDState(ctx, authority, config.StateDir, config.WorkDir,
					deliveryID, runState.Generation, canonical.HeadSHA)
			}
			if inspectErr == nil {
				gateErr = authority.ProjectFinalHumanGate(ctx, gateLease, runState.Generation,
					canonical.HeadSHA, time.Now().UTC())
			} else {
				gateErr = inspectErr
			}
			gateErr = errors.Join(gateErr, authority.ReleaseLease(context.Background(), gateLease),
				finalRepositoryLock.Close())
			if gateErr != nil {
				return gateErr
			}
			return emitSuperviseStatus(deliveryID, decision)
		case supervisor.DecisionBlocked:
			waited, waitErr := waitForHumanDecision(ctx, authority, config, deliveryID, decision)
			if waitErr != nil {
				return waitErr
			}
			if waited {
				continue
			}
			if emitErr := emitSuperviseStatus(deliveryID, decision); emitErr != nil {
				return emitErr
			}
			return commandExitError{code: 10, err: errors.New(decision.Reason)}
		case supervisor.DecisionDispatch:
			expectedModel, modelErr := modelForUnitType(config, registry, snapshot.Next.UnitType)
			if modelErr != nil {
				return modelErr
			}
			unitRunner, modelErr := runner.WithModel(expectedModel)
			if modelErr != nil {
				return modelErr
			}
			err := runHeadless(ctx, unitRunner, config, registry, deliveryID, issue, "", decision.Command, nil, &snapshot, confirmDepth, false, decisionActor, decisionBasis)
			if err == nil {
				continue
			}
			if recovery.ShouldRetry(err) {
				integrationRetryBoundary()
				continue
			}
			blocked := decision
			blocked.Kind = supervisor.DecisionBlocked
			blocked.Reason = string(recovery.Classify(gsd.Result{Terminal: gsd.TerminalError, Err: err}).Class)
			waited, waitErr := waitForHumanDecision(ctx, authority, config, deliveryID, blocked)
			if waitErr != nil {
				return waitErr
			}
			if waited {
				continue
			}
			if emitErr := emitSuperviseStatus(deliveryID, blocked); emitErr != nil {
				return emitErr
			}
			return commandExitError{code: 10, err: err}
		default:
			return fmt.Errorf("unknown supervise decision %q", decision.Kind)
		}
	}
}

func waitForHumanDecision(
	ctx context.Context,
	authority *store.Store,
	config fileConfig,
	deliveryID string,
	decision supervisor.Decision,
) (bool, error) {
	runState, err := authority.GetDeliveryRun(ctx, deliveryID)
	if err != nil {
		return false, err
	}
	if runState.State != domain.RunAwaitingDecision || exitAwaitingDecision() {
		return false, nil
	}
	interval := time.Duration(config.HeartbeatSeconds) * time.Second
	if interval <= 0 {
		interval = 15 * time.Second
	}
	for {
		if err := emitSuperviseStatus(deliveryID, decision); err != nil {
			return true, err
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return true, ctx.Err()
		case <-timer.C:
		}
		_, repositoryLock, err := lockDeliveryWorkspace(config)
		if err != nil {
			return true, err
		}
		owner := fmt.Sprintf("decision-poll-%d", time.Now().UTC().UnixNano())
		lease, err := authority.AcquireReconciliationLease(ctx, deliveryID, owner, time.Now().UTC(), 90*time.Second)
		if err != nil {
			return true, errors.Join(err, repositoryLock.Close())
		}
		canonical, inspectErr := shepherdgit.Inspect(ctx, config.WorkDir)
		delivery, deliveryErr := authority.GetDelivery(ctx, deliveryID)
		inspectErr = errors.Join(inspectErr, deliveryErr)
		if inspectErr == nil && canonical.Branch != delivery.Branch {
			inspectErr = errors.New("decision polling requires the immutable delivery branch")
		}
		pollCutoff := time.Now().UTC()
		if inspectErr == nil {
			_, inspectErr = consumeGitHubDecisionReplies(ctx, authority, lease, decisionReplyPollerFactory(),
				deliveryID, runState.Generation, canonical.HeadSHA,
				decisionApplicationFence(ctx, repositoryLock, config.WorkDir, delivery.Branch, canonical.HeadSHA))
		}
		if inspectErr == nil {
			inspectErr = expireUnansweredDecisionRequests(ctx, authority, lease, deliveryID, pollCutoff,
				decisionApplicationFence(ctx, repositoryLock, config.WorkDir, delivery.Branch, canonical.HeadSHA))
		}
		releaseErr := authority.ReleaseLease(context.Background(), lease)
		if err := errors.Join(inspectErr, releaseErr, repositoryLock.Close()); err != nil {
			return true, err
		}
		runState, err = authority.GetDeliveryRun(ctx, deliveryID)
		if err != nil {
			return true, err
		}
		if runState.State != domain.RunAwaitingDecision {
			return true, nil
		}
	}
}

func requireMatchingAttemptDispatch(canonical, attempt gsd.WorkflowSnapshot) error {
	if canonical.MilestoneID != attempt.MilestoneID || canonical.Phase != attempt.Phase || canonical.Next.Action != attempt.Next.Action || canonical.Next.UnitType != attempt.Next.UnitType || canonical.Next.UnitID != attempt.Next.UnitID {
		return fmt.Errorf("attempt GSD state does not match canonical dispatch: canonical=%s/%s %s %s/%s attempt=%s/%s %s %s/%s", canonical.MilestoneID, canonical.Phase, canonical.Next.Action, canonical.Next.UnitType, canonical.Next.UnitID, attempt.MilestoneID, attempt.Phase, attempt.Next.Action, attempt.Next.UnitType, attempt.Next.UnitID)
	}
	return nil
}

func rejectAmbiguousAttemptRecovery(ctx context.Context, authorityStore *store.Store, deliveryID string) error {
	blocked, err := authorityStore.HasBlockedPromotionJournal(ctx, deliveryID)
	if err != nil {
		return err
	}
	if blocked {
		return errors.New("blocked promotion journal requires journal-specific human recovery")
	}
	records, err := authorityStore.ListAttemptWorktrees(ctx, deliveryID)
	if err != nil {
		return err
	}
	if attemptRecoveryAmbiguous(records) {
		return errors.New("ambiguous attempt ownership blocks governed repository access")
	}
	return nil
}

func superviseQuery(ctx context.Context, authorityStore *store.Store, runner *gsd.Runner, config fileConfig, deliveryID string) (snapshot gsd.WorkflowSnapshot, returnErr error) {
	_, repositoryLock, err := lockDeliveryWorkspace(config)
	if err != nil {
		return gsd.WorkflowSnapshot{}, err
	}
	defer func() { returnErr = errors.Join(returnErr, repositoryLock.Close()) }()
	owner := fmt.Sprintf("supervise-query-%d", time.Now().UTC().UnixNano())
	lease, err := authorityStore.AcquireLease(ctx, deliveryID, owner, time.Now().UTC(), 90*time.Second)
	if err != nil {
		return gsd.WorkflowSnapshot{}, err
	}
	defer func() { returnErr = errors.Join(returnErr, authorityStore.ReleaseLease(context.Background(), lease)) }()
	if err := rejectAmbiguousAttemptRecovery(ctx, authorityStore, deliveryID); err != nil {
		return gsd.WorkflowSnapshot{}, err
	}
	snapshot, err = runner.Query(ctx)
	if err != nil {
		return gsd.WorkflowSnapshot{}, err
	}
	if err := authorityStore.CheckLease(ctx, lease, time.Now().UTC()); err != nil {
		return gsd.WorkflowSnapshot{}, err
	}
	if err := repositoryLock.Check(); err != nil {
		return gsd.WorkflowSnapshot{}, err
	}
	return snapshot, nil
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
	manager, repositoryLock, err := lockDeliveryWorkspace(config)
	if err != nil {
		return err
	}
	defer func() { _ = repositoryLock.Close() }()
	if err := recoverExistingPromotionAtStartup(ctx, authority, manager, repositoryLock, deliveryID); err != nil {
		return err
	}
	owner := fmt.Sprintf("repair-%d", time.Now().UTC().UnixNano())
	lease, err := authority.AcquireLease(ctx, deliveryID, owner, time.Now().UTC(), 90*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = authority.ReleaseLease(context.Background(), lease) }()
	if err := repositoryLock.Check(); err != nil {
		return err
	}
	if err := rejectAmbiguousAttemptRecovery(ctx, authority, deliveryID); err != nil {
		return err
	}
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
	manager, repositoryLock, err := lockDeliveryWorkspace(config)
	if err != nil {
		return err
	}
	defer func() { _ = repositoryLock.Close() }()
	if err := recoverExistingPromotionAtStartup(ctx, authority, manager, repositoryLock, deliveryID); err != nil {
		return err
	}
	if _, _, err := ensureIssueDelivery(ctx, authority, config, issue, contextHash); err != nil {
		return err
	}
	owner := fmt.Sprintf("adopt-%d", time.Now().UTC().UnixNano())
	lease, err := authority.AcquireLease(ctx, deliveryID, owner, time.Now().UTC(), 90*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = authority.ReleaseLease(context.Background(), lease) }()
	if err := repositoryLock.Check(); err != nil {
		return err
	}
	if err := rejectAmbiguousAttemptRecovery(ctx, authority, deliveryID); err != nil {
		return err
	}
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
	manager, repositoryLock, err := lockDeliveryWorkspace(config)
	if err != nil {
		return err
	}
	defer func() { _ = repositoryLock.Close() }()
	if err := recoverExistingPromotionAtStartup(ctx, authority, manager, repositoryLock, deliveryID); err != nil {
		return err
	}
	owner := fmt.Sprintf("query-%d", time.Now().UTC().UnixNano())
	lease, err := authority.AcquireLease(ctx, deliveryID, owner, time.Now().UTC(), 90*time.Second)
	if err != nil {
		return err
	}
	defer func() { _ = authority.ReleaseLease(context.Background(), lease) }()
	if err := repositoryLock.Check(); err != nil {
		return err
	}
	if err := rejectAmbiguousAttemptRecovery(ctx, authority, deliveryID); err != nil {
		return err
	}
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
	if !filepath.IsAbs(config.WorkDir) || !filepath.IsAbs(config.GSDHome) || !filepath.IsAbs(config.StateDir) || !filepath.IsAbs(config.AttemptRoot) {
		return fileConfig{}, errors.New("work_dir, gsd_home, state_dir, and attempt_root must be absolute")
	}
	for name, value := range map[string]string{
		"work_dir": config.WorkDir, "gsd_home": config.GSDHome,
		"state_dir": config.StateDir, "attempt_root": config.AttemptRoot,
	} {
		if filepath.Clean(value) != value {
			return fileConfig{}, fmt.Errorf("%s must be a clean absolute path", name)
		}
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
	if within, err := pathWithin(config.WorkDir, config.AttemptRoot); err != nil || within {
		return fileConfig{}, errors.New("attempt_root must be outside the canonical work directory")
	}
	if within, err := pathWithin(config.StateDir, config.AttemptRoot); err != nil || within {
		return fileConfig{}, errors.New("protected state_dir must not contain attempt_root")
	}
	if within, err := pathWithin(config.AttemptRoot, config.StateDir); err != nil || within {
		return fileConfig{}, errors.New("attempt_root must not contain protected state_dir")
	}
	return config, nil
}

func lockDeliveryWorkspace(config fileConfig) (*workspace.Manager, *workspace.RepositoryLock, error) {
	manager, err := workspace.NewManager(config.WorkDir, config.AttemptRoot)
	if err != nil {
		return nil, nil, err
	}
	lock, err := manager.TryAcquireRepositoryLock()
	if err != nil {
		return nil, nil, err
	}
	return manager, lock, nil
}

func recoverExistingPromotionAtStartup(ctx context.Context, authorityStore *store.Store, manager *workspace.Manager, repositoryLock *workspace.RepositoryLock, deliveryID string) error {
	delivery, err := authorityStore.GetDelivery(ctx, deliveryID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	owner := fmt.Sprintf("promotion-recovery-%d", time.Now().UTC().UnixNano())
	lease, err := authorityStore.AcquireReconciliationLease(ctx, deliveryID, owner, time.Now().UTC(), 90*time.Second)
	if err != nil {
		return err
	}
	recoveryErr := recoverPostGitPromotionJournals(ctx, authorityStore, manager, repositoryLock, delivery, lease)
	if recoveryErr == nil {
		blocked, blockedErr := authorityStore.HasBlockedPromotionJournal(ctx, deliveryID)
		recoveryErr = blockedErr
		if blockedErr == nil && blocked {
			recoveryErr = errors.New("blocked promotion journal requires journal-specific human recovery")
		}
	}
	if recoveryErr == nil {
		journals, listErr := authorityStore.ListIncompletePromotionJournals(ctx, deliveryID)
		recoveryErr = listErr
		if listErr == nil && len(journals) > 0 {
			recoveryErr = errors.New("pre-Git promotion recovery requires supervise decision reconciliation")
		}
	}
	releaseErr := authorityStore.ReleaseLease(context.Background(), lease)
	return errors.Join(recoveryErr, releaseErr)
}

func resolveHumanClearedAttempts(ctx context.Context, authorityStore *store.Store, config fileConfig,
	deliveryID string, heldManager *workspace.Manager, heldLease *store.Lease,
) error {
	blocked, err := authorityStore.HasBlockedPromotionJournal(ctx, deliveryID)
	if err != nil {
		return err
	}
	if blocked {
		return errors.New("blocked promotion journal requires journal-specific human recovery")
	}
	manager := heldManager
	var lease store.Lease
	var repositoryLock *workspace.RepositoryLock
	ownsFence := heldManager == nil || heldLease == nil
	if ownsFence {
		manager, repositoryLock, err = lockDeliveryWorkspace(config)
		if err != nil {
			return err
		}
		owner := fmt.Sprintf("human-recovery-%d", time.Now().UTC().UnixNano())
		lease, err = authorityStore.AcquireReconciliationLease(ctx, deliveryID, owner,
			time.Now().UTC(), 90*time.Second)
		if err != nil {
			_ = repositoryLock.Close()
			return err
		}
	} else {
		lease = *heldLease
	}
	records, resolveErr := authorityStore.ListAttemptWorktrees(ctx, deliveryID)
	if resolveErr == nil {
		for _, record := range records {
			hasJournal, journalErr := authorityStore.AttemptHasPromotionJournal(ctx, record.Key())
			if journalErr != nil {
				resolveErr = journalErr
				break
			}
			if hasJournal && record.State == store.AttemptWorktreePromoting {
				resolveErr = errors.New("journal-owned promotion requires journal-specific human recovery")
				break
			}
			if record.State == store.AttemptWorktreeCleanupComplete ||
				(record.ResourcesCreated && record.State != store.AttemptWorktreeRunning && record.State != store.AttemptWorktreePromoting) {
				continue
			}
			attempt := workspace.AttemptWorktree{Root: record.Path, Branch: record.Branch,
				Identity: workspace.AttemptIdentity{DeliveryID: record.DeliveryID, Generation: record.Generation,
					UnitID: record.UnitID, Attempt: record.Attempt, BaseHead: record.BaseHead}}
			if resolveErr = manager.ProveOwnedResourcesAbsent(ctx, attempt); resolveErr != nil {
				break
			}
			resolveErr = authorityStore.ResolveHumanClearedAttempt(ctx, record.Key(), record.ControllerOwner, record.ControllerEpoch, lease)
			if resolveErr != nil {
				break
			}
		}
	}
	if !ownsFence {
		return resolveErr
	}
	releaseErr := authorityStore.ReleaseLease(context.Background(), lease)
	lockErr := repositoryLock.Close()
	return errors.Join(resolveErr, releaseErr, lockErr)
}

type attemptLifecycle struct {
	store *store.Store
	key   store.AttemptWorktreeKey
	owner string
	epoch int64
	state store.AttemptWorktreeState
}

func (l *attemptLifecycle) transition(ctx context.Context, next store.AttemptWorktreeState, update store.AttemptWorktreeUpdate) error {
	record, err := l.store.TransitionAttemptWorktree(ctx, l.key, l.owner, l.epoch, next, update)
	if err != nil {
		return err
	}
	l.state = record.State
	return nil
}

func (l *attemptLifecycle) retain(ctx context.Context, failureClass string) error {
	if l == nil || l.state == store.AttemptWorktreeRetainedForRecovery ||
		l.state == store.AttemptWorktreeCleanupPending || l.state == store.AttemptWorktreeCleanupComplete ||
		l.state == store.AttemptWorktreeCleanupBlocked {
		return nil
	}
	switch l.state {
	case store.AttemptWorktreeCreated, store.AttemptWorktreePrepared, store.AttemptWorktreeRunning,
		store.AttemptWorktreeValidated, store.AttemptWorktreeRatified:
		return l.transition(ctx, store.AttemptWorktreeRetainedForRecovery, store.AttemptWorktreeUpdate{FailureClass: failureClass})
	default:
		return nil
	}
}

func attemptRecoveryAmbiguous(records []store.AttemptWorktreeRecord) bool {
	for _, record := range records {
		if record.State == store.AttemptWorktreeCleanupComplete {
			continue
		}
		if !record.ResourcesCreated || record.State == store.AttemptWorktreeRunning || record.State == store.AttemptWorktreePromoting {
			return true
		}
	}
	return false
}

func reconcileInterruptedHostDelivery(ctx context.Context, authorityStore *store.Store,
	manager *workspace.Manager, repositoryLock *workspace.RepositoryLock, config fileConfig,
	deliveryID string, lease store.Lease,
) error {
	records, err := authorityStore.ListAttemptWorktrees(ctx, deliveryID)
	if err != nil {
		return err
	}
	ungovernedRecords := make([]store.AttemptWorktreeRecord, 0, len(records))
	for _, record := range records {
		if record.State == store.AttemptWorktreePromoting {
			hasJournal, journalErr := authorityStore.AttemptHasPromotionJournal(ctx, record.Key())
			if journalErr != nil {
				return journalErr
			}
			if hasJournal {
				continue
			}
		}
		ungovernedRecords = append(ungovernedRecords, record)
	}
	if attemptRecoveryAmbiguous(ungovernedRecords) {
		return errors.Join(errors.New("ambiguous live or partially owned attempt requires human recovery"),
			authorityStore.ReconcileInterruptedDelivery(ctx, lease, domain.RunAwaitingDecision))
	}
	reconcileErr := reconcileAttemptWorktrees(ctx, authorityStore, manager, repositoryLock, deliveryID, lease)
	target := domain.RunReady
	if reconcileErr == nil {
		runState, runErr := authorityStore.GetDeliveryRun(ctx, deliveryID)
		delivery, deliveryErr := authorityStore.GetDelivery(ctx, deliveryID)
		canonical, inspectErr := shepherdgit.Inspect(ctx, config.WorkDir)
		reconcileErr = errors.Join(runErr, deliveryErr, inspectErr)
		if reconcileErr == nil {
			outstanding, gateErr := authorityStore.HasOutstandingDecisionRequest(ctx, deliveryID,
				delivery.Repository, delivery.Issue, delivery.PullRequest, runState.Generation,
				canonical.HeadSHA)
			reconcileErr = gateErr
			if outstanding {
				target = domain.RunAwaitingDecision
			}
		}
	}
	if reconcileErr != nil {
		target = domain.RunAwaitingDecision
	}
	return errors.Join(reconcileErr,
		authorityStore.ReconcileInterruptedDelivery(ctx, lease, target))
}

func reconcileAttemptWorktrees(ctx context.Context, authorityStore *store.Store, manager *workspace.Manager, repositoryLock *workspace.RepositoryLock, deliveryID string, lease store.Lease) error {
	if err := repositoryLock.Check(); err != nil {
		return err
	}
	records, err := authorityStore.ListAttemptWorktrees(ctx, deliveryID)
	if err != nil {
		return err
	}
	for _, record := range records {
		if record.State == store.AttemptWorktreeCleanupComplete {
			continue
		}
		if record.State == store.AttemptWorktreePromoting {
			hasJournal, journalErr := authorityStore.AttemptHasPromotionJournal(ctx, record.Key())
			if journalErr != nil {
				return journalErr
			}
			if hasJournal {
				continue
			}
		}
		claimed := record
		if record.ControllerOwner == lease.Owner && record.ControllerEpoch == lease.Epoch {
			switch record.State {
			case store.AttemptWorktreePromoted:
				claimed, err = authorityStore.TransitionAttemptWorktree(ctx, record.Key(), lease.Owner, lease.Epoch,
					store.AttemptWorktreeCleanupPending, store.AttemptWorktreeUpdate{})
				if err != nil {
					return err
				}
			case store.AttemptWorktreeCleanupPending:
			default:
				continue
			}
		} else {
			var claimErr error
			claimed, claimErr = authorityStore.ClaimAttemptWorktreeCleanup(ctx, record.Key(), record.ControllerOwner, record.ControllerEpoch, lease.Owner, lease.Epoch)
			if claimErr != nil {
				return claimErr
			}
		}
		expectedHead := claimed.CandidateHead
		if expectedHead == "" {
			expectedHead = claimed.BaseHead
		}
		owned := workspace.OwnedAttempt{Attempt: workspace.AttemptWorktree{
			Root: claimed.Path, Branch: claimed.Branch,
			Identity: workspace.AttemptIdentity{DeliveryID: claimed.DeliveryID, Generation: claimed.Generation,
				UnitID: claimed.UnitID, Attempt: claimed.Attempt, BaseHead: claimed.BaseHead},
		}, ExpectedHead: expectedHead}
		if leaseErr := authorityStore.CheckLease(ctx, lease, time.Now().UTC()); leaseErr != nil {
			return leaseErr
		}
		if lockErr := repositoryLock.Check(); lockErr != nil {
			return lockErr
		}
		result, cleanupErr := reconcileOwnedAttempt(ctx, manager, owned)
		if cleanupErr != nil || result == workspace.ReconcileBlocked {
			transitionErr := (&attemptLifecycle{store: authorityStore, key: claimed.Key(), owner: lease.Owner,
				epoch: lease.Epoch, state: store.AttemptWorktreeCleanupPending}).transition(ctx,
				store.AttemptWorktreeCleanupBlocked, store.AttemptWorktreeUpdate{CleanupError: "owned_resource_mismatch_or_cleanup_failure"})
			return errors.Join(cleanupErr, transitionErr)
		}
		if result != workspace.ReconcileComplete {
			return errors.New("attempt reconciliation did not reach a terminal cleanup state")
		}
		if _, transitionErr := authorityStore.TransitionAttemptWorktree(ctx, claimed.Key(), lease.Owner, lease.Epoch,
			store.AttemptWorktreeCleanupComplete, store.AttemptWorktreeUpdate{}); transitionErr != nil {
			return transitionErr
		}
	}
	return nil
}

func runHeadless(ctx context.Context, runner *gsd.Runner, config fileConfig, registry gsd.UnitRegistry, deliveryID string, issue int, contextHash, command string, args []string, supervisedSnapshot *gsd.WorkflowSnapshot, confirmDepth, continueUnit bool, decisionActor, decisionBasis string) (returnErr error) {
	if err := os.MkdirAll(config.StateDir, 0o700); err != nil {
		return err
	}
	authority, err := store.Open(ctx, filepath.Join(config.StateDir, "authority.db"))
	if err != nil {
		return err
	}
	defer authority.Close()
	attemptManager, repositoryLock, err := lockDeliveryWorkspace(config)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, repositoryLock.Close()) }()
	if err := recoverExistingPromotionAtStartup(ctx, authority, attemptManager, repositoryLock, deliveryID); err != nil {
		return err
	}
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
	executionLeaseTTL := time.Duration(2*config.TimeoutSeconds+120) * time.Second
	lease, err := authority.AcquireLease(ctx, deliveryID, executionID, time.Now().UTC(), executionLeaseTTL)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, authority.ReleaseLease(context.Background(), lease)) }()
	if err := rejectAmbiguousAttemptRecovery(ctx, authority, deliveryID); err != nil {
		return err
	}
	startSnapshot, err := shepherdgit.Inspect(ctx, config.WorkDir)
	if err != nil {
		return err
	}
	if startSnapshot.Branch != delivery.Branch {
		return errors.New("canonical worktree is not on the authorized delivery branch")
	}
	if err := shepherdgit.RequireClean(startSnapshot); err != nil {
		return err
	}
	trustedUnit := command
	var before gsd.WorkflowSnapshot
	if supervisedSnapshot != nil {
		before = *supervisedSnapshot
	} else {
		var queryErr error
		before, queryErr = runner.Query(ctx)
		if queryErr != nil {
			return fmt.Errorf("pre-run fenced query failed: %w", queryErr)
		}
	}
	canonicalUnitType := ""
	canonicalUnit := false
	if command == "next" && before.Next.Action != "dispatch" {
		return errors.New("next is allowed only for an explicit canonical dispatch")
	}
	if command != "next" || before.Next.Action == "dispatch" {
		canonicalUnitType, canonicalUnit, err = registry.CanonicalUnitForInvocation(command, before.Next.UnitType)
		if err != nil {
			return err
		}
	}
	if canonicalUnit {
		if before.Next.UnitID == "" {
			return errors.New("canonical dispatch requires a non-empty unit ID")
		}
		if command == "discuss" && before.Next.UnitID != before.MilestoneID {
			return errors.New("discussion target differs from the canonical unit ID")
		}
		trustedUnit = canonicalUnitType + "/" + before.Next.UnitID
	} else if before.Next.UnitID != "" {
		trustedUnit = before.Next.UnitType + "/" + before.Next.UnitID
	}
	if canonicalUnit && config.Runtime == "host" {
		if err := reconcileAttemptWorktrees(ctx, authority, attemptManager, repositoryLock, deliveryID, lease); err != nil {
			return fmt.Errorf("reconcile durable attempt worktrees: %w", err)
		}
	}
	if command == "discuss" {
		if before.Next.Action != "dispatch" || before.Next.UnitType != "discuss-milestone" || before.MilestoneID == "" {
			return errors.New("targeted discuss is allowed only when the canonical next unit is discuss-milestone")
		}
		args = []string{before.MilestoneID}
	}
	if err := rejectUnqualifiedContinuation(continueUnit); err != nil {
		return err
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
	if canonicalUnit {
		contractRunner, contractErr := runner.WithUnitContract(canonicalUnitType)
		if contractErr != nil {
			return contractErr
		}
		runner = contractRunner
	}
	attempt, err := authority.BeginAttempt(ctx, deliveryID, executionID)
	if err != nil {
		return err
	}
	finishedAttempt := false
	defer func() {
		if !finishedAttempt {
			if finishErr := authority.FinishAttempt(context.Background(), deliveryID, executionID, domain.RunFailed); finishErr != nil {
				returnErr = errors.Join(returnErr, finishErr)
			}
		}
	}()
	effects, err := openExternalEffectController(ctx, authority, lease, config, deliveryID, issue,
		attempt.Generation, startSnapshot.HeadSHA)
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, effects.Close()) }()
	recoveryDispatch, dispatchErr := waitForRecoveryDispatch(ctx, authority, lease, deliveryID, attempt.Generation, trustedUnit, startSnapshot.HeadSHA, executionID)
	if dispatchErr != nil {
		switch {
		case errors.Is(dispatchErr, store.ErrRetryBudgetExhausted), errors.Is(dispatchErr, store.ErrRecoveryDecisionPending):
			class := recovery.FailureRetryExhausted
			if errors.Is(dispatchErr, store.ErrRecoveryDecisionPending) {
				class = recovery.FailureHumanRequired
			}
			decisionErr := recovery.MarkDecision(recovery.Failure{Class: class}, recovery.ActionAwaitDecision, dispatchErr)
			decisionRequestErr := persistAndPublishDecisionRequest(ctx, authority, effects,
				deliveryID, issue, trustedUnit, attempt.Generation, startSnapshot.HeadSHA, string(class), decisionErr)
			target := domain.RunAwaitingDecision
			if decisionRequestErr != nil {
				publicationFailure := recovery.Classify(gsd.Result{Terminal: gsd.TerminalError, Err: decisionRequestErr})
				if publicationFailure.Class == recovery.FailureGitHubPublishUncertain || publicationFailure.Class == recovery.FailureOutboxUncertain {
					decisionErr = recovery.MarkDecision(publicationFailure, recovery.ActionBlock, decisionErr)
					target = domain.RunBlocked
				}
			}
			finishErr := authority.FinishAttempt(ctx, deliveryID, executionID, target)
			finishedAttempt = finishErr == nil
			return commandExitError{code: 10, err: errors.Join(decisionErr, decisionRequestErr, finishErr)}
		case errors.Is(dispatchErr, store.ErrRecoveryTerminal):
			finishErr := authority.FinishAttempt(ctx, deliveryID, executionID, domain.RunBlocked)
			finishedAttempt = finishErr == nil
			return commandExitError{code: 10, err: errors.Join(recovery.MarkDecision(recovery.Failure{Class: recovery.FailureHumanRequired}, recovery.ActionBlock, dispatchErr), finishErr)}
		default:
			return dispatchErr
		}
	}
	if recoveryDispatch.SelectedAction != "" {
		lockErr := repositoryLock.Check()
		_, _, fingerprintErr := canonicalRecoveryFingerprint(ctx, config.WorkDir)
		authorityHash, authorityErr := recoveryAuthorityScopeHash(ctx, recoveryGovernanceRequest{
			Config: config, DeliveryID: deliveryID, Generation: attempt.Generation, UnitID: trustedUnit,
			HeadSHA: startSnapshot.HeadSHA, WriteScope: issueContext.WriteScope, AuthorizedBranch: delivery.Branch,
		})
		leaseErr := authority.CheckLease(ctx, lease, time.Now().UTC())
		if lockErr != nil || fingerprintErr != nil || authorityErr != nil || leaseErr != nil || authorityHash != recoveryDispatch.AuthorityScopeHash {
			if leaseErr != nil {
				finishedAttempt = true
				return errors.Join(errors.New("recovery dispatch authority changed during backoff"), lockErr,
					fingerprintErr, authorityErr, leaseErr)
			}
			dispositionErr := authority.FailRecoveryDispatch(ctx, recoveryDispatch, executionID, lease.Epoch)
			return errors.Join(errors.New("recovery dispatch authority changed during backoff"), lockErr,
				fingerprintErr, authorityErr, dispositionErr)
		}
	}
	if err := executeRecoveryPlan(ctx, recoveryDispatch, authority, attemptManager, repositoryLock, lease, config); err != nil {
		dispositionErr := authority.FailRecoveryDispatch(ctx, recoveryDispatch, executionID, lease.Epoch)
		return errors.Join(fmt.Errorf("execute bounded recovery plan: %w", err), dispositionErr)
	}
	if err := authority.CheckLease(ctx, lease, time.Now().UTC()); err != nil {
		finishedAttempt = true
		return err
	}
	unitAttemptKey := store.UnitAttemptKey{DeliveryID: deliveryID, Generation: attempt.Generation,
		UnitID: trustedUnit, HeadSHA: startSnapshot.HeadSHA}
	unitAttempt, err := authority.BeginUnitAttemptFenced(ctx, unitAttemptKey, totalUnitAttemptLimit(config.MaxUnitAttempts), executionID, lease.Epoch)
	if err != nil {
		if errors.Is(err, store.ErrRecoveryClaimFenced) {
			finishedAttempt = true
			return err
		}
		dispositionErr := authority.FailRecoveryDispatch(ctx, recoveryDispatch, executionID, lease.Epoch)
		target := domain.RunFailed
		if errors.Is(err, store.ErrRetryBudgetExhausted) {
			target = domain.RunAwaitingDecision
		}
		if finishErr := authority.FinishAttempt(ctx, deliveryID, executionID, target); finishErr != nil {
			return errors.Join(err, dispositionErr, finishErr)
		}
		finishedAttempt = true
		if errors.Is(err, store.ErrRetryBudgetExhausted) {
			return commandExitError{code: 10, err: errors.Join(err, dispositionErr)}
		}
		return errors.Join(err, dispositionErr)
	}
	unitAttemptFinished := false
	defer func() {
		if !unitAttemptFinished {
			if finishErr := authority.FinishUnitAttemptFenced(context.Background(), unitAttemptKey, unitAttempt.Attempts, "controller_interrupted", executionID, lease.Epoch); finishErr != nil {
				finishedAttempt = true
				returnErr = errors.Join(returnErr, finishErr)
			}
		}
	}()
	executionWorkDir := config.WorkDir
	executionRunner := runner
	var attemptWorktree *workspace.AttemptWorktree
	var lifecycle *attemptLifecycle
	if canonicalUnit && config.Runtime != "host" {
		return errors.New("canonical unit supervision requires host runtime disposable attempt worktrees")
	}
	if canonicalUnit && config.Runtime == "host" {
		identity := workspace.AttemptIdentity{DeliveryID: deliveryID, Generation: attempt.Generation,
			UnitID: trustedUnit, Attempt: unitAttempt.Attempts, BaseHead: startSnapshot.HeadSHA}
		planned, planErr := attemptManager.Plan(identity)
		if planErr != nil {
			return planErr
		}
		record := store.AttemptWorktreeRecord{DeliveryID: deliveryID, Generation: attempt.Generation,
			UnitID: trustedUnit, Attempt: unitAttempt.Attempts, Branch: planned.Branch, Path: planned.Root,
			BaseHead: startSnapshot.HeadSHA, State: store.AttemptWorktreeCreated,
			ControllerOwner: lease.Owner, ControllerEpoch: lease.Epoch}
		if err := authority.CreateAttemptWorktree(ctx, record); err != nil {
			return err
		}
		lifecycle = &attemptLifecycle{store: authority, key: record.Key(), owner: lease.Owner,
			epoch: lease.Epoch, state: store.AttemptWorktreeCreated}
		defer func() {
			returnErr = errors.Join(returnErr, lifecycle.retain(context.Background(), "controller_interrupted"))
		}()
		attemptTree, createErr := attemptManager.Create(ctx, identity)
		if createErr != nil {
			return errors.Join(createErr, lifecycle.retain(context.Background(), "worktree_creation_failure"))
		}
		attemptWorktree = &attemptTree
		if verifyErr := attemptManager.VerifyCreatedAttempt(ctx, attemptTree); verifyErr != nil {
			return errors.Join(verifyErr, lifecycle.retain(context.Background(), "resource_verification_failure"))
		}
		if _, confirmErr := authority.ConfirmAttemptWorktreeResources(ctx, record.Key(), lease.Owner, lease.Epoch); confirmErr != nil {
			return errors.Join(confirmErr, lifecycle.retain(context.Background(), "resource_confirmation_failure"))
		}
		if prepareErr := prepareAttemptGSDState(ctx, attemptManager, attemptTree); prepareErr != nil {
			return errors.Join(fmt.Errorf("prepare attempt GSD state: %w", prepareErr), lifecycle.retain(context.Background(), "preparation_failure"))
		}
		if err := lifecycle.transition(ctx, store.AttemptWorktreePrepared, store.AttemptWorktreeUpdate{}); err != nil {
			return err
		}
		executionWorkDir = attemptTree.Root
		executionRunner, err = runner.WithWorkDir(executionWorkDir)
		if err != nil {
			return errors.Join(err, lifecycle.retain(context.Background(), "runner_rebind_failure"))
		}
		attemptSnapshot, queryErr := executionRunner.Query(ctx)
		if queryErr != nil {
			return errors.Join(fmt.Errorf("%w: %v", errPreDispatchAttemptQuery, queryErr), lifecycle.retain(context.Background(), "pre_dispatch_query_failure"))
		}
		if err := requireMatchingAttemptDispatch(before, attemptSnapshot); err != nil {
			return errors.Join(err, lifecycle.retain(context.Background(), "pre_dispatch_identity_mismatch"))
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
	var observedIdentityErr error
	observer := gsd.Observer{
		Event: func(event gsd.Event) {
			if identityErr := recordObservedRuntimeIdentity(event, expectedModel, "high", &observedModel, &observedThinking); identityErr != nil && observedIdentityErr == nil {
				observedIdentityErr = identityErr
				cancel()
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
					if err := appendAndRequestDecisionSummary(questionCtx, decisions, effects, deliveryID, executionID, trustedUnit, question, response, decisionActor, decisionBasis); err != nil {
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
			if err := appendAndRequestDecisionSummary(questionCtx, decisions, effects, deliveryID, executionID, trustedUnit, question, response, decisionActor, decisionBasis); err != nil {
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
	sessionsDir := filepath.Join(config.GSDHome, "agent", "sessions")
	if config.Runtime == "podman" {
		sessionsDir = filepath.Join(config.StateDir, "runtime", "sessions")
	}
	var sessionBaseline gsd.SessionIdentityBaseline
	if executionRunner.RequiresSessionBinding() {
		sessionBaseline, err = gsd.CaptureSessionIdentityBaseline(sessionsDir, executionWorkDir)
		if err != nil {
			return fmt.Errorf("capture top-level session baseline: %w", err)
		}
	}
	var scopeMu sync.Mutex
	var scopeViolation error
	var scopeMonitorDone chan struct{}
	var stopScopeMonitor context.CancelFunc
	if canonicalUnit {
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
	if lifecycle != nil {
		if err := lifecycle.transition(ctx, store.AttemptWorktreeRunning, store.AttemptWorktreeUpdate{}); err != nil {
			if stopScopeMonitor != nil {
				stopScopeMonitor()
			}
			if scopeMonitorDone != nil {
				<-scopeMonitorDone
			}
			return err
		}
	}
	result := executionRunner.Run(runCtx, command, args, observer)
	if observedIdentityErr != nil {
		result.Terminal = gsd.TerminalError
		result.Err = joinTerminalFailure(observedIdentityErr, result.Err)
	}
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
		result.Err = joinTerminalFailure(restoreErr, result.Err)
	}
	var sessionEvidence gsd.SessionIdentityEvidence
	if executionRunner.RequiresSessionBinding() || observedModel == "" || observedThinking == "" {
		var model, thinking string
		var identityErr error
		if executionRunner.RequiresSessionBinding() {
			sessionEvidence, identityErr = gsd.ReadSessionIdentityEvidenceForRun(sessionsDir, executionWorkDir, sessionBaseline, result.Started, expectedModel, "high")
			model, thinking = sessionEvidence.Model, sessionEvidence.Thinking
		} else {
			model, thinking, identityErr = gsd.ReadSessionIdentitySince(sessionsDir, executionWorkDir, result.Started)
		}
		if identityErr != nil {
			result.Terminal = gsd.TerminalError
			result.Err = joinTerminalFailure(fmt.Errorf("current top-level session identity: %w", identityErr), result.Err)
		} else if (model != "" && model != expectedModel) || (thinking != "" && thinking != "high") {
			result.Terminal = gsd.TerminalError
			result.Err = joinTerminalFailure(fmt.Errorf("current top-level session identity is %s/%s, expected %s/high", model, thinking, expectedModel), result.Err)
		} else {
			if observedModel == "" {
				observedModel = model
			} else if model != "" && observedModel != model {
				result.Terminal = gsd.TerminalError
				result.Err = joinTerminalFailure(errors.New("live and session model identity differ"), result.Err)
			}
			if observedThinking == "" {
				observedThinking = thinking
			} else if thinking != "" && observedThinking != thinking {
				result.Terminal = gsd.TerminalError
				result.Err = joinTerminalFailure(errors.New("live and session thinking identity differ"), result.Err)
			}
			appendActivity("model.activity", "session_metadata", trustedUnit, model, "", time.Now().UTC())
		}
	}
	if canonicalUnit && executionRunner.RequiresSessionBinding() && result.Terminal == gsd.TerminalSuccess {
		if observedModel == "" || observedThinking == "" || sessionEvidence.SessionID == "" || sessionEvidence.Fingerprint == "" || sessionEvidence.Model != observedModel || sessionEvidence.Thinking != observedThinking {
			result.Terminal = gsd.TerminalError
			result.Err = joinTerminalFailure(errors.New("complete current-run unit identity evidence is required"), result.Err)
		} else if completionErr := gsd.ValidateSessionSuccessfulCompletion(sessionsDir,
			sessionEvidence.SessionID); completionErr != nil {
			result.Terminal = gsd.TerminalError
			result.Err = joinTerminalFailure(fmt.Errorf("current top-level session completion: %w", completionErr), result.Err)
		} else if identityErr := authority.RecordUnitAttemptIdentity(ctx, store.UnitAttemptIdentity{
			UnitAttemptKey: unitAttemptKey, Attempt: unitAttempt.Attempts, Model: observedModel, Thinking: observedThinking,
			SessionID: sessionEvidence.SessionID, SessionFingerprint: sessionEvidence.Fingerprint,
			StartedAt: result.Started, EndedAt: result.Ended,
		}); identityErr != nil {
			result.Terminal = gsd.TerminalError
			result.Err = joinTerminalFailure(fmt.Errorf("persist unit attempt runtime identity: %w", identityErr), result.Err)
		}
	}
	postPhase := ""
	activityMu.Lock()
	spoolErr := activityErr
	activityMu.Unlock()
	if spoolErr != nil {
		result.Terminal = gsd.TerminalError
		result.Err = joinTerminalFailure(fmt.Errorf("durable activity append failed: %w", spoolErr), result.Err)
	}
	if result.Terminal != gsd.TerminalRejected {
		snapshot, queryErr := executionRunner.Query(ctx)
		if queryErr != nil {
			result.Terminal = gsd.TerminalError
			result.Err = joinTerminalFailure(fmt.Errorf("post-run query reconciliation failed: %w", queryErr), result.Err)
		} else {
			postPhase = snapshot.Phase
			if snapshot.MilestoneID != "" {
				if bindErr := authority.BindMilestone(ctx, deliveryID, snapshot.MilestoneID); bindErr != nil {
					result.Terminal = gsd.TerminalError
					result.Err = joinTerminalFailure(bindErr, result.Err)
				}
			}
			terminal, reconcileErr := gsd.Reconcile(registry, command, result, before, snapshot)
			result.Terminal = terminal
			if reconcileErr != nil {
				result.Err = joinTerminalFailure(reconcileErr, result.Err)
			}
		}
	}
	if result.Terminal == gsd.TerminalSuccess && (observedModel != expectedModel || observedThinking != "high") {
		result.Terminal = gsd.TerminalError
		result.Err = fmt.Errorf("effective runtime identity was not observed as %s/high", expectedModel)
	}
	if err := authority.CheckLease(ctx, lease, time.Now().UTC()); err != nil {
		result.Terminal = gsd.TerminalError
		result.Err = joinTerminalFailure(err, result.Err)
	}
	candidateWorkDir := config.WorkDir
	candidateHead := ""
	if result.Terminal == gsd.TerminalSuccess && canonicalUnit {
		message := "chore(gsd): checkpoint " + strings.ReplaceAll(before.Next.UnitID, "/", " ")
		if attemptWorktree != nil {
			if attemptManager == nil {
				result.Terminal = gsd.TerminalError
				result.Err = errors.New("attempt manager missing for candidate unit")
			} else if lockErr := repositoryLock.Check(); lockErr != nil {
				result.Terminal = gsd.TerminalError
				result.Err = lockErr
			} else if head, checkpointErr := attemptManager.CheckpointCandidate(ctx, *attemptWorktree, issueContext.WriteScope, message); checkpointErr != nil {
				result.Terminal = gsd.TerminalError
				result.Err = fmt.Errorf("checkpoint candidate canonical unit: %w", checkpointErr)
			} else if head == startSnapshot.HeadSHA {
				result.Terminal = gsd.TerminalError
				result.Err = recovery.MarkFailure(recovery.FailureArtifactMissing, true,
					errors.New("successful canonical unit did not advance the candidate head"))
			} else if lifecycle == nil {
				result.Terminal = gsd.TerminalError
				result.Err = errors.New("durable attempt lifecycle missing at candidate checkpoint")
			} else if _, persistErr := authority.RecordAttemptWorktreeCandidate(ctx, lifecycle.key, lifecycle.owner, lifecycle.epoch, head); persistErr != nil {
				result.Terminal = gsd.TerminalError
				result.Err = fmt.Errorf("persist attempt candidate: %w", persistErr)
			} else {
				candidateHead = head
				candidateWorkDir = attemptWorktree.Root
				appendActivity("transition", "checkpointed_attempt_candidate", trustedUnit, "", "", time.Now().UTC())
			}
		} else if head, checkpointErr := shepherdgit.CheckpointWithinScopes(ctx, config.WorkDir, issueContext.WriteScope, message); checkpointErr != nil {
			result.Terminal = gsd.TerminalError
			result.Err = fmt.Errorf("checkpoint successful canonical unit: %w", checkpointErr)
		} else if head == startSnapshot.HeadSHA {
			result.Terminal = gsd.TerminalError
			result.Err = recovery.MarkFailure(recovery.FailureArtifactMissing, true,
				errors.New("successful canonical unit did not advance the candidate head"))
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
	} else if result.Terminal == gsd.TerminalSuccess && canonicalUnit {
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
			} else if err := persistSuccessProof(ctx, authority, validator, config, registry, issueContext.WriteScope, candidateWorkDir, deliveryID, trustedUnit, before.Next.UnitType, issueContext.PRBase, attempt.Generation, unitAttempt.Attempts, stateVersion, startSnapshot.HeadSHA, candidateHead, result.ObservedWorkflowTools, lifecycle); err != nil {
				result.Terminal = gsd.TerminalError
				result.Err = err
			} else if attemptWorktree != nil {
				if lifecycle == nil {
					result.Terminal = gsd.TerminalError
					result.Err = errors.New("durable attempt lifecycle missing before promotion")
				} else if promotionSnapshot, inspectErr := shepherdgit.Inspect(ctx, config.WorkDir); inspectErr != nil {
					result.Terminal = gsd.TerminalError
					result.Err = inspectErr
				} else if promotionSnapshot.Branch != delivery.Branch || promotionSnapshot.HeadSHA != startSnapshot.HeadSHA {
					result.Terminal = gsd.TerminalError
					result.Err = errors.New("authorized delivery branch or head changed before promotion")
				} else if cleanErr := shepherdgit.RequireClean(promotionSnapshot); cleanErr != nil {
					result.Terminal = gsd.TerminalError
					result.Err = cleanErr
				} else if err := authority.CheckLease(ctx, lease, time.Now().UTC()); err != nil {
					result.Terminal = gsd.TerminalError
					result.Err = err
				} else if err := repositoryLock.Check(); err != nil {
					result.Terminal = gsd.TerminalError
					result.Err = err
				} else if err := promoteRatifiedAttempt(ctx, authority, attemptManager, repositoryLock, lease, delivery, *attemptWorktree, lifecycle); err != nil {
					result.Terminal = gsd.TerminalError
					result.Err = fmt.Errorf("promote ratified candidate: %w", err)
				} else if err := lifecycle.transition(ctx, store.AttemptWorktreeCleanupPending, store.AttemptWorktreeUpdate{}); err != nil {
					result.Terminal = gsd.TerminalError
					result.Err = err
				} else if lockErr := repositoryLock.Check(); lockErr != nil {
					result.Terminal = gsd.TerminalError
					result.Err = lockErr
				} else if cleanupResult, cleanupErr := reconcileOwnedAttempt(ctx, attemptManager, workspace.OwnedAttempt{Attempt: *attemptWorktree, ExpectedHead: candidateHead}); cleanupErr != nil || cleanupResult != workspace.ReconcileComplete {
					result.Terminal = gsd.TerminalError
					result.Err = fmt.Errorf("cleanup promoted attempt: %w", cleanupErr)
					if transitionErr := lifecycle.transition(ctx, store.AttemptWorktreeCleanupBlocked, store.AttemptWorktreeUpdate{CleanupError: "owned_resource_mismatch_or_cleanup_failure"}); transitionErr != nil {
						result.Err = errors.Join(result.Err, transitionErr)
					}
				} else if err := lifecycle.transition(ctx, store.AttemptWorktreeCleanupComplete, store.AttemptWorktreeUpdate{}); err != nil {
					result.Terminal = gsd.TerminalError
					result.Err = err
				} else {
					appendActivity("transition", "promoted_ratified_attempt", trustedUnit, "", "", time.Now().UTC())
				}
			}
		}
	}
	endSnapshot, snapshotErr = shepherdgit.Inspect(ctx, config.WorkDir)
	if snapshotErr != nil {
		result.Terminal = gsd.TerminalError
		result.Err = joinTerminalFailure(result.Err, errors.New("inspect worktree after validation and promotion"))
	} else if endSnapshot.Branch != delivery.Branch {
		result.Terminal = gsd.TerminalError
		result.Err = joinTerminalFailure(result.Err, errors.New("canonical worktree left the authorized delivery branch"))
	} else if cleanErr := shepherdgit.RequireClean(endSnapshot); cleanErr != nil {
		result.Terminal = gsd.TerminalError
		result.Err = joinTerminalFailure(result.Err, recovery.MarkFailure(recovery.FailureDirtyTree, false,
			errors.New("worktree must be clean after validation and promotion")))
	} else if err := authority.RecordAttemptHeads(ctx, deliveryID, executionID, startSnapshot.HeadSHA, endSnapshot.HeadSHA); err != nil {
		result.Terminal = gsd.TerminalError
		result.Err = joinTerminalFailure(err, result.Err)
	}
	appendActivity("run.terminal", string(result.Terminal), "", "", "", result.Ended)
	activityMu.Lock()
	terminalSpoolErr := activityErr
	activityMu.Unlock()
	if terminalSpoolErr != nil {
		result.Terminal = gsd.TerminalError
		result.Err = joinTerminalFailure(fmt.Errorf("durable terminal append failed: %w", terminalSpoolErr), result.Err)
	}
	if result.Terminal != gsd.TerminalSuccess {
		terminalErr := result.Err
		if terminalErr == nil {
			terminalErr = errors.New("runtime did not provide an error")
		}
		if diagnostic := terminalDiagnostic(result.Stderr); diagnostic != "" {
			terminalErr = fmt.Errorf("%w; runtime: %s", terminalErr, diagnostic)
		}
		result.Err = terminalErr
	}
	if lifecycle != nil && result.Terminal != gsd.TerminalSuccess {
		if retainErr := lifecycle.retain(ctx, classifyUnitFailure(result)); retainErr != nil {
			result.Err = errors.Join(result.Err, retainErr)
		}
	}
	unitOutcome := string(result.Terminal)
	if result.Err != nil {
		unitOutcome = classifyUnitFailure(result)
	}
	if errors.Is(result.Err, gsd.ErrMutatingSkip) {
		unitOutcome = "mutating_skip"
	}
	if err := authority.FinishUnitAttemptFenced(ctx, unitAttemptKey, unitAttempt.Attempts, unitOutcome, executionID, lease.Epoch); err != nil {
		finishedAttempt = true
		return joinTerminalFailure(result.Err, err)
	} else {
		unitAttemptFinished = true
	}
	selectedRecoveryAction := recovery.Action("")
	if result.Terminal != gsd.TerminalSuccess && result.Err != nil && !errors.Is(result.Err, gsd.ErrMutatingSkip) {
		failure := recovery.Classify(result)
		if !canonicalUnit {
			if policy, policyErr := recovery.PolicyFor(failure); policyErr == nil && policy.PlannerEligible {
				failure = recovery.Failure{Class: recovery.FailureHumanRequired}
			}
		}
		decision, recoveryErr := governRecovery(ctx, recoveryGovernanceRequest{
			Store: authority, Config: config, Issue: issue, DeliveryID: deliveryID,
			Generation: attempt.Generation, UnitID: trustedUnit, UnitAttempt: unitAttempt.Attempts,
			HeadSHA: startSnapshot.HeadSHA, ExecutionID: executionID, Failure: failure,
			TerminalError: result.Err, WriteScope: issueContext.WriteScope, AuthorizedBranch: delivery.Branch, Lease: lease,
			Observe: func(status, model, action string) {
				appendActivity("recovery", status, trustedUnit, model, action, time.Now().UTC())
			},
		})
		selectedRecoveryAction = decision.Action
		if recoveryErr != nil {
			result.Err = joinTerminalFailure(result.Err, recoveryErr)
		}
		result.Err = recovery.MarkDecision(failure, selectedRecoveryAction, result.Err)
	}
	targetState := recoveryRunState(&result, postPhase, selectedRecoveryAction)
	if targetState == domain.RunAwaitingDecision {
		class := classifyUnitFailure(result)
		decisionHead := endSnapshot.HeadSHA
		questionEffects := effects
		var publicationErr error
		if len(decisionHead) != 40 {
			publicationErr = errors.New("decision publication requires the current canonical head")
		} else if decisionHead != effects.facts.HeadSHA {
			questionEffects, publicationErr = openExternalEffectController(ctx, authority, lease, config,
				deliveryID, issue, attempt.Generation, decisionHead)
			if publicationErr == nil {
				defer func() { returnErr = errors.Join(returnErr, questionEffects.Close()) }()
			}
		}
		if publicationErr == nil {
			publicationErr = persistAndPublishDecisionRequest(ctx, authority, questionEffects, deliveryID,
				issue, trustedUnit, attempt.Generation, decisionHead, class, result.Err)
		}
		if publicationErr != nil {
			result.Terminal = gsd.TerminalError
			publicationFailure := recovery.Classify(gsd.Result{Terminal: gsd.TerminalError, Err: publicationErr})
			result.Err = joinTerminalFailure(result.Err, publicationErr)
			if publicationFailure.Class == recovery.FailureGitHubPublishUncertain || publicationFailure.Class == recovery.FailureOutboxUncertain {
				result.Err = recovery.MarkDecision(publicationFailure, recovery.ActionBlock, result.Err)
				targetState = domain.RunBlocked
			} else {
				targetState = domain.RunFailed
			}
		}
	}
	if targetState == domain.RunHumanGate {
		gateErr := validateFinalGateGSDState(ctx, authority, config.StateDir, config.WorkDir,
			deliveryID, attempt.Generation, endSnapshot.HeadSHA)
		if gateErr == nil {
			gateErr = authority.ResolveRecoveredPromotionDecisions(ctx, lease, time.Now().UTC())
		}
		if gateErr == nil {
			gateErr = authority.ProjectFinalHumanGate(ctx, lease, attempt.Generation,
				endSnapshot.HeadSHA, time.Now().UTC())
		}
		if gateErr != nil {
			result.Terminal = gsd.TerminalError
			result.Err = joinTerminalFailure(result.Err, gateErr)
			targetState = domain.RunFailed
		} else {
			integrationFinalGateBoundary()
			finishedAttempt = true
		}
	}
	if !finishedAttempt {
		if err := authority.FinishAttempt(ctx, deliveryID, executionID, targetState); err != nil {
			result.Terminal = gsd.TerminalError
			result.Err = joinTerminalFailure(err, result.Err)
		} else {
			finishedAttempt = true
		}
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
					results <- recovery.MarkFailure(recovery.FailureWriteScopeBreach, false,
						fmt.Errorf("live write-scope breach: changed path %q is outside the issue write scope", paths[0]))
					return
				}
			}
		}
	}()
	return results
}

func recordObservedRuntimeIdentity(event gsd.Event, expectedModel, expectedThinking string, model, thinking *string) error {
	if !event.IsTopLevelIdentity() {
		return nil
	}
	if event.Model != "" {
		if event.Model != expectedModel || (*model != "" && *model != event.Model) {
			return recovery.MarkFailure(recovery.FailureModelMismatch, false,
				fmt.Errorf("unexpected top-level model transition to %s; expected %s", event.Model, expectedModel))
		}
		*model = event.Model
	}
	if event.Kind == gsd.EventThinkingSelect && event.Thinking != "" {
		if event.Thinking != expectedThinking || (*thinking != "" && *thinking != event.Thinking) {
			return recovery.MarkFailure(recovery.FailureThinkingMismatch, false,
				fmt.Errorf("unexpected top-level thinking transition to %s; expected %s", event.Thinking, expectedThinking))
		}
		*thinking = event.Thinking
	}
	return nil
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

func rejectUnqualifiedContinuation(requested bool) error {
	if requested {
		return errors.New("--continue-unit is unavailable for disposable governed units until prior-attempt session binding is qualified")
	}
	return nil
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

func classifyUnitFailure(result gsd.Result) string {
	return string(recovery.Classify(result).Class)
}

func isAutomaticallyRetryable(err error) bool {
	return recovery.ShouldRetry(err)
}

func recoveryRunState(result *gsd.Result, postPhase string, action recovery.Action) domain.RunState {
	switch action {
	case recovery.ActionRetrySameUnit, recovery.ActionRetryAfterBackoff, recovery.ActionRunRecoveryPlan:
		return domain.RunReady
	case recovery.ActionAwaitDecision:
		return domain.RunAwaitingDecision
	case recovery.ActionBlock:
		return domain.RunBlocked
	case recovery.ActionFinalHumanGate:
		return domain.RunHumanGate
	case "":
		return targetRunState(result.Terminal, result.Err, postPhase)
	default:
		result.Err = joinTerminalFailure(result.Err, errors.New("unknown selected recovery action"))
		return domain.RunBlocked
	}
}

type recoveryGovernanceRequest struct {
	Store            *store.Store
	Config           fileConfig
	Issue            int
	DeliveryID       string
	Generation       int64
	UnitID           string
	UnitAttempt      int64
	HeadSHA          string
	ExecutionID      string
	Failure          recovery.Failure
	TerminalError    error
	WriteScope       []string
	AuthorizedBranch string
	Lease            store.Lease
	Observe          func(status, model, action string)
}

type recoveryGovernanceDecision struct {
	Action      recovery.Action
	NextRetryAt time.Time
}

func governRecovery(ctx context.Context, request recoveryGovernanceRequest) (recoveryGovernanceDecision, error) {
	if err := request.Store.CheckLease(ctx, request.Lease, time.Now().UTC()); err != nil {
		return recoveryGovernanceDecision{Action: recovery.ActionBlock}, err
	}
	policy, err := recovery.PolicyFor(request.Failure)
	if err != nil {
		return recoveryGovernanceDecision{Action: recovery.ActionBlock}, err
	}
	now := time.Now().UTC()
	key := store.RecoveryBudgetKey{
		DeliveryID:   request.DeliveryID,
		Generation:   request.Generation,
		UnitID:       request.UnitID,
		HeadSHA:      request.HeadSHA,
		FailureClass: string(request.Failure.Class),
	}
	maxAttempts := int64(request.Config.MaxUnitAttempts)
	failureHash := sha256.Sum256([]byte(request.TerminalError.Error()))
	reservation := store.RecoveryReservation{
		RecoveryBudgetKey: key,
		UnitAttempt:       request.UnitAttempt,
		ClaimToken:        recoveryClaimToken(request),
		ControllerOwner:   request.ExecutionID,
		ControllerEpoch:   request.Lease.Epoch,
		PolicyVersion:     recovery.PolicyVersion,
		MaxAttempts:       maxAttempts,
		BaseBackoff:       time.Second,
		MaxBackoff:        30 * time.Second,
		FailureHash:       "sha256:" + hex.EncodeToString(failureHash[:]),
		Diagnostic:        string(request.Failure.Class),
		Reversible:        request.Failure.Reversible,
		ExhaustedAction:   policy.ExhaustedAction,
		Now:               now,
	}
	budget, reserveErr := request.Store.ReserveRecoveryAttempt(ctx, reservation)
	if errors.Is(reserveErr, store.ErrRetryBudgetExhausted) {
		request.observe("budget_exhausted", "", string(budget.SelectedAction))
		return recoveryGovernanceDecision{Action: budget.SelectedAction}, reserveErr
	}
	if reserveErr != nil {
		request.observe("budget_rejected", "", string(recovery.ActionBlock))
		return recoveryGovernanceDecision{Action: recovery.ActionBlock}, reserveErr
	}
	request.observe("budget_reserved", "", string(request.Failure.Class))
	if !policy.PlannerEligible {
		if err := request.Store.CheckLease(ctx, request.Lease, time.Now().UTC()); err != nil {
			return recoveryGovernanceDecision{Action: recovery.ActionBlock}, err
		}
		if err := request.Store.CompleteRecoveryDecision(ctx, key, request.UnitAttempt, reservation.ClaimToken, request.ExecutionID, request.Lease.Epoch, policy.DirectAction, now); err != nil {
			return recoveryGovernanceDecision{Action: recovery.ActionBlock}, err
		}
		request.observe("action_selected", "", string(policy.DirectAction))
		return recoveryGovernanceDecision{Action: policy.DirectAction}, nil
	}
	beforeGit, beforeGSD, err := canonicalRecoveryFingerprint(ctx, request.Config.WorkDir)
	if err != nil {
		_ = request.Store.RejectRecoveryAttempt(ctx, key, request.UnitAttempt, reservation.ClaimToken, request.ExecutionID, request.Lease.Epoch, recovery.ActionBlock, "canonical fingerprint unavailable", now)
		return recoveryGovernanceDecision{Action: recovery.ActionBlock}, err
	}
	planner, err := recoveryPlannerFactory(request.Config)
	if err != nil {
		_ = request.Store.RejectRecoveryAttempt(ctx, key, request.UnitAttempt, reservation.ClaimToken, request.ExecutionID, request.Lease.Epoch, recovery.ActionBlock, "planner unavailable", now)
		return recoveryGovernanceDecision{Action: recovery.ActionBlock}, err
	}
	authorityHash, err := recoveryAuthorityScopeHash(ctx, request)
	if err != nil {
		rejectErr := request.Store.RejectRecoveryAttempt(ctx, key, request.UnitAttempt, reservation.ClaimToken, request.ExecutionID, request.Lease.Epoch, recovery.ActionBlock, "authority scope unavailable", now)
		return recoveryGovernanceDecision{Action: recovery.ActionBlock}, errors.Join(err, rejectErr)
	}
	request.observe("planner_requested", recovery.RequiredModel, string(request.Failure.Class))
	planned, planErr := planner.Plan(ctx, recovery.Request{
		Issue:              request.Issue,
		DeliveryID:         request.DeliveryID,
		Generation:         request.Generation,
		UnitID:             request.UnitID,
		Attempt:            request.UnitAttempt,
		HeadSHA:            request.HeadSHA,
		Failure:            request.Failure,
		EvidenceHash:       reservation.FailureHash,
		AuthorityScopeHash: authorityHash,
		ControllerBackoff:  budget.Backoff,
	})
	if planErr == nil {
		planErr = recovery.ValidateResult(recovery.Request{
			Issue: request.Issue, DeliveryID: request.DeliveryID, Generation: request.Generation,
			UnitID: request.UnitID, Attempt: request.UnitAttempt, HeadSHA: request.HeadSHA,
			Failure: request.Failure, EvidenceHash: reservation.FailureHash,
			AuthorityScopeHash: authorityHash, ControllerBackoff: budget.Backoff,
		}, planned, time.Now().UTC())
	}
	if planErr != nil {
		rejectErr := request.Store.RejectRecoveryAttempt(ctx, key, request.UnitAttempt, reservation.ClaimToken, request.ExecutionID, request.Lease.Epoch, recovery.ActionBlock, "planner evidence rejected", time.Now().UTC())
		request.observe("planner_rejected", recovery.RequiredModel, string(recovery.ActionBlock))
		return recoveryGovernanceDecision{Action: recovery.ActionBlock}, errors.Join(planErr, rejectErr)
	}
	afterGit, afterGSD, fingerprintErr := canonicalRecoveryFingerprint(ctx, request.Config.WorkDir)
	if fingerprintErr != nil || beforeGit != afterGit || beforeGSD != afterGSD {
		rejectErr := request.Store.RejectRecoveryAttempt(ctx, key, request.UnitAttempt, reservation.ClaimToken, request.ExecutionID, request.Lease.Epoch, recovery.ActionBlock, "canonical state changed during planning", time.Now().UTC())
		request.observe("planner_rejected", planned.ObservedModel, string(recovery.ActionBlock))
		return recoveryGovernanceDecision{Action: recovery.ActionBlock}, errors.Join(errors.New("canonical Git or GSD state changed during recovery planning"), fingerprintErr, rejectErr)
	}
	if err := request.Store.CheckLease(ctx, request.Lease, time.Now().UTC()); err != nil {
		return recoveryGovernanceDecision{Action: recovery.ActionBlock}, err
	}
	outcome := store.RecoveryOutcome{
		RecoveryBudgetKey:         key,
		UnitAttempt:               request.UnitAttempt,
		ClaimToken:                reservation.ClaimToken,
		ControllerOwner:           request.ExecutionID,
		ControllerEpoch:           request.Lease.Epoch,
		PlannerRequestNonce:       planned.RequestNonce,
		EvidenceHash:              planned.EvidenceHash,
		AuthorityScopeHash:        planned.AuthorityScopeHash,
		PlannerEvidenceHash:       planned.PlannerEvidenceHash,
		PlannerSessionID:          planned.SessionID,
		PlannerSessionFingerprint: planned.SessionFingerprint,
		ObservedModel:             planned.ObservedModel,
		Thinking:                  planned.Thinking,
		SelectedAction:            planned.Action,
		BoundedPlan:               planned.BoundedPlanSteps,
		IssuedAt:                  planned.IssuedAt,
		ExpiresAt:                 planned.ExpiresAt,
	}
	if err := request.Store.CompleteRecoveryAttempt(ctx, outcome); err != nil {
		return recoveryGovernanceDecision{Action: recovery.ActionBlock}, err
	}
	request.observe("planner_result", planned.ObservedModel, string(planned.Action))
	request.observe("action_selected", planned.ObservedModel, string(planned.Action))
	if recovery.ShouldRetry(recovery.MarkDecision(request.Failure, planned.Action, request.TerminalError)) {
		request.observe("backoff_scheduled", planned.ObservedModel, string(planned.Action))
	}
	return recoveryGovernanceDecision{Action: planned.Action, NextRetryAt: budget.NextRetryAt}, nil
}

func (r recoveryGovernanceRequest) observe(status, model, action string) {
	if r.Observe != nil {
		r.Observe(status, model, action)
	}
}

func totalUnitAttemptLimit(perClassMax int) int64 {
	if perClassMax <= 0 {
		return 1
	}
	return 1 + int64(len(recovery.AllFailureClasses())*perClassMax)
}

func recoveryClaimToken(request recoveryGovernanceRequest) string {
	digest := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%d\x00%s\x00%d\x00%s", request.DeliveryID, request.Generation, request.UnitID, request.UnitAttempt, request.ExecutionID)))
	return hex.EncodeToString(digest[:16])
}

func recoveryAuthorityScopeHash(ctx context.Context, request recoveryGovernanceRequest) (string, error) {
	snapshot, err := shepherdgit.Inspect(ctx, request.Config.WorkDir)
	if err != nil {
		return "", err
	}
	if snapshot.HeadSHA != request.HeadSHA {
		return "", recovery.MarkFailure(recovery.FailureStaleHead, false, errors.New("canonical head changed before recovery authority binding"))
	}
	if snapshot.Branch != request.AuthorizedBranch {
		return "", recovery.MarkFailure(recovery.FailureStaleHead, false, errors.New("canonical branch differs from authorized recovery branch"))
	}
	scopes := append([]string(nil), request.WriteScope...)
	sort.Strings(scopes)
	gsdManifest, err := workspace.BuildGSDManifest(ctx, filepath.Join(request.Config.WorkDir, ".gsd"))
	if err != nil {
		return "", err
	}
	raw, err := json.Marshal(struct {
		DeliveryID string   `json:"delivery_id"`
		Generation int64    `json:"generation"`
		UnitID     string   `json:"unit_id"`
		HeadSHA    string   `json:"head_sha"`
		Branch     string   `json:"branch"`
		WriteScope []string `json:"write_scope"`
		GSDHash    string   `json:"gsd_hash"`
	}{
		DeliveryID: request.DeliveryID,
		Generation: request.Generation,
		UnitID:     request.UnitID,
		HeadSHA:    request.HeadSHA,
		Branch:     request.AuthorizedBranch,
		WriteScope: scopes,
		GSDHash:    gsdManifest.Hash,
	})
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:]), nil
}

func canonicalRecoveryFingerprint(ctx context.Context, workDir string) (string, string, error) {
	snapshot, err := shepherdgit.Inspect(ctx, workDir)
	if err != nil {
		return "", "", err
	}
	if err := shepherdgit.RequireClean(snapshot); err != nil {
		return "", "", recovery.MarkFailure(recovery.FailureDirtyTree, false, err)
	}
	manifest, err := workspace.BuildGSDManifest(ctx, filepath.Join(workDir, ".gsd"))
	if err != nil {
		return "", "", err
	}
	bound := sha256.Sum256([]byte(snapshot.Branch + "\x00" + manifest.Hash))
	return snapshot.HeadSHA, "sha256:" + hex.EncodeToString(bound[:]), nil
}

func waitForRecoveryDispatch(ctx context.Context, authority *store.Store, lease store.Lease, deliveryID string, generation int64, unitID, headSHA, executionID string) (store.RecoveryAttempt, error) {
	for {
		if err := authority.CheckLease(ctx, lease, time.Now().UTC()); err != nil {
			return store.RecoveryAttempt{}, err
		}
		attempt, err := authority.ClaimRecoveryDispatch(ctx, deliveryID, generation, unitID, headSHA, executionID, lease.Epoch, time.Now().UTC())
		if err == nil {
			return attempt, nil
		}
		if !errors.Is(err, store.ErrRecoveryBackoffPending) {
			return store.RecoveryAttempt{}, err
		}
		wait := time.Until(attempt.NextRetryAt)
		if wait <= 0 {
			continue
		}
		timer := time.NewTimer(wait)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return store.RecoveryAttempt{}, ctx.Err()
		case <-timer.C:
		}
	}
}

func executeRecoveryPlan(ctx context.Context, attempt store.RecoveryAttempt, authority *store.Store, manager *workspace.Manager, repositoryLock *workspace.RepositoryLock, lease store.Lease, config fileConfig) error {
	if attempt.SelectedAction == "" {
		return nil
	}
	if err := recovery.ValidatePlan(attempt.SelectedAction, attempt.BoundedPlan); err != nil {
		return err
	}
	if attempt.SelectedAction != recovery.ActionRunRecoveryPlan {
		return nil
	}
	for _, step := range attempt.BoundedPlan {
		switch step.Primitive {
		case recovery.PrimitiveInspectRetainedAttempt:
			records, err := authority.ListAttemptWorktrees(ctx, attempt.DeliveryID)
			if err != nil {
				return err
			}
			if attemptRecoveryAmbiguous(records) {
				return errors.New("retained recovery attempt ownership is ambiguous")
			}
		case recovery.PrimitiveReconcileAttemptResources:
			if err := reconcileAttemptWorktrees(ctx, authority, manager, repositoryLock, attempt.DeliveryID, lease); err != nil {
				return err
			}
		case recovery.PrimitiveVerifyExpectedArtifacts:
			head, _, err := canonicalRecoveryFingerprint(ctx, config.WorkDir)
			if err != nil {
				return err
			}
			if head != attempt.HeadSHA {
				return recovery.MarkFailure(recovery.FailureStaleHead, false, errors.New("canonical head changed before recovery dispatch"))
			}
		case recovery.PrimitiveRetryFreshAttempt:
			// Fresh attempt resources are created immediately after this typed marker.
		default:
			return fmt.Errorf("unsupported recovery primitive %q", step.Primitive)
		}
	}
	return nil
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
		Repository: config.Repository, PullRequest: config.PullRequest,
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
		delivery.Repository != config.Repository || delivery.PullRequest != config.PullRequest ||
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

func appendAndRequestDecisionSummary(ctx context.Context, decisionStore *decisionlog.Store,
	effects externalEffectRequester, deliveryID, executionID, unitID string, question gsd.Question,
	response gsd.UIResponse, actorValue, basis string) error {
	if effects == nil {
		return errors.New("external effect requester is required")
	}
	if err := appendDecision(decisionStore, deliveryID, executionID, unitID, question, response, actorValue, basis); err != nil {
		return err
	}
	snapshot, err := decisionStore.Snapshot(deliveryID)
	if err != nil {
		return fmt.Errorf("read durable decision snapshot: %w", err)
	}
	if _, err := effects.RequestSummary(ctx, snapshot); err != nil {
		return recovery.MarkFailure(recovery.FailureOutboxUncertain, false,
			fmt.Errorf("request durable decision-summary effect: %w", err))
	}
	return nil
}

func recordOperationalDecision(ctx context.Context, decisionStore *decisionlog.Store,
	effects externalEffectRequester, deliveryID, unitID, question, answer, actor, basis string) error {
	if strings.TrimSpace(unitID) == "" || strings.TrimSpace(question) == "" || strings.TrimSpace(answer) == "" {
		return errors.New("decision unit, question, and answer are required")
	}
	executionID := fmt.Sprintf("decision-%d", time.Now().UTC().UnixNano())
	return appendAndRequestDecisionSummary(ctx, decisionStore, effects, deliveryID, executionID, unitID,
		gsd.Question{ID: "operational-decision", Title: question}, gsd.UIResponse{Value: answer}, actor, basis)
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
	var totalBytes int64
	var entries int
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
			entries++
			if entries > workspace.MaxGSDManifestEntries {
				return errors.New("GSD validation artifact entry limit exceeded")
			}
			if entry.IsDir() {
				return nil
			}
			if !info.Mode().IsRegular() {
				return errors.New("GSD validation artifact must be a regular file")
			}
			if info.Size() > workspace.MaxGSDManifestFileBytes || totalBytes > workspace.MaxGSDManifestTotalBytes-info.Size() {
				return errors.New("GSD validation artifact byte limit exceeded")
			}
			totalBytes += info.Size()
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

func persistSuccessProof(ctx context.Context, authorityStore *store.Store, validator validation.Validator, config fileConfig, registry gsd.UnitRegistry, scopes []string, workDir, deliveryID, unitID, unitType, baseBranch string, generation, attempt, stateVersion int64, startHead, candidateHead string, observedWorkflowTools []string, lifecycle *attemptLifecycle) error {
	metadata, ok := registry.Lookup(unitType)
	if !ok {
		return fmt.Errorf("missing official unit metadata for %s", unitType)
	}
	observedWorkflowTools = append([]string(nil), observedWorkflowTools...)
	sort.Strings(observedWorkflowTools)
	observedWorkflowTools = slices.Compact(observedWorkflowTools)
	for _, required := range metadata.RequiredWorkflowTools {
		if !slices.Contains(observedWorkflowTools, required) {
			return recovery.MarkFailure(recovery.FailureArtifactInvalid, true,
				fmt.Errorf("successful canonical unit lacks required workflow transition %s", required))
		}
	}
	artifacts, err := shepherdgit.ArtifactManifest(ctx, workDir, startHead, candidateHead, scopes)
	if err != nil {
		if errors.Is(err, shepherdgit.ErrWriteScopeBreach) {
			return recovery.MarkFailure(recovery.FailureWriteScopeBreach, false, err)
		}
		return recovery.MarkFailure(recovery.FailureArtifactInvalid, true, err)
	}
	gsdManifest, err := workspace.SnapshotGSDManifest(ctx, filepath.Join(workDir, ".gsd"), config.StateDir)
	if err != nil {
		return recovery.MarkFailure(recovery.FailureArtifactInvalid, true, err)
	}
	baseGSDManifest, err := workspace.SnapshotGSDManifest(ctx, filepath.Join(config.WorkDir, ".gsd"), config.StateDir)
	if err != nil {
		return recovery.MarkFailure(recovery.FailureArtifactInvalid, true, err)
	}
	gsdStateChanged := filepath.Clean(workDir) != filepath.Clean(config.WorkDir) &&
		gsdManifest.Hash != baseGSDManifest.Hash
	if len(artifacts) == 0 && !gsdStateChanged {
		return recovery.MarkFailure(recovery.FailureArtifactMissing, true,
			errors.New("successful canonical unit did not change a governed artifact"))
	}
	stateArtifacts, err := gsdStateArtifacts(workDir)
	if err != nil {
		return recovery.MarkFailure(recovery.FailureArtifactInvalid, true, err)
	}
	artifacts = append(artifacts, stateArtifacts...)
	if len(artifacts) == 0 {
		return recovery.MarkFailure(recovery.FailureArtifactMissing, true,
			errors.New("successful canonical unit produced no verifiable artifact"))
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
		ObservedWorkflowTools []string               `json:"observed_workflow_tools"`
		GSDManifestHash       string                 `json:"gsd_manifest_hash"`
		Artifacts             []shepherdgit.Artifact `json:"artifacts"`
	}{UnitType: unitType, PhaseChain: metadata.PhaseChain, RequiredWorkflowTools: metadata.RequiredWorkflowTools,
		ObservedWorkflowTools: observedWorkflowTools, GSDManifestHash: gsdManifest.Hash, Artifacts: artifacts}
	raw, err := json.Marshal(manifest)
	if err != nil {
		return err
	}
	evidenceHash := sha256.Sum256(raw)
	artifactHashes := make([]validation.ArtifactHash, 0, len(artifacts))
	for _, artifact := range artifacts {
		artifactHashes = append(artifactHashes, validation.ArtifactHash{Path: artifact.Path, Hash: artifact.Hash, Deleted: artifact.Deleted})
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
		reversible := errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
		return recovery.MarkFailure(recovery.FailureValidationFailed, reversible, err)
	}
	if err := integrationPostValidationBoundary(workDir); err != nil {
		return recovery.MarkFailure(recovery.FailureValidationFailed, false, err)
	}
	if validationResult.SessionID == "" {
		return recovery.MarkFailure(recovery.FailureValidationFailed, false,
			errors.New("validator session identity is required"))
	}
	if validationResult.EvidenceHash != evidenceHashValue {
		return recovery.MarkFailure(recovery.FailureValidationFailed, false,
			errors.New("validator evidence hash does not match candidate artifacts"))
	}
	issuedAt := validationResult.IssuedAt
	if issuedAt.IsZero() {
		return recovery.MarkFailure(recovery.FailureValidationFailed, false,
			errors.New("validator issue time is required"))
	}
	if lifecycle != nil {
		if err := lifecycle.transition(ctx, store.AttemptWorktreeValidated, store.AttemptWorktreeUpdate{CandidateHead: candidateHead, ValidatedHead: candidateHead}); err != nil {
			return err
		}
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
		return recovery.MarkFailure(recovery.FailureRatificationFailed, false, err)
	}
	proofIDHash := sha256.Sum256([]byte(deliveryID + ":" + unitID + ":" + candidateHead + ":" + hex.EncodeToString(evidenceHash[:])))
	proofRecord := store.ArtifactProof{
		ProofID: hex.EncodeToString(proofIDHash[:8]), DeliveryID: deliveryID, Generation: generation, UnitID: unitID, Attempt: attempt,
		StartHead: startHead, CandidateHead: candidateHead, ValidatedHead: attestation.HeadSHA,
		ExpectedArtifact: string(raw), ArtifactHash: "sha256:" + hex.EncodeToString(evidenceHash[:]),
		Validator: attestation.Validator, Thinking: attestation.Thinking,
		Ratified: attestation.HeadSHA == candidateHead && attestation.Validator == validationResult.ObservedModel,
	}
	attestationRecord := store.AttestationRecord{
		Repository: attestation.Repository, PR: attestation.PR, BaseBranch: attestation.BaseBranch,
		BaseHead: attestation.BaseSHA, CandidateHead: attestation.HeadSHA, ObservedHead: validationResult.ObservedHead,
		RunID: deliveryID, Generation: generation, UnitID: unitID, Attempt: attempt, StateVersion: stateVersion,
		ContractHash: attestation.ContractHash, EvidenceHash: attestation.EvidenceHash,
		ValidatorSessionID: attestation.ValidatorSessionID, HeadSHA: attestation.HeadSHA,
		Validator: attestation.Validator, Thinking: attestation.Thinking, Verdict: attestation.Verdict,
		LocalGates: attestation.LocalGates, UAT: attestation.UAT, MilestoneValid: attestation.MilestoneValid,
		CreatedAt: attestation.IssuedAt, ExpiresAt: attestation.ExpiresAt,
	}
	if lifecycle != nil {
		record, err := authorityStore.RatifyAttemptWorktree(ctx, lifecycle.key, lifecycle.owner, lifecycle.epoch, proofRecord, attestationRecord)
		if err != nil {
			return err
		}
		lifecycle.state = record.State
		return nil
	}
	if err := authorityStore.PutArtifactProof(ctx, proofRecord); err != nil {
		return err
	}
	return authorityStore.PutAttestation(ctx, attestationRecord)
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

type decisionReplyPoller interface {
	PollDecisionReplies(context.Context, shepherdgithub.QuestionRequest, string) ([]shepherdgithub.Reply, error)
}

func expireUnansweredDecisionRequests(ctx context.Context, authority *store.Store, lease store.Lease,
	deliveryID string, eligibilityCutoff time.Time, applyFence func() error) error {
	requests, err := authority.ListOpenDecisionRequests(ctx, deliveryID)
	if err != nil {
		return err
	}
	for _, request := range requests {
		if !request.ExpiresAt.After(eligibilityCutoff) {
			if applyFence == nil {
				return errors.New("decision expiry application fence is required")
			}
			if err := applyFence(); err != nil {
				return err
			}
			commitAt := time.Now().UTC()
			if err := authority.ExpireDecisionRequestAndBlock(ctx, lease, request.RequestID, commitAt); err != nil {
				return err
			}
		}
	}
	return nil
}

func decisionApplicationFence(ctx context.Context, repositoryLock *workspace.RepositoryLock,
	workDir, branch, headSHA string) func() error {
	return func() error {
		if err := repositoryLock.Check(); err != nil {
			return err
		}
		canonical, err := shepherdgit.Inspect(ctx, workDir)
		if err != nil {
			return err
		}
		if canonical.Branch != branch || canonical.HeadSHA != headSHA {
			return errors.New("decision reply authority changed during polling")
		}
		return nil
	}
}

func consumeGitHubDecisionReplies(ctx context.Context, authority *store.Store, lease store.Lease,
	client decisionReplyPoller, deliveryID string, generation int64, headSHA string,
	applyFence func() error) (bool, error) {
	requests, err := authority.ListDecisionAnswerRecovery(ctx, deliveryID, generation, headSHA)
	if err != nil {
		return false, err
	}
	type candidate struct {
		request store.DecisionRequest
		reply   shepherdgithub.Reply
	}
	candidates := make([]candidate, 0, len(requests))
	for _, request := range requests {
		if request.AcceptedAnswer != "" && (request.AcceptedActorID <= 0 || request.AcceptedCommentID <= 0) {
			return false, store.ErrLegacyDecisionReply
		}
		reply := shepherdgithub.Reply{RequestID: request.RequestID, Option: request.AcceptedAnswer,
			Author: request.AcceptedBy, AuthorID: request.AcceptedActorID, CommentID: request.AcceptedCommentID,
			CreatedAt: request.AcceptedAt}
		if request.AcceptedAnswer == "" {
			replies, err := client.PollDecisionReplies(ctx, shepherdgithub.QuestionRequest{
				RequestID: request.RequestID, Repository: request.Repository, Issue: request.Issue,
				PullRequest: request.PullRequest,
				DeliveryID:  deliveryID, UnitID: request.UnitID, Generation: request.Generation, HeadSHA: request.HeadSHA,
				Evidence: request.Evidence, Options: request.Options, RecommendedOption: request.RecommendedOption,
				SafeDefault: request.SafeDefault, ExpiresAt: request.ExpiresAt, Mention: "karthik-sivadas",
				QuestionCommentID: request.GitHubCommentID,
			}, "karthik-sivadas")
			if err != nil {
				return false, err
			}
			if len(replies) == 0 {
				continue
			}
			reply = replies[0]
		}
		candidates = append(candidates, candidate{request: request, reply: reply})
	}
	apply := func(candidate candidate) error {
		if applyFence == nil {
			return errors.New("decision reply application fence is required")
		}
		if err := applyFence(); err != nil {
			return err
		}
		return authority.ApplyDecisionRequestAnswer(ctx, lease, candidate.request.RequestID,
			candidate.reply.Option, candidate.reply.Author, candidate.reply.AuthorID, candidate.reply.CommentID,
			candidate.request.Generation, candidate.request.HeadSHA, candidate.reply.CreatedAt, time.Now().UTC())
	}
	for _, candidate := range candidates {
		if candidate.reply.Option != "retry" && candidate.reply.Option != "continue" {
			if err := apply(candidate); err != nil {
				return false, err
			}
			return true, nil
		}
	}
	applied := false
	for _, candidate := range candidates {
		if err := apply(candidate); err != nil {
			return applied, err
		}
		applied = true
	}
	return applied, nil
}

func persistAndPublishDecisionRequest(ctx context.Context, authority *store.Store,
	effects *externalEffectController, deliveryID string, issue int, unitID string, generation int64,
	headSHA, failureClass string, cause error) error {
	if effects == nil {
		return errors.New("external effect controller is required")
	}
	basis := failureClass
	if cause != nil {
		basis = failureClass + ": redacted local diagnostic available in activity log"
	}
	hash := sha256.Sum256([]byte(deliveryID + ":" + strconv.FormatInt(generation, 10) + ":" + unitID + ":" + headSHA + ":" + failureClass))
	requestID := "decision-" + hex.EncodeToString(hash[:8])
	expiresAt := time.Now().UTC().Add(integrationDecisionTTL())
	request := store.DecisionRequest{
		RequestID: requestID, DeliveryID: deliveryID, Repository: effects.facts.Repository,
		Issue: issue, PullRequest: effects.facts.PullRequest,
		UnitID: unitID, Generation: generation, HeadSHA: headSHA, Kind: failureClass,
		Evidence: boundedEvidence(basis), Options: []string{"retry", "stop"}, RecommendedOption: "retry",
		SafeDefault: "stop", ExpiresAt: expiresAt, Status: store.DecisionRequestOpen,
	}
	stored, err := authority.UpsertDecisionRequest(ctx, request)
	if err != nil {
		return err
	}
	result, err := effects.RequestQuestion(ctx, stored)
	if err != nil {
		return recovery.MarkFailure(recovery.FailureOutboxUncertain, false,
			fmt.Errorf("request durable decision-question effect: %w", err))
	}
	if err := authority.MarkDecisionRequestPublishedFenced(ctx, effects.lease, stored,
		result.ExternalID, time.Now().UTC()); err != nil {
		return recovery.MarkFailure(recovery.FailureOutboxUncertain, false, err)
	}
	return nil
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
