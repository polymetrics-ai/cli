# Overview

Dwolla is a payments platform exposing a HAL+JSON REST API. This bundle reads Dwolla customers,
events, exchange partners, and business classifications, and writes customer/funding-source/
transfer/webhook-subscription/beneficial-owner lifecycle mutations, through `https://api.dwolla.com`
using OAuth2 client-credentials auth. It was originally migrated from `internal/connectors/dwolla`
(the hand-written connector this bundle replaces at capability parity; the legacy package stays
registered and unchanged until wave6's registry flip) and was expanded in Pass B to the full
documented JSON-body API surface (researched against developers.dwolla.com and the official Go/
Kotlin SDKs' request/response structs, since legacy itself shipped zero writes to migrate).

## Auth setup

Provide `client_id` and `client_secret` secrets; the engine's declarative
`oauth2_client_credentials` auth mode mints and caches a Bearer token via the OAuth2
client-credentials grant against `{{ config.base_url }}/token` (Dwolla hosts the token endpoint at
`/token` on the same host as the API, matching legacy's `dwollaTokenURL` derivation), refreshing
automatically before expiry — matching legacy's `connsdk.OAuth2ClientCredentials`. Both secrets
flow only into the token exchange and are never logged. Every request also sends
`Accept: application/vnd.dwolla.v1.hal+json` (Dwolla's HAL media type, matching legacy's
`dwollaHALAccept`).

Set `base_url` to `https://api-sandbox.dwolla.com` for Dwolla's sandbox environment, or leave unset
for production (`https://api.dwolla.com`, the default). See Known limits for the narrowed
`environment` config surface.

## Streams notes

Four top-level list streams, all sharing Dwolla's HAL pagination shape: records live under
`_embedded.<key>` and the next page is an absolute URL at `_links.next.href`
(`pagination.type: next_url`, `next_url_path: "_links.next.href"`), matching legacy's `harvest`
loop exactly. The initial request for every stream sends `limit=25` (legacy's
`dwollaDefaultPageSize`).

- `customers` (`GET /customers`, records at `_embedded.customers`).
- `events` (`GET /events`, records at `_embedded.events`).
- `exchange_partners` (`GET /exchange-partners`, records at `_embedded.exchange-partners` — the
  HAL embed key uses a hyphen, matching legacy's `embedKey: "exchange-partners"`; dotted-path
  extraction splits only on `.`, so the hyphenated segment is unaffected).
- `business_classifications` (`GET /business-classifications`, records at
  `_embedded.business-classifications`; reference data with no `created` timestamp, matching
  legacy's own catalog which omits `CursorFields` for this stream).

Every other stream's primary key is `["id"]` and incremental cursor field is `created`, matching
legacy's uniform catalog — but no stream declares an `incremental` block: Dwolla's list endpoints
accept no server-side `updated_since`-style filter parameter, matching legacy's own `InitialState`
(always an empty cursor; `start_date` config accepted by legacy's own comment but never actually
consumed as a request filter either).

## Write actions & risks

`capabilities.write: true` as of Pass B. Legacy was entirely read-only ("Dwolla is an upstream
source with no reverse-ETL surface"), but the full documented API also exposes a substantial
JSON-body-only lifecycle-mutation surface that does not require any file upload. Real money
movement (`POST /transfers`, `POST /mass-payments`) is deliberately **excluded**, not migrated —
see Known limits. 15 actions:

- `create_customer` / `update_customer` (`POST /customers`, `POST /customers/{id}`) — creates or
  updates a personal/business/receive-only/unverified customer; Dwolla's own identity-verification
  rules apply per `type`. `update_customer` can also deactivate/reactivate a customer via `status`.
- `create_funding_source` (`POST /customers/{id}/funding-sources`) — attaches a bank-account
  funding source, either unverified (`routingNumber`/`accountNumber`, needing later micro-deposit
  verification) or pre-verified (`plaidToken`/`onDemandAuthorizationId`).
- `update_funding_source` (`POST /funding-sources/{id}`) — renames it or replaces its unverified
  routing/account numbers.
- `remove_funding_source` (`POST /funding-sources/{id}` with `{"removed": true}`,
  `confirm: destructive`) — Dwolla has no hard delete for funding sources; this soft-removes it so
  it can no longer send/receive transfers. **Not reversible.**
- `initiate_micro_deposits` / `verify_micro_deposits` (`POST /funding-sources/{id}/
  {initiate,verify}-micro-deposits`) — the two-step micro-deposit bank-account verification flow;
  Dwolla locks the funding source after repeated failed verify attempts.
- `cancel_transfer` (`POST /transfers/{id}` with `{"status": "cancelled"}`,
  `confirm: destructive`) — cancels a still-**pending** transfer before it clears; only succeeds
  while pending, and is **not reversible**. Distinct from initiating a transfer, which remains
  excluded.
- `create_webhook_subscription` / `update_webhook_subscription` / `delete_webhook_subscription`
  (`POST /webhook-subscriptions`, `POST /webhook-subscriptions/{id}` with `{"paused": bool}`,
  `DELETE /webhook-subscriptions/{id}`) — registers, pauses/resumes, or permanently deletes
  (`confirm: destructive`, **not reversible**) a webhook subscription. Dwolla caps active
  subscriptions at 10 (Sandbox) / 5 (Production).
- `create_beneficial_owner` / `update_beneficial_owner` / `remove_beneficial_owner`
  (`POST /customers/{id}/beneficial-owners`, `POST /beneficial-owners/{id}`,
  `DELETE /beneficial-owners/{id}`) — manages a business-verified customer's 25%+ equity holders;
  create/update carry SSN/date-of-birth/address PII submitted to Dwolla for identity verification.
  `remove_beneficial_owner` is `confirm: destructive` and **not reversible**.
- `certify_beneficial_ownership` (`POST /customers/{id}/beneficial-ownership` with
  `{"status": "certified"}`) — the Account Admin's attestation that all beneficial-owner
  information is accurate; required before a business customer can transact.

Every action's `record_schema` requires only fields the real Dwolla API itself requires (verified
against `github.com/kolanos/dwolla-v2-go`'s request structs and developers.dwolla.com); path-only
`initiate_micro_deposits`/`delete_webhook_subscription`/`remove_beneficial_owner` use `body_type`
`"none"` since Dwolla accepts no body for these, and `remove_funding_source`/`cancel_transfer`/
`update_webhook_subscription`/`certify_beneficial_ownership` use `body_fields` to send only the
fixed status-transition field Dwolla expects, not an arbitrary record dump.

## Known limits

- **Real money movement is out of scope, deliberately, not a gap**: `POST /transfers` (initiate a
  transfer) and `POST /mass-payments` (initiate up to 5,000 transfers in one batch) are excluded as
  `destructive_admin` — see `api_surface.json`. `cancel_transfer` is covered because cancelling a
  still-pending transfer prevents money from moving, the opposite risk profile.
- **Multipart/file-upload endpoints are out of scope** — the engine's write dialect supports only
  `json`/`form`(url-encoded)/`none` body types, never `multipart/form-data`. Identity-verification
  document upload (`POST /customers/{id}/documents`, `POST /beneficial-owners/{id}/documents`)
  requires an actual file payload and cannot be expressed.
- The **Dwolla Balance-tier Labels feature** (`/customers/{id}/labels`, a fund-reservation ledger)
  is a distinct, plan-gated sub-product legacy never touched; excluded as `requires_elevated_scope`.
- On-demand-authorizations, IAV tokens, funding-source tokens, and KBA question/answer sessions are
  all short-lived session credentials or identity-verification-flow steps tightly coupled to a
  specific customer-onboarding sequence, not independent syncable records or mutations; excluded as
  `non_data_endpoint`.
- **Dynamic (fixture-replay) conformance checks are marked `skip_dynamic` at the bundle level**
  (`metadata.json`'s `conformance` block), covering both reads and the Pass B write actions.
  `oauth2_client_credentials` auth's `token_url` is derived from `{{ config.base_url }}/token`;
  conformance's `withReplayURL` only overrides `b.HTTP.URL` (the base request URL used for
  stream/check/write paths), never `RuntimeConfig.Config["base_url"]` itself, so the `token_url`
  template still resolves to the synthetic non-secret value
  (`"synthetic-conformance-value/token"`), an unreachable non-URL — the OAuth token exchange fails
  before any declarative stream/check/write request is ever issued, so every auth-resolving dynamic
  check (including `write_request_shape`/`delete_semantics`) would otherwise fail identically and
  uninformatively. Static checks (spec/schema validity, `interpolations_resolve`, docs/fixtures
  presence, secret redaction, `surface_complete`, `write_schemas_valid`) still run and pass. This
  bundle has no Tier-2 `AuthHook` (auth is fully declarative `oauth2_client_credentials`), so there
  is no `paritytest/dwolla` package for this wave; the read/pagination/schema-projection shape and
  every Pass B write action's method/path/body construction are proven by structural review against
  the real Dwolla API (`github.com/kolanos/dwolla-v2-go`'s request/response structs,
  developers.dwolla.com) instead of a legacy Go write path to compare against (legacy shipped zero
  writes). Matches `clazar`'s and `sendpulse`'s identical documented precedent for the read side,
  and `gmail`'s precedent for combining a bundle-level skip marker with a non-empty `writes.json`.
- Legacy accepted an `environment` config value (`api`/`api-sandbox`) as an alternative to an
  explicit `base_url` override, deriving the production/sandbox host from it. The engine's
  `spec.json` `"default"` materialization only supports a fixed literal default, not one derived
  from another config value at read/check time — the same limitation `docs/migration/
  conventions.md` documents for sentry's `hostname`-derived base URL. This bundle narrows the
  config surface to `base_url` only (defaulting to production); a caller who needs the sandbox
  host sets `base_url` to `https://api-sandbox.dwolla.com` directly instead of `environment:
  api-sandbox`. This narrows accepted CONFIGURATION surface only, never emitted record data.
- `funding-sources`/`beneficial-owners`/`transfers` per-customer-scoped list READS remain out of
  scope (they require a per-customer partition the top-level list streams do not need); the
  corresponding WRITE actions above are still fully covered since a write's target id comes from
  the caller's own record, not from this connector's own read side. See `api_surface.json`'s
  `excluded` entries for the complete accounting.
- `metadata.json` declares no `rate_limit` block: legacy enforces no client-side rate limiting for
  Dwolla, so none is added here either (matches legacy's real, lack-of, throttling behavior).
