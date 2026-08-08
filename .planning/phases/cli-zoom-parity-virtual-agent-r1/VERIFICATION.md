# Verification Checklist — Zoom Virtual Agent parity, R1

## Lifecycle

- [x] GSD command provenance was resolved with `scripts/gsd sources`.
- [x] Required skills and canonical connector/CLI references are recorded in `PLAN.md`.
- [x] Live provider artifact URL, retrieval date, HTTP result, byte count, SHA-256, and exact
  thirteen-operation audit are recorded before RED.
- [x] Test-only RED failure was captured verbatim before production declaration changes and pushed
  in `cd9cfaa96`.
- [x] No missing engine foundation was discovered; ordinary typed direct read/write contracts cover
  the full provider category without broadening transport capability.
- [x] Connector declaration, generated output, docs, website catalog, fixture lifecycle, and
  inline verify-work/manual code-review evidence are complete.

## Source parity

- [x] All thirteen live-artifact endpoints match exactly one reconciled command: nine direct reads
  and four direct writes. A method/path set diff against the fetched artifact is empty.
- [x] `surface-reconcile --check --notes-contains provider_module=virtual-agent` is clean after
  reconciling exactly thirteen rows.
- [x] The global endpoint-ledger delta contains nine new `rest_read` entries under the `zoom` key;
  the SHA-256 hash of every non-Zoom ledger entry is unchanged.
- [x] Zero Zoom rows are `unsafe_or_disallowed` (audited from `api_surface.json`).
- [x] Every article body field comes from the published request schema; no response-only page/date/
  token value becomes a request flag.
- [x] Delete asserts documented status-only `204` success and typed destructive confirmation;
  create-sync asserts a no-body POST.
- [x] Article, knowledge-base, sync, engagement, consumer, query, transcript, survey, variable,
  operator, report, and generic token fields are redacted in tested preview/result paths.

## Runtime/docs checks

- [x] Focused Zoom, engine, commandrunner, connectorgen, CLI, and vet checks pass; `make lint`
  reports 0 issues.
- [x] Fresh binary passes `pm help zoom`, bare `pm zoom`, bare `pm zoom virtual-agent`, and every
  exact Virtual Agent command `--help`; all thirteen routes are reachable.
- [x] Isolated fixtures prove exact fixed method/path/auth/body/status, no invented query/paging
  input, redaction, no-body sync semantics, and destructive deletion confirmation.
- [x] Surface sync/reconciliation/validation, Zoom-only endpoint-ledger delta, docs validation,
  generated website data (twice idempotent), website typecheck, and scoped Make gates pass.

## verify-work evidence

```text
go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/...
go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/... -run '^TestVirtualAgent'
go test -count=1 -timeout 20m ./internal/connectors/commandrunner ./internal/connectors/engine ./cmd/connectorgen
go test -count=1 -timeout 20m ./internal/cli
go vet ./internal/connectors/commandrunner ./internal/connectors/engine ./internal/connectors/defs/zoom ./cmd/connectorgen ./internal/cli
go run ./cmd/connectorgen validate
go run ./cmd/connectorgen surface-sync
go run ./cmd/connectorgen surface-sync --check
go run ./cmd/connectorgen surface-reconcile --notes-contains provider_module=virtual-agent
go run ./cmd/connectorgen surface-reconcile --check --notes-contains provider_module=virtual-agent
go run ./cmd/pm docs generate --dir docs/cli --connectors-dir docs/connectors
make tidy-check
make docs-check
make smoke-no-build
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-boundary
make release-workflow-check
make lint
cd website && pnpm run gen:website-data && pnpm run gen:website-data
cd website && pnpm run typecheck && pnpm run lint
go build -o .tmp/pm ./cmd/pm
```

All commands passed. Website lint emitted 13 pre-existing warnings and no errors; no warning is in
this slice's generated data.

## Manual code-review evidence

The canonical runtime cannot register this provider-category phase and the parent contract forbids
spawning roles, so `code-review` was performed inline after resolving its GSD sources. Review of
the final diff found no blocking issue:

- every new command binds exactly one live-artifact method/path and is preflighted through the real
  command runner;
- all request inputs are fixed typed path/body members from the provider schema; no generic URL,
  header, HTTP, shell, or JSON-body input exists;
- mutation lifecycle requires plan, no-network preview, explicit approval, and execution; delete
  additionally requires typed destructive confirmation, and no-content is status-only;
- all response-bearing commands use redacted policies with provider-specific and generic secret
  field protection;
- generated docs/site output was regenerated, unrelated generated connector manuals were restored
  wholesale, the non-Zoom catalog and endpoint-ledger hashes are unchanged, and connector-boundary
  passes.
