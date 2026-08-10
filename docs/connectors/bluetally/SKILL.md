---
name: pm-bluetally
description: BlueTally connector knowledge and safe action guide.
---

# pm-bluetally

## Purpose

Reads BlueTally IT asset management data (assets, employees, licenses, maintenances, accessories) through the BlueTally REST API.

## Icon

- id: pm-sample
- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- api_key (secret)

## ETL Streams

- assets:
  - primary key: id
  - cursor: updated_at
  - fields: asset_id(string), asset_name(string), asset_serial(string), category_id(integer), category_name(string), created_at(string), currency(string), department_id(integer), id(integer), location_id(integer), notes(string), product_id(integer), product_name(string), purchase_cost(number), purchase_date(string), status_id(integer), supplier_id(integer), updated_at(string), warranty_expiration_date(string)
- employees:
  - primary key: id
  - cursor: updated_at
  - fields: archived(boolean), created_at(string), department_id(integer), email(string), id(integer), location_id(integer), manager_id(integer), name(string), notes(string), number_of_accessories(integer), number_of_assets(integer), number_of_consumables(integer), number_of_licenses(integer), title(string), updated_at(string)
- licenses:
  - primary key: id
  - cursor: updated_at
  - fields: available(integer), category_id(integer), created_at(string), currency(string), department_id(integer), expiration_date(string), id(integer), license_type(string), licensed_to_email(string), licensed_to_name(string), location_id(integer), manufacturer_id(integer), minimum_seats(integer), name(string), notes(string), number_of_seats(integer), order_number(string), purchase_cost(number), purchase_date(string), supplier_id(integer), termination_date(string), unit_cost(number), updated_at(string)
- maintenances:
  - primary key: id
  - cursor: updated_at
  - fields: asset_id(integer), cost(number), created_at(string), end_date(string), id(integer), name(string), notes(string), start_date(string), supplier_id(integer), type(string), updated_at(string)
- accessories:
  - primary key: id
  - cursor: updated_at
  - fields: available(integer), category_id(integer), created_at(string), currency(string), department_id(integer), id(integer), location_id(integer), manufacturer_id(integer), model_number(string), name(string), notes(string), purchase_cost(number), purchase_date(string), quantity(integer), supplier_id(integer), updated_at(string)
- components:
  - primary key: id
  - cursor: updated_at
  - fields: available(integer), category_id(integer), checked_out_to(array), created_at(string), currency(string), custom_fields(array), department_id(integer), id(integer), location_id(integer), logo(string), manufacturer_id(integer), minimum_quantity(integer), model_number(string), name(string), notes(string), order_number(string), purchase_cost(number), purchase_date(string), quantity(integer), supplier_id(integer), unit_cost(number), updated_at(string)
- consumables:
  - primary key: id
  - cursor: updated_at
  - fields: available(integer), category_id(integer), checked_out_to_employees(array), checked_out_to_locations(array), created_at(string), currency(string), custom_fields(array), department_id(integer), id(integer), location_id(integer), logo(string), manufacturer_id(integer), minimum_quantity(integer), model_number(string), name(string), notes(string), order_number(string), purchase_cost(number), purchase_date(string), supplier_id(integer), unit_cost(number), updated_at(string)
- categories:
  - primary key: id
  - cursor: updated_at
  - fields: accessories(array), assets(array), components(array), consumables(array), created_at(string), eula(string), id(integer), licenses(array), logo(string), minimum_quantity(integer), name(string), number_of_accessories(integer), number_of_assets(integer), number_of_components(integer), number_of_consumables(integer), number_of_deployable_assets(integer), number_of_licenses(integer), number_of_products(integer), products(array), skip_checkout_emails(boolean), type(string), updated_at(string)
- departments:
  - primary key: id
  - cursor: updated_at
  - fields: accessories(array), assets(array), components(array), consumables(array), created_at(string), email(string), employees(array), id(integer), licenses(array), name(string), number_of_accessories(integer), number_of_assets(integer), number_of_components(integer), number_of_consumables(integer), number_of_employees(integer), number_of_licenses(integer), phone(string), updated_at(string)
- depreciations:
  - primary key: id
  - cursor: updated_at
  - fields: assets(array), created_at(string), id(integer), licenses(array), minimum_value(number), months(integer), name(string), number_of_assets(integer), number_of_licenses(integer), number_of_products(integer), products(array), updated_at(string)
- locations:
  - primary key: id
  - cursor: updated_at
  - fields: accessories(array), address_line_1(string), address_line_2(string), assets(array), checked_out_assets(array), city(string), components(array), consumables(array), country(string), created_at(string), currency(string), custom_fields(array), email(string), employees(array), id(integer), licenses(array), logo(string), name(string), number_of_accessories(integer), number_of_assets(integer), number_of_checked_out_assets(integer), number_of_components(integer), number_of_consumables(integer), number_of_employees(integer), number_of_licenses(integer), phone(string), state(string), updated_at(string), zip(string)
- manufacturers:
  - primary key: id
  - cursor: updated_at
  - fields: accessories(array), assets(array), components(array), consumables(array), created_at(string), id(integer), licenses(array), logo(string), name(string), notes(string), number_of_accessories(integer), number_of_assets(integer), number_of_components(integer), number_of_consumables(integer), number_of_licenses(integer), number_of_products(integer), products(array), support_email(string), support_phone(string), support_url(string), updated_at(string), url(string)
- products:
  - primary key: id
  - cursor: updated_at
  - fields: archived(boolean), assets(array), category_id(integer), created_at(string), custom_fields(array), default_purchase_cost(number), depreciation_id(integer), end_of_life_date(string), end_of_life_months(integer), end_of_life_type(string), id(integer), logo(string), manufacturer_id(integer), minimum_quantity(integer), model_number(string), name(string), notes(string), number_of_assets(integer), number_of_deployable_assets(integer), updated_at(string)
- statuses:
  - primary key: id
  - cursor: updated_at
  - fields: assets(array), created_at(string), id(integer), name(string), notes(string), number_of_assets(integer), show_in_nav(boolean), type(string), updated_at(string)
- suppliers:
  - primary key: id
  - cursor: updated_at
  - fields: accessories(array), address_line_1(string), address_line_2(string), assets(array), city(string), components(array), consumables(array), contact_name(string), country(string), created_at(string), email(string), fax(string), id(integer), licenses(array), logo(string), maintenances(array), name(string), notes(string), number_of_accessories(integer), number_of_assets(integer), number_of_components(integer), number_of_consumables(integer), number_of_licenses(integer), number_of_maintenances(integer), phone(string), state(string), updated_at(string), url(string), zip(string)
- audits:
  - primary key: id
  - cursor: updated_at
  - fields: asset_id(integer), audit_date(string), audit_failed_reason(string), audit_status(string), completed(boolean), created_at(string), id(integer), location_id(integer), next_audit_date(string), notes(string), scheduled(boolean), updated_at(string), user_id(integer)
- activity:
  - primary key: timestamp, item_id, event
  - cursor: timestamp
  - fields: checked_out_to_from_email(string), checked_out_to_from_id(integer), checked_out_to_from_name(string), checked_out_to_from_type(string), event(string), item_id(integer), item_name(string), notes(string), timestamp(string), type(string), user_email(string), user_id(integer), user_name(string)
- tenants:
  - primary key: tenant_id
  - fields: tenant_id(integer), tenant_name(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external BlueTally API read of IT asset management data
- approval: none; read-only API
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Run BlueTally's declared streams and reverse-ETL actions.
- Usage: pm bluetally <command> [flags]
- Read streams
- Reverse ETL writes
- Other Commands
  - accessories list - Run the accessories ETL stream [intent=etl availability=implemented stream=accessories]
  - activity list - Run the activity ETL stream [intent=etl availability=implemented stream=activity]
  - api delete accessories id - Documented DELETE /accessories/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.delete.accessories-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete assets id - Documented DELETE /assets/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.delete.assets-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete audits id - Documented DELETE /audits/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.delete.audits-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete categories id - Documented DELETE /categories/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.delete.categories-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete components id - Documented DELETE /components/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.delete.components-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete consumables id - Documented DELETE /consumables/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.delete.consumables-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete departments id - Documented DELETE /departments/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.delete.departments-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete depreciations id - Documented DELETE /depreciations/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.delete.depreciations-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete employees id - Documented DELETE /employees/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.delete.employees-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete licenses id - Documented DELETE /licenses/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.delete.licenses-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete locations id - Documented DELETE /locations/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.delete.locations-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete maintenances id - Documented DELETE /maintenances/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.delete.maintenances-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete manufacturers id - Documented DELETE /manufacturers/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.delete.manufacturers-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete products id - Documented DELETE /products/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.delete.products-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete statuses id - Documented DELETE /statuses/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.delete.statuses-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api delete suppliers id - Documented DELETE /suppliers/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.delete.suppliers-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api get accessories id - Documented GET /accessories/{id} (not implemented) [intent=direct_read availability=not_implemented operation=bluetally.get.accessories-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get assets id - Documented GET /assets/{id} (not implemented) [intent=direct_read availability=not_implemented operation=bluetally.get.assets-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get audits id - Documented GET /audits/{id} (not implemented) [intent=direct_read availability=not_implemented operation=bluetally.get.audits-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get categories id - Documented GET /categories/{id} (not implemented) [intent=direct_read availability=not_implemented operation=bluetally.get.categories-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get components id - Documented GET /components/{id} (not implemented) [intent=direct_read availability=not_implemented operation=bluetally.get.components-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get consumables id - Documented GET /consumables/{id} (not implemented) [intent=direct_read availability=not_implemented operation=bluetally.get.consumables-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get departments id - Documented GET /departments/{id} (not implemented) [intent=direct_read availability=not_implemented operation=bluetally.get.departments-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get depreciations id - Documented GET /depreciations/{id} (not implemented) [intent=direct_read availability=not_implemented operation=bluetally.get.depreciations-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get employees id - Documented GET /employees/{id} (not implemented) [intent=direct_read availability=not_implemented operation=bluetally.get.employees-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get licenses id - Documented GET /licenses/{id} (not implemented) [intent=direct_read availability=not_implemented operation=bluetally.get.licenses-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get locations id - Documented GET /locations/{id} (not implemented) [intent=direct_read availability=not_implemented operation=bluetally.get.locations-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get maintenances id - Documented GET /maintenances/{id} (not implemented) [intent=direct_read availability=not_implemented operation=bluetally.get.maintenances-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get manufacturers id - Documented GET /manufacturers/{id} (not implemented) [intent=direct_read availability=not_implemented operation=bluetally.get.manufacturers-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get products id - Documented GET /products/{id} (not implemented) [intent=direct_read availability=not_implemented operation=bluetally.get.products-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get statuses id - Documented GET /statuses/{id} (not implemented) [intent=direct_read availability=not_implemented operation=bluetally.get.statuses-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api get suppliers id - Documented GET /suppliers/{id} (not implemented) [intent=direct_read availability=not_implemented operation=bluetally.get.suppliers-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
  - api post accessories - Documented POST /accessories (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.accessories]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post assets - Documented POST /assets (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.assets]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post audits - Documented POST /audits (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.audits]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post categories - Documented POST /categories (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.categories]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post checkin accessory - Documented POST /checkin/accessory (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.checkin-accessory]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post checkin asset - Documented POST /checkin/asset (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.checkin-asset]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post checkin component - Documented POST /checkin/component (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.checkin-component]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post checkin license - Documented POST /checkin/license (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.checkin-license]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post checkout accessory - Documented POST /checkout/accessory (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.checkout-accessory]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post checkout asset - Documented POST /checkout/asset (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.checkout-asset]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post checkout component - Documented POST /checkout/component (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.checkout-component]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post checkout consumable - Documented POST /checkout/consumable (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.checkout-consumable]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post checkout license - Documented POST /checkout/license (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.checkout-license]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post components - Documented POST /components (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.components]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post consumables - Documented POST /consumables (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.consumables]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post departments - Documented POST /departments (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.departments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post depreciations - Documented POST /depreciations (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.depreciations]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post employees - Documented POST /employees (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.employees]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post licenses - Documented POST /licenses (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.licenses]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post locations - Documented POST /locations (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.locations]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post maintenances - Documented POST /maintenances (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.maintenances]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post manufacturers - Documented POST /manufacturers (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.manufacturers]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post products - Documented POST /products (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.products]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post statuses - Documented POST /statuses (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.statuses]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api post suppliers - Documented POST /suppliers (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.post.suppliers]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put accessories id - Documented PUT /accessories/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.put.accessories-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put assets id - Documented PUT /assets/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.put.assets-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put audits id - Documented PUT /audits/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.put.audits-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put categories id - Documented PUT /categories/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.put.categories-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put components id - Documented PUT /components/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.put.components-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put consumables id - Documented PUT /consumables/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.put.consumables-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put departments id - Documented PUT /departments/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.put.departments-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put depreciations id - Documented PUT /depreciations/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.put.depreciations-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put employees id - Documented PUT /employees/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.put.employees-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put licenses id - Documented PUT /licenses/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.put.licenses-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put locations id - Documented PUT /locations/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.put.locations-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put maintenances id - Documented PUT /maintenances/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.put.maintenances-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put manufacturers id - Documented PUT /manufacturers/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.put.manufacturers-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put products id - Documented PUT /products/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.put.products-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put statuses id - Documented PUT /statuses/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.put.statuses-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - api put suppliers id - Documented PUT /suppliers/{id} (not implemented) [intent=direct_write availability=not_implemented operation=bluetally.put.suppliers-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
  - assets list - Run the assets ETL stream [intent=etl availability=implemented stream=assets]
  - audits list - Run the audits ETL stream [intent=etl availability=implemented stream=audits]
  - categories list - Run the categories ETL stream [intent=etl availability=implemented stream=categories]
  - components list - Run the components ETL stream [intent=etl availability=implemented stream=components]
  - consumables list - Run the consumables ETL stream [intent=etl availability=implemented stream=consumables]
  - departments list - Run the departments ETL stream [intent=etl availability=implemented stream=departments]
  - depreciations list - Run the depreciations ETL stream [intent=etl availability=implemented stream=depreciations]
  - employees list - Run the employees ETL stream [intent=etl availability=implemented stream=employees]
  - licenses list - Run the licenses ETL stream [intent=etl availability=implemented stream=licenses]
  - locations list - Run the locations ETL stream [intent=etl availability=implemented stream=locations]
  - maintenances list - Run the maintenances ETL stream [intent=etl availability=implemented stream=maintenances]
  - manufacturers list - Run the manufacturers ETL stream [intent=etl availability=implemented stream=manufacturers]
  - products list - Run the products ETL stream [intent=etl availability=implemented stream=products]
  - statuses list - Run the statuses ETL stream [intent=etl availability=implemented stream=statuses]
  - suppliers list - Run the suppliers ETL stream [intent=etl availability=implemented stream=suppliers]
  - tenants list - Run the tenants ETL stream [intent=etl availability=implemented stream=tenants]

## Commands

### Inspect as a manual

```bash
pm connectors inspect bluetally
```

### Inspect as structured JSON

```bash
pm connectors inspect bluetally --json
```

## Agent Rules

- Run pm connectors inspect bluetally before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
