# Help Scout parity — TDD ledger

Red-green-refactor evidence for `.planning/phases/help-scout-parity-sweep-r1`.

## 1. Baseline before any production edit

| Check | Result |
| --- | --- |
| Artifact | `developer.helpscout.com/mailbox-api/` left-nav → **146 endpoint pages**, each fetched individually |
| Documented operations | **144** (GET 79, POST 21, PUT 20, PATCH 6, DELETE 18) |
| Ledger target | 146 (pages) · sweep derivation 145 (literal paths) — **both are dedup drift, not missing endpoints** |
| `api_surface.json` rows | **8** — 4 `covered_by`, **4 legacy `excluded`**, one of them a **wildcard** standing for 33 endpoints |
| `operation_ledger_version` | **unset** |
| `writes.json` | **ABSENT** (`capabilities.write: false`) |
| `cli_surface.json` / `operations.json` | **ABSENT** |
| Pre-existing help-scout tests (finding F5 check) | **none** |

## 2. RED — committed failing, before production edits

`cmd/connectorgen/help_scout_api_surface_test.go`, run against the real bundle:

```
=== RUN   TestHelpScoutAPISurfaceOperationLedger
    help_scout_api_surface_test.go:97: operation_ledger_version = 0, want 1
    help_scout_api_surface_test.go:128: "GET /v2/reports/*" is a wildcard, not an operation; enumerate the endpoints it stands for
    help_scout_api_surface_test.go:168: 4 legacy excluded row(s) remain, want 0
    help_scout_api_surface_test.go:171: endpoints = 8, want 144 documented operations
    help_scout_api_surface_test.go:174: covered(4)+blocked(0) = 4, want 144
    help_scout_api_surface_test.go:177: totalByMethod = map[GET:6 POST:2], want map[DELETE:18 GET:79 PATCH:6 POST:21 PUT:20]
    help_scout_api_surface_test.go:182: expected DELETE /v2/customers/{customerId} (the sync and async deletes are one operation; async is a query parameter)
    help_scout_api_surface_test.go:189: expected the attachment file download endpoint
    help_scout_api_surface_test.go:192: expected the attachment data endpoint (base64 in JSON, NOT a binary download)
    help_scout_api_surface_test.go:196: expected out-of-base v3 row "GET /v3/conversations/{conversationId}"
    help_scout_api_surface_test.go:196: expected out-of-base v3 row "GET /v3/conversations/{conversationId}/threads"
    help_scout_api_surface_test.go:196: expected out-of-base v3 row "GET /v3/customers"
    help_scout_api_surface_test.go:196: expected out-of-base v3 row "GET /v3/system-users"
    help_scout_api_surface_test.go:196: expected out-of-base v3 row "GET /v3/system-users/{systemUserId}"
--- FAIL: TestHelpScoutAPISurfaceOperationLedger (0.00s)
FAIL	polymetrics.ai/cmd/connectorgen	0.581s
```

**Red was authored before it could be run** — the shared Go build cache had been corrupted by a
host-wide disk-full window and `go test` would not build. Rather than claim evidence it did not
have, the test was committed with `red_confirmed: false` and an explicit blocker note; the run above
was made and captured in a **second** commit once the cache recovered. No authoring happened in
between.

## 3. GREEN evidence

| Gate | Result |
| --- | --- |
| `TestHelpScoutAPISurfaceOperationLedger` | **PASS** |
| **Whole** `cmd/connectorgen` package (finding F5) | `ok polymetrics.ai/cmd/connectorgen 12.059s` |
| `connectorgen validate internal/connectors/defs/help-scout` | **0 findings** |
| `connectorgen surface-sync --check` | 551 scanned, 0 filled / 0 corrected |
| `TestEveryImplementedCommandPassesRuntimePreflight` | **PASS** |
| Commands reachable | **139/139** by running the built binary; 0 unreachable |
| Per-command paging flags | **none** |
| Endpoint-ledger delta | **none** — the only operation is a `binary_download`, and the ledger only covers `rest_read`/`provider_search` |

Final shape: **144** rows = **139 covered** + **5 blocked**, 0 excluded, 0 blank, 0 duplicates, no
query-string rows, no wildcard rows, `operation_ledger_version: 1`.
24 streams · 49 direct reads · 1 binary download · 65 write actions → **139 commands**.
`capabilities.write` false → true.

## 4. Two defects the gates caught, recorded because both were mine

1. **The dedup survivor named the write action.** The sync and async customer deletes collapse to one
   operation. The *async* page won the dedup, so the generated action was
   `delete_customer_asynchronously` — with a path that performs the **synchronous** delete. A command
   that says async and does sync is worse than either. Renamed to `delete_customer` (named for the
   endpoint) with `--async` exposed as an optional `omit_when_absent` query flag, so collapsing the
   count loses no documented behaviour.
2. **An assumed primary key.** New stream schemas were given `x-primary-key: ["id"]`; Help Scout's
   user-status records key on `userId`. `connectorgen validate` rejected it as
   `primary_key_missing`, and the generator now takes the key from the record the provider actually
   returns.

## 5. Refactor / safety notes

- No test weakened, skipped, narrowed, or deleted. This test is **new** and strictly additive.
- The `/v2` → host-root rebase that would unblock the five v3 rows is **not** performed here: it
  changes a shipped `spec.json` default and needs a config migration. Recorded as the named
  dependency instead.
- No paging flags hand-authored.
