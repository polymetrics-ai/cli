# Overview

Drift is a conversational-marketing platform. This bundle reads Drift users, accounts,
conversations, and contacts through the Drift REST API (`https://driftapi.com`) using Bearer
auth. It is read-only, migrated from `internal/connectors/drift` (the hand-written connector this
bundle replaces at capability parity); the legacy package stays registered and unchanged until
wave6's registry flip.

## Auth setup

Provide a Drift OAuth access token via the `access_token` secret; it is used only for Bearer auth
(`Authorization: Bearer <access_token>`) and is never logged. Legacy accepted three secret-key
aliases (`credentials.access_token`, `access_token`, `credentials_access_token`) for robustness to
how the secret is surfaced; this bundle's declarative `secrets.<key>` reference is a single exact
key, so only the canonical `access_token` key is supported here — see Known limits.

## Streams notes

Four streams, three distinct pagination shapes (matching legacy's `paginationKind` enum exactly):

- `users` (`GET /users/list`, records at `data`) — single-shot, no pagination.
- `accounts` (`GET /accounts`, records at `data.accounts`) — `next_url` pagination reading the
  absolute next-page URL from `data.next`; the initial request sends `size=65&index=0` (Drift's
  page-size cap for this endpoint, matching legacy's `driftAccountsPageSize`).
- `conversations` (`GET /conversations/list`, records at `data`) — `cursor` pagination with
  `token_path: pagination.next` and `stop_path: pagination.more`: the next page's `next` query
  param is read from the body's `pagination.next`, and pagination stops when `pagination.more` is
  falsy (matching legacy's `more != "true" || next == ""` stop condition) — the initial request
  sends `limit=50` (legacy's `driftConversationsPageSize`).
- `contacts` (`GET /contacts`, records at `data`) — single-shot, no pagination; an optional
  `email` config value is sent as a query filter (`omit_when_absent: true`) when set, matching
  legacy's `contacts`-only email lookup passthrough.

None of the four streams expose a server-side incremental filter in the legacy connector (Drift's
list endpoints accept no `updated_since`-style parameter), so no stream declares an `incremental`
block — matching legacy's `InitialState` always returning an empty cursor (full refresh only).

## Write actions & risks

None. Drift is read-only in both legacy and this bundle (`capabilities.write: false`); no
`writes.json` file is shipped.

## Known limits

- Secret-key aliasing: legacy accepted `credentials.access_token`/`access_token`/
  `credentials_access_token` as equivalent secret keys (first-match-wins). This bundle's
  declarative auth only resolves a single exact secret key (`access_token`) — any caller must use
  that canonical key. This narrows accepted CONFIGURATION surface (which secret key name works),
  never emitted record data, so it is documented here rather than in the parity-deviation ledger.
- Full Drift API surface (playbook management, message sending, widget configuration) is out of
  scope for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting for
  Drift, so none is added here either (matches legacy's real, lack-of, throttling behavior).
