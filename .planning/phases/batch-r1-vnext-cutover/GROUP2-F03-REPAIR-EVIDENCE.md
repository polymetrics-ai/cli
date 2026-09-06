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
bytes mechanically from that immutable baseline plus only the specified
completed target-file records and formatter boundaries. The bounded report is
`/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-group2-red-reconstruction-report.md`
(report SHA-256 `f51e2c8d3011662368c4b1958e4fc378df9de30a2150515d630e7427190f953b`).
It identifies baseline blob `e29ee4a316be686b4e6e91e4d3f5f0c2a5421539`
(baseline SHA-256 `48456f317c6ae2cc68b3ee085324f50f47946b288bf93a45aeceb5e013fa3890`),
records physical lines `10235`, `10427`, `10560`, `10567`, and `10588` plus
the formatter boundary at the actual RED record physical `10596`, and produces
the exact reconstructed final test SHA-256
`041f44816b2bd103a9b133dea196d6a68c243e64c6751375944ca2d5feb3a228` (261
lines). That hash is a mechanically derived historical snapshot, not a hash
recorded when the test executed, an independently observed full working-tree
identity, a current-source hash, or a new run. The report binds the real
physical `10596` command/exit/package duration above and excludes the earlier
compile-only setup failure.

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

## Follow-up dynamic proof matrix — steer 058 (incomplete)

The focused post-repair selector proves only the rows stated above. The
following frozen-ledger coverage remains required before Group 2 can be called
complete; all later additions are labelled post-repair coverage, never
retroactive RED evidence.

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
