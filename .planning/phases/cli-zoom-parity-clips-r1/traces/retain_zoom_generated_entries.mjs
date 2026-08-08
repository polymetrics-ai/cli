// Retain only the freshly generated Zoom entries in aggregate generated catalogs.
//
// This is intentionally a mechanical post-generation scope filter, not a
// hand-authored catalog merge: every Zoom value comes from the whole-repository
// generator output, while all unrelated entries return exactly to HEAD. It is
// needed because the generator also repairs an unrelated pre-existing Warehouse
// description drift during this provider-category slice.
import { execFileSync } from 'node:child_process';
import { readFileSync, writeFileSync } from 'node:fs';

const entries = [
  { path: 'docs/connectors/catalog/all-connectors.json', indent: 2, trailingNewline: true },
  { path: 'website/data/connectors.generated.json', indent: 2, trailingNewline: false },
  { path: 'website/lib/connectors.catalog.data.generated.json', indent: 0, trailingNewline: true },
];

for (const entry of entries) {
  const baseline = JSON.parse(execFileSync('git', ['show', `HEAD:${entry.path}`], {
    encoding: 'utf8',
    maxBuffer: 64 * 1024 * 1024,
  }));
  const generated = JSON.parse(readFileSync(entry.path, 'utf8'));
  const freshZoom = generated.find((item) => item.slug === 'zoom' || item.name === 'zoom');
  const baselineIndex = baseline.findIndex((item) => item.slug === 'zoom' || item.name === 'zoom');
  if (!freshZoom || baselineIndex < 0) {
    throw new Error(`Zoom entry missing from ${entry.path}`);
  }
  baseline[baselineIndex] = freshZoom;
  writeFileSync(entry.path, JSON.stringify(baseline, null, entry.indent) + (entry.trailingNewline ? '\n' : ''));
}
