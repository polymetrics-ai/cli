#!/usr/bin/env node

// Persisted, resume-safe crawl for Square's rendered API reference. It only
// promotes a complete operation list after every API group from the root page
// is saved in crawl-progress.json. Partial results remain explicitly partial.
import { readFileSync, writeFileSync } from 'node:fs';
import { createHash } from 'node:crypto';

const progressPath = new URL('../crawl-progress.json', import.meta.url);
const root = 'https://developer.squareup.com/reference/square';
const delay = milliseconds => new Promise(resolve => setTimeout(resolve, milliseconds));
const decode = value => value
  .replace(/<wbr\s*\/?\s*>/gi, '')
  .replace(/<[^>]+>/g, '')
  .replace(/&amp;/g, '&')
  .replace(/&#x27;/g, "'")
  .trim();

function readProgress() {
  return JSON.parse(readFileSync(progressPath, 'utf8'));
}

function writeProgress(progress) {
  progress.updated_at = '2026-08-19T00:00:00Z';
  writeFileSync(progressPath, `${JSON.stringify(progress, null, 2)}\n`);
}

async function get(url) {
  const response = await fetch(url);
  const body = Buffer.from(await response.arrayBuffer());
  if (!response.ok) {
    throw Object.assign(new Error(`HTTP ${response.status} ${url}`), { status: response.status });
  }
  return body;
}

function groupURLs(html) {
  return [...new Set([...html.matchAll(/href="(\/reference\/square\/[a-z0-9-]+-api)"/g)].map(match => new URL(match[1], root).href))].sort();
}

function groupOperations(html, groupURL) {
  const rows = [];
  const matcher = /<a\s+data-testid="reference-card--link"[^>]*href="([^"]+)"[\s\S]*?<span[^>]*http-method-string_[^>]*>(GET|POST|PUT|PATCH|DELETE)<\/span><span[^>]*>([\s\S]*?)<\/span>/g;
  for (const match of html.matchAll(matcher)) {
    const path = decode(match[3]);
    if (!path.startsWith('/')) throw new Error(`unparseable path ${JSON.stringify(path)} in ${groupURL}`);
    rows.push({ method: match[2], path, source_url: new URL(match[1], root).href });
  }
  return rows;
}

async function fetchWithBackoff(url, progress, crawl) {
  for (let attempt = 0; attempt < 5; attempt += 1) {
    try {
      return await get(url);
    } catch (error) {
      if (error.status !== 429 || attempt === 4) throw error;
      const backoffMilliseconds = 2000 * (2 ** attempt);
      crawl.state = 'partial_rate_limited';
      crawl.coverage_confidence = 'partial';
      crawl.last_error = `${error.message}; retry ${attempt + 1}/4 after ${backoffMilliseconds}ms`;
      crawl.resume_strategy = 'The persisted group_pages records are authoritative progress. Resume only missing pages, with bounded sequential exponential backoff; never promote a count before every root-discovered group succeeds.';
      writeProgress(progress);
      await delay(backoffMilliseconds);
    }
  }
  throw new Error(`unreachable retry state for ${url}`);
}

const progress = readProgress();
const crawl = progress.crawls.square;
crawl.group_pages ||= {};

const rootBody = await fetchWithBackoff(root, progress, crawl);
const groups = groupURLs(rootBody.toString('utf8'));
crawl.groups_total = groups.length;

for (const groupURL of groups) {
  if (crawl.group_pages[groupURL]) continue;
  try {
    const body = await fetchWithBackoff(groupURL, progress, crawl);
    crawl.group_pages[groupURL] = {
      sha256: createHash('sha256').update(body).digest('hex'),
      bytes: body.length,
      operations: groupOperations(body.toString('utf8'), groupURL),
    };
    crawl.groups_retrieved = Object.keys(crawl.group_pages).length;
    crawl.coverage_confidence = 'partial';
    crawl.last_error = null;
    writeProgress(progress);
    // The reference host rate-limits concurrent fetches. One small gap keeps
    // a resumed crawl polite and makes each saved page a durable checkpoint.
    await delay(1250);
  } catch (error) {
    crawl.state = error.status === 429 ? 'partial_rate_limited' : 'partial_error';
    crawl.coverage_confidence = 'partial';
    crawl.groups_retrieved = Object.keys(crawl.group_pages).length;
    crawl.last_error = error.message;
    crawl.resume_strategy = 'Resume the saved crawl; do not use the partial group or operation total as a source-lock denominator.';
    writeProgress(progress);
    console.error(`square crawl paused: ${error.message}`);
    process.exitCode = 75;
    break;
  }
}

if (Object.keys(crawl.group_pages).length === groups.length) {
  const all = Object.values(crawl.group_pages).flatMap(page => page.operations);
  const seen = new Set();
  const unique = all.filter(operation => {
    const key = `${operation.method} ${operation.path}`;
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
  crawl.state = 'complete';
  crawl.groups_retrieved = groups.length;
  crawl.operations_found = unique.length;
  crawl.operation_counts = Object.groupBy(unique, operation => operation.method);
  for (const [method, rows] of Object.entries(crawl.operation_counts)) crawl.operation_counts[method] = rows.length;
  crawl.coverage_confidence = 'complete_rendered_reference';
  crawl.resume_strategy = null;
  crawl.last_error = null;
  crawl.completed_operations = unique;
  writeProgress(progress);
  console.log(`square crawl complete: ${groups.length}/${groups.length} groups, ${unique.length} unique operations`);
} else {
  console.log(`square crawl partial: ${Object.keys(crawl.group_pages).length}/${groups.length} groups; no source-lock count promoted`);
}
