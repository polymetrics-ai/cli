# Twenty S7 fixtures + docs.md + conformance/certify plan

Issue: #284 (`twenty S7: fixtures + docs.md + conformance & certify`). Parent: #277 / #285. Branch: `feat/284-twenty-fixtures-certify` from parent head `23c2277d827da31e1c1629793c830a129eeb5e7e`.

## Issue contract / acceptance
- Scope from issue #284: `fixtures/**`, `docs.md`; dependencies #279-#283; refs parent #277.
- Deliver fixtures for streams + writes, `docs.md`, `pm connectors certify twenty`, parity-deviation ledger entries for documented gaps (target: none).
- Acceptance requested: `make connectorgen-validate`, `make verify`, focused tests, `pm connectors certify twenty` green.
- Current run constraint: no PR open in worker; push branch only. No live Twenty API or credentials.

## Required skills / workflows
- Skills loaded: `gsd-core`, `caveman`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-safety`, `golang-security`, `golang-documentation`, `golang-design-patterns`, `golang-structs-interfaces`.
- Required Go implementation skill path `.pi/skills/go-implementation/SKILL.md` is absent in this worktree (`ENOENT`); fallback is required-skills routing plus loaded Go skill set above.
- GSD adapter: `scripts/gsd doctor` passed; `scripts/gsd prompt programming-loop init --phase twenty-s7-fixtures-certify --dry-run` failed (`scripts/gsd: unknown GSD command: programming-loop`); fallback prompt captured with `scripts/gsd prompt gsd-quick "twenty S7 fixtures docs conformance certify issue #284"`.
- Execution decision for this worker cycle: `local_critical_path` because this is an already-isolated worker cwd/branch and Pi worker has no recursive `subagent` tool.
- No secrets, no live Twenty API, no reverse ETL execution, no new dependencies.

## Scope
- Add credential-free synthetic stream fixtures under `internal/connectors/defs/twenty/fixtures/streams/**` for the full 28-object read surface.
- Add synthetic write fixtures under `internal/connectors/defs/twenty/fixtures/writes/**` for create/update/batch/delete request-shape conformance, without executing live writes.
- Do not add live `fixtures/check.json` unless replay can be proven credential-free and non-live.
- Refresh `internal/connectors/defs/twenty/docs.md` with S7 fixture/certification notes and parity-deviation ledger.
- Add focused conformance coverage assertion in `internal/connectors/conformance/twenty_test.go` only for S7 fixture coverage.
- Commit generated website connector data only because `docs.md` is embedded there by the existing generator output.

## TDD slices
1. Red baseline already captured: focused coverage test failed before fixtures existed (139 missing/failing checks: 27 missing stream fixture checks + 112 write request-shape checks).
2. Green fixture slice: generated deterministic stream/write fixtures from `streams.json`, `schemas/*.json`, and `writes.json`.
3. Verification slice: run JSON parse, `connectorgen validate`, focused conformance/Twenty coverage, `go test ./...`, `go vet ./...`, `go build ./cmd/pm`, `pm connectors certify twenty` credential-free behavior, website data idempotency, then update artifacts.
4. Commit/push slice: commit coherent S7 result and push `origin/feat/284-twenty-fixtures-certify`; do not open PR.

## Collision rules
- Original S7 scope touched only `internal/connectors/defs/twenty/fixtures/**`, `internal/connectors/defs/twenty/docs.md`, `internal/connectors/conformance/twenty_test.go`, S7 phase artifacts, and generated website/catalog data if required by generator/idempotency.
- Correction scope for VERIFY-TURN59 is explicitly narrowed to `internal/connectors/certify/**` focused harness fixes/tests plus these S7 phase artifacts. Do not touch connector defs unless the certify fix proves impossible without it.
- Do not edit `api_surface.json`, `streams.json`, `writes.json`, `schemas/**`, unrelated CLI/code surfaces, `go.mod`/`go.sum`, other connectors, `.github/**`, `main`, parent PR/roadmap/state files, or coordinator `.planning/auto-loop/**` files.

## Correction plan — 2026-07-11 VERIFY-TURN59
- Objective: make credential-free localhost `pm connectors certify twenty` green for PR #322 head beyond `2d909649899b0c26d0e9822af70232d4af26e89d`; commit/push only `feat/284-twenty-fixtures-certify`; stop at correction handoff.
- Red evidence captured before production edits:
  - `./pm help connectors certify` rendered connector manual/help, exit 0.
  - Non-full localhost certify with placeholder env exited 2; manual rerun of kept workdir ETL returned `record is missing cursor field "updated_at"`.
  - Full localhost certify with `--full --skip write` exited 2; current run is still dominated by bootstrap cursor failure before reaching the known long-name raw-path failure.
- Slice 1 TDD: add focused certify tests for pre-bootstrap stream metadata cursor selection from connector inspect/catalog shape (`updatedAt`, not fallback `updated_at`) and for bounded deterministic certify names/path components for long Twenty stream names.
- Slice 1 implementation: seed `catalogStreamSpecs` from `connectors inspect --json` before `connection_create`; use bounded filesystem-safe names for full-sweep connection/table/capture artifacts without changing CLI stdout/stderr contracts.
- Gates: focused new tests; `go test ./internal/connectors/certify -run '<new-or-focused-tests>' -count=1`; conformance Twenty focused tests; `connectorgen validate`; `go vet ./...`; `go build ./cmd/pm`; `gofmt -l cmd internal`; credential-free non-full and `--full --skip write` localhost certify. `make verify` only if safe (otherwise record reverse ETL safety blocker).
- Execution decision: `local_critical_path` — this Pi worker is already in isolated issue worktree and has no recursive `subagent` tool.

## Correction plan — 2026-07-12 REVIEW-TURN62 / CORRECT-TURN63
- Objective: fix PR #322 review findings F1-F4 at head `4b06439d36ac44f7ae506ad6c12358bd6a748d9e`, commit one green correction slice, push only `origin/feat/284-twenty-fixtures-certify`, stop before verify/review/integrate.
- GSD adapter: `scripts/gsd doctor` PASS; `scripts/gsd prompt programming-loop init --phase twenty-s7-fixtures-certify --dry-run` still unavailable (`unknown GSD command: programming-loop`); fallback prompt captured with `scripts/gsd prompt gsd-quick "twenty S7 review correction F1 F2 F3 F4 issue #284"`.
- Skills loaded: `gsd-core`, `caveman`, go-implementation fallback from coordinator worktree plus `references/go-rules.md`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`.
- Rule notes: go-rules #1/#2/#5/#7 for wrapped/checked/lowercase errors; #15/#37/#38 for zero-value/JSON contracts; #40-#43 for CLI JSON/help/exit behavior; `golang-testing` #1/#5 for named observable behavior tests; `golang-security` trust-boundary and filesystem/no-secret guidance; `golang-safety` #2/#7 for safe assertions/resource cleanup; `golang-design-patterns` #20/#21 no new deps and testability.
- Slice F1: default no-`--stream` certify stream from seeded catalog/manual stream specs: first cursor stream, else first stream; use hardcoded connector fallback only if specs unavailable. Add Twenty no-`--stream` regression.
- Slice F2: extend conformance write fixtures to distinguish subset object `body` from exact decoded `body_exact` (array/object/scalar) and `no_body`; update Twenty `batch_*` fixtures to assert top-level arrays and `delete_*` fixtures to assert no body. Add regression tests.
- Slice F3: `fixture_conformance` stage loads defs and runs conformance when bundle fixtures exist (Twenty included); skip only missing bundle/fixtures with truthful reason; conformance failures fail certify.
- Slice F4: refresh `internal/connectors/defs/twenty/docs.md` to remove resolved `updatedAt`/long-name limitations; keep credential-gated live certification, local fixture conformance, and reverse ETL plan → preview → approval → execute caveats.
- Red evidence plan: run focused tests for F1-F3 before production edits and record failures, plus `./pm help connectors certify` before any `pm connectors certify` command.
- Green gates: `gofmt -w cmd internal`; focused certify/conformance tests; Twenty conformance; `go run ./cmd/connectorgen validate internal/connectors/defs --json`; `go vet ./...`; `go build ./cmd/pm`; `gofmt -l cmd internal`; credential-free localhost `pm connectors certify twenty` when safe; `make verify` only if reverse-ETL smoke remains blocked by safety.
- Execution decision: `local_critical_path` — isolated worker cwd/branch assigned; no recursive `subagent` tool in worker context.
