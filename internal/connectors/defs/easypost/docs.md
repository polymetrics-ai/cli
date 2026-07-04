# Overview

EasyPost is a declarative-HTTP connector for the EasyPost REST API v2
(`https://api.easypost.com/v2`). The original legacy connector exposed five read streams
(`shipments`, `trackers`, `addresses`, `parcels`, `insurances`); this Pass B bundle expands the
documented EasyPost surface with additional list streams for batches, carrier accounts, carrier
metadata/types, end shippers, events, claims, pickups, refunds, scan forms, child users, referral
customers, and webhooks, plus documented write actions that fit the JSON write dialect.

## Auth setup

Provide the EasyPost API key via the `username` secret. EasyPost uses HTTP Basic auth with the API
key as the username and an empty password, so the bundle sends `Authorization: Basic
base64(<api_key>:)`. `base_url` defaults to `https://api.easypost.com/v2` and may be overridden for
tests/proxies.

## Streams notes

EasyPost paginated list endpoints return records newest-first with `has_more` and the next page
requested by passing the last record ID as `before_id`. The bundle sends `page_size=100`, the
documented maximum, on paginated list streams. `child_users` is the documented exception using
`after_id` pagination. Existing legacy streams keep schema projection to preserve the hand-written
connector's explicit field mapping; newly added Pass B streams use `projection: "passthrough"` so
the full documented EasyPost object payload remains available beyond the compact catalog schema.

Incremental streams that support EasyPost's `start_datetime` lower bound (`shipments`, `trackers`,
`addresses`, `parcels`, `insurances`, `batches`, `events`, `claims`, `pickups`, `refunds`,
`scan_forms`) use `created_at` as the cursor and fall back to `start_date` on a fresh sync.

## Write actions & risks

`writes.json` covers the documented object lifecycle and operational actions that can be expressed
as a single JSON/empty-body HTTP request: address/parcel/customs/shipment/tracker/batch/end-shipper/
insurance/order/pickup/refund/scan-form/report/Luma/user/webhook creation and updates, shipment and
batch buy/refund/label/form actions, and idempotent deletes for trackers, child users, and webhooks.

Several writes are operational or charge-bearing in EasyPost: buying shipments, orders, batches, or
Luma shipments can purchase real postage; insurance and refund actions can affect billing/refund
state; pickup buy/cancel changes real carrier pickup state. Those actions are marked
`confirm: "destructive"` and all writes require reverse-ETL plan preview/approval before execution.

## Known limits

- `page_size` and `max_pages` are not runtime-configurable in this bundle. Legacy allowed those as
  config overrides; the declarative cursor paginator sends EasyPost's documented maximum
  `page_size=100` and stops on `has_more=false`.
- Account credential/admin and billing primitives are intentionally excluded from the executable
  write surface where they create/delete API keys, store carrier credentials, manage payment
  methods, or move funds. They remain enumerated in `api_surface.json` with explicit risk reasons.
- Single-ID detail GETs are generally excluded as `duplicate_of` when the corresponding list stream
  already emits the same object shape. Detail-only resources such as CustomsInfo, CustomsItem, and
  Order are covered by creation writes but have no bulk list stream to sync.
- Label/form/report URLs are emitted as metadata. The connector does not download binary label,
  form, scan-form, or report files.
