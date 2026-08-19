# pm connectors inspect ezofficeinventory

```text
NAME
  pm connectors inspect ezofficeinventory - EZOfficeInventory connector manual

SYNOPSIS
  pm connectors inspect ezofficeinventory
  pm connectors inspect ezofficeinventory --json
  pm credentials add <name> --connector ezofficeinventory [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes EZOfficeInventory assets, inventory items, stock assets, members, locations, groups, vendors, and purchase orders through the EZOfficeInventory REST API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  mode
  subdomain (required)
  api_key (secret) (required)

ETL STREAMS
  assets:
    primary key: identifier
    fields: asset_type(string), assigned_to_user_email(string), assigned_to_user_name(string), created_at(string), description(string), group_id(integer), identifier(integer), location_id(integer), location_name(string), name(string), price(string), purchased_on(string), updated_at(string)
  inventories:
    primary key: identifier
    fields: asset_type(string), created_at(string), description(string), group_id(integer), identifier(integer), location_id(integer), location_name(string), name(string), net_quantity(integer), price(string), updated_at(string)
  asset_stocks:
    primary key: identifier
    fields: asset_type(string), assigned_to_user_email(string), assigned_to_user_name(string), created_at(string), description(string), group_id(integer), identifier(integer), location_id(integer), location_name(string), name(string), price(string), purchased_on(string), updated_at(string)
  members:
    primary key: id
    fields: contact_type(string), country(string), created_at(string), email(string), first_name(string), full_name(string), id(integer), last_name(string), role_id(integer), role_name(string), status(string)
  locations:
    primary key: id
    fields: city(string), country(string), created_at(string), description(string), id(integer), name(string), parent_id(integer), state(string), status(string), street1(string), street2(string), updated_at(string), zipcode(string)
  groups:
    primary key: id
    fields: active(boolean), asset_depreciation_mode(string), assets_count(integer), company_id(integer), created_at(string), description(string), hidden_on_web_store(boolean), id(integer), name(string), updated_at(string)
  vendors:
    primary key: id
    fields: assets_count(integer), company_id(integer), created_at(string), id(integer), name(string), services_count(integer), status(integer), updated_at(string)
  purchase_orders:
    primary key: id
    fields: approver_type(string), company_id(integer), created_at(string), created_by_id(integer), id(integer), net_amount(string), paid_amount(string), payable_amount(string), po_type(string), requested_by_id(integer), sequence_num(integer), state(string), updated_at(string), vendor_id(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  create_asset:
    endpoint: POST /assets.api
    required fields: fixed_asset[name], fixed_asset[group_id], fixed_asset[location_id]
    risk: external mutation; creates a new asset record; approval required
  update_asset:
    endpoint: PUT /assets/{{ record.id }}.api
    required fields: id
    risk: external mutation; approval required
  create_member:
    endpoint: POST /members.api
    required fields: user[email], user[first_name], user[last_name], user[role_id]
    risk: external mutation; creates a new member/user account; approval required
  update_member:
    endpoint: PUT /members/{{ record.id }}.api
    required fields: id
    risk: external mutation; approval required
  create_location:
    endpoint: POST /locations.api
    required fields: location[name]
    risk: external mutation; creates a new location; approval required
  update_location:
    endpoint: PUT /locations/{{ record.id }}.api
    required fields: id
    risk: external mutation; approval required
  create_group:
    endpoint: POST /groups.api
    required fields: group[name]
    risk: external mutation; creates a new asset group/classification; approval required
  update_group:
    endpoint: PUT /groups/{{ record.id }}.api
    required fields: id
    risk: external mutation; approval required
  create_vendor:
    endpoint: POST /vendors.api
    required fields: vendor[name]
    risk: external mutation; creates a new vendor; approval required
  update_vendor:
    endpoint: PUT /vendors/{{ record.id }}.api
    required fields: id
    risk: external mutation; approval required
  create_purchase_order:
    endpoint: POST /purchase_orders.api
    required fields: vendor_id
    risk: external mutation; creates a new purchase order (financial document); approval required

SECURITY
  read risk: external EZOfficeInventory API read of asset, inventory, member, location, group, vendor, and purchase order data
  write risk: external mutation of asset, member, location, group, vendor, and purchase order records; create/update only, no delete actions implemented
  approval: writes require approval; reads are unrestricted
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect ezofficeinventory

  # Inspect as structured JSON
  pm connectors inspect ezofficeinventory --json

AGENT WORKFLOW
  - Run pm connectors inspect ezofficeinventory before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
