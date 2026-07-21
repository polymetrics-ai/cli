import assert from "node:assert/strict";
import { createHash } from "node:crypto";
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

function reviewDigest(review: IndependentReviewRecord): string {
	return createHash("sha256").update(JSON.stringify({
		idempotencyMarker: review.idempotencyMarker,
		repository: review.repository,
		workItemId: review.workItemId,
		pullRequest: review.pullRequest,
		generation: review.generation,
		baseSha: review.baseSha,
		headSha: review.headSha,
		changedPaths: review.changedPaths,
		allowedScopes: review.allowedScopes,
		completedAt: review.completedAt,
		verdict: review.verdict,
		findings: review.findings,
	})).digest("hex");
}

function attestation(review: IndependentReviewRecord, overrides: Record<string, unknown> = {}) {
	return {
		schemaVersion: 1,
		authority: "controller",
		sessionId: "session-478-review",
		runId: "run-478-review-1",
		provider: "openai-codex",
		model: "gpt-5.6-sol",
		reasoningEffort: "xhigh",
		readOnly: true,
		repository: review.repository,
		workItemId: review.workItemId,
		pullRequest: review.pullRequest,
		generation: review.generation,
		baseSha: review.baseSha,
		headSha: review.headSha,
		changedPaths: review.changedPaths,
		allowedScopes: review.allowedScopes,
		reviewMarker: review.idempotencyMarker,
		resultDigest: reviewDigest(review),
		completedAt: review.completedAt,
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

test("an exact clean review with controller-owned session attestation satisfies the route while movement invalidates it", () => {
	const review = cleanReview();
	assert.equal(reviewCoversExactRange(review, baseSha, headSha), true);
	assert.deepEqual(reconcileIndependentReview({ target: target(), reviews: [review], attestations: [attestation(review)] } as never), {
		kind: "satisfied",
		review,
	});

	for (const moved of [
		target({ headSha: "c".repeat(40) }),
		target({ baseSha: "d".repeat(40) }),
	]) {
		const decision = reconcileIndependentReview({ target: moved, reviews: [review], attestations: [attestation(review)] } as never);
		assert.equal(decision.kind, "dispatch");
		if (decision.kind === "dispatch") {
			assert.notEqual(decision.work.idempotencyMarker, review.idempotencyMarker);
		}
	}
});

test("reviewer-self-attested execution metadata cannot replace controller-owned session provenance", () => {
	const review = cleanReview();
	assert.equal(reconcileIndependentReview({ target: target(), reviews: [review] }).kind, "dispatch");
	for (const forged of [
		attestation(review, { authority: "reviewer" }),
		attestation(review, { provider: "anthropic" }),
		attestation(review, { model: "gpt-5.5" }),
		attestation(review, { reasoningEffort: "high" }),
		attestation(review, { readOnly: false }),
		attestation(review, { resultDigest: "0".repeat(64) }),
	]) {
		assert.throws(
			() => reconcileIndependentReview({ target: target(), reviews: [review], attestations: [forged] } as never),
			/attestation|session|provenance|digest|route|provider|model|read.only/i,
		);
	}
});

test("review generation is positive at target and record boundaries", () => {
	assert.throws(() => createIndependentReviewWork(target({ generation: 0 })), /generation|positive/i);
	assert.throws(() => validateIndependentReviewRecord({ ...cleanReview(), generation: 0 }), /generation|marker|positive/i);
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
