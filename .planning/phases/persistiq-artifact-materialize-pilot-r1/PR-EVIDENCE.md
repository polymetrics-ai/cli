# Generator capability PR evidence

**Date:** 2026-08-08
**Scope:** generator policy and its PersistIQ pilot only. No eligible-392
connector surface was fetched, generated, or changed in this branch.

## Capability covered

`connectorgen batch materialize` now:

- maps every operation parsed from an OpenAPI 3.x or Swagger 2.0 artifact into
  `api_surface.json` and the operation catalog;
- preserves source-surface endpoints absent from that artifact with the exact
  `discrepancy=present-in-surface-absent-from-artifact` marker;
- emits `availability=not_implemented` plus a machine-checkable
  `named_dependency=<slug>` note where the runtime cannot execute the command;
- never labels an unavailable command `implemented`; and
- leaves runtime preflight to the single `batch gate` over the staged result,
  rather than running a repository-wide sweep during each materialization.

The parser accepts only OpenAPI 3.x or Swagger 2.0 documents. Unsupported
versions are rejected as non-artifacts. The strict-version regression test is
`TestParseBatchOpenAPIArtifactRequiresOpenAPI3OrSwagger2`.

## Red/green evidence

The captain-policy red tests and their green implementation are recorded in
[`TDD-LEDGER.md`](TDD-LEDGER.md). The focused green run was:

```text
go test -timeout 20m ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner
ok   polymetrics.ai/cmd/connectorgen
ok   polymetrics.ai/internal/connectors/engine
ok   polymetrics.ai/internal/connectors/commandrunner
```

No existing production connector bundle changed. The path audit is captured in
[`pr-evidence-2026-08-08/no-production-bundle-changes.txt`](pr-evidence-2026-08-08/no-production-bundle-changes.txt),
and the existing bundle corpus contains zero `not_implemented` commands before
this capability is consumed.

## Existing-corpus regression gates

These are one-time generator-PR checks over the unchanged embedded corpus:

| Gate | Result | Evidence |
|---|---|---|
| `connectorgen validate internal/connectors/defs --json` | 551 connectors, 0 findings, 0 warnings | [`all-551-validate.json`](pr-evidence-2026-08-08/all-551-validate.json) |
| `connectorgen surface-sync --check` | 551 scanned, 0 fields filled, 0 corrected | [`all-551-surface-sync.txt`](pr-evidence-2026-08-08/all-551-surface-sync.txt) |
| `TestEveryImplementedCommandPassesRuntimePreflight` | pass | [`all-551-runtime-preflight.txt`](pr-evidence-2026-08-08/all-551-runtime-preflight.txt) |
| `go vet ./cmd/connectorgen ./internal/connectors/engine ./internal/connectors/commandrunner` | pass | command exit 0 |
| `go build ./cmd/pm` | pass | command exit 0 |
| `go run ./cmd/agentcontractgen check` | pass | canonical contract current |

The changed code adds no behavior to existing bundles: the new availability
and discrepancy fields are absent from all existing connector definitions, and
the all-551 gates above remain green.

## PersistIQ pilot rerun

The fresh artifact was OpenAPI 3.0.1, 47,796 bytes, SHA-256
`0bf3e1ecbfbf6215360b5bb8f9d4fda816df4e1872470a00b529fb3e8b80946f`.

| Measure | Result |
|---|---:|
| Artifact operations mapped | 21 |
| ETL / direct_read / reverse_etl / direct_write / binary_download / unclassified | 11 / 1 / 7 / 2 / 0 / 0 |
| Implemented commands | 21 |
| Named-dependency commands | 3 |
| Flagged discrepancies | 3 |
| Reachable real-binary command paths | 24/24 |
| Implemented commands reachable | 21/21 |
| Failed candidates | 0 |

Wall-clock timings from the rerun:

| Step | Time |
|---|---:|
| Identify artifact link | 0.03s |
| Map 21 operations | 0.03s |
| Fetch, digest, parse | 2.70s |
| Materialize, static gates, runtime preflight, binary reachability | 50.07s |
| Report collation | 0.09s |
| **Total** | **52.92s** |

The complete operation map, artifact, generated bundle, gate reports, binary
reachability results, and timing files are under
[`rerun-2026-08-08/`](rerun-2026-08-08/). PersistIQ was implemented according
to static evidence, **not certified**, and **never exercised against the
provider**; no credentials were used.

## Generalization validation before merge

At captain direction, three eligible, deliberately different shapes were
validated as staged evidence only; no generated production connector bundle
was added:

| Connector | Shape | Mapped | Implemented | Named dependency | Discrepancy | Reachable | Result |
|---|---|---:|---:|---:|---:|---:|---|
| watchmode | 23-read OpenAPI 3.0.3 | 23 | 13 | 32 | 22 | 45/45 | pass |
| docuseal | 7-read/16-write OpenAPI 3.1.0 | 0 | 0 | 0 | 0 | 0 | **failed: 11 top-level webhooks rejected as artifact inventory unknown** |
| float | 44-read/51-write Swagger 2.0 | 0 | 0 | 0 | 0 | 0 | **failed: external path-item reference not exhaustively resolvable** |

Watchmode mapped all 23 artifact operations as `direct_read` (0 ETL, 0
reverse_etl, 0 direct_write, 0 binary_download, 0 unclassified). Its 22
source-surface-only rows were retained with
`present-in-surface-absent-from-artifact`, proving the discrepancy path does
not refuse the connector. `connectorgen validate` had 0 findings,
`surface-sync --check` had no drift, batch gate included 1/1 with 13 runtime
preflightable commands, and the real binary reached all 45 command help
paths (13 implemented and 32 visible not-implemented commands).

Watchmode validation timings were: identify 0.04s; fetch/digest 2.52s; map
0.02s; batch plan 1.78s; materialize/parse 0.65s; validate 0.67s;
surface-sync derive/check 0.68s/0.64s; batch gate 0.66s; existing-corpus
runtime-preflight regression 5.28s; staged binary build 9.71s; bare namespace
2.48s; 45-command reachability 54.70s; report 0.06s; total 79.89s.

The complete evidence, artifact hashes, failure reports, mapping, generated
staged bundle, and reachability TSV are under
`.planning/phases/persistiq-artifact-materialize-pilot-r1/generalization-validation-2026-08-08/`.
The suggested Web Scraper artifact was fetched but not selected because the
ledger marks it `partner_gated` and the existing planner correctly refuses
non-public candidates. Ding Connect returned HTTP 403 twice and was replaced
by Float for the Swagger-2 attempt.

**Generalization result: NOT READY.** DocuSeal and Float fail before mapping,
so the generator capability cannot be called generalized or ready. The
eligible 392 remain untouched. Certification remains withheld; no provider
operation was exercised.

## Deferred work

The eligible 392 run is a separate follow-up after this generator capability
lands. It is intentionally absent from this PR.
