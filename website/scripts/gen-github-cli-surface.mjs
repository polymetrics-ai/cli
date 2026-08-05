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
    'Bounded GET downloads that write a file to a caller-supplied `--dest-root`; no response body is rendered.',
  ],
  ['direct_write', 'Blocked unless a safe reverse-ETL execution model is designed.'],
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

function currentSurfaceTable() {
  const commands = readJSON('cli_surface.json').commands ?? [];
  const endpoints = readJSON('api_surface.json').endpoints ?? [];
  const streams = readJSON('streams.json').streams ?? [];
  const writes = readJSON('writes.json').actions ?? [];

  const availability = countBy(commands, 'availability');
  assertAccountedFor(
    'availability',
    availability,
    AVAILABILITY_ROWS.map(([, value]) => value),
    commands.length,
  );

  // An api_surface row is either joined to a command/stream/write by covered_by
  // or is a blocked ledger-only row; it is never both.
  const covered = endpoints.filter((endpoint) => endpoint?.covered_by).length;
  const excluded = endpoints.filter((endpoint) => !endpoint?.covered_by).length;

  const rows = [
    ['Declared command entries', commands.length],
    ...AVAILABILITY_ROWS.map(([label, value]) => [label, availability.get(value) ?? 0]),
    ['GitHub read streams', streams.length],
    ['GitHub write actions', writes.length],
    ['Tracked REST endpoints', endpoints.length],
    ['Covered REST endpoints', covered],
    ['Excluded REST endpoints', excluded],
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

let page = readFileSync(PAGE, 'utf8');
page = replaceBlock(page, 'current-surface', currentSurfaceTable());
page = replaceBlock(page, 'execution-model', executionModelTable());
writeFileSync(PAGE, page);
console.log(`wrote ${PAGE}`);
