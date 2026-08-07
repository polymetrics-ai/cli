# UAT — MySQL container harness R1

This issue-shaped firstmate task uses the documented inline GSD fallback: the project-local Pi
adapter resolved the `verify-work` command, but no compatible isolated GSD verifier is available.
The following acceptance evidence was therefore evaluated directly.

| Deliverable | Automated evidence | Result |
| --- | --- | --- |
| Explicit-context reusable harness | `go test ./internal/connectors/native/dbtest`; harness tests cover pinned tags, dynamic loopback ports, image ownership, cleanup order, reset order, and disk reporting. | pass |
| MySQL source connector | `go test ./internal/connectors/native/mysql` plus the bundle-registry native-override assertion cover dynamic catalog/read/binlog contracts and public native registration. | pass |
| Real engine proof | `POLYMETRICS_DATABASE_INTEGRATION=1 POLYMETRICS_PODMAN_CONNECTION=<machine> POLYMETRICS_DATABASE_RECLAIM_DISK=1 go test -tags=databaseintegration -count=1 -v ./internal/connectors/native/mysql` passed in 33.8 s against MySQL 8.4.11, including four `sslmode` subtests checked against the server's negotiated `Ssl_cipher`. | pass |
| No leak after the live proof | Explicit `podman --connection <machine>` listings for containers, volumes, and images all returned no records. The machine's sparse disk file moved by +0.2 MiB end to end. | pass |
| Safe absence behavior | Without the opt-in or the mandatory Podman connection the tagged test visibly skips with its reason; with the opt-in, startup/reachability is a test failure rather than a green pass. | pass |
| Connector catalog and docs parity | Connector docs/catalog and website generated artifacts were regenerated; focused CLI catalog/golden/manual tests, `pm help connectors`, `pm connectors`, and `pm connectors inspect mysql --json` passed. | pass |

Verdict: pass. No human-judgment-only acceptance item remains for this implementation slice.
