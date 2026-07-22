import assert from "node:assert/strict";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import type { ProductionParentPlanDocument } from "./autonomous-production-contract.ts";
import { ProductionEffectJournal } from "./autonomous-effect-journal.ts";
import {
	createProductionAutonomousState,
	evolveProductionState,
	type ProductionAutonomousState,
} from "./autonomous-production-state.ts";
import { ProductionParentMergeEffectJournal } from "./production-parent-merge-effect.ts";

const HEAD = "a".repeat(40);

function plan(): ProductionParentPlanDocument {
	return {
		schemaVersion: 2,
		planId: "parent-merge-effect",
		parentIssue: 479,
		repository: "owner/repo",
		title: "Crash-safe parent observation",
		objective: "Observe the exact parent gate once",
		parentBranch: "feat/parent",
		parentBaseBranch: "main",
		actorAllowlist: ["maintainer"],
		decisionExpiresAt: "2026-08-01T00:00:00.000Z",
		children: [{
			id: "child",
			issue: 480,
			title: "child",
			task: "complete child",
			slug: "child",
			dependsOn: [],
			access: "mutating",
			writeScopes: ["owned/child"],
			requiredSkills: ["javascript-testing-patterns"],
			verification: [{ id: "tests", executable: "node", args: ["--test"], cwd: ".", timeoutMs: 30_000, maxOutputBytes: 1_000_000 }],
			humanGates: [],
			maxAttempts: 1,
			maxCorrections: 1,
		}],
	};
}

function waitingState(manifest: ProductionParentPlanDocument): ProductionAutonomousState {
	const initial = createProductionAutonomousState(manifest, {
		runId: "run-parent-merge",
		now: new Date("2026-07-22T10:00:00.000Z"),
	});
	return evolveProductionState(initial, {
		issue: initial.parentIssue,
		revision: initial.revision,
		generation: initial.generation,
		runId: initial.runId,
	}, (draft) => {
		draft.children[0].status = "succeeded";
		draft.children[0].stage = "succeeded";
		draft.children[0].checkpoint = {
			summary: "integrated",
			integrationReceiptDigest: "b".repeat(64),
		};
		draft.status = "waiting_human";
		draft.stage = "human_decision";
		draft.humanGate = {
			repository: manifest.repository,
			pullRequest: 438,
			generation: 1,
			head: HEAD,
			requestId: "parent-merge-request",
			status: "pending",
		};
	}, new Date("2026-07-22T10:00:01.000Z"));
}

test("parent merge effect is observed durably and applied only after its exact controller CAS", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-parent-merge-effect-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const manifest = plan();
	const before = waitingState(manifest);
	let externalObservations = 0;
	const journal = new ProductionEffectJournal(root);
	const adapter = new ProductionParentMergeEffectJournal({
		journal,
		parentGate: {
			prepare() { return { requestId: "parent-merge-request" }; },
			async request() { return { requestId: "parent-merge-request" }; },
			async observe() {
				externalObservations += 1;
				return {
					status: "merged" as const,
					repository: manifest.repository,
					pullRequest: 438,
					head: HEAD,
					mergedAt: "2026-07-22T10:00:02.000Z",
					mergeCommitSha: "c".repeat(40),
					revision: 9,
					observedAt: "2026-07-22T10:00:03.000Z",
				};
			},
			async close() {},
		},
	});

	const effect = await adapter.observe(manifest, before, new AbortController().signal);
	assert.equal(externalObservations, 1);
	assert.equal(effect.observation.status, "merged");
	const observed = await journal.load(effect.effectKey);
	assert.equal(observed?.phase, "observed");
	assert.equal(observed?.kind, "parent_merge_observation");
	assert.equal(observed?.runId, before.runId);
	assert.equal(observed?.generation, before.generation);
	assert.deepEqual(observed?.recoveryDescriptor, {
		operation: "parent_merge_observation",
		parentIssue: before.parentIssue,
		repository: before.repository,
		planId: before.planId,
		planDigest: before.planDigest,
		parentBranch: before.parentBranch,
		parentBaseBranch: before.parentBaseBranch,
		runId: before.runId,
		resourceGeneration: before.resourceGeneration,
		generation: before.generation,
		stateRevision: before.revision,
		pullRequest: before.humanGate!.pullRequest,
		requestId: before.humanGate!.requestId,
		head: before.humanGate!.head,
	});
	await assert.rejects(adapter.acknowledge(effect.effectKey, before), /exact controller CAS/i);
	assert.equal((await journal.load(effect.effectKey))?.phase, "observed");

	const after = evolveProductionState(before, {
		issue: before.parentIssue,
		revision: before.revision,
		generation: before.generation,
		runId: before.runId,
	}, (draft) => {
		draft.humanGate!.status = "merged";
		draft.humanGate!.mergeEvidence = {
			mergedAt: "2026-07-22T10:00:02.000Z",
			mergeCommitSha: "c".repeat(40),
			revision: 9,
			observedAt: "2026-07-22T10:00:03.000Z",
		};
		draft.status = "completed";
		draft.stage = "completed";
	}, new Date("2026-07-22T10:00:04.000Z"));
	await adapter.acknowledge(effect.effectKey, after);
	assert.equal((await journal.load(effect.effectKey))?.phase, "applied");
});
