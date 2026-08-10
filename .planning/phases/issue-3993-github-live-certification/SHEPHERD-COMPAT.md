---
phase: "3993"
mode: manual_compatibility
automatic_shepherd_approval: false
status: compatible_for_harness_gate_only
---

# Shepherd-compatible gate evidence — Issue #3993

This is **manual compatibility evidence, not an automatic Shepherd approval**.
Per the final-gate instruction, #3995 is not treated as integrated for this
child and no Shepherd driver, auto-loop trace, checkpoint, or validator verdict
was created. The delivery contract also prohibits spawning a Shepherd role.

The evidence below applies only to the #3993 harness slice. It does not approve
GitHub provider certification.

| Compatibility dimension | Ground-truth evidence | Manual result |
| --- | --- | --- |
| Correct stage | Five preserved task commits were rebased onto `origin/feat/3988-github-certification`; inline `verify-work` and standard `code-review` prompts/artifacts are current. | pass |
| Artifact validity | `UAT.md`, `VERIFICATION.md`, `REVIEW.md`, `RUN-STATE.json`, and `TDD-LEDGER.md` distinguish the harness pass from the 0/665 certification failure. | pass |
| Gates respected | No default-branch push/merge, live rerun, credential disclosure, provider write, #3990 coordinator, #3994 action path, or #3992 scheduler was added. | pass |
| Real progress | [#4020](https://github.com/polymetrics-ai/cli/issues/4020) records the RED/GREEN correction that closed the direct/config write-scope escape before process launch; this is correction loop 1 of 5. | pass |
| No hallucinated result | The report remains 665 attemptable / 856 blocked and 0 proven / 665 terminal-bound failures; successful isolated App and DuckDB/Parquet controls are stated separately. | pass |
| No scope conflict | The diff is limited to the GitHub harness, generated self-check artifacts, and issue-3993 evidence. The external admission and outbound paths are recorded as dependencies only. | pass |

## Compatibility verdict

**Proceed to local delivery gates for the #3993 harness slice.** This is a
manual compatibility conclusion, not a substitute for the future automatic
Shepherd validator.

The parent GitHub certification gate remains **blocked**, not approved: the
barrier measurement is still 0/665, #3990 owns coordinated admission policy,
767 mutations lack cleanup-safe fixtures, and #3994/#3992 own the outbound
warehouse-to-GitHub foundations.
