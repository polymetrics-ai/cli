# Summary — transport source eligibility club

## Delivered

- GitHub's declaration now contains a concrete, definition-matched source allowlist, including
  `commits`; the registry still refuses an undeclared stream before executor lookup or I/O.
- The general declarative source owns bounded collection batching, explicit `max_pages` semantics,
  source-bound checkpoint candidates, empty-result handling, and resume/replay behavior.
- PostgreSQL `polling_watermark` is implemented through its definition-owned native binder, strict
  `(cursor, primary-key tuple)` keyset reader, `PollingPreflight`, and the shared
  `engine.PollingSourceExecutor`.
- Cross-family production composition is asserted through `app.Open`; the binary integration tests
  hold the real GitHub → warehouse → PostgreSQL certification path.

## Verification

Red: retained failed-contract traces in `traces/`.

Green: focused package and race tests, vet, build, docs generation/check, connector generation and
surface drift checks, contract checks, and GSD workflow verification passed. See `VERIFICATION.md`
for commands and the edge matrix.

The post-PR certification gap was compared cleanly before change: the merge base passed while this
branch failed from its renamed declaration-owned source reference. Both certification registration
assertions were updated to the new exact `declarative_stream_source` ID and now pass without
weakening the missing-registration guard.

## Pending external evidence

The live Docker/PostgreSQL and authenticated `rails/rails` commits certification was not run: the
task explicitly says not to retry the unavailable shared container runtime and the certification
credential is pending rotation. No live count is claimed.
