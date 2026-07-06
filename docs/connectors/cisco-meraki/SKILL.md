---
name: pm-cisco-meraki
description: Cisco Meraki connector knowledge and safe action guide.
---

# pm-cisco-meraki

## Purpose

Reads and writes Cisco Meraki organizations, networks, devices, admins, licenses, configuration templates, policy objects, branding policies, SAML roles, and organization audit logs from the Meraki Dashboard API v1.

## Icon

- asset: icons/pm-sample.svg
- source: polymetrics
- review_status: polymetrics

## Capabilities

- check=true catalog=true read=true write=true query=false
- Integration type: api

## Authentication

- Use pm credentials add with --from-env or --value-stdin for secret fields.

## Configuration

- base_url
- mode
- page_size
- api_key (secret)

## ETL Streams

- organizations:
  - primary key: id
  - fields: api(), cloud(), id(), licensing(), management(), name(), url()
- organization_networks:
  - primary key: id
  - fields: enrollmentString(), id(), isBoundToConfigTemplate(), name(), notes(), organizationId(), productTypes(), tags(), timeZone(), url()
- organization_devices:
  - primary key: serial
  - fields: address(), firmware(), lat(), lng(), mac(), model(), name(), networkId(), organizationId(), productType(), serial(), tags()
- organization_admins:
  - primary key: id
  - fields: accountStatus(), authenticationMethod(), email(), hasApiKey(), id(), lastActive(), name(), networks(), orgAccess(), organizationId(), tags(), twoFactorAuthEnabled()
- organization_licenses:
  - primary key: id
  - fields: activationDate(), claimDate(), deviceSerial(), durationInDays(), expirationDate(), headLicenseId(), id(), licenseType(), networkId(), orderNumber(), organizationId(), seatCount(), state(), totalDurationInDays()
- organization_config_templates:
  - primary key: id
  - fields: id(), name(), organizationId(), productTypes(), timeZone()
- organization_policy_objects:
  - primary key: id
  - cursor: updatedAt
  - fields: category(), cidr(), createdAt(), groupIds(), id(), name(), networkIds(), organizationId(), type(), updatedAt()
- organization_branding_policies:
  - primary key: organizationId, name
  - fields: adminSettings(), customLogo(), enabled(), helpSettings(), name(), organizationId()
- organization_saml_roles:
  - primary key: id
  - fields: camera(), id(), networks(), orgAccess(), organizationId(), role(), tags()
- organization_configuration_changes:
  - primary key: organizationId, ts, label
  - cursor: ts
  - fields: adminEmail(), adminId(), adminName(), client(), label(), networkId(), networkName(), networkUrl(), newValue(), oldValue(), organizationId(), page(), ssidName(), ssidNumber(), ts()
- organization_api_requests:
  - primary key: organizationId, ts, path, method
  - cursor: ts
  - fields: adminId(), client(), host(), method(), operationId(), organizationId(), path(), queryString(), responseCode(), sourceIp(), ts(), userAgent(), version()

## Sync Modes

- ETL sync modes: full_refresh_append, full_refresh_overwrite, full_refresh_overwrite_deduped

## Reverse ETL Actions

- create_network:
  - endpoint: POST /organizations/{{ record.organizationId }}/networks
  - required fields: organizationId
  - risk: external mutation; creates a new Meraki network under an organization
- update_network:
  - endpoint: PUT /networks/{{ record.id }}
  - required fields: id
  - risk: external mutation; updates an existing Meraki network's name, timezone, tags, or notes
- delete_network:
  - endpoint: DELETE /networks/{{ record.id }}
  - required fields: id
  - risk: irreversible external deletion of a Meraki network and all its devices' configuration; approval required
- update_device:
  - endpoint: PUT /devices/{{ record.serial }}
  - required fields: serial
  - risk: external mutation; updates an existing device's name, tags, physical location, or notes
- create_admin:
  - endpoint: POST /organizations/{{ record.organizationId }}/admins
  - required fields: organizationId
  - risk: external mutation; grants a new dashboard administrator access to this organization
- update_admin:
  - endpoint: PUT /organizations/{{ record.organizationId }}/admins/{{ record.id }}
  - required fields: organizationId, id
  - risk: external mutation; changes an existing dashboard administrator's organization/network/tag access privileges
- delete_admin:
  - endpoint: DELETE /organizations/{{ record.organizationId }}/admins/{{ record.id }}
  - required fields: organizationId, id
  - risk: irreversible external revocation of a dashboard administrator's access to this organization; approval required
- create_config_template:
  - endpoint: POST /organizations/{{ record.organizationId }}/configTemplates
  - required fields: organizationId
  - risk: external mutation; creates a new configuration template under an organization
- update_config_template:
  - endpoint: PUT /organizations/{{ record.organizationId }}/configTemplates/{{ record.id }}
  - required fields: organizationId, id
  - risk: external mutation; renames or retimezones an existing configuration template
- delete_config_template:
  - endpoint: DELETE /organizations/{{ record.organizationId }}/configTemplates/{{ record.id }}
  - required fields: organizationId, id
  - risk: irreversible external deletion of a configuration template; may unbind networks currently attached to it; approval required
- create_policy_object:
  - endpoint: POST /organizations/{{ record.organizationId }}/policyObjects
  - required fields: organizationId
  - risk: external mutation; creates a new network policy object (CIDR/FQDN definition) used by firewall/traffic-shaping rules
- update_policy_object:
  - endpoint: PUT /organizations/{{ record.organizationId }}/policyObjects/{{ record.id }}
  - required fields: organizationId, id
  - risk: external mutation; changes an existing network policy object's definition, potentially altering live firewall/traffic-shaping behavior
- delete_policy_object:
  - endpoint: DELETE /organizations/{{ record.organizationId }}/policyObjects/{{ record.id }}
  - required fields: organizationId, id
  - risk: irreversible external deletion of a network policy object; may break firewall/traffic-shaping rules that reference it; approval required

## Security

- read risk: external Cisco Meraki Dashboard API read of organization, network, device, admin, license, configuration-template, policy-object, branding-policy, SAML-role, and audit-log configuration/state
- write risk: external mutation of Meraki network/device/admin/policy-object/configuration-template configuration; affects live network management state an operator relies on
- Never pass secret values in chat, shell arguments, logs, docs, or JSON output.

## Commands

### Inspect as a manual

```bash
pm connectors inspect cisco-meraki
```

### Inspect as structured JSON

```bash
pm connectors inspect cisco-meraki --json
```

## Agent Rules

- Run pm connectors inspect cisco-meraki before creating credentials or plans.
- Use --json only when the caller needs structured output; use the manual for human-readable guidance.
- Never ask the user to paste secret values into chat.
- For reverse ETL writes, create a plan, show the preview, wait for explicit approval, then run with the approval token.
