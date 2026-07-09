# Overview

Reads Front contacts, conversations, inboxes, tags, teammates, and channels through the Front Core
REST API.

Readable streams: `contacts`, `conversations`, `inboxes`, `tags`, `teammates`, `channels`.

This connector is read-only; no write actions are declared.

Service API documentation: https://dev.frontapp.com/reference/introduction.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Front API token, sent as a Bearer token (Authorization:
  Bearer <api_key>). Never logged.
- `base_url` (optional, string); default `https://api2.frontapp.com`; format `uri`; Front API base
  URL override for tests or proxies.
- `page_limit` (optional, string); default `50`; Records per page (1-100).

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api2.frontapp.com`, `page_limit=50`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/inboxes`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `_pagination.next`;
next URLs stay on the configured API host.

- `contacts`: GET `/contacts` - records path `_results`; query `limit`=`{{ config.page_limit }}`;
  follows a next-page URL from the response body; URL path `_pagination.next`; next URLs stay on the
  configured API host.
- `conversations`: GET `/conversations` - records path `_results`; query `limit`=`{{
  config.page_limit }}`; follows a next-page URL from the response body; URL path
  `_pagination.next`; next URLs stay on the configured API host.
- `inboxes`: GET `/inboxes` - records path `_results`; query `limit`=`{{ config.page_limit }}`;
  follows a next-page URL from the response body; URL path `_pagination.next`; next URLs stay on the
  configured API host.
- `tags`: GET `/tags` - records path `_results`; query `limit`=`{{ config.page_limit }}`; follows a
  next-page URL from the response body; URL path `_pagination.next`; next URLs stay on the
  configured API host.
- `teammates`: GET `/teammates` - records path `_results`; query `limit`=`{{ config.page_limit }}`;
  follows a next-page URL from the response body; URL path `_pagination.next`; next URLs stay on the
  configured API host.
- `channels`: GET `/channels` - records path `_results`; query `limit`=`{{ config.page_limit }}`;
  follows a next-page URL from the response body; URL path `_pagination.next`; next URLs stay on the
  configured API host.

## Write actions & risks

This connector is read-only. Read behavior: external Front API read of contact, conversation, inbox,
tag, teammate, and channel data.

## Known limits

- Batch defaults: read_page_size=50.
- API coverage includes 6 stream-backed endpoint group(s).
- CLI surface metadata currently marks those 6 stream-backed commands as implemented and leaves
  representative write, direct-read, binary, admin, analytics, and event surfaces planned for
  follow-up issue lanes (#190-#195).
- Other documented endpoints are not exposed by this connector where they are classified as
  non_data_endpoint=1, out_of_scope=3.
