# pm connectors inspect assemblyai

```text
NAME
  pm connectors inspect assemblyai - AssemblyAI connector manual

SYNOPSIS
  pm connectors inspect assemblyai
  pm connectors inspect assemblyai --json
  pm credentials add <name> --connector assemblyai [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads AssemblyAI transcripts and per-transcript sentence, paragraph, and subtitle references through the AssemblyAI REST API.

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  No secret authentication is required for this connector.

CONFIGURATION
  No connector-specific config fields.

SECURITY
  read risk: connector-specific
  write risk: connector-specific
  approval: external mutations require preview and approval
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect assemblyai

  # Inspect as structured JSON
  pm connectors inspect assemblyai --json

AGENT WORKFLOW
  - Run pm connectors inspect assemblyai before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
