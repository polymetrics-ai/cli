```
NAME
  pm flow - create and run job-backed multi-step flows

SYNOPSIS
  pm flow create --file flow.json [--json]
  pm flow plan --file flow.json [--json]
  pm flow preview --file flow.json [--json]
  pm flow run <name> [--force] [--json]
  pm flow run --file flow.json [--force] [--json]
  pm flow status <name> [--flows-dir .polymetrics/flows] [--json]
  pm flow list [--flows-dir .polymetrics/flows] [--json]

DESCRIPTION
  Flow manifests compose sync, query, rlm, and action steps. Dependencies are
  inferred from in/out warehouse tables. RLM steps reuse pm rlm analyzers and
  may reference a spec path relative to the flow manifest file.

  Create stores a flow only after every external job reference resolves
  positively. A sync step's job is an existing ETL connection. An action
  step's job is an existing reverse-ETL plan that has completed its one-time
  plan → preview → approval → execute lifecycle. Missing, malformed,
  unrecognised, or unapproved jobs are refused before the flow file is written.

CONNECTION-SCOPED SOURCE READS
  A query step may set "connection" to scope every warehouse table view used
  by its SQL. A sync or action step instead names its existing job. The action
  source connection, source table, mappings, destination action, credential,
  and confirmation policy are derived from the approved reverse-ETL job; they
  cannot be supplied inline. Use _unattributed only for a root-level table that
  no connection owns. When same-named tables have several owners, omitting the
  applicable selector refuses the read instead of choosing one.
  A case-equivalent spelling whose owner cannot be decided also fails closed;
  set "connection" to a known healthy owner rather than relying on an
  unscoped query.

  Query example:
  {"id":"query-acme","kind":"query","connection":"acme",
   "sql":"SELECT * FROM records","in":[],"out":[]}

  Approved action job fragment:
  {"id":"send","kind":"action","job":"rplan_0123456789abcdef",
   "action_cfg":{"read_back_stream":"targets"}}

ACTION EXECUTION
  An action uses the selected warehouse rows and the destination connector's
  typed ValidateWrite and Write methods; it never accepts a raw URL, generic
  HTTP write, SQL write, or operation request. Approve the reverse-ETL job once
  at connection, schema, preview, mapping, destination action, credential
  revision, and confirmation-policy granularity; then reference that job from
  the flow. No approval token or authorization reference is accepted by flow
  create or run.

  Every run reloads the job and revalidates that standing authorization before
  any provider request. It derives a payload-bound prepared-execution identity,
  validates the target, writes once, reads the target stream back, and persists
  the safe identity and opaque receipt before the action checkpoint succeeds.

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
  Action steps inherit their job's durable, revocable standing authorization.
  Credential revision, manifest/schema, source scope, mappings, destination
  action, confirmation policy, expiry, and revocation drift stop before write.
  Prepared identities are evidence, not secrets or reusable authority.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
