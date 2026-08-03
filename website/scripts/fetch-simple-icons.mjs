// Fetches vetted Simple Icons SVGs listed in the canonical connector icon registry.
// Run: node scripts/fetch-simple-icons.mjs               fetch and write checksum-verified SVGs
//      node scripts/fetch-simple-icons.mjs --update-lockfile   re-pin digests, write no SVGs
//
// The Simple Icons list is intentionally curated in internal/connectors/icon_data.json.
// Do not infer icons from arbitrary docs hosts such as GitHub, ReadMe, or Apiary:
// that produces false brand matches.
//
// Fetched content is checksum-pinned per connector by website/data/simple-icons.lock.json;
// see docs/migration/icon-registry-single-source.md for the lockfile contract and the
// CodeQL alert #93 disposition.

import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

import {
  assertLockfileCoversTargets,
  collectSimpleIconTargets,
  readSimpleIconsLockfile,
  resolveSimpleIconRequest,
  sha256Hex,
  writeSimpleIconsLockfile,
  writeVerifiedSimpleIcon,
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

function failWith(error, prefix) {
  fail(prefix ? `${prefix}: ${error.message}` : error.message);
  throw error;
}

let simpleIcons;
try {
  simpleIcons = collectSimpleIconTargets(JSON.parse(readFileSync(ICON_DATA, 'utf8')));
} catch (error) {
  failWith(error);
}

const lockfile = UPDATE_LOCKFILE ? {} : readSimpleIconsLockfile(LOCKFILE);
if (!UPDATE_LOCKFILE) {
  try {
    assertLockfileCoversTargets(lockfile, simpleIcons);
  } catch (error) {
    failWith(error);
  }
}

let written = 0;
let recorded = 0;
for (const icon of simpleIcons) {
  let request;
  try {
    request = resolveSimpleIconRequest(DOCS_CONNECTORS, icon);
  } catch (error) {
    failWith(error, icon.connector);
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
    recorded += 1;
    continue;
  }

  try {
    writeVerifiedSimpleIcon(lockfile, icon, svg, outputPath);
  } catch (error) {
    failWith(error, icon.connector);
  }
  written += 1;
}

if (UPDATE_LOCKFILE) {
  writeSimpleIconsLockfile(LOCKFILE, lockfile);
  console.log(
    `Recorded ${recorded} Simple Icons connector digests in website/data/simple-icons.lock.json. ` +
      'No SVGs were written; re-run without --update-lockfile to fetch verified assets.',
  );
} else {
  console.log(`Fetched ${written} Simple Icons SVGs into docs/connectors/icons.`);
}
