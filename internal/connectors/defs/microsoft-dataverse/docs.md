# Overview

Reads Microsoft Dataverse accounts, contacts, leads, opportunities, and users through the Web API.

Readable streams: `accounts`, `contacts`, `leads`, `opportunities`, `systemusers`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://learn.microsoft.com/en-us/power-apps/developer/data-platform/webapi/overview.

## Auth setup

Connection fields:

- `base_url` (required, string); format `uri`; Microsoft Dataverse Web API base URL, e.g.
  https://org.crm.dynamics.com/api/data/v9.2.
- `client_id` (optional, secret, string); Azure AD application (client) ID used for the OAuth2
  client-credentials grant. Never logged.
- `client_secret` (optional, secret, string); Azure AD application client secret, sent only in the
  OAuth2 client-credentials token-exchange request; never logged.
- `login_base_url` (optional, string); default `https://login.microsoftonline.com`; format `uri`;
  Azure AD login base URL override for tests or proxies; combined with tenant_id to derive the
  OAuth2 token endpoint.
- `max_pages` (optional, string); Permissive parse: empty, "all", or "unlimited" means unbounded.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records requested per @odata.nextLink page
  (1-5000), sent as the $top query parameter on the first request of each stream.
- `scope` (required, string); OAuth2 client-credentials scope, e.g.
  https://org.crm.dynamics.com/.default (Dataverse's own resource-scope convention: the scheme+host
  of base_url, with a /.default suffix).
- `tenant_id` (optional, secret, string); Azure AD tenant ID (GUID or verified domain), used to
  derive the per-tenant token endpoint.
- `token_url` (optional, string); format `uri`; Full OAuth2 token endpoint override. When set, takes
  priority over the derived login_base_url/tenant_id endpoint.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`, `tenant_id`.

Default configuration values: `login_base_url=https://login.microsoftonline.com`, `page_size=100`.

Authentication behavior:

- OAuth 2.0 client credentials authentication using `config.token_url`, `secrets.client_id`,
  `secrets.client_secret`, `config.scope` when `{{ config.token_url }}`.
- OAuth 2.0 client credentials authentication using `config.login_base_url`, `secrets.tenant_id`,
  `secrets.client_id`, `secrets.client_secret`, `config.scope`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/accounts` with query `$top`=`1`.

## Streams notes

Default pagination: single request; no pagination.

- `accounts`: GET `/accounts` - records path `value`.
- `contacts`: GET `/contacts` - records path `value`.
- `leads`: GET `/leads` - records path `value`.
- `opportunities`: GET `/opportunities` - records path `value`.
- `systemusers`: GET `/systemusers` - records path `value`.

## Write actions & risks

This connector is read-only. Read behavior: external Microsoft Dataverse Web API read of CRM
records.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=4.
