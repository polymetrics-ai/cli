---
status: resolved
trigger: "PR #4168 verify failed after the initial delivery; rebase onto the moved integration base, reproduce the failed job, and fix only branch-owned causes."
created: 2026-08-15
updated: 2026-08-15
---

# Symptoms

- Expected: PR #4168 passes the repository `verify` workflow after generated artifacts are current.
- Actual: the GitHub `verify` job failed.
- Error: `TestCertifyCLISingleConnectorPassExitsZero` exited 2 because the schedule certification
  sub-stages required an obsolete `authorization_reference`.
- Timeline: reported after #4167 merged into `integration/4015-mvp-flat-r1`.
- Reproduction: run the exact failed job step locally after rebasing onto the updated integration base.

# Current Focus

- hypothesis: confirmed — `internal/connectors/certify/stages_glue.go` retained the pre-correction
  schedule-authorization certification contract after production removed that carrier
- test: make the scripted driver reject authorization flags/fields, then rerun the exact fresh-binary
  CLI test and the focused certification package
- expecting: schedule create/list/install/remove pass without any authority carrier, with crontab
  cleanup still byte-identical
- next_action: update the certification test contract first, then the harness implementation
- reasoning_checkpoint: the failure is branch-owned and unrelated to #4167; the rebased production
  schedule path is behaving according to the captain's corrected model
- tdd_checkpoint: RED reproduced locally with the exact CI test

# Evidence

- timestamp: 2026-08-15T13:20:00Z
  observation: branch rebased without conflicts onto origin/integration/4015-mvp-flat-r1 at 6b9cfe492
- timestamp: 2026-08-15T13:33:00Z
  observation: GitHub run 31886502846 failed only `verify` → `certify-timing` →
    `TestCertifyCLISingleConnectorPassExitsZero`; schedule list/install/remove each reported
    "authorization reference was not preserved"
- timestamp: 2026-08-15T13:38:00Z
  observation: focused local reproduction failed identically; captured argv included the obsolete
    `--authorization auth_0123456789abcdef` supplied by `stageScheduleRoundtrip`
- timestamp: 2026-08-15T14:20:00Z
  observation: corrected scripted glue test, exact fresh-binary CLI test, focused certification
    package, and `certify-timing` all pass; one-pass regeneration and all drift checks are clean
- timestamp: 2026-08-15T14:47:00Z
  observation: hosted run 31890181420 passed the corrected certify-timing step, then the full suite
    exposed `TestPostgresManagedTargetMissingAndConsumedApprovalErrorsAreTyped`: branch-owned
    `GetReversePlan` returned `state path is required` for an explicitly in-memory App
- timestamp: 2026-08-15T15:05:00Z
  observation: the exact PostgreSQL typed-error test, related App authorization tests, installed-
    flow fresh-binary proof, focused lint, final regeneration, and drift checks all pass

# Eliminated

- The moved PostgreSQL base (#4167) did not cause this failure: the first failing command and all
  failed assertions are in this branch's corrected schedule-authorization proof surface.

The second failure is not eliminated as base-owned: #4168 changed `GetReversePlan` itself, and the
focused test reproduces the resulting early store-path error.

# Resolution

- root_cause: first, production removed the superseded per-schedule authorization carrier while
    certification still required it; second, branch-owned live plan reload was applied
    unconditionally to explicitly in-memory App fixtures
- fix: remove the legacy flag/reference; reject authority carriers in scripted scheduling; scan
    real schedule envelopes and crontab output while retaining direct-flow and cleanup assertions
- verification: first correction is green in local and hosted certify-timing; second correction and
    cleanup lint gap are green locally and pending a new hosted verify run
- files_changed: internal/connectors/certify/stages_glue.go,
    internal/connectors/certify/scripted_cli_test.go, internal/app/app.go, and GSD evidence
