# Overview

Reads Flowlu CRM accounts, leads, tasks, projects, invoices, and agile issues through the Flowlu
REST API.

Readable streams: `accounts`, `leads`, `tasks`, `projects`, `invoices`, `agile_issues`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.flowlu.com/api/.

## Auth setup

Connection fields:

- `api_key` (optional, secret, string); Flowlu API key, sent as the api_key query parameter on every
  request.
- `company` (required, string); Your Flowlu account subdomain (the <company> in
  https://<company>.flowlu.com). Used to derive base_url as
  https://{company}.flowlu.com/api/v1/module.

Secret fields are redacted in logs and write previews: `api_key`.

Authentication behavior:

- API key authentication in query parameter `api_key` using `secrets.api_key`.

Requests use base URL `https://{{ config.company }}.flowlu.com/api/v1/module` after applying
configuration defaults.

Connection checks call GET `crm/account/list`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `count`; starts at
1; page size 100.

- `accounts`: GET `crm/account/list` - records path `response.items`; page-number pagination; page
  parameter `page`; size parameter `count`; starts at 1; page size 100.
- `leads`: GET `crm/lead/list` - records path `response.items`; page-number pagination; page
  parameter `page`; size parameter `count`; starts at 1; page size 100.
- `tasks`: GET `task/tasks/list` - records path `response.items`; page-number pagination; page
  parameter `page`; size parameter `count`; starts at 1; page size 100.
- `projects`: GET `st/projects/list` - records path `response.items`; page-number pagination; page
  parameter `page`; size parameter `count`; starts at 1; page size 100.
- `invoices`: GET `fin/invoice/list` - records path `response.items`; page-number pagination; page
  parameter `page`; size parameter `count`; starts at 1; page size 100.
- `agile_issues`: GET `agile/issues/list` - records path `response.items`; page-number pagination;
  page parameter `page`; size parameter `count`; starts at 1; page size 100.

## Write actions & risks

This connector is read-only. Read behavior: external Flowlu CRM read of
accounts/leads/tasks/projects/invoices/agile issues.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 6 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
