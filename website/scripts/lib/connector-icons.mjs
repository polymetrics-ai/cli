import {
  copyFileSync,
  existsSync,
  mkdirSync,
  rmSync,
} from 'node:fs';
import { dirname, relative, resolve, sep } from 'node:path';

export const validConnectorIconPath = (value) =>
  /^icons\/(?:[A-Za-z0-9._-]+\/)?[A-Za-z0-9._-]+\.svg$/.test(value);

function assertInside(root, target, label) {
  const rel = relative(root, target);
  if (rel.startsWith('..') || rel === '..' || rel.includes(`..${sep}`) || rel === '') {
    throw new Error(`${label} escapes expected root: ${target}`);
  }
}

function resolveIconPath(root, iconPath, label) {
  if (!validConnectorIconPath(iconPath)) {
    throw new Error(`Invalid connector icon path: ${iconPath}`);
  }
  const target = resolve(root, iconPath);
  assertInside(root, target, label);
  return target;
}

export function syncConnectorIcons(paths, { sourceRoot, publicRoot }) {
  const sortedPaths = [...new Set(paths)].sort();
  for (const iconPath of sortedPaths) {
    const source = resolveIconPath(sourceRoot, iconPath, 'connector icon source');
    if (!existsSync(source)) {
      throw new Error(`Missing connector icon asset: ${iconPath}`);
    }
  }

  const outputRoot = resolve(publicRoot, 'icons');
  assertInside(publicRoot, outputRoot, 'connector icon output root');
  rmSync(outputRoot, { recursive: true, force: true });
  mkdirSync(outputRoot, { recursive: true });

  for (const iconPath of sortedPaths) {
    const source = resolveIconPath(sourceRoot, iconPath, 'connector icon source');
    const output = resolveIconPath(publicRoot, iconPath, 'connector icon output');
    mkdirSync(dirname(output), { recursive: true });
    copyFileSync(source, output);
  }
}
