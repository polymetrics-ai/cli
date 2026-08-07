# Gmail parity — TDD ledger

Red-green-refactor evidence for `.planning/phases/gmail-parity-sweep-r1`.

## 1. Baseline before any production edit

| Check | Result |
| --- | --- |
| Artifact | Google Discovery `gmail:v1`, **revision 20260803**, 217,687 bytes, HTTP 200, public |
| Documented operations | **79** (GET 30, POST 28, DELETE 10, PUT 8, PATCH 3) |
| Ledger target | 79 — **reconciles exactly** |
| `api_surface.json` rows | 79, method split already matches the artifact |
| Dispositions | 10 `covered_by.stream` · 35 `covered_by.write` · **34 legacy `excluded`** · 0 blank |
| `operation_ledger_version` | **unset** |
| `cli_surface.json` | **ABSENT** |
| `operations.json` | **ABSENT** |
| Pre-existing gmail tests (finding F5 check) | **none** — `ls cmd/connectorgen/ \| grep '^gmail'` empty, and no other connectorgen test references gmail |

## 2. RED — committed failing, before production edits

Added `cmd/connectorgen/gmail_api_surface_test.go`. Observed failure, captured verbatim:

```
--- FAIL: TestGmailAPISurfaceOperationLedger (0.00s)
    gmail_api_surface_test.go:80: operation_ledger_version = 0, want 1 (the v2 provenance ledger is required)
    gmail_api_surface_test.go:139: 34 legacy excluded row(s) remain, want 0; re-disposition them as covered or blocked-with-named-dependency
    gmail_api_surface_test.go:146: covered(45)+blocked(0) = 45, want 79
FAIL	polymetrics.ai/cmd/connectorgen	0.499s
```

**The test isolates the real gap precisely.** The `endpoints = 79` and `totalByMethod` assertions
**pass** on today's bundle, because gmail's *count* is already right. What is wrong is its
*dispositions*. Three failures, no noise.

The test additionally enforces, for every row:

- **exactly one** disposition — never zero (blank), never two;
- a blocked row must carry a **reason**, a **source citation**, and a **named dependency**
  (an explicit `dependency` field or an issue reference), so "blocked" can never be a shrug;
- **zero** legacy `excluded` rows, since that category is not one of the three this sweep accepts;
- no duplicate `method + path` key.

## 3. Planned assertions to turn green

| Assertion | Satisfied by |
| --- | --- |
| `operation_ledger_version: 1` | v2 provenance ledger added to `api_surface.json` |
| 0 legacy `excluded` | all 34 re-dispositioned per `PLAN.md` |
| covered + blocked = 79 | 26 promoted to covered, 13 to blocked-with-named-dependency |
| blocked rows name a dependency | 11 CSE → Workspace add-on citation; 2 bulk → **#514** |
| promoted endpoints present | `watch`, `stop`, smimeInfo, detail GETs |
| individually reachable | `cli_surface.json` authored from scratch |
| `TestEveryImplementedCommandPassesRuntimePreflight` | ledger resync via `connectorgen surface-sync` |

## 4. GREEN evidence

_(not yet — authoring is the next step)_

## 5. Refactor / safety notes

- No test weakened, skipped, narrowed, or deleted. This test is **new** and strictly additive.
- The shared engine must not be widened from inside this connector PR; the array-body gap
  (`batchDelete`/`batchModify`) belongs to **#514** and is recorded as a named block, not worked
  around.
- `attachments.get` must not be marked implemented unless a bounded engine **download** capability
  genuinely exists — to be checked, not assumed.
