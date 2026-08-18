# Code review — CLI package test-ceiling foundation

## Scope

- `internal/cli/github_transport_binary_test.go`
- `internal/cli/certify_cli_test.go`
- Phase evidence under `.planning/phases/issue-4015-cli-package-test-ceiling-r1/`

## Review result

No actionable findings.

| Area | Review evidence | Result |
| --- | --- | --- |
| Exactly-once construction | `sync.Once` establishes the fixture path/error before every caller returns; build arguments and source directory remain fixed to the existing helper contract. | pass |
| Failure propagation | A temporary-directory or build failure is stored and fails every requesting test; `TestMain` still runs the existing certification-budget guard. | pass |
| Resource lifecycle | The fixture has one package-owned temporary directory. `TestMain` removes it after `m.Run()` and turns cleanup failure into a non-zero exit before the explicit `os.Exit`. | pass |
| Test strength | The existing real-binary lifecycle proof gained an identity assertion; its command, independent root, provider behavior, approvals, readback, and cleanup assertions remain unchanged. | pass |
| Concurrency | `sync.Once` supports future concurrent callers. `go test -race -timeout 20m ./internal/cli -run '^TestPMBinaryExecutesIssueLabelWarehouseTransportLifecycle$'` passed in 61.353s. Package-wide parallelism is intentionally not introduced because process-global `t.Setenv` and `t.Chdir` tests remain. | pass |
| Security / boundary | No external data reaches `exec.Command`; the existing constant `go build -o <temporary-path> ./cmd/pm` call remains. `make connector-boundary` and aggregate `make verify` report zero findings. | pass |
| Static analysis | `gofmt`, `go vet ./...`, `make lint`, and `git diff --check` passed. | pass |

## Disposition

- Accepted findings: none.
- Declined findings: none.
- Deferred findings: none.
- Automated PR review: pending PR creation. Claude is the primary route; Copilot is fallback-only if Claude coverage is unavailable.
