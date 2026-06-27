# EVAL-PLAN — Phase 2: RLM Deterministic Backend

## Purpose

Evaluate that the deterministic backend meets its behavioral contracts before the phase is considered done. This is not model evaluation — there is no model in this phase. This is correctness and property evaluation.

---

## E1 — Determinism property

**What:** Given identical InTable content and identical Spec, two runs must produce byte-for-byte identical OutTable files.

**How:** `TestDeterminismSameInputSameOutput` in `internal/rlm/deterministic_test.go`:
1. Build 10 records with randomized field values (seeded deterministically).
2. Run `DeterministicAnalyzer.Run` twice with same inputs.
3. Assert both OutTable files are byte-identical (compare via `os.ReadFile` + `bytes.Equal`).

**Pass threshold:** 100% — any non-determinism is a blocking failure.

---

## E2 — Scoring accuracy

**What:** For a known spec and known input, the output score must match hand-calculated expected values.

**How:** `TestScoringWeightedSum` in `deterministic_test.go`:
- 2 features: `email` (weight=0.5, score_if_set=1.0, default=0.0) and `company` (weight=0.5, score_if_set=1.0, default=0.0).
- Record A: email set, company absent → expected score = (0.5*1.0 + 0.5*0.0) / (0.5+0.5) = 0.5.
- Record B: both set → expected score = 1.0.
- Record C: neither set → expected score = 0.0.

**Pass threshold:** Scores match expected within `1e-9` (float64 precision).

---

## E3 — Fixture parity

**What:** Applying the same spec to `DefaultFixtureRows` via both `FixtureAnalyzer` and `DeterministicAnalyzer` (with fixture rows written to a temp InTable) must yield identical `_rlm_score` values for each record.

**How:** `TestFixtureScoresMatchDeterministic` in `fixture_test.go`.

**Pass threshold:** 100% match across all fixture rows.

---

## E4 — End-to-end offline flow

**What:** The `likely_customers.json` spec + fixture backend produces a sensible scoring distribution (not all 0.0, not all 1.0; top record has score >= 0.5).

**How:** `TestLikelyCustomersFlowOffline` in `e2e_test.go`.

**Pass threshold:** At least one record scores >= 0.5. Scores span at least 2 distinct values across the fixture set.

---

## E5 — CLI contract

**What:** CLI flags, exit codes, and JSON envelope match API-CONTRACT.md.

**How:** `TestRLMRunJSONOutput` in `internal/cli/rlm_cli_test.go` — parse stdout with `json.Unmarshal`, assert all required fields present and correctly typed.

**Pass threshold:** 100% — any missing field or wrong type is a contract violation.

---

## E6 — Model stub gate

**What:** `--mode model` never silently succeeds; always returns a clear error.

**How:** `TestRLMRunModelStub` (CLI) and `TestModelStubReturnsNotImplemented` (unit).

**Pass threshold:** 100%. Any path that lets model mode complete a run is a regression.

---

## Eval execution

All evals run as standard Go tests:
```bash
go test ./internal/rlm/... -v -count=1
go test ./internal/cli/... -run TestRLM -v -count=1
```

`-count=1` disables test caching to ensure determinism checks actually re-execute.
