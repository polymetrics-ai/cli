```
NAME
  pm connectors - inspect connector definitions, streams, and write actions

SYNOPSIS
  pm connectors list [--all] [--json]
  pm connectors catalog [--capability read|write|cdc|query] [--stage stage] [--json]
  pm connectors inspect <name> [--json]
  pm connectors help
  pm connectors certify <name> [--from-env field=ENV] [--config key=value] [--stream name] [--skip write] [--write] [--full] [--keep-workdir] [--json]
  pm connectors certify --all --credentials-file <file> [--parallel n] [--resume] [--write=false | --skip=write] [--json]
  pm connectors certify --sweep [--credentials-file <file>] [--older-than 24h] [--json]

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
    --stage. Example stages include alpha, beta, and ga.

  inspect <name>
    Prints a man-style connector manual for a bare connector name. Use --json
    to print structured metadata for agents. Inspection is metadata-only and
    does not resolve credentials.

  help
    Prints this connectors namespace manual. Use inspect <name> for a
    connector-specific manual or structured metadata.

  certify
    Runs the connector certification harness through the in-process CLI. A
    single run can use local connector behavior or user-named credential
    variable references. Batch mode requires --all and --credentials-file;
    --write=false or --skip=write overrides every credential-file write entry.
    Only --skip=write is implemented. Boolean controls accept only true or
    false. Explicit --parallel must be from 1 through 32 and workers are capped
    by queued connectors. --resume reuses completed prior reports only when
    its exact schema, connector manifest, effective non-secret options, and
    environment-reference fingerprint match; other reports rerun. Sweep mode
    retries cleanup from the durable per-connector ledger in a fresh temporary
    workspace and requires plan, successful preview, approval, then execute.
    --older-than must be greater than zero and no more than 8760h. Unsupported,
    mode-inapplicable, malformed, and unknown controls fail before credential
    loading, telemetry, runner, sweep, or write effects instead of becoming
    no-ops.

    Before a certification report is completed, CLI usage errors exit 2,
    validation errors exit 3, and setup or runtime errors exit 1. These errors
    emit the normal CLI Error envelope, not a completed certification report.
    A completed certification report exits 0 for pass, 2 for certification
    failure, or 3 for leaked resources (which dominate other report outcomes).
    JSON reports use ConnectorCertification, ConnectorCertificationBatch, or
    ConnectorCertificationSweep envelopes.

    Credential values must come from environment references. Credential files
    are regular non-symlink YAML files no larger than 1 MiB, use version 1 and
    known fields, name at least one locally registered connector, and carry
    valid environment-variable references. Secret-schema fields are rejected
    from config and must use from_env. Credential-file exec entries are rejected;
    no external credential command is run. Secret values and approval
    tokens are never command operands, output, events, telemetry, or report
    fields. Reports, history, progress, and ledgers use restrictive local files.
    Live checks and writes are opt-in harness behavior, not performed by
    connector inspection or help.

EXAMPLES
  pm connectors
  pm connectors --json
  pm connectors list
  pm connectors list --all --json
  pm connectors catalog --capability write --stage ga --json
  pm connectors inspect github
  pm connectors inspect github --json
  pm connectors certify sample --json
  pm connectors certify --all --credentials-file certify/creds.yaml --json
  pm connectors certify --sweep --credentials-file certify/creds.yaml --older-than 24h --json
  pm credentials add github-public --connector github --config owner=octocat --config repo=Hello-World --config auth_type=public
  pm credentials add github-token --connector github --config owner=OWNER --config repo=REPO --config auth_type=token --from-env token=GITHUB_TOKEN
  pm credentials add github-app --connector github --config owner=OWNER --config repo=REPO --config auth_type=github_app --config app_id=12345 --config installation_id=67890 --value-stdin private_key < app.pem

SECURITY
  Connector inspection never reads credentials. Certification accepts only
  environment-variable credential references; credential-file exec is rejected.
  Secret values are excluded from output and reports.
  Live certification writes remain explicitly opt-in and cleanup-gated;
  credential-file writes additionally require sandbox=true.

EXIT STATUS
  0 command success or completed certification pass
  1 setup/runtime error before a certification report completes
  2 CLI usage error or completed certification failure
  3 CLI validation error or leaked resource in a completed certification

```
