# pm connectors inspect dropbox-sign

```text
NAME
  pm connectors inspect dropbox-sign - Dropbox Sign connector manual

SYNOPSIS
  pm connectors inspect dropbox-sign
  pm connectors inspect dropbox-sign --json
  pm credentials add <name> --connector dropbox-sign [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads Dropbox Sign (HelloSign) signature requests, templates, team members, and account details, and writes signature-request/template/team/account lifecycle mutations, through the Dropbox Sign REST API.

ICON
  id: simple-icons-dropbox
  asset: icons/simple-icons/dropbox.svg
  title: Dropbox
  simple_icon_slug: dropbox
  simple_icon_hex: 0061FF
  source: simple-icons
  license: CC0-1.0
  review_status: cc0_with_trademark_caveat
  review_url: https://simpleicons.org/?q=Dropbox
  match: curated-alias
  matched_by: dropbox

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  mode
  api_key (secret)

ETL STREAMS
  signature_requests:
    primary key: signature_request_id
    cursor: created_at
    fields: created_at(integer), has_error(boolean), is_complete(boolean), is_declined(boolean), message(string), requester_email_address(string), signature_request_id(string), subject(string), test_mode(boolean), title(string)
  templates:
    primary key: template_id
    cursor: updated_at
    fields: is_creator(boolean), is_embedded(boolean), is_locked(boolean), message(string), template_id(string), title(string), updated_at(integer)
  team_members:
    primary key: account_id
    fields: account_id(string), email_address(string), role(string)
  account:
    primary key: account_id
    fields: account_id(string), email_address(string), is_paid_hf(boolean), is_paid_hs(boolean), locale(string), role_code(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  update_signature_request:
    endpoint: POST /signature_request/update/{{ record.signature_request_id }}
    required fields: signature_request_id, signature_id
    risk: external mutation; changes a signer's email address or name on an in-progress signature request, redirecting where the next request/reminder is delivered; approval required
  cancel_signature_request:
    endpoint: POST /signature_request/cancel/{{ record.signature_request_id }}
    required fields: signature_request_id
    risk: destructive external mutation; cancels an incomplete signature request, this action is not reversible; approval required
  remind_signature_request:
    endpoint: POST /signature_request/remind/{{ record.signature_request_id }}
    required fields: signature_request_id, email_address
    risk: external mutation; sends an email reminder to a signer; cannot be sent again within 1 hour of the last reminder (manual or automatic)
  release_hold_signature_request:
    endpoint: POST /signature_request/release_hold/{{ record.signature_request_id }}
    required fields: signature_request_id
    risk: external mutation; releases a held signature request created from an UnclaimedDraft, immediately sending requests to all signers; approval required
  remove_signature_request:
    endpoint: POST /signature_request/remove/{{ record.signature_request_id }}
    required fields: signature_request_id
    risk: destructive external mutation; removes the caller's access to a completed signature request from the account's list view, this action is not reversible; approval required
  delete_template:
    endpoint: POST /template/delete/{{ record.template_id }}
    required fields: template_id
    risk: destructive external mutation; completely deletes a template from the account, this action is not reversible; approval required
  add_template_user:
    endpoint: POST /template/add_user/{{ record.template_id }}
    required fields: template_id
    risk: external mutation; grants the specified account (which must already be a Team member) access to a template
  remove_template_user:
    endpoint: POST /template/remove_user/{{ record.template_id }}
    required fields: template_id
    risk: external mutation; revokes the specified account's access to a template
  create_team:
    endpoint: POST /team/create
    risk: external mutation; creates a new Team and makes the calling account its member; fails if the caller already belongs to a Team
  update_team:
    endpoint: PUT /team
    required fields: name
    risk: external mutation; renames the caller's own Team
  add_team_member:
    endpoint: PUT /team/add_member
    risk: external mutation; invites or moves a user onto the caller's Team, creating a new Dropbox Sign account for the invited email if one does not already exist
  remove_team_member:
    endpoint: POST /team/remove_member
    risk: destructive external mutation; removes a user from the caller's Team; optionally transfers the removed account's documents to another account (Enterprise plans only), which is not reversible; approval required
  update_account:
    endpoint: PUT /account
    risk: external mutation; updates the caller's account settings (currently limited to the event callback URL and locale)

SECURITY
  read risk: external Dropbox Sign API read of signature requests, templates, team members, and account data
  write risk: external mutation of signature requests (update/cancel/remind/release_hold/remove), templates (delete/add_user/remove_user), teams (create/update/add_member/remove_member), and account settings; several actions are destructive/not reversible (cancel_signature_request, remove_signature_request, delete_template, remove_team_member) and require approval
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run Dropbox Sign's declared streams and reverse-ETL actions.
  Usage: pm dropbox-sign <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    account list - Run the account ETL stream [intent=etl availability=implemented stream=account]
    add team member apply - Plan and execute the add team member reverse-ETL action [intent=reverse_etl availability=implemented write=add_team_member]; approval: requires plan, preview, approval, and execute; risk: external mutation; invites or moves a user onto the caller's Team, creating a new Dropbox Sign account for the invited email if one does not already exist
    add template user apply - Plan and execute the add template user reverse-ETL action [intent=reverse_etl availability=implemented write=add_template_user]; approval: requires plan, preview, approval, and execute; risk: external mutation; grants the specified account (which must already be a Team member) access to a template; flags: --template_id (required)
    api delete api-app client-id - Documented DELETE /api_app/{client_id} (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.delete.api-app-client-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete fax fax-id - Documented DELETE /fax/{fax_id} (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.delete.fax-fax-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete fax-line - Documented DELETE /fax_line (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.delete.fax-line]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete team destroy - Documented DELETE /team/destroy (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.delete.team-destroy]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get api-app client-id - Documented GET /api_app/{client_id} (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.api-app-client-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api-app list - Documented GET /api_app/list (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.api-app-list]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get bulk-send-job bulk-send-job-id - Documented GET /bulk_send_job/{bulk_send_job_id} (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.bulk-send-job-bulk-send-job-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get bulk-send-job list - Documented GET /bulk_send_job/list (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.bulk-send-job-list]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get embedded sign-url signature-id - Documented GET /embedded/sign_url/{signature_id} (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.embedded-sign-url-signature-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get fax fax-id - Documented GET /fax/{fax_id} (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.fax-fax-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get fax files fax-id - Documented GET /fax/files/{fax_id} (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.fax-files-fax-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get fax list - Documented GET /fax/list (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.fax-list]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get fax-line - Documented GET /fax_line (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.fax-line]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get fax-line area-codes - Documented GET /fax_line/area_codes (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.fax-line-area-codes]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get fax-line list - Documented GET /fax_line/list (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.fax-line-list]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get signature-request files signature-request-id - Documented GET /signature_request/files/{signature_request_id} (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.signature-request-files-signature-request-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get signature-request files-as-data-uri signature-request-id - Documented GET /signature_request/files_as_data_uri/{signature_request_id} (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.signature-request-files-as-data-uri-signature-request-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get signature-request files-as-file-url signature-request-id - Documented GET /signature_request/files_as_file_url/{signature_request_id} (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.signature-request-files-as-file-url-signature-request-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get signature-request signature-request-id - Documented GET /signature_request/{signature_request_id} (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.signature-request-signature-request-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get team - Documented GET /team (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.team]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get team info - Documented GET /team/info (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.team-info]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get team invites - Documented GET /team/invites (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.team-invites]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get team sub-teams team-id - Documented GET /team/sub_teams/{team_id} (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.team-sub-teams-team-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get template files template-id - Documented GET /template/files/{template_id} (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.template-files-template-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get template files-as-data-uri template-id - Documented GET /template/files_as_data_uri/{template_id} (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.template-files-as-data-uri-template-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get template files-as-file-url template-id - Documented GET /template/files_as_file_url/{template_id} (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.template-files-as-file-url-template-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get template template-id - Documented GET /template/{template_id} (not implemented) [intent=direct_read availability=not_implemented operation=dropbox-sign.get.template-template-id]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api post account create - Documented POST /account/create (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.account-create]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post account verify - Documented POST /account/verify (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.account-verify]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api-app - Documented POST /api_app (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.api-app]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post embedded edit-url template-id - Documented POST /embedded/edit_url/{template_id} (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.embedded-edit-url-template-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post fax send - Documented POST /fax/send (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.fax-send]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post fax-line create - Documented POST /fax_line/create (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.fax-line-create]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post notification retry - Documented POST /notification/retry (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.notification-retry]; approval: not implemented: the provider webhook-URL mutation has no dedicated security-reviewed execution contract; risk: high; notes: named_dependency=review.webhook_url_mutation: the provider webhook-URL mutation has no dedicated security-reviewed execution contract
    api post oauth token - Documented POST /oauth/token (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.oauth-token]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post oauth token-refresh - Documented POST /oauth/token?refresh (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.oauth-token-refresh]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post report create - Documented POST /report/create (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.report-create]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post signature-request bulk-create-embedded-with-template - Documented POST /signature_request/bulk_create_embedded_with_template (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.signature-request-bulk-create-embedded-with-template]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post signature-request bulk-send-with-template - Documented POST /signature_request/bulk_send_with_template (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.signature-request-bulk-send-with-template]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post signature-request create-embedded - Documented POST /signature_request/create_embedded (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.signature-request-create-embedded]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post signature-request create-embedded-with-template - Documented POST /signature_request/create_embedded_with_template (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.signature-request-create-embedded-with-template]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post signature-request send - Documented POST /signature_request/send (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.signature-request-send]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post signature-request send-with-template - Documented POST /signature_request/send_with_template (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.signature-request-send-with-template]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post template create - Documented POST /template/create (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.template-create]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post template create-embedded-draft - Documented POST /template/create_embedded_draft (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.template-create-embedded-draft]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post template update template-id - Documented POST /template/update/{template_id} (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.template-update-template-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post template update-files template-id - Documented POST /template/update_files/{template_id} (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.template-update-files-template-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post unclaimed-draft create - Documented POST /unclaimed_draft/create (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.unclaimed-draft-create]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post unclaimed-draft create-embedded - Documented POST /unclaimed_draft/create_embedded (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.unclaimed-draft-create-embedded]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post unclaimed-draft create-embedded-with-template - Documented POST /unclaimed_draft/create_embedded_with_template (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.unclaimed-draft-create-embedded-with-template]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post unclaimed-draft edit-and-resend signature-request-id - Documented POST /unclaimed_draft/edit_and_resend/{signature_request_id} (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.post.unclaimed-draft-edit-and-resend-signature-request-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put api-app client-id - Documented PUT /api_app/{client_id} (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.put.api-app-client-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put fax-line add-user - Documented PUT /fax_line/add_user (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.put.fax-line-add-user]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put fax-line remove-user - Documented PUT /fax_line/remove_user (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.put.fax-line-remove-user]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put signature-request edit signature-request-id - Documented PUT /signature_request/edit/{signature_request_id} (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.put.signature-request-edit-signature-request-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put signature-request edit-embedded signature-request-id - Documented PUT /signature_request/edit_embedded/{signature_request_id} (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.put.signature-request-edit-embedded-signature-request-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put signature-request edit-embedded-with-template signature-request-id - Documented PUT /signature_request/edit_embedded_with_template/{signature_request_id} (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.put.signature-request-edit-embedded-with-template-signature-request-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put signature-request edit-with-template signature-request-id - Documented PUT /signature_request/edit_with_template/{signature_request_id} (not implemented) [intent=direct_write availability=not_implemented operation=dropbox-sign.put.signature-request-edit-with-template-signature-request-id]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    cancel signature request apply - Plan and execute the cancel signature request reverse-ETL action [intent=reverse_etl availability=implemented write=cancel_signature_request]; approval: requires plan, preview, approval, and execute; risk: destructive external mutation; cancels an incomplete signature request, this action is not reversible; approval required; flags: --signature_request_id (required)
    create team apply - Plan and execute the create team reverse-ETL action [intent=reverse_etl availability=implemented write=create_team]; approval: requires plan, preview, approval, and execute; risk: external mutation; creates a new Team and makes the calling account its member; fails if the caller already belongs to a Team
    delete template apply - Plan and execute the delete template reverse-ETL action [intent=reverse_etl availability=implemented write=delete_template]; approval: requires plan, preview, approval, and execute; risk: destructive external mutation; completely deletes a template from the account, this action is not reversible; approval required; flags: --template_id (required)
    release hold signature request apply - Plan and execute the release hold signature request reverse-ETL action [intent=reverse_etl availability=implemented write=release_hold_signature_request]; approval: requires plan, preview, approval, and execute; risk: external mutation; releases a held signature request created from an UnclaimedDraft, immediately sending requests to all signers; approval required; flags: --signature_request_id (required)
    remind signature request apply - Plan and execute the remind signature request reverse-ETL action [intent=reverse_etl availability=implemented write=remind_signature_request]; approval: requires plan, preview, approval, and execute; risk: external mutation; sends an email reminder to a signer; cannot be sent again within 1 hour of the last reminder (manual or automatic); flags: --email_address (required), --signature_request_id (required)
    remove signature request apply - Plan and execute the remove signature request reverse-ETL action [intent=reverse_etl availability=implemented write=remove_signature_request]; approval: requires plan, preview, approval, and execute; risk: destructive external mutation; removes the caller's access to a completed signature request from the account's list view, this action is not reversible; approval required; flags: --signature_request_id (required)
    remove team member apply - Plan and execute the remove team member reverse-ETL action [intent=reverse_etl availability=implemented write=remove_team_member]; approval: requires plan, preview, approval, and execute; risk: destructive external mutation; removes a user from the caller's Team; optionally transfers the removed account's documents to another account (Enterprise plans only), which is not reversible; approval required
    remove template user apply - Plan and execute the remove template user reverse-ETL action [intent=reverse_etl availability=implemented write=remove_template_user]; approval: requires plan, preview, approval, and execute; risk: external mutation; revokes the specified account's access to a template; flags: --template_id (required)
    signature requests list - Run the signature requests ETL stream [intent=etl availability=implemented stream=signature_requests]
    team members list - Run the team members ETL stream [intent=etl availability=implemented stream=team_members]
    templates list - Run the templates ETL stream [intent=etl availability=implemented stream=templates]
    update account apply - Plan and execute the update account reverse-ETL action [intent=reverse_etl availability=implemented write=update_account]; approval: requires plan, preview, approval, and execute; risk: external mutation; updates the caller's account settings (currently limited to the event callback URL and locale)
    update signature request apply - Plan and execute the update signature request reverse-ETL action [intent=reverse_etl availability=implemented write=update_signature_request]; approval: requires plan, preview, approval, and execute; risk: external mutation; changes a signer's email address or name on an in-progress signature request, redirecting where the next request/reminder is delivered; approval required; flags: --signature_id (required), --signature_request_id (required)
    update team apply - Plan and execute the update team reverse-ETL action [intent=reverse_etl availability=implemented write=update_team]; approval: requires plan, preview, approval, and execute; risk: external mutation; renames the caller's own Team; flags: --name (required)

EXAMPLES
  # Inspect as a manual
  pm connectors inspect dropbox-sign

  # Inspect as structured JSON
  pm connectors inspect dropbox-sign --json

AGENT WORKFLOW
  - Run pm connectors inspect dropbox-sign before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
