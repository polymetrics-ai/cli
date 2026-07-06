# Overview

Reads and writes Cisco Meraki organizations, networks, devices, admins, licenses, configuration
templates, policy objects, branding policies, SAML roles, and organization audit logs from the
Meraki Dashboard API v1.

Readable streams: `organizations`, `organization_networks`, `organization_devices`,
`organization_admins`, `organization_licenses`, `organization_config_templates`,
`organization_policy_objects`, `organization_branding_policies`, `organization_saml_roles`,
`organization_configuration_changes`, `organization_api_requests`.

Write actions: `create_network`, `update_network`, `delete_network`, `update_device`,
`create_admin`, `update_admin`, `delete_admin`, `create_config_template`, `update_config_template`,
`delete_config_template`, `create_policy_object`, `update_policy_object`, `delete_policy_object`.

Service API documentation: https://developer.cisco.com/meraki/api-v1/.

## Auth setup

Connection fields:

- `api_key` (required, secret, string); Cisco Meraki Dashboard API key. Used only for Bearer auth;
  never logged.
- `base_url` (optional, string); default `https://api.meraki.com/api/v1`; format `uri`; Cisco Meraki
  API base URL override for tests or proxies.
- `mode` (optional, string).
- `page_size` (optional, integer); default `1000`; Records requested per page (perPage query param).
  Meraki caps this at 1000.

Secret fields are redacted in logs and write previews: `api_key`.

Default configuration values: `base_url=https://api.meraki.com/api/v1`, `page_size=1000`.

Authentication behavior:

- Bearer token authentication using `secrets.api_key`.

Requests use the configured `base_url` value after applying defaults.

Connection checks call GET `/organizations` with query `perPage`=`1`.

## Streams notes

Default pagination: follows RFC 5988 Link headers with rel=next.

- `organizations`: GET `/organizations` - records at response root; query `perPage`=`{{
  config.page_size }}`; follows RFC 5988 Link headers with rel=next.
- `organization_networks`: GET `/organizations/{{ fanout.id }}/networks` - records at response root;
  query `perPage`=`{{ config.page_size }}`; follows RFC 5988 Link headers with rel=next; fan-out;
  ids from request `/organizations?perPage={{ config.page_size }}`; id field `id`; id inserted into
  the request path; stamps `organizationId`.
- `organization_devices`: GET `/organizations/{{ fanout.id }}/devices` - records at response root;
  query `perPage`=`{{ config.page_size }}`; follows RFC 5988 Link headers with rel=next; fan-out;
  ids from request `/organizations?perPage={{ config.page_size }}`; id field `id`; id inserted into
  the request path; stamps `organizationId`.
- `organization_admins`: GET `/organizations/{{ fanout.id }}/admins` - records at response root;
  query `perPage`=`{{ config.page_size }}`; follows RFC 5988 Link headers with rel=next; fan-out;
  ids from request `/organizations?perPage={{ config.page_size }}`; id field `id`; id inserted into
  the request path; stamps `organizationId`.
- `organization_licenses`: GET `/organizations/{{ fanout.id }}/licenses` - records at response root;
  query `perPage`=`{{ config.page_size }}`; follows RFC 5988 Link headers with rel=next; fan-out;
  ids from request `/organizations?perPage={{ config.page_size }}`; id field `id`; id inserted into
  the request path; stamps `organizationId`.
- `organization_config_templates`: GET `/organizations/{{ fanout.id }}/configTemplates` - records at
  response root; follows RFC 5988 Link headers with rel=next; fan-out; ids from request
  `/organizations?perPage={{ config.page_size }}`; id field `id`; id inserted into the request path;
  stamps `organizationId`.
- `organization_policy_objects`: GET `/organizations/{{ fanout.id }}/policyObjects` - records at
  response root; query `perPage`=`{{ config.page_size }}`; follows RFC 5988 Link headers with
  rel=next; fan-out; ids from request `/organizations?perPage={{ config.page_size }}`; id field
  `id`; id inserted into the request path; stamps `organizationId`.
- `organization_branding_policies`: GET `/organizations/{{ fanout.id }}/brandingPolicies` - records
  at response root; follows RFC 5988 Link headers with rel=next; fan-out; ids from request
  `/organizations?perPage={{ config.page_size }}`; id field `id`; id inserted into the request path;
  stamps `organizationId`.
- `organization_saml_roles`: GET `/organizations/{{ fanout.id }}/samlRoles` - records at response
  root; follows RFC 5988 Link headers with rel=next; fan-out; ids from request
  `/organizations?perPage={{ config.page_size }}`; id field `id`; id inserted into the request path;
  stamps `organizationId`.
- `organization_configuration_changes`: GET `/organizations/{{ fanout.id }}/configurationChanges` -
  records at response root; query `perPage`=`{{ config.page_size }}`; follows RFC 5988 Link headers
  with rel=next; fan-out; ids from request `/organizations?perPage={{ config.page_size }}`; id field
  `id`; id inserted into the request path; stamps `organizationId`.
- `organization_api_requests`: GET `/organizations/{{ fanout.id }}/apiRequests` - records at
  response root; query `perPage`=`{{ config.page_size }}`; follows RFC 5988 Link headers with
  rel=next; fan-out; ids from request `/organizations?perPage={{ config.page_size }}`; id field
  `id`; id inserted into the request path; stamps `organizationId`.

## Write actions & risks

Overall write risk: external mutation of Meraki
network/device/admin/policy-object/configuration-template configuration; affects live network
management state an operator relies on.

Reverse ETL writes should be planned, previewed, approved, and then executed. Declared actions:

- `create_network`: POST `/organizations/{{ record.organizationId }}/networks` - kind `create`; body
  type `json`; path fields `organizationId`; required record fields `organizationId`, `name`,
  `productTypes`; accepted fields `name`, `notes`, `organizationId`, `productTypes`, `tags`,
  `timeZone`; risk: external mutation; creates a new Meraki network under an organization.
- `update_network`: PUT `/networks/{{ record.id }}` - kind `update`; body type `json`; path fields
  `id`; required record fields `id`; accepted fields `enrollmentString`, `id`, `name`, `notes`,
  `tags`, `timeZone`; risk: external mutation; updates an existing Meraki network's name, timezone,
  tags, or notes.
- `delete_network`: DELETE `/networks/{{ record.id }}` - kind `delete`; body type `none`; path
  fields `id`; required record fields `id`; accepted fields `id`; missing records treated as success
  for status `404`; risk: irreversible external deletion of a Meraki network and all its devices'
  configuration; approval required.
- `update_device`: PUT `/devices/{{ record.serial }}` - kind `update`; body type `json`; path fields
  `serial`; required record fields `serial`; accepted fields `address`, `lat`, `lng`,
  `moveMapMarker`, `name`, `notes`, `serial`, `tags`; risk: external mutation; updates an existing
  device's name, tags, physical location, or notes.
- `create_admin`: POST `/organizations/{{ record.organizationId }}/admins` - kind `create`; body
  type `json`; path fields `organizationId`; required record fields `organizationId`, `email`,
  `name`, `orgAccess`; accepted fields `email`, `name`, `networks`, `orgAccess`, `organizationId`,
  `tags`; risk: external mutation; grants a new dashboard administrator access to this organization.
- `update_admin`: PUT `/organizations/{{ record.organizationId }}/admins/{{ record.id }}` - kind
  `update`; body type `json`; path fields `organizationId`, `id`; required record fields
  `organizationId`, `id`; accepted fields `id`, `name`, `networks`, `orgAccess`, `organizationId`,
  `tags`; risk: external mutation; changes an existing dashboard administrator's
  organization/network/tag access privileges.
- `delete_admin`: DELETE `/organizations/{{ record.organizationId }}/admins/{{ record.id }}` - kind
  `delete`; body type `none`; path fields `organizationId`, `id`; required record fields
  `organizationId`, `id`; accepted fields `id`, `organizationId`; missing records treated as success
  for status `404`; risk: irreversible external revocation of a dashboard administrator's access to
  this organization; approval required.
- `create_config_template`: POST `/organizations/{{ record.organizationId }}/configTemplates` - kind
  `create`; body type `json`; path fields `organizationId`; required record fields `organizationId`,
  `name`; accepted fields `copyFromNetworkId`, `name`, `organizationId`, `timeZone`; risk: external
  mutation; creates a new configuration template under an organization.
- `update_config_template`: PUT `/organizations/{{ record.organizationId }}/configTemplates/{{
  record.id }}` - kind `update`; body type `json`; path fields `organizationId`, `id`; required
  record fields `organizationId`, `id`; accepted fields `id`, `name`, `organizationId`, `timeZone`;
  risk: external mutation; renames or retimezones an existing configuration template.
- `delete_config_template`: DELETE `/organizations/{{ record.organizationId }}/configTemplates/{{
  record.id }}` - kind `delete`; body type `none`; path fields `organizationId`, `id`; required
  record fields `organizationId`, `id`; accepted fields `id`, `organizationId`; missing records
  treated as success for status `404`; risk: irreversible external deletion of a configuration
  template; may unbind networks currently attached to it; approval required.
- `create_policy_object`: POST `/organizations/{{ record.organizationId }}/policyObjects` - kind
  `create`; body type `json`; path fields `organizationId`; required record fields `organizationId`,
  `name`, `category`, `type`; accepted fields `category`, `cidr`, `fqdn`, `groupIds`, `ip`, `mask`,
  `name`, `organizationId`, `type`; risk: external mutation; creates a new network policy object
  (CIDR/FQDN definition) used by firewall/traffic-shaping rules.
- `update_policy_object`: PUT `/organizations/{{ record.organizationId }}/policyObjects/{{ record.id
  }}` - kind `update`; body type `json`; path fields `organizationId`, `id`; required record fields
  `organizationId`, `id`; accepted fields `cidr`, `fqdn`, `groupIds`, `id`, `ip`, `mask`, `name`,
  `organizationId`; risk: external mutation; changes an existing network policy object's definition,
  potentially altering live firewall/traffic-shaping behavior.
- `delete_policy_object`: DELETE `/organizations/{{ record.organizationId }}/policyObjects/{{
  record.id }}` - kind `delete`; body type `none`; path fields `organizationId`, `id`; required
  record fields `organizationId`, `id`; accepted fields `id`, `organizationId`; missing records
  treated as success for status `404`; risk: irreversible external deletion of a network policy
  object; may break firewall/traffic-shaping rules that reference it; approval required.

## Known limits

- Batch defaults: read_page_size=1000.
- API coverage includes 11 stream-backed endpoint group(s), 13 write-backed endpoint group(s).
- Other documented endpoints are not exposed by this connector where they are classified as
  destructive_admin=2, duplicate_of=9, non_data_endpoint=1, out_of_scope=14,
  requires_elevated_scope=7.
