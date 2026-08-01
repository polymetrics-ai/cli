# TDD LEDGER — issue #3247 Marketo parity

## Red

Command:

```bash
gofmt -w cmd/connectorgen/marketo_full_surface_test.go && go test ./cmd/connectorgen -run TestMarketo -count=1
```

Result: failed as expected before production connector edits. Current Marketo bundle had no `writes.json`, no full-surface command/operation metadata, and `metadata.capabilities.write=false`.

Failure excerpt:

```text
read ../../internal/connectors/defs/marketo/writes.json: no such file or directory
Marketo capabilities read/write = true/false, want true/true
```

## Green

Commands now pass after the Marketo definition expansion and safety fixes:

```bash
go test ./cmd/connectorgen -run TestMarketo -count=1
go run ./cmd/connectorgen validate internal/connectors/defs --json
```

Current asserted Marketo parity counts:

- 327 official AdobeDocs Swagger operations.
- 117 fixture-backed ETL/changefeed streams.
- 28 bounded redacted direct reads.
- 158 typed reverse-ETL write actions.
- 24 blocked/not-applicable operation rows.
- 303 CLI commands: 117 ETL, 28 direct reads, 158 reverse ETL.

## Safety regression tests added

`cmd/connectorgen/marketo_full_surface_test.go` now asserts:

- Marketo read/write capabilities are enabled without enabling generic query.
- Full operation ledger counts remain reconciled.
- Write paths do not embed query strings; write-query operations are blocked until a typed write query map exists.
- Destructive writes require `confirm: destructive`, required path/body selectors, and required array item selectors.
- Response-only fields such as `errors`, `reasons`, `seq`, and operation-status fields are absent from write schemas.
- CLI flags are unique, kebab-case, and map to declared schemas.
- ETL/direct path-parameter commands include explicit config/path guidance in examples/notes.

## Refactor / review fixes

Automated reviewer findings were dispositioned by blocking unsafe query-selector writes (`merge_leads`, bulk imports, query-only asset writes, associate lead, remove leads from list), tightening destructive schemas, pruning response/result fields from write schemas, fixing duplicate CLI flags, and adding path-parameter guidance.
