// Generates website/lib/connectors.catalog.generated.ts and the large JSON
// payload it imports from website/data/connectors.generated.json. The source
// JSON data is generated from internal/connectors/defs/<name>/ bundles by
// gen-connector-bundles.mjs.
//
// Run: node scripts/gen-connector-catalog.mjs   (or: npm run gen:catalog)

import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import { mapCLISurface } from './lib/cli-surface.mjs';

const __dirname = dirname(fileURLToPath(import.meta.url));
const DATA = resolve(__dirname, '../data/connectors.generated.json');
const OUT = resolve(__dirname, '../lib/connectors.catalog.generated.ts');
const JSON_OUT = resolve(__dirname, '../lib/connectors.catalog.data.generated.json');

const FEATURED = new Set([
  'GitHub', 'Stripe', 'Postgres', 'MySQL', 'MongoDB', 'Snowflake', 'BigQuery',
  'Redshift', 'Salesforce', 'HubSpot', 'Slack', 'Notion', 'Shopify', 'Linear',
  'Jira', 'Zendesk', 'Intercom', 'Google Sheets', 'Airtable', 'Mixpanel',
  'Amplitude', 'Datadog', 'Twilio', 'SendGrid', 'S3', 'Asana', 'Monday',
  'GitLab', 'Bitbucket', 'Confluence', 'PagerDuty', 'Klaviyo', 'Mailchimp',
  'Pipedrive', 'Close', 'Copper', 'Freshdesk', 'Zoom', 'Google Analytics',
  'Google Ads', 'Facebook Marketing', 'Amazon Ads', 'Amazon Seller Partner',
  'BambooHR', 'Xero', 'Chargebee', 'Braintree', 'PayPal', 'Square',
  'Typeform', 'SurveyMonkey', 'Calendly', 'Drift', 'Customer.io', 'Braze',
  'Auth0', 'Azure Blob Storage', 'Google Cloud Storage', 'Elasticsearch',
  'ClickHouse', 'Oracle', 'DynamoDB', 'Redis', 'Kafka',
]);

const CATEGORY_LABELS = {
  api: 'API',
  database: 'Database',
  file: 'File',
  vectorstore: 'Vector store',
  message_queue: 'Message queue',
  queue: 'Queue',
  accounting: 'Accounting',
  other: 'Other',
};

const trim = (value) => (typeof value === 'string' ? value.trim() : '');
const titleCase = (value) =>
  trim(value)
    .split(/[_\s-]+/)
    .filter(Boolean)
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ');
const categoryLabel = (value) => CATEGORY_LABELS[value] ?? titleCase(value) ?? 'API';

const raw = JSON.parse(readFileSync(DATA, 'utf8'));
const items = Array.isArray(raw) ? raw : raw.connectors ?? [];

function mapCapabilities(capabilities = {}) {
  return {
    check: capabilities.check === true,
    read: capabilities.read === true,
    write: capabilities.write === true,
    query: capabilities.query === true,
    cdc: capabilities.cdc === true,
    dynamicSchema: capabilities.dynamic_schema === true,
  };
}

function capabilityLabels(capabilities) {
  return Object.entries({
    check: 'check',
    read: 'read',
    write: 'write',
    query: 'query',
    cdc: 'cdc',
    dynamicSchema: 'dynamic schema',
  })
    .filter(([key]) => capabilities[key])
    .map(([, label]) => label);
}

function mapConnector(item) {
  const capabilities = mapCapabilities(item.capabilities);
  const streams = (Array.isArray(item.streams) ? item.streams : [])
    .map((stream) => ({
      name: trim(stream.name),
      primaryKey: Array.isArray(stream.primary_key)
        ? stream.primary_key.filter((key) => trim(key)).map((key) => trim(key))
        : [],
      cursor: trim(stream.cursor),
      incremental: stream.incremental === true,
    }))
    .filter((stream) => stream.name);

  const writeActions = (Array.isArray(item.write_actions) ? item.write_actions : [])
    .map((action) => ({
      name: trim(action.name),
      method: trim(action.method).toUpperCase(),
      kind: trim(action.kind),
    }))
    .filter((action) => action.name);

  const docsUrl = trim(item.docs_url);
  const releaseStage = trim(item.release_stage);
  const category = trim(item.integration_type) || 'api';

  return {
    slug: trim(item.slug),
    name: trim(item.name),
    description: trim(item.description),
    category,
    categoryLabel: categoryLabel(category),
    releaseStage,
    status: 'available',
    capabilities,
    capabilityLabels: capabilityLabels(capabilities),
    streams,
    writeActions,
    cliSurface: mapCLISurface(item.cli_surface, { keyStyle: 'camel' }),
    docsMd: trim(item.docs_md),
    docs: docsUrl ? [{ title: 'Service API documentation', type: 'api_reference', url: docsUrl }] : [],
    docUrl: docsUrl,
    appDocUrl: docsUrl,
    icon: item.icon ?? null,
    featured: FEATURED.has(trim(item.name)),
  };
}

const all = items
  .filter((item) => item && trim(item.slug) && trim(item.name))
  .map(mapConnector)
  .sort((a, b) => a.name.localeCompare(b.name, 'en', { sensitivity: 'base' }));

const categoryCounts = {};
const releaseStageCounts = {};
const capabilityCounts = {
  check: 0,
  read: 0,
  write: 0,
  query: 0,
  cdc: 0,
  dynamicSchema: 0,
};

for (const c of all) {
  categoryCounts[c.category] = (categoryCounts[c.category] || 0) + 1;
  releaseStageCounts[c.releaseStage || 'unknown'] =
    (releaseStageCounts[c.releaseStage || 'unknown'] || 0) + 1;
  for (const key of Object.keys(capabilityCounts)) {
    if (c.capabilities[key]) capabilityCounts[key]++;
  }
}

const banner =
  `// AUTO-GENERATED by scripts/gen-connector-catalog.mjs - DO NOT EDIT.\n` +
  `// Source: website/data/connectors.generated.json\n` +
  `// Run \`npm run gen:catalog\` to regenerate.\n\n`;

const body =
  banner +
  `import catalogData from './connectors.catalog.data.generated.json';\n` +
  `import type { ConnectorCapabilities, ConnectorMeta } from './connectors.types';\n\n` +
  `export const CONNECTOR_CATALOG = catalogData as ConnectorMeta[];\n\n` +
  `export const CONNECTOR_CATALOG_COUNT = ${all.length};\n\n` +
  `export const CONNECTOR_CATEGORY_COUNTS = ${JSON.stringify(categoryCounts)} as Record<string, number>;\n\n` +
  `export const CONNECTOR_RELEASE_STAGE_COUNTS = ${JSON.stringify(releaseStageCounts)} as Record<string, number>;\n\n` +
  `export const CONNECTOR_CAPABILITY_COUNTS = ${JSON.stringify(capabilityCounts)} as Record<keyof ConnectorCapabilities, number>;\n\n` +
  `const BY_SLUG: Record<string, ConnectorMeta> = Object.fromEntries(\n` +
  `  CONNECTOR_CATALOG.map((c) => [c.slug, c]),\n);\n\n` +
  `export function connectorBySlug(slug: string): ConnectorMeta | undefined {\n` +
  `  return BY_SLUG[slug];\n}\n`;

mkdirSync(dirname(OUT), { recursive: true });
const catalogJson = JSON.stringify(all, null, 0);
writeFileSync(JSON_OUT, `${catalogJson}\n`, 'utf8');
writeFileSync(OUT, body, 'utf8');

console.log(
  `Wrote ${all.length} connectors to lib/connectors.catalog.generated.ts and ` +
    `lib/connectors.catalog.data.generated.json ` +
    `(${((body.length + catalogJson.length) / 1024).toFixed(0)} KB).\n` +
    `Categories: ${JSON.stringify(categoryCounts)}\n` +
    `Capabilities: ${JSON.stringify(capabilityCounts)}\n` +
    `Featured: ${all.filter((c) => c.featured).length}`,
);
