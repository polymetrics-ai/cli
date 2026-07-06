# Overview

Reads Primetric employees, projects, clients, and roles through OAuth-authenticated REST list
endpoints.

Readable streams: `employees`, `projects`, `clients`, `roles`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developer.primetric.com/.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://api.primetric.com/api/v1`; format `uri`; Primetric
  API base URL override for tests or proxies.
- `client_id` (required, secret, string); Primetric OAuth2 client-credentials client ID. Used only
  for the token exchange; never logged.
- `client_secret` (required, secret, string); Primetric OAuth2 client-credentials client secret.
  Used only for the token exchange; never logged.
- `token_url` (optional, string); default `https://api.primetric.com/oauth/token`; format `uri`.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`.

Default configuration values: `base_url=https://api.primetric.com/api/v1`,
`token_url=https://api.primetric.com/oauth/token`.

Authentication behavior:

- OAuth 2.0 client credentials authentication using `config.token_url`, `secrets.client_id`,
  `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/employees`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; starts at 1; page size 50.

- `employees`: GET `/employees` - records path `data`; page-number pagination; page parameter
  `page`; starts at 1; page size 50; computed output fields `name`.
- `projects`: GET `/projects` - records path `data`; page-number pagination; page parameter `page`;
  starts at 1; page size 50.
- `clients`: GET `/clients` - records path `data`; page-number pagination; page parameter `page`;
  starts at 1; page size 50.
- `roles`: GET `/roles` - records path `data`; page-number pagination; page parameter `page`; starts
  at 1; page size 50.

## Write actions & risks

This connector is read-only. Read behavior: external Primetric API read of employee, project,
client, and role data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 4 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=3.
