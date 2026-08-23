---
name: pm-sonar-cloud
description: SonarCloud connector knowledge and safe action guide.
---

# pm-sonar-cloud

## Purpose

Reads SonarCloud issues, components, projects, hotspots, rules, metrics, languages, quality gates, measures, webhooks, and project analyses through the Web API; writes webhook lifecycle, issue comment/assign/tag/transition, and project-tag mutations.

## Icon

- id: sonarcloud
- asset: icons/sonarcloud.svg
- source: upstream_registry
- review_status: upstream_seeded
- review_url: https://sonarcloud.io/web_api

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- component_keys
- end_date
- mode
- organization
- page_size
- start_date
- user_token (secret) (required)

## ETL Streams

- issues:
  - primary key: key
  - cursor: createdAt
  - fields: author(string), component(string), createdAt(string), creationDate(string), key(string), line(integer), message(string), organization(string), project(string), rule(string), severity(string), status(string), tags(array), type(string), updateDate(string)
- components:
  - primary key: key
  - cursor: createdAt
  - fields: createdAt(string), key(string), name(string), organization(string), project(string), qualifier(string), visibility(string)
- quality_gates:
  - primary key: id
  - cursor: createdAt
  - fields: createdAt(string), id(string), isBuiltIn(boolean), isDefault(boolean), name(string)
- measures:
  - primary key: metric
  - cursor: createdAt
  - fields: bestValue(boolean), component(string), createdAt(string), metric(string), value(string)
- projects:
  - primary key: key
  - fields: key(string), lastAnalysisDate(string), name(string), organization(string), qualifier(string), revision(string), visibility(string)
- hotspots:
  - primary key: key
  - fields: assignee(string), author(string), component(string), creationDate(string), key(string), line(integer), message(string), project(string), securityCategory(string), status(string), updateDate(string), vulnerabilityProbability(string)
- languages:
  - primary key: key
  - fields: key(string), name(string)
- metrics:
  - primary key: key
  - fields: custom(boolean), description(string), direction(integer), domain(string), hidden(boolean), id(string), key(string), name(string), qualitative(boolean), type(string)
- rules:
  - primary key: key
  - fields: createdAt(string), isExternal(boolean), isTemplate(boolean), key(string), lang(string), langName(string), name(string), repo(string), severity(string), status(string), tags(array), type(string), updatedAt(string)
- webhooks:
  - primary key: key
  - fields: hasSecret(boolean), key(string), name(string), url(string)
- project_analyses:
  - primary key: key
  - fields: buildString(string), date(string), events(array), key(string), projectVersion(string), revision(string)

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_webhook:
  - endpoint: POST /api/webhooks/create
  - required fields: name, organization, url
  - risk: external mutation; creates a project or organization webhook that will receive analysis-completion callbacks; approval required
- update_webhook:
  - endpoint: POST /api/webhooks/update
  - required fields: webhook, name, url
  - risk: external mutation; changes an existing webhook's callback URL/secret; approval required
- delete_webhook:
  - endpoint: POST /api/webhooks/delete
  - required fields: webhook
  - risk: external mutation; permanently removes a webhook; approval required
- add_issue_comment:
  - endpoint: POST /api/issues/add_comment
  - required fields: issue, text
  - risk: external mutation; adds a permanent comment to an issue; approval required
- assign_issue:
  - endpoint: POST /api/issues/assign
  - required fields: issue
  - risk: external mutation; assigns or unassigns (empty assignee) an issue; approval required
- set_issue_tags:
  - endpoint: POST /api/issues/set_tags
  - required fields: issue
  - risk: external mutation; replaces an issue's full tag set (empty tags clears them); approval required
- do_issue_transition:
  - endpoint: POST /api/issues/do_transition
  - required fields: issue, transition
  - risk: external mutation; moves an issue through its workflow (e.g. resolve, wontfix, falsepositive); some transitions require elevated project permissions on the live API; approval required
- set_project_tags:
  - endpoint: POST /api/project_tags/set
  - required fields: project, tags
  - risk: external mutation; replaces a project's full tag set; approval required

## Security

- read risk: external SonarCloud API read of issues, components, projects, hotspots, rules, metrics, languages, quality gates, measures, webhooks, and project analyses
- write risk: external SonarCloud API mutation of webhooks (create/update/delete), issue comments/assignment/tags/workflow transitions, and project tags
- approval: required for all write actions; each is an external, user-visible mutation on a connected SonarCloud organization or project
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect sonar-cloud
```

### Inspect as structured JSON

```bash
pm connectors inspect sonar-cloud --json
```

## Agent Rules

- Run pm connectors inspect sonar-cloud before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
