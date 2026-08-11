# SUMMARY — issue #3995 shared connector-certification Shepherd gate

## Delivered boundary

The canonical delivery contract now declares a versioned, read-only
`connector-certification-shepherd` gate. `internal/agentcontract` decodes the generated
capability, flow, status, and evidence JSON with strict schemas, evaluates one connector and
transition deterministically, and returns `PROCEED`, `RETRY`, or `HALT` with stable failure
coordinates. `EnforceCertificationGate` blocks `integrate_sub_pr`, `accepted`, `ready_parent`,
and `human_ready` unless the decision is `PROCEED`.

The single canonical projection registry now generates and verifies Claude, Codex, Pi, and
OpenCode worker inputs. The shared projection block is byte-equivalent across all four harnesses;
OpenCode is registered rather than handled as a one-off adapter.

## Acceptance proof

- The initial RED test loaded the generated GitHub baseline and required the unavailable evaluator
  to return `RETRY` with `capability/github/capability:check/live_evidence`; its undefined-symbol
  failure is recorded in `TDD-LEDGER.md`.
- The implemented gate returns that exact `RETRY` for the current zero-certified GitHub baseline
  and returns `PROCEED` for a complete temporary generated fixture.
- Isolated binding defects preserve their exact cell/evidence coordinates. Invalid schemas,
  unknown fields, omitted adapter-local inputs, unmatched/malformed sidecars, and unredacted proof
  bodies halt closed.
- Evaluation is tested as read-only. It cannot create evidence, access credentials, invoke a
  provider, or change `cmd/connectorgen/certification*.go`.

## Delivery state

All local verification and manual review gates are green. The two fixed manual-review findings are
tracked by child [#4024](https://github.com/polymetrics-ai/cli/issues/4024) in correction round 1
of 5: the remediation commit is `842f1c271` and the verification record is `d511186bc`.
`no-mistakes` and child-PR publication are the remaining delivery steps.

#3984 is consumed at `815dc1ab65380e03f6e0c078ba36030baaec21ea`. #3985's concrete canon is
available at `da7747a796049601a179a97c025bfb05f011f1e8`, although #3985 remains formally open.
#3989 remains a held integration gate: this change accepts only the declared proof-schema version
and halts unknown versions rather than inventing future proof fields.

## Correction round 2

The correction round adds empty-fingerprint rejection, semantic proof JSON comparison, complete
matrix/connector topology validation, root-confined artifact loading, and producer-equivalent named
delivery limitations. The canonical contract now renders the exact read-only
`agentcontractgen certification-gate` argv for every registered harness, so protected transitions
have an executable gate surface rather than an uncalled helper.

The focused review command
`go test -timeout 20m ./internal/agentcontract ./cmd/agentcontractgen -count=1` passed. The current
GitHub artifact still returns deterministic `RETRY` with
`capability/github/capability:check/live_evidence`; a complete producer-valid fixture may return
`PROCEED`. Correction tracking is split deliberately: #4024 covers proof comparison/fingerprint
consumption, #4028 invalid artifacts, and #4030 protected transitions.

## Correction round 3

The producer and consumer now import one exact four-kind flow catalog, so a certification artifact
cannot omit, add, or remap flow coverage. The consumer validates every structurally valid
live-evidence binding before deriving capability/workflow/sync/flow completion, final connector
status, and baseline aggregates; report drift halts deterministically. Pointer failures retain a
trusted cell coordinate and only expose a safe canonical evidence record.

The read-only `agentcontractgen certification-gate` boundary now requires an explicit canonical
absolute non-symlink root and non-symlink canonical contract for all protected transitions. Missing,
relative, traversing, or symlinked roots, as well as help, block with an encoded `HALT`; only
`PROCEED` can exit zero. Focused correction-round verification passed for the complete
`internal/agentcontract` and `cmd/agentcontractgen` packages, focused certification-generator
tests/check, projection sync/check, and explicit all-four transition command tests.

## Correction round 4

The forbidden round-three direct dependency from `cmd/connectorgen/certification*.go` was removed.
The unchanged producer remains the source of `flow-matrix.json`; `agentcontractgen sync/check`
generates and verifies an importable flow-kind catalog for the consumer. Flow overrides preserve
all immutable base facts, raw and override evidence bind before resolution, and proof comparisons
retain JSON numbers without `float64` precision loss. Focused package, producer-matrix, catalog,
and changed-path checks passed; the GitHub baseline remains the required deterministic `RETRY`.

## Correction round 5

Raw flow pairs now derive their applicability, declared/implemented conjunctions, and exact
not-applicable code/reason from each canonical flow-kind source/destination role before evidence
or report processing. Fixture/live evidence remains evidence-only, and the correction-round-four
override rule remains intact. The catalog type and defensive accessor are stable non-generated
code; its generated file now contains only data registration, so sync compiles and restores an
absent catalog from `flow-matrix.json`. Focused Green checks, including an actual missing-file
bootstrap, the unchanged producer matrix check, catalog check, and empty forbidden-path audit,
passed while the GitHub baseline retained its deterministic `RETRY`. The prior intentional root
instruction also required refreshing stale deterministic render-hash test expectations.
