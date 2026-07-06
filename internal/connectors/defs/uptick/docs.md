# Overview

Reads Uptick field service management data through the Uptick REST API using OAuth2 password-grant
auth.

Readable streams: `tasks`, `clients`, `properties`, `invoices`, `assets`, `quotes`,
`purchaseorders`, `forms`, `users`, `teams`, `stockitems`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://support.uptickhq.com/en/articles/6728314-uptick-api-overview-and-patch-notes.

## Auth setup

Connection fields:

- `base_url` (required, string); format `uri`; Uptick tenant base URL (per-tenant instance, e.g.
  https://demo-fire.onuptick.com). Required; must be an absolute http(s) URL with a host to bound
  SSRF risk.
- `client_id` (required, secret, string); Uptick OAuth 2.0 client ID for the password grant. Used
  only in the token-request form; never logged.
- `client_secret` (required, secret, string); Uptick OAuth 2.0 client secret. Used only in the
  token-request form; never logged.
- `page_size` (optional, string); default `100`; Records per page (1-200), sent as the page_size
  query param.
- `password` (required, secret, string); Uptick account password for the OAuth 2.0 password grant.
  Used only in the token-request form; never logged.
- `start_date` (optional, string); RFC3339 lower bound for the updatedsince query param on a fresh
  (no-cursor) sync. Optional; omitted entirely when unset (full sync).
- `username` (required, string); Uptick OAuth 2.0 Resource Owner Password Credentials grant username
  (e.g. an operator email). Not a secret; sent alongside the password in the token-request form.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`, `password`.

Default configuration values: `page_size=100`.

Authentication behavior:

- Connector-specific authentication using `config.username`, `secrets.password`, `config.base_url`,
  `secrets.client_id`, `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/v2.14/clients/` with query
`fields[Client]`=`id,created,updated,ref,name,is_active,contact_name,contact_email,contact_phone_bh,address,notes`;
`page_size`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `links.next`; next URLs
stay on the configured API host.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `tasks`: GET `/api/v2.14/tasks/` - records path `data`; query
  `fields[Task]`=`id,created,updated,deleted,ref,description,is_active,name,due,status,client,property,priority`;
  `ordering`=`-updated`; `page_size`=`{{ config.page_size }}`; follows a next-page URL from the
  response body; URL path `links.next`; next URLs stay on the configured API host; incremental
  cursor `updated`; sent as `updatedsince`; formatted as `rfc3339`; initial lower bound from
  `start_date`.
- `clients`: GET `/api/v2.14/clients/` - records path `data`; query
  `fields[Client]`=`id,created,updated,ref,name,is_active,contact_name,contact_email,contact_phone_bh,address,notes`;
  `ordering`=`-updated`; `page_size`=`{{ config.page_size }}`; follows a next-page URL from the
  response body; URL path `links.next`; next URLs stay on the configured API host; incremental
  cursor `updated`; sent as `updatedsince`; formatted as `rfc3339`; initial lower bound from
  `start_date`.
- `properties`: GET `/api/v2.14/properties/` - records path `data`; query
  `fields[Property]`=`id,created,updated,name,ref,address,timezone,status,coords`;
  `ordering`=`-updated`; `page_size`=`{{ config.page_size }}`; follows a next-page URL from the
  response body; URL path `links.next`; next URLs stay on the configured API host; incremental
  cursor `updated`; sent as `updatedsince`; formatted as `rfc3339`; initial lower bound from
  `start_date`.
- `invoices`: GET `/api/v2.14/invoices/` - records path `data`; query
  `fields[Invoice]`=`id,created,updated,number,ref,description,currency,date,due_date,status,subtotal,gst,total,is_overdue,is_sent,property,task`;
  `ordering`=`-updated`; `page_size`=`{{ config.page_size }}`; follows a next-page URL from the
  response body; URL path `links.next`; next URLs stay on the configured API host; incremental
  cursor `updated`; sent as `updatedsince`; formatted as `rfc3339`; initial lower bound from
  `start_date`.
- `assets`: GET `/api/v2.14/assets/` - records path `data`; query
  `fields[Asset]`=`id,created,updated,deleted,is_active,ref,uptick_ref,label,location,status,make,model,size,barcode,serviced_date,property,type,variant`;
  `ordering`=`-updated`; `page_size`=`{{ config.page_size }}`; follows a next-page URL from the
  response body; URL path `links.next`; next URLs stay on the configured API host; incremental
  cursor `updated`; sent as `updatedsince`; formatted as `rfc3339`; initial lower bound from
  `start_date`.
- `quotes`: GET `/api/v2.14/quotes/` - records path `data`; query `page_size`=`{{ config.page_size
  }}`; follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host; emits passthrough records.
- `purchaseorders`: GET `/api/v2.14/purchaseorders/` - records path `data`; query `page_size`=`{{
  config.page_size }}`; follows a next-page URL from the response body; URL path `links.next`; next
  URLs stay on the configured API host; emits passthrough records.
- `forms`: GET `/api/v2.14/forms/` - records path `data`; query `page_size`=`{{ config.page_size
  }}`; follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host; emits passthrough records.
- `users`: GET `/api/v2.14/users/` - records path `data`; query `page_size`=`{{ config.page_size
  }}`; follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host; emits passthrough records.
- `teams`: GET `/api/v2.14/teams/` - records path `data`; query `page_size`=`{{ config.page_size
  }}`; follows a next-page URL from the response body; URL path `links.next`; next URLs stay on the
  configured API host; emits passthrough records.
- `stockitems`: GET `/api/v2.14/stockitems/` - records path `data`; query `page_size`=`{{
  config.page_size }}`; follows a next-page URL from the response body; URL path `links.next`; next
  URLs stay on the configured API host; emits passthrough records.

## Write actions & risks

This connector is read-only. Read behavior: external Uptick field service management API reads for
tasks, clients, properties, invoices, assets, quotes, purchase orders, forms, users, teams, and
stock items.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 11 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=4.
