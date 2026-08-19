# Code review — cli-unknown-subcommand-false-success-r1

## Method

Manual inline review, the repository-mandated fallback because the canonical single-worker contract
forbids spawning a reviewer role. Reviewed `internal/cli/cli.go`, the new CLI regression coverage,
and the regenerated transcript diff.

## Findings

No actionable findings.

- The resolver validates only the display/help path, then returns the existing
  `connectorCommandUsageError`; ordinary execution continues to use `commandrunner.Preflight`.
  This avoids creating a parallel execution validator.
- A connector root (empty normalized path), an exact declared command, and a declared one-segment
  group keep their prior successful manuals. All other requested help paths now consistently report
  the complete unresolved path.
- The error stays a `usageErrorf`, so the existing exit-2 mapping, terminal sanitization, and JSON
  `usage_error` envelope are reused rather than duplicated.
- The golden regeneration contains one intentional entry only; it records the previously silent
  deep help path under `--json` and no connector metadata or generated command surface changed.
