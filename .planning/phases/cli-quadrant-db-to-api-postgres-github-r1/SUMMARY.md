# Summary — DB → API PostgreSQL → GitHub route R1

The shipped `pm` binary now accepts the exact PostgreSQL polling-watermark source descriptor for the closed GitHub issue-label destination. It maps only the two definition-declared inputs, `target_issue` and `label`, into the two definition-owned actions:

- `full_append` / append strategy → `add_issue_labels`.
- keyed `incremental_upsert` / merge strategy → `set_issue_labels`.

The destination reopens the durable warehouse workset, verifies the row equals the preview-bound target pair, writes with approval, reads the provider back, then checkpoints. Unsupported actions, ineligible modes, null/malformed rows, and tombstones are refused before provider write I/O.

Proof is deliberately split:

- The Docker PostgreSQL + `httptest` GitHub test is deterministic, CI-safe simulated coverage through the production HTTP boundary, including zero-row, null, replay, interruption/resume, and delete-unavailable edges.
- The separately env-gated live test uses real GitHub HTTPS in retained private `karthik-sivadas/pm-parity-proof-db-to-api`. Its final independent labels are issue #1 `pm-db-api-live-add` and issue #2 `pm-db-api-live-set`.

This proves the DB→API route only. GitHub live write coverage remains two actions of 607; the other 605 actions and the API→API quadrant are not certified by this change.
