# pm connectors inspect google-classroom

```text
NAME
  pm connectors inspect google-classroom - Google Classroom connector manual

SYNOPSIS
  pm connectors inspect google-classroom
  pm connectors inspect google-classroom --json
  pm credentials add <name> --connector google-classroom [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Google Classroom courses, teachers, students, course work, and announcements through the Classroom REST API using an OAuth2 refresh token. In architecture v2 this quarantine bundle dispatches live reads through a Tier-2 hook that delegates to the legacy connector until the wave 6 cutover.

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
  base_url
  mode
  client_id (secret) (required)
  client_refresh_token (secret) (required)
  client_secret (secret) (required)

ETL STREAMS
  courses:
    primary key: id
    cursor: updateTime
    fields: alternateLink(string), calendarId(string), courseGroupEmail(string), courseState(string), creationTime(string), description(string), descriptionHeading(string), enrollmentCode(string), guardiansEnabled(boolean), id(string), name(string), ownerId(string), room(string), section(string), teacherGroupEmail(string), updateTime(string)
  teachers:
    primary key: courseId, userId
    fields: courseId(string), emailAddress(string), fullName(string), photoUrl(string), userId(string)
  students:
    primary key: courseId, userId
    fields: courseId(string), emailAddress(string), fullName(string), photoUrl(string), userId(string)
  course_work:
    primary key: id
    cursor: updateTime
    fields: alternateLink(string), courseId(string), creationTime(string), description(string), dueDate(string), id(string), maxPoints(number), state(string), title(string), updateTime(string), workType(string)
  announcements:
    primary key: id
    cursor: updateTime
    fields: alternateLink(string), courseId(string), creationTime(string), creatorUserId(string), id(string), state(string), text(string), updateTime(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

SECURITY
  read risk: external Google Classroom API reads performed by the legacy connector via a Tier-2 hook
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
