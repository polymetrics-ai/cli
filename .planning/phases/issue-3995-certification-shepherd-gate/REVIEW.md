# REVIEW — issue #3995 shared connector-certification Shepherd gate

## Method

The issue-named phase has no registered numbered GSD adapter phase, so `code-review` was executed
inline under the documented fallback. Review covered the canonical contract, strict decoder,
evaluator, evidence binding, projection renderer/checker, generated Claude/Codex/Pi/OpenCode
files, tests, and changed-path scope. No provider, credential, connector-bundle, transport, or
`cmd/connectorgen/certification*.go` path is changed.

Automated review evidence before this report:

- `go test -timeout 20m ./internal/agentcontract -count=1`
- `go test -timeout 20m ./cmd/agentcontractgen -count=1`
- `go vet ./...`, `go build ./cmd/pm`, `make lint`, and
  `go run ./cmd/agentcontractgen check`
- generated-projection equivalence/drift tests and the individual repository gates recorded in
  `VERIFICATION.md`

## Findings and dispositions

| ID | Severity | Finding | Disposition | Evidence |
| --- | --- | --- | --- | --- |
| R-1 | medium | The first proof consumer only required `body.value` to be valid JSON. A forged accepted sidecar could therefore retain a raw scalar rather than a repository-salted fingerprint. | Fixed. `validateHTTPBody` now mirrors the proof boundary: `none`, `opaque`, and `json` forms are versioned/typed and every JSON scalar must be a fingerprint (or `null`). Header/query/statement values also accept only fingerprint sequences. | `TestCertificationGateRejectsUnredactedProofBody` passes; `TestCertificationGateMatchesSemanticallyEquivalentEvidenceProof` preserves safe formatting tolerance. |
| R-2 | low | Raw `json.RawMessage` whitespace could make a semantically identical matrix pointer and sidecar appear unequal. | Fixed. Pointer/record proof comparison serializes the typed proof into canonical JSON before comparison. | `TestCertificationGateMatchesSemanticallyEquivalentEvidenceProof` passes. |

## Verdict

PASS after correction round 1 of 5. There are no open actionable review findings. The remaining
external dependency is #3989: its eventual proof-schema version must be integrated explicitly;
the present gate halts unknown proof versions rather than inventing fields.

## Correction tracking

The proof-consumer fixes are tracked as child issue [#4024](https://github.com/polymetrics-ai/cli/issues/4024)
under #3995, with `Refs #3988`. It owns both R-1 fingerprint-only proof consumption and R-2
canonical semantic JSON comparison. The production remediation is commit `842f1c271`; the
verification record is `d511186bc`. The issue is deliberately open pending this child branch's PR
review and the parent acceptance flow.

## Correction round 2 re-review

| ID | Severity | Finding | Disposition | Tracking |
| --- | --- | --- | --- | --- |
| C2-1 | error | Empty proof fingerprint sequences could satisfy proof validation. | Fixed: empty sequences now halt; the sidecar and pointer are both rejected. | #4024 |
| C2-2 | error | JSON object-member order caused semantically equal proof bodies to mismatch. | Fixed: typed proofs are decoded to semantic JSON values before comparison. | #4024 |
| C2-3 | error | Missing sync-mode/primitive or flow-pair cells, and inconsistent connector rosters, could be omitted before a certified status proceeded. | Fixed: topology and connector identity are validated across capability, flow, and status artifacts. | #4028 |
| C2-4 | error | Symlink ancestors and non-regular evidence records could escape the supplied input root. | Fixed: the shared root-bound reader rejects both before decoding. | #4028 |
| C2-5 | warning | A producer-valid false delivery guarantee with a named limitation halted in the consumer. | Fixed: the consumer mirrors the producer's named-limitation rule. | #4028 |
| C2-6 | error | The gate lacked an executable production transition surface. | Fixed: canonical projections now carry the exact read-only `agentcontractgen certification-gate` argv. | #4030 |

The six reported correction tests were written before their production edits. The one permitted
focused verification then passed:

```sh
go test -timeout 20m ./internal/agentcontract ./cmd/agentcontractgen -count=1
```

The command reported `ok` for both packages. This review phase intentionally did not run full
repository tests, lint, CI, a PR action, or any outer delivery gate.

## Current verdict

PASS after correction round 2 of 5. The checked-in GitHub artifact deterministically returns
`RETRY` with `capability/github/capability:check/live_evidence`; malformed, escaped, and
producer-invalid inputs halt, while a complete producer-valid fixture can proceed. Remaining
validation, branch/PR work, and human gates remain owned by the outer executor.

## Correction round 3 re-review

| ID | Severity | Finding | Disposition | Tracking |
| --- | --- | --- | --- | --- |
| C3-1 | error | The consumer accepted a caller-supplied flow-kind inventory, allowing omitted, added, or remapped producer flow kinds to change pair coverage. | Fixed: `connectorgen` and the consumer import one exact four-kind catalog and reject any inventory or mapping drift. | #4028 |
| C3-2 | error | Completion fields, connector statuses, and baseline aggregates could be forged independently of structurally valid, matched evidence. | Fixed: all completion reports, statuses, and aggregates are recomputed before the target evaluation and any disagreement halts. | #4028 |
| C3-3 | error | An omitted gate root could select a parent repository and evaluate it instead of failing closed. | Fixed: the transition command accepts only one explicit canonical absolute non-symlink root with a non-symlink contract. | #4030 |
| C3-4 | error | Invalid live-evidence pointers lost their trustworthy cell/evidence coordinates in a generic matrix halt. | Fixed: malformed pointers halt with their trusted cell ID, a safe record ID only when canonical, and the fixed `invalid_pointer` reason. | #4028 |

Focused correction-round verification passed for `internal/agentcontract`, `cmd/agentcontractgen`,
certification-generator coverage, projection sync/check, and the four protected command
transitions. It did not run an outer pipeline phase.

## Current verdict after correction round 3

PASS after correction round 3 of 5. The checked-in GitHub artifact retains deterministic `RETRY`
with `capability/github/capability:check/live_evidence`; a complete producer-valid fixture
`PROCEED`s, while malformed, mismatched, and escaped input halts. Remaining delivery phases,
branch/PR work, and human gates remain owned by the outer executor.

## Correction round 4 re-review

| ID | Severity | Finding | Disposition | Tracking |
| --- | --- | --- | --- | --- |
| C4-1 | error | Round three directly changed `cmd/connectorgen/certification*.go`, violating the forbidden-path constraint. | Fixed: the producer path is restored unchanged; an `agentcontractgen`-generated catalog now projects `flow-matrix.json` for the consumer, and sync/check detect its drift. | #4028 |
| C4-2 | error | A flow override could promote immutable producer facts to a green resolved cell. | Fixed: an applicable override must preserve applicable, declared, implemented, fixture, and not-applicable facts; only fully bound live-evidence facts may differ. | #4028 |
| C4-3 | error | Resolved evidence validation masked raw base pair-set pointers. | Fixed: every raw pair-set cell and each override cell is bound before resolved reports are derived; safe missing, mismatched, and wrong-coordinate records halt with exact coordinates. | #4028 |
| C4-4 | error | Semantic proof comparison rounded distinct JSON numbers through `float64`. | Fixed: proof decoding uses `UseNumber`, so distinct 64-bit `original_bytes` values halt as a mismatch. | #4024 |

The focused Round 4 package, producer-matrix, catalog sync/check, and changed-path checks passed.
No provider, credential, network, evidence-creation, CI, PR, or outer pipeline action ran.

## Current verdict after correction round 4

PASS after correction round 4 of 5. The checked-in GitHub artifact remains deterministic `RETRY`
with `capability/github/capability:check/live_evidence`; complete valid fixtures proceed; malformed,
mismatched, escaped, overridden-invalid, and precision-mismatched inputs halt.

## Correction round 5 re-review

| ID | Severity | Finding | Disposition | Tracking |
| --- | --- | --- | --- | --- |
| C5-1 | error | A flow pair could remain green after an endpoint role became not-applicable or lost its declared/implemented facts, and a not-applicable code/reason could drift independently. | Fixed: each raw pair coordinate now derives applicability, declared/implemented conjunctions, and the exact not-applicable code/reason from its canonical flow-kind endpoint roles before evidence binding or report derivation. | #4028 |
| C5-2 | warning | The generated catalog was the only source file defining its package API, so sync could not compile to restore an absent catalog. | Fixed: stable non-generated type/accessor code fails closed for missing, empty, or invalid data; `flow_gen.go` now carries only generated data registration and sync recreated it from the matrix. | #4028 |

The focused Round 5 catalog-bootstrap, consumer, unchanged-producer, projection, and changed-path
checks passed. No provider, credential, network, evidence-creation, live connector, transport,
warehouse, delivery-contract, generated worker, CI, PR, or outer pipeline action ran.

## Current verdict after correction round 5

PASS after correction round 5 of 5. The GitHub baseline remains deterministic `RETRY` with
`capability/github/capability:check/live_evidence`; complete valid fixtures proceed; all reviewed
role/pair, catalog, malformed, mismatched, escaped, overridden-invalid, and precision-mismatched
inputs halt deterministically.
