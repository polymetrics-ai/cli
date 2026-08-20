import assert from 'node:assert/strict';
import { mkdtempSync, readFileSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join, resolve } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

import { generateDocsData } from './gen-docs-data.mjs';

const scriptDir = dirname(fileURLToPath(import.meta.url));
const websiteRoot = resolve(scriptDir, '..');

test('generated website docs match canonical content', (t) => {
  const outputDir = mkdtempSync(join(tmpdir(), 'polymetrics-website-docs-'));
  t.after(() => rmSync(outputDir, { recursive: true, force: true }));

  const outputPath = join(outputDir, 'docs.generated.ts');
  generateDocsData({ outputPath });

  const generated = readFileSync(outputPath, 'utf8');
  const tracked = readFileSync(join(websiteRoot, 'lib', 'docs.generated.ts'), 'utf8');
  assert.equal(generated, tracked);

  const secondOutputPath = join(outputDir, 'docs.generated.second.ts');
  generateDocsData({ outputPath: secondOutputPath });
  assert.equal(readFileSync(secondOutputPath, 'utf8'), generated, 'generation must be deterministic across consecutive runs');
});

test('generated website data retains runtime CLI semantics', () => {
  const pages = [];
  const outputDir = mkdtempSync(join(tmpdir(), 'polymetrics-website-semantics-'));
  try {
    generateDocsData({ outputPath: join(outputDir, 'docs.generated.ts') });
    const generated = readFileSync(join(outputDir, 'docs.generated.ts'), 'utf8');
    const urls = [...generated.matchAll(/"url": "([^"]+)"/gu)].map((match) => match[1]);
    assert.equal(new Set(urls).size, urls.length, 'generated page URLs must be unique');
    assert.ok(urls.every((url) => url.startsWith('/docs') && !url.includes('\\\\') && !url.includes('//docs')));
    pages.push(generated);
  } finally {
    rmSync(outputDir, { recursive: true, force: true });
  }
  const generated = pages[0];
  assert.match(generated, /pm etl transport postgres-managed-target/u);
  assert.match(generated, /--source <connector>:<credential-name>/u);
  assert.match(generated, /--destination <connector>:<credential-name>/u);
  assert.match(generated, /`7` \| Policy refusal/u);
});
