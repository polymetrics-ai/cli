# Overview

Intruder is a vulnerability-scanning platform. This bundle reads Intruder issues, scans, and
targets through the Intruder REST API (`https://api.intruder.io/v1`), read-only and full-refresh
only (the API exposes no incremental cursor for any of these objects). It migrates
`internal/connectors/intruder` (the legacy hand-written connector, kept registered and unchanged
until wave6's registry flip).

## Auth setup

Provide an Intruder API access token via the `access_token` secret; it is used only for Bearer
auth (`Authorization: Bearer <access_token>`) and is never logged.

## Streams notes

All 3 streams (`issues`, `scans`, `targets`) share the same shape: `GET` against the Intruder list
endpoint, records at `results`, primary key `["id"]`. Pagination is `offset_limit`
(`limit`/`offset` query params, `page_size: 100`, matching legacy's default `pageSize`/
`maxPageSize` of 100) — the engine requests successive pages until a page returns fewer records
than the page size. No stream has a usable incremental cursor field in the legacy connector, so
none is declared here either (full-refresh only, matching legacy exactly).

## Write actions & risks

None. Intruder is a read-only source connector; no `writes.json` is declared.

## Known limits

- **`occurrences` stream is NOT migrated in this wave.** Legacy's `occurrences` stream is a
  sub-resource fan-out: it first lists every issue id from `/issues`, then reads
  `/issues/{id}/occurrences` once per issue, tagging every occurrence with its parent `issue_id`.
  This "list parent ids, then fan out to a per-parent sub-resource" shape has no Tier-1
  declarative expression (`streams.json` has no directive to slice a list stream by another
  stream's ids and issue N follow-up requests) — per `docs/migration/conventions.md` §1/§6, this
  is a legitimate Tier-2 `StreamHook` trigger ("sub-resource fan-out reads"). Implementing it
  correctly (matching legacy's exact per-issue pagination and `issue_id` tagging) requires a
  `hooks/intruder/hooks.go` `StreamHook`, out of scope for this JSON-only wave. See
  `api_surface.json`'s `excluded` entry for `/issues/{id}/occurrences`.
- The 3 migrated streams (`issues`, `scans`, `targets`) are full capability parity with legacy's
  equivalent streams: same fields, same pagination behavior, same auth.
