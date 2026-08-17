# pm connectors inspect okta

```text
NAME
  pm connectors inspect okta - Okta connector manual

SYNOPSIS
  pm connectors inspect okta
  pm connectors inspect okta --json
  pm credentials add <name> --connector okta [--config key=value] [--from-env field=ENV] [--value-stdin field]

DESCRIPTION
  Reads and writes documented Okta Admin Management API resources through the Okta REST APIs.

ICON
  asset: icons/okta.svg
  source: official
  review_status: official_verified
  review_url: https://developer.okta.com/docs/reference/

CAPABILITIES
  check=true catalog=true read=true write=true query=false
  Integration type: api

AUTHENTICATION
  Use pm credentials add with --from-env or --value-stdin for secret fields.

CONFIGURATION
  aaguid
  api_service_id
  api_token_id
  app_id
  app_instance_id
  app_name
  assignment_id
  auth_server_id
  authenticator_id
  base_url
  behavior_id
  brand_id
  bundle_id
  captcha_id
  claim_id
  client_id
  connection_id
  contact_type
  csr_id
  custom_telephony_provider_id
  customization_id
  device_assurance_id
  device_id
  device_integration_id
  domain
  domain_id
  email_domain_id
  email_server_id
  enrollment_id
  entitlement_id
  event_hook_id
  external_id
  factor_id
  feature_id
  feature_name
  filter
  grant_id
  group_id
  group_or_external_id
  group_rule_id
  id
  identity_source_id
  idp_csr_id
  idp_id
  inline_hook_id
  key_id
  kid
  linked_object_name
  log_stream_id
  log_stream_type
  mapping_id
  member_id
  method_type
  notification_type
  oauth_client_id
  path
  permission_type
  policy_id
  pool_id
  posture_check_id
  principal_rate_limit_id
  push_provider_id
  realm_id
  relationship_name
  resource_id
  resource_set_id_or_label
  result_id
  role_assignment_id
  role_id_or_encoded_role_id
  role_id_or_label
  role_ref
  rule_id
  schema_id
  scope_id
  secret_id
  security_event_provider_id
  session_id
  start_date
  stream_id
  template_id
  template_name
  theme_id
  token_id
  transaction_id
  trusted_origin_id
  type
  type_id
  update_id
  user_id
  user_id_or_login
  zone_id
  access_token (secret)
  api_token (secret)

ETL STREAMS
  users:
    primary key: id
    fields: created(), email(), id(), last_login(), login(), status()
  groups:
    primary key: id
    fields: created(), description(), id(), name()
  system_logs:
    primary key: uuid
    cursor: published
    fields: display_message(), event_type(), published(), uuid()
  well_known_app_authenticator_configuration:
    primary key: name
    fields: appAuthenticatorEnrollEndpoint(), authenticatorId(), createdDate(), key(), lastUpdated(), name(), orgId(), settings(), supportedMethods(), type()
  well_known_apple_app_site_association:
  well_known_assetlinks_json:
  well_known_okta_organization:
    primary key: id
    fields: _links(), id(), pipeline()
  well_known_ssf_configuration:
    fields: authorization_schemes(), configuration_endpoint(), default_subjects(), delivery_methods_supported(), issuer(), jwks_uri(), spec_version(), verification_endpoint()
  well_known_webauthn:
  api_v1_agent_pools:
    primary key: id
    fields: _links(), agents(), disruptedAgents(), id(), inactiveAgents(), name(), operationalStatus(), type()
  api_v1_agent_pools_pool_id_updates:
    primary key: id
    fields: _links(), agentType(), agents(), enabled(), id(), name(), notifyAdmin(), reason(), schedule(), sortOrder(), status(), targetVersion()
  api_v1_agent_pools_pool_id_updates_settings:
    fields: agentType(), continueOnError(), latestVersion(), minimalSupportedVersion(), poolId(), poolName(), releaseChannel()
  api_v1_agent_pools_pool_id_updates_update_id:
    primary key: id
    fields: _links(), agentType(), agents(), enabled(), id(), name(), notifyAdmin(), reason(), schedule(), sortOrder(), status(), targetVersion()
  api_v1_api_tokens:
    primary key: id
    fields: _link(), clientName(), created(), expiresAt(), id(), lastUpdated(), name(), network(), tokenWindow(), userId()
  api_v1_api_tokens_api_token_id:
    primary key: id
    fields: _link(), clientName(), created(), expiresAt(), id(), lastUpdated(), name(), network(), tokenWindow(), userId()
  api_v1_apps:
    primary key: id
    fields: _embedded(), _links(), accessibility(), created(), expressConfiguration(), features(), id(), label(), lastUpdated(), licensing(), orn(), profile(), signOnMode(), status(), universalLogout(), visibility()
  api_v1_apps_app_id:
    primary key: id
    fields: _embedded(), _links(), accessibility(), created(), expressConfiguration(), features(), id(), label(), lastUpdated(), licensing(), orn(), profile(), signOnMode(), status(), universalLogout(), visibility()
  api_v1_apps_app_id_connections_default:
    fields: _links(), authScheme(), baseUrl(), profile(), status()
  api_v1_apps_app_id_connections_default_jwks:
    fields: jwks()
  api_v1_apps_app_id_credentials_csrs:
    primary key: id
    fields: _links(), created(), csr(), id(), kty()
  api_v1_apps_app_id_credentials_csrs_csr_id:
    primary key: id
    fields: _links(), created(), csr(), id(), kty()
  api_v1_apps_app_id_credentials_jwks:
    primary key: id
    fields: _links(), created(), id(), lastUpdated()
  api_v1_apps_app_id_credentials_jwks_key_id:
    primary key: id
    fields: _links(), created(), id(), lastUpdated()
  api_v1_apps_app_id_credentials_keys:
    fields: created(), e(), expiresAt(), kid(), kty(), lastUpdated(), n(), use(), x5c(), x5t#S256()
  api_v1_apps_app_id_credentials_keys_key_id:
    fields: created(), e(), expiresAt(), kid(), kty(), lastUpdated(), n(), use(), x5c(), x5t#S256()
  api_v1_apps_app_id_credentials_secrets:
    primary key: id
    fields: _links(), client_secret(), created(), id(), lastUpdated(), secret_hash(), status()
  api_v1_apps_app_id_credentials_secrets_secret_id:
    primary key: id
    fields: _links(), client_secret(), created(), id(), lastUpdated(), secret_hash(), status()
  api_v1_apps_app_id_cwo_connections:
    primary key: id
    fields: created(), id(), lastUpdated(), requestingAppInstanceId(), resourceAppInstanceId(), status()
  api_v1_apps_app_id_cwo_connections_connection_id:
    primary key: id
    fields: created(), id(), lastUpdated(), requestingAppInstanceId(), resourceAppInstanceId(), status()
  api_v1_apps_app_id_features:
    primary key: name
    fields: _links(), description(), name(), status()
  api_v1_apps_app_id_features_feature_name:
    primary key: name
    fields: _links(), description(), name(), status()
  api_v1_apps_app_id_federated_claims:
    primary key: id
    fields: created(), expression(), id(), lastUpdated(), name()
  api_v1_apps_app_id_federated_claims_claim_id:
    primary key: name
    fields: expression(), name()
  api_v1_apps_app_id_grants:
    primary key: id
    fields: _embedded(), _links(), clientId(), created(), createdBy(), id(), issuer(), lastUpdated(), scopeId(), source(), status(), userId()
  api_v1_apps_app_id_grants_grant_id:
    primary key: id
    fields: _embedded(), _links(), clientId(), created(), createdBy(), id(), issuer(), lastUpdated(), scopeId(), source(), status(), userId()
  api_v1_apps_app_id_group_push_mappings:
    primary key: id
    fields: _links(), appConfig(), created(), errorSummary(), id(), lastPush(), lastUpdated(), sourceGroupId(), status(), targetGroupId()
  api_v1_apps_app_id_group_push_mappings_mapping_id:
    primary key: id
    fields: _links(), appConfig(), created(), errorSummary(), id(), lastPush(), lastUpdated(), sourceGroupId(), status(), targetGroupId()
  api_v1_apps_app_id_groups:
    primary key: id
    fields: _embedded(), _links(), id(), lastUpdated(), priority(), profile()
  api_v1_apps_app_id_groups_group_id:
    primary key: id
    fields: _embedded(), _links(), id(), lastUpdated(), priority(), profile()
  api_v1_apps_app_id_tokens:
    primary key: id
    fields: _embedded(), _links(), clientId(), created(), expiresAt(), id(), issuer(), lastUpdated(), scopes(), status(), userId()
  api_v1_apps_app_id_tokens_token_id:
    primary key: id
    fields: _embedded(), _links(), clientId(), created(), expiresAt(), id(), issuer(), lastUpdated(), scopes(), status(), userId()
  api_v1_apps_app_id_users:
    primary key: id
    fields: _embedded(), _links(), created(), credentials(), externalId(), id(), lastSync(), lastUpdated(), passwordChanged(), profile(), scope(), status(), statusChanged(), syncState()
  api_v1_apps_app_id_users_user_id:
    primary key: id
    fields: _embedded(), _links(), created(), credentials(), externalId(), id(), lastSync(), lastUpdated(), passwordChanged(), profile(), scope(), status(), statusChanged(), syncState()
  api_v1_authenticators:
    primary key: id
    fields: _links(), created(), description(), id(), key(), lastUpdated(), name(), status(), type()
  api_v1_authenticators_authenticator_id:
    primary key: id
    fields: _links(), created(), description(), id(), key(), lastUpdated(), name(), status(), type()
  api_v1_authenticators_authenticator_id_aaguids:
    primary key: name
    fields: _links(), aaguid(), attestationRootCertificates(), authenticatorCharacteristics(), name()
  api_v1_authenticators_authenticator_id_aaguids_aaguid:
    primary key: name
    fields: _links(), aaguid(), attestationRootCertificates(), authenticatorCharacteristics(), name()
  api_v1_authenticators_authenticator_id_methods:
    fields: _links(), status(), type()
  api_v1_authenticators_authenticator_id_methods_method_type:
    fields: _links(), status(), type()
  api_v1_authorization_servers:
    primary key: id
    fields: _links(), accessTokenEncryptedResponseAlgorithm(), audiences(), created(), credentials(), description(), id(), issuer(), issuerMode(), jwks(), jwks_uri(), lastUpdated(), name(), status()
  api_v1_authorization_servers_auth_server_id:
    primary key: id
    fields: _links(), accessTokenEncryptedResponseAlgorithm(), audiences(), created(), credentials(), description(), id(), issuer(), issuerMode(), jwks(), jwks_uri(), lastUpdated(), name(), status()
  api_v1_authorization_servers_auth_server_id_associated_servers:
    primary key: id
    fields: _links(), accessTokenEncryptedResponseAlgorithm(), audiences(), created(), credentials(), description(), id(), issuer(), issuerMode(), jwks(), jwks_uri(), lastUpdated(), name(), status()
  api_v1_authorization_servers_auth_server_id_claims:
    primary key: id
    fields: _links(), alwaysIncludeInToken(), claimType(), conditions(), group_filter_type(), id(), name(), status(), system(), value(), valueType()
  api_v1_authorization_servers_auth_server_id_claims_claim_id:
    primary key: id
    fields: _links(), alwaysIncludeInToken(), claimType(), conditions(), group_filter_type(), id(), name(), status(), system(), value(), valueType()
  api_v1_authorization_servers_auth_server_id_clients:
    fields: _links(), client_id(), client_name(), client_uri(), logo_uri()
  api_v1_authorization_servers_auth_server_id_clients_client_id_tokens:
    primary key: id
    fields: _embedded(), _links(), clientId(), created(), expiresAt(), id(), issuer(), lastUpdated(), scopes(), status(), userId()
  api_v1_authorization_servers_auth_server_id_clients_client_id_tokens_token_id:
    primary key: id
    fields: _embedded(), _links(), clientId(), created(), expiresAt(), id(), issuer(), lastUpdated(), scopes(), status(), userId()
  api_v1_authorization_servers_auth_server_id_credentials_keys:
    fields: _links(), alg(), e(), kid(), kty(), n(), status(), use()
  api_v1_authorization_servers_auth_server_id_credentials_keys_key_id:
    fields: _links(), alg(), e(), kid(), kty(), n(), status(), use()
  api_v1_authorization_servers_auth_server_id_policies:
    primary key: id
    fields: _links(), conditions(), created(), description(), id(), lastUpdated(), name(), priority(), status(), system(), type()
  api_v1_authorization_servers_auth_server_id_policies_policy_id:
    primary key: id
    fields: _links(), conditions(), created(), description(), id(), lastUpdated(), name(), priority(), status(), system(), type()
  api_v1_authorization_servers_auth_server_id_policies_policy_id_rules:
    primary key: id
    fields: _links(), actions(), conditions(), created(), id(), lastUpdated(), name(), priority(), status(), system(), type()
  api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id:
    primary key: id
    fields: _links(), actions(), conditions(), created(), id(), lastUpdated(), name(), priority(), status(), system(), type()
  api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys:
    primary key: id
    fields: _links(), created(), e(), id(), kid(), kty(), lastUpdated(), n(), status(), use()
  api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id:
    primary key: id
    fields: _links(), created(), e(), id(), kid(), kty(), lastUpdated(), n(), status(), use()
  api_v1_authorization_servers_auth_server_id_scopes:
    primary key: id
    fields: _links(), consent(), default(), description(), displayName(), id(), metadataPublish(), name(), optional(), system()
  api_v1_authorization_servers_auth_server_id_scopes_scope_id:
    primary key: id
    fields: _links(), consent(), default(), description(), displayName(), id(), metadataPublish(), name(), optional(), system()
  api_v1_behaviors:
    primary key: id
    fields: _link(), created(), id(), lastUpdated(), name(), status(), type()
  api_v1_behaviors_behavior_id:
    primary key: id
    fields: _link(), created(), id(), lastUpdated(), name(), status(), type()
  api_v1_bot_protection_configuration:
    fields: _links(), enforcementType(), level(), mode(), supportedFlows()
  api_v1_brands:
    primary key: id
    fields: agreeToCustomPrivacyPolicy(), customPrivacyPolicyUrl(), defaultApp(), emailDomainId(), id(), isDefault(), locale(), name(), removePoweredByOkta()
  api_v1_brands_brand_id:
    primary key: id
    fields: agreeToCustomPrivacyPolicy(), customPrivacyPolicyUrl(), defaultApp(), emailDomainId(), id(), isDefault(), locale(), name(), removePoweredByOkta()
  api_v1_brands_brand_id_domains:
    fields: domains()
  api_v1_brands_brand_id_pages_error:
    fields: _embedded(), _links()
  api_v1_brands_brand_id_pages_error_customized:
    fields: contentSecurityPolicySetting()
  api_v1_brands_brand_id_pages_error_default:
    fields: contentSecurityPolicySetting()
  api_v1_brands_brand_id_pages_error_preview:
    fields: contentSecurityPolicySetting()
  api_v1_brands_brand_id_pages_sign_in:
    fields: _embedded(), _links()
  api_v1_brands_brand_id_pages_sign_in_customized:
    fields: contentSecurityPolicySetting(), widgetCustomizations(), widgetVersion()
  api_v1_brands_brand_id_pages_sign_in_default:
    fields: contentSecurityPolicySetting(), widgetCustomizations(), widgetVersion()
  api_v1_brands_brand_id_pages_sign_in_preview:
    fields: contentSecurityPolicySetting(), widgetCustomizations(), widgetVersion()
  api_v1_brands_brand_id_pages_sign_out_customized:
    fields: type(), url()
  api_v1_brands_brand_id_templates_email:
    primary key: name
    fields: _embedded(), _links(), name()
  api_v1_brands_brand_id_templates_email_template_name:
    primary key: name
    fields: _embedded(), _links(), name()
  api_v1_brands_brand_id_templates_email_template_name_customizations:
    primary key: id
    fields: _links(), created(), id(), isDefault(), language(), lastUpdated()
  api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id:
    primary key: id
    fields: _links(), created(), id(), isDefault(), language(), lastUpdated()
  api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id_preview:
    fields: _links(), body(), subject()
  api_v1_brands_brand_id_templates_email_template_name_default_content:
    fields: _links()
  api_v1_brands_brand_id_templates_email_template_name_default_content_preview:
    fields: _links(), body(), subject()
  api_v1_brands_brand_id_templates_email_template_name_settings:
    fields: _links(), recipients()
  api_v1_brands_brand_id_themes:
    primary key: id
    fields: _links(), backgroundImage(), emailTemplateTouchPointVariant(), endUserDashboardTouchPointVariant(), errorPageTouchPointVariant(), favicon(), id(), loadingPageTouchPointVariant(), logo(), primaryColorContrastHex(), primaryColorHex(), secondaryColorContrastHex(), secondaryColorHex(), signInPageTouchPointVariant()
  api_v1_brands_brand_id_themes_theme_id:
    primary key: id
    fields: _links(), backgroundImage(), emailTemplateTouchPointVariant(), endUserDashboardTouchPointVariant(), errorPageTouchPointVariant(), favicon(), id(), loadingPageTouchPointVariant(), logo(), primaryColorContrastHex(), primaryColorHex(), secondaryColorContrastHex(), secondaryColorHex(), signInPageTouchPointVariant()
  api_v1_brands_brand_id_well_known_uris:
    fields: _embedded(), _links()
  api_v1_brands_brand_id_well_known_uris_path:
    fields: _links(), representation()
  api_v1_brands_brand_id_well_known_uris_path_customized:
    fields: _links(), representation()
  api_v1_captchas:
    primary key: id
    fields: _links(), id(), name(), secretKey(), siteKey(), type()
  api_v1_captchas_captcha_id:
    primary key: id
    fields: _links(), id(), name(), secretKey(), siteKey(), type()
  api_v1_device_assurances:
    primary key: id
    fields: _links(), createdBy(), createdDate(), devicePostureChecks(), displayRemediationMode(), gracePeriod(), id(), lastUpdate(), lastUpdatedBy(), name(), platform()
  api_v1_device_assurances_device_assurance_id:
    primary key: id
    fields: _links(), createdBy(), createdDate(), devicePostureChecks(), displayRemediationMode(), gracePeriod(), id(), lastUpdate(), lastUpdatedBy(), name(), platform()
  api_v1_device_integrations:
    primary key: id
    fields: _links(), displayName(), id(), metadata(), name(), platform(), status()
  api_v1_device_integrations_device_integration_id:
    primary key: id
    fields: _links(), displayName(), id(), metadata(), name(), platform(), status()
  api_v1_device_posture_checks:
    primary key: id
    fields: _links(), createdBy(), createdDate(), description(), id(), lastUpdate(), lastUpdatedBy(), mappingType(), name(), platform(), query(), remediationSettings(), type(), variableName()
  api_v1_device_posture_checks_default:
    primary key: id
    fields: _links(), createdBy(), createdDate(), description(), id(), lastUpdate(), lastUpdatedBy(), mappingType(), name(), platform(), query(), remediationSettings(), type(), variableName()
  api_v1_device_posture_checks_posture_check_id:
    primary key: id
    fields: _links(), createdBy(), createdDate(), description(), id(), lastUpdate(), lastUpdatedBy(), mappingType(), name(), platform(), query(), remediationSettings(), type(), variableName()
  api_v1_devices:
    fields: _embedded()
  api_v1_devices_device_id:
    fields: providers()
  api_v1_devices_device_id_users:
    fields: created(), managementStatus(), screenLockType(), user()
  api_v1_directories_app_instance_id_groups_group_id_query_result_id:
    primary key: id
    fields: id(), profile()
  api_v1_domains:
    fields: domains()
  api_v1_domains_domain_id:
    primary key: id
    fields: _links(), brandId(), certificateSourceType(), dnsRecords(), domain(), id(), publicCertificate(), validationStatus()
  api_v1_dr_status:
    fields: status()
  api_v1_dr_status_domain:
    fields: status()
  api_v1_email_domains:
    fields: displayName(), userName()
  api_v1_email_domains_email_domain_id:
    fields: displayName(), userName()
  api_v1_email_servers:
    fields: email-servers()
  api_v1_email_servers_email_server_id:
    primary key: id
    fields: alias(), authType(), enabled(), host(), id(), port(), username()
  api_v1_event_hooks:
    primary key: id
    fields: _links(), channel(), created(), createdBy(), description(), events(), id(), lastUpdated(), name(), status(), verificationStatus()
  api_v1_event_hooks_event_hook_id:
    primary key: id
    fields: _links(), channel(), created(), createdBy(), description(), events(), id(), lastUpdated(), name(), status(), verificationStatus()
  api_v1_features:
    primary key: id
    fields: _links(), description(), id(), name(), stage(), status(), type()
  api_v1_features_feature_id:
    primary key: id
    fields: _links(), description(), id(), name(), stage(), status(), type()
  api_v1_features_feature_id_dependencies:
    primary key: id
    fields: _links(), description(), id(), name(), stage(), status(), type()
  api_v1_features_feature_id_dependents:
    primary key: id
    fields: _links(), description(), id(), name(), stage(), status(), type()
  api_v1_first_party_app_settings_app_name:
    fields: sessionIdleTimeoutMinutes(), sessionMaxLifetimeMinutes()
  api_v1_groups_rules:
    primary key: id
    fields: _embedded(), actions(), conditions(), created(), id(), lastUpdated(), name(), status(), type()
  api_v1_groups_rules_group_rule_id:
    primary key: id
    fields: _embedded(), actions(), conditions(), created(), id(), lastUpdated(), name(), status(), type()
  api_v1_groups_group_id:
    primary key: id
    fields: _embedded(), _links(), created(), id(), lastMembershipUpdated(), lastUpdated(), objectClass(), profile(), type()
  api_v1_groups_group_id_apps:
    primary key: id
    fields: _embedded(), _links(), accessibility(), created(), expressConfiguration(), features(), id(), label(), lastUpdated(), licensing(), orn(), profile(), signOnMode(), status(), universalLogout(), visibility()
  api_v1_groups_group_id_owners:
    primary key: id
    fields: displayName(), id(), lastUpdated(), originId(), originType(), resolved(), type()
  api_v1_groups_group_id_roles:
    primary key: id
    fields: _embedded(), _links(), assignmentType(), created(), id(), label(), lastUpdated(), status(), type()
  api_v1_groups_group_id_roles_role_assignment_id:
    primary key: id
    fields: _embedded(), _links(), assignmentType(), created(), id(), label(), lastUpdated(), status(), type()
  api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps:
    primary key: id
    fields: _links(), category(), description(), displayName(), features(), id(), lastUpdated(), name(), signOnModes(), status(), verificationStatus(), website()
  api_v1_groups_group_id_roles_role_assignment_id_targets_groups:
    primary key: id
    fields: _embedded(), _links(), created(), id(), lastMembershipUpdated(), lastUpdated(), objectClass(), profile(), type()
  api_v1_groups_group_id_users:
    primary key: id
    fields: _embedded(), _links(), activated(), created(), credentials(), id(), lastLogin(), lastUpdated(), passwordChanged(), profile(), realmId(), status(), statusChanged(), transitioningToStatus(), type()
  api_v1_hook_keys:
    primary key: id
    fields: created(), id(), isUsed(), keyId(), lastUpdated(), name()
  api_v1_hook_keys_public_key_id:
    fields: alg(), e(), kid(), kty(), n(), use()
  api_v1_hook_keys_id:
    primary key: id
    fields: created(), id(), isUsed(), keyId(), lastUpdated(), name()
  api_v1_iam_assignees_users:
    primary key: id
    fields: _links(), id(), orn()
  api_v1_iam_governance_bundles:
    fields: _links(), bundles()
  api_v1_iam_governance_bundles_bundle_id:
    primary key: id
    fields: _links(), description(), id(), name(), orn(), status()
  api_v1_iam_governance_bundles_bundle_id_entitlements:
    fields: _links(), entitlements()
  api_v1_iam_governance_bundles_bundle_id_entitlements_entitlement_id_values:
    fields: _links(), entitlementValues()
  api_v1_iam_governance_opt_in:
    fields: _links(), optInStatus()
  api_v1_iam_resource_sets:
    fields: _links(), resource-sets()
  api_v1_iam_resource_sets_resource_set_id_or_label:
    primary key: id
    fields: _links(), created(), description(), id(), label(), lastUpdated()
  api_v1_iam_resource_sets_resource_set_id_or_label_bindings:
    primary key: id
    fields: _links(), id()
  api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label:
    primary key: id
    fields: _links(), id()
  api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members:
    fields: _links(), members()
  api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members_member_id:
    primary key: id
    fields: _links(), created(), id(), lastUpdated()
  api_v1_iam_resource_sets_resource_set_id_or_label_resources:
    fields: _links(), resources()
  api_v1_iam_resource_sets_resource_set_id_or_label_resources_resource_id:
    primary key: id
    fields: _links(), conditions(), created(), id(), lastUpdated(), orn()
  api_v1_iam_roles:
    primary key: id
    fields: _links(), created(), description(), id(), label(), lastUpdated()
  api_v1_iam_roles_role_id_or_label:
    primary key: id
    fields: _links(), created(), description(), id(), label(), lastUpdated()
  api_v1_iam_roles_role_id_or_label_permissions:
    fields: _links(), conditions(), created(), label(), lastUpdated()
  api_v1_iam_roles_role_id_or_label_permissions_permission_type:
    fields: _links(), conditions(), created(), label(), lastUpdated()
  api_v1_identity_sources_identity_source_id_groups_group_or_external_id:
    primary key: id
    fields: externalId(), id(), profile()
  api_v1_identity_sources_identity_source_id_groups_group_or_external_id_membership:
    fields: memberExternalIds()
  api_v1_identity_sources_identity_source_id_sessions:
    primary key: id
    fields: created(), id(), identitySourceId(), importType(), lastUpdated(), status()
  api_v1_identity_sources_identity_source_id_sessions_session_id:
    primary key: id
    fields: created(), id(), identitySourceId(), importType(), lastUpdated(), status()
  api_v1_identity_sources_identity_source_id_users_external_id:
    primary key: id
    fields: created(), externalId(), id(), lastUpdated(), profile()
  api_v1_idps:
    primary key: id
    fields: _links(), created(), id(), issuerMode(), lastUpdated(), name(), policy(), properties(), protocol(), status(), type()
  api_v1_idps_credentials_keys:
    fields: created(), e(), expiresAt(), kid(), kty(), lastUpdated(), n(), use(), x5c(), x5t#S256()
  api_v1_idps_credentials_keys_kid:
    fields: created(), e(), expiresAt(), kid(), kty(), lastUpdated(), n(), use(), x5c(), x5t#S256()
  api_v1_idps_idp_id:
    primary key: id
    fields: _links(), created(), id(), issuerMode(), lastUpdated(), name(), policy(), properties(), protocol(), status(), type()
  api_v1_idps_idp_id_credentials_csrs:
    primary key: id
    fields: _links(), created(), csr(), id(), kty()
  api_v1_idps_idp_id_credentials_csrs_idp_csr_id:
    primary key: id
    fields: _links(), created(), csr(), id(), kty()
  api_v1_idps_idp_id_credentials_keys:
    fields: created(), e(), expiresAt(), kid(), kty(), lastUpdated(), n(), use(), x5c(), x5t#S256()
  api_v1_idps_idp_id_credentials_keys_active:
    fields: created(), e(), expiresAt(), kid(), kty(), lastUpdated(), n(), use(), x5c(), x5t#S256()
  api_v1_idps_idp_id_credentials_keys_kid:
    fields: created(), e(), expiresAt(), kid(), kty(), lastUpdated(), n(), use(), x5c(), x5t#S256()
  api_v1_idps_idp_id_users:
    primary key: id
    fields: _embedded(), _links(), created(), externalId(), id(), lastUpdated(), profile()
  api_v1_idps_idp_id_users_user_id:
    primary key: id
    fields: _embedded(), _links(), created(), externalId(), id(), lastUpdated(), profile()
  api_v1_idps_idp_id_users_user_id_credentials_tokens:
    primary key: id
    fields: expiresAt(), id(), scopes(), token(), tokenAuthScheme(), tokenType()
  api_v1_inline_hooks:
    primary key: id
    fields: _links(), channel(), created(), id(), lastUpdated(), name(), status(), type(), version()
  api_v1_inline_hooks_inline_hook_id:
    primary key: id
    fields: _links(), channel(), created(), id(), lastUpdated(), name(), status(), type(), version()
  api_v1_log_streams:
    primary key: id
    fields: _links(), created(), id(), lastUpdated(), name(), status(), type()
  api_v1_log_streams_log_stream_id:
    primary key: id
    fields: _links(), created(), id(), lastUpdated(), name(), status(), type()
  api_v1_mappings:
    primary key: id
    fields: _links(), id(), source(), target()
  api_v1_mappings_mapping_id:
    primary key: id
    fields: _links(), id(), properties(), source(), target()
  api_v1_meta_schemas_apps_app_id_default:
    primary key: id
    fields: $schema(), _links(), created(), definitions(), id(), lastUpdated(), name(), properties(), title(), type()
  api_v1_meta_schemas_group_default:
    primary key: id
    fields: $schema(), _links(), created(), definitions(), description(), id(), lastUpdated(), name(), properties(), title(), type()
  api_v1_meta_schemas_log_stream:
    primary key: id
    fields: $schema(), _links(), errorMessage(), id(), oneOf(), pattern(), properties(), required(), title(), type()
  api_v1_meta_schemas_log_stream_log_stream_type:
    primary key: id
    fields: $schema(), _links(), errorMessage(), id(), oneOf(), pattern(), properties(), required(), title(), type()
  api_v1_meta_schemas_user_linked_objects:
    fields: _links(), associated(), primary()
  api_v1_meta_schemas_user_linked_objects_linked_object_name:
    fields: _links(), associated(), primary()
  api_v1_meta_schemas_user_schema_id:
    primary key: id
    fields: $schema(), _links(), created(), definitions(), id(), lastUpdated(), name(), properties(), title(), type()
  api_v1_meta_types_user:
    primary key: id
    fields: _links(), created(), createdBy(), default(), description(), displayName(), id(), lastUpdated(), lastUpdatedBy(), name()
  api_v1_meta_types_user_type_id:
    primary key: id
    fields: _links(), created(), createdBy(), default(), description(), displayName(), id(), lastUpdated(), lastUpdatedBy(), name()
  api_v1_meta_uischemas:
    primary key: id
    fields: _links(), created(), id(), lastUpdated(), uiSchema()
  api_v1_meta_uischemas_id:
    primary key: id
    fields: _links(), created(), id(), lastUpdated(), uiSchema()
  api_v1_org:
    primary key: id
    fields: _links(), address1(), address2(), city(), companyName(), country(), created(), endUserSupportHelpURL(), expiresAt(), id(), lastUpdated(), phoneNumber(), postalCode(), state(), status(), subdomain(), supportPhoneNumber(), website()
  api_v1_org_captcha:
    fields: _links(), captchaId(), enabledPages()
  api_v1_org_contacts:
    fields: _links(), contactType()
  api_v1_org_contacts_contact_type:
    fields: _links(), userId()
  api_v1_org_factors_yubikey_token_tokens:
    primary key: id
    fields: _embedded(), _links(), created(), id(), lastUpdated(), lastVerified(), profile(), status()
  api_v1_org_factors_yubikey_token_tokens_token_id:
    primary key: id
    fields: _embedded(), _links(), created(), id(), lastUpdated(), lastVerified(), profile(), status()
  api_v1_org_org_settings_third_party_admin_setting:
    fields: thirdPartyAdmin()
  api_v1_org_preferences:
    fields: _links(), showEndUserFooter()
  api_v1_org_privacy_aerial:
    fields: _links(), accountId(), grantedBy(), grantedDate()
  api_v1_org_privacy_okta_communication:
    fields: _links(), optOutEmailUsers()
  api_v1_org_privacy_okta_support:
    fields: _links(), caseNumber(), expiration(), support()
  api_v1_org_privacy_okta_support_cases:
    fields: supportCases()
  api_v1_org_settings_auto_assign_admin_app_setting:
    fields: autoAssignAdminAppSetting()
  api_v1_org_settings_client_privileges_setting:
    fields: clientPrivilegesSetting()
  api_v1_policies:
    primary key: id
    fields: _embedded(), _links(), created(), description(), id(), lastUpdated(), name(), priority(), status(), system(), type()
  api_v1_policies_policy_id:
    primary key: id
    fields: _embedded(), _links(), created(), description(), id(), lastUpdated(), name(), priority(), status(), system(), type()
  api_v1_policies_policy_id_app:
    primary key: id
    fields: _embedded(), _links(), accessibility(), created(), expressConfiguration(), features(), id(), label(), lastUpdated(), licensing(), orn(), profile(), signOnMode(), status(), universalLogout(), visibility()
  api_v1_policies_policy_id_mappings:
    primary key: id
    fields: _links(), id()
  api_v1_policies_policy_id_mappings_mapping_id:
    primary key: id
    fields: _links(), id()
  api_v1_policies_policy_id_rules:
    primary key: id
    fields: _links(), created(), id(), lastUpdated(), name(), priority(), status(), system(), type()
  api_v1_policies_policy_id_rules_rule_id:
    primary key: id
    fields: _links(), created(), id(), lastUpdated(), name(), priority(), status(), system(), type()
  api_v1_principal_rate_limits:
    primary key: id
    fields: createdBy(), createdDate(), defaultConcurrencyPercentage(), defaultPercentage(), id(), lastUpdate(), lastUpdatedBy(), orgId(), principalId(), principalType()
  api_v1_principal_rate_limits_principal_rate_limit_id:
    primary key: id
    fields: createdBy(), createdDate(), defaultConcurrencyPercentage(), defaultPercentage(), id(), lastUpdate(), lastUpdatedBy(), orgId(), principalId(), principalType()
  api_v1_push_providers:
    primary key: id
    fields: _links(), id(), lastUpdatedDate(), name(), providerType()
  api_v1_push_providers_push_provider_id:
    primary key: id
    fields: _links(), id(), lastUpdatedDate(), name(), providerType()
  api_v1_rate_limit_settings_admin_notifications:
    fields: notificationsEnabled()
  api_v1_rate_limit_settings_per_client:
    fields: defaultMode(), useCaseModeOverrides()
  api_v1_rate_limit_settings_warning_threshold:
    fields: warningThreshold()
  api_v1_realm_assignments:
    primary key: id
    fields: _links(), actions(), conditions(), created(), domains(), id(), isDefault(), lastUpdated(), name(), priority(), status()
  api_v1_realm_assignments_operations:
    fields: _links(), assignmentOperation(), numUserMoved(), realmId(), realmName()
  api_v1_realm_assignments_assignment_id:
    primary key: id
    fields: _links(), actions(), conditions(), created(), domains(), id(), isDefault(), lastUpdated(), name(), priority(), status()
  api_v1_realms:
    primary key: id
    fields: _links(), created(), id(), isDefault(), lastUpdated(), profile()
  api_v1_realms_realm_id:
    primary key: id
    fields: _links(), created(), id(), isDefault(), lastUpdated(), profile()
  api_v1_roles_role_ref_subscriptions:
    fields: _links(), channels(), notificationType(), status()
  api_v1_roles_role_ref_subscriptions_notification_type:
    fields: _links(), channels(), notificationType(), status()
  api_v1_security_events_providers:
    primary key: id
    fields: _links(), id(), name(), settings(), status(), type()
  api_v1_security_events_providers_security_event_provider_id:
    primary key: id
    fields: _links(), id(), name(), settings(), status(), type()
  api_v1_sessions_session_id:
    primary key: id
    fields: _links(), amr(), createdAt(), expiresAt(), id(), idp(), lastFactorVerification(), lastPasswordVerification(), login(), status(), userId()
  api_v1_ssf_stream:
    fields: aud(), delivery(), events_delivered(), events_requested(), events_supported(), format(), iss(), min_verification_interval(), stream_id()
  api_v1_ssf_stream_status:
    fields: status(), stream_id()
  api_v1_telephony_providers:
    primary key: id
    fields: enabled(), id(), isPrimaryProvider(), providerCapability(), providerName(), providerSettings(), providerSid()
  api_v1_telephony_providers_custom_telephony_provider_id:
    primary key: id
    fields: enabled(), id(), isPrimaryProvider(), providerCapability(), providerName(), providerSettings(), providerSid()
  api_v1_templates_sms:
    primary key: id
    fields: created(), id(), lastUpdated(), name(), template(), translations(), type()
  api_v1_templates_sms_template_id:
    primary key: id
    fields: created(), id(), lastUpdated(), name(), template(), translations(), type()
  api_v1_threats_configuration:
    fields: _links(), action(), created(), excludeZones(), lastUpdated()
  api_v1_trusted_origins:
    primary key: id
    fields: _links(), created(), createdBy(), id(), lastUpdated(), lastUpdatedBy(), name(), origin(), scopes(), status()
  api_v1_trusted_origins_trusted_origin_id:
    fields: allowedOktaApps(), type()
  api_v1_users_id:
    fields: _embedded()
  api_v1_users_id_app_links:
    primary key: id
    fields: appAssignmentId(), appInstanceId(), appName(), credentialsSetup(), hidden(), id(), label(), linkUrl(), logoUrl(), sortOrder()
  api_v1_users_id_blocks:
    fields: appliesTo(), type()
  api_v1_users_id_groups:
    primary key: id
    fields: _embedded(), _links(), created(), id(), lastMembershipUpdated(), lastUpdated(), objectClass(), profile(), type()
  api_v1_users_id_idps:
    primary key: id
    fields: _links(), created(), id(), issuerMode(), lastUpdated(), name(), policy(), properties(), protocol(), status(), type()
  api_v1_users_user_id_or_login_linked_objects_relationship_name:
    fields: _links()
  api_v1_users_user_id_authenticator_enrollments:
    primary key: id
    fields: _links(), created(), id(), key(), lastUpdated(), name(), profile(), status(), type()
  api_v1_users_user_id_authenticator_enrollments_enrollment_id:
    primary key: id
    fields: _links(), created(), id(), key(), lastUpdated(), name(), profile(), status(), type()
  api_v1_users_user_id_classification:
    fields: lastUpdated(), type()
  api_v1_users_user_id_clients:
    fields: _links(), client_id(), client_name(), client_uri(), logo_uri()
  api_v1_users_user_id_clients_client_id_grants:
    primary key: id
    fields: _embedded(), _links(), clientId(), created(), createdBy(), id(), issuer(), lastUpdated(), scopeId(), source(), status(), userId()
  api_v1_users_user_id_clients_client_id_tokens:
    primary key: id
    fields: _embedded(), _links(), clientId(), created(), expiresAt(), id(), issuer(), lastUpdated(), scopes(), status(), userId()
  api_v1_users_user_id_clients_client_id_tokens_token_id:
    primary key: id
    fields: _embedded(), _links(), clientId(), created(), expiresAt(), id(), issuer(), lastUpdated(), scopes(), status(), userId()
  api_v1_users_user_id_devices:
    fields: created(), device(), deviceUserId()
  api_v1_users_user_id_factors:
    primary key: id
    fields: _embedded(), _links(), created(), factorType(), id(), lastUpdated(), profile(), provider(), status(), vendorName()
  api_v1_users_user_id_factors_catalog:
    fields: _embedded(), _links(), enrollment(), factorType(), provider(), status(), vendorName()
  api_v1_users_user_id_factors_questions:
    fields: answer(), question(), questionText()
  api_v1_users_user_id_factors_factor_id:
    primary key: id
    fields: _embedded(), _links(), created(), factorType(), id(), lastUpdated(), profile(), provider(), status(), vendorName()
  api_v1_users_user_id_factors_factor_id_transactions_transaction_id:
    fields: factorResult()
  api_v1_users_user_id_grants:
    primary key: id
    fields: _embedded(), _links(), clientId(), created(), createdBy(), id(), issuer(), lastUpdated(), scopeId(), source(), status(), userId()
  api_v1_users_user_id_grants_grant_id:
    primary key: id
    fields: _embedded(), _links(), clientId(), created(), createdBy(), id(), issuer(), lastUpdated(), scopeId(), source(), status(), userId()
  api_v1_users_user_id_risk:
    fields: _links(), riskLevel()
  api_v1_users_user_id_roles:
    primary key: id
    fields: _embedded(), _links(), assignmentType(), created(), id(), label(), lastUpdated(), status(), type()
  api_v1_users_user_id_roles_role_assignment_id:
    primary key: id
    fields: _embedded(), _links(), assignmentType(), created(), id(), label(), lastUpdated(), status(), type()
  api_v1_users_user_id_roles_role_assignment_id_governance:
    fields: _links(), grants()
  api_v1_users_user_id_roles_role_assignment_id_governance_grant_id:
    fields: _links(), bundleId(), expirationDate(), grantId(), type()
  api_v1_users_user_id_roles_role_assignment_id_governance_grant_id_resources:
    fields: _links(), resources()
  api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps:
    primary key: id
    fields: _links(), category(), description(), displayName(), features(), id(), lastUpdated(), name(), signOnModes(), status(), verificationStatus(), website()
  api_v1_users_user_id_roles_role_assignment_id_targets_groups:
    primary key: id
    fields: _embedded(), _links(), created(), id(), lastMembershipUpdated(), lastUpdated(), objectClass(), profile(), type()
  api_v1_users_user_id_roles_role_id_or_encoded_role_id_targets:
    fields: _links(), assignmentType(), expiration(), orn()
  api_v1_users_user_id_subscriptions:
    fields: _links(), channels(), notificationType(), status()
  api_v1_users_user_id_subscriptions_notification_type:
    fields: _links(), channels(), notificationType(), status()
  api_v1_zones:
    primary key: id
    fields: _links(), created(), id(), lastUpdated(), name(), status(), system(), type(), usage()
  api_v1_zones_zone_id:
    primary key: id
    fields: _links(), created(), id(), lastUpdated(), name(), status(), system(), type(), usage()
  attack_protection_api_v1_authenticator_settings:
    fields: verifyKnowledgeSecondWhen2faRequired()
  attack_protection_api_v1_user_lockout_settings:
    fields: preventBruteForceLockoutFromUnknownDevices()
  integrations_api_v1_api_services:
    primary key: id
    fields: _links(), configGuideUrl(), createdAt(), createdBy(), grantedScopes(), id(), name(), properties(), type()
  integrations_api_v1_api_services_api_service_id:
    primary key: id
    fields: _links(), configGuideUrl(), createdAt(), createdBy(), grantedScopes(), id(), name(), properties(), type()
  integrations_api_v1_api_services_api_service_id_credentials_secrets:
    primary key: id
    fields: _links(), client_secret(), created(), id(), lastUpdated(), secret_hash(), status()
  oauth2_v1_clients_client_id_roles:
    primary key: id
    fields: _embedded(), _links(), assignmentType(), created(), id(), label(), lastUpdated(), status(), type()
  oauth2_v1_clients_client_id_roles_role_assignment_id:
    primary key: id
    fields: _embedded(), _links(), assignmentType(), created(), id(), label(), lastUpdated(), status(), type()
  oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps:
    primary key: id
    fields: _links(), category(), description(), displayName(), features(), id(), lastUpdated(), name(), signOnModes(), status(), verificationStatus(), website()
  oauth2_v1_clients_client_id_roles_role_assignment_id_targets_groups:
    primary key: id
    fields: _embedded(), _links(), created(), id(), lastMembershipUpdated(), lastUpdated(), objectClass(), profile(), type()
  okta_personal_settings_api_v1_export_blocklists:
    fields: domains()
  privileged_access_api_v1_okta_service_accounts:
    primary key: id
    fields: created(), description(), email(), id(), lastUpdated(), name(), oktaUserId(), ownerGroupIds(), ownerUserIds(), status(), statusDetail(), username()
  privileged_access_api_v1_okta_service_accounts_id:
    primary key: id
    fields: created(), description(), email(), id(), lastUpdated(), name(), oktaUserId(), ownerGroupIds(), ownerUserIds(), status(), statusDetail(), username()
  privileged_access_api_v1_service_accounts:
    primary key: id
    fields: containerGlobalName(), containerInstanceName(), containerOrn(), created(), description(), id(), lastUpdated(), name(), ownerGroupIds(), ownerUserIds(), password(), status(), statusDetail(), username()
  privileged_access_api_v1_service_accounts_id:
    primary key: id
    fields: containerGlobalName(), containerInstanceName(), containerOrn(), created(), description(), id(), lastUpdated(), name(), ownerGroupIds(), ownerUserIds(), password(), status(), statusDetail(), username()
  webauthn_registration_api_v1_users_user_id_enrollments:
    primary key: id
    fields: _links(), created(), factorType(), id(), lastUpdated(), profile(), provider(), status(), vendorName()

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_api_v1_agent_pools_pool_id_updates:
    endpoint: POST /api/v1/agentPools/{{ record.pool_id }}/updates
    required fields: pool_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_agent_pools_pool_id_updates_settings:
    endpoint: POST /api/v1/agentPools/{{ record.pool_id }}/updates/settings
    required fields: pool_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_agent_pools_pool_id_updates_update_id:
    endpoint: POST /api/v1/agentPools/{{ record.pool_id }}/updates/{{ record.update_id }}
    required fields: pool_id, update_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_agent_pools_pool_id_updates_update_id:
    endpoint: DELETE /api/v1/agentPools/{{ record.pool_id }}/updates/{{ record.update_id }}
    required fields: pool_id, update_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_agent_pools_pool_id_updates_update_id_activate:
    endpoint: POST /api/v1/agentPools/{{ record.pool_id }}/updates/{{ record.update_id }}/activate
    required fields: pool_id, update_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_agent_pools_pool_id_updates_update_id_deactivate:
    endpoint: POST /api/v1/agentPools/{{ record.pool_id }}/updates/{{ record.update_id }}/deactivate
    required fields: pool_id, update_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_agent_pools_pool_id_updates_update_id_pause:
    endpoint: POST /api/v1/agentPools/{{ record.pool_id }}/updates/{{ record.update_id }}/pause
    required fields: pool_id, update_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_agent_pools_pool_id_updates_update_id_resume:
    endpoint: POST /api/v1/agentPools/{{ record.pool_id }}/updates/{{ record.update_id }}/resume
    required fields: pool_id, update_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_agent_pools_pool_id_updates_update_id_retry:
    endpoint: POST /api/v1/agentPools/{{ record.pool_id }}/updates/{{ record.update_id }}/retry
    required fields: pool_id, update_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_agent_pools_pool_id_updates_update_id_stop:
    endpoint: POST /api/v1/agentPools/{{ record.pool_id }}/updates/{{ record.update_id }}/stop
    required fields: pool_id, update_id
    risk: high: external Okta admin API mutation; approval required
  delete_api_v1_api_tokens_current:
    endpoint: DELETE /api/v1/api-tokens/current
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_api_tokens_api_token_id:
    endpoint: PUT /api/v1/api-tokens/{{ record.api_token_id }}
    required fields: api_token_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_api_tokens_api_token_id:
    endpoint: DELETE /api/v1/api-tokens/{{ record.api_token_id }}
    required fields: api_token_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_apps:
    endpoint: POST /api/v1/apps
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_apps_app_id:
    endpoint: PUT /api/v1/apps/{{ record.app_id }}
    required fields: app_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_apps_app_id:
    endpoint: DELETE /api/v1/apps/{{ record.app_id }}
    required fields: app_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_apps_app_id_connections_default:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/connections/default
    required fields: app_id
    risk: medium: external Okta admin API mutation; approval required
  execute_api_v1_apps_app_id_connections_default_lifecycle_activate:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/connections/default/lifecycle/activate
    required fields: app_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_apps_app_id_connections_default_lifecycle_deactivate:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/connections/default/lifecycle/deactivate
    required fields: app_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_apps_app_id_credentials_csrs:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/credentials/csrs
    required fields: app_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_apps_app_id_credentials_csrs_csr_id:
    endpoint: DELETE /api/v1/apps/{{ record.app_id }}/credentials/csrs/{{ record.csr_id }}
    required fields: app_id, csr_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_apps_app_id_credentials_jwks:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/credentials/jwks
    required fields: app_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_apps_app_id_credentials_jwks_key_id:
    endpoint: DELETE /api/v1/apps/{{ record.app_id }}/credentials/jwks/{{ record.key_id }}
    required fields: app_id, key_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_apps_app_id_credentials_jwks_key_id_lifecycle_activate:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/credentials/jwks/{{ record.key_id }}/lifecycle/activate
    required fields: app_id, key_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_apps_app_id_credentials_jwks_key_id_lifecycle_deactivate:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/credentials/jwks/{{ record.key_id }}/lifecycle/deactivate
    required fields: app_id, key_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_apps_app_id_credentials_secrets:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/credentials/secrets
    required fields: app_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_apps_app_id_credentials_secrets_secret_id:
    endpoint: DELETE /api/v1/apps/{{ record.app_id }}/credentials/secrets/{{ record.secret_id }}
    required fields: app_id, secret_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_apps_app_id_credentials_secrets_secret_id_lifecycle_activate:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/credentials/secrets/{{ record.secret_id }}/lifecycle/activate
    required fields: app_id, secret_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_apps_app_id_credentials_secrets_secret_id_lifecycle_deactivate:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/credentials/secrets/{{ record.secret_id }}/lifecycle/deactivate
    required fields: app_id, secret_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_apps_app_id_cwo_connections:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/cwo/connections
    required fields: app_id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_apps_app_id_cwo_connections_connection_id:
    endpoint: PATCH /api/v1/apps/{{ record.app_id }}/cwo/connections/{{ record.connection_id }}
    required fields: app_id, connection_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_apps_app_id_cwo_connections_connection_id:
    endpoint: DELETE /api/v1/apps/{{ record.app_id }}/cwo/connections/{{ record.connection_id }}
    required fields: app_id, connection_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_apps_app_id_features_feature_name:
    endpoint: PUT /api/v1/apps/{{ record.app_id }}/features/{{ record.feature_name }}
    required fields: app_id, feature_name
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_apps_app_id_federated_claims:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/federated-claims
    required fields: app_id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_apps_app_id_federated_claims_claim_id:
    endpoint: PUT /api/v1/apps/{{ record.app_id }}/federated-claims/{{ record.claim_id }}
    required fields: app_id, claim_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_apps_app_id_federated_claims_claim_id:
    endpoint: DELETE /api/v1/apps/{{ record.app_id }}/federated-claims/{{ record.claim_id }}
    required fields: app_id, claim_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_apps_app_id_grants:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/grants
    required fields: app_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_apps_app_id_grants_grant_id:
    endpoint: DELETE /api/v1/apps/{{ record.app_id }}/grants/{{ record.grant_id }}
    required fields: app_id, grant_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_apps_app_id_group_push_mappings:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/group-push/mappings
    required fields: app_id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_apps_app_id_group_push_mappings_mapping_id:
    endpoint: PATCH /api/v1/apps/{{ record.app_id }}/group-push/mappings/{{ record.mapping_id }}
    required fields: app_id, mapping_id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_apps_app_id_groups_group_id:
    endpoint: PUT /api/v1/apps/{{ record.app_id }}/groups/{{ record.group_id }}
    required fields: app_id, group_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_apps_app_id_groups_group_id:
    endpoint: DELETE /api/v1/apps/{{ record.app_id }}/groups/{{ record.group_id }}
    required fields: app_id, group_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_apps_app_id_interclient_allowed_apps:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/interclient-allowed-apps
    required fields: app_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_apps_app_id_interclient_allowed_apps_allowed_app_id:
    endpoint: DELETE /api/v1/apps/{{ record.app_id }}/interclient-allowed-apps/{{ record.allowed_app_id }}
    required fields: app_id, allowed_app_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_apps_app_id_lifecycle_activate:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/lifecycle/activate
    required fields: app_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_apps_app_id_lifecycle_deactivate:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/lifecycle/deactivate
    required fields: app_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_apps_app_id_policies_policy_id:
    endpoint: PUT /api/v1/apps/{{ record.app_id }}/policies/{{ record.policy_id }}
    required fields: app_id, policy_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_apps_app_id_tokens:
    endpoint: DELETE /api/v1/apps/{{ record.app_id }}/tokens
    required fields: app_id
    risk: high: external Okta admin API mutation; approval required
  delete_api_v1_apps_app_id_tokens_token_id:
    endpoint: DELETE /api/v1/apps/{{ record.app_id }}/tokens/{{ record.token_id }}
    required fields: app_id, token_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_apps_app_id_users:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/users
    required fields: app_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_apps_app_id_users_user_id:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/users/{{ record.user_id }}
    required fields: app_id, user_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_apps_app_id_users_user_id:
    endpoint: DELETE /api/v1/apps/{{ record.app_id }}/users/{{ record.user_id }}
    required fields: app_id, user_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_apps_app_name_app_id_oauth2_callback:
    endpoint: POST /api/v1/apps/{{ record.app_name }}/{{ record.app_id }}/oauth2/callback
    required fields: app_name, app_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_authenticators:
    endpoint: POST /api/v1/authenticators
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_authenticators_authenticator_id:
    endpoint: PUT /api/v1/authenticators/{{ record.authenticator_id }}
    required fields: authenticator_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_authenticators_authenticator_id_aaguids:
    endpoint: POST /api/v1/authenticators/{{ record.authenticator_id }}/aaguids
    required fields: authenticator_id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_authenticators_authenticator_id_aaguids_aaguid:
    endpoint: PUT /api/v1/authenticators/{{ record.authenticator_id }}/aaguids/{{ record.aaguid }}
    required fields: authenticator_id, aaguid
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_authenticators_authenticator_id_aaguids_aaguid_2:
    endpoint: PATCH /api/v1/authenticators/{{ record.authenticator_id }}/aaguids/{{ record.aaguid }}
    required fields: authenticator_id, aaguid
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_authenticators_authenticator_id_aaguids_aaguid:
    endpoint: DELETE /api/v1/authenticators/{{ record.authenticator_id }}/aaguids/{{ record.aaguid }}
    required fields: authenticator_id, aaguid
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_authenticators_authenticator_id_lifecycle_activate:
    endpoint: POST /api/v1/authenticators/{{ record.authenticator_id }}/lifecycle/activate
    required fields: authenticator_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_authenticators_authenticator_id_lifecycle_deactivate:
    endpoint: POST /api/v1/authenticators/{{ record.authenticator_id }}/lifecycle/deactivate
    required fields: authenticator_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_authenticators_authenticator_id_methods_method_type:
    endpoint: PUT /api/v1/authenticators/{{ record.authenticator_id }}/methods/{{ record.method_type }}
    required fields: authenticator_id, method_type
    risk: medium: external Okta admin API mutation; approval required
  execute_api_v1_authenticators_authenticator_id_methods_method_type_lifecycle_activate:
    endpoint: POST /api/v1/authenticators/{{ record.authenticator_id }}/methods/{{ record.method_type }}/lifecycle/activate
    required fields: authenticator_id, method_type
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_authenticators_authenticator_id_methods_method_type_lifecycle_deactivate:
    endpoint: POST /api/v1/authenticators/{{ record.authenticator_id }}/methods/{{ record.method_type }}/lifecycle/deactivate
    required fields: authenticator_id, method_type
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_authenticators_authenticator_id_methods_web_authn_method_type_verify_rp_id_domain:
    endpoint: POST /api/v1/authenticators/{{ record.authenticator_id }}/methods/{{ record.web_authn_method_type }}/verify-rp-id-domain
    required fields: authenticator_id, web_authn_method_type
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_authorization_servers:
    endpoint: POST /api/v1/authorizationServers
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_authorization_servers_auth_server_id:
    endpoint: PUT /api/v1/authorizationServers/{{ record.auth_server_id }}
    required fields: auth_server_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_authorization_servers_auth_server_id:
    endpoint: DELETE /api/v1/authorizationServers/{{ record.auth_server_id }}
    required fields: auth_server_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_authorization_servers_auth_server_id_associated_servers:
    endpoint: POST /api/v1/authorizationServers/{{ record.auth_server_id }}/associatedServers
    required fields: auth_server_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_authorization_servers_auth_server_id_associated_servers_associated_server_id:
    endpoint: DELETE /api/v1/authorizationServers/{{ record.auth_server_id }}/associatedServers/{{ record.associated_server_id }}
    required fields: auth_server_id, associated_server_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_authorization_servers_auth_server_id_claims:
    endpoint: POST /api/v1/authorizationServers/{{ record.auth_server_id }}/claims
    required fields: auth_server_id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_authorization_servers_auth_server_id_claims_claim_id:
    endpoint: PUT /api/v1/authorizationServers/{{ record.auth_server_id }}/claims/{{ record.claim_id }}
    required fields: auth_server_id, claim_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_authorization_servers_auth_server_id_claims_claim_id:
    endpoint: DELETE /api/v1/authorizationServers/{{ record.auth_server_id }}/claims/{{ record.claim_id }}
    required fields: auth_server_id, claim_id
    risk: high: external Okta admin API mutation; approval required
  delete_api_v1_authorization_servers_auth_server_id_clients_client_id_tokens:
    endpoint: DELETE /api/v1/authorizationServers/{{ record.auth_server_id }}/clients/{{ record.client_id }}/tokens
    required fields: auth_server_id, client_id
    risk: high: external Okta admin API mutation; approval required
  delete_api_v1_authorization_servers_auth_server_id_clients_client_id_tokens_token_id:
    endpoint: DELETE /api/v1/authorizationServers/{{ record.auth_server_id }}/clients/{{ record.client_id }}/tokens/{{ record.token_id }}
    required fields: auth_server_id, client_id, token_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_authorization_servers_auth_server_id_credentials_lifecycle_key_rotate:
    endpoint: POST /api/v1/authorizationServers/{{ record.auth_server_id }}/credentials/lifecycle/keyRotate
    required fields: auth_server_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_authorization_servers_auth_server_id_lifecycle_activate:
    endpoint: POST /api/v1/authorizationServers/{{ record.auth_server_id }}/lifecycle/activate
    required fields: auth_server_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_authorization_servers_auth_server_id_lifecycle_deactivate:
    endpoint: POST /api/v1/authorizationServers/{{ record.auth_server_id }}/lifecycle/deactivate
    required fields: auth_server_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_authorization_servers_auth_server_id_policies:
    endpoint: POST /api/v1/authorizationServers/{{ record.auth_server_id }}/policies
    required fields: auth_server_id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_authorization_servers_auth_server_id_policies_policy_id:
    endpoint: PUT /api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id }}
    required fields: auth_server_id, policy_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_authorization_servers_auth_server_id_policies_policy_id:
    endpoint: DELETE /api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id }}
    required fields: auth_server_id, policy_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_lifecycle_activate:
    endpoint: POST /api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id }}/lifecycle/activate
    required fields: auth_server_id, policy_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_lifecycle_deactivate:
    endpoint: POST /api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id }}/lifecycle/deactivate
    required fields: auth_server_id, policy_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules:
    endpoint: POST /api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id }}/rules
    required fields: auth_server_id, policy_id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id:
    endpoint: PUT /api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id }}/rules/{{ record.rule_id }}
    required fields: auth_server_id, policy_id, rule_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id:
    endpoint: DELETE /api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id }}/rules/{{ record.rule_id }}
    required fields: auth_server_id, policy_id, rule_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id_lifecycle_activate:
    endpoint: POST /api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id }}/rules/{{ record.rule_id }}/lifecycle/activate
    required fields: auth_server_id, policy_id, rule_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id_lifecycle_deactivate:
    endpoint: POST /api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id }}/rules/{{ record.rule_id }}/lifecycle/deactivate
    required fields: auth_server_id, policy_id, rule_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys:
    endpoint: POST /api/v1/authorizationServers/{{ record.auth_server_id }}/resourceservercredentials/keys
    required fields: auth_server_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id:
    endpoint: DELETE /api/v1/authorizationServers/{{ record.auth_server_id }}/resourceservercredentials/keys/{{ record.key_id }}
    required fields: auth_server_id, key_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id_lifecycle_activate:
    endpoint: POST /api/v1/authorizationServers/{{ record.auth_server_id }}/resourceservercredentials/keys/{{ record.key_id }}/lifecycle/activate
    required fields: auth_server_id, key_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id_lifecycle_deactivate:
    endpoint: POST /api/v1/authorizationServers/{{ record.auth_server_id }}/resourceservercredentials/keys/{{ record.key_id }}/lifecycle/deactivate
    required fields: auth_server_id, key_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_authorization_servers_auth_server_id_scopes:
    endpoint: POST /api/v1/authorizationServers/{{ record.auth_server_id }}/scopes
    required fields: auth_server_id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_authorization_servers_auth_server_id_scopes_scope_id:
    endpoint: PUT /api/v1/authorizationServers/{{ record.auth_server_id }}/scopes/{{ record.scope_id }}
    required fields: auth_server_id, scope_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_authorization_servers_auth_server_id_scopes_scope_id:
    endpoint: DELETE /api/v1/authorizationServers/{{ record.auth_server_id }}/scopes/{{ record.scope_id }}
    required fields: auth_server_id, scope_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_behaviors:
    endpoint: POST /api/v1/behaviors
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_behaviors_behavior_id:
    endpoint: PUT /api/v1/behaviors/{{ record.behavior_id }}
    required fields: behavior_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_behaviors_behavior_id:
    endpoint: DELETE /api/v1/behaviors/{{ record.behavior_id }}
    required fields: behavior_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_behaviors_behavior_id_lifecycle_activate:
    endpoint: POST /api/v1/behaviors/{{ record.behavior_id }}/lifecycle/activate
    required fields: behavior_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_behaviors_behavior_id_lifecycle_deactivate:
    endpoint: POST /api/v1/behaviors/{{ record.behavior_id }}/lifecycle/deactivate
    required fields: behavior_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_bot_protection_configuration:
    endpoint: POST /api/v1/bot-protection/configuration
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_brands:
    endpoint: POST /api/v1/brands
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_brands_brand_id:
    endpoint: PUT /api/v1/brands/{{ record.brand_id }}
    required fields: brand_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_brands_brand_id:
    endpoint: DELETE /api/v1/brands/{{ record.brand_id }}
    required fields: brand_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_brands_brand_id_pages_error_customized:
    endpoint: PUT /api/v1/brands/{{ record.brand_id }}/pages/error/customized
    required fields: brand_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_brands_brand_id_pages_error_customized:
    endpoint: DELETE /api/v1/brands/{{ record.brand_id }}/pages/error/customized
    required fields: brand_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_brands_brand_id_pages_error_preview:
    endpoint: PUT /api/v1/brands/{{ record.brand_id }}/pages/error/preview
    required fields: brand_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_brands_brand_id_pages_error_preview:
    endpoint: DELETE /api/v1/brands/{{ record.brand_id }}/pages/error/preview
    required fields: brand_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_brands_brand_id_pages_sign_in_customized:
    endpoint: PUT /api/v1/brands/{{ record.brand_id }}/pages/sign-in/customized
    required fields: brand_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_brands_brand_id_pages_sign_in_customized:
    endpoint: DELETE /api/v1/brands/{{ record.brand_id }}/pages/sign-in/customized
    required fields: brand_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_brands_brand_id_pages_sign_in_preview:
    endpoint: PUT /api/v1/brands/{{ record.brand_id }}/pages/sign-in/preview
    required fields: brand_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_brands_brand_id_pages_sign_in_preview:
    endpoint: DELETE /api/v1/brands/{{ record.brand_id }}/pages/sign-in/preview
    required fields: brand_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_brands_brand_id_pages_sign_out_customized:
    endpoint: PUT /api/v1/brands/{{ record.brand_id }}/pages/sign-out/customized
    required fields: brand_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_brands_brand_id_templates_email_template_name_customizations:
    endpoint: POST /api/v1/brands/{{ record.brand_id }}/templates/email/{{ record.template_name }}/customizations
    required fields: brand_id, template_name
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_brands_brand_id_templates_email_template_name_customizations:
    endpoint: DELETE /api/v1/brands/{{ record.brand_id }}/templates/email/{{ record.template_name }}/customizations
    required fields: brand_id, template_name
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id:
    endpoint: PUT /api/v1/brands/{{ record.brand_id }}/templates/email/{{ record.template_name }}/customizations/{{ record.customization_id }}
    required fields: brand_id, template_name, customization_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id:
    endpoint: DELETE /api/v1/brands/{{ record.brand_id }}/templates/email/{{ record.template_name }}/customizations/{{ record.customization_id }}
    required fields: brand_id, template_name, customization_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_brands_brand_id_templates_email_template_name_settings:
    endpoint: PUT /api/v1/brands/{{ record.brand_id }}/templates/email/{{ record.template_name }}/settings
    required fields: brand_id, template_name
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_brands_brand_id_templates_email_template_name_test:
    endpoint: POST /api/v1/brands/{{ record.brand_id }}/templates/email/{{ record.template_name }}/test
    required fields: brand_id, template_name
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_brands_brand_id_themes_theme_id:
    endpoint: PUT /api/v1/brands/{{ record.brand_id }}/themes/{{ record.theme_id }}
    required fields: brand_id, theme_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_brands_brand_id_themes_theme_id_background_image:
    endpoint: DELETE /api/v1/brands/{{ record.brand_id }}/themes/{{ record.theme_id }}/background-image
    required fields: brand_id, theme_id
    risk: high: external Okta admin API mutation; approval required
  delete_api_v1_brands_brand_id_themes_theme_id_favicon:
    endpoint: DELETE /api/v1/brands/{{ record.brand_id }}/themes/{{ record.theme_id }}/favicon
    required fields: brand_id, theme_id
    risk: high: external Okta admin API mutation; approval required
  delete_api_v1_brands_brand_id_themes_theme_id_logo:
    endpoint: DELETE /api/v1/brands/{{ record.brand_id }}/themes/{{ record.theme_id }}/logo
    required fields: brand_id, theme_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_brands_brand_id_well_known_uris_path_customized:
    endpoint: PUT /api/v1/brands/{{ record.brand_id }}/well-known-uris/{{ record.path }}/customized
    required fields: brand_id, path
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_captchas:
    endpoint: POST /api/v1/captchas
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_captchas_captcha_id:
    endpoint: POST /api/v1/captchas/{{ record.captcha_id }}
    required fields: captcha_id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_captchas_captcha_id:
    endpoint: PUT /api/v1/captchas/{{ record.captcha_id }}
    required fields: captcha_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_captchas_captcha_id:
    endpoint: DELETE /api/v1/captchas/{{ record.captcha_id }}
    required fields: captcha_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_device_assurances:
    endpoint: POST /api/v1/device-assurances
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_device_assurances_device_assurance_id:
    endpoint: PUT /api/v1/device-assurances/{{ record.device_assurance_id }}
    required fields: device_assurance_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_device_assurances_device_assurance_id:
    endpoint: DELETE /api/v1/device-assurances/{{ record.device_assurance_id }}
    required fields: device_assurance_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_device_integrations_device_integration_id_lifecycle_activate:
    endpoint: POST /api/v1/device-integrations/{{ record.device_integration_id }}/lifecycle/activate
    required fields: device_integration_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_device_integrations_device_integration_id_lifecycle_deactivate:
    endpoint: POST /api/v1/device-integrations/{{ record.device_integration_id }}/lifecycle/deactivate
    required fields: device_integration_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_device_posture_checks:
    endpoint: POST /api/v1/device-posture-checks
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_device_posture_checks_posture_check_id:
    endpoint: PUT /api/v1/device-posture-checks/{{ record.posture_check_id }}
    required fields: posture_check_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_device_posture_checks_posture_check_id:
    endpoint: DELETE /api/v1/device-posture-checks/{{ record.posture_check_id }}
    required fields: posture_check_id
    risk: high: external Okta admin API mutation; approval required
  delete_api_v1_devices_device_id:
    endpoint: DELETE /api/v1/devices/{{ record.device_id }}
    required fields: device_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_devices_device_id_lifecycle_activate:
    endpoint: POST /api/v1/devices/{{ record.device_id }}/lifecycle/activate
    required fields: device_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_devices_device_id_lifecycle_deactivate:
    endpoint: POST /api/v1/devices/{{ record.device_id }}/lifecycle/deactivate
    required fields: device_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_devices_device_id_lifecycle_suspend:
    endpoint: POST /api/v1/devices/{{ record.device_id }}/lifecycle/suspend
    required fields: device_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_devices_device_id_lifecycle_unsuspend:
    endpoint: POST /api/v1/devices/{{ record.device_id }}/lifecycle/unsuspend
    required fields: device_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_directories_app_instance_id_groups_modify:
    endpoint: POST /api/v1/directories/{{ record.app_instance_id }}/groups/modify
    required fields: app_instance_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_directories_app_instance_id_groups_group_id_query:
    endpoint: POST /api/v1/directories/{{ record.app_instance_id }}/groups/{{ record.group_id }}/query
    required fields: app_instance_id, group_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_domains:
    endpoint: POST /api/v1/domains
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_domains_domain_id:
    endpoint: PUT /api/v1/domains/{{ record.domain_id }}
    required fields: domain_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_domains_domain_id:
    endpoint: DELETE /api/v1/domains/{{ record.domain_id }}
    required fields: domain_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_domains_domain_id_certificate:
    endpoint: PUT /api/v1/domains/{{ record.domain_id }}/certificate
    required fields: domain_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_domains_domain_id_verify:
    endpoint: POST /api/v1/domains/{{ record.domain_id }}/verify
    required fields: domain_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_dr_failback:
    endpoint: POST /api/v1/dr/failback
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_dr_failover:
    endpoint: POST /api/v1/dr/failover
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_email_domains:
    endpoint: POST /api/v1/email-domains
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_email_domains_email_domain_id:
    endpoint: PUT /api/v1/email-domains/{{ record.email_domain_id }}
    required fields: email_domain_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_email_domains_email_domain_id:
    endpoint: DELETE /api/v1/email-domains/{{ record.email_domain_id }}
    required fields: email_domain_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_email_domains_email_domain_id_verify:
    endpoint: POST /api/v1/email-domains/{{ record.email_domain_id }}/verify
    required fields: email_domain_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_email_servers:
    endpoint: POST /api/v1/email-servers
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_email_servers_email_server_id:
    endpoint: PATCH /api/v1/email-servers/{{ record.email_server_id }}
    required fields: email_server_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_email_servers_email_server_id:
    endpoint: DELETE /api/v1/email-servers/{{ record.email_server_id }}
    required fields: email_server_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_email_servers_email_server_id_test:
    endpoint: POST /api/v1/email-servers/{{ record.email_server_id }}/test
    required fields: email_server_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_event_hooks:
    endpoint: POST /api/v1/eventHooks
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_event_hooks_event_hook_id:
    endpoint: PUT /api/v1/eventHooks/{{ record.event_hook_id }}
    required fields: event_hook_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_event_hooks_event_hook_id:
    endpoint: DELETE /api/v1/eventHooks/{{ record.event_hook_id }}
    required fields: event_hook_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_event_hooks_event_hook_id_lifecycle_activate:
    endpoint: POST /api/v1/eventHooks/{{ record.event_hook_id }}/lifecycle/activate
    required fields: event_hook_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_event_hooks_event_hook_id_lifecycle_deactivate:
    endpoint: POST /api/v1/eventHooks/{{ record.event_hook_id }}/lifecycle/deactivate
    required fields: event_hook_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_event_hooks_event_hook_id_lifecycle_verify:
    endpoint: POST /api/v1/eventHooks/{{ record.event_hook_id }}/lifecycle/verify
    required fields: event_hook_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_features_feature_id_lifecycle:
    endpoint: POST /api/v1/features/{{ record.feature_id }}/{{ record.lifecycle }}
    required fields: feature_id, lifecycle
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_first_party_app_settings_app_name:
    endpoint: PUT /api/v1/first-party-app-settings/{{ record.app_name }}
    required fields: app_name
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_groups:
    endpoint: POST /api/v1/groups
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_groups_rules:
    endpoint: POST /api/v1/groups/rules
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_groups_rules_group_rule_id:
    endpoint: PUT /api/v1/groups/rules/{{ record.group_rule_id }}
    required fields: group_rule_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_groups_rules_group_rule_id:
    endpoint: DELETE /api/v1/groups/rules/{{ record.group_rule_id }}
    required fields: group_rule_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_groups_rules_group_rule_id_lifecycle_activate:
    endpoint: POST /api/v1/groups/rules/{{ record.group_rule_id }}/lifecycle/activate
    required fields: group_rule_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_groups_rules_group_rule_id_lifecycle_deactivate:
    endpoint: POST /api/v1/groups/rules/{{ record.group_rule_id }}/lifecycle/deactivate
    required fields: group_rule_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_groups_group_id:
    endpoint: PUT /api/v1/groups/{{ record.group_id }}
    required fields: group_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_groups_group_id:
    endpoint: DELETE /api/v1/groups/{{ record.group_id }}
    required fields: group_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_groups_group_id_owners:
    endpoint: POST /api/v1/groups/{{ record.group_id }}/owners
    required fields: group_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_groups_group_id_owners_owner_id:
    endpoint: DELETE /api/v1/groups/{{ record.group_id }}/owners/{{ record.owner_id }}
    required fields: group_id, owner_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_groups_group_id_roles:
    endpoint: POST /api/v1/groups/{{ record.group_id }}/roles
    required fields: group_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_groups_group_id_roles_role_assignment_id:
    endpoint: DELETE /api/v1/groups/{{ record.group_id }}/roles/{{ record.role_assignment_id }}
    required fields: group_id, role_assignment_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name:
    endpoint: PUT /api/v1/groups/{{ record.group_id }}/roles/{{ record.role_assignment_id }}/targets/catalog/apps/{{ record.app_name }}
    required fields: group_id, role_assignment_id, app_name
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name:
    endpoint: DELETE /api/v1/groups/{{ record.group_id }}/roles/{{ record.role_assignment_id }}/targets/catalog/apps/{{ record.app_name }}
    required fields: group_id, role_assignment_id, app_name
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id:
    endpoint: PUT /api/v1/groups/{{ record.group_id }}/roles/{{ record.role_assignment_id }}/targets/catalog/apps/{{ record.app_name }}/{{ record.app_id }}
    required fields: group_id, role_assignment_id, app_name, app_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id:
    endpoint: DELETE /api/v1/groups/{{ record.group_id }}/roles/{{ record.role_assignment_id }}/targets/catalog/apps/{{ record.app_name }}/{{ record.app_id }}
    required fields: group_id, role_assignment_id, app_name, app_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_groups_group_id_roles_role_assignment_id_targets_groups_target_group_id:
    endpoint: PUT /api/v1/groups/{{ record.group_id }}/roles/{{ record.role_assignment_id }}/targets/groups/{{ record.target_group_id }}
    required fields: group_id, role_assignment_id, target_group_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_groups_group_id_roles_role_assignment_id_targets_groups_target_group_id:
    endpoint: DELETE /api/v1/groups/{{ record.group_id }}/roles/{{ record.role_assignment_id }}/targets/groups/{{ record.target_group_id }}
    required fields: group_id, role_assignment_id, target_group_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_groups_group_id_users_user_id:
    endpoint: PUT /api/v1/groups/{{ record.group_id }}/users/{{ record.user_id }}
    required fields: group_id, user_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_groups_group_id_users_user_id:
    endpoint: DELETE /api/v1/groups/{{ record.group_id }}/users/{{ record.user_id }}
    required fields: group_id, user_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_hook_keys:
    endpoint: POST /api/v1/hook-keys
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_hook_keys_id:
    endpoint: PUT /api/v1/hook-keys/{{ record.id }}
    required fields: id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_hook_keys_id:
    endpoint: DELETE /api/v1/hook-keys/{{ record.id }}
    required fields: id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_iam_governance_bundles:
    endpoint: POST /api/v1/iam/governance/bundles
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_iam_governance_bundles_bundle_id:
    endpoint: PUT /api/v1/iam/governance/bundles/{{ record.bundle_id }}
    required fields: bundle_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_iam_governance_bundles_bundle_id:
    endpoint: DELETE /api/v1/iam/governance/bundles/{{ record.bundle_id }}
    required fields: bundle_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_iam_governance_opt_in:
    endpoint: POST /api/v1/iam/governance/optIn
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_iam_governance_opt_out:
    endpoint: POST /api/v1/iam/governance/optOut
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_iam_resource_sets:
    endpoint: POST /api/v1/iam/resource-sets
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_iam_resource_sets_resource_set_id_or_label:
    endpoint: PUT /api/v1/iam/resource-sets/{{ record.resource_set_id_or_label }}
    required fields: resource_set_id_or_label
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_iam_resource_sets_resource_set_id_or_label:
    endpoint: DELETE /api/v1/iam/resource-sets/{{ record.resource_set_id_or_label }}
    required fields: resource_set_id_or_label
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_iam_resource_sets_resource_set_id_or_label_bindings:
    endpoint: POST /api/v1/iam/resource-sets/{{ record.resource_set_id_or_label }}/bindings
    required fields: resource_set_id_or_label
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label:
    endpoint: DELETE /api/v1/iam/resource-sets/{{ record.resource_set_id_or_label }}/bindings/{{ record.role_id_or_label }}
    required fields: resource_set_id_or_label, role_id_or_label
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members:
    endpoint: PATCH /api/v1/iam/resource-sets/{{ record.resource_set_id_or_label }}/bindings/{{ record.role_id_or_label }}/members
    required fields: resource_set_id_or_label, role_id_or_label
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members_member_id:
    endpoint: DELETE /api/v1/iam/resource-sets/{{ record.resource_set_id_or_label }}/bindings/{{ record.role_id_or_label }}/members/{{ record.member_id }}
    required fields: resource_set_id_or_label, role_id_or_label, member_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_iam_resource_sets_resource_set_id_or_label_resources:
    endpoint: POST /api/v1/iam/resource-sets/{{ record.resource_set_id_or_label }}/resources
    required fields: resource_set_id_or_label
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_iam_resource_sets_resource_set_id_or_label_resources:
    endpoint: PATCH /api/v1/iam/resource-sets/{{ record.resource_set_id_or_label }}/resources
    required fields: resource_set_id_or_label
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_iam_resource_sets_resource_set_id_or_label_resources_resource_id:
    endpoint: PUT /api/v1/iam/resource-sets/{{ record.resource_set_id_or_label }}/resources/{{ record.resource_id }}
    required fields: resource_set_id_or_label, resource_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_iam_resource_sets_resource_set_id_or_label_resources_resource_id:
    endpoint: DELETE /api/v1/iam/resource-sets/{{ record.resource_set_id_or_label }}/resources/{{ record.resource_id }}
    required fields: resource_set_id_or_label, resource_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_iam_roles:
    endpoint: POST /api/v1/iam/roles
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_iam_roles_role_id_or_label:
    endpoint: PUT /api/v1/iam/roles/{{ record.role_id_or_label }}
    required fields: role_id_or_label
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_iam_roles_role_id_or_label:
    endpoint: DELETE /api/v1/iam/roles/{{ record.role_id_or_label }}
    required fields: role_id_or_label
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_iam_roles_role_id_or_label_permissions_permission_type:
    endpoint: POST /api/v1/iam/roles/{{ record.role_id_or_label }}/permissions/{{ record.permission_type }}
    required fields: role_id_or_label, permission_type
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_iam_roles_role_id_or_label_permissions_permission_type:
    endpoint: PUT /api/v1/iam/roles/{{ record.role_id_or_label }}/permissions/{{ record.permission_type }}
    required fields: role_id_or_label, permission_type
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_iam_roles_role_id_or_label_permissions_permission_type:
    endpoint: DELETE /api/v1/iam/roles/{{ record.role_id_or_label }}/permissions/{{ record.permission_type }}
    required fields: role_id_or_label, permission_type
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_identity_sources_identity_source_id_groups:
    endpoint: POST /api/v1/identity-sources/{{ record.identity_source_id }}/groups
    required fields: identity_source_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_identity_sources_identity_source_id_groups_group_or_external_id:
    endpoint: POST /api/v1/identity-sources/{{ record.identity_source_id }}/groups/{{ record.group_or_external_id }}
    required fields: identity_source_id, group_or_external_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_identity_sources_identity_source_id_groups_group_or_external_id:
    endpoint: DELETE /api/v1/identity-sources/{{ record.identity_source_id }}/groups/{{ record.group_or_external_id }}
    required fields: identity_source_id, group_or_external_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_identity_sources_identity_source_id_groups_group_or_external_id_membership:
    endpoint: POST /api/v1/identity-sources/{{ record.identity_source_id }}/groups/{{ record.group_or_external_id }}/membership
    required fields: identity_source_id, group_or_external_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_identity_sources_identity_source_id_groups_group_or_external_id_membership_member_external_id:
    endpoint: DELETE /api/v1/identity-sources/{{ record.identity_source_id }}/groups/{{ record.group_or_external_id }}/membership/{{ record.member_external_id }}
    required fields: identity_source_id, group_or_external_id, member_external_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_identity_sources_identity_source_id_sessions:
    endpoint: POST /api/v1/identity-sources/{{ record.identity_source_id }}/sessions
    required fields: identity_source_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_identity_sources_identity_source_id_sessions_session_id:
    endpoint: DELETE /api/v1/identity-sources/{{ record.identity_source_id }}/sessions/{{ record.session_id }}
    required fields: identity_source_id, session_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_delete:
    endpoint: POST /api/v1/identity-sources/{{ record.identity_source_id }}/sessions/{{ record.session_id }}/bulk-delete
    required fields: identity_source_id, session_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_group_memberships_delete:
    endpoint: POST /api/v1/identity-sources/{{ record.identity_source_id }}/sessions/{{ record.session_id }}/bulk-group-memberships-delete
    required fields: identity_source_id, session_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_group_memberships_upsert:
    endpoint: POST /api/v1/identity-sources/{{ record.identity_source_id }}/sessions/{{ record.session_id }}/bulk-group-memberships-upsert
    required fields: identity_source_id, session_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_groups_delete:
    endpoint: POST /api/v1/identity-sources/{{ record.identity_source_id }}/sessions/{{ record.session_id }}/bulk-groups-delete
    required fields: identity_source_id, session_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_groups_upsert:
    endpoint: POST /api/v1/identity-sources/{{ record.identity_source_id }}/sessions/{{ record.session_id }}/bulk-groups-upsert
    required fields: identity_source_id, session_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_upsert:
    endpoint: POST /api/v1/identity-sources/{{ record.identity_source_id }}/sessions/{{ record.session_id }}/bulk-upsert
    required fields: identity_source_id, session_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_identity_sources_identity_source_id_sessions_session_id_start_import:
    endpoint: POST /api/v1/identity-sources/{{ record.identity_source_id }}/sessions/{{ record.session_id }}/start-import
    required fields: identity_source_id, session_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_identity_sources_identity_source_id_users:
    endpoint: POST /api/v1/identity-sources/{{ record.identity_source_id }}/users
    required fields: identity_source_id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_identity_sources_identity_source_id_users_external_id:
    endpoint: PUT /api/v1/identity-sources/{{ record.identity_source_id }}/users/{{ record.external_id }}
    required fields: identity_source_id, external_id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_identity_sources_identity_source_id_users_external_id_2:
    endpoint: PATCH /api/v1/identity-sources/{{ record.identity_source_id }}/users/{{ record.external_id }}
    required fields: identity_source_id, external_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_identity_sources_identity_source_id_users_external_id:
    endpoint: DELETE /api/v1/identity-sources/{{ record.identity_source_id }}/users/{{ record.external_id }}
    required fields: identity_source_id, external_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_idps:
    endpoint: POST /api/v1/idps
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_idps_credentials_keys:
    endpoint: POST /api/v1/idps/credentials/keys
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_idps_credentials_keys_kid:
    endpoint: PUT /api/v1/idps/credentials/keys/{{ record.kid }}
    required fields: kid
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_idps_credentials_keys_kid:
    endpoint: DELETE /api/v1/idps/credentials/keys/{{ record.kid }}
    required fields: kid
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_idps_idp_id:
    endpoint: PUT /api/v1/idps/{{ record.idp_id }}
    required fields: idp_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_idps_idp_id:
    endpoint: DELETE /api/v1/idps/{{ record.idp_id }}
    required fields: idp_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_idps_idp_id_credentials_csrs:
    endpoint: POST /api/v1/idps/{{ record.idp_id }}/credentials/csrs
    required fields: idp_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_idps_idp_id_credentials_csrs_idp_csr_id:
    endpoint: DELETE /api/v1/idps/{{ record.idp_id }}/credentials/csrs/{{ record.idp_csr_id }}
    required fields: idp_id, idp_csr_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_idps_idp_id_lifecycle_activate:
    endpoint: POST /api/v1/idps/{{ record.idp_id }}/lifecycle/activate
    required fields: idp_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_idps_idp_id_lifecycle_deactivate:
    endpoint: POST /api/v1/idps/{{ record.idp_id }}/lifecycle/deactivate
    required fields: idp_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_idps_idp_id_users_user_id:
    endpoint: POST /api/v1/idps/{{ record.idp_id }}/users/{{ record.user_id }}
    required fields: idp_id, user_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_idps_idp_id_users_user_id:
    endpoint: DELETE /api/v1/idps/{{ record.idp_id }}/users/{{ record.user_id }}
    required fields: idp_id, user_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_inline_hooks:
    endpoint: POST /api/v1/inlineHooks
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_inline_hooks_inline_hook_id:
    endpoint: POST /api/v1/inlineHooks/{{ record.inline_hook_id }}
    required fields: inline_hook_id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_inline_hooks_inline_hook_id:
    endpoint: PUT /api/v1/inlineHooks/{{ record.inline_hook_id }}
    required fields: inline_hook_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_inline_hooks_inline_hook_id:
    endpoint: DELETE /api/v1/inlineHooks/{{ record.inline_hook_id }}
    required fields: inline_hook_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_inline_hooks_inline_hook_id_execute:
    endpoint: POST /api/v1/inlineHooks/{{ record.inline_hook_id }}/execute
    required fields: inline_hook_id
    risk: medium: external Okta admin API mutation; approval required
  execute_api_v1_inline_hooks_inline_hook_id_lifecycle_activate:
    endpoint: POST /api/v1/inlineHooks/{{ record.inline_hook_id }}/lifecycle/activate
    required fields: inline_hook_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_inline_hooks_inline_hook_id_lifecycle_deactivate:
    endpoint: POST /api/v1/inlineHooks/{{ record.inline_hook_id }}/lifecycle/deactivate
    required fields: inline_hook_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_log_streams:
    endpoint: POST /api/v1/logStreams
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_log_streams_log_stream_id:
    endpoint: PUT /api/v1/logStreams/{{ record.log_stream_id }}
    required fields: log_stream_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_log_streams_log_stream_id:
    endpoint: DELETE /api/v1/logStreams/{{ record.log_stream_id }}
    required fields: log_stream_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_log_streams_log_stream_id_lifecycle_activate:
    endpoint: POST /api/v1/logStreams/{{ record.log_stream_id }}/lifecycle/activate
    required fields: log_stream_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_log_streams_log_stream_id_lifecycle_deactivate:
    endpoint: POST /api/v1/logStreams/{{ record.log_stream_id }}/lifecycle/deactivate
    required fields: log_stream_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_mappings_mapping_id:
    endpoint: POST /api/v1/mappings/{{ record.mapping_id }}
    required fields: mapping_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_meta_schemas_apps_app_id_default:
    endpoint: POST /api/v1/meta/schemas/apps/{{ record.app_id }}/default
    required fields: app_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_meta_schemas_group_default:
    endpoint: POST /api/v1/meta/schemas/group/default
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_meta_schemas_user_linked_objects:
    endpoint: POST /api/v1/meta/schemas/user/linkedObjects
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_meta_schemas_user_linked_objects_linked_object_name:
    endpoint: DELETE /api/v1/meta/schemas/user/linkedObjects/{{ record.linked_object_name }}
    required fields: linked_object_name
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_meta_schemas_user_schema_id:
    endpoint: POST /api/v1/meta/schemas/user/{{ record.schema_id }}
    required fields: schema_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_meta_types_user:
    endpoint: POST /api/v1/meta/types/user
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_meta_types_user_type_id:
    endpoint: POST /api/v1/meta/types/user/{{ record.type_id }}
    required fields: type_id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_meta_types_user_type_id:
    endpoint: PUT /api/v1/meta/types/user/{{ record.type_id }}
    required fields: type_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_meta_types_user_type_id:
    endpoint: DELETE /api/v1/meta/types/user/{{ record.type_id }}
    required fields: type_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_meta_uischemas:
    endpoint: POST /api/v1/meta/uischemas
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_meta_uischemas_id:
    endpoint: PUT /api/v1/meta/uischemas/{{ record.id }}
    required fields: id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_meta_uischemas_id:
    endpoint: DELETE /api/v1/meta/uischemas/{{ record.id }}
    required fields: id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_org:
    endpoint: POST /api/v1/org
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_org:
    endpoint: PUT /api/v1/org
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_org_captcha:
    endpoint: PUT /api/v1/org/captcha
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_org_captcha:
    endpoint: DELETE /api/v1/org/captcha
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_org_contacts_contact_type:
    endpoint: PUT /api/v1/org/contacts/{{ record.contact_type }}
    required fields: contact_type
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_org_email_bounces_remove_list:
    endpoint: POST /api/v1/org/email/bounces/remove-list
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_org_factors_yubikey_token_tokens:
    endpoint: POST /api/v1/org/factors/yubikey_token/tokens
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_org_org_settings_third_party_admin_setting:
    endpoint: POST /api/v1/org/orgSettings/thirdPartyAdminSetting
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_org_preferences_hide_end_user_footer:
    endpoint: POST /api/v1/org/preferences/hideEndUserFooter
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_org_preferences_show_end_user_footer:
    endpoint: POST /api/v1/org/preferences/showEndUserFooter
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_org_privacy_aerial_grant:
    endpoint: POST /api/v1/org/privacy/aerial/grant
    risk: medium: external Okta admin API mutation; approval required
  execute_api_v1_org_privacy_aerial_revoke:
    endpoint: POST /api/v1/org/privacy/aerial/revoke
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_org_privacy_okta_communication_opt_in:
    endpoint: POST /api/v1/org/privacy/oktaCommunication/optIn
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_org_privacy_okta_communication_opt_out:
    endpoint: POST /api/v1/org/privacy/oktaCommunication/optOut
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_org_privacy_okta_support_cases_case_number:
    endpoint: PATCH /api/v1/org/privacy/oktaSupport/cases/{{ record.case_number }}
    required fields: case_number
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_org_privacy_okta_support_extend:
    endpoint: POST /api/v1/org/privacy/oktaSupport/extend
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_org_privacy_okta_support_grant:
    endpoint: POST /api/v1/org/privacy/oktaSupport/grant
    risk: medium: external Okta admin API mutation; approval required
  execute_api_v1_org_privacy_okta_support_revoke:
    endpoint: POST /api/v1/org/privacy/oktaSupport/revoke
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_org_settings_auto_assign_admin_app_setting:
    endpoint: POST /api/v1/org/settings/autoAssignAdminAppSetting
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_org_settings_client_privileges_setting:
    endpoint: PUT /api/v1/org/settings/clientPrivilegesSetting
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_orgs:
    endpoint: POST /api/v1/orgs
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_policies:
    endpoint: POST /api/v1/policies
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_policies_policy_id:
    endpoint: PUT /api/v1/policies/{{ record.policy_id }}
    required fields: policy_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_policies_policy_id:
    endpoint: DELETE /api/v1/policies/{{ record.policy_id }}
    required fields: policy_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_policies_policy_id_clone:
    endpoint: POST /api/v1/policies/{{ record.policy_id }}/clone
    required fields: policy_id
    risk: medium: external Okta admin API mutation; approval required
  execute_api_v1_policies_policy_id_lifecycle_activate:
    endpoint: POST /api/v1/policies/{{ record.policy_id }}/lifecycle/activate
    required fields: policy_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_policies_policy_id_lifecycle_deactivate:
    endpoint: POST /api/v1/policies/{{ record.policy_id }}/lifecycle/deactivate
    required fields: policy_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_policies_policy_id_mappings:
    endpoint: POST /api/v1/policies/{{ record.policy_id }}/mappings
    required fields: policy_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_policies_policy_id_mappings_mapping_id:
    endpoint: DELETE /api/v1/policies/{{ record.policy_id }}/mappings/{{ record.mapping_id }}
    required fields: policy_id, mapping_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_policies_policy_id_rules:
    endpoint: POST /api/v1/policies/{{ record.policy_id }}/rules
    required fields: policy_id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_policies_policy_id_rules_rule_id:
    endpoint: PUT /api/v1/policies/{{ record.policy_id }}/rules/{{ record.rule_id }}
    required fields: policy_id, rule_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_policies_policy_id_rules_rule_id:
    endpoint: DELETE /api/v1/policies/{{ record.policy_id }}/rules/{{ record.rule_id }}
    required fields: policy_id, rule_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_policies_policy_id_rules_rule_id_lifecycle_activate:
    endpoint: POST /api/v1/policies/{{ record.policy_id }}/rules/{{ record.rule_id }}/lifecycle/activate
    required fields: policy_id, rule_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_policies_policy_id_rules_rule_id_lifecycle_deactivate:
    endpoint: POST /api/v1/policies/{{ record.policy_id }}/rules/{{ record.rule_id }}/lifecycle/deactivate
    required fields: policy_id, rule_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_principal_rate_limits:
    endpoint: POST /api/v1/principal-rate-limits
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_principal_rate_limits_principal_rate_limit_id:
    endpoint: PUT /api/v1/principal-rate-limits/{{ record.principal_rate_limit_id }}
    required fields: principal_rate_limit_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_push_providers:
    endpoint: POST /api/v1/push-providers
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_push_providers_push_provider_id:
    endpoint: PUT /api/v1/push-providers/{{ record.push_provider_id }}
    required fields: push_provider_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_push_providers_push_provider_id:
    endpoint: DELETE /api/v1/push-providers/{{ record.push_provider_id }}
    required fields: push_provider_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_rate_limit_settings_admin_notifications:
    endpoint: PUT /api/v1/rate-limit-settings/admin-notifications
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_rate_limit_settings_per_client:
    endpoint: PUT /api/v1/rate-limit-settings/per-client
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_rate_limit_settings_warning_threshold:
    endpoint: PUT /api/v1/rate-limit-settings/warning-threshold
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_realm_assignments:
    endpoint: POST /api/v1/realm-assignments
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_realm_assignments_operations:
    endpoint: POST /api/v1/realm-assignments/operations
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_realm_assignments_assignment_id:
    endpoint: PUT /api/v1/realm-assignments/{{ record.assignment_id }}
    required fields: assignment_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_realm_assignments_assignment_id:
    endpoint: DELETE /api/v1/realm-assignments/{{ record.assignment_id }}
    required fields: assignment_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_realm_assignments_assignment_id_lifecycle_activate:
    endpoint: POST /api/v1/realm-assignments/{{ record.assignment_id }}/lifecycle/activate
    required fields: assignment_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_realm_assignments_assignment_id_lifecycle_deactivate:
    endpoint: POST /api/v1/realm-assignments/{{ record.assignment_id }}/lifecycle/deactivate
    required fields: assignment_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_realms:
    endpoint: POST /api/v1/realms
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_realms_realm_id:
    endpoint: PUT /api/v1/realms/{{ record.realm_id }}
    required fields: realm_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_realms_realm_id:
    endpoint: DELETE /api/v1/realms/{{ record.realm_id }}
    required fields: realm_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_roles_role_ref_subscriptions_notification_type_subscribe:
    endpoint: POST /api/v1/roles/{{ record.role_ref }}/subscriptions/{{ record.notification_type }}/subscribe
    required fields: role_ref, notification_type
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_roles_role_ref_subscriptions_notification_type_unsubscribe:
    endpoint: POST /api/v1/roles/{{ record.role_ref }}/subscriptions/{{ record.notification_type }}/unsubscribe
    required fields: role_ref, notification_type
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_security_events_providers:
    endpoint: POST /api/v1/security-events-providers
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_security_events_providers_security_event_provider_id:
    endpoint: PUT /api/v1/security-events-providers/{{ record.security_event_provider_id }}
    required fields: security_event_provider_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_security_events_providers_security_event_provider_id:
    endpoint: DELETE /api/v1/security-events-providers/{{ record.security_event_provider_id }}
    required fields: security_event_provider_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_security_events_providers_security_event_provider_id_lifecycle_activate:
    endpoint: POST /api/v1/security-events-providers/{{ record.security_event_provider_id }}/lifecycle/activate
    required fields: security_event_provider_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_security_events_providers_security_event_provider_id_lifecycle_deactivate:
    endpoint: POST /api/v1/security-events-providers/{{ record.security_event_provider_id }}/lifecycle/deactivate
    required fields: security_event_provider_id
    risk: high: external Okta admin API mutation; approval required
  delete_api_v1_sessions_session_id:
    endpoint: DELETE /api/v1/sessions/{{ record.session_id }}
    required fields: session_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_sessions_session_id_lifecycle_refresh:
    endpoint: POST /api/v1/sessions/{{ record.session_id }}/lifecycle/refresh
    required fields: session_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_ssf_stream:
    endpoint: POST /api/v1/ssf/stream
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_ssf_stream:
    endpoint: PUT /api/v1/ssf/stream
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_ssf_stream_2:
    endpoint: PATCH /api/v1/ssf/stream
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_ssf_stream:
    endpoint: DELETE /api/v1/ssf/stream
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_ssf_stream_verification:
    endpoint: POST /api/v1/ssf/stream/verification
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_telephony_providers:
    endpoint: POST /api/v1/telephony-providers
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_telephony_providers_custom_telephony_provider_id:
    endpoint: PATCH /api/v1/telephony-providers/{{ record.custom_telephony_provider_id }}
    required fields: custom_telephony_provider_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_telephony_providers_custom_telephony_provider_id:
    endpoint: DELETE /api/v1/telephony-providers/{{ record.custom_telephony_provider_id }}
    required fields: custom_telephony_provider_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_telephony_providers_custom_telephony_provider_id_lifecycle_activate:
    endpoint: POST /api/v1/telephony-providers/{{ record.custom_telephony_provider_id }}/lifecycle/activate
    required fields: custom_telephony_provider_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_telephony_providers_custom_telephony_provider_id_lifecycle_deactivate:
    endpoint: POST /api/v1/telephony-providers/{{ record.custom_telephony_provider_id }}/lifecycle/deactivate
    required fields: custom_telephony_provider_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_telephony_providers_custom_telephony_provider_id_set_as_primary:
    endpoint: POST /api/v1/telephony-providers/{{ record.custom_telephony_provider_id }}/setAsPrimary
    required fields: custom_telephony_provider_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_telephony_providers_custom_telephony_provider_id_test:
    endpoint: POST /api/v1/telephony-providers/{{ record.custom_telephony_provider_id }}/test
    required fields: custom_telephony_provider_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_templates_sms:
    endpoint: POST /api/v1/templates/sms
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_templates_sms_template_id:
    endpoint: POST /api/v1/templates/sms/{{ record.template_id }}
    required fields: template_id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_templates_sms_template_id:
    endpoint: PUT /api/v1/templates/sms/{{ record.template_id }}
    required fields: template_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_templates_sms_template_id:
    endpoint: DELETE /api/v1/templates/sms/{{ record.template_id }}
    required fields: template_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_threats_configuration:
    endpoint: POST /api/v1/threats/configuration
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_trusted_origins:
    endpoint: POST /api/v1/trustedOrigins
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_trusted_origins_trusted_origin_id:
    endpoint: PUT /api/v1/trustedOrigins/{{ record.trusted_origin_id }}
    required fields: trusted_origin_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_trusted_origins_trusted_origin_id:
    endpoint: DELETE /api/v1/trustedOrigins/{{ record.trusted_origin_id }}
    required fields: trusted_origin_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_trusted_origins_trusted_origin_id_lifecycle_activate:
    endpoint: POST /api/v1/trustedOrigins/{{ record.trusted_origin_id }}/lifecycle/activate
    required fields: trusted_origin_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_trusted_origins_trusted_origin_id_lifecycle_deactivate:
    endpoint: POST /api/v1/trustedOrigins/{{ record.trusted_origin_id }}/lifecycle/deactivate
    required fields: trusted_origin_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_users:
    endpoint: POST /api/v1/users
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_users_id:
    endpoint: POST /api/v1/users/{{ record.id }}
    required fields: id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_users_id:
    endpoint: PUT /api/v1/users/{{ record.id }}
    required fields: id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_users_id:
    endpoint: DELETE /api/v1/users/{{ record.id }}
    required fields: id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_users_id_lifecycle_activate:
    endpoint: POST /api/v1/users/{{ record.id }}/lifecycle/activate
    required fields: id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_users_id_lifecycle_deactivate:
    endpoint: POST /api/v1/users/{{ record.id }}/lifecycle/deactivate
    required fields: id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_users_id_lifecycle_expire_password:
    endpoint: POST /api/v1/users/{{ record.id }}/lifecycle/expire_password
    required fields: id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_users_id_lifecycle_expire_password_with_temp_password:
    endpoint: POST /api/v1/users/{{ record.id }}/lifecycle/expire_password_with_temp_password
    required fields: id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_users_id_lifecycle_reactivate:
    endpoint: POST /api/v1/users/{{ record.id }}/lifecycle/reactivate
    required fields: id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_users_id_lifecycle_reset_factors:
    endpoint: POST /api/v1/users/{{ record.id }}/lifecycle/reset_factors
    required fields: id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_users_id_lifecycle_suspend:
    endpoint: POST /api/v1/users/{{ record.id }}/lifecycle/suspend
    required fields: id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_users_id_lifecycle_unlock:
    endpoint: POST /api/v1/users/{{ record.id }}/lifecycle/unlock
    required fields: id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_users_id_lifecycle_unsuspend:
    endpoint: POST /api/v1/users/{{ record.id }}/lifecycle/unsuspend
    required fields: id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_users_user_id_or_login_linked_objects_primary_relationship_name_primary_user_id:
    endpoint: PUT /api/v1/users/{{ record.user_id_or_login }}/linkedObjects/{{ record.primary_relationship_name }}/{{ record.primary_user_id }}
    required fields: user_id_or_login, primary_relationship_name, primary_user_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_users_user_id_or_login_linked_objects_relationship_name:
    endpoint: DELETE /api/v1/users/{{ record.user_id_or_login }}/linkedObjects/{{ record.relationship_name }}
    required fields: user_id_or_login, relationship_name
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_users_user_id_authenticator_enrollments_phone:
    endpoint: POST /api/v1/users/{{ record.user_id }}/authenticator-enrollments/phone
    required fields: user_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_users_user_id_authenticator_enrollments_tac:
    endpoint: POST /api/v1/users/{{ record.user_id }}/authenticator-enrollments/tac
    required fields: user_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_users_user_id_authenticator_enrollments_enrollment_id:
    endpoint: DELETE /api/v1/users/{{ record.user_id }}/authenticator-enrollments/{{ record.enrollment_id }}
    required fields: user_id, enrollment_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_users_user_id_classification:
    endpoint: PUT /api/v1/users/{{ record.user_id }}/classification
    required fields: user_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_users_user_id_clients_client_id_grants:
    endpoint: DELETE /api/v1/users/{{ record.user_id }}/clients/{{ record.client_id }}/grants
    required fields: user_id, client_id
    risk: high: external Okta admin API mutation; approval required
  delete_api_v1_users_user_id_clients_client_id_tokens:
    endpoint: DELETE /api/v1/users/{{ record.user_id }}/clients/{{ record.client_id }}/tokens
    required fields: user_id, client_id
    risk: high: external Okta admin API mutation; approval required
  delete_api_v1_users_user_id_clients_client_id_tokens_token_id:
    endpoint: DELETE /api/v1/users/{{ record.user_id }}/clients/{{ record.client_id }}/tokens/{{ record.token_id }}
    required fields: user_id, client_id, token_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_users_user_id_credentials_change_password:
    endpoint: POST /api/v1/users/{{ record.user_id }}/credentials/change_password
    required fields: user_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_users_user_id_credentials_change_recovery_question:
    endpoint: POST /api/v1/users/{{ record.user_id }}/credentials/change_recovery_question
    required fields: user_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_users_user_id_credentials_forgot_password:
    endpoint: POST /api/v1/users/{{ record.user_id }}/credentials/forgot_password
    required fields: user_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_users_user_id_credentials_forgot_password_recovery_question:
    endpoint: POST /api/v1/users/{{ record.user_id }}/credentials/forgot_password_recovery_question
    required fields: user_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_users_user_id_factors:
    endpoint: POST /api/v1/users/{{ record.user_id }}/factors
    required fields: user_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_users_user_id_factors_factor_id:
    endpoint: DELETE /api/v1/users/{{ record.user_id }}/factors/{{ record.factor_id }}
    required fields: user_id, factor_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_users_user_id_factors_factor_id_lifecycle_activate:
    endpoint: POST /api/v1/users/{{ record.user_id }}/factors/{{ record.factor_id }}/lifecycle/activate
    required fields: user_id, factor_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_users_user_id_factors_factor_id_resend:
    endpoint: POST /api/v1/users/{{ record.user_id }}/factors/{{ record.factor_id }}/resend
    required fields: user_id, factor_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_users_user_id_factors_factor_id_verify:
    endpoint: POST /api/v1/users/{{ record.user_id }}/factors/{{ record.factor_id }}/verify
    required fields: user_id, factor_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_users_user_id_grants:
    endpoint: DELETE /api/v1/users/{{ record.user_id }}/grants
    required fields: user_id
    risk: high: external Okta admin API mutation; approval required
  delete_api_v1_users_user_id_grants_grant_id:
    endpoint: DELETE /api/v1/users/{{ record.user_id }}/grants/{{ record.grant_id }}
    required fields: user_id, grant_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_users_user_id_risk:
    endpoint: PUT /api/v1/users/{{ record.user_id }}/risk
    required fields: user_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_users_user_id_roles:
    endpoint: POST /api/v1/users/{{ record.user_id }}/roles
    required fields: user_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_users_user_id_roles_role_assignment_id:
    endpoint: DELETE /api/v1/users/{{ record.user_id }}/roles/{{ record.role_assignment_id }}
    required fields: user_id, role_assignment_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps:
    endpoint: PUT /api/v1/users/{{ record.user_id }}/roles/{{ record.role_assignment_id }}/targets/catalog/apps
    required fields: user_id, role_assignment_id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name:
    endpoint: PUT /api/v1/users/{{ record.user_id }}/roles/{{ record.role_assignment_id }}/targets/catalog/apps/{{ record.app_name }}
    required fields: user_id, role_assignment_id, app_name
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name:
    endpoint: DELETE /api/v1/users/{{ record.user_id }}/roles/{{ record.role_assignment_id }}/targets/catalog/apps/{{ record.app_name }}
    required fields: user_id, role_assignment_id, app_name
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id:
    endpoint: PUT /api/v1/users/{{ record.user_id }}/roles/{{ record.role_assignment_id }}/targets/catalog/apps/{{ record.app_name }}/{{ record.app_id }}
    required fields: user_id, role_assignment_id, app_name, app_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id:
    endpoint: DELETE /api/v1/users/{{ record.user_id }}/roles/{{ record.role_assignment_id }}/targets/catalog/apps/{{ record.app_name }}/{{ record.app_id }}
    required fields: user_id, role_assignment_id, app_name, app_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_users_user_id_roles_role_assignment_id_targets_groups_group_id:
    endpoint: PUT /api/v1/users/{{ record.user_id }}/roles/{{ record.role_assignment_id }}/targets/groups/{{ record.group_id }}
    required fields: user_id, role_assignment_id, group_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_users_user_id_roles_role_assignment_id_targets_groups_group_id:
    endpoint: DELETE /api/v1/users/{{ record.user_id }}/roles/{{ record.role_assignment_id }}/targets/groups/{{ record.group_id }}
    required fields: user_id, role_assignment_id, group_id
    risk: high: external Okta admin API mutation; approval required
  delete_api_v1_users_user_id_sessions:
    endpoint: DELETE /api/v1/users/{{ record.user_id }}/sessions
    required fields: user_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_users_user_id_subscriptions_notification_type_subscribe:
    endpoint: POST /api/v1/users/{{ record.user_id }}/subscriptions/{{ record.notification_type }}/subscribe
    required fields: user_id, notification_type
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_users_user_id_subscriptions_notification_type_unsubscribe:
    endpoint: POST /api/v1/users/{{ record.user_id }}/subscriptions/{{ record.notification_type }}/unsubscribe
    required fields: user_id, notification_type
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_zones:
    endpoint: POST /api/v1/zones
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_zones_zone_id:
    endpoint: PUT /api/v1/zones/{{ record.zone_id }}
    required fields: zone_id
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_zones_zone_id:
    endpoint: DELETE /api/v1/zones/{{ record.zone_id }}
    required fields: zone_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_zones_zone_id_lifecycle_activate:
    endpoint: POST /api/v1/zones/{{ record.zone_id }}/lifecycle/activate
    required fields: zone_id
    risk: high: external Okta admin API mutation; approval required
  execute_api_v1_zones_zone_id_lifecycle_deactivate:
    endpoint: POST /api/v1/zones/{{ record.zone_id }}/lifecycle/deactivate
    required fields: zone_id
    risk: high: external Okta admin API mutation; approval required
  update_attack_protection_api_v1_authenticator_settings:
    endpoint: PUT /attack-protection/api/v1/authenticator-settings
    risk: medium: external Okta admin API mutation; approval required
  update_attack_protection_api_v1_user_lockout_settings:
    endpoint: PUT /attack-protection/api/v1/user-lockout-settings
    risk: medium: external Okta admin API mutation; approval required
  create_integrations_api_v1_api_services:
    endpoint: POST /integrations/api/v1/api-services
    risk: medium: external Okta admin API mutation; approval required
  delete_integrations_api_v1_api_services_api_service_id:
    endpoint: DELETE /integrations/api/v1/api-services/{{ record.api_service_id }}
    required fields: api_service_id
    risk: high: external Okta admin API mutation; approval required
  create_integrations_api_v1_api_services_api_service_id_credentials_secrets:
    endpoint: POST /integrations/api/v1/api-services/{{ record.api_service_id }}/credentials/secrets
    required fields: api_service_id
    risk: medium: external Okta admin API mutation; approval required
  delete_integrations_api_v1_api_services_api_service_id_credentials_secrets_secret_id:
    endpoint: DELETE /integrations/api/v1/api-services/{{ record.api_service_id }}/credentials/secrets/{{ record.secret_id }}
    required fields: api_service_id, secret_id
    risk: high: external Okta admin API mutation; approval required
  execute_integrations_api_v1_api_services_api_service_id_credentials_secrets_secret_id_lifecycle_activate:
    endpoint: POST /integrations/api/v1/api-services/{{ record.api_service_id }}/credentials/secrets/{{ record.secret_id }}/lifecycle/activate
    required fields: api_service_id, secret_id
    risk: high: external Okta admin API mutation; approval required
  execute_integrations_api_v1_api_services_api_service_id_credentials_secrets_secret_id_lifecycle_deactivate:
    endpoint: POST /integrations/api/v1/api-services/{{ record.api_service_id }}/credentials/secrets/{{ record.secret_id }}/lifecycle/deactivate
    required fields: api_service_id, secret_id
    risk: high: external Okta admin API mutation; approval required
  create_oauth2_v1_clients_client_id_roles:
    endpoint: POST /oauth2/v1/clients/{{ record.client_id }}/roles
    required fields: client_id
    risk: medium: external Okta admin API mutation; approval required
  delete_oauth2_v1_clients_client_id_roles_role_assignment_id:
    endpoint: DELETE /oauth2/v1/clients/{{ record.client_id }}/roles/{{ record.role_assignment_id }}
    required fields: client_id, role_assignment_id
    risk: high: external Okta admin API mutation; approval required
  update_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name:
    endpoint: PUT /oauth2/v1/clients/{{ record.client_id }}/roles/{{ record.role_assignment_id }}/targets/catalog/apps/{{ record.app_name }}
    required fields: client_id, role_assignment_id, app_name
    risk: medium: external Okta admin API mutation; approval required
  delete_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name:
    endpoint: DELETE /oauth2/v1/clients/{{ record.client_id }}/roles/{{ record.role_assignment_id }}/targets/catalog/apps/{{ record.app_name }}
    required fields: client_id, role_assignment_id, app_name
    risk: high: external Okta admin API mutation; approval required
  update_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id:
    endpoint: PUT /oauth2/v1/clients/{{ record.client_id }}/roles/{{ record.role_assignment_id }}/targets/catalog/apps/{{ record.app_name }}/{{ record.app_id }}
    required fields: client_id, role_assignment_id, app_name, app_id
    risk: medium: external Okta admin API mutation; approval required
  delete_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id:
    endpoint: DELETE /oauth2/v1/clients/{{ record.client_id }}/roles/{{ record.role_assignment_id }}/targets/catalog/apps/{{ record.app_name }}/{{ record.app_id }}
    required fields: client_id, role_assignment_id, app_name, app_id
    risk: high: external Okta admin API mutation; approval required
  update_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_groups_group_id:
    endpoint: PUT /oauth2/v1/clients/{{ record.client_id }}/roles/{{ record.role_assignment_id }}/targets/groups/{{ record.group_id }}
    required fields: client_id, role_assignment_id, group_id
    risk: medium: external Okta admin API mutation; approval required
  delete_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_groups_group_id:
    endpoint: DELETE /oauth2/v1/clients/{{ record.client_id }}/roles/{{ record.role_assignment_id }}/targets/groups/{{ record.group_id }}
    required fields: client_id, role_assignment_id, group_id
    risk: high: external Okta admin API mutation; approval required
  update_okta_personal_settings_api_v1_edit_feature:
    endpoint: PUT /okta-personal-settings/api/v1/edit-feature
    risk: medium: external Okta admin API mutation; approval required
  update_okta_personal_settings_api_v1_export_blocklists:
    endpoint: PUT /okta-personal-settings/api/v1/export-blocklists
    risk: medium: external Okta admin API mutation; approval required
  create_privileged_access_api_v1_okta_service_accounts:
    endpoint: POST /privileged-access/api/v1/okta-service-accounts
    risk: medium: external Okta admin API mutation; approval required
  update_privileged_access_api_v1_okta_service_accounts_id:
    endpoint: PATCH /privileged-access/api/v1/okta-service-accounts/{{ record.id }}
    required fields: id
    risk: medium: external Okta admin API mutation; approval required
  delete_privileged_access_api_v1_okta_service_accounts_id:
    endpoint: DELETE /privileged-access/api/v1/okta-service-accounts/{{ record.id }}
    required fields: id
    risk: high: external Okta admin API mutation; approval required
  create_privileged_access_api_v1_service_accounts:
    endpoint: POST /privileged-access/api/v1/service-accounts
    risk: medium: external Okta admin API mutation; approval required
  update_privileged_access_api_v1_service_accounts_id:
    endpoint: PATCH /privileged-access/api/v1/service-accounts/{{ record.id }}
    required fields: id
    risk: medium: external Okta admin API mutation; approval required
  delete_privileged_access_api_v1_service_accounts_id:
    endpoint: DELETE /privileged-access/api/v1/service-accounts/{{ record.id }}
    required fields: id
    risk: high: external Okta admin API mutation; approval required
  execute_webauthn_registration_api_v1_activate:
    endpoint: POST /webauthn-registration/api/v1/activate
    risk: high: external Okta admin API mutation; approval required
  create_webauthn_registration_api_v1_enroll:
    endpoint: POST /webauthn-registration/api/v1/enroll
    risk: medium: external Okta admin API mutation; approval required
  create_webauthn_registration_api_v1_initiate_fulfillment_request:
    endpoint: POST /webauthn-registration/api/v1/initiate-fulfillment-request
    risk: medium: external Okta admin API mutation; approval required
  create_webauthn_registration_api_v1_send_pin:
    endpoint: POST /webauthn-registration/api/v1/send-pin
    risk: medium: external Okta admin API mutation; approval required
  delete_webauthn_registration_api_v1_users_user_id_enrollments_authenticator_enrollment_id:
    endpoint: DELETE /webauthn-registration/api/v1/users/{{ record.user_id }}/enrollments/{{ record.authenticator_enrollment_id }}
    required fields: user_id, authenticator_enrollment_id
    risk: high: external Okta admin API mutation; approval required
  create_webauthn_registration_api_v1_users_user_id_enrollments_authenticator_enrollment_id_mark_error:
    endpoint: POST /webauthn-registration/api/v1/users/{{ record.user_id }}/enrollments/{{ record.authenticator_enrollment_id }}/mark-error
    required fields: user_id, authenticator_enrollment_id
    risk: medium: external Okta admin API mutation; approval required

SECURITY
  read risk: external Okta admin API reads across users, groups, logs, apps, policies, authenticators, brands, roles, and related resources
  write risk: external Okta admin API mutations including lifecycle, provisioning, credential, policy, app, user, group, and delete operations
  approval: required for every write action; deletes and lifecycle actions are high risk
  Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

EXAMPLES
  # Inspect as a manual
  pm connectors inspect okta

  # Inspect as structured JSON
  pm connectors inspect okta --json

AGENT WORKFLOW
  - Run pm connectors inspect okta before creating credentials or plans.
  - Use --json only when the caller needs structured output; use the manual for human-readable guidance.
  - Never ask the user to paste secret values into chat.
  - For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.

EXIT STATUS
  0 success
  1 runtime error
  2 usage error

```
