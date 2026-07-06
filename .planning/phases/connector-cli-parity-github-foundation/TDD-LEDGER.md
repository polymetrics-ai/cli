# TDD Ledger: Connector CLI parity GitHub foundation

## Red

Command:

```bash
go test ./internal/coordination/issueguard
```

Expected failure:

```text
undefined: ValidatePR
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

## Automated Review Follow-Up

Added regression coverage after automated PR review:

- reject Conventional Commit scopes accepted by the older loose local regex but rejected by the
  repository `pr-title` workflow
- reject ambiguous issue relationships such as `Related to #123`, `Issue #123`, and
  `References #123`

Command:

```bash
go test ./cmd/prissueguard ./internal/coordination/issueguard
```

Evidence:

- Targeted guard tests passed.
- `make verify` passed after the guard changes.

## CodeRabbit Review Follow-Up

Added CLI exit-code coverage after CodeRabbit review:

- valid PR input returns exit code `0`
- invalid PR title/body returns exit code `1`
- unreadable body file returns exit code `2`

Command:

```bash
go test ./cmd/prissueguard ./internal/coordination/issueguard
```

Evidence:

- Targeted guard tests passed.
- YAML parsing, JSON parsing, PR body guard, and diff whitespace checks passed after the review
  fixes.

## Harness gap

The GSD skill references `scripts/programming-loop.mjs` and `scripts/tdd-gate.mjs`, but this
worktree does not contain those helper scripts. This phase follows the same loop manually and records
evidence here.
