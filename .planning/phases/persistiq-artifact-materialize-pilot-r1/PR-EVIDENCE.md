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

## Deferred work

The eligible 392 run is a separate follow-up after this generator capability
lands. It is intentionally absent from this PR.
