# SPEC: Agentic ETL Platform

## User Stories

- As an agent, I can inspect connector capabilities and schemas without accessing secrets.
- As an agent, I can run ETL against many paginated pages without exhausting memory.
- As a human, I can generate detailed CLI docs and agent skills from the current binary.
- As an operator, I can diagnose CLI failures by stable error category and exit code.
- As a security reviewer, I can verify that unsafe terminal/control characters and path traversal are rejected.

## Functional Requirements

- Add stable CLI error categories: usage, validation, auth, connector, runtime, policy, internal.
- Sanitize untrusted strings before printing to terminal-facing output.
- Validate identifiers and paths at parse boundaries.
- Add connector manifest fields for config, secrets, streams, pagination, sync modes, and risk.
- Generate `SKILL.md` files for shared, connector, ETL, reverse ETL, runtime, and GitHub recipes.
- Stream ETL through bounded destination writes.
- Persist ETL checkpoint/progress metadata on run completion and failure.

## Constraints

- Single Go binary.
- Default mode remains dependency-free.
- No new Go module dependencies.
- Existing commands remain backward compatible where possible.
