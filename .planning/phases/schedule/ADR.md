# ADR — Phase 3: Scheduling

Date: 2026-06-27

---

## ADR-001: No resident daemon

**Status:** Accepted

**Context:** We need recurring flow execution. Options: (a) a long-running `pm` daemon that wakes on a timer, (b) delegate to the OS scheduler, (c) delegate to Temporal.

**Decision:** Option (b) as the default — `pm schedule install` emits a native OS timer file. Option (c) is opt-in via `POLYMETRICS_TEMPORAL_ADDR`.

**Rationale:**
- A resident daemon would require a PID file, startup management, and crash recovery — all complexity that the OS scheduler already handles.
- launchd and systemd-user are the standard mechanisms for user-level periodic jobs on macOS and Linux respectively.
- The "no daemon" constraint is a project invariant (dependency-free default).

**Consequences:** Users must re-install the timer if they move the `pm` binary. Documented in RUNBOOK.md.

---

## ADR-002: Stdlib-only cron parser

**Status:** Accepted

**Context:** Several Go cron libraries exist (robfig/cron, etc.). We could import one.

**Decision:** Write a minimal 5-field cron parser in stdlib only.

**Rationale:**
- Project invariant: no new third-party dependencies on the default path.
- The use case is validation + `Next()` computation for `list`. We do not need a full scheduling engine.
- A focused parser is ~100 lines and fully testable.

**Consequences:** We only support standard 5-field cron (no `@reboot`, no 6-field seconds, no named months/days). This covers all practical use cases for flow scheduling.

---

## ADR-003: Temporal as opt-in via env var

**Status:** Accepted

**Context:** Temporal provides overlap-prevention, retries, and durable history — valuable for production scheduling. But it requires infrastructure.

**Decision:** `SelectBackend` checks `POLYMETRICS_TEMPORAL_ADDR` via `internal/runtimecheck.FromEnv()`. If set and reachable, TemporalBackend is selected. Otherwise, OS backend is used.

**Rationale:**
- Consistent with how Postgres and Dragonfly are made optional in this codebase.
- `internal/runtimecheck` already has the Temporal probe logic.
- No new code path for "does Temporal exist"; reuses existing adapter.

**Consequences:** HUMAN GATE — the Temporal path requires `POLYMETRICS_TEMPORAL_ADDR` set by the user. If Temporal is unreachable, `SelectBackend` falls back to the OS backend rather than erroring. This prevents breakage when the env var is set but Temporal is down.

---

## ADR-004: `ProbeFunc` injection for testability

**Status:** Accepted

**Context:** `SelectBackend` needs to check Temporal reachability. In tests, we cannot hit a real Temporal server.

**Decision:** `SelectBackend` accepts an optional `probe func(ctx context.Context, addr string) bool` parameter. Tests inject a fake probe. Production code passes `nil`, which falls back to `runtimecheck.Doctor`.

**Rationale:** Clean seam, no global state, stdlib-compatible, no test-only build tags.

**Consequences:** Callers must pass `nil` for the default behavior. The CLI wiring in `internal/cli/schedule.go` always passes `nil`.

---

## ADR-005: Unit files written with mode 0600

**Status:** Accepted

**Context:** launchd plist and systemd unit files contain the path to the `pm` binary and the flow name. They must not be world-readable/writable.

**Decision:** All files written by `pm schedule install` use `os.WriteFile(path, data, 0600)`.

**Rationale:** Reduces the attack surface per THREAT-MODEL TH-6.

**Consequences:** Other local user accounts on the same machine cannot read the schedule configuration. This is the desired behavior.
