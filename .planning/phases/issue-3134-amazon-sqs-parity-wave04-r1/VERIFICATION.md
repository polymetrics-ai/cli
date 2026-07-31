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
