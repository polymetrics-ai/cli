# pm connectors inspect bitbucket

```text
NAME
  pm connectors inspect bitbucket - Bitbucket connector manual

SYNOPSIS
  pm connectors inspect bitbucket
  pm connectors inspect bitbucket --json
  pm credentials add <name> --connector bitbucket [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Bitbucket Cloud repositories, branches, commits, tags, pull requests, issues, pipelines, deployments, downloads metadata, webhooks, branch restrictions, projects, and snippets; exposes approval-gated write plans for selected repository, issue, pull request, pipeline, webhook, branch restriction, and snippet mutations.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  repo_slug
  start_date
  workspace
  access_token (secret)

ETL STREAMS
  repositories:
    primary key: uuid
    fields: created_on(), description(), fork_policy(), full_name(), id(), is_private(), language(), links(), mainbranch(), name(), project(), repository(), scm(), slug(), state(), title(), type(), updated_on(), uuid(), workspace()
  branches:
    primary key: name
    fields: created_on(), default_merge_strategy(), description(), full_name(), id(), links(), merge_strategies(), name(), repository(), slug(), state(), target(), target_hash(), title(), type(), updated_on(), uuid(), workspace()
  commits:
    primary key: hash
    cursor: date
    fields: author(), created_on(), date(), description(), full_name(), hash(), id(), links(), message(), name(), parents(), repository(), slug(), state(), summary(), title(), type(), updated_on(), uuid(), workspace()
  tags:
    primary key: name
    fields: created_on(), description(), full_name(), id(), links(), name(), repository(), slug(), state(), target(), target_hash(), title(), type(), updated_on(), uuid(), workspace()
  pull_requests:
    primary key: id
    cursor: updated_on
    fields: author(), author_display_name(), close_source_branch(), comment_count(), created_on(), description(), destination(), destination_branch(), full_name(), id(), links(), name(), participants(), repository(), reviewers(), slug(), source(), source_branch(), state(), summary(), task_count(), title(), type(), updated_on(), uuid(), workspace()
  issues:
    primary key: id
    cursor: updated_on
    fields: assignee(), assignee_display_name(), component(), content(), created_on(), description(), full_name(), id(), kind(), links(), milestone(), name(), priority(), reporter(), reporter_display_name(), repository(), slug(), state(), title(), type(), updated_on(), uuid(), version(), votes(), watches(), workspace()
  pipelines:
    primary key: uuid
    cursor: created_on
    fields: build_number(), completed_on(), created_on(), creator(), description(), duration_in_seconds(), full_name(), id(), links(), name(), repository(), slug(), state(), target(), title(), trigger(), type(), updated_on(), uuid(), workspace()
  deployments:
    primary key: uuid
    cursor: created_on
    fields: created_on(), deployment_state(), description(), environment(), full_name(), id(), last_update_time(), links(), name(), release(), repository(), slug(), state(), title(), type(), updated_on(), uuid(), workspace()
  downloads:
    primary key: name
    fields: created_on(), description(), full_name(), id(), links(), name(), repository(), size(), slug(), state(), title(), type(), updated_on(), user(), uuid(), workspace()
  webhooks:
    primary key: uuid
    fields: active(), created_on(), description(), events(), full_name(), id(), links(), name(), repository(), slug(), state(), subject_type(), title(), type(), updated_on(), url(), uuid(), workspace()
  branch_restrictions:
    primary key: id
    fields: branch_match_kind(), created_on(), description(), full_name(), groups(), id(), kind(), links(), name(), pattern(), repository(), slug(), state(), title(), type(), updated_on(), users(), uuid(), value(), workspace()
  projects:
    primary key: key
    cursor: updated_on
    fields: created_on(), description(), full_name(), id(), is_private(), key(), links(), name(), repository(), slug(), state(), title(), type(), updated_on(), uuid(), workspace()
  snippets:
    primary key: id
    cursor: updated_on
    fields: created_on(), creator(), description(), files(), full_name(), id(), is_private(), links(), name(), owner(), repository(), scm(), slug(), state(), title(), type(), updated_on(), uuid(), workspace()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_repository:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}
    risk: creates a Bitbucket repository in the configured workspace
  create_issue:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/issues
    risk: creates a visible Bitbucket issue and may notify repository participants
  update_issue:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/issues/{{ record.issue_id }}
    required fields: issue_id
    risk: mutates an existing Bitbucket issue
  create_pull_request:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests
    risk: creates a visible pull request and may notify reviewers
  merge_pull_request:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pull_request_id }}/merge
    required fields: pull_request_id
    risk: merges a pull request into its destination branch
  decline_pull_request:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pull_request_id }}/decline
    required fields: pull_request_id
    risk: declines a Bitbucket pull request
  run_pipeline:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines
    risk: starts a Bitbucket pipeline run that may consume CI minutes and deploy artifacts
  stop_pipeline:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines/{{ record.pipeline_uuid }}/stopPipeline
    required fields: pipeline_uuid
    risk: stops an in-flight Bitbucket pipeline
  create_webhook:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/hooks
    risk: creates an outbound webhook that sends repository events to an external URL
  delete_webhook:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/hooks/{{ record.uid }}
    required fields: uid
    risk: deletes a repository webhook and may interrupt downstream automation
  create_branch_restriction:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/branch-restrictions
    risk: changes repository branch protection behavior
  create_snippet:
    endpoint: POST /snippets/{{ config.workspace }}
    risk: creates a Bitbucket snippet that may publish code or text content

SECURITY
  read risk: Bitbucket Cloud REST API reads scoped to the configured workspace/repository; binary payloads and local git workflows are not executed.
  write risk: Selected Bitbucket mutations are explicit reverse ETL actions only; destructive/admin/sensitive operations are blocked or require typed confirmation metadata.
  approval: Reverse ETL writes require plan, preview, approval, execute; destructive/admin operations require typed confirmation when exposed.
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Work with Bitbucket Cloud repositories from the command line.
  Usage: pm bitbucket <command> <subcommand> [flags]
  Source CLI: bb/Bitbucket Cloud REST (https://developer.atlassian.com/cloud/bitbucket/rest/)
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --credential (string): Use a saved Bitbucket connector credential.: maps_to=connection
    --connection (string): Alias for --credential.: maps_to=connection
    --workspace (string): Bitbucket workspace slug.: maps_to=config.workspace
    --repo (string): Bitbucket repository slug.: maps_to=config.repo_slug
    --limit (integer): Maximum records to emit for stream commands.
    --max-bytes (integer): Maximum bytes for direct-read JSON responses.
  Repository Commands
    repo list - List repositories in a workspace [intent=etl availability=implemented stream=repositories]
    repo view - View repository details [intent=direct_read availability=implemented]; flags: --repo
    repo create - Create a repository [intent=reverse_etl availability=implemented write=create_repository]; approval: reverse ETL writes require plan, preview, approval, execute. Typed confirmation may be required by policy.; risk: Creates a repository in the configured workspace.; flags: --name, --description, --is-private
    repo delete - Delete a repository [intent=direct_write availability=unsafe_or_disallowed]; notes: Repository deletion is destructive admin behavior and is not exposed as a connector write.
    repo clone - Clone a repository locally [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Depends on local git and filesystem state; no local git executor is enabled.
    branch list - List repository branches [intent=etl availability=implemented stream=branches]
    commit list - List repository commits [intent=etl availability=implemented stream=commits]
    tag list - List repository tags [intent=etl availability=implemented stream=tags]
    download list - List repository downloads [intent=etl availability=implemented stream=downloads]
    download get - Download a repository file asset [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Binary download to local filesystem requires explicit max-byte, destination, overwrite, and archive policies before execution.
  Pull Request Commands
    pull-request list - List pull requests [intent=etl availability=implemented stream=pull_requests]; flags: --state
    pull-request view - View pull request details [intent=direct_read availability=implemented]; flags: --pull-request-id
    pull-request create - Create a pull request [intent=reverse_etl availability=implemented write=create_pull_request]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a visible pull request in the configured repository.; flags: --source-branch, --destination-branch, --title
    pull-request merge - Merge a pull request [intent=reverse_etl availability=implemented write=merge_pull_request]; approval: reverse ETL writes require plan, preview, approval, execute. Typed confirmation may be required by policy.; risk: Merges code into the destination branch.; flags: --pull-request-id, --message
    pull-request decline - Decline a pull request [intent=reverse_etl availability=implemented write=decline_pull_request]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Declines an open pull request.; flags: --pull-request-id, --message
  Issue Tracker Commands
    issue list - List issues [intent=etl availability=implemented stream=issues]; flags: --state
    issue view - View issue details [intent=direct_read availability=implemented]; flags: --issue-id
    issue create - Create an issue [intent=reverse_etl availability=implemented write=create_issue]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a visible issue in the configured repository.; flags: --title, --kind, --priority
    issue edit - Edit an issue [intent=reverse_etl availability=implemented write=update_issue]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Mutates an existing issue.; flags: --issue-id, --title, --state
    issue comment - Comment on an issue [intent=reverse_etl availability=partial]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Adds a visible issue comment.; notes: Bitbucket comment bodies require nested content objects; use reverse ETL writes once nested body mapping is modeled.
  Pipelines And Deployments Commands
    pipeline list - List pipelines [intent=etl availability=implemented stream=pipelines]
    pipeline view - View pipeline details [intent=direct_read availability=implemented]; flags: --pipeline-uuid
    pipeline run - Run a pipeline [intent=reverse_etl availability=implemented write=run_pipeline]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Starts a Bitbucket pipeline execution.
    pipeline stop - Stop a pipeline [intent=reverse_etl availability=implemented write=stop_pipeline]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Stops an in-flight Bitbucket pipeline.; flags: --pipeline-uuid
    deployment list - List deployments [intent=etl availability=implemented stream=deployments]
  Workspace And Administration Commands
    workspace list - List accessible workspaces [intent=direct_read availability=implemented]
    project list - List workspace projects [intent=direct_read availability=implemented]
    webhook list - List repository webhooks [intent=etl availability=implemented stream=webhooks]
    webhook create - Create a repository webhook [intent=reverse_etl availability=implemented write=create_webhook]; approval: reverse ETL writes require plan, preview, approval, execute. URL policy review is required.; risk: Creates an outbound webhook that can send repository events to an external URL.; flags: --url, --event
    webhook delete - Delete a repository webhook [intent=reverse_etl availability=implemented write=delete_webhook]; approval: reverse ETL writes require plan, preview, approval, execute. Typed confirmation may be required by policy.; risk: Deletes an existing webhook and may interrupt downstream automation.; flags: --uid
    branch-restriction list - List branch restrictions [intent=etl availability=implemented stream=branch_restrictions]
    branch-restriction create - Create a branch restriction [intent=reverse_etl availability=implemented write=create_branch_restriction]; approval: reverse ETL writes require plan, preview, approval, execute. Typed confirmation may be required for admin policy changes.; risk: Changes repository branch protection behavior.; flags: --kind, --pattern
    branch-restriction delete - Delete a branch restriction [intent=reverse_etl availability=partial]; approval: reverse ETL writes require plan, preview, approval, execute. Typed confirmation is required for admin policy changes.; risk: Removes branch protection behavior.; notes: Blocked until branch restriction delete confirmation UX is modeled.
    snippet list - List snippets [intent=direct_read availability=implemented]
    snippet create - Create a snippet [intent=reverse_etl availability=implemented write=create_snippet]; approval: reverse ETL writes require plan, preview, approval, execute. Content redaction review is required.; risk: Creates a Bitbucket snippet and may publish code or text content.; flags: --title, --content, --is-private
  Local Workflow Commands
    auth status - Show credential status [intent=auth availability=unsupported_local unsupported local workflow]; notes: Use `pm credentials inspect <name> --redacted`; this metadata does not read secrets.
    config view - Show Bitbucket command configuration [intent=config availability=unsupported_local unsupported local workflow]; notes: Connector-local CLI configuration is not implemented.
    api - Call an arbitrary Bitbucket API endpoint [intent=raw_api availability=unsafe_or_disallowed]; notes: Generic raw API calls are forbidden. Add reviewed direct-read or reverse-ETL operations instead.
  Help topics:
    safety - Bitbucket writes remain plan, preview, approval, execute; generic raw API writes are disallowed.
    coverage - All 331 official Swagger operations are enumerated in api_surface.json and operations.json; only reviewed app intents execute.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect bitbucket

  # Inspect as structured JSON
  pm connectors inspect bitbucket --json

AGENT WORKFLOW
  - Run pm connectors inspect bitbucket before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
