# Overview

Apptivo is a wave2 fan-out declarative-HTTP migration, Pass-B surface-reviewed. It reads Apptivo
CRM customers, contacts, leads, and opportunities through the Apptivo REST DAO API
(`https://app.apptivo.com/app/dao/v6/<object>`), and deletes a customer record via the documented
`deleteCustomer` DAO action. This bundle originally targeted read-only capability parity with
`internal/connectors/apptivo` (the hand-written connector it migrates); Pass B
(`docs/migration/conventions.md` §8, `api_surface.json`) researched the full public per-object DAO
API reference for the 4 legacy-parity CRM objects (customers/contacts/leads/opportunities — see
`api_surface.json`'s scope note on why the ~30 other Apptivo app objects are out of scope for this
pass) and added the one action expressible in the current write dialect (`remove_customer`); every
create/update action across all 4 objects is a documented `ENGINE_GAP` (see Write actions & risks).
The legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Apptivo authenticates every request with two credentials sent as query parameters: `apiKey` and
`accessKey`. Provide both as the `api_key`/`access_key` secrets; `api_key` is wired via the
`api_key_query` auth mode (`apiKey={{ secrets.api_key }}`) and `access_key` is wired via a static
per-stream `query` entry (`accessKey={{ secrets.access_key }}`) — the engine's `stream.Query`
templates resolve against `secrets.*` exactly like `auth` does, so both credentials reach every
request the same way legacy's `apptivoAuth` closure (`apptivo.go:241`) sets both query params on
every call. Neither value is ever logged. `base_url` defaults to `https://app.apptivo.com` and may
be overridden for tests/proxies.

## Streams notes

All 4 streams (`customers`, `contacts`, `leads`, `opportunities`) share the same shape: `GET`
against the Apptivo DAO `getAll` action (`query: {"a": "getAll"}`), records at the `data` body key.
Pagination follows Apptivo's offset convention (`pagination.type: offset_limit`,
`limit_param: numRecords`, `offset_param: startIndex`, `page_size: 100`) — a page shorter than 100
records terminates the loop, matching legacy's `harvest` (`apptivo.go:139`) exactly.

## Write actions & risks

`capabilities.write` is now `true` (Pass B). Legacy's own `Write` unconditionally returned
`connectors.ErrUnsupportedOperation` — this is a genuinely new capability this bundle adds beyond
legacy parity, not a migrated legacy behavior, following `docs/migration/conventions.md`'s Pass B
full-surface-expansion mandate to implement every dialect-expressible mutation.

- `remove_customer` (`GET /app/dao/v6/customers?a=delete&customerId=...`) — irreversibly deletes a
  CRM customer record. `kind: "delete"`, `confirm: "destructive"`; approval required.
  Apptivo's documented `deleteCustomer` request (`https://www.apptivo.com/developer-api/
  customers-api-reference/deletecustomer/`) is a plain URL hit with `customerId`/`apiKey`/
  `accessKey` as query parameters and no HTTP method stated in the vendor docs (the sample URL's
  own shape — a GET-style query string, no request body — is the only concrete evidence available);
  this bundle uses `GET` to match that literal documented URL rather than guessing `POST`/`DELETE`.
  `path_fields: ["id"]` excludes `id` from the (empty, `body_type: none`) body since the customer id
  is carried entirely in the query string embedded in `path`, not as a path segment or body field —
  `{{ record.id }}` is still urlencoded by `InterpolatePath`'s default per-reference encoding
  exactly as any other path-embedded reference would be.

**Every create/update action for all 4 objects is an `ENGINE_GAP`, not implemented** — see
`api_surface.json`'s `createCustomer`/`updateCustomer` (and the equivalent contacts/leads/
opportunities actions') excluded entries for the full reasoning: Apptivo's real wire contract sends
the ENTIRE record as one JSON-serialized string inside a single query parameter
(`customerData={...}`), and `writes.json`'s `WriteAction` type has no `query` field at all (only
`path`/`path_fields`/`body_type`/`body_fields`), nor does any interpolation filter in this dialect
serialize an object to a JSON string for template embedding. Closing this would need either a
`query` field added to `WriteAction` plus a `to_json`-shaped filter, or a `WriteHook` — the latter a
disproportionate 3rd-hook-interface Tier-2 escalation for a single connector. `contacts`/`leads`/
`opportunities` additionally have no documented `delete` action at all (unlike customers), so no
delete action was added for those three objects either.

## Known limits

- **`startIndex=0` is sent explicitly on the first page request; legacy omits it entirely.**
  Legacy's `harvest` (`apptivo.go:148`) only sets `startIndex` in the query when `offset > 0`,
  mirroring the "upstream inject_on_first_request:false behaviour" its own comment describes — the
  API is documented to treat a missing `startIndex` as `0`. The engine's `offset_limit` paginator
  (`connsdk.OffsetPaginator.Start()`) always sends `offset_param` (here `startIndex=0`) on the
  first request, with no "omit on first page" option in the dialect. Per legacy's own comment this
  is accepted-behaviorally-identical by the API (a missing `startIndex` and an explicit `0` are the
  same request from Apptivo's point of view), so this never changes emitted record data for any
  input legacy itself would accept — an ACCEPTABLE parity deviation, not a data-changing one.
- `page_size`/`max_pages` config overrides legacy exposes (`apptivoPageSize`/`apptivoMaxPages`,
  clamped 1-500 / `all`/`unlimited`) are not runtime-configurable here: the engine's
  `offset_limit` paginator's `PageSize` is a static int set once in `streams.json`, not
  template-resolvable, and no config knob feeds a `MaxPages` cap for this paginator type.
  `spec.json` intentionally does not declare `page_size`/`max_pages` (a declared-but-unwireable key
  is worse than an absent one, per conventions.md F6).
