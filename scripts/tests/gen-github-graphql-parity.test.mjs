import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

import { buildSourceLock } from "../github-combined-operation-ledger.mjs";
import {
  buildGitHubGraphQLParityArtifacts,
  validateGitHubGraphQLParityArtifacts,
} from "../gen-github-graphql-parity.mjs";

const scriptsDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const fixtureDir = path.join(scriptsDir, "testdata", "github-combined-operation-ledger");

async function miniLock() {
  const [restText, graphqlSchema] = await Promise.all([
    readFile(path.join(fixtureDir, "mini-openapi.json"), "utf8"),
    readFile(path.join(fixtureDir, "mini-schema.graphql"), "utf8"),
  ]);
  return buildSourceLock({
    restDocument: JSON.parse(restText),
    graphqlSchema,
    capturedAt: "2026-08-09",
    restSource: { url: "https://example.test/rest.json", commit: "abcdef0123456789" },
    graphqlSource: { url: "https://example.test/schema.graphql" },
  });
}

function emptyBundle() {
  return {
    operations: { operations: [] },
    cli: { tagline: "GitHub", usage: "pm github", commands: [] },
    surface: { api: "GitHub", endpoints: [] },
  };
}

test("generates one fixed source-derived contract per GraphQL root", async () => {
  const lock = await miniLock();
  const generated = buildGitHubGraphQLParityArtifacts({ lock, bundle: emptyBundle() });

  assert.equal(generated.operations.length, lock.counts.graphql_query + lock.counts.graphql_mutation);
  assert.equal(generated.commands.length, generated.operations.length);
  assert.equal(generated.transport.method, "POST");
  assert.equal(generated.transport.path, "/graphql");
  assert.deepEqual(
    generated.transport.covered_by.operations,
    generated.operations.map((operation) => operation.id),
    "the shared physical transport must name every fixed operation exactly once",
  );

  const enterprise = generated.operations.find((operation) => operation.id === "github.graphql.mutation.create-enterprise-organization");
  assert.ok(enterprise, "enterprise organization root canary is generated");
  assert.equal(enterprise.kind, "graphql_mutation");
  assert.match(enterprise.graphql.document, /^mutation GitHubMutationCreateEnterpriseOrganization\(/u);
  assert.match(enterprise.graphql.document, /createEnterpriseOrganization\(input: \$input\)/u);
  assert.deepEqual(enterprise.graphql.variables_schema.required, ["input"]);
  assert.equal(enterprise.graphql.variables_schema.properties.input.additionalProperties, false);
  assert.equal(enterprise.graphql.variables_schema.properties.input.properties.login.type, "string");

  const node = generated.operations.find((operation) => operation.id === "github.graphql.query.node");
  assert.match(node.graphql.document, /node\(id: \$id\) \{ __typename \}/u);
  assert.doesNotMatch(node.graphql.document, /caller(?:Selection|Document)|\$selection/u);
  assert.deepEqual(node.graphql.variables_schema.required, ["id"]);

  const nodes = generated.commands.find((command) => command.operation === "github.graphql.query.nodes");
  assert.deepEqual(nodes.flags, [{ name: "ids", type: "json", required: true, maps_to: "body.ids" }]);
  assert.deepEqual(nodes.api_surface, [{ method: "POST", path: "/graphql" }]);
  assert.equal(nodes.output_policy, "json_redacted");

  assert.doesNotThrow(() => validateGitHubGraphQLParityArtifacts({ lock, generated }));
});

test("fails closed for a missing canary, duplicate root command, unbounded list, or unclassified deleteIssue", async () => {
  const lock = await miniLock();
  assert.throws(
    () => buildGitHubGraphQLParityArtifacts({ lock: { ...lock, graphql: { ...lock.graphql, mutation_fields: lock.graphql.mutation_fields.filter((field) => field.name !== "createEnterpriseOrganization") } }, bundle: emptyBundle() }),
    /createEnterpriseOrganization/u,
  );

  const generated = buildGitHubGraphQLParityArtifacts({ lock, bundle: emptyBundle() });
  generated.commands.push({ ...generated.commands[0] });
  assert.throws(() => validateGitHubGraphQLParityArtifacts({ lock, generated }), /duplicate.*command|command.*duplicate/ui);

  const unbounded = buildGitHubGraphQLParityArtifacts({ lock, bundle: emptyBundle() });
  const nodes = unbounded.operations.find((operation) => operation.id === "github.graphql.query.nodes");
  delete nodes.graphql.variables_schema.properties.ids.maxItems;
  assert.throws(() => validateGitHubGraphQLParityArtifacts({ lock, generated: unbounded }), /maxItems/u);
});
