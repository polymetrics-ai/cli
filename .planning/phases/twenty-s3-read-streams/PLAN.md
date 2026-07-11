# Twenty S3 read streams — PLAN

Issue: #280
Parent: #277
PR: #290
Branch: `feat/280-twenty-streams`
Base: `feat/277-twenty-connector-parity` @ `b4895064`

## Turn36 correction plan

- Operator decision: Option 1 — S3 owns the minimal fixtures for streams it declares; no quality-gate exception. S7 #284 refines/expands fixtures later.
- Restore `internal/connectors/defs/twenty/api_surface.json` to full S3 read coverage: 56 GET rows, 56 `covered_by.stream`, 0 `excluded`, 0 `covered_by.direct_read`.
- Map both `/rest/<object>` and `/rest/<object>/{id}` to the same existing snake_case stream.
- Keep the existing 28 snake_case streams in `streams.json`; edit only if validation proves inconsistency.
- Keep minimal `internal/connectors/defs/twenty/fixtures/streams/**` fixtures because conformance binds when streams are declared.
- Remove unowned `internal/connectors/engine/twenty_bundle_test.go`.
- Do not edit schemas, writes, CLI surface, docs/website, engine/source beyond deleting the unowned test, go.mod/go.sum, or credentialed/live/reverse-ETL paths.

## Review F1 correction plan (2026-07-11)

- Accepted finding F1: `streams.json` declares cursor `pagination.limit_param: "limit"`, but the engine only consumes `limit_param` for `offset_limit`; current cursor reads omit the intended `limit=60`.
- Scope: update only `internal/connectors/defs/twenty/streams.json`, `internal/connectors/defs/twenty/fixtures/streams/attachments/page_1.json`, `internal/connectors/defs/twenty/fixtures/streams/attachments/page_2.json`, and this phase's GSD evidence files.
- For all 28 streams, remove the no-op `pagination.limit_param` and add static `query: {"limit": "60"}` while preserving cursor pagination fields `type`, `cursor_param`, `token_path`, and `stop_path`.
- Update the minimal attachments fixtures so page 1 expects `{ "limit": "60" }` and page 2 expects `{ "limit": "60", "starting_after": "attachment_fixture_cursor_1" }`.
- Red validation before production edit: run the review-fix Python invariant and capture its failure against current head `0c3b2c97`.
- Green gates: run the exact required command block from the review-fix task, then commit `fix(twenty): send bounded stream page size (Refs #280)` and push `origin feat/280-twenty-streams`.
- Do not edit schemas, writes, CLI surface, docs.md/docs/website, Go source, go.mod/go.sum, or secrets.

## GSD / execution mode

- GSD adapter: repo-local Pi (`scripts/gsd`).
- Review-fix commands run this session: `scripts/gsd doctor`; `scripts/gsd list`; `scripts/gsd prompt programming-loop init --phase twenty-s3-read-streams --dry-run` (failed: `unknown GSD command: programming-loop`); `scripts/gsd prompt quick --validate "S3 #280 PR #290 review-fix F1: remove no-op cursor limit_param and add static stream query limit=60 for Twenty streams"` (adapter prompt generated).
- Manual-GSD fallback: active because `programming-loop` is not exposed in this adapter; equivalent PLAN / TDD-LEDGER / VERIFICATION artifacts maintained here.
- Spawn decision: `local_critical_path` — user explicitly said do not spawn subagents; this worker owns exactly this issue/branch/cwd.
- Required skills loaded: `gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-lint`.
- Required patch note: `.planning/auto-loop/tasks/S3/full-scope-correction-uncommitted.patch` was requested as reading input but is absent in this checkout (`ENOENT`); correction proceeds from the issue comments/operator contract and current branch state.

## TDD / validation slices

1. RED validation: prove current branch is invalid with api_surface coverage counts and the unowned engine test path still present.
2. GREEN slice: rewrite only `api_surface.json` endpoint coverage/explanatory scope and delete `internal/connectors/engine/twenty_bundle_test.go`.
3. Focused gates: JSON parse/count assertions, `connectorgen validate`, Twenty conformance, GSD workflow verification, package tests.
4. Broad gates: `gofmt -w cmd internal`, `go vet ./...`, `go test ./... -count=1`, `go build ./cmd/pm`, `gofmt -l cmd internal`; do not run `make verify` because this task forbids reverse ETL execution and `make verify` depends on `smoke` -> `./pm reverse run`.
5. Update verification artifacts with exact command results, commit `fix(twenty): restore full S3 read coverage (Refs #280)`, push `origin feat/280-twenty-streams`.
6. Review F1 slice: capture red invariant for missing static stream `limit=60`, update `streams.json` plus attachments fixtures, run required focused gates and `go test ./... -count=1` if time allows, then commit `fix(twenty): send bounded stream page size (Refs #280)` and push `origin feat/280-twenty-streams`.

## CLI help/docs/website parity

Connector surface changes only. S6 #283 owns CLI surface/help/docs/website. For this S3 correction: runtime CLI help, docs/cli, website, and generated help/manual artifacts are not applicable; no CLI command behavior changes.

## Safety notes

- No secrets requested, printed, stored, or summarized.
- No credentialed Twenty checks.
- No new dependencies.
- No reverse ETL execution.
- Parent PR #285 merge to `main` remains human-gated.
