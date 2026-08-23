# Issue 4325 — Verification Checklist

## Required structural checks

- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] `go run ./cmd/connectorgen surface-sync --check`
- [ ] `make connector-runtime-preflight`
- [ ] `make connector-canon-check`
- [ ] `go test -timeout 20m ./cmd/connectorgen ./internal/connectors/commandrunner ./internal/cli`
- [ ] `go build ./cmd/pm`

## Behavioral checks

- [ ] Fresh credential-free source retrieval and exact method/path comparison
      for each non-GitHub batch lock.
- [ ] Built-binary probes for a representative implemented command in every
      applicable class; expected result is exactly the credential boundary.
- [ ] CircleCI, Sentry, and Vercel no longer return `unknown command` for their
      adopted terminal commands.
- [ ] `pm connectors`, `pm connectors inspect <name> --json`, and changed
      command help retain discovery/help behavior.
- [ ] Every batch connector’s inspection retains live certification pending.
- [ ] GitHub lock/descriptor byte and SHA-256 assertions equal the captain’s
      required values; `github/rate_limits.json` is unmodified.

## Release checks

- [ ] Full serial `make verify` immediately before every push.
- [ ] Independent Gate B rerun returns GO.
- [ ] GSD `verify-work` and `code-review` records are complete, including all
      actionable finding dispositions.
- [ ] PR is conventional-commit titled, contains `Refs #4325`, and API base
      read-back returns `main`.
