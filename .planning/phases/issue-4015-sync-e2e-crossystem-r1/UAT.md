# UAT — issue #4015 cross-system pipelines

Mode: inline/manual `verify-work` fallback. The official prompt was resolved with `scripts/gsd prompt verify-work issue-4015-sync-e2e-crossystem-r1`; the repository's canonical single-worker contract forbids spawning a delivery role.

| Deliverable | Independent acceptance evidence | Result |
| --- | --- | --- |
| D1 — PostgreSQL → GitHub | Fresh binary delivered one PostgreSQL row through owned Parquet and approved `update_label`; separate GitHub GET/list matched exact content/count; replay skipped `0/0` and did not duplicate. | Pass |
| D2 — GitHub → PostgreSQL | First run matched an independent 10-label source count and pgx read-back. Replay completed `0/0`, while pgx found zero rows and the named sample absent, contrary to `full_overwrite`. | Fail — product defect recorded |
| D3 — GitHub → GitHub | Exact-`since` source, independent warehouse query, approved `update_issue_comment`, and separate GitHub GET/list all matched comment `5328121289` and exact count 1. | Pass |
| D4 — installed schedule and persisted flow | Exact task-local crontab payload ran the stored two-step flow; schedule inspection recorded terminal state and prepared identity; separate GitHub read-back remained exact. | Pass |
| D5 — cleanup | Every created label/release/tag/comment from all attempts returned HTTP 404; container and volume name queries found no task resources. | Pass |

No human judgment is required for the route facts: each is a directly observed live assertion. Overall UAT is `verified_with_finding`, not fully passed, because D2 violates its declared sync mode. Product remediation is intentionally deferred to a separate lane.
