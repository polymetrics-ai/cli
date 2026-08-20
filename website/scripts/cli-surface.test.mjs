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

const pageFlags = JSON.parse(
  readFileSync(
    resolve(
      dirname(fileURLToPath(import.meta.url)),
      '../../internal/connectors/direct_read_page_flags.json',
    ),
    'utf8',
  ),
);

const approvalFlags = JSON.parse(
  readFileSync(
    resolve(
      dirname(fileURLToPath(import.meta.url)),
      '../../internal/connectors/reverse_etl_approval_flags.json',
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

test('a text_export command documents the runtime destination flags', () => {
  const mapped = mapCLISurface(
    surfaceWith({
      intent: 'text_export',
      flags: [{ name: 'report-id', type: 'string', maps_to: 'path.report_id' }],
    }),
  );

  assert.deepEqual(flagNames(mapped), ['report-id', ...sharedFlags.map((flag) => flag.name)]);
});

test('other intents are left exactly as the bundle declares them', () => {
  const mapped = mapCLISurface(
    surfaceWith({ intent: 'etl', flags: [{ name: 'issue-id', type: 'string' }] }),
  );
  assert.deepEqual(flagNames(mapped), ['issue-id']);
});

test('write commands document the shared approval stdin marker exactly once', () => {
  for (const intent of ['reverse_etl', 'direct_write']) {
    const once = mapCLISurface(surfaceWith({ intent }));
    assert.deepEqual(once.global_flags.map((flag) => flag.name), approvalFlags.map((flag) => flag.name));

    const twice = mapCLISurface(once, { keyStyle: 'camel' });
    assert.deepEqual(twice.globalFlags.map((flag) => flag.name), approvalFlags.map((flag) => flag.name));
  }
});

// A direct read returns ONE page, so the website must document how to ask for
// the next one. The flags come from the same JSON the runtime embeds, so the
// catalog page cannot advertise a different navigation surface than the CLI.
test('a direct_read command documents the runtime page flags', () => {
  const mapped = mapCLISurface(
    surfaceWith({ intent: 'direct_read', flags: [{ name: 'issue-id', type: 'string' }] }),
  );
  assert.deepEqual(flagNames(mapped), ['issue-id', ...pageFlags.map((flag) => flag.name)]);
});

test('re-mapping a direct_read command does not duplicate the page flags', () => {
  const once = mapCLISurface(
    surfaceWith({ intent: 'direct_read', flags: [{ name: 'issue-id', type: 'string' }] }),
  );
  const twice = mapCLISurface({ ...once, commands: once.commands });
  assert.deepEqual(flagNames(twice), flagNames(once));
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

test('declared repeatable flags remain in website data', () => {
  const mapped = mapCLISurface(
    surfaceWith({
      intent: 'etl',
      flags: [{ name: 'header-x-mode', type: 'string', repeatable: true, maps_to: 'header.X-Mode' }],
    }),
  );

  assert.equal(mapped.commands[0].flags[0].repeatable, true);
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
