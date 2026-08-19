#!/usr/bin/env node

// Derives connector-local request-field evidence from the provider-owned raw
// OpenAPI matrix. This is deliberately a research artifact, not the final
// citation metadata convention maintained by the separate conventions lane.

const fs = require("fs");
const path = require("path");

const phaseDir = __dirname;
const repoRoot = path.resolve(phaseDir, "../../..");
const bundleDir = path.join(repoRoot, "internal/connectors/defs/recurly");
const matrix = JSON.parse(fs.readFileSync(path.join(phaseDir, "RECURLY-PROVIDER-FIELD-RESEARCH.json"), "utf8"));
const outputPath = path.join(phaseDir, "RECURLY-LOCAL-REQUEST-FIELD-RESEARCH.json");

function readBundleJSON(name) {
  return JSON.parse(fs.readFileSync(path.join(bundleDir, `${name}.json`), "utf8"));
}

function normalizeField(field) {
  return field.replaceAll("[]", "");
}

function configTemplateFields(value, location) {
  if (typeof value !== "string") return [];
  return [...value.matchAll(/\{\{\s*config\.([A-Za-z0-9_]+)\s*\}\}/g)]
    .map((match) => `${location}.${match[1]}`);
}

function schemaFields(schema, prefix, sourcePath, required, fields) {
  if (!schema || typeof schema !== "object") return;
  if (schema.properties && typeof schema.properties === "object") {
    const requiredNames = new Set(Array.isArray(schema.required) ? schema.required : []);
    for (const [name, child] of Object.entries(schema.properties)) {
      const next = prefix ? `${prefix}.${name}` : name;
      const nextSource = `${sourcePath}.properties.${name}`;
      fields.push({
        field: `body.${next}`,
        source_path: nextSource,
        local_required: requiredNames.has(name),
      });
      schemaFields(child, next, nextSource, requiredNames.has(name), fields);
    }
  }
  if (schema.items) {
    schemaFields(schema.items, `${prefix}[]`, `${sourcePath}.items`, required, fields);
  }
}

const api = readBundleJSON("api_surface");
const streams = new Map(readBundleJSON("streams").streams.map((stream) => [stream.name, stream]));
const writes = new Map(readBundleJSON("writes").actions.map((action) => [action.name, action]));
const commands = new Map(readBundleJSON("cli_surface").commands.map((command) => [command.path, command]));
const providerByEndpoint = new Map(matrix.operations.map((operation) => [
  `${operation.method} ${operation.path}`,
  new Map(operation.fields.map((field) => [normalizeField(field.field), field])),
]));

const rows = [];
for (const endpoint of api.endpoints) {
  const endpointKey = `${endpoint.method} ${endpoint.path}`;
  const providerFields = providerByEndpoint.get(endpointKey);
  if (!providerFields) {
    throw new Error(`provider matrix has no endpoint for ${endpointKey}`);
  }

  const localFields = [];
  if (endpoint.covered_by?.stream) {
    const stream = streams.get(endpoint.covered_by.stream);
    if (!stream) throw new Error(`${endpointKey} references missing stream ${endpoint.covered_by.stream}`);
    for (const field of configTemplateFields(stream.path, "path")) {
      localFields.push({field, source_path: `streams.${stream.name}.path`, local_required: false});
    }
    for (const [name, value] of Object.entries(stream.query || {})) {
      const configured = configTemplateFields(value, "query");
      if (configured.length > 0) {
        for (const field of configured) {
          localFields.push({field, source_path: `streams.${stream.name}.query.${name}`, local_required: false});
        }
      } else {
        localFields.push({field: `query.${name}`, source_path: `streams.${stream.name}.query.${name}`, local_required: false});
      }
    }
    for (const local of localFields) {
      local.kind = "stream";
      local.owner = stream.name;
    }
  } else if (endpoint.covered_by?.write) {
    const action = writes.get(endpoint.covered_by.write);
    if (!action) throw new Error(`${endpointKey} references missing write ${endpoint.covered_by.write}`);
    for (const name of action.path_fields || []) {
      localFields.push({
        field: `path.${name}`,
        source_path: `writes.${action.name}.path_fields.${name}`,
        local_required: Array.isArray(action.record_schema?.required) && action.record_schema.required.includes(name),
      });
    }
    const bodyFields = [];
    schemaFields(action.record_schema, "", `writes.${action.name}.record_schema`, false, bodyFields);
    for (const field of bodyFields) {
      const topLevel = field.field.slice("body.".length).split(".")[0].replace(/\[\]$/, "");
      if (!(action.path_fields || []).includes(topLevel)) localFields.push(field);
    }
    for (const local of localFields) {
      local.kind = "write";
      local.owner = action.name;
    }
  } else {
    const commandPaths = [endpoint.covered_by?.direct_read, ...(endpoint.covered_by?.direct_reads || [])].filter(Boolean);
    if (commandPaths.length === 0) throw new Error(`${endpointKey} has no executable coverage`);
    for (const commandPath of commandPaths) {
      const command = commands.get(commandPath);
      if (!command) throw new Error(`${endpointKey} references missing command ${commandPath}`);
      for (const flag of command.flags || []) {
        localFields.push({
          field: flag.maps_to,
          source_path: `cli_surface.commands.${command.path}.flags.${flag.name}`,
          local_required: flag.required === true,
          kind: "command",
          owner: command.path,
        });
      }
    }
  }

  for (const local of localFields) {
    const provider = providerFields.get(normalizeField(local.field));
    if (!provider) {
      throw new Error(`${endpointKey} ${local.owner} field ${local.field} has no provider evidence`);
    }
    rows.push({
      method: endpoint.method,
      path: endpoint.path,
      kind: local.kind,
      owner: local.owner,
      local_field: local.field,
      local_source_path: local.source_path,
      local_required: local.local_required,
      provider_field: provider.field,
      provider_required: provider.required,
      provider_requiredness_rationale: provider.requiredness_rationale,
      source_url: provider.source_url,
      evidence_type: provider.evidence_type,
      evidence_path: provider.evidence_path,
      schema_refs: provider.schema_refs,
      confidence: provider.confidence,
      matching_rule: local.field === provider.field ? "exact" : "array-item notation normalized against provider schema",
    });
  }
}

rows.sort((left, right) => [left.method, left.path, left.kind, left.owner, left.local_field]
  .join("\u0000").localeCompare([right.method, right.path, right.kind, right.owner, right.local_field].join("\u0000")));

const reconciliation = {
  purpose: "Raw provider evidence for every locally exposed Recurly request field; not final citation metadata until the shared citation convention lands.",
  provider: matrix.provider,
  provider_reference_url: matrix.provider_reference_url,
  openapi_source_url: matrix.openapi_source_url,
  openapi_sha256: matrix.openapi_sha256,
  documented_operation_count: matrix.documented_operation_count,
  local_request_field_count: rows.length,
  uncited_or_unmatched_field_count: 0,
  fields: rows,
};

fs.writeFileSync(outputPath, `${JSON.stringify(reconciliation, null, 2)}\n`);
console.log(`wrote ${outputPath} (${rows.length} locally exposed request fields, 0 unmatched)`);
