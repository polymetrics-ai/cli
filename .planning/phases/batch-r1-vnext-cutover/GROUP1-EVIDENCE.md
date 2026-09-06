# CP11 Group 1 bounded evidence

## F-04 original-behaviour RED observation

- Fixture: `.cache/cp11-f01-f03-original-7e014d00`, materialized by `git archive` from `7e014d00e2faf4ccf54e68b03dc1bb9c261463d3` (tree `60bd84ebddfb12dd6d2d0c6ab9c6d9087b716590`). It is a disposable local fixture, not an authoritative checkout, branch, or publication target, and remains preserved for later review.
- Relevant overlay identities: historical fixture `vnext_publication_repair_test.go` plus its F-04 classification seam SHA-256 `49abfa4eb0b0ffc8f34b4d4c5a8937bef0ac0b1a63a9317cc4fde0736175a11b`; bounded child control `cp11_f04_original_behavior_test.go` SHA-256 `d8faa2d4fcad83ea7213366efd309b985cdd118b3d28e3d1e23becc4528dd363`. The fixture also contains the earlier-prepared but **unrun** F-01 overlay; it is not used as F-04 evidence.
- Pre-marker command/result retained as history: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestCP11F04OriginalSnapshotOracleFailuresAreBounded$' -v` → exit 0 / `ok polymetrics.ai/cmd/connectorgen 5.571s`. It confirmed the FIFO block and mixed-observation children but did not yet carry the parent-visible boundary marker, so it is not credited as the strengthened FIFO witness.
- Strengthened command/result: the same command after adding the marker → exit 0 / `ok polymetrics.ai/cmd/connectorgen 5.565s`; the outer test passes only after observing the intended original failures in independently bounded children.
- FIFO child: after old `Lstat` classified regular A, the child wrote the exact parent-visible marker `post-classification FIFO replacement installed` immediately after installing FIFO B and before the old `os.ReadFile(path)`. The parent observed that marker before beginning its 350 ms bounded hang interval, then issued `SIGKILL` and boundedly reaped the still-blocked child: `ORIGINAL F-04 witness: child classified A and installed FIFO B before the old pathname open blocked` and `ORIGINAL F-04 observed: regular-to-FIFO replacement blocked the old pathname oracle until the bounded harness killed and reaped it`. The marker distinguishes the actual check→replacement→open failure from slow child startup.
- Symlink child: the old oracle retained regular A metadata but followed B to bytes outside that pathname, then emitted `ORIGINAL F-04: pathname oracle mixed A classification with symlink replacement B bytes`.
- Directory child: it retained A directory metadata then recursed by pathname into B and emitted `ORIGINAL F-04: pathname oracle mixed A classification with directory replacement B bytes`.
- Each child records distinct A/B inode evidence before emitting its failure. The harness bounds startup-to-observation at 350 ms for FIFO after the marker and one second for nonblocking cases, arms reaping immediately after `Start`, and bounds cleanup reaping at one second. An initial F-04 harness run is retained as a failed fixture setup: it observed the FIFO hang and child mixed-observation output but exited 1 because its parent looked for raw bytes in JSON's base64 payload. The fixture assertion was corrected; the final command above is the authoritative bounded observation. No active production code, provider, credential, database, shared daemon, or protected checkout was touched.

## F-08 original-harness negative control

- Same exact baseline fixture. Its dedicated fixture control `cp11_f08_original_behavior_test.go` has SHA-256 `279af039828da378f3883b8bf47808d9acde372515974b769f3ca3681125aba2`.
- Command: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestCP11F08OriginalHarnessLeavesChildOnReadinessFailure$' -v` → exit 0 / `ok polymetrics.ai/cmd/connectorgen 4.989s`, because the parent passes only after observing the original failure and safely cleaning it up.
- The fixture's inner `unarmed-readiness-failure` process starts the exact `sleeper` child, writes **that child PID** to the parent-owned file, then fails before installing a cleanup owner. The outer `TestCP11F08OriginalHarnessLeavesChildOnReadinessFailure` reads that exact PID, proves it is live with `kill(pid, 0)`, and logs `ORIGINAL F-08 observed: exact child PID 36295 remained live after the unarmed readiness failure`.
- The original unarmed fixture does not get credit for reaping its child. The outer test owns a PID-specific remediation attempt: it registers a cleanup guard for that exact PID before the liveness check, directly sends `SIGKILL` only to that PID, and polls for that same PID's `ESRCH` before setting its `cleaned` guard. The literal historical log label `ORIGINAL F-08 bounded cleanup reaped exact child PID 36295` is imprecise: this outer process did not directly `Wait` for the orphaned grandchild, so the evidence is an exact-PID kill/absence observation, not proof of direct reap ownership. OS/init cleanup is not attributed to the outer harness. Cleanup `SIGKILL` is not a signal-success witness.

## Group 1 active GREEN matrix

- Tested active uncommitted source/test identity: SHA-256 `e47f98a1691fbf1ea262b0571cbe5be28b96c42f0f197bd2e8ed3c7e5f6815b0` for the exact binary diff across `vnext_publication_repair.go`, `vnext_publication_repair_test.go`, `vnext_publication_test.go`, and `vnext_publication_observation_test.go`. This retains the documented premature F-01 patch but does not accept it before Group 2's baseline regression.
- Command: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextPublicationTreeSnapshotRefusesAtoBReplacement|TestVNextPublicationTreeSnapshotRetainsNestedRegularBytes|TestConnectorgenMainPreservesNonConsumingSignalTermination|TestConnectorgenMainLockRenderSignalCancellationAndRetry|TestVNextPublicationBoundedChildReapsWithheldFIFOReadiness|TestVNextGenerationPublisherResumesInterruptedBaseAuthorityPreparation|TestVNextPublicationFIFOReaderRefusesBeforeBlockingOpen)$' -v` → exit 0 / `ok polymetrics.ai/cmd/connectorgen 17.659s`.
- F-04: child witnesses logged distinct A/B identities before the descriptor-bound open in each regular→FIFO, regular→symlink, and directory A→B case. The active helper then refused FIFO as nonregular, refused the symlink at the no-follow file boundary without the B secret appearing in output, and rejected directory B when its opened descriptor disagreed with A. A nested regular-file positive retained its bytes. Existing interrupted-base-authority `--check` and four FIFO refusal/preservation caller cases also passed using the new oracle.
- F-08: every child in the two signal tests starts through `vNextPublicationStartBoundedChildForTest`, which starts its direct `command.Wait` goroutine and cleanup owner immediately after a successful `Start`; no test path invokes a raw unbounded `Wait`. The lock-render test logged child PID `68022` holding the connector-directory descriptor while the parent lock was held before sending real `SIGINT`, then asserted exit 1/context cancellation, no success output, unchanged complete selected/control/authority/generation snapshot, release, and retry. The non-consuming child received real `SIGINT`/`SIGTERM` variants and exited as signaled. The withheld FIFO readiness/nonterminating child returned within 100 ms, was killed by its immediate owner, and logged direct `Wait` completion; that direct Wait proof is distinct from the original orphan's PID-absence observation.

Group 1 F-04/F-08 evidence is now complete for execution purposes. The same active diff still contains the separately preserved, prematurely started F-01 production patch; it remains unaccepted until the Group 2 original-behaviour fixture runs and its complete capture/candidate/mutating matrix is recorded.

## 2026-09-06 — CP11 e77 seven-finding Group 1 repair (steers 051/053)

- **F-04-R original negative control preserved before helper replacement:**
  [GROUP1-F04-ORIGINAL-CONTROL.md](GROUP1-F04-ORIGINAL-CONTROL.md) retains
  the complete command/output, parent-visible A/B boundary, exact old-helper
  seam, hashes, reconstructable helper hunk, and full 197-line bounded child
  snapshot. It ran at parent `67ff7a7ababdbd4d91d9a0b5f9b9d6705fb3c189`:
  `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestCP11F04ROriginalWitnessCanMixPathObservations$' -v` exited 0 in 4.5s
  (`ok ... 1.199s`) only because it observed the FIFO pathname-read block and
  A-classification/B-payload symlink and directory substitutions. It is an
  original-behaviour oracle witness, never repair GREEN; it was not rerun.
- **F-04-R repair:** `vNextPublicationFileWitnessForTest` now retains a
  no-follow/nonblocking regular-file descriptor via the existing
  `vNextPublicationDirectory` path; its `FileInfo` and `io.ReadAll` bytes come
  from that same descriptor. `vNextPublicationDirectoryWitnessForTest` retains
  the opened directory descriptor and reads `metadata.json` relative to it.
  A test-only after-open callback forces regular A→FIFO B, regular A→symlink B,
  and directory A→B after the actual descriptor/open classification. All three
  return retained A identity/bytes without B reads or FIFO blocking.
- **F-08-R negative and repair controls:** the test child reaches the old lsof
  directory-open observation and a test-owned pre-flock gate while the parent
  holds the exact directory lock. Before the parent releases that gate, no
  contention acknowledgement exists; directory-open is therefore explicitly
  disqualified as readiness. `LockContention` is a nil-by-default test-only
  callback invoked only after the real `LOCK_NB` returns `EWOULDBLOCK`/`EAGAIN`
  on the retained directory identity. The child calls the real `runMain`
  lock-render path, so real SIGINT and SIGTERM occur only after the parent has
  compared that acknowledgement to its held descriptor identity. The old broad
  non-consuming signal assertion now compares `WaitStatus.Signal()` with the
  sent SIGINT/SIGTERM; cleanup SIGKILL remains excluded.
- **Focused GREEN command:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestCP11F04RWitnessRetainsOpenedObjectAcrossReplacement|TestConnectorgenMainSignalsOnlyAfterExactLockContention|TestConnectorgenMainPreservesNonConsumingSignalTermination)$' -v` exited 0 in 12.4s
  (`ok polymetrics.ai/cmd/connectorgen 9.198s`). Its selector covers FIFO,
  symlink, directory, real-main SIGINT, real-main SIGTERM, exact default
  signal status, no-success/cancellation, selected/control/authority/generation
  preservation, direct waits and ordinary retry. It neither claims provider,
  power-loss, no-mistakes, release, nor CP11 acceptance.
- **Focused race command:** the same selector under `go test -race -count=1
  -timeout 20m ./cmd/connectorgen ... -v` exited 0 in 20.2s (`ok
  polymetrics.ai/cmd/connectorgen 15.924s`). The first race run exposed only a
  test-binary post-boundary completion bound: successful FIFO/directory child
  outputs could arrive after its one-second normal-wait budget. The helper was
  unchanged; the owner widened that normal (non-cleanup) bound to three seconds,
  retained immediate cleanup ownership, and then reran both race and normal.
- **Exact tested source/test identities:**
  `vnext_publication.go` `b67349746761d31faa789b300a4284dce4974343f2123bf8000f3534dc626892`;
  `vnext_lock_cli.go` `962bc07c21098c2826d7c7f3770852ab62d4f70dd3bc56ea2f98dba7bc5ee84d`;
  `vnext_publication_durable_matrix_test.go` `532d0cc07bf0b0754d4a557303a7bf83b5d0e64259ea69213db68a6c81d85d5c`;
  `vnext_publication_test.go` `2927c544fb4e4995fff5843f5264c14108b6f5fcfd4c6e80b904bb5f14679390`;
  `vnext_publication_witness_observation_test.go` `b42a9a03140e8c8cac27d5661ae62b7f408241e26adf3b9dd44ba4b87186316a`;
  `vnext_publication_lock_contention_test.go` `12b2c177132404c24df7a4a3ea0d196a883a6813d836a2a347ebfcc659fb08ab`.
