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
  id: okta
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
    fields: created(string), email(string), id(string), last_login(string), login(string), status(string)
  groups:
    primary key: id
    fields: created(string), description(string), id(string), name(string)
  system_logs:
    primary key: uuid
    cursor: published
    fields: display_message(string), event_type(string), published(string), uuid(string)
  well_known_app_authenticator_configuration:
    primary key: name
    fields: appAuthenticatorEnrollEndpoint(string), authenticatorId(string), createdDate(string), key(string), lastUpdated(string), name(string), orgId(string), settings(object), supportedMethods(array), type(string)
  well_known_apple_app_site_association:
  well_known_assetlinks_json:
  well_known_okta_organization:
    primary key: id
    fields: _links(object), id(string), pipeline(string)
  well_known_ssf_configuration:
    fields: authorization_schemes(array), configuration_endpoint(string), default_subjects(string), delivery_methods_supported(array), issuer(string), jwks_uri(string), spec_version(string), verification_endpoint(string)
  well_known_webauthn:
  api_v1_agent_pools:
    primary key: id
    fields: _links(object), agents(array), disruptedAgents(integer), id(string), inactiveAgents(integer), name(string), operationalStatus(string), type(string)
  api_v1_agent_pools_pool_id_updates:
    primary key: id
    fields: _links(object), agentType(string), agents(array), enabled(boolean), id(string), name(string), notifyAdmin(boolean), reason(string), schedule(object), sortOrder(integer), status(string), targetVersion(string)
  api_v1_agent_pools_pool_id_updates_settings:
    fields: agentType(string), continueOnError(boolean), latestVersion(string), minimalSupportedVersion(string), poolId(string), poolName(string), releaseChannel(string)
  api_v1_agent_pools_pool_id_updates_update_id:
    primary key: id
    fields: _links(object), agentType(string), agents(array), enabled(boolean), id(string), name(string), notifyAdmin(boolean), reason(string), schedule(object), sortOrder(integer), status(string), targetVersion(string)
  api_v1_api_tokens:
    primary key: id
    fields: _link(object), clientName(string), created(string), expiresAt(string), id(string), lastUpdated(string), name(string), network(object), tokenWindow(string), userId(string)
  api_v1_api_tokens_api_token_id:
    primary key: id
    fields: _link(object), clientName(string), created(string), expiresAt(string), id(string), lastUpdated(string), name(string), network(object), tokenWindow(string), userId(string)
  api_v1_apps:
    primary key: id
    fields: _embedded(object), _links(object), accessibility(object), created(string), expressConfiguration(object), features(array), id(string), label(string), lastUpdated(string), licensing(object), orn(string), profile(object), signOnMode(string), status(string), universalLogout(object), visibility(object)
  api_v1_apps_app_id:
    primary key: id
    fields: _embedded(object), _links(object), accessibility(object), created(string), expressConfiguration(object), features(array), id(string), label(string), lastUpdated(string), licensing(object), orn(string), profile(object), signOnMode(string), status(string), universalLogout(object), visibility(object)
  api_v1_apps_app_id_connections_default:
    fields: _links(object), authScheme(string), baseUrl(string), profile(object), status(string)
  api_v1_apps_app_id_connections_default_jwks:
    fields: jwks(object)
  api_v1_apps_app_id_credentials_csrs:
    primary key: id
    fields: _links(object), created(string), csr(string), id(string), kty(string)
  api_v1_apps_app_id_credentials_csrs_csr_id:
    primary key: id
    fields: _links(object), created(string), csr(string), id(string), kty(string)
  api_v1_apps_app_id_credentials_jwks:
    primary key: id
    fields: _links(object), created(string), id(string), lastUpdated(string)
  api_v1_apps_app_id_credentials_jwks_key_id:
    primary key: id
    fields: _links(object), created(string), id(string), lastUpdated(string)
  api_v1_apps_app_id_credentials_keys:
    fields: created(string), e(string), expiresAt(string), kid(string), kty(string), lastUpdated(string), n(string), use(string), x5c(array), x5t#S256(string)
  api_v1_apps_app_id_credentials_keys_key_id:
    fields: created(string), e(string), expiresAt(string), kid(string), kty(string), lastUpdated(string), n(string), use(string), x5c(array), x5t#S256(string)
  api_v1_apps_app_id_credentials_secrets:
    primary key: id
    fields: _links(object), client_secret(string), created(string), id(string), lastUpdated(string), secret_hash(string), status(string)
  api_v1_apps_app_id_credentials_secrets_secret_id:
    primary key: id
    fields: _links(object), client_secret(string), created(string), id(string), lastUpdated(string), secret_hash(string), status(string)
  api_v1_apps_app_id_cwo_connections:
    primary key: id
    fields: created(string), id(string), lastUpdated(string), requestingAppInstanceId(string), resourceAppInstanceId(string), status(string)
  api_v1_apps_app_id_cwo_connections_connection_id:
    primary key: id
    fields: created(string), id(string), lastUpdated(string), requestingAppInstanceId(string), resourceAppInstanceId(string), status(string)
  api_v1_apps_app_id_features:
    primary key: name
    fields: _links(object), description(string), name(string), status(string)
  api_v1_apps_app_id_features_feature_name:
    primary key: name
    fields: _links(object), description(string), name(string), status(string)
  api_v1_apps_app_id_federated_claims:
    primary key: id
    fields: created(string), expression(string), id(string), lastUpdated(string), name(string)
  api_v1_apps_app_id_federated_claims_claim_id:
    primary key: name
    fields: expression(string), name(string)
  api_v1_apps_app_id_grants:
    primary key: id
    fields: _embedded(object), _links(object), clientId(string), created(string), createdBy(object), id(string), issuer(string), lastUpdated(string), scopeId(string), source(string), status(string), userId(string)
  api_v1_apps_app_id_grants_grant_id:
    primary key: id
    fields: _embedded(object), _links(object), clientId(string), created(string), createdBy(object), id(string), issuer(string), lastUpdated(string), scopeId(string), source(string), status(string), userId(string)
  api_v1_apps_app_id_group_push_mappings:
    primary key: id
    fields: _links(object), appConfig(object), created(string), errorSummary(string), id(string), lastPush(string), lastUpdated(string), sourceGroupId(string), status(string), targetGroupId(string)
  api_v1_apps_app_id_group_push_mappings_mapping_id:
    primary key: id
    fields: _links(object), appConfig(object), created(string), errorSummary(string), id(string), lastPush(string), lastUpdated(string), sourceGroupId(string), status(string), targetGroupId(string)
  api_v1_apps_app_id_groups:
    primary key: id
    fields: _embedded(object), _links(object), id(string), lastUpdated(string), priority(integer), profile(object)
  api_v1_apps_app_id_groups_group_id:
    primary key: id
    fields: _embedded(object), _links(object), id(string), lastUpdated(string), priority(integer), profile(object)
  api_v1_apps_app_id_tokens:
    primary key: id
    fields: _embedded(object), _links(object), clientId(string), created(string), expiresAt(string), id(string), issuer(string), lastUpdated(string), scopes(array), status(string), userId(string)
  api_v1_apps_app_id_tokens_token_id:
    primary key: id
    fields: _embedded(object), _links(object), clientId(string), created(string), expiresAt(string), id(string), issuer(string), lastUpdated(string), scopes(array), status(string), userId(string)
  api_v1_apps_app_id_users:
    primary key: id
    fields: _embedded(object), _links(object), created(string), credentials(object), externalId(string), id(string), lastSync(string), lastUpdated(string), passwordChanged(string), profile(object), scope(string), status(string), statusChanged(string), syncState(string)
  api_v1_apps_app_id_users_user_id:
    primary key: id
    fields: _embedded(object), _links(object), created(string), credentials(object), externalId(string), id(string), lastSync(string), lastUpdated(string), passwordChanged(string), profile(object), scope(string), status(string), statusChanged(string), syncState(string)
  api_v1_authenticators:
    primary key: id
    fields: _links(object), created(string), description(string), id(string), key(string), lastUpdated(string), name(string), status(string), type(string)
  api_v1_authenticators_authenticator_id:
    primary key: id
    fields: _links(object), created(string), description(string), id(string), key(string), lastUpdated(string), name(string), status(string), type(string)
  api_v1_authenticators_authenticator_id_aaguids:
    primary key: name
    fields: _links(object), aaguid(string), attestationRootCertificates(array), authenticatorCharacteristics(object), name(string)
  api_v1_authenticators_authenticator_id_aaguids_aaguid:
    primary key: name
    fields: _links(object), aaguid(string), attestationRootCertificates(array), authenticatorCharacteristics(object), name(string)
  api_v1_authenticators_authenticator_id_methods:
    fields: _links(object), status(string), type(string)
  api_v1_authenticators_authenticator_id_methods_method_type:
    fields: _links(object), status(string), type(string)
  api_v1_authorization_servers:
    primary key: id
    fields: _links(object), accessTokenEncryptedResponseAlgorithm(string), audiences(array), created(string), credentials(object), description(string), id(string), issuer(string), issuerMode(string), jwks(object), jwks_uri(string), lastUpdated(string), name(string), status(string)
  api_v1_authorization_servers_auth_server_id:
    primary key: id
    fields: _links(object), accessTokenEncryptedResponseAlgorithm(string), audiences(array), created(string), credentials(object), description(string), id(string), issuer(string), issuerMode(string), jwks(object), jwks_uri(string), lastUpdated(string), name(string), status(string)
  api_v1_authorization_servers_auth_server_id_associated_servers:
    primary key: id
    fields: _links(object), accessTokenEncryptedResponseAlgorithm(string), audiences(array), created(string), credentials(object), description(string), id(string), issuer(string), issuerMode(string), jwks(object), jwks_uri(string), lastUpdated(string), name(string), status(string)
  api_v1_authorization_servers_auth_server_id_claims:
    primary key: id
    fields: _links(object), alwaysIncludeInToken(boolean), claimType(string), conditions(object), group_filter_type(string), id(string), name(string), status(string), system(boolean), value(string), valueType(string)
  api_v1_authorization_servers_auth_server_id_claims_claim_id:
    primary key: id
    fields: _links(object), alwaysIncludeInToken(boolean), claimType(string), conditions(object), group_filter_type(string), id(string), name(string), status(string), system(boolean), value(string), valueType(string)
  api_v1_authorization_servers_auth_server_id_clients:
    fields: _links(object), client_id(string), client_name(string), client_uri(string), logo_uri(string)
  api_v1_authorization_servers_auth_server_id_clients_client_id_tokens:
    primary key: id
    fields: _embedded(object), _links(object), clientId(string), created(string), expiresAt(string), id(string), issuer(string), lastUpdated(string), scopes(array), status(string), userId(string)
  api_v1_authorization_servers_auth_server_id_clients_client_id_tokens_token_id:
    primary key: id
    fields: _embedded(object), _links(object), clientId(string), created(string), expiresAt(string), id(string), issuer(string), lastUpdated(string), scopes(array), status(string), userId(string)
  api_v1_authorization_servers_auth_server_id_credentials_keys:
    fields: _links(object), alg(string), e(string), kid(string), kty(string), n(string), status(string), use(string)
  api_v1_authorization_servers_auth_server_id_credentials_keys_key_id:
    fields: _links(object), alg(string), e(string), kid(string), kty(string), n(string), status(string), use(string)
  api_v1_authorization_servers_auth_server_id_policies:
    primary key: id
    fields: _links(object), conditions(object), created(string), description(string), id(string), lastUpdated(string), name(string), priority(integer), status(string), system(boolean), type(string)
  api_v1_authorization_servers_auth_server_id_policies_policy_id:
    primary key: id
    fields: _links(object), conditions(object), created(string), description(string), id(string), lastUpdated(string), name(string), priority(integer), status(string), system(boolean), type(string)
  api_v1_authorization_servers_auth_server_id_policies_policy_id_rules:
    primary key: id
    fields: _links(object), actions(object), conditions(object), created(string), id(string), lastUpdated(string), name(string), priority(integer), status(string), system(boolean), type(string)
  api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id:
    primary key: id
    fields: _links(object), actions(object), conditions(object), created(string), id(string), lastUpdated(string), name(string), priority(integer), status(string), system(boolean), type(string)
  api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys:
    primary key: id
    fields: _links(object), created(string), e(string), id(string), kid(string), kty(string), lastUpdated(string), n(string), status(string), use(string)
  api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id:
    primary key: id
    fields: _links(object), created(string), e(string), id(string), kid(string), kty(string), lastUpdated(string), n(string), status(string), use(string)
  api_v1_authorization_servers_auth_server_id_scopes:
    primary key: id
    fields: _links(object), consent(string), default(boolean), description(string), displayName(string), id(string), metadataPublish(string), name(string), optional(boolean), system(boolean)
  api_v1_authorization_servers_auth_server_id_scopes_scope_id:
    primary key: id
    fields: _links(object), consent(string), default(boolean), description(string), displayName(string), id(string), metadataPublish(string), name(string), optional(boolean), system(boolean)
  api_v1_behaviors:
    primary key: id
    fields: _link(object), created(string), id(string), lastUpdated(string), name(string), status(string), type(string)
  api_v1_behaviors_behavior_id:
    primary key: id
    fields: _link(object), created(string), id(string), lastUpdated(string), name(string), status(string), type(string)
  api_v1_bot_protection_configuration:
    fields: _links(object), enforcementType(string), level(string), mode(string), supportedFlows(array)
  api_v1_brands:
    primary key: id
    fields: agreeToCustomPrivacyPolicy(boolean), customPrivacyPolicyUrl(string), defaultApp(object), emailDomainId(string), id(string), isDefault(boolean), locale(string), name(string), removePoweredByOkta(boolean)
  api_v1_brands_brand_id:
    primary key: id
    fields: agreeToCustomPrivacyPolicy(boolean), customPrivacyPolicyUrl(string), defaultApp(object), emailDomainId(string), id(string), isDefault(boolean), locale(string), name(string), removePoweredByOkta(boolean)
  api_v1_brands_brand_id_domains:
    fields: domains(array)
  api_v1_brands_brand_id_pages_error:
    fields: _embedded(object), _links(object)
  api_v1_brands_brand_id_pages_error_customized:
    fields: contentSecurityPolicySetting(object)
  api_v1_brands_brand_id_pages_error_default:
    fields: contentSecurityPolicySetting(object)
  api_v1_brands_brand_id_pages_error_preview:
    fields: contentSecurityPolicySetting(object)
  api_v1_brands_brand_id_pages_sign_in:
    fields: _embedded(object), _links(object)
  api_v1_brands_brand_id_pages_sign_in_customized:
    fields: contentSecurityPolicySetting(object), widgetCustomizations(object), widgetVersion(string)
  api_v1_brands_brand_id_pages_sign_in_default:
    fields: contentSecurityPolicySetting(object), widgetCustomizations(object), widgetVersion(string)
  api_v1_brands_brand_id_pages_sign_in_preview:
    fields: contentSecurityPolicySetting(object), widgetCustomizations(object), widgetVersion(string)
  api_v1_brands_brand_id_pages_sign_out_customized:
    fields: type(string), url(string)
  api_v1_brands_brand_id_templates_email:
    primary key: name
    fields: _embedded(object), _links(object), name(string)
  api_v1_brands_brand_id_templates_email_template_name:
    primary key: name
    fields: _embedded(object), _links(object), name(string)
  api_v1_brands_brand_id_templates_email_template_name_customizations:
    primary key: id
    fields: _links(object), created(string), id(string), isDefault(boolean), language(string), lastUpdated(string)
  api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id:
    primary key: id
    fields: _links(object), created(string), id(string), isDefault(boolean), language(string), lastUpdated(string)
  api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id_preview:
    fields: _links(object), body(string), subject(string)
  api_v1_brands_brand_id_templates_email_template_name_default_content:
    fields: _links(object)
  api_v1_brands_brand_id_templates_email_template_name_default_content_preview:
    fields: _links(object), body(string), subject(string)
  api_v1_brands_brand_id_templates_email_template_name_settings:
    fields: _links(object), recipients(string)
  api_v1_brands_brand_id_themes:
    primary key: id
    fields: _links(object), backgroundImage(string), emailTemplateTouchPointVariant(string), endUserDashboardTouchPointVariant(string), errorPageTouchPointVariant(string), favicon(string), id(string), loadingPageTouchPointVariant(string), logo(string), primaryColorContrastHex(string), primaryColorHex(string), secondaryColorContrastHex(string), secondaryColorHex(string), signInPageTouchPointVariant(string)
  api_v1_brands_brand_id_themes_theme_id:
    primary key: id
    fields: _links(object), backgroundImage(string), emailTemplateTouchPointVariant(string), endUserDashboardTouchPointVariant(string), errorPageTouchPointVariant(string), favicon(string), id(string), loadingPageTouchPointVariant(string), logo(string), primaryColorContrastHex(string), primaryColorHex(string), secondaryColorContrastHex(string), secondaryColorHex(string), signInPageTouchPointVariant(string)
  api_v1_brands_brand_id_well_known_uris:
    fields: _embedded(object), _links(object)
  api_v1_brands_brand_id_well_known_uris_path:
    fields: _links(object), representation(object)
  api_v1_brands_brand_id_well_known_uris_path_customized:
    fields: _links(object), representation(object)
  api_v1_captchas:
    primary key: id
    fields: _links(object), id(string), name(string), secretKey(string), siteKey(string), type(string)
  api_v1_captchas_captcha_id:
    primary key: id
    fields: _links(object), id(string), name(string), secretKey(string), siteKey(string), type(string)
  api_v1_device_assurances:
    primary key: id
    fields: _links(object), createdBy(string), createdDate(string), devicePostureChecks(object), displayRemediationMode(string), gracePeriod(object), id(string), lastUpdate(string), lastUpdatedBy(string), name(string), platform(string)
  api_v1_device_assurances_device_assurance_id:
    primary key: id
    fields: _links(object), createdBy(string), createdDate(string), devicePostureChecks(object), displayRemediationMode(string), gracePeriod(object), id(string), lastUpdate(string), lastUpdatedBy(string), name(string), platform(string)
  api_v1_device_integrations:
    primary key: id
    fields: _links(object), displayName(string), id(string), metadata(object), name(string), platform(string), status(string)
  api_v1_device_integrations_device_integration_id:
    primary key: id
    fields: _links(object), displayName(string), id(string), metadata(object), name(string), platform(string), status(string)
  api_v1_device_posture_checks:
    primary key: id
    fields: _links(object), createdBy(string), createdDate(string), description(string), id(string), lastUpdate(string), lastUpdatedBy(string), mappingType(string), name(string), platform(string), query(string), remediationSettings(object), type(string), variableName(string)
  api_v1_device_posture_checks_default:
    primary key: id
    fields: _links(object), createdBy(string), createdDate(string), description(string), id(string), lastUpdate(string), lastUpdatedBy(string), mappingType(string), name(string), platform(string), query(string), remediationSettings(object), type(string), variableName(string)
  api_v1_device_posture_checks_posture_check_id:
    primary key: id
    fields: _links(object), createdBy(string), createdDate(string), description(string), id(string), lastUpdate(string), lastUpdatedBy(string), mappingType(string), name(string), platform(string), query(string), remediationSettings(object), type(string), variableName(string)
  api_v1_devices:
    fields: _embedded(object)
  api_v1_devices_device_id:
    fields: providers(array)
  api_v1_devices_device_id_users:
    fields: created(string), managementStatus(string), screenLockType(string), user(object)
  api_v1_directories_app_instance_id_groups_group_id_query_result_id:
    primary key: id
    fields: id(string), profile(object)
  api_v1_domains:
    fields: domains(array)
  api_v1_domains_domain_id:
    primary key: id
    fields: _links(object), brandId(string), certificateSourceType(string), dnsRecords(array), domain(string), id(string), publicCertificate(object), validationStatus(string)
  api_v1_dr_status:
    fields: status(array)
  api_v1_dr_status_domain:
    fields: status(array)
  api_v1_email_domains:
    fields: displayName(string), userName(string)
  api_v1_email_domains_email_domain_id:
    fields: displayName(string), userName(string)
  api_v1_email_servers:
    fields: email-servers(array)
  api_v1_email_servers_email_server_id:
    primary key: id
    fields: alias(string), authType(string), enabled(boolean), host(string), id(string), port(integer), username(string)
  api_v1_event_hooks:
    primary key: id
    fields: _links(object), channel(object), created(string), createdBy(string), description(string), events(object), id(string), lastUpdated(string), name(string), status(string), verificationStatus(string)
  api_v1_event_hooks_event_hook_id:
    primary key: id
    fields: _links(object), channel(object), created(string), createdBy(string), description(string), events(object), id(string), lastUpdated(string), name(string), status(string), verificationStatus(string)
  api_v1_features:
    primary key: id
    fields: _links(object), description(string), id(string), name(string), stage(object), status(string), type(string)
  api_v1_features_feature_id:
    primary key: id
    fields: _links(object), description(string), id(string), name(string), stage(object), status(string), type(string)
  api_v1_features_feature_id_dependencies:
    primary key: id
    fields: _links(object), description(string), id(string), name(string), stage(object), status(string), type(string)
  api_v1_features_feature_id_dependents:
    primary key: id
    fields: _links(object), description(string), id(string), name(string), stage(object), status(string), type(string)
  api_v1_first_party_app_settings_app_name:
    fields: sessionIdleTimeoutMinutes(integer), sessionMaxLifetimeMinutes(integer)
  api_v1_groups_rules:
    primary key: id
    fields: _embedded(object), actions(object), conditions(object), created(string), id(string), lastUpdated(string), name(string), status(string), type(string)
  api_v1_groups_rules_group_rule_id:
    primary key: id
    fields: _embedded(object), actions(object), conditions(object), created(string), id(string), lastUpdated(string), name(string), status(string), type(string)
  api_v1_groups_group_id:
    primary key: id
    fields: _embedded(object), _links(object), created(string), id(string), lastMembershipUpdated(string), lastUpdated(string), objectClass(array), profile(object), type(string)
  api_v1_groups_group_id_apps:
    primary key: id
    fields: _embedded(object), _links(object), accessibility(object), created(string), expressConfiguration(object), features(array), id(string), label(string), lastUpdated(string), licensing(object), orn(string), profile(object), signOnMode(string), status(string), universalLogout(object), visibility(object)
  api_v1_groups_group_id_owners:
    primary key: id
    fields: displayName(string), id(string), lastUpdated(string), originId(string), originType(string), resolved(boolean), type(string)
  api_v1_groups_group_id_roles:
    primary key: id
    fields: _embedded(object), _links(object), assignmentType(string), created(string), id(string), label(string), lastUpdated(string), status(string), type(string)
  api_v1_groups_group_id_roles_role_assignment_id:
    primary key: id
    fields: _embedded(object), _links(object), assignmentType(string), created(string), id(string), label(string), lastUpdated(string), status(string), type(string)
  api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps:
    primary key: id
    fields: _links(object), category(string), description(string), displayName(string), features(array), id(string), lastUpdated(string), name(string), signOnModes(array), status(string), verificationStatus(string), website(string)
  api_v1_groups_group_id_roles_role_assignment_id_targets_groups:
    primary key: id
    fields: _embedded(object), _links(object), created(string), id(string), lastMembershipUpdated(string), lastUpdated(string), objectClass(array), profile(object), type(string)
  api_v1_groups_group_id_users:
    primary key: id
    fields: _embedded(object), _links(object), activated(string), created(string), credentials(object), id(string), lastLogin(string), lastUpdated(string), passwordChanged(string), profile(object), realmId(string), status(string), statusChanged(string), transitioningToStatus(string), type(object)
  api_v1_hook_keys:
    primary key: id
    fields: created(string), id(string), isUsed(string), keyId(string), lastUpdated(string), name(string)
  api_v1_hook_keys_public_key_id:
    fields: alg(string), e(string), kid(string), kty(string), n(string), use(string)
  api_v1_hook_keys_id:
    primary key: id
    fields: created(string), id(string), isUsed(string), keyId(string), lastUpdated(string), name(string)
  api_v1_iam_assignees_users:
    primary key: id
    fields: _links(object), id(string), orn(string)
  api_v1_iam_governance_bundles:
    fields: _links(object), bundles(array)
  api_v1_iam_governance_bundles_bundle_id:
    primary key: id
    fields: _links(object), description(string), id(string), name(string), orn(string), status(string)
  api_v1_iam_governance_bundles_bundle_id_entitlements:
    fields: _links(object), entitlements(array)
  api_v1_iam_governance_bundles_bundle_id_entitlements_entitlement_id_values:
    fields: _links(object), entitlementValues(array)
  api_v1_iam_governance_opt_in:
    fields: _links(object), optInStatus(string)
  api_v1_iam_resource_sets:
    fields: _links(object), resource-sets(array)
  api_v1_iam_resource_sets_resource_set_id_or_label:
    primary key: id
    fields: _links(object), created(string), description(string), id(string), label(string), lastUpdated(string)
  api_v1_iam_resource_sets_resource_set_id_or_label_bindings:
    primary key: id
    fields: _links(object), id(string)
  api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label:
    primary key: id
    fields: _links(object), id(string)
  api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members:
    fields: _links(object), members(array)
  api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members_member_id:
    primary key: id
    fields: _links(object), created(string), id(string), lastUpdated(string)
  api_v1_iam_resource_sets_resource_set_id_or_label_resources:
    fields: _links(object), resources(array)
  api_v1_iam_resource_sets_resource_set_id_or_label_resources_resource_id:
    primary key: id
    fields: _links(object), conditions(object), created(string), id(string), lastUpdated(string), orn(string)
  api_v1_iam_roles:
    primary key: id
    fields: _links(object), created(string), description(string), id(string), label(string), lastUpdated(string)
  api_v1_iam_roles_role_id_or_label:
    primary key: id
    fields: _links(object), created(string), description(string), id(string), label(string), lastUpdated(string)
  api_v1_iam_roles_role_id_or_label_permissions:
    fields: _links(object), conditions(object), created(string), label(string), lastUpdated(string)
  api_v1_iam_roles_role_id_or_label_permissions_permission_type:
    fields: _links(object), conditions(object), created(string), label(string), lastUpdated(string)
  api_v1_identity_sources_identity_source_id_groups_group_or_external_id:
    primary key: id
    fields: externalId(string), id(string), profile(object)
  api_v1_identity_sources_identity_source_id_groups_group_or_external_id_membership:
    fields: memberExternalIds(array)
  api_v1_identity_sources_identity_source_id_sessions:
    primary key: id
    fields: created(string), id(string), identitySourceId(string), importType(string), lastUpdated(string), status(string)
  api_v1_identity_sources_identity_source_id_sessions_session_id:
    primary key: id
    fields: created(string), id(string), identitySourceId(string), importType(string), lastUpdated(string), status(string)
  api_v1_identity_sources_identity_source_id_users_external_id:
    primary key: id
    fields: created(string), externalId(string), id(string), lastUpdated(string), profile(object)
  api_v1_idps:
    primary key: id
    fields: _links(object), created(string), id(string), issuerMode(string), lastUpdated(string), name(string), policy(object), properties(object), protocol(object), status(string), type(string)
  api_v1_idps_credentials_keys:
    fields: created(string), e(string), expiresAt(string), kid(string), kty(string), lastUpdated(string), n(string), use(string), x5c(array), x5t#S256(string)
  api_v1_idps_credentials_keys_kid:
    fields: created(string), e(string), expiresAt(string), kid(string), kty(string), lastUpdated(string), n(string), use(string), x5c(array), x5t#S256(string)
  api_v1_idps_idp_id:
    primary key: id
    fields: _links(object), created(string), id(string), issuerMode(string), lastUpdated(string), name(string), policy(object), properties(object), protocol(object), status(string), type(string)
  api_v1_idps_idp_id_credentials_csrs:
    primary key: id
    fields: _links(object), created(string), csr(string), id(string), kty(string)
  api_v1_idps_idp_id_credentials_csrs_idp_csr_id:
    primary key: id
    fields: _links(object), created(string), csr(string), id(string), kty(string)
  api_v1_idps_idp_id_credentials_keys:
    fields: created(string), e(string), expiresAt(string), kid(string), kty(string), lastUpdated(string), n(string), use(string), x5c(array), x5t#S256(string)
  api_v1_idps_idp_id_credentials_keys_active:
    fields: created(string), e(string), expiresAt(string), kid(string), kty(string), lastUpdated(string), n(string), use(string), x5c(array), x5t#S256(string)
  api_v1_idps_idp_id_credentials_keys_kid:
    fields: created(string), e(string), expiresAt(string), kid(string), kty(string), lastUpdated(string), n(string), use(string), x5c(array), x5t#S256(string)
  api_v1_idps_idp_id_users:
    primary key: id
    fields: _embedded(object), _links(object), created(string), externalId(string), id(string), lastUpdated(string), profile(object)
  api_v1_idps_idp_id_users_user_id:
    primary key: id
    fields: _embedded(object), _links(object), created(string), externalId(string), id(string), lastUpdated(string), profile(object)
  api_v1_idps_idp_id_users_user_id_credentials_tokens:
    primary key: id
    fields: expiresAt(string), id(string), scopes(array), token(string), tokenAuthScheme(string), tokenType(string)
  api_v1_inline_hooks:
    primary key: id
    fields: _links(object), channel(object), created(string), id(string), lastUpdated(string), name(string), status(string), type(string), version(string)
  api_v1_inline_hooks_inline_hook_id:
    primary key: id
    fields: _links(object), channel(object), created(string), id(string), lastUpdated(string), name(string), status(string), type(string), version(string)
  api_v1_log_streams:
    primary key: id
    fields: _links(object), created(string), id(string), lastUpdated(string), name(string), status(string), type(string)
  api_v1_log_streams_log_stream_id:
    primary key: id
    fields: _links(object), created(string), id(string), lastUpdated(string), name(string), status(string), type(string)
  api_v1_mappings:
    primary key: id
    fields: _links(object), id(string), source(object), target(object)
  api_v1_mappings_mapping_id:
    primary key: id
    fields: _links(object), id(string), properties(object), source(object), target(object)
  api_v1_meta_schemas_apps_app_id_default:
    primary key: id
    fields: $schema(string), _links(object), created(string), definitions(object), id(string), lastUpdated(string), name(string), properties(object), title(string), type(string)
  api_v1_meta_schemas_group_default:
    primary key: id
    fields: $schema(string), _links(object), created(string), definitions(object), description(string), id(string), lastUpdated(string), name(string), properties(object), title(string), type(string)
  api_v1_meta_schemas_log_stream:
    primary key: id
    fields: $schema(string), _links(object), errorMessage(object), id(string), oneOf(array), pattern(string), properties(object), required(array), title(string), type(string)
  api_v1_meta_schemas_log_stream_log_stream_type:
    primary key: id
    fields: $schema(string), _links(object), errorMessage(object), id(string), oneOf(array), pattern(string), properties(object), required(array), title(string), type(string)
  api_v1_meta_schemas_user_linked_objects:
    fields: _links(object), associated(object), primary(object)
  api_v1_meta_schemas_user_linked_objects_linked_object_name:
    fields: _links(object), associated(object), primary(object)
  api_v1_meta_schemas_user_schema_id:
    primary key: id
    fields: $schema(string), _links(object), created(string), definitions(object), id(string), lastUpdated(string), name(string), properties(object), title(string), type(string)
  api_v1_meta_types_user:
    primary key: id
    fields: _links(object), created(string), createdBy(string), default(boolean), description(string), displayName(string), id(string), lastUpdated(string), lastUpdatedBy(string), name(string)
  api_v1_meta_types_user_type_id:
    primary key: id
    fields: _links(object), created(string), createdBy(string), default(boolean), description(string), displayName(string), id(string), lastUpdated(string), lastUpdatedBy(string), name(string)
  api_v1_meta_uischemas:
    primary key: id
    fields: _links(object), created(string), id(string), lastUpdated(string), uiSchema(object)
  api_v1_meta_uischemas_id:
    primary key: id
    fields: _links(object), created(string), id(string), lastUpdated(string), uiSchema(object)
  api_v1_org:
    primary key: id
    fields: _links(object), address1(string), address2(string), city(string), companyName(string), country(string), created(string), endUserSupportHelpURL(string), expiresAt(string), id(string), lastUpdated(string), phoneNumber(string), postalCode(string), state(string), status(string), subdomain(string), supportPhoneNumber(string), website(string)
  api_v1_org_captcha:
    fields: _links(object), captchaId(string), enabledPages(array)
  api_v1_org_contacts:
    fields: _links(object), contactType(string)
  api_v1_org_contacts_contact_type:
    fields: _links(object), userId(string)
  api_v1_org_factors_yubikey_token_tokens:
    primary key: id
    fields: _embedded(object), _links(object), created(string), id(string), lastUpdated(string), lastVerified(string), profile(object), status(string)
  api_v1_org_factors_yubikey_token_tokens_token_id:
    primary key: id
    fields: _embedded(object), _links(object), created(string), id(string), lastUpdated(string), lastVerified(string), profile(object), status(string)
  api_v1_org_org_settings_third_party_admin_setting:
    fields: thirdPartyAdmin(boolean)
  api_v1_org_preferences:
    fields: _links(object), showEndUserFooter(boolean)
  api_v1_org_privacy_aerial:
    fields: _links(object), accountId(string), grantedBy(string), grantedDate(string)
  api_v1_org_privacy_okta_communication:
    fields: _links(object), optOutEmailUsers(boolean)
  api_v1_org_privacy_okta_support:
    fields: _links(object), caseNumber(string), expiration(string), support(string)
  api_v1_org_privacy_okta_support_cases:
    fields: supportCases(array)
  api_v1_org_settings_auto_assign_admin_app_setting:
    fields: autoAssignAdminAppSetting(boolean)
  api_v1_org_settings_client_privileges_setting:
    fields: clientPrivilegesSetting(boolean)
  api_v1_policies:
    primary key: id
    fields: _embedded(object), _links(object), created(string), description(string), id(string), lastUpdated(string), name(string), priority(integer), status(string), system(boolean), type(string)
  api_v1_policies_policy_id:
    primary key: id
    fields: _embedded(object), _links(object), created(string), description(string), id(string), lastUpdated(string), name(string), priority(integer), status(string), system(boolean), type(string)
  api_v1_policies_policy_id_app:
    primary key: id
    fields: _embedded(object), _links(object), accessibility(object), created(string), expressConfiguration(object), features(array), id(string), label(string), lastUpdated(string), licensing(object), orn(string), profile(object), signOnMode(string), status(string), universalLogout(object), visibility(object)
  api_v1_policies_policy_id_mappings:
    primary key: id
    fields: _links(object), id(string)
  api_v1_policies_policy_id_mappings_mapping_id:
    primary key: id
    fields: _links(object), id(string)
  api_v1_policies_policy_id_rules:
    primary key: id
    fields: _links(object), created(string), id(string), lastUpdated(string), name(string), priority(integer), status(string), system(boolean), type(string)
  api_v1_policies_policy_id_rules_rule_id:
    primary key: id
    fields: _links(object), created(string), id(string), lastUpdated(string), name(string), priority(integer), status(string), system(boolean), type(string)
  api_v1_principal_rate_limits:
    primary key: id
    fields: createdBy(string), createdDate(string), defaultConcurrencyPercentage(integer), defaultPercentage(integer), id(string), lastUpdate(string), lastUpdatedBy(string), orgId(string), principalId(string), principalType(string)
  api_v1_principal_rate_limits_principal_rate_limit_id:
    primary key: id
    fields: createdBy(string), createdDate(string), defaultConcurrencyPercentage(integer), defaultPercentage(integer), id(string), lastUpdate(string), lastUpdatedBy(string), orgId(string), principalId(string), principalType(string)
  api_v1_push_providers:
    primary key: id
    fields: _links(object), id(string), lastUpdatedDate(string), name(string), providerType(string)
  api_v1_push_providers_push_provider_id:
    primary key: id
    fields: _links(object), id(string), lastUpdatedDate(string), name(string), providerType(string)
  api_v1_rate_limit_settings_admin_notifications:
    fields: notificationsEnabled(boolean)
  api_v1_rate_limit_settings_per_client:
    fields: defaultMode(string), useCaseModeOverrides(object)
  api_v1_rate_limit_settings_warning_threshold:
    fields: warningThreshold(integer)
  api_v1_realm_assignments:
    primary key: id
    fields: _links(object), actions(object), conditions(object), created(string), domains(array), id(string), isDefault(boolean), lastUpdated(string), name(string), priority(integer), status(string)
  api_v1_realm_assignments_operations:
    fields: _links(object), assignmentOperation(object), numUserMoved(number), realmId(string), realmName(string)
  api_v1_realm_assignments_assignment_id:
    primary key: id
    fields: _links(object), actions(object), conditions(object), created(string), domains(array), id(string), isDefault(boolean), lastUpdated(string), name(string), priority(integer), status(string)
  api_v1_realms:
    primary key: id
    fields: _links(object), created(string), id(string), isDefault(boolean), lastUpdated(string), profile(object)
  api_v1_realms_realm_id:
    primary key: id
    fields: _links(object), created(string), id(string), isDefault(boolean), lastUpdated(string), profile(object)
  api_v1_roles_role_ref_subscriptions:
    fields: _links(object), channels(array), notificationType(string), status(string)
  api_v1_roles_role_ref_subscriptions_notification_type:
    fields: _links(object), channels(array), notificationType(string), status(string)
  api_v1_security_events_providers:
    primary key: id
    fields: _links(object), id(string), name(string), settings(object), status(string), type(string)
  api_v1_security_events_providers_security_event_provider_id:
    primary key: id
    fields: _links(object), id(string), name(string), settings(object), status(string), type(string)
  api_v1_sessions_session_id:
    primary key: id
    fields: _links(object), amr(array), createdAt(string), expiresAt(string), id(string), idp(object), lastFactorVerification(string), lastPasswordVerification(string), login(string), status(string), userId(string)
  api_v1_ssf_stream:
    fields: aud(string), delivery(object), events_delivered(array), events_requested(array), events_supported(array), format(string), iss(string), min_verification_interval(integer), stream_id(string)
  api_v1_ssf_stream_status:
    fields: status(string), stream_id(string)
  api_v1_telephony_providers:
    primary key: id
    fields: enabled(boolean), id(string), isPrimaryProvider(boolean), providerCapability(string), providerName(string), providerSettings(object), providerSid(string)
  api_v1_telephony_providers_custom_telephony_provider_id:
    primary key: id
    fields: enabled(boolean), id(string), isPrimaryProvider(boolean), providerCapability(string), providerName(string), providerSettings(object), providerSid(string)
  api_v1_templates_sms:
    primary key: id
    fields: created(string), id(string), lastUpdated(string), name(string), template(string), translations(object), type(string)
  api_v1_templates_sms_template_id:
    primary key: id
    fields: created(string), id(string), lastUpdated(string), name(string), template(string), translations(object), type(string)
  api_v1_threats_configuration:
    fields: _links(object), action(string), created(string), excludeZones(array), lastUpdated(string)
  api_v1_trusted_origins:
    primary key: id
    fields: _links(object), created(string), createdBy(string), id(string), lastUpdated(string), lastUpdatedBy(string), name(string), origin(string), scopes(array), status(string)
  api_v1_trusted_origins_trusted_origin_id:
    fields: allowedOktaApps(array), type(string)
  api_v1_users_id:
    fields: _embedded(object)
  api_v1_users_id_app_links:
    primary key: id
    fields: appAssignmentId(string), appInstanceId(string), appName(string), credentialsSetup(boolean), hidden(boolean), id(string), label(string), linkUrl(string), logoUrl(string), sortOrder(integer)
  api_v1_users_id_blocks:
    fields: appliesTo(string), type(string)
  api_v1_users_id_groups:
    primary key: id
    fields: _embedded(object), _links(object), created(string), id(string), lastMembershipUpdated(string), lastUpdated(string), objectClass(array), profile(object), type(string)
  api_v1_users_id_idps:
    primary key: id
    fields: _links(object), created(string), id(string), issuerMode(string), lastUpdated(string), name(string), policy(object), properties(object), protocol(object), status(string), type(string)
  api_v1_users_user_id_or_login_linked_objects_relationship_name:
    fields: _links(object)
  api_v1_users_user_id_authenticator_enrollments:
    primary key: id
    fields: _links(object), created(string), id(string), key(string), lastUpdated(string), name(string), profile(object), status(string), type(string)
  api_v1_users_user_id_authenticator_enrollments_enrollment_id:
    primary key: id
    fields: _links(object), created(string), id(string), key(string), lastUpdated(string), name(string), profile(object), status(string), type(string)
  api_v1_users_user_id_classification:
    fields: lastUpdated(string), type(string)
  api_v1_users_user_id_clients:
    fields: _links(object), client_id(string), client_name(string), client_uri(string), logo_uri(string)
  api_v1_users_user_id_clients_client_id_grants:
    primary key: id
    fields: _embedded(object), _links(object), clientId(string), created(string), createdBy(object), id(string), issuer(string), lastUpdated(string), scopeId(string), source(string), status(string), userId(string)
  api_v1_users_user_id_clients_client_id_tokens:
    primary key: id
    fields: _embedded(object), _links(object), clientId(string), created(string), expiresAt(string), id(string), issuer(string), lastUpdated(string), scopes(array), status(string), userId(string)
  api_v1_users_user_id_clients_client_id_tokens_token_id:
    primary key: id
    fields: _embedded(object), _links(object), clientId(string), created(string), expiresAt(string), id(string), issuer(string), lastUpdated(string), scopes(array), status(string), userId(string)
  api_v1_users_user_id_devices:
    fields: created(string), device(object), deviceUserId(string)
  api_v1_users_user_id_factors:
    primary key: id
    fields: _embedded(object), _links(object), created(string), factorType(string), id(string), lastUpdated(string), profile(object), provider(string), status(string), vendorName(string)
  api_v1_users_user_id_factors_catalog:
    fields: _embedded(object), _links(object), enrollment(string), factorType(string), provider(string), status(string), vendorName(string)
  api_v1_users_user_id_factors_questions:
    fields: answer(string), question(string), questionText(string)
  api_v1_users_user_id_factors_factor_id:
    primary key: id
    fields: _embedded(object), _links(object), created(string), factorType(string), id(string), lastUpdated(string), profile(object), provider(string), status(string), vendorName(string)
  api_v1_users_user_id_factors_factor_id_transactions_transaction_id:
    fields: factorResult(string)
  api_v1_users_user_id_grants:
    primary key: id
    fields: _embedded(object), _links(object), clientId(string), created(string), createdBy(object), id(string), issuer(string), lastUpdated(string), scopeId(string), source(string), status(string), userId(string)
  api_v1_users_user_id_grants_grant_id:
    primary key: id
    fields: _embedded(object), _links(object), clientId(string), created(string), createdBy(object), id(string), issuer(string), lastUpdated(string), scopeId(string), source(string), status(string), userId(string)
  api_v1_users_user_id_risk:
    fields: _links(object), riskLevel(string)
  api_v1_users_user_id_roles:
    primary key: id
    fields: _embedded(object), _links(object), assignmentType(string), created(string), id(string), label(string), lastUpdated(string), status(string), type(string)
  api_v1_users_user_id_roles_role_assignment_id:
    primary key: id
    fields: _embedded(object), _links(object), assignmentType(string), created(string), id(string), label(string), lastUpdated(string), status(string), type(string)
  api_v1_users_user_id_roles_role_assignment_id_governance:
    fields: _links(object), grants(array)
  api_v1_users_user_id_roles_role_assignment_id_governance_grant_id:
    fields: _links(object), bundleId(string), expirationDate(string), grantId(string), type(string)
  api_v1_users_user_id_roles_role_assignment_id_governance_grant_id_resources:
    fields: _links(object), resources(array)
  api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps:
    primary key: id
    fields: _links(object), category(string), description(string), displayName(string), features(array), id(string), lastUpdated(string), name(string), signOnModes(array), status(string), verificationStatus(string), website(string)
  api_v1_users_user_id_roles_role_assignment_id_targets_groups:
    primary key: id
    fields: _embedded(object), _links(object), created(string), id(string), lastMembershipUpdated(string), lastUpdated(string), objectClass(array), profile(object), type(string)
  api_v1_users_user_id_roles_role_id_or_encoded_role_id_targets:
    fields: _links(object), assignmentType(string), expiration(string), orn(string)
  api_v1_users_user_id_subscriptions:
    fields: _links(object), channels(array), notificationType(string), status(string)
  api_v1_users_user_id_subscriptions_notification_type:
    fields: _links(object), channels(array), notificationType(string), status(string)
  api_v1_zones:
    primary key: id
    fields: _links(object), created(string), id(string), lastUpdated(string), name(string), status(string), system(boolean), type(string), usage(string)
  api_v1_zones_zone_id:
    primary key: id
    fields: _links(object), created(string), id(string), lastUpdated(string), name(string), status(string), system(boolean), type(string), usage(string)
  attack_protection_api_v1_authenticator_settings:
    fields: verifyKnowledgeSecondWhen2faRequired(boolean)
  attack_protection_api_v1_user_lockout_settings:
    fields: preventBruteForceLockoutFromUnknownDevices(boolean)
  integrations_api_v1_api_services:
    primary key: id
    fields: _links(object), configGuideUrl(string), createdAt(string), createdBy(string), grantedScopes(array), id(string), name(string), properties(object), type(string)
  integrations_api_v1_api_services_api_service_id:
    primary key: id
    fields: _links(object), configGuideUrl(string), createdAt(string), createdBy(string), grantedScopes(array), id(string), name(string), properties(object), type(string)
  integrations_api_v1_api_services_api_service_id_credentials_secrets:
    primary key: id
    fields: _links(object), client_secret(string), created(string), id(string), lastUpdated(string), secret_hash(string), status(string)
  oauth2_v1_clients_client_id_roles:
    primary key: id
    fields: _embedded(object), _links(object), assignmentType(string), created(string), id(string), label(string), lastUpdated(string), status(string), type(string)
  oauth2_v1_clients_client_id_roles_role_assignment_id:
    primary key: id
    fields: _embedded(object), _links(object), assignmentType(string), created(string), id(string), label(string), lastUpdated(string), status(string), type(string)
  oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps:
    primary key: id
    fields: _links(object), category(string), description(string), displayName(string), features(array), id(string), lastUpdated(string), name(string), signOnModes(array), status(string), verificationStatus(string), website(string)
  oauth2_v1_clients_client_id_roles_role_assignment_id_targets_groups:
    primary key: id
    fields: _embedded(object), _links(object), created(string), id(string), lastMembershipUpdated(string), lastUpdated(string), objectClass(array), profile(object), type(string)
  okta_personal_settings_api_v1_export_blocklists:
    fields: domains(array)
  privileged_access_api_v1_okta_service_accounts:
    primary key: id
    fields: created(string), description(string), email(string), id(string), lastUpdated(string), name(string), oktaUserId(string), ownerGroupIds(array), ownerUserIds(array), status(string), statusDetail(string), username(string)
  privileged_access_api_v1_okta_service_accounts_id:
    primary key: id
    fields: created(string), description(string), email(string), id(string), lastUpdated(string), name(string), oktaUserId(string), ownerGroupIds(array), ownerUserIds(array), status(string), statusDetail(string), username(string)
  privileged_access_api_v1_service_accounts:
    primary key: id
    fields: containerGlobalName(string), containerInstanceName(string), containerOrn(string), created(string), description(string), id(string), lastUpdated(string), name(string), ownerGroupIds(array), ownerUserIds(array), password(string), status(string), statusDetail(string), username(string)
  privileged_access_api_v1_service_accounts_id:
    primary key: id
    fields: containerGlobalName(string), containerInstanceName(string), containerOrn(string), created(string), description(string), id(string), lastUpdated(string), name(string), ownerGroupIds(array), ownerUserIds(array), password(string), status(string), statusDetail(string), username(string)
  webauthn_registration_api_v1_users_user_id_enrollments:
    primary key: id
    fields: _links(object), created(string), factorType(string), id(string), lastUpdated(string), profile(object), provider(string), status(string), vendorName(string)

SYNC MODES
  ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped, incremental_append, incremental_append_deduped

REVERSE ETL ACTIONS
  create_api_v1_agent_pools_pool_id_updates:
    endpoint: POST /api/v1/agentPools/{{ record.pool_id }}/updates
    required fields: pool_id
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_agent_pools_pool_id_updates_settings:
    endpoint: POST /api/v1/agentPools/{{ record.pool_id }}/updates/settings
    required fields: pool_id, agentType
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
    required fields: signOnMode, label
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_apps_app_id:
    endpoint: PUT /api/v1/apps/{{ record.app_id }}
    required fields: app_id, signOnMode, label
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_apps_app_id:
    endpoint: DELETE /api/v1/apps/{{ record.app_id }}
    required fields: app_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_apps_app_id_connections_default:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/connections/default
    required fields: app_id, profile
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
    required fields: app_id, connection_id, status
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
    required fields: app_id, issuer, scopeId
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_apps_app_id_grants_grant_id:
    endpoint: DELETE /api/v1/apps/{{ record.app_id }}/grants/{{ record.grant_id }}
    required fields: app_id, grant_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_apps_app_id_group_push_mappings:
    endpoint: POST /api/v1/apps/{{ record.app_id }}/group-push/mappings
    required fields: app_id, sourceGroupId
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_apps_app_id_group_push_mappings_mapping_id:
    endpoint: PATCH /api/v1/apps/{{ record.app_id }}/group-push/mappings/{{ record.mapping_id }}
    required fields: app_id, mapping_id, status
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
    required fields: app_id, id
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
    required fields: auth_server_id, policy_id, name, conditions, type
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id:
    endpoint: PUT /api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id }}/rules/{{ record.rule_id }}
    required fields: auth_server_id, policy_id, rule_id, name, conditions, type
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
    required fields: auth_server_id, name
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_authorization_servers_auth_server_id_scopes_scope_id:
    endpoint: PUT /api/v1/authorizationServers/{{ record.auth_server_id }}/scopes/{{ record.scope_id }}
    required fields: auth_server_id, scope_id, name
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_authorization_servers_auth_server_id_scopes_scope_id:
    endpoint: DELETE /api/v1/authorizationServers/{{ record.auth_server_id }}/scopes/{{ record.scope_id }}
    required fields: auth_server_id, scope_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_behaviors:
    endpoint: POST /api/v1/behaviors
    required fields: name, type
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_behaviors_behavior_id:
    endpoint: PUT /api/v1/behaviors/{{ record.behavior_id }}
    required fields: behavior_id, name, type
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
    required fields: level, mode
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_brands:
    endpoint: POST /api/v1/brands
    required fields: name
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_brands_brand_id:
    endpoint: PUT /api/v1/brands/{{ record.brand_id }}
    required fields: brand_id, name
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
    required fields: brand_id, type
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_brands_brand_id_templates_email_template_name_customizations:
    endpoint: POST /api/v1/brands/{{ record.brand_id }}/templates/email/{{ record.template_name }}/customizations
    required fields: brand_id, template_name, language
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_brands_brand_id_templates_email_template_name_customizations:
    endpoint: DELETE /api/v1/brands/{{ record.brand_id }}/templates/email/{{ record.template_name }}/customizations
    required fields: brand_id, template_name
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id:
    endpoint: PUT /api/v1/brands/{{ record.brand_id }}/templates/email/{{ record.template_name }}/customizations/{{ record.customization_id }}
    required fields: brand_id, template_name, customization_id, language
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id:
    endpoint: DELETE /api/v1/brands/{{ record.brand_id }}/templates/email/{{ record.template_name }}/customizations/{{ record.customization_id }}
    required fields: brand_id, template_name, customization_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_brands_brand_id_templates_email_template_name_settings:
    endpoint: PUT /api/v1/brands/{{ record.brand_id }}/templates/email/{{ record.template_name }}/settings
    required fields: brand_id, template_name, recipients
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_brands_brand_id_templates_email_template_name_test:
    endpoint: POST /api/v1/brands/{{ record.brand_id }}/templates/email/{{ record.template_name }}/test
    required fields: brand_id, template_name
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_brands_brand_id_themes_theme_id:
    endpoint: PUT /api/v1/brands/{{ record.brand_id }}/themes/{{ record.theme_id }}
    required fields: brand_id, theme_id, primaryColorHex, secondaryColorHex, signInPageTouchPointVariant, endUserDashboardTouchPointVariant, errorPageTouchPointVariant, emailTemplateTouchPointVariant
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
    required fields: brand_id, path, representation
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
    required fields: app_instance_id, id, parameters
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_directories_app_instance_id_groups_group_id_query:
    endpoint: POST /api/v1/directories/{{ record.app_instance_id }}/groups/{{ record.group_id }}/query
    required fields: app_instance_id, group_id, attributes
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_domains:
    endpoint: POST /api/v1/domains
    required fields: certificateSourceType, domain
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_domains_domain_id:
    endpoint: PUT /api/v1/domains/{{ record.domain_id }}
    required fields: domain_id, brandId
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_domains_domain_id:
    endpoint: DELETE /api/v1/domains/{{ record.domain_id }}
    required fields: domain_id
    risk: high: external Okta admin API mutation; approval required
  update_api_v1_domains_domain_id_certificate:
    endpoint: PUT /api/v1/domains/{{ record.domain_id }}/certificate
    required fields: domain_id, certificate, certificateChain, privateKey, type
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
    required fields: displayName, userName
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_email_domains_email_domain_id:
    endpoint: PUT /api/v1/email-domains/{{ record.email_domain_id }}
    required fields: email_domain_id, displayName, userName
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
    required fields: alias, enabled, host, port, username, authType
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_email_servers_email_server_id:
    endpoint: PATCH /api/v1/email-servers/{{ record.email_server_id }}
    required fields: email_server_id, authType
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_email_servers_email_server_id:
    endpoint: DELETE /api/v1/email-servers/{{ record.email_server_id }}
    required fields: email_server_id
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_email_servers_email_server_id_test:
    endpoint: POST /api/v1/email-servers/{{ record.email_server_id }}/test
    required fields: email_server_id, fromAddress, toAddress
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_event_hooks:
    endpoint: POST /api/v1/eventHooks
    required fields: name, events, channel
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_event_hooks_event_hook_id:
    endpoint: PUT /api/v1/eventHooks/{{ record.event_hook_id }}
    required fields: event_hook_id, name, events, channel
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
    required fields: group_id, type
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
    required fields: description, label, resources
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
    required fields: resource_set_id_or_label, resourceOrnOrUrl, conditions
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
    required fields: label, description, permissions
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_iam_roles_role_id_or_label:
    endpoint: PUT /api/v1/iam/roles/{{ record.role_id_or_label }}
    required fields: role_id_or_label, label, description
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
    required fields: x5c
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
    required fields: created, id, lastUpdated, name, status, type, _links
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_log_streams_log_stream_id:
    endpoint: PUT /api/v1/logStreams/{{ record.log_stream_id }}
    required fields: log_stream_id, name, type
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
    required fields: mapping_id, properties
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
    required fields: name, displayName
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_meta_types_user_type_id:
    endpoint: POST /api/v1/meta/types/user/{{ record.type_id }}
    required fields: type_id
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_meta_types_user_type_id:
    endpoint: PUT /api/v1/meta/types/user/{{ record.type_id }}
    required fields: type_id, name, displayName, description
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
    required fields: accountId
    risk: medium: external Okta admin API mutation; approval required
  execute_api_v1_org_privacy_aerial_revoke:
    endpoint: POST /api/v1/org/privacy/aerial/revoke
    required fields: accountId
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
    required fields: admin, edition, name, subdomain
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_policies:
    endpoint: POST /api/v1/policies
    required fields: name, type
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_policies_policy_id:
    endpoint: PUT /api/v1/policies/{{ record.policy_id }}
    required fields: policy_id, name, type
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
    required fields: principalId, principalType
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_principal_rate_limits_principal_rate_limit_id:
    endpoint: PUT /api/v1/principal-rate-limits/{{ record.principal_rate_limit_id }}
    required fields: principal_rate_limit_id, principalId, principalType
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
    required fields: notificationsEnabled
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_rate_limit_settings_per_client:
    endpoint: PUT /api/v1/rate-limit-settings/per-client
    required fields: defaultMode
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_rate_limit_settings_warning_threshold:
    endpoint: PUT /api/v1/rate-limit-settings/warning-threshold
    required fields: warningThreshold
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
    required fields: name, settings, type
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_security_events_providers_security_event_provider_id:
    endpoint: PUT /api/v1/security-events-providers/{{ record.security_event_provider_id }}
    required fields: security_event_provider_id, name, settings, type
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
    required fields: events_requested, delivery
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_ssf_stream:
    endpoint: PUT /api/v1/ssf/stream
    required fields: events_requested, delivery
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_ssf_stream_2:
    endpoint: PATCH /api/v1/ssf/stream
    required fields: events_requested, delivery
    risk: medium: external Okta admin API mutation; approval required
  delete_api_v1_ssf_stream:
    endpoint: DELETE /api/v1/ssf/stream
    risk: high: external Okta admin API mutation; approval required
  create_api_v1_ssf_stream_verification:
    endpoint: POST /api/v1/ssf/stream/verification
    required fields: stream_id
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
    required fields: action
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
    required fields: profile
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
    required fields: user_id, riskLevel
    risk: medium: external Okta admin API mutation; approval required
  create_api_v1_users_user_id_roles:
    endpoint: POST /api/v1/users/{{ record.user_id }}/roles
    required fields: user_id, type
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
    required fields: name, type
    risk: medium: external Okta admin API mutation; approval required
  update_api_v1_zones_zone_id:
    endpoint: PUT /api/v1/zones/{{ record.zone_id }}
    required fields: zone_id, name, type
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
    required fields: type, grantedScopes
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
    required fields: client_id, type
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
    required fields: name, oktaUserId
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
    required fields: name, containerOrn, username, password
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

COMMAND SURFACE
  Run Okta's declared streams and reverse-ETL actions.
  Usage: pm okta <command> [flags]
  Read streams
  Reverse ETL writes
  Other Commands
    api delete api v1 apps appid group-push mappings mappingid - Documented DELETE /api/v1/apps/{appId}/group-push/mappings/{mappingId} (not implemented) [intent=direct_write availability=not_implemented operation=okta.delete.api-v1-apps-appid-group-push-mappings-mappingid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api get api v1 apps appid interclient-allowed-apps - Documented GET /api/v1/apps/{appId}/interclient-allowed-apps (not implemented) [intent=direct_read availability=not_implemented operation=okta.get.api-v1-apps-appid-interclient-allowed-apps]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 apps appid interclient-target-apps - Documented GET /api/v1/apps/{appId}/interclient-target-apps (not implemented) [intent=direct_read availability=not_implemented operation=okta.get.api-v1-apps-appid-interclient-target-apps]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 apps appid sso saml metadata - Documented GET /api/v1/apps/{appId}/sso/saml/metadata (not implemented) [intent=direct_read availability=not_implemented operation=okta.get.api-v1-apps-appid-sso-saml-metadata]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 brands brandid pages sign-in widget-versions - Documented GET /api/v1/brands/{brandId}/pages/sign-in/widget-versions (not implemented) [intent=direct_read availability=not_implemented operation=okta.get.api-v1-brands-brandid-pages-sign-in-widget-versions]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: medium; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 devices deviceid os-accounts - Documented GET /api/v1/devices/{deviceId}/os-accounts (not implemented) [intent=direct_read availability=not_implemented operation=okta.get.api-v1-devices-deviceid-os-accounts]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api get api v1 devices deviceid os-accounts osaccountid - Documented GET /api/v1/devices/{deviceId}/os-accounts/{osAccountId} (not implemented) [intent=direct_read availability=not_implemented operation=okta.get.api-v1-devices-deviceid-os-accounts-osaccountid]; approval: not implemented: the direct-read executor has no non-redacting output policy for this provider operation; risk: low; notes: named_dependency=engine.direct_read_executor: the direct-read executor has no non-redacting output policy for this provider operation; flags: --page, --page-cursor
    api patch api v1 apps appid groups groupid - Documented PATCH /api/v1/apps/{appId}/groups/{groupId} (not implemented) [intent=direct_write availability=not_implemented operation=okta.patch.api-v1-apps-appid-groups-groupid]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 apps appid credentials csrs csrid lifecycle publish - Documented POST /api/v1/apps/{appId}/credentials/csrs/{csrId}/lifecycle/publish (not implemented) [intent=direct_write availability=not_implemented operation=okta.post.api-v1-apps-appid-credentials-csrs-csrid-lifecycle-publish]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 apps appid credentials keys generate - Documented POST /api/v1/apps/{appId}/credentials/keys/generate (not implemented) [intent=direct_write availability=not_implemented operation=okta.post.api-v1-apps-appid-credentials-keys-generate]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 apps appid credentials keys keyid clone - Documented POST /api/v1/apps/{appId}/credentials/keys/{keyId}/clone (not implemented) [intent=direct_write availability=not_implemented operation=okta.post.api-v1-apps-appid-credentials-keys-keyid-clone]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 apps appid logo - Documented POST /api/v1/apps/{appId}/logo (not implemented) [intent=direct_write availability=not_implemented operation=okta.post.api-v1-apps-appid-logo]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 brands brandid themes themeid background-image - Documented POST /api/v1/brands/{brandId}/themes/{themeId}/background-image (not implemented) [intent=direct_write availability=not_implemented operation=okta.post.api-v1-brands-brandid-themes-themeid-background-image]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 brands brandid themes themeid favicon - Documented POST /api/v1/brands/{brandId}/themes/{themeId}/favicon (not implemented) [intent=direct_write availability=not_implemented operation=okta.post.api-v1-brands-brandid-themes-themeid-favicon]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 brands brandid themes themeid logo - Documented POST /api/v1/brands/{brandId}/themes/{themeId}/logo (not implemented) [intent=direct_write availability=not_implemented operation=okta.post.api-v1-brands-brandid-themes-themeid-logo]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 idps idpid credentials csrs idpcsrid lifecycle publish - Documented POST /api/v1/idps/{idpId}/credentials/csrs/{idpCsrId}/lifecycle/publish (not implemented) [intent=direct_write availability=not_implemented operation=okta.post.api-v1-idps-idpid-credentials-csrs-idpcsrid-lifecycle-publish]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 idps idpid credentials keys generate - Documented POST /api/v1/idps/{idpId}/credentials/keys/generate (not implemented) [intent=direct_write availability=not_implemented operation=okta.post.api-v1-idps-idpid-credentials-keys-generate]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 idps idpid credentials keys kid clone - Documented POST /api/v1/idps/{idpId}/credentials/keys/{kid}/clone (not implemented) [intent=direct_write availability=not_implemented operation=okta.post.api-v1-idps-idpid-credentials-keys-kid-clone]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 policies simulate - Documented POST /api/v1/policies/simulate (not implemented) [intent=direct_write availability=not_implemented operation=okta.post.api-v1-policies-simulate]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post api v1 users id lifecycle reset-password - Documented POST /api/v1/users/{id}/lifecycle/reset_password (not implemented) [intent=direct_write availability=not_implemented operation=okta.post.api-v1-users-id-lifecycle-reset-password]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api post security api v1 security-events - Documented POST /security/api/v1/security-events (not implemented) [intent=direct_write availability=not_implemented operation=okta.post.security-api-v1-security-events]; approval: not implemented: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract; risk: high; notes: named_dependency=engine.rest_write_operation_contract: the REST write executor lacks a reviewed operation-specific body, risk, approval, and execution contract
    api v1 agent pools list - Run the api v1 agent pools ETL stream [intent=etl availability=implemented stream=api_v1_agent_pools]
    api v1 agent pools pool id updates list - Run the api v1 agent pools pool id updates ETL stream [intent=etl availability=implemented stream=api_v1_agent_pools_pool_id_updates]
    api v1 agent pools pool id updates settings list - Run the api v1 agent pools pool id updates settings ETL stream [intent=etl availability=implemented stream=api_v1_agent_pools_pool_id_updates_settings]
    api v1 agent pools pool id updates update id list - Run the api v1 agent pools pool id updates update id ETL stream [intent=etl availability=implemented stream=api_v1_agent_pools_pool_id_updates_update_id]
    api v1 api tokens api token id list - Run the api v1 api tokens api token id ETL stream [intent=etl availability=implemented stream=api_v1_api_tokens_api_token_id]
    api v1 api tokens list - Run the api v1 api tokens ETL stream [intent=etl availability=implemented stream=api_v1_api_tokens]
    api v1 apps app id connections default jwks list - Run the api v1 apps app id connections default jwks ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_connections_default_jwks]
    api v1 apps app id connections default list - Run the api v1 apps app id connections default ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_connections_default]
    api v1 apps app id credentials csrs csr id list - Run the api v1 apps app id credentials csrs csr id ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_credentials_csrs_csr_id]
    api v1 apps app id credentials csrs list - Run the api v1 apps app id credentials csrs ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_credentials_csrs]
    api v1 apps app id credentials jwks key id list - Run the api v1 apps app id credentials jwks key id ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_credentials_jwks_key_id]
    api v1 apps app id credentials jwks list - Run the api v1 apps app id credentials jwks ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_credentials_jwks]
    api v1 apps app id credentials keys key id list - Run the api v1 apps app id credentials keys key id ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_credentials_keys_key_id]
    api v1 apps app id credentials keys list - Run the api v1 apps app id credentials keys ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_credentials_keys]
    api v1 apps app id credentials secrets list - Run the api v1 apps app id credentials secrets ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_credentials_secrets]
    api v1 apps app id credentials secrets secret id list - Run the api v1 apps app id credentials secrets secret id ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_credentials_secrets_secret_id]
    api v1 apps app id cwo connections connection id list - Run the api v1 apps app id cwo connections connection id ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_cwo_connections_connection_id]
    api v1 apps app id cwo connections list - Run the api v1 apps app id cwo connections ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_cwo_connections]
    api v1 apps app id features feature name list - Run the api v1 apps app id features feature name ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_features_feature_name]
    api v1 apps app id features list - Run the api v1 apps app id features ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_features]
    api v1 apps app id federated claims claim id list - Run the api v1 apps app id federated claims claim id ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_federated_claims_claim_id]
    api v1 apps app id federated claims list - Run the api v1 apps app id federated claims ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_federated_claims]
    api v1 apps app id grants grant id list - Run the api v1 apps app id grants grant id ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_grants_grant_id]
    api v1 apps app id grants list - Run the api v1 apps app id grants ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_grants]
    api v1 apps app id group push mappings list - Run the api v1 apps app id group push mappings ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_group_push_mappings]
    api v1 apps app id group push mappings mapping id list - Run the api v1 apps app id group push mappings mapping id ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_group_push_mappings_mapping_id]
    api v1 apps app id groups group id list - Run the api v1 apps app id groups group id ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_groups_group_id]
    api v1 apps app id groups list - Run the api v1 apps app id groups ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_groups]
    api v1 apps app id list - Run the api v1 apps app id ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id]
    api v1 apps app id tokens list - Run the api v1 apps app id tokens ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_tokens]
    api v1 apps app id tokens token id list - Run the api v1 apps app id tokens token id ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_tokens_token_id]
    api v1 apps app id users list - Run the api v1 apps app id users ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_users]
    api v1 apps app id users user id list - Run the api v1 apps app id users user id ETL stream [intent=etl availability=implemented stream=api_v1_apps_app_id_users_user_id]
    api v1 apps list - Run the api v1 apps ETL stream [intent=etl availability=implemented stream=api_v1_apps]
    api v1 authenticators authenticator id aaguids aaguid list - Run the api v1 authenticators authenticator id aaguids aaguid ETL stream [intent=etl availability=implemented stream=api_v1_authenticators_authenticator_id_aaguids_aaguid]
    api v1 authenticators authenticator id aaguids list - Run the api v1 authenticators authenticator id aaguids ETL stream [intent=etl availability=implemented stream=api_v1_authenticators_authenticator_id_aaguids]
    api v1 authenticators authenticator id list - Run the api v1 authenticators authenticator id ETL stream [intent=etl availability=implemented stream=api_v1_authenticators_authenticator_id]
    api v1 authenticators authenticator id methods list - Run the api v1 authenticators authenticator id methods ETL stream [intent=etl availability=implemented stream=api_v1_authenticators_authenticator_id_methods]
    api v1 authenticators authenticator id methods method type list - Run the api v1 authenticators authenticator id methods method type ETL stream [intent=etl availability=implemented stream=api_v1_authenticators_authenticator_id_methods_method_type]
    api v1 authenticators list - Run the api v1 authenticators ETL stream [intent=etl availability=implemented stream=api_v1_authenticators]
    api v1 authorization servers auth server id associated servers list - Run the api v1 authorization servers auth server id associated servers ETL stream [intent=etl availability=implemented stream=api_v1_authorization_servers_auth_server_id_associated_servers]
    api v1 authorization servers auth server id claims claim id list - Run the api v1 authorization servers auth server id claims claim id ETL stream [intent=etl availability=implemented stream=api_v1_authorization_servers_auth_server_id_claims_claim_id]
    api v1 authorization servers auth server id claims list - Run the api v1 authorization servers auth server id claims ETL stream [intent=etl availability=implemented stream=api_v1_authorization_servers_auth_server_id_claims]
    api v1 authorization servers auth server id clients client id tokens list - Run the api v1 authorization servers auth server id clients client id tokens ETL stream [intent=etl availability=implemented stream=api_v1_authorization_servers_auth_server_id_clients_client_id_tokens]
    api v1 authorization servers auth server id clients client id tokens token id list - Run the api v1 authorization servers auth server id clients client id tokens token id ETL stream [intent=etl availability=implemented stream=api_v1_authorization_servers_auth_server_id_clients_client_id_tokens_token_id]
    api v1 authorization servers auth server id clients list - Run the api v1 authorization servers auth server id clients ETL stream [intent=etl availability=implemented stream=api_v1_authorization_servers_auth_server_id_clients]
    api v1 authorization servers auth server id credentials keys key id list - Run the api v1 authorization servers auth server id credentials keys key id ETL stream [intent=etl availability=implemented stream=api_v1_authorization_servers_auth_server_id_credentials_keys_key_id]
    api v1 authorization servers auth server id credentials keys list - Run the api v1 authorization servers auth server id credentials keys ETL stream [intent=etl availability=implemented stream=api_v1_authorization_servers_auth_server_id_credentials_keys]
    api v1 authorization servers auth server id list - Run the api v1 authorization servers auth server id ETL stream [intent=etl availability=implemented stream=api_v1_authorization_servers_auth_server_id]
    api v1 authorization servers auth server id policies list - Run the api v1 authorization servers auth server id policies ETL stream [intent=etl availability=implemented stream=api_v1_authorization_servers_auth_server_id_policies]
    api v1 authorization servers auth server id policies policy id list - Run the api v1 authorization servers auth server id policies policy id ETL stream [intent=etl availability=implemented stream=api_v1_authorization_servers_auth_server_id_policies_policy_id]
    api v1 authorization servers auth server id policies policy id rules list - Run the api v1 authorization servers auth server id policies policy id rules ETL stream [intent=etl availability=implemented stream=api_v1_authorization_servers_auth_server_id_policies_policy_id_rules]
    api v1 authorization servers auth server id policies policy id rules rule id list - Run the api v1 authorization servers auth server id policies policy id rules rule id ETL stream [intent=etl availability=implemented stream=api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id]
    api v1 authorization servers auth server id resourceservercredentials keys key id list - Run the api v1 authorization servers auth server id resourceservercredentials keys key id ETL stream [intent=etl availability=implemented stream=api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id]
    api v1 authorization servers auth server id resourceservercredentials keys list - Run the api v1 authorization servers auth server id resourceservercredentials keys ETL stream [intent=etl availability=implemented stream=api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys]
    api v1 authorization servers auth server id scopes list - Run the api v1 authorization servers auth server id scopes ETL stream [intent=etl availability=implemented stream=api_v1_authorization_servers_auth_server_id_scopes]
    api v1 authorization servers auth server id scopes scope id list - Run the api v1 authorization servers auth server id scopes scope id ETL stream [intent=etl availability=implemented stream=api_v1_authorization_servers_auth_server_id_scopes_scope_id]
    api v1 authorization servers list - Run the api v1 authorization servers ETL stream [intent=etl availability=implemented stream=api_v1_authorization_servers]
    api v1 behaviors behavior id list - Run the api v1 behaviors behavior id ETL stream [intent=etl availability=implemented stream=api_v1_behaviors_behavior_id]
    api v1 behaviors list - Run the api v1 behaviors ETL stream [intent=etl availability=implemented stream=api_v1_behaviors]
    api v1 bot protection configuration list - Run the api v1 bot protection configuration ETL stream [intent=etl availability=implemented stream=api_v1_bot_protection_configuration]
    api v1 brands brand id domains list - Run the api v1 brands brand id domains ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_domains]
    api v1 brands brand id list - Run the api v1 brands brand id ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id]
    api v1 brands brand id pages error customized list - Run the api v1 brands brand id pages error customized ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_pages_error_customized]
    api v1 brands brand id pages error default list - Run the api v1 brands brand id pages error default ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_pages_error_default]
    api v1 brands brand id pages error list - Run the api v1 brands brand id pages error ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_pages_error]
    api v1 brands brand id pages error preview list - Run the api v1 brands brand id pages error preview ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_pages_error_preview]
    api v1 brands brand id pages sign in customized list - Run the api v1 brands brand id pages sign in customized ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_pages_sign_in_customized]
    api v1 brands brand id pages sign in default list - Run the api v1 brands brand id pages sign in default ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_pages_sign_in_default]
    api v1 brands brand id pages sign in list - Run the api v1 brands brand id pages sign in ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_pages_sign_in]
    api v1 brands brand id pages sign in preview list - Run the api v1 brands brand id pages sign in preview ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_pages_sign_in_preview]
    api v1 brands brand id pages sign out customized list - Run the api v1 brands brand id pages sign out customized ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_pages_sign_out_customized]
    api v1 brands brand id templates email list - Run the api v1 brands brand id templates email ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_templates_email]
    api v1 brands brand id templates email template name customizations customization id list - Run the api v1 brands brand id templates email template name customizations customization id ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id]
    api v1 brands brand id templates email template name customizations customization id preview list - Run the api v1 brands brand id templates email template name customizations customization id preview ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id_preview]
    api v1 brands brand id templates email template name customizations list - Run the api v1 brands brand id templates email template name customizations ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_templates_email_template_name_customizations]
    api v1 brands brand id templates email template name default content list - Run the api v1 brands brand id templates email template name default content ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_templates_email_template_name_default_content]
    api v1 brands brand id templates email template name default content preview list - Run the api v1 brands brand id templates email template name default content preview ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_templates_email_template_name_default_content_preview]
    api v1 brands brand id templates email template name list - Run the api v1 brands brand id templates email template name ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_templates_email_template_name]
    api v1 brands brand id templates email template name settings list - Run the api v1 brands brand id templates email template name settings ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_templates_email_template_name_settings]
    api v1 brands brand id themes list - Run the api v1 brands brand id themes ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_themes]
    api v1 brands brand id themes theme id list - Run the api v1 brands brand id themes theme id ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_themes_theme_id]
    api v1 brands brand id well known uris list - Run the api v1 brands brand id well known uris ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_well_known_uris]
    api v1 brands brand id well known uris path customized list - Run the api v1 brands brand id well known uris path customized ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_well_known_uris_path_customized]
    api v1 brands brand id well known uris path list - Run the api v1 brands brand id well known uris path ETL stream [intent=etl availability=implemented stream=api_v1_brands_brand_id_well_known_uris_path]
    api v1 brands list - Run the api v1 brands ETL stream [intent=etl availability=implemented stream=api_v1_brands]
    api v1 captchas captcha id list - Run the api v1 captchas captcha id ETL stream [intent=etl availability=implemented stream=api_v1_captchas_captcha_id]
    api v1 captchas list - Run the api v1 captchas ETL stream [intent=etl availability=implemented stream=api_v1_captchas]
    api v1 device assurances device assurance id list - Run the api v1 device assurances device assurance id ETL stream [intent=etl availability=implemented stream=api_v1_device_assurances_device_assurance_id]
    api v1 device assurances list - Run the api v1 device assurances ETL stream [intent=etl availability=implemented stream=api_v1_device_assurances]
    api v1 device integrations device integration id list - Run the api v1 device integrations device integration id ETL stream [intent=etl availability=implemented stream=api_v1_device_integrations_device_integration_id]
    api v1 device integrations list - Run the api v1 device integrations ETL stream [intent=etl availability=implemented stream=api_v1_device_integrations]
    api v1 device posture checks default list - Run the api v1 device posture checks default ETL stream [intent=etl availability=implemented stream=api_v1_device_posture_checks_default]
    api v1 device posture checks list - Run the api v1 device posture checks ETL stream [intent=etl availability=implemented stream=api_v1_device_posture_checks]
    api v1 device posture checks posture check id list - Run the api v1 device posture checks posture check id ETL stream [intent=etl availability=implemented stream=api_v1_device_posture_checks_posture_check_id]
    api v1 devices device id list - Run the api v1 devices device id ETL stream [intent=etl availability=implemented stream=api_v1_devices_device_id]
    api v1 devices device id users list - Run the api v1 devices device id users ETL stream [intent=etl availability=implemented stream=api_v1_devices_device_id_users]
    api v1 devices list - Run the api v1 devices ETL stream [intent=etl availability=implemented stream=api_v1_devices]
    api v1 directories app instance id groups group id query result id list - Run the api v1 directories app instance id groups group id query result id ETL stream [intent=etl availability=implemented stream=api_v1_directories_app_instance_id_groups_group_id_query_result_id]
    api v1 domains domain id list - Run the api v1 domains domain id ETL stream [intent=etl availability=implemented stream=api_v1_domains_domain_id]
    api v1 domains list - Run the api v1 domains ETL stream [intent=etl availability=implemented stream=api_v1_domains]
    api v1 dr status domain list - Run the api v1 dr status domain ETL stream [intent=etl availability=implemented stream=api_v1_dr_status_domain]
    api v1 dr status list - Run the api v1 dr status ETL stream [intent=etl availability=implemented stream=api_v1_dr_status]
    api v1 email domains email domain id list - Run the api v1 email domains email domain id ETL stream [intent=etl availability=implemented stream=api_v1_email_domains_email_domain_id]
    api v1 email domains list - Run the api v1 email domains ETL stream [intent=etl availability=implemented stream=api_v1_email_domains]
    api v1 email servers email server id list - Run the api v1 email servers email server id ETL stream [intent=etl availability=implemented stream=api_v1_email_servers_email_server_id]
    api v1 email servers list - Run the api v1 email servers ETL stream [intent=etl availability=implemented stream=api_v1_email_servers]
    api v1 event hooks event hook id list - Run the api v1 event hooks event hook id ETL stream [intent=etl availability=implemented stream=api_v1_event_hooks_event_hook_id]
    api v1 event hooks list - Run the api v1 event hooks ETL stream [intent=etl availability=implemented stream=api_v1_event_hooks]
    api v1 features feature id dependencies list - Run the api v1 features feature id dependencies ETL stream [intent=etl availability=implemented stream=api_v1_features_feature_id_dependencies]
    api v1 features feature id dependents list - Run the api v1 features feature id dependents ETL stream [intent=etl availability=implemented stream=api_v1_features_feature_id_dependents]
    api v1 features feature id list - Run the api v1 features feature id ETL stream [intent=etl availability=implemented stream=api_v1_features_feature_id]
    api v1 features list - Run the api v1 features ETL stream [intent=etl availability=implemented stream=api_v1_features]
    api v1 first party app settings app name list - Run the api v1 first party app settings app name ETL stream [intent=etl availability=implemented stream=api_v1_first_party_app_settings_app_name]
    api v1 groups group id apps list - Run the api v1 groups group id apps ETL stream [intent=etl availability=implemented stream=api_v1_groups_group_id_apps]
    api v1 groups group id list - Run the api v1 groups group id ETL stream [intent=etl availability=implemented stream=api_v1_groups_group_id]
    api v1 groups group id owners list - Run the api v1 groups group id owners ETL stream [intent=etl availability=implemented stream=api_v1_groups_group_id_owners]
    api v1 groups group id roles list - Run the api v1 groups group id roles ETL stream [intent=etl availability=implemented stream=api_v1_groups_group_id_roles]
    api v1 groups group id roles role assignment id list - Run the api v1 groups group id roles role assignment id ETL stream [intent=etl availability=implemented stream=api_v1_groups_group_id_roles_role_assignment_id]
    api v1 groups group id roles role assignment id targets catalog apps list - Run the api v1 groups group id roles role assignment id targets catalog apps ETL stream [intent=etl availability=implemented stream=api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps]
    api v1 groups group id roles role assignment id targets groups list - Run the api v1 groups group id roles role assignment id targets groups ETL stream [intent=etl availability=implemented stream=api_v1_groups_group_id_roles_role_assignment_id_targets_groups]
    api v1 groups group id users list - Run the api v1 groups group id users ETL stream [intent=etl availability=implemented stream=api_v1_groups_group_id_users]
    api v1 groups rules group rule id list - Run the api v1 groups rules group rule id ETL stream [intent=etl availability=implemented stream=api_v1_groups_rules_group_rule_id]
    api v1 groups rules list - Run the api v1 groups rules ETL stream [intent=etl availability=implemented stream=api_v1_groups_rules]
    api v1 hook keys id list - Run the api v1 hook keys id ETL stream [intent=etl availability=implemented stream=api_v1_hook_keys_id]
    api v1 hook keys list - Run the api v1 hook keys ETL stream [intent=etl availability=implemented stream=api_v1_hook_keys]
    api v1 hook keys public key id list - Run the api v1 hook keys public key id ETL stream [intent=etl availability=implemented stream=api_v1_hook_keys_public_key_id]
    api v1 iam assignees users list - Run the api v1 iam assignees users ETL stream [intent=etl availability=implemented stream=api_v1_iam_assignees_users]
    api v1 iam governance bundles bundle id entitlements entitlement id values list - Run the api v1 iam governance bundles bundle id entitlements entitlement id values ETL stream [intent=etl availability=implemented stream=api_v1_iam_governance_bundles_bundle_id_entitlements_entitlement_id_values]
    api v1 iam governance bundles bundle id entitlements list - Run the api v1 iam governance bundles bundle id entitlements ETL stream [intent=etl availability=implemented stream=api_v1_iam_governance_bundles_bundle_id_entitlements]
    api v1 iam governance bundles bundle id list - Run the api v1 iam governance bundles bundle id ETL stream [intent=etl availability=implemented stream=api_v1_iam_governance_bundles_bundle_id]
    api v1 iam governance bundles list - Run the api v1 iam governance bundles ETL stream [intent=etl availability=implemented stream=api_v1_iam_governance_bundles]
    api v1 iam governance opt in list - Run the api v1 iam governance opt in ETL stream [intent=etl availability=implemented stream=api_v1_iam_governance_opt_in]
    api v1 iam resource sets list - Run the api v1 iam resource sets ETL stream [intent=etl availability=implemented stream=api_v1_iam_resource_sets]
    api v1 iam resource sets resource set id or label bindings list - Run the api v1 iam resource sets resource set id or label bindings ETL stream [intent=etl availability=implemented stream=api_v1_iam_resource_sets_resource_set_id_or_label_bindings]
    api v1 iam resource sets resource set id or label bindings role id or label list - Run the api v1 iam resource sets resource set id or label bindings role id or label ETL stream [intent=etl availability=implemented stream=api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label]
    api v1 iam resource sets resource set id or label bindings role id or label members list - Run the api v1 iam resource sets resource set id or label bindings role id or label members ETL stream [intent=etl availability=implemented stream=api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members]
    api v1 iam resource sets resource set id or label bindings role id or label members member id list - Run the api v1 iam resource sets resource set id or label bindings role id or label members member id ETL stream [intent=etl availability=implemented stream=api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members_member_id]
    api v1 iam resource sets resource set id or label list - Run the api v1 iam resource sets resource set id or label ETL stream [intent=etl availability=implemented stream=api_v1_iam_resource_sets_resource_set_id_or_label]
    api v1 iam resource sets resource set id or label resources list - Run the api v1 iam resource sets resource set id or label resources ETL stream [intent=etl availability=implemented stream=api_v1_iam_resource_sets_resource_set_id_or_label_resources]
    api v1 iam resource sets resource set id or label resources resource id list - Run the api v1 iam resource sets resource set id or label resources resource id ETL stream [intent=etl availability=implemented stream=api_v1_iam_resource_sets_resource_set_id_or_label_resources_resource_id]
    api v1 iam roles list - Run the api v1 iam roles ETL stream [intent=etl availability=implemented stream=api_v1_iam_roles]
    api v1 iam roles role id or label list - Run the api v1 iam roles role id or label ETL stream [intent=etl availability=implemented stream=api_v1_iam_roles_role_id_or_label]
    api v1 iam roles role id or label permissions list - Run the api v1 iam roles role id or label permissions ETL stream [intent=etl availability=implemented stream=api_v1_iam_roles_role_id_or_label_permissions]
    api v1 iam roles role id or label permissions permission type list - Run the api v1 iam roles role id or label permissions permission type ETL stream [intent=etl availability=implemented stream=api_v1_iam_roles_role_id_or_label_permissions_permission_type]
    api v1 identity sources identity source id groups group or external id list - Run the api v1 identity sources identity source id groups group or external id ETL stream [intent=etl availability=implemented stream=api_v1_identity_sources_identity_source_id_groups_group_or_external_id]
    api v1 identity sources identity source id groups group or external id membership list - Run the api v1 identity sources identity source id groups group or external id membership ETL stream [intent=etl availability=implemented stream=api_v1_identity_sources_identity_source_id_groups_group_or_external_id_membership]
    api v1 identity sources identity source id sessions list - Run the api v1 identity sources identity source id sessions ETL stream [intent=etl availability=implemented stream=api_v1_identity_sources_identity_source_id_sessions]
    api v1 identity sources identity source id sessions session id list - Run the api v1 identity sources identity source id sessions session id ETL stream [intent=etl availability=implemented stream=api_v1_identity_sources_identity_source_id_sessions_session_id]
    api v1 identity sources identity source id users external id list - Run the api v1 identity sources identity source id users external id ETL stream [intent=etl availability=implemented stream=api_v1_identity_sources_identity_source_id_users_external_id]
    api v1 idps credentials keys kid list - Run the api v1 idps credentials keys kid ETL stream [intent=etl availability=implemented stream=api_v1_idps_credentials_keys_kid]
    api v1 idps credentials keys list - Run the api v1 idps credentials keys ETL stream [intent=etl availability=implemented stream=api_v1_idps_credentials_keys]
    api v1 idps idp id credentials csrs idp csr id list - Run the api v1 idps idp id credentials csrs idp csr id ETL stream [intent=etl availability=implemented stream=api_v1_idps_idp_id_credentials_csrs_idp_csr_id]
    api v1 idps idp id credentials csrs list - Run the api v1 idps idp id credentials csrs ETL stream [intent=etl availability=implemented stream=api_v1_idps_idp_id_credentials_csrs]
    api v1 idps idp id credentials keys active list - Run the api v1 idps idp id credentials keys active ETL stream [intent=etl availability=implemented stream=api_v1_idps_idp_id_credentials_keys_active]
    api v1 idps idp id credentials keys kid list - Run the api v1 idps idp id credentials keys kid ETL stream [intent=etl availability=implemented stream=api_v1_idps_idp_id_credentials_keys_kid]
    api v1 idps idp id credentials keys list - Run the api v1 idps idp id credentials keys ETL stream [intent=etl availability=implemented stream=api_v1_idps_idp_id_credentials_keys]
    api v1 idps idp id list - Run the api v1 idps idp id ETL stream [intent=etl availability=implemented stream=api_v1_idps_idp_id]
    api v1 idps idp id users list - Run the api v1 idps idp id users ETL stream [intent=etl availability=implemented stream=api_v1_idps_idp_id_users]
    api v1 idps idp id users user id credentials tokens list - Run the api v1 idps idp id users user id credentials tokens ETL stream [intent=etl availability=implemented stream=api_v1_idps_idp_id_users_user_id_credentials_tokens]
    api v1 idps idp id users user id list - Run the api v1 idps idp id users user id ETL stream [intent=etl availability=implemented stream=api_v1_idps_idp_id_users_user_id]
    api v1 idps list - Run the api v1 idps ETL stream [intent=etl availability=implemented stream=api_v1_idps]
    api v1 inline hooks inline hook id list - Run the api v1 inline hooks inline hook id ETL stream [intent=etl availability=implemented stream=api_v1_inline_hooks_inline_hook_id]
    api v1 inline hooks list - Run the api v1 inline hooks ETL stream [intent=etl availability=implemented stream=api_v1_inline_hooks]
    api v1 log streams list - Run the api v1 log streams ETL stream [intent=etl availability=implemented stream=api_v1_log_streams]
    api v1 log streams log stream id list - Run the api v1 log streams log stream id ETL stream [intent=etl availability=implemented stream=api_v1_log_streams_log_stream_id]
    api v1 mappings list - Run the api v1 mappings ETL stream [intent=etl availability=implemented stream=api_v1_mappings]
    api v1 mappings mapping id list - Run the api v1 mappings mapping id ETL stream [intent=etl availability=implemented stream=api_v1_mappings_mapping_id]
    api v1 meta schemas apps app id default list - Run the api v1 meta schemas apps app id default ETL stream [intent=etl availability=implemented stream=api_v1_meta_schemas_apps_app_id_default]
    api v1 meta schemas group default list - Run the api v1 meta schemas group default ETL stream [intent=etl availability=implemented stream=api_v1_meta_schemas_group_default]
    api v1 meta schemas log stream list - Run the api v1 meta schemas log stream ETL stream [intent=etl availability=implemented stream=api_v1_meta_schemas_log_stream]
    api v1 meta schemas log stream log stream type list - Run the api v1 meta schemas log stream log stream type ETL stream [intent=etl availability=implemented stream=api_v1_meta_schemas_log_stream_log_stream_type]
    api v1 meta schemas user linked objects linked object name list - Run the api v1 meta schemas user linked objects linked object name ETL stream [intent=etl availability=implemented stream=api_v1_meta_schemas_user_linked_objects_linked_object_name]
    api v1 meta schemas user linked objects list - Run the api v1 meta schemas user linked objects ETL stream [intent=etl availability=implemented stream=api_v1_meta_schemas_user_linked_objects]
    api v1 meta schemas user schema id list - Run the api v1 meta schemas user schema id ETL stream [intent=etl availability=implemented stream=api_v1_meta_schemas_user_schema_id]
    api v1 meta types user list - Run the api v1 meta types user ETL stream [intent=etl availability=implemented stream=api_v1_meta_types_user]
    api v1 meta types user type id list - Run the api v1 meta types user type id ETL stream [intent=etl availability=implemented stream=api_v1_meta_types_user_type_id]
    api v1 meta uischemas id list - Run the api v1 meta uischemas id ETL stream [intent=etl availability=implemented stream=api_v1_meta_uischemas_id]
    api v1 meta uischemas list - Run the api v1 meta uischemas ETL stream [intent=etl availability=implemented stream=api_v1_meta_uischemas]
    api v1 org captcha list - Run the api v1 org captcha ETL stream [intent=etl availability=implemented stream=api_v1_org_captcha]
    api v1 org contacts contact type list - Run the api v1 org contacts contact type ETL stream [intent=etl availability=implemented stream=api_v1_org_contacts_contact_type]
    api v1 org contacts list - Run the api v1 org contacts ETL stream [intent=etl availability=implemented stream=api_v1_org_contacts]
    api v1 org factors yubikey token tokens list - Run the api v1 org factors yubikey token tokens ETL stream [intent=etl availability=implemented stream=api_v1_org_factors_yubikey_token_tokens]
    api v1 org factors yubikey token tokens token id list - Run the api v1 org factors yubikey token tokens token id ETL stream [intent=etl availability=implemented stream=api_v1_org_factors_yubikey_token_tokens_token_id]
    api v1 org list - Run the api v1 org ETL stream [intent=etl availability=implemented stream=api_v1_org]
    api v1 org org settings third party admin setting list - Run the api v1 org org settings third party admin setting ETL stream [intent=etl availability=implemented stream=api_v1_org_org_settings_third_party_admin_setting]
    api v1 org preferences list - Run the api v1 org preferences ETL stream [intent=etl availability=implemented stream=api_v1_org_preferences]
    api v1 org privacy aerial list - Run the api v1 org privacy aerial ETL stream [intent=etl availability=implemented stream=api_v1_org_privacy_aerial]
    api v1 org privacy okta communication list - Run the api v1 org privacy okta communication ETL stream [intent=etl availability=implemented stream=api_v1_org_privacy_okta_communication]
    api v1 org privacy okta support cases list - Run the api v1 org privacy okta support cases ETL stream [intent=etl availability=implemented stream=api_v1_org_privacy_okta_support_cases]
    api v1 org privacy okta support list - Run the api v1 org privacy okta support ETL stream [intent=etl availability=implemented stream=api_v1_org_privacy_okta_support]
    api v1 org settings auto assign admin app setting list - Run the api v1 org settings auto assign admin app setting ETL stream [intent=etl availability=implemented stream=api_v1_org_settings_auto_assign_admin_app_setting]
    api v1 org settings client privileges setting list - Run the api v1 org settings client privileges setting ETL stream [intent=etl availability=implemented stream=api_v1_org_settings_client_privileges_setting]
    api v1 policies list - Run the api v1 policies ETL stream [intent=etl availability=implemented stream=api_v1_policies]
    api v1 policies policy id app list - Run the api v1 policies policy id app ETL stream [intent=etl availability=implemented stream=api_v1_policies_policy_id_app]
    api v1 policies policy id list - Run the api v1 policies policy id ETL stream [intent=etl availability=implemented stream=api_v1_policies_policy_id]
    api v1 policies policy id mappings list - Run the api v1 policies policy id mappings ETL stream [intent=etl availability=implemented stream=api_v1_policies_policy_id_mappings]
    api v1 policies policy id mappings mapping id list - Run the api v1 policies policy id mappings mapping id ETL stream [intent=etl availability=implemented stream=api_v1_policies_policy_id_mappings_mapping_id]
    api v1 policies policy id rules list - Run the api v1 policies policy id rules ETL stream [intent=etl availability=implemented stream=api_v1_policies_policy_id_rules]
    api v1 policies policy id rules rule id list - Run the api v1 policies policy id rules rule id ETL stream [intent=etl availability=implemented stream=api_v1_policies_policy_id_rules_rule_id]
    api v1 principal rate limits list - Run the api v1 principal rate limits ETL stream [intent=etl availability=implemented stream=api_v1_principal_rate_limits]
    api v1 principal rate limits principal rate limit id list - Run the api v1 principal rate limits principal rate limit id ETL stream [intent=etl availability=implemented stream=api_v1_principal_rate_limits_principal_rate_limit_id]
    api v1 push providers list - Run the api v1 push providers ETL stream [intent=etl availability=implemented stream=api_v1_push_providers]
    api v1 push providers push provider id list - Run the api v1 push providers push provider id ETL stream [intent=etl availability=implemented stream=api_v1_push_providers_push_provider_id]
    api v1 rate limit settings admin notifications list - Run the api v1 rate limit settings admin notifications ETL stream [intent=etl availability=implemented stream=api_v1_rate_limit_settings_admin_notifications]
    api v1 rate limit settings per client list - Run the api v1 rate limit settings per client ETL stream [intent=etl availability=implemented stream=api_v1_rate_limit_settings_per_client]
    api v1 rate limit settings warning threshold list - Run the api v1 rate limit settings warning threshold ETL stream [intent=etl availability=implemented stream=api_v1_rate_limit_settings_warning_threshold]
    api v1 realm assignments assignment id list - Run the api v1 realm assignments assignment id ETL stream [intent=etl availability=implemented stream=api_v1_realm_assignments_assignment_id]
    api v1 realm assignments list - Run the api v1 realm assignments ETL stream [intent=etl availability=implemented stream=api_v1_realm_assignments]
    api v1 realm assignments operations list - Run the api v1 realm assignments operations ETL stream [intent=etl availability=implemented stream=api_v1_realm_assignments_operations]
    api v1 realms list - Run the api v1 realms ETL stream [intent=etl availability=implemented stream=api_v1_realms]
    api v1 realms realm id list - Run the api v1 realms realm id ETL stream [intent=etl availability=implemented stream=api_v1_realms_realm_id]
    api v1 roles role ref subscriptions list - Run the api v1 roles role ref subscriptions ETL stream [intent=etl availability=implemented stream=api_v1_roles_role_ref_subscriptions]
    api v1 roles role ref subscriptions notification type list - Run the api v1 roles role ref subscriptions notification type ETL stream [intent=etl availability=implemented stream=api_v1_roles_role_ref_subscriptions_notification_type]
    api v1 security events providers list - Run the api v1 security events providers ETL stream [intent=etl availability=implemented stream=api_v1_security_events_providers]
    api v1 security events providers security event provider id list - Run the api v1 security events providers security event provider id ETL stream [intent=etl availability=implemented stream=api_v1_security_events_providers_security_event_provider_id]
    api v1 sessions session id list - Run the api v1 sessions session id ETL stream [intent=etl availability=implemented stream=api_v1_sessions_session_id]
    api v1 ssf stream list - Run the api v1 ssf stream ETL stream [intent=etl availability=implemented stream=api_v1_ssf_stream]
    api v1 ssf stream status list - Run the api v1 ssf stream status ETL stream [intent=etl availability=implemented stream=api_v1_ssf_stream_status]
    api v1 telephony providers custom telephony provider id list - Run the api v1 telephony providers custom telephony provider id ETL stream [intent=etl availability=implemented stream=api_v1_telephony_providers_custom_telephony_provider_id]
    api v1 telephony providers list - Run the api v1 telephony providers ETL stream [intent=etl availability=implemented stream=api_v1_telephony_providers]
    api v1 templates sms list - Run the api v1 templates sms ETL stream [intent=etl availability=implemented stream=api_v1_templates_sms]
    api v1 templates sms template id list - Run the api v1 templates sms template id ETL stream [intent=etl availability=implemented stream=api_v1_templates_sms_template_id]
    api v1 threats configuration list - Run the api v1 threats configuration ETL stream [intent=etl availability=implemented stream=api_v1_threats_configuration]
    api v1 trusted origins list - Run the api v1 trusted origins ETL stream [intent=etl availability=implemented stream=api_v1_trusted_origins]
    api v1 trusted origins trusted origin id list - Run the api v1 trusted origins trusted origin id ETL stream [intent=etl availability=implemented stream=api_v1_trusted_origins_trusted_origin_id]
    api v1 users id app links list - Run the api v1 users id app links ETL stream [intent=etl availability=implemented stream=api_v1_users_id_app_links]
    api v1 users id blocks list - Run the api v1 users id blocks ETL stream [intent=etl availability=implemented stream=api_v1_users_id_blocks]
    api v1 users id groups list - Run the api v1 users id groups ETL stream [intent=etl availability=implemented stream=api_v1_users_id_groups]
    api v1 users id idps list - Run the api v1 users id idps ETL stream [intent=etl availability=implemented stream=api_v1_users_id_idps]
    api v1 users id list - Run the api v1 users id ETL stream [intent=etl availability=implemented stream=api_v1_users_id]
    api v1 users user id authenticator enrollments enrollment id list - Run the api v1 users user id authenticator enrollments enrollment id ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_authenticator_enrollments_enrollment_id]
    api v1 users user id authenticator enrollments list - Run the api v1 users user id authenticator enrollments ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_authenticator_enrollments]
    api v1 users user id classification list - Run the api v1 users user id classification ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_classification]
    api v1 users user id clients client id grants list - Run the api v1 users user id clients client id grants ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_clients_client_id_grants]
    api v1 users user id clients client id tokens list - Run the api v1 users user id clients client id tokens ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_clients_client_id_tokens]
    api v1 users user id clients client id tokens token id list - Run the api v1 users user id clients client id tokens token id ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_clients_client_id_tokens_token_id]
    api v1 users user id clients list - Run the api v1 users user id clients ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_clients]
    api v1 users user id devices list - Run the api v1 users user id devices ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_devices]
    api v1 users user id factors catalog list - Run the api v1 users user id factors catalog ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_factors_catalog]
    api v1 users user id factors factor id list - Run the api v1 users user id factors factor id ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_factors_factor_id]
    api v1 users user id factors factor id transactions transaction id list - Run the api v1 users user id factors factor id transactions transaction id ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_factors_factor_id_transactions_transaction_id]
    api v1 users user id factors list - Run the api v1 users user id factors ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_factors]
    api v1 users user id factors questions list - Run the api v1 users user id factors questions ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_factors_questions]
    api v1 users user id grants grant id list - Run the api v1 users user id grants grant id ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_grants_grant_id]
    api v1 users user id grants list - Run the api v1 users user id grants ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_grants]
    api v1 users user id or login linked objects relationship name list - Run the api v1 users user id or login linked objects relationship name ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_or_login_linked_objects_relationship_name]
    api v1 users user id risk list - Run the api v1 users user id risk ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_risk]
    api v1 users user id roles list - Run the api v1 users user id roles ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_roles]
    api v1 users user id roles role assignment id governance grant id list - Run the api v1 users user id roles role assignment id governance grant id ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_roles_role_assignment_id_governance_grant_id]
    api v1 users user id roles role assignment id governance grant id resources list - Run the api v1 users user id roles role assignment id governance grant id resources ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_roles_role_assignment_id_governance_grant_id_resources]
    api v1 users user id roles role assignment id governance list - Run the api v1 users user id roles role assignment id governance ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_roles_role_assignment_id_governance]
    api v1 users user id roles role assignment id list - Run the api v1 users user id roles role assignment id ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_roles_role_assignment_id]
    api v1 users user id roles role assignment id targets catalog apps list - Run the api v1 users user id roles role assignment id targets catalog apps ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps]
    api v1 users user id roles role assignment id targets groups list - Run the api v1 users user id roles role assignment id targets groups ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_roles_role_assignment_id_targets_groups]
    api v1 users user id roles role id or encoded role id targets list - Run the api v1 users user id roles role id or encoded role id targets ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_roles_role_id_or_encoded_role_id_targets]
    api v1 users user id subscriptions list - Run the api v1 users user id subscriptions ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_subscriptions]
    api v1 users user id subscriptions notification type list - Run the api v1 users user id subscriptions notification type ETL stream [intent=etl availability=implemented stream=api_v1_users_user_id_subscriptions_notification_type]
    api v1 zones list - Run the api v1 zones ETL stream [intent=etl availability=implemented stream=api_v1_zones]
    api v1 zones zone id list - Run the api v1 zones zone id ETL stream [intent=etl availability=implemented stream=api_v1_zones_zone_id]
    attack protection api v1 authenticator settings list - Run the attack protection api v1 authenticator settings ETL stream [intent=etl availability=implemented stream=attack_protection_api_v1_authenticator_settings]
    attack protection api v1 user lockout settings list - Run the attack protection api v1 user lockout settings ETL stream [intent=etl availability=implemented stream=attack_protection_api_v1_user_lockout_settings]
    create api v1 agent pools pool id updates apply - Plan and execute the create api v1 agent pools pool id updates reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_agent_pools_pool_id_updates]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --pool_id (required)
    create api v1 agent pools pool id updates settings apply - Plan and execute the create api v1 agent pools pool id updates settings reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_agent_pools_pool_id_updates_settings]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --agentType (required), --pool_id (required)
    create api v1 agent pools pool id updates update id apply - Plan and execute the create api v1 agent pools pool id updates update id reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_agent_pools_pool_id_updates_update_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --pool_id (required), --update_id (required)
    create api v1 apps app id connections default apply - Plan and execute the create api v1 apps app id connections default reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_api_v1_apps_app_id_connections_default]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create api v1 apps app id credentials csrs apply - Plan and execute the create api v1 apps app id credentials csrs reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_apps_app_id_credentials_csrs]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required)
    create api v1 apps app id credentials jwks apply - Plan and execute the create api v1 apps app id credentials jwks reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_apps_app_id_credentials_jwks]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required)
    create api v1 apps app id credentials secrets apply - Plan and execute the create api v1 apps app id credentials secrets reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_apps_app_id_credentials_secrets]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required)
    create api v1 apps app id cwo connections apply - Plan and execute the create api v1 apps app id cwo connections reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_apps_app_id_cwo_connections]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required)
    create api v1 apps app id federated claims apply - Plan and execute the create api v1 apps app id federated claims reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_apps_app_id_federated_claims]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required)
    create api v1 apps app id grants apply - Plan and execute the create api v1 apps app id grants reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_apps_app_id_grants]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required), --issuer (required), --scopeId (required)
    create api v1 apps app id group push mappings apply - Plan and execute the create api v1 apps app id group push mappings reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_apps_app_id_group_push_mappings]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required), --sourceGroupId (required)
    create api v1 apps app id interclient allowed apps apply - Plan and execute the create api v1 apps app id interclient allowed apps reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_apps_app_id_interclient_allowed_apps]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required)
    create api v1 apps app id users apply - Plan and execute the create api v1 apps app id users reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_apps_app_id_users]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required), --id (required)
    create api v1 apps app id users user id apply - Plan and execute the create api v1 apps app id users user id reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_apps_app_id_users_user_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required), --user_id (required)
    create api v1 apps app name app id oauth2 callback apply - Plan and execute the create api v1 apps app name app id oauth2 callback reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_apps_app_name_app_id_oauth2_callback]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required), --app_name (required)
    create api v1 apps apply - Plan and execute the create api v1 apps reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_apps]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --label (required), --signOnMode (required)
    create api v1 authenticators apply - Plan and execute the create api v1 authenticators reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_authenticators]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 authenticators authenticator id aaguids apply - Plan and execute the create api v1 authenticators authenticator id aaguids reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_authenticators_authenticator_id_aaguids]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --authenticator_id (required)
    create api v1 authenticators authenticator id methods web authn method type verify rp id domain apply - Plan and execute the create api v1 authenticators authenticator id methods web authn method type verify rp id domain reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_authenticators_authenticator_id_methods_web_authn_method_type_verify_rp_id_domain]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --authenticator_id (required), --web_authn_method_type (required)
    create api v1 authorization servers apply - Plan and execute the create api v1 authorization servers reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_authorization_servers]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 authorization servers auth server id associated servers apply - Plan and execute the create api v1 authorization servers auth server id associated servers reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_authorization_servers_auth_server_id_associated_servers]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --auth_server_id (required)
    create api v1 authorization servers auth server id claims apply - Plan and execute the create api v1 authorization servers auth server id claims reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_authorization_servers_auth_server_id_claims]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --auth_server_id (required)
    create api v1 authorization servers auth server id policies apply - Plan and execute the create api v1 authorization servers auth server id policies reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_authorization_servers_auth_server_id_policies]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --auth_server_id (required)
    create api v1 authorization servers auth server id policies policy id rules apply - Plan and execute the create api v1 authorization servers auth server id policies policy id rules reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create api v1 authorization servers auth server id resourceservercredentials keys apply - Plan and execute the create api v1 authorization servers auth server id resourceservercredentials keys reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --auth_server_id (required)
    create api v1 authorization servers auth server id scopes apply - Plan and execute the create api v1 authorization servers auth server id scopes reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_authorization_servers_auth_server_id_scopes]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --auth_server_id (required), --name (required)
    create api v1 behaviors apply - Plan and execute the create api v1 behaviors reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_behaviors]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --name (required), --type (required)
    create api v1 bot protection configuration apply - Plan and execute the create api v1 bot protection configuration reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_bot_protection_configuration]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --level (required), --mode (required)
    create api v1 brands apply - Plan and execute the create api v1 brands reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_brands]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --name (required)
    create api v1 brands brand id templates email template name customizations apply - Plan and execute the create api v1 brands brand id templates email template name customizations reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_brands_brand_id_templates_email_template_name_customizations]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --brand_id (required), --language (required), --template_name (required)
    create api v1 brands brand id templates email template name test apply - Plan and execute the create api v1 brands brand id templates email template name test reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_brands_brand_id_templates_email_template_name_test]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --brand_id (required), --template_name (required)
    create api v1 captchas apply - Plan and execute the create api v1 captchas reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_captchas]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 captchas captcha id apply - Plan and execute the create api v1 captchas captcha id reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_captchas_captcha_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --captcha_id (required)
    create api v1 device assurances apply - Plan and execute the create api v1 device assurances reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_device_assurances]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 device posture checks apply - Plan and execute the create api v1 device posture checks reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_device_posture_checks]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 directories app instance id groups group id query apply - Plan and execute the create api v1 directories app instance id groups group id query reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_directories_app_instance_id_groups_group_id_query]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_instance_id (required), --attributes (required), --group_id (required)
    create api v1 directories app instance id groups modify apply - Plan and execute the create api v1 directories app instance id groups modify reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_api_v1_directories_app_instance_id_groups_modify]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create api v1 domains apply - Plan and execute the create api v1 domains reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_domains]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --certificateSourceType (required), --domain (required)
    create api v1 domains domain id verify apply - Plan and execute the create api v1 domains domain id verify reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_domains_domain_id_verify]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --domain_id (required)
    create api v1 dr failback apply - Plan and execute the create api v1 dr failback reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_dr_failback]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 dr failover apply - Plan and execute the create api v1 dr failover reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_dr_failover]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 email domains apply - Plan and execute the create api v1 email domains reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_email_domains]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --displayName (required), --userName (required)
    create api v1 email domains email domain id verify apply - Plan and execute the create api v1 email domains email domain id verify reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_email_domains_email_domain_id_verify]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --email_domain_id (required)
    create api v1 email servers apply - Plan and execute the create api v1 email servers reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_email_servers]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --alias (required), --authType (required), --enabled (required), --host (required), --port (required), --username (required)
    create api v1 email servers email server id test apply - Plan and execute the create api v1 email servers email server id test reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_email_servers_email_server_id_test]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --email_server_id (required), --fromAddress (required), --toAddress (required)
    create api v1 event hooks apply - Plan and execute the create api v1 event hooks reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_api_v1_event_hooks]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create api v1 features feature id lifecycle apply - Plan and execute the create api v1 features feature id lifecycle reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_features_feature_id_lifecycle]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --feature_id (required), --lifecycle (required)
    create api v1 groups apply - Plan and execute the create api v1 groups reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_groups]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 groups group id owners apply - Plan and execute the create api v1 groups group id owners reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_groups_group_id_owners]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --group_id (required)
    create api v1 groups group id roles apply - Plan and execute the create api v1 groups group id roles reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_groups_group_id_roles]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --group_id (required), --type (required)
    create api v1 groups rules apply - Plan and execute the create api v1 groups rules reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_groups_rules]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 hook keys apply - Plan and execute the create api v1 hook keys reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_hook_keys]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 iam governance bundles apply - Plan and execute the create api v1 iam governance bundles reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_iam_governance_bundles]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 iam governance opt in apply - Plan and execute the create api v1 iam governance opt in reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_iam_governance_opt_in]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 iam governance opt out apply - Plan and execute the create api v1 iam governance opt out reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_iam_governance_opt_out]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 iam resource sets apply - Plan and execute the create api v1 iam resource sets reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_iam_resource_sets]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --description (required), --label (required), --resources (required)
    create api v1 iam resource sets resource set id or label bindings apply - Plan and execute the create api v1 iam resource sets resource set id or label bindings reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_iam_resource_sets_resource_set_id_or_label_bindings]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --resource_set_id_or_label (required)
    create api v1 iam resource sets resource set id or label resources apply - Plan and execute the create api v1 iam resource sets resource set id or label resources reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_api_v1_iam_resource_sets_resource_set_id_or_label_resources]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create api v1 iam roles apply - Plan and execute the create api v1 iam roles reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_iam_roles]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --description (required), --label (required), --permissions (required)
    create api v1 iam roles role id or label permissions permission type apply - Plan and execute the create api v1 iam roles role id or label permissions permission type reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_iam_roles_role_id_or_label_permissions_permission_type]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --permission_type (required), --role_id_or_label (required)
    create api v1 identity sources identity source id groups apply - Plan and execute the create api v1 identity sources identity source id groups reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_identity_sources_identity_source_id_groups]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --identity_source_id (required)
    create api v1 identity sources identity source id groups group or external id apply - Plan and execute the create api v1 identity sources identity source id groups group or external id reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_identity_sources_identity_source_id_groups_group_or_external_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --group_or_external_id (required), --identity_source_id (required)
    create api v1 identity sources identity source id groups group or external id membership apply - Plan and execute the create api v1 identity sources identity source id groups group or external id membership reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_identity_sources_identity_source_id_groups_group_or_external_id_membership]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --group_or_external_id (required), --identity_source_id (required)
    create api v1 identity sources identity source id sessions apply - Plan and execute the create api v1 identity sources identity source id sessions reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_identity_sources_identity_source_id_sessions]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --identity_source_id (required)
    create api v1 identity sources identity source id sessions session id bulk delete apply - Plan and execute the create api v1 identity sources identity source id sessions session id bulk delete reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_delete]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --identity_source_id (required), --session_id (required)
    create api v1 identity sources identity source id sessions session id bulk group memberships delete apply - Plan and execute the create api v1 identity sources identity source id sessions session id bulk group memberships delete reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_group_memberships_delete]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --identity_source_id (required), --session_id (required)
    create api v1 identity sources identity source id sessions session id bulk group memberships upsert apply - Plan and execute the create api v1 identity sources identity source id sessions session id bulk group memberships upsert reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_group_memberships_upsert]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --identity_source_id (required), --session_id (required)
    create api v1 identity sources identity source id sessions session id bulk groups delete apply - Plan and execute the create api v1 identity sources identity source id sessions session id bulk groups delete reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_groups_delete]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --identity_source_id (required), --session_id (required)
    create api v1 identity sources identity source id sessions session id bulk groups upsert apply - Plan and execute the create api v1 identity sources identity source id sessions session id bulk groups upsert reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_groups_upsert]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --identity_source_id (required), --session_id (required)
    create api v1 identity sources identity source id sessions session id bulk upsert apply - Plan and execute the create api v1 identity sources identity source id sessions session id bulk upsert reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_upsert]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --identity_source_id (required), --session_id (required)
    create api v1 identity sources identity source id sessions session id start import apply - Plan and execute the create api v1 identity sources identity source id sessions session id start import reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_identity_sources_identity_source_id_sessions_session_id_start_import]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --identity_source_id (required), --session_id (required)
    create api v1 identity sources identity source id users apply - Plan and execute the create api v1 identity sources identity source id users reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_identity_sources_identity_source_id_users]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --identity_source_id (required)
    create api v1 idps apply - Plan and execute the create api v1 idps reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_idps]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 idps credentials keys apply - Plan and execute the create api v1 idps credentials keys reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_idps_credentials_keys]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --x5c (required)
    create api v1 idps idp id credentials csrs apply - Plan and execute the create api v1 idps idp id credentials csrs reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_idps_idp_id_credentials_csrs]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --idp_id (required)
    create api v1 idps idp id users user id apply - Plan and execute the create api v1 idps idp id users user id reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_idps_idp_id_users_user_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --idp_id (required), --user_id (required)
    create api v1 inline hooks apply - Plan and execute the create api v1 inline hooks reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_inline_hooks]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 inline hooks inline hook id apply - Plan and execute the create api v1 inline hooks inline hook id reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_inline_hooks_inline_hook_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --inline_hook_id (required)
    create api v1 inline hooks inline hook id execute apply - Plan and execute the create api v1 inline hooks inline hook id execute reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_inline_hooks_inline_hook_id_execute]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --inline_hook_id (required)
    create api v1 log streams apply - Plan and execute the create api v1 log streams reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_api_v1_log_streams]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create api v1 mappings mapping id apply - Plan and execute the create api v1 mappings mapping id reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_api_v1_mappings_mapping_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create api v1 meta schemas apps app id default apply - Plan and execute the create api v1 meta schemas apps app id default reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_meta_schemas_apps_app_id_default]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required)
    create api v1 meta schemas group default apply - Plan and execute the create api v1 meta schemas group default reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_meta_schemas_group_default]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 meta schemas user linked objects apply - Plan and execute the create api v1 meta schemas user linked objects reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_meta_schemas_user_linked_objects]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 meta schemas user schema id apply - Plan and execute the create api v1 meta schemas user schema id reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_meta_schemas_user_schema_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --schema_id (required)
    create api v1 meta types user apply - Plan and execute the create api v1 meta types user reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_meta_types_user]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --displayName (required), --name (required)
    create api v1 meta types user type id apply - Plan and execute the create api v1 meta types user type id reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_meta_types_user_type_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --type_id (required)
    create api v1 meta uischemas apply - Plan and execute the create api v1 meta uischemas reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_meta_uischemas]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 org apply - Plan and execute the create api v1 org reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_org]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 org email bounces remove list apply - Plan and execute the create api v1 org email bounces remove list reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_org_email_bounces_remove_list]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 org factors yubikey token tokens apply - Plan and execute the create api v1 org factors yubikey token tokens reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_org_factors_yubikey_token_tokens]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 org org settings third party admin setting apply - Plan and execute the create api v1 org org settings third party admin setting reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_org_org_settings_third_party_admin_setting]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 org preferences hide end user footer apply - Plan and execute the create api v1 org preferences hide end user footer reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_org_preferences_hide_end_user_footer]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 org preferences show end user footer apply - Plan and execute the create api v1 org preferences show end user footer reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_org_preferences_show_end_user_footer]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 org privacy aerial grant apply - Plan and execute the create api v1 org privacy aerial grant reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_org_privacy_aerial_grant]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --accountId (required)
    create api v1 org privacy okta communication opt in apply - Plan and execute the create api v1 org privacy okta communication opt in reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_org_privacy_okta_communication_opt_in]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 org privacy okta communication opt out apply - Plan and execute the create api v1 org privacy okta communication opt out reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_org_privacy_okta_communication_opt_out]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 org privacy okta support extend apply - Plan and execute the create api v1 org privacy okta support extend reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_org_privacy_okta_support_extend]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 org privacy okta support grant apply - Plan and execute the create api v1 org privacy okta support grant reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_org_privacy_okta_support_grant]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 org settings auto assign admin app setting apply - Plan and execute the create api v1 org settings auto assign admin app setting reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_org_settings_auto_assign_admin_app_setting]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 orgs apply - Plan and execute the create api v1 orgs reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_api_v1_orgs]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create api v1 policies apply - Plan and execute the create api v1 policies reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_policies]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --name (required), --type (required)
    create api v1 policies policy id clone apply - Plan and execute the create api v1 policies policy id clone reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_policies_policy_id_clone]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --policy_id (required)
    create api v1 policies policy id mappings apply - Plan and execute the create api v1 policies policy id mappings reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_policies_policy_id_mappings]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --policy_id (required)
    create api v1 policies policy id rules apply - Plan and execute the create api v1 policies policy id rules reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_policies_policy_id_rules]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --policy_id (required)
    create api v1 principal rate limits apply - Plan and execute the create api v1 principal rate limits reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_principal_rate_limits]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --principalId (required), --principalType (required)
    create api v1 push providers apply - Plan and execute the create api v1 push providers reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_push_providers]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 realm assignments apply - Plan and execute the create api v1 realm assignments reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_realm_assignments]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 realm assignments operations apply - Plan and execute the create api v1 realm assignments operations reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_realm_assignments_operations]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 realms apply - Plan and execute the create api v1 realms reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_realms]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 roles role ref subscriptions notification type subscribe apply - Plan and execute the create api v1 roles role ref subscriptions notification type subscribe reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_roles_role_ref_subscriptions_notification_type_subscribe]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --notification_type (required), --role_ref (required)
    create api v1 roles role ref subscriptions notification type unsubscribe apply - Plan and execute the create api v1 roles role ref subscriptions notification type unsubscribe reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_roles_role_ref_subscriptions_notification_type_unsubscribe]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --notification_type (required), --role_ref (required)
    create api v1 security events providers apply - Plan and execute the create api v1 security events providers reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_api_v1_security_events_providers]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create api v1 ssf stream apply - Plan and execute the create api v1 ssf stream reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_api_v1_ssf_stream]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create api v1 ssf stream verification apply - Plan and execute the create api v1 ssf stream verification reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_ssf_stream_verification]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --stream_id (required)
    create api v1 telephony providers apply - Plan and execute the create api v1 telephony providers reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_telephony_providers]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 telephony providers custom telephony provider id set as primary apply - Plan and execute the create api v1 telephony providers custom telephony provider id set as primary reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_telephony_providers_custom_telephony_provider_id_set_as_primary]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --custom_telephony_provider_id (required)
    create api v1 telephony providers custom telephony provider id test apply - Plan and execute the create api v1 telephony providers custom telephony provider id test reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_telephony_providers_custom_telephony_provider_id_test]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --custom_telephony_provider_id (required)
    create api v1 templates sms apply - Plan and execute the create api v1 templates sms reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_templates_sms]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 templates sms template id apply - Plan and execute the create api v1 templates sms template id reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_templates_sms_template_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --template_id (required)
    create api v1 threats configuration apply - Plan and execute the create api v1 threats configuration reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_threats_configuration]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --action (required)
    create api v1 trusted origins apply - Plan and execute the create api v1 trusted origins reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_trusted_origins]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create api v1 users apply - Plan and execute the create api v1 users reverse-ETL action [intent=reverse_etl availability=not_implemented write=create_api_v1_users]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    create api v1 users id apply - Plan and execute the create api v1 users id reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_users_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --id (required)
    create api v1 users user id authenticator enrollments phone apply - Plan and execute the create api v1 users user id authenticator enrollments phone reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_users_user_id_authenticator_enrollments_phone]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --user_id (required)
    create api v1 users user id authenticator enrollments tac apply - Plan and execute the create api v1 users user id authenticator enrollments tac reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_users_user_id_authenticator_enrollments_tac]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --user_id (required)
    create api v1 users user id credentials change password apply - Plan and execute the create api v1 users user id credentials change password reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_users_user_id_credentials_change_password]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --user_id (required)
    create api v1 users user id credentials change recovery question apply - Plan and execute the create api v1 users user id credentials change recovery question reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_users_user_id_credentials_change_recovery_question]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --user_id (required)
    create api v1 users user id credentials forgot password apply - Plan and execute the create api v1 users user id credentials forgot password reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_users_user_id_credentials_forgot_password]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --user_id (required)
    create api v1 users user id credentials forgot password recovery question apply - Plan and execute the create api v1 users user id credentials forgot password recovery question reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_users_user_id_credentials_forgot_password_recovery_question]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --user_id (required)
    create api v1 users user id factors apply - Plan and execute the create api v1 users user id factors reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_users_user_id_factors]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --user_id (required)
    create api v1 users user id factors factor id resend apply - Plan and execute the create api v1 users user id factors factor id resend reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_users_user_id_factors_factor_id_resend]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --factor_id (required), --user_id (required)
    create api v1 users user id factors factor id verify apply - Plan and execute the create api v1 users user id factors factor id verify reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_users_user_id_factors_factor_id_verify]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --factor_id (required), --user_id (required)
    create api v1 users user id roles apply - Plan and execute the create api v1 users user id roles reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_users_user_id_roles]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --type (required), --user_id (required)
    create api v1 users user id subscriptions notification type subscribe apply - Plan and execute the create api v1 users user id subscriptions notification type subscribe reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_users_user_id_subscriptions_notification_type_subscribe]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --notification_type (required), --user_id (required)
    create api v1 users user id subscriptions notification type unsubscribe apply - Plan and execute the create api v1 users user id subscriptions notification type unsubscribe reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_users_user_id_subscriptions_notification_type_unsubscribe]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --notification_type (required), --user_id (required)
    create api v1 zones apply - Plan and execute the create api v1 zones reverse-ETL action [intent=reverse_etl availability=implemented write=create_api_v1_zones]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --name (required), --type (required)
    create integrations api v1 api services api service id credentials secrets apply - Plan and execute the create integrations api v1 api services api service id credentials secrets reverse-ETL action [intent=reverse_etl availability=implemented write=create_integrations_api_v1_api_services_api_service_id_credentials_secrets]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --api_service_id (required)
    create integrations api v1 api services apply - Plan and execute the create integrations api v1 api services reverse-ETL action [intent=reverse_etl availability=implemented write=create_integrations_api_v1_api_services]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --grantedScopes (required), --type (required)
    create oauth2 v1 clients client id roles apply - Plan and execute the create oauth2 v1 clients client id roles reverse-ETL action [intent=reverse_etl availability=implemented write=create_oauth2_v1_clients_client_id_roles]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --client_id (required), --type (required)
    create privileged access api v1 okta service accounts apply - Plan and execute the create privileged access api v1 okta service accounts reverse-ETL action [intent=reverse_etl availability=implemented write=create_privileged_access_api_v1_okta_service_accounts]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --name (required), --oktaUserId (required)
    create privileged access api v1 service accounts apply - Plan and execute the create privileged access api v1 service accounts reverse-ETL action [intent=reverse_etl availability=implemented write=create_privileged_access_api_v1_service_accounts]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --containerOrn (required), --name (required), --password (required), --username (required)
    create webauthn registration api v1 enroll apply - Plan and execute the create webauthn registration api v1 enroll reverse-ETL action [intent=reverse_etl availability=implemented write=create_webauthn_registration_api_v1_enroll]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create webauthn registration api v1 initiate fulfillment request apply - Plan and execute the create webauthn registration api v1 initiate fulfillment request reverse-ETL action [intent=reverse_etl availability=implemented write=create_webauthn_registration_api_v1_initiate_fulfillment_request]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create webauthn registration api v1 send pin apply - Plan and execute the create webauthn registration api v1 send pin reverse-ETL action [intent=reverse_etl availability=implemented write=create_webauthn_registration_api_v1_send_pin]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    create webauthn registration api v1 users user id enrollments authenticator enrollment id mark error apply - Plan and execute the create webauthn registration api v1 users user id enrollments authenticator enrollment id mark error reverse-ETL action [intent=reverse_etl availability=implemented write=create_webauthn_registration_api_v1_users_user_id_enrollments_authenticator_enrollment_id_mark_error]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --authenticator_enrollment_id (required), --user_id (required)
    delete api v1 agent pools pool id updates update id apply - Plan and execute the delete api v1 agent pools pool id updates update id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_agent_pools_pool_id_updates_update_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --pool_id (required), --update_id (required)
    delete api v1 api tokens api token id apply - Plan and execute the delete api v1 api tokens api token id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_api_tokens_api_token_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --api_token_id (required)
    delete api v1 api tokens current apply - Plan and execute the delete api v1 api tokens current reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_api_tokens_current]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required
    delete api v1 apps app id apply - Plan and execute the delete api v1 apps app id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_apps_app_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required)
    delete api v1 apps app id credentials csrs csr id apply - Plan and execute the delete api v1 apps app id credentials csrs csr id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_apps_app_id_credentials_csrs_csr_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required), --csr_id (required)
    delete api v1 apps app id credentials jwks key id apply - Plan and execute the delete api v1 apps app id credentials jwks key id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_apps_app_id_credentials_jwks_key_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required), --key_id (required)
    delete api v1 apps app id credentials secrets secret id apply - Plan and execute the delete api v1 apps app id credentials secrets secret id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_apps_app_id_credentials_secrets_secret_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required), --secret_id (required)
    delete api v1 apps app id cwo connections connection id apply - Plan and execute the delete api v1 apps app id cwo connections connection id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_apps_app_id_cwo_connections_connection_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required), --connection_id (required)
    delete api v1 apps app id federated claims claim id apply - Plan and execute the delete api v1 apps app id federated claims claim id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_apps_app_id_federated_claims_claim_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required), --claim_id (required)
    delete api v1 apps app id grants grant id apply - Plan and execute the delete api v1 apps app id grants grant id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_apps_app_id_grants_grant_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required), --grant_id (required)
    delete api v1 apps app id groups group id apply - Plan and execute the delete api v1 apps app id groups group id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_apps_app_id_groups_group_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required), --group_id (required)
    delete api v1 apps app id interclient allowed apps allowed app id apply - Plan and execute the delete api v1 apps app id interclient allowed apps allowed app id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_apps_app_id_interclient_allowed_apps_allowed_app_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --allowed_app_id (required), --app_id (required)
    delete api v1 apps app id tokens apply - Plan and execute the delete api v1 apps app id tokens reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_apps_app_id_tokens]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required)
    delete api v1 apps app id tokens token id apply - Plan and execute the delete api v1 apps app id tokens token id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_apps_app_id_tokens_token_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required), --token_id (required)
    delete api v1 apps app id users user id apply - Plan and execute the delete api v1 apps app id users user id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_apps_app_id_users_user_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required), --user_id (required)
    delete api v1 authenticators authenticator id aaguids aaguid apply - Plan and execute the delete api v1 authenticators authenticator id aaguids aaguid reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_authenticators_authenticator_id_aaguids_aaguid]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --aaguid (required), --authenticator_id (required)
    delete api v1 authorization servers auth server id apply - Plan and execute the delete api v1 authorization servers auth server id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_authorization_servers_auth_server_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --auth_server_id (required)
    delete api v1 authorization servers auth server id associated servers associated server id apply - Plan and execute the delete api v1 authorization servers auth server id associated servers associated server id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_authorization_servers_auth_server_id_associated_servers_associated_server_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --associated_server_id (required), --auth_server_id (required)
    delete api v1 authorization servers auth server id claims claim id apply - Plan and execute the delete api v1 authorization servers auth server id claims claim id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_authorization_servers_auth_server_id_claims_claim_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --auth_server_id (required), --claim_id (required)
    delete api v1 authorization servers auth server id clients client id tokens apply - Plan and execute the delete api v1 authorization servers auth server id clients client id tokens reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_authorization_servers_auth_server_id_clients_client_id_tokens]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --auth_server_id (required), --client_id (required)
    delete api v1 authorization servers auth server id clients client id tokens token id apply - Plan and execute the delete api v1 authorization servers auth server id clients client id tokens token id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_authorization_servers_auth_server_id_clients_client_id_tokens_token_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --auth_server_id (required), --client_id (required), --token_id (required)
    delete api v1 authorization servers auth server id policies policy id apply - Plan and execute the delete api v1 authorization servers auth server id policies policy id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_authorization_servers_auth_server_id_policies_policy_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --auth_server_id (required), --policy_id (required)
    delete api v1 authorization servers auth server id policies policy id rules rule id apply - Plan and execute the delete api v1 authorization servers auth server id policies policy id rules rule id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --auth_server_id (required), --policy_id (required), --rule_id (required)
    delete api v1 authorization servers auth server id resourceservercredentials keys key id apply - Plan and execute the delete api v1 authorization servers auth server id resourceservercredentials keys key id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --auth_server_id (required), --key_id (required)
    delete api v1 authorization servers auth server id scopes scope id apply - Plan and execute the delete api v1 authorization servers auth server id scopes scope id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_authorization_servers_auth_server_id_scopes_scope_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --auth_server_id (required), --scope_id (required)
    delete api v1 behaviors behavior id apply - Plan and execute the delete api v1 behaviors behavior id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_behaviors_behavior_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --behavior_id (required)
    delete api v1 brands brand id apply - Plan and execute the delete api v1 brands brand id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_brands_brand_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --brand_id (required)
    delete api v1 brands brand id pages error customized apply - Plan and execute the delete api v1 brands brand id pages error customized reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_brands_brand_id_pages_error_customized]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --brand_id (required)
    delete api v1 brands brand id pages error preview apply - Plan and execute the delete api v1 brands brand id pages error preview reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_brands_brand_id_pages_error_preview]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --brand_id (required)
    delete api v1 brands brand id pages sign in customized apply - Plan and execute the delete api v1 brands brand id pages sign in customized reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_brands_brand_id_pages_sign_in_customized]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --brand_id (required)
    delete api v1 brands brand id pages sign in preview apply - Plan and execute the delete api v1 brands brand id pages sign in preview reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_brands_brand_id_pages_sign_in_preview]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --brand_id (required)
    delete api v1 brands brand id templates email template name customizations apply - Plan and execute the delete api v1 brands brand id templates email template name customizations reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_brands_brand_id_templates_email_template_name_customizations]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --brand_id (required), --template_name (required)
    delete api v1 brands brand id templates email template name customizations customization id apply - Plan and execute the delete api v1 brands brand id templates email template name customizations customization id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --brand_id (required), --customization_id (required), --template_name (required)
    delete api v1 brands brand id themes theme id background image apply - Plan and execute the delete api v1 brands brand id themes theme id background image reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_brands_brand_id_themes_theme_id_background_image]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --brand_id (required), --theme_id (required)
    delete api v1 brands brand id themes theme id favicon apply - Plan and execute the delete api v1 brands brand id themes theme id favicon reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_brands_brand_id_themes_theme_id_favicon]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --brand_id (required), --theme_id (required)
    delete api v1 brands brand id themes theme id logo apply - Plan and execute the delete api v1 brands brand id themes theme id logo reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_brands_brand_id_themes_theme_id_logo]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --brand_id (required), --theme_id (required)
    delete api v1 captchas captcha id apply - Plan and execute the delete api v1 captchas captcha id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_captchas_captcha_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --captcha_id (required)
    delete api v1 device assurances device assurance id apply - Plan and execute the delete api v1 device assurances device assurance id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_device_assurances_device_assurance_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --device_assurance_id (required)
    delete api v1 device posture checks posture check id apply - Plan and execute the delete api v1 device posture checks posture check id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_device_posture_checks_posture_check_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --posture_check_id (required)
    delete api v1 devices device id apply - Plan and execute the delete api v1 devices device id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_devices_device_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --device_id (required)
    delete api v1 domains domain id apply - Plan and execute the delete api v1 domains domain id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_domains_domain_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --domain_id (required)
    delete api v1 email domains email domain id apply - Plan and execute the delete api v1 email domains email domain id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_email_domains_email_domain_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --email_domain_id (required)
    delete api v1 email servers email server id apply - Plan and execute the delete api v1 email servers email server id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_email_servers_email_server_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --email_server_id (required)
    delete api v1 event hooks event hook id apply - Plan and execute the delete api v1 event hooks event hook id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_event_hooks_event_hook_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --event_hook_id (required)
    delete api v1 groups group id apply - Plan and execute the delete api v1 groups group id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_groups_group_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --group_id (required)
    delete api v1 groups group id owners owner id apply - Plan and execute the delete api v1 groups group id owners owner id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_groups_group_id_owners_owner_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --group_id (required), --owner_id (required)
    delete api v1 groups group id roles role assignment id apply - Plan and execute the delete api v1 groups group id roles role assignment id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_groups_group_id_roles_role_assignment_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --group_id (required), --role_assignment_id (required)
    delete api v1 groups group id roles role assignment id targets catalog apps app name app id apply - Plan and execute the delete api v1 groups group id roles role assignment id targets catalog apps app name app id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required), --app_name (required), --group_id (required), --role_assignment_id (required)
    delete api v1 groups group id roles role assignment id targets catalog apps app name apply - Plan and execute the delete api v1 groups group id roles role assignment id targets catalog apps app name reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_name (required), --group_id (required), --role_assignment_id (required)
    delete api v1 groups group id roles role assignment id targets groups target group id apply - Plan and execute the delete api v1 groups group id roles role assignment id targets groups target group id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_groups_group_id_roles_role_assignment_id_targets_groups_target_group_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --group_id (required), --role_assignment_id (required), --target_group_id (required)
    delete api v1 groups group id users user id apply - Plan and execute the delete api v1 groups group id users user id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_groups_group_id_users_user_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --group_id (required), --user_id (required)
    delete api v1 groups rules group rule id apply - Plan and execute the delete api v1 groups rules group rule id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_groups_rules_group_rule_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --group_rule_id (required)
    delete api v1 hook keys id apply - Plan and execute the delete api v1 hook keys id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_hook_keys_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --id (required)
    delete api v1 iam governance bundles bundle id apply - Plan and execute the delete api v1 iam governance bundles bundle id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_iam_governance_bundles_bundle_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --bundle_id (required)
    delete api v1 iam resource sets resource set id or label apply - Plan and execute the delete api v1 iam resource sets resource set id or label reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_iam_resource_sets_resource_set_id_or_label]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --resource_set_id_or_label (required)
    delete api v1 iam resource sets resource set id or label bindings role id or label apply - Plan and execute the delete api v1 iam resource sets resource set id or label bindings role id or label reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --resource_set_id_or_label (required), --role_id_or_label (required)
    delete api v1 iam resource sets resource set id or label bindings role id or label members member id apply - Plan and execute the delete api v1 iam resource sets resource set id or label bindings role id or label members member id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members_member_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --member_id (required), --resource_set_id_or_label (required), --role_id_or_label (required)
    delete api v1 iam resource sets resource set id or label resources resource id apply - Plan and execute the delete api v1 iam resource sets resource set id or label resources resource id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_iam_resource_sets_resource_set_id_or_label_resources_resource_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --resource_id (required), --resource_set_id_or_label (required)
    delete api v1 iam roles role id or label apply - Plan and execute the delete api v1 iam roles role id or label reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_iam_roles_role_id_or_label]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --role_id_or_label (required)
    delete api v1 iam roles role id or label permissions permission type apply - Plan and execute the delete api v1 iam roles role id or label permissions permission type reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_iam_roles_role_id_or_label_permissions_permission_type]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --permission_type (required), --role_id_or_label (required)
    delete api v1 identity sources identity source id groups group or external id apply - Plan and execute the delete api v1 identity sources identity source id groups group or external id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_identity_sources_identity_source_id_groups_group_or_external_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --group_or_external_id (required), --identity_source_id (required)
    delete api v1 identity sources identity source id groups group or external id membership member external id apply - Plan and execute the delete api v1 identity sources identity source id groups group or external id membership member external id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_identity_sources_identity_source_id_groups_group_or_external_id_membership_member_external_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --group_or_external_id (required), --identity_source_id (required), --member_external_id (required)
    delete api v1 identity sources identity source id sessions session id apply - Plan and execute the delete api v1 identity sources identity source id sessions session id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_identity_sources_identity_source_id_sessions_session_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --identity_source_id (required), --session_id (required)
    delete api v1 identity sources identity source id users external id apply - Plan and execute the delete api v1 identity sources identity source id users external id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_identity_sources_identity_source_id_users_external_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --external_id (required), --identity_source_id (required)
    delete api v1 idps credentials keys kid apply - Plan and execute the delete api v1 idps credentials keys kid reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_idps_credentials_keys_kid]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --kid (required)
    delete api v1 idps idp id apply - Plan and execute the delete api v1 idps idp id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_idps_idp_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --idp_id (required)
    delete api v1 idps idp id credentials csrs idp csr id apply - Plan and execute the delete api v1 idps idp id credentials csrs idp csr id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_idps_idp_id_credentials_csrs_idp_csr_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --idp_csr_id (required), --idp_id (required)
    delete api v1 idps idp id users user id apply - Plan and execute the delete api v1 idps idp id users user id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_idps_idp_id_users_user_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --idp_id (required), --user_id (required)
    delete api v1 inline hooks inline hook id apply - Plan and execute the delete api v1 inline hooks inline hook id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_inline_hooks_inline_hook_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --inline_hook_id (required)
    delete api v1 log streams log stream id apply - Plan and execute the delete api v1 log streams log stream id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_log_streams_log_stream_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --log_stream_id (required)
    delete api v1 meta schemas user linked objects linked object name apply - Plan and execute the delete api v1 meta schemas user linked objects linked object name reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_meta_schemas_user_linked_objects_linked_object_name]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --linked_object_name (required)
    delete api v1 meta types user type id apply - Plan and execute the delete api v1 meta types user type id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_meta_types_user_type_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --type_id (required)
    delete api v1 meta uischemas id apply - Plan and execute the delete api v1 meta uischemas id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_meta_uischemas_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --id (required)
    delete api v1 org captcha apply - Plan and execute the delete api v1 org captcha reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_org_captcha]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required
    delete api v1 policies policy id apply - Plan and execute the delete api v1 policies policy id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_policies_policy_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --policy_id (required)
    delete api v1 policies policy id mappings mapping id apply - Plan and execute the delete api v1 policies policy id mappings mapping id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_policies_policy_id_mappings_mapping_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --mapping_id (required), --policy_id (required)
    delete api v1 policies policy id rules rule id apply - Plan and execute the delete api v1 policies policy id rules rule id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_policies_policy_id_rules_rule_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --policy_id (required), --rule_id (required)
    delete api v1 push providers push provider id apply - Plan and execute the delete api v1 push providers push provider id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_push_providers_push_provider_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --push_provider_id (required)
    delete api v1 realm assignments assignment id apply - Plan and execute the delete api v1 realm assignments assignment id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_realm_assignments_assignment_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --assignment_id (required)
    delete api v1 realms realm id apply - Plan and execute the delete api v1 realms realm id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_realms_realm_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --realm_id (required)
    delete api v1 security events providers security event provider id apply - Plan and execute the delete api v1 security events providers security event provider id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_security_events_providers_security_event_provider_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --security_event_provider_id (required)
    delete api v1 sessions session id apply - Plan and execute the delete api v1 sessions session id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_sessions_session_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --session_id (required)
    delete api v1 ssf stream apply - Plan and execute the delete api v1 ssf stream reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_ssf_stream]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required
    delete api v1 telephony providers custom telephony provider id apply - Plan and execute the delete api v1 telephony providers custom telephony provider id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_telephony_providers_custom_telephony_provider_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --custom_telephony_provider_id (required)
    delete api v1 templates sms template id apply - Plan and execute the delete api v1 templates sms template id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_templates_sms_template_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --template_id (required)
    delete api v1 trusted origins trusted origin id apply - Plan and execute the delete api v1 trusted origins trusted origin id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_trusted_origins_trusted_origin_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --trusted_origin_id (required)
    delete api v1 users id apply - Plan and execute the delete api v1 users id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_users_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --id (required)
    delete api v1 users user id authenticator enrollments enrollment id apply - Plan and execute the delete api v1 users user id authenticator enrollments enrollment id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_users_user_id_authenticator_enrollments_enrollment_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --enrollment_id (required), --user_id (required)
    delete api v1 users user id clients client id grants apply - Plan and execute the delete api v1 users user id clients client id grants reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_users_user_id_clients_client_id_grants]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --client_id (required), --user_id (required)
    delete api v1 users user id clients client id tokens apply - Plan and execute the delete api v1 users user id clients client id tokens reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_users_user_id_clients_client_id_tokens]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --client_id (required), --user_id (required)
    delete api v1 users user id clients client id tokens token id apply - Plan and execute the delete api v1 users user id clients client id tokens token id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_users_user_id_clients_client_id_tokens_token_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --client_id (required), --token_id (required), --user_id (required)
    delete api v1 users user id factors factor id apply - Plan and execute the delete api v1 users user id factors factor id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_users_user_id_factors_factor_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --factor_id (required), --user_id (required)
    delete api v1 users user id grants apply - Plan and execute the delete api v1 users user id grants reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_users_user_id_grants]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --user_id (required)
    delete api v1 users user id grants grant id apply - Plan and execute the delete api v1 users user id grants grant id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_users_user_id_grants_grant_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --grant_id (required), --user_id (required)
    delete api v1 users user id or login linked objects relationship name apply - Plan and execute the delete api v1 users user id or login linked objects relationship name reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_users_user_id_or_login_linked_objects_relationship_name]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --relationship_name (required), --user_id_or_login (required)
    delete api v1 users user id roles role assignment id apply - Plan and execute the delete api v1 users user id roles role assignment id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_users_user_id_roles_role_assignment_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --role_assignment_id (required), --user_id (required)
    delete api v1 users user id roles role assignment id targets catalog apps app name app id apply - Plan and execute the delete api v1 users user id roles role assignment id targets catalog apps app name app id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required), --app_name (required), --role_assignment_id (required), --user_id (required)
    delete api v1 users user id roles role assignment id targets catalog apps app name apply - Plan and execute the delete api v1 users user id roles role assignment id targets catalog apps app name reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_name (required), --role_assignment_id (required), --user_id (required)
    delete api v1 users user id roles role assignment id targets groups group id apply - Plan and execute the delete api v1 users user id roles role assignment id targets groups group id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_users_user_id_roles_role_assignment_id_targets_groups_group_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --group_id (required), --role_assignment_id (required), --user_id (required)
    delete api v1 users user id sessions apply - Plan and execute the delete api v1 users user id sessions reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_users_user_id_sessions]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --user_id (required)
    delete api v1 zones zone id apply - Plan and execute the delete api v1 zones zone id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_api_v1_zones_zone_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --zone_id (required)
    delete integrations api v1 api services api service id apply - Plan and execute the delete integrations api v1 api services api service id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_integrations_api_v1_api_services_api_service_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --api_service_id (required)
    delete integrations api v1 api services api service id credentials secrets secret id apply - Plan and execute the delete integrations api v1 api services api service id credentials secrets secret id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_integrations_api_v1_api_services_api_service_id_credentials_secrets_secret_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --api_service_id (required), --secret_id (required)
    delete oauth2 v1 clients client id roles role assignment id apply - Plan and execute the delete oauth2 v1 clients client id roles role assignment id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_oauth2_v1_clients_client_id_roles_role_assignment_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --client_id (required), --role_assignment_id (required)
    delete oauth2 v1 clients client id roles role assignment id targets catalog apps app name app id apply - Plan and execute the delete oauth2 v1 clients client id roles role assignment id targets catalog apps app name app id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required), --app_name (required), --client_id (required), --role_assignment_id (required)
    delete oauth2 v1 clients client id roles role assignment id targets catalog apps app name apply - Plan and execute the delete oauth2 v1 clients client id roles role assignment id targets catalog apps app name reverse-ETL action [intent=reverse_etl availability=implemented write=delete_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_name (required), --client_id (required), --role_assignment_id (required)
    delete oauth2 v1 clients client id roles role assignment id targets groups group id apply - Plan and execute the delete oauth2 v1 clients client id roles role assignment id targets groups group id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_groups_group_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --client_id (required), --group_id (required), --role_assignment_id (required)
    delete privileged access api v1 okta service accounts id apply - Plan and execute the delete privileged access api v1 okta service accounts id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_privileged_access_api_v1_okta_service_accounts_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --id (required)
    delete privileged access api v1 service accounts id apply - Plan and execute the delete privileged access api v1 service accounts id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_privileged_access_api_v1_service_accounts_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --id (required)
    delete webauthn registration api v1 users user id enrollments authenticator enrollment id apply - Plan and execute the delete webauthn registration api v1 users user id enrollments authenticator enrollment id reverse-ETL action [intent=reverse_etl availability=implemented write=delete_webauthn_registration_api_v1_users_user_id_enrollments_authenticator_enrollment_id]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --authenticator_enrollment_id (required), --user_id (required)
    execute api v1 agent pools pool id updates update id activate apply - Plan and execute the execute api v1 agent pools pool id updates update id activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_agent_pools_pool_id_updates_update_id_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --pool_id (required), --update_id (required)
    execute api v1 agent pools pool id updates update id deactivate apply - Plan and execute the execute api v1 agent pools pool id updates update id deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_agent_pools_pool_id_updates_update_id_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --pool_id (required), --update_id (required)
    execute api v1 agent pools pool id updates update id pause apply - Plan and execute the execute api v1 agent pools pool id updates update id pause reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_agent_pools_pool_id_updates_update_id_pause]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --pool_id (required), --update_id (required)
    execute api v1 agent pools pool id updates update id resume apply - Plan and execute the execute api v1 agent pools pool id updates update id resume reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_agent_pools_pool_id_updates_update_id_resume]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --pool_id (required), --update_id (required)
    execute api v1 agent pools pool id updates update id retry apply - Plan and execute the execute api v1 agent pools pool id updates update id retry reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_agent_pools_pool_id_updates_update_id_retry]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --pool_id (required), --update_id (required)
    execute api v1 agent pools pool id updates update id stop apply - Plan and execute the execute api v1 agent pools pool id updates update id stop reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_agent_pools_pool_id_updates_update_id_stop]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --pool_id (required), --update_id (required)
    execute api v1 apps app id connections default lifecycle activate apply - Plan and execute the execute api v1 apps app id connections default lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_apps_app_id_connections_default_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required)
    execute api v1 apps app id connections default lifecycle deactivate apply - Plan and execute the execute api v1 apps app id connections default lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_apps_app_id_connections_default_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required)
    execute api v1 apps app id credentials jwks key id lifecycle activate apply - Plan and execute the execute api v1 apps app id credentials jwks key id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_apps_app_id_credentials_jwks_key_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required), --key_id (required)
    execute api v1 apps app id credentials jwks key id lifecycle deactivate apply - Plan and execute the execute api v1 apps app id credentials jwks key id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_apps_app_id_credentials_jwks_key_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required), --key_id (required)
    execute api v1 apps app id credentials secrets secret id lifecycle activate apply - Plan and execute the execute api v1 apps app id credentials secrets secret id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_apps_app_id_credentials_secrets_secret_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required), --secret_id (required)
    execute api v1 apps app id credentials secrets secret id lifecycle deactivate apply - Plan and execute the execute api v1 apps app id credentials secrets secret id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_apps_app_id_credentials_secrets_secret_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required), --secret_id (required)
    execute api v1 apps app id lifecycle activate apply - Plan and execute the execute api v1 apps app id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_apps_app_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required)
    execute api v1 apps app id lifecycle deactivate apply - Plan and execute the execute api v1 apps app id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_apps_app_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --app_id (required)
    execute api v1 authenticators authenticator id lifecycle activate apply - Plan and execute the execute api v1 authenticators authenticator id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_authenticators_authenticator_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --authenticator_id (required)
    execute api v1 authenticators authenticator id lifecycle deactivate apply - Plan and execute the execute api v1 authenticators authenticator id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_authenticators_authenticator_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --authenticator_id (required)
    execute api v1 authenticators authenticator id methods method type lifecycle activate apply - Plan and execute the execute api v1 authenticators authenticator id methods method type lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_authenticators_authenticator_id_methods_method_type_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --authenticator_id (required), --method_type (required)
    execute api v1 authenticators authenticator id methods method type lifecycle deactivate apply - Plan and execute the execute api v1 authenticators authenticator id methods method type lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_authenticators_authenticator_id_methods_method_type_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --authenticator_id (required), --method_type (required)
    execute api v1 authorization servers auth server id credentials lifecycle key rotate apply - Plan and execute the execute api v1 authorization servers auth server id credentials lifecycle key rotate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_authorization_servers_auth_server_id_credentials_lifecycle_key_rotate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --auth_server_id (required)
    execute api v1 authorization servers auth server id lifecycle activate apply - Plan and execute the execute api v1 authorization servers auth server id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_authorization_servers_auth_server_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --auth_server_id (required)
    execute api v1 authorization servers auth server id lifecycle deactivate apply - Plan and execute the execute api v1 authorization servers auth server id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_authorization_servers_auth_server_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --auth_server_id (required)
    execute api v1 authorization servers auth server id policies policy id lifecycle activate apply - Plan and execute the execute api v1 authorization servers auth server id policies policy id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --auth_server_id (required), --policy_id (required)
    execute api v1 authorization servers auth server id policies policy id lifecycle deactivate apply - Plan and execute the execute api v1 authorization servers auth server id policies policy id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --auth_server_id (required), --policy_id (required)
    execute api v1 authorization servers auth server id policies policy id rules rule id lifecycle activate apply - Plan and execute the execute api v1 authorization servers auth server id policies policy id rules rule id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --auth_server_id (required), --policy_id (required), --rule_id (required)
    execute api v1 authorization servers auth server id policies policy id rules rule id lifecycle deactivate apply - Plan and execute the execute api v1 authorization servers auth server id policies policy id rules rule id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --auth_server_id (required), --policy_id (required), --rule_id (required)
    execute api v1 authorization servers auth server id resourceservercredentials keys key id lifecycle activate apply - Plan and execute the execute api v1 authorization servers auth server id resourceservercredentials keys key id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --auth_server_id (required), --key_id (required)
    execute api v1 authorization servers auth server id resourceservercredentials keys key id lifecycle deactivate apply - Plan and execute the execute api v1 authorization servers auth server id resourceservercredentials keys key id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --auth_server_id (required), --key_id (required)
    execute api v1 behaviors behavior id lifecycle activate apply - Plan and execute the execute api v1 behaviors behavior id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_behaviors_behavior_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --behavior_id (required)
    execute api v1 behaviors behavior id lifecycle deactivate apply - Plan and execute the execute api v1 behaviors behavior id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_behaviors_behavior_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --behavior_id (required)
    execute api v1 device integrations device integration id lifecycle activate apply - Plan and execute the execute api v1 device integrations device integration id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_device_integrations_device_integration_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --device_integration_id (required)
    execute api v1 device integrations device integration id lifecycle deactivate apply - Plan and execute the execute api v1 device integrations device integration id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_device_integrations_device_integration_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --device_integration_id (required)
    execute api v1 devices device id lifecycle activate apply - Plan and execute the execute api v1 devices device id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_devices_device_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --device_id (required)
    execute api v1 devices device id lifecycle deactivate apply - Plan and execute the execute api v1 devices device id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_devices_device_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --device_id (required)
    execute api v1 devices device id lifecycle suspend apply - Plan and execute the execute api v1 devices device id lifecycle suspend reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_devices_device_id_lifecycle_suspend]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --device_id (required)
    execute api v1 devices device id lifecycle unsuspend apply - Plan and execute the execute api v1 devices device id lifecycle unsuspend reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_devices_device_id_lifecycle_unsuspend]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --device_id (required)
    execute api v1 event hooks event hook id lifecycle activate apply - Plan and execute the execute api v1 event hooks event hook id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_event_hooks_event_hook_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --event_hook_id (required)
    execute api v1 event hooks event hook id lifecycle deactivate apply - Plan and execute the execute api v1 event hooks event hook id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_event_hooks_event_hook_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --event_hook_id (required)
    execute api v1 event hooks event hook id lifecycle verify apply - Plan and execute the execute api v1 event hooks event hook id lifecycle verify reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_event_hooks_event_hook_id_lifecycle_verify]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --event_hook_id (required)
    execute api v1 groups rules group rule id lifecycle activate apply - Plan and execute the execute api v1 groups rules group rule id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_groups_rules_group_rule_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --group_rule_id (required)
    execute api v1 groups rules group rule id lifecycle deactivate apply - Plan and execute the execute api v1 groups rules group rule id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_groups_rules_group_rule_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --group_rule_id (required)
    execute api v1 idps idp id lifecycle activate apply - Plan and execute the execute api v1 idps idp id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_idps_idp_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --idp_id (required)
    execute api v1 idps idp id lifecycle deactivate apply - Plan and execute the execute api v1 idps idp id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_idps_idp_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --idp_id (required)
    execute api v1 inline hooks inline hook id lifecycle activate apply - Plan and execute the execute api v1 inline hooks inline hook id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_inline_hooks_inline_hook_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --inline_hook_id (required)
    execute api v1 inline hooks inline hook id lifecycle deactivate apply - Plan and execute the execute api v1 inline hooks inline hook id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_inline_hooks_inline_hook_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --inline_hook_id (required)
    execute api v1 log streams log stream id lifecycle activate apply - Plan and execute the execute api v1 log streams log stream id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_log_streams_log_stream_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --log_stream_id (required)
    execute api v1 log streams log stream id lifecycle deactivate apply - Plan and execute the execute api v1 log streams log stream id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_log_streams_log_stream_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --log_stream_id (required)
    execute api v1 org privacy aerial revoke apply - Plan and execute the execute api v1 org privacy aerial revoke reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_org_privacy_aerial_revoke]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --accountId (required)
    execute api v1 org privacy okta support revoke apply - Plan and execute the execute api v1 org privacy okta support revoke reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_org_privacy_okta_support_revoke]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required
    execute api v1 policies policy id lifecycle activate apply - Plan and execute the execute api v1 policies policy id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_policies_policy_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --policy_id (required)
    execute api v1 policies policy id lifecycle deactivate apply - Plan and execute the execute api v1 policies policy id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_policies_policy_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --policy_id (required)
    execute api v1 policies policy id rules rule id lifecycle activate apply - Plan and execute the execute api v1 policies policy id rules rule id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_policies_policy_id_rules_rule_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --policy_id (required), --rule_id (required)
    execute api v1 policies policy id rules rule id lifecycle deactivate apply - Plan and execute the execute api v1 policies policy id rules rule id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_policies_policy_id_rules_rule_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --policy_id (required), --rule_id (required)
    execute api v1 realm assignments assignment id lifecycle activate apply - Plan and execute the execute api v1 realm assignments assignment id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_realm_assignments_assignment_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --assignment_id (required)
    execute api v1 realm assignments assignment id lifecycle deactivate apply - Plan and execute the execute api v1 realm assignments assignment id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_realm_assignments_assignment_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --assignment_id (required)
    execute api v1 security events providers security event provider id lifecycle activate apply - Plan and execute the execute api v1 security events providers security event provider id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_security_events_providers_security_event_provider_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --security_event_provider_id (required)
    execute api v1 security events providers security event provider id lifecycle deactivate apply - Plan and execute the execute api v1 security events providers security event provider id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_security_events_providers_security_event_provider_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --security_event_provider_id (required)
    execute api v1 sessions session id lifecycle refresh apply - Plan and execute the execute api v1 sessions session id lifecycle refresh reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_sessions_session_id_lifecycle_refresh]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --session_id (required)
    execute api v1 telephony providers custom telephony provider id lifecycle activate apply - Plan and execute the execute api v1 telephony providers custom telephony provider id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_telephony_providers_custom_telephony_provider_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --custom_telephony_provider_id (required)
    execute api v1 telephony providers custom telephony provider id lifecycle deactivate apply - Plan and execute the execute api v1 telephony providers custom telephony provider id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_telephony_providers_custom_telephony_provider_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --custom_telephony_provider_id (required)
    execute api v1 trusted origins trusted origin id lifecycle activate apply - Plan and execute the execute api v1 trusted origins trusted origin id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_trusted_origins_trusted_origin_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --trusted_origin_id (required)
    execute api v1 trusted origins trusted origin id lifecycle deactivate apply - Plan and execute the execute api v1 trusted origins trusted origin id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_trusted_origins_trusted_origin_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --trusted_origin_id (required)
    execute api v1 users id lifecycle activate apply - Plan and execute the execute api v1 users id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_users_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --id (required)
    execute api v1 users id lifecycle deactivate apply - Plan and execute the execute api v1 users id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_users_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --id (required)
    execute api v1 users id lifecycle expire password apply - Plan and execute the execute api v1 users id lifecycle expire password reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_users_id_lifecycle_expire_password]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --id (required)
    execute api v1 users id lifecycle expire password with temp password apply - Plan and execute the execute api v1 users id lifecycle expire password with temp password reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_users_id_lifecycle_expire_password_with_temp_password]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --id (required)
    execute api v1 users id lifecycle reactivate apply - Plan and execute the execute api v1 users id lifecycle reactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_users_id_lifecycle_reactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --id (required)
    execute api v1 users id lifecycle reset factors apply - Plan and execute the execute api v1 users id lifecycle reset factors reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_users_id_lifecycle_reset_factors]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --id (required)
    execute api v1 users id lifecycle suspend apply - Plan and execute the execute api v1 users id lifecycle suspend reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_users_id_lifecycle_suspend]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --id (required)
    execute api v1 users id lifecycle unlock apply - Plan and execute the execute api v1 users id lifecycle unlock reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_users_id_lifecycle_unlock]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --id (required)
    execute api v1 users id lifecycle unsuspend apply - Plan and execute the execute api v1 users id lifecycle unsuspend reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_users_id_lifecycle_unsuspend]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --id (required)
    execute api v1 users user id factors factor id lifecycle activate apply - Plan and execute the execute api v1 users user id factors factor id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_users_user_id_factors_factor_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --factor_id (required), --user_id (required)
    execute api v1 zones zone id lifecycle activate apply - Plan and execute the execute api v1 zones zone id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_zones_zone_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --zone_id (required)
    execute api v1 zones zone id lifecycle deactivate apply - Plan and execute the execute api v1 zones zone id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_api_v1_zones_zone_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --zone_id (required)
    execute integrations api v1 api services api service id credentials secrets secret id lifecycle activate apply - Plan and execute the execute integrations api v1 api services api service id credentials secrets secret id lifecycle activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_integrations_api_v1_api_services_api_service_id_credentials_secrets_secret_id_lifecycle_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --api_service_id (required), --secret_id (required)
    execute integrations api v1 api services api service id credentials secrets secret id lifecycle deactivate apply - Plan and execute the execute integrations api v1 api services api service id credentials secrets secret id lifecycle deactivate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_integrations_api_v1_api_services_api_service_id_credentials_secrets_secret_id_lifecycle_deactivate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required; flags: --api_service_id (required), --secret_id (required)
    execute webauthn registration api v1 activate apply - Plan and execute the execute webauthn registration api v1 activate reverse-ETL action [intent=reverse_etl availability=implemented write=execute_webauthn_registration_api_v1_activate]; approval: requires plan, preview, approval, and execute; risk: high: external Okta admin API mutation; approval required
    groups list - Run the groups ETL stream [intent=etl availability=implemented stream=groups]
    integrations api v1 api services api service id credentials secrets list - Run the integrations api v1 api services api service id credentials secrets ETL stream [intent=etl availability=implemented stream=integrations_api_v1_api_services_api_service_id_credentials_secrets]
    integrations api v1 api services api service id list - Run the integrations api v1 api services api service id ETL stream [intent=etl availability=implemented stream=integrations_api_v1_api_services_api_service_id]
    integrations api v1 api services list - Run the integrations api v1 api services ETL stream [intent=etl availability=implemented stream=integrations_api_v1_api_services]
    oauth2 v1 clients client id roles list - Run the oauth2 v1 clients client id roles ETL stream [intent=etl availability=implemented stream=oauth2_v1_clients_client_id_roles]
    oauth2 v1 clients client id roles role assignment id list - Run the oauth2 v1 clients client id roles role assignment id ETL stream [intent=etl availability=implemented stream=oauth2_v1_clients_client_id_roles_role_assignment_id]
    oauth2 v1 clients client id roles role assignment id targets catalog apps list - Run the oauth2 v1 clients client id roles role assignment id targets catalog apps ETL stream [intent=etl availability=implemented stream=oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps]
    oauth2 v1 clients client id roles role assignment id targets groups list - Run the oauth2 v1 clients client id roles role assignment id targets groups ETL stream [intent=etl availability=implemented stream=oauth2_v1_clients_client_id_roles_role_assignment_id_targets_groups]
    okta personal settings api v1 export blocklists list - Run the okta personal settings api v1 export blocklists ETL stream [intent=etl availability=implemented stream=okta_personal_settings_api_v1_export_blocklists]
    privileged access api v1 okta service accounts id list - Run the privileged access api v1 okta service accounts id ETL stream [intent=etl availability=implemented stream=privileged_access_api_v1_okta_service_accounts_id]
    privileged access api v1 okta service accounts list - Run the privileged access api v1 okta service accounts ETL stream [intent=etl availability=implemented stream=privileged_access_api_v1_okta_service_accounts]
    privileged access api v1 service accounts id list - Run the privileged access api v1 service accounts id ETL stream [intent=etl availability=implemented stream=privileged_access_api_v1_service_accounts_id]
    privileged access api v1 service accounts list - Run the privileged access api v1 service accounts ETL stream [intent=etl availability=implemented stream=privileged_access_api_v1_service_accounts]
    system logs list - Run the system logs ETL stream [intent=etl availability=implemented stream=system_logs]
    update api v1 api tokens api token id apply - Plan and execute the update api v1 api tokens api token id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_api_tokens_api_token_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --api_token_id (required)
    update api v1 apps app id apply - Plan and execute the update api v1 apps app id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_apps_app_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required), --label (required), --signOnMode (required)
    update api v1 apps app id cwo connections connection id apply - Plan and execute the update api v1 apps app id cwo connections connection id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_apps_app_id_cwo_connections_connection_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required), --connection_id (required), --status (required)
    update api v1 apps app id features feature name apply - Plan and execute the update api v1 apps app id features feature name reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_apps_app_id_features_feature_name]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required), --feature_name (required)
    update api v1 apps app id federated claims claim id apply - Plan and execute the update api v1 apps app id federated claims claim id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_apps_app_id_federated_claims_claim_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required), --claim_id (required)
    update api v1 apps app id group push mappings mapping id apply - Plan and execute the update api v1 apps app id group push mappings mapping id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_apps_app_id_group_push_mappings_mapping_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required), --mapping_id (required), --status (required)
    update api v1 apps app id groups group id apply - Plan and execute the update api v1 apps app id groups group id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_apps_app_id_groups_group_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required), --group_id (required)
    update api v1 apps app id policies policy id apply - Plan and execute the update api v1 apps app id policies policy id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_apps_app_id_policies_policy_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required), --policy_id (required)
    update api v1 authenticators authenticator id aaguids aaguid 2 apply - Plan and execute the update api v1 authenticators authenticator id aaguids aaguid 2 reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_authenticators_authenticator_id_aaguids_aaguid_2]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --aaguid (required), --authenticator_id (required)
    update api v1 authenticators authenticator id aaguids aaguid apply - Plan and execute the update api v1 authenticators authenticator id aaguids aaguid reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_authenticators_authenticator_id_aaguids_aaguid]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --aaguid (required), --authenticator_id (required)
    update api v1 authenticators authenticator id apply - Plan and execute the update api v1 authenticators authenticator id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_authenticators_authenticator_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --authenticator_id (required)
    update api v1 authenticators authenticator id methods method type apply - Plan and execute the update api v1 authenticators authenticator id methods method type reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_authenticators_authenticator_id_methods_method_type]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --authenticator_id (required), --method_type (required)
    update api v1 authorization servers auth server id apply - Plan and execute the update api v1 authorization servers auth server id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_authorization_servers_auth_server_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --auth_server_id (required)
    update api v1 authorization servers auth server id claims claim id apply - Plan and execute the update api v1 authorization servers auth server id claims claim id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_authorization_servers_auth_server_id_claims_claim_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --auth_server_id (required), --claim_id (required)
    update api v1 authorization servers auth server id policies policy id apply - Plan and execute the update api v1 authorization servers auth server id policies policy id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_authorization_servers_auth_server_id_policies_policy_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --auth_server_id (required), --policy_id (required)
    update api v1 authorization servers auth server id policies policy id rules rule id apply - Plan and execute the update api v1 authorization servers auth server id policies policy id rules rule id reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update api v1 authorization servers auth server id scopes scope id apply - Plan and execute the update api v1 authorization servers auth server id scopes scope id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_authorization_servers_auth_server_id_scopes_scope_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --auth_server_id (required), --name (required), --scope_id (required)
    update api v1 behaviors behavior id apply - Plan and execute the update api v1 behaviors behavior id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_behaviors_behavior_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --behavior_id (required), --name (required), --type (required)
    update api v1 brands brand id apply - Plan and execute the update api v1 brands brand id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_brands_brand_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --brand_id (required), --name (required)
    update api v1 brands brand id pages error customized apply - Plan and execute the update api v1 brands brand id pages error customized reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_brands_brand_id_pages_error_customized]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --brand_id (required)
    update api v1 brands brand id pages error preview apply - Plan and execute the update api v1 brands brand id pages error preview reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_brands_brand_id_pages_error_preview]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --brand_id (required)
    update api v1 brands brand id pages sign in customized apply - Plan and execute the update api v1 brands brand id pages sign in customized reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_brands_brand_id_pages_sign_in_customized]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --brand_id (required)
    update api v1 brands brand id pages sign in preview apply - Plan and execute the update api v1 brands brand id pages sign in preview reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_brands_brand_id_pages_sign_in_preview]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --brand_id (required)
    update api v1 brands brand id pages sign out customized apply - Plan and execute the update api v1 brands brand id pages sign out customized reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_brands_brand_id_pages_sign_out_customized]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --brand_id (required), --type (required)
    update api v1 brands brand id templates email template name customizations customization id apply - Plan and execute the update api v1 brands brand id templates email template name customizations customization id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --brand_id (required), --customization_id (required), --language (required), --template_name (required)
    update api v1 brands brand id templates email template name settings apply - Plan and execute the update api v1 brands brand id templates email template name settings reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_brands_brand_id_templates_email_template_name_settings]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --brand_id (required), --recipients (required), --template_name (required)
    update api v1 brands brand id themes theme id apply - Plan and execute the update api v1 brands brand id themes theme id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_brands_brand_id_themes_theme_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --brand_id (required), --emailTemplateTouchPointVariant (required), --endUserDashboardTouchPointVariant (required), --errorPageTouchPointVariant (required), --primaryColorHex (required), --secondaryColorHex (required), --signInPageTouchPointVariant (required), --theme_id (required)
    update api v1 brands brand id well known uris path customized apply - Plan and execute the update api v1 brands brand id well known uris path customized reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_api_v1_brands_brand_id_well_known_uris_path_customized]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update api v1 captchas captcha id apply - Plan and execute the update api v1 captchas captcha id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_captchas_captcha_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --captcha_id (required)
    update api v1 device assurances device assurance id apply - Plan and execute the update api v1 device assurances device assurance id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_device_assurances_device_assurance_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --device_assurance_id (required)
    update api v1 device posture checks posture check id apply - Plan and execute the update api v1 device posture checks posture check id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_device_posture_checks_posture_check_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --posture_check_id (required)
    update api v1 domains domain id apply - Plan and execute the update api v1 domains domain id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_domains_domain_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --brandId (required), --domain_id (required)
    update api v1 domains domain id certificate apply - Plan and execute the update api v1 domains domain id certificate reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_domains_domain_id_certificate]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --certificate (required), --certificateChain (required), --domain_id (required), --privateKey (required), --type (required)
    update api v1 email domains email domain id apply - Plan and execute the update api v1 email domains email domain id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_email_domains_email_domain_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --displayName (required), --email_domain_id (required), --userName (required)
    update api v1 email servers email server id apply - Plan and execute the update api v1 email servers email server id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_email_servers_email_server_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --authType (required), --email_server_id (required)
    update api v1 event hooks event hook id apply - Plan and execute the update api v1 event hooks event hook id reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_api_v1_event_hooks_event_hook_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update api v1 first party app settings app name apply - Plan and execute the update api v1 first party app settings app name reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_first_party_app_settings_app_name]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_name (required)
    update api v1 groups group id apply - Plan and execute the update api v1 groups group id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_groups_group_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --group_id (required)
    update api v1 groups group id roles role assignment id targets catalog apps app name app id apply - Plan and execute the update api v1 groups group id roles role assignment id targets catalog apps app name app id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required), --app_name (required), --group_id (required), --role_assignment_id (required)
    update api v1 groups group id roles role assignment id targets catalog apps app name apply - Plan and execute the update api v1 groups group id roles role assignment id targets catalog apps app name reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_name (required), --group_id (required), --role_assignment_id (required)
    update api v1 groups group id roles role assignment id targets groups target group id apply - Plan and execute the update api v1 groups group id roles role assignment id targets groups target group id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_groups_group_id_roles_role_assignment_id_targets_groups_target_group_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --group_id (required), --role_assignment_id (required), --target_group_id (required)
    update api v1 groups group id users user id apply - Plan and execute the update api v1 groups group id users user id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_groups_group_id_users_user_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --group_id (required), --user_id (required)
    update api v1 groups rules group rule id apply - Plan and execute the update api v1 groups rules group rule id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_groups_rules_group_rule_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --group_rule_id (required)
    update api v1 hook keys id apply - Plan and execute the update api v1 hook keys id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_hook_keys_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --id (required)
    update api v1 iam governance bundles bundle id apply - Plan and execute the update api v1 iam governance bundles bundle id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_iam_governance_bundles_bundle_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --bundle_id (required)
    update api v1 iam resource sets resource set id or label apply - Plan and execute the update api v1 iam resource sets resource set id or label reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_iam_resource_sets_resource_set_id_or_label]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --resource_set_id_or_label (required)
    update api v1 iam resource sets resource set id or label bindings role id or label members apply - Plan and execute the update api v1 iam resource sets resource set id or label bindings role id or label members reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --resource_set_id_or_label (required), --role_id_or_label (required)
    update api v1 iam resource sets resource set id or label resources apply - Plan and execute the update api v1 iam resource sets resource set id or label resources reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_iam_resource_sets_resource_set_id_or_label_resources]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --resource_set_id_or_label (required)
    update api v1 iam resource sets resource set id or label resources resource id apply - Plan and execute the update api v1 iam resource sets resource set id or label resources resource id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_iam_resource_sets_resource_set_id_or_label_resources_resource_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --resource_id (required), --resource_set_id_or_label (required)
    update api v1 iam roles role id or label apply - Plan and execute the update api v1 iam roles role id or label reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_iam_roles_role_id_or_label]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --description (required), --label (required), --role_id_or_label (required)
    update api v1 iam roles role id or label permissions permission type apply - Plan and execute the update api v1 iam roles role id or label permissions permission type reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_iam_roles_role_id_or_label_permissions_permission_type]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --permission_type (required), --role_id_or_label (required)
    update api v1 identity sources identity source id users external id 2 apply - Plan and execute the update api v1 identity sources identity source id users external id 2 reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_identity_sources_identity_source_id_users_external_id_2]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --external_id (required), --identity_source_id (required)
    update api v1 identity sources identity source id users external id apply - Plan and execute the update api v1 identity sources identity source id users external id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_identity_sources_identity_source_id_users_external_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --external_id (required), --identity_source_id (required)
    update api v1 idps credentials keys kid apply - Plan and execute the update api v1 idps credentials keys kid reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_idps_credentials_keys_kid]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --kid (required)
    update api v1 idps idp id apply - Plan and execute the update api v1 idps idp id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_idps_idp_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --idp_id (required)
    update api v1 inline hooks inline hook id apply - Plan and execute the update api v1 inline hooks inline hook id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_inline_hooks_inline_hook_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --inline_hook_id (required)
    update api v1 log streams log stream id apply - Plan and execute the update api v1 log streams log stream id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_log_streams_log_stream_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --log_stream_id (required), --name (required), --type (required)
    update api v1 meta types user type id apply - Plan and execute the update api v1 meta types user type id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_meta_types_user_type_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --description (required), --displayName (required), --name (required), --type_id (required)
    update api v1 meta uischemas id apply - Plan and execute the update api v1 meta uischemas id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_meta_uischemas_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --id (required)
    update api v1 org apply - Plan and execute the update api v1 org reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_org]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    update api v1 org captcha apply - Plan and execute the update api v1 org captcha reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_org_captcha]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    update api v1 org contacts contact type apply - Plan and execute the update api v1 org contacts contact type reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_org_contacts_contact_type]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --contact_type (required)
    update api v1 org privacy okta support cases case number apply - Plan and execute the update api v1 org privacy okta support cases case number reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_org_privacy_okta_support_cases_case_number]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --case_number (required)
    update api v1 org settings client privileges setting apply - Plan and execute the update api v1 org settings client privileges setting reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_org_settings_client_privileges_setting]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    update api v1 policies policy id apply - Plan and execute the update api v1 policies policy id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_policies_policy_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --name (required), --policy_id (required), --type (required)
    update api v1 policies policy id rules rule id apply - Plan and execute the update api v1 policies policy id rules rule id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_policies_policy_id_rules_rule_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --policy_id (required), --rule_id (required)
    update api v1 principal rate limits principal rate limit id apply - Plan and execute the update api v1 principal rate limits principal rate limit id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_principal_rate_limits_principal_rate_limit_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --principalId (required), --principalType (required), --principal_rate_limit_id (required)
    update api v1 push providers push provider id apply - Plan and execute the update api v1 push providers push provider id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_push_providers_push_provider_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --push_provider_id (required)
    update api v1 rate limit settings admin notifications apply - Plan and execute the update api v1 rate limit settings admin notifications reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_rate_limit_settings_admin_notifications]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --notificationsEnabled (required)
    update api v1 rate limit settings per client apply - Plan and execute the update api v1 rate limit settings per client reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_rate_limit_settings_per_client]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --defaultMode (required)
    update api v1 rate limit settings warning threshold apply - Plan and execute the update api v1 rate limit settings warning threshold reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_rate_limit_settings_warning_threshold]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --warningThreshold (required)
    update api v1 realm assignments assignment id apply - Plan and execute the update api v1 realm assignments assignment id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_realm_assignments_assignment_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --assignment_id (required)
    update api v1 realms realm id apply - Plan and execute the update api v1 realms realm id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_realms_realm_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --realm_id (required)
    update api v1 security events providers security event provider id apply - Plan and execute the update api v1 security events providers security event provider id reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_api_v1_security_events_providers_security_event_provider_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update api v1 ssf stream 2 apply - Plan and execute the update api v1 ssf stream 2 reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_api_v1_ssf_stream_2]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update api v1 ssf stream apply - Plan and execute the update api v1 ssf stream reverse-ETL action [intent=reverse_etl availability=not_implemented write=update_api_v1_ssf_stream]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; notes: named_dependency=engine.reverse_etl_scalar_flag_contract: the reverse-ETL command surface cannot faithfully expose this action's required object or array record fields as scalar flags
    update api v1 telephony providers custom telephony provider id apply - Plan and execute the update api v1 telephony providers custom telephony provider id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_telephony_providers_custom_telephony_provider_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --custom_telephony_provider_id (required)
    update api v1 templates sms template id apply - Plan and execute the update api v1 templates sms template id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_templates_sms_template_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --template_id (required)
    update api v1 trusted origins trusted origin id apply - Plan and execute the update api v1 trusted origins trusted origin id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_trusted_origins_trusted_origin_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --trusted_origin_id (required)
    update api v1 users id apply - Plan and execute the update api v1 users id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_users_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --id (required)
    update api v1 users user id classification apply - Plan and execute the update api v1 users user id classification reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_users_user_id_classification]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --user_id (required)
    update api v1 users user id or login linked objects primary relationship name primary user id apply - Plan and execute the update api v1 users user id or login linked objects primary relationship name primary user id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_users_user_id_or_login_linked_objects_primary_relationship_name_primary_user_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --primary_relationship_name (required), --primary_user_id (required), --user_id_or_login (required)
    update api v1 users user id risk apply - Plan and execute the update api v1 users user id risk reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_users_user_id_risk]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --riskLevel (required), --user_id (required)
    update api v1 users user id roles role assignment id targets catalog apps app name app id apply - Plan and execute the update api v1 users user id roles role assignment id targets catalog apps app name app id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required), --app_name (required), --role_assignment_id (required), --user_id (required)
    update api v1 users user id roles role assignment id targets catalog apps app name apply - Plan and execute the update api v1 users user id roles role assignment id targets catalog apps app name reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_name (required), --role_assignment_id (required), --user_id (required)
    update api v1 users user id roles role assignment id targets catalog apps apply - Plan and execute the update api v1 users user id roles role assignment id targets catalog apps reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --role_assignment_id (required), --user_id (required)
    update api v1 users user id roles role assignment id targets groups group id apply - Plan and execute the update api v1 users user id roles role assignment id targets groups group id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_users_user_id_roles_role_assignment_id_targets_groups_group_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --group_id (required), --role_assignment_id (required), --user_id (required)
    update api v1 zones zone id apply - Plan and execute the update api v1 zones zone id reverse-ETL action [intent=reverse_etl availability=implemented write=update_api_v1_zones_zone_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --name (required), --type (required), --zone_id (required)
    update attack protection api v1 authenticator settings apply - Plan and execute the update attack protection api v1 authenticator settings reverse-ETL action [intent=reverse_etl availability=implemented write=update_attack_protection_api_v1_authenticator_settings]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    update attack protection api v1 user lockout settings apply - Plan and execute the update attack protection api v1 user lockout settings reverse-ETL action [intent=reverse_etl availability=implemented write=update_attack_protection_api_v1_user_lockout_settings]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    update oauth2 v1 clients client id roles role assignment id targets catalog apps app name app id apply - Plan and execute the update oauth2 v1 clients client id roles role assignment id targets catalog apps app name app id reverse-ETL action [intent=reverse_etl availability=implemented write=update_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_id (required), --app_name (required), --client_id (required), --role_assignment_id (required)
    update oauth2 v1 clients client id roles role assignment id targets catalog apps app name apply - Plan and execute the update oauth2 v1 clients client id roles role assignment id targets catalog apps app name reverse-ETL action [intent=reverse_etl availability=implemented write=update_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --app_name (required), --client_id (required), --role_assignment_id (required)
    update oauth2 v1 clients client id roles role assignment id targets groups group id apply - Plan and execute the update oauth2 v1 clients client id roles role assignment id targets groups group id reverse-ETL action [intent=reverse_etl availability=implemented write=update_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_groups_group_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --client_id (required), --group_id (required), --role_assignment_id (required)
    update okta personal settings api v1 edit feature apply - Plan and execute the update okta personal settings api v1 edit feature reverse-ETL action [intent=reverse_etl availability=implemented write=update_okta_personal_settings_api_v1_edit_feature]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    update okta personal settings api v1 export blocklists apply - Plan and execute the update okta personal settings api v1 export blocklists reverse-ETL action [intent=reverse_etl availability=implemented write=update_okta_personal_settings_api_v1_export_blocklists]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required
    update privileged access api v1 okta service accounts id apply - Plan and execute the update privileged access api v1 okta service accounts id reverse-ETL action [intent=reverse_etl availability=implemented write=update_privileged_access_api_v1_okta_service_accounts_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --id (required)
    update privileged access api v1 service accounts id apply - Plan and execute the update privileged access api v1 service accounts id reverse-ETL action [intent=reverse_etl availability=implemented write=update_privileged_access_api_v1_service_accounts_id]; approval: requires plan, preview, approval, and execute; risk: medium: external Okta admin API mutation; approval required; flags: --id (required)
    users list - Run the users ETL stream [intent=etl availability=implemented stream=users]
    webauthn registration api v1 users user id enrollments list - Run the webauthn registration api v1 users user id enrollments ETL stream [intent=etl availability=implemented stream=webauthn_registration_api_v1_users_user_id_enrollments]
    well known app authenticator configuration list - Run the well known app authenticator configuration ETL stream [intent=etl availability=implemented stream=well_known_app_authenticator_configuration]
    well known apple app site association list - Run the well known apple app site association ETL stream [intent=etl availability=implemented stream=well_known_apple_app_site_association]
    well known assetlinks json list - Run the well known assetlinks json ETL stream [intent=etl availability=implemented stream=well_known_assetlinks_json]
    well known okta organization list - Run the well known okta organization ETL stream [intent=etl availability=implemented stream=well_known_okta_organization]
    well known ssf configuration list - Run the well known ssf configuration ETL stream [intent=etl availability=implemented stream=well_known_ssf_configuration]
    well known webauthn list - Run the well known webauthn ETL stream [intent=etl availability=implemented stream=well_known_webauthn]

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
