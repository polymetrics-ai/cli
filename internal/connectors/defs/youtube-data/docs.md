# Overview

This bundle reads YouTube channels, videos, playlists, playlist items, comment threads, search
results, video categories, and i18n region/language reference data through the YouTube Data API v3
(`GET {base_url}/channels|videos|playlists|playlistItems|commentThreads|search|videoCategories|
i18nRegions|i18nLanguages`). It migrates `internal/connectors/youtube-data` (the hand-written
connector, which had only the 3 legacy-parity streams); the legacy package stays registered and
unchanged until wave6's registry flip. Pass B full-surface expansion
(`docs/migration/conventions.md`) added the 6 new streams, all reachable with this bundle's
existing api_key-only auth. Read-only: `capabilities.write` stays `false` — see Known limits for
why every YouTube Data write operation is out of reach of this connector's credential model, not
silently skipped.

## Auth setup

Provide `api_key` (secret), sent as the `key` query parameter on every request via `auth: [{"mode":
"api_key_query", "param": "key", "value": "{{ secrets.api_key }}"}]`, matching legacy's
`connsdk.APIKeyQuery("key", key)`. Never logged.

## Streams notes

All 3 streams (`channels`, `videos`, `playlists`) send `part=snippet,statistics` (matching legacy's
`q.Set("part", "snippet,statistics")` on every request, even though `playlists` items carry no
`statistics` object in practice — legacy sends the identical static `part` value for all three
streams and this bundle mirrors that exactly). Each stream additionally sends an optional `id`
query parameter: `channels` reads it from `config.channel_ids`, `videos`/`playlists` read it from
`config.ids` — both via `stream.Query`'s `omit_when_absent: true` opt-in dialect, so the parameter
is omitted entirely (not sent empty) when unset, matching legacy's `baseQuery`-equivalent
conditional (`youtube_data.go:89-97`).

Records are extracted from the top-level `items` array. `channels` computes `title` from
`record.snippet.title` and `view_count` from `record.statistics.viewCount` (a numeric-looking
string, YouTube's real wire shape — schema types it `["string","null"]`, not `integer`, to avoid a
type-widening deviation). `videos` and `playlists` both compute `title` from
`record.snippet.title` and `published_at` from `record.snippet.publishedAt`, matching legacy's
`mapVideo` function, which the legacy `playlists` stream also reuses verbatim (legacy has no
separate `mapPlaylist`).

**No `check` block is declared** (deliberate, not an oversight): legacy's `Check` performs
config/secret presence validation only (`base_url` well-formedness, `api_key` non-empty) and issues
**no HTTP request at all** (`youtube_data.go:46-60`). The engine's `Check` dispatch, when
`streams.json`'s `base.check` is unset, still resolves auth (validating `api_key` is configured)
via `newRuntime` but performs no network call and returns `nil` — this is the exact, honest
parity representation of legacy's no-network-call Check, not a narrowed capability. `conformance`'s
`check_fixture` dynamic check structurally Skips (no fixture needed) when `HTTP.Check == nil`.

**Pass B new streams**: `playlist_items`, `comment_threads`, `search`, `video_categories`,
`i18n_regions`, `i18n_languages` — all reachable with API-key-only auth (confirmed against the live
API reference; every genuinely OAuth2-gated resource/method is excluded, see `api_surface.json` and
Known limits below).

- `playlist_items` (`GET /playlistItems`) fans out over `config.playlist_ids` (comma-separated,
  `fan_out.ids_from.config_key`), sending one paginated request sequence per playlist id with
  `playlistId` as the fanned-out query parameter, stamping `playlist_id` onto every emitted record.
  Cursor-paginated (`pageToken`/`nextPageToken`, YouTube's uniform pagination convention across
  every list method). `video_id` is computed from `contentDetails.videoId`.
- `comment_threads` (`GET /commentThreads`) fans out over `config.video_ids` the identical way,
  stamping `video_id`. `text`/`published_at` are computed from the nested
  `snippet.topLevelComment.snippet.{textDisplay,publishedAt}` path (a comment thread embeds its own
  top-level comment; replies themselves are a separate, unmigrated `comments` sub-resource — see
  Known limits).
- `search` (`GET /search`) is scoped to `type=video` (a static, non-templated query value) —
  YouTube's `search` resource's own `id` field is a variant object (`id.videoId`/`id.channelId`/
  `id.playlistId` depending on match type), and the dialect's single-reference `computed_fields`
  can express exactly one of those paths per stream; restricting to videos keeps `id` (computed
  from `id.videoId`) always populated rather than null for 2 of 3 result kinds. `q`
  (`config.search_query`) and `channelId` (`config.channel_ids`, reusing the same config key the
  `channels` stream already declares) are both optional/`omit_when_absent`.
- `video_categories` (`GET /videoCategories`) requires either `id` or `regionCode`; this stream
  always sends `regionCode` (`config.region_code`, `default: "US"` materialized via the engine's
  spec-default mechanism, §3) since there is no all-regions listing.
- `i18n_regions`/`i18n_languages` are plain unpaginated reference-data lists with no id-scoping
  required at all.

## Write actions & risks

None — unchanged from legacy, but for a different reason than before. Legacy predates OAuth2
support entirely (api_key-only), so it was read-only by simple omission. Pass B research confirms
this is not a narrowing this migration introduced: **every** YouTube Data API v3 mutation
(videos/playlists/playlistItems/comments/commentThreads/channels/subscriptions/captions/
channelSections insert/update/delete, plus thumbnails.set/channelBanners.insert/watermarks.set)
requires OAuth2 user authorization against the target resource's owning channel — the API has no
API-key-authenticated write path at all. This bundle's `spec.json` declares only `api_key`; adding
OAuth2 client-credentials/authorization-code support would be a genuine credential-model expansion
(a new `spec.json` auth surface, likely `oauth2_authorization_code` — not currently a supported
`auth` mode in this engine's dialect for a 3-legged user-consent flow) rather than a Pass B
declarative addition. This is the closed-vocabulary `requires_elevated_scope` exclusion category
used throughout `api_surface.json`'s write-adjacent entries; treat as an `AUTH_COMPLEX`-class
blocker if OAuth2 write support is ever prioritized, not a gap to silently paper over.

## Known limits

- Every YouTube Data write operation requires OAuth2 (see Write actions & risks above) — this
  bundle's api_key-only auth cannot reach any of them; `capabilities.write` stays `false`.
- **`comments` (individual comment/reply list) is not migrated.** `comment_threads` already embeds
  each thread's own top-level comment (`snippet.topLevelComment`), but a thread's REPLIES live under
  the separate `comments.list` endpoint (filtered by `parentId`), which would need its own
  fan_out over comment_thread ids — a further nesting level deferred past this pass; Pass B
  breadth-vs-cost triage.
- **`activities`/`subscriptions`/`channelSections` public (channelId-scoped) list variants are not
  migrated**, even though they technically work without OAuth2 when scoped by `channelId` rather
  than `mine=true`: their public cuts are narrower/lower-value than data already reachable via the
  migrated streams (videos, playlistItems, search) and were triaged out this pass; see
  `api_surface.json` for the specific reasoning per resource.
- Full YouTube Data API surface (~19 resource types; captions, thumbnails, channel banners,
  watermarks, members, membership levels, video abuse report reasons) is out of scope for this
  pass; see `api_surface.json`'s per-endpoint `excluded` entries for the specific reason each was
  left out.
- `view_count` is typed as a nullable string (YouTube's actual wire representation for
  `statistics.viewCount`), not an integer — this matches both legacy's untyped `connectors.Record`
  passthrough and the real API shape; no widening deviation.
