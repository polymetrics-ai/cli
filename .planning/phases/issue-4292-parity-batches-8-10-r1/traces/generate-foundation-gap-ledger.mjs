#!/usr/bin/env node

// Emits the captain-required operation-level shared-foundation gap fan-out.
// This is intentionally separate from declaration-pending connector work: a
// row here is an open provider-neutral capability and never implies disabled
// or N/A provider behavior.
import { readFile, writeFile } from "node:fs/promises";
import { resolve } from "node:path";

const root = resolve(import.meta.dirname, "../../../..");
const phase = resolve(root, ".planning/phases/issue-4292-parity-batches-8-10-r1");
const defs = resolve(root, "internal/connectors/defs");
const output = resolve(phase, "FOUNDATION-GAP-LEDGER.json");
const names = ["brex","zoho-books","testrail","amplitude","posthog","metabase","dbt","looker","mode","dremio","coda","clickup-api","calendly","greenhouse","lever-hiring","ashby","workable","recruitee","hibob","factorial","datadog","pagerduty","auth0","okta","firehydrant","adobe-commerce-magento","commercetools","recharge","docuseal","eventbrite"];
const json = async (path) => JSON.parse(await readFile(path, "utf8"));
const pretty = (value) => `${JSON.stringify(value, null, 2)}\n`;
const batch = (name) => 8 + Math.floor(names.indexOf(name) / 10);
const gap = {
  id: "operation-evidence-projection-v1",
  missing_provider_neutral_capability: "operation-level projection joining provider source trace, canonical mapping, enabled runtime reachability, generated CLI and website rows, and executable fixture/conformance evidence",
  owning_issue_or_lane: "common-foundations / captain pre-merge gate",
  status: "open",
  closure_verification: "run this generator with --check; every provider operation must have enabled runtime, CLI, website, and executable fixture/conformance cells, or a provider-evidenced capability-absent N/A cell",
  failing_validator_or_runtime_evidence: "PREMERGE-GATE.md: current batch ledger has source/canonical mapping and action-level CLI dispositions but lacks the common operation-level runtime, website, and fixture/conformance projection",
};
const inventoryGap = {
  id: "provider-operation-inventory-unavailable-v1",
  missing_provider_neutral_capability: "complete provider operation inventory for a public source that is unavailable or dynamic-instance-dependent",
  owning_issue_or_lane: "common-foundations / captain pre-merge gate",
  status: "open",
  closure_verification: "pin a complete provider operation inventory, regenerate the source disposition, and replace this connector-level inventory row with exact operation rows",
  failing_validator_or_runtime_evidence: "source disposition records a non-enumerable provider surface; an unenumerated provider operation cannot be silently treated as N/A or merge-ready",
};

const rows = [];
for (const connector of names) {
  const sourcePath = resolve(defs, connector, "sources", `${connector}-declaration-disposition.json`);
  const disposition = await json(sourcePath);
  const lock = await json(resolve(defs, connector, disposition.source_basis.source_lock));
  const entries = disposition.ledger_dispositions;
  for (const entry of entries) {
    const source = entry.source ?? {};
    const doc = (lock.rest.documents ?? []).find((candidate) => candidate.source_url === source.source_url);
    const surfaces = [entry.parity_class];
    if (entry.reverse_etl_eligibility?.state === "foundation-gap") surfaces.push("reverse_etl");
    rows.push({
      row_id: `${gap.id}:${connector}:${source.source_id}`,
      gap_id: gap.id,
      connector,
      batch: batch(connector),
      provider_operation: { method: entry.method, path: entry.path, operation_id: source.operation_id, source_id: source.source_id },
      source_trace: {
        url: source.source_url ?? disposition.source_basis.source_url,
        revision: { provider_version: "not-declared-by-pinned-source", captured_at: lock.captured_at },
        sha256: doc?.sha256 ?? lock.rest.source_bundle?.sha256,
        source_location: source.source_location,
      },
      canonical_mapping: { parity_class: entry.parity_class, api_surface: entry.api_surface?.operation?.status ?? "not-mapped" },
      affected_surfaces: surfaces,
      failing_validator_or_runtime_evidence: gap.failing_validator_or_runtime_evidence,
      missing_provider_neutral_capability: gap.missing_provider_neutral_capability,
      owning_issue_or_lane: gap.owning_issue_or_lane,
      status: gap.status,
      closure_verification: gap.closure_verification,
      merge_ready: false,
    });
  }
  if (entries.length === 0) {
    rows.push({
      row_id: `${inventoryGap.id}:${connector}`,
      gap_id: inventoryGap.id,
      connector,
      batch: batch(connector),
      provider_operation: { state: "non-enumerable-provider-inventory" },
      source_trace: { url: disposition.source_basis.source_url, revision: { provider_version: "not-declared-by-pinned-source", captured_at: lock.captured_at }, sha256: lock.rest.source_bundle?.sha256, source_location: disposition.source_basis.source_lock },
      canonical_mapping: { parity_class: "not-enumerable", api_surface: "not-mapped" },
      affected_surfaces: ["etl", "reverse_etl", "direct_read", "direct_write", "binary_download", "binary_upload"],
      failing_validator_or_runtime_evidence: inventoryGap.failing_validator_or_runtime_evidence,
      missing_provider_neutral_capability: inventoryGap.missing_provider_neutral_capability,
      owning_issue_or_lane: inventoryGap.owning_issue_or_lane,
      status: inventoryGap.status,
      closure_verification: inventoryGap.closure_verification,
      merge_ready: false,
    });
  }
}
const batchRollups = [8, 9, 10].map((value) => ({ batch: value, open_gap_rows: rows.filter((row) => row.batch === value).length, affected_connectors: [...new Set(rows.filter((row) => row.batch === value).map((row) => row.connector))].length }));
const catalog = [gap, inventoryGap].map((item) => ({ ...item, fan_out: { affected_operations: rows.filter((row) => row.gap_id === item.id).length, affected_connectors: new Set(rows.filter((row) => row.gap_id === item.id).map((row) => row.connector)).size } }));
const document = { schema_version: 1, issue: 4292, gap_catalog: catalog, rows, rollups: { batches: batchRollups, portfolio: { open_gap_rows: rows.length, affected_connectors: new Set(rows.map((row) => row.connector)).size, merge_ready: false } } };
if (process.argv.includes("--check")) {
  if (await readFile(output, "utf8") !== pretty(document)) throw new Error("foundation gap ledger drift; regenerate it");
  process.stdout.write(`verified ${rows.length} foundation-gap rows\n`);
} else {
  await writeFile(output, pretty(document));
  process.stdout.write(`generated ${rows.length} foundation-gap rows\n`);
}
