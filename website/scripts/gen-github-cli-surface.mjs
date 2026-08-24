// Rewrites the count tables in website/content/docs/github-cli-surface.mdx from
// the GitHub connector bundle.
// Run: node scripts/gen-github-cli-surface.mjs
// Reads: internal/connectors/defs/github/{cli_surface,api_surface,streams,writes}.json
// Emits: website/content/docs/github-cli-surface.mdx (marked blocks only)
//
// The page used to hand-copy these numbers and drifted to roughly a quarter of
// the real command count, on the same page that argues declared metadata must
// match reality. Only the marked blocks are generated; every other line on the
// page, including each intent's runtime rule, stays hand-written.

import { readFileSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const BUNDLE = resolve(__dirname, '../../internal/connectors/defs/github');
const PAGE = resolve(__dirname, '../content/docs/github-cli-surface.mdx');
const SOURCE_LOCK = resolve(BUNDLE, 'sources/github-operation-source-lock.json');
const COMBINED_LEDGER = resolve(__dirname, '../../.planning/phases/github-parity-extract-r1/GITHUB-COMBINED-OPERATION-LEDGER.json');

// Runtime rules are editorial, not derivable, so they live here rather than in
// the bundle. An intent with no rule is a hard error: a silently blank rule
// would reintroduce exactly the unverified prose this generator removes.
const INTENT_RULES = [
  ['etl', 'Reads map to existing streams or planned stream/direct-read coverage.'],
  [
    'reverse_etl',
    'Commands with explicit `record.*` flag mappings create stored reverse plans; execution requires preview and approval.',
  ],
  ['direct_read', 'Constrained read commands that do not naturally fit streams.'],
  [
    'binary_download',
    'Bounded GET downloads that require declared response status, media, and redirect policy before writing a file beneath a caller-supplied `--dest-root`; no response body is rendered.',
  ],
  [
    'binary_upload',
    'Declared, bounded local-file uploads follow plan → preview → approval → execute; the provider URL, byte cap, media policy, and request body remain connector-owned.',
  ],
  [
    'direct_write',
    'Fixed declared writes follow plan → preview → approval → execute; destructive operations also require typed confirmation.',
  ],
  ['local_workflow', 'Browser, git, shell, extension, completion, or local config behavior.'],
  ['auth', 'Credential lifecycle commands; never print token material.'],
  ['config', 'Local config behavior, not connector data extraction.'],
  ['raw_api', 'Must remain constrained and approval-gated before any implementation.'],
];

const AVAILABILITY_ROWS = [
  ['Implemented command entries', 'implemented'],
  ['Partial command entries', 'partial'],
  ['Planned command entries', 'planned'],
  ['Unsupported API entries', 'unsupported_api'],
  ['Unsupported local workflow entries', 'unsupported_local'],
  ['Unsafe or disallowed entries', 'unsafe_or_disallowed'],
];

const readJSON = (name) => JSON.parse(readFileSync(resolve(BUNDLE, name), 'utf8'));

function sourceCounts(sourceLock) {
  const counts = sourceLock?.counts;
  const names = ['rest', 'graphql_query', 'graphql_mutation', 'total'];
  if (!counts || names.some((name) => !Number.isInteger(counts[name]) || counts[name] < 0)) {
    throw new Error('GitHub source lock has no complete non-negative operation counts');
  }
  if (counts.rest + counts.graphql_query + counts.graphql_mutation !== counts.total) {
    throw new Error('GitHub source lock operation counts do not sum to total');
  }
  return counts;
}

function isLegacyGraphQLBinding(endpoint) {
  return endpoint?.method === 'GRAPHQL';
}

function isFixedGraphQLTransport(endpoint) {
  return endpoint?.method === 'POST'
    && endpoint?.path === '/graphql'
    && Array.isArray(endpoint?.covered_by?.operations);
}

function sourceProgress(ledger, source) {
  if (!ledger || JSON.stringify(ledger.counts) !== JSON.stringify(source)) {
    throw new Error('GitHub combined ledger counts do not match the source lock');
  }
  const metrics = [
    ['inventory', 'classified'],
    ['implementation', 'implemented'],
    ['live_proof', 'proven'],
  ];
  const progress = {};
  for (const [name, numeratorName] of metrics) {
    const metric = ledger.progress?.[name];
    if (!metric || !Number.isInteger(metric[numeratorName]) || metric[numeratorName] < 0 || metric[numeratorName] > source.total || metric.total !== source.total) {
      throw new Error(`GitHub combined ledger ${name} progress is incomplete`);
    }
    const expectedPercent = Number(((metric[numeratorName] / source.total) * 100).toFixed(2));
    if (metric.percent !== expectedPercent) {
      throw new Error(`GitHub combined ledger ${name} percentage does not match its numerator`);
    }
    progress[name] = metric;
  }
  return progress;
}

function countBy(items, key) {
  const counts = new Map();
  for (const item of items) {
    const value = typeof item?.[key] === 'string' ? item[key].trim() : '';
    counts.set(value, (counts.get(value) ?? 0) + 1);
  }
  return counts;
}

function assertAccountedFor(label, counts, known, total) {
  const unknown = [...counts.keys()].filter((value) => !known.includes(value));
  if (unknown.length > 0) {
    throw new Error(`${label}: bundle declares unhandled value(s) ${unknown.join(', ')}`);
  }
  const summed = known.reduce((sum, value) => sum + (counts.get(value) ?? 0), 0);
  if (summed !== total) {
    throw new Error(`${label}: rows sum to ${summed}, bundle declares ${total}`);
  }
}

export function currentSurfaceTable({ commands, endpoints, streams, writes, sourceLock, ledger }) {
  const source = sourceCounts(sourceLock);
  const progress = sourceProgress(ledger, source);

  const availability = countBy(commands, 'availability');
  assertAccountedFor(
    'availability',
    availability,
    AVAILABILITY_ROWS.map(([, value]) => value),
    commands.length,
  );

  // GraphQL shares one physical POST /graphql transport, which is intentionally
  // not a REST OpenAPI operation. Keep that row separate so a future source
  // change cannot silently inflate the REST denominator in the website.
  const restEndpoints = endpoints.filter(
    (endpoint) => !isLegacyGraphQLBinding(endpoint) && !isFixedGraphQLTransport(endpoint),
  );
  const fixedGraphQLTransports = endpoints.filter(isFixedGraphQLTransport);
  const legacyGraphQLBindings = endpoints.filter(isLegacyGraphQLBinding);
  if (restEndpoints.length !== source.rest) {
    throw new Error(`GitHub REST endpoint count = ${restEndpoints.length}, source lock declares ${source.rest}`);
  }
  if (fixedGraphQLTransports.length !== 1) {
    throw new Error(`GitHub fixed GraphQL transport count = ${fixedGraphQLTransports.length}, want 1`);
  }
  const fixedGraphQLOperations = fixedGraphQLTransports[0].covered_by.operations;
  if (
    new Set(fixedGraphQLOperations).size !== fixedGraphQLOperations.length
    || fixedGraphQLOperations.some((operation) => typeof operation !== 'string' || operation.trim() === '')
  ) {
    throw new Error('GitHub fixed GraphQL operation bindings must be unique non-empty strings');
  }
  const sourceGraphQLOperations = fixedGraphQLOperations.filter((operation) =>
    /^github\.graphql\.(?:query|mutation)\./u.test(operation));
  const supplementalGraphQLOperations = fixedGraphQLOperations.filter((operation) =>
    !/^github\.graphql\.(?:query|mutation)\./u.test(operation));
  const expectedGraphQLRoots = source.graphql_query + source.graphql_mutation;
  if (
    sourceGraphQLOperations.length !== expectedGraphQLRoots
    || new Set(sourceGraphQLOperations).size !== expectedGraphQLRoots
  ) {
    throw new Error(
      `GitHub fixed GraphQL root-operation bindings = ${sourceGraphQLOperations.length}, source lock declares ${expectedGraphQLRoots}`,
    );
  }

  const coveredREST = restEndpoints.filter((endpoint) => endpoint?.covered_by).length;
  const blockedREST = restEndpoints.filter((endpoint) => !endpoint?.covered_by).length;

  const rows = [
    ['Declared command entries', commands.length],
    ...AVAILABILITY_ROWS.map(([label, value]) => [label, availability.get(value) ?? 0]),
    ['GitHub read streams', streams.length],
    ['GitHub write actions', writes.length],
    ['Pinned REST operations', source.rest],
    ['Pinned GraphQL Query roots', source.graphql_query],
    ['Pinned GraphQL Mutation roots', source.graphql_mutation],
    ['Pinned source operations', source.total],
    ['Source inventory classification', `${progress.inventory.classified}/${progress.inventory.total} (${progress.inventory.percent}%)`],
    ['Executable implementation', `${progress.implementation.implemented}/${progress.implementation.total} (${progress.implementation.percent}%)`],
    ['Current-head live proof', `${progress.live_proof.proven}/${progress.live_proof.total} (${progress.live_proof.percent}%)`],
    ['Tracked REST endpoints', restEndpoints.length],
    ['Covered REST endpoints', coveredREST],
    ['Blocked REST endpoints', blockedREST],
    ['Fixed GraphQL transport endpoints', fixedGraphQLTransports.length],
    ['Fixed GraphQL root-operation bindings', sourceGraphQLOperations.length],
    ['Supplemental fixed GraphQL bindings', supplementalGraphQLOperations.length],
    ['Legacy GraphQL compatibility bindings', legacyGraphQLBindings.length],
  ];

  return [
    '| Metric | Value |',
    '|---|---:|',
    ...rows.map(([label, value]) => `| ${label} | ${value} |`),
  ].join('\n');
}

function executionModelTable() {
  const commands = readJSON('cli_surface.json').commands ?? [];
  const intents = countBy(commands, 'intent');
  assertAccountedFor(
    'intent',
    intents,
    INTENT_RULES.map(([intent]) => intent),
    commands.length,
  );

  return [
    '| Intent | Current count | Runtime rule |',
    '|---|---:|---|',
    ...INTENT_RULES.map(
      ([intent, rule]) => `| \`${intent}\` | ${intents.get(intent) ?? 0} | ${rule} |`,
    ),
  ].join('\n');
}

// replaceBlock swaps the body between the generated markers, leaving the
// markers in place so the page keeps declaring which lines are not hand-owned.
function replaceBlock(page, id, body) {
  const open = `{/* generated:${id} — regenerate with \`pnpm run gen:website-data\`; do not hand-edit */}`;
  const close = `{/* /generated:${id} */}`;
  const start = page.indexOf(open);
  const end = page.indexOf(close);
  if (start < 0 || end < 0 || end < start) {
    throw new Error(`github-cli-surface.mdx is missing the generated:${id} markers`);
  }
  return `${page.slice(0, start)}${open}\n\n${body}\n\n${page.slice(end)}`;
}

function main() {
  const commands = readJSON('cli_surface.json').commands ?? [];
  const endpoints = readJSON('api_surface.json').endpoints ?? [];
  const streams = readJSON('streams.json').streams ?? [];
  const writes = readJSON('writes.json').actions ?? [];
  const sourceLock = JSON.parse(readFileSync(SOURCE_LOCK, 'utf8'));
  const ledger = JSON.parse(readFileSync(COMBINED_LEDGER, 'utf8'));

  let page = readFileSync(PAGE, 'utf8');
  page = replaceBlock(page, 'current-surface', currentSurfaceTable({ commands, endpoints, streams, writes, sourceLock, ledger }));
  page = replaceBlock(page, 'execution-model', executionModelTable());
  writeFileSync(PAGE, page);
  console.log(`wrote ${PAGE}`);
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main();
}
