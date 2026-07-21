import assert from "node:assert/strict";
import { spawnSync } from "node:child_process";
import test from "node:test";

import {
	DependencyGraphError,
	scopesCollide,
	selectReadyWork,
	validateDependencyGraph,
	type DependencyWorkItem,
} from "./dependency-graph.ts";

function item(overrides: Partial<DependencyWorkItem> & Pick<DependencyWorkItem, "id">): DependencyWorkItem {
	return {
		id: overrides.id,
		dependsOn: [],
		status: "pending",
		access: "mutating",
		writeScopes: [`src/${overrides.id}`],
		...overrides,
	};
}

test("validates a closed dependency DAG and returns deterministic topological order", () => {
	const result = validateDependencyGraph([
		item({ id: "review", access: "read_only", writeScopes: [], dependsOn: ["worker"] }),
		item({ id: "worker", dependsOn: ["plan"] }),
		item({ id: "plan", access: "read_only", writeScopes: [] }),
	]);
	assert.deepEqual(result.topologicalOrder, ["plan", "worker", "review"]);
	assert.deepEqual(result.items.map((candidate) => candidate.id), ["plan", "review", "worker"]);
});

test("cycles, self edges, unknown dependencies, and duplicate IDs fail closed", () => {
	const cases: Array<[string, DependencyWorkItem[], string]> = [
		["cycle", [item({ id: "a", dependsOn: ["b"] }), item({ id: "b", dependsOn: ["a"] })], "cycle"],
		["self", [item({ id: "a", dependsOn: ["a"] })], "cycle"],
		["unknown", [item({ id: "a", dependsOn: ["missing"] })], "unknown_dependency"],
		["duplicate", [item({ id: "a" }), item({ id: "a" })], "duplicate_id"],
	];
	for (const [name, items, code] of cases) {
		assert.throws(
			() => validateDependencyGraph(items),
			(error: unknown) => error instanceof DependencyGraphError && error.code === code,
			name,
		);
	}
});

test("ambiguous, broad, traversing, globbed, duplicate, and redundant write scopes fail closed", () => {
	const scopes = [
		[], ["."], ["/absolute"], ["src/../secret"], ["src\\windows"], ["src/*"],
		["src/a", "src/a"], ["src", "src/a"], [" src/a"],
	];
	for (const writeScopes of scopes) {
		assert.throws(
			() => validateDependencyGraph([item({ id: "worker", writeScopes })]),
			(error: unknown) => error instanceof DependencyGraphError && error.code === "ambiguous_scope",
			JSON.stringify(writeScopes),
		);
	}
	assert.throws(
		() => validateDependencyGraph([item({ id: "reader", access: "read_only", writeScopes: ["src"] })]),
		(error: unknown) => error instanceof DependencyGraphError && error.code === "ambiguous_scope",
	);
});

test("scope collision is segment-aware and detects ancestor overlap", () => {
	assert.equal(scopesCollide(["src/policy"], ["src/policy/file.ts"]), true);
	assert.equal(scopesCollide(["src/policy"], ["src/policy-two"]), false);
	assert.equal(scopesCollide(["a/file.ts"], ["b/file.ts"]), false);
});

test("scope collision conservatively folds Darwin/Git case and Unicode aliases", () => {
	assert.equal(scopesCollide(["src/Readme.md"], ["src/README.md"]), true);
	assert.equal(scopesCollide(["src/caf\u00e9/file.ts"], ["src/cafe\u0301/file.ts"]), true);
	assert.equal(scopesCollide(["src/Stra\u00dfe/file.ts"], ["src/STRASSE/file.ts"]), true);
	assert.equal(scopesCollide(["src/\u1e9e/file.ts"], ["src/\u00df/file.ts"]), true);
	assert.equal(scopesCollide(["src/\u1e9e/file.ts"], ["src/ss/file.ts"]), true);
	assert.throws(
		() => validateDependencyGraph([item({ id: "aliases", writeScopes: ["src/Readme.md", "src/README.md"] })]),
		(error: unknown) => error instanceof DependencyGraphError && error.code === "ambiguous_scope",
	);
	assert.throws(
		() => validateDependencyGraph([item({ id: "sharp-s", writeScopes: ["src/\u1e9e", "src/ss"] })]),
		(error: unknown) => error instanceof DependencyGraphError && error.code === "ambiguous_scope",
	);
});

test("canonical ordering is explicit ECMAScript code-unit order, independent of locale collation", () => {
	const graph = validateDependencyGraph([
		item({ id: "\u00e4", access: "read_only", writeScopes: [] }),
		item({ id: "z", access: "read_only", writeScopes: [] }),
	]);
	assert.deepEqual(graph.items.map((candidate) => candidate.id), ["z", "\u00e4"]);
	assert.deepEqual(graph.topologicalOrder, ["z", "\u00e4"]);
});

test("dependency/status-incoherent and non-exact work-item snapshots fail closed", () => {
	for (const status of ["running", "succeeded", "failed"] as const) {
		assert.throws(
			() => validateDependencyGraph([
				item({ id: "dependency", status: "pending" }),
				item({ id: "dependent", status, dependsOn: ["dependency"] }),
			]),
			(error: unknown) => error instanceof DependencyGraphError && error.code === "incoherent_status",
		);
	}
	assert.throws(
		() => validateDependencyGraph([{
			...item({ id: "extra" }),
			unexpected: true,
		} as DependencyWorkItem]),
		(error: unknown) => error instanceof DependencyGraphError && error.code === "invalid_item",
	);
});

test("selects a maximum safe set instead of a greedy first-fit set", () => {
	const items = [
		item({ id: "central", writeScopes: ["src"] }),
		item({ id: "leaf-a", writeScopes: ["src/a"] }),
		item({ id: "leaf-b", writeScopes: ["src/b"] }),
	];
	const result = selectReadyWork(items, { maxConcurrency: 2 });
	assert.deepEqual(result, { kind: "selected", itemIds: ["leaf-a", "leaf-b"] });
});

test("selects every dependency-ready non-colliding lane up to remaining concurrency", () => {
	const items = [
		item({ id: "done", status: "succeeded" }),
		item({ id: "worker-a", dependsOn: ["done"], writeScopes: ["src/a"] }),
		item({ id: "worker-b", dependsOn: ["done"], writeScopes: ["src/b"] }),
		item({ id: "blocked", dependsOn: ["later"], writeScopes: ["src/c"] }),
		item({ id: "later", status: "pending", access: "read_only", writeScopes: [] }),
	];
	assert.deepEqual(selectReadyWork(items, { maxConcurrency: 3 }), {
		kind: "selected",
		itemIds: ["later", "worker-a", "worker-b"],
	});
});

test("read-only research and review coexist with mutators while mutating collisions serialize", () => {
	const items = [
		item({ id: "running-writer", status: "running", writeScopes: ["src/shared"] }),
		item({ id: "colliding-writer", writeScopes: ["src/shared/file.ts"] }),
		item({ id: "research", access: "read_only", writeScopes: [] }),
		item({ id: "review", access: "read_only", writeScopes: [] }),
	];
	assert.deepEqual(selectReadyWork(items, { maxConcurrency: 4 }), {
		kind: "selected",
		itemIds: ["research", "review"],
	});
});

test("selection reports capacity, completion, or exactly one repository blocker", () => {
	assert.deepEqual(selectReadyWork([], { maxConcurrency: 1 }), { kind: "complete" });
	assert.deepEqual(selectReadyWork([item({ id: "running", status: "running" })], { maxConcurrency: 1 }), {
		kind: "at_capacity",
	});
	assert.deepEqual(selectReadyWork([item({ id: "done", status: "succeeded" })], { maxConcurrency: 1 }), {
		kind: "complete",
	});
	const dependencyBlocked = [
		item({ id: "first", status: "failed" }),
		item({ id: "second", dependsOn: ["first"] }),
	];
	assert.deepEqual(selectReadyWork(dependencyBlocked, { maxConcurrency: 2 }), {
		kind: "blocked",
		blocker: "not_spawned_dependency_blocked",
	});
	const collisionBlocked = [
		item({ id: "running", status: "running", writeScopes: ["src/shared"] }),
		item({ id: "waiting", writeScopes: ["src/shared/file.ts"] }),
	];
	assert.deepEqual(selectReadyWork(collisionBlocked, { maxConcurrency: 2 }), {
		kind: "blocked",
		blocker: "not_spawned_write_scope_collision",
	});
});

test("selection is deterministic and does not mutate caller-owned items", () => {
	const items = [item({ id: "b", writeScopes: ["b"] }), item({ id: "a", writeScopes: ["a"] })];
	const before = structuredClone(items);
	const first = selectReadyWork(items, { maxConcurrency: 2 });
	const second = selectReadyWork(items, { maxConcurrency: 2 });
	assert.deepEqual(first, second);
	assert.deepEqual(first, { kind: "selected", itemIds: ["a", "b"] });
	assert.deepEqual(items, before);
});

test("a hostile 64-item conflict component is rejected within a bounded subprocess deadline", () => {
	const script = `
		import { selectReadyWork } from "./.pi/extensions/shepherd/dependency-graph.ts";
		const count = 64;
		const items = Array.from({ length: count }, (_, index) => ({
			id: String(index).padStart(2, "0"),
			dependsOn: [],
			status: "pending",
			access: "mutating",
			writeScopes: ["edges/" + index, "edges/" + ((index + count - 1) % count)],
		}));
		try {
			selectReadyWork(items, { maxConcurrency: 64 });
			process.stdout.write("not_rejected");
		} catch (error) {
			process.stdout.write(error && typeof error === "object" && "code" in error ? String(error.code) : "untyped");
		}
	`;
	const started = performance.now();
	const result = spawnSync(process.execPath, ["--input-type=module", "--eval", script], {
		cwd: process.cwd(),
		encoding: "utf8",
		timeout: 1_000,
	});
	assert.equal(result.signal, null, `scheduler exceeded deadline: ${result.signal ?? result.error?.message}`);
	assert.equal(result.status, 0, result.stderr);
	assert.equal(result.stdout, "conflict_component_too_large");
	assert.ok(performance.now() - started < 1_000, "scheduler must remain below the subprocess deadline");
});
