# Verification Checklist — Zoom SCIM2 parity, R1

## Lifecycle

- [x] GSD command provenance was resolved with `scripts/gsd sources`.
- [x] Required skills and canonical connector/CLI references are recorded in `PLAN.md`.
- [x] Live provider artifact URL, retrieval date, HTTP result, byte count, SHA-256, and exact
  eleven-operation audit are recorded before RED.
- [x] RED failure captured verbatim before production changes and pushed in `a415f53b4`; the later
  literal-root-redaction RED was pushed separately in `fdcea0a23`.
- [x] Required reusable foundations red/green tested and separated from connector authoring:
  operation-scoped direct-read origin/auth (`027bb66f4`), named root-object mapping (`543b1f3d9`),
  and literal-root JSON-member redaction (`9542b444c`) are pushed.
- [x] Connector declaration, generated output, docs, and website catalog are ready for this
  SCIM2 slice commit.
- [x] Inline verify-work and manual code-review evidence are complete under the documented
  single-worker fallback.

## Source parity

- [x] All eleven SCIM2 ledger rows have exactly one executable disposition (`4 direct_read`, `7
  direct_write`), confirmed by scoped `surface-reconcile --check`.
- [x] Zero Zoom rows are `unsafe_or_disallowed` (audited with the API-surface JSON query).
- [x] Every command accepts only documented typed path/body members; no paging flags are invented.
- [x] Each 204 action proves status-only success and no request body when none is declared.
- [x] User/group data is redacted from previews, errors, and results according to declared policy,
  including literal SCIM extension URN keys that contain dots.

## Runtime/docs checks

- [x] Focused engine/commandrunner/app/Zoom tests and vet pass; `make lint` reports 0 issues.
- [x] Fresh binary passes `pm help zoom`, bare `pm zoom`, bare `pm zoom scim2`, and every exact
  SCIM2 command `--help`; all eleven routes are reachable.
- [x] Isolated fixtures prove declared root origin/auth, exact method/path/body/status,
  destructive confirmation, no paging input, redaction, and no-body semantics.
- [x] Surface sync/reconciliation/validation, Zoom-only endpoint-ledger delta, docs validation,
  generated website data (twice idempotent), website typecheck, CLI tests, and scoped Make gates
  pass.

## verify-work evidence

```text
go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/...
go test -count=1 -timeout 20m ./internal/connectors/commandrunner ./internal/connectors/engine ./cmd/connectorgen
go test -count=1 -timeout 20m ./internal/cli
go vet ./internal/connectors/commandrunner ./internal/connectors/engine ./internal/connectors/defs/zoom ./cmd/connectorgen ./internal/cli
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
```

All commands passed. Website lint emitted 13 pre-existing warnings and no errors; no warning is in
this slice's generated data.

## Manual code-review evidence

The canonical runtime cannot register this provider-category phase and the parent contract forbids
spawning roles, so `code-review` was performed inline after resolving its GSD sources. Review of the
final diff found no blocking issue:

- every new command binds exactly one live-artifact method/path and is preflighted through the real
  command runner;
- every SCIM transport declaration fixes the provider root and bearer secret in the bundle, with no
  caller-supplied URL/header/body escape hatch;
- deletes have destructive confirmation, and all documented 204 paths return status-only success;
- root object schemas remain operation-scoped, named, bounded, and redacted; the extension-key
  redaction regression has a focused RED/GREEN test;
- generated docs/site output was regenerated, unrelated connector manuals restored wholesale, and
  connector-boundary passes.
