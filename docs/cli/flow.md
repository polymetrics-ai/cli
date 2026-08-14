```
NAME
  pm flow - plan, preview, run, list, and inspect multi-step flows

SYNOPSIS
  pm flow plan --file flow.json [--json]
  pm flow preview --file flow.json [--json]
  pm flow run --file flow.json [--force] [--json]
  pm flow status <name> [--flows-dir .polymetrics/flows] [--json]
  pm flow list [--flows-dir .polymetrics/flows] [--json]

DESCRIPTION
  Flow manifests compose sync, query, rlm, and action steps. Dependencies are
  inferred from in/out warehouse tables. RLM steps reuse pm rlm analyzers and
  may reference a spec path relative to the flow manifest file.

CONNECTION-SCOPED SOURCE READS
  A query step may set "connection" to scope every warehouse table view used
  by its SQL. An action step sets "source_connection" inside "action_cfg" to
  scope its "source_table". Use _unattributed only for a root-level table that
  no connection owns. When same-named tables have several owners, omitting the
  applicable manifest selector refuses the read instead of choosing one.
  A case-equivalent spelling whose owner cannot be decided also fails closed;
  set "connection" to a known healthy owner rather than relying on an
  unscoped query.

  Query example:
  {"id":"query-acme","kind":"query","connection":"acme",
   "sql":"SELECT * FROM records","in":[],"out":[]}

  Action source selector fragment:
  "action_cfg": {"source_table":"records","source_connection":"acme"}

RLM STEP EXAMPLE
  {
    "id": "score",
    "kind": "rlm",
    "spec": "lead-score.json",
    "mode": "fixture",
    "in": [],
    "out": ["lead_scores"]
  }

SECURITY
  Read-only sync, query, and rlm steps run through existing app primitives.
  Action steps remain approval-gated.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
