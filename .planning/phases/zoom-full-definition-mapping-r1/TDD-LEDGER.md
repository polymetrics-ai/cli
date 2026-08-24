# TDD ledger — Zoom full definition mapping

Issue: #4265
Phase: `zoom-full-definition-mapping-r1`

## Source-lock integrity

Red: the first attempted source lock serialized zero modules, so it could not substantiate any
declaration.

Green: the corrected public-only acquisition pins 35 module documents, their individual and
aggregate SHA-256 digests, byte counts, OpenAPI versions, server roots, method counts, and 1,937
source operations. An exact method/origin-path comparison finds 1,911 of 1,913 ledger identities:
1,908 blocked rows plus the three preserved ETL rows. The remaining two ledger-only Phone paths and
26 source-only paths are committed as explicit reconciliation records.

Refactor: server-root normalization is recorded in the source lock (`server_url`, `path`, and
`source_path`) so the standard `/v2` routes and the key-connector `/api/v2` route are compared
without changing their provider path contracts.

## Rendered-reference v3 migration probe (2026-08-24)

Red: after #4332 landed, the source-lock assertions were changed to require 35
`rendered_reference` documents, immutable per-document capture/provenance evidence, 1,937 nested
operation identities, and empty aggregate OpenAPI versions. The preserved v2 lock failed those
assertions, and `connectorgen validate` then advanced from the former `retrieval` parse rejection
to the missing canonical source descriptor.

Green (probe only): a mechanical v3 conversion preserved every document hash/byte count, all 1,937
operation IDs/routes/locations, and the existing provider citations. `go test -count=1 -timeout
20m ./internal/connectors/defs/zoom -run
'^(TestPinnedSourceCrosswalkAccountsForEveryIdentity|TestMissingFoundationGapRowsAreSourceLockedAndRollUp)$'`
passed against that conversion.

Terminal boundary: `go run ./cmd/connectorgen source-import zoom --out
internal/connectors/defs/zoom/sources/zoom-operation-descriptor.json` then fetched only the first
declared artifact and received HTTP 404 for `accounts`; no verified cache contained its SHA-256.
The cited provider URL was not fetched. Per the 2026-08-23 stable-capture decision, source bytes
may not be re-created from an unavailable URL or accepted without re-verification. The probe was
reverted and the original v2 lock restored byte-for-byte at SHA-256
`2e102acffd89467374405829abd994714f994f237c4a38c4ad0d9a553c42c3f7` pending the separate attested
mirror foundation.

## Declaration contract tests

Red: `TestSourceBackedOperationInventoryKeepsEveryContractVisible` first observed zero source-backed
operation contracts (wanted 1,820 candidates after protected/external cohorts), and the reverse-ETL
fixture test observed zero warehouse destination actions (wanted two). A first generated operation
pass also failed the operations schema because provider parameter enum values included non-strings;
the source fact is retained in the crosswalk, while the unsupported field is omitted rather than
coerced in `operations.json`.

Green: the connector-local inventory now contains 1,748 source-backed operation contracts (776
`rest_read`, 971 `rest_write`, and one `binary_download`) with 311 typed destructive DELETE
contracts. The two legacy typed write actions retain their plan/preview and loopback fixture
coverage; this statement is historical and does not certify the new destination template.
`TestPinnedSourceCrosswalkAccountsForEveryIdentity` proves the 35-document, 1,937-operation lock
and every 1,911 exact / 26 source-only / two ledger-only identity result.
`TestDeclarationDispositionAccountsForThePinnedSourceAndLedger` proves the 1,913 ledger rows plus
26 source-only rows have a declared or disabled result, and every disabled result has a
fixed-vocabulary reason, evidence, and recovery state. `go test -timeout 20m
./internal/connectors/defs/zoom -count=1`, `go run ./cmd/connectorgen validate
internal/connectors/defs/zoom --json`, and `go run ./cmd/connectorgen surface-sync
internal/connectors/defs --check` pass.

Red: before `sync_transport.json`, Zoom inspection projected no source role even though the merged
definition-owned adapter could execute the three existing declared streams.

Green (historical source-only checkpoint): `TestSourceTransportDeclaresEveryExecutableZoomStream`
proved the complete users/meetings/webinars source allowlist, declarative executor reference, five
source modes, delivery semantics, and conformance run. The later #4304 continuation adds the closed
destination declaration and preserves this source assertion.

## Candidate and live-proof boundary

Red: Zoom had no `certification.json`, so candidate generation could schedule neither a bounded
read nor the 204 typed mutation declarations. The first generated matrix also showed that an
accepted REST-read live record alone has `fixture_tested=false` for `operation:rest_read`.

Green: `TestCertificationCandidatesDescribeOneBoundedReadAndDeferWrites` proves one authenticated
self-read candidate plus all 204 explicitly unassessed mutation candidates. An isolated external
proof run obtained a short-lived OAuth token in memory, returned HTTP 200 for both guarded GET
exchanges, and imported a fingerprint-only `observed_operations` record. `TestAcceptedLiveReadProofDoesNotOverstateCertification`
proves that live evidence stays uncertified until the shared matrix can project an exact
operation-specific fixture. The recovery is logged as
`operation-specific-fixture-evidence-projection`; no auth, engine, generator, or status code was
changed by this lane.

Red: the attempted bounded full certification passed the Zoom `users` and `meetings`
full-refresh-append stages, but its catalog stage rejected every preserved stream for lacking
`cursor_fields`. The bounded Webinar request additionally received provider HTTP 400.

Green: verified against the SHA-pinned public OpenAPI response schemas that `user_created_at` /
`created_at` is a date-time creation timestamp for users and `created_at` is the creation timestamp
for meetings and webinars. Each stream now projects and declares its exact `created_at` cursor; no
watermark is invented. The fresh full run passed catalog plus Users and Meetings append ETL and query
read-back. The provider response was HTTP 400, code 200: Webinar plan is missing; it is recorded as
a recoverable `requires-paid-tier` certification finding with the account identifier redacted.

Red: the exact successful source stages cannot become accepted capability evidence: the fixture stage
unconditionally returns a wave0 no-bundle skip, all unrelated stream and schedule failures aggregate
into `Report.Passed=false`, and the external evidence importer correctly refuses a non-passing report.

Green: `definition-fixture-conformance-certification-stage`,
`capability-scoped-live-evidence-publication`, and `schedule-roundtrip-source-only-skip` record the
three exact engine limitations and minimal connector-neutral remedies. The lane leaves all three
cells uncertified rather than treating a partial full-run report as accepted proof.

Refactor criteria: run `connectorgen validate`, `connectorgen surface-sync --check`,
`make connector-boundary`, and then the full `make verify` before a push.

## Runnable-command correction

Red: `cli_surface.json` has only five commands despite 1,748 operation contracts, and 314 rows
labelled `requires-elevated-scope` plus 473 rows labelled `unsafe-to-exercise` are invisible to a
user even though a missing OAuth scope should be surfaced by the provider at runtime and an ordinary
delete must take the existing guarded write lifecycle.

Green: `TestRunnableOperationContractsHaveCommands` now proves 505 direct reads and 202 typed
no-body scalar write actions, including 185 guarded deletes, are each bound to a unique implemented
command and exact API-surface endpoint. Together with the three preserved ETL and two existing
fixture-backed actions, Zoom exposes 712 commands and 204 write actions. The 707 newly mapped rows
are `implemented-pending-certification`; none are certified. Elevated-scope and ordinary-delete
rejections are zero. The disposition report records 1,131 disabled ledger rows using only the
captain-approved reasons and the status format is `documented`, `commands`, `writes`, `deletes`,
`disabled`, and `ENABLED%`.

Red: the first command materializer mapped every direct-read parameter to `path.*` and copied
`allow_empty` to non-string flags. `connectorgen validate` rejected the latter, while a repeated
`surface-sync --check` exposed parameter drift and raw paging cursors.

Green: direct-read flags now map to the source-declared path or query location; the accepted
operation parameter set excludes provider paging mechanics before `surface-sync` derives flags.
The focused bundle test, validation, and a second `surface-sync --check` pass are green with zero
raw cursor flags and zero generated-field drift.

Safety stop: a REST write without a source-backed root request contract is not promoted through an
invented `--input` wrapper. The current runner accepts dotted `body.*` flags but only permits a
literal root body for `direct_read` (`internal/connectors/commandrunner/runner.go:1372-1430`). The
471 affected Zoom JSON-body writes stay disabled on `rest-write-root-json-input`; 25 source
contracts needing array query serialization stay disabled on `rest-query-array-encoding` until an
approved foundation change makes their typed transport possible.

## Reverse-ETL destination readiness under merge freeze

Red (historical): the prior tests proved runnable source-operation commands, but did not make the
complete three-way relationship between `cli_surface.json`, `writes.json`, and generated mutation
candidates an explicit invariant. Removing an action or candidate could therefore leave the future
typed-destination input incomplete until a later destination declaration failed closed.

Green: `TestEveryTypedZoomActionHasReverseETLCommandAndCandidate` asserts exact 204-way set equality:
every implemented reverse-ETL command names one unique typed action, every action has one command
and one `write_action`/`reverse_plan` candidate, every command address agrees with its action method,
and all candidates retained the deferred typed-destination classification. The current cohort is 11
creates, 8 updates, and 185 destructive deletes. The #4304 continuation now supersedes the
source-only part of this checkpoint with explicit destination and eligibility evidence.

## #4304 production typed destination continuation

Red: after merging #4304, Zoom still declares only a source transport and every generated mutation
candidate says `generic_typed_destination_executor_deferred`. A connector-local assertion therefore
fails until a definition-owned destination names an exact typed action, action/source input mapping,
acknowledgement, delivery facts, conformance evidence, and one strategy per declared mode.

Green: the Zoom bundle declares `declarative_api/declarative_typed_destination` for the existing
`zoom_users_userssotokendelete` action. The contract binds the already-declared `users` stream's
`id` to its normalized `user_id` action input (interpolated as provider `{userId}`), is restricted
to `full_append` / `append`, records keyed/durable acknowledgement
facts, and remains on the action's existing destructive plan/preview/approval lifecycle. No live
provider call occurs. Candidate classification changes from the obsolete executor-missing family to
an explicitly uncertified, fixture-and-live-proof-pending declared transport family; this does not
make the other 205 typed commands unreachable.

Refactor: replace the stale generic-destination gap with a seven-surface readiness ledger that
cross-checks the 1,937 source and 1,913 ledger inventory records, committed command/action bindings,
transport eligibility, and certification limitations. Run generated artifacts and all repository
gates before push.

## Delete-only destination reconciliation (2026-08-23)

Red: after the source-key audit identified eight `users.id -> user_id` candidates, the focused
transport and eligibility tests were changed to refuse an ordinary destination replay. They failed
because `sync_transport.json` still declared `declarative_typed_destination`, the eligibility ledger
still selected one DELETE and deferred seven as multiplicity, and the missing-foundation catalog
still treated the problem as dispatch/multiplicity rather than destructive replay semantics.

Green: `go test -count=1 -timeout 20m ./internal/connectors/defs/zoom -run
'^(TestMissingFoundationGapRowsAreSourceLockedAndRollUp|TestZoomTransportDeclaresTheExecutableSourceOnlyUntilDeleteSemanticsExists|TestReverseETLEligibilityDisposesEveryTypedAction|TestSevenSurfaceReadinessAccountsForEveryProviderIdentity|TestLaneOwnedMeetingLifecycleActionsAreClosedAndReachable)$'`
passes after removing the destination declaration. The source transport remains executable; all
eight DELETE commands remain implemented and directly CLI-reachable; the readiness and eligibility
ledgers record zero destination-bound actions and eight direct-CLI-only delete-semantic actions.
The source-traced `declarative-typed-destination-delete-semantics` gap cites
`internal/app/issue_label_warehouse_transport.go:944`, which refuses DELETE as an ordinary
`full_append` apply action. No alternate non-delete mapping and no transport bypass was introduced.

Manual-GSD boundary: this is connector-owned declaration reconciliation, not a new shared
foundation. A tombstone-aware delete destination remains a separately scoped shared capability;
this lane keeps the commands reachable without claiming they are sync destinations.

Red: after the certification source declaration was corrected, `go run ./cmd/connectorgen
certification-matrix . --connector zoom --check` and `go run ./cmd/connectorgen
certification-sweep . --connector zoom --check` both reported generated-artifact drift.

Green: rerunning `certification-candidates`, `certification-matrix`, and `certification-sweep` for
Zoom followed by each `--check` regenerated a 717-row sweep with 714 CLI commands and no drift.
The generated mutation candidates now state that the typed actions are direct CLI actions and cite
the delete-semantics limitation instead of claiming a declared destination.

## Lane-owned Meeting lifecycle continuation

Manual-GSD fallback: the relaunch supplied an already-dirty connector branch after the production
need for a reversible lifecycle became explicit. Reconstructing the original pre-edit state would
manufacture archival Red evidence, so that earlier state is not claimed.

Red: `TestLaneOwnedMeetingLifecycleActionsAreClosedAndReachable` now loads exact source-local
create/update/delete replay fixtures. It fails only on the existing Meeting delete action before
I/O: `omit_when_absent` rejects missing optional `record.occurrence_id` despite its declared
object-form query policy. Create and update also exposed unsupported `minimum` / `minLength` keywords, removed
from their connector-local schemas rather than worked around in the engine. The delete failure is
downstream evidence for issue #4305 (`cli-rest-structured-body-r1`), which alone owns the
declaration-closed shared-engine fix; no shared-engine edit is retained in this branch.

Foundation rehearsal Green: captain-approved exact SHA
`c3f83cbf6eabbae00219566fb02719ca2d6c480d` passed
`TestZoomMeetingDeleteOptionalQueryRehearsal` in an isolated detached worktree. The exact Meeting
DELETE action retained a present optional field and omitted both absent record-derived optional
query fields through a safety-approved loopback request. See
[`FOUNDATION-REHEARSAL.md`](FOUNDATION-REHEARSAL.md) for the command and constraints.

Current-branch Red remains: the preserved Zoom branch does not contain that Foundation revision,
so its connector-local lifecycle test remains an expected pre-I/O RED rather than a promoted
fixture pass. After Foundation lands on `main`, rerun all three lifecycle fixture request-shape
checks and then the built-binary lane-owned create/read-back/update/delete-cleanup proof after the
final #4304 persisted App/CLI dispatch head. The seven-surface and reverse-ETL eligibility ledgers
continue to account for 714 commands and all 206 typed actions; no certification is promoted by
the rehearsal.

## Captain-required missing-foundation ledger

Red: `zoom-foundation-gaps.json` names shared limitations only as aggregate counts. It does not
provide an exact source-locked provider operation row, document hash/revision, fan-out, owner,
closure command, or merge-readiness state for each affected operation. A missing executor could
therefore be confused with an ordinary disabled or non-applicable connector row.

Green: `zoom-missing-foundation-gaps.json` has a deduplicated shared catalog plus source-locked
operation rows and source-module/portfolio rollups. A connector-local test verifies every row's
provider identity/digest against the pin, every catalog reference, a non-empty capability/owner/
closure check, exact fan-out, and `merge_ready_enabled=false` for each open row. It also verifies
that no row calls an open shared gap disabled or non-applicable and that runtime CLI reachability is
reported separately rather than overwritten.
