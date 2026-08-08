# UAT — MySQL container harness R1

This firstmate task uses the documented inline GSD fallback. The project-local adapter resolved
`verify-work`, but the delivery contract provides no compatible isolated verifier for this job.

| Deliverable | Evidence | Result |
| --- | --- | --- |
| Reusable direct-endpoint harness | `go test -count=1 -timeout 5m ./internal/connectors/native/dbtest`; unit coverage spans pinned tags, endpoint identity/capacity proof, loopback allocation, resource ownership, cleanup/interrupt order, disk reporting, and serial slots. | pass |
| MySQL source connector | Focused native, definition, engine, and command-runner tests; internal reader coverage only. Public CDC remains false with no public descriptor or executor. | pass |
| Real engine proof | The prior task-owned-machine result predates the direct-endpoint contract; outer validation must re-run the tagged MySQL proof with a proven local endpoint. | pending |
| No retained engine resources | Unit coverage proves removal only for generated container/volume/run-image resources; the source image is retained. | pass |
| Safe absence behavior | Missing opt-in/endpoint visibly skips. Once opted in, a startup/reachability error is a red test with a sanitized stage reason. | pass |
| #3902 pagination contract | MySQL exposes no HTTP/API direct-read surface; regression test makes the ETL/direct-read distinction explicit and live data proves internal multi-page draining. | pass |
| PostgreSQL TLS compatibility | Definition, TLS parser, and pgx pool tests prove matching mode/CA/server-name enforcement through Check/Catalog/Read; write path untouched. | pass |
| Generated/docs parity | `pm docs generate`, website catalog generator, docs validation, connector validation, and surface sync ran cleanly; `internal/cli` passed without blind transcript regeneration. | pass |

Verdict: **focused endpoint-contract verification passed**. The outer executor owns the remaining
live, PR, and CI phases.
