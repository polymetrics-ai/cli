import assert from "node:assert/strict";
import test from "node:test";

import { buildCases, resolveLiveBoundary, summarizeCaseMovement } from "../github-live-cases.mjs";

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
