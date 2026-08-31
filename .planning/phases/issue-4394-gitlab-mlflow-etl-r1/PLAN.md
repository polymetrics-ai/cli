# GitLab #4394 MLflow metrics-history ETL — plan

## Task Delivery Header

- Issue: Refs #4394 / #4413 — one GitLab source-bound MLflow full-refresh ETL slice.
- Base branch: `fm/cli-top100-declaration-batch-r1` at immutable commit `0a708dea5e0024a173b19959d2c43f2bf5a6e0f2`.
- Delivery: candidate branch `codex/4394-gitlab-mlflow-etl-r1` only; no parent or `main` integration without independent review.
- Scope: GitLab connector-definition artifacts, focused GitLab tests, and this issue-local evidence only.
- Verification: retained-lock/matrix reconciliation; exact source-bound preflight; two-page TLS-dial-redirect fixture into local DuckDB; 502 no-materialization; no-credential/preflight no-I/O; JSON, vet, agent-contract, and diff checks.

## Frozen source fact

The sole candidate is
`gitlab.rest.getApiV4ProjectsIdMlMlflowApi20MlflowMetricsGetHistory`:

- `GET /api/v4/projects/{id}/ml/mlflow/api/2.0/mlflow/metrics/get-history`
- required source inputs: path `id`, query `run_id`, and query `metric_key`
- provider continuation: query `page_token` -> JSON `next_page_token`
- records: JSON `metrics[]`
- retained citation: `sources/gitlab-operation-source-lock.json`, SHA-256
  `b33e80759af3a529d71a0bb58c8c76d65c9f0b20774a042196c6d3c0c57310bd`.

`project_id` is an explicit connector configuration key mapped to the locked
path parameter `{id}`. It is not inferred from a bare parameter, method, or
operation name. The two required MLflow query inputs likewise remain explicit
connector configuration keys.

## Atlas and substrate preflight

| Need | Existing capability | Decision |
| --- | --- | --- |
| Source identity, method, route, record and pagination admission | `source.projection-admission.v1` | Reuse existing source-bound `stream_etl` preflight. |
| Provider collection -> DuckDB full refresh | `warehouse.stage-etl.v1` | Reuse only if the source-bound stream and sync contract load without runtime changes. |
| Full-overwrite/full-append source transport | `transport.sync-contract.v1` | Reuse by adding only the declared MLflow stream to GitLab's existing source allowlist. |
| Test-only source-origin fixture | existing source-bound origin guard | Preserve locked GitLab origin and use a local TLS dial redirect only in test code; no `base_url` bypass. |

No Foundation Atlas entry is planned unless the red test exposes an actual
shared runtime deficiency. A connector-specific source/matrix validator may
prove exact schema and continuation provenance, but it must not conceal a
runtime requirement.

## Red -> green -> refactor

1. **Red (this checkpoint):** add an issue-local TDD record and a focused
   GitLab test that requires the exact retained source ID, the explicit
   `project_id -> {id}` configuration binding, required query bindings,
   `metrics` record path, and `page_token -> next_page_token` cursor. The test
   also mutates source ID, method, path, schema, and pagination to prove the
   existing source-bound preflight rejects each structural loss.
2. **Green (only after the red evidence is reported):** add the field-complete
   stream, schema, source-bound `stream_etl` operation, connector spec keys,
   source transport allowlist/contract evidence, and exact matrix mapping.
   Promote only this ETL cell to `implemented`.
3. **Witness:** use a two-page local TLS-dial-redirect fixture while retaining
   `https://gitlab.com/api/v4` as the selected source origin. Assert the second
   request carries the returned `page_token`, DuckDB receives both metrics,
   a 502 creates no materialization, and preflight/credential checks perform
   no provider I/O.
4. **Refactor/verify:** retain direct-read, write, delete, binary, GLQL, and
   webhook rows unchanged; run the scoped normal tests first and request any
   heavy race slot separately.

## Explicit exclusions

- No source-lock, descriptor, crosswalk, shared engine, generic transport,
  certification, credential, Atlas, or CLI-runtime changes.
- No GLQL POST-body cursor ETL, inbound webhook receiver/sync, binary work,
  bracket-query aliases, or reverse-ETL change.
- No real provider credential or network I/O.

## GSD execution record

`scripts/gsd doctor` and the issue-phase sources/prompts were resolved before
the plan. This narrow issue worktree is not a registered roadmap phase and the
task forbids additional runtime agents, so the rendered discuss/plan/execute/
verify/review steps are recorded manually here without weakening TDD or the
independent-review gate.
