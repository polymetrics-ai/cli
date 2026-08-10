#!/usr/bin/env node

import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { validateLabBoundary } from "./github-live-lab.mjs";

const CERTIFICATION_OWNER = "polymetrics-cert";
const SCRIPT_DIR = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(SCRIPT_DIR, "..");
const SURFACE_PATH = path.join(ROOT, "internal/connectors/defs/github/cli_surface.json");
const REASON_FAMILIES = Object.freeze([
  "mutation_outside_pinned_repo",
  "org_or_enterprise",
  "secret_material",
  "app_auth",
  "binary_resource",
  "no_cleanup_safe_fixture",
  "other",
]);

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

function flagValue(flag, command, boundary) {
  const name = String(flag.name || "");
  if (Array.isArray(flag.values) && flag.values.length > 0) {
    return String(flag.values[0]);
  }
  if (flag.type === "boolean") return "false";
  if (flag.type === "integer") {
    if (name === "issue-number") return "5";
    return "1";
  }
  if (flag.type === "json") {
    if (name.includes("options")) return '{"provider-live":"provider-live"}';
    if (name.includes("ids")) return '["provider-live"]';
    return '{"provider-live":"provider-live"}';
  }
  if (flag.type === "string_array") return "provider-live";
  if (name === "owner" || name === "repo-owner" || name === "username" || name === "user" || name === "target-user") {
    return boundary.owner;
  }
  if (name === "repo" || name === "repository") return boundary.repo;
  if (name === "repo-name") return boundary.repo;
  if (name === "org") return boundary.owner;
  if (name === "q") {
    if (command?.path === "search users") return boundary.owner;
    if (command?.path === "search topics") return "polymetrics";
    if (command?.path === "search commits") return "provider-live";
    return `repo:${boundary.owner}/${boundary.repo}`;
  }
  if (name === "path") return "README.md";
  if (name === "ref" && command?.path === "git ref view") return "heads/main";
  if (name === "ref" || name === "branch" || name === "base-ref" || name === "target-commitish") return "main";
  if (name === "sha" || name === "commit-sha" || name === "tree-sha") return "HEAD";
  if (name === "basehead") return "main...main";
  if (name === "text" || name === "context" || name === "body" || name === "title" || name === "comment") return "provider-live";
  if (name === "issue" || name === "issue-number") return "5";
  if (name === "head") return "main";
  if (name === "tag-name") return "provider-live";
  if (name === "archive-format") return "tarball";
  return "provider-live";
}

function flagArgs(command, mode, boundary) {
  const args = [];
  for (const flag of command.flags || []) {
    const mapsTo = String(flag.maps_to || "");
    const include =
      mode === "reverse_etl" ||
      mode === "direct_write" ||
      mapsTo.startsWith("path.") ||
      flag.required === true ||
      (mode === "etl" && flag.name === "state") ||
      (mode === "direct_read" && ["ref", "q", "path"].includes(flag.name));
    if (include) args.push(`--${flag.name}`, flagValue(flag, command, boundary));
  }
  return args;
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
  const families = Object.fromEntries(REASON_FAMILIES.map((family) => [family, 0]));
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

/** Compare actual case classification against the frozen ledger by reason family. */
export function summarizeCaseMovement({ baselineCases, currentCases }) {
  if (!Array.isArray(baselineCases) || !Array.isArray(currentCases)) {
    throw new Error("case movement requires baseline and current case arrays");
  }
  const baseline = tallyCases(baselineCases);
  const current = tallyCases(currentCases);
  const movement = Object.fromEntries(
    REASON_FAMILIES.map((family) => [family, current.families[family] - baseline.families[family]]),
  );
  return { baseline, current, movement };
}

function directReadCase(command, boundary) {
  return { command: command.path, args: flagArgs(command, command.intent, boundary) };
}

function etlCase(command, boundary) {
  return { command: command.path, args: flagArgs(command, command.intent, boundary) };
}

function binaryCase(command, boundary) {
  return { command: command.path, args: flagArgs(command, command.intent, boundary) };
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
  return {
    schema_version: 2,
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
  };
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  if (!options.boundary || !options.out) {
    throw new Error("usage: github-live-cases.mjs --boundary <path> --out <path> [--baseline <frozen-cases-path>]");
  }
  const [surface, boundaryFile, baselineFile] = await Promise.all([
    readFile(SURFACE_PATH, "utf8"),
    readFile(path.resolve(options.boundary), "utf8"),
    options.baseline ? readFile(path.resolve(options.baseline), "utf8") : Promise.resolve(null),
  ]);
  const output = path.resolve(options.out);
  const boundary = resolveLiveBoundary(JSON.parse(boundaryFile));
  const result = buildCases(JSON.parse(surface), boundary);
  if (baselineFile !== null) {
    const baseline = JSON.parse(baselineFile);
    result.measurement = summarizeCaseMovement({
      baselineCases: baseline.cases,
      currentCases: result.cases,
    });
  }
  result.classification = tallyCases(result.cases);
  await writeFile(output, `${JSON.stringify(result, null, 2)}\n`, { mode: 0o600 });
  process.stdout.write(`github live cases: executable=${result.classification.attemptable} untestable=${result.classification.blocked} output=${output}\n`);
}

if (path.resolve(process.argv[1] || "") === fileURLToPath(import.meta.url)) {
  main().catch((error) => {
    process.stderr.write(`github live cases: ${error instanceof Error ? error.message : "generation failed"}\n`);
    process.exitCode = 1;
  });
}
