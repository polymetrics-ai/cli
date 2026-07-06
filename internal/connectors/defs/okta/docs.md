# Overview

Reads and writes documented Okta Admin Management API resources through the Okta REST APIs.

Readable streams: `users`, `groups`, `system_logs`, `well_known_app_authenticator_configuration`,
`well_known_apple_app_site_association`, `well_known_assetlinks_json`,
`well_known_okta_organization`, `well_known_ssf_configuration`, `well_known_webauthn`,
`api_v1_agent_pools`, `api_v1_agent_pools_pool_id_updates`,
`api_v1_agent_pools_pool_id_updates_settings`, `api_v1_agent_pools_pool_id_updates_update_id`,
`api_v1_api_tokens`, `api_v1_api_tokens_api_token_id`, `api_v1_apps`, `api_v1_apps_app_id`,
`api_v1_apps_app_id_connections_default`, `api_v1_apps_app_id_connections_default_jwks`,
`api_v1_apps_app_id_credentials_csrs`, `api_v1_apps_app_id_credentials_csrs_csr_id`,
`api_v1_apps_app_id_credentials_jwks`, `api_v1_apps_app_id_credentials_jwks_key_id`,
`api_v1_apps_app_id_credentials_keys`, `api_v1_apps_app_id_credentials_keys_key_id`,
`api_v1_apps_app_id_credentials_secrets`, `api_v1_apps_app_id_credentials_secrets_secret_id`,
`api_v1_apps_app_id_cwo_connections`, `api_v1_apps_app_id_cwo_connections_connection_id`,
`api_v1_apps_app_id_features`, `api_v1_apps_app_id_features_feature_name`,
`api_v1_apps_app_id_federated_claims`, `api_v1_apps_app_id_federated_claims_claim_id`,
`api_v1_apps_app_id_grants`, `api_v1_apps_app_id_grants_grant_id`,
`api_v1_apps_app_id_group_push_mappings`, `api_v1_apps_app_id_group_push_mappings_mapping_id`,
`api_v1_apps_app_id_groups`, `api_v1_apps_app_id_groups_group_id`, `api_v1_apps_app_id_tokens`,
`api_v1_apps_app_id_tokens_token_id`, `api_v1_apps_app_id_users`,
`api_v1_apps_app_id_users_user_id`, `api_v1_authenticators`,
`api_v1_authenticators_authenticator_id`, `api_v1_authenticators_authenticator_id_aaguids`,
`api_v1_authenticators_authenticator_id_aaguids_aaguid`,
`api_v1_authenticators_authenticator_id_methods`,
`api_v1_authenticators_authenticator_id_methods_method_type`, `api_v1_authorization_servers`,
`api_v1_authorization_servers_auth_server_id`,
`api_v1_authorization_servers_auth_server_id_associated_servers`,
`api_v1_authorization_servers_auth_server_id_claims`,
`api_v1_authorization_servers_auth_server_id_claims_claim_id`,
`api_v1_authorization_servers_auth_server_id_clients`,
`api_v1_authorization_servers_auth_server_id_clients_client_id_tokens`,
`api_v1_authorization_servers_auth_server_id_clients_client_id_tokens_token_id`,
`api_v1_authorization_servers_auth_server_id_credentials_keys`,
`api_v1_authorization_servers_auth_server_id_credentials_keys_key_id`,
`api_v1_authorization_servers_auth_server_id_policies`,
`api_v1_authorization_servers_auth_server_id_policies_policy_id`,
`api_v1_authorization_servers_auth_server_id_policies_policy_id_rules`,
`api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id`,
`api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys`,
`api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id`,
`api_v1_authorization_servers_auth_server_id_scopes`,
`api_v1_authorization_servers_auth_server_id_scopes_scope_id`, `api_v1_behaviors`,
`api_v1_behaviors_behavior_id`, `api_v1_bot_protection_configuration`, `api_v1_brands`,
`api_v1_brands_brand_id`, `api_v1_brands_brand_id_domains`, `api_v1_brands_brand_id_pages_error`,
`api_v1_brands_brand_id_pages_error_customized`, `api_v1_brands_brand_id_pages_error_default`,
`api_v1_brands_brand_id_pages_error_preview`, `api_v1_brands_brand_id_pages_sign_in`,
`api_v1_brands_brand_id_pages_sign_in_customized`, `api_v1_brands_brand_id_pages_sign_in_default`,
`api_v1_brands_brand_id_pages_sign_in_preview`, `api_v1_brands_brand_id_pages_sign_out_customized`,
`api_v1_brands_brand_id_templates_email`, `api_v1_brands_brand_id_templates_email_template_name`,
`api_v1_brands_brand_id_templates_email_template_name_customizations`,
`api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id`,
`api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id_preview`,
`api_v1_brands_brand_id_templates_email_template_name_default_content`,
`api_v1_brands_brand_id_templates_email_template_name_default_content_preview`,
`api_v1_brands_brand_id_templates_email_template_name_settings`, `api_v1_brands_brand_id_themes`,
`api_v1_brands_brand_id_themes_theme_id`, `api_v1_brands_brand_id_well_known_uris`,
`api_v1_brands_brand_id_well_known_uris_path`,
`api_v1_brands_brand_id_well_known_uris_path_customized`, `api_v1_captchas`,
`api_v1_captchas_captcha_id`, `api_v1_device_assurances`,
`api_v1_device_assurances_device_assurance_id`, `api_v1_device_integrations`,
`api_v1_device_integrations_device_integration_id`, `api_v1_device_posture_checks`,
`api_v1_device_posture_checks_default`, `api_v1_device_posture_checks_posture_check_id`,
`api_v1_devices`, `api_v1_devices_device_id`, `api_v1_devices_device_id_users`,
`api_v1_directories_app_instance_id_groups_group_id_query_result_id`, `api_v1_domains`,
`api_v1_domains_domain_id`, `api_v1_dr_status`, `api_v1_dr_status_domain`, `api_v1_email_domains`,
`api_v1_email_domains_email_domain_id`, `api_v1_email_servers`,
`api_v1_email_servers_email_server_id`, `api_v1_event_hooks`, `api_v1_event_hooks_event_hook_id`,
`api_v1_features`, `api_v1_features_feature_id`, `api_v1_features_feature_id_dependencies`,
`api_v1_features_feature_id_dependents`, `api_v1_first_party_app_settings_app_name`,
`api_v1_groups_rules`, `api_v1_groups_rules_group_rule_id`, `api_v1_groups_group_id`,
`api_v1_groups_group_id_apps`, `api_v1_groups_group_id_owners`, `api_v1_groups_group_id_roles`,
`api_v1_groups_group_id_roles_role_assignment_id`,
`api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps`,
`api_v1_groups_group_id_roles_role_assignment_id_targets_groups`, `api_v1_groups_group_id_users`,
`api_v1_hook_keys`, `api_v1_hook_keys_public_key_id`, `api_v1_hook_keys_id`,
`api_v1_iam_assignees_users`, `api_v1_iam_governance_bundles`,
`api_v1_iam_governance_bundles_bundle_id`, `api_v1_iam_governance_bundles_bundle_id_entitlements`,
`api_v1_iam_governance_bundles_bundle_id_entitlements_entitlement_id_values`,
`api_v1_iam_governance_opt_in`, `api_v1_iam_resource_sets`,
`api_v1_iam_resource_sets_resource_set_id_or_label`,
`api_v1_iam_resource_sets_resource_set_id_or_label_bindings`,
`api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label`,
`api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members`,
`api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members_member_id`,
`api_v1_iam_resource_sets_resource_set_id_or_label_resources`,
`api_v1_iam_resource_sets_resource_set_id_or_label_resources_resource_id`, `api_v1_iam_roles`,
`api_v1_iam_roles_role_id_or_label`, `api_v1_iam_roles_role_id_or_label_permissions`,
`api_v1_iam_roles_role_id_or_label_permissions_permission_type`,
`api_v1_identity_sources_identity_source_id_groups_group_or_external_id`,
`api_v1_identity_sources_identity_source_id_groups_group_or_external_id_membership`,
`api_v1_identity_sources_identity_source_id_sessions`,
`api_v1_identity_sources_identity_source_id_sessions_session_id`,
`api_v1_identity_sources_identity_source_id_users_external_id`, `api_v1_idps`,
`api_v1_idps_credentials_keys`, `api_v1_idps_credentials_keys_kid`, `api_v1_idps_idp_id`,
`api_v1_idps_idp_id_credentials_csrs`, `api_v1_idps_idp_id_credentials_csrs_idp_csr_id`,
`api_v1_idps_idp_id_credentials_keys`, `api_v1_idps_idp_id_credentials_keys_active`,
`api_v1_idps_idp_id_credentials_keys_kid`, `api_v1_idps_idp_id_users`,
`api_v1_idps_idp_id_users_user_id`, `api_v1_idps_idp_id_users_user_id_credentials_tokens`,
`api_v1_inline_hooks`, `api_v1_inline_hooks_inline_hook_id`, `api_v1_log_streams`,
`api_v1_log_streams_log_stream_id`, `api_v1_mappings`, `api_v1_mappings_mapping_id`,
`api_v1_meta_schemas_apps_app_id_default`, `api_v1_meta_schemas_group_default`,
`api_v1_meta_schemas_log_stream`, `api_v1_meta_schemas_log_stream_log_stream_type`,
`api_v1_meta_schemas_user_linked_objects`,
`api_v1_meta_schemas_user_linked_objects_linked_object_name`, `api_v1_meta_schemas_user_schema_id`,
`api_v1_meta_types_user`, `api_v1_meta_types_user_type_id`, `api_v1_meta_uischemas`,
`api_v1_meta_uischemas_id`, `api_v1_org`, `api_v1_org_captcha`, `api_v1_org_contacts`,
`api_v1_org_contacts_contact_type`, `api_v1_org_factors_yubikey_token_tokens`,
`api_v1_org_factors_yubikey_token_tokens_token_id`,
`api_v1_org_org_settings_third_party_admin_setting`, `api_v1_org_preferences`,
`api_v1_org_privacy_aerial`, `api_v1_org_privacy_okta_communication`,
`api_v1_org_privacy_okta_support`, `api_v1_org_privacy_okta_support_cases`,
`api_v1_org_settings_auto_assign_admin_app_setting`,
`api_v1_org_settings_client_privileges_setting`, `api_v1_policies`, `api_v1_policies_policy_id`,
`api_v1_policies_policy_id_app`, `api_v1_policies_policy_id_mappings`,
`api_v1_policies_policy_id_mappings_mapping_id`, `api_v1_policies_policy_id_rules`,
`api_v1_policies_policy_id_rules_rule_id`, `api_v1_principal_rate_limits`,
`api_v1_principal_rate_limits_principal_rate_limit_id`, `api_v1_push_providers`,
`api_v1_push_providers_push_provider_id`, `api_v1_rate_limit_settings_admin_notifications`,
`api_v1_rate_limit_settings_per_client`, `api_v1_rate_limit_settings_warning_threshold`,
`api_v1_realm_assignments`, `api_v1_realm_assignments_operations`,
`api_v1_realm_assignments_assignment_id`, `api_v1_realms`, `api_v1_realms_realm_id`,
`api_v1_roles_role_ref_subscriptions`, `api_v1_roles_role_ref_subscriptions_notification_type`,
`api_v1_security_events_providers`, `api_v1_security_events_providers_security_event_provider_id`,
`api_v1_sessions_session_id`, `api_v1_ssf_stream`, `api_v1_ssf_stream_status`,
`api_v1_telephony_providers`, `api_v1_telephony_providers_custom_telephony_provider_id`,
`api_v1_templates_sms`, `api_v1_templates_sms_template_id`, `api_v1_threats_configuration`,
`api_v1_trusted_origins`, `api_v1_trusted_origins_trusted_origin_id`, `api_v1_users_id`,
`api_v1_users_id_app_links`, `api_v1_users_id_blocks`, `api_v1_users_id_groups`,
`api_v1_users_id_idps`, `api_v1_users_user_id_or_login_linked_objects_relationship_name`,
`api_v1_users_user_id_authenticator_enrollments`,
`api_v1_users_user_id_authenticator_enrollments_enrollment_id`,
`api_v1_users_user_id_classification`, `api_v1_users_user_id_clients`,
`api_v1_users_user_id_clients_client_id_grants`, `api_v1_users_user_id_clients_client_id_tokens`,
`api_v1_users_user_id_clients_client_id_tokens_token_id`, `api_v1_users_user_id_devices`,
`api_v1_users_user_id_factors`, `api_v1_users_user_id_factors_catalog`,
`api_v1_users_user_id_factors_questions`, `api_v1_users_user_id_factors_factor_id`,
`api_v1_users_user_id_factors_factor_id_transactions_transaction_id`, `api_v1_users_user_id_grants`,
`api_v1_users_user_id_grants_grant_id`, `api_v1_users_user_id_risk`, `api_v1_users_user_id_roles`,
`api_v1_users_user_id_roles_role_assignment_id`,
`api_v1_users_user_id_roles_role_assignment_id_governance`,
`api_v1_users_user_id_roles_role_assignment_id_governance_grant_id`,
`api_v1_users_user_id_roles_role_assignment_id_governance_grant_id_resources`,
`api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps`,
`api_v1_users_user_id_roles_role_assignment_id_targets_groups`,
`api_v1_users_user_id_roles_role_id_or_encoded_role_id_targets`,
`api_v1_users_user_id_subscriptions`, `api_v1_users_user_id_subscriptions_notification_type`,
`api_v1_zones`, `api_v1_zones_zone_id`, `attack_protection_api_v1_authenticator_settings`,
`attack_protection_api_v1_user_lockout_settings`, `integrations_api_v1_api_services`,
`integrations_api_v1_api_services_api_service_id`,
`integrations_api_v1_api_services_api_service_id_credentials_secrets`,
`oauth2_v1_clients_client_id_roles`, `oauth2_v1_clients_client_id_roles_role_assignment_id`,
`oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps`,
`oauth2_v1_clients_client_id_roles_role_assignment_id_targets_groups`,
`okta_personal_settings_api_v1_export_blocklists`, `privileged_access_api_v1_okta_service_accounts`,
`privileged_access_api_v1_okta_service_accounts_id`, `privileged_access_api_v1_service_accounts`,
`privileged_access_api_v1_service_accounts_id`,
`webauthn_registration_api_v1_users_user_id_enrollments`.

Write actions: `create_api_v1_agent_pools_pool_id_updates`,
`create_api_v1_agent_pools_pool_id_updates_settings`,
`create_api_v1_agent_pools_pool_id_updates_update_id`,
`delete_api_v1_agent_pools_pool_id_updates_update_id`,
`execute_api_v1_agent_pools_pool_id_updates_update_id_activate`,
`execute_api_v1_agent_pools_pool_id_updates_update_id_deactivate`,
`execute_api_v1_agent_pools_pool_id_updates_update_id_pause`,
`execute_api_v1_agent_pools_pool_id_updates_update_id_resume`,
`execute_api_v1_agent_pools_pool_id_updates_update_id_retry`,
`execute_api_v1_agent_pools_pool_id_updates_update_id_stop`, `delete_api_v1_api_tokens_current`,
`update_api_v1_api_tokens_api_token_id`, `delete_api_v1_api_tokens_api_token_id`,
`create_api_v1_apps`, `update_api_v1_apps_app_id`, `delete_api_v1_apps_app_id`,
`create_api_v1_apps_app_id_connections_default`,
`execute_api_v1_apps_app_id_connections_default_lifecycle_activate`,
`execute_api_v1_apps_app_id_connections_default_lifecycle_deactivate`,
`create_api_v1_apps_app_id_credentials_csrs`, `delete_api_v1_apps_app_id_credentials_csrs_csr_id`,
`create_api_v1_apps_app_id_credentials_jwks`, `delete_api_v1_apps_app_id_credentials_jwks_key_id`,
`execute_api_v1_apps_app_id_credentials_jwks_key_id_lifecycle_activate`,
`execute_api_v1_apps_app_id_credentials_jwks_key_id_lifecycle_deactivate`,
`create_api_v1_apps_app_id_credentials_secrets`,
`delete_api_v1_apps_app_id_credentials_secrets_secret_id`,
`execute_api_v1_apps_app_id_credentials_secrets_secret_id_lifecycle_activate`,
`execute_api_v1_apps_app_id_credentials_secrets_secret_id_lifecycle_deactivate`,
`create_api_v1_apps_app_id_cwo_connections`,
`update_api_v1_apps_app_id_cwo_connections_connection_id`,
`delete_api_v1_apps_app_id_cwo_connections_connection_id`,
`update_api_v1_apps_app_id_features_feature_name`, `create_api_v1_apps_app_id_federated_claims`,
`update_api_v1_apps_app_id_federated_claims_claim_id`,
`delete_api_v1_apps_app_id_federated_claims_claim_id`, `create_api_v1_apps_app_id_grants`,
`delete_api_v1_apps_app_id_grants_grant_id`, `create_api_v1_apps_app_id_group_push_mappings`,
`update_api_v1_apps_app_id_group_push_mappings_mapping_id`,
`update_api_v1_apps_app_id_groups_group_id`, `delete_api_v1_apps_app_id_groups_group_id`,
`create_api_v1_apps_app_id_interclient_allowed_apps`,
`delete_api_v1_apps_app_id_interclient_allowed_apps_allowed_app_id`,
`execute_api_v1_apps_app_id_lifecycle_activate`, `execute_api_v1_apps_app_id_lifecycle_deactivate`,
`update_api_v1_apps_app_id_policies_policy_id`, `delete_api_v1_apps_app_id_tokens`,
`delete_api_v1_apps_app_id_tokens_token_id`, `create_api_v1_apps_app_id_users`,
`create_api_v1_apps_app_id_users_user_id`, `delete_api_v1_apps_app_id_users_user_id`,
`create_api_v1_apps_app_name_app_id_oauth2_callback`, `create_api_v1_authenticators`,
`update_api_v1_authenticators_authenticator_id`,
`create_api_v1_authenticators_authenticator_id_aaguids`,
`update_api_v1_authenticators_authenticator_id_aaguids_aaguid`,
`update_api_v1_authenticators_authenticator_id_aaguids_aaguid_2`,
`delete_api_v1_authenticators_authenticator_id_aaguids_aaguid`,
`execute_api_v1_authenticators_authenticator_id_lifecycle_activate`,
`execute_api_v1_authenticators_authenticator_id_lifecycle_deactivate`,
`update_api_v1_authenticators_authenticator_id_methods_method_type`,
`execute_api_v1_authenticators_authenticator_id_methods_method_type_lifecycle_activate`,
`execute_api_v1_authenticators_authenticator_id_methods_method_type_lifecycle_deactivate`,
`create_api_v1_authenticators_authenticator_id_methods_web_authn_method_type_verify_rp_id_domain`,
`create_api_v1_authorization_servers`, `update_api_v1_authorization_servers_auth_server_id`,
`delete_api_v1_authorization_servers_auth_server_id`,
`create_api_v1_authorization_servers_auth_server_id_associated_servers`,
`delete_api_v1_authorization_servers_auth_server_id_associated_servers_associated_server_id`,
`create_api_v1_authorization_servers_auth_server_id_claims`,
`update_api_v1_authorization_servers_auth_server_id_claims_claim_id`,
`delete_api_v1_authorization_servers_auth_server_id_claims_claim_id`,
`delete_api_v1_authorization_servers_auth_server_id_clients_client_id_tokens`,
`delete_api_v1_authorization_servers_auth_server_id_clients_client_id_tokens_token_id`,
`execute_api_v1_authorization_servers_auth_server_id_credentials_lifecycle_key_rotate`,
`execute_api_v1_authorization_servers_auth_server_id_lifecycle_activate`,
`execute_api_v1_authorization_servers_auth_server_id_lifecycle_deactivate`,
`create_api_v1_authorization_servers_auth_server_id_policies`,
`update_api_v1_authorization_servers_auth_server_id_policies_policy_id`, and 349 more.

Service API documentation: https://developer.okta.com/docs/api/openapi/okta-management/management/.

## Auth setup

Connection fields:

- `aaguid` (optional, string); Path parameter aaguid for
  api_v1_authenticators_authenticator_id_aaguids_aaguid.
- `access_token` (optional, secret, string); OAuth access token, sent as Bearer when api_token is
  absent.
- `api_service_id` (optional, string); Path parameter apiServiceId for
  integrations_api_v1_api_services_api_service_id.
- `api_token` (optional, secret, string); Okta API token, sent as Authorization: SSWS.
- `api_token_id` (optional, string); Path parameter apiTokenId for api_v1_api_tokens_api_token_id.
- `app_id` (optional, string); Path parameter appId for api_v1_apps_app_id.
- `app_instance_id` (optional, string); Path parameter appInstanceId for
  api_v1_directories_app_instance_id_groups_group_id_query_result_id.
- `app_name` (optional, string); Path parameter appName for
  api_v1_first_party_app_settings_app_name.
- `assignment_id` (optional, string); Path parameter assignmentId for
  api_v1_realm_assignments_assignment_id.
- `auth_server_id` (optional, string); Path parameter authServerId for
  api_v1_authorization_servers_auth_server_id.
- `authenticator_id` (optional, string); Path parameter authenticatorId for
  api_v1_authenticators_authenticator_id.
- `base_url` (required, string); format `uri`; Okta org API base URL, e.g.
  https://your-org.okta.com.
- `behavior_id` (optional, string); Path parameter behaviorId for api_v1_behaviors_behavior_id.
- `brand_id` (optional, string); Path parameter brandId for api_v1_brands_brand_id.
- `bundle_id` (optional, string); Path parameter bundleId for
  api_v1_iam_governance_bundles_bundle_id.
- `captcha_id` (optional, string); Path parameter captchaId for api_v1_captchas_captcha_id.
- `claim_id` (optional, string); Path parameter claimId for
  api_v1_apps_app_id_federated_claims_claim_id.
- `client_id` (optional, string); Path parameter clientId for
  api_v1_authorization_servers_auth_server_id_clients_client_id_tokens.
- `connection_id` (optional, string); Path parameter connectionId for
  api_v1_apps_app_id_cwo_connections_connection_id.
- `contact_type` (optional, string); Path parameter contactType for
  api_v1_org_contacts_contact_type.
- `csr_id` (optional, string); Path parameter csrId for api_v1_apps_app_id_credentials_csrs_csr_id.
- `custom_telephony_provider_id` (optional, string); Path parameter customTelephonyProviderId for
  api_v1_telephony_providers_custom_telephony_provider_id.
- `customization_id` (optional, string); Path parameter customizationId for
  api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id.
- `device_assurance_id` (optional, string); Path parameter deviceAssuranceId for
  api_v1_device_assurances_device_assurance_id.
- `device_id` (optional, string); Path parameter deviceId for api_v1_devices_device_id.
- `device_integration_id` (optional, string); Path parameter deviceIntegrationId for
  api_v1_device_integrations_device_integration_id.
- `domain` (optional, string); Path parameter domain for api_v1_dr_status_domain.
- `domain_id` (optional, string); Path parameter domainId for api_v1_domains_domain_id.
- `email_domain_id` (optional, string); Path parameter emailDomainId for
  api_v1_email_domains_email_domain_id.
- `email_server_id` (optional, string); Path parameter emailServerId for
  api_v1_email_servers_email_server_id.
- `enrollment_id` (optional, string); Path parameter enrollmentId for
  api_v1_users_user_id_authenticator_enrollments_enrollment_id.
- `entitlement_id` (optional, string); Path parameter entitlementId for
  api_v1_iam_governance_bundles_bundle_id_entitlements_entitlement_id_values.
- `event_hook_id` (optional, string); Path parameter eventHookId for
  api_v1_event_hooks_event_hook_id.
- `external_id` (optional, string); Path parameter externalId for
  api_v1_identity_sources_identity_source_id_users_external_id.
- `factor_id` (optional, string); Path parameter factorId for
  api_v1_users_user_id_factors_factor_id.
- `feature_id` (optional, string); Path parameter featureId for api_v1_features_feature_id.
- `feature_name` (optional, string); Path parameter featureName for
  api_v1_apps_app_id_features_feature_name.
- `filter` (optional, string); Required query parameter filter for api_v1_principal_rate_limits.
- `grant_id` (optional, string); Path parameter grantId for api_v1_apps_app_id_grants_grant_id.
- `group_id` (optional, string); Path parameter groupId for api_v1_apps_app_id_groups_group_id.
- `group_or_external_id` (optional, string); Path parameter groupOrExternalId for
  api_v1_identity_sources_identity_source_id_groups_group_or_external_id.
- `group_rule_id` (optional, string); Path parameter groupRuleId for
  api_v1_groups_rules_group_rule_id.
- `id` (optional, string); Path parameter id for api_v1_hook_keys_id.
- `identity_source_id` (optional, string); Path parameter identitySourceId for
  api_v1_identity_sources_identity_source_id_groups_group_or_external_id.
- `idp_csr_id` (optional, string); Path parameter idpCsrId for
  api_v1_idps_idp_id_credentials_csrs_idp_csr_id.
- `idp_id` (optional, string); Path parameter idpId for api_v1_idps_idp_id.
- `inline_hook_id` (optional, string); Path parameter inlineHookId for
  api_v1_inline_hooks_inline_hook_id.
- `key_id` (optional, string); Path parameter keyId for api_v1_apps_app_id_credentials_jwks_key_id.
- `kid` (optional, string); Path parameter kid for api_v1_idps_credentials_keys_kid.
- `linked_object_name` (optional, string); Path parameter linkedObjectName for
  api_v1_meta_schemas_user_linked_objects_linked_object_name.
- `log_stream_id` (optional, string); Path parameter logStreamId for
  api_v1_log_streams_log_stream_id.
- `log_stream_type` (optional, string); Path parameter logStreamType for
  api_v1_meta_schemas_log_stream_log_stream_type.
- `mapping_id` (optional, string); Path parameter mappingId for
  api_v1_apps_app_id_group_push_mappings_mapping_id.
- `member_id` (optional, string); Path parameter memberId for
  api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members_member_id.
- `method_type` (optional, string); Path parameter methodType for
  api_v1_authenticators_authenticator_id_methods_method_type.
- `notification_type` (optional, string); Path parameter notificationType for
  api_v1_roles_role_ref_subscriptions_notification_type.
- `oauth_client_id` (optional, string); Required query parameter oauthClientId for
  well_known_app_authenticator_configuration.
- `path` (optional, string); Path parameter path for api_v1_brands_brand_id_well_known_uris_path.
- `permission_type` (optional, string); Path parameter permissionType for
  api_v1_iam_roles_role_id_or_label_permissions_permission_type.
- `policy_id` (optional, string); Path parameter policyId for
  api_v1_authorization_servers_auth_server_id_policies_policy_id.
- `pool_id` (optional, string); Path parameter poolId for api_v1_agent_pools_pool_id_updates.
- `posture_check_id` (optional, string); Path parameter postureCheckId for
  api_v1_device_posture_checks_posture_check_id.
- `principal_rate_limit_id` (optional, string); Path parameter principalRateLimitId for
  api_v1_principal_rate_limits_principal_rate_limit_id.
- `push_provider_id` (optional, string); Path parameter pushProviderId for
  api_v1_push_providers_push_provider_id.
- `realm_id` (optional, string); Path parameter realmId for api_v1_realms_realm_id.
- `relationship_name` (optional, string); Path parameter relationshipName for
  api_v1_users_user_id_or_login_linked_objects_relationship_name.
- `resource_id` (optional, string); Path parameter resourceId for
  api_v1_iam_resource_sets_resource_set_id_or_label_resources_resource_id.
- `resource_set_id_or_label` (optional, string); Path parameter resourceSetIdOrLabel for
  api_v1_iam_resource_sets_resource_set_id_or_label.
- `result_id` (optional, string); Path parameter resultId for
  api_v1_directories_app_instance_id_groups_group_id_query_result_id.
- `role_assignment_id` (optional, string); Path parameter roleAssignmentId for
  api_v1_groups_group_id_roles_role_assignment_id.
- `role_id_or_encoded_role_id` (optional, string); Path parameter roleIdOrEncodedRoleId for
  api_v1_users_user_id_roles_role_id_or_encoded_role_id_targets.
- `role_id_or_label` (optional, string); Path parameter roleIdOrLabel for
  api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label.
- `role_ref` (optional, string); Path parameter roleRef for api_v1_roles_role_ref_subscriptions.
- `rule_id` (optional, string); Path parameter ruleId for
  api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id.
- `schema_id` (optional, string); Path parameter schemaId for api_v1_meta_schemas_user_schema_id.
- `scope_id` (optional, string); Path parameter scopeId for
  api_v1_authorization_servers_auth_server_id_scopes_scope_id.
- `secret_id` (optional, string); Path parameter secretId for
  api_v1_apps_app_id_credentials_secrets_secret_id.
- `security_event_provider_id` (optional, string); Path parameter securityEventProviderId for
  api_v1_security_events_providers_security_event_provider_id.
- `session_id` (optional, string); Path parameter sessionId for
  api_v1_identity_sources_identity_source_id_sessions_session_id.
- `start_date` (optional, string); format `date-time`; Optional lower bound for system log since
  filtering.
- `stream_id` (optional, string); Required query parameter stream_id for api_v1_ssf_stream_status.
- `template_id` (optional, string); Path parameter templateId for api_v1_templates_sms_template_id.
- `template_name` (optional, string); Path parameter templateName for
  api_v1_brands_brand_id_templates_email_template_name.
- `theme_id` (optional, string); Path parameter themeId for api_v1_brands_brand_id_themes_theme_id.
- `token_id` (optional, string); Path parameter tokenId for api_v1_apps_app_id_tokens_token_id.
- `transaction_id` (optional, string); Path parameter transactionId for
  api_v1_users_user_id_factors_factor_id_transactions_transaction_id.
- `trusted_origin_id` (optional, string); Path parameter trustedOriginId for
  api_v1_trusted_origins_trusted_origin_id.
- `type` (optional, string); Required query parameter type for api_v1_policies.
- `type_id` (optional, string); Path parameter typeId for api_v1_meta_types_user_type_id.
- `update_id` (optional, string); Path parameter updateId for
  api_v1_agent_pools_pool_id_updates_update_id.
- `user_id` (optional, string); Path parameter userId for api_v1_apps_app_id_users_user_id.
- `user_id_or_login` (optional, string); Path parameter userIdOrLogin for
  api_v1_users_user_id_or_login_linked_objects_relationship_name.
- `zone_id` (optional, string); Path parameter zoneId for api_v1_zones_zone_id.

Secret fields are redacted in logs and write previews: `access_token`, `api_token`.

Authentication behavior:

- API key authentication in `Authorization` with prefix `SSWS` using `secrets.api_token` when `{{
  secrets.api_token }}`.
- Bearer token authentication using `secrets.access_token` when `{{ secrets.access_token }}`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/api/v1/users` with query `limit`=`1`.

## Streams notes

Default pagination: single request; no pagination.

Pagination by stream: link_header: `users`, `groups`, `system_logs`, `api_v1_agent_pools`,
`api_v1_agent_pools_pool_id_updates`, `api_v1_api_tokens`, `api_v1_apps`,
`api_v1_apps_app_id_credentials_csrs`, `api_v1_apps_app_id_credentials_jwks`,
`api_v1_apps_app_id_credentials_keys`, `api_v1_apps_app_id_credentials_secrets`,
`api_v1_apps_app_id_cwo_connections`, `api_v1_apps_app_id_features`,
`api_v1_apps_app_id_federated_claims`, `api_v1_apps_app_id_grants`,
`api_v1_apps_app_id_group_push_mappings`, `api_v1_apps_app_id_groups`, `api_v1_apps_app_id_tokens`,
`api_v1_apps_app_id_users`, `api_v1_authenticators`,
`api_v1_authenticators_authenticator_id_aaguids`, `api_v1_authenticators_authenticator_id_methods`,
`api_v1_authorization_servers`, `api_v1_authorization_servers_auth_server_id_associated_servers`,
`api_v1_authorization_servers_auth_server_id_claims`,
`api_v1_authorization_servers_auth_server_id_clients`,
`api_v1_authorization_servers_auth_server_id_clients_client_id_tokens`,
`api_v1_authorization_servers_auth_server_id_credentials_keys`,
`api_v1_authorization_servers_auth_server_id_policies`,
`api_v1_authorization_servers_auth_server_id_policies_policy_id_rules`,
`api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys`,
`api_v1_authorization_servers_auth_server_id_scopes`, `api_v1_behaviors`, `api_v1_brands`,
`api_v1_brands_brand_id_domains`, `api_v1_brands_brand_id_templates_email`,
`api_v1_brands_brand_id_templates_email_template_name_customizations`,
`api_v1_brands_brand_id_themes`, `api_v1_captchas`, `api_v1_device_assurances`,
`api_v1_device_integrations`, `api_v1_device_posture_checks`,
`api_v1_device_posture_checks_default`, `api_v1_devices`, `api_v1_devices_device_id_users`,
`api_v1_domains`, `api_v1_email_domains`, `api_v1_email_servers`, `api_v1_event_hooks`,
`api_v1_features`, `api_v1_features_feature_id_dependencies`,
`api_v1_features_feature_id_dependents`, `api_v1_groups_rules`, `api_v1_groups_group_id_apps`,
`api_v1_groups_group_id_owners`, `api_v1_groups_group_id_roles`,
`api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps`,
`api_v1_groups_group_id_roles_role_assignment_id_targets_groups`, `api_v1_groups_group_id_users`,
`api_v1_hook_keys`, and 72 more; none: `well_known_app_authenticator_configuration`,
`well_known_apple_app_site_association`, `well_known_assetlinks_json`,
`well_known_okta_organization`, `well_known_ssf_configuration`, `well_known_webauthn`,
`api_v1_agent_pools_pool_id_updates_settings`, `api_v1_agent_pools_pool_id_updates_update_id`,
`api_v1_api_tokens_api_token_id`, `api_v1_apps_app_id`, `api_v1_apps_app_id_connections_default`,
`api_v1_apps_app_id_connections_default_jwks`, `api_v1_apps_app_id_credentials_csrs_csr_id`,
`api_v1_apps_app_id_credentials_jwks_key_id`, `api_v1_apps_app_id_credentials_keys_key_id`,
`api_v1_apps_app_id_credentials_secrets_secret_id`,
`api_v1_apps_app_id_cwo_connections_connection_id`, `api_v1_apps_app_id_features_feature_name`,
`api_v1_apps_app_id_federated_claims_claim_id`, `api_v1_apps_app_id_grants_grant_id`,
`api_v1_apps_app_id_group_push_mappings_mapping_id`, `api_v1_apps_app_id_groups_group_id`,
`api_v1_apps_app_id_tokens_token_id`, `api_v1_apps_app_id_users_user_id`,
`api_v1_authenticators_authenticator_id`, `api_v1_authenticators_authenticator_id_aaguids_aaguid`,
`api_v1_authenticators_authenticator_id_methods_method_type`,
`api_v1_authorization_servers_auth_server_id`,
`api_v1_authorization_servers_auth_server_id_claims_claim_id`,
`api_v1_authorization_servers_auth_server_id_clients_client_id_tokens_token_id`,
`api_v1_authorization_servers_auth_server_id_credentials_keys_key_id`,
`api_v1_authorization_servers_auth_server_id_policies_policy_id`,
`api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id`,
`api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id`,
`api_v1_authorization_servers_auth_server_id_scopes_scope_id`, `api_v1_behaviors_behavior_id`,
`api_v1_bot_protection_configuration`, `api_v1_brands_brand_id`,
`api_v1_brands_brand_id_pages_error`, `api_v1_brands_brand_id_pages_error_customized`,
`api_v1_brands_brand_id_pages_error_default`, `api_v1_brands_brand_id_pages_error_preview`,
`api_v1_brands_brand_id_pages_sign_in`, `api_v1_brands_brand_id_pages_sign_in_customized`,
`api_v1_brands_brand_id_pages_sign_in_default`, `api_v1_brands_brand_id_pages_sign_in_preview`,
`api_v1_brands_brand_id_pages_sign_out_customized`,
`api_v1_brands_brand_id_templates_email_template_name`,
`api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id`,
`api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id_preview`,
`api_v1_brands_brand_id_templates_email_template_name_default_content`,
`api_v1_brands_brand_id_templates_email_template_name_default_content_preview`,
`api_v1_brands_brand_id_templates_email_template_name_settings`,
`api_v1_brands_brand_id_themes_theme_id`, `api_v1_brands_brand_id_well_known_uris`,
`api_v1_brands_brand_id_well_known_uris_path`,
`api_v1_brands_brand_id_well_known_uris_path_customized`, `api_v1_captchas_captcha_id`,
`api_v1_device_assurances_device_assurance_id`, `api_v1_device_integrations_device_integration_id`,
and 92 more.

Incremental streams use their declared cursor fields and send lower-bound parameters only when a
lower bound is available.

- `users`: GET `/api/v1/users` - records path `.`; query `limit`=`200`; follows RFC 5988 Link
  headers with rel=next; computed output fields `email`, `last_login`, `login`.
- `groups`: GET `/api/v1/groups` - records path `.`; query `limit`=`200`; follows RFC 5988 Link
  headers with rel=next; computed output fields `description`, `name`.
- `system_logs`: GET `/api/v1/logs` - records path `.`; query `limit`=`200`; follows RFC 5988 Link
  headers with rel=next; incremental cursor `published`; sent as `since`; formatted as `rfc3339`;
  initial lower bound from `start_date`; computed output fields `display_message`, `event_type`.
- `well_known_app_authenticator_configuration`: GET `/.well-known/app-authenticator-configuration` -
  records path `.`; query `oauthClientId`=`{{ config.oauth_client_id }}`; emits passthrough records.
- `well_known_apple_app_site_association`: GET `/.well-known/apple-app-site-association` - records
  at response root; emits passthrough records.
- `well_known_assetlinks_json`: GET `/.well-known/assetlinks.json` - records path `.`; emits
  passthrough records.
- `well_known_okta_organization`: GET `/.well-known/okta-organization` - records at response root;
  emits passthrough records.
- `well_known_ssf_configuration`: GET `/.well-known/ssf-configuration` - records at response root;
  emits passthrough records.
- `well_known_webauthn`: GET `/.well-known/webauthn` - records at response root; emits passthrough
  records.
- `api_v1_agent_pools`: GET `/api/v1/agentPools` - records path `.`; query `limit`=`200`; follows
  RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_agent_pools_pool_id_updates`: GET `/api/v1/agentPools/{{ config.pool_id }}/updates` -
  records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `api_v1_agent_pools_pool_id_updates_settings`: GET `/api/v1/agentPools/{{ config.pool_id
  }}/updates/settings` - records at response root; emits passthrough records.
- `api_v1_agent_pools_pool_id_updates_update_id`: GET `/api/v1/agentPools/{{ config.pool_id
  }}/updates/{{ config.update_id }}` - records at response root; emits passthrough records.
- `api_v1_api_tokens`: GET `/api/v1/api-tokens` - records path `.`; query `limit`=`200`; follows RFC
  5988 Link headers with rel=next; emits passthrough records.
- `api_v1_api_tokens_api_token_id`: GET `/api/v1/api-tokens/{{ config.api_token_id }}` - records at
  response root; emits passthrough records.
- `api_v1_apps`: GET `/api/v1/apps` - records path `.`; query `limit`=`200`; follows RFC 5988 Link
  headers with rel=next; emits passthrough records.
- `api_v1_apps_app_id`: GET `/api/v1/apps/{{ config.app_id }}` - records at response root; emits
  passthrough records.
- `api_v1_apps_app_id_connections_default`: GET `/api/v1/apps/{{ config.app_id
  }}/connections/default` - records at response root; emits passthrough records.
- `api_v1_apps_app_id_connections_default_jwks`: GET `/api/v1/apps/{{ config.app_id
  }}/connections/default/jwks` - records at response root; emits passthrough records.
- `api_v1_apps_app_id_credentials_csrs`: GET `/api/v1/apps/{{ config.app_id }}/credentials/csrs` -
  records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `api_v1_apps_app_id_credentials_csrs_csr_id`: GET `/api/v1/apps/{{ config.app_id
  }}/credentials/csrs/{{ config.csr_id }}` - records at response root; emits passthrough records.
- `api_v1_apps_app_id_credentials_jwks`: GET `/api/v1/apps/{{ config.app_id }}/credentials/jwks` -
  records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `api_v1_apps_app_id_credentials_jwks_key_id`: GET `/api/v1/apps/{{ config.app_id
  }}/credentials/jwks/{{ config.key_id }}` - records at response root; emits passthrough records.
- `api_v1_apps_app_id_credentials_keys`: GET `/api/v1/apps/{{ config.app_id }}/credentials/keys` -
  records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `api_v1_apps_app_id_credentials_keys_key_id`: GET `/api/v1/apps/{{ config.app_id
  }}/credentials/keys/{{ config.key_id }}` - records at response root; emits passthrough records.
- `api_v1_apps_app_id_credentials_secrets`: GET `/api/v1/apps/{{ config.app_id
  }}/credentials/secrets` - records path `.`; query `limit`=`200`; follows RFC 5988 Link headers
  with rel=next; emits passthrough records.
- `api_v1_apps_app_id_credentials_secrets_secret_id`: GET `/api/v1/apps/{{ config.app_id
  }}/credentials/secrets/{{ config.secret_id }}` - records at response root; emits passthrough
  records.
- `api_v1_apps_app_id_cwo_connections`: GET `/api/v1/apps/{{ config.app_id }}/cwo/connections` -
  records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `api_v1_apps_app_id_cwo_connections_connection_id`: GET `/api/v1/apps/{{ config.app_id
  }}/cwo/connections/{{ config.connection_id }}` - records at response root; emits passthrough
  records.
- `api_v1_apps_app_id_features`: GET `/api/v1/apps/{{ config.app_id }}/features` - records path `.`;
  query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_apps_app_id_features_feature_name`: GET `/api/v1/apps/{{ config.app_id }}/features/{{
  config.feature_name }}` - records at response root; emits passthrough records.
- `api_v1_apps_app_id_federated_claims`: GET `/api/v1/apps/{{ config.app_id }}/federated-claims` -
  records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `api_v1_apps_app_id_federated_claims_claim_id`: GET `/api/v1/apps/{{ config.app_id
  }}/federated-claims/{{ config.claim_id }}` - records at response root; emits passthrough records.
- `api_v1_apps_app_id_grants`: GET `/api/v1/apps/{{ config.app_id }}/grants` - records path `.`;
  query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_apps_app_id_grants_grant_id`: GET `/api/v1/apps/{{ config.app_id }}/grants/{{
  config.grant_id }}` - records at response root; emits passthrough records.
- `api_v1_apps_app_id_group_push_mappings`: GET `/api/v1/apps/{{ config.app_id
  }}/group-push/mappings` - records path `.`; query `limit`=`200`; follows RFC 5988 Link headers
  with rel=next; emits passthrough records.
- `api_v1_apps_app_id_group_push_mappings_mapping_id`: GET `/api/v1/apps/{{ config.app_id
  }}/group-push/mappings/{{ config.mapping_id }}` - records at response root; emits passthrough
  records.
- `api_v1_apps_app_id_groups`: GET `/api/v1/apps/{{ config.app_id }}/groups` - records path `.`;
  query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_apps_app_id_groups_group_id`: GET `/api/v1/apps/{{ config.app_id }}/groups/{{
  config.group_id }}` - records at response root; emits passthrough records.
- `api_v1_apps_app_id_tokens`: GET `/api/v1/apps/{{ config.app_id }}/tokens` - records path `.`;
  query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_apps_app_id_tokens_token_id`: GET `/api/v1/apps/{{ config.app_id }}/tokens/{{
  config.token_id }}` - records path `.`; emits passthrough records.
- `api_v1_apps_app_id_users`: GET `/api/v1/apps/{{ config.app_id }}/users` - records path `.`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_apps_app_id_users_user_id`: GET `/api/v1/apps/{{ config.app_id }}/users/{{ config.user_id
  }}` - records at response root; emits passthrough records.
- `api_v1_authenticators`: GET `/api/v1/authenticators` - records path `.`; query `limit`=`200`;
  follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_authenticators_authenticator_id`: GET `/api/v1/authenticators/{{ config.authenticator_id
  }}` - records at response root; emits passthrough records.
- `api_v1_authenticators_authenticator_id_aaguids`: GET `/api/v1/authenticators/{{
  config.authenticator_id }}/aaguids` - records path `.`; query `limit`=`200`; follows RFC 5988 Link
  headers with rel=next; emits passthrough records.
- `api_v1_authenticators_authenticator_id_aaguids_aaguid`: GET `/api/v1/authenticators/{{
  config.authenticator_id }}/aaguids/{{ config.aaguid }}` - records at response root; emits
  passthrough records.
- `api_v1_authenticators_authenticator_id_methods`: GET `/api/v1/authenticators/{{
  config.authenticator_id }}/methods` - records path `.`; query `limit`=`200`; follows RFC 5988 Link
  headers with rel=next; emits passthrough records.
- `api_v1_authenticators_authenticator_id_methods_method_type`: GET `/api/v1/authenticators/{{
  config.authenticator_id }}/methods/{{ config.method_type }}` - records at response root; emits
  passthrough records.
- `api_v1_authorization_servers`: GET `/api/v1/authorizationServers` - records path `.`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_authorization_servers_auth_server_id`: GET `/api/v1/authorizationServers/{{
  config.auth_server_id }}` - records at response root; emits passthrough records.
- `api_v1_authorization_servers_auth_server_id_associated_servers`: GET
  `/api/v1/authorizationServers/{{ config.auth_server_id }}/associatedServers` - records path `.`;
  query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_authorization_servers_auth_server_id_claims`: GET `/api/v1/authorizationServers/{{
  config.auth_server_id }}/claims` - records path `.`; query `limit`=`200`; follows RFC 5988 Link
  headers with rel=next; emits passthrough records.
- `api_v1_authorization_servers_auth_server_id_claims_claim_id`: GET
  `/api/v1/authorizationServers/{{ config.auth_server_id }}/claims/{{ config.claim_id }}` - records
  at response root; emits passthrough records.
- `api_v1_authorization_servers_auth_server_id_clients`: GET `/api/v1/authorizationServers/{{
  config.auth_server_id }}/clients` - records path `.`; query `limit`=`200`; follows RFC 5988 Link
  headers with rel=next; emits passthrough records.
- `api_v1_authorization_servers_auth_server_id_clients_client_id_tokens`: GET
  `/api/v1/authorizationServers/{{ config.auth_server_id }}/clients/{{ config.client_id }}/tokens` -
  records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `api_v1_authorization_servers_auth_server_id_clients_client_id_tokens_token_id`: GET
  `/api/v1/authorizationServers/{{ config.auth_server_id }}/clients/{{ config.client_id }}/tokens/{{
  config.token_id }}` - records path `.`; emits passthrough records.
- `api_v1_authorization_servers_auth_server_id_credentials_keys`: GET
  `/api/v1/authorizationServers/{{ config.auth_server_id }}/credentials/keys` - records path `.`;
  query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_authorization_servers_auth_server_id_credentials_keys_key_id`: GET
  `/api/v1/authorizationServers/{{ config.auth_server_id }}/credentials/keys/{{ config.key_id }}` -
  records at response root; emits passthrough records.
- `api_v1_authorization_servers_auth_server_id_policies`: GET `/api/v1/authorizationServers/{{
  config.auth_server_id }}/policies` - records path `.`; query `limit`=`200`; follows RFC 5988 Link
  headers with rel=next; emits passthrough records.
- `api_v1_authorization_servers_auth_server_id_policies_policy_id`: GET
  `/api/v1/authorizationServers/{{ config.auth_server_id }}/policies/{{ config.policy_id }}` -
  records at response root; emits passthrough records.
- `api_v1_authorization_servers_auth_server_id_policies_policy_id_rules`: GET
  `/api/v1/authorizationServers/{{ config.auth_server_id }}/policies/{{ config.policy_id }}/rules` -
  records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id`: GET
  `/api/v1/authorizationServers/{{ config.auth_server_id }}/policies/{{ config.policy_id }}/rules/{{
  config.rule_id }}` - records at response root; emits passthrough records.
- `api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys`: GET
  `/api/v1/authorizationServers/{{ config.auth_server_id }}/resourceservercredentials/keys` -
  records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id`: GET
  `/api/v1/authorizationServers/{{ config.auth_server_id }}/resourceservercredentials/keys/{{
  config.key_id }}` - records at response root; emits passthrough records.
- `api_v1_authorization_servers_auth_server_id_scopes`: GET `/api/v1/authorizationServers/{{
  config.auth_server_id }}/scopes` - records path `.`; query `limit`=`200`; follows RFC 5988 Link
  headers with rel=next; emits passthrough records.
- `api_v1_authorization_servers_auth_server_id_scopes_scope_id`: GET
  `/api/v1/authorizationServers/{{ config.auth_server_id }}/scopes/{{ config.scope_id }}` - records
  at response root; emits passthrough records.
- `api_v1_behaviors`: GET `/api/v1/behaviors` - records path `.`; query `limit`=`200`; follows RFC
  5988 Link headers with rel=next; emits passthrough records.
- `api_v1_behaviors_behavior_id`: GET `/api/v1/behaviors/{{ config.behavior_id }}` - records at
  response root; emits passthrough records.
- `api_v1_bot_protection_configuration`: GET `/api/v1/bot-protection/configuration` - records at
  response root; emits passthrough records.
- `api_v1_brands`: GET `/api/v1/brands` - records path `.`; query `limit`=`200`; follows RFC 5988
  Link headers with rel=next; emits passthrough records.
- `api_v1_brands_brand_id`: GET `/api/v1/brands/{{ config.brand_id }}` - records at response root;
  emits passthrough records.
- `api_v1_brands_brand_id_domains`: GET `/api/v1/brands/{{ config.brand_id }}/domains` - records at
  response root; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough
  records.
- `api_v1_brands_brand_id_pages_error`: GET `/api/v1/brands/{{ config.brand_id }}/pages/error` -
  records at response root; emits passthrough records.
- `api_v1_brands_brand_id_pages_error_customized`: GET `/api/v1/brands/{{ config.brand_id
  }}/pages/error/customized` - records at response root; emits passthrough records.
- `api_v1_brands_brand_id_pages_error_default`: GET `/api/v1/brands/{{ config.brand_id
  }}/pages/error/default` - records at response root; emits passthrough records.
- `api_v1_brands_brand_id_pages_error_preview`: GET `/api/v1/brands/{{ config.brand_id
  }}/pages/error/preview` - records at response root; emits passthrough records.
- `api_v1_brands_brand_id_pages_sign_in`: GET `/api/v1/brands/{{ config.brand_id }}/pages/sign-in` -
  records at response root; emits passthrough records.
- `api_v1_brands_brand_id_pages_sign_in_customized`: GET `/api/v1/brands/{{ config.brand_id
  }}/pages/sign-in/customized` - records at response root; emits passthrough records.
- `api_v1_brands_brand_id_pages_sign_in_default`: GET `/api/v1/brands/{{ config.brand_id
  }}/pages/sign-in/default` - records at response root; emits passthrough records.
- `api_v1_brands_brand_id_pages_sign_in_preview`: GET `/api/v1/brands/{{ config.brand_id
  }}/pages/sign-in/preview` - records at response root; emits passthrough records.
- `api_v1_brands_brand_id_pages_sign_out_customized`: GET `/api/v1/brands/{{ config.brand_id
  }}/pages/sign-out/customized` - records at response root; emits passthrough records.
- `api_v1_brands_brand_id_templates_email`: GET `/api/v1/brands/{{ config.brand_id
  }}/templates/email` - records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with
  rel=next; emits passthrough records.
- `api_v1_brands_brand_id_templates_email_template_name`: GET `/api/v1/brands/{{ config.brand_id
  }}/templates/email/{{ config.template_name }}` - records at response root; emits passthrough
  records.
- `api_v1_brands_brand_id_templates_email_template_name_customizations`: GET `/api/v1/brands/{{
  config.brand_id }}/templates/email/{{ config.template_name }}/customizations` - records path `.`;
  query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id`: GET
  `/api/v1/brands/{{ config.brand_id }}/templates/email/{{ config.template_name }}/customizations/{{
  config.customization_id }}` - records at response root; emits passthrough records.
- `api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id_preview`:
  GET `/api/v1/brands/{{ config.brand_id }}/templates/email/{{ config.template_name
  }}/customizations/{{ config.customization_id }}/preview` - records at response root; emits
  passthrough records.
- `api_v1_brands_brand_id_templates_email_template_name_default_content`: GET `/api/v1/brands/{{
  config.brand_id }}/templates/email/{{ config.template_name }}/default-content` - records at
  response root; emits passthrough records.
- `api_v1_brands_brand_id_templates_email_template_name_default_content_preview`: GET
  `/api/v1/brands/{{ config.brand_id }}/templates/email/{{ config.template_name
  }}/default-content/preview` - records at response root; emits passthrough records.
- `api_v1_brands_brand_id_templates_email_template_name_settings`: GET `/api/v1/brands/{{
  config.brand_id }}/templates/email/{{ config.template_name }}/settings` - records at response
  root; emits passthrough records.
- `api_v1_brands_brand_id_themes`: GET `/api/v1/brands/{{ config.brand_id }}/themes` - records path
  `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_brands_brand_id_themes_theme_id`: GET `/api/v1/brands/{{ config.brand_id }}/themes/{{
  config.theme_id }}` - records at response root; emits passthrough records.
- `api_v1_brands_brand_id_well_known_uris`: GET `/api/v1/brands/{{ config.brand_id
  }}/well-known-uris` - records at response root; emits passthrough records.
- `api_v1_brands_brand_id_well_known_uris_path`: GET `/api/v1/brands/{{ config.brand_id
  }}/well-known-uris/{{ config.path }}` - records at response root; emits passthrough records.
- `api_v1_brands_brand_id_well_known_uris_path_customized`: GET `/api/v1/brands/{{ config.brand_id
  }}/well-known-uris/{{ config.path }}/customized` - records at response root; emits passthrough
  records.
- `api_v1_captchas`: GET `/api/v1/captchas` - records path `.`; query `limit`=`200`; follows RFC
  5988 Link headers with rel=next; emits passthrough records.
- `api_v1_captchas_captcha_id`: GET `/api/v1/captchas/{{ config.captcha_id }}` - records at response
  root; emits passthrough records.
- `api_v1_device_assurances`: GET `/api/v1/device-assurances` - records path `.`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_device_assurances_device_assurance_id`: GET `/api/v1/device-assurances/{{
  config.device_assurance_id }}` - records at response root; emits passthrough records.
- `api_v1_device_integrations`: GET `/api/v1/device-integrations` - records path `.`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_device_integrations_device_integration_id`: GET `/api/v1/device-integrations/{{
  config.device_integration_id }}` - records at response root; emits passthrough records.
- `api_v1_device_posture_checks`: GET `/api/v1/device-posture-checks` - records path `.`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_device_posture_checks_default`: GET `/api/v1/device-posture-checks/default` - records path
  `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_device_posture_checks_posture_check_id`: GET `/api/v1/device-posture-checks/{{
  config.posture_check_id }}` - records at response root; emits passthrough records.
- `api_v1_devices`: GET `/api/v1/devices` - records path `.`; query `limit`=`200`; follows RFC 5988
  Link headers with rel=next; emits passthrough records.
- `api_v1_devices_device_id`: GET `/api/v1/devices/{{ config.device_id }}` - records at response
  root; emits passthrough records.
- `api_v1_devices_device_id_users`: GET `/api/v1/devices/{{ config.device_id }}/users` - records
  path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough
  records.
- `api_v1_directories_app_instance_id_groups_group_id_query_result_id`: GET `/api/v1/directories/{{
  config.app_instance_id }}/groups/{{ config.group_id }}/query/{{ config.result_id }}` - records at
  response root; emits passthrough records.
- `api_v1_domains`: GET `/api/v1/domains` - records at response root; query `limit`=`200`; follows
  RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_domains_domain_id`: GET `/api/v1/domains/{{ config.domain_id }}` - records at response
  root; emits passthrough records.
- `api_v1_dr_status`: GET `/api/v1/dr/status` - records at response root; emits passthrough records.
- `api_v1_dr_status_domain`: GET `/api/v1/dr/status/{{ config.domain }}` - records at response root;
  emits passthrough records.
- `api_v1_email_domains`: GET `/api/v1/email-domains` - records path `.`; query `limit`=`200`;
  follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_email_domains_email_domain_id`: GET `/api/v1/email-domains/{{ config.email_domain_id }}` -
  records at response root; emits passthrough records.
- `api_v1_email_servers`: GET `/api/v1/email-servers` - records at response root; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_email_servers_email_server_id`: GET `/api/v1/email-servers/{{ config.email_server_id }}` -
  records at response root; emits passthrough records.
- `api_v1_event_hooks`: GET `/api/v1/eventHooks` - records path `.`; query `limit`=`200`; follows
  RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_event_hooks_event_hook_id`: GET `/api/v1/eventHooks/{{ config.event_hook_id }}` - records
  at response root; emits passthrough records.
- `api_v1_features`: GET `/api/v1/features` - records path `.`; query `limit`=`200`; follows RFC
  5988 Link headers with rel=next; emits passthrough records.
- `api_v1_features_feature_id`: GET `/api/v1/features/{{ config.feature_id }}` - records at response
  root; emits passthrough records.
- `api_v1_features_feature_id_dependencies`: GET `/api/v1/features/{{ config.feature_id
  }}/dependencies` - records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with
  rel=next; emits passthrough records.
- `api_v1_features_feature_id_dependents`: GET `/api/v1/features/{{ config.feature_id }}/dependents`
  - records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `api_v1_first_party_app_settings_app_name`: GET `/api/v1/first-party-app-settings/{{
  config.app_name }}` - records at response root; emits passthrough records.
- `api_v1_groups_rules`: GET `/api/v1/groups/rules` - records path `.`; query `limit`=`200`; follows
  RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_groups_rules_group_rule_id`: GET `/api/v1/groups/rules/{{ config.group_rule_id }}` -
  records at response root; emits passthrough records.
- `api_v1_groups_group_id`: GET `/api/v1/groups/{{ config.group_id }}` - records at response root;
  emits passthrough records.
- `api_v1_groups_group_id_apps`: GET `/api/v1/groups/{{ config.group_id }}/apps` - records path `.`;
  query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_groups_group_id_owners`: GET `/api/v1/groups/{{ config.group_id }}/owners` - records path
  `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_groups_group_id_roles`: GET `/api/v1/groups/{{ config.group_id }}/roles` - records path
  `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_groups_group_id_roles_role_assignment_id`: GET `/api/v1/groups/{{ config.group_id
  }}/roles/{{ config.role_assignment_id }}` - records at response root; emits passthrough records.
- `api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps`: GET `/api/v1/groups/{{
  config.group_id }}/roles/{{ config.role_assignment_id }}/targets/catalog/apps` - records path `.`;
  query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_groups_group_id_roles_role_assignment_id_targets_groups`: GET `/api/v1/groups/{{
  config.group_id }}/roles/{{ config.role_assignment_id }}/targets/groups` - records path `.`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_groups_group_id_users`: GET `/api/v1/groups/{{ config.group_id }}/users` - records path
  `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_hook_keys`: GET `/api/v1/hook-keys` - records path `.`; query `limit`=`200`; follows RFC
  5988 Link headers with rel=next; emits passthrough records.
- `api_v1_hook_keys_public_key_id`: GET `/api/v1/hook-keys/public/{{ config.key_id }}` - records at
  response root; emits passthrough records.
- `api_v1_hook_keys_id`: GET `/api/v1/hook-keys/{{ config.id }}` - records at response root; emits
  passthrough records.
- `api_v1_iam_assignees_users`: GET `/api/v1/iam/assignees/users` - records path `value`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_iam_governance_bundles`: GET `/api/v1/iam/governance/bundles` - records at response root;
  query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_iam_governance_bundles_bundle_id`: GET `/api/v1/iam/governance/bundles/{{ config.bundle_id
  }}` - records at response root; emits passthrough records.
- `api_v1_iam_governance_bundles_bundle_id_entitlements`: GET `/api/v1/iam/governance/bundles/{{
  config.bundle_id }}/entitlements` - records at response root; query `limit`=`200`; follows RFC
  5988 Link headers with rel=next; emits passthrough records.
- `api_v1_iam_governance_bundles_bundle_id_entitlements_entitlement_id_values`: GET
  `/api/v1/iam/governance/bundles/{{ config.bundle_id }}/entitlements/{{ config.entitlement_id
  }}/values` - records at response root; query `limit`=`200`; follows RFC 5988 Link headers with
  rel=next; emits passthrough records.
- `api_v1_iam_governance_opt_in`: GET `/api/v1/iam/governance/optIn` - records at response root;
  emits passthrough records.
- `api_v1_iam_resource_sets`: GET `/api/v1/iam/resource-sets` - records at response root; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_iam_resource_sets_resource_set_id_or_label`: GET `/api/v1/iam/resource-sets/{{
  config.resource_set_id_or_label }}` - records at response root; emits passthrough records.
- `api_v1_iam_resource_sets_resource_set_id_or_label_bindings`: GET `/api/v1/iam/resource-sets/{{
  config.resource_set_id_or_label }}/bindings` - records path `roles`; query `limit`=`200`; follows
  RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label`: GET
  `/api/v1/iam/resource-sets/{{ config.resource_set_id_or_label }}/bindings/{{
  config.role_id_or_label }}` - records at response root; emits passthrough records.
- `api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members`: GET
  `/api/v1/iam/resource-sets/{{ config.resource_set_id_or_label }}/bindings/{{
  config.role_id_or_label }}/members` - records at response root; query `limit`=`200`; follows RFC
  5988 Link headers with rel=next; emits passthrough records.
- `api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members_member_id`:
  GET `/api/v1/iam/resource-sets/{{ config.resource_set_id_or_label }}/bindings/{{
  config.role_id_or_label }}/members/{{ config.member_id }}` - records at response root; emits
  passthrough records.
- `api_v1_iam_resource_sets_resource_set_id_or_label_resources`: GET `/api/v1/iam/resource-sets/{{
  config.resource_set_id_or_label }}/resources` - records at response root; query `limit`=`200`;
  follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_iam_resource_sets_resource_set_id_or_label_resources_resource_id`: GET
  `/api/v1/iam/resource-sets/{{ config.resource_set_id_or_label }}/resources/{{ config.resource_id
  }}` - records at response root; emits passthrough records.
- `api_v1_iam_roles`: GET `/api/v1/iam/roles` - records path `roles`; query `limit`=`200`; follows
  RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_iam_roles_role_id_or_label`: GET `/api/v1/iam/roles/{{ config.role_id_or_label }}` -
  records at response root; emits passthrough records.
- `api_v1_iam_roles_role_id_or_label_permissions`: GET `/api/v1/iam/roles/{{ config.role_id_or_label
  }}/permissions` - records path `permissions`; query `limit`=`200`; follows RFC 5988 Link headers
  with rel=next; emits passthrough records.
- `api_v1_iam_roles_role_id_or_label_permissions_permission_type`: GET `/api/v1/iam/roles/{{
  config.role_id_or_label }}/permissions/{{ config.permission_type }}` - records at response root;
  emits passthrough records.
- `api_v1_identity_sources_identity_source_id_groups_group_or_external_id`: GET
  `/api/v1/identity-sources/{{ config.identity_source_id }}/groups/{{ config.group_or_external_id
  }}` - records at response root; emits passthrough records.
- `api_v1_identity_sources_identity_source_id_groups_group_or_external_id_membership`: GET
  `/api/v1/identity-sources/{{ config.identity_source_id }}/groups/{{ config.group_or_external_id
  }}/membership` - records at response root; query `limit`=`200`; follows RFC 5988 Link headers with
  rel=next; emits passthrough records.
- `api_v1_identity_sources_identity_source_id_sessions`: GET `/api/v1/identity-sources/{{
  config.identity_source_id }}/sessions` - records path `.`; query `limit`=`200`; follows RFC 5988
  Link headers with rel=next; emits passthrough records.
- `api_v1_identity_sources_identity_source_id_sessions_session_id`: GET `/api/v1/identity-sources/{{
  config.identity_source_id }}/sessions/{{ config.session_id }}` - records at response root; emits
  passthrough records.
- `api_v1_identity_sources_identity_source_id_users_external_id`: GET `/api/v1/identity-sources/{{
  config.identity_source_id }}/users/{{ config.external_id }}` - records at response root; emits
  passthrough records.
- `api_v1_idps`: GET `/api/v1/idps` - records path `.`; query `limit`=`200`; follows RFC 5988 Link
  headers with rel=next; emits passthrough records.
- `api_v1_idps_credentials_keys`: GET `/api/v1/idps/credentials/keys` - records path `.`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_idps_credentials_keys_kid`: GET `/api/v1/idps/credentials/keys/{{ config.kid }}` - records
  at response root; emits passthrough records.
- `api_v1_idps_idp_id`: GET `/api/v1/idps/{{ config.idp_id }}` - records at response root; emits
  passthrough records.
- `api_v1_idps_idp_id_credentials_csrs`: GET `/api/v1/idps/{{ config.idp_id }}/credentials/csrs` -
  records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `api_v1_idps_idp_id_credentials_csrs_idp_csr_id`: GET `/api/v1/idps/{{ config.idp_id
  }}/credentials/csrs/{{ config.idp_csr_id }}` - records at response root; emits passthrough
  records.
- `api_v1_idps_idp_id_credentials_keys`: GET `/api/v1/idps/{{ config.idp_id }}/credentials/keys` -
  records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `api_v1_idps_idp_id_credentials_keys_active`: GET `/api/v1/idps/{{ config.idp_id
  }}/credentials/keys/active` - records path `.`; query `limit`=`200`; follows RFC 5988 Link headers
  with rel=next; emits passthrough records.
- `api_v1_idps_idp_id_credentials_keys_kid`: GET `/api/v1/idps/{{ config.idp_id
  }}/credentials/keys/{{ config.kid }}` - records at response root; emits passthrough records.
- `api_v1_idps_idp_id_users`: GET `/api/v1/idps/{{ config.idp_id }}/users` - records path `.`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_idps_idp_id_users_user_id`: GET `/api/v1/idps/{{ config.idp_id }}/users/{{ config.user_id
  }}` - records at response root; emits passthrough records.
- `api_v1_idps_idp_id_users_user_id_credentials_tokens`: GET `/api/v1/idps/{{ config.idp_id
  }}/users/{{ config.user_id }}/credentials/tokens` - records path `.`; query `limit`=`200`; follows
  RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_inline_hooks`: GET `/api/v1/inlineHooks` - records path `.`; query `limit`=`200`; follows
  RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_inline_hooks_inline_hook_id`: GET `/api/v1/inlineHooks/{{ config.inline_hook_id }}` -
  records at response root; emits passthrough records.
- `api_v1_log_streams`: GET `/api/v1/logStreams` - records path `.`; query `limit`=`200`; follows
  RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_log_streams_log_stream_id`: GET `/api/v1/logStreams/{{ config.log_stream_id }}` - records
  at response root; emits passthrough records.
- `api_v1_mappings`: GET `/api/v1/mappings` - records path `.`; query `limit`=`200`; follows RFC
  5988 Link headers with rel=next; emits passthrough records.
- `api_v1_mappings_mapping_id`: GET `/api/v1/mappings/{{ config.mapping_id }}` - records at response
  root; emits passthrough records.
- `api_v1_meta_schemas_apps_app_id_default`: GET `/api/v1/meta/schemas/apps/{{ config.app_id
  }}/default` - records at response root; emits passthrough records.
- `api_v1_meta_schemas_group_default`: GET `/api/v1/meta/schemas/group/default` - records at
  response root; emits passthrough records.
- `api_v1_meta_schemas_log_stream`: GET `/api/v1/meta/schemas/logStream` - records path `.`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_meta_schemas_log_stream_log_stream_type`: GET `/api/v1/meta/schemas/logStream/{{
  config.log_stream_type }}` - records at response root; emits passthrough records.
- `api_v1_meta_schemas_user_linked_objects`: GET `/api/v1/meta/schemas/user/linkedObjects` - records
  path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough
  records.
- `api_v1_meta_schemas_user_linked_objects_linked_object_name`: GET
  `/api/v1/meta/schemas/user/linkedObjects/{{ config.linked_object_name }}` - records at response
  root; emits passthrough records.
- `api_v1_meta_schemas_user_schema_id`: GET `/api/v1/meta/schemas/user/{{ config.schema_id }}` -
  records at response root; emits passthrough records.
- `api_v1_meta_types_user`: GET `/api/v1/meta/types/user` - records path `.`; query `limit`=`200`;
  follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_meta_types_user_type_id`: GET `/api/v1/meta/types/user/{{ config.type_id }}` - records at
  response root; emits passthrough records.
- `api_v1_meta_uischemas`: GET `/api/v1/meta/uischemas` - records path `.`; query `limit`=`200`;
  follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_meta_uischemas_id`: GET `/api/v1/meta/uischemas/{{ config.id }}` - records at response
  root; emits passthrough records.
- `api_v1_org`: GET `/api/v1/org` - records at response root; emits passthrough records.
- `api_v1_org_captcha`: GET `/api/v1/org/captcha` - records at response root; emits passthrough
  records.
- `api_v1_org_contacts`: GET `/api/v1/org/contacts` - records path `.`; query `limit`=`200`; follows
  RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_org_contacts_contact_type`: GET `/api/v1/org/contacts/{{ config.contact_type }}` - records
  at response root; emits passthrough records.
- `api_v1_org_factors_yubikey_token_tokens`: GET `/api/v1/org/factors/yubikey_token/tokens` -
  records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `api_v1_org_factors_yubikey_token_tokens_token_id`: GET
  `/api/v1/org/factors/yubikey_token/tokens/{{ config.token_id }}` - records at response root; emits
  passthrough records.
- `api_v1_org_org_settings_third_party_admin_setting`: GET
  `/api/v1/org/orgSettings/thirdPartyAdminSetting` - records at response root; emits passthrough
  records.
- `api_v1_org_preferences`: GET `/api/v1/org/preferences` - records at response root; emits
  passthrough records.
- `api_v1_org_privacy_aerial`: GET `/api/v1/org/privacy/aerial` - records at response root; emits
  passthrough records.
- `api_v1_org_privacy_okta_communication`: GET `/api/v1/org/privacy/oktaCommunication` - records at
  response root; emits passthrough records.
- `api_v1_org_privacy_okta_support`: GET `/api/v1/org/privacy/oktaSupport` - records at response
  root; emits passthrough records.
- `api_v1_org_privacy_okta_support_cases`: GET `/api/v1/org/privacy/oktaSupport/cases` - records at
  response root; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough
  records.
- `api_v1_org_settings_auto_assign_admin_app_setting`: GET
  `/api/v1/org/settings/autoAssignAdminAppSetting` - records at response root; emits passthrough
  records.
- `api_v1_org_settings_client_privileges_setting`: GET
  `/api/v1/org/settings/clientPrivilegesSetting` - records at response root; emits passthrough
  records.
- `api_v1_policies`: GET `/api/v1/policies` - records path `.`; query `limit`=`200`; `type`=`{{
  config.type }}`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_policies_policy_id`: GET `/api/v1/policies/{{ config.policy_id }}` - records at response
  root; emits passthrough records.
- `api_v1_policies_policy_id_app`: GET `/api/v1/policies/{{ config.policy_id }}/app` - records path
  `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_policies_policy_id_mappings`: GET `/api/v1/policies/{{ config.policy_id }}/mappings` -
  records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `api_v1_policies_policy_id_mappings_mapping_id`: GET `/api/v1/policies/{{ config.policy_id
  }}/mappings/{{ config.mapping_id }}` - records at response root; emits passthrough records.
- `api_v1_policies_policy_id_rules`: GET `/api/v1/policies/{{ config.policy_id }}/rules` - records
  path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough
  records.
- `api_v1_policies_policy_id_rules_rule_id`: GET `/api/v1/policies/{{ config.policy_id }}/rules/{{
  config.rule_id }}` - records at response root; emits passthrough records.
- `api_v1_principal_rate_limits`: GET `/api/v1/principal-rate-limits` - records path `.`; query
  `filter`=`{{ config.filter }}`; `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `api_v1_principal_rate_limits_principal_rate_limit_id`: GET `/api/v1/principal-rate-limits/{{
  config.principal_rate_limit_id }}` - records at response root; emits passthrough records.
- `api_v1_push_providers`: GET `/api/v1/push-providers` - records path `.`; query `limit`=`200`;
  follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_push_providers_push_provider_id`: GET `/api/v1/push-providers/{{ config.push_provider_id
  }}` - records at response root; emits passthrough records.
- `api_v1_rate_limit_settings_admin_notifications`: GET
  `/api/v1/rate-limit-settings/admin-notifications` - records at response root; emits passthrough
  records.
- `api_v1_rate_limit_settings_per_client`: GET `/api/v1/rate-limit-settings/per-client` - records at
  response root; emits passthrough records.
- `api_v1_rate_limit_settings_warning_threshold`: GET
  `/api/v1/rate-limit-settings/warning-threshold` - records at response root; emits passthrough
  records.
- `api_v1_realm_assignments`: GET `/api/v1/realm-assignments` - records path `.`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_realm_assignments_operations`: GET `/api/v1/realm-assignments/operations` - records path
  `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_realm_assignments_assignment_id`: GET `/api/v1/realm-assignments/{{ config.assignment_id
  }}` - records at response root; emits passthrough records.
- `api_v1_realms`: GET `/api/v1/realms` - records path `.`; query `limit`=`200`; follows RFC 5988
  Link headers with rel=next; emits passthrough records.
- `api_v1_realms_realm_id`: GET `/api/v1/realms/{{ config.realm_id }}` - records at response root;
  emits passthrough records.
- `api_v1_roles_role_ref_subscriptions`: GET `/api/v1/roles/{{ config.role_ref }}/subscriptions` -
  records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `api_v1_roles_role_ref_subscriptions_notification_type`: GET `/api/v1/roles/{{ config.role_ref
  }}/subscriptions/{{ config.notification_type }}` - records at response root; emits passthrough
  records.
- `api_v1_security_events_providers`: GET `/api/v1/security-events-providers` - records path `.`;
  query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_security_events_providers_security_event_provider_id`: GET
  `/api/v1/security-events-providers/{{ config.security_event_provider_id }}` - records at response
  root; emits passthrough records.
- `api_v1_sessions_session_id`: GET `/api/v1/sessions/{{ config.session_id }}` - records at response
  root; emits passthrough records.
- `api_v1_ssf_stream`: GET `/api/v1/ssf/stream` - records path `.`; emits passthrough records.
- `api_v1_ssf_stream_status`: GET `/api/v1/ssf/stream/status` - records at response root; query
  `stream_id`=`{{ config.stream_id }}`; emits passthrough records.
- `api_v1_telephony_providers`: GET `/api/v1/telephony-providers` - records path `.`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_telephony_providers_custom_telephony_provider_id`: GET `/api/v1/telephony-providers/{{
  config.custom_telephony_provider_id }}` - records at response root; emits passthrough records.
- `api_v1_templates_sms`: GET `/api/v1/templates/sms` - records path `.`; query `limit`=`200`;
  follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_templates_sms_template_id`: GET `/api/v1/templates/sms/{{ config.template_id }}` - records
  at response root; emits passthrough records.
- `api_v1_threats_configuration`: GET `/api/v1/threats/configuration` - records at response root;
  emits passthrough records.
- `api_v1_trusted_origins`: GET `/api/v1/trustedOrigins` - records path `.`; query `limit`=`200`;
  follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_trusted_origins_trusted_origin_id`: GET `/api/v1/trustedOrigins/{{
  config.trusted_origin_id }}` - records path `scopes`; emits passthrough records.
- `api_v1_users_id`: GET `/api/v1/users/{{ config.id }}` - records at response root; emits
  passthrough records.
- `api_v1_users_id_app_links`: GET `/api/v1/users/{{ config.id }}/appLinks` - records path `.`;
  query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_users_id_blocks`: GET `/api/v1/users/{{ config.id }}/blocks` - records path `.`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_users_id_groups`: GET `/api/v1/users/{{ config.id }}/groups` - records path `.`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_users_id_idps`: GET `/api/v1/users/{{ config.id }}/idps` - records path `.`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_users_user_id_or_login_linked_objects_relationship_name`: GET `/api/v1/users/{{
  config.user_id_or_login }}/linkedObjects/{{ config.relationship_name }}` - records path `.`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_users_user_id_authenticator_enrollments`: GET `/api/v1/users/{{ config.user_id
  }}/authenticator-enrollments` - records at response root; query `limit`=`200`; follows RFC 5988
  Link headers with rel=next; emits passthrough records.
- `api_v1_users_user_id_authenticator_enrollments_enrollment_id`: GET `/api/v1/users/{{
  config.user_id }}/authenticator-enrollments/{{ config.enrollment_id }}` - records at response
  root; emits passthrough records.
- `api_v1_users_user_id_classification`: GET `/api/v1/users/{{ config.user_id }}/classification` -
  records at response root; emits passthrough records.
- `api_v1_users_user_id_clients`: GET `/api/v1/users/{{ config.user_id }}/clients` - records path
  `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_users_user_id_clients_client_id_grants`: GET `/api/v1/users/{{ config.user_id
  }}/clients/{{ config.client_id }}/grants` - records path `.`; query `limit`=`200`; follows RFC
  5988 Link headers with rel=next; emits passthrough records.
- `api_v1_users_user_id_clients_client_id_tokens`: GET `/api/v1/users/{{ config.user_id
  }}/clients/{{ config.client_id }}/tokens` - records path `.`; query `limit`=`200`; follows RFC
  5988 Link headers with rel=next; emits passthrough records.
- `api_v1_users_user_id_clients_client_id_tokens_token_id`: GET `/api/v1/users/{{ config.user_id
  }}/clients/{{ config.client_id }}/tokens/{{ config.token_id }}` - records path `.`; emits
  passthrough records.
- `api_v1_users_user_id_devices`: GET `/api/v1/users/{{ config.user_id }}/devices` - records path
  `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_users_user_id_factors`: GET `/api/v1/users/{{ config.user_id }}/factors` - records path
  `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_users_user_id_factors_catalog`: GET `/api/v1/users/{{ config.user_id }}/factors/catalog` -
  records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `api_v1_users_user_id_factors_questions`: GET `/api/v1/users/{{ config.user_id
  }}/factors/questions` - records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with
  rel=next; emits passthrough records.
- `api_v1_users_user_id_factors_factor_id`: GET `/api/v1/users/{{ config.user_id }}/factors/{{
  config.factor_id }}` - records at response root; emits passthrough records.
- `api_v1_users_user_id_factors_factor_id_transactions_transaction_id`: GET `/api/v1/users/{{
  config.user_id }}/factors/{{ config.factor_id }}/transactions/{{ config.transaction_id }}` -
  records at response root; emits passthrough records.
- `api_v1_users_user_id_grants`: GET `/api/v1/users/{{ config.user_id }}/grants` - records path `.`;
  query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_users_user_id_grants_grant_id`: GET `/api/v1/users/{{ config.user_id }}/grants/{{
  config.grant_id }}` - records at response root; emits passthrough records.
- `api_v1_users_user_id_risk`: GET `/api/v1/users/{{ config.user_id }}/risk` - records at response
  root; emits passthrough records.
- `api_v1_users_user_id_roles`: GET `/api/v1/users/{{ config.user_id }}/roles` - records path `.`;
  query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_users_user_id_roles_role_assignment_id`: GET `/api/v1/users/{{ config.user_id }}/roles/{{
  config.role_assignment_id }}` - records at response root; emits passthrough records.
- `api_v1_users_user_id_roles_role_assignment_id_governance`: GET `/api/v1/users/{{ config.user_id
  }}/roles/{{ config.role_assignment_id }}/governance` - records at response root; emits passthrough
  records.
- `api_v1_users_user_id_roles_role_assignment_id_governance_grant_id`: GET `/api/v1/users/{{
  config.user_id }}/roles/{{ config.role_assignment_id }}/governance/{{ config.grant_id }}` -
  records at response root; emits passthrough records.
- `api_v1_users_user_id_roles_role_assignment_id_governance_grant_id_resources`: GET
  `/api/v1/users/{{ config.user_id }}/roles/{{ config.role_assignment_id }}/governance/{{
  config.grant_id }}/resources` - records at response root; emits passthrough records.
- `api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps`: GET `/api/v1/users/{{
  config.user_id }}/roles/{{ config.role_assignment_id }}/targets/catalog/apps` - records path `.`;
  query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_users_user_id_roles_role_assignment_id_targets_groups`: GET `/api/v1/users/{{
  config.user_id }}/roles/{{ config.role_assignment_id }}/targets/groups` - records path `.`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_users_user_id_roles_role_id_or_encoded_role_id_targets`: GET `/api/v1/users/{{
  config.user_id }}/roles/{{ config.role_id_or_encoded_role_id }}/targets` - records path `.`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `api_v1_users_user_id_subscriptions`: GET `/api/v1/users/{{ config.user_id }}/subscriptions` -
  records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `api_v1_users_user_id_subscriptions_notification_type`: GET `/api/v1/users/{{ config.user_id
  }}/subscriptions/{{ config.notification_type }}` - records at response root; emits passthrough
  records.
- `api_v1_zones`: GET `/api/v1/zones` - records path `.`; query `limit`=`200`; follows RFC 5988 Link
  headers with rel=next; emits passthrough records.
- `api_v1_zones_zone_id`: GET `/api/v1/zones/{{ config.zone_id }}` - records at response root; emits
  passthrough records.
- `attack_protection_api_v1_authenticator_settings`: GET
  `/attack-protection/api/v1/authenticator-settings` - records at response root; emits passthrough
  records.
- `attack_protection_api_v1_user_lockout_settings`: GET
  `/attack-protection/api/v1/user-lockout-settings` - records at response root; emits passthrough
  records.
- `integrations_api_v1_api_services`: GET `/integrations/api/v1/api-services` - records path `.`;
  query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `integrations_api_v1_api_services_api_service_id`: GET `/integrations/api/v1/api-services/{{
  config.api_service_id }}` - records at response root; emits passthrough records.
- `integrations_api_v1_api_services_api_service_id_credentials_secrets`: GET
  `/integrations/api/v1/api-services/{{ config.api_service_id }}/credentials/secrets` - records path
  `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `oauth2_v1_clients_client_id_roles`: GET `/oauth2/v1/clients/{{ config.client_id }}/roles` -
  records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `oauth2_v1_clients_client_id_roles_role_assignment_id`: GET `/oauth2/v1/clients/{{
  config.client_id }}/roles/{{ config.role_assignment_id }}` - records at response root; emits
  passthrough records.
- `oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps`: GET
  `/oauth2/v1/clients/{{ config.client_id }}/roles/{{ config.role_assignment_id
  }}/targets/catalog/apps` - records path `.`; query `limit`=`200`; follows RFC 5988 Link headers
  with rel=next; emits passthrough records.
- `oauth2_v1_clients_client_id_roles_role_assignment_id_targets_groups`: GET `/oauth2/v1/clients/{{
  config.client_id }}/roles/{{ config.role_assignment_id }}/targets/groups` - records path `.`;
  query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `okta_personal_settings_api_v1_export_blocklists`: GET
  `/okta-personal-settings/api/v1/export-blocklists` - records at response root; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.
- `privileged_access_api_v1_okta_service_accounts`: GET
  `/privileged-access/api/v1/okta-service-accounts` - records path `.`; query `limit`=`200`; follows
  RFC 5988 Link headers with rel=next; emits passthrough records.
- `privileged_access_api_v1_okta_service_accounts_id`: GET
  `/privileged-access/api/v1/okta-service-accounts/{{ config.id }}` - records at response root;
  emits passthrough records.
- `privileged_access_api_v1_service_accounts`: GET `/privileged-access/api/v1/service-accounts` -
  records path `.`; query `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits
  passthrough records.
- `privileged_access_api_v1_service_accounts_id`: GET `/privileged-access/api/v1/service-accounts/{{
  config.id }}` - records at response root; emits passthrough records.
- `webauthn_registration_api_v1_users_user_id_enrollments`: GET
  `/webauthn-registration/api/v1/users/{{ config.user_id }}/enrollments` - records path `.`; query
  `limit`=`200`; follows RFC 5988 Link headers with rel=next; emits passthrough records.

## Write actions & risks

Overall write risk: external Okta admin API mutations including lifecycle, provisioning, credential,
policy, app, user, group, and delete operations.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_api_v1_agent_pools_pool_id_updates`: POST `/api/v1/agentPools/{{ record.pool_id
  }}/updates` - kind `create`; body type `json`; path fields `pool_id`; required record fields
  `pool_id`; accepted fields `_links`, `agentType`, `agents`, `enabled`, `id`, `name`,
  `notifyAdmin`, `pool_id`, `reason`, `schedule`, `sortOrder`, `status`, `targetVersion`; risk:
  medium: external Okta admin API mutation; approval required.
- `create_api_v1_agent_pools_pool_id_updates_settings`: POST `/api/v1/agentPools/{{ record.pool_id
  }}/updates/settings` - kind `create`; body type `json`; path fields `pool_id`; required record
  fields `pool_id`, `agentType`; accepted fields `agentType`, `continueOnError`, `latestVersion`,
  `minimalSupportedVersion`, `poolId`, `poolName`, `pool_id`, `releaseChannel`; risk: medium:
  external Okta admin API mutation; approval required.
- `create_api_v1_agent_pools_pool_id_updates_update_id`: POST `/api/v1/agentPools/{{ record.pool_id
  }}/updates/{{ record.update_id }}` - kind `create`; body type `json`; path fields `pool_id`,
  `update_id`; required record fields `pool_id`, `update_id`; accepted fields `_links`, `agentType`,
  `agents`, `enabled`, `id`, `name`, `notifyAdmin`, `pool_id`, `reason`, `schedule`, `sortOrder`,
  `status`, `targetVersion`, `update_id`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_api_v1_agent_pools_pool_id_updates_update_id`: DELETE `/api/v1/agentPools/{{
  record.pool_id }}/updates/{{ record.update_id }}` - kind `delete`; body type `none`; path fields
  `pool_id`, `update_id`; required record fields `pool_id`, `update_id`; accepted fields `pool_id`,
  `update_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_agent_pools_pool_id_updates_update_id_activate`: POST `/api/v1/agentPools/{{
  record.pool_id }}/updates/{{ record.update_id }}/activate` - kind `custom`; body type `none`; path
  fields `pool_id`, `update_id`; required record fields `pool_id`, `update_id`; accepted fields
  `pool_id`, `update_id`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_agent_pools_pool_id_updates_update_id_deactivate`: POST `/api/v1/agentPools/{{
  record.pool_id }}/updates/{{ record.update_id }}/deactivate` - kind `custom`; body type `none`;
  path fields `pool_id`, `update_id`; required record fields `pool_id`, `update_id`; accepted fields
  `pool_id`, `update_id`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_agent_pools_pool_id_updates_update_id_pause`: POST `/api/v1/agentPools/{{
  record.pool_id }}/updates/{{ record.update_id }}/pause` - kind `custom`; body type `none`; path
  fields `pool_id`, `update_id`; required record fields `pool_id`, `update_id`; accepted fields
  `pool_id`, `update_id`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_agent_pools_pool_id_updates_update_id_resume`: POST `/api/v1/agentPools/{{
  record.pool_id }}/updates/{{ record.update_id }}/resume` - kind `custom`; body type `none`; path
  fields `pool_id`, `update_id`; required record fields `pool_id`, `update_id`; accepted fields
  `pool_id`, `update_id`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_agent_pools_pool_id_updates_update_id_retry`: POST `/api/v1/agentPools/{{
  record.pool_id }}/updates/{{ record.update_id }}/retry` - kind `custom`; body type `none`; path
  fields `pool_id`, `update_id`; required record fields `pool_id`, `update_id`; accepted fields
  `pool_id`, `update_id`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_agent_pools_pool_id_updates_update_id_stop`: POST `/api/v1/agentPools/{{
  record.pool_id }}/updates/{{ record.update_id }}/stop` - kind `custom`; body type `none`; path
  fields `pool_id`, `update_id`; required record fields `pool_id`, `update_id`; accepted fields
  `pool_id`, `update_id`; risk: high: external Okta admin API mutation; approval required.
- `delete_api_v1_api_tokens_current`: DELETE `/api/v1/api-tokens/current` - kind `delete`; body type
  `none`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  high: external Okta admin API mutation; approval required.
- `update_api_v1_api_tokens_api_token_id`: PUT `/api/v1/api-tokens/{{ record.api_token_id }}` - kind
  `update`; body type `json`; path fields `api_token_id`; required record fields `api_token_id`;
  accepted fields `api_token_id`, `clientName`, `created`, `name`, `network`, `userId`; risk:
  medium: external Okta admin API mutation; approval required.
- `delete_api_v1_api_tokens_api_token_id`: DELETE `/api/v1/api-tokens/{{ record.api_token_id }}` -
  kind `delete`; body type `none`; path fields `api_token_id`; required record fields
  `api_token_id`; accepted fields `api_token_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: high: external Okta admin API mutation; approval
  required.
- `create_api_v1_apps`: POST `/api/v1/apps` - kind `create`; body type `json`; required record
  fields `signOnMode`, `label`; accepted fields `_embedded`, `_links`, `accessibility`, `created`,
  `expressConfiguration`, `features`, `id`, `label`, `lastUpdated`, `licensing`, `orn`, `profile`,
  `signOnMode`, `status`, `universalLogout`, `visibility`; risk: medium: external Okta admin API
  mutation; approval required.
- `update_api_v1_apps_app_id`: PUT `/api/v1/apps/{{ record.app_id }}` - kind `update`; body type
  `json`; path fields `app_id`; required record fields `app_id`, `signOnMode`, `label`; accepted
  fields `_embedded`, `_links`, `accessibility`, `app_id`, `created`, `expressConfiguration`,
  `features`, `id`, `label`, `lastUpdated`, `licensing`, `orn`, `profile`, `signOnMode`, `status`,
  `universalLogout`, `visibility`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_api_v1_apps_app_id`: DELETE `/api/v1/apps/{{ record.app_id }}` - kind `delete`; body type
  `none`; path fields `app_id`; required record fields `app_id`; accepted fields `app_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: high: external Okta
  admin API mutation; approval required.
- `create_api_v1_apps_app_id_connections_default`: POST `/api/v1/apps/{{ record.app_id
  }}/connections/default` - kind `create`; body type `json`; path fields `app_id`; required record
  fields `app_id`, `profile`; accepted fields `app_id`, `baseUrl`, `profile`; risk: medium: external
  Okta admin API mutation; approval required.
- `execute_api_v1_apps_app_id_connections_default_lifecycle_activate`: POST `/api/v1/apps/{{
  record.app_id }}/connections/default/lifecycle/activate` - kind `custom`; body type `none`; path
  fields `app_id`; required record fields `app_id`; accepted fields `app_id`; risk: high: external
  Okta admin API mutation; approval required.
- `execute_api_v1_apps_app_id_connections_default_lifecycle_deactivate`: POST `/api/v1/apps/{{
  record.app_id }}/connections/default/lifecycle/deactivate` - kind `custom`; body type `none`; path
  fields `app_id`; required record fields `app_id`; accepted fields `app_id`; risk: high: external
  Okta admin API mutation; approval required.
- `create_api_v1_apps_app_id_credentials_csrs`: POST `/api/v1/apps/{{ record.app_id
  }}/credentials/csrs` - kind `create`; body type `json`; path fields `app_id`; required record
  fields `app_id`; accepted fields `app_id`, `subject`, `subjectAltNames`; risk: medium: external
  Okta admin API mutation; approval required.
- `delete_api_v1_apps_app_id_credentials_csrs_csr_id`: DELETE `/api/v1/apps/{{ record.app_id
  }}/credentials/csrs/{{ record.csr_id }}` - kind `delete`; body type `none`; path fields `app_id`,
  `csr_id`; required record fields `app_id`, `csr_id`; accepted fields `app_id`, `csr_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: high: external Okta
  admin API mutation; approval required.
- `create_api_v1_apps_app_id_credentials_jwks`: POST `/api/v1/apps/{{ record.app_id
  }}/credentials/jwks` - kind `create`; body type `json`; path fields `app_id`; required record
  fields `app_id`; accepted fields `app_id`, `kid`, `status`; risk: medium: external Okta admin API
  mutation; approval required.
- `delete_api_v1_apps_app_id_credentials_jwks_key_id`: DELETE `/api/v1/apps/{{ record.app_id
  }}/credentials/jwks/{{ record.key_id }}` - kind `delete`; body type `none`; path fields `app_id`,
  `key_id`; required record fields `app_id`, `key_id`; accepted fields `app_id`, `key_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: high: external Okta
  admin API mutation; approval required.
- `execute_api_v1_apps_app_id_credentials_jwks_key_id_lifecycle_activate`: POST `/api/v1/apps/{{
  record.app_id }}/credentials/jwks/{{ record.key_id }}/lifecycle/activate` - kind `custom`; body
  type `none`; path fields `app_id`, `key_id`; required record fields `app_id`, `key_id`; accepted
  fields `app_id`, `key_id`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_apps_app_id_credentials_jwks_key_id_lifecycle_deactivate`: POST `/api/v1/apps/{{
  record.app_id }}/credentials/jwks/{{ record.key_id }}/lifecycle/deactivate` - kind `custom`; body
  type `none`; path fields `app_id`, `key_id`; required record fields `app_id`, `key_id`; accepted
  fields `app_id`, `key_id`; risk: high: external Okta admin API mutation; approval required.
- `create_api_v1_apps_app_id_credentials_secrets`: POST `/api/v1/apps/{{ record.app_id
  }}/credentials/secrets` - kind `create`; body type `json`; path fields `app_id`; required record
  fields `app_id`; accepted fields `app_id`, `client_secret`, `status`; risk: medium: external Okta
  admin API mutation; approval required.
- `delete_api_v1_apps_app_id_credentials_secrets_secret_id`: DELETE `/api/v1/apps/{{ record.app_id
  }}/credentials/secrets/{{ record.secret_id }}` - kind `delete`; body type `none`; path fields
  `app_id`, `secret_id`; required record fields `app_id`, `secret_id`; accepted fields `app_id`,
  `secret_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_apps_app_id_credentials_secrets_secret_id_lifecycle_activate`: POST
  `/api/v1/apps/{{ record.app_id }}/credentials/secrets/{{ record.secret_id }}/lifecycle/activate` -
  kind `custom`; body type `none`; path fields `app_id`, `secret_id`; required record fields
  `app_id`, `secret_id`; accepted fields `app_id`, `secret_id`; risk: high: external Okta admin API
  mutation; approval required.
- `execute_api_v1_apps_app_id_credentials_secrets_secret_id_lifecycle_deactivate`: POST
  `/api/v1/apps/{{ record.app_id }}/credentials/secrets/{{ record.secret_id }}/lifecycle/deactivate`
  - kind `custom`; body type `none`; path fields `app_id`, `secret_id`; required record fields
  `app_id`, `secret_id`; accepted fields `app_id`, `secret_id`; risk: high: external Okta admin API
  mutation; approval required.
- `create_api_v1_apps_app_id_cwo_connections`: POST `/api/v1/apps/{{ record.app_id
  }}/cwo/connections` - kind `create`; body type `json`; path fields `app_id`; required record
  fields `app_id`; accepted fields `app_id`, `created`, `id`, `lastUpdated`,
  `requestingAppInstanceId`, `resourceAppInstanceId`, `status`; risk: medium: external Okta admin
  API mutation; approval required.
- `update_api_v1_apps_app_id_cwo_connections_connection_id`: PATCH `/api/v1/apps/{{ record.app_id
  }}/cwo/connections/{{ record.connection_id }}` - kind `update`; body type `json`; path fields
  `app_id`, `connection_id`; required record fields `app_id`, `connection_id`, `status`; accepted
  fields `app_id`, `connection_id`, `status`; risk: medium: external Okta admin API mutation;
  approval required.
- `delete_api_v1_apps_app_id_cwo_connections_connection_id`: DELETE `/api/v1/apps/{{ record.app_id
  }}/cwo/connections/{{ record.connection_id }}` - kind `delete`; body type `none`; path fields
  `app_id`, `connection_id`; required record fields `app_id`, `connection_id`; accepted fields
  `app_id`, `connection_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Okta admin API mutation; approval required.
- `update_api_v1_apps_app_id_features_feature_name`: PUT `/api/v1/apps/{{ record.app_id
  }}/features/{{ record.feature_name }}` - kind `update`; body type `json`; path fields `app_id`,
  `feature_name`; required record fields `app_id`, `feature_name`; accepted fields `app_id`,
  `create`, `feature_name`, `update`; risk: medium: external Okta admin API mutation; approval
  required.
- `create_api_v1_apps_app_id_federated_claims`: POST `/api/v1/apps/{{ record.app_id
  }}/federated-claims` - kind `create`; body type `json`; path fields `app_id`; required record
  fields `app_id`; accepted fields `app_id`, `expression`, `name`; risk: medium: external Okta admin
  API mutation; approval required.
- `update_api_v1_apps_app_id_federated_claims_claim_id`: PUT `/api/v1/apps/{{ record.app_id
  }}/federated-claims/{{ record.claim_id }}` - kind `update`; body type `json`; path fields
  `app_id`, `claim_id`; required record fields `app_id`, `claim_id`; accepted fields `app_id`,
  `claim_id`, `created`, `expression`, `id`, `lastUpdated`, `name`; risk: medium: external Okta
  admin API mutation; approval required.
- `delete_api_v1_apps_app_id_federated_claims_claim_id`: DELETE `/api/v1/apps/{{ record.app_id
  }}/federated-claims/{{ record.claim_id }}` - kind `delete`; body type `none`; path fields
  `app_id`, `claim_id`; required record fields `app_id`, `claim_id`; accepted fields `app_id`,
  `claim_id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  high: external Okta admin API mutation; approval required.
- `create_api_v1_apps_app_id_grants`: POST `/api/v1/apps/{{ record.app_id }}/grants` - kind
  `create`; body type `json`; path fields `app_id`; required record fields `app_id`, `issuer`,
  `scopeId`; accepted fields `_embedded`, `_links`, `app_id`, `clientId`, `created`, `createdBy`,
  `id`, `issuer`, `lastUpdated`, `scopeId`, `source`, `status`, `userId`; risk: medium: external
  Okta admin API mutation; approval required.
- `delete_api_v1_apps_app_id_grants_grant_id`: DELETE `/api/v1/apps/{{ record.app_id }}/grants/{{
  record.grant_id }}` - kind `delete`; body type `none`; path fields `app_id`, `grant_id`; required
  record fields `app_id`, `grant_id`; accepted fields `app_id`, `grant_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: high: external Okta admin API
  mutation; approval required.
- `create_api_v1_apps_app_id_group_push_mappings`: POST `/api/v1/apps/{{ record.app_id
  }}/group-push/mappings` - kind `create`; body type `json`; path fields `app_id`; required record
  fields `app_id`, `sourceGroupId`; accepted fields `appConfig`, `app_id`, `sourceGroupId`,
  `status`, `targetGroupId`, `targetGroupName`; risk: medium: external Okta admin API mutation;
  approval required.
- `update_api_v1_apps_app_id_group_push_mappings_mapping_id`: PATCH `/api/v1/apps/{{ record.app_id
  }}/group-push/mappings/{{ record.mapping_id }}` - kind `update`; body type `json`; path fields
  `app_id`, `mapping_id`; required record fields `app_id`, `mapping_id`, `status`; accepted fields
  `app_id`, `mapping_id`, `status`; risk: medium: external Okta admin API mutation; approval
  required.
- `update_api_v1_apps_app_id_groups_group_id`: PUT `/api/v1/apps/{{ record.app_id }}/groups/{{
  record.group_id }}` - kind `update`; body type `json`; path fields `app_id`, `group_id`; required
  record fields `app_id`, `group_id`; accepted fields `_embedded`, `_links`, `app_id`, `group_id`,
  `id`, `lastUpdated`, `priority`, `profile`; risk: medium: external Okta admin API mutation;
  approval required.
- `delete_api_v1_apps_app_id_groups_group_id`: DELETE `/api/v1/apps/{{ record.app_id }}/groups/{{
  record.group_id }}` - kind `delete`; body type `none`; path fields `app_id`, `group_id`; required
  record fields `app_id`, `group_id`; accepted fields `app_id`, `group_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: high: external Okta admin API
  mutation; approval required.
- `create_api_v1_apps_app_id_interclient_allowed_apps`: POST `/api/v1/apps/{{ record.app_id
  }}/interclient-allowed-apps` - kind `create`; body type `json`; path fields `app_id`; required
  record fields `app_id`; accepted fields `app_id`, `id`; risk: medium: external Okta admin API
  mutation; approval required.
- `delete_api_v1_apps_app_id_interclient_allowed_apps_allowed_app_id`: DELETE `/api/v1/apps/{{
  record.app_id }}/interclient-allowed-apps/{{ record.allowed_app_id }}` - kind `delete`; body type
  `none`; path fields `app_id`, `allowed_app_id`; required record fields `app_id`, `allowed_app_id`;
  accepted fields `allowed_app_id`, `app_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_apps_app_id_lifecycle_activate`: POST `/api/v1/apps/{{ record.app_id
  }}/lifecycle/activate` - kind `custom`; body type `none`; path fields `app_id`; required record
  fields `app_id`; accepted fields `app_id`; risk: high: external Okta admin API mutation; approval
  required.
- `execute_api_v1_apps_app_id_lifecycle_deactivate`: POST `/api/v1/apps/{{ record.app_id
  }}/lifecycle/deactivate` - kind `custom`; body type `none`; path fields `app_id`; required record
  fields `app_id`; accepted fields `app_id`; risk: high: external Okta admin API mutation; approval
  required.
- `update_api_v1_apps_app_id_policies_policy_id`: PUT `/api/v1/apps/{{ record.app_id }}/policies/{{
  record.policy_id }}` - kind `update`; body type `none`; path fields `app_id`, `policy_id`;
  required record fields `app_id`, `policy_id`; accepted fields `app_id`, `policy_id`; risk: medium:
  external Okta admin API mutation; approval required.
- `delete_api_v1_apps_app_id_tokens`: DELETE `/api/v1/apps/{{ record.app_id }}/tokens` - kind
  `delete`; body type `none`; path fields `app_id`; required record fields `app_id`; accepted fields
  `app_id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  high: external Okta admin API mutation; approval required.
- `delete_api_v1_apps_app_id_tokens_token_id`: DELETE `/api/v1/apps/{{ record.app_id }}/tokens/{{
  record.token_id }}` - kind `delete`; body type `none`; path fields `app_id`, `token_id`; required
  record fields `app_id`, `token_id`; accepted fields `app_id`, `token_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: high: external Okta admin API
  mutation; approval required.
- `create_api_v1_apps_app_id_users`: POST `/api/v1/apps/{{ record.app_id }}/users` - kind `create`;
  body type `json`; path fields `app_id`; required record fields `app_id`, `id`; accepted fields
  `_embedded`, `_links`, `app_id`, `created`, `credentials`, `externalId`, `id`, `lastSync`,
  `lastUpdated`, `passwordChanged`, `profile`, `scope`, `status`, `statusChanged`, `syncState`;
  risk: medium: external Okta admin API mutation; approval required.
- `create_api_v1_apps_app_id_users_user_id`: POST `/api/v1/apps/{{ record.app_id }}/users/{{
  record.user_id }}` - kind `create`; body type `json`; path fields `app_id`, `user_id`; required
  record fields `app_id`, `user_id`; accepted fields `app_id`, `credentials`, `user_id`; risk:
  medium: external Okta admin API mutation; approval required.
- `delete_api_v1_apps_app_id_users_user_id`: DELETE `/api/v1/apps/{{ record.app_id }}/users/{{
  record.user_id }}` - kind `delete`; body type `none`; path fields `app_id`, `user_id`; required
  record fields `app_id`, `user_id`; accepted fields `app_id`, `user_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: high: external Okta admin API
  mutation; approval required.
- `create_api_v1_apps_app_name_app_id_oauth2_callback`: POST `/api/v1/apps/{{ record.app_name }}/{{
  record.app_id }}/oauth2/callback` - kind `create`; body type `none`; path fields `app_name`,
  `app_id`; required record fields `app_name`, `app_id`; accepted fields `app_id`, `app_name`; risk:
  medium: external Okta admin API mutation; approval required.
- `create_api_v1_authenticators`: POST `/api/v1/authenticators` - kind `create`; body type `none`;
  risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_authenticators_authenticator_id`: PUT `/api/v1/authenticators/{{
  record.authenticator_id }}` - kind `update`; body type `none`; path fields `authenticator_id`;
  required record fields `authenticator_id`; accepted fields `authenticator_id`; risk: medium:
  external Okta admin API mutation; approval required.
- `create_api_v1_authenticators_authenticator_id_aaguids`: POST `/api/v1/authenticators/{{
  record.authenticator_id }}/aaguids` - kind `create`; body type `json`; path fields
  `authenticator_id`; required record fields `authenticator_id`; accepted fields `aaguid`,
  `attestationRootCertificates`, `authenticatorCharacteristics`, `authenticator_id`; risk: medium:
  external Okta admin API mutation; approval required.
- `update_api_v1_authenticators_authenticator_id_aaguids_aaguid`: PUT `/api/v1/authenticators/{{
  record.authenticator_id }}/aaguids/{{ record.aaguid }}` - kind `update`; body type `json`; path
  fields `authenticator_id`, `aaguid`; required record fields `authenticator_id`, `aaguid`; accepted
  fields `aaguid`, `attestationRootCertificates`, `authenticatorCharacteristics`,
  `authenticator_id`, `name`; risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_authenticators_authenticator_id_aaguids_aaguid_2`: PATCH `/api/v1/authenticators/{{
  record.authenticator_id }}/aaguids/{{ record.aaguid }}` - kind `update`; body type `json`; path
  fields `authenticator_id`, `aaguid`; required record fields `authenticator_id`, `aaguid`; accepted
  fields `aaguid`, `attestationRootCertificates`, `authenticatorCharacteristics`,
  `authenticator_id`, `name`; risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_authenticators_authenticator_id_aaguids_aaguid`: DELETE `/api/v1/authenticators/{{
  record.authenticator_id }}/aaguids/{{ record.aaguid }}` - kind `delete`; body type `none`; path
  fields `authenticator_id`, `aaguid`; required record fields `authenticator_id`, `aaguid`; accepted
  fields `aaguid`, `authenticator_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_authenticators_authenticator_id_lifecycle_activate`: POST
  `/api/v1/authenticators/{{ record.authenticator_id }}/lifecycle/activate` - kind `custom`; body
  type `none`; path fields `authenticator_id`; required record fields `authenticator_id`; accepted
  fields `authenticator_id`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_authenticators_authenticator_id_lifecycle_deactivate`: POST
  `/api/v1/authenticators/{{ record.authenticator_id }}/lifecycle/deactivate` - kind `custom`; body
  type `none`; path fields `authenticator_id`; required record fields `authenticator_id`; accepted
  fields `authenticator_id`; risk: high: external Okta admin API mutation; approval required.
- `update_api_v1_authenticators_authenticator_id_methods_method_type`: PUT
  `/api/v1/authenticators/{{ record.authenticator_id }}/methods/{{ record.method_type }}` - kind
  `update`; body type `json`; path fields `authenticator_id`, `method_type`; required record fields
  `authenticator_id`, `method_type`; accepted fields `_links`, `authenticator_id`, `method_type`,
  `status`, `type`; risk: medium: external Okta admin API mutation; approval required.
- `execute_api_v1_authenticators_authenticator_id_methods_method_type_lifecycle_activate`: POST
  `/api/v1/authenticators/{{ record.authenticator_id }}/methods/{{ record.method_type
  }}/lifecycle/activate` - kind `custom`; body type `none`; path fields `authenticator_id`,
  `method_type`; required record fields `authenticator_id`, `method_type`; accepted fields
  `authenticator_id`, `method_type`; risk: high: external Okta admin API mutation; approval
  required.
- `execute_api_v1_authenticators_authenticator_id_methods_method_type_lifecycle_deactivate`: POST
  `/api/v1/authenticators/{{ record.authenticator_id }}/methods/{{ record.method_type
  }}/lifecycle/deactivate` - kind `custom`; body type `none`; path fields `authenticator_id`,
  `method_type`; required record fields `authenticator_id`, `method_type`; accepted fields
  `authenticator_id`, `method_type`; risk: high: external Okta admin API mutation; approval
  required.
- `create_api_v1_authenticators_authenticator_id_methods_web_authn_method_type_verify_rp_id_domain`:
  POST `/api/v1/authenticators/{{ record.authenticator_id }}/methods/{{ record.web_authn_method_type
  }}/verify-rp-id-domain` - kind `create`; body type `none`; path fields `authenticator_id`,
  `web_authn_method_type`; required record fields `authenticator_id`, `web_authn_method_type`;
  accepted fields `authenticator_id`, `web_authn_method_type`; risk: medium: external Okta admin API
  mutation; approval required.
- `create_api_v1_authorization_servers`: POST `/api/v1/authorizationServers` - kind `create`; body
  type `json`; accepted fields `_links`, `accessTokenEncryptedResponseAlgorithm`, `audiences`,
  `created`, `credentials`, `description`, `id`, `issuer`, `issuerMode`, `jwks`, `jwks_uri`,
  `lastUpdated`, `name`, `status`; risk: medium: external Okta admin API mutation; approval
  required.
- `update_api_v1_authorization_servers_auth_server_id`: PUT `/api/v1/authorizationServers/{{
  record.auth_server_id }}` - kind `update`; body type `json`; path fields `auth_server_id`;
  required record fields `auth_server_id`; accepted fields `_links`,
  `accessTokenEncryptedResponseAlgorithm`, `audiences`, `auth_server_id`, `created`, `credentials`,
  `description`, `id`, `issuer`, `issuerMode`, `jwks`, `jwks_uri`, `lastUpdated`, `name`, `status`;
  risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_authorization_servers_auth_server_id`: DELETE `/api/v1/authorizationServers/{{
  record.auth_server_id }}` - kind `delete`; body type `none`; path fields `auth_server_id`;
  required record fields `auth_server_id`; accepted fields `auth_server_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: high: external Okta admin API
  mutation; approval required.
- `create_api_v1_authorization_servers_auth_server_id_associated_servers`: POST
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/associatedServers` - kind `create`; body
  type `json`; path fields `auth_server_id`; required record fields `auth_server_id`; accepted
  fields `auth_server_id`, `trusted`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_api_v1_authorization_servers_auth_server_id_associated_servers_associated_server_id`:
  DELETE `/api/v1/authorizationServers/{{ record.auth_server_id }}/associatedServers/{{
  record.associated_server_id }}` - kind `delete`; body type `none`; path fields `auth_server_id`,
  `associated_server_id`; required record fields `auth_server_id`, `associated_server_id`; accepted
  fields `associated_server_id`, `auth_server_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: high: external Okta admin API mutation; approval
  required.
- `create_api_v1_authorization_servers_auth_server_id_claims`: POST `/api/v1/authorizationServers/{{
  record.auth_server_id }}/claims` - kind `create`; body type `json`; path fields `auth_server_id`;
  required record fields `auth_server_id`; accepted fields `_links`, `alwaysIncludeInToken`,
  `auth_server_id`, `claimType`, `conditions`, `group_filter_type`, `id`, `name`, `status`,
  `system`, `value`, `valueType`; risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_authorization_servers_auth_server_id_claims_claim_id`: PUT
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/claims/{{ record.claim_id }}` - kind
  `update`; body type `json`; path fields `auth_server_id`, `claim_id`; required record fields
  `auth_server_id`, `claim_id`; accepted fields `_links`, `alwaysIncludeInToken`, `auth_server_id`,
  `claimType`, `claim_id`, `conditions`, `group_filter_type`, `id`, `name`, `status`, `system`,
  `value`, `valueType`; risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_authorization_servers_auth_server_id_claims_claim_id`: DELETE
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/claims/{{ record.claim_id }}` - kind
  `delete`; body type `none`; path fields `auth_server_id`, `claim_id`; required record fields
  `auth_server_id`, `claim_id`; accepted fields `auth_server_id`, `claim_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Okta admin
  API mutation; approval required.
- `delete_api_v1_authorization_servers_auth_server_id_clients_client_id_tokens`: DELETE
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/clients/{{ record.client_id }}/tokens` -
  kind `delete`; body type `none`; path fields `auth_server_id`, `client_id`; required record fields
  `auth_server_id`, `client_id`; accepted fields `auth_server_id`, `client_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Okta admin
  API mutation; approval required.
- `delete_api_v1_authorization_servers_auth_server_id_clients_client_id_tokens_token_id`: DELETE
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/clients/{{ record.client_id }}/tokens/{{
  record.token_id }}` - kind `delete`; body type `none`; path fields `auth_server_id`, `client_id`,
  `token_id`; required record fields `auth_server_id`, `client_id`, `token_id`; accepted fields
  `auth_server_id`, `client_id`, `token_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_authorization_servers_auth_server_id_credentials_lifecycle_key_rotate`: POST
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/credentials/lifecycle/keyRotate` - kind
  `custom`; body type `json`; path fields `auth_server_id`; required record fields `auth_server_id`;
  accepted fields `auth_server_id`, `use`; risk: high: external Okta admin API mutation; approval
  required.
- `execute_api_v1_authorization_servers_auth_server_id_lifecycle_activate`: POST
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/lifecycle/activate` - kind `custom`;
  body type `none`; path fields `auth_server_id`; required record fields `auth_server_id`; accepted
  fields `auth_server_id`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_authorization_servers_auth_server_id_lifecycle_deactivate`: POST
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/lifecycle/deactivate` - kind `custom`;
  body type `none`; path fields `auth_server_id`; required record fields `auth_server_id`; accepted
  fields `auth_server_id`; risk: high: external Okta admin API mutation; approval required.
- `create_api_v1_authorization_servers_auth_server_id_policies`: POST
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/policies` - kind `create`; body type
  `json`; path fields `auth_server_id`; required record fields `auth_server_id`; accepted fields
  `_links`, `auth_server_id`, `conditions`, `created`, `description`, `id`, `lastUpdated`, `name`,
  `priority`, `status`, `system`, `type`; risk: medium: external Okta admin API mutation; approval
  required.
- `update_api_v1_authorization_servers_auth_server_id_policies_policy_id`: PUT
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id }}` - kind
  `update`; body type `json`; path fields `auth_server_id`, `policy_id`; required record fields
  `auth_server_id`, `policy_id`; accepted fields `_links`, `auth_server_id`, `conditions`,
  `created`, `description`, `id`, `lastUpdated`, `name`, `policy_id`, `priority`, `status`,
  `system`, `type`; risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_authorization_servers_auth_server_id_policies_policy_id`: DELETE
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id }}` - kind
  `delete`; body type `none`; path fields `auth_server_id`, `policy_id`; required record fields
  `auth_server_id`, `policy_id`; accepted fields `auth_server_id`, `policy_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Okta admin
  API mutation; approval required.
- `execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_lifecycle_activate`: POST
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id
  }}/lifecycle/activate` - kind `custom`; body type `none`; path fields `auth_server_id`,
  `policy_id`; required record fields `auth_server_id`, `policy_id`; accepted fields
  `auth_server_id`, `policy_id`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_lifecycle_deactivate`:
  POST `/api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id
  }}/lifecycle/deactivate` - kind `custom`; body type `none`; path fields `auth_server_id`,
  `policy_id`; required record fields `auth_server_id`, `policy_id`; accepted fields
  `auth_server_id`, `policy_id`; risk: high: external Okta admin API mutation; approval required.
- `create_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules`: POST
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id }}/rules` -
  kind `create`; body type `json`; path fields `auth_server_id`, `policy_id`; required record fields
  `auth_server_id`, `policy_id`, `name`, `conditions`, `type`; accepted fields `_links`, `actions`,
  `auth_server_id`, `conditions`, `created`, `id`, `lastUpdated`, `name`, `policy_id`, `priority`,
  `status`, `system`, `type`; risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id`: PUT
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id }}/rules/{{
  record.rule_id }}` - kind `update`; body type `json`; path fields `auth_server_id`, `policy_id`,
  `rule_id`; required record fields `auth_server_id`, `policy_id`, `rule_id`, `name`, `conditions`,
  `type`; accepted fields `_links`, `actions`, `auth_server_id`, `conditions`, `created`, `id`,
  `lastUpdated`, `name`, `policy_id`, `priority`, `rule_id`, `status`, `system`, `type`; risk:
  medium: external Okta admin API mutation; approval required.
- `delete_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id`: DELETE
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id }}/rules/{{
  record.rule_id }}` - kind `delete`; body type `none`; path fields `auth_server_id`, `policy_id`,
  `rule_id`; required record fields `auth_server_id`, `policy_id`, `rule_id`; accepted fields
  `auth_server_id`, `policy_id`, `rule_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id_lifecycle_activate`:
  POST `/api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id
  }}/rules/{{ record.rule_id }}/lifecycle/activate` - kind `custom`; body type `none`; path fields
  `auth_server_id`, `policy_id`, `rule_id`; required record fields `auth_server_id`, `policy_id`,
  `rule_id`; accepted fields `auth_server_id`, `policy_id`, `rule_id`; risk: high: external Okta
  admin API mutation; approval required.
- `execute_api_v1_authorization_servers_auth_server_id_policies_policy_id_rules_rule_id_lifecycle_deactivate`:
  POST `/api/v1/authorizationServers/{{ record.auth_server_id }}/policies/{{ record.policy_id
  }}/rules/{{ record.rule_id }}/lifecycle/deactivate` - kind `custom`; body type `none`; path fields
  `auth_server_id`, `policy_id`, `rule_id`; required record fields `auth_server_id`, `policy_id`,
  `rule_id`; accepted fields `auth_server_id`, `policy_id`, `rule_id`; risk: high: external Okta
  admin API mutation; approval required.
- `create_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys`: POST
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/resourceservercredentials/keys` - kind
  `create`; body type `json`; path fields `auth_server_id`; required record fields `auth_server_id`;
  accepted fields `auth_server_id`, `e`, `kid`, `kty`, `n`, `status`, `use`; risk: medium: external
  Okta admin API mutation; approval required.
- `delete_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id`: DELETE
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/resourceservercredentials/keys/{{
  record.key_id }}` - kind `delete`; body type `none`; path fields `auth_server_id`, `key_id`;
  required record fields `auth_server_id`, `key_id`; accepted fields `auth_server_id`, `key_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: high:
  external Okta admin API mutation; approval required.
- `execute_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id_lifecycle_activate`:
  POST `/api/v1/authorizationServers/{{ record.auth_server_id }}/resourceservercredentials/keys/{{
  record.key_id }}/lifecycle/activate` - kind `custom`; body type `none`; path fields
  `auth_server_id`, `key_id`; required record fields `auth_server_id`, `key_id`; accepted fields
  `auth_server_id`, `key_id`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_authorization_servers_auth_server_id_resourceservercredentials_keys_key_id_lifecycle_deactivate`:
  POST `/api/v1/authorizationServers/{{ record.auth_server_id }}/resourceservercredentials/keys/{{
  record.key_id }}/lifecycle/deactivate` - kind `custom`; body type `none`; path fields
  `auth_server_id`, `key_id`; required record fields `auth_server_id`, `key_id`; accepted fields
  `auth_server_id`, `key_id`; risk: high: external Okta admin API mutation; approval required.
- `create_api_v1_authorization_servers_auth_server_id_scopes`: POST `/api/v1/authorizationServers/{{
  record.auth_server_id }}/scopes` - kind `create`; body type `json`; path fields `auth_server_id`;
  required record fields `auth_server_id`, `name`; accepted fields `_links`, `auth_server_id`,
  `consent`, `default`, `description`, `displayName`, `id`, `metadataPublish`, `name`, `optional`,
  `system`; risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_authorization_servers_auth_server_id_scopes_scope_id`: PUT
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/scopes/{{ record.scope_id }}` - kind
  `update`; body type `json`; path fields `auth_server_id`, `scope_id`; required record fields
  `auth_server_id`, `scope_id`, `name`; accepted fields `_links`, `auth_server_id`, `consent`,
  `default`, `description`, `displayName`, `id`, `metadataPublish`, `name`, `optional`, `scope_id`,
  `system`; risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_authorization_servers_auth_server_id_scopes_scope_id`: DELETE
  `/api/v1/authorizationServers/{{ record.auth_server_id }}/scopes/{{ record.scope_id }}` - kind
  `delete`; body type `none`; path fields `auth_server_id`, `scope_id`; required record fields
  `auth_server_id`, `scope_id`; accepted fields `auth_server_id`, `scope_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Okta admin
  API mutation; approval required.
- `create_api_v1_behaviors`: POST `/api/v1/behaviors` - kind `create`; body type `json`; required
  record fields `name`, `type`; accepted fields `_link`, `created`, `id`, `lastUpdated`, `name`,
  `status`, `type`; risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_behaviors_behavior_id`: PUT `/api/v1/behaviors/{{ record.behavior_id }}` - kind
  `update`; body type `json`; path fields `behavior_id`; required record fields `behavior_id`,
  `name`, `type`; accepted fields `_link`, `behavior_id`, `created`, `id`, `lastUpdated`, `name`,
  `status`, `type`; risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_behaviors_behavior_id`: DELETE `/api/v1/behaviors/{{ record.behavior_id }}` - kind
  `delete`; body type `none`; path fields `behavior_id`; required record fields `behavior_id`;
  accepted fields `behavior_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_behaviors_behavior_id_lifecycle_activate`: POST `/api/v1/behaviors/{{
  record.behavior_id }}/lifecycle/activate` - kind `custom`; body type `none`; path fields
  `behavior_id`; required record fields `behavior_id`; accepted fields `behavior_id`; risk: high:
  external Okta admin API mutation; approval required.
- `execute_api_v1_behaviors_behavior_id_lifecycle_deactivate`: POST `/api/v1/behaviors/{{
  record.behavior_id }}/lifecycle/deactivate` - kind `custom`; body type `none`; path fields
  `behavior_id`; required record fields `behavior_id`; accepted fields `behavior_id`; risk: high:
  external Okta admin API mutation; approval required.
- `create_api_v1_bot_protection_configuration`: POST `/api/v1/bot-protection/configuration` - kind
  `create`; body type `json`; required record fields `level`, `mode`; accepted fields `_links`,
  `enforcementType`, `level`, `mode`, `supportedFlows`; risk: medium: external Okta admin API
  mutation; approval required.
- `create_api_v1_brands`: POST `/api/v1/brands` - kind `create`; body type `json`; required record
  fields `name`; accepted fields `name`; risk: medium: external Okta admin API mutation; approval
  required.
- `update_api_v1_brands_brand_id`: PUT `/api/v1/brands/{{ record.brand_id }}` - kind `update`; body
  type `json`; path fields `brand_id`; required record fields `brand_id`, `name`; accepted fields
  `agreeToCustomPrivacyPolicy`, `brand_id`, `customPrivacyPolicyUrl`, `defaultApp`, `emailDomainId`,
  `locale`, `name`, `removePoweredByOkta`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_api_v1_brands_brand_id`: DELETE `/api/v1/brands/{{ record.brand_id }}` - kind `delete`;
  body type `none`; path fields `brand_id`; required record fields `brand_id`; accepted fields
  `brand_id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  high: external Okta admin API mutation; approval required.
- `update_api_v1_brands_brand_id_pages_error_customized`: PUT `/api/v1/brands/{{ record.brand_id
  }}/pages/error/customized` - kind `update`; body type `json`; path fields `brand_id`; required
  record fields `brand_id`; accepted fields `brand_id`, `contentSecurityPolicySetting`; risk:
  medium: external Okta admin API mutation; approval required.
- `delete_api_v1_brands_brand_id_pages_error_customized`: DELETE `/api/v1/brands/{{ record.brand_id
  }}/pages/error/customized` - kind `delete`; body type `none`; path fields `brand_id`; required
  record fields `brand_id`; accepted fields `brand_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: high: external Okta admin API mutation; approval
  required.
- `update_api_v1_brands_brand_id_pages_error_preview`: PUT `/api/v1/brands/{{ record.brand_id
  }}/pages/error/preview` - kind `update`; body type `json`; path fields `brand_id`; required record
  fields `brand_id`; accepted fields `brand_id`, `contentSecurityPolicySetting`; risk: medium:
  external Okta admin API mutation; approval required.
- `delete_api_v1_brands_brand_id_pages_error_preview`: DELETE `/api/v1/brands/{{ record.brand_id
  }}/pages/error/preview` - kind `delete`; body type `none`; path fields `brand_id`; required record
  fields `brand_id`; accepted fields `brand_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: high: external Okta admin API mutation; approval
  required.
- `update_api_v1_brands_brand_id_pages_sign_in_customized`: PUT `/api/v1/brands/{{ record.brand_id
  }}/pages/sign-in/customized` - kind `update`; body type `json`; path fields `brand_id`; required
  record fields `brand_id`; accepted fields `brand_id`, `contentSecurityPolicySetting`,
  `widgetCustomizations`, `widgetVersion`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_api_v1_brands_brand_id_pages_sign_in_customized`: DELETE `/api/v1/brands/{{
  record.brand_id }}/pages/sign-in/customized` - kind `delete`; body type `none`; path fields
  `brand_id`; required record fields `brand_id`; accepted fields `brand_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: high: external Okta admin API
  mutation; approval required.
- `update_api_v1_brands_brand_id_pages_sign_in_preview`: PUT `/api/v1/brands/{{ record.brand_id
  }}/pages/sign-in/preview` - kind `update`; body type `json`; path fields `brand_id`; required
  record fields `brand_id`; accepted fields `brand_id`, `contentSecurityPolicySetting`,
  `widgetCustomizations`, `widgetVersion`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_api_v1_brands_brand_id_pages_sign_in_preview`: DELETE `/api/v1/brands/{{ record.brand_id
  }}/pages/sign-in/preview` - kind `delete`; body type `none`; path fields `brand_id`; required
  record fields `brand_id`; accepted fields `brand_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: high: external Okta admin API mutation; approval
  required.
- `update_api_v1_brands_brand_id_pages_sign_out_customized`: PUT `/api/v1/brands/{{ record.brand_id
  }}/pages/sign-out/customized` - kind `update`; body type `json`; path fields `brand_id`; required
  record fields `brand_id`, `type`; accepted fields `brand_id`, `type`, `url`; risk: medium:
  external Okta admin API mutation; approval required.
- `create_api_v1_brands_brand_id_templates_email_template_name_customizations`: POST
  `/api/v1/brands/{{ record.brand_id }}/templates/email/{{ record.template_name }}/customizations` -
  kind `create`; body type `json`; path fields `brand_id`, `template_name`; required record fields
  `brand_id`, `template_name`, `language`; accepted fields `_links`, `brand_id`, `created`, `id`,
  `isDefault`, `language`, `lastUpdated`, `template_name`; risk: medium: external Okta admin API
  mutation; approval required.
- `delete_api_v1_brands_brand_id_templates_email_template_name_customizations`: DELETE
  `/api/v1/brands/{{ record.brand_id }}/templates/email/{{ record.template_name }}/customizations` -
  kind `delete`; body type `none`; path fields `brand_id`, `template_name`; required record fields
  `brand_id`, `template_name`; accepted fields `brand_id`, `template_name`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: high: external Okta admin API
  mutation; approval required.
- `update_api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id`: PUT
  `/api/v1/brands/{{ record.brand_id }}/templates/email/{{ record.template_name }}/customizations/{{
  record.customization_id }}` - kind `update`; body type `json`; path fields `brand_id`,
  `template_name`, `customization_id`; required record fields `brand_id`, `template_name`,
  `customization_id`, `language`; accepted fields `_links`, `brand_id`, `created`,
  `customization_id`, `id`, `isDefault`, `language`, `lastUpdated`, `template_name`; risk: medium:
  external Okta admin API mutation; approval required.
- `delete_api_v1_brands_brand_id_templates_email_template_name_customizations_customization_id`:
  DELETE `/api/v1/brands/{{ record.brand_id }}/templates/email/{{ record.template_name
  }}/customizations/{{ record.customization_id }}` - kind `delete`; body type `none`; path fields
  `brand_id`, `template_name`, `customization_id`; required record fields `brand_id`,
  `template_name`, `customization_id`; accepted fields `brand_id`, `customization_id`,
  `template_name`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: high: external Okta admin API mutation; approval required.
- `update_api_v1_brands_brand_id_templates_email_template_name_settings`: PUT `/api/v1/brands/{{
  record.brand_id }}/templates/email/{{ record.template_name }}/settings` - kind `update`; body type
  `json`; path fields `brand_id`, `template_name`; required record fields `brand_id`,
  `template_name`, `recipients`; accepted fields `brand_id`, `recipients`, `template_name`; risk:
  medium: external Okta admin API mutation; approval required.
- `create_api_v1_brands_brand_id_templates_email_template_name_test`: POST `/api/v1/brands/{{
  record.brand_id }}/templates/email/{{ record.template_name }}/test` - kind `create`; body type
  `none`; path fields `brand_id`, `template_name`; required record fields `brand_id`,
  `template_name`; accepted fields `brand_id`, `template_name`; risk: medium: external Okta admin
  API mutation; approval required.
- `update_api_v1_brands_brand_id_themes_theme_id`: PUT `/api/v1/brands/{{ record.brand_id
  }}/themes/{{ record.theme_id }}` - kind `update`; body type `json`; path fields `brand_id`,
  `theme_id`; required record fields `brand_id`, `theme_id`, `primaryColorHex`, `secondaryColorHex`,
  `signInPageTouchPointVariant`, `endUserDashboardTouchPointVariant`, `errorPageTouchPointVariant`,
  `emailTemplateTouchPointVariant`; accepted fields `_links`, `brand_id`,
  `emailTemplateTouchPointVariant`, `endUserDashboardTouchPointVariant`,
  `errorPageTouchPointVariant`, `loadingPageTouchPointVariant`, `primaryColorContrastHex`,
  `primaryColorHex`, `secondaryColorContrastHex`, `secondaryColorHex`,
  `signInPageTouchPointVariant`, `theme_id`; risk: medium: external Okta admin API mutation;
  approval required.
- `delete_api_v1_brands_brand_id_themes_theme_id_background_image`: DELETE `/api/v1/brands/{{
  record.brand_id }}/themes/{{ record.theme_id }}/background-image` - kind `delete`; body type
  `none`; path fields `brand_id`, `theme_id`; required record fields `brand_id`, `theme_id`;
  accepted fields `brand_id`, `theme_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: high: external Okta admin API mutation; approval required.
- `delete_api_v1_brands_brand_id_themes_theme_id_favicon`: DELETE `/api/v1/brands/{{ record.brand_id
  }}/themes/{{ record.theme_id }}/favicon` - kind `delete`; body type `none`; path fields
  `brand_id`, `theme_id`; required record fields `brand_id`, `theme_id`; accepted fields `brand_id`,
  `theme_id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  high: external Okta admin API mutation; approval required.
- `delete_api_v1_brands_brand_id_themes_theme_id_logo`: DELETE `/api/v1/brands/{{ record.brand_id
  }}/themes/{{ record.theme_id }}/logo` - kind `delete`; body type `none`; path fields `brand_id`,
  `theme_id`; required record fields `brand_id`, `theme_id`; accepted fields `brand_id`, `theme_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: high:
  external Okta admin API mutation; approval required.
- `update_api_v1_brands_brand_id_well_known_uris_path_customized`: PUT `/api/v1/brands/{{
  record.brand_id }}/well-known-uris/{{ record.path }}/customized` - kind `update`; body type
  `json`; path fields `brand_id`, `path`; required record fields `brand_id`, `path`,
  `representation`; accepted fields `brand_id`, `path`, `representation`; risk: medium: external
  Okta admin API mutation; approval required.
- `create_api_v1_captchas`: POST `/api/v1/captchas` - kind `create`; body type `json`; accepted
  fields `_links`, `id`, `name`, `secretKey`, `siteKey`, `type`; risk: medium: external Okta admin
  API mutation; approval required.
- `create_api_v1_captchas_captcha_id`: POST `/api/v1/captchas/{{ record.captcha_id }}` - kind
  `create`; body type `json`; path fields `captcha_id`; required record fields `captcha_id`;
  accepted fields `_links`, `captcha_id`, `id`, `name`, `secretKey`, `siteKey`, `type`; risk:
  medium: external Okta admin API mutation; approval required.
- `update_api_v1_captchas_captcha_id`: PUT `/api/v1/captchas/{{ record.captcha_id }}` - kind
  `update`; body type `json`; path fields `captcha_id`; required record fields `captcha_id`;
  accepted fields `_links`, `captcha_id`, `id`, `name`, `secretKey`, `siteKey`, `type`; risk:
  medium: external Okta admin API mutation; approval required.
- `delete_api_v1_captchas_captcha_id`: DELETE `/api/v1/captchas/{{ record.captcha_id }}` - kind
  `delete`; body type `none`; path fields `captcha_id`; required record fields `captcha_id`;
  accepted fields `captcha_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Okta admin API mutation; approval required.
- `create_api_v1_device_assurances`: POST `/api/v1/device-assurances` - kind `create`; body type
  `json`; accepted fields `_links`, `createdBy`, `createdDate`, `devicePostureChecks`,
  `displayRemediationMode`, `gracePeriod`, `id`, `lastUpdate`, `lastUpdatedBy`, `name`, `platform`;
  risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_device_assurances_device_assurance_id`: PUT `/api/v1/device-assurances/{{
  record.device_assurance_id }}` - kind `update`; body type `json`; path fields
  `device_assurance_id`; required record fields `device_assurance_id`; accepted fields `_links`,
  `createdBy`, `createdDate`, `devicePostureChecks`, `device_assurance_id`,
  `displayRemediationMode`, `gracePeriod`, `id`, `lastUpdate`, `lastUpdatedBy`, `name`, `platform`;
  risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_device_assurances_device_assurance_id`: DELETE `/api/v1/device-assurances/{{
  record.device_assurance_id }}` - kind `delete`; body type `none`; path fields
  `device_assurance_id`; required record fields `device_assurance_id`; accepted fields
  `device_assurance_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_device_integrations_device_integration_id_lifecycle_activate`: POST
  `/api/v1/device-integrations/{{ record.device_integration_id }}/lifecycle/activate` - kind
  `custom`; body type `none`; path fields `device_integration_id`; required record fields
  `device_integration_id`; accepted fields `device_integration_id`; risk: high: external Okta admin
  API mutation; approval required.
- `execute_api_v1_device_integrations_device_integration_id_lifecycle_deactivate`: POST
  `/api/v1/device-integrations/{{ record.device_integration_id }}/lifecycle/deactivate` - kind
  `custom`; body type `none`; path fields `device_integration_id`; required record fields
  `device_integration_id`; accepted fields `device_integration_id`; risk: high: external Okta admin
  API mutation; approval required.
- `create_api_v1_device_posture_checks`: POST `/api/v1/device-posture-checks` - kind `create`; body
  type `json`; accepted fields `_links`, `createdBy`, `createdDate`, `description`, `id`,
  `lastUpdate`, `lastUpdatedBy`, `mappingType`, `name`, `platform`, `query`, `remediationSettings`,
  `type`, `variableName`; risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_device_posture_checks_posture_check_id`: PUT `/api/v1/device-posture-checks/{{
  record.posture_check_id }}` - kind `update`; body type `json`; path fields `posture_check_id`;
  required record fields `posture_check_id`; accepted fields `_links`, `createdBy`, `createdDate`,
  `description`, `id`, `lastUpdate`, `lastUpdatedBy`, `mappingType`, `name`, `platform`,
  `posture_check_id`, `query`, `remediationSettings`, `type`, `variableName`; risk: medium: external
  Okta admin API mutation; approval required.
- `delete_api_v1_device_posture_checks_posture_check_id`: DELETE `/api/v1/device-posture-checks/{{
  record.posture_check_id }}` - kind `delete`; body type `none`; path fields `posture_check_id`;
  required record fields `posture_check_id`; accepted fields `posture_check_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Okta admin
  API mutation; approval required.
- `delete_api_v1_devices_device_id`: DELETE `/api/v1/devices/{{ record.device_id }}` - kind
  `delete`; body type `none`; path fields `device_id`; required record fields `device_id`; accepted
  fields `device_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_devices_device_id_lifecycle_activate`: POST `/api/v1/devices/{{ record.device_id
  }}/lifecycle/activate` - kind `custom`; body type `none`; path fields `device_id`; required record
  fields `device_id`; accepted fields `device_id`; risk: high: external Okta admin API mutation;
  approval required.
- `execute_api_v1_devices_device_id_lifecycle_deactivate`: POST `/api/v1/devices/{{ record.device_id
  }}/lifecycle/deactivate` - kind `custom`; body type `none`; path fields `device_id`; required
  record fields `device_id`; accepted fields `device_id`; risk: high: external Okta admin API
  mutation; approval required.
- `execute_api_v1_devices_device_id_lifecycle_suspend`: POST `/api/v1/devices/{{ record.device_id
  }}/lifecycle/suspend` - kind `custom`; body type `none`; path fields `device_id`; required record
  fields `device_id`; accepted fields `device_id`; risk: high: external Okta admin API mutation;
  approval required.
- `execute_api_v1_devices_device_id_lifecycle_unsuspend`: POST `/api/v1/devices/{{ record.device_id
  }}/lifecycle/unsuspend` - kind `custom`; body type `none`; path fields `device_id`; required
  record fields `device_id`; accepted fields `device_id`; risk: high: external Okta admin API
  mutation; approval required.
- `create_api_v1_directories_app_instance_id_groups_modify`: POST `/api/v1/directories/{{
  record.app_instance_id }}/groups/modify` - kind `create`; body type `json`; path fields
  `app_instance_id`; required record fields `app_instance_id`, `id`, `parameters`; accepted fields
  `app_instance_id`, `id`, `parameters`; risk: medium: external Okta admin API mutation; approval
  required.
- `create_api_v1_directories_app_instance_id_groups_group_id_query`: POST `/api/v1/directories/{{
  record.app_instance_id }}/groups/{{ record.group_id }}/query` - kind `create`; body type `json`;
  path fields `app_instance_id`, `group_id`; required record fields `app_instance_id`, `group_id`,
  `attributes`; accepted fields `app_instance_id`, `attributes`, `group_id`; risk: medium: external
  Okta admin API mutation; approval required.
- `create_api_v1_domains`: POST `/api/v1/domains` - kind `create`; body type `json`; required record
  fields `certificateSourceType`, `domain`; accepted fields `certificateSourceType`, `domain`; risk:
  medium: external Okta admin API mutation; approval required.
- `update_api_v1_domains_domain_id`: PUT `/api/v1/domains/{{ record.domain_id }}` - kind `update`;
  body type `json`; path fields `domain_id`; required record fields `domain_id`, `brandId`; accepted
  fields `brandId`, `domain_id`; risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_domains_domain_id`: DELETE `/api/v1/domains/{{ record.domain_id }}` - kind
  `delete`; body type `none`; path fields `domain_id`; required record fields `domain_id`; accepted
  fields `domain_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Okta admin API mutation; approval required.
- `update_api_v1_domains_domain_id_certificate`: PUT `/api/v1/domains/{{ record.domain_id
  }}/certificate` - kind `update`; body type `json`; path fields `domain_id`; required record fields
  `domain_id`, `certificate`, `certificateChain`, `privateKey`, `type`; accepted fields
  `certificate`, `certificateChain`, `domain_id`, `privateKey`, `type`; risk: medium: external Okta
  admin API mutation; approval required.
- `create_api_v1_domains_domain_id_verify`: POST `/api/v1/domains/{{ record.domain_id }}/verify` -
  kind `create`; body type `none`; path fields `domain_id`; required record fields `domain_id`;
  accepted fields `domain_id`; risk: medium: external Okta admin API mutation; approval required.
- `create_api_v1_dr_failback`: POST `/api/v1/dr/failback` - kind `create`; body type `json`;
  accepted fields `domains`; risk: medium: external Okta admin API mutation; approval required.
- `create_api_v1_dr_failover`: POST `/api/v1/dr/failover` - kind `create`; body type `json`;
  accepted fields `domains`; risk: medium: external Okta admin API mutation; approval required.
- `create_api_v1_email_domains`: POST `/api/v1/email-domains` - kind `create`; body type `json`;
  required record fields `displayName`, `userName`; accepted fields `displayName`, `userName`; risk:
  medium: external Okta admin API mutation; approval required.
- `update_api_v1_email_domains_email_domain_id`: PUT `/api/v1/email-domains/{{
  record.email_domain_id }}` - kind `update`; body type `json`; path fields `email_domain_id`;
  required record fields `email_domain_id`, `displayName`, `userName`; accepted fields
  `displayName`, `email_domain_id`, `userName`; risk: medium: external Okta admin API mutation;
  approval required.
- `delete_api_v1_email_domains_email_domain_id`: DELETE `/api/v1/email-domains/{{
  record.email_domain_id }}` - kind `delete`; body type `none`; path fields `email_domain_id`;
  required record fields `email_domain_id`; accepted fields `email_domain_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Okta admin
  API mutation; approval required.
- `create_api_v1_email_domains_email_domain_id_verify`: POST `/api/v1/email-domains/{{
  record.email_domain_id }}/verify` - kind `create`; body type `none`; path fields
  `email_domain_id`; required record fields `email_domain_id`; accepted fields `email_domain_id`;
  risk: medium: external Okta admin API mutation; approval required.
- `create_api_v1_email_servers`: POST `/api/v1/email-servers` - kind `create`; body type `json`;
  required record fields `alias`, `enabled`, `host`, `port`, `username`, `authType`; accepted fields
  `alias`, `authType`, `enabled`, `host`, `id`, `port`, `username`; risk: medium: external Okta
  admin API mutation; approval required.
- `update_api_v1_email_servers_email_server_id`: PATCH `/api/v1/email-servers/{{
  record.email_server_id }}` - kind `update`; body type `json`; path fields `email_server_id`;
  required record fields `email_server_id`, `authType`; accepted fields `alias`, `authType`,
  `email_server_id`, `enabled`, `host`, `id`, `port`, `username`; risk: medium: external Okta admin
  API mutation; approval required.
- `delete_api_v1_email_servers_email_server_id`: DELETE `/api/v1/email-servers/{{
  record.email_server_id }}` - kind `delete`; body type `none`; path fields `email_server_id`;
  required record fields `email_server_id`; accepted fields `email_server_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Okta admin
  API mutation; approval required.
- `create_api_v1_email_servers_email_server_id_test`: POST `/api/v1/email-servers/{{
  record.email_server_id }}/test` - kind `create`; body type `json`; path fields `email_server_id`;
  required record fields `email_server_id`, `fromAddress`, `toAddress`; accepted fields
  `email_server_id`, `fromAddress`, `toAddress`; risk: medium: external Okta admin API mutation;
  approval required.
- `create_api_v1_event_hooks`: POST `/api/v1/eventHooks` - kind `create`; body type `json`; required
  record fields `name`, `events`, `channel`; accepted fields `_links`, `channel`, `created`,
  `createdBy`, `description`, `events`, `id`, `lastUpdated`, `name`, `status`, `verificationStatus`;
  risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_event_hooks_event_hook_id`: PUT `/api/v1/eventHooks/{{ record.event_hook_id }}` -
  kind `update`; body type `json`; path fields `event_hook_id`; required record fields
  `event_hook_id`, `name`, `events`, `channel`; accepted fields `_links`, `channel`, `created`,
  `createdBy`, `description`, `event_hook_id`, `events`, `id`, `lastUpdated`, `name`, `status`,
  `verificationStatus`; risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_event_hooks_event_hook_id`: DELETE `/api/v1/eventHooks/{{ record.event_hook_id }}`
  - kind `delete`; body type `none`; path fields `event_hook_id`; required record fields
  `event_hook_id`; accepted fields `event_hook_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: high: external Okta admin API mutation; approval
  required.
- `execute_api_v1_event_hooks_event_hook_id_lifecycle_activate`: POST `/api/v1/eventHooks/{{
  record.event_hook_id }}/lifecycle/activate` - kind `custom`; body type `none`; path fields
  `event_hook_id`; required record fields `event_hook_id`; accepted fields `event_hook_id`; risk:
  high: external Okta admin API mutation; approval required.
- `execute_api_v1_event_hooks_event_hook_id_lifecycle_deactivate`: POST `/api/v1/eventHooks/{{
  record.event_hook_id }}/lifecycle/deactivate` - kind `custom`; body type `none`; path fields
  `event_hook_id`; required record fields `event_hook_id`; accepted fields `event_hook_id`; risk:
  high: external Okta admin API mutation; approval required.
- `execute_api_v1_event_hooks_event_hook_id_lifecycle_verify`: POST `/api/v1/eventHooks/{{
  record.event_hook_id }}/lifecycle/verify` - kind `custom`; body type `none`; path fields
  `event_hook_id`; required record fields `event_hook_id`; accepted fields `event_hook_id`; risk:
  high: external Okta admin API mutation; approval required.
- `create_api_v1_features_feature_id_lifecycle`: POST `/api/v1/features/{{ record.feature_id }}/{{
  record.lifecycle }}` - kind `create`; body type `none`; path fields `feature_id`, `lifecycle`;
  required record fields `feature_id`, `lifecycle`; accepted fields `feature_id`, `lifecycle`; risk:
  medium: external Okta admin API mutation; approval required.
- `update_api_v1_first_party_app_settings_app_name`: PUT `/api/v1/first-party-app-settings/{{
  record.app_name }}` - kind `update`; body type `json`; path fields `app_name`; required record
  fields `app_name`; accepted fields `app_name`, `sessionIdleTimeoutMinutes`,
  `sessionMaxLifetimeMinutes`; risk: medium: external Okta admin API mutation; approval required.
- `create_api_v1_groups`: POST `/api/v1/groups` - kind `create`; body type `json`; accepted fields
  `profile`; risk: medium: external Okta admin API mutation; approval required.
- `create_api_v1_groups_rules`: POST `/api/v1/groups/rules` - kind `create`; body type `json`;
  accepted fields `actions`, `conditions`, `name`, `type`; risk: medium: external Okta admin API
  mutation; approval required.
- `update_api_v1_groups_rules_group_rule_id`: PUT `/api/v1/groups/rules/{{ record.group_rule_id }}`
  - kind `update`; body type `json`; path fields `group_rule_id`; required record fields
  `group_rule_id`; accepted fields `_embedded`, `actions`, `conditions`, `created`, `group_rule_id`,
  `id`, `lastUpdated`, `name`, `status`, `type`; risk: medium: external Okta admin API mutation;
  approval required.
- `delete_api_v1_groups_rules_group_rule_id`: DELETE `/api/v1/groups/rules/{{ record.group_rule_id
  }}` - kind `delete`; body type `none`; path fields `group_rule_id`; required record fields
  `group_rule_id`; accepted fields `group_rule_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: high: external Okta admin API mutation; approval
  required.
- `execute_api_v1_groups_rules_group_rule_id_lifecycle_activate`: POST `/api/v1/groups/rules/{{
  record.group_rule_id }}/lifecycle/activate` - kind `custom`; body type `none`; path fields
  `group_rule_id`; required record fields `group_rule_id`; accepted fields `group_rule_id`; risk:
  high: external Okta admin API mutation; approval required.
- `execute_api_v1_groups_rules_group_rule_id_lifecycle_deactivate`: POST `/api/v1/groups/rules/{{
  record.group_rule_id }}/lifecycle/deactivate` - kind `custom`; body type `none`; path fields
  `group_rule_id`; required record fields `group_rule_id`; accepted fields `group_rule_id`; risk:
  high: external Okta admin API mutation; approval required.
- `update_api_v1_groups_group_id`: PUT `/api/v1/groups/{{ record.group_id }}` - kind `update`; body
  type `json`; path fields `group_id`; required record fields `group_id`; accepted fields
  `group_id`, `profile`; risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_groups_group_id`: DELETE `/api/v1/groups/{{ record.group_id }}` - kind `delete`;
  body type `none`; path fields `group_id`; required record fields `group_id`; accepted fields
  `group_id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  high: external Okta admin API mutation; approval required.
- `create_api_v1_groups_group_id_owners`: POST `/api/v1/groups/{{ record.group_id }}/owners` - kind
  `create`; body type `json`; path fields `group_id`; required record fields `group_id`; accepted
  fields `group_id`, `id`, `type`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_api_v1_groups_group_id_owners_owner_id`: DELETE `/api/v1/groups/{{ record.group_id
  }}/owners/{{ record.owner_id }}` - kind `delete`; body type `none`; path fields `group_id`,
  `owner_id`; required record fields `group_id`, `owner_id`; accepted fields `group_id`, `owner_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: high:
  external Okta admin API mutation; approval required.
- `create_api_v1_groups_group_id_roles`: POST `/api/v1/groups/{{ record.group_id }}/roles` - kind
  `create`; body type `json`; path fields `group_id`; required record fields `group_id`, `type`;
  accepted fields `group_id`, `type`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_api_v1_groups_group_id_roles_role_assignment_id`: DELETE `/api/v1/groups/{{
  record.group_id }}/roles/{{ record.role_assignment_id }}` - kind `delete`; body type `none`; path
  fields `group_id`, `role_assignment_id`; required record fields `group_id`, `role_assignment_id`;
  accepted fields `group_id`, `role_assignment_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: high: external Okta admin API mutation; approval
  required.
- `update_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name`: PUT
  `/api/v1/groups/{{ record.group_id }}/roles/{{ record.role_assignment_id
  }}/targets/catalog/apps/{{ record.app_name }}` - kind `update`; body type `none`; path fields
  `group_id`, `role_assignment_id`, `app_name`; required record fields `group_id`,
  `role_assignment_id`, `app_name`; accepted fields `app_name`, `group_id`, `role_assignment_id`;
  risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name`: DELETE
  `/api/v1/groups/{{ record.group_id }}/roles/{{ record.role_assignment_id
  }}/targets/catalog/apps/{{ record.app_name }}` - kind `delete`; body type `none`; path fields
  `group_id`, `role_assignment_id`, `app_name`; required record fields `group_id`,
  `role_assignment_id`, `app_name`; accepted fields `app_name`, `group_id`, `role_assignment_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: high:
  external Okta admin API mutation; approval required.
- `update_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id`: PUT
  `/api/v1/groups/{{ record.group_id }}/roles/{{ record.role_assignment_id
  }}/targets/catalog/apps/{{ record.app_name }}/{{ record.app_id }}` - kind `update`; body type
  `none`; path fields `group_id`, `role_assignment_id`, `app_name`, `app_id`; required record fields
  `group_id`, `role_assignment_id`, `app_name`, `app_id`; accepted fields `app_id`, `app_name`,
  `group_id`, `role_assignment_id`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_api_v1_groups_group_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id`:
  DELETE `/api/v1/groups/{{ record.group_id }}/roles/{{ record.role_assignment_id
  }}/targets/catalog/apps/{{ record.app_name }}/{{ record.app_id }}` - kind `delete`; body type
  `none`; path fields `group_id`, `role_assignment_id`, `app_name`, `app_id`; required record fields
  `group_id`, `role_assignment_id`, `app_name`, `app_id`; accepted fields `app_id`, `app_name`,
  `group_id`, `role_assignment_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: high: external Okta admin API mutation; approval required.
- `update_api_v1_groups_group_id_roles_role_assignment_id_targets_groups_target_group_id`: PUT
  `/api/v1/groups/{{ record.group_id }}/roles/{{ record.role_assignment_id }}/targets/groups/{{
  record.target_group_id }}` - kind `update`; body type `none`; path fields `group_id`,
  `role_assignment_id`, `target_group_id`; required record fields `group_id`, `role_assignment_id`,
  `target_group_id`; accepted fields `group_id`, `role_assignment_id`, `target_group_id`; risk:
  medium: external Okta admin API mutation; approval required.
- `delete_api_v1_groups_group_id_roles_role_assignment_id_targets_groups_target_group_id`: DELETE
  `/api/v1/groups/{{ record.group_id }}/roles/{{ record.role_assignment_id }}/targets/groups/{{
  record.target_group_id }}` - kind `delete`; body type `none`; path fields `group_id`,
  `role_assignment_id`, `target_group_id`; required record fields `group_id`, `role_assignment_id`,
  `target_group_id`; accepted fields `group_id`, `role_assignment_id`, `target_group_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: high: external Okta
  admin API mutation; approval required.
- `update_api_v1_groups_group_id_users_user_id`: PUT `/api/v1/groups/{{ record.group_id }}/users/{{
  record.user_id }}` - kind `update`; body type `none`; path fields `group_id`, `user_id`; required
  record fields `group_id`, `user_id`; accepted fields `group_id`, `user_id`; risk: medium: external
  Okta admin API mutation; approval required.
- `delete_api_v1_groups_group_id_users_user_id`: DELETE `/api/v1/groups/{{ record.group_id
  }}/users/{{ record.user_id }}` - kind `delete`; body type `none`; path fields `group_id`,
  `user_id`; required record fields `group_id`, `user_id`; accepted fields `group_id`, `user_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: high:
  external Okta admin API mutation; approval required.
- `create_api_v1_hook_keys`: POST `/api/v1/hook-keys` - kind `create`; body type `json`; accepted
  fields `name`; risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_hook_keys_id`: PUT `/api/v1/hook-keys/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `id`, `name`; risk: medium:
  external Okta admin API mutation; approval required.
- `delete_api_v1_hook_keys_id`: DELETE `/api/v1/hook-keys/{{ record.id }}` - kind `delete`; body
  type `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Okta admin
  API mutation; approval required.
- `create_api_v1_iam_governance_bundles`: POST `/api/v1/iam/governance/bundles` - kind `create`;
  body type `json`; accepted fields `description`, `entitlements`, `name`; risk: medium: external
  Okta admin API mutation; approval required.
- `update_api_v1_iam_governance_bundles_bundle_id`: PUT `/api/v1/iam/governance/bundles/{{
  record.bundle_id }}` - kind `update`; body type `json`; path fields `bundle_id`; required record
  fields `bundle_id`; accepted fields `bundle_id`, `description`, `entitlements`, `name`; risk:
  medium: external Okta admin API mutation; approval required.
- `delete_api_v1_iam_governance_bundles_bundle_id`: DELETE `/api/v1/iam/governance/bundles/{{
  record.bundle_id }}` - kind `delete`; body type `none`; path fields `bundle_id`; required record
  fields `bundle_id`; accepted fields `bundle_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: high: external Okta admin API mutation; approval
  required.
- `create_api_v1_iam_governance_opt_in`: POST `/api/v1/iam/governance/optIn` - kind `create`; body
  type `none`; risk: medium: external Okta admin API mutation; approval required.
- `create_api_v1_iam_governance_opt_out`: POST `/api/v1/iam/governance/optOut` - kind `create`; body
  type `none`; risk: medium: external Okta admin API mutation; approval required.
- `create_api_v1_iam_resource_sets`: POST `/api/v1/iam/resource-sets` - kind `create`; body type
  `json`; required record fields `description`, `label`, `resources`; accepted fields `description`,
  `label`, `resources`; risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_iam_resource_sets_resource_set_id_or_label`: PUT `/api/v1/iam/resource-sets/{{
  record.resource_set_id_or_label }}` - kind `update`; body type `json`; path fields
  `resource_set_id_or_label`; required record fields `resource_set_id_or_label`; accepted fields
  `_links`, `created`, `description`, `id`, `label`, `lastUpdated`, `resource_set_id_or_label`;
  risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_iam_resource_sets_resource_set_id_or_label`: DELETE `/api/v1/iam/resource-sets/{{
  record.resource_set_id_or_label }}` - kind `delete`; body type `none`; path fields
  `resource_set_id_or_label`; required record fields `resource_set_id_or_label`; accepted fields
  `resource_set_id_or_label`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Okta admin API mutation; approval required.
- `create_api_v1_iam_resource_sets_resource_set_id_or_label_bindings`: POST
  `/api/v1/iam/resource-sets/{{ record.resource_set_id_or_label }}/bindings` - kind `create`; body
  type `json`; path fields `resource_set_id_or_label`; required record fields
  `resource_set_id_or_label`; accepted fields `members`, `resource_set_id_or_label`, `role`; risk:
  medium: external Okta admin API mutation; approval required.
- `delete_api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label`: DELETE
  `/api/v1/iam/resource-sets/{{ record.resource_set_id_or_label }}/bindings/{{
  record.role_id_or_label }}` - kind `delete`; body type `none`; path fields
  `resource_set_id_or_label`, `role_id_or_label`; required record fields `resource_set_id_or_label`,
  `role_id_or_label`; accepted fields `resource_set_id_or_label`, `role_id_or_label`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: high: external Okta
  admin API mutation; approval required.
- `update_api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members`:
  PATCH `/api/v1/iam/resource-sets/{{ record.resource_set_id_or_label }}/bindings/{{
  record.role_id_or_label }}/members` - kind `update`; body type `json`; path fields
  `resource_set_id_or_label`, `role_id_or_label`; required record fields `resource_set_id_or_label`,
  `role_id_or_label`; accepted fields `additions`, `resource_set_id_or_label`, `role_id_or_label`;
  risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_iam_resource_sets_resource_set_id_or_label_bindings_role_id_or_label_members_member_id`:
  DELETE `/api/v1/iam/resource-sets/{{ record.resource_set_id_or_label }}/bindings/{{
  record.role_id_or_label }}/members/{{ record.member_id }}` - kind `delete`; body type `none`; path
  fields `resource_set_id_or_label`, `role_id_or_label`, `member_id`; required record fields
  `resource_set_id_or_label`, `role_id_or_label`, `member_id`; accepted fields `member_id`,
  `resource_set_id_or_label`, `role_id_or_label`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: high: external Okta admin API mutation; approval
  required.
- `create_api_v1_iam_resource_sets_resource_set_id_or_label_resources`: POST
  `/api/v1/iam/resource-sets/{{ record.resource_set_id_or_label }}/resources` - kind `create`; body
  type `json`; path fields `resource_set_id_or_label`; required record fields
  `resource_set_id_or_label`, `resourceOrnOrUrl`, `conditions`; accepted fields `conditions`,
  `resourceOrnOrUrl`, `resource_set_id_or_label`; risk: medium: external Okta admin API mutation;
  approval required.
- `update_api_v1_iam_resource_sets_resource_set_id_or_label_resources`: PATCH
  `/api/v1/iam/resource-sets/{{ record.resource_set_id_or_label }}/resources` - kind `update`; body
  type `json`; path fields `resource_set_id_or_label`; required record fields
  `resource_set_id_or_label`; accepted fields `additions`, `resource_set_id_or_label`; risk: medium:
  external Okta admin API mutation; approval required.
- `update_api_v1_iam_resource_sets_resource_set_id_or_label_resources_resource_id`: PUT
  `/api/v1/iam/resource-sets/{{ record.resource_set_id_or_label }}/resources/{{ record.resource_id
  }}` - kind `update`; body type `json`; path fields `resource_set_id_or_label`, `resource_id`;
  required record fields `resource_set_id_or_label`, `resource_id`; accepted fields `conditions`,
  `resource_id`, `resource_set_id_or_label`; risk: medium: external Okta admin API mutation;
  approval required.
- `delete_api_v1_iam_resource_sets_resource_set_id_or_label_resources_resource_id`: DELETE
  `/api/v1/iam/resource-sets/{{ record.resource_set_id_or_label }}/resources/{{ record.resource_id
  }}` - kind `delete`; body type `none`; path fields `resource_set_id_or_label`, `resource_id`;
  required record fields `resource_set_id_or_label`, `resource_id`; accepted fields `resource_id`,
  `resource_set_id_or_label`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Okta admin API mutation; approval required.
- `create_api_v1_iam_roles`: POST `/api/v1/iam/roles` - kind `create`; body type `json`; required
  record fields `label`, `description`, `permissions`; accepted fields `description`, `label`,
  `permissions`; risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_iam_roles_role_id_or_label`: PUT `/api/v1/iam/roles/{{ record.role_id_or_label }}`
  - kind `update`; body type `json`; path fields `role_id_or_label`; required record fields
  `role_id_or_label`, `label`, `description`; accepted fields `description`, `label`,
  `role_id_or_label`; risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_iam_roles_role_id_or_label`: DELETE `/api/v1/iam/roles/{{ record.role_id_or_label
  }}` - kind `delete`; body type `none`; path fields `role_id_or_label`; required record fields
  `role_id_or_label`; accepted fields `role_id_or_label`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: high: external Okta admin API mutation; approval
  required.
- `create_api_v1_iam_roles_role_id_or_label_permissions_permission_type`: POST `/api/v1/iam/roles/{{
  record.role_id_or_label }}/permissions/{{ record.permission_type }}` - kind `create`; body type
  `json`; path fields `role_id_or_label`, `permission_type`; required record fields
  `role_id_or_label`, `permission_type`; accepted fields `conditions`, `permission_type`,
  `role_id_or_label`; risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_iam_roles_role_id_or_label_permissions_permission_type`: PUT `/api/v1/iam/roles/{{
  record.role_id_or_label }}/permissions/{{ record.permission_type }}` - kind `update`; body type
  `json`; path fields `role_id_or_label`, `permission_type`; required record fields
  `role_id_or_label`, `permission_type`; accepted fields `conditions`, `permission_type`,
  `role_id_or_label`; risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_iam_roles_role_id_or_label_permissions_permission_type`: DELETE
  `/api/v1/iam/roles/{{ record.role_id_or_label }}/permissions/{{ record.permission_type }}` - kind
  `delete`; body type `none`; path fields `role_id_or_label`, `permission_type`; required record
  fields `role_id_or_label`, `permission_type`; accepted fields `permission_type`,
  `role_id_or_label`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Okta admin API mutation; approval required.
- `create_api_v1_identity_sources_identity_source_id_groups`: POST `/api/v1/identity-sources/{{
  record.identity_source_id }}/groups` - kind `create`; body type `json`; path fields
  `identity_source_id`; required record fields `identity_source_id`; accepted fields `externalId`,
  `identity_source_id`, `profile`; risk: medium: external Okta admin API mutation; approval
  required.
- `create_api_v1_identity_sources_identity_source_id_groups_group_or_external_id`: POST
  `/api/v1/identity-sources/{{ record.identity_source_id }}/groups/{{ record.group_or_external_id
  }}` - kind `create`; body type `json`; path fields `identity_source_id`, `group_or_external_id`;
  required record fields `identity_source_id`, `group_or_external_id`; accepted fields `externalId`,
  `group_or_external_id`, `identity_source_id`, `profile`; risk: medium: external Okta admin API
  mutation; approval required.
- `delete_api_v1_identity_sources_identity_source_id_groups_group_or_external_id`: DELETE
  `/api/v1/identity-sources/{{ record.identity_source_id }}/groups/{{ record.group_or_external_id
  }}` - kind `delete`; body type `none`; path fields `identity_source_id`, `group_or_external_id`;
  required record fields `identity_source_id`, `group_or_external_id`; accepted fields
  `group_or_external_id`, `identity_source_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: high: external Okta admin API mutation; approval required.
- `create_api_v1_identity_sources_identity_source_id_groups_group_or_external_id_membership`: POST
  `/api/v1/identity-sources/{{ record.identity_source_id }}/groups/{{ record.group_or_external_id
  }}/membership` - kind `create`; body type `json`; path fields `identity_source_id`,
  `group_or_external_id`; required record fields `identity_source_id`, `group_or_external_id`;
  accepted fields `group_or_external_id`, `identity_source_id`, `memberExternalId`; risk: medium:
  external Okta admin API mutation; approval required.
- `delete_api_v1_identity_sources_identity_source_id_groups_group_or_external_id_membership_member_external_id`:
  DELETE `/api/v1/identity-sources/{{ record.identity_source_id }}/groups/{{
  record.group_or_external_id }}/membership/{{ record.member_external_id }}` - kind `delete`; body
  type `none`; path fields `identity_source_id`, `group_or_external_id`, `member_external_id`;
  required record fields `identity_source_id`, `group_or_external_id`, `member_external_id`;
  accepted fields `group_or_external_id`, `identity_source_id`, `member_external_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: high: external Okta
  admin API mutation; approval required.
- `create_api_v1_identity_sources_identity_source_id_sessions`: POST `/api/v1/identity-sources/{{
  record.identity_source_id }}/sessions` - kind `create`; body type `none`; path fields
  `identity_source_id`; required record fields `identity_source_id`; accepted fields
  `identity_source_id`; risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_identity_sources_identity_source_id_sessions_session_id`: DELETE
  `/api/v1/identity-sources/{{ record.identity_source_id }}/sessions/{{ record.session_id }}` - kind
  `delete`; body type `none`; path fields `identity_source_id`, `session_id`; required record fields
  `identity_source_id`, `session_id`; accepted fields `identity_source_id`, `session_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: high: external Okta
  admin API mutation; approval required.
- `create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_delete`: POST
  `/api/v1/identity-sources/{{ record.identity_source_id }}/sessions/{{ record.session_id
  }}/bulk-delete` - kind `create`; body type `json`; path fields `identity_source_id`, `session_id`;
  required record fields `identity_source_id`, `session_id`; accepted fields `entityType`,
  `identity_source_id`, `profiles`, `session_id`; risk: medium: external Okta admin API mutation;
  approval required.
- `create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_group_memberships_delete`:
  POST `/api/v1/identity-sources/{{ record.identity_source_id }}/sessions/{{ record.session_id
  }}/bulk-group-memberships-delete` - kind `create`; body type `json`; path fields
  `identity_source_id`, `session_id`; required record fields `identity_source_id`, `session_id`;
  accepted fields `identity_source_id`, `memberships`, `session_id`; risk: medium: external Okta
  admin API mutation; approval required.
- `create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_group_memberships_upsert`:
  POST `/api/v1/identity-sources/{{ record.identity_source_id }}/sessions/{{ record.session_id
  }}/bulk-group-memberships-upsert` - kind `create`; body type `json`; path fields
  `identity_source_id`, `session_id`; required record fields `identity_source_id`, `session_id`;
  accepted fields `identity_source_id`, `memberships`, `session_id`; risk: medium: external Okta
  admin API mutation; approval required.
- `create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_groups_delete`: POST
  `/api/v1/identity-sources/{{ record.identity_source_id }}/sessions/{{ record.session_id
  }}/bulk-groups-delete` - kind `create`; body type `json`; path fields `identity_source_id`,
  `session_id`; required record fields `identity_source_id`, `session_id`; accepted fields
  `externalIds`, `identity_source_id`, `session_id`; risk: medium: external Okta admin API mutation;
  approval required.
- `create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_groups_upsert`: POST
  `/api/v1/identity-sources/{{ record.identity_source_id }}/sessions/{{ record.session_id
  }}/bulk-groups-upsert` - kind `create`; body type `json`; path fields `identity_source_id`,
  `session_id`; required record fields `identity_source_id`, `session_id`; accepted fields
  `identity_source_id`, `profiles`, `session_id`; risk: medium: external Okta admin API mutation;
  approval required.
- `create_api_v1_identity_sources_identity_source_id_sessions_session_id_bulk_upsert`: POST
  `/api/v1/identity-sources/{{ record.identity_source_id }}/sessions/{{ record.session_id
  }}/bulk-upsert` - kind `create`; body type `json`; path fields `identity_source_id`, `session_id`;
  required record fields `identity_source_id`, `session_id`; accepted fields `entityType`,
  `identity_source_id`, `profiles`, `session_id`; risk: medium: external Okta admin API mutation;
  approval required.
- `create_api_v1_identity_sources_identity_source_id_sessions_session_id_start_import`: POST
  `/api/v1/identity-sources/{{ record.identity_source_id }}/sessions/{{ record.session_id
  }}/start-import` - kind `create`; body type `none`; path fields `identity_source_id`,
  `session_id`; required record fields `identity_source_id`, `session_id`; accepted fields
  `identity_source_id`, `session_id`; risk: medium: external Okta admin API mutation; approval
  required.
- `create_api_v1_identity_sources_identity_source_id_users`: POST `/api/v1/identity-sources/{{
  record.identity_source_id }}/users` - kind `create`; body type `json`; path fields
  `identity_source_id`; required record fields `identity_source_id`; accepted fields `externalId`,
  `identity_source_id`, `profile`; risk: medium: external Okta admin API mutation; approval
  required.
- `update_api_v1_identity_sources_identity_source_id_users_external_id`: PUT
  `/api/v1/identity-sources/{{ record.identity_source_id }}/users/{{ record.external_id }}` - kind
  `update`; body type `json`; path fields `identity_source_id`, `external_id`; required record
  fields `identity_source_id`, `external_id`; accepted fields `externalId`, `external_id`,
  `identity_source_id`, `profile`; risk: medium: external Okta admin API mutation; approval
  required.
- `update_api_v1_identity_sources_identity_source_id_users_external_id_2`: PATCH
  `/api/v1/identity-sources/{{ record.identity_source_id }}/users/{{ record.external_id }}` - kind
  `update`; body type `json`; path fields `identity_source_id`, `external_id`; required record
  fields `identity_source_id`, `external_id`; accepted fields `external_id`, `identity_source_id`,
  `profile`; risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_identity_sources_identity_source_id_users_external_id`: DELETE
  `/api/v1/identity-sources/{{ record.identity_source_id }}/users/{{ record.external_id }}` - kind
  `delete`; body type `none`; path fields `identity_source_id`, `external_id`; required record
  fields `identity_source_id`, `external_id`; accepted fields `external_id`, `identity_source_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: high:
  external Okta admin API mutation; approval required.
- `create_api_v1_idps`: POST `/api/v1/idps` - kind `create`; body type `json`; accepted fields
  `_links`, `created`, `id`, `issuerMode`, `lastUpdated`, `name`, `policy`, `properties`,
  `protocol`, `status`, `type`; risk: medium: external Okta admin API mutation; approval required.
- `create_api_v1_idps_credentials_keys`: POST `/api/v1/idps/credentials/keys` - kind `create`; body
  type `json`; required record fields `x5c`; accepted fields `x5c`; risk: medium: external Okta
  admin API mutation; approval required.
- `update_api_v1_idps_credentials_keys_kid`: PUT `/api/v1/idps/credentials/keys/{{ record.kid }}` -
  kind `update`; body type `json`; path fields `kid`; required record fields `kid`; accepted fields
  `created`, `e`, `expiresAt`, `kid`, `kty`, `lastUpdated`, `n`, `use`, `x5c`, `x5t#S256`; risk:
  medium: external Okta admin API mutation; approval required.
- `delete_api_v1_idps_credentials_keys_kid`: DELETE `/api/v1/idps/credentials/keys/{{ record.kid }}`
  - kind `delete`; body type `none`; path fields `kid`; required record fields `kid`; accepted
  fields `kid`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: high: external Okta admin API mutation; approval required.
- `update_api_v1_idps_idp_id`: PUT `/api/v1/idps/{{ record.idp_id }}` - kind `update`; body type
  `json`; path fields `idp_id`; required record fields `idp_id`; accepted fields `_links`,
  `created`, `id`, `idp_id`, `issuerMode`, `lastUpdated`, `name`, `policy`, `properties`,
  `protocol`, `status`, `type`; risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_idps_idp_id`: DELETE `/api/v1/idps/{{ record.idp_id }}` - kind `delete`; body type
  `none`; path fields `idp_id`; required record fields `idp_id`; accepted fields `idp_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: high: external Okta
  admin API mutation; approval required.
- `create_api_v1_idps_idp_id_credentials_csrs`: POST `/api/v1/idps/{{ record.idp_id
  }}/credentials/csrs` - kind `create`; body type `json`; path fields `idp_id`; required record
  fields `idp_id`; accepted fields `idp_id`, `subject`, `subjectAltNames`; risk: medium: external
  Okta admin API mutation; approval required.
- `delete_api_v1_idps_idp_id_credentials_csrs_idp_csr_id`: DELETE `/api/v1/idps/{{ record.idp_id
  }}/credentials/csrs/{{ record.idp_csr_id }}` - kind `delete`; body type `none`; path fields
  `idp_id`, `idp_csr_id`; required record fields `idp_id`, `idp_csr_id`; accepted fields
  `idp_csr_id`, `idp_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_idps_idp_id_lifecycle_activate`: POST `/api/v1/idps/{{ record.idp_id
  }}/lifecycle/activate` - kind `custom`; body type `none`; path fields `idp_id`; required record
  fields `idp_id`; accepted fields `idp_id`; risk: high: external Okta admin API mutation; approval
  required.
- `execute_api_v1_idps_idp_id_lifecycle_deactivate`: POST `/api/v1/idps/{{ record.idp_id
  }}/lifecycle/deactivate` - kind `custom`; body type `none`; path fields `idp_id`; required record
  fields `idp_id`; accepted fields `idp_id`; risk: high: external Okta admin API mutation; approval
  required.
- `create_api_v1_idps_idp_id_users_user_id`: POST `/api/v1/idps/{{ record.idp_id }}/users/{{
  record.user_id }}` - kind `create`; body type `json`; path fields `idp_id`, `user_id`; required
  record fields `idp_id`, `user_id`; accepted fields `externalId`, `idp_id`, `user_id`; risk:
  medium: external Okta admin API mutation; approval required.
- `delete_api_v1_idps_idp_id_users_user_id`: DELETE `/api/v1/idps/{{ record.idp_id }}/users/{{
  record.user_id }}` - kind `delete`; body type `none`; path fields `idp_id`, `user_id`; required
  record fields `idp_id`, `user_id`; accepted fields `idp_id`, `user_id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: high: external Okta admin API
  mutation; approval required.
- `create_api_v1_inline_hooks`: POST `/api/v1/inlineHooks` - kind `create`; body type `json`;
  accepted fields `channel`, `name`, `type`, `version`; risk: medium: external Okta admin API
  mutation; approval required.
- `create_api_v1_inline_hooks_inline_hook_id`: POST `/api/v1/inlineHooks/{{ record.inline_hook_id
  }}` - kind `create`; body type `json`; path fields `inline_hook_id`; required record fields
  `inline_hook_id`; accepted fields `channel`, `inline_hook_id`, `name`, `version`; risk: medium:
  external Okta admin API mutation; approval required.
- `update_api_v1_inline_hooks_inline_hook_id`: PUT `/api/v1/inlineHooks/{{ record.inline_hook_id }}`
  - kind `update`; body type `json`; path fields `inline_hook_id`; required record fields
  `inline_hook_id`; accepted fields `channel`, `inline_hook_id`, `name`, `version`; risk: medium:
  external Okta admin API mutation; approval required.
- `delete_api_v1_inline_hooks_inline_hook_id`: DELETE `/api/v1/inlineHooks/{{ record.inline_hook_id
  }}` - kind `delete`; body type `none`; path fields `inline_hook_id`; required record fields
  `inline_hook_id`; accepted fields `inline_hook_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: high: external Okta admin API mutation; approval
  required.
- `create_api_v1_inline_hooks_inline_hook_id_execute`: POST `/api/v1/inlineHooks/{{
  record.inline_hook_id }}/execute` - kind `create`; body type `json`; path fields `inline_hook_id`;
  required record fields `inline_hook_id`; accepted fields `data`, `eventType`, `inline_hook_id`,
  `source`; risk: medium: external Okta admin API mutation; approval required.
- `execute_api_v1_inline_hooks_inline_hook_id_lifecycle_activate`: POST `/api/v1/inlineHooks/{{
  record.inline_hook_id }}/lifecycle/activate` - kind `custom`; body type `none`; path fields
  `inline_hook_id`; required record fields `inline_hook_id`; accepted fields `inline_hook_id`; risk:
  high: external Okta admin API mutation; approval required.
- `execute_api_v1_inline_hooks_inline_hook_id_lifecycle_deactivate`: POST `/api/v1/inlineHooks/{{
  record.inline_hook_id }}/lifecycle/deactivate` - kind `custom`; body type `none`; path fields
  `inline_hook_id`; required record fields `inline_hook_id`; accepted fields `inline_hook_id`; risk:
  high: external Okta admin API mutation; approval required.
- `create_api_v1_log_streams`: POST `/api/v1/logStreams` - kind `create`; body type `json`; required
  record fields `created`, `id`, `lastUpdated`, `name`, `status`, `type`, `_links`; accepted fields
  `_links`, `created`, `id`, `lastUpdated`, `name`, `status`, `type`; risk: medium: external Okta
  admin API mutation; approval required.
- `update_api_v1_log_streams_log_stream_id`: PUT `/api/v1/logStreams/{{ record.log_stream_id }}` -
  kind `update`; body type `json`; path fields `log_stream_id`; required record fields
  `log_stream_id`, `name`, `type`; accepted fields `log_stream_id`, `name`, `type`; risk: medium:
  external Okta admin API mutation; approval required.
- `delete_api_v1_log_streams_log_stream_id`: DELETE `/api/v1/logStreams/{{ record.log_stream_id }}`
  - kind `delete`; body type `none`; path fields `log_stream_id`; required record fields
  `log_stream_id`; accepted fields `log_stream_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: high: external Okta admin API mutation; approval
  required.
- `execute_api_v1_log_streams_log_stream_id_lifecycle_activate`: POST `/api/v1/logStreams/{{
  record.log_stream_id }}/lifecycle/activate` - kind `custom`; body type `none`; path fields
  `log_stream_id`; required record fields `log_stream_id`; accepted fields `log_stream_id`; risk:
  high: external Okta admin API mutation; approval required.
- `execute_api_v1_log_streams_log_stream_id_lifecycle_deactivate`: POST `/api/v1/logStreams/{{
  record.log_stream_id }}/lifecycle/deactivate` - kind `custom`; body type `none`; path fields
  `log_stream_id`; required record fields `log_stream_id`; accepted fields `log_stream_id`; risk:
  high: external Okta admin API mutation; approval required.
- `create_api_v1_mappings_mapping_id`: POST `/api/v1/mappings/{{ record.mapping_id }}` - kind
  `create`; body type `json`; path fields `mapping_id`; required record fields `mapping_id`,
  `properties`; accepted fields `mapping_id`, `properties`; risk: medium: external Okta admin API
  mutation; approval required.
- `create_api_v1_meta_schemas_apps_app_id_default`: POST `/api/v1/meta/schemas/apps/{{ record.app_id
  }}/default` - kind `create`; body type `json`; path fields `app_id`; required record fields
  `app_id`; accepted fields `$schema`, `_links`, `app_id`, `created`, `definitions`, `id`,
  `lastUpdated`, `name`, `properties`, `title`, `type`; risk: medium: external Okta admin API
  mutation; approval required.
- `create_api_v1_meta_schemas_group_default`: POST `/api/v1/meta/schemas/group/default` - kind
  `create`; body type `json`; accepted fields `$schema`, `_links`, `created`, `definitions`,
  `description`, `id`, `lastUpdated`, `name`, `properties`, `title`, `type`; risk: medium: external
  Okta admin API mutation; approval required.
- `create_api_v1_meta_schemas_user_linked_objects`: POST `/api/v1/meta/schemas/user/linkedObjects` -
  kind `create`; body type `json`; accepted fields `_links`, `associated`, `primary`; risk: medium:
  external Okta admin API mutation; approval required.
- `delete_api_v1_meta_schemas_user_linked_objects_linked_object_name`: DELETE
  `/api/v1/meta/schemas/user/linkedObjects/{{ record.linked_object_name }}` - kind `delete`; body
  type `none`; path fields `linked_object_name`; required record fields `linked_object_name`;
  accepted fields `linked_object_name`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: high: external Okta admin API mutation; approval required.
- `create_api_v1_meta_schemas_user_schema_id`: POST `/api/v1/meta/schemas/user/{{ record.schema_id
  }}` - kind `create`; body type `json`; path fields `schema_id`; required record fields
  `schema_id`; accepted fields `$schema`, `_links`, `created`, `definitions`, `id`, `lastUpdated`,
  `name`, `properties`, `schema_id`, `title`, `type`; risk: medium: external Okta admin API
  mutation; approval required.
- `create_api_v1_meta_types_user`: POST `/api/v1/meta/types/user` - kind `create`; body type `json`;
  required record fields `name`, `displayName`; accepted fields `_links`, `created`, `createdBy`,
  `default`, `description`, `displayName`, `id`, `lastUpdated`, `lastUpdatedBy`, `name`; risk:
  medium: external Okta admin API mutation; approval required.
- `create_api_v1_meta_types_user_type_id`: POST `/api/v1/meta/types/user/{{ record.type_id }}` -
  kind `create`; body type `json`; path fields `type_id`; required record fields `type_id`; accepted
  fields `description`, `displayName`, `type_id`; risk: medium: external Okta admin API mutation;
  approval required.
- `update_api_v1_meta_types_user_type_id`: PUT `/api/v1/meta/types/user/{{ record.type_id }}` - kind
  `update`; body type `json`; path fields `type_id`; required record fields `type_id`, `name`,
  `displayName`, `description`; accepted fields `description`, `displayName`, `name`, `type_id`;
  risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_meta_types_user_type_id`: DELETE `/api/v1/meta/types/user/{{ record.type_id }}` -
  kind `delete`; body type `none`; path fields `type_id`; required record fields `type_id`; accepted
  fields `type_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: high: external Okta admin API mutation; approval required.
- `create_api_v1_meta_uischemas`: POST `/api/v1/meta/uischemas` - kind `create`; body type `json`;
  accepted fields `uiSchema`; risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_meta_uischemas_id`: PUT `/api/v1/meta/uischemas/{{ record.id }}` - kind `update`;
  body type `json`; path fields `id`; required record fields `id`; accepted fields `id`, `uiSchema`;
  risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_meta_uischemas_id`: DELETE `/api/v1/meta/uischemas/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `id`; required record fields `id`; accepted fields `id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: high:
  external Okta admin API mutation; approval required.
- `create_api_v1_org`: POST `/api/v1/org` - kind `create`; body type `json`; accepted fields
  `_links`, `address1`, `address2`, `city`, `companyName`, `country`, `created`,
  `endUserSupportHelpURL`, `expiresAt`, `id`, `lastUpdated`, `phoneNumber`, `postalCode`, `state`,
  `status`, `subdomain`, `supportPhoneNumber`, `website`; risk: medium: external Okta admin API
  mutation; approval required.
- `update_api_v1_org`: PUT `/api/v1/org` - kind `update`; body type `json`; accepted fields
  `_links`, `address1`, `address2`, `city`, `companyName`, `country`, `created`,
  `endUserSupportHelpURL`, `expiresAt`, `id`, `lastUpdated`, `phoneNumber`, `postalCode`, `state`,
  `status`, `subdomain`, `supportPhoneNumber`, `website`; risk: medium: external Okta admin API
  mutation; approval required.
- `update_api_v1_org_captcha`: PUT `/api/v1/org/captcha` - kind `update`; body type `json`; accepted
  fields `_links`, `captchaId`, `enabledPages`; risk: medium: external Okta admin API mutation;
  approval required.
- `delete_api_v1_org_captcha`: DELETE `/api/v1/org/captcha` - kind `delete`; body type `none`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: high:
  external Okta admin API mutation; approval required.
- `update_api_v1_org_contacts_contact_type`: PUT `/api/v1/org/contacts/{{ record.contact_type }}` -
  kind `update`; body type `json`; path fields `contact_type`; required record fields
  `contact_type`; accepted fields `_links`, `contact_type`, `userId`; risk: medium: external Okta
  admin API mutation; approval required.
- `create_api_v1_org_email_bounces_remove_list`: POST `/api/v1/org/email/bounces/remove-list` - kind
  `create`; body type `json`; accepted fields `emailAddresses`; risk: medium: external Okta admin
  API mutation; approval required.
- `create_api_v1_org_factors_yubikey_token_tokens`: POST `/api/v1/org/factors/yubikey_token/tokens`
  - kind `create`; body type `json`; accepted fields `aesKey`, `privateId`, `publicId`,
  `serialNumber`; risk: medium: external Okta admin API mutation; approval required.
- `create_api_v1_org_org_settings_third_party_admin_setting`: POST
  `/api/v1/org/orgSettings/thirdPartyAdminSetting` - kind `create`; body type `json`; accepted
  fields `thirdPartyAdmin`; risk: medium: external Okta admin API mutation; approval required.
- `create_api_v1_org_preferences_hide_end_user_footer`: POST
  `/api/v1/org/preferences/hideEndUserFooter` - kind `create`; body type `none`; risk: medium:
  external Okta admin API mutation; approval required.
- `create_api_v1_org_preferences_show_end_user_footer`: POST
  `/api/v1/org/preferences/showEndUserFooter` - kind `create`; body type `none`; risk: medium:
  external Okta admin API mutation; approval required.
- `create_api_v1_org_privacy_aerial_grant`: POST `/api/v1/org/privacy/aerial/grant` - kind `create`;
  body type `json`; required record fields `accountId`; accepted fields `accountId`; risk: medium:
  external Okta admin API mutation; approval required.
- `execute_api_v1_org_privacy_aerial_revoke`: POST `/api/v1/org/privacy/aerial/revoke` - kind
  `custom`; body type `json`; required record fields `accountId`; accepted fields `accountId`; risk:
  high: external Okta admin API mutation; approval required.
- `create_api_v1_org_privacy_okta_communication_opt_in`: POST
  `/api/v1/org/privacy/oktaCommunication/optIn` - kind `create`; body type `none`; risk: medium:
  external Okta admin API mutation; approval required.
- `create_api_v1_org_privacy_okta_communication_opt_out`: POST
  `/api/v1/org/privacy/oktaCommunication/optOut` - kind `create`; body type `none`; risk: medium:
  external Okta admin API mutation; approval required.
- `update_api_v1_org_privacy_okta_support_cases_case_number`: PATCH
  `/api/v1/org/privacy/oktaSupport/cases/{{ record.case_number }}` - kind `update`; body type
  `json`; path fields `case_number`; required record fields `case_number`; accepted fields
  `caseNumber`, `case_number`, `impersonation`, `selfAssigned`, `subject`; risk: medium: external
  Okta admin API mutation; approval required.
- `create_api_v1_org_privacy_okta_support_extend`: POST `/api/v1/org/privacy/oktaSupport/extend` -
  kind `create`; body type `none`; risk: medium: external Okta admin API mutation; approval
  required.
- `create_api_v1_org_privacy_okta_support_grant`: POST `/api/v1/org/privacy/oktaSupport/grant` -
  kind `create`; body type `none`; risk: medium: external Okta admin API mutation; approval
  required.
- `execute_api_v1_org_privacy_okta_support_revoke`: POST `/api/v1/org/privacy/oktaSupport/revoke` -
  kind `custom`; body type `none`; risk: high: external Okta admin API mutation; approval required.
- `create_api_v1_org_settings_auto_assign_admin_app_setting`: POST
  `/api/v1/org/settings/autoAssignAdminAppSetting` - kind `create`; body type `json`; accepted
  fields `autoAssignAdminAppSetting`; risk: medium: external Okta admin API mutation; approval
  required.
- `update_api_v1_org_settings_client_privileges_setting`: PUT
  `/api/v1/org/settings/clientPrivilegesSetting` - kind `update`; body type `json`; accepted fields
  `clientPrivilegesSetting`; risk: medium: external Okta admin API mutation; approval required.
- `create_api_v1_orgs`: POST `/api/v1/orgs` - kind `create`; body type `json`; required record
  fields `admin`, `edition`, `name`, `subdomain`; accepted fields `_links`, `admin`, `created`,
  `edition`, `id`, `lastUpdated`, `name`, `settings`, `status`, `subdomain`, `token`, `tokenType`,
  `website`; risk: medium: external Okta admin API mutation; approval required.
- `create_api_v1_policies`: POST `/api/v1/policies` - kind `create`; body type `json`; required
  record fields `name`, `type`; accepted fields `_embedded`, `_links`, `created`, `description`,
  `id`, `lastUpdated`, `name`, `priority`, `status`, `system`, `type`; risk: medium: external Okta
  admin API mutation; approval required.
- `update_api_v1_policies_policy_id`: PUT `/api/v1/policies/{{ record.policy_id }}` - kind `update`;
  body type `json`; path fields `policy_id`; required record fields `policy_id`, `name`, `type`;
  accepted fields `_embedded`, `_links`, `created`, `description`, `id`, `lastUpdated`, `name`,
  `policy_id`, `priority`, `status`, `system`, `type`; risk: medium: external Okta admin API
  mutation; approval required.
- `delete_api_v1_policies_policy_id`: DELETE `/api/v1/policies/{{ record.policy_id }}` - kind
  `delete`; body type `none`; path fields `policy_id`; required record fields `policy_id`; accepted
  fields `policy_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Okta admin API mutation; approval required.
- `create_api_v1_policies_policy_id_clone`: POST `/api/v1/policies/{{ record.policy_id }}/clone` -
  kind `create`; body type `none`; path fields `policy_id`; required record fields `policy_id`;
  accepted fields `policy_id`; risk: medium: external Okta admin API mutation; approval required.
- `execute_api_v1_policies_policy_id_lifecycle_activate`: POST `/api/v1/policies/{{ record.policy_id
  }}/lifecycle/activate` - kind `custom`; body type `none`; path fields `policy_id`; required record
  fields `policy_id`; accepted fields `policy_id`; risk: high: external Okta admin API mutation;
  approval required.
- `execute_api_v1_policies_policy_id_lifecycle_deactivate`: POST `/api/v1/policies/{{
  record.policy_id }}/lifecycle/deactivate` - kind `custom`; body type `none`; path fields
  `policy_id`; required record fields `policy_id`; accepted fields `policy_id`; risk: high: external
  Okta admin API mutation; approval required.
- `create_api_v1_policies_policy_id_mappings`: POST `/api/v1/policies/{{ record.policy_id
  }}/mappings` - kind `create`; body type `json`; path fields `policy_id`; required record fields
  `policy_id`; accepted fields `policy_id`, `resourceId`, `resourceType`; risk: medium: external
  Okta admin API mutation; approval required.
- `delete_api_v1_policies_policy_id_mappings_mapping_id`: DELETE `/api/v1/policies/{{
  record.policy_id }}/mappings/{{ record.mapping_id }}` - kind `delete`; body type `none`; path
  fields `policy_id`, `mapping_id`; required record fields `policy_id`, `mapping_id`; accepted
  fields `mapping_id`, `policy_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: high: external Okta admin API mutation; approval required.
- `create_api_v1_policies_policy_id_rules`: POST `/api/v1/policies/{{ record.policy_id }}/rules` -
  kind `create`; body type `json`; path fields `policy_id`; required record fields `policy_id`;
  accepted fields `_links`, `created`, `id`, `lastUpdated`, `name`, `policy_id`, `priority`,
  `status`, `system`, `type`; risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_policies_policy_id_rules_rule_id`: PUT `/api/v1/policies/{{ record.policy_id
  }}/rules/{{ record.rule_id }}` - kind `update`; body type `json`; path fields `policy_id`,
  `rule_id`; required record fields `policy_id`, `rule_id`; accepted fields `_links`, `created`,
  `id`, `lastUpdated`, `name`, `policy_id`, `priority`, `rule_id`, `status`, `system`, `type`; risk:
  medium: external Okta admin API mutation; approval required.
- `delete_api_v1_policies_policy_id_rules_rule_id`: DELETE `/api/v1/policies/{{ record.policy_id
  }}/rules/{{ record.rule_id }}` - kind `delete`; body type `none`; path fields `policy_id`,
  `rule_id`; required record fields `policy_id`, `rule_id`; accepted fields `policy_id`, `rule_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: high:
  external Okta admin API mutation; approval required.
- `execute_api_v1_policies_policy_id_rules_rule_id_lifecycle_activate`: POST `/api/v1/policies/{{
  record.policy_id }}/rules/{{ record.rule_id }}/lifecycle/activate` - kind `custom`; body type
  `none`; path fields `policy_id`, `rule_id`; required record fields `policy_id`, `rule_id`;
  accepted fields `policy_id`, `rule_id`; risk: high: external Okta admin API mutation; approval
  required.
- `execute_api_v1_policies_policy_id_rules_rule_id_lifecycle_deactivate`: POST `/api/v1/policies/{{
  record.policy_id }}/rules/{{ record.rule_id }}/lifecycle/deactivate` - kind `custom`; body type
  `none`; path fields `policy_id`, `rule_id`; required record fields `policy_id`, `rule_id`;
  accepted fields `policy_id`, `rule_id`; risk: high: external Okta admin API mutation; approval
  required.
- `create_api_v1_principal_rate_limits`: POST `/api/v1/principal-rate-limits` - kind `create`; body
  type `json`; required record fields `principalId`, `principalType`; accepted fields `createdBy`,
  `createdDate`, `defaultConcurrencyPercentage`, `defaultPercentage`, `id`, `lastUpdate`,
  `lastUpdatedBy`, `orgId`, `principalId`, `principalType`; risk: medium: external Okta admin API
  mutation; approval required.
- `update_api_v1_principal_rate_limits_principal_rate_limit_id`: PUT
  `/api/v1/principal-rate-limits/{{ record.principal_rate_limit_id }}` - kind `update`; body type
  `json`; path fields `principal_rate_limit_id`; required record fields `principal_rate_limit_id`,
  `principalId`, `principalType`; accepted fields `createdBy`, `createdDate`,
  `defaultConcurrencyPercentage`, `defaultPercentage`, `id`, `lastUpdate`, `lastUpdatedBy`, `orgId`,
  `principalId`, `principalType`, `principal_rate_limit_id`; risk: medium: external Okta admin API
  mutation; approval required.
- `create_api_v1_push_providers`: POST `/api/v1/push-providers` - kind `create`; body type `json`;
  accepted fields `_links`, `id`, `lastUpdatedDate`, `name`, `providerType`; risk: medium: external
  Okta admin API mutation; approval required.
- `update_api_v1_push_providers_push_provider_id`: PUT `/api/v1/push-providers/{{
  record.push_provider_id }}` - kind `update`; body type `json`; path fields `push_provider_id`;
  required record fields `push_provider_id`; accepted fields `_links`, `id`, `lastUpdatedDate`,
  `name`, `providerType`, `push_provider_id`; risk: medium: external Okta admin API mutation;
  approval required.
- `delete_api_v1_push_providers_push_provider_id`: DELETE `/api/v1/push-providers/{{
  record.push_provider_id }}` - kind `delete`; body type `none`; path fields `push_provider_id`;
  required record fields `push_provider_id`; accepted fields `push_provider_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Okta admin
  API mutation; approval required.
- `update_api_v1_rate_limit_settings_admin_notifications`: PUT
  `/api/v1/rate-limit-settings/admin-notifications` - kind `update`; body type `json`; required
  record fields `notificationsEnabled`; accepted fields `notificationsEnabled`; risk: medium:
  external Okta admin API mutation; approval required.
- `update_api_v1_rate_limit_settings_per_client`: PUT `/api/v1/rate-limit-settings/per-client` -
  kind `update`; body type `json`; required record fields `defaultMode`; accepted fields
  `defaultMode`, `useCaseModeOverrides`; risk: medium: external Okta admin API mutation; approval
  required.
- `update_api_v1_rate_limit_settings_warning_threshold`: PUT
  `/api/v1/rate-limit-settings/warning-threshold` - kind `update`; body type `json`; required record
  fields `warningThreshold`; accepted fields `warningThreshold`; risk: medium: external Okta admin
  API mutation; approval required.
- `create_api_v1_realm_assignments`: POST `/api/v1/realm-assignments` - kind `create`; body type
  `json`; accepted fields `actions`, `conditions`, `name`, `priority`; risk: medium: external Okta
  admin API mutation; approval required.
- `create_api_v1_realm_assignments_operations`: POST `/api/v1/realm-assignments/operations` - kind
  `create`; body type `json`; accepted fields `assignmentId`; risk: medium: external Okta admin API
  mutation; approval required.
- `update_api_v1_realm_assignments_assignment_id`: PUT `/api/v1/realm-assignments/{{
  record.assignment_id }}` - kind `update`; body type `json`; path fields `assignment_id`; required
  record fields `assignment_id`; accepted fields `actions`, `assignment_id`, `conditions`, `name`,
  `priority`; risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_realm_assignments_assignment_id`: DELETE `/api/v1/realm-assignments/{{
  record.assignment_id }}` - kind `delete`; body type `none`; path fields `assignment_id`; required
  record fields `assignment_id`; accepted fields `assignment_id`; missing records treated as success
  for status `404`; confirmation `destructive`; risk: high: external Okta admin API mutation;
  approval required.
- `execute_api_v1_realm_assignments_assignment_id_lifecycle_activate`: POST
  `/api/v1/realm-assignments/{{ record.assignment_id }}/lifecycle/activate` - kind `custom`; body
  type `none`; path fields `assignment_id`; required record fields `assignment_id`; accepted fields
  `assignment_id`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_realm_assignments_assignment_id_lifecycle_deactivate`: POST
  `/api/v1/realm-assignments/{{ record.assignment_id }}/lifecycle/deactivate` - kind `custom`; body
  type `none`; path fields `assignment_id`; required record fields `assignment_id`; accepted fields
  `assignment_id`; risk: high: external Okta admin API mutation; approval required.
- `create_api_v1_realms`: POST `/api/v1/realms` - kind `create`; body type `json`; accepted fields
  `profile`; risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_realms_realm_id`: PUT `/api/v1/realms/{{ record.realm_id }}` - kind `update`; body
  type `json`; path fields `realm_id`; required record fields `realm_id`; accepted fields `profile`,
  `realm_id`; risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_realms_realm_id`: DELETE `/api/v1/realms/{{ record.realm_id }}` - kind `delete`;
  body type `none`; path fields `realm_id`; required record fields `realm_id`; accepted fields
  `realm_id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  high: external Okta admin API mutation; approval required.
- `create_api_v1_roles_role_ref_subscriptions_notification_type_subscribe`: POST `/api/v1/roles/{{
  record.role_ref }}/subscriptions/{{ record.notification_type }}/subscribe` - kind `create`; body
  type `none`; path fields `role_ref`, `notification_type`; required record fields `role_ref`,
  `notification_type`; accepted fields `notification_type`, `role_ref`; risk: medium: external Okta
  admin API mutation; approval required.
- `create_api_v1_roles_role_ref_subscriptions_notification_type_unsubscribe`: POST `/api/v1/roles/{{
  record.role_ref }}/subscriptions/{{ record.notification_type }}/unsubscribe` - kind `create`; body
  type `none`; path fields `role_ref`, `notification_type`; required record fields `role_ref`,
  `notification_type`; accepted fields `notification_type`, `role_ref`; risk: medium: external Okta
  admin API mutation; approval required.
- `create_api_v1_security_events_providers`: POST `/api/v1/security-events-providers` - kind
  `create`; body type `json`; required record fields `name`, `settings`, `type`; accepted fields
  `name`, `settings`, `type`; risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_security_events_providers_security_event_provider_id`: PUT
  `/api/v1/security-events-providers/{{ record.security_event_provider_id }}` - kind `update`; body
  type `json`; path fields `security_event_provider_id`; required record fields
  `security_event_provider_id`, `name`, `settings`, `type`; accepted fields `name`,
  `security_event_provider_id`, `settings`, `type`; risk: medium: external Okta admin API mutation;
  approval required.
- `delete_api_v1_security_events_providers_security_event_provider_id`: DELETE
  `/api/v1/security-events-providers/{{ record.security_event_provider_id }}` - kind `delete`; body
  type `none`; path fields `security_event_provider_id`; required record fields
  `security_event_provider_id`; accepted fields `security_event_provider_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Okta admin
  API mutation; approval required.
- `execute_api_v1_security_events_providers_security_event_provider_id_lifecycle_activate`: POST
  `/api/v1/security-events-providers/{{ record.security_event_provider_id }}/lifecycle/activate` -
  kind `custom`; body type `none`; path fields `security_event_provider_id`; required record fields
  `security_event_provider_id`; accepted fields `security_event_provider_id`; risk: high: external
  Okta admin API mutation; approval required.
- `execute_api_v1_security_events_providers_security_event_provider_id_lifecycle_deactivate`: POST
  `/api/v1/security-events-providers/{{ record.security_event_provider_id }}/lifecycle/deactivate` -
  kind `custom`; body type `none`; path fields `security_event_provider_id`; required record fields
  `security_event_provider_id`; accepted fields `security_event_provider_id`; risk: high: external
  Okta admin API mutation; approval required.
- `delete_api_v1_sessions_session_id`: DELETE `/api/v1/sessions/{{ record.session_id }}` - kind
  `delete`; body type `none`; path fields `session_id`; required record fields `session_id`;
  accepted fields `session_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_sessions_session_id_lifecycle_refresh`: POST `/api/v1/sessions/{{
  record.session_id }}/lifecycle/refresh` - kind `custom`; body type `none`; path fields
  `session_id`; required record fields `session_id`; accepted fields `session_id`; risk: high:
  external Okta admin API mutation; approval required.
- `create_api_v1_ssf_stream`: POST `/api/v1/ssf/stream` - kind `create`; body type `json`; required
  record fields `events_requested`, `delivery`; accepted fields `delivery`, `events_requested`,
  `format`; risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_ssf_stream`: PUT `/api/v1/ssf/stream` - kind `update`; body type `json`; required
  record fields `events_requested`, `delivery`; accepted fields `aud`, `delivery`,
  `events_delivered`, `events_requested`, `events_supported`, `format`, `iss`,
  `min_verification_interval`, `stream_id`; risk: medium: external Okta admin API mutation; approval
  required.
- `update_api_v1_ssf_stream_2`: PATCH `/api/v1/ssf/stream` - kind `update`; body type `json`;
  required record fields `events_requested`, `delivery`; accepted fields `aud`, `delivery`,
  `events_delivered`, `events_requested`, `events_supported`, `format`, `iss`,
  `min_verification_interval`, `stream_id`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_api_v1_ssf_stream`: DELETE `/api/v1/ssf/stream` - kind `delete`; body type `none`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: high: external Okta
  admin API mutation; approval required.
- `create_api_v1_ssf_stream_verification`: POST `/api/v1/ssf/stream/verification` - kind `create`;
  body type `json`; required record fields `stream_id`; accepted fields `state`, `stream_id`; risk:
  medium: external Okta admin API mutation; approval required.
- `create_api_v1_telephony_providers`: POST `/api/v1/telephony-providers` - kind `create`; body type
  `json`; accepted fields `providerAuthToken`, `providerCapability`, `providerName`,
  `providerSettings`, `providerSid`; risk: medium: external Okta admin API mutation; approval
  required.
- `update_api_v1_telephony_providers_custom_telephony_provider_id`: PATCH
  `/api/v1/telephony-providers/{{ record.custom_telephony_provider_id }}` - kind `update`; body type
  `json`; path fields `custom_telephony_provider_id`; required record fields
  `custom_telephony_provider_id`; accepted fields `custom_telephony_provider_id`, `id`,
  `providerAuthToken`, `providerSettings`, `providerSid`; risk: medium: external Okta admin API
  mutation; approval required.
- `delete_api_v1_telephony_providers_custom_telephony_provider_id`: DELETE
  `/api/v1/telephony-providers/{{ record.custom_telephony_provider_id }}` - kind `delete`; body type
  `none`; path fields `custom_telephony_provider_id`; required record fields
  `custom_telephony_provider_id`; accepted fields `custom_telephony_provider_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Okta admin
  API mutation; approval required.
- `execute_api_v1_telephony_providers_custom_telephony_provider_id_lifecycle_activate`: POST
  `/api/v1/telephony-providers/{{ record.custom_telephony_provider_id }}/lifecycle/activate` - kind
  `custom`; body type `none`; path fields `custom_telephony_provider_id`; required record fields
  `custom_telephony_provider_id`; accepted fields `custom_telephony_provider_id`; risk: high:
  external Okta admin API mutation; approval required.
- `execute_api_v1_telephony_providers_custom_telephony_provider_id_lifecycle_deactivate`: POST
  `/api/v1/telephony-providers/{{ record.custom_telephony_provider_id }}/lifecycle/deactivate` -
  kind `custom`; body type `none`; path fields `custom_telephony_provider_id`; required record
  fields `custom_telephony_provider_id`; accepted fields `custom_telephony_provider_id`; risk: high:
  external Okta admin API mutation; approval required.
- `create_api_v1_telephony_providers_custom_telephony_provider_id_set_as_primary`: POST
  `/api/v1/telephony-providers/{{ record.custom_telephony_provider_id }}/setAsPrimary` - kind
  `create`; body type `none`; path fields `custom_telephony_provider_id`; required record fields
  `custom_telephony_provider_id`; accepted fields `custom_telephony_provider_id`; risk: medium:
  external Okta admin API mutation; approval required.
- `create_api_v1_telephony_providers_custom_telephony_provider_id_test`: POST
  `/api/v1/telephony-providers/{{ record.custom_telephony_provider_id }}/test` - kind `create`; body
  type `json`; path fields `custom_telephony_provider_id`; required record fields
  `custom_telephony_provider_id`; accepted fields `countryCodeIso2`, `custom_telephony_provider_id`,
  `factor`, `phoneNumber`; risk: medium: external Okta admin API mutation; approval required.
- `create_api_v1_templates_sms`: POST `/api/v1/templates/sms` - kind `create`; body type `json`;
  accepted fields `created`, `id`, `lastUpdated`, `name`, `template`, `translations`, `type`; risk:
  medium: external Okta admin API mutation; approval required.
- `create_api_v1_templates_sms_template_id`: POST `/api/v1/templates/sms/{{ record.template_id }}` -
  kind `create`; body type `json`; path fields `template_id`; required record fields `template_id`;
  accepted fields `created`, `id`, `lastUpdated`, `name`, `template`, `template_id`, `translations`,
  `type`; risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_templates_sms_template_id`: PUT `/api/v1/templates/sms/{{ record.template_id }}` -
  kind `update`; body type `json`; path fields `template_id`; required record fields `template_id`;
  accepted fields `created`, `id`, `lastUpdated`, `name`, `template`, `template_id`, `translations`,
  `type`; risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_templates_sms_template_id`: DELETE `/api/v1/templates/sms/{{ record.template_id }}`
  - kind `delete`; body type `none`; path fields `template_id`; required record fields
  `template_id`; accepted fields `template_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: high: external Okta admin API mutation; approval required.
- `create_api_v1_threats_configuration`: POST `/api/v1/threats/configuration` - kind `create`; body
  type `json`; required record fields `action`; accepted fields `_links`, `action`, `created`,
  `excludeZones`, `lastUpdated`; risk: medium: external Okta admin API mutation; approval required.
- `create_api_v1_trusted_origins`: POST `/api/v1/trustedOrigins` - kind `create`; body type `json`;
  accepted fields `name`, `origin`, `scopes`; risk: medium: external Okta admin API mutation;
  approval required.
- `update_api_v1_trusted_origins_trusted_origin_id`: PUT `/api/v1/trustedOrigins/{{
  record.trusted_origin_id }}` - kind `update`; body type `json`; path fields `trusted_origin_id`;
  required record fields `trusted_origin_id`; accepted fields `_links`, `created`, `createdBy`,
  `id`, `lastUpdated`, `lastUpdatedBy`, `name`, `origin`, `scopes`, `status`, `trusted_origin_id`;
  risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_trusted_origins_trusted_origin_id`: DELETE `/api/v1/trustedOrigins/{{
  record.trusted_origin_id }}` - kind `delete`; body type `none`; path fields `trusted_origin_id`;
  required record fields `trusted_origin_id`; accepted fields `trusted_origin_id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Okta admin
  API mutation; approval required.
- `execute_api_v1_trusted_origins_trusted_origin_id_lifecycle_activate`: POST
  `/api/v1/trustedOrigins/{{ record.trusted_origin_id }}/lifecycle/activate` - kind `custom`; body
  type `none`; path fields `trusted_origin_id`; required record fields `trusted_origin_id`; accepted
  fields `trusted_origin_id`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_trusted_origins_trusted_origin_id_lifecycle_deactivate`: POST
  `/api/v1/trustedOrigins/{{ record.trusted_origin_id }}/lifecycle/deactivate` - kind `custom`; body
  type `none`; path fields `trusted_origin_id`; required record fields `trusted_origin_id`; accepted
  fields `trusted_origin_id`; risk: high: external Okta admin API mutation; approval required.
- `create_api_v1_users`: POST `/api/v1/users` - kind `create`; body type `json`; required record
  fields `profile`; accepted fields `credentials`, `groupIds`, `profile`, `realmId`, `type`; risk:
  medium: external Okta admin API mutation; approval required.
- `create_api_v1_users_id`: POST `/api/v1/users/{{ record.id }}` - kind `create`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `credentials`, `id`, `profile`,
  `realmId`, `type`; risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_users_id`: PUT `/api/v1/users/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `credentials`, `id`, `profile`,
  `realmId`, `type`; risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_users_id`: DELETE `/api/v1/users/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Okta admin
  API mutation; approval required.
- `execute_api_v1_users_id_lifecycle_activate`: POST `/api/v1/users/{{ record.id
  }}/lifecycle/activate` - kind `custom`; body type `none`; path fields `id`; required record fields
  `id`; accepted fields `id`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_users_id_lifecycle_deactivate`: POST `/api/v1/users/{{ record.id
  }}/lifecycle/deactivate` - kind `custom`; body type `none`; path fields `id`; required record
  fields `id`; accepted fields `id`; risk: high: external Okta admin API mutation; approval
  required.
- `execute_api_v1_users_id_lifecycle_expire_password`: POST `/api/v1/users/{{ record.id
  }}/lifecycle/expire_password` - kind `custom`; body type `none`; path fields `id`; required record
  fields `id`; accepted fields `id`; risk: high: external Okta admin API mutation; approval
  required.
- `execute_api_v1_users_id_lifecycle_expire_password_with_temp_password`: POST `/api/v1/users/{{
  record.id }}/lifecycle/expire_password_with_temp_password` - kind `custom`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; risk: high: external Okta admin
  API mutation; approval required.
- `execute_api_v1_users_id_lifecycle_reactivate`: POST `/api/v1/users/{{ record.id
  }}/lifecycle/reactivate` - kind `custom`; body type `none`; path fields `id`; required record
  fields `id`; accepted fields `id`; risk: high: external Okta admin API mutation; approval
  required.
- `execute_api_v1_users_id_lifecycle_reset_factors`: POST `/api/v1/users/{{ record.id
  }}/lifecycle/reset_factors` - kind `custom`; body type `none`; path fields `id`; required record
  fields `id`; accepted fields `id`; risk: high: external Okta admin API mutation; approval
  required.
- `execute_api_v1_users_id_lifecycle_suspend`: POST `/api/v1/users/{{ record.id
  }}/lifecycle/suspend` - kind `custom`; body type `none`; path fields `id`; required record fields
  `id`; accepted fields `id`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_users_id_lifecycle_unlock`: POST `/api/v1/users/{{ record.id }}/lifecycle/unlock`
  - kind `custom`; body type `none`; path fields `id`; required record fields `id`; accepted fields
  `id`; risk: high: external Okta admin API mutation; approval required.
- `execute_api_v1_users_id_lifecycle_unsuspend`: POST `/api/v1/users/{{ record.id
  }}/lifecycle/unsuspend` - kind `custom`; body type `none`; path fields `id`; required record
  fields `id`; accepted fields `id`; risk: high: external Okta admin API mutation; approval
  required.
- `update_api_v1_users_user_id_or_login_linked_objects_primary_relationship_name_primary_user_id`:
  PUT `/api/v1/users/{{ record.user_id_or_login }}/linkedObjects/{{ record.primary_relationship_name
  }}/{{ record.primary_user_id }}` - kind `update`; body type `none`; path fields
  `user_id_or_login`, `primary_relationship_name`, `primary_user_id`; required record fields
  `user_id_or_login`, `primary_relationship_name`, `primary_user_id`; accepted fields
  `primary_relationship_name`, `primary_user_id`, `user_id_or_login`; risk: medium: external Okta
  admin API mutation; approval required.
- `delete_api_v1_users_user_id_or_login_linked_objects_relationship_name`: DELETE `/api/v1/users/{{
  record.user_id_or_login }}/linkedObjects/{{ record.relationship_name }}` - kind `delete`; body
  type `none`; path fields `user_id_or_login`, `relationship_name`; required record fields
  `user_id_or_login`, `relationship_name`; accepted fields `relationship_name`, `user_id_or_login`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: high:
  external Okta admin API mutation; approval required.
- `create_api_v1_users_user_id_authenticator_enrollments_phone`: POST `/api/v1/users/{{
  record.user_id }}/authenticator-enrollments/phone` - kind `create`; body type `none`; path fields
  `user_id`; required record fields `user_id`; accepted fields `user_id`; risk: medium: external
  Okta admin API mutation; approval required.
- `create_api_v1_users_user_id_authenticator_enrollments_tac`: POST `/api/v1/users/{{ record.user_id
  }}/authenticator-enrollments/tac` - kind `create`; body type `none`; path fields `user_id`;
  required record fields `user_id`; accepted fields `user_id`; risk: medium: external Okta admin API
  mutation; approval required.
- `delete_api_v1_users_user_id_authenticator_enrollments_enrollment_id`: DELETE `/api/v1/users/{{
  record.user_id }}/authenticator-enrollments/{{ record.enrollment_id }}` - kind `delete`; body type
  `none`; path fields `user_id`, `enrollment_id`; required record fields `user_id`, `enrollment_id`;
  accepted fields `enrollment_id`, `user_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: high: external Okta admin API mutation; approval required.
- `update_api_v1_users_user_id_classification`: PUT `/api/v1/users/{{ record.user_id
  }}/classification` - kind `update`; body type `json`; path fields `user_id`; required record
  fields `user_id`; accepted fields `type`, `user_id`; risk: medium: external Okta admin API
  mutation; approval required.
- `delete_api_v1_users_user_id_clients_client_id_grants`: DELETE `/api/v1/users/{{ record.user_id
  }}/clients/{{ record.client_id }}/grants` - kind `delete`; body type `none`; path fields
  `user_id`, `client_id`; required record fields `user_id`, `client_id`; accepted fields
  `client_id`, `user_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Okta admin API mutation; approval required.
- `delete_api_v1_users_user_id_clients_client_id_tokens`: DELETE `/api/v1/users/{{ record.user_id
  }}/clients/{{ record.client_id }}/tokens` - kind `delete`; body type `none`; path fields
  `user_id`, `client_id`; required record fields `user_id`, `client_id`; accepted fields
  `client_id`, `user_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Okta admin API mutation; approval required.
- `delete_api_v1_users_user_id_clients_client_id_tokens_token_id`: DELETE `/api/v1/users/{{
  record.user_id }}/clients/{{ record.client_id }}/tokens/{{ record.token_id }}` - kind `delete`;
  body type `none`; path fields `user_id`, `client_id`, `token_id`; required record fields
  `user_id`, `client_id`, `token_id`; accepted fields `client_id`, `token_id`, `user_id`; missing
  records treated as success for status `404`; confirmation `destructive`; risk: high: external Okta
  admin API mutation; approval required.
- `create_api_v1_users_user_id_credentials_change_password`: POST `/api/v1/users/{{ record.user_id
  }}/credentials/change_password` - kind `create`; body type `json`; path fields `user_id`; required
  record fields `user_id`; accepted fields `newPassword`, `oldPassword`, `revokeSessions`,
  `user_id`; risk: medium: external Okta admin API mutation; approval required.
- `create_api_v1_users_user_id_credentials_change_recovery_question`: POST `/api/v1/users/{{
  record.user_id }}/credentials/change_recovery_question` - kind `create`; body type `json`; path
  fields `user_id`; required record fields `user_id`; accepted fields `password`, `provider`,
  `recovery_question`, `user_id`; risk: medium: external Okta admin API mutation; approval required.
- `create_api_v1_users_user_id_credentials_forgot_password`: POST `/api/v1/users/{{ record.user_id
  }}/credentials/forgot_password` - kind `create`; body type `none`; path fields `user_id`; required
  record fields `user_id`; accepted fields `user_id`; risk: medium: external Okta admin API
  mutation; approval required.
- `create_api_v1_users_user_id_credentials_forgot_password_recovery_question`: POST
  `/api/v1/users/{{ record.user_id }}/credentials/forgot_password_recovery_question` - kind
  `create`; body type `json`; path fields `user_id`; required record fields `user_id`; accepted
  fields `password`, `provider`, `recovery_question`, `user_id`; risk: medium: external Okta admin
  API mutation; approval required.
- `create_api_v1_users_user_id_factors`: POST `/api/v1/users/{{ record.user_id }}/factors` - kind
  `create`; body type `json`; path fields `user_id`; required record fields `user_id`; accepted
  fields `_embedded`, `_links`, `created`, `factorType`, `id`, `lastUpdated`, `profile`, `provider`,
  `status`, `user_id`, `vendorName`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_api_v1_users_user_id_factors_factor_id`: DELETE `/api/v1/users/{{ record.user_id
  }}/factors/{{ record.factor_id }}` - kind `delete`; body type `none`; path fields `user_id`,
  `factor_id`; required record fields `user_id`, `factor_id`; accepted fields `factor_id`,
  `user_id`; missing records treated as success for status `404`; confirmation `destructive`; risk:
  high: external Okta admin API mutation; approval required.
- `execute_api_v1_users_user_id_factors_factor_id_lifecycle_activate`: POST `/api/v1/users/{{
  record.user_id }}/factors/{{ record.factor_id }}/lifecycle/activate` - kind `custom`; body type
  `json`; path fields `user_id`, `factor_id`; required record fields `user_id`, `factor_id`;
  accepted fields `factor_id`, `passCode`, `user_id`; risk: high: external Okta admin API mutation;
  approval required.
- `create_api_v1_users_user_id_factors_factor_id_resend`: POST `/api/v1/users/{{ record.user_id
  }}/factors/{{ record.factor_id }}/resend` - kind `create`; body type `json`; path fields
  `user_id`, `factor_id`; required record fields `user_id`, `factor_id`; accepted fields
  `factorType`, `factor_id`, `user_id`; risk: medium: external Okta admin API mutation; approval
  required.
- `create_api_v1_users_user_id_factors_factor_id_verify`: POST `/api/v1/users/{{ record.user_id
  }}/factors/{{ record.factor_id }}/verify` - kind `create`; body type `json`; path fields
  `user_id`, `factor_id`; required record fields `user_id`, `factor_id`; accepted fields
  `factor_id`, `passCode`, `user_id`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_api_v1_users_user_id_grants`: DELETE `/api/v1/users/{{ record.user_id }}/grants` - kind
  `delete`; body type `none`; path fields `user_id`; required record fields `user_id`; accepted
  fields `user_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: high: external Okta admin API mutation; approval required.
- `delete_api_v1_users_user_id_grants_grant_id`: DELETE `/api/v1/users/{{ record.user_id
  }}/grants/{{ record.grant_id }}` - kind `delete`; body type `none`; path fields `user_id`,
  `grant_id`; required record fields `user_id`, `grant_id`; accepted fields `grant_id`, `user_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: high:
  external Okta admin API mutation; approval required.
- `update_api_v1_users_user_id_risk`: PUT `/api/v1/users/{{ record.user_id }}/risk` - kind `update`;
  body type `json`; path fields `user_id`; required record fields `user_id`, `riskLevel`; accepted
  fields `riskLevel`, `riskReason`, `user_id`; risk: medium: external Okta admin API mutation;
  approval required.
- `create_api_v1_users_user_id_roles`: POST `/api/v1/users/{{ record.user_id }}/roles` - kind
  `create`; body type `json`; path fields `user_id`; required record fields `user_id`, `type`;
  accepted fields `type`, `user_id`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_api_v1_users_user_id_roles_role_assignment_id`: DELETE `/api/v1/users/{{ record.user_id
  }}/roles/{{ record.role_assignment_id }}` - kind `delete`; body type `none`; path fields
  `user_id`, `role_assignment_id`; required record fields `user_id`, `role_assignment_id`; accepted
  fields `role_assignment_id`, `user_id`; missing records treated as success for status `404`;
  confirmation `destructive`; risk: high: external Okta admin API mutation; approval required.
- `update_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps`: PUT `/api/v1/users/{{
  record.user_id }}/roles/{{ record.role_assignment_id }}/targets/catalog/apps` - kind `update`;
  body type `none`; path fields `user_id`, `role_assignment_id`; required record fields `user_id`,
  `role_assignment_id`; accepted fields `role_assignment_id`, `user_id`; risk: medium: external Okta
  admin API mutation; approval required.
- `update_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name`: PUT
  `/api/v1/users/{{ record.user_id }}/roles/{{ record.role_assignment_id }}/targets/catalog/apps/{{
  record.app_name }}` - kind `update`; body type `none`; path fields `user_id`,
  `role_assignment_id`, `app_name`; required record fields `user_id`, `role_assignment_id`,
  `app_name`; accepted fields `app_name`, `role_assignment_id`, `user_id`; risk: medium: external
  Okta admin API mutation; approval required.
- `delete_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name`: DELETE
  `/api/v1/users/{{ record.user_id }}/roles/{{ record.role_assignment_id }}/targets/catalog/apps/{{
  record.app_name }}` - kind `delete`; body type `none`; path fields `user_id`,
  `role_assignment_id`, `app_name`; required record fields `user_id`, `role_assignment_id`,
  `app_name`; accepted fields `app_name`, `role_assignment_id`, `user_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: high: external Okta admin API
  mutation; approval required.
- `update_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id`: PUT
  `/api/v1/users/{{ record.user_id }}/roles/{{ record.role_assignment_id }}/targets/catalog/apps/{{
  record.app_name }}/{{ record.app_id }}` - kind `update`; body type `none`; path fields `user_id`,
  `role_assignment_id`, `app_name`, `app_id`; required record fields `user_id`,
  `role_assignment_id`, `app_name`, `app_id`; accepted fields `app_id`, `app_name`,
  `role_assignment_id`, `user_id`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_api_v1_users_user_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id`:
  DELETE `/api/v1/users/{{ record.user_id }}/roles/{{ record.role_assignment_id
  }}/targets/catalog/apps/{{ record.app_name }}/{{ record.app_id }}` - kind `delete`; body type
  `none`; path fields `user_id`, `role_assignment_id`, `app_name`, `app_id`; required record fields
  `user_id`, `role_assignment_id`, `app_name`, `app_id`; accepted fields `app_id`, `app_name`,
  `role_assignment_id`, `user_id`; missing records treated as success for status `404`; confirmation
  `destructive`; risk: high: external Okta admin API mutation; approval required.
- `update_api_v1_users_user_id_roles_role_assignment_id_targets_groups_group_id`: PUT
  `/api/v1/users/{{ record.user_id }}/roles/{{ record.role_assignment_id }}/targets/groups/{{
  record.group_id }}` - kind `update`; body type `none`; path fields `user_id`,
  `role_assignment_id`, `group_id`; required record fields `user_id`, `role_assignment_id`,
  `group_id`; accepted fields `group_id`, `role_assignment_id`, `user_id`; risk: medium: external
  Okta admin API mutation; approval required.
- `delete_api_v1_users_user_id_roles_role_assignment_id_targets_groups_group_id`: DELETE
  `/api/v1/users/{{ record.user_id }}/roles/{{ record.role_assignment_id }}/targets/groups/{{
  record.group_id }}` - kind `delete`; body type `none`; path fields `user_id`,
  `role_assignment_id`, `group_id`; required record fields `user_id`, `role_assignment_id`,
  `group_id`; accepted fields `group_id`, `role_assignment_id`, `user_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: high: external Okta admin API
  mutation; approval required.
- `delete_api_v1_users_user_id_sessions`: DELETE `/api/v1/users/{{ record.user_id }}/sessions` -
  kind `delete`; body type `none`; path fields `user_id`; required record fields `user_id`; accepted
  fields `user_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: high: external Okta admin API mutation; approval required.
- `create_api_v1_users_user_id_subscriptions_notification_type_subscribe`: POST `/api/v1/users/{{
  record.user_id }}/subscriptions/{{ record.notification_type }}/subscribe` - kind `create`; body
  type `none`; path fields `user_id`, `notification_type`; required record fields `user_id`,
  `notification_type`; accepted fields `notification_type`, `user_id`; risk: medium: external Okta
  admin API mutation; approval required.
- `create_api_v1_users_user_id_subscriptions_notification_type_unsubscribe`: POST `/api/v1/users/{{
  record.user_id }}/subscriptions/{{ record.notification_type }}/unsubscribe` - kind `create`; body
  type `none`; path fields `user_id`, `notification_type`; required record fields `user_id`,
  `notification_type`; accepted fields `notification_type`, `user_id`; risk: medium: external Okta
  admin API mutation; approval required.
- `create_api_v1_zones`: POST `/api/v1/zones` - kind `create`; body type `json`; required record
  fields `name`, `type`; accepted fields `_links`, `created`, `id`, `lastUpdated`, `name`, `status`,
  `system`, `type`, `usage`; risk: medium: external Okta admin API mutation; approval required.
- `update_api_v1_zones_zone_id`: PUT `/api/v1/zones/{{ record.zone_id }}` - kind `update`; body type
  `json`; path fields `zone_id`; required record fields `zone_id`, `name`, `type`; accepted fields
  `_links`, `created`, `id`, `lastUpdated`, `name`, `status`, `system`, `type`, `usage`, `zone_id`;
  risk: medium: external Okta admin API mutation; approval required.
- `delete_api_v1_zones_zone_id`: DELETE `/api/v1/zones/{{ record.zone_id }}` - kind `delete`; body
  type `none`; path fields `zone_id`; required record fields `zone_id`; accepted fields `zone_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: high:
  external Okta admin API mutation; approval required.
- `execute_api_v1_zones_zone_id_lifecycle_activate`: POST `/api/v1/zones/{{ record.zone_id
  }}/lifecycle/activate` - kind `custom`; body type `none`; path fields `zone_id`; required record
  fields `zone_id`; accepted fields `zone_id`; risk: high: external Okta admin API mutation;
  approval required.
- `execute_api_v1_zones_zone_id_lifecycle_deactivate`: POST `/api/v1/zones/{{ record.zone_id
  }}/lifecycle/deactivate` - kind `custom`; body type `none`; path fields `zone_id`; required record
  fields `zone_id`; accepted fields `zone_id`; risk: high: external Okta admin API mutation;
  approval required.
- `update_attack_protection_api_v1_authenticator_settings`: PUT
  `/attack-protection/api/v1/authenticator-settings` - kind `update`; body type `json`; accepted
  fields `verifyKnowledgeSecondWhen2faRequired`; risk: medium: external Okta admin API mutation;
  approval required.
- `update_attack_protection_api_v1_user_lockout_settings`: PUT
  `/attack-protection/api/v1/user-lockout-settings` - kind `update`; body type `json`; accepted
  fields `preventBruteForceLockoutFromUnknownDevices`; risk: medium: external Okta admin API
  mutation; approval required.
- `create_integrations_api_v1_api_services`: POST `/integrations/api/v1/api-services` - kind
  `create`; body type `json`; required record fields `type`, `grantedScopes`; accepted fields
  `grantedScopes`, `properties`, `type`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_integrations_api_v1_api_services_api_service_id`: DELETE
  `/integrations/api/v1/api-services/{{ record.api_service_id }}` - kind `delete`; body type `none`;
  path fields `api_service_id`; required record fields `api_service_id`; accepted fields
  `api_service_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: high: external Okta admin API mutation; approval required.
- `create_integrations_api_v1_api_services_api_service_id_credentials_secrets`: POST
  `/integrations/api/v1/api-services/{{ record.api_service_id }}/credentials/secrets` - kind
  `create`; body type `none`; path fields `api_service_id`; required record fields `api_service_id`;
  accepted fields `api_service_id`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_integrations_api_v1_api_services_api_service_id_credentials_secrets_secret_id`: DELETE
  `/integrations/api/v1/api-services/{{ record.api_service_id }}/credentials/secrets/{{
  record.secret_id }}` - kind `delete`; body type `none`; path fields `api_service_id`, `secret_id`;
  required record fields `api_service_id`, `secret_id`; accepted fields `api_service_id`,
  `secret_id`; missing records treated as success for status `404`; confirmation `destructive`;
  risk: high: external Okta admin API mutation; approval required.
- `execute_integrations_api_v1_api_services_api_service_id_credentials_secrets_secret_id_lifecycle_activate`:
  POST `/integrations/api/v1/api-services/{{ record.api_service_id }}/credentials/secrets/{{
  record.secret_id }}/lifecycle/activate` - kind `custom`; body type `none`; path fields
  `api_service_id`, `secret_id`; required record fields `api_service_id`, `secret_id`; accepted
  fields `api_service_id`, `secret_id`; risk: high: external Okta admin API mutation; approval
  required.
- `execute_integrations_api_v1_api_services_api_service_id_credentials_secrets_secret_id_lifecycle_deactivate`:
  POST `/integrations/api/v1/api-services/{{ record.api_service_id }}/credentials/secrets/{{
  record.secret_id }}/lifecycle/deactivate` - kind `custom`; body type `none`; path fields
  `api_service_id`, `secret_id`; required record fields `api_service_id`, `secret_id`; accepted
  fields `api_service_id`, `secret_id`; risk: high: external Okta admin API mutation; approval
  required.
- `create_oauth2_v1_clients_client_id_roles`: POST `/oauth2/v1/clients/{{ record.client_id }}/roles`
  - kind `create`; body type `json`; path fields `client_id`; required record fields `client_id`,
  `type`; accepted fields `client_id`, `type`; risk: medium: external Okta admin API mutation;
  approval required.
- `delete_oauth2_v1_clients_client_id_roles_role_assignment_id`: DELETE `/oauth2/v1/clients/{{
  record.client_id }}/roles/{{ record.role_assignment_id }}` - kind `delete`; body type `none`; path
  fields `client_id`, `role_assignment_id`; required record fields `client_id`,
  `role_assignment_id`; accepted fields `client_id`, `role_assignment_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: high: external Okta admin API
  mutation; approval required.
- `update_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name`: PUT
  `/oauth2/v1/clients/{{ record.client_id }}/roles/{{ record.role_assignment_id
  }}/targets/catalog/apps/{{ record.app_name }}` - kind `update`; body type `none`; path fields
  `client_id`, `role_assignment_id`, `app_name`; required record fields `client_id`,
  `role_assignment_id`, `app_name`; accepted fields `app_name`, `client_id`, `role_assignment_id`;
  risk: medium: external Okta admin API mutation; approval required.
- `delete_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name`:
  DELETE `/oauth2/v1/clients/{{ record.client_id }}/roles/{{ record.role_assignment_id
  }}/targets/catalog/apps/{{ record.app_name }}` - kind `delete`; body type `none`; path fields
  `client_id`, `role_assignment_id`, `app_name`; required record fields `client_id`,
  `role_assignment_id`, `app_name`; accepted fields `app_name`, `client_id`, `role_assignment_id`;
  missing records treated as success for status `404`; confirmation `destructive`; risk: high:
  external Okta admin API mutation; approval required.
- `update_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id`:
  PUT `/oauth2/v1/clients/{{ record.client_id }}/roles/{{ record.role_assignment_id
  }}/targets/catalog/apps/{{ record.app_name }}/{{ record.app_id }}` - kind `update`; body type
  `none`; path fields `client_id`, `role_assignment_id`, `app_name`, `app_id`; required record
  fields `client_id`, `role_assignment_id`, `app_name`, `app_id`; accepted fields `app_id`,
  `app_name`, `client_id`, `role_assignment_id`; risk: medium: external Okta admin API mutation;
  approval required.
- `delete_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_catalog_apps_app_name_app_id`:
  DELETE `/oauth2/v1/clients/{{ record.client_id }}/roles/{{ record.role_assignment_id
  }}/targets/catalog/apps/{{ record.app_name }}/{{ record.app_id }}` - kind `delete`; body type
  `none`; path fields `client_id`, `role_assignment_id`, `app_name`, `app_id`; required record
  fields `client_id`, `role_assignment_id`, `app_name`, `app_id`; accepted fields `app_id`,
  `app_name`, `client_id`, `role_assignment_id`; missing records treated as success for status
  `404`; confirmation `destructive`; risk: high: external Okta admin API mutation; approval
  required.
- `update_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_groups_group_id`: PUT
  `/oauth2/v1/clients/{{ record.client_id }}/roles/{{ record.role_assignment_id }}/targets/groups/{{
  record.group_id }}` - kind `update`; body type `none`; path fields `client_id`,
  `role_assignment_id`, `group_id`; required record fields `client_id`, `role_assignment_id`,
  `group_id`; accepted fields `client_id`, `group_id`, `role_assignment_id`; risk: medium: external
  Okta admin API mutation; approval required.
- `delete_oauth2_v1_clients_client_id_roles_role_assignment_id_targets_groups_group_id`: DELETE
  `/oauth2/v1/clients/{{ record.client_id }}/roles/{{ record.role_assignment_id }}/targets/groups/{{
  record.group_id }}` - kind `delete`; body type `none`; path fields `client_id`,
  `role_assignment_id`, `group_id`; required record fields `client_id`, `role_assignment_id`,
  `group_id`; accepted fields `client_id`, `group_id`, `role_assignment_id`; missing records treated
  as success for status `404`; confirmation `destructive`; risk: high: external Okta admin API
  mutation; approval required.
- `update_okta_personal_settings_api_v1_edit_feature`: PUT
  `/okta-personal-settings/api/v1/edit-feature` - kind `update`; body type `json`; accepted fields
  `enableEnduserEntryPoints`, `enableExportApps`; risk: medium: external Okta admin API mutation;
  approval required.
- `update_okta_personal_settings_api_v1_export_blocklists`: PUT
  `/okta-personal-settings/api/v1/export-blocklists` - kind `update`; body type `json`; accepted
  fields `domains`; risk: medium: external Okta admin API mutation; approval required.
- `create_privileged_access_api_v1_okta_service_accounts`: POST
  `/privileged-access/api/v1/okta-service-accounts` - kind `create`; body type `json`; required
  record fields `name`, `oktaUserId`; accepted fields `description`, `name`, `oktaUserId`,
  `ownerGroupIds`, `ownerUserIds`; risk: medium: external Okta admin API mutation; approval
  required.
- `update_privileged_access_api_v1_okta_service_accounts_id`: PATCH
  `/privileged-access/api/v1/okta-service-accounts/{{ record.id }}` - kind `update`; body type
  `json`; path fields `id`; required record fields `id`; accepted fields `description`, `id`,
  `name`, `ownerGroupIds`, `ownerUserIds`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_privileged_access_api_v1_okta_service_accounts_id`: DELETE
  `/privileged-access/api/v1/okta-service-accounts/{{ record.id }}` - kind `delete`; body type
  `none`; path fields `id`; required record fields `id`; accepted fields `id`; missing records
  treated as success for status `404`; confirmation `destructive`; risk: high: external Okta admin
  API mutation; approval required.
- `create_privileged_access_api_v1_service_accounts`: POST
  `/privileged-access/api/v1/service-accounts` - kind `create`; body type `json`; required record
  fields `name`, `containerOrn`, `username`, `password`; accepted fields `containerGlobalName`,
  `containerInstanceName`, `containerOrn`, `created`, `description`, `id`, `lastUpdated`, `name`,
  `ownerGroupIds`, `ownerUserIds`, `password`, `status`, `statusDetail`, `username`; risk: medium:
  external Okta admin API mutation; approval required.
- `update_privileged_access_api_v1_service_accounts_id`: PATCH
  `/privileged-access/api/v1/service-accounts/{{ record.id }}` - kind `update`; body type `json`;
  path fields `id`; required record fields `id`; accepted fields `description`, `id`, `name`,
  `ownerGroupIds`, `ownerUserIds`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_privileged_access_api_v1_service_accounts_id`: DELETE
  `/privileged-access/api/v1/service-accounts/{{ record.id }}` - kind `delete`; body type `none`;
  path fields `id`; required record fields `id`; accepted fields `id`; missing records treated as
  success for status `404`; confirmation `destructive`; risk: high: external Okta admin API
  mutation; approval required.
- `execute_webauthn_registration_api_v1_activate`: POST `/webauthn-registration/api/v1/activate` -
  kind `custom`; body type `json`; accepted fields `credResponses`, `fulfillmentProvider`,
  `pinResponseJwe`, `serial`, `userId`, `version`, `yubicoSigningJwks`; risk: high: external Okta
  admin API mutation; approval required.
- `create_webauthn_registration_api_v1_enroll`: POST `/webauthn-registration/api/v1/enroll` - kind
  `create`; body type `json`; accepted fields `enrollmentRpIds`, `fulfillmentProvider`, `userId`,
  `yubicoTransportKeyJWK`; risk: medium: external Okta admin API mutation; approval required.
- `create_webauthn_registration_api_v1_initiate_fulfillment_request`: POST
  `/webauthn-registration/api/v1/initiate-fulfillment-request` - kind `create`; body type `json`;
  accepted fields `fulfillmentData`, `fulfillmentProvider`, `userId`; risk: medium: external Okta
  admin API mutation; approval required.
- `create_webauthn_registration_api_v1_send_pin`: POST `/webauthn-registration/api/v1/send-pin` -
  kind `create`; body type `json`; accepted fields `authenticatorEnrollmentId`,
  `fulfillmentProvider`, `userId`; risk: medium: external Okta admin API mutation; approval
  required.
- `delete_webauthn_registration_api_v1_users_user_id_enrollments_authenticator_enrollment_id`:
  DELETE `/webauthn-registration/api/v1/users/{{ record.user_id }}/enrollments/{{
  record.authenticator_enrollment_id }}` - kind `delete`; body type `none`; path fields `user_id`,
  `authenticator_enrollment_id`; required record fields `user_id`, `authenticator_enrollment_id`;
  accepted fields `authenticator_enrollment_id`, `user_id`; missing records treated as success for
  status `404`; confirmation `destructive`; risk: high: external Okta admin API mutation; approval
  required.
- `create_webauthn_registration_api_v1_users_user_id_enrollments_authenticator_enrollment_id_mark_error`:
  POST `/webauthn-registration/api/v1/users/{{ record.user_id }}/enrollments/{{
  record.authenticator_enrollment_id }}/mark-error` - kind `create`; body type `none`; path fields
  `user_id`, `authenticator_enrollment_id`; required record fields `user_id`,
  `authenticator_enrollment_id`; accepted fields `authenticator_enrollment_id`, `user_id`; risk:
  medium: external Okta admin API mutation; approval required.

## Known limits

- Batch defaults: read_page_size=200, write_batch_size=1.
- API coverage includes 284 stream-backed endpoint group(s), 429 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  binary_payload=7, out_of_scope=12.
