```
NAME
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

GITHUB CERTIFICATION
  Full certification passed for the current GitHub connector surface. The
  certificate accounted for 509 API endpoints: 440 covered and 69 explicitly
  blocked. It covered 37 catalog streams, 2 implemented direct-read command
  families, and 231 write actions. The safe create_label write lifecycle passed
  with read-back verification and cleanup; destructive/admin/binary surfaces are
  not executed blindly, and binary download surfaces remain safely blocked until
  a bounded binary executor and destination policy exist.

FRESHCHAT COMMAND SURFACE
  pm freshchat
  pm freshchat --help
    Print the Freshchat connector command surface without resolving credentials.
    The help output includes implemented ETL stream commands such as user list,
    agent list, conversation fields, report status, and metrics instant, plus
    reverse ETL command metadata for approved plan/preview/execute workflows.

  pm freshchat user list --credential <name> --limit 50 --json
    Runs the stream-backed users command through a saved Freshchat credential.
    Use --config for scoped detail streams until direct-read path flags are
    implemented.

  Freshchat writes remain reverse ETL only: plan, preview, approval token, then
  execute. Multipart file/image upload commands are listed as excluded and do
  not expose a raw upload escape hatch.

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

```
