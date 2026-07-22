import assert from "node:assert/strict";
import { createHash } from "node:crypto";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import { ProductionEffectJournal, productionEffectKey } from "./autonomous-effect-journal.ts";
import {
	createProductionAutonomousState,
	ProductionFileStateStore,
	validateProductionAutonomousState,
	type ProductionAutonomousState,
} from "./autonomous-production-state.ts";
import type {
	ProductionEffectKind,
	ProductionEffectRecord,
	ProductionParentPlanDocument,
} from "./autonomous-production-contract.ts";
import { ProductionRecoveryBarrier } from "./autonomous-recovery.ts";
import {
	ProductionEffectRecoveryAuthority,
	type ProductionRecoveryProbe,
	type ProductionRecoveryProbeTable,
} from "./production-effect-recovery-authority.ts";

const KINDS: ProductionEffectKind[] = [
	"workspace_claim", "agent_implementation", "agent_correction", "shell_verification", "git_commit", "git_push",
	"child_pull_request", "independent_review", "child_integration", "parent_refresh", "child_head_reconciliation",
	"human_request", "human_consume", "parent_merge_observation",
];
const SHA = "a".repeat(40);

function plan(): ProductionParentPlanDocument {
	return {
		schemaVersion: 2,
		planId: "production-recovery-matrix",
		parentIssue: 479,
		repository: "acme/widgets",
		title: "Production recovery matrix",
		objective: "Recover every external effect exactly once.",
		parentBranch: "feat/479-parent",
		parentBaseBranch: "main",
		actorAllowlist: ["maintainer"],
		decisionExpiresAt: "2099-07-22T00:00:00.000Z",
		children: [{
			id: "state",
			issue: 480,
			title: "State lane",
			task: "Implement the state lane.",
			slug: "state-lane",
			dependsOn: [],
			access: "mutating",
			writeScopes: [".pi/extensions/shepherd"],
			requiredSkills: ["golang-testing"],
			verification: [{
				id: "focused",
				executable: "node",
				args: ["--test"],
				cwd: ".",
				timeoutMs: 30_000,
				maxOutputBytes: 65_536,
			}],
			humanGates: [],
			maxAttempts: 2,
			maxCorrections: 1,
		}],
	};
}

function currentState(kind: ProductionEffectKind): ProductionAutonomousState {
	const state = createProductionAutonomousState(plan(), {
		runId: "run-1",
		now: new Date("2026-07-22T10:00:00.000Z"),
		maxConcurrency: 1,
		timeoutMs: 30_000,
	});
	state.stage = "child_lifecycle";
	state.children[0].status = "running";
	state.children[0].stage = "workspace";
	state.children[0].attempts = 1;
	if (kind === "human_consume") {
		state.status = "waiting_human";
		state.stage = "human_decision";
		state.children[0].status = "blocked";
		state.childGate = {
			childId: "state",
			repository: state.repository,
			issue: 480,
			generation: 1,
			requestId: "request-479",
			reason: "retry_budget_exhausted",
			status: "pending",
		};
	}
	if (kind === "parent_merge_observation") {
		state.status = "waiting_human";
		state.stage = "human_decision";
		state.children[0].status = "succeeded";
		state.children[0].stage = "succeeded";
		state.children[0].checkpoint = {
			summary: "child integrated before parent gate",
			integrationReceiptDigest: "c".repeat(64),
		};
		state.humanGate = {
			repository: state.repository,
			pullRequest: 438,
			generation: 1,
			head: SHA,
			requestId: "request-479",
			status: "pending",
		};
	}
	return validateProductionAutonomousState(state);
}

function canonical(value: unknown): unknown {
	if (Array.isArray(value)) return value.map(canonical);
	if (value !== null && typeof value === "object") {
		return Object.fromEntries(Object.entries(value as Record<string, unknown>)
			.sort(([left], [right]) => left.localeCompare(right))
			.map(([key, item]) => [key, canonical(item)]));
	}
	return value;
}

function descriptor(kind: ProductionEffectKind): Record<string, unknown> {
	if (kind === "human_request" || kind === "human_consume") {
		return { generation: 1, head: SHA, pullRequest: 438, repository: "acme/widgets", requestId: "request-479" };
	}
	if (kind === "parent_merge_observation") return {
		operation: "parent_merge_observation",
		generation: 1,
		head: SHA,
		pullRequest: 438,
		repository: "acme/widgets",
		requestId: "request-479",
		stateRevision: 1,
	};
	return { kind, marker: `recover-${kind}` };
}

function makeIntent(kind: ProductionEffectKind) {
	const recoveryDescriptor = canonical(descriptor(kind));
	const intentDigest = createHash("sha256").update(JSON.stringify(recoveryDescriptor)).digest("hex");
	const coordinates = {
		kind,
		runId: "run-1",
		generation: 1,
		...(kind === "parent_merge_observation" ? {} : { childId: "state" }),
		intentDigest,
	};
	return { key: productionEffectKey(coordinates), ...coordinates, recoveryDescriptor };
}

function projectedState(kind: ProductionEffectKind, record: ProductionEffectRecord, current: ProductionAutonomousState) {
	const projected = structuredClone(current);
	projected.revision += 1;
	projected.updatedAt = "2026-07-22T10:00:01.000Z";
	if (kind === "human_request") {
		projected.status = "waiting_human";
		projected.stage = "human_decision";
		projected.children[0].status = "blocked";
		projected.childGate = {
			childId: "state",
			repository: projected.repository,
			issue: 480,
			generation: 1,
			requestId: "request-479",
			reason: "retry_budget_exhausted",
			status: "pending",
		};
	} else if (kind === "human_consume") {
		projected.childGate!.status = "aborted";
		projected.status = "failed";
		projected.stage = "blocked";
		projected.terminalBlocker = "human aborted the exhausted child";
		projected.children[0].status = "failed";
		projected.children[0].stage = "failed";
	} else if (kind === "parent_merge_observation") {
		projected.status = "completed";
		projected.stage = "completed";
		projected.humanGate!.status = "merged";
		projected.humanGate!.mergeEvidence = {
			mergedAt: "2026-07-22T10:00:00.500Z",
			mergeCommitSha: "b".repeat(40),
			revision: 9,
			observedAt: "2026-07-22T10:00:00.750Z",
		};
	} else {
		projected.children[0].checkpoint = {
			summary: `authoritatively recovered ${kind}`,
			effectKey: record.key,
			effectKeys: [record.key],
		};
	}
	return validateProductionAutonomousState(projected);
}

function probes(probe: ProductionRecoveryProbe): ProductionRecoveryProbeTable {
	return Object.fromEntries(KINDS.map((kind) => [kind, probe])) as ProductionRecoveryProbeTable;
}

type CrashWindow = "before_call" | "after_effect" | "after_observe" | "after_state_cas";

for (const kind of KINDS) {
	for (const crash of ["before_call", "after_effect", "after_observe", "after_state_cas"] as CrashWindow[]) {
		test(`${kind}: fresh-process recovery at ${crash} performs zero duplicate external mutations`, async (t) => {
			const root = await mkdtemp(join(tmpdir(), `shepherd-authority-${kind}-`));
			t.after(() => rm(root, { recursive: true, force: true }));
			const stateStore = new ProductionFileStateStore(join(root, "state"));
			const initial = await stateStore.create(currentState(kind));
			const journal = new ProductionEffectJournal(join(root, "state"));
			const intent = makeIntent(kind);
			let record = await journal.prepare(intent);
			let externalApplied = false;
			let mutations = 0;
			const mutate = () => {
				if (externalApplied) throw new Error("duplicate external mutation");
				externalApplied = true;
				mutations += 1;
			};
			const resultDigest = createHash("sha256").update(`result:${kind}`).digest("hex");
			const probe: ProductionRecoveryProbe = async ({ record: exact, currentState: current }) => externalApplied
				? { status: "applied", resultDigest, projectedState: projectedState(kind, exact, current) }
				: { status: "absent" };
			const authority = () => new ProductionEffectRecoveryAuthority({
				stateRoot: join(root, "state"), issue: 479, stateStore, probes: probes(probe),
			});

			if (crash !== "before_call") mutate();
			if (crash === "after_observe" || crash === "after_state_cas") {
				const observed = await authority().observe(record, new AbortController().signal);
				assert.equal(observed.status, "applied");
				record = await journal.observe(record.key, { runId: "run-1", generation: 1 }, resultDigest);
			}
			if (crash === "after_state_cas") await authority().apply(record, new AbortController().signal);

			const restarted = new ProductionRecoveryBarrier(new ProductionEffectJournal(join(root, "state")), authority());
			await restarted.open({ runId: "run-1", generation: 1 });
			if (crash === "before_call") {
				assert.equal(await journal.load(intent.key), undefined, "absence resets the WAL before replay");
				mutate();
				record = await journal.prepare(intent);
				const observed = await authority().observe(record, new AbortController().signal);
				assert.equal(observed.status, "applied");
				record = await journal.observe(record.key, { runId: "run-1", generation: 1 }, resultDigest);
				await authority().apply(record, new AbortController().signal);
				await journal.apply(record.key, { runId: "run-1", generation: 1 });
			}

			assert.equal(mutations, 1);
			assert.equal((await journal.load(intent.key))?.phase, "applied");
			assert.notDeepEqual(await stateStore.load(479), initial);
			assert.deepEqual(
				await new ProductionRecoveryBarrier(new ProductionEffectJournal(join(root, "state")), authority())
					.open({ runId: "run-1", generation: 1 }),
				{ reconciled: 0 },
			);
			assert.equal(mutations, 1);
		});
	}
}

for (const observation of ["pending", "approved_waiting_for_merge", "rejected", "invalidated"] as const) {
	test(`parent merge recovery binds exact ${observation} revision projection`, async (t) => {
		const root = await mkdtemp(join(tmpdir(), `shepherd-parent-observation-${observation}-`));
		t.after(() => rm(root, { recursive: true, force: true }));
		const stateRoot = join(root, "state");
		const store = new ProductionFileStateStore(stateRoot);
		const current = await store.create(currentState("parent_merge_observation"));
		const journal = new ProductionEffectJournal(stateRoot);
		let record = await journal.prepare(makeIntent("parent_merge_observation"));
		const projected = structuredClone(current);
		projected.revision += 1;
		projected.updatedAt = "2026-07-22T10:00:01.000Z";
		if (observation === "rejected") {
			projected.status = "failed";
			projected.stage = "blocked";
			projected.terminalBlocker = "human rejected the exact parent merge";
			projected.humanGate!.status = "rejected";
		} else if (observation === "invalidated") {
			projected.status = "running";
			projected.stage = "schedule";
			projected.invalidatedParentGates = [{
				...projected.humanGate!,
				status: "invalidated",
				invalidationEvidence: {
					currentHead: "b".repeat(40),
					revision: 9,
					observedAt: "2026-07-22T10:00:00.750Z",
				},
			}];
			delete projected.humanGate;
		}
		const exactProjected = validateProductionAutonomousState(projected);
		const resultDigest = createHash("sha256").update(`parent:${observation}`).digest("hex");
		const authority = new ProductionEffectRecoveryAuthority({
			stateRoot,
			issue: 479,
			stateStore: store,
			probes: probes(async () => ({ status: "applied", resultDigest, projectedState: exactProjected })),
		});
		assert.deepEqual(await authority.observe(record, new AbortController().signal), {
			status: "applied",
			resultDigest,
		});
		record = await journal.observe(record.key, { runId: record.runId, generation: record.generation }, resultDigest);
		await authority.apply(record, new AbortController().signal);
		await journal.apply(record.key, { runId: record.runId, generation: record.generation });
		assert.deepEqual(await store.load(479), exactProjected);
	});
}

test("recovery authority rejects a non-exhaustive probe table before any evidence read", async () => {
	assert.throws(() => new ProductionEffectRecoveryAuthority({
		stateRoot: "/tmp/shepherd-incomplete-recovery",
		issue: 479,
		stateStore: { async load() { return undefined; }, async create(value) { return value; }, async compareAndSwap(_fence, value) { return value; } },
		probes: { git_push: async () => ({ status: "absent" }) } as never,
	}), /exhaustive|probe/i);
});
