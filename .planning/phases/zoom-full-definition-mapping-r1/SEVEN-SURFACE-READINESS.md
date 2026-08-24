# Zoom seven-surface readiness

Issue: #4265
Base: `main` after the temporary reverse-ETL foundation stack landed. The retained-artifact
foundation is landed: 34 current direct API Hub OpenAPI documents are retained and source-traced;
the historic Accounts capture is explicitly unavailable rather than silently replaced.

[`sources/zoom-seven-surface-readiness.json`](../../../internal/connectors/defs/zoom/sources/zoom-seven-surface-readiness.json)
is the committed machine-readable reconciliation ledger. It contains 1,939 records: all 1,937
source-locked provider operations plus two ledger-only identities. Every record carries its pinned
provider provenance, parity class, all seven surface cells, any command/action/transport binding,
implementation state, certification state, and a recoverable gap when it is not implemented.

## Missing-foundation delivery

[`sources/zoom-missing-foundation-gaps.json`](../../../internal/connectors/defs/zoom/sources/zoom-missing-foundation-gaps.json)
is the required companion ledger. It has one deduplicated catalog of 11 shared gaps and
1,329 source-locked operation-gap rows across 1,299 provider operations. Every row records the
exact provider operation, document URL/revision/SHA-256, relevant surface(s), current validator or
runtime evidence, and an explicit `merge_ready_enabled: false`. The catalog supplies the shared
provider-neutral capability, issue/lane ownership (or an explicit unassigned foundation backlog),
status, exact closure commands, and complete operation fan-out. It also rolls up the 29 source
module batches and the Zoom portfolio.

This is deliberately not a second disabled/N/A inventory. An independently implemented command
keeps its runtime CLI reachability fact in the row; an open shared gap only blocks merge-readiness
and certification. The catalog-only REST operation-coverage gap has zero currently unmitigated Zoom
operations and remains visible rather than being silently dropped. The ledger does not use a
Zoom-specific executor or a connector-side workaround for any shared gap.

| Measure | Count | Honest status |
| --- | ---: | --- |
| Documented source operations | 1,937 | Reconciled; this is the source lock, not a completeness claim. |
| Retained current source operations | 1,871 | 34 exact first-party OpenAPI artifacts; all method/path/operation-ID sets match the preserved historical inventory. |
| Explicit unavailable historical operations | 66 | Accounts remains represented in the crosswalk and ledger, but its rotating capture is HTTP 404 with no verified historic bytes. |
| Reconciliation records | 1,939 | Includes two ledger-only identities. |
| Declared operation contracts | 1,748 | 776 reads, 971 writes, one binary-read contract. |
| Command-bound installed-CLI entries | 714 | 505 direct reads, three ETL streams, and 206 reverse-ETL commands, plus capability/transport entries. |
| Typed write actions / direct CLI reverse-ETL | 206 / 206 | Every typed action is directly user-reachable with existing approval and destructive safeguards. |
| ETL-bound streams | 3 | Preserved `users`, `meetings`, and `webinars` streams. |
| Reverse-ETL destination-bound actions | 0 | No ordinary replay destination is declared for a provider DELETE action. |
| Binary read / write contracts | 1 / 34 | Zero executable binary commands: Clip download awaits redirect-origin evidence; uploads await the `file_upload` executor and multipart-policy contract. |
| Disabled inventory rows | 1,155 | Each preserves a source or declaration-level reason; this lane does not call them provider-complete parity. |
| Certified operations | 0 | No command is silently certified; every reconciliation record is explicitly uncertified. |

## Reverse-ETL disposition

[`reverse-etl-eligibility.json`](../../../internal/connectors/defs/zoom/reverse-etl-eligibility.json)
gives all 206 typed actions a disposition. Destructive, privileged, and uncommon actions are not
made ineligible for those reasons; their existing plan → preview → approval → per-unit authorization
path remains mandatory.

| Reverse-ETL disposition | Count | Meaning |
| --- | ---: | --- |
| Directly CLI-reachable | 206 | All named actions have an implemented reverse-ETL command. |
| DELETE actions with an exact source-key overlap | 8 | `users.id` maps to normalized action input `user_id`, but each target action is a provider DELETE. |
| Direct CLI only — delete semantics | 8 | They remain implemented direct CLI commands, but no ordinary replay destination is declared: `internal/app/issue_label_warehouse_transport.go:944` rejects DELETE as `full_append`. |
| Direct CLI only — missing exact source fields | 197 | The definition has no stream that provides every required action input. No values or body fields are invented. |
| Direct CLI only — no source key | 1 | A source-record mapping cannot choose an unkeyed action safely. |

The former #4304 stack and optional-query behavior are now present through `main`. The rehearsal at
`c3f83cbf6eabbae00219566fb02719ca2d6c480d` remains isolated evidence in
[`FOUNDATION-REHEARSAL.md`](FOUNDATION-REHEARSAL.md), not certification. No Zoom-specific
dispatcher or generic writer is added. The subsequent delete-only reconciliation removes the
invalid destination declaration rather than replaying a provider DELETE or substituting another
action merely to claim coverage.

## Certification boundary

The connector definition cannot receive its current complete validation result while the Accounts
source is explicitly unavailable. `sources/zoom-source-repin-report.json` records 34 authorized
re-pins from rotating Next-data URLs to current first-party OpenAPI documents; `source-retain`
checked in their exact raw bytes. `source-import --check` stops before provider or cache fallback at
the Accounts unavailable reason, so no canonical descriptor is created from an error body. After
that source gap is resolved and the persisted App/CLI path is exercised, certification requires built-binary proof using only the registered
secret-store reference at process execution time: authenticated read; a unique lane-owned
create/read-back/update/delete-cleanup flow; ETL; reverse-ETL plan/apply/acknowledgement plus
independent provider read-back; and a binary round-trip where supported. No pre-existing resource,
account/security state, or destructive user action may be exercised. Fixture evidence, live proof,
independent readback, and cleanup remain separate fields in the ledger; none are currently promoted
to certification.

## Scope

This is a complete reconciliation and explicit eligibility accounting pass, not a claim that all
documented Zoom operations are executable. The 1,155 inventory-only rows remain visibly accounted
for with their exact current contract limitation. Provider entitlement, OAuth scope, destructive
classification, or lack of automated-live safety never turns an already-faithfully-modelled command
into an unavailable command; they restrict provider execution or certification instead. An operation
without an authored command/action contract remains an implementation gap, not a fabricated
foundation gap.

The captain's later hard gate is stricter than this branch's historical accounting: no operation
with an open foundation-gap row contributes to a merge-ready verdict, and no provider-wide CLI or
website reachability claim is made until every source operation is mapped, enabled, and proven
through the required six-surface evidence.

## Reproduction

The ledger is mechanically derived from the committed disposition, CLI surface, sweep, and
reverse-ETL eligibility records. Its invariants are enforced by:

```bash
go test -count=1 -timeout 20m ./internal/connectors/defs/zoom \
  -run '^(TestReverseETLEligibilityDisposesEveryTypedAction|TestSevenSurfaceReadinessAccountsForEveryProviderIdentity)$'
go run ./cmd/connectorgen validate internal/connectors/defs/zoom --json
go run ./cmd/connectorgen certification-matrix --connector zoom --check
go run ./cmd/connectorgen certification-sweep --connector zoom --check
go run ./cmd/connectorgen certification-candidates --connector zoom --check
go test -count=1 -timeout 20m ./internal/connectors/defs/zoom \
  -run '^TestMissingFoundationGapRowsAreSourceLockedAndRollUp$'
```

The two certification generators are currently green after the delete-only reconciliation:
`certification-candidates --check`, `certification-matrix --check`, and
`certification-sweep --check` report no drift. Full `connectorgen validate` and
`surface-sync --check` remain intentionally pending stable capture attestation; they must not be
made to pass by rewriting the preserved rendered-reference source evidence.
