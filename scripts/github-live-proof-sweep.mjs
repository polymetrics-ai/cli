#!/usr/bin/env node

import { createHash } from "node:crypto";
import { readFile, writeFile, chmod } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";
import { spawn } from "node:child_process";

const CONNECTOR = "github";
const APPROVED_TEST_OWNER = "karthik-sivadas";
const APPROVED_TEST_REPO = "pm-live-test-direct-read-20260808081515";
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
  "approve",
  "authorization",
  "client-secret",
  "key",
  "password",
  "private-key",
  "secret",
  "token",
  "value",
]);
const TOKEN_PATTERNS = [
  /\bgh[pousr]_[A-Za-z0-9_-]+\b/gi,
  /\bgithub_pat_[A-Za-z0-9_-]+\b/gi,
  /\b(?:bearer|token)\s+[A-Za-z0-9._~+\/-]{12,}\b/gi,
];
const OUTPUT_LIMIT_BYTES = 2 * 1024 * 1024;
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

function redactText(value) {
  let out = String(value ?? "");
  for (const pattern of TOKEN_PATTERNS) {
    out = out.replace(pattern, "<redacted>");
  }
  return out;
}

function hasTokenShapedValue(value) {
  const text = String(value ?? "");
  return TOKEN_PATTERNS.some((pattern) => {
    pattern.lastIndex = 0;
    return pattern.test(text);
  });
}

function redactInvocation(invocation) {
  if (!Array.isArray(invocation)) {
    throw new Error("proof invocation must be an argument array");
  }
  const out = [];
  let redactNext = false;
  for (const rawArgument of invocation) {
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
    const withoutPrefix = argument.slice(2);
    const equals = withoutPrefix.indexOf("=");
    const name = (equals === -1 ? withoutPrefix : withoutPrefix.slice(0, equals)).toLowerCase();
    if (SENSITIVE_ARGUMENT_NAMES.has(name)) {
      if (equals === -1) {
        out.push(argument);
        redactNext = true;
      } else {
        out.push(`--${name}=<redacted>`);
      }
      continue;
    }
    out.push(redactText(argument));
  }
  return out;
}

function safeAssertion(assertion) {
  if (!assertion || typeof assertion !== "object" || Array.isArray(assertion)) {
    throw new Error("proof assertion must be an object");
  }
  const out = {};
  for (const [key, value] of Object.entries(assertion)) {
    if (typeof value === "string") {
      out[key] = redactText(value);
    } else if (typeof value === "number" || typeof value === "boolean" || value === null) {
      out[key] = value;
    } else {
      throw new Error("proof assertion values must be scalar");
    }
  }
  return out;
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
    if (!record || typeof record !== "object" || Array.isArray(record)) {
      throw new Error("proof record must be an object");
    }
    if (reportContainsForbiddenField(record)) {
      throw new Error(`proof record for ${String(record.command || "<unknown>")} contains raw execution data`);
    }
    if (hasTokenShapedValue(JSON.stringify(record))) {
      throw new Error(`proof record for ${String(record.command || "<unknown>")} contains credential-shaped data`);
    }
    const command = String(record.command || "").trim();
    if (!expected.has(command)) {
      throw new Error(`proof record names unimplemented or unknown command ${JSON.stringify(command)}`);
    }
    if (byCommand.has(command)) {
      throw new Error(`duplicate terminal result for ${JSON.stringify(command)}`);
    }
    if (!TERMINAL_STATES.has(record.state)) {
      throw new Error(`invalid terminal state for ${JSON.stringify(command)}`);
    }
    if (record.state === "proven") {
      if (record.http_status !== undefined &&
          (!Number.isInteger(record.http_status) || record.http_status < 200 || record.http_status >= 300)) {
        throw new Error(`proven result for ${JSON.stringify(command)} has an invalid HTTP status`);
      }
      if (!record.assertion || record.assertion.matched !== true || typeof record.assertion.kind !== "string") {
        throw new Error(`proven result for ${JSON.stringify(command)} requires a matched returned-data assertion`);
      }
    } else if (!concreteReason(record.reason)) {
      throw new Error(`${record.state} result for ${JSON.stringify(command)} requires a concrete reason`);
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

function requireStringMap(value, name) {
  if (!value || typeof value !== "object" || Array.isArray(value)) {
    throw new Error(`${name} must be an object`);
  }
  const out = {};
  for (const [key, item] of Object.entries(value)) {
    if (typeof item !== "string" && typeof item !== "number") {
      throw new Error(`${name}.${key} must be a scalar`);
    }
    if (SENSITIVE_ARGUMENT_NAMES.has(key.toLowerCase())) {
      throw new Error(`${name}.${key} may not carry a credential or grant`);
    }
    out[key] = String(item);
  }
  return out;
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
      continue;
    }
    if (argument.startsWith(`${flag}=`)) {
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
        throw new Error(
          `write case ${JSON.stringify(command)} may not override the dedicated repository ${flag}`,
        );
      }
    }
  }
}

function validateCaseFile(caseFile, expected, owner, repo, surfaceCommands) {
  if (!caseFile || caseFile.connector !== CONNECTOR) {
    throw new Error("live proof cases must be explicitly GitHub-only");
  }
  const repository = caseFile.test_repository;
  if (!repository || repository.owner !== owner || repository.repo !== repo) {
    throw new Error("case file test_repository must exactly match --test-owner and --test-repo");
  }
  if (!Array.isArray(caseFile.cases)) {
    throw new Error("case file must contain a cases array");
  }
  const context = requireStringMap(caseFile.context || {}, "context");
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
    const hasUntestableReason = typeof item.untestable_reason === "string";
    if (hasUntestableReason && !concreteReason(item.untestable_reason)) {
      throw new Error(`untestable case for ${JSON.stringify(command)} requires a concrete reason`);
    }
    if (!hasUntestableReason) {
      if (!Array.isArray(item.args) || item.args.some((argument) => typeof argument !== "string")) {
        throw new Error(`executable case for ${JSON.stringify(command)} requires string args`);
      }
      for (const argument of item.args) {
        if (isReservedExecutionFlag(argument)) {
          throw new Error(`case ${JSON.stringify(command)} may not override lifecycle or credential flags`);
        }
      }
      const surfaceCommand = surfaceCommands.get(command);
      const isWrite = surfaceCommand?.intent === "reverse_etl" || surfaceCommand?.intent === "direct_write";
      if (isWrite) {
        validateWriteRepositoryTarget(
          command,
          interpolateArguments(item.args, {
            ...context,
            test_owner: owner,
            test_repo: repo,
            test_repository: `${owner}/${repo}`,
          }),
          owner,
          repo,
        );
      }
      if (item.readback !== undefined) {
        if (!item.readback || typeof item.readback !== "object" || Array.isArray(item.readback)) {
          throw new Error(`readback for ${JSON.stringify(command)} must be an object`);
        }
        const readbackCommand = String(item.readback.command || "").trim();
        const readbackSurfaceCommand = surfaceCommands.get(readbackCommand);
        if (!readbackSurfaceCommand || !expected.includes(readbackCommand)) {
          throw new Error(`readback for ${JSON.stringify(command)} names an unknown implemented command`);
        }
        if (readbackSurfaceCommand.intent === "reverse_etl" || readbackSurfaceCommand.intent === "direct_write") {
          throw new Error(`readback for ${JSON.stringify(command)} may not invoke another write`);
        }
        if (!Array.isArray(item.readback.args) || item.readback.args.some((argument) => typeof argument !== "string")) {
          throw new Error(`readback for ${JSON.stringify(command)} requires string args`);
        }
        for (const argument of item.readback.args) {
          if (isReservedExecutionFlag(argument)) {
            throw new Error(`readback for ${JSON.stringify(command)} may not override lifecycle or credential flags`);
          }
        }
        validateWriteRepositoryTarget(
          `${command} readback`,
          interpolateArguments(item.readback.args, {
            ...context,
            test_owner: owner,
            test_repo: repo,
            test_repository: `${owner}/${repo}`,
          }),
          owner,
          repo,
        );
      }
    }
    commands.set(command, item);
  }
  if (commands.size !== expected.length) {
    const missing = expected.find((command) => !commands.has(command));
    throw new Error(`case file is incomplete; missing case for ${JSON.stringify(missing)}`);
  }
  return {
    context,
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

function runProcess(binary, args, cwd, stdin = "") {
  return new Promise((resolve, reject) => {
    const child = spawn(binary, args, { cwd, stdio: ["pipe", "pipe", "pipe"] });
    let stdout = "";
    let stderr = "";
    let bytes = 0;
    let overflow = false;
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
    child.stdin.on("error", reject);
    child.on("error", reject);
    child.on("close", (code, signal) => resolve({ code, signal, stdout, stderr, overflow }));
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

export async function runWriteLifecycle({ binary, root, credential, command, args, readback, cwd }) {
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
  const runEnvelope = parseJSONOutput(await runProcess(binary, executeArgs, cwd, grant + "\n"), "write execution");
  const run = runEnvelope?.run;
  if (runEnvelope?.kind !== "ReverseRun" || run?.status !== "completed" || run?.records_succeeded !== 1 || run?.records_failed !== 0) {
    throw new Error("write execution did not report one completed provider mutation");
  }
  const operation = run.operation_direct_write;
  let readbackResult;
  if (readback) {
    const readbackArgs = interpolateArguments(readback.args, {
      test_owner: APPROVED_TEST_OWNER,
      test_repo: APPROVED_TEST_REPO,
      test_repository: `${APPROVED_TEST_OWNER}/${APPROVED_TEST_REPO}`,
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

async function validateCredentialScope(binary, root, credential, owner, repo, cwd) {
  const result = await runProcess(binary, ["credentials", "inspect", credential, "--root", root, "--json"], cwd);
  const envelope = parseJSONOutput(result, "credential inspection");
  const metadata = envelope?.credential;
  if (envelope?.kind !== "Credential" || metadata?.connector !== CONNECTOR) {
    throw new Error("named credential is not a GitHub credential");
  }
  if (metadata?.config?.owner !== owner || metadata?.config?.repo !== repo) {
    throw new Error("named credential is not scoped to the dedicated private test repository");
  }
}

async function executeLive(options) {
  if (options.connector !== undefined) {
    throw new Error("this runner is GitHub-only; --connector is not accepted");
  }
  if (options._.length !== 0) {
    throw new Error("live proof runner accepts options only");
  }
  const binary = requiredOption(options, "pm");
  const root = requiredOption(options, "root");
  const credential = requiredOption(options, "credential");
  const suppliedOwner = String(options["test-owner"] || "").trim();
  const suppliedRepo = String(options["test-repo"] || "").trim();
  if ((suppliedOwner && suppliedOwner !== APPROVED_TEST_OWNER) || (suppliedRepo && suppliedRepo !== APPROVED_TEST_REPO)) {
    throw new Error("live proof is hard-pinned to the approved private test repository");
  }
  const owner = APPROVED_TEST_OWNER;
  const repo = APPROVED_TEST_REPO;
  const casesPath = requiredOption(options, "cases");
  const reportPath = requiredOption(options, "report");
  const surface = await loadJSON(SURFACE_PATH);
  const binaryBytes = await readFile(binary);
  const caseBytes = await readFile(casesPath);
  const expected = enumerateImplementedCommands(surface);
  const surfaceCommands = new Map(surface.commands.map((command) => [command.path, command]));
  const cases = validateCaseFile(await loadJSON(casesPath), expected, owner, repo, surfaceCommands);
  await validateCredentialScope(binary, root, credential, owner, repo, REPOSITORY_ROOT);

  const records = [];
  for (const command of expected) {
    const caseItem = cases.cases.get(command);
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
    const invocation = shellSafeInvocation(command, args, credential, root);
    try {
    const result = isWrite
        ? await runWriteLifecycle({
            binary,
            root,
            credential,
            command,
            args,
            readback: caseItem.readback,
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
      records.push(redactForReport({
        command,
        state: "proven",
        invocation,
        http_status: result.httpStatus,
        assertion: result.assertion,
      }));
    } catch (error) {
      const diagnostic = error instanceof Error ? error.message : "live execution failed without a safe diagnostic";
      records.push(redactForReport({
        command,
        state: "failed",
        invocation,
        http_status: error?.httpStatus,
        reason: concreteReason(diagnostic)
          ? diagnostic
          : `${command} live execution failed without a safe provider diagnostic`,
      }));
    }
  }

  const tally = validateProofRecords(expected, records);
  const report = {
    schema_version: 1,
    connector: CONNECTOR,
    status: "credentialed_live",
    generated_at: new Date().toISOString(),
    surface_sha256: createHash("sha256").update(JSON.stringify(surface)).digest("hex"),
    binary_sha256: createHash("sha256").update(binaryBytes).digest("hex"),
    case_file_sha256: createHash("sha256").update(caseBytes).digest("hex"),
    test_repository: "<dedicated-private-test-repository>",
    implemented_commands: expected.length,
    tally,
    records,
  };
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
  const binary = requiredOption(options, "pm");
  const reportPath = requiredOption(options, "report");
  const reason = requiredOption(options, "reason");
  const surface = await loadJSON(SURFACE_PATH);
  const expected = enumerateImplementedCommands(surface);
  const records = expected.map((command) =>
    redactForReport({ command, state: "untestable", reason }),
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
    blocker: reason,
    tally,
    records,
  };
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
