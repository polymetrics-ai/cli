// Builds the per-connector descriptor map the enrichment Workflow grounds on.
// Reads the authoritative catalog, emits a slug-keyed map with ONLY the fields
// the research/verify agents need (vendor docs + secret/config field names) —
// deliberately excludes the upstream (Airbyte) documentation_url so enrichment
// sources stay vendor-only. Output: <scratchpad>/connectors-input.json + slugs.
import { readFileSync, writeFileSync, mkdirSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const CATALOG = resolve(__dirname, '../../internal/connectors/catalog_data.json');
const OUT_DIR = process.argv[2] || resolve(__dirname, '../.enrich');
const items = JSON.parse(readFileSync(CATALOG, 'utf8'));

const isAirbyteConnector = (e) =>
  /airbyte/i.test(e.slug || '') || (e.name || '').toLowerCase() === 'airbyte';

const map = {};
for (const e of items) {
  if (!e?.slug || !e?.name) continue;
  const keepAirbyte = isAirbyteConnector(e);
  const appDoc = e.application_documentation_url || '';
  map[e.slug] = {
    slug: e.slug,
    name: e.name,
    type: e.type,
    category: e.source_type || 'other',
    language: e.language || '',
    secretFields: Array.isArray(e.secret_fields) ? e.secret_fields : [],
    configFields: (Array.isArray(e.config_fields) ? e.config_fields : []).map((f) => ({
      name: f.name,
      required: !!f.required,
      secret: !!f.secret,
    })),
    // vendor docs only (NOT the Airbyte catalog documentation_url); also drop
    // any airbyte.com vendor links so research never grounds on Airbyte.
    appDocUrl: keepAirbyte ? appDoc : /airbyte/i.test(appDoc) ? '' : appDoc,
    officialDocs: (Array.isArray(e.official_application_docs) ? e.official_application_docs : [])
      .filter((d) => keepAirbyte || !/airbyte/i.test(`${d.title || ''} ${d.url || ''}`))
      .map((d) => ({ title: d.title, type: d.type || '', url: d.url })),
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
