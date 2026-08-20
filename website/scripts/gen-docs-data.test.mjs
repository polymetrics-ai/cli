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
});
