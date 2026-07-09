# TDD Ledger: Bitbucket full-surface implementation

## Red evidence

Pending: add fail-first tests that assert Bitbucket coverage increases from 30/331 to full typed coverage and that binary direct-read policy is supported.

## Green evidence

Pending.

## Refactor evidence

Pending.

## Manual GSD fallback

`scripts/gsd prompt programming-loop init --phase issue-79-bitbucket-full-surface --dry-run` returned `scripts/gsd: unknown GSD command: programming-loop`; manual GSD loop remains active.

## Red evidence update

```bash
go test ./cmd/connectorgen -run Bitbucket -count=1
```

Result: failed as expected before full typed coverage: `covered endpoints = 30, want all 331 Bitbucket Swagger operations covered by typed surfaces`.

```bash
go test ./internal/connectors/engine -run TestDirectReadReturnsBoundedBitbucketBinaryAsBase64JSON -count=1
```

Result: failed as expected before adding the bounded binary direct-read policy: `direct read output policy "bitbucket_binary_base64" is not supported`.

## Green evidence

Implemented full typed Bitbucket coverage:

- `api_surface.json`: 331/331 official Swagger operations covered by typed surfaces; 0 blocked operation rows.
- GET operations: 179/179 covered by direct-read commands; existing streams remain covered.
- Mutations: 152/152 covered by named reverse-ETL write actions.
- CLI commands: 342 implemented commands after generated typed operation commands.
- Added bounded `bitbucket_binary_base64` direct-read policy for binary/text GET endpoints; it returns JSON/base64, enforces max bytes, and performs no filesystem writes.
- Added declarative write query support for provider endpoints with query parameters, without introducing raw HTTP execution.

Green commands:

```bash
go test ./internal/connectors/engine -run 'TestWriteQueryTemplatesUseRecordFieldsAndStayOutOfJSONBody|TestDirectReadReturnsBoundedBitbucketBinaryAsBase64JSON' -count=1
go test ./cmd/connectorgen -run Bitbucket -count=1
go test ./internal/cli -run Bitbucket -count=1
go test ./internal/connectors/conformance -run 'TestConformance/bitbucket' -count=1
go test ./internal/connectors/engine -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
```

Results: passed.

## Refactor evidence

- Regenerated Bitbucket connector docs/catalog and website connector data.
- Ran `gofmt -w cmd internal` through `make verify`.
- Kept `pm bitbucket api` raw API command `unsafe_or_disallowed`; generated commands are typed `operation <method> ...` commands bound to exact Swagger endpoints and explicit output/write policies.
