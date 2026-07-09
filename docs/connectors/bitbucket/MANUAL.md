# pm connectors inspect bitbucket

```text
NAME
  pm connectors inspect bitbucket - Bitbucket connector manual

SYNOPSIS
  pm connectors inspect bitbucket
  pm connectors inspect bitbucket --json
  pm credentials add <name> --connector bitbucket [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Bitbucket Cloud repositories, branches, commits, tags, pull requests, issues, pipelines, deployments, downloads metadata, webhooks, branch restrictions, projects, and snippets; exposes approval-gated write plans for selected repository, issue, pull request, pipeline, webhook, branch restriction, and snippet mutations.

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
  base_url
  repo_slug
  start_date
  workspace
  access_token (secret)

ETL STREAMS
  repositories:
    primary key: uuid
    fields: created_on(), description(), fork_policy(), full_name(), id(), is_private(), language(), links(), mainbranch(), name(), project(), repository(), scm(), slug(), state(), title(), type(), updated_on(), uuid(), workspace()
  branches:
    primary key: name
    fields: created_on(), default_merge_strategy(), description(), full_name(), id(), links(), merge_strategies(), name(), repository(), slug(), state(), target(), target_hash(), title(), type(), updated_on(), uuid(), workspace()
  commits:
    primary key: hash
    cursor: date
    fields: author(), created_on(), date(), description(), full_name(), hash(), id(), links(), message(), name(), parents(), repository(), slug(), state(), summary(), title(), type(), updated_on(), uuid(), workspace()
  tags:
    primary key: name
    fields: created_on(), description(), full_name(), id(), links(), name(), repository(), slug(), state(), target(), target_hash(), title(), type(), updated_on(), uuid(), workspace()
  pull_requests:
    primary key: id
    cursor: updated_on
    fields: author(), author_display_name(), close_source_branch(), comment_count(), created_on(), description(), destination(), destination_branch(), full_name(), id(), links(), name(), participants(), repository(), reviewers(), slug(), source(), source_branch(), state(), summary(), task_count(), title(), type(), updated_on(), uuid(), workspace()
  issues:
    primary key: id
    cursor: updated_on
    fields: assignee(), assignee_display_name(), component(), content(), created_on(), description(), full_name(), id(), kind(), links(), milestone(), name(), priority(), reporter(), reporter_display_name(), repository(), slug(), state(), title(), type(), updated_on(), uuid(), version(), votes(), watches(), workspace()
  pipelines:
    primary key: uuid
    cursor: created_on
    fields: build_number(), completed_on(), created_on(), creator(), description(), duration_in_seconds(), full_name(), id(), links(), name(), repository(), slug(), state(), target(), title(), trigger(), type(), updated_on(), uuid(), workspace()
  deployments:
    primary key: uuid
    cursor: created_on
    fields: created_on(), deployment_state(), description(), environment(), full_name(), id(), last_update_time(), links(), name(), release(), repository(), slug(), state(), title(), type(), updated_on(), uuid(), workspace()
  downloads:
    primary key: name
    fields: created_on(), description(), full_name(), id(), links(), name(), repository(), size(), slug(), state(), title(), type(), updated_on(), user(), uuid(), workspace()
  webhooks:
    primary key: uuid
    fields: active(), created_on(), description(), events(), full_name(), id(), links(), name(), repository(), slug(), state(), subject_type(), title(), type(), updated_on(), url(), uuid(), workspace()
  branch_restrictions:
    primary key: id
    fields: branch_match_kind(), created_on(), description(), full_name(), groups(), id(), kind(), links(), name(), pattern(), repository(), slug(), state(), title(), type(), updated_on(), users(), uuid(), value(), workspace()
  projects:
    primary key: key
    cursor: updated_on
    fields: created_on(), description(), full_name(), id(), is_private(), key(), links(), name(), repository(), slug(), state(), title(), type(), updated_on(), uuid(), workspace()
  snippets:
    primary key: id
    cursor: updated_on
    fields: created_on(), creator(), description(), files(), full_name(), id(), is_private(), links(), name(), owner(), repository(), scm(), slug(), state(), title(), type(), updated_on(), uuid(), workspace()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_repository:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}
    risk: creates a Bitbucket repository in the configured workspace
  create_issue:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/issues
    risk: creates a visible Bitbucket issue and may notify repository participants
  update_issue:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/issues/{{ record.issue_id }}
    required fields: issue_id
    risk: mutates an existing Bitbucket issue
  create_pull_request:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests
    risk: creates a visible pull request and may notify reviewers
  merge_pull_request:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pull_request_id }}/merge
    required fields: pull_request_id
    risk: merges a pull request into its destination branch
  decline_pull_request:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pull_request_id }}/decline
    required fields: pull_request_id
    risk: declines a Bitbucket pull request
  run_pipeline:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines
    risk: starts a Bitbucket pipeline run that may consume CI minutes and deploy artifacts
  stop_pipeline:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines/{{ record.pipeline_uuid }}/stopPipeline
    required fields: pipeline_uuid
    risk: stops an in-flight Bitbucket pipeline
  create_webhook:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/hooks
    risk: creates an outbound webhook that sends repository events to an external URL
  delete_webhook:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/hooks/{{ record.uid }}
    required fields: uid
    risk: deletes a repository webhook and may interrupt downstream automation
  create_branch_restriction:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/branch-restrictions
    risk: changes repository branch protection behavior
  create_snippet:
    endpoint: POST /snippets/{{ config.workspace }}
    risk: creates a Bitbucket snippet that may publish code or text content
  op_put_addon:
    endpoint: PUT /addon
    risk: Update an installed app
  op_delete_addon:
    endpoint: DELETE /addon
    risk: Delete an app
  op_put_repositories_workspace_repo_slug:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}
    risk: Update a repository
  op_delete_repositories_workspace_repo_slug:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}
    risk: Delete a repository
  op_put_repositories_workspace_repo_slug_branch_restrictions_id:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/branch-restrictions/{{ record.id }}
    required fields: id
    risk: Update a branch restriction rule
  op_delete_repositories_workspace_repo_slug_branch_restrictions_id:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/branch-restrictions/{{ record.id }}
    required fields: id
    risk: Delete a branch restriction rule
  op_put_repositories_workspace_repo_slug_branching_model_settings:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/branching-model/settings
    risk: Update the branching model config for a repository
  op_post_repositories_workspace_repo_slug_commit_commit_approve:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/commit/{{ record.commit }}/approve
    required fields: commit
    risk: Approve a commit
  op_delete_repositories_workspace_repo_slug_commit_commit_approve:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/commit/{{ record.commit }}/approve
    required fields: commit
    risk: Unapprove a commit
  op_post_repositories_workspace_repo_slug_commit_commit_comments:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/commit/{{ record.commit }}/comments
    required fields: commit
    risk: Create comment for a commit
  op_put_repositories_workspace_repo_slug_commit_commit_comments_comment_id:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/commit/{{ record.commit }}/comments/{{ record.comment_id }}
    required fields: commit, comment_id
    risk: Update a commit comment
  op_delete_repositories_workspace_repo_slug_commit_commit_comments_comment_id:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/commit/{{ record.commit }}/comments/{{ record.comment_id }}
    required fields: commit, comment_id
    risk: Delete a commit comment
  op_put_repositories_workspace_repo_slug_commit_commit_properties_app_key_property_name:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/commit/{{ record.commit }}/properties/{{ record.app_key }}/{{ record.property_name }}
    required fields: commit, app_key, property_name
    risk: Update a commit application property
  op_delete_repositories_workspace_repo_slug_commit_commit_properties_app_key_property_name:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/commit/{{ record.commit }}/properties/{{ record.app_key }}/{{ record.property_name }}
    required fields: commit, app_key, property_name
    risk: Delete a commit application property
  op_put_repositories_workspace_repo_slug_commit_commit_reports_reportid:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/commit/{{ record.commit }}/reports/{{ record.reportId }}
    required fields: commit, reportId
    risk: Create or update a report
  op_delete_repositories_workspace_repo_slug_commit_commit_reports_reportid:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/commit/{{ record.commit }}/reports/{{ record.reportId }}
    required fields: commit, reportId
    risk: Delete a report
  op_post_repositories_workspace_repo_slug_commit_commit_reports_reportid_annotations:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/commit/{{ record.commit }}/reports/{{ record.reportId }}/annotations
    required fields: commit, reportId
    risk: Bulk create or update annotations
  op_put_repositories_workspace_repo_slug_commit_commit_reports_reportid_annotations_annotationid:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/commit/{{ record.commit }}/reports/{{ record.reportId }}/annotations/{{ record.annotationId }}
    required fields: commit, reportId, annotationId
    risk: Create or update an annotation
  op_delete_repositories_workspace_repo_slug_commit_commit_reports_reportid_annotations_annotationid:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/commit/{{ record.commit }}/reports/{{ record.reportId }}/annotations/{{ record.annotationId }}
    required fields: commit, reportId, annotationId
    risk: Delete an annotation
  op_post_repositories_workspace_repo_slug_commit_commit_statuses_build:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/commit/{{ record.commit }}/statuses/build
    required fields: commit
    risk: Create a build status for a commit
  op_put_repositories_workspace_repo_slug_commit_commit_statuses_build_key:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/commit/{{ record.commit }}/statuses/build/{{ record.key }}
    required fields: commit, key
    risk: Update a build status for a commit
  op_post_repositories_workspace_repo_slug_commits:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/commits
    risk: List commits with include/exclude
  op_post_repositories_workspace_repo_slug_commits_revision:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/commits/{{ record.revision }}
    required fields: revision
    risk: List commits for revision using include/exclude
  op_put_repositories_workspace_repo_slug_default_reviewers_target_username:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/default-reviewers/{{ record.target_username }}
    required fields: target_username
    risk: Add a user to the default reviewers
  op_delete_repositories_workspace_repo_slug_default_reviewers_target_username:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/default-reviewers/{{ record.target_username }}
    required fields: target_username
    risk: Remove a user from the default reviewers
  op_post_repositories_workspace_repo_slug_deploy_keys:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/deploy-keys
    risk: Add a repository deploy key
  op_put_repositories_workspace_repo_slug_deploy_keys_key_id:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/deploy-keys/{{ record.key_id }}
    required fields: key_id
    risk: Update a repository deploy key
  op_delete_repositories_workspace_repo_slug_deploy_keys_key_id:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/deploy-keys/{{ record.key_id }}
    required fields: key_id
    risk: Delete a repository deploy key
  op_post_repositories_workspace_repo_slug_deployments_config_environments_environment_uuid_variables:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/deployments_config/environments/{{ record.environment_uuid }}/variables
    required fields: environment_uuid
    risk: Create a variable for an environment
  op_put_repositories_workspace_repo_slug_deployments_config_environments_environment_uuid_variables_variable_uuid:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/deployments_config/environments/{{ record.environment_uuid }}/variables/{{ record.variable_uuid }}
    required fields: environment_uuid, variable_uuid
    risk: Update a variable for an environment
  op_delete_repositories_workspace_repo_slug_deployments_config_environments_environment_uuid_variables_variable_uuid:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/deployments_config/environments/{{ record.environment_uuid }}/variables/{{ record.variable_uuid }}
    required fields: environment_uuid, variable_uuid
    risk: Delete a variable for an environment
  op_post_repositories_workspace_repo_slug_downloads:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/downloads
    risk: Upload a download artifact
  op_delete_repositories_workspace_repo_slug_downloads_filename:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/downloads/{{ record.filename }}
    required fields: filename
    risk: Delete a download artifact
  op_post_repositories_workspace_repo_slug_environments:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/environments
    risk: Create an environment
  op_delete_repositories_workspace_repo_slug_environments_environment_uuid:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/environments/{{ record.environment_uuid }}
    required fields: environment_uuid
    risk: Delete an environment
  op_post_repositories_workspace_repo_slug_environments_environment_uuid_changes:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/environments/{{ record.environment_uuid }}/changes
    required fields: environment_uuid
    risk: Update an environment
  op_post_repositories_workspace_repo_slug_forks:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/forks
    risk: Fork a repository
  op_put_repositories_workspace_repo_slug_hooks_uid:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/hooks/{{ record.uid }}
    required fields: uid
    risk: Update a webhook for a repository
  op_post_repositories_workspace_repo_slug_issues_export:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/issues/export
    risk: Export issues
  op_post_repositories_workspace_repo_slug_issues_import:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/issues/import
    risk: Import issues
  op_delete_repositories_workspace_repo_slug_issues_issue_id:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/issues/{{ record.issue_id }}
    required fields: issue_id
    risk: Delete an issue
  op_post_repositories_workspace_repo_slug_issues_issue_id_attachments:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/issues/{{ record.issue_id }}/attachments
    required fields: issue_id
    risk: Upload an attachment to an issue
  op_delete_repositories_workspace_repo_slug_issues_issue_id_attachments_path:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/issues/{{ record.issue_id }}/attachments/{{ record.path }}
    required fields: issue_id, path
    risk: Delete an attachment for an issue
  op_post_repositories_workspace_repo_slug_issues_issue_id_changes:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/issues/{{ record.issue_id }}/changes
    required fields: issue_id
    risk: Modify the state of an issue
  op_post_repositories_workspace_repo_slug_issues_issue_id_comments:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/issues/{{ record.issue_id }}/comments
    required fields: issue_id
    risk: Create a comment on an issue
  op_put_repositories_workspace_repo_slug_issues_issue_id_comments_comment_id:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/issues/{{ record.issue_id }}/comments/{{ record.comment_id }}
    required fields: issue_id, comment_id
    risk: Update a comment on an issue
  op_delete_repositories_workspace_repo_slug_issues_issue_id_comments_comment_id:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/issues/{{ record.issue_id }}/comments/{{ record.comment_id }}
    required fields: issue_id, comment_id
    risk: Delete a comment on an issue
  op_put_repositories_workspace_repo_slug_issues_issue_id_vote:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/issues/{{ record.issue_id }}/vote
    required fields: issue_id
    risk: Vote for an issue
  op_delete_repositories_workspace_repo_slug_issues_issue_id_vote:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/issues/{{ record.issue_id }}/vote
    required fields: issue_id
    risk: Remove vote for an issue
  op_put_repositories_workspace_repo_slug_issues_issue_id_watch:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/issues/{{ record.issue_id }}/watch
    required fields: issue_id
    risk: Watch an issue
  op_delete_repositories_workspace_repo_slug_issues_issue_id_watch:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/issues/{{ record.issue_id }}/watch
    required fields: issue_id
    risk: Stop watching an issue
  op_put_repositories_workspace_repo_slug_override_settings:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/override-settings
    risk: Set the inheritance state for repository settings
  op_put_repositories_workspace_repo_slug_permissions_config_groups_group_slug:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/permissions-config/groups/{{ record.group_slug }}
    required fields: group_slug
    risk: Update an explicit group permission for a repository
  op_delete_repositories_workspace_repo_slug_permissions_config_groups_group_slug:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/permissions-config/groups/{{ record.group_slug }}
    required fields: group_slug
    risk: Delete an explicit group permission for a repository
  op_put_repositories_workspace_repo_slug_permissions_config_users_selected_user_id:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/permissions-config/users/{{ record.selected_user_id }}
    required fields: selected_user_id
    risk: Update an explicit user permission for a repository
  op_delete_repositories_workspace_repo_slug_permissions_config_users_selected_user_id:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/permissions-config/users/{{ record.selected_user_id }}
    required fields: selected_user_id
    risk: Delete an explicit user permission for a repository
  op_delete_repositories_workspace_repo_slug_pipelines_config_caches:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines-config/caches
    risk: Delete caches
  op_delete_repositories_workspace_repo_slug_pipelines_config_caches_cache_uuid:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines-config/caches/{{ record.cache_uuid }}
    required fields: cache_uuid
    risk: Delete a cache
  op_post_repositories_workspace_repo_slug_pipelines_config_runners:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines-config/runners
    risk: Create repository runner
  op_put_repositories_workspace_repo_slug_pipelines_config_runners_runner_uuid:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines-config/runners/{{ record.runner_uuid }}
    required fields: runner_uuid
    risk: Update repository runner
  op_delete_repositories_workspace_repo_slug_pipelines_config_runners_runner_uuid:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines-config/runners/{{ record.runner_uuid }}
    required fields: runner_uuid
    risk: Delete repository runner
  op_put_repositories_workspace_repo_slug_pipelines_config:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines_config
    risk: Update configuration
  op_put_repositories_workspace_repo_slug_pipelines_config_build_number:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines_config/build_number
    risk: Update the next build number
  op_post_repositories_workspace_repo_slug_pipelines_config_schedules:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines_config/schedules
    risk: Create a schedule
  op_put_repositories_workspace_repo_slug_pipelines_config_schedules_schedule_uuid:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines_config/schedules/{{ record.schedule_uuid }}
    required fields: schedule_uuid
    risk: Update a schedule
  op_delete_repositories_workspace_repo_slug_pipelines_config_schedules_schedule_uuid:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines_config/schedules/{{ record.schedule_uuid }}
    required fields: schedule_uuid
    risk: Delete a schedule
  op_put_repositories_workspace_repo_slug_pipelines_config_ssh_key_pair:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines_config/ssh/key_pair
    risk: Update SSH key pair
  op_delete_repositories_workspace_repo_slug_pipelines_config_ssh_key_pair:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines_config/ssh/key_pair
    risk: Delete SSH key pair
  op_post_repositories_workspace_repo_slug_pipelines_config_ssh_known_hosts:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines_config/ssh/known_hosts
    risk: Create a known host
  op_put_repositories_workspace_repo_slug_pipelines_config_ssh_known_hosts_known_host_uuid:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines_config/ssh/known_hosts/{{ record.known_host_uuid }}
    required fields: known_host_uuid
    risk: Update a known host
  op_delete_repositories_workspace_repo_slug_pipelines_config_ssh_known_hosts_known_host_uuid:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines_config/ssh/known_hosts/{{ record.known_host_uuid }}
    required fields: known_host_uuid
    risk: Delete a known host
  op_post_repositories_workspace_repo_slug_pipelines_config_variables:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines_config/variables
    risk: Create a variable for a repository
  op_put_repositories_workspace_repo_slug_pipelines_config_variables_variable_uuid:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines_config/variables/{{ record.variable_uuid }}
    required fields: variable_uuid
    risk: Update a variable for a repository
  op_delete_repositories_workspace_repo_slug_pipelines_config_variables_variable_uuid:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pipelines_config/variables/{{ record.variable_uuid }}
    required fields: variable_uuid
    risk: Delete a variable for a repository
  op_put_repositories_workspace_repo_slug_properties_app_key_property_name:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/properties/{{ record.app_key }}/{{ record.property_name }}
    required fields: app_key, property_name
    risk: Update a repository application property
  op_delete_repositories_workspace_repo_slug_properties_app_key_property_name:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/properties/{{ record.app_key }}/{{ record.property_name }}
    required fields: app_key, property_name
    risk: Delete a repository application property
  op_put_repositories_workspace_repo_slug_pullrequests_pull_request_id:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pull_request_id }}
    required fields: pull_request_id
    risk: Update a pull request
  op_post_repositories_workspace_repo_slug_pullrequests_pull_request_id_approve:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pull_request_id }}/approve
    required fields: pull_request_id
    risk: Approve a pull request
  op_delete_repositories_workspace_repo_slug_pullrequests_pull_request_id_approve:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pull_request_id }}/approve
    required fields: pull_request_id
    risk: Unapprove a pull request
  op_post_repositories_workspace_repo_slug_pullrequests_pull_request_id_comments:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pull_request_id }}/comments
    required fields: pull_request_id
    risk: Create a comment on a pull request
  op_put_repositories_workspace_repo_slug_pullrequests_pull_request_id_comments_comment_id:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pull_request_id }}/comments/{{ record.comment_id }}
    required fields: pull_request_id, comment_id
    risk: Update a comment on a pull request
  op_delete_repositories_workspace_repo_slug_pullrequests_pull_request_id_comments_comment_id:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pull_request_id }}/comments/{{ record.comment_id }}
    required fields: pull_request_id, comment_id
    risk: Delete a comment on a pull request
  op_post_repositories_workspace_repo_slug_pullrequests_pull_request_id_comments_comment_id_resolve:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pull_request_id }}/comments/{{ record.comment_id }}/resolve
    required fields: pull_request_id, comment_id
    risk: Resolve a comment thread
  op_delete_repositories_workspace_repo_slug_pullrequests_pull_request_id_comments_comment_id_resolve:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pull_request_id }}/comments/{{ record.comment_id }}/resolve
    required fields: pull_request_id, comment_id
    risk: Reopen a comment thread
  op_post_repositories_workspace_repo_slug_pullrequests_pull_request_id_request_changes:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pull_request_id }}/request-changes
    required fields: pull_request_id
    risk: Request changes for a pull request
  op_delete_repositories_workspace_repo_slug_pullrequests_pull_request_id_request_changes:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pull_request_id }}/request-changes
    required fields: pull_request_id
    risk: Remove change request for a pull request
  op_post_repositories_workspace_repo_slug_pullrequests_pull_request_id_tasks:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pull_request_id }}/tasks
    required fields: pull_request_id
    risk: Create a task on a pull request
  op_put_repositories_workspace_repo_slug_pullrequests_pull_request_id_tasks_task_id:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pull_request_id }}/tasks/{{ record.task_id }}
    required fields: pull_request_id, task_id
    risk: Update a task on a pull request
  op_delete_repositories_workspace_repo_slug_pullrequests_pull_request_id_tasks_task_id:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pull_request_id }}/tasks/{{ record.task_id }}
    required fields: pull_request_id, task_id
    risk: Delete a task on a pull request
  op_put_repositories_workspace_repo_slug_pullrequests_pullrequest_id_properties_app_key_property_name:
    endpoint: PUT /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pullrequest_id }}/properties/{{ record.app_key }}/{{ record.property_name }}
    required fields: pullrequest_id, app_key, property_name
    risk: Update a pull request application property
  op_delete_repositories_workspace_repo_slug_pullrequests_pullrequest_id_properties_app_key_property_name:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/pullrequests/{{ record.pullrequest_id }}/properties/{{ record.app_key }}/{{ record.property_name }}
    required fields: pullrequest_id, app_key, property_name
    risk: Delete a pull request application property
  op_post_repositories_workspace_repo_slug_refs_branches:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/refs/branches
    risk: Create a branch
  op_delete_repositories_workspace_repo_slug_refs_branches_name:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/refs/branches/{{ record.name }}
    required fields: name
    risk: Delete a branch
  op_post_repositories_workspace_repo_slug_refs_tags:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/refs/tags
    risk: Create a tag
  op_delete_repositories_workspace_repo_slug_refs_tags_name:
    endpoint: DELETE /repositories/{{ config.workspace }}/{{ config.repo_slug }}/refs/tags/{{ record.name }}
    required fields: name
    risk: Delete a tag
  op_post_repositories_workspace_repo_slug_src:
    endpoint: POST /repositories/{{ config.workspace }}/{{ config.repo_slug }}/src
    risk: Create a commit by uploading a file
  op_post_snippets:
    endpoint: POST /snippets
    risk: Create a snippet
  op_put_snippets_workspace_encoded_id:
    endpoint: PUT /snippets/{{ config.workspace }}/{{ record.encoded_id }}
    required fields: encoded_id
    risk: Update a snippet
  op_delete_snippets_workspace_encoded_id:
    endpoint: DELETE /snippets/{{ config.workspace }}/{{ record.encoded_id }}
    required fields: encoded_id
    risk: Delete a snippet
  op_post_snippets_workspace_encoded_id_comments:
    endpoint: POST /snippets/{{ config.workspace }}/{{ record.encoded_id }}/comments
    required fields: encoded_id
    risk: Create a comment on a snippet
  op_put_snippets_workspace_encoded_id_comments_comment_id:
    endpoint: PUT /snippets/{{ config.workspace }}/{{ record.encoded_id }}/comments/{{ record.comment_id }}
    required fields: encoded_id, comment_id
    risk: Update a comment on a snippet
  op_delete_snippets_workspace_encoded_id_comments_comment_id:
    endpoint: DELETE /snippets/{{ config.workspace }}/{{ record.encoded_id }}/comments/{{ record.comment_id }}
    required fields: encoded_id, comment_id
    risk: Delete a comment on a snippet
  op_put_snippets_workspace_encoded_id_watch:
    endpoint: PUT /snippets/{{ config.workspace }}/{{ record.encoded_id }}/watch
    required fields: encoded_id
    risk: Watch a snippet
  op_delete_snippets_workspace_encoded_id_watch:
    endpoint: DELETE /snippets/{{ config.workspace }}/{{ record.encoded_id }}/watch
    required fields: encoded_id
    risk: Stop watching a snippet
  op_put_snippets_workspace_encoded_id_node_id:
    endpoint: PUT /snippets/{{ config.workspace }}/{{ record.encoded_id }}/{{ record.node_id }}
    required fields: encoded_id, node_id
    risk: Update a previous revision of a snippet
  op_delete_snippets_workspace_encoded_id_node_id:
    endpoint: DELETE /snippets/{{ config.workspace }}/{{ record.encoded_id }}/{{ record.node_id }}
    required fields: encoded_id, node_id
    risk: Delete a previous revision of a snippet
  op_post_teams_username_pipelines_config_variables:
    endpoint: POST /teams/{{ record.username }}/pipelines_config/variables
    required fields: username
    risk: Create a variable for a user
  op_put_teams_username_pipelines_config_variables_variable_uuid:
    endpoint: PUT /teams/{{ record.username }}/pipelines_config/variables/{{ record.variable_uuid }}
    required fields: username, variable_uuid
    risk: Update a variable for a team
  op_delete_teams_username_pipelines_config_variables_variable_uuid:
    endpoint: DELETE /teams/{{ record.username }}/pipelines_config/variables/{{ record.variable_uuid }}
    required fields: username, variable_uuid
    risk: Delete a variable for a team
  op_post_users_selected_user_gpg_keys:
    endpoint: POST /users/{{ record.selected_user }}/gpg-keys
    required fields: selected_user
    risk: Add a new GPG key
  op_delete_users_selected_user_gpg_keys_fingerprint:
    endpoint: DELETE /users/{{ record.selected_user }}/gpg-keys/{{ record.fingerprint }}
    required fields: selected_user, fingerprint
    risk: Delete a GPG key
  op_post_users_selected_user_pipelines_config_variables:
    endpoint: POST /users/{{ record.selected_user }}/pipelines_config/variables
    required fields: selected_user
    risk: Create a variable for a user
  op_put_users_selected_user_pipelines_config_variables_variable_uuid:
    endpoint: PUT /users/{{ record.selected_user }}/pipelines_config/variables/{{ record.variable_uuid }}
    required fields: selected_user, variable_uuid
    risk: Update a variable for a user
  op_delete_users_selected_user_pipelines_config_variables_variable_uuid:
    endpoint: DELETE /users/{{ record.selected_user }}/pipelines_config/variables/{{ record.variable_uuid }}
    required fields: selected_user, variable_uuid
    risk: Delete a variable for a user
  op_put_users_selected_user_properties_app_key_property_name:
    endpoint: PUT /users/{{ record.selected_user }}/properties/{{ record.app_key }}/{{ record.property_name }}
    required fields: selected_user, app_key, property_name
    risk: Update a user application property
  op_delete_users_selected_user_properties_app_key_property_name:
    endpoint: DELETE /users/{{ record.selected_user }}/properties/{{ record.app_key }}/{{ record.property_name }}
    required fields: selected_user, app_key, property_name
    risk: Delete a user application property
  op_post_users_selected_user_ssh_keys:
    endpoint: POST /users/{{ record.selected_user }}/ssh-keys
    required fields: selected_user
    risk: Add a new SSH key
  op_put_users_selected_user_ssh_keys_key_id:
    endpoint: PUT /users/{{ record.selected_user }}/ssh-keys/{{ record.key_id }}
    required fields: selected_user, key_id
    risk: Update a SSH key
  op_delete_users_selected_user_ssh_keys_key_id:
    endpoint: DELETE /users/{{ record.selected_user }}/ssh-keys/{{ record.key_id }}
    required fields: selected_user, key_id
    risk: Delete a SSH key
  op_post_workspaces_workspace_hooks:
    endpoint: POST /workspaces/{{ config.workspace }}/hooks
    risk: Create a webhook for a workspace
  op_put_workspaces_workspace_hooks_uid:
    endpoint: PUT /workspaces/{{ config.workspace }}/hooks/{{ record.uid }}
    required fields: uid
    risk: Update a webhook for a workspace
  op_delete_workspaces_workspace_hooks_uid:
    endpoint: DELETE /workspaces/{{ config.workspace }}/hooks/{{ record.uid }}
    required fields: uid
    risk: Delete a webhook for a workspace
  op_post_workspaces_workspace_pipelines_config_runners:
    endpoint: POST /workspaces/{{ config.workspace }}/pipelines-config/runners
    risk: Create workspace runner
  op_put_workspaces_workspace_pipelines_config_runners_runner_uuid:
    endpoint: PUT /workspaces/{{ config.workspace }}/pipelines-config/runners/{{ record.runner_uuid }}
    required fields: runner_uuid
    risk: Update workspace runner
  op_delete_workspaces_workspace_pipelines_config_runners_runner_uuid:
    endpoint: DELETE /workspaces/{{ config.workspace }}/pipelines-config/runners/{{ record.runner_uuid }}
    required fields: runner_uuid
    risk: Delete workspace runner
  op_post_workspaces_workspace_pipelines_config_variables:
    endpoint: POST /workspaces/{{ config.workspace }}/pipelines-config/variables
    risk: Create a variable for a workspace
  op_put_workspaces_workspace_pipelines_config_variables_variable_uuid:
    endpoint: PUT /workspaces/{{ config.workspace }}/pipelines-config/variables/{{ record.variable_uuid }}
    required fields: variable_uuid
    risk: Update variable for a workspace
  op_delete_workspaces_workspace_pipelines_config_variables_variable_uuid:
    endpoint: DELETE /workspaces/{{ config.workspace }}/pipelines-config/variables/{{ record.variable_uuid }}
    required fields: variable_uuid
    risk: Delete a variable for a workspace
  op_post_workspaces_workspace_projects:
    endpoint: POST /workspaces/{{ config.workspace }}/projects
    risk: Create a project in a workspace
  op_put_workspaces_workspace_projects_project_key:
    endpoint: PUT /workspaces/{{ config.workspace }}/projects/{{ record.project_key }}
    required fields: project_key
    risk: Update a project for a workspace
  op_delete_workspaces_workspace_projects_project_key:
    endpoint: DELETE /workspaces/{{ config.workspace }}/projects/{{ record.project_key }}
    required fields: project_key
    risk: Delete a project for a workspace
  op_put_workspaces_workspace_projects_project_key_branching_model_settings:
    endpoint: PUT /workspaces/{{ config.workspace }}/projects/{{ record.project_key }}/branching-model/settings
    required fields: project_key
    risk: Update the branching model config for a project
  op_put_workspaces_workspace_projects_project_key_default_reviewers_selected_user:
    endpoint: PUT /workspaces/{{ config.workspace }}/projects/{{ record.project_key }}/default-reviewers/{{ record.selected_user }}
    required fields: project_key, selected_user
    risk: Add the specific user as a default reviewer for the project
  op_delete_workspaces_workspace_projects_project_key_default_reviewers_selected_user:
    endpoint: DELETE /workspaces/{{ config.workspace }}/projects/{{ record.project_key }}/default-reviewers/{{ record.selected_user }}
    required fields: project_key, selected_user
    risk: Remove the specific user from the project's default reviewers
  op_post_workspaces_workspace_projects_project_key_deploy_keys:
    endpoint: POST /workspaces/{{ config.workspace }}/projects/{{ record.project_key }}/deploy-keys
    required fields: project_key
    risk: Create a project deploy key
  op_delete_workspaces_workspace_projects_project_key_deploy_keys_key_id:
    endpoint: DELETE /workspaces/{{ config.workspace }}/projects/{{ record.project_key }}/deploy-keys/{{ record.key_id }}
    required fields: project_key, key_id
    risk: Delete a deploy key from a project
  op_put_workspaces_workspace_projects_project_key_permissions_config_groups_group_slug:
    endpoint: PUT /workspaces/{{ config.workspace }}/projects/{{ record.project_key }}/permissions-config/groups/{{ record.group_slug }}
    required fields: project_key, group_slug
    risk: Update an explicit group permission for a project
  op_delete_workspaces_workspace_projects_project_key_permissions_config_groups_group_slug:
    endpoint: DELETE /workspaces/{{ config.workspace }}/projects/{{ record.project_key }}/permissions-config/groups/{{ record.group_slug }}
    required fields: project_key, group_slug
    risk: Delete an explicit group permission for a project
  op_put_workspaces_workspace_projects_project_key_permissions_config_users_selected_user_id:
    endpoint: PUT /workspaces/{{ config.workspace }}/projects/{{ record.project_key }}/permissions-config/users/{{ record.selected_user_id }}
    required fields: project_key, selected_user_id
    risk: Update an explicit user permission for a project
  op_delete_workspaces_workspace_projects_project_key_permissions_config_users_selected_user_id:
    endpoint: DELETE /workspaces/{{ config.workspace }}/projects/{{ record.project_key }}/permissions-config/users/{{ record.selected_user_id }}
    required fields: project_key, selected_user_id
    risk: Delete an explicit user permission for a project

SECURITY
  read risk: Bitbucket Cloud REST API reads scoped to the configured workspace/repository; binary payloads and local git workflows are not executed.
  write risk: Selected Bitbucket mutations are explicit reverse ETL actions only; destructive/admin/sensitive operations are blocked or require typed confirmation metadata.
  approval: Reverse ETL writes require plan, preview, approval, execute; destructive/admin operations require typed confirmation when exposed.
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Work with Bitbucket Cloud repositories from the command line.
  Usage: pm bitbucket <command> <subcommand> [flags]
  Source CLI: bb/Bitbucket Cloud REST (https://developer.atlassian.com/cloud/bitbucket/rest/)
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --credential (string): Use a saved Bitbucket connector credential.: maps_to=connection
    --connection (string): Alias for --credential.: maps_to=connection
    --workspace (string): Bitbucket workspace slug.: maps_to=config.workspace
    --repo (string): Bitbucket repository slug.: maps_to=config.repo_slug
    --limit (integer): Maximum records to emit for stream commands.
    --max-bytes (integer): Maximum bytes for direct-read JSON responses.
  Repository Commands
    repo list - List repositories in a workspace [intent=etl availability=implemented stream=repositories]
    repo view - View repository details [intent=direct_read availability=implemented]; flags: --repo
    repo create - Create a repository [intent=reverse_etl availability=implemented write=create_repository]; approval: reverse ETL writes require plan, preview, approval, execute. Typed confirmation may be required by policy.; risk: Creates a repository in the configured workspace.; flags: --name, --description, --is-private
    repo delete - Delete a repository [intent=direct_write availability=unsafe_or_disallowed]; notes: Repository deletion is destructive admin behavior and is not exposed as a connector write.
    repo clone - Clone a repository locally [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Depends on local git and filesystem state; no local git executor is enabled.
    branch list - List repository branches [intent=etl availability=implemented stream=branches]
    commit list - List repository commits [intent=etl availability=implemented stream=commits]
    tag list - List repository tags [intent=etl availability=implemented stream=tags]
    download list - List repository downloads [intent=etl availability=implemented stream=downloads]
    download get - Download a repository file asset [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Binary download to local filesystem requires explicit max-byte, destination, overwrite, and archive policies before execution.
  Pull Request Commands
    pull-request list - List pull requests [intent=etl availability=implemented stream=pull_requests]; flags: --state
    pull-request view - View pull request details [intent=direct_read availability=implemented]; flags: --pull-request-id
    pull-request create - Create a pull request [intent=reverse_etl availability=implemented write=create_pull_request]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a visible pull request in the configured repository.; flags: --source-branch, --destination-branch, --title
    pull-request merge - Merge a pull request [intent=reverse_etl availability=implemented write=merge_pull_request]; approval: reverse ETL writes require plan, preview, approval, execute. Typed confirmation may be required by policy.; risk: Merges code into the destination branch.; flags: --pull-request-id, --message, --async
    pull-request decline - Decline a pull request [intent=reverse_etl availability=implemented write=decline_pull_request]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Declines an open pull request.; flags: --pull-request-id, --message
  Issue Tracker Commands
    issue list - List issues [intent=etl availability=implemented stream=issues]; flags: --state
    issue view - View issue details [intent=direct_read availability=implemented]; flags: --issue-id
    issue create - Create an issue [intent=reverse_etl availability=implemented write=create_issue]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Creates a visible issue in the configured repository.; flags: --title, --kind, --priority
    issue edit - Edit an issue [intent=reverse_etl availability=implemented write=update_issue]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Mutates an existing issue.; flags: --issue-id, --title, --state
    issue comment - Comment on an issue [intent=reverse_etl availability=partial]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Adds a visible issue comment.; notes: Bitbucket comment bodies require nested content objects; use reverse ETL writes once nested body mapping is modeled.
  Pipelines And Deployments Commands
    pipeline list - List pipelines [intent=etl availability=implemented stream=pipelines]
    pipeline view - View pipeline details [intent=direct_read availability=implemented]; flags: --pipeline-uuid
    pipeline run - Run a pipeline [intent=reverse_etl availability=implemented write=run_pipeline]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Starts a Bitbucket pipeline execution.
    pipeline stop - Stop a pipeline [intent=reverse_etl availability=implemented write=stop_pipeline]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Stops an in-flight Bitbucket pipeline.; flags: --pipeline-uuid
    deployment list - List deployments [intent=etl availability=implemented stream=deployments]
  Workspace And Administration Commands
    workspace list - List accessible workspaces [intent=direct_read availability=implemented]
    project list - List workspace projects [intent=direct_read availability=implemented]
    webhook list - List repository webhooks [intent=etl availability=implemented stream=webhooks]
    webhook create - Create a repository webhook [intent=reverse_etl availability=implemented write=create_webhook]; approval: reverse ETL writes require plan, preview, approval, execute. URL policy review is required.; risk: Creates an outbound webhook that can send repository events to an external URL.; flags: --url, --event
    webhook delete - Delete a repository webhook [intent=reverse_etl availability=implemented write=delete_webhook]; approval: reverse ETL writes require plan, preview, approval, execute. Typed confirmation may be required by policy.; risk: Deletes an existing webhook and may interrupt downstream automation.; flags: --uid
    branch-restriction list - List branch restrictions [intent=etl availability=implemented stream=branch_restrictions]
    branch-restriction create - Create a branch restriction [intent=reverse_etl availability=implemented write=create_branch_restriction]; approval: reverse ETL writes require plan, preview, approval, execute. Typed confirmation may be required for admin policy changes.; risk: Changes repository branch protection behavior.; flags: --kind, --pattern
    branch-restriction delete - Delete a branch restriction [intent=reverse_etl availability=partial]; approval: reverse ETL writes require plan, preview, approval, execute. Typed confirmation is required for admin policy changes.; risk: Removes branch protection behavior.; notes: Blocked until branch restriction delete confirmation UX is modeled.
    snippet list - List snippets [intent=direct_read availability=implemented]
    snippet create - Create a snippet [intent=reverse_etl availability=implemented write=create_snippet]; approval: reverse ETL writes require plan, preview, approval, execute. Content redaction review is required.; risk: Creates a Bitbucket snippet and may publish code or text content.; flags: --title, --content, --is-private
  Local Workflow Commands
    auth status - Show credential status [intent=auth availability=unsupported_local unsupported local workflow]; notes: Use `pm credentials inspect <name> --redacted`; this metadata does not read secrets.
    config view - Show Bitbucket command configuration [intent=config availability=unsupported_local unsupported local workflow]; notes: Connector-local CLI configuration is not implemented.
    api - Call an arbitrary Bitbucket API endpoint [intent=raw_api availability=unsafe_or_disallowed]; notes: Generic raw API calls are forbidden. Add reviewed direct-read or reverse-ETL operations instead.
  Typed Operation Commands
    operation put addon - Update an installed app [intent=reverse_etl availability=implemented write=op_put_addon]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update an installed app
    operation delete addon - Delete an app [intent=reverse_etl availability=implemented write=op_delete_addon]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete an app
    operation get addon addon-key client-key - Get the client key of a Connect addon [intent=direct_read availability=implemented]; flags: --addon-key
    operation get hook-events - Get a webhook resource [intent=direct_read availability=implemented]
    operation get hook-events subject-type - List subscribable webhook types [intent=direct_read availability=implemented]; flags: --subject-type
    operation get repositories - List public repositories [intent=direct_read availability=implemented]; flags: --after, --role, --q, --sort
    operation get repositories workspace - List repositories in a workspace [intent=direct_read availability=implemented]; flags: --workspace, --role, --q, --sort
    operation put repositories workspace repo-slug - Update a repository [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a repository
    operation delete repositories workspace repo-slug - Delete a repository [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a repository; flags: --redirect-to
    operation get repositories workspace repo-slug branch-restrictions - List branch restrictions [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --kind, --pattern
    operation get repositories workspace repo-slug branch-restrictions id - Get a branch restriction rule [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --id
    operation put repositories workspace repo-slug branch-restrictions id - Update a branch restriction rule [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_branch_restrictions_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a branch restriction rule; flags: --id
    operation delete repositories workspace repo-slug branch-restrictions id - Delete a branch restriction rule [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_branch_restrictions_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a branch restriction rule; flags: --id
    operation get repositories workspace repo-slug branching-model - Get the branching model for a repository [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation get repositories workspace repo-slug branching-model settings - Get the branching model config for a repository [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation put repositories workspace repo-slug branching-model settings - Update the branching model config for a repository [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_branching_model_settings]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update the branching model config for a repository
    operation get repositories workspace repo-slug commit commit - Get a commit [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --commit
    operation post repositories workspace repo-slug commit commit approve - Approve a commit [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_commit_commit_approve]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Approve a commit; flags: --commit
    operation delete repositories workspace repo-slug commit commit approve - Unapprove a commit [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_commit_commit_approve]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Unapprove a commit; flags: --commit
    operation get repositories workspace repo-slug commit commit comments - List a commit's comments [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --commit, --q, --sort
    operation post repositories workspace repo-slug commit commit comments - Create comment for a commit [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_commit_commit_comments]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create comment for a commit; flags: --commit
    operation get repositories workspace repo-slug commit commit comments comment-id - Get a commit comment [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --commit, --comment-id
    operation put repositories workspace repo-slug commit commit comments comment-id - Update a commit comment [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_commit_commit_comments_comment_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a commit comment; flags: --commit, --comment-id
    operation delete repositories workspace repo-slug commit commit comments comment-id - Delete a commit comment [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_commit_commit_comments_comment_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a commit comment; flags: --commit, --comment-id
    operation get repositories workspace repo-slug commit commit properties app-key property-name - Get a commit application property [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --commit, --app-key, --property-name
    operation put repositories workspace repo-slug commit commit properties app-key property-name - Update a commit application property [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_commit_commit_properties_app_key_property_name]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a commit application property; flags: --commit, --app-key, --property-name
    operation delete repositories workspace repo-slug commit commit properties app-key property-name - Delete a commit application property [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_commit_commit_properties_app_key_property_name]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a commit application property; flags: --commit, --app-key, --property-name
    operation get repositories workspace repo-slug commit commit pullrequests - List pull requests that contain a commit [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --commit, --page, --pagelen
    operation get repositories workspace repo-slug commit commit reports - List reports [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --commit
    operation get repositories workspace repo-slug commit commit reports reportid - Get a report [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --commit, --reportId
    operation put repositories workspace repo-slug commit commit reports reportid - Create or update a report [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_commit_commit_reports_reportid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create or update a report; flags: --commit, --reportId
    operation delete repositories workspace repo-slug commit commit reports reportid - Delete a report [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_commit_commit_reports_reportid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a report; flags: --commit, --reportId
    operation get repositories workspace repo-slug commit commit reports reportid annotations - List annotations [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --commit, --reportId
    operation post repositories workspace repo-slug commit commit reports reportid annotations - Bulk create or update annotations [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_commit_commit_reports_reportid_annotations]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Bulk create or update annotations; flags: --commit, --reportId
    operation get repositories workspace repo-slug commit commit reports reportid annotations annotationid - Get an annotation [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --commit, --reportId, --annotationId
    operation put repositories workspace repo-slug commit commit reports reportid annotations annotationid - Create or update an annotation [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_commit_commit_reports_reportid_annotations_annotationid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create or update an annotation; flags: --commit, --reportId, --annotationId
    operation delete repositories workspace repo-slug commit commit reports reportid annotations annotationid - Delete an annotation [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_commit_commit_reports_reportid_annotations_annotationid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete an annotation; flags: --commit, --reportId, --annotationId
    operation get repositories workspace repo-slug commit commit statuses - List commit statuses for a commit [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --commit, --refname, --q, --sort
    operation post repositories workspace repo-slug commit commit statuses build - Create a build status for a commit [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_commit_commit_statuses_build]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a build status for a commit; flags: --commit
    operation get repositories workspace repo-slug commit commit statuses build key - Get a build status for a commit [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --commit, --key
    operation put repositories workspace repo-slug commit commit statuses build key - Update a build status for a commit [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_commit_commit_statuses_build_key]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a build status for a commit; flags: --commit, --key
    operation get repositories workspace repo-slug commits - List commits [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation post repositories workspace repo-slug commits - List commits with include/exclude [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_commits]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: List commits with include/exclude
    operation get repositories workspace repo-slug commits revision - List commits for revision [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --revision
    operation post repositories workspace repo-slug commits revision - List commits for revision using include/exclude [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_commits_revision]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: List commits for revision using include/exclude; flags: --revision
    operation get repositories workspace repo-slug components - List components [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation get repositories workspace repo-slug components component-id - Get a component for issues [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --component-id
    operation get repositories workspace repo-slug default-reviewers - List default reviewers [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation get repositories workspace repo-slug default-reviewers target-username - Get a default reviewer [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --target-username
    operation put repositories workspace repo-slug default-reviewers target-username - Add a user to the default reviewers [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_default_reviewers_target_username]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Add a user to the default reviewers; flags: --target-username
    operation delete repositories workspace repo-slug default-reviewers target-username - Remove a user from the default reviewers [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_default_reviewers_target_username]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Remove a user from the default reviewers; flags: --target-username
    operation get repositories workspace repo-slug deploy-keys - List repository deploy keys [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation post repositories workspace repo-slug deploy-keys - Add a repository deploy key [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_deploy_keys]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Add a repository deploy key
    operation get repositories workspace repo-slug deploy-keys key-id - Get a repository deploy key [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --key-id
    operation put repositories workspace repo-slug deploy-keys key-id - Update a repository deploy key [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_deploy_keys_key_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a repository deploy key; flags: --key-id
    operation delete repositories workspace repo-slug deploy-keys key-id - Delete a repository deploy key [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_deploy_keys_key_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a repository deploy key; flags: --key-id
    operation get repositories workspace repo-slug deployments - List deployments [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation get repositories workspace repo-slug deployments deployment-uuid - Get a deployment [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --deployment-uuid
    operation get repositories workspace repo-slug deployments-config environments environment-uuid variables - List variables for an environment [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --environment-uuid
    operation post repositories workspace repo-slug deployments-config environments environment-uuid variables - Create a variable for an environment [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_deployments_config_environments_environment_uuid_variables]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a variable for an environment; flags: --environment-uuid
    operation put repositories workspace repo-slug deployments-config environments environment-uuid variables variable-uuid - Update a variable for an environment [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_deployments_config_environments_environment_uuid_variables_variable_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a variable for an environment; flags: --environment-uuid, --variable-uuid
    operation delete repositories workspace repo-slug deployments-config environments environment-uuid variables variable-uuid - Delete a variable for an environment [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_deployments_config_environments_environment_uuid_variables_variable_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a variable for an environment; flags: --environment-uuid, --variable-uuid
    operation get repositories workspace repo-slug diff spec - Compare two commits [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --spec, --context, --path, --ignore-whitespace, --binary, --renames, --merge, --topic
    operation get repositories workspace repo-slug diffstat spec - Compare two commit diff stats [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --spec
    operation get repositories workspace repo-slug downloads - List download artifacts [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation post repositories workspace repo-slug downloads - Upload a download artifact [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_downloads]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Upload a download artifact
    operation get repositories workspace repo-slug downloads filename - Get a download artifact link [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --filename
    operation delete repositories workspace repo-slug downloads filename - Delete a download artifact [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_downloads_filename]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a download artifact; flags: --filename
    operation get repositories workspace repo-slug effective-branching-model - Get the effective, or currently applied, branching model for a repository [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation get repositories workspace repo-slug effective-default-reviewers - List effective default reviewers [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation get repositories workspace repo-slug environments - List environments [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation post repositories workspace repo-slug environments - Create an environment [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_environments]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create an environment
    operation get repositories workspace repo-slug environments environment-uuid - Get an environment [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --environment-uuid
    operation delete repositories workspace repo-slug environments environment-uuid - Delete an environment [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_environments_environment_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete an environment; flags: --environment-uuid
    operation post repositories workspace repo-slug environments environment-uuid changes - Update an environment [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_environments_environment_uuid_changes]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update an environment; flags: --environment-uuid
    operation get repositories workspace repo-slug file-conflicts spec - Get file conflicts for a commit spec [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --spec
    operation get repositories workspace repo-slug filehistory commit path - List commits that modified a file [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --commit, --path, --renames, --q, --sort
    operation get repositories workspace repo-slug forks - List repository forks [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --role, --q, --sort
    operation post repositories workspace repo-slug forks - Fork a repository [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_forks]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Fork a repository
    operation get repositories workspace repo-slug hooks - List webhooks for a repository [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation get repositories workspace repo-slug hooks uid - Get a webhook for a repository [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --uid
    operation put repositories workspace repo-slug hooks uid - Update a webhook for a repository [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_hooks_uid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a webhook for a repository; flags: --uid
    operation get repositories workspace repo-slug issues - List issues [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation post repositories workspace repo-slug issues export - Export issues [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_issues_export]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Export issues
    operation get repositories workspace repo-slug issues export repo-name--issues--task-id-.zip - Check issue export status [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --repo-name, --task-id
    operation get repositories workspace repo-slug issues import - Check issue import status [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation post repositories workspace repo-slug issues import - Import issues [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_issues_import]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Import issues
    operation delete repositories workspace repo-slug issues issue-id - Delete an issue [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_issues_issue_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete an issue; flags: --issue-id
    operation get repositories workspace repo-slug issues issue-id attachments - List attachments for an issue [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --issue-id
    operation post repositories workspace repo-slug issues issue-id attachments - Upload an attachment to an issue [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_issues_issue_id_attachments]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Upload an attachment to an issue; flags: --issue-id
    operation get repositories workspace repo-slug issues issue-id attachments path - Get attachment for an issue [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --issue-id, --path
    operation delete repositories workspace repo-slug issues issue-id attachments path - Delete an attachment for an issue [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_issues_issue_id_attachments_path]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete an attachment for an issue; flags: --issue-id, --path
    operation get repositories workspace repo-slug issues issue-id changes - List changes on an issue [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --issue-id, --q, --sort
    operation post repositories workspace repo-slug issues issue-id changes - Modify the state of an issue [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_issues_issue_id_changes]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Modify the state of an issue; flags: --issue-id
    operation get repositories workspace repo-slug issues issue-id changes change-id - Get issue change object [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --issue-id, --change-id
    operation get repositories workspace repo-slug issues issue-id comments - List comments on an issue [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --issue-id, --q
    operation post repositories workspace repo-slug issues issue-id comments - Create a comment on an issue [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_issues_issue_id_comments]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a comment on an issue; flags: --issue-id
    operation get repositories workspace repo-slug issues issue-id comments comment-id - Get a comment on an issue [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --issue-id, --comment-id
    operation put repositories workspace repo-slug issues issue-id comments comment-id - Update a comment on an issue [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_issues_issue_id_comments_comment_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a comment on an issue; flags: --issue-id, --comment-id
    operation delete repositories workspace repo-slug issues issue-id comments comment-id - Delete a comment on an issue [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_issues_issue_id_comments_comment_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a comment on an issue; flags: --issue-id, --comment-id
    operation get repositories workspace repo-slug issues issue-id vote - Check if current user voted for an issue [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --issue-id
    operation put repositories workspace repo-slug issues issue-id vote - Vote for an issue [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_issues_issue_id_vote]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Vote for an issue; flags: --issue-id
    operation delete repositories workspace repo-slug issues issue-id vote - Remove vote for an issue [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_issues_issue_id_vote]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Remove vote for an issue; flags: --issue-id
    operation get repositories workspace repo-slug issues issue-id watch - Check if current user is watching a issue [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --issue-id
    operation put repositories workspace repo-slug issues issue-id watch - Watch an issue [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_issues_issue_id_watch]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Watch an issue; flags: --issue-id
    operation delete repositories workspace repo-slug issues issue-id watch - Stop watching an issue [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_issues_issue_id_watch]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Stop watching an issue; flags: --issue-id
    operation get repositories workspace repo-slug merge-base revspec - Get the common ancestor between two commits [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --revspec
    operation get repositories workspace repo-slug milestones - List milestones [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation get repositories workspace repo-slug milestones milestone-id - Get a milestone [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --milestone-id
    operation get repositories workspace repo-slug override-settings - Retrieve the inheritance state for repository settings [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation put repositories workspace repo-slug override-settings - Set the inheritance state for repository settings [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_override_settings]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Set the inheritance state for repository settings
    operation get repositories workspace repo-slug patch spec - Get a patch for two commits [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --spec
    operation get repositories workspace repo-slug permissions-config groups - List explicit group permissions for a repository [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation get repositories workspace repo-slug permissions-config groups group-slug - Get an explicit group permission for a repository [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --group-slug
    operation put repositories workspace repo-slug permissions-config groups group-slug - Update an explicit group permission for a repository [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_permissions_config_groups_group_slug]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update an explicit group permission for a repository; flags: --group-slug
    operation delete repositories workspace repo-slug permissions-config groups group-slug - Delete an explicit group permission for a repository [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_permissions_config_groups_group_slug]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete an explicit group permission for a repository; flags: --group-slug
    operation get repositories workspace repo-slug permissions-config users - List explicit user permissions for a repository [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation get repositories workspace repo-slug permissions-config users selected-user-id - Get an explicit user permission for a repository [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --selected-user-id
    operation put repositories workspace repo-slug permissions-config users selected-user-id - Update an explicit user permission for a repository [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_permissions_config_users_selected_user_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update an explicit user permission for a repository; flags: --selected-user-id
    operation delete repositories workspace repo-slug permissions-config users selected-user-id - Delete an explicit user permission for a repository [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_permissions_config_users_selected_user_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete an explicit user permission for a repository; flags: --selected-user-id
    operation get repositories workspace repo-slug pipelines - List pipelines [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --creator.uuid, --target.ref-type, --target.ref-name, --target.branch, --target.commit.hash, --target.selector.pattern, --target.selector.type, --created-on, --trigger-type, --status, --sort, --page, --pagelen
    operation get repositories workspace repo-slug pipelines-config caches - List caches [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation delete repositories workspace repo-slug pipelines-config caches - Delete caches [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_pipelines_config_caches]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete caches; flags: --name
    operation delete repositories workspace repo-slug pipelines-config caches cache-uuid - Delete a cache [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_pipelines_config_caches_cache_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a cache; flags: --cache-uuid
    operation get repositories workspace repo-slug pipelines-config caches cache-uuid content-uri - Get cache content URI [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --cache-uuid
    operation get repositories workspace repo-slug pipelines-config runners - Get repository runners [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation post repositories workspace repo-slug pipelines-config runners - Create repository runner [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_pipelines_config_runners]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create repository runner
    operation get repositories workspace repo-slug pipelines-config runners runner-uuid - Get repository runner [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --runner-uuid
    operation put repositories workspace repo-slug pipelines-config runners runner-uuid - Update repository runner [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_pipelines_config_runners_runner_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update repository runner; flags: --runner-uuid
    operation delete repositories workspace repo-slug pipelines-config runners runner-uuid - Delete repository runner [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_pipelines_config_runners_runner_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete repository runner; flags: --runner-uuid
    operation get repositories workspace repo-slug pipelines pipeline-uuid steps - List steps for a pipeline [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pipeline-uuid
    operation get repositories workspace repo-slug pipelines pipeline-uuid steps step-uuid - Get a step of a pipeline [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pipeline-uuid, --step-uuid
    operation get repositories workspace repo-slug pipelines pipeline-uuid steps step-uuid log - Get log file for a step [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pipeline-uuid, --step-uuid
    operation get repositories workspace repo-slug pipelines pipeline-uuid steps step-uuid logs log-uuid - Get the logs for the build container or a service container for a given step of a pipeline. [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pipeline-uuid, --step-uuid, --log-uuid
    operation get repositories workspace repo-slug pipelines pipeline-uuid steps step-uuid test-reports - Get a summary of test reports for a given step of a pipeline. [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pipeline-uuid, --step-uuid
    operation get repositories workspace repo-slug pipelines pipeline-uuid steps step-uuid test-reports test-cases - Get test cases for a given step of a pipeline. [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pipeline-uuid, --step-uuid
    operation get repositories workspace repo-slug pipelines pipeline-uuid steps step-uuid test-reports test-cases test-case-uuid test-case-reasons - Get test case reasons (output) for a given test case in a step of a pipeline. [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pipeline-uuid, --step-uuid, --test-case-uuid
    operation get repositories workspace repo-slug pipelines-config - Get configuration [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation put repositories workspace repo-slug pipelines-config - Update configuration [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_pipelines_config]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update configuration
    operation put repositories workspace repo-slug pipelines-config build-number - Update the next build number [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_pipelines_config_build_number]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update the next build number
    operation get repositories workspace repo-slug pipelines-config schedules - List schedules [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation post repositories workspace repo-slug pipelines-config schedules - Create a schedule [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_pipelines_config_schedules]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a schedule
    operation get repositories workspace repo-slug pipelines-config schedules schedule-uuid - Get a schedule [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --schedule-uuid
    operation put repositories workspace repo-slug pipelines-config schedules schedule-uuid - Update a schedule [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_pipelines_config_schedules_schedule_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a schedule; flags: --schedule-uuid
    operation delete repositories workspace repo-slug pipelines-config schedules schedule-uuid - Delete a schedule [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_pipelines_config_schedules_schedule_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a schedule; flags: --schedule-uuid
    operation get repositories workspace repo-slug pipelines-config schedules schedule-uuid executions - List executions of a schedule [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --schedule-uuid
    operation get repositories workspace repo-slug pipelines-config ssh key-pair - Get SSH key pair [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation put repositories workspace repo-slug pipelines-config ssh key-pair - Update SSH key pair [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_pipelines_config_ssh_key_pair]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update SSH key pair
    operation delete repositories workspace repo-slug pipelines-config ssh key-pair - Delete SSH key pair [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_pipelines_config_ssh_key_pair]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete SSH key pair
    operation get repositories workspace repo-slug pipelines-config ssh known-hosts - List known hosts [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation post repositories workspace repo-slug pipelines-config ssh known-hosts - Create a known host [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_pipelines_config_ssh_known_hosts]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a known host
    operation get repositories workspace repo-slug pipelines-config ssh known-hosts known-host-uuid - Get a known host [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --known-host-uuid
    operation put repositories workspace repo-slug pipelines-config ssh known-hosts known-host-uuid - Update a known host [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_pipelines_config_ssh_known_hosts_known_host_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a known host; flags: --known-host-uuid
    operation delete repositories workspace repo-slug pipelines-config ssh known-hosts known-host-uuid - Delete a known host [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_pipelines_config_ssh_known_hosts_known_host_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a known host; flags: --known-host-uuid
    operation get repositories workspace repo-slug pipelines-config variables - List variables for a repository [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation post repositories workspace repo-slug pipelines-config variables - Create a variable for a repository [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_pipelines_config_variables]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a variable for a repository
    operation get repositories workspace repo-slug pipelines-config variables variable-uuid - Get a variable for a repository [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --variable-uuid
    operation put repositories workspace repo-slug pipelines-config variables variable-uuid - Update a variable for a repository [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_pipelines_config_variables_variable_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a variable for a repository; flags: --variable-uuid
    operation delete repositories workspace repo-slug pipelines-config variables variable-uuid - Delete a variable for a repository [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_pipelines_config_variables_variable_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a variable for a repository; flags: --variable-uuid
    operation get repositories workspace repo-slug properties app-key property-name - Get a repository application property [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --app-key, --property-name
    operation put repositories workspace repo-slug properties app-key property-name - Update a repository application property [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_properties_app_key_property_name]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a repository application property; flags: --app-key, --property-name
    operation delete repositories workspace repo-slug properties app-key property-name - Delete a repository application property [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_properties_app_key_property_name]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a repository application property; flags: --app-key, --property-name
    operation get repositories workspace repo-slug pullrequests - List pull requests [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --state
    operation get repositories workspace repo-slug pullrequests activity - List a pull request activity log [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation put repositories workspace repo-slug pullrequests pull-request-id - Update a pull request [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_pullrequests_pull_request_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a pull request; flags: --pull-request-id
    operation get repositories workspace repo-slug pullrequests pull-request-id activity - List a pull request activity log [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pull-request-id
    operation post repositories workspace repo-slug pullrequests pull-request-id approve - Approve a pull request [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_pullrequests_pull_request_id_approve]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Approve a pull request; flags: --pull-request-id
    operation delete repositories workspace repo-slug pullrequests pull-request-id approve - Unapprove a pull request [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_pullrequests_pull_request_id_approve]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Unapprove a pull request; flags: --pull-request-id
    operation get repositories workspace repo-slug pullrequests pull-request-id comments - List comments on a pull request [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pull-request-id
    operation post repositories workspace repo-slug pullrequests pull-request-id comments - Create a comment on a pull request [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_pullrequests_pull_request_id_comments]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a comment on a pull request; flags: --pull-request-id
    operation get repositories workspace repo-slug pullrequests pull-request-id comments comment-id - Get a comment on a pull request [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pull-request-id, --comment-id
    operation put repositories workspace repo-slug pullrequests pull-request-id comments comment-id - Update a comment on a pull request [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_pullrequests_pull_request_id_comments_comment_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a comment on a pull request; flags: --pull-request-id, --comment-id
    operation delete repositories workspace repo-slug pullrequests pull-request-id comments comment-id - Delete a comment on a pull request [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_pullrequests_pull_request_id_comments_comment_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a comment on a pull request; flags: --pull-request-id, --comment-id
    operation post repositories workspace repo-slug pullrequests pull-request-id comments comment-id resolve - Resolve a comment thread [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_pullrequests_pull_request_id_comments_comment_id_resolve]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Resolve a comment thread; flags: --pull-request-id, --comment-id
    operation delete repositories workspace repo-slug pullrequests pull-request-id comments comment-id resolve - Reopen a comment thread [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_pullrequests_pull_request_id_comments_comment_id_resolve]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Reopen a comment thread; flags: --pull-request-id, --comment-id
    operation get repositories workspace repo-slug pullrequests pull-request-id commits - List commits on a pull request [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pull-request-id
    operation get repositories workspace repo-slug pullrequests pull-request-id conflicts - Get file conflicts for a pull request [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pull-request-id
    operation get repositories workspace repo-slug pullrequests pull-request-id diff - List changes in a pull request [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pull-request-id
    operation get repositories workspace repo-slug pullrequests pull-request-id diffstat - Get the diff stat for a pull request [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pull-request-id
    operation get repositories workspace repo-slug pullrequests pull-request-id merge task-status task-id - Get the merge task status for a pull request [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pull-request-id, --task-id
    operation get repositories workspace repo-slug pullrequests pull-request-id patch - Get the patch for a pull request [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pull-request-id
    operation post repositories workspace repo-slug pullrequests pull-request-id request-changes - Request changes for a pull request [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_pullrequests_pull_request_id_request_changes]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Request changes for a pull request; flags: --pull-request-id
    operation delete repositories workspace repo-slug pullrequests pull-request-id request-changes - Remove change request for a pull request [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_pullrequests_pull_request_id_request_changes]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Remove change request for a pull request; flags: --pull-request-id
    operation get repositories workspace repo-slug pullrequests pull-request-id statuses - List commit statuses for a pull request [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pull-request-id, --q, --sort
    operation get repositories workspace repo-slug pullrequests pull-request-id tasks - List tasks on a pull request [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pull-request-id, --q, --sort, --pagelen
    operation post repositories workspace repo-slug pullrequests pull-request-id tasks - Create a task on a pull request [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_pullrequests_pull_request_id_tasks]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a task on a pull request; flags: --pull-request-id
    operation get repositories workspace repo-slug pullrequests pull-request-id tasks task-id - Get a task on a pull request [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pull-request-id, --task-id
    operation put repositories workspace repo-slug pullrequests pull-request-id tasks task-id - Update a task on a pull request [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_pullrequests_pull_request_id_tasks_task_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a task on a pull request; flags: --pull-request-id, --task-id
    operation delete repositories workspace repo-slug pullrequests pull-request-id tasks task-id - Delete a task on a pull request [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_pullrequests_pull_request_id_tasks_task_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a task on a pull request; flags: --pull-request-id, --task-id
    operation get repositories workspace repo-slug pullrequests pullrequest-id properties app-key property-name - Get a pull request application property [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --pullrequest-id, --app-key, --property-name
    operation put repositories workspace repo-slug pullrequests pullrequest-id properties app-key property-name - Update a pull request application property [intent=reverse_etl availability=implemented write=op_put_repositories_workspace_repo_slug_pullrequests_pullrequest_id_properties_app_key_property_name]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a pull request application property; flags: --pullrequest-id, --app-key, --property-name
    operation delete repositories workspace repo-slug pullrequests pullrequest-id properties app-key property-name - Delete a pull request application property [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_pullrequests_pullrequest_id_properties_app_key_property_name]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a pull request application property; flags: --pullrequest-id, --app-key, --property-name
    operation get repositories workspace repo-slug refs - List branches and tags [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --q, --sort
    operation get repositories workspace repo-slug refs branches - List open branches [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --q, --sort
    operation post repositories workspace repo-slug refs branches - Create a branch [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_refs_branches]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a branch
    operation get repositories workspace repo-slug refs branches name - Get a branch [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --name
    operation delete repositories workspace repo-slug refs branches name - Delete a branch [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_refs_branches_name]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a branch; flags: --name
    operation get repositories workspace repo-slug refs tags - List tags [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --q, --sort
    operation post repositories workspace repo-slug refs tags - Create a tag [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_refs_tags]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a tag
    operation get repositories workspace repo-slug refs tags name - Get a tag [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --name
    operation delete repositories workspace repo-slug refs tags name - Delete a tag [intent=reverse_etl availability=implemented write=op_delete_repositories_workspace_repo_slug_refs_tags_name]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a tag; flags: --name
    operation get repositories workspace repo-slug src - Get the root directory of the main branch [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --format
    operation post repositories workspace repo-slug src - Create a commit by uploading a file [intent=reverse_etl availability=implemented write=op_post_repositories_workspace_repo_slug_src]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a commit by uploading a file; flags: --message, --author, --parents, --files, --branch
    operation get repositories workspace repo-slug src commit path - Get file or directory contents [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --commit, --path, --format, --q, --sort, --max-depth
    operation get repositories workspace repo-slug versions - List defined versions for issues [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation get repositories workspace repo-slug versions version-id - Get a defined version for issues [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --version-id
    operation get repositories workspace repo-slug watchers - List repositories watchers [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug
    operation get snippets - List snippets [intent=direct_read availability=implemented]; flags: --role
    operation post snippets - Create a snippet [intent=reverse_etl availability=implemented write=op_post_snippets]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a snippet
    operation get snippets workspace encoded-id - Get a snippet [intent=direct_read availability=implemented]; flags: --workspace, --encoded-id
    operation put snippets workspace encoded-id - Update a snippet [intent=reverse_etl availability=implemented write=op_put_snippets_workspace_encoded_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a snippet; flags: --encoded-id
    operation delete snippets workspace encoded-id - Delete a snippet [intent=reverse_etl availability=implemented write=op_delete_snippets_workspace_encoded_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a snippet; flags: --encoded-id
    operation get snippets workspace encoded-id comments - List comments on a snippet [intent=direct_read availability=implemented]; flags: --workspace, --encoded-id
    operation post snippets workspace encoded-id comments - Create a comment on a snippet [intent=reverse_etl availability=implemented write=op_post_snippets_workspace_encoded_id_comments]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a comment on a snippet; flags: --encoded-id
    operation get snippets workspace encoded-id comments comment-id - Get a comment on a snippet [intent=direct_read availability=implemented]; flags: --workspace, --encoded-id, --comment-id
    operation put snippets workspace encoded-id comments comment-id - Update a comment on a snippet [intent=reverse_etl availability=implemented write=op_put_snippets_workspace_encoded_id_comments_comment_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a comment on a snippet; flags: --encoded-id, --comment-id
    operation delete snippets workspace encoded-id comments comment-id - Delete a comment on a snippet [intent=reverse_etl availability=implemented write=op_delete_snippets_workspace_encoded_id_comments_comment_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a comment on a snippet; flags: --encoded-id, --comment-id
    operation get snippets workspace encoded-id commits - List snippet changes [intent=direct_read availability=implemented]; flags: --workspace, --encoded-id
    operation get snippets workspace encoded-id commits revision - Get a previous snippet change [intent=direct_read availability=implemented]; flags: --workspace, --encoded-id, --revision
    operation get snippets workspace encoded-id files path - Get a snippet's raw file at HEAD [intent=direct_read availability=implemented]; flags: --workspace, --encoded-id, --path
    operation get snippets workspace encoded-id watch - Check if the current user is watching a snippet [intent=direct_read availability=implemented]; flags: --workspace, --encoded-id
    operation put snippets workspace encoded-id watch - Watch a snippet [intent=reverse_etl availability=implemented write=op_put_snippets_workspace_encoded_id_watch]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Watch a snippet; flags: --encoded-id
    operation delete snippets workspace encoded-id watch - Stop watching a snippet [intent=reverse_etl availability=implemented write=op_delete_snippets_workspace_encoded_id_watch]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Stop watching a snippet; flags: --encoded-id
    operation get snippets workspace encoded-id watchers - List users watching a snippet [intent=direct_read availability=implemented]; flags: --workspace, --encoded-id
    operation get snippets workspace encoded-id node-id - Get a previous revision of a snippet [intent=direct_read availability=implemented]; flags: --workspace, --encoded-id, --node-id
    operation put snippets workspace encoded-id node-id - Update a previous revision of a snippet [intent=reverse_etl availability=implemented write=op_put_snippets_workspace_encoded_id_node_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a previous revision of a snippet; flags: --encoded-id, --node-id
    operation delete snippets workspace encoded-id node-id - Delete a previous revision of a snippet [intent=reverse_etl availability=implemented write=op_delete_snippets_workspace_encoded_id_node_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a previous revision of a snippet; flags: --encoded-id, --node-id
    operation get snippets workspace encoded-id node-id files path - Get a snippet's raw file [intent=direct_read availability=implemented]; flags: --workspace, --encoded-id, --node-id, --path
    operation get snippets workspace encoded-id revision diff - Get snippet changes between versions [intent=direct_read availability=implemented]; flags: --workspace, --encoded-id, --revision, --path
    operation get snippets workspace encoded-id revision patch - Get snippet patch between versions [intent=direct_read availability=implemented]; flags: --workspace, --encoded-id, --revision
    operation get teams username pipelines-config variables - List variables for an account [intent=direct_read availability=implemented]; flags: --username
    operation post teams username pipelines-config variables - Create a variable for a user [intent=reverse_etl availability=implemented write=op_post_teams_username_pipelines_config_variables]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a variable for a user; flags: --username
    operation get teams username pipelines-config variables variable-uuid - Get a variable for a team [intent=direct_read availability=implemented]; flags: --username, --variable-uuid
    operation put teams username pipelines-config variables variable-uuid - Update a variable for a team [intent=reverse_etl availability=implemented write=op_put_teams_username_pipelines_config_variables_variable_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a variable for a team; flags: --username, --variable-uuid
    operation delete teams username pipelines-config variables variable-uuid - Delete a variable for a team [intent=reverse_etl availability=implemented write=op_delete_teams_username_pipelines_config_variables_variable_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a variable for a team; flags: --username, --variable-uuid
    operation get teams username search code - Search for code in a team's repositories [intent=direct_read availability=implemented]; flags: --username, --search-query, --page, --pagelen
    operation get user - Get current user [intent=direct_read availability=implemented]
    operation get user emails - List email addresses for current user [intent=direct_read availability=implemented]
    operation get user emails email - Get an email address for current user [intent=direct_read availability=implemented]; flags: --email
    operation get user permissions repositories - List repository permissions for a user [intent=direct_read availability=implemented]; flags: --q, --sort
    operation get user permissions workspaces - List workspaces for the current user [intent=direct_read availability=implemented]; flags: --q, --sort
    operation get user workspaces - List workspaces for the current user [intent=direct_read availability=implemented]; flags: --sort, --administrator
    operation get user workspaces workspace permission - Get user permission on a workspace [intent=direct_read availability=implemented]; flags: --workspace
    operation get user workspaces workspace permissions repositories - List repository permissions in a workspace for a user [intent=direct_read availability=implemented]; flags: --workspace, --q, --sort
    operation get users selected-user - Get a user [intent=direct_read availability=implemented]; flags: --selected-user
    operation get users selected-user gpg-keys - List GPG keys [intent=direct_read availability=implemented]; flags: --selected-user
    operation post users selected-user gpg-keys - Add a new GPG key [intent=reverse_etl availability=implemented write=op_post_users_selected_user_gpg_keys]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Add a new GPG key; flags: --selected-user
    operation get users selected-user gpg-keys fingerprint - Get a GPG key [intent=direct_read availability=implemented]; flags: --selected-user, --fingerprint
    operation delete users selected-user gpg-keys fingerprint - Delete a GPG key [intent=reverse_etl availability=implemented write=op_delete_users_selected_user_gpg_keys_fingerprint]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a GPG key; flags: --selected-user, --fingerprint
    operation get users selected-user pipelines-config variables - List variables for a user [intent=direct_read availability=implemented]; flags: --selected-user
    operation post users selected-user pipelines-config variables - Create a variable for a user [intent=reverse_etl availability=implemented write=op_post_users_selected_user_pipelines_config_variables]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a variable for a user; flags: --selected-user
    operation get users selected-user pipelines-config variables variable-uuid - Get a variable for a user [intent=direct_read availability=implemented]; flags: --selected-user, --variable-uuid
    operation put users selected-user pipelines-config variables variable-uuid - Update a variable for a user [intent=reverse_etl availability=implemented write=op_put_users_selected_user_pipelines_config_variables_variable_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a variable for a user; flags: --selected-user, --variable-uuid
    operation delete users selected-user pipelines-config variables variable-uuid - Delete a variable for a user [intent=reverse_etl availability=implemented write=op_delete_users_selected_user_pipelines_config_variables_variable_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a variable for a user; flags: --selected-user, --variable-uuid
    operation get users selected-user properties app-key property-name - Get a user application property [intent=direct_read availability=implemented]; flags: --selected-user, --app-key, --property-name
    operation put users selected-user properties app-key property-name - Update a user application property [intent=reverse_etl availability=implemented write=op_put_users_selected_user_properties_app_key_property_name]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a user application property; flags: --selected-user, --app-key, --property-name
    operation delete users selected-user properties app-key property-name - Delete a user application property [intent=reverse_etl availability=implemented write=op_delete_users_selected_user_properties_app_key_property_name]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a user application property; flags: --selected-user, --app-key, --property-name
    operation get users selected-user search code - Search for code in a user's repositories [intent=direct_read availability=implemented]; flags: --selected-user, --search-query, --page, --pagelen
    operation get users selected-user ssh-keys - List SSH keys [intent=direct_read availability=implemented]; flags: --selected-user
    operation post users selected-user ssh-keys - Add a new SSH key [intent=reverse_etl availability=implemented write=op_post_users_selected_user_ssh_keys]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Add a new SSH key; flags: --selected-user, --expires-on
    operation get users selected-user ssh-keys key-id - Get a SSH key [intent=direct_read availability=implemented]; flags: --selected-user, --key-id
    operation put users selected-user ssh-keys key-id - Update a SSH key [intent=reverse_etl availability=implemented write=op_put_users_selected_user_ssh_keys_key_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a SSH key; flags: --selected-user, --key-id
    operation delete users selected-user ssh-keys key-id - Delete a SSH key [intent=reverse_etl availability=implemented write=op_delete_users_selected_user_ssh_keys_key_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a SSH key; flags: --selected-user, --key-id
    operation get workspaces workspace - Get a workspace [intent=direct_read availability=implemented]; flags: --workspace
    operation get workspaces workspace hooks - List webhooks for a workspace [intent=direct_read availability=implemented]; flags: --workspace
    operation post workspaces workspace hooks - Create a webhook for a workspace [intent=reverse_etl availability=implemented write=op_post_workspaces_workspace_hooks]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a webhook for a workspace
    operation get workspaces workspace hooks uid - Get a webhook for a workspace [intent=direct_read availability=implemented]; flags: --workspace, --uid
    operation put workspaces workspace hooks uid - Update a webhook for a workspace [intent=reverse_etl availability=implemented write=op_put_workspaces_workspace_hooks_uid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a webhook for a workspace; flags: --uid
    operation delete workspaces workspace hooks uid - Delete a webhook for a workspace [intent=reverse_etl availability=implemented write=op_delete_workspaces_workspace_hooks_uid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a webhook for a workspace; flags: --uid
    operation get workspaces workspace members - List users in a workspace [intent=direct_read availability=implemented]; flags: --workspace
    operation get workspaces workspace members member - Get user membership for a workspace [intent=direct_read availability=implemented]; flags: --workspace, --member
    operation get workspaces workspace permissions - List user permissions in a workspace [intent=direct_read availability=implemented]; flags: --workspace, --q
    operation get workspaces workspace permissions repositories - List all repository permissions for a workspace [intent=direct_read availability=implemented]; flags: --workspace, --q, --sort
    operation get workspaces workspace permissions repositories repo-slug - List a repository permissions for a workspace [intent=direct_read availability=implemented]; flags: --workspace, --repo-slug, --q, --sort
    operation get workspaces workspace pipelines-config identity oidc .well-known openid-configuration - Get OpenID configuration for OIDC in Pipelines [intent=direct_read availability=implemented]; flags: --workspace
    operation get workspaces workspace pipelines-config identity oidc keys.json - Get keys for OIDC in Pipelines [intent=direct_read availability=implemented]; flags: --workspace
    operation get workspaces workspace pipelines-config runners - Get workspace runners [intent=direct_read availability=implemented]; flags: --workspace
    operation post workspaces workspace pipelines-config runners - Create workspace runner [intent=reverse_etl availability=implemented write=op_post_workspaces_workspace_pipelines_config_runners]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create workspace runner
    operation get workspaces workspace pipelines-config runners runner-uuid - Get workspace runner [intent=direct_read availability=implemented]; flags: --workspace, --runner-uuid
    operation put workspaces workspace pipelines-config runners runner-uuid - Update workspace runner [intent=reverse_etl availability=implemented write=op_put_workspaces_workspace_pipelines_config_runners_runner_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update workspace runner; flags: --runner-uuid
    operation delete workspaces workspace pipelines-config runners runner-uuid - Delete workspace runner [intent=reverse_etl availability=implemented write=op_delete_workspaces_workspace_pipelines_config_runners_runner_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete workspace runner; flags: --runner-uuid
    operation get workspaces workspace pipelines-config variables - List variables for a workspace [intent=direct_read availability=implemented]; flags: --workspace
    operation post workspaces workspace pipelines-config variables - Create a variable for a workspace [intent=reverse_etl availability=implemented write=op_post_workspaces_workspace_pipelines_config_variables]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a variable for a workspace
    operation get workspaces workspace pipelines-config variables variable-uuid - Get variable for a workspace [intent=direct_read availability=implemented]; flags: --workspace, --variable-uuid
    operation put workspaces workspace pipelines-config variables variable-uuid - Update variable for a workspace [intent=reverse_etl availability=implemented write=op_put_workspaces_workspace_pipelines_config_variables_variable_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update variable for a workspace; flags: --variable-uuid
    operation delete workspaces workspace pipelines-config variables variable-uuid - Delete a variable for a workspace [intent=reverse_etl availability=implemented write=op_delete_workspaces_workspace_pipelines_config_variables_variable_uuid]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a variable for a workspace; flags: --variable-uuid
    operation post workspaces workspace projects - Create a project in a workspace [intent=reverse_etl availability=implemented write=op_post_workspaces_workspace_projects]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a project in a workspace
    operation get workspaces workspace projects project-key - Get a project for a workspace [intent=direct_read availability=implemented]; flags: --workspace, --project-key
    operation put workspaces workspace projects project-key - Update a project for a workspace [intent=reverse_etl availability=implemented write=op_put_workspaces_workspace_projects_project_key]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update a project for a workspace; flags: --project-key
    operation delete workspaces workspace projects project-key - Delete a project for a workspace [intent=reverse_etl availability=implemented write=op_delete_workspaces_workspace_projects_project_key]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a project for a workspace; flags: --project-key
    operation get workspaces workspace projects project-key branching-model - Get the branching model for a project [intent=direct_read availability=implemented]; flags: --workspace, --project-key
    operation get workspaces workspace projects project-key branching-model settings - Get the branching model config for a project [intent=direct_read availability=implemented]; flags: --workspace, --project-key
    operation put workspaces workspace projects project-key branching-model settings - Update the branching model config for a project [intent=reverse_etl availability=implemented write=op_put_workspaces_workspace_projects_project_key_branching_model_settings]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update the branching model config for a project; flags: --project-key
    operation get workspaces workspace projects project-key default-reviewers - List the default reviewers in a project [intent=direct_read availability=implemented]; flags: --workspace, --project-key
    operation get workspaces workspace projects project-key default-reviewers selected-user - Get a default reviewer [intent=direct_read availability=implemented]; flags: --workspace, --project-key, --selected-user
    operation put workspaces workspace projects project-key default-reviewers selected-user - Add the specific user as a default reviewer for the project [intent=reverse_etl availability=implemented write=op_put_workspaces_workspace_projects_project_key_default_reviewers_selected_user]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Add the specific user as a default reviewer for the project; flags: --project-key, --selected-user
    operation delete workspaces workspace projects project-key default-reviewers selected-user - Remove the specific user from the project's default reviewers [intent=reverse_etl availability=implemented write=op_delete_workspaces_workspace_projects_project_key_default_reviewers_selected_user]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Remove the specific user from the project's default reviewers; flags: --project-key, --selected-user
    operation get workspaces workspace projects project-key deploy-keys - List project deploy keys [intent=direct_read availability=implemented]; flags: --workspace, --project-key
    operation post workspaces workspace projects project-key deploy-keys - Create a project deploy key [intent=reverse_etl availability=implemented write=op_post_workspaces_workspace_projects_project_key_deploy_keys]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Create a project deploy key; flags: --project-key
    operation get workspaces workspace projects project-key deploy-keys key-id - Get a project deploy key [intent=direct_read availability=implemented]; flags: --workspace, --project-key, --key-id
    operation delete workspaces workspace projects project-key deploy-keys key-id - Delete a deploy key from a project [intent=reverse_etl availability=implemented write=op_delete_workspaces_workspace_projects_project_key_deploy_keys_key_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete a deploy key from a project; flags: --project-key, --key-id
    operation get workspaces workspace projects project-key permissions-config groups - List explicit group permissions for a project [intent=direct_read availability=implemented]; flags: --workspace, --project-key
    operation get workspaces workspace projects project-key permissions-config groups group-slug - Get an explicit group permission for a project [intent=direct_read availability=implemented]; flags: --workspace, --project-key, --group-slug
    operation put workspaces workspace projects project-key permissions-config groups group-slug - Update an explicit group permission for a project [intent=reverse_etl availability=implemented write=op_put_workspaces_workspace_projects_project_key_permissions_config_groups_group_slug]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update an explicit group permission for a project; flags: --project-key, --group-slug
    operation delete workspaces workspace projects project-key permissions-config groups group-slug - Delete an explicit group permission for a project [intent=reverse_etl availability=implemented write=op_delete_workspaces_workspace_projects_project_key_permissions_config_groups_group_slug]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete an explicit group permission for a project; flags: --project-key, --group-slug
    operation get workspaces workspace projects project-key permissions-config users - List explicit user permissions for a project [intent=direct_read availability=implemented]; flags: --workspace, --project-key
    operation get workspaces workspace projects project-key permissions-config users selected-user-id - Get an explicit user permission for a project [intent=direct_read availability=implemented]; flags: --workspace, --project-key, --selected-user-id
    operation put workspaces workspace projects project-key permissions-config users selected-user-id - Update an explicit user permission for a project [intent=reverse_etl availability=implemented write=op_put_workspaces_workspace_projects_project_key_permissions_config_users_selected_user_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Update an explicit user permission for a project; flags: --project-key, --selected-user-id
    operation delete workspaces workspace projects project-key permissions-config users selected-user-id - Delete an explicit user permission for a project [intent=reverse_etl availability=implemented write=op_delete_workspaces_workspace_projects_project_key_permissions_config_users_selected_user_id]; approval: reverse ETL writes require plan, preview, approval, execute; typed confirmation is required for destructive/admin/sensitive operations.; risk: Delete an explicit user permission for a project; flags: --project-key, --selected-user-id
    operation get workspaces workspace pullrequests selected-user - List workspace pull requests for a user [intent=direct_read availability=implemented]; flags: --workspace, --selected-user, --state
    operation get workspaces workspace search code - Search for code in a workspace [intent=direct_read availability=implemented]; flags: --workspace, --search-query, --page, --pagelen
    operation get workspaces workspace settings gpg public-key - Get the workspace system GPG public key(s) [intent=direct_read availability=implemented]; flags: --workspace
  Help topics:
    safety - Bitbucket writes remain plan, preview, approval, execute; generic raw API writes are disallowed.
    coverage - All 331 official Swagger operations are covered by typed streams, direct reads, or approval-gated reverse ETL writes; raw API remains disallowed.

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
