# TDD Ledger — issue #4158 / Production MVP verify green

## Test contract mapping

| Class | Planned separately named test | Observable assertion |
| --- | --- | --- |
| Happy path | `Test…ValidPostgresManagedTarget…DurableAcknowledgement` | The real managed-target construction path returns the durable acknowledgement / receipt, not only `err == nil`. |
| Bad path | `Test…NonPostgresHistoryRouteRefusesBeforeIO` | `errors.As` finds `*database.DatabaseWriteHistoryRouteError` with the intended reason; driver fake observes zero calls. |
| Edge case | `Test…<causal boundary>` | A cause-relevant boundary (chosen after reproducing) asserts exact acknowledgement/refusal and no extra I/O. |

## Evidence log

| ID | Stage | Command / action | Result |
| --- | --- | --- | --- |
| T0 | Plan | GSD command resolution + delivery artifacts | Green — recorded in `PLAN.md`; no production code touched. |
| T1 | Reproduction | Fresh-binary GitHub warehouse flow at `ef3c71caf` | Pending. |
| T2 | Reproduction | PostgreSQL managed-target live-control assertion at `ef3c71caf` | Pending. |
| T3 | Falsifier | Pre-#4150 / `#4150` / `#4155` causal comparison | Pending. |
| T4 | Red | New happy/bad/edge regression tests before production change | Pending. |
| T5 | Green | Minimal causal fix with focused tests | Pending. |
