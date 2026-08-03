// Fetches vetted Simple Icons SVGs listed in the canonical connector icon registry.
// Run: node scripts/fetch-simple-icons.mjs
//
// The Simple Icons list is intentionally curated in internal/connectors/icon_data.json.
// Do not infer icons from arbitrary docs hosts such as GitHub, ReadMe, or Apiary:
// that produces false brand matches.

import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import {
  readSimpleIconsLockfile,
  resolveSimpleIconRequest,
  sha256Hex,
  validSimpleIconPath,
  validSimpleIconSlug,
  verifyFetchedIconDigest,
  writeSimpleIconsLockfile,
} from './lib/simple-icons.mjs';

const __dirname = dirname(fileURLToPath(import.meta.url));
const ICON_DATA = resolve(__dirname, '../../internal/connectors/icon_data.json');
const DOCS_CONNECTORS = resolve(__dirname, '../../docs/connectors');
const LOCKFILE = resolve(__dirname, '../data/simple-icons.lock.json');
const UPDATE_LOCKFILE = process.argv.includes('--update-lockfile');

function fail(message) {
  console.error(`fetch-simple-icons: ${message}`);
  process.exit(1);
}

function tintSvg(svg, hex) {
  const clean = svg.replace(/<title>.*?<\/title>/s, '');
  const color = String(hex || '').replace(/^#/, '');
  if (!/^[0-9A-Fa-f]{6}$/.test(color)) return clean;
  const openTag = clean.slice(0, clean.indexOf('>') + 1);
  if (/\sfill=/.test(openTag)) return clean;
  return clean.replace('<svg ', `<svg fill="#${color.toUpperCase()}" `);
}

const registry = JSON.parse(readFileSync(ICON_DATA, 'utf8'));
if (!Array.isArray(registry)) fail('canonical icon registry must be an array');

const seenConnectors = new Set();
const seenPaths = new Set();
const simpleIcons = [];
for (const icon of registry) {
  const connector = String(icon.connector || '').trim();
  const slug = String(icon.simple_icon_slug || '').trim();
  const path = String(icon.path || '').trim();
  if (!connector || /^(source|destination)-/.test(connector)) {
    fail(`invalid connector key: ${connector}`);
  }
  if (seenConnectors.has(connector)) {
    fail(`duplicate connector key: ${connector}`);
  }
  seenConnectors.add(connector);
  if (!slug && icon.source !== 'simple-icons') continue;
  if (!slug || !validSimpleIconSlug(slug)) fail(`invalid simple_icon_slug for ${connector}: ${slug}`);
  if (!validSimpleIconPath(path)) fail(`invalid Simple Icons path for ${connector}: ${path}`);
  if (seenPaths.has(path)) fail(`duplicate Simple Icons path: ${path}`);
  seenPaths.add(path);
  simpleIcons.push({ connector, slug, path, hex: icon.simple_icon_hex });
}

const lockfile = UPDATE_LOCKFILE ? {} : readSimpleIconsLockfile(LOCKFILE);

let written = 0;
for (const icon of simpleIcons) {
  let request;
  try {
    request = resolveSimpleIconRequest(DOCS_CONNECTORS, icon);
  } catch (error) {
    fail(`${icon.connector}: ${error.message}`);
    throw error;
  }
  const { url, outputPath } = request;

  const response = await fetch(url);
  if (!response.ok) {
    fail(`could not fetch ${icon.slug}: HTTP ${response.status}`);
  }

  const svg = await response.text();
  if (!svg.trim().startsWith('<svg') || /<script/i.test(svg)) {
    fail(`unexpected SVG payload for ${icon.slug}`);
  }

  if (UPDATE_LOCKFILE) {
    lockfile[icon.connector] = { slug: icon.slug, sha256: sha256Hex(svg) };
  } else {
    try {
      verifyFetchedIconDigest(lockfile, icon, svg);
    } catch (error) {
      fail(`${icon.connector}: ${error.message}`);
      throw error;
    }
  }

  mkdirSync(dirname(outputPath), { recursive: true });
  writeFileSync(outputPath, tintSvg(svg, icon.hex), 'utf8');
  written += 1;
}

if (UPDATE_LOCKFILE) {
  writeSimpleIconsLockfile(LOCKFILE, lockfile);
  console.log(`Updated Simple Icons lockfile with ${Object.keys(lockfile).length} connector entries.`);
}

console.log(`Fetched ${written} Simple Icons SVGs into docs/connectors/icons.`);
