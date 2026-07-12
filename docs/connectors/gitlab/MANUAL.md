# pm connectors inspect gitlab

```text
NAME
  pm connectors inspect gitlab - GitLab connector manual

SYNOPSIS
  pm connectors inspect gitlab
  pm connectors inspect gitlab --json
  pm credentials add <name> --connector gitlab [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads GitLab projects, groups, users, and issues through the GitLab REST API v4.

ICON
  asset: icons/gitlab.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://docs.gitlab.com/ee/api/rest/deprecations.html

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  page_size
  start_date
  access_token (secret)

ETL STREAMS
  projects: 
    primary key: id
    cursor: last_activity_at
    fields: archived(), created_at(), default_branch(), description(), forks_count(), id(), last_activity_at(), name(), open_issues_count(), path(), path_with_namespace(), star_count(), visibility(), web_url()
  groups: 
    primary key: id
    cursor: created_at
    fields: created_at(), description(), full_name(), full_path(), id(), name(), parent_id(), path(), visibility(), web_url()
  users: 
    primary key: id
    cursor: created_at
    fields: bot(), created_at(), id(), is_admin(), name(), state(), username(), web_url()
  issues: 
    primary key: id
    cursor: updated_at
    fields: author_id(), closed_at(), created_at(), downvotes(), id(), iid(), project_id(), state(), title(), updated_at(), upvotes(), user_notes_count(), web_url()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  delete_api_v4_admin_ci_variables_key: 
    endpoint: DELETE /admin/ci/variables/{{ record.key }}
    required fields: key
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_admin_clusters_cluster_id: 
    endpoint: DELETE /admin/clusters/{{ record.cluster_id }}
    required fields: cluster_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_applications_id: 
    endpoint: DELETE /applications/{{ record.id }}
    required fields: id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_broadcast_messages_id: 
    endpoint: DELETE /broadcast_messages/{{ record.id }}
    required fields: id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_features_name: 
    endpoint: DELETE /features/{{ record.name }}
    required fields: name
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id: 
    endpoint: DELETE /groups/{{ record.id }}
    required fields: id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_debian_distributions_codename: 
    endpoint: DELETE /groups/{{ record.id }}/-/debian_distributions/{{ record.codename }}
    required fields: id, codename
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_packages_npm_package_package_name_dist_tags_tag: 
    endpoint: DELETE /groups/{{ record.id }}/-/packages/npm/-/package/*package_name/dist-tags/{{ record.tag }}
    required fields: id, tag
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_access_requests_user_id: 
    endpoint: DELETE /groups/{{ record.id }}/access_requests/{{ record.user_id }}
    required fields: id, user_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_badges_badge_id: 
    endpoint: DELETE /groups/{{ record.id }}/badges/{{ record.badge_id }}
    required fields: id, badge_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_billable_members_user_id: 
    endpoint: DELETE /groups/{{ record.id }}/billable_members/{{ record.user_id }}
    required fields: id, user_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_clusters_cluster_id: 
    endpoint: DELETE /groups/{{ record.id }}/clusters/{{ record.cluster_id }}
    required fields: id, cluster_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_custom_attributes_key: 
    endpoint: DELETE /groups/{{ record.id }}/custom_attributes/{{ record.key }}
    required fields: id, key
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_dependency_proxy_cache: 
    endpoint: DELETE /groups/{{ record.id }}/dependency_proxy/cache
    required fields: id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_deploy_tokens_token_id: 
    endpoint: DELETE /groups/{{ record.id }}/deploy_tokens/{{ record.token_id }}
    required fields: id, token_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_epics_epic_iid_award_emoji_award_id: 
    endpoint: DELETE /groups/{{ record.id }}/epics/{{ record.epic_iid }}/award_emoji/{{ record.award_id }}
    required fields: id, epic_iid, award_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_epics_epic_iid_notes_note_id_award_emoji_award_id: 
    endpoint: DELETE /groups/{{ record.id }}/epics/{{ record.epic_iid }}/notes/{{ record.note_id }}/award_emoji/{{ record.award_id }}
    required fields: id, epic_iid, note_id, award_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_integrations_slug: 
    endpoint: DELETE /groups/{{ record.id }}/integrations/{{ record.slug }}
    required fields: id, slug
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_invitations_email: 
    endpoint: DELETE /groups/{{ record.id }}/invitations/{{ record.email }}
    required fields: id, email
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_members_user_id: 
    endpoint: DELETE /groups/{{ record.id }}/members/{{ record.user_id }}
    required fields: id, user_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_members_user_id_override: 
    endpoint: DELETE /groups/{{ record.id }}/members/{{ record.user_id }}/override
    required fields: id, user_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_share_group_id: 
    endpoint: DELETE /groups/{{ record.id }}/share/{{ record.group_id }}
    required fields: id, group_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_shared_projects_project_id: 
    endpoint: DELETE /groups/{{ record.id }}/shared_projects/{{ record.project_id }}
    required fields: id, project_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_ssh_certificates_ssh_certificates_id: 
    endpoint: DELETE /groups/{{ record.id }}/ssh_certificates/{{ record.ssh_certificates_id }}
    required fields: id, ssh_certificates_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_uploads_secret_filename: 
    endpoint: DELETE /groups/{{ record.id }}/uploads/{{ record.secret }}/{{ record.filename }}
    required fields: id, secret, filename
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_uploads_upload_id: 
    endpoint: DELETE /groups/{{ record.id }}/uploads/{{ record.upload_id }}
    required fields: id, upload_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_variables_key: 
    endpoint: DELETE /groups/{{ record.id }}/variables/{{ record.key }}
    required fields: id, key
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_groups_id_wikis_slug: 
    endpoint: DELETE /groups/{{ record.id }}/wikis/{{ record.slug }}
    required fields: id, slug
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_hooks_hook_id: 
    endpoint: DELETE /hooks/{{ record.hook_id }}
    required fields: hook_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_hooks_hook_id_custom_headers_key: 
    endpoint: DELETE /hooks/{{ record.hook_id }}/custom_headers/{{ record.key }}
    required fields: hook_id, key
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_hooks_hook_id_url_variables_key: 
    endpoint: DELETE /hooks/{{ record.hook_id }}/url_variables/{{ record.key }}
    required fields: hook_id, key
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_integrations_jira_forge_subscriptions_id: 
    endpoint: DELETE /integrations/jira_forge/subscriptions/{{ record.id }}
    required fields: id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_namespaces_id_storage_limit_exclusion: 
    endpoint: DELETE /namespaces/{{ record.id }}/storage/limit_exclusion
    required fields: id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_organizations_id: 
    endpoint: DELETE /organizations/{{ record.id }}
    required fields: id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_packages_conan_v1_conans_package_name_package_version_package_username_package_channel: 
    endpoint: DELETE /packages/conan/v1/conans/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}
    required fields: package_name, package_version, package_username, package_channel
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_packages_npm_package_package_name_dist_tags_tag: 
    endpoint: DELETE /packages/npm/-/package/*package_name/dist-tags/{{ record.tag }}
    required fields: tag
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_personal_access_tokens_self: 
    endpoint: DELETE /personal_access_tokens/self
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_personal_access_tokens_id: 
    endpoint: DELETE /personal_access_tokens/{{ record.id }}
    required fields: id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id: 
    endpoint: DELETE /projects/{{ record.id }}
    required fields: id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_access_requests_user_id: 
    endpoint: DELETE /projects/{{ record.id }}/access_requests/{{ record.user_id }}
    required fields: id, user_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_alert_management_alerts_alert_iid_metric_images_metric_image_id: 
    endpoint: DELETE /projects/{{ record.id }}/alert_management_alerts/{{ record.alert_iid }}/metric_images/{{ record.metric_image_id }}
    required fields: id, alert_iid, metric_image_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_artifacts: 
    endpoint: DELETE /projects/{{ record.id }}/artifacts
    required fields: id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_badges_badge_id: 
    endpoint: DELETE /projects/{{ record.id }}/badges/{{ record.badge_id }}
    required fields: id, badge_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_cluster_agents_agent_id: 
    endpoint: DELETE /projects/{{ record.id }}/cluster_agents/{{ record.agent_id }}
    required fields: id, agent_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_cluster_agents_agent_id_tokens_token_id: 
    endpoint: DELETE /projects/{{ record.id }}/cluster_agents/{{ record.agent_id }}/tokens/{{ record.token_id }}
    required fields: id, agent_id, token_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_clusters_cluster_id: 
    endpoint: DELETE /projects/{{ record.id }}/clusters/{{ record.cluster_id }}
    required fields: id, cluster_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_custom_attributes_key: 
    endpoint: DELETE /projects/{{ record.id }}/custom_attributes/{{ record.key }}
    required fields: id, key
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_debian_distributions_codename: 
    endpoint: DELETE /projects/{{ record.id }}/debian_distributions/{{ record.codename }}
    required fields: id, codename
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_deploy_keys_key_id: 
    endpoint: DELETE /projects/{{ record.id }}/deploy_keys/{{ record.key_id }}
    required fields: id, key_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_deploy_tokens_token_id: 
    endpoint: DELETE /projects/{{ record.id }}/deploy_tokens/{{ record.token_id }}
    required fields: id, token_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_deployments_deployment_id: 
    endpoint: DELETE /projects/{{ record.id }}/deployments/{{ record.deployment_id }}
    required fields: id, deployment_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_environments_review_apps: 
    endpoint: DELETE /projects/{{ record.id }}/environments/review_apps
    required fields: id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_environments_environment_id: 
    endpoint: DELETE /projects/{{ record.id }}/environments/{{ record.environment_id }}
    required fields: id, environment_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_error_tracking_client_keys_key_id: 
    endpoint: DELETE /projects/{{ record.id }}/error_tracking/client_keys/{{ record.key_id }}
    required fields: id, key_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_feature_flags_feature_flag_name: 
    endpoint: DELETE /projects/{{ record.id }}/feature_flags/{{ record.feature_flag_name }}
    required fields: id, feature_flag_name
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_feature_flags_user_lists_iid: 
    endpoint: DELETE /projects/{{ record.id }}/feature_flags_user_lists/{{ record.iid }}
    required fields: id, iid
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_fork: 
    endpoint: DELETE /projects/{{ record.id }}/fork
    required fields: id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_freeze_periods_freeze_period_id: 
    endpoint: DELETE /projects/{{ record.id }}/freeze_periods/{{ record.freeze_period_id }}
    required fields: id, freeze_period_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_hooks_hook_id: 
    endpoint: DELETE /projects/{{ record.id }}/hooks/{{ record.hook_id }}
    required fields: id, hook_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_hooks_hook_id_custom_headers_key: 
    endpoint: DELETE /projects/{{ record.id }}/hooks/{{ record.hook_id }}/custom_headers/{{ record.key }}
    required fields: id, hook_id, key
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_hooks_hook_id_url_variables_key: 
    endpoint: DELETE /projects/{{ record.id }}/hooks/{{ record.hook_id }}/url_variables/{{ record.key }}
    required fields: id, hook_id, key
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_integrations_slug: 
    endpoint: DELETE /projects/{{ record.id }}/integrations/{{ record.slug }}
    required fields: id, slug
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_invitations_email: 
    endpoint: DELETE /projects/{{ record.id }}/invitations/{{ record.email }}
    required fields: id, email
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_issues_issue_iid: 
    endpoint: DELETE /projects/{{ record.id }}/issues/{{ record.issue_iid }}
    required fields: id, issue_iid
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_issues_issue_iid_award_emoji_award_id: 
    endpoint: DELETE /projects/{{ record.id }}/issues/{{ record.issue_iid }}/award_emoji/{{ record.award_id }}
    required fields: id, issue_iid, award_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_issues_issue_iid_links_issue_link_id: 
    endpoint: DELETE /projects/{{ record.id }}/issues/{{ record.issue_iid }}/links/{{ record.issue_link_id }}
    required fields: id, issue_iid, issue_link_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_issues_issue_iid_metric_images_metric_image_id: 
    endpoint: DELETE /projects/{{ record.id }}/issues/{{ record.issue_iid }}/metric_images/{{ record.metric_image_id }}
    required fields: id, issue_iid, metric_image_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_issues_issue_iid_notes_note_id_award_emoji_award_id: 
    endpoint: DELETE /projects/{{ record.id }}/issues/{{ record.issue_iid }}/notes/{{ record.note_id }}/award_emoji/{{ record.award_id }}
    required fields: id, issue_iid, note_id, award_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_job_token_scope_allowlist_target_project_id: 
    endpoint: DELETE /projects/{{ record.id }}/job_token_scope/allowlist/{{ record.target_project_id }}
    required fields: id, target_project_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_job_token_scope_groups_allowlist_target_group_id: 
    endpoint: DELETE /projects/{{ record.id }}/job_token_scope/groups_allowlist/{{ record.target_group_id }}
    required fields: id, target_group_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_jobs_job_id_artifacts: 
    endpoint: DELETE /projects/{{ record.id }}/jobs/{{ record.job_id }}/artifacts
    required fields: id, job_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_members_user_id: 
    endpoint: DELETE /projects/{{ record.id }}/members/{{ record.user_id }}
    required fields: id, user_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_merge_requests_merge_request_iid: 
    endpoint: DELETE /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}
    required fields: id, merge_request_iid
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_merge_requests_merge_request_iid_award_emoji_award_id: 
    endpoint: DELETE /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/award_emoji/{{ record.award_id }}
    required fields: id, merge_request_iid, award_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_merge_requests_merge_request_iid_context_commits: 
    endpoint: DELETE /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/context_commits
    required fields: id, merge_request_iid
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_merge_requests_merge_request_iid_draft_notes_draft_note_id: 
    endpoint: DELETE /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/draft_notes/{{ record.draft_note_id }}
    required fields: id, merge_request_iid, draft_note_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_merge_requests_merge_request_iid_notes_note_id_award_emoji_award_id: 
    endpoint: DELETE /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/notes/{{ record.note_id }}/award_emoji/{{ record.award_id }}
    required fields: id, merge_request_iid, note_id, award_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_packages_conan_v1_conans_package_name_package_version_package_username_package_channel: 
    endpoint: DELETE /projects/{{ record.id }}/packages/conan/v1/conans/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}
    required fields: id, package_name, package_version, package_username, package_channel
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_packages_conan_v2_conans_package_name_package_version_package_username_package_channel_revisions_recipe_revision: 
    endpoint: DELETE /projects/{{ record.id }}/packages/conan/v2/conans/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}/revisions/{{ record.recipe_revision }}
    required fields: id, package_name, package_version, package_username, package_channel, recipe_revision
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_packages_conan_v2_conans_package_name_package_version_package_username_package_channel_revisions_recipe_revision_packages_conan_package_reference_revisions_package_revision: 
    endpoint: DELETE /projects/{{ record.id }}/packages/conan/v2/conans/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}/revisions/{{ record.recipe_revision }}/packages/{{ record.conan_package_reference }}/revisions/{{ record.package_revision }}
    required fields: id, package_name, package_version, package_username, package_channel, recipe_revision, conan_package_reference, package_revision
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_packages_npm_package_package_name_dist_tags_tag: 
    endpoint: DELETE /projects/{{ record.id }}/packages/npm/-/package/*package_name/dist-tags/{{ record.tag }}
    required fields: id, tag
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_packages_nuget_package_name_package_version: 
    endpoint: DELETE /projects/{{ record.id }}/packages/nuget/*package_name/*package_version
    required fields: id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_packages_protection_rules_package_protection_rule_id: 
    endpoint: DELETE /projects/{{ record.id }}/packages/protection/rules/{{ record.package_protection_rule_id }}
    required fields: id, package_protection_rule_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_packages_package_id: 
    endpoint: DELETE /projects/{{ record.id }}/packages/{{ record.package_id }}
    required fields: id, package_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_packages_package_id_package_files_package_file_id: 
    endpoint: DELETE /projects/{{ record.id }}/packages/{{ record.package_id }}/package_files/{{ record.package_file_id }}
    required fields: id, package_id, package_file_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_pages: 
    endpoint: DELETE /projects/{{ record.id }}/pages
    required fields: id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_pages_domains_domain: 
    endpoint: DELETE /projects/{{ record.id }}/pages/domains/{{ record.domain }}
    required fields: id, domain
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_pipeline_schedules_pipeline_schedule_id: 
    endpoint: DELETE /projects/{{ record.id }}/pipeline_schedules/{{ record.pipeline_schedule_id }}
    required fields: id, pipeline_schedule_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_pipeline_schedules_pipeline_schedule_id_variables_key: 
    endpoint: DELETE /projects/{{ record.id }}/pipeline_schedules/{{ record.pipeline_schedule_id }}/variables/{{ record.key }}
    required fields: id, pipeline_schedule_id, key
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_pipelines_pipeline_id: 
    endpoint: DELETE /projects/{{ record.id }}/pipelines/{{ record.pipeline_id }}
    required fields: id, pipeline_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_protected_branches_name: 
    endpoint: DELETE /projects/{{ record.id }}/protected_branches/{{ record.name }}
    required fields: id, name
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_protected_tags_name: 
    endpoint: DELETE /projects/{{ record.id }}/protected_tags/{{ record.name }}
    required fields: id, name
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_registry_protection_repository_rules_protection_rule_id: 
    endpoint: DELETE /projects/{{ record.id }}/registry/protection/repository/rules/{{ record.protection_rule_id }}
    required fields: id, protection_rule_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_registry_protection_tag_rules_protection_rule_id: 
    endpoint: DELETE /projects/{{ record.id }}/registry/protection/tag/rules/{{ record.protection_rule_id }}
    required fields: id, protection_rule_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_registry_repositories_repository_id: 
    endpoint: DELETE /projects/{{ record.id }}/registry/repositories/{{ record.repository_id }}
    required fields: id, repository_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_registry_repositories_repository_id_tags: 
    endpoint: DELETE /projects/{{ record.id }}/registry/repositories/{{ record.repository_id }}/tags
    required fields: id, repository_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_registry_repositories_repository_id_tags_tag_name: 
    endpoint: DELETE /projects/{{ record.id }}/registry/repositories/{{ record.repository_id }}/tags/{{ record.tag_name }}
    required fields: id, repository_id, tag_name
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_releases_tag_name: 
    endpoint: DELETE /projects/{{ record.id }}/releases/{{ record.tag_name }}
    required fields: id, tag_name
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_releases_tag_name_assets_links_link_id: 
    endpoint: DELETE /projects/{{ record.id }}/releases/{{ record.tag_name }}/assets/links/{{ record.link_id }}
    required fields: id, tag_name, link_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_remote_mirrors_mirror_id: 
    endpoint: DELETE /projects/{{ record.id }}/remote_mirrors/{{ record.mirror_id }}
    required fields: id, mirror_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_repository_branches_branch: 
    endpoint: DELETE /projects/{{ record.id }}/repository/branches/{{ record.branch }}
    required fields: id, branch
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_repository_files_file_path: 
    endpoint: DELETE /projects/{{ record.id }}/repository/files/{{ record.file_path }}
    required fields: id, file_path
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_repository_merged_branches: 
    endpoint: DELETE /projects/{{ record.id }}/repository/merged_branches
    required fields: id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_repository_tags_tag_name: 
    endpoint: DELETE /projects/{{ record.id }}/repository/tags/{{ record.tag_name }}
    required fields: id, tag_name
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_runners_runner_id: 
    endpoint: DELETE /projects/{{ record.id }}/runners/{{ record.runner_id }}
    required fields: id, runner_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_secure_files_secure_file_id: 
    endpoint: DELETE /projects/{{ record.id }}/secure_files/{{ record.secure_file_id }}
    required fields: id, secure_file_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_services_slug: 
    endpoint: DELETE /projects/{{ record.id }}/services/{{ record.slug }}
    required fields: id, slug
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_share_group_id: 
    endpoint: DELETE /projects/{{ record.id }}/share/{{ record.group_id }}
    required fields: id, group_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_snippets_snippet_id: 
    endpoint: DELETE /projects/{{ record.id }}/snippets/{{ record.snippet_id }}
    required fields: id, snippet_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_snippets_snippet_id_award_emoji_award_id: 
    endpoint: DELETE /projects/{{ record.id }}/snippets/{{ record.snippet_id }}/award_emoji/{{ record.award_id }}
    required fields: id, snippet_id, award_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_snippets_snippet_id_notes_note_id_award_emoji_award_id: 
    endpoint: DELETE /projects/{{ record.id }}/snippets/{{ record.snippet_id }}/notes/{{ record.note_id }}/award_emoji/{{ record.award_id }}
    required fields: id, snippet_id, note_id, award_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_terraform_state_name: 
    endpoint: DELETE /projects/{{ record.id }}/terraform/state/{{ record.name }}
    required fields: id, name
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_terraform_state_name_lock: 
    endpoint: DELETE /projects/{{ record.id }}/terraform/state/{{ record.name }}/lock
    required fields: id, name
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_terraform_state_name_versions_serial: 
    endpoint: DELETE /projects/{{ record.id }}/terraform/state/{{ record.name }}/versions/{{ record.serial }}
    required fields: id, name, serial
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_terraform_state_protection_rules_terraform_state_protection_rule_id: 
    endpoint: DELETE /projects/{{ record.id }}/terraform/state_protection_rules/{{ record.terraform_state_protection_rule_id }}
    required fields: id, terraform_state_protection_rule_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_triggers_trigger_id: 
    endpoint: DELETE /projects/{{ record.id }}/triggers/{{ record.trigger_id }}
    required fields: id, trigger_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_uploads_secret_filename: 
    endpoint: DELETE /projects/{{ record.id }}/uploads/{{ record.secret }}/{{ record.filename }}
    required fields: id, secret, filename
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_uploads_upload_id: 
    endpoint: DELETE /projects/{{ record.id }}/uploads/{{ record.upload_id }}
    required fields: id, upload_id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_variables_key: 
    endpoint: DELETE /projects/{{ record.id }}/variables/{{ record.key }}
    required fields: id, key
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_projects_id_wikis_slug: 
    endpoint: DELETE /projects/{{ record.id }}/wikis/{{ record.slug }}
    required fields: id, slug
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_runners: 
    endpoint: DELETE /runners
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_runners_managers: 
    endpoint: DELETE /runners/managers
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_runners_id: 
    endpoint: DELETE /runners/{{ record.id }}
    required fields: id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_snippets_id: 
    endpoint: DELETE /snippets/{{ record.id }}
    required fields: id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  delete_api_v4_topics_id: 
    endpoint: DELETE /topics/{{ record.id }}
    required fields: id
    risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation
  patch_api_v4_jobs_id_trace: 
    endpoint: PATCH /jobs/{{ record.id }}/trace
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  patch_api_v4_projects_id_error_tracking_settings: 
    endpoint: PATCH /projects/{{ record.id }}/error_tracking/settings
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  patch_api_v4_projects_id_job_token_scope: 
    endpoint: PATCH /projects/{{ record.id }}/job_token_scope
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  patch_api_v4_projects_id_packages_protection_rules_package_protection_rule_id: 
    endpoint: PATCH /projects/{{ record.id }}/packages/protection/rules/{{ record.package_protection_rule_id }}
    required fields: id, package_protection_rule_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  patch_api_v4_projects_id_pages: 
    endpoint: PATCH /projects/{{ record.id }}/pages
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  patch_api_v4_projects_id_protected_branches_name: 
    endpoint: PATCH /projects/{{ record.id }}/protected_branches/{{ record.name }}
    required fields: id, name
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  patch_api_v4_projects_id_registry_protection_repository_rules_protection_rule_id: 
    endpoint: PATCH /projects/{{ record.id }}/registry/protection/repository/rules/{{ record.protection_rule_id }}
    required fields: id, protection_rule_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  patch_api_v4_projects_id_registry_protection_tag_rules_protection_rule_id: 
    endpoint: PATCH /projects/{{ record.id }}/registry/protection/tag/rules/{{ record.protection_rule_id }}
    required fields: id, protection_rule_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  patch_api_v4_projects_id_terraform_state_protection_rules_terraform_state_protection_rule_id: 
    endpoint: PATCH /projects/{{ record.id }}/terraform/state_protection_rules/{{ record.terraform_state_protection_rule_id }}
    required fields: id, terraform_state_protection_rule_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_admin_ci_variables: 
    endpoint: POST /admin/ci/variables
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_admin_clusters_add: 
    endpoint: POST /admin/clusters/add
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_admin_migrations_timestamp_mark: 
    endpoint: POST /admin/migrations/{{ record.timestamp }}/mark
    required fields: timestamp
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_applications: 
    endpoint: POST /applications
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_applications_id_renew_secret: 
    endpoint: POST /applications/{{ record.id }}/renew-secret
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_broadcast_messages: 
    endpoint: POST /broadcast_messages
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_bulk_imports: 
    endpoint: POST /bulk_imports
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_bulk_imports_import_id_cancel: 
    endpoint: POST /bulk_imports/{{ record.import_id }}/cancel
    required fields: import_id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_container_registry_event_events: 
    endpoint: POST /container_registry_event/events
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_deploy_keys: 
    endpoint: POST /deploy_keys
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_feature_flags_unleash_project_id_client_metrics: 
    endpoint: POST /feature_flags/unleash/{{ record.project_id }}/client/metrics
    required fields: project_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_feature_flags_unleash_project_id_client_register: 
    endpoint: POST /feature_flags/unleash/{{ record.project_id }}/client/register
    required fields: project_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_features_name: 
    endpoint: POST /features/{{ record.name }}
    required fields: name
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_geo_node_proxy_id_graphql: 
    endpoint: POST /geo/node_proxy/{{ record.id }}/graphql
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_geo_proxy_git_ssh_info_refs_receive_pack: 
    endpoint: POST /geo/proxy_git_ssh/info_refs_receive_pack
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_geo_proxy_git_ssh_info_refs_upload_pack: 
    endpoint: POST /geo/proxy_git_ssh/info_refs_upload_pack
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_geo_proxy_git_ssh_receive_pack: 
    endpoint: POST /geo/proxy_git_ssh/receive_pack
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_geo_proxy_git_ssh_upload_pack: 
    endpoint: POST /geo/proxy_git_ssh/upload_pack
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_geo_status: 
    endpoint: POST /geo/status
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_glql: 
    endpoint: POST /glql
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups: 
    endpoint: POST /groups
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_import: 
    endpoint: POST /groups/import
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_groups_import_authorize: 
    endpoint: POST /groups/import/authorize
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_groups_id_debian_distributions: 
    endpoint: POST /groups/{{ record.id }}/-/debian_distributions
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_id_packages_npm_npm_v1_security_advisories_bulk: 
    endpoint: POST /groups/{{ record.id }}/-/packages/npm/-/npm/v1/security/advisories/bulk
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_id_packages_npm_npm_v1_security_audits_quick: 
    endpoint: POST /groups/{{ record.id }}/-/packages/npm/-/npm/v1/security/audits/quick
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_id_access_requests: 
    endpoint: POST /groups/{{ record.id }}/access_requests
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_groups_id_access_tokens_self_rotate: 
    endpoint: POST /groups/{{ record.id }}/access_tokens/self/rotate
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_groups_id_archive: 
    endpoint: POST /groups/{{ record.id }}/archive
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_id_badges: 
    endpoint: POST /groups/{{ record.id }}/badges
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_id_clusters_user: 
    endpoint: POST /groups/{{ record.id }}/clusters/user
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_id_deploy_tokens: 
    endpoint: POST /groups/{{ record.id }}/deploy_tokens
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_groups_id_epics_epic_iid_award_emoji: 
    endpoint: POST /groups/{{ record.id }}/epics/{{ record.epic_iid }}/award_emoji
    required fields: id, epic_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_id_epics_epic_iid_notes_note_id_award_emoji: 
    endpoint: POST /groups/{{ record.id }}/epics/{{ record.epic_iid }}/notes/{{ record.note_id }}/award_emoji
    required fields: id, epic_iid, note_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_id_export: 
    endpoint: POST /groups/{{ record.id }}/export
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_groups_id_export_relations: 
    endpoint: POST /groups/{{ record.id }}/export_relations
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_groups_id_invitations: 
    endpoint: POST /groups/{{ record.id }}/invitations
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_id_ldap_sync: 
    endpoint: POST /groups/{{ record.id }}/ldap_sync
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_groups_id_members: 
    endpoint: POST /groups/{{ record.id }}/members
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_groups_id_members_approve_all: 
    endpoint: POST /groups/{{ record.id }}/members/approve_all
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_groups_id_members_user_id_override: 
    endpoint: POST /groups/{{ record.id }}/members/{{ record.user_id }}/override
    required fields: id, user_id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_groups_id_placeholder_reassignments: 
    endpoint: POST /groups/{{ record.id }}/placeholder_reassignments
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_id_placeholder_reassignments_authorize: 
    endpoint: POST /groups/{{ record.id }}/placeholder_reassignments/authorize
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_id_projects_project_id: 
    endpoint: POST /groups/{{ record.id }}/projects/{{ record.project_id }}
    required fields: id, project_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_id_restore: 
    endpoint: POST /groups/{{ record.id }}/restore
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_id_runners_reset_registration_token: 
    endpoint: POST /groups/{{ record.id }}/runners/reset_registration_token
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_groups_id_share: 
    endpoint: POST /groups/{{ record.id }}/share
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_groups_id_ssh_certificates: 
    endpoint: POST /groups/{{ record.id }}/ssh_certificates
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_id_transfer: 
    endpoint: POST /groups/{{ record.id }}/transfer
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_groups_id_transfer_to_organization: 
    endpoint: POST /groups/{{ record.id }}/transfer_to_organization
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_groups_id_unarchive: 
    endpoint: POST /groups/{{ record.id }}/unarchive
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_id_uploads: 
    endpoint: POST /groups/{{ record.id }}/uploads
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_id_uploads_authorize: 
    endpoint: POST /groups/{{ record.id }}/uploads/authorize
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_id_variables: 
    endpoint: POST /groups/{{ record.id }}/variables
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_groups_id_wikis: 
    endpoint: POST /groups/{{ record.id }}/wikis
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_groups_id_wikis_attachments: 
    endpoint: POST /groups/{{ record.id }}/wikis/attachments
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_hooks: 
    endpoint: POST /hooks
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_hooks_hook_id: 
    endpoint: POST /hooks/{{ record.hook_id }}
    required fields: hook_id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_import_bitbucket: 
    endpoint: POST /import/bitbucket
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_import_bitbucket_server: 
    endpoint: POST /import/bitbucket_server
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_import_github: 
    endpoint: POST /import/github
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_import_github_cancel: 
    endpoint: POST /import/github/cancel
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_import_github_gists: 
    endpoint: POST /import/github/gists
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_integrations_jira_connect_subscriptions: 
    endpoint: POST /integrations/jira_connect/subscriptions
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_integrations_jira_forge_installation_forge_token: 
    endpoint: POST /integrations/jira_forge/installation/forge_token
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_integrations_jira_forge_subscriptions: 
    endpoint: POST /integrations/jira_forge/subscriptions
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_integrations_slack_events: 
    endpoint: POST /integrations/slack/events
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_integrations_slack_interactions: 
    endpoint: POST /integrations/slack/interactions
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_integrations_slack_options: 
    endpoint: POST /integrations/slack/options
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_jobs_request: 
    endpoint: POST /jobs/request
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_jobs_id_artifacts: 
    endpoint: POST /jobs/{{ record.id }}/artifacts
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_jobs_id_artifacts_authorize: 
    endpoint: POST /jobs/{{ record.id }}/artifacts/authorize
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_markdown: 
    endpoint: POST /markdown
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_namespaces_id_storage_limit_exclusion: 
    endpoint: POST /namespaces/{{ record.id }}/storage/limit_exclusion
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_offline_exports: 
    endpoint: POST /offline_exports
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_offline_imports: 
    endpoint: POST /offline_imports
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_organizations: 
    endpoint: POST /organizations
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_packages_conan_v1_conans_package_name_package_version_package_username_package_channel_packages_conan_package_reference_upload_urls: 
    endpoint: POST /packages/conan/v1/conans/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}/packages/{{ record.conan_package_reference }}/upload_urls
    required fields: package_name, package_version, package_username, package_channel, conan_package_reference
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_packages_conan_v1_conans_package_name_package_version_package_username_package_channel_upload_urls: 
    endpoint: POST /packages/conan/v1/conans/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}/upload_urls
    required fields: package_name, package_version, package_username, package_channel
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_packages_npm_npm_v1_security_advisories_bulk: 
    endpoint: POST /packages/npm/-/npm/v1/security/advisories/bulk
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_packages_npm_npm_v1_security_audits_quick: 
    endpoint: POST /packages/npm/-/npm/v1/security/audits/quick
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_personal_access_tokens_self_rotate: 
    endpoint: POST /personal_access_tokens/self/rotate
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_personal_access_tokens_id_rotate: 
    endpoint: POST /personal_access_tokens/{{ record.id }}/rotate
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects: 
    endpoint: POST /projects
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_import: 
    endpoint: POST /projects/import
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_projects_import_relation: 
    endpoint: POST /projects/import-relation
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_projects_import_relation_authorize: 
    endpoint: POST /projects/import-relation/authorize
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_projects_import_authorize: 
    endpoint: POST /projects/import/authorize
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_projects_remote_import: 
    endpoint: POST /projects/remote-import
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_projects_remote_import_s3: 
    endpoint: POST /projects/remote-import-s3
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_projects_user_user_id: 
    endpoint: POST /projects/user/{{ record.user_id }}
    required fields: user_id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_ref_ref_trigger_pipeline: 
    endpoint: POST /projects/{{ record.id }}/(ref/{{ record.ref }}/)trigger/pipeline
    required fields: id, ref
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_access_requests: 
    endpoint: POST /projects/{{ record.id }}/access_requests
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_access_tokens_self_rotate: 
    endpoint: POST /projects/{{ record.id }}/access_tokens/self/rotate
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_alert_management_alerts_alert_iid_metric_images: 
    endpoint: POST /projects/{{ record.id }}/alert_management_alerts/{{ record.alert_iid }}/metric_images
    required fields: id, alert_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_alert_management_alerts_alert_iid_metric_images_authorize: 
    endpoint: POST /projects/{{ record.id }}/alert_management_alerts/{{ record.alert_iid }}/metric_images/authorize
    required fields: id, alert_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_archive: 
    endpoint: POST /projects/{{ record.id }}/archive
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_badges: 
    endpoint: POST /projects/{{ record.id }}/badges
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_catalog_publish: 
    endpoint: POST /projects/{{ record.id }}/catalog/publish
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_ci_lint: 
    endpoint: POST /projects/{{ record.id }}/ci/lint
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_cluster_agents: 
    endpoint: POST /projects/{{ record.id }}/cluster_agents
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_cluster_agents_agent_id_tokens: 
    endpoint: POST /projects/{{ record.id }}/cluster_agents/{{ record.agent_id }}/tokens
    required fields: id, agent_id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_clusters_user: 
    endpoint: POST /projects/{{ record.id }}/clusters/user
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_create_ci_config: 
    endpoint: POST /projects/{{ record.id }}/create_ci_config
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_debian_distributions: 
    endpoint: POST /projects/{{ record.id }}/debian_distributions
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_deploy_keys: 
    endpoint: POST /projects/{{ record.id }}/deploy_keys
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_deploy_keys_key_id_enable: 
    endpoint: POST /projects/{{ record.id }}/deploy_keys/{{ record.key_id }}/enable
    required fields: id, key_id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_deploy_tokens: 
    endpoint: POST /projects/{{ record.id }}/deploy_tokens
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_deployments: 
    endpoint: POST /projects/{{ record.id }}/deployments
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_deployments_deployment_id_approval: 
    endpoint: POST /projects/{{ record.id }}/deployments/{{ record.deployment_id }}/approval
    required fields: id, deployment_id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_environments: 
    endpoint: POST /projects/{{ record.id }}/environments
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_environments_stop_stale: 
    endpoint: POST /projects/{{ record.id }}/environments/stop_stale
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_environments_environment_id_stop: 
    endpoint: POST /projects/{{ record.id }}/environments/{{ record.environment_id }}/stop
    required fields: id, environment_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_error_tracking_client_keys: 
    endpoint: POST /projects/{{ record.id }}/error_tracking/client_keys
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_export: 
    endpoint: POST /projects/{{ record.id }}/export
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_projects_id_export_relations: 
    endpoint: POST /projects/{{ record.id }}/export_relations
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_projects_id_feature_flags: 
    endpoint: POST /projects/{{ record.id }}/feature_flags
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_feature_flags_user_lists: 
    endpoint: POST /projects/{{ record.id }}/feature_flags_user_lists
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_fork: 
    endpoint: POST /projects/{{ record.id }}/fork
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_fork_forked_from_id: 
    endpoint: POST /projects/{{ record.id }}/fork/{{ record.forked_from_id }}
    required fields: id, forked_from_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_freeze_periods: 
    endpoint: POST /projects/{{ record.id }}/freeze_periods
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_hooks: 
    endpoint: POST /projects/{{ record.id }}/hooks
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_projects_id_hooks_hook_id_events_hook_log_id_resend: 
    endpoint: POST /projects/{{ record.id }}/hooks/{{ record.hook_id }}/events/{{ record.hook_log_id }}/resend
    required fields: id, hook_id, hook_log_id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_projects_id_hooks_hook_id_test_trigger: 
    endpoint: POST /projects/{{ record.id }}/hooks/{{ record.hook_id }}/test/{{ record.trigger }}
    required fields: id, hook_id, trigger
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_projects_id_housekeeping: 
    endpoint: POST /projects/{{ record.id }}/housekeeping
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_import_git: 
    endpoint: POST /projects/{{ record.id }}/import/git
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_projects_id_import_project_members_project_id: 
    endpoint: POST /projects/{{ record.id }}/import_project_members/{{ record.project_id }}
    required fields: id, project_id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_integrations_mattermost_slash_commands_trigger: 
    endpoint: POST /projects/{{ record.id }}/integrations/mattermost_slash_commands/trigger
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_invitations: 
    endpoint: POST /projects/{{ record.id }}/invitations
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_issues: 
    endpoint: POST /projects/{{ record.id }}/issues
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_issues_issue_iid_add_spent_time: 
    endpoint: POST /projects/{{ record.id }}/issues/{{ record.issue_iid }}/add_spent_time
    required fields: id, issue_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_issues_issue_iid_award_emoji: 
    endpoint: POST /projects/{{ record.id }}/issues/{{ record.issue_iid }}/award_emoji
    required fields: id, issue_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_issues_issue_iid_clone: 
    endpoint: POST /projects/{{ record.id }}/issues/{{ record.issue_iid }}/clone
    required fields: id, issue_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_issues_issue_iid_links: 
    endpoint: POST /projects/{{ record.id }}/issues/{{ record.issue_iid }}/links
    required fields: id, issue_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_issues_issue_iid_metric_images: 
    endpoint: POST /projects/{{ record.id }}/issues/{{ record.issue_iid }}/metric_images
    required fields: id, issue_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_issues_issue_iid_metric_images_authorize: 
    endpoint: POST /projects/{{ record.id }}/issues/{{ record.issue_iid }}/metric_images/authorize
    required fields: id, issue_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_issues_issue_iid_move: 
    endpoint: POST /projects/{{ record.id }}/issues/{{ record.issue_iid }}/move
    required fields: id, issue_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_issues_issue_iid_notes_note_id_award_emoji: 
    endpoint: POST /projects/{{ record.id }}/issues/{{ record.issue_iid }}/notes/{{ record.note_id }}/award_emoji
    required fields: id, issue_iid, note_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_issues_issue_iid_reset_spent_time: 
    endpoint: POST /projects/{{ record.id }}/issues/{{ record.issue_iid }}/reset_spent_time
    required fields: id, issue_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_issues_issue_iid_reset_time_estimate: 
    endpoint: POST /projects/{{ record.id }}/issues/{{ record.issue_iid }}/reset_time_estimate
    required fields: id, issue_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_issues_issue_iid_time_estimate: 
    endpoint: POST /projects/{{ record.id }}/issues/{{ record.issue_iid }}/time_estimate
    required fields: id, issue_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_job_token_scope_allowlist: 
    endpoint: POST /projects/{{ record.id }}/job_token_scope/allowlist
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_job_token_scope_groups_allowlist: 
    endpoint: POST /projects/{{ record.id }}/job_token_scope/groups_allowlist
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_jobs_job_id_artifacts_keep: 
    endpoint: POST /projects/{{ record.id }}/jobs/{{ record.job_id }}/artifacts/keep
    required fields: id, job_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_jobs_job_id_cancel: 
    endpoint: POST /projects/{{ record.id }}/jobs/{{ record.job_id }}/cancel
    required fields: id, job_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_jobs_job_id_erase: 
    endpoint: POST /projects/{{ record.id }}/jobs/{{ record.job_id }}/erase
    required fields: id, job_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_jobs_job_id_play: 
    endpoint: POST /projects/{{ record.id }}/jobs/{{ record.job_id }}/play
    required fields: id, job_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_jobs_job_id_retry: 
    endpoint: POST /projects/{{ record.id }}/jobs/{{ record.job_id }}/retry
    required fields: id, job_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_members: 
    endpoint: POST /projects/{{ record.id }}/members
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_merge_requests: 
    endpoint: POST /projects/{{ record.id }}/merge_requests
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_merge_requests_merge_request_iid_add_spent_time: 
    endpoint: POST /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/add_spent_time
    required fields: id, merge_request_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_merge_requests_merge_request_iid_approve: 
    endpoint: POST /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/approve
    required fields: id, merge_request_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_merge_requests_merge_request_iid_award_emoji: 
    endpoint: POST /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/award_emoji
    required fields: id, merge_request_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_merge_requests_merge_request_iid_cancel_merge_when_pipeline_succeeds: 
    endpoint: POST /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/cancel_merge_when_pipeline_succeeds
    required fields: id, merge_request_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_merge_requests_merge_request_iid_context_commits: 
    endpoint: POST /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/context_commits
    required fields: id, merge_request_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_merge_requests_merge_request_iid_draft_notes: 
    endpoint: POST /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/draft_notes
    required fields: id, merge_request_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_merge_requests_merge_request_iid_draft_notes_bulk_publish: 
    endpoint: POST /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/draft_notes/bulk_publish
    required fields: id, merge_request_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_merge_requests_merge_request_iid_notes_note_id_award_emoji: 
    endpoint: POST /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/notes/{{ record.note_id }}/award_emoji
    required fields: id, merge_request_iid, note_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_merge_requests_merge_request_iid_pipelines: 
    endpoint: POST /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/pipelines
    required fields: id, merge_request_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_merge_requests_merge_request_iid_reset_spent_time: 
    endpoint: POST /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/reset_spent_time
    required fields: id, merge_request_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_merge_requests_merge_request_iid_reset_time_estimate: 
    endpoint: POST /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/reset_time_estimate
    required fields: id, merge_request_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_merge_requests_merge_request_iid_time_estimate: 
    endpoint: POST /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/time_estimate
    required fields: id, merge_request_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_merge_requests_merge_request_iid_unapprove: 
    endpoint: POST /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/unapprove
    required fields: id, merge_request_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_packages_composer: 
    endpoint: POST /projects/{{ record.id }}/packages/composer
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_packages_conan_v1_conans_package_name_package_version_package_username_package_channel_packages_conan_package_reference_upload_urls: 
    endpoint: POST /projects/{{ record.id }}/packages/conan/v1/conans/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}/packages/{{ record.conan_package_reference }}/upload_urls
    required fields: id, package_name, package_version, package_username, package_channel, conan_package_reference
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_packages_conan_v1_conans_package_name_package_version_package_username_package_channel_upload_urls: 
    endpoint: POST /projects/{{ record.id }}/packages/conan/v1/conans/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}/upload_urls
    required fields: id, package_name, package_version, package_username, package_channel
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_packages_helm_api_channel_charts: 
    endpoint: POST /projects/{{ record.id }}/packages/helm/api/{{ record.channel }}/charts
    required fields: id, channel
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_packages_helm_api_channel_charts_authorize: 
    endpoint: POST /projects/{{ record.id }}/packages/helm/api/{{ record.channel }}/charts/authorize
    required fields: id, channel
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_packages_npm_npm_v1_security_advisories_bulk: 
    endpoint: POST /projects/{{ record.id }}/packages/npm/-/npm/v1/security/advisories/bulk
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_packages_npm_npm_v1_security_audits_quick: 
    endpoint: POST /projects/{{ record.id }}/packages/npm/-/npm/v1/security/audits/quick
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_packages_protection_rules: 
    endpoint: POST /projects/{{ record.id }}/packages/protection/rules
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_packages_pypi: 
    endpoint: POST /projects/{{ record.id }}/packages/pypi
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_packages_pypi_authorize: 
    endpoint: POST /projects/{{ record.id }}/packages/pypi/authorize
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_packages_rpm: 
    endpoint: POST /projects/{{ record.id }}/packages/rpm
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_packages_rpm_authorize: 
    endpoint: POST /projects/{{ record.id }}/packages/rpm/authorize
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_packages_rubygems_api_v1_gems: 
    endpoint: POST /projects/{{ record.id }}/packages/rubygems/api/v1/gems
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_packages_rubygems_api_v1_gems_authorize: 
    endpoint: POST /projects/{{ record.id }}/packages/rubygems/api/v1/gems/authorize
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_pages_domains: 
    endpoint: POST /projects/{{ record.id }}/pages/domains
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_pipeline: 
    endpoint: POST /projects/{{ record.id }}/pipeline
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_pipeline_schedules: 
    endpoint: POST /projects/{{ record.id }}/pipeline_schedules
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_pipeline_schedules_pipeline_schedule_id_play: 
    endpoint: POST /projects/{{ record.id }}/pipeline_schedules/{{ record.pipeline_schedule_id }}/play
    required fields: id, pipeline_schedule_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_pipeline_schedules_pipeline_schedule_id_take_ownership: 
    endpoint: POST /projects/{{ record.id }}/pipeline_schedules/{{ record.pipeline_schedule_id }}/take_ownership
    required fields: id, pipeline_schedule_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_pipeline_schedules_pipeline_schedule_id_variables: 
    endpoint: POST /projects/{{ record.id }}/pipeline_schedules/{{ record.pipeline_schedule_id }}/variables
    required fields: id, pipeline_schedule_id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_pipelines_pipeline_id_cancel: 
    endpoint: POST /projects/{{ record.id }}/pipelines/{{ record.pipeline_id }}/cancel
    required fields: id, pipeline_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_pipelines_pipeline_id_retry: 
    endpoint: POST /projects/{{ record.id }}/pipelines/{{ record.pipeline_id }}/retry
    required fields: id, pipeline_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_protected_branches: 
    endpoint: POST /projects/{{ record.id }}/protected_branches
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_protected_tags: 
    endpoint: POST /projects/{{ record.id }}/protected_tags
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_registry_protection_repository_rules: 
    endpoint: POST /projects/{{ record.id }}/registry/protection/repository/rules
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_registry_protection_tag_rules: 
    endpoint: POST /projects/{{ record.id }}/registry/protection/tag/rules
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_releases: 
    endpoint: POST /projects/{{ record.id }}/releases
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_releases_tag_name_assets_links: 
    endpoint: POST /projects/{{ record.id }}/releases/{{ record.tag_name }}/assets/links
    required fields: id, tag_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_releases_tag_name_evidence: 
    endpoint: POST /projects/{{ record.id }}/releases/{{ record.tag_name }}/evidence
    required fields: id, tag_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_remote_mirrors: 
    endpoint: POST /projects/{{ record.id }}/remote_mirrors
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_remote_mirrors_mirror_id_sync: 
    endpoint: POST /projects/{{ record.id }}/remote_mirrors/{{ record.mirror_id }}/sync
    required fields: id, mirror_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_repository_branches: 
    endpoint: POST /projects/{{ record.id }}/repository/branches
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_repository_changelog: 
    endpoint: POST /projects/{{ record.id }}/repository/changelog
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_repository_commits: 
    endpoint: POST /projects/{{ record.id }}/repository/commits
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_repository_commits_sha_cherry_pick: 
    endpoint: POST /projects/{{ record.id }}/repository/commits/{{ record.sha }}/cherry_pick
    required fields: id, sha
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_repository_commits_sha_comments: 
    endpoint: POST /projects/{{ record.id }}/repository/commits/{{ record.sha }}/comments
    required fields: id, sha
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_repository_commits_sha_revert: 
    endpoint: POST /projects/{{ record.id }}/repository/commits/{{ record.sha }}/revert
    required fields: id, sha
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_repository_files_file_path: 
    endpoint: POST /projects/{{ record.id }}/repository/files/{{ record.file_path }}
    required fields: id, file_path
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_repository_tags: 
    endpoint: POST /projects/{{ record.id }}/repository/tags
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_repository_size: 
    endpoint: POST /projects/{{ record.id }}/repository_size
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_restore: 
    endpoint: POST /projects/{{ record.id }}/restore
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_runners: 
    endpoint: POST /projects/{{ record.id }}/runners
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_projects_id_runners_reset_registration_token: 
    endpoint: POST /projects/{{ record.id }}/runners/reset_registration_token
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_secure_files: 
    endpoint: POST /projects/{{ record.id }}/secure_files
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_services_mattermost_slash_commands_trigger: 
    endpoint: POST /projects/{{ record.id }}/services/mattermost_slash_commands/trigger
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_share: 
    endpoint: POST /projects/{{ record.id }}/share
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_snippets: 
    endpoint: POST /projects/{{ record.id }}/snippets
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_snippets_snippet_id_award_emoji: 
    endpoint: POST /projects/{{ record.id }}/snippets/{{ record.snippet_id }}/award_emoji
    required fields: id, snippet_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_snippets_snippet_id_notes_note_id_award_emoji: 
    endpoint: POST /projects/{{ record.id }}/snippets/{{ record.snippet_id }}/notes/{{ record.note_id }}/award_emoji
    required fields: id, snippet_id, note_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_star: 
    endpoint: POST /projects/{{ record.id }}/star
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_statuses_sha: 
    endpoint: POST /projects/{{ record.id }}/statuses/{{ record.sha }}
    required fields: id, sha
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_terraform_state_name: 
    endpoint: POST /projects/{{ record.id }}/terraform/state/{{ record.name }}
    required fields: id, name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_terraform_state_name_authorize: 
    endpoint: POST /projects/{{ record.id }}/terraform/state/{{ record.name }}/authorize
    required fields: id, name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_terraform_state_name_lock: 
    endpoint: POST /projects/{{ record.id }}/terraform/state/{{ record.name }}/lock
    required fields: id, name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_terraform_state_protection_rules: 
    endpoint: POST /projects/{{ record.id }}/terraform/state_protection_rules
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_triggers: 
    endpoint: POST /projects/{{ record.id }}/triggers
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_unarchive: 
    endpoint: POST /projects/{{ record.id }}/unarchive
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_unstar: 
    endpoint: POST /projects/{{ record.id }}/unstar
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_uploads: 
    endpoint: POST /projects/{{ record.id }}/uploads
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_uploads_authorize: 
    endpoint: POST /projects/{{ record.id }}/uploads/authorize
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_variables: 
    endpoint: POST /projects/{{ record.id }}/variables
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_projects_id_wikis: 
    endpoint: POST /projects/{{ record.id }}/wikis
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_projects_id_wikis_attachments: 
    endpoint: POST /projects/{{ record.id }}/wikis/attachments
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_runners: 
    endpoint: POST /runners
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_runners_reset_authentication_token: 
    endpoint: POST /runners/reset_authentication_token
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_runners_reset_registration_token: 
    endpoint: POST /runners/reset_registration_token
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_runners_verify: 
    endpoint: POST /runners/verify
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_runners_id_reset_authentication_token: 
    endpoint: POST /runners/{{ record.id }}/reset_authentication_token
    required fields: id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  post_api_v4_slack_trigger: 
    endpoint: POST /slack/trigger
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_snippets: 
    endpoint: POST /snippets
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_topics: 
    endpoint: POST /topics
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_topics_merge: 
    endpoint: POST /topics/merge
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  post_api_v4_usage_data_increment_counter: 
    endpoint: POST /usage_data/increment_counter
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_usage_data_increment_unique_users: 
    endpoint: POST /usage_data/increment_unique_users
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_usage_data_track_event: 
    endpoint: POST /usage_data/track_event
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_usage_data_track_events: 
    endpoint: POST /usage_data/track_events
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  post_api_v4_user_runners: 
    endpoint: POST /user/runners
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  put_api_v4_admin_batched_background_migrations_id_pause: 
    endpoint: PUT /admin/batched_background_migrations/{{ record.id }}/pause
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  put_api_v4_admin_batched_background_migrations_id_resume: 
    endpoint: PUT /admin/batched_background_migrations/{{ record.id }}/resume
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  put_api_v4_admin_batched_background_operations_id_restart: 
    endpoint: PUT /admin/batched_background_operations/{{ record.id }}/restart
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  put_api_v4_admin_batched_background_operations_id_stop: 
    endpoint: PUT /admin/batched_background_operations/{{ record.id }}/stop
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  put_api_v4_admin_ci_variables_key: 
    endpoint: PUT /admin/ci/variables/{{ record.key }}
    required fields: key
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_admin_clusters_cluster_id: 
    endpoint: PUT /admin/clusters/{{ record.cluster_id }}
    required fields: cluster_id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  put_api_v4_application_appearance: 
    endpoint: PUT /application/appearance
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_application_plan_limits: 
    endpoint: PUT /application/plan_limits
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_broadcast_messages_id: 
    endpoint: PUT /broadcast_messages/{{ record.id }}
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  put_api_v4_groups_id: 
    endpoint: PUT /groups/{{ record.id }}
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_debian_distributions_codename: 
    endpoint: PUT /groups/{{ record.id }}/-/debian_distributions/{{ record.codename }}
    required fields: id, codename
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_packages_npm_package_package_name_dist_tags_tag: 
    endpoint: PUT /groups/{{ record.id }}/-/packages/npm/-/package/*package_name/dist-tags/{{ record.tag }}
    required fields: id, tag
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_access_requests_user_id_approve: 
    endpoint: PUT /groups/{{ record.id }}/access_requests/{{ record.user_id }}/approve
    required fields: id, user_id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_groups_id_badges_badge_id: 
    endpoint: PUT /groups/{{ record.id }}/badges/{{ record.badge_id }}
    required fields: id, badge_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_clusters_cluster_id: 
    endpoint: PUT /groups/{{ record.id }}/clusters/{{ record.cluster_id }}
    required fields: id, cluster_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_custom_attributes_key: 
    endpoint: PUT /groups/{{ record.id }}/custom_attributes/{{ record.key }}
    required fields: id, key
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_groups_id_integrations_apple_app_store: 
    endpoint: PUT /groups/{{ record.id }}/integrations/apple-app-store
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_asana: 
    endpoint: PUT /groups/{{ record.id }}/integrations/asana
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_assembla: 
    endpoint: PUT /groups/{{ record.id }}/integrations/assembla
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_bamboo: 
    endpoint: PUT /groups/{{ record.id }}/integrations/bamboo
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_bugzilla: 
    endpoint: PUT /groups/{{ record.id }}/integrations/bugzilla
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_buildkite: 
    endpoint: PUT /groups/{{ record.id }}/integrations/buildkite
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_campfire: 
    endpoint: PUT /groups/{{ record.id }}/integrations/campfire
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_clickup: 
    endpoint: PUT /groups/{{ record.id }}/integrations/clickup
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_confluence: 
    endpoint: PUT /groups/{{ record.id }}/integrations/confluence
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_custom_issue_tracker: 
    endpoint: PUT /groups/{{ record.id }}/integrations/custom-issue-tracker
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_datadog: 
    endpoint: PUT /groups/{{ record.id }}/integrations/datadog
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_diffblue_cover: 
    endpoint: PUT /groups/{{ record.id }}/integrations/diffblue-cover
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_discord: 
    endpoint: PUT /groups/{{ record.id }}/integrations/discord
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_drone_ci: 
    endpoint: PUT /groups/{{ record.id }}/integrations/drone-ci
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_emails_on_push: 
    endpoint: PUT /groups/{{ record.id }}/integrations/emails-on-push
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_ewm: 
    endpoint: PUT /groups/{{ record.id }}/integrations/ewm
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_external_wiki: 
    endpoint: PUT /groups/{{ record.id }}/integrations/external-wiki
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_git_guardian: 
    endpoint: PUT /groups/{{ record.id }}/integrations/git-guardian
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_github: 
    endpoint: PUT /groups/{{ record.id }}/integrations/github
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_gitlab_slack_application: 
    endpoint: PUT /groups/{{ record.id }}/integrations/gitlab-slack-application
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_google_cloud_platform_artifact_registry: 
    endpoint: PUT /groups/{{ record.id }}/integrations/google-cloud-platform-artifact-registry
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_google_cloud_platform_workload_identity_federation: 
    endpoint: PUT /groups/{{ record.id }}/integrations/google-cloud-platform-workload-identity-federation
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_google_play: 
    endpoint: PUT /groups/{{ record.id }}/integrations/google-play
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_hangouts_chat: 
    endpoint: PUT /groups/{{ record.id }}/integrations/hangouts-chat
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_harbor: 
    endpoint: PUT /groups/{{ record.id }}/integrations/harbor
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_irker: 
    endpoint: PUT /groups/{{ record.id }}/integrations/irker
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_jenkins: 
    endpoint: PUT /groups/{{ record.id }}/integrations/jenkins
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_jira: 
    endpoint: PUT /groups/{{ record.id }}/integrations/jira
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_jira_cloud_app: 
    endpoint: PUT /groups/{{ record.id }}/integrations/jira-cloud-app
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_linear: 
    endpoint: PUT /groups/{{ record.id }}/integrations/linear
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_matrix: 
    endpoint: PUT /groups/{{ record.id }}/integrations/matrix
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_mattermost: 
    endpoint: PUT /groups/{{ record.id }}/integrations/mattermost
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_mattermost_slash_commands: 
    endpoint: PUT /groups/{{ record.id }}/integrations/mattermost-slash-commands
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_microsoft_teams: 
    endpoint: PUT /groups/{{ record.id }}/integrations/microsoft-teams
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_mock_ci: 
    endpoint: PUT /groups/{{ record.id }}/integrations/mock-ci
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_mock_monitoring: 
    endpoint: PUT /groups/{{ record.id }}/integrations/mock-monitoring
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_packagist: 
    endpoint: PUT /groups/{{ record.id }}/integrations/packagist
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_phorge: 
    endpoint: PUT /groups/{{ record.id }}/integrations/phorge
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_pipelines_email: 
    endpoint: PUT /groups/{{ record.id }}/integrations/pipelines-email
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_pivotaltracker: 
    endpoint: PUT /groups/{{ record.id }}/integrations/pivotaltracker
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_pumble: 
    endpoint: PUT /groups/{{ record.id }}/integrations/pumble
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_pushover: 
    endpoint: PUT /groups/{{ record.id }}/integrations/pushover
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_redmine: 
    endpoint: PUT /groups/{{ record.id }}/integrations/redmine
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_slack: 
    endpoint: PUT /groups/{{ record.id }}/integrations/slack
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_squash_tm: 
    endpoint: PUT /groups/{{ record.id }}/integrations/squash-tm
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_teamcity: 
    endpoint: PUT /groups/{{ record.id }}/integrations/teamcity
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_telegram: 
    endpoint: PUT /groups/{{ record.id }}/integrations/telegram
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_unify_circuit: 
    endpoint: PUT /groups/{{ record.id }}/integrations/unify-circuit
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_webex_teams: 
    endpoint: PUT /groups/{{ record.id }}/integrations/webex-teams
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_youtrack: 
    endpoint: PUT /groups/{{ record.id }}/integrations/youtrack
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_integrations_zentao: 
    endpoint: PUT /groups/{{ record.id }}/integrations/zentao
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_invitations_email: 
    endpoint: PUT /groups/{{ record.id }}/invitations/{{ record.email }}
    required fields: id, email
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_groups_id_members_member_id_approve: 
    endpoint: PUT /groups/{{ record.id }}/members/{{ record.member_id }}/approve
    required fields: id, member_id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_groups_id_members_user_id: 
    endpoint: PUT /groups/{{ record.id }}/members/{{ record.user_id }}
    required fields: id, user_id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_groups_id_members_user_id_state: 
    endpoint: PUT /groups/{{ record.id }}/members/{{ record.user_id }}/state
    required fields: id, user_id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_groups_id_variables_key: 
    endpoint: PUT /groups/{{ record.id }}/variables/{{ record.key }}
    required fields: id, key
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_groups_id_wikis_slug: 
    endpoint: PUT /groups/{{ record.id }}/wikis/{{ record.slug }}
    required fields: id, slug
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_hooks_hook_id: 
    endpoint: PUT /hooks/{{ record.hook_id }}
    required fields: hook_id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  put_api_v4_hooks_hook_id_custom_headers_key: 
    endpoint: PUT /hooks/{{ record.hook_id }}/custom_headers/{{ record.key }}
    required fields: hook_id, key
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_hooks_hook_id_url_variables_key: 
    endpoint: PUT /hooks/{{ record.hook_id }}/url_variables/{{ record.key }}
    required fields: hook_id, key
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_integrations_jira_forge_installation: 
    endpoint: PUT /integrations/jira_forge/installation
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_jobs_id: 
    endpoint: PUT /jobs/{{ record.id }}
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_namespaces_id: 
    endpoint: PUT /namespaces/{{ record.id }}
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  put_api_v4_packages_conan_v1_files_package_name_package_version_package_username_package_channel_recipe_revision_export_file_name: 
    endpoint: PUT /packages/conan/v1/files/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}/{{ record.recipe_revision }}/export/{{ record.file_name }}
    required fields: package_name, package_version, package_username, package_channel, recipe_revision, file_name
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  put_api_v4_packages_conan_v1_files_package_name_package_version_package_username_package_channel_recipe_revision_export_file_name_authorize: 
    endpoint: PUT /packages/conan/v1/files/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}/{{ record.recipe_revision }}/export/{{ record.file_name }}/authorize
    required fields: package_name, package_version, package_username, package_channel, recipe_revision, file_name
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  put_api_v4_packages_conan_v1_files_package_name_package_version_package_username_package_channel_recipe_revision_package_conan_package_reference_package_revision_file_name: 
    endpoint: PUT /packages/conan/v1/files/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}/{{ record.recipe_revision }}/package/{{ record.conan_package_reference }}/{{ record.package_revision }}/{{ record.file_name }}
    required fields: package_name, package_version, package_username, package_channel, recipe_revision, conan_package_reference, package_revision, file_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_packages_conan_v1_files_package_name_package_version_package_username_package_channel_recipe_revision_package_conan_package_reference_package_revision_file_name_authorize: 
    endpoint: PUT /packages/conan/v1/files/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}/{{ record.recipe_revision }}/package/{{ record.conan_package_reference }}/{{ record.package_revision }}/{{ record.file_name }}/authorize
    required fields: package_name, package_version, package_username, package_channel, recipe_revision, conan_package_reference, package_revision, file_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_packages_npm_package_package_name_dist_tags_tag: 
    endpoint: PUT /packages/npm/-/package/*package_name/dist-tags/{{ record.tag }}
    required fields: tag
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id: 
    endpoint: PUT /projects/{{ record.id }}
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_access_requests_user_id_approve: 
    endpoint: PUT /projects/{{ record.id }}/access_requests/{{ record.user_id }}/approve
    required fields: id, user_id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_projects_id_alert_management_alerts_alert_iid_metric_images_metric_image_id: 
    endpoint: PUT /projects/{{ record.id }}/alert_management_alerts/{{ record.alert_iid }}/metric_images/{{ record.metric_image_id }}
    required fields: id, alert_iid, metric_image_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_badges_badge_id: 
    endpoint: PUT /projects/{{ record.id }}/badges/{{ record.badge_id }}
    required fields: id, badge_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_clusters_cluster_id: 
    endpoint: PUT /projects/{{ record.id }}/clusters/{{ record.cluster_id }}
    required fields: id, cluster_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_custom_attributes_key: 
    endpoint: PUT /projects/{{ record.id }}/custom_attributes/{{ record.key }}
    required fields: id, key
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_projects_id_debian_distributions_codename: 
    endpoint: PUT /projects/{{ record.id }}/debian_distributions/{{ record.codename }}
    required fields: id, codename
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_deploy_keys_key_id: 
    endpoint: PUT /projects/{{ record.id }}/deploy_keys/{{ record.key_id }}
    required fields: id, key_id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_projects_id_deployments_deployment_id: 
    endpoint: PUT /projects/{{ record.id }}/deployments/{{ record.deployment_id }}
    required fields: id, deployment_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_environments_environment_id: 
    endpoint: PUT /projects/{{ record.id }}/environments/{{ record.environment_id }}
    required fields: id, environment_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_error_tracking_settings: 
    endpoint: PUT /projects/{{ record.id }}/error_tracking/settings
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_feature_flags_feature_flag_name: 
    endpoint: PUT /projects/{{ record.id }}/feature_flags/{{ record.feature_flag_name }}
    required fields: id, feature_flag_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_feature_flags_user_lists_iid: 
    endpoint: PUT /projects/{{ record.id }}/feature_flags_user_lists/{{ record.iid }}
    required fields: id, iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_freeze_periods_freeze_period_id: 
    endpoint: PUT /projects/{{ record.id }}/freeze_periods/{{ record.freeze_period_id }}
    required fields: id, freeze_period_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_hooks_hook_id: 
    endpoint: PUT /projects/{{ record.id }}/hooks/{{ record.hook_id }}
    required fields: id, hook_id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  put_api_v4_projects_id_hooks_hook_id_custom_headers_key: 
    endpoint: PUT /projects/{{ record.id }}/hooks/{{ record.hook_id }}/custom_headers/{{ record.key }}
    required fields: id, hook_id, key
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_projects_id_hooks_hook_id_url_variables_key: 
    endpoint: PUT /projects/{{ record.id }}/hooks/{{ record.hook_id }}/url_variables/{{ record.key }}
    required fields: id, hook_id, key
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_projects_id_integrations_apple_app_store: 
    endpoint: PUT /projects/{{ record.id }}/integrations/apple-app-store
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_asana: 
    endpoint: PUT /projects/{{ record.id }}/integrations/asana
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_assembla: 
    endpoint: PUT /projects/{{ record.id }}/integrations/assembla
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_bamboo: 
    endpoint: PUT /projects/{{ record.id }}/integrations/bamboo
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_bugzilla: 
    endpoint: PUT /projects/{{ record.id }}/integrations/bugzilla
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_buildkite: 
    endpoint: PUT /projects/{{ record.id }}/integrations/buildkite
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_campfire: 
    endpoint: PUT /projects/{{ record.id }}/integrations/campfire
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_clickup: 
    endpoint: PUT /projects/{{ record.id }}/integrations/clickup
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_confluence: 
    endpoint: PUT /projects/{{ record.id }}/integrations/confluence
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_custom_issue_tracker: 
    endpoint: PUT /projects/{{ record.id }}/integrations/custom-issue-tracker
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_datadog: 
    endpoint: PUT /projects/{{ record.id }}/integrations/datadog
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_diffblue_cover: 
    endpoint: PUT /projects/{{ record.id }}/integrations/diffblue-cover
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_discord: 
    endpoint: PUT /projects/{{ record.id }}/integrations/discord
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_drone_ci: 
    endpoint: PUT /projects/{{ record.id }}/integrations/drone-ci
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_emails_on_push: 
    endpoint: PUT /projects/{{ record.id }}/integrations/emails-on-push
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_ewm: 
    endpoint: PUT /projects/{{ record.id }}/integrations/ewm
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_external_wiki: 
    endpoint: PUT /projects/{{ record.id }}/integrations/external-wiki
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_git_guardian: 
    endpoint: PUT /projects/{{ record.id }}/integrations/git-guardian
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_github: 
    endpoint: PUT /projects/{{ record.id }}/integrations/github
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_gitlab_slack_application: 
    endpoint: PUT /projects/{{ record.id }}/integrations/gitlab-slack-application
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_google_cloud_platform_artifact_registry: 
    endpoint: PUT /projects/{{ record.id }}/integrations/google-cloud-platform-artifact-registry
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_google_cloud_platform_workload_identity_federation: 
    endpoint: PUT /projects/{{ record.id }}/integrations/google-cloud-platform-workload-identity-federation
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_google_play: 
    endpoint: PUT /projects/{{ record.id }}/integrations/google-play
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_hangouts_chat: 
    endpoint: PUT /projects/{{ record.id }}/integrations/hangouts-chat
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_harbor: 
    endpoint: PUT /projects/{{ record.id }}/integrations/harbor
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_irker: 
    endpoint: PUT /projects/{{ record.id }}/integrations/irker
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_jenkins: 
    endpoint: PUT /projects/{{ record.id }}/integrations/jenkins
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_jira: 
    endpoint: PUT /projects/{{ record.id }}/integrations/jira
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_jira_cloud_app: 
    endpoint: PUT /projects/{{ record.id }}/integrations/jira-cloud-app
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_linear: 
    endpoint: PUT /projects/{{ record.id }}/integrations/linear
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_matrix: 
    endpoint: PUT /projects/{{ record.id }}/integrations/matrix
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_mattermost: 
    endpoint: PUT /projects/{{ record.id }}/integrations/mattermost
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_mattermost_slash_commands: 
    endpoint: PUT /projects/{{ record.id }}/integrations/mattermost-slash-commands
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_microsoft_teams: 
    endpoint: PUT /projects/{{ record.id }}/integrations/microsoft-teams
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_mock_ci: 
    endpoint: PUT /projects/{{ record.id }}/integrations/mock-ci
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_mock_monitoring: 
    endpoint: PUT /projects/{{ record.id }}/integrations/mock-monitoring
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_packagist: 
    endpoint: PUT /projects/{{ record.id }}/integrations/packagist
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_phorge: 
    endpoint: PUT /projects/{{ record.id }}/integrations/phorge
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_pipelines_email: 
    endpoint: PUT /projects/{{ record.id }}/integrations/pipelines-email
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_pivotaltracker: 
    endpoint: PUT /projects/{{ record.id }}/integrations/pivotaltracker
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_pumble: 
    endpoint: PUT /projects/{{ record.id }}/integrations/pumble
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_pushover: 
    endpoint: PUT /projects/{{ record.id }}/integrations/pushover
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_redmine: 
    endpoint: PUT /projects/{{ record.id }}/integrations/redmine
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_slack: 
    endpoint: PUT /projects/{{ record.id }}/integrations/slack
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_squash_tm: 
    endpoint: PUT /projects/{{ record.id }}/integrations/squash-tm
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_teamcity: 
    endpoint: PUT /projects/{{ record.id }}/integrations/teamcity
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_telegram: 
    endpoint: PUT /projects/{{ record.id }}/integrations/telegram
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_unify_circuit: 
    endpoint: PUT /projects/{{ record.id }}/integrations/unify-circuit
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_webex_teams: 
    endpoint: PUT /projects/{{ record.id }}/integrations/webex-teams
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_youtrack: 
    endpoint: PUT /projects/{{ record.id }}/integrations/youtrack
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_integrations_zentao: 
    endpoint: PUT /projects/{{ record.id }}/integrations/zentao
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_invitations_email: 
    endpoint: PUT /projects/{{ record.id }}/invitations/{{ record.email }}
    required fields: id, email
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_issues_issue_iid: 
    endpoint: PUT /projects/{{ record.id }}/issues/{{ record.issue_iid }}
    required fields: id, issue_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_issues_issue_iid_metric_images_metric_image_id: 
    endpoint: PUT /projects/{{ record.id }}/issues/{{ record.issue_iid }}/metric_images/{{ record.metric_image_id }}
    required fields: id, issue_iid, metric_image_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_issues_issue_iid_reorder: 
    endpoint: PUT /projects/{{ record.id }}/issues/{{ record.issue_iid }}/reorder
    required fields: id, issue_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_members_user_id: 
    endpoint: PUT /projects/{{ record.id }}/members/{{ record.user_id }}
    required fields: id, user_id
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_projects_id_merge_requests_merge_request_iid: 
    endpoint: PUT /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}
    required fields: id, merge_request_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_merge_requests_merge_request_iid_draft_notes_draft_note_id: 
    endpoint: PUT /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/draft_notes/{{ record.draft_note_id }}
    required fields: id, merge_request_iid, draft_note_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_merge_requests_merge_request_iid_draft_notes_draft_note_id_publish: 
    endpoint: PUT /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/draft_notes/{{ record.draft_note_id }}/publish
    required fields: id, merge_request_iid, draft_note_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_merge_requests_merge_request_iid_merge: 
    endpoint: PUT /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/merge
    required fields: id, merge_request_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_merge_requests_merge_request_iid_rebase: 
    endpoint: PUT /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/rebase
    required fields: id, merge_request_iid
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_merge_requests_merge_request_iid_reset_approvals: 
    endpoint: PUT /projects/{{ record.id }}/merge_requests/{{ record.merge_request_iid }}/reset_approvals
    required fields: id, merge_request_iid
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_projects_id_packages_conan_v1_files_package_name_package_version_package_username_package_channel_recipe_revision_export_file_name: 
    endpoint: PUT /projects/{{ record.id }}/packages/conan/v1/files/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}/{{ record.recipe_revision }}/export/{{ record.file_name }}
    required fields: id, package_name, package_version, package_username, package_channel, recipe_revision, file_name
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  put_api_v4_projects_id_packages_conan_v1_files_package_name_package_version_package_username_package_channel_recipe_revision_export_file_name_authorize: 
    endpoint: PUT /projects/{{ record.id }}/packages/conan/v1/files/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}/{{ record.recipe_revision }}/export/{{ record.file_name }}/authorize
    required fields: id, package_name, package_version, package_username, package_channel, recipe_revision, file_name
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  put_api_v4_projects_id_packages_conan_v1_files_package_name_package_version_package_username_package_channel_recipe_revision_package_conan_package_reference_package_revision_file_name: 
    endpoint: PUT /projects/{{ record.id }}/packages/conan/v1/files/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}/{{ record.recipe_revision }}/package/{{ record.conan_package_reference }}/{{ record.package_revision }}/{{ record.file_name }}
    required fields: id, package_name, package_version, package_username, package_channel, recipe_revision, conan_package_reference, package_revision, file_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_conan_v1_files_package_name_package_version_package_username_package_channel_recipe_revision_package_conan_package_reference_package_revision_file_name_authorize: 
    endpoint: PUT /projects/{{ record.id }}/packages/conan/v1/files/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}/{{ record.recipe_revision }}/package/{{ record.conan_package_reference }}/{{ record.package_revision }}/{{ record.file_name }}/authorize
    required fields: id, package_name, package_version, package_username, package_channel, recipe_revision, conan_package_reference, package_revision, file_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_conan_v2_conans_package_name_package_version_package_username_package_channel_revisions_recipe_revision_files_file_name: 
    endpoint: PUT /projects/{{ record.id }}/packages/conan/v2/conans/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}/revisions/{{ record.recipe_revision }}/files/{{ record.file_name }}
    required fields: id, package_name, package_version, package_username, package_channel, recipe_revision, file_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_conan_v2_conans_package_name_package_version_package_username_package_channel_revisions_recipe_revision_files_file_name_authorize: 
    endpoint: PUT /projects/{{ record.id }}/packages/conan/v2/conans/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}/revisions/{{ record.recipe_revision }}/files/{{ record.file_name }}/authorize
    required fields: id, package_name, package_version, package_username, package_channel, recipe_revision, file_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_conan_v2_conans_package_name_package_version_package_username_package_channel_revisions_recipe_revision_packages_conan_package_reference_revisions_package_revision_files_file_name: 
    endpoint: PUT /projects/{{ record.id }}/packages/conan/v2/conans/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}/revisions/{{ record.recipe_revision }}/packages/{{ record.conan_package_reference }}/revisions/{{ record.package_revision }}/files/{{ record.file_name }}
    required fields: id, package_name, package_version, package_username, package_channel, recipe_revision, conan_package_reference, package_revision, file_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_conan_v2_conans_package_name_package_version_package_username_package_channel_revisions_recipe_revision_packages_conan_package_reference_revisions_package_revision_files_file_name_authorize: 
    endpoint: PUT /projects/{{ record.id }}/packages/conan/v2/conans/{{ record.package_name }}/{{ record.package_version }}/{{ record.package_username }}/{{ record.package_channel }}/revisions/{{ record.recipe_revision }}/packages/{{ record.conan_package_reference }}/revisions/{{ record.package_revision }}/files/{{ record.file_name }}/authorize
    required fields: id, package_name, package_version, package_username, package_channel, recipe_revision, conan_package_reference, package_revision, file_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_debian_file_name: 
    endpoint: PUT /projects/{{ record.id }}/packages/debian/{{ record.file_name }}
    required fields: id, file_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_debian_file_name_authorize: 
    endpoint: PUT /projects/{{ record.id }}/packages/debian/{{ record.file_name }}/authorize
    required fields: id, file_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_generic_package_name_package_version_path__file_name: 
    endpoint: PUT /projects/{{ record.id }}/packages/generic/{{ record.package_name }}/*package_version/(*path/){{ record.file_name }}
    required fields: id, package_name, file_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_generic_package_name_package_version_path__file_name_authorize: 
    endpoint: PUT /projects/{{ record.id }}/packages/generic/{{ record.package_name }}/*package_version/(*path/){{ record.file_name }}/authorize
    required fields: id, package_name, file_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_maven_path_file_name: 
    endpoint: PUT /projects/{{ record.id }}/packages/maven/*path/{{ record.file_name }}
    required fields: id, file_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_maven_path_file_name_authorize: 
    endpoint: PUT /projects/{{ record.id }}/packages/maven/*path/{{ record.file_name }}/authorize
    required fields: id, file_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_ml_models_model_version_id_files_path__file_name: 
    endpoint: PUT /projects/{{ record.id }}/packages/ml_models/{{ record.model_version_id }}/files/(*path/){{ record.file_name }}
    required fields: id, model_version_id, file_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_ml_models_model_version_id_files_path__file_name_authorize: 
    endpoint: PUT /projects/{{ record.id }}/packages/ml_models/{{ record.model_version_id }}/files/(*path/){{ record.file_name }}/authorize
    required fields: id, model_version_id, file_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_npm_package_package_name_dist_tags_tag: 
    endpoint: PUT /projects/{{ record.id }}/packages/npm/-/package/*package_name/dist-tags/{{ record.tag }}
    required fields: id, tag
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_npm_package_name: 
    endpoint: PUT /projects/{{ record.id }}/packages/npm/{{ record.package_name }}
    required fields: id, package_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_nuget: 
    endpoint: PUT /projects/{{ record.id }}/packages/nuget
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_nuget_authorize: 
    endpoint: PUT /projects/{{ record.id }}/packages/nuget/authorize
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_nuget_symbolpackage: 
    endpoint: PUT /projects/{{ record.id }}/packages/nuget/symbolpackage
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_nuget_symbolpackage_authorize: 
    endpoint: PUT /projects/{{ record.id }}/packages/nuget/symbolpackage/authorize
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_nuget_v2: 
    endpoint: PUT /projects/{{ record.id }}/packages/nuget/v2
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_nuget_v2_authorize: 
    endpoint: PUT /projects/{{ record.id }}/packages/nuget/v2/authorize
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_terraform_modules_module_name_module_system_module_version_file: 
    endpoint: PUT /projects/{{ record.id }}/packages/terraform/modules/{{ record.module_name }}/{{ record.module_system }}/*module_version/file
    required fields: id, module_name, module_system
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_packages_terraform_modules_module_name_module_system_module_version_file_authorize: 
    endpoint: PUT /projects/{{ record.id }}/packages/terraform/modules/{{ record.module_name }}/{{ record.module_system }}/*module_version/file/authorize
    required fields: id, module_name, module_system
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_pages_domains_domain: 
    endpoint: PUT /projects/{{ record.id }}/pages/domains/{{ record.domain }}
    required fields: id, domain
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_pages_domains_domain_verify: 
    endpoint: PUT /projects/{{ record.id }}/pages/domains/{{ record.domain }}/verify
    required fields: id, domain
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_pipeline_schedules_pipeline_schedule_id: 
    endpoint: PUT /projects/{{ record.id }}/pipeline_schedules/{{ record.pipeline_schedule_id }}
    required fields: id, pipeline_schedule_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_pipeline_schedules_pipeline_schedule_id_variables_key: 
    endpoint: PUT /projects/{{ record.id }}/pipeline_schedules/{{ record.pipeline_schedule_id }}/variables/{{ record.key }}
    required fields: id, pipeline_schedule_id, key
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_projects_id_pipelines_pipeline_id_metadata: 
    endpoint: PUT /projects/{{ record.id }}/pipelines/{{ record.pipeline_id }}/metadata
    required fields: id, pipeline_id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  put_api_v4_projects_id_releases_tag_name: 
    endpoint: PUT /projects/{{ record.id }}/releases/{{ record.tag_name }}
    required fields: id, tag_name
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_releases_tag_name_assets_links_link_id: 
    endpoint: PUT /projects/{{ record.id }}/releases/{{ record.tag_name }}/assets/links/{{ record.link_id }}
    required fields: id, tag_name, link_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_remote_mirrors_mirror_id: 
    endpoint: PUT /projects/{{ record.id }}/remote_mirrors/{{ record.mirror_id }}
    required fields: id, mirror_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_repository_branches_branch_protect: 
    endpoint: PUT /projects/{{ record.id }}/repository/branches/{{ record.branch }}/protect
    required fields: id, branch
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_repository_branches_branch_unprotect: 
    endpoint: PUT /projects/{{ record.id }}/repository/branches/{{ record.branch }}/unprotect
    required fields: id, branch
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_repository_files_file_path: 
    endpoint: PUT /projects/{{ record.id }}/repository/files/{{ record.file_path }}
    required fields: id, file_path
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_repository_submodules_submodule: 
    endpoint: PUT /projects/{{ record.id }}/repository/submodules/{{ record.submodule }}
    required fields: id, submodule
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_resource_groups_key: 
    endpoint: PUT /projects/{{ record.id }}/resource_groups/{{ record.key }}
    required fields: id, key
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_projects_id_services_apple_app_store: 
    endpoint: PUT /projects/{{ record.id }}/services/apple-app-store
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_asana: 
    endpoint: PUT /projects/{{ record.id }}/services/asana
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_assembla: 
    endpoint: PUT /projects/{{ record.id }}/services/assembla
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_bamboo: 
    endpoint: PUT /projects/{{ record.id }}/services/bamboo
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_bugzilla: 
    endpoint: PUT /projects/{{ record.id }}/services/bugzilla
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_buildkite: 
    endpoint: PUT /projects/{{ record.id }}/services/buildkite
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_campfire: 
    endpoint: PUT /projects/{{ record.id }}/services/campfire
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_clickup: 
    endpoint: PUT /projects/{{ record.id }}/services/clickup
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_confluence: 
    endpoint: PUT /projects/{{ record.id }}/services/confluence
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_custom_issue_tracker: 
    endpoint: PUT /projects/{{ record.id }}/services/custom-issue-tracker
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_datadog: 
    endpoint: PUT /projects/{{ record.id }}/services/datadog
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_diffblue_cover: 
    endpoint: PUT /projects/{{ record.id }}/services/diffblue-cover
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_discord: 
    endpoint: PUT /projects/{{ record.id }}/services/discord
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_drone_ci: 
    endpoint: PUT /projects/{{ record.id }}/services/drone-ci
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_emails_on_push: 
    endpoint: PUT /projects/{{ record.id }}/services/emails-on-push
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_ewm: 
    endpoint: PUT /projects/{{ record.id }}/services/ewm
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_external_wiki: 
    endpoint: PUT /projects/{{ record.id }}/services/external-wiki
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_git_guardian: 
    endpoint: PUT /projects/{{ record.id }}/services/git-guardian
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_github: 
    endpoint: PUT /projects/{{ record.id }}/services/github
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_gitlab_slack_application: 
    endpoint: PUT /projects/{{ record.id }}/services/gitlab-slack-application
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_google_cloud_platform_artifact_registry: 
    endpoint: PUT /projects/{{ record.id }}/services/google-cloud-platform-artifact-registry
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_google_cloud_platform_workload_identity_federation: 
    endpoint: PUT /projects/{{ record.id }}/services/google-cloud-platform-workload-identity-federation
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_google_play: 
    endpoint: PUT /projects/{{ record.id }}/services/google-play
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_hangouts_chat: 
    endpoint: PUT /projects/{{ record.id }}/services/hangouts-chat
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_harbor: 
    endpoint: PUT /projects/{{ record.id }}/services/harbor
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_irker: 
    endpoint: PUT /projects/{{ record.id }}/services/irker
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_jenkins: 
    endpoint: PUT /projects/{{ record.id }}/services/jenkins
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_jira: 
    endpoint: PUT /projects/{{ record.id }}/services/jira
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_jira_cloud_app: 
    endpoint: PUT /projects/{{ record.id }}/services/jira-cloud-app
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_linear: 
    endpoint: PUT /projects/{{ record.id }}/services/linear
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_matrix: 
    endpoint: PUT /projects/{{ record.id }}/services/matrix
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_mattermost: 
    endpoint: PUT /projects/{{ record.id }}/services/mattermost
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_mattermost_slash_commands: 
    endpoint: PUT /projects/{{ record.id }}/services/mattermost-slash-commands
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_microsoft_teams: 
    endpoint: PUT /projects/{{ record.id }}/services/microsoft-teams
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_mock_ci: 
    endpoint: PUT /projects/{{ record.id }}/services/mock-ci
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_mock_monitoring: 
    endpoint: PUT /projects/{{ record.id }}/services/mock-monitoring
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_packagist: 
    endpoint: PUT /projects/{{ record.id }}/services/packagist
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_phorge: 
    endpoint: PUT /projects/{{ record.id }}/services/phorge
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_pipelines_email: 
    endpoint: PUT /projects/{{ record.id }}/services/pipelines-email
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_pivotaltracker: 
    endpoint: PUT /projects/{{ record.id }}/services/pivotaltracker
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_pumble: 
    endpoint: PUT /projects/{{ record.id }}/services/pumble
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_pushover: 
    endpoint: PUT /projects/{{ record.id }}/services/pushover
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_redmine: 
    endpoint: PUT /projects/{{ record.id }}/services/redmine
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_slack: 
    endpoint: PUT /projects/{{ record.id }}/services/slack
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_squash_tm: 
    endpoint: PUT /projects/{{ record.id }}/services/squash-tm
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_teamcity: 
    endpoint: PUT /projects/{{ record.id }}/services/teamcity
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_telegram: 
    endpoint: PUT /projects/{{ record.id }}/services/telegram
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_unify_circuit: 
    endpoint: PUT /projects/{{ record.id }}/services/unify-circuit
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_webex_teams: 
    endpoint: PUT /projects/{{ record.id }}/services/webex-teams
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_youtrack: 
    endpoint: PUT /projects/{{ record.id }}/services/youtrack
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_services_zentao: 
    endpoint: PUT /projects/{{ record.id }}/services/zentao
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_snippets_snippet_id: 
    endpoint: PUT /projects/{{ record.id }}/snippets/{{ record.snippet_id }}
    required fields: id, snippet_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_transfer: 
    endpoint: PUT /projects/{{ record.id }}/transfer
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  put_api_v4_projects_id_triggers_trigger_id: 
    endpoint: PUT /projects/{{ record.id }}/triggers/{{ record.trigger_id }}
    required fields: id, trigger_id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_projects_id_variables_key: 
    endpoint: PUT /projects/{{ record.id }}/variables/{{ record.key }}
    required fields: id, key
    risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy
  put_api_v4_projects_id_wikis_slug: 
    endpoint: PUT /projects/{{ record.id }}/wikis/{{ record.slug }}
    required fields: id, slug
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_runners_id: 
    endpoint: PUT /runners/{{ record.id }}
    required fields: id
    risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation
  put_api_v4_snippets_id: 
    endpoint: PUT /snippets/{{ record.id }}
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_suggestions_batch_apply: 
    endpoint: PUT /suggestions/batch_apply
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_suggestions_id_apply: 
    endpoint: PUT /suggestions/{{ record.id }}/apply
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared
  put_api_v4_topics_id: 
    endpoint: PUT /topics/{{ record.id }}
    required fields: id
    risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared

SECURITY
  read risk: external GitLab API read of projects, groups, users, and issues
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Work with GitLab projects from the command line.
  Usage: pm gitlab <command> <subcommand> [flags]
  Source CLI: glab (https://gitlab.com/gitlab-org/cli/-/tree/main/docs/source)
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --connection (string): Use a saved GitLab connector credential and base URL scope: maps_to=connection
    --limit (integer): Limit records emitted by stream-backed commands: maps_to=limit
  Core Commands
    project list - List visible GitLab projects [intent=etl availability=implemented stream=projects]; flags: --search, --owned
    project view - View one GitLab project [intent=direct_read availability=implemented]; notes: Bounded direct read of one project; response is recursively redacted before output.; flags: --id
    project delete - DELETE /projects/{id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project metric-images delete - DELETE /projects/{id}/alert_management_alerts/{alert_iid}/metric_images/{metric_image_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_alert_management_alerts_alert_iid_metric_images_metric_image_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --alert-iid, --metric-image-id
    project artifacts delete - DELETE /projects/{id}/artifacts [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_artifacts]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project badges delete - DELETE /projects/{id}/badges/{badge_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_badges_badge_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --badge-id
    project cluster-agents delete - DELETE /projects/{id}/cluster_agents/{agent_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_cluster_agents_agent_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --agent-id
    project clusters delete - DELETE /projects/{id}/clusters/{cluster_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_clusters_cluster_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --cluster-id
    project custom-attributes delete - DELETE /projects/{id}/custom_attributes/{key} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_custom_attributes_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --key
    project debian-distributions delete - DELETE /projects/{id}/debian_distributions/{codename} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_debian_distributions_codename]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --codename
    project deploy-keys delete - DELETE /projects/{id}/deploy_keys/{key_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_deploy_keys_key_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --key-id
    project deployments delete - DELETE /projects/{id}/deployments/{deployment_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_deployments_deployment_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --deployment-id
    project review-apps delete - DELETE /projects/{id}/environments/review_apps [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_environments_review_apps]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project environments delete - DELETE /projects/{id}/environments/{environment_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_environments_environment_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --environment-id
    project client-keys delete - DELETE /projects/{id}/error_tracking/client_keys/{key_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_error_tracking_client_keys_key_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --key-id
    project feature-flags delete - DELETE /projects/{id}/feature_flags/{feature_flag_name} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_feature_flags_feature_flag_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --feature-flag-name
    project feature-flags-user-lists delete - DELETE /projects/{id}/feature_flags_user_lists/{iid} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_feature_flags_user_lists_iid]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --iid
    project fork delete - DELETE /projects/{id}/fork [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_fork]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project freeze-periods delete - DELETE /projects/{id}/freeze_periods/{freeze_period_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_freeze_periods_freeze_period_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --freeze-period-id
    project hooks delete - DELETE /projects/{id}/hooks/{hook_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_hooks_hook_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --hook-id
    project custom-headers delete - DELETE /projects/{id}/hooks/{hook_id}/custom_headers/{key} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_hooks_hook_id_custom_headers_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --hook-id, --key
    project integrations delete - DELETE /projects/{id}/integrations/{slug} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_integrations_slug]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --slug
    project invitations delete - DELETE /projects/{id}/invitations/{email} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_invitations_email]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --email
    project artifacts delete-2 - DELETE /projects/{id}/jobs/{job_id}/artifacts [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_jobs_job_id_artifacts]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --job-id
    project conans delete - DELETE /projects/{id}/packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_packages_conan_v1_conans_package_name_package_version_package_username_package_channel]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-name, --package-version, --package-username, --package-channel
    project revisions delete - DELETE /projects/{id}/packages/conan/v2/conans/{package_name}/{package_version}/{package_username}/{package_channel}/revisions/{recipe_revision} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_packages_conan_v2_conans_package_name_package_version_package_username_package_channel_revisions_recipe_revision]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision
    project revisions delete-2 - DELETE /projects/{id}/packages/conan/v2/conans/{package_name}/{package_version}/{package_username}/{package_channel}/revisions/{recipe_revision}/packages/{conan_package_reference}/revisions/{package_revision} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_packages_conan_v2_conans_package_name_package_version_package_username_package_channel_revisions_recipe_revision_packages_conan_package_reference_revisions_package_revision]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --conan-package-reference, --package-revision
    project dist-tags delete - DELETE /projects/{id}/packages/npm/-/package/*package_name/dist-tags/{tag} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_packages_npm_package_package_name_dist_tags_tag]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --tag
    project *package-version delete - DELETE /projects/{id}/packages/nuget/*package_name/*package_version [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_packages_nuget_package_name_package_version]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project rules delete - DELETE /projects/{id}/packages/protection/rules/{package_protection_rule_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_packages_protection_rules_package_protection_rule_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-protection-rule-id
    project packages delete - DELETE /projects/{id}/packages/{package_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_packages_package_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-id
    project package-files delete - DELETE /projects/{id}/packages/{package_id}/package_files/{package_file_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_packages_package_id_package_files_package_file_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-id, --package-file-id
    project pages delete - DELETE /projects/{id}/pages [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_pages]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project domains delete - DELETE /projects/{id}/pages/domains/{domain} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_pages_domains_domain]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --domain
    project pipeline-schedules delete - DELETE /projects/{id}/pipeline_schedules/{pipeline_schedule_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_pipeline_schedules_pipeline_schedule_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --pipeline-schedule-id
    project protected-branches delete - DELETE /projects/{id}/protected_branches/{name} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_protected_branches_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --name
    project protected-tags delete - DELETE /projects/{id}/protected_tags/{name} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_protected_tags_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --name
    project rules delete-2 - DELETE /projects/{id}/registry/protection/tag/rules/{protection_rule_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_registry_protection_tag_rules_protection_rule_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --protection-rule-id
    project releases delete - DELETE /projects/{id}/releases/{tag_name} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_releases_tag_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --tag-name
    project links delete - DELETE /projects/{id}/releases/{tag_name}/assets/links/{link_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_releases_tag_name_assets_links_link_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --tag-name, --link-id
    project remote-mirrors delete - DELETE /projects/{id}/remote_mirrors/{mirror_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_remote_mirrors_mirror_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --mirror-id
    project secure-files delete - DELETE /projects/{id}/secure_files/{secure_file_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_secure_files_secure_file_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --secure-file-id
    project services delete - DELETE /projects/{id}/services/{slug} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_services_slug]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --slug
    project share delete - DELETE /projects/{id}/share/{group_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_share_group_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --group-id
    project snippets delete - DELETE /projects/{id}/snippets/{snippet_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_snippets_snippet_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --snippet-id
    project award-emoji delete - DELETE /projects/{id}/snippets/{snippet_id}/award_emoji/{award_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_snippets_snippet_id_award_emoji_award_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --snippet-id, --award-id
    project award-emoji delete-2 - DELETE /projects/{id}/snippets/{snippet_id}/notes/{note_id}/award_emoji/{award_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_snippets_snippet_id_notes_note_id_award_emoji_award_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --snippet-id, --note-id, --award-id
    project state delete - DELETE /projects/{id}/terraform/state/{name} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_terraform_state_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --name
    project lock delete - DELETE /projects/{id}/terraform/state/{name}/lock [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_terraform_state_name_lock]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --name
    project versions delete - DELETE /projects/{id}/terraform/state/{name}/versions/{serial} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_terraform_state_name_versions_serial]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --name, --serial
    project state-protection-rules delete - DELETE /projects/{id}/terraform/state_protection_rules/{terraform_state_protection_rule_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_terraform_state_protection_rules_terraform_state_protection_rule_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --terraform-state-protection-rule-id
    project triggers delete - DELETE /projects/{id}/triggers/{trigger_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_triggers_trigger_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --trigger-id
    project uploads delete - DELETE /projects/{id}/uploads/{secret}/{filename} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_uploads_secret_filename]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --secret, --filename
    project uploads delete-2 - DELETE /projects/{id}/uploads/{upload_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_uploads_upload_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --upload-id
    project wikis delete - DELETE /projects/{id}/wikis/{slug} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_wikis_slug]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --slug
    project )search view - GET /projects/{id}/(-/)search [intent=direct_read availability=implemented]; flags: --id
    project metric-images view - GET /projects/{id}/alert_management_alerts/{alert_iid}/metric_images [intent=direct_read availability=implemented]; flags: --id, --alert-iid
    project download download - GET /projects/{id}/attestations/{attestation_iid}/download [intent=direct_read availability=implemented]; flags: --id, --attestation-iid
    project attestations view - GET /projects/{id}/attestations/{subject_digest} [intent=direct_read availability=implemented]; flags: --id, --subject-digest
    project audit-events view - GET /projects/{id}/audit_events [intent=direct_read availability=implemented]; flags: --id
    project audit-events view-2 - GET /projects/{id}/audit_events/{audit_event_id} [intent=direct_read availability=implemented]; flags: --id, --audit-event-id
    project avatar download - GET /projects/{id}/avatar [intent=direct_read availability=implemented]; flags: --id
    project badges view - GET /projects/{id}/badges [intent=direct_read availability=implemented]; flags: --id
    project render view - GET /projects/{id}/badges/render [intent=direct_read availability=implemented]; flags: --id
    project badges view-2 - GET /projects/{id}/badges/{badge_id} [intent=direct_read availability=implemented]; flags: --id, --badge-id
    project lint view - GET /projects/{id}/ci/lint [intent=direct_read availability=implemented]; flags: --id
    project cluster-agents view - GET /projects/{id}/cluster_agents [intent=direct_read availability=implemented]; flags: --id
    project cluster-agents view-2 - GET /projects/{id}/cluster_agents/{agent_id} [intent=direct_read availability=implemented]; flags: --id, --agent-id
    project clusters view - GET /projects/{id}/clusters [intent=direct_read availability=implemented]; flags: --id
    project clusters view-2 - GET /projects/{id}/clusters/{cluster_id} [intent=direct_read availability=implemented]; flags: --id, --cluster-id
    project custom-attributes view - GET /projects/{id}/custom_attributes [intent=direct_read availability=implemented]; flags: --id
    project custom-attributes view-2 - GET /projects/{id}/custom_attributes/{key} [intent=direct_read availability=implemented]; flags: --id, --key
    project debian-distributions view - GET /projects/{id}/debian_distributions [intent=direct_read availability=implemented]; flags: --id
    project debian-distributions view-2 - GET /projects/{id}/debian_distributions/{codename} [intent=direct_read availability=implemented]; flags: --id, --codename
    project key.asc view - GET /projects/{id}/debian_distributions/{codename}/key.asc [intent=direct_read availability=implemented]; flags: --id, --codename
    project deploy-keys view - GET /projects/{id}/deploy_keys [intent=direct_read availability=implemented]; flags: --id
    project deploy-keys view-2 - GET /projects/{id}/deploy_keys/{key_id} [intent=direct_read availability=implemented]; flags: --id, --key-id
    project deployments view - GET /projects/{id}/deployments [intent=direct_read availability=implemented]; flags: --id
    project deployments view-2 - GET /projects/{id}/deployments/{deployment_id} [intent=direct_read availability=implemented]; flags: --id, --deployment-id
    project environments view - GET /projects/{id}/environments [intent=direct_read availability=implemented]; flags: --id
    project environments view-2 - GET /projects/{id}/environments/{environment_id} [intent=direct_read availability=implemented]; flags: --id, --environment-id
    project client-keys view - GET /projects/{id}/error_tracking/client_keys [intent=direct_read availability=implemented]; flags: --id
    project settings view - GET /projects/{id}/error_tracking/settings [intent=direct_read availability=implemented]; flags: --id
    project events view - GET /projects/{id}/events [intent=direct_read availability=implemented]; flags: --id
    project export view - GET /projects/{id}/export [intent=direct_read availability=implemented]; flags: --id
    project download download-2 - GET /projects/{id}/export/download [intent=direct_read availability=implemented]; flags: --id
    project download download-3 - GET /projects/{id}/export_relations/download [intent=direct_read availability=implemented]; flags: --id
    project status view - GET /projects/{id}/export_relations/status [intent=direct_read availability=implemented]; flags: --id
    project feature-flags view - GET /projects/{id}/feature_flags [intent=direct_read availability=implemented]; flags: --id
    project feature-flags view-2 - GET /projects/{id}/feature_flags/{feature_flag_name} [intent=direct_read availability=implemented]; flags: --id, --feature-flag-name
    project feature-flags-user-lists view - GET /projects/{id}/feature_flags_user_lists [intent=direct_read availability=implemented]; flags: --id
    project feature-flags-user-lists view-2 - GET /projects/{id}/feature_flags_user_lists/{iid} [intent=direct_read availability=implemented]; flags: --id, --iid
    project forks view - GET /projects/{id}/forks [intent=direct_read availability=implemented]; flags: --id
    project freeze-periods view - GET /projects/{id}/freeze_periods [intent=direct_read availability=implemented]; flags: --id
    project freeze-periods view-2 - GET /projects/{id}/freeze_periods/{freeze_period_id} [intent=direct_read availability=implemented]; flags: --id, --freeze-period-id
    project hooks view - GET /projects/{id}/hooks [intent=direct_read availability=implemented]; flags: --id
    project hooks view-2 - GET /projects/{id}/hooks/{hook_id} [intent=direct_read availability=implemented]; flags: --id, --hook-id
    project events view-2 - GET /projects/{id}/hooks/{hook_id}/events [intent=direct_read availability=implemented]; flags: --id, --hook-id
    project import view - GET /projects/{id}/import [intent=direct_read availability=implemented]; flags: --id
    project integrations view - GET /projects/{id}/integrations [intent=direct_read availability=implemented]; flags: --id
    project integrations view-2 - GET /projects/{id}/integrations/{slug} [intent=direct_read availability=implemented]; flags: --id, --slug
    project invitations view - GET /projects/{id}/invitations [intent=direct_read availability=implemented]; flags: --id
    project jobs view - GET /projects/{id}/jobs [intent=direct_read availability=implemented]; flags: --id
    project download download-4 - GET /projects/{id}/jobs/artifacts/{ref_name}/download [intent=direct_read availability=implemented]; flags: --id, --ref-name
    project *artifact-path download - GET /projects/{id}/jobs/artifacts/{ref_name}/raw/*artifact_path [intent=direct_read availability=implemented]; flags: --id, --ref-name
    project jobs view-2 - GET /projects/{id}/jobs/{job_id} [intent=direct_read availability=implemented]; flags: --id, --job-id
    project artifacts download - GET /projects/{id}/jobs/{job_id}/artifacts [intent=direct_read availability=implemented]; flags: --id, --job-id
    project *artifact-path download-2 - GET /projects/{id}/jobs/{job_id}/artifacts/*artifact_path [intent=direct_read availability=implemented]; flags: --id, --job-id
    project tree download - GET /projects/{id}/jobs/{job_id}/artifacts/tree [intent=direct_read availability=implemented]; flags: --id, --job-id
    project trace download - GET /projects/{id}/jobs/{job_id}/trace [intent=direct_read availability=implemented]; flags: --id, --job-id
    project languages view - GET /projects/{id}/languages [intent=direct_read availability=implemented]; flags: --id
    project packages view - GET /projects/{id}/packages [intent=direct_read availability=implemented]; flags: --id
    project 1 download - GET /projects/{id}/packages/cargo/1/{package_name} [intent=direct_read availability=implemented]; flags: --id, --package-name
    project 2 download - GET /projects/{id}/packages/cargo/2/{package_name} [intent=direct_read availability=implemented]; flags: --id, --package-name
    project 3 download - GET /projects/{id}/packages/cargo/3/{first_char}/{package_name} [intent=direct_read availability=implemented]; flags: --id, --first-char, --package-name
    project config.json download - GET /projects/{id}/packages/cargo/config.json [intent=direct_read availability=implemented]; flags: --id
    project download download-5 - GET /projects/{id}/packages/cargo/{package_name}/{package_version}/download [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version
    project cargo download - GET /projects/{id}/packages/cargo/{prefix_1}/{prefix_2}/{package_name} [intent=direct_read availability=implemented]; flags: --id, --prefix-1, --prefix-2, --package-name
    project *package-name download - GET /projects/{id}/packages/composer/archives/*package_name [intent=direct_read availability=implemented]; flags: --id
    project search download - GET /projects/{id}/packages/conan/v1/conans/search [intent=direct_read availability=implemented]; flags: --id
    project conans download - GET /projects/{id}/packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel} [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel
    project digest download - GET /projects/{id}/packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/digest [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel
    project download-urls download - GET /projects/{id}/packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/download_urls [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel
    project packages download - GET /projects/{id}/packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/packages/{conan_package_reference} [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel, --conan-package-reference
    project digest download-2 - GET /projects/{id}/packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/packages/{conan_package_reference}/digest [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel, --conan-package-reference
    project download-urls download-2 - GET /projects/{id}/packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/packages/{conan_package_reference}/download_urls [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel, --conan-package-reference
    project search download-2 - GET /projects/{id}/packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/search [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel
    project export download - GET /projects/{id}/packages/conan/v1/files/{package_name}/{package_version}/{package_username}/{package_channel}/{recipe_revision}/export/{file_name} [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --file-name
    project package download - GET /projects/{id}/packages/conan/v1/files/{package_name}/{package_version}/{package_username}/{package_channel}/{recipe_revision}/package/{conan_package_reference}/{package_revision}/{file_name} [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --conan-package-reference, --package-revision, --file-name
    project ping download - GET /projects/{id}/packages/conan/v1/ping [intent=direct_read availability=implemented]; flags: --id
    project authenticate download - GET /projects/{id}/packages/conan/v1/users/authenticate [intent=direct_read availability=implemented]; flags: --id
    project check-credentials download - GET /projects/{id}/packages/conan/v1/users/check_credentials [intent=direct_read availability=implemented]; flags: --id
    project search download-3 - GET /projects/{id}/packages/conan/v2/conans/search [intent=direct_read availability=implemented]; flags: --id
    project latest download - GET /projects/{id}/packages/conan/v2/conans/{package_name}/{package_version}/{package_username}/{package_channel}/latest [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel
    project revisions download - GET /projects/{id}/packages/conan/v2/conans/{package_name}/{package_version}/{package_username}/{package_channel}/revisions [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel
    project files download - GET /projects/{id}/packages/conan/v2/conans/{package_name}/{package_version}/{package_username}/{package_channel}/revisions/{recipe_revision}/files [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision
    project files download-2 - GET /projects/{id}/packages/conan/v2/conans/{package_name}/{package_version}/{package_username}/{package_channel}/revisions/{recipe_revision}/files/{file_name} [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --file-name
    project latest download-2 - GET /projects/{id}/packages/conan/v2/conans/{package_name}/{package_version}/{package_username}/{package_channel}/revisions/{recipe_revision}/packages/{conan_package_reference}/latest [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --conan-package-reference
    project revisions download-2 - GET /projects/{id}/packages/conan/v2/conans/{package_name}/{package_version}/{package_username}/{package_channel}/revisions/{recipe_revision}/packages/{conan_package_reference}/revisions [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --conan-package-reference
    project files download-3 - GET /projects/{id}/packages/conan/v2/conans/{package_name}/{package_version}/{package_username}/{package_channel}/revisions/{recipe_revision}/packages/{conan_package_reference}/revisions/{package_revision}/files [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --conan-package-reference, --package-revision
    project files download-4 - GET /projects/{id}/packages/conan/v2/conans/{package_name}/{package_version}/{package_username}/{package_channel}/revisions/{recipe_revision}/packages/{conan_package_reference}/revisions/{package_revision}/files/{file_name} [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --conan-package-reference, --package-revision, --file-name
    project search download-4 - GET /projects/{id}/packages/conan/v2/conans/{package_name}/{package_version}/{package_username}/{package_channel}/revisions/{recipe_revision}/search [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision
    project search download-5 - GET /projects/{id}/packages/conan/v2/conans/{package_name}/{package_version}/{package_username}/{package_channel}/search [intent=direct_read availability=implemented]; flags: --id, --package-name, --package-version, --package-username, --package-channel
    project authenticate download-2 - GET /projects/{id}/packages/conan/v2/users/authenticate [intent=direct_read availability=implemented]; flags: --id
    project check-credentials download-2 - GET /projects/{id}/packages/conan/v2/users/check_credentials [intent=direct_read availability=implemented]; flags: --id
    project InRelease download - GET /projects/{id}/packages/debian/dists/*distribution/InRelease [intent=direct_read availability=implemented]; flags: --id
    project Release download - GET /projects/{id}/packages/debian/dists/*distribution/Release [intent=direct_read availability=implemented]; flags: --id
    project Release.gpg download - GET /projects/{id}/packages/debian/dists/*distribution/Release.gpg [intent=direct_read availability=implemented]; flags: --id
    project Packages download - GET /projects/{id}/packages/debian/dists/*distribution/{component}/binary-{architecture}/Packages [intent=direct_read availability=implemented]; flags: --id, --component, --architecture
    project SHA256 download - GET /projects/{id}/packages/debian/dists/*distribution/{component}/binary-{architecture}/by-hash/SHA256/{file_sha256} [intent=direct_read availability=implemented]; flags: --id, --component, --architecture, --file-sha256
    project Packages download-2 - GET /projects/{id}/packages/debian/dists/*distribution/{component}/debian-installer/binary-{architecture}/Packages [intent=direct_read availability=implemented]; flags: --id, --component, --architecture
    project SHA256 download-2 - GET /projects/{id}/packages/debian/dists/*distribution/{component}/debian-installer/binary-{architecture}/by-hash/SHA256/{file_sha256} [intent=direct_read availability=implemented]; flags: --id, --component, --architecture, --file-sha256
    project Sources download - GET /projects/{id}/packages/debian/dists/*distribution/{component}/source/Sources [intent=direct_read availability=implemented]; flags: --id, --component
    project SHA256 download-3 - GET /projects/{id}/packages/debian/dists/*distribution/{component}/source/by-hash/SHA256/{file_sha256} [intent=direct_read availability=implemented]; flags: --id, --component, --file-sha256
    project pool download - GET /projects/{id}/packages/debian/pool/{distribution}/{letter}/{package_name}/{package_version}/{file_name} [intent=direct_read availability=implemented]; flags: --id, --distribution, --letter, --package-name, --package-version, --file-name
    project (*path download - GET /projects/{id}/packages/generic/{package_name}/*package_version/(*path/){file_name} [intent=direct_read availability=implemented]; flags: --id, --package-name, --file-name
    project list download - GET /projects/{id}/packages/go/*module_name/@v/list [intent=direct_read availability=implemented]; flags: --id
    project @v download - GET /projects/{id}/packages/go/*module_name/@v/{module_version}.info [intent=direct_read availability=implemented]; flags: --id, --module-version
    project @v download-2 - GET /projects/{id}/packages/go/*module_name/@v/{module_version}.mod [intent=direct_read availability=implemented]; flags: --id, --module-version
    project @v download-3 - GET /projects/{id}/packages/go/*module_name/@v/{module_version}.zip [intent=direct_read availability=implemented]; flags: --id, --module-version
    project charts download - GET /projects/{id}/packages/helm/{channel}/charts/{file_name}.tgz [intent=direct_read availability=implemented]; flags: --id, --channel, --file-name
    project index.yaml download - GET /projects/{id}/packages/helm/{channel}/index.yaml [intent=direct_read availability=implemented]; flags: --id, --channel
    project *path download - GET /projects/{id}/packages/maven/*path/{file_name} [intent=direct_read availability=implemented]; flags: --id, --file-name
    project (*path download-2 - GET /projects/{id}/packages/ml_models/{model_version_id}/files/(*path/){file_name} [intent=direct_read availability=implemented]; flags: --id, --model-version-id, --file-name
    project *package-name download-2 - GET /projects/{id}/packages/npm/*package_name [intent=direct_read availability=implemented]; flags: --id
    project *file-name download - GET /projects/{id}/packages/npm/*package_name/-/*file_name [intent=direct_read availability=implemented]; flags: --id
    project dist-tags download - GET /projects/{id}/packages/npm/-/package/*package_name/dist-tags [intent=direct_read availability=implemented]; flags: --id
    project *package-filename download - GET /projects/{id}/packages/nuget/download/*package_name/*package_version/*package_filename [intent=direct_read availability=implemented]; flags: --id
    project index download - GET /projects/{id}/packages/nuget/download/*package_name/index [intent=direct_read availability=implemented]; flags: --id
    project index download-2 - GET /projects/{id}/packages/nuget/index [intent=direct_read availability=implemented]; flags: --id
    project *package-version download - GET /projects/{id}/packages/nuget/metadata/*package_name/*package_version [intent=direct_read availability=implemented]; flags: --id
    project index download-3 - GET /projects/{id}/packages/nuget/metadata/*package_name/index [intent=direct_read availability=implemented]; flags: --id
    project query download - GET /projects/{id}/packages/nuget/query [intent=direct_read availability=implemented]; flags: --id
    project *same-file-name download - GET /projects/{id}/packages/nuget/symbolfiles/*file_name/*signature/*same_file_name [intent=direct_read availability=implemented]; flags: --id
    project v2 download - GET /projects/{id}/packages/nuget/v2 [intent=direct_read availability=implemented]; flags: --id
    project $metadata download - GET /projects/{id}/packages/nuget/v2/$metadata [intent=direct_read availability=implemented]; flags: --id
    project rules download - GET /projects/{id}/packages/protection/rules [intent=direct_read availability=implemented]; flags: --id
    project *file-identifier download - GET /projects/{id}/packages/pypi/files/{sha256}/*file_identifier [intent=direct_read availability=implemented]; flags: --id, --sha256
    project simple download - GET /projects/{id}/packages/pypi/simple [intent=direct_read availability=implemented]; flags: --id
    project *package-name download-3 - GET /projects/{id}/packages/pypi/simple/*package_name [intent=direct_read availability=implemented]; flags: --id
    project *file-name download-2 - GET /projects/{id}/packages/rpm/*package_file_id/*file_name [intent=direct_read availability=implemented]; flags: --id
    project *file-name download-3 - GET /projects/{id}/packages/rpm/repodata/*file_name [intent=direct_read availability=implemented]; flags: --id
    project dependencies download - GET /projects/{id}/packages/rubygems/api/v1/dependencies [intent=direct_read availability=implemented]; flags: --id
    project gems download - GET /projects/{id}/packages/rubygems/gems/{file_name} [intent=direct_read availability=implemented]; flags: --id, --file-name
    project Marshal.4.8 download - GET /projects/{id}/packages/rubygems/quick/Marshal.4.8/{file_name} [intent=direct_read availability=implemented]; flags: --id, --file-name
    project rubygems download - GET /projects/{id}/packages/rubygems/{file_name} [intent=direct_read availability=implemented]; flags: --id, --file-name
    project modules download - GET /projects/{id}/packages/terraform/modules/{module_name}/{module_system} [intent=direct_read availability=implemented]; flags: --id, --module-name, --module-system
    project *module-version download - GET /projects/{id}/packages/terraform/modules/{module_name}/{module_system}/*module_version [intent=direct_read availability=implemented]; flags: --id, --module-name, --module-system
    project packages download-2 - GET /projects/{id}/packages/{package_id} [intent=direct_read availability=implemented]; flags: --id, --package-id
    project package-files download - GET /projects/{id}/packages/{package_id}/package_files [intent=direct_read availability=implemented]; flags: --id, --package-id
    project download download-6 - GET /projects/{id}/packages/{package_id}/package_files/{package_file_id}/download [intent=direct_read availability=implemented]; flags: --id, --package-id, --package-file-id
    project pages view - GET /projects/{id}/pages [intent=direct_read availability=implemented]; flags: --id
    project domains view - GET /projects/{id}/pages/domains [intent=direct_read availability=implemented]; flags: --id
    project domains view-2 - GET /projects/{id}/pages/domains/{domain} [intent=direct_read availability=implemented]; flags: --id, --domain
    project pipeline-schedules view - GET /projects/{id}/pipeline_schedules [intent=direct_read availability=implemented]; flags: --id
    project pipeline-schedules view-2 - GET /projects/{id}/pipeline_schedules/{pipeline_schedule_id} [intent=direct_read availability=implemented]; flags: --id, --pipeline-schedule-id
    project protected-branches view - GET /projects/{id}/protected_branches [intent=direct_read availability=implemented]; flags: --id
    project protected-branches view-2 - GET /projects/{id}/protected_branches/{name} [intent=direct_read availability=implemented]; flags: --id, --name
    project protected-tags view - GET /projects/{id}/protected_tags [intent=direct_read availability=implemented]; flags: --id
    project protected-tags view-2 - GET /projects/{id}/protected_tags/{name} [intent=direct_read availability=implemented]; flags: --id, --name
    project rules view - GET /projects/{id}/registry/protection/tag/rules [intent=direct_read availability=implemented]; flags: --id
    project repositories view - GET /projects/{id}/registry/repositories [intent=direct_read availability=implemented]; flags: --id
    project relation-imports view - GET /projects/{id}/relation-imports [intent=direct_read availability=implemented]; flags: --id
    project releases view - GET /projects/{id}/releases [intent=direct_read availability=implemented]; flags: --id
    project )(*suffix-path) view - GET /projects/{id}/releases/permalink/latest(/)(*suffix_path) [intent=direct_read availability=implemented]; flags: --id
    project releases view-2 - GET /projects/{id}/releases/{tag_name} [intent=direct_read availability=implemented]; flags: --id, --tag-name
    project links view - GET /projects/{id}/releases/{tag_name}/assets/links [intent=direct_read availability=implemented]; flags: --id, --tag-name
    project links view-2 - GET /projects/{id}/releases/{tag_name}/assets/links/{link_id} [intent=direct_read availability=implemented]; flags: --id, --tag-name, --link-id
    project *direct-asset-path download - GET /projects/{id}/releases/{tag_name}/downloads/*direct_asset_path [intent=direct_read availability=implemented]; flags: --id, --tag-name
    project remote-mirrors view - GET /projects/{id}/remote_mirrors [intent=direct_read availability=implemented]; flags: --id
    project remote-mirrors view-2 - GET /projects/{id}/remote_mirrors/{mirror_id} [intent=direct_read availability=implemented]; flags: --id, --mirror-id
    project public-key view - GET /projects/{id}/remote_mirrors/{mirror_id}/public_key [intent=direct_read availability=implemented]; flags: --id, --mirror-id
    project secure-files download - GET /projects/{id}/secure_files [intent=direct_read availability=implemented]; flags: --id
    project secure-files download-2 - GET /projects/{id}/secure_files/{secure_file_id} [intent=direct_read availability=implemented]; flags: --id, --secure-file-id
    project download download-7 - GET /projects/{id}/secure_files/{secure_file_id}/download [intent=direct_read availability=implemented]; flags: --id, --secure-file-id
    project services view - GET /projects/{id}/services [intent=direct_read availability=implemented]; flags: --id
    project services view-2 - GET /projects/{id}/services/{slug} [intent=direct_read availability=implemented]; flags: --id, --slug
    project share-locations view - GET /projects/{id}/share_locations [intent=direct_read availability=implemented]; flags: --id
    project snapshot view - GET /projects/{id}/snapshot [intent=direct_read availability=implemented]; flags: --id
    project snippets view - GET /projects/{id}/snippets [intent=direct_read availability=implemented]; flags: --id
    project snippets view-2 - GET /projects/{id}/snippets/{snippet_id} [intent=direct_read availability=implemented]; flags: --id, --snippet-id
    project award-emoji view - GET /projects/{id}/snippets/{snippet_id}/award_emoji [intent=direct_read availability=implemented]; flags: --id, --snippet-id
    project award-emoji view-2 - GET /projects/{id}/snippets/{snippet_id}/award_emoji/{award_id} [intent=direct_read availability=implemented]; flags: --id, --snippet-id, --award-id
    project raw download - GET /projects/{id}/snippets/{snippet_id}/files/{ref}/{file_path}/raw [intent=direct_read availability=implemented]; flags: --id, --snippet-id, --ref, --file-path
    project award-emoji view-3 - GET /projects/{id}/snippets/{snippet_id}/notes/{note_id}/award_emoji [intent=direct_read availability=implemented]; flags: --id, --snippet-id, --note-id
    project award-emoji view-4 - GET /projects/{id}/snippets/{snippet_id}/notes/{note_id}/award_emoji/{award_id} [intent=direct_read availability=implemented]; flags: --id, --snippet-id, --note-id, --award-id
    project raw download-2 - GET /projects/{id}/snippets/{snippet_id}/raw [intent=direct_read availability=implemented]; flags: --id, --snippet-id
    project user-agent-detail view - GET /projects/{id}/snippets/{snippet_id}/user_agent_detail [intent=direct_read availability=implemented]; flags: --id, --snippet-id
    project starrers view - GET /projects/{id}/starrers [intent=direct_read availability=implemented]; flags: --id
    project statistics view - GET /projects/{id}/statistics [intent=direct_read availability=implemented]; flags: --id
    project storage view - GET /projects/{id}/storage [intent=direct_read availability=implemented]; flags: --id
    project templates view - GET /projects/{id}/templates/{type} [intent=direct_read availability=implemented]; flags: --id, --type
    project templates view-2 - GET /projects/{id}/templates/{type}/{name} [intent=direct_read availability=implemented]; flags: --id, --type, --name
    project state download - GET /projects/{id}/terraform/state/{name} [intent=direct_read availability=implemented]; flags: --id, --name
    project versions download - GET /projects/{id}/terraform/state/{name}/versions/{serial} [intent=direct_read availability=implemented]; flags: --id, --name, --serial
    project state-protection-rules download - GET /projects/{id}/terraform/state_protection_rules [intent=direct_read availability=implemented]; flags: --id
    project transfer-locations view - GET /projects/{id}/transfer_locations [intent=direct_read availability=implemented]; flags: --id
    project triggers view - GET /projects/{id}/triggers [intent=direct_read availability=implemented]; flags: --id
    project triggers view-2 - GET /projects/{id}/triggers/{trigger_id} [intent=direct_read availability=implemented]; flags: --id, --trigger-id
    project uploads download - GET /projects/{id}/uploads [intent=direct_read availability=implemented]; flags: --id
    project uploads download-2 - GET /projects/{id}/uploads/{secret}/{filename} [intent=direct_read availability=implemented]; flags: --id, --secret, --filename
    project uploads download-3 - GET /projects/{id}/uploads/{upload_id} [intent=direct_read availability=implemented]; flags: --id, --upload-id
    project users view - GET /projects/{id}/users [intent=direct_read availability=implemented]; flags: --id
    project wikis view - GET /projects/{id}/wikis [intent=direct_read availability=implemented]; flags: --id
    project wikis view-2 - GET /projects/{id}/wikis/{slug} [intent=direct_read availability=implemented]; flags: --id, --slug
    project FindPackagesById\(\) download - GET /projects/{project_id}/packages/nuget/v2/FindPackagesById\(\) [intent=direct_read availability=implemented]; flags: --project-id
    project Packages\(Id='*package-name',Version='*package-version'\) download - GET /projects/{project_id}/packages/nuget/v2/Packages\(Id='*package_name',Version='*package_version'\) [intent=direct_read availability=implemented]; flags: --project-id
    project Packages\(\) download - GET /projects/{project_id}/packages/nuget/v2/Packages\(\) [intent=direct_read availability=implemented]; flags: --project-id
    project contributed-projects view - GET /users/{user_id}/contributed_projects [intent=direct_read availability=implemented]; flags: --user-id
    project view-2 - GET /users/{user_id}/projects [intent=direct_read availability=implemented]; flags: --user-id
    project starred-projects view - GET /users/{user_id}/starred_projects [intent=direct_read availability=implemented]; flags: --user-id
    project settings update - PATCH /projects/{id}/error_tracking/settings [intent=reverse_etl availability=implemented write=patch_api_v4_projects_id_error_tracking_settings]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project rules update - PATCH /projects/{id}/packages/protection/rules/{package_protection_rule_id} [intent=reverse_etl availability=implemented write=patch_api_v4_projects_id_packages_protection_rules_package_protection_rule_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-protection-rule-id
    project pages update - PATCH /projects/{id}/pages [intent=reverse_etl availability=implemented write=patch_api_v4_projects_id_pages]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project protected-branches update - PATCH /projects/{id}/protected_branches/{name} [intent=reverse_etl availability=implemented write=patch_api_v4_projects_id_protected_branches_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --name
    project rules update-2 - PATCH /projects/{id}/registry/protection/tag/rules/{protection_rule_id} [intent=reverse_etl availability=implemented write=patch_api_v4_projects_id_registry_protection_tag_rules_protection_rule_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --protection-rule-id
    project state-protection-rules update - PATCH /projects/{id}/terraform/state_protection_rules/{terraform_state_protection_rule_id} [intent=reverse_etl availability=implemented write=patch_api_v4_projects_id_terraform_state_protection_rules_terraform_state_protection_rule_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --terraform-state-protection-rule-id
    project create - POST /projects [intent=reverse_etl availability=implemented write=post_api_v4_projects]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    project import create - POST /projects/import [intent=reverse_etl availability=implemented write=post_api_v4_projects_import]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    project import-relation create - POST /projects/import-relation [intent=reverse_etl availability=implemented write=post_api_v4_projects_import_relation]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    project authorize create - POST /projects/import-relation/authorize [intent=reverse_etl availability=implemented write=post_api_v4_projects_import_relation_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    project authorize create-2 - POST /projects/import/authorize [intent=reverse_etl availability=implemented write=post_api_v4_projects_import_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    project remote-import create - POST /projects/remote-import [intent=reverse_etl availability=implemented write=post_api_v4_projects_remote_import]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    project remote-import-s3 create - POST /projects/remote-import-s3 [intent=reverse_etl availability=implemented write=post_api_v4_projects_remote_import_s3]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    project user create - POST /projects/user/{user_id} [intent=reverse_etl availability=implemented write=post_api_v4_projects_user_user_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --user-id
    project pipeline create - POST /projects/{id}/(ref/{ref}/)trigger/pipeline [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_ref_ref_trigger_pipeline]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --ref
    project metric-images create - POST /projects/{id}/alert_management_alerts/{alert_iid}/metric_images [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_alert_management_alerts_alert_iid_metric_images]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --alert-iid
    project authorize create-3 - POST /projects/{id}/alert_management_alerts/{alert_iid}/metric_images/authorize [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_alert_management_alerts_alert_iid_metric_images_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --alert-iid
    project archive create - POST /projects/{id}/archive [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_archive]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project badges create - POST /projects/{id}/badges [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_badges]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project publish create - POST /projects/{id}/catalog/publish [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_catalog_publish]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project lint create - POST /projects/{id}/ci/lint [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_ci_lint]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project cluster-agents create - POST /projects/{id}/cluster_agents [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_cluster_agents]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project user create-2 - POST /projects/{id}/clusters/user [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_clusters_user]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project create-ci-config create - POST /projects/{id}/create_ci_config [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_create_ci_config]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project debian-distributions create - POST /projects/{id}/debian_distributions [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_debian_distributions]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project deploy-keys create - POST /projects/{id}/deploy_keys [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_deploy_keys]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project enable create - POST /projects/{id}/deploy_keys/{key_id}/enable [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_deploy_keys_key_id_enable]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --key-id
    project deployments create - POST /projects/{id}/deployments [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_deployments]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project approval create - POST /projects/{id}/deployments/{deployment_id}/approval [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_deployments_deployment_id_approval]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --deployment-id
    project environments create - POST /projects/{id}/environments [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_environments]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project stop-stale create - POST /projects/{id}/environments/stop_stale [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_environments_stop_stale]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project stop create - POST /projects/{id}/environments/{environment_id}/stop [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_environments_environment_id_stop]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --environment-id
    project client-keys create - POST /projects/{id}/error_tracking/client_keys [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_error_tracking_client_keys]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project export create - POST /projects/{id}/export [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_export]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project export-relations create - POST /projects/{id}/export_relations [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_export_relations]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project feature-flags create - POST /projects/{id}/feature_flags [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_feature_flags]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project feature-flags-user-lists create - POST /projects/{id}/feature_flags_user_lists [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_feature_flags_user_lists]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project fork create - POST /projects/{id}/fork [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_fork]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project fork create-2 - POST /projects/{id}/fork/{forked_from_id} [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_fork_forked_from_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --forked-from-id
    project freeze-periods create - POST /projects/{id}/freeze_periods [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_freeze_periods]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project hooks create - POST /projects/{id}/hooks [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_hooks]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project resend create - POST /projects/{id}/hooks/{hook_id}/events/{hook_log_id}/resend [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_hooks_hook_id_events_hook_log_id_resend]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --hook-id, --hook-log-id
    project test create - POST /projects/{id}/hooks/{hook_id}/test/{trigger} [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_hooks_hook_id_test_trigger]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --hook-id, --trigger
    project housekeeping create - POST /projects/{id}/housekeeping [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_housekeeping]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project git create - POST /projects/{id}/import/git [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_import_git]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project trigger create - POST /projects/{id}/integrations/mattermost_slash_commands/trigger [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_integrations_mattermost_slash_commands_trigger]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project invitations create - POST /projects/{id}/invitations [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_invitations]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project keep create - POST /projects/{id}/jobs/{job_id}/artifacts/keep [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_jobs_job_id_artifacts_keep]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --job-id
    project cancel create - POST /projects/{id}/jobs/{job_id}/cancel [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_jobs_job_id_cancel]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --job-id
    project erase create - POST /projects/{id}/jobs/{job_id}/erase [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_jobs_job_id_erase]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --job-id
    project play create - POST /projects/{id}/jobs/{job_id}/play [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_jobs_job_id_play]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --job-id
    project retry create - POST /projects/{id}/jobs/{job_id}/retry [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_jobs_job_id_retry]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --job-id
    project composer create - POST /projects/{id}/packages/composer [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_packages_composer]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project upload-urls create - POST /projects/{id}/packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/packages/{conan_package_reference}/upload_urls [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_packages_conan_v1_conans_package_name_package_version_package_username_package_channel_packages_conan_package_reference_upload_urls]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-name, --package-version, --package-username, --package-channel, --conan-package-reference
    project upload-urls create-2 - POST /projects/{id}/packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/upload_urls [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_packages_conan_v1_conans_package_name_package_version_package_username_package_channel_upload_urls]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-name, --package-version, --package-username, --package-channel
    project charts create - POST /projects/{id}/packages/helm/api/{channel}/charts [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_packages_helm_api_channel_charts]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --channel
    project authorize create-4 - POST /projects/{id}/packages/helm/api/{channel}/charts/authorize [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_packages_helm_api_channel_charts_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --channel
    project bulk create - POST /projects/{id}/packages/npm/-/npm/v1/security/advisories/bulk [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_packages_npm_npm_v1_security_advisories_bulk]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project quick create - POST /projects/{id}/packages/npm/-/npm/v1/security/audits/quick [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_packages_npm_npm_v1_security_audits_quick]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project rules create - POST /projects/{id}/packages/protection/rules [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_packages_protection_rules]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project pypi create - POST /projects/{id}/packages/pypi [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_packages_pypi]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project authorize create-5 - POST /projects/{id}/packages/pypi/authorize [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_packages_pypi_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project rpm create - POST /projects/{id}/packages/rpm [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_packages_rpm]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project authorize create-6 - POST /projects/{id}/packages/rpm/authorize [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_packages_rpm_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project gems create - POST /projects/{id}/packages/rubygems/api/v1/gems [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_packages_rubygems_api_v1_gems]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project authorize create-7 - POST /projects/{id}/packages/rubygems/api/v1/gems/authorize [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_packages_rubygems_api_v1_gems_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project domains create - POST /projects/{id}/pages/domains [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_pages_domains]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project pipeline create-2 - POST /projects/{id}/pipeline [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_pipeline]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project pipeline-schedules create - POST /projects/{id}/pipeline_schedules [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_pipeline_schedules]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project play create-2 - POST /projects/{id}/pipeline_schedules/{pipeline_schedule_id}/play [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_pipeline_schedules_pipeline_schedule_id_play]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --pipeline-schedule-id
    project take-ownership create - POST /projects/{id}/pipeline_schedules/{pipeline_schedule_id}/take_ownership [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_pipeline_schedules_pipeline_schedule_id_take_ownership]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --pipeline-schedule-id
    project protected-branches create - POST /projects/{id}/protected_branches [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_protected_branches]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project protected-tags create - POST /projects/{id}/protected_tags [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_protected_tags]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project rules create-2 - POST /projects/{id}/registry/protection/tag/rules [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_registry_protection_tag_rules]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project releases create - POST /projects/{id}/releases [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_releases]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project links create - POST /projects/{id}/releases/{tag_name}/assets/links [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_releases_tag_name_assets_links]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --tag-name
    project evidence create - POST /projects/{id}/releases/{tag_name}/evidence [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_releases_tag_name_evidence]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --tag-name
    project remote-mirrors create - POST /projects/{id}/remote_mirrors [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_remote_mirrors]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project sync create - POST /projects/{id}/remote_mirrors/{mirror_id}/sync [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_remote_mirrors_mirror_id_sync]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --mirror-id
    project restore create - POST /projects/{id}/restore [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_restore]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project secure-files create - POST /projects/{id}/secure_files [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_secure_files]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project trigger create-2 - POST /projects/{id}/services/mattermost_slash_commands/trigger [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_services_mattermost_slash_commands_trigger]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project share create - POST /projects/{id}/share [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_share]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project snippets create - POST /projects/{id}/snippets [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_snippets]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project award-emoji create - POST /projects/{id}/snippets/{snippet_id}/award_emoji [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_snippets_snippet_id_award_emoji]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --snippet-id
    project award-emoji create-2 - POST /projects/{id}/snippets/{snippet_id}/notes/{note_id}/award_emoji [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_snippets_snippet_id_notes_note_id_award_emoji]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --snippet-id, --note-id
    project star create - POST /projects/{id}/star [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_star]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project statuses create - POST /projects/{id}/statuses/{sha} [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_statuses_sha]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --sha
    project state create - POST /projects/{id}/terraform/state/{name} [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_terraform_state_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --name
    project authorize create-8 - POST /projects/{id}/terraform/state/{name}/authorize [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_terraform_state_name_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --name
    project lock create - POST /projects/{id}/terraform/state/{name}/lock [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_terraform_state_name_lock]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --name
    project state-protection-rules create - POST /projects/{id}/terraform/state_protection_rules [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_terraform_state_protection_rules]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project triggers create - POST /projects/{id}/triggers [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_triggers]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project unarchive create - POST /projects/{id}/unarchive [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_unarchive]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project unstar create - POST /projects/{id}/unstar [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_unstar]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project uploads create - POST /projects/{id}/uploads [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_uploads]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project authorize create-9 - POST /projects/{id}/uploads/authorize [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_uploads_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project wikis create - POST /projects/{id}/wikis [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_wikis]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project attachments create - POST /projects/{id}/wikis/attachments [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_wikis_attachments]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project set - PUT /projects/{id} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project metric-images set - PUT /projects/{id}/alert_management_alerts/{alert_iid}/metric_images/{metric_image_id} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_alert_management_alerts_alert_iid_metric_images_metric_image_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --alert-iid, --metric-image-id
    project badges set - PUT /projects/{id}/badges/{badge_id} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_badges_badge_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --badge-id
    project clusters set - PUT /projects/{id}/clusters/{cluster_id} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_clusters_cluster_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --cluster-id
    project custom-attributes set - PUT /projects/{id}/custom_attributes/{key} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_custom_attributes_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --key
    project debian-distributions set - PUT /projects/{id}/debian_distributions/{codename} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_debian_distributions_codename]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --codename
    project deploy-keys set - PUT /projects/{id}/deploy_keys/{key_id} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_deploy_keys_key_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --key-id
    project deployments set - PUT /projects/{id}/deployments/{deployment_id} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_deployments_deployment_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --deployment-id
    project environments set - PUT /projects/{id}/environments/{environment_id} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_environments_environment_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --environment-id
    project settings set - PUT /projects/{id}/error_tracking/settings [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_error_tracking_settings]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project feature-flags set - PUT /projects/{id}/feature_flags/{feature_flag_name} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_feature_flags_feature_flag_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --feature-flag-name
    project feature-flags-user-lists set - PUT /projects/{id}/feature_flags_user_lists/{iid} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_feature_flags_user_lists_iid]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --iid
    project freeze-periods set - PUT /projects/{id}/freeze_periods/{freeze_period_id} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_freeze_periods_freeze_period_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --freeze-period-id
    project hooks set - PUT /projects/{id}/hooks/{hook_id} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_hooks_hook_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --hook-id
    project custom-headers set - PUT /projects/{id}/hooks/{hook_id}/custom_headers/{key} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_hooks_hook_id_custom_headers_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --hook-id, --key
    project apple-app-store set - PUT /projects/{id}/integrations/apple-app-store [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_apple_app_store]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project asana set - PUT /projects/{id}/integrations/asana [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_asana]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project assembla set - PUT /projects/{id}/integrations/assembla [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_assembla]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project bamboo set - PUT /projects/{id}/integrations/bamboo [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_bamboo]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project bugzilla set - PUT /projects/{id}/integrations/bugzilla [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_bugzilla]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project buildkite set - PUT /projects/{id}/integrations/buildkite [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_buildkite]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project campfire set - PUT /projects/{id}/integrations/campfire [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_campfire]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project clickup set - PUT /projects/{id}/integrations/clickup [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_clickup]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project confluence set - PUT /projects/{id}/integrations/confluence [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_confluence]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project custom-issue-tracker set - PUT /projects/{id}/integrations/custom-issue-tracker [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_custom_issue_tracker]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project datadog set - PUT /projects/{id}/integrations/datadog [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_datadog]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project diffblue-cover set - PUT /projects/{id}/integrations/diffblue-cover [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_diffblue_cover]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project discord set - PUT /projects/{id}/integrations/discord [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_discord]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project drone-ci set - PUT /projects/{id}/integrations/drone-ci [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_drone_ci]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project emails-on-push set - PUT /projects/{id}/integrations/emails-on-push [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_emails_on_push]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project ewm set - PUT /projects/{id}/integrations/ewm [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_ewm]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project external-wiki set - PUT /projects/{id}/integrations/external-wiki [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_external_wiki]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project git-guardian set - PUT /projects/{id}/integrations/git-guardian [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_git_guardian]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project github set - PUT /projects/{id}/integrations/github [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_github]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project gitlab-slack-application set - PUT /projects/{id}/integrations/gitlab-slack-application [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_gitlab_slack_application]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project google-cloud-platform-artifact-registry set - PUT /projects/{id}/integrations/google-cloud-platform-artifact-registry [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_google_cloud_platform_artifact_registry]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project google-cloud-platform-workload-identity-federation set - PUT /projects/{id}/integrations/google-cloud-platform-workload-identity-federation [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_google_cloud_platform_workload_identity_federation]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project google-play set - PUT /projects/{id}/integrations/google-play [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_google_play]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project hangouts-chat set - PUT /projects/{id}/integrations/hangouts-chat [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_hangouts_chat]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project harbor set - PUT /projects/{id}/integrations/harbor [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_harbor]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project irker set - PUT /projects/{id}/integrations/irker [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_irker]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project jenkins set - PUT /projects/{id}/integrations/jenkins [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_jenkins]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project jira set - PUT /projects/{id}/integrations/jira [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_jira]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project jira-cloud-app set - PUT /projects/{id}/integrations/jira-cloud-app [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_jira_cloud_app]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project linear set - PUT /projects/{id}/integrations/linear [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_linear]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project matrix set - PUT /projects/{id}/integrations/matrix [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_matrix]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project mattermost set - PUT /projects/{id}/integrations/mattermost [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_mattermost]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project mattermost-slash-commands set - PUT /projects/{id}/integrations/mattermost-slash-commands [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_mattermost_slash_commands]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project microsoft-teams set - PUT /projects/{id}/integrations/microsoft-teams [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_microsoft_teams]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project mock-ci set - PUT /projects/{id}/integrations/mock-ci [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_mock_ci]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project mock-monitoring set - PUT /projects/{id}/integrations/mock-monitoring [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_mock_monitoring]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project packagist set - PUT /projects/{id}/integrations/packagist [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_packagist]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project phorge set - PUT /projects/{id}/integrations/phorge [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_phorge]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project pivotaltracker set - PUT /projects/{id}/integrations/pivotaltracker [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_pivotaltracker]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project pumble set - PUT /projects/{id}/integrations/pumble [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_pumble]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project pushover set - PUT /projects/{id}/integrations/pushover [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_pushover]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project redmine set - PUT /projects/{id}/integrations/redmine [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_redmine]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project slack set - PUT /projects/{id}/integrations/slack [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_slack]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project squash-tm set - PUT /projects/{id}/integrations/squash-tm [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_squash_tm]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project teamcity set - PUT /projects/{id}/integrations/teamcity [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_teamcity]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project telegram set - PUT /projects/{id}/integrations/telegram [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_telegram]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project unify-circuit set - PUT /projects/{id}/integrations/unify-circuit [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_unify_circuit]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project webex-teams set - PUT /projects/{id}/integrations/webex-teams [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_webex_teams]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project youtrack set - PUT /projects/{id}/integrations/youtrack [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_youtrack]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project zentao set - PUT /projects/{id}/integrations/zentao [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_zentao]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project invitations set - PUT /projects/{id}/invitations/{email} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_invitations_email]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --email
    project export set - PUT /projects/{id}/packages/conan/v1/files/{package_name}/{package_version}/{package_username}/{package_channel}/{recipe_revision}/export/{file_name} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_conan_v1_files_package_name_package_version_package_username_package_channel_recipe_revision_export_file_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --file-name
    project authorize set - PUT /projects/{id}/packages/conan/v1/files/{package_name}/{package_version}/{package_username}/{package_channel}/{recipe_revision}/export/{file_name}/authorize [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_conan_v1_files_package_name_package_version_package_username_package_channel_recipe_revision_export_file_name_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --file-name
    project package set - PUT /projects/{id}/packages/conan/v1/files/{package_name}/{package_version}/{package_username}/{package_channel}/{recipe_revision}/package/{conan_package_reference}/{package_revision}/{file_name} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_conan_v1_files_package_name_package_version_package_username_package_channel_recipe_revision_package_conan_package_reference_package_revision_file_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --conan-package-reference, --package-revision, --file-name
    project authorize set-2 - PUT /projects/{id}/packages/conan/v1/files/{package_name}/{package_version}/{package_username}/{package_channel}/{recipe_revision}/package/{conan_package_reference}/{package_revision}/{file_name}/authorize [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_conan_v1_files_package_name_package_version_package_username_package_channel_recipe_revision_package_conan_package_reference_package_revision_file_name_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --conan-package-reference, --package-revision, --file-name
    project files set - PUT /projects/{id}/packages/conan/v2/conans/{package_name}/{package_version}/{package_username}/{package_channel}/revisions/{recipe_revision}/files/{file_name} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_conan_v2_conans_package_name_package_version_package_username_package_channel_revisions_recipe_revision_files_file_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --file-name
    project authorize set-3 - PUT /projects/{id}/packages/conan/v2/conans/{package_name}/{package_version}/{package_username}/{package_channel}/revisions/{recipe_revision}/files/{file_name}/authorize [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_conan_v2_conans_package_name_package_version_package_username_package_channel_revisions_recipe_revision_files_file_name_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --file-name
    project files set-2 - PUT /projects/{id}/packages/conan/v2/conans/{package_name}/{package_version}/{package_username}/{package_channel}/revisions/{recipe_revision}/packages/{conan_package_reference}/revisions/{package_revision}/files/{file_name} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_conan_v2_conans_package_name_package_version_package_username_package_channel_revisions_recipe_revision_packages_conan_package_reference_revisions_package_revision_files_file_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --conan-package-reference, --package-revision, --file-name
    project authorize set-4 - PUT /projects/{id}/packages/conan/v2/conans/{package_name}/{package_version}/{package_username}/{package_channel}/revisions/{recipe_revision}/packages/{conan_package_reference}/revisions/{package_revision}/files/{file_name}/authorize [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_conan_v2_conans_package_name_package_version_package_username_package_channel_revisions_recipe_revision_packages_conan_package_reference_revisions_package_revision_files_file_name_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --conan-package-reference, --package-revision, --file-name
    project debian set - PUT /projects/{id}/packages/debian/{file_name} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_debian_file_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --file-name
    project authorize set-5 - PUT /projects/{id}/packages/debian/{file_name}/authorize [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_debian_file_name_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --file-name
    project (*path set - PUT /projects/{id}/packages/generic/{package_name}/*package_version/(*path/){file_name} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_generic_package_name_package_version_path__file_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-name, --file-name
    project authorize set-6 - PUT /projects/{id}/packages/generic/{package_name}/*package_version/(*path/){file_name}/authorize [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_generic_package_name_package_version_path__file_name_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-name, --file-name
    project *path set - PUT /projects/{id}/packages/maven/*path/{file_name} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_maven_path_file_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --file-name
    project authorize set-7 - PUT /projects/{id}/packages/maven/*path/{file_name}/authorize [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_maven_path_file_name_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --file-name
    project (*path set-2 - PUT /projects/{id}/packages/ml_models/{model_version_id}/files/(*path/){file_name} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_ml_models_model_version_id_files_path__file_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --model-version-id, --file-name
    project authorize set-8 - PUT /projects/{id}/packages/ml_models/{model_version_id}/files/(*path/){file_name}/authorize [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_ml_models_model_version_id_files_path__file_name_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --model-version-id, --file-name
    project dist-tags set - PUT /projects/{id}/packages/npm/-/package/*package_name/dist-tags/{tag} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_npm_package_package_name_dist_tags_tag]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --tag
    project npm set - PUT /projects/{id}/packages/npm/{package_name} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_npm_package_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --package-name
    project nuget set - PUT /projects/{id}/packages/nuget [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_nuget]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project authorize set-9 - PUT /projects/{id}/packages/nuget/authorize [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_nuget_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project symbolpackage set - PUT /projects/{id}/packages/nuget/symbolpackage [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_nuget_symbolpackage]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project authorize set-10 - PUT /projects/{id}/packages/nuget/symbolpackage/authorize [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_nuget_symbolpackage_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project v2 set - PUT /projects/{id}/packages/nuget/v2 [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_nuget_v2]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project authorize set-11 - PUT /projects/{id}/packages/nuget/v2/authorize [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_nuget_v2_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project file set - PUT /projects/{id}/packages/terraform/modules/{module_name}/{module_system}/*module_version/file [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_terraform_modules_module_name_module_system_module_version_file]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --module-name, --module-system
    project authorize set-12 - PUT /projects/{id}/packages/terraform/modules/{module_name}/{module_system}/*module_version/file/authorize [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_packages_terraform_modules_module_name_module_system_module_version_file_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --module-name, --module-system
    project domains set - PUT /projects/{id}/pages/domains/{domain} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_pages_domains_domain]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --domain
    project verify set - PUT /projects/{id}/pages/domains/{domain}/verify [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_pages_domains_domain_verify]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --domain
    project pipeline-schedules set - PUT /projects/{id}/pipeline_schedules/{pipeline_schedule_id} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_pipeline_schedules_pipeline_schedule_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --pipeline-schedule-id
    project releases set - PUT /projects/{id}/releases/{tag_name} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_releases_tag_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --tag-name
    project links set - PUT /projects/{id}/releases/{tag_name}/assets/links/{link_id} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_releases_tag_name_assets_links_link_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --tag-name, --link-id
    project remote-mirrors set - PUT /projects/{id}/remote_mirrors/{mirror_id} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_remote_mirrors_mirror_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --mirror-id
    project apple-app-store set-2 - PUT /projects/{id}/services/apple-app-store [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_apple_app_store]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project asana set-2 - PUT /projects/{id}/services/asana [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_asana]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project assembla set-2 - PUT /projects/{id}/services/assembla [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_assembla]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project bamboo set-2 - PUT /projects/{id}/services/bamboo [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_bamboo]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project bugzilla set-2 - PUT /projects/{id}/services/bugzilla [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_bugzilla]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project buildkite set-2 - PUT /projects/{id}/services/buildkite [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_buildkite]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project campfire set-2 - PUT /projects/{id}/services/campfire [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_campfire]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project clickup set-2 - PUT /projects/{id}/services/clickup [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_clickup]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project confluence set-2 - PUT /projects/{id}/services/confluence [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_confluence]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project custom-issue-tracker set-2 - PUT /projects/{id}/services/custom-issue-tracker [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_custom_issue_tracker]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project datadog set-2 - PUT /projects/{id}/services/datadog [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_datadog]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project diffblue-cover set-2 - PUT /projects/{id}/services/diffblue-cover [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_diffblue_cover]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project discord set-2 - PUT /projects/{id}/services/discord [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_discord]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project drone-ci set-2 - PUT /projects/{id}/services/drone-ci [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_drone_ci]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project emails-on-push set-2 - PUT /projects/{id}/services/emails-on-push [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_emails_on_push]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project ewm set-2 - PUT /projects/{id}/services/ewm [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_ewm]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project external-wiki set-2 - PUT /projects/{id}/services/external-wiki [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_external_wiki]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project git-guardian set-2 - PUT /projects/{id}/services/git-guardian [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_git_guardian]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project github set-2 - PUT /projects/{id}/services/github [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_github]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project gitlab-slack-application set-2 - PUT /projects/{id}/services/gitlab-slack-application [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_gitlab_slack_application]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project google-cloud-platform-artifact-registry set-2 - PUT /projects/{id}/services/google-cloud-platform-artifact-registry [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_google_cloud_platform_artifact_registry]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project google-cloud-platform-workload-identity-federation set-2 - PUT /projects/{id}/services/google-cloud-platform-workload-identity-federation [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_google_cloud_platform_workload_identity_federation]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project google-play set-2 - PUT /projects/{id}/services/google-play [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_google_play]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project hangouts-chat set-2 - PUT /projects/{id}/services/hangouts-chat [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_hangouts_chat]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project harbor set-2 - PUT /projects/{id}/services/harbor [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_harbor]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project irker set-2 - PUT /projects/{id}/services/irker [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_irker]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project jenkins set-2 - PUT /projects/{id}/services/jenkins [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_jenkins]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project jira set-2 - PUT /projects/{id}/services/jira [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_jira]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project jira-cloud-app set-2 - PUT /projects/{id}/services/jira-cloud-app [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_jira_cloud_app]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project linear set-2 - PUT /projects/{id}/services/linear [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_linear]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project matrix set-2 - PUT /projects/{id}/services/matrix [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_matrix]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project mattermost set-2 - PUT /projects/{id}/services/mattermost [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_mattermost]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project mattermost-slash-commands set-2 - PUT /projects/{id}/services/mattermost-slash-commands [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_mattermost_slash_commands]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project microsoft-teams set-2 - PUT /projects/{id}/services/microsoft-teams [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_microsoft_teams]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project mock-ci set-2 - PUT /projects/{id}/services/mock-ci [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_mock_ci]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project mock-monitoring set-2 - PUT /projects/{id}/services/mock-monitoring [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_mock_monitoring]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project packagist set-2 - PUT /projects/{id}/services/packagist [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_packagist]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project phorge set-2 - PUT /projects/{id}/services/phorge [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_phorge]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project pivotaltracker set-2 - PUT /projects/{id}/services/pivotaltracker [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_pivotaltracker]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project pumble set-2 - PUT /projects/{id}/services/pumble [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_pumble]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project pushover set-2 - PUT /projects/{id}/services/pushover [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_pushover]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project redmine set-2 - PUT /projects/{id}/services/redmine [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_redmine]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project slack set-2 - PUT /projects/{id}/services/slack [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_slack]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project squash-tm set-2 - PUT /projects/{id}/services/squash-tm [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_squash_tm]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project teamcity set-2 - PUT /projects/{id}/services/teamcity [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_teamcity]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project telegram set-2 - PUT /projects/{id}/services/telegram [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_telegram]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project unify-circuit set-2 - PUT /projects/{id}/services/unify-circuit [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_unify_circuit]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project webex-teams set-2 - PUT /projects/{id}/services/webex-teams [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_webex_teams]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project youtrack set-2 - PUT /projects/{id}/services/youtrack [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_youtrack]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project zentao set-2 - PUT /projects/{id}/services/zentao [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_zentao]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project snippets set - PUT /projects/{id}/snippets/{snippet_id} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_snippets_snippet_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --snippet-id
    project transfer set - PUT /projects/{id}/transfer [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_transfer]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    project triggers set - PUT /projects/{id}/triggers/{trigger_id} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_triggers_trigger_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --trigger-id
    project wikis set - PUT /projects/{id}/wikis/{slug} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_wikis_slug]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --slug
    group list - List visible GitLab groups [intent=etl availability=implemented stream=groups]; flags: --search
    group view - View one GitLab group [intent=direct_read availability=implemented]; notes: Bounded direct read of one group; response is recursively redacted before output.; flags: --id
    group delete - DELETE /groups/{id} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group debian-distributions delete - DELETE /groups/{id}/-/debian_distributions/{codename} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_debian_distributions_codename]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --codename
    group dist-tags delete - DELETE /groups/{id}/-/packages/npm/-/package/*package_name/dist-tags/{tag} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_packages_npm_package_package_name_dist_tags_tag]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --tag
    group badges delete - DELETE /groups/{id}/badges/{badge_id} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_badges_badge_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --badge-id
    group clusters delete - DELETE /groups/{id}/clusters/{cluster_id} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_clusters_cluster_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --cluster-id
    group custom-attributes delete - DELETE /groups/{id}/custom_attributes/{key} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_custom_attributes_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --key
    group cache delete - DELETE /groups/{id}/dependency_proxy/cache [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_dependency_proxy_cache]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group award-emoji delete - DELETE /groups/{id}/epics/{epic_iid}/award_emoji/{award_id} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_epics_epic_iid_award_emoji_award_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --epic-iid, --award-id
    group award-emoji delete-2 - DELETE /groups/{id}/epics/{epic_iid}/notes/{note_id}/award_emoji/{award_id} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_epics_epic_iid_notes_note_id_award_emoji_award_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --epic-iid, --note-id, --award-id
    group integrations delete - DELETE /groups/{id}/integrations/{slug} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_integrations_slug]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --slug
    group invitations delete - DELETE /groups/{id}/invitations/{email} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_invitations_email]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --email
    group share delete - DELETE /groups/{id}/share/{group_id} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_share_group_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --group-id
    group shared-projects delete - DELETE /groups/{id}/shared_projects/{project_id} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_shared_projects_project_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --project-id
    group ssh-certificates delete - DELETE /groups/{id}/ssh_certificates/{ssh_certificates_id} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_ssh_certificates_ssh_certificates_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --ssh-certificates-id
    group uploads delete - DELETE /groups/{id}/uploads/{secret}/{filename} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_uploads_secret_filename]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --secret, --filename
    group uploads delete-2 - DELETE /groups/{id}/uploads/{upload_id} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_uploads_upload_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --upload-id
    group wikis delete - DELETE /groups/{id}/wikis/{slug} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_wikis_slug]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --slug
    group *package-name download - GET /group/{id}/-/packages/composer/*package_name [intent=direct_read availability=implemented]; flags: --id
    group p download - GET /group/{id}/-/packages/composer/p/{sha} [intent=direct_read availability=implemented]; flags: --id, --sha
    group *package-name download-2 - GET /group/{id}/-/packages/composer/p2/*package_name [intent=direct_read availability=implemented]; flags: --id
    group packages download - GET /group/{id}/-/packages/composer/packages [intent=direct_read availability=implemented]; flags: --id
    group )search view - GET /groups/{id}/(-/)search [intent=direct_read availability=implemented]; flags: --id
    group debian-distributions view - GET /groups/{id}/-/debian_distributions [intent=direct_read availability=implemented]; flags: --id
    group debian-distributions view-2 - GET /groups/{id}/-/debian_distributions/{codename} [intent=direct_read availability=implemented]; flags: --id, --codename
    group key.asc view - GET /groups/{id}/-/debian_distributions/{codename}/key.asc [intent=direct_read availability=implemented]; flags: --id, --codename
    group InRelease download - GET /groups/{id}/-/packages/debian/dists/*distribution/InRelease [intent=direct_read availability=implemented]; flags: --id
    group Release download - GET /groups/{id}/-/packages/debian/dists/*distribution/Release [intent=direct_read availability=implemented]; flags: --id
    group Release.gpg download - GET /groups/{id}/-/packages/debian/dists/*distribution/Release.gpg [intent=direct_read availability=implemented]; flags: --id
    group Packages download - GET /groups/{id}/-/packages/debian/dists/*distribution/{component}/binary-{architecture}/Packages [intent=direct_read availability=implemented]; flags: --id, --component, --architecture
    group SHA256 download - GET /groups/{id}/-/packages/debian/dists/*distribution/{component}/binary-{architecture}/by-hash/SHA256/{file_sha256} [intent=direct_read availability=implemented]; flags: --id, --component, --architecture, --file-sha256
    group Packages download-2 - GET /groups/{id}/-/packages/debian/dists/*distribution/{component}/debian-installer/binary-{architecture}/Packages [intent=direct_read availability=implemented]; flags: --id, --component, --architecture
    group SHA256 download-2 - GET /groups/{id}/-/packages/debian/dists/*distribution/{component}/debian-installer/binary-{architecture}/by-hash/SHA256/{file_sha256} [intent=direct_read availability=implemented]; flags: --id, --component, --architecture, --file-sha256
    group Sources download - GET /groups/{id}/-/packages/debian/dists/*distribution/{component}/source/Sources [intent=direct_read availability=implemented]; flags: --id, --component
    group SHA256 download-3 - GET /groups/{id}/-/packages/debian/dists/*distribution/{component}/source/by-hash/SHA256/{file_sha256} [intent=direct_read availability=implemented]; flags: --id, --component, --file-sha256
    group pool download - GET /groups/{id}/-/packages/debian/pool/{distribution}/{project_id}/{letter}/{package_name}/{package_version}/{file_name} [intent=direct_read availability=implemented]; flags: --id, --distribution, --project-id, --letter, --package-name, --package-version, --file-name
    group *path download - GET /groups/{id}/-/packages/maven/*path/{file_name} [intent=direct_read availability=implemented]; flags: --id, --file-name
    group *package-name download-3 - GET /groups/{id}/-/packages/npm/*package_name [intent=direct_read availability=implemented]; flags: --id
    group dist-tags download - GET /groups/{id}/-/packages/npm/-/package/*package_name/dist-tags [intent=direct_read availability=implemented]; flags: --id
    group index download - GET /groups/{id}/-/packages/nuget/index [intent=direct_read availability=implemented]; flags: --id
    group *package-version download - GET /groups/{id}/-/packages/nuget/metadata/*package_name/*package_version [intent=direct_read availability=implemented]; flags: --id
    group index download-2 - GET /groups/{id}/-/packages/nuget/metadata/*package_name/index [intent=direct_read availability=implemented]; flags: --id
    group query download - GET /groups/{id}/-/packages/nuget/query [intent=direct_read availability=implemented]; flags: --id
    group *same-file-name download - GET /groups/{id}/-/packages/nuget/symbolfiles/*file_name/*signature/*same_file_name [intent=direct_read availability=implemented]; flags: --id
    group v2 download - GET /groups/{id}/-/packages/nuget/v2 [intent=direct_read availability=implemented]; flags: --id
    group $metadata download - GET /groups/{id}/-/packages/nuget/v2/$metadata [intent=direct_read availability=implemented]; flags: --id
    group *file-identifier download - GET /groups/{id}/-/packages/pypi/files/{sha256}/*file_identifier [intent=direct_read availability=implemented]; flags: --id, --sha256
    group simple download - GET /groups/{id}/-/packages/pypi/simple [intent=direct_read availability=implemented]; flags: --id
    group *package-name download-4 - GET /groups/{id}/-/packages/pypi/simple/*package_name [intent=direct_read availability=implemented]; flags: --id
    group audit-events view - GET /groups/{id}/audit_events [intent=direct_read availability=implemented]; flags: --id
    group audit-events view-2 - GET /groups/{id}/audit_events/{audit_event_id} [intent=direct_read availability=implemented]; flags: --id, --audit-event-id
    group avatar download - GET /groups/{id}/avatar [intent=direct_read availability=implemented]; flags: --id
    group badges view - GET /groups/{id}/badges [intent=direct_read availability=implemented]; flags: --id
    group render view - GET /groups/{id}/badges/render [intent=direct_read availability=implemented]; flags: --id
    group badges view-2 - GET /groups/{id}/badges/{badge_id} [intent=direct_read availability=implemented]; flags: --id, --badge-id
    group clusters view - GET /groups/{id}/clusters [intent=direct_read availability=implemented]; flags: --id
    group clusters view-2 - GET /groups/{id}/clusters/{cluster_id} [intent=direct_read availability=implemented]; flags: --id, --cluster-id
    group custom-attributes view - GET /groups/{id}/custom_attributes [intent=direct_read availability=implemented]; flags: --id
    group custom-attributes view-2 - GET /groups/{id}/custom_attributes/{key} [intent=direct_read availability=implemented]; flags: --id, --key
    group descendant-groups view - GET /groups/{id}/descendant_groups [intent=direct_read availability=implemented]; flags: --id
    group award-emoji view - GET /groups/{id}/epics/{epic_iid}/award_emoji [intent=direct_read availability=implemented]; flags: --id, --epic-iid
    group award-emoji view-2 - GET /groups/{id}/epics/{epic_iid}/award_emoji/{award_id} [intent=direct_read availability=implemented]; flags: --id, --epic-iid, --award-id
    group award-emoji view-3 - GET /groups/{id}/epics/{epic_iid}/notes/{note_id}/award_emoji [intent=direct_read availability=implemented]; flags: --id, --epic-iid, --note-id
    group award-emoji view-4 - GET /groups/{id}/epics/{epic_iid}/notes/{note_id}/award_emoji/{award_id} [intent=direct_read availability=implemented]; flags: --id, --epic-iid, --note-id, --award-id
    group download download - GET /groups/{id}/export/download [intent=direct_read availability=implemented]; flags: --id
    group download download-2 - GET /groups/{id}/export_relations/download [intent=direct_read availability=implemented]; flags: --id
    group status view - GET /groups/{id}/export_relations/status [intent=direct_read availability=implemented]; flags: --id
    group shared view - GET /groups/{id}/groups/shared [intent=direct_read availability=implemented]; flags: --id
    group integrations view - GET /groups/{id}/integrations [intent=direct_read availability=implemented]; flags: --id
    group integrations view-2 - GET /groups/{id}/integrations/{slug} [intent=direct_read availability=implemented]; flags: --id, --slug
    group invitations view - GET /groups/{id}/invitations [intent=direct_read availability=implemented]; flags: --id
    group invited-groups view - GET /groups/{id}/invited_groups [intent=direct_read availability=implemented]; flags: --id
    group packages view - GET /groups/{id}/packages [intent=direct_read availability=implemented]; flags: --id
    group placeholder-reassignments view - GET /groups/{id}/placeholder_reassignments [intent=direct_read availability=implemented]; flags: --id
    group view-2 - GET /groups/{id}/projects [intent=direct_read availability=implemented]; flags: --id
    group shared view-2 - GET /groups/{id}/projects/shared [intent=direct_read availability=implemented]; flags: --id
    group provisioned-users view - GET /groups/{id}/provisioned_users [intent=direct_read availability=implemented]; flags: --id
    group repositories view - GET /groups/{id}/registry/repositories [intent=direct_read availability=implemented]; flags: --id
    group releases view - GET /groups/{id}/releases [intent=direct_read availability=implemented]; flags: --id
    group saml-users view - GET /groups/{id}/saml_users [intent=direct_read availability=implemented]; flags: --id
    group ssh-certificates view - GET /groups/{id}/ssh_certificates [intent=direct_read availability=implemented]; flags: --id
    group subgroups view - GET /groups/{id}/subgroups [intent=direct_read availability=implemented]; flags: --id
    group transfer-locations view - GET /groups/{id}/transfer_locations [intent=direct_read availability=implemented]; flags: --id
    group uploads download - GET /groups/{id}/uploads [intent=direct_read availability=implemented]; flags: --id
    group uploads download-2 - GET /groups/{id}/uploads/{secret}/{filename} [intent=direct_read availability=implemented]; flags: --id, --secret, --filename
    group uploads download-3 - GET /groups/{id}/uploads/{upload_id} [intent=direct_read availability=implemented]; flags: --id, --upload-id
    group wikis view - GET /groups/{id}/wikis [intent=direct_read availability=implemented]; flags: --id
    group wikis view-2 - GET /groups/{id}/wikis/{slug} [intent=direct_read availability=implemented]; flags: --id, --slug
    group view-3 - GET /projects/{id}/groups [intent=direct_read availability=implemented]; flags: --id
    group invited-groups view-2 - GET /projects/{id}/invited_groups [intent=direct_read availability=implemented]; flags: --id
    group resource-groups view - GET /projects/{id}/resource_groups [intent=direct_read availability=implemented]; flags: --id
    group resource-groups view-2 - GET /projects/{id}/resource_groups/{key} [intent=direct_read availability=implemented]; flags: --id, --key
    group current-job view - GET /projects/{id}/resource_groups/{key}/current_job [intent=direct_read availability=implemented]; flags: --id, --key
    group upcoming-jobs view - GET /projects/{id}/resource_groups/{key}/upcoming_jobs [intent=direct_read availability=implemented]; flags: --id, --key
    group create - POST /groups [intent=reverse_etl availability=implemented write=post_api_v4_groups]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    group import create - POST /groups/import [intent=reverse_etl availability=implemented write=post_api_v4_groups_import]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    group authorize create - POST /groups/import/authorize [intent=reverse_etl availability=implemented write=post_api_v4_groups_import_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    group debian-distributions create - POST /groups/{id}/-/debian_distributions [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_debian_distributions]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group bulk create - POST /groups/{id}/-/packages/npm/-/npm/v1/security/advisories/bulk [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_packages_npm_npm_v1_security_advisories_bulk]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group quick create - POST /groups/{id}/-/packages/npm/-/npm/v1/security/audits/quick [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_packages_npm_npm_v1_security_audits_quick]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group archive create - POST /groups/{id}/archive [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_archive]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group badges create - POST /groups/{id}/badges [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_badges]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group user create - POST /groups/{id}/clusters/user [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_clusters_user]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group award-emoji create - POST /groups/{id}/epics/{epic_iid}/award_emoji [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_epics_epic_iid_award_emoji]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --epic-iid
    group award-emoji create-2 - POST /groups/{id}/epics/{epic_iid}/notes/{note_id}/award_emoji [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_epics_epic_iid_notes_note_id_award_emoji]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --epic-iid, --note-id
    group export create - POST /groups/{id}/export [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_export]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group export-relations create - POST /groups/{id}/export_relations [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_export_relations]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group invitations create - POST /groups/{id}/invitations [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_invitations]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group ldap-sync create - POST /groups/{id}/ldap_sync [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_ldap_sync]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group placeholder-reassignments create - POST /groups/{id}/placeholder_reassignments [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_placeholder_reassignments]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group authorize create-2 - POST /groups/{id}/placeholder_reassignments/authorize [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_placeholder_reassignments_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group create-2 - POST /groups/{id}/projects/{project_id} [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_projects_project_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --project-id
    group restore create - POST /groups/{id}/restore [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_restore]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group share create - POST /groups/{id}/share [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_share]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group ssh-certificates create - POST /groups/{id}/ssh_certificates [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_ssh_certificates]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group transfer create - POST /groups/{id}/transfer [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_transfer]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group transfer-to-organization create - POST /groups/{id}/transfer_to_organization [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_transfer_to_organization]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group unarchive create - POST /groups/{id}/unarchive [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_unarchive]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group uploads create - POST /groups/{id}/uploads [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_uploads]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group authorize create-3 - POST /groups/{id}/uploads/authorize [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_uploads_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group wikis create - POST /groups/{id}/wikis [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_wikis]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group attachments create - POST /groups/{id}/wikis/attachments [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_wikis_attachments]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group set - PUT /groups/{id} [intent=reverse_etl availability=implemented write=put_api_v4_groups_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group debian-distributions set - PUT /groups/{id}/-/debian_distributions/{codename} [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_debian_distributions_codename]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --codename
    group dist-tags set - PUT /groups/{id}/-/packages/npm/-/package/*package_name/dist-tags/{tag} [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_packages_npm_package_package_name_dist_tags_tag]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --tag
    group badges set - PUT /groups/{id}/badges/{badge_id} [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_badges_badge_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --badge-id
    group clusters set - PUT /groups/{id}/clusters/{cluster_id} [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_clusters_cluster_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --cluster-id
    group custom-attributes set - PUT /groups/{id}/custom_attributes/{key} [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_custom_attributes_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --key
    group apple-app-store set - PUT /groups/{id}/integrations/apple-app-store [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_apple_app_store]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group asana set - PUT /groups/{id}/integrations/asana [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_asana]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group assembla set - PUT /groups/{id}/integrations/assembla [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_assembla]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group bamboo set - PUT /groups/{id}/integrations/bamboo [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_bamboo]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group bugzilla set - PUT /groups/{id}/integrations/bugzilla [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_bugzilla]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group buildkite set - PUT /groups/{id}/integrations/buildkite [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_buildkite]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group campfire set - PUT /groups/{id}/integrations/campfire [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_campfire]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group clickup set - PUT /groups/{id}/integrations/clickup [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_clickup]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group confluence set - PUT /groups/{id}/integrations/confluence [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_confluence]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group custom-issue-tracker set - PUT /groups/{id}/integrations/custom-issue-tracker [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_custom_issue_tracker]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group datadog set - PUT /groups/{id}/integrations/datadog [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_datadog]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group diffblue-cover set - PUT /groups/{id}/integrations/diffblue-cover [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_diffblue_cover]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group discord set - PUT /groups/{id}/integrations/discord [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_discord]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group drone-ci set - PUT /groups/{id}/integrations/drone-ci [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_drone_ci]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group emails-on-push set - PUT /groups/{id}/integrations/emails-on-push [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_emails_on_push]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group ewm set - PUT /groups/{id}/integrations/ewm [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_ewm]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group external-wiki set - PUT /groups/{id}/integrations/external-wiki [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_external_wiki]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group git-guardian set - PUT /groups/{id}/integrations/git-guardian [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_git_guardian]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group github set - PUT /groups/{id}/integrations/github [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_github]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group gitlab-slack-application set - PUT /groups/{id}/integrations/gitlab-slack-application [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_gitlab_slack_application]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group google-cloud-platform-artifact-registry set - PUT /groups/{id}/integrations/google-cloud-platform-artifact-registry [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_google_cloud_platform_artifact_registry]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group google-cloud-platform-workload-identity-federation set - PUT /groups/{id}/integrations/google-cloud-platform-workload-identity-federation [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_google_cloud_platform_workload_identity_federation]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group google-play set - PUT /groups/{id}/integrations/google-play [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_google_play]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group hangouts-chat set - PUT /groups/{id}/integrations/hangouts-chat [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_hangouts_chat]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group harbor set - PUT /groups/{id}/integrations/harbor [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_harbor]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group irker set - PUT /groups/{id}/integrations/irker [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_irker]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group jenkins set - PUT /groups/{id}/integrations/jenkins [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_jenkins]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group jira set - PUT /groups/{id}/integrations/jira [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_jira]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group jira-cloud-app set - PUT /groups/{id}/integrations/jira-cloud-app [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_jira_cloud_app]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group linear set - PUT /groups/{id}/integrations/linear [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_linear]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group matrix set - PUT /groups/{id}/integrations/matrix [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_matrix]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group mattermost set - PUT /groups/{id}/integrations/mattermost [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_mattermost]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group mattermost-slash-commands set - PUT /groups/{id}/integrations/mattermost-slash-commands [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_mattermost_slash_commands]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group microsoft-teams set - PUT /groups/{id}/integrations/microsoft-teams [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_microsoft_teams]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group mock-ci set - PUT /groups/{id}/integrations/mock-ci [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_mock_ci]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group mock-monitoring set - PUT /groups/{id}/integrations/mock-monitoring [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_mock_monitoring]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group packagist set - PUT /groups/{id}/integrations/packagist [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_packagist]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group phorge set - PUT /groups/{id}/integrations/phorge [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_phorge]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group pivotaltracker set - PUT /groups/{id}/integrations/pivotaltracker [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_pivotaltracker]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group pumble set - PUT /groups/{id}/integrations/pumble [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_pumble]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group pushover set - PUT /groups/{id}/integrations/pushover [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_pushover]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group redmine set - PUT /groups/{id}/integrations/redmine [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_redmine]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group slack set - PUT /groups/{id}/integrations/slack [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_slack]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group squash-tm set - PUT /groups/{id}/integrations/squash-tm [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_squash_tm]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group teamcity set - PUT /groups/{id}/integrations/teamcity [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_teamcity]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group telegram set - PUT /groups/{id}/integrations/telegram [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_telegram]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group unify-circuit set - PUT /groups/{id}/integrations/unify-circuit [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_unify_circuit]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group webex-teams set - PUT /groups/{id}/integrations/webex-teams [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_webex_teams]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group youtrack set - PUT /groups/{id}/integrations/youtrack [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_youtrack]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group zentao set - PUT /groups/{id}/integrations/zentao [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_zentao]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    group invitations set - PUT /groups/{id}/invitations/{email} [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_invitations_email]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --email
    group wikis set - PUT /groups/{id}/wikis/{slug} [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_wikis_slug]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --slug
    group resource-groups set - PUT /projects/{id}/resource_groups/{key} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_resource_groups_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --key
    user list - List GitLab users visible to the token [intent=etl availability=implemented stream=users]; flags: --search, --username
    user events - View events for a GitLab user [intent=direct_read availability=implemented]; notes: Bounded direct read of a user event page; response is recursively redacted before output.; flags: --id
    issue list - List GitLab issues visible to the token [intent=etl availability=implemented stream=issues]; flags: --state, --assignee-username, --label
    issue view - View issue details [intent=direct_read availability=implemented]; notes: Bounded direct read of one project issue; response is recursively redacted before output.; flags: --project-id, --issue-iid
    issue create - Create an issue [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would create a visible issue in a GitLab project.; notes: No GitLab write action is declared yet.
    issue update - Update an issue [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would mutate issue title, description, labels, assignees, or state.; notes: No GitLab write action is declared yet.
    issue close - Close an issue [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would change issue state in a GitLab project.; notes: No GitLab write action is declared yet.
    issue reopen - Reopen an issue [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would change issue state in a GitLab project.; notes: No GitLab write action is declared yet.
    issue delete - Delete an issue [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; destructive writes need an explicit policy and typed confirmation before dispatch.; risk: Deletes project data and may be irreversible.; notes: Destructive issue deletion is not exposed by this metadata slice.
    issue note - Add a note to an issue [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would create a visible note on an issue.; notes: No GitLab write action is declared yet.
    issue delete-2 - DELETE /projects/{id}/issues/{issue_iid} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_issues_issue_iid]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid
    issue award-emoji delete - DELETE /projects/{id}/issues/{issue_iid}/award_emoji/{award_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_issues_issue_iid_award_emoji_award_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid, --award-id
    issue links delete - DELETE /projects/{id}/issues/{issue_iid}/links/{issue_link_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_issues_issue_iid_links_issue_link_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid, --issue-link-id
    issue metric-images delete - DELETE /projects/{id}/issues/{issue_iid}/metric_images/{metric_image_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_issues_issue_iid_metric_images_metric_image_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid, --metric-image-id
    issue award-emoji delete-2 - DELETE /projects/{id}/issues/{issue_iid}/notes/{note_id}/award_emoji/{award_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_issues_issue_iid_notes_note_id_award_emoji_award_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid, --note-id, --award-id
    issue view-2 - GET /groups/{id}/issues [intent=direct_read availability=implemented]; flags: --id
    issue issues-statistics view - GET /groups/{id}/issues_statistics [intent=direct_read availability=implemented]; flags: --id
    issue view-3 - GET /issues/{id} [intent=direct_read availability=implemented]; flags: --id
    issue issues-statistics view-2 - GET /issues_statistics [intent=direct_read availability=implemented]
    issue view-4 - GET /projects/{id}/issues [intent=direct_read availability=implemented]; flags: --id
    issue resource-milestone-events view - GET /projects/{id}/issues/{eventable_id}/resource_milestone_events [intent=direct_read availability=implemented]; flags: --id, --eventable-id
    issue resource-milestone-events view-2 - GET /projects/{id}/issues/{eventable_id}/resource_milestone_events/{event_id} [intent=direct_read availability=implemented]; flags: --id, --eventable-id, --event-id
    issue award-emoji view - GET /projects/{id}/issues/{issue_iid}/award_emoji [intent=direct_read availability=implemented]; flags: --id, --issue-iid
    issue award-emoji view-2 - GET /projects/{id}/issues/{issue_iid}/award_emoji/{award_id} [intent=direct_read availability=implemented]; flags: --id, --issue-iid, --award-id
    issue closed-by view - GET /projects/{id}/issues/{issue_iid}/closed_by [intent=direct_read availability=implemented]; flags: --id, --issue-iid
    issue links view - GET /projects/{id}/issues/{issue_iid}/links [intent=direct_read availability=implemented]; flags: --id, --issue-iid
    issue links view-2 - GET /projects/{id}/issues/{issue_iid}/links/{issue_link_id} [intent=direct_read availability=implemented]; flags: --id, --issue-iid, --issue-link-id
    issue metric-images view - GET /projects/{id}/issues/{issue_iid}/metric_images [intent=direct_read availability=implemented]; flags: --id, --issue-iid
    issue award-emoji view-3 - GET /projects/{id}/issues/{issue_iid}/notes/{note_id}/award_emoji [intent=direct_read availability=implemented]; flags: --id, --issue-iid, --note-id
    issue award-emoji view-4 - GET /projects/{id}/issues/{issue_iid}/notes/{note_id}/award_emoji/{award_id} [intent=direct_read availability=implemented]; flags: --id, --issue-iid, --note-id, --award-id
    issue participants view - GET /projects/{id}/issues/{issue_iid}/participants [intent=direct_read availability=implemented]; flags: --id, --issue-iid
    issue time-stats view - GET /projects/{id}/issues/{issue_iid}/time_stats [intent=direct_read availability=implemented]; flags: --id, --issue-iid
    issue user-agent-detail view - GET /projects/{id}/issues/{issue_iid}/user_agent_detail [intent=direct_read availability=implemented]; flags: --id, --issue-iid
    issue issues-statistics view-3 - GET /projects/{id}/issues_statistics [intent=direct_read availability=implemented]; flags: --id
    issue create-2 - POST /projects/{id}/issues [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_issues]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    issue add-spent-time create - POST /projects/{id}/issues/{issue_iid}/add_spent_time [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_issues_issue_iid_add_spent_time]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid
    issue award-emoji create - POST /projects/{id}/issues/{issue_iid}/award_emoji [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_issues_issue_iid_award_emoji]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid
    issue clone create - POST /projects/{id}/issues/{issue_iid}/clone [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_issues_issue_iid_clone]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid
    issue links create - POST /projects/{id}/issues/{issue_iid}/links [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_issues_issue_iid_links]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid
    issue metric-images create - POST /projects/{id}/issues/{issue_iid}/metric_images [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_issues_issue_iid_metric_images]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid
    issue authorize create - POST /projects/{id}/issues/{issue_iid}/metric_images/authorize [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_issues_issue_iid_metric_images_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid
    issue move create - POST /projects/{id}/issues/{issue_iid}/move [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_issues_issue_iid_move]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid
    issue award-emoji create-2 - POST /projects/{id}/issues/{issue_iid}/notes/{note_id}/award_emoji [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_issues_issue_iid_notes_note_id_award_emoji]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid, --note-id
    issue reset-spent-time create - POST /projects/{id}/issues/{issue_iid}/reset_spent_time [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_issues_issue_iid_reset_spent_time]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid
    issue reset-time-estimate create - POST /projects/{id}/issues/{issue_iid}/reset_time_estimate [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_issues_issue_iid_reset_time_estimate]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid
    issue time-estimate create - POST /projects/{id}/issues/{issue_iid}/time_estimate [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_issues_issue_iid_time_estimate]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid
    issue set - PUT /projects/{id}/issues/{issue_iid} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_issues_issue_iid]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid
    issue metric-images set - PUT /projects/{id}/issues/{issue_iid}/metric_images/{metric_image_id} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_issues_issue_iid_metric_images_metric_image_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid, --metric-image-id
    issue reorder set - PUT /projects/{id}/issues/{issue_iid}/reorder [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_issues_issue_iid_reorder]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --issue-iid
    mr list - List merge requests [intent=etl availability=planned]; notes: Merge request stream coverage belongs to a future stream expansion or operation-ledger lane.
    mr view - View merge request details [intent=direct_read availability=planned]; notes: Single merge-request lookup requires bounded direct-read support.
    mr create - Create a merge request [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would create a visible merge request in a project.; notes: No GitLab write action is declared yet.
    mr update - Update a merge request [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would mutate merge request metadata or state.; notes: No GitLab write action is declared yet.
    mr merge - Merge a merge request [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; destructive or deployment-adjacent writes need explicit policy and typed confirmation.; risk: Merges code into the target branch and can trigger CI/CD or deployments.; notes: Not exposed until sensitive/admin policy and typed confirmation are implemented.
    mr approve - Approve a merge request [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would record a review approval on a merge request.; notes: No GitLab write action is declared yet.
    mr merge-requests delete - DELETE /projects/{id}/merge_requests/{merge_request_iid} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_merge_requests_merge_request_iid]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid
    mr award-emoji delete - DELETE /projects/{id}/merge_requests/{merge_request_iid}/award_emoji/{award_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_merge_requests_merge_request_iid_award_emoji_award_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid, --award-id
    mr context-commits delete - DELETE /projects/{id}/merge_requests/{merge_request_iid}/context_commits [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_merge_requests_merge_request_iid_context_commits]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid
    mr draft-notes delete - DELETE /projects/{id}/merge_requests/{merge_request_iid}/draft_notes/{draft_note_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_merge_requests_merge_request_iid_draft_notes_draft_note_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid, --draft-note-id
    mr award-emoji delete-2 - DELETE /projects/{id}/merge_requests/{merge_request_iid}/notes/{note_id}/award_emoji/{award_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_merge_requests_merge_request_iid_notes_note_id_award_emoji_award_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid, --note-id, --award-id
    mr merge-requests view - GET /groups/{id}/merge_requests [intent=direct_read availability=implemented]; flags: --id
    mr merge-requests view-2 - GET /merge_requests [intent=direct_read availability=implemented]
    mr merge-requests view-3 - GET /projects/{id}/deployments/{deployment_id}/merge_requests [intent=direct_read availability=implemented]; flags: --id, --deployment-id
    mr related-merge-requests view - GET /projects/{id}/issues/{issue_iid}/related_merge_requests [intent=direct_read availability=implemented]; flags: --id, --issue-iid
    mr merge-requests view-4 - GET /projects/{id}/merge_requests [intent=direct_read availability=implemented]; flags: --id
    mr resource-milestone-events view - GET /projects/{id}/merge_requests/{eventable_id}/resource_milestone_events [intent=direct_read availability=implemented]; flags: --id, --eventable-id
    mr resource-milestone-events view-2 - GET /projects/{id}/merge_requests/{eventable_id}/resource_milestone_events/{event_id} [intent=direct_read availability=implemented]; flags: --id, --eventable-id, --event-id
    mr merge-requests view-5 - GET /projects/{id}/merge_requests/{merge_request_iid} [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid
    mr approval-state view - GET /projects/{id}/merge_requests/{merge_request_iid}/approval_state [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid
    mr approvals view - GET /projects/{id}/merge_requests/{merge_request_iid}/approvals [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid
    mr award-emoji view - GET /projects/{id}/merge_requests/{merge_request_iid}/award_emoji [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid
    mr award-emoji view-2 - GET /projects/{id}/merge_requests/{merge_request_iid}/award_emoji/{award_id} [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid, --award-id
    mr changes view - GET /projects/{id}/merge_requests/{merge_request_iid}/changes [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid
    mr closes-issues view - GET /projects/{id}/merge_requests/{merge_request_iid}/closes_issues [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid
    mr commits view - GET /projects/{id}/merge_requests/{merge_request_iid}/commits [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid
    mr context-commits view - GET /projects/{id}/merge_requests/{merge_request_iid}/context_commits [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid
    mr diffs view - GET /projects/{id}/merge_requests/{merge_request_iid}/diffs [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid
    mr draft-notes view - GET /projects/{id}/merge_requests/{merge_request_iid}/draft_notes [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid
    mr draft-notes view-2 - GET /projects/{id}/merge_requests/{merge_request_iid}/draft_notes/{draft_note_id} [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid, --draft-note-id
    mr merge-ref view - GET /projects/{id}/merge_requests/{merge_request_iid}/merge_ref [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid
    mr award-emoji view-3 - GET /projects/{id}/merge_requests/{merge_request_iid}/notes/{note_id}/award_emoji [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid, --note-id
    mr award-emoji view-4 - GET /projects/{id}/merge_requests/{merge_request_iid}/notes/{note_id}/award_emoji/{award_id} [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid, --note-id, --award-id
    mr participants view - GET /projects/{id}/merge_requests/{merge_request_iid}/participants [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid
    mr raw-diffs download - GET /projects/{id}/merge_requests/{merge_request_iid}/raw_diffs [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid
    mr related-issues view - GET /projects/{id}/merge_requests/{merge_request_iid}/related_issues [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid
    mr reviewers view - GET /projects/{id}/merge_requests/{merge_request_iid}/reviewers [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid
    mr time-stats view - GET /projects/{id}/merge_requests/{merge_request_iid}/time_stats [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid
    mr versions view - GET /projects/{id}/merge_requests/{merge_request_iid}/versions [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid
    mr versions view-2 - GET /projects/{id}/merge_requests/{merge_request_iid}/versions/{version_id} [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid, --version-id
    mr merge-requests view-6 - GET /projects/{id}/repository/commits/{sha}/merge_requests [intent=direct_read availability=implemented]; flags: --id, --sha
    mr merge-requests create - POST /projects/{id}/merge_requests [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_merge_requests]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    mr add-spent-time create - POST /projects/{id}/merge_requests/{merge_request_iid}/add_spent_time [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_merge_requests_merge_request_iid_add_spent_time]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid
    mr approve create - POST /projects/{id}/merge_requests/{merge_request_iid}/approve [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_merge_requests_merge_request_iid_approve]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid
    mr award-emoji create - POST /projects/{id}/merge_requests/{merge_request_iid}/award_emoji [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_merge_requests_merge_request_iid_award_emoji]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid
    mr cancel-merge-when-pipeline-succeeds create - POST /projects/{id}/merge_requests/{merge_request_iid}/cancel_merge_when_pipeline_succeeds [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_merge_requests_merge_request_iid_cancel_merge_when_pipeline_succeeds]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid
    mr context-commits create - POST /projects/{id}/merge_requests/{merge_request_iid}/context_commits [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_merge_requests_merge_request_iid_context_commits]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid
    mr draft-notes create - POST /projects/{id}/merge_requests/{merge_request_iid}/draft_notes [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_merge_requests_merge_request_iid_draft_notes]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid
    mr bulk-publish create - POST /projects/{id}/merge_requests/{merge_request_iid}/draft_notes/bulk_publish [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_merge_requests_merge_request_iid_draft_notes_bulk_publish]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid
    mr award-emoji create-2 - POST /projects/{id}/merge_requests/{merge_request_iid}/notes/{note_id}/award_emoji [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_merge_requests_merge_request_iid_notes_note_id_award_emoji]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid, --note-id
    mr reset-spent-time create - POST /projects/{id}/merge_requests/{merge_request_iid}/reset_spent_time [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_merge_requests_merge_request_iid_reset_spent_time]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid
    mr reset-time-estimate create - POST /projects/{id}/merge_requests/{merge_request_iid}/reset_time_estimate [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_merge_requests_merge_request_iid_reset_time_estimate]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid
    mr time-estimate create - POST /projects/{id}/merge_requests/{merge_request_iid}/time_estimate [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_merge_requests_merge_request_iid_time_estimate]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid
    mr unapprove create - POST /projects/{id}/merge_requests/{merge_request_iid}/unapprove [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_merge_requests_merge_request_iid_unapprove]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid
    mr merge-requests set - PUT /projects/{id}/merge_requests/{merge_request_iid} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_merge_requests_merge_request_iid]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid
    mr draft-notes set - PUT /projects/{id}/merge_requests/{merge_request_iid}/draft_notes/{draft_note_id} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_merge_requests_merge_request_iid_draft_notes_draft_note_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid, --draft-note-id
    mr publish set - PUT /projects/{id}/merge_requests/{merge_request_iid}/draft_notes/{draft_note_id}/publish [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_merge_requests_merge_request_iid_draft_notes_draft_note_id_publish]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid, --draft-note-id
    mr merge set - PUT /projects/{id}/merge_requests/{merge_request_iid}/merge [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_merge_requests_merge_request_iid_merge]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid
    mr rebase set - PUT /projects/{id}/merge_requests/{merge_request_iid}/rebase [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_merge_requests_merge_request_iid_rebase]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid
    mr reset-approvals set - PUT /projects/{id}/merge_requests/{merge_request_iid}/reset_approvals [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_merge_requests_merge_request_iid_reset_approvals]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid
    repo clone - Clone a GitLab repository locally [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Requires a constrained local git executor and destination path policy; not a connector API dispatch.
    repo archive - Download a repository archive [intent=direct_read availability=planned]; notes: Binary archive downloads require explicit size and output-path policy before enabling.
    repo create - Create a GitLab project [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would create a project and allocate namespace resources.; notes: No GitLab write action is declared yet.
    repo update - Update GitLab project settings [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would mutate project settings.; notes: No GitLab write action is declared yet.
    repo delete - Delete a GitLab project [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; destructive admin writes need explicit policy and typed confirmation.; risk: Deletes a project and its repository data.; notes: Repository deletion is intentionally not exposed.
    repo transfer - Transfer a GitLab project to another namespace [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; admin writes need explicit policy and typed confirmation.; risk: Changes project ownership and namespace access boundaries.; notes: Not exposed by this metadata slice.
    repo rules delete - DELETE /projects/{id}/registry/protection/repository/rules/{protection_rule_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_registry_protection_repository_rules_protection_rule_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --protection-rule-id
    repo repositories delete - DELETE /projects/{id}/registry/repositories/{repository_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_registry_repositories_repository_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --repository-id
    repo tags delete - DELETE /projects/{id}/registry/repositories/{repository_id}/tags [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_registry_repositories_repository_id_tags]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --repository-id
    repo tags delete-2 - DELETE /projects/{id}/registry/repositories/{repository_id}/tags/{tag_name} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_registry_repositories_repository_id_tags_tag_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --repository-id, --tag-name
    repo branches delete - DELETE /projects/{id}/repository/branches/{branch} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_repository_branches_branch]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --branch
    repo files delete - DELETE /projects/{id}/repository/files/{file_path} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_repository_files_file_path]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --file-path
    repo merged-branches delete - DELETE /projects/{id}/repository/merged_branches [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_repository_merged_branches]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    repo tags delete-3 - DELETE /projects/{id}/repository/tags/{tag_name} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_repository_tags_tag_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --tag-name
    repo pipeline-refs view - GET /geo/repositories/{gl_repository}/pipeline_refs [intent=direct_read availability=implemented]; flags: --gl-repository
    repo rules view - GET /projects/{id}/registry/protection/repository/rules [intent=direct_read availability=implemented]; flags: --id
    repo tags view - GET /projects/{id}/registry/repositories/{repository_id}/tags [intent=direct_read availability=implemented]; flags: --id, --repository-id
    repo tags view-2 - GET /projects/{id}/registry/repositories/{repository_id}/tags/{tag_name} [intent=direct_read availability=implemented]; flags: --id, --repository-id, --tag-name
    repo archive download - GET /projects/{id}/repository/archive [intent=direct_read availability=implemented]; flags: --id
    repo blobs view - GET /projects/{id}/repository/blobs/{sha} [intent=direct_read availability=implemented]; flags: --id, --sha
    repo raw download - GET /projects/{id}/repository/blobs/{sha}/raw [intent=direct_read availability=implemented]; flags: --id, --sha
    repo branches view - GET /projects/{id}/repository/branches [intent=direct_read availability=implemented]; flags: --id
    repo branches view-2 - GET /projects/{id}/repository/branches/{branch} [intent=direct_read availability=implemented]; flags: --id, --branch
    repo changelog view - GET /projects/{id}/repository/changelog [intent=direct_read availability=implemented]; flags: --id
    repo commits view - GET /projects/{id}/repository/commits [intent=direct_read availability=implemented]; flags: --id
    repo commits view-2 - GET /projects/{id}/repository/commits/{sha} [intent=direct_read availability=implemented]; flags: --id, --sha
    repo comments view - GET /projects/{id}/repository/commits/{sha}/comments [intent=direct_read availability=implemented]; flags: --id, --sha
    repo diff view - GET /projects/{id}/repository/commits/{sha}/diff [intent=direct_read availability=implemented]; flags: --id, --sha
    repo refs view - GET /projects/{id}/repository/commits/{sha}/refs [intent=direct_read availability=implemented]; flags: --id, --sha
    repo sequence view - GET /projects/{id}/repository/commits/{sha}/sequence [intent=direct_read availability=implemented]; flags: --id, --sha
    repo signature view - GET /projects/{id}/repository/commits/{sha}/signature [intent=direct_read availability=implemented]; flags: --id, --sha
    repo statuses view - GET /projects/{id}/repository/commits/{sha}/statuses [intent=direct_read availability=implemented]; flags: --id, --sha
    repo compare view - GET /projects/{id}/repository/compare [intent=direct_read availability=implemented]; flags: --id
    repo contributors view - GET /projects/{id}/repository/contributors [intent=direct_read availability=implemented]; flags: --id
    repo files view - GET /projects/{id}/repository/files/{file_path} [intent=direct_read availability=implemented]; flags: --id, --file-path
    repo blame view - GET /projects/{id}/repository/files/{file_path}/blame [intent=direct_read availability=implemented]; flags: --id, --file-path
    repo raw download-2 - GET /projects/{id}/repository/files/{file_path}/raw [intent=direct_read availability=implemented]; flags: --id, --file-path
    repo health view - GET /projects/{id}/repository/health [intent=direct_read availability=implemented]; flags: --id
    repo merge-base view - GET /projects/{id}/repository/merge_base [intent=direct_read availability=implemented]; flags: --id
    repo tags view-3 - GET /projects/{id}/repository/tags [intent=direct_read availability=implemented]; flags: --id
    repo tags view-4 - GET /projects/{id}/repository/tags/{tag_name} [intent=direct_read availability=implemented]; flags: --id, --tag-name
    repo signature view-2 - GET /projects/{id}/repository/tags/{tag_name}/signature [intent=direct_read availability=implemented]; flags: --id, --tag-name
    repo tree view - GET /projects/{id}/repository/tree [intent=direct_read availability=implemented]; flags: --id
    repo branches check - HEAD /projects/{id}/repository/branches/{branch} [intent=direct_read availability=implemented]; flags: --id, --branch
    repo files check - HEAD /projects/{id}/repository/files/{file_path} [intent=direct_read availability=implemented]; flags: --id, --file-path
    repo blame check - HEAD /projects/{id}/repository/files/{file_path}/blame [intent=direct_read availability=implemented]; flags: --id, --file-path
    repo rules update - PATCH /projects/{id}/registry/protection/repository/rules/{protection_rule_id} [intent=reverse_etl availability=implemented write=patch_api_v4_projects_id_registry_protection_repository_rules_protection_rule_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --protection-rule-id
    repo rules create - POST /projects/{id}/registry/protection/repository/rules [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_registry_protection_repository_rules]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    repo branches create - POST /projects/{id}/repository/branches [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_repository_branches]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    repo changelog create - POST /projects/{id}/repository/changelog [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_repository_changelog]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    repo commits create - POST /projects/{id}/repository/commits [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_repository_commits]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    repo cherry-pick create - POST /projects/{id}/repository/commits/{sha}/cherry_pick [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_repository_commits_sha_cherry_pick]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --sha
    repo comments create - POST /projects/{id}/repository/commits/{sha}/comments [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_repository_commits_sha_comments]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --sha
    repo revert create - POST /projects/{id}/repository/commits/{sha}/revert [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_repository_commits_sha_revert]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --sha
    repo files create - POST /projects/{id}/repository/files/{file_path} [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_repository_files_file_path]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --file-path
    repo tags create - POST /projects/{id}/repository/tags [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_repository_tags]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    repo repository-size create - POST /projects/{id}/repository_size [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_repository_size]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    repo protect set - PUT /projects/{id}/repository/branches/{branch}/protect [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_repository_branches_branch_protect]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --branch
    repo unprotect set - PUT /projects/{id}/repository/branches/{branch}/unprotect [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_repository_branches_branch_unprotect]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --branch
    repo files set - PUT /projects/{id}/repository/files/{file_path} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_repository_files_file_path]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --file-path
    repo submodules set - PUT /projects/{id}/repository/submodules/{submodule} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_repository_submodules_submodule]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --submodule
  CI/CD Commands
    pipeline list - List CI/CD pipelines [intent=etl availability=planned]; notes: Pipelines are ETL candidates but are not current streams.
    pipeline view - View one CI/CD pipeline [intent=direct_read availability=planned]; notes: Pipeline detail reads require direct-read operation metadata.
    pipeline run - Run a CI/CD pipeline [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; execution requires plan, preview, approval, and typed confirmation policy.; risk: Starts CI/CD execution and may deploy or mutate environments.; notes: No pipeline-trigger write action is exposed by this metadata slice.
    pipeline cancel - Cancel a CI/CD pipeline [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would stop CI/CD execution for a project pipeline.; notes: No GitLab write action is declared yet.
    pipeline pipelines delete - DELETE /projects/{id}/pipelines/{pipeline_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_pipelines_pipeline_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --pipeline-id
    pipeline pipelines view - GET /projects/{id}/merge_requests/{merge_request_iid}/pipelines [intent=direct_read availability=implemented]; flags: --id, --merge-request-iid
    pipeline pipelines download - GET /projects/{id}/packages/{package_id}/pipelines [intent=direct_read availability=implemented]; flags: --id, --package-id
    pipeline pipelines view-2 - GET /projects/{id}/pipeline_schedules/{pipeline_schedule_id}/pipelines [intent=direct_read availability=implemented]; flags: --id, --pipeline-schedule-id
    pipeline pipelines view-3 - GET /projects/{id}/pipelines [intent=direct_read availability=implemented]; flags: --id
    pipeline latest view - GET /projects/{id}/pipelines/latest [intent=direct_read availability=implemented]; flags: --id
    pipeline pipelines view-4 - GET /projects/{id}/pipelines/{pipeline_id} [intent=direct_read availability=implemented]; flags: --id, --pipeline-id
    pipeline jobs view - GET /projects/{id}/pipelines/{pipeline_id}/jobs [intent=direct_read availability=implemented]; flags: --id, --pipeline-id
    pipeline test-report view - GET /projects/{id}/pipelines/{pipeline_id}/test_report [intent=direct_read availability=implemented]; flags: --id, --pipeline-id
    pipeline test-report-summary view - GET /projects/{id}/pipelines/{pipeline_id}/test_report_summary [intent=direct_read availability=implemented]; flags: --id, --pipeline-id
    pipeline trigger-jobs view - GET /projects/{id}/pipelines/{pipeline_id}/trigger_jobs [intent=direct_read availability=implemented]; flags: --id, --pipeline-id
    pipeline pipelines create - POST /projects/{id}/merge_requests/{merge_request_iid}/pipelines [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_merge_requests_merge_request_iid_pipelines]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --merge-request-iid
    pipeline cancel create - POST /projects/{id}/pipelines/{pipeline_id}/cancel [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_pipelines_pipeline_id_cancel]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --pipeline-id
    pipeline retry create - POST /projects/{id}/pipelines/{pipeline_id}/retry [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_pipelines_pipeline_id_retry]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --pipeline-id
    pipeline pipelines-email set - PUT /groups/{id}/integrations/pipelines-email [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_integrations_pipelines_email]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    pipeline pipelines-email set-2 - PUT /projects/{id}/integrations/pipelines-email [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_integrations_pipelines_email]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    pipeline metadata set - PUT /projects/{id}/pipelines/{pipeline_id}/metadata [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_pipelines_pipeline_id_metadata]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --pipeline-id
    pipeline pipelines-email set-3 - PUT /projects/{id}/services/pipelines-email [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_services_pipelines_email]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    job view - View a CI/CD job [intent=direct_read availability=planned]; notes: Job reads are bounded direct-read candidates.
    job artifact download - Download CI/CD job artifacts [intent=direct_read availability=planned]; notes: Binary downloads require explicit size limits and output destination policy before enabling.
    job view-2 - GET /job [intent=direct_read availability=implemented]
    job allowed-agents view - GET /job/allowed_agents [intent=direct_read availability=implemented]
    schedule list - List pipeline schedules [intent=etl availability=planned]; notes: Schedule stream coverage is deferred.
    schedule run - Run a pipeline schedule [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; execution requires plan, preview, approval, and typed confirmation policy.; risk: Starts scheduled CI/CD execution and may deploy or mutate environments.; notes: Not exposed by this metadata slice.
    runner list - List GitLab runners [intent=etl availability=planned]; notes: Runner inventory often requires elevated scope and is deferred to the operation ledger.
    runner register - Register a GitLab runner [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; admin infrastructure writes need explicit policy and typed confirmation.; risk: Changes CI execution infrastructure and may introduce privileged compute.; notes: Not exposed by this metadata slice.
    runner runners delete - DELETE /projects/{id}/runners/{runner_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_runners_runner_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --runner-id
    runner runners delete-2 - DELETE /runners [intent=reverse_etl availability=implemented write=delete_api_v4_runners]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    runner managers delete - DELETE /runners/managers [intent=reverse_etl availability=implemented write=delete_api_v4_runners_managers]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    runner runners delete-3 - DELETE /runners/{id} [intent=reverse_etl availability=implemented write=delete_api_v4_runners_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    runner runners view - GET /groups/{id}/runners [intent=direct_read availability=implemented]; flags: --id
    runner runners view-2 - GET /projects/{id}/runners [intent=direct_read availability=implemented]; flags: --id
    runner runners view-3 - GET /runners [intent=direct_read availability=implemented]
    runner all view - GET /runners/all [intent=direct_read availability=implemented]
    runner discovery view - GET /runners/router/discovery [intent=direct_read availability=implemented]
    runner runners view-4 - GET /runners/{id} [intent=direct_read availability=implemented]; flags: --id
    runner jobs view - GET /runners/{id}/jobs [intent=direct_read availability=implemented]; flags: --id
    runner managers view - GET /runners/{id}/managers [intent=direct_read availability=implemented]; flags: --id
    runner view - GET /runners/{id}/projects [intent=direct_read availability=implemented]; flags: --id
    runner runners create - POST /projects/{id}/runners [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_runners]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    runner runners create-2 - POST /runners [intent=reverse_etl availability=implemented write=post_api_v4_runners]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    runner verify create - POST /runners/verify [intent=reverse_etl availability=implemented write=post_api_v4_runners_verify]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    runner runners create-3 - POST /user/runners [intent=reverse_etl availability=implemented write=post_api_v4_user_runners]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    runner runners set - PUT /runners/{id} [intent=reverse_etl availability=implemented write=put_api_v4_runners_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
  Collaboration Commands
    label list - List labels [intent=etl availability=planned]; notes: Label stream coverage is deferred.
    label create - Create a label [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would create a label in a project or group.; notes: No GitLab write action is declared yet.
    label delete - Delete a label [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; destructive writes need explicit policy and typed confirmation.; risk: Deletes a label and may affect issue or merge request workflows.; notes: Not exposed by this metadata slice.
    milestone list - List milestones [intent=etl availability=planned]; notes: Milestone stream coverage is deferred.
    release list - List releases [intent=etl availability=planned]; notes: Release stream coverage is deferred.
    release create - Create a release [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would publish a release in a GitLab project.; notes: No GitLab write action is declared yet.
    release download - Download release assets [intent=direct_read availability=planned]; notes: Binary asset downloads require explicit size limits and output destination policy before enabling.
    snippet list - List snippets [intent=etl availability=planned]; notes: Snippet streams are deferred to the operation ledger.
    snippet create - Create a snippet [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would create project or personal snippet content.; notes: No GitLab write action is declared yet.
    todo list - List to-do items [intent=direct_read availability=planned]; notes: To-do items are user-scoped direct-read candidates.
    todo done - Mark a to-do item done [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: Would mutate a user-scoped to-do item.; notes: No GitLab write action is declared yet.
  Security And Administration Commands
    variable list - List CI/CD variables [intent=direct_read availability=planned]; risk: Variable metadata may be sensitive even when values are masked or hidden.; notes: Requires sensitive-field redaction policy before enabling.
    variable set - Create or update a CI/CD variable [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; sensitive writes require stdin/env input, preview redaction, approval, and typed confirmation.; risk: Writes secret or deployment-affecting configuration.; notes: Never request or store variable values in prompts or metadata.
    variable delete - Delete a CI/CD variable [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; destructive sensitive writes require explicit policy and typed confirmation.; risk: Deletes secret or deployment-affecting configuration.; notes: Not exposed by this metadata slice.
    variable variables delete - DELETE /admin/ci/variables/{key} [intent=reverse_etl availability=implemented write=delete_api_v4_admin_ci_variables_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --key
    variable variables delete-2 - DELETE /groups/{id}/variables/{key} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_variables_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --key
    variable url-variables delete - DELETE /hooks/{hook_id}/url_variables/{key} [intent=reverse_etl availability=implemented write=delete_api_v4_hooks_hook_id_url_variables_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --hook-id, --key
    variable url-variables delete-2 - DELETE /projects/{id}/hooks/{hook_id}/url_variables/{key} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_hooks_hook_id_url_variables_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --hook-id, --key
    variable variables delete-3 - DELETE /projects/{id}/pipeline_schedules/{pipeline_schedule_id}/variables/{key} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_pipeline_schedules_pipeline_schedule_id_variables_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --pipeline-schedule-id, --key
    variable variables delete-4 - DELETE /projects/{id}/variables/{key} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_variables_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --key
    variable variables view - GET /admin/ci/variables [intent=direct_read availability=implemented]
    variable variables view-2 - GET /admin/ci/variables/{key} [intent=direct_read availability=implemented]; flags: --key
    variable variables view-3 - GET /groups/{id}/variables [intent=direct_read availability=implemented]; flags: --id
    variable variables view-4 - GET /groups/{id}/variables/{key} [intent=direct_read availability=implemented]; flags: --id, --key
    variable variables view-5 - GET /projects/{id}/pipeline_schedules/{pipeline_schedule_id}/variables/{key} [intent=direct_read availability=implemented]; flags: --id, --pipeline-schedule-id, --key
    variable variables view-6 - GET /projects/{id}/pipelines/{pipeline_id}/variables [intent=direct_read availability=implemented]; flags: --id, --pipeline-id
    variable variables view-7 - GET /projects/{id}/variables [intent=direct_read availability=implemented]; flags: --id
    variable variables view-8 - GET /projects/{id}/variables/{key} [intent=direct_read availability=implemented]; flags: --id, --key
    variable variables create - POST /admin/ci/variables [intent=reverse_etl availability=implemented write=post_api_v4_admin_ci_variables]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution
    variable variables create-2 - POST /groups/{id}/variables [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_variables]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    variable variables create-3 - POST /projects/{id}/pipeline_schedules/{pipeline_schedule_id}/variables [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_pipeline_schedules_pipeline_schedule_id_variables]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --pipeline-schedule-id
    variable variables create-4 - POST /projects/{id}/variables [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_variables]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    variable variables set - PUT /admin/ci/variables/{key} [intent=reverse_etl availability=implemented write=put_api_v4_admin_ci_variables_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --key
    variable variables set-2 - PUT /groups/{id}/variables/{key} [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_variables_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --key
    variable url-variables set - PUT /hooks/{hook_id}/url_variables/{key} [intent=reverse_etl availability=implemented write=put_api_v4_hooks_hook_id_url_variables_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --hook-id, --key
    variable url-variables set-2 - PUT /projects/{id}/hooks/{hook_id}/url_variables/{key} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_hooks_hook_id_url_variables_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --hook-id, --key
    variable variables set-3 - PUT /projects/{id}/pipeline_schedules/{pipeline_schedule_id}/variables/{key} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_pipeline_schedules_pipeline_schedule_id_variables_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --pipeline-schedule-id, --key
    variable variables set-4 - PUT /projects/{id}/variables/{key} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_variables_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --key
    deploy-key list - List deploy keys [intent=direct_read availability=planned]; risk: Deploy key metadata can reveal access configuration.; notes: Requires operation-ledger classification and redaction policy before enabling.
    deploy-key add - Add a deploy key [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; access-control writes require explicit policy and typed confirmation.; risk: Grants repository access to a key.; notes: Not exposed by this metadata slice.
    ssh-key list - List account SSH keys [intent=direct_read availability=planned]; risk: Account key metadata can be sensitive.; notes: Requires redaction and account-scope policy before enabling.
    ssh-key add - Add an account SSH key [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; access-control writes require explicit policy and typed confirmation.; risk: Grants account access to a key.; notes: Not exposed by this metadata slice.
    gpg-key list - List account GPG keys [intent=direct_read availability=planned]; notes: Account-scoped key metadata requires direct-read policy before enabling.
    token list - List personal, project, or group tokens [intent=direct_read availability=unsafe_or_disallowed]; risk: Token metadata is sensitive and may require elevated scope.; notes: Blocked until sensitive/admin policy defines risk tiers and redaction.
    token rotate - Rotate a token [intent=reverse_etl availability=unsafe_or_disallowed]; approval: Blocked by default; credential lifecycle writes require human-approved policy.; risk: Mutates credentials and can break automation or reveal newly issued secret material.; notes: Not exposed by this metadata slice.
    token deploy-tokens delete - DELETE /groups/{id}/deploy_tokens/{token_id} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_deploy_tokens_token_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --token-id
    token self delete - DELETE /personal_access_tokens/self [intent=reverse_etl availability=implemented write=delete_api_v4_personal_access_tokens_self]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    token personal-access-tokens delete - DELETE /personal_access_tokens/{id} [intent=reverse_etl availability=implemented write=delete_api_v4_personal_access_tokens_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    token tokens delete - DELETE /projects/{id}/cluster_agents/{agent_id}/tokens/{token_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_cluster_agents_agent_id_tokens_token_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --agent-id, --token-id
    token deploy-tokens delete-2 - DELETE /projects/{id}/deploy_tokens/{token_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_deploy_tokens_token_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --token-id
    token allowlist delete - DELETE /projects/{id}/job_token_scope/allowlist/{target_project_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_job_token_scope_allowlist_target_project_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --target-project-id
    token groups-allowlist delete - DELETE /projects/{id}/job_token_scope/groups_allowlist/{target_group_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_job_token_scope_groups_allowlist_target_group_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --target-group-id
    token deploy-tokens view - GET /deploy_tokens [intent=direct_read availability=implemented]
    token deploy-tokens view-2 - GET /groups/{id}/deploy_tokens [intent=direct_read availability=implemented]; flags: --id
    token deploy-tokens view-3 - GET /groups/{id}/deploy_tokens/{token_id} [intent=direct_read availability=implemented]; flags: --id, --token-id
    token personal-access-tokens view - GET /personal_access_tokens [intent=direct_read availability=implemented]
    token self view - GET /personal_access_tokens/self [intent=direct_read availability=implemented]
    token associations view - GET /personal_access_tokens/self/associations [intent=direct_read availability=implemented]
    token personal-access-tokens view-2 - GET /personal_access_tokens/{id} [intent=direct_read availability=implemented]; flags: --id
    token tokens view - GET /projects/{id}/cluster_agents/{agent_id}/tokens [intent=direct_read availability=implemented]; flags: --id, --agent-id
    token tokens view-2 - GET /projects/{id}/cluster_agents/{agent_id}/tokens/{token_id} [intent=direct_read availability=implemented]; flags: --id, --agent-id, --token-id
    token deploy-tokens view-4 - GET /projects/{id}/deploy_tokens [intent=direct_read availability=implemented]; flags: --id
    token deploy-tokens view-5 - GET /projects/{id}/deploy_tokens/{token_id} [intent=direct_read availability=implemented]; flags: --id, --token-id
    token job-token-scope view - GET /projects/{id}/job_token_scope [intent=direct_read availability=implemented]; flags: --id
    token allowlist view - GET /projects/{id}/job_token_scope/allowlist [intent=direct_read availability=implemented]; flags: --id
    token groups-allowlist view - GET /projects/{id}/job_token_scope/groups_allowlist [intent=direct_read availability=implemented]; flags: --id
    token job-token-scope update - PATCH /projects/{id}/job_token_scope [intent=reverse_etl availability=implemented write=patch_api_v4_projects_id_job_token_scope]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    token rotate create - POST /groups/{id}/access_tokens/self/rotate [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_access_tokens_self_rotate]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    token deploy-tokens create - POST /groups/{id}/deploy_tokens [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_deploy_tokens]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    token reset-registration-token create - POST /groups/{id}/runners/reset_registration_token [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_runners_reset_registration_token]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    token forge-token create - POST /integrations/jira_forge/installation/forge_token [intent=reverse_etl availability=implemented write=post_api_v4_integrations_jira_forge_installation_forge_token]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution
    token rotate create-2 - POST /personal_access_tokens/self/rotate [intent=reverse_etl availability=implemented write=post_api_v4_personal_access_tokens_self_rotate]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution
    token rotate create-3 - POST /personal_access_tokens/{id}/rotate [intent=reverse_etl availability=implemented write=post_api_v4_personal_access_tokens_id_rotate]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    token rotate create-4 - POST /projects/{id}/access_tokens/self/rotate [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_access_tokens_self_rotate]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    token tokens create - POST /projects/{id}/cluster_agents/{agent_id}/tokens [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_cluster_agents_agent_id_tokens]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --agent-id
    token deploy-tokens create-2 - POST /projects/{id}/deploy_tokens [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_deploy_tokens]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    token allowlist create - POST /projects/{id}/job_token_scope/allowlist [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_job_token_scope_allowlist]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    token groups-allowlist create - POST /projects/{id}/job_token_scope/groups_allowlist [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_job_token_scope_groups_allowlist]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    token reset-registration-token create-2 - POST /projects/{id}/runners/reset_registration_token [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_runners_reset_registration_token]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    token reset-authentication-token create - POST /runners/reset_authentication_token [intent=reverse_etl availability=implemented write=post_api_v4_runners_reset_authentication_token]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution
    token reset-registration-token create-3 - POST /runners/reset_registration_token [intent=reverse_etl availability=implemented write=post_api_v4_runners_reset_registration_token]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution
    token reset-authentication-token create-2 - POST /runners/{id}/reset_authentication_token [intent=reverse_etl availability=implemented write=post_api_v4_runners_id_reset_authentication_token]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    securefile list - List secure files [intent=direct_read availability=unsafe_or_disallowed]; risk: Secure files may contain signing keys or other sensitive binary content.; notes: Blocked until sensitive binary-output policy exists.
    securefile download - Download secure files [intent=direct_read availability=unsafe_or_disallowed]; risk: Downloads sensitive binary material to local storage.; notes: Blocked until bounded executor, destination policy, and approval policy exist.
    container-registry list - List container registry repositories or tags [intent=etl availability=planned]; notes: Registry inventory is an ETL candidate, deferred to operation ledger classification.
    packages list - List package registry packages [intent=etl availability=planned]; notes: Package registry inventory is an ETL candidate, deferred to operation ledger classification.
    packages conans delete - DELETE /packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel} [intent=reverse_etl availability=implemented write=delete_api_v4_packages_conan_v1_conans_package_name_package_version_package_username_package_channel]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --package-name, --package-version, --package-username, --package-channel
    packages dist-tags delete - DELETE /packages/npm/-/package/*package_name/dist-tags/{tag} [intent=reverse_etl availability=implemented write=delete_api_v4_packages_npm_package_package_name_dist_tags_tag]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --tag
    packages search download - GET /packages/conan/v1/conans/search [intent=direct_read availability=implemented]
    packages conans download - GET /packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel} [intent=direct_read availability=implemented]; flags: --package-name, --package-version, --package-username, --package-channel
    packages digest download - GET /packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/digest [intent=direct_read availability=implemented]; flags: --package-name, --package-version, --package-username, --package-channel
    packages download-urls download - GET /packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/download_urls [intent=direct_read availability=implemented]; flags: --package-name, --package-version, --package-username, --package-channel
    packages download - GET /packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/packages/{conan_package_reference} [intent=direct_read availability=implemented]; flags: --package-name, --package-version, --package-username, --package-channel, --conan-package-reference
    packages digest download-2 - GET /packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/packages/{conan_package_reference}/digest [intent=direct_read availability=implemented]; flags: --package-name, --package-version, --package-username, --package-channel, --conan-package-reference
    packages download-urls download-2 - GET /packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/packages/{conan_package_reference}/download_urls [intent=direct_read availability=implemented]; flags: --package-name, --package-version, --package-username, --package-channel, --conan-package-reference
    packages search download-2 - GET /packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/search [intent=direct_read availability=implemented]; flags: --package-name, --package-version, --package-username, --package-channel
    packages export download - GET /packages/conan/v1/files/{package_name}/{package_version}/{package_username}/{package_channel}/{recipe_revision}/export/{file_name} [intent=direct_read availability=implemented]; flags: --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --file-name
    packages package download - GET /packages/conan/v1/files/{package_name}/{package_version}/{package_username}/{package_channel}/{recipe_revision}/package/{conan_package_reference}/{package_revision}/{file_name} [intent=direct_read availability=implemented]; flags: --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --conan-package-reference, --package-revision, --file-name
    packages ping download - GET /packages/conan/v1/ping [intent=direct_read availability=implemented]
    packages authenticate download - GET /packages/conan/v1/users/authenticate [intent=direct_read availability=implemented]
    packages check-credentials download - GET /packages/conan/v1/users/check_credentials [intent=direct_read availability=implemented]
    packages *path download - GET /packages/maven/*path/{file_name} [intent=direct_read availability=implemented]; flags: --file-name
    packages *package-name download - GET /packages/npm/*package_name [intent=direct_read availability=implemented]
    packages dist-tags download - GET /packages/npm/-/package/*package_name/dist-tags [intent=direct_read availability=implemented]
    packages v1 download - GET /packages/terraform/modules/v1/{module_namespace}/{module_name}/{module_system} [intent=direct_read availability=implemented]; flags: --module-namespace, --module-name, --module-system
    packages *module-version download - GET /packages/terraform/modules/v1/{module_namespace}/{module_name}/{module_system}/*module_version [intent=direct_read availability=implemented]; flags: --module-namespace, --module-name, --module-system
    packages download download - GET /packages/terraform/modules/v1/{module_namespace}/{module_name}/{module_system}/*module_version/download [intent=direct_read availability=implemented]; flags: --module-namespace, --module-name, --module-system
    packages file download - GET /packages/terraform/modules/v1/{module_namespace}/{module_name}/{module_system}/*module_version/file [intent=direct_read availability=implemented]; flags: --module-namespace, --module-name, --module-system
    packages download download-2 - GET /packages/terraform/modules/v1/{module_namespace}/{module_name}/{module_system}/download [intent=direct_read availability=implemented]; flags: --module-namespace, --module-name, --module-system
    packages versions download - GET /packages/terraform/modules/v1/{module_namespace}/{module_name}/{module_system}/versions [intent=direct_read availability=implemented]; flags: --module-namespace, --module-name, --module-system
    packages upload-urls create - POST /packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/packages/{conan_package_reference}/upload_urls [intent=reverse_etl availability=implemented write=post_api_v4_packages_conan_v1_conans_package_name_package_version_package_username_package_channel_packages_conan_package_reference_upload_urls]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --package-name, --package-version, --package-username, --package-channel, --conan-package-reference
    packages upload-urls create-2 - POST /packages/conan/v1/conans/{package_name}/{package_version}/{package_username}/{package_channel}/upload_urls [intent=reverse_etl availability=implemented write=post_api_v4_packages_conan_v1_conans_package_name_package_version_package_username_package_channel_upload_urls]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --package-name, --package-version, --package-username, --package-channel
    packages bulk create - POST /packages/npm/-/npm/v1/security/advisories/bulk [intent=reverse_etl availability=implemented write=post_api_v4_packages_npm_npm_v1_security_advisories_bulk]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    packages quick create - POST /packages/npm/-/npm/v1/security/audits/quick [intent=reverse_etl availability=implemented write=post_api_v4_packages_npm_npm_v1_security_audits_quick]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    packages export set - PUT /packages/conan/v1/files/{package_name}/{package_version}/{package_username}/{package_channel}/{recipe_revision}/export/{file_name} [intent=reverse_etl availability=implemented write=put_api_v4_packages_conan_v1_files_package_name_package_version_package_username_package_channel_recipe_revision_export_file_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --file-name
    packages authorize set - PUT /packages/conan/v1/files/{package_name}/{package_version}/{package_username}/{package_channel}/{recipe_revision}/export/{file_name}/authorize [intent=reverse_etl availability=implemented write=put_api_v4_packages_conan_v1_files_package_name_package_version_package_username_package_channel_recipe_revision_export_file_name_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --file-name
    packages package set - PUT /packages/conan/v1/files/{package_name}/{package_version}/{package_username}/{package_channel}/{recipe_revision}/package/{conan_package_reference}/{package_revision}/{file_name} [intent=reverse_etl availability=implemented write=put_api_v4_packages_conan_v1_files_package_name_package_version_package_username_package_channel_recipe_revision_package_conan_package_reference_package_revision_file_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --conan-package-reference, --package-revision, --file-name
    packages authorize set-2 - PUT /packages/conan/v1/files/{package_name}/{package_version}/{package_username}/{package_channel}/{recipe_revision}/package/{conan_package_reference}/{package_revision}/{file_name}/authorize [intent=reverse_etl availability=implemented write=put_api_v4_packages_conan_v1_files_package_name_package_version_package_username_package_channel_recipe_revision_package_conan_package_reference_package_revision_file_name_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --package-name, --package-version, --package-username, --package-channel, --recipe-revision, --conan-package-reference, --package-revision, --file-name
    packages dist-tags set - PUT /packages/npm/-/package/*package_name/dist-tags/{tag} [intent=reverse_etl availability=implemented write=put_api_v4_packages_npm_package_package_name_dist_tags_tag]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --tag
  Local Workflow Commands
    auth login - Authenticate glab locally [intent=auth availability=excluded]; notes: Polymetrics credentials are managed through pm credential flows and never through prompt text.
    config set - Set local glab configuration [intent=config availability=unsupported_local unsupported local workflow]; notes: Local glab configuration is outside the GitLab connector.
    alias set - Configure a local glab alias [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Local alias configuration is outside connector execution.
    completion - Generate shell completion scripts [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Shell completion generation is handled by the pm CLI, not by the GitLab connector surface.
  Additional Commands
    api - Make an arbitrary GitLab API request [intent=raw_api availability=unsafe_or_disallowed]; approval: Not exposed. Add typed operations instead of raw API access.; risk: Arbitrary API dispatch can bypass connector safety, approval, redaction, and operation-ledger classification.; notes: Generic raw HTTP reads/writes are intentionally disallowed.
    search code - Search code or project resources [intent=direct_read availability=planned]; notes: Search is a bounded direct-read candidate and must not become a raw API escape hatch.
    search view - GET /search [intent=direct_read availability=implemented]
    changelog generate - Generate changelogs from project history [intent=direct_read availability=planned]; notes: Changelog generation combines reads and local formatting; it needs an explicit bounded workflow before exposure.
    cluster list - List GitLab Agents for Kubernetes clusters [intent=direct_read availability=planned]; risk: Cluster metadata can be sensitive and may require elevated scope.; notes: Deferred to operation-ledger and sensitive/admin policy lanes.
    duo prompt - Interact with GitLab Duo [intent=docs_only availability=excluded]; notes: Interactive AI assistant behavior is outside connector ETL/reverse ETL scope.
    mcp serve - Run the glab MCP server [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Local server processes are outside connector command dispatch.
    opentofu state - Work with OpenTofu or Terraform integration state [intent=direct_read availability=planned]; risk: Infrastructure state can contain sensitive data and environment topology.; notes: Requires explicit operation classification and redaction policy before enabling.
    stack create - Create or manage stacked diffs [intent=local_workflow availability=unsupported_local unsupported local workflow]; notes: Stacked-diff workflow requires local git operations and is outside connector API dispatch.
  Other Commands
    admin clusters delete - DELETE /admin/clusters/{cluster_id} [intent=reverse_etl availability=implemented write=delete_api_v4_admin_clusters_cluster_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --cluster-id
    applications delete - DELETE /applications/{id} [intent=reverse_etl availability=implemented write=delete_api_v4_applications_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    broadcast-messages delete - DELETE /broadcast_messages/{id} [intent=reverse_etl availability=implemented write=delete_api_v4_broadcast_messages_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    features delete - DELETE /features/{name} [intent=reverse_etl availability=implemented write=delete_api_v4_features_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --name
    access access-requests delete - DELETE /groups/{id}/access_requests/{user_id} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_access_requests_user_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --user-id
    access billable-members delete - DELETE /groups/{id}/billable_members/{user_id} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_billable_members_user_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --user-id
    access members delete - DELETE /groups/{id}/members/{user_id} [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_members_user_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --user-id
    access override delete - DELETE /groups/{id}/members/{user_id}/override [intent=reverse_etl availability=implemented write=delete_api_v4_groups_id_members_user_id_override]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --user-id
    hooks delete - DELETE /hooks/{hook_id} [intent=reverse_etl availability=implemented write=delete_api_v4_hooks_hook_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --hook-id
    hooks custom-headers delete - DELETE /hooks/{hook_id}/custom_headers/{key} [intent=reverse_etl availability=implemented write=delete_api_v4_hooks_hook_id_custom_headers_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --hook-id, --key
    integrations subscriptions delete - DELETE /integrations/jira_forge/subscriptions/{id} [intent=reverse_etl availability=implemented write=delete_api_v4_integrations_jira_forge_subscriptions_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    namespaces limit-exclusion delete - DELETE /namespaces/{id}/storage/limit_exclusion [intent=reverse_etl availability=implemented write=delete_api_v4_namespaces_id_storage_limit_exclusion]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    organizations delete - DELETE /organizations/{id} [intent=reverse_etl availability=implemented write=delete_api_v4_organizations_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    access access-requests delete-2 - DELETE /projects/{id}/access_requests/{user_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_access_requests_user_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --user-id
    access members delete-2 - DELETE /projects/{id}/members/{user_id} [intent=reverse_etl availability=implemented write=delete_api_v4_projects_id_members_user_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --user-id
    snippets delete - DELETE /snippets/{id} [intent=reverse_etl availability=implemented write=delete_api_v4_snippets_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    topics delete - DELETE /topics/{id} [intent=reverse_etl availability=implemented write=delete_api_v4_topics_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: destructive GitLab mutation is blocked by default and requires explicit reverse-ETL policy plus typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    admin batched-background-migrations view - GET /admin/batched_background_migrations [intent=direct_read availability=implemented]
    admin batched-background-migrations view-2 - GET /admin/batched_background_migrations/{id} [intent=direct_read availability=implemented]; flags: --id
    admin batched-background-operations view - GET /admin/batched_background_operations [intent=direct_read availability=implemented]
    admin batched-background-operations view-2 - GET /admin/batched_background_operations/{id} [intent=direct_read availability=implemented]; flags: --id
    admin clusters view - GET /admin/clusters [intent=direct_read availability=implemented]
    admin clusters view-2 - GET /admin/clusters/{cluster_id} [intent=direct_read availability=implemented]; flags: --cluster-id
    admin tables view - GET /admin/databases/{database_name}/dictionary/tables/{table_name} [intent=direct_read availability=implemented]; flags: --database-name, --table-name
    admin pending view - GET /admin/migrations/pending [intent=direct_read availability=implemented]
    application appearance view - GET /application/appearance [intent=direct_read availability=implemented]
    application plan-limits view - GET /application/plan_limits [intent=direct_read availability=implemented]
    application statistics view - GET /application/statistics [intent=direct_read availability=implemented]
    applications view - GET /applications [intent=direct_read availability=implemented]
    avatar download - GET /avatar [intent=direct_read availability=implemented]
    broadcast-messages view - GET /broadcast_messages [intent=direct_read availability=implemented]
    broadcast-messages view-2 - GET /broadcast_messages/{id} [intent=direct_read availability=implemented]; flags: --id
    bulk-imports view - GET /bulk_imports [intent=direct_read availability=implemented]
    bulk-imports entities view - GET /bulk_imports/entities [intent=direct_read availability=implemented]
    bulk-imports view-2 - GET /bulk_imports/{import_id} [intent=direct_read availability=implemented]; flags: --import-id
    bulk-imports entities view-2 - GET /bulk_imports/{import_id}/entities [intent=direct_read availability=implemented]; flags: --import-id
    bulk-imports entities view-3 - GET /bulk_imports/{import_id}/entities/{entity_id} [intent=direct_read availability=implemented]; flags: --import-id, --entity-id
    bulk-imports failures view - GET /bulk_imports/{import_id}/entities/{entity_id}/failures [intent=direct_read availability=implemented]; flags: --import-id, --entity-id
    databases tables view - GET /databases/{database_name}/dictionary/tables [intent=direct_read availability=implemented]; flags: --database-name
    deploy-keys view - GET /deploy_keys [intent=direct_read availability=implemented]
    discover-cert-based-clusters view - GET /discover-cert-based-clusters [intent=direct_read availability=implemented]
    events view - GET /events [intent=direct_read availability=implemented]
    feature-flags unleash view - GET /feature_flags/unleash/{project_id} [intent=direct_read availability=implemented]; flags: --project-id
    feature-flags features view - GET /feature_flags/unleash/{project_id}/client/features [intent=direct_read availability=implemented]; flags: --project-id
    features view - GET /features [intent=direct_read availability=implemented]
    features definitions view - GET /features/definitions [intent=direct_read availability=implemented]
    geo proxy view - GET /geo/proxy [intent=direct_read availability=implemented]
    geo retrieve view - GET /geo/retrieve/{replicable_name}/{replicable_id} [intent=direct_read availability=implemented]; flags: --replicable-name, --replicable-id
    access access-requests view - GET /groups/{id}/access_requests [intent=direct_read availability=implemented]; flags: --id
    access billable-members view - GET /groups/{id}/billable_members [intent=direct_read availability=implemented]; flags: --id
    access indirect view - GET /groups/{id}/billable_members/{user_id}/indirect [intent=direct_read availability=implemented]; flags: --id, --user-id
    access memberships view - GET /groups/{id}/billable_members/{user_id}/memberships [intent=direct_read availability=implemented]; flags: --id, --user-id
    access members view - GET /groups/{id}/members [intent=direct_read availability=implemented]; flags: --id
    access all view - GET /groups/{id}/members/all [intent=direct_read availability=implemented]; flags: --id
    access all view-2 - GET /groups/{id}/members/all/{user_id} [intent=direct_read availability=implemented]; flags: --id, --user-id
    access members view-2 - GET /groups/{id}/members/{user_id} [intent=direct_read availability=implemented]; flags: --id, --user-id
    access pending-members view - GET /groups/{id}/pending_members [intent=direct_read availability=implemented]; flags: --id
    hooks view - GET /hooks [intent=direct_read availability=implemented]
    hooks view-2 - GET /hooks/{hook_id} [intent=direct_read availability=implemented]; flags: --hook-id
    integrations subscriptions view - GET /integrations/jira_forge/subscriptions [intent=direct_read availability=implemented]
    jobs artifacts download - GET /jobs/{id}/artifacts [intent=direct_read availability=implemented]; flags: --id
    keys view - GET /keys [intent=direct_read availability=implemented]
    keys view-2 - GET /keys/{id} [intent=direct_read availability=implemented]; flags: --id
    metadata view - GET /metadata [intent=direct_read availability=implemented]
    namespaces view - GET /namespaces [intent=direct_read availability=implemented]
    namespaces limit-exclusions view - GET /namespaces/storage/limit_exclusions [intent=direct_read availability=implemented]
    namespaces view-2 - GET /namespaces/{id} [intent=direct_read availability=implemented]; flags: --id
    namespaces exists view - GET /namespaces/{id}/exists [intent=direct_read availability=implemented]; flags: --id
    namespaces gitlab-subscription view - GET /namespaces/{id}/gitlab_subscription [intent=direct_read availability=implemented]; flags: --id
    offline-exports view - GET /offline_exports [intent=direct_read availability=implemented]
    offline-exports view-2 - GET /offline_exports/{id} [intent=direct_read availability=implemented]; flags: --id
    pages domains view - GET /pages/domains [intent=direct_read availability=implemented]
    access access-requests view-2 - GET /projects/{id}/access_requests [intent=direct_read availability=implemented]; flags: --id
    access members view-3 - GET /projects/{id}/members [intent=direct_read availability=implemented]; flags: --id
    access all view-3 - GET /projects/{id}/members/all [intent=direct_read availability=implemented]; flags: --id
    access all view-4 - GET /projects/{id}/members/all/{user_id} [intent=direct_read availability=implemented]; flags: --id, --user-id
    access members view-4 - GET /projects/{id}/members/{user_id} [intent=direct_read availability=implemented]; flags: --id, --user-id
    access pages-access view - GET /projects/{id}/pages_access [intent=direct_read availability=implemented]; flags: --id
    registry repositories view - GET /registry/repositories/{id} [intent=direct_read availability=implemented]; flags: --id
    snippets view - GET /snippets [intent=direct_read availability=implemented]
    snippets all view - GET /snippets/all [intent=direct_read availability=implemented]
    snippets public view - GET /snippets/public [intent=direct_read availability=implemented]
    snippets view-2 - GET /snippets/{id} [intent=direct_read availability=implemented]; flags: --id
    snippets raw download - GET /snippets/{id}/files/{ref}/{file_path}/raw [intent=direct_read availability=implemented]; flags: --id, --ref, --file-path
    snippets raw download-2 - GET /snippets/{id}/raw [intent=direct_read availability=implemented]; flags: --id
    snippets user-agent-detail view - GET /snippets/{id}/user_agent_detail [intent=direct_read availability=implemented]; flags: --id
    topics view - GET /topics [intent=direct_read availability=implemented]
    topics view-2 - GET /topics/{id} [intent=direct_read availability=implemented]; flags: --id
    usage-data metric-definitions view - GET /usage_data/metric_definitions [intent=direct_read availability=implemented]
    usage-data non-sql-metrics view - GET /usage_data/non_sql_metrics [intent=direct_read availability=implemented]
    usage-data queries view - GET /usage_data/queries [intent=direct_read availability=implemented]
    usage-data service-ping view - GET /usage_data/service_ping [intent=direct_read availability=implemented]
    user-counts view - GET /user_counts [intent=direct_read availability=implemented]
    version view - GET /version [intent=direct_read availability=implemented]
    web-commits public-key view - GET /web_commits/public_key [intent=direct_read availability=implemented]
    jobs trace update - PATCH /jobs/{id}/trace [intent=reverse_etl availability=implemented write=patch_api_v4_jobs_id_trace]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    admin add create - POST /admin/clusters/add [intent=reverse_etl availability=implemented write=post_api_v4_admin_clusters_add]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    admin mark create - POST /admin/migrations/{timestamp}/mark [intent=reverse_etl availability=implemented write=post_api_v4_admin_migrations_timestamp_mark]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --timestamp
    applications create - POST /applications [intent=reverse_etl availability=implemented write=post_api_v4_applications]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    applications renew-secret create - POST /applications/{id}/renew-secret [intent=reverse_etl availability=implemented write=post_api_v4_applications_id_renew_secret]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    broadcast-messages create - POST /broadcast_messages [intent=reverse_etl availability=implemented write=post_api_v4_broadcast_messages]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    bulk-imports create - POST /bulk_imports [intent=reverse_etl availability=implemented write=post_api_v4_bulk_imports]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    bulk-imports cancel create - POST /bulk_imports/{import_id}/cancel [intent=reverse_etl availability=implemented write=post_api_v4_bulk_imports_import_id_cancel]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --import-id
    container-registry-event events create - POST /container_registry_event/events [intent=reverse_etl availability=implemented write=post_api_v4_container_registry_event_events]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    deploy-keys create - POST /deploy_keys [intent=reverse_etl availability=implemented write=post_api_v4_deploy_keys]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution
    feature-flags metrics create - POST /feature_flags/unleash/{project_id}/client/metrics [intent=reverse_etl availability=implemented write=post_api_v4_feature_flags_unleash_project_id_client_metrics]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --project-id
    feature-flags register create - POST /feature_flags/unleash/{project_id}/client/register [intent=reverse_etl availability=implemented write=post_api_v4_feature_flags_unleash_project_id_client_register]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --project-id
    features create - POST /features/{name} [intent=reverse_etl availability=implemented write=post_api_v4_features_name]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --name
    geo graphql create - POST /geo/node_proxy/{id}/graphql [intent=reverse_etl availability=implemented write=post_api_v4_geo_node_proxy_id_graphql]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    geo info-refs-receive-pack create - POST /geo/proxy_git_ssh/info_refs_receive_pack [intent=reverse_etl availability=implemented write=post_api_v4_geo_proxy_git_ssh_info_refs_receive_pack]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    geo info-refs-upload-pack create - POST /geo/proxy_git_ssh/info_refs_upload_pack [intent=reverse_etl availability=implemented write=post_api_v4_geo_proxy_git_ssh_info_refs_upload_pack]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    geo receive-pack create - POST /geo/proxy_git_ssh/receive_pack [intent=reverse_etl availability=implemented write=post_api_v4_geo_proxy_git_ssh_receive_pack]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    geo upload-pack create - POST /geo/proxy_git_ssh/upload_pack [intent=reverse_etl availability=implemented write=post_api_v4_geo_proxy_git_ssh_upload_pack]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    geo status create - POST /geo/status [intent=reverse_etl availability=implemented write=post_api_v4_geo_status]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    glql create - POST /glql [intent=reverse_etl availability=implemented write=post_api_v4_glql]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    access access-requests create - POST /groups/{id}/access_requests [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_access_requests]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    access members create - POST /groups/{id}/members [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_members]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    access approve-all create - POST /groups/{id}/members/approve_all [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_members_approve_all]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    access override create - POST /groups/{id}/members/{user_id}/override [intent=reverse_etl availability=implemented write=post_api_v4_groups_id_members_user_id_override]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --user-id
    hooks create - POST /hooks [intent=reverse_etl availability=implemented write=post_api_v4_hooks]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    hooks create-2 - POST /hooks/{hook_id} [intent=reverse_etl availability=implemented write=post_api_v4_hooks_hook_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --hook-id
    import bitbucket create - POST /import/bitbucket [intent=reverse_etl availability=implemented write=post_api_v4_import_bitbucket]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    import bitbucket-server create - POST /import/bitbucket_server [intent=reverse_etl availability=implemented write=post_api_v4_import_bitbucket_server]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    import github create - POST /import/github [intent=reverse_etl availability=implemented write=post_api_v4_import_github]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    import cancel create - POST /import/github/cancel [intent=reverse_etl availability=implemented write=post_api_v4_import_github_cancel]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    import gists create - POST /import/github/gists [intent=reverse_etl availability=implemented write=post_api_v4_import_github_gists]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    integrations subscriptions create - POST /integrations/jira_connect/subscriptions [intent=reverse_etl availability=implemented write=post_api_v4_integrations_jira_connect_subscriptions]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    integrations subscriptions create-2 - POST /integrations/jira_forge/subscriptions [intent=reverse_etl availability=implemented write=post_api_v4_integrations_jira_forge_subscriptions]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    integrations events create - POST /integrations/slack/events [intent=reverse_etl availability=implemented write=post_api_v4_integrations_slack_events]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    integrations interactions create - POST /integrations/slack/interactions [intent=reverse_etl availability=implemented write=post_api_v4_integrations_slack_interactions]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    integrations options create - POST /integrations/slack/options [intent=reverse_etl availability=implemented write=post_api_v4_integrations_slack_options]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    jobs request create - POST /jobs/request [intent=reverse_etl availability=implemented write=post_api_v4_jobs_request]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    jobs artifacts create - POST /jobs/{id}/artifacts [intent=reverse_etl availability=implemented write=post_api_v4_jobs_id_artifacts]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    jobs authorize create - POST /jobs/{id}/artifacts/authorize [intent=reverse_etl availability=implemented write=post_api_v4_jobs_id_artifacts_authorize]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    markdown create - POST /markdown [intent=reverse_etl availability=implemented write=post_api_v4_markdown]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    namespaces limit-exclusion create - POST /namespaces/{id}/storage/limit_exclusion [intent=reverse_etl availability=implemented write=post_api_v4_namespaces_id_storage_limit_exclusion]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    offline-exports create - POST /offline_exports [intent=reverse_etl availability=implemented write=post_api_v4_offline_exports]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    offline-imports create - POST /offline_imports [intent=reverse_etl availability=implemented write=post_api_v4_offline_imports]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    organizations create - POST /organizations [intent=reverse_etl availability=implemented write=post_api_v4_organizations]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    access access-requests create-2 - POST /projects/{id}/access_requests [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_access_requests]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    access import-project-members create - POST /projects/{id}/import_project_members/{project_id} [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_import_project_members_project_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --project-id
    access members create-2 - POST /projects/{id}/members [intent=reverse_etl availability=implemented write=post_api_v4_projects_id_members]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    slack trigger create - POST /slack/trigger [intent=reverse_etl availability=implemented write=post_api_v4_slack_trigger]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    snippets create - POST /snippets [intent=reverse_etl availability=implemented write=post_api_v4_snippets]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    topics create - POST /topics [intent=reverse_etl availability=implemented write=post_api_v4_topics]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    topics merge create - POST /topics/merge [intent=reverse_etl availability=implemented write=post_api_v4_topics_merge]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    usage-data increment-counter create - POST /usage_data/increment_counter [intent=reverse_etl availability=implemented write=post_api_v4_usage_data_increment_counter]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    usage-data increment-unique-users create - POST /usage_data/increment_unique_users [intent=reverse_etl availability=implemented write=post_api_v4_usage_data_increment_unique_users]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    usage-data track-event create - POST /usage_data/track_event [intent=reverse_etl availability=implemented write=post_api_v4_usage_data_track_event]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    usage-data track-events create - POST /usage_data/track_events [intent=reverse_etl availability=implemented write=post_api_v4_usage_data_track_events]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution
    admin pause set - PUT /admin/batched_background_migrations/{id}/pause [intent=reverse_etl availability=implemented write=put_api_v4_admin_batched_background_migrations_id_pause]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    admin resume set - PUT /admin/batched_background_migrations/{id}/resume [intent=reverse_etl availability=implemented write=put_api_v4_admin_batched_background_migrations_id_resume]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    admin restart set - PUT /admin/batched_background_operations/{id}/restart [intent=reverse_etl availability=implemented write=put_api_v4_admin_batched_background_operations_id_restart]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    admin stop set - PUT /admin/batched_background_operations/{id}/stop [intent=reverse_etl availability=implemented write=put_api_v4_admin_batched_background_operations_id_stop]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    admin clusters set - PUT /admin/clusters/{cluster_id} [intent=reverse_etl availability=implemented write=put_api_v4_admin_clusters_cluster_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --cluster-id
    application appearance set - PUT /application/appearance [intent=reverse_etl availability=implemented write=put_api_v4_application_appearance]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    application plan-limits set - PUT /application/plan_limits [intent=reverse_etl availability=implemented write=put_api_v4_application_plan_limits]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    broadcast-messages set - PUT /broadcast_messages/{id} [intent=reverse_etl availability=implemented write=put_api_v4_broadcast_messages_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    access approve set - PUT /groups/{id}/access_requests/{user_id}/approve [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_access_requests_user_id_approve]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --user-id
    access approve set-2 - PUT /groups/{id}/members/{member_id}/approve [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_members_member_id_approve]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --member-id
    access members set - PUT /groups/{id}/members/{user_id} [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_members_user_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --user-id
    access state set - PUT /groups/{id}/members/{user_id}/state [intent=reverse_etl availability=implemented write=put_api_v4_groups_id_members_user_id_state]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --user-id
    hooks set - PUT /hooks/{hook_id} [intent=reverse_etl availability=implemented write=put_api_v4_hooks_hook_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --hook-id
    hooks custom-headers set - PUT /hooks/{hook_id}/custom_headers/{key} [intent=reverse_etl availability=implemented write=put_api_v4_hooks_hook_id_custom_headers_key]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --hook-id, --key
    integrations installation set - PUT /integrations/jira_forge/installation [intent=reverse_etl availability=implemented write=put_api_v4_integrations_jira_forge_installation]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    jobs set - PUT /jobs/{id} [intent=reverse_etl availability=implemented write=put_api_v4_jobs_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    namespaces set - PUT /namespaces/{id} [intent=reverse_etl availability=implemented write=put_api_v4_namespaces_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: admin or access-control GitLab mutation is blocked by default pending elevated-scope policy and typed confirmation; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    access approve set-3 - PUT /projects/{id}/access_requests/{user_id}/approve [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_access_requests_user_id_approve]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --user-id
    access members set-2 - PUT /projects/{id}/members/{user_id} [intent=reverse_etl availability=implemented write=put_api_v4_projects_id_members_user_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: sensitive GitLab mutation is blocked by default pending redaction, input-mode, approval, and typed confirmation policy; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id, --user-id
    snippets set - PUT /snippets/{id} [intent=reverse_etl availability=implemented write=put_api_v4_snippets_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    suggestions batch-apply set - PUT /suggestions/batch_apply [intent=reverse_etl availability=implemented write=put_api_v4_suggestions_batch_apply]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution
    suggestions apply set - PUT /suggestions/{id}/apply [intent=reverse_etl availability=implemented write=put_api_v4_suggestions_id_apply]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
    topics set - PUT /topics/{id} [intent=reverse_etl availability=implemented write=put_api_v4_topics_id]; approval: reverse ETL writes require plan, preview, approval, execute.; risk: GitLab mutation candidate is inventoried for reverse ETL but blocked until write schema, preview, approval, and execution policy are declared; notes: typed operation; use reverse ETL preview and approval before execution; flags: --id
  Help topics:
    authentication - Use saved Polymetrics credentials or environment/stdin-based credential loading; never pass token values in prompts.
    writes - GitLab write-like commands are either planned reverse ETL operations or blocked-by-default safety surfaces; planned writes must use reverse ETL plan, preview, approval, execute before dispatch.
    binary-downloads - Artifacts, archives, secure files, and release assets remain disabled until bounded size and output destination policies exist.
    raw-api - Generic raw API commands are intentionally disallowed; use typed streams, direct reads, or future typed write actions instead.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect gitlab

  # Inspect as structured JSON
  pm connectors inspect gitlab --json

AGENT WORKFLOW
  - Run pm connectors inspect gitlab before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
