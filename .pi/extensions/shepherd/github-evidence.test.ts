import assert from "node:assert/strict";
import { readFile } from "node:fs/promises";
import test from "node:test";

import {
	evaluateGitHubPullRequestEvidence,
	validateGitHubPullRequestEvidence,
	type GitHubEvidenceBlocker,
	type GitHubPullRequestEvidence,
} from "./github-evidence.ts";
import {
	createIndependentReviewWork,
	type IndependentReviewRecord,
} from "./review-router.ts";

const fixturePath = ".pi/extensions/shepherd/fixtures/issue-478/green-pull-request.json";
const baseSha = "a".repeat(40);
const headSha = "b".repeat(40);

async function fixture(): Promise<Record<string, unknown>> {
	return JSON.parse(await readFile(fixturePath, "utf8")) as Record<string, unknown>;
}

function cleanReview(overrides: Partial<IndependentReviewRecord> = {}): IndependentReviewRecord {
	return {
		...createIndependentReviewWork({
			repository: "github.com/polymetrics-ai/cli",
			workItemId: "evidence",
			pullRequest: 812,
			generation: 3,
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

async function evidence(overrides: Record<string, unknown> = {}): Promise<GitHubPullRequestEvidence> {
	return validateGitHubPullRequestEvidence({
		...await fixture(),
		reviews: [cleanReview()],
		...overrides,
	});
}

const expected = {
	number: 812,
	marker: "<!-- shepherd-child-pr:v1:471:evidence:0123456789abcdef01234567 -->",
	baseBranch: "feat/471-pi-agent-session-shepherd",
	headBranch: "feat/811-github-evidence",
	baseSha,
	headSha,
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
	for (const checks of [
		[],
		[{ name: "verify", status: "queued", conclusion: null, headSha }],
		[{ name: "verify", status: "in_progress", conclusion: null, headSha }],
		[{ name: "verify", status: "completed", conclusion: "failure", headSha }],
		[{ name: "verify", status: "completed", conclusion: "skipped", headSha }],
		[{ name: "verify", status: "completed", conclusion: "success", headSha: "c".repeat(40) }],
	]) {
		assert.ok(blockerCodes(evaluateGitHubPullRequestEvidence(await evidence({ checks }), expected)).includes("ci_not_green"));
	}
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

	const thread = { id: "RT-1", resolved: false, headSha };
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
	const duplicateCheck = { name: "verify", status: "completed", conclusion: "success", headSha };
	assert.throws(() => validateGitHubPullRequestEvidence({ ...raw, checks: [duplicateCheck, duplicateCheck] }), /duplicate|check/i);
	assert.throws(() => validateGitHubPullRequestEvidence({ ...raw, threads: Array.from({ length: 129 }, (_, index) => ({ id: `T-${index}`, resolved: true, headSha })) }), /bounded|threads|128/i);
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
