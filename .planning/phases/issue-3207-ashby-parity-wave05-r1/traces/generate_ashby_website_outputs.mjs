import { readFileSync, writeFileSync } from 'node:fs';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import { mapCLISurface } from '../../../../website/scripts/lib/cli-surface.mjs';

const root = resolve(dirname(fileURLToPath(import.meta.url)), '../../../..');
const bundleRoot = join(root, 'internal/connectors/defs/ashby');
const websiteDataPath = join(root, 'website/data/connectors.generated.json');
const websiteCatalogPath = join(root, 'website/lib/connectors.catalog.data.generated.json');

const metadata = readJSON(join(bundleRoot, 'metadata.json'));
const streamBundle = readJSON(join(bundleRoot, 'streams.json'));
const writeBundle = readJSON(join(bundleRoot, 'writes.json'));
const cliSurface = readJSON(join(bundleRoot, 'cli_surface.json'));
const docsMd = readFileSync(join(bundleRoot, 'docs.md'), 'utf8');
const websiteData = readJSON(websiteDataPath);
const existing = websiteData.find((item) => item.slug === 'ashby');
if (!existing) throw new Error('ashby is missing from website connector data');

const streams = (streamBundle.streams ?? []).map((stream) => {
  const schema = readJSON(join(bundleRoot, stream.schema));
  const primaryKey = schema['x-primary-key'] ?? stream.primary_key ?? stream.primaryKey ?? [];
  const cursor = stream.incremental?.cursor_field ?? stream.cursor ?? schema['x-cursor-field'] ?? '';
  return {
    name: trim(stream.name),
    primary_key: Array.isArray(primaryKey) ? primaryKey.map(trim).filter(Boolean) : [trim(primaryKey)].filter(Boolean),
    cursor: trim(cursor),
    incremental: Boolean(stream.incremental),
  };
});

const replacement = {
  slug: 'ashby',
  name: trim(metadata.display_name) || 'ashby',
  description: trim(metadata.description),
  docs_url: trim(metadata.docs_url),
  integration_type: trim(metadata.integration_type),
  release_stage: trim(metadata.release_stage),
  capabilities: {
    check: metadata.capabilities?.check === true,
    read: metadata.capabilities?.read === true,
    write: metadata.capabilities?.write === true,
    query: metadata.capabilities?.query === true,
    cdc: metadata.capabilities?.cdc === true,
    dynamic_schema: metadata.capabilities?.dynamic_schema === true,
  },
  streams,
  write_actions: (writeBundle.actions ?? []).map((action) => ({
    name: trim(action.name),
    method: trim(action.method).toUpperCase(),
    kind: trim(action.kind),
  })),
  cli_surface: mapCLISurface(cliSurface),
  docs_md: docsMd,
  icon: existing.icon,
};

websiteData[websiteData.indexOf(existing)] = replacement;
writeFileSync(websiteDataPath, JSON.stringify(websiteData, null, 2), 'utf8');

const websiteCatalog = readJSON(websiteCatalogPath);
const catalogEntry = websiteCatalog.find((item) => item.slug === 'ashby');
if (!catalogEntry) throw new Error('ashby is missing from website connector catalog');
catalogEntry.streams = streams.map((stream) => ({
  name: stream.name,
  primaryKey: stream.primary_key,
  cursor: stream.cursor,
  incremental: stream.incremental,
}));
catalogEntry.cliSurface = mapCLISurface(cliSurface, { keyStyle: 'camel' });
catalogEntry.docsMd = docsMd.trim();
writeFileSync(websiteCatalogPath, `${JSON.stringify(websiteCatalog)}\n`, 'utf8');

function readJSON(path) {
  return JSON.parse(readFileSync(path, 'utf8'));
}

function trim(value) {
  return typeof value === 'string' ? value.trim() : '';
}
