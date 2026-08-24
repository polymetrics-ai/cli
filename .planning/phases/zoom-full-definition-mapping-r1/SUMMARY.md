# Summary — Zoom runnable command surface

Issue: #4265
Phase: `zoom-full-definition-mapping-r1`

## Delivered

- Preserved 35 captured Zoom Developer Docs modules (12,127,228 bytes; 1,937 source operations)
  and crosswalked them with the 1,913-row provider ledger. #4332 landed v3 rendered-reference
  citations, and a migration probe correctly advanced to the next terminal cause: the first
  immutable Next-data artifact returns HTTP 404 and no verified cache contains the pinned bytes.
  The 2026-08-23 stable-capture decision requires an attested mirror; the original lock is retained
  byte-for-byte and no unavailable URL is fabricated as a fetchable artifact or descriptor.
- Declared 1,748 source-backed executor contracts: 776 `rest_read`, 971 `rest_write`, and one
  bounded binary-download contract, including 311 destructive DELETE contracts.
- Added 505 source-backed direct-read commands, 202 approval-gated no-body scalar write commands,
  and two narrow, source-backed JSON Meeting lifecycle actions (create and update). Together with
  three preserved ETL commands and two existing write actions, Zoom now has 714 runnable commands,
  206 typed write actions, and 185 typed guarded deletes.
- Rebuilt connector manuals, catalog/website data, root-help golden transcripts, and the generated
  operation endpoint ledger from the command declarations.
- Declared the connector-neutral declarative ETL source for users, meetings, and webinars. The
  eight `users.id → user_id` matches are provider DELETE actions, not ordinary replay actions, so
  no `declarative_typed_destination` is declared. All eight remain implemented direct CLI commands;
  the explicit source-traced `declarative-typed-destination-delete-semantics` gap names the future
  tombstone-aware transport requirement and the current refusal at
  `internal/app/issue_label_warehouse_transport.go:944`.
- Generated one bounded authenticated direct-read candidate and 206 typed mutation candidates.
  Historical observed GET evidence remains explicitly non-certifying; current credentialed proof
  must be rerun against the final #4304 App/CLI dispatch head.
- Rehearsed the exact Zoom Meeting DELETE optional-query declaration against captain-approved
  Foundation SHA `c3f83cbf6eabbae00219566fb02719ca2d6c480d` in an isolated temporary worktree.
  The synthetic safety-approved loopback request omitted absent optional record fields and retained
  a present one; no Foundation ancestry, credential, provider call, or certification was added.
- Added a reverse-ETL eligibility ledger: all 206 implemented reverse-ETL commands, connector-owned
  typed actions, and generated mutation candidates form an exact one-to-one set; each action has an
  explicit transport disposition without sacrificing direct CLI reachability.
- Recorded a disposition for every 1,913 ledger row and 26 source-only rows. The 1,129 disabled
  ledger rows use only `foundation-gap`, `schema-incompatible`, `provider-does-not-expose`, or
  `requires-paid-tier`, with evidence and recovery state.
- Added the required missing-foundation ledger: 11 deduplicated shared-capability catalog
  entries and 1,329 exact source-locked gap rows across 1,299 provider operations. Each row has
  provider document URL/revision/hash, affected surface(s), runtime evidence, an owner or explicit
  unassigned foundation backlog, closure command, complete fan-out, batch/portfolio rollups, and
  `merge_ready_enabled: false`. Runtime CLI reachability remains a separate fact and no gap is
  hidden as disabled or N/A.

## Honest limits

- This is not provider-wide complete parity. It does not claim a runnable command for unmodelled
  JSON-body writes (the two narrow Meeting lifecycle contracts are explicit exceptions), array-query
  contracts, file uploads, the bounded Clip download redirect case, paid-tier operations, or
  source/ledger mismatches.
- No Zoom operation is certified yet. The accepted REST-read live proof lacks an exact operation
  fixture because the shared matrix projects fixtures only for capabilities, not operation kinds;
  `operation-specific-fixture-evidence-projection` records the minimal recovery.
- The SHA-pinned public OpenAPI response schemas provide creation timestamps for all three preserved
  streams, so the connector projects and declares a `created_at` cursor without inferring a
  watermark. The Webinar probe returned Zoom HTTP 400, error code 200: Webinar plan is missing;
  it is a recoverable `requires-paid-tier` observation for this account, not a product defect.
- The rerun passed catalog, Users append ETL, Meetings append ETL, and their query read-backs.
  It cannot be imported as accepted capability evidence because the shared fixture-conformance stage
  is still hard-wired to fail, the all-stream report aggregates the paid Webinar refusal, and the
  source-only connector receives irrelevant flow/schedule checks. Each foundation gap and its
  minimal recovery is recorded; no partial report is called certification.
- Reverse-ETL mutations remain unassessed. No Zoom destination is declared, because every exact
  source-key overlap is a provider DELETE and ordinary source replay would be unsafe. The direct
  CLI commands are still implemented; a tombstone-aware shared delete destination is the exact
  bounded future requirement.
- The committed seven-surface ledger accounts for all 1,937 source operations plus two ledger-only
  identities. This is not provider-wide executable completeness: 1,155 inventory rows remain
  explicit implementation or technical-contract gaps, and no provider entitlement or destructive
  classification is presented as a reason to hide an otherwise-modelled command.
- No auth, engine, generator, certification allowlist, or status code was changed.
- The v3 multi-document and rendered-reference foundations are on `main`, but stable capture
  attestation is still required before the 35 rotating Next-data artifacts can be re-verified.
  No Zoom source lock is migrated or descriptor synthesized until that separate foundation lands.
- The later captain hard gate means this remains explicitly **not merge-ready**: open foundation
  gaps, incomplete all-operation CLI/website reachability, and the missing final six-surface live
  proof prevent any provider-wide completeness claim.
