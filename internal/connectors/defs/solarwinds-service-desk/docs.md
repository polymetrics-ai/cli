# Overview

Reads SolarWinds Service Desk incidents, problems, changes, change catalogs, releases, solutions,
catalog items, configuration items, users, sites, departments, roles, groups, categories,
hardware/mobile/other/software assets, printers, contracts, purchase orders, vendors, audits, and
risks; writes delete actions for every resource with a documented delete-by-id endpoint.

Readable streams: `incidents`, `users`, `departments`, `categories`, `problems`, `changes`,
`change_catalogs`, `releases`, `solutions`, `catalog_items`, `configuration_items`, `sites`,
`roles`, `groups`, `hardwares`, `mobiles`, `other_assets`, `softwares`, `printers`, `contracts`,
`purchase_orders`, `vendors`, `audits`, `risks`.

Write actions: `delete_incident`, `delete_problem`, `delete_change`, `delete_change_catalog`,
`delete_release`, `delete_solution`, `delete_catalog_item`, `delete_configuration_item`,
`delete_user`, `delete_site`, `delete_department`, `delete_role`, `delete_group`, `delete_category`,
`delete_hardware`, `delete_mobile`, `delete_other_asset`, `delete_contract`,
`delete_purchase_order`, `delete_vendor`.

Service API documentation: https://apidoc.samanage.com/.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); SolarWinds Service Desk API key (fallback name), sent as a
  Bearer token when api_key_2 is not configured. Never logged.
- `api_key_2` (optional, secret, string); SolarWinds Service Desk API key, sent as a Bearer token.
  Never logged.
- `base_url` (optional, string); default `https://api.samanage.com`; format `uri`; SolarWinds
  Service Desk (Samanage) API base URL override for tests or proxies.
- `mode` (optional, string).
- `page` (optional, string).
- `per_page` (optional, string).
- `start_date` (optional, string); format `date-time`.

Secret fields are redacted in logs and write previews: `api_key`, `api_key_2`.

Default configuration values: `base_url=https://api.samanage.com`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key_2` when `{{ secrets.api_key_2 }}`.
- Bearer token authentication using `secrets.api_key` when `{{ secrets.api_key }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/incidents.json`.

## Streams notes

Default pagination: single request; no pagination.

- `incidents`: GET `/incidents.json` - records at response root; query `page` from template `{{
  config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted
  when absent; `updated_after` from template `{{ config.start_date }}`, omitted when absent; emits
  passthrough records.
- `users`: GET `/users.json` - records at response root; query `page` from template `{{ config.page
  }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted when absent;
  `updated_after` from template `{{ config.start_date }}`, omitted when absent; emits passthrough
  records.
- `departments`: GET `/departments.json` - records at response root; query `page` from template `{{
  config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted
  when absent; `updated_after` from template `{{ config.start_date }}`, omitted when absent; emits
  passthrough records.
- `categories`: GET `/categories.json` - records at response root; query `page` from template `{{
  config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted
  when absent; `updated_after` from template `{{ config.start_date }}`, omitted when absent; emits
  passthrough records.
- `problems`: GET `/problems.json` - records at response root; query `page` from template `{{
  config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted
  when absent; `updated_after` from template `{{ config.start_date }}`, omitted when absent; emits
  passthrough records.
- `changes`: GET `/changes.json` - records at response root; query `page` from template `{{
  config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted
  when absent; `updated_after` from template `{{ config.start_date }}`, omitted when absent; emits
  passthrough records.
- `change_catalogs`: GET `/change_catalogs.json` - records at response root; query `page` from
  template `{{ config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page
  }}`, omitted when absent; emits passthrough records.
- `releases`: GET `/releases.json` - records at response root; query `page` from template `{{
  config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted
  when absent; emits passthrough records.
- `solutions`: GET `/solutions.json` - records at response root; query `page` from template `{{
  config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted
  when absent; emits passthrough records.
- `catalog_items`: GET `/catalog_items.json` - records at response root; query `page` from template
  `{{ config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`,
  omitted when absent; emits passthrough records.
- `configuration_items`: GET `/configuration_items.json` - records at response root; query `page`
  from template `{{ config.page }}`, omitted when absent; `per_page` from template `{{
  config.per_page }}`, omitted when absent; emits passthrough records.
- `sites`: GET `/sites.json` - records at response root; query `page` from template `{{ config.page
  }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted when absent;
  emits passthrough records.
- `roles`: GET `/roles.json` - records at response root; query `page` from template `{{ config.page
  }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted when absent;
  emits passthrough records.
- `groups`: GET `/groups.json` - records at response root; query `page` from template `{{
  config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted
  when absent; emits passthrough records.
- `hardwares`: GET `/hardwares.json` - records at response root; query `page` from template `{{
  config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted
  when absent; emits passthrough records.
- `mobiles`: GET `/mobiles.json` - records at response root; query `page` from template `{{
  config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted
  when absent; emits passthrough records.
- `other_assets`: GET `/other_assets.json` - records at response root; query `page` from template
  `{{ config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`,
  omitted when absent; emits passthrough records.
- `softwares`: GET `/softwares.json` - records at response root; query `page` from template `{{
  config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted
  when absent; emits passthrough records.
- `printers`: GET `/printers.json` - records at response root; query `page` from template `{{
  config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted
  when absent; emits passthrough records.
- `contracts`: GET `/contracts.json` - records at response root; query `page` from template `{{
  config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted
  when absent; emits passthrough records.
- `purchase_orders`: GET `/purchase_orders.json` - records at response root; query `page` from
  template `{{ config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page
  }}`, omitted when absent; emits passthrough records.
- `vendors`: GET `/vendors.json` - records at response root; query `page` from template `{{
  config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted
  when absent; emits passthrough records.
- `audits`: GET `/audits.json` - records at response root; query `page` from template `{{
  config.page }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted
  when absent; emits passthrough records.
- `risks`: GET `/risks.json` - records at response root; query `page` from template `{{ config.page
  }}`, omitted when absent; `per_page` from template `{{ config.per_page }}`, omitted when absent;
  emits passthrough records.

## Write actions & risks

Overall write risk: external SolarWinds Service Desk API delete mutations against incidents,
problems, changes, assets, and organizational records.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `delete_incident`: DELETE `/incidents/{{ record.id }}.json` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  permanently deletes a SolarWinds Service Desk incident record; approval required.
- `delete_problem`: DELETE `/problems/{{ record.id }}.json` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  permanently deletes a SolarWinds Service Desk problem record; approval required.
- `delete_change`: DELETE `/changes/{{ record.id }}.json` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  permanently deletes a SolarWinds Service Desk change record; approval required.
- `delete_change_catalog`: DELETE `/change_catalogs/{{ record.id }}.json` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: external
  mutation; permanently deletes a SolarWinds Service Desk change catalog record; approval required.
- `delete_release`: DELETE `/releases/{{ record.id }}.json` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  permanently deletes a SolarWinds Service Desk release record; approval required.
- `delete_solution`: DELETE `/solutions/{{ record.id }}.json` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  permanently deletes a SolarWinds Service Desk solution record; approval required.
- `delete_catalog_item`: DELETE `/catalog_items/{{ record.id }}.json` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: external
  mutation; permanently deletes a SolarWinds Service Desk catalog item record; approval required.
- `delete_configuration_item`: DELETE `/configuration_items/{{ record.id }}.json` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk:
  external mutation; permanently deletes a SolarWinds Service Desk configuration item record;
  approval required.
- `delete_user`: DELETE `/users/{{ record.id }}.json` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; risk: external mutation; permanently
  deletes a SolarWinds Service Desk user record; approval required.
- `delete_site`: DELETE `/sites/{{ record.id }}.json` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; risk: external mutation; permanently
  deletes a SolarWinds Service Desk site record; approval required.
- `delete_department`: DELETE `/departments/{{ record.id }}.json` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  permanently deletes a SolarWinds Service Desk department record; approval required.
- `delete_role`: DELETE `/roles/{{ record.id }}.json` - kind `delete`; body type `none`; path fields
  `id`; required record fields `id`; accepted fields `id`; risk: external mutation; permanently
  deletes a SolarWinds Service Desk role record; approval required.
- `delete_group`: DELETE `/groups/{{ record.id }}.json` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  permanently deletes a SolarWinds Service Desk group record; approval required.
- `delete_category`: DELETE `/categories/{{ record.id }}.json` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  permanently deletes a SolarWinds Service Desk category record; approval required.
- `delete_hardware`: DELETE `/hardwares/{{ record.id }}.json` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  permanently deletes a SolarWinds Service Desk hardware record; approval required.
- `delete_mobile`: DELETE `/mobiles/{{ record.id }}.json` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  permanently deletes a SolarWinds Service Desk mobile record; approval required.
- `delete_other_asset`: DELETE `/other_assets/{{ record.id }}.json` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: external
  mutation; permanently deletes a SolarWinds Service Desk other asset record; approval required.
- `delete_contract`: DELETE `/contracts/{{ record.id }}.json` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  permanently deletes a SolarWinds Service Desk contract record; approval required.
- `delete_purchase_order`: DELETE `/purchase_orders/{{ record.id }}.json` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk: external
  mutation; permanently deletes a SolarWinds Service Desk purchase order record; approval required.
- `delete_vendor`: DELETE `/vendors/{{ record.id }}.json` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: external mutation;
  permanently deletes a SolarWinds Service Desk vendor record; approval required.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 24 stream-backed endpoint group(s), 20 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=1, duplicate_of=23, out_of_scope=67.
