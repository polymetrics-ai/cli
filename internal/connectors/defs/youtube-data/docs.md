# Overview

This bundle reads YouTube channels, videos, and playlists through the YouTube Data API v3 (`GET
{base_url}/channels|videos|playlists`). It migrates `internal/connectors/youtube-data` (the
hand-written connector); the legacy package stays registered and unchanged until wave6's registry
flip. Read-only: `capabilities.write` is `false` and this bundle ships no `writes.json`.

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

## Write actions & risks

None. YouTube Data is modeled read-only in legacy (`capabilities.Write: false`); this bundle
matches that exactly and ships no `writes.json`.

## Known limits

- Full YouTube Data API surface (playlist items, search, comment threads, captions, live
  broadcasts) is out of scope for wave2; see `api_surface.json`'s `excluded: {category:
  out_of_scope}` entries.
- `view_count` is typed as a nullable string (YouTube's actual wire representation for
  `statistics.viewCount`), not an integer — this matches both legacy's untyped `connectors.Record`
  passthrough and the real API shape; no widening deviation.
