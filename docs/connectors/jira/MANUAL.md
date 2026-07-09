# pm connectors inspect jira

```text
NAME
  pm connectors inspect jira - Jira connector manual

SYNOPSIS
  pm connectors inspect jira
  pm connectors inspect jira --json
  pm credentials add <name> --connector jira [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Jira issues, projects, and users through the Jira Cloud REST API v3 using HTTP Basic auth (email + API token). Read-only.

ICON
  asset: icons/jira.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.atlassian.com/changelog/#

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  email
  api_token (secret)

ETL STREAMS
  issues:
    primary key: id
    cursor: updated
    fields: assignee(), created(), id(), issuetype(), key(), priority(), project(), reporter(), self(), status(), summary(), updated()
  projects:
    primary key: id
    fields: id(), isPrivate(), key(), name(), projectTypeKey(), self(), simplified(), style()
  users:
    primary key: accountId
    fields: accountId(), accountType(), active(), displayName(), emailAddress(), self()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Jira Cloud API read of issue, project, and user data
  approval: none; read-only, no reverse-ETL writes
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Work with Jira Cloud issues, projects, and users from the command line.
  Usage: pm jira <command> [flags]
  Source CLI: jira (https://developer.atlassian.com/cloud/jira/platform/swagger-v3.v3.json)
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --connection (string): Use a saved Jira connector credential and site scope.: maps_to=connection
  Issue Commands
    issue list - List Jira issues [intent=etl availability=implemented stream=issues]; notes: Executes the existing Jira issues stream. Advanced JQL, field selection, and single-issue rendering are deferred to later direct-read or stream-runner slices.
    issue search - Search issues with Jira query parameters [intent=etl availability=partial stream=issues]; notes: The connector has an issue-search stream, but gh-style/Jira-style ad-hoc JQL flags are not yet exposed as command flags.
    issue view - View one Jira issue [intent=direct_read availability=planned]; notes: Requires a bounded direct-read executor for caller-supplied issue IDs.
    issue create - Create an issue [intent=reverse_etl availability=planned]; approval: Future reverse ETL writes must use plan, preview, approval, and execute.; risk: Creates visible Jira project data.; notes: No Jira write action is declared in this metadata slice.
    issue edit - Edit an issue [intent=reverse_etl availability=planned]; approval: Future reverse ETL writes must use plan, preview, approval, and execute.; risk: Mutates existing Jira issue fields.; notes: Blocked until #107 classifies the operation and #110 defines sensitive/admin policy.
    issue transition - Transition an issue through a Jira workflow [intent=reverse_etl availability=planned]; approval: Future reverse ETL writes must use plan, preview, approval, and execute.; risk: Changes issue workflow state and may trigger Jira automation.; notes: Workflow transitions require explicit typed approval and field validation before execution.
    issue delete - Delete an issue [intent=direct_write availability=unsafe_or_disallowed]; notes: Issue deletion is destructive and is blocked by default.
    comment list - List comments for an issue [intent=direct_read availability=planned]; notes: Requires per-issue fan-out or a bounded direct-read command for caller-supplied issue IDs.
    comment add - Add an issue comment [intent=reverse_etl availability=planned]; approval: Future reverse ETL writes must use plan, preview, approval, and execute.; risk: Adds visible comment text to a Jira issue.; notes: No comment write action is declared in this metadata slice.
    worklog list - List issue worklogs [intent=direct_read availability=planned]; notes: Requires per-issue fan-out or a bounded direct-read command for caller-supplied issue IDs.
    worklog add - Add issue worklog time [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Any future support must use plan, preview, approval, and execute with typed confirmation.; risk: Writes time-tracking data to Jira issues.; notes: Time-tracking writes are sensitive business records and remain blocked by default.
    attachment metadata - Read issue attachment metadata [intent=direct_read availability=planned]; notes: Structured attachment metadata is a candidate for a bounded direct-read or fan-out stream after the operation ledger slice.
    attachment download - Download attachment content [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Attachment content returns binary payloads and needs an explicit bounded file-output policy before enabling.
  Project Commands
    project list - List Jira projects [intent=etl availability=implemented stream=projects]
    project view - View one Jira project [intent=direct_read availability=planned]; notes: Requires a bounded direct-read executor for caller-supplied project keys or IDs.
    project create - Create a Jira project [intent=direct_write availability=unsafe_or_disallowed]; notes: Project creation is an administrative action and is blocked by default.
    project delete - Delete a Jira project [intent=direct_write availability=unsafe_or_disallowed]; notes: Project deletion is destructive admin and is not exposed.
    component list - List project components [intent=direct_read availability=planned]; notes: Project subresources need project-ID fan-out or bounded direct reads.
    version list - List project versions [intent=direct_read availability=planned]; notes: Project version reads need project-ID fan-out or bounded direct reads.
  User And Team Commands
    user list - List Jira users visible to the site [intent=etl availability=implemented stream=users]
    user view - View one Jira user [intent=direct_read availability=planned]; notes: Requires a bounded direct-read executor for caller-supplied account IDs.
    myself view - View the authenticated Jira user [intent=direct_read availability=planned]; notes: The endpoint is currently used as the connector health check and is not yet exposed as a command.
    group list - List Jira groups [intent=direct_read availability=planned]; notes: Group APIs require elevated permissions on many sites and need operation-ledger risk classification.
  Agile Commands
    board list - List Jira Software boards [intent=direct_read availability=planned]; notes: Jira Software Agile APIs are outside the current REST v3 legacy stream set and need a separate ledger row.
    sprint list - List board sprints [intent=direct_read availability=planned]; notes: Requires Agile API coverage and board-ID scoping.
    epic list - List epics [intent=direct_read availability=planned]; notes: Requires Agile API or JQL-backed operation design; not part of the current legacy streams.
  Administration Commands
    workflow list - List Jira workflows [intent=direct_read availability=planned]; notes: Workflow metadata commonly requires admin permissions and needs #107/#110 classification.
    workflow publish - Publish a workflow draft [intent=direct_write availability=unsafe_or_disallowed]; notes: Workflow mutations are administrative and blocked by default.
    field list - List Jira fields [intent=direct_read availability=planned]; notes: Field metadata is useful but outside the current issue/project/user legacy stream set.
    permission scheme list - List permission schemes [intent=direct_read availability=planned]; notes: Permission scheme APIs are admin-oriented and need elevated-scope classification.
    screen list - List Jira screens [intent=direct_read availability=planned]; notes: Screen configuration APIs are admin-oriented and need elevated-scope classification.
    webhook list - List Jira webhooks [intent=direct_read availability=planned]; notes: Webhook APIs require explicit permission/risk review before command exposure.
  Local Workflow Commands
    auth login - Configure Jira credentials [intent=auth availability=unsupported_local unsupported local workflow]; notes: Use `pm credentials add jira ...` with environment variables, stdin, or secret manager input; do not pass secrets in prompt text or examples.
    config set - Set local Jira CLI configuration [intent=config availability=unsupported_local unsupported local workflow]; notes: Connector metadata does not write local CLI config. Use Polymetrics connection and credential commands.
    api - Call an arbitrary Jira REST endpoint [intent=raw_api availability=unsafe_or_disallowed]; notes: Generic raw Jira API calls are intentionally not exposed. Use typed streams, direct reads, or reverse ETL actions only.
  Help topics:
    auth - Use Polymetrics credentials for Jira site URL, email, and API token; never place secrets in command examples.
    writes - Jira writes are metadata-only in this slice and remain blocked until reverse ETL policy and operation ledger slices are complete.
    attachments - Attachment content is binary and requires explicit output bounds before any download executor can be enabled.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect jira

  # Inspect as structured JSON
  pm connectors inspect jira --json

AGENT WORKFLOW
  - Run pm connectors inspect jira before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
