---
name: pm-thrive-learning
description: Thrive Learning connector knowledge and safe action guide.
---

# pm-thrive-learning

## Purpose

Reads users, content, completions, assignments, audiences, tags, CPD records, and activity data through the Thrive Learning Public API.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=false query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- start_date
- username
- password (secret)

## ETL Streams

- users:
  - primary key: id
  - fields: created_at(), email(), id(), name(), updated_at()
- content:
  - primary key: id
  - fields: created_at(), id(), title(), type(), updated_at()
- completions:
  - primary key: id
  - fields: completed_at(), content_id(), id(), updated_at(), user_id()
- activities:
  - primary key: id
  - cursor: date
  - fields: contextId(), contextType(), date(), id(), name(), type(), user()
- contents_v1:
  - primary key: id
  - cursor: updatedAt
  - fields: createdAt(), description(), duration(), id(), isOfficial(), tags(), title(), type(), updatedAt()
- learning_completions:
  - primary key: id
  - fields: activeUntil(), completedAt(), completionType(), contentId(), contentVersion(), hadDueDate(), id(), isRPL(), skills(), userId()
- assignments:
  - primary key: id
  - cursor: updatedAt
  - fields: alternativeContentIds(), audienceId(), completionPeriod(), createdAt(), deletedAt(), hideAlternativeContent(), id(), isActive(), isDeleted(), primaryContentId(), recurrence(), updatedAt()
- assignment_enrolments:
  - primary key: id
  - cursor: updatedAt
  - fields: assignmentId(), assignment_id(), audienceId(), availableDate(), dueDate(), id(), lastCompletedAt(), primaryContentId(), status(), updatedAt(), userId()
- audiences:
  - primary key: id
  - cursor: updatedAt
  - fields: apiControlled(), category(), createdAt(), id(), name(), parent(), reference(), type(), updatedAt()
- audience_members:
  - primary key: audience_id, userId
  - fields: audience_id(), email(), reference(), userId()
- audience_managers:
  - primary key: audience_id, userId
  - fields: audience_id(), email(), permissions(), reference(), userId()
- tags:
  - primary key: id
  - fields: campaigns(), contents(), id(), tag()
- cpd_categories:
  - primary key: categoryId
  - fields: categoryId(), name()
- cpd_entries:
  - primary key: logEntryId
  - fields: activity(), category(), description(), durationMinutes(), entryDate(), isVerified(), logEntryId(), userId()
- cpd_requirements:
  - primary key: audienceRequirementId
  - fields: audienceId(), audienceRequirementId(), createdAt(), requiredMinutes()
- skill_levels:
  - primary key: value
  - fields: isEnabled(), name(), value()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Security

- read risk: external Thrive Learning API read of user, content, completion, assignment, audience, tag, and CPD data
- approval: none; read-only, no dialect-expressible write path could be safely conformance-verified for this connector (see docs.md Known limits' write-actions ENGINE_GAP)
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect thrive-learning
```

### Inspect as structured JSON

```bash
pm connectors inspect thrive-learning --json
```

## Agent Rules

- Run pm connectors inspect thrive-learning before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
