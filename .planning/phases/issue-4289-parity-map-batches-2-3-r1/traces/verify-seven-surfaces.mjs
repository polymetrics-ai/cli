#!/usr/bin/env node

import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";

const root = process.cwd();
const phase = path.join(root, ".planning", "phases", "issue-4289-parity-map-batches-2-3-r1");
const connectors = ["grafana", "trello", "slack", "n8n", "google-calendar", "gmail", "twilio", "amazon-sqs", "elasticsearch", "gong", "google-ads", "facebook-marketing", "linkedin-ads", "aircall", "xero", "paypal-transaction", "gocardless", "amazon-seller-partner", "miro"];

async function readJSON(file, fallback) {
  try {
    return JSON.parse(await readFile(file, "utf8"));
  } catch (error) {
    if (error.code === "ENOENT") return fallback;
    throw error;
  }
}

function endpointKey(method, pathname) {
  return `${String(method).toUpperCase()} ${pathname}`;
}

function executableOperation(operations, commands, row, kind) {
  const expected = endpointKey(row.method, row.api_surface.path);
  const operation = operations.find((candidate) =>
    candidate.kind === kind && endpointKey(candidate.rest?.method, candidate.rest?.path) === expected,
  );
  if (!operation) return false;
  return commands.some((command) =>
    command.availability === "implemented" &&
    command.operation === operation.id &&
    command.intent === (kind === "rest_write" ? "direct_write" : "direct_read"),
  );
}

function executableETL(commands, source, row) {
  if (!source) return false;
  const expected = endpointKey(row.method, row.api_surface.path);
  return commands.some((command) =>
    command.availability === "implemented" &&
    command.intent === "etl" &&
    command.api_surface?.some((endpoint) => endpointKey(endpoint.method, endpoint.path) === expected),
  );
}

const ledger = [];
let incomplete = false;
for (const connector of connectors) {
  const dir = path.join(root, "internal", "connectors", "defs", connector);
  const map = await readJSON(path.join(dir, "sources", `${connector}-declaration-disposition.json`), { ledger_dispositions: [] });
  const operations = (await readJSON(path.join(dir, "operations.json"), { operations: [] })).operations;
  const commands = (await readJSON(path.join(dir, "cli_surface.json"), { commands: [] })).commands;
  const transport = await readJSON(path.join(dir, "sync_transport.json"), null);
  const missing = { direct_read: 0, direct_write: 0, etl: 0, binary_read: 0, binary_write: 0 };
  const represented = { direct_read: 0, direct_write: 0, etl: 0, binary_read: 0, binary_write: 0 };

  for (const row of map.ledger_dispositions) {
    let ok = false;
    switch (row.parity_class) {
      case "direct_read":
        ok = executableOperation(operations, commands, row, "rest_read");
        break;
      case "direct_write":
        ok = executableOperation(operations, commands, row, "rest_write");
        break;
      case "etl":
        ok = executableETL(commands, transport?.source_transport, row);
        break;
      case "binary_read":
      case "binary_write":
        ok = operations.some((operation) => endpointKey(operation.binary?.method || operation.file?.method, operation.binary?.path || operation.file?.path) === endpointKey(row.method, row.api_surface.path));
        break;
      default:
        throw new Error(`${connector}: unsupported parity class ${row.parity_class}`);
    }
    (ok ? represented : missing)[row.parity_class]++;
    if (!ok) incomplete = true;
  }
  ledger.push({
    connector,
    source_operations: map.ledger_dispositions.length,
    represented,
    missing,
    reverse_etl: transport?.destination_transport
      ? { declaration: "present", application_dispatch: "foundation-pending" }
      : { declaration: "missing", application_dispatch: "not-applicable" },
    executable_commands: commands.filter((command) => command.availability === "implemented").length,
  });
}

await writeFile(path.join(phase, "SEVEN-SURFACE-LEDGER.json"), `${JSON.stringify({ schema_version: 1, foundation_dispatch: "pending latest fm/cli-reverse-etl-destination-r1", connectors: ledger }, null, 2)}\n`);
if (incomplete) {
  throw new Error("seven-surface reconciliation is incomplete; see SEVEN-SURFACE-LEDGER.json");
}
console.log(`verified ${ledger.length} connectors across direct, binary, ETL, reverse-ETL declaration, and CLI surfaces`);
