#!/usr/bin/env node

// Reconciles the seven required #4292 surfaces from connector-owned JSON only.
// It is deliberately a static/no-I/O assertion: provider source locks provide
// the operation inventory, and a transport declaration is admitted only when
// an existing typed action's required record fields are present in one existing
// source-stream schema. It never invents a provider request/body schema.

import { readFile, writeFile } from "node:fs/promises";
import { existsSync } from "node:fs";
import { resolve } from "node:path";

const root = resolve(import.meta.dirname, "../../../..");
const phase = resolve(root, ".planning/phases/issue-4292-parity-batches-8-10-r1");
const defs = resolve(root, "internal/connectors/defs");
const ledgerPath = resolve(phase, "SEVEN-SURFACE-LEDGER.json");
const summaryPath = resolve(phase, "SEVEN-SURFACE-SUMMARY.md");

const connectors = [
  "brex", "zoho-books", "testrail", "amplitude", "posthog", "metabase", "dbt", "looker", "mode", "dremio",
  "coda", "clickup-api", "calendly", "greenhouse", "lever-hiring", "ashby", "workable", "recruitee", "hibob", "factorial",
  "datadog", "pagerduty", "auth0", "okta", "firehydrant", "adobe-commerce-magento", "commercetools", "recharge", "docuseal", "eventbrite",
];

const batchFor = (name) => {
  const index = connectors.indexOf(name);
  if (index < 0) throw new Error(`unknown assigned connector ${name}`);
  return 8 + Math.floor(index / 10);
};
const file = (connector, name) => resolve(defs, connector, name);
const readJSON = async (path, fallback = undefined) => {
  try { return JSON.parse(await readFile(path, "utf8")); }
  catch (error) {
    if (fallback !== undefined && error.code === "ENOENT") return fallback;
    throw error;
  }
};
const pretty = (value) => `${JSON.stringify(value, null, 2)}\n`;
const assert = (condition, message) => { if (!condition) throw new Error(message); };
const array = (value) => Array.isArray(value) ? value : [];
const actionFields = (action) => Object.keys(action.record_schema?.properties ?? {});
const requiredFields = (action) => array(action.record_schema?.required);
const actionCLICommands = (surface, action) => array(surface.commands).filter((command) => command.write === action.name);

function sourceSchema(stream, schemas) {
  return schemas.get(stream.name) ?? { properties: {} };
}

function candidateFor(action, streams, schemas) {
  if (action.transport_binding) {
    return { state: "semantic-exclusion", reason: "action selects a different closed destination adapter through transport_binding" };
  }
  if (action.record_schema?.type !== "object") {
    return { state: "semantic-exclusion", reason: "closed declarative destination accepts object records only" };
  }
  const required = requiredFields(action);
  const fields = required.length > 0 ? required : actionFields(action);
  if (fields.length === 0) {
    return { state: "semantic-exclusion", reason: "action has no typed record field available for the required input_fields mapping" };
  }
  // The foundation validates names in the action-facing mapping as lowercase,
  // underscore/dash identifiers. A provider's case-preserving action input
  // cannot be changed here: the executor uses the input name to look up the
  // action's actual typed property. Keep this as an explicit foundation gap
  // rather than emitting a declaration the real loader rejects.
  if (fields.some((field) => /[A-Z]/.test(field))) {
    return {
      state: "foundation-gap",
      reason: `closed destination mapping rejects the provider's case-preserving input(s): ${fields.filter((field) => /[A-Z]/.test(field)).join(", ")}`,
      refusing_location: "internal/connectors/sync_transport.go:673",
      minimal_change: "allow case-preserving concrete input_fields identifiers so action property names remain exact",
    };
  }
  for (const stream of streams) {
    const properties = sourceSchema(stream, schemas).properties ?? {};
    if (fields.every((name) => Object.hasOwn(properties, name))) {
      return {
        state: "representable",
        stream: stream.name,
        inputs: fields.map((field) => ({ input: field, field })),
        required,
      };
    }
  }
  return { state: "semantic-exclusion", reason: `no executable source-stream schema provides every typed action input: ${fields.join(", ")}` };
}

function declaration(connector, streams, selected, candidates) {
  const document = {
    schema_version: 1,
    source_transport: {
      executor: { family: "declarative_api", id: "declarative_stream_source" },
      eligible_streams: streams.map((stream) => stream.name),
      modes: ["full_append"],
      delivery: { idempotency: "at_least_once", ordering: "source_ordered", deletes: "not_available" },
      conformance: { suite: "declarative_stream_transport_static", run_id: `${connector}_source_definition_v1` },
    },
  };
  if (selected) {
    document.destination_transport = {
      executor: { family: "declarative_api", id: "declarative_typed_destination" },
      // This admits every exact record-driven action. The current closed
      // destination declaration has one selected action per mode; actions not
      // selected below are recorded in the ledger as a foundation multiplicity
      // dependency rather than silently treated as unavailable.
      eligible_actions: candidates.map(({ action }) => action.name),
      modes: ["full_append"],
      delivery: { idempotency: "keyed", ordering: "source_ordered", deletes: "not_available" },
      conformance: { suite: "declarative_typed_destination_static", run_id: `${connector}_${selected.name}_full_append_v1` },
      acknowledgement: "durable_warehouse",
      apply_strategies: [{ mode: "full_append", strategy: "append", action: selected.name }],
      source_bindings: [{
        executor: { family: "declarative_api", id: "declarative_stream_source" },
        eligible_streams: [selected.candidate.stream],
        record_mapping: { kind: "input_fields", inputs: selected.candidate.inputs },
      }],
    };
  }
  return document;
}

function schemaTypes(schema) {
  if (typeof schema?.type === "string") return [schema.type];
  if (Array.isArray(schema?.type)) return schema.type;
  if (schema?.properties) return ["object"];
  if (schema?.items) return ["array"];
  return ["any"];
}

function requiredMappingPaths(schema, prefix = "") {
  const paths = [];
  for (const name of array(schema?.required)) {
    const child = schema?.properties?.[name];
    const path = prefix ? `${prefix}.${name}` : name;
    const childTypes = schemaTypes(child);
    if (childTypes.includes("object")) {
      const descendants = requiredMappingPaths(child, path);
      paths.push(...(descendants.length > 0 ? descendants : [path]));
    } else if (childTypes.includes("array")) {
      const descendants = requiredMappingPaths(child?.items, `${path}.0`);
      paths.push(...(descendants.length > 0 ? descendants : [path]));
    } else {
      paths.push(path);
    }
  }
  return paths;
}

function schemaAtPath(schema, path) {
  let current = schema;
  for (const part of path.split(".")) {
    if (schemaTypes(current).includes("array")) {
      if (!/^\d+$/.test(part)) return null;
      current = current.items;
    } else {
      current = current?.properties?.[part];
    }
    if (!current) return null;
  }
  return current;
}

function flagTypeFor(schema, atTopLevel) {
  const types = schemaTypes(schema);
  if (atTopLevel && types.length === 1 && (types[0] === "object" || types[0] === "array")) return "json";
  if (types.includes("string")) return "string";
  if (types.includes("integer")) return "integer";
  if (types.includes("number")) return "number";
  if (types.includes("boolean")) return "boolean";
  if (types.includes("array") && schemaTypes(schema?.items).length === 1 && schemaTypes(schema?.items)[0] === "string") return "string_array";
  return null;
}

function cliFlagName(path, seen) {
  const base = path
    .replace(/([a-z0-9])([A-Z])/g, "$1-$2")
    .replace(/[_.]/g, "-")
    .toLowerCase();
  let name = base;
  let suffix = 2;
  while (seen.has(name)) name = `${base}-${suffix++}`;
  seen.add(name);
  return name;
}

function materializeWriteFlags(action) {
  const required = new Set([...requiredMappingPaths(action.record_schema), ...array(action.path_fields)]);
  const flags = [];
  const seenNames = new Set();
  for (const path of [...required].sort()) {
    const schema = schemaAtPath(action.record_schema, path);
    const type = flagTypeFor(schema, !path.includes("."));
    if (!schema || !type) {
      return { flags: [], reason: `the closed CLI flag set cannot faithfully represent required record field ${path}` };
    }
    flags.push({
      name: cliFlagName(path, seenNames),
      type,
      summary: `Required ${path} record field.`,
      maps_to: `record.${path}`,
      required: true,
    });
  }
  return { flags };
}

const exactTemplateValue = /^\{\{\s*(record|config)\.([A-Za-z0-9_.-]+)\s*\}\}$/;
const exactRecordPlaceholder = /\{\{\s*record\.([A-Za-z0-9_.-]+)\s*\}\}/g;

function exactRouteBinding(action, endpoints) {
  const matches = array(endpoints).filter((endpoint) => endpoint?.covered_by?.write === action.name);
  if (matches.length !== 1) {
    return { reason: `exact api_surface binding requires one covered_by.write match for ${action.name}; found ${matches.length}` };
  }
  const endpoint = matches[0];
  if (endpoint.method !== action.method.toUpperCase()) {
    return { reason: `typed action method ${action.method.toUpperCase()} disagrees with covered api_surface method ${endpoint.method}` };
  }
  const pathParts = action.path.split("?");
  if (pathParts.length > 2) return { reason: "typed action path has more than one query separator" };
  let route = pathParts[0];
  const placeholders = [...route.matchAll(/\{\{([^}]*)\}\}/g)];
  for (const placeholder of placeholders) {
    const match = `{{${placeholder[1]}}}`.match(exactTemplateValue);
    if (!match || match[1] !== "record") return { reason: `typed action path uses unrecognized placeholder ${placeholder[0]}` };
  }
  route = route.replace(exactRecordPlaceholder, "{$1}");
  if (route !== endpoint.path) {
    return { reason: `canonical typed action path ${route} disagrees with covered api_surface path ${endpoint.path}` };
  }
  return { endpoint, query: pathParts[1] ?? "" };
}

function queryInputFlags(query, spec, action, seenNames) {
  if (!query) return { flags: [] };
  const flags = [];
  for (const item of query.split("&")) {
    const [parameter, value, ...extra] = item.split("=");
    if (!parameter || value === undefined || extra.length > 0) return { reason: `query input ${item} is not a single declared parameter template` };
    const match = value.match(exactTemplateValue);
    if (!match) return { reason: `query input ${parameter} uses unrecognized placeholder ${value}` };
    const [, source, field] = match;
    if (source === "record") {
      const fieldSchema = schemaAtPath(action.record_schema, field);
      const type = flagTypeFor(fieldSchema, !field.includes("."));
      if (!fieldSchema || !type) return { reason: `query record input ${parameter} cannot be represented by the closed CLI flag set` };
      flags.push({ name: cliFlagName(field, seenNames), type, summary: `Declared ${parameter} query input.`, maps_to: `record.${field}` });
      continue;
    }
    if (field.includes(".")) return { reason: `query config input ${parameter} has unsupported nested key ${field}` };
    const configSchema = spec?.properties?.[field];
    const type = flagTypeFor(configSchema, false);
    if (!configSchema || !type || type === "json" || type === "string_array") return { reason: `query config input ${parameter} has no representable typed config schema` };
    flags.push({ name: cliFlagName(field, seenNames), type, summary: `Declared ${parameter} config input.`, maps_to: `config.${field}` });
  }
  return { flags };
}

async function materializeWriteCLI(connector) {
  const [metadata, writesDocument, existing, apiSurface, spec] = await Promise.all([
    readJSON(file(connector, "metadata.json")),
    readJSON(file(connector, "writes.json"), { actions: [] }),
    readJSON(file(connector, "cli_surface.json"), null),
    readJSON(file(connector, "api_surface.json"), { endpoints: [] }),
    readJSON(file(connector, "spec.json"), { properties: {} }),
  ]);
  const surface = existing ?? {
    tagline: `Run ${metadata.display_name}'s declared typed write actions.`,
    usage: `pm ${connector} <command> [flags]`,
    groups: [],
    commands: [],
  };
  surface.groups = array(surface.groups);
  surface.commands = array(surface.commands);
  const commandsByWrite = new Map();
  const paths = new Set();
  for (const command of surface.commands) {
    paths.add(command.path);
    if (command.write) commandsByWrite.set(command.write, command);
  }
  let writeGroup = surface.groups.find((group) => /write|reverse/i.test(group.id));
  if (!writeGroup) {
    writeGroup = { id: "reverse_etl", title: "Reverse ETL writes", commands: [] };
    surface.groups.push(writeGroup);
  }
  writeGroup.commands = array(writeGroup.commands);
  for (const action of array(writesDocument.actions)) {
    const existingCommand = commandsByWrite.get(action.name);
    const route = exactRouteBinding(action, apiSurface.endpoints);
    const generated = existingCommand?.notes?.startsWith("Generated from the connector-owned typed action;");
    if (existingCommand && !generated && existingCommand.availability === "implemented") continue;
    const path = generated ? existingCommand.path : `${action.name.replaceAll("_", " ")} apply`;
    if (!generated) assert(!paths.has(path), `${connector}: generated command path ${path} conflicts with an existing command`);
    const materialized = materializeWriteFlags(action);
    const query = route.reason ? { flags: [], reason: route.reason } : queryInputFlags(route.query, spec, action, new Set(materialized.flags.map((flag) => flag.name)));
    const flags = [...materialized.flags];
    for (const flag of query.flags) {
      if (!flags.some((existingFlag) => existingFlag.maps_to === flag.maps_to)) flags.push(flag);
    }
    const reason = materialized.reason ?? query.reason;
    const configInput = flags.find((flag) => flag.maps_to.startsWith("config."));
    const availability = reason || configInput ? "partial" : "implemented";
    const blockedReason = reason
      ? `declaration-pending: ${reason}`
      : "foundation-gap: internal/connectors/commandrunner/runner.go:1565 rejects config.* flags in reverse-ETL recordOverrides; minimal change: apply declared config overrides through the existing closed config-override path before assembling the typed write record";
    const command = {
      path,
      summary: route.endpoint ? `${route.endpoint.method} ${route.endpoint.path} (${action.name})` : `Typed action ${action.name}`,
      intent: "reverse_etl",
      availability,
      write: action.name,
      ...(route.endpoint ? { api_surface: [{ method: route.endpoint.method, path: route.endpoint.path }] } : {}),
      flags,
      risk: action.risk,
      approval: availability === "partial"
        ? `Blocked pending a faithful CLI record binding: ${blockedReason}.`
        : "Reverse ETL writes require plan, preview, approval, and execute.",
      notes: availability === "partial"
        ? `Generated from the connector-owned typed action; ${blockedReason}.`
        : "Generated from the connector-owned typed action; execution remains plan-gated.",
      examples: [`pm ${connector} ${path} --plan <plan-name>`],
    };
    if (generated) {
      delete existingCommand.api_surface;
      Object.assign(existingCommand, command);
    } else {
      surface.commands.push(command);
      writeGroup.commands.push(path);
      paths.add(path);
    }
  }
  surface.commands.sort((left, right) => left.path.localeCompare(right.path));
  writeGroup.commands = [...new Set(writeGroup.commands)].sort();
  await writeFile(file(connector, "cli_surface.json"), pretty(surface));
}

async function inspect(connector) {
  const dir = resolve(defs, connector);
  const [streamsDocument, writesDocument, cliSurface, disposition] = await Promise.all([
    readJSON(file(connector, "streams.json"), { streams: [] }),
    readJSON(file(connector, "writes.json"), { actions: [] }),
    readJSON(file(connector, "cli_surface.json"), { commands: [] }),
    readJSON(file(connector, `sources/${connector}-declaration-disposition.json`)),
  ]);
  const streams = array(streamsDocument.streams);
  const schemas = new Map();
  for (const stream of streams) schemas.set(stream.name, await readJSON(resolve(dir, stream.schema)));
  const actions = array(writesDocument.actions);
  const candidates = [];
  const actionsLedger = actions.map((action) => {
    const candidate = candidateFor(action, streams, schemas);
    if (candidate.state === "representable") candidates.push({ action, candidate });
    const cli = actionCLICommands(cliSurface, action);
    const implementedCLI = cli.filter((command) => command.availability === "implemented");
    const partialCLI = cli.filter((command) => command.availability === "partial");
    return {
      action: action.name,
      kind: action.kind,
      direct_write_cli: cli.map((command) => ({ path: command.path, availability: command.availability, block_reason: command.availability === "partial" ? command.notes ?? command.approval : undefined })),
      direct_write_cli_status: implementedCLI.length > 0
        ? "implemented"
        : partialCLI.length > 0
          ? "partial-blocked"
        : "declaration-pending-cli-binding",
      reverse_etl_eligibility: candidate,
    };
  });
  const preferred = candidates.find(({ action }) => action.confirm !== "destructive" && action.confirmation?.kind !== "destructive") ?? candidates[0];
  const selected = preferred && { name: preferred.action.name, candidate: preferred.candidate };
  const source = disposition.summary?.state === "skipped" || disposition.summary?.state === "dynamic"
    ? { state: disposition.summary.state, operations: null }
    : { state: "mapped", operations: disposition.summary?.operations_found ?? 0 };
  const parity = Object.fromEntries(array(disposition.summary?.parity_class_counts).map(({ key, count }) => [key, count]));
  const syncPath = file(connector, "sync_transport.json");
  const sync = await readJSON(syncPath, null);
  return {
    batch: batchFor(connector),
    connector,
    provider_operations: source,
    surfaces: {
      binary_read: parity.binary_read ?? 0,
      binary_write: parity.binary_write ?? 0,
      direct_read: parity.direct_read ?? 0,
      direct_write: parity.direct_write ?? 0,
      etl: { executable_streams: streams.map((stream) => stream.name), source_transport: sync?.source_transport ? "declared-static" : "declaration-pending" },
      reverse_etl: {
        destination_transport: sync?.destination_transport ? "declared-static; application-dispatch-pending-foundation" : "declaration-pending",
        selected_initial_proof: selected?.name ?? null,
        exact_record_driven_actions: candidates.map(({ action }) => action.name),
        multiplicity_dependency: candidates.length > 1 ? "one apply_strategy action is selectable per mode; remaining exact actions await foundation multi-action selection" : null,
        actions: actionsLedger,
      },
      cli: {
        declared_commands: array(cliSurface.commands).length,
        implemented_commands: array(cliSurface.commands).filter((command) => command.availability === "implemented").length,
        partial_commands: array(cliSurface.commands).filter((command) => command.availability === "partial").length,
      },
    },
  };
}

function verifyDeclaration(row, sync) {
  const source = sync?.source_transport;
  assert(source, `${row.connector}: source transport declaration missing`);
  assert(source.executor?.family === "declarative_api" && source.executor?.id === "declarative_stream_source", `${row.connector}: source transport must select declarative_stream_source`);
  assert(JSON.stringify(source.eligible_streams) === JSON.stringify(row.surfaces.etl.executable_streams), `${row.connector}: source eligible streams must exactly match executable streams`);
  const destination = sync.destination_transport;
  const representable = row.surfaces.reverse_etl.exact_record_driven_actions;
  if (representable.length === 0) {
    assert(!destination, `${row.connector}: destination declaration cannot invent a record-driven action`);
    return;
  }
  assert(destination, `${row.connector}: representable actions need a destination declaration`);
  assert(destination.executor?.family === "declarative_api" && destination.executor?.id === "declarative_typed_destination", `${row.connector}: destination must select declarative_typed_destination`);
  assert(JSON.stringify(destination.eligible_actions) === JSON.stringify(representable), `${row.connector}: destination eligible actions must list every exact record-driven action`);
  assert(destination.acknowledgement === "durable_warehouse", `${row.connector}: destination acknowledgement must be durable_warehouse`);
  assert(destination.apply_strategies?.length === 1 && destination.apply_strategies[0].mode === "full_append", `${row.connector}: initial proof must select one full_append action`);
  assert(representable.includes(destination.apply_strategies[0].action), `${row.connector}: selected action is not exact-record-driven`);
  const mapping = destination.source_bindings?.[0]?.record_mapping;
  assert(mapping?.kind === "input_fields" && array(mapping.inputs).length > 0, `${row.connector}: destination requires exact input_fields mapping`);
}

function verifyWriteCLI(row) {
  for (const action of row.surfaces.reverse_etl.actions) {
    assert(action.direct_write_cli.length > 0, `${row.connector}: typed write action ${action.action} lacks a direct CLI binding`);
    assert(action.direct_write_cli_status !== "declaration-pending-cli-binding", `${row.connector}: typed write action ${action.action} has no executable or explicitly blocked CLI binding`);
  }
}

function summary(rows) {
  const lines = [
    "# Issue #4292 seven-surface reconciliation",
    "",
    "Generated entirely from the pinned source ledgers and existing connector-owned stream/action schemas. `declared-static` means structural, credential-free contract validation only; application-level generic destination dispatch remains pending the latest #4304 foundation integration and its installed App/CLI proof.",
    "",
    "| Batch | Connector | Provider operations | Direct reads | Direct writes | Binary R/W | Streams | Selected destination proof | Eligible typed actions | CLI implemented/partial/declared |",
    "| --- | --- | ---: | ---: | ---: | --- | ---: | --- | ---: | ---: |",
  ];
  for (const row of rows) {
    const s = row.surfaces;
    lines.push(`| ${row.batch} | ${row.connector} | ${row.provider_operations.operations ?? row.provider_operations.state} | ${s.direct_read} | ${s.direct_write} | ${s.binary_read}/${s.binary_write} | ${s.etl.executable_streams.length} | ${s.reverse_etl.selected_initial_proof ?? "—"} | ${s.reverse_etl.exact_record_driven_actions.length} | ${s.cli.implemented_commands}/${s.cli.partial_commands}/${s.cli.declared_commands} |`);
  }
  lines.push("", "Each typed write action has a machine-readable `reverse_etl_eligibility` disposition and `direct_write_cli_status` in `SEVEN-SURFACE-LEDGER.json`. `partial-blocked` has a directly invokable CLI path plus an exact closed-contract blocker; `declaration-pending-cli-binding` is an unfinished reachability obligation, not a safety exclusion. When more than one action is structurally representable, the declaration lists every eligible action but the current closed destination multiplicity selects one action per mode; unselected actions are explicitly pending the foundation's multi-action selection capability. Semantic exclusions name the exact record-schema incompatibility and remain subject to direct CLI reachability.", "");
  return lines.join("\n");
}

const args = process.argv.slice(2);
const writeIndex = args.indexOf("--write-declarations");
const writeCLIIndex = args.indexOf("--write-cli");
const checkIndex = args.indexOf("--check");
const selectedNames = args.filter((value) => !value.startsWith("--"));
const targets = selectedNames.length === 0 ? connectors : selectedNames;
for (const target of targets) assert(connectors.includes(target), `unassigned connector ${target}`);

const rows = [];
for (const connector of connectors) rows.push(await inspect(connector));
if (writeIndex >= 0) {
  for (const row of rows.filter((row) => targets.includes(row.connector))) {
    const streams = row.surfaces.etl.executable_streams.map((name) => ({ name }));
    const candidates = row.surfaces.reverse_etl.actions
      .filter((action) => action.reverse_etl_eligibility.state === "representable")
      .map((action) => ({ action: { name: action.action }, candidate: action.reverse_etl_eligibility }));
    const selected = row.surfaces.reverse_etl.selected_initial_proof
      ? { name: row.surfaces.reverse_etl.selected_initial_proof, candidate: candidates.find(({ action }) => action.name === row.surfaces.reverse_etl.selected_initial_proof).candidate }
      : null;
    await writeFile(file(row.connector, "sync_transport.json"), pretty(declaration(row.connector, streams, selected, candidates)));
  }
  // Re-inspect after writes so output represents the declaration state.
  rows.length = 0;
  for (const connector of connectors) rows.push(await inspect(connector));
}
if (writeCLIIndex >= 0) {
  for (const connector of targets) await materializeWriteCLI(connector);
  rows.length = 0;
  for (const connector of connectors) rows.push(await inspect(connector));
}
const document = { schema_version: 1, issue: 4292, application_dispatch: "pending-foundation-app-cli-integration", rows };
if (checkIndex >= 0) {
  for (const row of rows.filter((row) => targets.includes(row.connector))) {
    verifyDeclaration(row, await readJSON(file(row.connector, "sync_transport.json"), null));
    verifyWriteCLI(row);
  }
  assert(existsSync(ledgerPath) && existsSync(summaryPath), "seven-surface outputs are missing; run without --check");
  assert(await readFile(ledgerPath, "utf8") === pretty(document), "seven-surface machine ledger drift; regenerate it");
  assert(await readFile(summaryPath, "utf8") === summary(rows), "seven-surface human summary drift; regenerate it");
  process.stdout.write(`verified seven surfaces for ${targets.length} connector(s)\n`);
} else {
  await writeFile(ledgerPath, pretty(document));
  await writeFile(summaryPath, summary(rows));
  const changes = [
    writeIndex >= 0 ? `declared ${targets.length}` : "",
    writeCLIIndex >= 0 ? `materialized CLI for ${targets.length}` : "",
  ].filter(Boolean);
  process.stdout.write(`generated seven-surface ledger for ${rows.length} connectors${changes.length > 0 ? `; ${changes.join(", ")}` : ""}\n`);
}
