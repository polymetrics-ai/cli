import assert from "node:assert/strict";
import test from "node:test";

import {
  enumerateImplementedCommands,
  redactForReport,
  validateProofRecords,
} from "../github-live-proof-sweep.mjs";

const fixtureSurface = {
  commands: [
    { path: "issue list", availability: "implemented" },
    { path: "repo view", availability: "implemented" },
    { path: "repo delete", availability: "blocked" },
  ],
};

test("enumerates every and only implemented GitHub command", () => {
  assert.deepEqual(enumerateImplementedCommands(fixtureSurface), [
    "issue list",
    "repo view",
  ]);
});

test("rejects an omitted command instead of treating a sample as a sweep", () => {
  assert.throws(
    () =>
      validateProofRecords(["issue list", "repo view"], [
        {
          command: "issue list",
          state: "proven",
          http_status: 200,
          assertion: { kind: "returned-data", subject: "issues", matched: true },
        },
      ]),
    /missing terminal result for "repo view"/,
  );
});

test("requires a returned-data assertion or a concrete untestable reason", () => {
  assert.doesNotThrow(() =>
    validateProofRecords(["issue list", "repo view"], [
      {
        command: "issue list",
        state: "proven",
        http_status: 200,
        assertion: { kind: "returned-data", subject: "issues", matched: true },
      },
      {
        command: "repo view",
        state: "untestable",
        reason:
          "requires GitHub Enterprise Server administration, which the dedicated test token cannot hold",
      },
    ]),
  );

  assert.throws(
    () =>
      validateProofRecords(["issue list"], [
        { command: "issue list", state: "covered-by-fixture" },
      ]),
    /invalid terminal state/,
  );
});

test("redacts raw subprocess output before it can become a report record", () => {
  const raw = "request failed with token ghp_fixture_token_should_not_escape";
  const record = redactForReport({
    command: "repo view",
    state: "failed",
    stderr: raw,
    reason: "request failed",
  });

  assert.equal(JSON.stringify(record).includes("ghp_fixture_token_should_not_escape"), false);
  assert.deepEqual(record, {
    command: "repo view",
    state: "failed",
    reason: "request failed",
  });
});
