# Overview

Discord is a wave2 **partial** migration of `internal/connectors/discord`
(the hand-written legacy connector this bundle migrates; the legacy package
stays registered and unchanged until wave6's registry flip, and continues to
serve the `members` stream that this bundle does not implement — see "Known
limits"). This bundle reads Discord guild, channel, and role data for a
configured guild through the Discord REST API v10 using a bot token. It is
read-only: Discord is a read-only source here, so `capabilities.write` is
`false` and this bundle ships no `writes.json`.

## Auth setup

Provide `bot_token` as a secret: a Discord bot token, sent as `Authorization:
Bot <bot_token>` — identical to legacy's `connsdk.APIKeyHeader("Authorization",
secret, "Bot ")` construction. Provide `guild_id` (required, non-secret
config): every stream is scoped to `/guilds/{guild_id}/...`.

`base_url` defaults to `https://discord.com/api/v10` (materialized via
`spec.json`'s `"default"`, matching legacy's `discordDefaultBaseURL`). The
bundle sends Discord's required `DiscordBot (https://polymetrics.ai, 2.0)
polymetrics-go-cli` User-Agent on every request via `base.user_agent`,
matching legacy exactly (Discord's Cloudflare layer may block requests with a
missing/invalid User-Agent).

## Streams notes

- **`guilds`** reads a single object (`GET /guilds/{guild_id}`) and emits it
  as one record (`records.path: "."`, `single_object: true`), matching
  legacy's `readSingle`.
- **`channels`** and **`roles`** each read an unpaginated top-level JSON
  array (`GET /guilds/{guild_id}/channels` and `/roles` respectively,
  `records.path: ""`), matching legacy's `readArray` — Discord returns the
  full channel/role list in one response with no cursor, exactly as legacy
  assumed (`pageNone`).

None of the 3 implemented streams declare a `pagination` block (the engine's
default `none` paginator: exactly one request, matching legacy's `pageNone`/
`pageSingle` behavior for these three resources).

## Write actions & risks

None. Legacy `discord` is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `metadata.json` declares
`capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- **`members` stream is BLOCKED (ENGINE_GAP), not implemented in this
  bundle.** Legacy's `readMembers` (`discord.go:198-235`) paginates Discord's
  snowflake `after`=highest-user-id cursor by reading `memberAfterID(item)`
  — the **nested** `user.id` field of the last member on each page
  (`streams.go:196-201`; Discord's real Guild Member object has no top-level
  id at all, only a nested `user` object — confirmed against
  docs.discord.com/developers/resources/guild). The engine's `cursor`
  pagination's `last_record_field` variant
  (`paginate.go`'s `lastRecordFieldValue`) reads only a **flat top-level
  key** on the raw last record (`last[field]`) with no dotted-path
  traversal — there is no way to declare `user.id` as a `last_record_field`
  value in `streams.json`. This is a genuine engine dialect gap (not a
  convenience workaround): expressing this exact cursor derivation requires
  either (a) dotted-path support added to `lastRecordFieldValue`, or (b) a
  Tier-2 `RecordHook`/`StreamHook`, both out of scope for this wave (hard
  rule: no hooks/Go this wave). `internal/connectors/discord` (legacy) keeps
  serving `members` unchanged until a follow-up wave closes this gap;
  `metadata.json`'s description and `api_surface.json`'s `members` entry
  both document this explicitly. See `docs/migration` blocker type
  `ENGINE_GAP` for this connector's reported result.
- Full Discord API surface (emojis, invites, message send/moderation,
  webhooks, threads) is out of scope for wave2; see `api_surface.json`'s
  `excluded` entries. Only the 3 legacy-parity streams implemented here are
  covered.
- **Legacy's `mode: fixture` credential-free affordance is NOT part of this
  bundle.** Legacy's `readFixture`/`fixtureMode` (`discord.go:240-278`) emit
  synthetic records without any network call when `config.mode ==
  "fixture"` — this is a legacy-only testing convenience, not part of the
  live record shape; this bundle's own `fixtures/` directory is the wave2
  substitute used by `conformance`'s dynamic (fixture-replay) checks.
- All fixtures (`fixtures/streams/**`, `fixtures/check.json`) represent
  Discord's real wire shape (a bare top-level array for `channels`/`roles`,
  a bare top-level object for `guilds`) exactly as the API returns them.
