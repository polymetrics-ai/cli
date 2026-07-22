import assert from "node:assert/strict";
import test from "node:test";

import {
	AgentSessionProductionReviewAdapter,
	GhProductionReviewRepository,
	MemoryProductionReviewRepository,
	type ProductionReviewSession,
} from "./production-review-adapter.ts";
import type { ExternalCallContext } from "./github-orchestrator.ts";
import type { IndependentReviewTarget, IndependentReviewWork } from "./review-router.ts";

const BASE = "1".repeat(40);
const HEAD = "2".repeat(40);

function context(): ExternalCallContext {
	return {
		signal: new AbortController().signal,
		deadlineAt: new Date(Date.now() + 60_000).toISOString(),
		acknowledgeAbort() {},
	};
}

function target(headSha = HEAD): IndependentReviewTarget {
	return {
		repository: "acme/widgets",
		workItemId: "child-a",
		pullRequest: 42,
		generation: 3,
		baseBranch: "feat/parent",
		headBranch: "feat/42-child-a",
		baseSha: BASE,
		headSha,
		changedPaths: ["src/a.ts"],
		allowedScopes: ["src"],
	};
}

class Session implements ProductionReviewSession {
	runs: IndependentReviewWork[] = [];
	async run(work: IndependentReviewWork) {
		this.runs.push(work);
		return {
			sessionId: `session-${this.runs.length}`,
			runId: `run-${this.runs.length}`,
			completedAt: `2026-07-22T00:00:0${this.runs.length}.000Z`,
			verdict: "clean" as const,
			findings: [],
		};
	}
}

test("runs an exact-range xhigh read-only AgentSession and reuses its durable attested result", async () => {
	const repository = new MemoryProductionReviewRepository();
	const session = new Session();
	const adapter = new AgentSessionProductionReviewAdapter(repository, session);
	const first = await adapter.review(target(), context());
	const second = await adapter.review(target(), context());
	assert.equal(first.review.verdict, "clean");
	assert.equal(first.review.reasoningEffort, "xhigh");
	assert.equal(first.review.readOnly, true);
	assert.equal(session.runs.length, 1);
	assert.deepEqual(second, first);
	assert.deepEqual((await adapter.findAttestations(target(), context())).items, [first.attestation]);
});

test("head movement invalidates prior review and dispatches a fresh exact-range AgentSession", async () => {
	const repository = new MemoryProductionReviewRepository();
	const session = new Session();
	const adapter = new AgentSessionProductionReviewAdapter(repository, session);
	await adapter.review(target(), context());
	const moved = await adapter.review(target("3".repeat(40)), context());
	assert.equal(session.runs.length, 2);
	assert.equal(moved.review.headSha, "3".repeat(40));
	assert.notEqual(moved.review.idempotencyMarker, session.runs[0].idempotencyMarker);
});

test("stores structured findings and exact-head dispositions without converting them to clean review authority", async () => {
	const repository = new MemoryProductionReviewRepository();
	const session: ProductionReviewSession = {
		async run() {
			return {
				sessionId: "session-findings",
				runId: "run-findings",
				completedAt: "2026-07-22T00:00:01.000Z",
				verdict: "findings",
				findings: [{ id: "F-1", severity: "blocking", summary: "unsafe mutation" }],
			};
		},
	};
	const adapter = new AgentSessionProductionReviewAdapter(repository, session);
	const artifact = await adapter.review(target(), context());
	assert.equal(artifact.review.verdict, "findings");
	const dispositioned = await repository.recordDispositions(target(), [{
		findingId: "F-1",
		kind: "fixed",
		rationale: "fixed on the exact reviewed head",
		actor: "controller",
		headSha: HEAD,
		recordedAt: "2026-07-22T00:00:02.000Z",
	}], context());
	assert.equal(dispositioned.dispositions[0].findingId, "F-1");
	assert.equal((await adapter.review(target(), context())).review.verdict, "findings");
	assert.equal((await adapter.review(target("4".repeat(40)), context())).dispositions.length, 0);
});

test("GitHub marker repository survives restart, reuses publication, and rejects duplicate durable evidence", async () => {
	const comments: Array<Record<string, unknown>> = [];
	let nextId = 100;
	const execute = async (_file: "gh", args: readonly string[]) => {
		const method = args[args.indexOf("--method") + 1];
		const endpoint = args[3];
		if (method === "GET") return JSON.stringify(comments);
		const bodyArg = args[args.indexOf("-f") + 1];
		const body = bodyArg.slice("body=".length);
		if (method === "POST") {
			const comment = { id: nextId++, body, updated_at: "2026-07-22T00:00:02.000Z" };
			comments.push(comment);
			return JSON.stringify(comment);
		}
		if (method === "PATCH") {
			const id = Number(endpoint.split("/").at(-1));
			const comment = comments.find((entry) => entry.id === id)!;
			comment.body = body;
			return JSON.stringify(comment);
		}
		throw new Error("unexpected call");
	};
	const session = new Session();
	const first = new AgentSessionProductionReviewAdapter(new GhProductionReviewRepository({ execute }), session);
	const published = await first.review(target(), context());
	const restarted = new AgentSessionProductionReviewAdapter(new GhProductionReviewRepository({ execute }), session);
	assert.deepEqual(await restarted.review(target(), context()), published);
	assert.equal(session.runs.length, 1);
	comments.push({ ...comments.at(-1), id: nextId++ });
	await assert.rejects(restarted.review(target(), context()), /duplicate|ambiguous|revision/i);
});
