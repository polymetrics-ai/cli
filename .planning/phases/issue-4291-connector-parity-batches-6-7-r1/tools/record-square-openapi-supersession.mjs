#!/usr/bin/env node

// Preserve the rendered-crawl history as partial evidence, while recording
// that Square's complete official OpenAPI specification—not that crawl—is the
// settled source-lock denominator.
import { readFileSync, writeFileSync } from 'node:fs';

const progressPath = new URL('../crawl-progress.json', import.meta.url);
const progress = JSON.parse(readFileSync(progressPath, 'utf8'));
const crawl = progress.crawls.square;
const groups = Object.keys(crawl.group_pages ?? {});
if (groups.length !== 40) throw new Error(`expected 40 persisted Square group pages, found ${groups.length}`);
crawl.state = 'superseded_by_machine_spec';
crawl.groups_total = 40;
crawl.groups_retrieved = 40;
crawl.groups_extracted = 40;
crawl.operations_found = null;
crawl.coverage_confidence = 'partial';
crawl.resume_strategy = null;
crawl.static_extraction_defect = 'Seven fetched group pages supplied no static operation cards; their browser-rendered DOM cards were checkpointed, but the rendered crawler is not used as the settled source denominator.';
crawl.authoritative_source = {
  source_url: 'https://raw.githubusercontent.com/square/connect-api-specification/master/api.json',
  sha256: 'a0d0db22c202f68282fd359772461697b3f635df03d3b19b11bb60add7f8ff7c',
  bytes: 3279392,
  openapi: '3.0.0',
  operations_found: 334,
  coverage_confidence: 'complete_machine_readable_specification',
};
delete crawl.completed_operations;
delete crawl.operation_counts;
progress.updated_at = new Date().toISOString();
writeFileSync(progressPath, `${JSON.stringify(progress, null, 2)}\n`);
console.log('square rendered crawl recorded as partial and superseded by complete OpenAPI source');
