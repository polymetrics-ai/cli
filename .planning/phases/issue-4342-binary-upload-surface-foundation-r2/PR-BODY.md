Refs #4342

## Summary

- Add the declaration-only `binary_upload` CLI intent and route it through the existing approval-bound write-plan lifecycle; arbitrary URL, body, path, cap, and media choices are not accepted.
- Promote GitHub release-asset upload to that intent with its declared file field, 64 MiB cap, and media allow-list; regenerate its manual, skill, website, certification, evidence, and subject projections.
- Add a distinct upload certification capability/candidate/stage/sweep class. A blocked or plan-only command is reported `blocked`/`not_live`, never `pass`; `file_upload` remains declarable but non-executable.

## Behavioral evidence

- Red/green and deliberate-break records are in `.planning/phases/issue-4342-binary-upload-surface-foundation-r2/TDD-LEDGER.md`.
- The deliberate break removed planner admission: installed GitHub plus raw/base64/multipart commandrunner tests failed, then passed after restoration.
- Existing engine binary-write tests continue to cover exact-byte transfer and unsafe/missing/changed/oversize refusal before I/O. No credentialed provider call was authorized for this PR, so no live upload is claimed.

## Verification

- Passed: `go test` for engine, commandrunner, certify, and connectorgen (rerun after merging current main); `go vet ./...`; `go build ./cmd/pm`; and individual `make verify` component gates including lint, docs, certification projections, connector boundary, and release workflow.
- `pm github`, `pm help github`, and `pm github releases assets upload --help` confirmed contextual/bare help and the new public intent/approval text.
- Attempted `go test ./internal/cli`: generated skills were refreshed, but unrelated runtime tests require unavailable Redis endpoints `127.0.0.1:1` and `127.0.0.1:2`. Aggregate `go test ./...` and `make verify` were not run as single commands per repository memory-bound guidance; CI retains them.

## Delivery record

- GSD discuss/plan/execute/verify/review prompts and manual fallback evidence are recorded in the issue phase artifacts.
- Required skills: golang-how-to, golang-cli, golang-testing, golang-safety, golang-security, golang-error-handling, golang-design-patterns, golang-structs-interfaces, golang-documentation, vercel-react-best-practices, and vercel-composition-patterns.
