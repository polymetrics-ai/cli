# Issue 4325 — Batch 1 Connector Repair Context

## Decision record

- Repair all ten connector bundles as independent, sequential TDD slices whose
  aggregate acceptance target is the independent Gate B report.
- Treat the report's live provider method/path derivation as authoritative
  unless a fresh credential-free retrieval disproves it; do not lower a count
  or remove a route to make a declaration count agree.
- Maintain source integrity by re-pinning retrieved provider documents and
  regenerating all dependent bundle artifacts. GitHub is out of scope and its
  locked source/descriptor bytes and SHA-256 values are immutable acceptance
  assertions.
- Terminal reachability means dispatch resolves and, without a credential,
  returns `missing --credential`; `unknown command` is a defect. Live provider
  certification remains pending because this task has no authorized credentials.
- Every disabled row must either be connector-local `declaration-pending`, a
  defensible `unsafe-to-exercise`, or a real foundation gap that cites the
  current refusing `file:line`. `requires-elevated-scope` is not a valid reason.
- The recursive-schema importer dependency for Stripe is external issue #4323;
  sequence Stripe after the non-Stripe work and do not bypass importer refusal.
- No shared engine/runtime shim belongs in this issue. If one is necessary,
  create a separate foundation issue and stop the affected connector slice.

## CLI parity applicability

The repair changes connector command surfaces. Runtime command probes and
generated command discovery are in scope. Standalone manual/website prose is
not applicable unless source-generated documentation checks identify a changed
artifact; no hand-authored connector command help is introduced.

## Inline workflow fallback

The project-local GSD adapter generated the `discuss-phase` prompt. Compatible
isolated runtime agents are unavailable and the canonical contract disallows
role spawning, so this context records the required inline/manual execution.
