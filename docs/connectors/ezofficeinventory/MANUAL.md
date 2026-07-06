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
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  mode
  subdomain
  api_key (secret)

ETL STREAMS
  assets:
    primary key: identifier
    fields: asset_type(), assigned_to_user_email(), assigned_to_user_name(), created_at(), description(), group_id(), identifier(), location_id(), location_name(), name(), price(), purchased_on(), updated_at()
  inventories:
    primary key: identifier
    fields: asset_type(), created_at(), description(), group_id(), identifier(), location_id(), location_name(), name(), net_quantity(), price(), updated_at()
  asset_stocks:
    primary key: identifier
    fields: asset_type(), assigned_to_user_email(), assigned_to_user_name(), created_at(), description(), group_id(), identifier(), location_id(), location_name(), name(), price(), purchased_on(), updated_at()
  members:
    primary key: id
    fields: contact_type(), country(), created_at(), email(), first_name(), full_name(), id(), last_name(), role_id(), role_name(), status()
  locations:
    primary key: id
    fields: city(), country(), created_at(), description(), id(), name(), parent_id(), state(), status(), street1(), street2(), updated_at(), zipcode()
  groups:
    primary key: id
    fields: active(), asset_depreciation_mode(), assets_count(), company_id(), created_at(), description(), hidden_on_web_store(), id(), name(), updated_at()
  vendors:
    primary key: id
    fields: assets_count(), company_id(), created_at(), id(), name(), services_count(), status(), updated_at()
  purchase_orders:
    primary key: id
    fields: approver_type(), company_id(), created_at(), created_by_id(), id(), net_amount(), paid_amount(), payable_amount(), po_type(), requested_by_id(), sequence_num(), state(), updated_at(), vendor_id()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_asset:
    endpoint: POST /assets.api
    risk: external mutation; creates a new asset record; approval required
  update_asset:
    endpoint: PUT /assets/{{ record.id }}.api
    required fields: id
    risk: external mutation; approval required
  create_member:
    endpoint: POST /members.api
    risk: external mutation; creates a new member/user account; approval required
  update_member:
    endpoint: PUT /members/{{ record.id }}.api
    required fields: id
    risk: external mutation; approval required
  create_location:
    endpoint: POST /locations.api
    risk: external mutation; creates a new location; approval required
  update_location:
    endpoint: PUT /locations/{{ record.id }}.api
    required fields: id
    risk: external mutation; approval required
  create_group:
    endpoint: POST /groups.api
    risk: external mutation; creates a new asset group/classification; approval required
  update_group:
    endpoint: PUT /groups/{{ record.id }}.api
    required fields: id
    risk: external mutation; approval required
  create_vendor:
    endpoint: POST /vendors.api
    risk: external mutation; creates a new vendor; approval required
  update_vendor:
    endpoint: PUT /vendors/{{ record.id }}.api
    required fields: id
    risk: external mutation; approval required
  create_purchase_order:
    endpoint: POST /purchase_orders.api
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
