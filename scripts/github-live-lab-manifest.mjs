#!/usr/bin/env node

import { createHash } from "node:crypto";
import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { assertPMOnly } from "./github-live-lab.mjs";

const EXPECTED_PRE_SKIPPED_CASES = 957;
const COHORTS = Object.freeze([
  "personal_repo",
  "sandbox_org_free",
  "github_app_or_marketplace",
  "unavailable_entitlement",
]);
const CLEANUP_STRATEGIES = new Set([
  "not_applicable",
  "delete",
  "neutralize_and_retain",
  "explicit_retention_required",
]);
const SCRIPT_DIR = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(SCRIPT_DIR, "..");
const SURFACE_PATH = path.join(ROOT, "internal/connectors/defs/github/cli_surface.json");
const CASES_PATH = path.join(ROOT, ".planning/phases/github-parity-extract-r1/LIVE-PROOF-CASES.json");
const DEFAULT_OUTPUT = path.join(ROOT, ".planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-MANIFEST.json");

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
 * Classify by the provider capability that must exist before a safe fixture can
 * be attempted. The sequence is intentional: App authentication and named plan
 * features override otherwise ordinary organization/repository routing.
 */
export function classifyLabCohort(command) {
  if (!isPlainObject(command)) throw new Error("manifest command must be an object");
  const commandPath = String(command.path || "").trim();
  const api = firstAPI(command);
  const value = `${commandPath} ${api.path}`.toLowerCase();
  if (commandPath.startsWith("apps ") || commandPath === "installation view" || api.path === "/app" || api.path.includes("/app/") || api.path.includes("/installation") || api.path.includes("/marketplace_listing/")) {
    return {
      cohort: "github_app_or_marketplace",
      credential: { class: "github_app_or_installation", requirement: "GitHub App JWT or installation credential" },
      plan_feature: api.path.includes("marketplace") ? "GitHub App with draft Marketplace listing" : "GitHub App or installation authentication",
      target_allowlist_entry: "github_app_or_installation",
      target_kind: "github_app_or_installation",
      external_prerequisite: "Register/install the dedicated lab GitHub App through a documented PM surface, or record the exact interactive bootstrap impossibility.",
    };
  }
  if (/(?:\/enterprises?\/|enterprise|codespaces|copilot|code-scanning|code-quality|secret-scanning|dependabot|vulnerability-alert|private-vulnerability|security-advisories|\/billing\/|\/settings\/billing\/)/u.test(value)) {
    return {
      cohort: "unavailable_entitlement",
      credential: { class: "user_or_installation_with_feature", requirement: "credential with the named GitHub feature entitlement" },
      plan_feature: entitlementFor(api.path, commandPath),
      target_allowlist_entry: "entitlement_scoped_lab_target",
      target_kind: "entitlement_scoped_resource",
      external_prerequisite: `Enable or trial ${entitlementFor(api.path, commandPath)} on an isolated lab target with the least required permission.`,
    };
  }
  if (hasOrgTarget(command, api.path)) {
    return {
      cohort: "sandbox_org_free",
      credential: { class: "organization_admin_user", requirement: "GitHub Free sandbox organization owner/admin credential" },
      plan_feature: "GitHub Free sandbox organization",
      target_allowlist_entry: "sandbox_org",
      target_kind: "organization",
      external_prerequisite: "Create or name the dedicated GitHub Free sandbox organization and resolve its immutable organization ID through pm github.",
    };
  }
  return {
    cohort: "personal_repo",
    credential: { class: "personal_user", requirement: "dedicated personal lab credential scoped to a private lab repository" },
    plan_feature: "GitHub Free personal private repository",
    target_allowlist_entry: "personal_repo",
    target_kind: "repository_or_personal_account",
    external_prerequisite: null,
  };
}

function isWrite(command) {
  return command.intent === "reverse_etl" || command.intent === "direct_write";
}

function cleanupPlan(command, commands) {
  if (!isWrite(command)) return { command: null, strategy: "not_applicable" };
  if (command.path === "issue create") {
    const closeIssue = commands.find((item) => item.path === "issue close" && item.availability === "implemented");
    if (!closeIssue) throw new Error("issue create requires implemented issue close neutralization");
    return { command: closeIssue.path, strategy: "neutralize_and_retain" };
  }
  const namespace = String(command.path || "").split(" ")[0];
  const deleteCandidates = commands
    .filter((item) => isWrite(item) && item.availability === "implemented" && /\bdelete\b/u.test(String(item.path || "")))
    .sort((left, right) => String(left.path).localeCompare(String(right.path)));
  const sameNamespace = deleteCandidates.find((item) => String(item.path).startsWith(`${namespace} `));
  if (sameNamespace) return { command: sameNamespace.path, strategy: "delete" };
  return { command: null, strategy: "explicit_retention_required" };
}

function setupPM(cohort) {
  if (cohort === "github_app_or_marketplace") {
    return ["pm github apps get-authenticated --credential {{credential_name}} --root {{project_root}} --json"];
  }
  if (cohort === "sandbox_org_free") {
    return ["pm github repo view --credential {{credential_name}} --root {{project_root}} --json"];
  }
  if (cohort === "unavailable_entitlement") {
    return ["pm github rate-limit get --credential {{credential_name}} --root {{project_root}} --json"];
  }
  return ["pm github repo view --credential {{credential_name}} --root {{project_root}} --json"];
}

function assertionPM(command, classification) {
  if (classification.cohort === "github_app_or_marketplace") {
    return "pm github apps get-authenticated --credential {{credential_name}} --root {{project_root}} --json";
  }
  if (isWrite(command)) {
    return "pm github repo view --credential {{credential_name}} --root {{project_root}} --json";
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
  return "Independent PM read-back must show the run-owned fixture absent, neutralized, or explicitly retained with a reason.";
}

function earliestDivergence(historicalReason, classification) {
  if (classification.cohort === "github_app_or_marketplace") {
    return "Historical classifier stopped before PM dispatch because only a user credential was available; the App/installation boundary and credential class are the first new branch.";
  }
  if (classification.cohort === "unavailable_entitlement") {
    return "Historical classifier stopped before PM dispatch because the target feature/plan was unavailable; entitlement resolution is the first new branch.";
  }
  if (/outside the pinned|no approved cleanup-safe fixture|already exists|retain/i.test(historicalReason)) {
    return "Historical classifier stopped before PM dispatch at the pinned-repository cleanup boundary; immutable lab target resolution and fixture provisioning are the first new branch.";
  }
  return "Historical classifier stopped before PM dispatch because a resource/read-back fixture was absent; the parameterized PM fixture resolver is the first new branch.";
}

function rowFor({ command, caseItem, commands, index }) {
  const classification = classifyLabCohort(command);
  const api = firstAPI(command);
  const cleanup = cleanupPlan(command, commands);
  const destructive = isWrite(command) && (/\bdelete\b/u.test(command.path) || command.risk === "destructive");
  const historicalReason = String(caseItem.untestable_reason);
  const testPM = `pm github ${command.path} {{command_flags}} --credential {{credential_name}} --root {{project_root}} --json`;
  const cleanupPM = cleanup.command
    ? `pm github ${cleanup.command} {{cleanup_flags}} --credential {{credential_name}} --root {{project_root}} --json`
    : null;
  return {
    case_id: `github-live-lab-${String(index + 1).padStart(4, "0")}-${stableHash([command.path, api.method, api.path]).slice(0, 12)}`,
    command: command.path,
    intent: command.intent || "",
    api,
    historical_reason: historicalReason,
    cohort: classification.cohort,
    target: { kind: classification.target_kind, lifecycle: "resolve_slug_and_immutable_id_before_write" },
    target_allowlist_entry: classification.target_allowlist_entry,
    credential: classification.credential,
    plan_feature: classification.plan_feature,
    setup_pm: setupPM(classification.cohort),
    test_pm: testPM,
    assert_pm: assertionPM(command, classification),
    cleanup_pm: cleanupPM,
    cleanup_strategy: cleanup.strategy,
    destructive_acknowledgement: destructive
      ? "required: use the connector-provided typed destructive confirmation after preview"
      : "not required by the current command contract; reverse-ETL approval still applies to writes",
    residual_state_check: residualState(command, cleanup),
    earliest_divergence: earliestDivergence(historicalReason, classification),
    external_prerequisite: classification.external_prerequisite,
  };
}

/** Build one reproducible row for every preserved historical pre-skip. */
export function buildLabManifest({ surface, cases }) {
  if (!isPlainObject(surface) || !Array.isArray(surface.commands)) {
    throw new Error("GitHub CLI surface must contain commands");
  }
  if (!isPlainObject(cases) || !Array.isArray(cases.cases)) {
    throw new Error("preserved live case ledger must contain cases");
  }
  const commandsByPath = new Map();
  for (const command of surface.commands) {
    const commandPath = String(command?.path || "").trim();
    if (commandPath === "" || commandsByPath.has(commandPath)) {
      throw new Error("GitHub CLI surface has missing or duplicate command path");
    }
    commandsByPath.set(commandPath, command);
  }
  const historical = cases.cases
    .filter((item) => typeof item?.untestable_reason === "string")
    .sort((left, right) => String(left.command).localeCompare(String(right.command)));
  if (historical.length !== EXPECTED_PRE_SKIPPED_CASES) {
    throw new Error(`preserved case ledger has ${historical.length} pre-skipped rows, expected ${EXPECTED_PRE_SKIPPED_CASES}`);
  }
  const seen = new Set();
  const rows = historical.map((caseItem, index) => {
    const commandPath = String(caseItem.command || "").trim();
    if (seen.has(commandPath)) throw new Error(`preserved case ledger duplicates ${JSON.stringify(commandPath)}`);
    const command = commandsByPath.get(commandPath);
    if (!command) throw new Error(`preserved case ledger names unknown command ${JSON.stringify(commandPath)}`);
    seen.add(commandPath);
    return rowFor({ command, caseItem, commands: surface.commands, index });
  });
  const classTally = Object.fromEntries(COHORTS.map((cohort) => [cohort, 0]));
  for (const row of rows) classTally[row.cohort] += 1;
  return {
    schema_version: 1,
    connector: "github",
    source: {
      case_ledger: ".planning/phases/github-parity-extract-r1/LIVE-PROOF-CASES.json",
      case_ledger_sha256: stableHash(cases),
      cli_surface: "internal/connectors/defs/github/cli_surface.json",
      cli_surface_sha256: stableHash(surface),
      historical_pre_skipped_rows: historical.length,
    },
    policy: {
      provider_lifecycle: "pm_github_only",
      live_terminal_policy: "provider result required; no pre-skip is proof",
      classes: COHORTS,
    },
    class_tally: classTally,
    rows,
  };
}

/** Validate the generated artifact against both preserved input sources. */
export function validateLabManifest({ manifest, surface, cases }) {
  if (!isPlainObject(manifest) || manifest.schema_version !== 1 || manifest.connector !== "github") {
    throw new Error("lab manifest must be a schema-versioned GitHub artifact");
  }
  if (!Array.isArray(manifest.rows)) throw new Error("lab manifest must contain rows");
  const expected = buildLabManifest({ surface, cases });
  if (manifest.rows.length !== EXPECTED_PRE_SKIPPED_CASES) {
    throw new Error(`lab manifest rows must total ${EXPECTED_PRE_SKIPPED_CASES}`);
  }
  const seenCommands = new Set();
  const seenCaseIDs = new Set();
  const tally = Object.fromEntries(COHORTS.map((cohort) => [cohort, 0]));
  for (const row of manifest.rows) {
    if (!isPlainObject(row)) throw new Error("lab manifest row must be an object");
    const command = String(row.command || "").trim();
    const caseID = String(row.case_id || "").trim();
    if (command === "" || caseID === "") throw new Error("lab manifest row must have command and case_id");
    if (seenCommands.has(command) || seenCaseIDs.has(caseID)) throw new Error("lab manifest duplicates command or case_id");
    seenCommands.add(command);
    seenCaseIDs.add(caseID);
    if (!COHORTS.includes(row.cohort)) throw new Error(`lab manifest row ${JSON.stringify(command)} has unknown cohort`);
    tally[row.cohort] += 1;
    if (!Array.isArray(row.setup_pm) || row.setup_pm.length === 0) {
      throw new Error(`lab manifest row ${JSON.stringify(command)} needs PM setup commands`);
    }
    row.setup_pm.forEach(assertPMOnly);
    assertPMOnly(row.test_pm);
    assertPMOnly(row.assert_pm);
    if (row.cleanup_pm !== null) assertPMOnly(row.cleanup_pm);
    if (!CLEANUP_STRATEGIES.has(row.cleanup_strategy)) {
      throw new Error(`lab manifest row ${JSON.stringify(command)} has an invalid cleanup_strategy`);
    }
    if (["delete", "neutralize_and_retain"].includes(row.cleanup_strategy) && row.cleanup_pm === null) {
      throw new Error(`lab manifest row ${JSON.stringify(command)} requires a PM cleanup command`);
    }
    for (const field of ["historical_reason", "target_allowlist_entry", "plan_feature", "destructive_acknowledgement", "residual_state_check", "earliest_divergence"]) {
      if (typeof row[field] !== "string" || row[field].trim() === "") {
        throw new Error(`lab manifest row ${JSON.stringify(command)} is missing ${field}`);
      }
    }
    if (!isPlainObject(row.credential) || typeof row.credential.class !== "string") {
      throw new Error(`lab manifest row ${JSON.stringify(command)} is missing credential class`);
    }
  }
  if (JSON.stringify(tally) !== JSON.stringify(manifest.class_tally)) {
    throw new Error("lab manifest class tally does not match its rows");
  }
  if (JSON.stringify(tally) !== JSON.stringify(expected.class_tally)) {
    throw new Error("lab manifest classes drift from the source-derived classification");
  }
  const expectedCommands = expected.rows.map((row) => row.command).sort();
  if (JSON.stringify([...seenCommands].sort()) !== JSON.stringify(expectedCommands)) {
    throw new Error("lab manifest command set does not match preserved pre-skipped cases");
  }
  return { rows: manifest.rows.length, class_tally: tally };
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
  const content = `${JSON.stringify(manifest, null, 2)}\n`;
  if (options.check) {
    const existing = await readFile(output, "utf8");
    if (existing !== content) throw new Error(`manifest drift: regenerate ${output}`);
  } else {
    await writeFile(output, content, { encoding: "utf8", mode: 0o600 });
  }
  const result = validateLabManifest({ manifest, surface, cases });
  process.stdout.write(`github live lab manifest: rows=${result.rows} personal_repo=${result.class_tally.personal_repo} sandbox_org_free=${result.class_tally.sandbox_org_free} github_app_or_marketplace=${result.class_tally.github_app_or_marketplace} unavailable_entitlement=${result.class_tally.unavailable_entitlement}\n`);
}

if (path.resolve(process.argv[1] || "") === fileURLToPath(import.meta.url)) {
  main().catch((error) => {
    process.stderr.write(`github live lab manifest: ${error instanceof Error ? error.message : "generation failed"}\n`);
    process.exitCode = 1;
  });
}
