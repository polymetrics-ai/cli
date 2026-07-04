# Overview

Xero covers the Xero Accounting API at `https://api.xero.com/api.xro/2.0`. Pass B expands the
bundle from the six legacy read streams to the full official Accounting OpenAPI surface:
JSON/documented GET operations are streams, and JSON/no-body PUT, POST, and DELETE operations are
typed write actions.

The legacy streams (`invoices`, `contacts`, `accounts`, `bank_transactions`, `items`, `payments`)
keep their existing schema-mode projections and computed aliases for parity with
`internal/connectors/xero`. New docs-derived streams use passthrough projection with minimal
catalog schemas so live Xero response fields are preserved.

## Auth setup

Provide a Xero OAuth2 `access_token` secret and a `tenant_id` secret. The access token is sent as a
Bearer token. The tenant id is sent in the `Xero-tenant-id` header. `base_url` defaults to the Xero
Accounting API base URL and may be overridden for tests or proxies.

## Streams notes

Existing legacy streams retain their original page-number behavior (`page` starts at `1`, Xero's
fixed 100-record page size, stop on a short page). Additional streams are generated from
`xero_accounting.yaml`; endpoints that expose a `page` query parameter use the same page-number
pagination, while detail/report/history streams use `pagination: none`.

Binary PDF and attachment-content download endpoints are not streams. Attachment metadata list
endpoints are streams because they return JSON envelopes.

## Write actions & risks

`writes.json` contains JSON or no-body write actions for every dialect-expressible Accounting API
mutation. DELETE operations are marked destructive, and Xero POST operations whose operation name is
delete-like are also treated as destructive write actions.

Attachment upload endpoints are excluded because they require binary `application/octet-stream`
request bodies, which the declarative write dialect does not model.

## Known limits

- This bundle is scoped to the Accounting API only. Payroll, Assets, Files, Projects, and other
  Xero APIs use separate base paths and are outside this connector's metadata/legacy scope.
- Optional query filters from the OpenAPI spec are not modeled. The streams still cover the
  documented endpoints and preserve returned fields through passthrough projection.
- `tenant_id` is modeled only as a secret header value. Legacy also allowed a plain config fallback;
  callers should supply the tenant id as a secret for this bundle.
