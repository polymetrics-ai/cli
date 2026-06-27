# RUNBOOK — Phase 3: Scheduling

Date: 2026-06-27

---

## Normal operations

### Create a schedule

```bash
pm schedule create --name nightly-leads --cron "0 2 * * *" --flow likely-customers
```
Writes `<root>/schedules/nightly-leads.json`. Does NOT install any OS timer yet.

### Install a schedule (macOS)

```bash
pm schedule install nightly-leads
```
Writes `~/Library/LaunchAgents/ai.polymetrics.schedule.nightly-leads.plist` and runs `launchctl load -w <path>`. 
The payload that runs: `pm flow run likely-customers --json`.

### Install a schedule (Linux / systemd)

```bash
pm schedule install nightly-leads
```
Writes `~/.config/systemd/user/pm-schedule-nightly-leads.{service,timer}` and runs `systemctl --user enable --now pm-schedule-nightly-leads.timer`.

### Force crontab mode (any OS)

```bash
pm schedule install nightly-leads --crontab
```
Appends a crontab line via `crontab -l | cat - | crontab -` pattern.

### Install via Temporal (opt-in)

```bash
POLYMETRICS_TEMPORAL_ADDR=localhost:7233 pm schedule install nightly-leads
```
Registers a Temporal cron workflow. Requires Temporal to be reachable.

### List schedules

```bash
pm schedule list
pm schedule list --json
```
Shows name, cron, flow, next-run time (computed from `CronExpr.Next(time.Now())`), and installed backend if detectable.

### Remove a schedule

```bash
pm schedule remove nightly-leads
```
Removes the manifest file and, if the OS timer is present, unloads/disables/removes it.

---

## Incident procedures

### Schedule stopped firing (launchd)

1. Check if the plist is loaded: `launchctl list | grep ai.polymetrics`
2. If absent: `pm schedule install <name>` to reload.
3. Check launchd log: `log show --predicate 'subsystem == "ai.polymetrics"' --last 1h`
4. If the flow itself is failing: `pm flow run <name> --json` manually and inspect output.

### Schedule stopped firing (systemd)

1. `systemctl --user status pm-schedule-<name>.timer`
2. `journalctl --user -u pm-schedule-<name>.service --since "1 hour ago"`
3. Re-enable: `pm schedule remove <name> && pm schedule install <name>`

### Schedule stopped firing (Temporal)

1. Check Temporal UI or `tctl schedule describe --schedule-id pm-schedule-<name>`.
2. Check for paused state. Resume via Temporal UI or tctl.
3. Re-register: `pm schedule remove <name> && pm schedule install <name>`.

### Duplicate schedule installed

Symptom: two crontab lines for the same schedule.
Cause: `install` run twice without `remove` (for crontab backend only — launchd/systemd are idempotent by file path).
Fix: `pm schedule remove <name> && pm schedule install <name>`.

### `pm` binary path changed after install

The installed timer still points to the old binary path.
Fix: `pm schedule remove <name> && pm schedule install <name>` (re-resolves `os.Executable()`).

---

## Verification after phase completion

```bash
export GOTOOLCHAIN=auto
gofmt -w internal/schedule internal/cli
go vet ./...
go test ./internal/schedule/... ./internal/cli/...
go build ./cmd/pm
make verify
```
All must exit 0.
