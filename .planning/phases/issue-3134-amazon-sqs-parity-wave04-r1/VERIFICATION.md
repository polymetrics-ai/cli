# Verification Checklist — Amazon SQS parity wave04 r1

Required by task:

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/amazon-sqs` (focused temp defs root: 1 connector checked, 0 findings)
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/amazon-sqs' -count=1`
- [x] `go test ./internal/connectors/native/amazon-sqs -count=1`
- [!] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1` (passed earlier after golden refresh; post-hardening exact rerun was interrupted/timed out under host-wide concurrent parity test load before producing assertion output; relevant post-hardening subset covering golden transcripts, dynamic connector commands, connector catalog, root help, and inspect manifest passed in 146s)
- [x] `go build ./cmd/pm`
- [x] `make connector-boundary` (outcome clean)
- [!] `make verify` / `go test ./...` (attempted; full suites timed out in long-running `internal/cli` and `internal/connectors/certify` certification/golden tests while many other worktrees were concurrently running heavy `go test` processes on the same host; no Amazon SQS assertion failure was observed)
- [x] `git diff --check`

Optional/focused diagnostics if failures occur:

- [ ] `go test ./internal/connectors/commandrunner -run OperationDirectRead -count=1`
- [ ] `go test ./internal/app -run Reverse -count=1`

Results will be recorded before commit.

## 2026-08-02 corrective verification — issueguard removal

Forward corrective commit evidence:

- [x] `git diff --name-status origin/main...HEAD -- internal/coordination/issueguard cmd/prissueguard` showed the forbidden `internal/coordination/issueguard/guard.go` and `guard_test.go` deltas before restoration.
- [x] After restore, `git diff --exit-code origin/main -- internal/coordination/issueguard cmd/prissueguard` returned 0 for restored working tree content.
- [x] `go test ./internal/coordination/issueguard ./cmd/prissueguard -count=1`
- [x] `go run ./cmd/prissueguard --title "chore(amazon-sqs): stage unvalidated parity checkpoint" --body-file /tmp/pr3541-body-refs-3134.md`
- [x] `go test ./internal/connectors/native/amazon-sqs -count=1`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/amazon-sqs --json`
- [x] `go test ./internal/connectors/conformance -run '^TestConformance/amazon-sqs$' -count=1`
- [x] `git diff --check`
- [x] `go build ./cmd/pm`
- [x] `make connector-boundary`
- [ ] Normal push of corrective commit.
- [ ] Fresh no-mistakes validation with native Pi `openai-codex/gpt-5.6-sol` at `xhigh`, including post-guardrail revalidation and remote checks.

## 2026-08-02 review-finding fix verification

- [x] `go test ./internal/connectors/native/amazon-sqs -count=1`
- Focused coverage includes required reverse CLI metadata/examples, all ten formerly unexecuted typed operation paths, reachable sanitized certification errors, and pre-preview message-attribute validation.
- Per review-phase scope, no complete repository test, lint, push, PR, CI, live AWS, or credentialed connector phase was run.

## 2026-08-02 message-attribute semantic verification

- [x] `go test ./internal/connectors/native/amazon-sqs -count=1`
- Focused coverage validates SQS Number syntax/precision/range, BinaryValue standard base64 encoding, and AWSTraceHeader X-Ray shape before preview.
- No full repository, lint, push, PR, CI, live AWS, or credentialed connector phase was run.

## 2026-08-02 decoded-empty binary verification

- [x] `go test ./internal/connectors/native/amazon-sqs -count=1`
- Focused coverage rejects newline-only standard-base64 input that decodes to zero bytes before preview while preserving valid nonempty binary attributes.
- No full repository, lint, push, PR, CI, live AWS, or credentialed connector phase was run.

## 2026-08-03 CI verify staticcheck lint verification

- [x] `golangci-lint run ./internal/connectors/native/...` returns `0 issues` after the `strings.ContainsAny` correction (pre-fix run reproduced `S1003` at `message_attribute_values.go:17:28`, matching the PR #3541 `verify` annotation).
- [x] `go test ./internal/connectors/native/amazon-sqs -count=1`
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/amazon-sqs --json` reports no findings or warnings.
- [x] `pm docs validate --connectors-dir docs/connectors` passes.
- [x] Parity re-audit: all 23 official AWS SQS operations (https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_Operations.html) map to the `messages` stream, a typed direct read, a typed reverse-ETL write, or an explicit disposition; `api_surface.json` lists 23/23 endpoints with coverage mapping and `cli_surface.json` exposes 23/23 commands, all `[availability=implemented]` in `pm amazon-sqs --help`.
- No live AWS, credentialed connector, push, PR, or CI phase was run in this worktree; the commit is pushed by the no-mistakes pipeline.
