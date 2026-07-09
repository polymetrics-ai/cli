# Plan — issue #113 Monday stream runner

## Objective

Prove implemented Monday stream-backed commands execute through the generic connector command runner without live credentials.

## GSD mode

- `scripts/gsd prompt plan-phase issue-113-monday-stream-runner --skip-research` generated this plan prompt.
- `programming-loop` command unavailable; manual TDD fallback active.

## Slice

1. Red test: use the real Monday bundle/Hook connector from `bundleregistry`, an `httptest.Server`, and `commandrunner.Run` for `board list` (and possibly `item list`) to show command execution.
2. Green: existing #111 metadata should satisfy the runner; implement only if the test reveals a safe metadata/runtime gap.
3. Verify no credentials are read and no live network is used.

## Safety

No live monday.com calls. Test server returns sanitized fixture-shaped GraphQL responses. No reverse ETL execution.

## Verification

```bash
go test ./internal/connectors/commandrunner -run 'TestRunMonday' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```
