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

test("cycle 10 warning findings require later exact dispositions and a causally later clean review", async (t) => {
	const warningReview = cleanReview({
		completedAt: "2026-07-21T10:00:00.000Z",
		verdict: "findings",
		findings: [{ id: "W-10", severity: "warning", summary: "Disposition this warning before readiness." }],
	});
	const laterClean = cleanReview({ completedAt: "2026-07-21T12:00:00.000Z" });
	const reviews = [warningReview, laterClean];
	const expectedWithAttempts = {
		...expected,
		attestations: [
			attestation(warningReview, { sessionId: "session-478-warning", runId: "run-478-warning" }),
			attestation(laterClean, { sessionId: "session-478-warning-clean", runId: "run-478-warning-clean" }),
		],
	};
	const fixed = {
		findingId: "W-10",
		kind: "fixed" as const,
		rationale: "The warning was corrected and rereviewed.",
		actor: "maintainer",
		headSha,
		recordedAt: "2026-07-21T11:00:00.000Z",
	};
	const mustBlock: Array<[string, Record<string, unknown>]> = [
		["missing disposition", {}],
		["stale-head disposition", { dispositions: [{ ...fixed, headSha: "9".repeat(40) }] }],
		["pre-finding disposition", { dispositions: [{ ...fixed, recordedAt: "2026-07-21T09:59:59.000Z" }] }],
		["equal-time disposition", { dispositions: [{ ...fixed, recordedAt: warningReview.completedAt }] }],
		["unauthorized not-actionable disposition", { dispositions: [{ ...fixed, kind: "not_actionable" }] }],
		["clean review before disposition", { dispositions: [{ ...fixed, recordedAt: "2026-07-21T12:30:00.000Z" }] }],
	];
	for (const [name, overrides] of mustBlock) {
		await t.test(name, async () => {
			const result = evaluateGitHubPullRequestEvidence(
				await evidence({ reviews, ...overrides, observedAt: "2026-07-21T13:00:00.000Z" }),
				expectedWithAttempts,
			);
			assert.ok(
				blockerCodes(result).some((blocker) => blocker === "undispositioned_finding" || blocker === "review_missing"),
				name,
			);
		});
	}
	await t.test("later fixed disposition followed by later clean review", async () => {
		assert.equal(evaluateGitHubPullRequestEvidence(
			await evidence({ reviews, dispositions: [fixed] }),
			expectedWithAttempts,
		).kind, "eligible");
	});
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

test("cycle 4 exposes a complete controller-owned current check-policy observation", () => {
	const api = githubEvidenceApi as Record<string, unknown>;
	assert.equal(typeof api.createRequiredGitHubCheckPolicyObservation, "function");
	assert.equal(typeof api.validateRequiredGitHubCheckPolicyObservation, "function");
	const create = api.createRequiredGitHubCheckPolicyObservation as (value: unknown) => Record<string, unknown>;
	const validate = api.validateRequiredGitHubCheckPolicyObservation as (value: unknown) => Record<string, unknown>;
	const observation = create({
		schemaVersion: 1,
		authority: "controller",
		repository: basePolicy.repository,
		baseBranch: basePolicy.baseBranch,
		revision: basePolicy.revision,
		digest: basePolicy.digest,
		observedAt: "2026-07-21T12:06:00.000Z",
	});
	assert.deepEqual(validate(JSON.parse(JSON.stringify(observation))), observation);
	for (const moved of [
		{ ...observation, authority: "transport" },
		{ ...observation, repository: "github.com/other/cli" },
		{ ...observation, baseBranch: "main" },
		{ ...observation, revision: Number(observation.revision) + 1 },
		{ ...observation, digest: "f".repeat(64) },
	]) {
		if ("authority" in moved && moved.authority === "transport") assert.throws(() => validate(moved), /controller|authority|policy/i);
		else assert.notDeepEqual(validate(moved), observation);
	}
});

test("cycle 4 rejects every pseudo or symbolic ref in policies and PR evidence", async () => {
	const pullRequest = await fixture();
	for (const invalid of [
		"HEAD", "FETCH_HEAD", "ORIG_HEAD", "MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD",
		"REBASE_HEAD", "BISECT_HEAD", "AUTO_MERGE", "topic/ORIG_HEAD", "refs/heads/topic",
	]) {
		assert.throws(() => createRequiredGitHubCheckPolicy({
			schemaVersion: 1,
			repository: basePolicy.repository,
			baseBranch: invalid,
			revision: 7,
			requiredChecks: [{ name: "verify", producerId: "github-actions:workflow-verify" }],
		}), /branch|ref|pseudo|symbolic/i, invalid);
		assert.throws(() => validateGitHubPullRequestEvidence({ ...pullRequest, baseBranch: invalid }), /branch|ref|evidence/i, invalid);
		assert.throws(() => validateGitHubPullRequestEvidence({ ...pullRequest, headBranch: invalid }), /branch|ref|evidence/i, invalid);
	}
});

test("cycle 4 pre-bounds dense and sparse evidence arrays before descriptor materialization", () => {
	const dense = Array.from({ length: 129 }, (_, index) => ({ id: `check-${index}` }));
	const original = Object.getOwnPropertyDescriptors;
	let traversed = false;
	let rejection: unknown;
	Object.getOwnPropertyDescriptors = ((value: object) => {
		if (value === dense) {
			traversed = true;
			throw new Error("descriptor traversal must not occur");
		}
		return original(value);
	}) as typeof Object.getOwnPropertyDescriptors;
	try {
		validateGitHubPullRequestEvidence({ ...cycle3Evidence(), checks: dense });
	} catch (error) {
		rejection = error;
	} finally {
		Object.getOwnPropertyDescriptors = original;
	}
	assert.equal(traversed, false);
	assert.match(String(rejection), /bounded|checks|128/i);

	const sparse: unknown[] = [];
	sparse.length = 1_000_000;
	assert.throws(() => validateGitHubPullRequestEvidence({ ...cycle3Evidence(), reviews: sparse }), /bounded|dense|reviews|array/i);
});

test("cycle 5 allows only documented non-draft integrated PR states during receipt reauthorization", async () => {
	const merged = await evidence({ state: "merged", draft: false });
	const accepted = evaluateGitHubPullRequestEvidence(merged, expected, { allowIntegrated: true } as never);
	assert.equal(accepted.kind, "eligible");

	for (const invalid of [
		{ state: "merged" as const, draft: true },
		{ state: "closed" as const, draft: false },
	]) {
		const decision = evaluateGitHubPullRequestEvidence(await evidence(invalid), expected, { allowIntegrated: true } as never);
		assert.equal(decision.kind, "blocked");
	}
});

test("cycle 5 rejects cookie/session values in every PR review and disposition text boundary", async (t) => {
	const candidate = await evidence();
	const marker = "SYNTHETIC_CYCLE5_VALUE";
	const findingReview: IndependentReviewRecord = {
		...canonicalReview,
		verdict: "findings",
		findings: [{ id: "cycle-5-finding", severity: "warning", summary: "Synthetic review finding" }],
	};
	const cases: Array<[string, Record<string, unknown>]> = [
		["PR title", { ...candidate, title: `Set-Cookie: session=${marker}; HttpOnly` }],
		["PR body", { ...candidate, body: `${candidate.body}\nCookie: sid=${marker}` }],
		["review finding", {
			...candidate,
			reviews: [{
				...findingReview,
				findings: [{ id: "cycle-5-finding", severity: "warning", summary: `X-Session-Token: ${marker}` }],
			}],
		}],
		["finding disposition", {
			...candidate,
			reviews: [findingReview],
			dispositions: [{
				findingId: "cycle-5-finding",
				kind: "fixed",
				rationale: `session cookie=${marker}`,
				actor: "maintainer",
				headSha,
				recordedAt: "2026-07-21T12:01:00.000Z",
			}],
		}],
	];
	for (const [name, value] of cases) {
		await t.test(name, () => {
			assert.throws(() => validateGitHubPullRequestEvidence(value), /credential|secret|sensitive/i);
		});
	}
});

test("cycle 6 normalizes revoked evidence-array proxies before host array operations", () => {
	const revoked = Proxy.revocable([], {});
	revoked.revoke();
	let rejection: unknown;
	try {
		validateGitHubPullRequestEvidence({ ...cycle3Evidence(), reviews: revoked.proxy });
	} catch (error) {
		rejection = error;
	}
	assert.ok(rejection instanceof Error);
	assert.match(String(rejection), /invalid|bounded|shape|proxy|array/i);
	assert.doesNotMatch(String(rejection), /Cannot perform ['"]?IsArray|revoked/i);
});

test("cycle 6 applies the complete shared credential grammar to PR review boundaries", async (t) => {
	const candidate = await evidence();
	const findingReview: IndependentReviewRecord = {
		...canonicalReview,
		verdict: "findings",
		findings: [{ id: "cycle-6-finding", severity: "warning", summary: "Synthetic review finding" }],
	};
	const cases: Array<[string, Record<string, unknown>]> = [
		["PR title", { ...candidate, title: "//registry.invalid/:_authToken=SYNTHETIC_NPM_MARKER" }],
		["PR body", { ...candidate, body: `${candidate.body}\npassword SYNTHETIC_NETRC_MARKER` }],
		["review finding", {
			...candidate,
			reviews: [{
				...findingReview,
				findings: [{ id: "cycle-6-finding", severity: "warning", summary: "aws_secret_access_key = SYNTHETIC_AWS_MARKER" }],
			}],
		}],
		["finding disposition", {
			...candidate,
			reviews: [findingReview],
			dispositions: [{
				findingId: "cycle-6-finding",
				kind: "fixed",
				rationale: "credentials_file = /tmp/SYNTHETIC_CREDENTIAL_FILE",
				actor: "maintainer",
				headSha,
				recordedAt: "2026-07-21T12:01:00.000Z",
			}],
		}],
	];
	for (const [name, value] of cases) {
		await t.test(name, () => assert.throws(
			() => validateGitHubPullRequestEvidence(value),
			/credential|secret|sensitive/i,
		));
	}
});

test("cycle 7 applies every finite Kubernetes Docker and AWS form to all PR text boundaries", async (t) => {
	const candidate = await evidence();
	const findingReview: IndependentReviewRecord = {
		...canonicalReview,
		verdict: "findings",
		findings: [{ id: "cycle-7-finding", severity: "warning", summary: "Synthetic review finding" }],
	};
	const samples = [
		"client-key-data: SYNTHETIC_KUBERNETES_KEY_DATA",
		"token: SYNTHETIC_KUBERNETES_TOKEN",
		'{"auth":"SYNTHETIC_DOCKER_AUTH"}',
		'{"identitytoken":"SYNTHETIC_DOCKER_IDENTITY_TOKEN"}',
		"aws_access_key_id = SYNTHETIC_AWS_ACCESS_KEY_ID",
		"aws_secret_access_key = SYNTHETIC_AWS_SECRET_ACCESS_KEY",
		"aws_session_token = SYNTHETIC_AWS_SESSION_TOKEN",
		"ASIAABCDEFGHIJKLMNOP",
	];
	for (const [index, sample] of samples.entries()) {
		await t.test(`schema form ${index + 1}`, () => {
			const values: Record<string, unknown>[] = [
				{ ...candidate, title: sample },
				{ ...candidate, body: `${candidate.body}\n${sample}` },
				{ ...candidate, reviews: [{ ...findingReview, findings: [{ id: "cycle-7-finding", severity: "warning", summary: sample }] }] },
				{
					...candidate,
					reviews: [findingReview],
					dispositions: [{
						findingId: "cycle-7-finding",
						kind: "fixed",
						rationale: sample,
						actor: "maintainer",
						headSha,
						recordedAt: "2026-07-21T12:01:00.000Z",
					}],
				},
			];
			for (const value of values) {
				let rejection: unknown;
				try { validateGitHubPullRequestEvidence(value); } catch (error) { rejection = error; }
				assert.ok(rejection instanceof Error);
				assert.match(rejection.message, /credential|secret|sensitive/i);
				assert.doesNotMatch(rejection.message, /SYNTHETIC_/u);
			}
		});
	}
});

test("cycle 8 provider-neutral credential suffixes close every PR evidence text boundary", async (t) => {
	const candidate = await evidence();
	const findingReview: IndependentReviewRecord = {
		...canonicalReview,
		verdict: "findings",
		findings: [{ id: "cycle-8-finding", severity: "warning", summary: "Synthetic review finding" }],
	};
	const suffixAssignments = [
		"UNLISTED_ALPHA_AUTHORIZATION=SYNTHETIC_CYCLE8_AUTHORIZATION_MARKER",
		"UNLISTED_BRAVO_TOKEN=SYNTHETIC_CYCLE8_TOKEN_MARKER",
		"UNLISTED_CHARLIE_ACCESS_TOKEN=SYNTHETIC_CYCLE8_ACCESS_TOKEN_MARKER",
		"UNLISTED_DELTA_REFRESH_TOKEN=SYNTHETIC_CYCLE8_REFRESH_TOKEN_MARKER",
		"UNLISTED_ECHO_API_KEY=SYNTHETIC_CYCLE8_API_KEY_MARKER",
		"UNLISTED_FOXTROT_PASSWORD=SYNTHETIC_CYCLE8_PASSWORD_MARKER",
		"UNLISTED_GOLF_SECRET=SYNTHETIC_CYCLE8_SECRET_MARKER",
		"UNLISTED_HOTEL_CLIENT_SECRET=SYNTHETIC_CYCLE8_CLIENT_SECRET_MARKER",
		"UNLISTED_INDIA_PRIVATE_KEY=SYNTHETIC_CYCLE8_PRIVATE_KEY_MARKER",
		"UNLISTED_JULIET_DATABASE_URL=SYNTHETIC_CYCLE8_DATABASE_URL_MARKER",
		"UNLISTED_KILO_CREDENTIAL=SYNTHETIC_CYCLE8_CREDENTIAL_MARKER",
		"UNLISTED_LIMA_CREDENTIALS=SYNTHETIC_CYCLE8_CREDENTIALS_MARKER",
		"UNLISTED_MIKE_COOKIE=SYNTHETIC_CYCLE8_COOKIE_MARKER",
		"UNLISTED_NOVEMBER_COOKIES=SYNTHETIC_CYCLE8_COOKIES_MARKER",
		"UNLISTED_OSCAR_SET_COOKIE=SYNTHETIC_CYCLE8_SET_COOKIE_MARKER",
		"UNLISTED_PAPA_SESSION=SYNTHETIC_CYCLE8_SESSION_MARKER",
		"UNLISTED_QUEBEC_SESSION_ID=SYNTHETIC_CYCLE8_SESSION_ID_MARKER",
		"UNLISTED_ROMEO_SESSION_TOKEN=SYNTHETIC_CYCLE8_SESSION_TOKEN_MARKER",
		"UNLISTED_SIERRA_SESSION_COOKIE=SYNTHETIC_CYCLE8_SESSION_COOKIE_MARKER",
		"UNLISTED_TANGO_CSRF_TOKEN=SYNTHETIC_CYCLE8_CSRF_TOKEN_MARKER",
	];
	const finiteSchemaAssignments = [
		"client-key-data: SYNTHETIC_CYCLE8_KUBERNETES_KEY_DATA",
		"token: SYNTHETIC_CYCLE8_KUBERNETES_TOKEN",
		'{"auth":"SYNTHETIC_CYCLE8_DOCKER_AUTH"}',
		'{"identitytoken":"SYNTHETIC_CYCLE8_DOCKER_IDENTITY_TOKEN"}',
		"aws_access_key_id = SYNTHETIC_CYCLE8_AWS_ACCESS_KEY_ID",
		"aws_secret_access_key = SYNTHETIC_CYCLE8_AWS_SECRET_ACCESS_KEY",
		"aws_session_token = SYNTHETIC_CYCLE8_AWS_SESSION_TOKEN",
		"ASIAABCDEFGHIJKLMNOP",
	];
	const boundaryValues = (text: string): Record<string, unknown>[] => [
		{ ...candidate, title: text },
		{ ...candidate, body: `${candidate.body}\n${text}` },
		{
			...candidate,
			reviews: [{
				...findingReview,
				findings: [{ id: "cycle-8-finding", severity: "warning", summary: text }],
			}],
		},
		{
			...candidate,
			reviews: [findingReview],
			dispositions: [{
				findingId: "cycle-8-finding",
				kind: "fixed",
				rationale: text,
				actor: "maintainer",
				headSha,
				recordedAt: "2026-07-21T12:01:00.000Z",
			}],
		},
	];

	await t.test("classifies every recognized suffix with an unknown provider prefix", () => {
		for (const assignment of suffixAssignments) {
			for (const value of boundaryValues(assignment)) {
				assert.throws(
					() => validateGitHubPullRequestEvidence(value),
					/credential|secret|sensitive/i,
					assignment,
				);
			}
		}
	});

	await t.test("rejects without reflecting the classified synthetic value", () => {
		for (const assignment of suffixAssignments) {
			const marker = assignment.slice(assignment.indexOf("=") + 1);
			for (const value of boundaryValues(assignment)) {
				let rejection: unknown;
				try {
					validateGitHubPullRequestEvidence(value);
				} catch (error) {
					rejection = error;
				}
				assert.ok(rejection instanceof Error, assignment);
				assert.match(rejection.message, /credential|secret|sensitive/i, assignment);
				assert.doesNotMatch(rejection.message, new RegExp(marker), assignment);
			}
		}
	});

	await t.test("allows only the exact documented public FEATURE_TOKEN field", () => {
		for (const value of boundaryValues("FEATURE_TOKEN=non-sensitive-build-label")) {
			assert.doesNotThrow(() => validateGitHubPullRequestEvidence(value));
		}
		for (const value of boundaryValues("UNLISTED_FEATURE_TOKEN=SYNTHETIC_CYCLE8_NEARBY_MARKER")) {
			assert.throws(
				() => validateGitHubPullRequestEvidence(value),
				/credential|secret|sensitive/i,
			);
		}
	});

	await t.test("retains finite Kubernetes Docker and AWS forms at every text field", () => {
		for (const assignment of finiteSchemaAssignments) {
			for (const value of boundaryValues(assignment)) {
				assert.throws(
					() => validateGitHubPullRequestEvidence(value),
					/credential|secret|sensitive/i,
					assignment,
				);
			}
		}
	});
});

function cycle9EvidenceAssignment(nameLength: number, marker = "CYCLE9_GITHUB_EVIDENCE_MARKER"): string {
	const suffix = "_TOKEN";
	if (nameLength <= suffix.length) throw new Error("cycle 9 assignment length is too short");
	return `V${"A".repeat(nameLength - suffix.length - 1)}${suffix}=${marker}`;
}

test("cycle 9 parses the complete bounded assignment name before GitHub evidence acceptance", async (t) => {
	const candidate = await evidence();
	const marker = "CYCLE9_GITHUB_EVIDENCE_MARKER";
	const bodyPrefix = `${candidate.body}\n`;
	const largestName = 65_536 - Buffer.byteLength(bodyPrefix) - marker.length - 1;
	const rows = [
		["leading underscore", `_UNLISTED_TOKEN=${marker}`, true],
		["127 characters", cycle9EvidenceAssignment(127, marker), true],
		["128 characters", cycle9EvidenceAssignment(128, marker), true],
		["129 characters", cycle9EvidenceAssignment(129, marker), true],
		["256 characters", cycle9EvidenceAssignment(256, marker), true],
		["largest in-field name", cycle9EvidenceAssignment(largestName, marker), true],
		["over-field name", cycle9EvidenceAssignment(largestName + 1, marker), true],
		["exact public control", "FEATURE_TOKEN=non-sensitive-build-label", false],
	] as const;
	for (const [name, assignment, rejects] of rows) {
		await t.test(name, () => {
			if (!rejects) {
				assert.equal(validateGitHubPullRequestEvidence({ ...candidate, body: `${bodyPrefix}${assignment}` }).body,
					`${bodyPrefix}${assignment}`);
				return;
			}
			let rejection: unknown;
			try {
				validateGitHubPullRequestEvidence({ ...candidate, body: `${bodyPrefix}${assignment}` });
			} catch (error) {
				rejection = error;
			}
			assert.ok(rejection instanceof Error);
			assert.match(rejection.message, /credential|secret|sensitive|invalid|bounded/i);
			assert.doesNotMatch(rejection.message, new RegExp(marker, "u"));
		});
	}
});

const cycle10AssignmentSuffixes = [
	"AUTHORIZATION", "TOKEN", "ACCESS_TOKEN", "REFRESH_TOKEN", "API_KEY", "PASSWORD", "SECRET",
	"CLIENT_SECRET", "PRIVATE_KEY", "DATABASE_URL", "CREDENTIAL", "CREDENTIALS", "COOKIE", "COOKIES",
	"SET_COOKIE", "SESSION", "SESSION_ID", "SESSION_TOKEN", "SESSION_COOKIE", "CSRF_TOKEN",
] as const;

test("cycle 10 closes assignment operator case and index policy before GitHub evidence acceptance", async (t) => {
	const candidate = await evidence();
	const marker = "CYCLE10_GITHUB_EVIDENCE_MARKER";
	const prefix = `${candidate.body}\n`;
	const rows: Array<[string, string, boolean]> = [
		...cycle10AssignmentSuffixes.map((suffix): [string, string, boolean] =>
			[`append ${suffix}`, `ACME_${suffix}+=${marker}`, true]),
		["lowercase base", `acme_api_key=${marker}`, true],
		["mixed-case base append", `AcMe_ApI_KeY+=${marker}`, true],
		["numeric index", `ACME_API_KEY[0]=${marker}`, true],
		["associative index append", `ACME_API_KEY[slot]+=${marker}`, true],
		["exact public ordinary control", "FEATURE_TOKEN=enabled", false],
		["exact public append control", "FEATURE_TOKEN+=enabled", false],
		["indexed public-lookalike", `FEATURE_TOKEN[0]=${marker}`, true],
	];
	for (const [name, assignment, rejects] of rows) {
		await t.test(name, () => {
			const body = `${prefix}${assignment}`;
			if (!rejects) {
				assert.equal(validateGitHubPullRequestEvidence({ ...candidate, body }).body, body);
				return;
			}
			let rejection: unknown;
			try {
				validateGitHubPullRequestEvidence({ ...candidate, body });
			} catch (error) {
				rejection = error;
			}
			assert.ok(rejection instanceof Error, name);
			assert.match(rejection.message, /credential|secret|sensitive|invalid|bounded/i);
			assert.doesNotMatch(rejection.message, new RegExp(marker, "u"));
		});
	}
});

function cycle11SensitiveAssignmentTails(marker: string): ReadonlyArray<readonly [string, string]> {
	const values: ReadonlyArray<readonly [string, string]> = [
		["escaped double quote", `"alpha\\"${marker}"`],
		["escaped whitespace", `alpha\\ ${marker}`],
		["line continuation", `alpha\\\n${marker}`],
		["command substitution", `$(printf ${marker})`],
		["parameter expansion", `\${UNSAFE:-${marker}}`],
	];
	return ["=", "+="].flatMap((operator) => values.map(([name, value]) =>
		[`${operator} ${name}`, `ACME_API_KEY${operator}${value}`] as const));
}

test("cycle 11 keeps assignment-tail rejection generic before GitHub evidence acceptance", async (t) => {
	const candidate = await evidence();
	const marker = "CYCLE11_GITHUB_EVIDENCE_MARKER";
	for (const [name, assignment] of cycle11SensitiveAssignmentTails(marker)) {
		await t.test(name, () => {
			let rejection: unknown;
			try {
				validateGitHubPullRequestEvidence({ ...candidate, body: `${candidate.body}\n${assignment}` });
			} catch (error) {
				rejection = error;
			}
			assert.ok(rejection instanceof Error, name);
			assert.match(rejection.message, /credential|secret|sensitive|invalid|bounded/i);
			assert.doesNotMatch(rejection.message, new RegExp(marker, "u"));
			assert.doesNotMatch(rejection.message, /API_KEY/iu);
		});
	}
});
