# Overview

Metricool is an unquarantine migration of `internal/connectors/metricool` (`metricool.go` +
`streams.go`, the read-only legacy connector this bundle migrates; the legacy package stays
registered and unchanged until wave6's registry flip). It reads Metricool brand profiles and
per-brand Instagram, Facebook, LinkedIn, and TikTok post analytics through the Metricool REST API
(`https://app.metricool.com/api`). Metricool's analytics endpoints are not paginated and are scoped
per brand (`blogId`); legacy fans out one request per configured `blog_id` for every per-brand
stream, stamping `blogId` onto each emitted record — this is exactly the shape the engine's
`fan_out` dialect (`streams.json`'s `fan_out.ids_from.config_key`) was added to express (S4 engine
mini-wave item 2), which is why this connector was previously quarantined as an `ENGINE_GAP` and is
now buildable. Read-only: legacy's `Write` always returns `connectors.ErrUnsupportedOperation`
(`metricool.go:102-104`), and this bundle declares `capabilities.write: false` with no
`writes.json` to match.

## Auth setup

Provide one secret, `user_token` (Metricool user token), sent as the `X-Mc-Auth` header on every
request (`base.auth`'s `api_key_header` mode), matching legacy's
`connsdk.APIKeyHeader("X-Mc-Auth", secret, "")` (`metricool.go:261`) exactly. `user_id` (config,
required) is sent as the `userId` query parameter on every request, matching legacy
(`metricool.go:134,144`). `base_url` defaults to `https://app.metricool.com/api` and may be
overridden for tests/proxies.

## Streams notes

- **`brands`** (`GET /admin/simpleProfiles?userId=<user_id>`): account-wide, no fan-out (matches
  legacy's `perBlog: false`, `streams.go:41-47`). Records at the response root (`records.path:
  "."`). Primary key `["id"]`.
- **`instagram_posts`** (`GET /stats/instagram/posts`), **`facebook_posts`** (`GET
  /stats/facebook/posts`), **`linkedin_posts`** (`GET /stats/linkedin/posts`): all three share the
  identical shape — records at the response root, `dateLegacy` encoding (`start`/`end` query params,
  legacy's `YYYYMMDD` convention, `metricool.go:146-148`), and the fan-out block below. Primary key
  `["blogId", "postId"]`, cursor field `publishDate`.
- **`tiktok_posts`** (`GET /v2/analytics/posts/tiktok`): records at the `data` envelope (legacy's
  `dateV2` shape, `streams.go:70-74`), `from`/`to` query params (legacy's `YYYY-MM-DDTHH:MM:SS`
  convention, `metricool.go:149-151`), and the fan-out block below. Primary key `["blogId",
  "videoId"]`, cursor field `publishDate`.
- **Fan-out (all 4 per-brand streams)**: `fan_out.ids_from.config_key` names `blog_ids` (a
  comma-separated config list, matching legacy's `metricoolBlogIDs` parsing, `metricool.go:274-287`
  exactly — trimmed, empty entries dropped), `into.query_param` names `blogId` (matching legacy's
  `query.Set("blogId", blog)`, `metricool.go:180`), and `stamp_field` writes the current blog id onto
  every emitted record of that sub-sequence (matching legacy's `record["blogId"] = blog`,
  `metricool.go:195-197`) — one full unpaginated request per configured `blog_id`.
- **Date range passthrough — no server-side arithmetic or reformatting.** Legacy defaults to a
  60-day lookback and reformats a single `start_date`/`end_date` config pair into either
  `YYYYMMDD` (`dateLegacy` streams) or `YYYY-MM-DDTHH:MM:SS` (`tiktok_posts`) depending on the
  target endpoint (`metricool.go:292-322`). The declarative dialect has no date-arithmetic/reformat
  filter (`docs/migration/conventions.md` §3 lists `urlencode`/`unix_seconds`/`base64`/`join:<sep>`/
  `last_path_segment`/`const:<value>` only — no "add N days" or "reformat this date string"
  filter), so `start`/`end` (`instagram_posts`/`facebook_posts`/`linkedin_posts`) and `from`/`to`
  (`tiktok_posts`) are wired as direct passthrough query params off `config.start_date`/
  `config.end_date` (`omit_when_absent: true`, matching `appfigures`'s identical `start`/`end`
  passthrough pattern) — the caller must supply the value already formatted for the stream being
  read (see `spec.json`'s `start_date`/`end_date` descriptions for the exact per-stream format).
  There is no automatic 60-day-back default; leaving both unset omits the params entirely (an
  unbounded Metricool-side default, not legacy's specific 60-day window).
- Every field name on every stream's raw Metricool API response already matches its schema property
  name exactly — plain schema-mode projection copies every field by exact key match with zero
  `computed_fields` needed, preserving legacy's field-built `connectors.Record{...}` mapping
  (`streams.go:192-264`) field-for-field.

## Write actions & risks

None — Metricool is read-only. `capabilities.write: false`, no `writes.json` file, matching
legacy's `ErrUnsupportedOperation` (`metricool.go:102-104`): Metricool is a social-analytics source
with no reverse-ETL write surface.

## Known limits

- **No automatic date-range default.** As noted above, legacy defaults to a 60-day lookback
  (`metricoolDefaultLookbackDays`, `metricool.go:35,302`) when `start_date` is unset; this bundle
  has no equivalent default (the dialect performs no date arithmetic) — `start_date`/`end_date` are
  plain optional passthrough params, omitted entirely when unset. A caller relying on legacy's
  implicit 60-day window must now configure `start_date` explicitly.
- **`stamp_field`'s overwrite semantics are unconditional, unlike legacy's own stamp.** Legacy
  writes `record["blogId"] = blog` unconditionally on every per-brand record after mapping
  (`metricool.go:194-197`) — there is no existing-value check to begin with, so the engine's
  `fan_out.stamp_field` (always-overwrite, `read.go:355-357`) is a byte-for-byte behavioral match
  here, unlike bundles whose legacy stitch was conditional (see `finnworlds`'s docs.md for that
  contrast).
- **No config-driven `page_size`/`max_pages` override** — moot for this connector: Metricool's
  analytics endpoints return the full unpaginated dataset for one `blogId`/date-range combination
  per request (`pagination.type: none` on every stream, matching legacy's own single-request-per-
  blog harvest with no page-token support at all, `metricool.go:182-185`), so neither concept
  applies.
