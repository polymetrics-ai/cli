# Overview

Zendesk Chat is a wave2 fan-out migration of `internal/connectors/zendesk-chat`
(the hand-written legacy connector this bundle migrates; the legacy package
stays registered and unchanged until wave6's registry flip). It reads
Zendesk Chat agents, chats, departments, shortcuts, and triggers through the
Zendesk Chat REST API v2.

## Auth setup

Provide a Zendesk Chat OAuth access token via the `access_token` secret; it
is sent as `Authorization: Bearer <access_token>`, matching legacy's
`connsdk.Bearer(secret)` (`zendeskchat.go:277`), and is never logged. If
`access_token` is unset, `streams.json`'s `base.auth` falls through to
`mode: none` (an unauthenticated request) rather than hard-erroring at
bundle-load time — the same absent-secret-falsy `when` pattern used by
searxng's optional Bearer proxy; a real read against a live Zendesk Chat
account still requires the token, since the API itself rejects
unauthenticated requests.

**Documented config-surface deviation (base_url/subdomain)**: legacy accepts
either a bare `subdomain` config value (deriving
`https://<subdomain>.zendesk.com/api/v2/chat` itself) or a `base_url`
override (`zendeskChatBaseURL`, `zendeskchat.go:324-347`). The engine's
declarative `url` template is a single static string with no conditional
branching between two config keys, so this bundle requires `base_url`
directly (matching zendesk-support's identical documented deviation) rather
than declaring an unwireable `subdomain` property (conventions.md §3).
Every legacy-accepted `subdomain`-only configuration is still reachable by
supplying the equivalent `https://<subdomain>.zendesk.com/api/v2/chat` as
`base_url`.

## Streams notes

`agents`, `departments`, `shortcuts`, and `triggers` are simple, non-paginated,
top-level-JSON-array endpoints (`records.path: "."`, matching legacy's own
`recordsPath: "."`); primary key `["id"]`. None of the four declare an
`incremental` block, matching legacy exactly (`CursorFields: nil` on all
four in `zendeskChatStreams()`).

`chats` is Zendesk's incremental-export shape
(`{chats:[...], next_url:..., count:N}`), modeled as `pagination.type:
next_url` with `next_url_path: "next_url"` — the engine's `next_url`
paginator follows the response body's absolute next-page URL exactly like
legacy's `harvestChats` loop, stopping when `next_url` is empty (matching
legacy's `next == "" || len(records) == 0` stop rule; the engine's own
next_url paginator additionally loop-guards against requesting the same URL
twice). The stream declares `incremental.cursor_field: timestamp`,
`request_param: start_time`, `param_format: unix_seconds`,
`start_config_key: start_date` — reproducing legacy's `chatsStartTime`
exactly: the resolved lower bound (state cursor if set, else `start_date`)
is converted to a Unix-seconds string and sent as `start_time` only when it
resolves (an absent cursor AND absent `start_date` means no `start_time`
param at all, i.e. a full export, matching legacy's `if startTime != ""`
guard). `param_format: unix_seconds`'s digits-only-passthrough handles a
numeric state cursor identically to legacy's own `strconv.ParseInt` fast
path.

## Write actions & risks

None. Legacy `zendesk-chat` is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `metadata.json` declares
`capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- Full Zendesk Chat API surface (bans, visitors, chat search, account
  settings, roles, goals) is out of scope for wave2; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries. Only the 5 legacy-parity streams are
  implemented.
- **`chats`' `start_time` query param is re-sent on every paginated page,
  unlike legacy** (`bitly`'s identical, already-ledgered divergence,
  conventions.md's bitly docs.md): the engine's `readDeclarative` computes
  `stream`-level query params (including the `incremental.request_param`
  value) once and merges them into every page request
  (`engine/read.go`'s `mergeQuery`), whereas legacy's `harvestChats`
  explicitly clears its outgoing query (`query = url.Values{}`) once it
  starts following an absolute `next_url`. This is benign in DATA terms
  only: Zendesk's own `next_url` already embeds the correct pagination
  state, and re-applying an identical `start_time` value is a no-op
  replace, not a behavior change an operator or the API would observe
  differently — verified by this bundle's fixture, whose (sanctioned
  single-page, see below) `next_url` is empty, so no second-page divergence
  is exercised. If a live Zendesk Chat account's `next_url` ever omitted or
  diverged from the first-page `start_time`, this bundle's request would
  differ from legacy's; today it does not.
- **`chats`' fixture is single-page** (conventions.md §4's sanctioned
  `next_url` exception): the next-page URL for a `next_url` paginator is
  the replay server's own address, unknown until the harness picks a port
  at runtime, so a static fixture cannot embed a correct second page.
  `pagination_terminates` is proven against the `agents` stream (a
  different, non-paginated stream in this same bundle) instead, per the
  sanctioned pattern; no live `paritytest` package exists for this
  connector in this wave (JSON+docs-only fan-out; a follow-up wave may add
  one alongside a hooks/native decision for any Tier-2/3 escalation).
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's
  `readFixture` path (only reached when `config.mode == "fixture"`, a
  credential-free conformance-harness affordance) stamps a
  `previous_cursor` field (echoing `req.State["cursor"]` when set) onto
  every fixture-mode record (`zendeskchat.go:253-255`), plus synthetic
  filler fields (`connector`, `fixture`) not present in any live Zendesk
  Chat response shape. None of these are part of the LIVE record shape;
  this bundle's schemas target the live path only (`harvestArray`/
  `harvestChats`), per the same instruction bitly's migration followed.
  The engine's own conformance/fixture-replay harness provides the
  credential-free test affordance this bundle needs, so no fixture-mode
  equivalent is needed here.
