#!/usr/bin/env node

// Reverse-ETL preparation for #4292. This deliberately reports the typed
// direct-write actions that a connector already owns without promoting them to
// destinations: #4303 still has to supply the neutral typed-destination
// factory, and each future destination needs its own evidence, field bindings,
// acknowledgement, and sync-mode strategy.

import { readFile, writeFile } from "node:fs/promises";
import { resolve } from "node:path";

const root = resolve(import.meta.dirname, "../../../..");
const defsRoot = resolve(root, "internal/connectors/defs");
const phaseRoot = resolve(import.meta.dirname, "..");
const output = resolve(phaseRoot, "REVERSE-ETL-READINESS.md");
const inventoryOutput = resolve(phaseRoot, "REVERSE-ETL-TYPED-ACTION-INVENTORY.json");
const batches = {
  "8": ["brex", "zoho-books", "testrail", "amplitude", "posthog", "metabase", "dbt", "looker", "mode", "dremio"],
  "9": ["coda", "clickup-api", "calendly", "greenhouse", "lever-hiring", "ashby", "workable", "recruitee", "hibob", "factorial"],
  "10": ["datadog", "pagerduty", "auth0", "okta", "firehydrant", "adobe-commerce-magento", "commercetools", "recharge", "docuseal", "eventbrite"],
};
const gapID = "generic-typed-destination-executor";
const postFoundationRequirements = [
  "connector-owned conformance evidence",
  "explicit source-to-destination field bindings",
  "acknowledgement contract",
  "per-sync-mode apply strategies",
  "connector product-safety decision",
];

function assert(condition, message) {
  if (!condition) throw new Error(message);
}

async function readJSON(path) {
  return JSON.parse(await readFile(path, "utf8"));
}

function sourceState(lock) {
  if (lock.state === "skipped") return "skipped: no public API description";
  if (lock.state === "dynamic") return "dynamic: instance-dependent";
  return "mapped";
}

function count(value) {
  return value === null ? "—" : String(value);
}

function escapeCell(value) {
  return String(value).replaceAll("|", "\\|");
}

function summarize(connector, batch, sourceLock, disposition) {
  if (sourceLock.state === "skipped" || sourceLock.state === "dynamic") {
    assert(disposition.ledger_dispositions.length === 0, `${connector}: unavailable/dynamic source must not have direct-write rows`);
    return {
      batch,
      connector,
      source: sourceState(sourceLock),
      directWrites: null,
      actionBacked: null,
      uniqueActions: null,
      authoringPending: null,
      inventory: [],
    };
  }

  const rows = disposition.ledger_dispositions.filter((row) => row.parity_class === "direct_write");
  const actionBacked = rows.filter((row) => row.declaration?.writes?.length > 0);
  const uniqueActions = new Set(actionBacked.flatMap((row) => row.declaration.writes));
  for (const row of disposition.ledger_dispositions) {
    assert(row.parity_class !== "reverse_etl", `${connector}: reverse_etl may not be a primary endpoint class`);
  }
  for (const row of rows) {
    assert(row.reverse_etl_eligibility?.state === "foundation-gap", `${connector}: direct write has no foundation-gated reverse-ETL eligibility`);
    assert(row.reverse_etl_eligibility.foundation_gap?.id === gapID, `${connector}: direct write has the wrong reverse-ETL gap`);
    assert(!JSON.stringify(row).includes("transport_binding"), `${connector}: direct-write row invents a transport binding`);
  }
  for (const row of actionBacked) {
    assert(row.state === "enabled", `${connector}: typed-action-backed direct write must be enabled`);
  }
  const enabled = rows.filter((row) => row.state === "enabled");
  assert(enabled.length === actionBacked.length, `${connector}: an enabled direct write has no named typed action`);
  const inventory = actionBacked.map((row) => ({
    source_id: row.source.source_id,
    method: row.method,
    path: row.path,
    operation_id: row.source.operation_id,
    typed_action_ids: row.declaration.writes,
  }));

  return {
    batch,
    connector,
    source: sourceState(sourceLock),
    directWrites: rows.length,
    actionBacked: actionBacked.length,
    uniqueActions: uniqueActions.size,
    authoringPending: rows.length - actionBacked.length,
    inventory,
  };
}

function renderInventory(rows) {
  const inventory = rows.flatMap((row) => row.inventory);
  const expected = rows.reduce((sum, row) => sum + (row.actionBacked ?? 0), 0);
  assert(inventory.length === expected, `typed-action inventory count drift (${inventory.length} != ${expected})`);
  const byConnector = rows
    .filter((row) => row.inventory.length > 0)
    .map((row) => ({
      batch: row.batch,
      connector: row.connector,
      disposition_ledger: `internal/connectors/defs/${row.connector}/sources/${row.connector}-declaration-disposition.json`,
      operations: row.inventory,
    }));
  return `${JSON.stringify({
    schema_version: 1,
    status: "pre-foundation",
    purpose: "Exact source operation to typed direct-write action inventory for post-#4303 reverse-ETL destination preparation. This is not a destination declaration and does not establish product safety.",
    prohibited: ["transport_binding", "generic HTTP write executor"],
    current_reverse_etl_status: {
      state: "foundation-gap",
      foundation_gap_id: gapID,
      destination_declarations: "none",
    },
    post_foundation_declaration_requirements: postFoundationRequirements,
    typed_action_operation_bindings_by_connector: byConnector,
  }, null, 2)}\n`;
}

function render(rows) {
  const mapped = rows.filter((row) => row.directWrites !== null);
  const total = (key) => mapped.reduce((sum, row) => sum + row[key], 0);
  const lines = [
    "# Reverse-ETL readiness — issue #4292",
    "",
    "This is a preparation audit, not a destination declaration. Issue #4303 must first add the connector-neutral typed destination factory. Until then, every source-backed `direct_write` below retains `generic-typed-destination-executor`; a named typed action is necessary but insufficient for reverse ETL.",
    "",
    "A future connector-owned destination must additionally declare per-connector conformance evidence, explicit source-to-destination field bindings, acknowledgement, and per-sync-mode apply strategies. No `transport_binding` is declared here.",
    "",
    "## Batch 8–10 readiness",
    "",
    "| Batch | Connector | Provider source state | Direct-write operations | Action-backed rows | Unique typed actions | Typed-action authoring pending | Destination now |",
    "| --- | --- | --- | ---: | ---: | ---: | ---: | --- |",
  ];
  for (const row of rows) {
    lines.push(`| ${row.batch} | ${escapeCell(row.connector)} | ${escapeCell(row.source)} | ${count(row.directWrites)} | ${count(row.actionBacked)} | ${count(row.uniqueActions)} | ${count(row.authoringPending)} | none — #4303 foundation gap |`);
  }
  lines.push(
    `| **Total mapped** | **${mapped.length} connectors** | **${rows.length - mapped.length} skipped/dynamic** | **${total("directWrites")}** | **${total("actionBacked")}** | **${total("uniqueActions")}** | **${total("authoringPending")}** | **none** |`,
    "",
    "`Action-backed rows` are direct-write operations whose pinned route is already bound to a named typed `writes.json` action. `REVERSE-ETL-TYPED-ACTION-INVENTORY.json` preserves the exact source ID, route, source location, and action IDs for each of those 1,419 rows. `Typed-action authoring pending` is the remaining source-backed direct-write inventory; it is connector-local contract/safety/fixture work, not a shared engine gap. Neither artifact decides whether an operation is product-safe for reverse ETL.",
    "",
    "## Zoom critical-path preparation",
    "",
    "The captain directed that Zoom's 204-action destination cohort must gain destination declarations after #4303. That cohort is outside #4292's batch source inventory, so this report records **204 as a captain-directed planning target, not a provider-source operation total**. Current Zoom bundle evidence is read-only: `internal/connectors/defs/zoom/docs.md` records no declared Zoom write action. Once #4303 supplies the declaration schema, the Zoom lane must first identify the exact named typed action IDs and their pinned operations, then add the required per-connector evidence, explicit bindings, acknowledgement, and mode strategies. It must not add a `transport_binding` in advance.",
    "",
    "## Regeneration and assertions",
    "",
    "```bash",
    "node .planning/phases/issue-4292-parity-batches-8-10-r1/traces/audit-reverse-etl-readiness.mjs",
    "```",
    "",
    "The script fails if a source-backed direct write lacks the locked foundation gap, an enabled direct write lacks a named typed action, any row uses primary `reverse_etl`, or a `transport_binding` appears.",
  );
  return `${lines.join("\n")}\n`;
}

const rows = [];
for (const [batch, connectors] of Object.entries(batches)) {
  for (const connector of connectors) {
    const sourceDir = resolve(defsRoot, connector, "sources");
    const [sourceLock, disposition] = await Promise.all([
      readJSON(resolve(sourceDir, `${connector}-operation-source-lock.json`)),
      readJSON(resolve(sourceDir, `${connector}-declaration-disposition.json`)),
    ]);
    rows.push(summarize(connector, batch, sourceLock, disposition));
  }
}

const rendered = render(rows);
const renderedInventory = renderInventory(rows);
if (process.argv.includes("--check")) {
  const existing = await readFile(output, "utf8");
  const existingInventory = await readFile(inventoryOutput, "utf8");
  assert(existing === rendered, `${output}: readiness report drift; rerun this script without --check`);
  assert(existingInventory === renderedInventory, `${inventoryOutput}: typed-action inventory drift; rerun this script without --check`);
  process.stdout.write(`verified ${output} and ${inventoryOutput}\n`);
} else {
  await Promise.all([writeFile(output, rendered), writeFile(inventoryOutput, renderedInventory)]);
  process.stdout.write(`wrote ${output} and ${inventoryOutput}\n`);
}
