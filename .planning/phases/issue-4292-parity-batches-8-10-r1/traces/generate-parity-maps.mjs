#!/usr/bin/env node

// Source-map generator for issue #4292. It deliberately consumes only the
// checked-in, source-reviewed endpoint inventory and public documentation
// bytes. It does not create connector commands, schemas, operations, writes,
// or sync transport declarations.

import { createHash } from "node:crypto";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { resolve } from "node:path";

const root = resolve(import.meta.dirname, "../../../..");
const defsRoot = resolve(root, "internal/connectors/defs");
const generatedAt = new Date().toISOString();

const batches = {
  "8": ["brex", "zoho-books", "testrail", "amplitude", "posthog", "metabase", "dbt", "looker", "mode", "dremio"],
  "9": ["coda", "clickup-api", "calendly", "greenhouse", "lever-hiring", "ashby", "workable", "recruitee", "hibob", "factorial"],
  "10": ["datadog", "pagerduty", "auth0", "okta", "firehydrant", "adobe-commerce-magento", "commercetools", "recharge", "docuseal", "eventbrite"],
};

const sourceURLOverrides = {
  brex: "https://developer.brex.com/",
  "zoho-books": "https://www.zoho.com/books/api/v3/openapi-all.zip",
  dbt: "https://raw.githubusercontent.com/dbt-labs/dbt-cloud-openapi-spec/master/openapi-v2.yaml",
  dremio: "https://docs.dremio.com/25.x/reference/api/",
};

const browserSkippedSources = {
  testrail: {
    attempted_url: "https://support.testrail.com/hc/en-us/categories/7076541806228",
    reason: "no-public-api-description",
    retrieval_method: "browser",
    detail: "The official rendered reference was requested in Chrome without credentials and returned a Cloudflare security-verification page rather than the published API reference. Per issue #4292 decision, no source pin or operation ledger is fabricated.",
  },
};

const reverseETLEligibility = {
  state: "foundation-gap",
  foundation_gap: {
    id: "generic-typed-destination-executor",
    evidence: "internal/app/issue_label_warehouse_transport.go:85-95 (acb85dc03): issueLabelDestinationReference is the only destination factory and BuildDestination enforces the GitHub issue-label contract.",
    minimal_change: "register a connector-neutral typed destination DefinitionFactory selected by the definition, with per-connector evidence, explicit source bindings, acknowledgement and per-mode apply strategies",
  },
};

function pretty(value) {
  return `${JSON.stringify(value, null, 2)}\n`;
}

function hash(bytes) {
  return createHash("sha256").update(bytes).digest("hex");
}

function operationSlug(method, path) {
  return `${method.toLowerCase()}-${path}`
    .replaceAll("{", "")
    .replaceAll("}", "")
    .replaceAll(/[^a-zA-Z0-9]+/g, "-")
    .replaceAll(/^-+|-+$/g, "")
    .toLowerCase();
}

function hasBinaryCategory(endpoint) {
  return /binary|attachment|download|upload|file|export/i.test(endpoint.excluded?.category ?? "");
}

function primaryClass(endpoint, writeActions, streams) {
  if (endpoint.covered_by?.stream && streams.has(endpoint.covered_by.stream)) {
    return { key: "etl", state: "enabled", binding: { stream: endpoint.covered_by.stream } };
  }
  if (endpoint.covered_by?.write && writeActions.has(endpoint.covered_by.write)) {
    return { key: "direct_write", state: "enabled", binding: { write: endpoint.covered_by.write } };
  }
  if (hasBinaryCategory(endpoint)) {
    return {
      key: /^(GET|HEAD)$/i.test(endpoint.method) ? "binary_read" : "binary_write",
      state: "disabled",
      binding: null,
    };
  }
  return {
    key: /^(GET|HEAD)$/i.test(endpoint.method) ? "direct_read" : "direct_write",
    state: "disabled",
    binding: null,
  };
}

function declaredOperation(connector, endpoint, index, sourceURL, classification) {
  // Some provider inventories intentionally expose more than one documented
  // operation at the same method/path. The checked-in API surface preserves
  // those distinct entries, so its index is part of this map's stable source
  // identity rather than silently collapsing them by endpoint shape.
  const sourceID = `${connector}.rest.${operationSlug(endpoint.method, endpoint.path)}-api-surface-${index + 1}`;
  const source = {
    source_lock: `sources/${connector}-operation-source-lock.json`,
    source_id: sourceID,
    source_url: sourceURL,
    source_location: `api_surface.json:endpoints[${index}]`,
    operation_id: sourceID,
    deprecated: endpoint.excluded?.category === "deprecated",
  };
  const apiSurface = {
    method: endpoint.method,
    path: endpoint.path,
    operation: null,
    ...(endpoint.covered_by ? { covered_by: endpoint.covered_by } : {}),
    ...(endpoint.excluded ? { excluded: endpoint.excluded } : {}),
  };
  const directWrite = classification.key === "direct_write";
  const state = classification.state;
  const pendingID = `typed-operation-contract-${connector}`;
  const foundation = state === "enabled"
    ? {
        state: "present",
        evidence: classification.binding.stream
          ? `internal/connectors/defs/${connector}/streams.json: existing stream ${JSON.stringify(classification.binding.stream)} binds this source endpoint.`
          : `internal/connectors/defs/${connector}/writes.json: typed write action ${JSON.stringify(classification.binding.write)} binds this source endpoint.`,
        contract: classification.binding.stream
          ? null
          : { kind: "typed_write_action", id: classification.binding.write },
      }
    : {
        state: "present",
        evidence: "No shared engine change is requested for this row; the missing work is a connector-local declaration bound to the pinned operation.",
        declaration_pending: {
          id: pendingID,
          evidence: `internal/connectors/defs/${connector}/api_surface.json:endpoints[${index}]; ${source.source_location}.`,
          minimal_change: "Add a bounded connector-owned operation contract and runnable command from the pinned public document, or retain a disabled source disposition when the source shape is not executable.",
        },
      };
  const rejection = state === "enabled" ? null : {
    reason: "declaration-pending",
    recoverable: true,
    detail: "The engine shape is already available; this source operation awaits its connector-local typed contract and/or runnable command declaration.",
    evidence: `internal/connectors/defs/${connector}/api_surface.json:endpoints[${index}].`,
  };
  return {
    method: endpoint.method,
    path: endpoint.path,
    parity_class: classification.key,
    api_surface: apiSurface,
    source,
    state,
    foundation,
    ...(directWrite ? { reverse_etl_eligibility: reverseETLEligibility } : {}),
    rejection,
    declaration: state === "enabled"
      ? {
          status: classification.binding.stream
            ? `enabled; existing etl stream ${JSON.stringify(classification.binding.stream)} binds the pinned source contract`
            : `enabled; typed direct_write action ${JSON.stringify(classification.binding.write)} binds the pinned source contract`,
          ...(classification.binding.stream ? { stream: classification.binding.stream } : { write: classification.binding.write }),
        }
      : { status: `disabled; declaration-pending ${pendingID}`, contract: null },
  };
}

function classCounts(rows) {
  const counts = new Map();
  for (const row of rows) counts.set(row.parity_class, (counts.get(row.parity_class) ?? 0) + 1);
  return ["direct_read", "direct_write", "etl", "reverse_etl", "binary_read", "binary_write"]
    .map((key) => ({ key, count: counts.get(key) ?? 0 }));
}

async function fetchSource(url) {
  const response = await fetch(url, {
    redirect: "follow",
    headers: { "user-agent": "Polymetrics-source-lock/1.0 (+https://github.com/polymetrics-ai/cli)" },
    signal: AbortSignal.timeout(60_000),
  });
  if (!response.ok) throw new Error(`public documentation fetch returned HTTP ${response.status}`);
  const bytes = Buffer.from(await response.arrayBuffer());
  return {
    source_url: response.url,
    requested_url: url,
    sha256: hash(bytes),
    bytes: bytes.length,
    content_type: response.headers.get("content-type"),
    retrieval_method: "http",
    representation: "published-document",
  };
}

async function readJSON(path) {
  return JSON.parse(await readFile(path, "utf8"));
}

async function generateSkipped(connector, skip) {
  const sourceDir = resolve(defsRoot, connector, "sources");
  await mkdir(sourceDir, { recursive: true });
  const sourceLock = {
    schema_version: 2,
    connector,
    state: "skipped",
    captured_at: generatedAt,
    skip,
  };
  const crosswalk = {
    schema_version: 1,
    connector,
    state: "skipped",
    source_lock: `sources/${connector}-operation-source-lock.json`,
    generated_at: generatedAt,
    reason: skip.reason,
    source_operations: [],
  };
  const disposition = {
    schema_version: 1,
    connector,
    generated_at: generatedAt,
    source_basis: { state: "skipped", source_lock: `sources/${connector}-operation-source-lock.json`, reason: skip.reason },
    summary: { state: "skipped", reason: skip.reason, live_certification: "pending" },
    ledger_dispositions: [],
  };
  await Promise.all([
    writeFile(resolve(sourceDir, `${connector}-operation-source-lock.json`), pretty(sourceLock)),
    writeFile(resolve(sourceDir, `${connector}-operation-crosswalk.json`), pretty(crosswalk)),
    writeFile(resolve(sourceDir, `${connector}-declaration-disposition.json`), pretty(disposition)),
  ]);
}

async function generate(connector) {
  if (browserSkippedSources[connector]) return generateSkipped(connector, browserSkippedSources[connector]);
  const bundle = resolve(defsRoot, connector);
  const [metadata, apiSurface, streams, writes] = await Promise.all([
    readJSON(resolve(bundle, "metadata.json")),
    readJSON(resolve(bundle, "api_surface.json")),
    readJSON(resolve(bundle, "streams.json")),
    readJSON(resolve(bundle, "writes.json")).catch(() => ({ actions: [] })),
  ]);
  const sourceURL = sourceURLOverrides[connector] ?? metadata.docs_url;
  const sourceDocument = await fetchSource(sourceURL);
  const streamNames = new Set((streams.streams ?? []).map((stream) => stream.name));
  const actionNames = new Set((writes.actions ?? []).map((action) => action.name));
  const endpoints = apiSurface.endpoints ?? [];
  const rows = endpoints.map((endpoint, index) => declaredOperation(
    connector,
    endpoint,
    index,
    sourceDocument.source_url,
    primaryClass(endpoint, actionNames, streamNames),
  ));
  const sourceOperations = rows.map((row) => ({
    id: row.source.source_id,
    protocol: "rest",
    method: row.method,
    path: row.path,
    operation_id: row.source.operation_id,
    deprecated: row.source.deprecated,
    source_location: row.source.source_location,
  }));
  const directWrites = rows.filter((row) => row.parity_class === "direct_write");
  const enabled = rows.filter((row) => row.state === "enabled");
  const disabled = rows.filter((row) => row.state === "disabled");
  const sourceDir = resolve(bundle, "sources");
  await mkdir(sourceDir, { recursive: true });
  const sourceLock = {
    schema_version: 2,
    connector,
    captured_at: generatedAt,
    rest: {
      ...sourceDocument,
      inventory_basis: "The existing source-reviewed api_surface.json endpoint inventory is crosswalked verbatim; no request, response, pagination, or body schema is inferred by this map.",
      operation_counts: Object.fromEntries([...new Set(endpoints.map((endpoint) => endpoint.method))].sort().map((method) => [method, endpoints.filter((endpoint) => endpoint.method === method).length])),
      operations: sourceOperations,
    },
  };
  const crosswalk = {
    schema_version: 1,
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
          ? { state: "materialized", kind: row.parity_class, id: row.declaration.write ?? row.declaration.stream ?? null }
          : { state: "not_materialized", reason: row.rejection.reason },
      },
    })),
  };
  const disposition = {
    schema_version: 1,
    connector,
    generated_at: generatedAt,
    source_basis: {
      source_lock: `sources/${connector}-operation-source-lock.json`,
      source_url: sourceDocument.source_url,
      source_sha256: sourceDocument.sha256,
      source_bytes: sourceDocument.bytes,
      source_operation_count: rows.length,
    },
    summary: {
      api_surface_rows: endpoints.length,
      exact_source_rows: rows.length,
      declared_operations: rows.length,
      declared_percent: 100,
      enabled_operations: enabled.length,
      enabled_percent: rows.length === 0 ? 0 : Number(((enabled.length / rows.length) * 100).toFixed(2)),
      disabled_operations: disabled.length,
      documented_deletes: rows.filter((row) => row.method === "DELETE").length,
      enabled_deletes: enabled.filter((row) => row.method === "DELETE").length,
      parity_class_counts: classCounts(rows),
      stream_bindings: rows.filter((row) => row.declaration.stream).length,
      writes_actions: directWrites.filter((row) => row.declaration.write).length,
      terminal_commands: 0,
      live_certification: "pending",
      gap_ids: directWrites.length === 0 ? [] : ["generic-typed-destination-executor"],
      foundation_gaps: directWrites.length === 0 ? [] : [{ ...reverseETLEligibility.foundation_gap, count: directWrites.length, applies_to: "reverse_etl_eligibility attribute on direct_write operations" }],
      rejected_by_reason: [
        ...(disabled.length === 0 ? [] : [{ key: "declaration-pending", count: disabled.length }]),
      ],
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
      declaration_pending_ids: [pendingID(connector), `sync-transport-source-definition-${connector}`],
      declaration_pending: disabled.length === 0 ? [] : [{ id: pendingID(connector), count: disabled.length }],
    },
    ledger_dispositions: rows,
  };
  await Promise.all([
    writeFile(resolve(sourceDir, `${connector}-operation-source-lock.json`), pretty(sourceLock)),
    writeFile(resolve(sourceDir, `${connector}-operation-crosswalk.json`), pretty(crosswalk)),
    writeFile(resolve(sourceDir, `${connector}-declaration-disposition.json`), pretty(disposition)),
  ]);
}

function pendingID(connector) {
  return `typed-operation-contract-${connector}`;
}

const requestedBatch = process.argv.at(2);
if (!requestedBatch || !batches[requestedBatch]) {
  throw new Error(`usage: ${process.argv[1]} <8|9|10>`);
}
for (const connector of batches[requestedBatch]) {
  await generate(connector);
  process.stdout.write(`generated ${connector}\n`);
}
