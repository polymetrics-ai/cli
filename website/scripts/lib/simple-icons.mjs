import { createHash } from 'node:crypto';
import { existsSync, readFileSync, writeFileSync } from 'node:fs';
import { resolve } from 'node:path';

import { assertInside } from './connector-icons.mjs';

const SIMPLE_ICON_SLUG_PATTERN = /^[a-z0-9]+$/;
const SIMPLE_ICON_PATH_PATTERN = /^icons\/simple-icons\/[A-Za-z0-9._-]+\.svg$/;
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

export function sha256Hex(content) {
  return createHash('sha256').update(content, 'utf8').digest('hex');
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
