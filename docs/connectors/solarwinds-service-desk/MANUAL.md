# pm connectors inspect solarwinds-service-desk

```text
NAME
  pm connectors inspect solarwinds-service-desk - SolarWinds Service Desk connector manual

SYNOPSIS
  pm connectors inspect solarwinds-service-desk
  pm connectors inspect solarwinds-service-desk --json
  pm credentials add <name> --connector solarwinds-service-desk [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads SolarWinds Service Desk incidents, problems, changes, change catalogs, releases, solutions, catalog items, configuration items, users, sites, departments, roles, groups, categories, hardware/mobile/other/software assets, printers, contracts, purchase orders, vendors, audits, and risks; writes delete actions for every resource with a documented delete-by-id endpoint.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  page
  per_page
  start_date
  api_key (secret)
  api_key_2 (secret)

ETL STREAMS
  incidents:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  users:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  departments:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  categories:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  problems:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  changes:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  change_catalogs:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  releases:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  solutions:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  catalog_items:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  configuration_items:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  sites:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  roles:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  groups:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  hardwares:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  mobiles:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  other_assets:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  softwares:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  printers:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  contracts:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  purchase_orders:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  vendors:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  audits:
    primary key: id
    fields: created_at(), id(), name(), updated_at()
  risks:
    primary key: id
    fields: created_at(), id(), name(), updated_at()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  delete_incident:
    endpoint: DELETE /incidents/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk incident record; approval required
  delete_problem:
    endpoint: DELETE /problems/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk problem record; approval required
  delete_change:
    endpoint: DELETE /changes/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk change record; approval required
  delete_change_catalog:
    endpoint: DELETE /change_catalogs/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk change catalog record; approval required
  delete_release:
    endpoint: DELETE /releases/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk release record; approval required
  delete_solution:
    endpoint: DELETE /solutions/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk solution record; approval required
  delete_catalog_item:
    endpoint: DELETE /catalog_items/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk catalog item record; approval required
  delete_configuration_item:
    endpoint: DELETE /configuration_items/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk configuration item record; approval required
  delete_user:
    endpoint: DELETE /users/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk user record; approval required
  delete_site:
    endpoint: DELETE /sites/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk site record; approval required
  delete_department:
    endpoint: DELETE /departments/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk department record; approval required
  delete_role:
    endpoint: DELETE /roles/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk role record; approval required
  delete_group:
    endpoint: DELETE /groups/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk group record; approval required
  delete_category:
    endpoint: DELETE /categories/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk category record; approval required
  delete_hardware:
    endpoint: DELETE /hardwares/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk hardware record; approval required
  delete_mobile:
    endpoint: DELETE /mobiles/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk mobile record; approval required
  delete_other_asset:
    endpoint: DELETE /other_assets/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk other asset record; approval required
  delete_contract:
    endpoint: DELETE /contracts/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk contract record; approval required
  delete_purchase_order:
    endpoint: DELETE /purchase_orders/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk purchase order record; approval required
  delete_vendor:
    endpoint: DELETE /vendors/{{ record.id }}.json
    required fields: id
    risk: external mutation; permanently deletes a SolarWinds Service Desk vendor record; approval required

SECURITY
  read risk: external SolarWinds Service Desk API read of incident, problem, change, asset, and organizational (user/site/department/role/group) data
  write risk: external SolarWinds Service Desk API delete mutations against incidents, problems, changes, assets, and organizational records
  approval: required for all write actions; every write action is an irreversible external delete
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect solarwinds-service-desk

  # Inspect as structured JSON
  pm connectors inspect solarwinds-service-desk --json

AGENT WORKFLOW
  - Run pm connectors inspect solarwinds-service-desk before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
