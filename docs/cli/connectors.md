```
NAME
  pm connectors - inspect connector definitions, streams, and write actions

SYNOPSIS
  pm connectors list [--all] [--json]
  pm connectors catalog [--capability read|write|cdc|query] [--stage stage] [--json]
  pm connectors inspect <name> [--json]
  pm connectors help <name>
  pm connectors certify <connector> [--full] [--external-proof --full-parity] [--from-env field=ENV | --value-stdin field] [--json]

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
  pm credentials add github-public --connector github --config owner=octocat --config repo=Hello-World --config auth_type=public
  pm credentials add github-token --connector github --config owner=OWNER --config repo=REPO --config auth_type=token --from-env token=GITHUB_TOKEN
  pm credentials add github-app --connector github --config owner=OWNER --config repo=REPO --config auth_type=github_app --config app_id=12345 --config installation_id=67890 --value-stdin private_key < app.pem

SECURITY
  Connector inspection never reads credentials.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
