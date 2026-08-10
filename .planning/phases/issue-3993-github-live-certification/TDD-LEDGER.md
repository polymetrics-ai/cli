# TDD ledger — Issue #3993

| Slice | Red evidence | Green evidence | Refactor / notes |
| --- | --- | --- | --- |
| Boundary parameterization | Pending: frozen owner/repository and old pre-skip policy are rejected by a deterministic script test. | Pending | Inputs must remain immutable and must not accept protected owners/repositories. |
| Concurrent whole-surface runner | Pending: test rejects sequential launch and terminal records without independent read-back. | Pending | Preserve per-operation cleanup and bounded terminal status. |
| Manifest / bootstrap integrity | Pending: current stale artifacts fail both `--check` commands. | Pending | Regenerate only source-derived artifacts; no hand-authored count. |
| Live provider evidence | Not a unit-test red step: real-provider result follows the green harness. | Pending | Secret-free report, read-backs, cleanup and residue scan are mandatory. |
| Warehouse inbound proof | Existing measured result will be re-run with a built binary. | Pending | Outbound write stays deferred to #3994/#3992. |

