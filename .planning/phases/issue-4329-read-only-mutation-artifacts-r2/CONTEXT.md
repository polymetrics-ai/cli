# Issue #4329 — read-only mutation artifacts, r2

## Discuss-phase decision

The existing `source-cited-non-executable-mutation-foundation-r1` supports an
explicit per-operation disposition, but a connector which deliberately has no
write capability still needs every provider mutation enumerated by hand before
source import can retain it. That makes its independently valid read/ETL
surface all-or-nothing.

Source import will deterministically attach the existing source-cited
non-executable-mutation artifact to an exact provider mutation when, and only
when, the bundle declares `capabilities.write: false`. It keeps the immutable
source ID/method/path/citation and a named foundation gap; it creates no action,
request schema, transport, command, or partial claim. A manually authored
disposition remains supported for write-capable bundles and remains exact.

The automatic rule must not suppress an existing complete action or an
implemented action claim. A source-bound delete or reverse-ETL action with its
real executable contract remains implemented. `read_only` continues to mean a
non-mutating operation only; mutation artifacts are distinct.

Sentry and Vercel are acceptance vectors through hermetic, retained-source
fixtures shaped from their locked source identities. The Batch 1 worktree and
its definition/source bytes are read-only inputs to this foundation task.

## Inline GSD fallback

`scripts/gsd doctor`, command resolution, and generated prompts for
`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
`code-review` were run. Compatible isolated GSD workers are unavailable in this
direct-PR environment, so the lifecycle is executed inline and recorded here.
