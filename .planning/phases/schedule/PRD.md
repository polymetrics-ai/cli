# PRD — Phase 3: Scheduling (`pm schedule`)

Date: 2026-06-27
Phase: schedule (Phase 3 of 5)
Status: planning

---

## Problem

Flows built in phases 0–2 must run on a recurring basis (nightly enrichment, daily sync, hourly scoring). Users need a way to bind a cron expression to a flow name and install that schedule onto the host OS or Temporal — without running a separate daemon process.

## Goals

1. Let users create a named schedule that maps a cron expression to a flow name.
2. Let users install that schedule as a native OS timer (launchd plist on macOS, systemd-user timer on Linux, or a crontab line as a fallback) — no resident daemon.
3. Let users optionally install to Temporal if `POLYMETRICS_TEMPORAL_ADDR` is set (overlap-prevention, retries, durability).
4. All four verbs — `create`, `list`, `install`, `remove` — must work offline, stdlib-only, with zero new dependencies.

## Non-goals

- No resident daemon.
- No Windows support in this phase.
- No UI.
- No new third-party Go modules.
- No cron execution engine inside `pm` (OS/Temporal handles execution).

## Success criteria

- `pm schedule create --name nightly-leads --cron "0 2 * * *" --flow likely-customers` writes a manifest file.
- `pm schedule install nightly-leads` on macOS writes a `.plist` to `~/Library/LaunchAgents/` and loads it.
- `pm schedule install nightly-leads` on Linux writes a `.timer`/`.service` pair to `~/.config/systemd/user/` and runs `systemctl --user enable --now`.
- `pm schedule install nightly-leads --crontab` writes a crontab line regardless of OS.
- With `POLYMETRICS_TEMPORAL_ADDR` set, `install` registers a Temporal cron workflow instead.
- `pm schedule list` shows all manifests with next-run time.
- `pm schedule remove nightly-leads` removes the manifest and unloads/unregisters the timer.
- `make verify` green; all behavior covered by table-driven tests.

## Design-direction

Not applicable — this is a CLI/backend phase.
