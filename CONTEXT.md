# Polymetrics CLI Context

`pm` is a local-first CLI for ETL and reverse ETL. It is intended for humans and LLM agents.

## Core Syntax

```bash
pm <command> [options]
pm help <topic>
pm connectors inspect <connector> --json
pm etl run --connection <name> --stream <stream> --batch-size 100 --json
pm reverse plan <name> ...
pm reverse preview <plan-id> --json
pm reverse run <plan-id> --approve <approval-token> --json
```

## Safety

- JSON output is for agents; stderr is for human diagnostics.
- Connector manifests expose secret field names only, never secret values.
- Reverse ETL writes require approval.
- Dependency-free mode is default.
- Runtime-backed mode requires PostgreSQL, DragonflyDB, and Temporal health checks.

## Generated Agent Skills

```bash
pm skills generate --dir docs/skills --json
```

Use generated skills for connector-specific and recipe-specific command guidance.
