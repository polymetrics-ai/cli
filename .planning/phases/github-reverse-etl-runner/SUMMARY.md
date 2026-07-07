# Summary

This slice turns GitHub write-backed command-surface entries into safe reverse ETL command plans.

Implemented behavior:

- `pm github issue create --title ... --credential ...` creates a stored connector command plan.
- `pm github issue close --plan <id> --preview --json` previews without external writes.
- `pm github issue close --plan <id> --approve <token> --json` executes through the existing
  connector write engine after approval.
- The generic command runner remains read-only/non-mutating.
- Commands without explicit `record.*` flag mappings remain blocked.
- `--plan` execution through a provider command is bound to connector-command plans for the same
  connector and command path; normal reverse plans must still use `pm reverse run`.

Initial GitHub command mappings added:

- `issue create`
- `issue close`
- `repo deploy-key add`

Follow-up subissues should expand mappings for PRs, labels, releases, workflows, rulesets, arrays,
objects, and command-specific constants.
