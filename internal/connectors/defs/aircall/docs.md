# Overview

Aircall is a wave2 fan-out declarative-HTTP migration, expanded in Pass B to the full documented
Aircall surface (developer.aircall.io/api-references/). It reads Aircall calls, users, contacts,
numbers, teams, tags, and webhooks, and writes user/team/contact/tag/webhook create-update-delete
plus call archive/unarchive/comment/tag actions, through the Aircall REST API
(`https://api.aircall.io/v1/...`). This bundle originally targeted read-only capability parity with
`internal/connectors/aircall` (the hand-written connector it migrates, itself read-only); the legacy
package stays registered and unchanged until wave6's registry flip, so this bundle's write surface
is a genuine capability expansion beyond legacy, not a parity port.

## Auth setup

Provide `api_id` and `api_token` secrets; they are sent as HTTP Basic auth (`api_id:api_token`,
base64-encoded) and never logged, matching legacy's `connsdk.Basic(id, token)`
(`aircall.go:257`). `base_url` defaults to `https://api.aircall.io/v1` (legacy's
`aircallDefaultBaseURL`) and may be overridden for tests/proxies.

## Streams notes

All seven streams share Aircall's `meta.next_page_link` envelope: `GET /<resource>` returns
`{"<resource>":[...],"meta":{"next_page_link":<url|null>,...}}`, records live at the
resource-named top-level key (identical to the resource segment: `calls`/`users`/`contacts`/
`numbers`/`teams`/`tags`/`webhooks`). `tags` and `webhooks` are Pass B additions (the full
documented `GET /v1/tags` and `GET /v1/webhooks` list endpoints); neither carries a `created_at`/
`updated_at` field in Aircall's documented response shape, so neither declares an `incremental`
block (per §8's incremental truth table: no incremental filter, keep no `x-cursor-field`). Pagination
is `next_url` (`next_url_path: meta.next_page_link`) — Aircall
returns a fully-qualified absolute next-page URL, matching legacy's own "follow next_page_link
verbatim" behavior (`aircall.go:190-193`) and the engine's `next_url` paginator's same-host SSRF
guard (THREAT-MODEL §3), which passes cleanly in production since Aircall's own `next_page_link`
is always same-origin as `base_url`. `per_page=50` (Aircall's own default, `aircallDefaultPageSize`)
is a static per-stream query value re-sent on every page request the engine issues; see Known
limits for the same "re-sent vs. legacy's reset-to-empty-then-follow" divergence already documented
on bitly's bundle (this wave's `next_url` sibling pilot).

`calls` and `contacts` are the two streams Aircall's own API supports a `from` (unix-seconds) lower
bound filter on (legacy's `harvest`, `aircall.go:156-158`: `if fromUnix != "" && (endpoint.resource
== "calls" || endpoint.resource == "contacts")`); this bundle expresses that exact same-branch
gating via `incremental.request_param: "from"` (declared only on `calls`', cursor field
`started_at`, and `contacts`', cursor field `created_at`, `incremental` blocks) with
`param_format: unix_seconds` (matching legacy's `toUnixSeconds` conversion of the RFC3339
`start_date`/state cursor) — the engine's `buildInitialQuery` sends `request_param` only when the
formatted lower bound is non-empty, identical to legacy's own `if fromUnix != ""` gate. `users`,
`numbers`, and `teams` declare no `incremental` block and no `from` query key at all, matching
legacy's `harvest` never setting `from` for those three resources (the `base.Set("from", ...)` call
is conditioned on the resource name, so `users`/`numbers`/`teams` never receive it either way).

## Write actions & risks

Pass B flips `capabilities.write` to `true` (a genuine capability expansion beyond legacy, which was
fully read-only — "no obvious safe reverse-ETL surface" was legacy's own package doc, but the full
documented Aircall API does expose real dialect-expressible mutations). 18 actions in `writes.json`:

- **Users**: `create_user`/`update_user`/`delete_user` (`POST`/`PUT`/`DELETE /v1/users(/:id)`) —
  deleting a user permanently frees an Aircall agent seat/license; approval required for
  create/update/delete.
- **Teams**: `create_team`/`delete_team` (`POST`/`DELETE /v1/teams(/:id)`), plus
  `add_user_to_team`/`remove_user_from_team` (`POST`/`DELETE /v1/teams/:team_id/users/:user_id`,
  both bodyless path-parameterized mutations — `body_type: "none"`) for team membership.
- **Contacts**: `create_contact`/`update_contact`/`delete_contact` (`POST`/`PUT`/
  `DELETE /v1/contacts(/:id)`) — `update_contact`'s `PUT` replaces the full contact record
  including `phone_numbers`/`emails` arrays; the per-sub-item add/update/delete endpoints
  (`/contacts/:id/phone_numbers/...`, `/contacts/:id/emails/...`) are not separately modeled (see
  `api_surface.json`), since the same outcome is reachable via a full-record `update_contact` PUT.
- **Tags**: `create_tag`/`update_tag`/`delete_tag` (`POST`/`PUT`/`DELETE /v1/tags(/:id)`).
- **Webhooks**: `create_webhook`/`update_webhook`/`delete_webhook` (`POST`/`PUT`/
  `DELETE /v1/webhooks(/:id)`) — `create_webhook`/`update_webhook` register/repoint a live outbound
  HTTP callback; verify the target `url` before enabling.
- **Calls**: `archive_call`/`unarchive_call` (`PUT /v1/calls/:id/archive` /`/unarchive`, both
  bodyless), `comment_call` (`POST /v1/calls/:id/comments`, body restricted to `content` via
  `body_fields`), `tag_call` (`POST /v1/calls/:id/tags`, body restricted to `tag_ids`).

Every action uses `body_type: "json"` (default JSON body construction, `path_fields` excluding the
id(s) already in the path) except the four bodyless mutations above, which use `body_type: "none"`.
No action needs a hook: every one of these operations is a single JSON/bodyless HTTP request with no
signature auth, multipart body, or compound follow-up call. Excluded live-telephony/messaging/
dialer/conversation-intelligence/analytics-export mutations are documented per-endpoint in
`api_surface.json` (categories `destructive_admin`/`requires_elevated_scope`/`out_of_scope`, e.g.
starting an outbound call, sending a live SMS/WhatsApp message, or triggering an AI Voice Agent
call — all real-world telephony/messaging side effects, not reverse-ETL record mutations).

## Known limits

- **`next_url` fixtures are single-page, per the sanctioned exception (conventions.md §4).** A
  `next_url` stream's next-page URL is the replay server's own runtime address, unknown until the
  harness picks a port — a static fixture file cannot embed the correct absolute URL for a second
  page. Every stream in this bundle ships a single-page fixture (satisfies `fixtures_present`/
  `read_fixture_nonempty`); `pagination_terminates` passes on the first stream (`calls`) with its
  single page (`hits == len(pages) == 1`), which is not a 2-page pagination proof but is not a
  false failure either. Real 2-page `next_url`-following correctness for THIS bundle's exact request
  shape is proven by legacy's own existing test
  (`internal/connectors/aircall/aircall_test.go`'s `TestReadPaginatesAndAuthenticates`, which drives
  a real 2-page `httptest.Server` and asserts the second page is requested via the served
  `next_page_link`), plus the engine's own generic `next_url` paginator unit tests
  (`internal/connectors/engine/paginate_test.go`'s `TestNewPaginatorNextURLFollowsAbsoluteURL` and
  siblings) and read-path integration test
  (`internal/connectors/engine/read_test.go`'s `TestReadNextURLPaginationSetsBaseHostFromRequester`).
  This wave does not add a new `paritytest/aircall` package (out of scope per this wave's JSON-only
  mandate); a future wave adding hand-written parity suites should follow bitly's/calendly's
  `TestParity<Name>_..._TwoPagePagination` pattern for aircall specifically.
- **`per_page` is re-sent on every page request, unlike legacy's reset-to-empty-then-follow.** The
  engine's `readDeclarative` loop merges `stream.Query` into every page request regardless of
  pagination type, and `connsdk.Requester.resolveURL` re-applies that merged query onto the absolute
  next-page URL (replacing any same-named param already present). Legacy instead resets to an empty
  `url.Values{}` once it follows an absolute next-page URL (mirrors bitly's identical, already-ledgered
  divergence in this same wave — see bitly's `docs.md`). This is benign in DATA terms only because
  Aircall's own `next_page_link` already carries the identical `per_page` value the engine
  re-applies (the replace is idempotent); if Aircall's `next_page_link` ever diverged from
  `per_page`, this bundle's request would differ from legacy's — today it does not.
- **`per_page`/`max_pages` config overrides are not modeled.** Legacy exposes `per_page` (1-50,
  default 50) and `max_pages` (0/all/unlimited or a positive integer cap) as config-driven overrides
  (`aircallPageSize`/`aircallMaxPages`). The engine's `next_url` paginator has no config-driven
  page-size or request-count-cap knob at all (mirrors bitly's identical, already-ledgered
  limitation); `per_page`/`max_pages` are therefore not declared in `spec.json`, and this bundle
  sends Aircall's own default (`per_page=50`) as a static per-stream query literal.
- **Write actions are a Pass B capability expansion beyond legacy, not a parity port.** Legacy's
  `Write` always returns `connectors.ErrUnsupportedOperation` ("no obvious safe reverse-ETL
  surface"); this bundle's 18 write actions have no legacy behavior to stay in parity with, so there
  is no parity-deviation ledger entry for them (§5's meta-rule only applies to a deviation from
  legacy behavior). No `paritytest/aircall` package exists yet (out of scope per the JSON-only
  mandate); correctness for each write action is proven by `conformance`'s
  `write_request_shape:<action>` dynamic check against `fixtures/writes/<action>.json`.
- **Detail-by-id GET endpoints are not modeled as streams.** `/v1/users/:id`, `/v1/teams/:id`,
  `/v1/calls/:id`, `/v1/contacts/:id`, `/v1/tags/:id`, and `/v1/webhooks/:id` each require an id a
  caller must already have (typically obtained from the corresponding list stream); none is a bulk
  syncable resource in its own right, so each is `excluded: {category: "duplicate_of"}` in
  `api_surface.json` rather than a stream.
