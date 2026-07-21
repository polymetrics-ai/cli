import assert from "node:assert/strict";
import { mkdtemp, readFile, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import {
	FileHumanDecisionRepository,
	consumeHumanDecision,
	createHumanDecisionRecord,
	persistHumanDecisionRequest,
	recordHumanDecision,
	routeHumanDecisionTarget,
	type HumanDecisionBinding,
	type HumanDecisionRequestSpec,
} from "./human-decision.ts";

const head = "a".repeat(40);
const issueBinding: HumanDecisionBinding = {
	repository: "polymetrics-ai/cli",
	target: { kind: "issue", number: 471 },
	generation: 3,
};
const prBinding: HumanDecisionBinding = {
	repository: "polymetrics-ai/cli",
	target: { kind: "pull_request", number: 477 },
	generation: 3,
	headSha: head,
};

function spec(overrides: Partial<HumanDecisionRequestSpec> = {}): HumanDecisionRequestSpec {
	return {
		requestId: "req-477",
		gate: "requirements",
		binding: issueBinding,
		allowedOptions: ["approve", "reject"],
		actorAllowlist: ["maintainer-one", "Maintainer-Two"],
		expiresAt: "2026-07-22T10:00:00.000Z",
		question: "Approve the exact requirements for issue #471?",
		...overrides,
	};
}

test("routes requirements and scope only to the parent issue", () => {
	for (const gate of ["requirements", "scope"] as const) {
		assert.deepEqual(routeHumanDecisionTarget(gate, 471, 477), { kind: "issue", number: 471 });
	}
});

test("routes review, head, merge, and distinct parent merge gates only to the PR", () => {
	for (const gate of ["review", "head", "merge", "parent_merge"] as const) {
		assert.deepEqual(routeHumanDecisionTarget(gate, 471, 477), { kind: "pull_request", number: 477 });
	}
});

test("requires exact head binding for every PR gate and rejects head binding for issue gates", () => {
	assert.throws(() => createHumanDecisionRecord(spec({ gate: "review", binding: { ...prBinding, headSha: undefined } })), /head/i);
	assert.throws(() => createHumanDecisionRecord(spec({ gate: "requirements", binding: { ...issueBinding, headSha: head } })), /head/i);
	assert.throws(() => createHumanDecisionRecord(spec({ gate: "parent_merge", binding: { ...prBinding, headSha: "b".repeat(39) } })), /head/i);
});

test("parent merge remains a distinct exact-head request", () => {
	const parentMerge = createHumanDecisionRecord(spec({ gate: "parent_merge", binding: prBinding }));
	const merge = createHumanDecisionRecord(spec({ gate: "merge", binding: prBinding }));
	assert.equal(parentMerge.gate, "parent_merge");
	assert.notEqual(parentMerge.idempotencyMarker, merge.idempotencyMarker);
});

test("canonicalizes actors/options and rejects duplicate, malformed, expired, or secret-bearing requests", () => {
	const record = createHumanDecisionRecord(spec());
	assert.deepEqual(record.actorAllowlist, ["maintainer-one", "maintainer-two"]);
	assert.deepEqual(record.allowedOptions, ["approve", "reject"]);
	for (const candidate of [
		spec({ allowedOptions: ["approve", "approve"] }),
		spec({ actorAllowlist: ["same", "SAME"] }),
		spec({ requestId: "../escape" }),
		spec({ expiresAt: "2026-07-21T09:59:59.000Z" }),
		spec({ question: "Authorization: Bearer secret-value" }),
		spec({ question: "token github_pat_1234567890123456789012" }),
	]) {
		assert.throws(() => createHumanDecisionRecord(candidate, new Date("2026-07-21T10:00:00.000Z")));
	}
});

test("durably round-trips and rejects a changed retry specification after restart", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-477-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const first = new FileHumanDecisionRepository(root, { lockRetryMs: 1 });
	const created = await persistHumanDecisionRequest(first, spec(), new Date("2026-07-21T10:00:00.000Z"));
	const restarted = new FileHumanDecisionRepository(root, { lockRetryMs: 1 });
	assert.deepEqual(await restarted.load("req-477"), created);
	await assert.rejects(
		persistHumanDecisionRequest(restarted, spec({ allowedOptions: ["approve", "defer"] })),
		/conflict|differs/i,
	);
	const serialized = await readFile(join(root, "req-477.json"), "utf8");
	assert.equal(serialized.includes("github_pat_"), false);
});

test("persists minimal accepted evidence and consumes it exactly once across repository instances", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-477-consume-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const first = new FileHumanDecisionRepository(root, { lockRetryMs: 1 });
	const second = new FileHumanDecisionRepository(root, { lockRetryMs: 1 });
	await persistHumanDecisionRequest(first, spec(), new Date("2026-07-21T10:00:00.000Z"));
	await recordHumanDecision(first, "req-477", issueBinding, {
		option: "approve",
		actor: "maintainer-one",
		sourceUrl: "https://github.com/polymetrics-ai/cli/issues/471#issuecomment-2001",
		decidedAt: "2026-07-21T10:01:00.000Z",
	});

	const attempts = await Promise.allSettled([
		consumeHumanDecision(first, "req-477", issueBinding, new Date("2026-07-21T10:02:00.000Z")),
		consumeHumanDecision(second, "req-477", issueBinding, new Date("2026-07-21T10:02:00.000Z")),
	]);
	assert.equal(attempts.filter((attempt) => attempt.status === "fulfilled").length, 1);
	assert.equal(attempts.filter((attempt) => attempt.status === "rejected").length, 1);
	const accepted = attempts.find((attempt) => attempt.status === "fulfilled");
	assert.equal(accepted?.status === "fulfilled" && accepted.value.option, "approve");
	assert.deepEqual(Object.keys(accepted?.status === "fulfilled" ? accepted.value : {}).sort(), [
		"actor", "decidedAt", "option", "sourceUrl",
	]);
	await assert.rejects(consumeHumanDecision(first, "req-477", issueBinding), /consum/i);
});

test("fails closed when generation, repository, target, or head differs at decision time", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-477-binding-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const repository = new FileHumanDecisionRepository(root, { lockRetryMs: 1 });
	await persistHumanDecisionRequest(repository, spec({ gate: "review", binding: prBinding }));
	for (const binding of [
		{ ...prBinding, generation: 4 },
		{ ...prBinding, repository: "other/repo" },
		{ ...prBinding, target: { kind: "pull_request" as const, number: 478 } },
		{ ...prBinding, headSha: "b".repeat(40) },
	]) {
		await assert.rejects(recordHumanDecision(repository, "req-477", binding, {
			option: "approve",
			actor: "maintainer-one",
			sourceUrl: "https://github.com/polymetrics-ai/cli/pull/477#issuecomment-1",
			decidedAt: "2026-07-21T10:01:00.000Z",
		}), /binding|stale|target/i);
	}
});
