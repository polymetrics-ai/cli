# Overview

Reads and writes documented Salesforce Account Engagement (Pardot) API v5 JSON resources.

Readable streams: `prospects`, `prospect`, `campaigns`, `campaign`, `lists`, `list`, `users`,
`user`, `custom_fields`, `custom_field`, `custom_redirects`, `custom_redirect`, `dynamic_contents`,
`dynamic_content`, `emails`, `email`, `email_templates`, `email_template`, `list_emails`,
`list_email`, `list_email_stats`, `files`, `file`, `folders`, `folder`, `folder_contents`,
`folder_content`, `forms`, `form`, `form_fields`, `form_field`, `form_handlers`, `form_handler`,
`form_handler_fields`, `form_handler_field`, `landing_pages`, `landing_page`, `layout_templates`,
`layout_template`, `list_memberships`, `list_membership`, `opportunities`, `opportunity`,
`prospect_accounts`, `prospect_account`, `tags`, `tag`, `tagged_objects`, `tagged_object`,
`tracker_domains`, `tracker_domain`, `visitors`, `visitor`, `visits`, `visit`, `visitor_activities`,
`visitor_activity`, `visitor_page_views`, `visitor_page_view`, `engagement_studio_programs`,
`engagement_studio_program`, `lifecycle_stages`, `lifecycle_stage`, `lifecycle_histories`,
`lifecycle_history`, `account`, `bulk_actions`, `bulk_action`, `imports`, `import_job`,
`external_activities`, `external_activity`.

Write actions: `create_prospect`, `update_prospect`, `delete_prospect`,
`upsert_prospect_latest_by_email`, `undelete_prospect`, `add_tag_to_prospect`,
`remove_tag_from_prospect`, `connect_salesforce_campaign`, `add_tag_to_campaign`,
`remove_tag_from_campaign`, `create_list`, `update_list`, `delete_list`, `add_tag_to_list`,
`remove_tag_from_list`, `add_tag_to_user`, `remove_tag_from_user`, `create_custom_field`,
`update_custom_field`, `delete_custom_field`, `add_tag_to_custom_field`,
`remove_tag_from_custom_field`, `create_custom_redirect`, `update_custom_redirect`,
`delete_custom_redirect`, `add_tag_to_custom_redirect`, `remove_tag_from_custom_redirect`,
`create_dynamic_content`, `add_tag_to_dynamic_content`, `remove_tag_from_dynamic_content`,
`create_email`, `add_tag_to_email`, `remove_tag_from_email`, `copy_email_to_cms`,
`create_email_template`, `update_email_template`, `delete_email_template`,
`add_tag_to_email_template`, `remove_tag_from_email_template`, `create_list_email`, `update_file`,
`delete_file`, `add_tag_to_file`, `remove_tag_from_file`, `copy_file_to_cms`, `create_form`,
`delete_form`, `undelete_form`, `reorder_form_fields`, `add_tag_to_form`, `remove_tag_from_form`,
`copy_form_to_cms`, `create_form_field`, `add_form_field_dependent`, `add_form_field_progressive`,
`add_form_field_value`, `reorder_form_field_values`, `add_tag_to_form_field`,
`remove_tag_from_form_field`, `create_form_handler`, `update_form_handler`, `delete_form_handler`,
`add_tag_to_form_handler`, `remove_tag_from_form_handler`, `create_form_handler_field`,
`update_form_handler_field`, `delete_form_handler_field`, `create_landing_page`,
`add_tag_to_landing_page`, `remove_tag_from_landing_page`, `copy_landing_page_to_cms`,
`create_layout_template`, `update_layout_template`, `delete_layout_template`,
`add_tag_to_layout_template`, `remove_tag_from_layout_template`, `create_list_membership`,
`update_list_membership`, `delete_list_membership`, `add_tag_to_opportunity`, and 12 more.

Service API documentation: https://developer.salesforce.com/docs/marketing/pardot/overview.

## Auth setup

Connection fields:

- `access_token` (required, secret, string); Pre-issued Salesforce OAuth access token; sent as a
  Bearer token. Never logged. Token acquisition/refresh is intentionally outside this connector.
- `base_url` (optional, string); default `https://pi.pardot.com`; format `uri`; Pardot API base URL
  override for tests, sandboxes, or proxies.
- `business_unit_id` (required, string); Pardot Business Unit Id, sent as the
  Pardot-Business-Unit-Id header on every request.
- `id` (optional, integer/string); Record id used by detail streams.

Secret fields are redacted in logs and write previews: `access_token`.

Default configuration values: `base_url=https://pi.pardot.com`.

Authentication behavior:

- Bearer token authentication using `secrets.access_token`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/v5/objects/prospects` with query `fields`=`id,email`; `limit`=`1`.

## Streams notes

Default pagination: follows a next-page URL from the response body; URL path `nextPageUrl`; next
URLs stay on the configured API host.

Pagination by stream: next_url: `prospects`, `campaigns`, `lists`, `users`, `custom_fields`,
`custom_redirects`, `dynamic_contents`, `emails`, `email_templates`, `list_emails`, `files`,
`folders`, `folder_contents`, `forms`, `form_fields`, `form_handlers`, `form_handler_fields`,
`landing_pages`, `layout_templates`, `list_memberships`, `opportunities`, `prospect_accounts`,
`tags`, `tagged_objects`, `tracker_domains`, `visitors`, `visits`, `visitor_activities`,
`visitor_page_views`, `engagement_studio_programs`, `lifecycle_stages`, `lifecycle_histories`,
`bulk_actions`, `imports`, `external_activities`; none: `prospect`, `campaign`, `list`, `user`,
`custom_field`, `custom_redirect`, `dynamic_content`, `email`, `email_template`, `list_email`,
`list_email_stats`, `file`, `folder`, `folder_content`, `form`, `form_field`, `form_handler`,
`form_handler_field`, `landing_page`, `layout_template`, `list_membership`, `opportunity`,
`prospect_account`, `tag`, `tagged_object`, `tracker_domain`, `visitor`, `visit`,
`visitor_activity`, `visitor_page_view`, `engagement_studio_program`, `lifecycle_stage`,
`lifecycle_history`, `account`, `bulk_action`, `import_job`, `external_activity`.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `prospects`: GET `/api/v5/objects/prospects` - records path `values`; query
  `fields`=`id,email,firstName,lastName,createdAt,updatedAt`; `limit`=`200`; follows a next-page URL
  from the response body; URL path `nextPageUrl`; next URLs stay on the configured API host;
  incremental cursor `updatedAt`; formatted as `rfc3339`.
- `prospect`: GET `/api/v5/objects/prospects/{{ config.id }}` - records path `.`; query
  `fields`=`id,email,firstName,lastName,createdAt,updatedAt`; emits passthrough records.
- `campaigns`: GET `/api/v5/objects/campaigns` - records path `values`; query
  `fields`=`id,name,createdAt,updatedAt`; `limit`=`200`; follows a next-page URL from the response
  body; URL path `nextPageUrl`; next URLs stay on the configured API host; incremental cursor
  `updatedAt`; formatted as `rfc3339`.
- `campaign`: GET `/api/v5/objects/campaigns/{{ config.id }}` - records path `.`; query
  `fields`=`id,name,createdAt,updatedAt`; emits passthrough records.
- `lists`: GET `/api/v5/objects/lists` - records path `values`; query
  `fields`=`id,name,createdAt,updatedAt`; `limit`=`200`; follows a next-page URL from the response
  body; URL path `nextPageUrl`; next URLs stay on the configured API host; incremental cursor
  `updatedAt`; formatted as `rfc3339`.
- `list`: GET `/api/v5/objects/lists/{{ config.id }}` - records path `.`; query
  `fields`=`id,name,createdAt,updatedAt`; emits passthrough records.
- `users`: GET `/api/v5/objects/users` - records path `values`; query
  `fields`=`id,email,firstName,lastName,username,jobTitle,role,roleName,salesforceId,isDeleted,createdAt,updatedAt,createdById,updatedById,tagReplacementLanguage`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `user`: GET `/api/v5/objects/users/{{ config.id }}` - records path `.`; query
  `fields`=`id,email,firstName,lastName,username,jobTitle,role,roleName,salesforceId,isDeleted,createdAt,updatedAt,createdById,updatedById,tagReplacementLanguage`;
  emits passthrough records.
- `custom_fields`: GET `/api/v5/objects/custom-fields` - records path `values`; query
  `fields`=`id,name,fieldId,type,isRecordMultipleResponses,salesforceId,isUseValues,isRequired,isAnalyticsSynced,createdAt,updatedAt,createdById,updatedById`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `custom_field`: GET `/api/v5/objects/custom-fields/{{ config.id }}` - records path `.`; query
  `fields`=`id,name,fieldId,type,isRecordMultipleResponses,salesforceId,isUseValues,isRequired,isAnalyticsSynced,createdAt,updatedAt,createdById,updatedById`;
  emits passthrough records.
- `custom_redirects`: GET `/api/v5/objects/custom-redirects` - records path `values`; query
  `fields`=`id,name,campaignId,destinationUrl,folderId,trackerDomainId,createdAt,updatedAt,createdById,updatedById,isDeleted`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `custom_redirect`: GET `/api/v5/objects/custom-redirects/{{ config.id }}` - records path `.`;
  query
  `fields`=`id,name,campaignId,destinationUrl,folderId,trackerDomainId,createdAt,updatedAt,createdById,updatedById,isDeleted`;
  emits passthrough records.
- `dynamic_contents`: GET `/api/v5/objects/dynamic-contents` - records path `values`; query
  `fields`=`id,name,basedOnProspectApiFieldId,baseContent,folderId,trackerDomainId,embedCode,embedUrl,basedOn,tagReplacementLanguage,createdAt,updatedAt,isDeleted,createdById,updatedById`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `dynamic_content`: GET `/api/v5/objects/dynamic-contents/{{ config.id }}` - records path `.`;
  query
  `fields`=`id,name,basedOnProspectApiFieldId,baseContent,folderId,trackerDomainId,embedCode,embedUrl,basedOn,tagReplacementLanguage,createdAt,updatedAt,isDeleted,createdById,updatedById`;
  emits passthrough records.
- `emails`: GET `/api/v5/objects/emails` - records path `values`; query
  `fields`=`id,name,campaignId,prospectId,subject,emailTemplateId,trackerDomainId,folderId,clientType,createdById,listEmailId,sentAt,type,salesforceCmsId`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `email`: GET `/api/v5/objects/emails/{{ config.id }}` - records path `.`; query
  `fields`=`id,name,campaignId,prospectId,subject,emailTemplateId,trackerDomainId,folderId,clientType,createdById,listEmailId,sentAt,type,salesforceCmsId`;
  emits passthrough records.
- `email_templates`: GET `/api/v5/objects/email-templates` - records path `values`; query
  `fields`=`id,name,subject,isOneToOneEmail,isAutoResponderEmail,isDripEmail,isListEmail,trackerDomainId,folderId,type,campaignId,isDeleted,createdAt,updatedAt,createdById,updatedById,tagReplacementLanguage`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `email_template`: GET `/api/v5/objects/email-templates/{{ config.id }}` - records path `.`; query
  `fields`=`id,name,subject,isOneToOneEmail,isAutoResponderEmail,isDripEmail,isListEmail,trackerDomainId,folderId,type,campaignId,isDeleted,createdAt,updatedAt,createdById,updatedById,tagReplacementLanguage`;
  emits passthrough records.
- `list_emails`: GET `/api/v5/objects/list-emails` - records path `values`; query
  `fields`=`id,name,campaignId,subject,emailTemplateId,trackerDomainId,folderId,isPaused,isSent,isDeleted,clientType,createdById,updatedById,createdAt,updatedAt,sentAt,isOperational,type`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `list_email`: GET `/api/v5/objects/list-emails/{{ config.id }}` - records path `.`; query
  `fields`=`id,name,campaignId,subject,emailTemplateId,trackerDomainId,folderId,isPaused,isSent,isDeleted,clientType,createdById,updatedById,createdAt,updatedAt,sentAt,isOperational,type`;
  emits passthrough records.
- `list_email_stats`: GET `/api/v5/objects/list-emails/{{ config.id }}/stats` - records path `.`;
  query
  `fields`=`id,name,campaignId,subject,emailTemplateId,trackerDomainId,folderId,isPaused,isSent,isDeleted,clientType,createdById,updatedById,createdAt,updatedAt,sentAt,isOperational,type`;
  emits passthrough records.
- `files`: GET `/api/v5/objects/files` - records path `values`; query
  `fields`=`id,name,folderId,campaignId,vanityUrlPath,trackerDomainId,salesforceId,url,size,bitlyIsPersonalized,bitlyShortUrl,vanityUrl,isTracked,createdAt,updatedAt,createdById,updatedById,salesforceCmsId`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `file`: GET `/api/v5/objects/files/{{ config.id }}` - records path `.`; query
  `fields`=`id,name,folderId,campaignId,vanityUrlPath,trackerDomainId,salesforceId,url,size,bitlyIsPersonalized,bitlyShortUrl,vanityUrl,isTracked,createdAt,updatedAt,createdById,updatedById,salesforceCmsId`;
  emits passthrough records.
- `folders`: GET `/api/v5/objects/folders` - records path `values`; query
  `fields`=`id,name,parentFolderId,path,usePermissions,createdAt,updatedAt,createdById,updatedById`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `folder`: GET `/api/v5/objects/folders/{{ config.id }}` - records path `.`; query
  `fields`=`id,name,parentFolderId,path,usePermissions,createdAt,updatedAt,createdById,updatedById`;
  emits passthrough records.
- `folder_contents`: GET `/api/v5/objects/folder-contents` - records path `values`; query
  `fields`=`id,folderId,folderRef,objectType,objectId,objectName,objectRef,createdAt,updatedAt,createdById,updatedById`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `folder_content`: GET `/api/v5/objects/folder-contents/{{ config.id }}` - records path `.`; query
  `fields`=`id,folderId,folderRef,objectType,objectId,objectName,objectRef,createdAt,updatedAt,createdById,updatedById`;
  emits passthrough records.
- `forms`: GET `/api/v5/objects/forms` - records path `values`; query
  `fields`=`id,name,campaignId,layoutTemplateId,folderId,trackerDomainId,createdAt,updatedAt,createdById,updatedById,embedCode,isDeleted,isUseRedirectLocation,salesforceCmsId,salesforceId,url`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `form`: GET `/api/v5/objects/forms/{{ config.id }}` - records path `.`; query
  `fields`=`id,name,campaignId,layoutTemplateId,folderId,trackerDomainId,createdAt,updatedAt,createdById,updatedById,embedCode,isDeleted,isUseRedirectLocation,salesforceCmsId,salesforceId,url`;
  emits passthrough records.
- `form_fields`: GET `/api/v5/objects/form-fields` - records path `values`; query
  `fields`=`id,formId,prospectApiFieldId,type,dataFormat,label,description,errorMessage,cssClasses,isRequired,isAlwaysDisplay,isMaintainInitialValue,isDoNotPrefill,sortOrder,hasDependents,hasProgressives,hasValues,createdById,updatedById,createdAt,updatedAt`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `form_field`: GET `/api/v5/objects/form-fields/{{ config.id }}` - records path `.`; query
  `fields`=`id,formId,prospectApiFieldId,type,dataFormat,label,description,errorMessage,cssClasses,isRequired,isAlwaysDisplay,isMaintainInitialValue,isDoNotPrefill,sortOrder,hasDependents,hasProgressives,hasValues,createdById,updatedById,createdAt,updatedAt`;
  emits passthrough records.
- `form_handlers`: GET `/api/v5/objects/form-handlers` - records path `values`; query
  `fields`=`id,name,folderId,campaignId,trackerDomainId,isDataForwarded,isCookieless,salesforceId,embedCode,createdAt,createdById,updatedAt,updatedById,isDeleted`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `form_handler`: GET `/api/v5/objects/form-handlers/{{ config.id }}` - records path `.`; query
  `fields`=`id,name,folderId,campaignId,trackerDomainId,isDataForwarded,isCookieless,salesforceId,embedCode,createdAt,createdById,updatedAt,updatedById,isDeleted`;
  emits passthrough records.
- `form_handler_fields`: GET `/api/v5/objects/form-handler-fields` - records path `values`; query
  `fields`=`id,name,formHandlerId,prospectApiFieldId,fieldLabel,dataFormat,isRequired,createdAt,updatedAt,createdById,updatedById`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `form_handler_field`: GET `/api/v5/objects/form-handler-fields/{{ config.id }}` - records path
  `.`; query
  `fields`=`id,name,formHandlerId,prospectApiFieldId,fieldLabel,dataFormat,isRequired,createdAt,updatedAt,createdById,updatedById`;
  emits passthrough records.
- `landing_pages`: GET `/api/v5/objects/landing-pages` - records path `values`; query
  `fields`=`id,name,campaignId,folderId,formId,layoutTemplateId,title,description,isDoNotIndex,vanityUrlPath,redirectLocation,trackerDomainId,archiveDate,createdAt,updatedAt,createdById,updatedById,isDeleted,salesforceCmsId`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `landing_page`: GET `/api/v5/objects/landing-pages/{{ config.id }}` - records path `.`; query
  `fields`=`id,name,campaignId,folderId,formId,layoutTemplateId,title,description,isDoNotIndex,vanityUrlPath,redirectLocation,trackerDomainId,archiveDate,createdAt,updatedAt,createdById,updatedById,isDeleted,salesforceCmsId`;
  emits passthrough records.
- `layout_templates`: GET `/api/v5/objects/layout-templates` - records path `values`; query
  `fields`=`id,name,layoutContent,formContent,isIncludeDefaultCss,folderId,isDeleted,createdAt,updatedAt,createdById,updatedById,siteSearchContent`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `layout_template`: GET `/api/v5/objects/layout-templates/{{ config.id }}` - records path `.`;
  query
  `fields`=`id,name,layoutContent,formContent,isIncludeDefaultCss,folderId,isDeleted,createdAt,updatedAt,createdById,updatedById,siteSearchContent`;
  emits passthrough records.
- `list_memberships`: GET `/api/v5/objects/list-memberships` - records path `values`; query
  `fields`=`id,createdAt,updatedAt,createdById,updatedById,isDeleted,optedOut,listId,prospectId`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `list_membership`: GET `/api/v5/objects/list-memberships/{{ config.id }}` - records path `.`;
  query
  `fields`=`id,createdAt,updatedAt,createdById,updatedById,isDeleted,optedOut,listId,prospectId`;
  emits passthrough records.
- `opportunities`: GET `/api/v5/objects/opportunities` - records path `values`; query
  `fields`=`id,name,campaignId,closedAt,createdAt,createdById,updatedAt,updatedById,value,probability,stage,status,isDeleted`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `opportunity`: GET `/api/v5/objects/opportunities/{{ config.id }}` - records path `.`; query
  `fields`=`id,name,campaignId,closedAt,createdAt,createdById,updatedAt,updatedById,value,probability,stage,status,isDeleted`;
  emits passthrough records.
- `prospect_accounts`: GET `/api/v5/objects/prospect-accounts` - records path `values`; query
  `fields`=`id,name,salesforceId,isDeleted,annualRevenue,billingCity,billingCountry,billingState,billingZip,phone,website,createdAt,updatedAt,createdById,updatedById,assignedToId`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `prospect_account`: GET `/api/v5/objects/prospect-accounts/{{ config.id }}` - records path `.`;
  query
  `fields`=`id,name,salesforceId,isDeleted,annualRevenue,billingCity,billingCountry,billingState,billingZip,phone,website,createdAt,updatedAt,createdById,updatedById,assignedToId`;
  emits passthrough records.
- `tags`: GET `/api/v5/objects/tags` - records path `values`; query
  `fields`=`id,name,objectCount,createdById,updatedById,createdAt,updatedAt`; `limit`=`200`; follows
  a next-page URL from the response body; URL path `nextPageUrl`; next URLs stay on the configured
  API host; emits passthrough records.
- `tag`: GET `/api/v5/objects/tags/{{ config.id }}` - records path `.`; query
  `fields`=`id,name,objectCount,createdById,updatedById,createdAt,updatedAt`; emits passthrough
  records.
- `tagged_objects`: GET `/api/v5/objects/tagged-objects` - records path `values`; query
  `fields`=`id,tagId,objectType,objectId,objectName,createdAt,createdById`; `limit`=`200`; follows a
  next-page URL from the response body; URL path `nextPageUrl`; next URLs stay on the configured API
  host; emits passthrough records.
- `tagged_object`: GET `/api/v5/objects/tagged-objects/{{ config.id }}` - records path `.`; query
  `fields`=`id,tagId,objectType,objectId,objectName,createdAt,createdById`; emits passthrough
  records.
- `tracker_domains`: GET `/api/v5/objects/tracker-domains` - records path `values`; query
  `fields`=`id,domain,isPrimary,isDeleted,defaultCampaignId,httpsStatus,sslStatus,validationStatus,validatedAt,createdAt,updatedAt,createdById,updatedById`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `tracker_domain`: GET `/api/v5/objects/tracker-domains/{{ config.id }}` - records path `.`; query
  `fields`=`id,domain,isPrimary,isDeleted,defaultCampaignId,httpsStatus,sslStatus,validationStatus,validatedAt,createdAt,updatedAt,createdById,updatedById`;
  emits passthrough records.
- `visitors`: GET `/api/v5/objects/visitors` - records path `values`; query
  `fields`=`id,campaignParameter,contentParameter,createdAt,doNotSell,hostname,ipAddress,mediumParameter,pageViewCount,prospectId,campaignId,sourceParameter,termParameter,updatedAt,isIdentified`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `visitor`: GET `/api/v5/objects/visitors/{{ config.id }}` - records path `.`; query
  `fields`=`id,campaignParameter,contentParameter,createdAt,doNotSell,hostname,ipAddress,mediumParameter,pageViewCount,prospectId,campaignId,sourceParameter,termParameter,updatedAt,isIdentified`;
  emits passthrough records.
- `visits`: GET `/api/v5/objects/visits` - records path `values`; query
  `fields`=`id,visitorId,prospectId,visitorPageViewCount,firstVisitorPageViewAt,lastVisitorPageViewAt,durationInSeconds,campaignParameter,mediumParameter,sourceParameter,contentParameter,termParameter,createdAt,updatedAt`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `visit`: GET `/api/v5/objects/visits/{{ config.id }}` - records path `.`; query
  `fields`=`id,visitorId,prospectId,visitorPageViewCount,firstVisitorPageViewAt,lastVisitorPageViewAt,durationInSeconds,campaignParameter,mediumParameter,sourceParameter,contentParameter,termParameter,createdAt,updatedAt`;
  emits passthrough records.
- `visitor_activities`: GET `/api/v5/objects/visitor-activities` - records path `values`; query
  `fields`=`id,campaignId,createdAt,customRedirectId,details,emailId,emailTemplateId,fileId,formHandlerId,formId,landingPageId,listEmailId,opportunityId,prospectId,typeName,type,updatedAt,visitId,visitorId,visitorPageViewId`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `visitor_activity`: GET `/api/v5/objects/visitor-activities/{{ config.id }}` - records path `.`;
  query
  `fields`=`id,campaignId,createdAt,customRedirectId,details,emailId,emailTemplateId,fileId,formHandlerId,formId,landingPageId,listEmailId,opportunityId,prospectId,typeName,type,updatedAt,visitId,visitorId,visitorPageViewId`;
  emits passthrough records.
- `visitor_page_views`: GET `/api/v5/objects/visitor-page-views` - records path `values`; query
  `fields`=`id,url,title,visitorId,campaignId,visitId,durationInSeconds,salesforceId,createdAt`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `visitor_page_view`: GET `/api/v5/objects/visitor-page-views/{{ config.id }}` - records path `.`;
  query
  `fields`=`id,url,title,visitorId,campaignId,visitId,durationInSeconds,salesforceId,createdAt`;
  emits passthrough records.
- `engagement_studio_programs`: GET `/api/v5/objects/engagement-studio-programs` - records path
  `values`; query
  `fields`=`id,name,folderId,status,isDeleted,salesforceId,description,recipientListIds,suppressionListIds,createdAt,updatedAt,createdById,updatedById`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `engagement_studio_program`: GET `/api/v5/objects/engagement-studio-programs/{{ config.id }}` -
  records path `.`; query
  `fields`=`id,name,folderId,status,isDeleted,salesforceId,description,recipientListIds,suppressionListIds,createdAt,updatedAt,createdById,updatedById`;
  emits passthrough records.
- `lifecycle_stages`: GET `/api/v5/objects/lifecycle-stages` - records path `values`; query
  `fields`=`id,name,matchType,isLocked,position,isDeleted,createdAt,updatedAt`; `limit`=`200`;
  follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs stay on the
  configured API host; emits passthrough records.
- `lifecycle_stage`: GET `/api/v5/objects/lifecycle-stages/{{ config.id }}` - records path `.`;
  query `fields`=`id,name,matchType,isLocked,position,isDeleted,createdAt,updatedAt`; emits
  passthrough records.
- `lifecycle_histories`: GET `/api/v5/objects/lifecycle-histories` - records path `values`; query
  `fields`=`id,prospectId,lifecycleStageId,createdAt,updatedAt`; `limit`=`200`; follows a next-page
  URL from the response body; URL path `nextPageUrl`; next URLs stay on the configured API host;
  emits passthrough records.
- `lifecycle_history`: GET `/api/v5/objects/lifecycle-histories/{{ config.id }}` - records path `.`;
  query `fields`=`id,prospectId,lifecycleStageId,createdAt,updatedAt`; emits passthrough records.
- `account`: GET `/api/v5/objects/account` - records path `.`; query
  `fields`=`id,company,level,website,pluginCampaignId,addressOne,addressTwo,city,state,zip,territory,country,phone,fax,adminId,createdAt,updatedAt,maximumDailyApiCalls,apiCallsUsed,createdById,updatedById`;
  emits passthrough records.
- `bulk_actions`: GET `/api/v5/bulk-actions` - records path `values`; query
  `fields`=`id,object,bulkAction,status,origin,count,percentComplete,createdAt,updatedAt,createdById,updatedById,errorCount,errorsRef,fileName,processedCount`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `bulk_action`: GET `/api/v5/bulk-actions/{{ config.id }}` - records path `.`; query
  `fields`=`id,object,bulkAction,status,origin,count,percentComplete,createdAt,updatedAt,createdById,updatedById,errorCount,errorsRef,fileName,processedCount`;
  emits passthrough records.
- `imports`: GET `/api/v5/imports` - records path `values`; query
  `fields`=`id,operation,object,status,isExpired,batchesRef,createdAt,updatedAt,createdById,updatedById,createdCount,updatedCount,errorCount,errorRef`;
  `limit`=`200`; follows a next-page URL from the response body; URL path `nextPageUrl`; next URLs
  stay on the configured API host; emits passthrough records.
- `import_job`: GET `/api/v5/imports/{{ config.id }}` - records path `.`; query
  `fields`=`id,operation,object,status,isExpired,batchesRef,createdAt,updatedAt,createdById,updatedById,createdCount,updatedCount,errorCount,errorRef`;
  emits passthrough records.
- `external_activities`: GET `/api/v5/external-activities` - records path `values`; query
  `fields`=`id,extension,type,email,value,activityDate,createdAt,updatedAt`; `limit`=`200`; follows
  a next-page URL from the response body; URL path `nextPageUrl`; next URLs stay on the configured
  API host; emits passthrough records.
- `external_activity`: GET `/api/v5/external-activities/{{ config.id }}` - records path `.`; query
  `fields`=`id,extension,type,email,value,activityDate,createdAt,updatedAt`; emits passthrough
  records.

## Write actions & risks

Overall write risk: creates, updates, sends, tags, copies, restores, cancels, and deletes Salesforce
Account Engagement records through documented API v5 mutation endpoints.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_prospect`: POST `/api/v5/objects/prospects` - kind `create`; body type `json`; risk: POST
  /api/v5/objects/prospects in Salesforce Account Engagement.
- `update_prospect`: PATCH `/api/v5/objects/prospects/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: PATCH
  /api/v5/objects/prospects/{{ record.id }} in Salesforce Account Engagement.
- `delete_prospect`: DELETE `/api/v5/objects/prospects/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Deletes Salesforce Account
  Engagement prospect records.
- `upsert_prospect_latest_by_email`: POST `/api/v5/objects/prospects/do/upsertLatestByEmail` - kind
  `update`; body type `json`; risk: Invokes Pardot upsertLatestByEmail action for prospect.
- `undelete_prospect`: POST `/api/v5/objects/prospects/do/undelete` - kind `update`; body type
  `json`; risk: Invokes Pardot undelete action for prospect.
- `add_tag_to_prospect`: POST `/api/v5/objects/prospects/{{ record.id }}/do/addTag` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk:
  Invokes Pardot addTag action for prospect.
- `remove_tag_from_prospect`: POST `/api/v5/objects/prospects/{{ record.id }}/do/removeTag` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Invokes
  Pardot removeTag action for prospect.
- `connect_salesforce_campaign`: POST `/api/v5/objects/campaigns/{{ record.id
  }}/do/connectSalesforceCampaign` - kind `update`; body type `json`; path fields `id`; required
  record fields `id`; accepted fields `id`; risk: Invokes Pardot connectSalesforceCampaign action
  for campaign.
- `add_tag_to_campaign`: POST `/api/v5/objects/campaigns/{{ record.id }}/do/addTag` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk:
  Invokes Pardot addTag action for campaign.
- `remove_tag_from_campaign`: POST `/api/v5/objects/campaigns/{{ record.id }}/do/removeTag` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Invokes
  Pardot removeTag action for campaign.
- `create_list`: POST `/api/v5/objects/lists` - kind `create`; body type `json`; risk: POST
  /api/v5/objects/lists in Salesforce Account Engagement.
- `update_list`: PATCH `/api/v5/objects/lists/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: PATCH
  /api/v5/objects/lists/{{ record.id }} in Salesforce Account Engagement.
- `delete_list`: DELETE `/api/v5/objects/lists/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Deletes Salesforce Account Engagement
  list records.
- `add_tag_to_list`: POST `/api/v5/objects/lists/{{ record.id }}/do/addTag` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: Invokes
  Pardot addTag action for list.
- `remove_tag_from_list`: POST `/api/v5/objects/lists/{{ record.id }}/do/removeTag` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Invokes Pardot
  removeTag action for list.
- `add_tag_to_user`: POST `/api/v5/objects/users/{{ record.id }}/do/addTag` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: Invokes
  Pardot addTag action for user.
- `remove_tag_from_user`: POST `/api/v5/objects/users/{{ record.id }}/do/removeTag` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Invokes Pardot
  removeTag action for user.
- `create_custom_field`: POST `/api/v5/objects/custom-fields` - kind `create`; body type `json`;
  risk: POST /api/v5/objects/custom-fields in Salesforce Account Engagement.
- `update_custom_field`: PATCH `/api/v5/objects/custom-fields/{{ record.id }}` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: PATCH
  /api/v5/objects/custom-fields/{{ record.id }} in Salesforce Account Engagement.
- `delete_custom_field`: DELETE `/api/v5/objects/custom-fields/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Deletes Salesforce
  Account Engagement custom_field records.
- `add_tag_to_custom_field`: POST `/api/v5/objects/custom-fields/{{ record.id }}/do/addTag` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: Invokes Pardot addTag action for custom_field.
- `remove_tag_from_custom_field`: POST `/api/v5/objects/custom-fields/{{ record.id }}/do/removeTag`
  - kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  Invokes Pardot removeTag action for custom_field.
- `create_custom_redirect`: POST `/api/v5/objects/custom-redirects` - kind `create`; body type
  `json`; risk: POST /api/v5/objects/custom-redirects in Salesforce Account Engagement.
- `update_custom_redirect`: PATCH `/api/v5/objects/custom-redirects/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: PATCH /api/v5/objects/custom-redirects/{{ record.id }} in Salesforce Account Engagement.
- `delete_custom_redirect`: DELETE `/api/v5/objects/custom-redirects/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  Salesforce Account Engagement custom_redirect records.
- `add_tag_to_custom_redirect`: POST `/api/v5/objects/custom-redirects/{{ record.id }}/do/addTag` -
  kind `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: Invokes Pardot addTag action for custom_redirect.
- `remove_tag_from_custom_redirect`: POST `/api/v5/objects/custom-redirects/{{ record.id
  }}/do/removeTag` - kind `delete`; body type `none`; path fields `id`; required record fields `id`;
  accepted fields `id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Invokes Pardot removeTag action for custom_redirect.
- `create_dynamic_content`: POST `/api/v5/objects/dynamic-contents` - kind `create`; body type
  `json`; risk: POST /api/v5/objects/dynamic-contents in Salesforce Account Engagement.
- `add_tag_to_dynamic_content`: POST `/api/v5/objects/dynamic-contents/{{ record.id }}/do/addTag` -
  kind `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: Invokes Pardot addTag action for dynamic_content.
- `remove_tag_from_dynamic_content`: POST `/api/v5/objects/dynamic-contents/{{ record.id
  }}/do/removeTag` - kind `delete`; body type `none`; path fields `id`; required record fields `id`;
  accepted fields `id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Invokes Pardot removeTag action for dynamic_content.
- `create_email`: POST `/api/v5/objects/emails` - kind `create`; body type `json`; risk: POST
  /api/v5/objects/emails in Salesforce Account Engagement.
- `add_tag_to_email`: POST `/api/v5/objects/emails/{{ record.id }}/do/addTag` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: Invokes
  Pardot addTag action for email.
- `remove_tag_from_email`: POST `/api/v5/objects/emails/{{ record.id }}/do/removeTag` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Invokes
  Pardot removeTag action for email.
- `copy_email_to_cms`: POST `/api/v5/objects/emails/{{ record.id }}/do/copyToCms` - kind `update`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk:
  Invokes Pardot copyToCms action for email.
- `create_email_template`: POST `/api/v5/objects/email-templates` - kind `create`; body type `json`;
  risk: POST /api/v5/objects/email-templates in Salesforce Account Engagement.
- `update_email_template`: PATCH `/api/v5/objects/email-templates/{{ record.id }}` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: PATCH
  /api/v5/objects/email-templates/{{ record.id }} in Salesforce Account Engagement.
- `delete_email_template`: DELETE `/api/v5/objects/email-templates/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Deletes Salesforce
  Account Engagement email_template records.
- `add_tag_to_email_template`: POST `/api/v5/objects/email-templates/{{ record.id }}/do/addTag` -
  kind `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: Invokes Pardot addTag action for email_template.
- `remove_tag_from_email_template`: POST `/api/v5/objects/email-templates/{{ record.id
  }}/do/removeTag` - kind `delete`; body type `none`; path fields `id`; required record fields `id`;
  accepted fields `id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Invokes Pardot removeTag action for email_template.
- `create_list_email`: POST `/api/v5/objects/list-emails` - kind `create`; body type `json`; risk:
  POST /api/v5/objects/list-emails in Salesforce Account Engagement.
- `update_file`: PATCH `/api/v5/objects/files/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: PATCH
  /api/v5/objects/files/{{ record.id }} in Salesforce Account Engagement.
- `delete_file`: DELETE `/api/v5/objects/files/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Deletes Salesforce Account Engagement
  file records.
- `add_tag_to_file`: POST `/api/v5/objects/files/{{ record.id }}/do/addTag` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: Invokes
  Pardot addTag action for file.
- `remove_tag_from_file`: POST `/api/v5/objects/files/{{ record.id }}/do/removeTag` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Invokes Pardot
  removeTag action for file.
- `copy_file_to_cms`: POST `/api/v5/objects/files/{{ record.id }}/do/copyToCms` - kind `update`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk:
  Invokes Pardot copyToCms action for file.
- `create_form`: POST `/api/v5/objects/forms` - kind `create`; body type `json`; risk: POST
  /api/v5/objects/forms in Salesforce Account Engagement.
- `delete_form`: DELETE `/api/v5/objects/forms/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Deletes Salesforce Account Engagement
  form records.
- `undelete_form`: POST `/api/v5/objects/forms/do/undelete` - kind `update`; body type `json`; risk:
  Invokes Pardot undelete action for form.
- `reorder_form_fields`: POST `/api/v5/objects/forms/{{ record.id }}/do/reorderFormFields` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: Invokes Pardot reorderFormFields action for form.
- `add_tag_to_form`: POST `/api/v5/objects/forms/{{ record.id }}/do/addTag` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: Invokes
  Pardot addTag action for form.
- `remove_tag_from_form`: POST `/api/v5/objects/forms/{{ record.id }}/do/removeTag` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Invokes Pardot
  removeTag action for form.
- `copy_form_to_cms`: POST `/api/v5/objects/forms/{{ record.id }}/do/copyToCms` - kind `update`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; risk:
  Invokes Pardot copyToCms action for form.
- `create_form_field`: POST `/api/v5/objects/form-fields` - kind `create`; body type `json`; risk:
  POST /api/v5/objects/form-fields in Salesforce Account Engagement.
- `add_form_field_dependent`: POST `/api/v5/objects/form-fields/{{ record.id }}/do/addDependent` -
  kind `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: Invokes Pardot addDependent action for form_field.
- `add_form_field_progressive`: POST `/api/v5/objects/form-fields/{{ record.id }}/do/addProgressive`
  - kind `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: Invokes Pardot addProgressive action for form_field.
- `add_form_field_value`: POST `/api/v5/objects/form-fields/{{ record.id }}/do/addValue` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: Invokes Pardot addValue action for form_field.
- `reorder_form_field_values`: POST `/api/v5/objects/form-fields/{{ record.id
  }}/do/reorderFormFieldValues` - kind `update`; body type `json`; path fields `id`; required record
  fields `id`; accepted fields `id`; risk: Invokes Pardot reorderFormFieldValues action for
  form_field.
- `add_tag_to_form_field`: POST `/api/v5/objects/form-fields/{{ record.id }}/do/addTag` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: Invokes Pardot addTag action for form_field.
- `remove_tag_from_form_field`: POST `/api/v5/objects/form-fields/{{ record.id }}/do/removeTag` -
  kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  Invokes Pardot removeTag action for form_field.
- `create_form_handler`: POST `/api/v5/objects/form-handlers` - kind `create`; body type `json`;
  risk: POST /api/v5/objects/form-handlers in Salesforce Account Engagement.
- `update_form_handler`: PATCH `/api/v5/objects/form-handlers/{{ record.id }}` - kind `update`; body
  type `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: PATCH
  /api/v5/objects/form-handlers/{{ record.id }} in Salesforce Account Engagement.
- `delete_form_handler`: DELETE `/api/v5/objects/form-handlers/{{ record.id }}` - kind `delete`;
  body type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: Deletes Salesforce
  Account Engagement form_handler records.
- `add_tag_to_form_handler`: POST `/api/v5/objects/form-handlers/{{ record.id }}/do/addTag` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: Invokes Pardot addTag action for form_handler.
- `remove_tag_from_form_handler`: POST `/api/v5/objects/form-handlers/{{ record.id }}/do/removeTag`
  - kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  Invokes Pardot removeTag action for form_handler.
- `create_form_handler_field`: POST `/api/v5/objects/form-handler-fields` - kind `create`; body type
  `json`; risk: POST /api/v5/objects/form-handler-fields in Salesforce Account Engagement.
- `update_form_handler_field`: PATCH `/api/v5/objects/form-handler-fields/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: PATCH /api/v5/objects/form-handler-fields/{{ record.id }} in Salesforce Account Engagement.
- `delete_form_handler_field`: DELETE `/api/v5/objects/form-handler-fields/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  Salesforce Account Engagement form_handler_field records.
- `create_landing_page`: POST `/api/v5/objects/landing-pages` - kind `create`; body type `json`;
  risk: POST /api/v5/objects/landing-pages in Salesforce Account Engagement.
- `add_tag_to_landing_page`: POST `/api/v5/objects/landing-pages/{{ record.id }}/do/addTag` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: Invokes Pardot addTag action for landing_page.
- `remove_tag_from_landing_page`: POST `/api/v5/objects/landing-pages/{{ record.id }}/do/removeTag`
  - kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  Invokes Pardot removeTag action for landing_page.
- `copy_landing_page_to_cms`: POST `/api/v5/objects/landing-pages/{{ record.id }}/do/copyToCms` -
  kind `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: Invokes Pardot copyToCms action for landing_page.
- `create_layout_template`: POST `/api/v5/objects/layout-templates` - kind `create`; body type
  `json`; risk: POST /api/v5/objects/layout-templates in Salesforce Account Engagement.
- `update_layout_template`: PATCH `/api/v5/objects/layout-templates/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: PATCH /api/v5/objects/layout-templates/{{ record.id }} in Salesforce Account Engagement.
- `delete_layout_template`: DELETE `/api/v5/objects/layout-templates/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  Salesforce Account Engagement layout_template records.
- `add_tag_to_layout_template`: POST `/api/v5/objects/layout-templates/{{ record.id }}/do/addTag` -
  kind `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: Invokes Pardot addTag action for layout_template.
- `remove_tag_from_layout_template`: POST `/api/v5/objects/layout-templates/{{ record.id
  }}/do/removeTag` - kind `delete`; body type `none`; path fields `id`; required record fields `id`;
  accepted fields `id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: Invokes Pardot removeTag action for layout_template.
- `create_list_membership`: POST `/api/v5/objects/list-memberships` - kind `create`; body type
  `json`; risk: POST /api/v5/objects/list-memberships in Salesforce Account Engagement.
- `update_list_membership`: PATCH `/api/v5/objects/list-memberships/{{ record.id }}` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: PATCH /api/v5/objects/list-memberships/{{ record.id }} in Salesforce Account Engagement.
- `delete_list_membership`: DELETE `/api/v5/objects/list-memberships/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: Deletes
  Salesforce Account Engagement list_membership records.
- `add_tag_to_opportunity`: POST `/api/v5/objects/opportunities/{{ record.id }}/do/addTag` - kind
  `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields `id`;
  risk: Invokes Pardot addTag action for opportunity.
- `remove_tag_from_opportunity`: POST `/api/v5/objects/opportunities/{{ record.id }}/do/removeTag` -
  kind `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  Invokes Pardot removeTag action for opportunity.
- `create_tag`: POST `/api/v5/objects/tags` - kind `create`; body type `json`; risk: POST
  /api/v5/objects/tags in Salesforce Account Engagement.
- `update_tag`: PATCH `/api/v5/objects/tags/{{ record.id }}` - kind `update`; body type `json`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: PATCH
  /api/v5/objects/tags/{{ record.id }} in Salesforce Account Engagement.
- `delete_tag`: DELETE `/api/v5/objects/tags/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: Deletes Salesforce Account Engagement
  tag records.
- `merge_tags`: POST `/api/v5/objects/tags/{{ record.id }}/do/mergeTags` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: Invokes Pardot
  mergeTags action for tag.
- `assign_visitor_to_prospect`: POST `/api/v5/objects/visitors/{{ record.id }}/do/assignToProspect`
  - kind `update`; body type `json`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: Invokes Pardot assignToProspect action for visitor.
- `identify_visitor_company`: POST `/api/v5/objects/visitors/{{ record.id }}/do/identifyCompany` -
  kind `update`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: Invokes Pardot identifyCompany action for visitor.
- `update_bulk_action`: PATCH `/api/v5/bulk-actions/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `id`; risk: PATCH
  /api/v5/bulk-actions/{{ record.id }} in Salesforce Account Engagement.
- `create_import_job`: POST `/api/v5/imports` - kind `create`; body type `json`; risk: POST
  /api/v5/imports in Salesforce Account Engagement.
- `update_import_job`: PATCH `/api/v5/imports/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `id`; risk: PATCH
  /api/v5/imports/{{ record.id }} in Salesforce Account Engagement.
- `cancel_import`: POST `/api/v5/imports/{{ record.id }}/do/cancel` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: Invokes Pardot cancel
  action for import_job.
- `create_external_activity`: POST `/api/v5/external-activities` - kind `create`; body type `json`;
  risk: POST /api/v5/external-activities in Salesforce Account Engagement.

## Known limits

- Batch defaults: read_page_size=200, write_batch_size=1.
- API coverage includes 72 stream-backed endpoint group(s), 92 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=7.
