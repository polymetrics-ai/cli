---
coverage:
  - id: D1
    description: PostgreSQL to GitHub delivers one row through the warehouse and replays incrementally without duplication.
    verification:
      - kind: e2e
        ref: internal/cli/cross_system_live_certification_integration_test.go:TestPMBinaryExecutesLiveCrossSystemPipelines route 1
        status: pass
    human_judgment: false
  - id: D2
    description: GitHub to PostgreSQL full_overwrite replaces the complete destination on first run and replay.
    verification:
      - kind: e2e
        ref: internal/cli/cross_system_live_certification_integration_test.go:TestPMBinaryExecutesLiveCrossSystemPipelines route 2
        status: fail
    human_judgment: false
  - id: D3
    description: GitHub to GitHub updates one bounded issue comment through the warehouse and a typed approved action.
    verification:
      - kind: e2e
        ref: internal/cli/cross_system_live_certification_integration_test.go:TestPMBinaryExecutesLiveCrossSystemPipelines route 3
        status: pass
    human_judgment: false
  - id: D4
    description: The installed crontab entry point executes the persisted ETL and reverse-ETL flow and records terminal schedule state.
    verification:
      - kind: e2e
        ref: internal/cli/cross_system_live_certification_integration_test.go:TestPMBinaryExecutesLiveCrossSystemPipelines route 4
        status: pass
    human_judgment: false
  - id: D5
    description: Every task-created GitHub fixture is deleted and independently returns HTTP 404.
    verification:
      - kind: e2e
        ref: internal/cli/cross_system_live_certification_integration_test.go cleanup assertions
        status: pass
    human_judgment: false
---

# Summary — issue #4015 cross-system pipeline certification

## Outcome

Three routes are proven and one is broken with precise live evidence. PostgreSQL → GitHub, GitHub → GitHub, and persisted flow/schedule execution passed independent destination read-back, replay, and cleanup checks. GitHub → PostgreSQL delivered all 10 source labels correctly on its first run, but its `full_overwrite` replay completed `0/0` and replaced the target with an empty table instead of performing another full replacement.

This PR adds only an opt-in live characterization test and GSD evidence. It does not fix product code.

## Exact live samples

- PostgreSQL → GitHub: label `pm-cert-db-api-e10940f636b8`, exact count 1; replay `0/0` under `incremental_upsert`.
- GitHub → PostgreSQL: 10 source/target rows and the named label on first delivery; replay `0/0`, target count 0, named sample absent under `full_overwrite` (defect).
- GitHub → GitHub: comment `5328121289`, exact count 1, body `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru`; scheduled replay read/wrote `1/1` and remained singular.
- Schedule/flow: the task-local crontab payload invoked `pm --root <root> flow run pm-cert-cross-system-e10940f636b8 --json`; terminal state and prepared execution identity were recorded.
- Cleanup: the decisive label and comment independently returned HTTP 404, as did every diagnostic fixture created during the bounded run.

## Scale disposition

The optional 5 GB repeat was not attempted because route 2 failed correctness. No throughput or large-scale batching claim is made.

## Lifecycle

The issue-first discuss, TDD plan, execute, and verify-work workflows ran inline. Required Go skills are recorded in `PLAN.md`. CLI help/manual/website parity is not applicable because no command, flag, help, connector surface, documentation, or website behavior changed.
