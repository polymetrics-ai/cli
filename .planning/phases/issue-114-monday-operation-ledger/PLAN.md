# Plan — issue #114 Monday operation ledger

## Objective

Replace Monday's legacy 8-row `api_surface.json` with an operation-ledger inventory for the canonical Monday GraphQL API reference surface used by parent #82: 367 documented GraphQL operations, split into 87 query operations and 280 mutation operations.

## GSD mode

- `scripts/gsd prompt plan-phase issue-114-monday-operation-ledger --skip-research` generated this lane prompt.
- `programming-loop` command unavailable; manual TDD fallback active.

## Source/filter

- Source: `https://developer.monday.com/api-reference/reference/about-the-api-reference` plus sitemap reference pages.
- Count rule: headings under `# Queries` / `# Mutations` that contain GraphQL code examples.
- Parent-approved out-of-scope source pages: monday app framework/marketplace/platform docs (`app*`, app subscriptions/monetization, app blocks, marketplace discounts, batch trial extension, Ask Developer Docs, Platform API). This yields the required 367 / 87 / 280 count.

## Slice

1. Red test: assert Monday api surface has operation-ledger mode, 367 endpoint rows, 87 query rows, 280 mutation rows, no legacy `excluded`, and all implemented streams remain covered.
2. Green: generate `api_surface.json` ledger rows and `operations.json` metadata-only operation specs.
3. Refactor: ensure connectorgen validation remains clean; no writes become executable.

## Safety

All mutation/direct-read rows are metadata-only and blocked by default unless already covered by existing read streams. No arbitrary GraphQL execution, no secrets, no live calls.

## Verification

```bash
go test ./cmd/connectorgen -run 'TestMondayOperationLedger' -count=1
go test ./internal/connectors/engine -run 'TestBundleLoadEmbeddedMondayOperationLedger' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```
