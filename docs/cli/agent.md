```
NAME
  pm agent - produce typed command plans for external LLM agents

SYNOPSIS
  pm agent plan --request <text> [--json]

DESCRIPTION
  Agent planning is intentionally narrow in the MVP. It returns typed command
  suggestions and safety notes instead of executing arbitrary instructions.

SECURITY
  The agent command cannot read secrets, generate approval tokens, or run shell
  commands.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
