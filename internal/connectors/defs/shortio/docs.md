# Overview

Short.io is a wave2 fan-out declarative-HTTP migration. It reads Short.io links and domains
through the Short.io REST API (`GET https://api.short.io/api/...`). This bundle migrates
`internal/connectors/shortio` (the hand-written legacy connector) to a declarative defs bundle at
capability parity; the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Short.io API key via the `api_key` secret; it is sent verbatim as the raw `Authorization`
header value with **no** `Bearer ` prefix (`api_key_header` mode, header `Authorization`, empty
prefix — matching legacy's `connsdk.APIKeyHeader("Authorization", token, "")` at
`shortio.go:133`, Short.io's own non-standard convention) and is never logged. `base_url` defaults
to `https://api.short.io` (legacy's `shortioDefaultBaseURL`) and may be overridden for
tests/proxies.

## Streams notes

Both streams (`links`, `domains`) share Short.io's `nextPageToken` cursor pagination convention
(`pagination.type: cursor`, `cursor_param: nextPageToken`, `token_path: nextPageToken`), matching
legacy's `harvest` loop exactly (`shortio.go:90-122`): no `stop_path` is declared since legacy
stops purely on an absent/empty `nextPageToken`, with no separate boolean stop signal. `limit` is
a config-driven per-page-size override (`{{ config.page_size }}`, defaulting to legacy's own
default of 150 via `spec.json`'s `page_size` default and the query param's own `default: "150"`),
matching legacy's `pageSize` resolution (`shortio.go:213-226`).

Legacy's `shortioRecord` mapper (`shortio.go:147-149`) is shared **verbatim** by both endpoints —
both `links` and `domains` emit the identical 5-field record shape (`id`, `path`, `title`, `name`,
`updated_at`), even though `path` is a link-only concept and the domains stream's real API payload
is unlikely to ever populate it; this bundle reproduces that exact shared-mapper shape rather than
narrowing either stream's schema, matching legacy's actual emitted data field-for-field. `id` is
renamed from the raw `idString` (`computed_fields`), `title` is a pass-through rename (kept
explicit for symmetry with the `id`/`updated_at` renames), and `updated_at` is renamed from the raw
`updatedAt`. Neither stream exposes a real server-side incremental filter in legacy (no
date-range/updated-since query parameter is ever sent); `x-cursor-field: updated_at` is declared
purely as catalog/sort-key metadata matching legacy's own `CursorFields` declaration
(`shortio.go:168`), and no `incremental` block is declared on either stream, matching legacy's
full-refresh-only read behavior exactly.

## Write actions & risks

None. Short.io's legacy connector is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **Fallback field names are not modeled.** Legacy's shared `shortioRecord` mapper reads `id` with
  a fallback from `idString` to `id` (`first(item, "idString", "id")`), `title` with a fallback
  from `title` to `name` (`first(item, "title", "name")`), `name` with a fallback from `name` to
  `hostname` (`first(item, "name", "hostname")`), and `updated_at` with a fallback from `updatedAt`
  to `updated_at`. This bundle implements only the PRIMARY field of each pair
  (`idString`/`title`/`name`/`updatedAt`) — there is no coalesce/first-non-null filter in this
  dialect's `computed_fields` templating. Legacy's own test suite (`shortio_test.go`) only ever
  exercises the primary field names for every pair it asserts on (`idString`, `updatedAt`); this is
  judged an ACCEPTABLE, documented scope-narrowing rather than an `ENGINE_GAP`, per the `encharge`
  bundle's identical precedent for an unexercised defensive fallback. `name`'s own fallback to
  `hostname` is unexercised by legacy's tests at all (neither `name` nor `hostname` is asserted on);
  it is passed through by plain schema projection (no rename needed, since the schema property name
  matches the raw key already) rather than a `computed_fields` entry.
- **`max_pages` is not runtime-configurable.** Legacy exposes a `max_pages` config override
  (`0`/`all`/`unlimited` for unbounded, or a positive integer hard cap, `shortio.go:227-237`). The
  engine's `PaginationSpec.MaxPages` is a fixed bundle-authored literal, not config-driven; this
  bundle leaves it unset (unbounded), matching legacy's own default.
- **Legacy's fixture-mode-only stamped fields are not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`) stamps a `fixture: true` marker onto every emitted
  record (`shortio.go:157`); this is a credential-free conformance-harness affordance, not part of
  the live record shape, and is intentionally not reproduced — the engine's own
  `internal/connectors/conformance` fixture-replay harness provides the equivalent test affordance.
