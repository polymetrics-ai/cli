import assert from "node:assert/strict";
import { execFile } from "node:child_process";
import { mkdir, mkdtemp, rename, rm, symlink } from "node:fs/promises";
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
		register(factory, resolveWorktree = async (ctx) => ({
			cwd: ctx.cwd,
			repositoryIdentity: "a".repeat(64),
			worktreeIdentity: "b".repeat(64),
		}), autonomousFactory) {
			registerShepherdExtension(hosts, {
				resolveWorktree,
				createController(ctx, worktree) {
					const controller = factory(ctx, worktree);
					controllers.push(controller);
					return controller;
				},
				...(autonomousFactory ? {
					createAutonomousController(ctx, worktree) {
						const controller = autonomousFactory(ctx, worktree);
						controllers.push(controller);
						return controller;
					},
				} : {}),
			});
		},
	};
}

test("routes autonomous control separately from the retained read-only canary and exposes no merge command", async () => {
	const h = harness();
	const calls = [];
	const autonomousState = (issue, status = "waiting_human") => ({
		schemaVersion: 2,
		kind: "autonomous",
		issue,
		planId: "plan",
		runId: "run-auto",
		generation: 1,
		status,
		stage: status === "waiting_human" ? "HUMAN_DECISION" : "SCHEDULE",
		maxConcurrency: 2,
		timeoutMs: 900_000,
		createdAt: "2026-07-22T08:00:00Z",
		updatedAt: "2026-07-22T08:01:00Z",
		children: [],
	});
	h.register(
		() => ({
			async status() { calls.push("canary:status"); return undefined; },
			async start(command) { calls.push("canary:start"); return state(command.issue); },
			async resume(command) { return state(command.issue); },
			async stop(issue) { return state(issue, "stopped"); },
			async shutdown() {},
		}),
		undefined,
		() => ({
			async status(issue) { calls.push("auto:status"); return autonomousState(issue); },
			async start(command) { calls.push("auto:start"); return autonomousState(command.issue); },
			async resume(command) { calls.push("auto:resume"); return autonomousState(command.issue); },
			async stop(issue) { calls.push("auto:stop"); return autonomousState(issue, "stopped"); },
			async shutdown() {},
		}),
	);
	await h.command.handler("start --issue 479", h.context);
	await new Promise((resolve) => setImmediate(resolve));
	await h.command.handler("status --issue 479", h.context);
	await h.command.handler("canary --issue 471 --pr 472 --read-only --backend sdk-inproc --experimental", h.context);
	await new Promise((resolve) => setImmediate(resolve));
	await h.command.handler("merge-main --issue 479", h.context);
	assert.deepEqual(calls, ["auto:start", "auto:status", "canary:start"]);
	assert.match(h.notifications.at(-1).message, /unknown action/i);
	assert.doesNotMatch(h.notifications.find((entry) => /Commands:/.test(entry.message))?.message ?? "", /merge-main/);
});

test("renders durable production scheduler, budget, and exact parent-gate status", async () => {
	const h = harness();
	const productionState = {
		schemaVersion: 1,
		kind: "production_autonomous",
		parentIssue: 479,
		repository: "owner/repo",
		planId: "production-plan",
		planDigest: "d".repeat(64),
		parentBranch: "feat/parent",
		parentBaseBranch: "main",
		runId: "run-production",
		generation: 2,
		revision: 9,
		maxConcurrency: 2,
		timeoutMs: 900_000,
		status: "waiting_human",
		stage: "human_decision",
		createdAt: "2026-07-22T08:00:00.000Z",
		updatedAt: "2026-07-22T08:01:00.000Z",
		idleReason: "capacity_exhausted",
		humanGate: {
			repository: "owner/repo",
			pullRequest: 438,
			generation: 2,
			head: "a".repeat(40),
			requestId: "parent-gate-479-2",
			status: "pending",
		},
		children: [{
			id: "lane-a",
			issue: 501,
			slug: "lane-a",
			specDigest: "e".repeat(64),
			dependsOn: [],
			writeScopes: ["owned/a"],
			maxAttempts: 2,
			maxCorrections: 1,
			attempts: 1,
			authorizedAttempts: 0,
			corrections: 0,
			status: "succeeded",
			stage: "succeeded",
		}],
	};
	h.register(
		() => ({
			async status() { return undefined; },
			async start(command) { return state(command.issue); },
			async stop(issue) { return state(issue, "stopped"); },
			async shutdown() {},
		}),
		undefined,
		() => ({
			async status() { return productionState; },
			async start() { return productionState; },
			async resume() { return productionState; },
			async stop() { return productionState; },
			async shutdown() {},
		}),
	);
	await h.command.handler("status --issue 479", h.context);
	const rendered = h.notifications.at(-1).message;
	assert.match(rendered, /status=waiting_human stage=human_decision/);
	assert.match(rendered, /lane-a: succeeded\/succeeded attempts=1\/2 corrections=0\/1/);
	assert.match(rendered, /idle_reason=capacity_exhausted/);
	assert.match(rendered, /human_gate=parent_merge:pending pr=#438 head=aaaaaaaaaaaa/);
});

test("status follows an active canary instead of stale autonomous state for the same issue", async () => {
	const h = harness();
	const calls = [];
	let release;
	const gate = new Promise((resolve) => { release = resolve; });
	h.register(
		() => ({
			async status(issue) { calls.push("canary:status"); return state(issue, "running"); },
			async start(command) { calls.push("canary:start"); await gate; return state(command.issue); },
			async resume(command) { return state(command.issue); },
			async stop(issue) { return state(issue, "stopped"); },
			async shutdown() {},
		}),
		undefined,
		() => ({
			async status(issue) {
				calls.push("auto:status");
				return {
					schemaVersion: 2,
					kind: "autonomous",
					issue,
					planId: "old-plan",
					runId: "old-run",
					generation: 1,
					status: "waiting_human",
					stage: "HUMAN_DECISION",
					maxConcurrency: 2,
					timeoutMs: 900_000,
					createdAt: "2026-07-22T08:00:00Z",
					updatedAt: "2026-07-22T08:01:00Z",
					children: [],
				};
			},
			async start() { throw new Error("not expected"); },
			async resume() { throw new Error("not expected"); },
			async stop() { throw new Error("not expected"); },
			async shutdown() {},
		}),
	);

	await h.command.handler(
		"canary --issue 479 --pr 438 --read-only --backend sdk-inproc --experimental",
		h.context,
	);
	await h.command.handler("status --issue 479", h.context);
	assert.deepEqual(calls, ["canary:start", "canary:status"]);
	assert.match(h.notifications.at(-1).message, /status=running/);
	release();
	await new Promise((resolve) => setImmediate(resolve));
});

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
	assert.equal(otherWorktreeIdentity.repositoryIdentity, rootIdentity.repositoryIdentity);
	assert.notEqual(otherWorktreeIdentity.worktreeIdentity, rootIdentity.worktreeIdentity);
	assert.notEqual(otherWorktreeIdentity.cwd, rootIdentity.cwd);
	assert.match(rootIdentity.repositoryIdentity, /^[0-9a-f]{64}$/);
	assert.match(rootIdentity.worktreeIdentity, /^[0-9a-f]{64}$/);
	assert.deepEqual(Object.keys(rootIdentity).sort(), ["cwd", "repositoryIdentity", "worktreeIdentity"]);
});

test("canonical identity survives moves but changes for replacement repositories and remotes", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-stable-identity-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const original = join(root, "original");
	const moved = join(root, "moved");
	await mkdir(original);
	await execFileAsync("git", ["init", "--quiet", original]);
	await execFileAsync("git", ["-C", original, "remote", "add", "origin", "https://user-a@example.invalid/org/repo.git"]);
	const initial = await canonicalizeGitWorktree(original);
	await execFileAsync("git", ["-C", original, "remote", "set-url", "origin", "https://user-b@example.invalid/org/repo.git"]);
	const credentialVariant = await canonicalizeGitWorktree(original);
	assert.equal(credentialVariant.repositoryIdentity, initial.repositoryIdentity);
	assert.equal(credentialVariant.worktreeIdentity, initial.worktreeIdentity);

	await rename(original, moved);
	const afterMove = await canonicalizeGitWorktree(moved);
	assert.equal(afterMove.repositoryIdentity, initial.repositoryIdentity);
	assert.equal(afterMove.worktreeIdentity, initial.worktreeIdentity);

	await execFileAsync("git", ["-C", moved, "remote", "set-url", "origin", "https://example.invalid/org/other.git"]);
	const changedRemote = await canonicalizeGitWorktree(moved);
	assert.notEqual(changedRemote.repositoryIdentity, initial.repositoryIdentity);
	assert.notEqual(changedRemote.worktreeIdentity, initial.worktreeIdentity);

	await rm(moved, { recursive: true, force: true });
	await mkdir(moved);
	await execFileAsync("git", ["init", "--quiet", moved]);
	await execFileAsync("git", ["-C", moved, "remote", "add", "origin", "https://user-c@example.invalid/org/repo.git"]);
	const replacement = await canonicalizeGitWorktree(moved);
	assert.notEqual(replacement.repositoryIdentity, initial.repositoryIdentity);
	assert.notEqual(replacement.worktreeIdentity, initial.worktreeIdentity);
});

test("canonical Git lookup rejects an already-aborted setup signal", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-aborted-identity-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const repo = join(root, "repo");
	await mkdir(repo);
	await execFileAsync("git", ["init", "--quiet", repo]);
	const abortController = new AbortController();
	abortController.abort(new Error("cancel canonical lookup"));
	await assert.rejects(
		canonicalizeGitWorktree(repo, { signal: abortController.signal }),
		/lookup was aborted/,
	);
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
			? { cwd: "/real/worktree-a", repositoryIdentity: "a".repeat(64), worktreeIdentity: "b".repeat(64) }
			: { cwd: "/real/worktree-b", repositoryIdentity: "a".repeat(64), worktreeIdentity: "c".repeat(64) },
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

test("shutdown aborts and joins pre-controller worktree resolution", async () => {
	const h = harness();
	let resolutionStarted;
	const started = new Promise((resolve) => { resolutionStarted = resolve; });
	let observedAbort = false;
	h.register(
		() => { throw new Error("controller must not be created after setup cancellation"); },
		async (_ctx, options) => {
			resolutionStarted();
			await new Promise((resolve, reject) => {
				options.signal.addEventListener("abort", () => {
					observedAbort = true;
					reject(options.signal.reason);
				}, { once: true });
			});
			throw new Error("unreachable");
		},
	);
	const handling = h.command.handler(
		"canary --issue 397 --pr 438 --read-only --backend sdk-inproc --experimental",
		h.context,
	);
	await started;
	await h.shutdown();
	await handling;
	assert.equal(observedAbort, true);
	assert.equal(h.controllers.length, 0);
});

test("explicit stop aborts and joins matching unresolved controller initialization", { timeout: 2_000 }, async () => {
	const h = harness();
	let resolutionStarted;
	const started = new Promise((resolve) => { resolutionStarted = resolve; });
	let observedAbort = false;
	let creates = 0;
	h.register(
		() => {
			creates += 1;
			throw new Error("controller must not be created after initialization stop");
		},
		async (_ctx, options) => {
			resolutionStarted();
			await new Promise((resolve, reject) => {
				options.signal.addEventListener("abort", () => {
					observedAbort = true;
					reject(options.signal.reason);
				}, { once: true });
			});
			throw new Error("unreachable");
		},
	);
	const starting = h.command.handler("start --issue 479", h.context);
	await started;
	await h.command.handler("stop --issue 479", h.context);
	await starting;
	assert.equal(observedAbort, true);
	assert.equal(creates, 0);
	assert.match(h.notifications.at(-1).message, /initialization.*stopped/i);
});

test("shutdown closes a controller created late by an abort-ignoring resolver", async () => {
	const h = harness();
	let releaseResolution;
	const gate = new Promise((resolve) => { releaseResolution = resolve; });
	let resolutionStarted;
	const started = new Promise((resolve) => { resolutionStarted = resolve; });
	let starts = 0;
	let closes = 0;
	h.register(
		() => ({
			async status() { return undefined; },
			async start(command) { starts += 1; return state(command.issue); },
			async resume(command) { starts += 1; return state(command.issue); },
			async stop(issue) { return state(issue, "stopped"); },
			async shutdown() { closes += 1; },
		}),
		async (ctx) => {
			resolutionStarted();
			await gate;
			return { cwd: ctx.cwd, repositoryIdentity: "a".repeat(64), worktreeIdentity: "b".repeat(64) };
		},
	);
	const handling = h.command.handler(
		"canary --issue 397 --pr 438 --read-only --backend sdk-inproc --experimental",
		h.context,
	);
	await started;
	const shuttingDown = h.shutdown();
	releaseResolution();
	await Promise.all([handling, shuttingDown]);
	assert.equal(starts, 0);
	assert.equal(closes, 1);
});

test("shutdown fails its deadline when pre-controller resolution ignores abort forever", async (t) => {
	t.mock.timers.enable({ apis: ["setTimeout"] });
	const h = harness();
	let resolutionStarted;
	const started = new Promise((resolve) => { resolutionStarted = resolve; });
	h.register(
		() => { throw new Error("controller must not be created"); },
		async () => {
			resolutionStarted();
			await new Promise(() => {});
			throw new Error("unreachable");
		},
	);
	void h.command.handler(
		"canary --issue 397 --pr 438 --read-only --backend sdk-inproc --experimental",
		h.context,
	);
	await started;
	const shuttingDown = h.shutdown();
	t.mock.timers.tick(45_000);
	await assert.rejects(shuttingDown, /shutdown deadline exceeded/i);
});
