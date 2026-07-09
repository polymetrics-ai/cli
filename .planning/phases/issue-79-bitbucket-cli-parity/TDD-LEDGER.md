# TDD Ledger: Bitbucket CLI Parity Parent

## Red evidence

- Parent orchestration setup is planning-only.
- #90 started with `go test ./cmd/connectorgen -run TestBitbucketCLISurfaceMetadata -count=1`, which failed because `internal/connectors/defs/bitbucket/cli_surface.json` did not exist yet.
- Broader regression red: `go test ./...` initially failed in catalog-count tests after Bitbucket increased the registered connector/bundle counts.

## Green evidence

- Planning artifacts created before production edits.
- Required GSD/Pi health commands passed:
  - `scripts/gsd doctor`
  - `scripts/gsd verify-pi`
  - `scripts/gsd list --json`
- `scripts/gsd prompt plan-phase issue-79-bitbucket-cli-parity --skip-research --tdd` generated an official GSD planning prompt.
- #90 metadata-only Bitbucket seed bundle, CLI surface, generated docs/catalog artifacts, and generated website data are green.
- Full local gates passed: `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, `go run ./cmd/connectorgen validate internal/connectors/defs`.

## Manual GSD fallback

`programming-loop` is not exposed by the current `scripts/gsd` registry. Command attempted:

```bash
scripts/gsd prompt programming-loop init --phase issue-79-bitbucket-cli-parity --dry-run
```

Result:

```text
scripts/gsd: unknown GSD command: programming-loop
```

Fallback in use: manual GSD universal runtime loop with `.pi/prompts/pm-gsd-loop.md`; maintain plan, red test/validation, green implementation, refactor, verification, commit/push, and review evidence.

## Refactor evidence

- Updated connector/catalog hard-coded count tests for the new Bitbucket bundle.
- Reverted unrelated broad connector-manual formatting churn from `pm docs generate`; kept only Bitbucket docs/catalog additions needed by `pm docs validate`.

## Lanes

| Issue | Red | Green | Refactor | Notes |
|---:|---|---|---|---|
| #90 | complete | complete | complete | metadata-only seed bundle verified; commit/push pending |
| #91 | blocked | blocked | blocked | waits for #90 metadata |
| #92 | blocked | blocked | blocked | waits for #90 and stream definitions |
| #93 | blocked | blocked | blocked | avoid `api_surface.json` collision with #90 seed |
| #94 | blocked | blocked | blocked | waits for operation ledger |
| #95 | blocked | blocked | blocked | waits for operation ledger need classification |
| #96 | blocked | blocked | blocked | waits for operation ledger write risk classification |
