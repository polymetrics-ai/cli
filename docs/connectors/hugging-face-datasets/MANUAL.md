# pm connectors inspect hugging-face-datasets

```text
NAME
  pm connectors inspect hugging-face-datasets - Hugging Face - Datasets connector manual

SYNOPSIS
  pm connectors inspect hugging-face-datasets
  pm connectors inspect hugging-face-datasets --json
  pm credentials add <name> --connector hugging-face-datasets [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads dataset splits, per-split sizes, and rows from the Hugging Face dataset-viewer REST API. Read-only; an optional user access token unlocks gated and private datasets.

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
  pm connectors inspect hugging-face-datasets

  # Inspect as structured JSON
  pm connectors inspect hugging-face-datasets --json

AGENT WORKFLOW
  - Run pm connectors inspect hugging-face-datasets before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
