#!/usr/bin/env node

import { readFile, writeFile, chmod } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

import { assertPersistedArtifactSafe } from "./github-live-artifact-guard.mjs";

const SCRIPT_DIR = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(SCRIPT_DIR, "..");
const PHASE_DIR = path.join(ROOT, ".planning/phases/github-parity-extract-r1");
const LIVE_REPORT = path.join(PHASE_DIR, "LIVE-PROOF-REPORT.json");
const STATIC_PROOF = path.join(PHASE_DIR, "RATE-LIMIT-PROOF.json");
const OUTPUT = path.join(PHASE_DIR, "LIVE-RATE-LIMIT-PROOF.json");
const SHA256 = /^[a-f0-9]{64}$/iu;

function isPlainObject(value) {
  return value !== null && typeof value === "object" && !Array.isArray(value);
}

function requireSHA256(value, label) {
  if (typeof value !== "string" || !SHA256.test(value)) {
    throw new Error(`${label} must be a SHA-256 digest`);
  }
  return value;
}

function requireCount(value, label) {
  if (!Number.isSafeInteger(value) || value < 0) {
    throw new Error(`${label} must be a non-negative integer`);
  }
  return value;
}

export function buildHistoricalRateLimitProof({ legacyLive, staticProof }) {
  assertPersistedArtifactSafe(legacyLive, "legacy GitHub live proof report");
  assertPersistedArtifactSafe(staticProof, "GitHub rate-limit unit evidence");
  if (!isPlainObject(legacyLive) || legacyLive.schema_version !== 1 || legacyLive.connector !== "github") {
    throw new Error("legacy GitHub live proof report has an unsupported shape");
  }
  if (legacyLive.status !== "credentialed_live" || Object.hasOwn(legacyLive, "case_digest")) {
    throw new Error("legacy GitHub live proof report must retain only its archived case-file binding");
  }
  const caseFileSHA256 = requireSHA256(legacyLive.case_file_sha256, "legacy live proof case_file_sha256");
  const binarySHA256 = requireSHA256(legacyLive.binary_sha256, "legacy live proof binary_sha256");
  const commandRows = requireCount(legacyLive.implemented_commands, "legacy live proof implemented_commands");
  if (!isPlainObject(legacyLive.tally)) {
    throw new Error("legacy GitHub live proof report must contain a tally");
  }
  const returnedRows = requireCount(legacyLive.tally.proven, "legacy live proof tally.proven");
  if (!Array.isArray(legacyLive.records)) {
    throw new Error("legacy GitHub live proof report must contain records");
  }
  const rateLimit = legacyLive.records.find((record) => record?.command === "rate-limit get");
  if (!isPlainObject(rateLimit) || rateLimit.state !== "proven" || rateLimit.http_status !== 200 || !isPlainObject(rateLimit.assertion)) {
    throw new Error("legacy GitHub live proof report lacks the archived rate-limit observation");
  }
  if (!isPlainObject(staticProof) || staticProof.connector !== "github" || typeof staticProof.test !== "string" || !isPlainObject(staticProof.evidence)) {
    throw new Error("GitHub rate-limit unit evidence has an unsupported shape");
  }
  const sameScopeWaits = requireCount(staticProof.evidence.same_scope_local_waits, "rate-limit unit same_scope_local_waits");
  const independentScopeWaits = requireCount(staticProof.evidence.independent_scope_extra_waits, "rate-limit unit independent_scope_extra_waits");
  const observed429 = legacyLive.records.filter((record) => record?.http_status === 429).length;
  const report = {
    schema_version: 2,
    connector: "github",
    status: "historical_observation",
    proof_mode: "archived_credentialed_observation",
    current_certification: {
      status: "not_proven",
      rate_limit_get: "untestable",
    },
    case_file_sha256: caseFileSHA256,
    historical_observation: {
      binary_sha256: binarySHA256,
      command_rows: commandRows,
      returned_rows: returnedRows,
      observed_http_429: observed429,
      headroom_preserved: observed429 === 0,
      rate_limit_command: {
        command: rateLimit.command,
        http_status: rateLimit.http_status,
        assertion_kind: rateLimit.assertion.kind,
        assertion_matched: rateLimit.assertion.matched,
      },
    },
    admission_and_scope_unit: {
      test: staticProof.test,
      same_scope_waits: sameScopeWaits,
      independent_scope_waits: independentScopeWaits,
    },
    interpretation: "Archived observation only; current certification is not proven and rate-limit get remains untestable in the canonical live classifier.",
  };
  assertPersistedArtifactSafe(report, "historical GitHub rate-limit proof");
  return report;
}

async function main() {
  const [legacyLive, staticProof] = await Promise.all([
    readFile(LIVE_REPORT, "utf8").then(JSON.parse),
    readFile(STATIC_PROOF, "utf8").then(JSON.parse),
  ]);
  const report = buildHistoricalRateLimitProof({ legacyLive, staticProof });
  await writeFile(OUTPUT, `${JSON.stringify(report, null, 2)}\n`, { mode: 0o600 });
  await chmod(OUTPUT, 0o600);
  process.stdout.write(`github historical rate-limit observation: certification=${report.current_certification.status} 429=${report.historical_observation.observed_http_429}\n`);
}

if (path.resolve(process.argv[1] || "") === fileURLToPath(import.meta.url)) {
  main().catch((error) => {
    process.stderr.write(`github historical rate-limit observation: ${error instanceof Error ? error.message : "generation failed"}\n`);
    process.exitCode = 1;
  });
}
