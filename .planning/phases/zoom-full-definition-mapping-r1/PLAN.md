# Plan — Zoom full definition mapping

Issue: #4265 (parent mapping session)
Base: `origin/main` at `acb85dc03`

## Task Delivery Header

- Issue: Refs #4265 — Zoom connector parity mapping session.
- Base branch: `main` at `362d7ccf9`.
- Merges into: `fm/cli-zoom-full-definition-mapping-r1` → `main` through the existing draft parent PR #4285; captain approval remains required for any main merge.
- Delivery: a committed, locally-verified Zoom command surface, typed write actions, source contracts, source transport declaration, candidate declaration, and exhaustive disposition ledger. Certification proof is imported only for cells that have both an exact fixture and passing live proof; parent merge remains explicitly human-gated.
- Working branch: `fm/cli-zoom-full-definition-mapping-r1`.
- Task: pin public provider provenance, crosswalk Zoom's ledger, bind every executable source-backed contract to a runnable command and typed action where applicable, and record every other endpoint as disabled with evidence, a fixed-vocabulary reason, and recovery path.
- Verification: focused Zoom bundle tests, `connectorgen validate`, `surface-sync --check`, `connector-boundary`, the required non-Zoom certification regression, and `make verify` before pushing.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every declared endpoint is tied to the pinned provider source and Zoom ledger | live | The committed crosswalk counts every provider and ledger identity; validation rejects a declaration that lacks a matching ledger endpoint. |
| Every executable contract is user-reachable | local | `cli_surface.json` has 712 implemented commands bound to exact source contracts and `api_surface` rows; `writes.json` has 204 typed actions including 185 guarded deletes. |
| All non-delivered operations remain visible and actionable | local | The committed disposition ledger has one state, reason/evidence, and recoverability outcome per Zoom ledger endpoint. |
| No shared engine/auth/generator code is changed | live | `make connector-boundary` and the changed-path review show only Zoom definitions plus required GSD evidence. |

## Authoritative inventory

`api_surface.json` has 1,913 endpoints: 712 covered rows on this branch (the preserved three
streams, two source-backed warehouse actions, 505 direct reads, and 202 newly typed actions) and
1,131 explicitly blocked operation rows. Operation models:
846 direct reads, 684 sensitive reverse-ETL writes, 312 destructive actions, 11 admin reverse-ETL
writes, one binary read, and 54 deprecated. Methods: 881 GET, 392 POST, 269 PATCH, 52 PUT, and 319
DELETE.

## Invariant

Every declaration must bind to a real API-surface row. The ledger provides method, origin-relative path, source URL, risk, status, and blocked reason, but does not contain request/response schemas, pagination, or parameter contracts. Those are not inferred from endpoint paths/titles; they require a pinned Zoom OpenAPI/source artifact in `sources/zoom-operation-source-lock.json` before operations, writes, streams, schemas, and fixtures can be honestly generated.

## Mapping stages

1. Pin and hash the Zoom provider artifact; compare it against all 1,913 ledger method/path rows.
2. Derive source operation facts and explicit blocked/rejection records from the ledger/source pairing. The lock remains immutable after its initial pin; the crosswalk is the contract-bearing derivative.
3. Derive every executable operation as a command and every no-body scalar mutation as a typed
   write action with an exact source contract and approval policy. DELETE records retain
   `mutation_class=delete` and destructive confirmation. Hold body-schema, array-encoding,
   binary, upload, paid-tier, and source-mismatch cases in the disposition ledger; do not infer a
   substitute contract.
4. Preserve the existing streams and derive only source-backed schemas/fixtures. Declare `sync_transport.json` only for the merged connector-neutral declarative source adapter; do not invent a reverse-ETL destination binding while the factory remains GitHub issue-label-specific.
5. Generate a bounded direct-read candidate and all typed mutation candidate inventory. Read held OAuth values only at point of use, exchange them in memory, and import only fingerprint-only external proof.
6. Generate the required declared/blocked/per-class/foundation-gap report. Record a live operation as uncertified when the matrix lacks its exact fixture projection.

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

## Corrected runnable-command delivery (2026-08-19)

The prior mapping completion condition is superseded by
`data/PARITY-DELIVERABLE-CORRECTION.md`: an operation inventory is supporting evidence, not a user
capability. This phase now delivers every Zoom operation that has an executable operation contract
as an `availability: implemented` command in `cli_surface.json`, and it exposes each eligible
mutation through the guarded direct-write plan lifecycle. `ENABLED%` is calculated from runnable
commands, never from the operation inventory.

The implementation passes proceed in dependency order:

1. Add a red connector-local test that fails while Zoom exposes only five commands and while an
   `requires-elevated-scope` disposition is disabled. Green requires deterministic command coverage
   for every eligible `rest_read` and no ordinary DELETE classified `unsafe-to-exercise`.
2. Materialize source-backed direct-read commands with the declaration-owned path/query flags and
   output policy. Use the parameter-import paging exclusion so opaque cursors remain reachable only
   through `--page`/`--page-cursor`; do not hand-author them or infer pagination.
3. Materialize guarded reverse-ETL commands for source-backed no-body scalar DELETEs and other
   writes with a complete declared request contract. Every DELETE keeps `mutation_class: delete`, an exact
   `api_surface` binding, and plan/preview/approval/execute protection.
4. Reclassify `requires-elevated-scope` rows as enabled: required scopes are command metadata and
   a missing permission is a provider 403 at runtime, not a compile-time disable. Reclassify
   ordinary deletion from `unsafe-to-exercise`; reserve that reason for credential-minting or an
   equivalently dangerous operation. Continue to disable only paid-tier, foundation-gap,
   schema-incompatible, provider-not-exposed, and genuinely unsafe operations, each with evidence.
5. Do not manufacture a root JSON write payload. The current operation records do not retain a
   per-operation request-body schema and `operationDirectReadOverrides` only permits a root `body`
   mapping for `direct_read`. If that makes a write non-executable, record the exact engine gap and
   minimal external foundation change; do not alter the engine in this connector lane.

Each pass gets focused no-credential preflight evidence and the command-count/rejection evidence is
kept in the TDD ledger and verification record before the required full `make verify` gate.

## Required skills

Loaded for this connector/CLI/test/documentation work: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`,
`golang-structs-interfaces`, and `golang-documentation`. The work remains connector-local; no shared
engine, auth, generator, certification allowlist, or status code is modified.
