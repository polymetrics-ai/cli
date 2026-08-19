#!/usr/bin/env node

// Promotes Square only after square-reference-crawl.mjs has persisted all
// root-discovered groups. This guard prevents a plausible partial crawl from
// becoming the source-lock denominator.
import { createHash } from 'node:crypto';
import { readFileSync, writeFileSync } from 'node:fs';

const root = new URL('../../../../', import.meta.url);
const readJSON = path => JSON.parse(readFileSync(new URL(path, root), 'utf8'));
const writeJSON = (path, value) => writeFileSync(new URL(path, root), `${JSON.stringify(value, null, 2)}\n`);
const connector = 'square';
const referenceRoot = 'https://developer.squareup.com/reference/square';
const artifact = 'square-rendered-reference-2026-08-19';
const gap = {
  id: 'generic-typed-destination-executor',
  evidence: 'internal/app/issue_label_warehouse_transport.go:85-95: the only destination factory selects issueLabelDestinationReference and BuildDestination enforces the GitHub issue-label contract.',
  minimal_change: 'register a connector-neutral typed destination DefinitionFactory selected by the definition, with per-connector evidence, explicit source bindings, acknowledgement and per-mode apply strategies',
};
const crawl = readJSON('.planning/phases/issue-4291-connector-parity-batches-6-7-r1/crawl-progress.json').crawls.square;
if (crawl.state !== 'complete' || crawl.coverage_confidence !== 'complete_rendered_reference' || crawl.groups_retrieved !== crawl.groups_total) {
  throw new Error(`Square crawl is incomplete (${crawl.groups_retrieved ?? 0}/${crawl.groups_total ?? '?'}); refusing source-lock promotion`);
}
const sourceOperations = [...crawl.completed_operations].sort((left, right) => left.method.localeCompare(right.method) || left.path.localeCompare(right.path));
const streams = new Set(readJSON(`internal/connectors/defs/${connector}/streams.json`).streams.map(stream => stream.name));
const oldSurface = readJSON(`internal/connectors/defs/${connector}/api_surface.json`);
const oldByKey = new Map(oldSurface.endpoints.map(endpoint => [`${endpoint.method} ${endpoint.path}`, endpoint]));
const methodOrder = { DELETE: 0, GET: 1, POST: 2, PUT: 3, PATCH: 4 };
sourceOperations.sort((left, right) => methodOrder[left.method] - methodOrder[right.method] || left.path.localeCompare(right.path));
const slug = value => value.replace(/[^a-zA-Z0-9]+/g, ' ').trim().split(/\s+/).map(part => part[0]?.toUpperCase() + part.slice(1)).join('');
const operationCounts = Object.fromEntries(['GET', 'POST', 'PUT', 'PATCH', 'DELETE'].filter(method => sourceOperations.some(operation => operation.method === method)).map(method => [method, sourceOperations.filter(operation => operation.method === method).length]));
const manifest = Object.entries(crawl.group_pages).sort(([left], [right]) => left.localeCompare(right)).map(([url, page]) => ({ url, sha256: page.sha256, bytes: page.bytes }));
const manifestSHA = createHash('sha256').update(JSON.stringify(manifest)).digest('hex');
const totalBytes = manifest.reduce((sum, page) => sum + page.bytes, 0);
const sourceLockOperations = sourceOperations.map((operation, index) => ({
  id: `${connector}.rest.${operation.method.toLowerCase()}.${slug(operation.path)}${index}`,
  protocol: 'rest', method: operation.method, path: operation.path, operation_id: null, deprecated: false,
  source_location: 'rendered API reference operation card', source_url: operation.source_url,
}));
const coverage = { level: 'complete_rendered_reference', basis: `Crawled and persisted every ${crawl.groups_total} API group linked from ${referenceRoot}; the page-hash manifest is pinned in the crawl progress and its deterministic SHA-256 is ${manifestSHA}. The group pages yield ${sourceOperations.length} unique HTTP method/path operations.` };
const sourceLock = {
  schema_version: 2, connector, captured_at: '2026-08-19T00:00:00Z',
  rest: { source_url: referenceRoot, source_kind: 'complete_rendered_reference', source_document_count: crawl.groups_total, reference_manifest_sha256: manifestSHA, bytes: totalBytes, openapi: 'rendered-reference', info_version: 'Square API public reference', operation_counts: operationCounts, operations: sourceLockOperations },
  counts: { rest: sourceOperations.length, graphql_query: 0, graphql_mutation: 0, total: sourceOperations.length },
  operations_found: { rest: sourceOperations.length, graphql_query: 0, graphql_mutation: 0, total: sourceOperations.length }, coverage_confidence: coverage,
};
const isMutation = method => method !== 'GET';
const operationSpec = operation => operation.method === 'DELETE'
  ? { model: 'destructive_action', status: 'blocked', risk: 'high', blocked_by_default: true, reason: 'Documented provider deletion has no operation-specific plan, preview, approval, typed confirmation, and execution contract.' }
  : isMutation(operation.method)
    ? { model: 'disallowed', status: 'blocked', risk: 'high', blocked_by_default: true, reason: 'Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.' }
    : { model: 'direct_read', status: 'blocked', risk: 'medium', blocked_by_default: true, reason: 'Documented provider read has no connector-owned bounded operation contract and runnable command binding.' };
const classCounts = { direct_read: 0, direct_write: 0, etl: 0, reverse_etl: 0, binary_read: 0, binary_write: 0 };
const endpoints = sourceOperations.map(operation => {
  const previous = oldByKey.get(`${operation.method} ${operation.path}`);
  const stream = previous?.covered_by?.stream && streams.has(previous.covered_by.stream) ? previous.covered_by.stream : null;
  const parityClass = stream ? 'etl' : isMutation(operation.method) ? 'direct_write' : 'direct_read';
  classCounts[parityClass] += 1;
  return { method: operation.method, path: operation.path, provenance: { artifact, source_url: operation.source_url }, ...(stream ? { covered_by: { stream } } : { operation: operationSpec(operation) }), parityClass };
});
const ledger = endpoints.map((endpoint, index) => {
  const source = sourceLockOperations[index];
  const stream = endpoint.covered_by?.stream;
  const pendingID = stream ? `runnable-command-binding-${connector}` : `typed-operation-contract-${connector}`;
  const evidence = `internal/connectors/defs/${connector}/api_surface.json endpoints[${index}] is bound to the pinned provider operation; internal/connectors/defs/${connector}/sources/${connector}-operation-source-lock.json: ${source.source_location} at ${source.source_url}.`;
  const declaration = { status: `disabled; declaration-pending ${pendingID}`, contract: null };
  if (endpoint.parityClass === 'direct_write') declaration.reverse_etl = { eligible_typed_action: false, state: 'foundation-gap', foundation_gap: gap };
  return { method: endpoint.method, path: endpoint.path, parity_class: endpoint.parityClass, api_surface: { method: endpoint.method, path: endpoint.path, operation: endpoint.operation ?? null, covered_by: endpoint.covered_by ?? null }, source: { source_lock: `sources/${connector}-operation-source-lock.json`, source_id: source.id, source_url: source.source_url, source_location: source.source_location, operation_id: null, deprecated: false }, state: 'disabled', foundation: { state: 'present', evidence: 'No shared engine change is requested for this direct operation; the missing work is a connector-local declaration bound to the pinned provider operation.', declaration_pending: { id: pendingID, evidence, minimal_change: stream ? 'Add a connector-owned runnable command bound to this pinned stream operation, or retain it as declaration-pending; no engine change is required.' : 'Derive a bounded connector-owned operation contract and command from the pinned document, or retain a disabled source disposition when the source shape is not executable.' } }, rejection: { reason: 'declaration-pending', recoverable: true, detail: 'The engine shape is already available; this source operation awaits its connector-local typed contract and/or runnable command declaration.', evidence }, declaration };
});
const directWrites = classCounts.direct_write;
const disposition = {
  schema_version: 1, connector, generated_at: '2026-08-19T00:00:00Z', source_basis: { source_lock: `sources/${connector}-operation-source-lock.json`, source_url: referenceRoot, source_sha256: manifestSHA, source_bytes: totalBytes, source_operation_count: sourceOperations.length, operations_found: sourceOperations.length, coverage_confidence: coverage },
  summary: { api_surface_rows: sourceOperations.length, exact_source_rows: sourceOperations.length, declared_operations: sourceOperations.length, operations_found: sourceOperations.length, coverage_confidence: coverage, enabled_operations: 0, enabled_percent: 0, disabled_operations: sourceOperations.length, documented_deletes: operationCounts.DELETE ?? 0, enabled_deletes: 0, parity_class_counts: Object.entries(classCounts).map(([key, count]) => ({ key, count })), stream_bindings: classCounts.etl, writes_actions: 0, terminal_commands: 0, live_certification: 'pending', gap_ids: [gap.id], foundation_gaps: [{ id: gap.id, count: directWrites }], rejected_by_reason: [{ key: 'declaration-pending', count: sourceOperations.length }], transport: { contract: 'docs/sync-transport-definition.md', source_transport: { state: 'declaration-pending', declaration_pending: { id: `sync-transport-source-definition-${connector}`, evidence: `docs/sync-transport-definition.md:15-38 lists the declaration fields; internal/connectors/defs/${connector}/sync_transport.json is absent.`, minimal_change: 'Add the connector-owned source transport declaration and conformance evidence; no engine change is required.' } }, destination_transport: { state: 'foundation-gap', foundation_gap: gap } }, declaration_pending_ids: [`runnable-command-binding-${connector}`, `typed-operation-contract-${connector}`, `sync-transport-source-definition-${connector}`], declaration_pending: [{ id: `runnable-command-binding-${connector}`, count: classCounts.etl }, { id: `typed-operation-contract-${connector}`, count: sourceOperations.length - classCounts.etl }], runnable_cli_surface_commands: 0, typed_write_actions: 0, endpoint_bound_cli_commands: 0, endpoint_bound_typed_write_actions: 0 },
  ledger_dispositions: ledger, source_only_dispositions: [], notes: [`Complete six-class source-locked map regenerated from all ${crawl.groups_total} Square rendered-reference groups; every ${sourceOperations.length} pinned operation has exactly one primary parity class.`, 'No partial group count was promoted: the source lock is generated only after all root-discovered groups are durably persisted.', 'Direct writes remain direct_write operations. Reverse-ETL eligibility is a separate attribute and remains foundation-gapped because destination execution is GitHub issue-label bound.'],
};
writeJSON(`internal/connectors/defs/${connector}/sources/${connector}-operation-source-lock.json`, sourceLock);
writeJSON(`internal/connectors/defs/${connector}/api_surface.json`, { api: 'Square API public reference', docs: referenceRoot, reviewed_at: '2026-08-19', operation_ledger_version: 2, scope: `Complete provider-reference inventory generated from every operation in all ${crawl.groups_total} cited rendered API-reference groups (${sourceOperations.length} unique HTTP operations). Existing stream bindings are retained only where their exact method/path remains in the complete crawl.`, artifacts: [{ id: artifact, url: referenceRoot, retrieved_at: '2026-08-19', sha256: manifestSHA }], endpoints: endpoints.map(({ parityClass, ...endpoint }) => endpoint) });
writeJSON(`internal/connectors/defs/${connector}/sources/${connector}-declaration-disposition.json`, disposition);
console.log(`${connector}: ${sourceOperations.length} source rows, ${classCounts.etl} ETL bindings, ${directWrites} direct writes`);
