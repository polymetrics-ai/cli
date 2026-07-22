import assert from "node:assert/strict";
import test from "node:test";

import {
	productionPlanDigest,
	validateProductionParentPlan,
	type ProductionParentPlanDocument,
} from "./autonomous-production-contract.ts";

function plan(): ProductionParentPlanDocument {
	return {
		schemaVersion: 2,
		planId: "plan-479",
		parentIssue: 479,
		repository: "polymetrics/polymetrics",
		title: "Production Shepherd",
		objective: "Complete the autonomous lifecycle",
		parentBranch: "feat/471-pi-agent-session-shepherd",
		parentBaseBranch: "main",
		actorAllowlist: ["maintainer"],
		decisionExpiresAt: "2026-08-01T00:00:00.000Z",
		children: [{
			id: "state",
			issue: 501,
			title: "State",
			task: "Implement durable state",
			slug: "durable-state",
			dependsOn: [],
			access: "mutating",
			writeScopes: [".pi/extensions/shepherd"],
			requiredSkills: ["javascript-testing-patterns"],
			verification: [{
				id: "state-tests",
				executable: "node",
				args: ["--test", ".pi/extensions/shepherd/autonomous-production-state.test.ts"],
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

test("production plan is canonical, stable, mutating, bounded, and shell-free", () => {
	const validated = validateProductionParentPlan(plan(), 479);
	assert.equal(productionPlanDigest(validated), productionPlanDigest(structuredClone(validated)));
	assert.equal(validated.children[0].access, "mutating");
	assert.equal(validated.children[0].verification[0].executable, "node");
	assert.deepEqual(validated.children[0].verification[0].args.slice(0, 1), ["--test"]);
});

test("top-level read-only work, shell commands, traversal, controls, and hostile records fail closed", () => {
	const readOnly = plan() as unknown as { children: Array<Record<string, unknown>> };
	readOnly.children[0].access = "read_only";
	assert.throws(() => validateProductionParentPlan(readOnly), /must be mutating/);

	const shell = plan();
	shell.children[0].verification[0].executable = "/bin/sh";
	assert.throws(() => validateProductionParentPlan(shell), /allowlistable/);

	const traversal = plan();
	traversal.children[0].writeScopes = ["../outside"];
	assert.throws(() => validateProductionParentPlan(traversal), /inside the worktree/);

	const controls = plan();
	controls.children[0].task = "unsafe\u001b[31m";
	assert.throws(() => validateProductionParentPlan(controls), /safe text/);

	const accessor = plan() as unknown as Record<string, unknown>;
	Object.defineProperty(accessor, "title", { enumerable: true, get: () => "surprise" });
	assert.throws(() => validateProductionParentPlan(accessor), /shape|data propert|plain data|exact/i);
});
