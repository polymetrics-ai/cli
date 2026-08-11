# Code review — Issue #3981 managed-target ownership and provisioning

## Method

The `code-review` command was resolved through `scripts/gsd sources` and its
prompt generated with `scripts/gsd prompt`. The canonical contract requires one
inline worker for this issue-local phase, so this is the documented manual
inline review rather than a skipped gate.

Reviewed the complete child diff against `feat/3972-postgres-parity`, the
truth-table state transitions, validation boundaries, native-lock lifecycle,
error rendering, and final static/race evidence.

## Findings and disposition

| Severity | Finding | Disposition |
| --- | --- | --- |
| Fixed before verdict | A provisioner-local mutex did not protect concurrent callers using distinct provisioner instances. | #4038 was created and linked from #3981 before the fix. The driver-neutral target-lock port is acquired before observation and held through reassertion; the fake proves every create held it and alternates two provisioners. |
| None remaining | Raw credential/display material could influence physical names or errors. | PASS: names use a length-delimited opaque hash of the source artifact triple/table; values have no display/credential field and driver errors are not rendered. |
| None remaining | A pre-existing, foreign, moved, replaced, missing, unreadable, or drifted target could be adopted/evolved. | PASS: the closed state machine only creates from absent namespace plus absent control; all other required states return typed refusal without a fake mutation. |
| None remaining | Cancellation or a same-process/cross-provisioner race could return a successful unasserted mutation. | PASS: context checks occur before lock, observe, create, and post-create result; target lock plus re-observation is covered by `-race -count=10`. |
| None remaining | Scope expansion into a native driver, generic SQL, DDL, transport, CDC, capability, CLI, help, docs, or website surface. | PASS: diff contains only the database shared contract/tests and planning evidence; scope scan found only comments referring to the deliberately absent ALTER path. |

## Verdict

**PASS.** No remaining actionable source, security, concurrency, or scope
finding. The review correction count is 1/5 (#4038); it is already verified by
the final local gates. Human review and merge remain required.
