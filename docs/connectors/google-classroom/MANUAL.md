# pm connectors inspect google-classroom

```text
NAME
  pm connectors inspect google-classroom - Google Classroom connector manual

SYNOPSIS
  pm connectors inspect google-classroom
  pm connectors inspect google-classroom --json
  pm credentials add <name> --connector google-classroom [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Classroom courses and course-scoped resources through fixed REST routes and OAuth2 refresh-token authentication.

ICON
  id: simple-icons-googleclassroom
  asset: icons/simple-icons/googleclassroom.svg
  title: Google Classroom
  simple_icon_slug: googleclassroom
  simple_icon_hex: 0F9D58
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Google%20Classroom
  match: exact-name-or-slug
  matched_by: google-classroom

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  client_id (secret) (required)
  client_refresh_token (secret) (required)
  client_secret (secret) (required)

ETL STREAMS
  courses:
    primary key: id
    cursor: updateTime
    fields: alternateLink(string), courseState(string), creationTime(string), description(string), id(string), name(string), ownerId(string), section(string), updateTime(string)
  teachers:
    primary key: courseId, userId
    fields: courseId(string), emailAddress(string), fullName(string), photoUrl(string), userId(string)
  students:
    primary key: courseId, userId
    fields: courseId(string), emailAddress(string), fullName(string), photoUrl(string), userId(string)
  course_work:
    primary key: id
    cursor: updateTime
    fields: alternateLink(string), courseId(string), creationTime(string), description(string), id(string), maxPoints(number), state(string), title(string), updateTime(string), workType(string)
  announcements:
    primary key: id
    cursor: updateTime
    fields: alternateLink(string), courseId(string), creationTime(string), creatorUserId(string), id(string), state(string), text(string), updateTime(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

SECURITY
  read risk: Bounded Classroom reads use fixed OAuth2 and REST routes.
  write risk: unsupported
  approval: none; read-only
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect google-classroom

  # Inspect as structured JSON
  pm connectors inspect google-classroom --json

AGENT WORKFLOW
  - Run pm connectors inspect google-classroom before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
