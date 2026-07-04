# Overview

Vitally is a customer-success platform. This bundle reads and writes customer-success **accounts**,
**users**, **notes**, **conversations**, **tasks**, and **NPS responses** via the Vitally REST API
(default base URL `https://rest.vitally.io`). It originally migrated
`internal/connectors/vitally` (188 loc, accounts-only, read-only), which stays registered and
unchanged until wave6's registry flip. Pass B full-surface expansion (this revision) researched the
live REST API docs (`docs.vitally.io/en/collections/10410457-rest-api`) and found the documented
API "currently supports creating, updating, retrieving and listing Users, Accounts, Conversations,
Tasks, Notes and NPS Responses" — every one of those 6 resources is now a covered stream, and every
create/update/delete action Vitally documents for them is now a `writes.json` action.
`capabilities.write` is now `true`. Engine-vs-legacy parity for the original accounts-only surface
is tested in `internal/connectors/paritytest/vitally/parity_test.go` per SPEC.md §6's per-connector
parity-test-package decision; the 5 newly-added streams/writes have no legacy counterpart to prove
parity against (they did not exist in `internal/connectors/vitally`) and are proven by this
bundle's own fixture-replay conformance suite instead.

## Auth setup

Vitally authenticates via a single secret, `basic_auth_header`, which must contain the **entire,
pre-built `Authorization` header value** (e.g. `Basic <base64-encoded apiKey:>`) — not a bare API
key or token. This mirrors legacy exactly: `vitally.go:100-104` builds the requester's
authenticator as `connsdk.APIKeyHeader("Authorization", auth, "")`, i.e. an arbitrary-header
authenticator with an **empty prefix**, so whatever string is configured in `basic_auth_header` is
sent as the `Authorization` header value completely unmodified (no re-encoding, no `Basic ` prefix
added by the connector itself — the caller is expected to have already produced the full header
value, typically by base64-encoding an `apiKey:` pair themselves before configuring the secret).
This matches the real API's own documented auth (`Authorization: Basic <base64 apiKey:>`, one
colon-terminated value with no password half).

This bundle reproduces that exact behavior using the engine's `api_key_header` auth mode:
`{"mode":"api_key_header","header":"Authorization","value":"{{ secrets.basic_auth_header }}","prefix":""}`.
The engine's `basic` auth mode was deliberately **not** used here, even though Vitally's own API
uses HTTP Basic auth conceptually — `basic` mode base64-encodes a `username:password` pair at
request time (`connsdk.Basic`), which would require decomposing `basic_auth_header` back into its
constituent parts (information the connector never receives; the secret already IS the encoded
header). Using `api_key_header` with an empty prefix is the byte-exact reproduction, asserted by
`TestParityVitally_AuthorizationHeaderByteExact`.

## Streams notes

All 6 List endpoints share the real API's uniform cursor-pagination envelope
(`{"results": [...], "next": <cursor-or-null>}`), documented on the REST API Overview page: `limit`
(default/max 100, wired via `spec.json`'s new `page_size` config, default `100`) and `from` (the
opaque cursor returned as `next`) — modeled as `pagination: {"type": "cursor", "cursor_param":
"from", "token_path": "next"}` at the bundle level, shared by every stream. `next: null` stops
pagination (no `stop_path` needed — Vitally's own docs guarantee `next` is exactly `null` at the
end, unlike Zendesk's "may still be populated" caveat).

- **`accounts`** (`GET /resources/accounts`, `results` array): unchanged path from the original
  migration, now with the full real record shape (health/NPS scores, CSM/AE ids, timestamps,
  `keyRoles`, `segments`) rather than legacy's narrow `{id, name, traits}` projection — Pass B's
  full-surface mandate. The previously-undocumented-here `status` filter (`active` / `churned` /
  `activeOrChurned`, defaulting server-side to `active`) is now wired via `spec.json`'s
  `account_status` enum and the optional-query dialect (`omit_when_absent: true`) — this is now
  expressible because the engine's `stream.Query` object form (conventions.md §3) exists; it did
  not exist at accounts-only migration time (see the superseded "Known limits" entry below, now
  resolved).
- **`users`** (`GET /resources/users`, `results` array): full real record shape including nested
  `accounts`/`organizations` association arrays (read-only associations — see Known limits).
- **`notes`** (`GET /resources/notes`, `results` array): real field names are camelCase
  (`accountId`, `noteDate`, `authorId`, etc.); every schema property beyond a same-name passthrough
  is wired via `computed_fields` renames to this bundle's snake_case schema.
- **`conversations`** (`GET /resources/conversations`, `results` array): the list endpoint's own
  docs note it does NOT include the `messages` array (messages are fetched per-conversation via the
  detail GET, which this bundle excludes as `duplicate_of`/out_of_scope — see `api_surface.json`).
- **`tasks`** (`GET /resources/tasks`, `results` array): same camelCase-to-snake_case
  `computed_fields` pattern as notes.
- **`nps_responses`** (`GET /resources/npsResponses`, `results` array): NPS responses are unique on
  `externalId`, so `create_nps_response`'s POST doubles as an upsert (Vitally's own documented
  behavior — see Write actions & risks).

All 6 streams declare `x-cursor-field` (`updatedAt`/`updated_at`) matching each resource's own
`updatedAt` field, but **no `incremental` block**: the real List endpoints support only a `sortBy`
ordering hint (`updatedAt` default or `createdAt`), never a since/after server-side filter
parameter, so there is no `request_param` for the engine's `incremental` machinery to drive
(conventions.md §8 rule 2's truth table: declare `incremental` only when legacy/the real API
genuinely supports a server-side filter — it does not here).

## Write actions & risks

14 write actions across all 6 resources, added in this Pass B expansion (legacy shipped none —
`Write` always returned `ErrUnsupportedOperation`). All are external mutations visible to the
vendor's CS team; **approval required** for every one:

- `create_account` / `update_account` — `POST`/`PUT /resources/accounts[/{id}]`.
- `create_user` / `update_user` — `POST`/`PUT /resources/users[/{id}]`. Vitally's REST API does not
  support unlinking a User from an Account/Organization (only new associations) — this bundle does
  not attempt to model unlinking as a write action since there is no corresponding endpoint.
- `create_note` / `update_note` / `delete_note` — `POST`/`PUT`/`DELETE /resources/notes[/{id}]`.
  `update_note`'s `tags` field is documented as **replace-the-whole-set**, not merge: omitting
  `tags` from the update body leaves existing tags untouched, but including a partial `tags` array
  overwrites the full set — ported verbatim from Vitally's own docs, not this bundle's design
  choice (see the `record_schema` description on `update_note`).
- `create_conversation` / `update_conversation` / `delete_conversation` —
  `POST`/`PUT`/`DELETE /resources/conversations[/{id}]`. Vitally's own docs are explicit that these
  endpoints **do not start real conversations or send outbound messages** — they only create/update
  a historical record inside Vitally as a reference point.
  `delete_conversation` is `confirm: "destructive"` (deletes all messages too).
- `create_task` / `update_task` — `POST`/`PUT /resources/tasks[/{id}]`. There is no documented
  `DELETE /resources/tasks/{id}` endpoint, so no `delete_task` action exists (unlike notes/
  conversations, which do document DELETE).
- `create_nps_response` / `update_nps_response` — `POST`/`PUT /resources/npsResponses[/{id}]`.
  `create_nps_response` is a documented upsert-by-`externalId`: POSTing an `externalId` that
  already exists updates that response rather than erroring — this is Vitally's own behavior, not
  a bundle-side merge.

`delete_note`/`delete_conversation` both declare `delete.missing_ok_status: [404]` (idempotent
delete — a second delete of an already-deleted/archived record is treated as success, matching the
general repo-wide delete-semantics convention, conventions.md §3).

## Known limits

- **`Check` now dials the network; legacy's `Check` never did.** Legacy `Check` (`vitally.go:33-47`)
  validates config/secret presence offline only — it never issues an HTTP request. This bundle's
  `base.check` (`streams.json`) issues a real `GET /resources/accounts`, matching the wave's
  general "fail loud, not fail silent" preference for `Check` (a bad credential or unreachable host
  is now caught at `Check` time instead of surfacing on the first `Read`). This is a deliberate,
  strictly-improving behavior change with zero record-data impact, but it IS a behavior deviation
  worth naming explicitly: a `Check` that previously always succeeded offline (given well-formed
  config/secrets) can now fail on a network outage or an invalid credential that legacy would only
  have caught on `Read`.
- **RESOLVED (Pass B): the optional `status` query filter is now modeled.** The original
  accounts-only migration's "Known limits" recorded this as an out-of-scope gap because the engine
  dialect at that time had no way to send a query param only when a config value is set. The
  `stream.Query` opt-in optional-query object form (conventions.md §3, `omit_when_absent`) now
  exists and is used for `accounts`' `status` filter (via `spec.json`'s new `account_status` enum
  property) and every stream's incremental-adjacent filtering needs generally.
  Resolved rather than re-documented as an open gap, per conventions.md §5's meta-rule.
- **`users`' `accounts`/`organizations` association arrays are read-only pass-through.** The real
  API documents that NEW associations can be established by including `accountIds`/
  `organizationIds` in `create_user`'s body, but **unlinking** a User from an Account/Organization
  is explicitly NOT supported via this REST API (Vitally's docs point to a separate Analytics API
  `POST /analytics/v1/unlink` endpoint instead) — out of scope for this connector, which targets
  the REST API only.
- **Conversations' `messages` array is not modeled as stream data.** The list endpoint's own docs
  state messages are omitted from list reads; only the single-conversation detail GET includes them
  (excluded from this bundle's surface — see `api_surface.json`). A future pass could model a
  `messages`-per-conversation fan-out stream (the engine's `fan_out` dialect, conventions.md §3,
  would fit this shape) if message-level sync becomes a requirement.
- **Legacy's `fixture` mode is not part of the bundle.** Legacy's `mode=fixture` config value
  short-circuits network access and emits one synthetic record stamped with a `fixture: true`
  field (`vitally.go:64-65,107-112`) — this is a legacy testing affordance, not part of the real
  API's record shape, and parity (for the original accounts stream) is asserted against legacy's
  LIVE read path via `httptest` (SPEC.md §5.1's xkcd note documents the same principle for
  fixture-mode connectors generally).
