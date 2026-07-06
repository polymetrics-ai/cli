# Overview

Reads and writes Smartsheet sheets, rows, folders, reports, dashboards, users, webhooks,
attachments, discussions, proofs, update requests, and workspace metadata.

Readable streams: `sheets`, `sheet_rows`, `contacts`, `contact`, `events`, `favorites`, `favorite`,
`folder_metadata`, `folder_children`, `folder_path`, `home_contents`, `groups`, `group`,
`home_folders`, `reports`, `report`, `report_path`, `report_publish`, `asset_shares`, `asset_share`,
`sheet`, `sheet_attachments`, `sheet_attachment`, `sheet_attachment_versions`,
`sheet_automation_rules`, `sheet_automation_rule`, `sheet_columns`, `sheet_column`, `sheet_comment`,
`sheet_cross_sheet_references`, `sheet_cross_sheet_reference`, `sheet_discussions`,
`sheet_discussion`, `sheet_discussion_attachments`, `sheet_path`, `sheet_proofs`, `sheet_proof`,
`sheet_proof_attachments`, `sheet_proof_discussions`, `sheet_proof_request_actions`,
`sheet_proof_versions`, `sheet_publish`, `sheet_row`, `sheet_row_attachments`, `sheet_cell_history`,
`sheet_row_discussions`, `sheet_sent_update_requests`, `sheet_sent_update_request`, `sheet_summary`,
`sheet_summary_fields`, `sheet_update_requests`, `sheet_update_request`, `sheet_version`,
`dashboards`, `dashboard`, `dashboard_path`, `dashboard_publish`, `users`, `current_user`,
`org_sheets`, `user`, `user_alternate_emails`, `user_alternate_email`, `user_plans`, `webhooks`,
`webhook`, `workspaces`, `workspace_metadata`, `workspace_children`.

Write actions: `add_favorites`, `delete_favorite`, `update_folder`, `copy_folder`,
`create_folder_in_folder`, `move_folder`, `create_sheet_in_folder`, `create_group`, `update_group`,
`add_group_members`, `delete_group_member`, `create_home_folder`, `create_report`,
`add_report_columns`, `update_report_definition`, `set_report_publish`, `add_report_scope`,
`remove_report_scope`, `create_sheet`, `update_sheet`, `attach_url_to_sheet`,
`delete_sheet_attachment`, `delete_sheet_attachment_versions`, `update_automation_rule`,
`delete_automation_rule`, `add_sheet_columns`, `update_sheet_column`, `delete_sheet_column`,
`update_comment`, `delete_comment`, `attach_url_to_comment`, `copy_sheet`,
`create_cross_sheet_reference`, `create_sheet_discussion`, `delete_sheet_discussion`,
`create_discussion_comment`, `move_sheet`, `update_proof`, `delete_proof`,
`create_proof_discussion`, `create_proof_request`, `delete_proof_requests`, `create_proof_version`,
`delete_proof_version`, `set_sheet_publish`, `add_sheet_rows`, `update_sheet_rows`, `copy_rows`,
`move_rows`, `attach_url_to_row`, `create_row_discussion`, `delete_sent_update_request`,
`add_summary_fields`, `update_summary_fields`, `create_update_request`, `delete_update_request`,
`sort_sheet_rows`, `update_dashboard`, `copy_dashboard`, `move_dashboard`, `set_dashboard_publish`,
`add_alternate_emails`, `delete_alternate_email`, `make_alternate_email_primary`, `create_webhook`,
`update_webhook`, `delete_webhook`, `create_workspace`, `update_workspace`, `copy_workspace`,
`create_workspace_folder`, `create_sheet_in_workspace`.

Service API documentation: https://developers.smartsheet.com/api/smartsheet/openapi/home.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Smartsheet API access token, used as the Bearer
  credential. Never logged.
- `alternate_email_id` (optional, string); Alternate email ID for alternate email streams and
  writes.
- `asset_id` (optional, string); Smartsheet asset ID used by share read streams.
- `asset_type` (optional, string); default `sheet`; Smartsheet asset type used by share read
  streams.
- `attachment_id` (optional, string); Attachment ID for attachment streams and delete/version
  actions.
- `automation_rule_id` (optional, string); Automation rule ID for automation rule streams and
  writes.
- `base_url` (optional, string); default `https://api.smartsheet.com/2.0`; format `uri`; Smartsheet
  API base URL override for tests or proxies.
- `column_id` (optional, string); Column ID for column and cell-history streams.
- `comment_id` (optional, string); Comment ID for comment streams and comment writes.
- `config_spreadsheet_id` (optional, string); Smartsheet config spreadsheet id value used by
  templated streams or writes.
- `contact_id` (optional, string); Contact ID for contact detail streams.
- `cross_sheet_reference_id` (optional, string); Cross-sheet reference ID for cross-sheet reference
  detail streams.
- `discussion_id` (optional, string); Discussion ID for discussion streams and comment writes.
- `favorite_id` (optional, string); Favorite object ID for favorite detail streams.
- `favorite_type` (optional, string); Favorite type for favorite detail streams.
- `field_id` (optional, string); Summary field ID for summary-field image writes.
- `folder_id` (optional, string); Folder ID for folder streams and folder-scoped writes.
- `group_id` (optional, string); Group ID for group streams and member writes.
- `page_size` (optional, string); default `100`; Records per page for the hook-backed sheet_rows
  stream. Declarative page-number streams use the bundle pageSize of 100.
- `plan_id` (optional, string); Plan ID for user plan membership actions.
- `proof_id` (optional, string); Proof ID for proof streams and proof-scoped writes.
- `report_id` (optional, string); Report ID for report streams and report-scoped writes.
- `row_id` (optional, string); Row ID for row streams and row-scoped writes.
- `sent_update_request_id` (optional, string); Sent update request ID for sent update request
  streams and deletes.
- `share_id` (optional, string); Share ID for share detail/update/delete actions.
- `sheet_id` (optional, string); Sheet ID for sheet streams and sheet-scoped writes.
- `sight_id` (optional, string); Dashboard (sight) ID for dashboard streams and writes.
- `spreadsheet_id` (optional, string); Smartsheet sheet or report ID to read rows from; required for
  the hook-backed sheet_rows stream.
- `update_request_id` (optional, string); Update request ID for update request streams and writes.
- `user_id` (optional, string); User ID for user and alternate email streams and writes.
- `webhook_id` (optional, string); Webhook ID for webhook streams and writes.
- `workspace_id` (optional, string); Workspace ID for workspace streams and workspace-scoped writes.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `asset_type=sheet`, `base_url=https://api.smartsheet.com/2.0`,
`page_size=100`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/sheets` with query `pageSize`=`1`.

## Streams notes

Default pagination: page-number pagination; page parameter `page`; size parameter `pageSize`; starts
at 1.

Pagination by stream: cursor: `events`, `folder_children`, `asset_shares`, `dashboards`,
`user_plans`, `workspaces`, `workspace_children`; none: `contact`, `favorite`, `folder_metadata`,
`folder_path`, `home_contents`, `group`, `reports`, `report`, `report_path`, `report_publish`,
`asset_share`, `sheet`, `sheet_attachment`, `sheet_automation_rule`, `sheet_column`,
`sheet_comment`, `sheet_cross_sheet_reference`, `sheet_discussion`, `sheet_path`, `sheet_proof`,
`sheet_publish`, `sheet_row`, `sheet_sent_update_request`, `sheet_summary`, `sheet_update_request`,
`sheet_version`, `dashboard`, `dashboard_path`, `dashboard_publish`, `current_user`, `org_sheets`,
`user`, `user_alternate_emails`, `user_alternate_email`, `webhook`, `workspace_metadata`;
page_number: `sheets`, `sheet_rows`, `contacts`, `favorites`, `groups`, `home_folders`,
`sheet_attachments`, `sheet_attachment_versions`, `sheet_automation_rules`, `sheet_columns`,
`sheet_cross_sheet_references`, `sheet_discussions`, `sheet_discussion_attachments`, `sheet_proofs`,
`sheet_proof_attachments`, `sheet_proof_discussions`, `sheet_proof_request_actions`,
`sheet_proof_versions`, `sheet_row_attachments`, `sheet_cell_history`, `sheet_row_discussions`,
`sheet_sent_update_requests`, `sheet_summary_fields`, `sheet_update_requests`, `users`, `webhooks`.

- `sheets`: GET `/sheets` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `pageSize`; starts at 1; emits passthrough records.
- `sheet_rows`: GET `/sheets/{{ config.spreadsheet_id }}` - records path `rows`; query
  `include`=`rows`; page-number pagination; page parameter `page`; size parameter `pageSize`; starts
  at 1.
- `contacts`: GET `/contacts` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `pageSize`; starts at 1.
- `contact`: GET `/contacts/{{ config.contact_id }}` - single-object response.
- `events`: GET `/events` - records path `data`; query `maxCount`=`100`; cursor pagination; cursor
  parameter `streamPosition`; next token from `nextStreamPosition`; stop flag `moreAvailable`.
- `favorites`: GET `/favorites` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `pageSize`; starts at 1.
- `favorite`: GET `/favorites/{{ config.favorite_type }}/{{ config.favorite_id }}` - single-object
  response.
- `folder_metadata`: GET `/folders/{{ config.folder_id }}/metadata` - single-object response.
- `folder_children`: GET `/folders/{{ config.folder_id }}/children` - records path `data`; query
  `maxItems`=`100`; cursor pagination; cursor parameter `lastKey`; next token from `lastKey`.
- `folder_path`: GET `/folders/{{ config.folder_id }}/path` - single-object response.
- `home_contents`: GET `/folders/personal` - single-object response.
- `groups`: GET `/groups` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `pageSize`; starts at 1.
- `group`: GET `/groups/{{ config.group_id }}` - single-object response.
- `home_folders`: GET `/home/folders` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `pageSize`; starts at 1.
- `reports`: GET `/reports` - records path `data`.
- `report`: GET `/reports/{{ config.report_id }}` - single-object response.
- `report_path`: GET `/reports/{{ config.report_id }}/path` - single-object response.
- `report_publish`: GET `/reports/{{ config.report_id }}/publish` - single-object response.
- `asset_shares`: GET `/shares` - records path `items`; query `assetId`=`{{ config.asset_id }}`;
  `assetType`=`{{ config.asset_type }}`; `maxItems`=`100`; cursor pagination; cursor parameter
  `lastKey`; next token from `lastKey`.
- `asset_share`: GET `/shares/{{ config.share_id }}` - single-object response; query `assetId`=`{{
  config.asset_id }}`; `assetType`=`{{ config.asset_type }}`.
- `sheet`: GET `/sheets/{{ config.sheet_id }}` - single-object response.
- `sheet_attachments`: GET `/sheets/{{ config.sheet_id }}/attachments` - records path `data`;
  page-number pagination; page parameter `page`; size parameter `pageSize`; starts at 1.
- `sheet_attachment`: GET `/sheets/{{ config.sheet_id }}/attachments/{{ config.attachment_id }}` -
  single-object response.
- `sheet_attachment_versions`: GET `/sheets/{{ config.sheet_id }}/attachments/{{
  config.attachment_id }}/versions` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `pageSize`; starts at 1.
- `sheet_automation_rules`: GET `/sheets/{{ config.sheet_id }}/automationrules` - records path
  `data`; page-number pagination; page parameter `page`; size parameter `pageSize`; starts at 1.
- `sheet_automation_rule`: GET `/sheets/{{ config.sheet_id }}/automationrules/{{
  config.automation_rule_id }}` - single-object response.
- `sheet_columns`: GET `/sheets/{{ config.sheet_id }}/columns` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `pageSize`; starts at 1.
- `sheet_column`: GET `/sheets/{{ config.sheet_id }}/columns/{{ config.column_id }}` - single-object
  response.
- `sheet_comment`: GET `/sheets/{{ config.sheet_id }}/comments/{{ config.comment_id }}` -
  single-object response.
- `sheet_cross_sheet_references`: GET `/sheets/{{ config.sheet_id }}/crosssheetreferences` - records
  path `data`; page-number pagination; page parameter `page`; size parameter `pageSize`; starts at
  1.
- `sheet_cross_sheet_reference`: GET `/sheets/{{ config.sheet_id }}/crosssheetreferences/{{
  config.cross_sheet_reference_id }}` - single-object response.
- `sheet_discussions`: GET `/sheets/{{ config.sheet_id }}/discussions` - records path `data`;
  page-number pagination; page parameter `page`; size parameter `pageSize`; starts at 1.
- `sheet_discussion`: GET `/sheets/{{ config.sheet_id }}/discussions/{{ config.discussion_id }}` -
  single-object response.
- `sheet_discussion_attachments`: GET `/sheets/{{ config.sheet_id }}/discussions/{{
  config.discussion_id }}/attachments` - records path `data`; page-number pagination; page parameter
  `page`; size parameter `pageSize`; starts at 1.
- `sheet_path`: GET `/sheets/{{ config.sheet_id }}/path` - single-object response.
- `sheet_proofs`: GET `/sheets/{{ config.sheet_id }}/proofs` - records path `data`; page-number
  pagination; page parameter `page`; size parameter `pageSize`; starts at 1.
- `sheet_proof`: GET `/sheets/{{ config.sheet_id }}/proofs/{{ config.proof_id }}` - single-object
  response.
- `sheet_proof_attachments`: GET `/sheets/{{ config.sheet_id }}/proofs/{{ config.proof_id
  }}/attachments` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `pageSize`; starts at 1.
- `sheet_proof_discussions`: GET `/sheets/{{ config.sheet_id }}/proofs/{{ config.proof_id
  }}/discussions` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `pageSize`; starts at 1.
- `sheet_proof_request_actions`: GET `/sheets/{{ config.sheet_id }}/proofs/{{ config.proof_id
  }}/requestactions` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `pageSize`; starts at 1.
- `sheet_proof_versions`: GET `/sheets/{{ config.sheet_id }}/proofs/{{ config.proof_id }}/versions`
  - records path `data`; page-number pagination; page parameter `page`; size parameter `pageSize`;
  starts at 1.
- `sheet_publish`: GET `/sheets/{{ config.sheet_id }}/publish` - single-object response.
- `sheet_row`: GET `/sheets/{{ config.sheet_id }}/rows/{{ config.row_id }}` - single-object
  response.
- `sheet_row_attachments`: GET `/sheets/{{ config.sheet_id }}/rows/{{ config.row_id }}/attachments`
  - records path `data`; page-number pagination; page parameter `page`; size parameter `pageSize`;
  starts at 1.
- `sheet_cell_history`: GET `/sheets/{{ config.sheet_id }}/rows/{{ config.row_id }}/columns/{{
  config.column_id }}/history` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `pageSize`; starts at 1.
- `sheet_row_discussions`: GET `/sheets/{{ config.sheet_id }}/rows/{{ config.row_id }}/discussions`
  - records path `data`; page-number pagination; page parameter `page`; size parameter `pageSize`;
  starts at 1.
- `sheet_sent_update_requests`: GET `/sheets/{{ config.sheet_id }}/sentupdaterequests` - records
  path `data`; page-number pagination; page parameter `page`; size parameter `pageSize`; starts at
  1.
- `sheet_sent_update_request`: GET `/sheets/{{ config.sheet_id }}/sentupdaterequests/{{
  config.sent_update_request_id }}` - single-object response.
- `sheet_summary`: GET `/sheets/{{ config.sheet_id }}/summary` - single-object response.
- `sheet_summary_fields`: GET `/sheets/{{ config.sheet_id }}/summary/fields` - records path `data`;
  page-number pagination; page parameter `page`; size parameter `pageSize`; starts at 1.
- `sheet_update_requests`: GET `/sheets/{{ config.sheet_id }}/updaterequests` - records path `data`;
  page-number pagination; page parameter `page`; size parameter `pageSize`; starts at 1.
- `sheet_update_request`: GET `/sheets/{{ config.sheet_id }}/updaterequests/{{
  config.update_request_id }}` - single-object response.
- `sheet_version`: GET `/sheets/{{ config.sheet_id }}/version` - single-object response.
- `dashboards`: GET `/sights` - records path `data`; query `maxItems`=`100`; cursor pagination;
  cursor parameter `lastKey`; next token from `lastKey`.
- `dashboard`: GET `/sights/{{ config.sight_id }}` - single-object response.
- `dashboard_path`: GET `/sights/{{ config.sight_id }}/path` - single-object response.
- `dashboard_publish`: GET `/sights/{{ config.sight_id }}/publish` - single-object response.
- `users`: GET `/users` - records path `data`; page-number pagination; page parameter `page`; size
  parameter `pageSize`; starts at 1.
- `current_user`: GET `/users/me` - single-object response.
- `org_sheets`: GET `/users/sheets` - records path `data`.
- `user`: GET `/users/{{ config.user_id }}` - single-object response.
- `user_alternate_emails`: GET `/users/{{ config.user_id }}/alternateemails` - records path `data`.
- `user_alternate_email`: GET `/users/{{ config.user_id }}/alternateemails/{{
  config.alternate_email_id }}` - single-object response.
- `user_plans`: GET `/users/{{ config.user_id }}/plans` - records path `data`; query
  `maxItems`=`100`; cursor pagination; cursor parameter `lastKey`; next token from `lastKey`.
- `webhooks`: GET `/webhooks` - records path `data`; page-number pagination; page parameter `page`;
  size parameter `pageSize`; starts at 1.
- `webhook`: GET `/webhooks/{{ config.webhook_id }}` - single-object response.
- `workspaces`: GET `/workspaces` - records path `data`; query `maxItems`=`100`; cursor pagination;
  cursor parameter `lastKey`; next token from `lastKey`.
- `workspace_metadata`: GET `/workspaces/{{ config.workspace_id }}/metadata` - single-object
  response.
- `workspace_children`: GET `/workspaces/{{ config.workspace_id }}/children` - records path `data`;
  query `maxItems`=`100`; cursor pagination; cursor parameter `lastKey`; next token from `lastKey`.

## Write actions & risks

Overall write risk: creates, updates, copies, moves, publishes, shares, and deletes Smartsheet
business objects including rows, sheets, reports, folders, comments, attachments, proofs, update
requests, webhooks, and workspaces; destructive deletes require approval.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `add_favorites`: POST `/favorites` - kind `create`; body type `json`; risk: adds one or more
  favorites.
- `delete_favorite`: DELETE `/favorites/{{ record.favorite_type }}/{{ record.favorite_id }}` - kind
  `delete`; body type `none`; path fields `favorite_type`, `favorite_id`; required record fields
  `favorite_type`, `favorite_id`; accepted fields `favorite_id`, `favorite_type`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: removes one favorite.
- `update_folder`: PUT `/folders/{{ record.folder_id }}` - kind `update`; body type `json`; path
  fields `folder_id`; required record fields `folder_id`; accepted fields `folder_id`; risk: updates
  folder metadata.
- `copy_folder`: POST `/folders/{{ record.folder_id }}/copy` - kind `custom`; body type `json`; path
  fields `folder_id`; required record fields `folder_id`; accepted fields `folder_id`; risk: copies
  a folder to a destination.
- `create_folder_in_folder`: POST `/folders/{{ record.folder_id }}/folders` - kind `create`; body
  type `json`; path fields `folder_id`; required record fields `folder_id`; accepted fields
  `folder_id`; risk: creates a child folder.
- `move_folder`: POST `/folders/{{ record.folder_id }}/move` - kind `custom`; body type `json`; path
  fields `folder_id`; required record fields `folder_id`; accepted fields `folder_id`; risk: moves a
  folder to a destination.
- `create_sheet_in_folder`: POST `/folders/{{ record.folder_id }}/sheets` - kind `create`; body type
  `json`; path fields `folder_id`; required record fields `folder_id`; accepted fields `folder_id`;
  risk: creates a sheet in a folder.
- `create_group`: POST `/groups` - kind `create`; body type `json`; risk: creates an organization
  group.
- `update_group`: PUT `/groups/{{ record.group_id }}` - kind `update`; body type `json`; path fields
  `group_id`; required record fields `group_id`; accepted fields `group_id`; risk: updates an
  organization group.
- `add_group_members`: POST `/groups/{{ record.group_id }}/members` - kind `create`; body type
  `json`; path fields `group_id`; required record fields `group_id`; accepted fields `group_id`;
  risk: adds members to a group.
- `delete_group_member`: DELETE `/groups/{{ record.group_id }}/members/{{ record.user_id }}` - kind
  `delete`; body type `none`; path fields `group_id`, `user_id`; required record fields `group_id`,
  `user_id`; accepted fields `group_id`, `user_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: removes a member from a group.
- `create_home_folder`: POST `/home/folders` - kind `create`; body type `json`; risk: creates a
  folder in home.
- `create_report`: POST `/reports` - kind `create`; body type `json`; risk: creates a report.
- `add_report_columns`: POST `/reports/{{ record.report_id }}/columns` - kind `create`; body type
  `json`; path fields `report_id`; required record fields `report_id`; accepted fields `report_id`;
  risk: adds columns to a report.
- `update_report_definition`: PUT `/reports/{{ record.report_id }}/definition` - kind `update`; body
  type `json`; path fields `report_id`; required record fields `report_id`; accepted fields
  `report_id`; risk: updates a report definition.
- `set_report_publish`: PUT `/reports/{{ record.report_id }}/publish` - kind `update`; body type
  `json`; path fields `report_id`; required record fields `report_id`; accepted fields `report_id`;
  risk: updates report publish settings.
- `add_report_scope`: POST `/reports/{{ record.report_id }}/scope` - kind `create`; body type
  `json`; path fields `report_id`; required record fields `report_id`; accepted fields `report_id`;
  risk: adds sheets to a report scope.
- `remove_report_scope`: DELETE `/reports/{{ record.report_id }}/scope` - kind `delete`; body type
  `json`; path fields `report_id`; required record fields `report_id`; accepted fields `report_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: removes
  sheets from a report scope.
- `create_sheet`: POST `/sheets` - kind `create`; body type `json`; risk: creates a sheet in the
  Sheets folder.
- `update_sheet`: PUT `/sheets/{{ record.sheet_id }}` - kind `update`; body type `json`; path fields
  `sheet_id`; required record fields `sheet_id`; accepted fields `sheet_id`; risk: updates sheet
  metadata.
- `attach_url_to_sheet`: POST `/sheets/{{ record.sheet_id }}/attachments` - kind `create`; body type
  `json`; path fields `sheet_id`; required record fields `sheet_id`; accepted fields `sheet_id`;
  risk: adds a URL attachment to a sheet.
- `delete_sheet_attachment`: DELETE `/sheets/{{ record.sheet_id }}/attachments/{{
  record.attachment_id }}` - kind `delete`; body type `none`; path fields `sheet_id`,
  `attachment_id`; required record fields `sheet_id`, `attachment_id`; accepted fields
  `attachment_id`, `sheet_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes a sheet attachment.
- `delete_sheet_attachment_versions`: DELETE `/sheets/{{ record.sheet_id }}/attachments/{{
  record.attachment_id }}/versions` - kind `delete`; body type `none`; path fields `sheet_id`,
  `attachment_id`; required record fields `sheet_id`, `attachment_id`; accepted fields
  `attachment_id`, `sheet_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes all versions of a sheet attachment.
- `update_automation_rule`: PUT `/sheets/{{ record.sheet_id }}/automationrules/{{
  record.automation_rule_id }}` - kind `update`; body type `json`; path fields `sheet_id`,
  `automation_rule_id`; required record fields `sheet_id`, `automation_rule_id`; accepted fields
  `automation_rule_id`, `sheet_id`; risk: updates an automation rule.
- `delete_automation_rule`: DELETE `/sheets/{{ record.sheet_id }}/automationrules/{{
  record.automation_rule_id }}` - kind `delete`; body type `none`; path fields `sheet_id`,
  `automation_rule_id`; required record fields `sheet_id`, `automation_rule_id`; accepted fields
  `automation_rule_id`, `sheet_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: deletes an automation rule.
- `add_sheet_columns`: POST `/sheets/{{ record.sheet_id }}/columns` - kind `create`; body type
  `json`; path fields `sheet_id`; required record fields `sheet_id`; accepted fields `sheet_id`;
  risk: adds columns to a sheet.
- `update_sheet_column`: PUT `/sheets/{{ record.sheet_id }}/columns/{{ record.column_id }}` - kind
  `update`; body type `json`; path fields `sheet_id`, `column_id`; required record fields
  `sheet_id`, `column_id`; accepted fields `column_id`, `sheet_id`; risk: updates a sheet column.
- `delete_sheet_column`: DELETE `/sheets/{{ record.sheet_id }}/columns/{{ record.column_id }}` -
  kind `delete`; body type `none`; path fields `sheet_id`, `column_id`; required record fields
  `sheet_id`, `column_id`; accepted fields `column_id`, `sheet_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: deletes a sheet column.
- `update_comment`: PUT `/sheets/{{ record.sheet_id }}/comments/{{ record.comment_id }}` - kind
  `update`; body type `json`; path fields `sheet_id`, `comment_id`; required record fields
  `sheet_id`, `comment_id`; accepted fields `comment_id`, `sheet_id`; risk: edits a comment.
- `delete_comment`: DELETE `/sheets/{{ record.sheet_id }}/comments/{{ record.comment_id }}` - kind
  `delete`; body type `none`; path fields `sheet_id`, `comment_id`; required record fields
  `sheet_id`, `comment_id`; accepted fields `comment_id`, `sheet_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: deletes a comment.
- `attach_url_to_comment`: POST `/sheets/{{ record.sheet_id }}/comments/{{ record.comment_id
  }}/attachments` - kind `create`; body type `json`; path fields `sheet_id`, `comment_id`; required
  record fields `sheet_id`, `comment_id`; accepted fields `comment_id`, `sheet_id`; risk: adds a URL
  attachment to a comment.
- `copy_sheet`: POST `/sheets/{{ record.sheet_id }}/copy` - kind `custom`; body type `json`; path
  fields `sheet_id`; required record fields `sheet_id`; accepted fields `sheet_id`; risk: copies a
  sheet.
- `create_cross_sheet_reference`: POST `/sheets/{{ record.sheet_id }}/crosssheetreferences` - kind
  `create`; body type `json`; path fields `sheet_id`; required record fields `sheet_id`; accepted
  fields `sheet_id`; risk: creates cross-sheet references.
- `create_sheet_discussion`: POST `/sheets/{{ record.sheet_id }}/discussions` - kind `create`; body
  type `json`; path fields `sheet_id`; required record fields `sheet_id`; accepted fields
  `sheet_id`; risk: creates a sheet discussion.
- `delete_sheet_discussion`: DELETE `/sheets/{{ record.sheet_id }}/discussions/{{
  record.discussion_id }}` - kind `delete`; body type `none`; path fields `sheet_id`,
  `discussion_id`; required record fields `sheet_id`, `discussion_id`; accepted fields
  `discussion_id`, `sheet_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes a sheet discussion.
- `create_discussion_comment`: POST `/sheets/{{ record.sheet_id }}/discussions/{{
  record.discussion_id }}/comments` - kind `create`; body type `json`; path fields `sheet_id`,
  `discussion_id`; required record fields `sheet_id`, `discussion_id`; accepted fields
  `discussion_id`, `sheet_id`; risk: adds a comment to a discussion.
- `move_sheet`: POST `/sheets/{{ record.sheet_id }}/move` - kind `custom`; body type `json`; path
  fields `sheet_id`; required record fields `sheet_id`; accepted fields `sheet_id`; risk: moves a
  sheet.
- `update_proof`: PUT `/sheets/{{ record.sheet_id }}/proofs/{{ record.proof_id }}` - kind `update`;
  body type `json`; path fields `sheet_id`, `proof_id`; required record fields `sheet_id`,
  `proof_id`; accepted fields `proof_id`, `sheet_id`; risk: updates proof status.
- `delete_proof`: DELETE `/sheets/{{ record.sheet_id }}/proofs/{{ record.proof_id }}` - kind
  `delete`; body type `none`; path fields `sheet_id`, `proof_id`; required record fields `sheet_id`,
  `proof_id`; accepted fields `proof_id`, `sheet_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: deletes a proof.
- `create_proof_discussion`: POST `/sheets/{{ record.sheet_id }}/proofs/{{ record.proof_id
  }}/discussions` - kind `create`; body type `json`; path fields `sheet_id`, `proof_id`; required
  record fields `sheet_id`, `proof_id`; accepted fields `proof_id`, `sheet_id`; risk: creates a
  proof discussion.
- `create_proof_request`: POST `/sheets/{{ record.sheet_id }}/proofs/{{ record.proof_id }}/requests`
  - kind `create`; body type `json`; path fields `sheet_id`, `proof_id`; required record fields
  `sheet_id`, `proof_id`; accepted fields `proof_id`, `sheet_id`; risk: creates proof requests.
- `delete_proof_requests`: DELETE `/sheets/{{ record.sheet_id }}/proofs/{{ record.proof_id
  }}/requests` - kind `delete`; body type `none`; path fields `sheet_id`, `proof_id`; required
  record fields `sheet_id`, `proof_id`; accepted fields `proof_id`, `sheet_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: deletes proof requests.
- `create_proof_version`: POST `/sheets/{{ record.sheet_id }}/proofs/{{ record.proof_id }}/versions`
  - kind `create`; body type `json`; path fields `sheet_id`, `proof_id`; required record fields
  `sheet_id`, `proof_id`; accepted fields `proof_id`, `sheet_id`; risk: creates a proof version from
  JSON metadata.
- `delete_proof_version`: DELETE `/sheets/{{ record.sheet_id }}/proofs/{{ record.proof_id
  }}/versions` - kind `delete`; body type `none`; path fields `sheet_id`, `proof_id`; required
  record fields `sheet_id`, `proof_id`; accepted fields `proof_id`, `sheet_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: deletes a proof version.
- `set_sheet_publish`: PUT `/sheets/{{ record.sheet_id }}/publish` - kind `update`; body type
  `json`; path fields `sheet_id`; required record fields `sheet_id`; accepted fields `sheet_id`;
  risk: updates sheet publish settings.
- `add_sheet_rows`: POST `/sheets/{{ record.sheet_id }}/rows` - kind `create`; body type `json`;
  path fields `sheet_id`; required record fields `sheet_id`; accepted fields `sheet_id`; risk: adds
  rows to a sheet.
- `update_sheet_rows`: PUT `/sheets/{{ record.sheet_id }}/rows` - kind `update`; body type `json`;
  path fields `sheet_id`; required record fields `sheet_id`; accepted fields `sheet_id`; risk:
  updates rows in a sheet.
- `copy_rows`: POST `/sheets/{{ record.sheet_id }}/rows/copy` - kind `custom`; body type `json`;
  path fields `sheet_id`; required record fields `sheet_id`; accepted fields `sheet_id`; risk:
  copies rows to another sheet.
- `move_rows`: POST `/sheets/{{ record.sheet_id }}/rows/move` - kind `custom`; body type `json`;
  path fields `sheet_id`; required record fields `sheet_id`; accepted fields `sheet_id`; risk: moves
  rows to another sheet.
- `attach_url_to_row`: POST `/sheets/{{ record.sheet_id }}/rows/{{ record.row_id }}/attachments` -
  kind `create`; body type `json`; path fields `sheet_id`, `row_id`; required record fields
  `sheet_id`, `row_id`; accepted fields `row_id`, `sheet_id`; risk: adds a URL attachment to a row.
- `create_row_discussion`: POST `/sheets/{{ record.sheet_id }}/rows/{{ record.row_id }}/discussions`
  - kind `create`; body type `json`; path fields `sheet_id`, `row_id`; required record fields
  `sheet_id`, `row_id`; accepted fields `row_id`, `sheet_id`; risk: creates a row discussion.
- `delete_sent_update_request`: DELETE `/sheets/{{ record.sheet_id }}/sentupdaterequests/{{
  record.sent_update_request_id }}` - kind `delete`; body type `none`; path fields `sheet_id`,
  `sent_update_request_id`; required record fields `sheet_id`, `sent_update_request_id`; accepted
  fields `sent_update_request_id`, `sheet_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: deletes a sent update request.
- `add_summary_fields`: POST `/sheets/{{ record.sheet_id }}/summary/fields` - kind `create`; body
  type `json`; path fields `sheet_id`; required record fields `sheet_id`; accepted fields
  `sheet_id`; risk: adds sheet summary fields.
- `update_summary_fields`: PUT `/sheets/{{ record.sheet_id }}/summary/fields` - kind `update`; body
  type `json`; path fields `sheet_id`; required record fields `sheet_id`; accepted fields
  `sheet_id`; risk: updates sheet summary fields.
- `create_update_request`: POST `/sheets/{{ record.sheet_id }}/updaterequests` - kind `create`; body
  type `json`; path fields `sheet_id`; required record fields `sheet_id`; accepted fields
  `sheet_id`; risk: creates an update request.
- `delete_update_request`: DELETE `/sheets/{{ record.sheet_id }}/updaterequests/{{
  record.update_request_id }}` - kind `delete`; body type `none`; path fields `sheet_id`,
  `update_request_id`; required record fields `sheet_id`, `update_request_id`; accepted fields
  `sheet_id`, `update_request_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes an update request.
- `sort_sheet_rows`: POST `/sheets/{{ record.sheet_id }}/sort` - kind `custom`; body type `json`;
  path fields `sheet_id`; required record fields `sheet_id`; accepted fields `sheet_id`; risk: sorts
  rows in a sheet.
- `update_dashboard`: PUT `/sights/{{ record.sight_id }}` - kind `update`; body type `json`; path
  fields `sight_id`; required record fields `sight_id`; accepted fields `sight_id`; risk: updates
  dashboard metadata.
- `copy_dashboard`: POST `/sights/{{ record.sight_id }}/copy` - kind `custom`; body type `json`;
  path fields `sight_id`; required record fields `sight_id`; accepted fields `sight_id`; risk:
  copies a dashboard.
- `move_dashboard`: POST `/sights/{{ record.sight_id }}/move` - kind `custom`; body type `json`;
  path fields `sight_id`; required record fields `sight_id`; accepted fields `sight_id`; risk: moves
  a dashboard.
- `set_dashboard_publish`: PUT `/sights/{{ record.sight_id }}/publish` - kind `update`; body type
  `json`; path fields `sight_id`; required record fields `sight_id`; accepted fields `sight_id`;
  risk: updates dashboard publish settings.
- `add_alternate_emails`: POST `/users/{{ record.user_id }}/alternateemails` - kind `create`; body
  type `json`; path fields `user_id`; required record fields `user_id`; accepted fields `user_id`;
  risk: adds alternate email addresses to a user.
- `delete_alternate_email`: DELETE `/users/{{ record.user_id }}/alternateemails/{{
  record.alternate_email_id }}` - kind `delete`; body type `none`; path fields `user_id`,
  `alternate_email_id`; required record fields `user_id`, `alternate_email_id`; accepted fields
  `alternate_email_id`, `user_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: deletes an alternate email address.
- `make_alternate_email_primary`: POST `/users/{{ record.user_id }}/alternateemails/{{
  record.alternate_email_id }}/makeprimary` - kind `custom`; body type `none`; path fields
  `user_id`, `alternate_email_id`; required record fields `user_id`, `alternate_email_id`; accepted
  fields `alternate_email_id`, `user_id`; risk: makes an alternate email primary.
- `create_webhook`: POST `/webhooks` - kind `create`; body type `json`; risk: creates a webhook.
- `update_webhook`: PUT `/webhooks/{{ record.webhook_id }}` - kind `update`; body type `json`; path
  fields `webhook_id`; required record fields `webhook_id`; accepted fields `webhook_id`; risk:
  updates a webhook.
- `delete_webhook`: DELETE `/webhooks/{{ record.webhook_id }}` - kind `delete`; body type `none`;
  path fields `webhook_id`; required record fields `webhook_id`; accepted fields `webhook_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: deletes a
  webhook.
- `create_workspace`: POST `/workspaces` - kind `create`; body type `json`; risk: creates a
  workspace.
- `update_workspace`: PUT `/workspaces/{{ record.workspace_id }}` - kind `update`; body type `json`;
  path fields `workspace_id`; required record fields `workspace_id`; accepted fields `workspace_id`;
  risk: updates a workspace.
- `copy_workspace`: POST `/workspaces/{{ record.workspace_id }}/copy` - kind `custom`; body type
  `json`; path fields `workspace_id`; required record fields `workspace_id`; accepted fields
  `workspace_id`; risk: copies a workspace.
- `create_workspace_folder`: POST `/workspaces/{{ record.workspace_id }}/folders` - kind `create`;
  body type `json`; path fields `workspace_id`; required record fields `workspace_id`; accepted
  fields `workspace_id`; risk: creates a folder in a workspace.
- `create_sheet_in_workspace`: POST `/workspaces/{{ record.workspace_id }}/sheets` - kind `create`;
  body type `json`; path fields `workspace_id`; required record fields `workspace_id`; accepted
  fields `workspace_id`; risk: creates a sheet in a workspace.

## Known limits

- Batch defaults: read_page_size=100.
- API coverage includes 69 stream-backed endpoint group(s), 72 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=9, destructive_admin=5, duplicate_of=1, non_data_endpoint=7, out_of_scope=9,
  requires_elevated_scope=10.
