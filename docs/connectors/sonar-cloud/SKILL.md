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

## Command Surface

- Declared sonar-cloud API commands.
- Usage: pm sonar-cloud <command> [flags]
- Global flags:
  - --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
- Other Commands
  - operations post-api-authentication-logout - Declared direct write: POST /api/authentication/logout. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: ends the caller's own HTTP session; not a syncable record and actively counter-productive for a long-lived connector credential; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-authentication-validate - Declared direct read: GET /api/authentication/validate. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-ce-activity - Declared direct read: GET /api/ce/activity. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-ce-activity-status - Declared direct read: GET /api/ce/activity_status. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-ce-component - Declared direct read: GET /api/ce/component. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-ce-task - Declared direct read: GET /api/ce/task. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-components-search - Declared etl: GET /api/components/search. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations get-api-components-show - Declared direct read: GET /api/components/show. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-components-tree - Declared direct read: GET /api/components/tree. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-duplications-show - Declared direct read: GET /api/duplications/show. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-favorites-add - Declared direct write: POST /api/favorites/add. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: per-user UI favorite-star bookmarking of a project; a personal UI preference, not organizational analysis data or a syncable record; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-favorites-remove - Declared direct write: POST /api/favorites/remove. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: per-user UI favorite-star bookmarking of a project; a personal UI preference, not organizational analysis data or a syncable record; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-favorites-search - Declared direct read: GET /api/favorites/search. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-favourites-index - Declared direct read: GET /api/favourites/index. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-hotspots-change-status - Declared direct write: POST /api/hotspots/change_status. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Security Hotspots are deprecated and being replaced by security issues/vulnerabilities per the api/hotspots service's own catalog description (deprecatedSince 16 June 2026); the read-side api/hotspots/search is still covered for continuity, but this mutation is not implemented against a deprecated model; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-hotspots-search - Declared etl: GET /api/hotspots/search. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations get-api-hotspots-show - Declared direct read: GET /api/hotspots/show. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-issues-add-comment - Declared direct write: POST /api/issues/add_comment. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/issues/add_comment.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-issues-assign - Declared direct write: POST /api/issues/assign. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/issues/assign.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-issues-authors - Declared direct read: GET /api/issues/authors. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-issues-bulk-change - Declared direct write: POST /api/issues/bulk_change. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: multi-issue batch mutation (assign/transition/tag/comment across an issue-key list in one call); the single-issue mutations (assign_issue/set_issue_tags/do_issue_transition/add_issue_comment) already cover the same primitive operations this dialect's per-record write shape supports, applied one issue at a time; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-issues-changelog - Declared direct read: GET /api/issues/changelog. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-issues-delete-comment - Declared direct write: POST /api/issues/delete_comment. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: deletes a specific issue comment by its own comment key; a rarely-needed cleanup action with no corresponding read stream of comment keys to source an id from; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-issues-do-transition - Declared direct write: POST /api/issues/do_transition. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/issues/do_transition.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-issues-edit-comment - Declared direct write: POST /api/issues/edit_comment. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: edits a specific issue comment by its own comment key; a rarely-needed cleanup action with no corresponding read stream of comment keys to source an id from; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-issues-search - Declared etl: GET /api/issues/search. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations post-api-issues-set-severity - Declared direct write: POST /api/issues/set_severity. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: deprecated 25 Aug 2023 per the catalog; manual issue severity override has been superseded by the impacts/cleanCodeAttribute model; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-issues-set-tags - Declared direct write: POST /api/issues/set_tags. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/issues/set_tags.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-issues-set-type - Declared direct write: POST /api/issues/set_type. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: deprecated 25 Aug 2023 per the catalog; manual issue type override has been superseded by the impacts/cleanCodeAttribute model; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-issues-tags - Declared direct read: GET /api/issues/tags. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-languages-list - Declared etl: GET /api/languages/list. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations get-api-measures-component - Declared direct read: GET /api/measures/component. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-measures-component-tree - Declared direct read: GET /api/measures/component_tree. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-measures-search-history - Declared direct read: GET /api/measures/search_history. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-metrics-domains - Declared direct read: GET /api/metrics/domains. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-metrics-search - Declared etl: GET /api/metrics/search. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations get-api-metrics-types - Declared direct read: GET /api/metrics/types. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-notifications-add - Declared direct write: POST /api/notifications/add. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: per-user personal email-notification subscription preference, not organizational analysis data; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-notifications-list - Declared direct read: GET /api/notifications/list. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-notifications-remove - Declared direct write: POST /api/notifications/remove. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: per-user personal email-notification subscription preference, not organizational analysis data; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-permissions-add-group - Declared direct write: POST /api/permissions/add_group. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: grants a permission to a group at global or project scope; account/access-control administration, out of remit for a data-sync connector; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-permissions-add-group-to-template - Declared direct write: POST /api/permissions/add_group_to_template. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: permission-template administration; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-permissions-add-project-creator-to-template - Declared direct write: POST /api/permissions/add_project_creator_to_template. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: permission-template administration; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-permissions-add-user - Declared direct write: POST /api/permissions/add_user. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: grants a permission to a user at global or project scope; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-permissions-add-user-to-template - Declared direct write: POST /api/permissions/add_user_to_template. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: permission-template administration; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-permissions-apply-template - Declared direct write: POST /api/permissions/apply_template. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: applies a permission template to a project, overwriting its existing access grants; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-permissions-bulk-apply-template - Declared direct write: POST /api/permissions/bulk_apply_template. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: bulk permission-template application across many projects at once; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-permissions-create-template - Declared direct write: POST /api/permissions/create_template. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: permission-template administration; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-permissions-delete-template - Declared direct write: POST /api/permissions/delete_template. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: irreversibly deletes a permission template; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-permissions-remove-group - Declared direct write: POST /api/permissions/remove_group. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: revokes a permission from a group; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-permissions-remove-group-from-template - Declared direct write: POST /api/permissions/remove_group_from_template. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: permission-template administration; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-permissions-remove-project-creator-from-template - Declared direct write: POST /api/permissions/remove_project_creator_from_template. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: permission-template administration; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-permissions-remove-user - Declared direct write: POST /api/permissions/remove_user. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: revokes a permission from a user; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-permissions-remove-user-from-template - Declared direct write: POST /api/permissions/remove_user_from_template. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: permission-template administration; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-permissions-search-templates - Declared direct read: GET /api/permissions/search_templates. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-permissions-set-default-template - Declared direct write: POST /api/permissions/set_default_template. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: changes which permission template new projects get by default; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-permissions-update-template - Declared direct write: POST /api/permissions/update_template. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: permission-template administration; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-project-analyses-create-event - Declared direct write: POST /api/project_analyses/create_event. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: annotates a specific analysis with a custom marker event; a niche write with no corresponding common sync use case, and its target analysis key is only obtainable via the already-covered project_analyses read stream, adding a fan-out dependency not worth the complexity for this action alone; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-project-analyses-delete - Declared direct write: POST /api/project_analyses/delete. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: irreversibly deletes a historical analysis snapshot and all its associated measures/issues-at-that-point-in-time; destructive project-data administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-project-analyses-delete-event - Declared direct write: POST /api/project_analyses/delete_event. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: irreversibly deletes an analysis marker event; destructive project-data administration; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-project-analyses-search - Declared etl: GET /api/project_analyses/search. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations post-api-project-analyses-set-baseline - Declared direct write: POST /api/project_analyses/set_baseline. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: changes which historical analysis the New Code Definition is measured against; project-administrator-scoped configuration, not a data-sync mutation; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-project-analyses-unset-baseline - Declared direct write: POST /api/project_analyses/unset_baseline. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: reverts the New Code Definition baseline; project-administrator-scoped configuration, not a data-sync mutation; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-project-analyses-update-event - Declared direct write: POST /api/project_analyses/update_event. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: renames a specific analysis marker event; a niche write with no corresponding common sync use case, and its target event key is only obtainable via the already-covered project_analyses read stream's nested events[], adding a fan-out dependency not worth the complexity for this action alone; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-project-badges-ai-code-assurance - Declared direct read: GET /api/project_badges/ai_code_assurance. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-project-badges-measure - Declared direct read: GET /api/project_badges/measure. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-project-badges-quality-gate - Declared direct read: GET /api/project_badges/quality_gate. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-project-branches-delete - Declared direct write: POST /api/project_branches/delete. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: irreversibly deletes a project branch and its analysis history; destructive project-data administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-project-branches-list - Declared direct read: GET /api/project_branches/list. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-project-branches-rename - Declared direct write: POST /api/project_branches/rename. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: renames a project's main branch; project-administrator-scoped configuration, not a data-sync mutation; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-project-links-create - Declared direct write: POST /api/project_links/create. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: manages a project's auxiliary reference links (homepage/CI/issue-tracker URLs); a project-metadata convenience field, not core analysis data; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-project-links-delete - Declared direct write: POST /api/project_links/delete. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: manages a project's auxiliary reference links; a project-metadata convenience field, not core analysis data; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-project-links-search - Declared direct read: GET /api/project_links/search. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-project-pull-requests-delete - Declared direct write: POST /api/project_pull_requests/delete. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: irreversibly deletes a project pull-request analysis and its history; destructive project-data administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-project-pull-requests-list - Declared direct read: GET /api/project_pull_requests/list. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-project-tags-search - Declared direct read: GET /api/project_tags/search. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-project-tags-set - Declared direct write: POST /api/project_tags/set. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/project_tags/set.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-projects-bulk-delete - Declared direct write: POST /api/projects/bulk_delete. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: irreversibly deletes multiple projects and all their analysis history in one call; destructive project-data administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-projects-bulk-update-key - Declared direct write: POST /api/projects/bulk_update_key. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: deprecated since 7.6 per the catalog; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-projects-create - Declared direct write: POST /api/projects/create. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: requires the 'Create Projects' organization permission per the action's own description; project provisioning is an organization-administration action, not a typical data-sync mutation; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-projects-delete - Declared direct write: POST /api/projects/delete. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: irreversibly deletes a project and all its analysis history; destructive project-data administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-projects-search - Declared etl: GET /api/projects/search. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations post-api-projects-update-key - Declared direct write: POST /api/projects/update_key. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: renames a project's unique key, which is also its primary-key identifier for every other stream in this bundle; requires project-administrator rights and changing it would break cross-stream record correlation; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-projects-update-visibility - Declared direct write: POST /api/projects/update_visibility. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: changes a project's public/private visibility; requires project-administrator rights, an access-control change rather than a data-sync mutation; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-properties-index - Declared direct read: GET /api/properties/index. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-qualitygates-copy - Declared direct write: POST /api/qualitygates/copy. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: deprecated 16 September 2025 per the catalog; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-qualitygates-create - Declared direct write: POST /api/qualitygates/create. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: deprecated 16 September 2025 per the catalog; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-qualitygates-create-condition - Declared direct write: POST /api/qualitygates/create_condition. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: deprecated 16 September 2025 per the catalog; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-qualitygates-delete-condition - Declared direct write: POST /api/qualitygates/delete_condition. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: deprecated 16 September 2025 per the catalog; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-qualitygates-deselect - Declared direct write: POST /api/qualitygates/deselect. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: deprecated 16 September 2025 per the catalog; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-qualitygates-destroy - Declared direct write: POST /api/qualitygates/destroy. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: deprecated 16 September 2025 per the catalog; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-qualitygates-get-by-project - Declared direct read: GET /api/qualitygates/get_by_project. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-qualitygates-list - Declared etl: GET /api/qualitygates/list. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations get-api-qualitygates-project-status - Declared direct read: GET /api/qualitygates/project_status. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-qualitygates-rename - Declared direct write: POST /api/qualitygates/rename. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: deprecated 16 September 2025 per the catalog; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-qualitygates-search - Declared direct read: GET /api/qualitygates/search. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-qualitygates-select - Declared direct write: POST /api/qualitygates/select. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: deprecated 16 September 2025 per the catalog; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-qualitygates-set-as-default - Declared direct write: POST /api/qualitygates/set_as_default. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: deprecated 16 September 2025 per the catalog; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-qualitygates-show - Declared direct read: GET /api/qualitygates/show. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-qualitygates-unset-default - Declared direct write: POST /api/qualitygates/unset_default. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: deprecated since 7.0 per the catalog; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-qualitygates-update-condition - Declared direct write: POST /api/qualitygates/update_condition. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: deprecated 16 September 2025 per the catalog; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-qualityprofiles-activate-rule - Declared direct write: POST /api/qualityprofiles/activate_rule. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: requires 'Administer Quality Profiles' organization permission per the underlying action; quality-profile administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-qualityprofiles-activate-rules - Declared direct write: POST /api/qualityprofiles/activate_rules. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: requires 'Administer Quality Profiles' organization permission; quality-profile administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-qualityprofiles-add-project - Declared direct write: POST /api/qualityprofiles/add_project. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: requires project-administrator rights; quality-profile administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-qualityprofiles-backup - Declared direct read: GET /api/qualityprofiles/backup. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-qualityprofiles-change-parent - Declared direct write: POST /api/qualityprofiles/change_parent. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: requires 'Administer Quality Profiles' organization permission; quality-profile administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-qualityprofiles-changelog - Declared direct read: GET /api/qualityprofiles/changelog. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-qualityprofiles-copy - Declared direct write: POST /api/qualityprofiles/copy. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: requires 'Administer Quality Profiles' organization permission; quality-profile administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-qualityprofiles-create - Declared direct write: POST /api/qualityprofiles/create. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: requires 'Administer Quality Profiles' organization permission; quality-profile administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-qualityprofiles-deactivate-rule - Declared direct write: POST /api/qualityprofiles/deactivate_rule. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: requires 'Administer Quality Profiles' organization permission; quality-profile administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-qualityprofiles-deactivate-rules - Declared direct write: POST /api/qualityprofiles/deactivate_rules. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: requires 'Administer Quality Profiles' organization permission; quality-profile administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-qualityprofiles-delete - Declared direct write: POST /api/qualityprofiles/delete. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: irreversibly deletes a quality profile; destructive quality-profile administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-qualityprofiles-export - Declared direct read: GET /api/qualityprofiles/export. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-qualityprofiles-exporters - Declared direct read: GET /api/qualityprofiles/exporters. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-qualityprofiles-importers - Declared direct read: GET /api/qualityprofiles/importers. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-qualityprofiles-inheritance - Declared direct read: GET /api/qualityprofiles/inheritance. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-qualityprofiles-projects - Declared direct read: GET /api/qualityprofiles/projects. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-qualityprofiles-remove-project - Declared direct write: POST /api/qualityprofiles/remove_project. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: requires project-administrator rights; quality-profile administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-qualityprofiles-rename - Declared direct write: POST /api/qualityprofiles/rename. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: requires 'Administer Quality Profiles' organization permission; quality-profile administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-qualityprofiles-restore - Declared direct write: POST /api/qualityprofiles/restore. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: requires 'Administer Quality Profiles' organization permission; quality-profile administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-qualityprofiles-restore-built-in - Declared direct write: POST /api/qualityprofiles/restore_built_in. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: deprecated since 6.4 per the catalog; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-qualityprofiles-search - Declared direct read: GET /api/qualityprofiles/search. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-qualityprofiles-set-default - Declared direct write: POST /api/qualityprofiles/set_default. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: requires 'Administer Quality Profiles' organization permission; quality-profile administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-rules-repositories - Declared direct read: GET /api/rules/repositories. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-rules-search - Declared etl: GET /api/rules/search. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations get-api-rules-show - Declared direct read: GET /api/rules/show. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-rules-tags - Declared direct read: GET /api/rules/tags. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-rules-update - Declared direct write: POST /api/rules/update. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: requires 'Administer Quality Profiles' organization permission per the action's own description; custom-rule authoring/rule-metadata administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-settings-list-definitions - Declared direct read: GET /api/settings/list_definitions. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-settings-reset - Declared direct write: POST /api/settings/reset. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: resets an instance/project setting to its default value; configuration administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-settings-set - Declared direct write: POST /api/settings/set. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: changes an instance/project configuration setting; requires administrator rights, configuration administration rather than a data-sync mutation; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-settings-values - Declared direct read: GET /api/settings/values. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-sources-raw - Declared direct read: GET /api/sources/raw. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-sources-scm - Declared direct read: GET /api/sources/scm. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-sources-show - Declared direct read: GET /api/sources/show. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-timemachine-index - Declared direct read: GET /api/timemachine/index. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-user-groups-add-user - Declared direct write: POST /api/user_groups/add_user. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: adds a user to a group; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-user-groups-create - Declared direct write: POST /api/user_groups/create. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: creates a user group; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-user-groups-delete - Declared direct write: POST /api/user_groups/delete. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: irreversibly deletes a user group; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-user-groups-remove-user - Declared direct write: POST /api/user_groups/remove_user. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: removes a user from a group; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-user-groups-search - Declared direct read: GET /api/user_groups/search. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-user-groups-update - Declared direct write: POST /api/user_groups/update. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: renames/redescribes a user group; account/access-control administration, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-user-groups-users - Declared direct read: GET /api/user_groups/users. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-user-properties-index - Declared direct read: GET /api/user_properties/index. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-user-tokens-generate - Declared direct write: POST /api/user_tokens/generate. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: mints a new personal access token for a user; a credential-issuance action with security implications well outside a data-sync connector's remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-user-tokens-revoke - Declared direct write: POST /api/user_tokens/revoke. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: irreversibly revokes a personal access token; a credential-lifecycle action with security implications, out of remit; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-user-tokens-search - Declared direct read: GET /api/user_tokens/search. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-users-groups - Declared direct read: GET /api/users/groups. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations post-api-webhooks-create - Declared direct write: POST /api/webhooks/create. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/webhooks/create.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations post-api-webhooks-delete - Declared direct write: POST /api/webhooks/delete. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/webhooks/delete.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-webhooks-deliveries - Declared direct read: GET /api/webhooks/deliveries. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-webhooks-delivery - Declared direct read: GET /api/webhooks/delivery. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-webhooks-list - Declared etl: GET /api/webhooks/list. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.
  - operations post-api-webhooks-update - Declared direct write: POST /api/webhooks/update. [intent=direct_write availability=partial]; approval: direct_write commands require plan, preview, approval, execute; risk: Declared provider mutation: POST /api/webhooks/update.; notes: Unavailable: no bounded direct-write executor is declared for this provider operation.
  - operations get-api-webservices-list - Declared direct read: GET /api/webservices/list. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-webservices-response-example - Declared direct read: GET /api/webservices/response_example. [intent=direct_read availability=partial]; notes: Unavailable: no bounded read executor is declared for this provider operation.; flags: --page, --page-cursor
  - operations get-api-measures-search - Declared etl: GET /api/measures/search. [intent=etl availability=partial]; notes: Unavailable: no ETL executor is declared for this provider operation.

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
