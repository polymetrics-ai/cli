// Fetches vetted Simple Icons SVGs listed in website/data/icon_overrides.json.
// Run: node scripts/fetch-simple-icons.mjs
//
// The override list is intentionally curated. Do not infer icons from arbitrary
// docs hosts such as GitHub, ReadMe, or Apiary: that produces false brand matches.

import { mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const OVERRIDES = resolve(__dirname, '../data/icon_overrides.json');
const PUBLIC_CONNECTORS = resolve(__dirname, '../public/connectors');
const SIMPLE_ICON_CDN = 'https://cdn.simpleicons.org';

function fail(message) {
  console.error(`fetch-simple-icons: ${message}`);
  process.exit(1);
}

function validIconPath(path) {
  return /^icons\/simple-icons\/[A-Za-z0-9._-]+\.svg$/.test(path);
}

function tintSvg(svg, hex) {
  const clean = svg.replace(/<title>.*?<\/title>/s, '');
  const color = String(hex || '').replace(/^#/, '');
  if (!/^[0-9A-Fa-f]{6}$/.test(color)) return clean;
  const openTag = clean.slice(0, clean.indexOf('>') + 1);
  if (/\sfill=/.test(openTag)) return clean;
  return clean.replace('<svg ', `<svg fill="#${color.toUpperCase()}" `);
}

const overrides = JSON.parse(readFileSync(OVERRIDES, 'utf8'));
if (!Array.isArray(overrides)) fail('icon_overrides.json must be an array');

let written = 0;
for (const icon of overrides) {
  const slug = String(icon.simple_icon_slug || '').trim();
  const path = String(icon.path || '').trim();
  if (!slug || !/^[A-Za-z0-9._-]+$/.test(slug)) fail(`invalid simple_icon_slug: ${slug}`);
  if (!validIconPath(path)) fail(`invalid icon path for ${icon.connector}: ${path}`);

  const response = await fetch(`${SIMPLE_ICON_CDN}/${slug}`);
  if (!response.ok) {
    fail(`could not fetch ${slug}: HTTP ${response.status}`);
  }

  const svg = await response.text();
  if (!svg.trim().startsWith('<svg') || /<script/i.test(svg)) {
    fail(`unexpected SVG payload for ${slug}`);
  }

  const out = resolve(PUBLIC_CONNECTORS, path);
  mkdirSync(dirname(out), { recursive: true });
  writeFileSync(out, tintSvg(svg, icon.simple_icon_hex), 'utf8');
  written += 1;
}

console.log(`Fetched ${written} Simple Icons SVGs.`);
