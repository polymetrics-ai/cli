# pm connectors inspect thrive-learning

```text
NAME
  pm connectors inspect thrive-learning - Thrive Learning connector manual

SYNOPSIS
  pm connectors inspect thrive-learning
  pm connectors inspect thrive-learning --json
  pm credentials add <name> --connector thrive-learning [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads users, content, completions, assignments, audiences, tags, CPD records, and activity data through the Thrive Learning Public API.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  start_date
  username (required)
  password (secret) (required)

ETL STREAMS
  users:
    primary key: id
    fields: created_at(string), email(string), id(string), name(string), updated_at(string)
  content:
    primary key: id
    fields: created_at(string), id(string), title(string), type(string), updated_at(string)
  completions:
    primary key: id
    fields: completed_at(string), content_id(string), id(string), updated_at(string), user_id(string)
  activities:
    primary key: id
    cursor: date
    fields: contextId(string), contextType(string), date(string), id(string), name(string), type(string), user(string)
  contents_v1:
    primary key: id
    cursor: updatedAt
    fields: createdAt(string), description(string), duration(number), id(string), isOfficial(boolean), tags(array), title(string), type(string), updatedAt(string)
  learning_completions:
    primary key: id
    fields: activeUntil(string), completedAt(string), completionType(string), contentId(string), contentVersion(number), hadDueDate(boolean), id(string), isRPL(boolean), skills(array), userId(string)
  assignments:
    primary key: id
    cursor: updatedAt
    fields: alternativeContentIds(array), audienceId(string), completionPeriod(object), createdAt(string), deletedAt(string), hideAlternativeContent(boolean), id(string), isActive(boolean), isDeleted(boolean), primaryContentId(string), recurrence(object), updatedAt(string)
  assignment_enrolments:
    primary key: id
    cursor: updatedAt
    fields: assignmentId(string), assignment_id(string), audienceId(string), availableDate(string), dueDate(string), id(string), lastCompletedAt(string), primaryContentId(string), status(string), updatedAt(string), userId(string)
  audiences:
    primary key: id
    cursor: updatedAt
    fields: apiControlled(boolean), category(string), createdAt(string), id(string), name(string), parent(string), reference(string), type(string), updatedAt(string)
  audience_members:
    primary key: audience_id, userId
    fields: audience_id(string), email(string), reference(string), userId(string)
  audience_managers:
    primary key: audience_id, userId
    fields: audience_id(string), email(string), permissions(object), reference(string), userId(string)
  tags:
    primary key: id
    fields: campaigns(array), contents(array), id(string), tag(string)
  cpd_categories:
    primary key: categoryId
    fields: categoryId(string), name(string)
  cpd_entries:
    primary key: logEntryId
    fields: activity(string), category(string), description(string), durationMinutes(number), entryDate(string), isVerified(boolean), logEntryId(string), userId(string)
  cpd_requirements:
    primary key: audienceRequirementId
    fields: audienceId(string), audienceRequirementId(string), createdAt(string), requiredMinutes(integer)
  skill_levels:
    primary key: value
    fields: isEnabled(boolean), name(string), value(number)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: external Thrive Learning API read of user, content, completion, assignment, audience, tag, and CPD data
  approval: none; read-only, no dialect-expressible write path could be safely execution-contract-verified for this connector (see docs.md Known limits' write-actions ENGINE_GAP)
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect thrive-learning

  # Inspect as structured JSON
  pm connectors inspect thrive-learning --json

AGENT WORKFLOW
  - Run pm connectors inspect thrive-learning before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
