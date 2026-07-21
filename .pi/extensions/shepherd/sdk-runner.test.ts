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

function makeSdk(overrides = {}) {
	const calls = { reload: 0, prompt: 0, abort: 0, idle: 0, unsubscribe: 0, dispose: 0 };
	let resourceOptions;
	let sessionOptions;
	let listener;
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
		},
		async waitForIdle() { calls.idle += 1; },
		async abort() { calls.abort += 1; },
		dispose() { calls.dispose += 1; },
		getLastAssistantText: () => JSON.stringify({
			...request().binding,
			summary: "read-only inspection passed",
			dimensions,
			observedMutation: false,
		}),
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
		emit(event) { listener?.(event); },
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
	malformed.session.getLastAssistantText = () => "not-json";
	await assert.rejects(
		new SdkAgentRunner(malformed.sdk, malformed.modelRegistry).run(request()),
		/evidence must be one JSON object/,
	);
	assert.equal(malformed.calls.dispose, 1);

	const oversized = makeSdk();
	oversized.session.getLastAssistantText = () => "x".repeat(65);
	await assert.rejects(
		new SdkAgentRunner(oversized.sdk, oversized.modelRegistry, { maxAssistantBytes: 64 }).run(request()),
		/output limit/,
	);
	assert.equal(oversized.calls.dispose, 1);
});

test("aborts and cleans up when the event budget is exceeded", async () => {
	const harness = makeSdk();
	harness.session.prompt = async () => {
		harness.emit({ type: "first" });
		harness.emit({ type: "second" });
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
	harness.session.prompt = () => new Promise((resolve) => { release = resolve; });
	const runner = new SdkAgentRunner(harness.sdk, harness.modelRegistry);
	const pending = runner.run(request());
	await new Promise((resolve) => setTimeout(resolve, 0));
	await runner.abort("other-run");
	assert.equal(harness.calls.abort, 0);
	await runner.abort("run-1");
	assert.equal(harness.calls.abort, 1);
	release();
	await pending;
});
