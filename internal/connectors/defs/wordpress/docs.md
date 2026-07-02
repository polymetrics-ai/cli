# Overview

WordPress is a wave2 fan-out declarative-HTTP migration. It reads public WordPress REST API
resources — posts, pages, comments, media, users, categories, and tags — from any WordPress site
exposing the standard REST API (`GET {base_url}/wp-json/wp/v2/...`). This bundle is
engine-vs-legacy parity-tested against `internal/connectors/wordpress` (the hand-written connector
it migrates); the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Credentials are optional. When both `username` and `password` secrets are set, HTTP Basic auth is
applied (`Authorization: Basic ...`), matching legacy's `connsdk.Basic(username, password)`
(`wordpress.go:131-136`) — legacy applies Basic auth only when BOTH secrets are non-empty; if only
one is set, legacy sends no auth at all. This bundle's `auth` candidate gates on `secrets.password`
truthiness alone (the `when` grammar has no two-key AND), so a username-only-set, password-unset
edge case would, unlike legacy, still omit auth (password absent, so the bearer candidate's `when`
is false) — behaviorally identical to legacy in that specific case. The only divergence is a
password-set-but-username-unset configuration, which this bundle would send as Basic auth with an
empty username where legacy would send no auth at all; this is a narrow, rarely-hit edge case
(documented in Known limits) rather than a common-path deviation.

`base_url` must be provided explicitly (required). Legacy also accepts a bare `domain` config value
and derives `https://` + domain when `base_url` is unset (`wordpress.go:166-178`) — the engine's
`spec.json` `"default"` mechanism only supports a fixed literal, not a value derived from another
config key, so this bundle narrows the config surface to `base_url` only (see Known limits).

## Streams notes

All 7 streams (`posts`, `pages`, `comments`, `media`, `users`, `categories`, `tags`) are `GET` list
endpoints under `/wp-json/wp/v2/`; each response body is a bare top-level JSON array
(`records.path: "."`, matching legacy's `connsdk.Harvest(..., ".", ...)`). Pagination is 1-based
`page_number` (`page_param: page`, `size_param: per_page`, matching legacy's
`connsdk.PageNumberPaginator{PageParam: "page", SizeParam: "per_page", StartPage: 1}`), stopping on
a short page.

Legacy sends `after={{start_date}}` unconditionally on **every** stream's request whenever
`start_date` is configured (`wordpress.go:112-115`, computed once before the per-stream endpoint
lookup) — including `users`, `categories`, and `tags`, none of which have a `date` field or an
`after` filter semantics on the real WordPress REST API (WordPress silently ignores the
unrecognized param on those endpoints). This bundle reproduces that exact behavior: every stream's
`query` declares `"after": {"template": "{{ config.start_date }}", "omit_when_absent": true}`
(omitted entirely when `start_date` is unset, present verbatim when set), matching legacy's
`if after != "" { base.Set("after", after) }` gate.

Only `posts`, `pages`, `comments`, and `media` declare `x-cursor-field: date` and an `incremental`
block (`request_param: after`, `start_config_key: start_date`) — these 4 streams have a genuine
`date` field. `users`/`categories`/`tags` still send the (inert) `after` param when `start_date` is
configured, matching legacy's request shape exactly, but declare no `x-cursor-field`/`incremental`
block: legacy's own `streams()` sets `CursorFields: []string{"date"}` unconditionally for every
stream even though `users`/`categories`/`tags` never emit a `date` field at all — this bundle does
not reproduce that inconsistency in the schema layer, since `x-cursor-field` must name a property
that actually exists in the same schema (a hard `connectorgen validate` rule); it is a legacy
catalog-metadata quirk with no bearing on emitted record data.

## Write actions & risks

None. Legacy `Write` always returns `connectors.ErrUnsupportedOperation`; `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Known limits

- **`domain`-derived `base_url` is not modeled; `base_url` is required instead.** Legacy accepts
  either an explicit `base_url` or a bare `domain` (deriving `https://<domain>`) when `base_url` is
  unset. The engine's `spec.json` `"default"` materialization only supports a fixed literal value,
  not one derived from another config key (the same class of gap documented for sentry's
  hostname-derived URL and chargebee's site-derived URL in `docs/migration/conventions.md`). This
  bundle requires `base_url` explicitly; a caller previously relying on bare `domain` must now pass
  the fully-qualified URL.
- **Basic-auth "both secrets required" is approximated by a single-key `when` gate on
  `password`.** See Auth setup above: the engine's `when` grammar supports only a single
  reference's truthiness/equality/membership, not a two-key AND. A password-set/username-unset
  configuration would send Basic auth with an empty username, whereas legacy would send no auth at
  all in that specific case. Every other combination (both set, both unset, username-only-set)
  behaves identically to legacy.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes both as config-driven
  overrides (`pageSize`/`maxPages`, `wordpress.go:192-223`, including `"all"`/`"unlimited"` string
  values for unbounded `max_pages`). The engine's `page_number` paginator's `PageSize`/`MaxPages`
  fields are plain integers with no template/config-driven override mechanism, so neither can be
  wired to a `spec.json` config value; both are declared as fixed values in `streams.json`'s
  `base.pagination` block (`page_size: 100`, matching legacy's `defaultPageSize`; `max_pages: 1`,
  matching legacy's own default when `max_pages` is unset). `page_size`/`max_pages` are therefore
  not declared in `spec.json` at all (F6, REVIEW.md: a declared-but-unwireable config key is worse
  than an absent one).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached
  when `config.mode == "fixture"`) emits a fixed 2-record shape
  (`{"id","slug","date","status"}`) regardless of which stream was requested — not a real
  live-path shape for `users`/`categories`/`tags` (which have no `date`/`status` fields) or
  `comments` (which has no `slug` field). This bundle's schemas and fixtures target the live path
  only, per convention.
- **Single-page fixtures only, matching `max_pages: 1`'s real, always-enforced behavior.** Since
  `page_size`/`max_pages` cannot be config-driven (previous bullet), `max_pages: 1` is a hard,
  unconfigurable cap in this bundle exactly as it is in legacy's own unset-`max_pages` default —
  every stream genuinely only ever fetches one page in practice, so a second fixture page would
  describe a request the connector never issues. This bundle ships single-page fixtures for every
  stream (the identical precedent set by `defs/searxng`'s own `max_pages: 1` streams), rather than
  a misleading 2-page fixture that `pagination_terminates` could never actually reach.
