# Issue #4347 — run state

## Foundation state

Implemented and locally verified for the current-main GitHub lock. The branch
retains 14,471,636 raw bytes (12,920,264 REST + 1,551,372 GraphQL) under the
connector-owned tracked mirror; `source-import github --check` is hermetic.
It now includes `origin/main` through `72fe0ba88`; canonical source-import
regenerated the 18 resulting GitHub CLI projection changes, and the check is
green again.

## Blocking data requirement

The exact historical bytes for Elasticsearch and Zoom accounts are absent from
reachable Git objects. Firstmate authorized an explicit real-document re-pin;
GitHub GraphQL used that documented path and is retained. Elasticsearch is a
separate connector lane and must do the same in its own PR. Zoom accounts is an
HTTP-404 error response, not re-pin material, and is irrecoverable unless its
historic raw bytes turn up elsewhere. Firstmate inbox 004 prohibits importing
those lane files here; `LANE-ADOPTION.md` records the precise adoption work.

## Explicit unavailable state

An all-unavailable v3 source lock now reaches its declared terminal cause
without requiring a nonexistent retained-artifact manifest. This keeps Zoom
accounts honestly irrecoverable (HTTP 404/no verified historic copy) while
preserving mandatory retained copies for every actual source artifact.

## Post-rollup projection repair

Current-main schema generation renamed an optional parameter gap from `oneOf`
to non-scalar serialization. Projection now decides whether a zero-filter read
is executable from the source declaration's requiredness, restoring the 18
valid GitHub direct reads. The complete local 633-route fresh-binary fixture,
source inventory, and generator package suite are green.
