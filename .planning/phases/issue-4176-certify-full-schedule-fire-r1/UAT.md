# UAT — Issue #4176

GSD `verify-work` was performed inline. The deliverables are fully automated backend assertions; no visual or subjective review is required.

| ID | Deliverable | Automated evidence | Result |
| --- | --- | --- | --- |
| D1 | `--full` is a strict stage-set superset | `TestFullCertificationStageSetIsStrictSuperset` compares actual Runner reports and requires full-only sweep coverage. | pass |
| D2 | Installed schedule execution is observed | `TestGlueStagesScheduleFireObservesInstalledFlowAndRestoresBackend` plus the real `TestCertifyCLISingleConnectorPassExitsZero` route require `ScheduleFire`, flow `ok`, schedule `succeeded`, and cleanup. | pass |
| D3 | Unavailable scheduler daemon is honest | `TestGlueStagesScheduleFireRefusalFailsBeforeRemovalSuccess` proves pre-fire refusal cannot pass; `TestGlueStagesScheduleFireEmptyBackendIsRestoredAndDaemonIsNotLive` proves the explicit `not_live` boundary and byte-for-byte empty cleanup. | pass |

Verdict: **verified.** The controlled direct-fire path is covered. A real scheduler daemon trigger remains deliberately and visibly `not_live`; no credentialed or operator scheduler was authorized for this worktree.
