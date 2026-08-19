import assert from "node:assert/strict";
import test from "node:test";

import {
  buildProofModel,
  validateProofModel,
} from "../github-parity-proof.mjs";
import { classifyHelpResult, classifyInvocationResult } from "../github-command-reachability.mjs";

function currentBundle() {
  return {
    surface: {
      api: "GitHub REST API",
      endpoints: [],
    },
    streams: { streams: [] },
    writes: { actions: [] },
    operations: { operations: [] },
    cli: { commands: [] },
  };
}

test("proof model accounts every declared GitHub surface and generic-only route", async () => {
  const model = await buildProofModel();

  assert.equal(model.counts.endpoints, model.bundle.surface.endpoints.length);
  assert.equal(model.counts.coveredEndpoints + model.counts.blockedEndpoints, model.counts.endpoints);
  assert.equal(model.genericOnly.streams.length, 23);
  assert.equal(model.genericOnly.writeActions.length, 38);
  assert.equal(model.endpointLedger.length, model.counts.endpoints);
  assert.equal(model.commandLedger.length, model.counts.commands);
  assert.equal(validateProofModel(model), true);
});

test("proof model binds the shared GraphQL transport to every declared fixed operation", async () => {
  const model = await buildProofModel();
  const transport = model.endpointLedger.find((row) => row.method === "POST" && row.path === "/graphql");
  assert.ok(transport, "generated GraphQL transport must be present in the endpoint ledger");
  assert.equal(transport.coverage?.kind, "operation");
  const expected = model.bundle.operations.operations
    .filter((operation) => /^github\.graphql\.(?:query|mutation)\./u.test(operation.id))
    .map((operation) => operation.id)
    .sort();
  assert.deepEqual(transport.coverage?.targets, expected);
  assert.deepEqual(transport.links.operations, expected);
  assert.ok(transport.links.commands.length > 0, "transport operations must retain their command routes");
});

test("proof validator rejects an omitted endpoint instead of accepting a summary", async () => {
  const model = await buildProofModel();
  const incomplete = {
    ...model,
    endpointLedger: model.endpointLedger.slice(1),
  };

  assert.throws(
    () => validateProofModel(incomplete),
    new RegExp(`endpoint ledger has ${model.endpointLedger.length - 1} rows, want ${model.endpointLedger.length}`),
  );
});

test("proof validator rejects a covered_by target that is not declared", async () => {
  const model = await buildProofModel();
  const endpoint = model.endpointLedger.find((row) => row.coverage?.kind === "write");
  assert.ok(endpoint);
  endpoint.links.write_actions = ["not-a-real-action"];

  assert.throws(
    () => validateProofModel(model),
    /unknown write action "not-a-real-action"/,
  );
});

test("proof validator rejects a command ledger that loses an operation alias", async () => {
  const model = await buildProofModel();
  const operationCommand = model.commandLedger.find((row) => row.operation);
  assert.ok(operationCommand);
  operationCommand.operation = "github.operation.that.does.not.exist";

  assert.throws(
    () => validateProofModel(model),
    /unknown operation "github\.operation\.that\.does\.not\.exist"/,
  );
});

test("proof validator does not accept an empty source bundle", () => {
  const model = buildProofModel(currentBundle());
  assert.throws(() => validateProofModel(model), /source bundle has no endpoints/);
});

test("binary reachability requires the exact rendered command name", () => {
  assert.deepEqual(
    classifyHelpResult("issue list", {
      code: 0,
      stdout: "NAME\n  pm github issue list - List issues\n",
      stderr: "",
    }),
    { state: "reachable", rendered_name: "pm github issue list" },
  );
  assert.deepEqual(
    classifyHelpResult("issue list", {
      code: 0,
      stdout: "NAME\n  pm github - GitHub command surface\n",
      stderr: "",
    }),
    { state: "unreachable", reason: "rendered namespace help instead of the declared command" },
  );
});

test("binary reachability requires implemented commands to hit the no-credential runtime boundary", () => {
  const implemented = { path: "issue list", availability: "implemented" };
  assert.deepEqual(
    classifyInvocationResult(implemented, { code: 1, stdout: "", stderr: "error: missing --credential" }),
    { state: "reachable", runtime_state: "missing_credential" },
  );
  assert.deepEqual(
    classifyInvocationResult(implemented, { code: 1, stdout: "", stderr: "error: unknown command \"issue list\"" }),
    { state: "unreachable", reason: "binary returned unknown command" },
  );
  assert.deepEqual(
    classifyInvocationResult(implemented, { code: 1, stdout: "", stderr: "missing --input" }),
    { state: "unreachable", reason: "implemented command did not reach the missing-credential boundary" },
  );
  assert.deepEqual(
    classifyInvocationResult({ path: "auth token", availability: "unsupported_local" }, { code: 1, stdout: "", stderr: "blocked" }),
    { state: "reachable", runtime_state: "declared_non_executable" },
  );
});
