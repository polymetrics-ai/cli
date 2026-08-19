<!-- CI repair handoff: the outer PR phase must replace PR #3542's live body with this file verbatim. -->

## Intent

Complete the documented Ashby connector parity slice against the current official API inventory.
The connector now records all 212 documented REST and webhook operations exactly once: 71 fixture-backed streams, 9 bounded direct reads, 98 typed reverse-ETL writes, and 34 explicitly blocked operations.

## Linked Issue

Refs #3207

## What Changed

- expanded the Ashby operation ledger, typed schemas, fixtures, command surface, and generated documentation;
- routed bounded reads and typed writes through Ashby-owned validation and engine paths;
- kept destructive writes behind plan, preview, explicit approval, and execute safeguards; and
- documented unsupported webhook, multipart, upload-handle, conditional-side-effect, and checkpoint-foundation operations without exposing generic request escape hatches.

## Testing

- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go test ./internal/connectors/conformance -run 'TestConformance/ashby' -count=1`
- `go test ./internal/connectors/native/ashby ./internal/connectors/hooks/ashby -count=1`
- `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout 10m`
- `go vet ./internal/connectors/... ./internal/cli/...`
- `go build ./cmd/pm`
- `make connector-boundary`
- `make verify`

## Safety

No provider credentials, live provider calls, provider writes, infrastructure changes, or generic HTTP, shell, file, SQL, raw-query, arbitrary-body, or passthrough surfaces were used or added.

## Pipeline

Draft PR #3542's live body currently contains only a bare issue URL, which does not satisfy the issue-first guard. The outer PR phase must replace the live body with this file so `Refs #3207` is present before rerunning `require-linked-issue`.
