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

The JSON ledger now contains one named eligibility disposition for every one
of the 491 typed actions: four are `eligible_bound_fixture_mapping` and 487
are eligible but pending the exact source-identity, nested-object, or
action-scoped mapping dependency named alongside that action. The action-set
hash remains an anti-drift selector for each connector. These actions are not
excluded for risk, privilege, destructiveness, or the lack of live credentials.

The current generic destination can model one action mapping per destination
declaration. It resolves that mapping from source executor and stream, not the
selected action. `action-scoped-source-binding` is therefore the precise
remaining foundation dependency for multi-action destination coverage. The
application/CLI dispatch integration remains upstream in #4304; no row in this
ledger claims provider-live reverse-ETL deployment.

## Documented-operation command reachability boundary

The source crosswalk-to-current-surface join records 3,366 of 4,378 documented
operations without a declared command binding: Docker Hub 11, Notion 5, Stripe
581, Bitbucket 134, GitLab 1,751, CircleCI 95, Sentry 220, Vercel 378, Asana
164, and Jira 27. This count excludes the documented surface-only variants
that are not in each pinned provider source denominator.

Existing operation-level rejections identify every affected endpoint. The
current connector-local schema cannot give one of those rejected endpoints a
direct CLI command: `checkCLISurfaceEndpointCoverage` rejects its unbound
`api_surface` reference and `resolvePreflightCommand` rejects an
operation-backed partial command. The needed closed foundation is
`declaration-bound-disabled-command-surface`, which would produce a precise
`BlockedCommandError` for exactly one declared endpoint without permitting
generic provider I/O. This connector lane does not implement that shared change
without a keyed foundation decision.
