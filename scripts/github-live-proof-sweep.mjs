#!/usr/bin/env node

import { createHash } from "node:crypto";
import { readFile, writeFile, chmod } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";
import { spawn } from "node:child_process";

import {
  CASE_REASON_FAMILIES,
  HISTORICAL_TERMINAL_MEASUREMENT,
  resolveLiveBoundary,
  validateCanonicalCaseArtifact,
} from "./github-live-cases.mjs";
import {
  assertPersistedArtifactSafe,
  assertSafePersistedScalar,
  redactPersistedText,
  stableJSONString,
} from "./github-live-artifact-guard.mjs";

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
const OUTPUT_LIMIT_BYTES = 2 * 1024 * 1024;
const PROCESS_TIMEOUT_MS = 45_000;
const PROCESS_KILL_GRACE_MS = 1_000;
const APP_INSTALLATION_REPOSITORY_PREFLIGHT = "apps list-repos-accessible-to-installation";
const CANONICAL_GITHUB_API_ORIGIN = "https://api.github.com";
const BUILT_PM_IN_PROCESS = "built_pm_in_process";
const EXTERNAL_PM_PER_OPERATION = "external_pm_per_operation";
const EXECUTION_MODELS = new Set([
  BUILT_PM_IN_PROCESS,
  EXTERNAL_PM_PER_OPERATION,
]);
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
const RESERVED_EXECUTION_FLAGS = ["--credential", "--connection", "--root", "--approve", "--approval-token-stdin", "--plan", "--confirm"];
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
  return redactPersistedText(value);
}

function redactConfigValue(value) {
  const raw = String(value ?? "");
  const separator = raw.indexOf("=");
  if (separator < 1) return "<redacted>";
  const key = raw.slice(0, separator);
  return `${redactText(key)}=<redacted>`;
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
  assertSafePersistedScalar(value, label);
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

function isReservedExecutionFlag(argument) {
  return RESERVED_EXECUTION_FLAGS.some((flag) => argument === flag || argument.startsWith(`${flag}=`));
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
    assertPersistedArtifactSafe(record, "proof record");
    const { command, state } = validateProofRecordShape(record);
    if (reportContainsForbiddenField(record)) {
      throw new Error("proof record contains raw execution data");
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

function requireExecutionModel(value, label) {
  const executionModel = requireSafeText(value, label);
  if (!EXECUTION_MODELS.has(executionModel)) {
    throw new Error(`${label} is not a supported execution model`);
  }
  return executionModel;
}

function assertCurrentCertificationExecutionModel(executionModel) {
  if (executionModel !== BUILT_PM_IN_PROCESS) {
    throw new Error("credentialed live evidence requires the built_pm_in_process execution model");
  }
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

export function validateProofReport(candidate, expectedCommands, { caseDigest } = {}) {
  if (!isPlainObject(candidate)) {
    throw new Error("proof report must be an object");
  }
  assertPersistedArtifactSafe(candidate, "proof report");
  const status = requireSafeText(candidate.status, "proof report.status");
  const fields = status === "credentialed_live"
    ? new Set([
      "schema_version",
      "connector",
      "status",
      "generated_at",
      "surface_sha256",
      "binary_sha256",
      "case_digest",
      "execution_model",
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
        "execution_model",
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
  if (report.schema_version !== 2 || requireSafeText(report.connector, "proof report.connector") !== CONNECTOR) {
    throw new Error("proof report must describe the GitHub proof schema");
  }
  const generatedAt = requireSafeText(report.generated_at, "proof report.generated_at");
  if (!/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$/u.test(generatedAt) || Number.isNaN(Date.parse(generatedAt))) {
    throw new Error("proof report.generated_at must be an ISO timestamp");
  }
  requireSHA256(report.surface_sha256, "proof report.surface_sha256");
  requireSHA256(report.binary_sha256, "proof report.binary_sha256");
  const executionModel = requireExecutionModel(report.execution_model, "proof report.execution_model");
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
    assertCurrentCertificationExecutionModel(executionModel);
    const digest = requireSHA256(report.case_digest, "proof report.case_digest");
    if (caseDigest !== undefined && digest !== caseDigest) {
      throw new Error("proof report case digest does not match the canonical case artifact");
    }
    validateReportBoundary(report.run_boundary);
    const launch = requireKnownFields(report.launch, new Set(["strategy", "operations_released"]), "proof report.launch");
    if (requireSafeText(launch.strategy, "proof report.launch.strategy") !== "single_barrier_release") {
      throw new Error("proof report launch strategy is not the required barrier release");
    }
    requireNonNegativeInteger(launch.operations_released, "proof report.launch.operations_released");
  } else {
    if (executionModel !== EXTERNAL_PM_PER_OPERATION) {
      throw new Error("external blocker evidence requires the external_pm_per_operation execution model");
    }
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

function validateCaseClassification(value) {
  const classification = requireKnownFields(value, new Set([
    "total",
    "attemptable",
    "blocked",
    "families",
    "direct_read",
  ]), "live proof cases.classification");
  for (const key of ["total", "attemptable", "blocked"]) {
    requireNonNegativeInteger(classification[key], `live proof cases.classification.${key}`);
  }
  if (classification.total !== classification.attemptable + classification.blocked) {
    throw new Error("live proof cases.classification must account for every case");
  }
  const families = requireKnownFields(
    classification.families,
    new Set(CASE_REASON_FAMILIES),
    "live proof cases.classification.families",
  );
  let familyTotal = 0;
  for (const family of CASE_REASON_FAMILIES) {
    requireNonNegativeInteger(families[family], `live proof cases.classification.families.${family}`);
    familyTotal += families[family];
  }
  if (familyTotal !== classification.blocked) {
    throw new Error("live proof cases.classification families must account for every blocked case");
  }
  const directRead = requireKnownFields(
    classification.direct_read,
    new Set(["total", "attemptable", "blocked"]),
    "live proof cases.classification.direct_read",
  );
  for (const key of ["total", "attemptable", "blocked"]) {
    requireNonNegativeInteger(directRead[key], `live proof cases.classification.direct_read.${key}`);
  }
  if (directRead.total !== directRead.attemptable + directRead.blocked) {
    throw new Error("live proof cases.classification.direct_read must account for every direct read");
  }
  return classification;
}

function validateCaseMeasurement(value, classification) {
  const measurement = requireKnownFields(value, new Set([
    "historical_terminal_measurement",
    "current",
    "movement",
  ]), "live proof cases.measurement");
  const historical = requireKnownFields(
    measurement.historical_terminal_measurement,
    new Set(Object.keys(HISTORICAL_TERMINAL_MEASUREMENT)),
    "live proof cases.measurement.historical_terminal_measurement",
  );
  if (stableJSONString(historical) !== stableJSONString(HISTORICAL_TERMINAL_MEASUREMENT)) {
    throw new Error("live proof cases.measurement must retain the fixed historical terminal measurement");
  }
  const current = requireKnownFields(
    measurement.current,
    new Set(["total", "attemptable", "blocked"]),
    "live proof cases.measurement.current",
  );
  const expectedCurrent = {
    total: classification.total,
    attemptable: classification.attemptable,
    blocked: classification.blocked,
  };
  if (stableJSONString(current) !== stableJSONString(expectedCurrent)) {
    throw new Error("live proof cases.measurement current tally does not match the classifier");
  }
  const movement = requireKnownFields(
    measurement.movement,
    new Set(["total", "attemptable", "blocked"]),
    "live proof cases.measurement.movement",
  );
  const expectedMovement = {
    total: current.total - historical.total,
    attemptable: current.attemptable - historical.failed,
    blocked: current.blocked - historical.untestable,
  };
  if (stableJSONString(movement) !== stableJSONString(expectedMovement)) {
    throw new Error("live proof cases.measurement movement does not match the classifier");
  }
}

function validatePersistedCaseFile(caseFile) {
  assertPersistedArtifactSafe(caseFile, "live proof cases");
  const source = requireKnownFields(caseFile, new Set([
    "schema_version",
    "connector",
    "test_repository",
    "context",
    "source",
    "cases",
    "case_digest",
    "measurement",
    "classification",
  ]), "live proof cases");
  if (source.schema_version !== 3) {
    throw new Error("live proof cases.schema_version must be the canonical schema version");
  }
  if (requireSafeText(source.connector, "live proof cases.connector") !== CONNECTOR) {
    throw new Error("live proof cases must be explicitly GitHub-only");
  }
  requireSafeText(source.source, "live proof cases.source");
  requireSHA256(source.case_digest, "live proof cases.case_digest");
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
  const classification = validateCaseClassification(source.classification);
  validateCaseMeasurement(source.measurement, classification);
}

function validateCaseFile(caseFile, boundary, surface) {
  validatePersistedCaseFile(caseFile);
  validateCaseExecutionConstraints(caseFile, boundary, surface);
  const canonical = validateCanonicalCaseArtifact({ caseFile, surface, boundary });
  return {
    context: canonical.context,
    cases: new Map(canonical.cases.map((item) => [item.command, item])),
    caseDigest: canonical.case_digest,
  };
}

function flagValues(args, name) {
  const flag = `--${name}`;
  const values = [];
  for (let index = 0; index < args.length; index += 1) {
    const argument = args[index];
    if (argument === flag) {
      const value = args[index + 1];
      if (!value || value.startsWith("--")) {
        throw new Error(`${flag} requires a value in a live proof case`);
      }
      values.push(value);
      index += 1;
    } else if (argument.startsWith(`${flag}=`)) {
      const value = argument.slice(flag.length + 1);
      if (!value) {
        throw new Error(`${flag} requires a value in a live proof case`);
      }
      values.push(value);
    }
  }
  return values;
}

function validateWriteRepositoryTarget(command, args, owner, repo) {
  for (const [flag, expected] of [["owner", owner], ["repo", repo]]) {
    for (const actual of flagValues(args, flag)) {
      if (actual !== expected) {
        throw new Error(`write case ${JSON.stringify(command)} may not override the dedicated repository ${flag}`);
      }
    }
  }
}

function validateCaseExecutionConstraints(caseFile, boundary, surface) {
  const surfaceCommands = new Map(surface.commands.map((command) => [command.path, command]));
  for (const item of caseFile.cases) {
    if (item.untestable_reason !== undefined) continue;
    for (const argument of item.args) {
      if (isReservedExecutionFlag(argument)) {
        throw new Error(`case ${JSON.stringify(item.command)} may not override lifecycle or credential flags`);
      }
    }
    const surfaceCommand = surfaceCommands.get(item.command);
    if (surfaceCommand?.intent === "reverse_etl" || surfaceCommand?.intent === "direct_write") {
      validateWriteRepositoryTarget(
        item.command,
        interpolateArguments(item.args, caseFile.context),
        boundary.owner,
        boundary.repo,
      );
    }
    if (item.readback === undefined) continue;
    for (const argument of item.readback.args) {
      if (isReservedExecutionFlag(argument)) {
        throw new Error(`readback for ${JSON.stringify(item.command)} may not override lifecycle or credential flags`);
      }
    }
    const readbackCommand = surfaceCommands.get(item.readback.command);
    if (!readbackCommand || readbackCommand.availability !== "implemented") {
      throw new Error(`readback for ${JSON.stringify(item.command)} names an unknown implemented command`);
    }
    if (readbackCommand.intent === "reverse_etl" || readbackCommand.intent === "direct_write") {
      throw new Error(`readback for ${JSON.stringify(item.command)} may not invoke another write`);
    }
    validateWriteRepositoryTarget(
      `${item.command} readback`,
      interpolateArguments(item.readback.args, caseFile.context),
      boundary.owner,
      boundary.repo,
    );
  }
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

export function runProcess(binary, args, cwd, { timeoutMs = PROCESS_TIMEOUT_MS, stdin = "" } = {}) {
  if (!Number.isInteger(timeoutMs) || timeoutMs <= 0) {
    throw new Error("process timeout must be a positive integer");
  }
  return new Promise((resolve, reject) => {
    const child = spawn(binary, args, { cwd, stdio: ["pipe", "pipe", "pipe"] });
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
    child.stdin.once("error", (error) => finish(() => reject(error)));
    child.once("error", (error) => finish(() => reject(error)));
    child.once("close", (code, signal) =>
      finish(() => resolve({ code, signal, stdout, stderr, overflow, timedOut, timeoutMs })),
    );
    child.stdin.end(stdin);
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

export async function runWriteLifecycle({ binary, root, credential, command, args, readback, owner, repo, cwd }) {
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
    "--approval-token-stdin",
    ...(challenge ? ["--confirm", challenge] : []),
    "--root",
    root,
    "--json",
  ];
  const runEnvelope = parseJSONOutput(
    await runProcess(binary, executeArgs, cwd, { stdin: grant + "\n" }),
    "write execution",
  );
  const run = runEnvelope?.run;
  if (runEnvelope?.kind !== "ReverseRun" || run?.status !== "completed" || run?.records_succeeded !== 1 || run?.records_failed !== 0) {
    throw new Error("write execution did not report one completed provider mutation");
  }
  const operation = run.operation_direct_write;
  let readbackResult;
  if (readback) {
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
    readbackResult = assertReadEnvelope(readbackEnvelope, readback.command);
  }
  return {
    httpStatus: Number.isInteger(operation?.status) ? operation.status : undefined,
    assertion: {
      kind: readbackResult ? "reverse-write-readback" : "reverse-write-result",
      subject: command,
      matched: true,
      ...(readbackResult ? { readback_kind: readbackResult.assertion.kind } : {}),
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
  const expected = enumerateImplementedCommands(surface);
  const surfaceCommands = new Map(surface.commands.map((command) => [command.path, command]));
  const cases = validateCaseFile(await loadJSON(casesPath), boundary, surface);
  // This script launches a child `pm` per operation. It must not touch a
  // credential or provider while pretending it can produce the in-process
  // current-certification evidence owned by `pm connectors certify`.
  assertCurrentCertificationExecutionModel(EXTERNAL_PM_PER_OPERATION);
  const binaryBytes = await readFile(binary);
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
    schema_version: 2,
    connector: CONNECTOR,
    status: "credentialed_live",
    generated_at: new Date().toISOString(),
    surface_sha256: createHash("sha256").update(JSON.stringify(surface)).digest("hex"),
    binary_sha256: createHash("sha256").update(binaryBytes).digest("hex"),
    case_digest: cases.caseDigest,
    execution_model: EXTERNAL_PM_PER_OPERATION,
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
  validateProofReport(report, expected, { caseDigest: cases.caseDigest });
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
    schema_version: 2,
    connector: CONNECTOR,
    generated_at: new Date().toISOString(),
    status: "external_blocker",
    surface_sha256: createHash("sha256").update(JSON.stringify(surface)).digest("hex"),
    binary_sha256: createHash("sha256").update(binaryBytes).digest("hex"),
    execution_model: EXTERNAL_PM_PER_OPERATION,
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
