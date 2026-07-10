# Claude Review Rubric — polymetrics-ai/cli

Claude reads this on every `@claude` review. Keep it short and high-signal.
This reviewer is **on-demand and complementary** to CodeRabbit (primary auto-review)
and Copilot (fallback) — spend effort where an LLM that has read the *whole codebase*
adds value, not on style nits the linters and CodeRabbit already catch.

## 0. Load context first (don't review the diff blind)
- Read `CLAUDE.md`, `AGENTS.md`, and the relevant `.agents/*` contract files.
- Read `CONTRIBUTING.md` (Conventional Commits, issue-first PRs).
- Read the changed files **and** their package neighbors, interfaces, and callers.
- For connector changes, read `internal/connectors/defs` and how `connectorgen`
  validates them (`make connectorgen-validate`).

## 1. Correctness & bugs (highest priority)
- Logic errors, off-by-one, nil derefs, incorrect edge-case handling.
- Concurrency: goroutine leaks, unguarded shared state / data races, missing
  `context.Context` propagation and cancellation, blocking on unbuffered channels.
- Resource lifecycle: `defer Close()`, DuckDB connections/statements, file handles,
  HTTP bodies. CGO/`-tags duckdb` paths get extra scrutiny.

## 2. Go idioms & error handling
- Errors wrapped with `%w` and context; no silently swallowed errors; no `panic`
  in library paths.
- Accept interfaces, return structs; small focused interfaces.
- No premature abstraction; follows patterns already in the touched package.
- `gofmt`/`go vet`-clean; naming matches surrounding code.

## 3. Architecture & fit (where whole-repo context matters)
- Change lives in the right layer (`cmd/pm` = CLI wiring; `internal/*` = logic).
- Doesn't duplicate an existing helper; reuses connector/ETL/reverse-ETL abstractions.
- Public surface / CLI flags stay consistent with existing UX; no breaking changes
  without a note.
- Connector defs stay declarative and pass `connectorgen validate`.

## 4. Tests
- New/changed logic has tests (table-driven where idiomatic).
- Edge cases and error paths covered, not just the happy path.
- Smoke-test-relevant flows (`make smoke`) still make sense after the change.

## 5. Security & data handling
- No hardcoded secrets/credentials; connector auth handled via existing paths.
- Input validation on external/config-driven data; SQL built safely for DuckDB.
- No sensitive data written to logs.
- License-header / Elastic License 2.0 expectations respected for new files.

## Output format
- **Inline comments** (`mcp__github_inline_comment__create_inline_comment`,
  `confirmed:true`) for anything tied to specific lines — cite `file:line`.
- **One** top-level `gh pr comment` summary: a 2–4 line verdict + the top 3–5
  findings ranked by severity, and an explicit "no blocking issues" if clean.
- Prefer fewer, correct, high-confidence findings over exhaustive nitpicking.
