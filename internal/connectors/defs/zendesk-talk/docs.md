# Overview

Zendesk Talk is a wave2 fan-out migration of `internal/connectors/zendesk-talk`
(the hand-written legacy connector this bundle migrates; the legacy package
stays registered and unchanged until wave6's registry flip). It reads
Zendesk Talk phone numbers, greetings, greeting categories, IVRs, and
per-agent activity statistics through the Zendesk Talk (voice) REST API.

## Auth setup

Zendesk Talk supports two credential shapes, both declared as `when`-gated
candidates on `streams.json`'s `base.auth` (first match wins) — the golden
dual-auth pattern (conventions.md §3):

1. **OAuth2 access token** (checked first, matching legacy's own precedence
   — `zendesktalk.go:253-256` checks `credentials.access_token` before
   `credentials.api_token`): provide `access_token` as a secret; sent as
   `Authorization: Bearer <access_token>`.
2. **API token** (Zendesk Admin Center > Apps and integrations > APIs >
   Zendesk API): provide `api_token` (secret) and `email` (secret) — the
   agent email address that owns the token. Sent via HTTP Basic as
   `Authorization: Basic base64("<email>/token:<api_token>")`, byte-for-byte
   identical to legacy's `connsdk.Basic(email+"/token", apiToken)`
   construction (`zendesktalk.go:263`).

**Documented config-surface deviation (base_url/subdomain)**: legacy accepts
a bare `subdomain` config value and derives
`https://<subdomain>.zendesk.com/api/v2/channels/voice` itself
(`zendeskTalkBaseURL`, `zendesk_talk.go:289-317`), or an explicit `base_url`
override with automatic Talk-API-path suffixing. The engine's declarative
`url` template is a single static string with no conditional branching or
suffix-appending logic, so this bundle requires `base_url` to already
include the full Talk API path (e.g.
`https://acme.zendesk.com/api/v2/channels/voice`) — matching
zendesk-support's and zendesk-chat's identical documented deviation
(conventions.md §3: a declared-but-unwireable `subdomain` key is worse than
an absent one). Every legacy-accepted `subdomain`-only configuration is
still reachable by supplying the fully-qualified URL as `base_url`.

## Streams notes

All 5 streams (`phone_numbers`, `greetings`, `greeting_categories`, `ivrs`,
`agents_activity`) share the same shape: `GET` against the Zendesk Talk
collection endpoint (`agents_activity`'s path is `/stats/agents_activity`,
matching legacy's `resource: "stats/agents_activity"`), records extracted
from the response's own top-level key (matching legacy's `recordsKey`
routing table), and `per_page=100` sent via each stream's static `query`
(matches legacy's `zendeskTalkPageSize` constant) — mirrors stripe's
`limit=100`-via-static-query precedent (conventions.md's F6 lesson).
`phone_numbers` declares `x-cursor-field: created_at` for manifest-surface
parity (legacy's own `CursorFields: []string{"created_at"}"`), but — as with
zendesk-support — no stream declares an `incremental` block: legacy's
`harvest()` (`zendesk_talk.go:129-168`) never sends any lower-bound filter
on any stream, always reading the full collection. `agents_activity`'s
primary key is `["agent_id"]`, matching legacy's own
`PrimaryKey: []string{"agent_id"}` (this stream has no independent `id`
field at all — a per-agent snapshot row).

Pagination follows Zendesk Talk's `next_page` body-field convention
(`pagination.type: next_url`, `next_url_path: "next_page"`): the next page's
full URL is read from the response body's `next_page` key and requested
verbatim until it is `null`/empty, matching legacy's `harvest()` loop
exactly (`next == "" || next == "null"` stop rule, `zendesk_talk.go:158-161`).
This is the same "absolute next-page URL read from the body" shape as
bitly's `pagination.next`/aircall's `next_url`, just under a different body
key.

## Write actions & risks

None. Legacy `zendesk-talk` is read-only (`Write` returns
`connectors.ErrUnsupportedOperation`); `metadata.json` declares
`capabilities.write: false` and this bundle ships no `writes.json`.

## Known limits

- Full Zendesk Talk API surface (calls, call legs, current queue activity,
  account overview, lines, quality scores) is out of scope for wave2; see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries. Only the 5 legacy-parity streams are
  implemented.
- **Every stream's fixture is single-page** (conventions.md §4's sanctioned
  `next_url` exception, applied uniformly here since every stream in this
  bundle shares the identical `next_url`/`next_page` pagination shape at
  the base level): the next-page URL is the replay server's own address,
  unknown until the harness picks a port at runtime, so a static fixture
  cannot embed a correct second page for any stream. `pagination_terminates`
  is proven against the bundle's first stream (`phone_numbers`), whose
  single fixture page sets `next_page: null` so the read terminates after
  exactly one request — the same termination proof any single-page
  `next_url` fixture provides. No live `paritytest` package exists for this
  connector in this wave (JSON+docs-only fan-out); a follow-up wave may add
  one to exercise real 2-page `next_page` correctness, matching bitly's/
  calendly's precedent.
- **`per_page=100` is not runtime-configurable.** Legacy hardcodes
  `zendeskTalkPageSize = 100` with no config override at all (unlike
  zendesk-chat/zendesk-support, legacy zendesk-talk never reads a
  `page_size` config key), so this bundle's static `query: {"per_page":
  "100"}` is a direct, unconditional port — not a narrowing.
- **Legacy's fixture-mode-only fields are not modeled.** Legacy's
  `readFixture` path (only reached when `config.mode == "fixture"`, a
  credential-free conformance-harness affordance) stamps a
  `previous_cursor` field (echoing `req.State["cursor"]` when set) onto
  every fixture-mode record (`zendesk_talk.go:214-218`), plus many
  synthetic-only filler fields not present in any live Zendesk Talk response
  shape. None of these are part of the LIVE record shape; this bundle's
  schemas target the live path only (`harvest()`), per the same instruction
  bitly's and zendesk-chat's migrations followed. The engine's own
  conformance/fixture-replay harness provides the credential-free test
  affordance this bundle needs, so no fixture-mode equivalent is needed
  here.
