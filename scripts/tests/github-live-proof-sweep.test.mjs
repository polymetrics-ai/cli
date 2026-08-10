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
  validateCredentialScope,
  validateProofReport,
  validateProofRecords,
} from "../github-live-proof-sweep.mjs";
import { buildCases, resolveLiveBoundary } from "../github-live-cases.mjs";

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

test("redacts every config value by key before it can enter a proof record", () => {
  const record = redactForReport({
    command: "repo view",
    state: "failed",
    invocation: [
      "pm",
      "github",
      "repo",
      "view",
      "--config",
      "private_key=synthetic-private-material",
      "--config=base_url=https://synthetic.invalid",
    ],
    reason: "the synthetic command was rejected before a provider request",
  });

  assert.deepEqual(record.invocation, [
    "pm",
    "github",
    "repo",
    "view",
    "--config",
    "private_key=<redacted>",
    "--config=base_url=<redacted>",
  ]);
  assert.equal(JSON.stringify(record).includes("synthetic-private-material"), false);
  assert.equal(JSON.stringify(record).includes("https://synthetic.invalid"), false);
});

test("requires the canonical origin and a fully paginated App installation repository identity", async () => {
  const temp = await mkdtemp(path.join(os.tmpdir(), "github-live-proof-preflight-"));
  try {
    const fakePM = path.join(temp, "fake-pm.mjs");
    const logPath = path.join(temp, "invocations.jsonl");
    const boundary = {
      schema_version: 1,
      run_id: "github-live-proof-preflight-test",
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
    };
    const expectedBoundary = resolveLiveBoundary(boundary);
    const credential = {
      kind: "Credential",
      credential: {
        connector: "github",
        config: {
          owner: "polymetrics-cert",
          repo: "pm-live-lab-test",
          base_url: "https://api.github.com/",
          auth_type: "github_app",
          app_id: "12345",
          installation_id: "67890",
        },
        secret_fields: ["private_key"],
      },
    };
    const outsideOriginCredential = structuredClone(credential);
    outsideOriginCredential.credential.config.base_url = "https://outside.example";
    const firstPage = {
      kind: "ConnectorCommandDirectRead",
      connector: "github",
      command: "apps list-repos-accessible-to-installation",
      status: 200,
      response: {
        total_count: 2,
        repositories: [
          {
            id: "R_other",
            full_name: "polymetrics-cert/another-repository",
            owner: { login: "polymetrics-cert", id: "O_certification" },
          },
        ],
      },
      page: {
        strategy: "page_number",
        records: 1,
        size: 1,
        number: 1,
        has_more: true,
        next_number: 2,
        complete: false,
        reason: "more_pages",
      },
    };
    const secondPage = {
      kind: "ConnectorCommandDirectRead",
      connector: "github",
      command: "apps list-repos-accessible-to-installation",
      status: 200,
      response: {
        total_count: 2,
        repositories: [
          {
            id: "R_certification",
            full_name: "polymetrics-cert/pm-live-lab-test",
            owner: { login: "polymetrics-cert", id: "O_certification" },
          },
        ],
      },
      page: {
        strategy: "page_number",
        records: 1,
        size: 1,
        number: 2,
        has_more: false,
        complete: true,
      },
    };
    await writeFile(fakePM, [
      "#!/usr/bin/env node",
      "import { appendFileSync } from 'node:fs';",
      `const logPath = ${JSON.stringify(logPath)};`,
      `const credential = ${JSON.stringify(credential)};`,
      `const outsideOriginCredential = ${JSON.stringify(outsideOriginCredential)};`,
      `const firstPage = ${JSON.stringify(firstPage)};`,
      `const secondPage = ${JSON.stringify(secondPage)};`,
      "const args = process.argv.slice(2);",
      "appendFileSync(logPath, JSON.stringify(args) + '\\n');",
      "if (args[0] === 'credentials' && args[1] === 'inspect') {",
      "  process.stdout.write(JSON.stringify(args[2] === 'outside-origin' ? outsideOriginCredential : credential) + '\\n');",
      "  process.exit(0);",
      "}",
      "if (args[0] === 'github' && args[1] === 'apps' && args[2] === 'list-repos-accessible-to-installation') {",
      "  process.stdout.write(JSON.stringify(args.includes('--page') ? secondPage : firstPage) + '\\n');",
      "  process.exit(0);",
      "}",
      "process.exit(1);",
    ].join("\n"), "utf8");
    await chmod(fakePM, 0o755);
    const surface = {
      commands: [
        {
          path: "apps list-repos-accessible-to-installation",
          availability: "implemented",
          intent: "direct_read",
          flags: [],
          api_surface: [{ method: "GET", path: "/installation/repositories" }],
        },
      ],
    };

    const result = await validateCredentialScope({
      binary: fakePM,
      root: temp,
      credential: "github-live-proof",
      boundary: expectedBoundary,
      surface,
      cwd: temp,
    });
    assert.deepEqual(result.record.assertion, {
      kind: "app-installation-repository-boundary",
      subject: "apps list-repos-accessible-to-installation",
      matched: true,
      pages: 2,
    });

    const invocations = (await readFile(logPath, "utf8")).trim().split("\n").map((line) => JSON.parse(line));
    const preflights = invocations.filter((args) => args[0] === "github");
    assert.equal(preflights.length, 2);
    assert.equal(preflights[0].includes("--page"), false);
    assert.equal(preflights[1][preflights[1].indexOf("--page") + 1], "2");
    assert.equal(preflights[1].includes("--json"), true);

    await assert.rejects(
      validateCredentialScope({
        binary: fakePM,
        root: temp,
        credential: "outside-origin",
        boundary: expectedBoundary,
        surface,
        cwd: temp,
      }),
      /canonical GitHub API origin/i,
    );
    const afterRejectedOrigin = (await readFile(logPath, "utf8")).trim().split("\n").map((line) => JSON.parse(line));
    assert.equal(afterRejectedOrigin.filter((args) => args[0] === "github").length, 2);
  } finally {
    await rm(temp, { recursive: true, force: true });
  }
});

test("rejects unsafe boundary and report scalars without retaining their contents", () => {
  const secret = "-----BEGIN PRIVATE KEY-----\nnot-a-key\n-----END PRIVATE KEY-----";
  const report = {
    schema_version: 1,
    connector: "github",
    status: "credentialed_live",
    generated_at: "2026-08-11T00:00:00.000Z",
    surface_sha256: "a".repeat(64),
    binary_sha256: "b".repeat(64),
    case_file_sha256: "c".repeat(64),
    test_repository: "<run-owned-boundary-repository>",
    run_boundary: {
      run_id: secret,
      owner: "polymetrics-cert",
      repo: "pm-live-lab-test",
      owner_id: "O_certification",
      repo_id: "R_certification",
      organization_id: "O_certification",
    },
    launch: { strategy: "single_barrier_release", operations_released: 0 },
    implemented_commands: 1,
    tally: { proven: 0, untestable: 1, failed: 0 },
    records: [
      {
        command: "repo view",
        state: "untestable",
        reason: "the immutable boundary fixture is unavailable for this controlled proof",
      },
    ],
  };
  let rejected;
  try {
    validateProofReport(report, ["repo view"]);
  } catch (error) {
    rejected = error;
  }
  assert.ok(rejected instanceof Error);
  assert.equal(rejected.message.includes(secret), false);
  assert.match(rejected.message, /unsafe|report|boundary/i);
  assert.throws(
    () => validateProofReport({ ...report, unexpected: "field" }, ["repo view"]),
    /unsupported field/i,
  );
});

test("rejects untyped target, transport, secret, and lifecycle case inputs before starting pm", async () => {
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
    const boundary = {
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
    };
    await writeFile(boundaryPath, JSON.stringify(boundary), "utf8");

    const surface = JSON.parse(await (await import("node:fs/promises")).readFile(
      path.join(path.dirname(fileURLToPath(import.meta.url)), "../../internal/connectors/defs/github/cli_surface.json"),
      "utf8",
    ));
    const baseCases = buildCases(surface, resolveLiveBoundary(boundary));
    const runner = path.join(path.dirname(fileURLToPath(import.meta.url)), "../github-live-proof-sweep.mjs");
    const attempts = [
      {
        command: "orgs update",
        args: ["--org", "outside-the-run-owned-organization"],
        expected: /remains untestable/i,
      },
      {
        command: "orgs delete",
        args: ["--org=outside-the-run-owned-organization"],
        expected: /remains untestable/i,
      },
      {
        command: "orgs list-members",
        args: ["--org", "outside-the-run-owned-organization"],
        expected: /source-derived command descriptor and run-owned boundary/i,
      },
      {
        command: "repo view",
        args: ["--config=base_url=https://synthetic.invalid"],
        expected: /may not override connector configuration/i,
      },
      {
        command: "orgs update",
        args: ["--config", "private_key=synthetic-private-material"],
        expected: /secret --config key/i,
        absent: "synthetic-private-material",
      },
      {
        command: "repo view",
        args: ["--plan=existing-plan"],
        expected: /lifecycle or credential flags/i,
      },
      {
        command: "orgs update",
        args: ["--org", "polymetrics-cert"],
        readback: { command: "repo view", args: ["--approve=existing-grant"] },
        expected: /lifecycle or credential flags/i,
      },
      {
        command: "repo view",
        args: [],
        expected: /unsupported field/i,
        mutate: (target) => { target.undeclared = "field"; },
      },
      {
        command: "repo view",
        args: [],
        expected: /unsafe credential-like material/i,
        absent: "-----BEGIN PRIVATE KEY-----",
        mutate: (target) => {
          delete target.args;
          target.untestable_reason = "-----BEGIN PRIVATE KEY-----\\nsynthetic\\n-----END PRIVATE KEY-----";
        },
      },
    ];
    for (const attempt of attempts) {
      const cases = JSON.parse(JSON.stringify(baseCases));
      const target = cases.cases.find((item) => item.command === attempt.command);
      delete target.untestable_reason;
      target.args = attempt.args;
      if (attempt.readback) target.readback = attempt.readback;
      if (attempt.mutate) attempt.mutate(target);
      await writeFile(
        casesPath,
        JSON.stringify(cases),
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
      const output = `${result.stdout}\n${result.stderr}`;
      assert.match(output, attempt.expected);
      if (attempt.absent) assert.equal(output.includes(attempt.absent), false);
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
        "app_installation_credential_unavailable",
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
    assert.deepEqual(report.blocker, {
      code: "app_installation_credential_unavailable",
      message: "The captain-authorized GitHub App installation credential is unavailable to this proof runner.",
    });

    const rejected = spawnSync(
      process.execPath,
      [
        runner,
        "--external-blocker",
        "--pm",
        process.execPath,
        "--report",
        reportPath,
        "--reason",
        "unapproved free-form blocker narrative",
      ],
      { encoding: "utf8" },
    );
    assert.notEqual(rejected.status, 0);
    assert.match(`${rejected.stdout}\n${rejected.stderr}`, /external blocker code/i);
    assert.equal(`${rejected.stdout}\n${rejected.stderr}`.includes("unapproved free-form blocker narrative"), false);
  } finally {
    await rm(temp, { recursive: true, force: true });
  }
});
