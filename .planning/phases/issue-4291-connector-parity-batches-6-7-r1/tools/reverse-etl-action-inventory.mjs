#!/usr/bin/env node

// This inventory is intentionally not a sync transport declaration. It gives
// #4303 adoption work the complete connector-owned direct-write source set and
// distinguishes actions already typed from operations that still need a closed
// action contract. No transport_binding, source binding, acknowledgement, or
// apply strategy can be authored until the neutral destination schema exists.
import { existsSync, readFileSync, writeFileSync } from 'node:fs';

const connectors = [
  'close-com', 'outreach', 'salesloft', 'copper', 'zoho-bigin', 'klaviyo', 'braze', 'customer-io', 'intercom', 'freshdesk',
  'segment', 'activecampaign', 'iterable', 'help-scout', 'gorgias', 'service-now', 'chatwoot', 'chargebee', 'square', 'braintree',
];
const root = new URL('../../../../', import.meta.url);
const readJSON = path => JSON.parse(readFileSync(new URL(path, root), 'utf8'));
const output = new URL('../REVERSE-ETL-ACTION-INVENTORY.json', import.meta.url);
const inventory = [];

for (const connector of connectors) {
  const base = `internal/connectors/defs/${connector}`;
  const disposition = readJSON(`${base}/sources/${connector}-declaration-disposition.json`);
  const writesPath = new URL(`${base}/writes.json`, root);
  const writes = existsSync(writesPath) ? readJSON(`${base}/writes.json`).actions ?? [] : [];
  const actions = new Map(writes.map(action => [action.name, action]));
  const operations = disposition.ledger_dispositions
    .filter(row => row.parity_class === 'direct_write')
    .map(row => {
      const actionName = row.api_surface.covered_by?.write ?? null;
      const action = actionName ? actions.get(actionName) : null;
      return {
        method: row.method,
        path: row.path,
        source_url: row.source.source_url,
        source_location: row.source.source_location,
        typed_action: actionName,
        typed_action_status: action ? 'authored' : 'needs_authoring',
        action_kind: action?.kind ?? null,
        current_reverse_etl_state: row.declaration.reverse_etl.state,
        current_foundation_gap: row.declaration.reverse_etl.foundation_gap.id,
        transport_qualification: 'pending #4303 connector-neutral typed destination contract',
      };
    });
  inventory.push({
    connector,
    documented_direct_writes: operations.length,
    typed_actions_authored: operations.filter(operation => operation.typed_action_status === 'authored').length,
    actions_needing_authoring: operations.filter(operation => operation.typed_action_status === 'needs_authoring').length,
    operations,
  });
}

const total = key => inventory.reduce((sum, connector) => sum + connector[key], 0);
writeFileSync(output, `${JSON.stringify({
  schema_version: 1,
  generated_at: '2026-08-19T00:00:00Z',
  issue: 4291,
  foundation_issue: 4303,
  prohibition: 'This inventory does not declare transport_binding, sync_transport, source bindings, acknowledgement, or apply strategies. The current destination executor is GitHub issue-label bound.',
  totals: {
    documented_direct_writes: total('documented_direct_writes'),
    typed_actions_authored: total('typed_actions_authored'),
    actions_needing_authoring: total('actions_needing_authoring'),
  },
  connectors: inventory,
}, null, 2)}\n`);
console.log(`reverse-etl action inventory: ${total('documented_direct_writes')} direct writes; ${total('typed_actions_authored')} typed actions; ${total('actions_needing_authoring')} need authoring`);
