// Rewrites the count tables in website/content/docs/github-cli-surface.mdx from
// GitHub's immutable schema-4 source lock and rendered execution JSON.
// Run: node scripts/gen-github-cli-surface.mjs

import { readFileSync, writeFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const __dirname = dirname(fileURLToPath(import.meta.url));
const BUNDLE = resolve(__dirname, '../../internal/connectors/defs/github');
const PAGE = resolve(__dirname, '../content/docs/github-cli-surface.mdx');

const INTENT_RULES = [
  ['etl', 'Reads bind to rendered streams and keep warehouse materialization separate from direct reads.'],
  [
    'reverse_etl',
    'Commands with explicit `record.*` flag mappings create stored reverse plans; execution requires preview and approval.',
  ],
  ['direct_read', 'Constrained reads bind to one rendered provider operation.'],
  [
    'binary_download',
    'Bounded downloads require declared response status, media, redirect, size, and destination policy.',
  ],
  [
    'binary_upload',
    'Bounded local-file uploads follow plan → preview → approval → execute with connector-owned media and size policy.',
  ],
  [
    'direct_write',
    'Fixed rendered writes follow plan → preview → approval → execute; destructive operations require typed confirmation.',
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

const LANE_ROWS = [
  ['Direct read lane', 'direct_read'],
  ['Direct write lane', 'direct_write'],
  ['Binary download lane', 'binary_download'],
  ['Binary upload lane', 'binary_upload'],
  ['ETL lane', 'etl'],
  ['Reverse ETL lane', 'reverse_etl'],
  ['Sync transport lane', 'sync_transport'],
];

const readJSON = (name) => JSON.parse(readFileSync(resolve(BUNDLE, name), 'utf8'));

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

function assertVNextLock(sourceLock) {
  if (sourceLock?.schema_version !== 4 || sourceLock?.connector !== 'github') {
    throw new Error('GitHub source lock must be schema 4 for connector github');
  }
  if (!Array.isArray(sourceLock.operations)) {
    throw new Error('GitHub source lock operations must be an array');
  }
  const laneKeys = Object.keys(sourceLock.lanes ?? {}).sort();
  const expected = LANE_ROWS.map(([, key]) => key).sort();
  if (JSON.stringify(laneKeys) !== JSON.stringify(expected)) {
    throw new Error('GitHub source lock must declare exactly seven lanes');
  }
}

export function currentSurfaceTable({ commands, streams, writes, sourceLock }) {
  assertVNextLock(sourceLock);
  const availability = countBy(commands, 'availability');
  assertAccountedFor(
    'availability',
    availability,
    AVAILABILITY_ROWS.map(([, value]) => value),
    commands.length,
  );

  const endpointBoundCommands = commands.filter((command) =>
    Array.isArray(command?.api_surface) && command.api_surface.length > 0).length;
  const operationBoundCommands = commands.filter((command) =>
    typeof command?.operation === 'string' && command.operation.trim() !== '').length;

  const rows = [
    ['Rendered command entries', commands.length],
    ...AVAILABILITY_ROWS.map(([label, value]) => [label, availability.get(value) ?? 0]),
    ['Rendered read streams', streams.length],
    ['Rendered write actions', writes.length],
    ['Authored canonical operations', sourceLock.operations.length],
    ['Operation-bound commands', operationBoundCommands],
    ['Endpoint-bound commands', endpointBoundCommands],
    ...LANE_ROWS.map(([label, key]) => [label, sourceLock.lanes[key]]),
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
  const streams = readJSON('streams.json').streams ?? [];
  const writes = readJSON('writes.json').actions ?? [];
  const sourceLock = readJSON('source.lock.json');

  let page = readFileSync(PAGE, 'utf8');
  page = replaceBlock(page, 'current-surface', currentSurfaceTable({ commands, streams, writes, sourceLock }));
  page = replaceBlock(page, 'execution-model', executionModelTable());
  writeFileSync(PAGE, page);
  console.log(`wrote ${PAGE}`);
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main();
}
