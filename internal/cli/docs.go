package cli

const rootHelp = `NAME
  pm - local-first Polymetrics AI ETL and reverse ETL CLI

SYNOPSIS
  pm <command> [options]

DESCRIPTION
  pm runs a dependency-free local Polymetrics AI MVP. It manages credentials,
  connectors, ETL, reverse ETL plans, local warehouse tables, and agent-safe
  JSON output from one Go binary.

  Every command group is also a manual page. Run pm connectors, pm etl,
  pm credentials, or any other command group without a subcommand to read its
  documentation. Use --json on a command group to return the same manual in a
  machine-readable envelope for agents.

COMMANDS
  init              create a .polymetrics project
  connectors        list and inspect connectors
  credentials       add, test, inspect, list, and remove credentials
  connections       create and list source-to-destination connections
  catalog           refresh or show source catalogs
  etl               run ETL and inspect run status
  query             query local warehouse tables
  reverse           list, plan, preview, run, and inspect reverse ETL
  agent             produce typed plans for external agents
  runtime           check PostgreSQL, DragonflyDB, and Temporal dependencies
  perf              compare dependency-free and runtime-backed performance
  docs              generate markdown command docs
  skills            generate agent SKILL.md files
  version           print build version metadata
  help, man         show detailed documentation

HUMAN QUICK START
  pm init
  pm credentials add sample-local --connector sample
  pm credentials add warehouse-local --connector warehouse
  pm connections create sample_to_warehouse --source sample:sample-local --destination warehouse:warehouse-local --stream customers --primary-key id --cursor updated_at --table sample_customers
  pm etl run --connection sample_to_warehouse --stream customers
  pm query run --table sample_customers --limit 5

AGENT CONTRACT
  Use --json for machine-readable output.
  Use pm <command> to inspect command manuals before executing workflows.
  Use pm connectors inspect <name> --json before selecting connector config.
  Do not ask users for secret values in chat; use --from-env field=ENV or
  --value-stdin field.
  Reverse ETL external mutations require plan, preview, approval, and run.

SECURITY
  Secrets are stored encrypted under .polymetrics/vault. JSON output never
  includes decrypted secret values.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
`

var docs = map[string]string{
	"":            rootHelp,
	"pm":          rootHelp,
	"credentials": credentialsHelp,
	"etl":         etlHelp,
	"reverse":     reverseHelp,
	"connectors":  connectorsHelp,
	"connections": connectionsHelp,
	"catalog":     catalogHelp,
	"query":       queryHelp,
	"agent":       agentHelp,
	"runtime":     runtimeHelp,
	"perf":        perfHelp,
	"docs":        docsHelp,
	"skills":      skillsHelp,
	"version":     versionHelp,
}

const versionHelp = `NAME
  pm version - print build version metadata

SYNOPSIS
  pm version [--json]

DESCRIPTION
  Prints the release version, git commit, and build date embedded into release
  binaries. Development builds print dev, none, and unknown unless overridden
  with Go linker flags.

OPTIONS
  --json    render machine-readable JSON

EXIT STATUS
  0 success
`

const credentialsHelp = `NAME
  pm credentials - manage encrypted connector credentials

SYNOPSIS
  pm credentials add <name> --connector <connector> [--from-env field=ENV] [--value-stdin field] [--config key=value]
  pm credentials list [--json]
  pm credentials inspect <name> [--json]
  pm credentials test <name> [--json]
  pm credentials remove <name>

DESCRIPTION
  Credentials combine non-secret connector config with encrypted secret fields.
  Secrets should be supplied through environment variables or stdin, not shell
  arguments. Use --from-env field=ENV for non-interactive setup. Use
  --value-stdin field for multiline secrets such as GitHub App PEM keys.

OPTIONS
  --connector name       connector that owns the credential
  --from-env field=ENV   read one secret field from an environment variable
  --value-stdin field    read one secret field from standard input
  --config key=value     store non-secret connector config
  --root path            project root containing .polymetrics
  --json                 render machine-readable JSON

SECURITY
  Secret values are encrypted with AES-GCM in .polymetrics/vault and are not
  stored in state.json. Inspection output shows only secret field names.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
`

const connectorsHelp = `NAME
  pm connectors - inspect built-in connector capabilities and native Go catalog

SYNOPSIS
  pm connectors list [--all] [--json]
  pm connectors catalog [--type source|destination] [--stage stage] [--json]
  pm connectors port-plan --all [--json]
  pm connectors port-plan <catalog-slug> [--json]
  pm connectors inspect <name-or-catalog-slug> [--json]
  pm connectors help <name-or-catalog-slug>

DESCRIPTION
  pm ships with built-in runnable connectors and a generated connector catalog.
  Built-in connectors are compiled into the binary and expose explicit runtime
  capabilities. Catalog connectors expose documentation, configuration schema,
  secret field names, sync support, native implementation status, and the Go
  runtime family used by the native binding.

  Catalog entries are native-Go-only. pm does not execute connector container
  images. implementation_status=enabled means the connector has a Go runtime
  binding and fixture-backed conformance coverage. Connector-specific live API
  behavior is documented in each connector manual when available.

CATALOG
  The generated native-Go-only catalog contains 646 validated connectors:
  590 sources and 56 destinations. Use --all or the catalog subcommand when an
  agent needs to discover the complete connector universe.

GITHUB AUTHENTICATION
  public
    Unauthenticated public repository reads. Configure repository=owner/repo.
    This mode cannot execute reverse ETL writes.

  token
    Bearer-token auth for classic PATs, fine-grained PATs, OAuth tokens,
    GitHub Actions GITHUB_TOKEN, or pre-generated installation tokens. Store the
    secret as token, personalAccessToken, oauthToken, accessToken,
    installationToken, or githubToken.

  github_app
    Server-to-server GitHub App auth. Configure auth_type=github_app, app_id,
    and installation_id. Store the app private key with --value-stdin
    private_key or --from-env private_key_base64=ENV. pm signs a short-lived JWT
    and exchanges it for a one-hour installation token.

  unsupported
    Password auth and SSH keys do not authenticate GitHub REST API requests.

GITHUB ETL STREAMS
  issues
    Reads repository issues through /repos/{owner}/{repo}/issues and filters out
    pull requests returned by the Issues API. Primary key: node_id. Cursor:
    updated_at.

  pull_requests
    Reads repository pull requests through /repos/{owner}/{repo}/pulls. Primary
    key: node_id. Cursor: updated_at.

  Pagination defaults to one page. Set --config max_pages=0, all, or unlimited
  to read pages until the GitHub endpoint is exhausted.

GITHUB REVERSE ETL ACTIONS
  The built-in github connector can execute approved reverse ETL write actions
  such as create_issue, create_pull_request, comment_issue, update_issue,
  update_pull_request, request_reviewers, merge_pull_request, labels,
  milestones, releases, workflow runs, pull request reviews, and repository
  file create/update/delete.

ACTIONS
  list
    Prints runnable built-in connectors by default. Use --all to print the full
    generated catalog. Use --json when an agent needs stable structured output.

  catalog
    Prints the generated connector catalog, optionally filtered by --type and
    --stage. Example stages include alpha, beta, and generally_available.

  port-plan
    Prints native Go implementation plans for catalog connectors. Plans include
    runtime family, priority wave, ETL work, reverse ETL boundary, database CDC
    requirements, and conformance tests.

  inspect <name>
    Prints a man-style connector manual for built-in or catalog-only
    connectors. Use --json to print either the structured manifest or catalog
    definition for agents. Inspection is metadata-only and does not resolve
    credentials.

  help <name>
    Alias for the human connector manual.

EXAMPLES
  pm connectors
  pm connectors --json
  pm connectors list
  pm connectors list --all --json
  pm connectors catalog --type destination --stage generally_available --json
  pm connectors port-plan --all --json
  pm connectors port-plan source-postgres
  pm connectors port-plan source-mysql
  pm connectors port-plan source-mongodb-v2
  pm connectors inspect github
  pm connectors inspect source-github
  pm connectors inspect destination-postgres
  pm connectors inspect github --json
  pm credentials add github-public --connector github --config repository=octocat/Hello-World
  pm credentials add github-token --connector github --config repository=OWNER/REPO --from-env token=GITHUB_TOKEN
  pm credentials add github-app --connector github --config repository=OWNER/REPO --config auth_type=github_app --config app_id=12345 --config installation_id=67890 --value-stdin private_key < app.pem

SECURITY
  Connector inspection never reads credentials.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
`

const connectionsHelp = `NAME
  pm connections - configure source-to-destination sync connections

SYNOPSIS
  pm connections create <name> --source connector:credential --destination connector:credential --stream stream [--sync-mode mode] [--cursor field] [--primary-key field] [--table table]
  pm connections list [--json]

DESCRIPTION
  A connection joins one source endpoint to one destination endpoint and stores
  stream-level sync settings.

SYNC MODES
  full_refresh_append              read all source records and append them
  full_refresh_overwrite           read all source records and replace final output
  full_refresh_overwrite_deduped   replace final output and keep latest row per primary key
  incremental_append               append records at or after the saved cursor
  incremental_append_deduped       append raw history and materialize latest row per primary key

  Incremental modes require --cursor. Deduped modes require --primary-key. When
  a connector manifest declares defaults, pm fills them during connection
  creation.

SECURITY
  Connections reference credentials by name only.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
`

const catalogHelp = `NAME
  pm catalog - discover and display source streams

SYNOPSIS
  pm catalog refresh --connection <name> [--json]
  pm catalog show --connection <name> [--json]

DESCRIPTION
  Catalog commands call the source connector and store a local snapshot.

SECURITY
  Catalog output includes schemas and stream names, never secret values.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
`

const etlHelp = `NAME
  pm etl - run local ETL syncs

SYNOPSIS
  pm etl check --connector <name> [--config key=value] [--json]
  pm etl catalog --connector <name> [--config key=value] [--json]
  pm etl read --connector <name> [--stream stream] [--limit n] [--config key=value] [--json]
  pm etl run --connection <name> --stream <stream> [--batch-size n] [--runtime] [--json]
  pm etl status <run-id> [--json]

DESCRIPTION
  ETL can directly check, catalog, and read enabled runtime connectors by
  catalog slug or built-in connector name. Catalog entries with
  implementation_status=planned_native_port are inspectable through
  pm connectors inspect and pm connectors port-plan, but cannot execute ETL
  until a native Go port passes conformance and is enabled.

  ETL runs read from a configured source connector, add Polymetrics metadata
  fields, and write records to the destination connector. The MVP warehouse
  destination stores tables as JSONL files.

  ETL writes destination records in bounded batches. Use --batch-size for large
  paginated streams when you want tighter memory bounds.

  With --runtime, ETL also requires healthy PostgreSQL, DragonflyDB, and Temporal
  endpoints. It acquires a Dragonfly lease and appends a PostgreSQL run-ledger
  record after the local ETL completes.

DIRECT NATIVE CONNECTOR COMMANDS
  check
    Calls the connector check operation and returns status=ok on success.

  catalog
    Calls the connector catalog/discover operation and prints available streams.

  read
    Reads fixture-backed or live records from a connector stream with a hard
    output limit. Use --json for stable agent output.

SOURCE STREAMS
  sample.customers
    Deterministic customer fixture stream. Primary key: id. Cursor: updated_at.

  sample.events
    Deterministic event fixture stream. Primary key: id. Cursor: occurred_at.

  file.file
    Local JSONL or CSV file stream. Configure path and optionally stream.

  github.issues
    Repository issues excluding pull requests. Primary key: node_id. Cursor:
    updated_at. Supports public, token, and github_app auth.

  github.pull_requests
    Repository pull requests. Primary key: node_id. Cursor: updated_at.
    Supports public, token, and github_app auth.

DESTINATIONS
  warehouse
    Local JSONL warehouse tables. Supports append, overwrite, append_dedup, and
    overwrite_dedup destination behavior through ETL sync modes.

SYNC MODES
  full_refresh_append
    Reads every source record and appends to the final JSONL table. Duplicates
    across runs are expected.

  full_refresh_overwrite
    Reads every source record into a temp final file, then atomically replaces
    the final JSONL table only after the run succeeds.

  full_refresh_overwrite_deduped
    Reads every source record, writes current-generation raw JSONL, dedupes by
    primary key and cursor, then atomically replaces the final JSONL table.

  incremental_append
    Reads records at or after the saved cursor and appends accepted records.
    Cursor state advances only after successful writes.

  incremental_append_deduped
    Appends accepted records to raw JSONL history and materializes a final JSONL
    table with one latest row per primary key. Delete/tombstone records remove
    the row from final output.

SECURITY
  ETL resolves credentials in memory and stores only credential references.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
`

const queryHelp = `NAME
  pm query - inspect local warehouse data

SYNOPSIS
  pm query run --table <table> [--limit n] [--json]
  pm query run --sql "select * from <table> limit n" [--json]

DESCRIPTION
  The MVP query engine supports table reads and a small SELECT * FROM parser.

SECURITY
  Query output can contain data rows. Agent callers should use small limits.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
`

const reverseHelp = `NAME
  pm reverse - plan, preview, approve, and execute reverse ETL

SYNOPSIS
  pm reverse <command> [flags]

USAGE
  pm reverse list [--json]
  pm reverse plan <name> --source-table <table> --destination connector:credential --map source:dest [--json]
  pm reverse preview <plan-id> [--json]
  pm reverse run <plan-id> --approve <token> [--json]
  pm reverse status <run-id> [--json]

DESCRIPTION
  Reverse ETL reads local warehouse rows, maps fields, and writes records to a
  destination connector. The outbox connector records writes as JSONL. The
  github connector can execute approved issue and pull request mutations.

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
    token from the human plan output.

  status
    Show a completed or failed reverse ETL run by run ID.

FLAGS
  --source-table table         local warehouse table to read
  --destination connector:cred destination endpoint
  --map source:dest            field mapping, repeatable
  --action action              destination action; GitHub writes require an explicit action
  --limit n                    maximum source rows to include in the plan
  --approve token              approval token required by run
  --json                       render machine-readable JSON
  --root path                  project root containing .polymetrics

GITHUB ACTIONS
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
  Run pm connectors inspect github --json to inspect GitHub write actions.
  Run pm skills generate --dir docs/skills --json for agent-specific workflows.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
`

const agentHelp = `NAME
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
`

const docsHelp = `NAME
  pm docs - generate CLI documentation

SYNOPSIS
  pm docs generate --dir <path>
  pm docs validate [--connectors-dir <path>]

DESCRIPTION
  Writes embedded command documentation as markdown files. Generation also
  writes connector manuals under a connector docs directory. By default, when
  --dir is docs/cli, connector docs are written to docs/connectors.

  Validation checks every registered connector has a generated MANUAL.md with
  required human and agent workflow sections. This is intended for CI hooks and
  local preflight checks before adding or changing connectors.

OPTIONS
  --dir path             command docs output directory
  --connectors-dir path  connector docs output directory

SECURITY
  Generated docs contain no credentials.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
`

const skillsHelp = `NAME
  pm skills - generate agent skills

SYNOPSIS
  pm skills generate --dir <path> [--json]

DESCRIPTION
  Generates Codex/Claude-compatible SKILL.md files from the current CLI and
  connector manifests. Generated skills describe safe commands, connector
  streams, secret field names, and approval boundaries. Secret values are never
  read from the vault or written to generated files.

OPTIONS
  --dir path     destination directory for generated skills
  --json         render machine-readable generation summary

SECURITY
  Skill generation is metadata-only. It does not resolve credentials, read
  encrypted secret values, or contact external APIs.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
`

const runtimeHelp = `NAME
  pm runtime - inspect external runtime dependencies

SYNOPSIS
  pm runtime doctor [--json]

DESCRIPTION
  Checks PostgreSQL, DragonflyDB, and Temporal using the configured endpoints.
  Defaults match the local Compose stack in deploy/compose.

ENVIRONMENT
  POLYMETRICS_POSTGRES_URL
  POLYMETRICS_DRAGONFLY_ADDR
  POLYMETRICS_TEMPORAL_ADDR

SECURITY
  PostgreSQL passwords are redacted in command output.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
`

const perfHelp = `NAME
  pm perf - compare dependency-free and dependency-backed runtime paths

SYNOPSIS
  pm perf compare [--iterations n] [--runtime] [--json]
  pm perf sync-modes [--records n] [--json]

DESCRIPTION
  Runs repeated local ETL loops and reports elapsed time, average operation time,
  and records per second. Without --runtime, only the dependency-free path runs.
  With --runtime, the command also checks PostgreSQL, DragonflyDB, and Temporal,
  acquires a Dragonfly lease, appends a PostgreSQL ledger record, and compares
  that path against the dependency-free baseline.

  The sync-modes subcommand runs a synthetic local file-to-warehouse benchmark
  for every supported ETL sync mode and reports each mode's duration and records
  per second.

SECURITY
  Performance output contains counts and durations only.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
`
