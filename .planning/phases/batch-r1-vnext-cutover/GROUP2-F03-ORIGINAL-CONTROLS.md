# CP11 Group 2 F-03 original-behaviour controls

## Scope and provenance

This record preserves the pre-repair controls required by the coordinated
F-03-A/F-03-B/F-03-C ownership/error group. They ran after the Group 1 commit
`69246943bdcb5c3cdc39c08a7cf1664f4af811aa` and before any Group 2 production
repair. The narrow seams below are nil by default and make no production
decision: they retain real opened descriptors, perform each real `Close`, and
then return their named test completion cause. They exist solely to exercise
the current error/cleanup paths deterministically.

- `cmd/connectorgen/vnext_publication_group2_original_test.go` SHA-256:
  `48456f317c6ae2cc68b3ee085324f50f47946b288bf93a45aeceb5e013fa3890`
- `cmd/connectorgen/vnext_publication.go` SHA-256:
  `f0986947f04933632179504e738399c5bfc7f4c7d2b05b35a8ceea5f7cd899cb`
- `cmd/connectorgen/vnext_publication_dir.go` SHA-256:
  `ed2ef6fefb3b8450246b86688114fe5f162c1a8a0ebac55d3bdba268a2d993a0`
- `cmd/connectorgen/vnext_publication_repair.go` SHA-256:
  `d6ecb3510cb244beb490e7400ee48b0149a9267c8db86417eba8ec771dcbbda7`

The source identity above is retained in the immediately following local
test-control commit. This document does not relabel a defect-reproduction pass
as GREEN.

## Command and observed result

```text
go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestCP11F03AOriginalPostRecordFailureStrandsPreparedAuthority|TestCP11F03BOriginalCompoundCausesAreNotInspectable|TestCP11F03COriginalTemporaryCleanupDeletesReplacementB|TestCP11F03COriginalQuarantineCleanupDeletesReplacementB)$' -v

=== RUN   TestCP11F03AOriginalPostRecordFailureStrandsPreparedAuthority
--- PASS: TestCP11F03AOriginalPostRecordFailureStrandsPreparedAuthority (0.58s)
=== RUN   TestCP11F03COriginalTemporaryCleanupDeletesReplacementB
--- PASS: TestCP11F03COriginalTemporaryCleanupDeletesReplacementB (0.35s)
=== RUN   TestCP11F03COriginalQuarantineCleanupDeletesReplacementB
--- PASS: TestCP11F03COriginalQuarantineCleanupDeletesReplacementB (0.80s)
=== RUN   TestCP11F03BOriginalCompoundCausesAreNotInspectable
=== RUN   TestCP11F03BOriginalCompoundCausesAreNotInspectable/definitions_and_connector_close
=== RUN   TestCP11F03BOriginalCompoundCausesAreNotInspectable/missing_control_parent_close
=== RUN   TestCP11F03BOriginalCompoundCausesAreNotInspectable/staged_file_sync_and_close
--- PASS: TestCP11F03BOriginalCompoundCausesAreNotInspectable (0.04s)
    --- PASS: TestCP11F03BOriginalCompoundCausesAreNotInspectable/definitions_and_connector_close (0.00s)
    --- PASS: TestCP11F03BOriginalCompoundCausesAreNotInspectable/missing_control_parent_close (0.00s)
    --- PASS: TestCP11F03BOriginalCompoundCausesAreNotInspectable/staged_file_sync_and_close (0.04s)
PASS
ok  	polymetrics.ai/cmd/connectorgen	2.687s
```

The command exits zero only because each control asserts the demonstrated
unrepaired defect; it is RED evidence, not a repaired acceptance result.

## Observed defects

- **F-03-A:** after a real durable `prepared.json` record has closed, the
  injected post-record frontier returns an error. The original unwind removes
  the referenced anchors and leaves prepared-only authority; a fresh publisher
  refuses recovery for its missing anchor.
- **F-03-B:** real definitions-root and connector-root closes return distinct
  causes, but only the latter remains inspectable. A real failed missing-control
  open plus parent close loses the `fs.ErrNotExist` cause, so the current
  consumer reports a found/error instead of distinguishing compound absence.
  A real writable staged file reaches the before-sync failure frontier and its
  real close completion; only the primary cause remains inspectable. The
  source-audited sibling table remains mandatory for the coordinated repair;
  these are bounded executable representatives, not a claim that every sibling
  was independently faulted here.
- **F-03-C:** after actual temporary/quarantine A has opened, the control moves
  A aside and installs an empty directory B. The current cleanup re-observes B
  and deletes it. The temporary test also retains A's real `control` blocker;
  the quarantine test retains the stale generation so the result is not
  overstated as arbitrary deletion.

All controls use disposable local filesystem fixtures only. They do not access
provider credentials, a customer database, the protected `.cache/` fixture, or
the shared daemon.
