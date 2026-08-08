#!/usr/bin/env node

import { readFile, writeFile, chmod } from "node:fs/promises";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

const SCRIPT_DIR = path.dirname(fileURLToPath(import.meta.url));
const ROOT = path.resolve(SCRIPT_DIR, "..");
const PHASE_DIR = path.join(ROOT, ".planning/phases/github-parity-extract-r1");
const LIVE_REPORT = path.join(PHASE_DIR, "LIVE-PROOF-REPORT.json");
const STATIC_PROOF = path.join(PHASE_DIR, "RATE-LIMIT-PROOF.json");
const OUTPUT = path.join(PHASE_DIR, "LIVE-RATE-LIMIT-PROOF.json");

async function main() {
  const live = JSON.parse(await readFile(LIVE_REPORT, "utf8"));
  const staticProof = JSON.parse(await readFile(STATIC_PROOF, "utf8"));
  if (live.status !== "credentialed_live" || live.tally.failed !== 0) {
    throw new Error("credentialed live report is not an accepted zero-failure run");
  }
  const rateLimit = live.records.find((record) => record.command === "rate-limit get");
  if (!rateLimit || rateLimit.state !== "proven" || rateLimit.http_status !== 200) {
    throw new Error("credentialed live report did not prove rate-limit get");
  }
  const observed429 = live.records.filter((record) => record.http_status === 429).length;
  const report = {
    schema_version: 1,
    connector: "github",
    status: "PROVEN",
    proof_mode: "current_head_credentialed",
    binary_sha256: live.binary_sha256,
    case_file_sha256: live.case_file_sha256,
    live_command_rows: live.implemented_commands,
    live_proven_rows: live.tally.proven,
    observed_http_429: observed429,
    headroom_preserved: observed429 === 0,
    rate_limit_command: {
      command: rateLimit.command,
      http_status: rateLimit.http_status,
      assertion_kind: rateLimit.assertion.kind,
      assertion_matched: rateLimit.assertion.matched,
    },
    admission_and_scope_unit: {
      status: staticProof.status,
      test: staticProof.test,
      same_scope_waits: staticProof.evidence.same_scope_local_waits,
      independent_scope_waits: staticProof.evidence.independent_scope_extra_waits,
    },
    interpretation: "The bounded live sweep observed no HTTP 429 and did not intentionally exhaust the provider budget; the unit proof covers admission ordering and same-scope isolation without spending live quota.",
  };
  await writeFile(OUTPUT, `${JSON.stringify(report, null, 2)}\n`, { mode: 0o600 });
  await chmod(OUTPUT, 0o600);
  process.stdout.write(`github live rate-limit proof: ${report.status} 429=${observed429}\n`);
}

if (path.resolve(process.argv[1] || "") === fileURLToPath(import.meta.url)) {
  main().catch((error) => {
    process.stderr.write(`github live rate-limit proof: ${error instanceof Error ? error.message : "generation failed"}\n`);
    process.exitCode = 1;
  });
}
