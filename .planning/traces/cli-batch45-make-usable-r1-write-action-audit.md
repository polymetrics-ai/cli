# Verified-source write-action audit — Phase 2, 2026-08-24

## Scope and correction

This audit answers the follow-up for Pipedrive, Squarespace, and Zendesk Support.
The prior source-contract probe established that their verified pinned source does
not admit an `idempotency_key_header` for any existing `writes.json` action. That
rules each action out as a `declarative_typed_destination`; it does **not**, on
its own, establish that the existing write action contradicts its provider API.

For every action below, the declaration's effective method and path were compared
against the byte-for-byte verified source lock operation. `{{ record.<name> }}`
was normalized to one provider path-variable position. For Squarespace, the
declaration path is relative to its pinned `spec.json` default
`https://api.squarespace.com/1.0`, so `/webhook_subscriptions` is compared as
`/1.0/webhook_subscriptions`. This check is limited to operation identity; it is
not a substitute for a complete provider request-body conformance audit.

Result: **109/109 source operation identities match.** There is no observed
method/path contradiction in the merged `writes.json` contracts. The sources
also do not declare a usable idempotency header for these actions:

- Pipedrive's verified full OpenAPI source contains no `idempotency` declaration.
- Zendesk Support's verified full OpenAPI source contains no `idempotency`
  declaration.
- Squarespace's verified full OpenAPI source declares `Idempotency-Key` only for
  `POST /1.0/commerce/inventory/adjustments` and
  `POST /1.0/commerce/orders`, neither of which is an existing action here.

The engine can make a one-shot source-matched request without a header, but
disables retries for a non-idempotent action (`engine/write.go:writeRequester`).
Therefore these actions can be provider-method/path-correct yet remain ineligible
for durable typed-destination delivery. This does not certify every existing CLI
row as binary-reachable; the Phase 1 binary evidence remains the only such proof.

## Pipedrive — 17 matching actions

Pinned source: `https://developers.pipedrive.com/docs/api/v1/openapi.yaml`,
SHA-256 `302b0d7c2c1a6cb96a2d299717c6be0c2cf3eac6dfd884ea8352962ebf501c2b`.

| Declared action | Declaration claims | Pinned source states |
| --- | --- | --- |
| `create_lead` | `POST /leads` | `POST /leads` (`addLead`) |
| `update_lead` | `PATCH /leads/{id}` | `PATCH /leads/{id}` (`updateLead`) |
| `delete_lead` | `DELETE /leads/{id}` | `DELETE /leads/{id}` (`deleteLead`) |
| `create_note` | `POST /notes` | `POST /notes` (`addNote`) |
| `update_note` | `PUT /notes/{id}` | `PUT /notes/{id}` (`updateNote`) |
| `delete_note` | `DELETE /notes/{id}` | `DELETE /notes/{id}` (`deleteNote`) |
| `create_filter` | `POST /filters` | `POST /filters` (`addFilter`) |
| `update_filter` | `PUT /filters/{id}` | `PUT /filters/{id}` (`updateFilter`) |
| `delete_filter` | `DELETE /filters/{id}` | `DELETE /filters/{id}` (`deleteFilter`) |
| `create_activity_type` | `POST /activityTypes` | `POST /activityTypes` (`addActivityType`) |
| `update_activity_type` | `PUT /activityTypes/{id}` | `PUT /activityTypes/{id}` (`updateActivityType`) |
| `delete_activity_type` | `DELETE /activityTypes/{id}` | `DELETE /activityTypes/{id}` (`deleteActivityType`) |
| `create_lead_label` | `POST /leadLabels` | `POST /leadLabels` (`addLeadLabel`) |
| `update_lead_label` | `PATCH /leadLabels/{id}` | `PATCH /leadLabels/{id}` (`updateLeadLabel`) |
| `delete_lead_label` | `DELETE /leadLabels/{id}` | `DELETE /leadLabels/{id}` (`deleteLeadLabel`) |
| `create_webhook` | `POST /webhooks` | `POST /webhooks` (`addWebhook`) |
| `delete_webhook` | `DELETE /webhooks/{id}` | `DELETE /webhooks/{id}` (`deleteWebhook`) |

## Squarespace — 2 matching actions

Pinned source:
`https://developers.squarespace.com/commerce-apis/latest/schema-processor-version-version-latest.json`,
SHA-256 `eff1274e6e87cfa998a5125c2ebf53ee459202d108598dacf6507b32b2b2debc`.

| Declared action | Declaration claims (effective path) | Pinned source states |
| --- | --- | --- |
| `create_webhook_subscription` | `POST /1.0/webhook_subscriptions` | `POST /1.0/webhook_subscriptions` (`createWebhookSubscription`) |
| `delete_webhook_subscription` | `DELETE /1.0/webhook_subscriptions/{id}` | `DELETE /1.0/webhook_subscriptions/{subscriptionId}` (`deleteWebhookSubscription`) |

## Zendesk Support — 90 matching actions

Pinned source: `https://developer.zendesk.com/zendesk/oas.yaml`, SHA-256
`a487892c8e1f3feeba96c234148be69fddd50afce17bf30437bcb8de36d9a0c8`.

| Declared action | Declaration claims | Pinned source states |
| --- | --- | --- |
| `create_ticket` | `POST /api/v2/tickets` | `POST /api/v2/tickets` (`CreateTicket`) |
| `update_ticket` | `PUT /api/v2/tickets/{id}` | `PUT /api/v2/tickets/{ticket_id}` (`UpdateTicket`) |
| `delete_ticket` | `DELETE /api/v2/tickets/{id}` | `DELETE /api/v2/tickets/{ticket_id}` (`DeleteTicket`) |
| `create_user` | `POST /api/v2/users` | `POST /api/v2/users` (`CreateUser`) |
| `update_user` | `PUT /api/v2/users/{id}` | `PUT /api/v2/users/{user_id}` (`UpdateUser`) |
| `delete_user` | `DELETE /api/v2/users/{id}` | `DELETE /api/v2/users/{user_id}` (`DeleteUser`) |
| `create_organization` | `POST /api/v2/organizations` | `POST /api/v2/organizations` (`CreateOrganization`) |
| `update_organization` | `PUT /api/v2/organizations/{id}` | `PUT /api/v2/organizations/{organization_id}` (`UpdateOrganization`) |
| `delete_organization` | `DELETE /api/v2/organizations/{id}` | `DELETE /api/v2/organizations/{organization_id}` (`DeleteOrganization`) |
| `create_group` | `POST /api/v2/groups` | `POST /api/v2/groups` (`CreateGroup`) |
| `update_group` | `PUT /api/v2/groups/{id}` | `PUT /api/v2/groups/{group_id}` (`UpdateGroup`) |
| `delete_group` | `DELETE /api/v2/groups/{id}` | `DELETE /api/v2/groups/{group_id}` (`DeleteGroup`) |
| `create_macro` | `POST /api/v2/macros` | `POST /api/v2/macros` (`CreateMacro`) |
| `update_macro` | `PUT /api/v2/macros/{id}` | `PUT /api/v2/macros/{macro_id}` (`UpdateMacro`) |
| `delete_macro` | `DELETE /api/v2/macros/{id}` | `DELETE /api/v2/macros/{macro_id}` (`DeleteMacro`) |
| `create_trigger` | `POST /api/v2/triggers` | `POST /api/v2/triggers` (`CreateTrigger`) |
| `update_trigger` | `PUT /api/v2/triggers/{id}` | `PUT /api/v2/triggers/{trigger_id}` (`UpdateTrigger`) |
| `delete_trigger` | `DELETE /api/v2/triggers/{id}` | `DELETE /api/v2/triggers/{trigger_id}` (`DeleteTrigger`) |
| `create_automation` | `POST /api/v2/automations` | `POST /api/v2/automations` (`CreateAutomation`) |
| `update_automation` | `PUT /api/v2/automations/{id}` | `PUT /api/v2/automations/{automation_id}` (`UpdateAutomation`) |
| `delete_automation` | `DELETE /api/v2/automations/{id}` | `DELETE /api/v2/automations/{automation_id}` (`DeleteAutomation`) |
| `delete_api_token` | `DELETE /api/v2/api_tokens/{id}` | `DELETE /api/v2/api_tokens/{id}` (`DeleteApiToken`) |
| `create_view` | `POST /api/v2/views` | `POST /api/v2/views` (`CreateView`) |
| `update_view` | `PUT /api/v2/views/{id}` | `PUT /api/v2/views/{view_id}` (`UpdateView`) |
| `delete_view` | `DELETE /api/v2/views/{id}` | `DELETE /api/v2/views/{view_id}` (`DeleteView`) |
| `create_ticket_field` | `POST /api/v2/ticket_fields` | `POST /api/v2/ticket_fields` (`CreateTicketField`) |
| `update_ticket_field` | `PUT /api/v2/ticket_fields/{id}` | `PUT /api/v2/ticket_fields/{ticket_field_id}` (`UpdateTicketField`) |
| `delete_ticket_field` | `DELETE /api/v2/ticket_fields/{id}` | `DELETE /api/v2/ticket_fields/{ticket_field_id}` (`DeleteTicketField`) |
| `update_account_email_settings` | `PUT /api/v2/account/email_settings` | `PUT /api/v2/account/email_settings` (`UpdateAccountEmailSettings`) |
| `create_approval_request` | `POST /api/v2/approval_requests` | `POST /api/v2/approval_requests` (`CreateApprovalRequest`) |
| `update_attachment` | `PUT /api/v2/attachments/{attachment_id}` | `PUT /api/v2/attachments/{attachment_id}` (`UpdateAttachment`) |
| `create_bookmark` | `POST /api/v2/bookmarks` | `POST /api/v2/bookmarks` (`CreateBookmark`) |
| `create_brand` | `POST /api/v2/brands` | `POST /api/v2/brands` (`CreateBrand`) |
| `update_brand` | `PUT /api/v2/brands/{brand_id}` | `PUT /api/v2/brands/{brand_id}` (`UpdateBrand`) |
| `create_ticket_or_voicemail_ticket` | `POST /api/v2/channels/voice/tickets` | `POST /api/v2/channels/voice/tickets` (`CreateTicketOrVoicemailTicket`) |
| `create_custom_object` | `POST /api/v2/custom_objects` | `POST /api/v2/custom_objects` (`CreateCustomObject`) |
| `create_access_rule` | `POST /api/v2/custom_objects/{custom_object_key}/access_rules` | `POST /api/v2/custom_objects/{custom_object_key}/access_rules` (`CreateAccessRule`) |
| `update_access_rule` | `PATCH /api/v2/custom_objects/{custom_object_key}/access_rules/{id}` | `PATCH /api/v2/custom_objects/{custom_object_key}/access_rules/{id}` (`UpdateAccessRule`) |
| `create_custom_object_field` | `POST /api/v2/custom_objects/{custom_object_key}/fields` | `POST /api/v2/custom_objects/{custom_object_key}/fields` (`CreateCustomObjectField`) |
| `custom_object_record_bulk_jobs` | `POST /api/v2/custom_objects/{custom_object_key}/jobs` | `POST /api/v2/custom_objects/{custom_object_key}/jobs` (`CustomObjectRecordBulkJobs`) |
| `create_custom_object_record` | `POST /api/v2/custom_objects/{custom_object_key}/records` | `POST /api/v2/custom_objects/{custom_object_key}/records` (`CreateCustomObjectRecord`) |
| `upsert_custom_object_record_by_external_id_or_name` | `PATCH /api/v2/custom_objects/{custom_object_key}/records` | `PATCH /api/v2/custom_objects/{custom_object_key}/records` (`UpsertCustomObjectRecordByExternalIdOrName`) |
| `update_custom_object_record_attachment` | `PUT /api/v2/custom_objects/{custom_object_key}/records/{record_id}/attachments/{id}` | `PUT /api/v2/custom_objects/{custom_object_key}/records/{record_id}/attachments/{id}` (`UpdateCustomObjectRecordAttachment`) |
| `create_object_trigger` | `POST /api/v2/custom_objects/{custom_object_key}/triggers` | `POST /api/v2/custom_objects/{custom_object_key}/triggers` (`CreateObjectTrigger`) |
| `update_object_trigger` | `PUT /api/v2/custom_objects/{custom_object_key}/triggers/{trigger_id}` | `PUT /api/v2/custom_objects/{custom_object_key}/triggers/{trigger_id}` (`UpdateObjectTrigger`) |
| `update_many_object_triggers` | `PUT /api/v2/custom_objects/{custom_object_key}/triggers/update_many` | `PUT /api/v2/custom_objects/{custom_object_key}/triggers/update_many` (`UpdateManyObjectTriggers`) |
| `bulk_update_default_custom_status` | `PUT /api/v2/custom_status/default` | `PUT /api/v2/custom_status/default` (`BulkUpdateDefaultCustomStatus`) |
| `create_custom_status` | `POST /api/v2/custom_statuses` | `POST /api/v2/custom_statuses` (`CreateCustomStatus`) |
| `update_custom_status` | `PUT /api/v2/custom_statuses/{custom_status_id}` | `PUT /api/v2/custom_statuses/{custom_status_id}` (`UpdateCustomStatus`) |
| `create_ticket_form_statuses_for_custom_status` | `POST /api/v2/custom_statuses/{custom_status_id}/ticket_form_statuses` | `POST /api/v2/custom_statuses/{custom_status_id}/ticket_form_statuses` (`CreateTicketFormStatusesForCustomStatus`) |
| `create_deletion_schedule` | `POST /api/v2/deletion_schedules` | `POST /api/v2/deletion_schedules` (`CreateDeletionSchedule`) |
| `update_deletion_schedule` | `PUT /api/v2/deletion_schedules/{deletion_schedule_id}` | `PUT /api/v2/deletion_schedules/{deletion_schedule_id}` (`UpdateDeletionSchedule`) |
| `create_group_membership` | `POST /api/v2/group_memberships` | `POST /api/v2/group_memberships` (`CreateGroupMembership`) |
| `group_membership_bulk_create` | `POST /api/v2/group_memberships/create_many` | `POST /api/v2/group_memberships/create_many` (`GroupMembershipBulkCreate`) |
| `ticket_import` | `POST /api/v2/imports/tickets` | `POST /api/v2/imports/tickets` (`TicketImport`) |
| `ticket_bulk_import` | `POST /api/v2/imports/tickets/create_many` | `POST /api/v2/imports/tickets/create_many` (`TicketBulkImport`) |
| `create_itam_asset_type` | `POST /api/v2/it_asset_management/asset_types` | `POST /api/v2/it_asset_management/asset_types` (`CreateItamAssetType`) |
| `create_itam_asset_type_field` | `POST /api/v2/it_asset_management/asset_types/{asset_type_id}/fields` | `POST /api/v2/it_asset_management/asset_types/{asset_type_id}/fields` (`CreateItamAssetTypeField`) |
| `create_itam_asset` | `POST /api/v2/it_asset_management/assets` | `POST /api/v2/it_asset_management/assets` (`CreateItamAsset`) |
| `itam_asset_bulk_jobs` | `POST /api/v2/it_asset_management/assets/jobs` | `POST /api/v2/it_asset_management/assets/jobs` (`ItamAssetBulkJobs`) |
| `create_itam_location` | `POST /api/v2/it_asset_management/locations` | `POST /api/v2/it_asset_management/locations` (`CreateItamLocation`) |
| `create_itam_status` | `POST /api/v2/it_asset_management/statuses` | `POST /api/v2/it_asset_management/statuses` (`CreateItamStatus`) |
| `update_itam_status` | `PATCH /api/v2/it_asset_management/statuses/{status_id}` | `PATCH /api/v2/it_asset_management/statuses/{status_id}` (`UpdateItamStatus`) |
| `update_many_macros` | `PUT /api/v2/macros/update_many` | `PUT /api/v2/macros/update_many` (`UpdateManyMacros`) |
| `create_organization_subscription` | `POST /api/v2/organization_subscriptions` | `POST /api/v2/organization_subscriptions` (`CreateOrganizationSubscription`) |
| `create_saved_search` | `POST /api/v2/saved_searches` | `POST /api/v2/saved_searches` (`CreateSavedSearch`) |
| `update_saved_search` | `PUT /api/v2/saved_searches/{id}` | `PUT /api/v2/saved_searches/{id}` (`UpdateSavedSearch`) |
| `bulk_recover_suspended_tickets` | `PUT /api/v2/suspended_tickets/bulk_recover` | `PUT /api/v2/suspended_tickets/bulk_recover` (`BulkRecoverSuspendedTickets`) |
| `create_task_list_template` | `POST /api/v2/task_list_templates` | `POST /api/v2/task_list_templates` (`CreateTaskListTemplate`) |
| `update_task_list_template` | `PUT /api/v2/task_list_templates/{task_list_template_id}` | `PUT /api/v2/task_list_templates/{task_list_template_id}` (`UpdateTaskListTemplate`) |
| `create_ticket_content_pin` | `POST /api/v2/ticket_content_pins` | `POST /api/v2/ticket_content_pins` (`CreateTicketContentPin`) |
| `create_ticket_form_statuses` | `POST /api/v2/ticket_forms/{ticket_form_id}/ticket_form_statuses` | `POST /api/v2/ticket_forms/{ticket_form_id}/ticket_form_statuses` (`CreateTicketFormStatuses`) |
| `update_ticket_form_statuses` | `PUT /api/v2/ticket_forms/{ticket_form_id}/ticket_form_statuses` | `PUT /api/v2/ticket_forms/{ticket_form_id}/ticket_form_statuses` (`UpdateTicketFormStatuses`) |
| `update_ticket_form_status_by_id` | `PUT /api/v2/ticket_forms/{ticket_form_id}/ticket_form_statuses/{ticket_form_status_id}` | `PUT /api/v2/ticket_forms/{ticket_form_id}/ticket_form_statuses/{ticket_form_status_id}` (`UpdateTicketFormStatusById`) |
| `create_task_list` | `POST /api/v2/tickets/{ticket_id}/task_lists` | `POST /api/v2/tickets/{ticket_id}/task_lists` (`CreateTaskList`) |
| `tickets_create_many` | `POST /api/v2/tickets/create_many` | `POST /api/v2/tickets/create_many` (`TicketsCreateMany`) |
| `create_trigger_category` | `POST /api/v2/trigger_categories` | `POST /api/v2/trigger_categories` (`CreateTriggerCategory`) |
| `update_trigger_category` | `PATCH /api/v2/trigger_categories/{trigger_category_id}` | `PATCH /api/v2/trigger_categories/{trigger_category_id}` (`UpdateTriggerCategory`) |
| `batch_operate_trigger_categories` | `POST /api/v2/trigger_categories/jobs` | `POST /api/v2/trigger_categories/jobs` (`BatchOperateTriggerCategories`) |
| `update_many_triggers` | `PUT /api/v2/triggers/update_many` | `PUT /api/v2/triggers/update_many` (`UpdateManyTriggers`) |
| `create_user_group_membership` | `POST /api/v2/users/{user_id}/group_memberships` | `POST /api/v2/users/{user_id}/group_memberships` (`CreateUserGroupMembership`) |
| `update_current_user_settings` | `PUT /api/v2/users/me/settings` | `PUT /api/v2/users/me/settings` (`UpdateCurrentUserSettings`) |
| `create_workspace` | `POST /api/v2/workspaces` | `POST /api/v2/workspaces` (`CreateWorkspace`) |
| `update_workspace` | `PUT /api/v2/workspaces/{workspace_id}` | `PUT /api/v2/workspaces/{workspace_id}` (`UpdateWorkspace`) |
| `reorder_workspaces` | `PUT /api/v2/workspaces/reorder` | `PUT /api/v2/workspaces/reorder` (`ReorderWorkspaces`) |
| `update_permission_policy` | `PATCH /api/v2/custom_objects/{custom_object_key}/permission_policies/{id}` | `PATCH /api/v2/custom_objects/{custom_object_key}/permission_policies/{id}` (`UpdatePermissionPolicy`) |
| `bulk_set_agent_attribute_values_job` | `POST /api/v2/routing/agents/instance_values/jobs` | `POST /api/v2/routing/agents/instance_values/jobs` (`BulkSetAgentAttributeValuesJob`) |
| `create_many_users` | `POST /api/v2/users/create_many` | `POST /api/v2/users/create_many` (`CreateManyUsers`) |
| `create_or_update_many_users` | `POST /api/v2/users/create_or_update_many` | `POST /api/v2/users/create_or_update_many` (`CreateOrUpdateManyUsers`) |
| `request_user_create` | `POST /api/v2/users/request_create` | `POST /api/v2/users/request_create` (`RequestUserCreate`) |
