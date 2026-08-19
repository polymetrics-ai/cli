```
NAME
  pm worker - operate the optional RLM worker

SYNOPSIS
  pm worker serve [--json]
  pm worker status [--json]

DESCRIPTION
  Starts or inspects the optional Temporal-backed worker used by RLM agent
  mode. Runtime services are opt-in; use pm runtime doctor before starting a
  runtime-backed workflow.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
