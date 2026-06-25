```
NAME
  pm skills - generate agent skills

SYNOPSIS
  pm skills generate --dir <path> [--json]

DESCRIPTION
  Generates Codex/Claude-compatible SKILL.md files from the current CLI and
  connector manifests. Generated skills describe safe commands, connector
  streams, secret field names, and approval boundaries. Secret values are never
  read from the vault or written to generated files.

OPTIONS
  --dir path     destination directory for generated skills
  --json         render machine-readable generation summary

SECURITY
  Skill generation is metadata-only. It does not resolve credentials, read
  encrypted secret values, or contact external APIs.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
