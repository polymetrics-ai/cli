// Generates website/data/connectors.generated.json from the connector bundles.
// Run: node scripts/gen-connector-bundles.mjs
// Reads: internal/connectors/defs/<name>/{metadata.json,streams.json,writes.json,docs.md,cli_surface.json?}
// Emits: website/data/connectors.generated.json

import {
  readFileSync,
  writeFileSync,
  mkdirSync,
  readdirSync,
} from 'node:fs';
import { dirname, resolve, join } from 'node:path';
import { fileURLToPath } from 'node:url';

import { mapCLISurface } from './lib/cli-surface.mjs';
import {
  assertInside,
  collectConnectorIconPaths,
  syncConnectorIcons,
  validConnectorIconPath,
} from './lib/connector-icons.mjs';

const __dirname = dirname(fileURLToPath(import.meta.url));
const DEFS_ROOT = resolve(__dirname, '../../internal/connectors/defs');
const ICON_DATA = resolve(__dirname, '../../internal/connectors/icon_data.json');
const ICON_SOURCE_ROOT = resolve(__dirname, '../../docs/connectors');
const ICON_PUBLIC_ROOT = resolve(__dirname, '../public/connectors');
const OUT = resolve(__dirname, '../data/connectors.generated.json');

function readJSON(filePath) {
  try {
    return JSON.parse(readFileSync(filePath, 'utf8'));
  } catch {
    return null;
  }
}

function readMD(filePath) {
  try {
    return readFileSync(filePath, 'utf8');
  } catch {
    return '';
  }
}

function trim(value) {
  return typeof value === 'string' ? value.trim() : '';
}

function bool(value) {
  return value === true;
}

function normalizePrimaryKey(value) {
  if (Array.isArray(value)) {
    return value.filter((item) => trim(item)).map((item) => trim(item));
  }
  if (trim(value)) return [trim(value)];
  return [];
}

function readSchema(base, schemaPath) {
  if (!trim(schemaPath)) return null;
  const target = resolve(base, schemaPath);
  assertInside(base, target, 'stream schema');
  return readJSON(target);
}

const iconRaw = readJSON(ICON_DATA) ?? [];
const iconEntries = Array.isArray(iconRaw) ? iconRaw : [];
const copiedIconPaths = collectConnectorIconPaths(iconEntries);
const iconByConnector = new Map();
for (const icon of iconEntries) {
  const connector = trim(icon?.connector);
  if (!connector) continue;
  if (/^(source|destination)-/.test(connector)) {
    throw new Error(`Connector icon registry key must be bare: ${connector}`);
  }
  if (iconByConnector.has(connector)) {
    throw new Error(`Duplicate connector icon registry key: ${connector}`);
  }
  iconByConnector.set(connector, icon);
}

function mapIcon(slug) {
  const icon = iconByConnector.get(slug);
  if (!icon) {
    throw new Error(`Missing canonical connector icon registry entry for ${slug}`);
  }

  const path = trim(icon.path);
  if (!validConnectorIconPath(path)) {
    throw new Error(`Invalid connector icon path for ${slug}: ${path}`);
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

const entries = readdirSync(DEFS_ROOT, { withFileTypes: true })
  .filter(d => d.isDirectory())
  .map(d => d.name);

const connectors = [];

for (const dirName of entries) {
  const base = join(DEFS_ROOT, dirName);
  
  const metadata = readJSON(join(base, 'metadata.json'));
  if (!metadata) continue;

  const slug = trim(metadata.name || dirName);
  if (!slug) {
    throw new Error(`Connector bundle has empty name: ${dirName}`);
  }
  if (/^(source|destination)-/.test(slug)) {
    throw new Error(`Connector bundle must use a bare name: ${slug}`);
  }
  
  const streamsData = readJSON(join(base, 'streams.json'));
  const writesData = readJSON(join(base, 'writes.json'));
  const cliSurface = mapCLISurface(readJSON(join(base, 'cli_surface.json')));
  const docsMd = readMD(join(base, 'docs.md'));
  
  const streams = (streamsData?.streams ?? [])
    .map((stream) => {
      const schema = readSchema(base, stream.schema);
      const primaryKey = normalizePrimaryKey(
        schema?.['x-primary-key'] ?? stream.primary_key ?? stream.primaryKey,
      );
      const cursor = trim(
        stream.incremental?.cursor_field ??
          stream.cursor ??
          schema?.['x-cursor-field'] ??
          '',
      );

      return {
        name: trim(stream.name),
        primary_key: primaryKey,
        cursor,
        incremental: !!stream.incremental,
      };
    })
    .filter((stream) => stream.name);

  const writeActions = (writesData?.actions ?? [])
    .map((action) => ({
      name: trim(action.name),
      method: trim(action.method).toUpperCase(),
      kind: trim(action.kind),
    }))
    .filter((action) => action.name);
  
  const capabilities = metadata.capabilities ?? {};
  
  connectors.push({
    slug,
    name: trim(metadata.display_name) || slug,
    description: trim(metadata.description),
    docs_url: trim(metadata.docs_url),
    integration_type: trim(metadata.integration_type),
    release_stage: trim(metadata.release_stage),
    capabilities: {
      check: bool(capabilities.check),
      read: bool(capabilities.read),
      write: bool(capabilities.write),
      query: bool(capabilities.query),
      cdc: bool(capabilities.cdc),
      dynamic_schema: bool(capabilities.dynamic_schema),
    },
    streams,
    write_actions: writeActions,
    cli_surface: cliSurface,
    docs_md: docsMd,
    icon: mapIcon(slug),
  });
}

// Sort alphabetically by name
connectors.sort((a, b) => a.name.localeCompare(b.name, 'en', { sensitivity: 'base' }));

syncConnectorIcons(copiedIconPaths, {
  sourceRoot: ICON_SOURCE_ROOT,
  publicRoot: ICON_PUBLIC_ROOT,
});

mkdirSync(dirname(OUT), { recursive: true });
writeFileSync(OUT, JSON.stringify(connectors, null, 2), 'utf8');

console.log(
  `Wrote ${connectors.length} connectors to data/connectors.generated.json; ` +
    `${copiedIconPaths.size} icons copied.`,
);

// Report which connectors have write actions
const withWrites = connectors.filter(c => c.write_actions.length > 0);
console.log(`Connectors with write actions: ${withWrites.length}`);
