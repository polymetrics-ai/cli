# Overview

GoCardless is a declarative HTTP connector for the official GoCardless REST API. This Pass B bundle preserves the four legacy record projections for `payments`, `mandates`, `payouts`, and `refunds`, then adds the remaining documented JSON list/detail endpoints as passthrough streams plus documented bearer-authenticated mutation endpoints as write actions.

## Auth setup

Provide a GoCardless access token via the `access_token` secret; it is sent as `Authorization: Bearer <access_token>` and is never logged, matching legacy `connsdk.Bearer(secret)`. `base_url` is required by this bundle; set it to `https://api.gocardless.com` for live or `https://api-sandbox.gocardless.com` for sandbox. `gocardless_version` defaults to `2015-07-06` and is sent as the `GoCardless-Version` header.

`gc_key_id` is optional and is sent as `Gc-Key-Id` when configured. It is only needed for the encrypted bank account details endpoint, but the declarative engine has bundle-level headers rather than stream-scoped headers, so it is omitted unless explicitly supplied.

## Streams notes

The legacy streams remain first and keep schema projection so their emitted records match `internal/connectors/gocardless`: `payments`, `mandates`, `payouts`, and `refunds`. They keep the legacy incremental `created_at[gt]` filter and flatten documented `links` relationship IDs with `computed_fields`. Newly added streams use `projection: passthrough` with permissive schemas so documented fields are not silently dropped.

Cursor-paginated list endpoints use GoCardless's `meta.cursors.after` envelope and send `limit=50`, matching the legacy default page size. Detail endpoints and simple list endpoints such as `institutions` disable pagination. Path-scoped and selector streams require the corresponding config key named in `spec.json` when that stream is selected, for example `billing_request_id`, `payout_id`, or `balances_creditor_id`; these keys are optional globally because unrelated streams do not need them.

## Write actions & risks

The bundle declares 76 write actions for documented POST, PUT, and DELETE endpoints that the Tier-1 dialect can express as one HTTP request per record. GoCardless request bodies are wrapper-shaped in the public docs, so write records should include the documented top-level body key, such as `payments`, `customers`, `billing_requests`, or `data` for action endpoints. DELETE actions send no request body, treat 404 as idempotent missing-ok, and are marked `confirm: destructive`. Reverse ETL must still follow plan, preview, approval, and execute.

## Known limits

- `gocardless_environment`-based live/sandbox base-URL derivation is dropped; `base_url` is required instead. The engine can materialize fixed defaults but cannot derive one config value from another.
- `page_size` and `max_pages` remain legacy runtime concepts, but this declarative bundle uses the fixed documented page size of 50 for cursor lists and relies on empty `meta.cursors.after` for termination.
- Legacy `linkField` also falls back to top-level relationship fields; the documented GoCardless wire shape uses nested `links`, so the legacy streams model that documented shape directly.
- Newly added stream schemas are permissive passthrough schemas derived from documented response wrappers, not hand-curated warehouse schemas. The existing four legacy streams remain narrow to preserve emitted-record parity.
- Connect OAuth token exchange, introspection, and revocation endpoints use the separate `connect.gocardless.com` form flow and are excluded as `non_data_endpoint` in `api_surface.json`.
