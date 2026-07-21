import assert from "node:assert/strict";
import test from "node:test";

import { registerShepherdExtension } from "./extension.ts";

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
		register(factory) {
			registerShepherdExtension(hosts, {
				createController(ctx) {
					const controller = factory(ctx);
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
