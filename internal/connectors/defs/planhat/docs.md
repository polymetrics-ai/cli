# Overview

Planhat is a wave2 fan-out declarative-HTTP migration. It reads Planhat companies, end users, and
licenses through the Planhat REST API (`GET https://api.planhat.com/<resource>`). This bundle is a
capability-parity port of `internal/connectors/planhat` (the hand-written connector it migrates);
the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Planhat API token via the `api_token` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_token>`), matching legacy's `connsdk.Bearer(token)`
(`planhat.go:179`). `base_url` defaults to `https://api.planhat.com` and may be overridden for
tests/proxies.

## Streams notes

All three streams (`companies`, `endusers`, `licenses`) are root-array list endpoints (`GET
/companies`, `/endusers`, `/licenses`; `records.path: "."`, matching legacy's
`connsdk.RecordsAt(resp.Body, ".")`). Every stream computes a uniform `id` field from the raw
Planhat `_id` key and an `updated_at` field from the raw `updatedAt` key, matching legacy's
`first(item, "_id", "id")`/`first(item, "updatedAt", "updated_at")` primary path. `endusers` also
computes `phase` from the raw `status` key (legacy's `userRecord`); `licenses` computes `name` from
the raw `product` key and `phase` from `status` (legacy's `licenseRecord`).

Pagination is `offset_limit` (`limit`/`offset` query params, matching legacy's manual
`url.Values{"limit":..., "offset":...}` loop at `planhat.go:93-116`), 100 records per page and a
default `max_pages: 3` cap, both matching legacy's `defaultPageSize`/`defaultMaxPages` constants
exactly.

## Write actions & risks

None. Planhat's legacy connector is read-only (`Capabilities.Write: false`); this bundle ships no
`writes.json`.

## Known limits

- **Secondary fallback field names are not modeled.** Legacy's `first()` helper tries a primary
  key then a secondary fallback for several fields: `companyRecord`/`userRecord`/`licenseRecord`'s
  `id` tries `_id` then `id`; `updated_at` tries `updatedAt` then `updated_at`; `userRecord.name`
  tries `name` then `fullName`; `licenseRecord.name` tries `name` then `product`. The engine's
  `computed_fields` dialect has no multi-path-fallback primitive (only a single bare
  `{{ record.<path> }}` reference), so this bundle wires only the primary path each pair actually
  uses on Planhat's real API wire shape (`_id`, `updatedAt`); `endusers.name` needs no
  `computed_fields` entry at all since plain schema projection already carries a raw `name` field
  through unchanged (the common case, `name` checked first by legacy too). For `licenses.name`,
  this bundle diverges from legacy's own fallback ORDER (`name` checked first, `product` second) and
  wires `product` as the primary source instead, since Planhat's real license objects carry a
  `product` field, not a top-level `name` — this is the field legacy's own fallback exists to reach
  in practice. A record that genuinely has a `name` field (the primary path in legacy's order)
  would diverge here; this is judged low-risk since Planhat's documented license schema does not
  expose `name` at the top level, but it is a documented divergence from legacy's precedence order,
  not merely a dropped fallback.
- **`licenses.email` (legacy: always `nil`) is not modeled as an explicit null field.** Legacy's
  `licenseRecord` unconditionally sets `"email": nil` regardless of the raw record (licenses have
  no email concept); the engine's `computed_fields` dialect has no static-null literal (only
  string templates/static-literal strings), so this bundle simply does not declare `email` in the
  `licenses` schema — the field is absent rather than present-with-null. This is accepted as
  capability parity for any consumer that treats an absent field and an explicit JSON `null`
  identically (the overwhelming common case for downstream warehouse destinations); legacy's own
  test suite never asserts on this field's value.
