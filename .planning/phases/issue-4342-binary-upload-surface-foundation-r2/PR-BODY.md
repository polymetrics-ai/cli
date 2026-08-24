Refs #4342

## Summary

- Add the declaration-only `binary_upload` CLI intent and route it through the existing approval-bound write-plan lifecycle; arbitrary URL, body, path, cap, and media choices are not accepted.
- Promote GitHub release-asset upload to that intent with its declared file field, 64 MiB cap, and media allow-list; regenerate its manual, skill, website, certification, evidence, and subject projections.
- Add a distinct upload certification capability/candidate/stage/sweep class. A blocked or plan-only command is reported `blocked`/`not_live`, never `pass`; `file_upload` remains declarable but non-executable.
- Bind the public GitHub upload host only to the declared public API origin; retain only `201` as a success; withhold `file_path`; enforce raw/base64 media declarations; and accept user files relative to the project root rather than `.polymetrics`.
- For this one `binary_upload` command, change the lifecycle from the inherited generic `reverse_etl` optional-preview behavior to required persisted preview before its human-only approval token. The existing generic no-confirmation `reverse_etl` behavior is intentionally unchanged and out of scope here.

## Behavioral evidence

- Red/green and deliberate-break records are in `.planning/phases/issue-4342-binary-upload-surface-foundation-r2/TDD-LEDGER.md`.
- The deliberate breaks removed planner admission and then the binary persisted-preview gate; the public commandrunner and App lifecycle tests failed, then passed after restoration.
- The App lifecycle test covers project-root source resolution, withheld path state/output, no pre-preview token or I/O, persisted preview, approval, changed-file zero-I/O refusal, exact bytes/SHA-256, and the retained `201` receipt. Engine tests separately prove Enterprise-to-public-host refusal, exact-`201` policy, and media-policy pre-I/O refusal.
- The fresh authorized GitHub proof uploaded 32 bytes to `uploads.github.com`, received `201`, read back an identical SHA-256, and then verified oversize, arbitrary-media, and missing-file refusal with no partial provider asset. It deleted the proof asset and retained the draft empty for audit. The bounded non-secret record is `LIVE-PROOF.md`. The generic certification stage remains `not_live`, not `pass`, until it can own transfer/read-back/cleanup evidence itself.
- F-4343-01 through F-4343-08 are closed by their declared origin, redaction, exact-status, media, lifecycle, real-provider, guidance, and project-root fixes. F-4343-09 is fixed only for `binary_upload`: plan has no token, persisted preview mints it, and run requires that preview. The inherited generic no-confirmation `reverse_etl` behavior remains explicitly out of scope.

## Verification

- Passed full changed-package suites: `GOFLAGS='-p=3' go test -count=1 -timeout 20m ./internal/connectors/engine`, `./internal/connectors/commandrunner` (31.809s), `./internal/app` (493.942s), `./internal/connectors/certify`, `./internal/connectors/bundleregistry`, and `./internal/connectors`.
- Passed final generators and artifacts: `connectorgen validate`, `surface-sync --check`, `operation-evidence --check`, `certification-subject --check`, GitHub certification matrix/candidates/sweep checks, and `git diff --check`.
- Passed `GOFLAGS='-p=3' go vet ./...`, `go build ./cmd/pm`, `make tidy-check`, `docs-check-no-build`, `smoke-no-build`, `agent-contract-check`, `github-parity-artifacts-check`, `connector-boundary`, `connector-canon-check`, `release-workflow-check`, and `lint`.
- Passed `cd website && npm run typecheck && npm run test:scripts` (34 tests). `pm github`, `pm help github`, and `pm github releases assets upload --help` confirmed contextual/bare help; a fresh initialized project returns `error: missing --credential` for the public command.
- Aggregate `go test ./...` and `make verify` were deliberately not run as single commands under the repository's memory-bound guidance; CI retains those aggregate checks.

## Delivery record

- GSD discuss/plan/execute/verify/review prompts and manual fallback evidence are recorded in the issue phase artifacts.
- Required skills: golang-how-to, golang-cli, golang-testing, golang-safety, golang-security, golang-error-handling, golang-design-patterns, golang-structs-interfaces, golang-documentation, vercel-react-best-practices, and vercel-composition-patterns.
