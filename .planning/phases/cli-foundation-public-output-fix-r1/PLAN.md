# Foundation public-output repair r1

## Task Delivery Header

- Issue: Refs #4307 — feat(engine): enforce declaration-owned headers and bounded transfer operations
- Base branch: `fm/cli-current-foundations-main-integration-r1`
- Merges into: `fm/cli-current-foundations-main-integration-r1` → `main`
- Delivery: Atomic committed and non-force-pushed green sub-groups on `fm/cli-foundation-public-output-fix-r1`, ready for Firstmate's later no-mistakes/PR stage; no merge or release.
- Working branch: `fm/cli-foundation-public-output-fix-r1`
- Task: Repair only the independent public-output and declaration-authority findings `FND-B10`–`FND-B14` and `FND-W02` from immutable commit `c9824b5837f487acaa2c2a39126d29cf401d7fb5`.
- Verification: Red-green-refactor targeted Go tests with hermetic provider doubles; affected-package tests, `go vet ./...`, `go build ./cmd/pm`, generator/parity gates as applicable, `git diff --check`, and the non-test `make verify` gates listed in AGENTS.md.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Ordinary provider values are unchanged while configured secrets and concrete encodings are masked. | fake | The brief prohibits credentials and requires hermetic doubles; a deterministic double returns ordinary IDs, occurrence IDs, token-shaped text, and configured-secret representations, and assertions distinguish every preserved and masked scalar. |
| Public cursors, direct-read receipts, and non-JSON diagnostics contain no configured secret. | fake | A hermetic response double emits JSON, plain text, and cursor/receipt forms with direct, escaped, and encoded secret values; serialized public output is asserted secret-free while ordinary values remain. |
| Invalid GitHub App restrictions fail before authenticated I/O. | fake | A GitHub App transport double counts requests; malformed restrictions return an error with count zero, while a valid restriction reaches the double once. |
| Binary downloads and status checks accept only declared safe parameter bindings. | fake | Operation doubles record a request count and complete request shape; undeclared and unsafe bindings fail with count zero, declared happy paths retain exact request values. |

## TDD slices and checkpoints

1. **Public value preservation and public-output masking — red → green → refactor.** Add failing tests for name-shaped SQS/ID/token values, cursor/receipt output, and printable JSON/non-JSON secret encodings. Replace heuristic redaction with configured-material-only sanitization. Commit and push only after the focused tests pass.
2. **GitHub App restriction admission — red → green → refactor.** Add valid, malformed, and edge restriction cases that observe authenticated provider I/O. Make malformed restrictions fail closed before the request is authenticated or sent. Commit and push after package tests pass.
3. **Binary/status declaration authority — red → green → refactor.** Add happy, undeclared, and unsafe binding tests at the installed operation path. Enforce the same declaration-owned admission before binary and status provider I/O. Commit and push after focused tests pass.
4. **Integration verification and review.** Run the adapter-recorded inline `verify-work` and `code-review` checks, record exact commands/results, then commit/push any review fixes separately.

## Lifecycle, skills, and parity

- GSD path: `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`, resolved through `scripts/gsd sources` and executed inline because no compatible isolated roles may be spawned.
- Loaded skills: `golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, and `golang-concurrency`.
- CLI parity: no command, flag, help, manual, or website surface is intentionally changed. Confirm the affected output remains machine-readable; record a not-applicable result only after inspecting the final diff.
- Exclusions: no source-import/certification changes, reverse-ETL action binding changes, production connector-definition changes, credentials, generated-runtime artifacts, or external provider calls.
