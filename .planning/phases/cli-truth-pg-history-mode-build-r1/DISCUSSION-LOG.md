# Discussion log — PostgreSQL history-mode truthfulness repair

`scripts/gsd prompt discuss-phase cli-truth-pg-history-mode-build-r1` was
resolved inline because this direct-PR worker has no compatible Pi runtime for
the interactive GSD role workflow, and repository policy forbids spawning
separate planner/reviewer roles. The task brief and shared captain decision
provide the required answers.

| Area | Decision | Evidence |
| --- | --- | --- |
| Product scope | Build the existing claimed PostgreSQL capability; do not shrink the claim. | `data/truthfulness-build-shared-context.md`; F2 of `data/cli-mvp-truthfulness-sweep-r1/report.md`. |
| Connector ownership | PostgreSQL only. | `internal/connectors/defs/postgres/sync_transport.json`. |
| History meaning | Use the existing `dedupe_history` writer semantics, not a newly invented representation. | `managed_target_write.go`, `database_write_session.go`, and the existing PostgreSQL live history test. |
| Atomicity | Preserve one run-scoped publication and durable receipt before checkpoint. | Existing managed-target executor and PR #4184 constraint. |
| Proof | Use the built `pm` binary against a live PostgreSQL source and target; query the target separately for update and replay state. | Task acceptance and shared proof bar. |
| Docs parity | Existing prose is correct; prove generators remain stable rather than create spurious wording churn. | Advertised surfaces named in F2. |

No `needs-decision:` item is open.
