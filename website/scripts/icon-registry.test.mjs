import assert from 'node:assert/strict';
import {
  existsSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  readdirSync,
  rmSync,
  writeFileSync,
} from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join, relative, resolve, sep } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

import {
  collectConnectorIconPaths,
  syncConnectorIcons,
  validConnectorIconPath,
} from './lib/connector-icons.mjs';
import {
  resolveSimpleIconRequest,
  validSimpleIconSlug,
} from './lib/simple-icons.mjs';

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

function listSVGPaths(root, current = root) {
  if (!existsSync(current)) return [];
  const paths = [];
  for (const entry of readdirSync(current, { withFileTypes: true })) {
    const target = join(current, entry.name);
    if (entry.isDirectory()) {
      paths.push(...listSVGPaths(root, target));
    } else if (entry.isFile() && entry.name.endsWith('.svg')) {
      paths.push(relative(root, target).split(sep).join('/'));
    }
  }
  return paths.sort();
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

  const canonicalPaths = new Set(listSVGPaths(resolve(docsRoot, 'icons')));
  for (const outputPath of listSVGPaths(resolve(websiteRoot, 'icons'))) {
    assert.ok(canonicalPaths.has(outputPath), `website-only generated icon must be removed: icons/${outputPath}`);
  }
});

test('icon sync removes stale output after canonical path changes', (t) => {
  const fixtureRoot = mkdtempSync(join(tmpdir(), 'polymetrics-icon-sync-'));
  t.after(() => rmSync(fixtureRoot, { recursive: true, force: true }));

  const sourceRoot = resolve(fixtureRoot, 'docs/connectors');
  const publicRoot = resolve(fixtureRoot, 'website/public/connectors');
  const originalSource = resolve(sourceRoot, 'icons/original.svg');
  mkdirSync(dirname(originalSource), { recursive: true });
  writeFileSync(originalSource, '<svg xmlns="http://www.w3.org/2000/svg"></svg>');

  syncConnectorIcons(new Set(['icons/original.svg']), { sourceRoot, publicRoot });
  assert.equal(existsSync(resolve(publicRoot, 'icons/original.svg')), true);

  rmSync(originalSource);
  const replacementSource = resolve(sourceRoot, 'icons/simple-icons/replacement.svg');
  mkdirSync(dirname(replacementSource), { recursive: true });
  writeFileSync(replacementSource, '<svg xmlns="http://www.w3.org/2000/svg"></svg>');

  syncConnectorIcons(new Set(['icons/simple-icons/replacement.svg']), { sourceRoot, publicRoot });
  assert.equal(existsSync(resolve(publicRoot, 'icons/original.svg')), false);
  assert.equal(existsSync(resolve(publicRoot, 'icons/simple-icons/replacement.svg')), true);
});

test('icon sync rejects escaped paths before mutating generated output', (t) => {
  const fixtureRoot = mkdtempSync(join(tmpdir(), 'polymetrics-icon-boundary-'));
  t.after(() => rmSync(fixtureRoot, { recursive: true, force: true }));

  const sourceRoot = resolve(fixtureRoot, 'docs/connectors');
  const publicRoot = resolve(fixtureRoot, 'website/public/connectors');
  const outsideSource = resolve(sourceRoot, 'outside.svg');
  const insideSource = resolve(sourceRoot, 'icons/outside.svg');
  const outsideOutput = resolve(publicRoot, 'outside.svg');
  const retainedOutput = resolve(publicRoot, 'icons/retained.svg');
  for (const path of [outsideSource, insideSource, outsideOutput, retainedOutput]) {
    mkdirSync(dirname(path), { recursive: true });
    writeFileSync(path, path.endsWith('outside.svg') ? 'outside' : 'retained');
  }

  for (const iconPath of [
    'icons/../outside.svg',
    'icons/./outside.svg',
    'icons//outside.svg',
    'icons/simple-icons/../outside.svg',
    '../icons/outside.svg',
    '/icons/outside.svg',
    'icons\\outside.svg',
  ]) {
    assert.equal(validConnectorIconPath(iconPath), false, `${iconPath} must be rejected`);
    assert.throws(
      () => syncConnectorIcons(new Set([iconPath]), { sourceRoot, publicRoot }),
      /Invalid connector icon path/,
    );
    assert.equal(readFileSync(outsideOutput, 'utf8'), 'outside');
    assert.equal(readFileSync(retainedOutput, 'utf8'), 'retained');
  }
});

test('registry path collection rejects invalid unimplemented rows', () => {
  assert.throws(
    () => collectConnectorIconPaths([
      {
        connector: 'registry-only',
        implemented: false,
        path: 'icons/../outside.svg',
      },
    ]),
    /Invalid connector icon path for registry-only/,
  );

  assert.deepEqual(
    [...collectConnectorIconPaths([
      {
        connector: 'registry-only',
        implemented: false,
        path: 'icons/registry-only.svg',
      },
    ])],
    ['icons/registry-only.svg'],
  );
});

test('website scripts consume only the canonical registry', () => {
  const bundleScript = readFileSync(resolve(__dirname, 'gen-connector-bundles.mjs'), 'utf8');
  assert.equal(bundleScript.includes('icon_overrides'), false, 'bundle generator must not read website overrides');
  assert.equal(bundleScript.includes('stripPrefix'), false, 'bundle generator must not strip legacy prefixes');

  const fetchScript = readFileSync(resolve(__dirname, 'fetch-simple-icons.mjs'), 'utf8');
  assert.match(fetchScript, /internal\/connectors\/icon_data\.json/);
  assert.match(fetchScript, /docs\/connectors/);
  assert.equal(fetchScript.includes('icon_overrides'), false, 'Simple Icons fetcher must read canonical registry only');
  assert.match(
    fetchScript,
    /resolveSimpleIconRequest/,
    'Simple Icons fetcher must validate the slug and output path before any fetch or write',
  );
});

test('Simple Icons slug validation accepts only bare lowercase alphanumeric identifiers', () => {
  assert.equal(validSimpleIconSlug('github'), true);
  assert.equal(validSimpleIconSlug('1password'), true);
  assert.equal(validSimpleIconSlug(''), false, 'empty slug must be rejected');
  assert.equal(validSimpleIconSlug('foo/bar'), false, 'slug with a path separator must be rejected');
  assert.equal(validSimpleIconSlug('https://evil.example/x'), false, 'slug with a scheme must be rejected');
  assert.equal(validSimpleIconSlug('foo bar'), false, 'slug with whitespace must be rejected');
  assert.equal(validSimpleIconSlug('foo.bar'), false, 'slug with a dot must be rejected');
  assert.equal(validSimpleIconSlug('Foo'), false, 'slug with uppercase must be rejected');
  assert.equal(validSimpleIconSlug(undefined), false, 'non-string slug must be rejected');
});

test('Simple Icons request resolution enforces the CodeQL-flagged input boundaries', () => {
  const root = resolve(repoRoot, 'docs/connectors');
  const validIcon = { slug: 'github', path: 'icons/simple-icons/github.svg' };

  assert.throws(
    () => resolveSimpleIconRequest(root, { ...validIcon, path: '../outside.svg' }),
    /escapes/,
    '../-traversing path must be rejected before anything is written',
  );
  assert.throws(
    () => resolveSimpleIconRequest(root, { ...validIcon, path: 'icons/simple-icons/../../../etc/passwd.svg' }),
    /escapes/,
    'nested ../-traversing path must be rejected before anything is written',
  );
  assert.throws(
    () => resolveSimpleIconRequest(root, { ...validIcon, path: '/etc/passwd' }),
    /escapes/,
    'absolute path must be rejected',
  );
  assert.throws(
    () => resolveSimpleIconRequest(root, { ...validIcon, path: 'index.md' }),
    /must stay under icons\/simple-icons\//,
    'in-tree non-icon path must be rejected before anything is written',
  );
  assert.throws(
    () => resolveSimpleIconRequest(root, { ...validIcon, path: 'icons/github.svg' }),
    /must stay under icons\/simple-icons\//,
    'in-tree curated icon path must be rejected before anything is written',
  );
  assert.throws(
    () => resolveSimpleIconRequest(root, { ...validIcon, path: undefined }),
    /Invalid Simple Icons path/,
    'missing path must be rejected before anything is written',
  );
  assert.throws(
    () => resolveSimpleIconRequest(root, { ...validIcon, slug: 'foo/bar' }),
    /Invalid Simple Icons slug/,
    'slug containing / must be rejected before any fetch',
  );
  assert.throws(
    () => resolveSimpleIconRequest(root, { ...validIcon, slug: 'https://evil.example/x' }),
    /Invalid Simple Icons slug/,
    'slug containing a scheme must be rejected before any fetch',
  );
  assert.throws(
    () => resolveSimpleIconRequest(root, { ...validIcon, slug: '' }),
    /Invalid Simple Icons slug/,
    'empty slug must be rejected',
  );

  const { url, outputPath } = resolveSimpleIconRequest(root, validIcon);
  assert.equal(url, 'https://cdn.simpleicons.org/github', 'valid bare slug must resolve unchanged');
  assert.equal(outputPath, resolve(root, validIcon.path), 'valid in-tree path must resolve unchanged');
});
