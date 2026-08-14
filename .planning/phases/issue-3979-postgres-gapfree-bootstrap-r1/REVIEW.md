# Code review — #3979 PostgreSQL gap-free bootstrap

## Scope and route

- Reviewed range: `a59a456dd..HEAD` plus the uncommitted bootstrap implementation before its implementation commit.
- Target connector: native PostgreSQL only. No shared mapping, workset, target, CLI, documentation, dependency, or excluded-issue file changed.
- Review route: inline/manual code review. The Firstmate delivery brief explicitly makes CI, not an automated reviewer, the delivery gate. Opening the non-draft PR may still trigger repository automation; it is not being counted as a prerequisite here.

## Findings and disposition

1. **Accepted during review — snapshot warehouse candidate needed to be stageable.** The first coordinator page shape exposed only the opaque barrier and could not let a connection-owned WAL/Parquet stage bind the full source checkpoint. `BootstrapSnapshotPage` now carries a cloned `CandidateCheckpoint`; the live receiver asserts it validates and has exactly the same source/barrier/schema binding as the page.
2. **Accepted during review — imported snapshot cleanup failure must remain observable.** `readBootstrapSnapshot` now joins an explicit rollback error with the operation error, matching the existing typed-snapshot resource pattern.
3. **Accepted during review — generated snapshot names are rendered as a SQL literal.** The private exported-snapshot name validator is ASCII-only and permits only PostgreSQL's observed generated token alphabet before rendering; no operator value reaches that literal.
4. **Accepted from CI — staticcheck QF1001.** The generated-snapshot rune predicate is now named before negation, preserving the strict ASCII allowlist while conforming to the repository lint gate. `golangci-lint run ./internal/connectors/native/postgres/...` passes.
5. **No remaining blocking finding.** The coordinator owns no target handle. Its snapshot receiver is documented as the connection-owned WAL/Parquet durability boundary and carries a full candidate checkpoint; post-barrier records retain #3977's receipt → checkpoint → LSN acknowledgement ordering.

## Verification after review fixes

- `go test -timeout 20m ./internal/connectors/native/postgres/...` — pass.
- Direct Docker/Colima `databaseintegration` command in `traces/green-live-bootstrap.txt` — pass.
- `go vet ./internal/connectors/native/postgres/...`, `go build ./cmd/pm`, `go run ./cmd/agentcontractgen check`, `go run ./cmd/connectorgen validate`, and `go run ./cmd/connectorgen surface-sync --check` — pass.
