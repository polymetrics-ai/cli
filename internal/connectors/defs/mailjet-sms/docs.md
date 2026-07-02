# Overview

Mailjet SMS is a wave2 fan-out declarative-HTTP migration. It reads outbound SMS messages and SMS
counts from the Mailjet SMS API (`GET https://api.mailjet.com/v4/...`), full refresh, read-only.
This bundle migrates `internal/connectors/mailjet-sms` (the hand-written connector it replaces at
capability parity); the legacy package stays registered and unchanged until wave6's registry flip.

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
every request of the `sms` stream, matching legacy's date-window filter
(`mailjet-sms.go:123-128`); both use the opt-in optional-query object dialect
(`"omit_when_absent": true`) so an unset date bound is left off the request entirely rather than
sent empty or hard-erroring, matching legacy's own `if from := ...; from != ""` conditional.

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

## Write actions & risks

None. The Mailjet SMS API is read-only for this connector (legacy's `Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **`max_pages` is not modeled.** Legacy exposes a config-driven `max_pages` hard request-count cap
  (`mailjetMaxPages`) on the `sms` stream's read loop. The engine's `offset_limit` paginator has no
  `MaxPages`-equivalent knob wired to a config value; pagination is bounded only by the short-page
  stop signal, matching Mailjet's own real termination behavior. `max_pages` is not declared in
  `spec.json`.
- Full Mailjet SMS API surface (sending SMS, single-message lookup by ID) is out of scope for
  wave2; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability
  expansion"}` entries. Only the 2 legacy-parity read streams are implemented.
