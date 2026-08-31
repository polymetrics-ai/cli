# GitLab #4394 MLflow metrics-history ETL — TDD ledger

| Slice | Red evidence | Green evidence | Negative / boundary proof |
| --- | --- | --- | --- |
| Source-bound full-refresh declaration | Recorded against the frozen base: `GOCACHE=/private/tmp/gocache-gitlab-mlflow-etl-r1 go test -count=1 ./internal/connectors/defs/gitlab -run '^TestGitLabMLflowMetricsHistoryETLDeclarationIsSourceBound$'` failed because the exact ETL cell was `mapped_unproven`, as intended. | Green: the exact source ID is now the only promoted GitLab matrix ETL cell, bound to `mlflow_metrics_history`, its closed schema, explicit config selectors, and `full_refresh`. The focused matrix/declaration test passed. | Exact retained source ID, method, path, record path, schema, explicit config bindings, and cursor pair must all agree. |
| Preflight structural guard | Recorded: the focused table mutates source ID, method, path, schema, and pagination. | Green: the focused declaration test passed, including the structural mutation table. | Source-bound preflight rejects each mismatch before credential resolution or provider I/O. |
| Runtime witness | Red plan recorded before artifact promotion. | Green: the focused provider-to-DuckDB witness, 502/no-materialization, and missing-credential/no-I/O tests passed. | Test-only TLS dial redirect retains source-locked origin; 502 leaves no materialized table rows. |

Only the exact MLflow metrics-history ETL cell is promoted. All other GitLab
ETL candidates retain their prior disposition.
