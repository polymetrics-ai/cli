# Plan — retained-source mapping-only bridge

## 1. Red: command and safety boundaries

Create focused tests for a missing `retained-source-mapping` command and pure construction path. The tests must assert that all eight retained v2 locks produce a 2,340-ID source-only contract/report, while normal `canonical_evidence` admission remains false.

Create negative tests for malformed/missing source contracts and fragments, a conflicting provider `operationId`, GraphQL inventory, mixed or invalid matrix representations, missing/duplicate/unknown source IDs, invalid lanes/states, and any attempt to claim implementation or non-source artifact binding.

## 2. Green: reuse retained parsing without importing runtime behavior

Add a small `cmd/connectorgen/retainedsourcemapping.go` module. It will:

1. Parse command options and constrain connector and definitions paths.
2. Read and verify the frozen Batch R1 cohort membership and digest using the existing cohort checker.
3. Read one connector's source lock with existing strict decoding.
4. Apply a new, narrow retained-mapping eligibility predicate; it is separate from and must not change normal canonical-evidence import admission.
5. Reassemble the provider document through existing generic source-contract logic, parse/import it only in memory, and validate exact locked REST identity projection.
6. Strictly decode either generic source-lane matrix wire shape, convert every exact source ID into a retention-only seven-lane contract, and validate/reconcile that contract using existing `internal/connectors` logic.
7. Emit deterministic source-only JSON/report data. No filesystem output is allowed.

## 3. Refactor: deterministic encoding and direct safety proofs

Sort source IDs and lanes, use a single fixed accounting partition priority only for contract denominator bookkeeping, and leave every applicable lane as an overlay. The report must call this an accounting partition, not a runtime lane choice.

Run `gofmt`, focused/race/vet tests, `jq`, cohort check, `agentcontractgen check`, and diff check. Record any baseline failures separately rather than modifying unrelated connector artifacts.

## Non-goals

- No source lock, source lane matrix, source descriptor, root enabled contract, operations, writes, streams, schemas, API surface, CLI surface, transport, engine, runtime, certification, Atlas, or connector JSON change.
- No provider HTTP fetch, credential read, or materialization.
- No call to `runSourceImport`, `runSourceMaterialize`, source projection, or engine bundle load.
- No automatic operational contract promotion.

## Commit checkpoints

1. Planning evidence and red suite are staged only after the explicit red command/test proof.
2. One green commit after all focused verification passes.
3. Candidate-only push after local green gates. No integration/rebase/merge.
