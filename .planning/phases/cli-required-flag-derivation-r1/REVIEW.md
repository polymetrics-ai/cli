# Code review — CLI required-flag derivation r1

## Scope and method

An inline manual code review covered the generic generator change, its
repository-wide invariant, command-runner typed error, CLI serialization, and
the generated GitHub artifacts. The review used:

- `git diff --check origin/integration/4015-mvp-flat-r1...HEAD`
- the focused and full changed-package tests recorded in `VERIFICATION.md`
- `make connector-boundary` and `make connectorgen-surface-sync`
- two complete generator passes followed by `git diff --exit-code`

## Findings and disposition

No unresolved findings.

- Requiredness only derives from `rest.parameters` with `in: path` and
  `required: true`; query and body contracts remain untouched.
- The focused regression test covers both a missing required field and an
  explicit `required: false`, and asserts required query input remains
  optional under this path-only rule.
- `MissingRequiredFlagError` preserves the existing message while enabling the
  CLI to classify the caller error as `usage_error` without text parsing. The
  command-runner and CLI tests both prove refusal occurs before provider I/O.
- The all-bundle test has no connector identifier or example-list exception,
  and the connector-boundary check passes without an allowlist edit.
- Generated files were inspected through their source inputs and deterministic
  generator output rather than hand-edited.

## Automated review route

This direct-PR branch will rely on the repository's automatic Claude review on
PR open. No manual `@claude review` request is made for the initial commit;
the PR body records the pending automatic-review status for the coordinator.

## Website generated-data gap review

After PR 4209's `Website generated data` failure, the repository generator ran
twice. The first and second passes produced no `website/**` diff, and
`git diff --exit-code -- website` passed. Therefore no generated website file
was staged or baselined without an intended required-flag change.
