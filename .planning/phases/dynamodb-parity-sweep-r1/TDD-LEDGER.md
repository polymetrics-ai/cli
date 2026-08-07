# DynamoDB parity TDD ledger

## Red / validation baseline before production edits

Taken before any edit to `internal/connectors/defs/dynamodb`:

- `go run ./cmd/connectorgen validate internal/connectors/defs/dynamodb`: passes today with 0
  findings — the existing bundle is internally consistent but describes almost nothing. Validate
  alone cannot detect a missing operation surface, which is precisely the gap this test closes.
- Baseline bundle: **9** `api_surface.json` endpoints (1 `covered_by`, 8 legacy `excluded`), no
  `operation_ledger_version`, `capabilities.write: false`, and no `cli_surface.json`,
  `operations.json`, or `writes.json`. **0 of 58 documented operations are reachable as
  `pm dynamodb <command>`.**
- Official re-derivation from AWS's service model (botocore
  `data/dynamodb/2012-08-10/service-2.json`, byte-identical at tag `1.43.66`, set-equal to
  `API_Operations.html`): **58 operations, 27 read / 31 write, none deprecated**.
- Ledger drift: sweep carried forward **57**. Proven stale-by-one, not wrong-when-taken — botocore
  1.35.0 / 1.38.0 / 1.40.0 each declare 57 with no `SearchVectors`; current declares 58 with it.
  Writes reconcile at 31 with no adjustment, which independently corroborates both numbers.

### Red evidence

`go test ./cmd/connectorgen -run TestDynamoDBAPISurfaceOperationLedger` fails as designed:

```
--- FAIL: TestDynamoDBAPISurfaceOperationLedger (0.00s)
    dynamodb_api_surface_test.go:60: operation_ledger_version is unset; the v2 provenance ledger is required
    dynamodb_api_surface_test.go:64: api_surface declares 9 endpoints, want 58 documented operations
    dynamodb_api_surface_test.go:114: 8 legacy excluded row(s) remain; operation_ledger_version mode requires operation rows
    dynamodb_api_surface_test.go:117: covered(1)+blocked(0) = 1, want 58
```

## Planned tests / assertions

`cmd/connectorgen/dynamodb_api_surface_test.go` asserts:

- 58 endpoint rows, matching the derived documented-operation total;
- the 27/31 read/write split sums to 58;
- `operation_ledger_version` is set (v2 provenance mode);
- **no duplicate endpoint row** — load-bearing here, because every DynamoDB operation is `POST /`
  selected by `X-Amz-Target`, so rows keyed on the bare path would collapse 58 operations into 1;
- every row carries **exactly one** disposition — none blank, none doubled;
- no legacy `excluded` rows survive under v2 mode;
- every blocked row is `blocked_by_default` with `status: blocked`, a source citation, and a
  `named_dependency=` note.

Still to run at green: `connectorgen validate`, `TestEveryImplementedCommandPassesRuntimePreflight`,
`surface-sync --check`, conformance, `make certify-timing`, docs-check, connector-boundary, and a
real binary run of each scope.

## Refactor / safety notes

- Do not loosen validate, the connector boundary gate, certify, or the runtime preflight test.
- Nothing marked `implemented` unless its command runs; blocked rows name their dependency.
- Streams (4 ops) and DAX (21 ops) stay out — a naive `API_*.html` scrape yields 84, not 58.
- Fixtures stay synthetic and secret-free; no credential or token-derived value is ever emitted.
- Keep the diff scoped to dynamodb.
