# Connector Rollout Validation Gates

These gates are mandatory for every connector rollout slice. A slice is not integrated until all
applicable gates pass. The coordinator owns merge validation; workers report gate evidence in the
handoff.

## JSON parse gate

- Run `jq .` on every edited JSON file under `internal/connectors/defs/<name>/` and
  `.planning/`.
- A single malformed file fails the slice. The engine loads bundles defensively, but a parse
  error is a release blocker.

## connectorgen validate gate

- `go run ./cmd/connectorgen validate internal/connectors/defs --json`
- Must return `findings: []` and `warnings: []` for the full defs tree (547+ connectors).
- Common findings: write action with no `covered_by` entry in `api_surface.json`, schema mismatch,
  missing required file (`metadata.json`/`spec.json`/`streams.json`), unknown execution model.

## Stacked intra-connector slice gate (intermediate sub-PRs only)

When one connector rollout is decomposed into stacked slices (sub-PRs into a parent branch, e.g.
foundation → schemas → reads → writes → deletes → CLI → fixtures/docs), a coverage-complete
`api_surface.json` necessarily references streams/writes owned by not-yet-integrated slices. For
those intermediate sub-PRs only, the connectorgen validate gate is staged:

- The bundle must stay **loader-valid at every commit**: `metadata.json`, `spec.json`, `docs.md`
  present (a stub refined by its owning slice is fine) and `streams.json` present and schema-valid
  (`{"base": …, "streams": []}` is valid) unless `capabilities.dynamic_schema` is true. The defs
  tree is embedded, so one invalid bundle breaks every package that loads the registry: `go build`,
  `go vet ./...`, and repo-wide `go test ./...` must be green on every slice.
- `api_surface.json` grows incrementally: each endpoint row lands in the SAME slice that declares
  its covering stream/write (conformance `surface_complete` is bidirectional — rows over
  undeclared streams fail it, as do declared streams without rows — so dangling references are
  never a valid intermediate state). The foundation slice ships an empty-endpoints skeleton whose
  `scope` note names the research doc as the source of truth; `connectorgen validate` must return
  `findings: []` and `warnings: []` on EVERY slice.
- Drift protection: every slice that adds endpoint rows must reconcile the manifest against the
  research doc in its PR body — cumulative row count after the slice (e.g. reads → 56, +writes →
  140, +deletes → 168) and zero rows for operations the research doc does not contain. The
  parent-finalize gate fails if the final manifest row count differs from the research doc's
  operation count.
- Every other gate in this document applies unchanged to every slice.
- The final slice / parent-finalize gate is absolute: `findings: []` and `warnings: []` for the
  full defs tree, plus `make verify`, smoke, and certify. The parent PR into `main` remains
  human-gated and must never merge with tolerated findings outstanding.

## Secret-scan gate

- No secret values, API tokens, OAuth tokens, PEM private keys, passwords, or bearer strings in
  any artifact: docs, examples, previews, errors, test fixtures, or generated website data.
- Secret **fields** (`api_key`, `private_key`, …) are allowed as schema field names; their
  **values** are not. Use env-var placeholders (`PM_<NAME>_TOKEN`) in examples and fixtures.
- Sensitive/admin reverse-ETL operations are modeled as `sensitive_reverse_etl` /
  `admin_reverse_etl` / `destructive_admin` and remain blocked by default.

## Source-link gate

- Every `api_surface.json` endpoint row has a non-empty `source_url` pointing at official
  provider documentation (not a blog, not a third-party mirror).
- Provider CLI inventory rows cite the provider's CLI docs when a CLI exists.

## Operation-classification gate

- Every `api_surface.json` row has an `execution_model`. API-backed commands may not remain
  `partial` / `planned` / `unsupported_api` unless the row documents the gap and the reason.
- `reverse_etl` commands in `cli_surface.json` either map to a declared write action with
  `record.*` flag mappings (`availability=implemented`) or are `availability=partial` with a
  note explaining the blocker.

## Build and test gates

- `gofmt -l cmd internal` clean.
- `go vet ./...` clean.
- `go build ./cmd/pm` succeeds.
- Focused package tests pass: `go test ./internal/connectors/<name>/... ./internal/connectors/engine ./internal/connectors/commandrunner ./cmd/connectorgen -count=1`.
- `make verify` when feasible (note: the certify package is slow (~300s) and the `go test`
  timeout is bumped to 20m for headroom; a timeout is not verification completion).

## Website idempotency gate

- When connector docs or `*_surface.json` files change: `cd website && pnpm run gen:website-data`
  must be idempotent — regenerate twice with no diff — and the regenerated files must be
  committed so the "Verify generated website data" CI step passes.

## Review gate

- Claude auto-review where the base branch is `main`; for stacked sub-PRs into the parent
  branch where Claude is disabled, the parent PR must receive review coverage for the
  integrated range before the slice is considered integrated (see
  `.agents/agentic-delivery/workflows/automated-review-routing-loop.md`).
