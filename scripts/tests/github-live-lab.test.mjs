import assert from "node:assert/strict";
import { chmod, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import os from "node:os";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

import {
  appendCleanupEntry,
  capturePMErrorEnvelope,
  authorizeAccountBootstrapProbe,
  assertPMOnly,
  assertAccountBootstrapProbeInvocation,
  authorizeBootstrapRepoCreate,
  authorizeBootstrapRepoDiscovery,
  authorizeLabTarget,
  executeBootstrapPMFixtureStep,
  executeBootstrapPMDiscoveryStep,
  executePMFixtureStep,
  readCleanupLedger,
  assertBoundLabLabelAbsent,
  assertBoundLabLabelProperties,
  assertBoundLabDeployKeyAbsent,
  assertBoundLabIssueAbsent,
  resolveBoundLabDeployKey,
  resolveBoundLabLabel,
  resolveBoundLabIssue,
  resolveBootstrapRepositoryTarget,
  runPMProcess,
  runPMScopedRead,
  runPMAccountBootstrapProbe,
  runPMPlannedWrite,
  waitForBoundLabDeployKeyAbsent,
  waitForBoundLabDeployKey,
  waitForBoundLabIssue,
  validateCleanupLedger,
  validateLabBoundary,
  validateLabTerminalRecords,
} from "../github-live-lab.mjs";
import {
  buildBootstrapProbeInventory,
  validateBootstrapProbeInventory,
} from "../github-live-bootstrap-probes.mjs";
import {
  buildLabManifest,
  classifyLabCohort,
  validateLabManifest,
} from "../github-live-lab-manifest.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const phaseDir = path.join(root, ".planning/phases/github-parity-extract-r1");

const boundary = {
  schema_version: 1,
  run_id: "github-live-lab-test-20260809",
  default_deny: true,
  protected_owners: ["polymetrics-ai"],
  protected_repositories: [
    { owner_slug: "polymetrics-ai", repo_slug: "cli" },
  ],
  working_repositories: [
    { owner_slug: "polymetrics-ai", repo_slug: "cli" },
  ],
  allowed_targets: [
    {
      key: "personal_repo",
      resource_type: "repository",
      owner_slug: "lab-owner",
      owner_id: "O_lab_owner",
      repo_slug: "pm-live-lab-001",
      repo_id: "R_lab_repository",
      run_owned: true,
    },
  ],
};

async function loadJSON(relativePath) {
  return JSON.parse(await readFile(path.join(root, relativePath), "utf8"));
}

test("derives exactly the historical 957 pre-skipped cases into mutually exclusive PM-only cohorts", async () => {
  const surface = await loadJSON("internal/connectors/defs/github/cli_surface.json");
  const cases = await loadJSON(".planning/phases/github-parity-extract-r1/LIVE-PROOF-CASES.json");
  const manifest = buildLabManifest({ surface, cases, generatedAt: "2026-08-09T00:00:00.000Z" });
  const historical = cases.cases
    .filter((item) => typeof item.untestable_reason === "string")
    .map((item) => item.command)
    .sort();

  assert.equal(historical.length, 957);
  assert.equal(manifest.rows.length, 957);
  assert.deepEqual(manifest.rows.map((row) => row.command).sort(), historical);
  assert.deepEqual(
    Object.keys(manifest.class_tally).sort(),
    ["github_app_or_marketplace", "personal_repo", "sandbox_org_free", "unavailable_entitlement"],
  );
  assert.equal(Object.values(manifest.class_tally).reduce((sum, count) => sum + count, 0), 957);

  for (const row of manifest.rows) {
    assert.equal(typeof row.case_id, "string");
    assert.equal(typeof row.historical_reason, "string");
    assert.equal(typeof row.credential.class, "string");
    assert.equal(typeof row.plan_feature, "string");
    assert.equal(typeof row.target_allowlist_entry, "string");
    assert.equal(typeof row.destructive_acknowledgement, "string");
    assert.equal(typeof row.residual_state_check, "string");
    assert.equal(typeof row.earliest_divergence, "string");
    assert.equal(typeof row.cleanup_strategy, "string");
    assert.equal(Array.isArray(row.setup_pm), true);
    assert.match(row.test_pm, /^pm github /);
    assert.match(row.assert_pm, /^pm github /);
    assert.equal(row.cleanup_pm === null || row.cleanup_pm.startsWith("pm github "), true);
    assert.equal(row.cohort in manifest.class_tally, true);
    assert.equal(/(?:^|\s)(?:gh|curl)(?:\s|$)|https?:\/\/|browser/i.test(JSON.stringify(row)), false);
  }

  assert.doesNotThrow(() => validateLabManifest({ manifest, surface, cases }));

  const issueCreate = manifest.rows.find((row) => row.command === "issue create");
  assert.equal(issueCreate?.cleanup_strategy, "neutralize_and_retain");
  assert.match(issueCreate?.cleanup_pm || "", /^pm github issue close /);
});

test("classifies personal, sandbox organization, App, and unavailable-entitlement rows by provider requirement", () => {
  assert.equal(
    classifyLabCohort({ path: "issue create", intent: "reverse_etl", api_surface: [{ method: "POST", path: "/repos/{owner}/{repo}/issues" }] }).cohort,
    "personal_repo",
  );
  assert.equal(
    classifyLabCohort({ path: "orgs create-webhook", intent: "direct_write", api_surface: [{ method: "POST", path: "/orgs/{org}/hooks" }] }).cohort,
    "sandbox_org_free",
  );
  assert.equal(
    classifyLabCohort({ path: "apps get-authenticated", intent: "direct_read", api_surface: [{ method: "GET", path: "/app" }] }).cohort,
    "github_app_or_marketplace",
  );
  assert.equal(
    classifyLabCohort({ path: "codespaces create", intent: "direct_write", api_surface: [{ method: "POST", path: "/user/codespaces" }] }).cohort,
    "unavailable_entitlement",
  );
});

test("lab boundary default-denies protected, working, unresolved, ambiguous, and slug-ID-mismatched writes", () => {
  assert.doesNotThrow(() => validateLabBoundary(boundary));
  assert.doesNotThrow(() => authorizeLabTarget(boundary, {
    resource_type: "repository",
    owner_slug: "lab-owner",
    owner_id: "O_lab_owner",
    repo_slug: "pm-live-lab-001",
    repo_id: "R_lab_repository",
  }));

  for (const target of [
    {
      resource_type: "repository",
      owner_slug: "polymetrics-ai",
      owner_id: "O_production",
      repo_slug: "cli",
      repo_id: "R_production",
    },
    {
      resource_type: "repository",
      owner_slug: "lab-owner",
      owner_id: "O_lab_owner",
      repo_slug: "pm-live-lab-001",
    },
    {
      resource_type: "repository",
      owner_slug: "lab-owner",
      owner_id: "O_other_owner",
      repo_slug: "pm-live-lab-001",
      repo_id: "R_lab_repository",
    },
  ]) {
    assert.throws(() => authorizeLabTarget(boundary, target), /denied|immutable|ambiguous|match/i);
  }

  const ambiguous = structuredClone(boundary);
  ambiguous.allowed_targets.push({ ...boundary.allowed_targets[0], key: "same-target-again" });
  assert.throws(() => validateLabBoundary(ambiguous), /ambiguous/i);
});

test("a denied target cannot reach the PM fixture executor", async () => {
  let started = false;
  await assert.rejects(
    executePMFixtureStep({
      boundary,
      target: {
        resource_type: "repository",
        owner_slug: "polymetrics-ai",
        owner_id: "O_production",
        repo_slug: "cli",
        repo_id: "R_production",
      },
      invocation: "pm github repo view --json",
      execute: async () => {
        started = true;
      },
    }),
    /denied/i,
  );
  assert.equal(started, false);
});

test("archived cleanup evidence never reauthorizes a retired provider target", async () => {
  const current = structuredClone(boundary);
  const retiredTarget = structuredClone(boundary.allowed_targets[0]);
  current.run_id = "github-live-lab-current-20260810";
  current.allowed_targets = [{
    ...retiredTarget,
    owner_slug: "captain-owner",
    owner_id: "O_captain_owner",
    repo_slug: "pm-live-test-current",
    repo_id: "R_captain_current",
  }];
  current.historical_runs = [{ run_id: boundary.run_id, target: retiredTarget }];
  const entries = [
    { schema_version: 1, run_id: boundary.run_id, action: "initialized", note: "retired run" },
    {
      schema_version: 1,
      run_id: boundary.run_id,
      action: "created",
      fixture_id: "repository:R_lab_repository",
      target: retiredTarget,
      pm_command: "pm github repo create --name pm-live-lab-001",
      provider_id: "R_lab_repository",
    },
  ];

  assert.doesNotThrow(() => validateLabBoundary(current));
  assert.doesNotThrow(() => validateCleanupLedger({ entries, boundary: current }));
  assert.throws(() => authorizeLabTarget(current, retiredTarget), /denied|match|allowlist/i);

  let started = false;
  await assert.rejects(
    runPMScopedRead({
      binary: "pm",
      root: "/tmp/github-live-lab-root",
      credentialName: "lab-credential",
      command: "issue list",
      commandArgs: ["--config", "owner=lab-owner", "--config", "repo=pm-live-lab-001"],
      boundary: current,
      target: retiredTarget,
      run: async () => {
        started = true;
        throw new Error("retired target must not reach PM");
      },
    }),
    /denied|match|allowlist/i,
  );
  assert.equal(started, false);
});

test("a bootstrap principal may create exactly one private run-owned repository and nothing else", async () => {
  const bootstrapBoundary = structuredClone(boundary);
  bootstrapBoundary.bootstrap_principals = [
    {
      key: "personal_private_repo_bootstrap",
      resource_type: "authenticated_user",
      user_slug: "lab-owner",
      user_id: "U_lab_owner",
      allowed_command: "repo create",
      requested_repo_slug: "pm-live-lab-001",
      required_private: true,
      required_auto_init: true,
      purpose: "create exactly one run-owned private lab repository",
    },
  ];
  const request = {
    command: "repo create",
    principal: { user_slug: "lab-owner", user_id: "U_lab_owner" },
    record: { name: "pm-live-lab-001", private: true, auto_init: true },
  };
  assert.doesNotThrow(() => authorizeBootstrapRepoCreate(bootstrapBoundary, request));
  for (const invalid of [
    { ...request, command: "repo delete" },
    { ...request, principal: { user_slug: "lab-owner", user_id: "U_other" } },
    { ...request, record: { ...request.record, name: "not-run-owned" } },
    { ...request, record: { ...request.record, private: false } },
  ]) {
    assert.throws(() => authorizeBootstrapRepoCreate(bootstrapBoundary, invalid), /denied|immutable|private|bootstrap/i);
  }

  let started = false;
  await assert.rejects(
    executeBootstrapPMFixtureStep({
      boundary: bootstrapBoundary,
      request: { ...request, record: { ...request.record, name: "outside-boundary" } },
      invocation: "pm github repo create --name outside-boundary --private --auto-init",
      execute: async () => {
        started = true;
      },
    }),
    /denied|bootstrap/i,
  );
  assert.equal(started, false);
});

test("bootstrap discovery accepts only one private PM-listed repository under the authenticated immutable user", async () => {
  const bootstrapBoundary = structuredClone(boundary);
  bootstrapBoundary.allowed_targets = [];
  bootstrapBoundary.bootstrap_principals = [
    {
      key: "personal_private_repo_bootstrap",
      resource_type: "authenticated_user",
      user_slug: "lab-owner",
      user_id: "U_lab_owner",
      allowed_command: "repo create",
      requested_repo_slug: "pm-live-lab-001",
      required_private: true,
      required_auto_init: true,
      purpose: "create exactly one run-owned private lab repository",
    },
  ];
  const request = {
    command: "repos list-for-authenticated-user",
    principal: { user_slug: "lab-owner", user_id: "U_lab_owner" },
    repository: { owner_slug: "lab-owner", repo_slug: "pm-live-lab-001" },
  };
  assert.doesNotThrow(() => authorizeBootstrapRepoDiscovery(bootstrapBoundary, request));
  assert.deepEqual(
    resolveBootstrapRepositoryTarget({
      envelope: {
        kind: "ConnectorCommandDirectRead",
        connector: "github",
        command: "repos list-for-authenticated-user",
        status: 200,
        response: [
          {
            id: 12345,
            name: "pm-live-lab-001",
            private: true,
            owner: { login: "lab-owner", id: "U_lab_owner" },
          },
        ],
      },
      principal: request.principal,
      repository: request.repository,
    }),
    {
      resource_type: "repository",
      owner_slug: "lab-owner",
      owner_id: "U_lab_owner",
      repo_slug: "pm-live-lab-001",
      repo_id: "12345",
      run_owned: true,
    },
  );

  for (const invalid of [
    { ...request, command: "repo delete" },
    { ...request, principal: { user_slug: "lab-owner", user_id: "U_other" } },
    { ...request, repository: { owner_slug: "lab-owner", repo_slug: "outside-boundary" } },
  ]) {
    assert.throws(() => authorizeBootstrapRepoDiscovery(bootstrapBoundary, invalid), /denied|immutable|bootstrap|discovery/i);
  }

  for (const response of [
    [
      { id: 12345, name: "pm-live-lab-001", private: false, owner: { login: "lab-owner", id: "U_lab_owner" } },
    ],
    [
      { id: 12345, name: "pm-live-lab-001", private: true, owner: { login: "lab-owner", id: "U_other" } },
    ],
    [
      { id: 12345, name: "pm-live-lab-001", private: true, owner: { login: "lab-owner", id: "U_lab_owner" } },
      { id: 67890, name: "pm-live-lab-001", private: true, owner: { login: "lab-owner", id: "U_lab_owner" } },
    ],
  ]) {
    assert.throws(
      () => resolveBootstrapRepositoryTarget({
        envelope: { kind: "ConnectorCommandDirectRead", connector: "github", command: "repos list-for-authenticated-user", status: 200, response },
        principal: request.principal,
        repository: request.repository,
      }),
      /private|authenticated|exactly one|immutable/i,
    );
  }

  let started = false;
  await assert.rejects(
    executeBootstrapPMDiscoveryStep({
      boundary: bootstrapBoundary,
      request,
      invocation: "pm github repos list-for-authenticated-user --credential lab-credential --config owner=lab-owner --json",
      execute: async () => {
        started = true;
      },
    }),
    /denied|bootstrap|discovery/i,
  );
  assert.equal(started, false);
});

test("retains the credential-pinned repo-view control and sanitized owner/repo flag regression", async () => {
  const [surface, report, divergences] = await Promise.all([
    loadJSON("internal/connectors/defs/github/cli_surface.json"),
    loadJSON(".planning/phases/github-parity-extract-r1/LIVE-PROOF-REPORT.json"),
    loadJSON(".planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-DIVERGENCES.json"),
  ]);
  const repoView = surface.commands.find((command) => command.path === "repo view");
  assert.deepEqual(repoView?.flags || [], []);
  const control = report.records.find((record) => record.command === "repo view" && record.state === "proven");
  assert.equal(Array.isArray(control?.invocation), true);
  assert.deepEqual(control.invocation.slice(0, 4), ["pm", "github", "repo", "view"]);
  assert.equal(control.invocation.includes("--credential"), true);
  assert.equal(control.invocation.includes("--root"), true);
  assert.equal(control.invocation.includes("--json"), true);
  assert.equal(control.invocation.some((argument) => ["--owner", "--repo", "--config", "--connection"].includes(argument)), false);

  const regression = divergences.records.find((record) => record.id === "repo-view-owner-repo-flags-rejected");
  assert.deepEqual(
    {
      command: regression?.command,
      outcome: regression?.outcome,
      provider_request_started: regression?.provider_request_started,
    },
    { command: "repo view", outcome: "flag_parse_rejected", provider_request_started: false },
  );
  assert.match(regression.control_invocation_shape, /--credential <profile> --root <isolated-project> --json/);
  assert.match(regression.malformed_invocation_shape, /--owner <owner> --repo <repo>/);
  assert.equal(/pm-live-test-direct-read|github-live-proof|https?:\/\//i.test(JSON.stringify(regression)), false);
});

test("records personal-repository cohort results only after immutable target binding and independently verified lifecycle state", async () => {
  const [manifest, boundaryFile, report, probes, entries] = await Promise.all([
    loadJSON(".planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-MANIFEST.json"),
    loadJSON(".planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOUNDARY.json"),
    loadJSON(".planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-REPORT.json"),
    loadJSON(".planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOOTSTRAP-PROBES.json"),
    readCleanupLedger(path.join(phaseDir, "GITHUB-LIVE-LAB-CLEANUP.jsonl")),
  ]);
  const terminal = validateLabTerminalRecords(report.terminal_records);
  assert.deepEqual(terminal, { terminal_records: 11 });
  assert.deepEqual(report.tally, {
    terminal_records: 11,
    proven: 9,
    failed: 1,
    credential_blocker: 1,
    entitlement_blocker: 0,
  });
  const results = new Map(report.terminal_records.map((record) => [record.command, record]));
  for (const command of [
    "repo create",
    "issue create",
    "apps get-authenticated",
    "apps list-subscriptions-for-authenticated-user",
    "label create",
    "label edit",
    "label delete",
    "issue edit",
    "issue comment",
    "repo deploy-key add",
    "repo deploy-key delete",
  ]) {
    const result = results.get(command);
    const manifestRow = manifest.rows.find((row) => row.case_id === result?.case_id);
    const expected = command.startsWith("apps ")
      ? { command, state: command === "apps get-authenticated" ? "credential_blocker" : "proven", cohort: "github_app_or_marketplace" }
      : { command, state: command === "repo deploy-key delete" ? "failed" : "proven", cohort: "personal_repo" };
    assert.deepEqual({ command: result?.command, state: result?.state, cohort: manifestRow?.cohort }, expected);
  }
  assert.match(results.get("apps get-authenticated")?.reason || "", /401|App JWT|installation credential/i);
  assert.match(results.get("apps list-subscriptions-for-authenticated-user")?.assertion || "", /200|PM-only/i);
  assert.deepEqual(report.bootstrap_probes?.organization, {
    affected_case_count: 291,
    result: "pm_surface_missing_organization_create",
    delete_not_invoked: "no run-owned immutable organization target exists in the boundary",
  });
  assert.deepEqual(report.bootstrap_probes?.github_app_manifest, {
    affected_case_count: 33,
    result: "pm_surface_missing_manifest_code_issuer",
    conversion_not_invoked: "no PM-issued manifest conversion code exists",
  });
  assert.deepEqual(report.bootstrap_probes?.live_results, [
    { command: "apps get-authenticated", outcome: "credential_or_entitlement_rejected", http_status: 401 },
    { command: "apps list-subscriptions-for-authenticated-user", outcome: "success", http_status: 200 },
  ]);
  assert.equal(JSON.stringify({ report: report.bootstrap_probes, probes }).match(/credential_name|token|response|stdout|stderr/i), null);
  const target = boundaryFile.allowed_targets.find((entry) => entry.key === "personal_repo");
  assert.ok(target);
  assert.equal(target?.run_owned, true);
  assert.deepEqual(
    { owner_slug: target?.owner_slug, owner_id: target?.owner_id, repo_slug: target?.repo_slug, repo_id: target?.repo_id },
    { owner_slug: "karthik-sivadas", owner_id: "6113982", repo_slug: "pm-live-test-direct-read-20260808081515", repo_id: "1327549621" },
  );
  const historical = boundaryFile.historical_runs?.find((entry) => entry.run_id === report.run_id);
  assert.ok(historical);
  assert.notDeepEqual(historical?.target, target);
  assert.throws(() => authorizeLabTarget(boundaryFile, historical?.target), /denied|match|allowlist/i);
  const created = entries.find((entry) => entry.action === "created" && entry.fixture_id === `repository:${historical?.target.repo_id}`);
  const readBack = entries.find((entry) => entry.action === "read_back" && entry.fixture_id === `repository:${historical?.target.repo_id}`);
  assert.equal(created?.provider_id, historical?.target.repo_id);
  assert.equal(readBack?.provider_id, historical?.target.repo_id);
  const issue = entries.find((entry) => entry.action === "created" && String(entry.fixture_id || "").startsWith("issue:"));
  assert.ok(issue);
  const issueEvents = entries.filter((entry) => entry.fixture_id === issue.fixture_id).map((entry) => entry.action);
  assert.deepEqual(issueEvents, ["created", "read_back", "neutralized", "retained"]);
  const issueRetention = entries.find((entry) => entry.fixture_id === issue.fixture_id && entry.action === "retained");
  assert.match(issueRetention?.retention_reason || "", /unsafe_or_disallowed/i);
  const issueFixtures = [...new Set(entries.filter((entry) => String(entry.fixture_id || "").startsWith("issue:")).map((entry) => entry.fixture_id))];
  assert.equal(issueFixtures.length, 4);
  const issueEventGroups = issueFixtures.map((fixtureID) => ({
    fixtureID,
    events: entries.filter((entry) => entry.fixture_id === fixtureID).map((entry) => entry.action),
  }));
  const retainedBeforeEdit = issueEventGroups.filter(({ events }) => JSON.stringify(events) === JSON.stringify(["created", "read_back", "neutralized", "retained"]));
  assert.equal(retainedBeforeEdit.length, 3);
  const editableIssue = issueEventGroups.find(({ events }) => JSON.stringify(events) === JSON.stringify(["created", "read_back", "read_back", "read_back", "neutralized", "retained"]));
  assert.ok(editableIssue);
  assert.match(results.get("issue edit")?.assertion || "", /PM-only|title|body/i);
  assert.match(results.get("issue comment")?.assertion || "", /PM-only|comment/i);
  const label = entries.find((entry) => String(entry.fixture_id || "").startsWith("label:"));
  assert.ok(label);
  assert.equal(typeof label?.provider_id, "string");
  const labelEvents = entries.filter((entry) => entry.fixture_id === label.fixture_id).map((entry) => entry.action);
  assert.deepEqual(labelEvents, ["created", "read_back", "read_back", "cleanup_completed"]);
  assert.match(results.get("label create")?.assertion || "", /PM-only|read-back/i);
  assert.match(results.get("label edit")?.assertion || "", /PM-only|read-back/i);
  assert.match(results.get("label delete")?.assertion || "", /PM-only|absent/i);
  const deployKey = entries.find((entry) => String(entry.fixture_id || "").startsWith("deploy_key:"));
  assert.ok(deployKey);
  assert.equal(typeof deployKey?.provider_id, "string");
  const deployKeyEvents = entries.filter((entry) => entry.fixture_id === deployKey.fixture_id).map((entry) => entry.action);
  assert.deepEqual(deployKeyEvents, ["created", "read_back", "cleanup_failed"]);
  assert.match(results.get("repo deploy-key add")?.assertion || "", /PM-only|read-only|read-back/i);
  assert.match(results.get("repo deploy-key delete")?.reason || "", /false success|independent PM|same immutable/i);
  const deployKeyCleanupFailure = entries.find((entry) => entry.fixture_id === deployKey.fixture_id && entry.action === "cleanup_failed");
  assert.match(deployKeyCleanupFailure?.residual_state || "", /same generated immutable|pending local safety fix/i);
  assert.doesNotThrow(() => validateCleanupLedger({ entries, boundary: boundaryFile }));
});

test("PM process runner keeps supplied approval material out of argv", async () => {
  const temp = await mkdtemp(path.join(os.tmpdir(), "github-live-lab-stdin-"));
  try {
    const binary = path.join(temp, "fake-pm");
    const argvPath = path.join(temp, "argv");
    const stdinPath = path.join(temp, "stdin");
    await writeFile(
      binary,
      `#!/bin/sh\nprintf '%s\\n' "$@" > ${JSON.stringify(argvPath)}\ncat > ${JSON.stringify(stdinPath)}\nprintf '{"kind":"ok"}\\n'\n`,
      "utf8",
    );
    await chmod(binary, 0o755);

    const result = await runPMProcess(
      binary,
      ["reverse", "run", "rplan_test", "--approval-token-stdin"],
      "transient-grant\n",
    );
    assert.equal(result.code, 0);
    const argv = await readFile(argvPath, "utf8");
    assert.equal(argv.includes("--approval-token-stdin"), true);
    assert.equal(argv.includes("--approve"), false);
    assert.equal(argv.includes("transient-grant"), false);
    assert.equal(await readFile(stdinPath, "utf8"), "transient-grant\n");
  } finally {
    await rm(temp, { recursive: true, force: true });
  }
});

test("bootstrap write lifecycle keeps approval material process-only across plan, preview, and execute", async () => {
  const bootstrapBoundary = structuredClone(boundary);
  bootstrapBoundary.bootstrap_principals = [
    {
      key: "personal_private_repo_bootstrap",
      resource_type: "authenticated_user",
      user_slug: "lab-owner",
      user_id: "U_lab_owner",
      allowed_command: "repo create",
      requested_repo_slug: "pm-live-lab-001",
      required_private: true,
      required_auto_init: true,
      purpose: "create exactly one run-owned private lab repository",
    },
  ];
  const request = {
    command: "repo create",
    principal: { user_slug: "lab-owner", user_id: "U_lab_owner" },
    record: { name: "pm-live-lab-001", private: true, auto_init: true },
  };
  const calls = [];
  const result = await runPMPlannedWrite({
    binary: "pm",
    root: "/tmp/github-live-lab-root",
    credentialName: "github-live-proof",
    command: "repo create",
    recordArgs: ["--name", "pm-live-lab-001", "--private", "--auto-init"],
    boundary: bootstrapBoundary,
    bootstrapRequest: request,
    run: async (args, stdin) => {
      calls.push({ args, stdin });
      if (args.includes("--preview")) return { code: 0, stdout: "Approval token: transient-grant\n", stderr: "" };
      if (args.includes("--approval-token-stdin")) {
        return {
          code: 0,
          stdout: JSON.stringify({
            kind: "ReverseRun",
            run: { status: "completed", records_succeeded: 1, records_failed: 0, operation_direct_write: { status: 201 } },
          }),
          stderr: "",
        };
      }
      return { code: 0, stdout: "Created connector command plan plan-test-id\n", stderr: "" };
    },
  });

  assert.equal(calls.length, 3);
  assert.equal(calls[0].args.includes("--credential"), true);
  assert.equal(calls[1].args.includes("--preview"), true);
  assert.equal(calls[2].args.includes("--approval-token-stdin"), true);
  assert.equal(calls[2].args.includes("--approve"), false);
  assert.equal(calls[2].args.includes("transient-grant"), false);
  assert.equal(calls[2].stdin, "transient-grant\n");
  assert.deepEqual(result, { command: "repo create", http_status: 201, records_succeeded: 1, records_failed: 0 });
  assert.equal(JSON.stringify(result).includes("transient-grant"), false);
});

test("planned write re-supplies the exact record flags for withheld-field preview and execution", async () => {
  const recordArgs = [
    "--config", "owner=lab-owner",
    "--config", "repo=pm-live-lab-001",
    "--issue-number", "42",
    "--title", "lab title after edit",
    "--body", "lab body after edit",
  ];
  const calls = [];

  const result = await runPMPlannedWrite({
    binary: "pm",
    root: "/tmp/github-live-lab-root",
    credentialName: "github-live-proof",
    command: "issue edit",
    recordArgs,
    boundary,
    target: boundary.allowed_targets[0],
    run: async (args, stdin) => {
      calls.push({ args, stdin });
      if (args.includes("--preview")) return { code: 0, stdout: "Approval token: transient-grant\n", stderr: "" };
      if (args.includes("--approval-token-stdin")) {
        return {
          code: 0,
          stdout: JSON.stringify({
            kind: "ReverseRun",
            run: { status: "completed", records_succeeded: 1, records_failed: 0, operation_direct_write: { status: 200 } },
          }),
          stderr: "",
        };
      }
      return { code: 0, stdout: "Created connector command plan plan-test-id\n", stderr: "" };
    },
  });

  assert.equal(calls.length, 3);
  assert.deepEqual(calls[0].args.slice(3, 3 + recordArgs.length), recordArgs);
  assert.deepEqual(calls[1].args.slice(6, 6 + recordArgs.length), recordArgs);
  assert.deepEqual(calls[2].args.slice(6, 6 + recordArgs.length), recordArgs);
  assert.equal(calls[1].args.includes("--credential"), false);
  assert.equal(calls[2].args.includes("--credential"), false);
  assert.equal(calls[2].args.includes("--approval-token-stdin"), true);
  assert.equal(calls[2].args.includes("transient-grant"), false);
  assert.equal(calls[2].stdin, "transient-grant\n");
  assert.deepEqual(result, { command: "issue edit", http_status: 200, records_succeeded: 1, records_failed: 0 });
  assert.equal(JSON.stringify(result).includes("transient-grant"), false);
});

test("planned writes reject caller-provided approval stdin markers", async () => {
  let started = false;
  await assert.rejects(
    runPMPlannedWrite({
      binary: "pm",
      root: "/tmp/github-live-lab-root",
      credentialName: "github-live-proof",
      command: "issue create",
      recordArgs: ["--approval-token-stdin"],
      boundary,
      target: boundary.allowed_targets[0],
      run: async () => {
        started = true;
        throw new Error("PM process must not start");
      },
    }),
    /may not override lifecycle or credential flags/u,
  );
  assert.equal(started, false);
});

test("normal planned writes bind the exact repository target before any PM process starts", async () => {
  const base = {
    binary: "pm",
    root: "/tmp/github-live-lab-root",
    credentialName: "lab-credential",
    command: "issue create",
    boundary,
    recordArgs: [
      "--config", "owner=lab-owner",
      "--config", "repo=pm-live-lab-001",
      "--title", "lab fixture title",
    ],
  };
  const protectedTarget = {
    resource_type: "repository",
    owner_slug: "polymetrics-ai",
    owner_id: "O_production",
    repo_slug: "cli",
    repo_id: "R_production",
  };

  let started = false;
  await assert.rejects(
    runPMPlannedWrite({
      ...base,
      target: protectedTarget,
      run: async () => {
        started = true;
        throw new Error("PM process should not start for a protected target");
      },
    }),
    /lab target denied|protected/i,
  );
  assert.equal(started, false);

  started = false;
  await assert.rejects(
    runPMPlannedWrite({
      ...base,
      target: boundary.allowed_targets[0],
      recordArgs: [
        "--config", "owner=lab-owner",
        "--config", "repo=outside-boundary",
        "--title", "lab fixture title",
      ],
      run: async () => {
        started = true;
        throw new Error("PM process should not start for a mismatched repository scope");
      },
    }),
    /repository scope|target/i,
  );
  assert.equal(started, false);
});

test("normal PM reads use the same immutable target scope before the process starts", async () => {
  let started = false;
  await assert.rejects(
    runPMScopedRead({
      binary: "pm",
      root: "/tmp/github-live-lab-root",
      credentialName: "lab-credential",
      command: "issue list",
      commandArgs: ["--config", "owner=polymetrics-ai", "--config", "repo=cli"],
      boundary,
      target: {
        resource_type: "repository",
        owner_slug: "polymetrics-ai",
        owner_id: "O_production",
        repo_slug: "cli",
        repo_id: "R_production",
      },
      run: async () => {
        started = true;
        throw new Error("PM process should not start for protected read target");
      },
    }),
    /lab target denied|protected/i,
  );
  assert.equal(started, false);

  const result = await runPMScopedRead({
    binary: "pm",
    root: "/tmp/github-live-lab-root",
    credentialName: "lab-credential",
    command: "issue list",
    commandArgs: ["--config", "owner=lab-owner", "--config", "repo=pm-live-lab-001", "--state", "all"],
    boundary,
    target: boundary.allowed_targets[0],
    run: async () => ({
      code: 0,
      stdout: JSON.stringify({ kind: "ConnectorCommandRead", connector: "github", command: "issue list", count: 0, records: [] }),
      stderr: "",
    }),
  });
  assert.deepEqual(result, { kind: "ConnectorCommandRead", connector: "github", command: "issue list", count: 0, records: [] });
});

test("label read-back resolves one generated provider ID and rejects absent, duplicate, or malformed results", () => {
  const envelope = {
    kind: "ConnectorCommandRead",
    connector: "github",
    command: "label list",
    count: 2,
    records: [
      { id: 77, name: "unrelated-label", color: "ffffff" },
      { id: 88, name: "pm-live-lab-label-001", color: "0e8a16" },
    ],
  };
  assert.doesNotThrow(() => assertBoundLabLabelAbsent({ envelope, name: "missing-label" }));
  assert.deepEqual(resolveBoundLabLabel({ envelope, name: "pm-live-lab-label-001" }), {
    id: "88",
    name: "pm-live-lab-label-001",
  });
  assert.deepEqual(
    assertBoundLabLabelProperties({
      envelope,
      name: "pm-live-lab-label-001",
      color: "0e8a16",
    }),
    { id: "88", name: "pm-live-lab-label-001" },
  );
  assert.throws(
    () => assertBoundLabLabelProperties({ envelope, name: "pm-live-lab-label-001", color: "ffffff" }),
    /color|expected/i,
  );
  for (const candidate of [
    { ...envelope, command: "issue list" },
    { ...envelope, records: [] },
    { ...envelope, records: [...envelope.records, { id: 99, name: "pm-live-lab-label-001" }] },
  ]) {
    assert.throws(
      () => resolveBoundLabLabel({ envelope: candidate, name: "pm-live-lab-label-001" }),
      /label list|exactly one|malformed/i,
    );
  }
  assert.throws(
    () => assertBoundLabLabelAbsent({ envelope, name: "pm-live-lab-label-001" }),
    /already exists|exactly/i,
  );
});

test("deploy-key read-back retains only a generated immutable ID/title and refuses non-read-only or ambiguous keys", () => {
  const envelope = {
    kind: "ConnectorCommandRead",
    connector: "github",
    command: "repo deploy-key list",
    count: 2,
    records: [
      { id: 77, title: "unrelated deploy key", read_only: true },
      { id: 88, title: "pm-live-lab-deploy-key-001", read_only: true },
    ],
  };
  assert.doesNotThrow(() => assertBoundLabDeployKeyAbsent({ envelope, title: "missing generated deploy key" }));
  const key = resolveBoundLabDeployKey({ envelope, title: "pm-live-lab-deploy-key-001", readOnly: true });
  assert.deepEqual(key, { id: "88", title: "pm-live-lab-deploy-key-001" });
  assert.equal("key" in key, false);
  assert.throws(
    () => resolveBoundLabDeployKey({ envelope: { ...envelope, command: "label list" }, title: "pm-live-lab-deploy-key-001" }),
    /deploy-key list|exactly one|malformed/i,
  );
  assert.throws(
    () => resolveBoundLabDeployKey({ envelope: { ...envelope, records: [{ id: 88, title: "pm-live-lab-deploy-key-001", read_only: false }] }, title: "pm-live-lab-deploy-key-001", readOnly: true }),
    /read-only|expected/i,
  );
  assert.throws(
    () => resolveBoundLabDeployKey({ envelope: { ...envelope, records: [...envelope.records, { id: 99, title: "pm-live-lab-deploy-key-001", read_only: true }] }, title: "pm-live-lab-deploy-key-001" }),
    /exactly one|duplicate/i,
  );
  assert.throws(
    () => assertBoundLabDeployKeyAbsent({ envelope, title: "pm-live-lab-deploy-key-001" }),
    /already exists|exactly/i,
  );
});

test("PM deploy-key read-back retries only stale successful PM list responses and never returns key material", async () => {
  const stale = {
    kind: "ConnectorCommandRead",
    connector: "github",
    command: "repo deploy-key list",
    count: 0,
    records: [],
  };
  const current = {
    kind: "ConnectorCommandRead",
    connector: "github",
    command: "repo deploy-key list",
    count: 1,
    records: [{ id: 12, title: "pm-live-lab-deploy-key-retry", read_only: true }],
  };
  let reads = 0;
  const pauses = [];
  const deployKey = await waitForBoundLabDeployKey({
    read: async () => {
      const result = reads === 0 ? stale : current;
      reads += 1;
      return result;
    },
    title: "pm-live-lab-deploy-key-retry",
    sleep: async (milliseconds) => { pauses.push(milliseconds); },
  });
  assert.deepEqual(deployKey, { id: "12", title: "pm-live-lab-deploy-key-retry", attempts: 2 });
  assert.equal("key" in deployKey, false);
  assert.deepEqual(pauses, [1000]);

  reads = 0;
  await assert.rejects(
    waitForBoundLabDeployKey({
      read: async () => {
        reads += 1;
        throw new Error("PM scoped read failed with provider status 403");
      },
      title: "pm-live-lab-deploy-key-retry",
      sleep: async () => { throw new Error("provider failure must not be retried"); },
    }),
    /provider status 403/i,
  );
  assert.equal(reads, 1);
});

test("PM deploy-key absence read-back retries stale successful lists and returns only attempt count", async () => {
  const stillPresent = {
    kind: "ConnectorCommandRead",
    connector: "github",
    command: "repo deploy-key list",
    count: 1,
    records: [{ id: 12, title: "pm-live-lab-deploy-key-absent", read_only: true }],
  };
  const absent = { ...stillPresent, count: 0, records: [] };
  let reads = 0;
  const pauses = [];
  const result = await waitForBoundLabDeployKeyAbsent({
    read: async () => {
      const envelope = reads === 0 ? stillPresent : absent;
      reads += 1;
      return envelope;
    },
    title: "pm-live-lab-deploy-key-absent",
    sleep: async (milliseconds) => { pauses.push(milliseconds); },
  });
  assert.deepEqual(result, { attempts: 2 });
  assert.deepEqual(pauses, [1000]);

  reads = 0;
  await assert.rejects(
    waitForBoundLabDeployKeyAbsent({
      read: async () => {
        reads += 1;
        throw new Error("PM scoped read failed with provider status 403");
      },
      title: "pm-live-lab-deploy-key-absent",
      sleep: async () => { throw new Error("provider failure must not be retried"); },
    }),
    /provider status 403/i,
  );
  assert.equal(reads, 1);
});

test("issue read-back resolves one generated issue and asserts returned edit/comment state without retaining the record", () => {
  const envelope = {
    kind: "ConnectorCommandRead",
    connector: "github",
    command: "issue list",
    count: 2,
    records: [
      { node_id: "I_other", number: 1, title: "unrelated", body: "other", state: "closed", comments: 0 },
      { node_id: "I_lab_issue_002", number: 2, title: "pm-live-lab-issue-002", body: "edited body", state: "open", comments: 1 },
    ],
  };
  assert.doesNotThrow(() => assertBoundLabIssueAbsent({ envelope, title: "missing generated issue" }));
  assert.deepEqual(
    resolveBoundLabIssue({
      envelope,
      title: "pm-live-lab-issue-002",
      body: "edited body",
      state: "open",
      expectedComments: 1,
      minComments: 1,
    }),
    { id: "I_lab_issue_002", number: 2 },
  );
  assert.throws(
    () => resolveBoundLabIssue({ envelope, title: "pm-live-lab-issue-002", body: "old body" }),
    /body|expected/i,
  );
  assert.throws(
    () => resolveBoundLabIssue({ envelope, title: "pm-live-lab-issue-002", minComments: 2 }),
    /comment/i,
  );
  assert.throws(
    () => resolveBoundLabIssue({ envelope, title: "pm-live-lab-issue-002", expectedComments: 0 }),
    /comment/i,
  );
  for (const candidate of [
    { ...envelope, command: "label list" },
    { ...envelope, records: [] },
    { ...envelope, records: [...envelope.records, { ...envelope.records[1], node_id: "I_duplicate" }] },
  ]) {
    assert.throws(
      () => resolveBoundLabIssue({ envelope: candidate, title: "pm-live-lab-issue-002" }),
      /issue list|exactly one|malformed/i,
    );
  }
  assert.throws(
    () => assertBoundLabIssueAbsent({ envelope, title: "pm-live-lab-issue-002" }),
    /already exists|exactly/i,
  );
});

test("PM issue read-back retries only stale successful PM list responses and returns no provider record", async () => {
  const stale = {
    kind: "ConnectorCommandRead",
    connector: "github",
    command: "issue list",
    count: 0,
    records: [],
  };
  const current = {
    kind: "ConnectorCommandRead",
    connector: "github",
    command: "issue list",
    count: 1,
    records: [{ node_id: "I_lab_issue_retry", number: 7, title: "pm-live-lab-issue-retry", body: "created body", state: "open", comments: 0 }],
  };
  let reads = 0;
  const pauses = [];
  const issue = await waitForBoundLabIssue({
    read: async () => {
      const result = reads === 0 ? stale : current;
      reads += 1;
      return result;
    },
    title: "pm-live-lab-issue-retry",
    body: "created body",
    state: "open",
    expectedComments: 0,
    sleep: async (milliseconds) => { pauses.push(milliseconds); },
  });
  assert.deepEqual(issue, { id: "I_lab_issue_retry", number: 7, attempts: 2 });
  assert.equal(reads, 2);
  assert.deepEqual(pauses, [1000]);
  assert.equal("records" in issue, false);

  reads = 0;
  await assert.rejects(
    waitForBoundLabIssue({
      read: async () => {
        reads += 1;
        throw new Error("PM scoped read failed with provider status 401");
      },
      title: "pm-live-lab-issue-retry",
      sleep: async () => { throw new Error("provider failure must not be retried"); },
    }),
    /provider status 401/i,
  );
  assert.equal(reads, 1);
});

test("fixture commands are PM-only and cleanup ledger remains append-only, redacted, and idempotent", async () => {
  assert.doesNotThrow(() => assertPMOnly("pm github issue create --title {{fixture_title}}"));
  for (const forbidden of ["gh repo create", "curl https://api.github.com/user", "browser open github.com"]) {
    assert.throws(() => assertPMOnly(forbidden), /PM-only/i);
  }

  const temp = await mkdtemp(path.join(os.tmpdir(), "github-live-lab-ledger-"));
  const ledgerPath = path.join(temp, "cleanup.jsonl");
  try {
    const created = {
      schema_version: 1,
      run_id: boundary.run_id,
      fixture_id: "issue-001",
      action: "created",
      target: boundary.allowed_targets[0],
      pm_command: "pm github issue create --title {{fixture_title}}",
      provider_id: "I_lab_issue_001",
    };
    const absent = {
      ...created,
      action: "cleanup_already_absent",
      pm_command: "pm github issue list --state all --json",
    };
    const failedCreated = {
      ...created,
      fixture_id: "deploy-key-001",
      pm_command: "pm github repo deploy-key add --read-only",
      provider_id: "42",
    };
    const failedCleanup = {
      ...failedCreated,
      action: "cleanup_failed",
      pm_command: "pm github repo deploy-key delete --confirm <typed-destructive-acknowledgement>",
      residual_state: "independent PM list retained the same generated immutable deploy key",
    };
    await appendCleanupEntry(ledgerPath, created);
    await appendCleanupEntry(ledgerPath, absent);
    await appendCleanupEntry(ledgerPath, failedCreated);
    await appendCleanupEntry(ledgerPath, failedCleanup);
    const entries = await readCleanupLedger(ledgerPath);
    assert.equal(entries.length, 4);
    assert.doesNotThrow(() => validateCleanupLedger({ entries, boundary }));

    await assert.rejects(
      appendCleanupEntry(ledgerPath, { ...created, credential: "not-permitted" }),
      /forbidden|secret|credential/i,
    );
    assert.equal((await readCleanupLedger(ledgerPath)).length, 4);
    await assert.rejects(
      appendCleanupEntry(ledgerPath, { ...failedCleanup, fixture_id: "deploy-key-002", residual_state: undefined }),
      /cleanup_failed.*residual_state/i,
    );
    assert.equal((await readCleanupLedger(ledgerPath)).length, 4);
  } finally {
    await rm(temp, { recursive: true, force: true });
  }
});

test("captures a complete safe PM Error envelope without inventing a provider status", () => {
  const invocation = "pm github repos list-for-authenticated-user --credential lab-credential --root /tmp/github-live-lab --json";
  const noStatus = {
    api_version: "polymetrics.ai/v1",
    kind: "Error",
    error: {
      category: "internal",
      code: "internal_error",
      message: "direct read GET /user/repos: status 403 was not exposed structurally",
    },
  };
  assert.deepEqual(
    capturePMErrorEnvelope({ invocation, envelope: noStatus }),
    { invocation, envelope: noStatus, provider_status: null },
  );
  const withStatus = { ...noStatus, status: 403 };
  assert.deepEqual(
    capturePMErrorEnvelope({ invocation, envelope: withStatus }),
    { invocation, envelope: withStatus, provider_status: 403 },
  );
  assert.throws(
    () => capturePMErrorEnvelope({ invocation, envelope: { ...noStatus, token: "ghp_example_do_not_store" } }),
    /secret|credential|forbidden/i,
  );
  assert.throws(
    () => capturePMErrorEnvelope({ invocation, envelope: { ...noStatus, status: 99 } }),
    /status/i,
  );
});

test("lab terminal accounting requires exactly one safe terminal result per executed command", () => {
  assert.doesNotThrow(() => validateLabTerminalRecords([
    { command: "issue create", state: "proven", assertion: "readback-matched" },
    { command: "issue list", state: "failed", reason: "provider rejected the isolated fixture request with a safe diagnostic" },
  ]));
  assert.throws(
    () => validateLabTerminalRecords([
      { command: "issue create", state: "proven", assertion: "readback-matched" },
      { command: "issue create", state: "failed", reason: "duplicate terminal result" },
    ]),
    /duplicate/i,
  );
});

test("external bootstrap probes are fixed PM direct reads and reject writes or repository selectors before dispatch", async () => {
  assert.deepEqual(
    authorizeAccountBootstrapProbe({ command: "apps get-authenticated" }),
    {
      id: "github_app_authentication",
      command: "apps get-authenticated",
      method: "GET",
      path: "/app",
      credential_requirement: "GitHub App JWT or installation credential",
    },
  );
  assert.throws(
    () => authorizeAccountBootstrapProbe({ command: "apps create-from-manifest" }),
    /bootstrap probe|direct read|not allowed/i,
  );
  assert.throws(
    () => assertAccountBootstrapProbeInvocation("pm github apps get-authenticated --credential lab-credential --config owner=lab-owner --root /tmp/lab --json"),
    /forbids|fixed|bootstrap/i,
  );

  let started = false;
  await assert.rejects(
    runPMAccountBootstrapProbe({
      binary: "pm",
      root: "/tmp/github-live-lab-root",
      credentialName: "lab-credential",
      probe: { command: "apps create-from-manifest" },
      run: async () => {
        started = true;
        throw new Error("a write must never reach the PM process");
      },
    }),
    /bootstrap probe|direct read|not allowed/i,
  );
  assert.equal(started, false);

  const rejected = await runPMAccountBootstrapProbe({
    binary: "pm",
    root: "/tmp/github-live-lab-root",
    credentialName: "lab-credential",
    probe: { command: "apps get-authenticated" },
    run: async () => ({ code: 1, stdout: "", stderr: "provider status 401" }),
  });
  assert.deepEqual(rejected, {
    command: "apps get-authenticated",
    outcome: "credential_or_entitlement_rejected",
    http_status: 401,
  });
  assert.equal(JSON.stringify(rejected).includes("provider status"), false);
});

test("source-derived bootstrap probe inventory proves the organization/App surface boundary and affected case counts", async () => {
  const [surface, apiSurface, manifest] = await Promise.all([
    loadJSON("internal/connectors/defs/github/cli_surface.json"),
    loadJSON("internal/connectors/defs/github/api_surface.json"),
    loadJSON(".planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-MANIFEST.json"),
  ]);
  const inventory = buildBootstrapProbeInventory({ surface, apiSurface, manifest });
  assert.deepEqual(inventory.organization, {
    affected_case_count: 291,
    create_command: null,
    delete_command: {
      command: "orgs delete",
      method: "DELETE",
      path: "/orgs/{org}",
      availability: "implemented",
    },
    result: "pm_surface_missing_organization_create",
  });
  assert.deepEqual(inventory.github_app_manifest, {
    affected_case_count: 33,
    conversion_command: {
      command: "apps create-from-manifest",
      method: "POST",
      path: "/app-manifests/{code}/conversions",
      required_flags: ["code"],
    },
    code_issuer_commands: [],
    result: "pm_surface_missing_manifest_code_issuer",
  });
  assert.deepEqual(inventory.account_probes, [
    {
      id: "github_app_authentication",
      command: "apps get-authenticated",
      method: "GET",
      path: "/app",
      credential_requirement: "GitHub App JWT or installation credential",
    },
    {
      id: "marketplace_user_subscriptions",
      command: "apps list-subscriptions-for-authenticated-user",
      method: "GET",
      path: "/user/marketplace_purchases",
      credential_requirement: "GitHub Marketplace user entitlement credential",
    },
  ]);
  assert.doesNotThrow(() => validateBootstrapProbeInventory({ inventory, surface, apiSurface, manifest }));
});

test("records exact PM-surface and GitHub credential divergences without a provider fallback or retained account data", async () => {
  const divergences = await loadJSON(".planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-DIVERGENCES.json");
  assert.equal(divergences.current_run_id, "github-live-lab-20260810-target-rebind");
  const records = new Map(divergences.records.map((record) => [record.id, record]));
  assert.deepEqual(
    records.get("github-authenticated-repository-list-rate-limit-scope-error"),
    {
      id: "github-authenticated-repository-list-rate-limit-scope-error",
      run_id: "github-live-lab-20260810-target-rebind",
      command: "repos list-for-authenticated-user",
      phase: "authenticated repository bootstrap discovery",
      state: "local_preflight_defect_observed",
      invocation: "pm github repos list-for-authenticated-user --credential github-pm-live-test-20260808081515 --root /tmp/fm-cli-github-parity-extract-r1/pm-lab --json",
      executed_binary: "/tmp/fm-cli-github-parity-extract-r1/pm-lab/pm",
      pm_exit: 1,
      provider_request_started: false,
      provider_status: null,
      error_envelope: {
        api_version: "polymetrics.ai/v1",
        error: {
          category: "internal",
          code: "internal_error",
          message: "rate-limit policy \"authenticated-user\" requires non-secret config \"rate_limit_account\" for its declared scope",
        },
        kind: "Error",
      },
      diagnosis: {
        local_stage: "GitHub engine newRuntime rate-limit resolution before requester dispatch",
        affected_scope: "GitHub only: every runtime request selected by the whole-connector authenticated-user policy; not shared all-connectors direct-read machinery",
        source: "internal/connectors/defs/github/rate_limits.json + internal/connectors/engine/rate_limit_runtime.go",
        contract_gap: "GitHub's declared config describes rate_limit_account as optional, but the matching runtime policy fails without it and CLI serializes that local configuration error as internal.",
      },
      resolution: "Use the approved non-secret account coordination subject in the target-scoped lab credential, then fix and test the configuration/error-classification contract before relying on this cohort.",
    },
  );
  assert.deepEqual(
    records.get("github-repo-view-current-target-control"),
    {
      id: "github-repo-view-current-target-control",
      run_id: "github-live-lab-20260810-target-rebind",
      command: "repo view",
      phase: "current target current-head read control",
      state: "proven_read_with_projection_limit",
      invocation: "pm github repo view --credential github-pm-live-test-20260808081515 --root /tmp/fm-cli-github-parity-extract-r1/pm-lab --json",
      executed_binary: "/tmp/fm-cli-github-parity-extract-r1/pm-lab/pm",
      pm_exit: 0,
      provider_request_started: true,
      provider_status: null,
      output_envelope: {
        kind: "ConnectorCommandRead",
        connector: "github",
        command: "repo view",
      },
      assertions: {
        immutable_repository_id: "1327549621",
        exact_owner_repo: "karthik-sivadas/pm-live-test-direct-read-20260808081515",
        private: true,
        archived: "not_exposed_by_repo_view_projection",
      },
      limitation: "repo view is the repository ETL stream, and its committed projected schema has no archived field; the successful PM control cannot establish that property.",
      resolution: "The captain's independently verified unarchived state remains the authority for target admission; the boundary is pinned to the PM-read immutable repository ID and this control supplies no override flags.",
    },
  );
  assert.deepEqual(
    records.get("org-bootstrap-create-command-absent"),
    {
      id: "org-bootstrap-create-command-absent",
      command: "orgs create",
      phase: "sandbox organization bootstrap",
      state: "pm_surface_missing",
      provider_request_started: false,
      outcome: "no registered PM command maps to POST /user/orgs or POST /organizations",
      resolution: "Do not create or delete an organization until a PM-only creation path can bind one run-owned immutable organization target.",
    },
  );
  assert.deepEqual(
    records.get("app-manifest-code-issuer-absent"),
    {
      id: "app-manifest-code-issuer-absent",
      command: "apps create-from-manifest",
      phase: "GitHub App fixture bootstrap",
      state: "pm_surface_missing",
      provider_request_started: false,
      outcome: "the PM command requires --code, and no registered PM command issues a manifest conversion code",
      resolution: "Do not create an App until a PM-only manifest-code issuance path can supply a code and bind a run-owned App identity.",
    },
  );
  assert.deepEqual(
    records.get("app-authentication-user-credential-401"),
    {
      id: "app-authentication-user-credential-401",
      command: "apps get-authenticated",
      phase: "GitHub App credential probe",
      state: "credential_or_entitlement_rejected",
      provider_request_started: true,
      http_status: 401,
      resolution: "A dedicated App JWT or installation credential is required before App-authenticated fixture work can begin.",
    },
  );
  assert.deepEqual(
    records.get("marketplace-user-read-no-fixture-bootstrap"),
    {
      id: "marketplace-user-read-no-fixture-bootstrap",
      command: "apps list-subscriptions-for-authenticated-user",
      phase: "GitHub Marketplace credential probe",
      state: "proven_read",
      provider_request_started: true,
      http_status: 200,
      resolution: "This read proves the user route is reachable but does not create a Marketplace listing, App, plan, or installation fixture.",
    },
  );
  assert.deepEqual(
    records.get("issue-list-read-after-write-visibility-delay"),
    {
      id: "issue-list-read-after-write-visibility-delay",
      command: "issue list",
      phase: "personal repository editable issue lifecycle",
      state: "provider_visibility_delay_observed",
      provider_request_started: true,
      outcome: "the immediate independent PM list after a completed PM issue create did not yet satisfy the exact generated-record assertion; a later PM-only list did",
      resolution: "Retry only a successful PM issue-list assertion for at most six attempts; propagate PM credential, entitlement, scope, and provider errors immediately, and retain no provider payload.",
    },
  );
  assert.equal(/credential_name|token|response|stdout|stderr|https?:\/\//i.test(JSON.stringify(divergences.records.slice(1))), false);
});
