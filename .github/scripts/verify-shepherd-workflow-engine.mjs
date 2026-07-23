import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { resolve } from "node:path";

const EXPECTED_SOURCE = "npm:pi-workflow-engine@0.12.0";
const EXPECTED_VERSION = "0.12.0";
const EXPECTED_RESOLVED = "https://registry.npmjs.org/pi-workflow-engine/-/pi-workflow-engine-0.12.0.tgz";
const EXPECTED_INTEGRITY = "sha512-DX+e2U03raK8o8YbwnDUcAQSKNZm0v1J6jWS+bk2j2kEFihLmZCf0sUlrHWou1kWC3Zw+CA4HCgqpjLWlmtcRg==";
const root = process.cwd();

async function readJson(path) {
	return JSON.parse(await readFile(resolve(root, path), "utf8"));
}

const settings = await readJson(".pi/settings.json");
assert.deepEqual(settings.packages, [EXPECTED_SOURCE], "project settings must register only the exact reviewed workflow-engine package");

const lock = await readJson(".pi/npm/package-lock.json");
assert.equal(lock.lockfileVersion, 3, "workflow-engine install must use npm lockfile v3");
assert.equal(lock.packages?.[""]?.dependencies?.["pi-workflow-engine"], EXPECTED_VERSION, "root dependency must stay exact");
const locked = lock.packages?.["node_modules/pi-workflow-engine"];
assert.equal(locked?.version, EXPECTED_VERSION, "locked workflow-engine version drifted");
assert.equal(locked?.resolved, EXPECTED_RESOLVED, "locked workflow-engine tarball drifted");
assert.equal(locked?.integrity, EXPECTED_INTEGRITY, "locked workflow-engine integrity drifted");

const installed = await readJson(".pi/npm/node_modules/pi-workflow-engine/package.json");
assert.equal(installed.name, "pi-workflow-engine");
assert.equal(installed.version, EXPECTED_VERSION, "installed workflow-engine version drifted");
assert.deepEqual(
	installed.pi?.extensions,
	[".pi/extensions/pi-workflow-engine/index.ts"],
	"installed workflow-engine public extension entry drifted",
);

console.log(`verified pi-workflow-engine ${EXPECTED_VERSION} provenance and public extension entry`);
