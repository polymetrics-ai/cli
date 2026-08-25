#!/usr/bin/env node

// ServiceNow's provider boundary is six Table API templates, not a finite set
// of customer tables. This ledger pass makes the four currently unauthored
// templates explicit without manufacturing table names, schemas, or actions.

import { readFile, writeFile } from "node:fs/promises";
import { resolve } from "node:path";

const phase = resolve(import.meta.dirname, "..");
const root = resolve(phase, "../../..");
const connector = "service-now";
const source = "https://www.servicenow.com/docs/bundle/zurich-api-reference/page/integrate/inbound-rest/concept/c_TableAPI.html";
const file = resolve(root, "internal/connectors/defs", connector, "sources", `${connector}-declaration-disposition.json`);
const apiFile = resolve(root, "internal/connectors/defs", connector, "api_surface.json");
const evidenceFile = resolve(phase, "OPERATION-SURFACE-EVIDENCE.json");

const pending = new Map([
  ["service-now.table_api.get.0", {
    path: "tables list",
    intent: "etl",
    component: "internal/connectors/defs/service-now/streams.json: fixed incidents/users/groups aliases do not expose the documented table_name template as one closed user-selected stream",
    reason: "The provider table_name, record schema, and ACL-visible fields are instance-dependent. Add a connector-owned dynamic-table selection and schema-discovery contract before making the generic template executable; no engine gap is claimed.",
  }],
  ["service-now.table_api.get.1", {
    path: "tables get",
    intent: "direct_read",
    component: "internal/connectors/defs/service-now/operations.json: absent fixed direct-read operation contract for the table_name/sys_id template",
    reason: "A direct command needs a fixed table identity plus source-backed request/response schema. The public template cannot supply a finite instance table schema, so this remains connector declaration work.",
  }],
  ["service-now.table_api.put.3", {
    path: "tables replace apply",
    intent: "reverse_etl",
    component: "internal/connectors/defs/service-now/writes.json: no typed replace action for the table_name/sys_id template",
    reason: "Add a connector-owned typed action only for a source-backed fixed table/schema selection; do not infer customer field contracts from the public template.",
  }],
  ["service-now.table_api.delete.5", {
    path: "tables delete apply",
    intent: "reverse_etl",
    component: "internal/connectors/defs/service-now/writes.json: no typed delete action for the table_name/sys_id template",
    reason: "Add a connector-owned typed action only for a source-backed fixed table/schema selection. Destructive confirmation is execution metadata, not a reason to omit the documented DELETE template.",
  }],
]);

async function readJSON(path) { return JSON.parse(await readFile(path, "utf8")); }
async function writeJSON(path, value) { await writeFile(path, `${JSON.stringify(value, null, 2)}\n`); }

const disposition = await readJSON(file);
let changed = 0;
for (const row of disposition.ledger_dispositions) {
  const plan = pending.get(row.source.source_id);
  if (!plan) continue;
  row.declaration = {
    ...row.declaration,
    status: `disabled; declaration-pending intended ${plan.intent} CLI command ${JSON.stringify(plan.path)}`,
    command: {
      state: "declaration-pending",
      intended_path: plan.path,
      intent: plan.intent,
      lane: row.parity_class,
      source_cli_path: `${row.method} ${row.path}`,
      source_url: source,
      execution_component: plan.component,
      reason: plan.reason,
    },
  };
  row.rejection = {
    ...row.rejection,
    detail: plan.reason,
  };
  changed += 1;
}
if (changed !== pending.size) throw new Error(`updated ${changed} ServiceNow templates, expected ${pending.size}`);
disposition.notes = [
  ...(disposition.notes ?? []).filter((note) => !note.includes("2026-08-25 ServiceNow template CLI pending mapping")),
  "2026-08-25 ServiceNow template CLI pending mapping: the fixed GET/PUT/DELETE templates retain intended command paths and concrete missing components. Instance table names, fields, and ACLs stay dynamic-schema declaration work; no customer operation identity is invented.",
];
await writeJSON(file, disposition);

// One public Table API template may back the existing incident, user, and
// group contracts. The plural coverage form is the canonical representation:
// duplicate method/path rows collapse neither a provider operation nor any
// typed action, while the CLI safety map can see every valid action.
const apiSurface = await readJSON(apiFile);
const sharedTemplateWrites = new Map([
  ["POST /api/now/table/{table_name}", ["create_incident", "create_user", "create_group"]],
  ["PATCH /api/now/table/{table_name}/{sys_id}", ["update_incident", "update_user", "update_group"]],
]);
for (const [key, targets] of sharedTemplateWrites) {
  const [method, path] = key.split(" ");
  const matching = apiSurface.endpoints.filter((endpoint) => endpoint.method === method && endpoint.path === path);
  const observed = matching.map((endpoint) => endpoint.covered_by?.write).sort();
  if (matching.length !== targets.length || JSON.stringify(observed) !== JSON.stringify([...targets].sort())) {
    throw new Error(`${connector}:${key}: expected exactly the source-bound write projections ${targets.join(", ")}`);
  }
  const canonical = {
    ...matching[0],
    covered_by: { writes: targets },
  };
  let emitted = false;
  apiSurface.endpoints = apiSurface.endpoints.flatMap((endpoint) => {
    if (endpoint.method !== method || endpoint.path !== path) return [endpoint];
    if (emitted) return [];
    emitted = true;
    return [canonical];
  });
}
await writeJSON(apiFile, apiSurface);

const evidence = await readJSON(evidenceFile);
let evidenceChanged = 0;
for (const row of evidence.rows) {
  if (row.connector !== connector) continue;
  const plan = pending.get(row.provider_operation?.source_id);
  if (!plan) continue;
  row.canonical_mapping.declaration_status = `disabled; declaration-pending intended ${plan.intent} CLI command ${JSON.stringify(plan.path)}`;
  row.canonical_mapping.contract = {
    execution_component: plan.component,
    source_cli_path: `${row.provider_operation.method} ${row.provider_operation.path}`,
    source_url: source,
  };
  row.surfaces.executable_cli = {
    state: "declaration_pending_source_template_contract",
    binding: { intended_path: plan.path, intent: plan.intent },
    reason: plan.reason,
  };
  evidenceChanged += 1;
}
if (evidenceChanged !== pending.size) throw new Error(`updated ${evidenceChanged} operation-evidence rows, expected ${pending.size}`);
await writeJSON(evidenceFile, evidence);

console.log(`recorded ${changed} ServiceNow fixed-template CLI declaration-pending rows and normalized two shared template projections`);
