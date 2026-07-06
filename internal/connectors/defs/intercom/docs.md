# Overview

Reads Intercom contacts, companies, conversations, admins, and tags through the Intercom REST API.

Readable streams: `contacts`, `companies`, `conversations`, `admins`, `tags`.

This connector is read-only; no write actions are declared.

Service API documentation:
https://developers.intercom.com/docs/build-an-integration/learn-more/rest-apis/unversioned-changes#unversioned-changes.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Intercom access token. Used only for Bearer auth; never
  logged.
- `api_version` (optional, string); Optional Intercom-Version header value; when unset, the header
  is omitted and Intercom uses the account's default API version.
- `base_url` (optional, string); default `https://api.intercom.io`; format `uri`; Intercom API base
  URL override for tests or proxies.
- `page_size` (optional, string); default `50`; Records per page (1-150).

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.intercom.io`, `page_size=50`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/admins`.

## Streams notes

Default pagination: cursor pagination; cursor parameter `starting_after`; next token from
`pages.next.starting_after`.

- `contacts`: GET `/contacts` - records path `data`; query `per_page`=`50`; cursor pagination;
  cursor parameter `starting_after`; next token from `pages.next.starting_after`.
- `companies`: GET `/companies` - records path `data`; query `per_page`=`50`; cursor pagination;
  cursor parameter `starting_after`; next token from `pages.next.starting_after`.
- `conversations`: GET `/conversations` - records path `data`; query `per_page`=`50`; cursor
  pagination; cursor parameter `starting_after`; next token from `pages.next.starting_after`.
- `admins`: GET `/admins` - records path `data`; cursor pagination; cursor parameter
  `starting_after`; next token from `pages.next.starting_after`.
- `tags`: GET `/tags` - records path `data`; cursor pagination; cursor parameter `starting_after`;
  next token from `pages.next.starting_after`.

## Write actions & risks

This connector is read-only. Read behavior: external Intercom API read of contact and conversation
data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, duplicate_of=1, out_of_scope=3.
