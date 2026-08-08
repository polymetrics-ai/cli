# TDD Ledger — Zoom Quality Management documented-operation parity, R1

## RED — planned

The red checkpoint must change only `internal/connectors/defs/zoom/command_surface_test.go` and
synthetic Quality Management fixtures. It asserts the documented six-operation target before any
Zoom production declaration changes:

- covered operations: `12 → 18`
- locally blocked: `1830 → 1824`
- direct reads: `8 → 13`
- writes: `1 → 2`
- five fixed GET paths, two required detail IDs, no invented paging flags, response redaction, and
  one exact POST JSON body with a successful fixture `201`.

Run `go test -count=1 ./internal/connectors/defs/zoom/...`, capture its failure verbatim below,
then commit and push that red state before creating `operations.json`, `writes.json`, metadata,
source coverage, or documentation production changes.

## GREEN — pending

Record the focused test, surface/validator, binary, docs, and review evidence after the declarations
exist and the red test becomes green.
