# Overview

ConfigCat is a feature flag management platform. This bundle reads and writes ConfigCat data
through the ConfigCat Public Management API (`https://api.configcat.com`): the 5 legacy-parity
streams (organizations, products, configs, environments, tags) plus, as of this Pass B
full-surface expansion, 22 additional read streams (config/environment/setting/segment/webhook/
permission-group/integration/proxy-profile detail lookups, deleted settings, SDK keys,
config+environment setting values, segments/webhooks/permission-groups/audit-logs product-scoped
lists, members, stale flags, the authenticated user's own profile, and a tag detail lookup) and 12
write actions (create/update/delete for configs, environments, feature flags/settings, and tags).
It migrates `internal/connectors/configcat` (the hand-written legacy connector, which stays
registered and unchanged until wave6's registry flip) at parity for the original 5 streams; every
Pass B addition is new coverage with no legacy counterpart to match, verified directly against
ConfigCat's published OpenAPI 3.0 spec (`https://api.configcat.com/docs/v1/swagger.json`, linked
from the `docs_url` above).

This bundle was UNBLOCKED from `docs/migration/quarantine.json` once the engine gained the
`stream.fan_out` dialect (S4 engine mini-wave item 2) — legacy's `readNested` first lists
`/v1/products`, then issues one request per product id (`/v1/products/{id}/configs` etc.),
stamping `product_id` onto every nested record, which the pre-increment declarative dialect had
no mechanism to express short of a Tier-2 `StreamHook`.

## Auth setup

Provide a ConfigCat Public Management API password via the `password` secret; it is used only for
HTTP Basic auth and is never logged. The Basic auth username is resolved with the same precedence
as legacy's `configcatUsername`: the (non-secret) `username` config value if set, else a
defensively-checked `username` secret if set, else an empty username — expressed as three ordered
`basic` auth candidates gated by `when` (first-match-wins, matching legacy's own
config-then-secrets fallback order exactly). `base_url` defaults to
`https://api.configcat.com` and may be overridden for tests/proxies.

## Streams notes

`organizations` (`GET /v1/organizations`) and `products` (`GET /v1/products`) are flat, top-level
JSON array endpoints (`records.path: ""`), matching legacy's `readList` exactly; ConfigCat's
Public Management API paginates neither (legacy declares no pagination for either), so this
bundle declares no `pagination` block (`type: none`, the default).

`configs`, `environments`, and `tags` are nested-under-product resources: legacy's `readNested`
first lists every accessible product (`GET /v1/products`), then reads the sub-resource once per
product id, stamping `product_id` onto every record. This bundle reproduces that exact pattern
with `stream.fan_out`: `ids_from.request` issues a preliminary `GET /v1/products` listing
(the SAME endpoint the `products` stream itself reads, extracting `productId` off each record);
`into.path_var` makes the resolved product id referenceable in the stream's own `path` as
`{{ fanout.id }}` (e.g. `/v1/products/{{ fanout.id }}/configs`); `stamp_field: product_id` writes
the current product id onto every emitted record of that stream, after projection/computed_fields
— exactly matching legacy's `readList`'s conditional `rec["product_id"] = productID` stamp (the
stamped id and legacy's own nested-`product.productId`-derived fallback are always the same value
for records returned under that product's own endpoint, so the two approaches are behaviorally
identical for every record legacy itself would emit).

Every stream's `product_id`/`organization_id` cross-reference field is a renamed camelCase→snake_case
copy of the raw API field (`{{ record.organizationId }}`, `{{ record.product.productId }}`, etc.),
matching legacy's per-stream `mapRecord` functions field-for-field.

None of the 5 streams exposes a legacy-recognized incremental cursor field — ConfigCat's Public
Management API surfaces configuration metadata, not an event stream; legacy's own catalog
publishes no `CursorFields` for any stream. All 5 streams are full-refresh only.

`check` issues a single bounded `GET /v1/organizations`, mirroring legacy's `Check` implementation
exactly (a bounded read of the organizations list confirms auth and connectivity without
mutating anything).

### Pass B streams (22 new)

Config-driven detail lookups (one request each, no fan_out): `config` (`GET /v1/configs/
{config_id}`), `environment` (`GET /v1/environments/{environment_id}`), `settings`
(`GET /v1/configs/{config_id}/settings`, records at `""`), `setting` (`GET /v1/settings/
{setting_id}`), `deleted_settings` (`GET /v1/configs/{config_id}/deleted-settings`), `sdk_keys`
(`GET /v1/configs/{config_id}/environments/{environment_id}`, a single-object `{primary,
secondary}` response), `config_setting_values` (`GET /v1/configs/{config_id}/environments/
{environment_id}/values`), `segment` (`GET /v1/segments/{segment_id}`), `webhook`
(`GET /v1/webhooks/{webhook_id}`), `permission_group` (`GET /v1/permissions/
{permission_group_id}`), `integration` (`GET /v1/integrations/{integration_id}`, distinct from the
product-scoped `integrations` list, whose real response envelope is `{"integrations": [...]}` —
`records.path: "integrations"`), `proxy_profile` (`GET /v1/proxy-profiles/{proxy_profile_id}`),
`stale_flags` (`GET /v1/products/{product_id}/staleflags` — a single nested aggregate object per
product, `{productId, name, configs: [...], environments: [...]}` describing WHICH configs/
settings are stale per environment, not a flat list of stale-flag records; modeled as a
single-object stream, matching the API's own real shape rather than forcing a flat-list fiction),
`me` (`GET /v1/me`, the authenticated user's own `{email, fullName}`), `tag` (`GET /v1/tags/
{tag_id}`).

Product-scoped lists reusing the exact same `fan_out` pattern as `configs`/`environments`/`tags`
(preliminary `GET /v1/products` listing, `into.path_var`, `stamp_field: product_id`): `segments`
(`GET /v1/products/{{ fanout.id }}/segments`), `webhooks` (`.../webhooks`), `permission_groups`
(`.../permissions`), `audit_logs` (`.../auditlogs`, plus optional `configId`/`environmentId`
query filters via `config.audit_log_config_id`/`audit_log_environment_id`, both
`omit_when_absent`).

Organization-scoped, config-driven (not fanned out — ConfigCat's real API scopes these to a
single organization the caller names, not "every accessible organization"): `proxy_profiles`
(`GET /v1/organizations/{organization_id}/proxy-profiles`, records at `profiles`), `members`
(`GET /v1/organizations/{organization_id}/members`).

Product-scoped, config-driven (not fanned out — a per-product detail lookup, not a list to
enumerate across every product): `integrations` (`GET /v1/products/{product_id}/integrations`,
records at `integrations`; `product_id` is required for this stream even though the SAME
resource family's `segments`/`webhooks`/`permission_groups` streams fan out automatically, because
`integrations` returns one aggregate object per product, not naturally one-record-per-item without
already knowing which product to ask about — matching `stale_flags`' identical config-driven
per-product shape).

## Write actions & risks

12 actions across 4 resource families (`capabilities.write` is now `true`;
`metadata.json`'s `risk.write` summarizes the shared external-mutation exposure):

- **Configs** (`create_config`/`update_config`/`delete_config`): `create_config` requires `name`
  and posts to the configured `config.product_id`; `update_config`/`delete_config` path-template
  `{{ record.configId }}`. `delete_config` declares `delete.missing_ok_status: [404]` (idempotent
  delete) and cascades to every setting/flag defined in that config on the real API.
- **Environments** (`create_environment`/`update_environment`/`delete_environment`): same shape as
  configs, one level down; `delete_environment` cascades to every flag VALUE scoped to that
  environment (not the flag definitions themselves, which live at the config level).
- **Feature flags / settings** (`create_flag`/`update_flag`/`delete_flag`): `create_flag` requires
  `key`+`name`+`settingType` (`boolean`/`string`/`int`/`double`, ConfigCat's own `SettingType`
  enum) and posts to the configured `config.config_id`. `update_flag`/`delete_flag` mutate
  **METADATA ONLY** (name/hint/tags/order) — ConfigCat draws a hard line in its own API surface
  between a flag's metadata (modeled here, a `PUT /v1/settings/{settingId}` call) and its VALUE
  (the per-environment targeting rules/rollout percentages actually served to SDKs, at a
  completely different endpoint family, `/v1/settings/{settingKeyOrId}/value` and friends) — this
  bundle deliberately does NOT model flag-value mutation (see Known limits); `update_flag` can
  never change what a running application observes when it evaluates the flag.
- **Tags** (`create_tag`/`update_tag`/`delete_tag`): `create_tag` requires `name`, posts to the
  configured `config.product_id`; `delete_tag` untags every flag that referenced it (ConfigCat's
  own cascade behavior, not a bundle-side side effect).

## Known limits

- **Flag VALUE mutation is deliberately NOT modeled** — the single largest exclusion category in
  `api_surface.json` (`/v1/settings/{settingKeyOrId}/value` and its `/v1/environments/
  {environmentId}/settings/{settingId}/value` alias, both v1 and v2, plus the bulk
  `/v1/configs/{configId}/environments/{environmentId}/values` POST): ConfigCat's real targeting-
  rule body is a complex, JSON-Patch-shaped structure (percentage rollout rules, user-targeting
  comparators, a running `version`/optimistic-concurrency counter) that a flat `record_schema`
  write action cannot safely represent without risking silent corruption of live flag evaluation
  logic for applications currently reading that flag — this is a genuinely different risk profile
  from metadata mutation (renaming a flag can't break a live app; replacing its targeting rules
  incorrectly can). `update_flag`/`create_flag`/`delete_flag` in this bundle touch ONLY metadata.
- **Segment mutation (update/delete/create) is deliberately NOT modeled**, for the identical
  live-behavior-safety reason as flag values: a segment's comparison-rule definition is referenced
  by the targeting rules of every flag that uses it, so mutating one changes evaluation behavior
  for every referencing flag across every environment simultaneously.
- Organization/product-admin-only surfaces (member/invitation management, permission-group
  mutation, integration/webhook/Jira wiring, Proxy-profile deployment configuration, code-reference
  upload) are excluded as `requires_elevated_scope`/`out_of_scope` — see `api_surface.json` for the
  specific reason on each excluded endpoint.
- `configs`/`environments`/`tags`/`segments`/`webhooks`/`permission_groups`/`audit_logs` all fan
  out across every accessible product; a workspace with many products issues one request per
  product per stream per sync, matching legacy's own `readNested` cost profile exactly for the
  original 3 fan_out streams, extended consistently to the 4 new ones (no new request-count
  regression introduced by this migration).
- `fixtures/streams/{configs,environments,tags,segments,webhooks,permission_groups,audit_logs}/
  page_1.json` each record the preliminary `/v1/products` listing; `page_2.json` (and `page_3.json`
  for the original 3) record the per-product sub-resource response, exercising the fan-out path
  under `conformance`'s replay harness end to end (mirrors `cisco-meraki`'s identical fan-out
  fixture shape). Every other Pass B stream is a single-request config-driven detail lookup with
  exactly one fixture page (no pagination declared anywhere in this bundle — ConfigCat's Public
  Management API paginates none of its list endpoints this bundle covers).
