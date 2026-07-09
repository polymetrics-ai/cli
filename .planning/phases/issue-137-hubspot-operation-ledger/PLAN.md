# Plan: Issue #137 HubSpot Operation Ledger

Date: 2026-07-09
Parent issue: #132
Sub-issue: https://github.com/polymetrics-ai/cli/issues/137
Parent branch: `feat/132-hubspot-cli-parity`
Active branch: `feat/137-hubspot-operation-ledger`
Stack base: `feat/134-hubspot-cli-surface-metadata` until #134 is integrated into the parent branch

## GSD command path

- `scripts/gsd doctor` — passed.
- `scripts/gsd verify-pi` — passed.
- `scripts/gsd list --json` — passed, 69 commands.
- `scripts/gsd prompt programming-loop init --phase issue-137-hubspot-operation-ledger --dry-run` — blocked: pinned registry does not contain `programming-loop`.
- Active fallback: manual universal programming loop using `scripts/gsd prompt plan-phase issue-137-hubspot-operation-ledger --skip-research`, with TDD ledger and verification checklist maintained before production edits.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-context`
- `golang-concurrency`
- `golang-documentation`
- `golang-graphql`
- `golang-lint`

## Objective

Account for every unique official HubSpot OpenAPI method/path operation in `internal/connectors/defs/hubspot/api_surface.json` using the official public OpenAPI collection. The ledger must preserve full-surface safety: sensitive, admin, destructive, binary, and write operations are classified as typed gated app-operation candidates, not hidden behind generic raw tools.

## Scope for this slice

In scope:

- Fetch/read the official HubSpot public OpenAPI collection in a temporary directory only.
- Parse OpenAPI JSON files, excluding Postman collection files, and deduplicate by `(HTTP method, path)`.
- Generate a deterministic `api_surface.json` ledger with `operation_ledger_version: 1` and the current official baseline: 401 OpenAPI files, 4,396 raw versioned operations, 3,060 unique operations, method counts GET 1,038 / POST 1,314 / PUT 169 / PATCH 232 / DELETE 307.
- Expand the operation-ledger schema/validator vocabulary if necessary so blocked app-operation rows can classify `stream_etl`, `query_etl`, `reverse_etl`, and binary/file operation candidates without pretending they are implemented.
- Add fail-first tests for HubSpot ledger counts, method distribution, row classifiers, and no legacy `excluded` rows in operation-ledger mode.
- Update HubSpot docs/planning status with the generated inventory totals and classification counts.

Out of scope for this slice:

- Live HubSpot credentials or live API checks.
- Executing HubSpot reads/writes.
- Creating thousands of unsafe generic write actions without fixed request schemas.
- Runtime direct-read dispatcher work (#138).
- Stream schemas/fixtures and stream runner coverage (#136).
- POST body/search/query/binary runtime executor work (#139).
- Sensitive/admin write confirmation/redaction policy and fixture-backed write execution (#140).

## Safety interpretation

- A ledger row is an inventory classification, not runtime dispatch.
- Non-excluded mutations must not become executable until they are named `writes.json` actions with fixed schemas and plan → preview → approval → execute. This slice records them as `operation.model: reverse_etl`, `admin_reverse_etl`, `sensitive_reverse_etl`, or `destructive_action` with `blocked_by_default: true` and issue-linked reasons.
- Binary/file rows must not become executable until #139 supplies bounded max bytes, destination policy, no-overwrite/path traversal protections, and redaction.
- No `raw_api`, `direct_write`, generic HTTP write, generic shell write, or generic SQL write is introduced.

## TDD plan

1. Red: add a test proving HubSpot `api_surface.json` must contain all 3,060 official unique method/path operations and the official method breakdown.
2. Red: add a test proving operation-ledger mode rejects legacy `excluded` rows and requires a classifier on every row.
3. Red/green: extend schema/validator operation model vocabulary for non-executable app-operation candidate rows when needed.
4. Green: generate the deterministic HubSpot operation ledger and docs status.
5. Refactor: keep generation deterministic and document the official source, baseline counts, classification heuristic, and limitations.

## Targeted verification

```bash
gofmt -w cmd internal
go test ./cmd/connectorgen -run 'HubSpot|APISurfaceOperationLedger' -count=1
go run ./cmd/connectorgen validate internal/connectors/defs
python3 -m json.tool internal/connectors/defs/hubspot/api_surface.json >/dev/null
```

## Broad verification before handoff

```bash
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## CLI help/docs/website parity

This slice changes connector metadata/docs but does not add executable `pm hubspot ...` runtime commands.

- Runtime `pm hubspot` help: not applicable until #135/#136.
- `pm help connectors` and `pm connectors inspect hubspot --json`: applicable after metadata/docs update.
- `docs/connectors/hubspot/**`, connector catalog, and generated manuals: update if inspection output changes.
- Website docs: not applicable unless #135 help/docs renderer work is pulled into scope.

## Human gates

- New dependencies.
- Live HubSpot credentials or live HubSpot API checks.
- Auth scope changes.
- Destructive/admin external actions.
- Reverse ETL execution.
- Binary transfer runtime enablement without explicit destination policy.
- Generic raw write tools.
