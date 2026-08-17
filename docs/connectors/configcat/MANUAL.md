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
    fields: name(), organization_id()
  products:
    primary key: product_id
    fields: approve_required(), description(), name(), order(), organization_id(), product_id(), reason_required()
  configs:
    primary key: config_id
    fields: config_id(), description(), evaluation_version(), migrated_config_id(), name(), order(), product_id()
  environments:
    primary key: environment_id
    fields: approve_required(), color(), description(), environment_id(), name(), order(), product_id(), reason_required()
  tags:
    primary key: tag_id
    fields: color(), name(), product_id(), tag_id()
  config:
    primary key: configId
    fields: configId(), description(), evaluationVersion(), migratedConfigId(), name(), order(), product()
  environment:
    primary key: environmentId
    fields: approveRequired(), color(), description(), environmentId(), name(), order(), product(), reasonRequired()
  settings:
    primary key: settingId
    fields: configId(), configName(), createdAt(), hint(), isJson(), key(), name(), order(), predefinedVariations(), settingId(), settingType(), tags()
  setting:
    primary key: settingId
    fields: configId(), configName(), createdAt(), hint(), isJson(), key(), name(), order(), predefinedVariations(), settingId(), settingType(), tags()
  deleted_settings:
    primary key: key
    fields: hint(), key(), name(), settingType()
  sdk_keys:
    primary key: primary
    fields: primary(), secondary()
  config_setting_values:
    primary key: readOnly
    fields: config(), environment(), featureFlagLimitations(), readOnly(), settingValues()
  segments:
    primary key: segmentId
    fields: createdAt(), creatorEmail(), creatorFullName(), description(), lastUpdaterEmail(), lastUpdaterFullName(), name(), product(), segmentId(), updatedAt(), usage()
  segment:
    primary key: segmentId
    fields: comparator(), comparisonAttribute(), comparisonValue(), createdAt(), creatorEmail(), creatorFullName(), description(), lastUpdaterEmail(), lastUpdaterFullName(), name(), product(), segmentId(), updatedAt()
  webhooks:
    primary key: webhookId
    fields: config(), content(), environment(), httpMethod(), url(), webHookHeaders(), webhookId()
  webhook:
    primary key: webhookId
    fields: config(), content(), environment(), httpMethod(), url(), webHookHeaders(), webhookId()
  permission_groups:
    primary key: permissionGroupId
    fields: accessType(), canCreateOrUpdateConfig(), canCreateOrUpdateEnvironment(), canCreateOrUpdateSetting(), canDeleteConfig(), canDeleteEnvironment(), canDeleteSetting(), canManageMembers(), name(), permissionGroupId(), product()
  permission_group:
    primary key: permissionGroupId
    fields: accessType(), canCreateOrUpdateConfig(), canCreateOrUpdateEnvironment(), canCreateOrUpdateSetting(), canDeleteConfig(), canDeleteEnvironment(), canDeleteSetting(), canManageMembers(), name(), permissionGroupId(), product()
  integrations:
    primary key: integrationId
    fields: configIds(), environmentIds(), integrationId(), integrationType(), name(), parameters(), product()
  integration:
    primary key: integrationId
    fields: configIds(), environmentIds(), integrationId(), integrationType(), name(), parameters(), product()
  proxy_profiles:
    primary key: proxyProfileId
    fields: connectionPreferences(), description(), lastAccessedAt(), name(), proxyProfileId(), sdkKeySelectionRules()
  proxy_profile:
    primary key: proxyProfileId
    fields: connectionPreferences(), description(), lastAccessedAt(), name(), proxyProfileId(), sdkKeySelectionRules()
  members:
    primary key: userId
    fields: email(), fullName(), twoFactorEnabled(), userId()
  audit_logs:
    primary key: auditLogId
    fields: actionTarget(), auditLogDateTime(), auditLogId(), auditLogType(), auditLogTypeEnum(), details(), modelVersion(), truncated(), userEmail(), userName(), where(), why()
  stale_flags:
    primary key: productId
    fields: configs(), environments(), name(), productId()
  me:
    primary key: email
    fields: email(), fullName()
  tag:
    primary key: tagId
    fields: color(), name(), product(), tagId()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

REVERSE ETL ACTIONS
  create_config:
    endpoint: POST /v1/products/{{ config.product_id }}/configs
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
    risk: creates a new ConfigCat feature flag/setting within the configured config; low risk, no data destruction
  update_flag:
    endpoint: PUT /v1/settings/{{ record.settingId }}
    required fields: settingId
    risk: replaces an existing ConfigCat feature flag/setting's metadata (name/hint/tags); does not itself change the flag's evaluated VALUE in any environment
  delete_flag:
    endpoint: DELETE /v1/settings/{{ record.settingId }}
    required fields: settingId
    risk: permanently deletes a ConfigCat feature flag/setting and its values in every environment; destructive, external mutation; approval required
  create_tag:
    endpoint: POST /v1/products/{{ config.product_id }}/tags
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
