#!/usr/bin/env node

// Generates only command surfaces that are already fully represented by a
// checked-in stream, source-locked operation, and api_surface projection. The
// source lock remains the provider boundary; api_surface is a consistency
// check, never an inventory input.

import { readFile, writeFile } from "node:fs/promises";
import { resolve } from "node:path";

const phase = resolve(import.meta.dirname, "..");
const root = resolve(phase, "../../..");
const cohort = ["activecampaign", "freshdesk", "iterable", "segment", "square"];

async function readJSON(file) {
  return JSON.parse(await readFile(file, "utf8"));
}

async function writeJSON(file, value) {
  await writeFile(file, `${JSON.stringify(value, null, 2)}\n`);
}

function words(name) {
  return name.split("_").map((word) => word[0].toUpperCase() + word.slice(1)).join(" ");
}

function commandPath(stream) {
  return `${stream.replaceAll("_", "-")} list`;
}

function sourceIndex(lock, connector) {
  if (lock.schema_version !== 3 || lock.connector !== connector || !Array.isArray(lock.rest?.source_documents)) {
    throw new Error(`${connector}: expected a schema-version 3 source lock with source_documents`);
  }
  const index = new Map();
  for (const document of lock.rest.source_documents) {
    if (!document.published_source?.source_url || !Array.isArray(document.operations)) {
      throw new Error(`${connector}: source document ${document.id} is incomplete`);
    }
    for (const operation of document.operations) {
      if (operation.protocol !== "rest") {
        continue;
      }
      if (document.kind === "rendered_reference" && !operation.citation_url) {
        throw new Error(`${connector}: rendered-reference operation ${operation.id} has no citation_url`);
      }
      const key = `${operation.method} ${operation.path}`;
      const matches = index.get(key) ?? [];
      matches.push({ document, operation });
      index.set(key, matches);
    }
  }
  return index;
}

const commandEvidence = new Map();

for (const connector of cohort) {
  const definitionRoot = resolve(root, "internal/connectors/defs", connector);
  const streams = await readJSON(resolve(definitionRoot, "streams.json"));
  const lock = await readJSON(resolve(definitionRoot, "sources", `${connector}-operation-source-lock.json`));
  const apiSurface = await readJSON(resolve(definitionRoot, "api_surface.json"));
  const sources = sourceIndex(lock, connector);
  const commands = [];

  for (const stream of streams.streams) {
    const endpoints = apiSurface.endpoints.filter((endpoint) => (
      endpoint.method === "GET" && endpoint.covered_by?.stream === stream.name
    ));
    if (endpoints.length !== 1) {
      throw new Error(`${connector}:${stream.name}: expected one existing GET api_surface stream projection, found ${endpoints.length}`);
    }
    const endpoint = endpoints[0];
    const sourceMatches = sources.get(`${endpoint.method} ${endpoint.path}`) ?? [];
    if (sourceMatches.length !== 1) {
      throw new Error(`${connector}:${stream.name}: expected one pinned source operation for ${endpoint.method} ${endpoint.path}, found ${sourceMatches.length}`);
    }
    const { document, operation } = sourceMatches[0];
    const sourceURL = operation.citation_url ?? document.published_source.source_url;
    const path = commandPath(stream.name);
    const command = {
      path,
      summary: `Read ${words(stream.name)} as ${words(connector)} ETL records.`,
      intent: "etl",
      availability: "implemented",
      stream: stream.name,
      source_cli_path: `${operation.method} ${operation.path}`,
      source_url: sourceURL,
      api_surface: [{ method: endpoint.method, path: endpoint.path }],
      examples: [`pm ${connector} ${path} --json`],
      notes: `Pinned source operation ${operation.id} from source document ${document.id}; streams.json is the declared execution component.`,
    };
    commands.push(command);
    commandEvidence.set(`${connector}:${operation.id}`, { command, document, operation, stream, endpoint });
  }

  if (commands.length !== streams.streams.length) {
    throw new Error(`${connector}: did not generate every checked-in stream command`);
  }
  const firstDocument = lock.rest.source_documents[0];
  const surface = {
    tagline: `Read ${words(connector)} source-locked ETL records.`,
    usage: `pm ${connector} <command> [flags]`,
    source_cli: {
      name: `${words(connector)} provider API`,
      docs: firstDocument.published_source.source_url,
      reference: "Pinned schema-v3 source documents; each ETL command retains its exact provider operation route and citation.",
      source: "provider_api",
    },
    groups: [{
      id: "etl",
      title: "ETL streams",
      commands: commands.map((command) => command.path.split(" ")[0]),
    }],
    global_flags: [
      { name: "credential", type: "string", summary: `Credential name to use for the ${words(connector)} request.` },
      { name: "connection", type: "string", summary: "Alias for --credential." },
      { name: "config", type: "string_array", summary: "Connector config override as key=value; never pass secret values here." },
      { name: "json", type: "boolean", summary: "Emit machine-readable JSON output." },
      { name: "limit", type: "integer", summary: "Maximum ETL records to emit." },
    ],
    commands,
  };
  await writeJSON(resolve(definitionRoot, "cli_surface.json"), surface);
}

for (const connector of cohort) {
  const definitionRoot = resolve(root, "internal/connectors/defs", connector);
  const dispositionFile = resolve(definitionRoot, "sources", `${connector}-declaration-disposition.json`);
  const disposition = await readJSON(dispositionFile);
  const etlRows = disposition.ledger_dispositions.filter((row) => row.parity_class === "etl");
  const changed = [];
  for (const row of etlRows) {
    const evidence = commandEvidence.get(`${connector}:${row.source.source_id}`);
    if (!evidence) {
      throw new Error(`${connector}:${row.source.source_id}: the ETL disposition has no exact source command evidence`);
    }
    const { command, document, operation, stream, endpoint } = evidence;
    if (row.method !== endpoint.method || row.path !== endpoint.path || row.api_surface?.covered_by?.stream !== stream.name) {
      throw new Error(`${connector}:${row.source.source_id}: disposition, stream, and API-surface route do not agree`);
    }
    row.state = "enabled";
    row.rejection = null;
    row.foundation = {
      state: "present",
      evidence: `internal/connectors/defs/${connector}/cli_surface.json: implemented ETL command ${JSON.stringify(command.path)} binds ${operation.method} ${operation.path}; streams.json stream ${JSON.stringify(stream.name)} is the execution component and the built binary must stop at missing --credential before provider I/O.`,
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
          state: "declaration-pending",
          evidence: `internal/connectors/defs/${connector}/sync_transport.json is absent; this executable ETL command uses the existing streams.json component, while persisted source transport remains connector-owned declaration work.`,
        },
      },
      source_document: {
        id: document.id,
        kind: document.kind,
        published_source_url: document.published_source.source_url,
        citation_url: operation.citation_url ?? null,
      },
    };
    changed.push(row);
  }
  if (changed.length !== etlRows.length || changed.length === 0) {
    throw new Error(`${connector}: expected to enable every ETL disposition row, changed ${changed.length} of ${etlRows.length}`);
  }
  const summary = disposition.summary;
  summary.enabled_operations = disposition.ledger_dispositions.filter((row) => row.state === "enabled").length;
  summary.disabled_operations = disposition.ledger_dispositions.filter((row) => row.state !== "enabled").length;
  summary.enabled_percent = Number(((summary.enabled_operations / disposition.ledger_dispositions.length) * 100).toFixed(2));
  summary.terminal_commands = changed.length;
  summary.runnable_cli_surface_commands = changed.length;
  summary.endpoint_bound_cli_commands = changed.length;
  summary.rejected_by_reason = summary.disabled_operations === 0 ? [] : [{ key: "declaration-pending", count: summary.disabled_operations }];
  summary.declaration_pending_ids = summary.declaration_pending_ids.filter((id) => id !== `runnable-command-binding-${connector}`);
  summary.declaration_pending = summary.declaration_pending.filter((entry) => entry.id !== `runnable-command-binding-${connector}`);
  disposition.notes = [
    ...(disposition.notes ?? []).filter((note) => !note.includes("2026-08-25 declaration-first ETL command materialization")),
    `2026-08-25 declaration-first ETL command materialization: enabled ${changed.length} exact stream/source pairs. Every command has its provider operation ID, canonical method/path, source citation, and streams.json execution component; persisted source transport remains declaration-pending until connector-owned transport/conformance evidence exists.`,
  ];
  await writeJSON(dispositionFile, disposition);
}

const operationEvidenceFile = resolve(phase, "OPERATION-SURFACE-EVIDENCE.json");
const operationEvidence = await readJSON(operationEvidenceFile);
let evidenceChanges = 0;
for (const row of operationEvidence.rows) {
  const evidence = commandEvidence.get(`${row.connector}:${row.provider_operation?.source_id}`);
  if (!evidence) {
    continue;
  }
  const { command, document, operation, stream, endpoint } = evidence;
  row.canonical_mapping.source_state = "enabled";
  row.canonical_mapping.declaration_status = `enabled; runnable ETL command ${JSON.stringify(command.path)} binds pinned source operation ${operation.id}`;
  row.canonical_mapping.contract = {
    execution_component: `internal/connectors/defs/${row.connector}/streams.json stream ${JSON.stringify(stream.name)}`,
    source_document: document.id,
    source_kind: document.kind,
    source_cli_path: command.source_cli_path,
    citation_url: command.source_url,
  };
  row.canonical_mapping.api_surface_binding = { stream: stream.name };
  row.canonical_mapping.rejection = null;
  row.surfaces.etl = {
    state: "executable_stream_pending_persisted_transport_declaration",
    source_state: "enabled",
    execution_component: `streams.json:${stream.name}`,
  };
  row.surfaces.executable_cli = {
    state: "generated_binding_present",
    binding: { stream: stream.name, path: command.path, source_cli_path: `${operation.method} ${endpoint.path}` },
  };
  evidenceChanges += 1;
}
if (evidenceChanges !== commandEvidence.size) {
  throw new Error(`operation evidence: changed ${evidenceChanges} rows, expected ${commandEvidence.size}`);
}
await writeJSON(operationEvidenceFile, operationEvidence);

console.log(`generated ${commandEvidence.size} source-locked ETL commands for ${cohort.join(", ")}`);
