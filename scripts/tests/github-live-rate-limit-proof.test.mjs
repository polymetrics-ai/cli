import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import path from "node:path";
import test from "node:test";
import { fileURLToPath } from "node:url";

import * as rateLimitProof from "../github-live-rate-limit-proof.mjs";

const root = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "../..");
const phaseDir = path.join(root, ".planning/phases/github-parity-extract-r1");

async function loadJSON(name) {
  return JSON.parse(await readFile(path.join(phaseDir, name), "utf8"));
}

test("retires the legacy rate proof without current certification semantics", async () => {
  const [legacyLive, staticProof, checkedIn] = await Promise.all([
    loadJSON("LIVE-PROOF-REPORT.json"),
    loadJSON("RATE-LIMIT-PROOF.json"),
    loadJSON("LIVE-RATE-LIMIT-PROOF.json"),
  ]);

  assert.equal(typeof rateLimitProof.buildHistoricalRateLimitProof, "function");
  const report = rateLimitProof.buildHistoricalRateLimitProof({ legacyLive, staticProof });

  assert.equal(report.status, "historical_observation");
  assert.equal(report.proof_mode, "archived_credentialed_observation");
  assert.deepEqual(report.current_certification, {
    status: "not_proven",
    rate_limit_get: "untestable",
  });
  assert.equal(report.case_file_sha256, legacyLive.case_file_sha256);
  assert.equal(Object.hasOwn(report, "case_digest"), false);
  assert.equal(JSON.stringify(report).includes('"PROVEN"'), false);
  assert.deepEqual(checkedIn, report);
});
