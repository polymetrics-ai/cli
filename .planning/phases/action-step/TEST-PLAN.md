# TEST-PLAN — Action Step (Phase 1)

## Test file

`internal/flow/action_test.go` — all behaviour tests for the action step.
`internal/cli/flow_cli_test.go` — CLI integration (extends existing file).

## Test framework

`github.com/stretchr/testify` (already in go.mod). `net/http/httptest` for fake destination.

## Test IDs and assertions

### T-10 — Action manifest validation
| Case | Input | Expected |
|------|-------|----------|
| T-10a | valid action step | no errors |
| T-10b | missing source_table | ErrManifestInvalid |
| T-10c | missing destination_connector | ErrManifestInvalid |
| T-10d | missing mappings | ErrManifestInvalid |
| T-10e | missing action field | defaults to "upsert", no error |

### T-11 — Idempotent writes (httptest.Server)
| Case | Input | Expected |
|------|-------|----------|
| T-11a | run action twice with same 5 records | server receives 5 records on run 1, 0 on run 2 |
| T-11b | run with 3 new + 2 seen records | server receives exactly 3 on second run |

### T-12 — Identity mapping
| Case | Input | Expected |
|------|-------|----------|
| T-12a | server returns ext ID in response | identity_map.json written with pm_id→ext_id |
| T-12b | re-run with same records | records skipped (identity map hit) |

### T-13 — Dedupe
| Case | Input | Expected |
|------|-------|----------|
| T-13a | 3 records, 2 unique emails | server receives 2 |
| T-13b | no duplicate fields | all records pass through |

### T-14 — Rate-limit backoff
| Case | Input | Expected |
|------|-------|----------|
| T-14a | server returns 429 ×2 then 200 | action succeeds; server called 3 times |
| T-14b | server always returns 429 | ErrMaxRetriesExhausted; record goes to DLQ |

### T-15 — DLQ
| Case | Input | Expected |
|------|-------|----------|
| T-15a | server returns 400 for one record | DLQ file exists; RecordsFailed==1 |
| T-15b | successful records not in DLQ | RecordsSucceeded correct |

### T-16 — Schema drift
| Case | Input | Expected |
|------|-------|----------|
| T-16a | field type changed (string→int) | ErrSchemaDrift; 0 records sent |
| T-16b | field removed | ErrSchemaDrift; 0 records sent |
| T-16c | new field added | no error; run proceeds |
| T-16d | no drift | no error |

### T-17 — Receipts
| Case | Input | Expected |
|------|-------|----------|
| T-17a | successful action run | ledger has ≥1 entry with Mode=="action", Status=="receipt" |
| T-17b | failed run | no receipt entry (or partial entry for succeeded records only) |

### T-18 — Engine approval gate
| Case | Input | Expected |
|------|-------|----------|
| T-18a | action step + no token | ErrApprovalRequired; ActionRunner not called |
| T-18b | action step + valid token | ActionRunner.Execute called |
| T-18c | action step + expired token | ErrTokenExpired |

### T-19 — CLI
| Case | Input | Expected |
|------|-------|----------|
| T-19a | pm flow plan myflow | JSON output with approval_token field |
| T-19b | pm flow run --token X myflow | engine called with token X |
| T-19c | pm flow run myflow (no token, has action step) | non-zero exit, error JSON |

## Ordering guarantee

Tests run in wave order. Each T-N is committed RED before B-N is written.
`go test ./internal/flow/... ./internal/cli/...` must pass at wave end.
`make verify` must pass at phase end.

## Non-negotiables

- Tests must never be weakened.
- `httptest.Server` is the only external system in tests (no live network).
- No test may import a package not already in go.mod.
