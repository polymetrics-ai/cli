import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

import {
  buildCases,
  canonicalCaseDigest,
  resolveLiveBoundary,
  summarizeCaseMovement,
} from "../github-live-cases.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");

const boundary = {
  schema_version: 1,
  run_id: "github-live-cert-test-20260810",
  default_deny: true,
  protected_owners: ["polymetrics-ai"],
  protected_repositories: [{ owner_slug: "polymetrics-ai", repo_slug: "cli" }],
  working_repositories: [{ owner_slug: "polymetrics-ai", repo_slug: "cli" }],
  allowed_targets: [
    {
      key: "certification_org",
      resource_type: "organization",
      org_slug: "polymetrics-cert",
      org_id: "O_certification",
      run_owned: true,
    },
    {
      key: "certification_repo",
      resource_type: "repository",
      owner_slug: "polymetrics-cert",
      owner_id: "O_certification",
      repo_slug: "pm-live-lab-20260810",
      repo_id: "R_certification",
      run_owned: true,
    },
  ],
};

const surface = {
  commands: [
    {
      path: "repo view",
      availability: "implemented",
      intent: "direct_read",
      flags: [
        { name: "owner", required: true, type: "string" },
        { name: "repo", required: true, type: "string" },
      ],
      api_surface: [{ method: "GET", path: "/repos/{owner}/{repo}" }],
    },
    {
      path: "orgs list-members",
      availability: "implemented",
      intent: "direct_read",
      flags: [{ name: "org", required: true, type: "string" }],
      api_surface: [{ method: "GET", path: "/orgs/{org}/members" }],
    },
  ],
};

test("resolves one immutable, run-owned organization and repository from the lab boundary", () => {
  assert.deepEqual(resolveLiveBoundary(boundary), {
    run_id: "github-live-cert-test-20260810",
    owner: "polymetrics-cert",
    owner_id: "O_certification",
    repo: "pm-live-lab-20260810",
    repo_id: "R_certification",
    organization_id: "O_certification",
  });
});

test("derives live-case flags from the supplied immutable boundary instead of historical constants", () => {
  const cases = buildCases(surface, resolveLiveBoundary(boundary));

  assert.deepEqual(cases.test_repository, {
    owner: "polymetrics-cert",
    repo: "pm-live-lab-20260810",
    owner_id: "O_certification",
    repo_id: "R_certification",
    organization_id: "O_certification",
  });
  assert.deepEqual(cases.context, {
    test_owner: "polymetrics-cert",
    test_repo: "pm-live-lab-20260810",
    test_repository: "polymetrics-cert/pm-live-lab-20260810",
  });
  assert.deepEqual(cases.cases.find((item) => item.command === "repo view"), {
    command: "repo view",
    args: ["--owner", "polymetrics-cert", "--repo", "pm-live-lab-20260810"],
  });
  assert.deepEqual(cases.cases.find((item) => item.command === "orgs list-members"), {
    command: "orgs list-members",
    args: ["--org", "polymetrics-cert"],
  });
});

test("permits only boundary-root reads and the installation repository preflight", () => {
  const cases = buildCases({
    commands: [
      {
        path: "apps list-repos-accessible-to-installation",
        availability: "implemented",
        intent: "direct_read",
        flags: [],
        api_surface: [{ method: "GET", path: "/installation/repositories" }],
      },
      {
        path: "repo actions permissions",
        availability: "implemented",
        intent: "direct_read",
        flags: [],
        api_surface: [{ method: "GET", path: "/repos/{owner}/{repo}/actions/permissions" }],
      },
      {
        path: "orgs list-members",
        availability: "implemented",
        intent: "direct_read",
        flags: [{ name: "org", required: true, type: "string", maps_to: "path.org" }],
        api_surface: [{ method: "GET", path: "/orgs/{org}/members" }],
      },
      {
        path: "workflows get",
        availability: "implemented",
        intent: "direct_read",
        flags: [{ name: "workflow-id", required: true, type: "integer", maps_to: "path.workflow_id" }],
        api_surface: [{ method: "GET", path: "/repos/{owner}/{repo}/actions/workflows/{workflow_id}" }],
      },
      {
        path: "gists get",
        availability: "implemented",
        intent: "direct_read",
        flags: [{ name: "gist-id", required: true, type: "string", maps_to: "path.gist_id" }],
        api_surface: [{ method: "GET", path: "/gists/{gist_id}" }],
      },
      {
        path: "actions enterprise cache",
        availability: "implemented",
        intent: "direct_read",
        flags: [{ name: "enterprise", required: true, type: "string", maps_to: "path.enterprise" }],
        api_surface: [{ method: "GET", path: "/enterprises/{enterprise}/actions/cache/retention-limit" }],
      },
      {
        path: "users authenticated",
        availability: "implemented",
        intent: "direct_read",
        flags: [],
        api_surface: [{ method: "GET", path: "/user/repos" }],
      },
      {
        path: "repo archive tarball",
        availability: "implemented",
        intent: "binary_download",
        flags: [{ name: "ref", type: "string", maps_to: "path.ref" }],
        api_surface: [{ method: "GET", path: "/repos/{owner}/{repo}/tarball/{ref}" }],
      },
      {
        path: "repo snapshot",
        availability: "implemented",
        intent: "etl",
        flags: [],
        api_surface: [{ method: "GET", path: "/repos/{owner}/{repo}" }],
      },
      {
        path: "project item-list",
        availability: "implemented",
        intent: "etl",
        flags: [{ name: "project-id", required: true, type: "string", maps_to: "query.project_id" }],
        api_surface: [{ method: "GRAPHQL", path: "ListProjectItems" }],
      },
      {
        path: "project list",
        availability: "implemented",
        intent: "etl",
        flags: [],
        api_surface: [{ method: "GRAPHQL", path: "ListProjects" }],
      },
      {
        path: "discussion view",
        availability: "implemented",
        intent: "etl",
        flags: [{ name: "number", required: true, type: "integer", maps_to: "query.number" }],
        api_surface: [{ method: "GRAPHQL", path: "ViewDiscussion" }],
      },
    ],
  }, resolveLiveBoundary(boundary));

  assert.deepEqual(cases.cases.filter((item) => item.args !== undefined), [
    { command: "apps list-repos-accessible-to-installation", args: [] },
    { command: "repo actions permissions", args: [] },
    { command: "orgs list-members", args: ["--org", "polymetrics-cert"] },
    { command: "repo snapshot", args: [] },
  ]);
  for (const item of cases.cases.filter((item) => item.untestable_reason !== undefined)) {
    assert.match(item.untestable_reason, /immutable.*boundary|typed fixture|targetless/i);
  }
  assert.equal(JSON.stringify(cases).includes("provider-live"), false);
});

test("keeps normalized GET organization PAT-governance paths untestable", () => {
  const cases = buildCases({
    commands: [
      {
        path: "orgs list-pat-grant-requests",
        availability: "implemented",
        intent: "direct_read",
        flags: [{ name: "org", required: true, type: "string", maps_to: "path.org" }],
        api_surface: [{ method: "GET", path: "/orgs/{org}/personal-access-token-requests/" }],
      },
      {
        path: "orgs list-pat-grants",
        availability: "implemented",
        intent: "direct_read",
        flags: [{ name: "org", required: true, type: "string", maps_to: "path.org" }],
        api_surface: [{ method: "GET", path: "/orgs/{org}/personal-access-tokens" }],
      },
      {
        path: "orgs list-pat-grants-with-query",
        availability: "implemented",
        intent: "direct_read",
        flags: [{ name: "org", required: true, type: "string", maps_to: "path.org" }],
        api_surface: [{ method: "GET", path: "/orgs/{org}/personal-access-tokens?visibility=all" }],
      },
    ],
  }, resolveLiveBoundary(boundary));

  for (const item of cases.cases) {
    assert.equal(item.args, undefined);
    assert.match(item.untestable_reason, /PAT governance|immutable.*boundary|targetless/i);
  }
});

test("derives the production classifier, digest, and static terminal movement", async () => {
  const productionSurface = JSON.parse(await readFile(
    path.join(root, "internal/connectors/defs/github/cli_surface.json"),
    "utf8",
  ));
  const cases = buildCases(productionSurface, resolveLiveBoundary(boundary));

  assert.deepEqual(
    {
      total: cases.classification.total,
      attemptable: cases.classification.attemptable,
      blocked: cases.classification.blocked,
      direct_read: cases.classification.direct_read,
    },
    {
      total: 1521,
      attemptable: 182,
      blocked: 1339,
      direct_read: { total: 639, attemptable: 169, blocked: 470 },
    },
  );
  assert.equal(cases.case_digest, canonicalCaseDigest(cases.cases));
  assert.deepEqual(cases.measurement, {
    historical_terminal_measurement: {
      total: 1521,
      proven: 0,
      failed: 665,
      untestable: 856,
      terminal_timeout_ms: 45000,
    },
    current: { total: 1521, attemptable: 182, blocked: 1339 },
    movement: { total: 0, attemptable: -483, blocked: 483 },
  });
});

test("reports reason-family movement instead of preserving a frozen pre-skip tally", () => {
  assert.deepEqual(
    summarizeCaseMovement({
      baselineCases: [
        { command: "one", untestable_reason: "one: mutation targets provider state outside the pinned repository" },
        { command: "two", untestable_reason: "two: requires GitHub App or installation authentication while only a user token is available" },
        { command: "three", untestable_reason: "three: no approved cleanup-safe fixture exists" },
      ],
      currentCases: [
        { command: "one", args: [] },
        { command: "two", args: [] },
        { command: "three", untestable_reason: "three: no per-operation fixture/read-back/inverse-cleanup resolver has been declared" },
      ],
    }),
    {
      baseline: {
        attemptable: 0,
        blocked: 3,
        families: {
          mutation_outside_pinned_repo: 1,
          org_or_enterprise: 0,
          secret_material: 0,
          app_auth: 1,
          binary_resource: 0,
          no_cleanup_safe_fixture: 1,
          other: 0,
        },
      },
      current: {
        attemptable: 2,
        blocked: 1,
        families: {
          mutation_outside_pinned_repo: 0,
          org_or_enterprise: 0,
          secret_material: 0,
          app_auth: 0,
          binary_resource: 0,
          no_cleanup_safe_fixture: 1,
          other: 0,
        },
      },
      movement: {
        mutation_outside_pinned_repo: -1,
        org_or_enterprise: 0,
        secret_material: 0,
        app_auth: -1,
        binary_resource: 0,
        no_cleanup_safe_fixture: 0,
        other: 0,
      },
    },
  );
});
