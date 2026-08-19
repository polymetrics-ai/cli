---
phase: 600
plan: 01
status: complete
issue: 3984
completed: 2026-08-10
---

# Plan 600-01 summary: truthful capability baseline

## Delivered

- Added `connectorgen certification-matrix`, an AST- and runtime-derived
  generator with exact-byte `--check` drift detection.
- Generated and committed the authoritative per-connector capability matrix;
  its only live-proof input is the strict
  `internal/connectors/certifications/evidence/*.json` namespace.
- Made the matrix reject generic N/A reasons and native
  `ErrUnsupportedOperation` stubs. PostgreSQL and MySQL Write appear as
  applicable `declared=false`, `implemented=false` rows.
- Added the focused Make gate, proof-bearing evidence writer, local salt
  boundary, and architecture documentation. The new generated status is
  surfaced by connector inspection as CERTIFIED or COMMUNITY BUILD,
  UNCERTIFIED; it never blocks a reachable connector.
- Updated connector help, generated CLI documentation, website documentation,
  and golden transcripts for that visible quality signal.

## TDD evidence

- **Red:** `TestCertification*` failed to compile before the generator existed;
  the full compiler transcript is retained in `TDD-LEDGER.md`.
- **Green:** `go test -timeout 20m ./cmd/connectorgen -run 'TestCertification'
  -count=1` passed in 92.786s.
- **Drift:** `make connectorgen-certification-matrix` passed with the generated
  output current.

## Generated baseline

The baseline has **556 connectors**, **21 discovered function kinds**, and
**0 capability-complete connectors**. All `live_tested` and `complete` totals
are zero: no legacy filename was promoted to proof.

| Function kind | Applicable | Declared | Implemented | Fixture tested | Live tested | Complete |
|---|---:|---:|---:|---:|---:|---:|
| capability:catalog | 556 | 556 | 556 | 0 | 0 | 0 |
| capability:cdc | 2 | 0 | 2 | 0 | 0 | 0 |
| capability:check | 556 | 555 | 556 | 463 | 0 | 0 |
| capability:dynamic_schema | 6 | 6 | 0 | 0 | 0 | 0 |
| capability:query | 1 | 1 | 0 | 0 | 0 | 0 |
| capability:read | 556 | 555 | 556 | 431 | 0 | 0 |
| capability:write | 279 | 242 | 272 | 208 | 0 | 0 |
| operation:binary_download | 12 | 12 | 12 | 0 | 0 | 0 |
| operation:browser_open | 0 | 0 | 0 | 0 | 0 | 0 |
| operation:composite | 1 | 1 | 0 | 0 | 0 | 0 |
| operation:file_upload | 5 | 5 | 0 | 0 | 0 | 0 |
| operation:graphql_mutation | 1 | 1 | 1 | 0 | 0 | 0 |
| operation:graphql_query | 1 | 1 | 0 | 0 | 0 | 0 |
| operation:local_file | 0 | 0 | 0 | 0 | 0 | 0 |
| operation:local_git | 1 | 1 | 0 | 0 | 0 | 0 |
| operation:provider_search | 0 | 0 | 0 | 0 | 0 | 0 |
| operation:rest_read | 23 | 23 | 23 | 0 | 0 | 0 |
| operation:rest_write | 6 | 6 | 6 | 0 | 0 | 0 |
| operation:stream_etl | 1 | 1 | 0 | 0 | 0 | 0 |
| operation:xml_export | 0 | 0 | 0 | 0 | 0 | 0 |
| operation:xml_import | 0 | 0 | 0 | 0 | 0 | 0 |

The legacy inventory is retained for the captain’s later decision: **17**
filename matches are ignored (**11** bundle contracts; **6** fixture/schema
files). It is not a source of certification evidence.

## Follow-on

Plan 600-02 extends this same generated command with the separately reported
flow, workflow, and sync-mode scoreboards. Per captain clarification, every
candidate is a source → local Parquet warehouse → destination round trip, may
use the same connector on both ends, and carries explicit delivery guarantees
separate from whether the round trip worked.
