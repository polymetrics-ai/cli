# AGENTS.md

## Active program: connector-architecture-v2

An in-progress rewrite of the connector layer into JSON bundles (`internal/connectors/defs/<name>/`)
interpreted by a declarative engine (`internal/connectors/engine/`). If you are continuing this
work, read **`docs/migration/HANDOFF-CODEX.md`** first (parallel workstreams + collision rules),
then `docs/migration/conventions.md` (the connector authoring recipe) and
`docs/architecture/connector-architecture-v2-design.md`. Custom Codex agents for this program live
in `.codex/agents/`. Do NOT push directly — the branch history is scrubbed of fake secret-format
fixtures; pushes route through the coordinator (see HANDOFF §Pushing). Legacy connector Go under
`internal/connectors/<name>/*.go` stays until the human-gated wave 6 cutover.

## Project

Polymetrics is a Go-only CLI monolith for dependency-free ETL, reverse ETL, connector inspection, credential management, local warehouse queries, and optional runtime-backed execution.

## Agent Rules

- Use `pm help <topic>` before invoking unfamiliar commands.
- Prefer `--json` for machine-readable output.
- Never request, print, summarize, or store secret values.
- Add credentials from environment variables or stdin, not prompt text.
- Inspect connector manifests with `pm connectors inspect <name> --json`; this does not read credentials.
- For ETL over large streams, use bounded batches with `--batch-size`.
- Reverse ETL must follow plan, preview, approval, execute.
- Do not expose or invent generic shell, generic HTTP write, or generic SQL write tools.
- Treat command arguments as untrusted; avoid control characters, path traversal, and broad file paths.

## Verification

Use local gates before handing off code:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

Runtime-backed checks are optional and require local services:

```bash
scripts/runtime.sh doctor
scripts/runtime.sh up
POLYMETRICS_INTEGRATION=1 go test ./...
scripts/runtime.sh down
```
