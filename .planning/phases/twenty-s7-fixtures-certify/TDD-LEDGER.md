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
