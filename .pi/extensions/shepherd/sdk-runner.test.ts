import assert from "node:assert/strict";
import test from "node:test";

import { SdkAgentRunner } from "./sdk-runner.ts";

const head = "a".repeat(40);
const dimensions = {
	correctStage: 1,
	artifactValid: 1,
	gatesRespected: 1,
	realProgress: 1,
	noHallucination: 1,
	noConflict: 1,
};

function request(overrides = {}) {
	const base = {
		runId: "run-1",
		laneId: "scout",
		role: "scout",
		cwd: "/tmp/pr-438",
		readOnly: true,
		provider: "openai-codex",
		model: "gpt-5.6-sol",
		thinking: "xhigh",
		tools: [],
		systemPrompt: "Read-only scout.",
		prompt: "Inspect exact head.",
		timeoutMs: 1_000,
		binding: {
			runId: "run-1",
			generation: 1,
			laneId: "scout",
			candidateHead: head,
			validationNonce: "nonce-1234567890",
			readOnly: true,
			provider: "openai-codex",
			model: "gpt-5.6-sol",
			thinking: "xhigh",
		},
	};
	return { ...base, ...overrides };
}

function evidenceText(overrides = {}) {
	return JSON.stringify({
		...request().binding,
		summary: "read-only inspection passed",
		dimensions,
		observedMutation: false,
		...overrides,
	});
}

function terminalMessage(stopReason = "stop", text = evidenceText(), timestamp = 1) {
	return {
		role: "assistant",
		content: [{ type: "text", text }],
		api: "responses",
		provider: "openai-codex",
		model: "gpt-5.6-sol",
		usage: {
			input: 0,
			output: 0,
			cacheRead: 0,
			cacheWrite: 0,
			totalTokens: 0,
			cost: { input: 0, output: 0, cacheRead: 0, cacheWrite: 0, total: 0 },
		},
		stopReason,
		timestamp,
	};
}

function settlesWithin(promise, timeoutMs, label) {
	return Promise.race([
		promise,
		new Promise((_, reject) => setTimeout(() => reject(new Error(`${label} did not settle`)), timeoutMs)),
	]);
}

function makeSdk(overrides = {}) {
	const calls = { reload: 0, prompt: 0, abort: 0, idle: 0, unsubscribe: 0, dispose: 0 };
	let resourceOptions;
	let sessionOptions;
	let listener;
	const emit = (event) => listener?.(event);
	const emitTerminal = (stopReason = "stop", text = evidenceText()) => {
		const message = terminalMessage(stopReason, text);
		emit({ type: "message_end", message });
		emit({ type: "agent_end", messages: [message], willRetry: false });
		return message;
	};
	const session = {
		model: { provider: "openai-codex", id: "gpt-5.6-sol" },
		thinkingLevel: "xhigh",
		sessionFile: undefined,
		getActiveToolNames: () => [],
		subscribe(nextListener) {
			listener = nextListener;
			return () => { calls.unsubscribe += 1; };
		},
		async prompt(_prompt, options) {
			calls.prompt += 1;
			assert.deepEqual(options, { expandPromptTemplates: false, source: "extension" });
			emitTerminal();
		},
		async waitForIdle() { calls.idle += 1; },
		async abort() { calls.abort += 1; },
		dispose() { calls.dispose += 1; },
		getLastAssistantText: () => evidenceText(),
	};
	const modelRegistry = {
		authStorage: { opaque: true },
		find: (provider, model) => provider === "openai-codex" && model === "gpt-5.6-sol" ? session.model : undefined,
		hasConfiguredAuth: () => true,
	};
	const sdk = {
		version: "0.80.6",
		requiredVersion: "0.80.6",
		getAgentDir: () => "/tmp/pi-agent",
		createSettingsManager: (settings, options) => ({ settings, options, kind: "settings" }),
		createSessionManager: (cwd) => ({ cwd, kind: "session" }),
		createResourceLoader(options) {
			resourceOptions = options;
			return { async reload() { calls.reload += 1; } };
		},
		async createSession(options) {
			sessionOptions = options;
			return { session, extensionsResult: { extensions: [], errors: [] } };
		},
		...overrides,
	};
	return {
		sdk,
		modelRegistry,
		calls,
		session,
		emit,
		emitTerminal,
		get resourceOptions() { return resourceOptions; },
		get sessionOptions() { return sessionOptions; },
	};
}

test("constructs an isolated exact-model session and always cleans it up", async () => {
	const harness = makeSdk();
	const runner = new SdkAgentRunner(harness.sdk, harness.modelRegistry);
	const result = await runner.run(request());
	assert.equal(result.summary, "read-only inspection passed");
	assert.equal(harness.calls.reload, 1);
	assert.equal(harness.resourceOptions.noExtensions, true);
	assert.equal(harness.resourceOptions.noSkills, true);
	assert.equal(harness.resourceOptions.noPromptTemplates, true);
	assert.equal(harness.resourceOptions.noContextFiles, true);
	assert.equal(harness.sessionOptions.model.id, "gpt-5.6-sol");
	assert.equal(harness.sessionOptions.thinkingLevel, "xhigh");
	assert.equal(harness.sessionOptions.noTools, "all");
	assert.deepEqual(harness.sessionOptions.tools, []);
	assert.deepEqual(harness.sessionOptions.customTools, []);
	assert.equal(harness.calls.abort, 1);
	assert.equal(harness.calls.unsubscribe, 1);
	assert.equal(harness.calls.dispose, 1);
});

test("fails closed on SDK, model, extension, tool, and persistence drift", async () => {
	const badVersion = makeSdk({ version: "0.81.0" });
	await assert.rejects(new SdkAgentRunner(badVersion.sdk, badVersion.modelRegistry).run(request()), /requires Pi 0.80.6/);

	const nested = makeSdk({
		async createSession() {
			return { session: nested.session, extensionsResult: { extensions: ["recursive"], errors: [] } };
		},
	});
	await assert.rejects(new SdkAgentRunner(nested.sdk, nested.modelRegistry).run(request()), /unexpectedly loaded extensions/);

	const wrongTools = makeSdk();
	wrongTools.session.getActiveToolNames = () => ["read"];
	await assert.rejects(new SdkAgentRunner(wrongTools.sdk, wrongTools.modelRegistry).run(request()), /zero active tools/);

	const persistent = makeSdk();
	persistent.session.sessionFile = "/tmp/forbidden.jsonl";
	await assert.rejects(new SdkAgentRunner(persistent.sdk, persistent.modelRegistry).run(request()), /unexpectedly enabled persistence/);
});

test("rejects every requested built-in or custom child tool before session creation", async () => {
	const harness = makeSdk();
	await assert.rejects(
		new SdkAgentRunner(harness.sdk, harness.modelRegistry).run(request({ tools: ["read"] })),
		/child tools are disabled/,
	);
	await assert.rejects(
		new SdkAgentRunner(harness.sdk, harness.modelRegistry).run({
			...request(),
			customTools: [{ name: "unsafe-custom-tool" }],
		}),
		/custom child tools are disabled/,
	);
	assert.equal(harness.calls.reload, 0);
});

test("fails closed on malformed and oversized assistant evidence", async () => {
	const malformed = makeSdk();
	malformed.session.getLastAssistantText = () => evidenceText({ summary: "stale accessor must not be trusted" });
	malformed.session.prompt = async () => { malformed.emitTerminal("stop", "not-json"); };
	await assert.rejects(
		new SdkAgentRunner(malformed.sdk, malformed.modelRegistry).run(request()),
		/evidence must be one JSON object/,
	);
	assert.equal(malformed.calls.dispose, 1);

	const oversized = makeSdk();
	oversized.session.prompt = async () => { oversized.emitTerminal("stop", "x".repeat(65)); };
	await assert.rejects(
		new SdkAgentRunner(oversized.sdk, oversized.modelRegistry, { maxAssistantBytes: 64 }).run(request()),
		/output limit/,
	);
	assert.equal(oversized.calls.dispose, 1);
});


for (const stopReason of ["stop", "error", "aborted", "length", "toolUse"]) {
	test(`${stopReason} terminal evidence is ${stopReason === "stop" ? "accepted" : "rejected"}`, async () => {
		const harness = makeSdk();
		harness.session.getLastAssistantText = () => evidenceText({ summary: "unverified accessor text" });
		harness.session.prompt = async () => { harness.emitTerminal(stopReason, evidenceText({ summary: `${stopReason} terminal` })); };
		const outcome = new SdkAgentRunner(harness.sdk, harness.modelRegistry).run(request());
		if (stopReason === "stop") {
			assert.equal((await outcome).summary, "stop terminal");
		} else {
			await assert.rejects(outcome, new RegExp(`terminal.*${stopReason}`, "i"));
		}
		assert.equal(harness.calls.dispose, 1);
	});
}

test("rejects incomplete or inconsistent terminal event sequences", async () => {
	const noAgentEnd = makeSdk();
	noAgentEnd.session.prompt = async () => {
		noAgentEnd.emit({ type: "message_end", message: terminalMessage() });
	};
	await assert.rejects(
		new SdkAgentRunner(noAgentEnd.sdk, noAgentEnd.modelRegistry).run(request()),
		/terminal event sequence/i,
	);
});

test("aborts and cleans up when the event budget is exceeded", async () => {
	const harness = makeSdk();
	harness.session.prompt = async () => {
		harness.emit({ type: "first" });
		harness.emit({ type: "second" });
		harness.emitTerminal();
	};
	await assert.rejects(
		new SdkAgentRunner(harness.sdk, harness.modelRegistry, { maxEvents: 1 }).run(request()),
		/event limit exceeded/,
	);
	assert.equal(harness.calls.abort, 1);
	assert.equal(harness.calls.unsubscribe, 1);
	assert.equal(harness.calls.dispose, 1);
});

test("times out by aborting once and still disposes the child session", async () => {
	const harness = makeSdk();
	harness.session.prompt = () => new Promise(() => undefined);
	await assert.rejects(
		new SdkAgentRunner(harness.sdk, harness.modelRegistry).run(request({ timeoutMs: 10 })),
		/timed out after 10ms/,
	);
	assert.equal(harness.calls.abort, 1);
	assert.equal(harness.calls.idle, 1);
	assert.equal(harness.calls.unsubscribe, 1);
	assert.equal(harness.calls.dispose, 1);
});

test("abort addresses only children owned by the exact run", async () => {
	let release;
	const harness = makeSdk();
	harness.session.prompt = () => new Promise((resolve) => {
		release = () => { harness.emitTerminal(); resolve(); };
	});
	const runner = new SdkAgentRunner(harness.sdk, harness.modelRegistry);
	const pending = runner.run(request());
	await new Promise((resolve) => setTimeout(resolve, 0));
	await runner.abort("other-run");
	assert.equal(harness.calls.abort, 0);
	await runner.abort("run-1");
	assert.equal(harness.calls.abort, 1);
	release();
	await assert.rejects(pending, /cancel/i);
});

test("abort racing prompt resolution rejects otherwise valid late evidence", async () => {
	const harness = makeSdk();
	let runner;
	harness.session.prompt = async () => {
		harness.emitTerminal("stop", evidenceText({ summary: "late but schema-valid" }));
		await runner.abort("run-1");
	};
	runner = new SdkAgentRunner(harness.sdk, harness.modelRegistry);
	await assert.rejects(runner.run(request()), /cancel/i);
	assert.equal(harness.calls.dispose, 1);
});

test("abort during session creation prevents every later prompt and disposes the child", async () => {
	let releaseCreation;
	const creationGate = new Promise((resolve) => { releaseCreation = resolve; });
	let creationStarted;
	const creationStartedGate = new Promise((resolve) => { creationStarted = resolve; });
	const harness = makeSdk();
	const originalCreate = harness.sdk.createSession;
	harness.sdk.createSession = async (options) => {
		creationStarted();
		await creationGate;
		return originalCreate(options);
	};
	const runner = new SdkAgentRunner(harness.sdk, harness.modelRegistry);
	const pending = runner.run(request());
	await creationStartedGate;
	await runner.abort("run-1");
	releaseCreation();
	await assert.rejects(pending, /abort|cancel/i);
	await runner.close();
	assert.equal(harness.calls.prompt, 0);
	assert.equal(harness.calls.dispose, 1);
});

test("hung abort and idle hooks are bounded and disposal still runs", async () => {
	const harness = makeSdk();
	harness.session.abort = () => new Promise(() => undefined);
	harness.session.waitForIdle = () => new Promise(() => undefined);
	const runner = new SdkAgentRunner(harness.sdk, harness.modelRegistry, { cleanupTimeoutMs: 10 });
	const bounded = Promise.race([
		runner.run(request()),
		new Promise((_, reject) => setTimeout(() => reject(new Error("runner cleanup did not settle")), 200)),
	]);
	await assert.rejects(bounded, /cleanup.*timed out/i);
	assert.equal(harness.calls.dispose, 1);
});

test("close settles within its cleanup deadline and disposes a hung active child", async () => {
	let releasePrompt;
	const promptGate = new Promise((resolve) => { releasePrompt = resolve; });
	let promptStarted;
	const promptStartedGate = new Promise((resolve) => { promptStarted = resolve; });
	const harness = makeSdk();
	harness.session.prompt = async () => { promptStarted(); await promptGate; };
	harness.session.waitForIdle = () => new Promise(() => undefined);
	const runner = new SdkAgentRunner(harness.sdk, harness.modelRegistry, { cleanupTimeoutMs: 10 });
	const pending = runner.run(request());
	const pendingFailure = assert.rejects(pending, /cleanup.*timed out|cancel/i);
	await promptStartedGate;
	await assert.rejects(runner.close(), /close timed out|failed to close/i);
	assert.equal(harness.calls.dispose, 1);
	releasePrompt();
	await pendingFailure;
});

test("request deadline covers hanging reload and hanging session creation", async () => {
	const hangingReload = makeSdk({
		createResourceLoader() {
			return { reload: () => new Promise(() => undefined) };
		},
	});
	await assert.rejects(
		settlesWithin(
			new SdkAgentRunner(hangingReload.sdk, hangingReload.modelRegistry).run(request({ timeoutMs: 10 })),
			200,
			"hanging reload",
		),
		/timed out after 10ms/,
	);
	assert.equal(hangingReload.calls.prompt, 0);

	const hangingCreate = makeSdk({ createSession: () => new Promise(() => undefined) });
	await assert.rejects(
		settlesWithin(
			new SdkAgentRunner(hangingCreate.sdk, hangingCreate.modelRegistry).run(request({ timeoutMs: 10 })),
			200,
			"hanging create",
		),
		/timed out after 10ms/,
	);
	assert.equal(hangingCreate.calls.prompt, 0);
});

test("late session creation after timeout is immediately cleaned up without prompting", async () => {
	let releaseCreation;
	const creationGate = new Promise((resolve) => { releaseCreation = resolve; });
	let idleStarted;
	const idleStartedGate = new Promise((resolve) => { idleStarted = resolve; });
	let releaseIdle;
	const idleGate = new Promise((resolve) => { releaseIdle = resolve; });
	const harness = makeSdk();
	harness.session.waitForIdle = () => {
		idleStarted();
		return idleGate;
	};
	const originalCreate = harness.sdk.createSession;
	harness.sdk.createSession = async (options) => {
		await creationGate;
		return originalCreate(options);
	};
	const runner = new SdkAgentRunner(harness.sdk, harness.modelRegistry);
	const pending = runner.run(request({ timeoutMs: 10 }));
	const pendingFailure = assert.rejects(pending, /timed out after 10ms/);
	await new Promise((resolve) => setTimeout(resolve, 30));
	releaseCreation();
	await pendingFailure;
	await idleStartedGate;
	const closing = runner.close();
	let closeSettled = false;
	void closing.then(() => { closeSettled = true; });
	await new Promise((resolve) => setTimeout(resolve, 0));
	assert.equal(closeSettled, false, "close must join late-session cleanup tracked during setup");
	assert.equal(harness.calls.dispose, 0);
	releaseIdle();
	await closing;
	assert.equal(harness.calls.prompt, 0);
	assert.equal(harness.calls.abort, 1);
	assert.equal(harness.calls.dispose, 1);
});

test("one end-to-end deadline spans setup, prompt, and teardown", async () => {
	const delay = (ms) => new Promise((resolve) => setTimeout(resolve, ms));
	const promptDeadline = makeSdk({
		createResourceLoader() {
			return { async reload() { await delay(70); } };
		},
	});
	promptDeadline.session.prompt = async () => {
		await delay(70);
		promptDeadline.emitTerminal();
	};
	const startedAt = Date.now();
	await assert.rejects(
		new SdkAgentRunner(promptDeadline.sdk, promptDeadline.modelRegistry).run(request({ timeoutMs: 100 })),
		/timed out after 100ms/,
	);
	assert.ok(Date.now() - startedAt < 160, "setup and prompt must not each receive a fresh deadline");
	assert.equal(promptDeadline.calls.dispose, 1);

	const teardownDeadline = makeSdk({
		createResourceLoader() {
			return { async reload() { await delay(30); } };
		},
	});
	teardownDeadline.session.prompt = async () => {
		await delay(30);
		teardownDeadline.emitTerminal();
	};
	teardownDeadline.session.waitForIdle = async () => { await delay(80); };
	const teardownStartedAt = Date.now();
	await assert.rejects(
		new SdkAgentRunner(teardownDeadline.sdk, teardownDeadline.modelRegistry).run(request({ timeoutMs: 100 })),
		/cleanup.*timed out/i,
	);
	assert.ok(Date.now() - teardownStartedAt < 150, "teardown must use only the request deadline remainder");
	assert.equal(teardownDeadline.calls.dispose, 1);
});

test("concurrent close callers share cleanup completion and the same failure", async () => {
	let releasePrompt;
	let promptStarted;
	const promptStartedGate = new Promise((resolve) => { promptStarted = resolve; });
	const harness = makeSdk();
	harness.session.prompt = () => new Promise((resolve) => {
		promptStarted();
		releasePrompt = resolve;
	});
	harness.session.waitForIdle = () => new Promise(() => undefined);
	const runner = new SdkAgentRunner(harness.sdk, harness.modelRegistry, { cleanupTimeoutMs: 10 });
	const pending = runner.run(request());
	const pendingFailure = assert.rejects(pending, /cleanup.*timed out|cancel/i);
	await promptStartedGate;
	const firstClose = runner.close();
	const secondClose = runner.close();
	assert.strictEqual(secondClose, firstClose);
	const firstFailure = await firstClose.catch((error) => error);
	const secondFailure = await secondClose.catch((error) => error);
	const retryFailure = await runner.close().catch((error) => error);
	assert.strictEqual(secondFailure, firstFailure);
	assert.strictEqual(retryFailure, firstFailure);
	assert.equal(harness.calls.dispose, 1);
	releasePrompt();
	await pendingFailure;
});

test("concurrent successful close callers do not settle before cleanup finishes", async () => {
	let promptStarted;
	const promptStartedGate = new Promise((resolve) => { promptStarted = resolve; });
	let releaseIdle;
	const idleGate = new Promise((resolve) => { releaseIdle = resolve; });
	const harness = makeSdk();
	harness.session.prompt = () => new Promise(() => { promptStarted(); });
	harness.session.waitForIdle = () => idleGate;
	const runner = new SdkAgentRunner(harness.sdk, harness.modelRegistry, { cleanupTimeoutMs: 1_000 });
	const pendingFailure = assert.rejects(runner.run(request()), /cancel/i);
	await promptStartedGate;
	const firstClose = runner.close();
	const secondClose = runner.close();
	assert.strictEqual(secondClose, firstClose);
	let settled = false;
	void secondClose.then(() => { settled = true; });
	await new Promise((resolve) => setTimeout(resolve, 0));
	assert.equal(settled, false);
	assert.equal(harness.calls.dispose, 0);
	releaseIdle();
	await firstClose;
	await pendingFailure;
	assert.equal(harness.calls.dispose, 1);
	assert.strictEqual(runner.close(), firstClose);
});

test("accepts platform-aware absolute worktree paths and rejects traversal", async () => {
	for (const cwd of ["/tmp/pr-438", "C:\\repo\\pr-438", "C:/repo/pr-438", "\\\\server\\share\\pr-438"]) {
		const harness = makeSdk();
		await new SdkAgentRunner(harness.sdk, harness.modelRegistry).run(request({ cwd }));
		assert.equal(harness.sessionOptions.cwd, cwd);
	}
	for (const cwd of ["relative/path", "C:repo\\pr-438", "/tmp/../secret", "C:\\repo\\..\\secret", "\\\\server\\share\\..\\secret"]) {
		const harness = makeSdk();
		await assert.rejects(
			new SdkAgentRunner(harness.sdk, harness.modelRegistry).run(request({ cwd })),
			/absolute non-traversing path/,
		);
		assert.equal(harness.calls.reload, 0);
	}
});
