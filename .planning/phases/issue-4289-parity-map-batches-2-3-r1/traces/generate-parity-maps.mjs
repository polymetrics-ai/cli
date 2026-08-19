#!/usr/bin/env node

import { execFileSync } from "node:child_process";
import { createHash } from "node:crypto";
import { existsSync } from "node:fs";
import { mkdir, readFile, writeFile } from "node:fs/promises";
import { gunzipSync } from "node:zlib";
import path from "node:path";

const root = process.cwd();
const generatedAt = "2026-08-19T00:00:00Z";
const batch2Draft = "fm/cli-top100-declaration-batch-r1-inc2-wip";
// HEAD retains the connector-owned command/stream bindings authored by this
// issue before source inventory expansion. It is input for binding metadata,
// never an operation-inventory boundary; the public lock below remains that.
const bindingRevision = "HEAD";
const batch2 = ["grafana", "trello", "slack", "n8n", "google-calendar", "gmail", "twilio", "amazon-sqs", "elasticsearch"];
const batch3 = ["gong", "google-ads", "facebook-marketing", "linkedin-ads", "aircall", "xero", "paypal-transaction", "gocardless", "amazon-seller-partner", "miro"];

// These are documentation artifacts only. No endpoint, credential, or provider
// action is invoked by this map generator.
const batch3Sources = {
  "gong": { url: "https://gong.app.gong.io/ajax/settings/api/documentation/specs?version=", parser: "openapi", confidence: { level: "complete", basis: "provider-published OpenAPI document" } },
  "google-ads": { url: "https://googleads.googleapis.com/$discovery/rest?version=v22", parser: "discovery", confidence: { level: "complete", basis: "provider-published Google Discovery document" } },
  "facebook-marketing": { url: "https://codeload.github.com/facebook/facebook-business-sdk-codegen/tar.gz/refs/heads/main", parser: "facebook-codegen", confidence: { level: "complete", basis: "all current provider-published Facebook Business SDK code-generation API declarations; Graph routes are explicitly documented as object-ID node/edge templates, so totals count named owner-type/method/edge declarations rather than fabricated runtime identifiers" } },
  "linkedin-ads": { url: "https://learn.microsoft.com/_sitemaps/linkedin_en-us_1.xml", parser: "linkedin-rendered", confidence: { level: "complete", basis: "provider sitemap plus every current LinkedIn Marketing rendered-reference page" } },
  "aircall": { url: "https://developers.aircall.io/api-references", parser: "aircall-rendered", confidence: { level: "complete", basis: "provider's single complete rendered Public API reference; every rendered endpoint index entry is parsed, while the two literal example IDs are excluded in favour of their documented parameter templates" } },
  "xero": { url: "https://raw.githubusercontent.com/XeroAPI/Xero-OpenAPI/master/xero_accounting.yaml", parser: "surface-machine", confidence: { level: "complete", basis: "provider-published Xero Accounting OpenAPI document" } },
  "paypal-transaction": {
    url: "https://codeload.github.com/paypal/paypal-rest-api-specifications/tar.gz/refs/heads/main",
    parser: "paypal-specifications",
    confidence: {
      level: "complete",
      basis: "all provider-published PayPal REST OpenAPI documents under openapi/ in the official paypal-rest-api-specifications repository"
    }
  },
  "gocardless": { url: "https://developer.gocardless.com/openapi-schema-public.json", parser: "openapi", confidence: { level: "complete", basis: "provider-published GoCardless OpenAPI document served to its public API reference" } },
  "amazon-seller-partner": { url: "https://codeload.github.com/amzn/selling-partner-api-models/tar.gz/refs/heads/main", parser: "amazon-models", confidence: { level: "complete", basis: "all provider-published Selling Partner OpenAPI model documents under models/" } },
  "miro": { url: "https://raw.githubusercontent.com/miroapp/api-clients/main/packages/generator/spec.json", parser: "openapi", confidence: { level: "complete", basis: "provider-published Miro OpenAPI document linked by the API reference" } }
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

function readBaselineSurface(connector) {
  return JSON.parse(execFileSync("git", ["show", `${bindingRevision}:internal/connectors/defs/${connector}/api_surface.json`], { encoding: "utf8" }));
}

function mergeBaselineSurface(connector, surface) {
  const baseline = readBaselineSurface(connector);
  const seen = new Set();
  const endpoints = [];
  for (const endpoint of [...baseline.endpoints, ...surface.endpoints]) {
    // Query-bearing provider surfaces may intentionally have more than one
    // binding for one normalized resource path (for example Grafana folders).
    // Deduplicate exact endpoint text only; source discovery below still uses
    // normalized keys to bind a documented operation to its existing contract.
    const key = `${endpoint.method.toUpperCase()} ${endpoint.path}`;
    if (seen.has(key)) continue;
    seen.add(key);
    endpoints.push(endpoint);
  }
  return { ...surface, endpoints };
}

function sourcePathToBundlePath(connector, input) {
  if (connector === "grafana" && !input.startsWith("/api/")) return `/api${input}`;
  if (connector === "gmail" && input.startsWith("/gmail/v1/")) return input.slice("/gmail/v1".length);
  if (connector === "twilio" && input.startsWith("/2010-04-01/")) return input.slice("/2010-04-01".length);
  return input;
}

function canonicalPath(input, preserveParameterNames = false) {
  const path = input
    .replace(/\?.*$/, "")
    .replace(/(^|\/):[^/]+(?=\/|$)/g, "$1{}")
    .replace(/\/+/g, "/")
    .replace(/\/$/, "") || "/";
  if (!preserveParameterNames) return path.replace(/\{[^}]+\}/g, "{}");
  return path.replace(/\{([^}]+)\}/g, (_match, name) => `{${String(name).trim().toLowerCase().replace(/[^a-z0-9]+/g, "_")}}`);
}

function operationKey(connector, method, pathname) {
  return `${method.toUpperCase()} ${canonicalPath(pathname, connector === "facebook-marketing")}`;
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

async function fetchPublicBytes(url) {
  const response = await fetch(url, { headers: { "User-Agent": "polymetrics-source-lock/1" } });
  if (!response.ok) throw new Error(`${url}: HTTP ${response.status}`);
  const bytes = Buffer.from(await response.arrayBuffer());
  return {
    bytes,
    sha256: createHash("sha256").update(bytes).digest("hex"),
    contentType: response.headers.get("content-type") || "application/octet-stream"
  };
}

async function fetchPublicJSON(url) {
  const response = await fetch(url, { headers: { "User-Agent": "polymetrics-source-lock/1" } });
  if (!response.ok) throw new Error(`${url}: HTTP ${response.status}`);
  const bytes = Buffer.from(await response.arrayBuffer());
  return {
    json: JSON.parse(bytes.toString("utf8")),
    sha256: createHash("sha256").update(bytes).digest("hex"),
    bytes: bytes.length,
    contentType: response.headers.get("content-type") || "application/json"
  };
}

async function fetchPublicText(url) {
  const response = await fetch(url, { headers: { "User-Agent": "polymetrics-source-lock/1" } });
  if (!response.ok) throw new Error(`${url}: HTTP ${response.status}`);
  const bytes = Buffer.from(await response.arrayBuffer());
  return {
    text: bytes.toString("utf8"),
    sha256: createHash("sha256").update(bytes).digest("hex"),
    bytes: bytes.length,
    contentType: response.headers.get("content-type") || "text/plain"
  };
}

const sourceMethods = new Set(["get", "post", "put", "patch", "delete", "head"]);

function openAPIOperations(connector, document, sourceURL) {
  const operations = [];
  for (const [pathname, item] of Object.entries(document.paths || {})) {
    for (const [method, operation] of Object.entries(item || {})) {
      if (!sourceMethods.has(method.toLowerCase())) continue;
      operations.push({
        id: sourceID(connector, method, pathname, operations.length),
        protocol: "rest",
        method: method.toUpperCase(),
        path: pathname,
        operation_id: operation?.operationId || null,
        deprecated: Boolean(operation?.deprecated),
        source_location: `paths[${JSON.stringify(pathname)}].${method.toLowerCase()}`,
        source_url: sourceURL
      });
    }
  }
  return operations;
}

function discoveryOperations(connector, document, sourceURL) {
  const operations = [];
  function visit(resources) {
    for (const resource of Object.values(resources || {})) {
      for (const method of Object.values(resource.methods || {})) {
        if (!sourceMethods.has(method.httpMethod?.toLowerCase())) continue;
        const pathname = `/${String(method.path || "").replace(/^\//, "")}`;
        operations.push({
          id: sourceID(connector, method.httpMethod, pathname, operations.length),
          protocol: "rest",
          method: method.httpMethod.toUpperCase(),
          path: pathname,
          operation_id: method.id || null,
          deprecated: Boolean(method.deprecated),
          source_location: `Discovery method ${JSON.stringify(method.id || pathname)}`,
          source_url: sourceURL
        });
      }
      visit(resource.resources);
    }
  }
  visit(document.resources);
  return operations;
}

async function amazonModelOperations(connector, archiveURL) {
  const archive = await fetchPublicBytes(archiveURL);
  const models = tarEntries(archive.bytes)
    .filter((entry) => /(?:^|\/)models\/.*\.json$/.test(entry.path))
    .map((entry) => ({
      path: entry.path.replace(/^.*?models\//, "models/"),
      content: entry.content
    }));
  const operations = [];
  const sourceDocuments = [];
  for (const model of models) {
    const sourceURL = `https://raw.githubusercontent.com/amzn/selling-partner-api-models/main/${model.path}`;
    const sha256 = createHash("sha256").update(model.content).digest("hex");
    sourceDocuments.push({ path: model.path, source_url: sourceURL, sha256, bytes: model.content.length });
    for (const operation of openAPIOperations(connector, JSON.parse(model.content.toString("utf8")), sourceURL)) {
      operation.id = `${connector}.rest.${model.path.replace(/[^A-Za-z0-9]+/g, ".")}.${operation.id.split(".").pop()}`;
      operation.source_location = `${model.path}:${operation.source_location}`;
      operations.push(operation);
    }
  }
  return { operations, bytes: archive.bytes.length, sha256: archive.sha256, contentType: archive.contentType, sourceDocuments };
}

async function paypalSpecificationsOperations(connector, archiveURL) {
  const archive = await fetchPublicBytes(archiveURL);
  const specifications = tarEntries(archive.bytes)
    .filter((entry) => /(?:^|\/)openapi\/[^/]+\.json$/.test(entry.path))
    .map((entry) => ({
      path: entry.path.replace(/^.*?openapi\//, "openapi/"),
      content: entry.content
    }))
    .sort((left, right) => left.path.localeCompare(right.path));
  if (specifications.length === 0) throw new Error("PayPal source archive has no OpenAPI documents");
  const operations = [];
  const sourceDocuments = [];
  for (const specification of specifications) {
    const sourceURL = `https://raw.githubusercontent.com/paypal/paypal-rest-api-specifications/main/${specification.path}`;
    const sha256 = createHash("sha256").update(specification.content).digest("hex");
    sourceDocuments.push({ path: specification.path, source_url: sourceURL, sha256, bytes: specification.content.length });
    const document = JSON.parse(specification.content.toString("utf8"));
    const prefix = specification.path.replace(/[^A-Za-z0-9]+/g, ".").replace(/^\.|\.$/g, "").toLowerCase();
    for (const [index, operation] of openAPIOperations(connector, document, sourceURL).entries()) {
      operations.push({
        ...operation,
        id: `${connector}.rest.${prefix}.${index + 1}`,
        source_location: `${specification.path}:${operation.source_location}`
      });
    }
  }
  return { operations, bytes: archive.bytes.length, sha256: archive.sha256, contentType: archive.contentType, sourceDocuments };
}

async function linkedInRenderedOperations(connector, sitemapURL) {
  const sitemap = await fetchPublicText(sitemapURL);
  const pages = [...sitemap.text.matchAll(/<loc>(https:\/\/learn\.microsoft\.com\/en-us\/linkedin\/marketing\/[^<]+)<\/loc>/g)]
    .map((match) => match[1])
    .filter((url) => url.includes("?view="));
  const sourceDocuments = [];
  const candidates = [];
  let bytes = sitemap.bytes;
  const hash = createHash("sha256").update(sitemap.sha256);
  for (let index = 0; index < pages.length; index += 8) {
    const batch = await Promise.all(pages.slice(index, index + 8).map(async (url) => ({ url, artifact: await fetchPublicText(url) })));
    for (const { url, artifact } of batch) {
      bytes += artifact.bytes;
      hash.update(artifact.sha256);
      sourceDocuments.push({ source_url: url, sha256: artifact.sha256, bytes: artifact.bytes });
      const rendered = artifact.text.replace(/&amp;/g, "&").replace(/&#x2F;/g, "/");
      for (const match of rendered.matchAll(/\b(GET|POST|PUT|PATCH|DELETE)\s+(https:\/\/api\.linkedin\.com\/(?:rest|v2)\/(?:[^\s'"`<{}]+|\{[^}]+\})+)/g)) {
        const method = match[1];
        const requestURL = match[2].replace(/[),.;]+$/, "");
        const parsed = new URL(requestURL);
        candidates.push({ method, path: decodeURIComponent(parsed.pathname), source_url: url, source_location: `rendered request ${method} ${requestURL}` });
      }
    }
  }
  const unique = new Map();
  for (const candidate of candidates) {
    const key = `${candidate.method} ${candidate.path}`;
    if (!unique.has(key)) unique.set(key, candidate);
  }
  const documented = [...unique.values()];
  const templates = documented.filter((candidate) => /\{[^}]+\}/.test(candidate.path));
  const isLiteralExample = (pathname) => pathname.split("/").some((segment) => /^(?:\d+|urn:[^/]+|[0-9a-f]{16,})$/i.test(segment));
  const matchesTemplate = (template, pathname) => {
    const expected = template.split("/");
    const actual = pathname.split("/");
    return expected.length === actual.length && expected.every((segment, index) => /\{[^}]+\}/.test(segment) || segment === actual[index]);
  };
  // Microsoft Learn renders literal IDs next to the reusable parameterized
  // request. Count the documented template, not each illustrative account or
  // organization value, while retaining literal segments that have no matching
  // parameterized operation in the complete reference.
  const operations = documented
    .filter((candidate) => !isLiteralExample(candidate.path) || !templates.some((template) => template.method === candidate.method && matchesTemplate(template.path, candidate.path)))
    .map((candidate, index) => ({
    id: sourceID(connector, candidate.method, candidate.path, index),
    protocol: "rest",
    method: candidate.method,
    path: candidate.path,
    operation_id: null,
    deprecated: false,
    source_location: candidate.source_location,
    source_url: candidate.source_url
  }));
  return { operations, bytes, sha256: hash.digest("hex"), sourceDocuments };
}

async function aircallRenderedOperations(connector, referenceURL) {
  const artifact = await fetchPublicText(referenceURL);
  const candidates = [];
  for (const match of artifact.text.matchAll(/<span class="http-method[^>]*">(GET|POST|PUT|PATCH|DELETE)<\/span>\s*(\/v\d+(?:\/[^\s<]*)?)/g)) {
    const method = match[1];
    const pathname = match[2].replace(/[),.;]+$/, "");
    // The reference lists two literal examples beside the parameterized users
    // endpoint (a numeric ID and an email). They are illustrations, not extra
    // operations; every other endpoint-index entry is source material.
    if (/(^|\/)\d+(?:\/|$)/.test(pathname) || pathname.includes("@")) continue;
    candidates.push({ method, path: pathname, source_location: `rendered endpoint index ${method} ${pathname}` });
  }
  const unique = new Map();
  for (const candidate of candidates) {
    const key = `${candidate.method} ${candidate.path}`;
    if (!unique.has(key)) unique.set(key, candidate);
  }
  const operations = [...unique.values()].map((candidate, index) => ({
    id: sourceID(connector, candidate.method, candidate.path, index),
    protocol: "rest",
    method: candidate.method,
    path: candidate.path,
    operation_id: null,
    deprecated: false,
    source_location: candidate.source_location,
    source_url: referenceURL
  }));
  return {
    operations,
    bytes: artifact.bytes,
    sha256: artifact.sha256,
    sourceDocuments: [{ source_url: referenceURL, sha256: artifact.sha256, bytes: artifact.bytes }]
  };
}

function facebookOwnerParameter(specPath) {
  const objectName = path.basename(specPath, ".json");
  const known = {
    Ad: "ad_id",
    AdAccount: "ad_account_id",
    AdSet: "adset_id",
    Campaign: "campaign_id",
    User: "user_id"
  };
  if (known[objectName]) return known[objectName];
  return `${objectName.replace(/([a-z0-9])([A-Z])/g, "$1_$2").replace(/[^A-Za-z0-9]+/g, "_").toLowerCase()}_id`;
}

function facebookGraphPath(specPath, api) {
  const owner = facebookOwnerParameter(specPath);
  const endpoint = String(api.endpoint || api.basePath || "").replace(/^\/+|\/+$/g, "");
  return endpoint ? `/{${owner}}/${endpoint}` : `/{${owner}}`;
}

function tarOctal(buffer, offset, length) {
  const raw = buffer.subarray(offset, offset + length).toString("utf8").replace(/\0.*$/, "").trim();
  return raw ? Number.parseInt(raw, 8) : 0;
}

function tarString(buffer, offset, length) {
  return buffer.subarray(offset, offset + length).toString("utf8").replace(/\0.*$/, "");
}

function tarEntries(gzip) {
  const tar = gunzipSync(gzip);
  const entries = [];
  let offset = 0;
  let pendingPath = null;
  while (offset + 512 <= tar.length) {
    const header = tar.subarray(offset, offset + 512);
    if (header.every((byte) => byte === 0)) break;
    const size = tarOctal(header, 124, 12);
    const type = tarString(header, 156, 1);
    const prefix = tarString(header, 345, 155);
    const named = [prefix, tarString(header, 0, 100)].filter(Boolean).join("/");
    const contentStart = offset + 512;
    const content = tar.subarray(contentStart, contentStart + size);
    if (type === "x") {
      for (const record of content.toString("utf8").split("\n")) {
        const marker = record.indexOf(" path=");
        if (marker >= 0) pendingPath = record.slice(marker + " path=".length);
      }
    } else if (type === "0" || type === "") {
      entries.push({ path: pendingPath || named, content });
      pendingPath = null;
    }
    offset = contentStart + Math.ceil(size / 512) * 512;
  }
  return entries;
}

async function facebookCodegenOperations(connector, archiveURL) {
  const archive = await fetchPublicBytes(archiveURL);
  const specs = tarEntries(archive.bytes)
    .filter((entry) => /(?:^|\/)api_specs\/specs\/[^/]+\.json$/.test(entry.path))
    .map((entry) => ({
      path: entry.path.replace(/^.*?api_specs\/specs\//, "api_specs/specs/"),
      content: entry.content
    }));
  const operations = [];
  const sourceDocuments = [];
  for (const spec of specs) {
    const sourceURL = `https://raw.githubusercontent.com/facebook/facebook-business-sdk-codegen/main/${spec.path}`;
    const sha256 = createHash("sha256").update(spec.content).digest("hex");
    sourceDocuments.push({ path: spec.path, source_url: sourceURL, sha256, bytes: spec.content.length });
    const document = JSON.parse(spec.content.toString("utf8"));
    for (const [apiIndex, api] of (document.apis || []).entries()) {
      if (!sourceMethods.has(String(api.method || "").toLowerCase())) continue;
      const pathname = facebookGraphPath(spec.path, api);
      operations.push({
        id: `${connector}.graph.${spec.path.replace(/[^A-Za-z0-9]+/g, ".")}.${apiIndex + 1}`,
        protocol: "rest",
        method: api.method.toUpperCase(),
        path: pathname,
        operation_id: api.name || null,
        deprecated: false,
        source_location: `${spec.path}:apis[${apiIndex}] (owner=${path.basename(spec.path, ".json")}; Graph node/edge template)`,
        source_url: sourceURL
      });
    }
  }
  return { operations, bytes: archive.bytes.length, sha256: archive.sha256, contentType: archive.contentType, sourceDocuments };
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
    id: "generic-typed-destination-app-dispatch",
    evidence: "internal/app/transport_dispatch.go:53-67: after preflight, the persisted App dispatch admits only semantic managed targets or local-warehouse dedupe targets; declarative_typed_destination is neither, so it is not selected for a real App/CLI run.",
    minimal_change: "admit the exact preflighted declarative_api/declarative_typed_destination reference through persisted App/CLI dispatch while preserving definition-selected mode, action, source binding, approval, and authorization checks"
  };
}

function destinationActionMultiplicityGap() {
  return {
    id: "declarative-typed-destination-action-multiplicity",
    evidence: "internal/connectors/sync_transport.go:388-415 rejects duplicate apply strategies and requires one strategy for every mode; ApplyStrategyFor at :471-480 resolves that single action. A connection therefore cannot select every eligible action with distinct schemas and input mappings.",
    minimal_change: "extend the closed destination contract to select one exact declaration-owned action and its exact input_fields binding per approved route, without accepting a caller-supplied operation, method, URL, or body"
  };
}

function actionNames(endpoint) {
  const covered = endpoint?.covered_by || {};
  return [covered.write, ...(Array.isArray(covered.writes) ? covered.writes : [])].filter(Boolean);
}

async function typedWriteActionDispositions(connector, rows) {
  let writes;
  try {
    writes = JSON.parse(await readFile(definitionPath(connector, "writes.json"), "utf8"));
  } catch (error) {
    if (error.code === "ENOENT") return [];
    throw error;
  }
  const mapped = new Map();
  for (const row of rows) {
    for (const action of actionNames(row.api_surface)) {
      const sourceIDs = mapped.get(action) || [];
      sourceIDs.push(row.source.source_id);
      mapped.set(action, sourceIDs);
    }
  }
  const multiplicity = destinationActionMultiplicityGap();
  return writes.actions.map((action) => {
    const sourceIDs = mapped.get(action.name) || [];
    return {
      action: action.name,
      semantic_eligibility: "eligible",
      closed_destination_contract: "representable",
      source_row_binding: sourceIDs.length > 0
        ? { state: "source-bound", source_ids: sourceIDs }
        : {
            state: "declaration-pending",
            detail: "The existing typed action is individually representable, but the pinned source inventory has no exact source-row binding for its base-relative or unmatched action path.",
            minimal_change: "Pin the provider operation that exactly backs this existing action, then add its source-row and installed direct-command declarations; do not infer a request contract."
          },
      destination_binding: {
        state: "foundation-gap",
        foundation_gap: multiplicity
      }
    };
  });
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
    // The ledger's path is the provider-published operation path. The nested
    // api_surface object records the exact connector binding when normalized
    // parameter spellings coalesce to one executable surface endpoint.
    path: source.path,
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
  const plan = batch3Sources[connector];
  let artifact;
  let operations;
  let sourceDocuments;
  if (plan.parser === "openapi") {
    artifact = await fetchPublicJSON(plan.url);
    operations = openAPIOperations(connector, artifact.json, plan.url);
  } else if (plan.parser === "discovery") {
    artifact = await fetchPublicJSON(plan.url);
    operations = discoveryOperations(connector, artifact.json, plan.url);
  } else if (plan.parser === "amazon-models") {
    const aggregate = await amazonModelOperations(connector, plan.url);
    artifact = { sha256: aggregate.sha256, bytes: aggregate.bytes, contentType: aggregate.contentType };
    operations = aggregate.operations;
    sourceDocuments = aggregate.sourceDocuments;
  } else if (plan.parser === "paypal-specifications") {
    const aggregate = await paypalSpecificationsOperations(connector, plan.url);
    artifact = { sha256: aggregate.sha256, bytes: aggregate.bytes, contentType: aggregate.contentType };
    operations = aggregate.operations;
    sourceDocuments = aggregate.sourceDocuments;
  } else if (plan.parser === "linkedin-rendered") {
    const aggregate = await linkedInRenderedOperations(connector, plan.url);
    artifact = { sha256: aggregate.sha256, bytes: aggregate.bytes, contentType: "application/vnd.microsoft.learn.rendered-reference-set+html" };
    operations = aggregate.operations;
    sourceDocuments = aggregate.sourceDocuments;
  } else if (plan.parser === "aircall-rendered") {
    const aggregate = await aircallRenderedOperations(connector, plan.url);
    artifact = { sha256: aggregate.sha256, bytes: aggregate.bytes, contentType: "text/html; rendered-aircall-reference" };
    operations = aggregate.operations;
    sourceDocuments = aggregate.sourceDocuments;
  } else if (plan.parser === "facebook-codegen") {
    const aggregate = await facebookCodegenOperations(connector, plan.url);
    artifact = { sha256: aggregate.sha256, bytes: aggregate.bytes, contentType: aggregate.contentType };
    operations = aggregate.operations;
    sourceDocuments = aggregate.sourceDocuments;
  } else {
    artifact = await fetchPublicArtifact(plan.url);
    operations = surface.endpoints.map((endpoint, index) => ({
      id: sourceID(connector, endpoint.method, endpoint.path, index),
      protocol: "rest",
      method: endpoint.method.toUpperCase(),
      path: endpoint.path,
      operation_id: endpoint.operation?.notes?.match(/operation_id=([^;\s]+)/)?.[1] || null,
      deprecated: endpoint.operation?.model === "deprecated",
      source_location: `documented endpoint ${endpoint.method.toUpperCase()} ${endpoint.path}`,
      source_url: plan.url
    }));
  }
  const rootCounts = { rest: operations.length, graphql_query: 0, graphql_mutation: 0, total: operations.length };
  const restCounts = { total: operations.length, by_kind: { rest: operations.length }, by_method: methodCounts(operations) };
  return {
    schema_version: 2,
    connector,
    captured_at: generatedAt,
    counts: rootCounts,
    rest: {
      source_url: plan.url,
      sha256: artifact.sha256,
      bytes: artifact.bytes,
      format: artifact.contentType,
      inventory_basis: plan.parser === "facebook-codegen"
        ? "all current Facebook Business SDK code-generation API declarations; each operation is an owner-type plus HTTP method plus Graph node/edge template, because the concrete node identifier is instance-dependent"
        : plan.parser === "surface-rendered"
          ? "complete rendered provider API reference reconciled to the connector-owned API surface"
        : plan.parser === "linkedin-rendered"
          ? "parsed from every current LinkedIn Marketing rendered-reference page listed in the provider sitemap"
          : plan.parser === "aircall-rendered"
            ? "parsed from every rendered endpoint-index entry in Aircall's complete Public API reference"
            : plan.parser === "paypal-specifications"
              ? "parsed from every provider-published PayPal REST OpenAPI document under openapi/ in the official specifications repository"
          : "parsed from the pinned provider machine-readable source artifact",
      coverage_confidence: plan.confidence,
      counts: restCounts,
      operation_counts: methodCounts(operations),
      operations,
      ...(plan.parser === "facebook-codegen" ? { dynamic_surface: { kind: "graph-node-edge", basis: "The provider source declares operations by owner type and edge while the concrete Graph path is selected by a runtime object ID; counts.total is the finite count of named source declarations, not an estimate of object instances." } } : {}),
      ...(sourceDocuments ? { source_documents: sourceDocuments } : {})
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
  const operations = sourceOperations(lock);
  lock.counts = { ...lock.counts, rest: operations.length, total: operations.length };
  lock.rest.counts = { total: operations.length, by_kind: { rest: operations.length }, by_method: methodCounts(operations) };
  lock.rest.coverage_confidence = { level: "complete", basis: "provider-published machine-readable OpenAPI, Swagger, Discovery, or service-model document" };
  return lock;
}

function mapSurface(connector, surface, lock) {
  const sourceURL = Object.values(lock).find((value) => value?.source_url)?.source_url || surface.docs;
  // Existing typed command/stream bindings remain executable API contracts.
  // They are retained, but never constrain discovery: every operation in the
  // complete source inventory below is added when it is absent. This prevents
  // an old api_surface from acting as the boundary of what the provider has.
  const endpoints = surface.endpoints.map((endpoint) => convertExcluded(endpoint, sourceURL));
  const index = new Map();
  for (const endpoint of endpoints) {
    const key = operationKey(connector, endpoint.method, endpoint.path);
    if (!index.has(key)) index.set(key, endpoint);
  }
  for (const source of sourceOperations(lock)) {
    const key = operationKey(connector, source.method, source.path);
    if (index.has(key)) continue;
    const endpoint = { method: source.method, path: source.path, operation: inferredOperation(source.method, source.source_url, source) };
    endpoints.push(endpoint);
    index.set(key, endpoint);
  }
  const rows = sourceOperations(lock).map((source) => {
    const key = operationKey(connector, source.method, source.path);
    const endpoint = index.get(key);
    return ledgerRow(connector, source, endpoint, `${connector}-operation-source-lock.json`);
  });
  return { surface: { ...surface, operation_ledger_version: 1, endpoints }, rows };
}

async function summary(connector, lock, rows) {
  const counts = ["direct_read", "direct_write", "etl", "reverse_etl", "binary_read", "binary_write"].map((key) => ({ key, count: rows.filter((row) => row.parity_class === key).length }));
  const enabledRows = rows.filter((row) => row.state === "enabled");
  const deletes = rows.filter((row) => row.method === "DELETE");
  const enabledDeletes = deletes.filter((row) => row.state === "enabled");
  const reverseETL = rows.filter((row) => row.declaration.reverse_etl_eligibility);
  const actionDispositions = await typedWriteActionDispositions(connector, rows);
  const declarationPendingRows = rows.filter((row) => row.state === "disabled");
  const declarationPendingByID = [...declarationPendingRows.reduce((counts, row) => {
    const id = row.foundation.declaration_pending?.id;
    if (id) counts.set(id, (counts.get(id) || 0) + 1);
    return counts;
  }, new Map()).entries()].map(([id, count]) => ({ id, count }));
  return {
    api_surface_rows: rows.length,
    exact_source_rows: rows.length,
    operations_found: rows.length,
    coverage_confidence: lock.rest.coverage_confidence,
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
    gap_ids: ["generic-typed-destination-app-dispatch", "declarative-typed-destination-action-multiplicity"],
    foundation_gaps: [
      { id: "generic-typed-destination-app-dispatch", count: reverseETL.length, scope: "destination_transport" },
      { id: "declarative-typed-destination-action-multiplicity", count: actionDispositions.length, scope: "destination_action_selection" }
    ],
    rejected_by_reason: [{ key: "declaration-pending", count: declarationPendingRows.length }],
    reverse_etl_eligibility: {
      state: "foundation-gap",
      typed_direct_write_operations: reverseETL.length,
      typed_write_actions: actionDispositions.length,
      foundation_gap: genericDestinationGap(),
      action_dispositions: actionDispositions
    },
    transport: transportSummary(connector),
    declaration_pending_ids: declarationPendingByID.map((entry) => entry.id),
    declaration_pending: declarationPendingByID
  };
}

async function main() {
  const requested = process.argv.slice(2);
  const known = new Set([...batch2, ...batch3]);
  for (const connector of requested) {
    if (!known.has(connector)) throw new Error(`unknown connector ${connector}`);
  }
  const targets = requested.length === 0 ? [...batch2, ...batch3] : requested;
  const reports = [];
  for (const connector of targets) {
    const surface = mergeBaselineSurface(connector, JSON.parse(await readFile(definitionPath(connector, "api_surface.json"), "utf8")));
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
      operations_found: mapped.rows.length,
      coverage_confidence: lock.rest.coverage_confidence
      },
      summary: await summary(connector, lock, mapped.rows),
      ledger_dispositions: mapped.rows
    };
    await mkdir(path.dirname(sourceFile(connector, lockName)), { recursive: true });
    await writeFile(sourceFile(connector, lockName), pretty(lock));
    await writeFile(sourceFile(connector, `${connector}-declaration-disposition.json`), pretty(disposition));
    await writeFile(definitionPath(connector, "api_surface.json"), pretty(mapped.surface));
    reports.push({ connector, ...disposition.summary });
  }
  if (requested.length === 0) {
    await writeFile(path.join(root, ".planning", "phases", "issue-4289-parity-map-batches-2-3-r1", "traces", "parity-map-summary.json"), pretty({ generated_at: generatedAt, connectors: reports }));
  }
}

await main();
