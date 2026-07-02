```
NAME
  pm rlm - score warehouse records with deterministic or agent RLM

SYNOPSIS
  pm rlm run --spec spec.json --in customers --out scored_customers --mode deterministic [--json]
  pm rlm run --spec spec.json --out scored_customers --mode fixture [--json]
  pm rlm run --spec spec.json --in customers --out scored_customers --mode agent --request "score leads" [--json]

DESCRIPTION
  RLM materializes scored records to the local warehouse. Deterministic and
  fixture modes run dependency-free. Model and agent modes are opt-in and
  runtime-backed.

SECURITY
  RLM output is data only. It does not send messages or mutate external systems.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
