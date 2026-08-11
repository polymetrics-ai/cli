#!/usr/bin/env node

import { createHash } from "node:crypto";
import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { assertPersistedArtifactSafe, stableJSONString } from "./github-live-artifact-guard.mjs";
import { validateLabBoundary } from "./github-live-lab.mjs";

const CERTIFICATION_OWNER = "polymetrics-cert";
const SCRIPT_DIR = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(SCRIPT_DIR, "..");
const SURFACE_PATH = path.join(ROOT, "internal/connectors/defs/github/cli_surface.json");
export const CASE_REASON_FAMILIES = Object.freeze([
  "mutation_outside_pinned_repo",
  "org_or_enterprise",
  "secret_material",
  "app_auth",
  "binary_resource",
  "no_cleanup_safe_fixture",
  "other",
]);
export const HISTORICAL_TERMINAL_MEASUREMENT = Object.freeze({
  total: 1521,
  proven: 0,
  failed: 665,
  untestable: 856,
  terminal_timeout_ms: 45000,
});
const APP_INSTALLATION_REPOSITORY_PREFLIGHT = "apps list-repos-accessible-to-installation";

function parseArgs(args) {
  const options = {};
  for (let index = 0; index < args.length; index += 1) {
    const argument = args[index];
    if (!argument.startsWith("--")) throw new Error(`unexpected argument ${argument}`);
    const equals = argument.indexOf("=");
    const name = argument.slice(2, equals === -1 ? undefined : equals);
    if (equals !== -1) {
      options[name] = argument.slice(equals + 1);
      continue;
    }
    const value = args[index + 1];
    if (!value || value.startsWith("--")) throw new Error(`--${name} requires a value`);
    options[name] = value;
    index += 1;
  }
  return options;
}

/**
 * Derive the one provider boundary a live run is allowed to address.  Slugs
 * alone are never enough: the organization and repository must both be
 * present as immutable, run-owned targets in the default-deny lab boundary.
 */
export function resolveLiveBoundary(candidate) {
  const boundary = validateLabBoundary(candidate);
  const organizations = boundary.allowed_targets.filter(
    (target) => target.resource_type === "organization" && target.org_slug === CERTIFICATION_OWNER,
  );
  if (organizations.length !== 1) {
    throw new Error("GitHub live boundary must contain exactly one run-owned Polymetrics-Cert organization target");
  }
  const organization = organizations[0];
  const repositories = boundary.allowed_targets.filter(
    (target) => target.resource_type === "repository" && target.owner_slug === organization.org_slug,
  );
  if (repositories.length !== 1) {
    throw new Error("GitHub live boundary must contain exactly one run-owned repository under Polymetrics-Cert");
  }
  const repository = repositories[0];
  if (repository.owner_id !== organization.org_id) {
    throw new Error("GitHub live boundary repository owner ID must match the run-owned Polymetrics-Cert organization ID");
  }
  return {
    run_id: boundary.run_id,
    owner: repository.owner_slug,
    owner_id: repository.owner_id,
    repo: repository.repo_slug,
    repo_id: repository.repo_id,
    organization_id: organization.org_id,
  };
}

function exactReadAPI(command) {
  const apis = Array.isArray(command?.api_surface) ? command.api_surface : [];
  if (apis.length !== 1) return null;
  const api = apis[0];
  if (!api || api.method !== "GET" || typeof api.path !== "string") return null;
  return api;
}

function normalizedRESTPath(value) {
  const pathValue = String(value || "").trim().split(/[?#]/u, 1)[0].replace(/\/{2,}/gu, "/").replace(/\/+$/u, "").toLowerCase();
  return pathValue === "" ? "/" : pathValue;
}

function canonicalRESTPath(value) {
  const raw = typeof value === "string" ? value : "";
  if (raw !== raw.trim() || !raw.startsWith("/") || /[?#\\%]/u.test(raw)) return null;
  return normalizedRESTPath(raw);
}

function isOrganizationPATGovernanceRead(command) {
  const api = exactReadAPI(command);
  if (!api || api.method !== "GET") return false;
  const pathValue = normalizedRESTPath(api.path);
  return pathValue === "/orgs/{org}/personal-access-token-requests" ||
    pathValue.startsWith("/orgs/{org}/personal-access-token-requests/") ||
    pathValue === "/orgs/{org}/personal-access-tokens" ||
    pathValue.startsWith("/orgs/{org}/personal-access-tokens/");
}

function boundaryRootTargetValues(apiPath, boundary) {
  const roots = [
    ["/repos/{owner}/{repo}", { owner: boundary.owner, repo: boundary.repo }],
    ["/orgs/{org}", { org: boundary.owner }],
    ["/organizations/{org_id}", { org_id: boundary.organization_id }],
    ["/repositories/{repository_id}", { repository_id: boundary.repo_id }],
  ];
  for (const [prefix, values] of roots) {
    if (apiPath === prefix || apiPath.startsWith(`${prefix}/`)) return values;
  }
  return null;
}

function pathTemplateNames(apiPath) {
  return [...apiPath.matchAll(/\{([a-zA-Z0-9_]+)\}/g)].map((match) => match[1]);
}

function pathTargetForFlag(flag, apiPath) {
  const mapsTo = String(flag?.maps_to || "");
  const mapped = /^path\.([a-zA-Z0-9_]+)$/u.exec(mapsTo)?.[1];
  if (mapped) return mapped;
  if (mapsTo !== "") return null;
  const name = String(flag?.name || "").toLowerCase().replaceAll("-", "_");
  return pathTemplateNames(apiPath).includes(name) ? name : null;
}

function configResolvedPathNames(apiPath) {
  return apiPath === "/repos/{owner}/{repo}" || apiPath.startsWith("/repos/{owner}/{repo}/")
    ? new Set(["owner", "repo"])
    : new Set();
}

function hasBoundaryRootTarget(command, boundary) {
  const api = exactReadAPI(command);
  if (!api) return false;
  const apiPath = canonicalRESTPath(api.path);
  if (!apiPath) return false;
  const targetValues = boundaryRootTargetValues(apiPath, boundary);
  if (!targetValues) return false;
  const resolved = configResolvedPathNames(apiPath);
  for (const flag of command.flags || []) {
    const target = pathTargetForFlag(flag, apiPath);
    if (target) {
      if (!(target in targetValues)) return false;
      resolved.add(target);
    }
  }
  return pathTemplateNames(apiPath).every((name) => name in targetValues && resolved.has(name));
}

function boundaryRootArgs(command, boundary) {
  if (!hasBoundaryRootTarget(command, boundary)) return null;
  const api = exactReadAPI(command);
  const apiPath = canonicalRESTPath(api.path);
  if (!apiPath) return null;
  const targetValues = boundaryRootTargetValues(apiPath, boundary);
  const args = [];
  const seen = new Set();
  for (const flag of command.flags || []) {
    const name = String(flag?.name || "");
    const target = pathTargetForFlag(flag, apiPath);
    if (target) {
      if (!(target in targetValues) || !/^[a-z][a-z0-9-]*$/iu.test(name) || seen.has(name)) return null;
      args.push(`--${name}`, targetValues[target]);
      seen.add(name);
      continue;
    }
    if (flag?.required === true) return null;
  }
  return args;
}

function boundaryRootETLArgs(command, boundary) {
  if (!hasBoundaryRootTarget(command, boundary)) return null;
  const api = exactReadAPI(command);
  const apiPath = canonicalRESTPath(api.path);
  if (!apiPath) return null;
  const targetValues = boundaryRootTargetValues(apiPath, boundary);
  const args = [];
  const seen = new Set();
  for (const flag of command.flags || []) {
    const name = String(flag?.name || "");
    const target = pathTargetForFlag(flag, apiPath);
    if (target) {
      if (!(target in targetValues) || !/^[a-z][a-z0-9-]*$/iu.test(name) || seen.has(name)) return null;
      args.push(`--${name}`, targetValues[target]);
      seen.add(name);
      continue;
    }
    if (name === "state" && flag.type === "enum" && Array.isArray(flag.values) && typeof flag.values[0] === "string") {
      args.push(`--${name}`, flag.values[0]);
      continue;
    }
    if (flag?.required === true) return null;
  }
  return args;
}

function isBoundaryRootETL(command, boundary) {
  return hasBoundaryRootTarget(command, boundary);
}

function untestableReadCase(command) {
  return {
    command: command.path,
    untestable_reason: reason(
      command.path,
      "the command target is not rooted in an immutable Polymetrics-Cert boundary entry or declared typed fixture; targetless live reads are limited to the App installation repository identity preflight",
    ),
  };
}

function untestablePATGovernanceCase(command) {
  return {
    command: command.path,
    untestable_reason: reason(
      command.path,
      "organization PAT governance is outside the declared live fixture scope and remains untestable",
    ),
  };
}

function isAppInstallationRepositoryPreflight(command) {
  const api = exactReadAPI(command);
  return command?.path === APP_INSTALLATION_REPOSITORY_PREFLIGHT &&
    command?.intent === "direct_read" &&
    Array.isArray(command?.flags) && command.flags.length === 0 &&
    api?.path === "/installation/repositories";
}

function readCase(command, boundary) {
  if (isAppInstallationRepositoryPreflight(command)) {
    return { command: command.path, args: [] };
  }
  if (isOrganizationPATGovernanceRead(command)) {
    return untestablePATGovernanceCase(command);
  }
  const args = boundaryRootArgs(command, boundary);
  if (args !== null) return { command: command.path, args };
  return untestableReadCase(command);
}

function reason(command, detail) {
  return `${command}: ${detail}`;
}

function reasonFamily(value) {
  const reasonText = String(value || "").toLowerCase();
  if (/github app|app or installation|installation authentication|marketplace/.test(reasonText)) {
    return "app_auth";
  }
  if (/secret|password|private[-_ ]?key|token|credential material|caller-supplied/.test(reasonText)) {
    return "secret_material";
  }
  if (/archive resource|artifact|binary resource|sbom/.test(reasonText)) {
    return "binary_resource";
  }
  if (/organization or enterprise|\/orgs\/|\/enterprises?\//.test(reasonText)) {
    return "org_or_enterprise";
  }
  if (/outside the pinned|pinned repository|retained repository/.test(reasonText)) {
    return "mutation_outside_pinned_repo";
  }
  if (/cleanup-safe|per-operation fixture\/read-back\/inverse-cleanup|fixture.*cleanup|cleanup.*fixture/.test(reasonText)) {
    return "no_cleanup_safe_fixture";
  }
  return "other";
}

function tallyCases(cases) {
  const families = Object.fromEntries(CASE_REASON_FAMILIES.map((family) => [family, 0]));
  let attemptable = 0;
  let blocked = 0;
  for (const item of cases) {
    if (typeof item?.untestable_reason === "string") {
      blocked += 1;
      families[reasonFamily(item.untestable_reason)] += 1;
    } else {
      attemptable += 1;
    }
  }
  return { attemptable, blocked, families };
}

export function deriveCaseClassification(surface, cases) {
  if (!surface || !Array.isArray(surface.commands) || !Array.isArray(cases)) {
    throw new Error("case classification requires a GitHub surface and case array");
  }
  const intents = new Map(surface.commands.map((command) => [command.path, command.intent]));
  const directReadCases = cases.filter((item) => intents.get(item.command) === "direct_read");
  const all = tallyCases(cases);
  const directRead = tallyCases(directReadCases);
  return {
    total: cases.length,
    ...all,
    direct_read: {
      total: directReadCases.length,
      attemptable: directRead.attemptable,
      blocked: directRead.blocked,
    },
  };
}

export function canonicalCaseDigest(cases) {
  if (!Array.isArray(cases)) throw new Error("canonical case digest requires an ordered case array");
  return createHash("sha256").update(stableJSONString(cases)).digest("hex");
}

export function deriveHistoricalTerminalMovement(classification) {
  if (!classification || !Number.isSafeInteger(classification.total) ||
      !Number.isSafeInteger(classification.attemptable) || !Number.isSafeInteger(classification.blocked)) {
    throw new Error("historical terminal movement requires a complete current classification");
  }
  return {
    historical_terminal_measurement: { ...HISTORICAL_TERMINAL_MEASUREMENT },
    current: {
      total: classification.total,
      attemptable: classification.attemptable,
      blocked: classification.blocked,
    },
    movement: {
      total: classification.total - HISTORICAL_TERMINAL_MEASUREMENT.total,
      attemptable: classification.attemptable - HISTORICAL_TERMINAL_MEASUREMENT.failed,
      blocked: classification.blocked - HISTORICAL_TERMINAL_MEASUREMENT.untestable,
    },
  };
}

function assertProductionClassification(surface, classification) {
  const implemented = surface.commands.filter((command) => command.availability === "implemented");
  if (implemented.length !== HISTORICAL_TERMINAL_MEASUREMENT.total) return;
  const expected = {
    total: 1521,
    attemptable: 182,
    blocked: 1339,
    direct_read: { total: 639, attemptable: 169, blocked: 470 },
  };
  const actual = {
    total: classification.total,
    attemptable: classification.attemptable,
    blocked: classification.blocked,
    direct_read: classification.direct_read,
  };
  if (stableJSONString(actual) !== stableJSONString(expected)) {
    throw new Error("current GitHub live classifier no longer matches the reviewed production classification");
  }
}

/** Compare actual case classification against the frozen ledger by reason family. */
export function summarizeCaseMovement({ baselineCases, currentCases }) {
  if (!Array.isArray(baselineCases) || !Array.isArray(currentCases)) {
    throw new Error("case movement requires baseline and current case arrays");
  }
  const baseline = tallyCases(baselineCases);
  const current = tallyCases(currentCases);
  const movement = Object.fromEntries(
    CASE_REASON_FAMILIES.map((family) => [family, current.families[family] - baseline.families[family]]),
  );
  return { baseline, current, movement };
}

function directReadCase(command, boundary) {
  return readCase(command, boundary);
}

function etlCase(command, boundary) {
  if (isOrganizationPATGovernanceRead(command)) return untestablePATGovernanceCase(command);
  if (!isBoundaryRootETL(command, boundary)) return untestableReadCase(command);
  const args = boundaryRootETLArgs(command, boundary);
  return args === null ? untestableReadCase(command) : { command: command.path, args };
}

function binaryCase(command, boundary) {
  return readCase(command, boundary);
}

function writeCase(command) {
  if ((command.flags || []).some((flag) => /secret|password|token|private-key|value/i.test(flag.name))) {
    return {
      command: command.path,
      untestable_reason: reason(command.path, "the action requires caller-supplied secret material that the credential-safety contract forbids this proof from creating or serializing"),
    };
  }
  return {
    command: command.path,
    untestable_reason: reason(command.path, "the run-owned boundary is available, but no per-operation fixture/read-back/inverse-cleanup resolver has been declared; dispatch would not be cleanup-safe"),
  };
}

export function buildCases(surface, boundary) {
  if (!boundary || boundary.owner !== CERTIFICATION_OWNER || !boundary.repo || !boundary.repo_id) {
    throw new Error("buildCases requires the immutable Polymetrics-Cert live boundary");
  }
  assertPersistedArtifactSafe(boundary, "GitHub live boundary");
  const commands = surface.commands.filter((command) => command.availability === "implemented");
  const cases = commands.map((command) => {
    if (command.intent === "direct_read") return directReadCase(command, boundary);
    if (command.intent === "etl") return etlCase(command, boundary);
    if (command.intent === "binary_download") return binaryCase(command, boundary);
    if (command.intent === "reverse_etl" || command.intent === "direct_write") return writeCase(command);
    return {
      command: command.path,
      untestable_reason: reason(command.path, "the command intent has no approved live executor classification in the frozen GitHub inventory"),
    };
  });
  const classification = deriveCaseClassification(surface, cases);
  assertProductionClassification(surface, classification);
  return {
    schema_version: 3,
    connector: "github",
    test_repository: {
      owner: boundary.owner,
      repo: boundary.repo,
      owner_id: boundary.owner_id,
      repo_id: boundary.repo_id,
      organization_id: boundary.organization_id,
    },
    context: {
      test_owner: boundary.owner,
      test_repo: boundary.repo,
      test_repository: `${boundary.owner}/${boundary.repo}`,
    },
    source: "internal/connectors/defs/github/cli_surface.json",
    cases,
    case_digest: canonicalCaseDigest(cases),
    classification,
    measurement: deriveHistoricalTerminalMovement(classification),
  };
}

export function validateCanonicalCaseArtifact({ caseFile, surface, boundary }) {
  assertPersistedArtifactSafe(caseFile, "live proof cases");
  const canonical = buildCases(surface, boundary);
  if (stableJSONString(caseFile) !== stableJSONString(canonical)) {
    throw new Error("live proof cases must exactly match the ordered canonical classifier artifact");
  }
  return canonical;
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  if (Object.keys(options).some((key) => key !== "boundary" && key !== "out")) {
    throw new Error("usage: github-live-cases.mjs --boundary <path> --out <path>");
  }
  if (!options.boundary || !options.out) {
    throw new Error("usage: github-live-cases.mjs --boundary <path> --out <path>");
  }
  const [surface, boundaryFile] = await Promise.all([
    readFile(SURFACE_PATH, "utf8"),
    readFile(path.resolve(options.boundary), "utf8"),
  ]);
  const output = path.resolve(options.out);
  const boundary = resolveLiveBoundary(JSON.parse(boundaryFile));
  const result = buildCases(JSON.parse(surface), boundary);
  assertPersistedArtifactSafe(result, "live proof cases");
  await writeFile(output, `${JSON.stringify(result, null, 2)}\n`, { mode: 0o600 });
  process.stdout.write(`github live cases: executable=${result.classification.attemptable} untestable=${result.classification.blocked} output=${output}\n`);
}

if (path.resolve(process.argv[1] || "") === fileURLToPath(import.meta.url)) {
  main().catch((error) => {
    process.stderr.write(`github live cases: ${error instanceof Error ? error.message : "generation failed"}\n`);
    process.exitCode = 1;
  });
}
