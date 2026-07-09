# TDD Ledger: Bitbucket CLI Surface Metadata

## Red tests

```bash
go test ./cmd/connectorgen -run TestBitbucketCLISurfaceMetadata -count=1
```

Result: failed as expected before production Bitbucket defs existed.

```text
--- FAIL: TestBitbucketCLISurfaceMetadata (0.00s)
    bitbucket_cli_surface_test.go:12: read bitbucket cli_surface.json: open ../../internal/connectors/defs/bitbucket/cli_surface.json: no such file or directory
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	0.431s
FAIL
```

Broader regression red found after adding the connector bundle:

```bash
go test ./...
```

Result: failed in `internal/cli/catalog_cli_test.go` because the hard-coded connector catalog count expected `551` after Bitbucket increased the catalog to `552`. `internal/connectors/bundleregistry/registry_test.go` also needed the bundle-count expectation updated from `547` to `548`.

## Green tests

```bash
jq . internal/connectors/defs/bitbucket/*.json
```

Result: passed.

```bash
go test ./cmd/connectorgen -run TestBitbucketCLISurfaceMetadata -count=1
```

Result: passed after adding the Bitbucket seed bundle and `cli_surface.json`.

```bash
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

Result: passed; `connectors_checked=548`, `findings=0`, `warnings=0`.

```bash
go test ./cmd/connectorgen -count=1
go test ./internal/connectors/engine ./internal/connectors/commandrunner ./cmd/connectorgen -count=1
go build ./cmd/pm
```

Result: passed.

```bash
go test ./internal/cli ./internal/connectors/bundleregistry -count=1
go test ./...
go vet ./...
go build ./cmd/pm
./pm docs validate --connectors-dir docs/connectors
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

Result: passed after updating catalog/bundle count tests and adding generated Bitbucket connector manual/catalog docs.

```bash
cd website && pnpm run gen:website-data
cd website && pnpm run gen:website-data
```

Result: passed and idempotent after second generation.

## Manual GSD fallback

`programming-loop` is unavailable through `scripts/gsd` in this checkout.

```bash
scripts/gsd prompt programming-loop init --phase issue-90-bitbucket-cli-surface --dry-run
```

Result:

```text
scripts/gsd: unknown GSD command: programming-loop
```

Fallback: manual GSD universal runtime loop with required plan, red, green, refactor, verification, commit/push, and review evidence.

## Refactor notes

- Kept #90 metadata-only: no implemented commands, no streams, no writes, no direct-read output policy, no binary/local executor.
- Added a minimal operation-ledger seed in `api_surface.json` only to document that #93 owns the complete 331-operation inventory.
- Regenerated website connector data because adding a connector bundle changes generated catalog files.
- Added generated Bitbucket connector manual/catalog docs needed by `pm docs validate`; reverted unrelated broad connector-manual formatting churn from the docs generator.

## Safety notes

- No credentialed connector checks.
- No Bitbucket secret values in examples or fixtures.
- No binary downloads, local git mutations, raw HTTP writes, or reverse ETL execution in this metadata slice.
