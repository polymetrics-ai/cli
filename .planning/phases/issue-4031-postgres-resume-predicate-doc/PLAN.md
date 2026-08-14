# Plan — Issue #4031 PostgreSQL Resume Predicate Documentation

## Objective

Make the PostgreSQL resume predicate in the #4015 certification architecture
copyable as PostgreSQL/pgx SQL by replacing symbolic placeholders with bound
positional placeholders, without changing the PostgreSQL implementation.

## Scope fence

Allowed changed paths:

- `docs/architecture/github-postgres-warehouse-certification.md`
- `internal/connectors/native/postgres/postgres_test.go`
- `.planning/phases/issue-4031-postgres-resume-predicate-doc/**`

Excluded: runtime reader implementation, polling algorithms, sync modes,
provider behavior, capability claims, certification outcomes/counts, GitHub
work, generated surfaces, and the #3855 branch.

## Source proof and decision

`internal/connectors/native/postgres/reader.go` uses pgx positional binding:
it builds `WHERE <cursor> > $1`, appends the lower bound to `args`, and calls
`pool.Query(ctx, sql, args...)`. Issue #3855 calls that scalar predicate a
legacy unsound shape and specifies a complete cursor/tie-breaker keyset
contract. Therefore the architecture example remains composite but uses valid
PostgreSQL placeholders:

```sql
WHERE cursor > $1
   OR (cursor = $1 AND primary_key > $2)
ORDER BY cursor, primary_key
```

The document names the binding order: cursor as argument one and primary key as
argument two. This is executable pgx/PostgreSQL syntax; it does not
assert that the legacy scalar reader already implements the architecture target.

## TDD execution

1. **Red** — add a focused documentation assertion in the existing PostgreSQL test package requiring the positional composite predicate and rejecting `$cursor` / `$primary_key`. Run only that test before the architecture edit; it must fail against the symbolic baseline.
2. **Green** — change only the architecture paragraph/code block and run the focused assertion again. It must pass.
3. **Refactor** — run gofmt (test file only), inspect the diff, and keep the wording concise. No behavior refactor is permitted.

## Required skills used

- `github-issue-first-delivery`
- `golang-documentation`
- `golang-how-to`
- `golang-testing`
- `golang-database`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, and `gsd-code-review`

## Verification plan

1. Focused red/green documentation assertion.
2. `go test -timeout 20m ./internal/connectors/native/postgres`.
3. `go vet ./internal/connectors/native/postgres` and applicable lint.
4. Build/check documentation with `make docs-check-no-build` after building the local binary, then `git diff --check` and scoped diff review.
5. `go run ./cmd/agentcontractgen check`.
6. Inline GSD verify/review evidence, then `no-mistakes axi run` with `--skip=push,pr,ci`, never `--yes`.

At most five substantive correction loops are permitted. A newly discovered
defect receives its own issue before any related implementation change.
