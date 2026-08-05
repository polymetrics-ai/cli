#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
fixture_agent_dir="$repo_root/scripts/tests/testdata/pi-clean-project/agent"

fail() {
  printf 'pi-clean-project-agents test failed: %s\n' "$1" >&2
  exit 1
}

bun_bin="$(command -v bun || true)"
if [[ -z "$bun_bin" ]]; then
  fail 'bun is required to execute the installed Pi extension test'
fi

node_modules="$(npm root -g 2>/dev/null || true)"
if [[ ! -d "$node_modules/@earendil-works/pi-coding-agent" ]]; then
  fail 'the installed Pi coding-agent package is required to execute the extension test'
fi
pi_dependency_modules="$node_modules/@earendil-works/pi-coding-agent/node_modules"
if [[ ! -d "$pi_dependency_modules" ]]; then
  fail 'the installed Pi coding-agent dependencies are required to execute the extension test'
fi

# Keep the JavaScript program single-quoted so Bun receives template literals verbatim.
# shellcheck disable=SC2016
PI_SUB_AGENT_DEPTH=0 \
PI_SUBAGENT_SESSION_DIR="" \
PI_CODING_AGENT_DIR="$fixture_agent_dir" \
NODE_PATH="$pi_dependency_modules:$node_modules" \
PI_CLEAN_PROJECT_REPO_ROOT="$repo_root" \
"$bun_bin" -e '
import assert from "node:assert/strict";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";

const repoRoot = process.env.PI_CLEAN_PROJECT_REPO_ROOT;
if (!repoRoot) throw new Error("PI_CLEAN_PROJECT_REPO_ROOT is required");
const fixtureAgentDir = process.env.PI_CODING_AGENT_DIR;
if (!fixtureAgentDir) throw new Error("PI_CODING_AGENT_DIR is required");

const agentsModule = await import(path.join(repoRoot, ".pi/extensions/pi-sub-agent/agents.ts"));
const discovery = agentsModule.discoverAgents(repoRoot, "clean-project");
const expectedNames = ["pm-connector-worker", "pm-delivery-worker"];
const discovered = discovery.agents.map((agent) => `${agent.name} (${agent.source})`).sort();
const unexpected = discovered.filter((agent) => !expectedNames.some((name) => agent.startsWith(`${name} (`)));
if (unexpected.length > 0) {
  throw new Error(`clean-project exposed ambient agents: ${unexpected.join(", ")}`);
}
assert.deepEqual(discovered, expectedNames.map((name) => `${name} (project)`));
assert.equal(agentsModule.shouldConfirmProjectAgents("clean-project"), true, "clean-project must retain project confirmation");
assert.equal(agentsModule.shouldConfirmProjectAgents("user"), false, "user scope must not require project confirmation");

const bothScopeDelivery = agentsModule.discoverAgents(repoRoot, "both").agents.find((agent) => agent.name === "pm-delivery-worker");
assert.equal(bothScopeDelivery?.source, "project", "project definitions must override same-named global definitions in both scope");

const symlinkRoot = fs.mkdtempSync(path.join(os.tmpdir(), "pi-clean-project-agents-"));
try {
  const canonicalDir = path.join(symlinkRoot, ".agents", "agentic-delivery", "canonical");
  const projectAgentsDir = path.join(symlinkRoot, ".pi", "agents");
  fs.mkdirSync(canonicalDir, { recursive: true });
  fs.mkdirSync(projectAgentsDir, { recursive: true });
  fs.copyFileSync(path.join(repoRoot, ".agents", "agentic-delivery", "canonical", "delivery-contract.json"), path.join(canonicalDir, "delivery-contract.json"));
  fs.symlinkSync(path.join(fixtureAgentDir, "agents", "ambient-global.md"), path.join(projectAgentsDir, "pm-delivery-worker.md"));
  assert.deepEqual(agentsModule.discoverAgents(symlinkRoot, "clean-project").agents, [], "clean-project must reject a generated worker symlinked to an ambient definition");
  fs.rmSync(path.join(symlinkRoot, ".pi"), { recursive: true, force: true });
  const ambientPiDir = path.join(symlinkRoot, "ambient-pi");
  fs.mkdirSync(path.join(ambientPiDir, "agents"), { recursive: true });
  fs.copyFileSync(path.join(fixtureAgentDir, "agents", "ambient-global.md"), path.join(ambientPiDir, "agents", "pm-delivery-worker.md"));
  fs.symlinkSync(ambientPiDir, path.join(symlinkRoot, ".pi"));
  assert.deepEqual(agentsModule.discoverAgents(symlinkRoot, "clean-project").agents, [], "clean-project must reject a symlinked .pi directory");
} finally {
  fs.rmSync(symlinkRoot, { recursive: true, force: true });
}

const deliveryWorkerPath = path.join(repoRoot, ".pi", "agents", "pm-delivery-worker.md");
const originalDeliveryWorker = fs.readFileSync(deliveryWorkerPath, "utf8");
try {
  const handEditedWorker = originalDeliveryWorker.replace(/^tools:\n(?:  - .*\n)+/m, "");
  assert.notEqual(handEditedWorker, originalDeliveryWorker, "test fixture must remove generated worker tools");
  fs.writeFileSync(deliveryWorkerPath, handEditedWorker);
  assert.deepEqual(agentsModule.discoverAgents(repoRoot, "clean-project").agents, [], "clean-project must reject a parseable generated worker that diverges from the canonical projection");
} finally {
  fs.writeFileSync(deliveryWorkerPath, originalDeliveryWorker);
}

const canonicalContractPath = path.join(repoRoot, ".agents", "agentic-delivery", "canonical", "delivery-contract.json");
const originalContract = fs.readFileSync(canonicalContractPath, "utf8");
try {
  const handEditedContract = originalContract.replace("\"clean_project_scope\": \"clean-project\"", "\"clean_project_scope\": \"project\"");
  assert.notEqual(handEditedContract, originalContract, "test fixture must alter the clean-project scope");
  fs.writeFileSync(canonicalContractPath, handEditedContract);
  assert.deepEqual(agentsModule.discoverAgents(repoRoot, "clean-project").agents, [], "clean-project must reject a parseable contract that fails canonical validation");
} finally {
  fs.writeFileSync(canonicalContractPath, originalContract);
}

const contract = JSON.parse(fs.readFileSync(path.join(repoRoot, ".agents/agentic-delivery/canonical/delivery-contract.json"), "utf8"));
const expectedTools = contract.pi_harness.child_tools;
for (const agent of discovery.agents) {
  assert.deepEqual(agent.tools, expectedTools, `${agent.name} must use the canonical child tool allowlist`);
  assert.equal(agent.tools.includes("subagent"), false, `${agent.name} must not receive subagent`);
}

const extensionEntry = fs.readFileSync(path.join(repoRoot, ".pi/extensions/pi-sub-agent/index.ts"), "utf8");
assert.match(extensionEntry, /from "\.\/child-policy\.js"/, "extension entry point must use the isolated child policy");
const extension = await import(path.join(repoRoot, ".pi/extensions/pi-sub-agent/index.ts"));
let registeredTool;
extension.default({
  getActiveTools: () => ["read", "grep", "find", "ls", "bash", "edit", "write", "subagent"],
  getThinkingLevel: () => "off",
  on: () => {},
  registerTool: (tool) => { registeredTool = tool; },
});
assert.ok(registeredTool, "extension must register the subagent tool");
const unknownResult = await registeredTool.execute(
  "test",
  { agent: "ambient-global", task: "must not run" },
  new AbortController().signal,
  () => {},
  { cwd: repoRoot, hasUI: false },
);
assert.match(unknownResult.content[0].text, /Unknown agent: "ambient-global"\. Available: pm-connector-worker, pm-delivery-worker/, "default tool execution must expose only canonical clean workers");
const unconfirmedProjectResult = await registeredTool.execute(
  "test",
  { agent: "pm-delivery-worker", task: "must not start without project trust confirmation" },
  new AbortController().signal,
  () => {},
  { cwd: repoRoot, hasUI: false },
);
assert.match(unconfirmedProjectResult.content[0].text, /running project-local agents requires confirmation/, "clean-project workers must keep the noninteractive confirmation refusal");
const extensionModule = await import(path.join(repoRoot, ".pi/extensions/pi-sub-agent/child-policy.ts"));
assert.deepEqual(
  extensionModule.resolveChildToolAllowlist(["read", "subagent"], ["read", "bash", "subagent"]),
  ["read"],
  "child tool policy must remove subagent even when requested",
);
assert.equal(extensionModule.canStartSubagent("1"), false, "nested child depth must be rejected before spawn");
assert.deepEqual(
  extensionModule.childInvocationArgs(undefined),
  ["--mode", "json", "-p", "--no-session", "--no-extensions"],
  "child Pi invocation must disable all extensions",
);
'
