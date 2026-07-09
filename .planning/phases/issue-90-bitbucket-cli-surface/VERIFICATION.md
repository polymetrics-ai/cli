# Verification: Bitbucket CLI Surface Metadata

Date: 2026-07-09

## Required checks for this slice

```bash
jq . internal/connectors/defs/bitbucket/*.json
go test ./cmd/connectorgen -run TestBitbucketCLISurfaceMetadata -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
go test ./cmd/connectorgen -count=1
go build ./cmd/pm
```

## Broader gates before handoff if implementation proceeds beyond metadata

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
make verify
```

## CLI help/docs/website parity

- Runtime help: exempt for #90 because no dispatcher/help renderer behavior is changed.
- Bare namespace behavior: exempt for #90.
- `docs/cli/**`: no CLI command behavior changed; connector manual/catalog docs were updated because docs validation requires every registered connector.
- `website/**`: generated website connector data updated and re-run idempotently.
- #91 still owns rendered Bitbucket runtime help/docs parity.

## Results

- `jq . internal/connectors/defs/bitbucket/*.json`: passed.
- `go test ./cmd/connectorgen -run TestBitbucketCLISurfaceMetadata -count=1`: red failed before bundle existed, green passed after implementation.
- `go run ./cmd/connectorgen validate internal/connectors/defs --json`: passed; `connectors_checked=548`, `findings=0`, `warnings=0`.
- `go test ./cmd/connectorgen -count=1`: passed.
- `go test ./internal/connectors/engine ./internal/connectors/commandrunner ./cmd/connectorgen -count=1`: passed.
- `go build ./cmd/pm`: passed.
- `./pm help connectors`: passed before connector inspection.
- `./pm connectors inspect bitbucket --json`: passed without reading credentials.
- `cd website && pnpm run gen:website-data` twice: passed; second run idempotent.
- `git diff --check`: passed.
- `go test ./internal/cli ./internal/connectors/bundleregistry -count=1`: passed after catalog/bundle count updates.
- `go vet ./...`: passed.
- `go test ./...`: passed.
- `go build ./cmd/pm`: passed after final changes.
- `./pm docs validate --connectors-dir docs/connectors`: passed after adding Bitbucket connector docs/catalog artifacts.
- `make verify`: passed.
- `go run ./cmd/connectorgen validate internal/connectors/defs`: passed; `connectors_checked=548`, `findings=0`.
- Commit/push checkpoint: `0e359d76` pushed to `feat/79-bitbucket-cli-parity`.
- CodeRabbit actionable finding disposition: accepted; tightened `direct_write` safety assertion in `cmd/connectorgen/bitbucket_cli_surface_test.go`.
- Post-review-fix gates: `go test ./cmd/connectorgen -run TestBitbucketCLISurfaceMetadata -count=1`, `go test ./cmd/connectorgen -count=1`, `go test ./...`, and `make verify` passed.
