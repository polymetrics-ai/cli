import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

import { buildSourceLock } from "../github-combined-operation-ledger.mjs";
import {
  buildGitHubGraphQLParityArtifacts,
  mergeGitHubGraphQLParityArtifacts,
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
  assert.match(node.graphql.document, /node\(id: \$id\) \{ __typename .* on User \{ __typename \} .* on Widget \{ __typename \}/u);
  assert.doesNotMatch(node.graphql.document, /caller(?:Selection|Document)|\$selection/u);
  assert.deepEqual(node.graphql.variables_schema.required, ["id"]);

  const nodes = generated.commands.find((command) => command.operation === "github.graphql.query.nodes");
  assert.deepEqual(nodes.flags, [{ name: "ids", type: "json", required: true, maps_to: "body.ids" }]);
  assert.deepEqual(nodes.api_surface, [{ method: "POST", path: "/graphql" }]);
  assert.equal(nodes.output_policy, "json_redacted");

  assert.doesNotThrow(() => validateGitHubGraphQLParityArtifacts({ lock, generated }));
});

test("preserves fixed supplemental GraphQL operations on the shared transport", async () => {
  const lock = await miniLock();
  const bundle = emptyBundle();
  bundle.operations.operations.push({ id: "github.repo.list", kind: "graphql_query" });
  bundle.surface.endpoints.push({
    method: "POST",
    path: "/graphql",
    covered_by: { operations: ["github.repo.list"] },
  });

  const generated = buildGitHubGraphQLParityArtifacts({ lock, bundle });
  const merged = mergeGitHubGraphQLParityArtifacts(bundle, generated);
  const transport = merged.surface.endpoints.find(
    (endpoint) => endpoint.method === "POST" && endpoint.path === "/graphql",
  );

  assert.deepEqual(transport.covered_by.operations, [
    "github.repo.list",
    ...generated.operations.map((operation) => operation.id),
  ]);
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

test("uses the declared environment-only secret contract only for source fields that carry secret values", async () => {
	const projectRoot = path.resolve(scriptsDir, "..");
	const lock = JSON.parse(await readFile(path.join(projectRoot, "internal", "connectors", "defs", "github", "sources", "github-operation-source-lock.json"), "utf8"));
	const generated = buildGitHubGraphQLParityArtifacts({ lock, bundle: emptyBundle() });
	const node = generated.operations.find((candidate) => candidate.id === "github.graphql.query.node");
	assert.match(node?.graphql?.document || "", /on Issue \{ id number title isPinned \}/u);
	assert.match(node?.graphql?.document || "", /on PullRequest \{ id number title isDraft \}/u);
	assert.match(node?.graphql?.document || "", /on Repository \{ id databaseId nameWithOwner \}/u);

	for (const name of ["createMigrationSource", "startOrganizationMigration", "startRepositoryMigration"]) {
		const suffix = name.replace(/([a-z0-9])([A-Z])/gu, "$1-$2").toLowerCase();
		const operation = generated.operations.find((candidate) => candidate.id === `github.graphql.mutation.${suffix}`);
		const command = generated.commands.find((candidate) => candidate.path === `graphql mutation ${suffix}`);
		assert.equal(operation?.mutation_class, "secret", `${name} is classified by its declared secret field`);
		assert.deepEqual(operation?.sensitive_policy, {
			input_mode: "env",
			redact_fields: ["body.input"],
			transform: "none",
			approval_mode: "typed_confirmation",
		});
		assert.equal(command?.availability, "implemented");
		assert.ok(command?.flags.some((flag) => flag.name === "input" && flag.type === "json" && flag.required === true && flag.env_only === true));
	}

	const regenerate = generated.operations.find((candidate) => candidate.id === "github.graphql.mutation.regenerate-verifiable-domain-token");
	assert.equal(regenerate?.mutation_class, "destructive", "a token-generation result is not an input secret");
	assert.equal(regenerate?.sensitive_policy, undefined);

	const deleteIssue = generated.commands.find((candidate) => candidate.path === "graphql mutation delete-issue");
	assert.equal(deleteIssue?.availability, "implemented");
	assert.equal(deleteIssue?.approval, "plan, preview, approval, execute (typed destructive confirmation)");
	const transferIssue = generated.commands.find((candidate) => candidate.path === "graphql mutation transfer-issue");
	assert.equal(transferIssue?.availability, "implemented");
	assert.equal(transferIssue?.approval, "plan, preview, approval, execute (typed destructive confirmation)");
});
