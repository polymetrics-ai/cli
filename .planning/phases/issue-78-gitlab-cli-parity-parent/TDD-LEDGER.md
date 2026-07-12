# TDD Ledger: GitLab CLI Parity Parent Orchestration

## 2026-07-09 — parent setup

Task type: parent orchestration and planning.

### GSD Evidence

- `scripts/gsd doctor` — passed.
- `scripts/gsd verify-pi` — passed.
- `scripts/gsd list --json` — completed.
- `scripts/gsd prompt plan-phase issue-78-gitlab-cli-parity --skip-research` — generated the parent planning prompt.
- `scripts/gsd prompt programming-loop init --phase issue-78-gitlab-cli-parity --dry-run` — unavailable (`unknown GSD command: programming-loop`); manual universal programming loop fallback recorded in `PLAN.md` and `RUN-STATE.json`.

### Red/Green Plan

- Parent orchestration slice is planning-only; no production behavior changed.
- Behavior-changing sub-issues must start with red tests before production edits.
- #83 red target: embedded GitLab CLI surface test fails until `internal/connectors/defs/gitlab/cli_surface.json` exists and maps the four current stream-backed commands.

### Safety Notes

- No secrets requested or read.
- No credentialed GitLab checks.
- No external writes or destructive actions.
- No generic write tools or reverse ETL execution.

## 2026-07-09 — sub-issue execution summary

- #83: Red/green embedded CLI surface metadata test; pushed commit `feat(gitlab): add CLI surface metadata`.
- #84: Red help tests showed `pm gitlab`/`pm help gitlab` lacked connector-aware manual routing; green path renders connector manuals and JSON `CommandManual` envelopes.
- #85: Added local fixture test proving `pm gitlab issue list` uses the generic stream runner and Bearer auth from env-sourced credentials.
- #86: Added GitLab OpenAPI operation ledger (1,144 official operations + `/users` compatibility row) and disk-load test for coverage/blocked-by-default invariants.
- #87: Red direct-read policy test failed for unsupported `json_redacted`; green path adds recursive redaction and four GitLab bounded direct-read commands.
- #88: GraphQL/advanced support recorded as not required for this REST-backed slice; no generic GraphQL mutation/raw body executor added.
- #89: Sensitive/admin/destructive operations remain blocked by default with risk tiers, approval text, typed confirmation policy markers, and redaction policy.
