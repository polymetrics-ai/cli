# Phase 600 verification checklist

**Issue:** #3984  
**Status:** planned

## Capability matrix

- [ ] Function kinds are AST-derived from the engine and connector runtime
      contracts; a new source kind appears as a matrix row.
- [ ] Every registry/bundle connector has every function-kind row.
- [ ] PostgreSQL and MySQL write rows report `implemented=false`.
- [ ] Applicable cells without accepted live evidence cannot complete a
      connector.
- [ ] Every non-applicable cell has a non-generic code and explanation.
- [ ] Existing `certification.json` filenames are reported as legacy inputs,
      not accepted live evidence.
- [ ] The generated capability artifact is deterministic and checkable.

## Pair-flow matrix

- [ ] Each connector exposes all four API/database endpoint roles with reasons
      for roles that do not apply.
- [ ] Every flow key includes source, destination, and flow kind.
- [ ] API destinations without durable acknowledgement are not implemented.
- [ ] A passed flow evidence record requires an independent destination
      readback reference.
- [ ] Final connector certification requires all applicable function and flow
      cells.
- [ ] Capability and flow baselines are reported separately from generated
      summaries.

## Local gates

- [ ] `go test -timeout 20m ./cmd/connectorgen`
- [ ] `go test -timeout 20m ./internal/connectors/...`
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pm`
- [ ] `go run ./cmd/connectorgen certification-matrix --check`
- [ ] Individual `make verify` sub-gates required by `AGENTS.md`
- [ ] Inline `verify-work` and `code-review` records completed
