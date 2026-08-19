#!/usr/bin/env node

// Integrity assertions for the #4292 declaration-map artifacts. The checks
// intentionally exercise the three generated artifacts together so a count
// cannot claim parity while one of the source lock, crosswalk, or ledger has
// silently drifted.

import { readFile } from "node:fs/promises";
import { resolve } from "node:path";

const root = resolve(import.meta.dirname, "../../../..");
const defsRoot = resolve(root, "internal/connectors/defs");

const batches = {
  "8": ["brex", "zoho-books", "testrail", "amplitude", "posthog", "metabase", "dbt", "looker", "mode", "dremio"],
  "9": ["coda", "clickup-api", "calendly", "greenhouse", "lever-hiring", "ashby", "workable", "recruitee", "hibob", "factorial"],
  "10": ["datadog", "pagerduty", "auth0", "okta", "firehydrant", "adobe-commerce-magento", "commercetools", "recharge", "docuseal", "eventbrite"],
};

const classKeys = new Set(["direct_read", "direct_write", "etl", "reverse_etl", "binary_read", "binary_write"]);
const reverseETLGapID = "generic-typed-destination-executor";
const reverseETLMinimalChange = "register a connector-neutral typed destination DefinitionFactory selected by the definition, with per-connector evidence, explicit source bindings, acknowledgement and per-mode apply strategies";

function assert(condition, message) {
  if (!condition) throw new Error(message);
}

async function readJSON(path) {
  return JSON.parse(await readFile(path, "utf8"));
}

function ids(items, idFor, label, connector) {
  const values = items.map(idFor);
  assert(new Set(values).size === values.length, `${connector}: duplicated ${label}`);
  return values;
}

function sameIDs(left, right, label, connector) {
  assert(left.length === right.length, `${connector}: ${label} count differs (${left.length} != ${right.length})`);
  for (const value of left) assert(right.includes(value), `${connector}: ${label} missing ${value}`);
}

function verifySkipped(connector, sourceLock, crosswalk, disposition) {
  assert(sourceLock.state === "skipped", `${connector}: source lock must be skipped`);
  assert(sourceLock.skip?.reason === "no-public-api-description", `${connector}: unexpected skip reason`);
  assert(sourceLock.skip?.retrieval_method === "browser", `${connector}: skip must retain browser retrieval method`);
  assert(!("sha256" in (sourceLock.rest ?? {})), `${connector}: skipped source must not fabricate a SHA-256 pin`);
  assert(crosswalk.state === "skipped" && crosswalk.reason === sourceLock.skip.reason, `${connector}: skipped crosswalk mismatch`);
  assert(disposition.summary?.state === "skipped" && disposition.summary.reason === sourceLock.skip.reason, `${connector}: skipped disposition mismatch`);
  assert(disposition.ledger_dispositions?.length === 0, `${connector}: skipped source must not fabricate operation rows`);
}

function verifyMapped(connector, sourceLock, crosswalk, disposition) {
  const source = sourceLock.rest;
  assert(typeof source?.source_url === "string" && source.source_url.length > 0, `${connector}: source URL missing`);
  assert(/^[a-f0-9]{64}$/.test(source.sha256 ?? ""), `${connector}: invalid source SHA-256`);
  assert(Number.isInteger(source.bytes) && source.bytes > 0, `${connector}: invalid source byte count`);
  assert(["http", "browser"].includes(source.retrieval_method), `${connector}: retrieval method missing`);
  assert(typeof source.representation === "string" && source.representation.length > 0, `${connector}: source representation missing`);

  const lockIDs = ids(source.operations ?? [], (operation) => operation.id, "source-lock operation ID", connector);
  const crosswalkIDs = ids(crosswalk.source_operations ?? [], (operation) => operation.source_id, "crosswalk source ID", connector);
  const ledger = disposition.ledger_dispositions ?? [];
  const ledgerIDs = ids(ledger, (row) => row.source?.source_id, "ledger source ID", connector);
  sameIDs(lockIDs, crosswalkIDs, "source lock / crosswalk IDs", connector);
  sameIDs(lockIDs, ledgerIDs, "source lock / ledger IDs", connector);
  assert(lockIDs.length > 0, `${connector}: operation map must not be empty`);

  assert(disposition.source_basis?.source_sha256 === source.sha256, `${connector}: disposition SHA-256 drift`);
  assert(disposition.source_basis?.source_bytes === source.bytes, `${connector}: disposition byte-count drift`);
  assert(disposition.source_basis?.source_operation_count === lockIDs.length, `${connector}: disposition operation count drift`);
  assert(disposition.summary?.api_surface_rows === lockIDs.length, `${connector}: api-surface count drift`);
  assert(disposition.summary?.exact_source_rows === lockIDs.length, `${connector}: exact-source count drift`);
  assert(disposition.summary?.declared_operations === lockIDs.length, `${connector}: declared-operation count drift`);

  const counted = new Map();
  for (const row of ledger) {
    assert(classKeys.has(row.parity_class), `${connector}: unknown parity class ${row.parity_class}`);
    assert(row.parity_class !== "reverse_etl", `${connector}: reverse_etl cannot be a primary endpoint parity class`);
    assert(typeof row.method === "string" && typeof row.path === "string", `${connector}: row is missing method/path`);
    assert(row.source?.operation_id === row.source?.source_id, `${connector}: source and operation IDs must agree`);
    assert(row.source?.source_url === source.source_url, `${connector}: ledger source URL drift`);
    assert(row.source?.source_location?.startsWith("api_surface.json:endpoints["), `${connector}: untraceable source location`);
    if (row.parity_class === "direct_write") {
      const eligibility = row.reverse_etl_eligibility;
      assert(eligibility?.state === "foundation-gap", `${connector}: direct write lacks reverse ETL foundation-gap state`);
      assert(eligibility.foundation_gap?.id === reverseETLGapID, `${connector}: direct write has wrong reverse ETL foundation gap`);
      assert(eligibility.foundation_gap?.evidence?.startsWith("internal/app/issue_label_warehouse_transport.go:85-95"), `${connector}: direct write has wrong reverse ETL evidence`);
      assert(eligibility.foundation_gap?.minimal_change === reverseETLMinimalChange, `${connector}: direct write has wrong reverse ETL minimal change`);
      if (row.declaration?.write) {
        assert(row.state === "enabled", `${connector}: typed write action must be enabled direct_write`);
        assert(row.foundation?.contract?.kind === "typed_write_action", `${connector}: typed write action lacks typed contract`);
      }
    } else {
      assert(!("reverse_etl_eligibility" in row), `${connector}: only direct_write may carry reverse ETL eligibility`);
    }
    counted.set(row.parity_class, (counted.get(row.parity_class) ?? 0) + 1);
  }

  const summaryCounts = new Map((disposition.summary?.parity_class_counts ?? []).map(({ key, count }) => [key, count]));
  for (const key of classKeys) assert(summaryCounts.get(key) === (counted.get(key) ?? 0), `${connector}: parity count drift for ${key}`);
  const enabled = ledger.filter((row) => row.state === "enabled").length;
  const deletes = ledger.filter((row) => row.method === "DELETE");
  assert(disposition.summary.enabled_operations === enabled, `${connector}: enabled count drift`);
  assert(disposition.summary.disabled_operations === ledger.length - enabled, `${connector}: disabled count drift`);
  assert(disposition.summary.documented_deletes === deletes.length, `${connector}: documented-delete count drift`);
  assert(disposition.summary.enabled_deletes === deletes.filter((row) => row.state === "enabled").length, `${connector}: enabled-delete count drift`);
}

const requestedBatch = process.argv.at(2);
if (!requestedBatch || !batches[requestedBatch]) throw new Error(`usage: ${process.argv[1]} <8|9|10>`);

for (const connector of batches[requestedBatch]) {
  const sourceDir = resolve(defsRoot, connector, "sources");
  const [sourceLock, crosswalk, disposition] = await Promise.all([
    readJSON(resolve(sourceDir, `${connector}-operation-source-lock.json`)),
    readJSON(resolve(sourceDir, `${connector}-operation-crosswalk.json`)),
    readJSON(resolve(sourceDir, `${connector}-declaration-disposition.json`)),
  ]);
  if (sourceLock.state === "skipped") verifySkipped(connector, sourceLock, crosswalk, disposition);
  else verifyMapped(connector, sourceLock, crosswalk, disposition);
  process.stdout.write(`verified ${connector}\n`);
}
