# UAT — MySQL container harness R1

This firstmate task uses the documented inline GSD fallback. The project-local adapter resolved
`verify-work`, but the delivery contract provides no compatible isolated verifier for this job.

| Deliverable | Evidence | Result |
| --- | --- | --- |
| Reusable explicit-connection harness | `go test -count=1 -timeout 20m ./internal/connectors/native/dbtest`; unit coverage spans pinned tags, loopback allocation, resource ownership, cleanup/interrupt order, disk reporting, and serial slots. | pass |
| MySQL source connector | Focused native, definition, engine, and command-runner tests; matching executable binlog descriptor and live server proof. | pass |
| Real engine proof | Explicit-Podman tagged MySQL test completed successfully with reclaim enabled; cache confirmation followed. It exercised check, discovery, multi-page full read, incremental read, TLS modes, and real binlog CDC records. | pass |
| No retained engine resources | Explicit scoped container/volume/image listings were empty after the test. The task-owned Podman machine was then stopped and removed. | pass |
| Safe absence behavior | Missing opt-in/connection visibly skips. Once opted in, a startup/reachability error is a red test with a sanitized stage reason. | pass |
| #3902 pagination contract | MySQL exposes no HTTP/API direct-read surface; regression test makes the ETL/direct-read distinction explicit and live data proves internal multi-page draining. | pass |
| PostgreSQL TLS compatibility | Definition, TLS parser, and pgx pool tests prove matching mode/CA/server-name enforcement through Check/Catalog/Read; write path untouched. | pass |
| Generated/docs parity | `pm docs generate`, website catalog generator, docs validation, connector validation, and surface sync ran cleanly; `internal/cli` passed without blind transcript regeneration. | pass |

Verdict: **pass**. Remaining work is the PR's automated review/CI route, not an unproven product
acceptance item.
