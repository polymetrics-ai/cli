# CONTEXT — issue #3852 output-policy declaration

## Phase mapping

Issue #3852 is a shared connector-schema foundation in roadmap workstream 3, Direct-Read,
Binary, and Native Surface Parity. `gh-axi issue subissue list 3852` reported zero sub-issues on
2026-08-06; this branch is the one issue-scoped delivery slice.

## Locked decisions

- A connector must be able to declare the already executable complete-output write policies:
  `none` and `json`. `json` returns the decoded response unchanged; `none` intentionally returns
  no response body.
- The CLI schema's closed enum must be the union of the direct-read and direct-write runtime
  policy sets, while retaining `binary_file_bounded` as an existing binary-download compatibility
  declaration. No existing policy or bundle is migrated in this issue.
- A regression test must compare the schema enum against the runtime policy registries so either
  a newly supported runtime policy without a schema entry or a new unhandled schema entry fails
  tests.
- The initial TDD test must fail on the current schema while also proving the direct-write runtime
  accepts and executes `json` without changing its body.
- No redaction behavior changes, connector bundle rewrites, raw-body escape hatches, credentials,
  live-provider calls, or command-runner changes in #3771-owned functions are in scope.

## Implementation boundary

- `internal/connectors/engine/schema/cli_surface.schema.json`: declaration enum only.
- `internal/connectors/commandrunner/runner.go`: only a behavior-preserving representation of the
  existing direct-read/write policy lists, outside #3771's owned functions, if needed for the
  drift test to enumerate both runtime sets.
- `internal/connectors/commandrunner/runner_test.go`: schema/runtime drift test.
- `internal/connectors/engine/bundle_test.go`: RED-to-GREEN bundle-load test which also exercises
  the existing `json` direct-write response policy.
- `docs/migration/conventions.md`: authoring guidance directs complete-output writes to `json` and
  no-body writes to `none`; legacy redacting/narrow policies remain documented as retained
  compatibility choices, not defaults.

## CLI/docs parity disposition

This changes connector command metadata but adds no CLI command, flag, help topic, manual page, or
website page. Runtime help, bare namespace behavior, `docs/cli/**`, website docs, generated manual
artifacts, and completions are therefore not applicable. The connector authoring guide is the
applicable documentation surface and must be updated and grep-checked.

## GSD execution note

The adapter was healthy (`scripts/gsd doctor`; 69 commands), and the sources for
`discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review` all resolved.
Commands are executed inline because this issue's delivery contract forbids spawned GSD roles.
That documented fallback preserves the discussion, TDD, verification, and review artifacts.
