---
name: pm-gitlab
description: GitLab connector knowledge and safe action guide.
---

# pm-gitlab

## Purpose

Reads GitLab projects, groups, users, and issues through the GitLab REST API v4.

## Icon

- asset: icons/gitlab.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://docs.gitlab.com/ee/api/rest/deprecations.html

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- page_size
- start_date
- access_token (secret)

## ETL Streams

- projects:
  - primary key: id
  - cursor: last_activity_at
  - fields: archived(), created_at(), default_branch(), description(), forks_count(), id(), last_activity_at(), name(), open_issues_count(), path(), path_with_namespace(), star_count(), visibility(), web_url()
- groups:
  - primary key: id
  - cursor: created_at
  - fields: created_at(), description(), full_name(), full_path(), id(), name(), parent_id(), path(), visibility(), web_url()
- users:
  - primary key: id
  - cursor: created_at
  - fields: bot(), created_at(), id(), is_admin(), name(), state(), username(), web_url()
- issues:
  - primary key: id
  - cursor: updated_at
  - fields: author_id(), closed_at(), created_at(), downvotes(), id(), iid(), project_id(), state(), title(), updated_at(), upvotes(), user_notes_count(), web_url()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external GitLab API read of projects, groups, users, and issues
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Work with GitLab projects from the command line.
- Usage: pm gitlab <command> <subcommand> [flags]
- Source CLI: glab (https://gitlab.com/gitlab-org/cli/-/tree/main/docs/source)
- Global flags:
  - --json (boolean): Write machine-readable JSON output.
  - --connection (string): Use a saved GitLab connector credential and base URL scope: maps_to=connection
  - --limit (integer): Limit records emitted by stream-backed commands: maps_to=limit
- Core Commands
  - project list - List visible GitLab projects [intent=etl availability=implemented stream=projects]; flags: --search, --owned
  - project view - View one GitLab project [intent=direct_read availability=implemented]; notes: Bounded direct read of one project; response is recursively redacted before output.; flags: --id
  - group list - List visible GitLab groups [intent=etl availability=implemented stream=groups]; flags: --search
  - group view - View one GitLab group [intent=direct_read availability=implemented]; notes: Bounded direct read of one group; response is recursively redacted before output.; flags: --id
  - user list - List GitLab users visible to the token [intent=etl availability=implemented stream=users]; flags: --search, --username
  - user events - View events for a GitLab user [intent=direct_read availability=implemented]; notes: Bounded direct read of a user event page; response is recursively redacted before output.; flags: --id
  - issue list - List GitLab issues visible to the token [intent=etl availability=implemented stream=issues]; flags: --state, --assignee-username, --label
  - issue view - View issue details [intent=direct_read availability=implemented]; notes: Bounded direct read of one project issue; response is recursively redacted before output.; flags: --project-id, --issue-iid
  - issue create - Create an issue [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would create a visible issue in a GitLab project.; notes: No GitLab write action is declared yet.
  - issue update - Update an issue [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would mutate issue title, description, labels, assignees, or state.; notes: No GitLab write action is declared yet.
  - issue close - Close an issue [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would change issue state in a GitLab project.; notes: No GitLab write action is declared yet.
  - issue reopen - Reopen an issue [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would change issue state in a GitLab project.; notes: No GitLab write action is declared yet.
  - issue delete - Delete an issue [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; destructive writes need an explicit policy and typed confirmation before dispatch.; risk: Deletes project data and may be irreversible.; notes: Destructive issue deletion is not exposed by this metadata slice.
  - issue note - Add a note to an issue [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would create a visible note on an issue.; notes: No GitLab write action is declared yet.
  - mr list - List merge requests [intent=etl availability=planned]; notes: Merge request stream coverage belongs to a future stream expansion or operation-ledger lane.
  - mr view - View merge request details [intent=direct_read availability=planned]; notes: Single merge-request lookup requires bounded direct-read support.
  - mr create - Create a merge request [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would create a visible merge request in a project.; notes: No GitLab write action is declared yet.
  - mr update - Update a merge request [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would mutate merge request metadata or state.; notes: No GitLab write action is declared yet.
  - mr merge - Merge a merge request [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; destructive or deployment-adjacent writes need explicit policy and typed confirmation.; risk: Merges code into the target branch and can trigger CI/CD or deployments.; notes: Not exposed until sensitive/admin policy and typed confirmation are implemented.
  - mr approve - Approve a merge request [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would record a review approval on a merge request.; notes: No GitLab write action is declared yet.
  - repo clone - Clone a GitLab repository locally [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Requires a constrained local git executor and destination path policy; not a connector API dispatch.
  - repo archive - Download a repository archive [intent=direct_read availability=planned]; notes: Binary archive downloads require explicit size and output-path policy before enabling.
  - repo create - Create a GitLab project [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would create a project and allocate namespace resources.; notes: No GitLab write action is declared yet.
  - repo update - Update GitLab project settings [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would mutate project settings.; notes: No GitLab write action is declared yet.
  - repo delete - Delete a GitLab project [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; destructive admin writes need explicit policy and typed confirmation.; risk: Deletes a project and its repository data.; notes: Repository deletion is intentionally not exposed.
  - repo transfer - Transfer a GitLab project to another namespace [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; admin writes need explicit policy and typed confirmation.; risk: Changes project ownership and namespace access boundaries.; notes: Not exposed by this metadata slice.
- CI/CD Commands
  - pipeline list - List CI/CD pipelines [intent=etl availability=planned]; notes: Pipelines are ETL candidates but are not current streams.
  - pipeline view - View one CI/CD pipeline [intent=direct_read availability=planned]; notes: Pipeline detail reads require direct-read operation metadata.
  - pipeline run - Run a CI/CD pipeline [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; execution requires plan, preview, approval, and typed confirmation policy.; risk: Starts CI/CD execution and may deploy or mutate environments.; notes: No pipeline-trigger write action is exposed by this metadata slice.
  - pipeline cancel - Cancel a CI/CD pipeline [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would stop CI/CD execution for a project pipeline.; notes: No GitLab write action is declared yet.
  - job view - View a CI/CD job [intent=direct_read availability=planned]; notes: Job reads are bounded direct-read candidates.
  - job artifact download - Download CI/CD job artifacts [intent=direct_read availability=planned]; notes: Binary downloads require explicit size limits and output destination policy before enabling.
  - schedule list - List pipeline schedules [intent=etl availability=planned]; notes: Schedule stream coverage is deferred.
  - schedule run - Run a pipeline schedule [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; execution requires plan, preview, approval, and typed confirmation policy.; risk: Starts scheduled CI/CD execution and may deploy or mutate environments.; notes: Not exposed by this metadata slice.
  - runner list - List GitLab runners [intent=etl availability=planned]; notes: Runner inventory often requires elevated scope and is deferred to the operation ledger.
  - runner register - Register a GitLab runner [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; admin infrastructure writes need explicit policy and typed confirmation.; risk: Changes CI execution infrastructure and may introduce privileged compute.; notes: Not exposed by this metadata slice.
- Collaboration Commands
  - label list - List labels [intent=etl availability=planned]; notes: Label stream coverage is deferred.
  - label create - Create a label [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would create a label in a project or group.; notes: No GitLab write action is declared yet.
  - label delete - Delete a label [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; destructive writes need explicit policy and typed confirmation.; risk: Deletes a label and may affect issue or merge request workflows.; notes: Not exposed by this metadata slice.
  - milestone list - List milestones [intent=etl availability=planned]; notes: Milestone stream coverage is deferred.
  - release list - List releases [intent=etl availability=planned]; notes: Release stream coverage is deferred.
  - release create - Create a release [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would publish a release in a GitLab project.; notes: No GitLab write action is declared yet.
  - release download - Download release assets [intent=direct_read availability=planned]; notes: Binary asset downloads require explicit size limits and output destination policy before enabling.
  - snippet list - List snippets [intent=etl availability=planned]; notes: Snippet streams are deferred to the operation ledger.
  - snippet create - Create a snippet [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would create project or personal snippet content.; notes: No GitLab write action is declared yet.
  - todo list - List to-do items [intent=direct_read availability=planned]; notes: To-do items are user-scoped direct-read candidates.
  - todo done - Mark a to-do item done [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would mutate a user-scoped to-do item.; notes: No GitLab write action is declared yet.
- Security And Administration Commands
  - variable list - List CI/CD variables [intent=direct_read availability=planned]; risk: Variable metadata may be sensitive even when values are masked or hidden.; notes: Requires sensitive-field redaction policy before enabling.
  - variable set - Create or update a CI/CD variable [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; sensitive writes require stdin/env input, preview redaction, approval, and typed confirmation.; risk: Writes secret or deployment-affecting configuration.; notes: Never request or store variable values in prompts or metadata.
  - variable delete - Delete a CI/CD variable [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; destructive sensitive writes require explicit policy and typed confirmation.; risk: Deletes secret or deployment-affecting configuration.; notes: Not exposed by this metadata slice.
  - deploy-key list - List deploy keys [intent=direct_read availability=planned]; risk: Deploy key metadata can reveal access configuration.; notes: Requires operation-ledger classification and redaction policy before enabling.
  - deploy-key add - Add a deploy key [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; access-control writes require explicit policy and typed confirmation.; risk: Grants repository access to a key.; notes: Not exposed by this metadata slice.
  - ssh-key list - List account SSH keys [intent=direct_read availability=planned]; risk: Account key metadata can be sensitive.; notes: Requires redaction and account-scope policy before enabling.
  - ssh-key add - Add an account SSH key [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; access-control writes require explicit policy and typed confirmation.; risk: Grants account access to a key.; notes: Not exposed by this metadata slice.
  - gpg-key list - List account GPG keys [intent=direct_read availability=planned]; notes: Account-scoped key metadata requires direct-read policy before enabling.
  - token list - List personal, project, or group tokens [intent=direct_read availability=unsafe_or_disallowed]; risk: Token metadata is sensitive and may require elevated scope.; notes: Blocked until sensitive/admin policy defines risk tiers and redaction.
  - token rotate - Rotate a token [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; credential lifecycle writes require human-approved policy.; risk: Mutates credentials and can break automation or reveal newly issued secret material.; notes: Not exposed by this metadata slice.
  - securefile list - List secure files [intent=direct_read availability=unsafe_or_disallowed]; risk: Secure files may contain signing keys or other sensitive binary content.; notes: Blocked until sensitive binary-output policy exists.
  - securefile download - Download secure files [intent=direct_read availability=unsafe_or_disallowed]; risk: Downloads sensitive binary material to local storage.; notes: Blocked until bounded executor, destination policy, and approval policy exist.
  - container-registry list - List container registry repositories or tags [intent=etl availability=planned]; notes: Registry inventory is an ETL candidate, deferred to operation ledger classification.
  - packages list - List package registry packages [intent=etl availability=planned]; notes: Package registry inventory is an ETL candidate, deferred to operation ledger classification.
- Local Workflow Commands
  - auth login - Authenticate glab locally [intent=auth availability=excluded]; notes: Polymetrics credentials are managed through pm credential flows and never through prompt text.
  - config set - Set local glab configuration [intent=config availability=unsupported_local unsupported local workflow]; notes: Local glab configuration is outside the GitLab connector.
  - alias set - Configure a local glab alias [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Local alias configuration is outside connector execution.
  - completion - Generate shell completion scripts [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Shell completion generation is handled by the pm CLI, not by the GitLab connector surface.
- Additional Commands
  - api - Make an arbitrary GitLab API request [intent=raw_api availability=unsafe_or_disallowed]; approval: Not exposed. Add typed operations instead of raw API access.; risk: Arbitrary API dispatch can bypass connector safety, approval, redaction, and operation-ledger classification.; notes: Generic raw HTTP reads/writes are intentionally disallowed.
  - search code - Search code or project resources [intent=direct_read availability=planned]; notes: Search is a bounded direct-read candidate and must not become a raw API escape hatch.
  - changelog generate - Generate changelogs from project history [intent=direct_read availability=planned]; notes: Changelog generation combines reads and local formatting; it needs an explicit bounded workflow before exposure.
  - cluster list - List GitLab Agents for Kubernetes clusters [intent=direct_read availability=planned]; risk: Cluster metadata can be sensitive and may require elevated scope.; notes: Deferred to operation-ledger and sensitive/admin policy lanes.
  - duo prompt - Interact with GitLab Duo [intent=docs_only availability=excluded]; notes: Interactive AI assistant behavior is outside connector ETL/reverse ETL scope.
  - mcp serve - Run the glab MCP server [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Local server processes are outside connector command dispatch.
  - opentofu state - Work with OpenTofu or Terraform integration state [intent=direct_read availability=planned]; risk: Infrastructure state can contain sensitive data and environment topology.; notes: Requires explicit operation classification and redaction policy before enabling.
  - stack create - Create or manage stacked diffs [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Stacked-diff workflow requires local git operations and is outside connector API dispatch.
- Help topics:
  - authentication - Use saved Polymetrics credentials or environment/stdin-based credential loading; never pass token values in prompts.
  - writes - GitLab writes are planned only and must use reverse ETL plan, preview, approval, execute before dispatch.
  - binary-downloads - Artifacts, archives, secure files, and release assets remain disabled until bounded size and output destination policies exist.
  - raw-api - Generic raw API commands are intentionally disallowed; use typed streams, direct reads, or write actions instead.

## Commands

### Inspect as a manual

```bash
pm connectors inspect gitlab
```

### Inspect as structured JSON

```bash
pm connectors inspect gitlab --json
```

## Agent Rules

- Run pm connectors inspect gitlab before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
