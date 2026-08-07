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

Authored the two operations across `api_surface.json`, `cli_surface.json`, `writes.json`,
`operations.json`. All four files round-trip exactly at `json.dumps(indent=2, ensure_ascii=False)`,
so the bundle diff is **180 insertions / 2 deletions** with zero formatting churn — the 2 deletions
are the `scope` and `reviewed_at` fields being updated to `59 paths / 69 operations` and `2026-08-07`.

### Gates

| Gate | Result |
| --- | --- |
| `connectorgen validate internal/connectors/defs/gong` | **0 findings** |
| `TestGongAPISurfaceOperationLedger` | **PASS** at 69 (was the red failure) |
| `TestEveryImplementedCommandPassesRuntimePreflight` | **PASS** — the #3890 backstop |
| `connectorgen surface-sync --check` | clean, 551 scanned, **0 drift** |
| conformance suite | **PASS** (17.9s) |
| `make certify-timing` | **PASS**, exit 0, **92 real CLI invocations at budget**, total 108.6s |
| `make connector-boundary` | **PASS**, exit 0, no gong findings |
| `make docs-check` | **PASS** |
| `TestGoldenTranscripts` | **PASS unchanged** |
| `make tidy-check` / `lint` / `smoke` / `agent-contract-check` / `connectorgen-validate` / `connectorgen-surface-sync` / `release-workflow-check` | **all PASS** |
| `gofmt -l cmd internal` / `go vet` | clean |

### Trap 1 — operation endpoint ledger: does not bite here, and the reason matters

`surface-sync --check` reports **0 drift** and preflight passes with **no** gong Targets entry in
`internal/connectors/defs/operation_endpoint_ledger.json`. That is correct, not a miss: gong's 13
ledger rows are all **POST `rest_read`** — endpoints whose method would otherwise read as a write.
`GET /v2/targets` is an unambiguous GET, so it needs no disambiguating entry. Notion's 18 direct
reads needed them; gong's does not.

### Trap 2 — golden transcripts: read before regenerating, and no regeneration was needed

`TestGoldenTranscripts` passes **unchanged**. gong was already in the root-help connector
command-surface list, so unlike notion (which joined it) nothing moved. Nothing regenerated blindly.

### Binary run — reading files is not verification

```
pm gong                              -> targets group now renders
pm gong targets --help               -> 2 commands, both availability=implemented
pm gong targets list --help          -> intent=direct_read, output_policy=json_redacted
pm gong targets upload-assignments --help
                                     -> intent=reverse_etl, write=upload_target_assignments,
                                        destructive --confirm challenge, 4 typed flags
pm gong targets list --workspaceId 123 --json          -> reaches runtime
pm gong targets upload-assignments ... --preview --json -> reaches runtime, then `missing --credential`
```

Both commands route through the real runtime: in a scratch `pm init` project the write preview
advances past project resolution to credential resolution. That is the deepest honest proof
available without a live Gong credential, and no credential was introduced or printed.

### Regenerated artifacts, each inspected by this lane before committing

- **Connector docs** — `pm docs generate` rewrote **1,034 files**. Compared before accepting:
  all non-gong churn is **pre-existing generator drift in `main`** (committed docs carry
  `field()` with empty types; the current generator emits `field(string)`). Reverted every non-gong
  file; kept gong's own generated output, which necessarily carries both the Targets additions and
  the same type correction, because a hand-stripped file would disagree with its own generator.
  Net: `docs/connectors/gong/{MANUAL.md,SKILL.md}`, 19 insertions / 12 deletions each.
  **Recorded as a repo-wide finding, not fixed here** — see PROGRESS.md.
- **Website catalog** — compared **by object, not by line**: 551 connectors before and after, none
  added, none removed, **exactly one changed (`Gong`)**, in both
  `website/data/connectors.generated.json` and `website/lib/connectors.catalog.data.generated.json`.
  The generator's own summary independently corroborates the correction-3 audit: `"database": 2`.

## 5. Refactor / safety notes

- No test was weakened, skipped, narrowed, or deleted to reach green. The only test edit **raises**
  the asserted count and **adds** two presence assertions.
- The binary upload reuses the existing `engine.MultipartSpec` contract already proven in gong by
  `upload_call_media` and `upload_crm_entities`. No ad-hoc HTTP, no new dependency.
- File path and content are redacted in plans, matching the two existing multipart exemplars.
