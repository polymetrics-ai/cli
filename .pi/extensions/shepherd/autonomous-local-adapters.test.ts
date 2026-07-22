import assert from "node:assert/strict";
import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import { LocalParentMergeGate, RepositoryManifestIntake } from "./autonomous-local-adapters.ts";

test("loads the repository MVP manifest and keeps the parent merge as a local human wait", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-manifest-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	await mkdir(join(root, ".planning", "shepherd"), { recursive: true });
	await writeFile(join(root, ".planning", "shepherd", "issue-471.json"), JSON.stringify({
		schemaVersion: 1,
		parentIssue: 471,
		planId: "mvp-plan",
		children: [{
			id: "help",
			issue: 501,
			title: "Help",
			task: "Implement help and tests.",
			dependsOn: [],
			access: "mutating",
			writeScopes: ["cmd/help"],
		}],
	}));
	const plan = await new RepositoryManifestIntake(root).load(471, new AbortController().signal);
	assert.equal(plan.planId, "mvp-plan");
	assert.deepEqual(plan.children.map((child) => child.id), ["help"]);
	const state = {
		schemaVersion: 2,
		kind: "autonomous",
		issue: 471,
		planId: plan.planId,
		runId: "run-1",
		generation: 1,
		status: "running",
		stage: "HUMAN_DECISION",
		maxConcurrency: 2,
		timeoutMs: 900_000,
		createdAt: "2026-07-22T08:00:00Z",
		updatedAt: "2026-07-22T08:01:00Z",
		children: [],
	};
	const gate = new LocalParentMergeGate();
	assert.deepEqual(await gate.request(state, new AbortController().signal), {
		requestId: "local-parent-merge-471-1",
	});
	assert.equal(await gate.observe(state, new AbortController().signal), "pending");
	assert.equal("merge" in gate, false);
});
