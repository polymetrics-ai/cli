import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

const phaseDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), '..');
const operationEvidencePath = path.join(phaseDir, 'OPERATION-SURFACE-EVIDENCE.json');
const foundationGapsPath = path.join(phaseDir, 'FOUNDATION-GAPS.json');
const routeFoundationSHA = '6410fe59c';
const destinationFoundationSHA = '609f23bb3861ba7bc2ef1f7bc5246f5751cf9e57';

function readJSON(file) {
  return JSON.parse(fs.readFileSync(file, 'utf8'));
}

function writeJSON(file, value) {
  fs.writeFileSync(file, `${JSON.stringify(value, null, 2)}\n`);
}

function requiredRow(rows, sourceID) {
  const row = rows.find((candidate) => candidate.provider_operation?.source_id === sourceID);
  if (!row) throw new Error(`missing operation-evidence row ${sourceID}`);
  return row;
}

const helpScoutRoutes = [
  ['help-scout.rest.get.V3ConversationsConversationId139', 'help-scout.v3_get_conversation', 'direct get-conversation-v3'],
  ['help-scout.rest.get.V3ConversationsConversationIdThreads140', 'help-scout.v3_list_conversation_threads', 'direct list-conversation-threads-v3'],
  ['help-scout.rest.get.V3Customers141', 'help-scout.v3_list_customers', 'direct list-customers-v3'],
  ['help-scout.rest.get.V3SystemUsers142', 'help-scout.v3_list_system_users', 'direct list-system-users-v3'],
  ['help-scout.rest.get.V3SystemUsersSystemUserId143', 'help-scout.v3_get_system_user', 'direct get-system-user-v3'],
];

const operationEvidence = readJSON(operationEvidencePath);
for (const [sourceID, operation, command] of helpScoutRoutes) {
  const row = requiredRow(operationEvidence.rows, sourceID);
  row.canonical_mapping = {
    source_state: 'enabled',
    declaration_status: 'enabled; runtime-preflight runnable command binds the source-locked v3 route',
    contract: { operation, direct_read: command },
    api_surface_binding: { direct_read: command },
    rejection: null,
  };
  row.surfaces.direct_read = { state: 'mapped', source_state: 'enabled', rejection: null };
  row.surfaces.executable_cli = { state: 'generated_binding_present', binding: { direct_read: command } };
  row.merge_readiness = { enabled: true, open_foundation_gap: null, merge_ready: false };
}

const customerRow = requiredRow(operationEvidence.rows, 'help-scout.rest.patch.V2CustomersCustomerId31');
customerRow.surfaces.reverse_etl = {
  state: 'bound_foundation_available_pending_connector_app_cli_proof',
  eligible_typed_action: true,
  foundation_gap: null,
};
customerRow.merge_readiness = { enabled: true, open_foundation_gap: null, merge_ready: false };
writeJSON(operationEvidencePath, operationEvidence);

const gaps = readJSON(foundationGapsPath);
const definitions = new Map(gaps.gap_definitions.map((definition) => [definition.stable_gap_id, definition]));
for (const [id, sha] of [
  ['declarative-operation-route-override', routeFoundationSHA],
  ['declarative-typed-destination-action-specific-source-bindings', destinationFoundationSHA],
]) {
  const definition = definitions.get(id);
  if (!definition) throw new Error(`missing gap definition ${id}`);
  definition.status = 'resolved';
  definition.resolved_by_foundation_sha = sha;
}

const newDefinitions = [
  {
    stable_gap_id: 'scalar_json_write_body',
    missing_provider_neutral_capability: 'A declaration-owned, schema-validated JSON request-body mode for exactly one top-level scalar rather than the existing object or array body forms.',
    owning_issue: '#4291 (foundation handoff required)',
    owning_lane: 'unassigned',
    status: 'open',
  },
  {
    stable_gap_id: 'structured_recursive_filter_input',
    missing_provider_neutral_capability: 'A closed operation-specific structured input/builder that validates the documented recursive provider filter grammar without admitting a generic JSON body escape hatch.',
    owning_issue: '#4291 (foundation handoff required)',
    owning_lane: 'unassigned',
    status: 'open',
  },
  {
    stable_gap_id: 'post_binary_text_export',
    missing_provider_neutral_capability: 'A bounded declaration-owned POST binary/text-export executor with exact request-body, response, and destination contracts.',
    owning_issue: '#4291 (foundation handoff required)',
    owning_lane: 'unassigned',
    status: 'open',
  },
  {
    stable_gap_id: 'put_operation_direct_read',
    missing_provider_neutral_capability: 'A bounded operation direct-read executor that faithfully supports the provider-defined PUT method and validated request body.',
    owning_issue: '#4291 (foundation handoff required)',
    owning_lane: 'unassigned',
    status: 'open',
  },
];
for (const definition of newDefinitions) {
  if (!definitions.has(definition.stable_gap_id)) gaps.gap_definitions.push(definition);
}

for (const row of gaps.rows) {
  if (row.stable_gap_id === 'declarative-operation-route-override') {
    row.status = 'resolved';
    row.resolved_by_foundation_sha = routeFoundationSHA;
    row.runtime_reachability = {
      status: 'enabled_fixture_proven',
      reason: 'Definition-owned mailbox_v3 route resolves the source-locked path before I/O; TestHelpScoutV3DirectReadsUseTheirDeclaredRoute passes.',
    };
  }
  if (row.stable_gap_id === 'declarative-typed-destination-action-specific-source-bindings') {
    row.status = 'resolved';
    row.resolved_by_foundation_sha = destinationFoundationSHA;
    row.runtime_reachability = {
      status: 'foundation_resolved_connector_app_cli_proof_pending',
      reason: 'The connector now owns a customers(id) -> update_customer(customerId) action-specific source binding; source binding is no longer a shared refusal.',
    };
  }
}

const gorgiasSource = {
  url: 'https://dash.readme.com/api/v1/api-registry/1qfhqbgmshn434r',
  version: 'Gorgias REST API public OpenAPI specification',
  sha256: 'b824de51f33dfa90c6c6115eef74ff4b62277634a471347b01d769176f161897',
  source_lock: 'sources/gorgias-operation-source-lock.json',
};
const closureVerification = [
  'Add focused executor tests proving the closed provider shape and rejecting generic/raw-body bypasses before I/O',
  'go run ./cmd/connectorgen validate',
  'go run ./cmd/connectorgen surface-sync --check',
  'go test -timeout 20m ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight -count=1',
  'exercise the installed generated command through no-credential preflight and a local fixture',
];
const gorgiasRows = [
  ['scalar_json_write_body', 'gorgias.rest.put.UpdateCustomerCustomFieldValue91', 'update-customer-custom-field-value', 'PUT', '/api/customers/{customer_id}/custom-fields/{id}', 'https://developers.gorgias.com/reference/update-customer-custom-field-value', 'direct_write', ['direct_write', 'reverse_etl', 'executable_cli_command'], 'internal/connectors/engine/write.go:674-692 builds the default JSON payload as map[string]any and has no scalar body type.', 'Add a closed scalar JSON write body type that validates exactly the provider-declared string/integer/boolean schema.'],
  ['scalar_json_write_body', 'gorgias.rest.put.UpdateTicketCustomField107', 'update-ticket-custom-field', 'PUT', '/api/tickets/{ticket_id}/custom-fields/{id}', 'https://developers.gorgias.com/reference/update-ticket-custom-field', 'direct_write', ['direct_write', 'reverse_etl', 'executable_cli_command'], 'internal/connectors/engine/write.go:674-692 builds the default JSON payload as map[string]any and has no scalar body type.', 'Add a closed scalar JSON write body type that validates exactly the provider-declared string/integer/boolean schema.'],
  ['structured_recursive_filter_input', 'gorgias.rest.post.GetStatistic71', 'get-statistic', 'POST', '/api/reporting/stats', 'https://developers.gorgias.com/reference/get-statistic', 'direct_read', ['direct_read', 'executable_cli_command'], 'internal/connectors/engine/direct_read.go:286-293 and :785-786 admit raw input only for the exact text/plain root-string contract, not an arbitrary JSON filter tree.', 'Add a closed structured filter-expression input/builder for the documented recursive 50-arm provider grammar.'],
  ['post_binary_text_export', 'gorgias.rest.post.DownloadLegacyStatistic77', 'download-legacy-statistic', 'POST', '/api/stats/{name}/download', 'https://developers.gorgias.com/reference/download-legacy-statistic', 'binary_read', ['binary_download', 'executable_cli_command'], 'internal/connectors/engine/binary_read.go:292-293 and :333-334 require GET.', 'Add a bounded POST-capable binary/text-export executor with exact request-body and response contracts.'],
  ['put_operation_direct_read', 'gorgias.rest.put.UpdateViewItems112', 'update-view-items', 'PUT', '/api/views/{view_id}/items', 'https://developers.gorgias.com/reference/update-view-items', 'direct_read', ['direct_read', 'executable_cli_command'], 'internal/connectors/engine/direct_read.go:429-432 accepts only GET or POST operation direct reads.', 'Add a bounded PUT-capable operation direct-read executor with the current route/body validation contract.'],
];
for (const [stable_gap_id, source_id, operation_id, method, pathValue, url, parity_class, affected_surfaces, failing, minimal] of gorgiasRows) {
  if (gaps.rows.some((row) => row.stable_gap_id === stable_gap_id && row.provider_operation?.source_id === source_id)) continue;
  gaps.rows.push({
    stable_gap_id,
    connector: 'gorgias',
    provider_operation: { source_id, operation_id, method, path: pathValue, canonical_mapping: { parity_class, cli_command: null, disposition: 'not_enabled_foundation_gap' } },
    source: { ...gorgiasSource, url },
    affected_surfaces,
    failing_evidence: [failing],
    missing_provider_neutral_capability: definitions.get(stable_gap_id)?.missing_provider_neutral_capability,
    runtime_reachability: { status: 'not_enabled', reason: `Open provider-neutral ${stable_gap_id} foundation gap; no connector-local shim or generic escape hatch is permitted.` },
    closure_requires_generated_projections: true,
    owning_issue: '#4291 (foundation handoff required)',
    owning_lane: 'unassigned',
    status: 'open',
    closure_verification: [...closureVerification, minimal],
  });
}

const batchForConnector = new Map([
  ['close-com', 'batch_6'], ['outreach', 'batch_6'], ['salesloft', 'batch_6'], ['copper', 'batch_6'], ['zoho-bigin', 'batch_6'], ['klaviyo', 'batch_6'], ['braze', 'batch_6'], ['customer-io', 'batch_6'], ['intercom', 'batch_6'], ['freshdesk', 'batch_6'],
  ['segment', 'batch_7'], ['activecampaign', 'batch_7'], ['iterable', 'batch_7'], ['help-scout', 'batch_7'], ['gorgias', 'batch_7'], ['service-now', 'batch_7'], ['chatwoot', 'batch_7'], ['chargebee', 'batch_7'], ['square', 'batch_7'], ['braintree', 'batch_7'],
]);
const definitionByID = new Map(gaps.gap_definitions.map((definition) => [definition.stable_gap_id, definition]));
function rollup(rows) {
  return [...definitionByID.values()].map((definition) => {
    const affected = rows.filter((row) => row.stable_gap_id === definition.stable_gap_id);
    const connectors = [...new Set(affected.map((row) => row.connector))].sort();
    return {
      stable_gap_id: definition.stable_gap_id,
      status: definition.status,
      affected_operations: affected.length,
      affected_connectors: connectors,
      connector_count: connectors.length,
      owning_issue: definition.owning_issue,
      owning_lane: definition.owning_lane,
    };
  }).filter((entry) => entry.affected_operations > 0);
}
gaps.rollups.per_gap = rollup(gaps.rows);
gaps.rollups.per_batch = {
  batch_6: rollup(gaps.rows.filter((row) => batchForConnector.get(row.connector) === 'batch_6')),
  batch_7: rollup(gaps.rows.filter((row) => batchForConnector.get(row.connector) === 'batch_7')),
};
gaps.rollups.portfolio = {
  open_gap_ids: gaps.gap_definitions.filter((definition) => definition.status === 'open').map((definition) => definition.stable_gap_id),
  merge_ready: false,
  merge_ready_reason: 'Open provider-neutral foundation gaps prevent every affected operation from contributing to a merge-ready verdict.',
};
writeJSON(foundationGapsPath, gaps);
