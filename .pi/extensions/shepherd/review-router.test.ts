import assert from "node:assert/strict";
import test from "node:test";

import {
	createIndependentReviewWork,
	reconcileIndependentReview,
	reviewCoversExactRange,
	validateIndependentReviewRecord,
	type IndependentReviewRecord,
	type IndependentReviewTarget,
} from "./review-router.ts";

const baseSha = "a".repeat(40);
const headSha = "b".repeat(40);

function target(overrides: Partial<IndependentReviewTarget> = {}): IndependentReviewTarget {
	return {
		repository: "github.com/polymetrics-ai/cli",
		workItemId: "issue-478",
		pullRequest: 812,
		generation: 3,
		baseSha,
		headSha,
		changedPaths: [".pi/extensions/shepherd/github-evidence.ts"],
		allowedScopes: [".pi/extensions/shepherd"],
		...overrides,
	};
}

function cleanReview(overrides: Partial<IndependentReviewRecord> = {}): IndependentReviewRecord {
	const work = createIndependentReviewWork(target());
	return {
		...work,
		completedAt: "2026-07-21T12:00:00.000Z",
		verdict: "clean",
		findings: [],
		...overrides,
	};
}

test("creates a deterministic declarative xhigh independent Codex review record", () => {
	const first = createIndependentReviewWork(target());
	const second = createIndependentReviewWork(target());
	assert.deepEqual(first, second);
	assert.deepEqual(
		{
			kind: first.kind,
			provider: first.provider,
			model: first.model,
			reasoningEffort: first.reasoningEffort,
			readOnly: first.readOnly,
		},
		{
			kind: "codex_independent",
			provider: "openai-codex",
			model: "gpt-5.6-sol",
			reasoningEffort: "xhigh",
			readOnly: true,
		},
	);
	assert.match(first.idempotencyMarker, /^<!-- shepherd-review:v1:/);
	assert.equal("run" in first, false);
	assert.equal("session" in first, false);
});

test("an exact clean review satisfies the route while head or base movement invalidates it", () => {
	const review = cleanReview();
	assert.equal(reviewCoversExactRange(review, baseSha, headSha), true);
	assert.deepEqual(reconcileIndependentReview({ target: target(), reviews: [review] }), {
		kind: "satisfied",
		review,
	});

	for (const moved of [
		target({ headSha: "c".repeat(40) }),
		target({ baseSha: "d".repeat(40) }),
	]) {
		const decision = reconcileIndependentReview({ target: moved, reviews: [review] });
		assert.equal(decision.kind, "dispatch");
		if (decision.kind === "dispatch") {
			assert.notEqual(decision.work.idempotencyMarker, review.idempotencyMarker);
		}
	}
});

test("a findings verdict never claims clean coverage", () => {
	const work = createIndependentReviewWork(target());
	const review = validateIndependentReviewRecord({
		...work,
		completedAt: "2026-07-21T12:00:00.000Z",
		verdict: "findings",
		findings: [{ id: "F-1", severity: "blocking", summary: "Head can move before integration." }],
	});
	assert.equal(reviewCoversExactRange(review, baseSha, headSha), false);
	assert.equal(reconcileIndependentReview({ target: target(), reviews: [review] }).kind, "dispatch");
});

test("rejects Claude, Copilot, generic Codex, human, wrong-model, and non-xhigh review records", () => {
	const canonical = cleanReview();
	const variants: Array<[string, unknown]> = [
		["Claude", { ...canonical, kind: "claude_primary", provider: "anthropic" }],
		["Copilot", { ...canonical, kind: "copilot", provider: "github" }],
		["generic Codex", { ...canonical, kind: "codex" }],
		["human", { ...canonical, kind: "human", provider: "github" }],
		["wrong model", { ...canonical, model: "gpt-5.5" }],
		["wrong effort", { ...canonical, reasoningEffort: "high" }],
		["writable", { ...canonical, readOnly: false }],
	];
	for (const [name, candidate] of variants) {
		assert.throws(() => validateIndependentReviewRecord(candidate), /independent|review|route|model|xhigh|read.only/i, name);
	}
});

test("fails closed on unknown fields, unsafe paths, oversized arrays, and marker tampering", () => {
	assert.throws(() => createIndependentReviewWork({ ...target(), changedPaths: Array.from({ length: 65 }, (_, index) => `src/${index}.ts`) }), /bounded|paths|64/i);
	assert.throws(() => createIndependentReviewWork({ ...target(), changedPaths: ["../outside.ts"] }), /path|scope/i);
	assert.throws(() => validateIndependentReviewRecord({ ...cleanReview(), unexpected: true }), /field|shape|review/i);
	assert.throws(() => validateIndependentReviewRecord({ ...cleanReview(), idempotencyMarker: "<!-- forged -->" }), /marker/i);
});

test("rejects proxied arrays without invoking their traps", () => {
	let trapInvoked = false;
	const paths = new Proxy([".pi/extensions/shepherd/review-router.ts"], {
		get() {
			trapInvoked = true;
			throw new Error("proxy trap must not execute");
		},
	});
	assert.throws(() => createIndependentReviewWork(target({ changedPaths: paths })), /array|shape|paths|proxy/i);
	assert.equal(trapInvoked, false);
});
