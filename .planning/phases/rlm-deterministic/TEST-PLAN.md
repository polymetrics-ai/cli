# TEST-PLAN — Phase 2: RLM Deterministic Backend

All tests use table-driven style matching existing repo conventions (`internal/connectors`, `internal/app`).
No new testing libraries — existing `testing` stdlib package only (repo does not use testify in `internal/rlm`; check `internal/app` for style reference and match it).

---

## Test files and coverage targets

### `internal/rlm/spec_test.go`
| Test | Behavior | Pass condition |
|---|---|---|
| `TestParseSpecValid` | Parse valid JSON spec with 2 features | Returns `*Spec`, nil error, feature count = 2 |
| `TestParseSpecMissingName` | `name` field absent | Returns error containing "name" |
| `TestParseSpecEmptyFeatures` | `features` is `[]` | Returns validation error |
| `TestParseSpecNegativeWeight` | Weight = -1.0 | Returns error containing "weight" |
| `TestParseSpecZeroWeight` | Weight = 0.0 | Allowed (contributes no score) |
| `TestParseSpecScoreIfGTMissingThreshold` | `score_if_gt` set, `threshold` nil | Returns validation error |

### `internal/rlm/deterministic_test.go`
| Test | Behavior | Pass condition |
|---|---|---|
| `TestDeterminismSameInputSameOutput` | Two runs, identical input | All `_rlm_score` values identical; row order identical |
| `TestScoringWeightedSum` | email=set(w=0.5), company=absent(w=0.5); score_if_set=1.0, default=0.0 | Score = 0.5 |
| `TestScoringScoreIfSet_Present` | field non-empty | Assigns `score_if_set` |
| `TestScoringScoreIfSet_Absent` | field empty string | Assigns `default` |
| `TestScoringScoreIfGT_Above` | numeric field 10 > threshold 5 | Assigns `score_if_gt` |
| `TestScoringScoreIfGT_Below` | numeric field 2 < threshold 5 | Assigns `default` |
| `TestScoringAllZeroWeights` | All weights = 0 | Score = 0.0, no panic |
| `TestScoringEmptyRecordSet` | 0 input rows | `RecordsRead=0`, `RecordsScored=0`, no error |
| `TestSortingByScoreDesc` | 3 records with distinct scores | Output order: highest score first |
| `TestSortingTiebreakerByRawID` | 2 records with identical scores | Ordered by `_polymetrics_raw_id` asc |
| `TestMaterializationWritesNDJSON` | Run on temp dir InTable | OutTable exists, correct field set per row |
| `TestMaterializationPreservesSourceFields` | Source record has `email` field | OutTable rows also have `email` field |
| `TestMaterializationAtomic` | No partial writes | No temp file left on success |
| `TestDryRunDoesNotWrite` | `DryRun=true` | OutTable file absent after run |

### `internal/rlm/fixture_test.go`
| Test | Behavior | Pass condition |
|---|---|---|
| `TestFixtureRunReturnsRows` | Any spec | `RecordsScored = len(DefaultFixtureRows)` |
| `TestFixtureScoresMatchDeterministic` | Same spec applied both ways | Score values identical for same input rows |
| `TestFixtureIgnoresInTable` | Non-existent InTable path | No error |
| `TestFixtureDryRun` | `DryRun=true` | OutTable not written |

### `internal/rlm/model_stub_test.go`
| Test | Behavior | Pass condition |
|---|---|---|
| `TestModelStubReturnsNotImplemented` | `ModelAnalyzer.Run` called | Returns `ErrNotImplemented`, zero RunResult |
| `TestModelModeString` | `ModelAnalyzer.Mode()` | Returns `"model"` |

### `internal/rlm/e2e_test.go`
| Test | Behavior | Pass condition |
|---|---|---|
| `TestLikelyCustomersFlowOffline` | Fixture backend + `likely_customers.json` spec | `RecordsScored >= 5`, top score >= 0.5, re-run is identical |

### `internal/cli/rlm_cli_test.go`
| Test | Behavior | Pass condition |
|---|---|---|
| `TestRLMRunDeterministic` | Full CLI, deterministic mode | Exit 0, JSON has `records_scored > 0` |
| `TestRLMRunFixture` | Full CLI, fixture mode | Exit 0, OutTable created |
| `TestRLMRunModelStub` | Full CLI, model mode | Exit non-0, error mentions "not implemented" |
| `TestRLMRunMissingSpec` | `--spec /nonexistent` | Exit 1 |
| `TestRLMRunMissingOutFlag` | No `--out` flag | Exit 1 |
| `TestRLMRunDryRun` | `--dry-run` flag | Exit 0, OutTable NOT written |
| `TestRLMRunJSONOutput` | `--json` flag | stdout is valid JSON object |

---

## TDD ordering (red-first contract)

1. Write T1.1 (spec tests) → run → confirm FAIL → commit red evidence.
2. Implement T1.2 → run → confirm PASS.
3. Write T2.1 (scoring tests) → run → confirm FAIL → commit.
4. Implement T2.2 → run → confirm PASS.
5. Write T3.1 (materialization tests) → run → confirm FAIL → commit.
6. Implement T3.2 → run → confirm PASS.
7. Write T4.1 (fixture tests) → run → confirm FAIL → commit.
8. Implement T4.2 → run → confirm PASS.
9. Write T5.1 (CLI tests) → run → confirm FAIL → commit.
10. Implement T5.2 → run → confirm PASS.
11. Write T7.1 (e2e test) → confirm PASS (fixture already implemented).
12. Full `make verify`.

---

## What is NOT tested in this phase

- Real model/Claude API calls (gated, Phase 4).
- Network calls of any kind (there are none).
- Postgres ledger path (tested in `internal/ledger`; RLM calls ledger interface only).
- Flow engine orchestration (Phase 0 / flow-engine phase).
