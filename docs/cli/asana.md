# pm asana

```text
NAME
  pm asana - Inspect and safely plan changes to Asana workspaces, projects, and tasks.

SYNOPSIS
  pm asana <topic> <leaf> [flags]

DESCRIPTION
  Inspect and safely plan changes to Asana workspaces, projects, and tasks.

COMMANDS
  Organization
    workspaces list - List available Asana workspaces. [intent=etl availability=implemented stream=workspaces]
    users list - List Asana users, optionally scoped to a workspace. [intent=etl availability=implemented stream=users]
    teams list - List teams in an Asana workspace. [intent=etl availability=implemented stream=teams]
    team-memberships list - List Asana team memberships. [intent=etl availability=implemented stream=team_memberships]
    workspace-memberships list - List memberships in an Asana workspace. [intent=etl availability=implemented stream=workspace_memberships]
  Work Management
    projects list - List Asana projects, optionally scoped to a workspace. [intent=etl availability=implemented stream=projects]
    projects create - Plan creation of an Asana project. [intent=reverse_etl availability=implemented write=create_project]; risk: external mutation; creates a new project visible to the team/workspace; approval required; approval: Requires plan, preview, explicit approval, then execute.
    projects update - Plan updates to an Asana project. [intent=reverse_etl availability=implemented write=update_project]; risk: external mutation; overwrites project fields (can archive, reassign owner); approval required; approval: Requires plan, preview, explicit approval, then execute.
    projects delete - Plan deletion of an Asana project. [intent=reverse_etl availability=implemented write=delete_project]; risk: irreversible external deletion of a project and its association with its tasks; approval required; approval: Requires plan, preview, explicit approval, then execute.
    project-statuses list - List statuses for discovered Asana projects. [intent=etl availability=implemented stream=project_statuses]
    sections list - List sections for discovered Asana projects. [intent=etl availability=implemented stream=sections]
    sections create - Plan creation of an Asana project section. [intent=reverse_etl availability=implemented write=create_section]; risk: external mutation; creates a new section in a project's board/list view; approval required; approval: Requires plan, preview, explicit approval, then execute.
    sections update - Plan updates to an Asana section. [intent=reverse_etl availability=implemented write=update_section]; risk: external mutation; renames a section; approval required; approval: Requires plan, preview, explicit approval, then execute.
    sections delete - Plan deletion of an empty Asana section. [intent=reverse_etl availability=implemented write=delete_section]; risk: irreversible external deletion of a section (Asana requires the section be empty of tasks first); approval required; approval: Requires plan, preview, explicit approval, then execute.
    tasks list - List Asana tasks with optional workspace, project, and assignee scopes. [intent=etl availability=implemented stream=tasks]
    tasks create - Plan creation of an Asana task. [intent=reverse_etl availability=implemented write=create_task]; risk: external mutation; creates a new task visible to the whole team/project; approval required; approval: Requires plan, preview, explicit approval, then execute.
    tasks update - Plan updates to an Asana task. [intent=reverse_etl availability=implemented write=update_task]; risk: external mutation; overwrites task fields (e.g. can mark completed, reassign, reschedule); approval required; approval: Requires plan, preview, explicit approval, then execute.
    tasks delete - Plan deletion of an Asana task. [intent=reverse_etl availability=implemented write=delete_task]; risk: irreversible external deletion of a task; approval required; approval: Requires plan, preview, explicit approval, then execute.
    tasks comment - Plan an Asana task comment. [intent=reverse_etl availability=implemented write=add_comment]; risk: external mutation; posts a comment visible to everyone with access to the task; approval required; approval: Requires plan, preview, explicit approval, then execute.
    stories list - List stories for discovered Asana tasks. [intent=etl availability=implemented stream=stories]
  Metadata
    tags list - List Asana tags, optionally scoped to a workspace. [intent=etl availability=implemented stream=tags]
    tags create - Plan creation of an Asana tag. [intent=reverse_etl availability=implemented write=create_tag]; risk: external mutation; creates a new workspace-visible tag; approval required; approval: Requires plan, preview, explicit approval, then execute.
    tags update - Plan updates to an Asana tag. [intent=reverse_etl availability=implemented write=update_tag]; risk: external mutation; renames/recolors a tag visible workspace-wide; approval required; approval: Requires plan, preview, explicit approval, then execute.
    tags delete - Plan deletion of an Asana tag. [intent=reverse_etl availability=implemented write=delete_tag]; risk: irreversible external deletion of a tag, removed from every task that carries it; approval required; approval: Requires plan, preview, explicit approval, then execute.
    custom-fields list - List custom fields in an Asana workspace. [intent=etl availability=implemented stream=custom_fields]

GLOBAL FLAGS
  --json - Write deterministic machine-readable JSON output.

SAFETY
  Help is generated from validated connector metadata. It does not read credentials, open project state, contact connector APIs, or execute commands.
  Mutation commands retain their declared plan, preview, approval, and execution policy.

EXIT STATUS
  0  Help rendered successfully.
  2  The connector help path is invalid.
```
