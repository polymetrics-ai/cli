```
NAME
  pm extract - classify a narrow natural-language read request

SYNOPSIS
  pm extract --request <text> [--sql <select>] [--limit <n>] [--json]
  pm extract --request <text> --in <table> --out <table> [--spec-name <name>] [--json]

DESCRIPTION
  extract is a hidden, narrow natural-language router for agents. It classifies
  a request as a simple read-only query or typed RLM analysis. Routing remains
  dependency-free when no optional LLM provider is configured.

  A simple-query route runs only validated read-only SQL supplied by --sql or
  suggested by the optional classifier. An RLM route executes only when both
  --in and --out are provided; otherwise it returns the routing decision and a
  note. Input and output values are bare warehouse table names, not paths.

OPTIONS
  --request <text>
    Natural-language request to classify. Required for execution.
  --sql <select>
    Explicit read-only SQL for a simple-query route.
  --limit <n>
    Maximum query rows. Default: 100.
  --in <table>
    Bare local warehouse input table for RLM analysis.
  --out <table>
    Bare local warehouse output table for RLM analysis.
  --spec-name <name>
    Result specification name. Default: extract.
  --provider <name>, --model <name>, --llm-base-url <url>
    Optional classifier overrides. Without a resolvable provider/model, extract
    uses its offline heuristic.

OUTPUT
  --json emits one ExtractResult envelope. Human output contains the same
  result as indented JSON. Diagnostics and errors remain on stderr.

SECURITY
  extract does not expose generic shell, HTTP write, SQL write, or unrestricted
  filesystem access. SQL remains read-only. RLM table reads and writes stay
  beneath <root>/.polymetrics/warehouse and reject path traversal and external
  final-link effects. Optional agent execution remains typed and runtime-gated.

EXIT STATUS
  0 success or contextual help
  1 runtime error
  2 usage error
  3 validation error

```
