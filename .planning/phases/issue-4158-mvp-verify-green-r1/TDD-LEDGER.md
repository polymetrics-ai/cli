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
| T1 | Reproduction | `go test -timeout 20m ./internal/cli -run '^TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip$' -count=2 -v` | **Red, repeatable:** two freshly built 152,132,226-byte binaries each exited 3 at `pm flow run`; aggregate run 46.778s. |
| T2 | Diagnostic | Token-checked temporary test instrumentation of the rejected flow command, reverted before this record | **Red:** typed `Error` envelope `validation/flow_job_reference_refused`; message names the empty action job as malformed. No credential material was logged. |
| T3 | Falsifier / divergent path | `git log -S'resolveManifestJobs'`; `git show 5c12fb536`; ancestry checks for `5aa5081af`, `779a2ee27` | **Disconfirmed shared root:** `#4168` introduced job resolution after both route-narrowing commits; `#4170` added the failing test afterwards. The test cannot run pre-#4150 because it did not yet exist. |
| T4 | Smallest counterfactual | Temporarily set only the flow action step's `job` to the produced reverse-plan ID and reduce its `action_cfg` to `read_back_stream`, then run the same fresh-binary test; restore original fixture | **Green:** 42.85s. Flow synced 1, action wrote 1, warehouse query returned 1, checkpoint and flow receipt persisted, replay/unapproved/auth/unsafe refusal checks all passed. |
| T5 | PostgreSQL scope check | `go test -tags=databaseintegration -timeout 20m ./internal/connectors/native/postgres -run '^TestPostgresManagedTargetDriverLiveControlAssertions$' -count=1 -v` | **Skipped as designed:** `POLYMETRICS_DATABASE_INTEGRATION=1` plus a direct runtime endpoint were absent. The GitHub reproduction does not instantiate a PostgreSQL route, so it is independently disproven as the shared cause. |
| T6 | TDD implementation | New happy/bad/edge tests | Blocked on firstmate's fixture-migration vs compatibility decision; no assertion was changed. |
