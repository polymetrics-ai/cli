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

Each matrix fixture records decoded CURRENT/JOURNAL values at the interruption
or refusal before restoring only test-owned A/B names. The former claim that
ordinary Publish creates no private repair authority is corrected: its normal
publication fixtures retain real transaction authorities (the audited probe
observed six). Group 3 must add descriptor-safe raw control and real
transaction/prepared/phase/anchor identity witnesses before it can claim the
caller/cut proof complete.

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

## 2026-09-06 — steer 061/062 public nested and durable-witness execution

This is a current test-only proof update, not a new independent review,
production behavior claim, or CP11 acceptance. It follows Group 2 checkpoint
`54746816735a964d0177a7a64646d29561f08180` and evidence checkpoint
`e1dced72fce88eed2f7dfd860cdeeead57d32972`.

- **F-02-P public recursion:**
  `TestVNextPublicationPublicNestedQuarantineBoundsChildOwnership` executes
  actual Recover-owned-stage cleanup, public Prune stale-generation cleanup,
  and successful Publish final-prune cleanup. For each, the root crossed into
  its real quarantine before either a nested directory A→B replacement after
  identity/before open, or an injected identity error only after `file.Stat()`
  on the real nested descriptor. Four no-GC repetitions per caller/fault show
  no numeric descriptor growth. The owned-stage row also descriptor-safely
  reads and decodes its actual stage-owner marker before cleanup. Before
  fixture reconstruction the test observes the public root absent, original
  root identity at the quarantine candidate, nested A/B identity/type/bytes
  (or retained nested A), nonempty partial residue, raw durable heads, and the
  retained private authority graph. Each fresh recovery must leave no open
  descriptor naming that unique test root (including its root, quarantine,
  nested-child, and connector-lock paths). Earlier owned siblings may have
  been deleted; neither the code nor this evidence promises all-or-nothing
  deletion or automatic restoration. Only then does the fixture replace its
  own candidate with a pre-call fixture backup and use fresh Recover.
- **F-05/F-06-P raw durable state:**
  `vNextPublicationOptionalFileWitnessForTest` and
  `vNextPublicationAssertDurableCutWitnessForTest` make CURRENT/JOURNAL
  presence and raw regular-file type/inode/bytes descriptor-bound. The witness
  decodes the same bytes, snapshots every selected/rejected/stale generation
  root, and opens the real control authority under a shared operation lock. It
  retains the marker and validates/snapshots each transaction's prepared
  record, phase chain and anchors before fixture restoration. Initial Publish
  does retain private authority; the old contrary sentence above is corrected.
- **Caller and variant coverage:** existing named public controls plus
  `TestVNextPublicationUnheldDurableRowsRetainEmptyLeaseReplacementB` execute
  explicit Prune, no-JOURNAL Recover/Open, Publish initial recovery,
  prepared/committed new-selected recovery, successful final Publish prune,
  rejected-new recovery, immediate rollback, owned-stage cleanup, and Check.
  The unheld durable rows have separate empty and nonempty B executions. Their
  B is the actual regular `.lease` member; Group 2 F-03-C remains the separate
  replacement-directory B proof.
- **Commands:**

  ```text
  go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextPublicationPublicNestedQuarantineBoundsChildOwnership|TestVNextPublicationUnheldDurableRowsRetainEmptyLeaseReplacementB|TestVNextGenerationPublisherRefusesLateLeaseReplacementAcrossPublicCleanupCallers|TestVNextGenerationPublisherCheckIsReadOnly|TestVNextPublication(CommittedJournalNewSelectedRecoveryRejectsLateLeaseReplacement|SuccessfulPublishFinalPruneRejectsLateLeaseReplacement|FreshRejectedNewRecoveryRejectsLateLeaseReplacement|ImmediateRollbackRejectsLateLeaseReplacementIdentityVariants))$'
  # exit 0, ok polymetrics.ai/cmd/connectorgen 30.270s

  go test -race -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextPublicationPublicNestedQuarantineBoundsChildOwnership|TestVNextPublicationUnheldDurableRowsRetainEmptyLeaseReplacementB|TestVNextGenerationPublisherRefusesLateLeaseReplacementAcrossPublicCleanupCallers|TestVNextGenerationPublisherCheckIsReadOnly|TestVNextPublication(CommittedJournalNewSelectedRecoveryRejectsLateLeaseReplacement|SuccessfulPublishFinalPruneRejectsLateLeaseReplacement|FreshRejectedNewRecoveryRejectsLateLeaseReplacement|ImmediateRollbackRejectsLateLeaseReplacementIdentityVariants))$'
  # exit 0, ok polymetrics.ai/cmd/connectorgen 42.177s
  ```

The earlier full-package results below are historical records for their stated
source trees. They are not relabelled as evidence for these later Group 3 test
changes; current full-package/static/fresh-review gates remain required.

## Steer 063 — object-kind correction and completed Group 3 matrix

The 063 clarification is evidence provenance, not new scope: Group 2 F-03-C
substitutes quarantine/allocation **directories** (empty or carrying foreign
bytes), while F-05/F-06 substitutes only the destructive generation's regular
`.lease` member with empty or nonempty lease-file B. The public nested F-02-P
case separately substitutes a nested **directory**. No test was rerun merely
for that wording correction. The eight preserved changing-source executions,
including the two intentionally retained fixture failures and their later
focused normal/race passes, are bound at
`/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-group3-focused-execution-receipts.json`;
they are not collapsed into a claim about one frozen source tree.

| Required caller/cut | Executable selector and B/fault variant | Pre-restoration state/resource proof | Limited recovery proof |
| --- | --- | --- | --- |
| Recover owned stage; Prune stale generation; Publish final stale-generation prune | `TestVNextPublicationPublicNestedQuarantineBoundsChildOwnership`: each caller × nested directory A→B after identity/before open and opened-child post-`Stat` failure; four attempts with GC disabled | Public root absent; quarantined candidate retains root identity; nested A/B or retained-child identity/type/bytes; nonempty partial residue. The owned-stage cell reads/decodes its real owner marker. Every cell observes raw heads and retained marker/transaction/prepared/phase/anchor graph before fixture reconstruction. | Only fixture-owned candidate replacement is reconstructed; fresh Recover succeeds; lsof finds no descriptor naming the unique root/quarantine/child/connector-lock tree and numeric process descriptors do not grow. |
| Explicit Prune; no-JOURNAL Recover/Open; Publish initial recovery | `TestVNextPublicationUnheldDurableRowsRetainEmptyLeaseReplacementB` (empty `.lease` B) and `TestVNextGenerationPublisherRefusesLateLeaseReplacementAcrossPublicCleanupCallers` (nonempty `.lease` B) | Selected CURRENT/no JOURNAL, A/B regular-file type/inode/bytes, selected/stale roots, raw controls, and real private authority are asserted at the actual refusal. | Restore only A; a fresh publisher Recover is required. |
| Prepared and committed JOURNAL, new selected | Empty B: `...UnheldDurableRows.../prepared-new-selected-recover` and `/committed-new-selected-recover`; nonempty B: existing prepared public-caller and `TestVNextPublicationCommittedJournalNewSelectedRecoveryRejectsLateLeaseReplacement` | `AfterCommitSync` is prepared old/new with new CURRENT; `BeforePrune` is committed old/new with new CURRENT. Raw control files, roots, and authority graph are descriptor-safe witnesses. | Fixture-only A restoration then fresh Recover. |
| Successful Publish final prune | Empty B: `...UnheldDurableRows.../successful-publish-final-prune`; nonempty B: `TestVNextPublicationSuccessfulPublishFinalPruneRejectsLateLeaseReplacement` | New CURRENT and committed old/new JOURNAL advance legitimately; pre-quarantine stale root identity/content and private authority remain observable before restore. | Fixture-only A restoration then fresh Recover. |
| Old-selected rejected-new recovery; immediate validation rollback | `TestVNextPublicationFreshRejectedNewRecoveryRejectsLateLeaseReplacement` and `TestVNextPublicationImmediateRollbackRejectsLateLeaseReplacementIdentityVariants`, each retaining existing empty/nonempty `.lease` B variants | Prepared old/new versus rollback-restored-old logical state, A/B identity/bytes, selected/rejected roots and durable/private witnesses are retained. | Existing fixture-only fresh recovery/cleanup path, with no restart or all-or-nothing claim. |
| Owned-stage cleanup; non-destructive Check | `TestVNextGenerationPublisherRefusesLateReplacedValidatedStageCleanup` and `TestVNextGenerationPublisherCheckIsReadOnly` | Stage uses its real owner marker, never an invented lease-deletion case. Check preserves selected control/tree without destructive navigation; the durable witness records raw controls and authority. | Stage replacement remains no-replace protected; Check creates no recovery claim. |

After adding the root-scoped descriptor and stage/control observations, the
current focused matrix passed on the changed test source:

```text
go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextPublicationPublicNestedQuarantineBoundsChildOwnership|TestVNextPublicationUnheldDurableRowsRetainEmptyLeaseReplacementB|TestVNextGenerationPublisherRefusesLateLeaseReplacementAcrossPublicCleanupCallers|TestVNextGenerationPublisherRefusesLateReplacedValidatedStageCleanup|TestVNextGenerationPublisherCheckIsReadOnly|TestVNextPublication(CommittedJournalNewSelectedRecoveryRejectsLateLeaseReplacement|SuccessfulPublishFinalPruneRejectsLateLeaseReplacement|FreshRejectedNewRecoveryRejectsLateLeaseReplacement|ImmediateRollbackRejectsLateLeaseReplacementIdentityVariants))$'
# exit 0, ok polymetrics.ai/cmd/connectorgen 32.789s

go test -race -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextPublicationPublicNestedQuarantineBoundsChildOwnership|TestVNextPublicationUnheldDurableRowsRetainEmptyLeaseReplacementB|TestVNextGenerationPublisherRefusesLateLeaseReplacementAcrossPublicCleanupCallers|TestVNextGenerationPublisherRefusesLateReplacedValidatedStageCleanup|TestVNextGenerationPublisherCheckIsReadOnly|TestVNextPublication(CommittedJournalNewSelectedRecoveryRejectsLateLeaseReplacement|SuccessfulPublishFinalPruneRejectsLateLeaseReplacement|FreshRejectedNewRecoveryRejectsLateLeaseReplacement|ImmediateRollbackRejectsLateLeaseReplacementIdentityVariants))$'
# exit 0, ok polymetrics.ai/cmd/connectorgen 45.628s
```

These focused results supersede neither the preserved 063 records nor the
later required exact-final-source package/race/static and independent-review
gates.

## Final current-source package and static boundary

After the final root-scoped resource and control-authority observations, the
complete package boundaries passed on this exact pre-commit source tree:

- `go test -count=1 -timeout 20m ./cmd/connectorgen` → exit 0 / `ok polymetrics.ai/cmd/connectorgen 319.928s`.
- `go test -race -count=1 -timeout 20m ./cmd/connectorgen` → exit 0 / `ok polymetrics.ai/cmd/connectorgen 783.834s`.

The final static boundary also passed: formatter, `git diff --check`,
`go vet ./cmd/connectorgen`, `go build ./cmd/connectorgen`, `go build ./cmd/pm`,
`go mod tidy -diff`, `go run ./cmd/agentcontractgen check`, and
`golangci-lint run --new-from-rev=HEAD ./cmd/connectorgen/...` (0 issues).
This is local hermetic proof only. It does not accept CP11, certify a provider,
or replace Firstmate's required independent exact-SHA review.

## Historical full-package revalidation after F-02 lint cleanup

The final test-only cleanup checks the two F-02 parent-directory `Close`
errors rather than discarding them. Its focused F-02 selector passed in 1.274s
and `golangci-lint run --new-from-rev=HEAD ./cmd/connectorgen/...` reported
`0 issues.`. Because this changed current test code after the first broad
run, the package and race gates were repeated on that then-current source
state, before the later Group 3 observation additions:

- `go test -count=1 -timeout 20m ./cmd/connectorgen` → exit 0 / `ok polymetrics.ai/cmd/connectorgen 263.677s`.
- `go test -race -count=1 -timeout 20m ./cmd/connectorgen` → exit 0 / `ok polymetrics.ai/cmd/connectorgen 691.666s`.

Current formatter diff, `go vet ./cmd/connectorgen`, both `cmd/connectorgen`
and `cmd/pm` builds, `go mod tidy -diff`, `agentcontractgen check`, and
`git diff --check` also passed. Earlier current-wave generation/canon/docs,
553-definition validation, runtime preflight, connector boundary, and release
checks remain recorded in the TDD/verification record; this test-only close
fix added no generated connector or public documentation change.

## Current independent-gate accounting — Firstmate 064

The behavioral candidate remains
`7481d1770a21cc95869fd10bf281f632af48c089` (tree
`a2e583336ffa8ad86a0de95110259342bfa6dab0`). This is an artifact-only
accounting addendum: it changes neither test nor production source and makes
no CP11 acceptance, provider, power-loss, release, or merge claim. The
candidate commit itself changes three test files and five evidence files;
`git diff --check 7481d177^ 7481d177` exits 0. The worktree continued to show
only protected untracked `.cache/`, which was not read, staged, or modified.

| Gate | Exact command / source | Current result |
| --- | --- | --- |
| Source lock | `make connectorgen-vnext-locks` (Makefile 94–95) | exit 0, `ok polymetrics.ai/cmd/connectorgen 206.821s` |
| Connector canon | `make connector-canon-check` (Makefile 103–107; `scripts/tests/connector-canon.sh`) | exit 0, `connector canon check: ok` |
| Definitions | `make connectorgen-validate` (Makefile 88–89) | exit 0, `553 connector(s) checked, 0 findings` |
| Foundation Atlas | `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestFoundationAtlasSelectorsResolve$'` | exit 0, package `1.326s` |
| Runtime preflight | `make connector-runtime-preflight` (Makefile 97–101) | exit 0; cached Go result, not fresh execution proof |
| Docs | `make docs-check` (Makefile 41–44, including build) | exit 0; `pm` built and connector docs validated |
| Release workflow | `make release-workflow-check` (Makefile 112–120) | exit 0; pinned dependencies, Homebrew notification, darwin/linux arm64/amd64 parity, tooling, size budget, and production layout reported passed |
| Boundary scanner | `make connector-boundary` (Makefile 109–110; `go run ./cmd/connectorgen boundary . --json`) | **Unresolved:** the concurrent wrapper retained neither session, exit status, nor output. Its processes later ended, but termination is not a pass and the command was not rerun. |

The full normal/race package receipts above remain the final exact-candidate
test receipts (`319.928s`/`783.834s`), rather than substitutions for these
different gates. The detailed external record is
`/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-7481d177-current-gate-report-064.md`
(SHA-256 `a9bcdf60d0fb4945e096216727a39344ae87816e18abde8b6bdb71022e2bc908`).
Independent Astra exact-SHA review remains required, and the boundary gate
remains pending Firstmate recovery or explicit rerun authority.


## 2026-09-06 — Firstmate 068/070/073 evidence correction; audit pending

This dated correction supersedes only earlier current-state claims; all
original receipts and historical text remain preserved. Behavioral/test source
remains `7481d1770a21cc95869fd10bf281f632af48c089`, tree
`a2e583336ffa8ad86a0de95110259342bfa6dab0`. Current descendant `afde575a`
changes evidence only. No test, production or dependency change accompanies
this correction.

The original current boundary scanner receipt was recovered by Firstmate068:
exit 0, wall 254.579597042s, clean 284 files/553 connectors, zero findings or
warnings, six existing exceptions. Its preserved JSON has SHA-256
`744d0d129e15c8eccbaf723dda0ca96487c8babc25ce3c2d53d6feea25ed5849`;
raw physical 13951 has SHA-256
`ec88ff8c01c57e8647206c78ff003698932b66be9420a737536073ddd75c07b7`.
The complete private 064 report after its 068 appendix has SHA-256
`a59da5a5752e5c2149ca474cf5b9ac5761fa494fe7ae470ee04084a5680ba779`;
its older `a9bcdf60...` hash is prior chronology, not a conflicting result.
All eight supplementary gates therefore have successful results, with runtime
preflight still explicitly cached. No scanner or unchanged test was rerun.
The separate boundary-package failure remains CP29 debt. The final normal
319.928s/race 783.834s receipts are intended-source pre-commit results;
new-only Group 3 lint zero does not establish whole-original-range lint green.

The complete fresh independent Astra/xhigh review returned CHANGES REQUIRED
at 7481 (SHA-256
`2d92ce239d19509aa1838c23d5d0b9f31f4e5784232a512ee2b0f93e70ca571c`).
Its five entries are unadjudicated: 7481-01 successful-create/helper-completion
ownership, 7481-02 compound EEXIST retry, 7481-03 allocation A/B proof,
7481-04 expected durable-state oracle, and 7481-05 nested phase/evidence.
Earlier Group 2 complete and Group 3 provisional-remediation language is
challenged by those findings and cannot be treated as current closure.
F-04-R/F-08-R are resolved by that review; other reviewed obligations retain
the report's partial dispositions. The whole original report and all five
entries/prior-seven dispositions are inputs to the independent complete-ledger
audit in REVIEW-CONVERGENCE.md. No repair, CP11 acceptance, CP12 advance,
provider work, no-mistakes run, push or merge follows from this record.


## 2026-09-06 — Complete Firstmate073 audit disposition

The independent fresh Astra/xhigh complete-ledger audit returned **CHANGES
REQUIRED: six actionable entries (one Medium, five Low)** against behavioral
candidate `7481d1770a21cc95869fd10bf281f632af48c089`. Its full 315-line original
text, all dispositions, ten lenses, ownership/consumer and eleven-caller/cut
evidence tables are preserved verbatim in [REVIEW-CONVERGENCE.md](REVIEW-CONVERGENCE.md)
under the dated complete-audit return. Original report SHA-256:
`bc109e85fdde9d1958b2cde7874a3f7b30b8e5d06b1b0c2764088fb34fa3e0a0`.
This supersedes current closure implications in earlier entries; it does not
rewrite their historical source, test receipts or original attribution.

7481-01/02 require common-record creation/partial-record ownership repair
(including prepared, marker and phase siblings) and complete compound-collision
error classification. 7481-03/04/06 require the previously stipulated exact
allocation A/B and fresh recovery witnesses, independent expected controls/
roots/stable history across every actual caller/cut including held-reader and
Check, and the omitted compound-error controls. The prior F03 dynamic matrix
is not completely closed. Missing proof does not establish renewed B deletion
or a current recursive descriptor leak.

7481-05 adjudicates the nested Publish fixture as initial recovery, not final
prune. Its extra final-prune label and post-Stat retained-child identity claim
are unsupported. Mandatory public nested Recover-stage/Prune-generation proof
and the separate actual final-prune lease matrix remain. No extra nested
final-prune runtime guarantee is required. F04R/F08R remain resolved; the
other old-seven obligations retain the audit's partial dispositions.

The new external four-cell marker/phase overlay exited 1 (package 1.079s,
wall 5.26s), proving desired fresh-recovery failures on unchanged source.
It is not production GREEN. Current normal/race and eight supplementary
receipts, cached preflight, recovered 068 scanner success, historical F07
chronology and CP29 debt retain their exact original limits. Current
original-base-range lint attribution is still an acceptance gate.

Only evidence aggregation/correction has occurred. No production/test repair
has started. The complete audited set now awaits Firstmate's separately
authored coordinated repair/proof scope; CP11 remains unaccepted and CP12
has not begun.
