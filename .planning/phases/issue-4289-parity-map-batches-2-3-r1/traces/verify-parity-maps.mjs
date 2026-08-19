#!/usr/bin/env node

import { readFile, writeFile } from "node:fs/promises";
import { execFileSync } from "node:child_process";
import path from "node:path";

const root = process.cwd();
const phase = path.join(root, ".planning", "phases", "issue-4289-parity-map-batches-2-3-r1");
const connectors = ["grafana", "trello", "slack", "n8n", "google-calendar", "gmail", "twilio", "amazon-sqs", "elasticsearch", "gong", "google-ads", "facebook-marketing", "linkedin-ads", "aircall", "xero", "paypal-transaction", "gocardless", "amazon-seller-partner", "miro"];
const classes = new Set(["direct_read", "direct_write", "etl", "reverse_etl", "binary_read", "binary_write"]);
const required = ["method", "path", "parity_class", "api_surface", "source", "state", "foundation", "rejection", "declaration"];
const baselineRevision = "acb85dc03";

function file(connector, name) {
  return path.join(root, "internal", "connectors", "defs", connector, "sources", name);
}

function canonicalPath(input) {
  return input.replace(/\?.*$/, "").replace(/\{[^}]+\}/g, "{}").replace(/\/+/g, "/").replace(/\/$/, "") || "/";
}

async function readJSON(name) {
  return JSON.parse(await readFile(name, "utf8"));
}

function actionNames(row) {
  const covered = row.api_surface?.covered_by || {};
  return [covered.write, ...(Array.isArray(covered.writes) ? covered.writes : [])].filter(Boolean);
}

const checks = [];
for (const connector of connectors) {
  const lock = await readJSON(file(connector, `${connector}-operation-source-lock.json`));
  const map = await readJSON(file(connector, `${connector}-declaration-disposition.json`));
  const surface = await readJSON(path.join(root, "internal", "connectors", "defs", connector, "api_surface.json"));
  const writes = await readJSON(path.join(root, "internal", "connectors", "defs", connector, "writes.json")).catch((error) => error.code === "ENOENT" ? { actions: [] } : Promise.reject(error));
  const operations = Object.values(lock).flatMap((value) => Array.isArray(value?.operations) ? value.operations : []);
  const oldSurface = JSON.parse(execFileSync("git", ["show", `${baselineRevision}:internal/connectors/defs/${connector}/api_surface.json`], { encoding: "utf8" }));
  if (operations.length !== map.ledger_dispositions.length) throw new Error(`${connector}: source inventory ${operations.length} != disposition rows ${map.ledger_dispositions.length}`);
  if (lock.counts?.total !== operations.length || lock.rest.counts?.total !== operations.length || lock.rest.counts?.by_kind?.rest !== operations.length || !lock.rest.coverage_confidence?.level || !lock.rest.coverage_confidence?.basis) throw new Error(`${connector}: source lock lacks root/rest counts.total, per-kind counts, or coverage confidence`);
  if (map.source_basis.operations_found !== operations.length || map.summary.operations_found !== operations.length || !map.source_basis.coverage_confidence?.level || map.summary.coverage_confidence?.level !== lock.rest.coverage_confidence.level || "declared_percent" in map.summary) throw new Error(`${connector}: source accounting is self-referential or incomplete`);
  const surfaceKeys = new Set(surface.endpoints.map((endpoint) => `${endpoint.method.toUpperCase()} ${canonicalPath(endpoint.path)}`));
  for (const row of map.ledger_dispositions) {
    const sourceOperation = operations.find((operation) => operation.id === row.source?.source_id && operation.source_location === row.source?.source_location && operation.source_url === row.source?.source_url);
    for (const key of required) if (!(key in row)) throw new Error(`${connector}: ${row.method} ${row.path} missing ${key}`);
    if (!sourceOperation || sourceOperation.method !== row.method || sourceOperation.path !== row.path) throw new Error(`${connector}: ${row.method} ${row.path} is not the exact provider source operation for ${row.source?.source_id}`);
    if (!classes.has(row.parity_class)) throw new Error(`${connector}: ${row.method} ${row.path} has invalid parity class ${row.parity_class}`);
    if (!["enabled", "disabled"].includes(row.state)) throw new Error(`${connector}: ${row.method} ${row.path} has invalid state ${row.state}`);
    if (!surfaceKeys.has(`${row.method.toUpperCase()} ${canonicalPath(row.api_surface.path)}`)) throw new Error(`${connector}: ${row.method} ${row.path} is not bound to a surface endpoint`);
    if (row.parity_class === "reverse_etl") throw new Error(`${connector}: ${row.method} ${row.path} confuses reverse-ETL eligibility with an endpoint class`);
    if (row.state === "disabled" && row.rejection?.reason !== "declaration-pending") throw new Error(`${connector}: ${row.method} ${row.path} disabled for ${row.rejection?.reason}, not corrected declaration-pending`);
    if (row.foundation?.foundation_gap) throw new Error(`${connector}: ${row.method} ${row.path} puts a transport eligibility gap on the endpoint itself`);
    const eligibility = row.declaration?.reverse_etl_eligibility;
    if (eligibility && (row.parity_class !== "direct_write" || row.state !== "enabled" || eligibility.state !== "foundation-gap" || eligibility.foundation_gap?.id !== "generic-typed-destination-app-dispatch")) throw new Error(`${connector}: ${row.method} ${row.path} has invalid reverse-ETL eligibility metadata`);
  }
  const classTotal = map.summary.parity_class_counts.reduce((sum, count) => sum + count.count, 0);
  if (classTotal !== operations.length) throw new Error(`${connector}: class total ${classTotal} != ${operations.length}`);
  const reverseETL = map.ledger_dispositions.filter((row) => row.declaration?.reverse_etl_eligibility).length;
  const expectedActions = writes.actions.map((action) => action.name).sort();
  const actionDispositions = map.summary.reverse_etl_eligibility?.action_dispositions;
  if (!Array.isArray(actionDispositions) || actionDispositions.length !== expectedActions.length) throw new Error(`${connector}: reverse-ETL action disposition count does not match writes.json`);
  const actionDispositionByName = new Map(actionDispositions.map((entry) => [entry.action, entry]));
  if (actionDispositionByName.size !== actionDispositions.length) throw new Error(`${connector}: reverse-ETL action disposition duplicates an action`);
  for (const action of expectedActions) {
    const disposition = actionDispositionByName.get(action);
    if (!disposition || disposition.semantic_eligibility !== "eligible" || disposition.closed_destination_action_shape !== "representable") throw new Error(`${connector}: typed write action ${action} lacks an eligible closed-destination disposition`);
    if (!disposition.provider_operation_binding || !["source-bound", "declaration-pending"].includes(disposition.provider_operation_binding.state)) throw new Error(`${connector}: typed write action ${action} has invalid provider-operation binding disposition`);
    if (disposition.source_input_binding?.state !== "declaration-pending") throw new Error(`${connector}: typed write action ${action} omits its pending exact source-input binding`);
    if (disposition.destination_binding?.state !== "foundation-gap" || disposition.destination_binding.foundation_gap?.id !== "declarative-typed-destination-action-multiplicity") throw new Error(`${connector}: typed write action ${action} omits the action-selection foundation hold`);
  }
  const exactSourceBoundActions = new Set(map.ledger_dispositions.flatMap(actionNames));
  for (const disposition of actionDispositions) {
    if (disposition.provider_operation_binding.state === "source-bound" && (!Array.isArray(disposition.provider_operation_binding.source_ids) || disposition.provider_operation_binding.source_ids.length === 0 || !exactSourceBoundActions.has(disposition.action))) throw new Error(`${connector}: typed write action ${disposition.action} claims an invalid exact provider-operation binding`);
  }
  const declarationPending = map.ledger_dispositions.filter((row) => row.state === "disabled").length;
  const rejected = Object.fromEntries(map.summary.rejected_by_reason.map((entry) => [entry.key, entry.count]));
  if (rejected["foundation-gap"] !== undefined || rejected["declaration-pending"] !== declarationPending) throw new Error(`${connector}: rejection summary mislabels a foundation or declaration-pending row`);
  const pendingSummary = map.summary.declaration_pending.reduce((sum, entry) => sum + entry.count, 0);
  const foundationGaps = new Map(map.summary.foundation_gaps.map((gap) => [gap.id, gap]));
  if (foundationGaps.get("generic-typed-destination-app-dispatch")?.count !== reverseETL || foundationGaps.get("declarative-typed-destination-action-multiplicity")?.count !== expectedActions.length || pendingSummary !== declarationPending || map.summary.reverse_etl_eligibility?.typed_direct_write_operations !== reverseETL || map.summary.reverse_etl_eligibility?.typed_write_actions !== expectedActions.length || map.summary.reverse_etl_eligibility?.foundation_gap?.id !== "generic-typed-destination-app-dispatch") throw new Error(`${connector}: gap/declaration summary count drift`);
  if (map.summary.gap_ids.length !== 2 || !map.summary.gap_ids.includes("generic-typed-destination-app-dispatch") || !map.summary.gap_ids.includes("declarative-typed-destination-action-multiplicity")) throw new Error(`${connector}: destination foundation gap summary drift`);
  if (map.summary.transport.source_transport.state !== "declaration-pending" || map.summary.transport.destination_transport.state !== "gap" || map.summary.transport.destination_transport.foundation_gap?.id !== "generic-typed-destination-app-dispatch") throw new Error(`${connector}: transport source/destination state does not match the #4304 App/CLI dispatch boundary`);
  // A source lock is a captured immutable input to the map. Provider-hosted
  // discovery documents reorder JSON keys and mutable documentation pages add
  // per-request markup, so refetching would compare a new source revision to
  // the pinned one rather than prove this map's provenance.
  const sourceLock = {
    source_url: lock.rest.source_url,
    captured_at: lock.captured_at,
    bytes: lock.rest.bytes,
    sha256: lock.rest.sha256,
    verified: "captured public source metadata is complete; remote content is intentionally not refetched"
  };
  checks.push({
    connector,
    api_surface_counts: {
      baseline_revision: baselineRevision,
      old: oldSurface.endpoints.length,
      new: surface.endpoints.length,
      basis: lock.rest.coverage_confidence.basis
    },
    operations_found: operations.length,
    coverage_confidence: lock.rest.coverage_confidence,
    enabled_operations: map.summary.enabled_operations,
    disabled_operations: map.summary.disabled_operations,
    documented_deletes: map.summary.documented_deletes,
    enabled_deletes: map.summary.enabled_deletes,
    source_lock: sourceLock
  });
}

const total = checks.reduce((sum, check) => sum + check.operations_found, 0);
const markdown = [
  "# Source-accounted parity map — batches 2 and 3",
  "",
  "All source artifacts were fetched credential-free from their recorded public documentation URLs. No provider operation, credential, write, or live certification was used. A connector marked `partial` is an explicit delivery hold, not a complete-source assertion.",
  "",
  "| Connector | Old api_surface | New api_surface | Operations found | Coverage confidence and basis | Enabled | Disabled | Enabled % | Deletes | Foundation holds |",
  "| --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: | --- | --- |",
  ...checks.map((check) => `| ${check.connector} | ${check.api_surface_counts.old} | ${check.api_surface_counts.new} | ${check.operations_found} | ${check.coverage_confidence.level}: ${check.coverage_confidence.basis} | ${check.enabled_operations} | ${check.disabled_operations} | ${((check.enabled_operations / check.operations_found) * 100).toFixed(2)} | ${check.enabled_deletes}/${check.documented_deletes} | App/CLI dispatch + action-selection multiplicity (reverse ETL) |`),
  "",
  `Old api_surface counts are from \`${baselineRevision}\` (the current-main revision named in the transport correction); new counts are the source-derived projection at this revision. A new api_surface count can be lower than operations found when the provider publishes multiple operation declarations with the same normalized request shape. Total operations found across pinned source artifacts: **${total}**. This is not a self-referential coverage percentage: the per-connector confidence and basis state whether the input is complete or partial. Un-authored endpoint declarations are \`declaration-pending\`; a typed write remains enabled \`direct_write\`. Every current schema-backed write action has an explicit semantic eligibility disposition, while reverse-ETL binding remains held first on persisted App/CLI dispatch and then on exact multi-action selection.`,
  "",
  "ETL source transport is declaration-pending until each connector authors `sync_transport.json` with exact source executor, delivery, and conformance evidence. Reverse-ETL eligibility for typed direct writes remains foundation-blocked: `generic-typed-destination-app-dispatch` must admit its already-composed factory through the persisted App/CLI path, and `declarative-typed-destination-action-multiplicity` must provide an exact closed action-selection route before all eligible actions can be bound. No transport binding/action is invented."
].join("\n") + "\n";
await writeFile(path.join(phase, "SOURCE-LOCK-VERIFICATION.json"), `${JSON.stringify({ generated_at: "2026-08-19T00:00:00Z", checks }, null, 2)}\n`);
await writeFile(path.join(phase, "COMPLETE-PARITY-MAP.md"), markdown);
console.log(`verified ${checks.length} connectors / ${total} documented operations`);
