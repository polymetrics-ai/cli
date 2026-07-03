# Overview

Thrive Learning is a wave2 fan-out declarative-HTTP migration of
`internal/connectors/thrive-learning` (the hand-written legacy connector this bundle migrates; the
legacy package stays registered and unchanged until wave6's registry flip). It reads users,
content, and learning completions through the Thrive Learning API (`GET
https://api.thrivelearning.com/{users,content,completions}`). Read-only.

## Auth setup

Provide a tenant username via the `username` config value and an API password via the `password`
secret; both are sent via HTTP Basic auth (`auth: [{"mode": "basic", ...}]`), matching legacy's
`connsdk.Basic(username, password)` (`thrive_learning.go:121`). `password` is never logged
(`x-secret: true`). This bundle sources `password` from `secrets.password` only; legacy's
`secret()` helper additionally falls back to `cfg.Config["password"]` when the secret is unset —
that fallback is not modeled here (a genuine, narrow scope reduction: a caller relying on a
config-plaintext password rather than the secret store would need to migrate to the secret,
documented in Known limits). `base_url` defaults to `https://api.thrivelearning.com` and may be
overridden for tests/proxies.

## Streams notes

Three streams, each a simple list endpoint sharing an identical shape: `users` (`GET /users`,
records at `items`), `content` (`GET /content`, records at `items`), `completions` (`GET
/completions`, records at `items`); every emitted field matches the raw API's own field names
exactly (`id`/`email`/`name`/`created_at`/`updated_at` for `users`,
`id`/`title`/`type`/`created_at`/`updated_at` for `content`,
`id`/`user_id`/`content_id`/`completed_at`/`updated_at` for `completions`). All three declare
primary key `["id"]`; legacy's own `Catalog` sets no `CursorFields` for any stream, so none of these
schemas declare `x-cursor-field`, matching exactly.

Legacy's `Read` calls `connsdk.Harvest` with a callback that does `return emit(connectors.Record(rec))`
verbatim (`thrive_learning.go:96-98`) — there is no `mapRecord`-style field-building step, so the
`fields(...)` list backing each `streamSpec` only describes legacy's advertised `Catalog` metadata,
not an actual projection filter; every raw field the API returns for a record reaches the emitted
output today. All three streams therefore declare `"projection": "passthrough"` to preserve that
exact verbatim-emit behavior; `schemas/*.json` remain a documentation surface only (the
advertised/expected shape) and are not widened to `additionalProperties: true`, matching the
pingdom/searxng precedent for this rule.

`config.start_date`, when set, is sent as the `updated_since` query parameter on **every** stream
(legacy's shared `q.Set("updated_since", start)` branch inside `Read`, applied identically
regardless of which stream is being read — `thrive_learning.go:91-94`) via the opt-in
optional-query object dialect (`"query": {"updated_since": {"template": "{{ config.start_date }}",
"omit_when_absent": true}}`, conventions.md §3): the parameter is left off the request entirely
when `start_date` is unset, exactly matching legacy's conditional branch, and sent verbatim
(legacy sends it as-is with no reformatting) when set. This is a static config-value filter, not a
true state-tracked incremental cursor — legacy never persists or reads back a sync cursor from
`req.State`, so no `incremental` block is declared on any stream (matching exactly: every sync
uses the same static `start_date`, if any, every time).

Pagination follows a 1-based page-number convention (`pagination.type: page_number`, `page_param:
page`, `size_param: limit`, `page_size: 100`) shared by all three streams — matches legacy's
`connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "limit", StartPage: 1, PageSize: size}`
(`thrive_learning.go:95-98`) exactly, including the short-page stop rule.

## Write actions & risks

None. Legacy `thrive-learning` is read-only (`Write` returns `connectors.ErrUnsupportedOperation`,
wrapped with a descriptive message); `metadata.json` declares `capabilities.write: false` and this
bundle ships no `writes.json`.

## Known limits

- Full Thrive Learning API surface (courses/pathways, groups, certificates) is out of scope; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}`
  entries — legacy itself never implemented these, so this is parity, not a reduction.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`thrive_learning.go:190-212`: `pageSize(cfg)` any positive integer defaulting to 100,
  `maxPages(cfg)` any non-negative integer defaulting to 0/unbounded, both erroring loudly on a
  non-numeric or negative value). The engine's `page_number` paginator constructor reads
  `PaginationSpec.PageSize` as a static bundle-level integer from `streams.json` and has no
  `MaxPages`-equivalent knob at all, so neither is wireable from `config.*` without inventing Go.
  This bundle hardcodes `page_size: 100` (legacy's own default) and leaves pagination unbounded
  (matching legacy's own `max_pages` default of 0/unbounded) — matching every input that does not
  explicitly override either value (the common case). Neither `page_size` nor `max_pages` is
  declared in `spec.json` (F6, REVIEW.md: a declared-but-unwireable key is worse than an absent
  one).
- **Password config-fallback is not modeled.** Legacy's `secret(cfg, "password")` helper falls back
  to a plaintext `cfg.Config["password"]` value when `cfg.Secrets["password"]` is unset
  (`thrive_learning.go:214-221`) — a narrow legacy affordance for callers that never migrated the
  password into the secret store. This bundle's `auth` spec reads only `secrets.password` (the
  dialect has no config-fallback-if-secret-absent primitive, and x-secret discipline treats a
  password as inherently secret-shaped regardless of where a caller happens to store it); a caller
  relying on the plaintext-config fallback must move the value to the `password` secret. This is a
  documented, narrow scope reduction, not a data-shape deviation — no request output changes for
  any caller already using the secret.
