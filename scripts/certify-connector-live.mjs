#!/usr/bin/env node

import { createHash, createHmac, randomBytes } from "node:crypto";
import { chmod, mkdir, mkdtemp, open, readFile, rename, rm, unlink, writeFile } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";
import { spawn } from "node:child_process";

const scriptPath = fileURLToPath(import.meta.url);
const repositoryRoot = path.resolve(path.dirname(scriptPath), "..");
const definitionsRoot = path.join(repositoryRoot, "internal", "connectors", "defs");
const evidenceRoot = path.join(repositoryRoot, "internal", "connectors", "certifications", "evidence");
const saltPath = path.join(repositoryRoot, "internal", "connectors", "certifications", ".fingerprint-salt");
const temporaryRoot = path.join(repositoryRoot, ".tmp", "live-certification");
const receiptRoot = path.join(repositoryRoot, ".planning", "live-certification-runs");
const markerPrefix = "{{pmcertfp:v1:";
const markerSuffix = "}}";
const credentialNote = "Only the credential use documented by this record's protocol exchanges was verified; no broader credential scope is claimed.";
const maxOutputBytes = 2 * 1024 * 1024;
const safeIdentifier = /^[A-Za-z0-9][A-Za-z0-9_.:-]{0,255}$/u;
const safeFieldName = /^[A-Za-z0-9][A-Za-z0-9_.-]*$/u;
const safeEnvironmentName = /^[A-Za-z_][A-Za-z0-9_]*$/u;
const httpMethods = new Set(["GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"]);
const nonPassOutcomes = new Set(["no_object", "wrong_credential", "entitlement", "product_defect"]);

function usage() {
  return `usage: certify-connector-live.mjs <connector> [options]

Run definition-declared direct-read candidates one at a time. Each successful
candidate is persisted immediately as a bounded observed-operations record and
checked before another candidate is attempted.

options:
  --pm <path>                       built pm binary (required for a live run)
  --credential-env <ENV>            environment variable holding the credential
  --credential-field <field>        credential secret field for pm credentials add
  --credential-config <key=value>   non-secret credential config; repeatable
  --root <path>                     run-owned pm project inside this repository
  --receipt-file <path>             repository-relative sanitized receipt path
  --stages-file <path>              repository-relative JSON selector for declared stage names
  --limit <n>                       stop after n candidates (default: all)
  --definition-check                validate definitions without running pm
  --help                            show this help
`;
}

function fail(message) {
  throw new Error(message);
}

function requireIdentifier(value, label) {
  const normalized = String(value || "").trim();
  if (!safeIdentifier.test(normalized)) fail(`${label} must be a safe identifier`);
  return normalized;
}

function requirePathText(value, label) {
  const normalized = String(value || "").trim();
  if (!normalized || /[\x00-\x1F\x7F]/u.test(normalized)) fail(`${label} is invalid`);
  return normalized;
}

function parseOptions(argv) {
  const result = { configs: [] };
  const positionals = [];
  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    if (!argument.startsWith("--")) {
      positionals.push(argument);
      continue;
    }
    if (argument === "--help" || argument === "--definition-check") {
      result[argument.slice(2)] = true;
      continue;
    }
    const key = argument.slice(2);
    const value = argv[index + 1];
    if (!value || value.startsWith("--")) fail(`${argument} requires a value`);
    index += 1;
    if (key === "credential-config") {
      result.configs.push(value);
    } else if (["pm", "credential-env", "credential-field", "root", "receipt-file", "stages-file", "limit"].includes(key)) {
      if (result[key] !== undefined) fail(`${argument} may be provided once`);
      result[key] = value;
    } else {
      fail(`unknown option ${argument}`);
    }
  }
  if (result.help) return result;
  if (positionals.length !== 1) fail("exactly one connector argument is required");
  result.connector = requireIdentifier(positionals[0], "connector");
  return result;
}

async function readJSON(file, label) {
  let raw;
  try {
    raw = await readFile(file, "utf8");
  } catch (error) {
    fail(`${label} is unavailable: ${error instanceof Error ? error.message : "read failed"}`);
  }
  try {
    return JSON.parse(raw);
  } catch {
    fail(`${label} is not valid JSON`);
  }
}

async function readOptionalJSON(file, label) {
  try {
    return await readJSON(file, label);
  } catch (error) {
    if (error instanceof Error && error.message.includes("ENOENT")) return null;
    throw error;
  }
}

async function selectedStageRoute(file, connector) {
  const source = requireObject(await readJSON(withinRepository(requirePathText(file, "--stages-file")), "stages file"), "stages file");
  if (source.connector !== connector) fail("stages file connector does not match the selected connector");
  if (!Array.isArray(source.stage_names) || source.stage_names.length === 0) fail("stages file must declare one or more stage_names");
  const stageNames = new Set(source.stage_names.map((stageName) => requireIdentifier(stageName, "stages file stage_name")));
  if (stageNames.size !== source.stage_names.length) fail("stages file has duplicate stage_names");
  const outcomes = new Map();
  if (source.nonpass_outcomes !== undefined) {
    const declared = requireObject(source.nonpass_outcomes, "stages file nonpass_outcomes");
    for (const [stageName, outcome] of Object.entries(declared)) {
      requireIdentifier(stageName, "stages file nonpass outcome stage_name");
      if (!stageNames.has(stageName)) fail(`stages file declares a nonpass outcome for unselected ${stageName}`);
      if (typeof outcome !== "string" || !nonPassOutcomes.has(outcome)) {
        fail(`stages file nonpass outcome for ${stageName} is invalid`);
      }
      outcomes.set(stageName, outcome);
    }
  }
  return { stageNames, outcomes };
}

function requireObject(value, label) {
  if (!value || typeof value !== "object" || Array.isArray(value)) fail(`${label} must be an object`);
  return value;
}

function findCandidateDeclarations(value, out = []) {
  if (Array.isArray(value)) {
    for (const item of value) findCandidateDeclarations(item, out);
    return out;
  }
  if (!value || typeof value !== "object") return out;
  if (typeof value.stage_name === "string" && typeof value.command === "string" &&
      Array.isArray(value.args) && Array.isArray(value.output_assertions)) {
    out.push(value);
  }
  for (const child of Object.values(value)) findCandidateDeclarations(child, out);
  return out;
}

function commandIndex(surface) {
  if (!Array.isArray(requireObject(surface, "cli_surface.json").commands)) {
    fail("cli_surface.json must declare commands");
  }
  const indexed = new Map();
  for (const item of surface.commands) {
    if (!item || typeof item !== "object") continue;
    const command = String(item.path || "").trim();
    if (command) indexed.set(command, item);
  }
  return indexed;
}

function normalizeCandidate(candidate, commands) {
  const stageName = requireIdentifier(candidate.stage_name, "candidate stage_name");
  const command = String(candidate.command || "").trim();
  if (!command || /[\x00-\x1F\x7F]/u.test(command)) fail(`candidate ${stageName} has an invalid command`);
  const surfaceCommand = commands.get(command);
  if (!surfaceCommand || surfaceCommand.availability !== "implemented") {
    fail(`candidate ${stageName} command is not implemented in cli_surface.json`);
  }
  if (surfaceCommand.intent !== "direct_read") {
    fail(`candidate ${stageName} is not a definition-declared direct read`);
  }
  if (!Array.isArray(surfaceCommand.api_surface) || surfaceCommand.api_surface.length !== 1) {
    fail(`candidate ${stageName} requires one declared API surface entry`);
  }
  const api = requireObject(surfaceCommand.api_surface[0], `candidate ${stageName} API surface`);
  const method = String(api.method || "").toUpperCase();
  const requestPath = String(api.path || "").trim();
  if (!httpMethods.has(method) || !requestPath) fail(`candidate ${stageName} API surface is incomplete`);
  if (candidate.output_assertions.length === 0) fail(`candidate ${stageName} has no produced-value assertion`);
  return { stageName, command, args: candidate.args, assertions: candidate.output_assertions, api: { method, path: requestPath } };
}

function configMap(rawConfigs) {
  const result = new Map();
  for (const raw of rawConfigs) {
    const separator = raw.indexOf("=");
    if (separator < 1) fail("--credential-config must be key=value");
    const key = raw.slice(0, separator);
    const value = raw.slice(separator + 1);
    if (!safeFieldName.test(key) || !value || /[\x00-\x1F\x7F]/u.test(value)) {
      fail("--credential-config must contain safe non-secret key=value");
    }
    if (result.has(key)) fail(`duplicate --credential-config ${key}`);
    result.set(key, value);
  }
  return result;
}

function declaredCredentialConfig(certification) {
  const source = requireObject(certification, "certification.json");
  const result = new Map();
  const declarationSources = [source];
  if (source.source !== undefined) declarationSources.push(requireObject(source.source, "certification.json source"));
  for (const declarationSource of declarationSources) {
    for (const sectionName of ["source_credential_defaults", "required_credential_config"]) {
      const section = declarationSource[sectionName];
      if (section === undefined) continue;
      const values = requireObject(section, `certification.json ${sectionName}`);
      for (const [key, value] of Object.entries(values)) {
        if (!safeFieldName.test(key) || typeof value !== "string" || !value || /[\x00-\x1F\x7F]/u.test(value)) {
          fail(`certification.json ${sectionName} must contain safe non-secret string values`);
        }
        result.set(key, value);
      }
    }
  }
  return result;
}

function candidateConfigValue(candidate, key, values) {
  if (values.has(key)) return values.get(key);
  for (const item of candidate.args) {
    if (item && typeof item === "object" && item.config_key === key && typeof item.default === "string") {
      return item.default;
    }
  }
  return "";
}

function renderArguments(candidate, connector, credentialName, values) {
  const rendered = [];
  for (const item of candidate.args) {
    const source = requireObject(item, `candidate ${candidate.stageName} arg`);
    if (source.connector === true && Object.keys(source).length === 1) {
      rendered.push(connector);
      continue;
    }
    if (source.source_credential === true && Object.keys(source).length === 1) {
      rendered.push(credentialName);
      continue;
    }
    if (typeof source.literal === "string" && Object.keys(source).every((key) => key === "literal" || key === "when_config_key")) {
      if (source.when_config_key !== undefined && !candidateConfigValue(candidate, source.when_config_key, values)) continue;
      if (/[\x00-\x1F\x7F]/u.test(source.literal)) fail(`candidate ${candidate.stageName} literal is unsafe`);
      rendered.push(source.literal);
      continue;
    }
    if (typeof source.config_key === "string" && Object.keys(source).every((key) =>
      ["config_key", "default", "when_config_key", "omit_when_empty"].includes(key))) {
      const key = source.config_key;
      if (!safeFieldName.test(key)) fail(`candidate ${candidate.stageName} configuration key is invalid`);
      if (source.when_config_key !== undefined && !candidateConfigValue(candidate, source.when_config_key, values)) continue;
      const value = candidateConfigValue(candidate, key, values);
      if (!value && source.omit_when_empty === true) continue;
      if (!value) fail(`candidate ${candidate.stageName} needs configuration ${key}`);
      if (/[\x00-\x1F\x7F]/u.test(value)) fail(`candidate ${candidate.stageName} configuration ${key} is unsafe`);
      rendered.push(value);
      continue;
    }
    fail(`candidate ${candidate.stageName} arg has an unsupported definition shape`);
  }
  if (rendered[0] !== connector || !rendered.includes(credentialName) || !rendered.includes("--json")) {
    fail(`candidate ${candidate.stageName} does not bind its connector, credential, and JSON output from the declaration`);
  }
  return rendered;
}

function jsonPointer(value, pointer) {
  if (pointer === "") return value;
  if (!pointer.startsWith("/")) fail("output assertion json_pointer must be RFC 6901");
  let current = value;
  for (const rawSegment of pointer.slice(1).split("/")) {
    const segment = rawSegment.replaceAll("~1", "/").replaceAll("~0", "~");
    if (Array.isArray(current)) {
      if (!/^(0|[1-9][0-9]*)$/u.test(segment)) return undefined;
      current = current[Number(segment)];
    } else if (current && typeof current === "object" && Object.prototype.hasOwnProperty.call(current, segment)) {
      current = current[segment];
    } else {
      return undefined;
    }
  }
  return current;
}

function matchesType(value, type) {
  if (type === "object") return value !== null && typeof value === "object" && !Array.isArray(value);
  if (type === "array") return Array.isArray(value);
  if (type === "object_or_array") return value !== null && typeof value === "object";
  if (type === "string") return typeof value === "string";
  if (type === "number") return typeof value === "number" && Number.isFinite(value);
  if (type === "boolean") return typeof value === "boolean";
  if (type === "null") return value === null;
  fail(`output assertion has unsupported value_type ${String(type)}`);
}

function assertProducedValues(envelope, assertions) {
  if (!envelope || typeof envelope !== "object" || !Object.prototype.hasOwnProperty.call(envelope, "response")) {
    return { ok: false, reason: "pm JSON envelope omitted provider response" };
  }
  for (const assertion of assertions) {
    const source = requireObject(assertion, "output assertion");
    const pointer = String(source.json_pointer || "");
    const actual = jsonPointer(envelope, pointer);
    if (actual === undefined) return { ok: false, reason: `returned response omitted ${pointer}` };
    if (source.value_type !== undefined && !matchesType(actual, source.value_type)) {
      return { ok: false, reason: `returned response at ${pointer} did not have declared type ${source.value_type}` };
    }
    if (Object.prototype.hasOwnProperty.call(source, "equals") && source.equals !== null &&
      JSON.stringify(actual) !== JSON.stringify(source.equals)) {
      return { ok: false, reason: `returned response at ${pointer} did not equal the declaration-owned value` };
    }
  }
  return { ok: true };
}

function runProcess(binary, args, cwd, timeoutMS = 45_000) {
  return new Promise((resolve, reject) => {
    const child = spawn(binary, args, { cwd, stdio: ["ignore", "pipe", "pipe"] });
    let stdout = "";
    let stderr = "";
    let size = 0;
    let overflow = false;
    let timedOut = false;
    const timeout = setTimeout(() => {
      timedOut = true;
      child.kill("SIGTERM");
      setTimeout(() => child.kill("SIGKILL"), 1_000).unref();
    }, timeoutMS);
    const consume = (target, chunk) => {
      size += chunk.length;
      if (size > maxOutputBytes) {
        overflow = true;
        child.kill("SIGTERM");
        return;
      }
      if (target === "stdout") stdout += chunk.toString("utf8");
      else stderr += chunk.toString("utf8");
    };
    child.once("error", (error) => { clearTimeout(timeout); reject(error); });
    child.stdout.on("data", (chunk) => consume("stdout", chunk));
    child.stderr.on("data", (chunk) => consume("stderr", chunk));
    child.once("close", (code, signal) => { clearTimeout(timeout); resolve({ code, signal, stdout, stderr, overflow, timedOut }); });
  });
}

function providerDiagnostic(result, credential) {
  let message = `${result.stdout}\n${result.stderr}`;
  if (credential) message = message.replaceAll(credential, "<credential-redacted>");
  message = message.trim().slice(0, 8_192);
  if (message) return message;
  if (result.timedOut) return "provider command exceeded the declared process timeout";
  if (result.overflow) return "provider command exceeded the bounded captured-output limit";
  return "provider command exited without a provider diagnostic";
}

function classifyNonPass(reason, source, routedOutcome) {
  if (routedOutcome !== undefined) return routedOutcome;
  const lower = reason.toLowerCase();
  if (/\b(?:http|status)\s+401\b/u.test(lower) || lower.includes("bad credentials") || lower.includes("authentication required")) return "wrong_credential";
  if (/\b(?:http|status)\s+(?:402|403)\b/u.test(lower) || lower.includes("payment required") || lower.includes("not authorized") || lower.includes("not accessible")) return "entitlement";
  if (/\b(?:http|status)\s+404\b/u.test(lower) || /\b(?:not found|no .+ found|does not exist|missing)\b/u.test(lower)) return "no_object";
  for (const unavailable of source.live_unavailable || []) {
    const parts = Array.isArray(unavailable?.contains) ? unavailable.contains.map((item) => String(item).toLowerCase()) : [];
    if (parts.length > 0 && parts.some((part) => lower.includes(part))) return "entitlement";
  }
  if (lower.includes("configuration") || lower.includes("fixture") || lower.includes("unresolved")) return "no_object";
  return "product_defect";
}

async function repositorySalt() {
  try {
    const current = await readFile(saltPath);
    if (current.length < 16) fail("repository fingerprint salt is too short");
    return current;
  } catch (error) {
    if (error?.code !== "ENOENT") throw error;
  }
  await mkdir(path.dirname(saltPath), { recursive: true, mode: 0o700 });
  const fresh = randomBytes(32);
  try {
    const file = await open(saltPath, "wx", 0o600);
    await file.writeFile(fresh);
    await file.close();
    return fresh;
  } catch (error) {
    if (error?.code === "EEXIST") return repositorySalt();
    throw error;
  }
}

function fingerprint(salt, value) {
  return `${markerPrefix}${createHmac("sha256", salt).update(String(value)).digest("hex")}${markerSuffix}`;
}

function sanitizedJSON(value, salt) {
  if (value === null) return null;
  if (Array.isArray(value)) return value.map((item) => sanitizedJSON(item, salt));
  if (typeof value === "object") {
    const out = {};
    for (const [key, child] of Object.entries(value)) {
      out[safeFieldName.test(key) ? key : fingerprint(salt, key)] = sanitizedJSON(child, salt);
    }
    return out;
  }
  return fingerprint(salt, value);
}

function responseBody(response, salt) {
  if (response === undefined) return { encoding: "none", value: null, original_bytes: 0, truncated: false };
  const raw = JSON.stringify(response);
  return { encoding: "json", value: sanitizedJSON(response, salt), original_bytes: Buffer.byteLength(raw), truncated: false };
}

function evidencePayload({ connector, stageName, binarySHA, invocation, credential, salt, api, envelope, runID }) {
  return {
    schema_version: 2,
    scope: "capability",
    status: "passed",
    credential_scope: "observed_operations",
    credential_note: credentialNote,
    credential_scope_proof: "protocol_exchanges",
    connector,
    function_kind: `command:${stageName}`,
    provider: connector,
    executed_at: new Date().toISOString(),
    run_id: runID,
    proof: {
      redaction_strategy: "repository_salted_hmac_sha256_v1",
      pm_binary_sha256: binarySHA,
      pm_command_fingerprint: fingerprint(salt, invocation.join("\u0000")),
      credential_fingerprints: [fingerprint(salt, credential)],
      http_exchanges: [{
        operation: stageName,
        request: {
          method: api.method,
          target: fingerprint(salt, api.path),
          query: [],
          headers: [],
          body: { encoding: "none", value: null, original_bytes: 0, truncated: false },
        },
        response: { status: envelope.status, headers: [], body: responseBody(envelope.response, salt) },
      }],
      database_exchanges: [],
    },
  };
}

function withinRepository(file) {
  const normalized = path.resolve(file);
  if (!normalized.startsWith(`${repositoryRoot}${path.sep}`)) fail("output path must stay inside the repository");
  return normalized;
}

async function atomicJSON(file, value) {
  await mkdir(path.dirname(file), { recursive: true, mode: 0o700 });
  const temporary = `${file}.${process.pid}.${randomBytes(4).toString("hex")}.tmp`;
  await writeFile(temporary, `${JSON.stringify(value, null, 2)}\n`, { mode: 0o600 });
  await rename(temporary, file);
  await chmod(file, 0o600);
}

async function writeEvidenceDraft(record, recordName) {
  const draft = path.join(temporaryRoot, "drafts", `${recordName}.json`);
  await atomicJSON(draft, record);
  return draft;
}

async function importEvidenceDraft(draft) {
  const result = await runProcess("go", ["run", "./cmd/connectorgen", "certification-evidence", "draft", "--draft", draft, "--repo-root", repositoryRoot], repositoryRoot, 120_000);
  return result.code === 0 && !result.timedOut && !result.overflow ? "" : providerDiagnostic(result, "");
}

async function generateConnectorMatrix(connector) {
  const result = await runProcess("go", ["run", "./cmd/connectorgen", "certification-matrix", "--connector", connector], repositoryRoot, 120_000);
  return result.code === 0 && !result.timedOut && !result.overflow ? "" : providerDiagnostic(result, "");
}

async function checkConnectorMatrix(connector) {
  const result = await runProcess("go", ["run", "./cmd/connectorgen", "certification-matrix", "--connector", connector, "--check"], repositoryRoot, 120_000);
  return result.code === 0 && !result.timedOut && !result.overflow ? "" : providerDiagnostic(result, "");
}

async function prepareProject(options, connector, runID) {
  if (options.root !== undefined) {
    const root = withinRepository(requirePathText(options.root, "--root"));
    await mkdir(root, { recursive: true, mode: 0o700 });
    return { root, ownsRoot: false, credential: requireIdentifier(`${connector}-cert-${runID}`, "generated credential name") };
  }
  await mkdir(temporaryRoot, { recursive: true, mode: 0o700 });
  const root = await mkdtemp(path.join(temporaryRoot, "run-"));
  return { root, ownsRoot: true, credential: requireIdentifier(`${connector}-cert-${runID}`, "generated credential name") };
}

async function addCredential(binary, project, connector, options, configs) {
  const environmentName = String(options["credential-env"] || "");
  if (!safeEnvironmentName.test(environmentName)) fail("--credential-env must name a safe environment variable");
  const credential = process.env[environmentName];
  if (!credential) fail(`credential environment ${environmentName} is unavailable`);
  const field = String(options["credential-field"] || "");
  if (!safeFieldName.test(field)) fail("--credential-field must be a safe credential field name");
  const init = await runProcess(binary, ["init", "--root", project.root, "--json"], repositoryRoot);
  if (init.code !== 0) fail("could not initialize the run-owned pm project");
  const args = ["credentials", "add", project.credential, "--connector", connector, "--from-env", `${field}=${environmentName}`];
  for (const [key, value] of configs) args.push("--config", `${key}=${value}`);
  args.push("--root", project.root, "--json");
  const added = await runProcess(binary, args, repositoryRoot);
  if (added.code !== 0) fail("could not add the run-owned credential");
  return credential;
}

async function cleanupProject(binary, project) {
  let credentialRemoved = false;
  try {
    const removed = await runProcess(binary, ["credentials", "remove", project.credential, "--root", project.root, "--json"], repositoryRoot);
    credentialRemoved = removed.code === 0;
  } catch {
    credentialRemoved = false;
  }
  if (project.ownsRoot) await rm(project.root, { recursive: true, force: true });
  return credentialRemoved;
}

function initialReceipt(connector, metadata, runID) {
  return {
    schema_version: 1,
    connector,
    integration_type: metadata.integration_type,
    run_id: runID,
    started_at: new Date().toISOString(),
    credential_scope: "observed_operations",
    records: [],
    summary: { executed: 0, certified: 0, no_object: 0, wrong_credential: 0, entitlement: 0, product_defect: 0 },
    local_cleanup: { credential_removed: null },
  };
}

function addReceipt(receipt, record) {
  receipt.records.push(record);
  if (record.executed) receipt.summary.executed += 1;
  if (Object.hasOwn(receipt.summary, record.outcome)) receipt.summary[record.outcome] += 1;
}

async function persistReceipt(file, receipt) {
  receipt.updated_at = new Date().toISOString();
  await atomicJSON(file, receipt);
}

async function main() {
  const options = parseOptions(process.argv.slice(2));
  if (options.help) {
    process.stdout.write(usage());
    return 0;
  }
  const definitionDirectory = path.join(definitionsRoot, options.connector);
  const [metadata, surface, certification, sweep] = await Promise.all([
    readJSON(path.join(definitionDirectory, "metadata.json"), "metadata.json"),
    readJSON(path.join(definitionDirectory, "cli_surface.json"), "cli_surface.json"),
    readOptionalJSON(path.join(definitionDirectory, "certification.json"), "certification.json"),
    readOptionalJSON(path.join(definitionDirectory, "certification-sweep.json"), "certification-sweep.json"),
  ]);
  if (requireObject(metadata, "metadata.json").name !== options.connector) {
    fail("metadata.json connector name does not match its directory");
  }
  const commands = commandIndex(surface);
  const allCandidates = certification === null ? [] : findCandidateDeclarations(certification).map((item) => normalizeCandidate(item, commands));
  if (new Set(allCandidates.map((item) => item.stageName)).size !== allCandidates.length) fail("certification.json has duplicate candidate stage names");
  let candidates = allCandidates;
  if (sweep !== null) {
    const sweepCommands = requireObject(sweep, "certification-sweep.json").commands;
    if (!Array.isArray(sweepCommands)) fail("certification-sweep.json must declare commands");
    const eligible = new Set(sweepCommands
      .filter((item) => item && typeof item === "object" && item.status === "eligible_pending_live")
      .map((item) => String(item.path || "").trim())
      .filter(Boolean));
    const declared = new Set(allCandidates.map((item) => item.command));
    for (const command of eligible) {
      if (!declared.has(command)) fail(`certification-sweep.json eligible command ${command} has no certification.json candidate`);
    }
    candidates = allCandidates.filter((item) => eligible.has(item.command));
  }
  if (options["stages-file"] !== undefined) {
    const selectedRoute = await selectedStageRoute(options["stages-file"], options.connector);
    const selected = selectedRoute.stageNames;
    const available = new Set(candidates.map((item) => item.stageName));
    for (const stageName of selected) {
      if (!available.has(stageName)) fail(`stages file selected ${stageName}, which is not an eligible declared candidate`);
    }
    candidates = candidates.filter((item) => selected.has(item.stageName));
    options.nonpassOutcomes = selectedRoute.outcomes;
  }
  const definitionConfigs = certification === null ? new Map() : declaredCredentialConfig(certification);
  if (options["definition-check"]) {
    process.stdout.write(`${options.connector} definition check: commands=${commands.size} candidates=${allCandidates.length} selected=${candidates.length} credential_configs=${definitionConfigs.size}\n`);
    return 0;
  }

  const runID = requireIdentifier(`${options.connector}-${Date.now()}-${randomBytes(6).toString("hex")}`, "run identifier");
  const receiptPath = options["receipt-file"] === undefined
    ? path.join(receiptRoot, `${options.connector}-${runID}.json`)
    : withinRepository(requirePathText(options["receipt-file"], "--receipt-file"));
  const receipt = initialReceipt(options.connector, metadata, runID);
  if (candidates.length === 0) {
    const reason = certification === null
      ? "connector definitions do not declare certification.json"
      : "connector definitions declare no produced-value certification candidates";
    addReceipt(receipt, { stage_name: "definition_candidates", executed: false, outcome: "no_object", reason });
    await persistReceipt(receiptPath, receipt);
    process.stdout.write(`${options.connector} certification: executed=0 certified=0 no_object=1 wrong_credential=0 entitlement=0 product_defect=0\n`);
    return 0;
  }

  const binary = requirePathText(options.pm, "--pm");
  const binarySHA = createHash("sha256").update(await readFile(binary)).digest("hex");
  const values = new Map(definitionConfigs);
  for (const [key, value] of configMap(options.configs)) values.set(key, value);
  const limit = options.limit === undefined ? candidates.length : Number.parseInt(options.limit, 10);
  if (!Number.isSafeInteger(limit) || limit < 1) fail("--limit must be a positive integer");
  const project = await prepareProject(options, options.connector, runID);
  let credential = "";
  try {
    credential = await addCredential(binary, project, options.connector, options, values);
    const salt = await repositorySalt();
    for (const candidate of candidates.slice(0, limit)) {
      let rendered;
      try {
        rendered = renderArguments(candidate, options.connector, project.credential, values);
      } catch (error) {
        addReceipt(receipt, { stage_name: candidate.stageName, command: candidate.command, executed: false, outcome: "no_object", reason: error instanceof Error ? error.message : "candidate configuration could not be rendered" });
        await persistReceipt(receiptPath, receipt);
        continue;
      }
      const args = [...rendered, "--root", project.root];
      const invocation = [binary, ...args];
      let result;
      try {
        result = await runProcess(binary, args, repositoryRoot);
      } catch (error) {
        addReceipt(receipt, { stage_name: candidate.stageName, command: candidate.command, invocation, executed: false, outcome: "product_defect", reason: error instanceof Error ? error.message : "provider command failed to start" });
        await persistReceipt(receiptPath, receipt);
        continue;
      }
      if (result.code !== 0 || result.timedOut || result.overflow) {
        const reason = providerDiagnostic(result, credential);
        addReceipt(receipt, { stage_name: candidate.stageName, command: candidate.command, invocation, executed: true, outcome: classifyNonPass(reason, certification?.source || {}, options.nonpassOutcomes?.get(candidate.stageName)), provider_response: reason });
        await persistReceipt(receiptPath, receipt);
        continue;
      }
      let envelope;
      try {
        envelope = JSON.parse(result.stdout);
      } catch {
        addReceipt(receipt, { stage_name: candidate.stageName, command: candidate.command, invocation, executed: true, outcome: "product_defect", reason: "pm exited successfully without machine-readable JSON output" });
        await persistReceipt(receiptPath, receipt);
        continue;
      }
      if (!Number.isInteger(envelope.status) || envelope.status < 100 || envelope.status > 599) {
        addReceipt(receipt, { stage_name: candidate.stageName, command: candidate.command, invocation, executed: true, outcome: "product_defect", reason: "pm JSON envelope omitted a valid provider HTTP status" });
        await persistReceipt(receiptPath, receipt);
        continue;
      }
      const assertion = assertProducedValues(envelope, candidate.assertions);
      if (!assertion.ok) {
        addReceipt(receipt, { stage_name: candidate.stageName, command: candidate.command, invocation, executed: true, outcome: "product_defect", reason: assertion.reason });
        await persistReceipt(receiptPath, receipt);
        continue;
      }
      const evidenceRunID = requireIdentifier(`${runID}-${candidate.stageName}`, "evidence run identifier");
      const record = evidencePayload({ connector: options.connector, stageName: candidate.stageName, binarySHA, invocation, credential, salt, api: candidate.api, envelope, runID: evidenceRunID });
      const evidenceName = requireIdentifier(`${options.connector}-${candidate.stageName}-${runID}`, "evidence file name");
      const evidencePath = path.join(evidenceRoot, `${evidenceName}.json`);
      if (JSON.stringify(record).includes(credential)) fail("refusing to persist a record containing credential material");
      let draftPath = "";
      let importFailure = "";
      try {
        draftPath = await writeEvidenceDraft(record, evidenceName);
        importFailure = await importEvidenceDraft(draftPath);
      } catch (error) {
        importFailure = `could not import drafted accepted evidence: ${error instanceof Error ? error.message : "write failed"}`;
      } finally {
        if (draftPath) await unlink(draftPath).catch(() => {});
      }
      if (importFailure) {
        addReceipt(receipt, { stage_name: candidate.stageName, command: candidate.command, invocation, executed: true, outcome: "product_defect", reason: importFailure });
        await persistReceipt(receiptPath, receipt);
        continue;
      }
      const generationFailure = await generateConnectorMatrix(options.connector);
      if (generationFailure) {
        addReceipt(receipt, { stage_name: candidate.stageName, command: candidate.command, invocation, executed: true, outcome: "product_defect", reason: `accepted evidence remains published because scoped certification-matrix generation failed: ${generationFailure}` });
        await persistReceipt(receiptPath, receipt);
        continue;
      }
      const matrixFailure = await checkConnectorMatrix(options.connector);
      if (matrixFailure) {
        addReceipt(receipt, { stage_name: candidate.stageName, command: candidate.command, invocation, executed: true, outcome: "product_defect", reason: `accepted evidence remains published because scoped certification-matrix --check failed: ${matrixFailure}` });
        await persistReceipt(receiptPath, receipt);
        continue;
      }
      addReceipt(receipt, {
        stage_name: candidate.stageName,
        command: candidate.command,
        invocation,
        executed: true,
        outcome: "certified",
        asserted_values: candidate.assertions,
        evidence: path.relative(repositoryRoot, evidencePath),
        request_shape_source: "cli_surface.json",
        provider_status: envelope.status,
      });
      await persistReceipt(receiptPath, receipt);
    }
  } finally {
    receipt.local_cleanup.credential_removed = await cleanupProject(binary, project);
    await persistReceipt(receiptPath, receipt);
  }
  process.stdout.write(`${options.connector} certification: executed=${receipt.summary.executed} certified=${receipt.summary.certified} no_object=${receipt.summary.no_object} wrong_credential=${receipt.summary.wrong_credential} entitlement=${receipt.summary.entitlement} product_defect=${receipt.summary.product_defect}\n`);
  return receipt.summary.product_defect === 0 ? 0 : 1;
}

main().then((code) => { process.exitCode = code; }).catch((error) => {
  process.stderr.write(`connector certification runner: ${error instanceof Error ? error.message : "failed"}\n`);
  process.exitCode = 2;
});
