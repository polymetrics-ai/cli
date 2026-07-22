import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import {
	createProductionAutonomousState,
	type ProductionAutonomousState,
} from "./autonomous-production-state.ts";
import {
	ProductionEffectJournal,
} from "./autonomous-effect-journal.ts";
import {
	ProductionLifecycleError,
	type ProductionParentPlanDocument,
	type ProductionStageCheckpoint,
	type ProductionWorkspaceBinding,
} from "./autonomous-production-contract.ts";
import type {
	ProductionChildPipelineContext,
} from "./production-controller.ts";
import {
	createRequiredGitHubCheckPolicy,
	type GitHubPullRequestEvidence,
} from "./github-evidence.ts";
import {
	createCanonicalPullRequestSnapshot,
	type ChildIntegrationReceipt,
	type ChildIntegrationDecision,
	type GitHubChildIssue,
	type MaterializedChildRecord,
	type OrchestrationCallContext,
	type ParentOrchestrationPlan,
} from "./github-orchestrator.ts";
import {
	createHumanDecisionRecord,
	type HumanDecisionRecord,
} from "./human-decision.ts";
import {
	createProductionOrchestrationPlan,
} from "./production-orchestration-plan.ts";
import {
	MemoryProductionReviewRepository,
	type ProductionReviewArtifact,
	type ProductionReviewRepository,
} from "./production-review-adapter.ts";
import {
	createAgentSessionAttestation,
	createIndependentReviewWork,
	type IndependentReviewFinding,
	type IndependentReviewTarget,
} from "./review-router.ts";
import type {
	ProductionWorkspaceSession,
} from "./production-workspace-lifecycle.ts";
import {
	ProductionChildPipeline,
	productionChildIntegrationReceiptDigest,
	type ProductionChildGitHubPort,
	type ProductionChildPipelineOptions,
	type ProductionExactHeadReviewPort,
	type ProductionParentHeadSource,
	type ProductionWorkspaceLifecyclePort,
} from "./production-child-pipeline.ts";

const SHA_A = "a".repeat(40);
const SHA_B = "b".repeat(40);
const SHA_C = "c".repeat(40);
const SHA_D = "d".repeat(40);
const DIGEST = "e".repeat(64);

function plan(): ProductionParentPlanDocument {
	return {
		schemaVersion: 2,
		planId: "production-479",
		parentIssue: 479,
		repository: "owner/repo",
		title: "Production Shepherd",
		objective: "Run one production child",
		parentBranch: "feat/471-parent",
		parentBaseBranch: "main",
		actorAllowlist: ["maintainer"],
		decisionExpiresAt: "2026-08-01T00:00:00.000Z",
		children: [{
			id: "child",
			issue: 501,
			title: "Production child",
			task: "Implement the child",
			slug: "production-child",
			dependsOn: [],
			access: "mutating",
			writeScopes: ["owned/child"],
			requiredSkills: ["javascript-testing-patterns"],
			verification: [{
				id: "tests",
				executable: "node",
				args: ["--test", "owned/child.test.ts"],
				cwd: ".",
				timeoutMs: 30_000,
				maxOutputBytes: 1_000_000,
			}],
			humanGates: ["review"],
			maxAttempts: 2,
			maxCorrections: 1,
		}],
	};
}

function orchestrationPlan(value = plan()): ParentOrchestrationPlan {
	return createProductionOrchestrationPlan(value, 1, [
		createRequiredGitHubCheckPolicy({
			schemaVersion: 1,
			repository: value.repository,
			baseBranch: value.parentBranch,
			revision: 1,
			requiredChecks: [{ name: "tests", producerId: "github-actions" }],
		}),
		createRequiredGitHubCheckPolicy({
			schemaVersion: 1,
			repository: value.repository,
			baseBranch: value.parentBaseBranch,
			revision: 1,
			requiredChecks: [{ name: "tests", producerId: "github-actions" }],
		}),
	]);
}

function binding(overrides: Partial<ProductionWorkspaceBinding> = {}): ProductionWorkspaceBinding {
	return {
		claimId: "claim-501",
		ownershipId: `production:${"1".repeat(64)}`,
		repositoryIdentity: "repository-identity",
		worktreeIdentity: "worktree-identity",
		cwd: "/trusted/worktrees/issue-501-production-child",
		branch: "feat/501-production-child",
		baseBranch: "feat/471-parent",
		baseHead: SHA_A,
		head: SHA_A,
		writeScopes: ["owned/child"],
		...overrides,
	};
}

function checkpointMerge(
	previous: ProductionStageCheckpoint | undefined,
	next: ProductionStageCheckpoint,
): ProductionStageCheckpoint {
	const effectKeys = [...new Set([
		...(previous?.effectKey === undefined ? [] : [previous.effectKey]),
		...(previous?.effectKeys ?? []),
		...(next.effectKey === undefined ? [] : [next.effectKey]),
		...(next.effectKeys ?? []),
	])];
	return {
		...(previous ?? { summary: next.summary }),
		...next,
		...(effectKeys.length === 0 ? {} : { effectKeys }),
		...(next.workspace === undefined && previous?.workspace !== undefined ? { workspace: previous.workspace } : {}),
		...(next.verification === undefined && previous?.verification !== undefined ? { verification: previous.verification } : {}),
		...(next.pullRequest === undefined && previous?.pullRequest !== undefined ? { pullRequest: previous.pullRequest } : {}),
		...(next.review === undefined && previous?.review !== undefined ? { review: previous.review } : {}),
	};
}

function contextFor(state: ProductionAutonomousState, value = plan()): ProductionChildPipelineContext {
	return {
		plan: value,
		state: structuredClone(state),
		child: structuredClone(value.children[0]),
		runtime: structuredClone(state.children[0]),
		runId: state.runId,
		resourceGeneration: state.resourceGeneration,
		generation: state.generation,
		timeoutMs: state.timeoutMs,
		signal: new AbortController().signal,
	};
}

function persist(state: ProductionAutonomousState, next: ProductionStageCheckpoint): void {
	const child = state.children[0];
	child.checkpoint = checkpointMerge(child.checkpoint, next);
	if (next.workspace !== undefined) child.ownership = structuredClone(next.workspace);
	state.revision += 1;
	state.updatedAt = new Date().toISOString();
}

function childIssue(parent: ParentOrchestrationPlan): GitHubChildIssue {
	const child = parent.children[0];
	return {
		number: 501,
		marker: child.markers.issue,
		title: child.title,
		body: child.issueBody,
		state: "open",
		parentIssue: parent.parentIssue,
	};
}

function pullRequestEvidence(
	parent: ParentOrchestrationPlan,
	child: MaterializedChildRecord,
	handoff: ReturnType<FakeWorkspaceSession["handoff"]>,
	overrides: Partial<GitHubPullRequestEvidence> = {},
): GitHubPullRequestEvidence {
	const observedAt = new Date(Date.now() - 2_000).toISOString();
	return {
		schemaVersion: 2,
		repository: parent.repository,
		workItemId: child.id,
		generation: parent.generation,
		number: 77,
		marker: child.markers.pullRequest,
		title: child.title,
		body: `${child.markers.pullRequest}\nProduction child`,
		state: "open",
		draft: false,
		baseBranch: child.prBase,
		headBranch: child.branch,
		baseSha: handoff.baseHead,
		headSha: handoff.head,
		changedPathsComplete: true,
		changedPaths: [...handoff.changedScope],
		allowedScopes: [...child.writeScopes],
		mergeState: "clean",
		policyDigest: parent.requiredCheckPolicies.find((entry) => entry.baseBranch === child.prBase)!.digest,
		checksComplete: true,
		checks: [],
		requestedChangesComplete: true,
		requestedChanges: [],
		threadsComplete: true,
		threads: [],
		reviews: [],
		reviewsComplete: true,
		dispositionsComplete: true,
		dispositions: [],
		revision: 1,
		observedAt,
		...overrides,
	};
}

function integrationReceipt(
	parent: ParentOrchestrationPlan,
	child: MaterializedChildRecord,
	evidence: GitHubPullRequestEvidence,
): ChildIntegrationReceipt {
	const at = new Date(Date.now() - 1_000).toISOString();
	return {
		childId: child.id,
		pullRequest: evidence.number,
		generation: parent.generation,
		marker: child.markers.pullRequest,
		baseSha: evidence.baseSha,
		headSha: evidence.headSha,
		parentBranch: parent.parentBranch,
		pullRequestSnapshot: createCanonicalPullRequestSnapshot(evidence),
		observation: { revision: evidence.revision, observedAt: evidence.observedAt, state: evidence.state },
		controllerProvenance: {
			authority: "controller",
			planDigest: parent.canonical.digest,
			policyDigest: evidence.policyDigest,
			policyRevision: 1,
			policyObservedAt: evidence.observedAt,
			changedPathDigest: DIGEST,
			reviewAuthorizationDigest: DIGEST,
			reviewResultDigest: DIGEST,
			reviewCompletedAt: evidence.observedAt,
			evidenceRevision: evidence.revision,
			observedAt: evidence.observedAt,
		},
		transportProvenance: {
			authority: "transport",
			idempotencyKey: "integration-501",
			intentDigest: DIGEST,
			revision: 1,
		},
		integratedAt: at,
	};
}

class FakeWorkspaceSession implements ProductionWorkspaceSession {
	#binding = binding();
	#dirty = true;
	#verificationFailures: number;
	readonly calls: string[];

	constructor(calls: string[], verificationFailures = 0) {
		this.calls = calls;
		this.#verificationFailures = verificationFailures;
	}

	get binding(): ProductionWorkspaceBinding { return structuredClone(this.#binding); }

	async implement() {
		this.calls.push("implement");
		return {
			role: "implementation" as const,
			status: "completed" as const,
			runId: "run-479",
			generation: 1,
			laneId: "child-implementation",
			candidateHead: this.#binding.head,
			validationNonce: DIGEST,
			observedMutation: true,
			findings: [],
			summary: "implemented",
		};
	}

	async correct() {
		this.calls.push("correct");
		this.#dirty = true;
		return {
			role: "correction" as const,
			status: "completed" as const,
			runId: "run-479",
			generation: 1,
			laneId: "child-correction",
			candidateHead: this.#binding.head,
			validationNonce: DIGEST,
			observedMutation: true,
			findings: [],
			summary: "corrected",
		};
	}

	async verify() {
		this.calls.push("verify");
		const failed = this.#verificationFailures > 0;
		if (failed) this.#verificationFailures -= 1;
		return [{
			id: "tests",
			executable: "node",
			args: ["--test", "owned/child.test.ts"],
			cwd: ".",
			status: failed ? "failed" as const : "passed" as const,
			exitCode: failed ? 1 : 0,
			signal: null,
			durationMs: 1,
			stdout: "ok",
			stderr: "",
			...(failed ? { failureKind: "exit" as const } : {}),
			outputTruncated: false,
		}];
	}

	async commit() {
		this.calls.push("commit");
		const previousHead = this.#binding.head;
		if (this.#dirty) this.#binding.head = previousHead === SHA_A ? SHA_C : SHA_D;
		const committed = this.#dirty;
		this.#dirty = false;
		return { committed, previousHead, head: this.#binding.head };
	}

	async push() {
		this.calls.push("push");
		return { branch: this.#binding.branch, head: this.#binding.head, remoteName: "origin" as const };
	}

	handoff() {
		return {
			issue: 501,
			branch: this.#binding.branch,
			prBase: this.#binding.baseBranch,
			baseHead: this.#binding.baseHead,
			head: this.#binding.head,
			changedScope: ["owned/child/file.ts"],
			verificationState: "passed" as const,
			repositoryIdentity: this.#binding.repositoryIdentity,
			worktreeIdentity: this.#binding.worktreeIdentity,
			dirty: false,
		};
	}

	async captureHandoff() { this.calls.push("handoff"); return this.handoff(); }

	async refreshParent(request: { previousParentHead: string; newParentHead: string; effectKey: string }) {
		this.calls.push("refresh");
		const previousHead = this.#binding.head;
		const previousBaseHead = this.#binding.baseHead;
		this.#binding.baseHead = request.newParentHead;
		this.#binding.head = SHA_D;
		return {
			outcome: "rebased" as const,
			previousBaseHead,
			baseHead: this.#binding.baseHead,
			previousHead,
			head: this.#binding.head,
			verificationInvalidated: true as const,
			reviewInvalidated: true as const,
		};
	}

	async join() { this.calls.push("join"); }
}

interface HarnessOptions {
	findings?: IndependentReviewFinding[];
	pullRequest?: Partial<GitHubPullRequestEvidence>;
	integration?: ChildIntegrationDecision;
	publicationTimeoutOnce?: boolean;
	verificationFailures?: number;
}

async function harness(t: { after(fn: () => void | Promise<void>): void }, options: HarnessOptions = {}) {
	const root = await mkdtemp(join(tmpdir(), "production-child-pipeline-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const effects = new ProductionEffectJournal(root);
	const calls: string[] = [];
	const session = new FakeWorkspaceSession(calls, options.verificationFailures);
	const lifecycle: ProductionWorkspaceLifecyclePort = {
		async claim() { calls.push("workspace"); return session; },
		async abort(runId) { calls.push(`abort:${runId}`); },
		async close() { calls.push("workspace-close"); },
	};
	const parent = orchestrationPlan();
	let currentParentHead = SHA_A;
	let physicalPullRequests = 0;
	let publicationAttempts = 0;
	let published: GitHubPullRequestEvidence | undefined;
	let receipt: ChildIntegrationReceipt | undefined;
	const github: ProductionChildGitHubPort = {
		async createPlan(_value: unknown, _context?: OrchestrationCallContext) { calls.push("plan"); return parent; },
		async ensureChildIssue(value) { calls.push("issue"); return childIssue(value); },
		async ensureChildPullRequest(value, child, handoff) {
			calls.push("pull-request");
			publicationAttempts += 1;
			if (published === undefined) {
				physicalPullRequests += 1;
				published = pullRequestEvidence(value, child, handoff, options.pullRequest);
				if (options.publicationTimeoutOnce) throw new Error("external operation timed out after publication");
			}
			return published;
		},
		async integrateChild(value, child, handoff) {
			calls.push("integrate");
			if (options.integration !== undefined) return options.integration;
			const evidence = published ?? pullRequestEvidence(value, child, handoff);
			receipt = integrationReceipt(value, child, evidence);
			currentParentHead = SHA_B;
			return { kind: "integrated", receipt, reused: false };
		},
		async stop() { calls.push("github-stop"); return { kind: "joined", active: 0, unacknowledged: 0 }; },
	};
	const reviews: ProductionReviewRepository = new MemoryProductionReviewRepository();
	const reviewer: ProductionExactHeadReviewPort = {
		async review(target, externalContext) {
			calls.push("review");
			const work = createIndependentReviewWork(target);
			const review = {
				...work,
				completedAt: new Date(Date.now() - 500).toISOString(),
				verdict: options.findings?.length ? "findings" as const : "clean" as const,
				findings: options.findings ?? [],
			};
			const attestation = createAgentSessionAttestation({ sessionId: "review-session", runId: "review-run", review });
			return reviews.publish({ schemaVersion: 1, review, attestation, dispositions: [] }, externalContext);
		},
	};
	const parentHeads: ProductionParentHeadSource = {
		async observe() {
			calls.push("parent-head");
			return { repository: "owner/repo", branch: "feat/471-parent", head: currentParentHead };
		},
	};
	const decisions = new Map<string, "pending" | "authorized" | "aborted">();
	const decisionRecords = new Map<string, HumanDecisionRecord>();
	const broker: ProductionChildPipelineOptions["decisionBroker"] = {
		async request(request) {
			calls.push("intervention-request");
			decisions.set(request.requestId, "pending");
			const record = createHumanDecisionRecord({
				requestId: request.requestId,
				gate: request.gate,
				binding: {
					repository: request.repository,
					target: request.gate === "scope"
						? { kind: "issue", number: request.parentIssue }
						: { kind: "pull_request", number: request.pullRequest },
					generation: request.generation,
					...(request.headSha === undefined ? {} : { headSha: request.headSha }),
				},
				actorAllowlist: [...request.actorAllowlist],
				expiresAt: request.expiresAt,
				question: request.question,
				allowedOptions: [...request.allowedOptions],
			}, new Date(Date.now() - 1_000));
			decisionRecords.set(request.requestId, record);
			return record;
		},
		async poll(requestId, binding) {
			calls.push("intervention-poll");
			const status = decisions.get(requestId) ?? "pending";
			const base = decisionRecords.get(requestId);
			if (base === undefined || JSON.stringify(base.binding) !== JSON.stringify(binding)) throw new Error("missing decision record");
			if (status === "pending") return structuredClone(base);
			const targetPath = binding.target.kind === "issue" ? "issues" : "pull";
			return {
				...base,
				status: "decided" as const,
				updatedAt: new Date(Date.now() - 250).toISOString(),
				requestComment: {
					id: 1,
					url: `https://github.com/owner/repo/${targetPath}/${binding.target.number}#issuecomment-1`,
					actor: "shepherd-bot",
					createdAt: new Date(Date.now() - 750).toISOString(),
				},
				decision: {
					option: status === "authorized" ? "authorize-one-retry" : "abort-child",
					actor: "maintainer",
					sourceUrl: `https://github.com/owner/repo/${targetPath}/${binding.target.number}#issuecomment-2`,
					decidedAt: new Date(Date.now() - 500).toISOString(),
				},
			};
		},
		async consume(requestId, binding) {
			calls.push("intervention-consume");
			const status = decisions.get(requestId);
			if (status !== "authorized" && status !== "aborted") throw new Error("decision is not consumable");
			const base = decisionRecords.get(requestId);
			if (base === undefined || JSON.stringify(base.binding) !== JSON.stringify(binding)) throw new Error("missing decision record");
			const targetPath = binding.target.kind === "issue" ? "issues" : "pull";
			return {
				...base,
				status: "consumed",
				requestComment: {
					id: 1,
					url: `https://github.com/owner/repo/${targetPath}/${binding.target.number}#issuecomment-1`,
					actor: "shepherd-bot",
					createdAt: new Date(Date.now() - 750).toISOString(),
				},
				updatedAt: new Date(Date.now() - 250).toISOString(),
				consumedAt: new Date(Date.now() - 250).toISOString(),
				decision: {
					option: status === "authorized" ? "authorize-one-retry" : "abort-child",
					actor: "maintainer",
					sourceUrl: `https://github.com/owner/repo/${targetPath}/${binding.target.number}#issuecomment-2`,
					decidedAt: new Date(Date.now() - 500).toISOString(),
				},
			};
		},
	};
	const recoveryCalls: string[] = [];
	const pipeline = new ProductionChildPipeline({
		workspaceLifecycle: lifecycle,
		github,
		reviewer,
		reviewRepository: reviews,
		effects,
		decisionBroker: broker,
		parentHeads,
		coordinator: {
			cwd: "/trusted/coordinator",
			repositoryIdentity: "repository-identity",
			worktreeIdentity: "coordinator-worktree",
			remoteName: "origin",
			remoteIdentity: "remote-identity",
			fetchEndpointIdentity: "fetch-identity",
			pushEndpointIdentity: "push-identity",
			defaultBranch: "main",
		},
		trustedWorktreeRoot: "/trusted/worktrees",
		dispositionActor: "shepherd-controller",
		recovery: {
			async observe(record) { recoveryCalls.push(`observe:${record.kind}`); return { resultDigest: record.intentDigest }; },
			async apply(record) { recoveryCalls.push(`apply:${record.kind}`); },
		},
		now: () => new Date(),
	});
	const state = createProductionAutonomousState(plan(), {
		runId: "run-479",
		maxConcurrency: 1,
		timeoutMs: 30_000,
	});
	return {
		root,
		pipeline,
		state,
		calls,
		session,
		effects,
		decisions,
		recoveryCalls,
		physicalPullRequests: () => physicalPullRequests,
		publicationAttempts: () => publicationAttempts,
		receipt: () => receipt,
		setParentHead(value: string) { currentParentHead = value; },
	};
}

async function persistAndAcknowledge(
	pipeline: ProductionChildPipeline,
	state: ProductionAutonomousState,
	checkpoint: ProductionStageCheckpoint,
): Promise<void> {
	persist(state, checkpoint);
	for (const key of new Set([
		...(checkpoint.effectKeys ?? []),
		...(checkpoint.effectKey === undefined ? [] : [checkpoint.effectKey]),
	])) await pipeline.acknowledge(key, structuredClone(state));
}

test("composes every child stage into exact durable checkpoints and never merges main", async (t) => {
	const h = await harness(t);
	const workspace = await h.pipeline.workspace(contextFor(h.state));
	assert.deepEqual(workspace.workspace, binding());
	await persistAndAcknowledge(h.pipeline, h.state, workspace);

	const implemented = await h.pipeline.implement(contextFor(h.state));
	await persistAndAcknowledge(h.pipeline, h.state, implemented);
	const verified = await h.pipeline.verify(contextFor(h.state));
	await persistAndAcknowledge(h.pipeline, h.state, verified);
	const published = await h.pipeline.publish(contextFor(h.state));
	assert.equal(published.pullRequest, 77);
	assert.equal(published.workspace?.head, SHA_C);
	assert.equal(published.effectKeys?.length, 3);
	await persistAndAcknowledge(h.pipeline, h.state, published);
	const reviewed = await h.pipeline.review(contextFor(h.state));
	assert.equal(reviewed.review?.status, "clean");
	assert.equal(reviewed.review?.head, SHA_C);
	await persistAndAcknowledge(h.pipeline, h.state, reviewed);
	const integrated = await h.pipeline.integrate(contextFor(h.state));
	assert.match(integrated.integrationReceiptDigest ?? "", /^[0-9a-f]{64}$/u);
	assert.equal(integrated.parentHead, SHA_B);
	await persistAndAcknowledge(h.pipeline, h.state, integrated);

	assert.deepEqual(await h.effects.listNonApplied(), []);
	assert.deepEqual(h.calls, [
		"parent-head", "workspace", "implement", "verify", "commit", "push", "handoff",
		"plan", "issue", "pull-request", "handoff", "review", "handoff", "parent-head",
		"integrate", "parent-head",
	]);
	assert.equal(h.calls.some((call) => /merge.*main|main.*merge/iu.test(call)), false);
	assert.equal(await h.pipeline.find({
		repository: "owner/repo",
		childId: "child",
		generation: 1,
		digest: integrated.integrationReceiptDigest!,
	}, new AbortController().signal).then((value) => productionChildIntegrationReceiptDigest(value!)), integrated.integrationReceiptDigest);
	await h.pipeline.join("run-479");
	await h.pipeline.close();
	assert.deepEqual(h.calls.slice(-3), ["join", "workspace-close", "github-stop"]);
});

test("corrects an exact failed verification before any pull request is published", async (t) => {
	const h = await harness(t, { verificationFailures: 1 });
	for (const stage of ["workspace", "implement"] as const) {
		const checkpoint = await h.pipeline[stage](contextFor(h.state));
		await persistAndAcknowledge(h.pipeline, h.state, checkpoint);
	}
	const failed = await h.pipeline.verify(contextFor(h.state));
	assert.equal(failed.verification?.status, "failed");
	assert.deepEqual(failed.verification?.commands, [{ id: "tests", status: "failed", failureKind: "exit" }]);
	await persistAndAcknowledge(h.pipeline, h.state, failed);
	h.state.children[0].corrections = 1;
	const corrected = await h.pipeline.correct(contextFor(h.state), ["Verification tests failed (exit)."]);
	await persistAndAcknowledge(h.pipeline, h.state, corrected);
	delete h.state.children[0].checkpoint?.verification;
	assert.equal(h.calls.includes("pull-request"), false);
	const passed = await h.pipeline.verify(contextFor(h.state));
	assert.equal(passed.verification?.status, "passed");
	await persistAndAcknowledge(h.pipeline, h.state, passed);
	await h.pipeline.close();
});

test("requests an issue-bound retry decision when a pre-PR correction budget is exhausted", async (t) => {
	const h = await harness(t, { verificationFailures: 1 });
	for (const stage of ["workspace", "implement", "verify"] as const) {
		const checkpoint = await h.pipeline[stage](contextFor(h.state));
		await persistAndAcknowledge(h.pipeline, h.state, checkpoint);
	}
	const request = await h.pipeline.requestIntervention(contextFor(h.state), "correction_budget_exhausted");
	assert.equal(request.pullRequest, undefined);
	assert.equal(request.head, undefined);
	h.state.childGate = {
		childId: "child", repository: "owner/repo", issue: 501, generation: 1,
		requestId: request.requestId, reason: "correction_budget_exhausted", status: "pending",
	};
	await h.pipeline.acknowledge(request.effectKey!, structuredClone(h.state));
	await h.pipeline.close();
});

test("records findings, correction dispositions, parent refresh, and exact child intervention consumption", async (t) => {
	const finding = { id: "BLOCK-1", severity: "blocking" as const, summary: "Fix the exact bug" };
	const h = await harness(t, { findings: [finding] });
	for (const stage of ["workspace", "implement", "verify", "publish"] as const) {
		const checkpoint = await h.pipeline[stage](contextFor(h.state));
		await persistAndAcknowledge(h.pipeline, h.state, checkpoint);
	}
	const reviewed = await h.pipeline.review(contextFor(h.state));
	assert.deepEqual(reviewed.review?.findings, [{ id: finding.id, summary: finding.summary }]);
	await persistAndAcknowledge(h.pipeline, h.state, reviewed);
	const corrected = await h.pipeline.correct(contextFor(h.state), [finding.summary]);
	await persistAndAcknowledge(h.pipeline, h.state, corrected);
	assert.equal(h.calls.includes("correct"), true);

	h.setParentHead(SHA_B);
	const refreshed = await h.pipeline.refresh(contextFor(h.state));
	assert.equal(refreshed.workspace?.baseHead, SHA_B);
	assert.equal(refreshed.workspace?.head, SHA_D);
	await persistAndAcknowledge(h.pipeline, h.state, refreshed);

	const request = await h.pipeline.requestIntervention(contextFor(h.state), "correction_budget_exhausted");
	assert.equal(request.pullRequest, undefined);
	assert.equal(request.head, undefined);
	h.state.childGate = {
		childId: "child",
		repository: "owner/repo",
		issue: 501,
		generation: 1,
		requestId: request.requestId,
		reason: "correction_budget_exhausted",
		status: "pending",
	};
	await h.pipeline.acknowledge(request.effectKey!, structuredClone(h.state));
	h.decisions.set(request.requestId, "authorized");
	const observed = await h.pipeline.observeIntervention(structuredClone(h.state), new AbortController().signal);
	assert.equal(observed.status, "authorized");
	h.state.childGate.status = "authorized";
	await h.pipeline.acknowledge(observed.effectKey!, structuredClone(h.state));
	assert.deepEqual(await h.effects.listNonApplied(), []);
});

test("reconciles a timed-out PR publication with one physical PR and stable effect identity", async (t) => {
	const h = await harness(t, { publicationTimeoutOnce: true });
	for (const stage of ["workspace", "implement", "verify"] as const) {
		const checkpoint = await h.pipeline[stage](contextFor(h.state));
		await persistAndAcknowledge(h.pipeline, h.state, checkpoint);
	}
	await assert.rejects(h.pipeline.publish(contextFor(h.state)), /timed out/u);
	const recovered = await h.pipeline.publish(contextFor(h.state));
	await persistAndAcknowledge(h.pipeline, h.state, recovered);
	assert.equal(h.physicalPullRequests(), 1);
	assert.equal(h.publicationAttempts(), 2);
	const journal = JSON.parse(await readFile(join(h.root, "production-effects.json"), "utf8")) as {
		records: Array<{ kind: string; key: string }>;
	};
	assert.equal(journal.records.filter((record) => record.kind === "child_pull_request").length, 1);
	assert.equal(journal.records.filter((record) => record.kind === "git_commit").length, 1);
	assert.equal(journal.records.filter((record) => record.kind === "git_push").length, 1);
	assert.equal(new Set(journal.records.map((record) => record.key)).size, journal.records.length);
});

test("fails closed on wrong or draft publication evidence", async (t) => {
	for (const [name, override] of [
		["wrong PR", { workItemId: "other-child" }],
		["draft PR", { draft: true }],
	] as const) {
		await t.test(name, async (childTest) => {
			const h = await harness(childTest, { pullRequest: override });
			for (const stage of ["workspace", "implement", "verify"] as const) {
				const checkpoint = await h.pipeline[stage](contextFor(h.state));
				await persistAndAcknowledge(h.pipeline, h.state, checkpoint);
			}
			await assert.rejects(h.pipeline.publish(contextFor(h.state)), ProductionLifecycleError);
		});
	}
});

test("classifies untrusted CI/review blockers and head movement without integrating", async (t) => {
	for (const [blocker, kind] of [
		["ci_not_green", "retryable"],
		["review_missing", "correction_required"],
		["undispositioned_finding", "correction_required"],
		["head_moved", "stale_parent"],
	] as const) {
		await t.test(blocker, async (childTest) => {
			const h = await harness(childTest, { integration: { kind: "blocked", blockers: [blocker] } });
			for (const stage of ["workspace", "implement", "verify", "publish", "review"] as const) {
				const checkpoint = await h.pipeline[stage](contextFor(h.state));
				await persistAndAcknowledge(h.pipeline, h.state, checkpoint);
			}
			await assert.rejects(
				h.pipeline.integrate(contextFor(h.state)),
				(error: unknown) => error instanceof ProductionLifecycleError && error.kind === kind,
			);
			const pending = (await h.effects.listNonApplied()).filter((record) => record.kind === "child_integration");
			assert.equal(pending.length, 1);
			assert.equal(pending[0].phase, "observed");
		});
	}
});

test("delegates prepared/observed crash recovery and makes close/abort idempotent", async (t) => {
	const h = await harness(t);
	const prepared = await h.effects.prepare({
		kind: "git_push",
		runId: "run-479",
		generation: 1,
		childId: "child",
		intentDigest: createHash("sha256").update("recover").digest("hex"),
		key: createHash("sha256").update(JSON.stringify({
			kind: "git_push",
			runId: "run-479",
			generation: 1,
			childId: "child",
			intentDigest: createHash("sha256").update("recover").digest("hex"),
		})).digest("hex"),
	});
	const observed = await h.pipeline.observe(prepared, new AbortController().signal);
	assert.match(observed.resultDigest, /^[0-9a-f]{64}$/u);
	await h.pipeline.apply({ ...prepared, phase: "observed", observedAt: new Date().toISOString(), resultDigest: observed.resultDigest }, new AbortController().signal);
	assert.deepEqual(h.recoveryCalls, ["observe:git_push", "apply:git_push"]);
	await Promise.all([h.pipeline.abort("run-479"), h.pipeline.abort("run-479")]);
	await Promise.all([h.pipeline.close(), h.pipeline.close()]);
	assert.equal(h.calls.filter((call) => call === "workspace-close").length, 1);
	assert.equal(h.calls.filter((call) => call === "github-stop").length, 1);
});
