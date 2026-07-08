---
description: Run a repo-local, runtime-neutral GSD workflow prompt
argument-hint: "<prompt-name|doctor|sources|init> [args...]"
---
You are running the repo-local GSD command adapter, not Claude-specific slash commands.

1. Use `scripts/gsd doctor` first if the workflow has not been checked this session.
2. For workflow prompts, run:

```bash
scripts/gsd prompt $ARGUMENTS
```

3. Read the generated prompt carefully and execute it using the available Pi tools.
4. Record the exact `scripts/gsd ...` command used in any planning trace you update.
5. Preserve repo safety rules from `AGENTS.md`: no secrets, no credentialed connector checks, no reverse ETL execution without plan/preview/approval/execute, no new dependencies without human approval, and no `cmd/` or `internal/` edits for planning-only issue #122.

If `$ARGUMENTS` is `doctor`, `sources`, `workflows`, or `init`, run the corresponding `scripts/gsd $ARGUMENTS` command and report the result.
