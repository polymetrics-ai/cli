# Context — CLI package test-ceiling foundation

## Task Delivery Header

- Issue: Refs #4015 — Production MVP certification foundation
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: A committed, pushed branch with a pull request open against `integration/4015-mvp-flat-r1`; its base is API-read-back verified and local repository verification evidence is recorded.
- Working branch: `fm/cli-cli-package-test-ceiling-r1`
- Task: Remove the `internal/cli` package's collision with Go's 20-minute per-test-binary timeout without weakening any test. Measure the existing suite, identify the proven execution bottleneck, implement one of the allowed generic mechanisms, and prove all existing tests still execute.
- Verification: Capture before/after test-name sets; record verbose timing evidence; run targeted and consumer tests, the repository verification entry points, and `./cmd/connectorgen`; if parallelism is selected, run the full package three consecutive times; verify the PR base through the GitHub API.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Existing tests remain included | live | `go test -list '.*' ./internal/cli` before and after creates equal normalized name sets; a missing test changes the set. |
| The timeout collision is materially removed | live | Full-package timing after the chosen mechanism stays at least 15% below the 20-minute ceiling while all assertions run. |
| The generic mechanism cannot silently omit tests | live | The implementation's regression test rejects an incomplete execution partition, if sharding is selected; otherwise the before/after name-set proof covers every test. |
| No regression is introduced | live | Targeted, consumer, and repository verification commands complete successfully; failures are recorded with their exact output and disposition. |

## Constraints and decisions

- The shared programme context was read from `/Users/karthiksivadas/karthik-agent-workspace/data/production-parity-shared-context.md` because it is not tracked on the integration base.
- No existing test may be skipped, shortened, tagged out, or otherwise weakened; real-binary integration coverage must remain intact.
- Do not make connector-specific shared-Go identifiers or allowlist changes.
- Measurement determines the mechanism. Candidate mechanisms are isolated top-level test parallelism, one-time binary construction, complete CI sharding, or package extraction. A larger redesign requires a `needs-decision` stop.
- GSD is executed inline because this non-Pi worker has no compatible isolated GSD runtime; this is a documented manual fallback, not a lifecycle waiver.
- Required skills loaded for investigation: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-benchmark`, `golang-performance`, and `golang-troubleshooting`. Load concurrency or CI-specific guidance before selecting those mechanisms.
