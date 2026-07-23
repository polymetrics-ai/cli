import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import { dirname, join } from "node:path";

const EXPECTED_VERSION = "0.80.10";
const prefix = dirname(dirname(process.execPath));
const packageRoot = process.env.SHEPHERD_PI_PACKAGE_ROOT ?? join(
	prefix,
	"lib",
	"node_modules",
	"@earendil-works",
	"pi-coding-agent",
);

async function readJson(path) {
	return JSON.parse(await readFile(path, "utf8"));
}

const installedPackages = new Map([
	["pi-coding-agent", join(packageRoot, "package.json")],
	["pi-agent-core", join(packageRoot, "node_modules", "@earendil-works", "pi-agent-core", "package.json")],
	["pi-ai", join(packageRoot, "node_modules", "@earendil-works", "pi-ai", "package.json")],
	["pi-tui", join(packageRoot, "node_modules", "@earendil-works", "pi-tui", "package.json")],
]);

for (const [name, path] of installedPackages) {
	const manifest = await readJson(path);
	assert.equal(manifest.version, EXPECTED_VERSION, `${name} must be exactly ${EXPECTED_VERSION}`);
}

const shrinkwrap = await readJson(join(packageRoot, "npm-shrinkwrap.json"));
assert.equal(shrinkwrap.lockfileVersion, 3, "Pi runtime must publish an npm v3 shrinkwrap");
assert.equal(shrinkwrap.packages[""].version, EXPECTED_VERSION);

for (const name of ["pi-agent-core", "pi-ai", "pi-tui"]) {
	const packagePath = `node_modules/@earendil-works/${name}`;
	const locked = shrinkwrap.packages[packagePath];
	assert.ok(locked, `shrinkwrap must contain ${name}`);
	assert.equal(locked.version, EXPECTED_VERSION, `${name} shrinkwrap version must be exact`);
	assert.equal(
		locked.resolved,
		`https://registry.npmjs.org/@earendil-works/${name}/-/${name}-${EXPECTED_VERSION}.tgz`,
		`${name} must resolve to the exact published tarball`,
	);
}

console.log(`verified exact Pi runtime family ${EXPECTED_VERSION}`);
