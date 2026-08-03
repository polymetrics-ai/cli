import { createHash } from 'node:crypto';
import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';

import { assertInside } from './connector-icons.mjs';

const SIMPLE_ICON_SLUG_PATTERN = /^[a-z0-9]+$/;
const SIMPLE_ICON_PATH_PATTERN = /^icons\/simple-icons\/[A-Za-z0-9._-]+\.svg$/;
const SIMPLE_ICON_SHA256_PATTERN = /^[0-9a-f]{64}$/;
const SIMPLE_ICON_CDN = 'https://cdn.simpleicons.org';

export function validSimpleIconSlug(value) {
  return typeof value === 'string' && SIMPLE_ICON_SLUG_PATTERN.test(value);
}

export function validSimpleIconPath(value) {
  return typeof value === 'string' && SIMPLE_ICON_PATH_PATTERN.test(value);
}

// Validates the two registry-authored inputs a Simple Icons fetch depends on
// before either reaches an outbound network request or a filesystem write:
// the slug that builds the CDN URL, and the path that builds the write
// destination. The path must both stay inside the docs connector root and keep
// the icons/simple-icons/<name>.svg shape, so an in-tree but non-icon
// destination cannot be overwritten. Throws before returning either value on
// any invalid input.
export function resolveSimpleIconRequest(docsConnectorsRoot, icon) {
  const slug = icon?.slug;
  if (!validSimpleIconSlug(slug)) {
    throw new Error(`Invalid Simple Icons slug: ${slug}`);
  }
  const iconPath = icon?.path;
  if (typeof iconPath !== 'string') {
    throw new Error(`Invalid Simple Icons path: ${iconPath}`);
  }
  const outputPath = resolve(docsConnectorsRoot, iconPath);
  assertInside(docsConnectorsRoot, outputPath, 'Simple Icons output path');
  if (!validSimpleIconPath(iconPath)) {
    throw new Error(`Simple Icons path must stay under icons/simple-icons/: ${iconPath}`);
  }
  return { url: `${SIMPLE_ICON_CDN}/${slug}`, outputPath };
}

// Selects the canonical registry rows the Simple Icons fetcher will request and
// applies the connector-key, slug, and path rules the fetch/write boundary
// depends on. The fetcher and the website script tests both derive the fetch
// target set from here, so lockfile coverage can be checked against the registry
// without issuing a request.
export function collectSimpleIconTargets(registry) {
  if (!Array.isArray(registry)) {
    throw new Error('canonical icon registry must be an array');
  }
  const seenConnectors = new Set();
  const seenPaths = new Set();
  const targets = [];
  for (const row of registry) {
    const connector = String(row?.connector || '').trim();
    const slug = String(row?.simple_icon_slug || '').trim();
    const path = String(row?.path || '').trim();
    if (!connector || /^(source|destination)-/.test(connector)) {
      throw new Error(`invalid connector key: ${connector}`);
    }
    if (seenConnectors.has(connector)) {
      throw new Error(`duplicate connector key: ${connector}`);
    }
    seenConnectors.add(connector);
    if (!slug && row?.source !== 'simple-icons') continue;
    if (!validSimpleIconSlug(slug)) {
      throw new Error(`invalid simple_icon_slug for ${connector}: ${slug}`);
    }
    if (!validSimpleIconPath(path)) {
      throw new Error(`invalid Simple Icons path for ${connector}: ${path}`);
    }
    if (seenPaths.has(path)) {
      throw new Error(`duplicate Simple Icons path: ${path}`);
    }
    seenPaths.add(path);
    targets.push({ connector, slug, path, hex: row.simple_icon_hex });
  }
  return targets;
}

// Requires the lockfile to describe exactly the connectors the fetcher would
// request. A connector added to the registry with no recorded digest, an entry
// left behind after its connector was dropped, and an entry whose slug no longer
// matches the registry are all hard errors, so registry/lockfile drift fails
// offline in CI instead of surfacing later as an unpinned fetch or a silently
// retained stale entry.
export function assertLockfileCoversTargets(lockfile, targets) {
  const entries = lockfile ?? {};
  const expected = new Map(targets.map((target) => [target.connector, target]));
  const recorded = new Set(Object.keys(entries));

  const missing = [...expected.keys()].filter((connector) => !recorded.has(connector)).sort();
  if (missing.length) {
    throw new Error(
      `missing Simple Icons lockfile entries for connectors: ${missing.join(', ')}; ` +
        'run "node scripts/fetch-simple-icons.mjs --update-lockfile" to record them deliberately',
    );
  }

  const stale = [...recorded].filter((connector) => !expected.has(connector)).sort();
  if (stale.length) {
    throw new Error(
      `stale Simple Icons lockfile entries for connectors no longer fetched: ${stale.join(', ')}; ` +
        'run "node scripts/fetch-simple-icons.mjs --update-lockfile" to drop them deliberately',
    );
  }

  for (const [connector, target] of expected) {
    const entry = entries[connector];
    if (entry?.slug !== target.slug) {
      throw new Error(
        `Simple Icons lockfile slug mismatch for connector ${connector}: ` +
          `registry has ${target.slug}, lockfile has ${entry?.slug}`,
      );
    }
    if (!SIMPLE_ICON_SHA256_PATTERN.test(entry?.sha256 ?? '')) {
      throw new Error(`Simple Icons lockfile entry for connector ${connector} has no valid sha256 digest`);
    }
  }
}

export function sha256Hex(content) {
  return createHash('sha256').update(content, 'utf8').digest('hex');
}

export function tintSvg(svg, hex) {
  const clean = svg.replace(/<title>.*?<\/title>/s, '');
  const color = String(hex || '').replace(/^#/, '');
  if (!/^[0-9A-Fa-f]{6}$/.test(color)) return clean;
  const openTag = clean.slice(0, clean.indexOf('>') + 1);
  if (/\sfill=/.test(openTag)) return clean;
  return clean.replace('<svg ', `<svg fill="#${color.toUpperCase()}" `);
}

// Verifies fetched Simple Icons SVG content against a per-connector lockfile
// entry before it is ever written to disk, so a compromised or unexpectedly
// changed upstream response cannot reach the filesystem. The lockfile is
// keyed by connector, not by icon or slug: two connectors that legitimately
// share one icon (for example two Zoho product connectors sharing the
// "zoho" slug) each carry their own entry and each verify independently, so
// one connector's entry going stale never masks the other. A connector with
// no lockfile entry is rejected rather than silently allowed through.
export function verifyFetchedIconDigest(lockfile, icon, content) {
  const entry = lockfile[icon.connector];
  const digest = sha256Hex(content);
  if (!entry) {
    throw new Error(
      `missing Simple Icons lockfile entry for connector ${icon.connector} (slug ${icon.slug}); ` +
        'run "node scripts/fetch-simple-icons.mjs --update-lockfile" to record it deliberately',
    );
  }
  if (entry.sha256 !== digest) {
    throw new Error(
      `Simple Icons content mismatch for connector ${icon.connector} (slug ${icon.slug}): ` +
        `expected sha256 ${entry.sha256}, got ${digest}; if this is an intentional upstream icon ` +
        'update, run "node scripts/fetch-simple-icons.mjs --update-lockfile" to record it',
    );
  }
}

// The only path from fetched Simple Icons content to disk. Digest verification
// runs first and throws on failure, so a body that does not match its recorded
// per-connector entry never reaches mkdirSync or writeFileSync and no partial
// output directory is created for it.
export function writeVerifiedSimpleIcon(lockfile, icon, content, outputPath) {
  verifyFetchedIconDigest(lockfile, icon, content);
  mkdirSync(dirname(outputPath), { recursive: true });
  writeFileSync(outputPath, tintSvg(content, icon.hex), 'utf8');
}

export function readSimpleIconsLockfile(path) {
  if (!existsSync(path)) return {};
  return JSON.parse(readFileSync(path, 'utf8'));
}

export function writeSimpleIconsLockfile(path, lockfile) {
  const sorted = {};
  for (const connector of Object.keys(lockfile).sort()) {
    sorted[connector] = lockfile[connector];
  }
  writeFileSync(path, `${JSON.stringify(sorted, null, 2)}\n`, 'utf8');
}
