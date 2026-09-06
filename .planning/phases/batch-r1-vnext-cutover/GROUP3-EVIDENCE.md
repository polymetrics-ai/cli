# CP11 Group 3 F-05/F-06/F-07 evidence

This record is a local, hermetic proof artifact for the coordinated CP11
repair wave. It is not CP11 acceptance, a provider-live result, or a substitute
for the required whole-package/race/static gates and fresh independent review.

## F-05 identity proof strengthening

- `TestVNextGenerationPublisherOpenKeepsValidatedDirectory` retains the opened
  directory descriptor A after a path-level A→B substitution. It records the
  returned descriptor's dev/inode/type against displaced A, the returned A
  `metadata.json` bytes, and B's distinct directory dev/inode/type and B
  metadata bytes.
- `TestVNextGenerationPublisherHeldGenerationUsesStableCleanupLock` records
  original/displaced lease A and replacement B before restoration for real
  explicit Prune, successful Publish cleanup, and Recover cleanup. Its B is
  deliberately empty just like A: the equal-byte/different-inode control proves
  the assertion is identity-based, not a size/byte proxy. CURRENT is compared
  to the expected cut (C after successful Publish, not mechanically to its
  pre-Publish selection); the recovered `AfterCommitSync` cut is correctly
  recorded as prepared-JOURNAL/new-selected before its committed replacement.
- `TestVNextGenerationPublisherRefusesLateLeaseReplacementAcrossPublicCleanupCallers`
  now records A/B dev/inode/type/bytes and selected controls for explicit
  Prune, no-journal Recover, Publish initial recovery, and transitive Open;
  its fifth subtest is relabelled prepared-JOURNAL/new-selected. The retained
  `TestVNextGenerationPublisherLateLeaseReplacementRetainsPublicGenerationCollision`
  remains the stronger A/B/C public collision witness. Immediate rollback has
  separate empty/nonempty B variants below.

This is a proof-only strengthening: production reader/lease behavior passed
before the oracle grew identity observations. The equal-byte control is not an
invented production RED.

## F-06 named durable caller/cut matrix

| Caller/cut | Executable fixture and durable observation |
| --- | --- |
| Explicit Prune | `RefusesLateLeaseReplacementAcrossPublicCleanupCallers/prune`: selected CURRENT survives, no JOURNAL, late A/B lease identity refusal. |
| No-journal Recover | `.../no_journal_recovery`: selected CURRENT survives, no JOURNAL, late A/B refusal. |
| No-journal Open | `.../open_transitive_recovery`: Open's recovery path reaches the same A/B refusal without returning a handle. |
| Publish initial recovery | `.../publish_initial_recovery`: recovery fails at the stale lease before new staging; existing CURRENT/no JOURNAL survive. The F-03 close-fault matrix separately exercises Publish entering a rejected-new recovery. |
| Prepared JOURNAL/new-selected | `.../prepared_journal/new-selected_recovery`: `AfterCommitSync` is before committed JOURNAL replacement; decoded JOURNAL is prepared with old/new and CURRENT is new. |
| True committed JOURNAL/new-selected | `TestVNextPublicationCommittedJournalNewSelectedRecoveryRejectsLateLeaseReplacement`: `BeforePrune` occurs only after `writeJournalLocked(committed)` returns; decoded CURRENT is new and JOURNAL is committed old/new before fresh recovery's A/B refusal. |
| Successful Publish final prune | `TestVNextPublicationSuccessfulPublishFinalPruneRejectsLateLeaseReplacement`: begins old-only, advances new CURRENT/committed JOURNAL, then the actual final stale-old prune refuses the late A/B replacement. |
| Old-selected rejected-new fresh restart | `TestVNextPublicationFreshRejectedNewRecoveryRejectsLateLeaseReplacement`: `AfterStageRename` leaves finalized new tree plus prepared old/new JOURNAL and old CURRENT; a fresh publisher substitutes new lease A with empty B and nonempty B in separate cases, refuses before removal, then fixture-only restoration lets a second fresh Recover remove new and JOURNAL while retaining old. |
| Immediate validation rollback | `TestVNextPublicationImmediateRollbackRejectsLateLeaseReplacementIdentityVariants`: active validation fails after new CURRENT, rollback restores old then late cleanup substitution refuses; empty/nonempty B variants retain validation and identity causes, controls, A and B before restored fresh recovery. |
| Stage cleanup | `TestVNextGenerationPublisherRefusesLateReplacedValidatedStageCleanup`: real owned-stage recovery checks its actual stage ownership contract, refuses replacement, and keeps selected CURRENT/no JOURNAL unchanged. No generation lease is invented for this row. |
| Non-destructive Check | `TestVNextGenerationPublisherCheckIsReadOnly`: Check preserves CURRENT bytes, leaves stale generation, and creates/removes no JOURNAL. |

Each matrix fixture records CURRENT/JOURNAL from decoded durable bytes at the
interruption or refusal before restoring only test-owned A/B names. No private
repair authority is created in these ordinary publication fixtures; the absence
of a repair transaction is therefore preserved rather than fabricated.

## F-03 public recovery caller completion

`TestVNextPublicationRecoveryCallersReportRestoreCurrentCloseError` now
contains Recover, Open, Prune, and `Publish_initial_recovery`. The latter calls
Publish against a genuine durable rejected-new state, so its `recoverLocked`
restore-CURRENT close failure is not confused with ordinary prepared/current/
committed Publish close cuts. It returns the injected close cause after durable
old CURRENT restoration, leaves the rejected tree/JOURNAL observable, then
restores only test-owned bytes for normal fresh recovery.

## Executed focused GREEN

`go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextPublicationRecoveryCallersReportRestoreCurrentCloseError|TestVNextGenerationPublisherOpenKeepsValidatedDirectory|TestVNextGenerationPublisherHeldGenerationUsesStableCleanupLock|TestVNextGenerationPublisherRefusesLateLeaseReplacementAcrossPublicCleanupCallers|TestVNextGenerationPublisherLateLeaseReplacementRetainsPublicGenerationCollision|TestVNextPublicationCommittedJournalNewSelectedRecoveryRejectsLateLeaseReplacement|TestVNextPublicationSuccessfulPublishFinalPruneRejectsLateLeaseReplacement|TestVNextPublicationFreshRejectedNewRecoveryRejectsLateLeaseReplacement|TestVNextPublicationImmediateRollbackRejectsLateLeaseReplacementIdentityVariants|TestVNextGenerationPublisherRefusesLateReplacedValidatedStageCleanup|TestVNextGenerationPublisherCheckIsReadOnly)$' -v`
exited 0 with `ok polymetrics.ai/cmd/connectorgen 16.675s`.

The command is a fresh focused execution of current uncommitted source/test
state. It does not run the whole package, race detector, provider, database,
or no-mistakes pipeline; those remain later gates.

## Current full-package revalidation after F-02 lint cleanup

The final test-only cleanup checks the two F-02 parent-directory `Close`
errors rather than discarding them. Its focused F-02 selector passed in 1.274s
and `golangci-lint run --new-from-rev=HEAD ./cmd/connectorgen/...` reported
`0 issues.`. Because this changed current test code after the first broad
run, the package and race gates were repeated on the final source state:

- `go test -count=1 -timeout 20m ./cmd/connectorgen` → exit 0 / `ok polymetrics.ai/cmd/connectorgen 263.677s`.
- `go test -race -count=1 -timeout 20m ./cmd/connectorgen` → exit 0 / `ok polymetrics.ai/cmd/connectorgen 691.666s`.

Current formatter diff, `go vet ./cmd/connectorgen`, both `cmd/connectorgen`
and `cmd/pm` builds, `go mod tidy -diff`, `agentcontractgen check`, and
`git diff --check` also passed. Earlier current-wave generation/canon/docs,
553-definition validation, runtime preflight, connector boundary, and release
checks remain recorded in the TDD/verification record; this test-only close
fix added no generated connector or public documentation change.
