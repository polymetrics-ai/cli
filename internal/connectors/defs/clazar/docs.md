# Overview

Clazar is a cloud go-to-market platform (AWS/Azure/GCP marketplace co-sell, listings, private
offers, contracts). This bundle reads Clazar buyers, listings, contracts, opportunities, private
offers, reseller offers, contacts, and metering records, and writes buyer/opportunity/contract/
private-offer/contact/metering-record mutations plus contract activation, through the Clazar REST
API (`https://api.clazar.io`, reference at `https://developers.clazar.io/reference`) using OAuth2
client credentials. It began as a wave2 fan-out migration of `internal/connectors/clazar` (the
hand-written connector it migrates; the legacy package stays registered and unchanged until
wave6's registry flip) and was expanded to the full documented REST surface in a Pass B pass.
`capabilities.write` is now `true`.

## Auth setup

Provide Clazar OAuth2 client-credentials `client_id`/`client_secret` secrets (`mode:
oauth2_client_credentials`); the engine POSTs a `grant_type=client_credentials` token request to
`token_url` (`{{ config.base_url }}/authenticate/`, matching legacy's `base + "/authenticate/"`
and the real documented `POST /authenticate/` token endpoint) and applies the resulting bearer
token to every request, refreshing before expiry (`connsdk.OAuth2ClientCredentials`, the same
authenticator legacy itself uses). Neither secret is ever logged. `base_url` defaults to
`https://api.clazar.io` and may be overridden for tests/proxies (the override also reshapes the
derived token endpoint, matching legacy's `clazarBaseURL` + `clazarTokenPath` composition).

## Streams notes

The original 5 legacy-parity streams (`buyers`, `listings`, `contracts`, `opportunities`,
`private_offers`) are unchanged: `GET` against the Clazar list endpoint, `response_format=common`
sent on every request (legacy's own static query value — kept verbatim for these 5 streams even
though the live API does not document this parameter, since it is harmless and changing it is out
of scope for a parity-preserving pass), records at `results`, primary key `["id"]`, incremental
cursor field `last_modified_at`. Pagination is `page`/`page_size` (`pagination.type: page_number`,
`page_param: page`, `size_param: page_size`, `start_page: 1`, `page_size: 100` — legacy's own
default and max), matching legacy's `connsdk.PageNumberPaginator{PageParam: "page", SizeParam:
"page_size", StartPage: 1, PageSize: pageSize}` exactly (page-size stop threshold: a page
returning fewer than `page_size` records ends the stream). Incremental reads send
`last_modified_at_after` (`incremental.request_param`) computed either from the sync's persisted
cursor or, on a fresh sync, from the RFC3339 `start_date` config value — identical to legacy's
`incrementalLowerBound` (`cursor, then start_date, then absent`).

**New in this pass** (researched directly against `developers.clazar.io/reference`'s published
per-endpoint pages, 2026-07-04):

- **`reseller_offers`** (`GET /reseller_offers/`): same base pagination as the 5 original
  streams; primary key `["id"]`. No `last_modified_at` field is documented on this resource (its
  own timestamp fields are `accepted_at`/`published_at`/`expiration_at`, none of which is a
  monotonic "last changed" cursor), so no `incremental` block is declared — every sync is a full
  refresh, matching how the API itself exposes this resource (no `*_after`/`*_before` filter is
  documented for a generic "changed since" cut, only per-field date-range filters that are not the
  same semantic).
- **`contacts`** (`GET /contacts/`): same base pagination; primary key `["id"]`. No incremental
  cursor either — `contacts_list`'s documented query params (`email`/`full_name`/`is_editable`/
  `phone_number`/`uuid`/`search`/`ordering`/`page`/`page_size`) include no `updated_at_after`-style
  filter despite the resource itself carrying an `updated_at` field, so a request-param incremental
  cannot be expressed without risking silently dropping records the API's own filter semantics
  don't actually guarantee; full refresh only.
- **`metering`** (`GET /metering/`): same base pagination; primary key `["id"]`. No incremental
  cursor (metering records are typically append-only/immutable once accepted; the documented
  filters are `cloud`/`contract_id`/`dimension`/`status`, not a changed-since timestamp).

## Write actions & risks

10 write actions, added in this Pass B pass (Clazar's real REST surface is a compact,
fully-enumerable resource set — see `api_surface.json`):

- **`update_buyer`** / **`update_opportunity`** / **`update_private_offer`** / **`update_contract`**
  (`PATCH /{resource}/{{ record.id }}/`): all four resources document the identical update shape —
  only `custom_properties` (a free-form key-value map) and `external_object_associations`
  (Salesforce/HubSpot/Orb linkage) are mutable; every other field is read-only server-side. Low-risk
  (metadata/linkage only, no marketplace state change).
- **`activate_contract`** (`POST /contracts/{{ record.id }}/activate/`, `body_type: none`): no
  request body; transitions a contract from pending to active in the underlying cloud marketplace.
  **Approval required** — this is a real, hard-to-reverse state transition, not a metadata edit.
- **`create_contact`** / **`update_contact`** / **`delete_contact`**: full CRUD on Clazar's
  standalone Contact resource (`email`/`full_name`/`phone_number`); this is Clazar-internal data
  with no external marketplace side effect, so create/update are low-risk. `delete_contact` uses
  `delete.missing_ok_status: [404]` (idempotent delete) but is still marked approval-required since
  deletion is irreversible.
- **`update_metering_record`** (`PATCH /metering/{{ record.id }}/`): only `custom_properties` is
  documented as mutable on an already-submitted metering record; low-risk.
- **`create_metering_records`** (`POST /metering/`, `body_fields: ["request"]`): the real API wraps
  the submitted array under a `"request"` key (`{"request": [{cloud, contract_id, dimension,
  quantity, ...}, ...]}`), not a bare array or a resource-named key — `body_fields` restricts the
  JSON body to exactly that wrapper key. **Approval required**: submitted metering records drive
  usage-based marketplace billing/invoicing and are effectively irreversible once processed.

## Known limits

- **Dynamic (fixture-replay) conformance checks are marked `skip_dynamic` at the bundle level**
  (`metadata.json`'s `conformance` block). `oauth2_client_credentials` auth's `token_url` is
  derived from `{{ config.base_url }}/authenticate/`; conformance's `withReplayURL` only
  overrides `b.HTTP.URL` (the base request URL used for stream/check paths), never
  `RuntimeConfig.Config["base_url"]` itself, so the `token_url` template still resolves to the
  synthetic non-secret value (`"synthetic-conformance-value/authenticate/"`), an unreachable
  non-URL — the OAuth token exchange fails before any declarative stream/check request is ever
  issued, so every auth-resolving dynamic check (`check_fixture`, `read_fixture_nonempty:*`,
  `pagination_terminates`, `records_match_schema`, `cursor_advances`, and — per the bundle-level
  marker's documented scope — every `write_request_shape`/`delete_semantics` check too) would
  otherwise fail identically and uninformatively, for a reason that has nothing to do with this
  bundle's actual read/write shape. Static checks (`spec_schema_valid`, `stream_schemas_valid`,
  `interpolations_resolve`, `docs_present`, `fixtures_present`, `secret_redaction`, etc.) are
  unaffected and still run, including for the new streams/writes/fixtures added in this pass. This
  bundle has no Tier-2 `AuthHook` (auth is fully declarative `oauth2_client_credentials`), so there
  is no `paritytest/clazar` package; the read/pagination/incremental/schema/write-shape correctness
  of this pass is proven by structural review against `developers.clazar.io/reference`'s published
  per-endpoint pages (fetched 2026-07-04) instead. Matches sendpulse's/gmail's identical documented
  precedent for combining a bundle-level skip marker with a populated `writes.json`.
- `reseller_offers`/`contacts`/`metering` are full-refresh only (no incremental block) — see
  Streams notes above for why the documented query-param surface does not support a safe
  changed-since filter for any of the three.
- The single `PATCH /reseller_offers/{uuid}/` endpoint and the `GET /analytics/datasets/
  {dataset_name}` polymorphic analytics passthrough remain out of scope this pass; see
  `api_surface.json`'s `excluded` entries for the specific reasons (near-duplicate update shape,
  and no single fixed response schema across dataset_name values, respectively).
