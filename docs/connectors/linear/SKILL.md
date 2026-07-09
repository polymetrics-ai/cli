---
name: pm-linear
description: Linear connector knowledge and safe action guide.
---

# pm-linear

## Purpose

Reads and writes Linear issues, teams, projects, users, and approved common mutations through fixed Linear GraphQL operations.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- auth_type
- base_url
- max_pages
- access_token (secret)
- api_key (secret)

## ETL Streams

- issues:
  - primary key: id
  - cursor: updated_at
  - fields: assignee_email(), assignee_id(), branch_name(), canceled_at(), completed_at(), createdAt(), created_at(), creator_id(), description(), estimate(), id(), identifier(), priority(), state_id(), state_name(), state_type(), team_id(), team_key(), title(), updatedAt(), updated_at(), url()
- teams:
  - primary key: id
  - cursor: updated_at
  - fields: createdAt(), created_at(), description(), id(), key(), name(), private(), updatedAt(), updated_at()
- projects:
  - primary key: id
  - cursor: updated_at
  - fields: canceled_at(), completed_at(), createdAt(), created_at(), description(), id(), name(), progress(), started_at(), state(), updatedAt(), updated_at(), url()
- users:
  - primary key: id
  - cursor: updated_at
  - fields: active(), admin(), createdAt(), created_at(), display_name(), email(), id(), name(), updatedAt(), updated_at()
- issue:
  - primary key: id
  - cursor: updated_at
  - fields: assignee_email(), assignee_id(), branch_name(), canceled_at(), completed_at(), createdAt(), created_at(), creator_id(), description(), estimate(), id(), identifier(), priority(), state_id(), state_name(), state_type(), team_id(), team_key(), title(), updatedAt(), updated_at(), url()
- team:
  - primary key: id
  - cursor: updated_at
  - fields: createdAt(), created_at(), description(), id(), key(), name(), private(), updatedAt(), updated_at()
- project:
  - primary key: id
  - cursor: updated_at
  - fields: canceled_at(), completed_at(), createdAt(), created_at(), description(), id(), name(), progress(), started_at(), state(), updatedAt(), updated_at(), url()
- user:
  - primary key: id
  - cursor: updated_at
  - fields: active(), admin(), createdAt(), created_at(), display_name(), email(), id(), name(), updatedAt(), updated_at()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

## Reverse ETL Actions

- create_issue:
  - endpoint: POST /graphql
  - risk: Creates a visible Linear issue in the configured workspace.
- update_issue:
  - endpoint: POST /graphql
  - risk: Mutates an existing Linear issue.
- comment_issue:
  - endpoint: POST /graphql
  - risk: Creates a visible comment on a Linear issue.
- create_project:
  - endpoint: POST /graphql
  - risk: Creates a visible Linear project.

## Security

- read risk: external Linear GraphQL API read of approved fixed documents
- write risk: approved Linear GraphQL mutations through reverse ETL plan, preview, approval, execute
- approval: writes require connector command plan/preview/approval; sensitive/admin/destructive operations remain blocked by default
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Command Surface

- Work with Linear issues, teams, projects, and users from the command line.
- Usage: pm linear <command> <subcommand> [flags]
- Source CLI: Linear app and GraphQL API (https://developers.linear.app/docs/graphql/working-with-the-graphql-api)
- Global flags:
  - --json (boolean): Write machine-readable JSON output.
  - --connection (string): Use a saved Linear connector credential.: maps_to=connection
- Core Linear Commands
  - issue list - List Linear issues through the implemented ETL stream. [intent=etl availability=implemented stream=issues]
  - issue view - View one Linear issue. [intent=direct_read availability=implemented stream=issue]; flags: --issue-id
  - issue create - Create a Linear issue. [intent=reverse_etl availability=implemented write=create_issue]; approval: reverse ETL writes require plan, preview, approval token, execute, and typed confirmation when sensitive or destructive.; risk: Creates a visible issue in the configured Linear workspace.; flags: --team-id, --title, --description, --assignee-id, --project-id, --state-id
  - issue update - Update a Linear issue. [intent=reverse_etl availability=implemented write=update_issue]; approval: reverse ETL writes require plan, preview, approval token, execute, and typed confirmation when sensitive or destructive.; risk: Mutates an existing Linear issue.; flags: --issue-id, --title, --description, --assignee-id, --project-id, --state-id
  - project list - List Linear projects through the implemented ETL stream. [intent=etl availability=implemented stream=projects]
  - project view - View one Linear project. [intent=direct_read availability=implemented stream=project]; flags: --project-id
  - project create - Create a Linear project. [intent=reverse_etl availability=implemented write=create_project]; approval: reverse ETL writes require plan, preview, approval token, execute, and typed confirmation when sensitive or destructive.; risk: Creates a visible project in Linear.; flags: --team-id, --name, --description
  - team list - List Linear teams through the implemented ETL stream. [intent=etl availability=implemented stream=teams]
  - team view - View one Linear team. [intent=direct_read availability=implemented stream=team]; flags: --team-id
  - user list - List Linear users through the implemented ETL stream. [intent=etl availability=implemented stream=users]
  - user view - View one Linear user. [intent=direct_read availability=implemented stream=user]; flags: --user-id
- Collaboration Commands
  - comment create - Create a comment on a Linear issue. [intent=reverse_etl availability=implemented write=comment_issue]; approval: reverse ETL writes require plan, preview, approval token, execute, and typed confirmation when sensitive or destructive.; risk: Creates a visible comment in Linear.; flags: --issue-id, --body
  - cycle list - List Linear cycles. [intent=etl availability=planned]; notes: Planned ETL candidate from the official GraphQL query surface; not part of the current four-stream baseline.
  - label list - List Linear issue labels. [intent=etl availability=planned]; notes: Planned ETL candidate from the official GraphQL query surface; not part of the current four-stream baseline.
  - workflow-state list - List Linear workflow states. [intent=etl availability=planned]; notes: Planned ETL candidate from the official GraphQL query surface; not part of the current four-stream baseline.
- Administrative And Integration Commands
  - workspace invite - Invite a user to a Linear workspace. [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default pending sensitive/admin policy, preview, approval token, and typed confirmation.; risk: Changes workspace membership and may grant access to private Linear data.; notes: Administrative membership mutation; not exposed as an executable command in this connector surface.
  - webhook create - Create a Linear webhook. [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default pending sensitive/admin policy, explicit destination allow-listing, preview, approval token, and typed confirmation.; risk: Creates an outbound integration endpoint and may expose workspace events to external systems.; notes: Administrative integration mutation; not exposed as an executable command in this connector surface.
  - webhook delete - Delete a Linear webhook. [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default pending sensitive/admin policy, preview, approval token, and typed confirmation.; risk: Deletes an integration endpoint and can disrupt external workflows.; notes: Destructive administrative mutation; not exposed as an executable command in this connector surface.
  - api graphql - Run an arbitrary Linear GraphQL operation. [intent=raw_api availability=unsafe_or_disallowed]; approval: Disallowed. Use fixed reviewed stream, direct-read, or reverse-ETL actions instead.; risk: A raw GraphQL surface could bypass stream/write review, redaction, and approval policy.; notes: Generic GraphQL query or mutation execution is intentionally not exposed.
  - auth login - Authenticate the Linear connector. [intent=auth availability=unsupported_local unsupported local workflow]; notes: Credentials are managed through Polymetrics credential commands and environment/stdin flows; secrets are never accepted in prompt text.
  - config set - Configure Linear command defaults. [intent=config availability=unsupported_local unsupported local workflow]; notes: Connector configuration is handled by saved connections and runtime flags, not provider-specific config mutation commands.
- Help topics:
  - authentication - Use `api_key` or `access_token` through saved credentials or environment/stdin flows; never put secrets in command text.
  - graphql-safety - Linear GraphQL operations are exposed only as reviewed fixed streams, direct reads, or reverse-ETL actions; Raw arbitrary GraphQL is disallowed.
  - writes - Implemented Linear mutations use fixed GraphQL documents and reverse ETL plan → preview → approval → execute; sensitive/admin/destructive mutations remain blocked by default.

## Commands

### Inspect as a manual

```bash
pm connectors inspect linear
```

### Inspect as structured JSON

```bash
pm connectors inspect linear --json
```

## Agent Rules

- Run pm connectors inspect linear before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
