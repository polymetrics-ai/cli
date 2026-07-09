# CLI Help, Manual, Docs, and Website Parity

Use this reference for every issue that adds, removes, renames, or changes a CLI command, subcommand, flag, output format, connector surface, help topic, generated manual page, or website documentation page.

## Parity rule

A CLI feature is not complete until all applicable surfaces are updated together:

1. **Runtime help** — `pm help <topic>`, `pm <command> --help`, and namespace bare invocations such as `pm connectors` show useful contextual help.
2. **Bare namespace behavior** — command groups with no action selected should render their help/subcommand summary and exit successfully instead of failing with a confusing missing-action error. Examples include namespace commands like `pm connectors`, `pm credentials`, `pm etl`, `pm reverse`, `pm flow`, and similar parent commands.
3. **CLI manual docs** — `docs/cli/**` contains the command, flags, examples, JSON/output behavior, safety notes, and cross-links.
4. **Website docs** — `website/**` documentation mirrors the CLI manual changes or explicitly links to the canonical CLI doc page.
5. **Generated/help previews** — any generated command reference, help preview, golden test fixture, or docs index is regenerated or updated.
6. **Completions/discovery** — command lists, help indexes, completions, and discoverability metadata include the new/changed command when applicable.
7. **Safety and machine-readable behavior** — docs and help mention `--json`/machine-readable output when supported, credential handling, write gates, and reverse ETL plan → preview → approval → execute semantics when relevant.

## Required implementation checklist

For CLI behavior-changing work, complete or explicitly mark not applicable:

- [ ] `pm <namespace>` with no subcommand prints namespace help/subcommand summary and exits 0 for command groups.
- [ ] `pm help <topic>` resolves the command or topic.
- [ ] `pm <command> --help` is accurate and includes examples or links when useful.
- [ ] Invalid actions still return a usage error and do not hide real failures behind help output.
- [ ] JSON output and stdout/stderr behavior are documented when supported.
- [ ] `docs/cli/<topic>.md` or relevant manual docs are updated.
- [ ] Website docs under `website/**` are updated or intentionally linked to canonical docs.
- [ ] Generated help/manual fixtures or docs indexes are updated.
- [ ] Tests cover help rendering, bare namespace behavior, and docs/manual parity when code changes.
- [ ] PR body lists the CLI help/manual/website parity checks and any intentional exemptions.

## Suggested verification commands

Use commands appropriate to the issue. Prefer existing project helpers when available.

```bash
pm help <topic>
pm <namespace>
pm <command> --help
pm <command> --json --help  # only if supported and meaningful
rg -n "<command>|<flag>|<topic>" docs/cli website
```

When code changes, add targeted Go tests for:

- bare namespace command behavior;
- `help` topic resolution;
- command/flag help text;
- generated docs or golden help output.

## Hard stops

- Do not add frontend/docs dependencies without human approval.
- Do not run credentialed connector checks for docs parity unless explicitly requested.
- Do not execute reverse ETL while validating docs/help parity.
- Do not expose generic shell, generic HTTP write, or generic SQL write tools in docs, help, or website examples.
