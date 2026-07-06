# pm connectors inspect sonar-cloud

```text
NAME
  pm connectors inspect sonar-cloud - SonarCloud connector manual

SYNOPSIS
  pm connectors inspect sonar-cloud
  pm connectors inspect sonar-cloud --json
  pm credentials add <name> --connector sonar-cloud [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads SonarCloud issues, components, projects, hotspots, rules, metrics, languages, quality gates, measures, webhooks, and project analyses through the Web API; writes webhook lifecycle, issue comment/assign/tag/transition, and project-tag mutations.

ICON
  asset: icons/sonarcloud.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://sonarcloud.io/web_api

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  component_keys
  end_date
  mode
  organization
  page_size
  start_date
  user_token (secret)

ETL STREAMS
  issues:
    primary key: key
    cursor: createdAt
    fields: author(), component(), createdAt(), creationDate(), key(), line(), message(), organization(), project(), rule(), severity(), status(), tags(), type(), updateDate()
  components:
    primary key: key
    cursor: createdAt
    fields: createdAt(), key(), name(), organization(), project(), qualifier(), visibility()
  quality_gates:
    primary key: id
    cursor: createdAt
    fields: createdAt(), id(), isBuiltIn(), isDefault(), name()
  measures:
    primary key: metric
    cursor: createdAt
    fields: bestValue(), component(), createdAt(), metric(), value()
  projects:
    primary key: key
    fields: key(), lastAnalysisDate(), name(), organization(), qualifier(), revision(), visibility()
  hotspots:
    primary key: key
    fields: assignee(), author(), component(), creationDate(), key(), line(), message(), project(), securityCategory(), status(), updateDate(), vulnerabilityProbability()
  languages:
    primary key: key
    fields: key(), name()
  metrics:
    primary key: key
    fields: custom(), description(), direction(), domain(), hidden(), id(), key(), name(), qualitative(), type()
  rules:
    primary key: key
    fields: createdAt(), isExternal(), isTemplate(), key(), lang(), langName(), name(), repo(), severity(), status(), tags(), type(), updatedAt()
  webhooks:
    primary key: key
    fields: hasSecret(), key(), name(), url()
  project_analyses:
    primary key: key
    fields: buildString(), date(), events(), key(), projectVersion(), revision()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_webhook:
    endpoint: POST /api/webhooks/create
    risk: external mutation; creates a project or organization webhook that will receive analysis-completion callbacks; approval required
  update_webhook:
    endpoint: POST /api/webhooks/update
    risk: external mutation; changes an existing webhook's callback URL/secret; approval required
  delete_webhook:
    endpoint: POST /api/webhooks/delete
    risk: external mutation; permanently removes a webhook; approval required
  add_issue_comment:
    endpoint: POST /api/issues/add_comment
    risk: external mutation; adds a permanent comment to an issue; approval required
  assign_issue:
    endpoint: POST /api/issues/assign
    risk: external mutation; assigns or unassigns (empty assignee) an issue; approval required
  set_issue_tags:
    endpoint: POST /api/issues/set_tags
    risk: external mutation; replaces an issue's full tag set (empty tags clears them); approval required
  do_issue_transition:
    endpoint: POST /api/issues/do_transition
    risk: external mutation; moves an issue through its workflow (e.g. resolve, wontfix, falsepositive); some transitions require elevated project permissions on the live API; approval required
  set_project_tags:
    endpoint: POST /api/project_tags/set
    risk: external mutation; replaces a project's full tag set; approval required

SECURITY
  read risk: external SonarCloud API read of issues, components, projects, hotspots, rules, metrics, languages, quality gates, measures, webhooks, and project analyses
  write risk: external SonarCloud API mutation of webhooks (create/update/delete), issue comments/assignment/tags/workflow transitions, and project tags
  approval: required for all write actions; each is an external, user-visible mutation on a connected SonarCloud organization or project
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect sonar-cloud

  # Inspect as structured JSON
  pm connectors inspect sonar-cloud --json

AGENT WORKFLOW
  - Run pm connectors inspect sonar-cloud before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
