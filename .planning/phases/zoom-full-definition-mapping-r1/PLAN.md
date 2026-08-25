# Plan — Zoom full definition mapping

Issue: #4265 (parent mapping session)
Base: `origin/main` at `acb85dc03`

## Task Delivery Header

- Issue: Refs #4265 — Zoom connector parity mapping session.
- Base branch: `fm/cli-reverse-etl-destination-r1` at `d814875a902be684cb2a38b94f7a8077f66b70b1`.
- Merges into: `fm/cli-zoom-full-definition-mapping-r1` → `fm/cli-reverse-etl-destination-r1` → `main` through the existing draft PR #4285; captain approval remains required for any main merge.
- Delivery: a committed, locally-verified Zoom command surface, typed write actions, source and destination transport declarations, candidate declaration, and exhaustive seven-surface readiness ledger. Certification proof is imported only for cells that have both an exact fixture and passing live proof; PR #4285 remains draft and parent merge remains explicitly human-gated.
- Working branch: `fm/cli-zoom-full-definition-mapping-r1`.
- Task: pin public provider provenance, crosswalk Zoom's ledger, bind every executable source-backed contract to a runnable command and typed action where applicable, declare the production typed reverse-ETL destination selected by #4304, and record every other endpoint as disabled only with a precise technical recovery path.
- Verification: focused Zoom bundle tests, the targeted CI-timeout regression, `connectorgen validate`, generated certification artifacts, `surface-sync --check`, `connector-boundary`, the required non-Zoom certification regression, and `make verify` before pushing.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every declared endpoint is tied to the pinned provider source and Zoom ledger | live | The committed crosswalk counts every provider and ledger identity; validation rejects a declaration that lacks a matching ledger endpoint. |
| Every executable contract is user-reachable | local | `cli_surface.json` has 714 implemented commands bound to exact source contracts and `api_surface` rows; `writes.json` has 206 typed actions including 185 guarded deletes. |
| All non-delivered operations remain visible and actionable | local | The committed disposition ledger has one state, reason/evidence, and recoverability outcome per Zoom ledger endpoint. |
| No shared engine/auth/generator code is changed | live | `make connector-boundary` and the changed-path review show only Zoom definitions plus required GSD evidence. |

## Authoritative inventory

`api_surface.json` has 1,913 endpoints: 714 covered rows on this branch (the preserved three
streams, two source-backed warehouse actions, 505 direct reads, and 204 typed actions) and
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
4. Preserve the existing streams and derive only source-backed schemas/fixtures. The initial checkpoint declared only the source adapter; the continuation adds a closed, connector-owned typed destination after #4304, never a generic writer.
5. Generate a bounded direct-read candidate and all typed mutation candidate inventory. Read held OAuth values only at point of use, exchange them in memory, and import only fingerprint-only external proof.
6. Generate the required declared/blocked/per-class/foundation-gap report. Record a live operation as uncertified when the matrix lacks its exact fixture projection.

### Captain-required missing-foundation ledger continuation

Before a foundation gap can be treated as a dependency, it must have a stable shared gap ID and
an operation-level, source-locked fan-out. The connector-owned
`sources/zoom-missing-foundation-gaps.json` therefore keeps one deduplicated gap catalog and one
row for each affected Zoom provider operation. Each row joins the catalog to preserve the exact
provider source URL, Next-data revision, document hash, operation identity, affected surface(s),
runtime/validator evidence, owner, status, and closure verification. It deliberately records
`merge_ready_enabled=false` for an open foundation gap while keeping any independently implemented
CLI command's reachability fact separate; a foundation gap must never be relabelled as disabled or
not-applicable to improve a count. Rollups are by source module batch and the one-connector
portfolio, with each catalog item listing its complete Zoom operation fan-out. No shared or
connector-specific workaround is authorized by this evidence work.

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

## Historical pre-#4304 merge-freeze readiness (2026-08-19)

Issue #4303 is the estate-wide merge prerequisite. Its connector-neutral typed destination factory
must select only a connector-owned named action, explicit source bindings, acknowledgement, per-mode
apply strategies, and connector-owned conformance evidence. Zoom must not predeclare a
`destination_transport` or invent a `transport_binding` before that schema and executor land.

The pre-#4303 connector-local readiness contract is therefore exhaustive set equality: every
implemented Zoom `reverse_etl` command must name exactly one `writes.json` typed action, and every
typed action must have exactly one generated mutation candidate with the same declaration ID. The
audit currently proves 204 commands, 204 distinct typed actions, and 204 candidates (11 creates,
8 updates, and 185 destructive deletes). A regression test keeps that future destination input
complete before the typed-destination foundation landed. This remains a useful one-to-one command/action/candidate invariant, not the current availability boundary.

## Relaunch continuation — post-foundation reconciliation (2026-08-23)

The temporary #4304 stack has landed through `main`; PR #4285 is again a draft PR to `main`.
The connector does not retain a destination declaration merely because the shared executor can
compose one: destination eligibility must preserve the provider action's semantics.

### Exact-SHA optional-query rehearsal (2026-08-22)

The former Foundation SHA `c3f83cbf6eabbae00219566fb02719ca2d6c480d` was rehearsed only
in an isolated detached temporary worktree. Its exact Zoom Meeting DELETE declaration emitted a
fixture-approved loopback request while omitting absent `occurrence_id` and
`cancel_meeting_reminder` query fields and retaining a present optional value. The full
SHA-bound evidence is in [`FOUNDATION-REHEARSAL.md`](FOUNDATION-REHEARSAL.md). The optional-query
Foundation has since reached `main`; this rehearsal changes neither certification nor final merge
readiness. #4332 has since landed the rendered-reference citation contract. A Zoom migration probe
then advanced source validation to its terminal capture boundary: every original Next-data artifact
URL now returns HTTP 404 and no verified cache contains the pinned bytes. The 2026-08-23 stable
capture decision requires an attested mirror for those 35 captures; this lane keeps the preserved
v2 lock unchanged and does not invent a v3 artifact or descriptor while that separate foundation
slice is absent.

The `users.id -> user_id` field overlap matches eight provider DELETE actions and no ordinary
non-delete action. A source row replay would therefore issue a destructive provider deletion;
`internal/app/issue_label_warehouse_transport.go:944` correctly refuses that as an ordinary
`full_append` apply action. The connector declares its three executable ETL sources only. All eight
DELETE commands remain implemented direct CLI commands with their existing destructive confirmation
and reverse-ETL approval lifecycle, but none is declared as a sync destination. The committed
source-traced gap `declarative-typed-destination-delete-semantics` records the bounded shared
capability needed for a future tombstone-aware transport. No live mutation is authorized by this
decision.

The continuation also replaces the stale generic-destination deferral with an auditable
seven-surface readiness ledger. Every documented source or ledger operation carries provider
provenance, parity classification, command/action binding where declared, source/destination
transport eligibility, implementation state, certification state, and a recoverable technical gap
when not enabled. Destructive, privileged, uncommon, binary, or not-live-certified operations stay
user-reachable whenever their established typed contract exists; only a named contract or executor
limitation may keep an operation disabled.

## Retained-source and command-surface follow-up (2026-08-24)

The retained-source foundation is now available on `main`. Zoom re-pins 34 current first-party
OpenAPI documents as exact connector-owned artifacts (11,719,368 bytes); the historic Accounts
capture is explicitly `unavailable` because its dated capture URL returns HTTP 404 and no verified
historic bytes exist. The lock records 1,871 current operations plus 66 historical Accounts
identities in the crosswalk. This is a blocking source-projection gap, not a fabricated artifact or
descriptor.

The first full package gate also discovered stale generated flags: canonical source-derived flags
had been added without removing legacy aliases, producing duplicate `maps_to` bindings that the
runtime correctly rejects. The connector-local repair removes only those duplicate aliases, keeps
each command implemented and user-reachable, and proves the no-credential boundary through the
full Zoom package test.

Before final delivery, fixture proof is insufficient. The connector additionally exposes the narrow
source-backed Meeting create/update contracts required for the lane-owned reversible lifecycle.
After merging the final updated #4304 head and
proving it is an ancestor, build `pm` and use the registered Zoom secret-store reference only at
process execution time to prove: authenticated read; a unique lane-owned create/read-back/update/
delete-cleanup sequence; ETL; reverse-ETL plan/apply/acknowledgement plus independent provider
read-back; and any documented, safely supported binary round-trip. No pre-existing resource,
account/security operation, raw secret, token, provider response, or credential identifier may be
stored in evidence. A missing secret reference or required scope is a keyed blocker, not a reason
to certify from fixtures.

## Required skills

Loaded for this connector/CLI/test/documentation work: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`,
`golang-structs-interfaces`, and `golang-documentation`. The work remains connector-local; no shared
engine, auth, generator, certification allowlist, or status code is modified.
