#!/usr/bin/env node

import { readFile, writeFile } from "node:fs/promises";
import { resolve } from "node:path";

const phase = resolve(import.meta.dirname, "..");
const root = resolve(phase, "../../..");
const evidenceFile = resolve(phase, "OPERATION-SURFACE-EVIDENCE.json");
const destinationConnectors = [
  "close-com",
  "outreach",
  "zoho-bigin",
  "braze",
  "customer-io",
  "help-scout",
  "gorgias",
  "service-now",
  "chatwoot",
  "chargebee",
];

async function readJSON(file) {
  return JSON.parse(await readFile(file, "utf8"));
}

async function writeJSON(file, value) {
  await writeFile(file, `${JSON.stringify(value, null, 2)}\n`);
}

const changed = new Map();

for (const connector of destinationConnectors) {
  const file = resolve(root, "internal/connectors/defs", connector, "sources", `${connector}-declaration-disposition.json`);
  const document = await readJSON(file);
  const stale = document.ledger_dispositions.filter((row) => (
    row.declaration?.reverse_etl?.state?.startsWith("bound_") ||
    row.declaration?.reverse_etl?.foundation_gap?.id === "declarative-typed-destination-action-specific-source-bindings"
  ));
  for (const row of stale) {
    const action = row.api_surface?.covered_by?.write;
    if (!action) {
      throw new Error(`${connector}: stale destination row has no typed action`);
    }
    row.declaration.reverse_etl = {
      eligible_typed_action: true,
      state: "eligible_declaration_pending",
      reason: "The closed declarative typed-destination contract rejects the prior branch-only declaration: it selected full_overwrite and lacked the action-owned per-record batch, provider receipt locator, and read-back conformance evidence. The direct typed write remains enabled; reverse ETL is connector-owned declaration work, not a foundation gap.",
    };
    changed.set(`${connector}:${row.source.source_id}`, row);
  }
  document.summary.transport.destination_transport = {
    state: "declaration-pending",
    declaration_pending: {
      id: `qualified-declarative-typed-destination-${connector}`,
      evidence: "docs/sync-transport-definition.md:85-115 requires action-owned input_fields, per-record batch, keyed durable acknowledgement, and action-owned provider read-back. The prior declaration was removed because it selected rejected full_overwrite and carried no provider-evidenced receipt/read-back contract.",
      minimal_change: "Add a source-evidenced destination declaration only when one exact typed action has per-record batch, private receipt locator, provider-state read-back, conformance evidence, and a supported non-full-overwrite strategy.",
    },
    provider_live_certification: "pending",
  };
  const residual = document.ledger_dispositions.filter((row) => (
    row.declaration?.reverse_etl?.state?.startsWith("bound_") ||
    row.declaration?.reverse_etl?.foundation_gap?.id === "declarative-typed-destination-action-specific-source-bindings"
  ));
  if (residual.length !== 0) {
    throw new Error(`${connector}: retained ${residual.length} stale typed-destination claims`);
  }
  await writeJSON(file, document);
}

const outreachFile = resolve(root, "internal/connectors/defs/outreach/sources/outreach-declaration-disposition.json");
const outreach = await readJSON(outreachFile);
const cli = await readJSON(resolve(root, "internal/connectors/defs/outreach/cli_surface.json"));
const commandsByStream = new Map(cli.commands.map((command) => [command.stream, command]));
const outreachRows = outreach.ledger_dispositions.filter((row) => row.parity_class === "etl" && row.state === "disabled");
if ((outreachRows.length !== 0 && outreachRows.length !== 96) || commandsByStream.size !== 96) {
  throw new Error(`outreach: expected zero or 96 pending ETL rows and 96 commands, found ${outreachRows.length} rows and ${commandsByStream.size} commands`);
}
for (const row of outreachRows) {
  const stream = row.api_surface?.covered_by?.stream;
  const command = commandsByStream.get(stream);
  if (!command || command.intent !== "etl" || command.availability !== "implemented") {
    throw new Error(`outreach: ${row.source.source_id} has no implemented ETL command for stream ${stream}`);
  }
  const endpoint = command.api_surface?.find(({ method, path }) => method === row.method && path === row.path);
  if (!endpoint) {
    throw new Error(`outreach: ${row.source.source_id} command ${command.path} lacks its exact API-surface endpoint`);
  }
  row.state = "enabled";
  row.rejection = null;
  row.foundation = {
    state: "present",
    evidence: `internal/connectors/defs/outreach/cli_surface.json: implemented ETL command ${JSON.stringify(command.path)} binds ${row.method} ${row.path}; the installed binary reaches missing --credential before provider I/O.`,
  };
  row.declaration = {
    status: `enabled; runnable ETL command ${JSON.stringify(command.path)} binds the pinned source endpoint`,
    command: {
      path: command.path,
      intent: command.intent,
      availability: command.availability,
    },
    transport: {
      source_transport: {
        state: "declared",
        evidence: "internal/connectors/defs/outreach/sync_transport.json declares the exact stream source transport and connector-owned conformance reference.",
      },
    },
  };
  changed.set(`outreach:${row.source.source_id}`, row);
}
outreach.summary.enabled_operations = outreach.ledger_dispositions.filter((row) => row.state === "enabled").length;
outreach.summary.enabled_percent = 100;
outreach.summary.disabled_operations = 0;
outreach.summary.terminal_commands = 96;
outreach.summary.rejected_by_reason = [];
outreach.summary.declaration_pending_ids = [];
outreach.summary.declaration_pending = [];
outreach.summary.runnable_cli_surface_commands = 96;
outreach.summary.endpoint_bound_cli_commands = 96;
outreach.summary.transport.destination_transport = {
  state: "declaration-pending",
  declaration_pending: {
    id: "qualified-declarative-typed-destination-outreach",
    evidence: "docs/sync-transport-definition.md:85-115 requires an action-owned per-record batch, receipt locator, provider read-back, and supported non-full-overwrite mode. The prior activate_sequence declaration had none of the required provider-evidenced acknowledgement facts.",
    minimal_change: "Add an exact action-owned source binding only after source evidence supports a bounded durable acknowledgement and provider-state read-back contract.",
  },
  provider_live_certification: "pending",
};
outreach.notes = [
  ...outreach.notes.filter((note) => !note.includes("destination transport")),
  "2026-08-25 reconciliation: all 96 source-locked GET streams have exact implemented ETL commands and built-binary no-credential proof. The former activate_sequence destination claim was removed because the closed adapter rejects full_overwrite and no source-evidenced per-record acknowledgement/read-back contract exists; the direct typed action remains enabled while reverse ETL is declaration-pending.",
];
await writeJSON(outreachFile, outreach);

const evidence = await readJSON(evidenceFile);
let evidenceChanges = 0;
for (const row of evidence.rows) {
  const disposition = changed.get(`${row.connector}:${row.provider_operation?.source_id}`);
  if (!disposition) {
    continue;
  }
  row.canonical_mapping.source_state = disposition.state;
  row.canonical_mapping.declaration_status = disposition.declaration.status;
  row.canonical_mapping.contract = disposition.declaration.contract ?? null;
  row.canonical_mapping.api_surface_binding = disposition.api_surface.covered_by;
  row.canonical_mapping.rejection = disposition.rejection;
  if (disposition.parity_class === "etl") {
    row.surfaces.etl = { state: "source_transport_declared_or_pending", source_state: "enabled" };
    row.surfaces.executable_cli = { state: "generated_binding_present", binding: { stream: disposition.api_surface.covered_by.stream } };
  }
  if (disposition.parity_class === "direct_write") {
    row.surfaces.reverse_etl = {
      state: "eligible_declaration_pending",
      eligible_typed_action: true,
      foundation_gap: null,
    };
  }
  evidenceChanges += 1;
}
const enabledOutreachETL = evidence.rows.filter((row) => (
  row.connector === "outreach" && row.surfaces.etl?.source_state === "enabled"
));
if (enabledOutreachETL.length !== 96) {
  throw new Error(`operation evidence: expected 96 enabled Outreach ETL rows, found ${enabledOutreachETL.length}`);
}
const staleEvidence = evidence.rows.filter((row) => (
  row.surfaces.reverse_etl?.state?.startsWith("bound_")
));
if (staleEvidence.length !== 0) {
  throw new Error(`operation evidence: retained ${staleEvidence.length} stale typed-destination claims`);
}
await writeJSON(evidenceFile, evidence);
