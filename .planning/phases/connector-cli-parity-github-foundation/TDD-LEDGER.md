# TDD Ledger: Connector CLI parity GitHub foundation

## Red

Command:

```bash
go test ./internal/coordination/issueguard
```

Expected failure:

```text
undefined: ValidatePRBody
```

Reason: tests were added before the issue-first PR guard implementation.

## Green

Commands:

```bash
go test ./internal/coordination/issueguard
go run ./cmd/prissueguard --title 'feat(github): add cli surface metadata' --body 'Closes #123'
go run ./cmd/prissueguard --title 'add cli surface metadata' --body 'no issue'
go test ./cmd/prissueguard ./internal/coordination/issueguard
go test ./...
```

Evidence:

- Targeted package tests passed.
- Valid PR title/body smoke check passed.
- Invalid PR title/body smoke check failed as expected.
- Full `go test ./...` passed.

## Harness gap

The GSD skill references `scripts/programming-loop.mjs` and `scripts/tdd-gate.mjs`, but this
worktree does not contain those helper scripts. This phase follows the same loop manually and records
evidence here.
