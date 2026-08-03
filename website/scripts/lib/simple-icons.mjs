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
