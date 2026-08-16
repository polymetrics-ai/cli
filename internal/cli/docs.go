package cli

const rootHelp = `NAME
  pm - local-first Polymetrics AI ETL and reverse ETL CLI

SYNOPSIS
  pm <command> [options]

DESCRIPTION
  pm runs a dependency-free local Polymetrics AI MVP. It manages credentials,
  connectors, ETL, reverse ETL plans, local warehouse tables, and agent-safe
  JSON output from one Go binary.

  Connectors expose catalog metadata across the catalog. Executable ETL read
  streams and approval-gated reverse ETL write actions are present only when a
  connector implements those surfaces. Use pm connectors inspect <name> to see
  a connector's streams, write=true/false, and write actions.

  Every command group is also a manual page. Run pm connectors, pm etl,
  pm credentials, or any other command group without a subcommand to read its
  documentation. Use --json on a command group to return the same manual in a
  machine-readable envelope for agents.

COMMANDS
  init              create a .polymetrics project
  connectors        list and inspect connector streams and write actions
  credentials       add, link, test, inspect, list, and remove credentials
  connections       create and list source-to-destination connections
  catalog           refresh or show source catalogs
  etl               run ETL stream reads and inspect run status
  query             query local warehouse tables
  reverse           list, plan, preview, run, and inspect reverse ETL writes
  flow              plan, preview, run, list, and inspect multi-step flows
  rlm               score warehouse records with deterministic or agent RLM
  schedule          create, inspect, install, fire, and remove authorized flow schedules
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
	"init":        initHelp,
	"help":        helpHelp,
	"man":         manHelp,
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
	"extract":     extractHelp,
	"worker":      workerHelp,
}

const initHelp = `NAME
  pm init - initialize a local Polymetrics project

SYNOPSIS
  pm init [--root path] [--json]

DESCRIPTION
  Creates the .polymetrics project directory at the selected root. Run this
  once before configuring credentials, connections, or local warehouse data.

OPTIONS
  --root path    project root that will contain .polymetrics
  --json         render machine-readable JSON

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
`

const helpHelp = `NAME
  pm help - show a detailed command manual

SYNOPSIS
  pm help [<topic>] [--json]
  pm help etl transport [--json]

DESCRIPTION
  With no topic, prints the root manual. Pass a command namespace to read its
  detailed manual before creating a project or supplying credentials.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
`

const manHelp = `NAME
  pm man - alias for pm help

SYNOPSIS
  pm man [<topic>] [--json]
  pm man etl transport [--json]

DESCRIPTION
  Prints the same command manuals as pm help. Use a command namespace as the
  topic to inspect its flags and workflow before execution.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
`

const extractHelp = `NAME
  pm extract - route a natural-language data request safely

SYNOPSIS
  pm extract --request <text> [--sql query] [--limit n] [--in table] [--out table] [--json]

DESCRIPTION
  Classifies a bounded natural-language request and routes it to a typed local
  query or RLM analysis path. It never accepts an unrestricted shell, HTTP, or
  SQL write operation.

OPTIONS
  --request text       request to classify
  --sql query          optional validated local query
  --limit n            maximum returned rows; default 100
  --in table           source table for an executable RLM route
  --out table          destination table for an executable RLM route
  --json               render machine-readable JSON

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
`

const workerHelp = `NAME
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
`

const configHelp = `NAME
  pm help config - configuration reference

SYNOPSIS
  pm help config
  pm <command> --root <path> [--json]

DESCRIPTION
  pm resolves typed invocation configuration once per CLI run. The loader uses a
  fresh Viper instance for each invocation and never uses the package-level Viper
  singleton, AutomaticEnv, or file watching. Root and json affect invocation
  bootstrap. Runtime, RLM, and schedule keys are consumed by migrated runtime,
  worker, agent image, schedule, and RLM non-secret call sites. Worker/RLM agent
  Temporal execution remains opt-in: runtime.temporal_addr must be explicitly
  set by env or config file for those paths, while runtime doctor keeps local
  Compose defaults. Malformed .polymetrics/config.yaml files fail as validation
  errors.

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
  discovery for the same invocation; use --root, POLYMETRICS_ROOT, or PM_ROOT to
  select a different project root before the file is read. If a file is
  malformed, --json, POLYMETRICS_JSON=true, or PM_JSON=true selects the same
  single JSON Error envelope on stdout used by other validation errors.

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
  pm credentials add <name> --connector <connector> [--provider-family family] [--auth-profile profile] [--link-credential credential] [--from-env field=ENV] [--value-stdin field] [--config key=value]
  pm credentials link <name> --to <credential> [--json]
  pm credentials list [--json]
  pm credentials inspect <name> [--json]
  pm credentials test <name> [--json]
  pm credentials remove <name>

DESCRIPTION
  Credentials combine non-secret connector config with encrypted secret fields.
  Secrets should be supplied through environment variables or stdin, not shell
  arguments. Use --from-env field=ENV for non-interactive setup. Use
  --value-stdin field for multiline secrets such as GitHub App PEM keys.

  Provider family defaults to the connector name and auth profile to default
  when omitted. Existing credentials receive the same defaults when their
  project is opened. Each unlinked credential receives an isolated protected
  binding. Links require matching effective declarations. For a cross-connector
  link, every credential in the resulting cohort must have both declarations
  supplied explicitly; matching defaults alone are not enough.

OPTIONS
  --connector name       connector that owns the credential
  --provider-family id   non-secret provider family declaration
  --auth-profile id      non-secret authentication compatibility declaration
  --link-credential id   join a compatible credential's binding on add
  --to credential        join a compatible credential's binding
  --from-env field=ENV   read one secret field from an environment variable
  --value-stdin field    read one secret field from standard input
  --config key=value     store non-secret connector config
  --root path            project root containing .polymetrics
  --json                 render machine-readable JSON

SECURITY
  Secret values are encrypted with AES-GCM in .polymetrics/vault and are not
  stored in state.json. Inspection output shows only secret field names.
  Provider family and auth profile are non-secret credential metadata. Credential
  bindings are protected project state and are never shown in credential output;
  internal coordination receives only opaque projections. Linking records
  identity metadata only: it does not change connector authentication, rate
  limits, or transport behavior.

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
  pm connectors certify <connector> [--full] [--write-only] [--external-proof --full-parity] [--from-env field=ENV | --value-stdin field] [--json]

DESCRIPTION
  pm ships with runnable connector definitions compiled into the binary. Most
  connectors are declarative JSON bundles interpreted by the connector engine;
  hooks or native components cover APIs and protocols that need custom behavior.

  Connectors expose local metadata for ETL read streams and reverse ETL write
  actions when those executable surfaces are implemented. Planned ledger-only
  connectors remain visible in the catalog without executable streams or write
  actions. Run pm connectors inspect <name> to see write=true/false, ETL
  STREAMS, REVERSE ETL ACTIONS, and the generated certification quality
  signal without reading credentials. COMMUNITY BUILD, UNCERTIFIED is a
  warning only; the connector remains reachable.

  JSON inspection also projects the closed sync_transport source and
  destination eligibility. A structurally valid destination that declares
  acknowledgement=none remains declared so inspection reports that policy, but
  runtime preflight refuses it; only durable_warehouse can execute. A declared
  role still requires externally verified conformance; it is not a certification
  claim.

  POLLING-WATERMARK ELIGIBILITY
  polling_watermark is a bounded polling scan, not CDC or change capture.
  Its declaration status, where one exists, is separate from the connector's
  CDC capability. A polling mode is executable only when runtime preflight
  accepts the specific connector, discovered catalog object, and destination
  binding after checking the declared native source and apply executors plus
  immutable conformance evidence. A planned, unsupported, or absent static
  declaration alone does not implement a polling mode. A connector that
  constructs an implemented declaration per selected catalog object can become
  eligible only after the same runtime preflight succeeds.

  An admitted source uses declared keyset ordering: a watermark and unique
  tie-breaker are checkpointed only after durable downstream acknowledgement.
  Delivery is at least once, so the inclusive resume boundary can replay an
  accepted record. Snapshot barriers are declaration-bound and are never
  silently replaced by a full scan. Polling cannot observe hard deletes after a
  row disappears; tombstones require a declared, cursor-advancing soft-delete
  mapping. State incompatibility, source identity mismatch, snapshot expiry,
  and retention failure require an explicit rebootstrap; pm never implies an
  automatic rescan.

  GitHub currently declares both closed transport roles, so pm connectors
  inspect github --json reports source and destination status=declared. That
  inspection result is metadata only and does not read credentials.

  For connectors with a declared rate-limit policy, inspection reports RATE
  LIMIT COORDINATION. Process-local policies coordinate only requests made by
  this pm process; they make no cross-process claim. Policies explicitly
  declaring require_shared refuse before a request when their optional shared
  coordinator is unavailable. A connector with both ordinary policies reports
  policy-scoped coordination. A certification-only require_shared overlay
  preserves the process-local default label and explicitly states the
  certification boundary. Inspection never exposes a rate scope, coordinator
  address, or credential.

  The catalog command is generated from declarative bundles and Tier-3 native
  connectors. pm does not execute connector container images or accept legacy
  source-/destination-prefixed names.

CATALOG
  The connector catalog is generated from local connector metadata. The current
  runtime catalog has 556 bare-name entries: 552 declarative bundles plus the
  local sample, file, warehouse, and outbox primitives. Use --all or the catalog
  subcommand when an agent needs to discover the complete connector universe.
  Use --capability read, write, cdc, or query to filter by executable surface.

GITHUB AUTHENTICATION
  public
    Unauthenticated public repository reads. Configure owner and repo plus
    public_access=true (or auth_type=public).
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
    to print structured metadata for agents, including the generated binary
    certification status and declared rate-limit coordination provenance when
    applicable. Inspection is metadata-only and does not resolve credentials or
    expose a rate scope. A connector is either CERTIFIED or COMMUNITY BUILD,
    UNCERTIFIED; the latter remains available with a warning.

  help <name>
    Alias for the human connector manual.

  certify <connector>
    Runs the legacy connector test harness. It does not set the generated
    CERTIFIED status; only proof-bearing certification records can do that.
    --external-proof is an explicit live HTTPS acceptance mode: it builds a
    fresh pm child binary, accepts credentials only from --from-env or
    --value-stdin,
    and writes a fingerprint-only transcript after a full-parity run. It
    requires --full-parity and refuses incomplete or truncated exchanges.
    --full-parity enables both the full read sweep and live writes; it refuses
    to claim parity unless every applicable declared write has a production
    mutation, independent read-back, and verified cleanup. Skipped, not_live,
    recovered_unverified, blocked, failed, or leaked actions are never folded
    into a pass result.
    --write-only is a bounded GitHub repository-fixture wave, not a parity
    claim: it is restricted to Polymetrics-Cert/pm-cert-3993-20260810-wz0fru,
    the captain-approved run-owned disposable fixture, and records every
    non-live boundary explicitly. Commit-comment actions currently require
    GitHub's "Metadata" repository permission (read) for their item read-back;
    without it they are reported blocked, never pass.
    After that permission is granted, enable their bounded re-run explicitly
    with --config certification_commit_comment_item_read=enabled; the default
    avoids creating a further unverified commit comment while it is missing.
    With --full --json from a source checkout,
    the report includes the API-surface inventory and provider-artifact
    provenance evidence separately from endpoint coverage and connector
    capabilities. Version-1 and pre-ledger inventories remain legacy_unverified
    during the staged migration. A complete version-2 ledger reports its ledger
    version, artifact count, endpoint count, and cited endpoint count; invalid
    version-2 provenance fails certification without enabling or changing any
    connector capability. When the connector declares coordinated rate limits,
    JSON may also contain safe rate_limit_events for attempts, observed resets,
    waits, and requests stopped before send; the events contain no credentials
    or rendered rate scopes.

    PostgreSQL's full database proof requires --write plus an explicit
    --stream schema.table and cursor_field configuration. It certifies live
    catalog discovery, one bounded typed relation read, and PostgreSQL's six
    declared polling-to-managed-target modes with independent target read-back.
    It does not claim every relation in a dynamic database or a direct
    writes.json action surface.

EXAMPLES
  pm connectors
  pm connectors --json
  pm connectors list
  pm connectors list --all --json
  pm connectors catalog --capability write --stage generally_available --json
  pm connectors inspect github
  pm connectors inspect github --json
  pm connectors certify sample --full --json
  pm connectors certify github --external-proof --full-parity --config owner=OWNER --config repo=REPO --from-env token=GITHUB_TOKEN --json
  pm connectors certify postgres --full --write --stream public.events --config host=DB_HOST --config port=5432 --config database=DB_NAME --config username=DB_USER --config schema=public --config cursor_field=sequence --from-env password=POSTGRES_PASSWORD --json
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
  pm connections create <name> --source connector:credential --destination connector:credential --stream stream [--sync-mode mode] [--cursor field] [--primary-key field] [--table table] [--transform-file plan.json] [--target-copy-workers n]
  pm connections list [--json]

DESCRIPTION
  A connection joins one source endpoint to one destination endpoint and stores
  stream-level sync settings.

TARGET COPY CAPACITY
  --target-copy-workers records the bounded target connection capacity for an
  immutable transformed full_overwrite COPY destination. PostgreSQL currently
  declares a maximum of 8, so its default is 2 and accepted values are 1..8;
  another destination is accepted only when its own transport declaration
  supplies a lower maximum. The saved policy is included in the closed target
  plan and preview. It is not a run flag and does not permit unordered apply.
  This release has one ordered COPY consumer; a second COPY lane remains a
  separately measured follow-on rather than an implied consequence of this
  configuration value.

TRANSFORMS
  --transform-file reads one bounded JSON TransformPlanV1 during creation. The
  file is validated against the source's typed catalog before any connection is
  saved. pm stores only the normalized closed plan and its SHA-256 hash; it
  never stores the filename or accepts arbitrary SQL. The admitted vocabulary
  is typed projection/rename, date, checked multiply/cast, upper, mod, and a
  not_equal filter.

CONNECTION NAMES
  A name may contain letters, digits, '-' and '_', must start with a letter or
  digit, and is limited to 128 characters. Names that differ only by letter case
  are refused. Ambiguous names are rejected at creation rather than rewritten,
  because two connections that cannot be told apart cannot own separate data.
  The name is a display value: the local warehouse keys its directories on a
  generated identifier, so renaming is safe and never moves data.

STREAM AND TABLE NAMES
  Against the local warehouse destination, --stream and --table become path
  components, so each may contain only letters, digits, '.', '-' and '_'. They
  are checked when the connection is created, because a name the warehouse
  cannot materialize would otherwise fail every sync of that connection.

  One local-warehouse connection cannot configure distinct --table spellings
  that differ only by ASCII letter case, such as records and RECORDS: DuckDB
  treats them as one identifier. Creation refuses that inventory before saving
  it. A legacy inventory is left unchanged on open, but any local sync refuses
  before changing run or warehouse state; create replacement connections whose
  destination table names differ by more than ASCII letter case.

SYNC MODES
  full_refresh_append              read all source records and append them
  full_refresh_overwrite           read all source records and replace final output
  full_refresh_overwrite_deduped   compatibility name for typed full_overwrite admission
  incremental_append               append records at or after the saved cursor
  incremental_append_deduped       compatibility name for typed incremental_dedupe admission
  incremental_dedupe               typed current-state dedupe for an admitted source-to-warehouse transport
  incremental_dedupe_history       typed source-version history for an admitted source-to-warehouse transport

  Incremental modes and deduped compatibility names require --cursor. Deduped
  modes require --primary-key. A static connector manifest advertises the full
  deduped compatibility name only with both fields, and incremental modes only
  with a declared incremental executor. The two deduped compatibility names use
  their typed contract and refuse before source I/O until a matching transport
  is admitted. The two raw typed modes are selected only when source and
  warehouse transport declarations, registrations, and preflight admit them;
  otherwise they refuse before source I/O. When a connector manifest declares
  defaults, pm fills them during connection creation.

POLLING-WATERMARK LIMITS
  polling_watermark is not a general connection mode and is not CDC. It can be
  selected only by a connector's declared native source/object/destination
  binding after runtime preflight succeeds. An admitted source resumes from its
  declared watermark plus unique tie-breaker after durable downstream
  acknowledgement, so replay is at least once. A polling scan cannot observe a
  hard delete after the row is gone; delete-aware history requires a declared
  cursor-advancing soft-delete mapping. Incompatible state, source identity
  changes, snapshot expiry, and retention failures require explicit
  rebootstrap rather than an automatic full scan.

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
  Catalog refresh calls the source connector and stores a local snapshot;
  catalog show reads that persisted snapshot.
  refresh deliberately fetches a new provider catalog; show reads the
  existing snapshot and marks it stale when its discovery expiry has passed.
  Refresh a stale catalog before relying on fields added by the provider.

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
  pm etl run --connection <name> --stream <stream> [--batch-size n] [--max-in-flight-batches n] [--runtime] [--json]
  pm etl status <run-id> [--json]
  pm etl transport github-issue-label plan --connection <name> [--json]
  pm etl transport github-issue-label preview <plan-id> [--json]
  pm etl run --connection <name> --stream <stream> --batch-size 1 --approval-plan <plan-id> [--approval-token-stdin] --confirm destructive [--json]
  pm etl transport postgres-managed-target plan --connection <name> --stream <stream> [--authorization-lifetime <24h..48h>] [--json]
  pm etl transport postgres-managed-target preview <plan-id> [--json]
  pm etl transport github-issue-label cleanup plan --connection <name> --forward-plan <plan-id> [--json]
  pm etl transport github-issue-label cleanup run <plan-id> --connection <name> --approval-token-stdin --confirm destructive [--json]

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
  The warehouse destination uses an appendable JSONL write-ahead log and rebuilds
  each final table as a single Parquet file.

  ETL and reverse ETL are separate first-class connector surfaces: ETL reads
  streams, while pm reverse executes connector write actions where the upstream
  API supports mutations.

  ETL writes destination records in bounded batches. Use --batch-size for large
  paginated streams when you want tighter memory bounds.

ORDERED PIPELINE
  --max-in-flight-batches selects a bounded 1..8 producer/consumer depth only
  for a transformed full_overwrite Arrow transport whose source and destination
  both declare ordered-pipeline support. Its admitted default is 2; 1 preserves
  the prior serial callback behavior. The bound covers retained batches as well
  as the Arrow byte-credit policy, preserves source order, and refuses an
  undeclared endpoint before source or destination I/O. It is not a generic
  --workers flag and does not create parallel destination COPY lanes.

  With --runtime, ETL also requires healthy PostgreSQL, DragonflyDB, and Temporal
  endpoints. It acquires a Dragonfly lease and appends a PostgreSQL run-ledger
  record after the local ETL completes.

CLOSED ISSUE-LABEL TRANSPORT
  The fixed two-action issue-label destination is not a generic writer. Its
  connector definition declares the admitted source executors, source streams,
  and bounded record mappings. A saved connection owns the repository, source
  selection, target issue, label, action, and credential configuration; the
  command accepts none of those provider details directly.

  An input-fields source supplies only the destination definition's declared
  inputs: target_issue (a positive integer) and label (a non-empty string).
  full_append selects add_issue_labels; incremental_upsert selects
  set_issue_labels and requires transport_allow_keyed=true. The row's derived
  issue and label must equal the plan-bound destination configuration, so
  values cannot drift after destructive approval. Null, malformed, mismatched,
  or delete/tombstone rows are refused before destination write I/O. Use
  --batch-size 1 for this singleton destination contract.

  Create a closed plan, preview it in human output to obtain an ephemeral
  approval token, then pass that token only as one bounded stdin line to:

    pm etl run --connection <name> --stream <stream> --batch-size 1 \
      --approval-plan <plan-id> --approval-token-stdin --confirm destructive

  The run keeps the source -> durable warehouse -> reopen -> typed GitHub
  mutation and durable acknowledgement -> independent read-back -> checkpoint
  order. After the first approved non-additive run, later runs with the exact
  same plan scope use --approval-plan and --confirm destructive without a new
  --approval-token-stdin. Changed, expired, or revoked scope is refused before
  a provider write. Cleanup is a separate typed remove-label plan, preview, and
  one-time approval. A declared GitHub missing-label DELETE is a successful
  cleanup; replaying approval is refused.

  A GitHub source selection and every independent target read-back inspect only
  the first GitHub issues page. The transport fails instead of requesting
  another page when the configured GitHub source or target issue is not there.

  Approval tokens are never accepted in argv, environment variables, files,
  JSON output, or persisted project state. Run pm etl transport for the exact
  closed lifecycle and its stdin-only token rule.

CLOSED POSTGRESQL MANAGED-TARGET TRANSPORT
  postgres-managed-target is a fixed definition-selected source-to-PostgreSQL
  route, not a generic SQL writer. The saved connection binds both credentials,
  the source stream and sealed schema, primary key, mode, and immutable managed
  destination identity. PostgreSQL sources use their typed relation catalog;
  declared API sources use their authoritative JSON stream schema.

  Create and preview a plan, then pass its one-time token only through stdin to
  the ordinary ETL run:

    pm etl transport postgres-managed-target plan \
      --connection <name> --stream <stream> [--authorization-lifetime <24h..48h>] --json
    pm etl transport postgres-managed-target preview <plan-id>
    pm etl run --connection <name> --stream <stream> --batch-size 1000 \
      --approval-plan <plan-id> --approval-token-stdin --confirm destructive

  The registered source stages bounded pages in the connection-owned warehouse.
  The registered destination reopens the sealed Parquet, derives an immutable
  workset, applies it through the native managed target, reads the receipt back,
  and advances the source checkpoint only after durable acknowledgement.
  With transport_bootstrap=true on an incremental_upsert PostgreSQL source
  credential, a gap-free snapshot barrier continues into committed pgoutput
  transactions and resumes from the acknowledged LSN after restart.

  Approval binds the live source schema and both credential revisions. The
  one-time preview token creates a durable, revocable authorization with a
  24h default lifetime (configurable from 24h through 48h at plan time).
  Each provider page fetch and staged PostgreSQL apply has its own bounded
  deadline, so a stalled unit stops without expiring the overall authority.
  Stale, replayed, authentication-refused, and permission-refused runs stop
  before a checkpoint advance. The public PostgreSQL connector remains
  write=false and this route accepts no raw SQL or target identifiers.

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
    Local Parquet warehouse tables. Supports append, overwrite, append_dedup, and
    overwrite_dedup destination behavior through ETL sync modes.

SYNC MODES
  full_refresh_append
    Reads every source record and appends to the write-ahead log, then rebuilds
    the final Parquet table from it. Duplicates across runs are expected.

  full_refresh_overwrite
    Replaces the write-ahead log with this run's records, then atomically
    replaces the final Parquet table only after the run succeeds.

  full_refresh_overwrite_deduped
    Compatibility name for typed full_overwrite admission. pm refuses before
    source I/O until a matching transport is admitted.

  incremental_append
    Reads records at or after the saved cursor and appends accepted records to
    the write-ahead log. Cursor state advances only after successful writes.

  incremental_append_deduped
    Compatibility name for typed incremental_dedupe admission. pm refuses
    before source I/O until a matching transport is admitted.

  incremental_dedupe
    For an admitted source-to-warehouse transport, retains one current record
    per declared primary key. It refuses before source I/O for other pairs.

  incremental_dedupe_history
    For an admitted source-to-warehouse transport, retains deduplicated source versions with _valid_from, _valid_to, and _is_current fields. It requires
    declared primary-key and cursor fields, and refuses before source I/O for
    other pairs.

  Incremental modes and deduped compatibility names require --cursor. Deduped
  modes require --primary-key. Static connector manifests advertise the full
  deduped compatibility name only with both fields, and incremental modes only
  with a declared incremental executor.

POLLING-WATERMARK LIMITS
  polling_watermark is a bounded keyset scan, not CDC or a generic database
  query. The runtime evaluates every mode against the declared source ordering,
  discovered object, destination binding, registered executors, and conformance
  evidence before source I/O. A durable checkpoint records the watermark and
  unique tie-breaker only after downstream acknowledgement, so accepted records
  may replay. Hard deletes are not observable unless the declaration supplies
  a cursor-advancing soft-delete mapping. Incompatible state, source identity
  mismatch, snapshot expiry, and retention failure require explicit rebootstrap;
  pm never converts them into an automatic full scan.

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
  pm query run --table <table> [--connection name] [--limit n] [--json]
  pm query run --sql "select status, count(*) from <table> group by status" [--json]
  pm query run --table <table> --agent-mode summary --fields id,email --sample 3
  pm query run --table <table> --agent-mode stream --fields id,email

DESCRIPTION
  Queries run on an embedded DuckDB engine over the warehouse's Parquet tables,
  so --sql accepts read-only SELECT and WITH statements in full: joins, filters,
  aggregates, GROUP BY, window functions and CTEs. Writes are refused.
  Agent mode can emit compact summary JSON or projected NDJSON rows to reduce
  token usage for external agents.

  Each connection materializes its tables into its own directory, so two
  connections can use the same table name without overwriting each other. When
  more than one connection has a table of the requested name, the read is
  refused and lists the owning connections; pass --connection to pick one.
  A legacy connection that itself configured distinct table spellings differing
  only by ASCII letter case is different: --connection cannot choose between
  one owner's destinations, so SQL references are refused. Use --table only
  with an exact resolver-visible spelling to inspect retained data, or create
  replacement connections whose destination table names differ by more than
  ASCII letter case.
  A table at the warehouse root belongs to no connection, because a reverse ETL
  run writing to the warehouse connector produced it rather than a sync, or it
  was seeded by hand. It is listed and selected as _unattributed.

FLAGS
  --table table              local warehouse table to scan
  --connection name          connection whose table to read; required only when
                             several connections share the table name; use
                             _unattributed for a root-level table
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
  pm schedule - run existing approved-job flows on a local scheduler

SYNOPSIS
  pm schedule create --name nightly --cron "0 2 * * *" --flow nightly_leads [--json]
  pm schedule list [--json]
  pm schedule inspect nightly [--json]
  pm schedule status nightly [--json]
  pm schedule install nightly [--crontab] [--json]
  pm schedule remove nightly [--crontab] [--json]
  pm schedule fire nightly [--json]

DESCRIPTION
  Approve each ETL or reverse-ETL job once, compose those existing approved jobs
  into a stored flow with pm flow create, then schedule that existing flow.
  Create refuses a missing or invalid flow before writing a schedule. Install
  revalidates the flow before touching the scheduler backend.

  The selected backend invokes exactly:

    pm --root <root> flow run <name> --json

  No approval token or authorization reference is placed in crontab, argv, a
  schedule manifest, or schedule JSON. Use inspect or status to view terminal
  flow status, safe prepared-execution identities, and opaque receipt IDs. Use
  --crontab on install or remove to force the crontab backend.

SECURITY
  Each unattended firing reloads every referenced job and revalidates credential
  revision, manifest/schema, source scope, mappings, destination action,
  confirmation policy, expiry, and revocation before a provider request. Any
  drift refuses and parks the schedule. Failed, rate-limited, ambiguous, or
  cleanup-failed writes also park or halt and never replay automatically.

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

  The sync-modes subcommand benchmarks local sync modes that materialize without
  a closed transport. Typed compatibility names that refuse before source I/O
  are excluded.

SECURITY
  Performance output contains counts and durations only.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error
`
