# UAT — MySQL container harness R1

This issue-shaped firstmate task uses the documented inline GSD fallback: the project-local Pi
adapter resolved the `verify-work` command, but no compatible isolated GSD verifier is available.
The following acceptance evidence was therefore evaluated directly.

| Deliverable | Automated evidence | Result |
| --- | --- | --- |
| Explicit-context reusable harness | `go test ./internal/connectors/native/dbtest`; harness tests cover pinned tags, dynamic loopback ports, image ownership, cleanup order, reset order, and disk reporting. | pass |
| MySQL source connector | `go test ./internal/connectors/native/mysql`; dynamic catalog/read/binlog component contracts pass. | pass |
| Real engine proof | `DOCKER_CONTEXT=colima POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_DATABASE_RESET_COLIMA=1 go test -tags=databaseintegration -count=1 -v ./internal/connectors/native/mysql` passed in 53.48 seconds against MySQL 8.4.11. | pass |
| No leak after the live proof | Explicit `docker --context colima` listings for containers, volumes, and images all returned no records after deferred cleanup and Colima reset. | pass |
| Safe absence behavior | The tagged test without the opt-in visibly skips with its Docker/Colima reason; with the opt-in, startup/reachability is a test failure rather than a green pass. | pass |
| Connector catalog and docs parity | Connector docs/catalog and website generated artifacts were regenerated; focused CLI catalog/golden/manual tests, `pm help connectors`, `pm connectors`, and `pm connectors inspect mysql --json` passed. | pass |

Verdict: pass. No human-judgment-only acceptance item remains for this implementation slice.
