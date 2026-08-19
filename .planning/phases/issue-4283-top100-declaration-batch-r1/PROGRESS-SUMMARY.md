# Top daily-use declarations — increment 001

This checkpoint pins public API descriptions and records declaration parity only. It performs no live credentialed certification; every connector retains `live_certification: pending`.

| Connector | Rank | Documented / declared | Disabled | CLI declarations | Write actions (delete) |
| --- | ---: | ---: | ---: | ---: | ---: |
| Docker Hub | 3 | 54 / 54 | 50 | 4 | 0 (0) |
| GitLab | 5 | 1,755 / 1,755 | 1,751 | 4 | 0 (0) |
| Jira | 7 | 617 / 617 | 27 | 590 | 292 (84) |
| Vercel | 8 | 400 / 400 | 378 | 0 | 18 (8) |
| Notion | 9 | 49 / 49 | 5 | 49 | 24 (4) |
| Stripe | 10 | 589 / 589 | 581 | 8 | 3 (1) |
| Bitbucket | 11 | 331 / 331 | 134 | 5 | 54 (53) |
| CircleCI | 12 | 111 / 111 | 95 | 0 | 7 (3) |
| Sentry | 13 | 223 / 223 | 220 | 0 | 0 (0) |
| Asana | 18 | 249 / 249 | 164 | 249 | 73 (4) |
| **Total** | — | **4,378 / 4,378 (100%)** | **3,405** | **909** | **471 (157)** |

All ten were chosen from the ranked daily-use shortlist because they expose a retrievable primary public OpenAPI description and already have a declarative API surface inventory. Docker Hub is included as the rank-3 proof job. Native MySQL and the unconverted Gitea surface are deferred to a later cohort, not rejected: this measured first batch isolates the public-description declaration pipeline.

Each bundle now has a pinned source-lock plus `operations.json`, `writes.json`, and a generated `certification-sweep.json`. The current engine has no safely registered generic declarative API source/destination transport, so no `sync_transport.json` was fabricated; see `FOUNDATION-GAPS.md` and #4093.

### Docker Hub full-parity proof

Docker Hub's 54 pinned method/path rows now have a connector-local exact
crosswalk and declaration-disposition ledger. The bundle materializes 49
source-contract inventory rows (23 `rest_read`, 26 `rest_write`, including 6
typed delete contracts) while retaining its four existing ETL stream bindings.
Those inventory rows are deliberately non-terminal: the API surface keeps 50
routes disabled (46 elevated-scope, three HEAD-executor foundation gaps, and
one source-deprecated login). No new command or reverse-ETL action is claimed,
so `certification.json` is not applicable and live certification remains
pending. This gives Docker Hub 54 / 54 explicit source dispositions (100%).

Transport parity is intentionally **blocked, recoverable, and recorded for all ten**: the ten dedicated `sync_transport` rejection entries carry `reason: foundation-gap`, `recoverable: true`, and the minimum safe recovery. See `TRANSPORT-GAP.md` before selecting increment 2.

The JSON progress ledger, exact per-operation rejection list, and foundation-gap reason index are adjacent to this summary.
