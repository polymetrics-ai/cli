// Generates website/data/connectors.generated.json from the connector bundles.
// Run: node scripts/gen-connector-bundles.mjs
// Reads: internal/connectors/defs/<name>/{metadata.json,streams.json,writes.json,docs.md,cli_surface.json?}
// Emits: website/data/connectors.generated.json

import {
  copyFileSync,
  existsSync,
  readFileSync,
  writeFileSync,
  mkdirSync,
  readdirSync,
} from 'node:fs';
import { dirname, relative, resolve, join, sep } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const DEFS_ROOT = resolve(__dirname, '../../internal/connectors/defs');
const ICON_DATA = resolve(__dirname, '../../internal/connectors/icon_data.json');
const ICON_OVERRIDES = resolve(__dirname, '../data/icon_overrides.json');
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

function stripPrefix(name) {
  return String(name || '').replace(/^(source|destination)-/, '');
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

function mapFlags(flags) {
  return (Array.isArray(flags) ? flags : [])
    .map((flag) => ({
      name: trim(flag.name),
      type: trim(flag.type),
      summary: trim(flag.summary),
      values: Array.isArray(flag.values) ? flag.values.map((value) => trim(value)).filter(Boolean) : [],
      maps_to: trim(flag.maps_to),
    }))
    .filter((flag) => flag.name);
}

function mapCLISurface(surface) {
  if (!surface || typeof surface !== 'object') return null;
  const commands = (Array.isArray(surface.commands) ? surface.commands : [])
    .map((command) => ({
      path: trim(command.path),
      summary: trim(command.summary),
      intent: trim(command.intent),
      availability: trim(command.availability),
      stream: trim(command.stream),
      write: trim(command.write),
      source_cli_path: trim(command.source_cli_path),
      source_url: trim(command.source_url),
      flags: mapFlags(command.flags),
      examples: Array.isArray(command.examples) ? command.examples.map((example) => trim(example)).filter(Boolean) : [],
      output_policy: trim(command.output_policy),
      risk: trim(command.risk),
      approval: trim(command.approval),
      notes: trim(command.notes),
    }))
    .filter((command) => command.path);

  if (!trim(surface.usage) && commands.length === 0) return null;

  return {
    tagline: trim(surface.tagline),
    usage: trim(surface.usage),
    source_cli: surface.source_cli
      ? {
          name: trim(surface.source_cli.name),
          docs: trim(surface.source_cli.docs),
          reference: trim(surface.source_cli.reference),
          source: trim(surface.source_cli.source),
        }
      : null,
    groups: (Array.isArray(surface.groups) ? surface.groups : [])
      .map((group) => ({
        id: trim(group.id),
        title: trim(group.title),
        commands: Array.isArray(group.commands) ? group.commands.map((command) => trim(command)).filter(Boolean) : [],
      }))
      .filter((group) => group.id || group.title || group.commands.length > 0),
    global_flags: mapFlags(surface.global_flags),
    commands,
    help_topics: (Array.isArray(surface.help_topics) ? surface.help_topics : [])
      .map((topic) => ({
        name: trim(topic.name),
        summary: trim(topic.summary),
      }))
      .filter((topic) => topic.name),
  };
}

function assertInside(root, target, label) {
  const rel = relative(root, target);
  if (rel.startsWith('..') || rel === '..' || rel.includes(`..${sep}`) || rel === '') {
    throw new Error(`${label} escapes expected root: ${target}`);
  }
}

function readSchema(base, schemaPath) {
  if (!trim(schemaPath)) return null;
  const target = resolve(base, schemaPath);
  assertInside(base, target, 'stream schema');
  return readJSON(target);
}

const validIconPath = (path) => /^icons\/(?:[A-Za-z0-9._-]+\/)?[A-Za-z0-9._-]+\.svg$/.test(path);
const iconRaw = readJSON(ICON_DATA) ?? [];
const iconOverrideRaw = readJSON(ICON_OVERRIDES) ?? [];
const iconByConnector = new Map();
for (const icon of Array.isArray(iconRaw) ? iconRaw : []) {
  if (!icon?.connector) continue;
  iconByConnector.set(icon.connector, icon);
  iconByConnector.set(stripPrefix(icon.connector), icon);
}

const iconOverrideByConnector = new Map();
for (const icon of Array.isArray(iconOverrideRaw) ? iconOverrideRaw : []) {
  if (!icon?.connector) continue;
  iconOverrideByConnector.set(icon.connector, icon);
  iconOverrideByConnector.set(stripPrefix(icon.connector), icon);
}

const copiedIconPaths = new Set();
for (const icon of Array.isArray(iconRaw) ? iconRaw : []) {
  const path = trim(icon?.path);
  if (path && validIconPath(path)) {
    copiedIconPaths.add(path);
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

function syncConnectorIcons(paths) {
  const outDir = resolve(ICON_PUBLIC_ROOT, 'icons');

  for (const iconPath of [...paths].sort()) {
    const src = resolveIconPath(ICON_SOURCE_ROOT, iconPath, 'connector icon source');
    if (!existsSync(src)) {
      throw new Error(`Missing connector icon asset: ${iconPath}`);
    }
  }

  mkdirSync(outDir, { recursive: true });

  for (const iconPath of [...paths].sort()) {
    const src = resolveIconPath(ICON_SOURCE_ROOT, iconPath, 'connector icon source');
    const out = resolveIconPath(ICON_PUBLIC_ROOT, iconPath, 'connector icon output');
    mkdirSync(dirname(out), { recursive: true });
    copyFileSync(src, out);
  }
}

function mapIcon(slug, metadata) {
  const candidates = [
    slug,
    metadata.name,
    stripPrefix(metadata.name),
    `source-${slug}`,
    `destination-${slug}`,
  ].filter(Boolean);

  const override = candidates.map((candidate) => iconOverrideByConnector.get(candidate)).find(Boolean);
  if (override) {
    const path = trim(override.path);
    if (!validIconPath(path)) {
      throw new Error(`Invalid connector icon override path for ${slug}: ${path}`);
    }

    return {
      id: trim(override.id),
      path,
      publicPath: `/connectors/${path}`,
      source: trim(override.source),
      reviewStatus: trim(override.review_status),
      reviewUrl: trim(override.review_url),
    };
  }

  const icon = candidates.map((candidate) => iconByConnector.get(candidate)).find(Boolean);
  if (!icon) return null;

  const path = trim(icon.path);
  if (!validIconPath(path)) {
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

  const slug = stripPrefix(metadata.name || dirName);
  if (!slug) {
    throw new Error(`Connector bundle has empty name: ${dirName}`);
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
    icon: mapIcon(slug, metadata),
  });
}

// Sort alphabetically by name
connectors.sort((a, b) => a.name.localeCompare(b.name, 'en', { sensitivity: 'base' }));

syncConnectorIcons(copiedIconPaths);

mkdirSync(dirname(OUT), { recursive: true });
writeFileSync(OUT, JSON.stringify(connectors, null, 2), 'utf8');

console.log(
  `Wrote ${connectors.length} connectors to data/connectors.generated.json; ` +
    `${copiedIconPaths.size} icons copied.`,
);

// Report which connectors have write actions
const withWrites = connectors.filter(c => c.write_actions.length > 0);
console.log(`Connectors with write actions: ${withWrites.length}`);
