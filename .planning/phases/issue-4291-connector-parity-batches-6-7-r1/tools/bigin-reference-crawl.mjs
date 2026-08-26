#!/usr/bin/env node

// Crawl Bigin's complete public v2 rendered reference. The sitemap is the
// provider's authoritative page inventory; progress is persisted after every
// successfully retrieved page so an interruption resumes without treating a
// partial inventory as a complete source lock.
import { createHash } from 'node:crypto';
import { existsSync, readFileSync, writeFileSync } from 'node:fs';

const sitemapURL = 'https://www.bigin.com/sitemap.xml';
const outputPath = new URL('../bigin-reference-crawl-progress.json', import.meta.url);
const parserVersion = 4;
const now = () => new Date().toISOString();
const writeJSON = value => writeFileSync(outputPath, `${JSON.stringify(value, null, 2)}\n`);
const sha256 = value => createHash('sha256').update(value).digest('hex');
const sleep = milliseconds => new Promise(resolve => setTimeout(resolve, milliseconds));

async function retrieve(url) {
  let lastError;
  for (let attempt = 1; attempt <= 4; attempt += 1) {
    try {
      const response = await fetch(url, { headers: { 'user-agent': 'Polymetrics connector source-lock recovery (+https://github.com/polymetrics-ai/cli)' } });
      const bytes = Buffer.from(await response.arrayBuffer());
      if (response.ok) return { bytes, status: response.status };
      lastError = new Error(`HTTP ${response.status}`);
      if (response.status !== 429 && response.status < 500) break;
    } catch (error) {
      lastError = error;
    }
    await sleep(1000 * 2 ** (attempt - 1));
  }
  throw new Error(`${url}: ${lastError?.message ?? 'unavailable'}`);
}

function endpointOperations(html) {
  const operations = new Map();
  const moduleExamples = new Set(['Accounts', 'Calls', 'Companies', 'Contacts', 'Deals', 'Events', 'Pipelines', 'Products', 'Tasks']);
  const text = html
    .replace(/<[^>]*>/g, ' ')
    .replace(/&nbsp;/g, ' ')
    .replace(/&amp;/g, '&')
    .replace(/\s+/g, ' ');
  const pattern = /\b(GET|POST|PUT|PATCH|DELETE)\s+(https:\/\/[^\s<>"']+\/bigin\/(?:bulk\/)?v2(?:\/[^\s<>"']*)?)/gi;
  for (const match of text.matchAll(pattern)) {
    const method = match[1].toUpperCase();
    const raw = match[2].replace(/(?:Copied!?|[.,;:])$/g, '');
    const marker = raw.includes('/bigin/bulk/v2') ? '/bigin/bulk/v2' : '/bigin/v2';
    const offset = raw.indexOf(marker);
    if (offset < 0) continue;
    const suffix = raw.slice(offset + marker.length) || '/';
    const path = `${marker === '/bigin/bulk/v2' ? marker : ''}${suffix}`
      .split(/[?#]/, 1)[0]
      .split('/')
      .map(segment => moduleExamples.has(segment) ? '{module_api_name}' : segment)
      .join('/');
    operations.set(`${method} ${path}`, { method, path });
  }
  return [...operations.values()].sort((left, right) => left.method.localeCompare(right.method) || left.path.localeCompare(right.path));
}

const sitemap = await retrieve(sitemapURL);
const sitemapText = sitemap.bytes.toString('utf8');
const urls = [...sitemapText.matchAll(/<loc>(.*?)<\/loc>/g)]
  .map(match => match[1])
  .filter(url => url.startsWith('https://www.bigin.com/developer/docs/apis/v2/'));
if (urls.length === 0) throw new Error(`${sitemapURL}: no Bigin v2 reference pages`);

let existing = existsSync(outputPath)
  ? JSON.parse(readFileSync(outputPath, 'utf8'))
  : { schema_version: 1, connector: 'zoho-bigin', pages: {} };
if (existing.connector !== 'zoho-bigin') throw new Error(`${outputPath.pathname}: wrong connector`);
if (existing.parser_version !== parserVersion) {
  existing = {
    schema_version: 1,
    connector: 'zoho-bigin',
    parser_version: parserVersion,
    cache_reset_reason: 'The previous extractor omitted the provider-rendered /bigin/bulk/v2 endpoint family; the current parser includes ordinary and bulk v2 operations before counting.',
    pages: {},
  };
}
const progress = {
  ...existing,
  schema_version: 1,
  connector: 'zoho-bigin',
  parser_version: parserVersion,
  sitemap: { url: sitemapURL, retrieved_at: now(), sha256: sha256(sitemap.bytes), bytes: sitemap.bytes.length, v2_pages_total: urls.length },
  pages: existing.pages ?? {},
  state: 'in_progress',
};
writeJSON(progress);

for (const [index, url] of urls.entries()) {
  if (progress.pages[url]?.status === 'retrieved') continue;
  try {
    const page = await retrieve(url);
    const html = page.bytes.toString('utf8');
    progress.pages[url] = {
      status: 'retrieved',
      retrieved_at: now(),
      http_status: page.status,
      sha256: sha256(page.bytes),
      bytes: page.bytes.length,
      operations: endpointOperations(html),
    };
    progress.resume_from = index + 1 < urls.length ? urls[index + 1] : null;
    progress.pages_retrieved = Object.values(progress.pages).filter(page => page.status === 'retrieved').length;
    progress.pages_total = urls.length;
    progress.coverage_confidence = {
      level: 'partial',
      basis: `Retrieved ${progress.pages_retrieved} of ${urls.length} provider-sitemap Bigin v2 reference pages; resume from ${progress.resume_from ?? 'completion'}.`,
    };
    writeJSON(progress);
    console.log(`bigin ${progress.pages_retrieved}/${urls.length}: ${url}`);
    await sleep(300);
  } catch (error) {
    progress.pages[url] = { status: 'failed', attempted_at: now(), error: error.message };
    progress.resume_from = url;
    progress.pages_retrieved = Object.values(progress.pages).filter(page => page.status === 'retrieved').length;
    progress.pages_total = urls.length;
    progress.coverage_confidence = {
      level: 'partial',
      basis: `Retrieved ${progress.pages_retrieved} of ${urls.length} provider-sitemap Bigin v2 reference pages; failed at ${url} and must resume from that page.`,
    };
    writeJSON(progress);
    throw error;
  }
}

const retrieved = urls.map(url => progress.pages[url]).filter(page => page?.status === 'retrieved');
if (retrieved.length !== urls.length) throw new Error(`Bigin crawl incomplete: ${retrieved.length}/${urls.length}`);
const operations = new Map();
for (const [url, page] of Object.entries(progress.pages)) {
  if (page.status !== 'retrieved') continue;
  for (const operation of page.operations) {
    const key = `${operation.method} ${operation.path}`;
    const current = operations.get(key);
    if (!current) operations.set(key, { ...operation, source_url: url });
  }
}
progress.state = 'complete';
progress.resume_from = null;
progress.pages_retrieved = urls.length;
progress.pages_total = urls.length;
progress.operations = [...operations.values()].sort((left, right) => left.method.localeCompare(right.method) || left.path.localeCompare(right.path));
progress.counts = Object.fromEntries(['GET', 'POST', 'PUT', 'PATCH', 'DELETE'].map(method => [method, progress.operations.filter(operation => operation.method === method).length]).filter(([, count]) => count > 0));
progress.counts.total = progress.operations.length;
progress.coverage_confidence = {
  level: 'complete_rendered_reference',
  basis: `Retrieved every one of ${urls.length} Bigin v2 API-reference pages enumerated by the provider sitemap at ${sitemapURL}; deduplicated the provider's displayed method/path pairs across regional host variants.`,
};
writeJSON(progress);
console.log(`bigin complete: ${urls.length} pages, ${progress.operations.length} unique operations`);
