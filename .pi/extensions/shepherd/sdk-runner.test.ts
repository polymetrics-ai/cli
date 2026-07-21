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

function request() {
	return {
		runId: "run-1",
		laneId: "scout",
		role: "scout",
		cwd: "/tmp/pr-438",
		readOnly: true,
		provider: "openai-codex",
		model: "gpt-5.6-sol",
		thinking: "xhigh",
		tools: ["read", "grep", "find", "ls"],
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
}

function makeSdk(overrides = {}) {
	const calls = { reload: 0, prompt: 0, abort: 0, idle: 0, unsubscribe: 0, dispose: 0 };
	let resourceOptions;
	let sessionOptions;
	const session = {
		model: { provider: "openai-codex", id: "gpt-5.6-sol" },
		thinkingLevel: "xhigh",
		sessionFile: undefined,
		getActiveToolNames: () => ["read", "grep", "find", "ls"],
		subscribe: () => () => { calls.unsubscribe += 1; },
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
	return { sdk, modelRegistry, calls, session, get resourceOptions() { return resourceOptions; }, get sessionOptions() { return sessionOptions; } };
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
	assert.deepEqual(harness.sessionOptions.tools, ["read", "grep", "find", "ls"]);
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
	wrongTools.session.getActiveToolNames = () => ["read", "bash"];
	await assert.rejects(new SdkAgentRunner(wrongTools.sdk, wrongTools.modelRegistry).run(request()), /tool allowlist mismatch/);

	const persistent = makeSdk();
	persistent.session.sessionFile = "/tmp/forbidden.jsonl";
	await assert.rejects(new SdkAgentRunner(persistent.sdk, persistent.modelRegistry).run(request()), /unexpectedly enabled persistence/);
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
