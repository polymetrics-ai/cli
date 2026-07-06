# Overview

Reads Typeform forms, workspaces, themes, and images through the Typeform REST API.

Readable streams: `forms`, `responses`, `workspaces`, `themes`, `images`.

This connector is read-only; no write actions are declared.

Service API documentation: https://www.typeform.com/developers/changelog/.

## Auth setup

Connection fields:

- `access_token` (optional, secret, string); Typeform personal access token / OAuth access token.
  Used only for Bearer auth; never logged.
- `base_url` (optional, string); default `https://api.typeform.com`; format `uri`; Typeform API base
  URL override for tests or proxies.
- `form_ids` (optional, string); Comma-separated list of Typeform form ids to fan out over for the
  responses stream (one GET /forms/{form_id}/responses request per id, form_id stamped onto every
  emitted record). Required for responses.
- `mode` (optional, string).
- `page_size` (optional, string); default `200`; Records per page (1-200).

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://api.typeform.com`, `page_size=200`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/forms` with query `page_size`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `page_size`;
starts at 1; page size 200.

- `forms`: GET `/forms` - records path `items`; page-number pagination; page parameter `page`; size
  parameter `page_size`; starts at 1; page size 200; computed output fields `is_public`,
  `self_href`, `theme_href`.
- `responses`: GET `/forms/{{ fanout.id }}/responses` - records path `items`; page-number
  pagination; page parameter `page`; size parameter `page_size`; starts at 1; page size 200;
  fan-out; ids from config field `form_ids`; id inserted into the request path; stamps `form_id`.
- `workspaces`: GET `/workspaces` - records path `items`; page-number pagination; page parameter
  `page`; size parameter `page_size`; starts at 1; page size 200; computed output fields
  `self_href`.
- `themes`: GET `/themes` - records path `items`; page-number pagination; page parameter `page`;
  size parameter `page_size`; starts at 1; page size 200.
- `images`: GET `/images` - records at response root; page-number pagination; page parameter `page`;
  size parameter `page_size`; starts at 1; page size 200.

## Write actions & risks

This connector is read-only. Read behavior: external Typeform API read of form, workspace, theme,
and image metadata.

## Known limits

- Batch defaults: read_page_size=200.
- API coverage includes 5 stream-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  out_of_scope=2.
