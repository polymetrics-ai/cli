# pm connectors inspect bitbucket

```text
NAME
  pm connectors inspect bitbucket - Bitbucket connector manual

SYNOPSIS
  pm connectors inspect bitbucket
  pm connectors inspect bitbucket --json
  pm credentials add <name> --connector bitbucket [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Bitbucket Cloud workspace, repository, pull request, issue, commit, pipeline, deployment, snippet, and project resources; exposes only closed-schema reverse-ETL actions and blocks mutations lacking request-shape evidence.

ICON
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
    primary key: id
    fields: created_on(), description(), fork_policy(), full_name(), has_issues(), has_wiki(), id(), is_private(), language(), links(), mainbranch(), name(), owner(), parent(), project(), scm(), size(), slug(), type(), updated_on(), uuid()
  hook_events:
    primary key: id
    fields: created_on(), id(), links(), name(), repository(), slug(), type(), updated_on(), uuid(), workspace()
  hook_events_subject_type:
    primary key: id
    fields: category(), created_on(), description(), event(), id(), label(), links(), name(), slug(), type(), updated_on(), uuid()
  repositories_workspace:
    primary key: id
    fields: created_on(), description(), fork_policy(), full_name(), has_issues(), has_wiki(), id(), is_private(), language(), links(), mainbranch(), name(), owner(), parent(), project(), scm(), size(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug:
    primary key: id
    fields: created_on(), description(), fork_policy(), full_name(), has_issues(), has_wiki(), id(), is_private(), language(), links(), mainbranch(), name(), owner(), parent(), project(), scm(), size(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_branch_restrictions:
    primary key: id
    fields: branch_match_kind(), branch_type(), created_on(), groups(), id(), kind(), links(), name(), pattern(), slug(), type(), updated_on(), users(), uuid(), value()
  repositories_workspace_repo_slug_branch_restrictions_id:
    primary key: id
    fields: branch_match_kind(), branch_type(), created_on(), groups(), id(), kind(), links(), name(), pattern(), slug(), type(), updated_on(), users(), uuid(), value()
  repositories_workspace_repo_slug_branching_model:
    primary key: id
    fields: branch_types(), created_on(), development(), id(), links(), name(), production(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_branching_model_settings:
    primary key: id
    fields: branch_types(), created_on(), development(), id(), links(), name(), production(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_commit_commit:
    primary key: id
    fields: author(), committer(), created_on(), date(), hash(), id(), links(), message(), name(), parents(), participants(), repository(), slug(), summary(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_commit_commit_comments:
    primary key: id
    fields: commit(), content(), created_on(), deleted(), id(), inline(), links(), name(), parent(), slug(), type(), updated_on(), user(), uuid()
  repositories_workspace_repo_slug_commit_commit_comments_comment_id:
    primary key: id
    fields: commit(), content(), created_on(), deleted(), id(), inline(), links(), name(), parent(), slug(), type(), updated_on(), user(), uuid()
  repositories_workspace_repo_slug_commit_commit_properties_app_key_property_name:
    primary key: id
    fields: _attributes(), created_on(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_commit_commit_pullrequests:
    primary key: id
    fields: author(), close_source_branch(), closed_by(), comment_count(), created_on(), destination(), draft(), id(), links(), merge_commit(), name(), participants(), queued(), reason(), rendered(), reviewers(), slug(), source(), state(), summary(), task_count(), title(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_commit_commit_reports:
    primary key: id
    fields: created_on(), data(), details(), external_id(), id(), link(), links(), logo_url(), name(), remote_link_enabled(), report_type(), reporter(), result(), slug(), title(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_commit_commit_reports_reportid:
    primary key: id
    fields: created_on(), data(), details(), external_id(), id(), link(), links(), logo_url(), name(), remote_link_enabled(), report_type(), reporter(), result(), slug(), title(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_commit_commit_reports_reportid_annotations:
    primary key: id
    fields: annotation_type(), created_on(), details(), external_id(), id(), line(), link(), links(), name(), path(), result(), severity(), slug(), summary(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_commit_commit_reports_reportid_annotations_annotationid:
    primary key: id
    fields: annotation_type(), created_on(), details(), external_id(), id(), line(), link(), links(), name(), path(), result(), severity(), slug(), summary(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_commit_commit_statuses:
    primary key: id
    fields: created_on(), description(), id(), key(), links(), name(), refname(), slug(), state(), type(), updated_on(), url(), uuid()
  repositories_workspace_repo_slug_commit_commit_statuses_build_key:
    primary key: id
    fields: created_on(), description(), id(), key(), links(), name(), refname(), slug(), state(), type(), updated_on(), url(), uuid()
  repositories_workspace_repo_slug_commits:
    primary key: id
    fields: author(), committer(), created_on(), date(), hash(), id(), links(), message(), name(), parents(), slug(), summary(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_commits_revision:
    primary key: id
    fields: author(), committer(), created_on(), date(), hash(), id(), links(), message(), name(), parents(), slug(), summary(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_components:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_components_component_id:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_default_reviewers:
    primary key: id
    fields: created_on(), display_name(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_default_reviewers_target_username:
    primary key: id
    fields: created_on(), display_name(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_deploy_keys:
    primary key: id
    fields: added_on(), comment(), created_on(), id(), key(), label(), last_used(), links(), name(), owner(), repository(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_deploy_keys_key_id:
    primary key: id
    fields: added_on(), comment(), created_on(), id(), key(), label(), last_used(), links(), name(), owner(), repository(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_deployments:
    primary key: id
    fields: created_on(), environment(), id(), links(), name(), release(), slug(), state(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_deployments_config_environments_environment_uuid_variables:
    primary key: id
    fields: created_on(), id(), key(), links(), name(), secured(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_deployments_deployment_uuid:
    primary key: id
    fields: created_on(), environment(), id(), links(), name(), release(), slug(), state(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_diffstat_spec:
    primary key: id
    fields: created_on(), id(), lines_added(), lines_removed(), links(), name(), new(), old(), slug(), status(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_downloads:
    primary key: name
    fields: created_on(), id(), links(), name(), size(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_effective_branching_model:
    primary key: id
    fields: branch_types(), created_on(), development(), id(), links(), name(), production(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_effective_default_reviewers:
    primary key: id
    fields: created_on(), id(), links(), name(), reviewer_type(), slug(), type(), updated_on(), user(), uuid()
  repositories_workspace_repo_slug_environments:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_environments_environment_uuid:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_file_conflicts_spec:
    primary key: id
    fields: created_on(), id(), links(), message(), name(), path(), scenario(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_filehistory_commit_path:
    primary key: id
    fields: attributes(), commit(), created_on(), escaped_path(), id(), links(), name(), path(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_forks:
    primary key: id
    fields: created_on(), description(), fork_policy(), full_name(), has_issues(), has_wiki(), id(), is_private(), language(), links(), mainbranch(), name(), owner(), parent(), project(), scm(), size(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_hooks:
    primary key: id
    fields: active(), created_at(), created_on(), description(), events(), id(), links(), name(), secret_set(), slug(), subject(), subject_type(), type(), updated_on(), url(), uuid()
  repositories_workspace_repo_slug_hooks_uid:
    primary key: id
    fields: active(), created_at(), created_on(), description(), events(), id(), links(), name(), secret_set(), slug(), subject(), subject_type(), type(), updated_on(), url(), uuid()
  repositories_workspace_repo_slug_issues:
    primary key: id
    fields: assignee(), component(), content(), created_on(), edited_on(), id(), kind(), links(), milestone(), name(), priority(), reporter(), repository(), slug(), state(), title(), type(), updated_on(), uuid(), version(), votes()
  repositories_workspace_repo_slug_issues_import:
    primary key: id
    fields: count(), created_on(), id(), links(), name(), pct(), phase(), slug(), status(), total(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_issues_issue_id:
    primary key: id
    fields: assignee(), component(), content(), created_on(), edited_on(), id(), kind(), links(), milestone(), name(), priority(), reporter(), repository(), slug(), state(), title(), type(), updated_on(), uuid(), version(), votes()
  repositories_workspace_repo_slug_issues_issue_id_attachments:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_issues_issue_id_changes:
    primary key: id
    fields: changes(), created_on(), id(), issue(), links(), message(), name(), slug(), type(), updated_on(), user(), uuid()
  repositories_workspace_repo_slug_issues_issue_id_changes_change_id:
    primary key: id
    fields: changes(), created_on(), id(), issue(), links(), message(), name(), slug(), type(), updated_on(), user(), uuid()
  repositories_workspace_repo_slug_issues_issue_id_comments:
    primary key: id
    fields: content(), created_on(), deleted(), id(), inline(), issue(), links(), name(), parent(), slug(), type(), updated_on(), user(), uuid()
  repositories_workspace_repo_slug_issues_issue_id_comments_comment_id:
    primary key: id
    fields: content(), created_on(), deleted(), id(), inline(), issue(), links(), name(), parent(), slug(), type(), updated_on(), user(), uuid()
  repositories_workspace_repo_slug_merge_base_revspec:
    primary key: id
    fields: author(), committer(), created_on(), date(), hash(), id(), links(), message(), name(), parents(), participants(), repository(), slug(), summary(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_milestones:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_milestones_milestone_id:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_override_settings:
    primary key: id
    fields: created_on(), id(), links(), name(), override_settings(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_permissions_config_groups:
    primary key: id
    fields: created_on(), group(), id(), links(), name(), permission(), repository(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_permissions_config_groups_group_slug:
    primary key: id
    fields: created_on(), group(), id(), links(), name(), permission(), repository(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_permissions_config_users:
    primary key: id
    fields: created_on(), id(), links(), name(), permission(), repository(), slug(), type(), updated_on(), user(), uuid()
  repositories_workspace_repo_slug_permissions_config_users_selected_user_id:
    primary key: id
    fields: created_on(), id(), links(), name(), permission(), repository(), slug(), type(), updated_on(), user(), uuid()
  repositories_workspace_repo_slug_pipelines:
    primary key: id
    fields: build_number(), build_seconds_used(), completed_on(), configuration_sources(), created_on(), creator(), id(), links(), name(), repository(), slug(), state(), target(), trigger(), type(), updated_on(), uuid(), variables()
  repositories_workspace_repo_slug_pipelines_config:
    primary key: id
    fields: created_on(), enabled(), id(), links(), name(), repository(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pipelines_config_caches:
    primary key: id
    fields: created_on(), file_size_bytes(), id(), key_hash(), links(), name(), path(), pipeline_uuid(), slug(), step_uuid(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pipelines_config_caches_cache_uuid_content_uri:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), type(), updated_on(), uri(), uuid()
  repositories_workspace_repo_slug_pipelines_config_runners:
    primary key: id
    fields: created_on(), id(), labels(), links(), name(), oauth_client(), slug(), state(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pipelines_config_runners_runner_uuid:
    primary key: id
    fields: created_on(), id(), labels(), links(), name(), oauth_client(), slug(), state(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pipelines_config_schedules:
    primary key: id
    fields: created_on(), cron_pattern(), enabled(), id(), links(), name(), slug(), target(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pipelines_config_schedules_schedule_uuid:
    primary key: id
    fields: created_on(), cron_pattern(), enabled(), id(), links(), name(), slug(), target(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pipelines_config_schedules_schedule_uuid_executions:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pipelines_config_ssh_key_pair:
    primary key: id
    fields: created_on(), id(), links(), name(), public_key(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pipelines_config_ssh_known_hosts:
    primary key: id
    fields: created_on(), hostname(), id(), links(), name(), public_key(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pipelines_config_ssh_known_hosts_known_host_uuid:
    primary key: id
    fields: created_on(), hostname(), id(), links(), name(), public_key(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pipelines_config_variables:
    primary key: id
    fields: created_on(), id(), key(), links(), name(), secured(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pipelines_config_variables_variable_uuid:
    primary key: id
    fields: created_on(), id(), key(), links(), name(), secured(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pipelines_pipeline_uuid:
    primary key: id
    fields: build_number(), build_seconds_used(), completed_on(), configuration_sources(), created_on(), creator(), id(), links(), name(), repository(), slug(), state(), target(), trigger(), type(), updated_on(), uuid(), variables()
  repositories_workspace_repo_slug_pipelines_pipeline_uuid_steps:
    primary key: id
    fields: completed_on(), created_on(), id(), image(), links(), name(), script_commands(), setup_commands(), slug(), started_on(), state(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pipelines_pipeline_uuid_steps_step_uuid:
    primary key: id
    fields: completed_on(), created_on(), id(), image(), links(), name(), script_commands(), setup_commands(), slug(), started_on(), state(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_properties_app_key_property_name:
    primary key: id
    fields: _attributes(), created_on(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pullrequests:
    primary key: id
    fields: author(), close_source_branch(), closed_by(), comment_count(), created_on(), destination(), draft(), id(), links(), merge_commit(), name(), participants(), queued(), reason(), rendered(), reviewers(), slug(), source(), state(), summary(), task_count(), title(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pullrequests_pull_request_id:
    primary key: id
    fields: author(), close_source_branch(), closed_by(), comment_count(), created_on(), destination(), draft(), id(), links(), merge_commit(), name(), participants(), queued(), reason(), rendered(), reviewers(), slug(), source(), state(), summary(), task_count(), title(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pullrequests_pull_request_id_commits:
    primary key: hash
    fields: author(), date(), hash(), id(), links(), message(), parents(), repository(), summary(), type()
  repositories_workspace_repo_slug_pullrequests_pull_request_id_diffstat:
    primary key: id
    fields: created_on(), id(), lines_added(), lines_removed(), links(), name(), new(), old(), slug(), status(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pullrequests_pull_request_id_comments:
    primary key: id
    fields: content(), created_on(), deleted(), id(), inline(), links(), name(), parent(), pending(), pullrequest(), resolution(), slug(), type(), updated_on(), user(), uuid()
  repositories_workspace_repo_slug_pullrequests_pull_request_id_comments_comment_id:
    primary key: id
    fields: content(), created_on(), deleted(), id(), inline(), links(), name(), parent(), pending(), pullrequest(), resolution(), slug(), type(), updated_on(), user(), uuid()
  repositories_workspace_repo_slug_pullrequests_pull_request_id_statuses:
    primary key: id
    fields: created_on(), description(), id(), key(), links(), name(), refname(), slug(), state(), type(), updated_on(), url(), uuid()
  repositories_workspace_repo_slug_pullrequests_pull_request_id_tasks:
    primary key: id
    fields: comment(), content(), created_on(), creator(), id(), links(), name(), pending(), resolved_by(), resolved_on(), slug(), state(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pullrequests_pull_request_id_tasks_task_id:
    primary key: id
    fields: comment(), content(), created_on(), creator(), id(), links(), name(), pending(), resolved_by(), resolved_on(), slug(), state(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_pullrequests_pullrequest_id_properties_app_key_property_name:
    primary key: id
    fields: _attributes(), created_on(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_refs:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), target(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_refs_branches:
    primary key: id
    fields: created_on(), default_merge_strategy(), id(), links(), merge_strategies(), name(), slug(), target(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_refs_branches_name:
    primary key: id
    fields: created_on(), default_merge_strategy(), id(), links(), merge_strategies(), name(), slug(), target(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_refs_tags:
    primary key: id
    fields: created_on(), date(), id(), links(), message(), name(), slug(), tagger(), target(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_refs_tags_name:
    primary key: id
    fields: created_on(), date(), id(), links(), message(), name(), slug(), tagger(), target(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_src:
    primary key: id
    fields: commit(), created_on(), id(), links(), name(), path(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_versions:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_versions_version_id:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  repositories_workspace_repo_slug_watchers:
    primary key: id
    fields: created_on(), display_name(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  snippets:
    primary key: id
    fields: created_on(), creator(), id(), is_private(), links(), name(), owner(), scm(), slug(), title(), type(), updated_on(), uuid()
  snippets_workspace:
    primary key: id
    fields: created_on(), creator(), id(), is_private(), links(), name(), owner(), scm(), slug(), title(), type(), updated_on(), uuid()
  snippets_workspace_encoded_id:
    primary key: id
    fields: created_on(), creator(), id(), is_private(), links(), name(), owner(), scm(), slug(), title(), type(), updated_on(), uuid()
  snippets_workspace_encoded_id_comments:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), snippet(), type(), updated_on(), uuid()
  snippets_workspace_encoded_id_comments_comment_id:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), snippet(), type(), updated_on(), uuid()
  snippets_workspace_encoded_id_commits:
    primary key: id
    fields: author(), committer(), created_on(), date(), hash(), id(), links(), message(), name(), parents(), slug(), snippet(), summary(), type(), updated_on(), uuid()
  snippets_workspace_encoded_id_commits_revision:
    primary key: id
    fields: author(), committer(), created_on(), date(), hash(), id(), links(), message(), name(), parents(), slug(), snippet(), summary(), type(), updated_on(), uuid()
  snippets_workspace_encoded_id_node_id:
    primary key: id
    fields: created_on(), creator(), id(), is_private(), links(), name(), owner(), scm(), slug(), title(), type(), updated_on(), uuid()
  snippets_workspace_encoded_id_watchers:
    primary key: id
    fields: created_on(), display_name(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  teams_username_pipelines_config_variables:
    primary key: id
    fields: created_on(), id(), key(), links(), name(), secured(), slug(), type(), updated_on(), uuid()
  teams_username_pipelines_config_variables_variable_uuid:
    primary key: id
    fields: created_on(), id(), key(), links(), name(), secured(), slug(), type(), updated_on(), uuid()
  user:
    primary key: id
    fields: created_on(), display_name(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  user_emails:
    primary key: id
    fields: created_on(), error(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  user_emails_email:
    primary key: id
    fields: created_on(), error(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  user_permissions_repositories:
    primary key: id
    fields: created_on(), id(), links(), name(), permission(), repository(), slug(), type(), updated_on(), user(), uuid()
  user_permissions_workspaces:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), type(), updated_on(), user(), uuid(), workspace()
  user_workspaces:
    primary key: id
    fields: administrator(), created_on(), id(), links(), name(), slug(), type(), updated_on(), uuid(), workspace()
  user_workspaces_workspace_permission:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), type(), updated_on(), user(), uuid(), workspace()
  user_workspaces_workspace_permissions_repositories:
    primary key: id
    fields: created_on(), id(), links(), name(), permission(), repository(), slug(), type(), updated_on(), user(), uuid()
  users_selected_user:
    primary key: id
    fields: created_on(), display_name(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  users_selected_user_gpg_keys:
    primary key: id
    fields: added_on(), created_on(), expires_on(), fingerprint(), id(), key(), key_id(), last_used(), links(), name(), owner(), parent_fingerprint(), slug(), subkeys(), type(), updated_on(), uuid()
  users_selected_user_gpg_keys_fingerprint:
    primary key: id
    fields: added_on(), created_on(), expires_on(), fingerprint(), id(), key(), key_id(), last_used(), links(), name(), owner(), parent_fingerprint(), slug(), subkeys(), type(), updated_on(), uuid()
  users_selected_user_pipelines_config_variables:
    primary key: id
    fields: created_on(), id(), key(), links(), name(), secured(), slug(), type(), updated_on(), uuid()
  users_selected_user_pipelines_config_variables_variable_uuid:
    primary key: id
    fields: created_on(), id(), key(), links(), name(), secured(), slug(), type(), updated_on(), uuid()
  users_selected_user_properties_app_key_property_name:
    primary key: id
    fields: _attributes(), created_on(), id(), links(), name(), slug(), type(), updated_on(), uuid()
  users_selected_user_ssh_keys:
    primary key: id
    fields: comment(), created_on(), expires_on(), fingerprint(), id(), key(), label(), last_used(), links(), name(), owner(), slug(), type(), updated_on(), uuid()
  users_selected_user_ssh_keys_key_id:
    primary key: id
    fields: comment(), created_on(), expires_on(), fingerprint(), id(), key(), label(), last_used(), links(), name(), owner(), slug(), type(), updated_on(), uuid()
  workspaces:
    primary key: id
    fields: created_on(), forking_mode(), id(), is_privacy_enforced(), is_private(), links(), name(), slug(), type(), updated_on(), uuid()
  workspaces_workspace:
    primary key: id
    fields: created_on(), forking_mode(), id(), is_privacy_enforced(), is_private(), links(), name(), slug(), type(), updated_on(), uuid()
  workspaces_workspace_hooks:
    primary key: id
    fields: active(), created_at(), created_on(), description(), events(), id(), links(), name(), secret_set(), slug(), subject(), subject_type(), type(), updated_on(), url(), uuid()
  workspaces_workspace_hooks_uid:
    primary key: id
    fields: active(), created_at(), created_on(), description(), events(), id(), links(), name(), secret_set(), slug(), subject(), subject_type(), type(), updated_on(), url(), uuid()
  workspaces_workspace_members:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), type(), updated_on(), user(), uuid(), workspace()
  workspaces_workspace_members_member:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), type(), updated_on(), user(), uuid(), workspace()
  workspaces_workspace_permissions:
    primary key: id
    fields: created_on(), id(), links(), name(), slug(), type(), updated_on(), user(), uuid(), workspace()
  workspaces_workspace_permissions_repositories:
    primary key: id
    fields: created_on(), id(), links(), name(), permission(), repository(), slug(), type(), updated_on(), user(), uuid()
  workspaces_workspace_permissions_repositories_repo_slug:
    primary key: id
    fields: created_on(), id(), links(), name(), permission(), repository(), slug(), type(), updated_on(), user(), uuid()
  workspaces_workspace_pipelines_config_runners:
    primary key: id
    fields: created_on(), id(), labels(), links(), name(), oauth_client(), slug(), state(), type(), updated_on(), uuid()
  workspaces_workspace_pipelines_config_runners_runner_uuid:
    primary key: id
    fields: created_on(), id(), labels(), links(), name(), oauth_client(), slug(), state(), type(), updated_on(), uuid()
  workspaces_workspace_pipelines_config_variables:
    primary key: id
    fields: created_on(), id(), key(), links(), name(), secured(), slug(), type(), updated_on(), uuid()
  workspaces_workspace_pipelines_config_variables_variable_uuid:
    primary key: id
    fields: created_on(), id(), key(), links(), name(), secured(), slug(), type(), updated_on(), uuid()
  workspaces_workspace_projects:
    primary key: id
    fields: created_on(), description(), has_publicly_visible_repos(), id(), is_private(), key(), links(), name(), owner(), slug(), type(), updated_on(), uuid()
  workspaces_workspace_projects_project_key:
    primary key: id
    fields: created_on(), description(), has_publicly_visible_repos(), id(), is_private(), key(), links(), name(), owner(), slug(), type(), updated_on(), uuid()
  workspaces_workspace_projects_project_key_branching_model:
    primary key: id
    fields: branch_types(), created_on(), development(), id(), links(), name(), production(), slug(), type(), updated_on(), uuid()
  workspaces_workspace_projects_project_key_branching_model_settings:
    primary key: id
    fields: branch_types(), created_on(), development(), id(), links(), name(), production(), slug(), type(), updated_on(), uuid()
  workspaces_workspace_projects_project_key_default_reviewers:
    primary key: id
    fields: created_on(), id(), links(), name(), reviewer_type(), slug(), type(), updated_on(), user(), uuid()
  workspaces_workspace_projects_project_key_default_reviewers_selected_user:
    primary key: id
    fields: account_id(), account_status(), created_on(), display_name(), has_2fa_enabled(), id(), is_staff(), links(), name(), nickname(), slug(), type(), updated_on(), uuid()
  workspaces_workspace_projects_project_key_deploy_keys:
    primary key: id
    fields: added_on(), comment(), created_by(), created_on(), id(), key(), label(), last_used(), links(), name(), project(), slug(), type(), updated_on(), uuid()
  workspaces_workspace_projects_project_key_deploy_keys_key_id:
    primary key: id
    fields: added_on(), comment(), created_by(), created_on(), id(), key(), label(), last_used(), links(), name(), project(), slug(), type(), updated_on(), uuid()
  workspaces_workspace_projects_project_key_permissions_config_groups:
    primary key: id
    fields: created_on(), group(), id(), links(), name(), permission(), project(), slug(), type(), updated_on(), uuid()
  workspaces_workspace_projects_project_key_permissions_config_groups_group_slug:
    primary key: id
    fields: created_on(), group(), id(), links(), name(), permission(), project(), slug(), type(), updated_on(), uuid()
  workspaces_workspace_projects_project_key_permissions_config_users:
    primary key: id
    fields: created_on(), id(), links(), name(), permission(), project(), slug(), type(), updated_on(), user(), uuid()
  workspaces_workspace_projects_project_key_permissions_config_users_selected_user_id:
    primary key: id
    fields: created_on(), id(), links(), name(), permission(), project(), slug(), type(), updated_on(), user(), uuid()
  workspaces_workspace_pullrequests_selected_user:
    primary key: id
    fields: author(), close_source_branch(), closed_by(), comment_count(), created_on(), destination(), draft(), id(), links(), merge_commit(), name(), participants(), queued(), reason(), rendered(), reviewers(), slug(), source(), state(), summary(), task_count(), title(), type(), updated_on(), uuid()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  delete_repositories_workspace_repo_slug:
    endpoint: DELETE /repositories/{{ record.workspace }}/{{ record.repo_slug }}
    required fields: workspace, repo_slug
    risk: Destructive DELETE /repositories/{workspace}/{repo_slug}; requires typed destructive confirmation plus reverse ETL plan, preview, explicit approval, execute.
  create_repositories_workspace_repo_slug:
    endpoint: POST /repositories/{{ record.workspace }}/{{ record.repo_slug }}
    required fields: workspace, repo_slug, scm, is_private
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
  write risk: Closed-schema Bitbucket Cloud mutations only; DELETE actions carry destructive confirmation and writes lacking request-shape evidence stay blocked.
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
    search code - Planned bounded Bitbucket provider search/query command; blocked pending shared provider-query foundation #2985. [intent=direct_read availability=planned]; approval: blocked pending #2985; risk: planned bounded provider query; no raw query escape hatch is exposed; notes: Connector-local operation metadata is present, but shared execution foundation is not claimed.
    downloads get - Planned bounded Bitbucket download artifact command; blocked pending binary transfer foundation. [intent=direct_read availability=planned]; approval: blocked pending bounded binary executor; risk: planned bounded artifact download; no generic byte-stream command is exposed; notes: Connector-local operation metadata is present, but shared execution foundation is not claimed.
  Help topics:
    bitbucket-auth - Use OAuth access tokens or username/app-password credentials from the credential store; never pass secrets in command text.
    bitbucket-writes - Bitbucket mutations are typed reverse-ETL actions with plan, preview, explicit approval, execute, and destructive confirmation for DELETE actions.
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
