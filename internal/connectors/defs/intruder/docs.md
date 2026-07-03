# Overview

Intruder is a vulnerability-scanning platform. This bundle reads Intruder issues, issue
occurrences, scans, and targets through the Intruder REST API (`https://api.intruder.io/v1`),
read-only and full-refresh only (the API exposes no incremental cursor for any of these objects).
It migrates `internal/connectors/intruder` (the legacy hand-written connector, kept registered and
unchanged until wave6's registry flip). **Status: complete** — all 4 legacy streams are now
migrated at Tier 1, including `occurrences` via the S4 engine mini-wave's `stream.fan_out`
primitive (see Streams notes).

## Auth setup

Provide an Intruder API access token via the `access_token` secret; it is used only for Bearer
auth (`Authorization: Bearer <access_token>`) and is never logged.

## Streams notes

The 3 top-level streams (`issues`, `scans`, `targets`) share the same shape: `GET` against the
Intruder list endpoint, records at `results`, primary key `["id"]`. Pagination is `offset_limit`
(`limit`/`offset` query params, `page_size: 100`, matching legacy's default `pageSize`/
`maxPageSize` of 100) — the engine requests successive pages until a page returns fewer records
than the page size. No stream has a usable incremental cursor field in the legacy connector, so
none is declared here either (full-refresh only, matching legacy exactly).

`occurrences` is a sub-resource fan-out, now expressible via the S4 engine mini-wave's
`stream.fan_out` primitive (`docs/migration/conventions.md` §3): `fan_out.ids_from.request` issues
a preliminary `GET /issues` (paginated to exhaustion with the stream's own `offset_limit` spec,
matching legacy's `collectIssueIDs`), extracting `id` off each record found at `results`
(`records_path`); each resolved id is then threaded into `occurrences`'s own `path` as `{{
fanout.id }}` (`into.path_var`, resolving to `GET /issues/{id}/occurrences`) and stamped onto every
emitted record's `issue_id` field (`fan_out.stamp_field`), matching legacy's `readOccurrences` +
`wrap`'s `if rec["issue_id"] == nil { rec["issue_id"] = issueID }` injection — Intruder's real
`/issues/{id}/occurrences` response never carries a native `issue_id` field on its rows, so the
fallback branch fires unconditionally in practice, and the fan-out's unconditional stamp
reproduces that outcome exactly. Pagination, `MaxPages`, and rate-limiting are independent per
issue id, matching legacy's per-id `harvest` call.

## Write actions & risks

None. Intruder is a read-only source connector; no `writes.json` is declared.

## Known limits

- **`occurrences`'s `issue_id` is typed as `["string", "integer", "null"]`, not legacy's bare
  `integer` (ACCEPTABLE, documented deviation).** `fan_out.stamp_field` always writes the fan-out
  id as the STRING extracted from the id-listing request (matching every other
  `stamp_field`-using bundle in this repo — see `docs/migration/conventions.md` §3, and
  appfollow's `app_collection_id`/cisco-meraki's `organizationId` for the same shape), whereas
  legacy's raw API response and `intruderOccurrenceRecord` mapping would carry an integer had the
  API ever populated it natively. The emitted VALUE is identical (the same issue id, formatted as
  its decimal string form); only the JSON type differs. The schema widens to accept both shapes
  rather than silently narrowing legacy's declared type.
- All 4 streams (`issues`, `occurrences`, `scans`, `targets`) are full capability parity with
  legacy's equivalent streams: same fields, same pagination behavior, same auth.
