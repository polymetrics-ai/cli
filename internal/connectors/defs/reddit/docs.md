# Overview
Complete parity with the verified Reddit OAuth API ledger (https://www.reddit.com/dev/api/): all **230 concrete documented route variants** are dispositioned exactly once. **225 are executable** through 50 ETL streams, bounded direct reads, or 122 typed reverse-ETL actions; the remaining five are named, machine-checkable Reddit constraints listed in [Known limits](#known-limits). The ledger expands Reddit's 202 unique documented method+path endpoint identities where its URI variants represent distinct datasets.

Per the captain's ruling, moderator and live-thread-owner operations are real commands, not ledger-only: a subreddit moderator or live-thread owner is a legitimate user. They function only when the connected account has the provider-granted role on the target subreddit or thread; a Reddit 403 is a runtime authorization outcome, not a reason to omit the operation.
Service API documentation: https://www.reddit.com/dev/api/.

## Auth setup

This connector supports two authentication modes, evaluated in this declared order (`streams.json`'s `base.auth`):
1. **`access_token`** (legacy, if set) -- a caller-supplied Bearer token used as-is. Reddit bearer tokens expire 1 hour after issuance and this connector does not refresh a bare `access_token`; not schedulable for anything longer-running than an hour.
2. **`refresh_token` + `client_id` (+ `client_secret`)** (recommended) -- exchanged for a fresh `access_token` via `POST https://www.reddit.com/api/v1/access_token` (`grant_type=refresh_token`) before every read/write/check, so scheduled and long-running syncs never see a stale token. Obtain `refresh_token` from a one-time `authorization_code` exchange with `duration=permanent` (https://github.com/reddit-archive/reddit/wiki/OAuth2#refreshing-the-token). This grants USER-CONTEXT access, unlike Reddit's `client_credentials` grant (deliberately not used here: Reddit's own docs state that grant mints an application-only token that never receives a refresh_token and cannot act on behalf of a user -- no moderation, messages, votes, or saves).
**Leave `access_token` unset if you configure `refresh_token`.** When both are present, `access_token` wins (declared first, for backward compatibility with existing stored connections) -- silently disabling the auto-refresh flow.
Every request sends `User-Agent: go:ai.polymetrics.cli:v1 (by /u/<reddit_username>)`, matching Reddit's required `<platform>:<app ID>:<version> (by /u/<reddit username>)` format; non-conforming User-Agents are drastically rate-limited.

### OAuth scopes

28 distinct OAuth scopes appear across the documented endpoints. Each executable command declares its required scope in its inspected command `risk` and `notes` metadata; use `pm connectors inspect reddit --json` before granting consent. Flagged here as unusually sensitive:

- **`privatemessages`** -- full DM inbox/sent content, plus a bulk account-resolution endpoint
- **`history`** -- another user's downvoted/hidden/saved items; whether these variants expose other users' data is NOT STATED in Reddit's docs and is unverified -- treat as if it could
- **`modmail`** -- private moderator conversations, including the user-facing side
- **`modnote`** -- moderators' private notes about named users
- **`account`** -- mutates preferences, blocks users
- **`vote`** -- a human-only mutation; `pm reddit vote` cannot be used in a scheduled sync or a bulk reverse-ETL plan

Moderator-privilege scopes (`modconfig`, `modflair`, `modlog`, `modmail`, `modnote`, `modothers`, `modposts`, `modself`, `modwiki`, `structuredstyles`, `livemanage`) ask the connected user to grant this connector moderator or live-thread-owner authority over their communities/threads. Only request them if the operator genuinely intends to run moderator or live-thread commands.

## Streams notes

50 streams, grouped by Reddit's own doc sections:

- **Account**: `friends_list`, `blocked_list`, `trusted_list`
- **Flair**: `flair_list`
- **Listings**: `best_posts`, `duplicate_posts`, `hot_posts`, `posts`, `rising_posts`, `top_posts`, `controversial_posts`
- **Live threads**: `live_thread_updates`, `live_thread_discussions`
- **Private messages**: `inbox_messages`, `unread_messages`, `sent_messages`
- **Moderation**: `mod_log`, `mod_reports_queue`, `mod_spam_queue`, `mod_modqueue`, `mod_unmoderated_queue`, `mod_edited_queue`
- **Search**: `search_results`
- **Subreddits**: `subreddit_moderators`, `subreddit_contributors`, `subreddit_wiki_contributors`, `subreddit_banned_users`, `subreddit_muted_users`, `subreddit_wiki_banned_users`, `my_subscribed_subreddits`, `my_contributor_subreddits`, `my_moderated_subreddits`, `my_streams_subreddits`, `subreddit_search_results`, `popular_subreddits`, `new_subreddits`, `user_search_results`, `popular_user_subreddits`, `new_user_subreddits`
- **Users**: `user_overview`, `user_submitted`, `user_comments`, `user_upvoted`, `user_downvoted`, `user_hidden`, `user_saved`, `user_gilded`
- **Wiki**: `wiki_page_discussions`, `wiki_page_revisions`

All listing streams share `posts`/`comments`'s cursor pagination (`after`/`before`, `limit` capped at Reddit's documented maximum, default page size 100). Streams whose documented endpoint uses a generic `{where}`/`{sort}`/`{location}` placeholder are expanded into one concrete stream per genuinely distinct resource (e.g. `user_submitted` vs `user_saved` are different datasets, not the same stream with a parameter) -- see `api_surface.json`'s top-level `scope` note for the full accounting.

## Write actions & risks

122 reverse-ETL write actions, all behind plan -> preview -> explicit approval -> execute. Destructive actions (delete/remove-shaped) additionally require a typed destructive confirmation, under their own canonical command name -- no synthetic shared 'delete' command hides distinct endpoints.

`emoji emoji-asset-upload-s3` and `widgets widget-image-upload-s3` acquire a Reddit-issued lease and then perform exactly one bounded HTTPS multipart form upload to the approved Amazon S3 host. The source must be a project-local regular image file and remains bound to its approved digest; Reddit OAuth credentials are never forwarded to S3. `subreddits upload-sr-img` is instead a direct, bounded multipart upload to Reddit (PNG or JPEG, maximum 500 KiB).

`pm reddit vote` is a direct, one-record, explicitly confirmed human command. Reddit's rule is: "votes must be cast by humans. That is, API clients proxying a human's action one-for-one are OK, but bots deciding how to vote on content or amplifying a human's vote are not." The command is non-batchable and therefore cannot run from a scheduled sync, a batch reverse-ETL plan, a vote-from-file workflow, or an automated write path.
Moderator-scoped and live-thread-scoped write actions function only when the connected account moderates the target subreddit / owns-or-contributes-to the target live thread; Reddit returns 403 otherwise. This is a runtime authorization outcome, not a connector defect -- see each command's `notes`/`risk` text (`cli_surface.json`) for its required scope.

Destructive (typed confirmation required): `del`, `del_msg`, `delete_sr_banner`, `delete_sr_header`, `delete_sr_icon`, `delete_sr_img`, `live_thread_delete_update`, `live_thread_leave_contributor`, `live_thread_unhide_discussion`, `me_friends_username`, `mod_conversations_conversation_id_highlight`, `mod_conversations_conversation_id_unmute`, `mod_notes`, `multi_multipath`, `multi_multipath_r_srname`, `remove`, `subreddit_emoji_emoji_name`, `unfriend`, `unhide`, `unignore_reports`, `unlock`, `unspoiler`, `widget_widget_id`.

## Known limits

- **The `comments` stream calls the undocumented `GET /r/{subreddit}/comments`**, not the documented per-article `GET [/r/subreddit]/comments/{article}`. The documented route is deliberately excluded as superseded because it returns a single post's two-listing comment tree and would require per-post fan-out; `api_surface.json` records that explicit disposition and the 100-QPM cost rationale.
- **Mixed-kind listings are projected through the `posts` schema.** Several documented endpoints (e.g. moderation queues, a user's overview/upvoted/saved history) can return a mix of Link and Comment things in the same listing; this connector projects every record through the `post` (Link) schema for simplicity, so Comment-kind items in those listings lose comment-only fields like `body`. Use the dedicated `user_comments` stream for a pure comment listing.
- **`flair_list` uses a different response envelope** (`{"users": [...]}`, not the standard `Listing` `data.children` wrapper); `records.path` is set to `users` for that one stream.
- **Two moderator-scoped S3-lease upload endpoints are compound writes**, not blocks: Reddit's lease response selects one bounded HTTPS S3 form upload. The connector accepts only Amazon S3 hosts, does not forward Reddit OAuth credentials, and does not expose a generic raw-upload escape hatch. `upload_sr_img` is a separate direct Reddit multipart endpoint, not a lease.
- **`GET /api/morechildren` is implemented as a direct read but is a hard concurrency constraint, not a code limitation**: Reddit's own docs state "you may only make one request at a time to this API endpoint" -- callers must serialize repeated calls themselves.
- Reddit enforces 100 queries per minute per OAuth client id. The connector declares and enforces a 100-RPM inter-request limiter for read and direct-read paging; on HTTP 429 also check the `X-Ratelimit-Used`, `X-Ratelimit-Remaining`, and `X-Ratelimit-Reset` response headers.
- Reddit's Data API terms require removing user content that has been deleted from Reddit and recommend purging stored Reddit content within 48 hours. This connector does not implement automatic deletion propagation or retention enforcement for rows it has already synced; honoring that obligation for any persisted data is the operator's responsibility.
- Reddit's Developer/Data API Terms require written permission/a contract for commercial use of the Data API. Whether/how that applies to this product is an open captain decision outside this connector's scope; using this connector in a monetized product without confirming that requirement is the operator's responsibility.
- Exactly five documented variants are non-executable, each with a machine-checkable `api_surface.json.excluded` reason: `GET /api/needs_captcha` (legacy captcha flow only), `POST /api/block_user` ("Only accessible to approved OAuth applications"), `POST /api/store_visits` ("Requires a subscription to reddit premium"), `GET /api/recommend/sr/{srnames}` ("DEPRECATED"), and the documented per-article comments tree (superseded by the existing explicit undocumented sub-wide comments stream). None is silently omitted.
