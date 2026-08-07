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

`go test ./cmd/connectorgen/ -run TestGorgias -v` now passes on the first run against the authored
bundle:

```
=== RUN   TestGorgiasAPISurfaceOperationLedger
--- PASS: TestGorgiasAPISurfaceOperationLedger (0.00s)
PASS
ok  	polymetrics.ai/cmd/connectorgen	0.487s
```

Full gate transcript (all whole-package, never a targeted `-run`, per the plan):

- `go run ./cmd/connectorgen validate internal/connectors/defs/gorgias` → 0 findings (one fix round:
  `update_custom_fields`/`update_customer_custom_field_values`/`update_ticket_custom_fields` needed a
  `body_schema` alongside `body_type: "json_array"`).
- `go test ./cmd/connectorgen/` (whole package) → PASS.
- `go run ./cmd/connectorgen surface-sync` (write mode, then `--check`) → regenerated
  `internal/connectors/defs/operation_endpoint_ledger.json` (purely additive, gorgias-only content
  verified by object diff), 0 field(s) filled/corrected in `cli_surface.json`/`operations.json`
  (the hand-authored content already matched what the tool derives).
- `go test ./internal/connectors/commandrunner/ -run TestEveryImplementedCommandPassesRuntimePreflight`
  → PASS across all 1483 implemented commands repo-wide (this failed with 42 gorgias findings before
  `surface-sync` regenerated the endpoint ledger — expected, since the ledger is what the runtime
  checks against).
- `go test ./internal/connectors/conformance/` (whole package) → PASS.
- `make connector-boundary` → `"outcome": "clean"`.
- `make docs-check`, `make lint` (0 issues), `make tidy-check`, `make agent-contract-check`,
  `make smoke-no-build`, `make release-workflow-check` → all pass.
- `gofmt -l cmd internal` and `go vet ./cmd/connectorgen/ ./internal/connectors/...` → clean.
- Full `go test ./internal/cli/...` → PASS (589.9s), including `TestGoldenTranscripts` after
  regenerating `testdata/golden_transcripts.json` with `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1`
  (diff read first: the only change across 9 subtests was the new `pm gorgias <command> - Gorgias:
  ...` line appearing in the root "CONNECTOR COMMANDS" listing now that gorgias has a
  `cli_surface.json`; verified word-diff-clean of anything else).
- Binary built and exercised: `pm connectors inspect gorgias --json`, bare `pm gorgias`, and
  `--help` on a stream/direct-read/write/binary-download command all exit 0; a scripted pass over
  all 108 `cli_surface.json` commands confirmed every one resolves `--help` with exit 0.
- Website catalog regenerated (`gen-connector-bundles.mjs`, `gen-connector-catalog.mjs`,
  `gen-connectors.mjs`); the 3 touched files' diffs were checked **by object**, not by line — only
  the `gorgias` entry differs in `connectors.generated.json` and
  `connectors.catalog.data.generated.json`; `connectors.catalog.generated.ts`'s single-line diff is
  the aggregate write-capability counter (234→235).
- `docs/cli` and `docs/connectors` regenerated via `./pm docs generate --dir docs/cli`; reverted
  every non-gorgias path under `docs/connectors/` (1031 files of pre-existing `main` drift, matching
  the plan's ~1,034-file warning), keeping only `docs/connectors/gorgias/{MANUAL.md,SKILL.md}`.

See `RUN-STATE.json`'s `tdd.green`/`tdd.green_evidence` fields for the compact form.
