# Issue #4176 TDD ledger

| Slice | Test class | Red evidence | Green evidence |
| --- | --- | --- | --- |
| Full stage-set inclusion | Happy path | `TestFullCertificationStageSetIsStrictSuperset` showed the audit's tail-stage reading does not reproduce on the dispatch base: the already-existing `runFullReadSweep` invokes flow and schedule per stream. The new test would fail if either execution path dropped a stage. | The test passes: Full contains every ordinary stage name, including `flow_roundtrip`, `schedule_roundtrip`, and `schedule_fire`, plus `full_sweep_connection_create_customers`. |
| Installed schedule fire | Happy path | `go test -timeout 20m ./internal/connectors/certify -run 'Test(FullCertificationStageSetIsStrictSuperset|GlueStagesScheduleFire)' -count=1 -v` failed before production edits because the report had no `schedule_fire` stage and the driver observed no command. | The same focused command passes after certification executes `pm schedule fire`, asserts `ScheduleFire`, flow `ok`, terminal status `succeeded`, and cleanup. `TestCertifyCLISingleConnectorPassExitsZero` also passed through the shipped CLI construction path. |
| Fire refusal | Bad path | The base had no fire stage to gate. | `TestGlueStagesScheduleFireRefusalFailsBeforeRemovalSuccess` sabotages the install response, asserts the exact `schedule_install did not pass` refusal, zero `schedule fire` invocations, a failed aggregate/report, and cleanup after the refusal. |
| Empty backend snapshot and real scheduler boundary | Edge case | The base reported schedule capability `pass` after only create/list/install/remove. | `TestGlueStagesScheduleFireEmptyBackendIsRestoredAndDaemonIsNotLive` proves the empty redirected backend is restored byte-for-byte and records `result=not_live` with the scheduler-daemon reason. |

## Test-contract mapping

- Happy: `TestFullCertificationStageSetIsStrictSuperset` compares actual report stage names, and `TestGlueStagesScheduleFireObservesInstalledFlowAndRestoresBackend` asserts `ScheduleFire`, terminal flow status, and backend cleanup.
- Bad: `TestGlueStagesScheduleFireRefusalFailsBeforeRemovalSuccess` asserts the exact named schedule fire/roundtrip failure and report failure.
- Edge: `TestGlueStagesScheduleFireEmptyBackendIsRestoredAndDaemonIsNotLive` covers the absent backend snapshot and explicit `not_live` scheduler-daemon classification.

No production edit may precede the red-test evidence. A failure is treated as correct unless disproven with recorded evidence; no golden or assertion will be relaxed to pass.
