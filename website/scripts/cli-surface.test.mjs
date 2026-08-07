import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { dirname, resolve } from 'node:path';
import { test } from 'node:test';
import { fileURLToPath } from 'node:url';

import { mapCLISurface } from './lib/cli-surface.mjs';

const sharedFlags = JSON.parse(
  readFileSync(
    resolve(
      dirname(fileURLToPath(import.meta.url)),
      '../../internal/connectors/binary_download_flags.json',
    ),
    'utf8',
  ),
);

const surfaceWith = (command) => ({
  usage: 'pm acme <command>',
  commands: [{ path: 'artifact download', flags: [], ...command }],
});

const flagNames = (surface) => surface.commands[0].flags.map((flag) => flag.name);

test('a binary_download command documents the runtime destination flags', () => {
  const mapped = mapCLISurface(
    surfaceWith({
      intent: 'binary_download',
      flags: [{ name: 'artifact-id', type: 'string', maps_to: 'path.artifact_id' }],
    }),
  );

  assert.deepEqual(flagNames(mapped), ['artifact-id', ...sharedFlags.map((flag) => flag.name)]);
  const destRoot = mapped.commands[0].flags.find((flag) => flag.name === 'dest-root');
  assert.equal(destRoot.required, true, '--dest-root must be marked required');
});

test('other intents are left exactly as the bundle declares them', () => {
  const mapped = mapCLISurface(
    surfaceWith({ intent: 'direct_read', flags: [{ name: 'issue-id', type: 'string' }] }),
  );
  assert.deepEqual(flagNames(mapped), ['issue-id']);
});

test('declared scalar minima remain in website data', () => {
  const mapped = mapCLISurface(
    surfaceWith({
      intent: 'etl',
      flags: [{ name: 'page-number', type: 'integer', minimum: 1, maps_to: 'config.page_number' }],
    }),
  );

  assert.equal(mapped.commands[0].flags[0].minimum, 1);
});

// gen-connector-catalog.mjs re-maps gen-connector-bundles.mjs output, so an
// unconditional append documented --dest-root twice on the catalog page.
test('re-mapping already-mapped output does not duplicate the destination flags', () => {
  const once = mapCLISurface(
    surfaceWith({
      intent: 'binary_download',
      flags: [{ name: 'artifact-id', type: 'string', maps_to: 'path.artifact_id' }],
    }),
  );
  const twice = mapCLISurface(once, { keyStyle: 'camel' });

  assert.deepEqual(flagNames(twice), flagNames(once));
});

test('a bundle-declared destination flag is not repeated', () => {
  const mapped = mapCLISurface(
    surfaceWith({
      intent: 'binary_download',
      flags: [{ name: 'dest-root', type: 'string', summary: 'bundle-declared', required: true }],
    }),
  );

  assert.equal(flagNames(mapped).filter((name) => name === 'dest-root').length, 1);
  assert.equal(mapped.commands[0].flags[0].summary, 'bundle-declared');
});
