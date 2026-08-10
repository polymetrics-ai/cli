import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

import { buildSourceLock } from "../github-combined-operation-ledger.mjs";
import { compareGitHubSourceDrift } from "../github-source-drift.mjs";

const scriptsDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const fixtureDir = path.join(scriptsDir, "testdata", "github-combined-operation-ledger");

async function sources() {
  const [restText, graphqlSchema] = await Promise.all([
    readFile(path.join(fixtureDir, "mini-openapi.json"), "utf8"),
    readFile(path.join(fixtureDir, "mini-schema.graphql"), "utf8"),
  ]);
  return { restText, graphqlSchema };
}

async function pinnedLock() {
  const { restText, graphqlSchema } = await sources();
  return buildSourceLock({
    restDocument: JSON.parse(restText),
    restText,
    graphqlSchema,
    capturedAt: "2026-08-09",
    restSource: { url: "https://example.test/rest.json", commit: "abcdef0123456789" },
    graphqlSource: { url: "https://example.test/schema.graphql" },
  });
}

test("source drift keeps separately pinned REST and GraphQL inventories stable", async () => {
  const [lock, { restText, graphqlSchema }] = await Promise.all([pinnedLock(), sources()]);
  const report = compareGitHubSourceDrift({
    lock,
    restText,
    graphqlSchema,
    restSource: { url: "https://example.test/rest.json", commit: "upstream-main" },
    graphqlSource: { url: "https://example.test/schema.graphql" },
    capturedAt: "2026-08-09",
  });

  assert.equal(report.has_drift, false);
  assert.deepEqual(report.pinned.counts, { rest: 3, graphql_query: 4, graphql_mutation: 2, total: 9 });
  assert.deepEqual(report.observed.counts, report.pinned.counts);
  assert.deepEqual(report.rest.added_operation_ids, []);
  assert.deepEqual(report.graphql.added_operation_ids, []);
});

test("source drift fails loudly for an upstream GraphQL root addition or REST change", async () => {
  const [lock, { restText, graphqlSchema }] = await Promise.all([pinnedLock(), sources()]);
  const changedREST = JSON.parse(restText);
  changedREST.paths["/new-widget"] = {
    get: { operationId: "widgets/list-new" },
  };
  const changedSchema = graphqlSchema.replace(
    "type Query {\n  viewer: User!",
    "type Query {\n  addedQuery: String\n  viewer: User!",
  );
  const report = compareGitHubSourceDrift({
    lock,
    restText: JSON.stringify(changedREST),
    graphqlSchema: changedSchema,
    restSource: { url: "https://example.test/rest.json", commit: "upstream-main" },
    graphqlSource: { url: "https://example.test/schema.graphql" },
    capturedAt: "2026-08-09",
  });

  assert.equal(report.has_drift, true);
  assert.deepEqual(report.rest.added_operation_ids, ["github.rest.widgets/list-new"]);
  assert.deepEqual(report.graphql.added_operation_ids, ["github.graphql.query.addedQuery"]);
  assert.equal(report.observed.counts.rest, 4);
  assert.equal(report.observed.counts.graphql_query, 5);
  assert.match(report.summary, /REST.*1 added.*GraphQL.*1 added/i);
});
