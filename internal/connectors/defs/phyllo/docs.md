# Overview

Phyllo is a wave2 fan-out declarative-HTTP migration. It conservatively reads Phyllo users,
accounts, profiles, and social contents through stable REST list endpoints
(`GET https://api.getphyllo.com/v1/...`) — legacy's own package doc notes Phyllo's public docs are
JS-backed, so only endpoints documented by connector references are modeled. This bundle is
engine-vs-legacy parity-tested against `internal/connectors/phyllo` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Phyllo `client_id` and `client_secret` secret pair; they are sent as HTTP Basic auth
(`Authorization: Basic base64(client_id:client_secret)`) and are never logged, matching legacy's
`connsdk.Basic(id, secretValue)` (`phyllo.go:127`). `base_url` defaults to the production host
`https://api.getphyllo.com` and may be set explicitly to
`https://api.sandbox.getphyllo.com`/`https://api.staging.getphyllo.com` to target a non-production
environment, or overridden entirely for tests/proxies.

## Streams notes

All 4 streams (`users`, `accounts`, `profiles`, `social_contents`) share the same shape: `GET`
against the Phyllo v1 list endpoint, records at the top-level `data` key, primary key `["id"]`, and
legacy's identical common field set (`id`, `platform`, `status`, `created_at`, `updated_at` —
`phyllo.go:104-106`). Pagination is offset+limit (`pagination.type: offset_limit`, `limit_param:
limit`, `offset_param: offset`, `page_size: 50`), stopping on a short page — identical to legacy's
`connsdk.OffsetPaginator{LimitParam: "limit", OffsetParam: "offset", PageSize: size}`. None of
legacy's 4 streams declare an incremental cursor field, so this bundle declares no `incremental`
block for any stream, matching legacy exactly (full-refresh reads only).

Legacy's `Read` passes `connsdk.Harvest` a callback that emits every record verbatim
(`func(rec connsdk.Record) error { return emit(connectors.Record(rec)) }`, `phyllo.go:86`) — there
is no `mapRecord`-style field-building or filtering anywhere in the read path; `commonFields()`
(`phyllo.go:104-106`) only documents legacy's `Catalog` metadata, it never gates what `Read` emits.
Every stream therefore declares `"projection": "passthrough"` (conventions.md §8 rule 1) so the
engine's default schema-mode projection does not silently drop any field of the actual Phyllo
response object. Each `schemas/*.json`'s `id`/`platform`/`status`/`created_at`/`updated_at`
properties remain a documentation surface describing the common field set, not an allow-list.

## Write actions & risks

None. Legacy `phyllo.Write` always returns `connectors.ErrUnsupportedOperation`;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`environment`-based base-URL derivation is dropped; `base_url` must be set explicitly for
  non-production hosts.** Legacy derives the effective base URL from a separate `environment`
  config value (`"api.sandbox"` -> `https://api.sandbox.getphyllo.com`, `"api.staging"` ->
  `https://api.staging.getphyllo.com`, anything else -> the production default) only when
  `base_url` itself is unset (`phyllo.go:129-142`). The engine's `spec.json` `"default"`
  materialization mechanism (conventions.md §3) fills in a FIXED literal for a genuinely-absent
  key; it cannot express "the default value is a function of ANOTHER config key's value" (the same
  documented limitation as sentry's hostname-derived URL and chargebee's site-derived URL,
  conventions.md §3's `spec.json "default"` section). This bundle therefore requires the caller to
  set `base_url` directly to the desired sandbox/staging/production host — a documented
  config-surface narrowing, not a silent behavior change: every legacy-accepted final base URL
  (production default, sandbox, staging, or a raw override) remains reachable, just via one
  config key (`base_url`) instead of two (`base_url` OR `environment`). `environment` is not
  declared in this bundle's `spec.json` (F6, REVIEW.md: a declared-but-unwireable config key is
  worse than an absent one).
- Full Phyllo API surface (user creation, social comments) is out of scope; see
  `api_surface.json`'s `excluded: {category: out_of_scope}` entries — legacy itself never
  implemented these.
