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
- Touch only `internal/connectors/defs/twenty/fixtures/**`, `internal/connectors/defs/twenty/docs.md`, `internal/connectors/conformance/twenty_test.go`, S7 phase artifacts, and generated website/catalog data if required by generator/idempotency.
- Do not edit `api_surface.json`, `streams.json`, `writes.json`, `schemas/**`, engine runtime code, CLI code, `go.mod`/`go.sum`, other connectors, `.github/**`, `main`, or parent state files.
