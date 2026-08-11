# Verification checklist — Issue #3974: typed database connector foundation

## Acceptance proof

- [x] Invalid/unknown manifests, logical types, resource policies, and driver
      declarations fail closed with secret-safe errors.
- [x] Definitions, catalogs, logical types, and read-plan projections are
      defensive / immutable from caller mutation.
- [x] Driver declaration alone cannot pass; unknown/mismatched driver,
      protocol, API version, evidence, and source-runner cases fail correctly.
- [x] Type compatibility permits only exact/lossless mappings and never falls
      back to string.
- [x] Source and target references preserve workspace, connector, and
      connection IDs without credential fields.
- [x] Structured catalog, key, fingerprint, deterministic read plan, bounds,
      and cancellation behavior have focused unit coverage.
- [x] PostgreSQL has only a compile-time reference driver seam and its public
      `write` / `cdc` capabilities remain false.
- [x] Warehouse mediation amendment: shared artifact/owner identity, isolated
      database inbound/outbound legs, per-leg descriptor admission,
      mode-declaration refusal, and MySQL layer-two conformance proof.
- [x] Red/Green TDD evidence and manual GSD lifecycle evidence are complete for
      the original slice and the captain amendment.

## Local gates

- [x] `gofmt -w internal/warehouse internal/connectors/database internal/synccontract internal/connectors/engine internal/connectors/native/postgres`
- [x] focused database, synccontract, engine, native-postgres tests with `-count=1`
- [x] race tests for `internal/warehouse`, `internal/connectors/database`, and `internal/synccontract`
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
- [x] generated `verify-work` and `code-review` prompts recorded with the
      manual inline fallback.

The captain amendment touches `internal/warehouse` and
`internal/connectors/database`; its focused, race, static/build, CLI, and
required make-gate checks were rerun and passed.

## Deliberate non-goals / parity

No user-facing command, flag, output, help topic, generated manual, website
page, executable write action, or CDC surface changes in this issue. CLI/docs/
website parity is therefore not applicable to the implementation itself. The
existing PostgreSQL metadata is tested to prevent an accidental user-visible
capability promotion.

## Evidence log

- **Red:** `traces/red-run.txt` preserves the failing test run made before any
  production implementation.
- **Focused Green:** the four changed/adjacent packages passed together after
  implementation and again after lint-driven hygiene refinements; the initial
  output is in `traces/green-run.txt`.
- **Direct new-package lint:**
  `golangci-lint run ./internal/connectors/database/... ./internal/synccontract/...`
  passed after fixing its initial three findings.
- **Repository gates passed:** `go vet ./...`, `go build ./cmd/pm`,
  `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`,
  `make agent-contract-check`, `make connectorgen-validate`,
  `make connectorgen-surface-sync`, `make connector-boundary`, and
  `make release-workflow-check`.
- **CLI regression:** `go test -timeout 20m ./internal/cli -count=1` completed
  successfully as a standalone final gate.
- **GSD verification/review:** generated `verify-work` and `code-review`
  prompts were executed through the documented manual inline fallback. The
  coverage-backed UAT result is in `UAT.md`; review findings and dispositions
  are in `REVIEW.md`.
- **Captain amendment Red/Green:**
  `traces/warehouse-layer-red-run.txt`,
  `traces/warehouse-owner-identity-red-run.txt`, and
  `traces/warehouse-multi-admission-red-run.txt` were captured before their
  respective production seams existed. The complete five-package Green result
  is preserved in `traces/warehouse-layer-green-run.txt`.
- **Captain amendment local validation:** focused and race tests, `go vet`,
  `go build`, direct shared-package lint, `go test ./internal/cli`, and all
  individual required make gates passed after the per-leg admission refinement.
  `go list -deps` also confirmed that `internal/warehouse` and
  `internal/connectors/database` import neither the PostgreSQL nor MySQL native
  package.
