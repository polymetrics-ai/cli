package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/flow"
	"polymetrics.ai/internal/schedule"
)

// pmcert:workflow schedule
func runSchedule(ctx context.Context, cfg config.Config, root string, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return usageErrorf("usage: pm schedule <create|list|inspect|status|install|remove|fire>")
	}
	switch args[0] {
	case "create":
		return runScheduleCreate(ctx, root, args[1:], stdout, jsonOut)
	case "list":
		return runScheduleList(root, args[1:], stdout, jsonOut)
	case "inspect", "status":
		return runScheduleInspect(root, args[1:], stdout, jsonOut)
	case "install":
		return runScheduleInstall(ctx, cfg, root, args[1:], stdout, jsonOut)
	case "remove":
		return runScheduleRemove(ctx, cfg, root, args[1:], stdout, jsonOut)
	case "fire":
		return runScheduleFire(ctx, cfg, root, args[1:], stdout, jsonOut)
	default:
		return usageErrorf("unknown schedule subcommand %q", args[0])
	}
}

func runScheduleCreate(ctx context.Context, root string, args []string, stdout io.Writer, jsonOut bool) error {
	flags := parseFlags(args)
	name := flags.first("name")
	cron := flags.first("cron")
	flow := flags.first("flow")

	if name == "" || cron == "" || flow == "" {
		return usageErrorf("pm schedule create requires --name, --cron, --flow")
	}

	if _, err := schedule.ParseCron(cron); err != nil {
		return validationErrorf("invalid --cron: %v", err)
	}
	if err := validateStoredScheduledFlow(ctx, root, flow); err != nil {
		return err
	}
	if _, found, err := schedule.FindByFlow(root, flow); err != nil {
		return err
	} else if found {
		return &schedule.FlowReferenceError{Flow: flow, Reason: schedule.FlowReferenceAmbiguous, Err: errors.New("flow already has a schedule")}
	}

	now := time.Now().UTC()
	m := schedule.Manifest{
		Name: name, Cron: cron, Flow: flow, CreatedAt: now, UpdatedAt: now,
	}

	if err := schedule.Save(root, m, false); err != nil {
		if isAlreadyExists(err) {
			return validationErrorf("%v", err)
		}
		return err
	}

	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "Schedule", "ok": true, "schedule": m, "status": schedule.FireState{Status: schedule.FireStatusReady}})
	}
	_, _ = fmt.Fprintf(stdout, "Created schedule %s (cron: %s, flow: %s, status: %s)\n", m.Name, m.Cron, m.Flow, schedule.FireStatusReady)
	return nil
}

func runScheduleList(root string, args []string, stdout io.Writer, jsonOut bool) error {
	manifests, err := schedule.List(root)
	if err != nil {
		return err
	}
	if manifests == nil {
		manifests = []schedule.Manifest{}
	}

	statuses := make(map[string]schedule.FireState, len(manifests))
	for _, m := range manifests {
		state, err := schedule.LoadFireState(root, m.Name)
		if err != nil {
			return err
		}
		statuses[m.Name] = state
		if jsonOut {
			continue
		}
		_, _ = fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\n", m.Name, m.Cron, m.Flow, state.Status)
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ScheduleList", "schedules": manifests, "statuses": statuses})
	}
	return nil
}

func runScheduleInspect(root string, args []string, stdout io.Writer, jsonOut bool) error {
	flags := parseFlags(args)
	positionals := flags.values["_"]
	if len(positionals) != 1 {
		return usageErrorf("pm schedule inspect <name>")
	}
	m, err := schedule.Load(root, positionals[0])
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return validationErrorf("schedule %q not found", positionals[0])
		}
		return err
	}
	state, err := schedule.LoadFireState(root, m.Name)
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ScheduleInspect", "schedule": m, "status": state})
	}
	_, _ = fmt.Fprintf(stdout, "Schedule: %s\nFlow: %s\nCron: %s\nStatus: %s\n", m.Name, m.Flow, m.Cron, state.Status)
	if len(state.LastFire.ReceiptIDs) > 0 {
		_, _ = fmt.Fprintf(stdout, "Receipt IDs: %s\n", strings.Join(state.LastFire.ReceiptIDs, ","))
	}
	if len(state.LastFire.PreparedExecutionIdentities) > 0 {
		_, _ = fmt.Fprintf(stdout, "Prepared executions: %s\n", strings.Join(state.LastFire.PreparedExecutionIdentities, ","))
	}
	return nil
}

func runScheduleInstall(ctx context.Context, cfg config.Config, root string, args []string, stdout io.Writer, jsonOut bool) error {
	flags := parseFlags(args)
	positionals := flags.values["_"]
	if len(positionals) == 0 {
		return usageErrorf("pm schedule install <name> [--crontab]")
	}
	name := positionals[0]
	forceCrontab := flags.first("crontab") == "true"

	m, err := schedule.Load(root, name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return validationErrorf("schedule %q not found", name)
		}
		return err
	}
	m.Root = root
	if err := validateStoredScheduledFlow(ctx, root, m.Flow); err != nil {
		return err
	}
	state, err := schedule.LoadFireState(root, m.Name)
	if err != nil {
		return err
	}

	pmBin, _ := os.Executable()
	backend := schedule.SelectBackendFromConfig(ctx, forceCrontab, nil, scheduleConfig(cfg))

	if err := backend.Install(ctx, m, pmBin); err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ScheduleInstall", "ok": true, "schedule": m, "status": state, "backend": string(backend.Kind())})
	}
	_, _ = fmt.Fprintf(stdout, "Installed schedule %s via %s (flow: %s, status: %s)\n", m.Name, backend.Kind(), m.Flow, state.Status)
	return nil
}

func runScheduleRemove(ctx context.Context, cfg config.Config, root string, args []string, stdout io.Writer, jsonOut bool) error {
	flags := parseFlags(args)
	positionals := flags.values["_"]
	if len(positionals) == 0 {
		return usageErrorf("pm schedule remove <name> [--crontab]")
	}
	name := positionals[0]
	forceCrontab := flags.first("crontab") == "true"

	m, err := schedule.Load(root, name)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return validationErrorf("schedule %q not found", name)
		}
		return err
	}
	state, err := schedule.LoadFireState(root, name)
	if err != nil {
		return err
	}

	if err := schedule.Delete(root, name); err != nil {
		return err
	}

	// Remove the manifest (and its fire lock) before touching an external
	// scheduler. A stale backend entry then fails before provider dispatch;
	// removing an installed entry while a fire owns its lease is never safe.
	// Backend removal remains best-effort because a missing scheduler binary
	// cannot resurrect the now-deleted, authorization-bound schedule.
	backendCfg := scheduleConfig(cfg)
	backend := schedule.SelectBackendFromConfig(ctx, forceCrontab, nil, backendCfg)
	_ = backend.Remove(ctx, name)
	if backend.Kind() != schedule.KindCrontab {
		_ = (schedule.CrontabBackend{File: backendCfg.CrontabFile}).Remove(ctx, name)
	}

	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ScheduleRemove", "ok": true, "name": name, "flow": m.Flow, "status": state})
	}
	_, _ = fmt.Fprintf(stdout, "Removed schedule %s (flow: %s, status: %s)\n", name, m.Flow, state.Status)
	return nil
}

func runScheduleFire(ctx context.Context, cfg config.Config, root string, args []string, stdout io.Writer, jsonOut bool) error {
	return withApp(root, func(a *app.App) error {
		return runScheduleFireWithApp(ctx, cfg, root, a, args, stdout, jsonOut)
	})
}

func runScheduleFireWithApp(ctx context.Context, cfg config.Config, root string, a *app.App, args []string, stdout io.Writer, jsonOut bool) error {
	flags := parseFlags(args)
	positionals := flags.values["_"]
	if len(positionals) != 1 {
		return usageErrorf("pm schedule fire <name>")
	}
	m, err := schedule.Load(root, positionals[0])
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return validationErrorf("schedule %q not found", positionals[0])
		}
		return err
	}
	if flags.first("authorization") != "" {
		return usageErrorf("schedule fire derives approval from flow jobs; --authorization is not accepted")
	}

	var flowOutput bytes.Buffer
	var result flow.RunResult
	err = flowRun(ctx, cfg, a, []string{m.Flow}, &flowOutput, true)
	if err == nil {
		if unmarshalErr := json.Unmarshal(flowOutput.Bytes(), &result); unmarshalErr != nil {
			err = fmt.Errorf("decode scheduled flow result: %w", unmarshalErr)
		}
	}
	if err != nil {
		return err
	}
	receiptIDs := flowReceiptIDs(result)
	state, err := schedule.LoadFireState(root, m.Name)
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "ScheduleFire", "ok": true, "schedule": m, "status": state, "flow": result})
	}
	_, _ = fmt.Fprintf(stdout, "Schedule %s fired: flow=%s status=%s receipts=%s\n", m.Name, result.FlowName, result.Status, strings.Join(receiptIDs, ","))
	return nil
}

func validateStoredScheduledFlow(ctx context.Context, root, flowName string) error {
	if _, _, err := schedule.FindByFlow(root, flowName); err != nil {
		return err
	}
	return withApp(root, func(a *app.App) error {
		path := filepath.Join(a.ProjectDir(), "flows", flowName+".json")
		manifest, err := readManifestFile(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return &schedule.FlowReferenceError{Flow: flowName, Reason: schedule.FlowReferenceMissing, Err: err}
			}
			return &schedule.FlowReferenceError{Flow: flowName, Reason: schedule.FlowReferenceInvalid, Err: err}
		}
		if manifest.Name != flowName {
			return &schedule.FlowReferenceError{Flow: flowName, Reason: schedule.FlowReferenceInvalid, Err: errors.New("stored manifest name does not match its reference")}
		}
		if _, err := resolveManifestJobs(ctx, a, manifest); err != nil {
			return &schedule.FlowReferenceError{Flow: flowName, Reason: schedule.FlowReferenceInvalid, Err: err}
		}
		return nil
	})
}

func flowReceiptIDs(result flow.RunResult) []string {
	ids := []string{}
	for _, step := range result.Steps {
		ids = append(ids, step.ReceiptIDs...)
	}
	sort.Strings(ids)
	return ids
}

func flowPreparedExecutionIdentities(result flow.RunResult) []string {
	identities := []string{}
	for _, step := range result.Steps {
		if step.PreparedExecutionIdentity != "" {
			identities = append(identities, step.PreparedExecutionIdentity)
		}
	}
	sort.Strings(identities)
	return identities
}

func scheduleFireStopReason(err error) schedule.FireStopReason {
	var changed *app.AuthorizationScopeChangedError
	if errors.As(err, &changed) {
		return schedule.FireStopScope
	}
	var revoked *app.AuthorizationRevokedError
	if errors.As(err, &revoked) {
		return schedule.FireStopRevoked
	}
	var expired *app.AuthorizationExpiredError
	if errors.As(err, &expired) {
		return schedule.FireStopExpired
	}
	var rateLimited *connsdk.RateLimitError
	if errors.As(err, &rateLimited) {
		return schedule.FireStopRateLimit
	}
	var rateBudgetRefused *connsdk.RateBudgetRefusalError
	if errors.As(err, &rateBudgetRefused) {
		return schedule.FireStopRateLimit
	}
	return schedule.FireStopFailed
}

func scheduleConfig(cfg config.Config) schedule.BackendConfig {
	backendCfg := schedule.BackendConfig{CrontabFile: cfg.Schedule.CrontabFile}
	if cfg.IsExplicit("runtime.temporal_addr") {
		backendCfg.TemporalAddr = cfg.Runtime.TemporalAddr
	}
	return backendCfg
}

func isAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	// schedule.Save returns a plain error with "already exists" in the message
	return containsStr(err.Error(), "already exists")
}

func containsStr(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
