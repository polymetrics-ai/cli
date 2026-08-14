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
6. The Docker VM capacity fallback remains inside `dbtest.Config`: it requires
   a pre-cached pinned image, uses `--pull=never`, has no network, is read-only,
   drops capabilities, sets no-new-privileges and a PID bound, parses only a
   fixed POSIX `df` mount result, and idempotently cleans its unique ephemeral
   probe. A missing or malformed probe remains a pre-mutation refusal.
7. A pre-existing capacity probe is refused before its name is claimed for
   cleanup, and an indeterminate probe is removed only after its per-run Docker
   ownership label and immutable container ID are re-proven. The enabled MySQL entrypoint passes raw
   environment values to the shared validators so whitespace and control
   characters cannot normalize into an accepted runtime configuration.

## Evidence

`go test -timeout 20m -run
'^(TestDockerVMCapacityUsesOnlyAPreCachedLockedDownProbe\|TestDockerVMCapacityRefusesAPreexistingProbeWithoutClaimingItsCleanup\|TestDockerVMCapacityPreservesAProbeCreatedAfterPrecheck\|TestCloseRemovesOnlyLabelOwnedCapacityProbe\|TestDockerVMCapacityRefusesAnUncachedOrMalformedProbe)$'
./internal/connectors/native/dbtest` passed in 0.350s; the broader existing
dbtest and MySQL evidence is recorded in `VERIFICATION.md` and `TDD-LEDGER.md`.
