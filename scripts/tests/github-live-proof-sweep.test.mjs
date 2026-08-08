import assert from "node:assert/strict";
import test from "node:test";
import { chmod, mkdtemp, rm, writeFile } from "node:fs/promises";
import { existsSync } from "node:fs";
import { spawnSync } from "node:child_process";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

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

test("rejects a write case that overrides the dedicated repository owner before starting pm", async () => {
  const temp = await mkdtemp(path.join(os.tmpdir(), "github-live-proof-case-"));
  try {
    const marker = path.join(temp, "pm-was-started");
    const fakePM = path.join(temp, "fake-pm");
    const casesPath = path.join(temp, "cases.json");
    const reportPath = path.join(temp, "report.json");
    const root = path.join(temp, "project");
    await writeFile(
      fakePM,
      `#!/bin/sh\n: > ${JSON.stringify(marker)}\nprintf '{}\\n'\n`,
      "utf8",
    );
    await chmod(fakePM, 0o755);

    const surface = JSON.parse(await (await import("node:fs/promises")).readFile(
      path.join(path.dirname(fileURLToPath(import.meta.url)), "../../internal/connectors/defs/github/cli_surface.json"),
      "utf8",
    ));
    const cases = surface.commands
      .filter((command) => command.availability === "implemented")
      .map((command) => ({
        command: command.path,
        untestable_reason:
          "requires a deliberately prepared live GitHub resource not created by this isolated test fixture",
      }));
    const target = cases.find((item) => item.command === "repos create-using-template");
    target.untestable_reason = undefined;
    target.args = ["--owner", "outside-the-dedicated-repository"];
    await writeFile(
      casesPath,
      JSON.stringify({
        connector: "github",
        test_repository: { owner: "dedicated-owner", repo: "dedicated-repo" },
        cases,
      }),
      "utf8",
    );

    const runner = path.join(path.dirname(fileURLToPath(import.meta.url)), "../github-live-proof-sweep.mjs");
    const result = spawnSync(
      process.execPath,
      [
        runner,
        "--pm", fakePM,
        "--root", root,
        "--credential", "github-live-proof",
        "--test-owner", "dedicated-owner",
        "--test-repo", "dedicated-repo",
        "--cases", casesPath,
        "--report", reportPath,
        "--execute-writes",
      ],
      { encoding: "utf8" },
    );

    assert.notEqual(result.status, 0);
    assert.match(`${result.stdout}\n${result.stderr}`, /dedicated repository owner/i);
    assert.equal(existsSync(marker), false, "case validation must finish before pm starts");
  } finally {
    await rm(temp, { recursive: true, force: true });
  }
});
