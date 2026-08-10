import assert from "node:assert/strict";
import test from "node:test";
import { chmod, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import { existsSync } from "node:fs";
import { spawnSync } from "node:child_process";
import os from "node:os";
import path from "node:path";
import { fileURLToPath } from "node:url";

import {
  assertReadEnvelope,
  executeBarrier,
  enumerateImplementedCommands,
  redactForReport,
  runProcess,
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

test("releases all applicable operations from one barrier rather than serially", async () => {
  const events = [];
  const results = await executeBarrier(["repo view", "issue list", "label list"], async (command) => {
    events.push(`started:${command}`);
    await Promise.resolve();
    events.push(`finished:${command}`);
    return command;
  });

  assert.deepEqual(events.slice(0, 3).sort(), [
    "started:issue list",
    "started:label list",
    "started:repo view",
  ]);
  assert.deepEqual(results.sort(), ["issue list", "label list", "repo view"]);
});

test("bounds a non-terminating pm child before it can stall a live proof run", async () => {
  const temp = await mkdtemp(path.join(os.tmpdir(), "github-live-proof-timeout-"));
  try {
    const sleeper = path.join(temp, "sleeper.mjs");
    await writeFile(sleeper, "setInterval(() => {}, 1000);\n", "utf8");

    const result = await runProcess(process.execPath, [sleeper], temp, { timeoutMs: 100 });

    assert.equal(result.timedOut, true);
    assert.equal(result.timeoutMs, 100);
    assert.ok(["SIGTERM", "SIGKILL"].includes(result.signal));
  } finally {
    await rm(temp, { recursive: true, force: true });
  }
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

test("accepts stream and binary live results without inventing an HTTP status", () => {
  const stream = assertReadEnvelope(
    {
      kind: "ConnectorCommandRead",
      connector: "github",
      command: "issue list",
      stream: "issues",
      count: 2,
      records: [{ number: 5 }, { number: 6 }],
    },
    "issue list",
  );
  assert.equal(stream.httpStatus, undefined);
  assert.deepEqual(stream.assertion, {
    kind: "stream-records",
    subject: "issue list",
    matched: true,
    count: 2,
  });

  const binary = assertReadEnvelope(
    {
      kind: "ConnectorCommandBinaryDownload",
      connector: "github",
      command: "repo archive tarball",
      record: { file_name: "repo.tar.gz", file_size_bytes: 12 },
    },
    "repo archive tarball",
  );
  assert.equal(binary.httpStatus, undefined);
  assert.deepEqual(binary.assertion, {
    kind: "binary-download-record",
    subject: "repo archive tarball",
    matched: true,
  });

  assert.deepEqual(
    validateProofRecords(
      ["issue list", "repo archive tarball"],
      [
        { command: "issue list", state: "proven", assertion: stream.assertion },
        { command: "repo archive tarball", state: "proven", assertion: binary.assertion },
      ],
    ),
    { proven: 2, untestable: 0, failed: 0 },
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
    const boundaryPath = path.join(temp, "boundary.json");
    const reportPath = path.join(temp, "report.json");
    const root = path.join(temp, "project");
    await writeFile(
      fakePM,
      `#!/bin/sh\n: > ${JSON.stringify(marker)}\nprintf '{}\\n'\n`,
      "utf8",
    );
    await chmod(fakePM, 0o755);
    await writeFile(
      boundaryPath,
      JSON.stringify({
        schema_version: 1,
        run_id: "github-live-proof-sweep-test",
        default_deny: true,
        protected_owners: ["polymetrics-ai"],
        protected_repositories: [{ owner_slug: "polymetrics-ai", repo_slug: "cli" }],
        working_repositories: [{ owner_slug: "polymetrics-ai", repo_slug: "cli" }],
        allowed_targets: [
          {
            key: "organization",
            resource_type: "organization",
            org_slug: "polymetrics-cert",
            org_id: "O_certification",
            run_owned: true,
          },
          {
            key: "repository",
            resource_type: "repository",
            owner_slug: "polymetrics-cert",
            owner_id: "O_certification",
            repo_slug: "pm-live-lab-test",
            repo_id: "R_certification",
            run_owned: true,
          },
        ],
      }),
      "utf8",
    );

    const surface = JSON.parse(await (await import("node:fs/promises")).readFile(
      path.join(path.dirname(fileURLToPath(import.meta.url)), "../../internal/connectors/defs/github/cli_surface.json"),
      "utf8",
    ));
    const baseCases = surface.commands
      .filter((command) => command.availability === "implemented")
      .map((command) => ({
        command: command.path,
        untestable_reason:
          "requires a deliberately prepared live GitHub resource not created by this isolated test fixture",
      }));
    const runner = path.join(path.dirname(fileURLToPath(import.meta.url)), "../github-live-proof-sweep.mjs");
    for (const args of [
      ["--owner", "outside-the-dedicated-repository"],
      ["--repo=outside-the-dedicated-repository"],
      [
        "--config",
        "owner=outside-the-dedicated-repository",
        "--config",
        "repo=outside-the-dedicated-repository",
      ],
    ]) {
      const cases = baseCases.map((item) => ({ ...item }));
      const target = cases.find((item) => item.command === "repos create-using-template");
      target.untestable_reason = undefined;
      target.args = args;
      target.readback = { command: "repo view", args: [] };
      await writeFile(
        casesPath,
        JSON.stringify({
          connector: "github",
          test_repository: {
            owner: "polymetrics-cert",
            repo: "pm-live-lab-test",
            owner_id: "O_certification",
            repo_id: "R_certification",
            organization_id: "O_certification",
          },
          cases,
        }),
        "utf8",
      );

      const result = spawnSync(
        process.execPath,
        [
          runner,
          "--pm", fakePM,
          "--root", root,
          "--credential", "github-live-proof",
          "--boundary", boundaryPath,
          "--test-owner", "polymetrics-cert",
          "--test-repo", "pm-live-lab-test",
          "--cases", casesPath,
          "--report", reportPath,
          "--execute-writes",
        ],
        { encoding: "utf8" },
      );

      assert.notEqual(result.status, 0);
      assert.match(`${result.stdout}\n${result.stderr}`, /dedicated repository|--config (owner|repo)/i);
      assert.equal(existsSync(marker), false, "case validation must finish before pm starts");
    }

    const cases = baseCases.map((item) => ({ ...item }));
    const target = cases.find((item) => item.command === "repos create-using-template");
    target.untestable_reason = undefined;
    target.args = ["--owner", "polymetrics-cert", "--repo", "pm-live-lab-test"];
    await writeFile(
      casesPath,
      JSON.stringify({
        connector: "github",
        test_repository: {
          owner: "polymetrics-cert",
          repo: "pm-live-lab-test",
          owner_id: "O_certification",
          repo_id: "R_certification",
          organization_id: "O_certification",
        },
        cases,
      }),
      "utf8",
    );
    const result = spawnSync(
      process.execPath,
      [
        runner,
        "--pm", fakePM,
        "--root", root,
        "--credential", "github-live-proof",
        "--boundary", boundaryPath,
        "--cases", casesPath,
        "--report", reportPath,
        "--execute-writes",
      ],
      { encoding: "utf8" },
    );
    assert.notEqual(result.status, 0);
    assert.match(`${result.stdout}\n${result.stderr}`, /requires an independent readback/i);
    assert.equal(existsSync(marker), false, "write cases without a read-back must fail before pm starts");
  } finally {
    await rm(temp, { recursive: true, force: true });
  }
});

test("records every implemented command as terminally untestable for an external live blocker", async () => {
  const temp = await mkdtemp(path.join(os.tmpdir(), "github-live-proof-blocker-"));
  try {
    const runner = path.join(path.dirname(fileURLToPath(import.meta.url)), "../github-live-proof-sweep.mjs");
    const reportPath = path.join(temp, "report.json");
    const result = spawnSync(
      process.execPath,
      [
        runner,
        "--external-blocker",
        "--pm",
        process.execPath,
        "--report",
        reportPath,
        "--reason",
        "approved private GitHub credential and dedicated test repository are unavailable in this isolated worktree",
      ],
      { encoding: "utf8" },
    );
    assert.equal(result.status, 0, `${result.stdout}\n${result.stderr}`);
    const report = JSON.parse(await readFile(reportPath, "utf8"));
    const expected = enumerateImplementedCommands(
      JSON.parse(await readFile(path.join(path.dirname(fileURLToPath(import.meta.url)), "../../internal/connectors/defs/github/cli_surface.json"), "utf8")),
    );
    assert.equal(report.status, "external_blocker");
    assert.deepEqual(report.tally, { proven: 0, untestable: expected.length, failed: 0 });
    assert.equal(report.records.length, expected.length);
    assert.equal(report.records.every((record) => record.state === "untestable"), true);
  } finally {
    await rm(temp, { recursive: true, force: true });
  }
});
