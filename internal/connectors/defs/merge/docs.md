# Overview

Merge (https://merge.dev) is a unified API that normalizes data across many third-party platforms
into common-model objects. This bundle reads the Merge ATS (Applicant Tracking System) Common
Model v1 category: candidates, applications, jobs, offers, departments, and users. Migrated from
`internal/connectors/merge` (read-only there and here; `Capabilities.Write` is false).

## Auth setup

Merge uses dual-token auth. Provide an account-wide API token via the `api_token` secret (sent as
`Authorization: Bearer <api_token>`) and a per-linked-account token via the `account_token` secret
(sent as a static `X-Account-Token` header selecting which connected end-customer account to read).
Both are required — legacy hard-errors `Check`/`Read` when either is absent, matching this bundle's
`spec.json` `required: ["api_token", "account_token"]`. Neither is ever logged.

## Streams notes

All 6 streams (`candidates`, `applications`, `jobs`, `offers`, `departments`, `users`) share the
identical shape: `GET <resource>` against the Merge ATS base URL, records at `results`, primary key
`["id"]`, incremental cursor field `modified_at`. Pagination follows Merge's `next`/`previous`
cursor convention (`pagination.type: cursor` with `token_path: "next"`, `cursor_param: "cursor"`;
no `stop_path` needed since Merge's own stop signal is simply an absent/null `next` token, matching
the paginator's default stop-on-empty-token behavior with no separate boolean flag to gate on).
`page_size` is sent as the static literal `100` (legacy's `mergeDefaultPageSize`/`mergeMaxPageSize`
are both 100); `spec.json`'s `page_size` property is informational only (documents the same default
and legacy's accepted 1-100 range) rather than templated into the query, matching stripe's golden
`limit` pattern (`docs/migration/conventions.md` ledger item 3) — a config-templated page-size
param would receive conformance's synthetic non-numeric placeholder value during dynamic checks
instead of a real page size, so a static literal is used to guarantee both parity with legacy's
default behavior and safe conformance replay. Incremental reads send `modified_after` as an RFC3339
value (`param_format: rfc3339`), computed either from the sync's persisted cursor or, on a fresh
sync, from the `start_date` config value — identical to legacy's `incrementalLowerBound`.

## Write actions & risks

Not applicable — Merge is read-only in this bundle (`capabilities.write: false`, no `writes.json`),
matching legacy's `Write` returning an error wrapping `connectors.ErrUnsupportedOperation`
unconditionally ("merge connector is read-only").

## Known limits

- **`base_url` narrows to the ATS category only**: legacy's `base_url` config can point at any
  Merge product category (HRIS, Accounting, CRM, Ticketing, File Storage) by overriding the default
  ATS v1 base; this bundle's stream set (candidates/applications/jobs/offers/departments/users) is
  ATS-specific and would not resolve against another category's endpoints. `base_url` remains
  overridable (e.g. for a test server pointed at the identical ATS shape) but pointing it at a
  non-ATS Merge category is out of scope — a full multi-category migration is Pass B.
- **`max_pages` config dropped**: legacy accepts a runtime `max_pages` config value (0/`all`/
  `unlimited` for unbounded, else a positive integer hard cap). The engine's
  `PaginationSpec.MaxPages` is a fixed integer baked into the bundle, not a per-request templated
  value, so there is no mechanism to wire a runtime config value into it. Since legacy's own default
  is unbounded (`max_pages` unset) and declaring a fixed cap here would silently change accepted
  behavior for any caller relying on an unbounded sync, `max_pages` is left undeclared (no cap
  enforced, matching legacy's default) rather than kept as dead, unwireable config (F6,
  `docs/migration/conventions.md`).
- Only the 6 legacy-parity ATS Common Model streams are implemented; the wider Merge ATS surface
  (scorecards, interviews, EEOCs, tags) and any write actions are out of scope for this wave — see
  `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B capability expansion"}`
  entries.
- A 2-page fixture is provided for `candidates` (proving the cursor paginator consumes each page
  exactly once and terminates on a null `next`); the remaining 5 streams ship single-page fixtures,
  matching the stripe golden's precedent of providing one full 2-page proof per bundle rather than
  per stream.
