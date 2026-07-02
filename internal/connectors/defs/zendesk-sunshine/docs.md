# Overview

Zendesk Sunshine's custom objects API lets a Zendesk account model arbitrary custom objects and
relationships between them. This bundle reads object types, objects, and relationships
(`GET {base_url}/objects/types`, `/objects/records`, `/relationships/records`). It migrates
`internal/connectors/zendesk-sunshine` (the hand-written legacy connector) to a declarative Tier-1
bundle at capability parity; the legacy package stays registered and unchanged until wave6's
registry flip.

## Auth setup

Requires `email` (config) and `api_token` (secret), combined into HTTP Basic auth as
`{email}/token` / `api_token` — the same "email/token" Zendesk API-token convention as
`zendesk-support` (`zendesk_sunshine.go:119`'s `connsdk.Basic(email+"/token", token)`). `base_url`
is required explicitly by this bundle (see Known limits for why the legacy subdomain-derivation
shortcut is not modeled).

## Streams notes

All three streams (`object_types`, `objects`, `relationships`) are single-page GET reads with no
pagination and no incremental support — legacy performs one unconditional request per stream and
emits every record from the response's top-level `data` array; this bundle does the same
(`records.path: "data"`, no `pagination` block declared).

`object_types`' raw API record carries its identifier under the field name `key`, not `id`; a
`computed_fields` rename (`"id": "{{ record.key }}"`) maps it to the schema's `id` property,
matching legacy's `mapObjectType` output field name exactly. `objects` and `relationships` records
already carry a natural `id` field matching the schema (Zendesk Sunshine's own JSON:API-flavored
wire shape), so plain schema projection copies it through with no rename needed; legacy's
`first(item["id"], item["key"])` fallback to `key` only matters for a record that omits `id`
entirely, which the documented Sunshine wire shape never does for these two streams — see Known
limits.

## Write actions & risks

None. `capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation` unconditionally.

## Known limits

- **`base_url` subdomain-derivation shortcut is not modeled.** Legacy derives `base_url` as:
  an explicit `base_url` config value, else (when `subdomain` is set)
  `https://{subdomain}.zendesk.com/api/sunshine`, else a hardcoded
  `https://example.zendesk.com/api/sunshine` default (`zendesk_sunshine.go:157-164`). The engine's
  `spec.json` `"default"` materialization mechanism only fills in a FIXED literal when a key is
  absent — it has no support for a default derived from another config key's value (the same
  documented limitation as sentry's hostname-derived URL / chargebee's site-derived URL, per
  `docs/migration/conventions.md` §3). This bundle therefore requires `base_url` explicitly and does
  not declare `subdomain` in `spec.json` at all (an undeclared-and-unwireable key would be dead
  config, per the "declared config must be consumed" rule) — a documented config-surface narrowing,
  not a silent behavior change: any caller that previously relied on subdomain-only configuration
  must now supply the full `base_url`.
- **`objects`/`relationships`' `key`-fallback id derivation is not modeled.** Legacy's `mapObject`/
  `mapRelationship` compute `id` as `first(item["id"], item["key"])`. Zendesk Sunshine's documented
  API always returns `id` for object and relationship records, so this fallback is
  defensive/unreachable on the real wire shape; the declarative dialect has no
  ordered-multi-field-fallback primitive (only a single bare `{{ record.<path> }}` reference or a
  filter chain), so only the `id`-present case is expressed. Deliberately out-of-scope edge case,
  not a defect.
