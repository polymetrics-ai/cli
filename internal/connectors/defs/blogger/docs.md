# Overview

Blogger is a Tier-2 (AuthHook) migration of `internal/connectors/blogger` (legacy, read-only
reference until the wave6 registry flip). It reads Google Blogger API v3 blogs, posts, pages, and
comments via the Google OAuth 2.0 **refresh-token grant** only — the 3-legged consent/acquisition
dance is out of scope (the refresh token arrives as a pre-issued secret; the credentials layer
already owns acquisition/storage). This bundle is engine-vs-legacy parity-tested against
`internal/connectors/blogger`. Read-only: legacy `blogger.go:494-496` always returns
`ErrUnsupportedOperation` from `Write`, and this bundle declares `capabilities.write: false` with
no `writes.json` to match.

## Auth setup

Provide three secrets: `client_id`, `client_secret`, and `client_refresh_token` (long-lived; never
logged), plus the `blog_id` config value identifying which blog to read. `hooks/blogger/hooks.go`
implements `AuthHook`, porting legacy `blogger.go`'s `refreshTokenAuth` almost verbatim: it POSTs
`grant_type=refresh_token` + `refresh_token` + `client_id` + `client_secret` to `token_url`
(default `https://oauth2.googleapis.com/token`, config-overridable), caches the resulting access
token until 60 seconds before its declared expiry, and sets `Authorization: Bearer <access_token>`
on every request.

`token_url` MUST resolve to an `https://` URL: the hook fails closed on a non-https or unparseable
override rather than sending the refresh token/client secret to an attacker-chosen endpoint
(mirrors gmail's identical hook-level tightening — legacy's own `bloggerTokenURL` accepted any
scheme the caller passed with no validation at all, so this is strictly safer, never stricter for
the one real production Google token endpoint). See Known limits.

The bundle's `base.auth` declares exactly one candidate: `{"mode": "custom", "hook": "blogger",
...}` — legacy has no alternate auth path (no static API key, no public/no-auth fallback), so there
is no `when`-gated bypass to declare.

## Streams notes

Four streams, all primary-keyed on `id` with `x-cursor-field: updated` (legacy's catalog publishes
`CursorFields: []string{"updated"}` for every stream, §8 rule 2): `blogs` (single-resource,
`pagination: {"type": "none"}`, `records.path: ""` selects the bare object response exactly like
legacy's `readSingle`), and `posts`/`pages`/`comments` (all three paginated via Blogger's
`pageToken`/`nextPageToken` cursor convention — `pagination.type: cursor` with
`token_path: nextPageToken`, `cursor_param: pageToken`, `page_size: 100` sent as `maxResults`,
matching legacy's `harvest` loop exactly).

`computed_fields` flatten legacy's nested-object reads (`nestedField(item, outer, inner)`,
streams.go:201-207) into the schema's flat column names: `posts_total`/`pages_total` from
`posts.totalItems`/`pages.totalItems` on the `blogs` resource; `blog_id`/`author_id`/
`author_display_name` from `blog.id`/`author.id`/`author.displayName` on every stream;
`replies_total` from `replies.totalItems` on `posts`; `post_id` from `post.id` on `comments`. Every
one of these `totalItems` fields is Blogger's real wire type — a JSON **string** (e.g. `"2"`), not
a number — so the schema types them `["integer", "string", "null"]` rather than coercing; legacy
itself never parses these as integers either (streams.go's `bloggerBlogRecord`/`bloggerPostRecord`
copy the raw `any` value straight through).

**No incremental sync mode wired**: legacy's own doc comment (blogger.go:104-106) states Blogger
list endpoints do not accept an arbitrary updated-since filter, so `InitialState` always seeds an
empty cursor and no stream declares `incremental.request_param`/`client_filtered` — matching
legacy, every stream's `incremental` block is a bare `cursor_field` only (full_refresh +
incremental_append sync-mode eligibility, never a server-side or client-side filter).

## Write actions & risks

None — Blogger is read-only. `capabilities.write: false`, no `writes.json` file, matching legacy's
`ErrUnsupportedOperation` (`blogger.go:494-496`).

## Known limits

- **`max_pages` is not modeled**: legacy accepts a `max_pages` config value (`0`/`all`/`unlimited`
  for unbounded, else a positive integer hard-capping request count per stream,
  `bloggerMaxPages`). The engine's `PaginationSpec.MaxPages` is a fixed bundle-declared integer,
  not a runtime-config-driven override — there is no declarative mechanism to thread a per-request
  `config.max_pages` value into the pagination spec's cap. Left undeclared in `spec.json` (F6: a
  spec property with no wired template is dead config) rather than declaring an unwireable
  `max_pages` field; pagination is otherwise unbounded here (matching legacy's own default of
  unbounded when `max_pages` is unset/`all`/`unlimited`), so this narrows only the rarely-used
  explicit-cap case, not the default behavior. Deferred to Pass B if the engine grows a
  config-driven `MaxPages` override.
- **`token_url` https-only enforcement is stricter than legacy**, which performed no scheme
  validation on `token_url` at all (`bloggerTokenURL`, blogger.go:435-443, only reads a raw
  string override with no `url.Parse` check). This is documented as a parity deviation (never
  stricter for any *production* Google OAuth endpoint, which is always https; strictly safer for
  the one SSRF-adjacent secret-bearing surface this connector has). See the parity-deviation
  ledger in `docs/migration/conventions.md` §5.
- **`base_url`/`blog_id` path validation**: legacy validates `base_url`'s scheme (http/https only)
  and requires a non-empty `blog_id`, both enforced structurally here instead: `base_url` is a
  `spec.json` `format: uri` property fed straight into `streams.json`'s `base.url` template (an
  invalid override surfaces as a request-time failure rather than a dedicated pre-flight
  validation error), and `blog_id` is `required` in `spec.json` so an absent value is a hard load
  error rather than legacy's dedicated `errors.New("blogger connector requires config blog_id")`
  message. Same rejected-input set, different (still honest) error-classification path — see
  conventions.md §5's postgres precedent for the same acceptable deviation shape.
- **Bundle-level `skip_dynamic` marker**: this bundle's sole auth candidate is `mode: custom` (no
  `when`-gated non-custom fallback, mirroring gmail exactly), and conformance's synthetic
  non-secret config can never carry a real `https` `token_url` the AuthHook's own guard would
  accept — every auth-resolving dynamic check (`check_fixture`, every `read_fixture_nonempty:
  <stream>`, `pagination_terminates`, `records_match_schema`, `cursor_advances`) would otherwise
  fail identically and uninformatively. `paritytest/blogger` (which wires the real `AuthHook` via
  `engine.HooksFor("blogger")`) is the authoritative parity/correctness bar for this connector's
  auth and read paths.
