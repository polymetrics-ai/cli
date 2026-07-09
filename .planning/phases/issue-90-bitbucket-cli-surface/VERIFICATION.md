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
- `docs/cli/**`: exempt for #90.
- `website/**`: run generation/idempotency only if connector metadata generation changes website data.
- #91 owns rendered help/docs parity.

## Results

Pending red/green implementation.
