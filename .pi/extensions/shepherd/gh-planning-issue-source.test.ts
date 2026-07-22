import assert from "node:assert/strict";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import {
	GhProductionPlanningIssueSource,
	type ProductionPlanningIssueExecutor,
} from "./gh-planning-issue-source.ts";

function responses(overrides: { parent?: Record<string, unknown>; permission?: string } = {}) {
	return {
		repository: {
			nameWithOwner: "acme/widgets",
			defaultBranchRef: { name: "main" },
			viewerPermission: overrides.permission ?? "MAINTAIN",
		},
		viewer: { login: "maintainer" },
		parent: {
			number: 479,
			node_id: "I_parent",
			title: "Build Shepherd",
			body: "Split this into safe implementation lanes.",
			state: "open",
			updated_at: "2026-07-22T10:00:00.000Z",
			...overrides.parent,
		},
		subissues: [{
			number: 501,
			title: "First lane",
			body: "Implement the first lane.",
			state: "open",
			updated_at: "2026-07-22T10:01:00.000Z",
		}],
	};
}

function executor(data = responses()): { execute: ProductionPlanningIssueExecutor; calls: string[][] } {
	const calls: string[][] = [];
	return {
		calls,
		async execute(args, options) {
			assert.ok(options.cwd.length > 0);
			assert.equal(options.signal.aborted, false);
			calls.push([...args]);
			const coordinate = args.join(" ");
			if (coordinate.startsWith("repo view")) return JSON.stringify(data.repository);
			if (coordinate.includes("/user")) return JSON.stringify(data.viewer);
			if (coordinate.includes("/issues/479/sub_issues")) {
				return JSON.stringify(coordinate.includes("page=1") ? data.subissues : []);
			}
			if (coordinate.includes("/issues/479")) return JSON.stringify(data.parent);
			throw new Error("unexpected typed GitHub call");
		},
	};
}

test("GitHub planning source reads exact open issue, maintainer, and complete bounded subissues", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-gh-planning-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const fake = executor();
	const source = new GhProductionPlanningIssueSource({
		execute: fake.execute,
		now: () => new Date("2026-07-22T10:02:00.000Z"),
	});
	const facts = await source.observe({ repositoryRoot: root, parentIssue: 479 }, {
		signal: new AbortController().signal,
		deadlineAt: "2099-07-22T11:00:00.000Z",
	});
	assert.equal(facts.repository, "acme/widgets");
	assert.equal(facts.defaultBranch, "main");
	assert.deepEqual(facts.viewer, { login: "maintainer", permission: "maintain" });
	assert.equal(facts.parent.number, 479);
	assert.equal(facts.subissues[0]?.number, 501);
	assert.equal(facts.complete, true);
	assert.match(facts.revisionDigest, /^[0-9a-f]{64}$/);
	assert.equal(fake.calls.filter((args) => args.join(" ").includes("sub_issues")).length, 1);
	assert.ok(fake.calls.some((args) => args.includes("X-GitHub-Api-Version: 2022-11-28")));
});

test("GitHub planning source rejects a closed or pull-request parent and insufficient viewer authority", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-gh-planning-reject-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	for (const [data, expected] of [
		[responses({ parent: { state: "closed" } }), /must be open/],
		[responses({ parent: { pull_request: { url: "https://example.invalid" } } }), /pull request/],
		[responses({ permission: "WRITE" }), /admin or maintainer/],
	] as const) {
		const fake = executor(data);
		const source = new GhProductionPlanningIssueSource({ execute: fake.execute });
		await assert.rejects(source.observe({ repositoryRoot: root, parentIssue: 479 }, {
			signal: new AbortController().signal,
			deadlineAt: "2099-07-22T11:00:00.000Z",
		}), expected);
	}
});

test("GitHub planning source honors cancellation before any command", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-gh-planning-cancel-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const fake = executor();
	const source = new GhProductionPlanningIssueSource({ execute: fake.execute });
	const controller = new AbortController();
	controller.abort(new Error("stop planning"));
	await assert.rejects(source.observe({ repositoryRoot: root, parentIssue: 479 }, {
		signal: controller.signal,
		deadlineAt: "2099-07-22T11:00:00.000Z",
	}), /stop planning/);
	assert.equal(fake.calls.length, 0);
});
