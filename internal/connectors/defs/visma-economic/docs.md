# Overview

Visma e-conomic is a Danish/Nordic accounting SaaS. This bundle reads the customer directory from
the e-conomic REST API (`GET {base_url}/customers`). It migrates
`internal/connectors/visma-economic` (the hand-written legacy connector), which stays registered
and unchanged until wave6's registry flip. Read-only; a single `customers` stream, no pagination
and no incremental filtering (matching legacy exactly).

## Auth setup

Two secrets are required: `app_secret_token` and `agreement_grant_token`, e-conomic's own
two-token app authentication scheme. Both are sent as static request headers
(`X-AppSecretToken` / `X-AgreementGrantToken`) on every request via `streams.json`'s `base.headers`
— e-conomic does not use a Bearer/Basic/API-key-query scheme, so `base.auth` is declared as a single
unconditional `{"mode": "none"}` (the credentials flow entirely through the two headers, not
through the `auth` dispatch). Both secrets are required in `spec.json`; an absent header-templated
secret is always a hard validate/runtime error (per the engine's header-resolution rule), matching
legacy's own `Check`/`requester` validation that rejects an empty `app_secret_token` or
`agreement_grant_token`.

`base_url` defaults to `https://restapi.e-conomic.com`, matching legacy's `defaultBaseURL` constant,
materialized via `spec.json`'s `"default"` value.

## Streams notes

`customers` reads `GET /customers` and extracts records from the response's top-level `collection`
array (e-conomic's list envelope), matching legacy's `connsdk.RecordsAt(resp.Body, "collection")`.
Legacy maps `id` from the raw `customerNumber` field, always stringified via `fmt.Sprint`
(`text(item["customerNumber"])`) regardless of its raw JSON type (e-conomic's wire shape is a bare
integer). This bundle reproduces that exact stringification with `computed_fields`'
`"id": "{{ record.customerNumber | last_path_segment }}"` — the `last_path_segment` filter forces
string output via `Interpolate` (a bare `{{ record.customerNumber }}` reference would instead trigger
typed extraction and copy the raw JSON integer, which is not what legacy emits) while passing a
delimiter-free numeric value through unchanged. `name`/`currency` are copied through by plain schema
projection (exact key match, no rename needed). There is no pagination and no incremental
cursor/filter on this stream in legacy, so neither is declared here.

## Write actions & risks

None. `capabilities.write` is `false`; legacy's `Write` is an unconditional
`connectors.ErrUnsupportedOperation` stub, and this bundle ships no `writes.json`.

## Known limits

- Only the `customers` stream is migrated (legacy's own only stream). e-conomic's much larger API
  surface (invoices, products, orders, accounting years, etc.) is intentionally out of scope for
  this wave — see `api_surface.json`.
- No pagination or incremental read support: legacy's `Read` issues exactly one unpaginated request
  per sync and ignores any `start_date`/cursor config for this connector (it has none), so this
  bundle declares neither a `pagination` block nor an `incremental` block, matching legacy's actual
  behavior exactly (not a scope narrowing — there is nothing to narrow).
