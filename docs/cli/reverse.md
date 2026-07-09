```
NAME
  pm reverse - plan, preview, approve, and execute reverse ETL

SYNOPSIS
  pm reverse <command> [flags]

USAGE
  pm reverse list [--json]
  pm reverse plan <name> --source-table <table> --destination connector:credential --map source:dest [--json]
  pm reverse preview <plan-id> [--json]
  pm reverse run <plan-id> --approve <token> [--confirm <challenge>] [--json]
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

COMMANDS
  list
    List reverse ETL plans and runs in the current project.

  plan
    Create a reverse ETL plan from a local warehouse table to a destination
    connector. A human-readable plan prints an approval token for the user.
    JSON output redacts the token.

  preview
    Show a stored plan, mapped sample rows, destination connector, action, and
    record count before execution.

  run
    Execute a stored plan only when --approve is supplied with the approval
    token from the human plan output. Destructive or sensitive plans can also
    require the typed --confirm challenge printed by the plan output.

  status
    Show a completed or failed reverse ETL run by run ID.

FLAGS
  --source-table table         local warehouse table to read
  --destination connector:cred destination endpoint
  --map source:dest            field mapping, repeatable
  --action action              destination write action; inspect shows names
  --limit n                    maximum source rows to include in the plan
  --approve token              approval token required by run
  --confirm challenge          typed confirmation required by gated plans
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
  pm reverse run rplan_abc123 --approve <approval-token>
  pm reverse status rrun_abc123 --json

SECURITY
  Execution requires an approval token created by a prior plan. JSON plan output
  omits the token so agents cannot silently self-approve external writes.
  Reverse ETL never exposes raw secret values.

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
