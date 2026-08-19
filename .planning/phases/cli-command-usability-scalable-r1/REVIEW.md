# Issue #4193 inline code review

## Scope reviewed

- `internal/cli/cobra_router.go`
- `internal/cli/cli.go`
- `internal/cli/docs.go`
- CLI regression tests, generated CLI manuals, and transcript fixture

## Findings

No unresolved findings.

The review specifically checked that:

- no per-command leaf-help switch remains;
- unknown commands and malformed approval carriers remain usage failures;
- help resolution occurs before `withApp`, credential resolution, and
  required-flag validation for documented legacy and declared connector leaves;
- the contextual ETL transport manual is represented as a resolver field, not
  a router switch, and retains its dedicated security/flag documentation;
- the hidden `extract` command still rejects a bare invocation while rendering
  help for either help spelling; and
- generated docs and transcript fixtures reflect the actual final output.

## Review method

Manual diff review, `git diff --check`, focused red/green tests, exhaustive
legacy and declaration-derived command-surface sweeps, the complete
`internal/cli` package suite, and the repository gates in `VERIFICATION.md`.

GitHub's configured Claude review remains the primary automated route on PR
open; its result is external to this local review record.
