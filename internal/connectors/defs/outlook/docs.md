# Overview

Reads Outlook messages, mail folders, and calendar events through Microsoft Graph using an OAuth 2.0
refresh-token grant.

Readable streams: `messages`, `mail_folders`, `events`.

This connector is read-only; no write actions are declared.

Service API documentation: https://learn.microsoft.com/en-us/graph/api/resources/mail-api-overview.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://graph.microsoft.com/v1.0`; format `uri`; Microsoft
  Graph API base URL override for tests or proxies.
- `client_id` (required, secret, string); Microsoft Entra ID (Azure AD) application (client) ID for
  the OAuth 2.0 refresh-token grant. Used only in the token-request form; never logged.
- `client_secret` (required, secret, string); Microsoft Entra ID application client secret. Used
  only in the token-request form; never logged.
- `max_pages` (optional, string); default `0`; Maximum pages; use 0, all, or unlimited to exhaust
  the stream.
- `mode` (optional, string).
- `page_size` (optional, string); default `100`; Records per page ($top), 1-999.
- `refresh_token` (required, secret, string); Long-lived Microsoft Graph OAuth 2.0 refresh token.
  Exchanged for a short-lived access token at token_url; never logged. The 3-legged
  consent/acquisition dance is out of scope for this connector (credentials layer already owns it).
- `scope` (optional, string); OAuth 2.0 scope(s) requested in the token exchange (space-separated).
  Optional; omitted from the token request entirely when unset.
- `tenant_id` (optional, string); default `common`; Microsoft Entra ID tenant ID (or
  "common"/"organizations"/"consumers"), used to derive token_url when token_url is not set
  directly.
- `token_url` (optional, string); format `uri`; Microsoft identity platform token endpoint override.
  When unset, it is derived from tenant_id
  (https://login.microsoftonline.com/{tenant_id}/oauth2/v2.0/token, tenant_id defaulting to
  "common"). MUST be an https URL with a host; the hook fails closed on an invalid value to prevent
  exfiltrating the refresh token/client secret to an attacker-chosen endpoint.

Secret fields are redacted in logs and write previews: `client_id`, `client_secret`,
`refresh_token`.

Default configuration values: `base_url=https://graph.microsoft.com/v1.0`, `max_pages=0`,
`page_size=100`, `tenant_id=common`.

Authentication behavior:

- Connector-specific authentication using `secrets.refresh_token`, `config.token_url`,
  `secrets.client_id`, `secrets.client_secret`, `config.scope`.
- The bearer token is cached until 60 seconds before its declared expiry; when the token endpoint
  omits an expiry, a default of 3600 seconds is assumed.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/me/messages`.

## Streams notes

Default pagination: single request; no pagination.

- `messages`: GET `/me/messages` - records path `value`.
- `mail_folders`: GET `/me/mailFolders` - records path `value`.
- `events`: GET `/me/events` - records path `value`.

The `messages` and `events` streams advertise cursor fields (`received_date_time` and
`last_modified_date_time` respectively), but reads are always full syncs; no server-side
incremental filtering is applied.

## Write actions & risks

This connector is read-only. Read behavior: external Microsoft Graph API read of the authenticated
mailbox's messages, mail folders, and calendar events.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=3.
