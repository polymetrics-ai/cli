# pm connectors inspect workramp

```text
NAME
  pm connectors inspect workramp - WorkRamp connector manual

SYNOPSIS
  pm connectors inspect workramp
  pm connectors inspect workramp --json
  pm credentials add <name> --connector workramp [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes WorkRamp users and groups, and reads guides, resources, and SCORM courses, through the real WorkRamp Employee Learning Cloud API (app.workramp.com/api/v1).

ICON
  asset: icons/workramp.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developers.workramp.com/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  max_pages
  mode
  page_size
  api_key (secret)

ETL STREAMS
  users:
    primary key: id
    cursor: updatedAt
    fields: createdAt(), email(), id(), isAdmin(), isDeleted(), isPermanentlyDeleted(), name(), updatedAt()
  groups:
    primary key: id
    cursor: updatedAt
    fields: createdAt(), description(), enterpriseId(), id(), name(), updatedAt()
  courses:
    primary key: id
    cursor: updated_at
    fields: created_at(), id(), num_total_tasks(), num_total_test_questions(), tags(), title(), updated_at()
  resources:
    primary key: id
    fields: createdAt(), description(), id(), name(), updatedAt()
  scorm_courses:
    primary key: id
    fields: created_at(), id(), time_estimate(), title()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_user:
    endpoint: POST /api/v1/users
    risk: creates a WorkRamp user account; approval required
  update_user:
    endpoint: POST /api/v1/users/{{ record.id }}
    required fields: id
    risk: updates a WorkRamp user account's attributes; approval required
  delete_user:
    endpoint: DELETE /api/v1/users/{{ record.id }}
    required fields: id
    risk: permanently deletes a WorkRamp user account; approval required
  create_group:
    endpoint: POST /api/v1/groups
    risk: creates a WorkRamp group; approval required
  update_group:
    endpoint: POST /api/v1/groups/{{ record.id }}
    required fields: id
    risk: updates a WorkRamp group's attributes; approval required

SECURITY
  read risk: external WorkRamp API read of user, group, guide, resource, and SCORM-course data
  write risk: external mutation of WorkRamp users and groups (create/update/delete); actions are attributed to the API key's owning admin user; approval required
  approval: writes require approval; reads are none
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect workramp

  # Inspect as structured JSON
  pm connectors inspect workramp --json

AGENT WORKFLOW
  - Run pm connectors inspect workramp before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
