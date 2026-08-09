# Phase cli-unknown-subcommand-false-success-r1 — reject unresolved connector help paths

## GSD and skills

- GSD setup passed: `scripts/gsd doctor`; `go run ./cmd/agentcontractgen check`.
- Resolved commands: `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and
  `code-review` via `scripts/gsd sources <command>`.
- Manual-GSD fallback: execute generated `scripts/gsd prompt` instructions inline because Pi is
  unavailable and this task/canonical contract forbids role spawning.
- Skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
  `golang-security`, and `golang-spf13-cobra`.

## Scope

`internal/cli` only, plus the required GSD evidence. No connector definition, generated surface,
manual, or website doc changes are expected: the command syntax and help text do not change; only
the erroneous success exit for an invalid path is corrected.

## TDD execution plan

| Order | Slice | RED evidence | GREEN implementation | Guard |
| --- | --- | --- | --- | --- |
| 1 | Connector help resolution | A valid connector with a bogus deep path and `--help` exits 0 and renders root help. | Validate the normalized help path against the command surface before rendering help; return `connectorCommandUsageError` when unresolved. | Connector root, group, and declared deep command help still exit 0. |
| 2 | Error contract | The bogus JSON invocation is currently a successful command-manual envelope. | The same invocation returns exit 2 and `usage_error` whose message names the unresolved path. | A connector-level unknown remains exit 2 with the connector-only message. |
| 3 | Generated transcript | Add the deep invalid help invocation to the golden matrix before regeneration. | Regenerate the transcript and inspect the diff. | No unrelated golden entries change. |

## CLI parity checklist

- [x] Bare `pm <connector>` retains contextual root help and exit 0.
- [x] `pm <connector> --help` retains root help and exit 0.
- [x] Real group and deep-command `--help` retain exit 0.
- [x] Invalid action returns a usage error.
- [x] JSON errors use the existing structured error envelope.
- [x] CLI manual/website updates are not applicable: syntax and help content are unchanged.

## Commit checkpoints

1. Planning/TDD evidence.
2. Red regression test proof.
3. Green implementation, regenerated golden transcript, and scoped verification.
4. Review-fix checkpoint only for bounded in-scope findings.
