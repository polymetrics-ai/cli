import assert from "node:assert/strict";
import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import {
	AgentSessionMvpLifecycle,
	createRepositoryScopedWorkspace,
	LocalParentMergeGate,
	RepositoryManifestIntake,
	type AgentSessionMvpRuntime,
} from "./autonomous-local-adapters.ts";
import type { AutonomousChildContext } from "./autonomous-controller.ts";
import type { RoleRunRequest } from "./agent-session-runtime.ts";
import { createToolPolicy } from "./tool-policy.ts";

function deferred<T>() {
	let resolve!: (value: T) => void;
	const promise = new Promise<T>((accept) => { resolve = accept; });
	return { promise, resolve };
}

function childContext(id: string, issue: number): AutonomousChildContext {
	return {
		parentIssue: 479,
		runId: "run-parallel",
		generation: 1,
		child: {
			id,
			issue,
			title: id,
			task: `implement ${id}`,
			dependsOn: [],
			access: "mutating",
			writeScopes: [`.pi/extensions/shepherd/${id}`],
		},
		timeoutMs: 30_000,
		signal: new AbortController().signal,
	};
}

test("loads the repository MVP manifest and keeps the parent merge as a local human wait", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-manifest-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	await mkdir(join(root, ".planning", "shepherd"), { recursive: true });
	await writeFile(join(root, ".planning", "shepherd", "issue-471.json"), JSON.stringify({
		schemaVersion: 1,
		parentIssue: 471,
		planId: "mvp-plan",
		children: [{
			id: "help",
			issue: 501,
			title: "Help",
			task: "Implement help and tests.",
			dependsOn: [],
			access: "mutating",
			writeScopes: ["cmd/help"],
		}],
	}));
	const plan = await new RepositoryManifestIntake(root).load(471, new AbortController().signal);
	assert.equal(plan.planId, "mvp-plan");
	assert.deepEqual(plan.children.map((child) => child.id), ["help"]);
	const state = {
		schemaVersion: 2,
		kind: "autonomous",
		issue: 471,
		planId: plan.planId,
		runId: "run-1",
		generation: 1,
		status: "running",
		stage: "HUMAN_DECISION",
		maxConcurrency: 2,
		timeoutMs: 900_000,
		createdAt: "2026-07-22T08:00:00Z",
		updatedAt: "2026-07-22T08:01:00Z",
		children: [],
	};
	const gate = new LocalParentMergeGate();
	assert.deepEqual(await gate.request(state, new AbortController().signal), {
		requestId: "local-parent-merge-471-1",
	});
	assert.equal(await gate.observe(state, new AbortController().signal), "pending");
	assert.equal("merge" in gate, false);
});

test("builds an exact plain workspace capability accepted by the hardened tool policy", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-workspace-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const workspace = createRepositoryScopedWorkspace("issue-501-alpha", root);
	assert.equal(Object.getPrototypeOf(workspace), Object.prototype);
	const policy = createToolPolicy({
		readOnly: false,
		workspace,
		authority: {
			workspaceId: workspace.id,
			readPrefixes: ["."],
			writePrefixes: [".pi/extensions/shepherd/alpha"],
			capabilityNames: [],
		},
		capabilities: [],
	});
	assert.deepEqual(policy.names, ["workspace_read", "workspace_edit", "workspace_write"]);
});

test("allocates one embedded runtime per child while reusing it across that child's roles", async () => {
	const gates: Array<ReturnType<typeof deferred<void>>> = [];
	const requests: RoleRunRequest[][] = [];
	let active = 0;
	let maxActive = 0;
	const lifecycle = new AgentSessionMvpLifecycle(() => {
		const index = requests.length;
		requests.push([]);
		const gate = deferred<void>();
		gates.push(gate);
		const runtime: AgentSessionMvpRuntime = {
			async run(request) {
				requests[index].push(request);
				active += 1;
				maxActive = Math.max(maxActive, active);
				await gate.promise;
				active -= 1;
				return {
					schemaVersion: 1,
					role: request.role,
					status: "completed",
					summary: `${request.role} complete`,
					observedMutation: request.role === "implementation",
					changedPaths: [],
					verification: [],
					findings: [],
					...request.binding,
				};
			},
			async abort() {},
			async close() {},
		};
		return runtime;
	}, process.cwd());

	const alpha = lifecycle.execute(childContext("alpha", 501));
	const beta = lifecycle.execute(childContext("beta", 502));
	while (active < 2) await new Promise((resolve) => setImmediate(resolve));
	assert.equal(requests.length, 2);
	assert.equal(maxActive, 2);
	assert.notEqual(requests[0][0].workspace.id, requests[1][0].workspace.id);
	gates[0].resolve();
	gates[1].resolve();
	await Promise.all([alpha, beta]);

	const verification = lifecycle.verify(childContext("alpha", 501));
	while (requests[0].length < 2) await new Promise((resolve) => setImmediate(resolve));
	assert.equal(requests.length, 2, "the same child lane must retain its isolated runtime");
	await verification;
	await lifecycle.close();
});
