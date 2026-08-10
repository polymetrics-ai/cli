# Phase 600 TDD ledger

**Issue:** #3984  
**Execution mode:** Inline single-worker fallback; no GSD role was spawned by
the canonical delivery contract.

## Capability matrix — Plan 01

| Slice | Red evidence | Green evidence | Refactor evidence | Status |
|---|---|---|---|---|
| Source-derived inventory and executor status | `go test -timeout 20m ./cmd/connectorgen -run TestCertification -count=1` failed to compile: `discoverFunctionKinds` and the matrix builder do not exist. | `go test -timeout 20m ./cmd/connectorgen -run 'TestCertification' -count=1` passed (92.786s); it proves AST inventory, real executor annotations, engine pointer methods, and PostgreSQL/MySQL Write stubs are applicable but `implemented=false`. | Reflection now retains pointer receiver method sets while source inspection resolves the concrete receiver. | green |
| Strict cells and certification completeness | Same RED run: `certificationCell`, `certificationComplete`, `notApplicableReason`, and strict evidence validation do not exist. | Same focused GREEN run passed: missing live proof blocks completion; generic N/A and malformed evidence are rejected. | Generic engine Write without a declared action is a named N/A; concrete native database Write stubs remain applicable false cells. | green |
| Deterministic artifact/drift gate | Same RED run: `checkGeneratedArtifact` does not exist. | `make connectorgen-certification-matrix` passed: 556 connectors, 0 capability-complete. | JSON is ordered slices and the gate compares exact bytes. | green |
| GitHub GraphQL executor classification | `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestCertificationMatrixDistinguishesGitHubGraphQLAndLocalGitExecutors$' -v` failed: GitHub `operation:graphql_query` reported `implemented=false`. | Same command passed (27.922s): GitHub `graphql_query` is implemented and the genuinely absent `local_git` executor remains false. | The existing direct-read executor annotation now names its supported fixed GraphQL query branch; deeper call-graph proof remains explicitly deferred. | green |

## Proof-bearing certification refinement — Captain directive 2026-08-10

| Slice | Red evidence | Green evidence | Refactor evidence | Status |
|---|---|---|---|---|
| Embedded proof required before a live claim | `go test -timeout 20m ./cmd/connectorgen -run '^TestCertificationRejectsProoflessAcceptedLiveEvidence$' -count=1` failed: `validateAcceptedEvidence() error = nil, want proof-bearing live evidence rejection`. | `go test -timeout 20m ./cmd/connectorgen -run '^TestCertification' -count=1` passed; proofless, malformed, narrow-scope, and unsafe evidence are rejected. | The accepted record embeds only a fingerprinted proof; a pass claim or external pointer cannot promote a cell. | green |
| Prepared-value fingerprinting before persistence | The original proof contract had no local-salt writer or prepared-value sanitization boundary. | Same focused GREEN command passed `TestCertificationSanitizesPreparedValuesBeforeProofPersistence` and `TestCertificationEvidenceWriterUsesRepositoryLocalSaltBeforePersistence`. | HMAC-SHA-256 uses a local 0600 repository salt; raw prepared values are never serialized and unknown transcript values are fingerprinted. | green |
| Proof record identifiers and filenames | Security review found that an unsafe provider or evidence filename could otherwise reach a typed record before later matrix validation. | Same focused GREEN command passed TestCertificationRejectsUnsafeEvidenceIdentifiers. | Provider, connector/scope identifiers, and direct evidence filenames are restricted to safe identifiers before the writer opens a file. | green |
| Artifact-first proof validation | The first strict proof test above exposed that invalid evidence could otherwise be read only after matrix generation. | Same focused GREEN command passed `TestCertificationArtifactProofValidationPrecedesCodeDrift`. | Check mode validates accepted evidence before comparing regenerated code-derived output, so a proof cannot become trusted merely because a matrix file exists. | green |

## Pair-flow matrix — Plan 02

| Slice | Red evidence | Green evidence | Refactor evidence | Status |
|---|---|---|---|---|
| Endpoint roles, workflows, and pair identity | The original capability-only RED compiler failure in the transcript below established the missing matrix contracts. The captain added flow/workflow scope during the same checkpoint; no separate flow-only RED transcript was captured, recorded here as an inline-GSD limitation rather than invented evidence. | `go test -timeout 20m ./cmd/connectorgen -run '^TestCertification' -count=1` passed `TestCertificationWorkflowWithoutEvidencePreventsCompletion` and `TestCertificationFlowPairAllowsGitHubToItselfThroughWarehouse`. | Workflow kinds come from narrow source annotations, and pair keys include source, warehouse mediator, destination, and flow kind. | green |
| Stable sync-mode scoreboard | Same capability-only RED baseline; no separate scope-expansion RED transcript was captured. | Same focused GREEN command passed `TestCertificationDiscoversStableWarehouseFacingSyncPrimitives`, `TestCertificationSyncModeDatabaseWriteStubIsNotImplemented`, and `TestCertificationChangeCaptureRequiresDatabaseReadIntoWarehouse`. | Modes derive from `synccontract.AllModes()`; the four primitive identities are stable and every impossible mode/primitive combination has a named reason. | green |
| Durable destination, readback proof, and final certification | Same capability-only RED baseline; no separate scope-expansion RED transcript was captured. | Same focused GREEN command passed `TestCertificationFlowEvidenceRequiresRoundTripProof`; full matrix generation produces zero final certifications. | A flow requires one real-pm round trip with independent warehouse and destination readbacks; delivery guarantees remain separate from working state. | green |

## Commands

Focused tests use `go test -timeout 20m ./cmd/connectorgen -run
'^TestCertification' -count=1`; generator checks use `go run
./cmd/connectorgen certification-matrix --check`. No test may contact a real
provider or use real credentials.

## Red command transcript — 2026-08-10

```
FAIL polymetrics.ai/cmd/connectorgen [build failed]
cmd/connectorgen/certificationmatrix_test.go:11:16: undefined: discoverFunctionKinds
cmd/connectorgen/certificationmatrix_test.go:60:17: undefined: buildCapabilityMatrix
cmd/connectorgen/certificationmatrix_test.go:77:10: undefined: certificationCell
cmd/connectorgen/certificationmatrix_test.go:96:9: undefined: validateCertificationCell
```

This is the intended red state: the tests were added before any matrix
production implementation. The next green implementation must make these
observable contracts pass without requesting credentials or performing provider
network I/O.
