import assert from 'node:assert/strict';
import { existsSync, readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const repoRoot = resolve(__dirname, '../..');
const iconDataPath = resolve(repoRoot, 'internal/connectors/icon_data.json');
const overridesPath = resolve(repoRoot, 'website/data/icon_overrides.json');
const docsRoot = resolve(repoRoot, 'docs/connectors');
const websiteRoot = resolve(repoRoot, 'website/public/connectors');

function readJSON(path) {
  return JSON.parse(readFileSync(path, 'utf8'));
}

function byConnector(entries) {
  return new Map(entries.map((entry) => [entry.connector, entry]));
}

test('canonical icon registry uses bare keys and owns source assets', () => {
  const entries = readJSON(iconDataPath);
  assert.ok(Array.isArray(entries), 'icon_data.json must be an array');
  const seen = new Set();
  for (const entry of entries) {
    assert.ok(entry.connector, 'entry must have a connector key');
    assert.equal(/^(source|destination)-/.test(entry.connector), false, `${entry.connector} must be bare`);
    assert.equal(seen.has(entry.connector), false, `${entry.connector} must not be duplicated`);
    seen.add(entry.connector);
    assert.match(entry.path, /^icons\/(?:[A-Za-z0-9._-]+\/)?[A-Za-z0-9._-]+\.svg$/);
    assert.ok(existsSync(resolve(docsRoot, entry.path)), `${entry.connector} canonical asset missing: ${entry.path}`);
  }

  const icons = byConnector(entries);
  assert.equal(icons.get('apify-dataset')?.path, 'icons/apify.svg');
  assert.equal(icons.get('apple-search-ads')?.path, 'icons/simple-icons/apple.svg');
});

test('website output has no icon mapping or SVG authority outside docs source tree', () => {
  assert.equal(existsSync(overridesPath), false, 'website/data/icon_overrides.json must not be an authored registry');
  for (const entry of readJSON(iconDataPath)) {
    assert.ok(existsSync(resolve(docsRoot, entry.path)), `canonical source missing for ${entry.path}`);
    assert.ok(existsSync(resolve(websiteRoot, entry.path)), `generated website copy missing for ${entry.path}`);
  }
});

test('website scripts consume only the canonical registry', () => {
  const bundleScript = readFileSync(resolve(__dirname, 'gen-connector-bundles.mjs'), 'utf8');
  assert.equal(bundleScript.includes('icon_overrides'), false, 'bundle generator must not read website overrides');
  assert.equal(bundleScript.includes('stripPrefix'), false, 'bundle generator must not strip legacy prefixes');

  const fetchScript = readFileSync(resolve(__dirname, 'fetch-simple-icons.mjs'), 'utf8');
  assert.match(fetchScript, /internal\/connectors\/icon_data\.json/);
  assert.match(fetchScript, /docs\/connectors/);
  assert.equal(fetchScript.includes('icon_overrides'), false, 'Simple Icons fetcher must read canonical registry only');
});
