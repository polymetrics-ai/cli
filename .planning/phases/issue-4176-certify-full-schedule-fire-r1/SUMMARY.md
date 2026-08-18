---
coverage:
  - id: D1
    description: Full certification retains every ordinary stage and adds full-only stream sweep coverage.
    requirement: issue-4176-full-superset
    verification:
      - kind: integration
        ref: internal/connectors/certify/stages_source_test.go — TestFullCertificationStageSetIsStrictSuperset
        status: pass
    human_judgment: false
  - id: D2
    description: Certification executes an installed schedule through ScheduleFire, observes terminal flow/status, and restores the backend.
    requirement: issue-4176-schedule-fire
    verification:
      - kind: integration
        ref: internal/connectors/certify/stages_glue_test.go — TestGlueStagesScheduleFireObservesInstalledFlowAndRestoresBackend
        status: pass
      - kind: unit
        ref: internal/connectors/certify/stages_glue_test.go — TestGlueStagesScheduleFireRefusalFailsBeforeRemovalSuccess
        status: pass
      - kind: e2e
        ref: internal/cli/certify_cli_test.go — TestCertifyCLISingleConnectorPassExitsZero
        status: pass
    human_judgment: false
  - id: D3
    description: The unstarted scheduler daemon is explicitly reported as not_live, while the empty isolated backend is restored byte-for-byte.
    requirement: issue-4176-not-live-boundary
    verification:
      - kind: integration
        ref: internal/connectors/certify/stages_glue_test.go — TestGlueStagesScheduleFireEmptyBackendIsRestoredAndDaemonIsNotLive
        status: pass
    human_judgment: false
---

# Summary — Issue #4176 certification schedule fire

## Delivered

- The full-stage audit claim was traced through the real runner: `runFullReadSweep` already invokes flow and schedule stages for each stream. `TestFullCertificationStageSetIsStrictSuperset` now prevents either path from silently dropping them.
- Certification now executes `pm schedule fire <name> --json` after installing the isolated crontab entry. It requires `ScheduleFire`, flow status `ok`, and terminal schedule status `succeeded` before backend removal.
- A failed install assertion refuses fire before its CLI call; the named fire and roundtrip stages fail, so successful cleanup cannot mask it.
- The report distinguishes this verified direct-fire path from a scheduler daemon trigger. The isolated crontab fixture never starts a daemon, so `Capabilities.Schedule.result` is explicitly `not_live` with a reason, never `pass` with a non-live excuse.

## Test-contract classes

- Happy: `TestFullCertificationStageSetIsStrictSuperset`; `TestGlueStagesScheduleFireObservesInstalledFlowAndRestoresBackend`.
- Bad: `TestGlueStagesScheduleFireRefusalFailsBeforeRemovalSuccess`.
- Edge: `TestGlueStagesScheduleFireEmptyBackendIsRestoredAndDaemonIsNotLive`.

## Verification

- `go test -timeout 20m ./internal/connectors/certify -count=1` passed.
- `go test -timeout 20m ./internal/cli -count=1` passed in 492.716s.
- Focused binary construction-path test `TestCertifyCLISingleConnectorPassExitsZero` passed in 30.602s and now asserts two `schedule_fire` stages in the full sample run.
- Formatting, vet, build, workflow/GSD, docs, lint, smoke, connector validation/surface sync/boundary, and release workflow checks passed; see `VERIFICATION.md`.

## Explicit non-live boundary

No real crontab/systemd/Temporal daemon or credentialed provider was activated in this worktree. The certification run does execute the product `schedule fire` command against its isolated installed backend entry, but does not claim a scheduler triggered that entry. That difference is represented in the persisted report as `not_live`.
