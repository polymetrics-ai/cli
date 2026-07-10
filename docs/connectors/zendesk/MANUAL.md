# pm connectors inspect zendesk

```text
NAME
  pm connectors inspect zendesk - Zendesk connector manual

SYNOPSIS
  pm connectors inspect zendesk
  pm connectors inspect zendesk --json
  pm credentials add <name> --connector zendesk [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Operation ledger for the official Zendesk Support API surface; executable streams, bounded direct reads, and reverse-ETL writes are mapped by later Zendesk CLI parity sub-issues.

ICON
  asset: icons/pm-sample.svg
  source: polymetrics
  review_status: polymetrics
  review_url: https://github.com/polymetrics-ai/cli

CAPABILITIES
  check=true catalog=true read=true write=false query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  base_url
  access_token (secret)
  api_token (secret)
  email (secret)

SECURITY
  read risk: bounded direct-read JSON commands are enabled for 282 typed GET operations; ETL streams and binary downloads are still gated by later Zendesk lanes
  write risk: operation-ledger bundle in issue #160; Zendesk mutations remain blocked until typed reverse-ETL schemas, approval text, redaction, and destructive confirmation are added
  approval: all external writes remain plan → preview → approval → execute and blocked in this slice
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Work with Zendesk Support API metadata from the command line.
  Usage: pm zendesk <command> [flags]
  Source CLI: Zendesk API reference (https://developer.zendesk.com/zendesk/oas.yaml)
  Global flags:
    --json (boolean): Write machine-readable JSON output.
    --connection (string): Use a saved Zendesk connector credential/account scope.: maps_to=connection
  Read Commands
    read candidates - Review Zendesk direct-read operations from the official OAS [intent=docs_only availability=planned]; notes: This bundle now exposes 282 typed, bounded JSON direct-read commands. ETL streams and binary downloads remain separate lanes.
    read get-sources-by-target - Get sources by target [intent=direct_read availability=implemented]; flags: --target_type, --target_id, --field_id, --source_type
    read get-account-email-settings - Show Email Settings [intent=direct_read availability=implemented]
    read show-account-settings - Show Settings [intent=direct_read availability=implemented]; flags: --authenticity_token
    read verify-subdomain-availability - Verify Subdomain Availability [intent=direct_read availability=implemented]; flags: --subdomain
    read list-activities - List Activities [intent=direct_read availability=implemented]; flags: --include
    read show-activity - Show Activity [intent=direct_read availability=implemented]; flags: --activity_id
    read count-activities - Count Activities [intent=direct_read availability=implemented]
    read list-approval-requests - List Approval Requests [intent=direct_read availability=implemented]; flags: --filter-status, --filter-assignee_user_id, --filter-assignee_group_id, --before_cursor, --after_cursor
    read list-audit-logs - List Audit Logs [intent=direct_read availability=implemented]; flags: --filter-source_type, --filter-source_id, --filter-actor_id, --filter-ip_address, --filter-created_at, --filter-action, --sort_by, --sort_order, --sort
    read show-audit-log - Show Audit Log [intent=direct_read availability=implemented]; flags: --audit_log_id
    read autocomplete-tags - Search Tags [intent=direct_read availability=implemented]
    read list-automations - List Automations [intent=direct_read availability=implemented]
    read show-automation - Show Automation [intent=direct_read availability=implemented]; flags: --automation_id
    read list-active-automations - List Active Automations [intent=direct_read availability=implemented]
    read search-automations - Search Automations [intent=direct_read availability=implemented]
    read list-brand-agents - List Brand Agent Memberships [intent=direct_read availability=implemented]
    read show-brand-agent-by-id - Show Brand Agent Membership [intent=direct_read availability=implemented]; flags: --brand_agent_id
    read list-brands - List Brands [intent=direct_read availability=implemented]; flags: --page, --per_page, --assignable_from, --include_deleted
    read show-brand - Show a Brand [intent=direct_read availability=implemented]; flags: --brand_id
    read list-brand-agents-by-brand - List Agents By Brand [intent=direct_read availability=implemented]; flags: --brand_id
    read check-host-mapping-validity-for-existing-brand - Check Host Mapping Validity for an Existing Brand [intent=direct_read availability=implemented]; flags: --brand_id
    read check-host-mapping-validity - Check Host Mapping Validity [intent=direct_read availability=implemented]
    read list-monitored-twitter-handles - List Monitored X Handles [intent=direct_read availability=implemented]
    read show-monitored-twitter-handle - Show Monitored X Handle [intent=direct_read availability=implemented]; flags: --monitored_twitter_handle_id
    read getting-twicket-status - List Ticket statuses [intent=direct_read availability=implemented]; flags: --comment_id, --ids
    read list-custom-objects - List Custom Objects [intent=direct_read availability=implemented]; flags: --include_ui_path
    read show-custom-object - Show Custom Object [intent=direct_read availability=implemented]; flags: --custom_object_key, --include_permissions_metadata, --include_ui_path
    read list-access-rules - List Access Rules [intent=direct_read availability=implemented]; flags: --custom_object_key
    read show-access-rule - Show Access Rule [intent=direct_read availability=implemented]; flags: --custom_object_key, --id
    read list-access-rule-definitions - List Access Rule Definitions [intent=direct_read availability=implemented]; flags: --custom_object_key
    read list-custom-object-fields - List Custom Object Fields [intent=direct_read availability=implemented]; flags: --custom_object_key
    read show-custom-object-field - Show Custom Object Field [intent=direct_read availability=implemented]; flags: --custom_object_key, --custom_object_field_key_or_id, --include_standard_fields
    read custom-object-fields-limit - Custom Object Fields Limit [intent=direct_read availability=implemented]; flags: --custom_object_key
    read list-permission-policies - List Permission Policies [intent=direct_read availability=implemented]; flags: --custom_object_key
    read show-permission-policy - Show Permission Policy [intent=direct_read availability=implemented]; flags: --custom_object_key, --id
    read list-custom-object-records - List Custom Object Records [intent=direct_read availability=implemented]; flags: --custom_object_key, --filter-ids, --filter-external_ids, --sort, --page-before, --page-after, --page-size
    read show-custom-object-record - Show Custom Object Record [intent=direct_read availability=implemented]; flags: --custom_object_key, --custom_object_record_id
    read autocomplete-custom-object-record-search - Autocomplete Custom Object Record Search [intent=direct_read availability=implemented]; flags: --custom_object_key, --name, --page-before, --page-after, --page-size, --field_id, --source, --filter-dynamic_values, --requester_id, --assignee_id, --organization_id
    read count-custom-object-records - Count Custom Object Records [intent=direct_read availability=implemented]; flags: --custom_object_key
    read search-custom-object-records - Search Custom Object Records [intent=direct_read availability=implemented]; flags: --custom_object_key, --query, --sort, --page-before, --page-after, --page-size
    read list-object-triggers - List Object Triggers [intent=direct_read availability=implemented]; flags: --custom_object_key
    read get-object-trigger - Show Object Trigger [intent=direct_read availability=implemented]; flags: --custom_object_key, --trigger_id
    read list-active-object-triggers - List Active Object Triggers [intent=direct_read availability=implemented]; flags: --custom_object_key
    read list-object-triggers-definitions - List Object Trigger Action and Condition Definitions [intent=direct_read availability=implemented]; flags: --custom_object_key
    read search-object-triggers - Search Object Triggers [intent=direct_read availability=implemented]; flags: --custom_object_key
    read custom-objects-limit - Custom Objects Limit [intent=direct_read availability=implemented]
    read custom-object-records-limit - Custom Object Records Limit [intent=direct_read availability=implemented]
    read list-custom-roles - List Custom Roles [intent=direct_read availability=implemented]
    read show-custom-role-by-id - Show Custom Role [intent=direct_read availability=implemented]; flags: --custom_role_id
    read list-custom-statuses - List Custom Ticket Statuses [intent=direct_read availability=implemented]; flags: --status_categories, --active, --default
    read show-custom-status - Show Custom Ticket Status [intent=direct_read availability=implemented]; flags: --custom_status_id
    read list-deleted-users - List Deleted Users [intent=direct_read availability=implemented]
    read show-deleted-user - Show Deleted User [intent=direct_read availability=implemented]; flags: --deleted_user_id
    read count-deleted-users - Count Deleted Users [intent=direct_read availability=implemented]
    read list-deletion-schedules - List Deletion Schedules [intent=direct_read availability=implemented]
    read get-deletion-schedule - Get Deletion Schedule [intent=direct_read availability=implemented]; flags: --deletion_schedule_id
    read list-email-notifications - List Email Notifications [intent=direct_read availability=implemented]
    read show-email-notification - Show Email Notification [intent=direct_read availability=implemented]; flags: --notification_id
    read show-many-email-notifications - Show Many Email Notifications [intent=direct_read availability=implemented]
    read list-end-user-identities - List End User Identities [intent=direct_read availability=implemented]; flags: --user_id, --type
    read show-end-user-identity - Show End User Identity [intent=direct_read availability=implemented]; flags: --user_id, --user_identity_id
    read list-group-memberships - List Memberships [intent=direct_read availability=implemented]
    read show-group-membership-by-id - Show Membership [intent=direct_read availability=implemented]; flags: --group_membership_id
    read list-assignable-group-memberships - List Assignable Memberships [intent=direct_read availability=implemented]
    read list-group-sla-policies - List Group SLA Policies [intent=direct_read availability=implemented]
    read show-group-sla-policy - Show Group SLA Policy [intent=direct_read availability=implemented]; flags: --group_sla_policy_id
    read retrieve-group-sla-policy-filter-definition-items - Retrieve Supported Filter Definition Items [intent=direct_read availability=implemented]
    read list-groups - List Groups [intent=direct_read availability=implemented]
    read show-group-by-id - Show Group [intent=direct_read availability=implemented]; flags: --group_id
    read list-group-memberships-by-group - List Memberships By Group [intent=direct_read availability=implemented]; flags: --group_id
    read list-assignable-group-memberships-by-group - List Assignable Memberships By Group [intent=direct_read availability=implemented]; flags: --group_id
    read list-group-users - List Users By Group [intent=direct_read availability=implemented]; flags: --group_id
    read count-group-users - Count Users By Group [intent=direct_read availability=implemented]; flags: --group_id
    read list-assignable-groups - List Assignable Groups [intent=direct_read availability=implemented]
    read count-groups - Count Groups [intent=direct_read availability=implemented]
    read list-itam-asset-types - List Asset Types [intent=direct_read availability=implemented]
    read show-itam-asset-type - Show Asset Type [intent=direct_read availability=implemented]; flags: --asset_type_id
    read list-itam-asset-type-fields - List Asset Fields [intent=direct_read availability=implemented]; flags: --asset_type_id
    read show-itam-asset-type-field - Show Asset Field [intent=direct_read availability=implemented]; flags: --asset_type_id, --asset_type_field_id
    read list-itam-assets - List Assets [intent=direct_read availability=implemented]; flags: --filter-ids, --filter-external_ids
    read show-itam-asset - Show Asset [intent=direct_read availability=implemented]; flags: --asset_id
    read search-itam-assets - Search Assets [intent=direct_read availability=implemented]; flags: --query, --sort, --page-before, --page-after, --page-size
    read list-itam-locations - List Asset Locations [intent=direct_read availability=implemented]
    read show-itam-location - Show Asset Location [intent=direct_read availability=implemented]; flags: --location_id
    read list-itam-statuses - List Asset Statuses [intent=direct_read availability=implemented]
    read show-itam-status - Show Asset Status [intent=direct_read availability=implemented]; flags: --status_id
    read list-job-statuses - List Job Statuses [intent=direct_read availability=implemented]
    read show-job-status - Show Job Status [intent=direct_read availability=implemented]; flags: --job_status_id
    read show-many-job-statuses - Show Many Job Statuses [intent=direct_read availability=implemented]; flags: --ids
    read list-locales - List Locales [intent=direct_read availability=implemented]
    read show-locale-by-id - Show Locale [intent=direct_read availability=implemented]; flags: --locale_id
    read list-locales-for-agent - List Locales for Agent [intent=direct_read availability=implemented]
    read show-current-locale - Show Current Locale [intent=direct_read availability=implemented]
    read detect-best-locale - Detect Best Language for User [intent=direct_read availability=implemented]
    read list-available-public-locales - List Available Public Locales [intent=direct_read availability=implemented]
    read list-macros - List Macros [intent=direct_read availability=implemented]
    read show-macro - Show Macro [intent=direct_read availability=implemented]; flags: --macro_id
    read show-changes-to-ticket - Show Changes to Ticket [intent=direct_read availability=implemented]; flags: --macro_id, --normalize_comment
    read list-macros-actions - List Supported Actions for Macros [intent=direct_read availability=implemented]
    read list-active-macros - List Active Macros [intent=direct_read availability=implemented]
    read list-macro-categories - List Macro Categories [intent=direct_read availability=implemented]
    read list-macro-action-definitions - List Macro Action Definitions [intent=direct_read availability=implemented]
    read show-derived-macro - Show Macro Replica [intent=direct_read availability=implemented]; flags: --ticket_id
    read search-macro - Search Macros [intent=direct_read availability=implemented]
    read list-o-auth-clients - List Clients [intent=direct_read availability=implemented]
    read show-client - Show Client [intent=direct_read availability=implemented]; flags: --oauth_client_id
    read list-global-o-auth-clients - List Global OAuth Clients [intent=direct_read availability=implemented]
    read show-global-client - Show Global OAuth Client [intent=direct_read availability=implemented]; flags: --global_client_id
    read global-o-auth-clients-token-summary - Show Token summary for Global OAuth Clients [intent=direct_read availability=implemented]
    read list-o-auth-tokens - List Tokens [intent=direct_read availability=implemented]
    read show-token - Show Token [intent=direct_read availability=implemented]; flags: --oauth_token_id
    read show-current-token - Show Current Token [intent=direct_read availability=implemented]
    read list-organization-fields - List Organization Fields [intent=direct_read availability=implemented]; flags: --resolve_dc
    read show-organization-field - Show Organization Field [intent=direct_read availability=implemented]; flags: --organization_field_id
    read list-organization-memberships - List Memberships [intent=direct_read availability=implemented]
    read show-organization-membership-by-id - Show Membership [intent=direct_read availability=implemented]; flags: --organization_membership_id
    read show-organization-merge - Show Organization Merge [intent=direct_read availability=implemented]; flags: --organization_merge_id
    read list-organization-subscriptions - List Organization Subscriptions [intent=direct_read availability=implemented]
    read show-organization-subscription - Show Organization Subscription [intent=direct_read availability=implemented]; flags: --organization_subscription_id
    read list-organizations - List Organizations [intent=direct_read availability=implemented]
    read show-organization - Show Organization [intent=direct_read availability=implemented]; flags: --organization_id, --include
    read list-organization-merges - List Organization Merges [intent=direct_read availability=implemented]; flags: --organization_id
    read list-organization-memberships-by-organization - List Organization Memberships by Organization [intent=direct_read availability=implemented]; flags: --organization_id
    read organization-related - Show Organization's Related Information [intent=direct_read availability=implemented]; flags: --organization_id
    read list-organization-requests - List Organization Requests [intent=direct_read availability=implemented]; flags: --organization_id, --sort_by, --sort_order
    read list-organization-subscriptions-by-organization - List Subscriptions By Organization [intent=direct_read availability=implemented]; flags: --organization_id
    read list-organization-tags - List Organization Tags [intent=direct_read availability=implemented]; flags: --organization_id
    read list-organization-tickets - List Organization Tickets [intent=direct_read availability=implemented]; flags: --organization_id
    read count-organization-tickets - Count Organization Tickets [intent=direct_read availability=implemented]; flags: --organization_id
    read list-organization-users - List Organization Users [intent=direct_read availability=implemented]; flags: --organization_id
    read count-organization-users - Count Organization Users [intent=direct_read availability=implemented]; flags: --organization_id
    read autocomplete-organizations - Autocomplete Organizations [intent=direct_read availability=implemented]
    read count-organizations - Count Organizations [intent=direct_read availability=implemented]
    read search-organizations - Search Organizations [intent=direct_read availability=implemented]
    read show-many-organizations - Show Many Organizations [intent=direct_read availability=implemented]
    read list-ticket-problems - List Ticket Problems [intent=direct_read availability=implemented]
    read list-queues - List queues [intent=direct_read availability=implemented]
    read show-queue-by-id - Show Queue [intent=direct_read availability=implemented]; flags: --queue_id
    read list-queue-definitions - List Queue Definitions [intent=direct_read availability=implemented]
    read list-support-addresses - List Support Addresses [intent=direct_read availability=implemented]
    read show-support-address - Show Support Address [intent=direct_read availability=implemented]; flags: --support_address_id
    read get-relationship-filter-definitions - Filter Definitions [intent=direct_read availability=implemented]; flags: --target_type, --source_type
    read list-remote-authentications - List Remote Authentications [intent=direct_read availability=implemented]; flags: --brand_id
    read list-requests - List Requests [intent=direct_read availability=implemented]; flags: --sort_by, --sort_order
    read show-request - Show Request [intent=direct_read availability=implemented]; flags: --request_id
    read list-comments - Listing Comments [intent=direct_read availability=implemented]; flags: --request_id, --since, --role
    read show-comment - Getting Comments [intent=direct_read availability=implemented]; flags: --request_id, --ticket_comment_id
    read list-ccd-requests - List CCD Requests [intent=direct_read availability=implemented]; flags: --sort_by, --sort_order
    read list-open-requests - List Open Requests [intent=direct_read availability=implemented]; flags: --sort_by, --sort_order
    read search-requests - Search Requests [intent=direct_read availability=implemented]; flags: --query
    read list-solved-requests - List Solved Requests [intent=direct_read availability=implemented]; flags: --sort_by, --sort_order
    read list-resource-collections - List Resource Collections [intent=direct_read availability=implemented]
    read retrieve-resource-collection - Show Resource Collection [intent=direct_read availability=implemented]; flags: --resource_collection_id
    read list-a-gent-attribute-values - List Agent Attribute Values [intent=direct_read availability=implemented]; flags: --user_id
    read list-many-agents-attribute-values - List Attribute Values for Many Agents [intent=direct_read availability=implemented]; flags: --filter-agent_ids, --page-before, --page-after, --page-size
    read list-account-attributes - List Account Attributes [intent=direct_read availability=implemented]
    read show-attribute - Show Attribute [intent=direct_read availability=implemented]; flags: --attribute_id
    read list-attribute-values - List Attribute Values for an Attribute [intent=direct_read availability=implemented]; flags: --attribute_id
    read show-attribute-value - Show Attribute Value [intent=direct_read availability=implemented]; flags: --attribute_id, --attribute_value_id
    read list-routing-attribute-definitions - List Routing Attribute Definitions [intent=direct_read availability=implemented]
    read list-tickets-fullfilled-by-user - List Tickets Fulfilled by a User [intent=direct_read availability=implemented]; flags: --ticket_ids
    read list-ticket-attribute-values - List Ticket Attribute Values [intent=direct_read availability=implemented]; flags: --ticket_id
    read list-satisfaction-ratings - List Satisfaction Ratings [intent=direct_read availability=implemented]
    read show-satisfaction-rating - Show Satisfaction Rating [intent=direct_read availability=implemented]; flags: --satisfaction_rating_id
    read count-satisfaction-ratings - Count Satisfaction Ratings [intent=direct_read availability=implemented]
    read list-satisfaction-rating-reasons - List Reasons for Satisfaction Rating [intent=direct_read availability=implemented]
    read show-satisfaction-ratings - Show Reason for Satisfaction Rating [intent=direct_read availability=implemented]; flags: --satisfaction_reason_id
    read list-saved-searches - List Saved Searches [intent=direct_read availability=implemented]
    read count-search-results - Show Results Count [intent=direct_read availability=implemented]; flags: --query
    read show-security-settings - Show Security Settings [intent=direct_read availability=implemented]; flags: --brand_id
    read list-sessions - List Sessions [intent=direct_read availability=implemented]
    read list-sharing-agreements - List Sharing Agreements [intent=direct_read availability=implemented]
    read show-sharing-agreement - Show a Sharing Agreement [intent=direct_read availability=implemented]; flags: --sharing_agreement_id
    read list-sla-policies - List SLA Policies [intent=direct_read availability=implemented]
    read show-sla-policy - Show SLA Policy [intent=direct_read availability=implemented]; flags: --sla_policy_id
    read retrieve-sla-policy-filter-definition-items - Retrieve Supported Filter Definition Items [intent=direct_read availability=implemented]
    read list-suspended-tickets - List Suspended Tickets [intent=direct_read availability=implemented]
    read show-suspended-tickets - Show Suspended Ticket [intent=direct_read availability=implemented]; flags: --id
    read list-tags - List Tags [intent=direct_read availability=implemented]
    read count-tags - Count Tags [intent=direct_read availability=implemented]
    read list-target-failures - List Target Failures [intent=direct_read availability=implemented]
    read show-target-failure - Show Target Failure [intent=direct_read availability=implemented]; flags: --target_failure_id
    read list-targets - List Targets [intent=direct_read availability=implemented]
    read show-target - Show Target [intent=direct_read availability=implemented]; flags: --target_id
    read list-task-list-templates - List Task List Templates [intent=direct_read availability=implemented]
    read show-task-list-template - Show Task List Template [intent=direct_read availability=implemented]; flags: --task_list_template_id
    read get-tasks-by-task-list-template-id - Get Tasks by Task List Template Id [intent=direct_read availability=implemented]; flags: --task_list_template_id
    read list-ticket-fields - List Ticket Fields [intent=direct_read availability=implemented]; flags: --locale, --creator
    read show-ticketfield - Show Ticket Field [intent=direct_read availability=implemented]; flags: --ticket_field_id
    read list-ticket-field-options - List Ticket Field Options [intent=direct_read availability=implemented]; flags: --ticket_field_id
    read show-ticket-field-option - Show Ticket Field Option [intent=direct_read availability=implemented]; flags: --ticket_field_id, --ticket_field_option_id
    read count-ticket-fields - Count Ticket Fields [intent=direct_read availability=implemented]
    read show-many-ticket-fields - Show Many Ticket Fields [intent=direct_read availability=implemented]; flags: --ids, --keys, --creator, --exclude_sub_selection_options
    read list-ticket-form-statuses - List Ticket Form Statuses [intent=direct_read availability=implemented]; flags: --ticket_form_id, --filter
    read show-many-ticket-form-statuses - Show Many Ticket Form Statuses [intent=direct_read availability=implemented]; flags: --ids
    read list-ticket-forms - List Ticket Forms [intent=direct_read availability=implemented]; flags: --active, --end_user_visible, --fallback_to_default, --form_type, --associated_to_brand, --locale
    read show-ticket-form - Show Ticket Form [intent=direct_read availability=implemented]; flags: --ticket_form_id
    read ticket-form-ticket-form-statuses - List Ticket Form Statuses of a Ticket Form [intent=direct_read availability=implemented]; flags: --ticket_form_id
    read show-many-ticket-forms - Show Many Ticket Forms [intent=direct_read availability=implemented]; flags: --ids, --active, --end_user_visible, --fallback_to_default, --associated_to_brand
    read show-ticket-metrics - Show Ticket Metrics [intent=direct_read availability=implemented]; flags: --ticket_metric_id
    read list-tickets - List Tickets [intent=direct_read availability=implemented]; flags: --external_id, --start_time
    read show-ticket - Show Ticket [intent=direct_read availability=implemented]; flags: --ticket_id, --reduced_payload_size, --remove_duplicate_fields
    read show-ticket-audit - Show Audit [intent=direct_read availability=implemented]; flags: --ticket_id, --ticket_audit_id
    read count-audits-for-ticket - Count Audits for a Ticket [intent=direct_read availability=implemented]; flags: --ticket_id
    read list-ticket-collaborators - List Collaborators for a Ticket [intent=direct_read availability=implemented]; flags: --ticket_id
    read count-ticket-comments - Count Ticket Comments [intent=direct_read availability=implemented]; flags: --ticket_id
    read list-conversation-log-for-ticket - List Conversation log for Ticket [intent=direct_read availability=implemented]; flags: --ticket_id
    read list-ticket-email-c-cs - List Email CCs for a Ticket [intent=direct_read availability=implemented]; flags: --ticket_id
    read list-ticket-followers - List Followers for a Ticket [intent=direct_read availability=implemented]; flags: --ticket_id
    read list-ticket-incidents - List Ticket Incidents [intent=direct_read availability=implemented]; flags: --ticket_id
    read show-ticket-after-changes - Show Ticket After Changes [intent=direct_read availability=implemented]; flags: --ticket_id, --macro_id, --normalize_comment
    read show-ticket-metrics-by-ticket - Show Ticket Metrics By Ticket [intent=direct_read availability=implemented]; flags: --ticket_id
    read list-resource-tags - List Resource Tags [intent=direct_read availability=implemented]; flags: --ticket_id
    read show-task-list - Show Task List [intent=direct_read availability=implemented]; flags: --ticket_id
    read count-tickets - Count Tickets [intent=direct_read availability=implemented]
    read list-recent-tickets - List Recent Tickets [intent=direct_read availability=implemented]
    read tickets-show-many - Show Multiple Tickets [intent=direct_read availability=implemented]; flags: --include
    read list-trigger-categories - List Ticket Trigger Categories [intent=direct_read availability=implemented]; flags: --page, --sort, --include
    read show-trigger-category-by-id - Show Ticket Trigger Category [intent=direct_read availability=implemented]; flags: --trigger_category_id
    read list-triggers - List Ticket Triggers [intent=direct_read availability=implemented]
    read get-trigger - Show Ticket Trigger [intent=direct_read availability=implemented]; flags: --trigger_id
    read list-trigger-revisions - List Ticket Trigger Revisions [intent=direct_read availability=implemented]; flags: --trigger_id
    read trigger-revision - Show Ticket Trigger Revision [intent=direct_read availability=implemented]; flags: --trigger_id, --trigger_revision_id
    read list-active-triggers - List Active Ticket Triggers [intent=direct_read availability=implemented]
    read list-trigger-action-condition-definitions - List Ticket Trigger Action and Condition Definitions [intent=direct_read availability=implemented]
    read search-triggers - Search Ticket Triggers [intent=direct_read availability=implemented]
    read list-user-fields - List User Fields [intent=direct_read availability=implemented]; flags: --resolve_dc
    read show-user-field - Show User Field [intent=direct_read availability=implemented]; flags: --user_field_id
    read list-user-field-options - List User Field Options [intent=direct_read availability=implemented]; flags: --user_field_id
    read show-user-field-option - Show a User Field Option [intent=direct_read availability=implemented]; flags: --user_field_id, --user_field_option_id
    read show-many-user-fields - Show Many User Fields [intent=direct_read availability=implemented]; flags: --keys
    read list-users - List Users [intent=direct_read availability=implemented]; flags: --brand_id
    read show-user - Show User [intent=direct_read availability=implemented]; flags: --user_id
    read list-user-brand-agents - List Brand Agent Memberships By User [intent=direct_read availability=implemented]; flags: --user_id
    read show-user-brand-agent-by-id - Show Brand Agent Membership By User [intent=direct_read availability=implemented]; flags: --user_id, --brand_agent_id
    read show-user-compliance-deletion-statuses - Show Compliance Deletion Statuses [intent=direct_read availability=implemented]; flags: --user_id, --application
    read get-user-entitlements-full - Get Full User Entitlements [intent=direct_read availability=implemented]; flags: --user_id
    read list-user-group-memberships - List Group Memberships by User [intent=direct_read availability=implemented]; flags: --user_id
    read show-user-group-membership-by-id - Show User's Group Membership [intent=direct_read availability=implemented]; flags: --user_id, --group_membership_id
    read list-user-groups - List User Groups [intent=direct_read availability=implemented]; flags: --user_id
    read count-user-groups - Count User Groups [intent=direct_read availability=implemented]; flags: --user_id
    read list-user-identities - List Identities [intent=direct_read availability=implemented]; flags: --user_id, --type
    read show-user-identity - Show Identity [intent=direct_read availability=implemented]; flags: --user_id, --user_identity_id
    read list-user-organization-memberships - List Organization Memberships by User [intent=direct_read availability=implemented]; flags: --user_id
    read show-organization-membership-by-user-id - Show Organization Membership by User [intent=direct_read availability=implemented]; flags: --user_id, --organization_membership_id
    read list-user-organization-subscriptions - List User's Organization Subscriptions [intent=direct_read availability=implemented]; flags: --user_id
    read list-user-organizations - List User Organizations [intent=direct_read availability=implemented]; flags: --user_id
    read count-user-organizations - Count User's Organizations [intent=direct_read availability=implemented]; flags: --user_id
    read get-user-password-requirements - List password requirements [intent=direct_read availability=implemented]; flags: --user_id
    read show-user-related - Show User Related Information [intent=direct_read availability=implemented]; flags: --user_id
    read list-user-requests - List User Requests [intent=direct_read availability=implemented]; flags: --user_id, --sort_by, --sort_order
    read list-user-sessions - List Sessions for User [intent=direct_read availability=implemented]; flags: --user_id
    read show-session - Show Session [intent=direct_read availability=implemented]; flags: --user_id, --session_id
    read list-user-tags - List User Tags [intent=direct_read availability=implemented]; flags: --user_id
    read list-user-assigned-tickets - List User Assigned Tickets [intent=direct_read availability=implemented]; flags: --user_id
    read count-user-assigned-tickets - Count User Assigned Tickets [intent=direct_read availability=implemented]; flags: --user_id
    read list-user-ccd-tickets - List User CCD Tickets [intent=direct_read availability=implemented]; flags: --user_id
    read count-user-ccd-tickets - Count User CCD Tickets [intent=direct_read availability=implemented]; flags: --user_id
    read list-user-followed-tickets - List User Followed Tickets [intent=direct_read availability=implemented]; flags: --user_id, --exclude_archived
    read list-user-requested-tickets - List User Requested Tickets [intent=direct_read availability=implemented]; flags: --user_id, --exclude_archived, --exclude_count
    read autocomplete-users - Autocomplete Users [intent=direct_read availability=implemented]; flags: --name, --phone, --filter, --per_page, --brand_id
    read count-users - Count Users [intent=direct_read availability=implemented]; flags: --brand_id
    read show-current-user - Show Self [intent=direct_read availability=implemented]
    read list-current-user-o-auth-clients - List Current User's Clients [intent=direct_read availability=implemented]
    read show-currently-authenticated-session - Show the Currently Authenticated Session [intent=direct_read availability=implemented]
    read renew-current-session - Renew the current session [intent=direct_read availability=implemented]
    read show-current-user-settings - Show Current User Settings [intent=direct_read availability=implemented]
    read search-users - Search Users [intent=direct_read availability=implemented]; flags: --query, --external_id, --brand_id
    read show-many-users - Show Many Users [intent=direct_read availability=implemented]; flags: --ids, --external_ids, --include_deleted, --brand_id
    read list-views - List Views [intent=direct_read availability=implemented]; flags: --access, --active, --group_id, --sort, --sort_by, --sort_order
    read show-view - Show View [intent=direct_read availability=implemented]; flags: --view_id
    read get-view-count - Count Tickets in View [intent=direct_read availability=implemented]; flags: --view_id
    read list-tickets-from-view - List Tickets From a View [intent=direct_read availability=implemented]; flags: --view_id, --sort_by, --sort_order
    read list-active-views - List Active Views [intent=direct_read availability=implemented]; flags: --access, --group_id, --sort_by, --sort_order
    read list-compact-views - List Views - Compact [intent=direct_read availability=implemented]
    read count-views - Count Views [intent=direct_read availability=implemented]
    read get-view-counts - Count Tickets in Views [intent=direct_read availability=implemented]; flags: --ids
    read list-view-definitions - List View Filter Definitions [intent=direct_read availability=implemented]
    read search-views - Search Views [intent=direct_read availability=implemented]; flags: --query, --access, --active, --group_id, --sort_by, --sort_order, --include
    read list-views-by-id - List Views By ID [intent=direct_read availability=implemented]; flags: --ids, --active
    read list-workspaces - List Workspaces [intent=direct_read availability=implemented]
    read show-workspace - Show Workspace [intent=direct_read availability=implemented]; flags: --workspace_id
    binary candidates - Review Zendesk binary/file read candidates from the official OAS [intent=direct_read availability=planned]; notes: Metadata inventory only. api_surface.json tracks 37 binary/file read candidates; #162 must define bounded size/path/download policy before execution.
  Write Candidates
    write candidates - Review Zendesk sensitive/admin write candidates from the official OAS [intent=reverse_etl availability=planned]; approval: reverse ETL execution remains plan → preview → approval → execute.; risk: Zendesk mutation candidates remain blocked until typed reverse-ETL schemas, risk text, approval text, redaction, and policy gates are added.; notes: Metadata inventory only. api_surface.json tracks 210 sensitive/admin reverse-ETL candidates.
    destructive candidates - Review Zendesk destructive write candidates from the official OAS [intent=reverse_etl availability=planned]; approval: reverse ETL execution remains plan → preview → approval → execute with typed confirmation for destructive actions.; risk: Zendesk DELETE operations are destructive and remain blocked until typed reverse-ETL schemas, risk text, approval text, redaction, and destructive confirmation are added.; notes: Metadata inventory only. api_surface.json tracks 85 destructive-action candidates.
  Blocked Metadata
    deprecated operations - Review deprecated Zendesk operations from the official OAS [intent=docs_only availability=unsafe_or_disallowed]; notes: Metadata inventory only. api_surface.json tracks 3 deprecated operation rows; #160 must confirm replacements or blockers.
  Help topics:
    zendesk-auth - Zendesk authentication uses OAuth bearer tokens or email/API-token Basic auth with secrets stored in credentials.
    zendesk-safety - Zendesk writes remain reverse-ETL gated: plan, preview, approval, execute, with destructive confirmation where required.
    zendesk-operation-ledger - The initial operation ledger blocks all official OAS operations until later lanes map exact executable surfaces.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect zendesk

  # Inspect as structured JSON
  pm connectors inspect zendesk --json

AGENT WORKFLOW
  - Run pm connectors inspect zendesk before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
