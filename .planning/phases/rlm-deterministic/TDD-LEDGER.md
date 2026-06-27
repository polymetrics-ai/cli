# TDD Ledger

Phase: rlm-deterministic

Record failing test evidence before production code for every behavior-adding task.

---

## T1.1 — Spec parsing tests

File: `internal/rlm/spec_test.go`
Status: **red-confirmed**
Run: `GOTOOLCHAIN=auto go test ./internal/rlm/... 2>&1`

All ParseSpec tests fail because the stub returns ErrNotImplemented:

```
--- FAIL: TestParseSpecValid (0.00s)
    spec_test.go:18: expected nil error, got rlm: model backend not implemented (requires Phase 4 approval)
--- FAIL: TestParseSpecMissingName (0.00s)
    spec_test.go:42: error "rlm: model backend not implemented (requires Phase 4 approval)" should mention 'name'
--- FAIL: TestParseSpecNegativeWeight (0.00s)
    spec_test.go:66: error "rlm: model backend not implemented (requires Phase 4 approval)" should mention 'weight'
--- FAIL: TestParseSpecZeroWeight (0.00s)
    spec_test.go:80: zero weight should be allowed, got error: rlm: model backend not implemented (requires Phase 4 approval)
```

---

## T2.1 — Deterministic scoring tests

File: `internal/rlm/deterministic_test.go`
Status: **red-confirmed**

scoreRecords and DeterministicAnalyzer.Run stubs return ErrNotImplemented:

```
--- FAIL: TestDeterminismSameInputSameOutput (0.00s)
    deterministic_test.go:39: first run error: rlm: model backend not implemented (requires Phase 4 approval)
--- FAIL: TestScoringWeightedSum (0.00s)
    deterministic_test.go:75: scoreRecords error: rlm: model backend not implemented (requires Phase 4 approval)
--- FAIL: TestScoringAllZeroWeights (0.00s)
    deterministic_test.go:187: error: rlm: model backend not implemented (requires Phase 4 approval)
--- FAIL: TestScoringEmptyRecordSet (0.00s)
    deterministic_test.go:199: error on empty input: rlm: model backend not implemented (requires Phase 4 approval)
--- FAIL: TestSortingByScoreDesc (0.00s)
    deterministic_test.go:217: error: rlm: model backend not implemented (requires Phase 4 approval)
--- FAIL: TestSortingTiebreakerByRawID (0.00s)
    deterministic_test.go:239: error: rlm: model backend not implemented (requires Phase 4 approval)
```

---

## T3.1 — Materialization tests

File: `internal/rlm/deterministic_test.go` (materialization cases)
Status: **red-confirmed**

DeterministicAnalyzer.Run stub returns ErrNotImplemented:

```
--- FAIL: TestMaterializationWritesNDJSON (0.00s)
    deterministic_test.go:290: Run error: rlm: model backend not implemented (requires Phase 4 approval)
--- FAIL: TestMaterializationPreservesSourceFields (0.00s)
    deterministic_test.go:342: Run error: rlm: model backend not implemented (requires Phase 4 approval)
--- FAIL: TestMaterializationAtomic (0.00s)
    deterministic_test.go:374: Run error: rlm: model backend not implemented (requires Phase 4 approval)
--- FAIL: TestDryRunDoesNotWrite (0.00s)
    deterministic_test.go:405: Run error: rlm: model backend not implemented (requires Phase 4 approval)
```

---

## T4.1 — Fixture backend tests

File: `internal/rlm/fixture_test.go`
Status: **red-confirmed**

FixtureAnalyzer.Run stub returns ErrNotImplemented:

```
--- FAIL: TestFixtureRunReturnsRows (0.00s)
    fixture_test.go:22: Run error: rlm: model backend not implemented (requires Phase 4 approval)
--- FAIL: TestFixtureIgnoresInTable (0.00s)
    fixture_test.go:82: fixture should ignore missing InTable, got error: rlm: model backend not implemented (requires Phase 4 approval)
--- FAIL: TestFixtureDryRun (0.00s)
    fixture_test.go:99: Run error: rlm: model backend not implemented (requires Phase 4 approval)
```

---

## Model stub tests — intentionally GREEN

File: `internal/rlm/model_stub_test.go`
Status: **green** (by design — the stub IS the permanent implementation until Phase 4)

`TestModelStubReturnsNotImplemented` and `TestModelModeString` pass because ModelAnalyzer
is permanently stubbed until a human approves Phase 4.

---

## T7.1 — E2E offline flow test

File: `internal/rlm/e2e_test.go`
Status: **red-confirmed**

Fails at ParseSpec (first dependency in chain):

```
--- FAIL: TestLikelyCustomersFlowOffline (0.00s)
    e2e_test.go:19: ParseSpec: rlm: model backend not implemented (requires Phase 4 approval)
```

---

## T5.1 — CLI tests

File: `internal/cli/rlm_cli_test.go`
Status: **red-confirmed** (package build failure — expected)

The `internal/cli` package has pre-existing untracked test files referencing unimplemented functions,
AND the `rlm` case is not yet wired in cli.go. Build failure:

```
# polymetrics.ai/internal/cli [polymetrics.ai/internal/cli.test]
internal/cli/runtime_record_test.go:21:11: undefined: runtimeETLLeaseRequest
internal/cli/runtime_record_test.go:25:12: undefined: runtimeETLRunRecord
FAIL	polymetrics.ai/internal/cli [build failed]
```

The rlm_cli_test.go tests will go red→green once T5.2 wires `case "rlm":` in cli.go.

---

## HUMAN GATE — Model backend (Phase 4)

**STOP. Do not implement ModelAnalyzer.Run.**

The model backend requires outbound network calls to the Claude API, credential configuration,
and changes to the network/credential surface of `pm`. A human must explicitly approve Phase 4
before any HTTP client code, credential lookup, or response caching is added to model.go.

---

## Full red-run output (captured 2026-06-27)

```
GOTOOLCHAIN=auto go test ./internal/rlm/... 2>&1

--- FAIL: TestDeterminismSameInputSameOutput (0.00s)
--- FAIL: TestScoringWeightedSum (0.00s)
--- FAIL: TestScoringScoreIfSet_Present (0.00s)
--- FAIL: TestScoringScoreIfSet_Absent (0.00s)
--- FAIL: TestScoringScoreIfGT_Above (0.00s)
--- FAIL: TestScoringScoreIfGT_Below (0.00s)
--- FAIL: TestScoringAllZeroWeights (0.00s)
--- FAIL: TestScoringEmptyRecordSet (0.00s)
--- FAIL: TestSortingByScoreDesc (0.00s)
--- FAIL: TestSortingTiebreakerByRawID (0.00s)
--- FAIL: TestMaterializationWritesNDJSON (0.00s)
--- FAIL: TestMaterializationPreservesSourceFields (0.00s)
--- FAIL: TestMaterializationAtomic (0.00s)
--- FAIL: TestDryRunDoesNotWrite (0.00s)
--- FAIL: TestLikelyCustomersFlowOffline (0.00s)
--- FAIL: TestFixtureRunReturnsRows (0.00s)
--- FAIL: TestFixtureIgnoresInTable (0.00s)
--- FAIL: TestFixtureDryRun (0.00s)
--- FAIL: TestParseSpecValid (0.00s)
--- FAIL: TestParseSpecMissingName (0.00s)
--- FAIL: TestParseSpecNegativeWeight (0.00s)
--- FAIL: TestParseSpecZeroWeight (0.00s)
FAIL
FAIL	polymetrics.ai/internal/rlm	0.356s
```
