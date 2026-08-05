# Plan: Google Analytics Data API documented parity (#3030–#3037)

Branch: `fm/cli-google-analytics-data-api-parity-wave03-r1`
Parent: #3030 · subissues: #3031–#3037

## GSD and required skills

- `scripts/gsd doctor` and `scripts/gsd list` passed; `scripts/gsd prompt discuss-phase 3030 --auto` and `scripts/gsd prompt plan-phase 3030 --tdd --auto` rendered the official workflow prompts.
- The active repository rule forbids the absent `programming-loop` command. This phase uses the documented manual GSD fallback: this plan, TDD ledger, and verification checklist are the execution record.
- Loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-documentation`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-lint`, `vercel-react-best-practices`, and `vercel-composition-patterns`. The catalog has no `frontend-design` or `web-design-guidelines` skill; the generated data-only website change requires no React component work.

## Provider-derived inventory

Google's [Data API REST reference](https://developers.google.com/analytics/devguides/reporting/data/v1/rest), retrieved 2026-08-05, explicitly publishes both Data API discovery documents. Both discovery artifacts have revision `20260803`:

| Artifact | Provider operations | Classification |
| --- | ---: | --- |
| `https://analyticsdata.googleapis.com/$discovery/rest?version=v1beta` | 11 | 10 reads, 1 write |
| `https://analyticsdata.googleapis.com/$discovery/rest?version=v1alpha` | 15 | 12 reads, 3 writes |
| Semantic union | **24** | **20 reads, 4 writes** |

The v1alpha `getMetadata` and `runReport` operations are semantically equivalent to their v1beta counterparts and are counted once, following the audit cohort policy. This replaces the invalid 10-row baseline; 24 is 2.4× the baseline and is approved scope, not a deferral.

## Slice boundaries and commit checkpoints

1. **Planning checkpoint** — update this plan/TDD/verification record for the approved 24-operation scope and push it.
2. **Provider-ledger checkpoint** — commit only `api_surface.json`: one provenance-bearing, exactly-once 24-operation ledger. This is intentionally separate from implementation review.
3. **Red/green alpha direct-read slice** — add failing connector-owned tests, then implement fixed, bounded v1alpha GET direct reads: property quota snapshot; audience-list get/list; recurring-audience-list get/list; report-task get/list. Add sanitized fixtures and CLI operation metadata. Preserve fixed paths, numeric property validation, output redaction, and fixture-only tests.
4. **Typed POST/query and write disposition** — complete the operations/CLI ledger for all 24 semantic operations. Implement a POST read only when the current direct-read contract accepts its closed request schema and has redacted fixture coverage. The remaining report/query operations may be blocked only on the named shared provider-query/redaction foundation #2985; the four asynchronous create operations may be blocked only pending a closed named reverse-ETL action with plan → preview → explicit approval → execute, redaction, and idempotency evidence.
5. **Parity surfaces** — regenerate/update connector docs, catalogs, website generated data, help/golden data, and the parent plus seven child issue addenda with final truthful counts. No generic HTTP, SQL, shell, or raw-body surface is added.
6. **Verification and delivery** — execute targeted gates, required local gates, `git diff --check`, code review, no-mistakes PR/CI path, then report the PR when CI first passes. Never merge.

## Execution record

- The provider-ledger-only checkpoint is commit `7c550c075`; it is deliberately separate from all implementation commits.
- The fixed v1alpha metadata-read slice is commit `6daf9e150`. The current implementation checkpoint extends it with the remaining six typed v1alpha operation declarations, provider-ledger count regression coverage, generated manuals/catalog/website data, and GA command golden transcripts.
- The manual GSD fallback was used after rendering the `execute-phase`, `verify-work`, and `code-review` prompts. Pi has no compatible isolated GSD worker runtime and repository policy forbids role spawning; execution, verification, and review evidence are therefore recorded in these phase artifacts and the PR body.

## TDD and safety rules

- Every new executable direct-read operation starts with a failing native connector test, then fixture and live-`httptest` green evidence. No Google credential or provider call is made.
- All path parameters retain existing numeric property-ID checks; resource IDs remain typed path parameters; response bodies use `json_redacted` and the existing 1 MiB limit.
- No write action is advertised without a closed record schema, redaction, plan/preview/approval/execute semantics, and fixture request-shape coverage.
- CLI parity means definition-owned fixed commands only. `pm connectors`, `pm help connectors`, the connector command help, manuals, website data, and golden data are checked where applicable.

## Required verification

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/google-analytics-data-api
go test ./internal/connectors/conformance -run 'TestConformance/google-analytics-data-api' -count=1
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1
go build ./cmd/pm
make connector-boundary
make verify
git diff --check
```

Additionally: focused `go test ./internal/connectors/native/google-analytics-data-api -count=1`, generated surface checks, `go vet` on changed packages, and the no-mistakes/automated-review evidence required for the PR.
