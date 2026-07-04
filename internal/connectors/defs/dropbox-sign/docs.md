# Overview

Dropbox Sign (formerly HelloSign) is an e-signature platform. This bundle reads Dropbox Sign
signature requests, templates, team members, and account details, and writes signature-request/
template/team/account lifecycle mutations, through the Dropbox Sign REST API
(`https://api.hellosign.com/v3`) using HTTP Basic auth. It was originally migrated from
`internal/connectors/dropbox-sign` (the hand-written connector this bundle replaces at capability
parity; the legacy package stays registered and unchanged until wave6's registry flip) and was
expanded in Pass B to the full documented JSON-body API surface (researched against the official
OpenAPI source, `github.com/hellosign/hellosign-openapi`).

## Auth setup

Provide a Dropbox Sign API key via the `api_key` secret; it is sent as the HTTP Basic username with
a blank password (`Authorization: Basic base64(api_key:)`), matching legacy's
`connsdk.Basic(secret, "")`, and is never logged.

## Streams notes

Four streams: `signature_requests` (`GET /signature_request/list`, records at
`signature_requests`), `templates` (`GET /template/list`, records at `templates`), `team_members`
(`GET /team/members`, records at `team_members`) share the base-level `page_number` pagination
(`page`/`page_size` query params, 100 records per page, stopping on a short page) — matches
legacy's `list_info.num_pages` page-increment loop for the common case where every page except the
last is exactly `page_size` long. `account` (`GET /account`, a single-object endpoint with no
`list_info`) overrides pagination to `none` at the stream level and takes no pagination query
params at all, matching legacy's `paginated := endpoint.recordsPath != "account"` special case.

Primary keys: `signature_request_id` (signature_requests), `template_id` (templates), `account_id`
(team_members, account). Incremental cursor fields (`created_at`/`updated_at`) are declared on the
schemas that have them, matching legacy's `CursorFields`, but no stream declares an `incremental`
block: Dropbox Sign's list endpoints accept no server-side `updated_since`-style filter parameter,
matching legacy's own read path (full refresh only, no incremental request param sent).

## Write actions & risks

`capabilities.write: true` as of Pass B. Legacy was entirely read-only (signature requests are
*created* via file uploads and signer flows that do not map onto reverse-ETL record writes — that
reasoning still holds and is why creation/sending remains unmodeled), but the full documented API
also exposes a substantial JSON-body-only *lifecycle-mutation* surface on already-existing
resources that requires no file payload at all. 13 actions:

- `update_signature_request` (`POST /signature_request/update/{signature_request_id}`) — updates a
  signer's `email_address`/`name` on an in-progress request (e.g. in response to a
  `signature_request_email_bounce` event).
- `cancel_signature_request` (`POST /signature_request/cancel/{signature_request_id}`,
  `confirm: destructive`) — cancels an incomplete request; **not reversible**.
- `remind_signature_request` (`POST /signature_request/remind/{signature_request_id}`) — emails a
  signer a reminder; Dropbox Sign itself rate-limits this to once per hour per signer (manual or
  automatic reminders share the same cooldown).
- `release_hold_signature_request` (`POST /signature_request/release_hold/{signature_request_id}`)
  — releases a request held from an UnclaimedDraft, immediately sending it to every signer.
- `remove_signature_request` (`POST /signature_request/remove/{signature_request_id}`,
  `confirm: destructive`) — removes the caller's access to a *completed* request from the account's
  list view; **not reversible**.
- `delete_template` (`POST /template/delete/{template_id}`, `confirm: destructive`) — completely
  deletes a template from the account; **not reversible**.
- `add_template_user` / `remove_template_user` (`POST /template/{add_user,remove_user}/
  {template_id}`) — grants/revokes an existing Team member's access to a template.
- `create_team` (`POST /team/create`) — creates a Team with the caller as its sole member (fails if
  the caller already belongs to one).
- `update_team` (`PUT /team`) — renames the caller's own Team.
- `add_team_member` (`PUT /team/add_member`) — invites/moves a user onto the caller's Team,
  provisioning a new Dropbox Sign account for the invited email if none exists yet.
- `remove_team_member` (`POST /team/remove_member`, `confirm: destructive`) — removes a member;
  optionally transfers their documents to another account (Enterprise plans only, `new_owner_
  email_address`), which is **not reversible**.
- `update_account` (`PUT /account`) — updates the caller's own account settings, currently limited
  by Dropbox Sign to `callback_url` and `locale`.

Every action's `record_schema` requires only fields the live OpenAPI spec (`github.com/hellosign/
hellosign-openapi`) itself requires; path-only actions (`cancel_signature_request`,
`release_hold_signature_request`, `remove_signature_request`, `delete_template`) use `body_type:
"none"` since Dropbox Sign accepts no request body at all for these.

## Known limits

- **Multipart/file-upload endpoints are entirely out of scope** — the engine's write dialect
  supports only `json`/`form`(url-encoded)/`none` body types, never `multipart/form-data`. Every
  Dropbox Sign endpoint that *creates or edits document content* (`signature_request/send*`,
  `signature_request/create_embedded*`, `signature_request/edit*`, `template/create*`, `template/
  update*`, `unclaimed_draft/*`, bulk-send with template, and the entire Fax product's `fax/send`)
  requires an actual file payload and cannot be expressed; see `api_surface.json`'s
  `binary_payload`-category entries.
- The **Fax and Fax Line product surfaces** (a separate Dropbox Sign entitlement — dedicated fax
  numbers, fax sending/receiving) are excluded wholesale: legacy never touched this product and it
  is unrelated to e-signature.
- `/team/destroy` (deletes the caller's entire Team, only when it has a single member) is excluded
  as `destructive_admin`: unlike every other team action, it targets no specific record and is an
  irreversible account-structure change, not a per-record mutation.
- `account_verify` (`POST /account/verify`) is excluded as `non_data_endpoint`: it is a read-like
  existence lookup by email address with no state change, not a genuine write despite the POST
  verb.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting for
  Dropbox Sign, so none is added here either (matches legacy's real, lack-of, throttling behavior).
  Dropbox Sign's own per-signer reminder cooldown (documented on `remind_signature_request` above)
  is a server-side business rule, not a client-side rate limit, and surfaces as an ordinary 4xx
  error via `error_map` like any other request the API rejects.
- Legacy's `pageCount` reads `list_info.num_pages` and stops once `page >= numPages` (or on an
  empty page); this bundle's `page_number` paginator instead stops purely on a short page
  (`recordCount < page_size`). Both converge on the identical final record set for every real
  Dropbox Sign response shape (every page except the last is exactly `page_size` long); only a
  final page that happens to be exactly `page_size` long causes legacy to issue one additional
  (empty) request that this bundle does not. Documented as a benign pagination-loop-count
  difference, not an emitted-data deviation.
