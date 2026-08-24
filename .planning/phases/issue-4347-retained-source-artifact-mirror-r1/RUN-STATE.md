# Issue #4347 — run state

## Foundation state

Implemented and locally verified for the current-main GitHub lock. The branch
retains 14,471,636 raw bytes (12,920,264 REST + 1,551,372 GraphQL) under the
connector-owned tracked mirror; `source-import github --check` is hermetic.

## Blocking data requirement

The exact historical bytes for Elasticsearch and Zoom accounts are absent from
reachable Git objects. Firstmate authorized an explicit real-document re-pin;
GitHub GraphQL used that documented path and is retained. Elasticsearch is a
separate connector lane and must do the same in its own PR. Zoom accounts is an
HTTP-404 error response, not re-pin material, and is irrecoverable unless its
historic raw bytes turn up elsewhere. Firstmate inbox 004 prohibits importing
those lane files here; `LANE-ADOPTION.md` records the precise adoption work.
