#!/usr/bin/env node

// Complete-map generator for connectors whose authoritative provider source is
// OpenAPI. It only retains an old coverage link when the refreshed method/path
// still exists and it resolves to a real stream, CLI command, or typed write.
import { createHash } from 'node:crypto';
import { execFileSync } from 'node:child_process';
import { existsSync, readFileSync, writeFileSync } from 'node:fs';

const connector = process.argv[2];
const root = new URL('../../../../', import.meta.url);
const readJSON = path => JSON.parse(readFileSync(new URL(path, root), 'utf8'));
const writeJSON = (path, value) => writeFileSync(new URL(path, root), `${JSON.stringify(value, null, 2)}\n`);
const gap = {
  id: 'generic-typed-destination-executor',
  evidence: 'internal/app/issue_label_warehouse_transport.go:85-95: the only destination factory selects issueLabelDestinationReference and BuildDestination enforces the GitHub issue-label contract.',
  minimal_change: 'register a connector-neutral typed destination DefinitionFactory selected by the definition, with per-connector evidence, explicit source bindings, acknowledgement and per-mode apply strategies',
};
const configs = {
  outreach: {
    sourceURL: 'https://api.outreach.io/api/v2/schema/openapi.json',
    artifact: 'outreach-openapi-3.0.3-2026-08-19',
    api: 'Outreach REST API v2 OpenAPI 3.0.3',
    info: 'Outreach REST API v2 (https://api.outreach.io/api/v2)',
    docs: 'https://api.outreach.io/api/v2/schema/openapi.json',
    supplements: [{
      source_url: 'https://developers.outreach.io/api/custom-objects',
      source_location: 'Custom Objects via API: generic CRUD endpoint declaration',
      operations: [
        ['GET', '/api/v2/schema'],
        ['GET', '/api/v2/customObjects/{objectName}'],
        ['GET', '/api/v2/customObjects/{objectName}/{id}'],
        ['POST', '/api/v2/customObjects/{objectName}'],
        ['PATCH', '/api/v2/customObjects/{objectName}/{id}'],
        ['DELETE', '/api/v2/customObjects/{objectName}/{id}'],
      ],
    }],
  },
  'customer-io': {
    sourceURL: 'https://docs.customer.io/files/journeys-app.json',
    artifact: 'customer-io-openapi-3.1.0-2026-08-19',
    api: 'Customer.io App API OpenAPI 3.1.0',
    info: 'Customer.io App API v1 (https://api.customer.io/v1, or https://api-eu.customer.io/v1 for the EU region)',
    docs: 'https://docs.customer.io/files/journeys-app.json',
  },
  square: {
    sourceURL: 'https://raw.githubusercontent.com/square/connect-api-specification/master/api.json',
    artifact: 'square-openapi-3.0.0-2026-08-19',
    api: 'Square API OpenAPI 3.0.0',
    info: 'Square API v2 public OpenAPI specification',
    docs: 'https://github.com/square/connect-api-specification/blob/master/api.json',
  },
};
const config = configs[connector];
if (!config) throw new Error(`unsupported connector ${connector}; add an authoritative-source config`);
const base = `internal/connectors/defs/${connector}`;
const response = await fetch(config.sourceURL);
const bytes = Buffer.from(await response.arrayBuffer());
if (!response.ok) throw new Error(`HTTP ${response.status} ${config.sourceURL}`);
const document = JSON.parse(bytes.toString('utf8'));
if (!document.paths || (!document.openapi && !document.swagger)) throw new Error(`${config.sourceURL} is not an OpenAPI document`);
const sourceHash = createHash('sha256').update(bytes).digest('hex');
const supplementArtifacts = await Promise.all((config.supplements ?? []).map(async supplement => {
  const response = await fetch(supplement.source_url);
  const body = Buffer.from(await response.arrayBuffer());
  if (!response.ok) throw new Error(`HTTP ${response.status} ${supplement.source_url}`);
  return { ...supplement, bytes: body.length, sha256: createHash('sha256').update(body).digest('hex') };
}));
const methodOrder = { delete: 0, get: 1, post: 2, put: 3, patch: 4 };
const methods = Object.keys(methodOrder);
const declaredServer = document.servers?.[0]?.url;
const serverPath = declaredServer && /^https?:\/\//.test(declaredServer) ? new URL(declaredServer).pathname.replace(/\/$/, '') : '';
const surfacePath = path => serverPath && serverPath !== '/' && !path.startsWith(serverPath + '/') ? `${serverPath}${path}` : path;
const sourceOperations = Object.entries(document.paths).flatMap(([path, item]) => methods
  .filter(method => item?.[method] && typeof item[method] === 'object')
  .map(method => ({ method: method.toUpperCase(), path: surfacePath(path), source_path: path, source_url: config.sourceURL, operation_id: item[method].operationId ?? null, deprecated: item[method].deprecated === true })))
  .concat((config.supplements ?? []).flatMap(supplement => supplement.operations.map(([method, path]) => ({ method, path, source_url: supplement.source_url, source_location: supplement.source_location, operation_id: null, deprecated: false }))))
  .sort((left, right) => methodOrder[left.method.toLowerCase()] - methodOrder[right.method.toLowerCase()] || left.path.localeCompare(right.path));
const baselineRef = process.env.SOURCE_MAP_BASELINE;
const oldSurface = baselineRef
  ? JSON.parse(execFileSync('git', ['show', `${baselineRef}:${base}/api_surface.json`], { encoding: 'utf8' }))
  : readJSON(`${base}/api_surface.json`);
const oldByKey = new Map(oldSurface.endpoints.map(endpoint => [`${endpoint.method} ${endpoint.path}`, endpoint]));
const streams = new Set(readJSON(`${base}/streams.json`).streams.map(stream => stream.name));
const writesPath = new URL(`${base}/writes.json`, root);
const writes = existsSync(writesPath) ? readJSON(`${base}/writes.json`).actions ?? [] : [];
const writesByName = new Set(writes.map(action => action.name));
const cliPath = new URL(`${base}/cli_surface.json`, root);
const commands = existsSync(cliPath) ? readJSON(`${base}/cli_surface.json`).commands ?? [] : [];
const commandsByPath = new Set(commands.map(command => command.path));
const validCoverage = endpoint => {
  const coverage = endpoint?.covered_by;
  if (!coverage) return null;
  if (coverage.stream && streams.has(coverage.stream)) return { stream: coverage.stream };
  if (coverage.write && writesByName.has(coverage.write)) return { write: coverage.write };
  if (coverage.direct_read && commandsByPath.has(coverage.direct_read)) return { direct_read: coverage.direct_read };
  return null;
};
const sluggify = value => value.replace(/[^a-zA-Z0-9]+/g, ' ').trim().split(/\s+/).map(part => part[0]?.toUpperCase() + part.slice(1)).join('');
const operationCounts = Object.fromEntries(methods.filter(method => sourceOperations.some(operation => operation.method === method.toUpperCase()))
  .map(method => [method.toUpperCase(), sourceOperations.filter(operation => operation.method === method.toUpperCase()).length]));
const restOperations = sourceOperations.map((operation, index) => ({
  id: `${connector}.rest.${operation.method.toLowerCase()}.${operation.operation_id ? sluggify(operation.operation_id) : sluggify(operation.path)}${index}`,
  protocol: 'rest', method: operation.method, path: operation.path, operation_id: operation.operation_id, deprecated: operation.deprecated,
  source_location: operation.source_location ?? `.paths[${JSON.stringify(operation.source_path)}].${operation.method.toLowerCase()}`,
  source_url: operation.source_url,
}));
const coverage = config.supplements?.length
  ? { level: 'complete_machine_readable_specification_with_rendered_dynamic_supplement', basis: `Parsed every operation in the provider-published OpenAPI ${document.openapi ?? document.swagger} document at ${config.sourceURL}, then added the provider’s documented generic custom-object routes from ${config.supplements.map(supplement => supplement.source_url).join(', ')}. Outreach documents that per-account custom-object schemas are dynamic, while these six generic route shapes are fixed.` }
  : { level: 'complete_machine_readable_specification', basis: `Parsed every operation in the provider-published OpenAPI ${document.openapi ?? document.swagger} document at ${config.sourceURL}.` };
const sourceLock = {
  schema_version: 2, connector, captured_at: '2026-08-19T00:00:00Z',
  rest: { source_url: config.sourceURL, source_kind: config.supplements?.length ? 'complete_machine_readable_specification_with_rendered_dynamic_supplement' : 'complete_machine_readable_specification', sha256: sourceHash, bytes: bytes.length, openapi: document.openapi ?? document.swagger, info_version: config.info, operation_counts: operationCounts, ...(supplementArtifacts.length ? { supplements: supplementArtifacts.map(supplement => ({ source_url: supplement.source_url, source_location: supplement.source_location, operation_count: supplement.operations.length, bytes: supplement.bytes, sha256: supplement.sha256 })) } : {}), operations: restOperations },
  counts: { rest: sourceOperations.length, graphql_query: 0, graphql_mutation: 0, total: sourceOperations.length },
  operations_found: { rest: sourceOperations.length, graphql_query: 0, graphql_mutation: 0, total: sourceOperations.length },
  coverage_confidence: coverage,
};
const isMutation = method => method !== 'GET';
const operationSpec = operation => {
  if (operation.method === 'DELETE') return { model: 'destructive_action', status: 'blocked', risk: 'high', blocked_by_default: true, reason: 'Documented provider deletion has no operation-specific plan, preview, approval, typed confirmation, and execution contract.' };
  if (isMutation(operation.method)) return { model: 'disallowed', status: 'blocked', risk: 'high', blocked_by_default: true, reason: 'Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.' };
  return { model: 'direct_read', status: 'blocked', risk: 'medium', blocked_by_default: true, reason: 'Documented provider read has no connector-owned bounded operation contract and runnable command binding.' };
};
const rows = sourceOperations.map((operation, index) => {
  const preserved = validCoverage(oldByKey.get(`${operation.method} ${operation.path}`));
  const parityClass = preserved?.stream ? 'etl' : isMutation(operation.method) ? 'direct_write' : 'direct_read';
  const endpoint = { method: operation.method, path: operation.path, provenance: { artifact: operation.source_url === config.sourceURL ? config.artifact : `${connector}-rendered-dynamic-supplement-1-2026-08-19`, source_url: operation.source_url } };
  if (preserved) endpoint.covered_by = preserved; else endpoint.operation = operationSpec(operation);
  return { endpoint, source: restOperations[index], parityClass, coverage: preserved };
});
const classCounts = { direct_read: 0, direct_write: 0, etl: 0, reverse_etl: 0, binary_read: 0, binary_write: 0 };
const ledger = rows.map((row, index) => {
  classCounts[row.parityClass] += 1;
  const enabled = Boolean(row.coverage?.write || row.coverage?.direct_read);
  const pendingID = row.coverage?.stream ? `runnable-command-binding-${connector}` : `typed-operation-contract-${connector}`;
  const evidence = `internal/connectors/defs/${connector}/api_surface.json endpoints[${index}] is bound to the pinned provider operation; internal/connectors/defs/${connector}/sources/${connector}-operation-source-lock.json: ${row.source.source_location}.`;
  const declaration = enabled
    ? { status: row.coverage.write ? 'enabled; runnable typed write action binds the pinned source contract' : 'enabled; runtime-preflight runnable command binds the pinned source contract', contract: row.coverage }
    : { status: `disabled; declaration-pending ${pendingID}`, contract: null };
  if (row.parityClass === 'direct_write') declaration.reverse_etl = { eligible_typed_action: Boolean(row.coverage?.write), state: 'foundation-gap', foundation_gap: gap };
  return {
    method: row.endpoint.method, path: row.endpoint.path, parity_class: row.parityClass,
    api_surface: { method: row.endpoint.method, path: row.endpoint.path, operation: row.endpoint.operation ?? null, covered_by: row.endpoint.covered_by ?? null },
    source: { source_lock: `sources/${connector}-operation-source-lock.json`, source_id: row.source.id, source_url: row.source.source_url, source_location: row.source.source_location, operation_id: row.source.operation_id, deprecated: row.source.deprecated },
    state: enabled ? 'enabled' : 'disabled',
    foundation: enabled ? { state: 'present', evidence: 'The connector-owned executable declaration is bound to this exact pinned source operation.' } : { state: 'present', evidence: 'No shared engine change is requested for this direct operation; the missing work is a connector-local declaration bound to the pinned provider operation.', declaration_pending: { id: pendingID, evidence, minimal_change: row.coverage?.stream ? 'Add a connector-owned runnable command bound to this pinned stream operation, or retain it as declaration-pending; no engine change is required.' : 'Derive a bounded connector-owned operation contract and command from the pinned document, or retain a disabled source disposition when the source shape is not executable.' } },
    rejection: enabled ? null : { reason: 'declaration-pending', recoverable: true, detail: 'The engine shape is already available; this source operation awaits its connector-local typed contract and/or runnable command declaration.', evidence },
    declaration,
  };
});
const enabled = ledger.filter(row => row.state === 'enabled').length;
const directWriteRows = classCounts.direct_write;
const endpointBoundWrites = ledger.filter(row => row.api_surface.covered_by?.write).length;
const endpointBoundCommands = ledger.filter(row => row.api_surface.covered_by?.direct_read).length;
const disposition = {
  schema_version: 1, connector, generated_at: '2026-08-19T00:00:00Z',
  source_basis: { source_lock: `sources/${connector}-operation-source-lock.json`, source_url: config.sourceURL, source_sha256: sourceHash, source_bytes: bytes.length, source_operation_count: sourceOperations.length, operations_found: sourceOperations.length, coverage_confidence: coverage },
  summary: {
    api_surface_rows: sourceOperations.length, exact_source_rows: sourceOperations.length, declared_operations: sourceOperations.length, operations_found: sourceOperations.length, coverage_confidence: coverage,
    enabled_operations: enabled, enabled_percent: Number((enabled * 100 / sourceOperations.length).toFixed(2)), disabled_operations: sourceOperations.length - enabled,
    documented_deletes: sourceOperations.filter(operation => operation.method === 'DELETE').length, enabled_deletes: ledger.filter(row => row.state === 'enabled' && row.method === 'DELETE').length,
    parity_class_counts: Object.entries(classCounts).map(([key, count]) => ({ key, count })), stream_bindings: classCounts.etl, writes_actions: writes.length, terminal_commands: commands.length,
    gap_ids: [gap.id], foundation_gaps: [{ id: gap.id, count: directWriteRows }], rejected_by_reason: [{ key: 'declaration-pending', count: sourceOperations.length - enabled }],
    transport: { contract: 'docs/sync-transport-definition.md', source_transport: { state: 'declaration-pending', declaration_pending: { id: `sync-transport-source-definition-${connector}`, evidence: `docs/sync-transport-definition.md:15-38 lists the declaration fields; internal/connectors/defs/${connector}/sync_transport.json is absent.`, minimal_change: 'Add the connector-owned source transport declaration and conformance evidence; no engine change is required.' } }, destination_transport: { state: 'foundation-gap', foundation_gap: gap } },
    declaration_pending_ids: [`runnable-command-binding-${connector}`, `typed-operation-contract-${connector}`, `sync-transport-source-definition-${connector}`], declaration_pending: [{ id: `runnable-command-binding-${connector}`, count: classCounts.etl }, { id: `typed-operation-contract-${connector}`, count: sourceOperations.length - enabled - classCounts.etl }],
    runnable_cli_surface_commands: commands.length, typed_write_actions: writes.length, endpoint_bound_cli_commands: endpointBoundCommands, endpoint_bound_typed_write_actions: endpointBoundWrites,
  },
  ledger_dispositions: ledger, source_only_dispositions: [],
  notes: [
    `Complete six-class source-locked map regenerated from the provider OpenAPI ${document.openapi ?? document.swagger} document; every ${sourceOperations.length} pinned operation has exactly one primary parity class.`,
    'An operation is enabled only when its exact refreshed method/path still resolves to an actual typed write action or runtime-preflight runnable command. Stream-only mappings remain ETL but declaration-pending.',
    'Direct writes remain direct_write operations. Reverse-ETL eligibility is a separate attribute and remains foundation-gapped because destination execution is GitHub issue-label bound.',
  ],
};
writeJSON(`${base}/sources/${connector}-operation-source-lock.json`, sourceLock);
writeJSON(`${base}/api_surface.json`, { api: config.api, docs: config.docs, reviewed_at: '2026-08-19', operation_ledger_version: 2, scope: `Complete provider-reference inventory generated from every operation in the cited OpenAPI ${document.openapi ?? document.swagger} document${config.supplements?.length ? ' plus its fixed generic dynamic-resource routes from the cited rendered provider reference' : ''} (${sourceOperations.length} HTTP operations). Existing executable bindings are retained only where their exact method/path still matches the refreshed provider source.`, artifacts: [{ id: config.artifact, url: config.sourceURL, retrieved_at: '2026-08-19', sha256: sourceHash }, ...supplementArtifacts.map((supplement, index) => ({ id: `${connector}-rendered-dynamic-supplement-${index + 1}-2026-08-19`, url: supplement.source_url, retrieved_at: '2026-08-19', sha256: supplement.sha256 }))], endpoints: rows.map(row => row.endpoint) });
writeJSON(`${base}/sources/${connector}-declaration-disposition.json`, disposition);
console.log(`${connector}: ${sourceOperations.length} source rows; enabled=${enabled}, writes=${endpointBoundWrites}, commands=${endpointBoundCommands}`);
