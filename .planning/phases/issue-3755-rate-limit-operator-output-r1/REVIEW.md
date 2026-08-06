# Code review — issue #3755 rate-limit operator output

`scripts/gsd prompt code-review 3755` was generated and applied inline. Pi has no compatible isolated reviewer and the canonical single-worker contract forbids role spawning, so this is the required manual fallback.

## Scope reviewed

- `RateLimitReport` data ownership, synchronization, fixed policy cap, cloning, JSON fields, and human rendering.
- Requester buffered/streaming activity callbacks and their retry-wait ordering.
- Engine declaration/selection bridge and the limiter's observational wait callback.
- ETL state persistence, failed-run behavior, human/JSON CLI rendering, test-only bundle integration, generated docs, and website data.

## Findings

- Critical: none.
- Warning: none.
- Info: removed the now-unused `failRun` wrapper after `RunETL` moved to the summary-aware failure path.
- Existing behavior, not introduced by this issue: `pm etl run --help` attempts execution instead of leaf help. No command or flag was added here; runtime/bare namespace manuals and generated documentation cover the changed output contract.

## Review conclusion

No reporting path accepts or renders a credential, binding, runtime subject, opaque scope, or `CredentialRevision`. The added callbacks cannot influence policy resolution, admission, reserve, retry, or sleep decisions. The report stores only coalesced scalar totals and at most 16 selected policy rows per connector.
