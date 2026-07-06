# Overview

Reads Freshsales (Freshworks CRM) contacts, sales accounts, deals, and leads through the Freshsales
REST API.

Readable streams: `contacts`, `sales_accounts`, `deals`, `leads`.

This connector is read-only; no write actions are declared.

Service API documentation: https://freshsales.io/api-docs.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Freshsales API key, sent as 'Authorization: Token
  token=<api_key>'. Never logged.
- `domain_name` (required, string); Freshsales account domain (e.g. mydomain.myfreshworks.com);
  combined with the fixed /crm/sales/api path to form the base URL.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `view_id` (optional, string); default `0`; Freshsales view id every stream's list endpoint is
  scoped to (<resource>/view/<view_id>). Defaults to '0', the default/all view alias.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `max_pages=0`, `view_id=0`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `Token token=` using `secrets.api_key`.

Requests use base URL `https://{{ config.domain_name }}/crm/sales/api` after applying configuration
defaults.

Connection checks call GET `/contacts/view/{{ config.view_id }}`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; no page-size parameter; starts at
1; page size 100.

- `contacts`: GET `/contacts/view/{{ config.view_id }}` - records path `contacts`; page-number
  pagination; page parameter `page`; no page-size parameter; starts at 1; page size 100.
- `sales_accounts`: GET `/sales_accounts/view/{{ config.view_id }}` - records path `sales_accounts`;
  page-number pagination; page parameter `page`; no page-size parameter; starts at 1; page size 100.
- `deals`: GET `/deals/view/{{ config.view_id }}` - records path `deals`; page-number pagination;
  page parameter `page`; no page-size parameter; starts at 1; page size 100.
- `leads`: GET `/leads/view/{{ config.view_id }}` - records path `leads`; page-number pagination;
  page parameter `page`; no page-size parameter; starts at 1; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external Freshsales API read of CRM contact, account,
deal, and lead data.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
