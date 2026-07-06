# Overview

Reads and writes EZOfficeInventory assets, inventory items, stock assets, members, locations,
groups, vendors, and purchase orders through the EZOfficeInventory REST API.

Readable streams: `assets`, `inventories`, `asset_stocks`, `members`, `locations`, `groups`,
`vendors`, `purchase_orders`.

Write actions: `create_asset`, `update_asset`, `create_member`, `update_member`, `create_location`,
`update_location`, `create_group`, `update_group`, `create_vendor`, `update_vendor`,
`create_purchase_order`.

Service API documentation: https://ezofficeinventory.com/developers.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); EZOfficeInventory API access token, sent as the token
  header. Never logged.
- `mode` (optional, string).
- `subdomain` (required, string); Your EZOfficeInventory account subdomain (the <subdomain> in
  https://<subdomain>.ezofficeinventory.com). Used to derive base_url as
  https://{subdomain}.ezofficeinventory.com.

Secret fields are redacted in logs and write previews: `api_key`.

Authentication behavior:

- API key authentication in `token` using `secrets.api_key`.

Requests use base URL `https://{{ config.subdomain }}.ezofficeinventory.com` after applying
configuration defaults.

Connection checks call GET `/members.api`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `per_page`; starts
at 1; page size 25.

- `assets`: GET `/assets.api` - records path `assets`; query `include_custom_fields`=`true`;
  `show_document_details`=`true`; `show_document_urls`=`true`; `show_image_urls`=`true`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 25.
- `inventories`: GET `/inventory.api` - records path `assets`; query `include_custom_fields`=`true`;
  `show_document_details`=`true`; `show_document_urls`=`true`; `show_image_urls`=`true`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 25.
- `asset_stocks`: GET `/stock_assets.api` - records path `assets`; query
  `include_custom_fields`=`true`; `show_document_details`=`true`; `show_document_urls`=`true`;
  `show_image_urls`=`true`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 25.
- `members`: GET `/members.api` - records path `members`; page-number pagination; page parameter
  `page`; size parameter `per_page`; starts at 1; page size 25.
- `locations`: GET `/locations/get_line_item_locations.api` - records path `locations`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 25.
- `groups`: GET `/assets/classification_view.api` - records path `groups`; query
  `show_document_details`=`true`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 25; computed output fields `active`, `asset_depreciation_mode`,
  `assets_count`, `company_id`, `created_at`, `description`, `hidden_on_web_store`, `id`, `name`,
  `updated_at`.
- `vendors`: GET `/assets/vendors.api` - records path `vendors`; query
  `include_custom_fields`=`true`; page-number pagination; page parameter `page`; size parameter
  `per_page`; starts at 1; page size 25; computed output fields `assets_count`, `company_id`,
  `created_at`, `id`, `name`, `services_count`, `status`, `updated_at`.
- `purchase_orders`: GET `/purchase_orders.api` - records path `purchase_orders`; page-number
  pagination; page parameter `page`; size parameter `per_page`; starts at 1; page size 25.

## Write actions & risks

Overall write risk: external mutation of asset, member, location, group, vendor, and purchase order
records; create/update only, no delete actions implemented.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_asset`: POST `/assets.api` - kind `create`; body type `form`; required record fields
  `fixed_asset[name]`, `fixed_asset[group_id]`, `fixed_asset[location_id]`; accepted fields
  `fixed_asset[group_id]`, `fixed_asset[identifier]`, `fixed_asset[image_url]`,
  `fixed_asset[location_id]`, `fixed_asset[manufacturer]`, `fixed_asset[name]`,
  `fixed_asset[purchased_on]`, `fixed_asset[sub_group_id]`; risk: external mutation; creates a new
  asset record; approval required.
- `update_asset`: PUT `/assets/{{ record.id }}.api` - kind `update`; body type `form`; path fields
  `id`; required record fields `id`; accepted fields `fixed_asset[description]`,
  `fixed_asset[group_id]`, `fixed_asset[identifier]`, `fixed_asset[location_id]`,
  `fixed_asset[name]`, `id`; risk: external mutation; approval required.
- `create_member`: POST `/members.api` - kind `create`; body type `form`; required record fields
  `user[email]`, `user[first_name]`, `user[last_name]`, `user[role_id]`; accepted fields
  `user[email]`, `user[employee_id]`, `user[first_name]`, `user[last_name]`, `user[login_enabled]`,
  `user[phone_number]`, `user[role_id]`; risk: external mutation; creates a new member/user account;
  approval required.
- `update_member`: PUT `/members/{{ record.id }}.api` - kind `update`; body type `form`; path fields
  `id`; required record fields `id`; accepted fields `id`, `user[first_name]`, `user[last_name]`,
  `user[phone_number]`, `user[role_id]`; risk: external mutation; approval required.
- `create_location`: POST `/locations.api` - kind `create`; body type `form`; required record fields
  `location[name]`; accepted fields `location[city]`, `location[description]`, `location[name]`,
  `location[state]`, `location[street1]`, `location[street2]`, `location[zipcode]`; risk: external
  mutation; creates a new location; approval required.
- `update_location`: PUT `/locations/{{ record.id }}.api` - kind `update`; body type `form`; path
  fields `id`; required record fields `id`; accepted fields `id`, `location[city]`,
  `location[description]`, `location[name]`, `location[state]`; risk: external mutation; approval
  required.
- `create_group`: POST `/groups.api` - kind `create`; body type `form`; required record fields
  `group[name]`; accepted fields `group[asset_depreciation_mode]`, `group[description]`,
  `group[name]`; risk: external mutation; creates a new asset group/classification; approval
  required.
- `update_group`: PUT `/groups/{{ record.id }}.api` - kind `update`; body type `form`; path fields
  `id`; required record fields `id`; accepted fields `group[description]`, `group[name]`, `id`;
  risk: external mutation; approval required.
- `create_vendor`: POST `/vendors.api` - kind `create`; body type `form`; required record fields
  `vendor[name]`; accepted fields `vendor[address]`, `vendor[contact_person_name]`,
  `vendor[description]`, `vendor[email]`, `vendor[fax]`, `vendor[name]`, `vendor[phone_number]`,
  `vendor[website]`; risk: external mutation; creates a new vendor; approval required.
- `update_vendor`: PUT `/vendors/{{ record.id }}.api` - kind `update`; body type `form`; path fields
  `id`; required record fields `id`; accepted fields `id`, `vendor[email]`, `vendor[name]`,
  `vendor[phone_number]`; risk: external mutation; approval required.
- `create_purchase_order`: POST `/purchase_orders.api` - kind `create`; body type `form`; required
  record fields `vendor_id`; accepted fields `vendor_id`; risk: external mutation; creates a new
  purchase order (financial document); approval required.

## Known limits

- Batch defaults: read_page_size=25.
- API coverage includes 8 stream-backed endpoint group(s), 11 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=4, destructive_admin=6, duplicate_of=14, non_data_endpoint=2, out_of_scope=102.
