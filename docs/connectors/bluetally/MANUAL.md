# pm connectors inspect bluetally

```text
NAME
  pm connectors inspect bluetally - BlueTally connector manual

SYNOPSIS
  pm connectors inspect bluetally
  pm connectors inspect bluetally --json
  pm credentials add <name> --connector bluetally [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads BlueTally IT asset management data (assets, employees, licenses, maintenances, accessories) through the BlueTally REST API.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  api_key (secret)

ETL STREAMS
  assets:
    primary key: id
    cursor: updated_at
    fields: asset_id(), asset_name(), asset_serial(), category_id(), category_name(), created_at(), currency(), department_id(), id(), location_id(), notes(), product_id(), product_name(), purchase_cost(), purchase_date(), status_id(), supplier_id(), updated_at(), warranty_expiration_date()
  employees:
    primary key: id
    cursor: updated_at
    fields: archived(), created_at(), department_id(), email(), id(), location_id(), manager_id(), name(), notes(), number_of_accessories(), number_of_assets(), number_of_consumables(), number_of_licenses(), title(), updated_at()
  licenses:
    primary key: id
    cursor: updated_at
    fields: available(), category_id(), created_at(), currency(), department_id(), expiration_date(), id(), license_type(), licensed_to_email(), licensed_to_name(), location_id(), manufacturer_id(), minimum_seats(), name(), notes(), number_of_seats(), order_number(), purchase_cost(), purchase_date(), supplier_id(), termination_date(), unit_cost(), updated_at()
  maintenances:
    primary key: id
    cursor: updated_at
    fields: asset_id(), cost(), created_at(), end_date(), id(), name(), notes(), start_date(), supplier_id(), type(), updated_at()
  accessories:
    primary key: id
    cursor: updated_at
    fields: available(), category_id(), created_at(), currency(), department_id(), id(), location_id(), manufacturer_id(), model_number(), name(), notes(), purchase_cost(), purchase_date(), quantity(), supplier_id(), updated_at()
  components:
    primary key: id
    cursor: updated_at
    fields: available(), category_id(), checked_out_to(), created_at(), currency(), custom_fields(), department_id(), id(), location_id(), logo(), manufacturer_id(), minimum_quantity(), model_number(), name(), notes(), order_number(), purchase_cost(), purchase_date(), quantity(), supplier_id(), unit_cost(), updated_at()
  consumables:
    primary key: id
    cursor: updated_at
    fields: available(), category_id(), checked_out_to_employees(), checked_out_to_locations(), created_at(), currency(), custom_fields(), department_id(), id(), location_id(), logo(), manufacturer_id(), minimum_quantity(), model_number(), name(), notes(), order_number(), purchase_cost(), purchase_date(), supplier_id(), unit_cost(), updated_at()
  categories:
    primary key: id
    cursor: updated_at
    fields: accessories(), assets(), components(), consumables(), created_at(), eula(), id(), licenses(), logo(), minimum_quantity(), name(), number_of_accessories(), number_of_assets(), number_of_components(), number_of_consumables(), number_of_deployable_assets(), number_of_licenses(), number_of_products(), products(), skip_checkout_emails(), type(), updated_at()
  departments:
    primary key: id
    cursor: updated_at
    fields: accessories(), assets(), components(), consumables(), created_at(), email(), employees(), id(), licenses(), name(), number_of_accessories(), number_of_assets(), number_of_components(), number_of_consumables(), number_of_employees(), number_of_licenses(), phone(), updated_at()
  depreciations:
    primary key: id
    cursor: updated_at
    fields: assets(), created_at(), id(), licenses(), minimum_value(), months(), name(), number_of_assets(), number_of_licenses(), number_of_products(), products(), updated_at()
  locations:
    primary key: id
    cursor: updated_at
    fields: accessories(), address_line_1(), address_line_2(), assets(), checked_out_assets(), city(), components(), consumables(), country(), created_at(), currency(), custom_fields(), email(), employees(), id(), licenses(), logo(), name(), number_of_accessories(), number_of_assets(), number_of_checked_out_assets(), number_of_components(), number_of_consumables(), number_of_employees(), number_of_licenses(), phone(), state(), updated_at(), zip()
  manufacturers:
    primary key: id
    cursor: updated_at
    fields: accessories(), assets(), components(), consumables(), created_at(), id(), licenses(), logo(), name(), notes(), number_of_accessories(), number_of_assets(), number_of_components(), number_of_consumables(), number_of_licenses(), number_of_products(), products(), support_email(), support_phone(), support_url(), updated_at(), url()
  products:
    primary key: id
    cursor: updated_at
    fields: archived(), assets(), category_id(), created_at(), custom_fields(), default_purchase_cost(), depreciation_id(), end_of_life_date(), end_of_life_months(), end_of_life_type(), id(), logo(), manufacturer_id(), minimum_quantity(), model_number(), name(), notes(), number_of_assets(), number_of_deployable_assets(), updated_at()
  statuses:
    primary key: id
    cursor: updated_at
    fields: assets(), created_at(), id(), name(), notes(), number_of_assets(), show_in_nav(), type(), updated_at()
  suppliers:
    primary key: id
    cursor: updated_at
    fields: accessories(), address_line_1(), address_line_2(), assets(), city(), components(), consumables(), contact_name(), country(), created_at(), email(), fax(), id(), licenses(), logo(), maintenances(), name(), notes(), number_of_accessories(), number_of_assets(), number_of_components(), number_of_consumables(), number_of_licenses(), number_of_maintenances(), phone(), state(), updated_at(), url(), zip()
  audits:
    primary key: id
    cursor: updated_at
    fields: asset_id(), audit_date(), audit_failed_reason(), audit_status(), completed(), created_at(), id(), location_id(), next_audit_date(), notes(), scheduled(), updated_at(), user_id()
  activity:
    primary key: timestamp, item_id, event
    cursor: timestamp
    fields: checked_out_to_from_email(), checked_out_to_from_id(), checked_out_to_from_name(), checked_out_to_from_type(), event(), item_id(), item_name(), notes(), timestamp(), type(), user_email(), user_id(), user_name()
  tenants:
    primary key: tenant_id
    fields: tenant_id(), tenant_name()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external BlueTally API read of IT asset management data
  approval: none; read-only API
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect bluetally

  # Inspect as structured JSON
  pm connectors inspect bluetally --json

AGENT WORKFLOW
  - Run pm connectors inspect bluetally before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
