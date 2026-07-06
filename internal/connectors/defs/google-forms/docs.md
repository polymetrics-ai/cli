# Overview

Reads Google Forms metadata, form items, and submitted responses through the Google Forms REST API
using an OAuth 2.0 refresh-token grant.

Readable streams: `forms`, `form_items`, `responses`.

This connector is read-only; no write actions are declared.

Service API documentation: https://developers.google.com/forms/api/reference/rest.

## Auth setup

Connection fields:

- `base_url` (optional, string); default `https://forms.googleapis.com/v1`; format `uri`; Google
  Forms API base URL override for tests or proxies.
- `client_id` (required, secret, string); Google OAuth 2.0 client ID for the refresh-token grant.
  Used only in the token-request form; never logged.
- `client_refresh_token` (required, secret, string); Long-lived Google OAuth 2.0 refresh token.
  Exchanged for a short-lived access token at token_url; never logged. The 3-legged
  consent/acquisition dance is out of scope for this connector (credentials layer already owns it).
- `client_secret` (optional, secret, string); Google OAuth 2.0 client secret (optional for some
  client types, e.g. installed-app/native clients). Used only in the token-request form; never
  logged.
- `form_id` (required, string); Comma-, space-, or newline-separated Google Form IDs to read (forms,
  form_items, responses all fan out over this list, one request sequence per form).
- `mode` (optional, string).
- `page_size` (optional, integer); default `5000`; Records requested per page (pageSize query param)
  for the responses stream. Google Forms caps this at 5000.
- `start_date` (optional, string); format `date-time`; RFC3339 lower bound on response submission
  timestamp (responses stream's filter=timestamp>=... query). The incremental cursor
  (previously-synced last_submitted_time), when present, overrides this on a repeat sync.
- `token_url` (optional, string); default `https://oauth2.googleapis.com/token`; format `uri`;
  Google OAuth 2.0 token endpoint override. MUST be https in production; the hook fails closed on a
  non-https or unparseable value to prevent exfiltrating the refresh token to an attacker-chosen
  endpoint.

Secret fields are redacted in logs and write previews: `client_id`, `client_refresh_token`,
`client_secret`.

Default configuration values: `base_url=https://forms.googleapis.com/v1`, `page_size=5000`,
`token_url=https://oauth2.googleapis.com/token`.

Authentication behavior:

- Connector-specific authentication using `secrets.client_refresh_token`, `config.token_url`,
  `secrets.client_id`, `secrets.client_secret`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/forms/{{ config.form_id }}`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: cursor: `responses`; none: `forms`, `form_items`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `forms`: GET `/forms/{{ fanout.id }}` - records path `.`; computed output fields `description`,
  `document_title`, `form_id`, `item_count`, `responder_uri`, `revision_id`, `title`; fan-out; ids
  from config field `form_id`; id inserted into the request path.
- `form_items`: GET `/forms/{{ fanout.id }}` - records path `items`; computed output fields
  `description`, `item_id`, `question_id`, `title`; fan-out; ids from config field `form_id`; id
  inserted into the request path; stamps `form_id`.
- `responses`: GET `/forms/{{ fanout.id }}/responses` - records path `responses`; query `filter`
  from template `timestamp >= {{ incremental.lower_bound }}`, omitted when absent; `pageSize`=`{{
  config.page_size }}`; cursor pagination; cursor parameter `pageToken`; next token from
  `nextPageToken`; incremental cursor `last_submitted_time`; formatted as `rfc3339`; initial lower
  bound from `start_date`; computed output fields `answers`, `create_time`, `last_submitted_time`,
  `respondent_email`, `response_id`, `total_score`; fan-out; ids from config field `form_id`; id
  inserted into the request path; stamps `form_id`.

## Write actions & risks

This connector is read-only. Read behavior: external Google Forms API read of form metadata, form
items, and submitted responses.

## Known limits

- Batch defaults: read_page_size=5000.
- API coverage includes 3 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=1, out_of_scope=4, requires_elevated_scope=1.
