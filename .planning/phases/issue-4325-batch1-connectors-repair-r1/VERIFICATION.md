# Issue 4325 — Verification Checklist

## Required structural checks

- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`
- [ ] `go run ./cmd/connectorgen surface-sync --check`
- [ ] `make connector-runtime-preflight`
- [ ] `make connector-canon-check`
- [ ] `go test -timeout 20m ./cmd/connectorgen ./internal/connectors/commandrunner ./internal/cli`
- [x] `go build ./cmd/pm` (2026-08-25 Jira slice)

## Behavioral checks

- [x] Baseline red: source-import fails because batch connector source locks are
      absent from current main; built binary probes for CircleCI, Sentry, and
      Vercel return `unknown command` (2026-08-23).
- [ ] Fresh credential-free source retrieval and exact method/path comparison
      for each non-GitHub batch lock.
- [ ] Built-binary probes for a representative implemented command in every
      applicable class; expected result is exactly the credential boundary.
- [ ] CircleCI, Sentry, and Vercel no longer return `unknown command` for their
      adopted terminal commands.
- [x] `pm connectors`, `pm connectors inspect jira --json`, and the changed
      Jira command `--help` retain discovery/help behavior (2026-08-25).
- [ ] Every batch connector’s inspection retains live certification pending.
- [ ] GitHub lock/descriptor byte and SHA-256 assertions equal the captain’s
      required values; `github/rate_limits.json` is unmodified.
- [x] Jira direct-write subset: manifest-reserved
      `api op-798e4bdcb516fc99a56c6b35b2bc97e67b65830a72dc867eeab1bb261c01b320`
      passes real preflight and a no-credential binary probe stops at
      `missing --credential`;
      source-cited `removeGroup` and `addWatcher` remain deferred without
      connector-local conditional-query or scalar-body approximations.

## 2026-08-25 Jira slice disposition

`go run ./cmd/connectorgen source-import jira --check`, the focused real
commandrunner preflight test, `go vet ./internal/connectors/commandrunner`,
and `go build ./cmd/pm` pass. Targeted validation has exactly 23 unrelated
pre-existing Jira gaps and none for `resetUserColumns`. Global checks remain
honest blockers: `surface-sync --check` stops first at Docker Hub's missing
source descriptor, and the full commandrunner package has 21 existing Docker
Hub runtime-preflight failures. No full `make verify` or push was attempted:
the current instruction prohibits pushing and the Issue 4325 Batch 1 gate is
not yet green.

## Release checks

- [ ] Full serial `make verify` immediately before every push.
- [ ] Independent Gate B rerun returns GO.
- [ ] GSD `verify-work` and `code-review` records are complete, including all
      actionable finding dispositions.
- [ ] PR is conventional-commit titled, contains `Refs #4325`, and API base
      read-back returns `main`.
