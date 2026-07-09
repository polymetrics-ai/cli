# TDD Ledger: Issue #95 bitbucket-advanced-engine

## Red evidence

Pending. Add failing tests before production edits for this lane.


## Red evidence update

```bash
go test ./cmd/connectorgen -run Bitbucket -count=1
```

Result: failed as expected before full Bitbucket parity metadata existed: `implemented commands = 0, want at least 10` and `api_surface endpoints = 8, want official Bitbucket Swagger count 331`.

```bash
go test ./internal/cli -run Bitbucket -count=1
```

Result: failed as expected before help/runtime command implementation: `help topic "bitbucket" not found`.

## Green evidence

Superseded by final green evidence below.

## Refactor evidence

Superseded by final refactor evidence below.

## Manual GSD fallback

`scripts/gsd prompt programming-loop init --phase issue-79-bitbucket-cli-parity --dry-run` returned `scripts/gsd: unknown GSD command: programming-loop`; manual GSD loop remains active.

## Final green evidence

Bitbucket is recorded as REST-only; no GraphQL executor or arbitrary raw API escape hatch was added.

```bash
go test ./cmd/connectorgen -run Bitbucket -count=1
go test ./internal/cli -run Bitbucket -count=1
go test ./internal/connectors/conformance -run 'TestConformance/bitbucket' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
```

Results: passed.

## Final verification evidence

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
./pm help bitbucket
./pm bitbucket
./pm bitbucket --help
npm --prefix website run gen:website-data
```

Results: passed. `make verify` initially exceeded the 900s harness timeout during the long connector certification package, then passed when rerun with a longer timeout; no code or dependency changes were needed between the timeout and the passing rerun.

## Refactor evidence

- Ran `gofmt -w cmd internal`.
- Added fixture-backed conformance coverage for Bitbucket streams and representative writes.
- Regenerated connector docs/catalog and website generated connector data.
