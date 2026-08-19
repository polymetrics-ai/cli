#!/usr/bin/env node

// Source-first parity-map generator for issue #4292. The provider operation
// inventory comes exclusively from extract-source-operations.go. In
// particular, an existing api_surface.json is a coverage crosswalk, never the
// boundary used to decide which provider operations exist.

import { createHash } from "node:crypto";
import { execFile as execFileCallback } from "node:child_process";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { promisify } from "node:util";
import { resolve } from "node:path";

const execFile = promisify(execFileCallback);
const root = resolve(import.meta.dirname, "../../../..");
const defsRoot = resolve(root, "internal/connectors/defs");
const extractor = resolve(import.meta.dirname, "extract-source-operations.go");
const generatedAt = new Date().toISOString();
const reviewedAt = generatedAt.slice(0, 10);

const batches = {
  "8": ["brex", "zoho-books", "testrail", "amplitude", "posthog", "metabase", "dbt", "looker", "mode", "dremio"],
  "9": ["coda", "clickup-api", "calendly", "greenhouse", "lever-hiring", "ashby", "workable", "recruitee", "hibob", "factorial"],
  "10": ["datadog", "pagerduty", "auth0", "okta", "firehydrant", "adobe-commerce-magento", "commercetools", "recharge", "docuseal", "eventbrite"],
};

const allConnectors = Object.values(batches).flat();

const reverseETLEligibility = {
  state: "foundation-gap",
  foundation_gap: {
    id: "generic-typed-destination-executor",
    evidence: "internal/app/issue_label_warehouse_transport.go:85-95 (acb85dc03): issueLabelDestinationReference is the only destination factory and BuildDestination enforces the GitHub issue-label contract.",
    minimal_change: "register a connector-neutral typed destination DefinitionFactory selected by the definition, with per-connector evidence, explicit source bindings, acknowledgement and per-mode apply strategies",
  },
};

const unavailableSources = {
  testrail: {
    attempted_url: "https://support.testrail.com/hc/en-us/categories/7076541806228",
    reason: "no-public-api-description",
    retrieval_method: "browser",
    detail: "The official rendered reference was requested in Chrome without credentials and returned a Cloudflare security-verification page rather than the published API reference. Per issue #4292 decision, no source pin or operation ledger is fabricated.",
  },
  eventbrite: {
    attempted_url: "https://www.eventbrite.com/platform/api",
    reason: "no-public-api-description",
    retrieval_method: "browser",
    detail: "The official rendered reference was requested in Chrome without credentials and returned the provider response {\"message\":\"Unauthorized\"} rather than a published API reference. Per issue #4292 source-lock rule, no source pin or operation ledger is fabricated.",
  },
  greenhouse: {
    attempted_url: "https://developers.greenhouse.io/harvest.html",
    reason: "no-public-api-description",
    retrieval_method: "browser",
    detail: "The provider's formerly public Harvest API reference was requested in Chrome without credentials and rendered the Greenhouse Recruiting sign-in page. No current public provider API description could be pinned, so this map does not reuse the old api_surface.json as a substitute source inventory.",
  },
};

const dynamicSources = {
  "adobe-commerce-magento": {
    source_url: "https://developer.adobe.com/commerce/webapi/rest/",
    retrieval_method: "http",
    reason: "dynamic-instance-dependent",
    detail: "Adobe Commerce REST routes and schemas are generated from the modules enabled on each configured Magento instance. The official reference documents that instance-specific contract; there is no provider-global static operation total to fabricate.",
  },
};

// Native hooks retain a real, provider-documented source operation where the
// native connector routes the stream itself. Their legacy HOOK rows remain in
// api_surface solely to satisfy the current native runtime coverage contract;
// they are never counted as source operations or disposition rows.
const nativeHookBindings = new Set(["metabase", "mode"]);

function pretty(value) {
  return `${JSON.stringify(value, null, 2)}\n`;
}

function hash(value) {
  return createHash("sha256").update(value).digest("hex");
}

function endpointKey(method, path) {
  return `${String(method).toUpperCase()} ${canonicalPath(path)}`;
}

function canonicalPath(path) {
  return String(path)
    .replace(/\{[^}]+\}/g, "{}")
    .replace(/\/+$/, "") || "/";
}

function isHTTPMethod(method) {
  return ["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"].includes(String(method).toUpperCase());
}

function operationSlug(method, path) {
  return `${method.toLowerCase()}-${path}`
    .replaceAll("{", "")
    .replaceAll("}", "")
    .replaceAll(/[^a-zA-Z0-9]+/g, "-")
    .replaceAll(/^-+|-+$/g, "")
    .toLowerCase();
}

function sourceBundle(documents) {
  const normalized = documents
    .map(({ source_url, sha256, bytes, retrieval_method, representation }) => ({ source_url, sha256, bytes, retrieval_method, representation }))
    .sort((left, right) => left.source_url.localeCompare(right.source_url));
  return {
    document_count: normalized.length,
    sha256: hash(JSON.stringify(normalized)),
    bytes: normalized.reduce((total, document) => total + document.bytes, 0),
  };
}

function countByMethod(operations) {
  const counts = {};
  for (const operation of operations) counts[operation.method] = (counts[operation.method] ?? 0) + 1;
  return Object.fromEntries(Object.entries(counts).sort(([left], [right]) => left.localeCompare(right)));
}

function classCounts(rows) {
  const counts = new Map();
  for (const row of rows) counts.set(row.parity_class, (counts.get(row.parity_class) ?? 0) + 1);
  return ["direct_read", "direct_write", "etl", "reverse_etl", "binary_read", "binary_write"]
    .map((key) => ({ key, count: counts.get(key) ?? 0 }));
}

async function readJSON(path) {
  return JSON.parse(await readFile(path, "utf8"));
}

async function readBaselineJSON(relativePath) {
  const { stdout } = await execFile("git", ["show", `HEAD:${relativePath}`], { cwd: root, maxBuffer: 32 * 1024 * 1024 });
  return JSON.parse(stdout);
}

async function extract(connector) {
  const { stdout } = await execFile("go", ["run", extractor, "--connector", connector], {
    cwd: root,
    maxBuffer: 128 * 1024 * 1024,
  });
  return JSON.parse(stdout);
}

async function fetchPublishedDocument(url, retrievalMethod) {
  const response = await fetch(url, {
    redirect: "follow",
    headers: { "user-agent": "Polymetrics-source-lock/2.0 (+https://github.com/polymetrics-ai/cli)" },
    signal: AbortSignal.timeout(60_000),
  });
  if (!response.ok) throw new Error(`${url}: HTTP ${response.status}`);
  const bytes = Buffer.from(await response.arrayBuffer());
  return {
    source_url: response.url,
    sha256: hash(bytes),
    bytes: bytes.length,
    content_type: response.headers.get("content-type"),
    retrieval_method: retrievalMethod,
    representation: "complete-rendered-reference",
  };
}

function mergeCoverage(entries, connector, key) {
  const covered = entries.map((entry) => entry.covered_by).filter(Boolean);
  if (covered.length === 0) return null;
  const streamNames = [...new Set(covered.map((item) => item.stream).filter(Boolean))];
  const writes = [...new Set(covered.flatMap((item) => [item.write, ...(item.writes ?? [])]).filter(Boolean))];
  const directReads = [...new Set(covered.flatMap((item) => [item.direct_read, ...(item.direct_reads ?? [])]).filter(Boolean))];
  const operations = [...new Set(covered.flatMap((item) => item.operations ?? []))];
  const merged = {};
  // api_surface's executable surface has a singular stream slot. Additional
  // existing streams sharing this same provider operation remain explicit
  // duplicate coverage rows in buildSurface, while the source disposition
  // records the operation once (as the provider publishes it once).
  if (streamNames.length === 1) merged.stream = streamNames[0];
  if (streamNames.length > 1) merged.stream = streamNames[0];
  if (writes.length === 1) merged.write = writes[0];
  if (writes.length > 1) merged.writes = writes;
  if (directReads.length === 1) merged.direct_read = directReads[0];
  if (directReads.length > 1) merged.direct_reads = directReads;
  if (operations.length > 0) merged.operations = operations;
  return Object.keys(merged).length === 0 ? null : merged;
}

function coverageWriteNames(coverage) {
  if (!coverage) return [];
  return [coverage.write, ...(coverage.writes ?? [])].filter(Boolean);
}

function coverageDirectReadNames(coverage) {
  if (!coverage) return [];
  return [coverage.direct_read, ...(coverage.direct_reads ?? [])].filter(Boolean);
}

function classification(endpoint, coverage, actions) {
  const writes = coverageWriteNames(coverage);
  const directReads = coverageDirectReadNames(coverage);
  if (coverage?.stream) return { key: "etl", state: "enabled", binding: { stream: coverage.stream } };
  if (writes.length > 0 && writes.every((write) => actions.has(write))) return { key: "direct_write", state: "enabled", binding: { writes } };
  if (directReads.length > 0) return { key: "direct_read", state: "enabled", binding: { directReads } };
  if (endpoint.operation?.model === "binary_read") return { key: "binary_read", state: "disabled", binding: null };
  if (/^(GET|HEAD)$/i.test(endpoint.method)) return { key: "direct_read", state: "disabled", binding: null };
  return { key: "direct_write", state: "disabled", binding: null };
}

function blockedOperation(sourceOperation) {
  const isRead = /^(GET|HEAD)$/i.test(sourceOperation.method);
  return {
    model: isRead ? "direct_read" : "local_workflow",
    status: "blocked",
    risk: sourceOperation.method === "DELETE" ? "high" : "medium",
    blocked_by_default: true,
    reason: isRead
      ? "Provider-published read operation has no connector-owned typed stream or bounded direct-read contract; it remains blocked until that declaration, schema, fixtures, and command or stream binding are added."
      : "Provider-published state-changing operation has no connector-owned typed direct-write action; it remains blocked until a bounded declaration, schema, fixtures, and command binding are added.",
    source_url: sourceOperation.source_url,
    notes: "Named dependency: connector-local typed operation declaration (schema, fixtures, and CLI binding) derived from this pinned provider operation.",
  };
}

function dispositionRow(connector, endpoint, index, sourceOperation, classInfo) {
  const sourceID = sourceOperation.id;
  const source = {
    source_lock: `sources/${connector}-operation-source-lock.json`,
    source_id: sourceID,
    source_url: sourceOperation.source_url,
    source_location: sourceOperation.source_location,
    operation_id: sourceOperation.operation_id,
    deprecated: sourceOperation.deprecated,
  };
  const writes = classInfo.binding?.writes ?? [];
  const directWrite = classInfo.key === "direct_write";
  const foundation = classInfo.state === "enabled"
    ? {
        state: "present",
        evidence: classInfo.binding.stream
          ? `internal/connectors/defs/${connector}/streams.json: existing stream ${JSON.stringify(classInfo.binding.stream)} binds this provider operation.`
          : classInfo.binding.directReads
            ? `internal/connectors/defs/${connector}/cli_surface.json: implemented direct-read command(s) ${JSON.stringify(classInfo.binding.directReads)} bind this provider operation.`
            : `internal/connectors/defs/${connector}/writes.json: typed write action(s) ${JSON.stringify(writes)} bind this provider operation.`,
        ...(classInfo.binding.stream || classInfo.binding.directReads ? {} : { contract: { kind: "typed_write_action", ids: writes } }),
      }
    : {
        state: "present",
        evidence: "No shared engine change is requested for this row; the missing work is a connector-local declaration bound to the pinned provider operation.",
        declaration_pending: {
          id: `typed-operation-contract-${connector}`,
          evidence: `sources/${connector}-operation-source-lock.json:${sourceOperation.source_location}.`,
          minimal_change: "Add a bounded connector-owned operation contract and runnable command from the pinned provider document, or retain a disabled source disposition when the source shape is not executable.",
        },
      };
  const rejection = classInfo.state === "enabled" ? null : {
    reason: "declaration-pending",
    recoverable: true,
    detail: "The engine shape is already available; this provider operation awaits its connector-local typed contract and/or runnable command declaration.",
    evidence: `sources/${connector}-operation-source-lock.json:${sourceOperation.source_location}.`,
  };
  return {
    method: endpoint.method,
    path: endpoint.path,
    parity_class: classInfo.key,
    api_surface: { method: endpoint.method, path: endpoint.path, ...(endpoint.covered_by ? { covered_by: endpoint.covered_by } : { operation: endpoint.operation }) },
    source,
    state: classInfo.state,
    foundation,
    ...(directWrite ? { reverse_etl_eligibility: reverseETLEligibility } : {}),
    rejection,
    declaration: classInfo.state === "enabled"
      ? (classInfo.binding.stream
        ? { status: `enabled; existing etl stream ${JSON.stringify(classInfo.binding.stream)} binds the pinned provider operation`, stream: classInfo.binding.stream }
        : classInfo.binding.directReads
          ? { status: `enabled; implemented direct_read command(s) ${JSON.stringify(classInfo.binding.directReads)} bind the pinned provider operation`, direct_reads: classInfo.binding.directReads }
          : { status: `enabled; typed direct_write action(s) ${JSON.stringify(writes)} bind the pinned provider operation`, writes })
      : { status: `disabled; declaration-pending typed-operation-contract-${connector}`, contract: null },
    source_index: index,
  };
}

function stripSourceIndex(row) {
  const { source_index, ...clean } = row;
  return clean;
}

function sourceRows(sourceOperations, sourceEndpoints, connector, actions) {
  return sourceOperations.map((sourceOperation, index) => {
    const endpoint = sourceEndpoints[index];
    return dispositionRow(connector, endpoint, index, sourceOperation, classification(endpoint, endpoint.covered_by, actions));
  });
}

function buildSurface(connector, oldSurface, sourceOperations) {
  const sourceKeys = new Set(sourceOperations.map((operation) => endpointKey(operation.method, operation.path)));
  const oldByKey = new Map();
  for (const endpoint of oldSurface.endpoints ?? []) {
    if (!isHTTPMethod(endpoint.method)) continue;
    const key = endpointKey(endpoint.method, endpoint.path);
    const entries = oldByKey.get(key) ?? [];
    entries.push(endpoint);
    oldByKey.set(key, entries);
  }

  const sourceEndpoints = sourceOperations.map((sourceOperation) => {
    const key = endpointKey(sourceOperation.method, sourceOperation.path);
    const oldEntries = oldByKey.get(key) ?? [];
    const coverage = mergeCoverage(oldEntries, connector, key);
    return {
      method: sourceOperation.method,
      path: sourceOperation.path,
      ...(coverage ? { covered_by: coverage } : { operation: blockedOperation(sourceOperation) }),
    };
  });

  const unmatched = [];
  for (const endpoint of oldSurface.endpoints ?? []) {
    if (!endpoint.covered_by || !isHTTPMethod(endpoint.method)) continue;
    const key = endpointKey(endpoint.method, endpoint.path);
    if (!sourceKeys.has(key)) unmatched.push(`${endpoint.method} ${endpoint.path} -> ${JSON.stringify(endpoint.covered_by)}`);
  }
  // A legacy runtime binding that no longer appears in the provider's current
  // published inventory is not a provider operation and never enters the
  // source map. It remains in api_surface only because conformance must keep
  // the already-shipped stream/action bound while a separate behavior change
  // updates or retires it. Keeping it is deliberately not a source claim.
  const retainedLegacyBindings = (oldSurface.endpoints ?? [])
    .filter((endpoint) => endpoint.covered_by && isHTTPMethod(endpoint.method) && !sourceKeys.has(endpointKey(endpoint.method, endpoint.path)));
  const retainedSharedBindings = [];
  for (const [key, entries] of oldByKey) {
    if (!sourceKeys.has(key)) continue;
    const streams = [...new Set(entries.map((entry) => entry.covered_by?.stream).filter(Boolean))];
    if (streams.length < 2) continue;
    for (const stream of streams.slice(1)) {
      const original = entries.find((entry) => entry.covered_by?.stream === stream);
      retainedSharedBindings.push({ method: original.method, path: original.path, covered_by: { stream } });
    }
  }
  const retainedHooks = nativeHookBindings.has(connector)
    ? (oldSurface.endpoints ?? []).filter((endpoint) => !isHTTPMethod(endpoint.method) && endpoint.covered_by)
    : [];
  const endpoints = [...sourceEndpoints, ...retainedSharedBindings, ...retainedLegacyBindings, ...retainedHooks]
    .sort((left, right) => `${left.method} ${left.path}`.localeCompare(`${right.method} ${right.path}`));
  return {
    surface: {
      ...oldSurface,
      reviewed_at: reviewedAt,
      operation_ledger_version: 1,
      scope: `Source-surface refresh ${reviewedAt}: ${sourceOperations.length} provider-documented REST method+path operations are derived from the pinned source-lock inventory, not from the prior api_surface.json. Every provider operation is either bound to an existing executable declaration or recorded blocked by default. ${retainedSharedBindings.length > 0 ? `${retainedSharedBindings.length} duplicate coverage row(s) preserve additional streams bound to the same provider operation and are excluded from the provider operation total.` : ""} ${retainedLegacyBindings.length > 0 ? `${retainedLegacyBindings.length} legacy runtime-binding row(s) are retained only to preserve current conformance and are excluded from the provider operation total.` : ""} ${retainedHooks.length > 0 ? `${retainedHooks.length} native HOOK coverage row(s) remain runtime-binding metadata only and are excluded from the provider operation total.` : ""}`.trim(),
      endpoints,
    },
    sourceEndpoints,
    legacyBindings: [...retainedSharedBindings, ...retainedLegacyBindings, ...retainedHooks],
  };
}

function mappedArtifacts(connector, oldCount, sourceResult, surfaceBuild, actions) {
  const sourceOperations = sourceResult.operations;
  const rows = sourceRows(sourceOperations, surfaceBuild.sourceEndpoints, connector, actions).map(stripSourceIndex);
  const documents = sourceResult.documents;
  const bundle = sourceBundle(documents);
  const counts = { total: sourceOperations.length, by_method: countByMethod(sourceOperations), by_kind: { rest: sourceOperations.length } };
  const sourceLock = {
    schema_version: 3,
    connector,
    captured_at: generatedAt,
    rest: {
      source_url: sourceResult.source_url,
      representation: sourceResult.representation,
      documents,
      source_bundle: bundle,
      counts,
      coverage_confidence: { level: sourceResult.coverage_confidence, basis: sourceResult.coverage_basis },
      operations: sourceOperations,
    },
  };
  const crosswalk = {
    schema_version: 2,
    connector,
    source_lock: `sources/${connector}-operation-source-lock.json`,
    generated_at: generatedAt,
    source_operations: rows.map((row) => ({
      ...row.source,
      method: row.method,
      path: row.path,
      crosswalk: {
        state: "exact",
        api_surface: row.api_surface,
        inventory: row.state === "enabled"
          ? { state: "materialized", kind: row.parity_class, id: row.declaration.stream ?? row.declaration.direct_reads ?? row.declaration.writes ?? null }
          : { state: "not_materialized", reason: row.rejection.reason },
      },
    })),
  };
  const enabled = rows.filter((row) => row.state === "enabled");
  const disabled = rows.filter((row) => row.state === "disabled");
  const directWrites = rows.filter((row) => row.parity_class === "direct_write");
  const disposition = {
    schema_version: 2,
    connector,
    generated_at: generatedAt,
    source_basis: {
      source_lock: `sources/${connector}-operation-source-lock.json`,
      source_url: sourceResult.source_url,
      source_bundle: bundle,
      source_operation_count: sourceOperations.length,
      coverage_confidence: { level: sourceResult.coverage_confidence, basis: sourceResult.coverage_basis },
    },
    summary: {
      old_api_surface_count: oldCount,
      api_surface_rows: sourceOperations.length,
      legacy_runtime_binding_rows: surfaceBuild.legacyBindings.length,
      operations_found: sourceOperations.length,
      coverage_confidence: { level: sourceResult.coverage_confidence, basis: sourceResult.coverage_basis },
      enabled_operations: enabled.length,
      enabled_percent: sourceOperations.length === 0 ? 0 : Number(((enabled.length / sourceOperations.length) * 100).toFixed(2)),
      disabled_operations: disabled.length,
      documented_deletes: rows.filter((row) => row.method === "DELETE").length,
      enabled_deletes: enabled.filter((row) => row.method === "DELETE").length,
      parity_class_counts: classCounts(rows),
      stream_bindings: rows.filter((row) => row.declaration.stream).length,
      write_actions: directWrites.filter((row) => row.declaration.writes?.length).length,
      terminal_commands: 0,
      live_certification: "pending",
      gap_ids: directWrites.length === 0 ? [] : ["generic-typed-destination-executor"],
      foundation_gaps: directWrites.length === 0 ? [] : [{ ...reverseETLEligibility.foundation_gap, count: directWrites.length, applies_to: "reverse_etl_eligibility attribute on direct_write operations" }],
      rejected_by_reason: disabled.length === 0 ? [] : [{ key: "declaration-pending", count: disabled.length }],
      transport: {
        contract: "docs/sync-transport-definition.md (PR #4286)",
        source_transport: {
          state: "declaration-pending",
          declaration_pending: {
            id: `sync-transport-source-definition-${connector}`,
            evidence: `docs/sync-transport-definition.md:15-38 lists the declaration fields; internal/connectors/defs/${connector}/sync_transport.json is absent.`,
            minimal_change: "Add the connector-owned source transport declaration and conformance evidence; no engine change is required.",
          },
        },
        destination_transport: reverseETLEligibility,
      },
      declaration_pending_ids: disabled.length === 0 ? [] : [`typed-operation-contract-${connector}`, `sync-transport-source-definition-${connector}`],
      declaration_pending: disabled.length === 0 ? [] : [{ id: `typed-operation-contract-${connector}`, count: disabled.length }],
    },
    ledger_dispositions: rows,
  };
  return { sourceLock, crosswalk, disposition };
}

function unavailableArtifacts(connector, oldCount, skip) {
  const sourceLock = {
    schema_version: 3,
    connector,
    state: "skipped",
    captured_at: generatedAt,
    skip,
    rest: {
      representation: "unavailable",
      retrieval_method: skip.retrieval_method,
      documents: [],
      counts: { total: null, by_method: {}, by_kind: { rest: null } },
      coverage_confidence: { level: "unavailable", basis: `${skip.reason}: ${skip.detail}` },
      operations: [],
    },
  };
  return {
    sourceLock,
    crosswalk: { schema_version: 2, connector, state: "skipped", source_lock: `sources/${connector}-operation-source-lock.json`, generated_at: generatedAt, reason: skip.reason, source_operations: [] },
    disposition: {
      schema_version: 2,
      connector,
      generated_at: generatedAt,
      source_basis: { state: "skipped", source_lock: `sources/${connector}-operation-source-lock.json`, reason: skip.reason, coverage_confidence: sourceLock.rest.coverage_confidence },
      summary: { state: "skipped", old_api_surface_count: oldCount, operations_found: null, coverage_confidence: sourceLock.rest.coverage_confidence, reason: skip.reason, live_certification: "pending" },
      ledger_dispositions: [],
    },
    report: { newCount: null, basis: `no-public-api-description; ${skip.retrieval_method}` },
  };
}

async function dynamicArtifacts(connector, oldCount, dynamic) {
  const document = await fetchPublishedDocument(dynamic.source_url, dynamic.retrieval_method);
  const confidence = { level: "dynamic-instance-dependent", basis: dynamic.detail };
  const sourceLock = {
    schema_version: 3,
    connector,
    state: "dynamic",
    captured_at: generatedAt,
    dynamic: { reason: dynamic.reason, detail: dynamic.detail },
    rest: {
      source_url: document.source_url,
      representation: "complete-rendered-reference",
      documents: [document],
      source_bundle: sourceBundle([document]),
      counts: { total: null, by_method: {}, by_kind: { rest: null } },
      coverage_confidence: confidence,
      operations: [],
    },
  };
  return {
    sourceLock,
    crosswalk: { schema_version: 2, connector, state: "dynamic", source_lock: `sources/${connector}-operation-source-lock.json`, generated_at: generatedAt, reason: dynamic.reason, source_operations: [] },
    disposition: {
      schema_version: 2,
      connector,
      generated_at: generatedAt,
      source_basis: { state: "dynamic", source_lock: `sources/${connector}-operation-source-lock.json`, reason: dynamic.reason, coverage_confidence: confidence },
      summary: { state: "dynamic", old_api_surface_count: oldCount, operations_found: null, coverage_confidence: confidence, reason: dynamic.reason, live_certification: "pending" },
      ledger_dispositions: [],
    },
    report: { newCount: null, basis: "dynamic instance-dependent" },
  };
}

async function writeArtifacts(connector, artifacts) {
  const sourceDir = resolve(defsRoot, connector, "sources");
  await mkdir(sourceDir, { recursive: true });
  await Promise.all([
    writeFile(resolve(sourceDir, `${connector}-operation-source-lock.json`), pretty(artifacts.sourceLock)),
    writeFile(resolve(sourceDir, `${connector}-operation-crosswalk.json`), pretty(artifacts.crosswalk)),
    writeFile(resolve(sourceDir, `${connector}-declaration-disposition.json`), pretty(artifacts.disposition)),
  ]);
}

async function generate(connector) {
  const bundle = resolve(defsRoot, connector);
  const [oldSurface, writes] = await Promise.all([
    readBaselineJSON(`internal/connectors/defs/${connector}/api_surface.json`),
    readJSON(resolve(bundle, "writes.json")).catch(() => ({ actions: [] })),
  ]);
  const oldCount = (oldSurface.endpoints ?? []).length;
  if (unavailableSources[connector]) {
    const artifacts = unavailableArtifacts(connector, oldCount, unavailableSources[connector]);
    await writeArtifacts(connector, artifacts);
    return { connector, oldCount, ...artifacts.report };
  }
  if (dynamicSources[connector]) {
    const artifacts = await dynamicArtifacts(connector, oldCount, dynamicSources[connector]);
    await writeArtifacts(connector, artifacts);
    return { connector, oldCount, ...artifacts.report };
  }
  const sourceResult = await extract(connector);
  const sourceSurface = buildSurface(connector, oldSurface, sourceResult.operations);
  const actions = new Set((writes.actions ?? []).map((action) => action.name));
  const artifacts = mappedArtifacts(connector, oldCount, sourceResult, sourceSurface, actions);
  await Promise.all([
    writeFile(resolve(bundle, "api_surface.json"), pretty(sourceSurface.surface)),
    writeArtifacts(connector, artifacts),
  ]);
  return { connector, oldCount, newCount: sourceResult.operations.length, basis: sourceResult.coverage_confidence === "machine-readable" ? "machine-readable spec" : "complete rendered reference", legacyBindings: sourceSurface.legacyBindings.length };
}

function batchFor(connector) {
  return Object.entries(batches).find(([, connectors]) => connectors.includes(connector))?.[0] ?? "?";
}

function reportResult(row) {
  const base = row.newCount === row.oldCount ? "no change" : row.newCount == null ? "not statically countable" : "regenerated";
  return row.legacyBindings > 0 ? `${base}; ${row.legacyBindings} runtime binding row(s) excluded` : base;
}

async function writeReport(rows) {
  const reportPath = resolve(root, ".planning/phases/issue-4292-parity-batches-8-10-r1/SOURCE-SURFACE-REPORT.md");
  const lines = [
    "# Issue #4292 provider-surface audit",
    "",
    `Generated ${generatedAt}. Each source-derived total below is the refreshed provider operation inventory; it is not derived from the previous api_surface.json. A no-change count is stated explicitly.`,
    "",
    "| Batch | Connector | Old api_surface count | New provider-operation count | Basis | Result |",
    "| --- | --- | ---: | ---: | --- | --- |",
    ...rows.map((row) => `| ${batchFor(row.connector)} | ${row.connector} | ${row.oldCount} | ${row.newCount ?? "—"} | ${row.basis} | ${reportResult(row)} |`),
    "",
    "`new provider-operation count` excludes any retained runtime-only coverage rows noted in the Result column. They preserve existing conformance while a behavior change updates or retires a binding that is not in the current provider inventory; they are never counted as evidence for a provider REST operation. Adobe Commerce is instance/module-dependent; TestRail, Eventbrite, and Greenhouse have no current credential-free public API description, so their totals are deliberately not fabricated.",
  ];
  await writeFile(reportPath, `${lines.join("\n")}\n`);
}

async function existingReport() {
  const rows = [];
  for (const connector of allConnectors) {
    const sourceDir = resolve(defsRoot, connector, "sources");
    const [sourceLock, disposition] = await Promise.all([
      readJSON(resolve(sourceDir, `${connector}-operation-source-lock.json`)),
      readJSON(resolve(sourceDir, `${connector}-declaration-disposition.json`)),
    ]);
    const level = disposition.summary?.coverage_confidence?.level;
    rows.push({
      connector,
      oldCount: disposition.summary?.old_api_surface_count,
      newCount: disposition.summary?.operations_found,
      basis: sourceLock.state === "skipped"
        ? `no-public-api-description; ${sourceLock.skip.retrieval_method}`
        : sourceLock.state === "dynamic"
          ? "dynamic instance-dependent"
          : level === "machine-readable"
            ? "machine-readable spec"
            : "complete rendered reference",
      legacyBindings: disposition.summary?.legacy_runtime_binding_rows ?? 0,
    });
  }
  await writeReport(rows);
}

const requested = process.argv.at(2);
const connectors = requested === "all" ? allConnectors : requested === "report" ? [] : batches[requested] ?? (allConnectors.includes(requested) ? [requested] : null);
if (!connectors) throw new Error(`usage: ${process.argv[1]} <8|9|10|all|report|connector>`);
if (requested === "report") {
  await existingReport();
  process.stdout.write("generated provider-surface report\n");
  process.exit(0);
}
const report = [];
for (const connector of connectors) {
  const row = await generate(connector);
  report.push(row);
  process.stdout.write(`generated ${connector}: ${row.oldCount} -> ${row.newCount ?? "dynamic/unavailable"}\n`);
}
if (requested === "all") await writeReport(report);
