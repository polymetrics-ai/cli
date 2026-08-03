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

## No-mistakes review findings to fix

Run `01KYXSJ0G64JJ9VV9RPN0YAR0Y` surfaced 5 source-backed findings before its agent fix failed with a Pi WebSocket error:

- `marketo-next-page-token`: add the documented initial `nextPageToken` requirement for activity/change/deleted-lead streams.
- `marketo-query-flags-unusable`: make required-query stream config templates compatible with CLI query overrides.
- `marketo-write-content-types`: align multipart/form/form body types for the cited write actions.
- `marketo-destructive-id-types`: type destructive integer selectors as integers.
- `marketo-write-fixtures-missing`: add fixture-only write request-shape fixtures for every Marketo write action.

Manual follow-up fixes were applied on top of the branch per no-mistakes terminal-failure guidance, since the review agent-fix step failed on a transport error (WebSocket) after surfacing findings rather than on a code defect. All five findings above are addressed. While verifying the fixes, `go run ./cmd/connectorgen validate` surfaced one additional finding introduced by the `marketo-next-page-token` fix: the three affected stream entries carried a top-level `description` field that is not part of the `streams.json` meta-schema (`additional property not allowed`). Fixed by removing the field from `get_lead_activities`, `get_deleted_leads`, and `get_lead_changes`; the equivalent guidance is already carried in the schema-valid `cli_surface.json` per-command `notes` and flag `summary` fields, which were already updated. `connectorgen validate` now reports 0 findings across all 550 connectors.

## Recovery-session re-verification

Recovered from the failed run `01KYXSJ0G64JJ9VV9RPN0YAR0Y` in a follow-up session: independently re-verified every one of the five fixes against the working-tree diff (nextPageToken added to `get_lead_activities`/`get_deleted_leads`/`get_lead_changes`, query-flag templates switched to `omit_when_absent` objects, `create_email_template`/`create_email_full_content`/`update_content`/`create_file` now `multipart` and `delete_folder` now `form`, destructive integer selectors retyped, and 158 `fixtures/writes/<action>.json` files present), then reran the full local gate set: `gofmt`, `go vet ./...`, `go build ./cmd/pm`, `go run ./cmd/connectorgen validate` (0 findings/550 connectors), the three `TestMarketo*` tests, `internal/connectors/conformance`, the full `internal/connectors/...` suite, `go run ./cmd/connectorgen boundary . --json` (clean), `pm docs generate --dir docs/cli` + `pnpm run gen:website-data` (regeneration produced only genuine Marketo-scoped diffs; 448 files of unrelated connector-doc drift from stale generator state were reverted), `git diff --check`, and `make verify` end-to-end (fmt/vet/test/build/docs-check/smoke/lint 0 issues/connectorgen-validate/connector-boundary/release-workflow-check all green). Ready for a fresh `no-mistakes axi run`.
