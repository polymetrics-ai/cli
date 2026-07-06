# Overview

Reads accessible customers and allow-listed Google Ads GAQL search resources (campaigns, ad groups)
through the Google Ads REST API. Read-only; arbitrary GAQL is not accepted.

Readable streams: `accessible_customers`, `campaigns`, `ad_groups`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.google.com/google-ads/api/rest/overview.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Google OAuth 2.0 access token with Google Ads read
  scope. Used only for Bearer auth; never logged. Acquisition/refresh is out of scope for this
  connector (credentials layer already owns it).
- `base_url` (optional, string); default `https://googleads.googleapis.com/v24`; format `uri`;
  Google Ads REST API base URL override for tests or proxies.
- `customer_id` (optional, string); Google Ads customer id to query GAQL search resources
  (campaigns, ad_groups) against. Required for those streams only; accessible_customers is
  account-scoped and does not reference it.
- `developer_token` (required, secret, string); Google Ads developer token, sent as the
  developer-token header on every request. Never logged.
- `login_customer_id` (optional, string); Optional manager (MCC) customer id sent as the
  login-customer-id header. Omitted entirely when unset.
- `max_pages` (optional, string); Maximum pages fetched per GAQL search stream. A positive integer,
  or 'all'/'unlimited' (default) for no cap.
- `mode` (optional, string).
- `page_size` (optional, integer); default `1000`; GAQL search pageSize (1-10000).

Secret fields are redacted in logs and write previews: `access_token`, `developer_token`.

Default configuration values: `base_url=https://googleads.googleapis.com/v24`, `page_size=1000`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `customers:listAccessibleCustomers`.

## Streams notes

Default pagination: single request; no pagination.

- `accessible_customers`: GET `customers:listAccessibleCustomers` - records path `resourceNames`.
- `campaigns`: POST `customers/{{ config.customer_id }}/googleAds:search` - records path `results`.
- `ad_groups`: POST `customers/{{ config.customer_id }}/googleAds:search` - records path `results`.

## Write actions & risks

This connector is read-only. Read behavior: external Google Ads API read of
customer/campaign/ad-group metadata.

## Known limits

- Batch defaults: read_page_size=1000.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=1, out_of_scope=7, requires_elevated_scope=1.
