import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
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
  assertLockfileCoversTargets,
  collectSimpleIconTargets,
  readSimpleIconsLockfile,
  resolveSimpleIconRequest,
  sha256Hex,
  validSimpleIconSlug,
  verifyFetchedIconDigest,
  writeSimpleIconsLockfile,
  writeVerifiedSimpleIcon,
} from './lib/simple-icons.mjs';

function digestOf(content) {
  return createHash('sha256').update(content, 'utf8').digest('hex');
}

const __dirname = dirname(fileURLToPath(import.meta.url));
const repoRoot = resolve(__dirname, '../..');
const iconDataPath = resolve(repoRoot, 'internal/connectors/icon_data.json');
const overridesPath = resolve(repoRoot, 'website/data/icon_overrides.json');
const docsRoot = resolve(repoRoot, 'docs/connectors');
const websiteRoot = resolve(repoRoot, 'website/public/connectors');
const lockfilePath = resolve(repoRoot, 'website/data/simple-icons.lock.json');

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
  assert.match(
    fetchScript,
    /writeVerifiedSimpleIcon/,
    'Simple Icons fetcher must write fetched content only through the digest-verified write helper',
  );
  assert.equal(
    fetchScript.includes('writeFileSync'),
    false,
    'Simple Icons fetcher must not own an SVG write sink that can bypass digest verification',
  );
  assert.match(
    fetchScript,
    /assertLockfileCoversTargets/,
    'Simple Icons fetcher must reject registry/lockfile drift before issuing any request',
  );
});

test('Simple Icons digest verification accepts content matching the recorded connector entry', () => {
  const content = '<svg>github</svg>';
  const lockfile = { github: { slug: 'github', sha256: digestOf(content) } };
  assert.doesNotThrow(() => verifyFetchedIconDigest(lockfile, { connector: 'github', slug: 'github' }, content));
});

test('Simple Icons digest verification rejects tampered content before anything is written', () => {
  const content = '<svg>github</svg>';
  const lockfile = { github: { slug: 'github', sha256: digestOf(content) } };
  assert.throws(
    () => verifyFetchedIconDigest(lockfile, { connector: 'github', slug: 'github' }, '<svg>tampered</svg>'),
    /content mismatch/,
    'a tampered body must be rejected before mkdirSync/writeFileSync are ever reached',
  );
});

test('Simple Icons digest verification rejects a fetched connector with no lockfile entry', () => {
  assert.throws(
    () => verifyFetchedIconDigest({}, { connector: 'github', slug: 'github' }, '<svg>github</svg>'),
    /missing Simple Icons lockfile entry/,
    'an unpinned fetch must never quietly succeed',
  );
});

test('Simple Icons digest verification checks connectors sharing one icon independently', () => {
  const content = '<svg>shared</svg>';
  const sharedLockfile = {
    'zoho-books': { slug: 'zoho', sha256: digestOf(content) },
    'zoho-desk': { slug: 'zoho', sha256: digestOf(content) },
  };
  // Duplicate digests under different connector keys are expected and legitimate; do not dedupe.
  assert.doesNotThrow(() => verifyFetchedIconDigest(sharedLockfile, { connector: 'zoho-books', slug: 'zoho' }, content));
  assert.doesNotThrow(() => verifyFetchedIconDigest(sharedLockfile, { connector: 'zoho-desk', slug: 'zoho' }, content));

  const oneStaleLockfile = {
    'zoho-books': { slug: 'zoho', sha256: digestOf(content) },
    'zoho-desk': { slug: 'zoho', sha256: digestOf('<svg>old</svg>') },
  };
  assert.doesNotThrow(
    () => verifyFetchedIconDigest(oneStaleLockfile, { connector: 'zoho-books', slug: 'zoho' }, content),
    'a sibling connector verifying correctly must not mask the other',
  );
  assert.throws(
    () => verifyFetchedIconDigest(oneStaleLockfile, { connector: 'zoho-desk', slug: 'zoho' }, content),
    /zoho-desk/,
    'one connector failing must name that connector and not silently pass',
  );
});

test('Simple Icons lockfile pins exactly the connectors the fetcher would fetch', () => {
  const targets = collectSimpleIconTargets(readJSON(iconDataPath));
  const lockfile = readSimpleIconsLockfile(lockfilePath);
  assert.ok(targets.length > 0, 'canonical registry must declare Simple Icons fetch targets');

  assert.deepEqual(
    Object.keys(lockfile).sort(),
    targets.map((target) => target.connector).sort(),
    'website/data/simple-icons.lock.json must pin every Simple Icons connector in the canonical registry and nothing else',
  );
  assert.doesNotThrow(() => assertLockfileCoversTargets(lockfile, targets));

  for (const target of targets) {
    assert.equal(
      lockfile[target.connector].slug,
      target.slug,
      `${target.connector} lockfile slug must match the canonical registry slug`,
    );
    assert.match(
      lockfile[target.connector].sha256,
      /^[0-9a-f]{64}$/,
      `${target.connector} must pin a sha256 digest`,
    );
  }

  const digests = new Map();
  for (const [connector, entry] of Object.entries(lockfile)) {
    digests.set(entry.sha256, [...(digests.get(entry.sha256) ?? []), connector]);
  }
  assert.ok(
    [...digests.values()].some((connectors) => connectors.length > 1),
    'connectors sharing one upstream icon keep their own duplicate-digest entries rather than being deduplicated',
  );
});

test('Simple Icons lockfile coverage rejects registry drift in either direction', () => {
  const targets = [
    { connector: 'zoho-books', slug: 'zoho', path: 'icons/simple-icons/zoho.svg' },
    { connector: 'zoho-desk', slug: 'zoho', path: 'icons/simple-icons/zohodesk.svg' },
  ];
  const pinned = {
    'zoho-books': { slug: 'zoho', sha256: digestOf('<svg>zoho</svg>') },
    'zoho-desk': { slug: 'zoho', sha256: digestOf('<svg>zoho</svg>') },
  };
  assert.doesNotThrow(() => assertLockfileCoversTargets(pinned, targets));

  const missingOne = { ...pinned };
  delete missingOne['zoho-desk'];
  assert.throws(
    () => assertLockfileCoversTargets(missingOne, targets),
    /missing Simple Icons lockfile entries for connectors: zoho-desk/,
    'a registry connector added without a recorded digest must fail offline, not at fetch time',
  );

  assert.throws(
    () => assertLockfileCoversTargets({ ...pinned, 'zoho-crm': { slug: 'zoho', sha256: digestOf('x') } }, targets),
    /stale Simple Icons lockfile entries for connectors no longer fetched: zoho-crm/,
    'an entry left behind after its connector was removed must not be silently retained',
  );

  assert.throws(
    () => assertLockfileCoversTargets({ ...pinned, 'zoho-desk': { slug: 'zohodesk', sha256: digestOf('x') } }, targets),
    /slug mismatch for connector zoho-desk/,
    'a repointed registry slug must invalidate its recorded entry',
  );

  assert.throws(
    () => assertLockfileCoversTargets({ ...pinned, 'zoho-desk': { slug: 'zoho', sha256: 'not-a-digest' } }, targets),
    /no valid sha256 digest/,
    'a malformed digest must not count as a pin',
  );
});

test('Simple Icons fetch targets cover slug-bearing and simple-icons-sourced registry rows only', () => {
  assert.deepEqual(
    collectSimpleIconTargets([
      {
        connector: 'github',
        simple_icon_slug: 'github',
        simple_icon_hex: '181717',
        path: 'icons/simple-icons/github.svg',
      },
      { connector: 'apify-dataset', path: 'icons/apify.svg' },
    ]),
    [{ connector: 'github', slug: 'github', path: 'icons/simple-icons/github.svg', hex: '181717' }],
  );

  assert.throws(() => collectSimpleIconTargets({}), /must be an array/);
  assert.throws(
    () => collectSimpleIconTargets([
      { connector: 'source-github', simple_icon_slug: 'github', path: 'icons/simple-icons/github.svg' },
    ]),
    /invalid connector key/,
  );
  assert.throws(
    () => collectSimpleIconTargets([
      { connector: 'github', simple_icon_slug: 'github', path: 'icons/simple-icons/github.svg' },
      { connector: 'github', simple_icon_slug: 'github', path: 'icons/simple-icons/github.svg' },
    ]),
    /duplicate connector key/,
  );
  assert.throws(
    () => collectSimpleIconTargets([
      { connector: 'github', source: 'simple-icons', path: 'icons/simple-icons/github.svg' },
    ]),
    /invalid simple_icon_slug/,
    'a simple-icons-sourced row with no slug must be rejected, not skipped',
  );
  assert.throws(
    () => collectSimpleIconTargets([
      { connector: 'github', simple_icon_slug: 'github', path: 'icons/github.svg' },
    ]),
    /invalid Simple Icons path/,
  );
  assert.throws(
    () => collectSimpleIconTargets([
      { connector: 'zoho-books', simple_icon_slug: 'zoho', path: 'icons/simple-icons/zoho.svg' },
      { connector: 'zoho-desk', simple_icon_slug: 'zoho', path: 'icons/simple-icons/zoho.svg' },
    ]),
    /duplicate Simple Icons path/,
  );
});

test('verified Simple Icons write is the only sink and never reaches disk on a digest failure', (t) => {
  const dir = mkdtempSync(join(tmpdir(), 'polymetrics-icon-write-'));
  t.after(() => rmSync(dir, { recursive: true, force: true }));

  const outputPath = join(dir, 'icons/simple-icons/github.svg');
  const content = '<svg viewBox="0 0 24 24"><title>GitHub</title></svg>';
  const icon = { connector: 'github', slug: 'github', hex: '181717' };
  const lockfile = { github: { slug: 'github', sha256: digestOf(content) } };

  assert.throws(
    () => writeVerifiedSimpleIcon(lockfile, icon, '<svg>tampered</svg>', outputPath),
    /content mismatch/,
  );
  assert.equal(existsSync(dirname(outputPath)), false, 'a rejected body must not even create its output directory');

  assert.throws(
    () => writeVerifiedSimpleIcon({}, icon, content, outputPath),
    /missing Simple Icons lockfile entry/,
  );
  assert.equal(existsSync(outputPath), false, 'an unpinned connector must not produce output');

  writeVerifiedSimpleIcon(lockfile, icon, content, outputPath);
  assert.equal(
    readFileSync(outputPath, 'utf8'),
    '<svg fill="#181717" viewBox="0 0 24 24"></svg>',
    'verified content is tinted and written unchanged otherwise',
  );
});

test('Simple Icons lockfile round-trips sorted by connector for reviewable diffs', (t) => {
  const dir = mkdtempSync(join(tmpdir(), 'polymetrics-icon-lockfile-'));
  t.after(() => rmSync(dir, { recursive: true, force: true }));
  const path = join(dir, 'simple-icons.lock.json');

  assert.deepEqual(readSimpleIconsLockfile(path), {}, 'a missing lockfile file reads as empty, not an error');

  writeSimpleIconsLockfile(path, {
    zeta: { slug: 'zeta', sha256: digestOf('z') },
    alpha: { slug: 'alpha', sha256: digestOf('a') },
  });
  const contents = readFileSync(path, 'utf8');
  assert.ok(
    contents.indexOf('"alpha"') < contents.indexOf('"zeta"'),
    'entries must be sorted by connector key for a stable, reviewable diff',
  );
  assert.deepEqual(readSimpleIconsLockfile(path), {
    alpha: { slug: 'alpha', sha256: digestOf('a') },
    zeta: { slug: 'zeta', sha256: digestOf('z') },
  });
});

test('sha256Hex is deterministic for identical content and differs for different content', () => {
  assert.equal(sha256Hex('same'), sha256Hex('same'));
  assert.notEqual(sha256Hex('same'), sha256Hex('different'));
  assert.equal(sha256Hex('same'), digestOf('same'));
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
