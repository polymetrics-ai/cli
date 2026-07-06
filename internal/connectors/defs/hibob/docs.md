# Overview

Reads HiBob HR data: employee profiles, company named lists, and people field definitions via the
HiBob REST API (read-only).

Readable streams: `profiles`, `named_lists`, `company_lists`.

This connector is read-only; no write actions are declared.

Service API documentation: https://apidocs.hibob.com/.

## Auth setup

Connection fields:

- `base_url` (required, string); format `uri`; HiBob API base URL: https://api.hibob.com/v1
  (production) or https://api.sandbox.hibob.com/v1 (sandbox).
- `password` (required, secret, string); HiBob service-user token, sent as the HTTP Basic auth
  password. Never logged.
- `username` (required, string); HiBob service-user id, sent as the HTTP Basic auth username.

Secret fields are redacted in logs and write previews: `password`.

Authentication behavior:

- HTTP Basic authentication using `config.username`, `secrets.password`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/profiles`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: none: `named_lists`, `company_lists`; offset_limit: `profiles`.

- `profiles`: GET `/profiles` - records path `employees`; offset/limit pagination; offset parameter
  `offset`; limit parameter `limit`; page size 50; computed output fields `personal_pronouns`,
  `work_department`, `work_isManager`, `work_site`, `work_startDate`, `work_title`.
- `named_lists`: GET `/company/named-lists` - records path `values`.
- `company_lists`: GET `/company/people/fields` - records at response root.

## Write actions & risks

This connector is read-only. Read behavior: external HiBob API read of employee profile and HR
metadata.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
