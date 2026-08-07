# Chatwoot parity — TDD ledger

Red-green-refactor evidence for `.planning/phases/chatwoot-parity-sweep-r1`.

## 0. F5 pre-check (existing surface tests)

```
$ ls cmd/connectorgen/ | grep '^chatwoot'
$ grep -rln chatwoot cmd/connectorgen/*_test.go
```

Both commands returned nothing: chatwoot has zero pre-existing `cmd/connectorgen` test files
(unlike gong, which carried two — `gong_api_surface_test.go` and `gong_full_surface_test.go` — and
a targeted `-run` would have missed the second). `chatwoot_api_surface_test.go` is a wholly new file.

## 1. Baseline before any production edit

Measured by reading the bundle directly (there is no built binary surface to run yet — chatwoot has
no `cli_surface.json`, so no chatwoot command is reachable at all pre-change):

| Check | Result |
| --- | --- |
| `internal/connectors/defs/chatwoot/` contents | `api_surface.json`, `docs.md`, `fixtures/`, `metadata.json`, `schemas/`, `spec.json`, `streams.json`, `writes.json` — **no `cli_surface.json`, no `operations.json`** |
| `api_surface.json` rows | 71 (13 `covered_by`, 58 legacy `excluded`, 0 blank) |
| `operation_ledger_version` | absent (implicit 0) |
| Streams | 7: conversations, contacts, inboxes, agents, teams, labels, messages |
| Write actions | 6: create_contact, update_contact, create_conversation, send_message, toggle_conversation_status, create_label |
| Documented operations reachable as `pm chatwoot <command>` | 0 (no `cli_surface.json` exists yet) |

Cross-referencing the 71 existing rows against the live 148-operation swagger: all 71 are genuine
real operations (**zero stale/hallucinated rows** — nothing to retire), and **77 real documented
operations are entirely absent** from the bundle.

## 2. RED — assertion drives the count, before authoring

Wrote `cmd/connectorgen/chatwoot_api_surface_test.go` (new file, modeled on
`cmd/connectorgen/notion_api_surface_test.go` and `cmd/connectorgen/gong_api_surface_test.go`)
asserting the re-derived target surface: 148 endpoints, `operation_ledger_version: 1`, the exact
per-method split (GET 64 / POST 42 / PATCH 22 / DELETE 18 / PUT 2), zero legacy `excluded` rows,
zero blank dispositions, exactly one disposition per row, no duplicate `method`+`path` key, every
blocked row carrying a reason + source citation (`source_url`) + a named dependency
(`notes` prefixed `named_dependency=`), a `model=duplicate` row carrying `duplicate_of`, the
disposition partition (101 executable / 47 blocked), the `covered_by` target split (7 stream / 34
direct_read / 60 write), the blocked-model split (23 `direct_read` / 13 `admin_reverse_etl` / 9
`sensitive_reverse_etl` / 1 `disallowed` / 1 `duplicate`), and — the hazard this phase's brief
called out explicitly — that **both** trailing-slash pair members
(`GET /api/v2/accounts/{account_id}/reports/conversations` and
`GET /api/v2/accounts/{account_id}/reports/conversations/`) survive as distinct row keys.

Ran against the untouched, pre-existing bundle:

```
$ go test ./cmd/connectorgen/ -run TestChatwootAPISurfaceOperationLedger -v
=== RUN   TestChatwootAPISurfaceOperationLedger
    chatwoot_api_surface_test.go:59: operation_ledger_version = 0, want 1
--- FAIL: TestChatwootAPISurfaceOperationLedger (0.00s)
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	0.508s
FAIL
```

The test bites on the very first assertion (the pre-existing bundle has no `operation_ledger_version`
at all, so it unmarshals as the Go zero value `0`) and fails via `t.Fatalf`, which is why later
assertions (row count, method split, disposition partition, trailing-slash presence) never get a
chance to run in this capture — the failure is genuine and observed, not authored to look plausible.
Committed red, on purpose, before any production edit.

## 3. Planned assertions to turn green

| Assertion | Satisfied by |
| --- | --- |
| `operation_ledger_version == 1` | set in the rewritten `api_surface.json` |
| 148 total rows, 0 duplicate `method+path` keys | full rewrite of `endpoints[]` from the re-derived swagger enumeration |
| GET 64 / POST 42 / PATCH 22 / DELETE 18 / PUT 2 | every row's `method` matches its real swagger operation |
| 0 legacy `excluded` rows, 0 blank dispositions, exactly 1 disposition/row | every prior `excluded` row converted to `covered_by` or `operation`; every new row gets exactly one |
| 101 `covered_by` (7 stream / 34 direct_read / 60 write) | 7 unchanged streams; 34 new `operations.json` `rest_read` entries + `cli_surface.json` `direct_read` commands; 60 `writes.json` actions (6 unchanged + 54 new) + `cli_surface.json` `reverse_etl` commands |
| 47 `operation` rows (23 `direct_read` / 13 `admin_reverse_etl` / 9 `sensitive_reverse_etl` / 1 `disallowed` / 1 `duplicate`), each with reason + `source_url` + `notes` prefixed `named_dependency=` | hand-authored blocked rows citing the single-base-URL engine constraint (v2 reports, platform API, public API, survey, profile), the SSO-link credential-emission rule (`disallowed`), the inbox-provisioning `oneOf` rule (`admin_reverse_etl`), and one spec-artifact duplicate row (`duplicate`, with `duplicate_of` set) |
| Both trailing-slash rows present and distinct | both written as literal, distinct `path` strings; verified nothing in the engine, `connectorgen validate`, or this test normalizes a trailing slash (see below) |
| `TestEveryImplementedCommandPassesRuntimePreflight` | `connectorgen surface-sync` resyncs `operation_endpoint_ledger.json` from the authored `operations.json`/`api_surface.json` after authoring |

### Trailing-slash hazard — verified structurally before authoring, not just asserted

Read every place in `internal/connectors/engine/`, `internal/connectors/commandrunner/`, and
`cmd/connectorgen/` that matches a declared endpoint path, to confirm none of them would collapse
the pair:

- `operation_endpoint_ledger.go`'s dedup key and runtime lookup: `entry.Path == endpointPath` (exact,
  `strings.EqualFold` only on `Method`).
- `cmd/connectorgen/validate.go`'s `surfaceEndpointKey`: `strings.ToUpper(strings.TrimSpace(method)) +
  " " + strings.TrimSpace(path)` — `TrimSpace` removes surrounding whitespace only, never a trailing
  `/` that is part of the path itself.
- This test's own `key := ep.Method + " " + ep.Path` — plain string concatenation, and
  `findOperation` (engine) looks up `operations.json` entries by `id`, never by path, so the two rows'
  distinct operation names (`chatwoot.account_conversation_metrics_get` /
  `chatwoot.agent_conversation_metrics_get`) never collide either.

Both rows are, in this bundle, `operation`-blocked (the `/api/v2` prefix is outside the single base
URL — see PLAN.md's implementation-findings section), so the hazard is proven at the `api_surface.json`
row level; the assertion still pins both literal path strings as distinct so a future author who
promotes either to executable inherits the same proof.

## 4. GREEN evidence

Filled in after authoring — see `VERIFICATION.md` and the final report for gate output, binary run
transcript, and regenerated-artifact diffs.

## 5. Refactor / safety notes

- No test was weakened, skipped, narrowed, or deleted to reach green. The only edits to
  `chatwoot_api_surface_test.go` after the red capture, if any, are additive.
- No credential or token-derived value is emitted anywhere in the bundle; the one operation that would
  return a credential-derived value (`GET /platform/api/v1/users/{id}/login`, a live SSO/impersonation
  URL) is permanently `disallowed`, matching notion's OAuth-endpoint precedent.
- No new dependency was added.
