import assert from 'node:assert/strict';
import { test } from 'node:test';

import { currentSurfaceTable } from './gen-github-cli-surface.mjs';

const sourceLock = {
  schema_version: 4,
  connector: 'github',
  lanes: {
    direct_read: 'implemented',
    direct_write: 'implemented',
    binary_download: 'implemented',
    binary_upload: 'unsupported',
    etl: 'implemented',
    reverse_etl: 'implemented',
    sync_transport: 'implemented',
  },
  operations: [{ id: 'one' }, { id: 'two' }],
};

const argumentsFor = (lock = sourceLock) => ({
  commands: [
    {
      availability: 'implemented',
      intent: 'direct_read',
      operation: 'one',
      api_surface: [{ method: 'GET', path: '/user' }],
    },
    { availability: 'unsafe_or_disallowed', intent: 'direct_write' },
  ],
  streams: [{ name: 'user' }],
  writes: [],
  sourceLock: lock,
});

test('current GitHub surface reports only schema-4 authoring and rendered execution facts', () => {
  const table = currentSurfaceTable(argumentsFor());
  assert.match(table, /\| Rendered command entries \| 2 \|/u);
  assert.match(table, /\| Rendered read streams \| 1 \|/u);
  assert.match(table, /\| Authored canonical operations \| 2 \|/u);
  assert.match(table, /\| Operation-bound commands \| 1 \|/u);
  assert.match(table, /\| Endpoint-bound commands \| 1 \|/u);
  assert.match(table, /\| Binary upload lane \| unsupported \|/u);
  assert.doesNotMatch(table, /certif|conformance|ledger|compatibility/iu);
});

test('current GitHub surface rejects incomplete lane declarations', () => {
  const lock = structuredClone(sourceLock);
  delete lock.lanes.sync_transport;
  assert.throws(
    () => currentSurfaceTable(argumentsFor(lock)),
    /must declare exactly seven lanes/u,
  );
});

test('current GitHub surface rejects the wrong source-lock version', () => {
  const lock = structuredClone(sourceLock);
  lock.schema_version = 3;
  assert.throws(
    () => currentSurfaceTable(argumentsFor(lock)),
    /must be schema 4/u,
  );
});
