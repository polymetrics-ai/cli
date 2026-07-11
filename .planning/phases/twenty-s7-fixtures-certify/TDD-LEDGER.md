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
- Blocker type: certify harness limitation outside S7 write scope; documented in `docs.md` and `VERIFICATION.md`.
