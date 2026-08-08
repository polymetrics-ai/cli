#!/usr/bin/env node
// Mechanical authoring trace for Zoom's CRC category. It reads the freshly
// fetched provider artifact, extracts each documented request member into a
// closed schema, and appends only CRC operations/commands. It deliberately
// leaves derivable command metadata (endpoint refs, output policy, path maps,
// and REST max_bytes) to `connectorgen surface-sync`.
import fs from "node:fs";

const defs = "internal/connectors/defs/zoom";
const source = fs.readFileSync(".tmp/zoom-crc.md", "utf8");
const sourceURL = "https://developers.zoom.us/docs/api/crc.md";

function fail(message) {
  throw new Error(`author_crc_bundle: ${message}`);
}

function sourceSection(title) {
  const marker = `### ${title}\n`;
  const start = source.indexOf(marker);
  if (start < 0) fail(`missing source section ${JSON.stringify(title)}`);
  const end = source.indexOf("\n### ", start + marker.length);
  return source.slice(start, end < 0 ? source.length : end);
}

function requestSchema(title, expectedProperties, options = {}) {
  const section = sourceSection(title);
  const requestStart = section.indexOf("#### Request Body");
  if (requestStart < 0) fail(`${title} has no request body`);
  const responseStart = section.indexOf("#### Responses", requestStart);
  if (responseStart < 0) fail(`${title} request body has no response boundary`);
  const body = section.slice(requestStart, responseStart);
  const properties = {};
  const required = [];
  const fields = /^- \*\*`([^`]+)`(?: \(required\))?\*\*\n\n\s+`([^`]+)`/gm;
  for (const match of body.matchAll(fields)) {
    const [, name, rawType] = match;
    const typeToken = rawType.trim().toLowerCase();
    let type;
    switch (true) {
      case typeToken.startsWith("string"):
        type = "string";
        break;
      case typeToken.startsWith("integer"):
        type = "integer";
        break;
      case typeToken.startsWith("number"):
        type = "number";
        break;
      case typeToken.startsWith("boolean"):
        type = "boolean";
        break;
      case typeToken.startsWith("array"):
        type = "array";
        break;
      case typeToken.startsWith("object"):
        type = "object";
        break;
      default:
        fail(`${title} field ${name} has unsupported provider type ${JSON.stringify(rawType)}`);
    }
    if (properties[name]) fail(`${title} repeats request field ${name}`);
    properties[name] = type === "array" ? {type, items: {type: "string"}} : {type};
    if (match[0].includes("(required)")) required.push(name);
  }
  if (Object.keys(properties).length !== expectedProperties) {
    fail(`${title} properties=${Object.keys(properties).length}, want ${expectedProperties}`);
  }
  const schema = {type: "object", additionalProperties: false};
  if (required.length > 0) schema.required = required;
  if (options.minProperties) schema.minProperties = options.minProperties;
  schema.properties = properties;
  return schema;
}

function addEnum(schema, field, values) {
  if (!schema.properties[field]) fail(`schema has no ${field} field for enum`);
  schema.properties[field].enum = values;
}

const deviceTypes = [
  "Cisco Series",
  "Polycom Group Series",
  "Polycom Trio",
  "Polycom HDX",
  "Polycom Debut",
  "Lifesize Icon",
  "Dolby",
];
const manageTypes = ["Full Management", "SIP Proxy"];
const logLevels = ["OFF", "ERROR", "WARN", "INFO", "DEBUG"];

const accountSettings = requestSchema("Update Cisco/Polycom Room Account Setting", 82, {minProperties: 1});
const createAPIConnector = requestSchema("Create an API Connector", 4);
const updateAPIConnector = requestSchema("Update an API Connector", 5, {minProperties: 1});
const createManagedRoom = requestSchema("Create a Managed Room", 9);
const updateManagedRoom = requestSchema("Update a Managed Room", 86, {minProperties: 1});
const createRoomTemplate = requestSchema("Create a Room Template", 53);
const updateRoomTemplate = requestSchema("Update a Room Template", 53, {minProperties: 1});

for (const schema of [accountSettings, updateManagedRoom, createRoomTemplate, updateRoomTemplate]) {
  if (schema.properties.device_type) addEnum(schema, "device_type", deviceTypes);
  if (schema.properties.manage_type) addEnum(schema, "manage_type", manageTypes);
}
for (const schema of [createAPIConnector, updateAPIConnector]) addEnum(schema, "log_level", logLevels);

const accountFields = [
  "account_id", "connector_id", "connector_log_level", "firmware_key", "ip_address",
  "time_server", "time_zone",
];
const connectorFields = [
  "connector_id", "description", "managed_rooms", "networks", "private_key",
];
const roomFields = [
  "auth_password", "auth_username", "connector_id", "device_id", "device_password",
  "device_username", "firmware_key", "ip_address", "name", "room_host_key",
  "serial_number", "template_id",
];
const templateFields = [
  "device_type", "name", "template_id", "time_server", "time_zone",
];
const participantFields = ["meeting_id", "participant_identifier_code", "participant_id"];

function read({id, summary, path, scopes, redact, description}) {
  return {
    id,
    kind: "rest_read",
    summary,
    description,
    source_url: sourceURL,
    risk: "high",
    approval: "none",
    output_policy: "json_redacted",
    auth_scopes: scopes,
    sensitive_policy: {redact_fields: redact},
    rest: {method: "GET", path},
  };
}

function write({
  id, summary, path, method, scopes, outputPolicy, mutationClass, description, redact,
  bodySchema, contentType, multipart, destructive = false, secret = false,
}) {
  const rest = {method, path};
  if (contentType) rest.content_type = contentType;
  if (bodySchema) rest.body_schema = bodySchema;
  if (multipart) rest.multipart = multipart;
  const sensitivePolicy = {redact_fields: redact};
  if (secret) {
    sensitivePolicy.input_mode = "env_or_stdin";
    sensitivePolicy.transform = "none";
    sensitivePolicy.approval_mode = "typed_confirmation";
  }
  const operation = {
    id,
    kind: "rest_write",
    summary,
    description,
    source_url: sourceURL,
    risk: "high",
    approval: "plan-preview-confirm-execute",
    output_policy: outputPolicy,
    auth_scopes: scopes,
    mutation_class: mutationClass,
    batchable: false,
    sensitive_policy: sensitivePolicy,
    rest,
  };
  if (destructive) {
    operation.destructive = true;
    operation.confirmation = {kind: "destructive"};
  }
  if (secret) operation.secret_sensitive = true;
  return operation;
}

const operations = [
  read({
    id: "zoom.get_crc_managed_room_account_setting",
    summary: "Get CRC managed room account setting",
    path: "/v2/crc/managed_rooms/account_setting",
    scopes: ["crc_account:read:admin", "crc:read:rooms_account_settings:admin"],
    redact: accountFields,
    description: "Bounded Zoom CRC read for the managed Cisco or Polycom account setting; account and connector values are redacted before output.",
  }),
  write({
    id: "zoom.update_crc_managed_room_account_setting",
    summary: "Update CRC managed room account setting",
    path: "/v2/crc/managed_rooms/account_setting",
    method: "PATCH",
    scopes: ["crc_account:write:admin", "crc:update:rooms_account_settings:admin"],
    outputPolicy: "none",
    mutationClass: "update",
    description: "Approval-gated update of the provider-defined CRC account setting. The documented 204 response is status-only success.",
    redact: accountFields,
    bodySchema: accountSettings,
    contentType: "multipart/form-data",
    multipart: {
      max_bytes: 1048576,
      parts: Object.keys(accountSettings.properties).map((field) => ({name: field, type: "field", field})),
    },
  }),
  read({
    id: "zoom.list_crc_api_connectors",
    summary: "List CRC API Connectors",
    path: "/v2/crc/api_connectors",
    scopes: ["apiconnector:read:admin", "crc:read:list_apiconnectors:admin"],
    redact: connectorFields,
    description: "Bounded Zoom CRC read for enhanced API connectors; connector identifiers and managed-network values are redacted before output.",
  }),
  write({
    id: "zoom.create_crc_api_connector",
    summary: "Create a CRC API Connector",
    path: "/v2/crc/api_connectors",
    method: "POST",
    scopes: ["apiconnector:write:admin", "crc:write:apiconnector:admin"],
    outputPolicy: "json_redacted",
    mutationClass: "create",
    description: "Approval-gated creation of one provider-defined CRC enhanced API Connector.",
    redact: connectorFields,
    bodySchema: createAPIConnector,
    contentType: "application/json",
  }),
  read({
    id: "zoom.get_crc_api_connector",
    summary: "Get a CRC API Connector",
    path: "/v2/crc/api_connectors/{connectorId}",
    scopes: ["apiconnector:read:admin", "crc:read:apiconnector:admin"],
    redact: connectorFields,
    description: "Bounded Zoom CRC read for one enhanced API Connector; identifiers and network details are redacted before output.",
  }),
  write({
    id: "zoom.delete_crc_api_connector",
    summary: "Delete a CRC API Connector",
    path: "/v2/crc/api_connectors/{connectorId}",
    method: "DELETE",
    scopes: ["apiconnector:write:admin", "crc:delete:apiconnector:admin"],
    outputPolicy: "none",
    mutationClass: "delete",
    destructive: true,
    description: "Destructive deletion of one provider-defined enhanced API Connector. The documented 204 response is status-only success.",
    redact: connectorFields,
  }),
  write({
    id: "zoom.update_crc_api_connector",
    summary: "Update a CRC API Connector",
    path: "/v2/crc/api_connectors/{connectorId}",
    method: "PATCH",
    scopes: ["apiconnector:write:admin", "crc:update:apiconnector:admin"],
    outputPolicy: "none",
    mutationClass: "update",
    description: "Approval-gated update of one provider-defined enhanced API Connector. The documented 204 response is status-only success.",
    redact: connectorFields,
    bodySchema: updateAPIConnector,
    contentType: "application/json",
  }),
  read({
    id: "zoom.get_crc_api_connector_private_key",
    summary: "Get a CRC API Connector private key",
    path: "/v2/crc/api_connectors/{connectorId}/private_key",
    scopes: ["apiconnector:read:admin", "crc:read:apiconnector_private_key:admin"],
    redact: ["connector_id", "private_key"],
    description: "Bounded Zoom CRC read for one enhanced API Connector private key. The key is always redacted before output.",
  }),
  write({
    id: "zoom.update_crc_api_connector_private_key",
    summary: "Regenerate a CRC API Connector private key",
    path: "/v2/crc/api_connectors/{connectorId}/private_key",
    method: "PATCH",
    scopes: ["apiconnector:write:admin", "crc:update:apiconnector_private_key:admin"],
    outputPolicy: "json_redacted",
    mutationClass: "secret",
    secret: true,
    description: "Approval-gated private-key regeneration. The returned key is only available through the declared redacted output policy and requires typed confirmation.",
    redact: ["connector_id", "private_key"],
  }),
  read({
    id: "zoom.list_crc_managed_rooms",
    summary: "List CRC managed rooms",
    path: "/v2/crc/managed_rooms",
    scopes: ["crc_rooms:read:admin", "crc:read:list_rooms:admin"],
    redact: roomFields,
    description: "Bounded Zoom CRC read for managed Cisco or Polycom rooms; room and network values are redacted before output.",
  }),
  write({
    id: "zoom.create_crc_managed_room",
    summary: "Create a CRC managed room",
    path: "/v2/crc/managed_rooms",
    method: "POST",
    scopes: ["crc_rooms:write:admin", "crc:write:room:admin"],
    outputPolicy: "json_redacted",
    mutationClass: "create",
    description: "Approval-gated creation of one provider-defined managed Cisco or Polycom room.",
    redact: roomFields,
    bodySchema: createManagedRoom,
    contentType: "application/json",
  }),
  read({
    id: "zoom.get_crc_managed_room",
    summary: "Get a CRC managed room",
    path: "/v2/crc/managed_rooms/{deviceId}",
    scopes: ["crc_rooms:read:admin", "crc:read:room:admin"],
    redact: roomFields,
    description: "Bounded Zoom CRC read for one managed Cisco or Polycom room; room and credential-like values are redacted before output.",
  }),
  write({
    id: "zoom.delete_crc_managed_room",
    summary: "Delete a CRC managed room",
    path: "/v2/crc/managed_rooms/{deviceId}",
    method: "DELETE",
    scopes: ["crc_rooms:write:admin", "crc:delete:room:admin"],
    outputPolicy: "none",
    mutationClass: "delete",
    destructive: true,
    description: "Destructive deletion of one provider-defined managed room. The documented 204 response is status-only success.",
    redact: roomFields,
  }),
  write({
    id: "zoom.update_crc_managed_room",
    summary: "Update a CRC managed room",
    path: "/v2/crc/managed_rooms/{deviceId}",
    method: "PATCH",
    scopes: ["crc_rooms:write:admin", "crc:update:room:admin"],
    outputPolicy: "none",
    mutationClass: "update",
    description: "Approval-gated update of one provider-defined managed room. The documented 204 response is status-only success.",
    redact: roomFields,
    bodySchema: updateManagedRoom,
    contentType: "application/json",
  }),
  read({
    id: "zoom.get_crc_participant_identifier_code",
    summary: "Get a CRC participant identifier code",
    path: "/v2/crc/participant_identifier_code",
    scopes: ["crc:read:admin", "crc:master", "crc:read:participant_identifier_code:admin", "crc:read:participant_identifier_code:master"],
    redact: participantFields,
    description: "Bounded Zoom CRC read for the participant identifier code used to authenticate CRC meetings; all identifier values are redacted before output.",
  }),
  read({
    id: "zoom.list_crc_room_templates",
    summary: "List CRC room templates",
    path: "/v2/crc/room_templates",
    scopes: ["crc_rooms:read:admin", "crc:read:list_rooms_templates:admin"],
    redact: templateFields,
    description: "Bounded Zoom CRC read for Cisco or Polycom room templates; template and deployment values are redacted before output.",
  }),
  write({
    id: "zoom.create_crc_room_template",
    summary: "Create a CRC room template",
    path: "/v2/crc/room_templates",
    method: "POST",
    scopes: ["crc_rooms:write:admin", "crc:write:rooms_template:admin"],
    outputPolicy: "json_redacted",
    mutationClass: "create",
    description: "Approval-gated creation of one provider-defined Cisco or Polycom room template.",
    redact: templateFields,
    bodySchema: createRoomTemplate,
    contentType: "application/json",
  }),
  read({
    id: "zoom.get_crc_room_template",
    summary: "Get a CRC room template",
    path: "/v2/crc/room_templates/{templateId}",
    scopes: ["crc_rooms:read:admin", "crc:read:rooms_template:admin"],
    redact: templateFields,
    description: "Bounded Zoom CRC read for one Cisco or Polycom room template; template and deployment values are redacted before output.",
  }),
  write({
    id: "zoom.delete_crc_room_template",
    summary: "Delete a CRC room template",
    path: "/v2/crc/room_templates/{templateId}",
    method: "DELETE",
    scopes: ["crc_rooms:write:admin", "crc:delete:rooms_template:admin"],
    outputPolicy: "none",
    mutationClass: "delete",
    destructive: true,
    description: "Destructive deletion of one provider-defined Cisco or Polycom room template. The documented 204 response is status-only success.",
    redact: templateFields,
  }),
  write({
    id: "zoom.update_crc_room_template",
    summary: "Update a CRC room template",
    path: "/v2/crc/room_templates/{templateId}",
    method: "PATCH",
    scopes: ["crc_rooms:write:admin", "crc:update:rooms_template:admin"],
    outputPolicy: "none",
    mutationClass: "update",
    description: "Approval-gated update of one provider-defined Cisco or Polycom room template. The documented 204 response is status-only success.",
    redact: templateFields,
    bodySchema: updateRoomTemplate,
    contentType: "application/json",
  }),
];

function pathFlag(name, summary) {
  return {name, type: "string", summary, required: true, allow_empty: false};
}

function rootObjectFlag(name, summary) {
  return {name, type: "json_object", summary, maps_to: "body", required: true};
}

function command({path, summary, intent, operation, sourceCLIPath, flags = [], redact, risk, approval, notes, example}) {
  return {
    path,
    summary,
    intent,
    availability: "implemented",
    operation,
    source_cli_path: sourceCLIPath,
    source_url: sourceURL,
    flags,
    redact_fields: redact,
    risk,
    approval,
    notes,
    examples: [example],
  };
}

const ordinaryApproval = "Requires plan, no-network preview, and explicit single-use approval before execute.";
const destructiveApproval = "Requires plan, no-network preview, explicit single-use approval, and destructive typed confirmation before execute.";
const secretApproval = "Requires plan, no-network preview, explicit single-use approval, typed confirmation, then execute; output remains redacted.";
const cli = [
  command({path: "crc managed-rooms account-setting get", summary: "Get the CRC managed-room account setting.", intent: "direct_read", operation: "zoom.get_crc_managed_room_account_setting", sourceCLIPath: "GET /crc/managed_rooms/account_setting", redact: accountFields, risk: "Reads sensitive CRC account settings.", approval: "Read-only; no write approval is required.", notes: "The provider artifact exposes response-only paging values, so no paging flag is exposed.", example: "pm zoom crc managed-rooms account-setting get --credential zoom-fixture --json"}),
  command({path: "crc managed-rooms account-setting update", summary: "Update the CRC managed-room account setting.", intent: "direct_write", operation: "zoom.update_crc_managed_room_account_setting", sourceCLIPath: "PATCH /crc/managed_rooms/account_setting", flags: [rootObjectFlag("account-settings", "Closed object of documented CRC account-setting fields.")], redact: accountFields, risk: "Changes a provider-defined CRC account setting.", approval: ordinaryApproval, notes: "The named account-settings object is closed to the documented multipart fields; it is not a generic JSON or HTTP write. Its 204 response is status-only success.", example: "pm zoom crc managed-rooms account-setting update --credential zoom-fixture --account-settings '{\"enable_1080p\":true}' --preview --json"}),
  command({path: "crc api-connectors list", summary: "List CRC API Connectors.", intent: "direct_read", operation: "zoom.list_crc_api_connectors", sourceCLIPath: "GET /crc/api_connectors", redact: connectorFields, risk: "Reads sensitive CRC API Connector details.", approval: "Read-only; no write approval is required.", notes: "The provider artifact exposes response-only paging values, so no paging flag is exposed.", example: "pm zoom crc api-connectors list --credential zoom-fixture --json"}),
  command({path: "crc api-connectors create", summary: "Create one CRC API Connector.", intent: "direct_write", operation: "zoom.create_crc_api_connector", sourceCLIPath: "POST /crc/api_connectors", flags: [rootObjectFlag("api-connector", "Closed object of documented CRC API Connector fields.")], redact: connectorFields, risk: "Creates a provider-defined CRC API Connector.", approval: ordinaryApproval, notes: "The named api-connector object is closed to the documented endpoint fields; it is not a generic JSON or HTTP write.", example: "pm zoom crc api-connectors create --credential zoom-fixture --api-connector '{\"description\":\"CRC connector\"}' --preview --json"}),
  command({path: "crc api-connectors get", summary: "Get one CRC API Connector.", intent: "direct_read", operation: "zoom.get_crc_api_connector", sourceCLIPath: "GET /crc/api_connectors/{connectorId}", flags: [pathFlag("connector-id", "CRC API Connector identifier.")], redact: connectorFields, risk: "Reads sensitive CRC API Connector details.", approval: "Read-only; no write approval is required.", notes: "The connector ID is a fixed provider path parameter and is redacted in output.", example: "pm zoom crc api-connectors get --credential zoom-fixture --connector-id connector --json"}),
  command({path: "crc api-connectors delete", summary: "Delete one CRC API Connector.", intent: "direct_write", operation: "zoom.delete_crc_api_connector", sourceCLIPath: "DELETE /crc/api_connectors/{connectorId}", flags: [pathFlag("connector-id", "CRC API Connector identifier to delete.")], redact: connectorFields, risk: "Deletes a provider-defined CRC API Connector.", approval: destructiveApproval, notes: "The documented 204 response is status-only success; no response body is invented.", example: "pm zoom crc api-connectors delete --credential zoom-fixture --connector-id connector --preview --json"}),
  command({path: "crc api-connectors update", summary: "Update one CRC API Connector.", intent: "direct_write", operation: "zoom.update_crc_api_connector", sourceCLIPath: "PATCH /crc/api_connectors/{connectorId}", flags: [pathFlag("connector-id", "CRC API Connector identifier to update."), rootObjectFlag("api-connector", "Closed object of documented CRC API Connector update fields.")], redact: connectorFields, risk: "Changes a provider-defined CRC API Connector.", approval: ordinaryApproval, notes: "The named api-connector object is closed to documented update members; its 204 response is status-only success.", example: "pm zoom crc api-connectors update --credential zoom-fixture --connector-id connector --api-connector '{\"log_level\":\"INFO\"}' --preview --json"}),
  command({path: "crc api-connectors private-key get", summary: "Get one CRC API Connector private key.", intent: "direct_read", operation: "zoom.get_crc_api_connector_private_key", sourceCLIPath: "GET /crc/api_connectors/{connectorId}/private_key", flags: [pathFlag("connector-id", "CRC API Connector identifier.")], redact: ["connector_id", "private_key"], risk: "Reads a private key only through redacted output.", approval: "Read-only; no write approval is required.", notes: "The declared json_redacted policy prevents the private key from reaching CLI output.", example: "pm zoom crc api-connectors private-key get --credential zoom-fixture --connector-id connector --json"}),
  command({path: "crc api-connectors private-key update", summary: "Regenerate one CRC API Connector private key.", intent: "direct_write", operation: "zoom.update_crc_api_connector_private_key", sourceCLIPath: "PATCH /crc/api_connectors/{connectorId}/private_key", flags: [pathFlag("connector-id", "CRC API Connector identifier whose private key is regenerated.")], redact: ["connector_id", "private_key"], risk: "Regenerates a private key and returns it only through redacted output.", approval: secretApproval, notes: "This provider-defined secret-returning action has no request body. Its declared typed confirmation and json_redacted output prevent private key disclosure.", example: "pm zoom crc api-connectors private-key update --credential zoom-fixture --connector-id connector --preview --json"}),
  command({path: "crc managed-rooms list", summary: "List CRC managed rooms.", intent: "direct_read", operation: "zoom.list_crc_managed_rooms", sourceCLIPath: "GET /crc/managed_rooms", redact: roomFields, risk: "Reads sensitive CRC managed-room details.", approval: "Read-only; no write approval is required.", notes: "The provider artifact exposes response-only paging values, so no paging flag is exposed.", example: "pm zoom crc managed-rooms list --credential zoom-fixture --json"}),
  command({path: "crc managed-rooms create", summary: "Create one CRC managed room.", intent: "direct_write", operation: "zoom.create_crc_managed_room", sourceCLIPath: "POST /crc/managed_rooms", flags: [rootObjectFlag("managed-room", "Closed object of documented CRC managed-room fields.")], redact: roomFields, risk: "Creates a provider-defined CRC managed room.", approval: ordinaryApproval, notes: "The named managed-room object is closed to documented endpoint fields; credentials and room identifiers are redacted.", example: "pm zoom crc managed-rooms create --credential zoom-fixture --managed-room '{\"name\":\"CRC room\"}' --preview --json"}),
  command({path: "crc managed-rooms get", summary: "Get one CRC managed room.", intent: "direct_read", operation: "zoom.get_crc_managed_room", sourceCLIPath: "GET /crc/managed_rooms/{deviceId}", flags: [pathFlag("device-id", "CRC managed-room device identifier.")], redact: roomFields, risk: "Reads sensitive CRC managed-room details.", approval: "Read-only; no write approval is required.", notes: "The device ID is a fixed provider path parameter and is redacted in output.", example: "pm zoom crc managed-rooms get --credential zoom-fixture --device-id device --json"}),
  command({path: "crc managed-rooms delete", summary: "Delete one CRC managed room.", intent: "direct_write", operation: "zoom.delete_crc_managed_room", sourceCLIPath: "DELETE /crc/managed_rooms/{deviceId}", flags: [pathFlag("device-id", "CRC managed-room device identifier to delete.")], redact: roomFields, risk: "Deletes a provider-defined CRC managed room.", approval: destructiveApproval, notes: "The documented 204 response is status-only success; no response body is invented.", example: "pm zoom crc managed-rooms delete --credential zoom-fixture --device-id device --preview --json"}),
  command({path: "crc managed-rooms update", summary: "Update one CRC managed room.", intent: "direct_write", operation: "zoom.update_crc_managed_room", sourceCLIPath: "PATCH /crc/managed_rooms/{deviceId}", flags: [pathFlag("device-id", "CRC managed-room device identifier to update."), rootObjectFlag("managed-room", "Closed object of documented CRC managed-room update fields.")], redact: roomFields, risk: "Changes a provider-defined CRC managed room.", approval: ordinaryApproval, notes: "The named managed-room object is closed to documented update members; its 204 response is status-only success.", example: "pm zoom crc managed-rooms update --credential zoom-fixture --device-id device --managed-room '{\"enable_1080p\":true}' --preview --json"}),
  command({path: "crc participant-identifier-code get", summary: "Get the CRC participant identifier code.", intent: "direct_read", operation: "zoom.get_crc_participant_identifier_code", sourceCLIPath: "GET /crc/participant_identifier_code", redact: participantFields, risk: "Reads a CRC meeting authentication identifier.", approval: "Read-only; no write approval is required.", notes: "The participant identifier code is redacted in output.", example: "pm zoom crc participant-identifier-code get --credential zoom-fixture --json"}),
  command({path: "crc room-templates list", summary: "List CRC room templates.", intent: "direct_read", operation: "zoom.list_crc_room_templates", sourceCLIPath: "GET /crc/room_templates", redact: templateFields, risk: "Reads sensitive CRC room-template details.", approval: "Read-only; no write approval is required.", notes: "The provider artifact exposes response-only paging values, so no paging flag is exposed.", example: "pm zoom crc room-templates list --credential zoom-fixture --json"}),
  command({path: "crc room-templates create", summary: "Create one CRC room template.", intent: "direct_write", operation: "zoom.create_crc_room_template", sourceCLIPath: "POST /crc/room_templates", flags: [rootObjectFlag("room-template", "Closed object of documented CRC room-template fields.")], redact: templateFields, risk: "Creates a provider-defined CRC room template.", approval: ordinaryApproval, notes: "The named room-template object is closed to documented endpoint fields; it is not a generic JSON or HTTP write.", example: "pm zoom crc room-templates create --credential zoom-fixture --room-template '{\"device_type\":\"Cisco Series\",\"name\":\"CRC template\"}' --preview --json"}),
  command({path: "crc room-templates get", summary: "Get one CRC room template.", intent: "direct_read", operation: "zoom.get_crc_room_template", sourceCLIPath: "GET /crc/room_templates/{templateId}", flags: [pathFlag("template-id", "CRC room-template identifier.")], redact: templateFields, risk: "Reads sensitive CRC room-template details.", approval: "Read-only; no write approval is required.", notes: "The template ID is a fixed provider path parameter and is redacted in output.", example: "pm zoom crc room-templates get --credential zoom-fixture --template-id template --json"}),
  command({path: "crc room-templates delete", summary: "Delete one CRC room template.", intent: "direct_write", operation: "zoom.delete_crc_room_template", sourceCLIPath: "DELETE /crc/room_templates/{templateId}", flags: [pathFlag("template-id", "CRC room-template identifier to delete.")], redact: templateFields, risk: "Deletes a provider-defined CRC room template.", approval: destructiveApproval, notes: "The documented 204 response is status-only success; no response body is invented.", example: "pm zoom crc room-templates delete --credential zoom-fixture --template-id template --preview --json"}),
  command({path: "crc room-templates update", summary: "Update one CRC room template.", intent: "direct_write", operation: "zoom.update_crc_room_template", sourceCLIPath: "PATCH /crc/room_templates/{templateId}", flags: [pathFlag("template-id", "CRC room-template identifier to update."), rootObjectFlag("room-template", "Closed object of documented CRC room-template update fields.")], redact: templateFields, risk: "Changes a provider-defined CRC room template.", approval: ordinaryApproval, notes: "The named room-template object is closed to documented update members; its 204 response is status-only success.", example: "pm zoom crc room-templates update --credential zoom-fixture --template-id template --room-template '{\"name\":\"Renamed template\"}' --preview --json"}),
];

if (operations.length !== 20 || cli.length !== 20) fail(`authored ${operations.length} operations and ${cli.length} commands, want 20 each`);

function indentedJSON(value, spaces) {
  const indent = " ".repeat(spaces);
  return JSON.stringify(value, null, 2).split("\n").map((line) => indent + line).join("\n");
}

function appendRootArray(file, values) {
  const raw = fs.readFileSync(file, "utf8");
  const marker = "\n  ]\n}\n";
  const at = raw.lastIndexOf(marker);
  if (at < 0) fail(`${file} lacks root array terminator`);
  const output = `${raw.slice(0, at)},\n${values.map((value) => indentedJSON(value, 4)).join(",\n")}${raw.slice(at)}`;
  JSON.parse(output);
  fs.writeFileSync(file, output);
}

function appendCLIGroup(raw, group) {
  const marker = "\n  ],\n  \"global_flags\":";
  const at = raw.indexOf(marker);
  if (at < 0) fail("cli_surface.json lacks groups terminator");
  return `${raw.slice(0, at)},\n${indentedJSON(group, 4)}${raw.slice(at)}`;
}

function appendCLICommands(raw, commands) {
  const marker = "\n  ],\n  \"help_topics\":";
  const at = raw.indexOf(marker);
  if (at < 0) fail("cli_surface.json lacks commands terminator");
  return `${raw.slice(0, at)},\n${commands.map((value) => indentedJSON(value, 4)).join(",\n")}${raw.slice(at)}`;
}

const operationsPath = `${defs}/operations.json`;
const cliPath = `${defs}/cli_surface.json`;
const metadataPath = `${defs}/metadata.json`;
const existingOps = JSON.parse(fs.readFileSync(operationsPath, "utf8"));
const existingCLI = JSON.parse(fs.readFileSync(cliPath, "utf8"));
for (const operation of operations) {
  if (existingOps.operations.some((candidate) => candidate.id === operation.id)) fail(`operation already exists: ${operation.id}`);
}
for (const command of cli) {
  if (existingCLI.commands.some((candidate) => candidate.path === command.path)) fail(`command already exists: ${command.path}`);
}
if (existingCLI.groups.some((group) => group.id === "crc")) fail("CRC group already exists");

appendRootArray(operationsPath, operations);
let cliRaw = fs.readFileSync(cliPath, "utf8");
cliRaw = appendCLIGroup(cliRaw, {id: "crc", title: "Conference Room Connector (CRC)", commands: ["crc"]});
cliRaw = appendCLICommands(cliRaw, cli);
const oldTagline = existingCLI.tagline;
const newTagline = oldTagline.replace("Clips, clinical-note", "Clips, CRC, clinical-note");
if (newTagline === oldTagline) fail("could not add CRC to CLI tagline");
cliRaw = cliRaw.replace(JSON.stringify(oldTagline), JSON.stringify(newTagline));
JSON.parse(cliRaw);
fs.writeFileSync(cliPath, cliRaw);

const metadata = JSON.parse(fs.readFileSync(metadataPath, "utf8"));
metadata.description = metadata.description.replace("and Clips metadata/collaborator/comment/chapter/transfer data", "Clips metadata/collaborator/comment/chapter/transfer data, and CRC account/API Connector/managed-room/room-template data");
metadata.risk.read = metadata.risk.read.replace("and Clips metadata/collaborator/comment/chapter/transfer data plus bounded Clip file download", "Clips metadata/collaborator/comment/chapter/transfer data, CRC account/API Connector/managed-room/room-template/participant-code data, plus bounded Clip file download");
metadata.risk.write = metadata.risk.write.replace("Clips star/collaborator/comment/chapter/share/ownership/file actions", "Clips star/collaborator/comment/chapter/share/ownership/file actions and CRC account/API Connector/managed-room/room-template actions");
metadata.risk.approval = metadata.risk.approval.replace("and Clips collaborator/comment/clip deletion plus ownership transfer", "Clips collaborator/comment/clip deletion plus ownership transfer, CRC API Connector/managed-room/room-template deletion, and CRC private-key regeneration");
fs.writeFileSync(metadataPath, `${JSON.stringify(metadata, null, 2)}\n`);

console.log(JSON.stringify({
  source: sourceURL,
  operations: operations.length,
  commands: cli.length,
  schemas: {
    account_settings: Object.keys(accountSettings.properties).length,
    create_api_connector: Object.keys(createAPIConnector.properties).length,
    update_api_connector: Object.keys(updateAPIConnector.properties).length,
    create_managed_room: Object.keys(createManagedRoom.properties).length,
    update_managed_room: Object.keys(updateManagedRoom.properties).length,
    create_room_template: Object.keys(createRoomTemplate.properties).length,
    update_room_template: Object.keys(updateRoomTemplate.properties).length,
  },
}, null, 2));
