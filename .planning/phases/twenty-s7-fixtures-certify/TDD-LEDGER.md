# TDD ledger — Twenty S7 fixtures/certify

## Skills loaded
- `gsd-core`; `caveman`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-safety`; `golang-security`; `golang-documentation`; `golang-design-patterns`; `golang-structs-interfaces`.
- Missing required repo-local implementation skill: `.pi/skills/go-implementation/SKILL.md` returned `ENOENT`; fallback recorded in PLAN/RUN-STATE.
- Rule reminders applied: table/named test cases and behavior assertions (`golang-testing` best practices #1/#5), no secrets/trust-boundary checks (`golang-security` security-thinking questions 1-3), safe type/resource handling (`golang-safety` #2/#7), checked/wrapped errors/single handling (`golang-error-handling` #1/#2/#7), concise accurate docs/no invented context (`golang-documentation` writing principles), CLI stdout/stderr/help parity (`golang-cli` I/O patterns), simple explicit design/no new deps (`golang-design-patterns` #20/#21), schema/field-tag awareness (`golang-structs-interfaces` field tags section).

## GSD adapter evidence
- PASS: `scripts/gsd doctor`.
- PASS: `scripts/gsd list`.
- FAIL/adapter gap: `scripts/gsd prompt programming-loop init --phase twenty-s7-fixtures-certify --dry-run` -> `scripts/gsd: unknown GSD command: programming-loop`.
- Fallback prompt captured: `scripts/gsd prompt gsd-quick "twenty S7 fixtures docs conformance certify issue #284"`.
- Execution decision: `local_critical_path` (isolated worker cwd/branch; no recursive subagent tool in worker context).

## Red baseline
- Added `TestTwentyFixturesCoverAllStreamsAndWrites` to pin S7 fixture coverage across 28 streams and 112 write actions.
- Red run after fixing the test to load on-disk defs (fixtures/api_surface are intentionally not embedded): `go test ./internal/connectors/conformance -run TestTwentyFixturesCoverAllStreamsAndWrites -count=1` failed with 139 missing/failing checks (27 missing stream fixture checks + 112 write request-shape checks; attachments stream fixture already existed).

## Green evidence
- Generated synthetic fixtures for all 28 stream directories and all 112 write actions (`fixtures/writes/*.json`).
- `go test ./internal/connectors/conformance -run TestTwentyFixturesCoverAllStreamsAndWrites -count=1` — PASS.
- `go test ./internal/connectors/conformance -run 'TestConformance/twenty' -count=1` — PASS.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json` — PASS (`findings=0`, `warnings=0`).
- `make connectorgen-validate` — PASS (`548 connector(s) checked, 0 findings`).
- `go test ./...` — PASS.

## Certify blocker evidence
- Credential-free `pm connectors certify twenty` was attempted against a local localhost fixture server with a placeholder env var, never live Twenty.
- Non-full certify failed because certify creates the first connection with default cursor `updated_at` before reading catalog; Twenty's catalog cursor is `updatedAt`.
- Full certify with `--stream attachments --full --skip write` passed surface/read/sync/flow/schedule capabilities but failed on the longest stream name due `_pm_raw` filename length.
- Blocker type: certify harness limitation outside original S7 fixture write scope; VERIFY-TURN59 correction now grants minimal `internal/connectors/certify/**` write scope to fix it.

## Correction ledger — 2026-07-11 VERIFY-TURN59
- Required reread complete: `AGENTS.md`, issue #284 body, issue-agent contract, GSD universal runtime loop, required skills routing, GSD Pi adapter, CLI help/docs parity, S7 `PLAN.md`/`TDD-LEDGER.md`/`VERIFICATION.md`/`RUN-STATE.json`/`SUMMARY.md`, connector architecture/migration/certification docs, worker handoff template.
- GSD adapter: `scripts/gsd doctor` PASS; `scripts/gsd list` PASS; `scripts/gsd prompt programming-loop init --phase twenty-s7-fixtures-certify --dry-run` FAIL (`scripts/gsd: unknown GSD command: programming-loop`); fallback prompt captured with `scripts/gsd prompt gsd-quick "twenty S7 certify correction issue #284"`.
- Skills loaded: repo `gsd-core`, `caveman`; local `.pi/skills/go-implementation/SKILL.md` missing; read fallback `/Users/karthiksivadas/Development/polymetrics-cli-agents/wt-twenty-run/.pi/skills/go-implementation/SKILL.md` and `references/go-rules.md`; global Go skills `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`.
- Rule notes for correction: go-implementation/go-rules #1/#2/#5/#7 (wrap/check/lowercase errors), #15/#37/#38 (zero values and JSON contracts), #40-#43 (CLI stdout/stderr/JSON/help contracts); `golang-testing` #1/#5 (named behavior tests); `golang-security` trust-boundary questions and filesystem/path traversal guidance; `golang-safety` #7 resource cleanup; `golang-design-patterns` #20/#21 no new deps/testability.
- Red validation before production edits:
  - `./pm help connectors certify` — exit 0, help/manual rendered; no credentials read.
  - `PM_TWENTY_CERT_TOKEN=fixture_token_placeholder ./pm connectors certify twenty --stream attachments --limit 1 --config base_url=http://127.0.0.1:57548 --from-env api_key=PM_TWENTY_CERT_TOKEN --keep-workdir --json` — exit 2; `etl_full_refresh_append` failed with `kind got="Error" want="ETLRun", exit got=1 want=0`.
  - Kept workdir manual `./pm etl run --connection cert_live --stream attachments --json --root <kept-workdir>` — exit 1; JSON error message `record is missing cursor field "updated_at"`.
  - `PM_TWENTY_CERT_TOKEN=fixture_token_placeholder ./pm connectors certify twenty --stream attachments --full --skip write --limit 1 --config base_url=http://127.0.0.1:57548 --from-env api_key=PM_TWENTY_CERT_TOKEN --keep-workdir --json` — exit 2; current full run still blocked by same cursor failure before long-path assertion can execute.
- Planned red tests: focused certify unit tests for inspect-seeded cursor metadata and bounded raw path component names for `message_channel_message_association_message_folders`.
- Red test run: `go test ./internal/connectors/certify -run 'TestInspectStreamSpecsSeedBootstrapCursor|TestFullSweepArtifactNamesAreBoundedForLongStream' -count=1` — FAIL/build failed with `undefined: streamSpecsFromInspectEnvelope`, `rc.captureFileName undefined`, `rc.liveTableName undefined`, `rc.captureTableName undefined`.
- Green implementation: `stageManualJSON` seeds `catalogStreamSpecs` from `connectors inspect --json` `manifest.streams` before bootstrap `connection_create`; `safeName` now bounds stream-derived path/name components with an 8-hex deterministic hash suffix; full-sweep live/capture/incremental table and capture file names use bounded helpers.
- Green tests:
  - `gofmt -l cmd internal && go test ./internal/connectors/certify -run 'TestInspectStreamSpecsSeedBootstrapCursor|TestFullSweepArtifactNamesAreBoundedForLongStream' -count=1` — PASS: `ok polymetrics.ai/internal/connectors/certify 0.534s`.
  - `go test ./internal/connectors/certify -run 'TestInspectStreamSpecsSeedBootstrapCursor|TestFullSweepArtifactNamesAreBoundedForLongStream|TestCatalogStreamSpecsFromStreams|TestFullSweepNamesAreStreamScoped|TestFullSweepSourceStagesAgainstSample|TestSourceStagesAgainstSample' -count=1` — PASS: `ok polymetrics.ai/internal/connectors/certify 46.341s`.
  - Final focused rerun `go test ./internal/connectors/certify -run 'TestInspectStreamSpecsSeedBootstrapCursor|TestFullSweepArtifactNamesAreBoundedForLongStream' -count=1` — PASS: `ok polymetrics.ai/internal/connectors/certify 0.525s`.
- Corrective certify gates with `/tmp/twenty_fixture_all_server.py` serving Twenty checked-in stream fixtures on localhost, placeholder env only:
  - Non-full command exited 0 with `.report.passed=true`.
  - Full command `--full --skip write` exited 0 with `.report.passed=true`.

## Correction ledger — 2026-07-12 REVIEW-TURN62 / CORRECT-TURN63
- Required reread complete: dispatch prompt, `AGENTS.md`, issue #284 body/acceptance, issue-agent contract, GSD universal runtime loop, GSD Pi adapter, required-skills routing, CLI help/docs/website parity, connector validation gates, S7 `PLAN.md`/`TDD-LEDGER.md`/`VERIFICATION.md`/`RUN-STATE.json`/`SUMMARY.md`, connector migration conventions/design docs, REVIEW-TURN62/VERIFY-TURN61, worker handoff template.
- GSD adapter: `scripts/gsd doctor` PASS; `scripts/gsd prompt programming-loop init --phase twenty-s7-fixtures-certify --dry-run` FAIL (`scripts/gsd: unknown GSD command: programming-loop`); fallback prompt captured with `scripts/gsd prompt gsd-quick "twenty S7 review correction F1 F2 F3 F4 issue #284"`.
- Skills loaded: `gsd-core`; `caveman`; go-implementation fallback and `go-rules.md`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-design-patterns`; `golang-structs-interfaces`; `golang-context`; `golang-concurrency`; `golang-documentation`.
- Rule notes for this correction: go-rules #1/#2/#5/#7 (checked/wrapped/lowercase errors), #15/#37/#38 (zero-values and JSON contracts), #40-#43 (CLI stdout/stderr/JSON/help); `golang-testing` #1/#5 (named behavior tests); `golang-security` trust-boundary/no-secrets/path-safety; `golang-safety` #2/#7; `golang-design-patterns` #20/#21.
- Planned red tests before production edits:
  - F1: certify stream default picks first cursor stream from Twenty inspect/catalog specs when `Options.Stream` is empty.
  - F2: conformance write fixture comparison fails without exact top-level array/no-body support and Twenty batch/delete fixtures prove exact request shape.
  - F3: certify `fixture_conformance` runs and passes for Twenty instead of stale unconditional skip.
- Red evidence captured before production edits:
  - `./pm help connectors certify >/tmp/twenty-s7-turn63-help.txt && echo 'help exit=0'` — PASS/exit 0; no credentials read.
  - `go test ./internal/connectors/certify -run 'TestStreamNameDefaultsToFirstCatalogCursorStream|TestStreamNameDefaultsToFirstCatalogStreamWhenNoCursor|TestFixtureConformanceRunsForTwentyBundle' -count=1` — FAIL: Twenty no-stream default got `"customers"`; `fixture_conformance` returned stale no-defs skip.
  - `go test ./internal/connectors/conformance -run 'TestCompareWriteExpectationSupportsExactTopLevelArrayBody|TestCompareWriteExpectationSupportsExplicitNoBody|TestCompareWriteExpectationPreservesSubsetBody|TestTwentyBatchAndDeleteFixturesAssertExactBodies' -count=1` — FAIL/build failed: missing `BodyPresent`, `BodyExact`, and `NoBody` support.
- Green implementation:
  - F1: `streamName` now chooses first known catalog/manual stream with `CursorField`, else first known stream; hardcoded `issues`/`customers` fallback only when specs are unavailable.
  - F2: conformance supports `expect.body_exact` exact decoded JSON bodies (including top-level arrays) and `expect.no_body`; legacy `expect.body` remains subset-map matching. Twenty `batch_*` fixtures assert top-level arrays; `delete_*` fixtures assert `no_body=true`.
  - F3: `fixture_conformance` loads on-disk defs fixtures when present and runs `conformance.RunBundle`; missing bundle/fixtures are truthful skips; real conformance failures fail `Report.Passed`.
  - F4: Twenty docs remove stale current-limit notes for `updatedAt` cursor / long stream names; live certification credential gate, local fixture conformance, and reverse ETL plan → preview → approval → execute remain documented.
- Green evidence:
  - `go test ./internal/connectors/certify -run 'TestStreamNameDefaultsToFirstCatalogCursorStream|TestStreamNameDefaultsToFirstCatalogStreamWhenNoCursor|TestFixtureConformanceRunsForTwentyBundle|TestFixtureConformanceFailureFailsReport|TestInspectStreamSpecsSeedBootstrapCursor|TestFullSweepArtifactNamesAreBoundedForLongStream|TestCatalogStreamSpecsFromStreams|TestFullSweepNamesAreStreamScoped|TestFullSweepSourceStagesAgainstSample|TestSourceStagesAgainstSample' -count=1` — PASS: `ok polymetrics.ai/internal/connectors/certify 44.503s`.
  - `go test ./internal/connectors/conformance -run 'TestCompareWriteExpectationSupportsExactTopLevelArrayBody|TestCompareWriteExpectationSupportsExplicitNoBody|TestCompareWriteExpectationPreservesSubsetBody|TestTwentyBatchAndDeleteFixturesAssertExactBodies|TestConformance/twenty|TestTwentyFixturesCoverAllStreamsAndWrites|TestWriteRequestShape_MatchesExpectBlock|TestWriteRequestShape_MismatchFails' -count=1` — PASS: `ok polymetrics.ai/internal/connectors/conformance 1.183s`.
  - `go test ./internal/connectors/conformance -count=1` — PASS: `ok polymetrics.ai/internal/connectors/conformance 9.392s`.
  - `go test ./internal/connectors/certify -count=1` — PASS: `ok polymetrics.ai/internal/connectors/certify 334.029s`.
  - `go test ./...` — PASS (slow packages included `internal/cli 151.975s`, `internal/connectors/certify 340.621s`).
  - JSON parse for edited Twenty write fixtures and phase JSON — PASS.
  - `go run ./cmd/connectorgen validate internal/connectors/defs --json` — PASS: `findings=[]`, `warnings=[]`.
  - `go vet ./...` — PASS/no output; `go build ./cmd/pm` — PASS/no output; `gofmt -l cmd internal` clean.
  - Fresh `/tmp/twenty-s7-turn63-pm help connectors certify` — PASS/exit 0.
  - Credential-free localhost `pm connectors certify twenty` without `--stream` — PASS/exit 0, `.report.passed=true`, `fixture_conformance.passed=true`.
  - Credential-free localhost `pm connectors certify twenty --full --skip write` without `--stream` — PASS/exit 0, `.report.passed=true`, `fixture_conformance.passed=true`.
  - `cd website && pnpm run gen:website-data` rerun idempotency — PASS; `git diff --check` — PASS/no output.
- Not run: `make verify` because `verify` includes `smoke-no-build`, which executes `./pm reverse run`; this correction forbids reverse ETL/destructive execution.
- Safety: no `TWENTY_API_KEY`, no live `api.twenty.com`, placeholder env only for localhost certify, no reverse ETL/destructive external execution, no new deps.

## Correction ledger — 2026-07-12 REVIEW-TURN65 / CORRECT-TURN66
- Required reread complete: dispatch prompt, `AGENTS.md`, issue #284 body/acceptance, issue-agent contract, GSD universal runtime loop, GSD Pi adapter, required-skills routing, CLI help/docs/website parity, connector validation gates, S7 `PLAN.md`/`TDD-LEDGER.md`/`VERIFICATION.md`/`RUN-STATE.json`/`SUMMARY.md`, connector migration docs/design, REVIEW-TURN65/VERIFY-TURN64, worker handoff template.
- GSD adapter: `scripts/gsd doctor` PASS; `scripts/gsd prompt programming-loop init --phase twenty-s7-fixtures-certify --dry-run` FAIL (`scripts/gsd: unknown GSD command: programming-loop`); manual fallback prompt captured with `scripts/gsd prompt gsd-quick "twenty S7 review correction TURN66 limit cap issue #284"`.
- Skills loaded: `gsd-core`; `caveman`; go-implementation fallback and `go-rules.md`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-design-patterns`; `golang-structs-interfaces`; `golang-context`; `golang-concurrency`; `golang-documentation`; `golang-lint`.
- Rule notes: go-rules #1/#2/#5/#7 (checked/wrapped/lowercase errors), #15/#37/#38 (zero values and JSON contracts), #40-#43 (CLI stdout/stderr/JSON/help); `golang-testing` #1/#5 (named behavior tests); `golang-security` trust-boundary/no-secrets; `golang-safety` #2/#7; `golang-design-patterns` #20/#21 no new deps/testability; `golang-context` #1/#2 propagation; `golang-concurrency` #1/#7 cancellation-aware loops.
- Planned red tests before production edits:
  - app warehouse ETL: `RunETLRequest{Limit: 2}` caps source reads/loaded rows and forwards `ReadRequest.Limit`.
  - app connector-destination ETL: same cap with non-warehouse destination and bounded batch writes.
  - certify: live ETL stage argv for `--limit 7` includes `--limit 7` for `etl_full_refresh_append`, `etl_incremental_append`, and `resume`.
- CLI parity note: no public help/docs change planned; `pm etl run --limit` is hidden/internal pass-through for certify safety, while `pm connectors certify --limit` remains documented/exposed.
- Red evidence captured before production edits:
  - `go test ./internal/app -run 'TestRunETLLimitCapsWarehouseRead|TestRunETLLimitCapsConnectorDestinationRead' -count=1` — FAIL/build failed: `unknown field Limit in struct literal of type RunETLRequest`.
  - `go test ./internal/connectors/certify -run TestSourceStagesPassLimitToLiveETLRuns -count=1` — FAIL: `etl_full_refresh_append argv = "pm etl run --connection cert_live --stream customers --json --root ...", want --limit 1`.
- Green implementation:
  - `RunETLRequest.Limit` carries an optional cap through `pm etl run --limit` into app ETL.
  - Warehouse and connector-destination ETL reads pass `ReadRequest.Limit`, wrap emitters with `connectors.LimitEmitter`, and treat `ErrReadLimitReached` as successful early stop via `connectors.IgnoreReadLimit` before final flush/checkpoint.
  - Certify live ETL stages (`etl_full_refresh_append`, `etl_incremental_append`, `resume`) invoke `pm etl run ... --limit <Options.Limit>`; capture replay stages keep replaying the bounded capture.
- Green evidence:
  - `gofmt -l cmd internal` — PASS/clean.
  - `go test ./internal/app -run 'TestRunETLLimitCapsWarehouseRead|TestRunETLLimitCapsConnectorDestinationRead' -count=1` — PASS: `ok polymetrics.ai/internal/app 2.394s`.
  - `go test ./internal/connectors/certify -run TestSourceStagesPassLimitToLiveETLRuns -count=1` — PASS: `ok polymetrics.ai/internal/connectors/certify 16.365s`.
  - `go test ./internal/app -run 'TestRunETLLimitCapsWarehouseRead|TestRunETLLimitCapsConnectorDestinationRead|TestRunETLWritesBoundedBatches' -count=1` — PASS: `ok polymetrics.ai/internal/app 3.461s`.
  - `go test ./internal/connectors/certify -run 'TestSourceStagesPassLimitToLiveETLRuns|TestSourceStagesAgainstSample' -count=1` — PASS: `ok polymetrics.ai/internal/connectors/certify 31.688s`.
  - `go test ./internal/app -count=1` — PASS: `ok polymetrics.ai/internal/app 17.506s`.
  - `go run ./cmd/connectorgen validate internal/connectors/defs --json` — PASS: `findings=[]`, `warnings=[]`, `connectors_checked=548`.
  - `go vet ./...` — PASS/no output.
  - `go build ./cmd/pm` — PASS/no output.
  - `go test ./...` — PASS (slow packages: `internal/cli 153.828s`, `internal/connectors/certify 362.625s`, `internal/app 17.344s`).
  - `./pm help connectors certify` — PASS/exit 0 before localhost certify.
  - Credential-free localhost `pm connectors certify twenty --limit 1` with placeholder env only — PASS: `.report.passed=true`, `read_records=1`, `stream=attachments`.
  - `git diff --check` — PASS/no output; phase `RUN-STATE.json` parsed with `jq`.
- Not run: `make verify` because `verify` includes `smoke-no-build`, which executes `./pm reverse run`; this correction forbids reverse ETL/destructive execution.
- Safety: no `TWENTY_API_KEY`, no live `api.twenty.com`, placeholder env only for localhost certify, no reverse ETL/destructive external execution, no new deps.
