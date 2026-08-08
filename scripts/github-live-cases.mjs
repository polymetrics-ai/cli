#!/usr/bin/env node

import { readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

const OWNER = "karthik-sivadas";
const REPO = "pm-live-test-direct-read-20260808081515";
const MAIN_SHA = "6da22c1572d6692b958756fdc1d487502fc915e9";
const SCRIPT_DIR = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(SCRIPT_DIR, "..");
const SURFACE_PATH = path.join(ROOT, "internal/connectors/defs/github/cli_surface.json");
const DEFAULT_OUTPUT = path.join(
  ROOT,
  ".planning/phases/github-parity-extract-r1/LIVE-PROOF-CASES.json",
);

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

function firstApiPath(command) {
  const entry = command.api_surface?.[0];
  return typeof entry?.path === "string" ? entry.path : "";
}

function flagValue(flag, command) {
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
    return OWNER;
  }
  if (name === "repo" || name === "repository") return REPO;
  if (name === "repo-name") return REPO;
  if (name === "org") return OWNER;
  if (name === "q") {
    if (command?.path === "search users") return "karthik";
    if (command?.path === "search topics") return "polymetrics";
    if (command?.path === "search commits") return "provider-live";
    return `repo:${OWNER}/${REPO}`;
  }
  if (name === "path") return "README.md";
  if (name === "ref" && command?.path === "git ref view") return "heads/main";
  if (name === "ref" || name === "branch" || name === "base-ref" || name === "target-commitish") return "main";
  if (name === "sha" || name === "commit-sha" || name === "tree-sha") return MAIN_SHA;
  if (name === "basehead") return "main...main";
  if (name === "text" || name === "context" || name === "body" || name === "title" || name === "comment") return "provider-live";
  if (name === "issue" || name === "issue-number") return "5";
  if (name === "head") return "main";
  if (name === "tag-name") return "provider-live";
  if (name === "archive-format") return "tarball";
  return "provider-live";
}

function flagArgs(command, mode) {
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
    if (include) args.push(`--${flag.name}`, flagValue(flag, command));
  }
  return args;
}

function reason(command, detail) {
  return `${command}: ${detail}`;
}

function hasPathPlaceholder(apiPath, name) {
  return apiPath.includes(`{${name}}`);
}

function liveReadBlockReason(command, apiPath) {
  const commandPath = command.path;
  if (commandPath === "activity list-repos-watched-by-user") {
    return reason(commandPath, "the approved account returns an empty provider response for watched repositories, so no returned JSON page can be asserted");
  }
  if (commandPath === "code-security-configuration view") {
    return reason(commandPath, "the retained repository has no configured code-security configuration resource for this read");
  }
  if (commandPath === "discussion view") {
    return reason(commandPath, "the retained repository has no approved discussion resource and the credentialed GraphQL account cannot supply one safely");
  }
  if (commandPath === "project list") {
    return reason(commandPath, "the approved credentialed account has no project-read scope for a safe live GraphQL result");
  }
  if (commandPath === "repo read-file" || commandPath === "repo read-dir" || commandPath === "readme view") {
    return reason(commandPath, "the retained validation repository is empty and has no README or content entry to read back");
  }
  if (commandPath === "git ref view") return "";
  if (commandPath === "license view") {
    return reason(commandPath, "the retained validation repository has no declared license resource to read");
  }
  if (commandPath === "codeowners errors view") {
    return reason(commandPath, "the retained validation repository has no CODEOWNERS file whose syntax errors can be read");
  }
  if (commandPath === "activity list-public-events-for-repo-network") {
    return reason(commandPath, "the pinned repository network has no approved public-event resource for this read");
  }
  if (commandPath === "packages list-docker-migration-conflicting-packages-for-authenticated-user" ||
      commandPath === "packages list-docker-migration-conflicting-packages-for-user") {
    return reason(commandPath, "the approved account has no Docker package migration-conflict resource to enumerate");
  }
  if (commandPath.startsWith("apps ") || commandPath === "installation view") {
    return reason(commandPath, "the endpoint requires GitHub App or installation authentication, while the approved credential is a user token");
  }
  if (apiPath.includes("/marketplace_listing/") || apiPath.includes("/app/") || apiPath === "/app" || apiPath.includes("/installation")) {
    return reason(commandPath, "the endpoint requires GitHub App or marketplace installation state not present in the approved user credential");
  }
  if (apiPath.includes("/branches/") && apiPath.includes("/protection")) {
    return reason(commandPath, "the pinned repository has no approved branch-protection resource and this endpoint requires repository administration scope");
  }
  if (apiPath.includes("/code-scanning/") || apiPath.includes("/code-quality/")) {
    return reason(commandPath, "the pinned repository has no approved code-security analysis resource or required feature scope");
  }
  if (apiPath.includes("/codespaces")) {
    return reason(commandPath, "the pinned repository and approved account have no Codespaces resource authorized for this read");
  }
  if (apiPath.includes("/actions/permissions/fork-pr-contributor-approval") ||
      apiPath.includes("/actions/permissions/selected-actions") ||
      apiPath.includes("/actions/organization-secrets") ||
      apiPath.includes("/actions/organization-variables")) {
    return reason(commandPath, "the pinned repository has no approved organization-level Actions policy resource for this read");
  }
  if (apiPath.includes("/interaction-limits")) {
    return reason(commandPath, "the pinned repository has no configured interaction-limit policy available to this credential");
  }
  if (apiPath.includes("/rulesets")) {
    return reason(commandPath, "the pinned repository has no approved ruleset or ruleset-history resource for this administrative read");
  }
  if (apiPath.includes("/settings/billing/") || apiPath.includes("/billing/")) {
    return reason(commandPath, "the approved personal account has no billing-report resource authorized for this endpoint");
  }
  if (apiPath.includes("/user/migrations") || apiPath.includes("/user/docker/conflicts")) {
    return reason(commandPath, "the approved account has no migration resource prepared for this isolated proof");
  }
  if (apiPath.includes("/repos/{owner}/{repo}/import")) {
    return reason(commandPath, "the pinned repository has no import job resource to read");
  }
  if (apiPath.includes("/issue-types") || apiPath.includes("/issue-field-values") || apiPath.includes("/parent")) {
    return reason(commandPath, "the pinned repository has no configured issue-type or issue-field resource for this read");
  }
  if (apiPath.includes("/pages")) {
    return reason(commandPath, "the pinned repository has no Pages deployment or health resource configured");
  }
  if (apiPath.includes("/private-vulnerability-reporting")) {
    return reason(commandPath, "the pinned repository has no private-vulnerability-reporting resource enabled");
  }
  if (apiPath.includes("/properties/values")) {
    return reason(commandPath, "the pinned repository has no custom-property values resource configured");
  }
  if (apiPath.includes("/secret-scanning/")) {
    return reason(commandPath, "the pinned repository has no approved secret-scanning custom-pattern or history resource");
  }
  if (apiPath.endsWith("/subscription")) {
    return reason(commandPath, "the approved credential has no repository subscription resource to read back");
  }
  if (apiPath.includes("/vulnerability-alerts")) {
    return reason(commandPath, "the pinned repository has no approved vulnerability-alert resource for this read");
  }
  if (commandPath === "user blocks check" || commandPath === "user following check" || commandPath === "user starred check") {
    return reason(commandPath, "the approved account has no positive relationship resource for this check, so the provider returns no successful data page");
  }
  if (commandPath.startsWith("users list-blocked-by-authenticated-user") ||
      commandPath.startsWith("users list-emails-for-authenticated-user") ||
      commandPath.startsWith("users list-gpg-keys-for-authenticated-user") ||
      commandPath.startsWith("users list-ssh-signing-keys-for-authenticated-user")) {
    return reason(commandPath, "the approved account has no safe account-resource result for this credential-scoped list endpoint");
  }
  if (commandPath.startsWith("projects list-for-user")) {
    return reason(commandPath, "the approved account has no project-v2 read scope or isolated project resource for this live read");
  }
  if (commandPath === "interactions get-restrictions-for-authenticated-user") {
    return reason(commandPath, "the approved account has no interaction-limit policy resource available to this credential");
  }
  return "";
}

function targetPrerequisiteReason(command, apiPath, flags) {
  if (apiPath.includes("/orgs/") || apiPath.includes("/enterprises/")) {
    return reason(command, "the endpoint requires an organization or enterprise resource outside the pinned personal test repository");
  }
  if (apiPath.includes("/pulls/{pull_number}") || apiPath.includes("/pulls/{number}")) {
    return reason(command, "the pinned private test repository has no approved pull request resource for this path");
  }
  const resourcePlaceholders = [...apiPath.matchAll(/\{([^}]+)\}/g)]
    .map((match) => match[1])
    .filter((name) => !["owner", "repo", "path", "ref", "branch", "issue_number", "username"].includes(name));
  if (resourcePlaceholders.length > 0) {
    return reason(command, `the endpoint requires existing provider resource identifier(s) ${resourcePlaceholders.join(", ")} that are not present in the retained test repository`);
  }
  const requiredResourceFlags = (flags || [])
    .filter((flag) => flag.required === true)
    .map((flag) => flag.name)
    .filter((name) => ![
      "owner", "repo", "repo-owner", "repo-name", "repository", "username", "user", "target-user",
      "org", "q", "path", "ref", "branch", "base-ref", "target-commitish", "sha", "commit-sha",
      "tree-sha", "basehead", "text", "context", "issue-number",
    ].includes(name));
  if (requiredResourceFlags.length > 0) {
    return reason(command, `the live target has no approved existing resource for required identifier flag(s) ${requiredResourceFlags.join(", ")}`);
  }
  return "";
}

function directReadCase(command) {
  const apiPath = firstApiPath(command);
  const args = flagArgs(command, command.intent);
  const blocked = liveReadBlockReason(command, apiPath);
  if (blocked) return { command: command.path, untestable_reason: blocked };
  const prerequisite = targetPrerequisiteReason(command.path, apiPath, command.flags);
  if (prerequisite) return { command: command.path, untestable_reason: prerequisite };
  return { command: command.path, args };
}

function etlCase(command) {
  const apiPath = firstApiPath(command);
  const blocked = liveReadBlockReason(command, apiPath);
  if (blocked) return { command: command.path, untestable_reason: blocked };
  if (command.path === "project item-list") {
    return {
      command: command.path,
      untestable_reason: reason(command.path, "the approved credential has no isolated project identifier whose items can be read without expanding repository scope"),
    };
  }
  if (apiPath && (apiPath.includes("/orgs/") || apiPath.includes("/enterprises/"))) {
    return { command: command.path, untestable_reason: targetPrerequisiteReason(command.path, apiPath, command.flags) };
  }
  return { command: command.path, args: flagArgs(command, command.intent) };
}

function binaryCase(command) {
  const apiPath = firstApiPath(command);
  return {
    command: command.path,
    untestable_reason: reason(
      command.path,
      apiPath.includes("/orgs/") || apiPath.includes("/user/")
        ? "the approved target has no prepared migration archive resource and creating one would expand mutation scope"
        : "the retained repository has no approved existing artifact, run, job, or SBOM binary resource for this download",
    ),
  };
}

function writeCase(command) {
  const reversible = new Map([
    ["repo archive", { command: "repo view", args: [] }],
    ["repo unarchive", { command: "repo view", args: [] }],
    ["issue close", { command: "issue list", args: ["--state", "all"] }],
    ["issue reopen", { command: "issue list", args: ["--state", "all"] }],
    ["issue lock", { command: "issue list", args: ["--state", "all"] }],
    ["issue unlock", { command: "issue list", args: ["--state", "all"] }],
  ]);
  const readback = reversible.get(command.path);
  if (readback) {
    return {
      command: command.path,
      args: flagArgs(command, command.intent),
      readback,
    };
  }
  if (command.path === "repo create") {
    return {
      command: command.path,
      untestable_reason: reason(command.path, "the approved repository already exists and the captain requires all mutations to remain pinned to and retain that repository"),
    };
  }
  if (command.path === "repo delete" || command.path === "issue delete") {
    return {
      command: command.path,
      untestable_reason: reason(command.path, "the approved live-test contract retains the pinned repository and does not authorize destructive deletion of its resources"),
    };
  }
  if ((command.flags || []).some((flag) => /secret|password|token|private-key|value/i.test(flag.name))) {
    return {
      command: command.path,
      untestable_reason: reason(command.path, "the action requires caller-supplied secret material that the credential-safety contract forbids this proof from creating or serializing"),
    };
  }
  const apiPath = firstApiPath(command);
  if (apiPath.includes("/orgs/") || apiPath.includes("/enterprises/") || !apiPath.includes("/repos/{owner}/{repo}")) {
    return {
      command: command.path,
      untestable_reason: reason(command.path, "the mutation targets provider state outside the pinned repository and no cleanup-safe resource is approved for this live proof"),
    };
  }
  return {
    command: command.path,
    untestable_reason: reason(command.path, "the mutation has no approved cleanup-safe fixture in the retained repository; running it would leave provider state behind"),
  };
}

function buildCases(surface) {
  const commands = surface.commands.filter((command) => command.availability === "implemented");
  const cases = commands.map((command) => {
    if (command.intent === "direct_read") return directReadCase(command);
    if (command.intent === "etl") return etlCase(command);
    if (command.intent === "binary_download") return binaryCase(command);
    if (command.intent === "reverse_etl" || command.intent === "direct_write") return writeCase(command);
    return {
      command: command.path,
      untestable_reason: reason(command.path, "the command intent has no approved live executor classification in the frozen GitHub inventory"),
    };
  });
  return {
    schema_version: 1,
    connector: "github",
    test_repository: { owner: OWNER, repo: REPO },
    context: {
      test_owner: OWNER,
      test_repo: REPO,
      test_repository: `${OWNER}/${REPO}`,
    },
    source: "internal/connectors/defs/github/cli_surface.json",
    cases,
  };
}

async function main() {
  const options = parseArgs(process.argv.slice(2));
  const output = path.resolve(options.out || DEFAULT_OUTPUT);
  const surface = JSON.parse(await readFile(SURFACE_PATH, "utf8"));
  const result = buildCases(surface);
  await writeFile(output, `${JSON.stringify(result, null, 2)}\n`, { mode: 0o600 });
  const tally = result.cases.reduce((counts, item) => {
    const key = item.untestable_reason ? "untestable" : "executable";
    counts[key] += 1;
    return counts;
  }, { executable: 0, untestable: 0 });
  process.stdout.write(`github live cases: executable=${tally.executable} untestable=${tally.untestable} output=${output}\n`);
}

if (path.resolve(process.argv[1] || "") === fileURLToPath(import.meta.url)) {
  main().catch((error) => {
    process.stderr.write(`github live cases: ${error instanceof Error ? error.message : "generation failed"}\n`);
    process.exitCode = 1;
  });
}

export { buildCases };
