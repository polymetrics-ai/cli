import {
  copyFileSync,
  existsSync,
  mkdirSync,
  rmSync,
} from 'node:fs';
import {
  dirname,
  isAbsolute,
  posix,
  relative,
  resolve,
  sep,
} from 'node:path';

export const validConnectorIconPath = (value) =>
  typeof value === 'string' &&
  !posix.isAbsolute(value) &&
  posix.normalize(value) === value &&
  /^icons\/(?:[A-Za-z0-9._-]+\/)?[A-Za-z0-9._-]+\.svg$/.test(value);

export function collectConnectorIconPaths(entries) {
  const paths = new Set();
  for (const entry of entries) {
    const connector = typeof entry?.connector === 'string'
      ? entry.connector.trim()
      : '';
    const iconPath = entry?.path;
    if (!validConnectorIconPath(iconPath)) {
      throw new Error(
        `Invalid connector icon path for ${connector || '<unknown>'}: ${iconPath ?? ''}`,
      );
    }
    paths.add(iconPath);
  }
  return paths;
}

export function assertInside(root, target, label) {
  const rel = relative(root, target);
  if (rel === '..' || rel.startsWith(`..${sep}`) || rel === '' || isAbsolute(rel)) {
    throw new Error(`${label} escapes expected root: ${target}`);
  }
}

function resolveContainedPath(root, path, label) {
  const target = resolve(root, path);
  assertInside(root, target, label);
  return target;
}

function resolveIconPath(root, iconPath, label) {
  if (!validConnectorIconPath(iconPath)) {
    throw new Error(`Invalid connector icon path: ${iconPath}`);
  }
  return resolveContainedPath(root, iconPath, label);
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
    const output = resolveContainedPath(
      outputRoot,
      iconPath.slice('icons/'.length),
      'connector icon output',
    );
    mkdirSync(dirname(output), { recursive: true });
    copyFileSync(source, output);
  }
}
