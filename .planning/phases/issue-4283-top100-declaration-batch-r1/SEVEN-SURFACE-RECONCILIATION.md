# Seven-surface reconciliation — PR #4294

This is the authoritative ten-connector reconciliation for the relaunch on
2026-08-20. The machine-readable counts, typed-action set hashes, fixtures and
exact source mappings are in `SEVEN-SURFACE-RECONCILIATION.json`.

| Connector | Documented operations | CLI implemented | Typed actions | Static destination proof |
|---|---:|---:|---:|---|
| Docker Hub | 54 | 45 | 20 | None: no fixture-proven target identity mapping |
| Notion | 49 | 45 | 24 | `views` → `update_view` |
| Stripe | 589 | 8 | 3 | `customers` → `update_customer` |
| Bitbucket | 331 | 3 | 54 | None: fixture lacks required target workspace identity |
| GitLab | 1,755 | 4 | 0 | Not applicable |
| CircleCI | 111 | 2 | 7 | `schedules` → `update_schedule` |
| Sentry | 223 | 0 | 0 | Not applicable |
| Vercel | 400 | 2 | 18 | `projects` → `update_project` |
| Asana | 249 | 82 | 73 | None: closed mapper cannot construct nested action data |
| Jira | 617 | 584 | 292 | None: representative typed input violates mapper identifier grammar |

All remaining typed actions are explicitly covered by the pinned action-set
selectors in the JSON ledger. They remain CLI-reachable where a command is
already declared. They are not excluded for risk, privilege, destructiveness or
the lack of live credentials.

The current generic destination can model one action mapping per destination
declaration. It resolves that mapping from source executor and stream, not the
selected action. `action-scoped-source-binding` is therefore the precise
remaining foundation dependency for multi-action destination coverage. The
application/CLI dispatch integration remains upstream in #4304; no row in this
ledger claims provider-live reverse-ETL deployment.
