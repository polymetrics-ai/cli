// Verifies that connector enrichment, generated catalog data, and connector
// icons cover the source catalog exactly.

import { existsSync, readFileSync, readdirSync, statSync } from 'node:fs';
import { dirname, join, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const CATALOG = resolve(__dirname, '../../internal/connectors/catalog_data.json');
const GENERATED_CATALOG = resolve(__dirname, '../lib/connectors.catalog.generated.ts');
const ENRICHMENT = resolve(__dirname, '../lib/connectors.enrichment.generated.ts');
const WEBSITE_ROOT = resolve(__dirname, '..');
const PUBLIC_ROOT = resolve(WEBSITE_ROOT, 'public');
const oldIconPhraseA = 'place' + 'holder ico' + 'n';
const oldIconPhraseB = 'stand-in for a real service lo' + 'go';
const PLACEHOLDER_PATTERNS = [
  /name\.charAt\(0\)/,
  new RegExp(oldIconPhraseA, 'i'),
  new RegExp(oldIconPhraseB, 'i'),
];

const raw = JSON.parse(readFileSync(CATALOG, 'utf8'));
const sourceCatalog = (Array.isArray(raw) ? raw : (raw.connectors ?? raw.definitions ?? []))
  .filter((e) => e && e.slug && e.name);
const catalogSlugs = new Set(sourceCatalog.map((e) => e.slug));

const generatedCatalogText = readFileSync(GENERATED_CATALOG, 'utf8');
const generatedCatalogMatch = generatedCatalogText.match(
  /export const CONNECTOR_CATALOG: ConnectorMeta\[\] = (.*);\n\nexport const CONNECTOR_CATALOG_COUNT/s,
);
if (!generatedCatalogMatch) {
  console.error(JSON.stringify({ ok: false, error: 'Could not parse generated connector catalog.' }, null, 2));
  process.exit(1);
}

const generatedCatalog = JSON.parse(generatedCatalogMatch[1]);
const generatedCatalogSlugs = new Set(generatedCatalog.map((e) => e.slug));

const text = readFileSync(ENRICHMENT, 'utf8');
const enrichmentSlugs = new Set([...text.matchAll(/^  '([^']+)': \{/gm)].map((match) => match[1]));

const missingGeneratedCatalog = [...catalogSlugs].filter((slug) => !generatedCatalogSlugs.has(slug)).sort();
const extraGeneratedCatalog = [...generatedCatalogSlugs].filter((slug) => !catalogSlugs.has(slug)).sort();
const missingEnrichment = [...catalogSlugs].filter((slug) => !enrichmentSlugs.has(slug)).sort();
const extraEnrichment = [...enrichmentSlugs].filter((slug) => !catalogSlugs.has(slug)).sort();
const missingIcons = generatedCatalog
  .filter((c) => c.icon && !existsSync(resolve(PUBLIC_ROOT, `.${c.icon.publicPath}`)))
  .map((c) => ({ slug: c.slug, publicPath: c.icon.publicPath }))
  .sort((a, b) => a.slug.localeCompare(b.slug));

function sourceFiles(root, files = []) {
  for (const entry of readdirSync(root)) {
    const path = join(root, entry);
    const rel = relative(WEBSITE_ROOT, path);
    if (rel.includes('.next') || rel.includes('node_modules') || rel.includes('public')) continue;
    const stat = statSync(path);
    if (stat.isDirectory()) {
      sourceFiles(path, files);
    } else if (/\.(ts|tsx|mjs|js)$/.test(entry) && !entry.endsWith('.generated.ts')) {
      files.push(path);
    }
  }
  return files;
}

const stalePlaceholderMatches = [];
for (const file of sourceFiles(WEBSITE_ROOT)) {
  const body = readFileSync(file, 'utf8');
  for (const pattern of PLACEHOLDER_PATTERNS) {
    const match = body.match(pattern);
    if (match) {
      stalePlaceholderMatches.push({
        file: relative(WEBSITE_ROOT, file),
        pattern: match[0],
      });
    }
  }
}

if (
  generatedCatalog.length !== sourceCatalog.length ||
  missingGeneratedCatalog.length > 0 ||
  extraGeneratedCatalog.length > 0 ||
  missingEnrichment.length > 0 ||
  extraEnrichment.length > 0 ||
  missingIcons.length > 0 ||
  stalePlaceholderMatches.length > 0
) {
  console.error(
    JSON.stringify(
      {
        ok: false,
        sourceCatalogCount: sourceCatalog.length,
        generatedCatalogCount: generatedCatalog.length,
        enrichmentCount: enrichmentSlugs.size,
        missingGeneratedCatalog,
        extraGeneratedCatalog,
        missingEnrichment,
        extraEnrichment,
        missingIcons,
        stalePlaceholderMatches,
      },
      null,
      2,
    ),
  );
  process.exit(1);
}

console.log(
  JSON.stringify(
    {
      ok: true,
      sourceCatalogCount: sourceCatalog.length,
      generatedCatalogCount: generatedCatalog.length,
      enrichmentCount: enrichmentSlugs.size,
      iconCount: generatedCatalog.filter((c) => c.icon).length,
    },
    null,
    2,
  ),
);
