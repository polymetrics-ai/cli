# Greenhouse parity — TDD ledger

Red-green-refactor evidence for `.planning/phases/greenhouse-parity-sweep-r1`.

## 1. Baseline before any production edit

| Check | Result |
| --- | --- |
| Artifact | `developers.greenhouse.io/harvest.html`, HTTP 200, **1,636,662 bytes** — byte-identical to the sweep derivation |
| Documented operations | **138** (GET 69, POST 28, PUT 8, PATCH 19, DELETE 14), 0 duplicates |
| Ledger target | 138 — **reconciles exactly** |
| `api_surface.json` rows | **129** — nine short |
| Dispositions | 126 `covered_by` · **3 legacy `excluded`** · 0 blocked · 0 blank |
| `operation_ledger_version` | **unset** |
| `cli_surface.json` | **ABSENT** |
| `operations.json` | **ABSENT** |
| `pm greenhouse` | `error: unknown command "greenhouse"` — 0 of 126 covered operations reachable |
| Pre-existing greenhouse tests (finding F5 check) | **none** — no `cmd/connectorgen` test referenced greenhouse |

## 2. RED — committed failing, before production edits

Added `cmd/connectorgen/greenhouse_api_surface_test.go` against the **real, unmodified** bundle.
Observed failure, captured verbatim:

```
=== RUN   TestGreenhouseAPISurfaceOperationLedger
    greenhouse_api_surface_test.go:85: operation_ledger_version = 0, want 1
    greenhouse_api_surface_test.go:148: 3 legacy excluded row(s) remain, want 0 (deprecated operations still count and must carry an operation disposition, not an excluded stub)
    greenhouse_api_surface_test.go:152: endpoints = 129, want 138 documented operations
    greenhouse_api_surface_test.go:155: covered(126)+blocked(0) = 126, want 138
    greenhouse_api_surface_test.go:158: totalByMethod = map[DELETE:12 GET:69 PATCH:13 POST:27 PUT:8], want map[DELETE:14 GET:69 PATCH:19 POST:28 PUT:8]
    greenhouse_api_surface_test.go:163: expected DELETE /tags/candidate/{tag_id} (its docs markup carries a stray &#39; and a placeholder with a literal space; a naive extraction drops it)
    greenhouse_api_surface_test.go:169: expected out-of-base row "DELETE /v2/jobs/{job_id}/openings"
    greenhouse_api_surface_test.go:169: expected out-of-base row "PATCH /v2/job_posts/{job_post_id}"
    greenhouse_api_surface_test.go:169: expected out-of-base row "PATCH /v2/job_posts/{job_post_id}/status"
    greenhouse_api_surface_test.go:169: expected out-of-base row "PATCH /v2/scheduled_interviews/{scheduled_interview_id}"
    greenhouse_api_surface_test.go:169: expected out-of-base row "PATCH /v2/users/"
    greenhouse_api_surface_test.go:169: expected out-of-base row "PATCH /v2/users/disable"
    greenhouse_api_surface_test.go:169: expected out-of-base row "PATCH /v2/users/enable"
    greenhouse_api_surface_test.go:169: expected out-of-base row "POST /v2/scheduled_interviews"
--- FAIL: TestGreenhouseAPISurfaceOperationLedger (0.00s)
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	0.575s
FAIL
```

**The failure is specific, and the `GET: 69` half of the method split already passes** — greenhouse's
read surface was complete. What is missing is nine mutations and a disposition model.

The test additionally enforces, for every row:

- **exactly one** disposition — never zero (blank), never two;
- a blocked row must carry a **reason**, a **source citation**, and a machine-checkable
  `Named dependency:` marker, so "blocked" can never be a shrug;
- **zero** legacy `excluded` rows — deprecated operations still count and must be dispositioned;
- no duplicate `method + path` key;
- no `WEBHOOK`-method rows (events are excluded from the operation surface by policy).

## 3. Planned assertions to turn green

| Assertion | Satisfied by |
| --- | --- |
| `operation_ledger_version: 1` | provenance ledger added to `api_surface.json` |
| 0 legacy `excluded` | 3 deprecated v1 mutations re-dispositioned as blocked, each naming its v2 replacement |
| endpoints = 138 | +1 recovered markup-damaged row, +8 Harvest v2 rows |
| method split matches | the 9 added rows are exactly 1 DELETE + 1 POST… see `PLAN.md` |
| blocked rows name a dependency | v2 rows → per-operation base URL override; deprecated rows → their v2 successor |
| `DELETE /tags/candidate/{tag_id}` present | recovered from the damaged declaration and added to `writes.json` |
| individually reachable | `cli_surface.json` + `operations.json` authored from scratch |
| `TestEveryImplementedCommandPassesRuntimePreflight` | endpoint ledger resync via `connectorgen surface-sync` |

## 4. GREEN evidence

_(not yet — authoring is the next step)_

## 5. Refactor / safety notes

- No test weakened, skipped, narrowed, or deleted. This test is **new** and strictly additive.
- The per-operation base URL override is **not** worked around from inside this connector. The 8 v2
  operations are recorded as blocked with a named dependency, exactly as chatwoot's 47 were.
- No paging flags are hand-authored; the foundation lane derives them.
