# Overview

Reddit is a wave2 fan-out declarative-HTTP migration. It reads recent subreddit posts and comments
through Reddit's OAuth API listing endpoints (`GET https://oauth.reddit.com/r/{subreddit}/...`).
This bundle targets capability parity with `internal/connectors/reddit` (the hand-written connector
it migrates); the legacy package stays registered and unchanged until wave6's registry flip.
Read-only: legacy's `Write` always returns `connectors.ErrUnsupportedOperation`, and this bundle
declares `capabilities.write: false` with no `writes.json` to match. OAuth token acquisition is
intentionally out of scope on both sides — the caller supplies a valid bearer token directly.

## Auth setup

Provide a Reddit OAuth access token via the `access_token` secret. It is sent as a Bearer token
(`Authorization: Bearer <access_token>`), matching legacy's `connsdk.Bearer(token)`
(`reddit.go:158`) exactly. `subreddit` (required config) scopes every request path
(`/r/{{ config.subreddit }}/...`), matching legacy's own allow-listed
`"r/"+subreddit+"/"+endpoint.path` construction (`reddit.go:104`) rather than accepting an
arbitrary request path; the engine's `InterpolatePath` rejects `..`/traversal segments in the
resolved path exactly like legacy's own `cleanSegment` validation (`reddit.go:245-250`), and
urlencodes the segment by default. `base_url` defaults to `https://oauth.reddit.com` and may be
overridden for tests/proxies.

## Streams notes

Two streams, both primary-keyed on `id`: `posts` (`GET /r/{subreddit}/new`) and `comments`
(`GET /r/{subreddit}/comments`), matching legacy's `redditEndpoints` path mapping
(`reddit.go:170-173`: `posts` hits Reddit's `new` listing, not a literal `posts` path). Records are
extracted from the `data.children` array (Reddit's Listing envelope: `{"kind":"Listing","data":
{"children":[{"kind":"t3","data":{...}}, ...],"after":...}}`), matching legacy's
`connsdk.RecordsAt(resp.Body, "data.children")` (`reddit.go:108`). Each child item is itself a
`{"kind":..., "data":{...}}` wrapper, so every field is projected via `computed_fields` reaching one
level deeper (`{{ record.data.id }}`, `{{ record.data.title }}`, etc. — the engine's
`record.<dotted.path>` reference walks nested `map[string]any` values), matching legacy's
`childData()` unwrap helper (`reddit.go:192-197`) exactly. `posts` maps `id`/`name`/`title`/
`subreddit`/`author`/`created_utc`/`permalink`; `comments` maps `id`/`name`/`body`/`subreddit`/
`author`/`created_utc`/`permalink` — both matching legacy's `postRecord`/`commentRecord`
(`reddit.go:182-190`) field-for-field via bare single-reference `computed_fields` templates, so the
engine's typed extraction preserves each field's real wire type: Reddit's `created_utc` is a JSON
number (confirmed against Reddit's own documented JSON API shape), declared `"number"` here rather
than legacy's Catalog-only `"string"` typing (legacy's generic `fields()` catalog helper types
every field as `"string"` regardless of actual wire shape — that is catalog metadata, not the
emitted record's real type, which this bundle's schema instead reflects per the typed-extraction
rule).

Every request also sends `raw_json=1` (matching legacy's own query param, `reddit.go:100`, which
prevents Reddit from HTML-entity-escaping punctuation in text fields).

Pagination is cursor-based on Reddit's own `after` token (`pagination.type: cursor`, `cursor_param:
after`, `token_path: data.after`), matching legacy's own `harvest` loop (`reddit.go:92-130`) exactly:
the engine's `tokenPathCursor` paginator stops purely on an absent/empty `data.after` value, with no
`stop_path` declared — legacy has no independent boolean stop signal either (unlike Zendesk's
`has_more`), so the two sides' termination behavior is identical (see
`docs/migration/conventions.md` §3's pagination table: a spec that never sets `stop_path` preserves
the exact prior stop-on-empty-token-only behavior).

Neither stream exposes a server-side incremental filter parameter in legacy (`Read` never sends a
date-filter query param — `harvest` only ever sends `limit`/`raw_json`/`after`), so this bundle
declares no `incremental` block for either stream, matching legacy exactly. Both schemas still
declare `x-cursor-field: created_utc`, mirroring legacy's own `Catalog` `CursorFields` declaration
(`reddit.go:177-178`) for informational/dedup-mode purposes only.

## Write actions & risks

None. Legacy's own package doc states OAuth submission/reply endpoints are intentionally out of
scope; `capabilities.write` is `false` and this bundle ships no `writes.json`, matching legacy's
`Write` returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **`body`/`title` fallback (comments) is not modeled.** Legacy's `commentRecord` maps `body` from
  `first(data, "body", "title")` (a defensive OR: use `body` if present, else fall back to `title`,
  `reddit.go:187-190`). The engine's `computed_fields` dialect has no coalesce/fallback-between-two-
  paths primitive; this bundle therefore projects only the primary field (`record.data.body`),
  matching Reddit's real, documented comment shape (`t1` comment listings always carry a `body`
  field) — the `title` branch is a defensive fallback for a shape not observed on Reddit's actual
  comment listing endpoint. This narrows legacy's defensive-only fallback behavior, never its
  accepted-input behavior for the real API's actual wire shape.
- **`page_size`/`max_pages` are not runtime-configurable.** Legacy exposes config-driven overrides
  (`redditDefaultPageSize`/`redditMaxPageSize`, `reddit.go:252-274`). The engine's `cursor` paginator
  has no `PageSize`/`MaxPages`-equivalent knob wired to a config value (the `cursor_param` request
  simply carries `limit` as a static per-stream `query` value, `streams.json`'s
  `"query": {"limit": "100", ...}`), with no per-request config-driven override mechanism. This
  bundle therefore fixes Reddit's own default (`limit=100`) and does not declare `page_size`/
  `max_pages` in `spec.json` at all (a declared-but-unwireable config key is worse than an absent
  one, per the bitly/searxng/pagerduty F6 precedent).
- **Fixture-mode-only fields are not modeled.** Legacy's `readFixture` path (only reached when
  `config.mode == "fixture"`, a credential-free conformance-harness affordance) stamps a fixed
  2-record set with a synthetic `t3` kind/subreddit/author regardless of stream (`reddit.go:132-143`);
  this is a test-only affordance, not part of the live record shape. The engine's own
  conformance/fixture-replay harness provides the credential-free test affordance this bundle needs,
  so no fixture-mode equivalent is needed here.
