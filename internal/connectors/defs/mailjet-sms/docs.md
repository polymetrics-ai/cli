# Overview

Mailjet SMS is a declarative-HTTP bundle for the documented Mailjet SMS API v4 surface at
`https://api.mailjet.com/v4`. It reads outbound SMS messages, message counts, single-message
details, and export job status; it also models the documented SMS send and export-request actions.
The legacy Go package remains read-only and registered until wave6's registry flip.

## Auth setup

Provide a Mailjet SMS API Bearer token via the `token` secret; it is sent as
`Authorization: Bearer <token>` and is never logged, matching legacy's `connsdk.Bearer(secret)`.

## Streams notes

`sms` (`GET /sms`) lists outbound SMS messages from the `{"Data":[...],"Count":N}` envelope,
records at `Data`. Pagination follows Mailjet's `Limit`/`Offset` convention
(`pagination.type: offset_limit`, `limit_param: Limit`, `offset_param: Offset`, `page_size: 100`,
matching legacy's `mailjetDefaultPageSize`) — the next page's `Offset` advances by `Limit` until a
page returns fewer than `Limit` records, matching legacy's `harvest`.

The optional `start_date`/`end_date` config values are sent as `FromTS`/`ToTS` query params on
list and count requests, matching legacy's date-window filter for `sms`. Pass B also exposes
documented optional filters: `status_code` -> `StatusCode`, `recipient` -> `To`, and `sms_ids` ->
`IDs` on the list stream. Each optional filter uses `"omit_when_absent": true`, so unset filters are
left off the request rather than sent empty.

Legacy's `smsRecord` flattens the nested `Status` (`Code`/`Name`/`Description`) and `Cost`
(`Value`/`Currency`) sub-objects into flat `status_code`/`status_name`/`status_description`/
`cost_value`/`cost_currency` columns while preserving the top-level scalar fields. This bundle
reproduces the identical flattening via `computed_fields` (`"status_code": "{{ record.Status.Code
}}"`, etc.) — each is a bare single `record.<path>` reference, so the engine's typed-extraction
rule preserves the field's native JSON type (`status_code`/`cost_value` stay numeric, never
stringified), matching legacy's own untyped `map[string]any` passthrough.

`sms_count` (`GET /sms/count`) is a single-object endpoint (no `Data` array, no pagination): the
response body itself (`{"Count":N}`) is the one record, expressed as `records.path: "."`, matching
legacy's non-paginated single-GET branch.

`sms_message` (`GET /sms/{sms_ID}`) reads the documented single-message endpoint using
`config.sms_id` and records at the response `Data` array. Its schema follows Mailjet's documented
`SmsMessage` field spelling (`MessageID`), while the legacy list stream keeps its established
`MessageId` spelling.

`sms_export` (`GET /sms/export/{Job_ID}`) reads the asynchronous export job status object using
`config.export_job_id`, emitted as a single root record. The nested `Status` object is flattened to
`status_code`, `status_name`, and `status_description` for a tabular shape.

## Write actions & risks

`send_sms` (`POST /sms-send`) sends one SMS message per record. Required fields are `From`, `To`,
and `Text`; the request body is JSON and uses the documented `NewSmsMessage` field names.

`request_sms_export` (`POST /sms/export`) creates an asynchronous export job. Required fields are
`FromTS` and `ToTS` as Unix-second integers; the export result is read through the `sms_export`
stream.

Both write actions perform external Mailjet SMS mutations and require operator approval. Fixtures
use synthetic phone numbers and message/export identifiers only.

## Known limits

- **`max_pages` is not modeled.** Legacy exposes a config-driven `max_pages` hard request-count cap
  (`mailjetMaxPages`) on the `sms` stream's read loop. The engine's `offset_limit` paginator has no
  `MaxPages`-equivalent knob wired to a config value; pagination is bounded only by the short-page
  stop signal, matching Mailjet's own real termination behavior. `max_pages` is not declared in
  `spec.json`.
- **`page_size` is fixed at 100 in the declarative stream.** The legacy package accepts a runtime
  `page_size`, but this dialect's `offset_limit.page_size` is a static field. The bundle therefore
  does not declare `page_size` in `spec.json` and uses Mailjet's default-sized legacy fixture shape.
- **Export download URLs are not fetched as a stream.** The SMS reference returns an export job
  object with a `URL`; downloading and parsing the generated CSV is a binary/file transfer follow-up,
  not a documented JSON API endpoint in this SMS reference surface.
