#!/usr/bin/env node
// Materializes and checks the source-backed parity-map evidence for issue #4290.
// It intentionally consumes only the existing normalized provider inventories;
// it neither invokes connector commands nor constructs provider requests.

import { createHash } from 'node:crypto';
import { mkdir, readFile, writeFile } from 'node:fs/promises';
import { join } from 'node:path';

const root = process.cwd();
const phase = '.planning/phases/issue-4290-map-parity-batches-45-r1';
const batches = {
  batch4: [
    'salesforce', 'hubspot', 'pipedrive', 'mailchimp', 'zendesk-support',
    'quickbooks', 'bamboo-hr', 'airtable', 'google-analytics-data-api', 'woocommerce',
  ],
  batch5: [
    'pinterest', 'tiktok-marketing', 'linear', 'buildkite', 'sonar-cloud',
    'launchdarkly', 'fastly', 'squarespace', 'ebay-fulfillment', 'shipstation',
  ],
};

const sourceOverrides = {
  'ebay-fulfillment': 'https://developer.ebay.com/develop/api/fulfillment-api/release-notes',
};

const browserSkips = {
  'tiktok-marketing': {
    reason: 'no-public-api-description',
    evidence: 'chrome-devtools-axi browser fetch on 2026-08-19: https://business-api.tiktok.com/portal/docs?id=1740029169927169 returned chrome-error://chromewebdata/ ERR_SSL_PROTOCOL_ERROR.',
  },
  'ebay-fulfillment': {
    reason: 'no-public-api-description',
    evidence: 'chrome-devtools-axi browser fetch on 2026-08-19: https://developer.ebay.com/develop/api/fulfillment-api/release-notes returned the official Error Page | eBay rather than API documentation.',
  },
};

const reverseETLGap = {
  id: 'generic-typed-destination-executor',
  evidence: 'internal/app/issue_label_warehouse_transport.go:85-95: the only destination DefinitionFactory builds and enforces the GitHub issue-label destination contract.',
  minimal_change: 'register a connector-neutral typed destination DefinitionFactory selected by the definition, with per-connector evidence, explicit source bindings, acknowledgement and per-mode apply strategies',
};

function hash(bytes) {
  return createHash('sha256').update(bytes).digest('hex');
}

function json(value) {
  return `${JSON.stringify(value, null, 2)}\n`;
}

function slug(value) {
  return String(value).toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '') || 'root';
}

function endpointOperationID(endpoint, index) {
  const notes = endpoint.operation?.notes || '';
  const match = notes.match(/(?:^|;)\s*operation_id=([^;]+)/);
  return match ? match[1].trim() : `${endpoint.method.toLowerCase()}-${slug(endpoint.path)}-${index + 1}`;
}

function operationSourceURL(endpoint, sourceURL) {
  return endpoint.operation?.source_url || sourceURL;
}

function exactCommand(cli, endpoint) {
  return (cli?.commands || []).find((command) =>
    (command.api_surface || []).some((binding) => binding.method === endpoint.method && binding.path === endpoint.path),
  ) || null;
}

function parityClass(endpoint, command) {
  if (endpoint.covered_by?.write) return 'direct_write';
  if (endpoint.covered_by?.stream) return 'etl';
  // Reverse ETL is a transport eligibility attribute of a typed direct write,
  // never an endpoint class. A legacy/planned reverse_etl command therefore
  // still describes a direct_write endpoint.
  if (command?.intent === 'reverse_etl') return 'direct_write';
  if (command?.intent && ['direct_read', 'direct_write', 'etl', 'binary_read', 'binary_write'].includes(command.intent)) return command.intent;
  if (endpoint.excluded?.category === 'binary_payload') return ['GET', 'HEAD'].includes(endpoint.method) ? 'binary_read' : 'binary_write';
  return ['GET', 'HEAD', 'OPTIONS'].includes(endpoint.method) ? 'direct_read' : 'direct_write';
}

function classifiedDisposition(connector, endpoint, index, cli, hasTransport, sourceURL) {
  const command = exactCommand(cli, endpoint);
  const parity = parityClass(endpoint, command);
  const operationID = endpointOperationID(endpoint, index);
  const isElevated = endpoint.excluded?.category === 'requires_elevated_scope';
  const isNonData = ['non_data_endpoint', 'deprecated'].includes(endpoint.excluded?.category) || endpoint.operation?.model === 'deprecated';
  const source = {
    source_lock: `sources/${connector}-operation-source-lock.json`,
    source_id: `${connector}.api_surface.${operationID}`,
    source_url: operationSourceURL(endpoint, sourceURL),
    source_location: `api_surface.json:endpoints[${index}]`,
    operation_id: operationID,
    deprecated: endpoint.excluded?.category === 'deprecated' || endpoint.operation?.model === 'deprecated',
  };
  const apiSurface = {
    method: endpoint.method,
    path: endpoint.path,
    operation: endpoint.operation || null,
    covered_by: endpoint.covered_by || null,
    excluded: endpoint.excluded || null,
  };

  if (endpoint.covered_by?.write) {
    return {
      method: endpoint.method,
      path: endpoint.path,
      parity_class: parity,
      api_surface: apiSurface,
      source,
      state: 'enabled',
      foundation: { state: 'present', evidence: `internal/connectors/defs/${connector}/writes.json: typed direct-write action "${endpoint.covered_by.write}" binds ${endpoint.method} ${endpoint.path}.` },
      rejection: null,
      declaration: {
        status: `enabled; typed direct-write action "${endpoint.covered_by.write}" binds the normalized source endpoint`,
        command: command ? { path: command.path, intent: command.intent, availability: command.availability } : null,
        reverse_etl_eligibility: { eligible: true, state: 'foundation-gap', foundation_gap: reverseETLGap },
      },
    };
  }

  if (isElevated) {
    return {
      method: endpoint.method,
      path: endpoint.path,
      parity_class: parity,
      api_surface: apiSurface,
      source,
      state: 'enabled',
      foundation: { state: 'present', evidence: 'Authorization scope is a runtime provider policy, not an engine refusal.' },
      rejection: null,
      declaration: {
        status: 'enabled; requires elevated scope at runtime',
        command: command ? { path: command.path, intent: command.intent, availability: command.availability } : null,
        runtime_metadata: { required_scope: endpoint.excluded.reason },
      },
    };
  }

  if (command?.availability === 'implemented') {
    const declaration = {
      status: `enabled; runnable ${parity} command "${command.path}" binds the normalized source endpoint`,
      command: { path: command.path, intent: command.intent, availability: command.availability },
    };
    if (parity === 'etl') {
      declaration.transport = hasTransport
        ? { source_transport: { state: 'declared', evidence: `internal/connectors/defs/${connector}/sync_transport.json` } }
        : { source_transport: { state: 'declaration-pending', evidence: `internal/connectors/defs/${connector}/sync_transport.json is absent; the merged declarative source contract is available but no connector-owned conformance claim exists.` } };
    }
    return {
      method: endpoint.method, path: endpoint.path, parity_class: parity, api_surface: apiSurface, source,
      state: 'enabled',
      foundation: { state: 'present', evidence: `internal/connectors/defs/${connector}/cli_surface.json: implemented ${parity} command binds ${endpoint.method} ${endpoint.path}.` },
      rejection: null,
      declaration,
    };
  }

  if (isNonData) {
    return {
      method: endpoint.method, path: endpoint.path, parity_class: parity, api_surface: apiSurface, source,
      state: 'disabled',
      foundation: { state: 'present', evidence: 'The engine class is available; the provider endpoint is not a data-operation declaration.' },
      rejection: {
        reason: 'provider-does-not-expose', recoverable: false,
        detail: endpoint.excluded?.reason || 'The provider marks the endpoint deprecated or non-data.',
        evidence: `api_surface.json:endpoints[${index}]`,
      },
      declaration: { status: 'disabled; provider endpoint is not an executable data-operation declaration', command: command ? { path: command.path, intent: command.intent, availability: command.availability } : null },
    };
  }

  const declaration = {
    status: `declaration-pending; no implemented connector-owned ${parity} command binds the normalized source endpoint`,
    command: command ? { path: command.path, intent: command.intent, availability: command.availability } : null,
  };
  if (parity === 'etl') {
    declaration.transport = hasTransport
      ? { source_transport: { state: 'declared', evidence: `internal/connectors/defs/${connector}/sync_transport.json` } }
      : { source_transport: { state: 'declaration-pending', evidence: `internal/connectors/defs/${connector}/sync_transport.json is absent; the source executor is connector-neutral but requires a connector-owned declaration and conformance evidence.` } };
  }
  return {
    method: endpoint.method, path: endpoint.path, parity_class: parity, api_surface: apiSurface, source,
    state: 'declaration-pending',
    foundation: { state: 'present', evidence: 'No current engine refusal is evidenced; a missing operation contract, command, or CLI surface is a declaration gap.' },
    rejection: { reason: 'declaration-pending', recoverable: true, detail: 'The normalized documented endpoint has no implemented connector-owned terminal declaration.', evidence: `api_surface.json:endpoints[${index}]` },
    declaration,
  };
}

async function fetchPublicSource(connector, sourceURL) {
  if (browserSkips[connector]) return { state: 'skipped', source_url: sourceURL, ...browserSkips[connector] };
  const response = await fetch(sourceURL, { redirect: 'follow', headers: { 'user-agent': 'Polymetrics source-lock/1.0' } });
  if (!response.ok) throw new Error(`${connector}: public source fetch ${response.status} for ${sourceURL}`);
  const bytes = Buffer.from(await response.arrayBuffer());
  return {
    state: 'fetched',
    source_url: response.url,
    requested_url: sourceURL,
    sha256: hash(bytes),
    bytes: bytes.length,
    content_type: response.headers.get('content-type') || 'unknown',
  };
}

function countBy(items, key) {
  const values = new Map();
  for (const item of items) values.set(item[key], (values.get(item[key]) || 0) + 1);
  return [...values.entries()].sort(([a], [b]) => a.localeCompare(b)).map(([key, count]) => ({ key, count }));
}

function summaryMarkdown(connector, map) {
  const summary = map.summary;
  return [
    `# ${connector} parity map`,
    '',
    `- Documented endpoints: ${summary.exact_source_rows}`,
    `- Enabled: ${summary.enabled_operations}`,
    `- Declaration pending: ${summary.declaration_pending_operations}`,
    `- Disabled: ${summary.disabled_operations}`,
    `- Documented DELETEs: ${summary.documented_deletes}; enabled DELETEs: ${summary.enabled_deletes}`,
    `- Parity classes: ${summary.parity_class_counts.map(({ key, count }) => `${key}=${count}`).join(', ')}`,
    `- Foundation gaps: ${summary.gap_ids.length ? summary.gap_ids.join(', ') : 'none'}`,
    `- Public source retrieval: ${map.source_basis.source_retrieval.state}`,
    '',
  ].join('\n');
}

async function materialize(connector) {
  const dir = join(root, 'internal/connectors/defs', connector);
  const sources = join(dir, 'sources');
  const apiSurfaceBytes = await readFile(join(dir, 'api_surface.json'));
  const apiSurface = JSON.parse(apiSurfaceBytes);
  const cli = await readJSON(join(dir, 'cli_surface.json'));
  const sourceURL = sourceOverrides[connector] || apiSurface.docs.split('; ')[0];
  const retrieval = await fetchPublicSource(connector, sourceURL);
  const hasTransport = await exists(join(dir, 'sync_transport.json'));
  const rows = apiSurface.endpoints.map((endpoint, index) => classifiedDisposition(connector, endpoint, index, cli, hasTransport, retrieval.source_url || sourceURL));
  const documentedDeletes = rows.filter(({ method }) => method === 'DELETE').length;
  const sourceLock = {
    schema_version: 1,
    connector,
    captured_at: '2026-08-19T00:00:00Z',
    source_retrieval: retrieval,
    normalized_inventory: {
      path: 'api_surface.json', sha256: hash(apiSurfaceBytes), bytes: apiSurfaceBytes.length,
      source_operation_count: rows.length,
    },
    rest: {
      source_url: retrieval.source_url || sourceURL,
      sha256: retrieval.sha256 || null,
      bytes: retrieval.bytes || null,
      operation_counts: countBy(rows, 'method'),
      operations: rows.map((row, index) => ({
        id: row.source.source_id, protocol: 'rest', method: row.method, path: row.path,
        operation_id: row.source.operation_id, deprecated: row.source.deprecated,
        source_location: row.source.source_location, api_surface_index: index,
      })),
    },
  };
  const map = {
    schema_version: 1,
    connector,
    generated_at: '2026-08-19T00:00:00Z',
    source_basis: {
      source_lock: `sources/${connector}-operation-source-lock.json`,
      source_url: retrieval.source_url || sourceURL,
      source_sha256: retrieval.sha256 || null,
      source_bytes: retrieval.bytes || null,
      source_operation_count: rows.length,
      normalized_inventory_sha256: hash(apiSurfaceBytes),
      source_retrieval: retrieval.state,
    },
    summary: {
      api_surface_rows: apiSurface.endpoints.length,
      exact_source_rows: rows.length,
      declared_operations: rows.length,
      declared_percent: 100,
      enabled_operations: rows.filter(({ state }) => state === 'enabled').length,
      declaration_pending_operations: rows.filter(({ state }) => state === 'declaration-pending').length,
      disabled_operations: rows.filter(({ state }) => state === 'disabled').length,
      documented_deletes: documentedDeletes,
      enabled_deletes: rows.filter(({ method, state }) => method === 'DELETE' && state === 'enabled').length,
      parity_class_counts: countBy(rows, 'parity_class'),
      stream_bindings: rows.filter(({ api_surface }) => api_surface.covered_by?.stream).length,
      writes_actions: rows.filter(({ api_surface }) => api_surface.covered_by?.write).length,
      terminal_commands: rows.filter(({ declaration }) => declaration.command?.availability === 'implemented').length,
      live_certification: 'pending',
      gap_ids: ['generic-typed-destination-executor'],
      foundation_gaps: [{ id: 'generic-typed-destination-executor', count: rows.filter(({ api_surface }) => api_surface.covered_by?.write).length, scope: 'shared_destination_runtime', note: 'Reverse-ETL eligibility is an attribute on typed direct writes, not an endpoint parity class.' }],
      rejected_by_reason: countBy(rows.filter(({ rejection }) => rejection).map(({ rejection }) => ({ reason: rejection.reason })), 'reason'),
      transport: {
        source_transport: hasTransport ? { state: 'declared', evidence: `internal/connectors/defs/${connector}/sync_transport.json` } : { state: 'declaration-pending', evidence: `internal/connectors/defs/${connector}/sync_transport.json is absent; the declarative source factory is connector-neutral but requires a connector-owned definition and conformance evidence.` },
        destination_transport: { state: 'gap', foundation_gap: reverseETLGap },
      },
    },
    ledger_dispositions: rows,
  };
  await mkdir(sources, { recursive: true });
  await writeFile(join(sources, `${connector}-operation-source-lock.json`), json(sourceLock));
  await writeFile(join(sources, `${connector}-declaration-disposition.json`), json(map));
  await writeFile(join(sources, `${connector}-parity-map-summary.md`), summaryMarkdown(connector, map));
}

async function readJSON(path) {
  try { return JSON.parse(await readFile(path)); } catch (error) { if (error.code === 'ENOENT') return null; throw error; }
}

async function exists(path) {
  try { await readFile(path); return true; } catch (error) { if (error.code === 'ENOENT') return false; throw error; }
}

function assert(condition, message) {
  if (!condition) throw new Error(message);
}

async function check(connector) {
  const dir = join(root, 'internal/connectors/defs', connector);
  const apiSurface = JSON.parse(await readFile(join(dir, 'api_surface.json')));
  const lock = JSON.parse(await readFile(join(dir, 'sources', `${connector}-operation-source-lock.json`)));
  const map = JSON.parse(await readFile(join(dir, 'sources', `${connector}-declaration-disposition.json`)));
  const wanted = apiSurface.endpoints.map(({ method, path }) => `${method}\u0000${path}`).sort();
  const locked = lock.rest.operations.map(({ method, path }) => `${method}\u0000${path}`).sort();
  const mapped = map.ledger_dispositions.map(({ method, path }) => `${method}\u0000${path}`).sort();
  assert(JSON.stringify(wanted) === JSON.stringify(locked), `${connector}: source lock does not exactly match api_surface`);
  assert(JSON.stringify(wanted) === JSON.stringify(mapped), `${connector}: disposition map does not exactly match api_surface`);
  assert(new Set(mapped).size === mapped.length, `${connector}: duplicate method/path disposition`);
  if (browserSkips[connector]) {
    assert(lock.source_retrieval.state === 'skipped' && lock.source_retrieval.reason === 'no-public-api-description', `${connector}: browser source skip not recorded`);
    assert(!lock.source_retrieval.sha256 && !lock.source_retrieval.bytes, `${connector}: skipped browser source must not fabricate a pin`);
  } else {
    assert(lock.source_retrieval.state === 'fetched', `${connector}: public source was not fetched`);
    assert(/^[a-f0-9]{64}$/.test(lock.source_retrieval.sha256 || ''), `${connector}: source lock is missing a SHA-256 pin`);
    assert(Number.isInteger(lock.source_retrieval.bytes) && lock.source_retrieval.bytes > 0, `${connector}: source lock is missing a positive byte count`);
  }
  for (const row of map.ledger_dispositions) {
    assert(['direct_read', 'direct_write', 'etl', 'reverse_etl', 'binary_read', 'binary_write'].includes(row.parity_class), `${connector}: invalid parity class`);
    assert(row.parity_class !== 'reverse_etl', `${connector}: reverse_etl must be an eligibility attribute, not an endpoint parity class`);
    if (row.rejection?.reason === 'foundation-gap') {
      assert(row.foundation?.foundation_gap?.evidence && row.foundation?.foundation_gap?.minimal_change, `${connector}: foundation gap lacks evidence or minimal change`);
      assert(row.foundation.foundation_gap.id === 'generic-typed-destination-executor', `${connector}: invented foundation gap ${row.foundation.foundation_gap.id}`);
    }
    if (row.declaration?.reverse_etl_eligibility) {
      const eligibility = row.declaration.reverse_etl_eligibility;
      assert(row.parity_class === 'direct_write' && row.state === 'enabled', `${connector}: typed write must be enabled direct_write`);
      assert(eligibility.eligible && eligibility.state === 'foundation-gap', `${connector}: reverse ETL eligibility must expose the foundation gap`);
      assert(eligibility.foundation_gap?.id === 'generic-typed-destination-executor' && eligibility.foundation_gap?.evidence && eligibility.foundation_gap?.minimal_change, `${connector}: malformed reverse ETL foundation evidence`);
    }
    if (row.state === 'declaration-pending') assert(row.rejection?.reason === 'declaration-pending', `${connector}: pending row has wrong reason`);
    if (row.api_surface?.excluded?.category === 'requires_elevated_scope') assert(row.state === 'enabled' && !row.rejection, `${connector}: elevated scope must stay enabled`);
  }
  assert(map.summary.documented_deletes === apiSurface.endpoints.filter(({ method }) => method === 'DELETE').length, `${connector}: DELETE count mismatch`);
}

async function main() {
  const [mode, batch] = process.argv.slice(2);
  if (!['write', 'check'].includes(mode) || !batches[batch]) throw new Error('usage: materialize-parity-maps.mjs <write|check> <batch4|batch5>');
  for (const connector of batches[batch]) {
    if (mode === 'write') await materialize(connector); else await check(connector);
    console.log(`${mode}: ${connector}`);
  }
}

await main();
