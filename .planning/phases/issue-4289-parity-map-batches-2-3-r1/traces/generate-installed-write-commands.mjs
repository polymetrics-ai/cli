#!/usr/bin/env node

import { readFile, writeFile } from "node:fs/promises";
import { execFile as execFileCallback } from "node:child_process";
import { promisify } from "node:util";
import path from "node:path";

const root = process.cwd();
const phase = path.join(root, ".planning", "phases", "issue-4289-parity-map-batches-2-3-r1");
const connectors = ["grafana", "trello", "slack", "n8n", "google-calendar", "gmail", "twilio", "amazon-sqs", "elasticsearch", "gong", "google-ads", "facebook-marketing", "linkedin-ads", "aircall", "xero", "paypal-transaction", "gocardless", "amazon-seller-partner", "miro"];
const write = process.argv.includes("--write");
const restoreFromHead = process.argv.includes("--restore-from-head");
const execFile = promisify(execFileCallback);

async function readJSON(file, fallback) {
  try {
    return JSON.parse(await readFile(file, "utf8"));
  } catch (error) {
    if (error.code === "ENOENT") return fallback;
    throw error;
  }
}

async function writeJSON(file, value) {
  await writeFile(file, `${JSON.stringify(value, null, 2)}\n`);
}

async function cliSurface(connector, dir) {
  if (!restoreFromHead) return { value: await readJSON(path.join(dir, "cli_surface.json"), { commands: [], groups: [] }), raw: null };
  const repositoryPath = `internal/connectors/defs/${connector}/cli_surface.json`;
  try {
    const { stdout } = await execFile("git", ["show", `HEAD:${repositoryPath}`], { cwd: root });
    return { value: JSON.parse(stdout), raw: stdout };
  } catch (error) {
    // Bundles without a pre-existing command surface are materialized below.
    if (error.code === 128) return { value: await readJSON(path.join(dir, "cli_surface.json"), { commands: [], groups: [] }), raw: null };
    throw error;
  }
}

function explicitTypes(node) {
  if (Array.isArray(node.type)) return node.type;
  return typeof node.type === "string" ? [node.type] : [];
}

function isArray(node) {
  return explicitTypes(node).includes("array");
}

function isObject(node) {
  return Object.keys(node.properties || {}).length > 0 || explicitTypes(node).includes("object") || explicitTypes(node).length === 0;
}

function effectiveTypes(node) {
  const types = explicitTypes(node);
  if (types.length > 0) return types;
  if (isArray(node)) return ["array"];
  if (isObject(node)) return ["object"];
  return ["any"];
}

function requiredPaths(node, prefix = "") {
  const out = [];
  for (const required of node.required || []) {
    const child = (node.properties || {})[required];
    const target = prefix ? `${prefix}.${required}` : required;
    const nested = requiredNodePaths(child, target);
    out.push(...(nested.length > 0 ? nested : [target]));
  }
  return out;
}

function requiredNodePaths(node, prefix) {
  if (!node || typeof node !== "object" || Array.isArray(node)) return [];
  if (isArray(node)) {
    if (!node.items || typeof node.items !== "object" || Array.isArray(node.items)) return [];
    const nested = requiredNodePaths(node.items, `${prefix}.0`);
    return nested.length > 0 ? nested : [prefix];
  }
  return isObject(node) ? requiredPaths(node, prefix) : [];
}

function recordNode(schema, target) {
  let node = schema;
  for (const segment of target.split(".")) {
    if (isArray(node)) {
      if (!/^([0-9]|[1-9][0-9]{0,2})$/.test(segment) || Number(segment) > 128 || !node.items || Array.isArray(node.items)) {
        throw new Error(`invalid array record path ${target}`);
      }
      node = node.items;
      continue;
    }
    if (!isObject(node) || !(segment in (node.properties || {}))) {
      throw new Error(`record field ${target} is not declared`);
    }
    node = node.properties[segment];
  }
  return node;
}

function flagType(node, field) {
  const types = effectiveTypes(node);
  if (types.length !== 1) throw new Error(`ambiguous ${field}: ${types.join(",")}`);
  switch (types[0]) {
    case "string": return "string";
    case "integer": return "integer";
    case "number": return "number";
    case "boolean": return "boolean";
    case "array": return node.items && effectiveTypes(node.items).length === 1 && effectiveTypes(node.items)[0] === "string" ? "string_array" : "json";
    case "object": return "json";
    default: throw new Error(`unsupported ${field}: ${types[0]}`);
  }
}

// This is the connector-local projection of connectorgen's materializedWriteFlags:
// a closed write record admits required scalar flags or one top-level JSON object/array
// field. It deliberately does not create a direct HTTP request body or accept a raw body.
function commandFlags(action) {
  const schema = action.record_schema;
  if (!schema || typeof schema !== "object" || Array.isArray(schema)) throw new Error("record_schema is required");
  const required = new Set();
  for (const field of requiredPaths(schema)) {
    const topLevel = field.split(".")[0];
    const node = recordNode(schema, topLevel);
    required.add(isObject(node) || isArray(node) ? topLevel : field);
  }
  for (const field of action.path_fields || []) required.add(field);
  return [...required].sort().map((field) => {
    if (field.includes(".")) throw new Error(`nested path field ${field} requires a closed top-level record container`);
    return {
      name: field,
      type: flagType(recordNode(schema, field), field),
      summary: `Required ${field.replaceAll("_", " ")} record field.`,
      maps_to: `record.${field}`,
      required: true
    };
  });
}

function refsForAction(rows, action) {
  const ids = new Set(action.provider_operation_binding?.source_ids || []);
  return rows
    .filter((row) => ids.has(row.source?.source_id))
    .map((row) => ({ method: row.api_surface.method, path: row.api_surface.path, source_url: row.source.source_url }))
    .sort((left, right) => left.path.localeCompare(right.path) || left.method.localeCompare(right.method));
}

function commandForAction(connector, action, refs) {
  const commandPath = `${action.name.replaceAll("_", " ")} apply`;
  return {
    path: commandPath,
    summary: `Plan and execute the ${action.name.replaceAll("_", " ")} reverse-ETL action.`,
    intent: "reverse_etl",
    availability: "implemented",
    write: action.name,
    source_url: refs[0].source_url,
    source_cli_path: refs.length === 1 ? `${refs[0].method} ${refs[0].path}` : undefined,
    api_surface: refs.map(({ method, path: endpointPath }) => ({ method, path: endpointPath })),
    flags: commandFlags(action),
    risk: action.risk,
    approval: "requires plan, preview, approval, and execute",
    examples: [`pm ${connector} ${commandPath} --plan <plan-name>`]
  };
}

function partialCommandForAction(connector, action, refs, reason) {
  const commandPath = `${action.name.replaceAll("_", " ")} apply`;
  return {
    path: commandPath,
    summary: `Show the exact ${action.name.replaceAll("_", " ")} action and its declaration-bound typed-input hold.`,
    intent: "reverse_etl",
    availability: "partial",
    write: action.name,
    source_url: refs[0].source_url,
    source_cli_path: refs.length === 1 ? `${refs[0].method} ${refs[0].path}` : undefined,
    api_surface: refs.map(({ method, path: endpointPath }) => ({ method, path: endpointPath })),
    risk: action.risk,
    approval: `Partial command: ${reason}. A declaration-bound scalar-union input capability is required before this exact action can collect every documented value shape; no raw request body is accepted.`,
    examples: [`pm ${connector} ${commandPath} --json`]
  };
}

const report = {
  schema_version: 1,
  mode: write ? "write" : "check",
  connectors: [],
  totals: { existing: 0, generated: 0, source_binding_pending: 0, partial: 0 },
  command_coverage: { actions: 0, implemented: 0, partial: 0, missing: 0 }
};
for (const connector of connectors) {
  const dir = path.join(root, "internal", "connectors", "defs", connector);
  const writes = (await readJSON(path.join(dir, "writes.json"), { actions: [] })).actions;
  if (writes.length === 0) continue;
  const loadedSurface = await cliSurface(connector, dir);
  const surface = loadedSurface.value;
  const metadata = await readJSON(path.join(dir, "metadata.json"), {});
  surface.commands ||= [];
  surface.groups ||= [];
  const dispositionMap = await readJSON(path.join(dir, "sources", `${connector}-declaration-disposition.json`), { ledger_dispositions: [], summary: {} });
  const actions = new Map((dispositionMap.summary?.reverse_etl_eligibility?.action_dispositions || []).map((disposition) => [disposition.action, disposition]));
  const installed = new Map();
  surface.commands.forEach((command, index) => {
    if ((command.availability === "implemented" || command.availability === "partial") && command.intent === "reverse_etl" && command.write) {
      installed.set(command.write, { command, index });
    }
  });
  const commandPaths = new Map(surface.commands.map((command) => [command.path, command.write || command.operation || command.stream || ""]));
  const outcome = { connector, existing: [], generated: [], source_binding_pending: [], partial: [] };
  const additions = [];
  let changed = false;
  for (const action of writes) {
    const disposition = actions.get(action.name);
    const existing = installed.get(action.name);
    if (existing?.command.availability === "implemented") {
      outcome.existing.push({ action: action.name, availability: "implemented" });
      continue;
    }
    if (disposition?.provider_operation_binding?.state !== "source-bound") {
      if (existing) outcome.existing.push({ action: action.name, availability: existing.command.availability });
      outcome.source_binding_pending.push(action.name);
      continue;
    }
    let refs;
    let command;
    try {
      refs = refsForAction(dispositionMap.ledger_dispositions, disposition);
      if (refs.length === 0) throw new Error("source-bound action has no exact disposition rows");
      command = commandForAction(connector, action, refs);
    } catch (error) {
      const sourceRefs = refs || refsForAction(dispositionMap.ledger_dispositions, disposition);
      if (sourceRefs.length === 0) throw new Error(`${connector}: source-bound action ${action.name} has no exact disposition rows`);
      command = partialCommandForAction(connector, action, sourceRefs, error.message);
      outcome.partial.push({ action: action.name, reason: error.message });
    }
    const prior = commandPaths.get(command.path);
    if (prior && prior !== action.name) throw new Error(`${connector}: generated command path ${command.path} collides with ${prior}`);
    if (existing) {
      // Preserve an authored command's public path while promoting a previous
      // partial object/array hold only after the same closed record-schema
      // projection proves every required field representable.
      command.path = existing.command.path;
      surface.commands[existing.index] = command;
      changed = true;
    } else {
      additions.push(command);
    }
    commandPaths.set(command.path, action.name);
    if (command.availability === "implemented") outcome.generated.push(action.name);
  }
  if (write && (additions.length > 0 || !surface.tagline || !surface.usage || changed)) {
    // Match connectorgen's materialized CLI envelope for bundles that previously
    // had no command surface. This does not add an executor or a generic writer.
    const displayName = metadata.display_name || connector;
    if (!surface.tagline) {
      surface.tagline = `Run ${displayName}'s declared streams and reverse-ETL actions.`;
      changed = true;
    }
    if (!surface.usage) {
      surface.usage = `pm ${connector} <command> [flags]`;
      changed = true;
    }
    surface.commands.push(...additions);
    changed ||= additions.length > 0;
    let group = surface.groups.find((candidate) => candidate.id === "write");
    if (!group && additions.length > 0) {
      group = {
        id: "write",
        title: "Reverse ETL writes",
        commands: surface.commands
          .filter((command) => (command.availability === "implemented" || command.availability === "partial") && command.intent === "reverse_etl" && command.write)
          .map((command) => command.path)
      };
      surface.groups.push(group);
      changed = true;
    }
    if (group && additions.length > 0) {
      group.commands = [...new Set([...group.commands, ...additions.map((command) => command.path)])];
      changed = true;
    }
    await writeJSON(path.join(dir, "cli_surface.json"), surface);
  } else if (write && restoreFromHead && loadedSurface.raw !== null) {
    // Preserve the original compact formatting when reconciliation made no
    // declaration change to a tracked connector surface.
    await writeFile(path.join(dir, "cli_surface.json"), loadedSurface.raw);
  }
  const coverageCommands = [...surface.commands, ...additions];
  const availabilityByAction = new Map(coverageCommands
    .filter((command) => command.intent === "reverse_etl" && command.write)
    .map((command) => [command.write, command.availability]));
  outcome.command_coverage = { actions: writes.length, implemented: 0, partial: 0, missing: 0 };
  for (const action of writes) {
    switch (availabilityByAction.get(action.name)) {
      case "implemented": outcome.command_coverage.implemented++; break;
      case "partial": outcome.command_coverage.partial++; break;
      default: outcome.command_coverage.missing++;
    }
  }
  for (const key of ["existing", "generated", "source_binding_pending", "partial"]) report.totals[key] += outcome[key].length;
  for (const key of ["actions", "implemented", "partial", "missing"]) report.command_coverage[key] += outcome.command_coverage[key];
  report.connectors.push(outcome);
}
await writeJSON(path.join(phase, "INSTALLED-WRITE-COMMAND-GENERATION.json"), report);
console.log(JSON.stringify(report.totals));
