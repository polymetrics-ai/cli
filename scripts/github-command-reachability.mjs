#!/usr/bin/env node

import { createHash } from "node:crypto";
import { chmod, mkdir, mkdtemp, readFile, rm, writeFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { spawn } from "node:child_process";

const SCRIPT_DIR = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(SCRIPT_DIR, "..");
const SURFACE = path.join(ROOT, "internal/connectors/defs/github/cli_surface.json");
const DEFAULT_REPORT = path.join(ROOT, ".planning/phases/github-parity-extract-r1/COMMAND-REACHABILITY.json");
const MAX_OUTPUT = 256 * 1024;

async function loadSurface() {
  return JSON.parse(await readFile(SURFACE, "utf8"));
}

function parseOptions(argv) {
  const options = {};
  for (let index = 0; index < argv.length; index += 1) {
    const argument = argv[index];
    if (!argument.startsWith("--")) throw new Error(`unexpected argument ${JSON.stringify(argument)}`);
    const equals = argument.indexOf("=");
    const name = argument.slice(2, equals === -1 ? undefined : equals);
    if (!name) throw new Error("empty option name");
    if (equals !== -1) {
      options[name] = argument.slice(equals + 1);
      continue;
    }
    const value = argv[index + 1];
    if (!value || value.startsWith("--")) throw new Error(`--${name} requires a value`);
    options[name] = value;
    index += 1;
  }
  return options;
}

function required(options, name) {
  const value = String(options[name] || "").trim();
  if (!value) throw new Error(`--${name} is required`);
  return value;
}

function runProcess(binary, args, cwd) {
  return new Promise((resolve, reject) => {
    const child = spawn(binary, args, { cwd, stdio: ["ignore", "pipe", "pipe"] });
    let stdout = "";
    let stderr = "";
    let size = 0;
    let overflow = false;
    const consume = (target, chunk) => {
      size += chunk.length;
      if (size > MAX_OUTPUT) {
        overflow = true;
        child.kill("SIGTERM");
        return;
      }
      if (target === "stdout") stdout += chunk.toString("utf8");
      else stderr += chunk.toString("utf8");
    };
    child.stdout.on("data", (chunk) => consume("stdout", chunk));
    child.stderr.on("data", (chunk) => consume("stderr", chunk));
    child.on("error", reject);
    child.on("close", (code, signal) => resolve({ code, signal, stdout, stderr, overflow }));
  });
}

function renderedName(stdout) {
  return /^  (pm github .+?)(?: - .*)?$/m.exec(stdout)?.[1] || "";
}

export function classifyHelpResult(command, result) {
  const expected = `pm github ${command}`;
  const actual = renderedName(result.stdout || "");
  if (actual === expected) return { state: "reachable", rendered_name: actual };
  if (actual.startsWith("pm github ")) {
    return { state: "unreachable", reason: "rendered namespace help instead of the declared command" };
  }
  if (/unknown command/i.test(`${result.stdout}\n${result.stderr}`)) {
    return { state: "unreachable", reason: "binary returned unknown command" };
  }
  if (result.overflow) return { state: "unreachable", reason: "help output exceeded the bounded capture limit" };
  return {
    state: "unreachable",
    reason: result.code === 0 ? "binary did not render the declared command name" : "binary command help failed",
  };
}

function summarize(records) {
  const out = { total: records.length, reachable: 0, unreachable: 0, byAvailability: {} };
  for (const record of records) {
    out[record.state] += 1;
    out.byAvailability[record.availability] ||= { total: 0, reachable: 0, unreachable: 0 };
    out.byAvailability[record.availability].total += 1;
    out.byAvailability[record.availability][record.state] += 1;
  }
  return out;
}

async function sweep(options) {
  const pm = path.resolve(required(options, "pm"));
  const baseRoot = path.resolve(required(options, "root"));
  const reportPath = path.resolve(options.report || DEFAULT_REPORT);
  const workers = Math.max(1, Math.min(16, Number.parseInt(options.workers || "8", 10) || 8));
  const surface = await loadSurface();
  const commands = surface.commands || [];
  if (!commands.length) throw new Error("GitHub command surface is empty");
  await mkdir(baseRoot, { recursive: true });
  const workerRoots = await Promise.all(
    Array.from({ length: Math.min(workers, commands.length) }, () => mkdtemp(path.join(baseRoot, "worker-"))),
  );
  try {
    for (const workerRoot of workerRoots) {
      const init = await runProcess(pm, ["init", "--root", workerRoot, "--json"], ROOT);
      if (init.code !== 0) throw new Error("pm init failed for reachability worker");
    }
    const records = new Array(commands.length);
    await Promise.all(
      workerRoots.map(async (workerRoot, workerIndex) => {
        for (let index = workerIndex; index < commands.length; index += workerRoots.length) {
          const command = commands[index];
          const result = await runProcess(
            pm,
            ["github", ...String(command.path).split(" "), "--help", "--root", workerRoot],
            ROOT,
          );
          records[index] = {
            command: command.path,
            availability: command.availability,
            intent: command.intent,
            ...classifyHelpResult(command.path, result),
            exit_code: result.code,
            evidence: "built binary rendered command help NAME line",
          };
        }
      }),
    );
    const summary = summarize(records);
    if (summary.unreachable !== 0) throw new Error(`binary reachability found ${summary.unreachable} unreachable command(s)`);
    const report = {
      schema_version: 1,
      connector: "github",
      binary_sha256: createHash("sha256").update(await readFile(pm)).digest("hex"),
      surface_sha256: createHash("sha256").update(JSON.stringify(surface)).digest("hex"),
      method: "one initialized isolated project per worker; exact rendered NAME line",
      summary,
      records,
    };
    await writeFile(reportPath, `${JSON.stringify(report, null, 2)}\n`, { mode: 0o600 });
    await chmod(reportPath, 0o600);
    process.stdout.write(`github binary reachability: ${JSON.stringify(summary)}\n`);
  } finally {
    await Promise.all(workerRoots.map((workerRoot) => rm(workerRoot, { recursive: true, force: true })));
  }
}

async function main() {
  await sweep(parseOptions(process.argv.slice(2)));
}

if (path.resolve(process.argv[1] || "") === fileURLToPath(import.meta.url)) {
  main().catch((error) => {
    process.stderr.write(`github binary reachability: ${error instanceof Error ? error.message : String(error)}\n`);
    process.exitCode = 1;
  });
}
