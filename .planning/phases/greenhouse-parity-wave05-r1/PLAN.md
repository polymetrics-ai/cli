# Greenhouse parity wave05 r1 plan

## Scope

Parent issue: #3199. Children: #3200-#3206. Connector: `greenhouse`.

Work is constrained to Greenhouse-owned connector definitions/fixtures/docs/generated surfaces plus this GSD artifact. No live Greenhouse calls, no credentials, no provider writes, no shared runtime behavior changes, no generic HTTP/query/body/file/shell escape hatches, no push/PR/no-mistakes pipeline.

## GSD path and skills

- `scripts/gsd doctor`: passed.
- Required command attempted: `scripts/gsd prompt programming-loop init --phase issue-3199-greenhouse-parity --dry-run`; adapter reported `unknown GSD command: programming-loop`.
- Manual GSD fallback recorded per adapter rules; `scripts/gsd prompt quick ... --dry-run` was generated and followed inline.
- Skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-spf13-cobra`; CLI parity reference loaded.

## Current evidence before production edits

- Official parent allocation from #3199: total 134 = 68 ETL reads + 1 CDC/changefeed + 63 reverse writes + 2 binary-file operations; target post-change disposition 126 implemented/fixture-tested, 8 blocked/planned, 0 excluded, 0 certified.
- Current bundle inventory: 69 streams, 57 write actions, api_surface rows 129 = 69 stream-covered + 57 write-covered + 3 deprecated exclusions.
- Fresh official Harvest HTML fetched from `https://developers.greenhouse.io/harvest.html` (ETag/last-modified per issue) and parsed for documented operations. Delta to resolve:
  - Add documented v2 users operations: `PATCH /v2/users`, `/v2/users/disable`, `/v2/users/enable`.
  - Add documented v2 job-post operations: `PATCH /v2/job_posts/{id}`, `/v2/job_posts/{id}/status`.
  - Add destructive async candidate-tag deletion ledger row: `DELETE /tags/candidate/{tag id}`.
  - Correct hiring-team write paths from current `/jobs/{id}/hiring_team` rows to official `/jobs/{id}` rows.
  - Convert existing deprecated exclusions into blocked/planned operation rows because parent count has excluded=0.
  - Reconcile binary attachment operations against the parent direct/binary lane and existing write contract truthfully.

## Implementation plan

1. Inventory and baseline validation
   - Run focused `connectorgen validate`, conformance, CLI/golden smoke if useful to capture current failures.
   - Parse official docs into a connector-local parity ledger summary.
2. Definition updates
   - Update `api_surface.json` to enumerate every official operation exactly once with source-linked coverage/blocker rows and post-change counts matching the parent where supported by docs.
   - Add fixture-backed typed writes for documented supported operations that fit the existing JSON write contract.
   - For unsupported binary/asynchronous/deprecated/destructive operations, use `operation` blocked/planned rows with exact official-source evidence and no generic escape hatch.
   - Fix Greenhouse hiring-team write paths and fixtures to match official Harvest docs.
   - Tighten destructive write schemas where touched; avoid arbitrary request bodies.
3. Fixtures/docs/generated surfaces
   - Add/update write fixtures for every new/changed action.
   - Update `docs.md` and generated Greenhouse `docs/connectors/greenhouse/{MANUAL.md,SKILL.md}` plus connector catalog JSON if generation/update is available.
4. Verification
   - Focused: `go run ./cmd/connectorgen validate internal/connectors/defs/greenhouse`; `go test ./internal/connectors/conformance -run 'TestConformance/greenhouse' -count=1`; `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`; `go vet ./internal/connectors/... ./internal/cli/...`; `go build ./cmd/pm`; `make connector-boundary`; `git diff --check`.
   - Broad: `make verify`.
5. GitHub and commit
   - Update parent and child issues once via `gh-axi` with truthful counts and verification evidence.
   - Commit clean branch. Do not push or open/update PR.

## Safety checkpoints

- No secret values in docs, fixtures, errors, issue comments, or status.
- No live provider calls or credentialed checks.
- Destructive actions remain approval-gated and schema-bounded.
- Binary uploads/downloads stay blocked unless a reviewed bounded transfer contract exists.
