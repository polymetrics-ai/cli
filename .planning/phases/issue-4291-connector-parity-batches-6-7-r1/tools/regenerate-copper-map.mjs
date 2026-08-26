#!/usr/bin/env node

// Regenerates the complete Copper source lock, API surface, and six-class
// disposition map from the persisted complete rendered-reference crawl.
import { readFileSync, writeFileSync } from 'node:fs';

const root = new URL('../../../../', import.meta.url);
const readJSON = path => JSON.parse(readFileSync(new URL(path, root), 'utf8'));
const writeJSON = (path, value) => writeFileSync(new URL(path, root), `${JSON.stringify(value, null, 2)}\n`);
const connector = 'copper';
const sourceURL = 'https://developer.copper.com/search/search_index.json';
const referenceRoot = 'https://developer.copper.com/';
const artifact = 'copper-rendered-reference-2026-08-19';
const gap = {
  id: 'generic-typed-destination-executor',
  evidence: 'internal/app/issue_label_warehouse_transport.go:85-95: the only destination factory selects issueLabelDestinationReference and BuildDestination enforces the GitHub issue-label contract.',
  minimal_change: 'register a connector-neutral typed destination DefinitionFactory selected by the definition, with per-connector evidence, explicit source bindings, acknowledgement and per-mode apply strategies',
};
const progress = readJSON('.planning/phases/issue-4291-connector-parity-batches-6-7-r1/crawl-progress.json').crawls.copper;
if (progress.state !== 'complete' || progress.coverage_confidence !== 'complete_rendered_reference') {
  throw new Error('Copper crawl is not complete; refusing to promote a partial source lock');
}

const streamBindings = new Map([
  ['POST /v1/people/search', 'people'],
  ['POST /v1/companies/search', 'companies'],
  ['POST /v1/opportunities/search', 'opportunities'],
  ['POST /v1/leads/search', 'leads'],
  ['POST /v1/tasks/search', 'tasks'],
]);
const methodOrder = { DELETE: 0, GET: 1, POST: 2, PUT: 3, PATCH: 4 };
const slug = value => value.replace(/[^a-zA-Z0-9]+/g, ' ').trim().split(/\s+/).map(piece => piece[0]?.toUpperCase() + piece.slice(1)).join('');
const fixProviderTypo = path => path === '/v1/webhooks{webhook_id}' ? '/v1/webhooks/{webhook_id}' : path;
const sourceOperations = progress.completed_operations
  .map(operation => ({ ...operation, raw_path: operation.path, path: fixProviderTypo(operation.path) }))
  .sort((left, right) => methodOrder[left.method] - methodOrder[right.method] || left.path.localeCompare(right.path));
const sourceCounts = Object.fromEntries(['GET', 'POST', 'PUT', 'PATCH', 'DELETE']
  .filter(method => sourceOperations.some(operation => operation.method === method))
  .map(method => [method, sourceOperations.filter(operation => operation.method === method).length]));

const sourceLockOperations = sourceOperations.map((operation, index) => ({
  id: `${connector}.rest.${operation.method.toLowerCase()}.${slug(operation.path)}${index}`,
  protocol: 'rest',
  method: operation.method,
  path: operation.path,
  operation_id: null,
  deprecated: false,
  source_location: operation.raw_path === operation.path
    ? `Copper rendered declaration: ${operation.title}`
    : `Copper rendered declaration: ${operation.title}; normalized provider markup ${operation.raw_path} to ${operation.path}`,
  source_url: operation.source_url,
}));
const coverage = {
  level: 'complete_rendered_reference',
  basis: 'Parsed all 637 provider-published MkDocs search-index documentation nodes. The declaration-form operation blocks yield 89 unique HTTP method/path operations; curl examples are intentionally excluded because they substitute literal request ids.',
};
const sourceLock = {
  schema_version: 2,
  connector,
  captured_at: '2026-08-19T00:00:00Z',
  rest: {
    source_url: referenceRoot,
    index_url: sourceURL,
    source_kind: 'complete_rendered_reference',
    source_document_count: progress.documents_total,
    sha256: progress.index_sha256,
    bytes: progress.index_bytes,
    openapi: 'rendered-reference',
    info_version: 'Copper Developer API public reference',
    operation_counts: sourceCounts,
    operations: sourceLockOperations,
  },
  counts: { rest: sourceOperations.length, graphql_query: 0, graphql_mutation: 0, total: sourceOperations.length },
  operations_found: { rest: sourceOperations.length, graphql_query: 0, graphql_mutation: 0, total: sourceOperations.length },
  coverage_confidence: coverage,
};

const writeMethods = new Set(['POST', 'PUT', 'PATCH', 'DELETE']);
const operationSpec = operation => {
  if (operation.method === 'DELETE') return { model: 'destructive_action', status: 'blocked', risk: 'high', blocked_by_default: true, reason: 'Documented provider deletion has no operation-specific plan, preview, approval, typed confirmation, and execution contract.' };
  if (writeMethods.has(operation.method)) return { model: 'disallowed', status: 'blocked', risk: 'high', blocked_by_default: true, reason: 'Documented provider mutation has no reviewed operation-specific body, risk, approval, and execution contract.' };
  return { model: 'direct_read', status: 'blocked', risk: 'medium', blocked_by_default: true, reason: 'Documented provider read has no connector-owned bounded operation contract and runnable command binding.' };
};
const endpointRows = sourceOperations.map(operation => {
  const stream = streamBindings.get(`${operation.method} ${operation.path}`);
  return {
    method: operation.method,
    path: operation.path,
    provenance: { artifact, source_url: operation.source_url },
    ...(stream ? { covered_by: { stream } } : { operation: operationSpec(operation) }),
  };
});
const apiSurface = {
  api: 'Copper Developer API public reference',
  docs: referenceRoot,
  reviewed_at: '2026-08-19',
  operation_ledger_version: 2,
  scope: 'Complete provider-reference inventory generated from every declaration-form endpoint in the cited 637-document rendered reference (89 unique HTTP operations). The five native ETL stream bindings are derived from internal/connectors/native/copper/streams.go:5-25, which routes each named resource through POST /<resource>/search.',
  artifacts: [{ id: artifact, url: sourceURL, retrieved_at: '2026-08-19', sha256: progress.index_sha256 }],
  endpoints: endpointRows,
};

const classCounts = { direct_read: 0, direct_write: 0, etl: 0, reverse_etl: 0, binary_read: 0, binary_write: 0 };
const ledger = endpointRows.map((endpoint, index) => {
  const stream = endpoint.covered_by?.stream;
  const parityClass = stream ? 'etl' : writeMethods.has(endpoint.method) ? 'direct_write' : 'direct_read';
  classCounts[parityClass] += 1;
  const pendingID = stream ? `runnable-command-binding-${connector}` : `typed-operation-contract-${connector}`;
  const source = sourceLockOperations[index];
  const evidence = `internal/connectors/defs/${connector}/api_surface.json endpoints[${index}] is bound to the pinned provider operation; internal/connectors/defs/${connector}/sources/${connector}-operation-source-lock.json: ${source.source_location}.`;
  const declaration = { status: `disabled; declaration-pending ${pendingID}`, contract: null };
  if (parityClass === 'direct_write') declaration.reverse_etl = { eligible_typed_action: false, state: 'foundation-gap', foundation_gap: gap };
  return {
    method: endpoint.method,
    path: endpoint.path,
    parity_class: parityClass,
    api_surface: { method: endpoint.method, path: endpoint.path, operation: endpoint.operation ?? null, covered_by: endpoint.covered_by ?? null },
    source: { source_lock: `sources/${connector}-operation-source-lock.json`, source_id: source.id, source_url: source.source_url, source_location: source.source_location, operation_id: null, deprecated: false },
    state: 'disabled',
    foundation: {
      state: 'present',
      evidence: 'No shared engine change is requested for this direct operation; the missing work is a connector-local declaration bound to the pinned provider operation.',
      declaration_pending: {
        id: pendingID,
        evidence,
        minimal_change: stream
          ? 'Add a connector-owned runnable command bound to this pinned stream operation, or retain it as declaration-pending; no engine change is required.'
          : 'Derive a bounded connector-owned operation contract and command from the pinned document, or retain a disabled source disposition when the source shape is not executable.',
      },
    },
    rejection: { reason: 'declaration-pending', recoverable: true, detail: 'The engine shape is already available; this source operation awaits its connector-local typed contract and/or runnable command declaration.', evidence },
    declaration,
  };
});
const directWrites = classCounts.direct_write;
const disposition = {
  schema_version: 1,
  connector,
  generated_at: '2026-08-19T00:00:00Z',
  source_basis: { source_lock: `sources/${connector}-operation-source-lock.json`, source_url: referenceRoot, source_sha256: progress.index_sha256, source_bytes: progress.index_bytes, source_operation_count: sourceOperations.length, operations_found: sourceOperations.length, coverage_confidence: coverage },
  summary: {
    api_surface_rows: sourceOperations.length,
    exact_source_rows: sourceOperations.length,
    declared_operations: sourceOperations.length,
    operations_found: sourceOperations.length,
    coverage_confidence: coverage,
    enabled_operations: 0,
    enabled_percent: 0,
    disabled_operations: sourceOperations.length,
    documented_deletes: sourceOperations.filter(operation => operation.method === 'DELETE').length,
    enabled_deletes: 0,
    parity_class_counts: Object.entries(classCounts).map(([key, count]) => ({ key, count })),
    stream_bindings: classCounts.etl,
    writes_actions: 0,
    terminal_commands: 0,
    live_certification: 'pending',
    gap_ids: [gap.id],
    foundation_gaps: [{ id: gap.id, count: directWrites }],
    rejected_by_reason: [{ key: 'declaration-pending', count: sourceOperations.length }],
    transport: {
      contract: 'docs/sync-transport-definition.md',
      source_transport: { state: 'declaration-pending', declaration_pending: { id: `sync-transport-source-definition-${connector}`, evidence: `docs/sync-transport-definition.md:15-38 lists the declaration fields; internal/connectors/defs/${connector}/sync_transport.json is absent.`, minimal_change: 'Add the connector-owned source transport declaration and conformance evidence; no engine change is required.' } },
      destination_transport: { state: 'foundation-gap', foundation_gap: gap },
    },
    declaration_pending_ids: [`runnable-command-binding-${connector}`, `typed-operation-contract-${connector}`, `sync-transport-source-definition-${connector}`],
    declaration_pending: [{ id: `runnable-command-binding-${connector}`, count: classCounts.etl }, { id: `typed-operation-contract-${connector}`, count: sourceOperations.length - classCounts.etl }],
    runnable_cli_surface_commands: 0,
    typed_write_actions: 0,
    endpoint_bound_cli_commands: 0,
    endpoint_bound_typed_write_actions: 0,
  },
  ledger_dispositions: ledger,
  source_only_dispositions: [],
  notes: [
    `Complete six-class source-locked map regenerated from all ${progress.documents_total} Copper rendered-reference documents; every ${sourceOperations.length} pinned operation has exactly one primary parity class.`,
    'The previous five synthetic HOOK rows were discarded. The native stream table proves the five POST /<resource>/search ETL bindings; all other documented operations are source-only until a connector-owned declaration exists.',
    'Direct writes remain direct_write operations. Reverse-ETL eligibility is a separate attribute and remains foundation-gapped because destination execution is GitHub issue-label bound.',
    'The provider’s View subscription by ID declaration omits a path separator in `/v1/webhooks{webhook_id}`; the map records the mechanical `/v1/webhooks/{webhook_id}` normalization and retains the endpoint-page citation.',
  ],
};

writeJSON(`internal/connectors/defs/${connector}/sources/${connector}-operation-source-lock.json`, sourceLock);
writeJSON(`internal/connectors/defs/${connector}/api_surface.json`, apiSurface);
writeJSON(`internal/connectors/defs/${connector}/sources/${connector}-declaration-disposition.json`, disposition);
console.log(`${connector}: ${sourceOperations.length} source rows, ${classCounts.etl} ETL bindings, ${directWrites} direct writes`);
