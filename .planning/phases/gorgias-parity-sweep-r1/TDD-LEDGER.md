# Gorgias documented-operation parity — TDD ledger

## Baseline (before this phase)

`internal/connectors/defs/gorgias/` was a 4-stream, read-only bundle:
`metadata.json` (`capabilities.write: false`), `spec.json`, `streams.json` (4 streams: `tickets`,
`customers`, `messages`, `satisfaction_surveys`), `api_surface.json` (11 rows: 4 `covered_by.stream`
rows plus 7 legacy `excluded` rows), `docs.md`, `schemas/*.json`, `fixtures/{check.json,streams/**}`.
No `operations.json`, `cli_surface.json`, or `writes.json` existed.

F5 check (per plan step 1): `ls cmd/connectorgen/ | grep '^gorgias'` and
`grep -rln gorgias cmd/connectorgen/*_test.go` both returned nothing before this phase — zero
pre-existing gorgias surface tests, so there was no risk of a second test file being missed the way
gong's `gong_full_surface_test.go` was flagged as a hazard in the plan.

## RED — cmd/connectorgen/gorgias_api_surface_test.go

Written against the target of **114 documented operations** (GET 46, POST 23, PUT 27, DELETE 18),
derived in the sweep-wide artifact pass (`.planning/phases/gorgias-parity-sweep-r1/PLAN.md` /
`RUN-STATE.json`) — **not re-derived here**, per the task instructions.

Asserted: `operation_ledger_version == 1`; exactly 114 endpoint rows; the per-method split; zero
blank dispositions; zero legacy `excluded` rows; exactly one disposition per row (`covered_by` XOR
`operation` XOR `excluded`); no duplicate `method+path`; every blocked (`operation`) row carries a
non-empty `reason`, a non-empty `source_url`, and a `notes` field prefixed `named_dependency=`; and
that a handful of specific endpoints (the 4 streams, the upload, the download) are present.

Modeled on `cmd/connectorgen/notion_api_surface_test.go` (single-disposition-per-row structure,
`named_dependency=` prefix check) and `cmd/connectorgen/gong_api_surface_test.go` (per-method map
assertions, duplicate-key detection). Unlike Notion (3 duplicate rows for split search arms),
Gorgias's 114 rows map 1:1 onto 114 unique `(method, path)` actions — no qualified/duplicate rows are
needed.

### Verbatim RED failure

To guarantee the failure was observed against a genuine pre-implementation state (not merely
asserted), the in-progress bundle authoring work (`operations.json`, `cli_surface.json`,
`writes.json`, and the modified `api_surface.json`/`metadata.json`/`docs.md`) was `git stash push -u`d
before writing and running the test, restoring the disk-committed 11-row/legacy-`excluded` baseline
described above. The test was then run for the first time against that restored baseline:

```
$ go test ./cmd/connectorgen/ -run TestGorgias -v
=== RUN   TestGorgiasAPISurfaceOperationLedger
    gorgias_api_surface_test.go:64: operation_ledger_version = 0, want 1
    gorgias_api_surface_test.go:67: api_surface declares 11 rows, want 114 documented operations
    gorgias_api_surface_test.go:120: 7 legacy excluded row(s) remain; operation_ledger_version mode requires operation rows
    gorgias_api_surface_test.go:123: covered(4)+blocked(0) = 4, want 114 rows
    gorgias_api_surface_test.go:133: GET: 9 rows, want 46
    gorgias_api_surface_test.go:133: POST: 1 rows, want 23
    gorgias_api_surface_test.go:133: PUT: 1 rows, want 27
    gorgias_api_surface_test.go:133: DELETE: 0 rows, want 18
    gorgias_api_surface_test.go:146: expected documented endpoint "GET /api/tickets"
    gorgias_api_surface_test.go:146: expected documented endpoint "GET /api/customers"
    gorgias_api_surface_test.go:146: expected documented endpoint "GET /api/messages"
    gorgias_api_surface_test.go:146: expected documented endpoint "GET /api/satisfaction-surveys"
    gorgias_api_surface_test.go:146: expected documented endpoint "POST /api/upload"
    gorgias_api_surface_test.go:146: expected documented endpoint "GET /api/{file_type}/download/{domain_hash}/{resource_name}"
--- FAIL: TestGorgiasAPISurfaceOperationLedger (0.00s)
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	0.505s
FAIL
```

A follow-up whole-package run (`go test ./cmd/connectorgen/`, no `-run`) confirmed
`TestGorgiasAPISurfaceOperationLedger` was the ONLY failure in the package (12.3s, all other tests
green) — the same whole-package discipline the plan requires for the gate phase, applied here too so
the red observation itself is not a targeted-run artifact.

After capturing this output, the stashed bundle-authoring work was restored (`git stash pop`) and
work continued toward GREEN.

## Planned assertions → GREEN evidence

| Assertion | Target | Evidence once green |
| --- | --- | --- |
| `operation_ledger_version == 1` | 1 | `api_surface.json` top-level field |
| Row count | 114 | generated `api_surface.json` |
| Method split | GET 46 / POST 23 / PUT 27 / DELETE 18 | generated `api_surface.json`, cross-checked against the fetched OpenAPI document independently in Python |
| Zero legacy `excluded` rows | 0 | every row carries `covered_by` or `operation`, never `excluded` |
| Zero blank dispositions | 0 | every row carries exactly one of `covered_by`/`operation` |
| No duplicate `method+path` | 0 | catalog cross-checked 1:1 against the spec's 114 unique operations before generation |
| Every blocked row has reason + source_url + `named_dependency=` note | 6 blocked rows | see `docs.md` Known limits and `api_surface.json`'s `operation` blocks |
| Streams / upload / download present | `GET /api/tickets`, `/customers`, `/messages`, `/satisfaction-surveys`, `POST /api/upload`, `GET /api/{file_type}/download/{domain_hash}/{resource_name}` | `covered_by.stream`/`covered_by.write`/`covered_by.direct_reads` rows |

## GREEN

Recorded once `go test ./cmd/connectorgen/` (whole package) passes with the authored bundle in place;
see `RUN-STATE.json`'s `tdd.green` field and the final report for the gate run transcript.
