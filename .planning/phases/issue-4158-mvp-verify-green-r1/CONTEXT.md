# Context — issue #4158 / Production MVP verify green

## Task Delivery Header

- Issue: Refs #4158 — restore durable managed-target acknowledgement while preserving non-PostgreSQL history-route refusal; includes the Production MVP fresh-binary GitHub → warehouse failure as the suspected same root cause.
- Base branch: `integration/4015-mvp-flat-r1` at `ef3c71caf`.
- Merges into: `integration/4015-mvp-flat-r1` → `main`.
- Working branch: `fm/cli-mvp-verify-green-r1`.
- Delivery: direct PR to `integration/4015-mvp-flat-r1`; never `main`; API-report the PR base after opening.
- Target connector: `postgres` managed target. The GitHub binary test is an upstream reproducer only; no GitHub bundle or surface change is in scope unless evidence disproves the shared managed-target route cause.
- GSD mode: `discuss-phase --auto`, `plan-phase --tdd --skip-research`, `execute-phase --interactive`, `verify-work --auto`, and `code-review --depth=deep`, all executed inline because this task forbids role spawning.

## Locked decisions

- The named tests are product-acceptance evidence, not stale-test candidates. Their assertions will not be relaxed, skipped, or re-baselined.
- Start from the external binary exit code / durable acknowledgement, then establish the earliest divergence and perform a bisect from immediately before `#4150`.
- Treat trigger, masking condition, and visible symptom as separate facts.
- Preserve the typed, pre-I/O non-PostgreSQL history-route refusal if admission is widened.
- No credentials are read, logged, or placed in artifacts. Existing hermetic fixtures and live-test control seams are the only permitted evidence path.

## Falsifiers to run before a fix

1. If the fresh-binary test passes repeatedly at `ef3c71caf`, the asserted regression is not reproducible in this worktree; record the repeat results and investigate only a proven environment delta.
2. If the PostgreSQL control test passes on `ef3c71caf`, it cannot support the shared-cause hypothesis; retain it as disconfirming evidence and scope this PR to the binary defect.
3. If the binary failure remains on the parent of `#4150`, neither `#4150` nor `#4155` introduced it; do not widen their predicate without a different causal path.
4. If a single route-admission condition flip does not change the observed result, route admission is not the causal divergence.
