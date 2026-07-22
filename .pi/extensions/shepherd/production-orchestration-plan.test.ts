import assert from "node:assert/strict";
import test from "node:test";

import type { ProductionParentPlanDocument } from "./autonomous-production-contract.ts";
import { createRequiredGitHubCheckPolicy } from "./github-evidence.ts";
import { createProductionOrchestrationPlan } from "./production-orchestration-plan.ts";

function plan(): ProductionParentPlanDocument {
	return {
		schemaVersion: 2,
		planId: "plan-479",
		parentIssue: 479,
		repository: "owner/repo",
		title: "Production Shepherd",
		objective: "Run production Shepherd",
		parentBranch: "feat/shepherd",
		parentBaseBranch: "main",
		actorAllowlist: ["maintainer"],
		decisionExpiresAt: "2026-08-01T00:00:00.000Z",
		children: [{
			id: "lane",
			issue: 501,
			title: "Lane",
			task: "Implement the lane",
			slug: "lane",
			dependsOn: [],
			access: "mutating",
			writeScopes: ["owned/lane"],
			requiredSkills: ["javascript-testing-patterns"],
			verification: [
				{ id: "tests", executable: "node", args: ["--test", "lane.test.ts"], cwd: ".", timeoutMs: 30_000, maxOutputBytes: 1_000_000 },
				{ id: "typecheck", executable: "node", args: ["tsc.js", "--noEmit"], cwd: ".", timeoutMs: 30_000, maxOutputBytes: 1_000_000 },
			],
			humanGates: ["review"],
			maxAttempts: 2,
			maxCorrections: 1,
		}],
	};
}

test("maps the production plan to the canonical stacked-PR orchestration plan", () => {
	const policies = ["feat/shepherd", "main"].map((baseBranch, index) => createRequiredGitHubCheckPolicy({
		schemaVersion: 1,
		repository: "owner/repo",
		baseBranch,
		revision: index + 1,
		requiredChecks: [{ name: "test", producerId: "github-actions" }],
	}));
	const result = createProductionOrchestrationPlan(plan(), 3, policies);
	assert.equal(result.generation, 3);
	assert.equal(result.children[0].access, "mutating");
	assert.equal(result.children[0].prBase, "feat/shepherd");
	assert.deepEqual(result.children[0].verification.map((item) => item.kind), ["test", "typecheck"]);
	assert.match(result.canonical.digest, /^[0-9a-f]{64}$/);
});
