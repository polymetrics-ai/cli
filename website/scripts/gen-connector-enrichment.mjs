// Generates website/lib/connectors.enrichment.generated.ts from the Go connector
// catalog plus verified per-slug overlays in website/.enrich/enr. Existing
// hand-curated entries listed in CURATED_KEYS are preserved; overlays are used
// for current catalog slugs; every remaining connector gets a conservative
// setup/auth entry derived only from catalog config fields, secret fields, and
// vendor documentation links.
//
// Run: node scripts/gen-connector-enrichment.mjs

import { existsSync, readFileSync, writeFileSync, mkdirSync, readdirSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const CATALOG = resolve(__dirname, '../../internal/connectors/catalog_data.json');
const ENRICH_DIR = resolve(__dirname, '../.enrich/enr');
const OUT = resolve(__dirname, '../lib/connectors.enrichment.generated.ts');

const CURATED_FIELD_REFERENCES = {
  'source-github': [
    'api_url',
    'credentials.access_token',
    'credentials.client_id',
    'credentials.client_secret',
    'credentials.personal_access_token',
    'repositories',
    'start_date',
  ],
  'source-notion': [
    'credentials.access_token',
    'credentials.client_id',
    'credentials.client_secret',
    'credentials.token',
    'start_date',
  ],
  'source-postgres': [
    'database',
    'host',
    'password',
    'port',
    'schemas',
    'ssl_mode',
    'ssl_mode.ca_certificate',
    'ssl_mode.client_certificate',
    'ssl_mode.client_key',
    'tunnel_method',
    'tunnel_method.ssh_key',
    'username',
  ],
  'source-salesforce': [
    'client_id',
    'client_secret',
    'is_sandbox',
    'refresh_token',
    'start_date',
  ],
  'source-stripe': [
    'account_id',
    'client_secret',
    'start_date',
  ],
};
const CURATED_KEYS = new Set(Object.keys(CURATED_FIELD_REFERENCES));

const FIELD_LIMIT = 18;
const SOURCE_LIMIT = 5;

const raw = JSON.parse(readFileSync(CATALOG, 'utf8'));
const items = (Array.isArray(raw) ? raw : (raw.connectors ?? raw.definitions ?? []))
  .filter((e) => e && e.slug && e.name)
  .sort((a, b) =>
    a.name.localeCompare(b.name, 'en', { sensitivity: 'base' }) ||
    String(a.type || '').localeCompare(String(b.type || '')),
  );

function extractExistingEntries() {
  if (!existsSync(OUT)) return new Map();

  const text = readFileSync(OUT, 'utf8');
  const matches = [...text.matchAll(/^  '([^']+)': \{/gm)];
  const entries = new Map();

  for (let i = 0; i < matches.length; i++) {
    const key = matches[i][1];
    if (!CURATED_KEYS.has(key)) continue;

    const start = matches[i].index;
    const end =
      i + 1 < matches.length ? matches[i + 1].index : text.indexOf('\n};', start);
    if (start === undefined || end === -1) continue;

    entries.set(key, text.slice(start, end).trimEnd());
  }

  return entries;
}

const existingCuratedEntries = extractExistingEntries();

const trim = (s) => (typeof s === 'string' ? s.trim() : '');
const capitalize = (s) => (s ? `${s[0].toUpperCase()}${s.slice(1)}` : s);
const isAirbyteConnector = (e) =>
  /airbyte/i.test(e.slug || '') || (e.name || '').toLowerCase() === 'airbyte';
const asArr = (v) => (Array.isArray(v) ? v : []);

function fieldNames(fields) {
  return fields.map((f) => trim(f.name)).filter(Boolean);
}

function truncateList(names, limit = FIELD_LIMIT) {
  if (names.length <= limit) return names;
  return [...names.slice(0, limit), `and ${names.length - limit} more`];
}

function codeList(names) {
  const shown = truncateList(names);
  return shown.map((name) => (name.startsWith('and ') ? name : `\`${name}\``)).join(', ');
}

function typedFieldList(fields) {
  const shown = truncateList(fields, FIELD_LIMIT);
  return shown
    .map((f) => {
      if (typeof f === 'string') return f;
      const name = trim(f.name);
      const type = trim(f.type);
      return type ? `\`${name}\` (${type})` : `\`${name}\``;
    })
    .join(', ');
}

function hasAny(name, patterns) {
  return patterns.some((pattern) => pattern.test(name));
}

function buildAuthMethods(secretFields) {
  const names = fieldNames(secretFields);
  if (names.length === 0) {
    return [
      {
        name: 'No secret fields listed',
        summary:
          'The catalog does not mark any secret fields for this connector. Configure the required non-secret fields before validating the connector.',
      },
    ];
  }

  const used = new Set();
  const methods = [];

  const oauthCandidates = names.filter((field) => /client_id$/i.test(field) || /client_secret$/i.test(field) || /refresh_token$/i.test(field) || /realm_id$/i.test(field) || /tenant_id$/i.test(field));
  const hasOAuthShape = oauthCandidates.length > 0;

  const addGroup = (name, patterns, summaryPrefix) => {
    const group = names.filter((field) => !used.has(field) && hasAny(field, patterns));
    if (group.length === 0) return;
    for (const field of group) used.add(field);
    methods.push({
      name,
      summary: `${summaryPrefix} Secret fields: ${codeList(group)}.`,
    });
  };

  if (hasOAuthShape) {
    addGroup(
      'OAuth 2.0 credentials',
      [/client_id$/i, /client_secret$/i, /refresh_token$/i, /access_token$/i, /realm_id$/i, /tenant_id$/i],
      'Use the provider OAuth flow or app registration to create the token/client credentials required by this connector.',
    );
  }
  addGroup(
    'API key or access token',
    [/api[_-]?key$/i, /apikey$/i, /api[_-]?token$/i, /access_token$/i, /personal_access_token$/i, /management_token$/i, /auth_token$/i, /token$/i],
    'Create a provider API key or token with the least privileges needed for the streams you plan to sync.',
  );
  addGroup(
    'Cloud access keys',
    [/access_key_id$/i, /secret_access_key$/i, /hmac_key/i, /aws_access_key_id$/i, /aws_secret_access_key$/i],
    'Use cloud IAM access credentials scoped to the bucket, dataset, or staging resources configured for this connector.',
  );
  addGroup(
    'Service account credentials',
    [/service_account/i, /credentials_json$/i, /credentials\.json$/i],
    'Use a service account or JSON credential generated by the provider.',
  );
  addGroup(
    'Username/password credentials',
    [/password$/i],
    'Use a dedicated user credential with the minimum database or application privileges required.',
  );
  addGroup(
    'Certificate or private-key credentials',
    [/private_key/i, /client_certificate/i, /client_key/i, /ca_certificate/i, /certificate/i],
    'Use provider-issued certificates or private keys when TLS, key-pair, or certificate authentication is enabled.',
  );
  addGroup(
    'SSH tunnel credentials',
    [/ssh_key/i, /tunnel_user_password/i],
    'Use these credentials only when routing the connector through an SSH tunnel or bastion host.',
  );

  const remaining = names.filter((field) => !used.has(field));
  if (remaining.length > 0) {
    methods.push({
      name: 'Connector-specific secrets',
      summary: `Provide the connector-specific secret values from the provider or target system. Secret fields: ${codeList(remaining)}.`,
    });
  }

  return methods;
}

function buildSources(e) {
  const keepAirbyte = isAirbyteConnector(e);
  const seen = new Set();
  const docs = [];

  const add = (title, url, type = '') => {
    const cleanTitle = trim(title) || trim(type) || 'Vendor documentation';
    const cleanURL = trim(url);
    if (!cleanURL || seen.has(cleanURL)) return;
    if (!keepAirbyte && /airbyte/i.test(`${cleanTitle} ${cleanURL}`)) return;
    seen.add(cleanURL);
    docs.push({ title: cleanTitle, url: cleanURL });
  };

  for (const doc of Array.isArray(e.official_application_docs) ? e.official_application_docs : []) {
    add(doc.title, doc.url, doc.type);
    if (docs.length >= SOURCE_LIMIT) return docs;
  }

  add(`${e.name} application documentation`, e.application_documentation_url);
  if (keepAirbyte) add(`${e.name} connector documentation`, e.documentation_url);

  return docs.slice(0, SOURCE_LIMIT);
}

function catalogSourceUrls(e) {
  const keepAirbyte = isAirbyteConnector(e);
  const urls = new Set();

  const add = (url) => {
    const cleanURL = trim(url);
    if (!cleanURL) return;
    if (!keepAirbyte && /airbyte/i.test(cleanURL)) return;
    urls.add(cleanURL);
  };

  for (const doc of Array.isArray(e.official_application_docs) ? e.official_application_docs : []) {
    add(doc.url);
  }
  add(e.application_documentation_url);
  if (keepAirbyte) add(e.documentation_url);

  return urls;
}

// De-brand any Airbyte mention that leaked into researched prose. The Airbyte
// connector itself is exempt; all other connector pages should speak in pm terms.
function debrand(s) {
  let t = typeof s === 'string' ? s : '';
  t = t.replace(/\bAn(\s+)Airbyte\b/g, 'A$1pm');
  t = t.replace(/\ban(\s+)Airbyte\b/g, 'a$1pm');
  t = t.replace(/\bAirbyte Cloud\b/gi, 'a hosted environment');
  t = t.replace(/airbytehq\/airbyte/gi, 'owner/repo');
  t = t.replace(/airbyteio/gi, 'example');
  t = t.replace(/airbyte_internal/gi, 'pm_internal');
  t = t.replace(/_airbyte_/gi, '_pm_');
  t = t.replace(/airbyte/gi, 'pm');
  return t.replace(/\s{2,}/g, ' ').trim();
}

function cleanText(value, keepAirbyte) {
  const text = typeof value === 'string' ? value.replace(/[\u0000-\u001F\u007F]/g, ' ').trim() : '';
  if (!text) return '';
  return keepAirbyte ? text.replace(/\s{2,}/g, ' ').trim() : debrand(text);
}

function normalizeOverlay(rawOverlay, e) {
  const keepAirbyte = isAirbyteConnector(e);
  const allowedSources = catalogSourceUrls(e);
  const seenSourceUrls = new Set();

  const overlay = {
    prerequisites: asArr(rawOverlay.prerequisites)
      .map((item) => cleanText(item, keepAirbyte))
      .filter(Boolean)
      .slice(0, 12),
    authMethods: asArr(rawOverlay.authMethods)
      .filter((method) => method && typeof method.name === 'string' && typeof method.summary === 'string')
      .map((method) => ({
        name: cleanText(method.name, keepAirbyte),
        summary: cleanText(method.summary, keepAirbyte),
      }))
      .filter((method) => method.name && method.summary)
      .slice(0, 8),
    setupSteps: asArr(rawOverlay.setupSteps)
      .filter((step) => step && typeof step.title === 'string' && typeof step.body === 'string')
      .map((step) => ({
        title: cleanText(step.title, keepAirbyte),
        body: cleanText(step.body, keepAirbyte),
      }))
      .filter((step) => step.title && step.body)
      .slice(0, 12),
    sources: asArr(rawOverlay.sources)
      .filter((source) => source && typeof source.title === 'string' && typeof source.url === 'string')
      .map((source) => ({
        title: cleanText(source.title, keepAirbyte),
        url: trim(source.url),
      }))
      .filter((source) => {
        if (!source.title || !/^https?:\/\//i.test(source.url)) return false;
        if (!keepAirbyte && /airbyte/i.test(`${source.title} ${source.url}`)) return false;
        if (allowedSources.size > 0 && !allowedSources.has(source.url)) return false;
        if (seenSourceUrls.has(source.url)) return false;
        seenSourceUrls.add(source.url);
        return true;
      })
      .slice(0, SOURCE_LIMIT),
  };

  if (overlay.authMethods.length === 0 && overlay.setupSteps.length === 0) {
    return null;
  }

  return overlay;
}

function loadResearchOverlays() {
  const bySlug = new Map(items.map((e) => [e.slug, e]));
  const overlays = new Map();
  let stale = 0;
  let invalid = 0;

  if (!existsSync(ENRICH_DIR)) return { overlays, stale, invalid };

  for (const file of readdirSync(ENRICH_DIR).filter((name) => name.endsWith('.json')).sort()) {
    const slug = file.slice(0, -'.json'.length);
    const catalogEntry = bySlug.get(slug);
    if (!catalogEntry) {
      stale++;
      continue;
    }

    try {
      const rawOverlay = JSON.parse(readFileSync(resolve(ENRICH_DIR, file), 'utf8'));
      const overlay = normalizeOverlay(rawOverlay, catalogEntry);
      if (overlay) overlays.set(slug, overlay);
      else invalid++;
    } catch (err) {
      invalid++;
      console.warn(`  ! skip ${slug}: invalid enrichment overlay (${err.message})`);
    }
  }

  return { overlays, stale, invalid };
}

function catalogFieldSet(e) {
  return new Set([
    ...(Array.isArray(e.config_fields) ? e.config_fields : []).map((f) => trim(f.name)),
    ...(Array.isArray(e.secret_fields) ? e.secret_fields : []).map(trim),
  ].filter(Boolean));
}

function curatedEntryIsCurrent(e) {
  const expected = CURATED_FIELD_REFERENCES[e.slug];
  if (!expected) return false;

  const fields = catalogFieldSet(e);
  return expected.every((name) => fields.has(name));
}

function buildGeneratedEntry(e) {
  const configFields = Array.isArray(e.config_fields) ? e.config_fields : [];
  const requiredFields = configFields.filter((f) => !!f.required && trim(f.name));
  const optionalFields = configFields.filter((f) => !f.required && trim(f.name));
  const secretNames = Array.isArray(e.secret_fields) ? e.secret_fields.map((name) => ({ name })) : [];
  const syncWindowFields = configFields.filter((f) =>
    /(^|_)(start|end)_date$|replication_start_date|lookback|window/i.test(trim(f.name)),
  );

  const requiredText = requiredFields.length > 0
    ? typedFieldList(requiredFields)
    : 'No required configuration fields are listed in the catalog.';
  const secretText = secretNames.length > 0
    ? codeList(fieldNames(secretNames))
    : 'No secret fields are listed in the catalog.';
  const optionalText = optionalFields.length > 0
    ? typedFieldList(optionalFields)
    : 'No optional configuration fields are listed in the catalog.';

  const prerequisites = [
    `Access to ${e.name} with permission to create or read the resources used by this ${e.type || 'connector'} connector.`,
    requiredFields.length > 0
      ? `Values for required configuration fields: ${requiredText}.`
      : requiredText,
    secretNames.length > 0
      ? `Credential material for secret fields: ${secretText}.`
      : secretText,
  ];

  const setupSteps = [
    {
      title: 'Review the connector schema',
      body: requiredFields.length > 0
        ? `Run \`pm connectors inspect ${e.slug} --json\` before entering credentials. Confirm these required fields match your ${e.name} account or environment: ${requiredText}.`
        : `Run \`pm connectors inspect ${e.slug} --json\` before entering credentials. This catalog entry does not list required configuration fields.`,
    },
    {
      title: secretNames.length > 0 ? 'Create or collect credentials' : 'Confirm authentication requirements',
      body: secretNames.length > 0
        ? `Create credentials in ${e.name} or the target platform, then store them only through credential management. Secret fields expected by the catalog: ${secretText}.`
        : 'This catalog entry has no secret fields. Confirm the provider does not require additional credentials outside the listed configuration before running the connector.',
    },
    {
      title: 'Fill configuration values',
      body: optionalFields.length > 0
        ? `Provide the required fields first, then decide whether any optional fields are needed for filtering, performance, or advanced connection behavior. Optional fields include: ${optionalText}.`
        : optionalText,
    },
  ];

  if (syncWindowFields.length > 0) {
    setupSteps.push({
      title: 'Set the sync window intentionally',
      body: `This connector exposes date or lookback controls: ${typedFieldList(syncWindowFields)}. Use the narrowest initial window that still covers the data you need, then widen it after validation if necessary.`,
    });
  }

  setupSteps.push({
    title: e.runtime_capabilities?.check
      ? 'Validate before running data movement'
      : 'Use metadata until the native port is enabled',
    body: e.runtime_capabilities?.check
      ? 'After credentials are stored through your normal secret-handling path, run `pm credentials test <connection_name> --json` before scheduling ETL or reverse ETL.'
      : `This catalog entry is inspectable with \`pm connectors inspect ${e.slug} --json\` and \`pm connectors port-plan ${e.slug} --json\`, but runtime checks and ETL are unavailable until implementation_status is enabled.`,
  });

  return {
    prerequisites,
    authMethods: buildAuthMethods(secretNames),
    setupSteps,
    sources: buildSources(e),
  };
}

function indent(text, spaces) {
  const pad = ' '.repeat(spaces);
  return text.split('\n').map((line) => `${pad}${line}`).join('\n');
}

function renderGeneratedEntry(e) {
  const value = JSON.stringify(buildGeneratedEntry(e), null, 2);
  return `  '${e.slug}': ${indent(value, 2).trimStart()},`;
}

let curatedPreserved = 0;
let overlayUsed = 0;
const { overlays, stale: staleOverlays, invalid: invalidOverlays } = loadResearchOverlays();
const entries = items.map((e) => {
  const existing = existingCuratedEntries.get(e.slug);
  if (existing && curatedEntryIsCurrent(e)) {
    curatedPreserved++;
    return existing;
  }

  const overlay = overlays.get(e.slug);
  if (overlay) {
    overlayUsed++;
    const value = JSON.stringify(overlay, null, 2);
    return `  '${e.slug}': ${indent(value, 2).trimStart()},`;
  }

  return renderGeneratedEntry(e);
});

const banner = `/**
 * Connector enrichment data — "Setup & Authentication" content.
 *
 * AUTO-GENERATED by scripts/gen-connector-enrichment.mjs — DO NOT EDIT.
 * Sources: internal/connectors/catalog_data.json and website/.enrich/enr/*.json
 * Run \`npm run gen:enrichment\` to regenerate.
 *
 * Hand-curated entries listed in the generator's CURATED_KEYS set are preserved.
 * Verified overlay entries are used for current catalog slugs. Fallback entries
 * use only catalog config fields, secret fields, and vendor docs.
 * Facts that could not be confirmed from those sources are omitted.
 *
 * REVIEWED: 2026-06-30
 */

export type ConnectorEnrichment = {
  prerequisites: string[];
  authMethods: { name: string; summary: string }[];
  setupSteps: { title: string; body: string }[];
  sources: { title: string; url: string }[];
};

export const CONNECTOR_ENRICHMENT: Record<string, ConnectorEnrichment> = {
`;

const footer = `
};

export const CONNECTOR_ENRICHMENT_COUNT = ${items.length};

export function connectorEnrichment(slug: string): ConnectorEnrichment | undefined {
  return CONNECTOR_ENRICHMENT[slug];
}
`;

const body = `${banner}${entries.join('\n\n')}\n${footer}`;

mkdirSync(dirname(OUT), { recursive: true });
writeFileSync(OUT, body, 'utf8');

console.log(
  `Wrote ${items.length} connector enrichment entries to lib/connectors.enrichment.generated.ts ` +
    `(${curatedPreserved} curated preserved, ${overlayUsed} overlays, ` +
    `${items.length - curatedPreserved - overlayUsed} generated; ` +
    `${staleOverlays} stale overlays ignored, ${invalidOverlays} invalid/empty overlays skipped).`,
);
