# GitLab #4394 MLflow metrics-history ETL — verification checklist

- [x] Exact base and GitHub remote verified before edits.
- [x] CodeGraph checked: no `.codegraph/` directory is present in this worktree.
- [x] Connector build order and Foundation Atlas capabilities inspected.
- [x] Red source/matrix/preflight test recorded against the untouched connector artifacts.
- [x] Green connector-only artifact materialization is authorized by the red evidence; the lane promotion remains pending green proof.
- [x] Two-page source-origin-preserving fake-server -> DuckDB witness passes.
- [x] 502 leaves no materialized table rows.
- [x] Missing credential/preflight path makes no provider I/O.
- [x] Focused normal source/matrix/CLI/contract tests, connector validation, JSON parsing, gofmt, and diff checks pass.
- [ ] Candidate commit/push and remote SHA verification complete.
- [ ] Independent review complete before any parent integration.
