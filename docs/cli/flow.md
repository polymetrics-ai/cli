```
NAME
  pm flow - plan, preview, run, list, and inspect multi-step flows

SYNOPSIS
  pm flow plan --file flow.json [--json]
  pm flow preview --file flow.json [--json]
  pm flow run --file flow.json [--force] [--json] [--progress ndjson]
  pm flow status <name> [--flows-dir .polymetrics/flows] [--json]
  pm flow list [--flows-dir .polymetrics/flows] [--json]

DESCRIPTION
  Flow manifests compose sync, query, rlm, and action steps. Dependencies are
  inferred from in/out warehouse tables. RLM steps reuse pm rlm analyzers and
  may reference a spec path relative to the flow manifest file.

PROGRESS
  On an eligible stdin+stdout TTY, pm flow run renders an inline pipeline-rail
  dashboard and leaves a truthful final frame in scrollback. Add --progress
  ndjson to stream the same sanitized flow progress events to stderr for agents.
  Stdout remains the dashboard/final human line or single JSON envelope. On
  failures, stderr may also include the final error diagnostic after progress
  events. CI, PM_NO_TUI, --plain, --json, --no-input, pipes, and TERM=dumb keep
  the plain path.

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
  3 validation error, including invalid UI/progress flag

```
