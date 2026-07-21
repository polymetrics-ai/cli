import assert from "node:assert/strict";
import { mkdtemp, readFile, rm, stat, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import { FileStateStore, sanitizeSummary } from "./state-store.ts";

function runState() {
	return {
		schemaVersion: 1,
		issue: 471,
		pr: 438,
		runId: "run-1",
		generation: 1,
		status: "completed",
		candidateHead: "a".repeat(40),
		validationNonce: "nonce-1234567890",
		createdAt: "2026-07-21T08:00:00Z",
		updatedAt: "2026-07-21T08:01:00Z",
		lanes: [],
	};
}

test("atomically persists mode-0600 state and reloads it", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-store-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const store = new FileStateStore(root);
	await store.save(runState());
	assert.deepEqual(await store.load(471), runState());
	const mode = (await stat(join(root, "issue-471.json"))).mode & 0o777;
	assert.equal(mode, 0o600);
	assert.equal((await readFile(join(root, "issue-471.json"), "utf8")).endsWith("\n"), true);
});

test("fails closed on malformed or identity-mismatched state", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-store-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const store = new FileStateStore(root);
	await writeFile(join(root, "issue-471.json"), "not-json", { mode: 0o600 });
	await assert.rejects(store.load(471), /invalid Shepherd state/);
	await writeFile(join(root, "issue-471.json"), JSON.stringify({ ...runState(), issue: 472 }), { mode: 0o600 });
	await assert.rejects(store.load(471), /issue identity mismatch/);
});

test("bounds and redacts summaries before persistence", () => {
	const secret = "ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890";
	const input = `Authorization: Bearer bearer-value token=plain-secret ${secret} ${"x".repeat(2000)}`;
	const output = sanitizeSummary(input, 512);
	assert.ok(output.length <= 512);
	assert.doesNotMatch(output, /bearer-value|plain-secret|ABCDEFGHIJKLMNOPQRSTUVWXYZ/);
	assert.match(output, /\[REDACTED\]/);
});
