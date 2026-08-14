# Verification checklist — Issue #3976

## Acceptance proof

- [x] RED: `dynamic_catalog_integration_test.go` was committed before
      production implementation and its exact expected compiler failure is in
      `traces/dynamic-catalog-red.txt`.
- [ ] Live typed catalog discovery proves two distinct PostgreSQL schema
      fixtures yield different correct catalogs by independent PostgreSQL
      catalog oracle inspection.
- [x] Configured database/schema/table identity and deterministic relation,
      column, and key ordering are retained.
- [x] Columns include nullability, native type identity/modifiers, supported
      logical type, ordinal, and ordered primary-key membership.
- [x] Unsupported or unsafe native shapes fail closed with named secret-safe
      errors rather than coarse static/string fallback.
- [x] The #4034 typed catalog foundation is used at PostgreSQL's shipping
      runtime boundary; no second disconnected static catalog remains.
- [x] Static behavior owned by #3980/#3982/#3983/#3987 is recorded and not
      changed by this child.
- [x] PostgreSQL `write`, `query`, and `cdc` capability truth remains unchanged.

## Planned local gates

- [x] focused PostgreSQL catalog unit tests
- [ ] opt-in live `dbtest` PostgreSQL catalog tests with independent oracle:
      compiled and visibly skipped because no explicit direct Podman endpoint
      or `POLYMETRICS_DATABASE_INTEGRATION=1` opt-in is configured
- [x] `go test -race -timeout 20m ./internal/connectors/native/postgres`
- [x] `go test -timeout 20m ./internal/cli -count=1`
- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `make tidy-check`
- [x] `make lint`
- [x] `make docs-check`
- [x] `make smoke-no-build`
- [x] `make agent-contract-check`
- [x] `make connectorgen-validate`
- [x] `make connectorgen-surface-sync`
- [x] `make connector-boundary`
- [x] `make release-workflow-check`
- [x] generated `verify-work` and `code-review` manual-inline records
- [ ] no-mistakes child pipeline without `--yes`, maximum five fresh correction
      loops

## Live-proof resumption

- [x] RED trace proves the new live full/cursor test cannot use the obsolete
      Podman-only harness wiring.
- [x] PostgreSQL dbtest uses an explicit `docker` or `podman` direct Unix
      endpoint and a pinned capacity probe, never a global runtime default.
- [x] A real catalog discovery reports the seeded table's native/logical
      metadata; a full read returns all asserted rows; a cursor-advanced read
      returns only its asserted rows.
- [x] The live proof records the actual no-cursor, nonexistent-column,
      nullable-cursor, and two-table connection-level cursor outcomes.
- [x] The historical CDC integration test is recorded as intentionally
      fail-closed/skipped, not treated as live success or re-enabled.

## Delivery holds

- [x] Correctly stacked draft child PR #4065 exists with `Refs #3976` and
      `Refs #3972`, targets `feat/3972-postgres-parity`, and has been safely
      merged onto #4064 head `c2e013324` without force/discard.
- [ ] Parent integration remains held until corrected #4058 is green and merged.
- [ ] Automated review coverage is recorded and every actionable finding is
      dispositioned before parent-branch integration.

## CLI/docs/website applicability

No new command, flag, or help topic is added. The existing PostgreSQL connector
description changed from `information_schema` to PostgreSQL system-catalog
discovery, so its defs documentation, generated connector manuals/catalog, and
website connector data were regenerated. Connector docs validation remains a
required final gate.

## Final local outcomes

- Focused package run: PostgreSQL, `database`, and `engine` all passed.
- PostgreSQL race run passed; `internal/cli` passed in 162.331 seconds.
- `go vet ./...`, `go build ./cmd/pm`, lint, docs, smoke, agent-contract,
  connector validation/surface sync/boundary, and release-workflow gates passed.
- The tagged live regression was invoked and skipped visibly before startup:
  this workspace has neither the explicit opt-in nor an explicit direct local
  Podman endpoint. It remains a delivery hold for a live-proof claim, not a
  passing integration result.
