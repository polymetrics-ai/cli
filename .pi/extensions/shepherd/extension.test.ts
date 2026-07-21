import assert from "node:assert/strict";
import { execFile } from "node:child_process";
import { mkdir, mkdtemp, rm, symlink } from "node:fs/promises";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";
import { promisify } from "node:util";

import { canonicalizeGitWorktree, registerShepherdExtension } from "./extension.ts";

const execFileAsync = promisify(execFile);

function state(issue, status = "completed") {
	return {
		schemaVersion: 1,
		issue,
		runId: `run-${issue}`,
		generation: 1,
		status,
		candidateHead: "a".repeat(40),
		validationNonce: "nonce-1234567890",
		createdAt: "2026-07-21T08:00:00Z",
		updatedAt: "2026-07-21T08:01:00Z",
		lanes: [],
	};
}

function harness() {
	let command;
	let shutdown;
	const hosts = {
		registerCommand(name, definition) {
			assert.equal(name, "pm-shepherd");
			command = definition;
		},
		on(event, handler) {
			assert.equal(event, "session_shutdown");
			shutdown = handler;
		},
	};
	const controllers = [];
	const notifications = [];
	const statuses = [];
	const context = {
		cwd: "/tmp/pr-438",
		modelRegistry: {},
		isIdle: () => true,
		ui: {
			notify: (message, level) => notifications.push({ message, level }),
			setStatus: (key, value) => statuses.push({ key, value }),
		},
	};
	return {
		hosts,
		controllers,
		notifications,
		statuses,
		context,
		get command() { return command; },
		get shutdown() { return shutdown; },
		register(factory, resolveWorktree = async (ctx) => ({ cwd: ctx.cwd, identity: ctx.cwd })) {
			registerShepherdExtension(hosts, {
				resolveWorktree,
				createController(ctx, worktree) {
					const controller = factory(ctx, worktree);
					controllers.push(controller);
					return controller;
				},
			});
		},
	};
}

test("help and status never dispatch an AgentSession run", async () => {
	const h = harness();
	let starts = 0;
	h.register(() => ({
		async status() { return undefined; },
		async start() { starts += 1; return state(471); },
		async resume() { starts += 1; return state(471); },
		async stop() { return state(471, "stopped"); },
		async shutdown() {},
	}));
	await h.command.handler("help", h.context);
	await h.command.handler("status --issue 471", h.context);
	assert.equal(starts, 0);
	assert.match(h.notifications[0].message, /AgentSession Shepherd/);
	assert.match(h.notifications[1].message, /No persisted/);
});

test("allows only one active embedded run across all issues", async () => {
	const h = harness();
	let release;
	const gate = new Promise((resolve) => { release = resolve; });
	let starts = 0;
	h.register(() => ({
		async status() { return undefined; },
		async start(command) { starts += 1; await gate; return state(command.issue); },
		async resume(command) { starts += 1; await gate; return state(command.issue); },
		async stop(issue) { return state(issue, "stopped"); },
		async shutdown() {},
	}));
	const flags = "--pr 438 --read-only --backend sdk-inproc --experimental";
	await h.command.handler(`canary --issue 397 ${flags}`, h.context);
	await h.command.handler(`canary --issue 471 ${flags}`, h.context);
	assert.equal(starts, 1);
	assert.match(h.notifications.at(-1).message, /already active.*#397/i);
	release();
	await new Promise((resolve) => setTimeout(resolve, 0));
});

test("reserves the process-wide launch slot before asynchronous worktree resolution", async () => {
	const h = harness();
	let releaseResolution;
	const resolutionGate = new Promise((resolve) => { releaseResolution = resolve; });
	let starts = 0;
	h.register(
		() => ({
			async status() { return undefined; },
			async start(command) { starts += 1; return state(command.issue); },
			async resume(command) { starts += 1; return state(command.issue); },
			async stop(issue) { return state(issue, "stopped"); },
			async shutdown() {},
		}),
		async (ctx) => {
			await resolutionGate;
			return { cwd: ctx.cwd, identity: ctx.cwd };
		},
	);
	const flags = "--pr 438 --read-only --backend sdk-inproc --experimental";
	const first = h.command.handler(`canary --issue 397 ${flags}`, h.context);
	const second = h.command.handler(`canary --issue 471 ${flags}`, h.context);
	releaseResolution();
	await Promise.all([first, second]);
	assert.equal(starts, 1);
	assert.match(h.notifications.find((entry) => /already active/i.test(entry.message))?.message ?? "", /#397/);
});

test("shutdown closes controllers, waits for the active run, and suppresses late UI output", async () => {
	const h = harness();
	let release;
	const gate = new Promise((resolve) => { release = resolve; });
	let closed = 0;
	h.register(() => ({
		async status() { return undefined; },
		async start(command) { await gate; return state(command.issue); },
		async resume(command) { await gate; return state(command.issue); },
		async stop(issue) { return state(issue, "stopped"); },
		async shutdown() { closed += 1; release(); },
	}));
	await h.command.handler(
		"canary --issue 397 --pr 438 --read-only --backend sdk-inproc --experimental",
		h.context,
	);
	const beforeShutdown = h.notifications.length;
	await h.shutdown();
	assert.equal(closed, 1);
	assert.equal(h.notifications.length, beforeShutdown);
});

test("canonical Git identity converges root, subdirectory, and symlink while separating worktrees", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-identity-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const primary = join(root, "primary");
	const secondary = join(root, "secondary");
	const alias = join(root, "alias");
	await mkdir(primary);
	await execFileAsync("git", ["init", "--quiet", primary]);
	await execFileAsync("git", ["-C", primary, "-c", "user.name=Shepherd Test", "-c", "user.email=shepherd@example.invalid", "commit", "--allow-empty", "--quiet", "-m", "seed"]);
	await mkdir(join(primary, "nested"));
	await symlink(primary, alias, "dir");
	await execFileAsync("git", ["-C", primary, "worktree", "add", "--quiet", "--detach", secondary, "HEAD"]);

	const rootIdentity = await canonicalizeGitWorktree(primary);
	const subdirIdentity = await canonicalizeGitWorktree(join(primary, "nested"));
	const symlinkIdentity = await canonicalizeGitWorktree(join(alias, "nested"));
	const otherWorktreeIdentity = await canonicalizeGitWorktree(secondary);
	assert.deepEqual(subdirIdentity, rootIdentity);
	assert.deepEqual(symlinkIdentity, rootIdentity);
	assert.notEqual(otherWorktreeIdentity.identity, rootIdentity.identity);
	assert.notEqual(otherWorktreeIdentity.cwd, rootIdentity.cwd);
});

test("controller cache is keyed by canonical worktree and issue", async () => {
	const h = harness();
	const canonicalContexts = [];
	let statusCalls = 0;
	const controller = () => ({
		async status() { statusCalls += 1; return undefined; },
		async start(command) { return state(command.issue); },
		async resume(command) { return state(command.issue); },
		async stop(issue) { return state(issue, "stopped"); },
		async shutdown() {},
	});
	h.register(
		(ctx, worktree) => {
			canonicalContexts.push({ cwd: ctx.cwd, worktree });
			return controller();
		},
		async (ctx) => ctx.cwd.startsWith("/alias-a")
			? { cwd: "/real/worktree-a", identity: "worktree-a" }
			: { cwd: "/real/worktree-b", identity: "worktree-b" },
	);
	await h.command.handler("status --issue 471", { ...h.context, cwd: "/alias-a/nested" });
	await h.command.handler("status --issue 471", { ...h.context, cwd: "/alias-a/symlink" });
	await h.command.handler("status --issue 471", { ...h.context, cwd: "/other-worktree" });
	assert.equal(statusCalls, 3);
	assert.equal(h.controllers.length, 2);
	assert.deepEqual(canonicalContexts.map((entry) => entry.cwd), ["/real/worktree-a", "/real/worktree-b"]);
});

test("shutdown propagates aggregated failures and retains failed controller ownership", async () => {
	const h = harness();
	let creates = 0;
	let statusCalls = 0;
	h.register(() => {
		creates += 1;
		return {
			async status() { statusCalls += 1; return undefined; },
			async start(command) { return state(command.issue); },
			async resume(command) { return state(command.issue); },
			async stop(issue) { return state(issue, "stopped"); },
			async shutdown() { throw new Error("cleanup failed"); },
		};
	});
	await h.command.handler("status --issue 471", h.context);
	await assert.rejects(
		h.shutdown(),
		(error) => error instanceof AggregateError && error.errors.some((entry) => /cleanup failed/.test(String(entry))),
	);
	await h.command.handler("status --issue 471", h.context);
	assert.equal(creates, 1);
	assert.equal(statusCalls, 2);
});

test("shutdown propagates its deadline and retains a controller that has not exited", async (t) => {
	t.mock.timers.enable({ apis: ["setTimeout"] });
	const h = harness();
	let creates = 0;
	h.register(() => {
		creates += 1;
		return {
			async status() { return undefined; },
			async start(command) { return state(command.issue); },
			async resume(command) { return state(command.issue); },
			async stop(issue) { return state(issue, "stopped"); },
			async shutdown() { await new Promise(() => {}); },
		};
	});
	await h.command.handler("status --issue 471", h.context);
	const shuttingDown = h.shutdown();
	t.mock.timers.tick(45_000);
	await assert.rejects(shuttingDown, /shutdown deadline exceeded/i);
	await h.command.handler("status --issue 471", h.context);
	assert.equal(creates, 1);
});
