# CP11 Group 2 F-03 repair evidence

## Intended-behaviour RED before Group 2 source repair

After the immutable original-control checkpoint
`8d1337829a28a0feaf21bb26062619b9d4a5b583`, the control assertions were
converted from defect observation to the desired ownership/error contract. No
Group 2 production behaviour had changed at this point. The first attempt was
a test setup compile failure (`Check` needed its artifact argument); it is not
credited. After that correction, the actual behavioural selector was:

```text
go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestCP11F03ARepairPostRecordFailureRetainsCoherentAuthority|TestCP11F03BRepairCompoundCausesRemainInspectable|TestCP11F03CRepairTemporaryCleanupPreservesReplacementB|TestCP11F03CRepairQuarantineCleanupPreservesReplacementB)$' -v
```

It exited 1 / `FAIL polymetrics.ai/cmd/connectorgen 2.567s` with all intended
failures observed:

- F-03-A found a prepared-only transaction after the post-record frontier.
- F-03-C found both temporary and quarantine replacement B paths absent after
  cleanup.
- F-03-B found the definitions-root cause flattened beside connector close,
  the missing-control absence cause flattened beside parent close, and the
  staged-file completion cause flattened beside the primary frontier.

The desired-assertion test source was uncommitted at that moment. The original
defect controls, their test-only seam identities, exact final source hashes and
2.687s output remain preserved immutably in
[GROUP2-F03-ORIGINAL-CONTROLS.md](GROUP2-F03-ORIGINAL-CONTROLS.md) and commit
`8d1337829a28a0feaf21bb26062619b9d4a5b583`.

Firstmate steer 056 independently reconstructed the desired-assertion test
bytes mechanically. The original report remains unchanged at
`/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-group2-red-reconstruction-report.md`
(SHA-256 `f51e2c8d3011662368c4b1958e4fc378df9de30a2150515d630e7427190f953b`).
Its corrective appendix is
`/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-group2-red-reconstruction-appendix.md`
(SHA-256 `e9f679bebb495aafe3ec1d8e7d96f71e11d8802fe630215e2dd2bc8c4f4aa190`),
with a derived snapshot and provenance JSON at the same prefix.

The appendix resolves the reconstruction route precisely. Immutable baseline
blob `e29ee4a316be686b4e6e91e4d3f5f0c2a5421539` has computed SHA-256
`48456f317c6ae2cc68b3ee085324f50f47946b288bf93a45aeceb5e013fa3890`.
The earlier add/update records `10234/10235` and `10426/10427`, with target
formatter records `10300/10301` and `10434/10435`, are baseline history and
were not reapplied. Only post-baseline desired conversion `10560`, unused-import
removal `10567`, literal target formatter `10575`, setup correction `10588`,
and the literal target formatter in the actual RED record `10596` comprise the
derivation. Physical `10483` is not a target formatter: it formats
`vnext_publication.go` only. The resulting 261-line test-file snapshot has
computed SHA-256 `041f44816b2bd103a9b133dea196d6a68c243e64c6751375944ca2d5feb3a228`.
That is a mechanically derived historical test-file snapshot, not a hash
recorded when the test executed, an independently observed full working-tree
identity, a current-source hash, or a new run. Physical `10596` preserves the
newline-separated literal `gofmt` then `go test` command and binds the recorded
exit 1/package duration 2.567s above; the earlier compile-only setup failure is
excluded.

The subsequent production edits are the coordinated F-03-A/B/C repair, not a
claim that this RED ran on repaired source. Empty and nonempty B, shared
writable record, capture completion, and every source-audited sibling remain
explicitly in the GREEN matrix.

## Focused GREEN

```text
go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestCP11F03ARepairPostRecordFailureRetainsCoherentAuthority|TestCP11F03BRepairCompoundCausesRemainInspectable|TestCP11F03CRepairTemporaryCleanupPreservesReplacementB|TestCP11F03CRepairQuarantineCleanupPreservesReplacementB)$' -v
```

The selector exited 0 / `ok polymetrics.ai/cmd/connectorgen 4.389s`.

- **F-03-A:** a complete, identity-bound prepared record is retained with its
  anchors across the post-record frontier and fresh `Recover`, `Check`, and
  ordinary `Publish` retry succeed. Before a complete record is established,
  the unwind removes a known prepared record before its anchors; an unknown or
  replaced record fails closed with its anchors retained.
- **F-03-B:** each exercised consumer retains both relevant causes through
  `errors.Is`: definitions/connector close, actual missing-control absence plus
  parent close, writable stage primary plus close, owned-stage cleanup, and
  capture preparation plus close. Pure absence remains the sole path returning
  `found=false` without an error.
- **F-03-C:** real public `Publish` temporary allocation and public `Prune`
  quarantine allocation preserve both the moved opened A and distinct empty and
  nonempty replacement B. The nonempty test verifies B's exact `foreign` bytes;
  the temporary control blocker and stale generation remain accounted for.

## F-03-B complete sibling audit

The focused dynamic controls exercise the main producer/consumer classes. The
remaining mandatory source siblings were checked in the same repair:

| Audited source class | Coordinated repair |
| --- | --- |
| Directory failed-open + parent close; parent close + opened-file close | `vnext_publication_dir.go` joins both causes. |
| Definitions-root + connector-root close | `openConnectorRoot` joins both real close results, including a failed child open plus root close. |
| Opened-control identity/read + close; absence consumer | `vNextPublicationReadControlBound` joins completion with identity/read failure and accepts only recursively pure `ErrNotExist` as absence. |
| Shared prepared/phase/authority-marker writable record | `vNextPublicationWriteControlRepairRecord` captures identity at creation, joins real close completion with Write/short-Write/Sync/assertion failure, and supplies the identity/creation state needed for owned cleanup. |
| Predecessor link + close | `createControlRepairLocked` joins an actual link failure with real predecessor close completion. |
| Staged-file Write/Sync + close; failed-stage owned cleanup | `writeStageFile` has one deferred real close owner; `stageLocked` joins known-owned remove/close results with the primary failure. |
| Capture-directory failure + Sync/Close | `beginControlCaptureLocked` records the prior fault, real Sync, and real Close without discarding either completion cause. |

The unchanged `resultErr == nil` policies in read-only authority/capture scan
helpers are deliberate: their existing primary error remains authoritative, no
writable completion or known-owned cleanup is pending, and this repair does not
claim blanket close propagation. Expected `ENOTEMPTY` during intentionally
retained quarantine remains a preservation result, not a fabricated cleanup
failure.

## Follow-up dynamic proof matrix — steer 058 (initial scope)

The focused post-repair selector proved only the rows stated above. This table
records the frozen ledger that was still required at that point; its cells are
executed below as post-repair coverage, never retroactive RED evidence.

| Audited contract | Current executed scope | Remaining executable proof |
| --- | --- | --- |
| F-03-A preparation frontiers | Post-complete-record injected frontier only. | Pre-record failure, Write and short-Write, Sync, real Close completion, post-record transaction Sync, and connector Sync; distinguish operation failure, partial result, and real operation plus injected completion. |
| F-03-A graph classes | Existing durable terminal-cut and valid bootstrap controls, but not this full frontiers matrix. | CURRENT/JOURNAL transitions; valid bootstrap present/present and absent/absent; reachable successor prior/intended present and logical-absence classes. At every cut observe prepared/phase/anchor graph and owned identities before fixture removal, then nonmutating Check and permitted fresh Recover/Check/retry. |
| F-03-B compound causes | Definitions/connector close, missing-control absence+parent close, staged Sync+Close, owned-stage cleanup, and capture pre-close+Close. | Real Open+parent Close and parent Close+opened-file Close; opened-control identity/stat/read+Close and pure-versus-compound absence at its consumer; shared prepared/phase/authority-marker Write/short-Write/Sync+Close; predecessor link+Close; staged writable completion+Close; capture primary with both Sync and Close. Each case needs `errors.Is`/`errors.As` at its relevant consumer/public caller and one actual Close owner. |
| F-03-C temporary and quarantine ownership | Public Publish temporary and public Prune quarantine each with empty/nonempty B; exact nonempty `foreign` bytes. | Bind the actual CURRENT/JOURNAL temporary uses and stage/generation quarantine reachability. Retain exact A/B identity/type/bytes and residue before restoration, meaningful primary/completion/cleanup causes, and bounded fresh recovery/retry. |

The intentionally bounded read-only close policy and intentional retained
`ENOTEMPTY` result remain in force. No source table, broad package receipt, or
synthetic isolated `errors.Join` substitutes for a listed resource-backed
case.

## Steer 058 executed resource-backed matrix

The current proof source is
`cmd/connectorgen/vnext_publication_group2_original_test.go`. It names the
actual producer/consumer functions rather than modelling them in isolation.
The narrowly inert seams it exercises are: `vNextPublicationControlRecordHooks`
and the record/directory-sync points in `vnext_publication.go` and
`vnext_publication_repair.go`; the opened-file-after-parent-close seam in
`vnext_publication_dir.go`; and the read-control completion/close seams in
`vnext_publication.go`. Nil hooks continue to call the direct descriptor
operations in production.

| Contract and exact selector | Executed resource-backed observation |
| --- | --- |
| F-03-A successor preparation frontiers — `TestCP11F03ARepairPreparationFrontierMatrix` | Three reachable transitions are run at before-record, actual short Write, actual Sync then injected completion, actual Close then injected completion, post-record, post-transaction Sync, and post-connector Sync: `JOURNAL` absent→present, `CURRENT` present→present, and `JOURNAL` present→logical-absence. Early failure retains no unexposed record/anchors; every later cut scans a complete phase-free prepared graph, identity-bound prior/intended anchors and unchanged public predecessor, verifies `Check` is nonmutating, then uses fresh `Recover`, `Check`, and retry. The injected Sync/Close outcomes are explicitly completion injection after the real operation, not a claim of a natural I/O or power-loss failure. |
| F-03-A valid base bootstrap — `TestVNextGenerationPublisherResumesInterruptedBaseAuthorityPreparation` and `TestCP11F03ARepairBasePresentPresentAuthorityRecovers` | The existing public `lock-render` witness retains both absent/absent first-`CURRENT` and second-`JOURNAL` cuts. The new isolated fixture retains a valid published generation/control but removes its historical authority graph, then interrupts a base prepared record for each present/present `CURRENT` and `JOURNAL` target. It asserts phase-free identity-bound anchors and nonmutating `Check` before fresh recovery, `Check`, and retry establish terminal `CURRENT`/`JOURNAL` heads. Fixture-only authority removal is not a production cleanup path. |
| F-03-B compound consumers — `TestCP11F03BRepairCompoundCausesRemainInspectable` | It observes both causes through their real consumers: failed open + parent Close; parent Close + opened raw-file Close; opened control identity/read completion + Close plus recursively pure absence; authority-marker/prepared/phase actual short Write, real Sync and real Close completion; predecessor link + Close; definitions/connector close; stage/capture primary plus Sync/Close and owned cleanup. Each injected completion calls the real descriptor operation first and asserts the joined causes with `errors.Is`; pure absence alone remains `found=false, err=nil`. |
| F-03-C temporary/quarantine ownership — `TestCP11F03CRepairTemporaryCleanupPreservesReplacementB`, `TestCP11F03CRepairCurrentAndJournalTemporaryPathsPreserveReplacementB`, `TestCP11F03CRepairQuarantineCleanupPreservesReplacementB`, and `TestCP11F03CRepairStaleStageQuarantinePreservesReplacementB` | Public Publish temporary, direct `CURRENT`/`JOURNAL` transition temporary, public stale-generation Prune quarantine, and stale-stage quarantine each move opened A and install a distinct replacement directory B: either empty or containing the exact foreign bytes. The test records A/B identity/type, exact nonempty B bytes, and residue before any fixture-only restoration. The direct transition refuses on identity change rather than claiming success; fresh recovery/Check/retry follows. The stale-stage test preserves the pre-cleanup residue before its explicit isolated-fixture cleanup and then proves fresh recovery/Prune retry. |

Current-source executions, all with `-count=1 -timeout 20m`, are:

- Normal F-03-A transition matrix: `go test ./cmd/connectorgen -run '^TestCP11F03ARepairPreparationFrontierMatrix$' -v` → `ok` in `20.804s` (test `19.85s`); present/present base: `TestCP11F03ARepairBasePresentPresentAuthorityRecovers` → `ok` in `2.617s`.
- Race F-03-A is bounded by state class so the terminal result is retained: absent→present → `ok` in `10.200s`, `CURRENT` present→present → `ok` in `10.089s`, `JOURNAL` present→logical-absence → `ok` in `11.507s`; present/present base `CURRENT` and `JOURNAL` → `ok` in `4.095s`.
- Normal F-03-B/C combined selector above → `ok` in `8.442s`; the same current-source selector under `-race` → `ok` in `11.749s`.
- After these current proof edits, `gofmt`, `git diff --check`, `go vet ./cmd/connectorgen`, `go build ./cmd/connectorgen`, `go run ./cmd/agentcontractgen check`, and `go mod tidy -diff` passed. This is focused/static validation, not a replacement for the later whole-package three-group boundary.

The earlier complete four-record receipt remains preserved at
`/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-group2-expanded-matrix-completion-receipts.json`:
physical `11540` normal package `20.344s`, and race physical `11549`,
`11563`, and `11577` package `27.750s`, `28.178s`, and `27.995s`, all exit 0
for the then-current two-successor-class matrix. It is neither lost nor
relabelled as the later three-class/current-source result.

This closes the resource-backed F-03 cells from steer 058. It does not accept
CP11, claim provider-live behavior, or replace the remaining Group 2
source/static/final-review boundary.

## Additional current validation

- The focused F-03 selector also passed under `-race` in 6.324s.
- `TestVNextGenerationPublisherRecoversEveryTerminalAuthorityDurableCut` passed
  in 23.293s across CURRENT/JOURNAL and prior-present/prior-absent rows at every
  existing durable terminal cut.
- `TestVNextGenerationPublisherResumesInterruptedBaseAuthorityPreparation`,
  `TestVNextPublicationPublishReturnsWritableCloseErrorAndFreshRecovery`, and
  `TestVNextPublicationAtomicCloseFailuresFollowTheirDurableCuts` passed in
  3.461s, preserving valid two-step bootstrap and writable close recovery.
- The original same-worker full-package receipt is retained at
  `/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-group2-intermediate-package-raw-result.json`:
  session `01a07295-93c2-7053-9c99-bb4ef2038f97`, physical record `10897`,
  raw SHA-256 `4fd7390eef0b432f8f5f983d3924f9360bde2993d51586fa48bc6658634e0255`,
  timestamp `2026-09-06T03:49:01.632Z`, and item
  `exec-bf5a4c93-588b-4dc9-83c9-1ca878debf18`. Its complete returned result is
  `go test -count=1 -timeout 20m ./cmd/connectorgen` → exit 0 / `ok
  polymetrics.ai/cmd/connectorgen 271.387s`, with wall time
  `273.672594250s`. This is a valid intermediate Group 2 package PASS, not an
  inferred process result or a later rerun. It exercised the then-uncommitted
  Group 2 source state; it does not bind a later working tree, complete the
  three-group normal/race wave, or accept CP11.
- `gofmt`, `git diff --check`, `go vet ./cmd/connectorgen`, `go build
  ./cmd/connectorgen`, `go run ./cmd/agentcontractgen check`, and `go mod tidy
  -diff` passed. The complete final three-group normal/race validation remains
  pending; the intermediate package receipt above is deliberately not promoted
  to that later boundary.
