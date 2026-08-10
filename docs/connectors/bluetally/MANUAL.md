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
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  api_key (secret) (required)

ETL STREAMS
  assets:
    primary key: id
    cursor: updated_at
    fields: asset_id(string), asset_name(string), asset_serial(string), category_id(integer), category_name(string), created_at(string), currency(string), department_id(integer), id(integer), location_id(integer), notes(string), product_id(integer), product_name(string), purchase_cost(number), purchase_date(string), status_id(integer), supplier_id(integer), updated_at(string), warranty_expiration_date(string)
  employees:
    primary key: id
    cursor: updated_at
    fields: archived(boolean), created_at(string), department_id(integer), email(string), id(integer), location_id(integer), manager_id(integer), name(string), notes(string), number_of_accessories(integer), number_of_assets(integer), number_of_consumables(integer), number_of_licenses(integer), title(string), updated_at(string)
  licenses:
    primary key: id
    cursor: updated_at
    fields: available(integer), category_id(integer), created_at(string), currency(string), department_id(integer), expiration_date(string), id(integer), license_type(string), licensed_to_email(string), licensed_to_name(string), location_id(integer), manufacturer_id(integer), minimum_seats(integer), name(string), notes(string), number_of_seats(integer), order_number(string), purchase_cost(number), purchase_date(string), supplier_id(integer), termination_date(string), unit_cost(number), updated_at(string)
  maintenances:
    primary key: id
    cursor: updated_at
    fields: asset_id(integer), cost(number), created_at(string), end_date(string), id(integer), name(string), notes(string), start_date(string), supplier_id(integer), type(string), updated_at(string)
  accessories:
    primary key: id
    cursor: updated_at
    fields: available(integer), category_id(integer), created_at(string), currency(string), department_id(integer), id(integer), location_id(integer), manufacturer_id(integer), model_number(string), name(string), notes(string), purchase_cost(number), purchase_date(string), quantity(integer), supplier_id(integer), updated_at(string)
  components:
    primary key: id
    cursor: updated_at
    fields: available(integer), category_id(integer), checked_out_to(array), created_at(string), currency(string), custom_fields(array), department_id(integer), id(integer), location_id(integer), logo(string), manufacturer_id(integer), minimum_quantity(integer), model_number(string), name(string), notes(string), order_number(string), purchase_cost(number), purchase_date(string), quantity(integer), supplier_id(integer), unit_cost(number), updated_at(string)
  consumables:
    primary key: id
    cursor: updated_at
    fields: available(integer), category_id(integer), checked_out_to_employees(array), checked_out_to_locations(array), created_at(string), currency(string), custom_fields(array), department_id(integer), id(integer), location_id(integer), logo(string), manufacturer_id(integer), minimum_quantity(integer), model_number(string), name(string), notes(string), order_number(string), purchase_cost(number), purchase_date(string), supplier_id(integer), unit_cost(number), updated_at(string)
  categories:
    primary key: id
    cursor: updated_at
    fields: accessories(array), assets(array), components(array), consumables(array), created_at(string), eula(string), id(integer), licenses(array), logo(string), minimum_quantity(integer), name(string), number_of_accessories(integer), number_of_assets(integer), number_of_components(integer), number_of_consumables(integer), number_of_deployable_assets(integer), number_of_licenses(integer), number_of_products(integer), products(array), skip_checkout_emails(boolean), type(string), updated_at(string)
  departments:
    primary key: id
    cursor: updated_at
    fields: accessories(array), assets(array), components(array), consumables(array), created_at(string), email(string), employees(array), id(integer), licenses(array), name(string), number_of_accessories(integer), number_of_assets(integer), number_of_components(integer), number_of_consumables(integer), number_of_employees(integer), number_of_licenses(integer), phone(string), updated_at(string)
  depreciations:
    primary key: id
    cursor: updated_at
    fields: assets(array), created_at(string), id(integer), licenses(array), minimum_value(number), months(integer), name(string), number_of_assets(integer), number_of_licenses(integer), number_of_products(integer), products(array), updated_at(string)
  locations:
    primary key: id
    cursor: updated_at
    fields: accessories(array), address_line_1(string), address_line_2(string), assets(array), checked_out_assets(array), city(string), components(array), consumables(array), country(string), created_at(string), currency(string), custom_fields(array), email(string), employees(array), id(integer), licenses(array), logo(string), name(string), number_of_accessories(integer), number_of_assets(integer), number_of_checked_out_assets(integer), number_of_components(integer), number_of_consumables(integer), number_of_employees(integer), number_of_licenses(integer), phone(string), state(string), updated_at(string), zip(string)
  manufacturers:
    primary key: id
    cursor: updated_at
    fields: accessories(array), assets(array), components(array), consumables(array), created_at(string), id(integer), licenses(array), logo(string), name(string), notes(string), number_of_accessories(integer), number_of_assets(integer), number_of_components(integer), number_of_consumables(integer), number_of_licenses(integer), number_of_products(integer), products(array), support_email(string), support_phone(string), support_url(string), updated_at(string), url(string)
  products:
    primary key: id
    cursor: updated_at
    fields: archived(boolean), assets(array), category_id(integer), created_at(string), custom_fields(array), default_purchase_cost(number), depreciation_id(integer), end_of_life_date(string), end_of_life_months(integer), end_of_life_type(string), id(integer), logo(string), manufacturer_id(integer), minimum_quantity(integer), model_number(string), name(string), notes(string), number_of_assets(integer), number_of_deployable_assets(integer), updated_at(string)
  statuses:
    primary key: id
    cursor: updated_at
    fields: assets(array), created_at(string), id(integer), name(string), notes(string), number_of_assets(integer), show_in_nav(boolean), type(string), updated_at(string)
  suppliers:
    primary key: id
    cursor: updated_at
    fields: accessories(array), address_line_1(string), address_line_2(string), assets(array), city(string), components(array), consumables(array), contact_name(string), country(string), created_at(string), email(string), fax(string), id(integer), licenses(array), logo(string), maintenances(array), name(string), notes(string), number_of_accessories(integer), number_of_assets(integer), number_of_components(integer), number_of_consumables(integer), number_of_licenses(integer), number_of_maintenances(integer), phone(string), state(string), updated_at(string), url(string), zip(string)
  audits:
    primary key: id
    cursor: updated_at
    fields: asset_id(integer), audit_date(string), audit_failed_reason(string), audit_status(string), completed(boolean), created_at(string), id(integer), location_id(integer), next_audit_date(string), notes(string), scheduled(boolean), updated_at(string), user_id(integer)
  activity:
    primary key: timestamp, item_id, event
    cursor: timestamp
    fields: checked_out_to_from_email(string), checked_out_to_from_id(integer), checked_out_to_from_name(string), checked_out_to_from_type(string), event(string), item_id(integer), item_name(string), notes(string), timestamp(string), type(string), user_email(string), user_id(integer), user_name(string)
  tenants:
    primary key: tenant_id
    fields: tenant_id(integer), tenant_name(string)

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
