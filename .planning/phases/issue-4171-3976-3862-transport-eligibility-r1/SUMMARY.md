# Summary — transport source eligibility club (#4171, #3862; #3976 deconflicted)

## Delivered

- GitHub's declaration now contains a concrete, definition-matched source allowlist, including
  `commits`; the registry still refuses an undeclared stream before executor lookup or I/O.
- The general declarative source owns bounded collection batching, explicit `max_pages` semantics,
  source-bound checkpoint candidates, empty-result handling, and resume/replay behavior.
- PostgreSQL `polling_watermark` remains `planned` with its blocking reason. The attempted adapter
  was removed because its bind lived inside `ReadTransport` after authentication/catalog I/O rather
  than in a shipped production preflight; PR 4175 owns #3976.
- Cross-family production composition is asserted through `app.Open`; the binary integration test
  holds the GitHub → warehouse → PostgreSQL certification path, pending the known live limits.

## Verification

Red: retained failed-contract traces in `traces/`.

Green: focused package tests, vet/build, docs/skills/golden generation checks, connector surface
drift checks, contract checks, and GSD workflow verification are recorded in `VERIFICATION.md`.

The post-PR certification gap was compared cleanly before change: the merge base passed while this
branch failed from its renamed declaration-owned source reference. Both certification registration
assertions were updated to the new exact `declarative_stream_source` ID and now pass without
weakening the missing-registration guard.

## Pending external evidence

The authenticated `rails/rails` commits certification was not run because the certification
credential is pending rotation; the Docker runtime is not retried. No live count is claimed.
