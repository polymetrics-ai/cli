#!/usr/bin/env node

import { createHash } from "node:crypto";
import { readFile, writeFile, chmod } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";
import { spawn } from "node:child_process";

import { buildCases, resolveLiveBoundary } from "./github-live-cases.mjs";

const CONNECTOR = "github";
const TERMINAL_STATES = new Set(["proven", "untestable", "failed"]);
const FORBIDDEN_RECORD_FIELDS = new Set([
  "stdout",
  "stderr",
  "output",
  "response",
  "body",
  "approval_token",
  "approval_grant",
  "token",
  "grant",
]);
const SENSITIVE_ARGUMENT_NAMES = new Set([
  "access-token",
  "approve",
  "authorization",
  "client-secret",
  "encrypted-value",
  "key",
  "password",
  "private-key",
  "private-key-base64",
  "secret",
  "token",
  "value",
]);
const LIFECYCLE_ARGUMENT_NAMES = new Set([
  "approve",
  "confirm",
  "connection",
  "credential",
  "plan",
  "preview",
  "root",
]);
const SENSITIVE_CONFIG_KEY = /(?:^|[-_])(?:approval|authorization|credential|grant|password|private[-_]?key|secret|token|value)(?:$|[-_])/iu;
const TOKEN_PATTERNS = [
  /\bgh[pousr]_[A-Za-z0-9_-]+\b/gi,
  /\bgithub_pat_[A-Za-z0-9_-]+\b/gi,
  /\b(?:bearer|token)\s+[A-Za-z0-9._~+\/-]{12,}\b/gi,
];
const OUTPUT_LIMIT_BYTES = 2 * 1024 * 1024;
const PROCESS_TIMEOUT_MS = 45_000;
const PROCESS_KILL_GRACE_MS = 1_000;
const APP_INSTALLATION_REPOSITORY_PREFLIGHT = "apps list-repos-accessible-to-installation";
const CANONICAL_GITHUB_API_ORIGIN = "https://api.github.com";
const CONTROL_CHARACTER = /[\u0000-\u001F\u007F]/u;
const PEM_ARMOR = /-----BEGIN [A-Z0-9][A-Z0-9 -]{0,80}-----/iu;
const PEM_BLOCK = /-----BEGIN [\s\S]{0,8192}?-----END [^-]{1,80}-----/gu;
const SAFE_IDENTIFIER = /^[A-Za-z0-9][A-Za-z0-9_.:-]{0,255}$/u;
const GITHUB_SLUG = /^[A-Za-z0-9](?:[A-Za-z0-9-]{0,98})$/u;
const MAX_INSTALLATION_PREFLIGHT_PAGES = 10_000;
const EXTERNAL_BLOCKERS = Object.freeze({
  app_installation_credential_unavailable:
    "The captain-authorized GitHub App installation credential is unavailable to this proof runner.",
  run_owned_boundary_unavailable:
    "The immutable Polymetrics-Cert run boundary is unavailable to this proof runner.",
  provider_admission_unavailable:
    "The external GitHub provider admission prerequisite is unavailable to this proof runner.",
});
const SCRIPT_DIR = path.dirname(fileURLToPath(import.meta.url));
const REPOSITORY_ROOT = path.resolve(SCRIPT_DIR, "..");
const SURFACE_PATH = path.join(
  REPOSITORY_ROOT,
  "internal/connectors/defs/github/cli_surface.json",
);

/**
 * Return the complete, stable set that a proof run is responsible for.  This
 * intentionally derives from the production GitHub bundle rather than a
 * hand-maintained command list.
 */
export function enumerateImplementedCommands(surface) {
  if (!surface || !Array.isArray(surface.commands)) {
    throw new Error("GitHub command surface must contain a commands array");
  }
  const commands = surface.commands
    .filter((command) => command?.availability === "implemented")
    .map((command) => String(command.path || "").trim());
  if (commands.some((command) => command === "")) {
    throw new Error("implemented GitHub command is missing a path");
  }
  const unique = new Set(commands);
  if (unique.size !== commands.length) {
    throw new Error("GitHub command surface has duplicate implemented command paths");
  }
  return [...unique].sort((left, right) => left.localeCompare(right));
}

function normalizeArgumentName(value) {
  return String(value || "").trim().toLowerCase().replaceAll("_", "-");
}

function isSensitiveArgumentName(value) {
  return SENSITIVE_ARGUMENT_NAMES.has(normalizeArgumentName(value));
}

function isSensitiveConfigKey(value) {
  return SENSITIVE_CONFIG_KEY.test(normalizeArgumentName(value));
}

function splitLongFlag(argument) {
  const text = String(argument);
  if (!text.startsWith("--") || text === "--") return null;
  const body = text.slice(2);
  const equals = body.indexOf("=");
  const rawName = equals === -1 ? body : body.slice(0, equals);
  return {
    rawName,
    name: normalizeArgumentName(rawName),
    hasValue: equals !== -1,
    value: equals === -1 ? undefined : body.slice(equals + 1),
  };
}

function redactText(value) {
  let out = String(value ?? "");
  out = out.replace(PEM_BLOCK, "<redacted>");
  for (const pattern of TOKEN_PATTERNS) {
    out = out.replace(pattern, "<redacted>");
  }
  return out;
}

function redactConfigValue(value) {
  const raw = String(value ?? "");
  const separator = raw.indexOf("=");
  if (separator < 1) return "<redacted>";
  const key = raw.slice(0, separator);
  return `${redactText(key)}=<redacted>`;
}

function hasTokenShapedValue(value) {
  const text = String(value ?? "");
  return TOKEN_PATTERNS.some((pattern) => {
    pattern.lastIndex = 0;
    return pattern.test(text);
  });
}

function isPlainObject(value) {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function requireKnownFields(value, fields, label) {
  if (!isPlainObject(value)) {
    throw new Error(`${label} must be an object`);
  }
  if (Object.keys(value).some((key) => !fields.has(key))) {
    throw new Error(`${label} contains an unsupported field`);
  }
  return value;
}

function requireSafeText(value, label) {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`${label} must be a non-empty string`);
  }
  if (CONTROL_CHARACTER.test(value) || PEM_ARMOR.test(value) || hasTokenShapedValue(value)) {
    throw new Error(`${label} contains unsafe credential-like material`);
  }
  return value.trim();
}

function requireSafeIdentifier(value, label) {
  const text = requireSafeText(value, label);
  if (!SAFE_IDENTIFIER.test(text)) {
    throw new Error(`${label} must be a well-formed immutable identifier`);
  }
  return text;
}

function requireGitHubSlug(value, label) {
  const slug = requireSafeText(value, label);
  if (!GITHUB_SLUG.test(slug)) {
    throw new Error(`${label} must be a well-formed GitHub slug`);
  }
  return slug.toLowerCase();
}

function requireProviderID(value, label) {
  if (typeof value === "number" && Number.isSafeInteger(value) && value > 0) return String(value);
  return requireSafeIdentifier(value, label);
}

function requireNonNegativeInteger(value, label) {
  if (!Number.isSafeInteger(value) || value < 0) {
    throw new Error(`${label} must be a non-negative integer`);
  }
  return value;
}

function redactInvocation(invocation) {
  if (!Array.isArray(invocation)) {
    throw new Error("proof invocation must be an argument array");
  }
  const out = [];
  let redactNext = false;
  for (let index = 0; index < invocation.length; index += 1) {
    const rawArgument = invocation[index];
    const argument = String(rawArgument);
    if (redactNext) {
      out.push("<redacted>");
      redactNext = false;
      continue;
    }
    if (!argument.startsWith("--")) {
      out.push(redactText(argument));
      continue;
    }
    const flag = splitLongFlag(argument);
    if (!flag) {
      out.push(redactText(argument));
      continue;
    }
    if (flag.name === "config") {
      if (!flag.hasValue) {
        out.push(argument);
        const next = invocation[index + 1];
        if (next !== undefined && !String(next).startsWith("--")) {
          out.push(redactConfigValue(next));
          index += 1;
        }
      } else {
        out.push(`--config=${redactConfigValue(flag.value)}`);
      }
      continue;
    }
    if (isSensitiveArgumentName(flag.name)) {
      if (!flag.hasValue) {
        out.push(argument);
        redactNext = true;
      } else {
        out.push(`--${flag.name}=<redacted>`);
      }
      continue;
    }
    out.push(redactText(argument));
  }
  return out;
}

function validateProofAssertion(assertion, label = "proof assertion") {
  const source = requireKnownFields(assertion, new Set([
    "kind",
    "subject",
    "matched",
    "count",
    "pages",
    "readback_kind",
  ]), label);
  const out = {};
  for (const [key, value] of Object.entries(source)) {
    if (typeof value === "string") {
      out[key] = requireSafeText(value, `${label}.${key}`);
    } else if (typeof value === "number" || typeof value === "boolean" || value === null) {
      if (typeof value === "number") requireNonNegativeInteger(value, `${label}.${key}`);
      out[key] = value;
    } else {
      throw new Error(`${label}.${key} must be a scalar`);
    }
  }
  return out;
}

function safeAssertion(assertion) {
  if (!assertion || typeof assertion !== "object" || Array.isArray(assertion)) {
    throw new Error("proof assertion must be an object");
  }
  const redacted = {};
  for (const [key, value] of Object.entries(assertion)) {
    redacted[key] = typeof value === "string" ? redactText(value) : value;
  }
  return validateProofAssertion(redacted);
}

/**
 * Project an internal execution result onto the deliberately small report
 * schema. Raw stdout/stderr, response bodies, grants and credentials are not
 * report fields and are discarded even if a caller accidentally supplies them.
 */
export function redactForReport(candidate) {
  if (!candidate || typeof candidate !== "object" || Array.isArray(candidate)) {
    throw new Error("proof record must be an object");
  }
  const record = {
    command: String(candidate.command || "").trim(),
    state: String(candidate.state || "").trim(),
  };
  if (Array.isArray(candidate.invocation)) {
    record.invocation = redactInvocation(candidate.invocation);
  }
  if (Number.isInteger(candidate.http_status)) {
    record.http_status = candidate.http_status;
  }
  if (candidate.assertion !== undefined) {
    record.assertion = safeAssertion(candidate.assertion);
  }
  if (candidate.reason !== undefined) {
    record.reason = redactText(candidate.reason).trim();
  }
  return record;
}

function concreteReason(reason) {
  const normalized = String(reason || "").trim();
  if (normalized.length < 24) {
    return false;
  }
  const generic = new Set([
    "not tested",
    "unknown",
    "n/a",
    "todo",
    "requires access",
    "unsupported",
  ]);
  return !generic.has(normalized.toLowerCase());
}

function reportContainsForbiddenField(value) {
  if (!value || typeof value !== "object") {
    return false;
  }
  if (Array.isArray(value)) {
    return value.some(reportContainsForbiddenField);
  }
  return Object.entries(value).some(([key, nested]) =>
    FORBIDDEN_RECORD_FIELDS.has(key) || reportContainsForbiddenField(nested),
  );
}

function validateProofInvocation(invocation, label) {
  if (!Array.isArray(invocation)) {
    throw new Error(`${label} must be an argument array`);
  }
  invocation.forEach((argument, index) => requireSafeText(argument, `${label}[${index}]`));
}

function validateProofRecordShape(record) {
  const source = requireKnownFields(record, new Set([
    "command",
    "state",
    "invocation",
    "http_status",
    "assertion",
    "reason",
  ]), "proof record");
  const command = requireSafeText(source.command, "proof record.command");
  const state = requireSafeText(source.state, "proof record.state");
  if (source.invocation !== undefined) {
    validateProofInvocation(source.invocation, "proof record.invocation");
  }
  if (source.http_status !== undefined &&
      (!Number.isSafeInteger(source.http_status) || source.http_status < 100 || source.http_status > 599)) {
    throw new Error("proof record.http_status must be a provider HTTP status");
  }
  if (source.assertion !== undefined) {
    validateProofAssertion(source.assertion, "proof record.assertion");
  }
  if (source.reason !== undefined) {
    requireSafeText(source.reason, "proof record.reason");
  }
  return { command, state };
}

/**
 * Validate exact, terminal accounting for a full sweep.  `failed` is allowed
 * while a repair run is in progress, but it is never a success condition for a
 * final report.
 */
export function validateProofRecords(expectedCommands, records) {
  if (!Array.isArray(expectedCommands) || !Array.isArray(records)) {
    throw new Error("proof accounting requires command and record arrays");
  }
  const expected = new Set(expectedCommands);
  if (expected.size !== expectedCommands.length) {
    throw new Error("expected implemented command list contains duplicates");
  }
  const byCommand = new Map();
  for (const record of records) {
    const { command, state } = validateProofRecordShape(record);
    if (reportContainsForbiddenField(record)) {
      throw new Error("proof record contains raw execution data");
    }
    if (hasTokenShapedValue(JSON.stringify(record))) {
      throw new Error("proof record contains credential-shaped data");
    }
    if (!expected.has(command)) {
      throw new Error("proof record names an unimplemented or unknown command");
    }
    if (byCommand.has(command)) {
      throw new Error("proof accounting contains a duplicate terminal result");
    }
    if (!TERMINAL_STATES.has(state)) {
      throw new Error("proof record has an invalid terminal state");
    }
    if (state === "proven") {
      if (record.http_status !== undefined &&
          (!Number.isInteger(record.http_status) || record.http_status < 200 || record.http_status >= 300)) {
        throw new Error("proven proof record has an invalid HTTP status");
      }
      if (!record.assertion || record.assertion.matched !== true || typeof record.assertion.kind !== "string") {
        throw new Error("proven proof record requires a matched returned-data assertion");
      }
    } else if (!concreteReason(record.reason)) {
      throw new Error("unproven proof record requires a concrete reason");
    }
    byCommand.set(command, record);
  }
  for (const command of expected) {
    if (!byCommand.has(command)) {
      throw new Error(`missing terminal result for ${JSON.stringify(command)}`);
    }
  }
  return summarizeProofRecords(records);
}

export function summarizeProofRecords(records) {
  const tally = { proven: 0, untestable: 0, failed: 0 };
  for (const record of records) {
    if (record?.state in tally) {
      tally[record.state] += 1;
    }
  }
  return tally;
}

function requireSHA256(value, label) {
  const digest = requireSafeText(value, label);
  if (!/^[a-f0-9]{64}$/u.test(digest)) {
    throw new Error(`${label} must be a SHA-256 digest`);
  }
  return digest;
}

function validateReportTally(value, expectedCount) {
  const tally = requireKnownFields(value, new Set(["proven", "untestable", "failed"]), "proof report.tally");
  for (const key of ["proven", "untestable", "failed"]) {
    requireNonNegativeInteger(tally[key], `proof report.tally.${key}`);
  }
  if (tally.proven + tally.untestable + tally.failed !== expectedCount) {
    throw new Error("proof report tally does not account for every implemented command");
  }
  return tally;
}

function validateReportBoundary(value) {
  const boundary = requireKnownFields(value, new Set([
    "run_id",
    "owner",
    "repo",
    "owner_id",
    "repo_id",
    "organization_id",
  ]), "proof report.run_boundary");
  requireSafeIdentifier(boundary.run_id, "proof report.run_boundary.run_id");
  requireGitHubSlug(boundary.owner, "proof report.run_boundary.owner");
  requireGitHubSlug(boundary.repo, "proof report.run_boundary.repo");
  requireProviderID(boundary.owner_id, "proof report.run_boundary.owner_id");
  requireProviderID(boundary.repo_id, "proof report.run_boundary.repo_id");
  requireProviderID(boundary.organization_id, "proof report.run_boundary.organization_id");
}

export function validateProofReport(candidate, expectedCommands) {
  if (!isPlainObject(candidate)) {
    throw new Error("proof report must be an object");
  }
  const status = requireSafeText(candidate.status, "proof report.status");
  const fields = status === "credentialed_live"
    ? new Set([
      "schema_version",
      "connector",
      "status",
      "generated_at",
      "surface_sha256",
      "binary_sha256",
      "case_file_sha256",
      "test_repository",
      "run_boundary",
      "launch",
      "implemented_commands",
      "tally",
      "records",
    ])
    : status === "external_blocker"
      ? new Set([
        "schema_version",
        "connector",
        "status",
        "generated_at",
        "surface_sha256",
        "binary_sha256",
        "test_repository",
        "implemented_commands",
        "blocker",
        "tally",
        "records",
      ])
      : null;
  if (!fields) {
    throw new Error("proof report has an unsupported status");
  }
  const report = requireKnownFields(candidate, fields, "proof report");
  if (report.schema_version !== 1 || requireSafeText(report.connector, "proof report.connector") !== CONNECTOR) {
    throw new Error("proof report must describe the GitHub proof schema");
  }
  const generatedAt = requireSafeText(report.generated_at, "proof report.generated_at");
  if (!/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$/u.test(generatedAt) || Number.isNaN(Date.parse(generatedAt))) {
    throw new Error("proof report.generated_at must be an ISO timestamp");
  }
  requireSHA256(report.surface_sha256, "proof report.surface_sha256");
  requireSHA256(report.binary_sha256, "proof report.binary_sha256");
  const implementedCommands = requireNonNegativeInteger(report.implemented_commands, "proof report.implemented_commands");
  if (!Array.isArray(expectedCommands) || implementedCommands !== expectedCommands.length) {
    throw new Error("proof report implemented command count does not match the current surface");
  }
  const expectedTestRepository = status === "credentialed_live"
    ? "<run-owned-boundary-repository>"
    : "<credentialed-live-proof-not-available>";
  if (requireSafeText(report.test_repository, "proof report.test_repository") !== expectedTestRepository) {
    throw new Error("proof report uses an unsupported test repository projection");
  }
  if (status === "credentialed_live") {
    requireSHA256(report.case_file_sha256, "proof report.case_file_sha256");
    validateReportBoundary(report.run_boundary);
    const launch = requireKnownFields(report.launch, new Set(["strategy", "operations_released"]), "proof report.launch");
    if (requireSafeText(launch.strategy, "proof report.launch.strategy") !== "single_barrier_release") {
      throw new Error("proof report launch strategy is not the required barrier release");
    }
    requireNonNegativeInteger(launch.operations_released, "proof report.launch.operations_released");
  } else {
    const blocker = requireKnownFields(report.blocker, new Set(["code", "message"]), "proof report.blocker");
    const code = requireSafeIdentifier(blocker.code, "proof report.blocker.code");
    if (!(code in EXTERNAL_BLOCKERS) || requireSafeText(blocker.message, "proof report.blocker.message") !== EXTERNAL_BLOCKERS[code]) {
      throw new Error("proof report blocker is not an approved fixed external blocker");
    }
  }
  const tally = validateReportTally(report.tally, implementedCommands);
  const recordsTally = validateProofRecords(expectedCommands, report.records);
  if (JSON.stringify(tally) !== JSON.stringify(recordsTally)) {
    throw new Error("proof report tally does not match its terminal records");
  }
  return tally;
}

/**
 * Queue every applicable operation before permitting any one to start.  The
 * caller owns the operation body so setup and case validation stay outside the
 * measured release.  This deliberately has no concurrency cap: provider
 * admission is part of the run's evidence, not a local policy to silently
 * serialize around.
 */
export async function executeBarrier(items, execute) {
  if (!Array.isArray(items)) throw new Error("barrier items must be an array");
  if (typeof execute !== "function") throw new Error("barrier executor must be a function");
  if (items.length === 0) return [];
  let ready = 0;
  let release;
  const gate = new Promise((resolve) => { release = resolve; });
  return Promise.all(items.map(async (item) => {
    ready += 1;
    if (ready === items.length) release();
    await gate;
    return execute(item);
  }));
}

function parseOptions(args) {
  const options = { _: [] };
  for (let index = 0; index < args.length; index += 1) {
    const argument = args[index];
    if (!argument.startsWith("--")) {
      options._.push(argument);
      continue;
    }
    const equals = argument.indexOf("=");
    const name = argument.slice(2, equals === -1 ? undefined : equals);
    if (!name) {
      throw new Error("empty option name");
    }
    if (equals !== -1) {
      options[name] = argument.slice(equals + 1);
      continue;
    }
    if (name === "self-test" || name === "execute-writes" || name === "external-blocker") {
      options[name] = true;
      continue;
    }
    const value = args[index + 1];
    if (!value || value.startsWith("--")) {
      throw new Error(`--${name} requires a value`);
    }
    options[name] = value;
    index += 1;
  }
  return options;
}

function requiredOption(options, name) {
  const value = String(options[name] || "").trim();
  if (!value) {
    throw new Error(`--${name} is required for a live GitHub proof run`);
  }
  return value;
}

async function loadJSON(file) {
  return JSON.parse(await readFile(file, "utf8"));
}

const CASE_REASON_FAMILIES = new Set([
  "mutation_outside_pinned_repo",
  "org_or_enterprise",
  "secret_material",
  "app_auth",
  "binary_resource",
  "no_cleanup_safe_fixture",
  "other",
]);

function validateCaseTally(value, label, { allowNegative = false } = {}) {
  const tally = requireKnownFields(value, new Set(["attemptable", "blocked", "families"]), label);
  const minimum = allowNegative ? Number.MIN_SAFE_INTEGER : 0;
  for (const key of ["attemptable", "blocked"]) {
    if (!Number.isSafeInteger(tally[key]) || tally[key] < minimum) {
      throw new Error(`${label}.${key} must be an integer`);
    }
  }
  const families = requireKnownFields(tally.families, CASE_REASON_FAMILIES, `${label}.families`);
  for (const family of CASE_REASON_FAMILIES) {
    if (!Number.isSafeInteger(families[family]) || families[family] < minimum) {
      throw new Error(`${label}.families must contain integer counts`);
    }
  }
}

function validatePersistedCaseFile(caseFile) {
  const source = requireKnownFields(caseFile, new Set([
    "schema_version",
    "connector",
    "test_repository",
    "context",
    "source",
    "cases",
    "measurement",
    "classification",
  ]), "live proof cases");
  if (!Number.isSafeInteger(source.schema_version) || source.schema_version < 1) {
    throw new Error("live proof cases.schema_version must be a positive integer");
  }
  requireSafeText(source.connector, "live proof cases.connector");
  requireSafeText(source.source, "live proof cases.source");
  const repository = requireKnownFields(source.test_repository, new Set([
    "owner",
    "repo",
    "owner_id",
    "repo_id",
    "organization_id",
  ]), "live proof cases.test_repository");
  for (const key of Object.keys(repository)) {
    requireSafeText(repository[key], `live proof cases.test_repository.${key}`);
  }
  const context = requireKnownFields(source.context, new Set([
    "test_owner",
    "test_repo",
    "test_repository",
  ]), "live proof cases.context");
  for (const key of Object.keys(context)) {
    requireSafeText(context[key], `live proof cases.context.${key}`);
  }
  if (!Array.isArray(source.cases)) {
    throw new Error("live proof cases.cases must be an array");
  }
  for (const item of source.cases) {
    const candidate = requireKnownFields(item, new Set([
      "command",
      "args",
      "readback",
      "untestable_reason",
    ]), "live proof case");
    requireSafeText(candidate.command, "live proof case.command");
    if (candidate.args !== undefined) {
      if (!Array.isArray(candidate.args)) throw new Error("live proof case.args must be an array");
      candidate.args.forEach((argument, index) => requireSafeText(argument, `live proof case.args[${index}]`));
    }
    if (candidate.untestable_reason !== undefined) {
      requireSafeText(candidate.untestable_reason, "live proof case.untestable_reason");
    }
    if (candidate.readback !== undefined) {
      const readback = requireKnownFields(candidate.readback, new Set(["command", "args"]), "live proof case.readback");
      requireSafeText(readback.command, "live proof case.readback.command");
      if (!Array.isArray(readback.args)) throw new Error("live proof case.readback.args must be an array");
      readback.args.forEach((argument, index) => requireSafeText(argument, `live proof case.readback.args[${index}]`));
    }
  }
  if (source.classification !== undefined) {
    validateCaseTally(source.classification, "live proof cases.classification");
  }
  if (source.measurement !== undefined) {
    const measurement = requireKnownFields(source.measurement, new Set(["baseline", "current", "movement"]), "live proof cases.measurement");
    validateCaseTally(measurement.baseline, "live proof cases.measurement.baseline");
    validateCaseTally(measurement.current, "live proof cases.measurement.current");
    validateCaseTally(measurement.movement, "live proof cases.measurement.movement", { allowNegative: true });
  }
}

function requireStringMap(value, name) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${name} must be an object`);
  }
  const out = {};
  for (const [key, item] of Object.entries(value)) {
    if (typeof item !== "string" && typeof item !== "number") {
      throw new Error(`${name}.${key} must be a scalar`);
    }
    if (isSensitiveConfigKey(key)) {
      throw new Error(`${name}.${key} may not carry a credential or grant`);
    }
    out[key] = String(item);
  }
  return out;
}

function parseCaseArguments(args, label) {
  if (!Array.isArray(args) || args.some((argument) => typeof argument !== "string")) {
    throw new Error(`${label} requires string args`);
  }
  const parsed = [];
  for (let index = 0; index < args.length; index += 1) {
    const argument = args[index];
    if (/\r|\n|\0/u.test(argument)) {
      throw new Error(`${label} may not contain control characters`);
    }
    const flag = splitLongFlag(argument);
    if (!flag || !/^[a-z][a-z0-9_-]*$/iu.test(flag.rawName)) {
      throw new Error(`${label} must use named long flags`);
    }
    if (!flag.hasValue) {
      const value = args[index + 1];
      if (typeof value === "string" && !value.startsWith("--")) {
        if (/\r|\n|\0/u.test(value)) {
          throw new Error(`${label} may not contain control characters`);
        }
        parsed.push({ ...flag, hasValue: true, value });
        index += 1;
        continue;
      }
      parsed.push(flag);
      continue;
    }
    parsed.push(flag);
  }
  return parsed;
}

function validateConfigOverride(value, label) {
  const separator = String(value || "").indexOf("=");
  if (separator < 1) {
    throw new Error(`${label} --config requires key=value`);
  }
  const key = String(value).slice(0, separator).trim();
  if (isSensitiveConfigKey(key)) {
    throw new Error(`${label} may not carry a secret --config key`);
  }
  throw new Error(`${label} may not override connector configuration`);
}

function validateFlagValue(flag, value, label) {
  if (value === "") {
    throw new Error(`${label} requires a non-empty flag value`);
  }
  if (flag.type === "boolean" && value !== "true" && value !== "false") {
    throw new Error(`${label} requires a boolean flag value`);
  }
  if (flag.type === "integer" && !/^-?\d+$/u.test(value)) {
    throw new Error(`${label} requires an integer flag value`);
  }
  if (flag.type === "enum" && (!Array.isArray(flag.values) || !flag.values.includes(value))) {
    throw new Error(`${label} requires a declared enum flag value`);
  }
  if (flag.type === "json") {
    try {
      JSON.parse(value);
    } catch {
      throw new Error(`${label} requires JSON flag value`);
    }
  }
}

function validateTypedCaseArguments(args, surfaceCommand, label) {
  const declarations = new Map(
    (surfaceCommand.flags || []).map((flag) => [normalizeArgumentName(flag.name), flag]),
  );
  for (const argument of parseCaseArguments(args, label)) {
    if (LIFECYCLE_ARGUMENT_NAMES.has(argument.name)) {
      throw new Error(`${label} may not override lifecycle or credential flags`);
    }
    if (argument.name === "config") {
      validateConfigOverride(argument.value, label);
    }
    const declaration = declarations.get(argument.name);
    if (!declaration) {
      throw new Error(`${label} uses a flag not declared by the GitHub command surface`);
    }
    if (!argument.hasValue) {
      if (declaration.type !== "boolean") {
        throw new Error(`${label} requires a value for --${declaration.name}`);
      }
      continue;
    }
    validateFlagValue(declaration, argument.value, `${label} --${declaration.name}`);
  }
}

function validateSourceDerivedArguments(args, expectedArgs, surfaceCommand, label) {
  validateTypedCaseArguments(args, surfaceCommand, label);
  if (JSON.stringify(args) !== JSON.stringify(expectedArgs)) {
    throw new Error(`${label} must exactly match the source-derived command descriptor and run-owned boundary`);
  }
}

function validateReadback(readback, command, expected, surfaceCommands) {
  if (!readback || typeof readback !== "object" || Array.isArray(readback)) {
    throw new Error(`readback for ${JSON.stringify(command)} must be an object`);
  }
  if (Object.keys(readback).some((key) => key !== "command" && key !== "args")) {
    throw new Error(`readback for ${JSON.stringify(command)} has unsupported fields`);
  }
  const readbackCommand = String(readback.command || "").trim();
  const readbackSurfaceCommand = surfaceCommands.get(readbackCommand);
  if (!readbackSurfaceCommand || !expected.includes(readbackCommand)) {
    throw new Error(`readback for ${JSON.stringify(command)} names an unknown implemented command`);
  }
  if (readbackSurfaceCommand.intent === "reverse_etl" || readbackSurfaceCommand.intent === "direct_write") {
    throw new Error(`readback for ${JSON.stringify(command)} may not invoke another write`);
  }
  validateTypedCaseArguments(readback.args, readbackSurfaceCommand, `readback for ${JSON.stringify(command)}`);
}

function validateCaseFile(caseFile, boundary, surface) {
  validatePersistedCaseFile(caseFile);
  const expected = enumerateImplementedCommands(surface);
  const canonical = buildCases(surface, boundary);
  const surfaceCommands = new Map(surface.commands.map((command) => [command.path, command]));
  if (!caseFile || caseFile.connector !== CONNECTOR) {
    throw new Error("live proof cases must be explicitly GitHub-only");
  }
  if (caseFile.schema_version !== canonical.schema_version || caseFile.source !== canonical.source) {
    throw new Error("live proof cases must be generated from the current GitHub command surface");
  }
  const repository = caseFile.test_repository;
  if (
    !repository ||
    repository.owner !== boundary.owner ||
    repository.repo !== boundary.repo ||
    repository.owner_id !== boundary.owner_id ||
    repository.repo_id !== boundary.repo_id ||
    repository.organization_id !== boundary.organization_id
  ) {
    throw new Error("case file test_repository must exactly match the immutable run-owned boundary");
  }
  if (!Array.isArray(caseFile.cases)) {
    throw new Error("case file must contain a cases array");
  }
  const context = requireStringMap(caseFile.context || {}, "context");
  if (JSON.stringify(context) !== JSON.stringify(canonical.context)) {
    throw new Error("case file context must exactly match the immutable run-owned boundary");
  }
  const canonicalByCommand = new Map(canonical.cases.map((item) => [item.command, item]));
  const commands = new Map();
  for (const item of caseFile.cases) {
    if (!item || typeof item !== "object" || Array.isArray(item)) {
      throw new Error("each live proof case must be an object");
    }
    const command = String(item.command || "").trim();
    if (!expected.includes(command)) {
      throw new Error(`case names unimplemented or unknown command ${JSON.stringify(command)}`);
    }
    if (commands.has(command)) {
      throw new Error(`duplicate live proof case for ${JSON.stringify(command)}`);
    }
    if (Object.keys(item).some((key) => !["command", "args", "readback", "untestable_reason"].includes(key))) {
      throw new Error(`case ${JSON.stringify(command)} has unsupported fields`);
    }
    const sourceCase = canonicalByCommand.get(command);
    const surfaceCommand = surfaceCommands.get(command);
    if (!sourceCase || !surfaceCommand) {
      throw new Error(`case ${JSON.stringify(command)} is missing its source command descriptor`);
    }
    if (item.readback !== undefined) {
      validateReadback(item.readback, command, expected, surfaceCommands);
      throw new Error(`case ${JSON.stringify(command)} does not declare a typed read-back lifecycle`);
    }
    if (sourceCase.untestable_reason !== undefined) {
      if (item.args !== undefined) {
        validateTypedCaseArguments(item.args, surfaceCommand, `case ${JSON.stringify(command)}`);
        throw new Error(`case ${JSON.stringify(command)} remains untestable without a typed fixture lifecycle`);
      }
      if (item.untestable_reason !== sourceCase.untestable_reason) {
        throw new Error(`case ${JSON.stringify(command)} must retain its source-derived untestable reason`);
      }
    } else {
      if (item.untestable_reason !== undefined) {
        throw new Error(`case ${JSON.stringify(command)} may not replace an executable source-derived case with an untestable reason`);
      }
      validateSourceDerivedArguments(
        item.args,
        sourceCase.args,
        surfaceCommand,
        `case ${JSON.stringify(command)}`,
      );
    }
    commands.set(command, item);
  }
  if (commands.size !== expected.length) {
    const missing = expected.find((command) => !commands.has(command));
    throw new Error(`case file is incomplete; missing case for ${JSON.stringify(missing)}`);
  }
  return {
    context: canonical.context,
    cases: commands,
  };
}

function interpolateArguments(args, context) {
  return args.map((argument) =>
    argument.replace(/\{\{([a-zA-Z0-9_]+)\}\}/g, (whole, key) => {
      if (!(key in context)) {
        throw new Error(`live proof context is missing ${key}`);
      }
      return context[key];
    }),
  );
}

function shellSafeInvocation(command, args, credential, root) {
  return [
    "pm",
    CONNECTOR,
    ...command.split(" "),
    ...args,
    "--credential",
    credential,
    "--root",
    root,
    "--json",
  ];
}

export function runProcess(binary, args, cwd, { timeoutMs = PROCESS_TIMEOUT_MS } = {}) {
  if (!Number.isInteger(timeoutMs) || timeoutMs <= 0) {
    throw new Error("process timeout must be a positive integer");
  }
  return new Promise((resolve, reject) => {
    const child = spawn(binary, args, { cwd, stdio: ["ignore", "pipe", "pipe"] });
    let stdout = "";
    let stderr = "";
    let bytes = 0;
    let overflow = false;
    let timedOut = false;
    let settled = false;
    let forceKill;
    const timer = setTimeout(() => {
      timedOut = true;
      child.kill("SIGTERM");
      forceKill = setTimeout(() => child.kill("SIGKILL"), PROCESS_KILL_GRACE_MS);
    }, timeoutMs);
    const finish = (callback) => {
      if (settled) {
        return;
      }
      settled = true;
      clearTimeout(timer);
      clearTimeout(forceKill);
      callback();
    };
    const consume = (target, chunk) => {
      bytes += chunk.length;
      if (bytes > OUTPUT_LIMIT_BYTES) {
        overflow = true;
        child.kill("SIGTERM");
        return;
      }
      if (target === "stdout") {
        stdout += chunk.toString("utf8");
      } else {
        stderr += chunk.toString("utf8");
      }
    };
    child.stdout.on("data", (chunk) => consume("stdout", chunk));
    child.stderr.on("data", (chunk) => consume("stderr", chunk));
    child.once("error", (error) => finish(() => reject(error)));
    child.once("close", (code, signal) =>
      finish(() => resolve({ code, signal, stdout, stderr, overflow, timedOut, timeoutMs })),
    );
  });
}

function httpStatusFromFailure(text) {
  const match = /\b(?:http|status)\s+(\d{3})\b/i.exec(text);
  return match ? Number.parseInt(match[1], 10) : undefined;
}

function parseJSONOutput(result, step) {
  if (result.overflow) {
    throw new Error(`${step} exceeded the bounded in-memory output limit`);
  }
  if (result.timedOut) {
    throw new Error(`${step} exceeded the ${result.timeoutMs} ms terminal bound`);
  }
  if (result.code !== 0) {
    const status = httpStatusFromFailure(`${result.stdout}\n${result.stderr}`);
    const statusDescription = status === undefined
      ? "without a provider HTTP status"
      : `with provider HTTP status ${status}`;
    const error = new Error(`${step} exited ${result.code ?? "without an exit code"} ${statusDescription}`);
    error.httpStatus = status;
    throw error;
  }
  try {
    return JSON.parse(result.stdout);
  } catch {
    throw new Error(`${step} did not produce machine-readable JSON`);
  }
}

export function assertReadEnvelope(envelope, command) {
  if (!envelope || envelope.connector !== CONNECTOR || envelope.command !== command) {
    throw new Error("pm response does not identify the requested GitHub command");
  }
  if (envelope.kind === "ConnectorCommandDirectRead") {
    if (!Number.isInteger(envelope.status)) {
      throw new Error("direct-read response does not expose provider HTTP status");
    }
    if (!("response" in envelope)) {
      throw new Error("direct-read response omitted returned data");
    }
    return {
      httpStatus: envelope.status,
      assertion: { kind: "direct-read-response", subject: command, matched: true },
    };
  }
  if (envelope.kind === "ConnectorCommandRead") {
    if (!Number.isInteger(envelope.count) || !Array.isArray(envelope.records)) {
      throw new Error("stream response omitted returned record accounting");
    }
    return {
      assertion: {
        kind: "stream-records",
        subject: command,
        matched: true,
        count: envelope.count,
      },
    };
  }
  if (envelope.kind === "ConnectorCommandBinaryDownload") {
    if (!envelope.record || typeof envelope.record !== "object" || Array.isArray(envelope.record)) {
      throw new Error("binary-download response omitted returned file accounting");
    }
    return {
      assertion: { kind: "binary-download-record", subject: command, matched: true },
    };
  }
  throw new Error(`unexpected non-write result kind ${JSON.stringify(envelope.kind || "")}`);
}

async function runWriteLifecycle({ binary, root, credential, command, args, readback, owner, repo, cwd }) {
  const planArgs = [CONNECTOR, ...command.split(" "), ...args, "--credential", credential, "--root", root, "--json"];
  const humanPlanArgs = planArgs.filter((argument) => argument !== "--json");
  const planResult = await runProcess(binary, humanPlanArgs, cwd);
  if (planResult.overflow || planResult.code !== 0) {
    throw new Error("write plan did not complete");
  }
  const planID = /Created connector command plan\s+(\S+)/.exec(planResult.stdout)?.[1] || "";
  if (!planID) {
    throw new Error("write plan response omitted plan identity");
  }
  const initialGrant = /Approval token:\s*(\S+)/.exec(planResult.stdout)?.[1] || "";
  const challenge = /Confirmation required:\s+--confirm\s+(\S+)/.exec(planResult.stdout)?.[1] || "";
  const previewArgs = [CONNECTOR, ...command.split(" "), "--plan", planID, "--preview", "--root", root];
  const preview = await runProcess(binary, previewArgs, cwd);
  if (preview.overflow || preview.code !== 0) {
    throw new Error("write preview did not complete");
  }
  const grant = /Approval token:\s*(\S+)/.exec(preview.stdout)?.[1] || initialGrant;
  if (!grant) {
    throw new Error("write preview omitted the single-use approval grant");
  }
  const executeArgs = [
    CONNECTOR,
    ...command.split(" "),
    "--plan",
    planID,
    "--approve",
    grant,
    ...(challenge ? ["--confirm", challenge] : []),
    "--root",
    root,
    "--json",
  ];
  const runEnvelope = parseJSONOutput(await runProcess(binary, executeArgs, cwd), "write execution");
  const run = runEnvelope?.run;
  if (runEnvelope?.kind !== "ReverseRun" || run?.status !== "completed" || run?.records_succeeded !== 1 || run?.records_failed !== 0) {
    throw new Error("write execution did not report one completed provider mutation");
  }
  const operation = run.operation_direct_write;
  const readbackArgs = interpolateArguments(readback.args, {
    test_owner: owner,
    test_repo: repo,
    test_repository: `${owner}/${repo}`,
  });
  const readbackEnvelope = parseJSONOutput(
    await runProcess(
      binary,
      [CONNECTOR, ...readback.command.split(" "), ...readbackArgs, "--credential", credential, "--root", root, "--json"],
      cwd,
    ),
    "write readback",
  );
  const readbackResult = assertReadEnvelope(readbackEnvelope, readback.command);
  return {
    httpStatus: Number.isInteger(operation?.status) ? operation.status : undefined,
    assertion: {
      kind: "reverse-write-readback",
      subject: command,
      matched: true,
      readback_kind: readbackResult.assertion.kind,
    },
  };
}

function installationPreflightCommand(surface) {
  const matches = Array.isArray(surface?.commands)
    ? surface.commands.filter((command) => command?.path === APP_INSTALLATION_REPOSITORY_PREFLIGHT)
    : [];
  if (matches.length !== 1) {
    throw new Error("GitHub surface does not declare exactly one App installation repository preflight");
  }
  const command = matches[0];
  const apis = Array.isArray(command.api_surface) ? command.api_surface : [];
  if (command.availability !== "implemented" || command.intent !== "direct_read" ||
      !Array.isArray(command.flags) || command.flags.length !== 0 || apis.length !== 1 ||
      apis[0]?.method !== "GET" || apis[0]?.path !== "/installation/repositories") {
    throw new Error("GitHub App installation repository preflight descriptor is not immutable and targetless");
  }
  return command;
}

function normalizeGitHubBaseOrigin(value) {
  if (value === undefined) return CANONICAL_GITHUB_API_ORIGIN;
  const raw = requireSafeText(value, "GitHub credential base_url");
  let parsed;
  try {
    parsed = new URL(raw);
  } catch {
    throw new Error("GitHub credential must use the canonical GitHub API origin");
  }
  if (parsed.protocol !== "https:" || parsed.origin !== CANONICAL_GITHUB_API_ORIGIN ||
      parsed.username !== "" || parsed.password !== "" || parsed.pathname !== "/" ||
      parsed.search !== "" || parsed.hash !== "") {
    throw new Error("GitHub credential must use the canonical GitHub API origin");
  }
  return CANONICAL_GITHUB_API_ORIGIN;
}

function validateAppCredentialMetadata(metadata, boundary) {
  if (!isPlainObject(metadata) || requireSafeText(metadata.connector, "credential connector") !== CONNECTOR) {
    throw new Error("named credential is not a GitHub credential");
  }
  const config = requireKnownFields(metadata.config, new Set([
    "owner",
    "repo",
    "base_url",
    "auth_type",
    "app_id",
    "installation_id",
  ]), "GitHub App credential configuration");
  const configuredOwner = requireGitHubSlug(config.owner, "GitHub credential owner");
  const configuredRepo = requireGitHubSlug(config.repo, "GitHub credential repository");
  if (configuredOwner !== boundary.owner || configuredRepo !== boundary.repo) {
    throw new Error("named credential is not scoped to the immutable Polymetrics-Cert boundary");
  }
  normalizeGitHubBaseOrigin(config.base_url);
  if (requireSafeText(config.auth_type, "GitHub credential auth_type") !== "github_app") {
    throw new Error("live proof requires a GitHub App installation credential");
  }
  requireSafeIdentifier(config.app_id, "GitHub App credential app_id");
  requireSafeIdentifier(config.installation_id, "GitHub App credential installation_id");
  if (!Array.isArray(metadata.secret_fields) || metadata.secret_fields.length !== 1) {
    throw new Error("live proof requires exactly one GitHub App private-key secret field");
  }
  const secretField = requireSafeIdentifier(metadata.secret_fields[0], "GitHub App credential secret field");
  if (secretField !== "private_key" && secretField !== "private_key_base64") {
    throw new Error("live proof requires a GitHub App private-key secret field");
  }
}

function normalizeInstallationRepository(record) {
  if (!isPlainObject(record)) {
    throw new Error("App installation repository preflight returned a malformed repository record");
  }
  const id = requireProviderID(record.id, "App installation repository ID");
  const fullName = requireSafeText(record.full_name, "App installation repository full_name");
  const parts = fullName.split("/");
  if (parts.length !== 2) {
    throw new Error("App installation repository preflight returned a malformed repository identity");
  }
  if (!isPlainObject(record.owner)) {
    throw new Error("App installation repository preflight returned a malformed owner identity");
  }
  const owner = record.owner;
  const ownerLogin = requireGitHubSlug(owner.login, "App installation repository owner.login");
  const ownerID = requireProviderID(owner.id, "App installation repository owner.id");
  const fullNameOwner = requireGitHubSlug(parts[0], "App installation repository full_name owner");
  const fullNameRepo = requireGitHubSlug(parts[1], "App installation repository full_name repository");
  if (fullNameOwner !== ownerLogin) {
    throw new Error("App installation repository preflight returned inconsistent owner identity");
  }
  return {
    id,
    full_name: `${fullNameOwner}/${fullNameRepo}`,
    owner: { login: ownerLogin, id: ownerID },
  };
}

function validateInstallationPreflightPage(envelope) {
  if (!isPlainObject(envelope) || envelope.kind !== "ConnectorCommandDirectRead" ||
      envelope.connector !== CONNECTOR || envelope.command !== APP_INSTALLATION_REPOSITORY_PREFLIGHT ||
      !Number.isSafeInteger(envelope.status) || envelope.status < 200 || envelope.status >= 300) {
    throw new Error("App installation repository preflight did not return the expected GitHub direct-read envelope");
  }
  const response = requireKnownFields(envelope.response, new Set(["total_count", "repositories"]), "App installation repository response");
  if (!Number.isSafeInteger(response.total_count) || response.total_count < 0 || !Array.isArray(response.repositories)) {
    throw new Error("App installation repository preflight returned malformed pagination data");
  }
  const page = requireKnownFields(envelope.page, new Set([
    "strategy",
    "records",
    "size",
    "number",
    "has_more",
    "next_number",
    "next_cursor",
    "complete",
    "reason",
  ]), "App installation repository page");
  if (!Number.isSafeInteger(page.records) || page.records !== response.repositories.length ||
      typeof page.complete !== "boolean" || typeof page.has_more !== "boolean") {
    throw new Error("App installation repository preflight page is incomplete or malformed");
  }
  return {
    repositories: response.repositories.map(normalizeInstallationRepository),
    totalCount: response.total_count,
    page,
    status: envelope.status,
  };
}

function nextInstallationPreflightPage(page) {
  if (page.complete) {
    if (page.has_more || page.next_number !== undefined || page.next_cursor !== undefined) {
      throw new Error("App installation repository preflight returned contradictory completion data");
    }
    return null;
  }
  if (page.has_more !== true) {
    throw new Error("App installation repository preflight cannot prove the full installation repository set");
  }
  if (page.next_number !== undefined) {
    if (!Number.isSafeInteger(page.next_number) || page.next_number < 1 || page.next_cursor !== undefined) {
      throw new Error("App installation repository preflight returned malformed next-page data");
    }
    return ["--page", String(page.next_number)];
  }
  if (page.next_cursor !== undefined) {
    return ["--page-cursor", requireSafeText(page.next_cursor, "App installation repository next cursor")];
  }
  throw new Error("App installation repository preflight cannot safely follow the next page");
}

export async function validateCredentialScope({ binary, root, credential, boundary, surface, cwd }) {
  const preflight = installationPreflightCommand(surface);
  const result = await runProcess(binary, ["credentials", "inspect", credential, "--root", root, "--json"], cwd);
  const envelope = parseJSONOutput(result, "credential inspection");
  if (envelope?.kind !== "Credential") {
    throw new Error("named credential is not a GitHub credential");
  }
  validateAppCredentialMetadata(envelope.credential, boundary);

  const repositories = new Map();
  const seenPages = new Set();
  let navigation = [];
  let totalCount;
  let pages = 0;
  let status;
  while (navigation !== null) {
    if (pages >= MAX_INSTALLATION_PREFLIGHT_PAGES) {
      throw new Error("App installation repository preflight exceeded its bounded pagination limit");
    }
    const pageKey = navigation.length === 0 ? "initial" : `${navigation[0]}:${navigation[1]}`;
    if (seenPages.has(pageKey)) {
      throw new Error("App installation repository preflight repeated a pagination position");
    }
    seenPages.add(pageKey);
    const pageEnvelope = parseJSONOutput(
      await runProcess(
        binary,
        [CONNECTOR, ...preflight.path.split(" "), ...navigation, "--credential", credential, "--root", root, "--json"],
        cwd,
      ),
      "App installation repository preflight",
    );
    const page = validateInstallationPreflightPage(pageEnvelope);
    pages += 1;
    status = page.status;
    if (totalCount === undefined) {
      totalCount = page.totalCount;
    } else if (totalCount !== page.totalCount) {
      throw new Error("App installation repository preflight returned inconsistent total counts");
    }
    for (const repository of page.repositories) {
      if (repositories.has(repository.id)) {
        throw new Error("App installation repository preflight returned a duplicate immutable repository ID");
      }
      repositories.set(repository.id, repository);
    }
    navigation = nextInstallationPreflightPage(page.page);
  }
  if (repositories.size !== totalCount) {
    throw new Error("App installation repository preflight did not return every installation repository");
  }
  const expectedName = `${boundary.owner}/${boundary.repo}`;
  const matches = [...repositories.values()].filter((repository) =>
    repository.id === boundary.repo_id &&
    repository.full_name === expectedName &&
    repository.owner.login === boundary.owner &&
    repository.owner.id === boundary.owner_id,
  );
  if (matches.length !== 1) {
    throw new Error("App installation repository preflight did not prove the immutable Polymetrics-Cert boundary");
  }
  return {
    record: redactForReport({
      command: preflight.path,
      state: "proven",
      http_status: status,
      assertion: {
        kind: "app-installation-repository-boundary",
        subject: preflight.path,
        matched: true,
        pages,
      },
    }),
  };
}

async function executeLive(options) {
  if (options.connector !== undefined) {
    throw new Error("this runner is GitHub-only; --connector is not accepted");
  }
  if (options._.length !== 0) {
    throw new Error("live proof runner accepts options only");
  }
  const binary = requireSafeText(requiredOption(options, "pm"), "live proof pm binary");
  const root = requireSafeText(requiredOption(options, "root"), "live proof root");
  const credential = requireSafeIdentifier(requiredOption(options, "credential"), "live proof credential");
  const boundaryPath = requireSafeText(requiredOption(options, "boundary"), "live proof boundary path");
  const boundary = resolveLiveBoundary(await loadJSON(boundaryPath));
  const suppliedOwner = options["test-owner"] === undefined
    ? ""
    : requireGitHubSlug(options["test-owner"], "live proof test owner");
  const suppliedRepo = options["test-repo"] === undefined
    ? ""
    : requireGitHubSlug(options["test-repo"], "live proof test repository");
  if ((suppliedOwner && suppliedOwner !== boundary.owner) || (suppliedRepo && suppliedRepo !== boundary.repo)) {
    throw new Error("live proof --test-owner/--test-repo must match the immutable run-owned boundary");
  }
  const owner = boundary.owner;
  const repo = boundary.repo;
  const casesPath = requireSafeText(requiredOption(options, "cases"), "live proof cases path");
  const reportPath = requireSafeText(requiredOption(options, "report"), "live proof report path");
  const surface = await loadJSON(SURFACE_PATH);
  const binaryBytes = await readFile(binary);
  const caseBytes = await readFile(casesPath);
  const expected = enumerateImplementedCommands(surface);
  const surfaceCommands = new Map(surface.commands.map((command) => [command.path, command]));
  const cases = validateCaseFile(await loadJSON(casesPath), boundary, surface);
  const credentialScope = await validateCredentialScope({
    binary,
    root,
    credential,
    boundary,
    surface,
    cwd: REPOSITORY_ROOT,
  });

  const records = [credentialScope.record];
  const operations = [];
  for (const command of expected) {
    const caseItem = cases.cases.get(command);
    if (command === APP_INSTALLATION_REPOSITORY_PREFLIGHT) {
      if (caseItem.untestable_reason) {
        throw new Error("App installation repository preflight may not be marked untestable");
      }
      continue;
    }
    if (caseItem.untestable_reason) {
      records.push(redactForReport({ command, state: "untestable", reason: caseItem.untestable_reason }));
      continue;
    }
    const args = interpolateArguments(caseItem.args, {
      ...cases.context,
      test_owner: owner,
      test_repo: repo,
      test_repository: `${owner}/${repo}`,
    });
    const surfaceCommand = surfaceCommands.get(command);
    const isWrite = surfaceCommand?.intent === "reverse_etl" || surfaceCommand?.intent === "direct_write";
    if (isWrite && !options["execute-writes"]) {
      throw new Error("--execute-writes is required before this runner will dispatch any GitHub mutation");
    }
    operations.push({ command, args, isWrite, readback: caseItem.readback });
  }

  const attempted = await executeBarrier(operations, async ({ command, args, isWrite, readback }) => {
    const invocation = shellSafeInvocation(command, args, credential, root);
    try {
      const result = isWrite
          ? await runWriteLifecycle({
              binary,
              root,
              credential,
              command,
              args,
              readback,
              owner,
              repo,
              cwd: REPOSITORY_ROOT,
            })
          : assertReadEnvelope(
              parseJSONOutput(
                await runProcess(
                  binary,
                  [CONNECTOR, ...command.split(" "), ...args, "--credential", credential, "--root", root, "--json"],
                  REPOSITORY_ROOT,
                ),
                "read execution",
              ),
              command,
            );
      return redactForReport({
        command,
        state: "proven",
        invocation,
        http_status: result.httpStatus,
        assertion: result.assertion,
      });
    } catch (error) {
      const diagnostic = error instanceof Error ? error.message : "live execution failed without a safe diagnostic";
      return redactForReport({
        command,
        state: "failed",
        invocation,
        http_status: error?.httpStatus,
        reason: concreteReason(diagnostic)
          ? diagnostic
          : `${command} live execution failed without a safe provider diagnostic`,
      });
    }
  });
  records.push(...attempted);

  const tally = validateProofRecords(expected, records);
  const report = {
    schema_version: 1,
    connector: CONNECTOR,
    status: "credentialed_live",
    generated_at: new Date().toISOString(),
    surface_sha256: createHash("sha256").update(JSON.stringify(surface)).digest("hex"),
    binary_sha256: createHash("sha256").update(binaryBytes).digest("hex"),
    case_file_sha256: createHash("sha256").update(caseBytes).digest("hex"),
    test_repository: "<run-owned-boundary-repository>",
    run_boundary: {
      run_id: boundary.run_id,
      owner: boundary.owner,
      repo: boundary.repo,
      owner_id: boundary.owner_id,
      repo_id: boundary.repo_id,
      organization_id: boundary.organization_id,
    },
    launch: {
      strategy: "single_barrier_release",
      operations_released: operations.length,
    },
    implemented_commands: expected.length,
    tally,
    records,
  };
  validateProofReport(report, expected);
  await writeFile(reportPath, `${JSON.stringify(report, null, 2)}\n`, { mode: 0o600 });
  await chmod(reportPath, 0o600);
  process.stdout.write(`${CONNECTOR} live proof: proven=${tally.proven} untestable=${tally.untestable} failed=${tally.failed}\n`);
  return tally.failed === 0 ? 0 : 1;
}

async function recordExternalBlocker(options) {
  const allowed = new Set(["_", "external-blocker", "pm", "report", "reason"]);
  for (const key of Object.keys(options)) {
    if (!allowed.has(key)) {
      throw new Error(`external-blocker mode does not accept --${key}`);
    }
  }
  if (options._.length !== 0) {
    throw new Error("external-blocker mode accepts options only");
  }
  const binary = requireSafeText(requiredOption(options, "pm"), "external blocker pm binary");
  const reportPath = requireSafeText(requiredOption(options, "report"), "external blocker report path");
  const blockerCode = requireSafeIdentifier(requiredOption(options, "reason"), "external blocker code");
  const blockerMessage = EXTERNAL_BLOCKERS[blockerCode];
  if (!blockerMessage) {
    throw new Error("external blocker code is not permitted");
  }
  const surface = await loadJSON(SURFACE_PATH);
  const expected = enumerateImplementedCommands(surface);
  const records = expected.map((command) =>
    redactForReport({ command, state: "untestable", reason: blockerMessage }),
  );
  const tally = validateProofRecords(expected, records);
  const binaryBytes = await readFile(binary);
  const report = {
    schema_version: 1,
    connector: CONNECTOR,
    generated_at: new Date().toISOString(),
    status: "external_blocker",
    surface_sha256: createHash("sha256").update(JSON.stringify(surface)).digest("hex"),
    binary_sha256: createHash("sha256").update(binaryBytes).digest("hex"),
    test_repository: "<credentialed-live-proof-not-available>",
    implemented_commands: expected.length,
    blocker: {
      code: blockerCode,
      message: blockerMessage,
    },
    tally,
    records,
  };
  validateProofReport(report, expected);
  await writeFile(reportPath, `${JSON.stringify(report, null, 2)}\n`, { mode: 0o600 });
  await chmod(reportPath, 0o600);
  process.stdout.write(`${CONNECTOR} live proof: external blocker; untestable=${tally.untestable}\n`);
  return 0;
}

async function selfTest() {
  const commands = enumerateImplementedCommands({
    commands: [
      { path: "repo view", availability: "implemented" },
      { path: "repo delete", availability: "blocked" },
      { path: "issue list", availability: "implemented" },
    ],
  });
  const records = [
    redactForReport({
      command: "issue list",
      state: "proven",
      invocation: ["pm", "github", "issue", "list", "--token", "ghp_fixture_secret"],
      http_status: 200,
      assertion: { kind: "returned-data", subject: "issues", matched: true },
    }),
    redactForReport({
      command: "repo view",
      state: "untestable",
      reason: "requires GitHub Enterprise Server administration, which this dedicated test token cannot hold",
    }),
  ];
  const tally = validateProofRecords(commands, records);
  if (tally.proven !== 1 || tally.untestable !== 1 || tally.failed !== 0) {
    throw new Error("self-test tally does not match the fixture");
  }
  if (JSON.stringify(records).includes("ghp_fixture_secret")) {
    throw new Error("self-test detected fixture credential leakage");
  }
  process.stdout.write("github live proof self-test: ok\n");
}

async function main() {
  const options = parseOptions(process.argv.slice(2));
  if (options["external-blocker"]) {
    return recordExternalBlocker(options);
  }
  if (options["self-test"]) {
    if (Object.keys(options).some((key) => !["_", "self-test"].includes(key)) || options._.length !== 0) {
      throw new Error("--self-test does not accept live-run options");
    }
    await selfTest();
    return 0;
  }
  return executeLive(options);
}

if (path.resolve(process.argv[1] || "") === fileURLToPath(import.meta.url)) {
  main()
    .then((exitCode) => {
      process.exitCode = exitCode;
    })
    .catch((error) => {
      process.stderr.write(`github live proof: ${redactText(error instanceof Error ? error.message : "failed")}\n`);
      process.exitCode = 2;
    });
}
