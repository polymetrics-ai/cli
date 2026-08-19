# Phase 600 verification checklist

**Issue:** #3984  
**Status:** verified locally; automated PR review pending

## Capability matrix

- [x] Function kinds are AST-derived from the engine and connector runtime
      contracts; a new source kind appears as a matrix row.
- [x] Every registry/bundle connector has every function-kind row.
- [x] PostgreSQL and MySQL write rows report `implemented=false`.
- [x] GitHub `operation:graphql_query` reports `implemented=true` through the
      GraphQL direct-read executor, while GitHub `operation:local_git` remains
      an applicable `implemented=false` cell.
- [x] Applicable cells without accepted live evidence cannot complete a
      connector.
- [x] Every non-applicable cell has a non-generic code and explanation.
- [x] Existing `certification.json` filenames are reported as legacy inputs,
      not accepted live evidence.
- [x] The generated capability artifact is deterministic and checkable.

## Pair-flow matrix

- [x] Each connector exposes all four API/database endpoint roles with reasons
      for roles that do not apply.
- [x] Every flow key includes source, warehouse mediator, destination, and flow
      kind, including GitHub-to-GitHub.
- [x] API destinations without durable acknowledgement are not implemented.
- [x] A passed flow evidence record requires an independent warehouse and destination
      readback reference.
- [x] Final connector certification requires all applicable function, workflow,
      sync-mode, and flow cells.
- [x] Capability and flow baselines are reported separately from generated
      summaries.

## Workflow and sync-mode scoreboard

- [x] ETL, reverse ETL, flow authoring, and schedule each have an explicit
      generated workflow cell.
- [x] Every mode from synccontract.AllModes() is crossed with the stable four
      warehouse-facing primitive identifiers.
- [x] Change capture is applicable only to database read into the warehouse.
- [x] PostgreSQL and MySQL database write cells are applicable but
      implemented=false; no database-write executor is inferred.
- [x] Uncertified connector inspection visibly reports COMMUNITY BUILD,
      UNCERTIFIED without changing reachability.

## Local gates

- [x] Focused certification tests: `go test -timeout 20m ./cmd/connectorgen
      -run '^TestCertification' -count=1`
- [x] Generated transcript test: `go test -timeout 20m ./internal/cli -run
      '^TestGoldenTranscripts$' -count=1`
- [x] Changed-package tests, vet, formatting, and build gates
- [x] Generator drift check is current
- [x] Individual Make verify sub-gates required by AGENTS.md
- [x] Inline verify-work and code-review records completed

## Recorded commands

- Focused and full connectorgen tests; full internal/cli, engine, and status
  package tests.
- Go vet for changed packages, gofmt diff for changed Go files, and builds of
  cmd/connectorgen and cmd/pm.
- tidy-check, lint, docs-check, smoke-no-build, agent-contract-check,
  connectorgen-validate, connectorgen-surface-sync,
  github-parity-artifacts-check, connectorgen-certification-matrix,
  connector-boundary, and release-workflow-check.
- scripts/verify-gsd-workflow as the final PR-lifecycle guard.
