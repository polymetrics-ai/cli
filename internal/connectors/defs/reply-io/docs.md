# Overview

Reply.io is a wave2 fan-out declarative-HTTP migration. It reads Reply.io people, campaigns,
tasks, and email accounts through the Reply.io REST API (`GET https://api.reply.io/v1/...`). This
bundle targets capability parity with `internal/connectors/reply-io` (the hand-written connector it
migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Reply.io API key via the `api_key` secret; it is sent as the `X-Api-Key` header
(`api_key_header` auth mode), matching legacy's `connsdk.APIKeyHeader("X-Api-Key", key, "")`
exactly, and is never logged. `base_url` defaults to `https://api.reply.io/v1` (legacy's
`defaultBaseURL`) and can be overridden for tests or proxies.

Note: Reply.io's own public documentation (`docs.reply.io`) describes a `Authorization: Bearer
<api_key>` header for its current (v3) API, but this bundle intentionally follows legacy's
`X-Api-Key` header instead, per this migration's ground-truth rule (legacy behavior takes priority
over documentation for accepted-input parity) — legacy targets an older/different API version and
this bundle migrates that exact behavior, not the current public docs.

## Streams notes

All four streams (`people`, `campaigns`, `tasks`, `email_accounts`) return a bare JSON array
(`records.path: ""`). Pagination is `page_number` (`page`/`limit`, static `page_size: 100` matching
legacy's `defaultPageSize`) with the identical short-page stop rule legacy's own
`connsdk.PageNumberPaginator` implements — an exact parity match, not an approximation.

Legacy applies four optional config-driven filters (`updated_after`, `created_after`, `email`,
`status`) uniformly to every stream's request. This bundle reproduces that exact behavior via the
opt-in optional-query dialect (`query.<param>.omit_when_absent: true`) declared identically on each
of the four streams' own `query` block (`base.query` has no `query` field in the engine dialect — a
declared `HTTPBase` field only supports `url`/`user_agent`/`headers`/`auth`/`pagination`/`check`/
`error_map`/`rate_limit` — so the shared filter set is duplicated per-stream rather than declared
once at the base level), so each filter is sent only when its corresponding config value is set,
and omitted entirely otherwise.

Every stream stamps a static-literal `stream` marker field (`"people"`/`"campaigns"`/`"tasks"`/
`"email_accounts"`) via `computed_fields`, matching legacy's own `out["stream"] = stream` line in
`mapRecord`. All four streams publish `updated_at` as `x-cursor-field` on their schema (matching
legacy's own descriptive `CursorFields: []string{"updated_at"}` in `Catalog()`, never wired to any
filter there either) — but Reply.io's list endpoints expose no server-side incremental filter
parameter and legacy's own `harvest` never applies one, so `streams.json` deliberately declares
**no `incremental` block at all** on any of the four streams. This matters at the catalog level, not
just at read time: the engine's `DerivedSyncModes` flips on `incremental_append`/
`incremental_append_deduped` whenever a stream's `StreamSpec.Incremental` is non-nil, independent of
whether that block actually carries a `request_param`/`client_filtered` filter — a bare
`cursor_field`-only `incremental` block is sufficient by itself to advertise incremental sync modes
this connector cannot honor (a caller selecting `incremental_append` would silently get every record
replayed on every sync). Declaring `x-cursor-field` on the schema (catalog metadata only, matching
legacy's descriptive `CursorFields`) while omitting `streams.json`'s `incremental` block entirely
(no sync-mode advertisement) is the correct, exact-parity shape — identical to this wave's sibling
`recurly` bundle, which hit the same non-incremental-legacy situation and made the same choice. Every
read is a full paginated sweep, matching legacy's true read behavior (no `request_param`/
`start_config_key`/`client_filtered` declared, and no `incremental` block to trigger sync-mode
derivation).

## Write actions & risks

None. `capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Legacy's multi-candidate records-path search is narrowed to a single fixed path.** Legacy's
  `recordsAt` helper tries a sequence of candidate body paths (the stream's own declared path,
  then `data`, `items`, `records`, `results`, then the bare root array) and uses whichever first
  yields a non-empty list — a defensive accommodation for Reply.io endpoints whose exact envelope
  shape was unconfirmed at legacy-authoring time. The engine's `records.path` names exactly one
  fixed path per stream; this bundle declares `""` (bare root array), the FIRST candidate legacy's
  own search tries (since `streamEndpoint.recordsPath` is empty for every stream in legacy, its
  candidate list begins with `""` before falling through to `data`/`items`/etc.). If Reply.io's real
  API wraps records in one of the fallback envelope keys instead of returning a bare array, this
  bundle would need updating to that confirmed shape — documented here as a scope narrowing, not
  a silent divergence, since legacy itself never confirmed which shape is authoritative.
- **Legacy's `id` fallback chain is narrowed to the raw `id` field only.** Legacy synthesizes a
  missing `id` from `first(out, "id", "uuid", "email", "name")` when the raw record has no `id`
  field at all. The `computed_fields` dialect has no OR-fallback primitive across multiple record
  paths, so this bundle relies on schema projection copying `id` directly when present; a record
  legacy would have back-filled from `uuid`/`email`/`name` would instead surface with a missing
  `id` here. Documented scope narrowing (identical narrowing class as this wave's
  retailexpress-by-maropost bundle).
- **`page_size`/`max_pages` config overrides are not modeled.** Legacy exposes `page_size` (1-100,
  default 100) and `max_pages` (0/all/unlimited or a positive integer cap) as config-driven
  overrides. The engine's `page_number` paginator has no config-driven page-size or
  request-count-cap knob (mirrors this wave's aha/referralhero/rentcast precedent); `page_size`/
  `max_pages` are therefore not declared in `spec.json`, and this bundle sends Reply.io's own
  default (`limit=100`) as a static pagination-block value.
- Full Reply.io API surface (contact lists, custom fields, blacklist rules, sequences, templates,
  webhooks) is out of scope for this wave; see `api_surface.json`'s `excluded` entries.
