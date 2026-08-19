#!/usr/bin/env node

// Copper's MkDocs search index is the provider-published rendered corpus: it
// contains all 637 documentation nodes. Persist each parsed node so an
// interrupted pass resumes at the next unrecorded location rather than
// treating a partial parse as a complete source lock.
import { createHash } from 'node:crypto';
import { readFileSync, writeFileSync } from 'node:fs';

const progressPath = new URL('../crawl-progress.json', import.meta.url);
const indexURL = 'https://developer.copper.com/search/search_index.json';
const referenceRoot = 'https://developer.copper.com/';

const progress = JSON.parse(readFileSync(progressPath, 'utf8'));
const crawl = progress.crawls.copper;
crawl.document_operations ||= {};
const delay = milliseconds => new Promise(resolve => setTimeout(resolve, milliseconds));

function save() {
  progress.updated_at = new Date().toISOString();
  writeFileSync(progressPath, `${JSON.stringify(progress, null, 2)}\n`);
}

function sourceURL(location) {
  return new URL(location, referenceRoot).href;
}

function normalizePath(path) {
  return path.replace(/\{\{([^}]+)\}\}/g, (_, name) => `{${name.replace(/^(?:example|delete|update)_/, '')}}`);
}

function operations(document) {
  // Only the rendered declaration form (not a curl example) is authoritative
  // for the route. Example requests often replace path parameters with ids.
  const matcher = /<code>\s*(GET|POST|PUT|PATCH|DELETE)\s+https:\/\/api\.copper\.com\/developer_api(\/v1\/[^\s<?]+)(?:\?[^\s<]*)?/gi;
  const found = [];
  for (const match of document.text.matchAll(matcher)) {
    found.push({
      method: match[1].toUpperCase(),
      path: normalizePath(match[2]),
      source_url: sourceURL(document.location),
      title: document.title,
    });
  }
  return found;
}

let response;
let body;
for (let attempt = 0; attempt < 5; attempt += 1) {
  response = await fetch(indexURL);
  body = Buffer.from(await response.arrayBuffer());
  if (response.ok) break;
  if (response.status !== 429 || attempt === 4) throw new Error(`HTTP ${response.status} ${indexURL}`);
  const backoffMilliseconds = 2000 * (2 ** attempt);
  crawl.state = 'partial_rate_limited';
  crawl.coverage_confidence = 'partial';
  crawl.last_error = `HTTP 429 ${indexURL}; retry ${attempt + 1}/4 after ${backoffMilliseconds}ms`;
  crawl.resume_strategy = 'Resume the persisted document_operations map after bounded backoff; never promote a partial document count into the source lock.';
  save();
  await delay(backoffMilliseconds);
}
if (!response?.ok) throw new Error(`could not retrieve ${indexURL}`);
const index = JSON.parse(body.toString('utf8'));
if (!Array.isArray(index.docs)) throw new Error('Copper search index did not contain docs[]');

crawl.state = 'in_progress';
crawl.reference_root = referenceRoot;
crawl.index_url = indexURL;
crawl.documents_total = index.docs.length;
crawl.index_bytes = body.length;
crawl.index_sha256 = createHash('sha256').update(body).digest('hex');
crawl.coverage_confidence = 'partial';
save();

for (const [position, document] of index.docs.entries()) {
  if (!crawl.document_operations[document.location]) {
    crawl.document_operations[document.location] = operations(document);
  }
  crawl.documents_retrieved = Object.keys(crawl.document_operations).length;
  // A short checkpoint interval makes Ctrl-C and transient host failures
  // resumable without continually rewriting the progress document.
  if ((position + 1) % 25 === 0 || position + 1 === index.docs.length) save();
}

if (crawl.documents_retrieved !== crawl.documents_total) {
  crawl.state = 'partial_error';
  crawl.coverage_confidence = 'partial';
  crawl.resume_strategy = 'Resume the persisted document_operations map; do not promote its partial routes into the source lock.';
  save();
  process.exitCode = 75;
} else {
  const seen = new Set();
  const complete = Object.values(crawl.document_operations).flat().filter(operation => {
    const key = `${operation.method} ${operation.path}`;
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
  crawl.state = 'complete';
  crawl.operations_found = complete.length;
  crawl.operation_counts = Object.fromEntries(['GET', 'POST', 'PUT', 'PATCH', 'DELETE'].map(method => [method, complete.filter(operation => operation.method === method).length]));
  crawl.completed_operations = complete;
  crawl.coverage_confidence = 'complete_rendered_reference';
  crawl.resume_strategy = null;
  save();
  console.log(`copper crawl complete: ${crawl.documents_retrieved}/${crawl.documents_total} documents, ${complete.length} unique operations`);
}
