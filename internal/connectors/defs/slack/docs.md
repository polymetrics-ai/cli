# Overview

Reads Slack workspace users, channels, and channel messages through the Slack Web API. Read-only.

Readable streams: `users`, `channels`, `channel_messages`.

This connector is read-only; no write actions are declared.

Service API documentation: https://api.slack.com/changelog.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Slack OAuth access token (xoxb-/xoxp-), sent as a
  Bearer Authorization header. Provide either api_token or access_token.
- `api_token` (optional, secret, string); Slack API token (xoxb- bot token or a direct API token),
  sent as a Bearer Authorization header. Provide either api_token or access_token; api_token takes
  precedence when both are set.
- `base_url` (optional, string); default `https://slack.com/api`; format `uri`; Slack Web API base
  URL override for tests or proxies.
- `channel_id` (optional, string); Slack channel ID to read history from. Required only for the
  channel_messages stream.
- `max_pages` (optional, string); Permissive parse: empty, "all", or "unlimited" means unbounded; a
  positive integer string caps the page count at that value.
- `page_size` (optional, string); default `200`; Records per page (1-1000, Slack's limit param).

Secret fields are redacted in logs and write previews: `access_token`, `api_token`.

Default configuration values: `base_url=https://slack.com/api`, `page_size=200`.

Authentication behavior:

- Bearer token authentication using `secrets.api_token` when `{{ secrets.api_token }}`.
- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.
- No authentication.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `auth.test`.

Note that Slack's Web API returns HTTP 200 even for auth failures and other errors; success is
signaled solely by the `ok` field in the JSON body, which the connector validates before trusting
a response.

## Streams notes

Default pagination: single request; no pagination.

Pagination follows Slack's `response_metadata.next_cursor` convention; each page is validated for
`ok: true` before its records and cursor are used.

- `users`: GET `users.list` - records path `members`.
- `channels`: GET `conversations.list` - records path `channels`.
- `channel_messages`: GET `conversations.history` - records path `messages`.

## Write actions & risks

This connector is read-only. Read behavior: external Slack Web API read of workspace
members/channels/channel message history.

## Known limits

- Batch defaults: read_page_size=200.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=5, requires_elevated_scope=1.
