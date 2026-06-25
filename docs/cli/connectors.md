```
NAME
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
  The generated native-Go-only catalog contains 647 validated connectors:
  591 sources and 56 destinations. Use --all or the catalog subcommand when an
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

```
