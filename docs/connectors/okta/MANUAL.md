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
  base_url (required)
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
  Run Okta's declared typed write actions.
  Usage: pm okta <command> [flags]
  Global flags:
    --approval-token-stdin (boolean): Read the approval token as one bounded line from standard input.
  Reverse ETL writes
  Other Commands
    create api v1 agent pools pool id updates apply - Typed action create_api_v1_agent_pools_pool_id_updates [intent=reverse_etl availability=partial write=create_api_v1_agent_pools_pool_id_updates]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates.; flags: --pool-id (required)
    create api v1 agent pools pool id updates settings apply - Typed action create_api_v1_agent_pools_pool_id_updates_settings [intent=reverse_etl availability=partial write=create_api_v1_agent_pools_pool_id_updates_settings]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates/settings disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates/settings.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates/settings disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates/settings.; flags: --agent-type (required), --pool-id (required)
    create api v1 agent pools pool id updates update id apply - Typed action create_api_v1_agent_pools_pool_id_updates_update_id [intent=reverse_etl availability=partial write=create_api_v1_agent_pools_pool_id_updates_update_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates/{update_id} disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates/{updateId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates/{update_id} disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates/{updateId}.; flags: --pool-id (required), --update-id (required)
    create api v1 apps app id connections default apply - Typed action create_api_v1_apps_app_id_connections_default [intent=reverse_etl availability=partial write=create_api_v1_apps_app_id_connections_default]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/connections/default disagrees with covered api_surface path /api/v1/apps/{appId}/connections/default.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/connections/default disagrees with covered api_surface path /api/v1/apps/{appId}/connections/default.; flags: --app-id (required), --profile (required)
    create api v1 apps app id credentials csrs apply - Typed action create_api_v1_apps_app_id_credentials_csrs [intent=reverse_etl availability=partial write=create_api_v1_apps_app_id_credentials_csrs]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/csrs disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/csrs.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/csrs disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/csrs.; flags: --app-id (required)
    create api v1 apps app id credentials jwks apply - Typed action create_api_v1_apps_app_id_credentials_jwks [intent=reverse_etl availability=partial write=create_api_v1_apps_app_id_credentials_jwks]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/jwks disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/jwks.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/jwks disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/jwks.; flags: --app-id (required)
    create api v1 apps app id credentials secrets apply - Typed action create_api_v1_apps_app_id_credentials_secrets [intent=reverse_etl availability=partial write=create_api_v1_apps_app_id_credentials_secrets]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/secrets disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/secrets.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/secrets disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/secrets.; flags: --app-id (required)
    create api v1 apps app id cwo connections apply - Typed action create_api_v1_apps_app_id_cwo_connections [intent=reverse_etl availability=partial write=create_api_v1_apps_app_id_cwo_connections]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/cwo/connections disagrees with covered api_surface path /api/v1/apps/{appId}/cwo/connections.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/cwo/connections disagrees with covered api_surface path /api/v1/apps/{appId}/cwo/connections.; flags: --app-id (required)
    create api v1 apps app id federated claims apply - Typed action create_api_v1_apps_app_id_federated_claims [intent=reverse_etl availability=partial write=create_api_v1_apps_app_id_federated_claims]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/federated-claims disagrees with covered api_surface path /api/v1/apps/{appId}/federated-claims.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/federated-claims disagrees with covered api_surface path /api/v1/apps/{appId}/federated-claims.; flags: --app-id (required)
    create api v1 apps app id grants apply - Typed action create_api_v1_apps_app_id_grants [intent=reverse_etl availability=partial write=create_api_v1_apps_app_id_grants]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/grants disagrees with covered api_surface path /api/v1/apps/{appId}/grants.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/grants disagrees with covered api_surface path /api/v1/apps/{appId}/grants.; flags: --app-id (required), --issuer (required), --scope-id (required)
    create api v1 apps app id group push mappings apply - Typed action create_api_v1_apps_app_id_group_push_mappings [intent=reverse_etl availability=partial write=create_api_v1_apps_app_id_group_push_mappings]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/group-push/mappings disagrees with covered api_surface path /api/v1/apps/{appId}/group-push/mappings.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/group-push/mappings disagrees with covered api_surface path /api/v1/apps/{appId}/group-push/mappings.; flags: --app-id (required), --source-group-id (required)
    create api v1 apps app id interclient allowed apps apply - Typed action create_api_v1_apps_app_id_interclient_allowed_apps [intent=reverse_etl availability=partial write=create_api_v1_apps_app_id_interclient_allowed_apps]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/interclient-allowed-apps disagrees with covered api_surface path /api/v1/apps/{appId}/interclient-allowed-apps.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/interclient-allowed-apps disagrees with covered api_surface path /api/v1/apps/{appId}/interclient-allowed-apps.; flags: --app-id (required)
    create api v1 apps app id users apply - Typed action create_api_v1_apps_app_id_users [intent=reverse_etl availability=partial write=create_api_v1_apps_app_id_users]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/users disagrees with covered api_surface path /api/v1/apps/{appId}/users.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/users disagrees with covered api_surface path /api/v1/apps/{appId}/users.; flags: --app-id (required), --id (required)
    create api v1 apps app id users user id apply - Typed action create_api_v1_apps_app_id_users_user_id [intent=reverse_etl availability=partial write=create_api_v1_apps_app_id_users_user_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/users/{user_id} disagrees with covered api_surface path /api/v1/apps/{appId}/users/{userId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/users/{user_id} disagrees with covered api_surface path /api/v1/apps/{appId}/users/{userId}.; flags: --app-id (required), --user-id (required)
    create api v1 apps app name app id oauth2 callback apply - Typed action create_api_v1_apps_app_name_app_id_oauth2_callback [intent=reverse_etl availability=partial write=create_api_v1_apps_app_name_app_id_oauth2_callback]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_name}/{app_id}/oauth2/callback disagrees with covered api_surface path /api/v1/apps/{appName}/{appId}/oauth2/callback.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_name}/{app_id}/oauth2/callback disagrees with covered api_surface path /api/v1/apps/{appName}/{appId}/oauth2/callback.; flags: --app-id (required), --app-name (required)
    create api v1 apps apply - POST /api/v1/apps (create_api_v1_apps) [intent=reverse_etl availability=implemented write=create_api_v1_apps]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --label (required), --sign-on-mode (required)
    create api v1 authenticators apply - POST /api/v1/authenticators (create_api_v1_authenticators) [intent=reverse_etl availability=implemented write=create_api_v1_authenticators]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 authenticators authenticator id aaguids apply - Typed action create_api_v1_authenticators_authenticator_id_aaguids [intent=reverse_etl availability=partial write=create_api_v1_authenticators_authenticator_id_aaguids]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/aaguids disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/aaguids.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/aaguids disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/aaguids.; flags: --authenticator-id (required)
    create api v1 authenticators authenticator id methods web authn method type verify rp id domain apply - Typed action create_api_v1_authenticators_authenticator_id_methods_web_authn_method_type_verify_rp_id_domain [intent=reverse_etl availability=partial write=create_api_v1_authenticators_authenticator_id_methods_web_authn_method_type_verify_rp_id_domain]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/methods/{web_authn_method_type}/verify-rp-id-domain disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/methods/{webAuthnMethodType}/verify-rp-id-domain.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/methods/{web_authn_method_type}/verify-rp-id-domain disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/methods/{webAuthnMethodType}/verify-rp-id-domain.; flags: --authenticator-id (required), --web-authn-method-type (required)
    create api v1 authorization servers apply - POST /api/v1/authorizationServers (create_api_v1_authorization_servers) [intent=reverse_etl availability=implemented write=create_api_v1_authorization_servers]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 authorization servers auth server id associated servers apply - Typed action create_api_v1_authorization_servers_auth_server_id_associated_servers [intent=reverse_etl availability=partial write=create_api_v1_authorization_servers_auth_server_id_associated_servers]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/associatedServers disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/associatedServers.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/associatedServers disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/associatedServers.; flags: --auth-server-id (required)
    create api v1 authorization servers auth server id claims apply - Typed action create_api_v1_authorization_servers_auth_server_id_claims [intent=reverse_etl availability=partial write=create_api_v1_authorization_servers_auth_server_id_claims]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/claims disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/claims.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/claims disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/claims.; flags: --auth-server-id (required)
    create api v1 authorization servers auth server id policies apply - Typed action create_api_v1_authorization_servers_auth_server_id_policies [intent=reverse_etl availability=partial write=create_api_v1_authorization_servers_auth_server_id_policies]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies.; flags: --auth-server-id (required)
    create api v1 authorization servers auth server id policies policy id rules apply - Typed action create_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules [intent=reverse_etl availability=partial write=create_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies/{policy_id}/rules disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies/{policyId}/rules.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies/{policy_id}/rules disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies/{policyId}/rules.; flags: --auth-server-id (required), --conditions (required), --name (required), --policy-id (required), --type (required)
    create api v1 authorization servers auth server id resourceservercredentials keys apply - Typed action create_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys [intent=reverse_etl availability=partial write=create_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/resourceservercredentials/keys disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/resourceservercredentials/keys.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/resourceservercredentials/keys disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/resourceservercredentials/keys.; flags: --auth-server-id (required)
    create api v1 authorization servers auth server id scopes apply - Typed action create_api_v1_authorization_servers_auth_server_id_scopes [intent=reverse_etl availability=partial write=create_api_v1_authorization_servers_auth_server_id_scopes]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/scopes disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/scopes.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/scopes disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/scopes.; flags: --auth-server-id (required), --name (required)
    create api v1 behaviors apply - POST /api/v1/behaviors (create_api_v1_behaviors) [intent=reverse_etl availability=implemented write=create_api_v1_behaviors]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --name (required), --type (required)
    create api v1 bot protection configuration apply - POST /api/v1/bot-protection/configuration (create_api_v1_bot_protection_configuration) [intent=reverse_etl availability=implemented write=create_api_v1_bot_protection_configuration]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --level (required), --mode (required)
    create api v1 brands apply - POST /api/v1/brands (create_api_v1_brands) [intent=reverse_etl availability=implemented write=create_api_v1_brands]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --name (required)
    create api v1 brands brand id templates email template name customizations apply - Typed action create_api_v1_brands_brand_id_templates_email_template_name_customizations [intent=reverse_etl availability=partial write=create_api_v1_brands_brand_id_templates_email_template_name_customizations]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/templates/email/{template_name}/customizations disagrees with covered api_surface path /api/v1/brands/{brandId}/templates/email/{templateName}/customizations.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/templates/email/{template_name}/customizations disagrees with covered api_surface path /api/v1/brands/{brandId}/templates/email/{templateName}/customizations.; flags: --brand-id (required), --language (required), --template-name (required)
    create api v1 brands brand id templates email template name test apply - Typed action create_api_v1_brands_brand_id_templates_email_template_name_test [intent=reverse_etl availability=partial write=create_api_v1_brands_brand_id_templates_email_template_name_test]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/templates/email/{template_name}/test disagrees with covered api_surface path /api/v1/brands/{brandId}/templates/email/{templateName}/test.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/templates/email/{template_name}/test disagrees with covered api_surface path /api/v1/brands/{brandId}/templates/email/{templateName}/test.; flags: --brand-id (required), --template-name (required)
    create api v1 captchas apply - POST /api/v1/captchas (create_api_v1_captchas) [intent=reverse_etl availability=implemented write=create_api_v1_captchas]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 captchas captcha id apply - Typed action create_api_v1_captchas_captcha_id [intent=reverse_etl availability=partial write=create_api_v1_captchas_captcha_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/captchas/{captcha_id} disagrees with covered api_surface path /api/v1/captchas/{captchaId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/captchas/{captcha_id} disagrees with covered api_surface path /api/v1/captchas/{captchaId}.; flags: --captcha-id (required)
    create api v1 device assurances apply - POST /api/v1/device-assurances (create_api_v1_device_assurances) [intent=reverse_etl availability=implemented write=create_api_v1_device_assurances]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 device posture checks apply - POST /api/v1/device-posture-checks (create_api_v1_device_posture_checks) [intent=reverse_etl availability=implemented write=create_api_v1_device_posture_checks]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 directories app instance id groups group id query apply - Typed action create_api_v1_directories_app_instance_id_groups_group_id_query [intent=reverse_etl availability=partial write=create_api_v1_directories_app_instance_id_groups_group_id_query]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/directories/{app_instance_id}/groups/{group_id}/query disagrees with covered api_surface path /api/v1/directories/{appInstanceId}/groups/{groupId}/query.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/directories/{app_instance_id}/groups/{group_id}/query disagrees with covered api_surface path /api/v1/directories/{appInstanceId}/groups/{groupId}/query.; flags: --app-instance-id (required), --attributes (required), --group-id (required)
    create api v1 directories app instance id groups modify apply - Typed action create_api_v1_directories_app_instance_id_groups_modify [intent=reverse_etl availability=partial write=create_api_v1_directories_app_instance_id_groups_modify]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/directories/{app_instance_id}/groups/modify disagrees with covered api_surface path /api/v1/directories/{appInstanceId}/groups/modify.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/directories/{app_instance_id}/groups/modify disagrees with covered api_surface path /api/v1/directories/{appInstanceId}/groups/modify.; flags: --app-instance-id (required), --id (required), --parameters (required)
    create api v1 domains apply - POST /api/v1/domains (create_api_v1_domains) [intent=reverse_etl availability=implemented write=create_api_v1_domains]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --certificate-source-type (required), --domain (required)
    create api v1 domains domain id verify apply - Typed action create_api_v1_domains_domain_id_verify [intent=reverse_etl availability=partial write=create_api_v1_domains_domain_id_verify]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/domains/{domain_id}/verify disagrees with covered api_surface path /api/v1/domains/{domainId}/verify.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/domains/{domain_id}/verify disagrees with covered api_surface path /api/v1/domains/{domainId}/verify.; flags: --domain-id (required)
    create api v1 dr failback apply - POST /api/v1/dr/failback (create_api_v1_dr_failback) [intent=reverse_etl availability=implemented write=create_api_v1_dr_failback]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 dr failover apply - POST /api/v1/dr/failover (create_api_v1_dr_failover) [intent=reverse_etl availability=implemented write=create_api_v1_dr_failover]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 email domains apply - POST /api/v1/email-domains (create_api_v1_email_domains) [intent=reverse_etl availability=implemented write=create_api_v1_email_domains]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --display-name (required), --user-name (required)
    create api v1 email domains email domain id verify apply - Typed action create_api_v1_email_domains_email_domain_id_verify [intent=reverse_etl availability=partial write=create_api_v1_email_domains_email_domain_id_verify]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/email-domains/{email_domain_id}/verify disagrees with covered api_surface path /api/v1/email-domains/{emailDomainId}/verify.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/email-domains/{email_domain_id}/verify disagrees with covered api_surface path /api/v1/email-domains/{emailDomainId}/verify.; flags: --email-domain-id (required)
    create api v1 email servers apply - POST /api/v1/email-servers (create_api_v1_email_servers) [intent=reverse_etl availability=implemented write=create_api_v1_email_servers]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --alias (required), --auth-type (required), --enabled (required), --host (required), --port (required), --username (required)
    create api v1 email servers email server id test apply - Typed action create_api_v1_email_servers_email_server_id_test [intent=reverse_etl availability=partial write=create_api_v1_email_servers_email_server_id_test]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/email-servers/{email_server_id}/test disagrees with covered api_surface path /api/v1/email-servers/{emailServerId}/test.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/email-servers/{email_server_id}/test disagrees with covered api_surface path /api/v1/email-servers/{emailServerId}/test.; flags: --email-server-id (required), --from-address (required), --to-address (required)
    create api v1 event hooks apply - POST /api/v1/eventHooks (create_api_v1_event_hooks) [intent=reverse_etl availability=implemented write=create_api_v1_event_hooks]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --channel (required), --events (required), --name (required)
    create api v1 features feature id lifecycle apply - Typed action create_api_v1_features_feature_id_lifecycle [intent=reverse_etl availability=partial write=create_api_v1_features_feature_id_lifecycle]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/features/{feature_id}/{lifecycle} disagrees with covered api_surface path /api/v1/features/{featureId}/{lifecycle}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/features/{feature_id}/{lifecycle} disagrees with covered api_surface path /api/v1/features/{featureId}/{lifecycle}.; flags: --feature-id (required), --lifecycle (required)
    create api v1 groups apply - POST /api/v1/groups (create_api_v1_groups) [intent=reverse_etl availability=implemented write=create_api_v1_groups]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 groups group id owners apply - Typed action create_api_v1_groups_group_id_owners [intent=reverse_etl availability=partial write=create_api_v1_groups_group_id_owners]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/groups/{group_id}/owners disagrees with covered api_surface path /api/v1/groups/{groupId}/owners.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/groups/{group_id}/owners disagrees with covered api_surface path /api/v1/groups/{groupId}/owners.; flags: --group-id (required)
    create api v1 groups group id roles apply - Typed action create_api_v1_groups_group_id_roles [intent=reverse_etl availability=partial write=create_api_v1_groups_group_id_roles]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/groups/{group_id}/roles disagrees with covered api_surface path /api/v1/groups/{groupId}/roles.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/groups/{group_id}/roles disagrees with covered api_surface path /api/v1/groups/{groupId}/roles.; flags: --group-id (required), --type (required)
    create api v1 groups rules apply - POST /api/v1/groups/rules (create_api_v1_groups_rules) [intent=reverse_etl availability=implemented write=create_api_v1_groups_rules]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 hook keys apply - POST /api/v1/hook-keys (create_api_v1_hook_keys) [intent=reverse_etl availability=implemented write=create_api_v1_hook_keys]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 iam governance bundles apply - POST /api/v1/iam/governance/bundles (create_api_v1_iam_governance_bundles) [intent=reverse_etl availability=implemented write=create_api_v1_iam_governance_bundles]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 iam governance opt in apply - POST /api/v1/iam/governance/optIn (create_api_v1_iam_governance_opt_in) [intent=reverse_etl availability=implemented write=create_api_v1_iam_governance_opt_in]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 iam governance opt out apply - POST /api/v1/iam/governance/optOut (create_api_v1_iam_governance_opt_out) [intent=reverse_etl availability=implemented write=create_api_v1_iam_governance_opt_out]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 iam resource sets apply - POST /api/v1/iam/resource-sets (create_api_v1_iam_resource_sets) [intent=reverse_etl availability=implemented write=create_api_v1_iam_resource_sets]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --description (required), --label (required), --resources (required)
    create api v1 iam resource sets resource set id or label bindings apply - Typed action create_api_v1_iam_resource_sets_resource_set_id_or_label_bindings [intent=reverse_etl availability=partial write=create_api_v1_iam_resource_sets_resource_set_id_or_label_bindings]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label}/bindings disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}/bindings.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label}/bindings disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}/bindings.; flags: --resource-set-id-or-label (required)
    create api v1 iam resource sets resource set id or label resources apply - Typed action create_api_v1_iam_resource_sets_resource_set_id_or_label_resources [intent=reverse_etl availability=partial write=create_api_v1_iam_resource_sets_resource_set_id_or_label_resources]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label}/resources disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}/resources.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label}/resources disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}/resources.; flags: --conditions (required), --resource-orn-or-url (required), --resource-set-id-or-label (required)
    create api v1 iam roles apply - POST /api/v1/iam/roles (create_api_v1_iam_roles) [intent=reverse_etl availability=implemented write=create_api_v1_iam_roles]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --description (required), --label (required), --permissions (required)
    create api v1 iam roles role id or label permissions permission type apply - Typed action create_api_v1_iam_roles_role_id_or_label_permissions_permission_type [intent=reverse_etl availability=partial write=create_api_v1_iam_roles_role_id_or_label_permissions_permission_type]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/iam/roles/{role_id_or_label}/permissions/{permission_type} disagrees with covered api_surface path /api/v1/iam/roles/{roleIdOrLabel}/permissions/{permissionType}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/iam/roles/{role_id_or_label}/permissions/{permission_type} disagrees with covered api_surface path /api/v1/iam/roles/{roleIdOrLabel}/permissions/{permissionType}.; flags: --permission-type (required), --role-id-or-label (required)
    create api v1 identity sources identity source id groups apply - Typed action create_api_v1_identity_sources_identity_source_id_groups [intent=reverse_etl availability=partial write=create_api_v1_identity_sources_identity_source_id_groups]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/groups disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/groups.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/groups disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/groups.; flags: --identity-source-id (required)
    create api v1 identity sources identity source id groups group or external id apply - Typed action create_api_v1_identity_sources_identity_source_id_groups_group_or_external_id [intent=reverse_etl availability=partial write=create_api_v1_identity_sources_identity_source_id_groups_group_or_external_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/groups/{group_or_external_id} disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/groups/{groupOrExternalId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/groups/{group_or_external_id} disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/groups/{groupOrExternalId}.; flags: --group-or-external-id (required), --identity-source-id (required)
    create api v1 identity sources identity source id groups group or external id membership apply - Typed action create_api_v1_identity_sources_identity_source_id_groups_group_or_external_id_membership [intent=reverse_etl availability=partial write=create_api_v1_identity_sources_identity_source_id_groups_group_or_external_id_membership]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/groups/{group_or_external_id}/membership disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/groups/{groupOrExternalId}/membership.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/groups/{group_or_external_id}/membership disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/groups/{groupOrExternalId}/membership.; flags: --group-or-external-id (required), --identity-source-id (required)
    create api v1 identity sources identity source id sessions apply - Typed action create_api_v1_identity_sources_identity_source_id_sessions [intent=reverse_etl availability=partial write=create_api_v1_identity_sources_identity_source_id_sessions]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/sessions disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/sessions.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/sessions disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/sessions.; flags: --identity-source-id (required)
    create api v1 identity sources identity source id sessions session id bulk delete apply - Typed action create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_delete [intent=reverse_etl availability=partial write=create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_delete]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/sessions/{session_id}/bulk-delete disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/sessions/{sessionId}/bulk-delete.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/sessions/{session_id}/bulk-delete disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/sessions/{sessionId}/bulk-delete.; flags: --identity-source-id (required), --session-id (required)
    create api v1 identity sources identity source id sessions session id bulk group memberships delete apply - Typed action create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_group_memberships_delete [intent=reverse_etl availability=partial write=create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_group_memberships_delete]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/sessions/{session_id}/bulk-group-memberships-delete disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/sessions/{sessionId}/bulk-group-memberships-delete.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/sessions/{session_id}/bulk-group-memberships-delete disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/sessions/{sessionId}/bulk-group-memberships-delete.; flags: --identity-source-id (required), --session-id (required)
    create api v1 identity sources identity source id sessions session id bulk group memberships upsert apply - Typed action create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_group_memberships_upsert [intent=reverse_etl availability=partial write=create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_group_memberships_upsert]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/sessions/{session_id}/bulk-group-memberships-upsert disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/sessions/{sessionId}/bulk-group-memberships-upsert.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/sessions/{session_id}/bulk-group-memberships-upsert disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/sessions/{sessionId}/bulk-group-memberships-upsert.; flags: --identity-source-id (required), --session-id (required)
    create api v1 identity sources identity source id sessions session id bulk groups delete apply - Typed action create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_groups_delete [intent=reverse_etl availability=partial write=create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_groups_delete]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/sessions/{session_id}/bulk-groups-delete disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/sessions/{sessionId}/bulk-groups-delete.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/sessions/{session_id}/bulk-groups-delete disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/sessions/{sessionId}/bulk-groups-delete.; flags: --identity-source-id (required), --session-id (required)
    create api v1 identity sources identity source id sessions session id bulk groups upsert apply - Typed action create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_groups_upsert [intent=reverse_etl availability=partial write=create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_groups_upsert]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/sessions/{session_id}/bulk-groups-upsert disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/sessions/{sessionId}/bulk-groups-upsert.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/sessions/{session_id}/bulk-groups-upsert disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/sessions/{sessionId}/bulk-groups-upsert.; flags: --identity-source-id (required), --session-id (required)
    create api v1 identity sources identity source id sessions session id bulk upsert apply - Typed action create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_upsert [intent=reverse_etl availability=partial write=create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_upsert]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/sessions/{session_id}/bulk-upsert disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/sessions/{sessionId}/bulk-upsert.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/sessions/{session_id}/bulk-upsert disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/sessions/{sessionId}/bulk-upsert.; flags: --identity-source-id (required), --session-id (required)
    create api v1 identity sources identity source id sessions session id start import apply - Typed action create_api_v1_identity_sources_identity_source_id_sessions_session_id_start_import [intent=reverse_etl availability=partial write=create_api_v1_identity_sources_identity_source_id_sessions_session_id_start_import]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/sessions/{session_id}/start-import disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/sessions/{sessionId}/start-import.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/sessions/{session_id}/start-import disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/sessions/{sessionId}/start-import.; flags: --identity-source-id (required), --session-id (required)
    create api v1 identity sources identity source id users apply - Typed action create_api_v1_identity_sources_identity_source_id_users [intent=reverse_etl availability=partial write=create_api_v1_identity_sources_identity_source_id_users]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/users disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/users.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/users disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/users.; flags: --identity-source-id (required)
    create api v1 idps apply - POST /api/v1/idps (create_api_v1_idps) [intent=reverse_etl availability=implemented write=create_api_v1_idps]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 idps credentials keys apply - POST /api/v1/idps/credentials/keys (create_api_v1_idps_credentials_keys) [intent=reverse_etl availability=implemented write=create_api_v1_idps_credentials_keys]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --x5c (required)
    create api v1 idps idp id credentials csrs apply - Typed action create_api_v1_idps_idp_id_credentials_csrs [intent=reverse_etl availability=partial write=create_api_v1_idps_idp_id_credentials_csrs]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/idps/{idp_id}/credentials/csrs disagrees with covered api_surface path /api/v1/idps/{idpId}/credentials/csrs.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/idps/{idp_id}/credentials/csrs disagrees with covered api_surface path /api/v1/idps/{idpId}/credentials/csrs.; flags: --idp-id (required)
    create api v1 idps idp id users user id apply - Typed action create_api_v1_idps_idp_id_users_user_id [intent=reverse_etl availability=partial write=create_api_v1_idps_idp_id_users_user_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/idps/{idp_id}/users/{user_id} disagrees with covered api_surface path /api/v1/idps/{idpId}/users/{userId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/idps/{idp_id}/users/{user_id} disagrees with covered api_surface path /api/v1/idps/{idpId}/users/{userId}.; flags: --idp-id (required), --user-id (required)
    create api v1 inline hooks apply - POST /api/v1/inlineHooks (create_api_v1_inline_hooks) [intent=reverse_etl availability=implemented write=create_api_v1_inline_hooks]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 inline hooks inline hook id apply - Typed action create_api_v1_inline_hooks_inline_hook_id [intent=reverse_etl availability=partial write=create_api_v1_inline_hooks_inline_hook_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/inlineHooks/{inline_hook_id} disagrees with covered api_surface path /api/v1/inlineHooks/{inlineHookId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/inlineHooks/{inline_hook_id} disagrees with covered api_surface path /api/v1/inlineHooks/{inlineHookId}.; flags: --inline-hook-id (required)
    create api v1 inline hooks inline hook id execute apply - Typed action create_api_v1_inline_hooks_inline_hook_id_execute [intent=reverse_etl availability=partial write=create_api_v1_inline_hooks_inline_hook_id_execute]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/inlineHooks/{inline_hook_id}/execute disagrees with covered api_surface path /api/v1/inlineHooks/{inlineHookId}/execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/inlineHooks/{inline_hook_id}/execute disagrees with covered api_surface path /api/v1/inlineHooks/{inlineHookId}/execute.; flags: --inline-hook-id (required)
    create api v1 log streams apply - POST /api/v1/logStreams (create_api_v1_log_streams) [intent=reverse_etl availability=implemented write=create_api_v1_log_streams]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --links (required), --created (required), --id (required), --last-updated (required), --name (required), --status (required), --type (required)
    create api v1 mappings mapping id apply - Typed action create_api_v1_mappings_mapping_id [intent=reverse_etl availability=partial write=create_api_v1_mappings_mapping_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/mappings/{mapping_id} disagrees with covered api_surface path /api/v1/mappings/{mappingId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/mappings/{mapping_id} disagrees with covered api_surface path /api/v1/mappings/{mappingId}.; flags: --mapping-id (required), --properties (required)
    create api v1 meta schemas apps app id default apply - Typed action create_api_v1_meta_schemas_apps_app_id_default [intent=reverse_etl availability=partial write=create_api_v1_meta_schemas_apps_app_id_default]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/meta/schemas/apps/{app_id}/default disagrees with covered api_surface path /api/v1/meta/schemas/apps/{appId}/default.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/meta/schemas/apps/{app_id}/default disagrees with covered api_surface path /api/v1/meta/schemas/apps/{appId}/default.; flags: --app-id (required)
    create api v1 meta schemas group default apply - POST /api/v1/meta/schemas/group/default (create_api_v1_meta_schemas_group_default) [intent=reverse_etl availability=implemented write=create_api_v1_meta_schemas_group_default]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 meta schemas user linked objects apply - POST /api/v1/meta/schemas/user/linkedObjects (create_api_v1_meta_schemas_user_linked_objects) [intent=reverse_etl availability=implemented write=create_api_v1_meta_schemas_user_linked_objects]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 meta schemas user schema id apply - Typed action create_api_v1_meta_schemas_user_schema_id [intent=reverse_etl availability=partial write=create_api_v1_meta_schemas_user_schema_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/meta/schemas/user/{schema_id} disagrees with covered api_surface path /api/v1/meta/schemas/user/{schemaId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/meta/schemas/user/{schema_id} disagrees with covered api_surface path /api/v1/meta/schemas/user/{schemaId}.; flags: --schema-id (required)
    create api v1 meta types user apply - POST /api/v1/meta/types/user (create_api_v1_meta_types_user) [intent=reverse_etl availability=implemented write=create_api_v1_meta_types_user]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --display-name (required), --name (required)
    create api v1 meta types user type id apply - Typed action create_api_v1_meta_types_user_type_id [intent=reverse_etl availability=partial write=create_api_v1_meta_types_user_type_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/meta/types/user/{type_id} disagrees with covered api_surface path /api/v1/meta/types/user/{typeId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/meta/types/user/{type_id} disagrees with covered api_surface path /api/v1/meta/types/user/{typeId}.; flags: --type-id (required)
    create api v1 meta uischemas apply - POST /api/v1/meta/uischemas (create_api_v1_meta_uischemas) [intent=reverse_etl availability=implemented write=create_api_v1_meta_uischemas]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 org apply - POST /api/v1/org (create_api_v1_org) [intent=reverse_etl availability=implemented write=create_api_v1_org]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 org email bounces remove list apply - POST /api/v1/org/email/bounces/remove-list (create_api_v1_org_email_bounces_remove_list) [intent=reverse_etl availability=implemented write=create_api_v1_org_email_bounces_remove_list]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 org factors yubikey token tokens apply - POST /api/v1/org/factors/yubikey_token/tokens (create_api_v1_org_factors_yubikey_token_tokens) [intent=reverse_etl availability=implemented write=create_api_v1_org_factors_yubikey_token_tokens]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 org org settings third party admin setting apply - POST /api/v1/org/orgSettings/thirdPartyAdminSetting (create_api_v1_org_org_settings_third_party_admin_setting) [intent=reverse_etl availability=implemented write=create_api_v1_org_org_settings_third_party_admin_setting]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 org preferences hide end user footer apply - POST /api/v1/org/preferences/hideEndUserFooter (create_api_v1_org_preferences_hide_end_user_footer) [intent=reverse_etl availability=implemented write=create_api_v1_org_preferences_hide_end_user_footer]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 org preferences show end user footer apply - POST /api/v1/org/preferences/showEndUserFooter (create_api_v1_org_preferences_show_end_user_footer) [intent=reverse_etl availability=implemented write=create_api_v1_org_preferences_show_end_user_footer]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 org privacy aerial grant apply - POST /api/v1/org/privacy/aerial/grant (create_api_v1_org_privacy_aerial_grant) [intent=reverse_etl availability=implemented write=create_api_v1_org_privacy_aerial_grant]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --account-id (required)
    create api v1 org privacy okta communication opt in apply - POST /api/v1/org/privacy/oktaCommunication/optIn (create_api_v1_org_privacy_okta_communication_opt_in) [intent=reverse_etl availability=implemented write=create_api_v1_org_privacy_okta_communication_opt_in]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 org privacy okta communication opt out apply - POST /api/v1/org/privacy/oktaCommunication/optOut (create_api_v1_org_privacy_okta_communication_opt_out) [intent=reverse_etl availability=implemented write=create_api_v1_org_privacy_okta_communication_opt_out]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 org privacy okta support extend apply - POST /api/v1/org/privacy/oktaSupport/extend (create_api_v1_org_privacy_okta_support_extend) [intent=reverse_etl availability=implemented write=create_api_v1_org_privacy_okta_support_extend]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 org privacy okta support grant apply - POST /api/v1/org/privacy/oktaSupport/grant (create_api_v1_org_privacy_okta_support_grant) [intent=reverse_etl availability=implemented write=create_api_v1_org_privacy_okta_support_grant]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 org settings auto assign admin app setting apply - POST /api/v1/org/settings/autoAssignAdminAppSetting (create_api_v1_org_settings_auto_assign_admin_app_setting) [intent=reverse_etl availability=implemented write=create_api_v1_org_settings_auto_assign_admin_app_setting]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 orgs apply - POST /api/v1/orgs (create_api_v1_orgs) [intent=reverse_etl availability=implemented write=create_api_v1_orgs]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --admin (required), --edition (required), --name (required), --subdomain (required)
    create api v1 policies apply - POST /api/v1/policies (create_api_v1_policies) [intent=reverse_etl availability=implemented write=create_api_v1_policies]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --name (required), --type (required)
    create api v1 policies policy id clone apply - Typed action create_api_v1_policies_policy_id_clone [intent=reverse_etl availability=partial write=create_api_v1_policies_policy_id_clone]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/clone disagrees with covered api_surface path /api/v1/policies/{policyId}/clone.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/clone disagrees with covered api_surface path /api/v1/policies/{policyId}/clone.; flags: --policy-id (required)
    create api v1 policies policy id mappings apply - Typed action create_api_v1_policies_policy_id_mappings [intent=reverse_etl availability=partial write=create_api_v1_policies_policy_id_mappings]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/mappings disagrees with covered api_surface path /api/v1/policies/{policyId}/mappings.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/mappings disagrees with covered api_surface path /api/v1/policies/{policyId}/mappings.; flags: --policy-id (required)
    create api v1 policies policy id rules apply - Typed action create_api_v1_policies_policy_id_rules [intent=reverse_etl availability=partial write=create_api_v1_policies_policy_id_rules]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/rules disagrees with covered api_surface path /api/v1/policies/{policyId}/rules.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/rules disagrees with covered api_surface path /api/v1/policies/{policyId}/rules.; flags: --policy-id (required)
    create api v1 principal rate limits apply - POST /api/v1/principal-rate-limits (create_api_v1_principal_rate_limits) [intent=reverse_etl availability=implemented write=create_api_v1_principal_rate_limits]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --principal-id (required), --principal-type (required)
    create api v1 push providers apply - POST /api/v1/push-providers (create_api_v1_push_providers) [intent=reverse_etl availability=implemented write=create_api_v1_push_providers]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 realm assignments apply - POST /api/v1/realm-assignments (create_api_v1_realm_assignments) [intent=reverse_etl availability=implemented write=create_api_v1_realm_assignments]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 realm assignments operations apply - POST /api/v1/realm-assignments/operations (create_api_v1_realm_assignments_operations) [intent=reverse_etl availability=implemented write=create_api_v1_realm_assignments_operations]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 realms apply - POST /api/v1/realms (create_api_v1_realms) [intent=reverse_etl availability=implemented write=create_api_v1_realms]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 roles role ref subscriptions notification type subscribe apply - Typed action create_api_v1_roles_role_ref_subscriptions_notification_type_subscribe [intent=reverse_etl availability=partial write=create_api_v1_roles_role_ref_subscriptions_notification_type_subscribe]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/roles/{role_ref}/subscriptions/{notification_type}/subscribe disagrees with covered api_surface path /api/v1/roles/{roleRef}/subscriptions/{notificationType}/subscribe.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/roles/{role_ref}/subscriptions/{notification_type}/subscribe disagrees with covered api_surface path /api/v1/roles/{roleRef}/subscriptions/{notificationType}/subscribe.; flags: --notification-type (required), --role-ref (required)
    create api v1 roles role ref subscriptions notification type unsubscribe apply - Typed action create_api_v1_roles_role_ref_subscriptions_notification_type_unsubscribe [intent=reverse_etl availability=partial write=create_api_v1_roles_role_ref_subscriptions_notification_type_unsubscribe]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/roles/{role_ref}/subscriptions/{notification_type}/unsubscribe disagrees with covered api_surface path /api/v1/roles/{roleRef}/subscriptions/{notificationType}/unsubscribe.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/roles/{role_ref}/subscriptions/{notification_type}/unsubscribe disagrees with covered api_surface path /api/v1/roles/{roleRef}/subscriptions/{notificationType}/unsubscribe.; flags: --notification-type (required), --role-ref (required)
    create api v1 security events providers apply - POST /api/v1/security-events-providers (create_api_v1_security_events_providers) [intent=reverse_etl availability=implemented write=create_api_v1_security_events_providers]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --name (required), --settings (required), --type (required)
    create api v1 ssf stream apply - POST /api/v1/ssf/stream (create_api_v1_ssf_stream) [intent=reverse_etl availability=implemented write=create_api_v1_ssf_stream]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --delivery (required), --events-requested (required)
    create api v1 ssf stream verification apply - POST /api/v1/ssf/stream/verification (create_api_v1_ssf_stream_verification) [intent=reverse_etl availability=implemented write=create_api_v1_ssf_stream_verification]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --stream-id (required)
    create api v1 telephony providers apply - POST /api/v1/telephony-providers (create_api_v1_telephony_providers) [intent=reverse_etl availability=implemented write=create_api_v1_telephony_providers]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 telephony providers custom telephony provider id set as primary apply - Typed action create_api_v1_telephony_providers_custom_telephony_provider_id_set_as_primary [intent=reverse_etl availability=partial write=create_api_v1_telephony_providers_custom_telephony_provider_id_set_as_primary]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/telephony-providers/{custom_telephony_provider_id}/setAsPrimary disagrees with covered api_surface path /api/v1/telephony-providers/{customTelephonyProviderId}/setAsPrimary.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/telephony-providers/{custom_telephony_provider_id}/setAsPrimary disagrees with covered api_surface path /api/v1/telephony-providers/{customTelephonyProviderId}/setAsPrimary.; flags: --custom-telephony-provider-id (required)
    create api v1 telephony providers custom telephony provider id test apply - Typed action create_api_v1_telephony_providers_custom_telephony_provider_id_test [intent=reverse_etl availability=partial write=create_api_v1_telephony_providers_custom_telephony_provider_id_test]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/telephony-providers/{custom_telephony_provider_id}/test disagrees with covered api_surface path /api/v1/telephony-providers/{customTelephonyProviderId}/test.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/telephony-providers/{custom_telephony_provider_id}/test disagrees with covered api_surface path /api/v1/telephony-providers/{customTelephonyProviderId}/test.; flags: --custom-telephony-provider-id (required)
    create api v1 templates sms apply - POST /api/v1/templates/sms (create_api_v1_templates_sms) [intent=reverse_etl availability=implemented write=create_api_v1_templates_sms]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 templates sms template id apply - Typed action create_api_v1_templates_sms_template_id [intent=reverse_etl availability=partial write=create_api_v1_templates_sms_template_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/templates/sms/{template_id} disagrees with covered api_surface path /api/v1/templates/sms/{templateId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/templates/sms/{template_id} disagrees with covered api_surface path /api/v1/templates/sms/{templateId}.; flags: --template-id (required)
    create api v1 threats configuration apply - POST /api/v1/threats/configuration (create_api_v1_threats_configuration) [intent=reverse_etl availability=implemented write=create_api_v1_threats_configuration]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --action (required)
    create api v1 trusted origins apply - POST /api/v1/trustedOrigins (create_api_v1_trusted_origins) [intent=reverse_etl availability=implemented write=create_api_v1_trusted_origins]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create api v1 users apply - POST /api/v1/users (create_api_v1_users) [intent=reverse_etl availability=implemented write=create_api_v1_users]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --profile (required)
    create api v1 users id apply - POST /api/v1/users/{id} (create_api_v1_users_id) [intent=reverse_etl availability=implemented write=create_api_v1_users_id]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    create api v1 users user id authenticator enrollments phone apply - Typed action create_api_v1_users_user_id_authenticator_enrollments_phone [intent=reverse_etl availability=partial write=create_api_v1_users_user_id_authenticator_enrollments_phone]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/authenticator-enrollments/phone disagrees with covered api_surface path /api/v1/users/{userId}/authenticator-enrollments/phone.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/authenticator-enrollments/phone disagrees with covered api_surface path /api/v1/users/{userId}/authenticator-enrollments/phone.; flags: --user-id (required)
    create api v1 users user id authenticator enrollments tac apply - Typed action create_api_v1_users_user_id_authenticator_enrollments_tac [intent=reverse_etl availability=partial write=create_api_v1_users_user_id_authenticator_enrollments_tac]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/authenticator-enrollments/tac disagrees with covered api_surface path /api/v1/users/{userId}/authenticator-enrollments/tac.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/authenticator-enrollments/tac disagrees with covered api_surface path /api/v1/users/{userId}/authenticator-enrollments/tac.; flags: --user-id (required)
    create api v1 users user id credentials change password apply - Typed action create_api_v1_users_user_id_credentials_change_password [intent=reverse_etl availability=partial write=create_api_v1_users_user_id_credentials_change_password]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/credentials/change_password disagrees with covered api_surface path /api/v1/users/{userId}/credentials/change_password.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/credentials/change_password disagrees with covered api_surface path /api/v1/users/{userId}/credentials/change_password.; flags: --user-id (required)
    create api v1 users user id credentials change recovery question apply - Typed action create_api_v1_users_user_id_credentials_change_recovery_question [intent=reverse_etl availability=partial write=create_api_v1_users_user_id_credentials_change_recovery_question]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/credentials/change_recovery_question disagrees with covered api_surface path /api/v1/users/{userId}/credentials/change_recovery_question.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/credentials/change_recovery_question disagrees with covered api_surface path /api/v1/users/{userId}/credentials/change_recovery_question.; flags: --user-id (required)
    create api v1 users user id credentials forgot password apply - Typed action create_api_v1_users_user_id_credentials_forgot_password [intent=reverse_etl availability=partial write=create_api_v1_users_user_id_credentials_forgot_password]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/credentials/forgot_password disagrees with covered api_surface path /api/v1/users/{userId}/credentials/forgot_password.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/credentials/forgot_password disagrees with covered api_surface path /api/v1/users/{userId}/credentials/forgot_password.; flags: --user-id (required)
    create api v1 users user id credentials forgot password recovery question apply - Typed action create_api_v1_users_user_id_credentials_forgot_password_recovery_question [intent=reverse_etl availability=partial write=create_api_v1_users_user_id_credentials_forgot_password_recovery_question]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/credentials/forgot_password_recovery_question disagrees with covered api_surface path /api/v1/users/{userId}/credentials/forgot_password_recovery_question.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/credentials/forgot_password_recovery_question disagrees with covered api_surface path /api/v1/users/{userId}/credentials/forgot_password_recovery_question.; flags: --user-id (required)
    create api v1 users user id factors apply - Typed action create_api_v1_users_user_id_factors [intent=reverse_etl availability=partial write=create_api_v1_users_user_id_factors]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/factors disagrees with covered api_surface path /api/v1/users/{userId}/factors.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/factors disagrees with covered api_surface path /api/v1/users/{userId}/factors.; flags: --user-id (required)
    create api v1 users user id factors factor id resend apply - Typed action create_api_v1_users_user_id_factors_factor_id_resend [intent=reverse_etl availability=partial write=create_api_v1_users_user_id_factors_factor_id_resend]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/factors/{factor_id}/resend disagrees with covered api_surface path /api/v1/users/{userId}/factors/{factorId}/resend.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/factors/{factor_id}/resend disagrees with covered api_surface path /api/v1/users/{userId}/factors/{factorId}/resend.; flags: --factor-id (required), --user-id (required)
    create api v1 users user id factors factor id verify apply - Typed action create_api_v1_users_user_id_factors_factor_id_verify [intent=reverse_etl availability=partial write=create_api_v1_users_user_id_factors_factor_id_verify]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/factors/{factor_id}/verify disagrees with covered api_surface path /api/v1/users/{userId}/factors/{factorId}/verify.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/factors/{factor_id}/verify disagrees with covered api_surface path /api/v1/users/{userId}/factors/{factorId}/verify.; flags: --factor-id (required), --user-id (required)
    create api v1 users user id roles apply - Typed action create_api_v1_users_user_id_roles [intent=reverse_etl availability=partial write=create_api_v1_users_user_id_roles]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/roles disagrees with covered api_surface path /api/v1/users/{userId}/roles.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/roles disagrees with covered api_surface path /api/v1/users/{userId}/roles.; flags: --type (required), --user-id (required)
    create api v1 users user id subscriptions notification type subscribe apply - Typed action create_api_v1_users_user_id_subscriptions_notification_type_subscribe [intent=reverse_etl availability=partial write=create_api_v1_users_user_id_subscriptions_notification_type_subscribe]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/subscriptions/{notification_type}/subscribe disagrees with covered api_surface path /api/v1/users/{userId}/subscriptions/{notificationType}/subscribe.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/subscriptions/{notification_type}/subscribe disagrees with covered api_surface path /api/v1/users/{userId}/subscriptions/{notificationType}/subscribe.; flags: --notification-type (required), --user-id (required)
    create api v1 users user id subscriptions notification type unsubscribe apply - Typed action create_api_v1_users_user_id_subscriptions_notification_type_unsubscribe [intent=reverse_etl availability=partial write=create_api_v1_users_user_id_subscriptions_notification_type_unsubscribe]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/subscriptions/{notification_type}/unsubscribe disagrees with covered api_surface path /api/v1/users/{userId}/subscriptions/{notificationType}/unsubscribe.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/subscriptions/{notification_type}/unsubscribe disagrees with covered api_surface path /api/v1/users/{userId}/subscriptions/{notificationType}/unsubscribe.; flags: --notification-type (required), --user-id (required)
    create api v1 zones apply - POST /api/v1/zones (create_api_v1_zones) [intent=reverse_etl availability=implemented write=create_api_v1_zones]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --name (required), --type (required)
    create integrations api v1 api services api service id credentials secrets apply - Typed action create_integrations_api_v1_api_services_api_service_id_credentials_secrets [intent=reverse_etl availability=partial write=create_integrations_api_v1_api_services_api_service_id_credentials_secrets]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /integrations/api/v1/api-services/{api_service_id}/credentials/secrets disagrees with covered api_surface path /integrations/api/v1/api-services/{apiServiceId}/credentials/secrets.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /integrations/api/v1/api-services/{api_service_id}/credentials/secrets disagrees with covered api_surface path /integrations/api/v1/api-services/{apiServiceId}/credentials/secrets.; flags: --api-service-id (required)
    create integrations api v1 api services apply - POST /integrations/api/v1/api-services (create_integrations_api_v1_api_services) [intent=reverse_etl availability=implemented write=create_integrations_api_v1_api_services]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --granted-scopes (required), --type (required)
    create oauth2 v1 clients client id roles apply - Typed action create_oauth2_v1_clients_client_id_roles [intent=reverse_etl availability=partial write=create_oauth2_v1_clients_client_id_roles]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /oauth2/v1/clients/{client_id}/roles disagrees with covered api_surface path /oauth2/v1/clients/{clientId}/roles.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /oauth2/v1/clients/{client_id}/roles disagrees with covered api_surface path /oauth2/v1/clients/{clientId}/roles.; flags: --client-id (required), --type (required)
    create privileged access api v1 okta service accounts apply - POST /privileged-access/api/v1/okta-service-accounts (create_privileged_access_api_v1_okta_service_accounts) [intent=reverse_etl availability=implemented write=create_privileged_access_api_v1_okta_service_accounts]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --name (required), --okta-user-id (required)
    create privileged access api v1 service accounts apply - POST /privileged-access/api/v1/service-accounts (create_privileged_access_api_v1_service_accounts) [intent=reverse_etl availability=implemented write=create_privileged_access_api_v1_service_accounts]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --container-orn (required), --name (required), --password (required), --username (required)
    create webauthn registration api v1 enroll apply - POST /webauthn-registration/api/v1/enroll (create_webauthn_registration_api_v1_enroll) [intent=reverse_etl availability=implemented write=create_webauthn_registration_api_v1_enroll]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create webauthn registration api v1 initiate fulfillment request apply - POST /webauthn-registration/api/v1/initiate-fulfillment-request (create_webauthn_registration_api_v1_initiate_fulfillment_request) [intent=reverse_etl availability=implemented write=create_webauthn_registration_api_v1_initiate_fulfillment_request]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create webauthn registration api v1 send pin apply - POST /webauthn-registration/api/v1/send-pin (create_webauthn_registration_api_v1_send_pin) [intent=reverse_etl availability=implemented write=create_webauthn_registration_api_v1_send_pin]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    create webauthn registration api v1 users user id enrollments authenticator enrollment id mark error apply - Typed action create_webauthn_registration_api_v1_users_user_id_enrollments_authenticator_enrollment_id_mark_error [intent=reverse_etl availability=partial write=create_webauthn_registration_api_v1_users_user_id_enrollments_authenticator_enrollment_id_mark_error]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /webauthn-registration/api/v1/users/{user_id}/enrollments/{authenticator_enrollment_id}/mark-error disagrees with covered api_surface path /webauthn-registration/api/v1/users/{userId}/enrollments/{authenticatorEnrollmentId}/mark-error.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /webauthn-registration/api/v1/users/{user_id}/enrollments/{authenticator_enrollment_id}/mark-error disagrees with covered api_surface path /webauthn-registration/api/v1/users/{userId}/enrollments/{authenticatorEnrollmentId}/mark-error.; flags: --authenticator-enrollment-id (required), --user-id (required)
    delete api v1 agent pools pool id updates update id apply - Typed action delete_api_v1_agent_pools_pool_id_updates_update_id [intent=reverse_etl availability=partial write=delete_api_v1_agent_pools_pool_id_updates_update_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates/{update_id} disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates/{updateId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates/{update_id} disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates/{updateId}.; flags: --pool-id (required), --update-id (required)
    delete api v1 api tokens api token id apply - Typed action delete_api_v1_api_tokens_api_token_id [intent=reverse_etl availability=partial write=delete_api_v1_api_tokens_api_token_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/api-tokens/{api_token_id} disagrees with covered api_surface path /api/v1/api-tokens/{apiTokenId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/api-tokens/{api_token_id} disagrees with covered api_surface path /api/v1/api-tokens/{apiTokenId}.; flags: --api-token-id (required)
    delete api v1 api tokens current apply - DELETE /api/v1/api-tokens/current (delete_api_v1_api_tokens_current) [intent=reverse_etl availability=implemented write=delete_api_v1_api_tokens_current]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    delete api v1 apps app id apply - Typed action delete_api_v1_apps_app_id [intent=reverse_etl availability=partial write=delete_api_v1_apps_app_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id} disagrees with covered api_surface path /api/v1/apps/{appId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id} disagrees with covered api_surface path /api/v1/apps/{appId}.; flags: --app-id (required)
    delete api v1 apps app id credentials csrs csr id apply - Typed action delete_api_v1_apps_app_id_credentials_csrs_csr_id [intent=reverse_etl availability=partial write=delete_api_v1_apps_app_id_credentials_csrs_csr_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/csrs/{csr_id} disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/csrs/{csrId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/csrs/{csr_id} disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/csrs/{csrId}.; flags: --app-id (required), --csr-id (required)
    delete api v1 apps app id credentials jwks key id apply - Typed action delete_api_v1_apps_app_id_credentials_jwks_key_id [intent=reverse_etl availability=partial write=delete_api_v1_apps_app_id_credentials_jwks_key_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/jwks/{key_id} disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/jwks/{keyId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/jwks/{key_id} disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/jwks/{keyId}.; flags: --app-id (required), --key-id (required)
    delete api v1 apps app id credentials secrets secret id apply - Typed action delete_api_v1_apps_app_id_credentials_secrets_secret_id [intent=reverse_etl availability=partial write=delete_api_v1_apps_app_id_credentials_secrets_secret_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/secrets/{secret_id} disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/secrets/{secretId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/secrets/{secret_id} disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/secrets/{secretId}.; flags: --app-id (required), --secret-id (required)
    delete api v1 apps app id cwo connections connection id apply - Typed action delete_api_v1_apps_app_id_cwo_connections_connection_id [intent=reverse_etl availability=partial write=delete_api_v1_apps_app_id_cwo_connections_connection_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/cwo/connections/{connection_id} disagrees with covered api_surface path /api/v1/apps/{appId}/cwo/connections/{connectionId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/cwo/connections/{connection_id} disagrees with covered api_surface path /api/v1/apps/{appId}/cwo/connections/{connectionId}.; flags: --app-id (required), --connection-id (required)
    delete api v1 apps app id federated claims claim id apply - Typed action delete_api_v1_apps_app_id_federated_claims_claim_id [intent=reverse_etl availability=partial write=delete_api_v1_apps_app_id_federated_claims_claim_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/federated-claims/{claim_id} disagrees with covered api_surface path /api/v1/apps/{appId}/federated-claims/{claimId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/federated-claims/{claim_id} disagrees with covered api_surface path /api/v1/apps/{appId}/federated-claims/{claimId}.; flags: --app-id (required), --claim-id (required)
    delete api v1 apps app id grants grant id apply - Typed action delete_api_v1_apps_app_id_grants_grant_id [intent=reverse_etl availability=partial write=delete_api_v1_apps_app_id_grants_grant_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/grants/{grant_id} disagrees with covered api_surface path /api/v1/apps/{appId}/grants/{grantId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/grants/{grant_id} disagrees with covered api_surface path /api/v1/apps/{appId}/grants/{grantId}.; flags: --app-id (required), --grant-id (required)
    delete api v1 apps app id groups group id apply - Typed action delete_api_v1_apps_app_id_groups_group_id [intent=reverse_etl availability=partial write=delete_api_v1_apps_app_id_groups_group_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/groups/{group_id} disagrees with covered api_surface path /api/v1/apps/{appId}/groups/{groupId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/groups/{group_id} disagrees with covered api_surface path /api/v1/apps/{appId}/groups/{groupId}.; flags: --app-id (required), --group-id (required)
    delete api v1 apps app id interclient allowed apps allowed app id apply - Typed action delete_api_v1_apps_app_id_interclient_allowed_apps_allowed_app_id [intent=reverse_etl availability=partial write=delete_api_v1_apps_app_id_interclient_allowed_apps_allowed_app_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/interclient-allowed-apps/{allowed_app_id} disagrees with covered api_surface path /api/v1/apps/{appId}/interclient-allowed-apps/{allowedAppId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/interclient-allowed-apps/{allowed_app_id} disagrees with covered api_surface path /api/v1/apps/{appId}/interclient-allowed-apps/{allowedAppId}.; flags: --allowed-app-id (required), --app-id (required)
    delete api v1 apps app id tokens apply - Typed action delete_api_v1_apps_app_id_tokens [intent=reverse_etl availability=partial write=delete_api_v1_apps_app_id_tokens]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/tokens disagrees with covered api_surface path /api/v1/apps/{appId}/tokens.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/tokens disagrees with covered api_surface path /api/v1/apps/{appId}/tokens.; flags: --app-id (required)
    delete api v1 apps app id tokens token id apply - Typed action delete_api_v1_apps_app_id_tokens_token_id [intent=reverse_etl availability=partial write=delete_api_v1_apps_app_id_tokens_token_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/tokens/{token_id} disagrees with covered api_surface path /api/v1/apps/{appId}/tokens/{tokenId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/tokens/{token_id} disagrees with covered api_surface path /api/v1/apps/{appId}/tokens/{tokenId}.; flags: --app-id (required), --token-id (required)
    delete api v1 apps app id users user id apply - Typed action delete_api_v1_apps_app_id_users_user_id [intent=reverse_etl availability=partial write=delete_api_v1_apps_app_id_users_user_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/users/{user_id} disagrees with covered api_surface path /api/v1/apps/{appId}/users/{userId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/users/{user_id} disagrees with covered api_surface path /api/v1/apps/{appId}/users/{userId}.; flags: --app-id (required), --user-id (required)
    delete api v1 authenticators authenticator id aaguids aaguid apply - Typed action delete_api_v1_authenticators_authenticator_id_aaguids_aaguid [intent=reverse_etl availability=partial write=delete_api_v1_authenticators_authenticator_id_aaguids_aaguid]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/aaguids/{aaguid} disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/aaguids/{aaguid}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/aaguids/{aaguid} disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/aaguids/{aaguid}.; flags: --aaguid (required), --authenticator-id (required)
    delete api v1 authorization servers auth server id apply - Typed action delete_api_v1_authorization_servers_auth_server_id [intent=reverse_etl availability=partial write=delete_api_v1_authorization_servers_auth_server_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}.; flags: --auth-server-id (required)
    delete api v1 authorization servers auth server id associated servers associated server id apply - Typed action delete_api_v1_authorization_servers_auth_server_id_associated_servers_associated_server_id [intent=reverse_etl availability=partial write=delete_api_v1_authorization_servers_auth_server_id_associated_servers_associated_server_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/associatedServers/{associated_server_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/associatedServers/{associatedServerId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/associatedServers/{associated_server_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/associatedServers/{associatedServerId}.; flags: --associated-server-id (required), --auth-server-id (required)
    delete api v1 authorization servers auth server id claims claim id apply - Typed action delete_api_v1_authorization_servers_auth_server_id_claims_claim_id [intent=reverse_etl availability=partial write=delete_api_v1_authorization_servers_auth_server_id_claims_claim_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/claims/{claim_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/claims/{claimId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/claims/{claim_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/claims/{claimId}.; flags: --auth-server-id (required), --claim-id (required)
    delete api v1 authorization servers auth server id clients client id tokens apply - Typed action delete_api_v1_authorization_servers_auth_server_id_clients_client_id_tokens [intent=reverse_etl availability=partial write=delete_api_v1_authorization_servers_auth_server_id_clients_client_id_tokens]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/clients/{client_id}/tokens disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/clients/{clientId}/tokens.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/clients/{client_id}/tokens disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/clients/{clientId}/tokens.; flags: --auth-server-id (required), --client-id (required)
    delete api v1 authorization servers auth server id clients client id tokens token id apply - Typed action delete_api_v1_authorization_servers_auth_server_id_clients_client_id_tokens_token_id [intent=reverse_etl availability=partial write=delete_api_v1_authorization_servers_auth_server_id_clients_client_id_tokens_token_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/clients/{client_id}/tokens/{token_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/clients/{clientId}/tokens/{tokenId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/clients/{client_id}/tokens/{token_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/clients/{clientId}/tokens/{tokenId}.; flags: --auth-server-id (required), --client-id (required), --token-id (required)
    delete api v1 authorization servers auth server id policies policy id apply - Typed action delete_api_v1_authorization_servers_auth_server_id_policies_policy_id [intent=reverse_etl availability=partial write=delete_api_v1_authorization_servers_auth_server_id_policies_policy_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies/{policy_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies/{policyId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies/{policy_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies/{policyId}.; flags: --auth-server-id (required), --policy-id (required)
    delete api v1 authorization servers auth server id policies policy id rules rule id apply - Typed action delete_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id [intent=reverse_etl availability=partial write=delete_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies/{policy_id}/rules/{rule_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies/{policyId}/rules/{ruleId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies/{policy_id}/rules/{rule_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies/{policyId}/rules/{ruleId}.; flags: --auth-server-id (required), --policy-id (required), --rule-id (required)
    delete api v1 authorization servers auth server id resourceservercredentials keys key id apply - Typed action delete_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id [intent=reverse_etl availability=partial write=delete_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/resourceservercredentials/keys/{key_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/resourceservercredentials/keys/{keyId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/resourceservercredentials/keys/{key_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/resourceservercredentials/keys/{keyId}.; flags: --auth-server-id (required), --key-id (required)
    delete api v1 authorization servers auth server id scopes scope id apply - Typed action delete_api_v1_authorization_servers_auth_server_id_scopes_scope_id [intent=reverse_etl availability=partial write=delete_api_v1_authorization_servers_auth_server_id_scopes_scope_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/scopes/{scope_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/scopes/{scopeId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/scopes/{scope_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/scopes/{scopeId}.; flags: --auth-server-id (required), --scope-id (required)
    delete api v1 behaviors behavior id apply - Typed action delete_api_v1_behaviors_behavior_id [intent=reverse_etl availability=partial write=delete_api_v1_behaviors_behavior_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/behaviors/{behavior_id} disagrees with covered api_surface path /api/v1/behaviors/{behaviorId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/behaviors/{behavior_id} disagrees with covered api_surface path /api/v1/behaviors/{behaviorId}.; flags: --behavior-id (required)
    delete api v1 brands brand id apply - Typed action delete_api_v1_brands_brand_id [intent=reverse_etl availability=partial write=delete_api_v1_brands_brand_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id} disagrees with covered api_surface path /api/v1/brands/{brandId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id} disagrees with covered api_surface path /api/v1/brands/{brandId}.; flags: --brand-id (required)
    delete api v1 brands brand id pages error customized apply - Typed action delete_api_v1_brands_brand_id_pages_error_customized [intent=reverse_etl availability=partial write=delete_api_v1_brands_brand_id_pages_error_customized]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/pages/error/customized disagrees with covered api_surface path /api/v1/brands/{brandId}/pages/error/customized.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/pages/error/customized disagrees with covered api_surface path /api/v1/brands/{brandId}/pages/error/customized.; flags: --brand-id (required)
    delete api v1 brands brand id pages error preview apply - Typed action delete_api_v1_brands_brand_id_pages_error_preview [intent=reverse_etl availability=partial write=delete_api_v1_brands_brand_id_pages_error_preview]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/pages/error/preview disagrees with covered api_surface path /api/v1/brands/{brandId}/pages/error/preview.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/pages/error/preview disagrees with covered api_surface path /api/v1/brands/{brandId}/pages/error/preview.; flags: --brand-id (required)
    delete api v1 brands brand id pages sign in customized apply - Typed action delete_api_v1_brands_brand_id_pages_sign_in_customized [intent=reverse_etl availability=partial write=delete_api_v1_brands_brand_id_pages_sign_in_customized]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/pages/sign-in/customized disagrees with covered api_surface path /api/v1/brands/{brandId}/pages/sign-in/customized.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/pages/sign-in/customized disagrees with covered api_surface path /api/v1/brands/{brandId}/pages/sign-in/customized.; flags: --brand-id (required)
    delete api v1 brands brand id pages sign in preview apply - Typed action delete_api_v1_brands_brand_id_pages_sign_in_preview [intent=reverse_etl availability=partial write=delete_api_v1_brands_brand_id_pages_sign_in_preview]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/pages/sign-in/preview disagrees with covered api_surface path /api/v1/brands/{brandId}/pages/sign-in/preview.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/pages/sign-in/preview disagrees with covered api_surface path /api/v1/brands/{brandId}/pages/sign-in/preview.; flags: --brand-id (required)
    delete api v1 brands brand id templates email template name customizations apply - Typed action delete_api_v1_brands_brand_id_templates_email_template_name_customizations [intent=reverse_etl availability=partial write=delete_api_v1_brands_brand_id_templates_email_template_name_customizations]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/templates/email/{template_name}/customizations disagrees with covered api_surface path /api/v1/brands/{brandId}/templates/email/{templateName}/customizations.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/templates/email/{template_name}/customizations disagrees with covered api_surface path /api/v1/brands/{brandId}/templates/email/{templateName}/customizations.; flags: --brand-id (required), --template-name (required)
    delete api v1 brands brand id templates email template name customizations customization id apply - Typed action delete_api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id [intent=reverse_etl availability=partial write=delete_api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/templates/email/{template_name}/customizations/{customization_id} disagrees with covered api_surface path /api/v1/brands/{brandId}/templates/email/{templateName}/customizations/{customizationId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/templates/email/{template_name}/customizations/{customization_id} disagrees with covered api_surface path /api/v1/brands/{brandId}/templates/email/{templateName}/customizations/{customizationId}.; flags: --brand-id (required), --customization-id (required), --template-name (required)
    delete api v1 brands brand id themes theme id background image apply - Typed action delete_api_v1_brands_brand_id_themes_theme_id_background_image [intent=reverse_etl availability=partial write=delete_api_v1_brands_brand_id_themes_theme_id_background_image]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/themes/{theme_id}/background-image disagrees with covered api_surface path /api/v1/brands/{brandId}/themes/{themeId}/background-image.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/themes/{theme_id}/background-image disagrees with covered api_surface path /api/v1/brands/{brandId}/themes/{themeId}/background-image.; flags: --brand-id (required), --theme-id (required)
    delete api v1 brands brand id themes theme id favicon apply - Typed action delete_api_v1_brands_brand_id_themes_theme_id_favicon [intent=reverse_etl availability=partial write=delete_api_v1_brands_brand_id_themes_theme_id_favicon]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/themes/{theme_id}/favicon disagrees with covered api_surface path /api/v1/brands/{brandId}/themes/{themeId}/favicon.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/themes/{theme_id}/favicon disagrees with covered api_surface path /api/v1/brands/{brandId}/themes/{themeId}/favicon.; flags: --brand-id (required), --theme-id (required)
    delete api v1 brands brand id themes theme id logo apply - Typed action delete_api_v1_brands_brand_id_themes_theme_id_logo [intent=reverse_etl availability=partial write=delete_api_v1_brands_brand_id_themes_theme_id_logo]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/themes/{theme_id}/logo disagrees with covered api_surface path /api/v1/brands/{brandId}/themes/{themeId}/logo.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/themes/{theme_id}/logo disagrees with covered api_surface path /api/v1/brands/{brandId}/themes/{themeId}/logo.; flags: --brand-id (required), --theme-id (required)
    delete api v1 captchas captcha id apply - Typed action delete_api_v1_captchas_captcha_id [intent=reverse_etl availability=partial write=delete_api_v1_captchas_captcha_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/captchas/{captcha_id} disagrees with covered api_surface path /api/v1/captchas/{captchaId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/captchas/{captcha_id} disagrees with covered api_surface path /api/v1/captchas/{captchaId}.; flags: --captcha-id (required)
    delete api v1 device assurances device assurance id apply - Typed action delete_api_v1_device_assurances_device_assurance_id [intent=reverse_etl availability=partial write=delete_api_v1_device_assurances_device_assurance_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/device-assurances/{device_assurance_id} disagrees with covered api_surface path /api/v1/device-assurances/{deviceAssuranceId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/device-assurances/{device_assurance_id} disagrees with covered api_surface path /api/v1/device-assurances/{deviceAssuranceId}.; flags: --device-assurance-id (required)
    delete api v1 device posture checks posture check id apply - Typed action delete_api_v1_device_posture_checks_posture_check_id [intent=reverse_etl availability=partial write=delete_api_v1_device_posture_checks_posture_check_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/device-posture-checks/{posture_check_id} disagrees with covered api_surface path /api/v1/device-posture-checks/{postureCheckId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/device-posture-checks/{posture_check_id} disagrees with covered api_surface path /api/v1/device-posture-checks/{postureCheckId}.; flags: --posture-check-id (required)
    delete api v1 devices device id apply - Typed action delete_api_v1_devices_device_id [intent=reverse_etl availability=partial write=delete_api_v1_devices_device_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/devices/{device_id} disagrees with covered api_surface path /api/v1/devices/{deviceId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/devices/{device_id} disagrees with covered api_surface path /api/v1/devices/{deviceId}.; flags: --device-id (required)
    delete api v1 domains domain id apply - Typed action delete_api_v1_domains_domain_id [intent=reverse_etl availability=partial write=delete_api_v1_domains_domain_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/domains/{domain_id} disagrees with covered api_surface path /api/v1/domains/{domainId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/domains/{domain_id} disagrees with covered api_surface path /api/v1/domains/{domainId}.; flags: --domain-id (required)
    delete api v1 email domains email domain id apply - Typed action delete_api_v1_email_domains_email_domain_id [intent=reverse_etl availability=partial write=delete_api_v1_email_domains_email_domain_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/email-domains/{email_domain_id} disagrees with covered api_surface path /api/v1/email-domains/{emailDomainId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/email-domains/{email_domain_id} disagrees with covered api_surface path /api/v1/email-domains/{emailDomainId}.; flags: --email-domain-id (required)
    delete api v1 email servers email server id apply - Typed action delete_api_v1_email_servers_email_server_id [intent=reverse_etl availability=partial write=delete_api_v1_email_servers_email_server_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/email-servers/{email_server_id} disagrees with covered api_surface path /api/v1/email-servers/{emailServerId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/email-servers/{email_server_id} disagrees with covered api_surface path /api/v1/email-servers/{emailServerId}.; flags: --email-server-id (required)
    delete api v1 event hooks event hook id apply - Typed action delete_api_v1_event_hooks_event_hook_id [intent=reverse_etl availability=partial write=delete_api_v1_event_hooks_event_hook_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/eventHooks/{event_hook_id} disagrees with covered api_surface path /api/v1/eventHooks/{eventHookId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/eventHooks/{event_hook_id} disagrees with covered api_surface path /api/v1/eventHooks/{eventHookId}.; flags: --event-hook-id (required)
    delete api v1 groups group id apply - Typed action delete_api_v1_groups_group_id [intent=reverse_etl availability=partial write=delete_api_v1_groups_group_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/groups/{group_id} disagrees with covered api_surface path /api/v1/groups/{groupId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/groups/{group_id} disagrees with covered api_surface path /api/v1/groups/{groupId}.; flags: --group-id (required)
    delete api v1 groups group id owners owner id apply - Typed action delete_api_v1_groups_group_id_owners_owner_id [intent=reverse_etl availability=partial write=delete_api_v1_groups_group_id_owners_owner_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/groups/{group_id}/owners/{owner_id} disagrees with covered api_surface path /api/v1/groups/{groupId}/owners/{ownerId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/groups/{group_id}/owners/{owner_id} disagrees with covered api_surface path /api/v1/groups/{groupId}/owners/{ownerId}.; flags: --group-id (required), --owner-id (required)
    delete api v1 groups group id roles role assignment id apply - Typed action delete_api_v1_groups_group_id_roles_role_assignment_id [intent=reverse_etl availability=partial write=delete_api_v1_groups_group_id_roles_role_assignment_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/groups/{group_id}/roles/{role_assignment_id} disagrees with covered api_surface path /api/v1/groups/{groupId}/roles/{roleAssignmentId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/groups/{group_id}/roles/{role_assignment_id} disagrees with covered api_surface path /api/v1/groups/{groupId}/roles/{roleAssignmentId}.; flags: --group-id (required), --role-assignment-id (required)
    delete api v1 groups group id roles role assignment id targets catalog apps app name app id apply - Typed action delete_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id [intent=reverse_etl availability=partial write=delete_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/groups/{group_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name}/{app_id} disagrees with covered api_surface path /api/v1/groups/{groupId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}/{appId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/groups/{group_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name}/{app_id} disagrees with covered api_surface path /api/v1/groups/{groupId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}/{appId}.; flags: --app-id (required), --app-name (required), --group-id (required), --role-assignment-id (required)
    delete api v1 groups group id roles role assignment id targets catalog apps app name apply - Typed action delete_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name [intent=reverse_etl availability=partial write=delete_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/groups/{group_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name} disagrees with covered api_surface path /api/v1/groups/{groupId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/groups/{group_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name} disagrees with covered api_surface path /api/v1/groups/{groupId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}.; flags: --app-name (required), --group-id (required), --role-assignment-id (required)
    delete api v1 groups group id roles role assignment id targets groups target group id apply - Typed action delete_api_v1_groups_group_id_roles_role_assignment_id_targets_groups_target_group_id [intent=reverse_etl availability=partial write=delete_api_v1_groups_group_id_roles_role_assignment_id_targets_groups_target_group_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/groups/{group_id}/roles/{role_assignment_id}/targets/groups/{target_group_id} disagrees with covered api_surface path /api/v1/groups/{groupId}/roles/{roleAssignmentId}/targets/groups/{targetGroupId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/groups/{group_id}/roles/{role_assignment_id}/targets/groups/{target_group_id} disagrees with covered api_surface path /api/v1/groups/{groupId}/roles/{roleAssignmentId}/targets/groups/{targetGroupId}.; flags: --group-id (required), --role-assignment-id (required), --target-group-id (required)
    delete api v1 groups group id users user id apply - Typed action delete_api_v1_groups_group_id_users_user_id [intent=reverse_etl availability=partial write=delete_api_v1_groups_group_id_users_user_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/groups/{group_id}/users/{user_id} disagrees with covered api_surface path /api/v1/groups/{groupId}/users/{userId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/groups/{group_id}/users/{user_id} disagrees with covered api_surface path /api/v1/groups/{groupId}/users/{userId}.; flags: --group-id (required), --user-id (required)
    delete api v1 groups rules group rule id apply - Typed action delete_api_v1_groups_rules_group_rule_id [intent=reverse_etl availability=partial write=delete_api_v1_groups_rules_group_rule_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/groups/rules/{group_rule_id} disagrees with covered api_surface path /api/v1/groups/rules/{groupRuleId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/groups/rules/{group_rule_id} disagrees with covered api_surface path /api/v1/groups/rules/{groupRuleId}.; flags: --group-rule-id (required)
    delete api v1 hook keys id apply - DELETE /api/v1/hook-keys/{id} (delete_api_v1_hook_keys_id) [intent=reverse_etl availability=implemented write=delete_api_v1_hook_keys_id]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    delete api v1 iam governance bundles bundle id apply - Typed action delete_api_v1_iam_governance_bundles_bundle_id [intent=reverse_etl availability=partial write=delete_api_v1_iam_governance_bundles_bundle_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/iam/governance/bundles/{bundle_id} disagrees with covered api_surface path /api/v1/iam/governance/bundles/{bundleId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/iam/governance/bundles/{bundle_id} disagrees with covered api_surface path /api/v1/iam/governance/bundles/{bundleId}.; flags: --bundle-id (required)
    delete api v1 iam resource sets resource set id or label apply - Typed action delete_api_v1_iam_resource_sets_resource_set_id_or_label [intent=reverse_etl availability=partial write=delete_api_v1_iam_resource_sets_resource_set_id_or_label]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label} disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label} disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}.; flags: --resource-set-id-or-label (required)
    delete api v1 iam resource sets resource set id or label bindings role id or label apply - Typed action delete_api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label [intent=reverse_etl availability=partial write=delete_api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label}/bindings/{role_id_or_label} disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}/bindings/{roleIdOrLabel}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label}/bindings/{role_id_or_label} disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}/bindings/{roleIdOrLabel}.; flags: --resource-set-id-or-label (required), --role-id-or-label (required)
    delete api v1 iam resource sets resource set id or label bindings role id or label members member id apply - Typed action delete_api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members_member_id [intent=reverse_etl availability=partial write=delete_api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members_member_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label}/bindings/{role_id_or_label}/members/{member_id} disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}/bindings/{roleIdOrLabel}/members/{memberId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label}/bindings/{role_id_or_label}/members/{member_id} disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}/bindings/{roleIdOrLabel}/members/{memberId}.; flags: --member-id (required), --resource-set-id-or-label (required), --role-id-or-label (required)
    delete api v1 iam resource sets resource set id or label resources resource id apply - Typed action delete_api_v1_iam_resource_sets_resource_set_id_or_label_resources_resource_id [intent=reverse_etl availability=partial write=delete_api_v1_iam_resource_sets_resource_set_id_or_label_resources_resource_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label}/resources/{resource_id} disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}/resources/{resourceId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label}/resources/{resource_id} disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}/resources/{resourceId}.; flags: --resource-id (required), --resource-set-id-or-label (required)
    delete api v1 iam roles role id or label apply - Typed action delete_api_v1_iam_roles_role_id_or_label [intent=reverse_etl availability=partial write=delete_api_v1_iam_roles_role_id_or_label]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/iam/roles/{role_id_or_label} disagrees with covered api_surface path /api/v1/iam/roles/{roleIdOrLabel}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/iam/roles/{role_id_or_label} disagrees with covered api_surface path /api/v1/iam/roles/{roleIdOrLabel}.; flags: --role-id-or-label (required)
    delete api v1 iam roles role id or label permissions permission type apply - Typed action delete_api_v1_iam_roles_role_id_or_label_permissions_permission_type [intent=reverse_etl availability=partial write=delete_api_v1_iam_roles_role_id_or_label_permissions_permission_type]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/iam/roles/{role_id_or_label}/permissions/{permission_type} disagrees with covered api_surface path /api/v1/iam/roles/{roleIdOrLabel}/permissions/{permissionType}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/iam/roles/{role_id_or_label}/permissions/{permission_type} disagrees with covered api_surface path /api/v1/iam/roles/{roleIdOrLabel}/permissions/{permissionType}.; flags: --permission-type (required), --role-id-or-label (required)
    delete api v1 identity sources identity source id groups group or external id apply - Typed action delete_api_v1_identity_sources_identity_source_id_groups_group_or_external_id [intent=reverse_etl availability=partial write=delete_api_v1_identity_sources_identity_source_id_groups_group_or_external_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/groups/{group_or_external_id} disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/groups/{groupOrExternalId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/groups/{group_or_external_id} disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/groups/{groupOrExternalId}.; flags: --group-or-external-id (required), --identity-source-id (required)
    delete api v1 identity sources identity source id groups group or external id membership member external id apply - Typed action delete_api_v1_identity_sources_identity_source_id_groups_group_or_external_id_membership_member_external_id [intent=reverse_etl availability=partial write=delete_api_v1_identity_sources_identity_source_id_groups_group_or_external_id_membership_member_external_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/groups/{group_or_external_id}/membership/{member_external_id} disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/groups/{groupOrExternalId}/membership/{memberExternalId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/groups/{group_or_external_id}/membership/{member_external_id} disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/groups/{groupOrExternalId}/membership/{memberExternalId}.; flags: --group-or-external-id (required), --identity-source-id (required), --member-external-id (required)
    delete api v1 identity sources identity source id sessions session id apply - Typed action delete_api_v1_identity_sources_identity_source_id_sessions_session_id [intent=reverse_etl availability=partial write=delete_api_v1_identity_sources_identity_source_id_sessions_session_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/sessions/{session_id} disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/sessions/{sessionId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/sessions/{session_id} disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/sessions/{sessionId}.; flags: --identity-source-id (required), --session-id (required)
    delete api v1 identity sources identity source id users external id apply - Typed action delete_api_v1_identity_sources_identity_source_id_users_external_id [intent=reverse_etl availability=partial write=delete_api_v1_identity_sources_identity_source_id_users_external_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/users/{external_id} disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/users/{externalId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/users/{external_id} disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/users/{externalId}.; flags: --external-id (required), --identity-source-id (required)
    delete api v1 idps credentials keys kid apply - DELETE /api/v1/idps/credentials/keys/{kid} (delete_api_v1_idps_credentials_keys_kid) [intent=reverse_etl availability=implemented write=delete_api_v1_idps_credentials_keys_kid]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --kid (required)
    delete api v1 idps idp id apply - Typed action delete_api_v1_idps_idp_id [intent=reverse_etl availability=partial write=delete_api_v1_idps_idp_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/idps/{idp_id} disagrees with covered api_surface path /api/v1/idps/{idpId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/idps/{idp_id} disagrees with covered api_surface path /api/v1/idps/{idpId}.; flags: --idp-id (required)
    delete api v1 idps idp id credentials csrs idp csr id apply - Typed action delete_api_v1_idps_idp_id_credentials_csrs_idp_csr_id [intent=reverse_etl availability=partial write=delete_api_v1_idps_idp_id_credentials_csrs_idp_csr_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/idps/{idp_id}/credentials/csrs/{idp_csr_id} disagrees with covered api_surface path /api/v1/idps/{idpId}/credentials/csrs/{idpCsrId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/idps/{idp_id}/credentials/csrs/{idp_csr_id} disagrees with covered api_surface path /api/v1/idps/{idpId}/credentials/csrs/{idpCsrId}.; flags: --idp-csr-id (required), --idp-id (required)
    delete api v1 idps idp id users user id apply - Typed action delete_api_v1_idps_idp_id_users_user_id [intent=reverse_etl availability=partial write=delete_api_v1_idps_idp_id_users_user_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/idps/{idp_id}/users/{user_id} disagrees with covered api_surface path /api/v1/idps/{idpId}/users/{userId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/idps/{idp_id}/users/{user_id} disagrees with covered api_surface path /api/v1/idps/{idpId}/users/{userId}.; flags: --idp-id (required), --user-id (required)
    delete api v1 inline hooks inline hook id apply - Typed action delete_api_v1_inline_hooks_inline_hook_id [intent=reverse_etl availability=partial write=delete_api_v1_inline_hooks_inline_hook_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/inlineHooks/{inline_hook_id} disagrees with covered api_surface path /api/v1/inlineHooks/{inlineHookId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/inlineHooks/{inline_hook_id} disagrees with covered api_surface path /api/v1/inlineHooks/{inlineHookId}.; flags: --inline-hook-id (required)
    delete api v1 log streams log stream id apply - Typed action delete_api_v1_log_streams_log_stream_id [intent=reverse_etl availability=partial write=delete_api_v1_log_streams_log_stream_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/logStreams/{log_stream_id} disagrees with covered api_surface path /api/v1/logStreams/{logStreamId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/logStreams/{log_stream_id} disagrees with covered api_surface path /api/v1/logStreams/{logStreamId}.; flags: --log-stream-id (required)
    delete api v1 meta schemas user linked objects linked object name apply - Typed action delete_api_v1_meta_schemas_user_linked_objects_linked_object_name [intent=reverse_etl availability=partial write=delete_api_v1_meta_schemas_user_linked_objects_linked_object_name]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/meta/schemas/user/linkedObjects/{linked_object_name} disagrees with covered api_surface path /api/v1/meta/schemas/user/linkedObjects/{linkedObjectName}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/meta/schemas/user/linkedObjects/{linked_object_name} disagrees with covered api_surface path /api/v1/meta/schemas/user/linkedObjects/{linkedObjectName}.; flags: --linked-object-name (required)
    delete api v1 meta types user type id apply - Typed action delete_api_v1_meta_types_user_type_id [intent=reverse_etl availability=partial write=delete_api_v1_meta_types_user_type_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/meta/types/user/{type_id} disagrees with covered api_surface path /api/v1/meta/types/user/{typeId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/meta/types/user/{type_id} disagrees with covered api_surface path /api/v1/meta/types/user/{typeId}.; flags: --type-id (required)
    delete api v1 meta uischemas id apply - DELETE /api/v1/meta/uischemas/{id} (delete_api_v1_meta_uischemas_id) [intent=reverse_etl availability=implemented write=delete_api_v1_meta_uischemas_id]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    delete api v1 org captcha apply - DELETE /api/v1/org/captcha (delete_api_v1_org_captcha) [intent=reverse_etl availability=implemented write=delete_api_v1_org_captcha]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    delete api v1 policies policy id apply - Typed action delete_api_v1_policies_policy_id [intent=reverse_etl availability=partial write=delete_api_v1_policies_policy_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/policies/{policy_id} disagrees with covered api_surface path /api/v1/policies/{policyId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/policies/{policy_id} disagrees with covered api_surface path /api/v1/policies/{policyId}.; flags: --policy-id (required)
    delete api v1 policies policy id mappings mapping id apply - Typed action delete_api_v1_policies_policy_id_mappings_mapping_id [intent=reverse_etl availability=partial write=delete_api_v1_policies_policy_id_mappings_mapping_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/mappings/{mapping_id} disagrees with covered api_surface path /api/v1/policies/{policyId}/mappings/{mappingId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/mappings/{mapping_id} disagrees with covered api_surface path /api/v1/policies/{policyId}/mappings/{mappingId}.; flags: --mapping-id (required), --policy-id (required)
    delete api v1 policies policy id rules rule id apply - Typed action delete_api_v1_policies_policy_id_rules_rule_id [intent=reverse_etl availability=partial write=delete_api_v1_policies_policy_id_rules_rule_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/rules/{rule_id} disagrees with covered api_surface path /api/v1/policies/{policyId}/rules/{ruleId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/rules/{rule_id} disagrees with covered api_surface path /api/v1/policies/{policyId}/rules/{ruleId}.; flags: --policy-id (required), --rule-id (required)
    delete api v1 push providers push provider id apply - Typed action delete_api_v1_push_providers_push_provider_id [intent=reverse_etl availability=partial write=delete_api_v1_push_providers_push_provider_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/push-providers/{push_provider_id} disagrees with covered api_surface path /api/v1/push-providers/{pushProviderId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/push-providers/{push_provider_id} disagrees with covered api_surface path /api/v1/push-providers/{pushProviderId}.; flags: --push-provider-id (required)
    delete api v1 realm assignments assignment id apply - Typed action delete_api_v1_realm_assignments_assignment_id [intent=reverse_etl availability=partial write=delete_api_v1_realm_assignments_assignment_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/realm-assignments/{assignment_id} disagrees with covered api_surface path /api/v1/realm-assignments/{assignmentId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/realm-assignments/{assignment_id} disagrees with covered api_surface path /api/v1/realm-assignments/{assignmentId}.; flags: --assignment-id (required)
    delete api v1 realms realm id apply - Typed action delete_api_v1_realms_realm_id [intent=reverse_etl availability=partial write=delete_api_v1_realms_realm_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/realms/{realm_id} disagrees with covered api_surface path /api/v1/realms/{realmId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/realms/{realm_id} disagrees with covered api_surface path /api/v1/realms/{realmId}.; flags: --realm-id (required)
    delete api v1 security events providers security event provider id apply - Typed action delete_api_v1_security_events_providers_security_event_provider_id [intent=reverse_etl availability=partial write=delete_api_v1_security_events_providers_security_event_provider_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/security-events-providers/{security_event_provider_id} disagrees with covered api_surface path /api/v1/security-events-providers/{securityEventProviderId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/security-events-providers/{security_event_provider_id} disagrees with covered api_surface path /api/v1/security-events-providers/{securityEventProviderId}.; flags: --security-event-provider-id (required)
    delete api v1 sessions session id apply - Typed action delete_api_v1_sessions_session_id [intent=reverse_etl availability=partial write=delete_api_v1_sessions_session_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/sessions/{session_id} disagrees with covered api_surface path /api/v1/sessions/{sessionId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/sessions/{session_id} disagrees with covered api_surface path /api/v1/sessions/{sessionId}.; flags: --session-id (required)
    delete api v1 ssf stream apply - DELETE /api/v1/ssf/stream (delete_api_v1_ssf_stream) [intent=reverse_etl availability=implemented write=delete_api_v1_ssf_stream]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    delete api v1 telephony providers custom telephony provider id apply - Typed action delete_api_v1_telephony_providers_custom_telephony_provider_id [intent=reverse_etl availability=partial write=delete_api_v1_telephony_providers_custom_telephony_provider_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/telephony-providers/{custom_telephony_provider_id} disagrees with covered api_surface path /api/v1/telephony-providers/{customTelephonyProviderId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/telephony-providers/{custom_telephony_provider_id} disagrees with covered api_surface path /api/v1/telephony-providers/{customTelephonyProviderId}.; flags: --custom-telephony-provider-id (required)
    delete api v1 templates sms template id apply - Typed action delete_api_v1_templates_sms_template_id [intent=reverse_etl availability=partial write=delete_api_v1_templates_sms_template_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/templates/sms/{template_id} disagrees with covered api_surface path /api/v1/templates/sms/{templateId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/templates/sms/{template_id} disagrees with covered api_surface path /api/v1/templates/sms/{templateId}.; flags: --template-id (required)
    delete api v1 trusted origins trusted origin id apply - Typed action delete_api_v1_trusted_origins_trusted_origin_id [intent=reverse_etl availability=partial write=delete_api_v1_trusted_origins_trusted_origin_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/trustedOrigins/{trusted_origin_id} disagrees with covered api_surface path /api/v1/trustedOrigins/{trustedOriginId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/trustedOrigins/{trusted_origin_id} disagrees with covered api_surface path /api/v1/trustedOrigins/{trustedOriginId}.; flags: --trusted-origin-id (required)
    delete api v1 users id apply - DELETE /api/v1/users/{id} (delete_api_v1_users_id) [intent=reverse_etl availability=implemented write=delete_api_v1_users_id]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    delete api v1 users user id authenticator enrollments enrollment id apply - Typed action delete_api_v1_users_user_id_authenticator_enrollments_enrollment_id [intent=reverse_etl availability=partial write=delete_api_v1_users_user_id_authenticator_enrollments_enrollment_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/authenticator-enrollments/{enrollment_id} disagrees with covered api_surface path /api/v1/users/{userId}/authenticator-enrollments/{enrollmentId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/authenticator-enrollments/{enrollment_id} disagrees with covered api_surface path /api/v1/users/{userId}/authenticator-enrollments/{enrollmentId}.; flags: --enrollment-id (required), --user-id (required)
    delete api v1 users user id clients client id grants apply - Typed action delete_api_v1_users_user_id_clients_client_id_grants [intent=reverse_etl availability=partial write=delete_api_v1_users_user_id_clients_client_id_grants]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/clients/{client_id}/grants disagrees with covered api_surface path /api/v1/users/{userId}/clients/{clientId}/grants.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/clients/{client_id}/grants disagrees with covered api_surface path /api/v1/users/{userId}/clients/{clientId}/grants.; flags: --client-id (required), --user-id (required)
    delete api v1 users user id clients client id tokens apply - Typed action delete_api_v1_users_user_id_clients_client_id_tokens [intent=reverse_etl availability=partial write=delete_api_v1_users_user_id_clients_client_id_tokens]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/clients/{client_id}/tokens disagrees with covered api_surface path /api/v1/users/{userId}/clients/{clientId}/tokens.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/clients/{client_id}/tokens disagrees with covered api_surface path /api/v1/users/{userId}/clients/{clientId}/tokens.; flags: --client-id (required), --user-id (required)
    delete api v1 users user id clients client id tokens token id apply - Typed action delete_api_v1_users_user_id_clients_client_id_tokens_token_id [intent=reverse_etl availability=partial write=delete_api_v1_users_user_id_clients_client_id_tokens_token_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/clients/{client_id}/tokens/{token_id} disagrees with covered api_surface path /api/v1/users/{userId}/clients/{clientId}/tokens/{tokenId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/clients/{client_id}/tokens/{token_id} disagrees with covered api_surface path /api/v1/users/{userId}/clients/{clientId}/tokens/{tokenId}.; flags: --client-id (required), --token-id (required), --user-id (required)
    delete api v1 users user id factors factor id apply - Typed action delete_api_v1_users_user_id_factors_factor_id [intent=reverse_etl availability=partial write=delete_api_v1_users_user_id_factors_factor_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/factors/{factor_id} disagrees with covered api_surface path /api/v1/users/{userId}/factors/{factorId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/factors/{factor_id} disagrees with covered api_surface path /api/v1/users/{userId}/factors/{factorId}.; flags: --factor-id (required), --user-id (required)
    delete api v1 users user id grants apply - Typed action delete_api_v1_users_user_id_grants [intent=reverse_etl availability=partial write=delete_api_v1_users_user_id_grants]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/grants disagrees with covered api_surface path /api/v1/users/{userId}/grants.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/grants disagrees with covered api_surface path /api/v1/users/{userId}/grants.; flags: --user-id (required)
    delete api v1 users user id grants grant id apply - Typed action delete_api_v1_users_user_id_grants_grant_id [intent=reverse_etl availability=partial write=delete_api_v1_users_user_id_grants_grant_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/grants/{grant_id} disagrees with covered api_surface path /api/v1/users/{userId}/grants/{grantId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/grants/{grant_id} disagrees with covered api_surface path /api/v1/users/{userId}/grants/{grantId}.; flags: --grant-id (required), --user-id (required)
    delete api v1 users user id or login linked objects relationship name apply - Typed action delete_api_v1_users_user_id_or_login_linked_objects_relationship_name [intent=reverse_etl availability=partial write=delete_api_v1_users_user_id_or_login_linked_objects_relationship_name]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id_or_login}/linkedObjects/{relationship_name} disagrees with covered api_surface path /api/v1/users/{userIdOrLogin}/linkedObjects/{relationshipName}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id_or_login}/linkedObjects/{relationship_name} disagrees with covered api_surface path /api/v1/users/{userIdOrLogin}/linkedObjects/{relationshipName}.; flags: --relationship-name (required), --user-id-or-login (required)
    delete api v1 users user id roles role assignment id apply - Typed action delete_api_v1_users_user_id_roles_role_assignment_id [intent=reverse_etl availability=partial write=delete_api_v1_users_user_id_roles_role_assignment_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/roles/{role_assignment_id} disagrees with covered api_surface path /api/v1/users/{userId}/roles/{roleAssignmentId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/roles/{role_assignment_id} disagrees with covered api_surface path /api/v1/users/{userId}/roles/{roleAssignmentId}.; flags: --role-assignment-id (required), --user-id (required)
    delete api v1 users user id roles role assignment id targets catalog apps app name app id apply - Typed action delete_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id [intent=reverse_etl availability=partial write=delete_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name}/{app_id} disagrees with covered api_surface path /api/v1/users/{userId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}/{appId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name}/{app_id} disagrees with covered api_surface path /api/v1/users/{userId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}/{appId}.; flags: --app-id (required), --app-name (required), --role-assignment-id (required), --user-id (required)
    delete api v1 users user id roles role assignment id targets catalog apps app name apply - Typed action delete_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name [intent=reverse_etl availability=partial write=delete_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name} disagrees with covered api_surface path /api/v1/users/{userId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name} disagrees with covered api_surface path /api/v1/users/{userId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}.; flags: --app-name (required), --role-assignment-id (required), --user-id (required)
    delete api v1 users user id roles role assignment id targets groups group id apply - Typed action delete_api_v1_users_user_id_roles_role_assignment_id_targets_groups_group_id [intent=reverse_etl availability=partial write=delete_api_v1_users_user_id_roles_role_assignment_id_targets_groups_group_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/roles/{role_assignment_id}/targets/groups/{group_id} disagrees with covered api_surface path /api/v1/users/{userId}/roles/{roleAssignmentId}/targets/groups/{groupId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/roles/{role_assignment_id}/targets/groups/{group_id} disagrees with covered api_surface path /api/v1/users/{userId}/roles/{roleAssignmentId}/targets/groups/{groupId}.; flags: --group-id (required), --role-assignment-id (required), --user-id (required)
    delete api v1 users user id sessions apply - Typed action delete_api_v1_users_user_id_sessions [intent=reverse_etl availability=partial write=delete_api_v1_users_user_id_sessions]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/sessions disagrees with covered api_surface path /api/v1/users/{userId}/sessions.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/sessions disagrees with covered api_surface path /api/v1/users/{userId}/sessions.; flags: --user-id (required)
    delete api v1 zones zone id apply - Typed action delete_api_v1_zones_zone_id [intent=reverse_etl availability=partial write=delete_api_v1_zones_zone_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/zones/{zone_id} disagrees with covered api_surface path /api/v1/zones/{zoneId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/zones/{zone_id} disagrees with covered api_surface path /api/v1/zones/{zoneId}.; flags: --zone-id (required)
    delete integrations api v1 api services api service id apply - Typed action delete_integrations_api_v1_api_services_api_service_id [intent=reverse_etl availability=partial write=delete_integrations_api_v1_api_services_api_service_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /integrations/api/v1/api-services/{api_service_id} disagrees with covered api_surface path /integrations/api/v1/api-services/{apiServiceId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /integrations/api/v1/api-services/{api_service_id} disagrees with covered api_surface path /integrations/api/v1/api-services/{apiServiceId}.; flags: --api-service-id (required)
    delete integrations api v1 api services api service id credentials secrets secret id apply - Typed action delete_integrations_api_v1_api_services_api_service_id_credentials_secrets_secret_id [intent=reverse_etl availability=partial write=delete_integrations_api_v1_api_services_api_service_id_credentials_secrets_secret_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /integrations/api/v1/api-services/{api_service_id}/credentials/secrets/{secret_id} disagrees with covered api_surface path /integrations/api/v1/api-services/{apiServiceId}/credentials/secrets/{secretId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /integrations/api/v1/api-services/{api_service_id}/credentials/secrets/{secret_id} disagrees with covered api_surface path /integrations/api/v1/api-services/{apiServiceId}/credentials/secrets/{secretId}.; flags: --api-service-id (required), --secret-id (required)
    delete oauth2 v1 clients client id roles role assignment id apply - Typed action delete_oauth2_v1_clients_client_id_roles_role_assignment_id [intent=reverse_etl availability=partial write=delete_oauth2_v1_clients_client_id_roles_role_assignment_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /oauth2/v1/clients/{client_id}/roles/{role_assignment_id} disagrees with covered api_surface path /oauth2/v1/clients/{clientId}/roles/{roleAssignmentId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /oauth2/v1/clients/{client_id}/roles/{role_assignment_id} disagrees with covered api_surface path /oauth2/v1/clients/{clientId}/roles/{roleAssignmentId}.; flags: --client-id (required), --role-assignment-id (required)
    delete oauth2 v1 clients client id roles role assignment id targets catalog apps app name app id apply - Typed action delete_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id [intent=reverse_etl availability=partial write=delete_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /oauth2/v1/clients/{client_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name}/{app_id} disagrees with covered api_surface path /oauth2/v1/clients/{clientId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}/{appId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /oauth2/v1/clients/{client_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name}/{app_id} disagrees with covered api_surface path /oauth2/v1/clients/{clientId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}/{appId}.; flags: --app-id (required), --app-name (required), --client-id (required), --role-assignment-id (required)
    delete oauth2 v1 clients client id roles role assignment id targets catalog apps app name apply - Typed action delete_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name [intent=reverse_etl availability=partial write=delete_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /oauth2/v1/clients/{client_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name} disagrees with covered api_surface path /oauth2/v1/clients/{clientId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /oauth2/v1/clients/{client_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name} disagrees with covered api_surface path /oauth2/v1/clients/{clientId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}.; flags: --app-name (required), --client-id (required), --role-assignment-id (required)
    delete oauth2 v1 clients client id roles role assignment id targets groups group id apply - Typed action delete_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_groups_group_id [intent=reverse_etl availability=partial write=delete_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_groups_group_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /oauth2/v1/clients/{client_id}/roles/{role_assignment_id}/targets/groups/{group_id} disagrees with covered api_surface path /oauth2/v1/clients/{clientId}/roles/{roleAssignmentId}/targets/groups/{groupId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /oauth2/v1/clients/{client_id}/roles/{role_assignment_id}/targets/groups/{group_id} disagrees with covered api_surface path /oauth2/v1/clients/{clientId}/roles/{roleAssignmentId}/targets/groups/{groupId}.; flags: --client-id (required), --group-id (required), --role-assignment-id (required)
    delete privileged access api v1 okta service accounts id apply - DELETE /privileged-access/api/v1/okta-service-accounts/{id} (delete_privileged_access_api_v1_okta_service_accounts_id) [intent=reverse_etl availability=implemented write=delete_privileged_access_api_v1_okta_service_accounts_id]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    delete privileged access api v1 service accounts id apply - DELETE /privileged-access/api/v1/service-accounts/{id} (delete_privileged_access_api_v1_service_accounts_id) [intent=reverse_etl availability=implemented write=delete_privileged_access_api_v1_service_accounts_id]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    delete webauthn registration api v1 users user id enrollments authenticator enrollment id apply - Typed action delete_webauthn_registration_api_v1_users_user_id_enrollments_authenticator_enrollment_id [intent=reverse_etl availability=partial write=delete_webauthn_registration_api_v1_users_user_id_enrollments_authenticator_enrollment_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /webauthn-registration/api/v1/users/{user_id}/enrollments/{authenticator_enrollment_id} disagrees with covered api_surface path /webauthn-registration/api/v1/users/{userId}/enrollments/{authenticatorEnrollmentId}.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /webauthn-registration/api/v1/users/{user_id}/enrollments/{authenticator_enrollment_id} disagrees with covered api_surface path /webauthn-registration/api/v1/users/{userId}/enrollments/{authenticatorEnrollmentId}.; flags: --authenticator-enrollment-id (required), --user-id (required)
    execute api v1 agent pools pool id updates update id activate apply - Typed action execute_api_v1_agent_pools_pool_id_updates_update_id_activate [intent=reverse_etl availability=partial write=execute_api_v1_agent_pools_pool_id_updates_update_id_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates/{update_id}/activate disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates/{updateId}/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates/{update_id}/activate disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates/{updateId}/activate.; flags: --pool-id (required), --update-id (required)
    execute api v1 agent pools pool id updates update id deactivate apply - Typed action execute_api_v1_agent_pools_pool_id_updates_update_id_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_agent_pools_pool_id_updates_update_id_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates/{update_id}/deactivate disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates/{updateId}/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates/{update_id}/deactivate disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates/{updateId}/deactivate.; flags: --pool-id (required), --update-id (required)
    execute api v1 agent pools pool id updates update id pause apply - Typed action execute_api_v1_agent_pools_pool_id_updates_update_id_pause [intent=reverse_etl availability=partial write=execute_api_v1_agent_pools_pool_id_updates_update_id_pause]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates/{update_id}/pause disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates/{updateId}/pause.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates/{update_id}/pause disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates/{updateId}/pause.; flags: --pool-id (required), --update-id (required)
    execute api v1 agent pools pool id updates update id resume apply - Typed action execute_api_v1_agent_pools_pool_id_updates_update_id_resume [intent=reverse_etl availability=partial write=execute_api_v1_agent_pools_pool_id_updates_update_id_resume]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates/{update_id}/resume disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates/{updateId}/resume.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates/{update_id}/resume disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates/{updateId}/resume.; flags: --pool-id (required), --update-id (required)
    execute api v1 agent pools pool id updates update id retry apply - Typed action execute_api_v1_agent_pools_pool_id_updates_update_id_retry [intent=reverse_etl availability=partial write=execute_api_v1_agent_pools_pool_id_updates_update_id_retry]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates/{update_id}/retry disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates/{updateId}/retry.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates/{update_id}/retry disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates/{updateId}/retry.; flags: --pool-id (required), --update-id (required)
    execute api v1 agent pools pool id updates update id stop apply - Typed action execute_api_v1_agent_pools_pool_id_updates_update_id_stop [intent=reverse_etl availability=partial write=execute_api_v1_agent_pools_pool_id_updates_update_id_stop]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates/{update_id}/stop disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates/{updateId}/stop.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/agentPools/{pool_id}/updates/{update_id}/stop disagrees with covered api_surface path /api/v1/agentPools/{poolId}/updates/{updateId}/stop.; flags: --pool-id (required), --update-id (required)
    execute api v1 apps app id connections default lifecycle activate apply - Typed action execute_api_v1_apps_app_id_connections_default_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_apps_app_id_connections_default_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/connections/default/lifecycle/activate disagrees with covered api_surface path /api/v1/apps/{appId}/connections/default/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/connections/default/lifecycle/activate disagrees with covered api_surface path /api/v1/apps/{appId}/connections/default/lifecycle/activate.; flags: --app-id (required)
    execute api v1 apps app id connections default lifecycle deactivate apply - Typed action execute_api_v1_apps_app_id_connections_default_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_apps_app_id_connections_default_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/connections/default/lifecycle/deactivate disagrees with covered api_surface path /api/v1/apps/{appId}/connections/default/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/connections/default/lifecycle/deactivate disagrees with covered api_surface path /api/v1/apps/{appId}/connections/default/lifecycle/deactivate.; flags: --app-id (required)
    execute api v1 apps app id credentials jwks key id lifecycle activate apply - Typed action execute_api_v1_apps_app_id_credentials_jwks_key_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_apps_app_id_credentials_jwks_key_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/jwks/{key_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/jwks/{keyId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/jwks/{key_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/jwks/{keyId}/lifecycle/activate.; flags: --app-id (required), --key-id (required)
    execute api v1 apps app id credentials jwks key id lifecycle deactivate apply - Typed action execute_api_v1_apps_app_id_credentials_jwks_key_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_apps_app_id_credentials_jwks_key_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/jwks/{key_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/jwks/{keyId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/jwks/{key_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/jwks/{keyId}/lifecycle/deactivate.; flags: --app-id (required), --key-id (required)
    execute api v1 apps app id credentials secrets secret id lifecycle activate apply - Typed action execute_api_v1_apps_app_id_credentials_secrets_secret_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_apps_app_id_credentials_secrets_secret_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/secrets/{secret_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/secrets/{secretId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/secrets/{secret_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/secrets/{secretId}/lifecycle/activate.; flags: --app-id (required), --secret-id (required)
    execute api v1 apps app id credentials secrets secret id lifecycle deactivate apply - Typed action execute_api_v1_apps_app_id_credentials_secrets_secret_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_apps_app_id_credentials_secrets_secret_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/secrets/{secret_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/secrets/{secretId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/credentials/secrets/{secret_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/apps/{appId}/credentials/secrets/{secretId}/lifecycle/deactivate.; flags: --app-id (required), --secret-id (required)
    execute api v1 apps app id lifecycle activate apply - Typed action execute_api_v1_apps_app_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_apps_app_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/apps/{appId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/apps/{appId}/lifecycle/activate.; flags: --app-id (required)
    execute api v1 apps app id lifecycle deactivate apply - Typed action execute_api_v1_apps_app_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_apps_app_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/apps/{appId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/apps/{appId}/lifecycle/deactivate.; flags: --app-id (required)
    execute api v1 authenticators authenticator id lifecycle activate apply - Typed action execute_api_v1_authenticators_authenticator_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_authenticators_authenticator_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/lifecycle/activate.; flags: --authenticator-id (required)
    execute api v1 authenticators authenticator id lifecycle deactivate apply - Typed action execute_api_v1_authenticators_authenticator_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_authenticators_authenticator_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/lifecycle/deactivate.; flags: --authenticator-id (required)
    execute api v1 authenticators authenticator id methods method type lifecycle activate apply - Typed action execute_api_v1_authenticators_authenticator_id_methods_method_type_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_authenticators_authenticator_id_methods_method_type_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/methods/{method_type}/lifecycle/activate disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/methods/{methodType}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/methods/{method_type}/lifecycle/activate disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/methods/{methodType}/lifecycle/activate.; flags: --authenticator-id (required), --method-type (required)
    execute api v1 authenticators authenticator id methods method type lifecycle deactivate apply - Typed action execute_api_v1_authenticators_authenticator_id_methods_method_type_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_authenticators_authenticator_id_methods_method_type_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/methods/{method_type}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/methods/{methodType}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/methods/{method_type}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/methods/{methodType}/lifecycle/deactivate.; flags: --authenticator-id (required), --method-type (required)
    execute api v1 authorization servers auth server id credentials lifecycle key rotate apply - Typed action execute_api_v1_authorization_servers_auth_server_id_credentials_lifecycle_key_rotate [intent=reverse_etl availability=partial write=execute_api_v1_authorization_servers_auth_server_id_credentials_lifecycle_key_rotate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/credentials/lifecycle/keyRotate disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/credentials/lifecycle/keyRotate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/credentials/lifecycle/keyRotate disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/credentials/lifecycle/keyRotate.; flags: --auth-server-id (required)
    execute api v1 authorization servers auth server id lifecycle activate apply - Typed action execute_api_v1_authorization_servers_auth_server_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_authorization_servers_auth_server_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/lifecycle/activate.; flags: --auth-server-id (required)
    execute api v1 authorization servers auth server id lifecycle deactivate apply - Typed action execute_api_v1_authorization_servers_auth_server_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_authorization_servers_auth_server_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/lifecycle/deactivate.; flags: --auth-server-id (required)
    execute api v1 authorization servers auth server id policies policy id lifecycle activate apply - Typed action execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies/{policy_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies/{policyId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies/{policy_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies/{policyId}/lifecycle/activate.; flags: --auth-server-id (required), --policy-id (required)
    execute api v1 authorization servers auth server id policies policy id lifecycle deactivate apply - Typed action execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies/{policy_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies/{policyId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies/{policy_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies/{policyId}/lifecycle/deactivate.; flags: --auth-server-id (required), --policy-id (required)
    execute api v1 authorization servers auth server id policies policy id rules rule id lifecycle activate apply - Typed action execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies/{policy_id}/rules/{rule_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies/{policyId}/rules/{ruleId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies/{policy_id}/rules/{rule_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies/{policyId}/rules/{ruleId}/lifecycle/activate.; flags: --auth-server-id (required), --policy-id (required), --rule-id (required)
    execute api v1 authorization servers auth server id policies policy id rules rule id lifecycle deactivate apply - Typed action execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies/{policy_id}/rules/{rule_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies/{policyId}/rules/{ruleId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies/{policy_id}/rules/{rule_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies/{policyId}/rules/{ruleId}/lifecycle/deactivate.; flags: --auth-server-id (required), --policy-id (required), --rule-id (required)
    execute api v1 authorization servers auth server id resourceservercredentials keys key id lifecycle activate apply - Typed action execute_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/resourceservercredentials/keys/{key_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/resourceservercredentials/keys/{keyId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/resourceservercredentials/keys/{key_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/resourceservercredentials/keys/{keyId}/lifecycle/activate.; flags: --auth-server-id (required), --key-id (required)
    execute api v1 authorization servers auth server id resourceservercredentials keys key id lifecycle deactivate apply - Typed action execute_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/resourceservercredentials/keys/{key_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/resourceservercredentials/keys/{keyId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/resourceservercredentials/keys/{key_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/resourceservercredentials/keys/{keyId}/lifecycle/deactivate.; flags: --auth-server-id (required), --key-id (required)
    execute api v1 behaviors behavior id lifecycle activate apply - Typed action execute_api_v1_behaviors_behavior_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_behaviors_behavior_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/behaviors/{behavior_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/behaviors/{behaviorId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/behaviors/{behavior_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/behaviors/{behaviorId}/lifecycle/activate.; flags: --behavior-id (required)
    execute api v1 behaviors behavior id lifecycle deactivate apply - Typed action execute_api_v1_behaviors_behavior_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_behaviors_behavior_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/behaviors/{behavior_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/behaviors/{behaviorId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/behaviors/{behavior_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/behaviors/{behaviorId}/lifecycle/deactivate.; flags: --behavior-id (required)
    execute api v1 device integrations device integration id lifecycle activate apply - Typed action execute_api_v1_device_integrations_device_integration_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_device_integrations_device_integration_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/device-integrations/{device_integration_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/device-integrations/{deviceIntegrationId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/device-integrations/{device_integration_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/device-integrations/{deviceIntegrationId}/lifecycle/activate.; flags: --device-integration-id (required)
    execute api v1 device integrations device integration id lifecycle deactivate apply - Typed action execute_api_v1_device_integrations_device_integration_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_device_integrations_device_integration_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/device-integrations/{device_integration_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/device-integrations/{deviceIntegrationId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/device-integrations/{device_integration_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/device-integrations/{deviceIntegrationId}/lifecycle/deactivate.; flags: --device-integration-id (required)
    execute api v1 devices device id lifecycle activate apply - Typed action execute_api_v1_devices_device_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_devices_device_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/devices/{device_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/devices/{deviceId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/devices/{device_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/devices/{deviceId}/lifecycle/activate.; flags: --device-id (required)
    execute api v1 devices device id lifecycle deactivate apply - Typed action execute_api_v1_devices_device_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_devices_device_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/devices/{device_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/devices/{deviceId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/devices/{device_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/devices/{deviceId}/lifecycle/deactivate.; flags: --device-id (required)
    execute api v1 devices device id lifecycle suspend apply - Typed action execute_api_v1_devices_device_id_lifecycle_suspend [intent=reverse_etl availability=partial write=execute_api_v1_devices_device_id_lifecycle_suspend]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/devices/{device_id}/lifecycle/suspend disagrees with covered api_surface path /api/v1/devices/{deviceId}/lifecycle/suspend.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/devices/{device_id}/lifecycle/suspend disagrees with covered api_surface path /api/v1/devices/{deviceId}/lifecycle/suspend.; flags: --device-id (required)
    execute api v1 devices device id lifecycle unsuspend apply - Typed action execute_api_v1_devices_device_id_lifecycle_unsuspend [intent=reverse_etl availability=partial write=execute_api_v1_devices_device_id_lifecycle_unsuspend]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/devices/{device_id}/lifecycle/unsuspend disagrees with covered api_surface path /api/v1/devices/{deviceId}/lifecycle/unsuspend.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/devices/{device_id}/lifecycle/unsuspend disagrees with covered api_surface path /api/v1/devices/{deviceId}/lifecycle/unsuspend.; flags: --device-id (required)
    execute api v1 event hooks event hook id lifecycle activate apply - Typed action execute_api_v1_event_hooks_event_hook_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_event_hooks_event_hook_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/eventHooks/{event_hook_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/eventHooks/{eventHookId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/eventHooks/{event_hook_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/eventHooks/{eventHookId}/lifecycle/activate.; flags: --event-hook-id (required)
    execute api v1 event hooks event hook id lifecycle deactivate apply - Typed action execute_api_v1_event_hooks_event_hook_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_event_hooks_event_hook_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/eventHooks/{event_hook_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/eventHooks/{eventHookId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/eventHooks/{event_hook_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/eventHooks/{eventHookId}/lifecycle/deactivate.; flags: --event-hook-id (required)
    execute api v1 event hooks event hook id lifecycle verify apply - Typed action execute_api_v1_event_hooks_event_hook_id_lifecycle_verify [intent=reverse_etl availability=partial write=execute_api_v1_event_hooks_event_hook_id_lifecycle_verify]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/eventHooks/{event_hook_id}/lifecycle/verify disagrees with covered api_surface path /api/v1/eventHooks/{eventHookId}/lifecycle/verify.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/eventHooks/{event_hook_id}/lifecycle/verify disagrees with covered api_surface path /api/v1/eventHooks/{eventHookId}/lifecycle/verify.; flags: --event-hook-id (required)
    execute api v1 groups rules group rule id lifecycle activate apply - Typed action execute_api_v1_groups_rules_group_rule_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_groups_rules_group_rule_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/groups/rules/{group_rule_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/groups/rules/{groupRuleId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/groups/rules/{group_rule_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/groups/rules/{groupRuleId}/lifecycle/activate.; flags: --group-rule-id (required)
    execute api v1 groups rules group rule id lifecycle deactivate apply - Typed action execute_api_v1_groups_rules_group_rule_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_groups_rules_group_rule_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/groups/rules/{group_rule_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/groups/rules/{groupRuleId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/groups/rules/{group_rule_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/groups/rules/{groupRuleId}/lifecycle/deactivate.; flags: --group-rule-id (required)
    execute api v1 idps idp id lifecycle activate apply - Typed action execute_api_v1_idps_idp_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_idps_idp_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/idps/{idp_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/idps/{idpId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/idps/{idp_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/idps/{idpId}/lifecycle/activate.; flags: --idp-id (required)
    execute api v1 idps idp id lifecycle deactivate apply - Typed action execute_api_v1_idps_idp_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_idps_idp_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/idps/{idp_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/idps/{idpId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/idps/{idp_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/idps/{idpId}/lifecycle/deactivate.; flags: --idp-id (required)
    execute api v1 inline hooks inline hook id lifecycle activate apply - Typed action execute_api_v1_inline_hooks_inline_hook_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_inline_hooks_inline_hook_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/inlineHooks/{inline_hook_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/inlineHooks/{inlineHookId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/inlineHooks/{inline_hook_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/inlineHooks/{inlineHookId}/lifecycle/activate.; flags: --inline-hook-id (required)
    execute api v1 inline hooks inline hook id lifecycle deactivate apply - Typed action execute_api_v1_inline_hooks_inline_hook_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_inline_hooks_inline_hook_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/inlineHooks/{inline_hook_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/inlineHooks/{inlineHookId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/inlineHooks/{inline_hook_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/inlineHooks/{inlineHookId}/lifecycle/deactivate.; flags: --inline-hook-id (required)
    execute api v1 log streams log stream id lifecycle activate apply - Typed action execute_api_v1_log_streams_log_stream_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_log_streams_log_stream_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/logStreams/{log_stream_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/logStreams/{logStreamId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/logStreams/{log_stream_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/logStreams/{logStreamId}/lifecycle/activate.; flags: --log-stream-id (required)
    execute api v1 log streams log stream id lifecycle deactivate apply - Typed action execute_api_v1_log_streams_log_stream_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_log_streams_log_stream_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/logStreams/{log_stream_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/logStreams/{logStreamId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/logStreams/{log_stream_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/logStreams/{logStreamId}/lifecycle/deactivate.; flags: --log-stream-id (required)
    execute api v1 org privacy aerial revoke apply - POST /api/v1/org/privacy/aerial/revoke (execute_api_v1_org_privacy_aerial_revoke) [intent=reverse_etl availability=implemented write=execute_api_v1_org_privacy_aerial_revoke]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --account-id (required)
    execute api v1 org privacy okta support revoke apply - POST /api/v1/org/privacy/oktaSupport/revoke (execute_api_v1_org_privacy_okta_support_revoke) [intent=reverse_etl availability=implemented write=execute_api_v1_org_privacy_okta_support_revoke]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    execute api v1 policies policy id lifecycle activate apply - Typed action execute_api_v1_policies_policy_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_policies_policy_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/policies/{policyId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/policies/{policyId}/lifecycle/activate.; flags: --policy-id (required)
    execute api v1 policies policy id lifecycle deactivate apply - Typed action execute_api_v1_policies_policy_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_policies_policy_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/policies/{policyId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/policies/{policyId}/lifecycle/deactivate.; flags: --policy-id (required)
    execute api v1 policies policy id rules rule id lifecycle activate apply - Typed action execute_api_v1_policies_policy_id_rules_rule_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_policies_policy_id_rules_rule_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/rules/{rule_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/policies/{policyId}/rules/{ruleId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/rules/{rule_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/policies/{policyId}/rules/{ruleId}/lifecycle/activate.; flags: --policy-id (required), --rule-id (required)
    execute api v1 policies policy id rules rule id lifecycle deactivate apply - Typed action execute_api_v1_policies_policy_id_rules_rule_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_policies_policy_id_rules_rule_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/rules/{rule_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/policies/{policyId}/rules/{ruleId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/rules/{rule_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/policies/{policyId}/rules/{ruleId}/lifecycle/deactivate.; flags: --policy-id (required), --rule-id (required)
    execute api v1 realm assignments assignment id lifecycle activate apply - Typed action execute_api_v1_realm_assignments_assignment_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_realm_assignments_assignment_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/realm-assignments/{assignment_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/realm-assignments/{assignmentId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/realm-assignments/{assignment_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/realm-assignments/{assignmentId}/lifecycle/activate.; flags: --assignment-id (required)
    execute api v1 realm assignments assignment id lifecycle deactivate apply - Typed action execute_api_v1_realm_assignments_assignment_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_realm_assignments_assignment_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/realm-assignments/{assignment_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/realm-assignments/{assignmentId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/realm-assignments/{assignment_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/realm-assignments/{assignmentId}/lifecycle/deactivate.; flags: --assignment-id (required)
    execute api v1 security events providers security event provider id lifecycle activate apply - Typed action execute_api_v1_security_events_providers_security_event_provider_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_security_events_providers_security_event_provider_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/security-events-providers/{security_event_provider_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/security-events-providers/{securityEventProviderId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/security-events-providers/{security_event_provider_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/security-events-providers/{securityEventProviderId}/lifecycle/activate.; flags: --security-event-provider-id (required)
    execute api v1 security events providers security event provider id lifecycle deactivate apply - Typed action execute_api_v1_security_events_providers_security_event_provider_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_security_events_providers_security_event_provider_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/security-events-providers/{security_event_provider_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/security-events-providers/{securityEventProviderId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/security-events-providers/{security_event_provider_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/security-events-providers/{securityEventProviderId}/lifecycle/deactivate.; flags: --security-event-provider-id (required)
    execute api v1 sessions session id lifecycle refresh apply - Typed action execute_api_v1_sessions_session_id_lifecycle_refresh [intent=reverse_etl availability=partial write=execute_api_v1_sessions_session_id_lifecycle_refresh]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/sessions/{session_id}/lifecycle/refresh disagrees with covered api_surface path /api/v1/sessions/{sessionId}/lifecycle/refresh.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/sessions/{session_id}/lifecycle/refresh disagrees with covered api_surface path /api/v1/sessions/{sessionId}/lifecycle/refresh.; flags: --session-id (required)
    execute api v1 telephony providers custom telephony provider id lifecycle activate apply - Typed action execute_api_v1_telephony_providers_custom_telephony_provider_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_telephony_providers_custom_telephony_provider_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/telephony-providers/{custom_telephony_provider_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/telephony-providers/{customTelephonyProviderId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/telephony-providers/{custom_telephony_provider_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/telephony-providers/{customTelephonyProviderId}/lifecycle/activate.; flags: --custom-telephony-provider-id (required)
    execute api v1 telephony providers custom telephony provider id lifecycle deactivate apply - Typed action execute_api_v1_telephony_providers_custom_telephony_provider_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_telephony_providers_custom_telephony_provider_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/telephony-providers/{custom_telephony_provider_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/telephony-providers/{customTelephonyProviderId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/telephony-providers/{custom_telephony_provider_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/telephony-providers/{customTelephonyProviderId}/lifecycle/deactivate.; flags: --custom-telephony-provider-id (required)
    execute api v1 trusted origins trusted origin id lifecycle activate apply - Typed action execute_api_v1_trusted_origins_trusted_origin_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_trusted_origins_trusted_origin_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/trustedOrigins/{trusted_origin_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/trustedOrigins/{trustedOriginId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/trustedOrigins/{trusted_origin_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/trustedOrigins/{trustedOriginId}/lifecycle/activate.; flags: --trusted-origin-id (required)
    execute api v1 trusted origins trusted origin id lifecycle deactivate apply - Typed action execute_api_v1_trusted_origins_trusted_origin_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_trusted_origins_trusted_origin_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/trustedOrigins/{trusted_origin_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/trustedOrigins/{trustedOriginId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/trustedOrigins/{trusted_origin_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/trustedOrigins/{trustedOriginId}/lifecycle/deactivate.; flags: --trusted-origin-id (required)
    execute api v1 users id lifecycle activate apply - POST /api/v1/users/{id}/lifecycle/activate (execute_api_v1_users_id_lifecycle_activate) [intent=reverse_etl availability=implemented write=execute_api_v1_users_id_lifecycle_activate]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    execute api v1 users id lifecycle deactivate apply - POST /api/v1/users/{id}/lifecycle/deactivate (execute_api_v1_users_id_lifecycle_deactivate) [intent=reverse_etl availability=implemented write=execute_api_v1_users_id_lifecycle_deactivate]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    execute api v1 users id lifecycle expire password apply - POST /api/v1/users/{id}/lifecycle/expire_password (execute_api_v1_users_id_lifecycle_expire_password) [intent=reverse_etl availability=implemented write=execute_api_v1_users_id_lifecycle_expire_password]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    execute api v1 users id lifecycle expire password with temp password apply - POST /api/v1/users/{id}/lifecycle/expire_password_with_temp_password (execute_api_v1_users_id_lifecycle_expire_password_with_temp_password) [intent=reverse_etl availability=implemented write=execute_api_v1_users_id_lifecycle_expire_password_with_temp_password]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    execute api v1 users id lifecycle reactivate apply - POST /api/v1/users/{id}/lifecycle/reactivate (execute_api_v1_users_id_lifecycle_reactivate) [intent=reverse_etl availability=implemented write=execute_api_v1_users_id_lifecycle_reactivate]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    execute api v1 users id lifecycle reset factors apply - POST /api/v1/users/{id}/lifecycle/reset_factors (execute_api_v1_users_id_lifecycle_reset_factors) [intent=reverse_etl availability=implemented write=execute_api_v1_users_id_lifecycle_reset_factors]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    execute api v1 users id lifecycle suspend apply - POST /api/v1/users/{id}/lifecycle/suspend (execute_api_v1_users_id_lifecycle_suspend) [intent=reverse_etl availability=implemented write=execute_api_v1_users_id_lifecycle_suspend]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    execute api v1 users id lifecycle unlock apply - POST /api/v1/users/{id}/lifecycle/unlock (execute_api_v1_users_id_lifecycle_unlock) [intent=reverse_etl availability=implemented write=execute_api_v1_users_id_lifecycle_unlock]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    execute api v1 users id lifecycle unsuspend apply - POST /api/v1/users/{id}/lifecycle/unsuspend (execute_api_v1_users_id_lifecycle_unsuspend) [intent=reverse_etl availability=implemented write=execute_api_v1_users_id_lifecycle_unsuspend]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    execute api v1 users user id factors factor id lifecycle activate apply - Typed action execute_api_v1_users_user_id_factors_factor_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_users_user_id_factors_factor_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/factors/{factor_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/users/{userId}/factors/{factorId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/factors/{factor_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/users/{userId}/factors/{factorId}/lifecycle/activate.; flags: --factor-id (required), --user-id (required)
    execute api v1 zones zone id lifecycle activate apply - Typed action execute_api_v1_zones_zone_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_api_v1_zones_zone_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/zones/{zone_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/zones/{zoneId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/zones/{zone_id}/lifecycle/activate disagrees with covered api_surface path /api/v1/zones/{zoneId}/lifecycle/activate.; flags: --zone-id (required)
    execute api v1 zones zone id lifecycle deactivate apply - Typed action execute_api_v1_zones_zone_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_api_v1_zones_zone_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/zones/{zone_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/zones/{zoneId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/zones/{zone_id}/lifecycle/deactivate disagrees with covered api_surface path /api/v1/zones/{zoneId}/lifecycle/deactivate.; flags: --zone-id (required)
    execute integrations api v1 api services api service id credentials secrets secret id lifecycle activate apply - Typed action execute_integrations_api_v1_api_services_api_service_id_credentials_secrets_secret_id_lifecycle_activate [intent=reverse_etl availability=partial write=execute_integrations_api_v1_api_services_api_service_id_credentials_secrets_secret_id_lifecycle_activate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /integrations/api/v1/api-services/{api_service_id}/credentials/secrets/{secret_id}/lifecycle/activate disagrees with covered api_surface path /integrations/api/v1/api-services/{apiServiceId}/credentials/secrets/{secretId}/lifecycle/activate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /integrations/api/v1/api-services/{api_service_id}/credentials/secrets/{secret_id}/lifecycle/activate disagrees with covered api_surface path /integrations/api/v1/api-services/{apiServiceId}/credentials/secrets/{secretId}/lifecycle/activate.; flags: --api-service-id (required), --secret-id (required)
    execute integrations api v1 api services api service id credentials secrets secret id lifecycle deactivate apply - Typed action execute_integrations_api_v1_api_services_api_service_id_credentials_secrets_secret_id_lifecycle_deactivate [intent=reverse_etl availability=partial write=execute_integrations_api_v1_api_services_api_service_id_credentials_secrets_secret_id_lifecycle_deactivate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /integrations/api/v1/api-services/{api_service_id}/credentials/secrets/{secret_id}/lifecycle/deactivate disagrees with covered api_surface path /integrations/api/v1/api-services/{apiServiceId}/credentials/secrets/{secretId}/lifecycle/deactivate.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /integrations/api/v1/api-services/{api_service_id}/credentials/secrets/{secret_id}/lifecycle/deactivate disagrees with covered api_surface path /integrations/api/v1/api-services/{apiServiceId}/credentials/secrets/{secretId}/lifecycle/deactivate.; flags: --api-service-id (required), --secret-id (required)
    execute webauthn registration api v1 activate apply - POST /webauthn-registration/api/v1/activate (execute_webauthn_registration_api_v1_activate) [intent=reverse_etl availability=implemented write=execute_webauthn_registration_api_v1_activate]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: high: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    update api v1 api tokens api token id apply - Typed action update_api_v1_api_tokens_api_token_id [intent=reverse_etl availability=partial write=update_api_v1_api_tokens_api_token_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/api-tokens/{api_token_id} disagrees with covered api_surface path /api/v1/api-tokens/{apiTokenId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/api-tokens/{api_token_id} disagrees with covered api_surface path /api/v1/api-tokens/{apiTokenId}.; flags: --api-token-id (required)
    update api v1 apps app id apply - Typed action update_api_v1_apps_app_id [intent=reverse_etl availability=partial write=update_api_v1_apps_app_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id} disagrees with covered api_surface path /api/v1/apps/{appId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id} disagrees with covered api_surface path /api/v1/apps/{appId}.; flags: --app-id (required), --label (required), --sign-on-mode (required)
    update api v1 apps app id cwo connections connection id apply - Typed action update_api_v1_apps_app_id_cwo_connections_connection_id [intent=reverse_etl availability=partial write=update_api_v1_apps_app_id_cwo_connections_connection_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/cwo/connections/{connection_id} disagrees with covered api_surface path /api/v1/apps/{appId}/cwo/connections/{connectionId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/cwo/connections/{connection_id} disagrees with covered api_surface path /api/v1/apps/{appId}/cwo/connections/{connectionId}.; flags: --app-id (required), --connection-id (required), --status (required)
    update api v1 apps app id features feature name apply - Typed action update_api_v1_apps_app_id_features_feature_name [intent=reverse_etl availability=partial write=update_api_v1_apps_app_id_features_feature_name]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/features/{feature_name} disagrees with covered api_surface path /api/v1/apps/{appId}/features/{featureName}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/features/{feature_name} disagrees with covered api_surface path /api/v1/apps/{appId}/features/{featureName}.; flags: --app-id (required), --feature-name (required)
    update api v1 apps app id federated claims claim id apply - Typed action update_api_v1_apps_app_id_federated_claims_claim_id [intent=reverse_etl availability=partial write=update_api_v1_apps_app_id_federated_claims_claim_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/federated-claims/{claim_id} disagrees with covered api_surface path /api/v1/apps/{appId}/federated-claims/{claimId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/federated-claims/{claim_id} disagrees with covered api_surface path /api/v1/apps/{appId}/federated-claims/{claimId}.; flags: --app-id (required), --claim-id (required)
    update api v1 apps app id group push mappings mapping id apply - Typed action update_api_v1_apps_app_id_group_push_mappings_mapping_id [intent=reverse_etl availability=partial write=update_api_v1_apps_app_id_group_push_mappings_mapping_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/group-push/mappings/{mapping_id} disagrees with covered api_surface path /api/v1/apps/{appId}/group-push/mappings/{mappingId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/group-push/mappings/{mapping_id} disagrees with covered api_surface path /api/v1/apps/{appId}/group-push/mappings/{mappingId}.; flags: --app-id (required), --mapping-id (required), --status (required)
    update api v1 apps app id groups group id apply - Typed action update_api_v1_apps_app_id_groups_group_id [intent=reverse_etl availability=partial write=update_api_v1_apps_app_id_groups_group_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/groups/{group_id} disagrees with covered api_surface path /api/v1/apps/{appId}/groups/{groupId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/groups/{group_id} disagrees with covered api_surface path /api/v1/apps/{appId}/groups/{groupId}.; flags: --app-id (required), --group-id (required)
    update api v1 apps app id policies policy id apply - Typed action update_api_v1_apps_app_id_policies_policy_id [intent=reverse_etl availability=partial write=update_api_v1_apps_app_id_policies_policy_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/apps/{app_id}/policies/{policy_id} disagrees with covered api_surface path /api/v1/apps/{appId}/policies/{policyId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/apps/{app_id}/policies/{policy_id} disagrees with covered api_surface path /api/v1/apps/{appId}/policies/{policyId}.; flags: --app-id (required), --policy-id (required)
    update api v1 authenticators authenticator id aaguids aaguid 2 apply - Typed action update_api_v1_authenticators_authenticator_id_aaguids_aaguid_2 [intent=reverse_etl availability=partial write=update_api_v1_authenticators_authenticator_id_aaguids_aaguid_2]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/aaguids/{aaguid} disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/aaguids/{aaguid}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/aaguids/{aaguid} disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/aaguids/{aaguid}.; flags: --aaguid (required), --authenticator-id (required)
    update api v1 authenticators authenticator id aaguids aaguid apply - Typed action update_api_v1_authenticators_authenticator_id_aaguids_aaguid [intent=reverse_etl availability=partial write=update_api_v1_authenticators_authenticator_id_aaguids_aaguid]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/aaguids/{aaguid} disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/aaguids/{aaguid}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/aaguids/{aaguid} disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/aaguids/{aaguid}.; flags: --aaguid (required), --authenticator-id (required)
    update api v1 authenticators authenticator id apply - Typed action update_api_v1_authenticators_authenticator_id [intent=reverse_etl availability=partial write=update_api_v1_authenticators_authenticator_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id} disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id} disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}.; flags: --authenticator-id (required)
    update api v1 authenticators authenticator id methods method type apply - Typed action update_api_v1_authenticators_authenticator_id_methods_method_type [intent=reverse_etl availability=partial write=update_api_v1_authenticators_authenticator_id_methods_method_type]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/methods/{method_type} disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/methods/{methodType}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authenticators/{authenticator_id}/methods/{method_type} disagrees with covered api_surface path /api/v1/authenticators/{authenticatorId}/methods/{methodType}.; flags: --authenticator-id (required), --method-type (required)
    update api v1 authorization servers auth server id apply - Typed action update_api_v1_authorization_servers_auth_server_id [intent=reverse_etl availability=partial write=update_api_v1_authorization_servers_auth_server_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}.; flags: --auth-server-id (required)
    update api v1 authorization servers auth server id claims claim id apply - Typed action update_api_v1_authorization_servers_auth_server_id_claims_claim_id [intent=reverse_etl availability=partial write=update_api_v1_authorization_servers_auth_server_id_claims_claim_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/claims/{claim_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/claims/{claimId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/claims/{claim_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/claims/{claimId}.; flags: --auth-server-id (required), --claim-id (required)
    update api v1 authorization servers auth server id policies policy id apply - Typed action update_api_v1_authorization_servers_auth_server_id_policies_policy_id [intent=reverse_etl availability=partial write=update_api_v1_authorization_servers_auth_server_id_policies_policy_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies/{policy_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies/{policyId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies/{policy_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies/{policyId}.; flags: --auth-server-id (required), --policy-id (required)
    update api v1 authorization servers auth server id policies policy id rules rule id apply - Typed action update_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id [intent=reverse_etl availability=partial write=update_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies/{policy_id}/rules/{rule_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies/{policyId}/rules/{ruleId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/policies/{policy_id}/rules/{rule_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/policies/{policyId}/rules/{ruleId}.; flags: --auth-server-id (required), --conditions (required), --name (required), --policy-id (required), --rule-id (required), --type (required)
    update api v1 authorization servers auth server id scopes scope id apply - Typed action update_api_v1_authorization_servers_auth_server_id_scopes_scope_id [intent=reverse_etl availability=partial write=update_api_v1_authorization_servers_auth_server_id_scopes_scope_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/scopes/{scope_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/scopes/{scopeId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/authorizationServers/{auth_server_id}/scopes/{scope_id} disagrees with covered api_surface path /api/v1/authorizationServers/{authServerId}/scopes/{scopeId}.; flags: --auth-server-id (required), --name (required), --scope-id (required)
    update api v1 behaviors behavior id apply - Typed action update_api_v1_behaviors_behavior_id [intent=reverse_etl availability=partial write=update_api_v1_behaviors_behavior_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/behaviors/{behavior_id} disagrees with covered api_surface path /api/v1/behaviors/{behaviorId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/behaviors/{behavior_id} disagrees with covered api_surface path /api/v1/behaviors/{behaviorId}.; flags: --behavior-id (required), --name (required), --type (required)
    update api v1 brands brand id apply - Typed action update_api_v1_brands_brand_id [intent=reverse_etl availability=partial write=update_api_v1_brands_brand_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id} disagrees with covered api_surface path /api/v1/brands/{brandId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id} disagrees with covered api_surface path /api/v1/brands/{brandId}.; flags: --brand-id (required), --name (required)
    update api v1 brands brand id pages error customized apply - Typed action update_api_v1_brands_brand_id_pages_error_customized [intent=reverse_etl availability=partial write=update_api_v1_brands_brand_id_pages_error_customized]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/pages/error/customized disagrees with covered api_surface path /api/v1/brands/{brandId}/pages/error/customized.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/pages/error/customized disagrees with covered api_surface path /api/v1/brands/{brandId}/pages/error/customized.; flags: --brand-id (required)
    update api v1 brands brand id pages error preview apply - Typed action update_api_v1_brands_brand_id_pages_error_preview [intent=reverse_etl availability=partial write=update_api_v1_brands_brand_id_pages_error_preview]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/pages/error/preview disagrees with covered api_surface path /api/v1/brands/{brandId}/pages/error/preview.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/pages/error/preview disagrees with covered api_surface path /api/v1/brands/{brandId}/pages/error/preview.; flags: --brand-id (required)
    update api v1 brands brand id pages sign in customized apply - Typed action update_api_v1_brands_brand_id_pages_sign_in_customized [intent=reverse_etl availability=partial write=update_api_v1_brands_brand_id_pages_sign_in_customized]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/pages/sign-in/customized disagrees with covered api_surface path /api/v1/brands/{brandId}/pages/sign-in/customized.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/pages/sign-in/customized disagrees with covered api_surface path /api/v1/brands/{brandId}/pages/sign-in/customized.; flags: --brand-id (required)
    update api v1 brands brand id pages sign in preview apply - Typed action update_api_v1_brands_brand_id_pages_sign_in_preview [intent=reverse_etl availability=partial write=update_api_v1_brands_brand_id_pages_sign_in_preview]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/pages/sign-in/preview disagrees with covered api_surface path /api/v1/brands/{brandId}/pages/sign-in/preview.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/pages/sign-in/preview disagrees with covered api_surface path /api/v1/brands/{brandId}/pages/sign-in/preview.; flags: --brand-id (required)
    update api v1 brands brand id pages sign out customized apply - Typed action update_api_v1_brands_brand_id_pages_sign_out_customized [intent=reverse_etl availability=partial write=update_api_v1_brands_brand_id_pages_sign_out_customized]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/pages/sign-out/customized disagrees with covered api_surface path /api/v1/brands/{brandId}/pages/sign-out/customized.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/pages/sign-out/customized disagrees with covered api_surface path /api/v1/brands/{brandId}/pages/sign-out/customized.; flags: --brand-id (required), --type (required)
    update api v1 brands brand id templates email template name customizations customization id apply - Typed action update_api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id [intent=reverse_etl availability=partial write=update_api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/templates/email/{template_name}/customizations/{customization_id} disagrees with covered api_surface path /api/v1/brands/{brandId}/templates/email/{templateName}/customizations/{customizationId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/templates/email/{template_name}/customizations/{customization_id} disagrees with covered api_surface path /api/v1/brands/{brandId}/templates/email/{templateName}/customizations/{customizationId}.; flags: --brand-id (required), --customization-id (required), --language (required), --template-name (required)
    update api v1 brands brand id templates email template name settings apply - Typed action update_api_v1_brands_brand_id_templates_email_template_name_settings [intent=reverse_etl availability=partial write=update_api_v1_brands_brand_id_templates_email_template_name_settings]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/templates/email/{template_name}/settings disagrees with covered api_surface path /api/v1/brands/{brandId}/templates/email/{templateName}/settings.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/templates/email/{template_name}/settings disagrees with covered api_surface path /api/v1/brands/{brandId}/templates/email/{templateName}/settings.; flags: --brand-id (required), --recipients (required), --template-name (required)
    update api v1 brands brand id themes theme id apply - Typed action update_api_v1_brands_brand_id_themes_theme_id [intent=reverse_etl availability=partial write=update_api_v1_brands_brand_id_themes_theme_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/themes/{theme_id} disagrees with covered api_surface path /api/v1/brands/{brandId}/themes/{themeId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/themes/{theme_id} disagrees with covered api_surface path /api/v1/brands/{brandId}/themes/{themeId}.; flags: --brand-id (required), --email-template-touch-point-variant (required), --end-user-dashboard-touch-point-variant (required), --error-page-touch-point-variant (required), --primary-color-hex (required), --secondary-color-hex (required), --sign-in-page-touch-point-variant (required), --theme-id (required)
    update api v1 brands brand id well known uris path customized apply - Typed action update_api_v1_brands_brand_id_well_known_uris_path_customized [intent=reverse_etl availability=partial write=update_api_v1_brands_brand_id_well_known_uris_path_customized]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/well-known-uris/{path}/customized disagrees with covered api_surface path /api/v1/brands/{brandId}/well-known-uris/{path}/customized.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/brands/{brand_id}/well-known-uris/{path}/customized disagrees with covered api_surface path /api/v1/brands/{brandId}/well-known-uris/{path}/customized.; flags: --brand-id (required), --path (required), --representation (required)
    update api v1 captchas captcha id apply - Typed action update_api_v1_captchas_captcha_id [intent=reverse_etl availability=partial write=update_api_v1_captchas_captcha_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/captchas/{captcha_id} disagrees with covered api_surface path /api/v1/captchas/{captchaId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/captchas/{captcha_id} disagrees with covered api_surface path /api/v1/captchas/{captchaId}.; flags: --captcha-id (required)
    update api v1 device assurances device assurance id apply - Typed action update_api_v1_device_assurances_device_assurance_id [intent=reverse_etl availability=partial write=update_api_v1_device_assurances_device_assurance_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/device-assurances/{device_assurance_id} disagrees with covered api_surface path /api/v1/device-assurances/{deviceAssuranceId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/device-assurances/{device_assurance_id} disagrees with covered api_surface path /api/v1/device-assurances/{deviceAssuranceId}.; flags: --device-assurance-id (required)
    update api v1 device posture checks posture check id apply - Typed action update_api_v1_device_posture_checks_posture_check_id [intent=reverse_etl availability=partial write=update_api_v1_device_posture_checks_posture_check_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/device-posture-checks/{posture_check_id} disagrees with covered api_surface path /api/v1/device-posture-checks/{postureCheckId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/device-posture-checks/{posture_check_id} disagrees with covered api_surface path /api/v1/device-posture-checks/{postureCheckId}.; flags: --posture-check-id (required)
    update api v1 domains domain id apply - Typed action update_api_v1_domains_domain_id [intent=reverse_etl availability=partial write=update_api_v1_domains_domain_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/domains/{domain_id} disagrees with covered api_surface path /api/v1/domains/{domainId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/domains/{domain_id} disagrees with covered api_surface path /api/v1/domains/{domainId}.; flags: --brand-id (required), --domain-id (required)
    update api v1 domains domain id certificate apply - Typed action update_api_v1_domains_domain_id_certificate [intent=reverse_etl availability=partial write=update_api_v1_domains_domain_id_certificate]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/domains/{domain_id}/certificate disagrees with covered api_surface path /api/v1/domains/{domainId}/certificate.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/domains/{domain_id}/certificate disagrees with covered api_surface path /api/v1/domains/{domainId}/certificate.; flags: --certificate (required), --certificate-chain (required), --domain-id (required), --private-key (required), --type (required)
    update api v1 email domains email domain id apply - Typed action update_api_v1_email_domains_email_domain_id [intent=reverse_etl availability=partial write=update_api_v1_email_domains_email_domain_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/email-domains/{email_domain_id} disagrees with covered api_surface path /api/v1/email-domains/{emailDomainId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/email-domains/{email_domain_id} disagrees with covered api_surface path /api/v1/email-domains/{emailDomainId}.; flags: --display-name (required), --email-domain-id (required), --user-name (required)
    update api v1 email servers email server id apply - Typed action update_api_v1_email_servers_email_server_id [intent=reverse_etl availability=partial write=update_api_v1_email_servers_email_server_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/email-servers/{email_server_id} disagrees with covered api_surface path /api/v1/email-servers/{emailServerId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/email-servers/{email_server_id} disagrees with covered api_surface path /api/v1/email-servers/{emailServerId}.; flags: --auth-type (required), --email-server-id (required)
    update api v1 event hooks event hook id apply - Typed action update_api_v1_event_hooks_event_hook_id [intent=reverse_etl availability=partial write=update_api_v1_event_hooks_event_hook_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/eventHooks/{event_hook_id} disagrees with covered api_surface path /api/v1/eventHooks/{eventHookId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/eventHooks/{event_hook_id} disagrees with covered api_surface path /api/v1/eventHooks/{eventHookId}.; flags: --channel (required), --event-hook-id (required), --events (required), --name (required)
    update api v1 first party app settings app name apply - Typed action update_api_v1_first_party_app_settings_app_name [intent=reverse_etl availability=partial write=update_api_v1_first_party_app_settings_app_name]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/first-party-app-settings/{app_name} disagrees with covered api_surface path /api/v1/first-party-app-settings/{appName}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/first-party-app-settings/{app_name} disagrees with covered api_surface path /api/v1/first-party-app-settings/{appName}.; flags: --app-name (required)
    update api v1 groups group id apply - Typed action update_api_v1_groups_group_id [intent=reverse_etl availability=partial write=update_api_v1_groups_group_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/groups/{group_id} disagrees with covered api_surface path /api/v1/groups/{groupId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/groups/{group_id} disagrees with covered api_surface path /api/v1/groups/{groupId}.; flags: --group-id (required)
    update api v1 groups group id roles role assignment id targets catalog apps app name app id apply - Typed action update_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id [intent=reverse_etl availability=partial write=update_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/groups/{group_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name}/{app_id} disagrees with covered api_surface path /api/v1/groups/{groupId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}/{appId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/groups/{group_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name}/{app_id} disagrees with covered api_surface path /api/v1/groups/{groupId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}/{appId}.; flags: --app-id (required), --app-name (required), --group-id (required), --role-assignment-id (required)
    update api v1 groups group id roles role assignment id targets catalog apps app name apply - Typed action update_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name [intent=reverse_etl availability=partial write=update_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/groups/{group_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name} disagrees with covered api_surface path /api/v1/groups/{groupId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/groups/{group_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name} disagrees with covered api_surface path /api/v1/groups/{groupId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}.; flags: --app-name (required), --group-id (required), --role-assignment-id (required)
    update api v1 groups group id roles role assignment id targets groups target group id apply - Typed action update_api_v1_groups_group_id_roles_role_assignment_id_targets_groups_target_group_id [intent=reverse_etl availability=partial write=update_api_v1_groups_group_id_roles_role_assignment_id_targets_groups_target_group_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/groups/{group_id}/roles/{role_assignment_id}/targets/groups/{target_group_id} disagrees with covered api_surface path /api/v1/groups/{groupId}/roles/{roleAssignmentId}/targets/groups/{targetGroupId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/groups/{group_id}/roles/{role_assignment_id}/targets/groups/{target_group_id} disagrees with covered api_surface path /api/v1/groups/{groupId}/roles/{roleAssignmentId}/targets/groups/{targetGroupId}.; flags: --group-id (required), --role-assignment-id (required), --target-group-id (required)
    update api v1 groups group id users user id apply - Typed action update_api_v1_groups_group_id_users_user_id [intent=reverse_etl availability=partial write=update_api_v1_groups_group_id_users_user_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/groups/{group_id}/users/{user_id} disagrees with covered api_surface path /api/v1/groups/{groupId}/users/{userId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/groups/{group_id}/users/{user_id} disagrees with covered api_surface path /api/v1/groups/{groupId}/users/{userId}.; flags: --group-id (required), --user-id (required)
    update api v1 groups rules group rule id apply - Typed action update_api_v1_groups_rules_group_rule_id [intent=reverse_etl availability=partial write=update_api_v1_groups_rules_group_rule_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/groups/rules/{group_rule_id} disagrees with covered api_surface path /api/v1/groups/rules/{groupRuleId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/groups/rules/{group_rule_id} disagrees with covered api_surface path /api/v1/groups/rules/{groupRuleId}.; flags: --group-rule-id (required)
    update api v1 hook keys id apply - PUT /api/v1/hook-keys/{id} (update_api_v1_hook_keys_id) [intent=reverse_etl availability=implemented write=update_api_v1_hook_keys_id]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    update api v1 iam governance bundles bundle id apply - Typed action update_api_v1_iam_governance_bundles_bundle_id [intent=reverse_etl availability=partial write=update_api_v1_iam_governance_bundles_bundle_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/iam/governance/bundles/{bundle_id} disagrees with covered api_surface path /api/v1/iam/governance/bundles/{bundleId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/iam/governance/bundles/{bundle_id} disagrees with covered api_surface path /api/v1/iam/governance/bundles/{bundleId}.; flags: --bundle-id (required)
    update api v1 iam resource sets resource set id or label apply - Typed action update_api_v1_iam_resource_sets_resource_set_id_or_label [intent=reverse_etl availability=partial write=update_api_v1_iam_resource_sets_resource_set_id_or_label]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label} disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label} disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}.; flags: --resource-set-id-or-label (required)
    update api v1 iam resource sets resource set id or label bindings role id or label members apply - Typed action update_api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members [intent=reverse_etl availability=partial write=update_api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label}/bindings/{role_id_or_label}/members disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}/bindings/{roleIdOrLabel}/members.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label}/bindings/{role_id_or_label}/members disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}/bindings/{roleIdOrLabel}/members.; flags: --resource-set-id-or-label (required), --role-id-or-label (required)
    update api v1 iam resource sets resource set id or label resources apply - Typed action update_api_v1_iam_resource_sets_resource_set_id_or_label_resources [intent=reverse_etl availability=partial write=update_api_v1_iam_resource_sets_resource_set_id_or_label_resources]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label}/resources disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}/resources.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label}/resources disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}/resources.; flags: --resource-set-id-or-label (required)
    update api v1 iam resource sets resource set id or label resources resource id apply - Typed action update_api_v1_iam_resource_sets_resource_set_id_or_label_resources_resource_id [intent=reverse_etl availability=partial write=update_api_v1_iam_resource_sets_resource_set_id_or_label_resources_resource_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label}/resources/{resource_id} disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}/resources/{resourceId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/iam/resource-sets/{resource_set_id_or_label}/resources/{resource_id} disagrees with covered api_surface path /api/v1/iam/resource-sets/{resourceSetIdOrLabel}/resources/{resourceId}.; flags: --resource-id (required), --resource-set-id-or-label (required)
    update api v1 iam roles role id or label apply - Typed action update_api_v1_iam_roles_role_id_or_label [intent=reverse_etl availability=partial write=update_api_v1_iam_roles_role_id_or_label]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/iam/roles/{role_id_or_label} disagrees with covered api_surface path /api/v1/iam/roles/{roleIdOrLabel}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/iam/roles/{role_id_or_label} disagrees with covered api_surface path /api/v1/iam/roles/{roleIdOrLabel}.; flags: --description (required), --label (required), --role-id-or-label (required)
    update api v1 iam roles role id or label permissions permission type apply - Typed action update_api_v1_iam_roles_role_id_or_label_permissions_permission_type [intent=reverse_etl availability=partial write=update_api_v1_iam_roles_role_id_or_label_permissions_permission_type]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/iam/roles/{role_id_or_label}/permissions/{permission_type} disagrees with covered api_surface path /api/v1/iam/roles/{roleIdOrLabel}/permissions/{permissionType}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/iam/roles/{role_id_or_label}/permissions/{permission_type} disagrees with covered api_surface path /api/v1/iam/roles/{roleIdOrLabel}/permissions/{permissionType}.; flags: --permission-type (required), --role-id-or-label (required)
    update api v1 identity sources identity source id users external id 2 apply - Typed action update_api_v1_identity_sources_identity_source_id_users_external_id_2 [intent=reverse_etl availability=partial write=update_api_v1_identity_sources_identity_source_id_users_external_id_2]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/users/{external_id} disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/users/{externalId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/users/{external_id} disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/users/{externalId}.; flags: --external-id (required), --identity-source-id (required)
    update api v1 identity sources identity source id users external id apply - Typed action update_api_v1_identity_sources_identity_source_id_users_external_id [intent=reverse_etl availability=partial write=update_api_v1_identity_sources_identity_source_id_users_external_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/users/{external_id} disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/users/{externalId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/identity-sources/{identity_source_id}/users/{external_id} disagrees with covered api_surface path /api/v1/identity-sources/{identitySourceId}/users/{externalId}.; flags: --external-id (required), --identity-source-id (required)
    update api v1 idps credentials keys kid apply - PUT /api/v1/idps/credentials/keys/{kid} (update_api_v1_idps_credentials_keys_kid) [intent=reverse_etl availability=implemented write=update_api_v1_idps_credentials_keys_kid]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --kid (required)
    update api v1 idps idp id apply - Typed action update_api_v1_idps_idp_id [intent=reverse_etl availability=partial write=update_api_v1_idps_idp_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/idps/{idp_id} disagrees with covered api_surface path /api/v1/idps/{idpId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/idps/{idp_id} disagrees with covered api_surface path /api/v1/idps/{idpId}.; flags: --idp-id (required)
    update api v1 inline hooks inline hook id apply - Typed action update_api_v1_inline_hooks_inline_hook_id [intent=reverse_etl availability=partial write=update_api_v1_inline_hooks_inline_hook_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/inlineHooks/{inline_hook_id} disagrees with covered api_surface path /api/v1/inlineHooks/{inlineHookId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/inlineHooks/{inline_hook_id} disagrees with covered api_surface path /api/v1/inlineHooks/{inlineHookId}.; flags: --inline-hook-id (required)
    update api v1 log streams log stream id apply - Typed action update_api_v1_log_streams_log_stream_id [intent=reverse_etl availability=partial write=update_api_v1_log_streams_log_stream_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/logStreams/{log_stream_id} disagrees with covered api_surface path /api/v1/logStreams/{logStreamId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/logStreams/{log_stream_id} disagrees with covered api_surface path /api/v1/logStreams/{logStreamId}.; flags: --log-stream-id (required), --name (required), --type (required)
    update api v1 meta types user type id apply - Typed action update_api_v1_meta_types_user_type_id [intent=reverse_etl availability=partial write=update_api_v1_meta_types_user_type_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/meta/types/user/{type_id} disagrees with covered api_surface path /api/v1/meta/types/user/{typeId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/meta/types/user/{type_id} disagrees with covered api_surface path /api/v1/meta/types/user/{typeId}.; flags: --description (required), --display-name (required), --name (required), --type-id (required)
    update api v1 meta uischemas id apply - PUT /api/v1/meta/uischemas/{id} (update_api_v1_meta_uischemas_id) [intent=reverse_etl availability=implemented write=update_api_v1_meta_uischemas_id]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    update api v1 org apply - PUT /api/v1/org (update_api_v1_org) [intent=reverse_etl availability=implemented write=update_api_v1_org]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    update api v1 org captcha apply - PUT /api/v1/org/captcha (update_api_v1_org_captcha) [intent=reverse_etl availability=implemented write=update_api_v1_org_captcha]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    update api v1 org contacts contact type apply - Typed action update_api_v1_org_contacts_contact_type [intent=reverse_etl availability=partial write=update_api_v1_org_contacts_contact_type]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/org/contacts/{contact_type} disagrees with covered api_surface path /api/v1/org/contacts/{contactType}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/org/contacts/{contact_type} disagrees with covered api_surface path /api/v1/org/contacts/{contactType}.; flags: --contact-type (required)
    update api v1 org privacy okta support cases case number apply - Typed action update_api_v1_org_privacy_okta_support_cases_case_number [intent=reverse_etl availability=partial write=update_api_v1_org_privacy_okta_support_cases_case_number]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/org/privacy/oktaSupport/cases/{case_number} disagrees with covered api_surface path /api/v1/org/privacy/oktaSupport/cases/{caseNumber}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/org/privacy/oktaSupport/cases/{case_number} disagrees with covered api_surface path /api/v1/org/privacy/oktaSupport/cases/{caseNumber}.; flags: --case-number (required)
    update api v1 org settings client privileges setting apply - PUT /api/v1/org/settings/clientPrivilegesSetting (update_api_v1_org_settings_client_privileges_setting) [intent=reverse_etl availability=implemented write=update_api_v1_org_settings_client_privileges_setting]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    update api v1 policies policy id apply - Typed action update_api_v1_policies_policy_id [intent=reverse_etl availability=partial write=update_api_v1_policies_policy_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/policies/{policy_id} disagrees with covered api_surface path /api/v1/policies/{policyId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/policies/{policy_id} disagrees with covered api_surface path /api/v1/policies/{policyId}.; flags: --name (required), --policy-id (required), --type (required)
    update api v1 policies policy id rules rule id apply - Typed action update_api_v1_policies_policy_id_rules_rule_id [intent=reverse_etl availability=partial write=update_api_v1_policies_policy_id_rules_rule_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/rules/{rule_id} disagrees with covered api_surface path /api/v1/policies/{policyId}/rules/{ruleId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/policies/{policy_id}/rules/{rule_id} disagrees with covered api_surface path /api/v1/policies/{policyId}/rules/{ruleId}.; flags: --policy-id (required), --rule-id (required)
    update api v1 principal rate limits principal rate limit id apply - Typed action update_api_v1_principal_rate_limits_principal_rate_limit_id [intent=reverse_etl availability=partial write=update_api_v1_principal_rate_limits_principal_rate_limit_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/principal-rate-limits/{principal_rate_limit_id} disagrees with covered api_surface path /api/v1/principal-rate-limits/{principalRateLimitId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/principal-rate-limits/{principal_rate_limit_id} disagrees with covered api_surface path /api/v1/principal-rate-limits/{principalRateLimitId}.; flags: --principal-id (required), --principal-type (required), --principal-rate-limit-id (required)
    update api v1 push providers push provider id apply - Typed action update_api_v1_push_providers_push_provider_id [intent=reverse_etl availability=partial write=update_api_v1_push_providers_push_provider_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/push-providers/{push_provider_id} disagrees with covered api_surface path /api/v1/push-providers/{pushProviderId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/push-providers/{push_provider_id} disagrees with covered api_surface path /api/v1/push-providers/{pushProviderId}.; flags: --push-provider-id (required)
    update api v1 rate limit settings admin notifications apply - PUT /api/v1/rate-limit-settings/admin-notifications (update_api_v1_rate_limit_settings_admin_notifications) [intent=reverse_etl availability=implemented write=update_api_v1_rate_limit_settings_admin_notifications]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --notifications-enabled (required)
    update api v1 rate limit settings per client apply - PUT /api/v1/rate-limit-settings/per-client (update_api_v1_rate_limit_settings_per_client) [intent=reverse_etl availability=implemented write=update_api_v1_rate_limit_settings_per_client]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --default-mode (required)
    update api v1 rate limit settings warning threshold apply - PUT /api/v1/rate-limit-settings/warning-threshold (update_api_v1_rate_limit_settings_warning_threshold) [intent=reverse_etl availability=implemented write=update_api_v1_rate_limit_settings_warning_threshold]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --warning-threshold (required)
    update api v1 realm assignments assignment id apply - Typed action update_api_v1_realm_assignments_assignment_id [intent=reverse_etl availability=partial write=update_api_v1_realm_assignments_assignment_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/realm-assignments/{assignment_id} disagrees with covered api_surface path /api/v1/realm-assignments/{assignmentId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/realm-assignments/{assignment_id} disagrees with covered api_surface path /api/v1/realm-assignments/{assignmentId}.; flags: --assignment-id (required)
    update api v1 realms realm id apply - Typed action update_api_v1_realms_realm_id [intent=reverse_etl availability=partial write=update_api_v1_realms_realm_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/realms/{realm_id} disagrees with covered api_surface path /api/v1/realms/{realmId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/realms/{realm_id} disagrees with covered api_surface path /api/v1/realms/{realmId}.; flags: --realm-id (required)
    update api v1 security events providers security event provider id apply - Typed action update_api_v1_security_events_providers_security_event_provider_id [intent=reverse_etl availability=partial write=update_api_v1_security_events_providers_security_event_provider_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/security-events-providers/{security_event_provider_id} disagrees with covered api_surface path /api/v1/security-events-providers/{securityEventProviderId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/security-events-providers/{security_event_provider_id} disagrees with covered api_surface path /api/v1/security-events-providers/{securityEventProviderId}.; flags: --name (required), --security-event-provider-id (required), --settings (required), --type (required)
    update api v1 ssf stream 2 apply - PATCH /api/v1/ssf/stream (update_api_v1_ssf_stream_2) [intent=reverse_etl availability=implemented write=update_api_v1_ssf_stream_2]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --delivery (required), --events-requested (required)
    update api v1 ssf stream apply - PUT /api/v1/ssf/stream (update_api_v1_ssf_stream) [intent=reverse_etl availability=implemented write=update_api_v1_ssf_stream]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --delivery (required), --events-requested (required)
    update api v1 telephony providers custom telephony provider id apply - Typed action update_api_v1_telephony_providers_custom_telephony_provider_id [intent=reverse_etl availability=partial write=update_api_v1_telephony_providers_custom_telephony_provider_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/telephony-providers/{custom_telephony_provider_id} disagrees with covered api_surface path /api/v1/telephony-providers/{customTelephonyProviderId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/telephony-providers/{custom_telephony_provider_id} disagrees with covered api_surface path /api/v1/telephony-providers/{customTelephonyProviderId}.; flags: --custom-telephony-provider-id (required)
    update api v1 templates sms template id apply - Typed action update_api_v1_templates_sms_template_id [intent=reverse_etl availability=partial write=update_api_v1_templates_sms_template_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/templates/sms/{template_id} disagrees with covered api_surface path /api/v1/templates/sms/{templateId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/templates/sms/{template_id} disagrees with covered api_surface path /api/v1/templates/sms/{templateId}.; flags: --template-id (required)
    update api v1 trusted origins trusted origin id apply - Typed action update_api_v1_trusted_origins_trusted_origin_id [intent=reverse_etl availability=partial write=update_api_v1_trusted_origins_trusted_origin_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/trustedOrigins/{trusted_origin_id} disagrees with covered api_surface path /api/v1/trustedOrigins/{trustedOriginId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/trustedOrigins/{trusted_origin_id} disagrees with covered api_surface path /api/v1/trustedOrigins/{trustedOriginId}.; flags: --trusted-origin-id (required)
    update api v1 users id apply - PUT /api/v1/users/{id} (update_api_v1_users_id) [intent=reverse_etl availability=implemented write=update_api_v1_users_id]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    update api v1 users user id classification apply - Typed action update_api_v1_users_user_id_classification [intent=reverse_etl availability=partial write=update_api_v1_users_user_id_classification]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/classification disagrees with covered api_surface path /api/v1/users/{userId}/classification.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/classification disagrees with covered api_surface path /api/v1/users/{userId}/classification.; flags: --user-id (required)
    update api v1 users user id or login linked objects primary relationship name primary user id apply - Typed action update_api_v1_users_user_id_or_login_linked_objects_primary_relationship_name_primary_user_id [intent=reverse_etl availability=partial write=update_api_v1_users_user_id_or_login_linked_objects_primary_relationship_name_primary_user_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id_or_login}/linkedObjects/{primary_relationship_name}/{primary_user_id} disagrees with covered api_surface path /api/v1/users/{userIdOrLogin}/linkedObjects/{primaryRelationshipName}/{primaryUserId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id_or_login}/linkedObjects/{primary_relationship_name}/{primary_user_id} disagrees with covered api_surface path /api/v1/users/{userIdOrLogin}/linkedObjects/{primaryRelationshipName}/{primaryUserId}.; flags: --primary-relationship-name (required), --primary-user-id (required), --user-id-or-login (required)
    update api v1 users user id risk apply - Typed action update_api_v1_users_user_id_risk [intent=reverse_etl availability=partial write=update_api_v1_users_user_id_risk]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/risk disagrees with covered api_surface path /api/v1/users/{userId}/risk.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/risk disagrees with covered api_surface path /api/v1/users/{userId}/risk.; flags: --risk-level (required), --user-id (required)
    update api v1 users user id roles role assignment id targets catalog apps app name app id apply - Typed action update_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id [intent=reverse_etl availability=partial write=update_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name}/{app_id} disagrees with covered api_surface path /api/v1/users/{userId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}/{appId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name}/{app_id} disagrees with covered api_surface path /api/v1/users/{userId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}/{appId}.; flags: --app-id (required), --app-name (required), --role-assignment-id (required), --user-id (required)
    update api v1 users user id roles role assignment id targets catalog apps app name apply - Typed action update_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name [intent=reverse_etl availability=partial write=update_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name} disagrees with covered api_surface path /api/v1/users/{userId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name} disagrees with covered api_surface path /api/v1/users/{userId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}.; flags: --app-name (required), --role-assignment-id (required), --user-id (required)
    update api v1 users user id roles role assignment id targets catalog apps apply - Typed action update_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps [intent=reverse_etl availability=partial write=update_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/roles/{role_assignment_id}/targets/catalog/apps disagrees with covered api_surface path /api/v1/users/{userId}/roles/{roleAssignmentId}/targets/catalog/apps.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/roles/{role_assignment_id}/targets/catalog/apps disagrees with covered api_surface path /api/v1/users/{userId}/roles/{roleAssignmentId}/targets/catalog/apps.; flags: --role-assignment-id (required), --user-id (required)
    update api v1 users user id roles role assignment id targets groups group id apply - Typed action update_api_v1_users_user_id_roles_role_assignment_id_targets_groups_group_id [intent=reverse_etl availability=partial write=update_api_v1_users_user_id_roles_role_assignment_id_targets_groups_group_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/users/{user_id}/roles/{role_assignment_id}/targets/groups/{group_id} disagrees with covered api_surface path /api/v1/users/{userId}/roles/{roleAssignmentId}/targets/groups/{groupId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/users/{user_id}/roles/{role_assignment_id}/targets/groups/{group_id} disagrees with covered api_surface path /api/v1/users/{userId}/roles/{roleAssignmentId}/targets/groups/{groupId}.; flags: --group-id (required), --role-assignment-id (required), --user-id (required)
    update api v1 zones zone id apply - Typed action update_api_v1_zones_zone_id [intent=reverse_etl availability=partial write=update_api_v1_zones_zone_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /api/v1/zones/{zone_id} disagrees with covered api_surface path /api/v1/zones/{zoneId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /api/v1/zones/{zone_id} disagrees with covered api_surface path /api/v1/zones/{zoneId}.; flags: --name (required), --type (required), --zone-id (required)
    update attack protection api v1 authenticator settings apply - PUT /attack-protection/api/v1/authenticator-settings (update_attack_protection_api_v1_authenticator_settings) [intent=reverse_etl availability=implemented write=update_attack_protection_api_v1_authenticator_settings]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    update attack protection api v1 user lockout settings apply - PUT /attack-protection/api/v1/user-lockout-settings (update_attack_protection_api_v1_user_lockout_settings) [intent=reverse_etl availability=implemented write=update_attack_protection_api_v1_user_lockout_settings]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    update oauth2 v1 clients client id roles role assignment id targets catalog apps app name app id apply - Typed action update_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id [intent=reverse_etl availability=partial write=update_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /oauth2/v1/clients/{client_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name}/{app_id} disagrees with covered api_surface path /oauth2/v1/clients/{clientId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}/{appId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /oauth2/v1/clients/{client_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name}/{app_id} disagrees with covered api_surface path /oauth2/v1/clients/{clientId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}/{appId}.; flags: --app-id (required), --app-name (required), --client-id (required), --role-assignment-id (required)
    update oauth2 v1 clients client id roles role assignment id targets catalog apps app name apply - Typed action update_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name [intent=reverse_etl availability=partial write=update_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /oauth2/v1/clients/{client_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name} disagrees with covered api_surface path /oauth2/v1/clients/{clientId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /oauth2/v1/clients/{client_id}/roles/{role_assignment_id}/targets/catalog/apps/{app_name} disagrees with covered api_surface path /oauth2/v1/clients/{clientId}/roles/{roleAssignmentId}/targets/catalog/apps/{appName}.; flags: --app-name (required), --client-id (required), --role-assignment-id (required)
    update oauth2 v1 clients client id roles role assignment id targets groups group id apply - Typed action update_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_groups_group_id [intent=reverse_etl availability=partial write=update_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_groups_group_id]; approval: Blocked pending a faithful CLI record binding: declaration-pending: canonical typed action path /oauth2/v1/clients/{client_id}/roles/{role_assignment_id}/targets/groups/{group_id} disagrees with covered api_surface path /oauth2/v1/clients/{clientId}/roles/{roleAssignmentId}/targets/groups/{groupId}.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; declaration-pending: canonical typed action path /oauth2/v1/clients/{client_id}/roles/{role_assignment_id}/targets/groups/{group_id} disagrees with covered api_surface path /oauth2/v1/clients/{clientId}/roles/{roleAssignmentId}/targets/groups/{groupId}.; flags: --client-id (required), --group-id (required), --role-assignment-id (required)
    update okta personal settings api v1 edit feature apply - PUT /okta-personal-settings/api/v1/edit-feature (update_okta_personal_settings_api_v1_edit_feature) [intent=reverse_etl availability=implemented write=update_okta_personal_settings_api_v1_edit_feature]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    update okta personal settings api v1 export blocklists apply - PUT /okta-personal-settings/api/v1/export-blocklists (update_okta_personal_settings_api_v1_export_blocklists) [intent=reverse_etl availability=implemented write=update_okta_personal_settings_api_v1_export_blocklists]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.
    update privileged access api v1 okta service accounts id apply - PATCH /privileged-access/api/v1/okta-service-accounts/{id} (update_privileged_access_api_v1_okta_service_accounts_id) [intent=reverse_etl availability=implemented write=update_privileged_access_api_v1_okta_service_accounts_id]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)
    update privileged access api v1 service accounts id apply - PATCH /privileged-access/api/v1/service-accounts/{id} (update_privileged_access_api_v1_service_accounts_id) [intent=reverse_etl availability=implemented write=update_privileged_access_api_v1_service_accounts_id]; approval: reverse ETL writes require plan, preview, approval, and execute.; risk: medium: external Okta admin API mutation; approval required; notes: Generated from the connector-owned typed action; execution remains plan-gated.; flags: --id (required)

SYNC TRANSPORT
  Source transport: declared
  Destination transport: unsupported
  A declared transport still requires runtime preflight and externally verified conformance; it is not a certification claim.
  Source executor: declarative_api/declarative_stream_source

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
