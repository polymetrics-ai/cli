# Mixpanel parity — TDD ledger

Red-green-refactor evidence for `.planning/phases/mixpanel-parity-sweep-r1`.

## 0. F5 check

```
$ ls cmd/connectorgen/ | grep '^mixpanel'
(no output, exit 1)
$ grep -rln mixpanel cmd/connectorgen/*_test.go
(no output)
```

No pre-existing mixpanel surface test. Clean slate — one new test file, not a targeted `-run` that
could miss a second pre-existing test (the gong trap this plan calls out).

## 1. Baseline before any production edit

Measured from the pre-existing bundle at `internal/connectors/defs/mixpanel/`, not assumed.

| Check | Result |
| --- | --- |
| `api_surface.json` | 47 rows: 10 `covered_by` (all `stream`), 37 legacy `excluded`, 0 blank |
| `operation_ledger_version` | absent (0) |
| `streams.json` | 10 streams (all reads); no `writes.json`, `operations.json`, or `cli_surface.json` |
| `metadata.json` capabilities | `read: true, write: false, query: false, dynamic_schema: false` |
| Auth | `mode: custom`, hook `mixpanel` (Basic auth, username/password/api_secret, config-then-secrets precedence) |

This bundle predates the 13-file provider-artifact ledger: it mixes legacy Query API v2.0 endpoints
(not part of the 104-operation ledger at all — never fetched in the 13 files) with a handful of
current-API list/detail GET streams, and excludes the entire admin/ingestion/query surface as
`requires_elevated_scope`/`non_data_endpoint`/`binary_payload`/etc. under the legacy `excluded`
vocabulary this sweep retires.

## 2. RED — the assertion drives the build, before authoring

Wrote `cmd/connectorgen/mixpanel_api_surface_test.go` asserting 104 endpoints, the exact by-method
split (GET 41 / POST 44 / PUT 5 / PATCH 4 / DELETE 10), `operation_ledger_version` set, zero legacy
`excluded` rows, zero blank dispositions, exactly one disposition per row, no duplicate
`method+path`, every blocked row carrying `reason` + `source_url` + a `named_dependency=`-prefixed
`notes` field, and an explicit single-occurrence assertion for `POST /import` (the identity-merge /
import-events collision this plan flags up front).

Observed failure, captured verbatim (`go test ./cmd/connectorgen/ -run TestMixpanelAPISurfaceOperationLedger -v`):

```
=== RUN   TestMixpanelAPISurfaceOperationLedger
    mixpanel_api_surface_test.go:61: operation_ledger_version = 0, want it set (nonzero)
    mixpanel_api_surface_test.go:65: endpoints = 47, want 104
--- FAIL: TestMixpanelAPISurfaceOperationLedger (0.00s)
FAIL	polymetrics.ai/cmd/connectorgen	0.510s
FAIL
```

Confirmed again running the WHOLE package (not a targeted `-run`), matching the plan's explicit
instruction and catching any second pre-existing test the F5 check might have missed:

```
$ go test ./cmd/connectorgen/
--- FAIL: TestMixpanelAPISurfaceOperationLedger (0.00s)
    mixpanel_api_surface_test.go:61: operation_ledger_version = 0, want it set (nonzero)
    mixpanel_api_surface_test.go:65: endpoints = 47, want 104
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	11.479s
FAIL
```

Only the new test fails; nothing else in the package regresses. The test bites for the right
reason (count, not a typo) and is committed red, on purpose, before any production edit.

## 3. Surprise flagged before continuing (recorded in `RUN-STATE.json.surprise`)

Two genuine surprises beyond the plan's own hazard list, both resolved without weakening any gate:

1. **Scale**: unlike gong/notion (already near-parity from an earlier dedicated build program,
   needing only a small top-off), mixpanel enters this sweep at 10 read-only streams with
   `capabilities.write=false` and no `writes.json`/`operations.json`/`cli_surface.json` at all.
   Reaching 104 requires a first full build across 13 admin/ingestion/query domains, not an
   increment.
2. **Engine host constraint**: `operations.json`-backed `direct_read`/`direct_write`/
   `binary_download` executors require a connector-relative path resolved against the ONE
   configured `base_url` (`isAbsoluteHTTPURL` is rejected for all three). Mixpanel's 13 files span
   5 real hosts. `streams.json`/`writes.json` both already support absolute-URL paths bypassing
   `base_url` (proven by 7 of the 10 existing streams and by `youtube-analytics`' `writes.json`), so
   every new read/write is modeled as a stream (`intent: etl`) or write action
   (`intent: reverse_etl`) with an explicit absolute URL, never through `operations.json`.

## 4. Planned assertions to turn green

| Assertion | Satisfied by |
| --- | --- |
| 104 endpoints, `operation_ledger_version` set | full `api_surface.json` rewrite |
| GET 41 / POST 44 / PUT 5 / PATCH 4 / DELETE 10 | per-operation method taken verbatim from the 13 real specs |
| 0 legacy `excluded`, 0 blank, exactly 1 disposition/row | every row is `covered_by` (stream/write) or `operation` (blocked) |
| Every blocked row: reason + source_url + `named_dependency=` notes | 4 blocked rows: `POST /jql` (generic script executor, disallowed), `GET /export` (cross-host NDJSON binary, no fitting executor), `POST /nessie/pipeline/create` and `POST /nessie/pipeline/edit` (oneOf/discriminator bodies, no single fixed contract) |
| `POST /import` exactly once | one `import_events` write action serves both `identity-merge` and `import-events` semantics (a `$merge` pseudo-event is an ordinary import record) |
| Every row individually reachable as `pm mixpanel <command>` | new `cli_surface.json` with one command per covered row |

## 5. GREEN evidence

Filled in after the bundle is authored — see the end of this file / commit history for the
post-green update.
