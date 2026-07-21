import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import {
	evaluateGitHubPullRequestEvidence,
	createRequiredGitHubCheckPolicy,
	validateGitHubPullRequestEvidence,
	type GitHubEvidenceBlocker,
	type GitHubPullRequestEvidence,
} from "./github-evidence.ts";
import {
	createIndependentReviewWork,
	createAgentSessionAttestation,
	independentReviewResultDigest,
	type AgentSessionAttestation,
	type IndependentReviewRecord,
} from "./review-router.ts";
import * as githubEvidenceApi from "./github-evidence.ts";
import * as reviewRouterApi from "./review-router.ts";

const fixturePath = ".pi/extensions/shepherd/fixtures/issue-478/green-pull-request.json";
const baseSha = "a".repeat(40);
const headSha = "b".repeat(40);

async function fixture(): Promise<Record<string, unknown>> {
	return {
		...JSON.parse(await readFile(fixturePath, "utf8")) as Record<string, unknown>,
		policyDigest: basePolicy.digest,
	};
}

function cleanReview(overrides: Partial<IndependentReviewRecord> = {}): IndependentReviewRecord {
	return {
		...createIndependentReviewWork({
			repository: "github.com/polymetrics-ai/cli",
			workItemId: "evidence",
		pullRequest: 812,
		generation: 3,
		baseBranch: "feat/471-pi-agent-session-shepherd",
		headBranch: "feat/811-github-evidence",
		baseSha,
			headSha,
			changedPaths: [".pi/extensions/shepherd/github-evidence.ts"],
			allowedScopes: [".pi/extensions/shepherd"],
		}),
		completedAt: "2026-07-21T12:00:00.000Z",
		verdict: "clean",
		findings: [],
		...overrides,
	};
}

function attestation(review: IndependentReviewRecord, overrides: Record<string, unknown> = {}): AgentSessionAttestation {
	return {
		...createAgentSessionAttestation({
			sessionId: "session-478-evidence",
			runId: "run-478-evidence-1",
			review,
		}),
		...overrides,
	} as AgentSessionAttestation;
}

async function evidence(overrides: Record<string, unknown> = {}): Promise<GitHubPullRequestEvidence> {
	return validateGitHubPullRequestEvidence({
		...await fixture(),
		reviews: [cleanReview()],
		...overrides,
	});
}

const canonicalReview = cleanReview();
const basePolicy = createRequiredGitHubCheckPolicy({
	schemaVersion: 1,
	repository: canonicalReview.repository,
	baseBranch: canonicalReview.baseBranch,
	revision: 7,
	requiredChecks: [{ name: "shepherd-tests", producerId: "github-actions:workflow-verify" }],
});
const expected = {
	repository: canonicalReview.repository,
	workItemId: canonicalReview.workItemId,
	generation: canonicalReview.generation,
	number: 812,
	marker: "<!-- shepherd-child-pr:v1:471:evidence:0123456789abcdef01234567 -->",
	baseBranch: "feat/471-pi-agent-session-shepherd",
	headBranch: "feat/811-github-evidence",
	baseSha,
	headSha,
	changedPathEvidence: {
		schemaVersion: 1 as const,
		authority: "controller" as const,
		repository: canonicalReview.repository,
		workItemId: canonicalReview.workItemId,
		pullRequest: canonicalReview.pullRequest,
		generation: canonicalReview.generation,
		baseSha,
		headSha,
		paths: [".pi/extensions/shepherd/github-evidence.ts"],
		complete: true as const,
		revision: 41,
		observedAt: "2026-07-21T11:58:00.000Z",
	},
	minimumObservation: { revision: 41, observedAt: "2026-07-21T11:58:00.000Z" },
	requiredCheckPolicy: basePolicy,
	reviewTarget: {
		repository: canonicalReview.repository,
		workItemId: canonicalReview.workItemId,
		pullRequest: canonicalReview.pullRequest,
		generation: canonicalReview.generation,
		baseBranch: canonicalReview.baseBranch,
		headBranch: canonicalReview.headBranch,
		baseSha: canonicalReview.baseSha,
		headSha: canonicalReview.headSha,
		changedPaths: canonicalReview.changedPaths,
		allowedScopes: canonicalReview.allowedScopes,
	},
	attestations: [attestation(canonicalReview)],
};

function blockerCodes(result: ReturnType<typeof evaluateGitHubPullRequestEvidence>): GitHubEvidenceBlocker[] {
	return result.kind === "blocked" ? result.blockers : [];
}

test("accepts only an open, green, resolved, exact-range independently reviewed PR", async () => {
	const candidate = await evidence();
	const decision = evaluateGitHubPullRequestEvidence(candidate, expected);
	assert.equal(decision.kind, "eligible");
	if (decision.kind === "eligible") assert.equal(decision.review.headSha, headSha);
});

test("failed, pending, skipped, missing, or stale-head checks are not green", async () => {
	const baseCheck = {
		id: "check-verify-case",
		name: "shepherd-tests",
		producerId: "github-actions:workflow-verify",
		sequence: 1,
		headSha,
		updatedAt: "2026-07-21T11:55:00.000Z",
		completedAt: "2026-07-21T11:55:00.000Z",
	};
	for (const checks of [
		[],
		[{ ...baseCheck, status: "queued", conclusion: null, completedAt: null }],
		[{ ...baseCheck, status: "in_progress", conclusion: null, completedAt: null }],
		[{ ...baseCheck, status: "completed", conclusion: "failure" }],
		[{ ...baseCheck, status: "completed", conclusion: "skipped" }],
		[{ ...baseCheck, status: "completed", conclusion: "success", headSha: "c".repeat(40) }],
	]) {
		assert.ok(blockerCodes(evaluateGitHubPullRequestEvidence(await evidence({ checks }), expected)).includes("ci_not_green"));
	}
});

test("required CI contexts use trusted producers and a complete deterministic latest rollup", async () => {
	const trusted = {
		id: "trusted-latest",
		name: "shepherd-tests",
		producerId: "github-actions:workflow-verify",
		sequence: 2,
		status: "completed",
		conclusion: "success",
		headSha,
		updatedAt: "2026-07-21T11:55:00.000Z",
		completedAt: "2026-07-21T11:55:00.000Z",
	};
	for (const changes of [
		{ checksComplete: false },
		{ checks: [{ ...trusted, producerId: "untrusted-app:42" }] },
		{ checks: [{ ...trusted, name: "unrelated-success" }] },
		{ checks: [{ ...trusted, status: "in_progress", conclusion: null, completedAt: null }] },
	]) {
		assert.ok(blockerCodes(evaluateGitHubPullRequestEvidence(await evidence(changes), expected as never)).includes("ci_not_green"));
	}
	const oldFailure = { ...trusted, id: "old-failure", sequence: 1, conclusion: "failure", updatedAt: "2026-07-21T11:00:00.000Z", completedAt: "2026-07-21T11:00:00.000Z" };
	assert.equal(evaluateGitHubPullRequestEvidence(await evidence({ checks: [trusted, oldFailure] }), expected as never).kind, "eligible");
	assert.equal(evaluateGitHubPullRequestEvidence(await evidence({ checks: [oldFailure, trusted] }), expected as never).kind, "eligible");
});

test("authoritative changed paths require exact set equality independent of ordering", async () => {
	const secondPath = ".pi/extensions/shepherd/review-router.ts";
	const expectedPaths = expected.changedPathEvidence.paths;
	const review = cleanReview({
		...createIndependentReviewWork({
			...expected.reviewTarget,
			changedPaths: [secondPath, ...expectedPaths],
		}),
	});
	const exactExpected = {
		...expected,
		changedPathEvidence: { ...expected.changedPathEvidence, paths: [...expectedPaths, secondPath] },
		reviewTarget: { ...expected.reviewTarget, changedPaths: [secondPath, ...expectedPaths] },
		attestations: [attestation(review)],
	};
	assert.equal(evaluateGitHubPullRequestEvidence(await evidence({
		changedPaths: [secondPath, ...expectedPaths],
		reviews: [review],
	}), exactExpected as never).kind, "eligible");
	for (const claimed of [[], expectedPaths, [...exactExpected.changedPathEvidence.paths, ".pi/extensions/shepherd/extra.ts"]]) {
		const mismatched = cleanReview({
			...createIndependentReviewWork({ ...exactExpected.reviewTarget, changedPaths: claimed }),
		});
		const result = evaluateGitHubPullRequestEvidence(await evidence({
			changedPaths: exactExpected.changedPathEvidence.paths,
			reviews: [mismatched],
		}), { ...exactExpected, attestations: [attestation(mismatched)] } as never);
		assert.ok(blockerCodes(result).includes("review_missing"));
	}
});

test("expected target makes equal-generation selection independent of review and page order", async () => {
	const correct = cleanReview();
	const unrelated = cleanReview({
		...createIndependentReviewWork({ ...expected.reviewTarget, workItemId: "unrelated" }),
		completedAt: "2026-07-21T12:01:00.000Z",
	});
	for (const reviews of [[correct, unrelated], [unrelated, correct]]) {
		const result = evaluateGitHubPullRequestEvidence(await evidence({ reviews }), {
			...expected,
			attestations: [attestation(correct), attestation(unrelated, { sessionId: "session-478-unrelated" })],
		} as never);
		assert.equal(result.kind, "eligible");
		if (result.kind === "eligible") assert.equal(result.review.workItemId, "evidence");
	}
	assert.ok(blockerCodes(evaluateGitHubPullRequestEvidence(await evidence({ reviewsComplete: false }), expected as never)).includes("review_missing"));
});

test("authoritative requested changes and unresolved review threads block integration", async () => {
	const change = {
		id: "R-1",
		actor: "reviewer",
		commitSha: headSha,
		dismissed: false,
		submittedAt: "2026-07-21T11:00:00.000Z",
	};
	assert.ok(blockerCodes(evaluateGitHubPullRequestEvidence(await evidence({ requestedChanges: [change] }), expected)).includes("changes_requested"));
	assert.equal(evaluateGitHubPullRequestEvidence(await evidence({ requestedChanges: [{ ...change, dismissed: true }] }), expected).kind, "eligible");

	const thread = { id: "RT-1", resolved: false, headSha, updatedAt: "2026-07-21T11:30:00.000Z" };
	assert.ok(blockerCodes(evaluateGitHubPullRequestEvidence(await evidence({ threads: [thread] }), expected)).includes("unresolved_thread"));
	assert.equal(evaluateGitHubPullRequestEvidence(await evidence({ threads: [{ ...thread, resolved: true }] }), expected).kind, "eligible");
});

test("every blocking finding needs an exact-current-head disposition plus a later clean review", async () => {
	const staleHead = "9".repeat(40);
	const findingsReview: IndependentReviewRecord = {
		...createIndependentReviewWork({
			repository: "github.com/polymetrics-ai/cli",
			workItemId: "evidence",
			pullRequest: 812,
			generation: 2,
			baseBranch: expected.baseBranch,
			headBranch: expected.headBranch,
			baseSha,
			headSha: staleHead,
			changedPaths: [".pi/extensions/shepherd/github-evidence.ts"],
			allowedScopes: [".pi/extensions/shepherd"],
		}),
		completedAt: "2026-07-21T10:00:00.000Z",
		verdict: "findings",
		findings: [{ id: "F-1", severity: "blocking", summary: "Reconcile before retry." }],
	};
	const reviews = [findingsReview, cleanReview()];
	assert.ok(blockerCodes(evaluateGitHubPullRequestEvidence(await evidence({ reviews }), expected)).includes("undispositioned_finding"));

	const dispositions = [{
		findingId: "F-1",
		kind: "fixed",
		rationale: "Re-read by marker before every retry.",
		actor: "maintainer",
		headSha,
		recordedAt: "2026-07-21T11:30:00.000Z",
	}];
	assert.equal(evaluateGitHubPullRequestEvidence(await evidence({ reviews, dispositions }), expected).kind, "eligible");
	assert.ok(blockerCodes(evaluateGitHubPullRequestEvidence(await evidence({ reviews, dispositions: [{ ...dispositions[0], headSha: staleHead }] }), expected)).includes("undispositioned_finding"));
	assert.ok(blockerCodes(evaluateGitHubPullRequestEvidence(await evidence({
		reviews,
		dispositions: [{ ...dispositions[0], kind: "not_actionable" }],
	}), expected)).includes("undispositioned_finding"));
});

test("head movement, stale reviewed ranges, topology drift, draft state, and merge conflicts fail closed", async () => {
	const staleReview: IndependentReviewRecord = {
		...createIndependentReviewWork({
			repository: "github.com/polymetrics-ai/cli",
			workItemId: "evidence",
			pullRequest: 812,
			generation: 3,
			baseBranch: expected.baseBranch,
			headBranch: expected.headBranch,
			baseSha,
			headSha: "c".repeat(40),
			changedPaths: [".pi/extensions/shepherd/github-evidence.ts"],
			allowedScopes: [".pi/extensions/shepherd"],
		}),
		completedAt: "2026-07-21T12:00:00.000Z",
		verdict: "clean",
		findings: [],
	};
	const cases: Array<[Record<string, unknown>, Partial<typeof expected>, GitHubEvidenceBlocker]> = [
		[{ headSha: "c".repeat(40) }, {}, "head_moved"],
		[{ reviews: [staleReview] }, {}, "review_missing"],
		[{ baseBranch: "main" }, {}, "topology_mismatch"],
		[{ draft: true }, {}, "draft"],
		[{ mergeState: "conflicting" }, {}, "merge_blocked"],
	];
	for (const [changes, expectedChanges, blocker] of cases) {
		const result = evaluateGitHubPullRequestEvidence(await evidence(changes), { ...expected, ...expectedChanges });
		assert.ok(blockerCodes(result).includes(blocker), blocker);
	}
	assert.equal(evaluateGitHubPullRequestEvidence(await evidence({ draft: true }), expected, { allowDraft: true }).kind, "eligible");
});

test("rejects aggregate review claims, unknown fields, duplicate IDs, and unbounded evidence", async () => {
	const raw = await fixture();
	assert.throws(() => validateGitHubPullRequestEvidence({ ...raw, reviewDecision: "APPROVED" }), /field|shape|evidence/i);
	const duplicateCheck = { id: "duplicate", name: "verify", producerId: "producer", sequence: 1, status: "completed", conclusion: "success", headSha, updatedAt: "2026-07-21T11:00:00.000Z", completedAt: "2026-07-21T11:00:00.000Z" };
	assert.throws(() => validateGitHubPullRequestEvidence({ ...raw, checks: [duplicateCheck, duplicateCheck] }), /duplicate|check/i);
	assert.throws(() => validateGitHubPullRequestEvidence({ ...raw, threads: Array.from({ length: 129 }, (_, index) => ({ id: `T-${index}`, resolved: true, headSha, updatedAt: "2026-07-21T11:00:00.000Z" })) }), /bounded|threads|128/i);
	assert.throws(() => validateGitHubPullRequestEvidence({ ...raw, body: "x".repeat(65_537) }), /body|bounded/i);
});

test("rejects ambiguous duplicate finding IDs across review generations", async () => {
	const raw = await fixture();
	const first = {
		...createIndependentReviewWork({
			repository: "github.com/polymetrics-ai/cli",
			workItemId: "evidence",
			pullRequest: 812,
			generation: 2,
			baseBranch: expected.baseBranch,
			headBranch: expected.headBranch,
			baseSha,
			headSha: "9".repeat(40),
			changedPaths: [".pi/extensions/shepherd/github-evidence.ts"],
			allowedScopes: [".pi/extensions/shepherd"],
		}),
		completedAt: "2026-07-21T10:00:00.000Z",
		verdict: "findings" as const,
		findings: [{ id: "F-duplicate", severity: "blocking" as const, summary: "First finding." }],
	};
	const second = {
		...createIndependentReviewWork({
			repository: "github.com/polymetrics-ai/cli",
			workItemId: "evidence",
			pullRequest: 812,
			generation: 3,
			baseBranch: expected.baseBranch,
			headBranch: expected.headBranch,
			baseSha,
			headSha,
			changedPaths: [".pi/extensions/shepherd/github-evidence.ts"],
			allowedScopes: [".pi/extensions/shepherd"],
		}),
		completedAt: "2026-07-21T12:00:00.000Z",
		verdict: "findings" as const,
		findings: [{ id: "F-duplicate", severity: "blocking" as const, summary: "Ambiguous second finding." }],
	};
	assert.throws(() => validateGitHubPullRequestEvidence({ ...raw, reviews: [first, second] }), /duplicate|finding/i);
});

test("rejects proxied evidence arrays without invoking their traps", async () => {
	const raw = await fixture();
	let trapInvoked = false;
	const checks = new Proxy([], {
		get() {
			trapInvoked = true;
			throw new Error("proxy trap must not execute");
		},
	});
	assert.throws(() => validateGitHubPullRequestEvidence({ ...raw, checks }), /array|shape|checks|proxy/i);
	assert.equal(trapInvoked, false);
});

const cycle3Repository = "github.com/polymetrics-ai/cli";
const cycle3BaseBranch = "feat/471-pi-agent-session-shepherd";
const cycle3HeadBranch = "feat/811-github-evidence";

function cycle3Policy(overrides: Record<string, unknown> = {}): Record<string, unknown> {
	const input = {
		schemaVersion: 1,
		repository: cycle3Repository,
		baseBranch: cycle3BaseBranch,
		revision: 7,
		requiredChecks: [
			{ name: "shepherd-tests", producerId: "github-actions:workflow-verify" },
			{ name: "strict-types", producerId: "github-actions:workflow-types" },
		],
		...overrides,
	};
	const create = (githubEvidenceApi as Record<string, unknown>).createRequiredGitHubCheckPolicy;
	return typeof create === "function"
		? (create as (value: unknown) => Record<string, unknown>)(input)
		: { ...input, digest: "0".repeat(64) };
}

function cycle3Review(overrides: Record<string, unknown> = {}): Record<string, unknown> {
	return {
		...cleanReview(),
		baseBranch: cycle3BaseBranch,
		headBranch: cycle3HeadBranch,
		completedAt: "2026-07-21T12:03:00.000Z",
		...overrides,
	};
}

function cycle3Attestation(review: Record<string, unknown>, overrides: Record<string, unknown> = {}): Record<string, unknown> {
	const create = (reviewRouterApi as Record<string, unknown>).createAgentSessionAttestation;
	if (typeof create === "function") {
		return {
			...(create as (value: unknown) => Record<string, unknown>)({
				sessionId: "session-478-cycle-3",
				runId: "run-478-cycle-3-1",
				review,
			}),
			...overrides,
		};
	}
	return { ...attestation(review as unknown as IndependentReviewRecord), ...overrides } as Record<string, unknown>;
}

function cycle3Evidence(overrides: Record<string, unknown> = {}): Record<string, unknown> {
	const policy = cycle3Policy();
	return {
		schemaVersion: 2,
		repository: cycle3Repository,
		workItemId: "evidence",
		generation: 3,
		number: 812,
		marker: expected.marker,
		title: "feat(shepherd): add GitHub evidence",
		body: `Refs #811\nRefs #471\n\n${expected.marker}`,
		state: "open",
		draft: false,
		baseBranch: cycle3BaseBranch,
		headBranch: cycle3HeadBranch,
		baseSha,
		headSha,
		changedPathsComplete: true,
		changedPaths: [...expected.changedPathEvidence.paths],
		allowedScopes: [".pi/extensions/shepherd"],
		mergeState: "clean",
		policyDigest: policy.digest,
		checksComplete: true,
		checks: [
			{
				id: "check-shepherd-tests",
				name: "shepherd-tests",
				producerId: "github-actions:workflow-verify",
				sequence: 1,
				status: "completed",
				conclusion: "success",
				headSha,
				updatedAt: "2026-07-21T11:55:00.000Z",
				completedAt: "2026-07-21T11:55:00.000Z",
			},
			{
				id: "check-strict-types",
				name: "strict-types",
				producerId: "github-actions:workflow-types",
				sequence: 1,
				status: "completed",
				conclusion: "success",
				headSha,
				updatedAt: "2026-07-21T11:56:00.000Z",
				completedAt: "2026-07-21T11:56:00.000Z",
			},
		],
		requestedChangesComplete: true,
		requestedChanges: [],
		threadsComplete: true,
		threads: [],
		reviewsComplete: true,
		reviews: [cycle3Review()],
		dispositionsComplete: true,
		dispositions: [],
		revision: 42,
		observedAt: "2026-07-21T12:05:00.000Z",
		...overrides,
	};
}

function cycle3Expected(overrides: Record<string, unknown> = {}): Record<string, unknown> {
	const review = cycle3Review();
	return {
		repository: cycle3Repository,
		workItemId: "evidence",
		generation: 3,
		number: 812,
		marker: expected.marker,
		baseBranch: cycle3BaseBranch,
		headBranch: cycle3HeadBranch,
		baseSha,
		headSha,
		changedPathEvidence: {
			schemaVersion: 1,
			authority: "controller",
			repository: cycle3Repository,
			workItemId: "evidence",
			pullRequest: 812,
			generation: 3,
			baseSha,
			headSha,
			paths: [...expected.changedPathEvidence.paths],
			complete: true,
			revision: 41,
			observedAt: "2026-07-21T11:58:00.000Z",
		},
		minimumObservation: {
			revision: 41,
			observedAt: "2026-07-21T11:58:00.000Z",
		},
		requiredCheckPolicy: cycle3Policy(),
		reviewTarget: {
			repository: cycle3Repository,
			workItemId: "evidence",
			pullRequest: 812,
			generation: 3,
			baseBranch: cycle3BaseBranch,
			headBranch: cycle3HeadBranch,
			baseSha,
			headSha,
			changedPaths: [...expected.changedPathEvidence.paths],
			allowedScopes: [".pi/extensions/shepherd"],
		},
		attestations: [cycle3Attestation(review)],
		...overrides,
	};
}

function evaluateCycle3(evidenceValue = cycle3Evidence(), expectedValue = cycle3Expected()) {
	return (evaluateGitHubPullRequestEvidence as unknown as (
		evidence: unknown,
		expectedValue: unknown,
	) => ReturnType<typeof evaluateGitHubPullRequestEvidence>)(evidenceValue, expectedValue);
}

test("cycle 3 binds outer PR evidence and independent review identity at every coordinate", () => {
	assert.equal(evaluateCycle3().kind, "eligible");
	const cases: Array<[string, unknown]> = [
		["repository", "github.com/other/cli"],
		["workItemId", "other-work"],
		["generation", 4],
		["number", 999],
		["baseBranch", "main"],
		["headBranch", "feat/999-other"],
		["baseSha", "c".repeat(40)],
		["headSha", "d".repeat(40)],
	];
	for (const [field, value] of cases) {
		const result = evaluateCycle3(cycle3Evidence({ [field]: value }));
		assert.equal(result.kind, "blocked", field);
		assert.ok(blockerCodes(result).some((code) => code === "resource_mismatch" || code === "topology_mismatch" || code === "head_moved"), field);
	}
});

test("cycle 3 requires independently complete nested evidence and a fresh controller observation", () => {
	for (const field of [
		"changedPathsComplete",
		"checksComplete",
		"requestedChangesComplete",
		"threadsComplete",
		"reviewsComplete",
		"dispositionsComplete",
	]) {
		const result = evaluateCycle3(cycle3Evidence({ [field]: false }));
		assert.equal(result.kind, "blocked", field);
		assert.ok(blockerCodes(result).includes("evidence_incomplete" as never), field);
	}
	const staleRevision = evaluateCycle3(cycle3Evidence({ revision: 40 }));
	assert.ok(blockerCodes(staleRevision).includes("stale_evidence" as never));
	const staleTimeExpected = cycle3Expected({
		changedPathEvidence: {
			...(cycle3Expected().changedPathEvidence as Record<string, unknown>),
			observedAt: "2026-07-21T12:05:00.000Z",
		},
		minimumObservation: { revision: 41, observedAt: "2026-07-21T12:05:00.000Z" },
	});
	const staleTime = evaluateCycle3(cycle3Evidence({ observedAt: "2026-07-21T12:04:00.000Z" }), staleTimeExpected);
	assert.ok(blockerCodes(staleTime).includes("stale_evidence" as never));
	const independentPaths = [".pi/extensions/shepherd/review-router.ts"];
	const independentReview = {
		...createIndependentReviewWork({
			...(cycle3Expected().reviewTarget as IndependentReviewRecord),
			changedPaths: independentPaths,
		}),
		completedAt: "2026-07-21T12:03:00.000Z",
		verdict: "clean" as const,
		findings: [],
	};
	const wrongIndependentPaths = cycle3Expected({
		changedPathEvidence: {
			...(cycle3Expected().changedPathEvidence as Record<string, unknown>),
			paths: independentPaths,
		},
		reviewTarget: { ...(cycle3Expected().reviewTarget as Record<string, unknown>), changedPaths: independentPaths },
		attestations: [cycle3Attestation(independentReview)],
	});
	assert.ok(blockerCodes(evaluateCycle3(cycle3Evidence(), wrongIndependentPaths)).includes("resource_mismatch"));
});

test("cycle 3 enforces event causality, sequence rollups, and a post-disposition clean review", () => {
	const baseCheck = (cycle3Evidence().checks as Array<Record<string, unknown>>)[0];
	assert.throws(() => evaluateCycle3(cycle3Evidence({
		checks: [{ ...baseCheck, status: "in_progress", conclusion: null, completedAt: baseCheck.completedAt }],
	})), /completed|pending|check|chronolog/i);
	assert.throws(() => evaluateCycle3(cycle3Evidence({
		checks: [{ ...baseCheck, updatedAt: "2026-07-21T12:06:00.000Z", completedAt: "2026-07-21T12:06:00.000Z" }],
	})), /future|observation|chronolog|timestamp/i);

	const latestFailure = {
		...baseCheck,
		id: "check-shepherd-tests-2",
		sequence: 2,
		conclusion: "failure",
		updatedAt: "2026-07-21T11:59:00.000Z",
		completedAt: "2026-07-21T11:59:00.000Z",
	};
	const otherCheck = (cycle3Evidence().checks as Array<Record<string, unknown>>)[1];
	assert.ok(blockerCodes(evaluateCycle3(cycle3Evidence({ checks: [latestFailure, otherCheck, baseCheck] }))).includes("ci_not_green"));
	assert.throws(() => evaluateCycle3(cycle3Evidence({
		checks: [baseCheck, { ...baseCheck, id: "duplicate-sequence" }, otherCheck],
	})), /sequence|ambiguous|duplicate/i);

	const findingReview = cycle3Review({
		completedAt: "2026-07-21T12:01:00.000Z",
		verdict: "findings",
		findings: [{ id: "C3-F-1", severity: "blocking", summary: "Durable retry is missing." }],
	});
	const disposition = {
		findingId: "C3-F-1",
		kind: "fixed",
		rationale: "Added durable conditional mutation reconciliation.",
		actor: "maintainer",
		headSha,
		recordedAt: "2026-07-21T12:02:00.000Z",
	};
	const oldClean = cycle3Review({ completedAt: "2026-07-21T12:00:00.000Z" });
	const expectedOld = cycle3Expected({ attestations: [cycle3Attestation(oldClean)] });
	assert.ok(blockerCodes(evaluateCycle3(cycle3Evidence({
		reviews: [oldClean, findingReview],
		dispositions: [disposition],
	}), expectedOld)).includes("review_missing"));
	assert.ok(blockerCodes(evaluateCycle3(cycle3Evidence({
		reviews: [oldClean, findingReview],
		dispositions: [{ ...disposition, recordedAt: "2026-07-21T12:00:30.000Z" }],
	}), expectedOld)).includes("undispositioned_finding"));
	const finalClean = cycle3Review({ completedAt: "2026-07-21T12:03:00.000Z" });
	assert.equal(evaluateCycle3(cycle3Evidence({
		reviews: [oldClean, findingReview, finalClean],
		dispositions: [disposition],
	}), cycle3Expected({ attestations: [cycle3Attestation(finalClean)] })).kind, "eligible");
});

test("cycle 3 uses only a versioned repository/base policy and blocks policy movement", () => {
	const api = githubEvidenceApi as Record<string, unknown>;
	assert.equal(typeof api.createRequiredGitHubCheckPolicy, "function");
	assert.equal(typeof api.validateRequiredGitHubCheckPolicy, "function");
	const policy = cycle3Policy();
	assert.match(String(policy.digest), /^[0-9a-f]{64}$/u);
	for (const policyChanges of [
		{ repository: "github.com/other/cli" },
		{ baseBranch: "main" },
		{ revision: 8 },
		{ digest: "f".repeat(64) },
	]) {
		assert.throws(
			() => evaluateCycle3(cycle3Evidence(), cycle3Expected({ requiredCheckPolicy: { ...policy, ...policyChanges } })),
			/policy|digest|repository|base|revision/i,
		);
	}
	const missingRequiredContext = cycle3Evidence({ checks: [(cycle3Evidence().checks as unknown[])[0]] });
	assert.ok(blockerCodes(evaluateCycle3(missingRequiredContext)).includes("ci_not_green"));
});

test("cycle 3 evidence rejects proxies, accessors, cycles, oversize, and secret-bearing malformed input safely", () => {
	let invoked = false;
	const accessor = { ...cycle3Evidence() };
	Object.defineProperty(accessor, "repository", {
		enumerable: true,
		get() {
			invoked = true;
			throw new Error("SECRET_TOKEN_SHOULD_NOT_ESCAPE");
		},
	});
	assert.throws(() => evaluateCycle3(accessor), /shape|accessor|evidence/i);
	assert.equal(invoked, false);
	const proxied = new Proxy(cycle3Evidence(), {
		get() {
			invoked = true;
			throw new Error("SECRET_TOKEN_SHOULD_NOT_ESCAPE");
		},
	});
	assert.throws(() => evaluateCycle3(proxied), /shape|proxy|evidence/i);
	assert.equal(invoked, false);
	const cyclic = cycle3Evidence();
	(cyclic.checks as unknown[]).push(cyclic);
	assert.throws(() => evaluateCycle3(cyclic), /shape|check|cycle|evidence/i);
	assert.throws(() => evaluateCycle3(cycle3Evidence({
		requestedChanges: Array.from({ length: 129 }, (_, index) => ({
			id: `R-${index}`,
			actor: "reviewer",
			commitSha: headSha,
			dismissed: true,
			submittedAt: "2026-07-21T11:00:00.000Z",
		})),
	})), /bounded|requested|128/i);
	try {
		evaluateCycle3({ ...cycle3Evidence(), unexpected: "SECRET_TOKEN_SHOULD_NOT_ESCAPE" });
		assert.fail("unknown secret-bearing field must fail");
	} catch (error) {
		assert.doesNotMatch(String(error), /SECRET_TOKEN_SHOULD_NOT_ESCAPE/u);
	}
});
