# Overview

Granola is a Tier-1 declarative-HTTP wave2 fan-out migration, **partial**: it reads Granola meeting
notes list metadata through the Granola public API. This bundle migrates the `notes` stream of
`internal/connectors/granola` (the hand-written connector); the legacy package stays registered and
unchanged until wave6's registry flip. The legacy connector's second stream, `detailed_notes`, is
**not** implemented in this bundle — see Known limits.

## Auth setup

Provide the Granola public API key (`grn_` prefix) as the `api_key` secret; it flows into Bearer
auth (`Authorization: Bearer <api_key>`) and is never logged. `base_url` defaults to
`https://public-api.granola.ai/v1`.

## Streams notes

`notes` (`GET /notes`) uses the base's `cursor` pagination: `cursor_param: cursor`,
`token_path: cursor` (the next page's cursor, read from the response body's `cursor` field), and
`stop_path: hasMore` — Granola's own `{"notes":[...],"hasMore":bool,"cursor":"..."}` envelope.
Legacy's stop condition is `hasMore != "true" || cursor == ""`; the engine's `stop_path` mechanism
(falsy on any value other than the literal string `"true"`) plus the paginator's own
stop-on-empty-token default reproduce this exactly. `limit` sends `config.page_size` (default `30`,
matching legacy's `granolaDefaultPageSize`/`granolaMaxPageSize`, both capped at 30 upstream).

**Incremental sync is genuinely stateful here, unlike this wave's other bundles** — legacy's
`incrementalLowerBound` reads `req.State["cursor"]` first (the persisted incremental cursor from a
prior sync), falling back to the `start_date` config value only on a fresh sync. This is expressed
with a proper `incremental` block: `cursor_field: created_at`, `request_param: created_after`,
`param_format: rfc3339`, `start_config_key: start_date`. The engine computes the exact same
lower-bound precedence (state cursor, falling back to `start_config_key`) via
`incremental.lower_bound`/`buildInitialQuery`.

`owner_name`/`owner_email` are derived via `computed_fields` from the raw nested `owner.{name,email}`
object (legacy's `ownerFields` helper), flattening it to match the schema's top-level fields.

`start_date` only accepts RFC3339 in this bundle (`spec.json`'s `format: date-time`). Legacy also
accepted a bare `YYYY-MM-DD` and normalized it to RFC3339 in code
(`time.Parse("2006-01-02", startDate)`); the engine's `rfc3339` `param_format` sends the configured
value verbatim with no normalization step, so a bare-date `start_date` would be sent un-normalized
to the API (`created_after=2026-01-15` instead of legacy's `2026-01-15T00:00:00Z`). Rather than risk
that silent divergence, this bundle narrows `start_date`'s accepted input to RFC3339 only — the
bare-`YYYY-MM-DD` convenience is dropped (documented config-surface narrowing, matching the
established pattern in other bundles' `start_date` fields, e.g. gitlab's).

## Write actions & risks

None. Granola is a read-only source in this connector (legacy `Capabilities.Write` is `false`); no
`writes.json` file is present.

## Known limits

- **`detailed_notes` is NOT implemented in this bundle** (blocked, not silently dropped). Legacy's
  `harvestDetailed` fans out from the paginated `notes` list to a per-note `GET /notes/{id}` detail
  fetch for every note, assembling the richer `detailed_notes` record (summary, transcript,
  attendees, calendar_event, folders) from each individual detail response. This is a genuine
  sub-resource-fan-out read — `conventions.md` §1 names this exact shape as a `StreamHook` (Tier 2)
  trigger, not expressible in `streams.json`'s declarative dialect (there is no mechanism to issue a
  second, per-record follow-up request keyed off a field from the first response). This wave's
  mandate is Tier-1 JSON-only (no Go, no hooks); implementing `detailed_notes` correctly requires a
  `hooks/granola/hooks.go` `StreamHook`, which is out of scope here and left for a follow-up hooks
  wave. `api_surface.json` marks `GET /notes/{id}` `excluded: {category: out_of_scope}` accordingly.
  This bundle's `notes` stream (list-view metadata only, no summary/transcript) is fully migrated
  and at parity with legacy's own `notes` stream.
- `start_date`'s bare-`YYYY-MM-DD` convenience is dropped (see Streams notes above); only RFC3339 is
  accepted.
- Full Granola API surface beyond `/notes` and `/notes/{id}` (if any) is out of scope for wave2.
