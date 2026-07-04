# Overview

Harvest is a declarative-HTTP migration over the Harvest v2 REST API
(`GET https://api.harvestapp.com/v2/...`). It preserves the legacy connector's five original
streams and expands the read surface to additional documented list/singleton resources: contacts,
company settings, task assignments, invoices, estimates, expenses, invoice and estimate item
categories, expense categories, and roles. The legacy package stays registered and unchanged until
wave6's registry flip.

## Auth setup

Provide a Harvest personal access token via the `api_token` secret and the Harvest account ID via
the `account_id` config value; both are required. The token is sent as a Bearer token
(`Authorization: Bearer <api_token>`, matching legacy's `connsdk.Bearer(token)`), and `account_id`
is sent as the required `Harvest-Account-Id` header on every request (matching legacy's
`DefaultHeaders`). Legacy also accepts a dotted `credentials.api_token` secret key (preferred) with
a flat `api_token` fallback, and reads `account_id` from secrets before config; this bundle
declares only the flat `api_token`/config `account_id` keys for a simpler, single spec surface —
see Known limits. `base_url` defaults to `https://api.harvestapp.com/v2` and may be overridden for
tests/proxies.

## Streams notes

Most streams are `GET` list endpoints whose records live at the top-level key matching the stream
name (`clients`/`projects`/etc.). `company` is a singleton endpoint modeled with
`records.single_object: true` and no pagination. Pagination follows Harvest's page-number-in-body
convention: the response's top-level `next_page` field is the next page NUMBER (or `null` when
exhausted) — modeled as
`pagination.type: cursor` with `token_path: "next_page"` and `cursor_param: "page"`, which reads
the token from the body and resends it as the `page` query param, matching legacy's `parsePage`
loop exactly (an absent/null/non-advancing `next_page` stops pagination). Primary key is `["id"]`
and incremental cursor field is `["updated_at"]` across list streams that expose `updated_at`.
Those streams send `updated_since={{ incremental.lower_bound }}` (the opt-in optional-query dialect,
`omit_when_absent: true`) — present with the RFC3339 lower bound (persisted sync cursor, or the
`start_date` config value on a fresh sync) and omitted entirely on a true full sync with no
lower bound at all, matching legacy's `incrementalLowerBound`/`harvest` exactly (legacy only sets
`updated_since` in its base query `if updatedSince != ""`).

`projects`, `contacts`, `invoices`, `estimates`, `expenses`, `task_assignments`, and
`time_entries` contain nested relationship objects. `computed_fields` promotes the stable related
IDs/names used by the flat schemas while preserving the raw nested object where the new stream
schema includes it.

## Write actions & risks

None. Harvest remains read-only in this connector (`capabilities.write: false`); this bundle ships
no `writes.json`, matching legacy's `Write` returning `connectors.ErrUnsupportedOperation`.
Documented POST/PATCH/DELETE endpoints create, mutate, archive, or delete live Harvest business
records and generally require administrator or manager permissions, so `api_surface.json` excludes
them with `requires_elevated_scope` or `destructive_admin` reasons instead of exposing broad writes.

## Known limits

- **The engine's `token_path` cursor loop guard differs slightly from legacy's stop condition —
  never incorrect, marginally more permissive.** Legacy's `harvest` loop stops when
  `next, ok := parsePage(nextPage); !ok || next <= page` — an explicit "next page number must
  strictly increase" check on every page. The engine's `tokenPathCursor` (used for
  `pagination.type: cursor` + `token_path`) stops on an absent/empty token (`next_page: null`,
  matching legacy's `!ok` branch) and additionally guards against looping by refusing to follow
  the exact same token TWICE IN A ROW (`engine/paginate.go`'s `tokenPathCursor.Next`), rather than
  legacy's stricter "next must be greater than current" check. For a well-behaved Harvest API
  (`next_page` is always either a strictly increasing page number or `null`), both stop at
  identical points; the engine's guard is only reachable in the pathological case of an API
  repeating a non-null `next_page` value, which would be a genuine Harvest API bug either side
  would need to defend against. No accepted-input record data differs between the two.
- **Alias config/secret keys are not modeled.** Legacy accepts `credentials.api_token` (a dotted
  secret key, checked first) as well as a flat `api_token` fallback, and reads `account_id` from
  secrets before config. This bundle declares only `api_token` (secret) and `account_id` (config)
  — the simpler, single-key spec surface — since a `spec.json` property with no template
  consuming it is dead config (conventions.md's query-templating note, F6, applied the same way to
  spec properties generally). Also not modeled: legacy's `replication_start_date` config alias for
  `start_date` (this bundle wires only `start_date`).
- **Legacy's fixture-mode-only `previous_cursor` echo field is not modeled.** Legacy's
  `readFixture` path (only reached when `config.mode == "fixture"`) stamps a `previous_cursor`
  field onto every fixture record when `req.State["cursor"]` happens to be set — a fixture-mode-only
  affordance, not part of the live record shape. This bundle's schemas and fixtures target the
  live record shape only; the engine's own conformance/fixture-replay harness provides the
  credential-free test affordance legacy's fixture mode was built for.
- Parameterized subresources such as invoice messages/payments, user rates, user project
  assignments, and project user assignments are not expressible as standalone declarative list
  streams without caller-supplied IDs or fan-out. They are documented in `api_surface.json`.
- Report endpoints require explicit date windows or emit aggregate metric rows with
  permission-sensitive amounts; they are excluded pending a dedicated report-stream contract.
