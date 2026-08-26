# TDD ledger — #4354

| Slice | Red evidence | Green evidence | Refactor / notes |
| --- | --- | --- | --- |
| Source inventory and schema-v3 evidence | Pending: run existing source-evidence path against imported Outreach lock and declaration rows. A shared-reader failure is recorded as a non-local dependency, not fixed here. | Pending | Must preserve every operation row; never satisfy a check by downgrading availability. |
| Command preflight boundary | Pending: test valid ETL/write/delete command resolution without a credential and a wrong source identity/method/path binding. | Pending | Fixture-only; no provider I/O or reverse execution. |
| Artifact currentness | Pending: run existing validation/surface-sync/certification checks. | Pending | Generator changes are prohibited unless Outreach-only and already provided by the project. |
