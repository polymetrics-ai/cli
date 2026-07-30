# pm connectors inspect jira

```text
NAME
  pm connectors inspect jira - Jira connector manual

SYNOPSIS
  pm connectors inspect jira
  pm connectors inspect jira --json
  pm credentials add <name> --connector jira [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Jira Cloud issues, projects, and users; exposes bounded JSON direct reads; and declares approval-gated reverse ETL actions from the official Jira Cloud REST API v3 ledger (269 executable write actions, 616 operations accounted).

ICON
  asset: icons/jira.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://developer.atlassian.com/changelog/#

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  email
  api_token (secret)

ETL STREAMS
  issues:
    primary key: id
    cursor: updated
    fields: assignee(), created(), id(), issuetype(), key(), priority(), project(), reporter(), self(), status(), summary(), updated()
  projects:
    primary key: id
    fields: id(), isPrivate(), key(), name(), projectTypeKey(), self(), simplified(), style()
  users:
    primary key: accountId
    fields: accountId(), accountType(), active(), displayName(), emailAddress(), self()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  set_banner:
    endpoint: PUT /rest/api/3/announcementBanner
    risk: Mutates Jira data/configuration for operation `setBanner`.
  update_multiple_custom_field_values:
    endpoint: POST /rest/api/3/app/field/value
    risk: Mutates Jira data/configuration for operation `updateMultipleCustomFieldValues`.
  update_custom_field_configuration:
    endpoint: PUT /rest/api/3/app/field/{{ record.path_fieldIdOrKey }}/context/configuration
    required fields: path_fieldIdOrKey, configurations
    risk: Mutates Jira data/configuration for operation `updateCustomFieldConfiguration`.
  update_custom_field_value:
    endpoint: PUT /rest/api/3/app/field/{{ record.path_fieldIdOrKey }}/value
    required fields: path_fieldIdOrKey
    risk: Mutates Jira data/configuration for operation `updateCustomFieldValue`.
  set_application_property:
    endpoint: PUT /rest/api/3/application-properties/{{ record.path_id }}
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `setApplicationProperty`.
  remove_attachment:
    endpoint: DELETE /rest/api/3/attachment/{{ record.path_id }}
    required fields: path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `removeAttachment`.
  submit_bulk_delete:
    endpoint: POST /rest/api/3/bulk/issues/delete
    required fields: selectedIssueIdsOrKeys
    risk: Mutates Jira data/configuration with destructive semantics for operation `submitBulkDelete`.
  submit_bulk_edit:
    endpoint: POST /rest/api/3/bulk/issues/fields
    required fields: editedFieldsInput, selectedActions, selectedIssueIdsOrKeys
    risk: Mutates Jira data/configuration for operation `submitBulkEdit`.
  submit_bulk_move:
    endpoint: POST /rest/api/3/bulk/issues/move
    required fields: targetToMultipleSourceMapping
    risk: Mutates Jira data/configuration for operation `submitBulkMove`.
  submit_bulk_transition:
    endpoint: POST /rest/api/3/bulk/issues/transition
    required fields: bulkTransitionInputs
    risk: Mutates Jira data/configuration for operation `submitBulkTransition`.
  submit_bulk_unwatch:
    endpoint: POST /rest/api/3/bulk/issues/unwatch
    required fields: selectedIssueIdsOrKeys
    risk: Mutates Jira data/configuration for operation `submitBulkUnwatch`.
  submit_bulk_watch:
    endpoint: POST /rest/api/3/bulk/issues/watch
    required fields: selectedIssueIdsOrKeys
    risk: Mutates Jira data/configuration for operation `submitBulkWatch`.
  delete_comment_property:
    endpoint: DELETE /rest/api/3/comment/{{ record.path_commentId }}/properties/{{ record.path_propertyKey }}
    required fields: path_commentId, path_propertyKey
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteCommentProperty`.
  create_component:
    endpoint: POST /rest/api/3/component
    risk: Mutates Jira data/configuration for operation `createComponent`.
  delete_component:
    endpoint: DELETE /rest/api/3/component/{{ record.path_id }}
    required fields: path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteComponent`.
  update_component:
    endpoint: PUT /rest/api/3/component/{{ record.path_id }}
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `updateComponent`.
  create_field_association_scheme:
    endpoint: POST /rest/api/3/config/fieldschemes
    required fields: name
    risk: Mutates Jira data/configuration for operation `createFieldAssociationScheme`.
  delete_field_association_scheme:
    endpoint: DELETE /rest/api/3/config/fieldschemes/{{ record.path_id }}
    required fields: path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteFieldAssociationScheme`.
  update_field_association_scheme:
    endpoint: PUT /rest/api/3/config/fieldschemes/{{ record.path_id }}
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `updateFieldAssociationScheme`.
  clone_field_association_scheme:
    endpoint: POST /rest/api/3/config/fieldschemes/{{ record.path_id }}/clone
    required fields: path_id, name
    risk: Mutates Jira data/configuration for operation `cloneFieldAssociationScheme`.
  select_time_tracking_implementation:
    endpoint: PUT /rest/api/3/configuration/timetracking
    required fields: key
    risk: Mutates Jira data/configuration for operation `selectTimeTrackingImplementation`.
  set_shared_time_tracking_configuration:
    endpoint: PUT /rest/api/3/configuration/timetracking/options
    required fields: defaultUnit, timeFormat, workingDaysPerWeek, workingHoursPerDay
    risk: Mutates Jira data/configuration for operation `setSharedTimeTrackingConfiguration`.
  create_dashboard:
    endpoint: POST /rest/api/3/dashboard
    required fields: editPermissions, name, sharePermissions
    risk: Mutates Jira data/configuration for operation `createDashboard`.
  bulk_edit_dashboards:
    endpoint: PUT /rest/api/3/dashboard/bulk/edit
    required fields: action, entityIds
    risk: Mutates Jira data/configuration for operation `bulkEditDashboards`.
  add_gadget:
    endpoint: POST /rest/api/3/dashboard/{{ record.path_dashboardId }}/gadget
    required fields: path_dashboardId
    risk: Mutates Jira data/configuration for operation `addGadget`.
  remove_gadget:
    endpoint: DELETE /rest/api/3/dashboard/{{ record.path_dashboardId }}/gadget/{{ record.path_gadgetId }}
    required fields: path_dashboardId, path_gadgetId
    risk: Permanently deletes or removes Jira data/configuration for operation `removeGadget`.
  update_gadget:
    endpoint: PUT /rest/api/3/dashboard/{{ record.path_dashboardId }}/gadget/{{ record.path_gadgetId }}
    required fields: path_dashboardId, path_gadgetId
    risk: Mutates Jira data/configuration for operation `updateGadget`.
  delete_dashboard_item_property:
    endpoint: DELETE /rest/api/3/dashboard/{{ record.path_dashboardId }}/items/{{ record.path_itemId }}/properties/{{ record.path_propertyKey }}
    required fields: path_dashboardId, path_itemId, path_propertyKey
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteDashboardItemProperty`.
  delete_dashboard:
    endpoint: DELETE /rest/api/3/dashboard/{{ record.path_id }}
    required fields: path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteDashboard`.
  update_dashboard:
    endpoint: PUT /rest/api/3/dashboard/{{ record.path_id }}
    required fields: path_id, editPermissions, name, sharePermissions
    risk: Mutates Jira data/configuration for operation `updateDashboard`.
  copy_dashboard:
    endpoint: POST /rest/api/3/dashboard/{{ record.path_id }}/copy
    required fields: path_id, editPermissions, name, sharePermissions
    risk: Mutates Jira data/configuration for operation `copyDashboard`.
  create_custom_field:
    endpoint: POST /rest/api/3/field
    required fields: name, type
    risk: Mutates Jira data/configuration for operation `createCustomField`.
  remove_associations:
    endpoint: DELETE /rest/api/3/field/association
    required fields: associationContexts, fields
    risk: Permanently deletes or removes Jira data/configuration for operation `removeAssociations`.
  create_associations:
    endpoint: PUT /rest/api/3/field/association
    required fields: associationContexts, fields
    risk: Mutates Jira data/configuration for operation `createAssociations`.
  update_custom_field:
    endpoint: PUT /rest/api/3/field/{{ record.path_fieldId }}
    required fields: path_fieldId
    risk: Mutates Jira data/configuration for operation `updateCustomField`.
  create_custom_field_context:
    endpoint: POST /rest/api/3/field/{{ record.path_fieldId }}/context
    required fields: path_fieldId, name
    risk: Mutates Jira data/configuration for operation `createCustomFieldContext`.
  delete_custom_field_context:
    endpoint: DELETE /rest/api/3/field/{{ record.path_fieldId }}/context/{{ record.path_contextId }}
    required fields: path_fieldId, path_contextId
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteCustomFieldContext`.
  update_custom_field_context:
    endpoint: PUT /rest/api/3/field/{{ record.path_fieldId }}/context/{{ record.path_contextId }}
    required fields: path_fieldId, path_contextId
    risk: Mutates Jira data/configuration for operation `updateCustomFieldContext`.
  add_issue_types_to_context:
    endpoint: PUT /rest/api/3/field/{{ record.path_fieldId }}/context/{{ record.path_contextId }}/issuetype
    required fields: path_fieldId, path_contextId, issueTypeIds
    risk: Mutates Jira data/configuration for operation `addIssueTypesToContext`.
  remove_issue_types_from_context:
    endpoint: POST /rest/api/3/field/{{ record.path_fieldId }}/context/{{ record.path_contextId }}/issuetype/remove
    required fields: path_fieldId, path_contextId, issueTypeIds
    risk: Mutates Jira data/configuration with destructive semantics for operation `removeIssueTypesFromContext`.
  create_custom_field_option:
    endpoint: POST /rest/api/3/field/{{ record.path_fieldId }}/context/{{ record.path_contextId }}/option
    required fields: path_fieldId, path_contextId
    risk: Mutates Jira data/configuration for operation `createCustomFieldOption`.
  update_custom_field_option:
    endpoint: PUT /rest/api/3/field/{{ record.path_fieldId }}/context/{{ record.path_contextId }}/option
    required fields: path_fieldId, path_contextId
    risk: Mutates Jira data/configuration for operation `updateCustomFieldOption`.
  reorder_custom_field_options:
    endpoint: PUT /rest/api/3/field/{{ record.path_fieldId }}/context/{{ record.path_contextId }}/option/move
    required fields: path_fieldId, path_contextId, customFieldOptionIds
    risk: Mutates Jira data/configuration for operation `reorderCustomFieldOptions`.
  delete_custom_field_option:
    endpoint: DELETE /rest/api/3/field/{{ record.path_fieldId }}/context/{{ record.path_contextId }}/option/{{ record.path_optionId }}
    required fields: path_fieldId, path_contextId, path_optionId
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteCustomFieldOption`.
  replace_custom_field_option:
    endpoint: DELETE /rest/api/3/field/{{ record.path_fieldId }}/context/{{ record.path_contextId }}/option/{{ record.path_optionId }}/issue
    required fields: path_fieldId, path_optionId, path_contextId
    risk: Permanently deletes or removes Jira data/configuration for operation `replaceCustomFieldOption`.
  assign_projects_to_custom_field_context:
    endpoint: PUT /rest/api/3/field/{{ record.path_fieldId }}/context/{{ record.path_contextId }}/project
    required fields: path_fieldId, path_contextId, projectIds
    risk: Mutates Jira data/configuration for operation `assignProjectsToCustomFieldContext`.
  remove_custom_field_context_from_projects:
    endpoint: POST /rest/api/3/field/{{ record.path_fieldId }}/context/{{ record.path_contextId }}/project/remove
    required fields: path_fieldId, path_contextId, projectIds
    risk: Mutates Jira data/configuration with destructive semantics for operation `removeCustomFieldContextFromProjects`.
  create_issue_field_option:
    endpoint: POST /rest/api/3/field/{{ record.path_fieldKey }}/option
    required fields: path_fieldKey, value
    risk: Mutates Jira data/configuration for operation `createIssueFieldOption`.
  delete_issue_field_option:
    endpoint: DELETE /rest/api/3/field/{{ record.path_fieldKey }}/option/{{ record.path_optionId }}
    required fields: path_fieldKey, path_optionId
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteIssueFieldOption`.
  update_issue_field_option:
    endpoint: PUT /rest/api/3/field/{{ record.path_fieldKey }}/option/{{ record.path_optionId }}
    required fields: path_fieldKey, path_optionId, id, value
    risk: Mutates Jira data/configuration for operation `updateIssueFieldOption`.
  replace_issue_field_option:
    endpoint: DELETE /rest/api/3/field/{{ record.path_fieldKey }}/option/{{ record.path_optionId }}/issue
    required fields: path_fieldKey, path_optionId
    risk: Permanently deletes or removes Jira data/configuration for operation `replaceIssueFieldOption`.
  delete_custom_field:
    endpoint: DELETE /rest/api/3/field/{{ record.path_id }}
    required fields: path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteCustomField`.
  restore_custom_field:
    endpoint: POST /rest/api/3/field/{{ record.path_id }}/restore
    required fields: path_id
    risk: Mutates Jira data/configuration with destructive semantics for operation `restoreCustomField`.
  trash_custom_field:
    endpoint: POST /rest/api/3/field/{{ record.path_id }}/trash
    required fields: path_id
    risk: Mutates Jira data/configuration with destructive semantics for operation `trashCustomField`.
  create_filter:
    endpoint: POST /rest/api/3/filter
    required fields: name
    risk: Mutates Jira data/configuration for operation `createFilter`.
  set_default_share_scope:
    endpoint: PUT /rest/api/3/filter/defaultShareScope
    required fields: scope
    risk: Mutates Jira data/configuration for operation `setDefaultShareScope`.
  delete_filter:
    endpoint: DELETE /rest/api/3/filter/{{ record.path_id }}
    required fields: path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteFilter`.
  update_filter:
    endpoint: PUT /rest/api/3/filter/{{ record.path_id }}
    required fields: path_id, name
    risk: Mutates Jira data/configuration for operation `updateFilter`.
  reset_columns:
    endpoint: DELETE /rest/api/3/filter/{{ record.path_id }}/columns
    required fields: path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `resetColumns`.
  set_columns:
    endpoint: PUT /rest/api/3/filter/{{ record.path_id }}/columns
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `setColumns`.
  delete_favourite_for_filter:
    endpoint: DELETE /rest/api/3/filter/{{ record.path_id }}/favourite
    required fields: path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteFavouriteForFilter`.
  set_favourite_for_filter:
    endpoint: PUT /rest/api/3/filter/{{ record.path_id }}/favourite
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `setFavouriteForFilter`.
  change_filter_owner:
    endpoint: PUT /rest/api/3/filter/{{ record.path_id }}/owner
    required fields: path_id, accountId
    risk: Mutates Jira data/configuration for operation `changeFilterOwner`.
  add_share_permission:
    endpoint: POST /rest/api/3/filter/{{ record.path_id }}/permission
    required fields: path_id, type
    risk: Mutates Jira data/configuration for operation `addSharePermission`.
  delete_share_permission:
    endpoint: DELETE /rest/api/3/filter/{{ record.path_id }}/permission/{{ record.path_permissionId }}
    required fields: path_id, path_permissionId
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteSharePermission`.
  bulk_pin_unpin_projects_async:
    endpoint: POST /rest/api/3/forge/panel/action/bulk/async
    required fields: moduleId, projectList
    risk: Mutates Jira data/configuration for operation `bulkPinUnpinProjectsAsync`.
  remove_group:
    endpoint: DELETE /rest/api/3/group
    risk: Permanently deletes or removes Jira data/configuration for operation `removeGroup`.
  create_group:
    endpoint: POST /rest/api/3/group
    required fields: name
    risk: Mutates Jira data/configuration for operation `createGroup`.
  remove_user_from_group:
    endpoint: DELETE /rest/api/3/group/user?accountId={{ record.query_accountId }}
    required fields: query_accountId
    risk: Permanently deletes or removes Jira data/configuration for operation `removeUserFromGroup`.
  add_user_to_group:
    endpoint: POST /rest/api/3/group/user
    risk: Mutates Jira data/configuration for operation `addUserToGroup`.
  create_issue:
    endpoint: POST /rest/api/3/issue
    risk: Mutates Jira data/configuration for operation `createIssue`.
  archive_issues_async:
    endpoint: POST /rest/api/3/issue/archive
    risk: Mutates Jira data/configuration with destructive semantics for operation `archiveIssuesAsync`.
  archive_issues:
    endpoint: PUT /rest/api/3/issue/archive
    risk: Mutates Jira data/configuration with destructive semantics for operation `archiveIssues`.
  create_issues:
    endpoint: POST /rest/api/3/issue/bulk
    risk: Mutates Jira data/configuration for operation `createIssues`.
  bulk_fetch_issues:
    endpoint: POST /rest/api/3/issue/bulkfetch
    required fields: issueIdsOrKeys
    risk: Mutates Jira data/configuration for operation `bulkFetchIssues`.
  bulk_set_issues_properties_list:
    endpoint: POST /rest/api/3/issue/properties
    risk: Mutates Jira data/configuration for operation `bulkSetIssuesPropertiesList`.
  bulk_set_issue_properties_by_issue:
    endpoint: POST /rest/api/3/issue/properties/multi
    risk: Mutates Jira data/configuration for operation `bulkSetIssuePropertiesByIssue`.
  bulk_delete_issue_property:
    endpoint: DELETE /rest/api/3/issue/properties/{{ record.path_propertyKey }}
    required fields: path_propertyKey
    optional fields: currentValue, entityIds
    risk: Permanently deletes or removes Jira data/configuration for operation `bulkDeleteIssueProperty`.
  bulk_set_issue_property:
    endpoint: PUT /rest/api/3/issue/properties/{{ record.path_propertyKey }}
    required fields: path_propertyKey
    risk: Mutates Jira data/configuration for operation `bulkSetIssueProperty`.
  unarchive_issues:
    endpoint: PUT /rest/api/3/issue/unarchive
    risk: Mutates Jira data/configuration with destructive semantics for operation `unarchiveIssues`.
  delete_issue:
    endpoint: DELETE /rest/api/3/issue/{{ record.path_issueIdOrKey }}
    required fields: path_issueIdOrKey
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteIssue`.
  edit_issue:
    endpoint: PUT /rest/api/3/issue/{{ record.path_issueIdOrKey }}
    required fields: path_issueIdOrKey
    risk: Mutates Jira data/configuration for operation `editIssue`.
  assign_issue:
    endpoint: PUT /rest/api/3/issue/{{ record.path_issueIdOrKey }}/assignee
    required fields: path_issueIdOrKey
    risk: Mutates Jira data/configuration for operation `assignIssue`.
  add_attachment:
    endpoint: POST /rest/api/3/issue/{{ record.path_issueIdOrKey }}/attachments
    required fields: path_issueIdOrKey, file_path
    risk: Mutates Jira data/configuration for operation `addAttachment`.
  add_comment:
    endpoint: POST /rest/api/3/issue/{{ record.path_issueIdOrKey }}/comment
    required fields: path_issueIdOrKey
    risk: Mutates Jira data/configuration for operation `addComment`.
  delete_comment:
    endpoint: DELETE /rest/api/3/issue/{{ record.path_issueIdOrKey }}/comment/{{ record.path_id }}
    required fields: path_issueIdOrKey, path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteComment`.
  update_comment:
    endpoint: PUT /rest/api/3/issue/{{ record.path_issueIdOrKey }}/comment/{{ record.path_id }}
    required fields: path_issueIdOrKey, path_id
    risk: Mutates Jira data/configuration for operation `updateComment`.
  notify:
    endpoint: POST /rest/api/3/issue/{{ record.path_issueIdOrKey }}/notify
    required fields: path_issueIdOrKey
    risk: Mutates Jira data/configuration for operation `notify`.
  delete_issue_property:
    endpoint: DELETE /rest/api/3/issue/{{ record.path_issueIdOrKey }}/properties/{{ record.path_propertyKey }}
    required fields: path_issueIdOrKey, path_propertyKey
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteIssueProperty`.
  delete_remote_issue_link_by_global_id:
    endpoint: DELETE /rest/api/3/issue/{{ record.path_issueIdOrKey }}/remotelink?globalId={{ record.query_globalId }}
    required fields: path_issueIdOrKey, query_globalId
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteRemoteIssueLinkByGlobalId`.
  create_or_update_remote_issue_link:
    endpoint: POST /rest/api/3/issue/{{ record.path_issueIdOrKey }}/remotelink
    required fields: path_issueIdOrKey, object
    risk: Mutates Jira data/configuration for operation `createOrUpdateRemoteIssueLink`.
  delete_remote_issue_link_by_id:
    endpoint: DELETE /rest/api/3/issue/{{ record.path_issueIdOrKey }}/remotelink/{{ record.path_linkId }}
    required fields: path_issueIdOrKey, path_linkId
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteRemoteIssueLinkById`.
  update_remote_issue_link:
    endpoint: PUT /rest/api/3/issue/{{ record.path_issueIdOrKey }}/remotelink/{{ record.path_linkId }}
    required fields: path_issueIdOrKey, path_linkId, object
    risk: Mutates Jira data/configuration for operation `updateRemoteIssueLink`.
  do_transition:
    endpoint: POST /rest/api/3/issue/{{ record.path_issueIdOrKey }}/transitions
    required fields: path_issueIdOrKey
    risk: Mutates Jira data/configuration for operation `doTransition`.
  remove_vote:
    endpoint: DELETE /rest/api/3/issue/{{ record.path_issueIdOrKey }}/votes
    required fields: path_issueIdOrKey
    risk: Permanently deletes or removes Jira data/configuration for operation `removeVote`.
  add_vote:
    endpoint: POST /rest/api/3/issue/{{ record.path_issueIdOrKey }}/votes
    required fields: path_issueIdOrKey
    risk: Mutates Jira data/configuration for operation `addVote`.
  remove_watcher:
    endpoint: DELETE /rest/api/3/issue/{{ record.path_issueIdOrKey }}/watchers
    required fields: path_issueIdOrKey
    risk: Permanently deletes or removes Jira data/configuration for operation `removeWatcher`.
  bulk_delete_worklogs:
    endpoint: DELETE /rest/api/3/issue/{{ record.path_issueIdOrKey }}/worklog
    required fields: path_issueIdOrKey, ids
    risk: Permanently deletes or removes Jira data/configuration for operation `bulkDeleteWorklogs`.
  add_worklog:
    endpoint: POST /rest/api/3/issue/{{ record.path_issueIdOrKey }}/worklog
    required fields: path_issueIdOrKey
    risk: Mutates Jira data/configuration for operation `addWorklog`.
  bulk_move_worklogs:
    endpoint: POST /rest/api/3/issue/{{ record.path_issueIdOrKey }}/worklog/move
    required fields: path_issueIdOrKey
    risk: Mutates Jira data/configuration for operation `bulkMoveWorklogs`.
  delete_worklog:
    endpoint: DELETE /rest/api/3/issue/{{ record.path_issueIdOrKey }}/worklog/{{ record.path_id }}
    required fields: path_issueIdOrKey, path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteWorklog`.
  update_worklog:
    endpoint: PUT /rest/api/3/issue/{{ record.path_issueIdOrKey }}/worklog/{{ record.path_id }}
    required fields: path_issueIdOrKey, path_id
    risk: Mutates Jira data/configuration for operation `updateWorklog`.
  delete_worklog_property:
    endpoint: DELETE /rest/api/3/issue/{{ record.path_issueIdOrKey }}/worklog/{{ record.path_worklogId }}/properties/{{ record.path_propertyKey }}
    required fields: path_issueIdOrKey, path_worklogId, path_propertyKey
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteWorklogProperty`.
  link_issues:
    endpoint: POST /rest/api/3/issueLink
    required fields: inwardIssue, outwardIssue, type
    risk: Mutates Jira data/configuration for operation `linkIssues`.
  delete_issue_link:
    endpoint: DELETE /rest/api/3/issueLink/{{ record.path_linkId }}
    required fields: path_linkId
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteIssueLink`.
  create_issue_link_type:
    endpoint: POST /rest/api/3/issueLinkType
    risk: Mutates Jira data/configuration for operation `createIssueLinkType`.
  delete_issue_link_type:
    endpoint: DELETE /rest/api/3/issueLinkType/{{ record.path_issueLinkTypeId }}
    required fields: path_issueLinkTypeId
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteIssueLinkType`.
  update_issue_link_type:
    endpoint: PUT /rest/api/3/issueLinkType/{{ record.path_issueLinkTypeId }}
    required fields: path_issueLinkTypeId
    risk: Mutates Jira data/configuration for operation `updateIssueLinkType`.
  create_issue_security_scheme:
    endpoint: POST /rest/api/3/issuesecurityschemes
    required fields: name
    risk: Mutates Jira data/configuration for operation `createIssueSecurityScheme`.
  set_default_levels:
    endpoint: PUT /rest/api/3/issuesecurityschemes/level/default
    required fields: defaultValues
    risk: Mutates Jira data/configuration for operation `setDefaultLevels`.
  associate_schemes_to_projects:
    endpoint: PUT /rest/api/3/issuesecurityschemes/project
    required fields: projectId, schemeId
    risk: Mutates Jira data/configuration for operation `associateSchemesToProjects`.
  update_issue_security_scheme:
    endpoint: PUT /rest/api/3/issuesecurityschemes/{{ record.path_id }}
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `updateIssueSecurityScheme`.
  delete_security_scheme:
    endpoint: DELETE /rest/api/3/issuesecurityschemes/{{ record.path_schemeId }}
    required fields: path_schemeId
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteSecurityScheme`.
  add_security_level:
    endpoint: PUT /rest/api/3/issuesecurityschemes/{{ record.path_schemeId }}/level
    required fields: path_schemeId
    risk: Mutates Jira data/configuration for operation `addSecurityLevel`.
  remove_level:
    endpoint: DELETE /rest/api/3/issuesecurityschemes/{{ record.path_schemeId }}/level/{{ record.path_levelId }}
    required fields: path_schemeId, path_levelId
    risk: Permanently deletes or removes Jira data/configuration for operation `removeLevel`.
  update_security_level:
    endpoint: PUT /rest/api/3/issuesecurityschemes/{{ record.path_schemeId }}/level/{{ record.path_levelId }}
    required fields: path_schemeId, path_levelId
    risk: Mutates Jira data/configuration for operation `updateSecurityLevel`.
  add_security_level_members:
    endpoint: PUT /rest/api/3/issuesecurityschemes/{{ record.path_schemeId }}/level/{{ record.path_levelId }}/member
    required fields: path_schemeId, path_levelId, securitySchemeLevelMembers
    risk: Mutates Jira data/configuration for operation `addSecurityLevelMembers`.
  remove_member_from_security_level:
    endpoint: DELETE /rest/api/3/issuesecurityschemes/{{ record.path_schemeId }}/level/{{ record.path_levelId }}/member/{{ record.path_memberId }}
    required fields: path_schemeId, path_levelId, path_memberId
    risk: Permanently deletes or removes Jira data/configuration for operation `removeMemberFromSecurityLevel`.
  create_issue_type:
    endpoint: POST /rest/api/3/issuetype
    required fields: name
    risk: Mutates Jira data/configuration for operation `createIssueType`.
  delete_issue_type:
    endpoint: DELETE /rest/api/3/issuetype/{{ record.path_id }}
    required fields: path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteIssueType`.
  update_issue_type:
    endpoint: PUT /rest/api/3/issuetype/{{ record.path_id }}
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `updateIssueType`.
  delete_issue_type_property:
    endpoint: DELETE /rest/api/3/issuetype/{{ record.path_issueTypeId }}/properties/{{ record.path_propertyKey }}
    required fields: path_issueTypeId, path_propertyKey
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteIssueTypeProperty`.
  create_issue_type_scheme:
    endpoint: POST /rest/api/3/issuetypescheme
    required fields: issueTypeIds, name
    risk: Mutates Jira data/configuration for operation `createIssueTypeScheme`.
  assign_issue_type_scheme_to_project:
    endpoint: PUT /rest/api/3/issuetypescheme/project
    required fields: issueTypeSchemeId, projectId
    risk: Mutates Jira data/configuration for operation `assignIssueTypeSchemeToProject`.
  delete_issue_type_scheme:
    endpoint: DELETE /rest/api/3/issuetypescheme/{{ record.path_issueTypeSchemeId }}
    required fields: path_issueTypeSchemeId
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteIssueTypeScheme`.
  update_issue_type_scheme:
    endpoint: PUT /rest/api/3/issuetypescheme/{{ record.path_issueTypeSchemeId }}
    required fields: path_issueTypeSchemeId
    risk: Mutates Jira data/configuration for operation `updateIssueTypeScheme`.
  add_issue_types_to_issue_type_scheme:
    endpoint: PUT /rest/api/3/issuetypescheme/{{ record.path_issueTypeSchemeId }}/issuetype
    required fields: path_issueTypeSchemeId, issueTypeIds
    risk: Mutates Jira data/configuration for operation `addIssueTypesToIssueTypeScheme`.
  reorder_issue_types_in_issue_type_scheme:
    endpoint: PUT /rest/api/3/issuetypescheme/{{ record.path_issueTypeSchemeId }}/issuetype/move
    required fields: path_issueTypeSchemeId, issueTypeIds
    risk: Mutates Jira data/configuration for operation `reorderIssueTypesInIssueTypeScheme`.
  remove_issue_type_from_issue_type_scheme:
    endpoint: DELETE /rest/api/3/issuetypescheme/{{ record.path_issueTypeSchemeId }}/issuetype/{{ record.path_issueTypeId }}
    required fields: path_issueTypeSchemeId, path_issueTypeId
    risk: Permanently deletes or removes Jira data/configuration for operation `removeIssueTypeFromIssueTypeScheme`.
  create_issue_type_screen_scheme:
    endpoint: POST /rest/api/3/issuetypescreenscheme
    required fields: issueTypeMappings, name
    risk: Mutates Jira data/configuration for operation `createIssueTypeScreenScheme`.
  assign_issue_type_screen_scheme_to_project:
    endpoint: PUT /rest/api/3/issuetypescreenscheme/project
    risk: Mutates Jira data/configuration for operation `assignIssueTypeScreenSchemeToProject`.
  delete_issue_type_screen_scheme:
    endpoint: DELETE /rest/api/3/issuetypescreenscheme/{{ record.path_issueTypeScreenSchemeId }}
    required fields: path_issueTypeScreenSchemeId
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteIssueTypeScreenScheme`.
  update_issue_type_screen_scheme:
    endpoint: PUT /rest/api/3/issuetypescreenscheme/{{ record.path_issueTypeScreenSchemeId }}
    required fields: path_issueTypeScreenSchemeId
    risk: Mutates Jira data/configuration for operation `updateIssueTypeScreenScheme`.
  append_mappings_for_issue_type_screen_scheme:
    endpoint: PUT /rest/api/3/issuetypescreenscheme/{{ record.path_issueTypeScreenSchemeId }}/mapping
    required fields: path_issueTypeScreenSchemeId, issueTypeMappings
    risk: Mutates Jira data/configuration for operation `appendMappingsForIssueTypeScreenScheme`.
  update_default_screen_scheme:
    endpoint: PUT /rest/api/3/issuetypescreenscheme/{{ record.path_issueTypeScreenSchemeId }}/mapping/default
    required fields: path_issueTypeScreenSchemeId, screenSchemeId
    risk: Mutates Jira data/configuration for operation `updateDefaultScreenScheme`.
  remove_mappings_from_issue_type_screen_scheme:
    endpoint: POST /rest/api/3/issuetypescreenscheme/{{ record.path_issueTypeScreenSchemeId }}/mapping/remove
    required fields: path_issueTypeScreenSchemeId, issueTypeIds
    risk: Mutates Jira data/configuration with destructive semantics for operation `removeMappingsFromIssueTypeScreenScheme`.
  update_precomputations:
    endpoint: POST /rest/api/3/jql/function/computation
    risk: Mutates Jira data/configuration for operation `updatePrecomputations`.
  remove_preference:
    endpoint: DELETE /rest/api/3/mypreferences?key={{ record.query_key }}
    required fields: query_key
    risk: Permanently deletes or removes Jira data/configuration for operation `removePreference`.
  create_notification_scheme:
    endpoint: POST /rest/api/3/notificationscheme
    required fields: name
    risk: Mutates Jira data/configuration for operation `createNotificationScheme`.
  update_notification_scheme:
    endpoint: PUT /rest/api/3/notificationscheme/{{ record.path_id }}
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `updateNotificationScheme`.
  add_notifications:
    endpoint: PUT /rest/api/3/notificationscheme/{{ record.path_id }}/notification
    required fields: path_id, notificationSchemeEvents
    risk: Mutates Jira data/configuration for operation `addNotifications`.
  delete_notification_scheme:
    endpoint: DELETE /rest/api/3/notificationscheme/{{ record.path_notificationSchemeId }}
    required fields: path_notificationSchemeId
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteNotificationScheme`.
  remove_notification_from_notification_scheme:
    endpoint: DELETE /rest/api/3/notificationscheme/{{ record.path_notificationSchemeId }}/notification/{{ record.path_notificationId }}
    required fields: path_notificationSchemeId, path_notificationId
    risk: Permanently deletes or removes Jira data/configuration for operation `removeNotificationFromNotificationScheme`.
  create_permission_scheme:
    endpoint: POST /rest/api/3/permissionscheme
    required fields: name
    risk: Mutates Jira data/configuration for operation `createPermissionScheme`.
  delete_permission_scheme:
    endpoint: DELETE /rest/api/3/permissionscheme/{{ record.path_schemeId }}
    required fields: path_schemeId
    risk: Permanently deletes or removes Jira data/configuration for operation `deletePermissionScheme`.
  update_permission_scheme:
    endpoint: PUT /rest/api/3/permissionscheme/{{ record.path_schemeId }}
    required fields: path_schemeId, name
    risk: Mutates Jira data/configuration for operation `updatePermissionScheme`.
  create_permission_grant:
    endpoint: POST /rest/api/3/permissionscheme/{{ record.path_schemeId }}/permission
    required fields: path_schemeId
    risk: Mutates Jira data/configuration for operation `createPermissionGrant`.
  delete_permission_scheme_entity:
    endpoint: DELETE /rest/api/3/permissionscheme/{{ record.path_schemeId }}/permission/{{ record.path_permissionId }}
    required fields: path_schemeId, path_permissionId
    risk: Permanently deletes or removes Jira data/configuration for operation `deletePermissionSchemeEntity`.
  create_plan:
    endpoint: POST /rest/api/3/plans/plan
    required fields: issueSources, name, scheduling
    risk: Mutates Jira data/configuration for operation `createPlan`.
  archive_plan:
    endpoint: PUT /rest/api/3/plans/plan/{{ record.path_planId }}/archive
    required fields: path_planId
    risk: Mutates Jira data/configuration with destructive semantics for operation `archivePlan`.
  duplicate_plan:
    endpoint: POST /rest/api/3/plans/plan/{{ record.path_planId }}/duplicate
    required fields: path_planId, name
    risk: Mutates Jira data/configuration for operation `duplicatePlan`.
  add_atlassian_team:
    endpoint: POST /rest/api/3/plans/plan/{{ record.path_planId }}/team/atlassian
    required fields: path_planId, id, planningStyle
    risk: Mutates Jira data/configuration for operation `addAtlassianTeam`.
  remove_atlassian_team:
    endpoint: DELETE /rest/api/3/plans/plan/{{ record.path_planId }}/team/atlassian/{{ record.path_atlassianTeamId }}
    required fields: path_planId, path_atlassianTeamId
    risk: Permanently deletes or removes Jira data/configuration for operation `removeAtlassianTeam`.
  create_plan_only_team:
    endpoint: POST /rest/api/3/plans/plan/{{ record.path_planId }}/team/planonly
    required fields: path_planId, name, planningStyle
    risk: Mutates Jira data/configuration for operation `createPlanOnlyTeam`.
  delete_plan_only_team:
    endpoint: DELETE /rest/api/3/plans/plan/{{ record.path_planId }}/team/planonly/{{ record.path_planOnlyTeamId }}
    required fields: path_planId, path_planOnlyTeamId
    risk: Permanently deletes or removes Jira data/configuration for operation `deletePlanOnlyTeam`.
  trash_plan:
    endpoint: PUT /rest/api/3/plans/plan/{{ record.path_planId }}/trash
    required fields: path_planId
    risk: Mutates Jira data/configuration with destructive semantics for operation `trashPlan`.
  create_priority:
    endpoint: POST /rest/api/3/priority
    required fields: name, statusColor
    risk: Mutates Jira data/configuration for operation `createPriority`.
  set_default_priority:
    endpoint: PUT /rest/api/3/priority/default
    required fields: id
    risk: Mutates Jira data/configuration for operation `setDefaultPriority`.
  move_priorities:
    endpoint: PUT /rest/api/3/priority/move
    required fields: ids
    risk: Mutates Jira data/configuration for operation `movePriorities`.
  delete_priority:
    endpoint: DELETE /rest/api/3/priority/{{ record.path_id }}
    required fields: path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deletePriority`.
  update_priority:
    endpoint: PUT /rest/api/3/priority/{{ record.path_id }}
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `updatePriority`.
  create_priority_scheme:
    endpoint: POST /rest/api/3/priorityscheme
    required fields: defaultPriorityId, name, priorityIds
    risk: Mutates Jira data/configuration for operation `createPriorityScheme`.
  delete_priority_scheme:
    endpoint: DELETE /rest/api/3/priorityscheme/{{ record.path_schemeId }}
    required fields: path_schemeId
    risk: Permanently deletes or removes Jira data/configuration for operation `deletePriorityScheme`.
  update_priority_scheme:
    endpoint: PUT /rest/api/3/priorityscheme/{{ record.path_schemeId }}
    required fields: path_schemeId
    risk: Mutates Jira data/configuration for operation `updatePriorityScheme`.
  create_project:
    endpoint: POST /rest/api/3/project
    required fields: key, name
    risk: Mutates Jira data/configuration for operation `createProject`.
  create_project_with_custom_template:
    endpoint: POST /rest/api/3/project-template
    risk: Mutates Jira data/configuration for operation `createProjectWithCustomTemplate`.
  edit_template:
    endpoint: PUT /rest/api/3/project-template/edit-template
    risk: Mutates Jira data/configuration for operation `editTemplate`.
  remove_template:
    endpoint: DELETE /rest/api/3/project-template/remove-template?templateKey={{ record.query_templateKey }}
    required fields: query_templateKey
    risk: Permanently deletes or removes Jira data/configuration for operation `removeTemplate`.
  save_template:
    endpoint: POST /rest/api/3/project-template/save-template
    risk: Mutates Jira data/configuration for operation `saveTemplate`.
  delete_project:
    endpoint: DELETE /rest/api/3/project/{{ record.path_projectIdOrKey }}
    required fields: path_projectIdOrKey
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteProject`.
  update_project:
    endpoint: PUT /rest/api/3/project/{{ record.path_projectIdOrKey }}
    required fields: path_projectIdOrKey
    risk: Mutates Jira data/configuration for operation `updateProject`.
  archive_project:
    endpoint: POST /rest/api/3/project/{{ record.path_projectIdOrKey }}/archive
    required fields: path_projectIdOrKey
    risk: Mutates Jira data/configuration with destructive semantics for operation `archiveProject`.
  update_project_avatar:
    endpoint: PUT /rest/api/3/project/{{ record.path_projectIdOrKey }}/avatar
    required fields: path_projectIdOrKey, id
    risk: Mutates Jira data/configuration for operation `updateProjectAvatar`.
  delete_project_avatar:
    endpoint: DELETE /rest/api/3/project/{{ record.path_projectIdOrKey }}/avatar/{{ record.path_id }}
    required fields: path_projectIdOrKey, path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteProjectAvatar`.
  remove_default_project_classification:
    endpoint: DELETE /rest/api/3/project/{{ record.path_projectIdOrKey }}/classification-level/default
    required fields: path_projectIdOrKey
    risk: Permanently deletes or removes Jira data/configuration for operation `removeDefaultProjectClassification`.
  update_default_project_classification:
    endpoint: PUT /rest/api/3/project/{{ record.path_projectIdOrKey }}/classification-level/default
    required fields: path_projectIdOrKey, id
    risk: Mutates Jira data/configuration for operation `updateDefaultProjectClassification`.
  delete_project_asynchronously:
    endpoint: POST /rest/api/3/project/{{ record.path_projectIdOrKey }}/delete
    required fields: path_projectIdOrKey
    risk: Mutates Jira data/configuration with destructive semantics for operation `deleteProjectAsynchronously`.
  toggle_feature_for_project:
    endpoint: PUT /rest/api/3/project/{{ record.path_projectIdOrKey }}/features/{{ record.path_featureKey }}
    required fields: path_projectIdOrKey, path_featureKey
    risk: Mutates Jira data/configuration for operation `toggleFeatureForProject`.
  delete_project_property:
    endpoint: DELETE /rest/api/3/project/{{ record.path_projectIdOrKey }}/properties/{{ record.path_propertyKey }}
    required fields: path_projectIdOrKey, path_propertyKey
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteProjectProperty`.
  restore:
    endpoint: POST /rest/api/3/project/{{ record.path_projectIdOrKey }}/restore
    required fields: path_projectIdOrKey
    risk: Mutates Jira data/configuration with destructive semantics for operation `restore`.
  delete_actor:
    endpoint: DELETE /rest/api/3/project/{{ record.path_projectIdOrKey }}/role/{{ record.path_id }}
    required fields: path_projectIdOrKey, path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteActor`.
  add_actor_users:
    endpoint: POST /rest/api/3/project/{{ record.path_projectIdOrKey }}/role/{{ record.path_id }}
    required fields: path_projectIdOrKey, path_id
    risk: Mutates Jira data/configuration for operation `addActorUsers`.
  set_actors:
    endpoint: PUT /rest/api/3/project/{{ record.path_projectIdOrKey }}/role/{{ record.path_id }}
    required fields: path_projectIdOrKey, path_id
    risk: Mutates Jira data/configuration for operation `setActors`.
  update_project_email:
    endpoint: PUT /rest/api/3/project/{{ record.path_projectId }}/email
    required fields: path_projectId
    risk: Mutates Jira data/configuration for operation `updateProjectEmail`.
  assign_permission_scheme:
    endpoint: PUT /rest/api/3/project/{{ record.path_projectKeyOrId }}/permissionscheme
    required fields: path_projectKeyOrId, id
    risk: Mutates Jira data/configuration for operation `assignPermissionScheme`.
  create_project_category:
    endpoint: POST /rest/api/3/projectCategory
    risk: Mutates Jira data/configuration for operation `createProjectCategory`.
  remove_project_category:
    endpoint: DELETE /rest/api/3/projectCategory/{{ record.path_id }}
    required fields: path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `removeProjectCategory`.
  update_project_category:
    endpoint: PUT /rest/api/3/projectCategory/{{ record.path_id }}
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `updateProjectCategory`.
  redact:
    endpoint: POST /rest/api/3/redact
    risk: Mutates Jira data/configuration for operation `redact`.
  create_resolution:
    endpoint: POST /rest/api/3/resolution
    required fields: name
    risk: Mutates Jira data/configuration for operation `createResolution`.
  set_default_resolution:
    endpoint: PUT /rest/api/3/resolution/default
    required fields: id
    risk: Mutates Jira data/configuration for operation `setDefaultResolution`.
  move_resolutions:
    endpoint: PUT /rest/api/3/resolution/move
    required fields: ids
    risk: Mutates Jira data/configuration for operation `moveResolutions`.
  delete_resolution:
    endpoint: DELETE /rest/api/3/resolution/{{ record.path_id }}?replaceWith={{ record.query_replaceWith }}
    required fields: path_id, query_replaceWith
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteResolution`.
  update_resolution:
    endpoint: PUT /rest/api/3/resolution/{{ record.path_id }}
    required fields: path_id, name
    risk: Mutates Jira data/configuration for operation `updateResolution`.
  create_project_role:
    endpoint: POST /rest/api/3/role
    risk: Mutates Jira data/configuration for operation `createProjectRole`.
  delete_project_role:
    endpoint: DELETE /rest/api/3/role/{{ record.path_id }}
    required fields: path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteProjectRole`.
  partial_update_project_role:
    endpoint: POST /rest/api/3/role/{{ record.path_id }}
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `partialUpdateProjectRole`.
  fully_update_project_role:
    endpoint: PUT /rest/api/3/role/{{ record.path_id }}
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `fullyUpdateProjectRole`.
  delete_project_role_actors_from_role:
    endpoint: DELETE /rest/api/3/role/{{ record.path_id }}/actors
    required fields: path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteProjectRoleActorsFromRole`.
  add_project_role_actors_to_role:
    endpoint: POST /rest/api/3/role/{{ record.path_id }}/actors
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `addProjectRoleActorsToRole`.
  create_screen:
    endpoint: POST /rest/api/3/screens
    required fields: name
    risk: Mutates Jira data/configuration for operation `createScreen`.
  add_field_to_default_screen:
    endpoint: POST /rest/api/3/screens/addToDefault/{{ record.path_fieldId }}
    required fields: path_fieldId
    risk: Mutates Jira data/configuration for operation `addFieldToDefaultScreen`.
  delete_screen:
    endpoint: DELETE /rest/api/3/screens/{{ record.path_screenId }}
    required fields: path_screenId
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteScreen`.
  update_screen:
    endpoint: PUT /rest/api/3/screens/{{ record.path_screenId }}
    required fields: path_screenId
    risk: Mutates Jira data/configuration for operation `updateScreen`.
  add_screen_tab:
    endpoint: POST /rest/api/3/screens/{{ record.path_screenId }}/tabs
    required fields: path_screenId, name
    risk: Mutates Jira data/configuration for operation `addScreenTab`.
  delete_screen_tab:
    endpoint: DELETE /rest/api/3/screens/{{ record.path_screenId }}/tabs/{{ record.path_tabId }}
    required fields: path_screenId, path_tabId
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteScreenTab`.
  rename_screen_tab:
    endpoint: PUT /rest/api/3/screens/{{ record.path_screenId }}/tabs/{{ record.path_tabId }}
    required fields: path_screenId, path_tabId, name
    risk: Mutates Jira data/configuration for operation `renameScreenTab`.
  add_screen_tab_field:
    endpoint: POST /rest/api/3/screens/{{ record.path_screenId }}/tabs/{{ record.path_tabId }}/fields
    required fields: path_screenId, path_tabId, fieldId
    risk: Mutates Jira data/configuration for operation `addScreenTabField`.
  remove_screen_tab_field:
    endpoint: DELETE /rest/api/3/screens/{{ record.path_screenId }}/tabs/{{ record.path_tabId }}/fields/{{ record.path_id }}
    required fields: path_screenId, path_tabId, path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `removeScreenTabField`.
  move_screen_tab_field:
    endpoint: POST /rest/api/3/screens/{{ record.path_screenId }}/tabs/{{ record.path_tabId }}/fields/{{ record.path_id }}/move
    required fields: path_screenId, path_tabId, path_id
    risk: Mutates Jira data/configuration for operation `moveScreenTabField`.
  move_screen_tab:
    endpoint: POST /rest/api/3/screens/{{ record.path_screenId }}/tabs/{{ record.path_tabId }}/move/{{ record.path_pos }}
    required fields: path_screenId, path_tabId, path_pos
    risk: Mutates Jira data/configuration for operation `moveScreenTab`.
  create_screen_scheme:
    endpoint: POST /rest/api/3/screenscheme
    required fields: name, screens
    risk: Mutates Jira data/configuration for operation `createScreenScheme`.
  delete_screen_scheme:
    endpoint: DELETE /rest/api/3/screenscheme/{{ record.path_screenSchemeId }}
    required fields: path_screenSchemeId
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteScreenScheme`.
  update_screen_scheme:
    endpoint: PUT /rest/api/3/screenscheme/{{ record.path_screenSchemeId }}
    required fields: path_screenSchemeId
    risk: Mutates Jira data/configuration for operation `updateScreenScheme`.
  delete_statuses_by_id:
    endpoint: DELETE /rest/api/3/statuses?id={{ record.query_id }}
    required fields: query_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteStatusesById`.
  create_statuses:
    endpoint: POST /rest/api/3/statuses
    required fields: scope, statuses
    risk: Mutates Jira data/configuration for operation `createStatuses`.
  update_statuses:
    endpoint: PUT /rest/api/3/statuses
    required fields: statuses
    risk: Mutates Jira data/configuration for operation `updateStatuses`.
  cancel_task:
    endpoint: POST /rest/api/3/task/{{ record.path_taskId }}/cancel
    required fields: path_taskId
    risk: Mutates Jira data/configuration with destructive semantics for operation `cancelTask`.
  create_ui_modification:
    endpoint: POST /rest/api/3/uiModifications
    required fields: name
    risk: Mutates Jira data/configuration for operation `createUiModification`.
  delete_ui_modification:
    endpoint: DELETE /rest/api/3/uiModifications/{{ record.path_uiModificationId }}
    required fields: path_uiModificationId
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteUiModification`.
  update_ui_modification:
    endpoint: PUT /rest/api/3/uiModifications/{{ record.path_uiModificationId }}
    required fields: path_uiModificationId
    risk: Mutates Jira data/configuration for operation `updateUiModification`.
  delete_avatar:
    endpoint: DELETE /rest/api/3/universal_avatar/type/{{ record.path_type }}/owner/{{ record.path_owningObjectId }}/avatar/{{ record.path_id }}
    required fields: path_type, path_owningObjectId, path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteAvatar`.
  remove_user:
    endpoint: DELETE /rest/api/3/user?accountId={{ record.query_accountId }}
    required fields: query_accountId
    risk: Permanently deletes or removes Jira data/configuration for operation `removeUser`.
  create_user:
    endpoint: POST /rest/api/3/user
    required fields: emailAddress, products
    risk: Mutates Jira data/configuration for operation `createUser`.
  reset_user_columns:
    endpoint: DELETE /rest/api/3/user/columns
    risk: Permanently deletes or removes Jira data/configuration for operation `resetUserColumns`.
  delete_user_property:
    endpoint: DELETE /rest/api/3/user/properties/{{ record.path_propertyKey }}
    required fields: path_propertyKey
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteUserProperty`.
  create_version:
    endpoint: POST /rest/api/3/version
    risk: Mutates Jira data/configuration for operation `createVersion`.
  update_version:
    endpoint: PUT /rest/api/3/version/{{ record.path_id }}
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `updateVersion`.
  merge_versions:
    endpoint: PUT /rest/api/3/version/{{ record.path_id }}/mergeto/{{ record.path_moveIssuesTo }}
    required fields: path_id, path_moveIssuesTo
    risk: Mutates Jira data/configuration for operation `mergeVersions`.
  move_version:
    endpoint: POST /rest/api/3/version/{{ record.path_id }}/move
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `moveVersion`.
  create_related_work:
    endpoint: POST /rest/api/3/version/{{ record.path_id }}/relatedwork
    required fields: path_id, category
    risk: Mutates Jira data/configuration for operation `createRelatedWork`.
  update_related_work:
    endpoint: PUT /rest/api/3/version/{{ record.path_id }}/relatedwork
    required fields: path_id, category
    risk: Mutates Jira data/configuration for operation `updateRelatedWork`.
  delete_and_replace_version:
    endpoint: POST /rest/api/3/version/{{ record.path_id }}/removeAndSwap
    required fields: path_id
    risk: Mutates Jira data/configuration with destructive semantics for operation `deleteAndReplaceVersion`.
  delete_related_work:
    endpoint: DELETE /rest/api/3/version/{{ record.path_versionId }}/relatedwork/{{ record.path_relatedWorkId }}
    required fields: path_versionId, path_relatedWorkId
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteRelatedWork`.
  delete_webhook_by_id:
    endpoint: DELETE /rest/api/3/webhook
    required fields: webhookIds
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteWebhookById`.
  register_dynamic_webhooks:
    endpoint: POST /rest/api/3/webhook
    required fields: url, webhooks
    risk: Mutates Jira data/configuration for operation `registerDynamicWebhooks`.
  refresh_webhooks:
    endpoint: PUT /rest/api/3/webhook/refresh
    required fields: webhookIds
    risk: Mutates Jira data/configuration for operation `refreshWebhooks`.
  update_workflow_transition_rule_configurations:
    endpoint: PUT /rest/api/3/workflow/rule/config
    required fields: workflows
    risk: Mutates Jira data/configuration for operation `updateWorkflowTransitionRuleConfigurations`.
  delete_workflow_transition_rule_configurations:
    endpoint: PUT /rest/api/3/workflow/rule/config/delete
    required fields: workflows
    risk: Mutates Jira data/configuration with destructive semantics for operation `deleteWorkflowTransitionRuleConfigurations`.
  delete_inactive_workflow:
    endpoint: DELETE /rest/api/3/workflow/{{ record.path_entityId }}
    required fields: path_entityId
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteInactiveWorkflow`.
  create_workflows:
    endpoint: POST /rest/api/3/workflows/create
    risk: Mutates Jira data/configuration for operation `createWorkflows`.
  update_workflows:
    endpoint: POST /rest/api/3/workflows/update
    risk: Mutates Jira data/configuration for operation `updateWorkflows`.
  create_workflow_scheme:
    endpoint: POST /rest/api/3/workflowscheme
    risk: Mutates Jira data/configuration for operation `createWorkflowScheme`.
  assign_scheme_to_project:
    endpoint: PUT /rest/api/3/workflowscheme/project
    required fields: projectId
    risk: Mutates Jira data/configuration for operation `assignSchemeToProject`.
  switch_workflow_scheme_for_project:
    endpoint: POST /rest/api/3/workflowscheme/project/switch
    risk: Mutates Jira data/configuration for operation `switchWorkflowSchemeForProject`.
  update_schemes:
    endpoint: POST /rest/api/3/workflowscheme/update
    required fields: description, id, name, version
    risk: Mutates Jira data/configuration for operation `updateSchemes`.
  delete_workflow_scheme:
    endpoint: DELETE /rest/api/3/workflowscheme/{{ record.path_id }}
    required fields: path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteWorkflowScheme`.
  update_workflow_scheme:
    endpoint: PUT /rest/api/3/workflowscheme/{{ record.path_id }}
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `updateWorkflowScheme`.
  create_workflow_scheme_draft_from_parent:
    endpoint: POST /rest/api/3/workflowscheme/{{ record.path_id }}/createdraft
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `createWorkflowSchemeDraftFromParent`.
  delete_default_workflow:
    endpoint: DELETE /rest/api/3/workflowscheme/{{ record.path_id }}/default
    required fields: path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteDefaultWorkflow`.
  update_default_workflow:
    endpoint: PUT /rest/api/3/workflowscheme/{{ record.path_id }}/default
    required fields: path_id, workflow
    risk: Mutates Jira data/configuration for operation `updateDefaultWorkflow`.
  delete_workflow_scheme_draft:
    endpoint: DELETE /rest/api/3/workflowscheme/{{ record.path_id }}/draft
    required fields: path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteWorkflowSchemeDraft`.
  update_workflow_scheme_draft:
    endpoint: PUT /rest/api/3/workflowscheme/{{ record.path_id }}/draft
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `updateWorkflowSchemeDraft`.
  delete_draft_default_workflow:
    endpoint: DELETE /rest/api/3/workflowscheme/{{ record.path_id }}/draft/default
    required fields: path_id
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteDraftDefaultWorkflow`.
  update_draft_default_workflow:
    endpoint: PUT /rest/api/3/workflowscheme/{{ record.path_id }}/draft/default
    required fields: path_id, workflow
    risk: Mutates Jira data/configuration for operation `updateDraftDefaultWorkflow`.
  delete_workflow_scheme_draft_issue_type:
    endpoint: DELETE /rest/api/3/workflowscheme/{{ record.path_id }}/draft/issuetype/{{ record.path_issueType }}
    required fields: path_id, path_issueType
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteWorkflowSchemeDraftIssueType`.
  set_workflow_scheme_draft_issue_type:
    endpoint: PUT /rest/api/3/workflowscheme/{{ record.path_id }}/draft/issuetype/{{ record.path_issueType }}
    required fields: path_id, path_issueType
    risk: Mutates Jira data/configuration for operation `setWorkflowSchemeDraftIssueType`.
  publish_draft_workflow_scheme:
    endpoint: POST /rest/api/3/workflowscheme/{{ record.path_id }}/draft/publish
    required fields: path_id
    risk: Mutates Jira data/configuration for operation `publishDraftWorkflowScheme`.
  delete_draft_workflow_mapping:
    endpoint: DELETE /rest/api/3/workflowscheme/{{ record.path_id }}/draft/workflow?workflowName={{ record.query_workflowName }}
    required fields: path_id, query_workflowName
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteDraftWorkflowMapping`.
  update_draft_workflow_mapping:
    endpoint: PUT /rest/api/3/workflowscheme/{{ record.path_id }}/draft/workflow?workflowName={{ record.query_workflowName }}
    required fields: path_id, query_workflowName
    risk: Mutates Jira data/configuration for operation `updateDraftWorkflowMapping`.
  delete_workflow_scheme_issue_type:
    endpoint: DELETE /rest/api/3/workflowscheme/{{ record.path_id }}/issuetype/{{ record.path_issueType }}
    required fields: path_id, path_issueType
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteWorkflowSchemeIssueType`.
  set_workflow_scheme_issue_type:
    endpoint: PUT /rest/api/3/workflowscheme/{{ record.path_id }}/issuetype/{{ record.path_issueType }}
    required fields: path_id, path_issueType
    risk: Mutates Jira data/configuration for operation `setWorkflowSchemeIssueType`.
  delete_workflow_mapping:
    endpoint: DELETE /rest/api/3/workflowscheme/{{ record.path_id }}/workflow?workflowName={{ record.query_workflowName }}
    required fields: path_id, query_workflowName
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteWorkflowMapping`.
  update_workflow_mapping:
    endpoint: PUT /rest/api/3/workflowscheme/{{ record.path_id }}/workflow?workflowName={{ record.query_workflowName }}
    required fields: path_id, query_workflowName
    risk: Mutates Jira data/configuration for operation `updateWorkflowMapping`.
  addon_properties_resource_delete_addon_property_delete:
    endpoint: DELETE /rest/atlassian-connect/1/addons/{{ record.path_addonKey }}/properties/{{ record.path_propertyKey }}
    required fields: path_addonKey, path_propertyKey
    risk: Permanently deletes or removes Jira data/configuration for operation `AddonPropertiesResource.deleteAddonProperty_delete`.
  dynamic_modules_resource_remove_modules_delete:
    endpoint: DELETE /rest/atlassian-connect/1/app/module/dynamic
    risk: Permanently deletes or removes Jira data/configuration for operation `DynamicModulesResource.removeModules_delete`.
  dynamic_modules_resource_register_modules_post:
    endpoint: POST /rest/atlassian-connect/1/app/module/dynamic
    required fields: modules
    risk: Mutates Jira data/configuration for operation `DynamicModulesResource.registerModules_post`.
  connect_to_forge_migration_task_submission_resource_submit_task_post:
    endpoint: POST /rest/atlassian-connect/1/migration/{{ record.path_connectKey }}/{{ record.path_jiraIssueFieldsKey }}/task
    required fields: path_connectKey, path_jiraIssueFieldsKey
    risk: Mutates Jira data/configuration for operation `ConnectToForgeMigrationTaskSubmissionResource.submitTask_post`.
  delete_forge_app_property:
    endpoint: DELETE /rest/forge/1/app/properties/{{ record.path_propertyKey }}
    required fields: path_propertyKey
    risk: Permanently deletes or removes Jira data/configuration for operation `deleteForgeAppProperty`.

SECURITY
  read risk: external Jira Cloud API read of issues, projects, users, and bounded JSON direct-read endpoints
  write risk: external Jira Cloud API mutation through named reverse ETL actions only; DELETE/destructive actions require typed confirmation
  approval: reverse ETL plan, preview, explicit approval, execute; destructive actions require --confirm destructive
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Work with Jira Cloud issues, projects, users, direct reads, and approval-gated write actions.
  Usage: pm jira <command> --credential <name> [flags] --json
  Source CLI: Atlassian Jira Cloud REST API v3 (https://developer.atlassian.com/cloud/jira/platform/swagger-v3.v3.json)
  Global flags:
    --credential (string): Polymetrics credential name for Jira.
    --max-bytes (integer): Maximum direct-read response bytes; capped by the connector command runner.
    --plan (boolean): Create a reverse ETL plan for write commands instead of executing.
    --preview (boolean): Include write request preview while creating a plan.
    --approve (string): Approval token for an existing reverse ETL plan.
    --confirm (string): Typed confirmation challenge for destructive plans.
  Issue
  Project
  User
  Announcement Banner
  Issue Custom Field Configuration Apps
  Issue Custom Field Values Apps
  Jira Settings
  Application Roles
  Issue Attachments
  Audit Records
  Avatars
  Issue Bulk Operations
  Issues
  Classification Levels
  Issue Comments
  Issue Comment Properties
  Project Components
  Field Schemes
  Time Tracking
  Issue Custom Field Options
  Dashboards
  App Data Policies
  Jira Expressions
  Issue Fields
  Issue Custom Field Associations
  Issue Custom Field Contexts
  Screens
  Issue Custom Field Options Apps
  Filters
  Filter Sharing
  Issue Panels
  Groups
  Group And User Picker
  License Metrics
  Issue Search
  Issue Properties
  Issue Watchers
  Issue Remote Links
  Issue Votes
  Issue Worklogs
  Issue Worklog Properties
  Issue Links
  Issue Link Types
  Issue Security Schemes
  Issue Security Level
  Issue Types
  Issue Type Properties
  Issue Type Schemes
  Issue Type Screen Schemes
  Jql
  Jql Functions Apps
  Labels
  Permissions
  Myself
  Issue Notification Schemes
  Permission Schemes
  Plans
  Teams In Plan
  Issue Priorities
  Priority Schemes
  Projects
  Project Templates
  Project Types
  Project Avatars
  Project Classification Levels
  Project Features
  Project Properties
  Project Roles
  Project Role Actors
  Project Versions
  Project Email
  Project Permission Schemes
  Project Categories
  Project Key And Name Validation
  Issue Redaction
  Issue Resolutions
  Screen Tabs
  Screen Tab Fields
  Screen Schemes
  Server Info
  Issue Navigator Settings
  Workflow Statuses
  Workflow Status Categories
  Status
  Tasks
  Ui Modifications Apps
  Users
  User Search
  User Properties
  Webhooks
  Workflows
  Workflow Transition Rules
  Workflow Schemes
  Workflow Scheme Project Associations
  Workflow Scheme Drafts
  App Properties
  Dynamic Modules
  App Migration
  Migration Of Connect Modules To Forge
  Service Registry
  Api
  Other Commands
    issue search - Read Jira issues. [intent=etl availability=implemented stream=issues]
    project search - Read Jira projects. [intent=etl availability=implemented stream=projects]
    user search - Read Jira users. [intent=etl availability=implemented stream=users]
    announcement-banner get-banner - Get announcement banner configuration [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    announcement-banner set-banner - Update announcement banner configuration [intent=reverse_etl availability=partial write=set_banner]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `setBanner`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.
    issue-custom-field-configuration-apps get-custom-fields-configurations - Bulk get custom field configurations [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; flags: --field-ids-or-keys, --id, --field-context-id, --issue-id, --project-key-or-id, --issue-type-id, --start-at, --max-results
    issue-custom-field-values-apps update-multiple-custom-field-values - Update custom fields [intent=reverse_etl availability=partial write=update_multiple_custom_field_values]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateMultipleCustomFieldValues`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    issue-custom-field-configuration-apps get-custom-field-configuration - Get custom field configurations [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --field-id-or-key, --id, --field-context-id, --issue-id, --project-key-or-id, --issue-type-id, --start-at, --max-results
    issue-custom-field-configuration-apps update-custom-field-configuration - Update custom field configurations [intent=reverse_etl availability=partial write=update_custom_field_configuration]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateCustomFieldConfiguration`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --field-id-or-key, --configurations
    issue-custom-field-values-apps update-custom-field-value - Update custom field value [intent=reverse_etl availability=partial write=update_custom_field_value]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateCustomFieldValue`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --field-id-or-key
    jira-settings get-application-property - Get application property [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --key, --permission-level, --key-filter
    jira-settings get-advanced-settings - Get advanced settings [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    jira-settings set-application-property - Set application property [intent=reverse_etl availability=partial write=set_application_property]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `setApplicationProperty`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --id
    application-roles get-all-application-roles - Get all application roles [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    application-roles get-application-role - Get application role [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --key
    issue-attachments get-attachment-content - Get attachment content [intent=direct_read availability=planned]; approval: Not executable in this slice.; risk: Binary response; blocked until a bounded file output policy exists.; notes: bounded binary download executor is not implemented in the shared command runner; row remains blocked until a connector-generic binary/file output policy exists
    issue-attachments get-attachment-meta - Get Jira attachment settings [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    issue-attachments get-attachment-thumbnail - Get attachment thumbnail [intent=direct_read availability=planned]; approval: Not executable in this slice.; risk: Binary response; blocked until a bounded file output policy exists.; notes: bounded binary download executor is not implemented in the shared command runner; row remains blocked until a connector-generic binary/file output policy exists
    issue-attachments remove-attachment - Delete attachment [intent=reverse_etl availability=implemented write=remove_attachment]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `removeAttachment`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    issue-attachments get-attachment - Get attachment metadata [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    issue-attachments expand-attachment-for-humans - Get all metadata for an expanded attachment [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    issue-attachments expand-attachment-for-machines - Get contents metadata for an expanded attachment [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    audit-records get-audit-records - Get audit records [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --offset, --limit, --filter, --from, --to
    avatars get-all-system-avatars - Get system avatars by type [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --type
    issue-bulk-operations submit-bulk-delete - Bulk delete issues [intent=reverse_etl availability=implemented write=submit_bulk_delete]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration with destructive semantics for operation `submitBulkDelete`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --selected-issue-ids-or-keys
    issue-bulk-operations get-bulk-editable-fields - Get bulk editable fields [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-ids-or-keys, --search-text, --ending-before, --starting-after
    issue-bulk-operations submit-bulk-edit - Bulk edit issues [intent=reverse_etl availability=partial write=submit_bulk_edit]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `submitBulkEdit`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits structured JSON flag support for required object fields.; flags: --selected-actions, --selected-issue-ids-or-keys
    issue-bulk-operations submit-bulk-move - Bulk move issues [intent=reverse_etl availability=partial write=submit_bulk_move]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `submitBulkMove`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    issue-bulk-operations get-available-transitions - Get available transitions [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-ids-or-keys, --ending-before, --starting-after
    issue-bulk-operations submit-bulk-transition - Bulk transition issue statuses [intent=reverse_etl availability=partial write=submit_bulk_transition]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `submitBulkTransition`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --bulk-transition-inputs
    issue-bulk-operations submit-bulk-unwatch - Bulk unwatch issues [intent=reverse_etl availability=implemented write=submit_bulk_unwatch]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `submitBulkUnwatch`.; flags: --selected-issue-ids-or-keys
    issue-bulk-operations submit-bulk-watch - Bulk watch issues [intent=reverse_etl availability=implemented write=submit_bulk_watch]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `submitBulkWatch`.; flags: --selected-issue-ids-or-keys
    issue-bulk-operations get-bulk-operation-progress - Get bulk issue operation progress [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --task-id
    issues get-bulk-changelogs - Bulk fetch changelogs [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; flags: --field-ids, --issue-ids-or-keys, --max-results, --next-page-token
    classification-levels get-all-user-data-classification-levels - Get all classification levels [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --status, --order-by
    issue-comments get-comments-by-ids - Get comments by IDs [intent=direct_read availability=partial]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; notes: Provider-style command awaits typed integer/object array body flag support before it can run safely.; flags: --ids, --expand
    issue-comment-properties get-comment-property-keys - Get comment property keys [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --comment-id
    issue-comment-properties delete-comment-property - Delete comment property [intent=reverse_etl availability=implemented write=delete_comment_property]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteCommentProperty`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --comment-id, --property-key
    issue-comment-properties get-comment-property - Get comment property [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --comment-id, --property-key
    issue-comment-properties set-comment-property - Set comment property [intent=reverse_etl availability=planned]; notes: Provider-style command is planned until raw or dynamic JSON request-body support can represent this required body safely.
    project-components find-components-for-projects - Find components for projects [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-ids-or-keys, --start-at, --max-results, --order-by, --query
    project-components create-component - Create component [intent=reverse_etl availability=partial write=create_component]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createComponent`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    project-components delete-component - Delete component [intent=reverse_etl availability=implemented write=delete_component]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteComponent`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    project-components get-component - Get component [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    project-components update-component - Update component [intent=reverse_etl availability=partial write=update_component]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateComponent`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --id
    project-components get-component-related-issues - Get component issues count [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    field-schemes get-field-association-schemes - Get field schemes [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id, --query, --start-at, --max-results
    field-schemes create-field-association-scheme - Create field scheme [intent=reverse_etl availability=implemented write=create_field_association_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createFieldAssociationScheme`.; flags: --name
    field-schemes remove-fields-associated-with-schemes - Remove fields associated with field schemes [intent=reverse_etl availability=planned]; notes: Provider-style command is planned until raw or dynamic JSON request-body support can represent this required body safely.
    field-schemes update-fields-associated-with-schemes - Update fields associated with field schemes [intent=reverse_etl availability=planned]; notes: Provider-style command is planned until raw or dynamic JSON request-body support can represent this required body safely.
    field-schemes remove-field-association-scheme-item-parameters - Remove field parameters [intent=reverse_etl availability=planned]; notes: Provider-style command is planned until raw or dynamic JSON request-body support can represent this required body safely.
    field-schemes update-field-association-scheme-item-parameters - Update field parameters [intent=reverse_etl availability=planned]; notes: Provider-style command is planned until raw or dynamic JSON request-body support can represent this required body safely.
    field-schemes get-projects-with-field-schemes - Get projects with field schemes [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --project-id
    field-schemes associate-projects-to-field-association-schemes - Associate projects to field schemes [intent=reverse_etl availability=planned]; notes: Provider-style command is planned until raw or dynamic JSON request-body support can represent this required body safely.
    field-schemes delete-field-association-scheme - Delete a field scheme [intent=reverse_etl availability=implemented write=delete_field_association_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteFieldAssociationScheme`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    field-schemes get-field-association-scheme-by-id - Get field scheme [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    field-schemes update-field-association-scheme - Update field scheme [intent=reverse_etl availability=partial write=update_field_association_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateFieldAssociationScheme`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --id
    field-schemes clone-field-association-scheme - Clone field scheme [intent=reverse_etl availability=implemented write=clone_field_association_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `cloneFieldAssociationScheme`.; flags: --id, --name
    field-schemes search-field-association-scheme-fields - Search field scheme fields [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --field-id, --id
    field-schemes get-field-association-scheme-item-parameters - Get field parameters [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id, --field-id
    field-schemes search-field-association-scheme-projects - Search field scheme projects [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --project-id, --id
    jira-settings get-configuration - Get global settings [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    time-tracking get-selected-time-tracking-implementation - Get selected time tracking provider [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    time-tracking select-time-tracking-implementation - Select time tracking provider [intent=reverse_etl availability=implemented write=select_time_tracking_implementation]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `selectTimeTrackingImplementation`.; flags: --key
    time-tracking get-available-time-tracking-implementations - Get all time tracking providers [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    time-tracking get-shared-time-tracking-configuration - Get time tracking settings [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    time-tracking set-shared-time-tracking-configuration - Set time tracking settings [intent=reverse_etl availability=partial write=set_shared_time_tracking_configuration]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `setSharedTimeTrackingConfiguration`.; notes: Provider-style command awaits numeric/structured flag support; use record-driven reverse ETL plans for this action.; flags: --default-unit, --time-format, --working-days-per-week, --working-hours-per-day
    issue-custom-field-options get-custom-field-option - Get custom field option [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    dashboards get-all-dashboards - Get all dashboards [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --filter, --start-at, --max-results
    dashboards create-dashboard - Create dashboard [intent=reverse_etl availability=partial write=create_dashboard]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createDashboard`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --edit-permissions, --name, --share-permissions
    dashboards bulk-edit-dashboards - Bulk edit dashboards [intent=reverse_etl availability=partial write=bulk_edit_dashboards]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `bulkEditDashboards`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --action, --entity-ids
    dashboards get-all-available-dashboard-gadgets - Get available gadgets [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    dashboards get-dashboards-paginated - Search for dashboards [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --dashboard-name, --account-id, --owner, --groupname, --group-id, --project-id, --order-by, --start-at, --max-results, --status, --expand
    dashboards get-all-gadgets - Get gadgets [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --dashboard-id, --module-key, --uri, --gadget-id
    dashboards add-gadget - Add gadget to dashboard [intent=reverse_etl availability=partial write=add_gadget]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `addGadget`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --dashboard-id
    dashboards remove-gadget - Remove gadget from dashboard [intent=reverse_etl availability=implemented write=remove_gadget]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `removeGadget`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --dashboard-id, --gadget-id
    dashboards update-gadget - Update gadget on dashboard [intent=reverse_etl availability=partial write=update_gadget]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateGadget`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --dashboard-id, --gadget-id
    dashboards get-dashboard-item-property-keys - Get dashboard item property keys [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --dashboard-id, --item-id
    dashboards delete-dashboard-item-property - Delete dashboard item property [intent=reverse_etl availability=implemented write=delete_dashboard_item_property]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteDashboardItemProperty`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --dashboard-id, --item-id, --property-key
    dashboards get-dashboard-item-property - Get dashboard item property [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --dashboard-id, --item-id, --property-key
    dashboards set-dashboard-item-property - Set dashboard item property [intent=reverse_etl availability=planned]; notes: Provider-style command is planned until raw or dynamic JSON request-body support can represent this required body safely.
    dashboards delete-dashboard - Delete dashboard [intent=reverse_etl availability=implemented write=delete_dashboard]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteDashboard`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    dashboards get-dashboard - Get dashboard [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    dashboards update-dashboard - Update dashboard [intent=reverse_etl availability=partial write=update_dashboard]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateDashboard`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --id, --edit-permissions, --name, --share-permissions
    dashboards copy-dashboard - Copy dashboard [intent=reverse_etl availability=partial write=copy_dashboard]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `copyDashboard`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --id, --edit-permissions, --name, --share-permissions
    app-data-policies get-policy - Get data policy for the workspace [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    app-data-policies get-policies - Get data policy for projects [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --ids
    issues get-events - Get events [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    jira-expressions analyse-expression - Analyse Jira expression [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; flags: --expressions, --check
    jira-expressions evaluate-jsis-jira-expression - Evaluate Jira expression using enhanced search API [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; flags: --expression, --expand
    issue-fields get-fields - Get fields [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    issue-fields create-custom-field - Create custom field [intent=reverse_etl availability=implemented write=create_custom_field]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createCustomField`.; flags: --name, --type
    issue-custom-field-associations remove-associations - Remove associations [intent=reverse_etl availability=partial write=remove_associations]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `removeAssociations`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    issue-custom-field-associations create-associations - Create associations [intent=reverse_etl availability=partial write=create_associations]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createAssociations`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --association-contexts, --fields
    issue-fields get-fields-paginated - Get fields paginated [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --type, --id, --query, --order-by, --expand, --project-ids
    issue-fields get-trashed-fields-paginated - Get fields in trash paginated [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --id, --query, --expand, --order-by
    issue-fields update-custom-field - Update custom field [intent=reverse_etl availability=partial write=update_custom_field]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateCustomField`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --field-id
    issue-fields get-field-project-associations - Get field project associations [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --field-id, --start-at, --max-results
    issue-custom-field-contexts get-contexts-for-field - Get custom field contexts [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --field-id, --is-any-issue-type, --is-global-context, --context-id, --start-at, --max-results
    issue-custom-field-contexts create-custom-field-context - Create custom field context [intent=reverse_etl availability=implemented write=create_custom_field_context]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createCustomFieldContext`.; flags: --field-id, --name
    issue-custom-field-contexts get-context-default-values - Get default values for a custom field grouped by context and issue type [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --field-id, --context-id, --issue-type-id, --start-at, --max-results
    issue-custom-field-contexts get-issue-type-mappings-for-contexts - Get issue types for custom field context [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --field-id, --context-id, --start-at, --max-results
    issue-custom-field-contexts get-custom-field-contexts-for-projects-and-issue-types - Get custom field contexts for projects and issue types [intent=direct_read availability=partial]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; notes: Provider-style command awaits typed integer/object array body flag support before it can run safely.; flags: --mappings, --field-id, --start-at, --max-results
    issue-custom-field-contexts get-project-context-mapping - Get project mappings for custom field context [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --field-id, --context-id, --start-at, --max-results
    issue-custom-field-contexts delete-custom-field-context - Delete custom field context [intent=reverse_etl availability=implemented write=delete_custom_field_context]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteCustomFieldContext`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --field-id, --context-id
    issue-custom-field-contexts update-custom-field-context - Update custom field context [intent=reverse_etl availability=partial write=update_custom_field_context]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateCustomFieldContext`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --field-id, --context-id
    issue-custom-field-contexts add-issue-types-to-context - Add issue types to context [intent=reverse_etl availability=implemented write=add_issue_types_to_context]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `addIssueTypesToContext`.; flags: --field-id, --context-id, --issue-type-ids
    issue-custom-field-contexts remove-issue-types-from-context - Remove issue types from context [intent=reverse_etl availability=implemented write=remove_issue_types_from_context]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration with destructive semantics for operation `removeIssueTypesFromContext`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --field-id, --context-id, --issue-type-ids
    issue-custom-field-options get-options-for-context - Get custom field options (context) [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --field-id, --context-id, --option-id, --only-options, --start-at, --max-results
    issue-custom-field-options create-custom-field-option - Create custom field options (context) [intent=reverse_etl availability=partial write=create_custom_field_option]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createCustomFieldOption`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --field-id, --context-id
    issue-custom-field-options update-custom-field-option - Update custom field options (context) [intent=reverse_etl availability=partial write=update_custom_field_option]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateCustomFieldOption`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --field-id, --context-id
    issue-custom-field-options reorder-custom-field-options - Reorder custom field options (context) [intent=reverse_etl availability=implemented write=reorder_custom_field_options]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `reorderCustomFieldOptions`.; flags: --field-id, --context-id, --custom-field-option-ids
    issue-custom-field-options delete-custom-field-option - Delete custom field options (context) [intent=reverse_etl availability=implemented write=delete_custom_field_option]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteCustomFieldOption`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --field-id, --context-id, --option-id
    issue-custom-field-options replace-custom-field-option - Replace custom field options [intent=reverse_etl availability=implemented write=replace_custom_field_option]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `replaceCustomFieldOption`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --field-id, --option-id, --context-id
    issue-custom-field-contexts assign-projects-to-custom-field-context - Assign custom field context to projects [intent=reverse_etl availability=implemented write=assign_projects_to_custom_field_context]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `assignProjectsToCustomFieldContext`.; flags: --field-id, --context-id, --project-ids
    issue-custom-field-contexts remove-custom-field-context-from-projects - Remove custom field context from projects [intent=reverse_etl availability=implemented write=remove_custom_field_context_from_projects]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration with destructive semantics for operation `removeCustomFieldContextFromProjects`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --field-id, --context-id, --project-ids
    screens get-screens-for-field - Get screens for a field [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --field-id, --start-at, --max-results, --expand
    issue-custom-field-options-apps get-all-issue-field-options - Get all issue field options [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --field-key
    issue-custom-field-options-apps create-issue-field-option - Create issue field option [intent=reverse_etl availability=implemented write=create_issue_field_option]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createIssueFieldOption`.; flags: --field-key, --value
    issue-custom-field-options-apps get-selectable-issue-field-options - Get selectable issue field options [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --project-id, --field-key
    issue-custom-field-options-apps get-visible-issue-field-options - Get visible issue field options [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --project-id, --field-key
    issue-custom-field-options-apps delete-issue-field-option - Delete issue field option [intent=reverse_etl availability=implemented write=delete_issue_field_option]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteIssueFieldOption`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --field-key, --option-id
    issue-custom-field-options-apps get-issue-field-option - Get issue field option [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --field-key, --option-id
    issue-custom-field-options-apps update-issue-field-option - Update issue field option [intent=reverse_etl availability=implemented write=update_issue_field_option]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateIssueFieldOption`.; flags: --field-key, --option-id, --id, --value
    issue-custom-field-options-apps replace-issue-field-option - Replace issue field option [intent=reverse_etl availability=implemented write=replace_issue_field_option]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `replaceIssueFieldOption`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --field-key, --option-id
    issue-fields delete-custom-field - Delete custom field [intent=reverse_etl availability=implemented write=delete_custom_field]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteCustomField`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    issue-fields restore-custom-field - Restore custom field from trash [intent=reverse_etl availability=implemented write=restore_custom_field]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration with destructive semantics for operation `restoreCustomField`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    issue-fields trash-custom-field - Move custom field to trash [intent=reverse_etl availability=implemented write=trash_custom_field]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration with destructive semantics for operation `trashCustomField`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    filters create-filter - Create filter [intent=reverse_etl availability=implemented write=create_filter]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createFilter`.; flags: --name
    filter-sharing get-default-share-scope - Get default share scope [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    filter-sharing set-default-share-scope - Set default share scope [intent=reverse_etl availability=implemented write=set_default_share_scope]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `setDefaultShareScope`.; flags: --scope
    filters get-favourite-filters - Get favorite filters [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --expand
    filters get-my-filters - Get my filters [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --expand, --include-favourites
    filters get-filters-paginated - Search for filters [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --filter-name, --account-id, --owner, --groupname, --group-id, --project-id, --id, --order-by, --start-at, --max-results, --expand, --override-share-permissions, --is-substring-match
    filters delete-filter - Delete filter [intent=reverse_etl availability=implemented write=delete_filter]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteFilter`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    filters get-filter - Get filter [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id, --expand, --override-share-permissions
    filters update-filter - Update filter [intent=reverse_etl availability=implemented write=update_filter]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateFilter`.; flags: --id, --name
    filters reset-columns - Reset columns [intent=reverse_etl availability=implemented write=reset_columns]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `resetColumns`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    filters get-columns - Get columns [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    filters set-columns - Set columns [intent=reverse_etl availability=partial write=set_columns]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `setColumns`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --id
    filters delete-favourite-for-filter - Remove filter as favorite [intent=reverse_etl availability=implemented write=delete_favourite_for_filter]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteFavouriteForFilter`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    filters set-favourite-for-filter - Add filter as favorite [intent=reverse_etl availability=implemented write=set_favourite_for_filter]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `setFavouriteForFilter`.; flags: --id
    filters change-filter-owner - Change filter owner [intent=reverse_etl availability=implemented write=change_filter_owner]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `changeFilterOwner`.; flags: --id, --account-id
    filter-sharing get-share-permissions - Get share permissions [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    filter-sharing add-share-permission - Add share permission [intent=reverse_etl availability=implemented write=add_share_permission]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `addSharePermission`.; flags: --id, --type
    filter-sharing delete-share-permission - Delete share permission [intent=reverse_etl availability=implemented write=delete_share_permission]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteSharePermission`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id, --permission-id
    filter-sharing get-share-permission - Get share permission [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id, --permission-id
    issue-panels bulk-pin-unpin-projects-async - Bulk pin or unpin issue panel to projects [intent=reverse_etl availability=partial write=bulk_pin_unpin_projects_async]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `bulkPinUnpinProjectsAsync`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --module-id, --project-list
    groups remove-group - Remove group [intent=reverse_etl availability=implemented write=remove_group]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `removeGroup`.; notes: Requires --confirm destructive when executing the approved reverse plan.
    groups create-group - Create group [intent=reverse_etl availability=implemented write=create_group]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createGroup`.; flags: --name
    groups bulk-get-groups - Bulk get groups [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --group-id, --group-name, --access-type, --application-key
    groups get-users-from-group - Get users from group [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --groupname, --group-id, --include-inactive-users, --start-at, --max-results
    groups remove-user-from-group - Remove user from group [intent=reverse_etl availability=implemented write=remove_user_from_group]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `removeUserFromGroup`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --account-id
    groups add-user-to-group - Add user to group [intent=reverse_etl availability=partial write=add_user_to_group]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `addUserToGroup`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.
    groups find-groups - Find groups [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --account-id, --query, --exclude, --exclude-id, --max-results, --case-insensitive, --user-name
    group-and-user-picker find-users-and-groups - Find users and groups [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --query, --max-results, --show-avatar, --field-id, --project-id, --issue-type-id, --avatar-size, --case-insensitive, --exclude-connect-addons, --include-ai-agents
    license-metrics get-license - Get license [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    issues create-issue - Create issue [intent=reverse_etl availability=partial write=create_issue]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createIssue`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    issues archive-issues-async - Archive issue(s) by JQL [intent=reverse_etl availability=partial write=archive_issues_async]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration with destructive semantics for operation `archiveIssuesAsync`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.
    issues archive-issues - Archive issue(s) by issue ID/key [intent=reverse_etl availability=partial write=archive_issues]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration with destructive semantics for operation `archiveIssues`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    issues create-issues - Bulk create issue [intent=reverse_etl availability=partial write=create_issues]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createIssues`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    issues bulk-fetch-issues - Bulk fetch issues [intent=reverse_etl availability=implemented write=bulk_fetch_issues]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `bulkFetchIssues`.; flags: --issue-ids-or-keys
    issues get-create-issue-meta-issue-types - Get create metadata issue types for a project [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id-or-key, --start-at, --max-results
    issues get-create-issue-meta-issue-type-id - Get create field metadata for a project and issue type id [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id-or-key, --issue-type-id, --start-at, --max-results
    issues get-issue-limit-report - Get issue limit report [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --is-returning-keys
    issue-search get-issue-picker-resource - Get issue picker suggestions [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --query, --current-jql, --current-issue-key, --current-project-id, --show-sub-tasks, --show-sub-task-parent
    issue-properties bulk-set-issues-properties-list - Bulk set issues properties by list [intent=reverse_etl availability=partial write=bulk_set_issues_properties_list]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `bulkSetIssuesPropertiesList`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    issue-properties bulk-set-issue-properties-by-issue - Bulk set issue properties by issue [intent=reverse_etl availability=partial write=bulk_set_issue_properties_by_issue]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `bulkSetIssuePropertiesByIssue`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    issue-properties bulk-delete-issue-property - Bulk delete issue property [intent=reverse_etl availability=partial write=bulk_delete_issue_property]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `bulkDeleteIssueProperty`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --property-key
    issue-properties bulk-set-issue-property - Bulk set issue property [intent=reverse_etl availability=partial write=bulk_set_issue_property]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `bulkSetIssueProperty`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --property-key
    issues unarchive-issues - Unarchive issue(s) by issue keys/ID [intent=reverse_etl availability=partial write=unarchive_issues]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration with destructive semantics for operation `unarchiveIssues`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    issue-watchers get-is-watching-issue-bulk - Get is watching issue bulk [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; flags: --issue-ids
    issues delete-issue - Delete issue [intent=reverse_etl availability=implemented write=delete_issue]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteIssue`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --issue-id-or-key
    issues get-issue - Get issue [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-id-or-key, --fields, --fields-by-keys, --expand, --properties, --update-history, --fail-fast
    issues edit-issue - Edit issue [intent=reverse_etl availability=partial write=edit_issue]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `editIssue`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --issue-id-or-key
    issues assign-issue - Assign issue [intent=reverse_etl availability=partial write=assign_issue]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `assignIssue`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --issue-id-or-key
    issue-attachments add-attachment - Add attachment [intent=reverse_etl availability=implemented write=add_attachment]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `addAttachment`.; flags: --issue-id-or-key, --file-path
    issues get-change-logs - Get changelogs [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-id-or-key, --start-at, --max-results
    issues get-change-logs-by-ids - Get changelogs by IDs [intent=direct_read availability=partial]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; notes: Provider-style command awaits typed integer/object array body flag support before it can run safely.; flags: --changelog-ids, --issue-id-or-key
    issue-comments get-comments - Get comments [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-id-or-key, --start-at, --max-results, --order-by, --expand
    issue-comments add-comment - Add comment [intent=reverse_etl availability=partial write=add_comment]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `addComment`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --issue-id-or-key
    issue-comments delete-comment - Delete comment [intent=reverse_etl availability=implemented write=delete_comment]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteComment`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --issue-id-or-key, --id
    issue-comments get-comment - Get comment [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-id-or-key, --id, --expand
    issue-comments update-comment - Update comment [intent=reverse_etl availability=partial write=update_comment]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateComment`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --issue-id-or-key, --id
    issues get-edit-issue-meta - Get edit issue metadata [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-id-or-key, --override-screen-security, --override-editable-flag
    issues notify - Send notification for issue [intent=reverse_etl availability=partial write=notify]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `notify`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --issue-id-or-key
    issue-properties get-issue-property-keys - Get issue property keys [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-id-or-key
    issue-properties delete-issue-property - Delete issue property [intent=reverse_etl availability=implemented write=delete_issue_property]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteIssueProperty`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --issue-id-or-key, --property-key
    issue-properties get-issue-property - Get issue property [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-id-or-key, --property-key
    issue-properties set-issue-property - Set issue property [intent=reverse_etl availability=planned]; notes: Provider-style command is planned until raw or dynamic JSON request-body support can represent this required body safely.
    issue-remote-links delete-remote-issue-link-by-global-id - Delete remote issue link by global ID [intent=reverse_etl availability=implemented write=delete_remote_issue_link_by_global_id]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteRemoteIssueLinkByGlobalId`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --issue-id-or-key, --global-id
    issue-remote-links get-remote-issue-links - Get remote issue links [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-id-or-key, --global-id
    issue-remote-links create-or-update-remote-issue-link - Create or update remote issue link [intent=reverse_etl availability=partial write=create_or_update_remote_issue_link]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createOrUpdateRemoteIssueLink`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits structured JSON flag support for required object fields.; flags: --issue-id-or-key
    issue-remote-links delete-remote-issue-link-by-id - Delete remote issue link by ID [intent=reverse_etl availability=implemented write=delete_remote_issue_link_by_id]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteRemoteIssueLinkById`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --issue-id-or-key, --link-id
    issue-remote-links get-remote-issue-link-by-id - Get remote issue link by ID [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-id-or-key, --link-id
    issue-remote-links update-remote-issue-link - Update remote issue link by ID [intent=reverse_etl availability=partial write=update_remote_issue_link]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateRemoteIssueLink`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits structured JSON flag support for required object fields.; flags: --issue-id-or-key, --link-id
    issues get-transitions - Get transitions [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-id-or-key, --expand, --transition-id, --skip-remote-only-condition, --include-unavailable-transitions, --sort-by-ops-bar-and-status
    issues do-transition - Transition issue [intent=reverse_etl availability=partial write=do_transition]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `doTransition`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --issue-id-or-key
    issue-votes remove-vote - Delete vote [intent=reverse_etl availability=implemented write=remove_vote]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `removeVote`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --issue-id-or-key
    issue-votes get-votes - Get votes [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-id-or-key
    issue-votes add-vote - Add vote [intent=reverse_etl availability=implemented write=add_vote]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `addVote`.; flags: --issue-id-or-key
    issue-watchers remove-watcher - Delete watcher [intent=reverse_etl availability=implemented write=remove_watcher]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `removeWatcher`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --issue-id-or-key
    issue-watchers get-issue-watchers - Get issue watchers [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-id-or-key
    issue-worklogs bulk-delete-worklogs - Bulk delete worklogs [intent=reverse_etl availability=partial write=bulk_delete_worklogs]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `bulkDeleteWorklogs`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --issue-id-or-key
    issue-worklogs get-issue-worklog - Get issue worklogs [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-id-or-key, --start-at, --max-results, --started-after, --started-before, --expand
    issue-worklogs add-worklog - Add worklog [intent=reverse_etl availability=partial write=add_worklog]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `addWorklog`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --issue-id-or-key
    issue-worklogs bulk-move-worklogs - Bulk move worklogs [intent=reverse_etl availability=partial write=bulk_move_worklogs]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `bulkMoveWorklogs`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --issue-id-or-key
    issue-worklogs delete-worklog - Delete worklog [intent=reverse_etl availability=implemented write=delete_worklog]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteWorklog`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --issue-id-or-key, --id
    issue-worklogs get-worklog - Get worklog [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-id-or-key, --id, --expand
    issue-worklogs update-worklog - Update worklog [intent=reverse_etl availability=partial write=update_worklog]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateWorklog`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --issue-id-or-key, --id
    issue-worklog-properties get-worklog-property-keys - Get worklog property keys [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-id-or-key, --worklog-id
    issue-worklog-properties delete-worklog-property - Delete worklog property [intent=reverse_etl availability=implemented write=delete_worklog_property]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteWorklogProperty`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --issue-id-or-key, --worklog-id, --property-key
    issue-worklog-properties get-worklog-property - Get worklog property [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-id-or-key, --worklog-id, --property-key
    issue-worklog-properties set-worklog-property - Set worklog property [intent=reverse_etl availability=planned]; notes: Provider-style command is planned until raw or dynamic JSON request-body support can represent this required body safely.
    issue-links link-issues - Create issue link [intent=reverse_etl availability=partial write=link_issues]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `linkIssues`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits structured JSON flag support for required object fields.
    issue-links delete-issue-link - Delete issue link [intent=reverse_etl availability=implemented write=delete_issue_link]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteIssueLink`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --link-id
    issue-links get-issue-link - Get issue link [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --link-id
    issue-link-types get-issue-link-types - Get issue link types [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    issue-link-types create-issue-link-type - Create issue link type [intent=reverse_etl availability=partial write=create_issue_link_type]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createIssueLinkType`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.
    issue-link-types delete-issue-link-type - Delete issue link type [intent=reverse_etl availability=implemented write=delete_issue_link_type]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteIssueLinkType`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --issue-link-type-id
    issue-link-types get-issue-link-type - Get issue link type [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-link-type-id
    issue-link-types update-issue-link-type - Update issue link type [intent=reverse_etl availability=partial write=update_issue_link_type]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateIssueLinkType`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --issue-link-type-id
    issue-security-schemes get-issue-security-schemes - Get issue security schemes [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    issue-security-schemes create-issue-security-scheme - Create issue security scheme [intent=reverse_etl availability=implemented write=create_issue_security_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createIssueSecurityScheme`.; flags: --name
    issue-security-schemes get-security-levels - Get issue security levels [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --id, --scheme-id, --only-default
    issue-security-schemes set-default-levels - Set default issue security levels [intent=reverse_etl availability=partial write=set_default_levels]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `setDefaultLevels`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --default-values
    issue-security-schemes get-security-level-members - Get issue security level members [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --id, --scheme-id, --level-id, --expand
    issue-security-schemes search-projects-using-security-schemes - Get projects using issue security schemes [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --issue-security-scheme-id, --project-id
    issue-security-schemes associate-schemes-to-projects - Associate security scheme to project [intent=reverse_etl availability=implemented write=associate_schemes_to_projects]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `associateSchemesToProjects`.; flags: --project-id, --scheme-id
    issue-security-schemes search-security-schemes - Search issue security schemes [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --id, --project-id
    issue-security-schemes get-issue-security-scheme - Get issue security scheme [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    issue-security-schemes update-issue-security-scheme - Update issue security scheme [intent=reverse_etl availability=partial write=update_issue_security_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateIssueSecurityScheme`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --id
    issue-security-level get-issue-security-level-members - Get issue security level members by issue security scheme [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-security-scheme-id, --start-at, --max-results, --issue-security-level-id, --expand
    issue-security-schemes delete-security-scheme - Delete issue security scheme [intent=reverse_etl availability=implemented write=delete_security_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteSecurityScheme`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --scheme-id
    issue-security-schemes add-security-level - Add issue security levels [intent=reverse_etl availability=partial write=add_security_level]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `addSecurityLevel`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --scheme-id
    issue-security-schemes remove-level - Remove issue security level [intent=reverse_etl availability=implemented write=remove_level]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `removeLevel`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --scheme-id, --level-id
    issue-security-schemes update-security-level - Update issue security level [intent=reverse_etl availability=partial write=update_security_level]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateSecurityLevel`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --scheme-id, --level-id
    issue-security-schemes add-security-level-members - Add issue security level members [intent=reverse_etl availability=partial write=add_security_level_members]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `addSecurityLevelMembers`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --scheme-id, --level-id
    issue-security-schemes remove-member-from-security-level - Remove member from issue security level [intent=reverse_etl availability=implemented write=remove_member_from_security_level]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `removeMemberFromSecurityLevel`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --scheme-id, --level-id, --member-id
    issue-types get-issue-all-types - Get all issue types for user [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    issue-types create-issue-type - Create issue type [intent=reverse_etl availability=implemented write=create_issue_type]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createIssueType`.; flags: --name
    issue-types get-issue-types-for-project - Get issue types for project [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id, --level
    issue-types delete-issue-type - Delete issue type [intent=reverse_etl availability=implemented write=delete_issue_type]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteIssueType`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    issue-types get-issue-type - Get issue type [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    issue-types update-issue-type - Update issue type [intent=reverse_etl availability=partial write=update_issue_type]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateIssueType`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --id
    issue-types get-alternative-issue-types - Get alternative issue types [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    issue-type-properties get-issue-type-property-keys - Get issue type property keys [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-type-id
    issue-type-properties delete-issue-type-property - Delete issue type property [intent=reverse_etl availability=implemented write=delete_issue_type_property]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteIssueTypeProperty`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --issue-type-id, --property-key
    issue-type-properties get-issue-type-property - Get issue type property [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-type-id, --property-key
    issue-type-properties set-issue-type-property - Set issue type property [intent=reverse_etl availability=planned]; notes: Provider-style command is planned until raw or dynamic JSON request-body support can represent this required body safely.
    issue-type-schemes get-all-issue-type-schemes - Get all issue type schemes [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --id, --order-by, --expand, --query-string
    issue-type-schemes create-issue-type-scheme - Create issue type scheme [intent=reverse_etl availability=implemented write=create_issue_type_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createIssueTypeScheme`.; flags: --issue-type-ids, --name
    issue-type-schemes get-issue-type-schemes-mapping - Get issue type scheme items [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --issue-type-scheme-id
    issue-type-schemes get-issue-type-scheme-for-projects - Get issue type schemes for projects [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --project-id
    issue-type-schemes assign-issue-type-scheme-to-project - Assign issue type scheme to project [intent=reverse_etl availability=implemented write=assign_issue_type_scheme_to_project]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `assignIssueTypeSchemeToProject`.; flags: --issue-type-scheme-id, --project-id
    issue-type-schemes delete-issue-type-scheme - Delete issue type scheme [intent=reverse_etl availability=implemented write=delete_issue_type_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteIssueTypeScheme`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --issue-type-scheme-id
    issue-type-schemes update-issue-type-scheme - Update issue type scheme [intent=reverse_etl availability=partial write=update_issue_type_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateIssueTypeScheme`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --issue-type-scheme-id
    issue-type-schemes add-issue-types-to-issue-type-scheme - Add issue types to issue type scheme [intent=reverse_etl availability=implemented write=add_issue_types_to_issue_type_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `addIssueTypesToIssueTypeScheme`.; flags: --issue-type-scheme-id, --issue-type-ids
    issue-type-schemes reorder-issue-types-in-issue-type-scheme - Change order of issue types [intent=reverse_etl availability=implemented write=reorder_issue_types_in_issue_type_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `reorderIssueTypesInIssueTypeScheme`.; flags: --issue-type-scheme-id, --issue-type-ids
    issue-type-schemes remove-issue-type-from-issue-type-scheme - Remove issue type from issue type scheme [intent=reverse_etl availability=implemented write=remove_issue_type_from_issue_type_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `removeIssueTypeFromIssueTypeScheme`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --issue-type-scheme-id, --issue-type-id
    issue-type-screen-schemes get-issue-type-screen-schemes - Get issue type screen schemes [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --id, --query-string, --order-by, --expand
    issue-type-screen-schemes create-issue-type-screen-scheme - Create issue type screen scheme [intent=reverse_etl availability=partial write=create_issue_type_screen_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createIssueTypeScreenScheme`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --issue-type-mappings, --name
    issue-type-screen-schemes get-issue-type-screen-scheme-mappings - Get issue type screen scheme items [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --issue-type-screen-scheme-id
    issue-type-screen-schemes get-issue-type-screen-scheme-project-associations - Get issue type screen schemes for projects [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --project-id
    issue-type-screen-schemes assign-issue-type-screen-scheme-to-project - Assign issue type screen scheme to project [intent=reverse_etl availability=partial write=assign_issue_type_screen_scheme_to_project]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `assignIssueTypeScreenSchemeToProject`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.
    issue-type-screen-schemes delete-issue-type-screen-scheme - Delete issue type screen scheme [intent=reverse_etl availability=implemented write=delete_issue_type_screen_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteIssueTypeScreenScheme`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --issue-type-screen-scheme-id
    issue-type-screen-schemes update-issue-type-screen-scheme - Update issue type screen scheme [intent=reverse_etl availability=partial write=update_issue_type_screen_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateIssueTypeScreenScheme`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --issue-type-screen-scheme-id
    issue-type-screen-schemes append-mappings-for-issue-type-screen-scheme - Append mappings to issue type screen scheme [intent=reverse_etl availability=partial write=append_mappings_for_issue_type_screen_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `appendMappingsForIssueTypeScreenScheme`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --issue-type-screen-scheme-id, --issue-type-mappings
    issue-type-screen-schemes update-default-screen-scheme - Update issue type screen scheme default screen scheme [intent=reverse_etl availability=implemented write=update_default_screen_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateDefaultScreenScheme`.; flags: --issue-type-screen-scheme-id, --screen-scheme-id
    issue-type-screen-schemes remove-mappings-from-issue-type-screen-scheme - Remove mappings from issue type screen scheme [intent=reverse_etl availability=implemented write=remove_mappings_from_issue_type_screen_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration with destructive semantics for operation `removeMappingsFromIssueTypeScreenScheme`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --issue-type-screen-scheme-id, --issue-type-ids
    issue-type-screen-schemes get-projects-for-issue-type-screen-scheme - Get issue type screen scheme projects [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --issue-type-screen-scheme-id, --start-at, --max-results, --query
    jql get-auto-complete - Get field reference data (GET) [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    jql get-auto-complete-post - Get field reference data (POST) [intent=direct_read availability=partial]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; notes: Provider-style command awaits typed integer/object array body flag support before it can run safely.; flags: --include-collapsed-fields, --project-ids
    jql get-field-auto-complete-for-query-string - Get field auto complete suggestions [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --field-name, --field-value, --predicate-name, --predicate-value
    jql-functions-apps get-precomputations - Get precomputations (apps) [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --function-key, --start-at, --max-results, --order-by
    jql-functions-apps update-precomputations - Update precomputations (apps) [intent=reverse_etl availability=partial write=update_precomputations]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updatePrecomputations`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    jql-functions-apps get-precomputations-by-id - Get precomputations by ID (apps) [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; flags: --precomputation-i-ds, --order-by
    issue-search match-issues - Check issues against JQL [intent=direct_read availability=partial]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; notes: Provider-style command awaits typed integer/object array body flag support before it can run safely.; flags: --issue-ids, --jqls
    jql parse-jql-queries - Parse JQL query [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; flags: --queries, --validation
    jql migrate-queries - Convert user identifiers to account IDs in JQL queries [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; flags: --query-strings
    jql sanitise-jql-queries - Sanitize JQL queries [intent=direct_read availability=partial]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; notes: Provider-style command awaits typed integer/object array body flag support before it can run safely.; flags: --queries
    labels get-all-labels - Get all labels [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results
    license-metrics get-approximate-license-count - Get approximate license count [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    license-metrics get-approximate-application-license-count - Get approximate application license count [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --application-key
    permissions get-my-permissions - Get my permissions [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-key, --project-id, --issue-key, --issue-id, --permissions, --project-uuid, --project-configuration-uuid, --comment-id
    myself remove-preference - Delete preference [intent=reverse_etl availability=implemented write=remove_preference]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `removePreference`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --key
    myself get-preference - Get preference [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --key
    myself get-locale - Get locale [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    myself get-current-user - Get current user [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --expand
    issue-notification-schemes get-notification-schemes - Get notification schemes paginated [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --id, --project-id, --only-default, --expand
    issue-notification-schemes create-notification-scheme - Create notification scheme [intent=reverse_etl availability=implemented write=create_notification_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createNotificationScheme`.; flags: --name
    issue-notification-schemes get-notification-scheme-to-project-mappings - Get projects using notification schemes paginated [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --notification-scheme-id, --project-id
    issue-notification-schemes get-notification-scheme - Get notification scheme [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id, --expand
    issue-notification-schemes update-notification-scheme - Update notification scheme [intent=reverse_etl availability=partial write=update_notification_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateNotificationScheme`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --id
    issue-notification-schemes add-notifications - Add notifications to notification scheme [intent=reverse_etl availability=partial write=add_notifications]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `addNotifications`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --id, --notification-scheme-events
    issue-notification-schemes delete-notification-scheme - Delete notification scheme [intent=reverse_etl availability=implemented write=delete_notification_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteNotificationScheme`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --notification-scheme-id
    issue-notification-schemes remove-notification-from-notification-scheme - Remove notification from notification scheme [intent=reverse_etl availability=implemented write=remove_notification_from_notification_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `removeNotificationFromNotificationScheme`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --notification-scheme-id, --notification-id
    permissions get-all-permissions - Get all permissions [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    permissions get-bulk-permissions - Get bulk permissions [intent=direct_read availability=partial]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; notes: Provider-style command awaits typed integer/object array body flag support before it can run safely.; flags: --account-id, --global-permissions, --project-permissions
    permissions get-permitted-projects - Get permitted projects [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; flags: --permissions
    permission-schemes get-all-permission-schemes - Get all permission schemes [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --expand
    permission-schemes create-permission-scheme - Create permission scheme [intent=reverse_etl availability=implemented write=create_permission_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createPermissionScheme`.; flags: --name
    permission-schemes delete-permission-scheme - Delete permission scheme [intent=reverse_etl availability=implemented write=delete_permission_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deletePermissionScheme`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --scheme-id
    permission-schemes get-permission-scheme - Get permission scheme [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --scheme-id, --expand
    permission-schemes update-permission-scheme - Update permission scheme [intent=reverse_etl availability=implemented write=update_permission_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updatePermissionScheme`.; flags: --scheme-id, --name
    permission-schemes get-permission-scheme-grants - Get permission scheme grants [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --scheme-id, --expand
    permission-schemes create-permission-grant - Create permission grant [intent=reverse_etl availability=partial write=create_permission_grant]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createPermissionGrant`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --scheme-id
    permission-schemes delete-permission-scheme-entity - Delete permission scheme grant [intent=reverse_etl availability=implemented write=delete_permission_scheme_entity]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deletePermissionSchemeEntity`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --scheme-id, --permission-id
    permission-schemes get-permission-scheme-grant - Get permission scheme grant [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --scheme-id, --permission-id, --expand
    plans get-plans - Get plans paginated [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --include-trashed, --include-archived, --cursor, --max-results
    plans create-plan - Create plan [intent=reverse_etl availability=partial write=create_plan]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createPlan`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits structured JSON flag support for required object fields.; flags: --issue-sources, --name
    plans get-plan - Get plan [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --plan-id, --use-group-id
    plans update-plan - Update plan [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updatePlan`.; notes: Blocked until declarative writes support application/json-patch+json request bodies.; flags: --plan-id
    plans archive-plan - Archive plan [intent=reverse_etl availability=implemented write=archive_plan]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration with destructive semantics for operation `archivePlan`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --plan-id
    plans duplicate-plan - Duplicate plan [intent=reverse_etl availability=implemented write=duplicate_plan]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `duplicatePlan`.; flags: --plan-id, --name
    teams-in-plan get-teams - Get teams in plan paginated [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --plan-id, --cursor, --max-results
    teams-in-plan add-atlassian-team - Add Atlassian team to plan [intent=reverse_etl availability=implemented write=add_atlassian_team]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `addAtlassianTeam`.; flags: --plan-id, --id, --planning-style
    teams-in-plan remove-atlassian-team - Remove Atlassian team from plan [intent=reverse_etl availability=implemented write=remove_atlassian_team]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `removeAtlassianTeam`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --plan-id, --atlassian-team-id
    teams-in-plan get-atlassian-team - Get Atlassian team in plan [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --plan-id, --atlassian-team-id
    teams-in-plan update-atlassian-team - Update Atlassian team in plan [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateAtlassianTeam`.; notes: Blocked until declarative writes support application/json-patch+json request bodies.; flags: --plan-id, --atlassian-team-id
    teams-in-plan create-plan-only-team - Create plan-only team [intent=reverse_etl availability=implemented write=create_plan_only_team]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createPlanOnlyTeam`.; flags: --plan-id, --name, --planning-style
    teams-in-plan delete-plan-only-team - Delete plan-only team [intent=reverse_etl availability=implemented write=delete_plan_only_team]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deletePlanOnlyTeam`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --plan-id, --plan-only-team-id
    teams-in-plan get-plan-only-team - Get plan-only team [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --plan-id, --plan-only-team-id
    teams-in-plan update-plan-only-team - Update plan-only team [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updatePlanOnlyTeam`.; notes: Blocked until declarative writes support application/json-patch+json request bodies.; flags: --plan-id, --plan-only-team-id
    plans trash-plan - Trash plan [intent=reverse_etl availability=implemented write=trash_plan]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration with destructive semantics for operation `trashPlan`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --plan-id
    issue-priorities create-priority - Create priority [intent=reverse_etl availability=implemented write=create_priority]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createPriority`.; flags: --name, --status-color
    issue-priorities set-default-priority - Set default priority [intent=reverse_etl availability=implemented write=set_default_priority]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `setDefaultPriority`.; flags: --id
    issue-priorities move-priorities - Move priorities [intent=reverse_etl availability=implemented write=move_priorities]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `movePriorities`.; flags: --ids
    issue-priorities search-priorities - Search priorities [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --id, --project-id, --priority-name, --only-default, --expand
    issue-priorities delete-priority - Delete priority [intent=reverse_etl availability=implemented write=delete_priority]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deletePriority`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    issue-priorities get-priority - Get priority [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    issue-priorities update-priority - Update priority [intent=reverse_etl availability=partial write=update_priority]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updatePriority`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --id
    priority-schemes get-priority-schemes - Get priority schemes [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --priority-id, --scheme-id, --scheme-name, --only-default, --order-by, --expand
    priority-schemes create-priority-scheme - Create priority scheme [intent=reverse_etl availability=partial write=create_priority_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createPriorityScheme`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --default-priority-id, --name, --priority-ids
    priority-schemes suggested-priorities-for-mappings - Suggested priorities for mappings [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; flags: --max-results, --scheme-id, --start-at
    priority-schemes get-available-priorities-by-priority-scheme - Get available priorities by priority scheme [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --query, --scheme-id, --exclude
    priority-schemes delete-priority-scheme - Delete priority scheme [intent=reverse_etl availability=implemented write=delete_priority_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deletePriorityScheme`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --scheme-id
    priority-schemes update-priority-scheme - Update priority scheme [intent=reverse_etl availability=partial write=update_priority_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updatePriorityScheme`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --scheme-id
    priority-schemes get-priorities-by-priority-scheme - Get priorities by priority scheme [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --scheme-id
    priority-schemes get-projects-by-priority-scheme - Get projects by priority scheme [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --project-id, --scheme-id, --query
    projects create-project - Create project [intent=reverse_etl availability=implemented write=create_project]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createProject`.; flags: --key, --name
    project-templates create-project-with-custom-template - Create custom project [intent=reverse_etl availability=partial write=create_project_with_custom_template]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createProjectWithCustomTemplate`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    project-templates edit-template - Edit a custom project template [intent=reverse_etl availability=partial write=edit_template]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `editTemplate`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    project-templates live-template - Gets a custom project template [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id, --template-key
    project-templates remove-template - Deletes a custom project template [intent=reverse_etl availability=implemented write=remove_template]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `removeTemplate`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --template-key
    project-templates save-template - Save a custom project template [intent=reverse_etl availability=partial write=save_template]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `saveTemplate`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    projects get-recent - Get recent projects [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --expand, --properties
    project-types get-all-project-types - Get all project types [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    project-types get-all-accessible-project-types - Get licensed project types [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    project-types get-project-type-by-key - Get project type by key [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-type-key
    project-types get-accessible-project-type-by-key - Get accessible project type by key [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-type-key
    projects delete-project - Delete project [intent=reverse_etl availability=implemented write=delete_project]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteProject`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --project-id-or-key
    projects get-project - Get project [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id-or-key, --expand, --properties
    projects update-project - Update project [intent=reverse_etl availability=partial write=update_project]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateProject`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --project-id-or-key
    projects archive-project - Archive project [intent=reverse_etl availability=implemented write=archive_project]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration with destructive semantics for operation `archiveProject`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --project-id-or-key
    project-avatars update-project-avatar - Set project avatar [intent=reverse_etl availability=implemented write=update_project_avatar]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateProjectAvatar`.; flags: --project-id-or-key, --id
    project-avatars delete-project-avatar - Delete project avatar [intent=reverse_etl availability=implemented write=delete_project_avatar]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteProjectAvatar`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --project-id-or-key, --id
    project-avatars get-all-project-avatars - Get all project avatars [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id-or-key
    project-classification-levels get-project-classification-config - Get the classification configuration for a project [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id-or-key
    project-classification-levels remove-default-project-classification - Remove the default data classification level from a project [intent=reverse_etl availability=implemented write=remove_default_project_classification]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `removeDefaultProjectClassification`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --project-id-or-key
    project-classification-levels get-default-project-classification - Get the default data classification level of a project [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id-or-key
    project-classification-levels update-default-project-classification - Update the default data classification level of a project [intent=reverse_etl availability=implemented write=update_default_project_classification]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateDefaultProjectClassification`.; flags: --project-id-or-key, --id
    project-components get-project-components-paginated - Get project components paginated [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id-or-key, --start-at, --max-results, --order-by, --component-source, --query
    project-components get-project-components - Get project components [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id-or-key, --component-source
    projects delete-project-asynchronously - Delete project asynchronously [intent=reverse_etl availability=implemented write=delete_project_asynchronously]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration with destructive semantics for operation `deleteProjectAsynchronously`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --project-id-or-key
    project-features get-features-for-project - Get project features [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id-or-key
    project-features toggle-feature-for-project - Set project feature state [intent=reverse_etl availability=partial write=toggle_feature_for_project]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `toggleFeatureForProject`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --project-id-or-key, --feature-key
    project-properties get-project-property-keys - Get project property keys [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id-or-key
    project-properties delete-project-property - Delete project property [intent=reverse_etl availability=implemented write=delete_project_property]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteProjectProperty`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --project-id-or-key, --property-key
    project-properties get-project-property - Get project property [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id-or-key, --property-key
    project-properties set-project-property - Set project property [intent=reverse_etl availability=planned]; notes: Provider-style command is planned until raw or dynamic JSON request-body support can represent this required body safely.
    projects restore - Restore deleted or archived project [intent=reverse_etl availability=implemented write=restore]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration with destructive semantics for operation `restore`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --project-id-or-key
    project-roles get-project-roles - Get project roles for project [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id-or-key
    project-role-actors delete-actor - Delete actors from project role [intent=reverse_etl availability=implemented write=delete_actor]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteActor`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --project-id-or-key, --id
    project-roles get-project-role - Get project role for project [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id-or-key, --id, --exclude-inactive-users
    project-role-actors add-actor-users - Add actors to project role [intent=reverse_etl availability=partial write=add_actor_users]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `addActorUsers`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --project-id-or-key, --id
    project-role-actors set-actors - Set actors for project role [intent=reverse_etl availability=partial write=set_actors]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `setActors`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --project-id-or-key, --id
    project-roles get-project-role-details - Get project role details [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id-or-key, --current-member, --exclude-connect-addons, --exclude-other-service-roles
    projects get-all-statuses - Get all statuses for project [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id-or-key
    project-versions get-project-versions-paginated - Get project versions paginated [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id-or-key, --start-at, --max-results, --order-by, --query, --status, --expand
    project-versions get-project-versions - Get project versions [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id-or-key, --expand
    project-email get-project-email - Get project's sender email [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id
    project-email update-project-email - Set project's sender email [intent=reverse_etl availability=partial write=update_project_email]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateProjectEmail`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --project-id
    projects get-hierarchy - Get project issue type hierarchy [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id
    project-permission-schemes get-project-issue-security-scheme - Get project issue security scheme [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-key-or-id
    projects get-notification-scheme-for-project - Get project notification scheme [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-key-or-id, --expand
    project-permission-schemes get-assigned-permission-scheme - Get assigned permission scheme [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-key-or-id, --expand
    project-permission-schemes assign-permission-scheme - Assign permission scheme [intent=reverse_etl availability=implemented write=assign_permission_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `assignPermissionScheme`.; flags: --project-key-or-id, --id
    project-permission-schemes get-security-levels-for-project - Get project issue security levels [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-key-or-id
    project-categories get-all-project-categories - Get all project categories [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    project-categories create-project-category - Create project category [intent=reverse_etl availability=partial write=create_project_category]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createProjectCategory`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.
    project-categories remove-project-category - Delete project category [intent=reverse_etl availability=implemented write=remove_project_category]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `removeProjectCategory`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    project-categories get-project-category-by-id - Get project category by ID [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    project-categories update-project-category - Update project category [intent=reverse_etl availability=partial write=update_project_category]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateProjectCategory`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --id
    issue-fields get-project-fields - Get fields for projects [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --project-id, --work-type-id, --field-id
    project-key-and-name-validation validate-project-key - Validate project key [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --key
    project-key-and-name-validation get-valid-project-key - Get valid project key [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --key
    project-key-and-name-validation get-valid-project-name - Get valid project name [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --name
    issue-redaction redact - Redact [intent=reverse_etl availability=partial write=redact]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `redact`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    issue-redaction get-redaction-status - Get redaction status [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --job-id
    issue-resolutions create-resolution - Create resolution [intent=reverse_etl availability=implemented write=create_resolution]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createResolution`.; flags: --name
    issue-resolutions set-default-resolution - Set default resolution [intent=reverse_etl availability=implemented write=set_default_resolution]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `setDefaultResolution`.; flags: --id
    issue-resolutions move-resolutions - Move resolutions [intent=reverse_etl availability=implemented write=move_resolutions]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `moveResolutions`.; flags: --ids
    issue-resolutions search-resolutions - Search resolutions [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --id, --only-default
    issue-resolutions delete-resolution - Delete resolution [intent=reverse_etl availability=implemented write=delete_resolution]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteResolution`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id, --replace-with
    issue-resolutions get-resolution - Get resolution [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    issue-resolutions update-resolution - Update resolution [intent=reverse_etl availability=implemented write=update_resolution]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateResolution`.; flags: --id, --name
    project-roles get-all-project-roles - Get all project roles [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    project-roles create-project-role - Create project role [intent=reverse_etl availability=partial write=create_project_role]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createProjectRole`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.
    project-roles delete-project-role - Delete project role [intent=reverse_etl availability=implemented write=delete_project_role]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteProjectRole`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    project-roles get-project-role-by-id - Get project role by ID [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    project-roles partial-update-project-role - Partial update project role [intent=reverse_etl availability=partial write=partial_update_project_role]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `partialUpdateProjectRole`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --id
    project-roles fully-update-project-role - Fully update project role [intent=reverse_etl availability=partial write=fully_update_project_role]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `fullyUpdateProjectRole`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --id
    project-role-actors delete-project-role-actors-from-role - Delete default actors from project role [intent=reverse_etl availability=implemented write=delete_project_role_actors_from_role]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteProjectRoleActorsFromRole`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    project-role-actors get-project-role-actors-for-role - Get default actors for project role [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    project-role-actors add-project-role-actors-to-role - Add default actors to project role [intent=reverse_etl availability=partial write=add_project_role_actors_to_role]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `addProjectRoleActorsToRole`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --id
    screens get-screens - Get screens [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --id, --query-string, --scope, --order-by
    screens create-screen - Create screen [intent=reverse_etl availability=implemented write=create_screen]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createScreen`.; flags: --name
    screens add-field-to-default-screen - Add field to default screen [intent=reverse_etl availability=implemented write=add_field_to_default_screen]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `addFieldToDefaultScreen`.; flags: --field-id
    screen-tabs get-bulk-screen-tabs - Get bulk screen tabs [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --screen-id, --tab-id, --start-at, --max-result
    screens delete-screen - Delete screen [intent=reverse_etl availability=implemented write=delete_screen]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteScreen`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --screen-id
    screens update-screen - Update screen [intent=reverse_etl availability=partial write=update_screen]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateScreen`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --screen-id
    screens get-available-screen-fields - Get available screen fields [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --screen-id
    screen-tabs get-all-screen-tabs - Get all screen tabs [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --screen-id, --project-key
    screen-tabs add-screen-tab - Create screen tab [intent=reverse_etl availability=implemented write=add_screen_tab]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `addScreenTab`.; flags: --screen-id, --name
    screen-tabs delete-screen-tab - Delete screen tab [intent=reverse_etl availability=implemented write=delete_screen_tab]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteScreenTab`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --screen-id, --tab-id
    screen-tabs rename-screen-tab - Update screen tab [intent=reverse_etl availability=implemented write=rename_screen_tab]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `renameScreenTab`.; flags: --screen-id, --tab-id, --name
    screen-tab-fields get-all-screen-tab-fields - Get all screen tab fields [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --screen-id, --tab-id, --project-key
    screen-tab-fields add-screen-tab-field - Add screen tab field [intent=reverse_etl availability=implemented write=add_screen_tab_field]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `addScreenTabField`.; flags: --screen-id, --tab-id, --field-id
    screen-tab-fields remove-screen-tab-field - Remove screen tab field [intent=reverse_etl availability=implemented write=remove_screen_tab_field]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `removeScreenTabField`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --screen-id, --tab-id, --id
    screen-tab-fields move-screen-tab-field - Move screen tab field [intent=reverse_etl availability=partial write=move_screen_tab_field]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `moveScreenTabField`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --screen-id, --tab-id, --id
    screen-tabs move-screen-tab - Move screen tab [intent=reverse_etl availability=implemented write=move_screen_tab]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `moveScreenTab`.; flags: --screen-id, --tab-id, --pos
    screen-schemes get-screen-schemes - Get screen schemes [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --id, --expand, --query-string, --order-by
    screen-schemes create-screen-scheme - Create screen scheme [intent=reverse_etl availability=partial write=create_screen_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createScreenScheme`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits structured JSON flag support for required object fields.; flags: --name
    screen-schemes delete-screen-scheme - Delete screen scheme [intent=reverse_etl availability=implemented write=delete_screen_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteScreenScheme`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --screen-scheme-id
    screen-schemes update-screen-scheme - Update screen scheme [intent=reverse_etl availability=partial write=update_screen_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateScreenScheme`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --screen-scheme-id
    issue-search count-issues - Count issues using JQL [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; flags: --jql
    issue-search search-and-reconsile-issues-using-jql - Search for issues using JQL enhanced search (GET) [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --jql, --next-page-token, --max-results, --fields, --expand, --properties, --fields-by-keys, --fail-fast, --reconcile-issues
    issue-search search-and-reconsile-issues-using-jql-post - Search for issues using JQL enhanced search (POST) [intent=direct_read availability=partial]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; notes: Provider-style command awaits typed integer/object array body flag support before it can run safely.; flags: --expand, --fields, --fields-by-keys, --jql, --max-results, --next-page-token, --properties, --reconcile-issues
    issue-security-level get-issue-security-level - Get issue security level [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    server-info get-server-info - Get Jira instance info [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    issue-navigator-settings get-issue-navigator-default-columns - Get issue navigator default columns [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    issue-navigator-settings set-issue-navigator-default-columns - Set issue navigator default columns [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `setIssueNavigatorDefaultColumns`.; notes: Blocked until declarative writes support Jira repeated columns form-field request bodies without file upload.; flags: --columns
    workflow-statuses get-statuses - Get all statuses [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    workflow-statuses get-status - Get status [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id-or-name
    workflow-status-categories get-status-categories - Get all status categories [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    workflow-status-categories get-status-category - Get status category [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id-or-key
    status delete-statuses-by-id - Bulk delete Statuses [intent=reverse_etl availability=implemented write=delete_statuses_by_id]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteStatusesById`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    status get-statuses-by-id - Bulk get statuses [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    status create-statuses - Bulk create statuses [intent=reverse_etl availability=partial write=create_statuses]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createStatuses`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits structured JSON flag support for required object fields.; flags: --statuses
    status update-statuses - Bulk update statuses [intent=reverse_etl availability=partial write=update_statuses]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateStatuses`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --statuses
    status get-statuses-by-name - Bulk get statuses by name [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --name, --project-id
    status search - Search statuses paginated [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id, --start-at, --max-results, --search-string, --status-category, --include-global-statuses
    status get-project-issue-type-usages-for-status - Get issue type usages by status and project [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --status-id, --project-id, --next-page-token, --max-results
    status get-project-usages-for-status - Get project usages by status [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --status-id, --next-page-token, --max-results
    status get-workflow-usages-for-status - Get workflow usages by status [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --status-id, --next-page-token, --max-results
    tasks get-task - Get task [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --task-id
    tasks cancel-task - Cancel task [intent=reverse_etl availability=implemented write=cancel_task]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration with destructive semantics for operation `cancelTask`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --task-id
    ui-modifications-apps get-ui-modifications - Get UI modifications [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --expand
    ui-modifications-apps create-ui-modification - Create UI modification [intent=reverse_etl availability=implemented write=create_ui_modification]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createUiModification`.; flags: --name
    ui-modifications-apps delete-ui-modification - Delete UI modification [intent=reverse_etl availability=implemented write=delete_ui_modification]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteUiModification`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --ui-modification-id
    ui-modifications-apps update-ui-modification - Update UI modification [intent=reverse_etl availability=partial write=update_ui_modification]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateUiModification`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --ui-modification-id
    avatars get-avatars - Get avatars [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --type, --entity-id
    avatars delete-avatar - Delete avatar [intent=reverse_etl availability=implemented write=delete_avatar]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteAvatar`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --type, --owning-object-id, --id
    avatars get-avatar-image-by-type - Get avatar image by type [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --type, --size, --format
    avatars get-avatar-image-by-id - Get avatar image by ID [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --type, --id, --size, --format
    avatars get-avatar-image-by-owner - Get avatar image by owner [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --type, --entity-id, --size, --format
    users remove-user - Delete user [intent=reverse_etl availability=implemented write=remove_user]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `removeUser`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --account-id
    users get-user - Get user [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --account-id, --username, --key, --expand
    users create-user - Create user [intent=reverse_etl availability=implemented write=create_user]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createUser`.; flags: --email-address, --products
    user-search find-bulk-assignable-users - Find users assignable to projects [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --query, --username, --account-id, --project-keys, --start-at, --max-results
    user-search find-assignable-users - Find users assignable to issues [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --query, --session-id, --username, --account-id, --project, --issue-key, --issue-id, --start-at, --max-results, --action-descriptor-id, --recommend, --account-type, --app-type
    users bulk-get-users - Bulk get users [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --username, --key, --account-id
    users bulk-get-users-migration - Get account IDs for users [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --username, --key
    users reset-user-columns - Reset user default columns [intent=reverse_etl availability=implemented write=reset_user_columns]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `resetUserColumns`.; notes: Requires --confirm destructive when executing the approved reverse plan.
    users get-user-default-columns - Get user default columns [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --account-id, --username
    users set-user-columns - Set user default columns [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `setUserColumns`.; notes: Blocked until declarative writes support Jira repeated columns form-field request bodies without file upload.; flags: --columns
    users get-user-email - Get user email [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --account-id
    users get-user-email-bulk - Get user email bulk [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --account-id
    users get-user-groups - Get user groups [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --account-id, --username, --key
    user-search find-users-with-all-permissions - Find users with permissions [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --query, --username, --account-id, --permissions, --issue-key, --project-key, --start-at, --max-results
    user-search find-users-for-picker - Find users for picker [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --query, --max-results, --show-avatar, --exclude, --exclude-account-ids, --avatar-size, --exclude-connect-users
    user-properties get-user-property-keys - Get user property keys [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --account-id, --user-key, --username
    user-properties delete-user-property - Delete user property [intent=reverse_etl availability=implemented write=delete_user_property]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteUserProperty`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --property-key
    user-properties get-user-property - Get user property [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --account-id, --user-key, --username, --property-key
    user-properties set-user-property - Set user property [intent=reverse_etl availability=planned]; notes: Provider-style command is planned until raw or dynamic JSON request-body support can represent this required body safely.
    user-search find-users - Find users [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --query, --username, --account-id, --start-at, --max-results, --property
    user-search find-users-by-query - Find users by query [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --query, --start-at, --max-results
    user-search find-user-keys-by-query - Find user keys by query [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --query, --start-at, --max-result
    user-search find-users-with-browse-permission - Find users with browse permission [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --query, --username, --account-id, --issue-key, --project-key, --start-at, --max-results
    users get-all-users-default - Get all users default [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --expand
    project-versions create-version - Create version [intent=reverse_etl availability=partial write=create_version]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createVersion`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    project-versions get-version - Get version [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id, --expand
    project-versions update-version - Update version [intent=reverse_etl availability=partial write=update_version]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateVersion`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --id
    project-versions merge-versions - Merge versions [intent=reverse_etl availability=implemented write=merge_versions]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `mergeVersions`.; flags: --id, --move-issues-to
    project-versions move-version - Move version [intent=reverse_etl availability=partial write=move_version]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `moveVersion`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --id
    project-versions get-version-related-issues - Get version's related issues count [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    project-versions get-related-work - Get related work [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    project-versions create-related-work - Create related work [intent=reverse_etl availability=implemented write=create_related_work]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createRelatedWork`.; flags: --id, --category
    project-versions update-related-work - Update related work [intent=reverse_etl availability=implemented write=update_related_work]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateRelatedWork`.; flags: --id, --category
    project-versions delete-and-replace-version - Delete and replace version [intent=reverse_etl availability=partial write=delete_and_replace_version]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration with destructive semantics for operation `deleteAndReplaceVersion`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --id
    project-versions get-version-unresolved-issues - Get version's unresolved issues count [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    project-versions delete-related-work - Delete related work [intent=reverse_etl availability=implemented write=delete_related_work]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteRelatedWork`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --version-id, --related-work-id
    webhooks delete-webhook-by-id - Delete webhooks by ID [intent=reverse_etl availability=partial write=delete_webhook_by_id]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteWebhookById`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    webhooks get-dynamic-webhooks-for-app - Get dynamic webhooks for app [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results
    webhooks register-dynamic-webhooks - Register dynamic webhooks [intent=reverse_etl availability=partial write=register_dynamic_webhooks]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `registerDynamicWebhooks`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --url, --webhooks
    webhooks get-failed-webhooks - Get failed webhooks [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --max-results, --after
    webhooks refresh-webhooks - Extend webhook life [intent=reverse_etl availability=partial write=refresh_webhooks]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `refreshWebhooks`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --webhook-ids
    workflows read-workflow-from-history - Read workflow version from history [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; flags: --version, --workflow-id
    workflows list-workflow-history - List workflow history entries [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; flags: --workflow-id, --expand
    workflow-transition-rules get-workflow-transition-rule-configurations - Get workflow transition rule configurations [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --types, --keys, --workflow-names, --with-tags, --draft, --expand
    workflow-transition-rules update-workflow-transition-rule-configurations - Update workflow transition rule configurations [intent=reverse_etl availability=partial write=update_workflow_transition_rule_configurations]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateWorkflowTransitionRuleConfigurations`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --workflows
    workflow-transition-rules delete-workflow-transition-rule-configurations - Delete workflow transition rule configurations [intent=reverse_etl availability=partial write=delete_workflow_transition_rule_configurations]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration with destructive semantics for operation `deleteWorkflowTransitionRuleConfigurations`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --workflows
    workflows delete-inactive-workflow - Delete inactive workflow [intent=reverse_etl availability=implemented write=delete_inactive_workflow]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteInactiveWorkflow`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --entity-id
    workflows get-workflow-project-issue-type-usages - Get issue types in a project that are using a given workflow [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --workflow-id, --project-id, --next-page-token, --max-results
    workflows get-project-usages-for-workflow - Get projects using a given workflow [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --workflow-id, --next-page-token, --max-results
    workflows get-workflow-scheme-usages-for-workflow - Get workflow schemes which are using a given workflow [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --workflow-id, --next-page-token, --max-results
    workflows read-workflows - Bulk get workflows [intent=direct_read availability=partial]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; notes: Provider-style command awaits typed integer/object array body flag support before it can run safely.; flags: --project-and-issue-types, --workflow-ids, --workflow-names
    workflows workflow-capabilities - Get available workflow capabilities [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --workflow-id, --project-id, --issue-type-id
    workflows create-workflows - Bulk create workflows [intent=reverse_etl availability=partial write=create_workflows]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createWorkflows`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    workflows validate-create-workflows - Validate create workflows [intent=direct_read availability=partial]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; notes: Provider-style command awaits structured JSON body flag support for the required payload object.
    workflows get-default-editor - Get the user's default workflow editor [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    workflows read-workflow-previews - Preview workflow [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; flags: --issue-type-ids, --project-id, --workflow-ids, --workflow-names
    workflows search-workflows - Search workflows [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results, --expand, --query-string, --order-by, --scope, --is-active, --project-id
    workflows update-workflows - Bulk update workflows [intent=reverse_etl availability=partial write=update_workflows]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateWorkflows`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    workflows validate-update-workflows - Validate update workflows [intent=direct_read availability=partial]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; notes: Provider-style command awaits structured JSON body flag support for the required payload object.
    workflow-schemes get-all-workflow-schemes - Get all workflow schemes [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --start-at, --max-results
    workflow-schemes create-workflow-scheme - Create workflow scheme [intent=reverse_etl availability=partial write=create_workflow_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createWorkflowScheme`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    workflow-scheme-project-associations get-workflow-scheme-project-associations - Get workflow scheme project associations [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --project-id
    workflow-scheme-project-associations assign-scheme-to-project - Assign workflow scheme to project [intent=reverse_etl availability=implemented write=assign_scheme_to_project]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `assignSchemeToProject`.; flags: --project-id
    workflow-schemes switch-workflow-scheme-for-project - Switch workflow scheme for project [intent=reverse_etl availability=partial write=switch_workflow_scheme_for_project]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `switchWorkflowSchemeForProject`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.
    workflow-schemes read-workflow-schemes - Bulk get workflow schemes [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; flags: --project-ids, --workflow-scheme-ids
    workflow-schemes update-schemes - Update workflow scheme [intent=reverse_etl availability=partial write=update_schemes]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateSchemes`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits structured JSON flag support for required object fields.; flags: --description, --id, --name
    workflow-schemes get-required-workflow-scheme-mappings - Get required status mappings for workflow scheme update [intent=direct_read availability=partial]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; notes: Provider-style command awaits typed integer/object array body flag support before it can run safely.; flags: --default-workflow-id, --id, --workflows-for-issue-types
    workflow-schemes delete-workflow-scheme - Delete workflow scheme [intent=reverse_etl availability=implemented write=delete_workflow_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteWorkflowScheme`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    workflow-schemes get-workflow-scheme - Get workflow scheme [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id, --return-draft-if-exists
    workflow-schemes update-workflow-scheme - Classic update workflow scheme [intent=reverse_etl availability=partial write=update_workflow_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateWorkflowScheme`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --id
    workflow-scheme-drafts create-workflow-scheme-draft-from-parent - Create draft workflow scheme [intent=reverse_etl availability=implemented write=create_workflow_scheme_draft_from_parent]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `createWorkflowSchemeDraftFromParent`.; flags: --id
    workflow-schemes delete-default-workflow - Delete default workflow [intent=reverse_etl availability=implemented write=delete_default_workflow]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteDefaultWorkflow`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    workflow-schemes get-default-workflow - Get default workflow [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id, --return-draft-if-exists
    workflow-schemes update-default-workflow - Update default workflow [intent=reverse_etl availability=implemented write=update_default_workflow]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateDefaultWorkflow`.; flags: --id, --workflow
    workflow-scheme-drafts delete-workflow-scheme-draft - Delete draft workflow scheme [intent=reverse_etl availability=implemented write=delete_workflow_scheme_draft]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteWorkflowSchemeDraft`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    workflow-scheme-drafts get-workflow-scheme-draft - Get draft workflow scheme [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    workflow-scheme-drafts update-workflow-scheme-draft - Update draft workflow scheme [intent=reverse_etl availability=partial write=update_workflow_scheme_draft]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateWorkflowSchemeDraft`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --id
    workflow-scheme-drafts delete-draft-default-workflow - Delete draft default workflow [intent=reverse_etl availability=implemented write=delete_draft_default_workflow]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteDraftDefaultWorkflow`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id
    workflow-scheme-drafts get-draft-default-workflow - Get draft default workflow [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id
    workflow-scheme-drafts update-draft-default-workflow - Update draft default workflow [intent=reverse_etl availability=implemented write=update_draft_default_workflow]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateDraftDefaultWorkflow`.; flags: --id, --workflow
    workflow-scheme-drafts delete-workflow-scheme-draft-issue-type - Delete workflow for issue type in draft workflow scheme [intent=reverse_etl availability=implemented write=delete_workflow_scheme_draft_issue_type]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteWorkflowSchemeDraftIssueType`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id, --issue-type
    workflow-scheme-drafts get-workflow-scheme-draft-issue-type - Get workflow for issue type in draft workflow scheme [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id, --issue-type
    workflow-scheme-drafts set-workflow-scheme-draft-issue-type - Set workflow for issue type in draft workflow scheme [intent=reverse_etl availability=partial write=set_workflow_scheme_draft_issue_type]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `setWorkflowSchemeDraftIssueType`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --id, --issue-type
    workflow-scheme-drafts publish-draft-workflow-scheme - Publish draft workflow scheme [intent=reverse_etl availability=partial write=publish_draft_workflow_scheme]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `publishDraftWorkflowScheme`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --id
    workflow-scheme-drafts delete-draft-workflow-mapping - Delete issue types for workflow in draft workflow scheme [intent=reverse_etl availability=implemented write=delete_draft_workflow_mapping]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteDraftWorkflowMapping`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id, --workflow-name
    workflow-scheme-drafts get-draft-workflow - Get issue types for workflows in draft workflow scheme [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id, --workflow-name
    workflow-scheme-drafts update-draft-workflow-mapping - Set issue types for workflow in workflow scheme [intent=reverse_etl availability=partial write=update_draft_workflow_mapping]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateDraftWorkflowMapping`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --id, --workflow-name
    workflow-schemes delete-workflow-scheme-issue-type - Delete workflow for issue type in workflow scheme [intent=reverse_etl availability=implemented write=delete_workflow_scheme_issue_type]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteWorkflowSchemeIssueType`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id, --issue-type
    workflow-schemes get-workflow-scheme-issue-type - Get workflow for issue type in workflow scheme [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id, --issue-type, --return-draft-if-exists
    workflow-schemes set-workflow-scheme-issue-type - Set workflow for issue type in workflow scheme [intent=reverse_etl availability=partial write=set_workflow_scheme_issue_type]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `setWorkflowSchemeIssueType`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits required-body flag validation.; flags: --id, --issue-type
    workflow-schemes delete-workflow-mapping - Delete issue types for workflow in workflow scheme [intent=reverse_etl availability=implemented write=delete_workflow_mapping]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteWorkflowMapping`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --id, --workflow-name
    workflow-schemes get-workflow - Get issue types for workflows in workflow scheme [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --id, --workflow-name, --return-draft-if-exists
    workflow-schemes update-workflow-mapping - Set issue types for workflow in workflow scheme [intent=reverse_etl availability=partial write=update_workflow_mapping]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `updateWorkflowMapping`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed JSON body flag support for this required request body.; flags: --id, --workflow-name
    workflow-schemes get-project-usages-for-workflow-scheme - Get projects which are using a given workflow scheme [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --workflow-scheme-id, --next-page-token, --max-results
    issue-worklogs get-ids-of-worklogs-deleted-since - Get IDs of deleted worklogs [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --since
    issue-worklogs get-worklogs-for-ids - Get worklogs [intent=direct_read availability=partial]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; notes: Provider-style command awaits typed integer/object array body flag support before it can run safely.; flags: --ids, --expand
    issue-worklogs get-ids-of-worklogs-modified-since - Get IDs of updated worklogs [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --since, --expand
    app-properties addon-properties-resource-get-addon-properties-get - Get app properties [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --addon-key
    app-properties addon-properties-resource-delete-addon-property-delete - Delete app property [intent=reverse_etl availability=implemented write=addon_properties_resource_delete_addon_property_delete]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `AddonPropertiesResource.deleteAddonProperty_delete`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --addon-key, --property-key
    app-properties addon-properties-resource-get-addon-property-get - Get app property [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --addon-key, --property-key
    app-properties addon-properties-resource-put-addon-property-put - Set app property [intent=reverse_etl availability=planned]; notes: Provider-style command is planned until raw or dynamic JSON request-body support can represent this required body safely.
    dynamic-modules dynamic-modules-resource-remove-modules-delete - Remove modules [intent=reverse_etl availability=implemented write=dynamic_modules_resource_remove_modules_delete]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `DynamicModulesResource.removeModules_delete`.; notes: Requires --confirm destructive when executing the approved reverse plan.
    dynamic-modules dynamic-modules-resource-get-modules-get - Get modules [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    dynamic-modules dynamic-modules-resource-register-modules-post - Register modules [intent=reverse_etl availability=partial write=dynamic_modules_resource_register_modules_post]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `DynamicModulesResource.registerModules_post`.; notes: Reverse ETL action is implemented for record-driven plans; provider-style one-off command awaits typed integer/object array flag support.; flags: --modules
    app-migration app-issue-field-value-update-resource-update-issue-fields-put - Bulk update custom field value [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `AppIssueFieldValueUpdateResource.updateIssueFields_put`.; notes: Blocked until declarative writes support required per-action Atlassian-Transfer-Id request headers.
    app-migration migration-resource-update-entity-properties-value-put - Bulk update entity properties [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `MigrationResource.updateEntityPropertiesValue_put`.; notes: Blocked until declarative writes support required per-action Atlassian-Transfer-Id request headers.; flags: --entity-type, --items
    app-migration migration-resource-workflow-rule-search-post - Get workflow transition rule configurations [intent=reverse_etl availability=planned]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `MigrationResource.workflowRuleSearch_post`.; notes: Blocked until declarative writes support required per-action Atlassian-Transfer-Id request headers.; flags: --rule-ids, --workflow-entity-id
    migration-of-connect-modules-to-forge connect-to-forge-migration-fetch-task-resource-fetch-migration-task-get - Get Connect issue field migration task [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --connect-key, --jira-issue-fields-key
    migration-of-connect-modules-to-forge connect-to-forge-migration-task-submission-resource-submit-task-post - Submit Connect issue field migration task [intent=reverse_etl availability=implemented write=connect_to_forge_migration_task_submission_resource_submit_task_post]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Mutates Jira data/configuration for operation `ConnectToForgeMigrationTaskSubmissionResource.submitTask_post`.; flags: --connect-key, --jira-issue-fields-key
    service-registry service-registry-resource-services-get - Retrieve the attributes of service registries [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --service-ids
    app-properties get-forge-app-property-keys - Get app property keys (Forge) [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.
    app-properties delete-forge-app-property - Delete app property (Forge) [intent=reverse_etl availability=implemented write=delete_forge_app_property]; approval: reverse ETL writes require plan, preview, explicit approval, and execute; destructive actions also require typed `destructive` confirmation.; risk: Permanently deletes or removes Jira data/configuration for operation `deleteForgeAppProperty`.; notes: Requires --confirm destructive when executing the approved reverse plan.; flags: --property-key
    app-properties get-forge-app-property - Get app property (Forge) [intent=direct_read availability=implemented]; approval: none; risk: Bounded JSON direct read; response is capped and redacted by policy.; flags: --property-key
    app-properties put-forge-app-property - Set app property (Forge) [intent=reverse_etl availability=planned]; notes: Provider-style command is planned until raw or dynamic JSON request-body support can represent this required body safely.
    api get-worklogs-by-issue-id-and-worklog-id - Get worklogs by issue id and worklog id [intent=direct_read availability=partial]; approval: none; risk: Bounded JSON provider read; response is capped and redacted by policy.; notes: Provider-style command awaits typed integer/object array body flag support before it can run safely.; flags: --requests
  Help topics:
    jira - Jira connector commands are generated from the official Jira Cloud REST API v3 ledger.
    jira-writes - Jira write commands create reverse ETL plans and require explicit approval; destructive writes require typed confirmation.
    jira-direct-read - Jira direct reads are fixed endpoint commands with bounded JSON responses and no raw API passthrough.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect jira

  # Inspect as structured JSON
  pm connectors inspect jira --json

AGENT WORKFLOW
  - Run pm connectors inspect jira before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
