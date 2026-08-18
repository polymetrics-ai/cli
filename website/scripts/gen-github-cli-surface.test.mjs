import assert from 'node:assert/strict';
import { test } from 'node:test';

import { currentSurfaceTable } from './gen-github-cli-surface.mjs';

const sourceLock = {
  counts: {
    rest: 2,
    graphql_query: 1,
    graphql_mutation: 2,
    total: 5,
  },
};

const ledger = {
  counts: sourceLock.counts,
  progress: {
    inventory: { classified: 5, total: 5, percent: 100 },
    implementation: { implemented: 3, total: 5, percent: 60 },
    live_proof: { proven: 1, total: 5, percent: 20 },
  },
};

const argumentsFor = (endpoints) => ({
  commands: [
    { availability: 'implemented', intent: 'direct_read' },
    { availability: 'unsafe_or_disallowed', intent: 'direct_write' },
  ],
  endpoints,
  streams: [],
  writes: [],
  sourceLock,
  ledger,
});

test('current GitHub surface keeps fixed GraphQL transport out of REST counts', () => {
  const table = currentSurfaceTable(argumentsFor([
    { method: 'GET', path: '/user', covered_by: { stream: 'user' } },
    { method: 'DELETE', path: '/user/resource' },
    {
      method: 'POST',
      path: '/graphql',
      covered_by: { operations: ['github.graphql.query.viewer', 'github.graphql.mutation.create-one', 'github.graphql.mutation.delete-one'] },
    },
    { method: 'GRAPHQL', path: 'query { viewer { login } }', covered_by: { direct_read: 'viewer' } },
  ]));

  assert.match(table, /\| Pinned REST operations \| 2 \|/u);
  assert.match(table, /\| Pinned GraphQL Query roots \| 1 \|/u);
  assert.match(table, /\| Pinned GraphQL Mutation roots \| 2 \|/u);
  assert.match(table, /\| Pinned source operations \| 5 \|/u);
  assert.match(table, /\| Source inventory classification \| 5\/5 \(100%\) \|/u);
  assert.match(table, /\| Executable implementation \| 3\/5 \(60%\) \|/u);
  assert.match(table, /\| Current-head live proof \| 1\/5 \(20%\) \|/u);
  assert.match(table, /\| Tracked REST endpoints \| 2 \|/u);
  assert.match(table, /\| Covered REST endpoints \| 1 \|/u);
  assert.match(table, /\| Blocked REST endpoints \| 1 \|/u);
  assert.match(table, /\| Fixed GraphQL root-operation bindings \| 3 \|/u);
  assert.match(table, /\| Supplemental fixed GraphQL bindings \| 0 \|/u);
  assert.match(table, /\| Legacy GraphQL compatibility bindings \| 1 \|/u);
});

test('current GitHub surface accounts for supplemental fixed GraphQL bindings separately', () => {
  const table = currentSurfaceTable(argumentsFor([
    { method: 'GET', path: '/user', covered_by: { stream: 'user' } },
    { method: 'DELETE', path: '/user/resource' },
    {
      method: 'POST',
      path: '/graphql',
      covered_by: {
        operations: [
          'github.repo.list',
          'github.graphql.query.viewer',
          'github.graphql.mutation.create-one',
          'github.graphql.mutation.delete-one',
        ],
      },
    },
  ]));

  assert.match(table, /\| Fixed GraphQL root-operation bindings \| 3 \|/u);
  assert.match(table, /\| Supplemental fixed GraphQL bindings \| 1 \|/u);
});

test('current GitHub surface rejects a missing or incomplete fixed GraphQL transport', () => {
  assert.throws(
    () => currentSurfaceTable(argumentsFor([
      { method: 'GET', path: '/user', covered_by: { stream: 'user' } },
      { method: 'DELETE', path: '/user/resource' },
    ])),
    /fixed GraphQL transport count = 0/u,
  );

  assert.throws(
    () => currentSurfaceTable(argumentsFor([
      { method: 'GET', path: '/user', covered_by: { stream: 'user' } },
      { method: 'DELETE', path: '/user/resource' },
      { method: 'POST', path: '/graphql', covered_by: { operations: ['github.graphql.query.viewer'] } },
    ])),
    /root-operation bindings = 1/u,
  );
});
