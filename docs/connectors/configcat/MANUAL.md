# pm connectors inspect configcat

```text
NAME
  pm connectors inspect configcat - ConfigCat connector manual

SYNOPSIS
  pm connectors inspect configcat
  pm connectors inspect configcat --json
  pm credentials add <name> --connector configcat [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes ConfigCat feature-flag platform data: organizations, products, configs, environments, settings/feature flags, deleted settings, SDK keys, segments, webhooks, permission groups, integrations, proxy profiles, members, audit logs, stale flags, tags, and the authenticated user's own profile through the ConfigCat Public Management API.

ICON
  id: configcat
  asset: icons/configcat.svg
  source: upstream_registry
  review_status: upstream_seeded
  review_url: https://api.configcat.com/docs/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  audit_log_config_id
  audit_log_environment_id
  base_url
  config_id
  environment_id
  integration_id
  mode
  organization_id
  permission_group_id
  product_id
  proxy_profile_id
  segment_id
  setting_id
  tag_id
  username
  webhook_id
  password (secret)

ETL STREAMS
  organizations:
    primary key: organization_id
    fields: name(string), organization_id(string)
  products:
    primary key: product_id
    fields: approve_required(boolean), description(string), name(string), order(integer), organization_id(string), product_id(string), reason_required(boolean)
  configs:
    primary key: config_id
    fields: config_id(string), description(string), evaluation_version(string), migrated_config_id(string), name(string), order(integer), product_id(string)
  environments:
    primary key: environment_id
    fields: approve_required(boolean), color(string), description(string), environment_id(string), name(string), order(integer), product_id(string), reason_required(boolean)
  tags:
    primary key: tag_id
    fields: color(string), name(string), product_id(string), tag_id(integer)
  config:
    primary key: configId
    fields: configId(string), description(string), evaluationVersion(string), migratedConfigId(string), name(string), order(integer), product(object)
  environment:
    primary key: environmentId
    fields: approveRequired(boolean), color(string), description(string), environmentId(string), name(string), order(integer), product(object), reasonRequired(boolean)
  settings:
    primary key: settingId
    fields: configId(string), configName(string), createdAt(string), hint(string), isJson(boolean), key(string), name(string), order(integer), predefinedVariations(array), settingId(integer), settingType(string), tags(array)
  setting:
    primary key: settingId
    fields: configId(string), configName(string), createdAt(string), hint(string), isJson(boolean), key(string), name(string), order(integer), predefinedVariations(array), settingId(integer), settingType(string), tags(array)
  deleted_settings:
    primary key: key
    fields: hint(string), key(string), name(string), settingType(string)
  sdk_keys:
    primary key: primary
    fields: primary(string), secondary(string)
  config_setting_values:
    primary key: readOnly
    fields: config(object), environment(object), featureFlagLimitations(object), readOnly(boolean), settingValues(array)
  segments:
    primary key: segmentId
    fields: createdAt(string), creatorEmail(string), creatorFullName(string), description(string), lastUpdaterEmail(string), lastUpdaterFullName(string), name(string), product(object), segmentId(string), updatedAt(string), usage(object)
  segment:
    primary key: segmentId
    fields: comparator(string), comparisonAttribute(string), comparisonValue(string), createdAt(string), creatorEmail(string), creatorFullName(string), description(string), lastUpdaterEmail(string), lastUpdaterFullName(string), name(string), product(object), segmentId(string), updatedAt(string)
  webhooks:
    primary key: webhookId
    fields: config(object), content(string), environment(object), httpMethod(string), url(string), webHookHeaders(array), webhookId(integer)
  webhook:
    primary key: webhookId
    fields: config(object), content(string), environment(object), httpMethod(string), url(string), webHookHeaders(array), webhookId(integer)
  permission_groups:
    primary key: permissionGroupId
    fields: accessType(string), canCreateOrUpdateConfig(boolean), canCreateOrUpdateEnvironment(boolean), canCreateOrUpdateSetting(boolean), canDeleteConfig(boolean), canDeleteEnvironment(boolean), canDeleteSetting(boolean), canManageMembers(boolean), name(string), permissionGroupId(integer), product(object)
  permission_group:
    primary key: permissionGroupId
    fields: accessType(string), canCreateOrUpdateConfig(boolean), canCreateOrUpdateEnvironment(boolean), canCreateOrUpdateSetting(boolean), canDeleteConfig(boolean), canDeleteEnvironment(boolean), canDeleteSetting(boolean), canManageMembers(boolean), name(string), permissionGroupId(integer), product(object)
  integrations:
    primary key: integrationId
    fields: configIds(array), environmentIds(array), integrationId(string), integrationType(string), name(string), parameters(object), product(object)
  integration:
    primary key: integrationId
    fields: configIds(array), environmentIds(array), integrationId(string), integrationType(string), name(string), parameters(object), product(object)
  proxy_profiles:
    primary key: proxyProfileId
    fields: connectionPreferences(object), description(string), lastAccessedAt(string), name(string), proxyProfileId(string), sdkKeySelectionRules(array)
  proxy_profile:
    primary key: proxyProfileId
    fields: connectionPreferences(object), description(string), lastAccessedAt(string), name(string), proxyProfileId(string), sdkKeySelectionRules(array)
  members:
    primary key: userId
    fields: email(string), fullName(string), twoFactorEnabled(boolean), userId(string)
  audit_logs:
    primary key: auditLogId
    fields: actionTarget(object), auditLogDateTime(string), auditLogId(string), auditLogType(string), auditLogTypeEnum(string), details(object), modelVersion(integer), truncated(boolean), userEmail(string), userName(string), where(object), why(object)
  stale_flags:
    primary key: productId
    fields: configs(array), environments(array), name(string), productId(string)
  me:
    primary key: email
    fields: email(string), fullName(string)
  tag:
    primary key: tagId
    fields: color(string), name(string), product(object), tagId(integer)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_config:
    endpoint: POST /v1/products/{{ config.product_id }}/configs
    required fields: name
    risk: creates a new ConfigCat config within the configured product; low risk, no data destruction
  update_config:
    endpoint: PUT /v1/configs/{{ record.configId }}
    required fields: configId
    risk: renames/reorders an existing ConfigCat config; may affect SDK-visible dashboard organization
  delete_config:
    endpoint: DELETE /v1/configs/{{ record.configId }}
    required fields: configId
    risk: permanently deletes a ConfigCat config and every feature flag/setting defined in it; destructive, external mutation; approval required
  create_environment:
    endpoint: POST /v1/products/{{ config.product_id }}/environments
    required fields: name
    risk: creates a new ConfigCat environment within the configured product; low risk, no data destruction
  update_environment:
    endpoint: PUT /v1/environments/{{ record.environmentId }}
    required fields: environmentId
    risk: renames/recolors an existing ConfigCat environment; may affect dashboard organization visible to other users
  delete_environment:
    endpoint: DELETE /v1/environments/{{ record.environmentId }}
    required fields: environmentId
    risk: permanently deletes a ConfigCat environment and every feature flag value/SDK key scoped to it; destructive, external mutation; approval required
  create_flag:
    endpoint: POST /v1/configs/{{ config.config_id }}/settings
    required fields: key, name, settingType
    risk: creates a new ConfigCat feature flag/setting within the configured config; low risk, no data destruction
  update_flag:
    endpoint: PUT /v1/settings/{{ record.settingId }}
    required fields: settingId, name
    risk: replaces an existing ConfigCat feature flag/setting's metadata (name/hint/tags); does not itself change the flag's evaluated VALUE in any environment
  delete_flag:
    endpoint: DELETE /v1/settings/{{ record.settingId }}
    required fields: settingId
    risk: permanently deletes a ConfigCat feature flag/setting and its values in every environment; destructive, external mutation; approval required
  create_tag:
    endpoint: POST /v1/products/{{ config.product_id }}/tags
    required fields: name
    risk: creates a new ConfigCat tag within the configured product; low risk, no data destruction
  update_tag:
    endpoint: PUT /v1/tags/{{ record.tagId }}
    required fields: tagId
    risk: renames/recolors an existing ConfigCat tag; affects every feature flag tagged with it
  delete_tag:
    endpoint: DELETE /v1/tags/{{ record.tagId }}
    required fields: tagId
    risk: permanently deletes a ConfigCat tag and untags every feature flag that used it; destructive, external mutation; approval required

SECURITY
  read risk: external ConfigCat Public Management API read of organization/product/config/environment/setting metadata plus segments, webhooks, permission groups, integrations, proxy profiles, members, and audit logs
  write risk: external mutation of ConfigCat configs, environments, feature flags/settings, and tags (create/update/delete); does not change a feature flag's evaluated VALUE in any environment (see docs.md)
  approval: required for delete_config/delete_environment/delete_flag/delete_tag (destructive, cascades to dependent data); create/update actions are lower risk but still mutate shared product configuration
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

COMMAND SURFACE
  Run ConfigCat's declared streams and reverse-ETL actions.
  Usage: pm configcat <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api delete v1 environments environmentid settings settingid integrationlinks integrationlinktype key - Documented DELETE /v1/environments/{environmentId}/settings/{settingId}/integrationLinks/{integrationLinkType}/{key} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.delete.v1-environments-environmentid-settings-settingid-integrationlinks-integrationlinktype-key]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 integrations integrationid - Documented DELETE /v1/integrations/{integrationId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.delete.v1-integrations-integrationid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 invitations invitationid - Documented DELETE /v1/invitations/{invitationId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.delete.v1-invitations-invitationid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 organizations organizationid members userid - Documented DELETE /v1/organizations/{organizationId}/members/{userId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.delete.v1-organizations-organizationid-members-userid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 permissions permissiongroupid - Documented DELETE /v1/permissions/{permissionGroupId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.delete.v1-permissions-permissiongroupid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 products productid - Documented DELETE /v1/products/{productId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.delete.v1-products-productid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 products productid members userid - Documented DELETE /v1/products/{productId}/members/{userId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.delete.v1-products-productid-members-userid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 proxy-profiles proxyprofileid - Documented DELETE /v1/proxy-profiles/{proxyProfileId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.delete.v1-proxy-profiles-proxyprofileid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 segments segmentid - Documented DELETE /v1/segments/{segmentId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.delete.v1-segments-segmentid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v1 webhooks webhookid - Documented DELETE /v1/webhooks/{webhookId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.delete.v1-webhooks-webhookid]; approval: not implemented: the provider webhook-URL mutation has no dedicated security-reviewed execution contract; risk: high; notes: named_dependency=review.webhook_url_mutation: the provider webhook-URL mutation has no dedicated security-reviewed execution contract
    api delete v2 change-request-comments commentid - Documented DELETE /v2/change-request-comments/{commentId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.delete.v2-change-request-comments-commentid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api delete v2 change-requests changerequestid proposed-changes settingid - Documented DELETE /v2/change-requests/{changeRequestId}/proposed-changes/{settingId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.delete.v2-change-requests-changerequestid-proposed-changes-settingid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get v1 environments environmentid settings settingid value - Documented GET /v1/environments/{environmentId}/settings/{settingId}/value (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v1-environments-environmentid-settings-settingid-value]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 integrationlink integrationlinktype key details - Documented GET /v1/integrationLink/{integrationLinkType}/{key}/details (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v1-integrationlink-integrationlinktype-key-details]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 organizations organizationid auditlogs - Documented GET /v1/organizations/{organizationId}/auditlogs (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v1-organizations-organizationid-auditlogs]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 organizations organizationid invitations - Documented GET /v1/organizations/{organizationId}/invitations (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v1-organizations-organizationid-invitations]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 organizations organizationid organization-limitations - Documented GET /v1/organizations/{organizationId}/organization-limitations (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v1-organizations-organizationid-organization-limitations]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 products productid - Documented GET /v1/products/{productId} (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v1-products-productid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 products productid invitations - Documented GET /v1/products/{productId}/invitations (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v1-products-productid-invitations]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 products productid members - Documented GET /v1/products/{productId}/members (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v1-products-productid-members]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 products productid preferences - Documented GET /v1/products/{productId}/preferences (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v1-products-productid-preferences]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 proxy-profiles proxyprofileid sdk-keys - Documented GET /v1/proxy-profiles/{proxyProfileId}/sdk-keys (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v1-proxy-profiles-proxyprofileid-sdk-keys]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 settings settingid code-references - Documented GET /v1/settings/{settingId}/code-references (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v1-settings-settingid-code-references]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 settings settingid predefined-variations - Documented GET /v1/settings/{settingId}/predefined-variations (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v1-settings-settingid-predefined-variations]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 settings settingkeyorid value - Documented GET /v1/settings/{settingKeyOrId}/value (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v1-settings-settingkeyorid-value]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 tags tagid settings - Documented GET /v1/tags/{tagId}/settings (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v1-tags-tagid-settings]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v1 webhooks webhookid keys - Documented GET /v1/webhooks/{webhookId}/keys (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v1-webhooks-webhookid-keys]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 change-requests changerequestid - Documented GET /v2/change-requests/{changeRequestId} (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v2-change-requests-changerequestid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 change-requests changerequestid proposed-changes - Documented GET /v2/change-requests/{changeRequestId}/proposed-changes (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v2-change-requests-changerequestid-proposed-changes]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 configs configid environments environmentid values - Documented GET /v2/configs/{configId}/environments/{environmentId}/values (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v2-configs-configid-environments-environmentid-values]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 environments environmentid settings settingid value - Documented GET /v2/environments/{environmentId}/settings/{settingId}/value (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v2-environments-environmentid-settings-settingid-value]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 organizations organizationid auditlogs - Documented GET /v2/organizations/{organizationId}/auditlogs (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v2-organizations-organizationid-auditlogs]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 organizations organizationid members - Documented GET /v2/organizations/{organizationId}/members (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v2-organizations-organizationid-members]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 products productid auditlogs - Documented GET /v2/products/{productId}/auditlogs (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v2-products-productid-auditlogs]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 products productid change-requests - Documented GET /v2/products/{productId}/change-requests (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v2-products-productid-change-requests]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get v2 settings settingkeyorid value - Documented GET /v2/settings/{settingKeyOrId}/value (not implemented) [intent=direct_read availability=not_implemented operation=configcat.get.v2-settings-settingkeyorid-value]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api patch v1 environments environmentid settings settingid value - Documented PATCH /v1/environments/{environmentId}/settings/{settingId}/value (not implemented) [intent=direct_write availability=not_implemented operation=configcat.patch.v1-environments-environmentid-settings-settingid-value]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch v1 proxy-profiles proxyprofileid - Documented PATCH /v1/proxy-profiles/{proxyProfileId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.patch.v1-proxy-profiles-proxyprofileid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch v1 settings settingid - Documented PATCH /v1/settings/{settingId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.patch.v1-settings-settingid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch v1 settings settingkeyorid value - Documented PATCH /v1/settings/{settingKeyOrId}/value (not implemented) [intent=direct_write availability=not_implemented operation=configcat.patch.v1-settings-settingkeyorid-value]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch v1 webhooks webhookid - Documented PATCH /v1/webhooks/{webhookId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.patch.v1-webhooks-webhookid]; approval: not implemented: the provider webhook-URL mutation has no dedicated security-reviewed execution contract; risk: high; notes: named_dependency=review.webhook_url_mutation: the provider webhook-URL mutation has no dedicated security-reviewed execution contract
    api patch v2 environments environmentid settings settingid value - Documented PATCH /v2/environments/{environmentId}/settings/{settingId}/value (not implemented) [intent=direct_write availability=not_implemented operation=configcat.patch.v2-environments-environmentid-settings-settingid-value]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api patch v2 settings settingkeyorid value - Documented PATCH /v2/settings/{settingKeyOrId}/value (not implemented) [intent=direct_write availability=not_implemented operation=configcat.patch.v2-settings-settingkeyorid-value]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 code-references - Documented POST /v1/code-references (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v1-code-references]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 code-references delete-reports - Documented POST /v1/code-references/delete-reports (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v1-code-references-delete-reports]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 configs configid environments environmentid values - Documented POST /v1/configs/{configId}/environments/{environmentId}/values (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v1-configs-configid-environments-environmentid-values]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 configs configid environments environmentid webhooks - Documented POST /v1/configs/{configId}/environments/{environmentId}/webhooks (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v1-configs-configid-environments-environmentid-webhooks]; approval: not implemented: the provider webhook-URL mutation has no dedicated security-reviewed execution contract; risk: high; notes: named_dependency=review.webhook_url_mutation: the provider webhook-URL mutation has no dedicated security-reviewed execution contract
    api post v1 environments environmentid settings settingid integrationlinks integrationlinktype key - Documented POST /v1/environments/{environmentId}/settings/{settingId}/integrationLinks/{integrationLinkType}/{key} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v1-environments-environmentid-settings-settingid-integrationlinks-integrationlinktype-key]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 jira connect - Documented POST /v1/jira/connect (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v1-jira-connect]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 jira environments environmentid settings settingid integrationlinks key - Documented POST /v1/jira/environments/{environmentId}/settings/{settingId}/integrationLinks/{key} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v1-jira-environments-environmentid-settings-settingid-integrationlinks-key]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 organizations organizationid members userid - Documented POST /v1/organizations/{organizationId}/members/{userId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v1-organizations-organizationid-members-userid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 organizations organizationid products - Documented POST /v1/organizations/{organizationId}/products (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v1-organizations-organizationid-products]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 organizations organizationid proxy-profiles - Documented POST /v1/organizations/{organizationId}/proxy-profiles (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v1-organizations-organizationid-proxy-profiles]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 products productid integrations - Documented POST /v1/products/{productId}/integrations (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v1-products-productid-integrations]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 products productid members invite - Documented POST /v1/products/{productId}/members/invite (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v1-products-productid-members-invite]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 products productid permissions - Documented POST /v1/products/{productId}/permissions (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v1-products-productid-permissions]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 products productid preferences - Documented POST /v1/products/{productId}/preferences (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v1-products-productid-preferences]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 products productid segments - Documented POST /v1/products/{productId}/segments (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v1-products-productid-segments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 proxy-profiles proxyprofileid sdk-keys deselect - Documented POST /v1/proxy-profiles/{proxyProfileId}/sdk-keys/deselect (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v1-proxy-profiles-proxyprofileid-sdk-keys-deselect]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 proxy-profiles proxyprofileid sdk-keys select - Documented POST /v1/proxy-profiles/{proxyProfileId}/sdk-keys/select (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v1-proxy-profiles-proxyprofileid-sdk-keys-select]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v1 proxy-profiles proxyprofileid secret - Documented POST /v1/proxy-profiles/{proxyProfileId}/secret (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v1-proxy-profiles-proxyprofileid-secret]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 change-requests changerequestid apply - Documented POST /v2/change-requests/{changeRequestId}/apply (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v2-change-requests-changerequestid-apply]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 change-requests changerequestid approve - Documented POST /v2/change-requests/{changeRequestId}/approve (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v2-change-requests-changerequestid-approve]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 change-requests changerequestid claim-ownership - Documented POST /v2/change-requests/{changeRequestId}/claim-ownership (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v2-change-requests-changerequestid-claim-ownership]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 change-requests changerequestid close - Documented POST /v2/change-requests/{changeRequestId}/close (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v2-change-requests-changerequestid-close]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 change-requests changerequestid comments - Documented POST /v2/change-requests/{changeRequestId}/comments (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v2-change-requests-changerequestid-comments]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 change-requests changerequestid proposed-changes settingid resolve-conflicts - Documented POST /v2/change-requests/{changeRequestId}/proposed-changes/{settingId}/resolve-conflicts (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v2-change-requests-changerequestid-proposed-changes-settingid-resolve-conflicts]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 change-requests changerequestid remove-approval - Documented POST /v2/change-requests/{changeRequestId}/remove-approval (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v2-change-requests-changerequestid-remove-approval]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 configs configid environments environmentid change-requests - Documented POST /v2/configs/{configId}/environments/{environmentId}/change-requests (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v2-configs-configid-environments-environmentid-change-requests]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post v2 configs configid environments environmentid values - Documented POST /v2/configs/{configId}/environments/{environmentId}/values (not implemented) [intent=direct_write availability=not_implemented operation=configcat.post.v2-configs-configid-environments-environmentid-values]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v1 environments environmentid settings settingid value - Documented PUT /v1/environments/{environmentId}/settings/{settingId}/value (not implemented) [intent=direct_write availability=not_implemented operation=configcat.put.v1-environments-environmentid-settings-settingid-value]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v1 integrations integrationid - Documented PUT /v1/integrations/{integrationId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.put.v1-integrations-integrationid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v1 permissions permissiongroupid - Documented PUT /v1/permissions/{permissionGroupId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.put.v1-permissions-permissiongroupid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v1 products productid - Documented PUT /v1/products/{productId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.put.v1-products-productid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v1 proxy-profiles proxyprofileid - Documented PUT /v1/proxy-profiles/{proxyProfileId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.put.v1-proxy-profiles-proxyprofileid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v1 segments segmentid - Documented PUT /v1/segments/{segmentId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.put.v1-segments-segmentid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v1 settings settingid predefined-variations - Documented PUT /v1/settings/{settingId}/predefined-variations (not implemented) [intent=direct_write availability=not_implemented operation=configcat.put.v1-settings-settingid-predefined-variations]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v1 settings settingkeyorid value - Documented PUT /v1/settings/{settingKeyOrId}/value (not implemented) [intent=direct_write availability=not_implemented operation=configcat.put.v1-settings-settingkeyorid-value]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v1 webhooks webhookid - Documented PUT /v1/webhooks/{webhookId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.put.v1-webhooks-webhookid]; approval: not implemented: the provider webhook-URL mutation has no dedicated security-reviewed execution contract; risk: high; notes: named_dependency=review.webhook_url_mutation: the provider webhook-URL mutation has no dedicated security-reviewed execution contract
    api put v2 change-request-comments commentid - Documented PUT /v2/change-request-comments/{commentId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.put.v2-change-request-comments-commentid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v2 change-requests changerequestid - Documented PUT /v2/change-requests/{changeRequestId} (not implemented) [intent=direct_write availability=not_implemented operation=configcat.put.v2-change-requests-changerequestid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v2 change-requests changerequestid proposed-changes - Documented PUT /v2/change-requests/{changeRequestId}/proposed-changes (not implemented) [intent=direct_write availability=not_implemented operation=configcat.put.v2-change-requests-changerequestid-proposed-changes]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v2 environments environmentid settings settingid value - Documented PUT /v2/environments/{environmentId}/settings/{settingId}/value (not implemented) [intent=direct_write availability=not_implemented operation=configcat.put.v2-environments-environmentid-settings-settingid-value]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api put v2 settings settingkeyorid value - Documented PUT /v2/settings/{settingKeyOrId}/value (not implemented) [intent=direct_write availability=not_implemented operation=configcat.put.v2-settings-settingkeyorid-value]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    audit logs list - Run the audit logs ETL stream [intent=etl availability=implemented stream=audit_logs]
    config list - Run the config ETL stream [intent=etl availability=implemented stream=config]
    config setting values list - Run the config setting values ETL stream [intent=etl availability=implemented stream=config_setting_values]
    configs list - Run the configs ETL stream [intent=etl availability=implemented stream=configs]
    create config apply - Plan and execute the create config reverse-ETL action [intent=reverse_etl availability=implemented write=create_config]; approval: requires plan, preview, approval, and execute; risk: creates a new ConfigCat config within the configured product; low risk, no data destruction; flags: --name (required)
    create environment apply - Plan and execute the create environment reverse-ETL action [intent=reverse_etl availability=implemented write=create_environment]; approval: requires plan, preview, approval, and execute; risk: creates a new ConfigCat environment within the configured product; low risk, no data destruction; flags: --name (required)
    create flag apply - Plan and execute the create flag reverse-ETL action [intent=reverse_etl availability=implemented write=create_flag]; approval: requires plan, preview, approval, and execute; risk: creates a new ConfigCat feature flag/setting within the configured config; low risk, no data destruction; flags: --key (required), --name (required), --settingType (required)
    create tag apply - Plan and execute the create tag reverse-ETL action [intent=reverse_etl availability=implemented write=create_tag]; approval: requires plan, preview, approval, and execute; risk: creates a new ConfigCat tag within the configured product; low risk, no data destruction; flags: --name (required)
    delete config apply - Plan and execute the delete config reverse-ETL action [intent=reverse_etl availability=implemented write=delete_config]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a ConfigCat config and every feature flag/setting defined in it; destructive, external mutation; approval required; flags: --configId (required)
    delete environment apply - Plan and execute the delete environment reverse-ETL action [intent=reverse_etl availability=implemented write=delete_environment]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a ConfigCat environment and every feature flag value/SDK key scoped to it; destructive, external mutation; approval required; flags: --environmentId (required)
    delete flag apply - Plan and execute the delete flag reverse-ETL action [intent=reverse_etl availability=implemented write=delete_flag]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a ConfigCat feature flag/setting and its values in every environment; destructive, external mutation; approval required; flags: --settingId (required)
    delete tag apply - Plan and execute the delete tag reverse-ETL action [intent=reverse_etl availability=implemented write=delete_tag]; approval: requires plan, preview, approval, and execute; risk: permanently deletes a ConfigCat tag and untags every feature flag that used it; destructive, external mutation; approval required; flags: --tagId (required)
    deleted settings list - Run the deleted settings ETL stream [intent=etl availability=implemented stream=deleted_settings]
    environment list - Run the environment ETL stream [intent=etl availability=implemented stream=environment]
    environments list - Run the environments ETL stream [intent=etl availability=implemented stream=environments]
    integration list - Run the integration ETL stream [intent=etl availability=implemented stream=integration]
    integrations list - Run the integrations ETL stream [intent=etl availability=implemented stream=integrations]
    me list - Run the me ETL stream [intent=etl availability=implemented stream=me]
    members list - Run the members ETL stream [intent=etl availability=implemented stream=members]
    organizations list - Run the organizations ETL stream [intent=etl availability=implemented stream=organizations]
    permission group list - Run the permission group ETL stream [intent=etl availability=implemented stream=permission_group]
    permission groups list - Run the permission groups ETL stream [intent=etl availability=implemented stream=permission_groups]
    products list - Run the products ETL stream [intent=etl availability=implemented stream=products]
    proxy profile list - Run the proxy profile ETL stream [intent=etl availability=implemented stream=proxy_profile]
    proxy profiles list - Run the proxy profiles ETL stream [intent=etl availability=implemented stream=proxy_profiles]
    sdk keys list - Run the sdk keys ETL stream [intent=etl availability=implemented stream=sdk_keys]
    segment list - Run the segment ETL stream [intent=etl availability=implemented stream=segment]
    segments list - Run the segments ETL stream [intent=etl availability=implemented stream=segments]
    setting list - Run the setting ETL stream [intent=etl availability=implemented stream=setting]
    settings list - Run the settings ETL stream [intent=etl availability=implemented stream=settings]
    stale flags list - Run the stale flags ETL stream [intent=etl availability=implemented stream=stale_flags]
    tag list - Run the tag ETL stream [intent=etl availability=implemented stream=tag]
    tags list - Run the tags ETL stream [intent=etl availability=implemented stream=tags]
    update config apply - Plan and execute the update config reverse-ETL action [intent=reverse_etl availability=implemented write=update_config]; approval: requires plan, preview, approval, and execute; risk: renames/reorders an existing ConfigCat config; may affect SDK-visible dashboard organization; flags: --configId (required)
    update environment apply - Plan and execute the update environment reverse-ETL action [intent=reverse_etl availability=implemented write=update_environment]; approval: requires plan, preview, approval, and execute; risk: renames/recolors an existing ConfigCat environment; may affect dashboard organization visible to other users; flags: --environmentId (required)
    update flag apply - Plan and execute the update flag reverse-ETL action [intent=reverse_etl availability=implemented write=update_flag]; approval: requires plan, preview, approval, and execute; risk: replaces an existing ConfigCat feature flag/setting's metadata (name/hint/tags); does not itself change the flag's evaluated VALUE in any environment; flags: --name (required), --settingId (required)
    update tag apply - Plan and execute the update tag reverse-ETL action [intent=reverse_etl availability=implemented write=update_tag]; approval: requires plan, preview, approval, and execute; risk: renames/recolors an existing ConfigCat tag; affects every feature flag tagged with it; flags: --tagId (required)
    webhook list - Run the webhook ETL stream [intent=etl availability=implemented stream=webhook]
    webhooks list - Run the webhooks ETL stream [intent=etl availability=implemented stream=webhooks]

EXAMPLES
  # Inspect as a manual
  pm connectors inspect configcat

  # Inspect as structured JSON
  pm connectors inspect configcat --json

AGENT WORKFLOW
  - Run pm connectors inspect configcat before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
