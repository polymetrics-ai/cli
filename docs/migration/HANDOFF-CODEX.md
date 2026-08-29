# Connector Migration Handoff

The 2026-07-04 parallel-rollout handoff was archived at
[docs/connector-canon/archive/superseded-repository-planning/migration-handoff-codex-2026-07-04.md](../connector-canon/archive/superseded-repository-planning/migration-handoff-codex-2026-07-04.md).
It contains stale branch state, inventory figures, and multi-worker instructions.

For every connector change, begin with the current
[connector delivery canon](../connector-canon/INDEX.md), then read
[migration conventions](conventions.md) for bundle-authoring mechanics and the
architecture design only for still-applicable implementation detail. Do not use
the archived rollout as a task template. Connector work may use a bounded named
cohort with an immutable source-lock ledger; a named shared foundation and its
affected mappings may share that PR after the Foundation Atlas check, without
relaxing source evidence, TDD, review, CI, safety, or warehouse mediation.
