# Code review — Issue #4083

Review mode: inline standard review. The canonical contract prohibits spawning
the GSD reviewer/fixer roles in this worktree.

## Reviewed scope

- `internal/connectors/native/dbtest/harness.go`
- `internal/connectors/native/dbtest/harness_test.go`
- `internal/connectors/native/mysql/mysql_integration_test.go`
- `internal/connectors/native/dbtest/README.md`
- `AGENTS.md`

## Findings

No unresolved critical, warning, or informational finding.

Reviewed specifically:

1. Docker and Podman command construction always carries the supplied socket;
   no environment/default/context selection was introduced.
2. The direct-local-Unix validation still refuses named, remote, hosted-Unix,
   and control-character endpoint forms for both runtime selections.
3. Docker daemon identity and data-root capacity are rechecked before target
   commands; no unproved VM/remote image-store capacity is accepted.
4. Docker/Podman absence classification does not swallow permission or daemon
   errors, and live test configuration cannot turn a missing runtime into a
   passing skip.
5. No connector production implementation, credentials, dependency, generic
   runtime command path, or external database endpoint entered scope.

## Evidence

`go test -race -timeout 20m ./internal/connectors/native/dbtest`, scoped vet,
package tests, tagged skip/fail-closed witnesses, build/docs validation, lint,
and individual repository contract/generator/boundary/release gates passed as
recorded in `VERIFICATION.md` and `TDD-LEDGER.md`.
