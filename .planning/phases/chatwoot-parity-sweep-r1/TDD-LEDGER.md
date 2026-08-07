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

Authored `api_surface.json` (148 rows, full rewrite), `cli_surface.json` and `operations.json` (new
files), and `writes.json` (6 unchanged actions + 54 new, 60 total) programmatically from the
re-derived swagger enumeration, cross-referencing real request/response schemas pulled from the same
`swagger.json` artifact (never hand-typed field names). All four files round-trip exactly at
`json.dumps(indent=2, ensure_ascii=False)+"\n"`.

### A real engine constraint discovered mid-authoring, and how it was resolved

First authoring pass used base-relative paths in `operations.json` (e.g. `/agent_bots`, stripping the
account-scoped prefix myself). `connectorgen surface-sync` then "corrected" 34 `cli_surface.json`
`api_surface` cross-references down to that same relative form, which broke `connectorgen validate`'s
`cli_surface` <-> `api_surface.json` cross-check (34 `cli_surface_unknown_target` findings) because
`api_surface.json`'s own rows correctly stayed on the full documented path. Root cause: unlike every
prior connector in this generation (gong, notion), chatwoot's single base URL bakes in a real path
segment (`/api/v1/accounts/{account_id}`), so "relative to base" and "the documented path" are not the
same string here. Fix: `operations.json`'s `rest.path` for a direct read uses the **full documented
path** (matching `api_surface.json` exactly) — the engine's own
`normalizeDirectReadPathForBaseURL` (`internal/connectors/engine/direct_read.go`) already strips the
base's own prefix at request time when the resolved path starts with it or equals it exactly, and
`{account_id}` resolves automatically from `cfg.Config`, no CLI flag needed. `writes.json` (via
`write.go`'s `joinURL`, which has no such stripping) correctly keeps genuinely relative paths, matching
the six pre-existing write actions' own convention. After the fix, `surface-sync` made 0 corrections
and `validate` passed 0 findings on the first subsequent run.

A second, UX-only defect was caught the same way: direct-read command paths initially used singular
resource nouns (`agent-bot list`) while write command paths used the plural form the write actions
were named after (`agent-bots create`), splitting one resource into two unrelated-looking top-level
`pm chatwoot` groups. Fixed by renaming every direct-read command's first word to match its sibling
write group exactly (`agent-bots list`, `agent-bots get`), verified by asserting every command's first
word resolves to a declared `groups[]` id with none left over.

### Gates (whole packages / whole repo, never a targeted `-run`, per the phase brief)

| Gate | Result |
| --- | --- |
| `go run ./cmd/connectorgen validate internal/connectors/defs/chatwoot` | 0 findings |
| `go run ./cmd/connectorgen validate internal/connectors/defs` (whole repo) | 551 connectors, 0 findings |
| `go test ./cmd/connectorgen/` (whole package) | PASS, including `TestChatwootAPISurfaceOperationLedger` |
| `go test ./internal/connectors/commandrunner/` (whole package, incl. `TestEveryImplementedCommandPassesRuntimePreflight`) | PASS |
| `go run ./cmd/connectorgen surface-sync` then `--check` | 0 filled / 0 corrected across 551 connectors; ledger diff confined to `"chatwoot"` (0 -> 34 entries) |
| `go test ./internal/connectors/conformance/` | PASS |
| `make connector-boundary` | `"outcome": "clean"`, 0 findings, 551 connectors loaded |
| `make docs-check` | PASS |
| `make tidy-check` / `lint` / `agent-contract-check` / `connectorgen-validate` / `connectorgen-surface-sync` / `smoke-no-build` / `release-workflow-check` | all PASS |
| `gofmt -l cmd internal` / `go vet ./cmd/connectorgen/ ./internal/connectors/...` | clean |
| `go test ./internal/cli/ -run TestGoldenTranscripts` | PASS after regenerating (see below) |

### Binary run — reading files is not verification

```
go build ./cmd/pm                                    -> OK
pm connectors inspect chatwoot --json                 -> full manifest, 7 streams, 60 write actions
pm chatwoot                                            -> 19 command groups render, exit 0
pm chatwoot agent-bots --help                          -> 5 commands (2 reads + 3 writes) render together
pm chatwoot conversations list --help                  -> stream, intent=etl, exit 0
pm chatwoot conversations meta --help                  -> direct read, 5 query flags, exit 0
pm chatwoot agent-bots delete --help                   -> destructive write, --confirm challenge, exit 0
pm chatwoot conversations update-custom-attributes --help -> availability=partial, honest NOTES explaining why
```

Every one of the 101 `cli_surface.json` commands' `--help` was driven programmatically:
`TOTAL=101 FAIL=0`.

In a scratch `pm init` project (no credential configured):
```
pm chatwoot conversations meta --json                  -> reaches runtime, fails "missing --credential"
pm chatwoot agent-bots create --preview --json          -> reaches runtime, fails "missing --credential"
```
Both a direct read and a reverse-ETL write preview route through the real runtime end to end, advancing
past project/command resolution to credential resolution — the deepest honest proof available without
a live Chatwoot credential, and no credential was introduced or printed.

**Trailing-slash pair, exercised concretely**: `grep -c "reports" internal/connectors/defs/chatwoot/cli_surface.json` is 0 — no CLI command exists for either row, because both are `blocked-with-named-dependency` (see the engine-constraint finding above), not because either was missed. Both rows were confirmed present, distinct, and identically dispositioned by direct inspection of `api_surface.json`.

### Regenerated artifacts, each inspected before committing

- **Connector docs**: `pm docs generate --dir docs/cli` rewrote 1,032 files; `docs/cli/` had zero
  chatwoot-related changes, and every non-chatwoot file under `docs/connectors/` was reverted
  (`git checkout --pathspec-from-file=...`), leaving exactly `docs/connectors/chatwoot/{MANUAL.md,SKILL.md}`
  changed (362 / 363 lines). Diff read in full: correctly reflects the new 60-write/34-direct-read/
  7-stream command surface, including the field-type annotation fix already seen as pre-existing `main`
  drift in gong's/notion's own lanes.
- **Bundle-internal `docs.md`**: not touched by `pm docs generate` (verified empirically — a stashed
  copy came back unchanged); hand-updated instead, since its "Known limits" section named the removed
  legacy `excluded` categories and the stale 7-stream/6-write counts.
- **`metadata.json`**: `description`/`risk` updated from the original narrow 6-write scope description
  to reflect the full account-scoped surface (also flows into `pm connectors inspect`/`catalog --json`).
- **Website catalog**: `gen-connector-bundles.mjs`, `gen-connector-catalog.mjs`, `gen-connectors.mjs`
  all run. Diff compared **by object, not by line**: `connectors.generated.json` and
  `connectors.catalog.data.generated.json` each show 551 connectors before and after, none added, none
  removed, exactly one changed (`Chatwoot`). The two `.ts` wrapper files did not change at all (chatwoot
  already existed in the name/slug list; only its capability details changed).
- **`TestGoldenTranscripts`**: read the diff **before** regenerating, per the phase brief. 9 of 89
  transcripts differed (`root_bare_manual`, `root_long_help`, `root_short_help`, `root_help_command`,
  `root_man_command`, `root_json_help`, `root_late_json_help`, `root_equals_form`, `root_space_form`) —
  every one of the connector command-surface listings embedded in root help. Verified programmatically
  that removing exactly one literal line (`pm chatwoot <command> - Chatwoot: ...`, alphabetically
  between bitbucket and crisp) from each "after" transcript reproduces its "before" byte-for-byte —
  nothing else moved. Regenerated via `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1`, matching notion's
  precedent exactly (chatwoot newly joins the list, the way notion did; gong was already in it and
  nothing moved there).

## 5. Refactor / safety notes

- No test was weakened, skipped, narrowed, or deleted to reach green. `chatwoot_api_surface_test.go`
  was not edited after the red capture; the bundle was authored to satisfy it as written.
- No credential or token-derived value is emitted anywhere in the bundle; the one operation that would
  return a credential-derived value (`GET /platform/api/v1/users/{id}/login`, a live SSO/impersonation
  URL) is permanently `disallowed`, matching notion's OAuth-endpoint precedent.
- No new dependency was added.
- Both defects caught during authoring (§4's operations.json relative-vs-full-path engine finding, and
  the singular/plural command-group split) were caught by re-running the real gates
  (`connectorgen validate`, `surface-sync`, and the bare `pm chatwoot` help render) after every
  generation pass, not by inspection alone — consistent with the phase brief's "reading files is not
  verification" instruction applied to the authoring loop itself, not just the final report.
- The diff stayed scoped to chatwoot plus the three shared, mechanically-regenerated artifacts every
  connector-parity phase in this programme touches by design: `operation_endpoint_ledger.json` (diff
  confined to the `"chatwoot"` key, 0 -> 34 entries, verified by object), the website catalog (551
  connectors before/after in both files, exactly one changed, verified by object), and
  `internal/cli/testdata/golden_transcripts.json` (9 of 89 transcripts changed, each by exactly the one
  expected added line, verified programmatically). No other connector's files were touched.
