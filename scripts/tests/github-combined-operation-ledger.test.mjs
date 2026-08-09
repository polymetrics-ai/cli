import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

import {
  buildCombinedOperationLedger,
  buildSourceLock,
  parseGraphQLRootOperations,
  validateCombinedOperationLedger,
} from "../github-combined-operation-ledger.mjs";

const scriptsDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const fixtureDir = path.join(scriptsDir, "testdata", "github-combined-operation-ledger");

async function fixtures() {
  const [restText, graphqlSchema] = await Promise.all([
    readFile(path.join(fixtureDir, "mini-openapi.json"), "utf8"),
    readFile(path.join(fixtureDir, "mini-schema.graphql"), "utf8"),
  ]);
  return { restDocument: JSON.parse(restText), graphqlSchema };
}

async function currentGitHubLedger() {
  const projectRoot = path.resolve(scriptsDir, "..");
  const definitionDir = path.join(projectRoot, "internal", "connectors", "defs", "github");
  const [lockText, surfaceText, operationsText, cliText] = await Promise.all([
    readFile(path.join(definitionDir, "sources", "github-operation-source-lock.json"), "utf8"),
    readFile(path.join(definitionDir, "api_surface.json"), "utf8"),
    readFile(path.join(definitionDir, "operations.json"), "utf8"),
    readFile(path.join(definitionDir, "cli_surface.json"), "utf8"),
  ]);
  return buildCombinedOperationLedger({
    lock: JSON.parse(lockText),
    bundle: {
      surface: JSON.parse(surfaceText),
      operations: JSON.parse(operationsText),
      cli: JSON.parse(cliText),
    },
  });
}

test("parses every root Query and Mutation field with multiline signatures and directives", async () => {
  const { graphqlSchema } = await fixtures();
  const roots = parseGraphQLRootOperations(graphqlSchema);
  assert.deepEqual(
    roots.map((field) => ({ root: field.root, name: field.name, signature: field.signature, deprecated: field.deprecated, preview: field.preview })),
    [
      {
        root: "Mutation",
        name: "createEnterpriseOrganization",
        signature: "createEnterpriseOrganization(input: CreateEnterpriseOrganizationInput!): CreateEnterpriseOrganizationPayload",
        deprecated: false,
        preview: true,
      },
      {
        root: "Mutation",
        name: "closeWidget",
        signature: "closeWidget(input: CloseWidgetInput!): CloseWidgetPayload",
        deprecated: false,
        preview: false,
      },
      { root: "Query", name: "viewer", signature: "viewer: User!", deprecated: false, preview: false },
      { root: "Query", name: "node", signature: "node(id: ID!): Node", deprecated: false, preview: false },
      { root: "Query", name: "nodes", signature: "nodes(ids: [ID!]!): [Node]!", deprecated: false, preview: false },
      {
        root: "Query",
        name: "widget",
        signature: "widget(owner: String! name: String!): Widget",
        deprecated: true,
        preview: false,
      },
    ],
  );
});

test("builds a source-locked combined ledger with the enterprise-organization canary", async () => {
  const { restDocument, graphqlSchema } = await fixtures();
  const lock = buildSourceLock({
    restDocument,
    graphqlSchema,
    capturedAt: "2026-08-09",
    restSource: { url: "https://example.test/rest.json", commit: "abcdef0123456789" },
    graphqlSource: { url: "https://example.test/schema.graphql" },
  });
  assert.deepEqual(lock.counts, { rest: 3, graphql_query: 4, graphql_mutation: 2, total: 9 });
  assert.equal(lock.graphql.mutation_fields.find((field) => field.name === "createEnterpriseOrganization")?.preview, true);

  const bundle = {
    surface: {
      endpoints: [
        { method: "GET", path: "/widgets", covered_by: { direct_read: "widget list" } },
        { method: "POST", path: "/widgets", operation: { reason: "Named dependency: typed create contract" } },
        { method: "GET", path: "/widgets/{widget_id}", covered_by: { direct_read: "widget view" } },
      ],
    },
    operations: {
      operations: [
        {
          id: "github.viewer",
          kind: "graphql_query",
          graphql: { document: "query Viewer { viewer { login } }", operation_name: "Viewer" },
        },
        {
          id: "github.close-widget",
          kind: "graphql_mutation",
          graphql: { document: "mutation CloseWidget { closeWidget(input: { id: \"widget-1\" }) { clientMutationId } }", operation_name: "CloseWidget" },
        },
        {
          id: "github.node",
          kind: "graphql_query",
          graphql: { document: "query Node { node(id: \"widget-1\") { ... on Widget { id } } }", operation_name: "Node" },
        },
      ],
    },
    cli: {
      commands: [
        { path: "widget list", availability: "implemented", api_surface: [{ method: "GET", path: "/widgets" }] },
        { path: "widget view", availability: "implemented", api_surface: [{ method: "GET", path: "/widgets/{widget_id}" }] },
        { path: "viewer view", availability: "implemented", operation: "github.viewer" },
        { path: "widget close", availability: "unsafe_or_disallowed", operation: "github.close-widget" },
        { path: "widget node", availability: "implemented", operation: "github.node" },
      ],
    },
  };
  const ledger = buildCombinedOperationLedger({ lock, bundle });
  assert.equal(ledger.counts.total, 9);
  assert.deepEqual(ledger.counts, { rest: 3, graphql_query: 4, graphql_mutation: 2, total: 9 });
  assert.equal(ledger.rows.length, 9);
  assert.equal(ledger.rows.find((row) => row.id === "github.graphql.query.viewer")?.implementation.state, "partially_implemented");
  const unavailableMutation = ledger.rows.find((row) => row.id === "github.graphql.mutation.closeWidget");
  assert.equal(unavailableMutation?.implementation.state, "declared_not_executable");
  assert.equal(unavailableMutation?.blocker?.category, "mapped_command_not_executable");
  assert.match(unavailableMutation?.blocker?.reason || "", /not executable/i);
  assert.match(unavailableMutation?.blocker?.unblocking_condition || "", /typed GraphQL operation contract/i);
  assert.deepEqual(ledger.rows.find((row) => row.id === "github.graphql.query.node")?.projection_matrix, {
    root_field: "node",
    policy: "fixed declared documents only; no caller-supplied GraphQL selection",
    possible_object_types: ["User", "Widget"],
    supported_object_types: ["Widget"],
    state: "fixed_projection_only",
  });
  assert.deepEqual(ledger.rows.find((row) => row.id === "github.graphql.query.nodes")?.projection_matrix, {
    root_field: "nodes",
    policy: "fixed declared documents only; no caller-supplied GraphQL selection",
    possible_object_types: ["User", "Widget"],
    supported_object_types: [],
    state: "no_supported_projection",
  });
  const canary = ledger.rows.find((row) => row.id === "github.graphql.mutation.createEnterpriseOrganization");
  assert.equal(canary?.implementation.state, "not_implemented");
  assert.match(canary?.blocker?.unblocking_condition || "", /typed GraphQL operation contract/i);
  for (const row of ledger.rows) {
    for (const field of ["id", "protocol", "source", "pm", "implementation", "auth", "fixture", "safety", "assertion", "cleanup", "terminal_evidence"]) {
      assert.ok(Object.hasOwn(row, field), `${row.id} missing ${field}`);
    }
  }
  assert.doesNotThrow(() => validateCombinedOperationLedger({ lock, ledger }));
});

test("imports typed GraphQL root contracts and source-derived projection possibilities", async () => {
  const { restDocument, graphqlSchema } = await fixtures();
  const lock = buildSourceLock({
    restDocument,
    graphqlSchema,
    capturedAt: "2026-08-09",
    restSource: { url: "https://example.test/rest.json", commit: "abcdef0123456789" },
    graphqlSource: { url: "https://example.test/schema.graphql" },
  });

  assert.equal(lock.schema_version, 2);
  const createEnterpriseOrganization = lock.graphql.mutation_fields.find((field) => field.name === "createEnterpriseOrganization");
  assert.deepEqual(createEnterpriseOrganization?.arguments, [
    { name: "input", type: { kind: "named", name: "CreateEnterpriseOrganizationInput", non_null: true } },
  ]);
  assert.deepEqual(createEnterpriseOrganization?.return_type, {
    kind: "named",
    name: "CreateEnterpriseOrganizationPayload",
    non_null: false,
  });
  assert.deepEqual(lock.graphql.type_system.input_objects.find((input) => input.name === "CreateEnterpriseOrganizationInput"), {
    name: "CreateEnterpriseOrganizationInput",
    fields: [
      { name: "enterpriseAdmin", type: { kind: "named", name: "String", non_null: true } },
      { name: "login", type: { kind: "named", name: "String", non_null: true } },
      { name: "profileName", type: { kind: "named", name: "String", non_null: true } },
    ],
  });
  assert.deepEqual(lock.graphql.type_system.enums.find((entry) => entry.name === "WidgetCloseReason"), {
    name: "WidgetCloseReason",
    values: ["DONE", "STALE"],
  });
  assert.deepEqual(lock.graphql.type_system.interfaces.find((entry) => entry.name === "Node"), {
    name: "Node",
    fields: [{ name: "id", type: { kind: "named", name: "ID", non_null: true } }],
    possible_types: ["User", "Widget"],
  });
  assert.deepEqual(lock.graphql.type_system.unions.find((entry) => entry.name === "SearchResult"), {
    name: "SearchResult",
    possible_types: ["User", "Widget"],
  });
});

test("fails closed when GraphQL root types or the enterprise typed input contract drift", async () => {
  const { restDocument, graphqlSchema } = await fixtures();
  const sources = {
    restSource: { url: "https://example.test/rest.json", commit: "abcdef0123456789" },
    graphqlSource: { url: "https://example.test/schema.graphql" },
    capturedAt: "2026-08-09",
  };
  assert.throws(
    () => buildSourceLock({ ...sources, restDocument, graphqlSchema: graphqlSchema.replace("CreateEnterpriseOrganizationInput!", "MissingInput!") }),
    /MissingInput|unknown GraphQL type/i,
  );
  assert.throws(
    () => buildSourceLock({ ...sources, restDocument, graphqlSchema: graphqlSchema.replace("input: CreateEnterpriseOrganizationInput!", "enterpriseSlug: String!") }),
    /createEnterpriseOrganization.*input.*CreateEnterpriseOrganizationInput/i,
  );
  assert.throws(
    () => buildSourceLock({ ...sources, restDocument, graphqlSchema: `${graphqlSchema}\n type User { id: ID! }\n` }),
    /duplicate (GraphQL )?(type|object) User/i,
  );
});

test("rejects a v2 lock that loses an explicit typed root argument contract", async () => {
  const { restDocument, graphqlSchema } = await fixtures();
  const lock = buildSourceLock({
    restDocument,
    graphqlSchema,
    capturedAt: "2026-08-09",
    restSource: { url: "https://example.test/rest.json", commit: "abcdef0123456789" },
    graphqlSource: { url: "https://example.test/schema.graphql" },
  });
  delete lock.graphql.query_fields.find((field) => field.name === "viewer").arguments;
  assert.throws(
    () => buildCombinedOperationLedger({
      lock,
      bundle: { surface: { endpoints: [] }, operations: { operations: [] }, cli: { commands: [] } },
    }),
    /viewer.*arguments/i,
  );
});

test("rejects a GraphQL source lock without createEnterpriseOrganization", async () => {
  const { restDocument, graphqlSchema } = await fixtures();
  await assert.rejects(
    async () => buildSourceLock({
      restDocument,
      graphqlSchema: graphqlSchema.replace(/createEnterpriseOrganization/g, "createEnterpriseOrgRemoved"),
      capturedAt: "2026-08-09",
      restSource: { url: "https://example.test/rest.json", commit: "abcdef0123456789" },
      graphqlSource: { url: "https://example.test/schema.graphql" },
    }),
    /createEnterpriseOrganization/i,
  );
});

test("records generic node projections and the disabled deleteIssue mutation from the real GitHub bundle", async () => {
  const ledger = await currentGitHubLedger();
  assert.deepEqual(ledger.counts, { rest: 1220, graphql_query: 31, graphql_mutation: 274, total: 1525 });
  const node = ledger.rows.find((row) => row.id === "github.graphql.query.node")?.projection_matrix;
  const nodes = ledger.rows.find((row) => row.id === "github.graphql.query.nodes")?.projection_matrix;
  assert.deepEqual(node?.supported_object_types, ["DraftIssue", "Issue", "ProjectV2", "PullRequest"]);
  assert.deepEqual(nodes?.supported_object_types, []);
  assert.ok(Array.isArray(node?.possible_object_types) && node.possible_object_types.length > 0);
  assert.ok(Array.isArray(nodes?.possible_object_types) && nodes.possible_object_types.length > 0);
  for (const typeName of node.supported_object_types) {
    assert.ok(node.possible_object_types.includes(typeName), `node support ${typeName} is not in the source-derived possible types`);
  }
  assert.deepEqual(nodes.possible_object_types, node.possible_object_types);
  assert.equal(node?.policy, "fixed declared documents only; no caller-supplied GraphQL selection");
  assert.equal(nodes?.policy, "fixed declared documents only; no caller-supplied GraphQL selection");
  assert.equal(node?.state, "fixed_projection_only");
  assert.equal(nodes?.state, "no_supported_projection");
  const deleteIssue = ledger.rows.find((row) => row.id === "github.graphql.mutation.deleteIssue");
  assert.equal(deleteIssue?.implementation.state, "declared_not_executable");
  assert.equal(deleteIssue?.blocker?.category, "mapped_command_not_executable");
  assert.match(deleteIssue?.blocker?.reason || "", /issue delete \(unsafe_or_disallowed\)/i);
  assert.equal(ledger.rows.find((row) => row.id === "github.graphql.mutation.createEnterpriseOrganization")?.implementation.state, "not_implemented");
});

test("permanent source-inventory gates do not retain the fixed four-operation GraphQL denominator", async () => {
  const projectRoot = path.resolve(scriptsDir, "..");
  const gates = [
    "cmd/connectorgen/github_documented_surface_test.go",
    "cmd/connectorgen/github_api_surface_test.go",
    "internal/connectors/certify/stages_surface_inventory_internal_test.go",
    "scripts/tests/github-parity-proof.test.mjs",
  ];
  for (const relativePath of gates) {
    const source = await readFile(path.join(projectRoot, relativePath), "utf8");
    assert.doesNotMatch(source, /githubGraphQLRows|GRAPHQL"\s*:\s*4|want 1224|endpoints:\s*1224/u, `${relativePath} still hard-codes the legacy GraphQL denominator`);
  }
});
