// Builds a slug-keyed descriptor map from the generated connector bundle data.
// Output: <scratchpad>/connectors-input.json + slugs.
import { readFileSync, writeFileSync, mkdirSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const CONNECTORS = resolve(__dirname, '../data/connectors.generated.json');
const OUT_DIR = process.argv[2] || resolve(__dirname, '../.enrich');
const items = JSON.parse(readFileSync(CONNECTORS, 'utf8'));

const map = {};
for (const e of items) {
  if (!e?.slug || !e?.name) continue;
  map[e.slug] = {
    slug: e.slug,
    name: e.name,
    category: e.integration_type || 'api',
    capabilities: e.capabilities || {},
    streams: (Array.isArray(e.streams) ? e.streams : []).map((stream) => stream.name),
    writeActions: (Array.isArray(e.write_actions) ? e.write_actions : []).map((action) => action.name),
    appDocUrl: e.docs_url || '',
    officialDocs: e.docs_url
      ? [{ title: 'Service API documentation', type: 'api_reference', url: e.docs_url }]
      : [],
  };
}

const slugs = Object.keys(map).sort();
mkdirSync(OUT_DIR, { recursive: true });
mkdirSync(resolve(OUT_DIR, 'enr'), { recursive: true });
writeFileSync(resolve(OUT_DIR, 'connectors-input.json'), JSON.stringify(map), 'utf8');
writeFileSync(resolve(OUT_DIR, 'slugs.json'), JSON.stringify(slugs), 'utf8');
const withVendor = slugs.filter((s) => map[s].officialDocs.length > 0 || map[s].appDocUrl).length;
console.log(
  `Wrote ${slugs.length} descriptors to ${OUT_DIR}/connectors-input.json ` +
    `(${withVendor} have vendor docs). enr/ dir ready.\nFirst 6 slugs: ${slugs.slice(0, 6).join(', ')}`,
);
