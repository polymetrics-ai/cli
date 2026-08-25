#!/usr/bin/env node

// Materialize only the distinct Bigin provider GET operations that already
// have an exact disposition row, api-surface stream projection, and declared
// stream. The generic /{module_api_name} operation intentionally produces one
// `records list` command; its configured-module aliases are not extra provider
// operations and must not be manufactured as source identities.

import { readFile, writeFile } from "node:fs/promises";
import { resolve } from "node:path";

const phase = resolve(import.meta.dirname, "..");
const root = resolve(phase, "../../..");
const connector = "zoho-bigin";
const definitionRoot = resolve(root, "internal/connectors/defs", connector);

async function readJSON(file) {
  return JSON.parse(await readFile(file, "utf8"));
}

async function writeJSON(file, value) {
  await writeFile(file, `${JSON.stringify(value, null, 2)}\n`);
}

function words(value) {
  return value.split("_").map((word) => word[0].toUpperCase() + word.slice(1)).join(" ");
}

function commandPath(stream) {
  return `${stream.replaceAll("_", "-")} list`;
}

const lock = await readJSON(resolve(definitionRoot, "sources", `${connector}-operation-source-lock.json`));
if (lock.schema_version !== 3 || !Array.isArray(lock.rest?.source_documents)) {
  throw new Error(`${connector}: expected schema-version 3 source_documents`);
}
const sourceOperations = new Map();
for (const document of lock.rest.source_documents) {
  if (document.kind !== "rendered_reference" || !document.published_source?.source_url) {
    throw new Error(`${connector}: source document ${document.id} lacks the rendered-reference contract`);
  }
  for (const operation of document.operations ?? []) {
    if (operation.protocol !== "rest") continue;
    if (!operation.citation_url) {
      throw new Error(`${connector}:${operation.id}: rendered-reference operation has no citation_url`);
    }
    if (sourceOperations.has(operation.id)) {
      throw new Error(`${connector}:${operation.id}: duplicate source operation ID`);
    }
    sourceOperations.set(operation.id, { document, operation });
  }
}

const [streams, apiSurface, disposition, surface, evidence] = await Promise.all([
  readJSON(resolve(definitionRoot, "streams.json")),
  readJSON(resolve(definitionRoot, "api_surface.json")),
  readJSON(resolve(definitionRoot, "sources", `${connector}-declaration-disposition.json`)),
  readJSON(resolve(definitionRoot, "cli_surface.json")),
  readJSON(resolve(phase, "OPERATION-SURFACE-EVIDENCE.json")),
]);

const streamsByName = new Map(streams.streams.map((stream) => [stream.name, stream]));
const etlRows = disposition.ledger_dispositions.filter((row) => row.parity_class === "etl");
if (etlRows.length !== 6) {
  throw new Error(`${connector}: expected six ETL disposition rows, found ${etlRows.length}`);
}

const generated = [];
const commandEvidence = new Map();
for (const row of etlRows) {
  const source = sourceOperations.get(row.source?.source_id);
  if (!source) {
    throw new Error(`${connector}:${row.source?.source_id}: source operation missing from lock`);
  }
  const { document, operation } = source;
  if (operation.method !== row.method || operation.path !== row.path || operation.citation_url !== row.source.source_url) {
    throw new Error(`${connector}:${operation.id}: disposition does not match the pinned operation`);
  }
  const streamName = row.api_surface?.covered_by?.stream;
  const stream = streamsByName.get(streamName);
  if (!stream) {
    throw new Error(`${connector}:${operation.id}: missing declared stream ${JSON.stringify(streamName)}`);
  }
  const endpoints = apiSurface.endpoints.filter((endpoint) => (
    endpoint.method === operation.method
    && endpoint.path === operation.path
    && endpoint.covered_by?.stream === streamName
    && endpoint.provenance?.source_url === operation.citation_url
  ));
  if (endpoints.length !== 1) {
    throw new Error(`${connector}:${operation.id}: expected one exact API-surface stream projection, found ${endpoints.length}`);
  }
  const path = commandPath(streamName);
  const command = {
    path,
    summary: `Read ${words(streamName)} as Zoho Bigin ETL records.`,
    intent: "etl",
    availability: "implemented",
    stream: streamName,
    source_cli_path: `${operation.method} ${operation.path}`,
    source_url: operation.citation_url,
    api_surface: [{ method: operation.method, path: operation.path }],
    examples: [`pm ${connector} ${path} --json`],
    notes: `Pinned source operation ${operation.id} from rendered source document ${document.id}; streams.json stream ${JSON.stringify(streamName)} is the declared execution component.`,
  };
  generated.push(command);
  commandEvidence.set(operation.id, { command, document, operation, stream, endpoint: endpoints[0] });
}

const generatedPaths = new Set(generated.map((command) => command.path));
const retainedCommands = surface.commands.filter((command) => !generatedPaths.has(command.path));
surface.tagline = "Read Zoho Bigin source-locked ETL records and run typed write plans.";
surface.source_cli = {
  name: "Zoho Bigin provider API",
  docs: generated[0].source_url,
  reference: "Pinned rendered-reference source citations, stream declarations, and connector-owned typed write contracts.",
  source: "provider_api",
};
surface.global_flags ??= [];
if (!surface.global_flags.some((flag) => flag.name === "limit")) {
  surface.global_flags.push({ name: "limit", type: "integer", summary: "Maximum ETL records to emit." });
}
surface.groups = surface.groups.filter((group) => group.id !== "etl");
surface.groups.push({ id: "etl", title: "ETL streams", commands: generated.map((command) => command.path.split(" ")[0]) });
surface.commands = [...retainedCommands, ...generated].sort((left, right) => left.path.localeCompare(right.path));
await writeJSON(resolve(definitionRoot, "cli_surface.json"), surface);

// api_surface represents the legacy configured-module aliases as duplicate
// projections of this one provider route. The CLI safety map is keyed by
// method/path, so retain each alias for stream conformance while placing the
// canonical source-bound `records` projection last for that key.
const recordIndex = apiSurface.endpoints.findIndex((endpoint) => (
  endpoint.method === "GET"
  && endpoint.path === "/{module_api_name}"
  && endpoint.covered_by?.stream === "records"
));
const lastModuleIndex = apiSurface.endpoints.reduce((last, endpoint, index) => (
  endpoint.method === "GET" && endpoint.path === "/{module_api_name}" ? index : last
), -1);
if (recordIndex < 0 || lastModuleIndex < 0) {
  throw new Error(`${connector}: missing canonical records API-surface projection`);
}
if (recordIndex < lastModuleIndex) {
  const [recordProjection] = apiSurface.endpoints.splice(recordIndex, 1);
  apiSurface.endpoints.splice(lastModuleIndex, 0, recordProjection);
}
const moduleProjections = apiSurface.endpoints.filter((endpoint) => endpoint.method === "GET" && endpoint.path === "/{module_api_name}");
if (moduleProjections.length !== 8 || moduleProjections.at(-1)?.covered_by?.stream !== "records") {
  throw new Error(`${connector}: canonical records projection did not become the method/path safety target`);
}
await writeJSON(resolve(definitionRoot, "api_surface.json"), apiSurface);

for (const row of etlRows) {
  const item = commandEvidence.get(row.source.source_id);
  const { command, document, operation, stream } = item;
  row.state = "enabled";
  row.rejection = null;
  row.foundation = {
    state: "present",
    evidence: `internal/connectors/defs/${connector}/cli_surface.json: implemented ETL command ${JSON.stringify(command.path)} binds ${operation.method} ${operation.path}; streams.json stream ${JSON.stringify(stream.name)} and sync_transport.json source transport are declared execution components; the built binary must stop at missing --credential before provider I/O.`,
  };
  row.declaration = {
    status: `enabled; runnable ETL command ${JSON.stringify(command.path)} binds pinned source operation ${operation.id}`,
    command: {
      path: command.path,
      intent: command.intent,
      availability: command.availability,
      source_cli_path: command.source_cli_path,
      source_url: command.source_url,
    },
    transport: {
      source_transport: {
        state: "declared",
        evidence: `internal/connectors/defs/${connector}/sync_transport.json declares declarative_stream_source eligibility for stream ${JSON.stringify(stream.name)} with keyed source delivery and declarative_stream_transport/zoho_bigin_fixture_source_v1 conformance.`,
      },
    },
    source_document: {
      id: document.id,
      kind: document.kind,
      published_source_url: document.published_source.source_url,
      citation_url: operation.citation_url,
    },
  };
}
const summary = disposition.summary;
summary.enabled_operations = disposition.ledger_dispositions.filter((row) => row.state === "enabled").length;
summary.disabled_operations = disposition.ledger_dispositions.filter((row) => row.state !== "enabled").length;
summary.enabled_percent = Number(((summary.enabled_operations / disposition.ledger_dispositions.length) * 100).toFixed(2));
summary.terminal_commands = surface.commands.length;
summary.runnable_cli_surface_commands = surface.commands.length;
summary.endpoint_bound_cli_commands = surface.commands.length;
summary.rejected_by_reason = summary.disabled_operations === 0 ? [] : [{ key: "declaration-pending", count: summary.disabled_operations }];
summary.declaration_pending_ids = (summary.declaration_pending_ids ?? []).filter((id) => id !== "runnable-command-binding-zoho-bigin");
summary.declaration_pending = (summary.declaration_pending ?? []).filter((entry) => entry.id !== "runnable-command-binding-zoho-bigin");
disposition.notes = [
  ...(disposition.notes ?? []).filter((note) => !note.includes("2026-08-26 source-bound ETL command materialization")),
  "2026-08-26 source-bound ETL command materialization: enabled six distinct provider GET operations. The generic /{module_api_name} operation is represented once as records list; its configured-module aliases are not separate provider operations. Every command retains a source ID, rendered citation, route, streams.json execution component, and declared source transport.",
];
await writeJSON(resolve(definitionRoot, "sources", `${connector}-declaration-disposition.json`), disposition);

let changedEvidence = 0;
for (const row of evidence.rows) {
  if (row.connector !== connector) continue;
  const item = commandEvidence.get(row.provider_operation?.source_id);
  if (!item) continue;
  const { command, document, operation, stream } = item;
  row.canonical_mapping.source_state = "enabled";
  row.canonical_mapping.declaration_status = `enabled; runnable ETL command ${JSON.stringify(command.path)} binds pinned source operation ${operation.id}`;
  row.canonical_mapping.contract = {
    execution_component: `internal/connectors/defs/${connector}/streams.json stream ${JSON.stringify(stream.name)}`,
    transport_component: `internal/connectors/defs/${connector}/sync_transport.json source_transport`,
    source_document: document.id,
    source_kind: document.kind,
    source_cli_path: command.source_cli_path,
    citation_url: command.source_url,
  };
  row.canonical_mapping.api_surface_binding = { stream: stream.name };
  row.canonical_mapping.rejection = null;
  row.surfaces.etl = {
    state: "executable_stream_transport_declared",
    source_state: "enabled",
    execution_component: `streams.json:${stream.name}`,
  };
  row.surfaces.executable_cli = {
    state: "generated_binding_present",
    binding: { stream: stream.name, path: command.path, source_cli_path: command.source_cli_path },
  };
  changedEvidence += 1;
}
if (changedEvidence !== generated.length) {
  throw new Error(`operation evidence: changed ${changedEvidence} rows, expected ${generated.length}`);
}
await writeJSON(resolve(phase, "OPERATION-SURFACE-EVIDENCE.json"), evidence);

console.log(`generated ${generated.length} exact Zoho Bigin ETL command bindings`);
