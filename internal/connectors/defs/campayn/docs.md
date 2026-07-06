# Overview

Reads and writes Campayn subscriber lists, signup forms, contacts, email campaigns, and calendar
reports through the Campayn REST API.

Readable streams: `lists`, `emails`, `reports`, `forms`, `contacts`.

Write actions: `add_contact`, `update_contact`, `unsubscribe_contact`.

Service API documentation: https://github.com/nebojsac/Campayn-API.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Campayn API key. Sent as "Authorization: TRUEREST
  apikey=<api_key>"; never logged.
- `base_url` (required, string); format `uri`; Campayn account API base URL, e.g.
  https://<sub_domain>.campayn.com/api/v1. No default: the real base is account-subdomain-specific,
  and the engine's spec-default mechanism only materializes a fixed literal, not a
  derived-from-another-key URL (see docs.md's Known limits).
- `mode` (optional, string).
- `report_from` (optional, string); Optional lower-bound microtime (Unix seconds, UTC) filter for
  the reports stream's GET /reports/calendar.json?from=... query param, per the API's documented
  calendar report filter. Omitted from the request entirely when unset.
- `report_to` (optional, string); Optional upper-bound microtime (Unix seconds, UTC) filter for the
  reports stream's GET /reports/calendar.json?to=... query param. Omitted from the request entirely
  when unset.

Secret fields are redacted in logs and write previews: `api_key`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `TRUEREST apikey=` using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/lists.json`.

## Streams notes

Default pagination: single request; no pagination.

- `lists`: GET `/lists.json` - records at response root.
- `emails`: GET `/emails.json` - records at response root.
- `reports`: GET `/reports/calendar.json` - records at response root; query `from` from template `{{
  config.report_from }}`, omitted when absent; `to` from template `{{ config.report_to }}`, omitted
  when absent.
- `forms`: GET `/lists/{{ fanout.id }}/forms.json` - records at response root; fan-out; ids from
  request `/lists.json`; id field `id`; id inserted into the request path; stamps `list_id`.
- `contacts`: GET `/lists/{{ fanout.id }}/contacts.json` - records at response root; fan-out; ids
  from request `/lists.json`; id field `id`; id inserted into the request path; stamps `list_id`.

## Write actions & risks

Overall write risk: external mutation of Campayn contacts and list-subscription state (add contact,
update contact, unsubscribe by id or email); no destructive delete endpoint is documented by the
upstream API.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `add_contact`: POST `/lists/{{ record.list_id }}/contacts.json` - kind `create`; body type `json`;
  path fields `list_id`; required record fields `list_id`, `email`; accepted fields `address`,
  `city`, `company`, `country`, `custom_fields`, `email`, `failOnDuplicate`, `first_name`,
  `last_name`, `list_id`, `phones`, `sites`, `social`, `state`, `title`, `zip`; risk: adds a new
  contact to a Campayn subscriber list; low-risk external mutation, no approval required.
- `update_contact`: PUT `/contacts/{{ record.id }}.json` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `address`, `city`, `company`, `country`,
  `custom_fields`, `email`, `first_name`, `id`, `last_name`, `phones`, `sites`, `social`, `state`,
  `title`, `zip`; risk: replaces a contact's full field set (the upstream API's own docs warn any
  field not sent in the body is removed); external mutation, no approval required.
- `unsubscribe_contact`: POST `/lists/{{ record.list_id }}/unsubscribe.json` - kind `update`; body
  type `json`; path fields `list_id`; required record fields `list_id`; accepted fields `email`,
  `id`, `list_id`; risk: unsubscribes a contact from a list by id (single contact) or email (every
  contact on the list sharing that email address); the docs note neither path shows up in Reporting;
  low-risk external mutation, no approval required.

## Known limits

- API coverage includes 5 stream-backed endpoint group(s), 3 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  duplicate_of=2, requires_elevated_scope=1.
