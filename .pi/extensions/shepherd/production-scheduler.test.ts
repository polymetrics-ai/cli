import assert from "node:assert/strict";
import test from "node:test";

import type { ProductionParentPlanDocument } from "./autonomous-production-contract.ts";
import { createProductionAutonomousState } from "./autonomous-production-state.ts";
import { selectProductionChildren } from "./production-scheduler.ts";

function plan(): ProductionParentPlanDocument {
	const child = (id: string, issue: number, dependsOn: string[], writeScopes: string[]) => ({
		id,
		issue,
		title: id,
		task: `implement ${id}`,
		slug: id,
		dependsOn,
		access: "mutating" as const,
		writeScopes,
		requiredSkills: ["javascript-testing-patterns"],
		verification: [{ id: `${id}-test`, executable: "node", args: ["--test", `${id}.test.ts`], cwd: ".", timeoutMs: 30_000, maxOutputBytes: 1_000_000 }],
		humanGates: [],
		maxAttempts: 2,
		maxCorrections: 1,
	});
	return {
		schemaVersion: 2,
		planId: "scheduler",
		parentIssue: 479,
		repository: "owner/repo",
		title: "scheduler",
		objective: "schedule",
		parentBranch: "feat/parent",
		parentBaseBranch: "main",
		actorAllowlist: ["maintainer"],
		decisionExpiresAt: "2026-08-01T00:00:00.000Z",
		children: [
			child("alpha", 501, [], ["owned/shared"]),
			child("beta", 502, [], ["owned/shared/beta"]),
			child("gamma", 503, [], ["owned/gamma"]),
			child("delta", 504, ["alpha"], ["owned/delta"]),
		],
	};
}

test("selects the maximum deterministic disjoint set within capacity", () => {
	const state = createProductionAutonomousState(plan(), { runId: "run", maxConcurrency: 2 });
	assert.deepEqual(selectProductionChildren(state), { kind: "dispatch", childIds: ["alpha", "gamma"] });
});

test("explains capacity, write-scope collision, dependencies, and completion without mutating state", () => {
	const state = createProductionAutonomousState(plan(), { runId: "run", maxConcurrency: 1 });
	state.children.find((child) => child.id === "alpha")!.status = "running";
	state.children.find((child) => child.id === "alpha")!.stage = "implementation";
	state.children.find((child) => child.id === "alpha")!.attempts = 1;
	assert.deepEqual(selectProductionChildren(state), { kind: "idle", reason: "capacity" });

	state.maxConcurrency = 2;
	state.children.find((child) => child.id === "gamma")!.status = "blocked";
	assert.deepEqual(selectProductionChildren(state), { kind: "idle", reason: "write_scope_collision" });

	state.children.find((child) => child.id === "alpha")!.status = "failed";
	state.children.find((child) => child.id === "beta")!.status = "blocked";
	assert.deepEqual(selectProductionChildren(state), { kind: "idle", reason: "dependencies" });

	for (const child of state.children) {
		child.status = "succeeded";
		child.stage = "succeeded";
		child.checkpoint = { summary: "integrated", integrationReceiptDigest: "a".repeat(64) };
	}
	assert.deepEqual(selectProductionChildren(state), { kind: "complete" });
});
