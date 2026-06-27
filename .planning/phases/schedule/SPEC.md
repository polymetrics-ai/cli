# SPEC — Phase 3: Scheduling

Date: 2026-06-27

---

## 1. Package: `internal/schedule`

Module path: `polymetrics.ai/internal/schedule`

### 1.1 Schedule manifest

Stored in `<pm-root>/schedules/<name>.json`.

```json
{
  "name": "nightly-leads",
  "cron": "0 2 * * *",
  "flow": "likely-customers",
  "created_at": "2026-06-27T00:00:00Z",
  "updated_at": "2026-06-27T00:00:00Z"
}
```

Fields:
- `name` — slug: `[a-z0-9][a-z0-9-]*`, max 64 chars.
- `cron` — standard 5-field POSIX cron expression.
- `flow` — name of an existing flow manifest (validated at `install` time, not `create` time).
- `created_at`, `updated_at` — RFC3339 UTC timestamps.

### 1.2 Cron validation

`ParseCron(expr string) (CronExpr, error)` — stdlib-only parser. Validates:
- Exactly 5 fields: minute, hour, dom, month, dow.
- Each field is `*`, a number, or a number range/list/step within its allowed range.
- Returns `CronExpr` with a `Next(t time.Time) time.Time` method for listing.

### 1.3 Backend selection

`SelectBackend(ctx context.Context) Backend` — returns one of:

| Condition | Backend |
|-----------|---------|
| `POLYMETRICS_TEMPORAL_ADDR` set and reachable (via `runtimecheck`) | `TemporalBackend` |
| macOS (`runtime.GOOS == "darwin"`) | `LaunchdBackend` |
| Linux with systemd user session | `SystemdBackend` |
| fallback | `CrontabBackend` |

`Backend` interface:

```go
type Backend interface {
    Install(ctx context.Context, m Manifest, pmBin string) error
    Remove(ctx context.Context, name string) error
    Kind() string // "launchd" | "systemd" | "crontab" | "temporal"
}
```

### 1.4 launchd backend

Label: `ai.polymetrics.schedule.<name>`
Plist path: `~/Library/LaunchAgents/ai.polymetrics.schedule.<name>.plist`
Payload program: `[pmBin, "flow", "run", flowName, "--json"]`
`RunAtLoad`: false
`StartCalendarInterval`: derived from cron fields (minute + hour + dom + month + dow mapped to plist keys).

After writing plist, runs `launchctl load -w <plist_path>`.
Remove runs `launchctl unload -w <plist_path>` then deletes the file.

### 1.5 systemd-user backend

Unit name: `pm-schedule-<name>`
Files written to: `~/.config/systemd/user/`
`pm-schedule-<name>.service`: runs `pmBin flow run <flowName> --json`.
`pm-schedule-<name>.timer`: `OnCalendar` derived from cron expression (converted via `cronToOnCalendar`).
After writing, runs `systemctl --user enable --now pm-schedule-<name>.timer`.
Remove: `systemctl --user disable --now`, delete files.

### 1.6 crontab backend

Appends a line: `<cron_expr>  <pmBin> flow run <flowName> --json  # pm-schedule-<name>`
Remove: removes that line.
Uses a sentinel comment so exact-match removal is possible.

### 1.7 Temporal backend

HUMAN GATE: requires `POLYMETRICS_TEMPORAL_ADDR` to be set and reachable. This path is opt-in.
Uses existing `internal/runtimecheck` to confirm Temporal is reachable before proceeding.
Registers a Temporal cron workflow schedule with the flow name as the workflow ID.
WorkflowID: `pm-schedule-<name>`
CronSchedule: the raw cron expression (Temporal accepts standard cron syntax).
The workflow body calls `pm flow run <flowName> --json` via the Temporal activity layer.

Note: Temporal workflow/activity registration requires the Temporal SDK (already in `go.mod`). No new deps needed.

### 1.8 `pm schedule` CLI verbs

All wired in `internal/cli/cli.go` under `case "schedule":` dispatching to `internal/schedule`.

```
pm schedule create  --name <name> --cron <expr> --flow <flow-name>
pm schedule list    [--json]
pm schedule install <name> [--crontab] [--json]
pm schedule remove  <name> [--json]
```

Exit codes: 0 success, 1 user error, 2 internal error — consistent with existing CLI pattern.

`--json` outputs an envelope: `{"ok": true, "schedule": {...}}` or `{"ok": false, "error": "..."}`.

## 2. File layout

```
internal/schedule/
    schedule.go          # Manifest type, Load/Save/List
    cron.go              # ParseCron, CronExpr, Next()
    backend.go           # Backend interface, SelectBackend
    launchd.go           # LaunchdBackend
    systemd.go           # SystemdBackend
    crontab.go           # CrontabBackend
    temporal.go          # TemporalBackend (build-gated by runtimecheck)
    schedule_test.go     # unit tests: cron parse, manifest round-trip
    cron_test.go         # table-driven cron validation + Next()
    launchd_test.go      # golden test: plist rendering
    systemd_test.go      # golden test: .service/.timer rendering
    crontab_test.go      # golden test: crontab line format
    backend_test.go      # SelectBackend: env-based selection logic
internal/cli/
    schedule.go          # runSchedule dispatcher (new file, wired into cli.go)
```

Golden test fixtures: `internal/schedule/testdata/`

## 3. Constraints

- stdlib only — no new imports beyond already-present `go.mod` entries.
- All file paths derived from `os.UserHomeDir()` — no hardcoded paths.
- `pmBin` for the installed payload resolved via `os.Executable()` at install time.
- `cronToOnCalendar` is a pure function (testable without systemd present).
- `startCalendarInterval` is a pure function (testable without launchd present).
- Tests that exec `launchctl`/`systemctl`/`crontab` are skipped when the binary is absent (`exec.LookPath` check in `t.Skip`).
