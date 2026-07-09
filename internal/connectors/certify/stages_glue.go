// Stage implementations for the "glue" stages of the certification harness
// (docs/architecture/connector-certification-design.md §A "Glue stages":
// 18 flow_roundtrip, 19 schedule_roundtrip) plus the two meta-stages that
// must see everything captured across the WHOLE run once flow/schedule
// stages are added (20 secret_redaction_live, 21 json_contract). Write
// stages (12-17) are out of scope for this file (stages_write.go, a
// separate later task).
package certify

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// --- stage 18: flow_roundtrip ---

// flowManifestStep/flowManifestFile mirror the ephemeral flow manifest
// generated for flow_roundtrip (design §D): a capture-backed etl (sync) step
// feeding a dependent query step. Built by hand (rather than importing
// internal/flow's own types) so its exact on-disk shape matches
// flow.FlowManifest's JSON tags without importing internal/flow (certify
// drives the CLI only, per package layout).
type flowManifestStep struct {
	ID         string   `json:"id"`
	Kind       string   `json:"kind"`
	Connection string   `json:"connection,omitempty"`
	Streams    []string `json:"streams,omitempty"`
	SQL        string   `json:"sql,omitempty"`
	In         []string `json:"in"`
	Out        []string `json:"out"`
}

type flowManifestFile struct {
	Version int                `json:"version"`
	Name    string             `json:"name"`
	Steps   []flowManifestStep `json:"steps"`
}

const (
	flowSyncStepID  = "cert_sync"
	flowQueryStepID = "cert_query"
)

// flowName / flowTable are deterministic per-connector names for the
// ephemeral cert flow manifest and the table its query step reads.
func flowName(connector string) string           { return "cert_flow_" + connector }
func flowTable(connector string) string          { return "cert_flow_" + connector }
func flowConnectionName(connector string) string { return "cert_flow_conn_" + connector }

func (rc *runContext) flowName() string {
	name := flowName(rc.opts.Connector)
	if rc.currentStream != "" {
		name += "_" + safeName(rc.currentStream)
	}
	return name
}

func (rc *runContext) flowTable() string {
	table := flowTable(rc.opts.Connector)
	if rc.currentStream != "" {
		table += "_" + safeName(rc.currentStream)
	}
	return table
}

func (rc *runContext) flowConnectionName() string {
	name := flowConnectionName(rc.opts.Connector)
	if rc.currentStream != "" {
		name += "_" + safeName(rc.currentStream)
	}
	return name
}

func (rc *runContext) flowQueryTable() string {
	table := "cert_flow_query_" + rc.opts.Connector
	if rc.currentStream != "" {
		table += "_" + safeName(rc.currentStream)
	}
	return table
}

func (rc *runContext) scheduleName() string {
	name := "cert-schedule-" + rc.opts.Connector
	if rc.currentStream != "" {
		name += "-" + safeScheduleName(rc.currentStream)
	}
	if len(name) > 64 {
		return name[:64]
	}
	return name
}

func safeScheduleName(name string) string {
	return strings.ReplaceAll(strings.ToLower(safeName(name)), "_", "-")
}

// stageFlowRoundtrip drives stage 18: it registers a dedicated capture-backed
// connection (mirroring stages 6/7/10's replay-through-file-connector
// pattern), writes a two-step flow manifest referencing it, then exercises
// plan / preview (dry_run, zero side effects) / run (completed) / status
// (both steps done) through the real `pm flow` CLI surface.
func stageFlowRoundtrip(rc *runContext, rep *Report) error {
	if rc.capturePath == "" {
		skipStage(rc, rep, "flow_roundtrip", "skipped: no capture available (etl_full_refresh_append did not produce one)")
		return nil
	}

	// setupCaptureConnection's "mode" parameter doubles as a registration key
	// (it is idempotent per certify run, keyed off rc.captureFileRegistered,
	// not off the mode value itself), so passing "flow" here is safe -- it
	// just ensures the shared cert-file credential exists.
	if ok, errMsg := rc.setupCaptureConnection(rep, "flow", flowTable(rc.opts.Connector)); !ok {
		recordStage(rc, rep, "flow_roundtrip", 1, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, errMsg
		})
		return nil
	}

	table := rc.flowTable()
	connName := rc.flowConnectionName()
	connOK := recordStage(rc, rep, "flow_connection_create", 1, func() (bool, CLIStageInfo, string) {
		res := rc.run("connections", "create", connName,
			"--source", "file:"+rc.fileCredentialName(),
			"--destination", "warehouse:"+warehouseCredentialName,
			"--stream", rc.captureStreamName(),
			"--primary-key", rc.primaryKey(),
			"--cursor", rc.cursorField(),
			"--sync-mode", "full_refresh_append",
			"--table", table,
			"--json")
		passed, errMsg := assertKind(rc, "flow_connection_create", res, "Connection", 0)
		return passed, cliInfoFrom(res), errMsg
	})
	if !connOK.Passed {
		recordStage(rc, rep, "flow_roundtrip", 1, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, "flow_roundtrip: capture connection for flow step failed"
		})
		return nil
	}

	flowsDir := filepath.Join(rc.root, ".polymetrics", "flows")
	if err := os.MkdirAll(flowsDir, 0o755); err != nil {
		recordStage(rc, rep, "flow_roundtrip", 1, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, fmt.Sprintf("flow_roundtrip: mkdir flows dir: %v", err)
		})
		return nil
	}

	name := rc.flowName()
	queryTable := rc.flowQueryTable()
	manifest := flowManifestFile{
		Version: 1,
		Name:    name,
		Steps: []flowManifestStep{
			{
				ID:         flowSyncStepID,
				Kind:       "sync",
				Connection: connName,
				Streams:    []string{rc.captureStreamName()},
				In:         []string{},
				Out:        []string{table},
			},
			{
				ID:   flowQueryStepID,
				Kind: "query",
				SQL:  "SELECT * FROM " + table,
				In:   []string{table},
				Out:  []string{queryTable},
			},
		},
	}
	raw, err := json.Marshal(manifest)
	if err != nil {
		recordStage(rc, rep, "flow_roundtrip", 1, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, fmt.Sprintf("flow_roundtrip: marshal manifest: %v", err)
		})
		return nil
	}
	manifestPath := filepath.Join(flowsDir, name+".json")
	if err := os.WriteFile(manifestPath, raw, 0o644); err != nil {
		recordStage(rc, rep, "flow_roundtrip", 1, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, fmt.Sprintf("flow_roundtrip: write manifest: %v", err)
		})
		return nil
	}

	// `pm flow run` (when driven with a real *app.App, as every in-process
	// cli.Run call here is) persists step checkpoints under the PROJECT dir
	// (a.ProjectDir(), i.e. <root>/.polymetrics) regardless of --flows-dir,
	// while `pm flow status` reads both the manifest AND the checkpoint
	// store from whatever --flows-dir it is given. So flow_status must be
	// pointed at the project dir (not the flows/ subdirectory used for
	// --file / flow_run's manifest lookup) to see the checkpoints flow_run
	// just wrote -- which in turn means a copy of the manifest must also
	// exist there under <name>.json for flow_status's own manifest lookup.
	projectDir := filepath.Join(rc.root, ".polymetrics")
	projectManifestPath := filepath.Join(projectDir, name+".json")
	if err := os.WriteFile(projectManifestPath, raw, 0o644); err != nil {
		recordStage(rc, rep, "flow_roundtrip", 1, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, fmt.Sprintf("flow_roundtrip: write project-dir manifest copy: %v", err)
		})
		return nil
	}

	var order []any
	planStage := recordStage(rc, rep, "flow_plan", 1, func() (bool, CLIStageInfo, string) {
		res := rc.run("flow", "plan", "--file", manifestPath, "--json")
		passed, errMsg := assertKind(rc, "flow_plan", res, "", 0)
		if !passed {
			return false, cliInfoFrom(res), errMsg
		}
		if status, _ := res.Envelope["status"].(string); status != "ok" {
			return false, cliInfoFrom(res), fmt.Sprintf("flow_plan: status=%q, want ok", status)
		}
		order, _ = res.Envelope["order"].([]any)
		if len(order) != 2 {
			return false, cliInfoFrom(res), fmt.Sprintf("flow_plan: order has %d steps, want 2: %v", len(order), order)
		}
		if s, _ := order[0].(string); s != flowSyncStepID {
			return false, cliInfoFrom(res), fmt.Sprintf("flow_plan: order[0]=%q, want %q (sync step must precede its dependent query step)", s, flowSyncStepID)
		}
		if s, _ := order[1].(string); s != flowQueryStepID {
			return false, cliInfoFrom(res), fmt.Sprintf("flow_plan: order[1]=%q, want %q", s, flowQueryStepID)
		}
		return true, cliInfoFrom(res), ""
	})

	// Zero-side-effects assertion for preview: the query step's output table
	// must not exist before OR after `flow preview` (dry_run never touches
	// the warehouse).
	preSideEffect := queryTableExists(rc, queryTable)

	previewStage := recordStage(rc, rep, "flow_preview", 1, func() (bool, CLIStageInfo, string) {
		res := rc.run("flow", "preview", "--file", manifestPath, "--json")
		passed, errMsg := assertKind(rc, "flow_preview", res, "", 0)
		if !passed {
			return false, cliInfoFrom(res), errMsg
		}
		if status, _ := res.Envelope["status"].(string); status != "dry_run" {
			return false, cliInfoFrom(res), fmt.Sprintf("flow_preview: status=%q, want dry_run", status)
		}
		if postSideEffect := queryTableExists(rc, queryTable); postSideEffect != preSideEffect {
			return false, cliInfoFrom(res), fmt.Sprintf("flow_preview: dry_run had a side effect (query table existence changed: before=%v after=%v)", preSideEffect, postSideEffect)
		}
		return true, cliInfoFrom(res), ""
	})

	runStage := recordStage(rc, rep, "flow_run", 1, func() (bool, CLIStageInfo, string) {
		res := rc.run("flow", "run", "--file", manifestPath, "--flows-dir", flowsDir, "--json")
		passed, errMsg := assertKind(rc, "flow_run", res, "", 0)
		if !passed {
			return false, cliInfoFrom(res), errMsg
		}
		status, _ := res.Envelope["status"].(string)
		if status != "ok" {
			return false, cliInfoFrom(res), fmt.Sprintf("flow_run: status=%q, want ok (run must complete)", status)
		}
		steps, _ := res.Envelope["steps"].([]any)
		if len(steps) != 2 {
			return false, cliInfoFrom(res), fmt.Sprintf("flow_run: steps has %d entries, want 2: %v", len(steps), steps)
		}
		for _, raw := range steps {
			step, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if st, _ := step["status"].(string); st != "ok" {
				id, _ := step["id"].(string)
				return false, cliInfoFrom(res), fmt.Sprintf("flow_run: step %q status=%q, want ok", id, st)
			}
		}
		return true, cliInfoFrom(res), ""
	})

	statusStage := recordStage(rc, rep, "flow_status", 1, func() (bool, CLIStageInfo, string) {
		res := rc.run("flow", "status", name, "--flows-dir", projectDir, "--json")
		passed, errMsg := assertKind(rc, "flow_status", res, "", 0)
		if !passed {
			return false, cliInfoFrom(res), errMsg
		}
		steps, _ := res.Envelope["steps"].([]any)
		if len(steps) != 2 {
			return false, cliInfoFrom(res), fmt.Sprintf("flow_status: steps has %d entries, want 2: %v", len(steps), steps)
		}
		for _, raw := range steps {
			step, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			if st, _ := step["status"].(string); st != "success" {
				id, _ := step["id"].(string)
				return false, cliInfoFrom(res), fmt.Sprintf("flow_status: step %q status=%q, want success", id, st)
			}
		}
		return true, cliInfoFrom(res), ""
	})

	passed := planStage.Passed && previewStage.Passed && runStage.Passed && statusStage.Passed
	recordStage(rc, rep, "flow_roundtrip", 1, func() (bool, CLIStageInfo, string) {
		if !passed {
			return false, CLIStageInfo{}, "flow_roundtrip: one or more of plan/preview/run/status failed (see named sub-stages)"
		}
		return true, CLIStageInfo{}, ""
	})
	if passed {
		rep.Capabilities.Flow = &CapabilityResult{Result: "pass"}
	} else {
		rep.Capabilities.Flow = &CapabilityResult{Result: "fail", Reason: "flow_roundtrip: see flow_plan/flow_preview/flow_run/flow_status stages"}
	}
	return nil
}

// queryTableExists runs `pm query run --table <table> --json` and reports
// whether the table is currently readable (exit 0, kind QueryResult). A
// non-existent warehouse table surfaces as a non-zero exit / non-QueryResult
// kind, so this is a cheap, CLI-only existence probe with no other side
// effects (query run never writes).
func queryTableExists(rc *runContext, table string) bool {
	res := rc.run("query", "run", "--table", table, "--json")
	return res.ExitCode == 0 && res.Kind == "QueryResult"
}

// --- stage 19: schedule_roundtrip ---

// scheduleCrontabFileName is the ephemeral crontab file the harness redirects
// ALL crontab reads/writes to for the duration of schedule_roundtrip
// (internal/schedule/crontab.go's PM_CRONTAB_FILE seam) — the real
// certification run must never touch the operator's actual crontab.
const scheduleCrontabFileName = "cert-crontab"

// stageScheduleRoundtrip drives stage 19 (design §D): snapshot the
// (redirected, ephemeral) crontab, create + list + install --crontab, assert
// the "# pm-schedule-<name>" sentinel is present, remove, assert the
// sentinel is absent AND the crontab is byte-identical to the pre-create
// snapshot. Any residue (leftover sentinel or crontab drift) is reported as
// Capabilities.Schedule.Residue = true, the same severity class as a leaked
// resource (design §D "residue -> leaked_schedule").
func stageScheduleRoundtrip(rc *runContext, rep *Report) error {
	crontabPath := filepath.Join(rc.root, scheduleCrontabFileName)

	// PM_CRONTAB_FILE redirects internal/schedule's CrontabBackend to this
	// ephemeral file for the duration of this stage, so the real operator
	// crontab is never read or written (schedule_test.go uses the identical
	// seam). Restored unconditionally via defer, including on early return.
	prevCrontabFile, hadPrevCrontabFile := os.LookupEnv("PM_CRONTAB_FILE")
	if err := os.Setenv("PM_CRONTAB_FILE", crontabPath); err != nil {
		recordStage(rc, rep, "schedule_roundtrip", 2, func() (bool, CLIStageInfo, string) {
			return false, CLIStageInfo{}, fmt.Sprintf("schedule_roundtrip: set PM_CRONTAB_FILE: %v", err)
		})
		return nil
	}
	defer func() {
		if hadPrevCrontabFile {
			_ = os.Setenv("PM_CRONTAB_FILE", prevCrontabFile)
		} else {
			_ = os.Unsetenv("PM_CRONTAB_FILE")
		}
	}()

	// Snapshot BEFORE any create/install activity (design §D "snapshot
	// crontab -l"). The file may not exist yet — that is a valid, empty
	// snapshot.
	snapshot, _ := os.ReadFile(crontabPath)

	name := rc.scheduleName()
	sentinel := "# pm-schedule-" + name

	createStage := recordStage(rc, rep, "schedule_create", 2, func() (bool, CLIStageInfo, string) {
		res := rc.run("schedule", "create",
			"--name", name,
			"--cron", "0 3 * * *",
			"--flow", rc.flowName(),
			"--json")
		passed, errMsg := assertKind(rc, "schedule_create", res, "Schedule", 0)
		return passed, cliInfoFrom(res), errMsg
	})

	listStage := recordStage(rc, rep, "schedule_list", 2, func() (bool, CLIStageInfo, string) {
		res := rc.run("schedule", "list", "--json")
		passed, errMsg := assertKind(rc, "schedule_list", res, "ScheduleList", 0)
		if !passed {
			return false, cliInfoFrom(res), errMsg
		}
		schedules, _ := res.Envelope["schedules"].([]any)
		found := false
		for _, item := range schedules {
			m, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if n, _ := m["name"].(string); n == name {
				found = true
				break
			}
		}
		if !found {
			return false, cliInfoFrom(res), fmt.Sprintf("schedule_list: schedule %q not present in list", name)
		}
		return true, cliInfoFrom(res), ""
	})

	installStage := recordStage(rc, rep, "schedule_install", 2, func() (bool, CLIStageInfo, string) {
		res := rc.run("schedule", "install", name, "--crontab", "--json")
		passed, errMsg := assertKind(rc, "schedule_install", res, "ScheduleInstall", 0)
		if !passed {
			return false, cliInfoFrom(res), errMsg
		}
		if backend, _ := res.Envelope["backend"].(string); backend != "crontab" {
			return false, cliInfoFrom(res), fmt.Sprintf("schedule_install: backend=%q, want crontab", backend)
		}
		content, err := os.ReadFile(crontabPath)
		if err != nil {
			return false, cliInfoFrom(res), fmt.Sprintf("schedule_install: read crontab file: %v", err)
		}
		if !strings.Contains(string(content), sentinel) {
			return false, cliInfoFrom(res), fmt.Sprintf("schedule_install: sentinel %q not present in crontab after install", sentinel)
		}
		return true, cliInfoFrom(res), ""
	})

	removeStage := recordStage(rc, rep, "schedule_remove", 2, func() (bool, CLIStageInfo, string) {
		res := rc.run("schedule", "remove", name, "--crontab", "--json")
		passed, errMsg := assertKind(rc, "schedule_remove", res, "ScheduleRemove", 0)
		return passed, cliInfoFrom(res), errMsg
	})

	sentinelAbsent := true
	crontabIdentical := true
	if removeStage.Passed {
		after, _ := os.ReadFile(crontabPath)
		sentinelAbsent = !strings.Contains(string(after), sentinel)
		crontabIdentical = string(after) == string(snapshot)
	}

	residue := !removeStage.Passed || !sentinelAbsent || !crontabIdentical
	if residue {
		// Force-remove the sentinel before reporting (design §D "harness
		// force-removes sentinel before reporting") so a failed certify run
		// never itself leaves a schedule installed.
		forceRemoveCrontabSentinel(crontabPath, sentinel)
	}

	overallPassed := createStage.Passed && listStage.Passed && installStage.Passed && removeStage.Passed && !residue
	recordStage(rc, rep, "schedule_roundtrip", 2, func() (bool, CLIStageInfo, string) {
		if !overallPassed {
			reason := "schedule_roundtrip: one or more of create/list/install/remove failed (see named sub-stages)"
			if residue && createStage.Passed && listStage.Passed && installStage.Passed && removeStage.Passed {
				reason = fmt.Sprintf("schedule_roundtrip: residue after remove (sentinel_absent=%v crontab_identical=%v)", sentinelAbsent, crontabIdentical)
			}
			return false, CLIStageInfo{}, reason
		}
		return true, CLIStageInfo{}, ""
	})

	result := "pass"
	reason := ""
	if !overallPassed {
		result = "fail"
		reason = "schedule_roundtrip: see schedule_create/schedule_list/schedule_install/schedule_remove stages"
	}
	rep.Capabilities.Schedule = &ScheduleResult{
		Result:  result,
		Backend: "crontab",
		Residue: residue,
		Reason:  reason,
	}
	return nil
}

// forceRemoveCrontabSentinel strips any line containing sentinel from the
// crontab file at path, best-effort (design §D "harness force-removes
// sentinel before reporting"). Errors are swallowed: this is a
// belt-and-suspenders cleanup on an already-failed stage, not a new
// assertion surface.
func forceRemoveCrontabSentinel(path, sentinel string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}
	lines := strings.Split(string(content), "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.Contains(line, sentinel) {
			continue
		}
		kept = append(kept, line)
	}
	_ = os.WriteFile(path, []byte(strings.Join(kept, "\n")), 0o644)
}
