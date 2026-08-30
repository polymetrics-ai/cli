# Stripe source-to-seven-lane matrix — #4418

## Task delivery header

| Field | Value |
| --- | --- |
| Issue refs | #4418 — Stripe — source-to-seven-lane matrix |
| Base branch and commit | `fm/cli-top100-declaration-batch-r1` at `dc481bac1a8b78d60ac0b4a2b0dfd1a9068ce8db` |
| Working branch | `fix/4418-stripe-track-a-r1` |
| Integration path | scoped commit pushed for the Batch R1 parent; no PR and no merge opened |
| Source authority | `internal/connectors/defs/stripe/sources/stripe-operation-source-lock.json` |
| Source identity | Stripe OpenAPI `2026-07-29.dahlia`, SHA-256 `3653ad45bbec54fcbe461c541c908355b715018bdf455a0e11b27bedb2cbdee5` |

## Fixed scope

This delivery is a mapping/certification-admission record only. It adds a seven-lane source matrix and a local Go contract test under the existing Stripe definition directory. It does not alter connector execution, generator behavior, transport, credentials, live I/O, or shared Foundation code.

Every retained source row remains present, including all `DELETE` and other mutation rows. A source fact may support a `mapped_unproven` candidate, but it never establishes an executable capability.

## Source facts discovered before implementation

| Fact | Locked evidence |
| --- | --- |
| Retained REST rows | 589 |
| GET rows | 263 |
| Mutation-verb rows | 326 (`POST` 294 + `DELETE` 32) |
| Paging-shaped GET candidates | 128: 121 `starting_after` cursor rows and 7 `page`/`next_page` rows |
| List-shaped but no documented continuation | `stripe.rest.GetAccountsAccountCapabilities`, `stripe.rest.GetReportingReportTypes` |
| Binary media evidence | `stripe.rest.GetQuotesQuotePdf` response `application/pdf`; `stripe.rest.PostFiles` request `multipart/form-data` |

The paging predicate is deliberately exact: a GET row has JSON `data` as an array and `has_more`, plus either `starting_after`, or both `page` and response `next_page`. The two list-shaped rows above remain visible but receive a source-evidenced non-applicable ETL/sync disposition because continuation is not documented for their operation.

## Foundation Atlas decision

Consulted Atlas entries: `source.projection-admission.v1`, `runtime.direct-execution.v1`, `transport.sync-contract.v1`, and `warehouse.reverse-etl.v1`.

No runtime foundation is named or selected. The matrix has zero `missing_foundation` cells because this task is not attempting to admit execution. Candidate cells are `mapped_unproven`; their evidence is source-local only. No captain decision is required.

## Source-local integration boundary

The matrix is connector-local at `internal/connectors/defs/stripe/sources/stripe-source-lane-matrix.json`, matching the current Track A convention. The completed #4293 shared checker owns a root-level multi-connector manifest, so composing this Stripe evidence into that shared manifest is a Batch R1 parent responsibility. This task intentionally does not create a competing shared manifest.

## GSD and skill record

`scripts/gsd doctor`, required prompt help, and source checks were reviewed. The Pi-local GSD adapter does not provide the specialized planning/researcher roles for this scoped task, and repository/task policy does not authorize delegation, so the required research/plan/test-ledger/verification artifacts are recorded inline.

Loaded skills: `connector-lane-build-order`, `go-engineering` (including fundamentals and agentic ETL guidance), `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-safety`, `golang-security`, and `golang-testing`.
