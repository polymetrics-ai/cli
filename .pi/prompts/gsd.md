---
description: Run an official GSD Core workflow through the repo-local Pi adapter
argument-hint: "<command> [args...]"
---
Run the repo-local official GSD Core adapter for Pi.

1. Run:

```bash
scripts/gsd doctor
scripts/gsd prompt $ARGUMENTS
```

2. Read the generated prompt and execute it using Pi tools.
3. Read `.agents/agentic-delivery/references/required-skills-routing.md` and load the required Go/design skills for implementation, review, CLI, docs, website, or connector work.
4. Record the exact `scripts/gsd ...` command and required skills used in planning traces when updating `.planning/`.
5. Follow `AGENTS.md` safety gates: no secrets, no new dependencies without approval, no credentialed connector checks unless requested, no generic raw write tools, and reverse ETL remains plan → preview → approval → execute.
6. For runtime/RLM/Pi-agent work involving Podman, PostgreSQL, DragonflyDB/Redis-compatible coordination, Temporal, `pm runtime`, `pm rlm`, `pm agent image`, `pm worker`, or website architecture docs, read `.agents/agentic-delivery/references/runtime-rlm-website-integration.md`.
7. For CLI command, flag, help, output, connector-surface, or website-doc changes, read `.agents/agentic-delivery/references/cli-help-docs-website-parity.md` and include runtime help, bare namespace behavior, `docs/cli/**`, website docs, generated help/manual artifacts, and tests in the plan.

If `$ARGUMENTS` is `doctor`, `list`, `version`, `sources <command>`, or `verify-pi`, run `scripts/gsd $ARGUMENTS` and report the result instead of starting a workflow.
