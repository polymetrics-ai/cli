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
  enumerateImplementedCommands,
  redactForReport,
  runWriteLifecycle,
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

test("write lifecycle carries the approval only through child stdin", async () => {
  const temp = await mkdtemp(path.join(os.tmpdir(), "github-live-proof-stdin-"));
  try {
    const binary = path.join(temp, "fake-pm");
    const argvPath = path.join(temp, "argv");
    const stdinPath = path.join(temp, "stdin");
    await writeFile(
      binary,
      `#!/bin/sh\ncase "$*" in\n  *--approval-token-stdin*)\n    printf '%s\\n' "$@" > ${JSON.stringify(argvPath)}\n    cat > ${JSON.stringify(stdinPath)}\n    printf '{"kind":"ReverseRun","run":{"status":"completed","records_succeeded":1,"records_failed":0,"operation_direct_write":{"status":200}}}\\n'\n    ;;\n  *--preview*)\n    printf 'Approval token: transient-grant\\n'\n    ;;\n  *)\n    printf 'Created connector command plan plan-test-id\\n'\n    ;;\nesac\n`,
      "utf8",
    );
    await chmod(binary, 0o755);

    const result = await runWriteLifecycle({
      binary,
      root: path.join(temp, "project"),
      credential: "github-live-proof",
      command: "issue close",
      args: ["--issue-number", "42"],
      cwd: temp,
    });
    assert.equal(result.httpStatus, 200);
    const argv = await readFile(argvPath, "utf8");
    assert.equal(argv.includes("--approval-token-stdin"), true);
    assert.equal(argv.includes("--approve"), false);
    assert.equal(argv.includes("transient-grant"), false);
    assert.equal(await readFile(stdinPath, "utf8"), "transient-grant\n");
  } finally {
    await rm(temp, { recursive: true, force: true });
  }
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
    const baseCases = surface.commands
      .filter((command) => command.availability === "implemented")
      .map((command) => ({
        command: command.path,
        untestable_reason:
          "requires a deliberately prepared live GitHub resource not created by this isolated test fixture",
      }));
    const runner = path.join(path.dirname(fileURLToPath(import.meta.url)), "../github-live-proof-sweep.mjs");
    for (const { args, want } of [
      {
        args: ["--owner", "outside-the-dedicated-repository"],
        want: /dedicated repository owner/i,
      },
      {
        args: ["--repo=outside-the-dedicated-repository"],
        want: /dedicated repository repo/i,
      },
      {
        args: ["--approval-token-stdin"],
        want: /may not override lifecycle or credential flags/i,
      },
      {
        args: ["--approve=argv-value"],
        want: /may not override lifecycle or credential flags/i,
      },
    ]) {
      const cases = baseCases.map((item) => ({ ...item }));
      const target = cases.find((item) => item.command === "repos create-using-template");
      target.untestable_reason = undefined;
      target.args = args;
      await writeFile(
        casesPath,
        JSON.stringify({
          connector: "github",
          test_repository: { owner: "karthik-sivadas", repo: "pm-live-test-direct-read-20260808081515" },
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
          "--test-owner", "karthik-sivadas",
          "--test-repo", "pm-live-test-direct-read-20260808081515",
          "--cases", casesPath,
          "--report", reportPath,
          "--execute-writes",
        ],
        { encoding: "utf8" },
      );

      assert.notEqual(result.status, 0);
      assert.match(`${result.stdout}\n${result.stderr}`, want);
      assert.equal(existsSync(marker), false, "case validation must finish before pm starts");
    }
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
