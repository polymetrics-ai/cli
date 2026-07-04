# Overview

WorkflowMax reads and writes jobs, clients, and client contacts through the real WorkflowMax
API v2 (`https://api.workflowmax2.com/v2/...`, documented at
[api-docs.workflowmax.com](https://api-docs.workflowmax.com/)). This bundle was originally
migrated from `internal/connectors/workflowmax` (the hand-written connector); this Pass B pass
researched the real, currently-documented v2 API (WorkflowMax's own docs note the legacy XML v1
API has been superseded by a JSON v2 API) and corrected the bundle's paths/auth/query surface to
match it, while keeping the same 3 legacy-parity resource families (jobs, clients, contacts) and
adding writes. `capabilities.write` is now `true`.

## Auth setup

Provide a WorkflowMax API v2 OAuth2 access token via the `access_token` secret; it is sent as a
Bearer token on every request (`mode: bearer`). WorkflowMax v2 additionally **requires** an
`account-id` header on every request identifying the Xero/WorkflowMax organisation — provide it
via the required `account_id` config property (`streams.json`'s `base.headers` sends
`account-id: {{ config.account_id }}` on every read AND write request, matching Stripe's
`Stripe-Account`-header golden pattern in `docs/migration/conventions.md` §3). `base_url` defaults
to `https://api.workflowmax2.com` (matching legacy's own `defaultBaseURL`, which correctly names
the real v2 host) and may be overridden for test proxies.

## Streams notes

`jobs` (`GET /v2/jobs`) and `clients` (`GET /v2/clients`) share the identical envelope (records at
the top-level `data` array, matching legacy's `recordsPath: "data"`) and `page_number` pagination
(`page`/`pageSize` query params — the real v2 parameter name is `pageSize`, not legacy's
`page_size`). The bundle sends a static `pageSize=100` and caps reads at one page, matching
legacy's own default `page_size=100` and unset-`max_pages` behavior. Legacy's optional runtime
`page_size`/`max_pages` overrides are not declared because the engine pagination block is static.
Both streams declare an optional `updatedSince` query param
(`{{ config.updated_since }}`, `omit_when_absent: true`) wired to the real v2 `updatedSince`
date-range filter documented for both endpoints — this is a genuinely new, real server-side filter
neither legacy nor the pre-Pass-B bundle modeled; it is deliberately NOT wired as an `incremental`
block (no state-cursor-driven auto-advance), since legacy itself never issued any server-side
filter and adding automatic cursor-driven filtering here would be new sync-mode behavior beyond
Pass B's full-surface-coverage scope, not a parity port. Leaving `updated_since` unset (the
default) reproduces legacy's exact always-full-read behavior.

`jobs` ships a single default-read conformance fixture page at the real `pageSize: 100`; the
static `max_pages: 1` cap stops there to match legacy's default read. `clients` ships a single
fixture page (2 records, an honestly short final page under a 100 page size).

`jobs` emits the real v2 field set: `uuid` (primary key), `jobNumber`, `name`, `clientUUID`,
`clientContactUUID`, `description`, `clientOrderNumber`, `budget`, `jobStatusUUID`,
`jobCategoryUUID`, `priority`, `startDate`, `dueDate`, `completedDate`. `clients` emits `uuid`
(primary key), `name`, `exportCode`, `clientManagerUUID`, `jobManagerUUID`, `referralSource`,
`prospect`, `archived`, `favorite`. Both streams declare `"projection": "passthrough"` (matching
legacy's verbatim-emit behavior and the post-wave2 §8 rule 1) so every real v2 field survives, not
just the ones enumerated in each schema. Both schemas declare `x-cursor-field: updated_at` for
legacy catalog parity (`CursorFields: []string{"updated_at"}`), but neither stream declares an
`incremental` block: the real v2 list responses do not surface a per-item `updatedAt` cursor field
in the base (no `includes=`) response shape this bundle reads, so `updatedSince` remains a plain
optional config filter (above) rather than a stateful cursor.

**`contacts` is no longer a stream.** The pre-Pass-B bundle declared a `GET /contacts` stream, but
WorkflowMax v2 has **no bare list endpoint for client contacts at all** — verified against
api-docs.workflowmax.com's full `client-contact` endpoint set, which is exactly `POST
/v2/clients/contacts` (create), `GET /v2/clients/contacts/{uuid}` (get by id), `PUT
/v2/clients/contacts/{uuid}` (update), and `DELETE /v2/clients/contacts/{uuid}` (delete) — never a
collection GET. `legacy`'s `GET /contacts` path does not correspond to any endpoint in the real,
currently-documented v2 API (nor the deprecated v1 XML API, which this bundle does not target).
Client contacts are exposed through the `create_client_contact`/`update_client_contact`/
`delete_client_contact` write actions instead; per-client contact data is also visible nested
inside a `clients` stream record's real API response when the (not-yet-modeled) `includes=contacts`
query is added in a future pass. See Known limits.

## Write actions & risks

`capabilities.write` is now `true` (previously `false`). Eight actions, all requiring the same
Bearer + `account-id` header pair as reads:

- `create_client` (`POST /v2/clients`) / `update_client` (`PUT /v2/clients/{uuid}`) /
  `delete_client` (`DELETE /v2/clients/{uuid}`) — creates, updates, or permanently deletes a
  WorkflowMax client record.
- `create_job` (`POST /v2/jobs`) / `delete_job` (`DELETE /v2/jobs/{identifier}`) — creates or
  permanently deletes a WorkflowMax job. `update_job` (`PUT /v2/jobs/{identifier}`) is
  **excluded** (see `api_surface.json`): its real request body accepts a large partial-update
  surface (status/category/priority/staff-assignment/custom-field transitions) not modeled by this
  pass.
- `create_client_contact` (`POST /v2/clients/contacts`) / `update_client_contact` (`PUT
  /v2/clients/contacts/{uuid}`) / `delete_client_contact` (`DELETE /v2/clients/contacts/{uuid}`).

All eight risk-annotated as external mutations requiring approval (`delete_*` further flagged as
permanent deletes). `delete_client`/`delete_job`/`delete_client_contact` use `body_type: "none"`
(pure path-parameterized DELETE, no body); the rest use `body_type: "json"`.

## Known limits

- **Legacy's paths did not correspond to the real API.** Legacy's `GET /jobs`, `GET /clients`, and
  `GET /contacts` (no `/v2` prefix, no `account-id` header) do not match any endpoint in
  WorkflowMax's real, currently-documented v2 API — the real v2 surface requires `/v2/<resource>`
  paths and a mandatory `account-id` header, and has no bare `GET /contacts` at all. This Pass B
  pass corrects the bundle to the real, working v2 surface (documented above) rather than
  preserving legacy's non-functional paths; this is a genuine bug-fix, not a parity-narrowing
  deviation, since legacy's original paths would 404/401 against the real live API.
- **`page_size`/`max_pages` config-driven per-request overrides are not modeled.** The engine's
  `page_number` paginator reads `page_size`/`max_pages` from the static `streams.json`
  pagination block only, so the bundle uses legacy's default `pageSize=100` and one-page cap and
  does not declare ignored runtime config keys for those overrides.
- **`updated_since` is a plain optional filter, not a stateful `incremental` block** — see Streams
  notes above for the reasoning; a future pass could add a real `incremental` block once a stable
  per-item cursor field is confirmed in the live response shape.
- **Client contacts have no list/sync stream** (see Streams notes above) — only get-by-id (used
  internally to validate updates) and the three write actions are covered; a bare contacts list
  simply does not exist in the real API.
- The remaining real WorkflowMax v2 surface (staff, billable tasks, costs, custom fields/rates,
  leads, quotes, quote-variations, invoices, payments, purchase orders, suppliers/supplier-contacts,
  timesheets, capacity-plan, and all binary document-upload endpoints) is out of scope for this
  migration; see `api_surface.json`'s `excluded` entries for the full per-endpoint accounting.
