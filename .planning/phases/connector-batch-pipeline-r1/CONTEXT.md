# Context — connector batch pipeline r1

## Goal

Build the reusable, batch-level control plane for connector authoring without
creating a connector bundle before the shared provenance and output-policy
foundations have landed. The control plane turns a live provider-artifact ledger
into a committed, deterministic batch manifest and later gates every listed
connector independently.

## Supplied evidence

- Live ledger: `/Users/karthiksivadas/karthik-agent-workspace/data/cli-provider-artifact-sweep-r1/ledger.json`.
  It was read directly, rather than through an ignored-path search. At planning
  time it contained 216 records, 176 `done` records, and 87 versioned
  `openapi`/`swagger` records.
- The initial five candidates are chosen from direct, provider-published,
  machine-readable OpenAPI artifacts with a recorded version and retrieval
  date—not from provider-name recognition.
- `cmd/connectorgen` already owns `new`, `gen`, `validate`, `surface-sync`,
  `boundary`, and `ownership`; the batch entry point belongs there rather than
  in a parallel script.
- `TestEveryImplementedCommandPassesRuntimePreflight` in
  `internal/connectors/commandrunner/runner_test.go` is the real runtime guard.
  The batch gate must call `commandrunner.Preflight`, not recreate its rules.

## Dependency state checked live

- #3773 is implemented by PR #3869. Its v2 contract adds a provider-artifact
  table with a stable ID, HTTPS URL, ISO full-date `retrieved_at`, optional
  SHA-256, and endpoint-local artifact/source-URL citations. PR #3869 was open
  with pending required checks at planning time.
- #3870 adds declarable non-redacting output policies. It was open with a
  pending security check at planning time.
- #3868 preserves connector command output and was also still open with full
  verification pending.

At the current-main recheck, #3870 merged as `ee26d20fc` and #3868 merged as
`50deaade9`; #3869 remained open and was the remaining bundle-authoring gate.
No schema, engine, or `internal/connectors/commandrunner/runner.go` change was
authorized in this slice.

## Decisions

- Add `connectorgen batch plan` now. It validates all required ledger evidence,
  chooses or verifies a bounded candidate set, and writes a reviewable manifest
  with no provider call and no bundle mutation.
- Add `connectorgen batch gate` now. It will be usable after authoring to run
  per-connector validation, `surface-sync` drift detection, and the real
  runtime preflight; a connector failure is recorded as a drop, while the
  others continue.
- The artifact-to-bundle stage is deliberately documented and staged, not
  prematurely implemented against a contract that is still landing. It must
  emit only v2 provenance and non-redacting policies after those foundations
  are merged.
- There is no supplied GitHub parent issue number for this task. The branch and
  phase name are the work tracker; no issue is invented or falsely closed.
