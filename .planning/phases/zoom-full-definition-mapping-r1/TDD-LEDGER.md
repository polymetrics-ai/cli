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

## Declaration contract tests

Red: `TestSourceBackedOperationInventoryKeepsEveryContractVisible` first observed zero source-backed
operation contracts (wanted 1,820 candidates after protected/external cohorts), and the reverse-ETL
fixture test observed zero warehouse destination actions (wanted two). A first generated operation
pass also failed the operations schema because provider parameter enum values included non-strings;
the source fact is retained in the crosswalk, while the unsupported field is omitted rather than
coerced in `operations.json`.

Green: the connector-local inventory now contains 1,748 source-backed operation contracts (776
`rest_read`, 971 `rest_write`, and one `binary_download`) with 311 typed destructive DELETE
contracts. The two real destination actions run through plan/preview and a loopback HTTP fixture.
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

Green: `TestSourceTransportDeclaresEveryExecutableZoomStream` proves the complete users/meetings/
webinars source allowlist, declarative executor reference, five source modes, delivery semantics,
and conformance run. Reverse ETL remains undeclared: `generic-typed-destination-executor` cites
`internal/app/issue_label_warehouse_transport.go:85-95`; no action binding is invented.

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
