# Asana source-cited mutation gap inventory

Captain-directed inventory recorded 2026-08-26 from the pinned Asana descriptor:
`internal/connectors/defs/asana/sources/asana-operation-descriptor.json`.

The command below returned 90 source-projection findings:

```sh
go run ./cmd/connectorgen validate internal/connectors/defs
```

Every source operation below is bound to the pinned OpenAPI capture at
`56796a67a3c093eedf55fd9682357957a2ebfd85` with SHA-256
`cb3b90f4e0af56035eab0c648974f625b942a28a7144aa6c2326e38ca0bb3d56`.

## Lane A — 25 source routes without a complete source-mapped action

Twenty-one of these source-cited provider mutations have no action or command.
They use the existing `source-cited-non-executable-mutation-foundation-r1`
disposition. Four have a working legacy delete command but its local `gid`
path field does not name the provider's exact path parameter; those four use
the explicit `source-path-parameter-alias-foundation-r1` partial-coverage
disposition instead of being falsely called absent.

| Source operation | Exact source endpoint |
| --- | --- |
| `asana.rest.deleteAllocation` | `DELETE /allocations/{allocation_gid}` |
| `asana.rest.deleteAttachment` | `DELETE /attachments/{attachment_gid}` |
| `asana.rest.deleteBudget` | `DELETE /budgets/{budget_gid}` |
| `asana.rest.deleteCustomField` | `DELETE /custom_fields/{custom_field_gid}` |
| `asana.rest.deleteGoal` | `DELETE /goals/{goal_gid}` |
| `asana.rest.deleteMembership` | `DELETE /memberships/{membership_gid}` |
| `asana.rest.deleteOooEntry` | `DELETE /ooo_entries/{ooo_entry_gid}` |
| `asana.rest.deletePortfolio` | `DELETE /portfolios/{portfolio_gid}` |
| `asana.rest.deleteProjectBrief` | `DELETE /project_briefs/{project_brief_gid}` |
| `asana.rest.deleteProjectStatus` | `DELETE /project_statuses/{project_status_gid}` |
| `asana.rest.deleteProjectTemplate` | `DELETE /project_templates/{project_template_gid}` |
| `asana.rest.deleteProject` | `DELETE /projects/{project_gid}` — legacy `projects delete` path-alias mapping |
| `asana.rest.deleteRate` | `DELETE /rates/{rate_gid}` |
| `asana.rest.deleteRole` | `DELETE /roles/{role_gid}` |
| `asana.rest.deleteSection` | `DELETE /sections/{section_gid}` — legacy `sections delete` path-alias mapping |
| `asana.rest.deleteStatus` | `DELETE /status_updates/{status_update_gid}` |
| `asana.rest.deleteStory` | `DELETE /stories/{story_gid}` |
| `asana.rest.deleteTag` | `DELETE /tags/{tag_gid}` — legacy `tags delete` path-alias mapping |
| `asana.rest.deleteTaskTemplate` | `DELETE /task_templates/{task_template_gid}` |
| `asana.rest.deleteTask` | `DELETE /tasks/{task_gid}` — legacy `tasks delete` path-alias mapping |
| `asana.rest.deleteTimeTrackingCategory` | `DELETE /time_tracking_categories/{time_tracking_category_gid}` |
| `asana.rest.deleteTimeTrackingEntry` | `DELETE /time_tracking_entries/{time_tracking_entry_gid}` |
| `asana.rest.deleteWebhook` | `DELETE /webhooks/{webhook_gid}` |
| `asana.rest.approveAccessRequest` | `POST /access_requests/{access_request_gid}/approve` |
| `asana.rest.rejectAccessRequest` | `POST /access_requests/{access_request_gid}/reject` |

## Lane B — 65 implemented reverse-ETL commands with incomplete source contracts

These operations already retain a provider-derived
`cli-request-schema-foundation-r1` gap: their source contracts contain a
dynamic/complex body and, for many, a non-scalar query parameter. The current
CLI action is still marked `implemented`; no command was changed by this
inventory.

| Source operation | Exact source endpoint |
| --- | --- |
| `asana.rest.createCustomField` | `POST /custom_fields` |
| `asana.rest.createEnumOptionForCustomField` | `POST /custom_fields/{custom_field_gid}/enum_options` |
| `asana.rest.insertEnumOptionForCustomField` | `POST /custom_fields/{custom_field_gid}/enum_options/insert` |
| `asana.rest.createGraphExport` | `POST /exports/graph` |
| `asana.rest.createResourceExport` | `POST /exports/resource` |
| `asana.rest.createGoal` | `POST /goals` |
| `asana.rest.addCustomFieldSettingForGoal` | `POST /goals/{goal_gid}/addCustomFieldSetting` |
| `asana.rest.addFollowers` | `POST /goals/{goal_gid}/addFollowers` |
| `asana.rest.addSupportingRelationship` | `POST /goals/{goal_gid}/addSupportingRelationship` |
| `asana.rest.createGoalMetric` | `POST /goals/{goal_gid}/setMetric` |
| `asana.rest.updateGoalMetric` | `POST /goals/{goal_gid}/setMetricCurrentValue` |
| `asana.rest.createStoryForGoal` | `POST /goals/{goal_gid}/stories` |
| `asana.rest.createOooEntry` | `POST /ooo_entries` |
| `asana.rest.createPortfolio` | `POST /portfolios` |
| `asana.rest.addCustomFieldSettingForPortfolio` | `POST /portfolios/{portfolio_gid}/addCustomFieldSetting` |
| `asana.rest.addItemForPortfolio` | `POST /portfolios/{portfolio_gid}/addItem` |
| `asana.rest.addMembersForPortfolio` | `POST /portfolios/{portfolio_gid}/addMembers` |
| `asana.rest.duplicatePortfolio` | `POST /portfolios/{portfolio_gid}/duplicate` |
| `asana.rest.instantiateProject` | `POST /project_templates/{project_template_gid}/instantiateProject` |
| `asana.rest.createProject` | `POST /projects` |
| `asana.rest.addCustomFieldSettingForProject` | `POST /projects/{project_gid}/addCustomFieldSetting` |
| `asana.rest.addFollowersForProject` | `POST /projects/{project_gid}/addFollowers` |
| `asana.rest.addMembersForProject` | `POST /projects/{project_gid}/addMembers` |
| `asana.rest.duplicateProject` | `POST /projects/{project_gid}/duplicate` |
| `asana.rest.createProjectBrief` | `POST /projects/{project_gid}/project_briefs` |
| `asana.rest.createProjectStatusForProject` | `POST /projects/{project_gid}/project_statuses` |
| `asana.rest.projectSaveAsTemplate` | `POST /projects/{project_gid}/saveAsTemplate` |
| `asana.rest.createSectionForProject` | `POST /projects/{project_gid}/sections` |
| `asana.rest.insertSectionForProject` | `POST /projects/{project_gid}/sections/insert` |
| `asana.rest.triggerRule` | `POST /rule_triggers/{rule_trigger_gid}/run` |
| `asana.rest.addTaskForSection` | `POST /sections/{section_gid}/addTask` |
| `asana.rest.createStatusForObject` | `POST /status_updates` |
| `asana.rest.createTag` | `POST /tags` |
| `asana.rest.instantiateTask` | `POST /task_templates/{task_template_gid}/instantiateTask` |
| `asana.rest.createTask` | `POST /tasks` |
| `asana.rest.addDependenciesForTask` | `POST /tasks/{task_gid}/addDependencies` |
| `asana.rest.addDependentsForTask` | `POST /tasks/{task_gid}/addDependents` |
| `asana.rest.addFollowersForTask` | `POST /tasks/{task_gid}/addFollowers` |
| `asana.rest.addProjectForTask` | `POST /tasks/{task_gid}/addProject` |
| `asana.rest.addTagForTask` | `POST /tasks/{task_gid}/addTag` |
| `asana.rest.duplicateTask` | `POST /tasks/{task_gid}/duplicate` |
| `asana.rest.setParentForTask` | `POST /tasks/{task_gid}/setParent` |
| `asana.rest.createStoryForTask` | `POST /tasks/{task_gid}/stories` |
| `asana.rest.createSubtaskForTask` | `POST /tasks/{task_gid}/subtasks` |
| `asana.rest.createTimeTrackingEntry` | `POST /tasks/{task_gid}/time_tracking_entries` |
| `asana.rest.createTeam` | `POST /teams` |
| `asana.rest.addUserForTeam` | `POST /teams/{team_gid}/addUser` |
| `asana.rest.createProjectForTeam` | `POST /teams/{team_gid}/projects` |
| `asana.rest.addUserForWorkspace` | `POST /workspaces/{workspace_gid}/addUser` |
| `asana.rest.createProjectForWorkspace` | `POST /workspaces/{workspace_gid}/projects` |
| `asana.rest.createTagForWorkspace` | `POST /workspaces/{workspace_gid}/tags` |
| `asana.rest.updateCustomField` | `PUT /custom_fields/{custom_field_gid}` |
| `asana.rest.updateEnumOption` | `PUT /enum_options/{enum_option_gid}` |
| `asana.rest.updateGoalRelationship` | `PUT /goal_relationships/{goal_relationship_gid}` |
| `asana.rest.updateGoal` | `PUT /goals/{goal_gid}` |
| `asana.rest.updateOooEntry` | `PUT /ooo_entries/{ooo_entry_gid}` |
| `asana.rest.updatePortfolio` | `PUT /portfolios/{portfolio_gid}` |
| `asana.rest.updateProjectBrief` | `PUT /project_briefs/{project_brief_gid}` |
| `asana.rest.updateProjectPortfolioSetting` | `PUT /project_portfolio_settings/{project_portfolio_setting_gid}` |
| `asana.rest.updateStory` | `PUT /stories/{story_gid}` |
| `asana.rest.updateTeam` | `PUT /teams/{team_gid}` |
| `asana.rest.updateTimeTrackingEntry` | `PUT /time_tracking_entries/{time_tracking_entry_gid}` |
| `asana.rest.updateUser` | `PUT /users/{user_gid}` |
| `asana.rest.updateWorkspace` | `PUT /workspaces/{workspace_gid}` |
| `asana.rest.updateUserForWorkspace` | `PUT /workspaces/{workspace_gid}/users/{user_gid}` |

## Resolution

The connector-owned `asana-mutation-dispositions.json` now retains an exact
source ID, method, and path for all 90 rows. Existing
`sourceNonExecutableMutationDisposition` remains limited to the 21 genuinely
absent actions. A new, distinct `sourcePartialMutationCoverageDisposition`
preserves the 65 typed-request-incomplete actions and four legacy path aliases
while recording their named missing foundations. The validator rejects that
partial form for a source-complete action, an absent action, a read, or a route
without an implemented declared command; provider-evidenced unsupported is not
represented by either missing-foundation disposition.
