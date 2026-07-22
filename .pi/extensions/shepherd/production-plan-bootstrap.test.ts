import assert from "node:assert/strict";
import { mkdir, mkdtemp, readFile, realpath, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import type { AgentSessionHandoff, RoleRunRequest } from "./agent-session-runtime.ts";
import { createParentOrchestrationPlan, type ParentOrchestrationPlan } from "./github-orchestrator.ts";
import { createRequiredGitHubCheckPolicy } from "./github-evidence.ts";
import { ProductionRepositoryPlanIntake } from "./production-intake.ts";
import { createToolPolicy } from "./tool-policy.ts";
import {
	AgentSessionProductionPlanSession,
	ProductionPlanBootstrapper,
	validateProductionParentPlanProposal,
	type ProductionParentPlanProposal,
	type ProductionPlanningIssueFacts,
} from "./production-plan-bootstrap.ts";

const HEAD = "a".repeat(40);
const REVISION = "b".repeat(64);

function facts(): ProductionPlanningIssueFacts {
	return {
		schemaVersion: 1,
		repository: "acme/widgets",
		defaultBranch: "main",
		viewer: { login: "maintainer", permission: "maintain" },
		parent: {
			number: 479,
			nodeId: "I_parent",
			title: "Build production Shepherd",
			body: "Implement the issue through independent child lanes.",
			state: "open",
			updatedAt: "2026-07-22T10:00:00.000Z",
		},
		subissues: [{
			number: 901,
			title: "Existing design evidence",
			body: "Use this as planning evidence only.",
			state: "open",
			updatedAt: "2026-07-22T10:01:00.000Z",
		}],
		complete: true,
		revisionDigest: REVISION,
		observedAt: "2026-07-22T10:02:00.000Z",
	};
}

function proposal(overrides: Partial<ProductionParentPlanProposal> = {}): ProductionParentPlanProposal {
	return {
		schemaVersion: 1,
		sourceRevisionDigest: REVISION,
		title: "Production Shepherd",
		objective: "Implement the parent issue through bounded parallel child work.",
		children: [{
			id: "lane-a",
			title: "Implement lane A",
			task: "Implement the first isolated production slice.",
			slug: "lane-a",
			dependsOn: [],
			access: "mutating",
			writeScopes: [".pi/extensions/shepherd/lane-a"],
			requiredSkills: ["javascript-testing-patterns"],
			verification: [{
				id: "tests",
				executable: "node",
				args: ["--test", ".pi/extensions/shepherd/lane-a.test.ts"],
				cwd: ".",
				timeoutMs: 30_000,
				maxOutputBytes: 65_536,
			}],
			humanGates: [],
			maxAttempts: 2,
			maxCorrections: 1,
		}],
		...overrides,
	};
}

function authority() {
	return {
		repository: "acme/widgets",
		parentIssue: 479,
		parentBranch: "feat/471-parent",
		parentBaseBranch: "main",
		candidateHead: HEAD,
	};
}

function handoff(request: RoleRunRequest): AgentSessionHandoff {
	return {
		schemaVersion: 1,
		...request.binding,
		role: request.role,
		status: "completed",
		summary: "submitted one bounded semantic plan",
		observedMutation: false,
		changedPaths: [],
		verification: [],
		findings: [],
	};
}

test("missing issue plan uses one xhigh planning session and materializes returned GitHub issue numbers", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-plan-bootstrap-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const intake = new ProductionRepositoryPlanIntake(root);
	const calls: string[] = [];
	const bootstrap = new ProductionPlanBootstrapper({
		repositoryRoot: root,
		stateRoot: join(root, "state"),
		intake,
		issueSource: {
			async observe() { calls.push("observe"); return facts(); },
		},
		authoritySource: {
			async observe() { calls.push("authority"); return authority(); },
		},
		planSession: {
			async propose() { calls.push("plan"); return proposal(); },
		},
		github: {
			async createPlan(value) {
				calls.push("create-plan");
				const objective = value as { children: Array<Record<string, unknown>> };
				assert.equal(Object.hasOwn(objective.children[0]!, "issue"), false);
				return createParentOrchestrationPlan(value, {
					schemaVersion: 1,
					requiredCheckPolicies: ["feat/471-parent", "main"].map((baseBranch, index) =>
						createRequiredGitHubCheckPolicy({
							schemaVersion: 1,
							repository: "acme/widgets",
							baseBranch,
							revision: index + 1,
							requiredChecks: [{ name: "test", producerId: "github-actions" }],
						})),
				});
			},
			async ensureChildIssue(_plan, childId) {
				calls.push(`issue:${childId}`);
				return { number: 731 };
			},
			async stop() {},
		},
		now: () => new Date("2026-07-22T10:03:00.000Z"),
	});

	const snapshot = await bootstrap.ensure(479, new AbortController().signal);
	assert.equal(snapshot.plan.children[0]?.issue, 731);
	assert.equal(snapshot.path, join(await realpath(root), ".planning", "shepherd", "issue-479.json"));
	assert.deepEqual(calls, ["observe", "authority", "plan", "create-plan", "issue:lane-a"]);
	assert.equal((JSON.parse(await readFile(snapshot.path, "utf8")) as { children: Array<{ issue: number }> }).children[0]?.issue, 731);

	const reused = await bootstrap.ensure(479, new AbortController().signal);
	assert.equal(reused.digest, snapshot.digest);
	assert.deepEqual(calls, ["observe", "authority", "plan", "create-plan", "issue:lane-a"]);
});

test("bootstrap never replaces an invalid existing human plan and creates no GitHub effect", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-plan-bootstrap-invalid-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	await mkdir(join(root, ".planning", "shepherd"), { recursive: true });
	const path = join(root, ".planning", "shepherd", "issue-479.json");
	await writeFile(path, "{invalid", { mode: 0o600 });
	let effects = 0;
	const bootstrap = new ProductionPlanBootstrapper({
		repositoryRoot: root,
		stateRoot: join(root, "state"),
		intake: new ProductionRepositoryPlanIntake(root),
		issueSource: { async observe() { effects += 1; return facts(); } },
		authoritySource: { async observe() { effects += 1; return authority(); } },
		planSession: { async propose() { effects += 1; return proposal(); } },
		github: {
			async createPlan(value) { effects += 1; return value as ParentOrchestrationPlan; },
			async ensureChildIssue() { effects += 1; return { number: 731 }; },
			async stop() {},
		},
	});
	await assert.rejects(bootstrap.ensure(479, new AbortController().signal), /not valid JSON/);
	assert.equal(effects, 0);
	assert.equal(await readFile(path, "utf8"), "{invalid");
});

test("proposal validation rejects issue-number invention and dependency cycles before GitHub mutation", () => {
	assert.throws(
		() => validateProductionParentPlanProposal({
			...proposal(),
			children: [{ ...proposal().children[0], issue: 999 }],
		}, facts(), authority()),
		/exact|field|issue/i,
	);
	const first = proposal().children[0]!;
	assert.throws(
		() => validateProductionParentPlanProposal({
			...proposal(),
			children: [
				{ ...first, id: "a", slug: "a", dependsOn: ["b"] },
				{ ...first, id: "b", slug: "b", dependsOn: ["a"], writeScopes: ["other"] },
			],
		}, facts(), authority()),
		/cycle/i,
	);
	const commandMutation = proposal();
	commandMutation.children[0]!.verification[0] = {
		...commandMutation.children[0]!.verification[0]!,
		executable: "git",
		args: ["push", "origin", "HEAD:main", "--force"],
	};
	assert.throws(
		() => validateProductionParentPlanProposal(commandMutation, facts(), authority()),
		/verification recipe|test recipe/i,
	);
	for (const verification of [
		{ ...proposal().children[0]!.verification[0]!, id: "unit tests" },
		{ ...proposal().children[0]!.verification[0]!, maxOutputBytes: 1 },
	]) {
		const unusable = proposal();
		unusable.children[0]!.verification[0] = verification;
		assert.throws(
			() => validateProductionParentPlanProposal(unusable, facts(), authority()),
			/verification (?:ID|output limit)/i,
		);
	}
	const duplicateVerification = proposal();
	duplicateVerification.children[0]!.verification.push({
		...duplicateVerification.children[0]!.verification[0]!,
		args: ["--test", ".pi/extensions/shepherd/other.test.ts"],
	});
	assert.throws(
		() => validateProductionParentPlanProposal(duplicateVerification, facts(), authority()),
		/duplicate verification/i,
	);
	const invalidScope = proposal();
	invalidScope.children[0]!.writeScopes = ["owned/./lane"];
	assert.throws(
		() => validateProductionParentPlanProposal(invalidScope, facts(), authority()),
		/scope/i,
	);
	const duplicateSlug = proposal();
	duplicateSlug.children.push({
		...structuredClone(duplicateSlug.children[0]!),
		id: "lane-b",
		writeScopes: ["owned/lane-b"],
	});
	assert.throws(
		() => validateProductionParentPlanProposal(duplicateSlug, facts(), authority()),
		/duplicate child slug/i,
	);
});

test("planning AgentSession can only submit one validated issue-less proposal through host_inspect", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-plan-session-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	await writeFile(join(root, "README.md"), "repository context");
	let requestSeen: RoleRunRequest | undefined;
	const session = new AgentSessionProductionPlanSession({
		repositoryRoot: root,
		agentSession: {
			async run(request) {
				requestSeen = request;
				assert.equal(request.role, "planning");
				assert.equal(request.authority.readOnly, true);
				assert.deepEqual(request.authority.capabilityNames, ["host_inspect"]);
				assert.equal(request.capabilities.length, 1);
				assert.equal(request.capabilities[0]?.name, "host_inspect");
				const policy = createToolPolicy({
					readOnly: true,
					workspace: request.workspace,
					authority: {
						workspaceId: request.authority.workspaceId,
						readPrefixes: request.authority.readPrefixes,
						writePrefixes: request.authority.writePrefixes,
						capabilityNames: request.authority.capabilityNames,
					},
					capabilities: request.capabilities,
				});
				assert.deepEqual(policy.names, ["workspace_read", "host_inspect"]);
				const result = await request.capabilities[0]!.execute(proposal(), request.signal);
				assert.equal(result.status, "ok");
				return handoff(request);
			},
			async abort() {},
		},
	});
	const result = await session.propose({
		facts: facts(),
		authority: authority(),
	}, {
		signal: new AbortController().signal,
		deadlineAt: "2026-07-22T11:00:00.000Z",
	});
	assert.deepEqual(result, proposal());
	assert.ok(requestSeen);
	assert.match(requestSeen.task, /submit/i);
});
