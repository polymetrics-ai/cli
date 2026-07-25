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

## Delivery-profile and review gate

- `pm_worker` must have committed and pushed coherent green slices only to its assigned branch;
  `coordinator_fanout` must remain uncommitted and return an exact path/verification handoff to the
  coordinator. Neither profile edits shared generated/bootstrap files without explicit scope.
- After exact-head verification for `pm_worker`, run v4
  `scripts/pm-review-system.py compile --scope ... --base ... --head ...`, require a ready manifest,
  render every packet with the canonical `render` command, and pass each rendered stdout unchanged
  to a fresh-context local Codex reviewer. Require one authenticated `synthesize` result that is
  clean with every finding dispositioned. Then require independent Shepherd `PROCEED` for the same
  exact base/head/tree before the slice is ready (see
  `.agents/agentic-delivery/workflows/local-codex-review-loop.md`).
