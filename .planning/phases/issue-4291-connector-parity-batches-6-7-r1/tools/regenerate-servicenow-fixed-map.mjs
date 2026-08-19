#!/usr/bin/env node

// ServiceNow's Table API has six fixed operation templates but an
// instance-defined table and field schema. This generator deliberately maps
// only that fixed platform surface; it never fabricates a count of customer
// tables or fields.
import { createHash } from 'node:crypto';
import { writeFileSync } from 'node:fs';

const connector = 'service-now';
const base = `internal/connectors/defs/${connector}`;
const root = new URL('../../../../', import.meta.url);
const writeJSON = (path, value) => writeFileSync(new URL(path, root), `${JSON.stringify(value, null, 2)}\n`);
const sourceURL = 'https://www.servicenow.com/docs/bundle/zurich-api-reference/page/integrate/inbound-rest/concept/c_TableAPI.html';
const source = await fetch(sourceURL);
const bytes = Buffer.from(await source.arrayBuffer());
if (!source.ok) throw new Error(`${sourceURL}: HTTP ${source.status}`);
const sha256 = createHash('sha256').update(bytes).digest('hex');
const gap = {
  id: 'generic-typed-destination-executor',
  evidence: 'internal/app/issue_label_warehouse_transport.go:85-95: the only destination factory selects issueLabelDestinationReference and BuildDestination enforces the GitHub issue-label contract.',
  minimal_change: 'register a connector-neutral typed destination DefinitionFactory selected by the definition, with per-connector evidence, explicit source bindings, acknowledgement and per-mode apply strategies',
};
const operations = [
  { method: 'GET', path: '/api/now/table/{table_name}', parity_class: 'etl', covered_by: { stream: 'incidents' } },
  { method: 'GET', path: '/api/now/table/{table_name}/{sys_id}', parity_class: 'direct_read' },
  { method: 'POST', path: '/api/now/table/{table_name}', parity_class: 'direct_write', covered_by: { write: 'create_incident' } },
  { method: 'PUT', path: '/api/now/table/{table_name}/{sys_id}', parity_class: 'direct_write' },
  { method: 'PATCH', path: '/api/now/table/{table_name}/{sys_id}', parity_class: 'direct_write', covered_by: { write: 'update_incident' } },
  { method: 'DELETE', path: '/api/now/table/{table_name}/{sys_id}', parity_class: 'direct_write' },
];
const coverageAliases = [
  { method: 'GET', path: '/api/now/table/{table_name}', covered_by: { stream: 'users' } },
  { method: 'GET', path: '/api/now/table/{table_name}', covered_by: { stream: 'groups' } },
  { method: 'POST', path: '/api/now/table/{table_name}', covered_by: { write: 'create_user' } },
  { method: 'POST', path: '/api/now/table/{table_name}', covered_by: { write: 'create_group' } },
  { method: 'PATCH', path: '/api/now/table/{table_name}/{sys_id}', covered_by: { write: 'update_user' } },
  { method: 'PATCH', path: '/api/now/table/{table_name}/{sys_id}', covered_by: { write: 'update_group' } },
];
const counts = Object.fromEntries(['GET', 'POST', 'PUT', 'PATCH', 'DELETE'].map(method => [method, operations.filter(operation => operation.method === method).length]).filter(([, count]) => count > 0));
const confidence = {
  level: 'fixed_platform_surface_with_dynamic_instance_schema',
  basis: 'The official ServiceNow Table API documents six uniform HTTP operation templates. The table name, fields, ACLs, and customer-defined tables are instance-dependent; this connector pins the six fixed templates and records the dynamic-schema basis rather than inventing a finite instance total.',
};
const sourceOperations = operations.map((operation, index) => ({
  id: `${connector}.table_api.${operation.method.toLowerCase()}.${index}`,
  protocol: 'rest', method: operation.method, path: operation.path, operation_id: null, deprecated: false,
  source_location: `${sourceURL}: Table API ${operation.method} ${operation.path}`,
  source_url: sourceURL,
}));
const sourceLock = {
  schema_version: 2, connector, captured_at: '2026-08-19T00:00:00Z',
  rest: { source_url: sourceURL, source_kind: 'fixed_platform_surface_with_dynamic_instance_schema', sha256, bytes: bytes.length, info_version: 'ServiceNow Zurich Table API', operation_counts: counts, operations: sourceOperations },
  counts: { rest: operations.length, graphql_query: 0, graphql_mutation: 0, total: operations.length },
  operations_found: { rest: operations.length, graphql_query: 0, graphql_mutation: 0, total: operations.length },
  coverage_confidence: confidence,
  dynamic_schema: {
    state: 'instance_dependent',
    basis: 'The request template accepts any table_name allowed by the target instance. Its dictionary fields and ACL-permitted data surface cannot be enumerated from the public platform reference.',
    fixed_platform_operations: operations.length,
    instance_operation_total: null,
  },
};
const ledger = operations.map((operation, index) => {
  const enabled = Boolean(operation.covered_by?.write);
  const pendingID = operation.parity_class === 'etl' ? `runnable-command-binding-${connector}` : `typed-operation-contract-${connector}`;
  const declaration = enabled
    ? { status: 'enabled; runnable typed write action binds the fixed provider operation template', contract: operation.covered_by }
    : { status: `disabled; declaration-pending ${pendingID}`, contract: null };
  if (operation.parity_class === 'direct_write') declaration.reverse_etl = { eligible_typed_action: enabled, state: 'foundation-gap', foundation_gap: gap };
  const evidence = `internal/connectors/defs/${connector}/sources/${connector}-operation-source-lock.json: rest.operations[${index}] pins the fixed Table API template; its table schema is instance-dependent.`;
  return {
    method: operation.method, path: operation.path, parity_class: operation.parity_class,
    api_surface: { method: operation.method, path: operation.path, operation: operation.covered_by ? null : { model: operation.parity_class === 'direct_write' ? 'disallowed' : operation.parity_class === 'etl' ? 'stream' : 'direct_read', status: 'blocked', risk: operation.parity_class === 'direct_write' ? 'high' : 'medium', blocked_by_default: true, reason: 'No exact connector-local runnable contract binds this fixed ServiceNow Table API operation template.' }, covered_by: operation.covered_by ?? null },
    source: { source_lock: `sources/${connector}-operation-source-lock.json`, source_id: sourceOperations[index].id, source_url: sourceURL, source_location: sourceOperations[index].source_location, operation_id: null, deprecated: false },
    state: enabled ? 'enabled' : 'disabled',
    foundation: enabled ? { state: 'present', evidence: 'A connector-owned typed action binds this fixed provider operation template.' } : { state: 'present', evidence: 'The platform operation is fixed, but a connector-local exact action or command is not yet bound.', declaration_pending: { id: pendingID, evidence, minimal_change: 'Add a source-backed connector-owned typed action or runnable command for this exact fixed template, or retain the operation as declaration-pending.' } },
    rejection: enabled ? null : { reason: 'declaration-pending', recoverable: true, detail: 'The fixed platform operation is documented; no exact connector-local runnable declaration binds it.', evidence },
    declaration,
  };
});
const classCounts = Object.fromEntries(['direct_read', 'direct_write', 'etl', 'reverse_etl', 'binary_read', 'binary_write'].map(key => [key, operations.filter(operation => operation.parity_class === key).length]));
const enabled = ledger.filter(row => row.state === 'enabled').length;
const disposition = {
  schema_version: 1, connector, generated_at: '2026-08-19T00:00:00Z',
  source_basis: { source_lock: `sources/${connector}-operation-source-lock.json`, source_url: sourceURL, source_sha256: sha256, source_bytes: bytes.length, source_operation_count: operations.length, operations_found: operations.length, coverage_confidence: confidence, dynamic_schema_basis: sourceLock.dynamic_schema },
  summary: {
    api_surface_rows: operations.length, exact_source_rows: operations.length, declared_operations: operations.length, operations_found: operations.length, coverage_confidence: confidence,
    dynamic_schema: sourceLock.dynamic_schema,
    enabled_operations: enabled, enabled_percent: Number((enabled * 100 / operations.length).toFixed(2)), disabled_operations: operations.length - enabled,
    documented_deletes: 1, enabled_deletes: 0,
    parity_class_counts: Object.entries(classCounts).map(([key, count]) => ({ key, count })), stream_bindings: classCounts.etl, writes_actions: 6, terminal_commands: 0,
    gap_ids: [gap.id], foundation_gaps: [{ id: gap.id, count: classCounts.direct_write }], rejected_by_reason: [{ key: 'declaration-pending', count: operations.length - enabled }],
    transport: { contract: 'docs/sync-transport-definition.md', source_transport: { state: 'declaration-pending', declaration_pending: { id: `sync-transport-source-definition-${connector}`, evidence: `docs/sync-transport-definition.md:15-38 lists the declaration fields; internal/connectors/defs/${connector}/sync_transport.json is absent.`, minimal_change: 'Add the connector-owned source transport declaration and conformance evidence; no engine change is required.' } }, destination_transport: { state: 'foundation-gap', foundation_gap: gap } },
    declaration_pending_ids: [`runnable-command-binding-${connector}`, `typed-operation-contract-${connector}`, `sync-transport-source-definition-${connector}`], declaration_pending: [{ id: `runnable-command-binding-${connector}`, count: 1 }, { id: `typed-operation-contract-${connector}`, count: operations.length - enabled - 1 }],
    runnable_cli_surface_commands: 0, typed_write_actions: 6, endpoint_bound_cli_commands: 0, endpoint_bound_typed_write_actions: 2,
  },
  ledger_dispositions: ledger,
  source_only_dispositions: [],
  notes: [
    'The Table API operation count is the fixed public platform-template count only. It is not a fabricated count of target-instance tables or fields.',
    'Direct writes are direct_write. Their reverse-ETL eligibility is a separate foundation-gapped attribute until #4303 provides the connector-neutral typed destination executor.',
    'No transport binding is declared in this map.',
  ],
};
writeJSON(`${base}/sources/${connector}-operation-source-lock.json`, sourceLock);
writeJSON(`${base}/api_surface.json`, { api: 'ServiceNow Table API fixed platform surface', docs: sourceURL, reviewed_at: '2026-08-19', operation_ledger_version: 2, scope: 'The six public Table API method/path templates are fixed and pinned. Target tables, fields, ACLs, and customer schema are instance-dependent and are explicitly recorded in the source lock rather than counted as a fabricated finite API surface. Six additional API-surface rows project the same documented templates to distinct existing streams/actions and do not increase the fixed six-operation denominator.', artifacts: [{ id: 'servicenow-zurich-table-api-2026-08-19', url: sourceURL, retrieved_at: '2026-08-19', sha256 }], endpoints: [...ledger.map(row => ({ method: row.method, path: row.path, provenance: { artifact: 'servicenow-zurich-table-api-2026-08-19', source_url: sourceURL }, ...(row.api_surface.covered_by ? { covered_by: row.api_surface.covered_by } : { operation: row.api_surface.operation }) })), ...coverageAliases.map(alias => ({ ...alias, provenance: { artifact: 'servicenow-zurich-table-api-2026-08-19', source_url: sourceURL } }))] });
writeJSON(`${base}/sources/${connector}-declaration-disposition.json`, disposition);
console.log(`service-now: ${operations.length} fixed templates; enabled=${enabled}; dynamic instance surface uncounted`);
