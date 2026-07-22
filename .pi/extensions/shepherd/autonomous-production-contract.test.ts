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

test("verification recipes reject arbitrary code and default-branch mutation before orchestration", () => {
	for (const [executable, args] of [
		["node", ["-e", "require('node:child_process').execSync('rm -rf .')"]],
		["git", ["push", "origin", "HEAD:main", "--force"]],
		["go", ["env", "-w", "GOPROXY=https://hostile.invalid"]],
		["make", ["deploy"]],
	] as const) {
		const unsafe = plan();
		unsafe.children[0]!.verification[0] = {
			...unsafe.children[0]!.verification[0]!, executable, args: [...args],
		};
		assert.throws(() => validateProductionParentPlan(unsafe), /verification recipe|test recipe/i);
	}
});

test("verification identifiers and output bounds are executable by the downstream host verifier", () => {
	const unsafeId = plan();
	unsafeId.children[0]!.verification[0]!.id = "unit tests";
	assert.throws(() => validateProductionParentPlan(unsafeId), /verification ID/i);

	const undersizedOutput = plan();
	undersizedOutput.children[0]!.verification[0]!.maxOutputBytes = 1;
	assert.throws(() => validateProductionParentPlan(undersizedOutput), /verification output limit/i);
});

test("every production child declares at least one implementation skill", () => {
	const missingSkills = plan();
	missingSkills.children[0]!.requiredSkills = [];
	assert.throws(() => validateProductionParentPlan(missingSkills), /required skills/i);
});

test("plan intake enforces GitHub-inline fields, scheduler scopes, unique slugs, and unique verification IDs", () => {
	for (const mutate of [
		(value: ProductionParentPlanDocument) => { value.title = " padded"; },
		(value: ProductionParentPlanDocument) => { value.objective = "line one\nline two"; },
		(value: ProductionParentPlanDocument) => { value.children[0]!.title = "Lane "; },
		(value: ProductionParentPlanDocument) => { value.children[0]!.task = "line one\nline two"; },
		(value: ProductionParentPlanDocument) => { value.children[0]!.writeScopes = ["owned/./lane"]; },
		(value: ProductionParentPlanDocument) => { value.children[0]!.writeScopes = ["owned/*"]; },
	]) {
		const invalid = plan();
		mutate(invalid);
		assert.throws(() => validateProductionParentPlan(invalid), /inline|scope|safe text/i);
	}

	const duplicateVerification = plan();
	duplicateVerification.children[0]!.verification.push({
		...duplicateVerification.children[0]!.verification[0]!,
		args: ["--test", ".pi/extensions/shepherd/other.test.ts"],
	});
	assert.throws(() => validateProductionParentPlan(duplicateVerification), /duplicate verification/i);

	const duplicateSlug = plan();
	duplicateSlug.children.push({
		...structuredClone(duplicateSlug.children[0]!),
		id: "pipeline",
		issue: 502,
		writeScopes: [".pi/extensions/shepherd/pipeline"],
	});
	assert.throws(() => validateProductionParentPlan(duplicateSlug), /duplicate child slug/i);
});
