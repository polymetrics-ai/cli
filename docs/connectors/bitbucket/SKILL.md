---
name: pm-bitbucket
description: Bitbucket connector knowledge and safe action guide.
---

# pm-bitbucket

## Purpose

Bitbucket Cloud connector metadata for repository, pull request, issue, pipeline, deployment, download, webhook, and workspace CLI parity planning.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics
- review_url: https://github.com/polymetrics-ai/cli

## Capabilities

- check=false catalog=true read=false write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- repo_slug
- workspace
- access_token (secret)

## Security

- read risk: Metadata-only seed; no Bitbucket read execution is enabled in this slice.
- write risk: No Bitbucket writes are enabled in this slice. Future reverse ETL actions must use plan, preview, approval, execute.
- approval: Destructive, admin, and sensitive Bitbucket actions remain blocked by default pending policy work.
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Work with Bitbucket Cloud repositories from the command line.
- Usage: pm bitbucket <command> <subcommand> [flags]
- Global flags:
  - --json (boolean): Write machine-readable JSON output.
  - --connection (string): Use a saved Bitbucket connector credential and scope.: maps_to=connection
  - --workspace (string): Bitbucket workspace slug.: maps_to=config.workspace
  - --repo (string): Bitbucket repository slug.: maps_to=config.repo_slug
- Repository Commands
  - repo list - List repositories in a workspace [intent=etl availability=planned]; notes: Future stream-backed command for workspace repository listings.
  - repo view - View repository details [intent=direct_read availability=planned]; notes: Safe direct-read candidate; blocked until #94 defines Bitbucket output policy and executor support.; flags: --repo
  - repo create - Create a repository [intent=reverse_etl availability=partial]; approval: Future reverse ETL writes require plan, preview, approval, execute, and typed confirmation when policy marks the target as admin-level.; risk: Creates a repository in the target workspace.; notes: No write action exists in this metadata slice.
  - repo delete - Delete a repository [intent=direct_write availability=unsafe_or_disallowed]; notes: Repository deletion is destructive admin behavior and is not exposed as a connector write.
  - repo clone - Clone a repository locally [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Depends on local git and filesystem state; no local git executor is enabled.
  - branch list - List repository branches [intent=etl availability=planned]; notes: Future stream-backed repository refs command.
  - commit list - List repository commits [intent=etl availability=planned]; notes: Future stream-backed commit history command.
  - tag list - List repository tags [intent=etl availability=planned]; notes: Future stream-backed tag command.
  - download list - List repository downloads [intent=etl availability=planned]; notes: Download metadata is a future stream candidate; binary transfer is separate and bounded.
  - download get - Download a repository file asset [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Binary download to local filesystem requires explicit max-byte, destination, overwrite, and archive policies before execution.
- Pull Request Commands
  - pull-request list - List pull requests [intent=etl availability=planned]; notes: Future stream-backed command.; flags: --state
  - pull-request view - View pull request details [intent=direct_read availability=planned]; notes: Safe direct-read candidate; output policy and executor deferred to #94.
  - pull-request create - Create a pull request [intent=reverse_etl availability=partial]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a visible pull request in the configured repository.; notes: No write action exists in this metadata slice.; flags: --source-branch, --destination-branch, --title
  - pull-request merge - Merge a pull request [intent=reverse_etl availability=partial]; approval: reverse ETL writes require plan, preview, approval, execute, and may require typed confirmation by policy.; risk: Merges code into the destination branch.; notes: High-impact repository mutation; no write action exists in this metadata slice.
  - pull-request decline - Decline a pull request [intent=reverse_etl availability=partial]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Declines an open pull request.; notes: No write action exists in this metadata slice.
- Issue Tracker Commands
  - issue list - List issues [intent=etl availability=planned]; notes: Future stream-backed issue tracker command.; flags: --state
  - issue view - View issue details [intent=direct_read availability=planned]; notes: Safe direct-read candidate; output policy and executor deferred to #94.
  - issue create - Create an issue [intent=reverse_etl availability=partial]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a visible issue in the configured repository.; notes: No write action exists in this metadata slice.
  - issue comment - Comment on an issue [intent=reverse_etl availability=partial]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Adds a visible issue comment.; notes: No write action exists in this metadata slice.
- Pipelines And Deployments Commands
  - pipeline list - List pipelines [intent=etl availability=planned]; notes: Future stream-backed pipeline command.
  - pipeline view - View pipeline details [intent=direct_read availability=planned]; notes: Safe direct-read candidate; output policy and executor deferred to #94.
  - pipeline run - Run a pipeline [intent=reverse_etl availability=partial]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Starts a Bitbucket pipeline execution.; notes: No write action exists in this metadata slice.
  - pipeline stop - Stop a pipeline [intent=reverse_etl availability=partial]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Stops an in-flight Bitbucket pipeline.; notes: No write action exists in this metadata slice.
  - deployment list - List deployments [intent=etl availability=planned]; notes: Future stream-backed deployment command.
- Workspace And Administration Commands
  - workspace list - List accessible workspaces [intent=direct_read availability=planned]; notes: Viewer-scoped direct-read candidate; executor deferred to #94.
  - project list - List workspace projects [intent=direct_read availability=planned]; notes: Workspace-scoped direct-read candidate; executor deferred to #94.
  - webhook list - List repository webhooks [intent=etl availability=planned]; notes: Future stream-backed webhook metadata command.
  - webhook create - Create a repository webhook [intent=reverse_etl availability=partial]; approval: reverse ETL writes require plan, preview, approval, execute, and URL policy review.; risk: Creates an outbound webhook that can send repository events to an external URL.; notes: No write action exists in this metadata slice.
  - webhook delete - Delete a repository webhook [intent=reverse_etl availability=partial]; approval: reverse ETL writes require plan, preview, approval, execute, and typed confirmation when policy marks the webhook as sensitive.; risk: Deletes an existing webhook and may interrupt downstream automation.; notes: No write action exists in this metadata slice.
  - branch-restriction list - List branch restrictions [intent=etl availability=planned]; notes: Future stream-backed branch policy command.
  - branch-restriction create - Create a branch restriction [intent=reverse_etl availability=partial]; approval: reverse ETL writes require plan, preview, approval, execute, and typed confirmation for admin policy changes.; risk: Changes repository branch protection behavior.; notes: No write action exists in this metadata slice.
  - branch-restriction delete - Delete a branch restriction [intent=reverse_etl availability=partial]; approval: reverse ETL writes require plan, preview, approval, execute, and typed confirmation for admin policy changes.; risk: Removes branch protection behavior.; notes: No write action exists in this metadata slice.
  - snippet list - List snippets [intent=direct_read availability=planned]; notes: Direct-read candidate; snippet file content and binary behavior require output policy review.
  - snippet create - Create a snippet [intent=reverse_etl availability=partial]; approval: reverse ETL writes require plan, preview, approval, execute, and content redaction review.; risk: Creates a Bitbucket snippet and may publish code or text content.; notes: No write action exists in this metadata slice.
- Local Workflow Commands
  - auth status - Show credential status [intent=auth availability=unsupported_local unsupported local workflow]; notes: Use `pm credentials inspect <name> --redacted`; this metadata does not read secrets.
  - config view - Show Bitbucket command configuration [intent=config availability=unsupported_local unsupported local workflow]; notes: Connector-local CLI configuration is not implemented.
  - api - Call an arbitrary Bitbucket API endpoint [intent=raw_api availability=unsafe_or_disallowed]; notes: Generic raw API calls are forbidden. Add reviewed direct-read or reverse-ETL operations instead.
- Help topics:
  - safety - Bitbucket writes remain plan, preview, approval, execute; generic raw API writes are disallowed.
  - coverage - This metadata slice is a command map only. Full 331-operation ledger coverage is owned by issue #93.

## Commands

### Inspect as a manual

```bash
pm connectors inspect bitbucket
```

### Inspect as structured JSON

```bash
pm connectors inspect bitbucket --json
```

## Agent Rules

- Run pm connectors inspect bitbucket before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
