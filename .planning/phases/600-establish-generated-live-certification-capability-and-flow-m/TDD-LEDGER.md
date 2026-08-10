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

## Proof-bearing certification refinement — Captain directive 2026-08-10

| Slice | Red evidence | Green evidence | Refactor evidence | Status |
|---|---|---|---|---|
| Embedded proof required before a live claim | `go test -timeout 20m ./cmd/connectorgen -run '^TestCertificationRejectsProoflessAcceptedLiveEvidence$' -count=1` failed: `validateAcceptedEvidence() error = nil, want proof-bearing live evidence rejection`. | Pending | Pending | red recorded |
| Prepared-value redaction before persistence | Pending: add sanitizer/unsafe-proof tests before implementation. | Pending | Pending | planned |
| Artifact-first proof validation | Pending: add checker-order test before implementation. | Pending | Pending | planned |

## Pair-flow matrix — Plan 02

| Slice | Red evidence | Green evidence | Refactor evidence | Status |
|---|---|---|---|---|
| Endpoint roles and pair identity | Pending: roles/reasons and source+destination key tests | Pending | Pending | planned |
| Durable destination and readback proof | Pending: API destination and missing-readback tests | Pending | Pending | planned |
| Compact pair artifact and final certification | Pending: pair-set resolver and aggregation test | Pending | Pending | planned |

## Commands

Focused tests will use `go test -timeout 20m ./cmd/connectorgen`; generator
checks will use `go run ./cmd/connectorgen certification-matrix --check`.
No test may contact a real provider or use real credentials.

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
