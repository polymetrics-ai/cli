# TDD Ledger — Issue #4031

| Stage | Check | Expected result | Evidence |
| --- | --- | --- | --- |
| Red | `TestCertificationArchitectureUsesExecutablePostgresResumeBindings` | Fails while the architecture uses `$cursor` / `$primary_key`. | `go test -timeout 20m -run '^TestCertificationArchitectureUsesExecutablePostgresResumeBindings$' ./internal/connectors/native/postgres` exited 1 because the required `$1`/`$2` predicate was absent. |
| Green | Same focused test after the documentation change. | Passes with `$1` reused for cursor and `$2` for primary key. | `go test -timeout 20m -run '^TestCertificationArchitectureUsesExecutablePostgresResumeBindings$' ./internal/connectors/native/postgres` passed. |
| Regression | `go test -timeout 20m ./internal/connectors/native/postgres` | Passes without a live database. | Passed in 0.91s. |
| Refactor | `gofmt`, `git diff --check`, scoped review. | No formatting or scope drift. | `gofmt -w internal/connectors/native/postgres/postgres_test.go`, `git diff --check`, and manual review passed. |

## Red rationale

The behavior under test is the architecture document's executable SQL contract,
not runtime behavior. The assertion deliberately reads the exact document and
rejects the two symbolic values so a future doc edit cannot silently return to
an unbound predicate.
