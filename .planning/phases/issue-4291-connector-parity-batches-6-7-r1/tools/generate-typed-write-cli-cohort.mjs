#!/usr/bin/env node

// Materialize CLI write plans only from an existing connector-owned typed
// writes.json action and its exact source-disposition/api_surface crosswalk.
// This intentionally has no HTTP/body escape hatch: the action owns method,
// path, record schema, body mapping, confirmation and risk.

import { readFile, writeFile } from "node:fs/promises";
import { resolve } from "node:path";

const phase = resolve(import.meta.dirname, "..");
const root = resolve(phase, "../../..");
const defaultCohort = ["close-com", "zoho-bigin", "braze", "customer-io", "chargebee"];
const sourceBoundOnly = process.argv.includes("--source-bound-only");
const cohort = process.argv.slice(2).filter((arg) => arg !== "--source-bound-only");
if (cohort.length === 0) {
  cohort.push(...defaultCohort);
}

async function readJSON(file) {
  return JSON.parse(await readFile(file, "utf8"));
}

async function writeJSON(file, value) {
  await writeFile(file, `${JSON.stringify(value, null, 2)}\n`);
}

function words(name) {
  return name.replaceAll("_", " ");
}

function actionCommandPath(name) {
  return `${words(name)} apply`;
}

function normalizedActionPath(path) {
  return path.split("?", 1)[0]
    .replace(/\{\{\s*record\.([A-Za-z0-9_]+)\s*\}\}/g, "{$1}")
    .replace(/\{\{\s*config\.([A-Za-z0-9_]+)\s*\}\}/g, "{$1}");
}

function routeSegments(path) {
  return path
    .replace(/\{[^}]+\}/g, "{}")
    .replace(/^\/+|\/+$/g, "")
    .split("/")
    .filter(Boolean);
}

// writes.json paths are connector-relative to config.base_url. Public sources
// can include that base's fixed path prefix (for example /api/v2) and naturally
// give path variables provider names rather than the action record names. The
// disposition's existing covered_by.write relation is the precise mapping;
// this check only refuses a different action-relative static route. A provider
// template variable may resolve to a connector-owned fixed resource selection.
function sameProviderRoute(actionPath, sourcePath) {
  const action = routeSegments(actionPath);
  const source = routeSegments(sourcePath);
  if (source.length < action.length) {
    return false;
  }
  const suffix = source.slice(source.length - action.length);
  return action.every((segment, index) => segment === suffix[index] || suffix[index] === "{}" || segment === "{}");
}

function cliFlagForProperty(name, schema, required) {
  const types = Array.isArray(schema.type) ? schema.type : [schema.type].filter(Boolean);
  if (types.length !== 1) {
    return { flag: null, reason: `record field ${name} has no single declared type` };
  }
  let type = types[0];
  const flag = {
    name,
    summary: `${required ? "Required" : "Optional"} ${words(name)} record field.`,
    maps_to: `record.${name}`,
  };
  if (required) {
    flag.required = true;
  }
  if (type === "string" && Array.isArray(schema.enum) && schema.enum.every((value) => typeof value === "string")) {
    flag.type = "enum";
    flag.values = schema.enum;
    return { flag };
  }
  switch (type) {
    case "string":
    case "integer":
    case "number":
    case "boolean":
      flag.type = type;
      return { flag };
    case "array":
      flag.type = schema.items?.type === "string" ? "string_array" : "json";
      return { flag };
    case "object":
      flag.type = "json";
      return { flag };
    default:
      return { flag: null, reason: `record field ${name} has unsupported type ${JSON.stringify(type)}` };
  }
}

function materializeAction(action, sourceRow) {
  const recordSchema = action.record_schema;
  if (recordSchema?.type !== "object" || !recordSchema.properties || typeof recordSchema.properties !== "object") {
    return { command: null, reason: "writes.json record_schema is not a top-level typed object" };
  }
  const required = new Set([...(recordSchema.required ?? []), ...(action.path_fields ?? [])]);
  const flags = [];
  for (const name of Object.keys(recordSchema.properties).sort()) {
    const result = cliFlagForProperty(name, recordSchema.properties[name], required.has(name));
    if (!result.flag) {
      return { command: null, reason: result.reason };
    }
    flags.push(result.flag);
  }
  for (const pathField of action.path_fields ?? []) {
    if (!recordSchema.properties[pathField]) {
      return { command: null, reason: `required path field ${pathField} is absent from the typed record_schema` };
    }
  }
  if (sourceRow.method !== action.method || !sameProviderRoute(normalizedActionPath(action.path), sourceRow.path)) {
    return {
      command: null,
      reason: `writes.json ${action.method} ${normalizedActionPath(action.path)} does not match source disposition ${sourceRow.method} ${sourceRow.path}`,
    };
  }
  const path = actionCommandPath(action.name);
  return {
    command: {
      path,
      summary: `Plan and execute the typed ${words(action.name)} write action.`,
      intent: "reverse_etl",
      availability: "implemented",
      write: action.name,
      source_cli_path: `${action.method} ${sourceRow.path}`,
      source_url: sourceRow.source.source_url,
      flags,
      risk: action.risk,
      approval: "reverse ETL writes require plan, preview, approval, execute",
      api_surface: [{ method: sourceRow.method, path: sourceRow.path }],
      examples: [`pm ${sourceRow.connector ?? "connector"} ${path} --json`],
      notes: `Pinned source operation ${sourceRow.source.source_id}; writes.json action ${action.name} is the declared method/path/body execution component.`,
    },
  };
}

const commandEvidence = new Map();
const deferredEvidence = new Map();
const connectorDocuments = new Map();

for (const connector of cohort) {
  const definitionRoot = resolve(root, "internal/connectors/defs", connector);
  const cliFile = resolve(definitionRoot, "cli_surface.json");
  const writes = await readJSON(resolve(definitionRoot, "writes.json"));
  const dispositionFile = resolve(definitionRoot, "sources", `${connector}-declaration-disposition.json`);
  const disposition = await readJSON(dispositionFile);
  let cli;
  try {
    cli = await readJSON(cliFile);
  } catch (error) {
    if (error?.code !== "ENOENT") {
      throw error;
    }
    const firstSourceURL = disposition.ledger_dispositions[0]?.source?.source_url;
    if (!firstSourceURL) {
      throw new Error(`${connector}: cannot initialize CLI surface without a pinned source URL`);
    }
    cli = {
      tagline: `Run ${words(connector)} source-backed typed write plans.`,
      usage: `pm ${connector} <command> [flags]`,
      source_cli: {
        name: `${words(connector)} provider API`,
        docs: firstSourceURL,
        reference: "Pinned source-disposition citations and connector-owned typed write contracts.",
        source: "provider_api",
      },
      groups: [],
      global_flags: [
        { name: "credential", type: "string", summary: `Credential name to use for the ${words(connector)} request.` },
        { name: "connection", type: "string", summary: "Alias for --credential." },
        { name: "config", type: "string_array", summary: "Connector config override as key=value; never pass secret values here." },
        { name: "json", type: "boolean", summary: "Emit machine-readable JSON output." },
      ],
      commands: [],
    };
  }
  // A rerun is deterministic: replace only commands that this generator owns.
  cli.commands = cli.commands.filter((command) => !command.notes?.includes("writes.json action") || command.intent !== "reverse_etl");
  const sourceRowsByAction = new Map();
  for (const row of disposition.ledger_dispositions) {
    const action = row.api_surface?.covered_by?.write;
    if (!action) {
      continue;
    }
    const rows = sourceRowsByAction.get(action) ?? [];
    rows.push(row);
    sourceRowsByAction.set(action, rows);
  }
  const generated = [];
  const deferred = [];
  for (const action of writes.actions) {
    const rows = sourceRowsByAction.get(action.name) ?? [];
    if (rows.length === 0 && sourceBoundOnly) {
      continue;
    }
    if (rows.length !== 1) {
      throw new Error(`${connector}:${action.name}: expected exactly one exact source-disposition write row, found ${rows.length}`);
    }
    const sourceRow = rows[0];
    sourceRow.connector = connector;
    const result = materializeAction(action, sourceRow);
    delete sourceRow.connector;
    if (!result.command) {
      deferred.push({ action, sourceRow, reason: result.reason });
      deferredEvidence.set(`${connector}:${sourceRow.source.source_id}`, { action, sourceRow, reason: result.reason });
      continue;
    }
    if (cli.commands.some((command) => command.path === result.command.path)) {
      throw new Error(`${connector}:${action.name}: CLI path ${JSON.stringify(result.command.path)} already exists`);
    }
    generated.push(result.command);
    commandEvidence.set(`${connector}:${sourceRow.source.source_id}`, { action, sourceRow, command: result.command });
  }
  if (!sourceBoundOnly && generated.length + deferred.length !== writes.actions.length) {
    throw new Error(`${connector}: action accounting is incomplete`);
  }
  connectorDocuments.set(connector, { definitionRoot, dispositionFile, disposition, cli, generated, deferred });
}

for (const [connector, document] of connectorDocuments) {
  const { definitionRoot, dispositionFile, disposition, cli, generated, deferred } = document;
  const sourceRowsByID = new Map(disposition.ledger_dispositions.map((row) => [row.source.source_id, row]));
  for (const command of generated) {
    const evidence = [...commandEvidence.values()].find((candidate) => candidate.command === command);
    const row = sourceRowsByID.get(evidence.sourceRow.source.source_id);
    row.declaration = {
      ...row.declaration,
      status: `enabled; typed direct-write action ${evidence.action.name} is reachable through implemented CLI command ${JSON.stringify(command.path)}`,
      command: {
        path: command.path,
        intent: command.intent,
        availability: command.availability,
        write: evidence.action.name,
        source_cli_path: command.source_cli_path,
        source_url: command.source_url,
      },
    };
    row.foundation = {
      state: "present",
      evidence: `internal/connectors/defs/${connector}/cli_surface.json: implemented command ${JSON.stringify(command.path)} binds typed write action ${evidence.action.name} to ${command.source_cli_path}; writes.json remains the fixed request/body contract.`,
      contract: { write: evidence.action.name },
    };
  }
  for (const { action, sourceRow, reason } of deferred) {
    const row = sourceRowsByID.get(sourceRow.source.source_id);
    row.declaration = {
      ...row.declaration,
      status: `enabled typed direct-write action ${action.name}; executable CLI declaration-pending`,
      command: {
        state: "declaration-pending",
        intended_path: actionCommandPath(action.name),
        intent: "reverse_etl",
        write: action.name,
        source_cli_path: `${action.method} ${sourceRow.path}`,
        source_url: sourceRow.source.source_url,
        execution_component: `internal/connectors/defs/${connector}/writes.json action ${action.name}`,
        reason,
      },
    };
  }
  cli.commands.push(...generated);
  cli.commands.sort((left, right) => left.path.localeCompare(right.path));
  cli.groups = (cli.groups ?? []).filter((group) => group.id !== "writes");
  if (generated.length > 0) {
    cli.groups.push({
      id: "writes",
      title: "Typed write plans",
      commands: generated.map((command) => command.path),
    });
  }
  const summary = disposition.summary;
  summary.terminal_commands = cli.commands.length;
  summary.runnable_cli_surface_commands = cli.commands.length;
  summary.endpoint_bound_cli_commands = cli.commands.length;
  disposition.notes = [
    ...(disposition.notes ?? []).filter((note) => !note.includes("2026-08-25 typed write CLI materialization")),
    `2026-08-25 typed write CLI materialization: ${generated.length} source-backed action commands are installed with writes.json-owned input/method/path contracts; ${deferred.length} actions remain explicit CLI declaration-pending because ${deferred.length === 1 ? "its" : "their"} record field contract cannot be represented faithfully by the established command flag vocabulary.`,
  ];
  await writeJSON(resolve(definitionRoot, "cli_surface.json"), cli);
  await writeJSON(dispositionFile, disposition);
}

const evidenceFile = resolve(phase, "OPERATION-SURFACE-EVIDENCE.json");
const evidence = await readJSON(evidenceFile);
let updates = 0;
for (const row of evidence.rows) {
  const generated = commandEvidence.get(`${row.connector}:${row.provider_operation?.source_id}`);
  const deferred = deferredEvidence.get(`${row.connector}:${row.provider_operation?.source_id}`);
  if (generated) {
    const { action, command } = generated;
    row.canonical_mapping.declaration_status = `enabled; typed direct-write action ${action.name} is reachable through implemented CLI command ${JSON.stringify(command.path)}`;
    row.canonical_mapping.contract = {
      ...(row.canonical_mapping.contract ?? {}),
      write: action.name,
      cli_path: command.path,
      source_cli_path: command.source_cli_path,
      execution_component: `internal/connectors/defs/${row.connector}/writes.json action ${action.name}`,
    };
    row.surfaces.executable_cli = {
      state: "generated_binding_present",
      binding: { write: action.name, path: command.path, source_cli_path: command.source_cli_path },
    };
    updates += 1;
  } else if (deferred) {
    row.canonical_mapping.declaration_status = `enabled typed direct-write action ${deferred.action.name}; executable CLI declaration-pending`;
    row.surfaces.executable_cli = {
      state: "declaration_pending_typed_action_cli_contract",
      binding: { write: deferred.action.name, intended_path: actionCommandPath(deferred.action.name) },
      reason: deferred.reason,
    };
    updates += 1;
  }
}
if (updates !== commandEvidence.size + deferredEvidence.size) {
  throw new Error(`operation evidence: updated ${updates} rows, expected ${commandEvidence.size + deferredEvidence.size}`);
}
await writeJSON(evidenceFile, evidence);

console.log(`generated ${commandEvidence.size} typed write CLI commands; ${deferredEvidence.size} remain declaration-pending for ${cohort.join(", ")}`);
