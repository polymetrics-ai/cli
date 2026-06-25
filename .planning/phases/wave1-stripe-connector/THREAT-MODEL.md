# THREAT-MODEL — Stripe connector

## Assets
- Stripe secret API key (`client_secret`, sk_...) in the pm vault. Customer PII in records.
- Reverse-ETL write capability against the live Stripe account (create/update customer).

## Risks & mitigations
- **Secret leakage**: never log/print `client_secret`; it only feeds `connsdk.Bearer`. Errors must
  not interpolate the key or full request bodies. connsdk's HTTPError truncates response bodies and
  is not logged by the connector.
- **PII exposure**: connector returns records to the warehouse; do not add debug logging of record
  contents. Reverse-ETL preview shows mapped sample only under the existing approval flow.
- **Unauthorized/over-broad writes**: `ValidateWrite` enforces an allow-list (create/update_customer
  in batch 1); unknown actions rejected. Writes stay plan→preview→approve→execute. Default fixture
  mode performs no external mutation.
- **SSRF via base_url**: `base_url` override is validated (https scheme, host present); default is
  api.stripe.com. No user-controlled path injection beyond fixed stream endpoints.
- **Rate limits**: connsdk.Requester honors 429/Retry-After + backoff.

## Non-applicable
No new build dependency (pure-Go connsdk + net/http). No CGO.
