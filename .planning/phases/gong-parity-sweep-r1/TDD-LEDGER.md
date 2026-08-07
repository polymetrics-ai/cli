# Gong parity — TDD ledger

Red-green-refactor evidence for `.planning/phases/gong-parity-sweep-r1`.

## 1. Baseline before any production edit

Measured by **running the built binary**, not by reading files.

| Check | Command | Result |
| --- | --- | --- |
| Build | `go build ./cmd/pm` | OK |
| Command surface | `pm gong` | 18 groups render, exit 0. **No `targets` group.** |
| Every declared command reachable | `pm gong <path> --help` over all 67 `cli_surface.json` commands | **67 ok, 0 failed** |
| Intent partition | `cli_surface.json` | 12 `etl` / 29 `direct_read` / 26 `reverse_etl` |
| Availability | `cli_surface.json` | 43 `implemented` / 24 `partial` |
| Dispositions | `api_surface.json` | 67 `covered_by`, **0 blank** |
| Surface test | `go test ./cmd/connectorgen -run TestGongAPISurfaceOperationLedger` | PASS at 67 |

Baseline is honest: nothing was falsely marked `implemented`, and the 24 `partial` rows carry a
named constraint (flat CLI flags cannot express records with complex object/array fields; the typed
reverse-ETL path is required). That is a real engine limitation, recorded — not a defect introduced
here and not something this connector can fix alone.

## 2. RED — assertion drives the count, before authoring

Derived 69 from the provider artifact (see `PLAN.md`), then updated
`cmd/connectorgen/gong_api_surface_test.go`:

- `endpoints` 67 → **69**
- `covered` 67 → **69**
- `totalByMethod` / `coveredByMethod`: GET 28 → **29**, POST 27 → **28** (PUT 8, DELETE 3, PATCH 1 unchanged)
- new presence assertions for `GET /v2/targets` and `POST /v2/targets/{targetId}/assignments`

Observed failure, captured verbatim:

```
=== RUN   TestGongAPISurfaceOperationLedger
    gong_api_surface_test.go:76: endpoints = 67, want 69
--- FAIL: TestGongAPISurfaceOperationLedger (0.00s)
FAIL	polymetrics.ai/cmd/connectorgen	0.505s
```

The test bites. Committed red, on purpose, before production edits.

## 3. Planned assertions to turn green

| Assertion | Satisfied by |
| --- | --- |
| 69 endpoints, 69 covered, 0 blank | 2 new `api_surface.json` rows |
| GET 29 | `GET /v2/targets` as a **direct read** |
| POST 28 | `POST /v2/targets/{targetId}/assignments` as a **multipart write** |
| Both Targets keys present | new `targets` command group |
| No duplicate endpoint key | new paths are unique — verified absent from the current 67 |
| `TestEveryImplementedCommandPassesRuntimePreflight` | operation endpoint ledger resynced via `connectorgen surface-sync` |

## 4. GREEN evidence

_(filled in as gates run)_

## 5. Refactor / safety notes

- No test was weakened, skipped, narrowed, or deleted to reach green. The only test edit **raises**
  the asserted count and **adds** two presence assertions.
- The binary upload reuses the existing `engine.MultipartSpec` contract already proven in gong by
  `upload_call_media` and `upload_crm_entities`. No ad-hoc HTTP, no new dependency.
- File path and content are redacted in plans, matching the two existing multipart exemplars.
