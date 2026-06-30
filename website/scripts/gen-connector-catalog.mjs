// Generates website/lib/connectors.catalog.generated.ts from the Go connector
// catalog (internal/connectors/catalog_data.json). This is the data backbone for
// the per-connector doc pages and the connector browser.
//
// Run: node scripts/gen-connector-catalog.mjs   (or: npm run gen:catalog)
//
// One entry per catalog slug (source-* / destination-*). Excludes the huge raw
// config_schema; keeps the structured config_fields summary. Output is committed
// so the website build has no dependency on the Go tree.

import { copyFileSync, existsSync, mkdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { dirname, relative, resolve, sep } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const CATALOG = resolve(__dirname, '../../internal/connectors/catalog_data.json');
const ICON_DATA = resolve(__dirname, '../../internal/connectors/icon_data.json');
const ICON_SOURCE_ROOT = resolve(__dirname, '../../docs/connectors');
const ICON_PUBLIC_ROOT = resolve(__dirname, '../public/connectors');
const OUT = resolve(__dirname, '../lib/connectors.catalog.generated.ts');

// Featured connectors get SearXNG-enriched setup/auth prose later (Phase 5).
// The 2 enabled + the most recognizable names.
const FEATURED = new Set([
  'GitHub', 'Stripe', 'Postgres', 'MySQL', 'MongoDB', 'Snowflake', 'BigQuery',
  'Redshift', 'Salesforce', 'HubSpot', 'Slack', 'Notion', 'Shopify', 'Linear',
  'Jira', 'Zendesk', 'Intercom', 'Google Sheets', 'Airtable', 'Mixpanel',
  'Amplitude', 'Segment', 'Datadog', 'Twilio', 'S3', 'Asana', 'GitLab',
  'Klaviyo', 'Mailchimp', 'Pipedrive', 'Zoom', 'Google Ads', 'Facebook Marketing',
  'Amazon Ads', 'Amazon Seller Partner', 'Greenhouse', 'BambooHR', 'NetSuite',
  'QuickBooks', 'Xero', 'PayPal', 'Square',
]);

const raw = JSON.parse(readFileSync(CATALOG, 'utf8'));
const iconRaw = JSON.parse(readFileSync(ICON_DATA, 'utf8'));
const items = Array.isArray(raw) ? raw : (raw.connectors ?? raw.definitions ?? []);
const iconByConnector = new Map();
for (const icon of Array.isArray(iconRaw) ? iconRaw : []) {
  if (!icon?.connector) continue;
  if (iconByConnector.has(icon.connector)) {
    throw new Error(`Duplicate connector icon metadata for ${icon.connector}`);
  }
  iconByConnector.set(icon.connector, icon);
}
const copiedIconPaths = new Set();

const trim = (s) => (typeof s === 'string' ? s.trim() : '');
const validIconPath = (path) => /^icons\/[A-Za-z0-9._-]+\.svg$/.test(path);

function assertInside(root, target, label) {
  const rel = relative(root, target);
  if (rel.startsWith('..') || rel === '..' || rel.includes(`..${sep}`) || rel === '') {
    throw new Error(`${label} escapes expected root: ${target}`);
  }
}

function resolveIconPath(root, iconPath, label) {
  if (!validIconPath(iconPath)) {
    throw new Error(`Invalid connector icon path: ${iconPath}`);
  }

  const target = resolve(root, iconPath);
  assertInside(root, target, label);
  return target;
}

// Strip Airbyte branding from connector-spec prose (descriptions are authored
// upstream). Applied to all connectors EXCEPT the Airbyte connector itself.
function sanitizeDesc(s) {
  let t = trim(s);
  if (!t) return '';
  t = t.replace(/<[^>]+>/g, ' '); // strip embedded HTML (often <a href="docs.airbyte.com">)
  t = t.replace(/https?:\/\/(?:docs\.|reference\.)?airbyte\.com\/\S*/gi, ''); // drop airbyte URLs
  t = t.replace(/airbytehq\/airbyte/gi, 'owner/repo');
  t = t.replace(/airbytehq/gi, 'owner');
  // De-brand technical identifiers / example values the word-boundary rule misses:
  t = t.replace(/airbyte_internal/gi, 'pm_internal'); // default raw-table schema name
  t = t.replace(/_airbyte_/gi, '_pm_'); // metadata column prefix (e.g. _airbyte_data)
  t = t.replace(/airbyteio/gi, 'example'); // example tenant in domain samples
  t = t.replace(/\bairbyte\b/gi, 'the connector');
  t = t.replace(/&amp;/g, '&').replace(/&lt;/g, '<').replace(/&gt;/g, '>');
  t = t.replace(/\s+([.,;:])/g, '$1').replace(/\s{2,}/g, ' ').trim();
  return t;
}
const isAirbyteConnector = (e) =>
  /airbyte/i.test(e.slug || '') || (e.name || '').toLowerCase() === 'airbyte';

function mapIcon(e) {
  const icon = iconByConnector.get(e.slug);
  if (!icon) return null;

  const path = trim(icon.path);
  if (!validIconPath(path)) {
    throw new Error(`Invalid connector icon path for ${e.slug}: ${path}`);
  }

  copiedIconPaths.add(path);
  return {
    id: trim(icon.id),
    path,
    publicPath: `/connectors/${path}`,
    source: trim(icon.source),
    reviewStatus: trim(icon.review_status),
    reviewUrl: trim(icon.review_url),
  };
}

function syncConnectorIcons(paths) {
  const outDir = resolve(ICON_PUBLIC_ROOT, 'icons');

  for (const iconPath of [...paths].sort()) {
    const src = resolveIconPath(ICON_SOURCE_ROOT, iconPath, 'connector icon source');
    if (!existsSync(src)) {
      throw new Error(`Missing connector icon asset: ${iconPath}`);
    }
  }

  rmSync(outDir, { recursive: true, force: true });
  mkdirSync(outDir, { recursive: true });

  for (const iconPath of [...paths].sort()) {
    const src = resolveIconPath(ICON_SOURCE_ROOT, iconPath, 'connector icon source');
    const out = resolveIconPath(ICON_PUBLIC_ROOT, iconPath, 'connector icon output');
    mkdirSync(dirname(out), { recursive: true });
    copyFileSync(src, out);
  }
}

function mapConnector(e) {
  const keepAirbyte = isAirbyteConnector(e);
  const desc = (s) => (keepAirbyte ? trim(s) : sanitizeDesc(s));
  return {
    slug: e.slug,
    name: e.name,
    type: e.type, // "source" | "destination"
    category: e.source_type || 'other', // api | database | file | vectorstore | message_queue
    releaseStage: e.release_stage || '',
    supportLevel: e.support_level || '',
    language: e.language || '',
    tags: Array.isArray(e.tags) ? e.tags : [],
    status: e.implementation_status, // enabled | planned_native_port | unsupported_deprecated
    runtimeKind: e.runtime_kind || '',
    capabilities: {
      metadata: !!e.runtime_capabilities?.metadata,
      check: !!e.runtime_capabilities?.check,
      catalog: !!e.runtime_capabilities?.catalog,
      read: !!e.runtime_capabilities?.read,
      write: !!e.runtime_capabilities?.write,
      query: !!e.runtime_capabilities?.query,
      etl: !!e.runtime_capabilities?.etl,
      reverseEtl: !!e.runtime_capabilities?.reverse_etl,
      unsupportedReason: trim(e.runtime_capabilities?.unsupported_reason),
    },
    syncModes: Array.isArray(e.supported_sync_modes) ? e.supported_sync_modes : [],
    incremental: !!e.supports_incremental,
    config: (Array.isArray(e.config_fields) ? e.config_fields : []).map((f) => ({
      name: f.name,
      type: f.type || '',
      description: desc(f.description),
      required: !!f.required,
      secret: !!f.secret,
    })),
    secrets: Array.isArray(e.secret_fields) ? e.secret_fields : [],
    // Drop any vendor-doc link that points at (or is titled after) Airbyte,
    // except on the Airbyte connector itself. A handful of utility connectors
    // (file, datagen, e2e-test, web-scrapper) only had airbyte.com docs; they
    // now show "No documentation links" like other doc-less connectors.
    docs: (Array.isArray(e.official_application_docs) ? e.official_application_docs : [])
      .filter((d) => keepAirbyte || !/airbyte/i.test(`${d.title || ''} ${d.url || ''}`))
      .map((d) => ({ title: d.title, type: d.type || '', url: d.url })),
    // Upstream (Airbyte) catalog doc URL — kept only for the Airbyte connector;
    // blanked elsewhere so no Airbyte links surface anywhere.
    docUrl: keepAirbyte ? (e.documentation_url || '') : '',
    appDocUrl: keepAirbyte
      ? (e.application_documentation_url || '')
      : /airbyte/i.test(e.application_documentation_url || '')
        ? ''
        : e.application_documentation_url || '',
    pmName: e.pm_connector_name || '',
    notes: desc(e.native_support_notes),
    icon: mapIcon(e),
    featured: FEATURED.has(e.name),
  };
}

const all = items
  .filter((e) => e && e.slug && e.name)
  .map(mapConnector)
  .sort((a, b) => a.name.localeCompare(b.name, 'en', { sensitivity: 'base' }) || a.type.localeCompare(b.type));

syncConnectorIcons(copiedIconPaths);

// ── Derived indexes ─────────────────────────────────────────────────────
const categoryCounts = {};
const statusCounts = {};
let sources = 0;
let destinations = 0;
for (const c of all) {
  categoryCounts[c.category] = (categoryCounts[c.category] || 0) + 1;
  statusCounts[c.status] = (statusCounts[c.status] || 0) + 1;
  if (c.type === 'source') sources++;
  else if (c.type === 'destination') destinations++;
}

const banner =
  `// AUTO-GENERATED by scripts/gen-connector-catalog.mjs - DO NOT EDIT.\n` +
  `// Source: internal/connectors/catalog_data.json\n` +
  `// Run \`npm run gen:catalog\` to regenerate.\n\n`;

const types =
  `export type ConnectorCapabilities = {\n` +
  `  metadata: boolean; check: boolean; catalog: boolean; read: boolean;\n` +
  `  write: boolean; query: boolean; etl: boolean; reverseEtl: boolean;\n` +
  `  unsupportedReason: string;\n};\n\n` +
  `export type ConnectorConfigField = {\n` +
  `  name: string; type: string; description: string; required: boolean; secret: boolean;\n};\n\n` +
  `export type ConnectorDocLink = { title: string; type: string; url: string };\n\n` +
  `export type ConnectorIconMeta = {\n` +
  `  id: string; path: string; publicPath: string; source: string;\n` +
  `  reviewStatus: string; reviewUrl: string;\n` +
  `};\n\n` +
  `export type ConnectorMeta = {\n` +
  `  slug: string;\n  name: string;\n  type: 'source' | 'destination';\n  category: string;\n` +
  `  releaseStage: string;\n  supportLevel: string;\n  language: string;\n  tags: string[];\n` +
  `  status: string;\n  runtimeKind: string;\n  capabilities: ConnectorCapabilities;\n` +
  `  syncModes: string[];\n  incremental: boolean;\n  config: ConnectorConfigField[];\n` +
  `  secrets: string[];\n  docs: ConnectorDocLink[];\n  docUrl: string;\n  appDocUrl: string;\n` +
  `  pmName: string;\n  notes: string;\n  icon: ConnectorIconMeta | null;\n  featured: boolean;\n};\n\n`;

const body =
  banner +
  types +
  `export const CONNECTOR_CATALOG: ConnectorMeta[] = ${JSON.stringify(all, null, 0)};\n\n` +
  `export const CONNECTOR_CATALOG_COUNT = ${all.length};\n\n` +
  `export const CONNECTOR_SOURCES = ${sources};\n` +
  `export const CONNECTOR_DESTINATIONS = ${destinations};\n\n` +
  `export const CONNECTOR_CATEGORY_COUNTS: Record<string, number> = ${JSON.stringify(categoryCounts)};\n\n` +
  `export const CONNECTOR_STATUS_COUNTS: Record<string, number> = ${JSON.stringify(statusCounts)};\n\n` +
  `const BY_SLUG: Record<string, ConnectorMeta> = Object.fromEntries(\n` +
  `  CONNECTOR_CATALOG.map((c) => [c.slug, c]),\n);\n\n` +
  `export function connectorBySlug(slug: string): ConnectorMeta | undefined {\n` +
  `  return BY_SLUG[slug];\n}\n`;

mkdirSync(dirname(OUT), { recursive: true });
writeFileSync(OUT, body, 'utf8');

console.log(
  `Wrote ${all.length} connectors (${sources} sources, ${destinations} destinations) ` +
    `to lib/connectors.catalog.generated.ts (${(body.length / 1024).toFixed(0)} KB).\n` +
    `Categories: ${JSON.stringify(categoryCounts)}\nStatus: ${JSON.stringify(statusCounts)}\n` +
    `Featured: ${all.filter((c) => c.featured).length}\n` +
    `Icons: ${copiedIconPaths.size} copied to public/connectors/icons`,
);
