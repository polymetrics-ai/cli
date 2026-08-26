#!/usr/bin/env node

import { execFileSync } from 'node:child_process';
import { readFileSync, writeFileSync } from 'node:fs';

const connectors = process.argv.slice(2);
if (connectors.length === 0) {
  throw new Error('usage: migrate-source-lock-v3.mjs <connector> [...]');
}

function sourceLockPath(connector) {
  return `internal/connectors/defs/${connector}/sources/${connector}-operation-source-lock.json`;
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, 'utf8'));
}

function committedLock(file) {
  return JSON.parse(execFileSync('git', ['show', `HEAD:${file}`], { encoding: 'utf8' }));
}

function documentForm(version) {
  if (/^3\.(0|1)\.\d+$/.test(version ?? '')) {
    return { kind: 'openapi', openapi: version };
  }
  if (version === '2.0') {
    return { kind: 'openapi', swagger: version };
  }
  return { kind: 'rendered_reference' };
}

function renderedContentType(connector) {
  return connector === 'zoho-bigin' ? 'application/xml' : 'text/html';
}

function coverageConfidence(legacy, connector) {
  if (legacy.coverage_confidence) {
    return legacy.coverage_confidence;
  }
  return {
    level: 'documented',
    basis: `The schema-version 2 ${connector} source lock pins the public provider documentation capture and its operation inventory; it makes no completeness claim.`,
  };
}

function migrate(connector) {
  const file = sourceLockPath(connector);
  const current = readJSON(file);
  const legacy = committedLock(file);
  if (current.schema_version !== 2 || legacy.schema_version !== 2 || current.connector !== connector || legacy.connector !== connector) {
    throw new Error(`${connector}: expected matching schema-version 2 source locks`);
  }
  const legacyRest = legacy.rest;
  const currentRest = current.rest;
  const form = documentForm(legacyRest.openapi);
  const rootOrigin = new URL(legacyRest.source_url).origin;
  const operations = currentRest.operations.map((operation) => {
    const citation = operation.source_url;
    if (!citation) {
      throw new Error(`${connector}: ${operation.id} has no legacy operation citation`);
    }
    if (form.kind === 'rendered_reference' && new URL(citation).origin !== rootOrigin) {
      throw new Error(`${connector}: ${operation.id} cites ${new URL(citation).origin}, not captured origin ${rootOrigin}`);
    }
    const migrated = {
      id: operation.id,
      protocol: operation.protocol,
      method: operation.method,
      path: operation.path,
      operation_id: operation.operation_id,
      deprecated: operation.deprecated,
      source_location: operation.source_location,
    };
    if (form.kind === 'rendered_reference') {
      migrated.citation_url = citation;
    }
    return migrated;
  });
  const artifact = {
    source_url: legacyRest.source_url,
    sha256: legacyRest.sha256,
    bytes: legacyRest.bytes,
    ...(form.openapi ? { openapi: form.openapi } : {}),
    ...(form.swagger ? { swagger: form.swagger } : {}),
  };
  const document = {
    id: 'primary',
    kind: form.kind,
    ...(form.kind === 'rendered_reference' ? { content_type: renderedContentType(connector) } : {}),
    artifact,
    published_source: {
      source_url: legacyRest.source_url,
      capture_url: legacyRest.source_url,
      sha256: legacyRest.sha256,
      bytes: legacyRest.bytes,
      adapter: 'schema-v2-source-lock-migration',
    },
    ...(legacyRest.info_version ? { info_version: legacyRest.info_version } : {}),
    operations,
  };
  const rest = {
    retrieval: `Migrated schema-version 2 ${legacyRest.source_kind ?? 'provider-documentation'} source-lock capture.`,
    openapi: form.openapi ? [form.openapi] : [],
    ...(form.kind === 'rendered_reference' || legacy.coverage_confidence ? { coverage_confidence: coverageConfidence(legacy, connector) } : {}),
    source_documents: [document],
  };
  const migrated = {
    schema_version: 3,
    connector,
    ...(current.captured_at ? { captured_at: current.captured_at } : {}),
    rest,
    counts: current.counts,
  };
  writeFileSync(file, `${JSON.stringify(migrated, null, 2)}\n`);
}

for (const connector of connectors) {
  migrate(connector);
}
