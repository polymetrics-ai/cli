# Overview

Reads QuickBooks Online customers, invoices, payments, accounts, and vendors through the v3 Query
API via the OAuth 2.0 refresh-token grant. Read-only.

Readable streams: `customers`, `invoices`, `payments`, `accounts`, `vendors`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://developer.intuit.com/app/developer/qbo/docs/api/accounting/all-entities/account.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://quickbooks.api.intuit.com`; format `uri`;
  QuickBooks Online Accounting API base URL override for tests or proxies. Use
  https://sandbox-quickbooks.api.intuit.com for a sandbox company.
- `client_id` (required, secret, string); Intuit Developer app Client ID, obtained from the app's
  Keys tab. Used only in the token-refresh request form; never logged.
- `client_secret` (required, secret, string); Intuit Developer app Client Secret, obtained from the
  app's Keys tab. Used only in the token-refresh request form; never logged.
- `max_pages` (optional, string); Permissive parse: empty, "all", or "unlimited" means unbounded; a
  positive integer string caps the page count at that value.
- `page_size` (optional, string); default `1000`; Records per page (1-1000, Query API MAXRESULTS).
- `realm_id` (required, secret, string); QuickBooks company (Realm) ID that scopes every Query API
  request. Treated as secret-adjacent (identifies a specific company) and validated as a path-safe
  segment (no '/', '?', '#', or '..').
- `refresh_token` (required, secret, string); Long-lived QuickBooks OAuth 2.0 refresh token (valid
  up to ~100 days), exchanged for a short-lived access token at token_url; never logged. The
  3-legged consent/acquisition dance is out of scope for this connector (credentials layer already
  owns it).
- `sandbox` (optional, string); default `false`; Documents whether this realm is a sandbox company
  (informational only in this pass; base_url is the operative environment switch -- see Known
  limits).
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound for the earliest record
  to replicate.
- `token_url` (optional, string); default
  `https://oauth2.bearer.token.intuit.com/oauth2/v1/tokens/bearer`; format `uri`; Intuit OAuth 2.0
  token endpoint override. MUST be https in production; the hook fails closed on a non-https or
  unparseable value to prevent exfiltrating the refresh token/client secret to an attacker-chosen
  endpoint.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`, `realm_id`,
`refresh_token`.

Default configuration values: `base_url=https://quickbooks.api.intuit.com`, `page_size=1000`,
`sandbox=false`, `token_url=https://oauth2.bearer.token.intuit.com/oauth2/v1/tokens/bearer`.

Authentication behavior:

- Connector-specific authentication using `secrets.refresh_token`, `config.token_url`,
  `secrets.client_id`, `secrets.client_secret`.
- Access tokens are valid for one hour and are cached until 60 seconds before expiry, at which
  point a fresh token is obtained.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `v3/company/{{ config.realm_id }}/query` with query `query`=`SELECT *
FROM Customer STARTPOSITION 1 MAXRESULTS 1`.

## Streams notes

Pagination embeds `STARTPOSITION` and `MAXRESULTS` inside the query string value itself, per the
QuickBooks Query API. Paging stops when a page returns fewer than `page_size` records or
`max_pages` is reached, whichever occurs first.

- `customers`: GET `v3/company/{{ config.realm_id }}/query` - records path `QueryResponse.Customer`.
- `invoices`: GET `v3/company/{{ config.realm_id }}/query` - records path `QueryResponse.Invoice`.
- `payments`: GET `v3/company/{{ config.realm_id }}/query` - records path `QueryResponse.Payment`.
- `accounts`: GET `v3/company/{{ config.realm_id }}/query` - records path `QueryResponse.Account`.
- `vendors`: GET `v3/company/{{ config.realm_id }}/query` - records path `QueryResponse.Vendor`.

## Write actions & risks

This connector is read-only. Read behavior: external QuickBooks Online v3 Query API read of
customer/invoice/payment/account/vendor entities.

## Known limits

- Batch defaults: read_page_size=1000.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, non_data_endpoint=1, out_of_scope=4.
