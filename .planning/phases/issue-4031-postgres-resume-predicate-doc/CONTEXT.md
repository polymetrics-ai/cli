# Issue #4031 — PostgreSQL Resume Predicate Documentation Context

## Domain

Correct the PostgreSQL certification architecture's resume SQL so the example
uses PostgreSQL/pgx positional parameter syntax without changing runtime
behavior.

## Locked decisions

- Keep the correction to `docs/architecture/github-postgres-warehouse-certification.md` and one narrowly scoped regression assertion in the existing PostgreSQL test package.
- Preserve the architecture's composite `(cursor, primary_key)` keyset requirement. This correction changes placeholder syntax only; it neither claims nor implements the legacy reader's missing tie-breaker behavior.
- Use `$1` for both cursor comparisons and `$2` for the primary-key tie-breaker. State the corresponding pgx argument order in the document.
- Keep the child branch `fix/4015-postgres-resume-predicate-doc` stacked on `docs/4015-connector-release-certification` and open a draft PR only to that base.
- Do not run credentialed or runtime-backed checks and do not edit the #3855 branch.

## Source alignment

- `internal/connectors/native/postgres/reader.go` constructs a query with the PostgreSQL positional placeholder `$1`, appends the lower bound to `args`, and invokes `pool.Query(ctx, sql, args...)`.
- #3855 identifies that scalar query as the intentionally legacy, unsound shape because it lacks the composite tie-breaker. Its acceptance contract requires the complete `(cursor, tie_breaker)` predicate; this architecture correction documents that target using valid positional bindings.

## Canonical refs

- `AGENTS.md` — repository delivery, safety, GSD, and verification contract.
- `.agents/agentic-delivery/references/required-skills-routing.md` — skill routing.
- `.agents/agentic-delivery/references/gsd-pi-adapter.md` — installed lifecycle adapter.
- `.agents/agentic-delivery/references/runtime-rlm-website-integration.md` — PostgreSQL architecture documentation safety boundary.
- `.agents/agentic-delivery/contracts/issue-agent-contract.md` — child issue and stacked PR contract.
- `docs/architecture/github-postgres-warehouse-certification.md` — corrected architecture surface.
- `internal/connectors/native/postgres/reader.go` — pgx parameter-binding source proof.
- `internal/connectors/native/postgres/postgres_test.go` — narrow regression-check home.
- GitHub issues #4015, #4031, and #3855 — parent, correction owner, and legacy-source rationale.

## Manual-GSD fallback

`scripts/gsd doctor` is healthy and all five lifecycle commands resolved via
`scripts/gsd sources`. The official workflows cannot initialize issue #4031 as
a phase: `gsd-sdk query init.phase-op 4031` reports `phase_found: false`
because `.planning/ROADMAP.md` deliberately archives its stale numbered roadmap
and directs current work to issue-first delivery. The repository contract
explicitly permits an inline/manual fallback when the lifecycle cannot run.
This directory records the generated `discuss-phase`, `plan-phase --tdd`,
`execute-phase`, `verify-work`, and `code-review` prompts plus their equivalent
single-worker evidence. No subagents are used; the correction is one bounded
issue and the current contract prohibits role spawning.
