# Overview

Drift is a conversational-marketing platform. This bundle reads Drift users, accounts,
conversations, contacts, and teams, and writes contact/account/message/conversation/
timeline-event/GDPR mutations through the Drift REST API (`https://driftapi.com`) using Bearer
auth. It originally migrated `internal/connectors/drift` (the hand-written connector this bundle
replaces at capability parity) as a read-only bundle; this Pass B pass adds the `teams` read stream
and 11 write actions, researched directly against `https://devdocs.drift.com/llms.txt`'s full
AI-agent doc index and each linked endpoint page. The legacy package stays registered and unchanged
until wave6's registry flip.

## Auth setup

Provide a Drift OAuth access token via the `access_token` secret; it is used only for Bearer auth
(`Authorization: Bearer <access_token>`) and is never logged. Legacy accepted three secret-key
aliases (`credentials.access_token`, `access_token`, `credentials_access_token`) for robustness to
how the secret is surfaced; this bundle's declarative `secrets.<key>` reference is a single exact
key, so only the canonical `access_token` key is supported here — see Known limits. The same token's
OAuth scopes (`contact_read`/`contact_write`, `conversation_read`/`conversation_write`,
`user_read`/`user_write`, `account_read`/`account_write`, `team_read`, `gdpr_read`/`gdpr_write`)
gate which streams/writes actually succeed; a token missing a write scope gets a 403 from Drift,
surfaced via this bundle's `error_map`.

## Streams notes

Five streams, three distinct pagination shapes (matching legacy's `paginationKind` enum exactly for
the original four; `teams` is new and unpaginated):

- `users` (`GET /users/list`, records at `data`) — single-shot, no pagination.
- `accounts` (`GET /accounts`, records at `data.accounts`) — `next_url` pagination reading the
  absolute next-page URL from `data.next`; the initial request sends `size=65&index=0` (Drift's
  page-size cap for this endpoint, matching legacy's `driftAccountsPageSize`).
- `conversations` (`GET /conversations/list`, records at `data`) — `cursor` pagination with
  `token_path: pagination.next` and `stop_path: pagination.more`: the next page's `next` query
  param is read from the body's `pagination.next`, and pagination stops when `pagination.more` is
  falsy (matching legacy's `more != "true" || next == ""` stop condition) — the initial request
  sends `limit=50` (legacy's `driftConversationsPageSize`). The current Drift docs name this same
  query parameter `page_token`; legacy's own wire behavior sends `next` (verified against
  `internal/connectors/drift/drift.go`'s `query.Set("next", next)`), and this bundle keeps that
  exact parity-locked parameter name rather than "fixing" it to match the docs, since legacy's
  behavior against the real API is the parity bar, not the docs' current parameter name.
- `contacts` (`GET /contacts`, records at `data`) — single-shot, no pagination; an optional
  `email` config value is sent as a query filter (`omit_when_absent: true`) when set, matching
  legacy's `contacts`-only email lookup passthrough.
- `teams` (`GET /teams/org`, records at `data`, new this pass) — single-shot, no pagination;
  returns every team in the org (id/name/members/owner/status/availability-mode fields).

None of the five streams expose a server-side incremental filter (Drift's list endpoints accept no
`updated_since`-style parameter), so no stream declares an `incremental` block — matching legacy's
`InitialState` always returning an empty cursor (full refresh only) for the original four, and
extending the same full-refresh-only shape to `teams`.

## Write actions & risks

11 actions, all against Drift's real documented wire shapes:

- `create_contact` (`POST /contacts`, body `{attributes: {email, externalId?}}`) — low risk, no
  approval required.
- `update_contact` (`PATCH /contacts/{id}`, body `{attributes: {...}}`) — mutates any standard or
  custom contact attribute; approval required.
- `delete_contact` (`DELETE /contacts/{id}`, idempotent on 404, `confirm: destructive`) —
  permanently removes a contact.
- `post_timeline_event` (`POST /contacts/timeline`, body `{contactId, event, createdAt?,
  externalId?, attributes?}`) — low risk, no approval required.
- `create_account` (`POST /accounts/create`, body `{ownerId, domain, name?, targeted?,
  customProperties?}`) — low risk, no approval required.
- `update_account` (`PATCH /accounts/update`, body `{accountId, ownerId, ...}` — Drift's real API
  requires `accountId` inside the body itself, not the path, unlike every other resource in this
  bundle) — approval required.
- `delete_account` (`DELETE /accounts/{account_id}`, idempotent on 404, `confirm: destructive`) —
  permanently removes an account.
- `create_message` (`POST /conversations/{conversation_id}/messages`, body `{type, body?, userId?,
  buttons?}`) — posts into a LIVE conversation; a `type: chat` message is visible to the end
  customer immediately, so this is `confirm: destructive` despite being a "create", since its
  external-facing blast radius is closer to an irreversible customer-visible action than an
  ordinary internal record creation.
- `create_conversation` (`POST /conversations/new`, body `{email}`) — approval required.
- `gdpr_retrieve` (`POST /gdpr/retrieve`, body `{email}`, `kind: custom`) — triggers Drift to
  compile and email a data-subject-access-request bundle to the account's admin; approval required.
- `gdpr_delete` (`POST /gdpr/delete`, body `{email}`, `confirm: destructive`) — permanently erases
  every contact/user record matching the given email; irreversible, approval required.

Excluded from this pass (see `api_surface.json` for the endpoint-by-endpoint reasoning): SCIM user
provisioning (a separate identity-management sub-product with its own SCIM auth model), the App
Admin API (`app/uninstall`/`app/token_info` — OAuth-app integration lifecycle, not customer data),
playbooks/conversational-landing-pages/booked-meetings (read-only detail/config resources with no
create/update/delete counterpart), custom-attribute-definition listing (schema introspection, not
object data), and `POST /emails/unsubscribe` — a genuine dialect gap: its real wire body is a bare
JSON array of email strings (`["a@x.com", "b@y.com"]`), not an object, and this engine's write
dialect (`write.go`) only ever constructs an object body from a record's fields; there is no way to
express a bare-array body in `writes.json` today.

## Known limits

- Secret-key aliasing: legacy accepted `credentials.access_token`/`access_token`/
  `credentials_access_token` as equivalent secret keys (first-match-wins). This bundle's
  declarative auth only resolves a single exact secret key (`access_token`) — any caller must use
  that canonical key. This narrows accepted CONFIGURATION surface (which secret key name works),
  never emitted record data, so it is documented here rather than in the parity-deviation ledger.
- `POST /emails/unsubscribe` is not implemented: its wire body is a bare JSON array, not an object,
  which `writes.json`'s per-record-object body-construction dialect cannot express (see Write
  actions & risks above). This is a genuine, single-occurrence dialect limitation, not a scoping
  choice — documented here rather than filed as an `ENGINE_GAP` mini-wave trigger since it has not
  recurred elsewhere in this migration (conventions.md §6's recurrence threshold is ≥3).
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting for
  Drift, so none is added here either (matches legacy's real, lack-of, throttling behavior).
- `create_message`/`create_conversation`/`gdpr_retrieve`/`gdpr_delete` all have real, immediate,
  customer- or compliance-facing side effects (a live chat message becomes visible instantly; a
  GDPR delete is irreversible) — treat every write against this connector as touching production
  customer-facing state, not an internal system-of-record update.
