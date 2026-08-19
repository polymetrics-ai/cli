#!/usr/bin/env node

// Cross-artifact integrity assertions for the #4292 source-first parity map.
// These fail if a source lock becomes a landing page, omits a total, or lets
// api_surface.json turn back into the operation-inventory boundary.

import { createHash } from "node:crypto";
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
const mappedConfidence = new Set(["machine-readable", "complete-rendered-reference"]);
const reverseETLGapID = "generic-typed-destination-executor";
const reverseETLMinimalChange = "register a connector-neutral typed destination DefinitionFactory selected by the definition, with per-connector evidence, explicit source bindings, acknowledgement and per-mode apply strategies";

function assert(condition, message) {
  if (!condition) throw new Error(message);
}

async function readJSON(path) {
  return JSON.parse(await readFile(path, "utf8"));
}

function ids(items, selector, label, connector) {
  const values = items.map(selector);
  assert(new Set(values).size === values.length, `${connector}: duplicated ${label}`);
  return values;
}

function sameIDs(left, right, label, connector) {
  assert(left.length === right.length, `${connector}: ${label} count differs (${left.length} != ${right.length})`);
  for (const value of left) assert(right.includes(value), `${connector}: ${label} missing ${value}`);
}

function validDocument(document, connector) {
  assert(typeof document.source_url === "string" && document.source_url.startsWith("https://"), `${connector}: document source URL missing`);
  assert(/^[a-f0-9]{64}$/.test(document.sha256 ?? ""), `${connector}: document SHA-256 missing`);
  assert(Number.isInteger(document.bytes) && document.bytes > 0, `${connector}: document byte count missing`);
  assert(["http", "browser"].includes(document.retrieval_method), `${connector}: document retrieval method missing`);
  assert(["machine-readable-spec", "complete-rendered-reference"].includes(document.representation), `${connector}: source document is not a complete provider specification/reference`);
}

function sourceBundle(documents) {
  const normalized = documents
    .map(({ source_url, sha256, bytes, retrieval_method, representation }) => ({ source_url, sha256, bytes, retrieval_method, representation }))
    .sort((left, right) => left.source_url.localeCompare(right.source_url));
  return {
    document_count: normalized.length,
    sha256: createHash("sha256").update(JSON.stringify(normalized)).digest("hex"),
    bytes: normalized.reduce((total, document) => total + document.bytes, 0),
  };
}

function endpointKey(method, path) {
  return `${String(method).toUpperCase()} ${String(path).replace(/\{[^}]+\}/g, "{}").replace(/\/+$/, "") || "/"}`;
}

function verifyUnavailable(connector, sourceLock, crosswalk, disposition) {
  assert(sourceLock.state === "skipped", `${connector}: unavailable source must be skipped`);
  assert(sourceLock.skip?.reason === "no-public-api-description", `${connector}: unavailable source reason drift`);
  assert(sourceLock.skip?.retrieval_method === "browser", `${connector}: unavailable source must record browser retrieval`);
  assert(sourceLock.rest?.counts?.total === null, `${connector}: unavailable source total must be explicit null`);
  assert(Array.isArray(sourceLock.rest?.operations) && sourceLock.rest.operations.length === 0, `${connector}: unavailable source must not fabricate operations`);
  assert(crosswalk.state === "skipped" && crosswalk.source_operations?.length === 0, `${connector}: unavailable crosswalk drift`);
  assert(disposition.summary?.state === "skipped" && disposition.summary.operations_found === null, `${connector}: unavailable disposition drift`);
  assert(disposition.ledger_dispositions?.length === 0, `${connector}: unavailable source must not fabricate ledger rows`);
}

function verifyDynamic(connector, sourceLock, crosswalk, disposition) {
  assert(sourceLock.state === "dynamic", `${connector}: dynamic source state missing`);
  assert(sourceLock.dynamic?.reason === "dynamic-instance-dependent", `${connector}: dynamic reason drift`);
  assert(sourceLock.rest?.counts?.total === null, `${connector}: dynamic source must not fabricate a total`);
  assert(Array.isArray(sourceLock.rest?.documents) && sourceLock.rest.documents.length > 0, `${connector}: dynamic source needs its published reference pin`);
  sourceLock.rest.documents.forEach((document) => validDocument(document, connector));
  assert(sourceLock.rest.coverage_confidence?.level === "dynamic-instance-dependent", `${connector}: dynamic confidence drift`);
  assert(crosswalk.state === "dynamic" && crosswalk.source_operations?.length === 0, `${connector}: dynamic crosswalk drift`);
  assert(disposition.summary?.state === "dynamic" && disposition.summary.operations_found === null, `${connector}: dynamic disposition drift`);
  assert(disposition.ledger_dispositions?.length === 0, `${connector}: dynamic source must not fabricate ledger rows`);
}

function verifyMapped(connector, sourceLock, crosswalk, disposition, surface) {
  const source = sourceLock.rest;
  assert(sourceLock.state == null, `${connector}: mapped source has an unexpected state`);
  assert(typeof source?.source_url === "string" && source.source_url.startsWith("https://"), `${connector}: source URL missing`);
  assert(["machine-readable-spec", "complete-rendered-reference"].includes(source.representation), `${connector}: source must be a complete spec or rendered reference`);
  assert(Array.isArray(source.documents) && source.documents.length > 0, `${connector}: exact source-document pins missing`);
  source.documents.forEach((document) => validDocument(document, connector));
  assert(JSON.stringify(source.source_bundle) === JSON.stringify(sourceBundle(source.documents)), `${connector}: source bundle digest/bytes drift`);
  assert(Number.isInteger(source.counts?.total) && source.counts.total > 0, `${connector}: counts.total missing`);
  assert(typeof source.counts?.by_method === "object" && Object.keys(source.counts.by_method).length > 0, `${connector}: per-method counts missing`);
  assert(source.counts?.by_kind?.rest === source.counts.total, `${connector}: REST-kind count drift`);
  assert(mappedConfidence.has(source.coverage_confidence?.level), `${connector}: coverage confidence missing`);
  assert(typeof source.coverage_confidence?.basis === "string" && source.coverage_confidence.basis.length > 0, `${connector}: coverage basis missing`);

  const sourceOperations = source.operations ?? [];
  const lockIDs = ids(sourceOperations, (operation) => operation.id, "source operation ID", connector);
  const crosswalkIDs = ids(crosswalk.source_operations ?? [], (operation) => operation.source_id, "crosswalk source ID", connector);
  const ledger = disposition.ledger_dispositions ?? [];
  const ledgerIDs = ids(ledger, (row) => row.source?.source_id, "ledger source ID", connector);
  sameIDs(lockIDs, crosswalkIDs, "source lock / crosswalk IDs", connector);
  sameIDs(lockIDs, ledgerIDs, "source lock / ledger IDs", connector);
  assert(source.counts.total === lockIDs.length, `${connector}: total does not match extracted operations`);
  const countedByMethod = sourceOperations.reduce((counts, operation) => {
    counts[operation.method] = (counts[operation.method] ?? 0) + 1;
    return counts;
  }, {});
  const sortedCountedByMethod = Object.fromEntries(Object.entries(countedByMethod).sort(([left], [right]) => left.localeCompare(right)));
  assert(JSON.stringify(source.counts.by_method) === JSON.stringify(sortedCountedByMethod), `${connector}: per-method count drift`);

  assert(disposition.source_basis?.source_operation_count === lockIDs.length, `${connector}: disposition source count drift`);
  assert(JSON.stringify(disposition.source_basis?.source_bundle) === JSON.stringify(source.source_bundle), `${connector}: disposition bundle drift`);
  assert(disposition.summary?.operations_found === lockIDs.length, `${connector}: operations-found drift`);
  assert(disposition.summary?.api_surface_rows === lockIDs.length, `${connector}: source API-surface row count drift`);
  assert(Number.isInteger(disposition.summary?.old_api_surface_count), `${connector}: prior api-surface count missing`);
  assert(!Object.hasOwn(disposition.summary ?? {}, "declared_percent"), `${connector}: declared_percent is forbidden`);
  assert(JSON.stringify(disposition.summary?.coverage_confidence) === JSON.stringify(source.coverage_confidence), `${connector}: confidence drift`);

  const surfaceKeys = new Set((surface.endpoints ?? []).map((endpoint) => endpointKey(endpoint.method, endpoint.path)));
  assert(surface.operation_ledger_version === 1, `${connector}: regenerated surface must use the blocked operation ledger`);
  for (const operation of sourceOperations) {
    assert(surfaceKeys.has(endpointKey(operation.method, operation.path)), `${connector}: source operation missing from regenerated api_surface`);
  }

  const sourceByID = new Map(sourceOperations.map((operation) => [operation.id, operation]));
  const parityCounts = new Map();
  for (const row of ledger) {
    assert(classKeys.has(row.parity_class), `${connector}: unknown parity class ${row.parity_class}`);
    assert(row.parity_class !== "reverse_etl", `${connector}: reverse_etl cannot be a primary endpoint parity class`);
    const operation = sourceByID.get(row.source?.source_id);
    assert(operation, `${connector}: ledger row lacks source operation`);
    assert(row.method === operation.method && row.path === operation.path, `${connector}: ledger route drift for ${row.source.source_id}`);
    assert(row.source?.source_url === operation.source_url && row.source?.source_location === operation.source_location, `${connector}: ledger source provenance drift`);
    assert(row.source?.operation_id === operation.operation_id, `${connector}: ledger operation ID drift`);
    if (row.parity_class === "direct_write") {
      const eligibility = row.reverse_etl_eligibility;
      assert(eligibility?.state === "foundation-gap", `${connector}: direct write lacks reverse-ETL eligibility status`);
      assert(eligibility.foundation_gap?.id === reverseETLGapID, `${connector}: direct write has wrong reverse-ETL gap`);
      assert(eligibility.foundation_gap?.evidence?.startsWith("internal/app/issue_label_warehouse_transport.go:85-95"), `${connector}: direct write has wrong reverse-ETL evidence`);
      assert(eligibility.foundation_gap?.minimal_change === reverseETLMinimalChange, `${connector}: direct write has wrong reverse-ETL minimal change`);
      if (row.declaration?.writes?.length > 0) assert(row.state === "enabled", `${connector}: typed direct write must be enabled`);
    } else {
      assert(!Object.hasOwn(row, "reverse_etl_eligibility"), `${connector}: only direct_write may carry reverse-ETL eligibility`);
    }
    parityCounts.set(row.parity_class, (parityCounts.get(row.parity_class) ?? 0) + 1);
  }
  const summaryCounts = new Map((disposition.summary?.parity_class_counts ?? []).map(({ key, count }) => [key, count]));
  for (const key of classKeys) assert(summaryCounts.get(key) === (parityCounts.get(key) ?? 0), `${connector}: parity count drift for ${key}`);
  const enabled = ledger.filter((row) => row.state === "enabled");
  assert(disposition.summary.enabled_operations === enabled.length, `${connector}: enabled count drift`);
  assert(disposition.summary.disabled_operations === ledger.length - enabled.length, `${connector}: disabled count drift`);
  assert(!JSON.stringify(disposition).includes("transport_binding"), `${connector}: invented transport binding is forbidden`);
}

const requestedBatch = process.argv.at(2);
if (!requestedBatch || !batches[requestedBatch]) throw new Error(`usage: ${process.argv[1]} <8|9|10>`);
for (const connector of batches[requestedBatch]) {
  const sourceDir = resolve(defsRoot, connector, "sources");
  const [sourceLock, crosswalk, disposition, surface] = await Promise.all([
    readJSON(resolve(sourceDir, `${connector}-operation-source-lock.json`)),
    readJSON(resolve(sourceDir, `${connector}-operation-crosswalk.json`)),
    readJSON(resolve(sourceDir, `${connector}-declaration-disposition.json`)),
    readJSON(resolve(defsRoot, connector, "api_surface.json")),
  ]);
  if (sourceLock.state === "skipped") verifyUnavailable(connector, sourceLock, crosswalk, disposition);
  else if (sourceLock.state === "dynamic") verifyDynamic(connector, sourceLock, crosswalk, disposition);
  else verifyMapped(connector, sourceLock, crosswalk, disposition, surface);
  process.stdout.write(`verified ${connector}\n`);
}
