# Context — durable parking resume claim race

## Task Delivery Header

- Issue: standalone direct-PR reliability repair; no numbered issue was supplied
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: direct PR, with the API-reported base recorded after opening
- Working branch: `fm/cli-durable-parking-resume-race-flake-r1`
- Task: remove the real cross-process durable-parking resume-claim race exposed
  by `TestCLIDurableParkingAdmissionAndResumeAcrossKilledProcess/resume-race`.
- Verification: red-first targeted test, a 20-consecutive-run process test under
  recorded machine load, race/package checks, build, relevant individual verify
  gates, GSD verification, and inline code review.

The required task-delivery-header template is absent from this integration base.
This document is the repository's documented manual fallback.

## Discuss-phase decisions

- Preserve process-level instrumentation that reports child exit code and output.
- Treat the existing test failure as a production coordination failure until the
  production call path proves otherwise; do not add retries, skips, reduced
  concurrency, or longer timeouts.
- Scope is limited to durable parking coordination/composition and its focused
  regression tests. No public command, flag, help, manual, or website behavior
  changes are intended.
- Use the real CLI → `app.Open` construction path for acceptance coverage. Any
  store-level test is supplementary, not a substitute for that process test.
- Execute the GSD lifecycle inline: this non-Pi worker cannot provide the
  compatible isolated workflow roles, and the canonical single-worker contract
  prohibits role spawning.

## Required skills loaded

- `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`
- `golang-security`, `golang-safety`, `golang-design-patterns`
- `golang-structs-interfaces`, `golang-context`, `golang-concurrency`

## CLI help/docs/website parity

Not applicable unless investigation discovers a public-surface change: this is
an internal correctness repair with no command, flag, output contract, help,
manual, connector surface, or website-doc change.
