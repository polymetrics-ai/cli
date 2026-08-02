import { resolve } from 'node:path';

import { assertInside } from './connector-icons.mjs';

const SIMPLE_ICON_SLUG_PATTERN = /^[a-z0-9]+$/;
const SIMPLE_ICON_CDN = 'https://cdn.simpleicons.org';

export function validSimpleIconSlug(value) {
  return typeof value === 'string' && SIMPLE_ICON_SLUG_PATTERN.test(value);
}

// Validates the two registry-authored inputs a Simple Icons fetch depends on
// before either reaches an outbound network request or a filesystem write:
// the slug that builds the CDN URL, and the path that builds the write
// destination. Throws before returning either value on any invalid input.
export function resolveSimpleIconRequest(docsConnectorsRoot, icon) {
  const slug = icon?.slug;
  if (!validSimpleIconSlug(slug)) {
    throw new Error(`Invalid Simple Icons slug: ${slug}`);
  }
  const outputPath = resolve(docsConnectorsRoot, icon.path);
  assertInside(docsConnectorsRoot, outputPath, 'Simple Icons output path');
  return { url: `${SIMPLE_ICON_CDN}/${slug}`, outputPath };
}
