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
matching legacy's `mapObjectType` output field name exactly. `objects` and `relationships` compute
`id` as `{{ coalesce record.id record.key }}`, reproducing legacy's `first(item["id"], item["key"])`
fallback (first present, non-null value) field-for-field; only the `id`-present case occurs on the
documented Sunshine wire shape, but the coalesce keeps the `key`-fallback parity legacy emits.

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
- **`objects`/`relationships` emit only `id`, `type`, and (`attributes` | `source`+`target`).**
  Legacy's `mapObject`/`mapRelationship` project exactly those fields and drop the raw API's
  `external_id`/`type_version`/`created_at`/`updated_at` (objects) and `created_at`/`updated_at`
  (relationships); this bundle's schemas declare only the legacy-emitted properties so schema-mode
  projection matches legacy's emitted record DATA exactly, rather than widening the record with
  raw fields legacy deliberately dropped.
