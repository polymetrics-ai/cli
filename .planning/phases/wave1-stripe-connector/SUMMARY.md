# SUMMARY — Wave 1 batch 1: Stripe connector (declarative-HTTP reference)

Status: **completed** (GO). `make verify` green; orchestrator review GO; TDD red-before-code enforced.

## What shipped
- **`internal/connectors/stripe/`** — the reference declarative-HTTP connector, built on `connsdk`:
  `stripe.go` (New/Name/init→RegisterFactory, Metadata read+write, Check, Catalog, Read with Bearer
  auth + `data[]` extraction + `has_more`/`starting_after` id-cursor pagination + StatefulReader
  cursor + fixture mode + SSRF-validated base_url), `streams.go` (customers/charges/invoices/
  subscriptions/products + mappers), `write.go` (allow-listed `create_customer`/`update_customer`
  with ValidateWrite/DryRunWrite/Write), `manifest.go`.
- **connsdk gained `DoForm`** (form-encoded POST) via a shared `do` core — JSON `Do` behavior
  unchanged; +2 tests. Lets Stripe writes reuse connsdk auth/retry.
- **Registered** via `registryset` blank import; **`source-stripe` catalog entry flipped** to
  enabled + `pm_connector_name=stripe`.

## Verification
- `make verify` exit 0; full `go test ./...` 11 pkgs ok; conformance length 647 (enabled 2 / planned 645).
- Parity: `stripe` + `source-stripe` → kind Connector, read+write.
- Security: secret never logged; base_url SSRF-validated; write allow-list enforced.
- TDD: 3 behavior tasks red-confirmed then green.

## Significance
This is the **template for the ~500 declarative-HTTP connectors**: a thin package composing connsdk
(Requester/Bearer/RecordsAt/cursor) with per-system stream defs + write actions. The next HTTP
connectors copy this shape; mostly stream/endpoint/auth data changes.

## Boundary / deferred
- Core streams only (more Stripe streams later). Reusable `connsdk.IdCursorPaginator` deferred
  (Stripe's has_more loop kept in-package per ADR).
- No Querier (analysis via the DuckDB warehouse).

## Next (Wave 1 continues)
- More GA connectors in batches (postgres+CDC, slack, hubspot, shopify, jira, notion, bigquery,
  snowflake, s3 …). A connector codegen scaffold (skeleton + registryset import + red conformance
  tests from a catalog entry) would accelerate the long tail — candidate next step.
