# SPEC — Stripe connector on connsdk

## Package
`internal/connectors/stripe/` (`package stripe`):
- `stripe.go` — `type Connector struct{}`; `func New() connectors.Connector`; `Name()=="stripe"`;
  `init(){ connectors.RegisterFactory("stripe", New) }`. Implements core interface + WriteValidator
  + DryRunWriter.
- `streams.go` — stream defs (name, primary_key=["id"], cursor=["created"], fields) + record mappers.
- `write.go` — allow-listed reverse-ETL actions + payload builders.
- `stripe_test.go` — httptest-backed read/pagination/incremental + write-validate + registry tests.
Add `_ "polymetrics/internal/connectors/stripe"` to `internal/connectors/registryset/registry_gen.go`.

## Config / auth / secrets
- secret `client_secret` (Stripe `sk_...`). Auth via `connsdk.Bearer(client_secret)`.
- config: `account_id` (optional → `Stripe-Account` header), `start_date` (RFC3339 → incremental
  lower bound via `created[gte]`), `base_url` (default `https://api.stripe.com/v1`), pagination knobs
  (`page_size` default 100, `max_pages`). `mode=fixture` short-circuits to deterministic fixtures
  (mirrors NativeCatalogConnector) so conformance works without live creds.

## Read (ETL)
- `connsdk.Requester{ BaseURL, Auth: Bearer(secret), DefaultHeaders, UserAgent }`.
- Per stream: GET `/<resource>` with `limit`, optional `created[gte]=<unix>` from cursor/start_date,
  optional `starting_after`. Records under `data[]` (use `connsdk.RecordsAt(body,"data")`).
- Pagination: loop while body `has_more==true`; set `starting_after` = id of last record. (Stripe's
  cursor is the last object id, not a body token — implement this loop with `connsdk.Requester` +
  `connsdk.StringAt(body,"has_more")`; a reusable connsdk "IdCursorPaginator" may be extracted later.)
- Incremental: track max `created` in StatefulReader state; emit cursor as unix seconds. Map records
  to flat connectors.Record per streams.go.

## Write (reverse-ETL)
- `ValidateWrite`: action ∈ allow-list {create_customer, update_customer} (start minimal; expandable).
  Reject unknown actions. `DryRunWrite` returns staged count + action, no external call.
- `Write`: POST form-encoded to the Stripe endpoint per action; never log secrets/PII; return
  WriteResult counts. (Kept minimal in batch 1 to prove the write template.)

## Fixture mode
When `cfg.Config["mode"]=="fixture"`: Check ok; Catalog returns the streams; Read emits 1–2 canned
records per stream; Write validates + returns a receipt-style success. Lets the existing conformance
harness exercise stripe without network/creds.

## Catalog flip
After green: set `source-stripe` `implementation_status=enabled`, `pm_connector_name=stripe` in
catalog_data.json so the alias resolves and conformance treats it as live.

## Out of scope
Remaining streams, Connect flows, Querier.
