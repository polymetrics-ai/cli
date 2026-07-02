# Overview

Delighted is a customer-experience (NPS/CSAT/CES) survey product. This bundle reads survey
responses, people, bounces, unsubscribes, and aggregate metrics through the Delighted REST API.
Read-only. This bundle migrates `internal/connectors/delighted` (the hand-written connector); the
legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Delighted API key via the `api_key` secret; it is sent as the HTTP Basic auth username
with a blank password (`auth: [{"mode": "basic", "username": "{{ secrets.api_key }}", "password":
""}]`), matching legacy's `connsdk.Basic(secret, "")`.

## Streams notes

- `survey_responses`, `people`, `bounces`, and `unsubscribes` are Delighted's `.json` list
  endpoints, each returning a top-level JSON array (`records.path: ""`). Pagination is
  `pagination.type: page_number` (`page_param: page`, `size_param: per_page`, `page_size: 100`,
  matching legacy's `delightedDefaultPageSize`/`delightedMaxPageSize`, both 100) — a page shorter
  than 100 records is the last page, exactly matching legacy's `harvest` stop rule.
- `survey_responses` is Delighted's only genuinely incremental list stream: `incremental.
  cursor_field: updated_at`, `request_param: since`, `param_format: unix_seconds`,
  `start_config_key: start_date`. The `since` request param is sent with the state cursor's value
  when a prior sync's cursor exists, else the `start_date` config value formatted to Unix seconds,
  else omitted entirely on a fresh full sync — identical to legacy's `incrementalLowerBound`,
  modulo the config key rename covered in Known limits.
- `metrics` (`GET /metrics.json`) returns a single JSON object, not a list (`records.path: ""`
  still applies — `connsdk.RecordsAt`/the engine's schema-projection path treat a top-level JSON
  object as one record); `pagination: {"type": "none"}` overrides the base page-number pagination
  for this stream only. Legacy also lets `metrics` accept the `since` filter (from the same
  state-cursor-or-config resolution as `survey_responses`), reproduced here with an `incremental`
  block whose `cursor_field` is deliberately the empty string: `metrics` has no natural per-record
  timestamp field to serve as a cursor (legacy declares `PrimaryKey: []` and no cursor fields for
  this stream either), but `request_param`/`param_format`/`start_config_key` still resolve and
  format the `since` value identically to any other incremental stream. An empty `cursor_field` is
  valid in this dialect — `connectorgen validate`'s `cursor_field_missing` rule and
  `conformance`'s equivalent only fire when `cursor_field` is non-empty.
- `people`/`bounces`/`unsubscribes` are full-refresh only, matching legacy (no cursor fields
  declared for any of the three).

## Write actions & risks

None. Delighted is a read-only source connector here (`capabilities.write: false`); legacy's
`Write` unconditionally returns `ErrUnsupportedOperation`.

## Known limits

- **Config key rename**: legacy's `since` config key is named `start_date` in this bundle's
  `spec.json` (the wire query parameter Delighted itself receives is still `since`, unchanged).
  This matches the established convention every other `unix_seconds` + `start_config_key` bundle in
  this codebase uses (stripe, chargebee, aircall all name this config key `start_date`) — and,
  concretely, conformance's dynamic checks only special-case a config property literally named
  `start_date` with a real parseable RFC3339 synthetic value (`runtimeConfigForEngine` in
  `internal/connectors/conformance/dynamic.go`); any other property name (including `since`)
  receives the generic non-parseable `"synthetic-conformance-value"` string, which
  `param_format: unix_seconds` cannot parse and would hard-fail every dynamic check for this
  stream. Renaming the config key (not the wire behavior) resolves this without weakening the real
  Unix-seconds wire contract Delighted's API actually requires.
- `start_date`'s accepted input formats are narrower than legacy's `since`: this bundle's
  `param_format: unix_seconds` (via the engine's `parseLowerBoundTime`) accepts an all-digits
  Unix-seconds string or a strict RFC3339 timestamp. Legacy's `normalizeSince` additionally accepted
  two no-timezone datetime forms (`"2006-01-02 15:04:05"` and `"2006-01-02T15:04:05"`); the engine's
  dialect has no per-connector custom time-parsing hook, so those two literal forms are no longer
  accepted config input (documented scope narrowing of ACCEPTED CONFIG INPUT SHAPES only — every
  value legacy would parse via Unix-seconds or RFC3339 behaves identically here).
- `page_size`/`max_pages` config keys legacy exposed are not declared in `spec.json`: the engine's
  `page_number` paginator reads its page size only from `streams.json`'s statically-declared
  `pagination` block (fixed at `100`, legacy's own default and max), with no mechanism to source it
  from `RuntimeConfig.Config` at read time — the same limitation documented for searxng's
  `page_size`/`max_pages` (`docs/migration/conventions.md`'s Tier-1 read-only variant section). A
  `spec.json` property no template ever consumes is dead config (F6, REVIEW.md).
- `promoter_percent`/`passive_percent`/`detractor_percent` are typed `number` in this bundle's
  `metrics` schema, matching legacy's own declared `Field.Type: "number"` catalog contract — legacy's
  in-code fixture-mode sample data inconsistently emitted these as strings (`"50.0"`); this bundle's
  fixtures use real JSON numbers instead, tightening to the catalog's own declared type rather than
  reproducing a fixture-authoring inconsistency in legacy's own test-only code path.
- Full Delighted API surface (creating/updating people, sending surveys) is out of scope for this
  wave; see `api_surface.json`'s `excluded` entries.
