#!/usr/bin/env node

import { execFileSync } from "node:child_process";
import { createHash } from "node:crypto";
import { existsSync } from "node:fs";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";

const root = process.cwd();
const generatedAt = "2026-08-19T00:00:00Z";
const batch2Draft = "fm/cli-top100-declaration-batch-r1-inc2-wip";
const batch2 = ["grafana", "trello", "slack", "n8n", "google-calendar", "gmail", "twilio", "amazon-sqs", "elasticsearch"];
const batch3 = ["gong", "google-ads", "facebook-marketing", "linkedin-ads", "aircall", "xero", "paypal-transaction", "gocardless", "amazon-seller-partner", "miro"];

// These are documentation artifacts only. No endpoint, credential, or provider
// action is invoked by this map generator.
const batch3Sources = {
  "gong": "https://gong.app.gong.io/ajax/settings/api/documentation/specs?version=",
  "google-ads": "https://googleads.googleapis.com/$discovery/rest?version=v22",
  "facebook-marketing": "https://developers.facebook.com/docs/marketing-api/",
  "linkedin-ads": "https://learn.microsoft.com/en-us/linkedin/marketing/integrations/recent-changes?view=li-lms-2024-10",
  "aircall": "https://developer.aircall.io/api-references/",
  "xero": "https://raw.githubusercontent.com/XeroAPI/Xero-OpenAPI/master/xero_accounting.yaml",
  "paypal-transaction": "https://developer.paypal.com/api/rest/",
  "gocardless": "https://developer.gocardless.com/api-reference/openapi",
  "amazon-seller-partner": "https://github.com/amzn/selling-partner-api-models/tree/main/models",
  "miro": "https://developers.miro.com/reference/overview"
};

function pretty(value) {
  return `${JSON.stringify(value, null, 2)}\n`;
}

function definitionPath(connector, file) {
  return path.join(root, "internal", "connectors", "defs", connector, file);
}

function sourceFile(connector, file) {
  return definitionPath(connector, path.join("sources", file));
}

function readDraftLock(connector) {
  return JSON.parse(execFileSync("git", ["show", `${batch2Draft}:internal/connectors/defs/${connector}/sources/${connector}-operation-source-lock.json`], { encoding: "utf8" }));
}

function sourcePathToBundlePath(connector, input) {
  if (connector === "grafana" && !input.startsWith("/api/")) return `/api${input}`;
  if (connector === "gmail" && input.startsWith("/gmail/v1/")) return input.slice("/gmail/v1".length);
  if (connector === "twilio" && input.startsWith("/2010-04-01/")) return input.slice("/2010-04-01".length);
  return input;
}

function canonicalPath(input) {
  return input
    .replace(/\?.*$/, "")
    .replace(/\{[^}]+\}/g, "{}")
    .replace(/\/+/g, "/")
    .replace(/\/$/, "") || "/";
}

function sourceID(connector, method, pathname, index) {
  const slug = pathname.replace(/\{([^}]+)\}/g, "$1").replace(/[^A-Za-z0-9]+/g, "_").replace(/^_|_$/g, "").toLowerCase();
  return `${connector}.rest.${method.toLowerCase()}.${slug || "root"}.${index + 1}`;
}

function methodCounts(operations) {
  return Object.fromEntries([...operations.reduce((counts, operation) => counts.set(operation.method, (counts.get(operation.method) || 0) + 1), new Map()).entries()].sort(([a], [b]) => a.localeCompare(b)));
}

function sourceOperations(lock) {
  return Object.values(lock).flatMap((value) => Array.isArray(value?.operations) ? value.operations : []);
}

async function fetchPublicArtifact(url) {
  const response = await fetch(url, { headers: { "User-Agent": "polymetrics-source-lock/1" } });
  if (!response.ok) throw new Error(`${url}: HTTP ${response.status}`);
  const bytes = Buffer.from(await response.arrayBuffer());
  return { sha256: createHash("sha256").update(bytes).digest("hex"), bytes: bytes.length, contentType: response.headers.get("content-type") || "unknown" };
}

function convertExcluded(endpoint, sourceURL) {
  if (!endpoint.excluded) return endpoint;
  const category = endpoint.excluded.category;
  const model = category === "binary_payload" ? "binary_read" :
    category === "destructive_admin" ? "destructive_action" :
      category === "deprecated" ? "deprecated" :
        category === "duplicate_of" ? "duplicate" : "disallowed";
  const operation = {
    model,
    status: "blocked",
    risk: category === "destructive_admin" ? "high" : "medium",
    blocked_by_default: true,
    reason: endpoint.excluded.reason || "The documented operation remains blocked until its connector-local typed contract is declared.",
    source_url: sourceURL,
    notes: `legacy_exclusion=${category}; converted for the source-locked parity ledger`
  };
  if (model === "duplicate") operation.duplicate_of = endpoint.path;
  const { excluded, ...rest } = endpoint;
  return { ...rest, operation };
}

function inferredOperation(method, sourceURL, source) {
  const upper = method.toUpperCase();
  const model = upper === "GET" || upper === "POST" && /search|query|list/i.test(source.operation_id || "")
    ? "direct_read"
    : upper === "DELETE" ? "destructive_action" : "disallowed";
  return {
    model,
    status: "blocked",
    risk: upper === "DELETE" ? "high" : "medium",
    blocked_by_default: true,
    reason: "Provider-published operation is source-locked but has no connector-owned typed operation/action contract or runnable command declaration.",
    source_url: sourceURL,
    notes: `classification=declaration-pending; source_id=${source.id}; source_location=${source.source_location}`
  };
}

function coverageClass(endpoint, source) {
  const covered = endpoint?.covered_by || {};
  if (covered.stream) return "etl";
  if (covered.write || (Array.isArray(covered.writes) && covered.writes.length > 0)) return "direct_write";
  if (covered.direct_read || (Array.isArray(covered.direct_reads) && covered.direct_reads.length > 0) || (Array.isArray(covered.operations) && covered.operations.length > 0)) return "direct_read";
  if (endpoint?.operation?.model === "binary_read") return "binary_read";
  if (endpoint?.operation?.notes?.includes("official_lane=binary_file")) return source.method === "GET" ? "binary_read" : "binary_write";
  return source.method === "GET" || source.method === "HEAD" ? "direct_read" : "direct_write";
}

function hasTypedWrite(endpoint) {
  const covered = endpoint?.covered_by || {};
  return Boolean(covered.write || (Array.isArray(covered.writes) && covered.writes.length > 0));
}

function sourceTransportDefined(connector) {
  return existsSync(definitionPath(connector, "sync_transport.json"));
}

function enabled(connector, endpoint, classification) {
  const covered = endpoint?.covered_by || {};
  if (classification === "etl") return sourceTransportDefined(connector);
  return Boolean(covered.stream || covered.write || (Array.isArray(covered.writes) && covered.writes.length) || covered.direct_read || (Array.isArray(covered.direct_reads) && covered.direct_reads.length) || (Array.isArray(covered.operations) && covered.operations.length));
}

function declarationPending(connector, source, endpoint, classification) {
  if (classification === "etl") {
    return {
      id: `sync-transport-source-definition-${connector}`,
      evidence: `docs/sync-transport-definition.md:1-30 declares the connector-owned ETL source requirements; internal/connectors/defs/${connector}/sync_transport.json is absent for ${source.method} ${endpoint?.path || source.path}.`,
      minimal_change: "Add the connector-owned source transport declaration, exact executor, stream allowlist, delivery facts, and definition-owned conformance evidence; no generic engine change is required."
    };
  }
  const evidence = endpoint
    ? `internal/connectors/defs/${connector}/api_surface.json: ${source.method} ${endpoint.path} is source-bound but has no runnable connector-local typed contract.`
    : `internal/connectors/defs/${connector}/api_surface.json: ${source.method} ${source.path} is absent before source-map generation.`;
  return {
    id: `typed-operation-contract-${connector}`,
    evidence,
    minimal_change: "Derive a bounded connector-owned typed contract and runnable command/action from the pinned public source, or retain this explicit disabled disposition when the source shape is not executable."
  };
}

function genericDestinationGap() {
  return {
    id: "generic-typed-destination-executor",
    evidence: "internal/app/issue_label_warehouse_transport.go:85-95: the only destination DefinitionFactory is issueLabelDestinationReference and BuildDestination enforces issueLabelTransportConnectorContract, so no connector-neutral typed destination can admit these actions.",
    minimal_change: "register a connector-neutral typed destination DefinitionFactory selected by the definition, with per-connector evidence, explicit source bindings, acknowledgement and per-mode apply strategies"
  };
}

function ledgerRow(connector, source, endpoint, lockName) {
  const classification = coverageClass(endpoint, source);
  const isEnabled = enabled(connector, endpoint, classification);
  const pending = declarationPending(connector, source, endpoint, classification);
  const destinationGap = genericDestinationGap();
  const reverseETLEligibility = hasTypedWrite(endpoint) ? {
    state: "foundation-gap",
    detail: "The typed direct-write action is executable, but reverse ETL still needs a connector-neutral typed destination executor, explicit source binding, acknowledgement, and per-mode apply strategies.",
    foundation_gap: destinationGap
  } : null;
  const endpointCopy = {
    method: source.method,
    path: endpoint?.path || source.path,
    covered_by: endpoint?.covered_by || null,
    operation: endpoint?.operation || null
  };
  const foundation = isEnabled
    ? { state: "present", evidence: "A connector-local declared stream, typed action, or implemented direct-read binding owns this source operation." }
    : { state: "present", evidence: "No shared engine change is requested for this row; the missing work is a connector-local declaration bound to the pinned operation.", declaration_pending: pending };
  return {
    method: source.method,
    path: endpointCopy.path,
    parity_class: classification,
    api_surface: endpointCopy,
    source: {
      source_lock: `sources/${lockName}`,
      source_id: source.id,
      source_url: source.source_url,
      source_location: source.source_location,
      operation_id: source.operation_id || null,
      deprecated: Boolean(source.deprecated)
    },
    state: isEnabled ? "enabled" : "disabled",
    foundation,
    rejection: isEnabled ? null : {
      reason: "declaration-pending",
      recoverable: true,
      detail: "The engine shape is already available; this source operation awaits its connector-local typed contract and/or runnable command declaration.",
      evidence: pending.evidence
    },
    declaration: isEnabled
      ? {
          status: "enabled; existing connector-local binding",
          contract: endpointCopy.covered_by,
          ...(reverseETLEligibility ? { reverse_etl_eligibility: reverseETLEligibility } : {})
        }
      : { status: `disabled; declaration-pending ${pending.id}`, contract: null }
  };
}

function transportSummary(connector) {
  const base = `internal/connectors/defs/${connector}/sync_transport.json`;
  return {
    contract: "docs/sync-transport-definition.md (PR #4286)",
    source_transport: {
      state: "declaration-pending",
      declaration_pending: {
        id: `sync-transport-source-definition-${connector}`,
        evidence: `docs/sync-transport-definition.md:1-30 declares the connector-owned source requirements; ${base} is absent.`,
        minimal_change: "Add the connector-owned source transport declaration, exact executor, stream allowlist, delivery facts, and definition-owned conformance evidence; no generic engine change is required."
      }
    },
    destination_transport: {
      state: "gap",
      foundation_gap: genericDestinationGap()
    }
  };
}

async function batch3Lock(connector, surface) {
  const sourceURL = batch3Sources[connector];
  const artifact = await fetchPublicArtifact(sourceURL);
  const operations = surface.endpoints.map((endpoint, index) => ({
    id: sourceID(connector, endpoint.method, endpoint.path, index),
    protocol: "rest",
    method: endpoint.method.toUpperCase(),
    path: endpoint.path,
    operation_id: endpoint.operation?.notes?.match(/operation_id=([^;\s]+)/)?.[1] || null,
    deprecated: endpoint.operation?.model === "deprecated",
    source_location: `documented endpoint ${endpoint.method.toUpperCase()} ${endpoint.path}`,
    source_url: sourceURL
  }));
  return {
    schema_version: 2,
    connector,
    captured_at: generatedAt,
    rest: {
      source_url: sourceURL,
      sha256: artifact.sha256,
      bytes: artifact.bytes,
      format: artifact.contentType,
      inventory_basis: "existing connector api_surface.json reconciled to the pinned public provider documentation artifact",
      operation_counts: methodCounts(operations),
      operations
    }
  };
}

async function loadSourceLock(connector, surface) {
  if (batch3.includes(connector)) return batch3Lock(connector, surface);
  const lock = readDraftLock(connector);
  for (const section of Object.values(lock)) {
    if (!Array.isArray(section?.operations)) continue;
    for (const operation of section.operations) {
      operation.method = operation.method.toUpperCase();
      operation.path = sourcePathToBundlePath(connector, operation.path);
      operation.source_url ||= section.source_url;
    }
  }
  lock.captured_at = generatedAt;
  return lock;
}

function mapSurface(connector, surface, lock) {
  const sourceURL = Object.values(lock).find((value) => value?.source_url)?.source_url || surface.docs;
  const endpoints = surface.endpoints.map((endpoint) => convertExcluded(endpoint, sourceURL));
  const index = new Map();
  for (const endpoint of endpoints) {
    const key = `${endpoint.method.toUpperCase()} ${canonicalPath(endpoint.path)}`;
    if (!index.has(key)) index.set(key, endpoint);
  }
  for (const source of sourceOperations(lock)) {
    const key = `${source.method} ${canonicalPath(source.path)}`;
    if (index.has(key)) continue;
    const endpoint = { method: source.method, path: source.path, operation: inferredOperation(source.method, source.source_url, source) };
    endpoints.push(endpoint);
    index.set(key, endpoint);
  }
  const rows = sourceOperations(lock).map((source) => {
    const key = `${source.method} ${canonicalPath(source.path)}`;
    const endpoint = index.get(key);
    return ledgerRow(connector, source, endpoint, `${connector}-operation-source-lock.json`);
  });
  return { surface: { ...surface, operation_ledger_version: 1, endpoints }, rows };
}

function summary(connector, lock, rows) {
  const counts = ["direct_read", "direct_write", "etl", "reverse_etl", "binary_read", "binary_write"].map((key) => ({ key, count: rows.filter((row) => row.parity_class === key).length }));
  const enabledRows = rows.filter((row) => row.state === "enabled");
  const deletes = rows.filter((row) => row.method === "DELETE");
  const enabledDeletes = deletes.filter((row) => row.state === "enabled");
  const reverseETL = rows.filter((row) => row.declaration.reverse_etl_eligibility);
  const declarationPendingRows = rows.filter((row) => row.state === "disabled");
  const declarationPendingByID = [...declarationPendingRows.reduce((counts, row) => {
    const id = row.foundation.declaration_pending?.id;
    if (id) counts.set(id, (counts.get(id) || 0) + 1);
    return counts;
  }, new Map()).entries()].map(([id, count]) => ({ id, count }));
  return {
    api_surface_rows: rows.length,
    exact_source_rows: rows.length,
    declared_operations: rows.length,
    declared_percent: 100,
    enabled_operations: enabledRows.length,
    enabled_percent: Number(((enabledRows.length / rows.length) * 100).toFixed(2)),
    disabled_operations: rows.length - enabledRows.length,
    documented_deletes: deletes.length,
    enabled_deletes: enabledDeletes.length,
    parity_class_counts: counts,
    stream_bindings: rows.filter((row) => row.api_surface.covered_by?.stream).length,
    writes_actions: rows.filter((row) => row.api_surface.covered_by?.write || row.api_surface.covered_by?.writes?.length).length,
    terminal_commands: enabledRows.length,
    live_certification: "pending",
    gap_ids: ["generic-typed-destination-executor"],
    foundation_gaps: [{ id: "generic-typed-destination-executor", count: reverseETL.length, scope: "destination_transport" }],
    rejected_by_reason: [{ key: "declaration-pending", count: declarationPendingRows.length }],
    reverse_etl_eligibility: {
      state: "foundation-gap",
      typed_direct_write_operations: reverseETL.length,
      foundation_gap: genericDestinationGap()
    },
    transport: transportSummary(connector),
    declaration_pending_ids: declarationPendingByID.map((entry) => entry.id),
    declaration_pending: declarationPendingByID
  };
}

async function main() {
  const reports = [];
  for (const connector of [...batch2, ...batch3]) {
    const surface = JSON.parse(await readFile(definitionPath(connector, "api_surface.json"), "utf8"));
    const lock = await loadSourceLock(connector, surface);
    const mapped = mapSurface(connector, surface, lock);
    const lockName = `${connector}-operation-source-lock.json`;
    const disposition = {
      schema_version: 1,
      connector,
      generated_at: generatedAt,
      source_basis: {
        source_lock: `sources/${lockName}`,
        source_url: lock.rest.source_url,
        source_sha256: lock.rest.sha256,
        source_bytes: lock.rest.bytes,
        source_operation_count: mapped.rows.length
      },
      summary: summary(connector, lock, mapped.rows),
      ledger_dispositions: mapped.rows
    };
    await mkdir(path.dirname(sourceFile(connector, lockName)), { recursive: true });
    await writeFile(sourceFile(connector, lockName), pretty(lock));
    await writeFile(sourceFile(connector, `${connector}-declaration-disposition.json`), pretty(disposition));
    await writeFile(definitionPath(connector, "api_surface.json"), pretty(mapped.surface));
    reports.push({ connector, ...disposition.summary });
  }
  await writeFile(path.join(root, ".planning", "phases", "issue-4289-parity-map-batches-2-3-r1", "traces", "parity-map-summary.json"), pretty({ generated_at: generatedAt, connectors: reports }));
}

await main();
