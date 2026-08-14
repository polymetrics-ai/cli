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
  password (secret) (required)

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
  ETL sync modes: full_refresh_append, full_refresh_overwrite

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
