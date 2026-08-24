#!/usr/bin/env node
// Converts only the issue #4292 batch 8-10 pre-foundation source-lock shape to
// main's authoritative sourceImportLockV3 wire contract. It never fetches a
// provider URL or infers an OpenAPI/Swagger form pin: missing pins remain
// deliberately absent so connectorgen reports that terminal mapping gap.
import fs from "node:fs";
import path from "node:path";

const connectors = [
  "brex", "zoho-books", "testrail", "amplitude", "posthog", "metabase", "dbt", "looker", "mode", "dremio",
  "coda", "clickup-api", "calendly", "greenhouse", "lever-hiring", "ashby", "workable", "recruitee", "hibob", "factorial",
  "datadog", "pagerduty", "auth0", "okta", "firehydrant", "adobe-commerce-magento", "commercetools", "recharge", "docuseal", "eventbrite",
];

const root = path.resolve(import.meta.dirname, "../../../..");

function lockPath(connector) {
  return path.join(root, "internal", "connectors", "defs", connector, "sources", `${connector}-operation-source-lock.json`);
}

function normalizedMediaType(contentType) {
  if (typeof contentType !== "string" || contentType.trim() === "") {
    throw new Error("document has no content_type");
  }
  return contentType.split(";", 1)[0].trim().toLowerCase();
}

function captureURL(sourceURL) {
  const url = new URL(sourceURL);
  url.search = "";
  url.hash = "";
  return url.toString();
}

function sourceKind(lock, document) {
  if (lock.state === "skipped") return "unavailable";
  if (lock.rest.representation === "complete-rendered-reference") return "rendered_reference";
  if (lock.rest.representation === "machine-readable-spec" && normalizedMediaType(document.content_type) === "application/zip") return "bundle";
  if (lock.rest.representation === "machine-readable-spec") return "openapi";
  throw new Error(`unsupported historical representation ${JSON.stringify(lock.rest.representation)}`);
}

function retrieval(lock, documents) {
  const methods = new Set(documents.map(document => document.retrieval_method).filter(Boolean));
  if (methods.size === 0 && lock.skip?.retrieval_method) methods.add(lock.skip.retrieval_method);
  if (methods.size === 0) throw new Error("lock has no historical retrieval method");
  return `preserved issue-4292 public capture via ${[...methods].sort().join(", ")}`;
}

function sourceDocument(document, index, kind, operations) {
  const parsedURL = new URL(document.source_url);
  const artifact = {
    source_url: document.source_url,
    sha256: document.sha256,
    bytes: document.bytes,
  };
  if (parsedURL.search) artifact.identity_query = true;
  const target = {
    id: `document-${String(index + 1).padStart(4, "0")}`,
    ...(kind === "openapi" ? {} : { kind }),
    ...(kind === "openapi" ? {} : { content_type: normalizedMediaType(document.content_type) }),
    artifact,
    published_source: {
      source_url: document.source_url,
      capture_url: captureURL(document.source_url),
      sha256: document.sha256,
      bytes: document.bytes,
      adapter: `issue-4292-${document.retrieval_method}-capture`,
    },
    operations,
  };
  return target;
}

function migrateAvailable(lock) {
  const documents = lock.rest.documents;
  if (!Array.isArray(documents) || documents.length === 0) throw new Error("available lock has no documents");
  const documentIndex = new Map();
  documents.forEach((document, index) => {
    if (!document?.source_url || documentIndex.has(document.source_url)) throw new Error(`invalid or duplicate document source_url ${JSON.stringify(document?.source_url)}`);
    documentIndex.set(document.source_url, index);
  });
  const operationsByDocument = documents.map(() => []);
  for (const operation of lock.rest.operations ?? []) {
    const index = documentIndex.get(operation.source_url);
    if (index === undefined) throw new Error(`operation ${operation.id} has no matching captured document`);
    const kind = sourceKind(lock, documents[index]);
    const target = {
      id: operation.id,
      protocol: operation.protocol,
      method: operation.method,
      path: operation.path,
      operation_id: operation.operation_id,
      deprecated: Boolean(operation.deprecated),
      source_location: operation.source_location,
    };
    if (kind === "rendered_reference") target.citation_url = operation.source_url;
    operationsByDocument[index].push(target);
  }
  const sourceDocuments = documents.map((document, index) => sourceDocument(document, index, sourceKind(lock, document), operationsByDocument[index]));
  return sourceDocuments;
}

function migrateUnavailable(lock) {
  if (lock.state !== "skipped" || !lock.skip?.detail) throw new Error("unavailable lock lacks preserved skip provenance");
  return [{
    id: "unavailable",
    kind: "unavailable",
    unavailable_reason: lock.skip.detail,
    operations: [],
  }];
}

for (const connector of connectors) {
  const filename = lockPath(connector);
  const lock = JSON.parse(fs.readFileSync(filename, "utf8"));
  if (lock.schema_version !== 3 || !lock.rest?.documents || lock.rest.source_documents) {
    throw new Error(`${connector}: expected the preserved pre-foundation schema-version collision`);
  }
  const sourceDocuments = lock.state === "skipped" ? migrateUnavailable(lock) : migrateAvailable(lock);
  const restCount = sourceDocuments.reduce((count, document) => count + document.operations.length, 0);
  const historicalCount = lock.rest.counts?.total;
  if (historicalCount !== null && historicalCount !== undefined && historicalCount !== restCount) {
    throw new Error(`${connector}: historical count ${historicalCount} does not match ${restCount} operations`);
  }
  const migrated = {
    schema_version: 3,
    connector: lock.connector,
    ...(lock.captured_at ? { captured_at: lock.captured_at } : {}),
    rest: {
      retrieval: retrieval(lock, lock.rest.documents),
      openapi: [],
      coverage_confidence: lock.rest.coverage_confidence,
      source_documents: sourceDocuments,
    },
    counts: { rest: restCount, graphql_query: 0, graphql_mutation: 0, total: restCount },
  };
  fs.writeFileSync(filename, `${JSON.stringify(migrated, null, 2)}\n`);
  process.stdout.write(`${connector}: ${sourceDocuments.length} source document(s), ${restCount} operation(s)\n`);
}
