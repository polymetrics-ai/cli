# TEST-PLAN — Stripe connector

## Gate
`make verify` (default lane; CGO-free). Connector is pure-Go (connsdk/net-http) — no tag needed.

## Red-first tests (`internal/connectors/stripe/stripe_test.go`)
1. **Read + pagination + auth** (httptest server):
   - Page 1 returns `{"data":[{id:cus_1,created:...},{id:cus_2,...}],"has_more":true}`; page 2
     (`starting_after=cus_2`) returns `{"data":[{id:cus_3,...}],"has_more":false}`.
   - Assert: `Authorization: Bearer sk_test` header sent; 3 customers emitted, mapped (id, created…);
     incremental cursor advances to max `created`.
2. **Write validate** — `ValidateWrite(create_customer)` ok; unknown action rejected; `DryRunWrite`
   returns staged count without calling the server.
3. **Registry** — `connectors.NewRegistry().Get("stripe")` resolves with Write capability (via the
   package's own `_test` import triggering init); after catalog flip, `source-stripe` resolves too.

## Parity / regression
- All existing tests stay green. `make verify` green.
- `pm connectors inspect stripe --json` → kind Connector, read+write.
- `NativeConformanceReports` length == catalog length (unchanged).

## Evidence
Red captured in TDD-LEDGER.md before code; green after.
