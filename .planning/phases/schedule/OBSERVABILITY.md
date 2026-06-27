# OBSERVABILITY — Phase 3: Scheduling

Date: 2026-06-27

---

## Philosophy

`pm schedule` is a manifest-and-install tool, not a long-running daemon. Observability is primarily through:
1. CLI output (human text or `--json` envelope).
2. OS-level scheduler logs (launchd, systemd journal, Temporal UI).
3. The flow run output — `pm flow run <name> --json` — which carries its own observability (Phase 0 ledger/receipts).

---

## CLI output fields

Every `--json` response includes:

```json
{
  "ok": true|false,
  "error": "..." ,          // present only on error
  "schedule": {...},        // on create/list
  "backend": "launchd",     // on install
  "unit": "/path/to/file"   // on install (OS backends)
}
```

These are machine-readable and consumed by agents.

---

## Structured log events

`pm schedule install` emits to stderr (human mode) or omits (json mode) a summary line:
```
schedule "nightly-leads" installed via launchd (~/Library/LaunchAgents/...)
```

Errors include the backend kind and the underlying OS error:
```json
{"ok": false, "error": "launchd: launchctl load failed: exit status 1"}
```

---

## OS-level observability

### launchd (macOS)
```bash
# List loaded pm schedules
launchctl list | grep ai.polymetrics

# View logs
log show --predicate 'subsystem CONTAINS "ai.polymetrics"' --last 1h
```

### systemd (Linux)
```bash
systemctl --user status pm-schedule-<name>.timer
journalctl --user -u pm-schedule-<name>.service --since "1 hour ago"
```

### Temporal (opt-in)
- Temporal Web UI at the configured address.
- `tctl schedule describe --schedule-id pm-schedule-<name>`

---

## No metrics daemon

This phase does not emit Prometheus/OpenTelemetry metrics. Schedule-level execution metrics (records synced, errors, latency) are emitted by `pm flow run` itself (Phase 0 observability). The schedule layer is thin — it only manages OS-level registration.

---

## Next-run time

`pm schedule list` computes `next_run` via `CronExpr.Next(time.Now())` and includes it in the JSON output. This is a client-side computation (no daemon required).
