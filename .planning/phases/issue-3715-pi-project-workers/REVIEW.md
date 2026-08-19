# REVIEW — issue #3715 Pi clean project-only workers

Mode: standard inline review from the generated `code-review` prompt. No reviewer role was spawned,
as required by the canonical flow and task scope.

## Review scope

- Canonical contract and Go projection generator/checker.
- Pi discovery, child invocation policy, generated workers, and runtime isolation test.
- Pi extension documentation and durable repository guidance.

## Dispositions

### R1 — clean role paths could follow a symlink outside the project

- Severity: warning
- Disposition: accepted and fixed during review.
- Finding: selecting only an expected filename was insufficient if that file could be a symlink to
  ambient/global content.
- Action: clean discovery now requires an exact regular file and fails closed; the executable Pi
  test creates that hostile symlink scenario. `TestSyncPiProjectionRejectsSymlinkParent` likewise
  proves generated-file sync stays inside its `os.Root`.
- Evidence: `bash scripts/tests/pi-clean-project-agents.sh` and
  `go test ./internal/agentcontract -count=1` pass.

## Final review

No critical, warning, or informational finding remains. The child process has three independent
delegation protections: a generated tool declaration without `subagent`, runtime stripping of
`subagent` from the bounded allowlist, and Pi `--no-extensions`; the existing depth guard is
retained. The full-file generator accepts only the six validated repository-relative projection
paths and creates new paths only for the required Pi targets.
