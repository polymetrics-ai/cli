# Overview

Reads QuickBooks Online customers, invoices, payments, accounts, and vendors through the v3 Query
API via the OAuth 2.0 refresh-token grant. The bundle also carries the complete r2 official
QuickBooks Online Accounting API qbo operation ledger from Intuit `EntityJsonObject_v1.json`:
43 ETL/read rows, 60 reverse-ETL write rows, 32 bounded direct/provider-query/search rows,
25 binary/file/report rows, and 1 CDC/changefeed row.

Executable coverage is intentionally limited to the five legacy-compatible read streams:
`customers`, `invoices`, `payments`, `accounts`, and `vendors`. All other official rows are tracked
exactly once in `api_surface.json` and remain blocked by default until typed schemas, bounds,
redaction, idempotency, fixtures, and any shared-runtime support exist.

Service API documentation:
https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/account.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://quickbooks.api.intuit.com`; format `uri`;
  QuickBooks Online Accounting API base URL override for tests or proxies. Use
  https://sandbox-quickbooks.api.intuit.com for a sandbox company.
- `client_id` (required, secret, string); Intuit Developer app Client ID, obtained from the app's
  Keys tab. Used only in the token-refresh Basic Authorization header; never logged.
- `client_secret` (required, secret, string); Intuit Developer app Client Secret, obtained from the
  app's Keys tab. Used only in the token-refresh Basic Authorization header; never logged.
- `max_pages` (optional, string); Permissive parse: empty, "all", or "unlimited" uses the
  connector's 10000-page safety cap and returns a clear error if that cap is reached; a positive
  integer string caps the page count at that value.
- `page_size` (optional, string); default `1000`; Records per page (1-1000, Query API MAXRESULTS).
- `realm_id` (required, string); QuickBooks company (Realm) ID that scopes every Query API request.
  It is not an OAuth secret, but it is validated as a numeric path segment before use.
- `refresh_token` (required, secret, string); Long-lived QuickBooks OAuth 2.0 refresh token,
  exchanged for a short-lived access token at `token_url`; never logged. The 3-legged consent and
  acquisition flow is out of scope for this connector because the credentials layer owns it.
- `sandbox` (optional, string); default `false`; Documents whether this realm is a sandbox company
  (informational only; `base_url` is the operative environment switch).
- `start_date` (optional, string); format `date-time`; retained as a forward-compatible config
  field. The current stream hook does not send a server-side incremental filter.
- `token_url` (optional, string); default
  `https://oauth.platform.intuit.com/oauth2/v1/tokens/bearer`; format `uri`; Intuit OAuth 2.0
  token endpoint override. The hook fails closed on a non-https or unparseable value to prevent
  sending refresh credentials to an attacker-chosen endpoint.

Secret fields are redacted in logs and previews: `client_id`, `client_secret`, and
`refresh_token`.

Default configuration values: `base_url=https://quickbooks.api.intuit.com`, `page_size=1000`,
`sandbox=false`, `token_url=https://oauth.platform.intuit.com/oauth2/v1/tokens/bearer`.

Authentication behavior:

- Connector-specific authentication uses `secrets.refresh_token`, `config.token_url`,
  `secrets.client_id`, and `secrets.client_secret`; client credentials are sent through HTTP Basic
  auth on the token refresh request.
- Access tokens are cached until 60 seconds before expiry, then refreshed through the OAuth 2.0
  refresh-token grant.

Connection checks call GET `v3/company/{{ config.realm_id }}/query` with query `query`=`SELECT *
FROM Customer STARTPOSITION 1 MAXRESULTS 1`.

## Streams notes

Pagination embeds `STARTPOSITION` and `MAXRESULTS` inside the Query API `query` string value.
Paging stops when a page returns fewer than `page_size` records or `max_pages` is reached, whichever
comes first. The stream hook checks context cancellation between pages and records.

- `customers`: GET `v3/company/{{ config.realm_id }}/query` - records path
  `QueryResponse.Customer`; fixture-backed.
- `invoices`: GET `v3/company/{{ config.realm_id }}/query` - records path
  `QueryResponse.Invoice`; fixture-backed.
- `payments`: GET `v3/company/{{ config.realm_id }}/query` - records path
  `QueryResponse.Payment`; fixture-backed.
- `accounts`: GET `v3/company/{{ config.realm_id }}/query` - records path
  `QueryResponse.Account`; fixture-backed.
- `vendors`: GET `v3/company/{{ config.realm_id }}/query` - records path
  `QueryResponse.Vendor`; fixture-backed.

The remaining official ETL/read rows are blocked/planned in `api_surface.json`; they are not
silently excluded.

## Write actions & risks

No executable write actions are declared in this bundle. Official QuickBooks create, update, delete,
void, send, and batch operations are represented as blocked operation-ledger rows rather than raw
write escapes.

Future write enablement must be action-by-action with a closed `record_schema`, redaction, risk
text, provider-supported idempotency notes, sanitized fixtures, and the existing reverse ETL
plan -> preview -> explicit approval -> execute flow. Destructive operations must also declare
`confirm: "destructive"`; they are not blanket-excluded as unsafe.

## Known limits

- Batch defaults: read_page_size=1000.
- API coverage includes 161 operation-ledger row(s): 5 stream-backed executable rows and 156
  blocked/planned rows.
- Official lane counts represented in the ledger: `etl_read=43`, `reverse_etl_write=60`,
  `direct_read_query_search=32`, `binary_file=25`, `cdc_changefeed=1`,
  `excluded_not_applicable=0`, `total=161`.
- Fixture-backed executable coverage remains uncertified: 5 ETL streams, 0 write actions,
  0 implemented direct-read operations, 0 binary operations, and 0 CDC/changefeed operations.
- The bundle-level conformance skip is intentional: QuickBooks uses a custom OAuth refresh-token
  AuthHook and a StreamHook whose pagination state is embedded inside the Query API string. Static
  validation, hook unit tests, sanitized fixtures, and archived parity evidence are the local proof
  for fixture-only behavior.
- Additional provider-query/search, report/PDF/attachment binary surfaces, CDC, and all mutations
  require typed connector definitions and/or shared-runtime support before execution. No raw SQL,
  arbitrary provider query, generic HTTP, shell, file, or passthrough escape hatch is exposed.
