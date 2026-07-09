# Plan — issue #115 Monday direct reads

## Objective

Enable a small, safe Monday direct-read slice using the fixed GraphQL direct-read engine from #116.

## GSD mode

- `scripts/gsd prompt plan-phase issue-115-monday-direct-reads --skip-research` generated this prompt.
- `programming-loop` command unavailable; manual TDD fallback active.

## Slice

1. Red tests: Monday commandrunner can execute implemented `me view` and `account view` commands against a local GraphQL replay server; metadata validation requires those commands to be covered by direct-read api surface rows.
2. Green: update `cli_surface.json`, `api_surface.json`, and targeted operation specs for `monday.me.get_me` and `monday.account.get_account` only.
3. Refactor: verify no board/item direct read overclaiming and no mutation/direct write becomes executable.

## Safety

- No live Monday calls; tests use `httptest.Server` fixtures.
- Only bundled fixed query documents execute; no arbitrary GraphQL input.
- Queries return bounded JSON and fail closed on GraphQL errors.

## Verification

```bash
go test ./internal/connectors/commandrunner -run 'TestRunMondayDirectRead' -count=1
go test ./cmd/connectorgen -run 'TestMondayDirectRead' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```
