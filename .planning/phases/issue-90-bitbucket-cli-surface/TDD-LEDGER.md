# TDD Ledger: Bitbucket CLI Surface Metadata

## Red tests

Planned first red:

```bash
go test ./cmd/connectorgen -run TestBitbucketCLISurfaceMetadata -count=1
```

Expected initial failure: missing `internal/connectors/defs/bitbucket/cli_surface.json` / missing Bitbucket bundle.

## Green tests

Pending.

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

Pending.

## Safety notes

- No credentialed connector checks.
- No Bitbucket secret values in examples or fixtures.
- No binary downloads, local git mutations, raw HTTP writes, or reverse ETL execution in this metadata slice.
