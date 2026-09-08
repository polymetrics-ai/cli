```
NAME
  pm connectors - inspect connector definitions, streams, and write actions

SYNOPSIS
  pm connectors list [--all] [--json]
  pm connectors catalog [--capability read|write|cdc|query] [--stage stage] [--json]
  pm connectors inspect <name> [--json]
  pm connectors inspect <name> --sources [--json]
  pm connectors inspect <name> --inventory <inventory> --source-id <id> --lane <lane> [--preflight] [--json]
  pm connectors help <name>

DESCRIPTION
  pm ships with deterministic execution JSON bundles compiled into the binary.
  Authors maintain schema-4 source.lock.json files, which are projected through
  a canonical per-operation descriptor into metadata, schemas, streams, writes,
  operations, CLI surface, and optional sync transport or rate limits. The
  runtime never reads source locks or authoring evidence. Hooks or native
  components cover APIs and protocols that need custom execution behavior.

  Connectors expose local metadata for ETL read streams and reverse ETL write
  actions when those executable surfaces are implemented. Connectors without
  executable streams or writes remain visible in the catalog with explicit
  unsupported lanes. Run pm connectors inspect <name> to see write=true/false,
  ETL STREAMS, REVERSE ETL ACTIONS, and sync transport without reading
  credentials.

  RETAINED SOURCE INSPECTION
  --sources lists the retained cohort's complete source operations and all seven
  lanes, including writes, removals and source-excluded lanes. Each observation
  includes its full connector/inventory/source identity, facts and citations.
  Non-cohort connectors explicitly report not_in_cohort; this is not a claim
  about the provider's complete API. The catalog never executes an operation.

  Select exactly one cell with --inventory, --source-id and --lane. Source IDs
  are opaque: quote them when needed and use the complete inventory and ID from
  discovery. Lane is direct_read, direct_write, binary_download, binary_upload,
  etl, reverse_etl or sync_transport. No source ID is treated as a command path.
  Without --preflight this is informational inspection (exit 0). --preflight
  attempts source/lane selection and reports the current typed disposition:
    source_mapping_unproven (exit 1): the source-to-capability association is
      unresolved; an existing ordinary command may still be supported.
    missing_foundation (exit 1): a precisely scoped, source-backed demand under
      its named owner remains unresolved. This does not approve a receiver.
    source_lane_not_applicable (exit 3): the retained source excludes this lane;
      catalog state remains not_applicable.
  A future binding_observed result means an admitted association was observed,
  with execution_checked=false. It is never an execution certificate.

  Source inspection uses compiled safe metadata before project config, vault,
  state, approval/stdin, execution construction or provider access. --root is
  accepted without opening that root; global flags/environment still select
  root/JSON bootstrap options. Malformed selected metadata remains a real error.
  Ordinary commands continue to use their execution bundles independently of
  source visibility, proof status or missing source associations.

  --sources cannot combine with tuple flags or --preflight. Tuple flags require
  explicit values exactly once. --sources and --preflight are presence flags.
  Unknown, incomplete, duplicate or conflicting selectors are usage errors
  (exit 2). Source flags belong only to connectors inspect. Help exits 0 without
  selecting a cell. In --json mode a refusal writes one Error envelope with
  source_selection on stdout and a sanitized error line on stderr. Text refusals
  write no successful result to stdout. Citations establish operation identity;
  absent lane evidence is explicitly unresolved, never invented.

  JSON inspection also projects the closed sync_transport source and
  destination eligibility. A structurally valid destination that declares
  acknowledgement=none remains declared so inspection reports that policy, but
  runtime preflight refuses it; only durable_warehouse can execute. A declared
  role executes only when its named runtime executor and mode are available.

  POLLING-WATERMARK ELIGIBILITY
  polling_watermark is a bounded polling scan, not CDC or change capture.
  Its execution status, where one exists, is separate from the connector's
  CDC capability. A polling mode is executable only when runtime preflight
  accepts the specific connector, discovered catalog object, and destination
  binding after checking the rendered native source and apply executors. A
  planned, unsupported, or absent static
  binding alone does not implement a polling mode. A connector that
  constructs an implemented binding per selected catalog object can become
  eligible only after the same runtime preflight succeeds.

  An executable source uses rendered keyset ordering: a watermark and unique
  tie-breaker are checkpointed only after durable downstream acknowledgement.
  Delivery is at least once, so the inclusive resume boundary can replay an
  accepted record. Snapshot barriers are execution-bound and are never
  silently replaced by a full scan. Polling cannot observe hard deletes after a
  row disappears; tombstones require a declared, cursor-advancing soft-delete
  mapping. State incompatibility, source identity mismatch, snapshot expiry,
  and snapshot expiry require an explicit rebootstrap; pm never implies an
  automatic rescan.

  GitHub currently declares both closed transport roles, so pm connectors
  inspect github --json reports source and destination status=declared. That
  inspection result is metadata only and does not read credentials.

  For connectors with a declared rate-limit policy, inspection reports RATE
  LIMIT COORDINATION. Process-local policies coordinate only requests made by
  this pm process; they make no cross-process claim. Policies explicitly
  declaring require_shared refuse before a request when their optional shared
  coordinator is unavailable. A connector with multiple policies reports
  policy-scoped coordination. Inspection never exposes a rate scope,
  coordinator address, or credential.

  For provider-style commands with bounded caller input, JSON inspection also
  reports request_execution_limits. Each row names the command flag, request
  mapping, effective byte limit and unit, and the versioned PM execution-policy
  provenance. These are local implementation resource limits, not claims that
  the provider declared the same maximum. Connector command help renders the
  same effective cap; exceeding it rejects before provider I/O and never
  truncates the value.

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

SCHEMA-BOUND STRUCTURED WRITE INPUTS
  Some provider-sourced direct-write commands expose a declared object or array
  as a typed json flag, for example --settings or --targets. The generated
  command help and connector manual name the accepted fields and their
  maps_to=body.<field> binding. The rendered operation—not the caller—owns
  the method, route, content type, headers, and nested schema. There is no raw
  --body flag and no method, path, content-type, action, or connector override.
  A malformed, unknown, missing, oversized, or schema-incompatible structured
  value is rejected before any provider request. Direct writes still use plan,
  preview, approval, confirmation where declared, and execute; approval binds
  the exact canonical structured payload.

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
    to print structured metadata for agents, including declared sync transport,
    rate-limit coordination, and request execution limits when applicable.
    Inspection is metadata-only and does not resolve credentials or expose a
    rate scope.

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

```
