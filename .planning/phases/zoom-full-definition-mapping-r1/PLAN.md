# Plan — Zoom full definition mapping

Issue: #4265 (parent mapping session)
Base: `origin/main` at `31bfe62eb`

## Task Delivery Header

- Issue: Refs #4265 — Zoom connector parity mapping session.
- Base branch: `main` at `31bfe62eb`.
- Merges into: `fm/cli-zoom-full-definition-mapping-r1` → `main` through the existing draft parent PR #4271; captain approval remains required for any main merge.
- Delivery: a committed, locally-verified connector-local Zoom definition and exhaustive disposition ledger; certification and parent merge are explicitly out of scope.
- Working branch: `fm/cli-zoom-full-definition-mapping-r1`.
- Task: pin public provider provenance, crosswalk Zoom's ledger, add only source-backed executable declarations, and record every other endpoint as disabled with evidence, a fixed-vocabulary reason, and recovery path.
- Verification: focused Zoom bundle tests, `connectorgen validate`, `surface-sync --check`, `connector-boundary`, the required non-Zoom certification regression, and `make verify` before pushing.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every declared endpoint is tied to the pinned provider source and Zoom ledger | live | The committed crosswalk counts every provider and ledger identity; validation rejects a declaration that lacks a matching ledger endpoint. |
| All non-delivered operations remain visible and actionable | live | The committed disposition ledger has one state, reason/evidence, and recoverability outcome per Zoom ledger endpoint. |
| No shared engine/auth/generator code is changed | live | `make connector-boundary` and the changed-path review show only Zoom definitions plus required GSD evidence. |

## Authoritative inventory

`api_surface.json` has 1,913 endpoints: five covered rows (the preserved three streams and two
source-backed warehouse actions) and 1,908 explicitly blocked operation rows. Operation models:
846 direct reads, 684 sensitive reverse-ETL writes, 312 destructive actions, 11 admin reverse-ETL
writes, one binary read, and 54 deprecated. Methods: 881 GET, 392 POST, 269 PATCH, 52 PUT, and 319
DELETE.

## Invariant

Every declaration must bind to a real API-surface row. The ledger provides method, origin-relative path, source URL, risk, status, and blocked reason, but does not contain request/response schemas, pagination, or parameter contracts. Those are not inferred from endpoint paths/titles; they require a pinned Zoom OpenAPI/source artifact in `sources/zoom-operation-source-lock.json` before operations, writes, streams, schemas, and fixtures can be honestly generated.

## Mapping stages

1. Pin and hash the Zoom provider artifact; compare it against all 1,913 ledger method/path rows.
2. Derive source operation facts and explicit blocked/rejection records from the ledger/source pairing. The lock remains immutable after its initial pin; the crosswalk is the contract-bearing derivative.
3. Derive only executable operation and write declarations that have a complete source contract, a real supported executor, typed schema, approval policy, and fixture proof. The two reviewed reverse-ETL actions are the only actual warehouse destination actions. DELETE records retain `mutation_class=delete` in the disposition ledger; do not invent 311 `writes.json` actions merely to mirror provider inventory.
4. Preserve the existing streams and derive only source-backed schemas/fixtures. Omit `sync_transport.json` until an executor identity and conformance run prove a closed source or destination transport; a placeholder is invalid.
5. Do not reconcile held credentials in this lane. No credential access or auth/engine change is permitted. Update metadata capabilities only for executable definitions.
6. Generate the required declared/blocked/per-class/foundation-gap report. Certification begins only after these surfaces validate.

## Source-lock result (2026-08-19)

The public Zoom Developer Docs catalog supplied 35 module-level OpenAPI documents through its
published Next data routes. The pinned documents identify themselves as OpenAPI 3.0.0; this is
recorded verbatim in the lock rather than being relabelled as the ledger's older 3.1.1 snapshot.
They contain 1,937 REST operations and 12,127,228 bytes in total.

The source's server URL is prepended before comparison: this is required for the ordinary
`https://api.zoom.us/v2` routes and the customer-managed-key `https://{keyConnectorLb}/api/v2`
route. That yields 1,911 exact matches across all 1,913 ledger identities: 1,908 blocked rows and
the three preserved stream-backed rows. The two unmatched ledger paths are the old Zoom Phone
`{callLogId}` path forms; the pinned source uses `{callElementId}`. They must remain
disabled/rejected with source evidence—parameter names are not silently rewritten. The source has
26 additional identities not present in the ledger; they remain disabled reconciliation records.

## Inline lifecycle fallback

The Pi adapter's lifecycle prompts were resolved and executed inline because this connector lane's
canonical single-worker ownership and its connector-local boundary prohibit spawning the plan's
proposed generator worker. The resulting implementation follows the same red/green ledger and
verification stages; no GSD role was skipped.

## Plan/source divergence

`DECLARATION-PLAN.md` proposes a new `cmd/connectorgen/zoom_parity.go` generator and treats the
source difference as 24 rows. Measured source-lock data instead show 26 source-only and two
ledger-only rows, and the lane is expressly limited to `internal/connectors/defs/zoom/`. This plan
therefore uses a committed connector-local crosswalk/disposition ledger rather than a new command,
and records the measured 26+2 split.
