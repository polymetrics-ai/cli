# Descriptor-Free Retention Admission — Context

**Issue:** Refs #4283 — Top-100 Batch R1 mapping-control sub-slice
**Base:** `fm/cli-top100-declaration-batch-r1@c9ae575a734514b728a5e6add7ff8b0e55233437`
**Working branch:** `codex/4283-artifact-projection-b2-r1`
**Delivery:** local green commit only; no push, no integration, independent review required.

## Decision

Some Batch R1 connectors retain an immutable source lock and a source-lane
matrix but intentionally do not retain the large historical canonical
descriptor. Their mapping evidence must remain inspectable without treating
that retained evidence as an executable declaration.

This slice adds a deliberately narrow `retention_only` enabled-contract mode.
It is mapping admission only. It may waive the *descriptor-presence* check
only after the existing source-lock bridge verifies an exact primary-lock
digest/byte identity and exact source-operation-ID reconciliation. Every lane
must remain nonimplemented and source-only. Any executable claim continues to
require the canonical descriptor and its existing projection validation. The
actual loaded bundle must additionally have no operations, writes, streams,
selected sync transport, or CLI command marked `implemented`.

## Frozen evidence

| Connector | Lock | SHA-256 | Bytes | Operations |
| --- | --- | --- | ---: | ---: |
| Jira | `sources/jira-operation-source-lock.json` | `e7136af43bf72cd4ea5ada91ec665b318b60008814122461d4436a43b6c732bf` | 2,456,011 | 617 |
| Sentry | `sources/sentry-operation-source-lock.json` | `b71216654e44cc18f5e262fbb5075df67f1504a123d4bcb51cc8e8cc74ebd435` | 3,868,570 | 223 |
| Vercel | `sources/vercel-operation-source-lock.json` | `74cb7ff3dc0b89cc344b13ac9c6d5f1d9b7d7a9356cfd6b5a779da51fd43da28` | 10,463,249 | 400 |

The historical descriptor imports that this change must not restore are
21,861,518 bytes (Jira), 3,305,062 bytes (Sentry), and 13,711,335 bytes
(Vercel).

## Scope and non-goals

- In scope: enabled-contract validation/schema, source-projection admission,
  Atlas/terminology guidance, and focused regression tests.
- Out of scope: source locks, matrices, crosswalks, connector definitions,
  executable artifacts, runtime/transport/credential/certification behavior,
  historical descriptor import, generated manuals, and skills.
- `retention_only` does not make a connector executable, enabled, certified,
  or provider-parity complete. Source IDs are opaque evidence keys, not paths:
  ordinary spaces and `/` remain valid provider spelling; only empty and
  control-character IDs are rejected.

## Process evidence

CodeGraph is unavailable because this worktree has no `.codegraph/` directory.
The GSD lifecycle prompts were generated with `scripts/gsd prompt`; role
spawning is intentionally not used because this bounded task forbids parallel
sub-agent delegation. The required Go engineering, connector-lane build-order,
Foundation Atlas, and repository agent-contract procedures were read before
the first code edit.
