#!/usr/bin/env node

import { createHash } from "node:crypto";
import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { assertPersistedArtifactSafe, stableJSONString } from "./github-live-artifact-guard.mjs";
import { buildCases } from "./github-live-cases.mjs";
import { assertPMOnly } from "./github-live-lab.mjs";

const COHORTS = Object.freeze([
  "run_owned_repository",
  "run_owned_organization",
  "github_app_installation",
  "feature_or_entitlement",
]);
const SCRIPT_DIR = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(SCRIPT_DIR, "..");
const SURFACE_PATH = path.join(ROOT, "internal/connectors/defs/github/cli_surface.json");
const CASES_PATH = path.join(ROOT, ".planning/phases/github-parity-extract-r1/LIVE-PROOF-CASES.json");
const DEFAULT_OUTPUT = path.join(ROOT, ".planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-MANIFEST.json");
const MANIFEST_CLASSIFIER_BOUNDARY = Object.freeze({
  owner: "polymetrics-cert",
  owner_id: "O_manifest_classifier",
  repo: "pm-live-lab-manifest",
  repo_id: "R_manifest_classifier",
  organization_id: "O_manifest_classifier",
});

function isPlainObject(value) {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function stableHash(value) {
  return createHash("sha256").update(JSON.stringify(value)).digest("hex");
}

function firstAPI(command) {
  const api = Array.isArray(command.api_surface) ? command.api_surface[0] : undefined;
  return {
    method: typeof api?.method === "string" ? api.method : "",
    path: typeof api?.path === "string" ? api.path : "",
  };
}

function hasOrgTarget(command, apiPath) {
  if (apiPath.includes("/orgs/") || apiPath.includes("/organizations/")) return true;
  return (command.flags || []).some((flag) => flag?.name === "org");
}

function entitlementFor(apiPath, commandPath) {
  const value = `${apiPath} ${commandPath}`.toLowerCase();
  if (value.includes("codespaces")) return "GitHub Codespaces entitlement";
  if (value.includes("copilot")) return "GitHub Copilot organization or enterprise entitlement";
  if (value.includes("code-scanning") || value.includes("code-quality") || value.includes("secret-scanning") || value.includes("dependabot") || value.includes("vulnerability-alert")) {
    return "GitHub Advanced Security or repository security feature entitlement";
  }
  if (value.includes("enterprise")) return "GitHub Enterprise Cloud entitlement";
  if (value.includes("billing") || value.includes("marketplace")) return "provider billing or Marketplace entitlement";
  if (value.includes("private-vulnerability") || value.includes("security-advisories")) return "GitHub security feature entitlement";
  return "named GitHub provider feature entitlement";
}

/**
 * Classify a current-surface command by the run-owned boundary or provider
 * feature it requires. These are fixture cohorts, not pre-skip reasons: a
 * cohort remains subject to real provider admission and read-back.
 */
export function classifyLabCohort(command) {
  if (!isPlainObject(command)) throw new Error("manifest command must be an object");
  const commandPath = String(command.path || "").trim();
  const api = firstAPI(command);
  const value = `${commandPath} ${api.path}`.toLowerCase();
  if (commandPath.startsWith("apps ") || commandPath === "installation view" || api.path === "/app" || api.path.includes("/app/") || api.path.includes("/installation") || api.path.includes("/marketplace_listing/")) {
    return {
      cohort: "github_app_installation",
      credential: { class: "github_app_or_installation", requirement: "GitHub App JWT or installation credential" },
      plan_feature: api.path.includes("marketplace") ? "GitHub App with draft Marketplace listing" : "GitHub App or installation authentication",
      target_allowlist_entry: "run_owned_github_app_installation",
      target_kind: "github_app_installation",
      external_prerequisite: "Bind the captain-provided App installation to the run and prove its immutable installation identity before dispatch.",
    };
  }
  if (/(?:\/enterprises?\/|enterprise|codespaces|copilot|code-scanning|code-quality|secret-scanning|dependabot|vulnerability-alert|private-vulnerability|security-advisories|\/billing\/|\/settings\/billing\/)/u.test(value)) {
    return {
      cohort: "feature_or_entitlement",
      credential: { class: "github_app_installation", requirement: "captain-provided GitHub App installation credential with the named feature entitlement" },
      plan_feature: entitlementFor(api.path, commandPath),
      target_allowlist_entry: "run_owned_entitlement_target",
      target_kind: "entitlement_scoped_resource",
      external_prerequisite: `Verify ${entitlementFor(api.path, commandPath)} on the run-owned Polymetrics-Cert boundary before dispatch.`,
    };
  }
  if (hasOrgTarget(command, api.path)) {
    return {
      cohort: "run_owned_organization",
      credential: { class: "github_app_installation", requirement: "captain-provided GitHub App installation credential" },
      plan_feature: "run-owned Polymetrics-Cert organization",
      target_allowlist_entry: "run_owned_organization",
      target_kind: "organization",
      external_prerequisite: "Resolve the immutable Polymetrics-Cert organization ID into the run boundary before dispatch.",
    };
  }
  return {
    cohort: "run_owned_repository",
    credential: { class: "github_app_installation", requirement: "captain-provided GitHub App installation credential" },
    plan_feature: "run-owned Polymetrics-Cert repository",
    target_allowlist_entry: "run_owned_repository",
    target_kind: "repository",
    external_prerequisite: null,
  };
}

function isWrite(command) {
  return command.intent === "reverse_etl" || command.intent === "direct_write";
}

function cleanupPlan(command) {
  if (!isWrite(command)) return { command: null, strategy: "not_applicable" };
  return { command: null, strategy: "explicit_retention_required" };
}

function lifecyclePlan(command, classifierCase) {
  if (typeof classifierCase?.untestable_reason === "string") {
    return {
      status: isWrite(command) && classifierCase.untestable_reason.includes("no per-operation fixture/read-back/inverse-cleanup resolver")
        ? "requires_typed_lifecycle"
        : "untestable",
      reason: classifierCase.untestable_reason,
    };
  }
  return {
    status: "ready_for_provider_result",
    reason: null,
  };
}

function assertionPM(command, lifecycle) {
  if (lifecycle.status !== "ready_for_provider_result") {
    return null;
  }
  return `pm github ${command.path} {{assertion_flags}} --credential {{credential_name}} --root {{project_root}} --json`;
}

function residualState(command, cleanup) {
  if (!isWrite(command)) return "No fixture was created; the PM response assertion is the residual-state check.";
  if (cleanup.strategy === "delete") {
    return "Independent PM read-back must show the run-owned fixture absent after typed-confirmed cleanup.";
  }
  if (cleanup.strategy === "neutralize_and_retain") {
    return "Independent PM read-back must show the fixture neutralized; the append-only cleanup ledger must retain it with the reason deletion is unavailable.";
  }
  return "No generic read-back or cleanup assertion is valid: this write remains untestable until a typed fixture lifecycle is declared.";
}

function earliestDivergence(baselineReason, classification) {
  if (classification.cohort === "github_app_installation") {
    return "The prior user-token classifier did not have the captain-provided App installation identity; bind that identity and then require a real PM response/read-back.";
  }
  if (classification.cohort === "feature_or_entitlement") {
    return "The trial and App credentials make provider admission observable; retain the named feature requirement and record its actual provider result rather than pre-skipping it.";
  }
  if (/outside the pinned|no approved cleanup-safe fixture|already exists|retain/i.test(baselineReason)) {
    return "The frozen personal-repository restriction is removed; resolve the immutable run-owned Polymetrics-Cert target and prove fixture lifecycle before dispatch.";
  }
  return "The command remains in the current surface; generate a run-bound input and require its real provider result/read-back rather than relying on the historical classification.";
}

function rowFor({ command, archivedCase, classifierCase, index }) {
  const classification = classifyLabCohort(command);
  const api = firstAPI(command);
  const cleanup = cleanupPlan(command);
  const lifecycle = lifecyclePlan(command, classifierCase);
  const destructive = isWrite(command) && (/\bdelete\b/u.test(command.path) || command.risk === "destructive");
  const baselineReason = typeof archivedCase?.untestable_reason === "string"
    ? archivedCase.untestable_reason
    : "not pre-skipped in the frozen case ledger; current surface membership is the source of truth";
  const testPM = lifecycle.status === "ready_for_provider_result"
    ? `pm github ${command.path} {{command_flags}} --credential {{credential_name}} --root {{project_root}} --json`
    : null;
  const cleanupPM = cleanup.command
    ? `pm github ${cleanup.command} {{cleanup_flags}} --credential {{credential_name}} --root {{project_root}} --json`
    : null;
  return {
    case_id: `github-live-lab-${String(index + 1).padStart(4, "0")}-${stableHash([command.path, api.method, api.path]).slice(0, 12)}`,
    command: command.path,
    intent: command.intent || "",
    api,
    baseline_reason: baselineReason,
    cohort: classification.cohort,
    target: { kind: classification.target_kind, lifecycle: "resolve_slug_and_immutable_id_before_write" },
    target_allowlist_entry: classification.target_allowlist_entry,
    credential: classification.credential,
    plan_feature: classification.plan_feature,
    setup_pm: null,
    test_pm: testPM,
    assert_pm: assertionPM(command, lifecycle),
    cleanup_pm: cleanupPM,
    cleanup_strategy: cleanup.strategy,
    lifecycle_status: lifecycle.status,
    ...(lifecycle.reason ? { lifecycle_reason: lifecycle.reason } : {}),
    destructive_acknowledgement: destructive
      ? "required: use the connector-provided typed destructive confirmation after preview"
      : "not required by the current command contract; reverse-ETL approval still applies to writes",
    residual_state_check: residualState(command, cleanup),
    earliest_divergence: earliestDivergence(baselineReason, classification),
    external_prerequisite: classification.external_prerequisite,
  };
}

/** Build one reproducible row for every currently implemented command. */
export function buildLabManifest({ surface, cases, boundary = MANIFEST_CLASSIFIER_BOUNDARY }) {
  if (!isPlainObject(surface) || !Array.isArray(surface.commands)) {
    throw new Error("GitHub CLI surface must contain commands");
  }
  if (!isPlainObject(cases) || !Array.isArray(cases.cases)) {
    throw new Error("baseline live case ledger must contain cases");
  }
  assertPersistedArtifactSafe(cases, "archived live case ledger");
  const classifierCases = buildCases(surface, boundary);
  const commandPaths = new Set();
  for (const command of surface.commands) {
    const commandPath = String(command?.path || "").trim();
    if (commandPath === "" || commandPaths.has(commandPath)) {
      throw new Error("GitHub CLI surface has missing or duplicate command path");
    }
    commandPaths.add(commandPath);
  }
  const baselineByCommand = new Map();
  for (const item of cases.cases) {
    const commandPath = String(item?.command || "").trim();
    if (commandPath === "" || baselineByCommand.has(commandPath)) {
      throw new Error("baseline case ledger has missing or duplicate command paths");
    }
    baselineByCommand.set(commandPath, item);
  }
  const classifierByCommand = new Map(classifierCases.cases.map((item) => [item.command, item]));
  const implemented = surface.commands
    .filter((command) => command.availability === "implemented")
    .sort((left, right) => String(left.path).localeCompare(String(right.path)));
  const rows = implemented.map((command, index) =>
    rowFor({
      command,
      archivedCase: baselineByCommand.get(command.path),
      classifierCase: classifierByCommand.get(command.path),
      index,
    }),
  );
  const classTally = Object.fromEntries(COHORTS.map((cohort) => [cohort, 0]));
  for (const row of rows) classTally[row.cohort] += 1;
  const manifest = {
    schema_version: 3,
    connector: "github",
    source: {
      archived_case_ledger: ".planning/phases/github-parity-extract-r1/LIVE-PROOF-CASES.json",
      archived_case_ledger_sha256: stableHash(cases),
      cli_surface: "internal/connectors/defs/github/cli_surface.json",
      cli_surface_sha256: stableHash(surface),
      implemented_rows: rows.length,
      canonical_case_digest: classifierCases.case_digest,
      current_classification: classifierCases.classification,
      static_movement: classifierCases.measurement,
    },
    policy: {
      provider_lifecycle: "pm_github_only",
      live_terminal_policy: "provider result is required for executable rows; untyped writes are withheld rather than counted as proof",
      write_lifecycle_policy: "a write requires a typed fixture, resource-specific read-back, and inverse cleanup contract before dispatch",
      classes: COHORTS,
    },
    class_tally: classTally,
    rows,
  };
  assertPersistedArtifactSafe(manifest, "lab manifest");
  return manifest;
}

/** Validate the generated artifact against both preserved input sources. */
export function validateLabManifest({ manifest, surface, cases, boundary = MANIFEST_CLASSIFIER_BOUNDARY }) {
  assertPersistedArtifactSafe(manifest, "lab manifest");
  const expected = buildLabManifest({ surface, cases, boundary });
  if (stableJSONString(manifest) !== stableJSONString(expected)) {
    throw new Error("lab manifest must exactly match the canonical classifier-derived artifact");
  }
  for (const row of manifest.rows) {
    if (row.setup_pm !== null) row.setup_pm.forEach(assertPMOnly);
    if (row.test_pm !== null) assertPMOnly(row.test_pm);
    if (row.assert_pm !== null) assertPMOnly(row.assert_pm);
    if (row.cleanup_pm !== null) assertPMOnly(row.cleanup_pm);
  }
  return { rows: expected.rows.length, class_tally: expected.class_tally };
}

function parseArgs(args) {
  const options = {};
  for (let index = 0; index < args.length; index += 1) {
    const argument = args[index];
    if (argument === "--check") {
      options.check = true;
      continue;
    }
    if (argument === "--out") {
      const value = args[index + 1];
      if (!value || value.startsWith("--")) throw new Error("--out requires a path");
      options.out = value;
      index += 1;
      continue;
    }
    throw new Error(`unexpected argument ${argument}`);
  }
  return options;
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  const output = path.resolve(options.out || DEFAULT_OUTPUT);
  const [surface, cases] = await Promise.all([
    JSON.parse(await readFile(SURFACE_PATH, "utf8")),
    JSON.parse(await readFile(CASES_PATH, "utf8")),
  ]);
  const manifest = buildLabManifest({ surface, cases });
  const result = validateLabManifest({ manifest, surface, cases });
  const content = `${JSON.stringify(manifest, null, 2)}\n`;
  if (options.check) {
    const existing = await readFile(output, "utf8");
    if (existing !== content) throw new Error(`manifest drift: regenerate ${output}`);
  } else {
    await writeFile(output, content, { encoding: "utf8", mode: 0o600 });
  }
  process.stdout.write(`github live lab manifest: rows=${result.rows} run_owned_repository=${result.class_tally.run_owned_repository} run_owned_organization=${result.class_tally.run_owned_organization} github_app_installation=${result.class_tally.github_app_installation} feature_or_entitlement=${result.class_tally.feature_or_entitlement}\n`);
}

if (path.resolve(process.argv[1] || "") === fileURLToPath(import.meta.url)) {
  main().catch((error) => {
    process.stderr.write(`github live lab manifest: ${error instanceof Error ? error.message : "generation failed"}\n`);
    process.exitCode = 1;
  });
}
