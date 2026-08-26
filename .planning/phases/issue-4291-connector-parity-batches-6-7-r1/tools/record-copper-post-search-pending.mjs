#!/usr/bin/env node

// Copper's five public list endpoints are POST search requests. Its legacy
// stream aliases are GET /__legacy_hook paths, so this records the exact
// source-backed declaration work instead of turning a POST/body contract into
// a false GET command.

import { readFile, writeFile } from "node:fs/promises";
import { resolve } from "node:path";

const phase = resolve(import.meta.dirname, "..");
const root = resolve(phase, "../../..");
const connector = "copper";
const file = resolve(root, "internal/connectors/defs", connector, "sources", `${connector}-declaration-disposition.json`);
const evidenceFile = resolve(phase, "OPERATION-SURFACE-EVIDENCE.json");
const plans = new Map([
  ["copper.rest.post.V1PeopleSearch70", "people search"],
  ["copper.rest.post.V1CompaniesSearch53", "companies search"],
  ["copper.rest.post.V1OpportunitiesSearch64", "opportunities search"],
  ["copper.rest.post.V1LeadsSearch61", "leads search"],
  ["copper.rest.post.V1TasksSearch75", "tasks search"],
]);

async function readJSON(path) { return JSON.parse(await readFile(path, "utf8")); }
async function writeJSON(path, value) { await writeFile(path, `${JSON.stringify(value, null, 2)}\n`); }

const disposition = await readJSON(file);
let changed = 0;
for (const row of disposition.ledger_dispositions) {
  const path = plans.get(row.source.source_id);
  if (!path) continue;
  const reason = "The documented source operation is a POST search. streams.json currently routes this alias through GET /__legacy_hook and has no pinned body/query contract, so it cannot faithfully become an ETL CLI command yet.";
  row.declaration = {
    ...row.declaration,
    status: `disabled; declaration-pending intended ETL command ${JSON.stringify(path)}`,
    command: {
      state: "declaration-pending",
      intended_path: path,
      intent: "etl",
      lane: "etl",
      source_cli_path: `${row.method} ${row.path}`,
      source_url: row.source.source_url,
      execution_component: `internal/connectors/defs/copper/streams.json: replace legacy ${JSON.stringify(row.api_surface.covered_by.stream)} GET hook with a source-backed POST search stream and declared request body/query schema`,
      reason,
      engine_support: "internal/connectors/engine/bundle.go:296-301 declares StreamSpec.Method and StreamSpec.Body for POST-body streams; no shared engine change is required.",
    },
  };
  row.rejection = { ...row.rejection, detail: reason };
  changed += 1;
}
if (changed !== plans.size) throw new Error(`updated ${changed} Copper search rows, expected ${plans.size}`);
disposition.notes = [
  ...(disposition.notes ?? []).filter((note) => !note.includes("2026-08-26 Copper POST-search pending mapping")),
  "2026-08-26 Copper POST-search pending mapping: all five source-locked search operations retain canonical command paths, citations, and the legacy-hook-to-POST-stream component gap. StreamSpec already supports POST/body streams; request contract authoring is connector work, not a foundation gap.",
];
await writeJSON(file, disposition);

const evidence = await readJSON(evidenceFile);
let evidenceChanged = 0;
for (const row of evidence.rows) {
  if (row.connector !== connector) continue;
  const path = plans.get(row.provider_operation?.source_id);
  if (!path) continue;
  row.canonical_mapping.declaration_status = `disabled; declaration-pending intended ETL command ${JSON.stringify(path)}`;
  row.canonical_mapping.contract = {
    execution_component: "streams.json POST search declaration replacing the legacy hook",
    source_cli_path: `${row.provider_operation.method} ${row.provider_operation.path}`,
    source_url: row.source.url,
  };
  row.surfaces.etl = { state: "declaration_pending_source_backed_post_search_stream", source_state: "disabled" };
  row.surfaces.executable_cli = {
    state: "declaration_pending_source_template_contract",
    binding: { intended_path: path, intent: "etl" },
    reason: "A source-backed POST search request/body contract must replace the legacy GET hook before this command can execute.",
  };
  evidenceChanged += 1;
}
if (evidenceChanged !== plans.size) throw new Error(`updated ${evidenceChanged} Copper evidence rows, expected ${plans.size}`);
await writeJSON(evidenceFile, evidence);

console.log(`recorded ${changed} Copper POST-search CLI declaration-pending rows`);
