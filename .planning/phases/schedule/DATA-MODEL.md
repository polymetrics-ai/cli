# DATA-MODEL — Phase 3: Scheduling

Date: 2026-06-27

---

## Storage: filesystem only (no schema migration required)

Schedules are stored as JSON files under `<pm-root>/schedules/`. No database tables, no migrations. This is a deliberate design choice — schedules must work without Postgres.

---

## Manifest file

Path: `<pm-root>/schedules/<name>.json`
Permissions: `0600`
Encoding: UTF-8 JSON

```json
{
  "name": "nightly-leads",
  "cron": "0 2 * * *",
  "flow": "likely-customers",
  "created_at": "2026-06-27T00:00:00Z",
  "updated_at": "2026-06-27T00:00:00Z"
}
```

Field constraints:

| Field | Type | Constraints |
|-------|------|-------------|
| name | string | slug `[a-z0-9][a-z0-9-]*`, 1–64 chars, unique per root |
| cron | string | valid 5-field cron expression (validated by `ParseCron`) |
| flow | string | non-empty, validated to exist at `install` time only |
| created_at | string | RFC3339 UTC |
| updated_at | string | RFC3339 UTC |

---

## OS-level artifacts (not managed as data models)

These are side effects of `pm schedule install`. They are not parsed by `pm` after creation — they are managed entirely by the OS scheduler.

### launchd plist

Path: `~/Library/LaunchAgents/ai.polymetrics.schedule.<name>.plist`
Label: `ai.polymetrics.schedule.<name>`

### systemd units

Paths: `~/.config/systemd/user/pm-schedule-<name>.{service,timer}`

### crontab line

Added to user's crontab via `crontab` command. Tagged with sentinel comment `# pm-schedule-<name>`.

### Temporal schedule (opt-in)

WorkflowID: `pm-schedule-<name>`
Stored in Temporal server — not on the local filesystem.

---

## Schema migration

None required. The manifest format is append-compatible. Old manifests without new fields will have Go zero-values for those fields, which is safe.

No human gate needed for the current manifest schema.
