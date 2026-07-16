package cli

const rootHelp = `NAME
  pm - local-first Polymetrics AI ETL and reverse ETL CLI

SYNOPSIS
  pm <command> [options]

DESCRIPTION
  pm runs a dependency-free local Polymetrics AI MVP. It manages credentials,
  connectors, ETL, reverse ETL plans, local warehouse tables, and agent-safe
  JSON output from one Go binary.

  Connectors expose ETL read streams across the catalog. Connectors whose APIs
  support mutations also expose approval-gated reverse ETL write actions. Use
  pm connectors inspect <name> to see a connector's streams, write=true/false,
  and write actions.

  Every command group is also a manual page. Run pm connectors, pm etl,
  pm credentials, or any other command group without a subcommand to read its
  documentation. Use --json on a command group to return the same manual in a
  machine-readable envelope for agents.

COMMANDS
  init              create a .polymetrics project
  connectors        list and inspect connector streams and write actions
  credentials       add, test, inspect, list, and remove credentials
  connections       create and list source-to-destination connections
  catalog           refresh or show source catalogs
  etl               run ETL stream reads and inspect run status
  query             query local warehouse tables
  reverse           list, plan, preview, run, and inspect reverse ETL writes
  flow              plan, preview, run, list, and inspect multi-step flows
  rlm               score warehouse records with deterministic or agent RLM
  schedule          create, list, install, and remove flow schedules
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
  Use pm connectors inspect <name> --json before selecting connector config,
  streams, or write actions.
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
	"flow":        flowHelp,
	"config":      configHelp,
	"rlm":         rlmHelp,
	"schedule":    scheduleHelp,
	"agent":       agentHelp,
	"runtime":     runtimeHelp,
	"perf":        perfHelp,
	"docs":        docsHelp,
	"skills":      skillsHelp,
	"version":     versionHelp,
}

const configHelp = `NAME
  pm help config - configuration reference

SYNOPSIS
  pm help config
  pm <command> --root <path> [--json]

DESCRIPTION
  pm resolves typed invocation configuration once per CLI run. The loader uses a
  fresh Viper instance for each invocation and never uses the package-level Viper
  singleton, AutomaticEnv, or file watching. Current command behavior continues
  to use legacy readers until the env-migration phase, but malformed
  .polymetrics/config.yaml files already fail as validation errors.

PRECEDENCE
  1. Bound global flags: --root and --json.
  2. Explicit POLYMETRICS_* environment variables.
  3. Documented PM_* legacy aliases when the primary POLYMETRICS_* variable is
     not set.
  4. .polymetrics/config.yaml under the invocation project root.
  5. Built-in defaults.

CONFIG FILE
  The config file path is <project-root>/.polymetrics/config.yaml. Missing files
  are allowed. The root key in a config file does not relocate config-file
  discovery for the same invocation; use --root or POLYMETRICS_ROOT to select a
  different project root before the file is read.

  Example:

    version: 1
    project: polymetrics-local
    warehouse:
      connector: warehouse
      path: .polymetrics/warehouse
    runtime:
      postgres_url: postgres://localhost:15433/polymetrics?sslmode=disable
      dragonfly_addr: localhost:6379
      temporal_addr: localhost:7233
    rlm:
      image: ghcr.io/polymetrics/rlm-agent:latest
      podman_bin: podman
      fake_runner: false
      embedded_worker: false
      llm:
        provider: openrouter
        base_url: https://openrouter.ai/api/v1
        model: ""
    schedule:
      crontab_file: ""

KEYS
  root
    Default: invocation root (.). Primary env: POLYMETRICS_ROOT. Alias: PM_ROOT.
    Flag: --root.

  json
    Default: false. Primary env: POLYMETRICS_JSON. Alias: PM_JSON. Flag: --json.

  version
    Default: 1. Primary env: POLYMETRICS_VERSION. Alias: PM_VERSION.

  project
    Default: polymetrics-local. Primary env: POLYMETRICS_PROJECT. Alias:
    PM_PROJECT.

  warehouse.connector
    Default: warehouse. Primary env: POLYMETRICS_WAREHOUSE_CONNECTOR. Alias:
    PM_WAREHOUSE_CONNECTOR.

  warehouse.path
    Default: .polymetrics/warehouse. Primary env: POLYMETRICS_WAREHOUSE_PATH.
    Alias: PM_WAREHOUSE_PATH.

  runtime.postgres_url
    Default: local Compose PostgreSQL DSN. Primary env: POLYMETRICS_POSTGRES_URL.
    Alias: PM_POSTGRES_URL. Command output redacts PostgreSQL userinfo.

  runtime.dragonfly_addr
    Default: localhost:6379. Primary env: POLYMETRICS_DRAGONFLY_ADDR. Alias:
    PM_DRAGONFLY_ADDR.

  runtime.temporal_addr
    Default: localhost:7233. Primary env: POLYMETRICS_TEMPORAL_ADDR. Alias:
    PM_TEMPORAL_ADDR.

  rlm.image
    Default: ghcr.io/polymetrics/rlm-agent:latest. Primary env:
    POLYMETRICS_RLM_IMAGE. Alias: PM_RLM_IMAGE.

  rlm.podman_bin
    Default: podman. Primary env: POLYMETRICS_PODMAN_BIN. Alias:
    PM_PODMAN_BIN.

  rlm.fake_runner
    Default: false. Primary env: POLYMETRICS_RLM_FAKE_RUNNER. Alias:
    PM_RLM_FAKE_RUNNER.

  rlm.embedded_worker
    Default: false. Primary env: POLYMETRICS_RLM_EMBEDDED_WORKER. Alias:
    PM_RLM_EMBEDDED_WORKER.

  rlm.llm.provider
    Default: openrouter. Primary env: POLYMETRICS_LLM_PROVIDER. Alias:
    PM_LLM_PROVIDER.

  rlm.llm.base_url
    Default: https://openrouter.ai/api/v1. Primary env: POLYMETRICS_LLM_BASE_URL.
    Alias: PM_LLM_BASE_URL.

  rlm.llm.model
    Default: empty. Primary env: POLYMETRICS_LLM_MODEL. Alias: PM_LLM_MODEL.

  schedule.crontab_file
    Default: empty. Primary env: POLYMETRICS_CRONTAB_FILE. Alias:
    PM_CRONTAB_FILE. Intended for local scheduler redirection and tests.

SECURITY
  Configuration is an allowlist. pm does not ingest arbitrary POLYMETRICS_* or
  PM_* variables. User-named credential env vars supplied to --from-env and
  connector certification credsfile entries are credential data, not app config.
  Do not store secret values in config.yaml or examples. LLM API keys such as
  PM_LLM_API_KEY and provider-specific keys remain environment-only secret
  inputs and are not documented with values.

EXIT STATUS
  0 success
  3 malformed config validation error
`

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
  pm connectors - inspect connector definitions, streams, and write actions

SYNOPSIS
  pm connectors list [--all] [--json]
  pm connectors catalog [--capability read|write|cdc|query] [--stage stage] [--json]
  pm connectors inspect <name> [--json]
  pm connectors help <name>

DESCRIPTION
  pm ships with runnable connector definitions compiled into the binary. Most
  connectors are declarative JSON bundles interpreted by the connector engine;
  hooks or native components cover APIs and protocols that need custom behavior.

  Each connector exposes ETL read streams. Connectors whose APIs expose
  mutation endpoints also declare reverse ETL write actions. Run
  pm connectors inspect <name> to see write=true/false, ETL STREAMS, and
  REVERSE ETL ACTIONS without reading credentials.

  The catalog command is generated from declarative bundles and Tier-3 native
  connectors. pm does not execute connector container images or accept legacy
  source-/destination-prefixed names.

CATALOG
  The connector catalog is generated from local connector metadata. The current
  runtime catalog has 551 bare-name entries: 547 declarative bundles plus the
  local sample, file, warehouse, and outbox primitives. Use --all or the catalog
  subcommand when an agent needs to discover the complete connector universe.
  Use --capability read, write, cdc, or query to filter by executable surface.

GITHUB AUTHENTICATION
  public
    Unauthenticated public repository reads. Configure owner and repo.
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

REVERSE ETL WRITE ACTIONS
  Reverse ETL writes are available for any connector whose API exposes
  mutations and whose definition declares write actions. They are not
  GitHub-only. Use pm connectors catalog --capability write --json to discover
  writable connectors; the rest are read-only because their APIs expose no
  supported mutations.

  Run pm connectors inspect <name> to see a connector's write=true/false
  capability, ETL streams, reverse ETL write actions, required fields, and risk
  notes.

  GitHub is one writable connector example. It supports approved write actions
  such as create_issue, create_pull_request, comment_issue, update_issue,
  update_pull_request, request_reviewers, merge_pull_request, labels,
  milestones, releases, workflow runs, pull request reviews, and repository
  file create/update/delete.

ACTIONS
  list
    Prints runnable connectors by default. Use --all to print the full
    connector catalog. Use --json when an agent needs stable structured output.

  catalog
    Prints connector catalog metadata, optionally filtered by --capability and
    --stage. Example stages include alpha, beta, and generally_available.

  inspect <name>
    Prints a man-style connector manual for a bare connector name. Use --json
    to print structured metadata for agents. Inspection is metadata-only and
    does not resolve credentials.

  help <name>
    Alias for the human connector manual.

EXAMPLES
  pm connectors
  pm connectors --json
  pm connectors list
  pm connectors list --all --json
  pm connectors catalog --capability write --stage generally_available --json
  pm connectors inspect github
  pm connectors inspect github --json
  pm credentials add github-public --connector github --config owner=octocat --config repo=Hello-World --config auth_type=public
  pm credentials add github-token --connector github --config owner=OWNER --config repo=REPO --config auth_type=token --from-env token=GITHUB_TOKEN
  pm credentials add github-app --connector github --config owner=OWNER --config repo=REPO --config auth_type=github_app --config app_id=12345 --config installation_id=67890 --value-stdin private_key < app.pem

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
  ETL can directly check, catalog, and read enabled connectors by name. The
  read surface comes from connector definitions: declarative JSON bundles
  interpreted by the connector engine, with hooks or native components where an
  API or protocol needs custom behavior. Use pm connectors inspect <name> to
  see available streams.

  Some catalog slugs remain migration metadata only. Those entries are still
  inspectable through pm connectors inspect, but cannot execute ETL until a
  runnable connector definition or component passes conformance and is enabled.

  ETL runs read records from a configured source connector stream, add
  Polymetrics metadata fields, and write records to the destination connector.
  The MVP warehouse destination stores tables as JSONL files.

  ETL and reverse ETL are separate first-class connector surfaces: ETL reads
  streams, while pm reverse executes connector write actions where the upstream
  API supports mutations.

  ETL writes destination records in bounded batches. Use --batch-size for large
  paginated streams when you want tighter memory bounds.

  With --runtime, ETL also requires healthy PostgreSQL, DragonflyDB, and Temporal
  endpoints. It acquires a Dragonfly lease and appends a PostgreSQL run-ledger
  record after the local ETL completes.

DIRECT CONNECTOR COMMANDS
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
  pm query run --table <table> --agent-mode summary --fields id,email --sample 3
  pm query run --table <table> --agent-mode stream --fields id,email

DESCRIPTION
  The MVP query engine supports table reads and a small SELECT * FROM parser.
  Agent mode can emit compact summary JSON or projected NDJSON rows to reduce
  token usage for external agents.

FLAGS
  --table table              local warehouse table to scan
  --sql sql                  read-only SQL query; takes precedence over --table
  --limit n                  maximum rows to read; default 100
  --fields a,b               project output to selected fields
  --agent-mode summary       emit a count, sorted field list, and sample rows
  --agent-mode stream        emit one projected JSON object per line
  --sample n                 summary sample size; default 3

SECURITY
  Query output can contain data rows. Agent callers should use --fields and
  small limits or --agent-mode summary.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
`

const flowHelp = `NAME
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
`

const rlmHelp = `NAME
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
`

const scheduleHelp = `NAME
  pm schedule - create, list, install, and remove flow schedules

SYNOPSIS
  pm schedule create --name nightly --cron "0 2 * * *" --flow nightly_leads [--json]
  pm schedule list [--json]
  pm schedule install nightly [--crontab] [--json]
  pm schedule remove nightly [--crontab] [--json]

DESCRIPTION
  Schedules bind a cron expression to a named flow and install it into the
  selected local scheduler backend. Use --crontab on install or remove to force
  the crontab backend. The payload is pm flow run.

SECURITY
  Schedules do not embed secret values. Flow execution still uses the normal
  project credential references and approval gates.

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
