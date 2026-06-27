# API-CONTRACT — Phase 3: Scheduling

Date: 2026-06-27

---

## CLI verbs

All commands honor `--json` for machine-readable output and `--root <dir>` for the pm project root (inherited from global flag parsing).

### `pm schedule create`

```
pm schedule create --name <name> --cron <expr> --flow <flow-name>
```

Flags:
- `--name` (required) — slug `[a-z0-9][a-z0-9-]*`, max 64 chars.
- `--cron` (required) — standard 5-field cron expression.
- `--flow` (required) — flow manifest name to bind.

Success stdout (human): `schedule "nightly-leads" created`
Success stdout (`--json`):
```json
{"ok": true, "schedule": {"name": "nightly-leads", "cron": "0 2 * * *", "flow": "likely-customers", "created_at": "2026-06-27T00:00:00Z"}}
```

Error stdout (`--json`):
```json
{"ok": false, "error": "cron expression invalid: minute field out of range"}
```

Exit codes: 0 success, 1 validation error, 2 internal/IO error.

---

### `pm schedule list`

```
pm schedule list [--json]
```

Success stdout (`--json`):
```json
{"ok": true, "schedules": [
  {"name": "nightly-leads", "cron": "0 2 * * *", "flow": "likely-customers", "next_run": "2026-06-28T02:00:00Z"}
]}
```

Empty list returns `{"ok": true, "schedules": []}`.

---

### `pm schedule install`

```
pm schedule install <name> [--crontab] [--json]
```

Flags:
- `--crontab` — force crontab backend regardless of OS/Temporal.

Success stdout (`--json`):
```json
{"ok": true, "schedule": "nightly-leads", "backend": "launchd", "unit": "~/Library/LaunchAgents/ai.polymetrics.schedule.nightly-leads.plist"}
```

Error if name not found:
```json
{"ok": false, "error": "schedule \"nightly-leads\" not found"}
```

HUMAN GATE note: when `backend` is `"temporal"`, the output includes `"temporal_addr": "<addr>"` (masked if credentials present).

---

### `pm schedule remove`

```
pm schedule remove <name> [--json]
```

Success stdout (`--json`):
```json
{"ok": true, "schedule": "nightly-leads", "removed": true}
```

If the manifest is not found but the operation is otherwise clean, `"removed": false`.

---

## Go package public API

```go
package schedule

// Manifest is the serialized form of a schedule.
type Manifest struct {
    Name      string    `json:"name"`
    Cron      string    `json:"cron"`
    Flow      string    `json:"flow"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

// CronExpr is a parsed cron expression.
type CronExpr struct { /* opaque */ }

// ParseCron validates and parses a 5-field cron expression.
func ParseCron(expr string) (CronExpr, error)

// Next returns the next scheduled time after t.
func (c CronExpr) Next(t time.Time) time.Time

// Save writes a manifest to <root>/schedules/<name>.json.
func Save(root string, m Manifest) error

// Load reads a manifest by name.
func Load(root, name string) (Manifest, error)

// List returns all manifests sorted by name.
func List(root string) ([]Manifest, error)

// Delete removes a manifest file (does not uninstall OS timer).
func Delete(root, name string) error

// Backend installs/removes OS timers.
type Backend interface {
    Install(ctx context.Context, m Manifest, pmBin string) error
    Remove(ctx context.Context, name string) error
    Kind() string
}

// SelectBackend chooses the appropriate backend.
// probe is called to test Temporal reachability (pass nil to use runtimecheck default).
func SelectBackend(ctx context.Context, forceCrontab bool, probe func(ctx context.Context, addr string) bool) Backend
```

All public functions return typed errors. No panics across package boundaries.

---

## Stability contract

- `Manifest` JSON schema is stable; adding fields is non-breaking, removing is a breaking change requiring a new phase.
- `--json` envelope shape is stable: `{"ok": bool, ...}` top-level keys are stable.
- CLI flag names are stable once released.
