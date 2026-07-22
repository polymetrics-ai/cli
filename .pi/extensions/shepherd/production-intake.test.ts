import assert from "node:assert/strict";
import { lstat, mkdir, mkdtemp, rm, symlink, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import type { ProductionParentPlanDocument } from "./autonomous-production-contract.ts";
import { ProductionRepositoryPlanIntake } from "./production-intake.ts";

function plan(): ProductionParentPlanDocument {
	return {
		schemaVersion: 2,
		planId: "plan-479",
		parentIssue: 479,
		repository: "owner/repo",
		title: "Production Shepherd",
		objective: "Run every stage",
		parentBranch: "feat/shepherd",
		parentBaseBranch: "main",
		actorAllowlist: ["maintainer"],
		decisionExpiresAt: "2026-08-01T00:00:00Z",
		children: [{
			id: "lane-a",
			issue: 501,
			title: "Lane A",
			task: "Implement A",
			slug: "lane-a",
			dependsOn: [],
			access: "mutating",
			writeScopes: ["owned/a"],
			requiredSkills: ["javascript-testing-patterns"],
			verification: [{
				id: "tests",
				executable: "node",
				args: ["--test", "owned/a.test.ts"],
				cwd: ".",
				timeoutMs: 30_000,
				maxOutputBytes: 1_048_576,
			}],
			humanGates: [],
			maxAttempts: 2,
			maxCorrections: 1,
		}],
	};
}

test("production intake returns a canonical exact-plan digest and honors cancellation", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-production-intake-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	await mkdir(join(root, ".planning", "shepherd"), { recursive: true });
	await writeFile(join(root, ".planning", "shepherd", "issue-479.json"), JSON.stringify(plan()));
	const intake = new ProductionRepositoryPlanIntake(root);
	const snapshot = await intake.load(479, new AbortController().signal);
	assert.equal(snapshot.plan.parentIssue, 479);
	assert.match(snapshot.digest, /^[0-9a-f]{64}$/);

	const cancelled = new AbortController();
	cancelled.abort(new Error("stop"));
	await assert.rejects(intake.load(479, cancelled.signal), /stop/);
});

test("production intake rejects a symlink plan and read-only top-level work", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-production-intake-hostile-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	await mkdir(join(root, ".planning", "shepherd"), { recursive: true });
	const outside = join(root, "outside.json");
	await writeFile(outside, JSON.stringify(plan()));
	await symlink(outside, join(root, ".planning", "shepherd", "issue-479.json"));
	const intake = new ProductionRepositoryPlanIntake(root);
	await assert.rejects(intake.load(479, new AbortController().signal), /non-symlink/);

	await rm(join(root, ".planning", "shepherd", "issue-479.json"));
	const hostile = plan() as unknown as { children: Array<Record<string, unknown>> };
	hostile.children[0].access = "read_only";
	await writeFile(join(root, ".planning", "shepherd", "issue-479.json"), JSON.stringify(hostile));
	await assert.rejects(intake.load(479, new AbortController().signal), /must be mutating/);
});

test("missing production intake explains the exact initialization file without mutating the repository", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-production-intake-missing-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const intake = new ProductionRepositoryPlanIntake(root);
	await assert.rejects(
		intake.load(479, new AbortController().signal),
		/production plan is unavailable; create \.planning\/shepherd\/issue-479\.json/,
	);
	await assert.rejects(
		lstat(join(root, ".planning")),
		(error) => typeof error === "object" && error !== null && "code" in error && error.code === "ENOENT",
	);
});
