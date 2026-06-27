# RELEASE-NOTES — Phase 3: Scheduling

Date: 2026-06-27
Status: draft (pre-implementation)

---

## What's new

### `pm schedule` — daemon-free flow scheduling

Bind a cron expression to any flow and install it as a native OS timer or Temporal cron workflow.

```bash
pm schedule create --name nightly-leads --cron "0 2 * * *" --flow likely-customers
pm schedule install nightly-leads
pm schedule list
pm schedule remove nightly-leads
```

**No resident daemon.** The `install` command writes a native OS timer (launchd plist on macOS, systemd-user timer on Linux, or a crontab line) whose payload is `pm flow run <name> --json`. The OS scheduler owns execution.

**Temporal opt-in.** Set `POLYMETRICS_TEMPORAL_ADDR` to route to a Temporal cron workflow instead, gaining overlap-prevention, automatic retries, and durable history.

**Force crontab.** `pm schedule install <name> --crontab` writes a crontab line regardless of OS.

---

## Behavior guarantees

- `create` validates the cron expression at write time (before any file is installed).
- `install` validates the binary path via `os.Executable()`.
- OS unit files are written with mode `0600`.
- Flow names and schedule names are slug-validated; no shell injection possible.
- `list` computes next-run time locally — no daemon needed.

---

## Not included in this phase

- Windows support.
- Overlap-prevention on non-Temporal backends (the OS handles this per-process; Temporal adds workflow-level locking).
- Schedule history or last-run tracking (that comes from `pm flow status` / the flow ledger from Phase 0).

---

## Upgrade notes

No schema migrations. No new dependencies. Existing `pm` users can add scheduling without any infrastructure changes.
