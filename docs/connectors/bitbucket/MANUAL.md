# pm connectors inspect bitbucket

```text
NAME
  pm connectors inspect bitbucket - Bitbucket connector manual

SYNOPSIS
  pm connectors inspect bitbucket
  pm connectors inspect bitbucket --json
  pm credentials add <name> --connector bitbucket [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Bitbucket Cloud workspace, repository, pull request, issue, commit, pipeline, deployment, snippet, and project resources; exposes closed-schema repository creation and path-only deletes while blocking untyped JSON-body and multipart writes.

ICON
  id: pm-sample
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  addon_key
  annotationId
  app_key
  base_url
  cache_uuid
  change_id
  comment_id
  commit
  component_id
  deployment_uuid
  email
  encoded_id
  environment_uuid
  filename
  fingerprint
  group_slug
  id
  issue_id
  key
  key_id
  known_host_uuid
  log_uuid
  member
  milestone_id
  name
  node_id
  path
  pipeline_uuid
  project_key
  property_name
  pull_request_id
  pullrequest_id
  repo_name
  repo_slug
  reportId
  revision
  revspec
  runner_uuid
  schedule_uuid
  selected_user
  selected_user_id
  spec
  start_date
  step_uuid
  subject_type
  target_username
  task_id
  test_case_uuid
  uid
  username
  variable_uuid
  version_id
  workspace
  access_token (secret)
  app_password (secret)

ETL STREAMS
  repositories:
    primary key: uuid
    fields: created_on(string), description(string), fork_policy(string), full_name(string), has_issues(string), has_wiki(string), is_private(string), language(string), links(object), mainbranch(string), name(string), owner(string), parent(string), project(string), scm(string), size(string), slug(string), type(string), updated_on(string), uuid(string)
  hook_events:
    fields: created_on(string), links(object), name(string), repository(string), slug(string), type(string), updated_on(string), uuid(string), workspace(string)
  hook_events_subject_type:
    primary key: event
    fields: category(string), created_on(string), description(string), event(string), label(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace:
    primary key: uuid
    fields: created_on(string), description(string), fork_policy(string), full_name(string), has_issues(string), has_wiki(string), is_private(string), language(string), links(object), mainbranch(string), name(string), owner(string), parent(string), project(string), scm(string), size(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug:
    primary key: uuid
    fields: created_on(string), description(string), fork_policy(string), full_name(string), has_issues(string), has_wiki(string), is_private(string), language(string), links(object), mainbranch(string), name(string), owner(string), parent(string), project(string), scm(string), size(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_branch_restrictions:
    primary key: id
    fields: branch_match_kind(string), branch_type(string), created_on(string), groups(string), id(string), kind(string), links(object), name(string), pattern(string), slug(string), type(string), updated_on(string), users(string), uuid(string), value(string)
  repositories_workspace_repo_slug_branch_restrictions_id:
    primary key: id
    fields: branch_match_kind(string), branch_type(string), created_on(string), groups(string), id(string), kind(string), links(object), name(string), pattern(string), slug(string), type(string), updated_on(string), users(string), uuid(string), value(string)
  repositories_workspace_repo_slug_branching_model:
    fields: branch_types(string), created_on(string), development(string), links(object), name(string), production(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_branching_model_settings:
    fields: branch_types(string), created_on(string), development(string), links(object), name(string), production(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_commit_commit:
    primary key: hash
    fields: author(string), committer(string), created_on(string), date(string), hash(string), links(object), message(string), name(string), parents(string), participants(string), repository(string), slug(string), summary(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_commit_commit_comments:
    primary key: id
    fields: commit(string), content(string), created_on(string), deleted(string), id(string), inline(string), links(object), name(string), parent(string), slug(string), type(string), updated_on(string), user(string), uuid(string)
  repositories_workspace_repo_slug_commit_commit_comments_comment_id:
    primary key: id
    fields: commit(string), content(string), created_on(string), deleted(string), id(string), inline(string), links(object), name(string), parent(string), slug(string), type(string), updated_on(string), user(string), uuid(string)
  repositories_workspace_repo_slug_commit_commit_properties_app_key_property_name:
    fields: _attributes(string), created_on(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_commit_commit_pullrequests:
    primary key: id
    fields: author(string), close_source_branch(string), closed_by(string), comment_count(string), created_on(string), destination(string), draft(string), id(string), links(object), merge_commit(string), name(string), participants(string), queued(string), reason(string), rendered(string), reviewers(string), slug(string), source(string), state(string), summary(string), task_count(string), title(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_commit_commit_reports:
    primary key: external_id
    fields: created_on(string), data(string), details(string), external_id(string), link(string), links(object), logo_url(string), name(string), remote_link_enabled(string), report_type(string), reporter(string), result(string), slug(string), title(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_commit_commit_reports_reportid:
    primary key: external_id
    fields: created_on(string), data(string), details(string), external_id(string), link(string), links(object), logo_url(string), name(string), remote_link_enabled(string), report_type(string), reporter(string), result(string), slug(string), title(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_commit_commit_reports_reportid_annotations:
    primary key: external_id
    fields: annotation_type(string), created_on(string), details(string), external_id(string), line(string), link(string), links(object), name(string), path(string), result(string), severity(string), slug(string), summary(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_commit_commit_reports_reportid_annotations_annotationid:
    primary key: external_id
    fields: annotation_type(string), created_on(string), details(string), external_id(string), line(string), link(string), links(object), name(string), path(string), result(string), severity(string), slug(string), summary(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_commit_commit_statuses:
    primary key: key
    fields: created_on(string), description(string), key(string), links(object), name(string), refname(string), slug(string), state(string), type(string), updated_on(string), url(string), uuid(string)
  repositories_workspace_repo_slug_commit_commit_statuses_build_key:
    primary key: key
    fields: created_on(string), description(string), key(string), links(object), name(string), refname(string), slug(string), state(string), type(string), updated_on(string), url(string), uuid(string)
  repositories_workspace_repo_slug_commits:
    primary key: hash
    fields: author(string), committer(string), created_on(string), date(string), hash(string), links(object), message(string), name(string), parents(string), slug(string), summary(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_commits_revision:
    primary key: hash
    fields: author(string), committer(string), created_on(string), date(string), hash(string), links(object), message(string), name(string), parents(string), slug(string), summary(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_components:
    primary key: id
    fields: created_on(string), id(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_components_component_id:
    primary key: id
    fields: created_on(string), id(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_default_reviewers:
    primary key: uuid
    fields: created_on(string), display_name(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_default_reviewers_target_username:
    primary key: uuid
    fields: created_on(string), display_name(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_deploy_keys:
    primary key: id
    fields: added_on(string), comment(string), created_on(string), id(string), key(string), label(string), last_used(string), links(object), name(string), owner(string), repository(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_deploy_keys_key_id:
    primary key: id
    fields: added_on(string), comment(string), created_on(string), id(string), key(string), label(string), last_used(string), links(object), name(string), owner(string), repository(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_deployments:
    primary key: uuid
    fields: created_on(string), environment(string), links(object), name(string), release(string), slug(string), state(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_deployments_config_environments_environment_uuid_variables:
    primary key: uuid
    fields: created_on(string), key(string), links(object), name(string), secured(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_deployments_deployment_uuid:
    primary key: uuid
    fields: created_on(string), environment(string), links(object), name(string), release(string), slug(string), state(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_diffstat_spec:
    fields: created_on(string), lines_added(string), lines_removed(string), links(object), name(string), new(string), old(string), slug(string), status(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_effective_branching_model:
    fields: branch_types(string), created_on(string), development(string), links(object), name(string), production(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_effective_default_reviewers:
    primary key: uuid
    fields: created_on(string), links(object), name(string), reviewer_type(string), slug(string), type(string), updated_on(string), user(string), uuid(string)
  repositories_workspace_repo_slug_environments:
    primary key: uuid
    fields: created_on(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_environments_environment_uuid:
    primary key: uuid
    fields: created_on(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_file_conflicts_spec:
    primary key: path
    fields: created_on(string), links(object), message(string), name(string), path(string), scenario(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_filehistory_commit_path:
    primary key: path
    fields: attributes(string), commit(string), created_on(string), escaped_path(string), links(object), name(string), path(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_forks:
    primary key: uuid
    fields: created_on(string), description(string), fork_policy(string), full_name(string), has_issues(string), has_wiki(string), is_private(string), language(string), links(object), mainbranch(string), name(string), owner(string), parent(string), project(string), scm(string), size(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_hooks:
    primary key: uuid
    fields: active(string), created_at(string), created_on(string), description(string), events(string), links(object), name(string), secret_set(string), slug(string), subject(string), subject_type(string), type(string), updated_on(string), url(string), uuid(string)
  repositories_workspace_repo_slug_hooks_uid:
    primary key: uuid
    fields: active(string), created_at(string), created_on(string), description(string), events(string), links(object), name(string), secret_set(string), slug(string), subject(string), subject_type(string), type(string), updated_on(string), url(string), uuid(string)
  repositories_workspace_repo_slug_issues:
    primary key: id
    fields: assignee(string), component(string), content(string), created_on(string), edited_on(string), id(string), kind(string), links(object), milestone(string), name(string), priority(string), reporter(string), repository(string), slug(string), state(string), title(string), type(string), updated_on(string), uuid(string), version(string), votes(string)
  repositories_workspace_repo_slug_issues_import:
    fields: count(string), created_on(string), links(object), name(string), pct(string), phase(string), slug(string), status(string), total(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_issues_issue_id:
    primary key: id
    fields: assignee(string), component(string), content(string), created_on(string), edited_on(string), id(string), kind(string), links(object), milestone(string), name(string), priority(string), reporter(string), repository(string), slug(string), state(string), title(string), type(string), updated_on(string), uuid(string), version(string), votes(string)
  repositories_workspace_repo_slug_issues_issue_id_attachments:
    primary key: name
    fields: created_on(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_issues_issue_id_changes:
    primary key: id
    fields: changes(string), created_on(string), id(string), issue(string), links(object), message(string), name(string), slug(string), type(string), updated_on(string), user(string), uuid(string)
  repositories_workspace_repo_slug_issues_issue_id_changes_change_id:
    primary key: id
    fields: changes(string), created_on(string), id(string), issue(string), links(object), message(string), name(string), slug(string), type(string), updated_on(string), user(string), uuid(string)
  repositories_workspace_repo_slug_issues_issue_id_comments:
    primary key: id
    fields: content(string), created_on(string), deleted(string), id(string), inline(string), issue(string), links(object), name(string), parent(string), slug(string), type(string), updated_on(string), user(string), uuid(string)
  repositories_workspace_repo_slug_issues_issue_id_comments_comment_id:
    primary key: id
    fields: content(string), created_on(string), deleted(string), id(string), inline(string), issue(string), links(object), name(string), parent(string), slug(string), type(string), updated_on(string), user(string), uuid(string)
  repositories_workspace_repo_slug_merge_base_revspec:
    primary key: hash
    fields: author(string), committer(string), created_on(string), date(string), hash(string), links(object), message(string), name(string), parents(string), participants(string), repository(string), slug(string), summary(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_milestones:
    primary key: id
    fields: created_on(string), id(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_milestones_milestone_id:
    primary key: id
    fields: created_on(string), id(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_override_settings:
    fields: created_on(string), links(object), name(string), override_settings(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_permissions_config_groups:
    fields: created_on(string), group(string), links(object), name(string), permission(string), repository(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_permissions_config_groups_group_slug:
    fields: created_on(string), group(string), links(object), name(string), permission(string), repository(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_permissions_config_users:
    fields: created_on(string), links(object), name(string), permission(string), repository(string), slug(string), type(string), updated_on(string), user(string), uuid(string)
  repositories_workspace_repo_slug_permissions_config_users_selected_user_id:
    fields: created_on(string), links(object), name(string), permission(string), repository(string), slug(string), type(string), updated_on(string), user(string), uuid(string)
  repositories_workspace_repo_slug_pipelines:
    primary key: uuid
    fields: build_number(string), build_seconds_used(string), completed_on(string), configuration_sources(string), created_on(string), creator(string), links(object), name(string), repository(string), slug(string), state(string), target(string), trigger(string), type(string), updated_on(string), uuid(string), variables(string)
  repositories_workspace_repo_slug_pipelines_config:
    fields: created_on(string), enabled(string), links(object), name(string), repository(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_pipelines_config_caches:
    primary key: key_hash
    fields: created_on(string), file_size_bytes(string), key_hash(string), links(object), name(string), path(string), pipeline_uuid(string), slug(string), step_uuid(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_pipelines_config_caches_cache_uuid_content_uri:
    primary key: uri
    fields: created_on(string), links(object), name(string), slug(string), type(string), updated_on(string), uri(string), uuid(string)
  repositories_workspace_repo_slug_pipelines_config_runners:
    primary key: uuid
    fields: created_on(string), labels(string), links(object), name(string), oauth_client(string), slug(string), state(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_pipelines_config_runners_runner_uuid:
    primary key: uuid
    fields: created_on(string), labels(string), links(object), name(string), oauth_client(string), slug(string), state(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_pipelines_config_schedules:
    primary key: uuid
    fields: created_on(string), cron_pattern(string), enabled(string), links(object), name(string), slug(string), target(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_pipelines_config_schedules_schedule_uuid:
    primary key: uuid
    fields: created_on(string), cron_pattern(string), enabled(string), links(object), name(string), slug(string), target(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_pipelines_config_schedules_schedule_uuid_executions:
    primary key: uuid
    fields: created_on(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_pipelines_config_ssh_key_pair:
    fields: created_on(string), links(object), name(string), public_key(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_pipelines_config_ssh_known_hosts:
    primary key: uuid
    fields: created_on(string), hostname(string), links(object), name(string), public_key(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_pipelines_config_ssh_known_hosts_known_host_uuid:
    primary key: uuid
    fields: created_on(string), hostname(string), links(object), name(string), public_key(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_pipelines_config_variables:
    primary key: uuid
    fields: created_on(string), key(string), links(object), name(string), secured(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_pipelines_config_variables_variable_uuid:
    primary key: uuid
    fields: created_on(string), key(string), links(object), name(string), secured(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_pipelines_pipeline_uuid:
    primary key: uuid
    fields: build_number(string), build_seconds_used(string), completed_on(string), configuration_sources(string), created_on(string), creator(string), links(object), name(string), repository(string), slug(string), state(string), target(string), trigger(string), type(string), updated_on(string), uuid(string), variables(string)
  repositories_workspace_repo_slug_pipelines_pipeline_uuid_steps:
    primary key: uuid
    fields: completed_on(string), created_on(string), image(string), links(object), name(string), script_commands(string), setup_commands(string), slug(string), started_on(string), state(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_pipelines_pipeline_uuid_steps_step_uuid:
    primary key: uuid
    fields: completed_on(string), created_on(string), image(string), links(object), name(string), script_commands(string), setup_commands(string), slug(string), started_on(string), state(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_properties_app_key_property_name:
    fields: _attributes(string), created_on(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_pullrequests:
    primary key: id
    fields: author(string), close_source_branch(string), closed_by(string), comment_count(string), created_on(string), destination(string), draft(string), id(string), links(object), merge_commit(string), name(string), participants(string), queued(string), reason(string), rendered(string), reviewers(string), slug(string), source(string), state(string), summary(string), task_count(string), title(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_pullrequests_pull_request_id:
    primary key: id
    fields: author(string), close_source_branch(string), closed_by(string), comment_count(string), created_on(string), destination(string), draft(string), id(string), links(object), merge_commit(string), name(string), participants(string), queued(string), reason(string), rendered(string), reviewers(string), slug(string), source(string), state(string), summary(string), task_count(string), title(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_pullrequests_pull_request_id_comments:
    primary key: id
    fields: content(string), created_on(string), deleted(string), id(string), inline(string), links(object), name(string), parent(string), pending(string), pullrequest(string), resolution(string), slug(string), type(string), updated_on(string), user(string), uuid(string)
  repositories_workspace_repo_slug_pullrequests_pull_request_id_comments_comment_id:
    primary key: id
    fields: content(string), created_on(string), deleted(string), id(string), inline(string), links(object), name(string), parent(string), pending(string), pullrequest(string), resolution(string), slug(string), type(string), updated_on(string), user(string), uuid(string)
  repositories_workspace_repo_slug_pullrequests_pull_request_id_statuses:
    primary key: key
    fields: created_on(string), description(string), key(string), links(object), name(string), refname(string), slug(string), state(string), type(string), updated_on(string), url(string), uuid(string)
  repositories_workspace_repo_slug_pullrequests_pull_request_id_tasks:
    primary key: id
    fields: comment(string), content(string), created_on(string), creator(string), id(string), links(object), name(string), pending(string), resolved_by(string), resolved_on(string), slug(string), state(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_pullrequests_pull_request_id_tasks_task_id:
    primary key: id
    fields: comment(string), content(string), created_on(string), creator(string), id(string), links(object), name(string), pending(string), resolved_by(string), resolved_on(string), slug(string), state(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_pullrequests_pullrequest_id_properties_app_key_property_name:
    fields: _attributes(string), created_on(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_refs:
    primary key: name
    fields: created_on(string), links(object), name(string), slug(string), target(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_refs_branches:
    primary key: name
    fields: created_on(string), default_merge_strategy(string), links(object), merge_strategies(string), name(string), slug(string), target(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_refs_branches_name:
    primary key: name
    fields: created_on(string), default_merge_strategy(string), links(object), merge_strategies(string), name(string), slug(string), target(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_refs_tags:
    primary key: name
    fields: created_on(string), date(string), links(object), message(string), name(string), slug(string), tagger(string), target(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_refs_tags_name:
    primary key: name
    fields: created_on(string), date(string), links(object), message(string), name(string), slug(string), tagger(string), target(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_src:
    fields: commit(string), created_on(string), links(object), name(string), path(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_versions:
    primary key: id
    fields: created_on(string), id(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_versions_version_id:
    primary key: id
    fields: created_on(string), id(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  repositories_workspace_repo_slug_watchers:
    primary key: uuid
    fields: created_on(string), display_name(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  snippets:
    primary key: id
    fields: created_on(string), creator(string), id(string), is_private(string), links(object), name(string), owner(string), scm(string), slug(string), title(string), type(string), updated_on(string), uuid(string)
  snippets_workspace:
    primary key: id
    fields: created_on(string), creator(string), id(string), is_private(string), links(object), name(string), owner(string), scm(string), slug(string), title(string), type(string), updated_on(string), uuid(string)
  snippets_workspace_encoded_id:
    primary key: id
    fields: created_on(string), creator(string), id(string), is_private(string), links(object), name(string), owner(string), scm(string), slug(string), title(string), type(string), updated_on(string), uuid(string)
  snippets_workspace_encoded_id_comments:
    primary key: id
    fields: created_on(string), id(string), links(object), name(string), slug(string), snippet(string), type(string), updated_on(string), uuid(string)
  snippets_workspace_encoded_id_comments_comment_id:
    primary key: id
    fields: created_on(string), id(string), links(object), name(string), slug(string), snippet(string), type(string), updated_on(string), uuid(string)
  snippets_workspace_encoded_id_commits:
    primary key: hash
    fields: author(string), committer(string), created_on(string), date(string), hash(string), links(object), message(string), name(string), parents(string), slug(string), snippet(string), summary(string), type(string), updated_on(string), uuid(string)
  snippets_workspace_encoded_id_commits_revision:
    primary key: hash
    fields: author(string), committer(string), created_on(string), date(string), hash(string), links(object), message(string), name(string), parents(string), slug(string), snippet(string), summary(string), type(string), updated_on(string), uuid(string)
  snippets_workspace_encoded_id_node_id:
    primary key: id
    fields: created_on(string), creator(string), id(string), is_private(string), links(object), name(string), owner(string), scm(string), slug(string), title(string), type(string), updated_on(string), uuid(string)
  snippets_workspace_encoded_id_watchers:
    primary key: uuid
    fields: created_on(string), display_name(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  teams_username_pipelines_config_variables:
    primary key: uuid
    fields: created_on(string), key(string), links(object), name(string), secured(string), slug(string), type(string), updated_on(string), uuid(string)
  teams_username_pipelines_config_variables_variable_uuid:
    primary key: uuid
    fields: created_on(string), key(string), links(object), name(string), secured(string), slug(string), type(string), updated_on(string), uuid(string)
  user:
    primary key: uuid
    fields: created_on(string), display_name(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  user_permissions_repositories:
    fields: created_on(string), links(object), name(string), permission(string), repository(string), slug(string), type(string), updated_on(string), user(string), uuid(string)
  user_permissions_workspaces:
    fields: created_on(string), links(object), name(string), slug(string), type(string), updated_on(string), user(string), uuid(string), workspace(string)
  user_workspaces:
    fields: administrator(string), created_on(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string), workspace(string)
  user_workspaces_workspace_permission:
    fields: created_on(string), links(object), name(string), slug(string), type(string), updated_on(string), user(string), uuid(string), workspace(string)
  user_workspaces_workspace_permissions_repositories:
    fields: created_on(string), links(object), name(string), permission(string), repository(string), slug(string), type(string), updated_on(string), user(string), uuid(string)
  users_selected_user:
    primary key: uuid
    fields: created_on(string), display_name(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  users_selected_user_gpg_keys:
    primary key: fingerprint
    fields: added_on(string), created_on(string), expires_on(string), fingerprint(string), key(string), key_id(string), last_used(string), links(object), name(string), owner(string), parent_fingerprint(string), slug(string), subkeys(string), type(string), updated_on(string), uuid(string)
  users_selected_user_gpg_keys_fingerprint:
    primary key: fingerprint
    fields: added_on(string), created_on(string), expires_on(string), fingerprint(string), key(string), key_id(string), last_used(string), links(object), name(string), owner(string), parent_fingerprint(string), slug(string), subkeys(string), type(string), updated_on(string), uuid(string)
  users_selected_user_pipelines_config_variables:
    primary key: uuid
    fields: created_on(string), key(string), links(object), name(string), secured(string), slug(string), type(string), updated_on(string), uuid(string)
  users_selected_user_pipelines_config_variables_variable_uuid:
    primary key: uuid
    fields: created_on(string), key(string), links(object), name(string), secured(string), slug(string), type(string), updated_on(string), uuid(string)
  users_selected_user_properties_app_key_property_name:
    fields: _attributes(string), created_on(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  users_selected_user_ssh_keys:
    primary key: fingerprint
    fields: comment(string), created_on(string), expires_on(string), fingerprint(string), key(string), label(string), last_used(string), links(object), name(string), owner(string), slug(string), type(string), updated_on(string), uuid(string)
  users_selected_user_ssh_keys_key_id:
    primary key: fingerprint
    fields: comment(string), created_on(string), expires_on(string), fingerprint(string), key(string), label(string), last_used(string), links(object), name(string), owner(string), slug(string), type(string), updated_on(string), uuid(string)
  workspaces:
    primary key: uuid
    fields: created_on(string), forking_mode(string), is_privacy_enforced(string), is_private(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  workspaces_workspace:
    primary key: uuid
    fields: created_on(string), forking_mode(string), is_privacy_enforced(string), is_private(string), links(object), name(string), slug(string), type(string), updated_on(string), uuid(string)
  workspaces_workspace_hooks:
    primary key: uuid
    fields: active(string), created_at(string), created_on(string), description(string), events(string), links(object), name(string), secret_set(string), slug(string), subject(string), subject_type(string), type(string), updated_on(string), url(string), uuid(string)
  workspaces_workspace_hooks_uid:
    primary key: uuid
    fields: active(string), created_at(string), created_on(string), description(string), events(string), links(object), name(string), secret_set(string), slug(string), subject(string), subject_type(string), type(string), updated_on(string), url(string), uuid(string)
  workspaces_workspace_members:
    fields: created_on(string), links(object), name(string), slug(string), type(string), updated_on(string), user(string), uuid(string), workspace(string)
  workspaces_workspace_members_member:
    fields: created_on(string), links(object), name(string), slug(string), type(string), updated_on(string), user(string), uuid(string), workspace(string)
  workspaces_workspace_permissions:
    fields: created_on(string), links(object), name(string), slug(string), type(string), updated_on(string), user(string), uuid(string), workspace(string)
  workspaces_workspace_permissions_repositories:
    fields: created_on(string), links(object), name(string), permission(string), repository(string), slug(string), type(string), updated_on(string), user(string), uuid(string)
  workspaces_workspace_permissions_repositories_repo_slug:
    fields: created_on(string), links(object), name(string), permission(string), repository(string), slug(string), type(string), updated_on(string), user(string), uuid(string)
  workspaces_workspace_pipelines_config_runners:
    primary key: uuid
    fields: created_on(string), labels(string), links(object), name(string), oauth_client(string), slug(string), state(string), type(string), updated_on(string), uuid(string)
  workspaces_workspace_pipelines_config_runners_runner_uuid:
    primary key: uuid
    fields: created_on(string), labels(string), links(object), name(string), oauth_client(string), slug(string), state(string), type(string), updated_on(string), uuid(string)
  workspaces_workspace_pipelines_config_variables:
    primary key: uuid
    fields: created_on(string), key(string), links(object), name(string), secured(string), slug(string), type(string), updated_on(string), uuid(string)
  workspaces_workspace_pipelines_config_variables_variable_uuid:
    primary key: uuid
    fields: created_on(string), key(string), links(object), name(string), secured(string), slug(string), type(string), updated_on(string), uuid(string)
  workspaces_workspace_projects:
    primary key: uuid
    fields: created_on(string), description(string), has_publicly_visible_repos(string), is_private(string), key(string), links(object), name(string), owner(string), slug(string), type(string), updated_on(string), uuid(string)
  workspaces_workspace_projects_project_key:
    primary key: uuid
    fields: created_on(string), description(string), has_publicly_visible_repos(string), is_private(string), key(string), links(object), name(string), owner(string), slug(string), type(string), updated_on(string), uuid(string)
  workspaces_workspace_projects_project_key_branching_model:
    fields: branch_types(string), created_on(string), development(string), links(object), name(string), production(string), slug(string), type(string), updated_on(string), uuid(string)
  workspaces_workspace_projects_project_key_branching_model_settings:
    fields: branch_types(string), created_on(string), development(string), links(object), name(string), production(string), slug(string), type(string), updated_on(string), uuid(string)
  workspaces_workspace_projects_project_key_default_reviewers:
    primary key: uuid
    fields: created_on(string), links(object), name(string), reviewer_type(string), slug(string), type(string), updated_on(string), user(string), uuid(string)
  workspaces_workspace_projects_project_key_default_reviewers_selected_user:
    primary key: account_id
    fields: account_id(string), account_status(string), created_on(string), display_name(string), has_2fa_enabled(string), is_staff(string), links(object), name(string), nickname(string), slug(string), type(string), updated_on(string), uuid(string)
  workspaces_workspace_projects_project_key_deploy_keys:
    primary key: id
    fields: added_on(string), comment(string), created_by(string), created_on(string), id(string), key(string), label(string), last_used(string), links(object), name(string), project(string), slug(string), type(string), updated_on(string), uuid(string)
  workspaces_workspace_projects_project_key_deploy_keys_key_id:
    primary key: id
    fields: added_on(string), comment(string), created_by(string), created_on(string), id(string), key(string), label(string), last_used(string), links(object), name(string), project(string), slug(string), type(string), updated_on(string), uuid(string)
  workspaces_workspace_projects_project_key_permissions_config_groups:
    fields: created_on(string), group(string), links(object), name(string), permission(string), project(string), slug(string), type(string), updated_on(string), uuid(string)
  workspaces_workspace_projects_project_key_permissions_config_groups_group_slug:
    fields: created_on(string), group(string), links(object), name(string), permission(string), project(string), slug(string), type(string), updated_on(string), uuid(string)
  workspaces_workspace_projects_project_key_permissions_config_users:
    fields: created_on(string), links(object), name(string), permission(string), project(string), slug(string), type(string), updated_on(string), user(string), uuid(string)
  workspaces_workspace_projects_project_key_permissions_config_users_selected_user_id:
    fields: created_on(string), links(object), name(string), permission(string), project(string), slug(string), type(string), updated_on(string), user(string), uuid(string)
  workspaces_workspace_pullrequests_selected_user:
    primary key: id
    fields: author(string), close_source_branch(string), closed_by(string), comment_count(string), created_on(string), destination(string), draft(string), id(string), links(object), merge_commit(string), name(string), participants(string), queued(string), reason(string), rendered(string), reviewers(string), slug(string), source(string), state(string), summary(string), task_count(string), title(string), type(string), updated_on(string), uuid(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite

REVERSE ETL ACTIONS
  delete_repositories_workspace_repo_slug:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}
    required fields: workspace, repo_slug
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  create_repositories_workspace_repo_slug:
    endpoint: POST /repositories/{{ record.workspace }}/{{ record.repo_slug }}
    required fields: workspace, repo_slug, scm
    risk: POST /repositories/{workspace}/{repo_slug} Bitbucket Cloud mutation; execute only through reverse ETL plan, preview, explicit approval, and connector redaction.
  delete_repositories_workspace_repo_slug_branch_restrictions_id:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/branch-restrictions/{{ record.id }}
    required fields: workspace, repo_slug, id
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/branch-restrictions/{id}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_commit_commit_approve:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/commit/{{ record.commit }}/approve
    required fields: workspace, repo_slug, commit
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/commit/{commit}/approve; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_commit_commit_comments_comment_id:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/commit/{{ record.commit }}/comments/{{ record.comment_id }}
    required fields: workspace, repo_slug, commit, comment_id
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/commit/{commit}/comments/{comment_id}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_commit_commit_properties_app_key_property_name:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/commit/{{ record.commit }}/properties/{{ record.app_key }}/{{ record.property_name }}
    required fields: workspace, repo_slug, commit, app_key, property_name
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/commit/{commit}/properties/{app_key}/{property_name}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_commit_commit_reports_reportid:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/commit/{{ record.commit }}/reports/{{ record.reportId }}
    required fields: workspace, repo_slug, commit, reportId
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/commit/{commit}/reports/{reportId}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_commit_commit_reports_reportid_annotations_annotationid:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/commit/{{ record.commit }}/reports/{{ record.reportId }}/annotations/{{ record.annotationId }}
    required fields: workspace, repo_slug, commit, reportId, annotationId
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/commit/{commit}/reports/{reportId}/annotations/{annotationId}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_default_reviewers_target_username:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/default-reviewers/{{ record.target_username }}
    required fields: workspace, repo_slug, target_username
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/default-reviewers/{target_username}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_deploy_keys_key_id:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/deploy-keys/{{ record.key_id }}
    required fields: workspace, repo_slug, key_id
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/deploy-keys/{key_id}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_deployments_config_environments_environment_uui_171d7214:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/deployments_config/environments/{{ record.environment_uuid }}/variables/{{ record.variable_uuid }}
    required fields: workspace, repo_slug, environment_uuid, variable_uuid
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/deployments_config/environments/{environment_uuid}/variables/{variable_uuid}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_downloads_filename:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/downloads/{{ record.filename }}
    required fields: workspace, repo_slug, filename
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/downloads/{filename}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_environments_environment_uuid:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/environments/{{ record.environment_uuid }}
    required fields: workspace, repo_slug, environment_uuid
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/environments/{environment_uuid}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_hooks_uid:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/hooks/{{ record.uid }}
    required fields: workspace, repo_slug, uid
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/hooks/{uid}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_issues_issue_id:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/issues/{{ record.issue_id }}
    required fields: workspace, repo_slug, issue_id
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/issues/{issue_id}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_issues_issue_id_attachments_path:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/issues/{{ record.issue_id }}/attachments/{{ record.path }}
    required fields: workspace, repo_slug, issue_id, path
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/issues/{issue_id}/attachments/{path}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_issues_issue_id_comments_comment_id:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/issues/{{ record.issue_id }}/comments/{{ record.comment_id }}
    required fields: workspace, repo_slug, issue_id, comment_id
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/issues/{issue_id}/comments/{comment_id}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_issues_issue_id_vote:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/issues/{{ record.issue_id }}/vote
    required fields: workspace, repo_slug, issue_id
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/issues/{issue_id}/vote; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_issues_issue_id_watch:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/issues/{{ record.issue_id }}/watch
    required fields: workspace, repo_slug, issue_id
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/issues/{issue_id}/watch; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_permissions_config_groups_group_slug:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/permissions-config/groups/{{ record.group_slug }}
    required fields: workspace, repo_slug, group_slug
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/permissions-config/groups/{group_slug}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_permissions_config_users_selected_user_id:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/permissions-config/users/{{ record.selected_user_id }}
    required fields: workspace, repo_slug, selected_user_id
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/permissions-config/users/{selected_user_id}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_pipelines_config_caches:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/pipelines-config/caches
    required fields: workspace, repo_slug
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/pipelines-config/caches; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_pipelines_config_caches_cache_uuid:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/pipelines-config/caches/{{ record.cache_uuid }}
    required fields: workspace, repo_slug, cache_uuid
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/pipelines-config/caches/{cache_uuid}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_pipelines_config_runners_runner_uuid:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/pipelines-config/runners/{{ record.runner_uuid }}
    required fields: workspace, repo_slug, runner_uuid
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/pipelines-config/runners/{runner_uuid}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_pipelines_config_schedules_schedule_uuid:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/pipelines_config/schedules/{{ record.schedule_uuid }}
    required fields: workspace, repo_slug, schedule_uuid
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/pipelines_config/schedules/{schedule_uuid}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_pipelines_config_ssh_key_pair:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/pipelines_config/ssh/key_pair
    required fields: workspace, repo_slug
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/pipelines_config/ssh/key_pair; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_pipelines_config_ssh_known_hosts_known_host_uuid:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/pipelines_config/ssh/known_hosts/{{ record.known_host_uuid }}
    required fields: workspace, repo_slug, known_host_uuid
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/pipelines_config/ssh/known_hosts/{known_host_uuid}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_pipelines_config_variables_variable_uuid:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/pipelines_config/variables/{{ record.variable_uuid }}
    required fields: workspace, repo_slug, variable_uuid
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/pipelines_config/variables/{variable_uuid}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_properties_app_key_property_name:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/properties/{{ record.app_key }}/{{ record.property_name }}
    required fields: workspace, repo_slug, app_key, property_name
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/properties/{app_key}/{property_name}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_pullrequests_pull_request_id_approve:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/pullrequests/{{ record.pull_request_id }}/approve
    required fields: workspace, repo_slug, pull_request_id
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/pullrequests/{pull_request_id}/approve; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_pullrequests_pull_request_id_comments_comment_id:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/pullrequests/{{ record.pull_request_id }}/comments/{{ record.comment_id }}
    required fields: workspace, repo_slug, pull_request_id, comment_id
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/pullrequests/{pull_request_id}/comments/{comment_id}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_pullrequests_pull_request_id_comments_comment_id_resolve:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/pullrequests/{{ record.pull_request_id }}/comments/{{ record.comment_id }}/resolve
    required fields: workspace, repo_slug, pull_request_id, comment_id
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/pullrequests/{pull_request_id}/comments/{comment_id}/resolve; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_pullrequests_pull_request_id_request_changes:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/pullrequests/{{ record.pull_request_id }}/request-changes
    required fields: workspace, repo_slug, pull_request_id
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/pullrequests/{pull_request_id}/request-changes; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_pullrequests_pull_request_id_tasks_task_id:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/pullrequests/{{ record.pull_request_id }}/tasks/{{ record.task_id }}
    required fields: workspace, repo_slug, pull_request_id, task_id
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/pullrequests/{pull_request_id}/tasks/{task_id}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_pullrequests_pullrequest_id_properties_app_key_629f4f2b:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/pullrequests/{{ record.pullrequest_id }}/properties/{{ record.app_key }}/{{ record.property_name }}
    required fields: workspace, repo_slug, pullrequest_id, app_key, property_name
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/pullrequests/{pullrequest_id}/properties/{app_key}/{property_name}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_refs_branches_name:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/refs/branches/{{ record.name }}
    required fields: workspace, repo_slug, name
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/refs/branches/{name}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_repositories_workspace_repo_slug_refs_tags_name:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}/refs/tags/{{ record.name }}
    required fields: workspace, repo_slug, name
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}/refs/tags/{name}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_snippets_workspace_encoded_id:
    endpoint: DELETE /snippets/{{ record.workspace }}/{{ record.encoded_id }}
    required fields: workspace, encoded_id
    risk: Destructive DELETE /snippets/{workspace}/{encoded_id}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_snippets_workspace_encoded_id_comments_comment_id:
    endpoint: DELETE /snippets/{{ record.workspace }}/{{ record.encoded_id }}/comments/{{ record.comment_id }}
    required fields: workspace, encoded_id, comment_id
    risk: Destructive DELETE /snippets/{workspace}/{encoded_id}/comments/{comment_id}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_snippets_workspace_encoded_id_watch:
    endpoint: DELETE /snippets/{{ record.workspace }}/{{ record.encoded_id }}/watch
    required fields: workspace, encoded_id
    risk: Destructive DELETE /snippets/{workspace}/{encoded_id}/watch; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_snippets_workspace_encoded_id_node_id:
    endpoint: DELETE /snippets/{{ record.workspace }}/{{ record.encoded_id }}/{{ record.node_id }}
    required fields: workspace, encoded_id, node_id
    risk: Destructive DELETE /snippets/{workspace}/{encoded_id}/{node_id}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_teams_username_pipelines_config_variables_variable_uuid:
    endpoint: DELETE /teams/{{ record.username }}/pipelines_config/variables/{{ record.variable_uuid }}
    required fields: username, variable_uuid
    risk: Destructive DELETE /teams/{username}/pipelines_config/variables/{variable_uuid}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_users_selected_user_gpg_keys_fingerprint:
    endpoint: DELETE /users/{{ record.selected_user }}/gpg-keys/{{ record.fingerprint }}
    required fields: selected_user, fingerprint
    risk: Destructive DELETE /users/{selected_user}/gpg-keys/{fingerprint}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_users_selected_user_pipelines_config_variables_variable_uuid:
    endpoint: DELETE /users/{{ record.selected_user }}/pipelines_config/variables/{{ record.variable_uuid }}
    required fields: selected_user, variable_uuid
    risk: Destructive DELETE /users/{selected_user}/pipelines_config/variables/{variable_uuid}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_users_selected_user_properties_app_key_property_name:
    endpoint: DELETE /users/{{ record.selected_user }}/properties/{{ record.app_key }}/{{ record.property_name }}
    required fields: selected_user, app_key, property_name
    risk: Destructive DELETE /users/{selected_user}/properties/{app_key}/{property_name}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_users_selected_user_ssh_keys_key_id:
    endpoint: DELETE /users/{{ record.selected_user }}/ssh-keys/{{ record.key_id }}
    required fields: selected_user, key_id
    risk: Destructive DELETE /users/{selected_user}/ssh-keys/{key_id}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_workspaces_workspace_hooks_uid:
    endpoint: DELETE /workspaces/{{ record.workspace }}/hooks/{{ record.uid }}
    required fields: workspace, uid
    risk: Destructive DELETE /workspaces/{workspace}/hooks/{uid}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_workspaces_workspace_pipelines_config_runners_runner_uuid:
    endpoint: DELETE /workspaces/{{ record.workspace }}/pipelines-config/runners/{{ record.runner_uuid }}
    required fields: workspace, runner_uuid
    risk: Destructive DELETE /workspaces/{workspace}/pipelines-config/runners/{runner_uuid}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_workspaces_workspace_pipelines_config_variables_variable_uuid:
    endpoint: DELETE /workspaces/{{ record.workspace }}/pipelines-config/variables/{{ record.variable_uuid }}
    required fields: workspace, variable_uuid
    risk: Destructive DELETE /workspaces/{workspace}/pipelines-config/variables/{variable_uuid}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_workspaces_workspace_projects_project_key:
    endpoint: DELETE /workspaces/{{ record.workspace }}/projects/{{ record.project_key }}
    required fields: workspace, project_key
    risk: Destructive DELETE /workspaces/{workspace}/projects/{project_key}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_workspaces_workspace_projects_project_key_default_reviewers_selected_user:
    endpoint: DELETE /workspaces/{{ record.workspace }}/projects/{{ record.project_key }}/default-reviewers/{{ record.selected_user }}
    required fields: workspace, project_key, selected_user
    risk: Destructive DELETE /workspaces/{workspace}/projects/{project_key}/default-reviewers/{selected_user}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_workspaces_workspace_projects_project_key_deploy_keys_key_id:
    endpoint: DELETE /workspaces/{{ record.workspace }}/projects/{{ record.project_key }}/deploy-keys/{{ record.key_id }}
    required fields: workspace, project_key, key_id
    risk: Destructive DELETE /workspaces/{workspace}/projects/{project_key}/deploy-keys/{key_id}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_workspaces_workspace_projects_project_key_permissions_config_groups_group_slug:
    endpoint: DELETE /workspaces/{{ record.workspace }}/projects/{{ record.project_key }}/permissions-config/groups/{{ record.group_slug }}
    required fields: workspace, project_key, group_slug
    risk: Destructive DELETE /workspaces/{workspace}/projects/{project_key}/permissions-config/groups/{group_slug}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  delete_workspaces_workspace_projects_project_key_permissions_config_users_selected_user_id:
    endpoint: DELETE /workspaces/{{ record.workspace }}/projects/{{ record.project_key }}/permissions-config/users/{{ record.selected_user_id }}
    required fields: workspace, project_key, selected_user_id
    risk: Destructive DELETE /workspaces/{workspace}/projects/{project_key}/permissions-config/users/{selected_user_id}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.

SECURITY
  read risk: Read-only Bitbucket Cloud REST calls against configured workspaces, repositories, or typed resource identifiers.
  write risk: Closed-schema Bitbucket writes only: repository creation and path-only DELETE actions; untyped JSON-body and multipart mutations stay blocked until typed schemas and bounded transfer foundations exist. Redaction claims apply to newly created plans that persist redact_fields metadata; stale-plan hydration is a shared runtime dependency outside this connector bundle.
  approval: Reverse ETL writes require plan, preview, explicit approval, and typed destructive confirmation for destructive actions.
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Read Bitbucket Cloud resources and plan typed repository mutations safely.
  Usage: pm bitbucket <command> [flags]
  Source CLI: Bitbucket Cloud REST API (OpenAPI 3.0 Bitbucket API 2.0)
  Global flags:
    --credential (string): Credential name to use for Bitbucket requests.
    --config (string_array): Connector config override as key=value.
    --json (boolean): Emit machine-readable JSON output.
    --limit (integer): Maximum ETL records to emit.
    --preview (boolean): Preview a reverse-ETL write without executing it.
    --approve (string): Approval token required to execute a reverse-ETL plan.
    --confirm (string): Typed confirmation for destructive reverse-ETL writes.
  Repositories
    repositories list - List public Bitbucket repositories as paginated ETL records. [intent=etl availability=implemented stream=repositories]
    repositories create - Create or initialize a Bitbucket repository through typed reverse ETL. [intent=reverse_etl availability=implemented write=create_repositories_workspace_repo_slug]; approval: Requires reverse ETL plan -> preview -> explicit approval -> execute.; risk: Creates a Bitbucket repository; reverse ETL plan, preview, explicit approval, execute are required.; flags: --workspace, --repo-slug, --scm, --private
    repositories delete - Delete a Bitbucket repository through a destructive typed reverse-ETL action. [intent=reverse_etl availability=implemented write=delete_repositories_workspace_repo_slug]; approval: Requires reverse ETL plan -> preview -> explicit approval -> execute plus typed destructive confirmation.; risk: Permanently deletes a Bitbucket repository; action has confirm: destructive.; flags: --workspace, --repo-slug
  Reverse ETL writes
  Planned bounded direct/binary operations
  Other Commands
    search code - Planned bounded Bitbucket provider search/query command; blocked pending shared provider-query foundation #2985. [intent=direct_read availability=planned operation=get_teams_username_search_code]; approval: blocked pending #2985; risk: planned bounded provider query; no raw query escape hatch is exposed; notes: Connector-local operation metadata is present, but shared execution foundation is not claimed.; flags: --page, --page-cursor
    downloads get - Planned bounded Bitbucket binary download command; blocked pending binary transfer foundation. [intent=direct_read availability=planned operation=get_repositories_workspace_repo_slug_downloads_filename]; approval: blocked pending bounded binary executor; risk: planned bounded binary transfer; no generic byte-stream command is exposed; notes: Connector-local operation metadata is present, but shared execution foundation is not claimed.; flags: --page, --page-cursor
  Help topics:
    bitbucket-auth - Use OAuth access tokens or username/app-password credentials from the credential store; never pass secrets in command text.
    bitbucket-writes - Implemented Bitbucket writes are closed-schema repository creation and path-only DELETE reverse-ETL actions; untyped JSON-body mutations remain blocked.
    bitbucket-binary-direct - Binary and provider-search operation ledger rows are present but command execution is blocked until the shared bounded-command foundations land.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect bitbucket

  # Inspect as structured JSON
  pm connectors inspect bitbucket --json

AGENT WORKFLOW
  - Run pm connectors inspect bitbucket before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
