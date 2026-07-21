import assert from "node:assert/strict";
import test from "node:test";

import {
	createIndependentReviewWork,
	createAgentSessionAttestation,
	independentReviewResultDigest,
	reconcileIndependentReview,
	reviewCoversExactRange,
	validateAgentSessionAttestation,
	validateIndependentReviewRecord,
	type IndependentReviewRecord,
	type IndependentReviewTarget,
} from "./review-router.ts";
import * as reviewRouterApi from "./review-router.ts";

const baseSha = "a".repeat(40);
const headSha = "b".repeat(40);

function target(overrides: Partial<IndependentReviewTarget> = {}): IndependentReviewTarget {
	return {
		repository: "github.com/polymetrics-ai/cli",
		workItemId: "issue-478",
		pullRequest: 812,
		generation: 3,
		baseBranch: "feat/471-pi-agent-session-shepherd",
		headBranch: "feat/811-github-evidence",
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

function attestation(review: IndependentReviewRecord, overrides: Record<string, unknown> = {}) {
	return {
		...createAgentSessionAttestation({
			sessionId: "session-478-review",
			runId: "run-478-review-1",
			review,
		}),
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

test("exports one canonical controller attestation digest, constructor, and validator", () => {
	assert.equal(typeof independentReviewResultDigest, "function");
	assert.equal(typeof createAgentSessionAttestation, "function");
	assert.equal(typeof validateAgentSessionAttestation, "function");
	const review = cleanReview();
	const digest = independentReviewResultDigest(review);
	const created = createAgentSessionAttestation({
		sessionId: "session-478-exported-api",
		runId: "run-478-exported-api-1",
		review,
	});
	assert.equal(created.resultDigest, digest);
	assert.deepEqual(validateAgentSessionAttestation(JSON.parse(JSON.stringify(created))), created);
	assert.throws(
		() => validateAgentSessionAttestation({
			...created,
			resultDigest: "0".repeat(64),
		}, review),
		/digest|attestation/i,
	);
});

test("binds review and attestation targets to exact base and head branches", () => {
	const branched = {
		...target(),
		baseBranch: "feat/471-pi-agent-session-shepherd",
		headBranch: "feat/811-github-evidence",
	};
	const work = createIndependentReviewWork(branched as never);
	assert.equal(work.baseBranch, branched.baseBranch);
	assert.equal(work.headBranch, branched.headBranch);
	assert.notEqual(
		createIndependentReviewWork({ ...branched, baseBranch: "main" } as never).idempotencyMarker,
		work.idempotencyMarker,
	);
});

test("same-marker attempts bind attestations by digest and target independent of ordering", () => {
	const older = cleanReview({ completedAt: "2026-07-21T11:59:00.000Z" });
	const newer = cleanReview({ completedAt: "2026-07-21T12:00:00.000Z" });
	for (const reviews of [[older, newer], [newer, older]]) {
		for (const attestations of [
			[attestation(older, { sessionId: "session-older", runId: "run-older" }), attestation(newer, { sessionId: "session-newer", runId: "run-newer" })],
			[attestation(newer, { sessionId: "session-newer", runId: "run-newer" }), attestation(older, { sessionId: "session-older", runId: "run-older" })],
		]) {
			const decision = reconcileIndependentReview({ target: target(), reviews, attestations } as never);
			assert.equal(decision.kind, "satisfied");
			if (decision.kind === "satisfied") assert.equal(decision.review.completedAt, newer.completedAt);
		}
	}

	const ambiguousA = {
		...cleanReview(),
		verdict: "findings" as const,
		findings: [{ id: "A", severity: "blocking" as const, summary: "First result." }],
	};
	const ambiguousB = {
		...cleanReview(),
		verdict: "findings" as const,
		findings: [{ id: "B", severity: "blocking" as const, summary: "Second result." }],
	};
	assert.throws(() => reconcileIndependentReview({
		target: target(),
		reviews: [ambiguousA, ambiguousB],
		attestations: [
			attestation(ambiguousA, { sessionId: "session-a", runId: "run-a" }),
			attestation(ambiguousB, { sessionId: "session-b", runId: "run-b" }),
		],
	} as never), /ambiguous|same.marker|digest/i);
});

test("cycle 4 rejects every pseudo or symbolic Git ref at review and attestation boundaries", () => {
	for (const invalid of [
		"HEAD", "FETCH_HEAD", "ORIG_HEAD", "MERGE_HEAD", "CHERRY_PICK_HEAD", "REVERT_HEAD",
		"REBASE_HEAD", "BISECT_HEAD", "AUTO_MERGE", "topic/FETCH_HEAD", "refs/heads/topic",
	]) {
		assert.throws(() => createIndependentReviewWork(target({ baseBranch: invalid })), /branch|ref|pseudo|symbolic/i, invalid);
		assert.throws(() => createIndependentReviewWork(target({ headBranch: invalid })), /branch|ref|pseudo|symbolic/i, invalid);
		assert.throws(() => validateIndependentReviewRecord({ ...cleanReview(), baseBranch: invalid }), /branch|ref|marker|review/i, invalid);
	}
});

test("cycle 5 rejects cookie and session response-header forms through the shared grammar", () => {
	const assertSafe = (reviewRouterApi as Record<string, unknown>).assertNoSensitiveText as
		| ((value: string, description?: string) => void)
		| undefined;
	assert.equal(typeof assertSafe, "function");
	for (const value of [
		"Set-Cookie: session_id=SYNTHETIC_SESSION_VALUE; HttpOnly; Secure",
		"Cookie: sid=SYNTHETIC_COOKIE_VALUE; theme=dark",
		"X-Session-Token: SYNTHETIC_HEADER_VALUE",
		"X-CSRF-Token: SYNTHETIC_CSRF_VALUE",
		"session cookie=SYNTHETIC_PROSE_VALUE",
	]) {
		assert.throws(() => assertSafe!(value, "cycle 5 synthetic field"), /credential|secret|sensitive/i, value);
		assert.notEqual((reviewRouterApi.redactSensitiveText as (input: string) => string)(value), value, value);
	}

	const review = cleanReview({
		verdict: "findings",
		findings: [{
			id: "cycle-5-cookie-finding",
			severity: "warning",
			summary: "Set-Cookie: session=SYNTHETIC_REVIEW_VALUE; HttpOnly",
		}],
	});
	assert.throws(() => validateIndependentReviewRecord(review), /credential|secret|sensitive/i);
});

test("cycle 5 byte-bounds raw JSON and schema-bounds oversized objects before descriptor expansion", () => {
	const decode = (reviewRouterApi as Record<string, unknown>).decodeBoundedJsonPayload as
		| ((value: string | Uint8Array, maximumBytes?: number) => unknown)
		| undefined;
	assert.equal(typeof decode, "function");
	assert.deepEqual(decode!("{\"schemaVersion\":1}", 64), { schemaVersion: 1 });
	assert.throws(
		() => decode!(`{\"payload\":\"${"x".repeat(1_024)}\"}`, 128),
		/byte|bound|oversize|payload/i,
	);

	const oversized: Record<string, unknown> = { ...target() };
	for (let index = 0; index < 10_000; index += 1) oversized[`extra-${index}`] = index;
	const original = Object.getOwnPropertyDescriptors;
	let expanded = false;
	Object.getOwnPropertyDescriptors = ((value: object) => {
		if (value === oversized) expanded = true;
		return original(value);
	}) as typeof Object.getOwnPropertyDescriptors;
	try {
		assert.throws(() => createIndependentReviewWork(oversized as unknown as IndependentReviewTarget), /unknown|bound|shape|field/i);
	} finally {
		Object.getOwnPropertyDescriptors = original;
	}
	assert.equal(expanded, false, "oversized envelope must reject before generic descriptor expansion");
});

test("cycle 4 uses collision-free session and run tuple identities", () => {
	const older = cleanReview({ completedAt: "2026-07-21T11:59:00.000Z" });
	const newer = cleanReview({ completedAt: "2026-07-21T12:00:00.000Z" });
	const distinct = [
		attestation(older, { sessionId: "a:b", runId: "c" }),
		attestation(newer, { sessionId: "a", runId: "b:c" }),
	];
	assert.equal(reconcileIndependentReview({ target: target(), reviews: [older, newer], attestations: distinct } as never).kind, "satisfied");
	assert.throws(() => reconcileIndependentReview({
		target: target(),
		reviews: [older, newer],
		attestations: [
			attestation(older, { sessionId: "same", runId: "tuple" }),
			attestation(newer, { sessionId: "same", runId: "tuple" }),
		],
	} as never), /duplicate|session|run|tuple/i);
});

test("cycle 4 rejects oversized arrays before materializing all property descriptors", () => {
	const oversized = Array.from({ length: 65 }, (_, index) => `src/${index}.ts`);
	const original = Object.getOwnPropertyDescriptors;
	let traversed = false;
	let rejection: unknown;
	Object.getOwnPropertyDescriptors = ((value: object) => {
		if (value === oversized) {
			traversed = true;
			throw new Error("descriptor traversal must not occur");
		}
		return original(value);
	}) as typeof Object.getOwnPropertyDescriptors;
	try {
		createIndependentReviewWork(target({ changedPaths: oversized }));
	} catch (error) {
		rejection = error;
	} finally {
		Object.getOwnPropertyDescriptors = original;
	}
	assert.equal(traversed, false);
	assert.match(String(rejection), /bounded|paths|64/i);
});

test("cycle 6 orders every attested exact-head attempt and separates stable clean authorization", () => {
	const olderClean = cleanReview({ completedAt: "2026-07-21T12:00:00.000Z" });
	const laterFindings = cleanReview({
		completedAt: "2026-07-21T12:01:00.000Z",
		verdict: "findings",
		findings: [{ id: "cycle-6-later-blocker", severity: "blocking", summary: "Later exact-head blocker." }],
	});
	const latestClean = cleanReview({ completedAt: "2026-07-21T12:02:00.000Z" });
	const attestations = (reviews: IndependentReviewRecord[]) => reviews.map((review, index) => attestation(review, {
		sessionId: `cycle-6-session-${index}`,
		runId: `cycle-6-run-${index}`,
	}));

	assert.equal(reconcileIndependentReview({
		target: target(),
		reviews: [olderClean, laterFindings],
		attestations: attestations([olderClean, laterFindings]),
	} as never).kind, "dispatch", "a later findings attempt must invalidate earlier clean authority");

	const recovered = reconcileIndependentReview({
		target: target(),
		reviews: [laterFindings, latestClean, olderClean],
		attestations: attestations([laterFindings, latestClean, olderClean]),
	} as never);
	assert.equal(recovered.kind, "satisfied");
	if (recovered.kind === "satisfied") assert.equal(recovered.review.completedAt, latestClean.completedAt);

	const semanticDigest = (reviewRouterApi as Record<string, unknown>).independentReviewAuthorizationDigest as
		| ((review: IndependentReviewRecord) => string)
		| undefined;
	assert.equal(typeof semanticDigest, "function");
	assert.notEqual(independentReviewResultDigest(olderClean), independentReviewResultDigest(latestClean));
	assert.equal(semanticDigest!(olderClean), semanticDigest!(latestClean));
});

test("cycle 6 pre-bounds exact Uint8Array receivers and normalizes revoked proxy host errors", () => {
	const payload = new Uint8Array(128);
	Object.defineProperty(payload, "byteLength", { value: 1, enumerable: true });
	const originalDecode = TextDecoder.prototype.decode;
	let decoded = false;
	TextDecoder.prototype.decode = function (...args: Parameters<TextDecoder["decode"]>): string {
		decoded = true;
		return originalDecode.apply(this, args);
	};
	try {
		assert.throws(
			() => reviewRouterApi.decodeBoundedJsonPayload(payload, 16),
			/oversized|byte|bound/i,
		);
	} finally {
		TextDecoder.prototype.decode = originalDecode;
	}
	assert.equal(decoded, false, "physical byte length must reject before decode");

	const revoked = Proxy.revocable({}, {});
	revoked.revoke();
	let rejection: unknown;
	try {
		reviewRouterApi.readBoundedExactRecord(revoked.proxy, [], [], "cycle 6 record");
	} catch (error) {
		rejection = error;
	}
	assert.ok(rejection instanceof Error);
	assert.match(String(rejection), /invalid|bounded|shape|proxy/i);
	assert.doesNotMatch(String(rejection), /Cannot perform ['"]?IsArray|revoked/i);
});

test("cycle 6 shared credential grammar covers standard credential-file forms", () => {
	const samples = [
		"//registry.invalid/:_authToken=SYNTHETIC_NPM_MARKER",
		"password SYNTHETIC_NETRC_MARKER",
		"aws_secret_access_key = SYNTHETIC_AWS_MARKER",
		"azure_client_secret=SYNTHETIC_AZURE_MARKER",
		"credentials_file = /tmp/SYNTHETIC_CREDENTIAL_FILE",
	];
	for (const sample of samples) {
		assert.throws(() => reviewRouterApi.assertNoSensitiveText(sample, "cycle 6 credential fixture"), /credential|secret|sensitive/i, sample);
		assert.notEqual(reviewRouterApi.redactSensitiveText(sample), sample, sample);
	}
});
