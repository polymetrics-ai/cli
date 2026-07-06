# Overview

Wasabi Stats API is a Tier-2 migration using an `AuthHook` for legacy's content-based auth branch
and a `RecordHook` for legacy's missing-`id` fallback. The legacy Go connector remains the data
authority for these stream names: it reads `GET v1/stats` for `bucket_stats` and `GET v1/accounts`
for `account_stats`, extracts records from a top-level `data` array, and emits each raw record
verbatim after adding an `id` only when the source record lacks one.

## Auth setup

Provide one secret, `api_key`. The hook inspects the resolved value exactly like legacy: a value
that splits on the first `:` into two parts uses HTTP Basic auth; any value without `:` is sent as
`Authorization: Bearer <api_key>`.

## Streams notes

Both streams are unpaginated and pass `start_date` only when `config.start_date` is set. `Check`
matches legacy's separate behavior by targeting `v1/stats` and declaring a `start_date` query
default.

Both streams use `projection: "passthrough"` because legacy emits the decoded record map directly,
not a field-built `connectors.Record`. The schemas intentionally list only the legacy catalog fields:
`bucket_stats` has `id`, `bucket`, `date`, and `storage_bytes`; `account_stats` has `id`, `date`,
`storage_bytes`, and `object_count`. Runtime passthrough still preserves any additional raw fields
legacy would have emitted.

The `RecordHook` preserves legacy's missing-id behavior: if `id` is absent, it derives one from the
first non-empty value of `bucket`, then `date`, then the stream name. `x-cursor-field: date` is kept
because legacy publishes `CursorFields: []string{"date"}`; no `incremental` block is declared because
legacy's `start_date` is a fixed config filter, not persisted cursor state.

## API surface

Wasabi's newer documented standalone utilization endpoints under `/v1/standalone/utilizations`
return different envelopes and PascalCase field names. They are documented in `api_surface.json` as
not wired into these legacy stream names because substituting them would change emitted records.

## Known limits

The newer standalone utilization endpoints are intentionally not substituted for the legacy
endpoints in this bundle. They may be useful in a future additive connector version or differently
named streams, but they are not byte-for-byte compatible with legacy `bucket_stats` and
`account_stats` records.

## Write actions & risks

None. Legacy returns `ErrUnsupportedOperation`, and the stats surface is read-only.
