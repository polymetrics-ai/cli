```
NAME
  pm reverse - plan, preview, approve, and execute reverse ETL

SYNOPSIS
  pm reverse <command> [flags]

USAGE
  pm reverse list [--json]
  pm reverse plan <name> --source-table <table> [--connection name] --destination connector:credential --map source:dest [--json]
  pm reverse preview <plan-id> [--<withheld-flag> <value>...]
    [--from-env <env-only-flag>=ENV]... [--json]
  pm reverse run <plan-id> --approval-token-stdin [--confirm <challenge>]
    [--<withheld-flag> <value>...] [--from-env <env-only-flag>=ENV]... [--json]
  pm reverse status <run-id> [--json]

DESCRIPTION
  Reverse ETL reads local warehouse rows, maps fields, and writes records
  through a connector write action. It is available for any connector that
  declares capabilities.write=true. Use pm connectors catalog --capability
  write --json to discover writable connectors. The remaining connectors are
  read-only because their APIs expose no supported mutations.

  Run pm connectors inspect <name> to see write=true/false, available ETL
  streams, and reverse ETL write actions for a connector. The outbox connector
  records writes as JSONL. GitHub is one example of an external API connector
  with approved mutation actions.

  The workflow is intentionally split into plan, preview, approval, and run.
  Agents can create and preview plans, but JSON plan output omits approval
  tokens so an agent cannot silently approve its own external mutation.
  The connector command runner does not mask ETL or reverse-ETL command records
  from declared redact_fields; it dispatches the values it was given.
  This runner policy does not change source-table output or other execution paths.

  A connector-command plan does not persist the fields declared sensitive by the
  write action it runs (writes.json redact_fields) or, for a direct_write
  operation, by that operation (operations.json sensitive_policy.redact_fields).
  A redact_fields list on the command itself is not consulted, so a command-level
  declaration withholds nothing; pm connectors inspect <name> --json shows every
  declaration, not only the binding one.
  Withheld keys are removed outright rather than stored as a placeholder, so they
  never reach the project state file. Preview and run therefore need those values
  re-supplied on the same command. Ordinary withheld fields use the connector
  command's own --<flag> <value> form. A field declared env_only must instead
  use --from-env <flag>=ENV, which reads the value from the named environment
  variable without placing it in argv. For example: pm reverse preview <plan-id>
  --from-env input=ENV, and use the same form again on pm reverse run.
  Where a declared field covers a subtree that several flags fill, those flags
  re-supply it. Only fields the plan actually removed are asked for, so a
  declared field you never supplied is never demanded back. A re-supplied value
  that does not match the one the plan was built from fails the plan-hash check
  before anything is dispatched. Nothing is re-persisted at preview or run.
  DryRunWrite engine preview warnings preserve the resolved execution request.
  Engine direct-read, operation-direct-read, and binary-download executors
  preserve bounded HTTP URL/query/body diagnostics before downstream rendering.
  Declared redact_fields remain compatible metadata, but do not replace values
  in DryRunWrite preview warnings. A stored source-table sample is an app-level
  summary; the engine preview is authoritative for approval.

  Destructive plans do not receive an approval token during planning. Preview
  performs the connector's no-network dry run, persists a digest of the complete
  staged request set and its execution identity, and only then issues a
  time-bounded token in human-readable output. Execution recomputes that digest
  before dispatch and also requires the closed typed confirmation --confirm
  destructive. HTTP DELETE is treated as destructive even when connector
  metadata omits a confirmation declaration.

  A connector may declare a write action non-batchable (batchable: false).
  Bulk plans over --source-table refuse such an action, naming the action and
  the individual pm command that still runs it. Those actions stay fully
  available one record at a time as pm <connector> <command>, which keeps the
  plan, preview, approval, and execute steps. Use it for operations that must
  never be fanned out over many rows under a single approval. It is separate
  from --confirm: batchable controls whether an action may run in bulk at all,
  --confirm controls how severe one call is.

COMMANDS
  list
    List reverse ETL plans and runs in the current project.

  plan
    Create a reverse ETL plan from a local warehouse table to a destination
    connector. A human-readable non-destructive plan prints an approval token;
    a destructive plan prints no token until preview succeeds. JSON output
    always omits tokens. A non-batchable destination action is refused here,
    before any plan or approval token exists.

    Each connection materializes its tables into its own directory, so several
    connections can hold a table of the same name. Pass --connection when they
    do. The connection is resolved once, here, and recorded on the plan, so
    preview and run keep reading the same table afterwards; neither takes a
    connection selector of its own. Use --connection _unattributed for a
    root-level table that no connection owns.

  preview
    Show a stored plan's mapped sample rows, action, and count. For a destructive
    plan, also materialize the request through the destination's no-network dry
    run, persist its digest, and issue the approval token in human-readable
    output. JSON omits the token. DryRunWrite engine preview warnings preserve
    the resolved execution request, including fields declared in redact_fields;
    that preview is what the digest binds before dispatch. A connector-command
    plan that withheld declared sensitive fields needs them re-supplied here:
    --from-env <flag>=ENV for an env_only field, or --<flag> <value> otherwise.
    The error names each missing flag.

  run
    Execute a stored plan only when the bare --approval-token-stdin marker is
    supplied and standard input contains the approval token as one bounded line
    from human-readable plan or preview output. Destructive plans require a
    matching persisted preview and the closed --confirm destructive value. A
    connector-command plan that withheld declared sensitive fields needs the
    same re-supply form: --from-env <flag>=ENV for an env_only field, or
    --<flag> <value> otherwise. A failed dispatch is recorded; pm does not
    automatically retry a failed dispatch.

  status
    Show a completed or failed reverse ETL run by run ID.

FLAGS
  --source-table table         local warehouse table to read
  --connection name            connection whose table to read; required only
                               when several connections share the table name
  --destination connector:cred destination endpoint
  --map source:dest            field mapping, repeatable
  --action action              destination write action; inspect shows names
  --limit n                    maximum source rows to include in the plan
  --approval-token-stdin       read the approval token as one bounded line from
                               standard input; the marker accepts no value
  --confirm challenge          typed confirmation required by gated plans
  --<withheld-flag> value      re-supply a non-env_only field the plan withheld;
                               the flag is connector-owned, never persisted
  --from-env flag=ENV          re-supply a declared env_only field from ENV;
                               its value never enters argv or project state
  --json                       render machine-readable JSON
  --root path                  project root containing .polymetrics

GITHUB ACTION EXAMPLES
  These are examples from one writable connector. Other connectors expose
  different actions; pm connectors inspect <name> is the authoritative list.

  create_issue
    Requires title. Optional body, labels, assignees, milestone, type.

  update_issue
    Requires issue_number or number. Optional title, body, state,
    state_reason, labels, assignees, milestone, type.

  comment_issue
    Requires issue_number, pull_number, or number plus body. Alias: comment_pr.

  create_pull_request
    Requires title, head, and base. Optional body, draft,
    maintainer_can_modify, labels, assignees, milestone, reviewers,
    team_reviewers. Aliases: create_pr, pr_create.

  update_pull_request
    Requires pull_number or number. Optional title, body, state, base,
    maintainer_can_modify, labels, assignees, milestone, reviewers,
    team_reviewers. Alias: update_pr.

  request_reviewers
    Requires pull_number or number plus reviewers or team_reviewers.

  merge_pull_request
    Requires pull_number or number. Optional commit_title, commit_message, sha,
    merge_method. Alias: merge_pr.

EXAMPLES
  pm reverse
  pm reverse list
  pm reverse plan customers_to_outbox --source-table sample_customers --destination outbox:outbox-local --map id:external_id --map email:email
  pm reverse plan prs_to_github --source-table github_pr_candidates --destination github:github-local --action create_pull_request --map title:title --map head:head --map base:base --map reviewers:reviewers
  pm reverse preview rplan_abc123 --json
  pm reverse run rplan_abc123 --approval-token-stdin
  pm reverse status rrun_abc123 --json

SECURITY
  Execution requires a time-bounded, single-use approval token on standard
  input. Destructive tokens are created only after preview; execution revalidates
  the preview digest before dispatch. JSON plan and preview output omit tokens
  so agents cannot silently self-approve external writes. The stdin carrier is
  one bounded line and the token is never accepted through command arguments,
  environment, or project files. A connector-command plan never
  persists the fields its write action declares in redact_fields, or its
  direct_write operation declares in sensitive_policy.redact_fields; they are
  re-supplied per invocation and are not written back at preview or run. A
  redact_fields list declared on the command itself is not a withholding
  guarantee and never has been. DryRunWrite engine
  preview warnings preserve the resolved execution request, including fields
  declared in redact_fields. Engine direct-read, operation-direct-read, and binary-
  download executors preserve bounded HTTP URL/query/body diagnostics before
  downstream rendering. These engine-level guarantees do not establish
  complete pm CLI output. Credential storage remains encrypted at rest.

LEARN MORE
  Run pm reverse --help for this manual.
  Run pm connectors inspect outbox --json to inspect the local outbox destination.
  Run pm connectors inspect <name> --json to inspect streams and write actions.
  Run pm connectors inspect github --json to inspect one connector's write actions.
  Run pm skills generate --dir docs/skills --json for agent-specific workflows.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
