# Discussion log — CLI package test-ceiling foundation

## GSD lifecycle evidence

- `scripts/gsd doctor` passed.
- Resolved `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review` through `scripts/gsd sources`.
- Generated the `discuss-phase` prompt with `scripts/gsd prompt discuss-phase issue-4015-cli-package-test-ceiling-r1`.
- Inline/manual fallback: this task runs outside Pi and the canonical contract forbids spawning the GSD role set. The worker will execute the generated lifecycle artifacts directly.

## Fixed decisions from the brief

1. The branch starts from `origin/integration/4015-mvp-flat-r1` at `aff232e9302afc5b856c58d1d068497337cc99cf`.
2. The 20-minute timeout is per test binary; the target is to remove this package's proximity to that ceiling generically, not to change an individual test.
3. Existing top-level test names are a release invariant. Capture the exact set before making production changes and compare it after.
4. Measure verbose per-test durations, real-binary build reuse, top-level parallelism, shared state, and available CI resources before selecting an implementation direction.
5. The PR will target `integration/4015-mvp-flat-r1`, reference #4015, and record before/after timing, rejected alternatives, red/green evidence, verification, and review routing.

## Open implementation question

Which allowed generic mechanism is both safe and sufficient? This is intentionally unresolved until the measured suite and shared-state audit are complete.
