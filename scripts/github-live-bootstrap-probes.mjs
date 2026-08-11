#!/usr/bin/env node

import { createHash } from "node:crypto";
import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { assertPersistedArtifactSafe } from "./github-live-artifact-guard.mjs";

const SCRIPT_DIR = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(SCRIPT_DIR, "..");
const SURFACE_PATH = path.join(ROOT, "internal/connectors/defs/github/cli_surface.json");
const API_SURFACE_PATH = path.join(ROOT, "internal/connectors/defs/github/api_surface.json");
const MANIFEST_PATH = path.join(ROOT, ".planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-MANIFEST.json");
const DEFAULT_OUTPUT = path.join(ROOT, ".planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-BOOTSTRAP-PROBES.json");

function isPlainObject(value) {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function stableHash(value) {
  return createHash("sha256").update(JSON.stringify(value)).digest("hex");
}

function requireCommand(surface, commandPath) {
  if (!isPlainObject(surface) || !Array.isArray(surface.commands)) {
    throw new Error("GitHub CLI surface must contain commands");
  }
  const command = surface.commands.find((entry) => entry?.path === commandPath);
  if (!command) throw new Error(`GitHub CLI surface is missing ${JSON.stringify(commandPath)}`);
  return command;
}

function exactAPI(command, method, apiPath) {
  const entry = (command.api_surface || []).find((candidate) => candidate?.method === method && candidate?.path === apiPath);
  if (!entry) throw new Error(`${JSON.stringify(command.path)} is missing ${method} ${apiPath}`);
  return entry;
}

function requiredFlags(command) {
  return (command.flags || [])
    .filter((flag) => flag?.required === true && typeof flag.name === "string")
    .map((flag) => flag.name)
    .sort();
}

function cohortCount(manifest, cohort) {
  if (!isPlainObject(manifest) || !Array.isArray(manifest.rows)) {
    throw new Error("GitHub lab manifest must contain rows");
  }
  return manifest.rows.filter((row) => row?.cohort === cohort).length;
}

function organizationCreateCommands(surface) {
  return surface.commands
    .filter((command) => (command.api_surface || []).some((api) => api?.method === "POST" && (api.path === "/user/orgs" || api.path === "/organizations")))
    .map((command) => command.path)
    .sort();
}

function manifestCodeIssuerCommands(surface) {
  return surface.commands
    .filter((command) => (command.api_surface || []).some((api) => typeof api?.path === "string" && api.path.startsWith("/app-manifests/") && api.path !== "/app-manifests/{code}/conversions"))
    .map((command) => command.path)
    .sort();
}

/**
 * Build a response-body-free map of the exact GitHub PM bootstrap surface.
 * This examines only current-head generated artifacts and the complete
 * implemented-surface manifest;
 * it cannot perform a provider call or resolve a production target.
 */
export function buildBootstrapProbeInventory({ surface, apiSurface, manifest }) {
  assertPersistedArtifactSafe(manifest, "GitHub live lab manifest");
  if (!isPlainObject(apiSurface) || !Array.isArray(apiSurface.endpoints)) {
    throw new Error("GitHub API surface must contain endpoints");
  }
  const orgDelete = requireCommand(surface, "orgs delete");
  exactAPI(orgDelete, "DELETE", "/orgs/{org}");
  if (orgDelete.availability !== "implemented") throw new Error("orgs delete must remain implemented for the bootstrap audit");
  const orgCreate = organizationCreateCommands(surface);
  if (orgCreate.length !== 0) throw new Error("organization bootstrap audit found an unreviewed PM creation command");

  const appConversion = requireCommand(surface, "apps create-from-manifest");
  exactAPI(appConversion, "POST", "/app-manifests/{code}/conversions");
  const appCodeIssuers = manifestCodeIssuerCommands(surface);
  if (appCodeIssuers.length !== 0) throw new Error("GitHub App bootstrap audit found an unreviewed PM manifest-code issuer");

  const documentedOrgCreateEndpoints = apiSurface.endpoints
    .filter((endpoint) => endpoint?.method === "POST" && (endpoint.path === "/user/orgs" || endpoint.path === "/organizations"))
    .map((endpoint) => ({ method: endpoint.method, path: endpoint.path }));
  if (documentedOrgCreateEndpoints.length !== 0) {
    throw new Error("GitHub API surface unexpectedly documents an unreviewed organization-create endpoint");
  }

  const inventory = {
    schema_version: 3,
    connector: "github",
    source: {
      cli_surface: {
        path: "internal/connectors/defs/github/cli_surface.json",
        sha256: stableHash(surface),
      },
      api_surface: {
        path: "internal/connectors/defs/github/api_surface.json",
        sha256: stableHash(apiSurface),
      },
      live_lab_manifest: {
        path: ".planning/phases/github-parity-extract-r1/GITHUB-LIVE-LAB-MANIFEST.json",
        sha256: stableHash(manifest),
      },
    },
    policy: {
      provider_operations: "pm_github_only",
      account_probes: "none; targetless direct reads remain untestable outside the credential-bound installation repository preflight",
      organization_delete: "not_invoked_without_run_owned_immutable_target_and_cleanup_provenance",
    },
    organization: {
      affected_case_count: cohortCount(manifest, "run_owned_organization"),
      create_command: null,
      delete_command: {
        command: orgDelete.path,
        method: "DELETE",
        path: "/orgs/{org}",
        availability: orgDelete.availability,
      },
      result: "pm_surface_missing_organization_create",
    },
    github_app_manifest: {
      affected_case_count: cohortCount(manifest, "github_app_installation"),
      conversion_command: {
        command: appConversion.path,
        method: "POST",
        path: "/app-manifests/{code}/conversions",
        required_flags: requiredFlags(appConversion),
      },
      code_issuer_commands: appCodeIssuers,
      result: "pm_surface_missing_manifest_code_issuer",
    },
    account_probes: [],
  };
  assertPersistedArtifactSafe(inventory, "GitHub bootstrap probe inventory");
  return inventory;
}

/** Fail closed if a checked-in probe inventory drifts from its source artifacts. */
export function validateBootstrapProbeInventory({ inventory, surface, apiSurface, manifest }) {
  assertPersistedArtifactSafe(inventory, "GitHub bootstrap probe inventory");
  if (!isPlainObject(inventory) || inventory.schema_version !== 3 || inventory.connector !== "github") {
    throw new Error("bootstrap probe inventory must be a schema-versioned GitHub artifact");
  }
  const expected = buildBootstrapProbeInventory({ surface, apiSurface, manifest });
  if (JSON.stringify(inventory) !== JSON.stringify(expected)) {
    throw new Error("bootstrap probe inventory drifted from current GitHub source artifacts");
  }
  return {
    organization_cases: expected.organization.affected_case_count,
    github_app_or_marketplace_cases: expected.github_app_manifest.affected_case_count,
  };
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
  const [surface, apiSurface, manifest] = await Promise.all([
    JSON.parse(await readFile(SURFACE_PATH, "utf8")),
    JSON.parse(await readFile(API_SURFACE_PATH, "utf8")),
    JSON.parse(await readFile(MANIFEST_PATH, "utf8")),
  ]);
  const inventory = buildBootstrapProbeInventory({ surface, apiSurface, manifest });
  validateBootstrapProbeInventory({ inventory, surface, apiSurface, manifest });
  const content = `${JSON.stringify(inventory, null, 2)}\n`;
  if (options.check) {
    const existing = await readFile(output, "utf8");
    if (existing !== content) throw new Error(`bootstrap probe inventory drift: regenerate ${output}`);
  } else {
    await writeFile(output, content, { encoding: "utf8", mode: 0o600 });
  }
  console.log(
    `github live bootstrap probes: organization_cases=${inventory.organization.affected_case_count} app_marketplace_cases=${inventory.github_app_manifest.affected_case_count}`,
  );
}

if (process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main().catch((error) => {
    console.error(`github live bootstrap probes: ${error.message}`);
    process.exitCode = 1;
  });
}
