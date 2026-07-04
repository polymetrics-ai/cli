# Overview

Customerly was quarantined in wave1 for an `ENGINE_GAP`: both streams (`users`, `leads`) genuinely
paginate from page 0 (legacy's `for page := 0; ...` sends `page=0` as the true first request), and
the engine's `page_number` paginator unconditionally coerced a zero `StartPage` to 1
(`connsdk.PageNumberPaginator.Start()`). This gap was closed by the S4 engine mini-wave's
`PaginationSpec.StartPage *int` (`"start_page": 0` is now distinguishable from an omitted key) —
this bundle is the unblock build using that dialect addition. It reads Customerly users, leads, and
accounts, and writes user/lead/tag/message/attribute/company mutations through the Customerly v1
REST API. This bundle migrates `internal/connectors/customerly` (the hand-written connector it
replaces at capability parity); the legacy package stays registered and unchanged until wave6's
registry flip.

**Pass B full-surface expansion** (2026-07-04): reviewed the full documented Customerly API
surface against the Apiary blueprint (`https://customerlypublic.docs.apiary.io/`, the only
published machine-readable reference for this API — `docs.customerly.io`'s collection index links
to prose tutorials, not a spec). Added the `accounts` read stream and 9 write actions
(`delete_user`, `delete_lead`, `unsubscribe_user`, `add_tag`, `delete_tag`, `send_message`,
`add_user_attributes`, `add_company_attributes`, `add_user_to_company`); `capabilities.write` is
now `true`. See `api_surface.json` for the complete endpoint-by-endpoint disposition and Known
limits below for what remains out of reach.

## Auth setup

Provide a Customerly API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged.

## Streams notes

Both streams (`users` -> `GET /users/list`, `leads` -> `GET /leads/list`) share the identical
shape: records live at `data.users`/`data.leads` respectively, and every request sends
`sort=last_update&sort_direction=desc` as static query params (matching legacy's `harvest`
query construction exactly). Pagination is genuinely 0-indexed
(`pagination.type: page_number`, `page_param: page`, `size_param: per_page`, `start_page: 0`,
`page_size: 50` matching legacy's `customerlyDefaultPageSize`) — the first request sends `page=0`,
matching legacy's `for page := 0; ...` loop exactly; a page returning fewer than `per_page` records
stops the read (legacy's `len(records) < pageSize` short-page stop).

Neither `users` nor `leads` declares an `incremental` block: legacy's `harvest` sends no
server-side filter parameter derived from a cursor or `start_date` config value at all (only the
static `sort`/`sort_direction`/`page`/`per_page` params above) — per conventions.md §8 rule 2, an
`incremental` block is only declared when legacy actually sends a server-side filter, which it does
not here. `x-cursor-field: last_update` is still declared on both schemas (matching legacy's
published `CursorFields`) for catalog/sync-mode-derivation parity even though no request-time
filtering happens.

`accounts` (`GET /accounts/`) is new in Pass B: a flat, unpaginated list of `{account_id, email}`
pairs for every account on the workspace (the Apiary blueprint documents no `page`/`per_page`
params for this endpoint at all), so its stream declares a `pagination: {"type": "none"}` override
against the base `page_number` spec; records live at the bare top-level `data` array.

## Write actions & risks

Nine write actions, all flat-JSON-body mutations against documented Customerly endpoints:

- `delete_user` (`DELETE /users?user_id=...`) / `delete_lead` (`DELETE /leads?email=...`) —
  irreversibly delete a user or lead and every conversation/survey/campaign record tied to them.
  Both declare `delete.missing_ok_status: [404]` (idempotent delete) and send no body
  (`body_type: none`); the identifier is threaded into the request path's query string via
  `path_fields`/an inline `{{ record.* }}` template (the write dialect has no separate `query`
  field on a write action — the identifier must live in the templated `path` string itself, exactly
  like the two GET-by-identifier lookups this bundle deliberately excludes as `duplicate_of` in
  `api_surface.json`).
- `unsubscribe_user` (`POST /users/unsubscribe/{user_id}`) — marks a live user unsubscribed from
  messaging; empty request body per the documented shape.
- `add_tag` (`POST /tags`) — adds (or, with `untag: true`, removes) a tag across a batch of
  `users`/`leads` contact references in one request; the record's `tag`/`untag`/`users`/`leads`
  fields map directly onto the documented body, no wrapper needed (unlike `create_user`/
  `create_lead`, `users`/`leads` here are already the correct array-of-refs shape the API expects,
  not a "wrap the whole record" pattern).
- `delete_tag` (`DELETE /tags`) — permanently removes a tag definition from the app (un-applies it
  from every contact that carried it); `body_fields: ["tag"]` restricts the body to the documented
  `{"tag": "..."}` shape.
- `send_message` (`POST /messages`) — sends a message from/to a user or admin on the caller's
  behalf; may notify the recipient.
- `add_user_attributes` (`POST /users/add-attributes/{user_id}`) / `add_company_attributes`
  (`POST /company/add-attributes/{company_id}`) — add/overwrite custom attribute values on a live
  user or company.
- `add_user_to_company` (`POST /users/add-to-company`) — links a user to a company (creating the
  company if it doesn't already exist); the API accepts any one of
  `internal_user_id`/`user_id`/`email` to identify the user, all declared as optional record
  properties.

`capabilities.write` is `true`; every action's `risk` field flags it as an external mutation
requiring approval (the two deletes explicitly "irreversible").

## Known limits

- **`max_pages` is not modeled.** Legacy exposes a config-driven `max_pages` hard request-count cap
  (`customerlyMaxPages`) independent of the short-page stop signal. The engine's `page_number`
  paginator has no `MaxPages`-equivalent config-driven knob (`PaginationSpec.MaxPages` is a fixed
  bundle-authored value, not wired to a runtime config key); pagination here is bounded only by the
  short-page stop signal, which is Customerly's own real termination behavior and is exercised on
  every read regardless of `max_pages`. `max_pages` is not declared in `spec.json` (a declared but
  unwireable key is worse than an absent one, per conventions.md F6).
- **`create_user`/`create_lead` are `ENGINE_GAP`, not implemented.** Both documented endpoints
  (`POST /v1/users`, `POST /v1/leads`) require the request body to wrap the record in a named
  bulk/singleton array — `{"users": [{...}]}` / `{"leads": [{...}]}` — and the response
  (`{"data": {"inserted": N, "errors": [...]}}`) confirms genuine bulk-array semantics, not a
  single-object convenience wrapper. The write dialect's body construction (`body_type`
  json/form/none, `body_fields` allow-list) has no primitive that nests a record inside a named
  array wrapper; `body_fields`/default body construction always produce a flat object of the
  record's own top-level keys. See `api_surface.json`'s `excluded` entries for both endpoints.
- **`GET /tags` (list all app tags) is `ENGINE_GAP`, not implemented as a stream.** Its response
  body is `{"data": ["marketing", "purchased", ...]}` — a bare JSON array of scalar strings, not
  objects. `connsdk.RecordsAt` only keeps array elements that decode as a JSON object (a scalar
  element is silently dropped, yielding zero records), and `records.keyed_object` explodes an
  object's VALUES (which must themselves be objects) rather than a flat array of strings. This is
  the same gap class as `ip2whois`'s `nameservers` deviation (conventions.md ledger item 12): no
  declarative primitive fans a bare scalar array into one record per element. See
  `api_surface.json`'s `excluded` entry.
- The two GET-by-identifier user/lead lookups (`GET /users?email=...`, `GET /users?user_id=...`,
  `GET /leads?email=...`) are excluded as `duplicate_of` the `users`/`leads` list streams: they
  fetch one record already fully covered by the paginated list, and the dialect has no per-record
  detail-fetch mechanism outside `fan_out` (which resolves a parent id LIST, not a caller-supplied
  single email/user_id at read time).
- Knowledge-base content management (collections/articles/writers — help-center authoring) is a
  distinct product surface from contact/CRM data and is out of scope for this connector; see
  `api_surface.json`'s `excluded` entries.
