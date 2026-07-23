import assert from "node:assert/strict";
import test from "node:test";

import {
	applyRegisteredProviderConfigs,
	assertEmbeddedModelAuth,
	ExtensionModelRuntimeOwner,
} from "./embedded-model-runtime.ts";

function deferred<T>(): { promise: Promise<T>; resolve(value: T): void; reject(reason: unknown): void } {
	let resolvePromise!: (value: T) => void;
	let rejectPromise!: (reason: unknown) => void;
	const promise = new Promise<T>((resolve, reject) => {
		resolvePromise = resolve;
		rejectPromise = reject;
	});
	return { promise, resolve: resolvePromise, reject: rejectPromise };
}

test("legacy and production adapters share the exact single-flight runtime initialization", async () => {
	const authorityAgentDir = "/authority/agent";
	const owner = new ExtensionModelRuntimeOwner<object>(authorityAgentDir);
	const gate = deferred<object>();
	let createCalls = 0;
	const initialize = () => {
		createCalls += 1;
		return gate.promise;
	};
	const legacyAdapter = () => owner.acquire(authorityAgentDir, initialize);
	const productionAdapter = () => owner.acquire(authorityAgentDir, initialize);

	const legacy = legacyAdapter();
	const production = productionAdapter();
	assert.strictEqual(legacy, production);
	assert.equal(createCalls, 0, "initialization remains lazy until the promise runs");

	await Promise.resolve();
	assert.equal(createCalls, 1);
	const runtime = {};
	gate.resolve(runtime);
	assert.strictEqual(await legacy, runtime);
	assert.strictEqual(await production, runtime);
	assert.strictEqual(owner.acquire(authorityAgentDir, initialize), legacy);
	assert.equal(createCalls, 1);
});

test("a mismatched agentDir fails closed without initializing or reusing the host runtime", async () => {
	const owner = new ExtensionModelRuntimeOwner<object>("/authority/agent");
	let createCalls = 0;
	const initialize = async () => {
		createCalls += 1;
		return {};
	};

	assert.throws(
		() => owner.acquire("/different/agent", initialize),
		/agentDir does not match extension host authority/,
	);
	assert.equal(createCalls, 0);
	await owner.acquire("/authority/agent", initialize);
	assert.throws(
		() => owner.acquire("/different/agent", initialize),
		/agentDir does not match extension host authority/,
	);
	assert.equal(createCalls, 1);
});

test("failed single-flight initialization is shared and a later caller can retry", async () => {
	const authorityAgentDir = "/authority/agent";
	const owner = new ExtensionModelRuntimeOwner<object>(authorityAgentDir);
	const firstGate = deferred<object>();
	const failure = new Error("runtime initialization failed");
	let createCalls = 0;
	const initialize = () => {
		createCalls += 1;
		return createCalls === 1 ? firstGate.promise : Promise.resolve({ attempt: createCalls });
	};

	const first = owner.acquire(authorityAgentDir, initialize);
	const concurrent = owner.acquire(authorityAgentDir, initialize);
	assert.strictEqual(first, concurrent);
	await Promise.resolve();
	assert.equal(createCalls, 1);
	firstGate.reject(failure);
	await assert.rejects(first, (error) => error === failure);
	await assert.rejects(concurrent, (error) => error === failure);

	const retry = owner.acquire(authorityAgentDir, initialize);
	assert.notStrictEqual(retry, first);
	assert.deepEqual(await retry, { attempt: 2 });
	assert.equal(createCalls, 2);
});

test("registered provider configuration is applied to the shared runtime", () => {
	const calls: string[] = [];
	const runtime = {
		unregisterProvider(provider: string) { calls.push(`unregister:${provider}`); },
		registerProvider(provider: string, _config: unknown) { calls.push(`register:${provider}`); },
	};
	const registry = {
		getRegisteredProviderIds: () => ["configured", "missing"],
		getRegisteredProviderConfig: (provider: string) => provider === "configured" ? {} : undefined,
	};

	applyRegisteredProviderConfigs(runtime, registry);
	assert.deepEqual(calls, ["unregister:configured", "register:configured"]);
});

test("embedded auth accepts the runtime store and fails closed for unavailable host-only auth", () => {
	assert.doesNotThrow(() => assertEmbeddedModelAuth(
		"openai-codex",
		true,
		{ configured: true, source: "stored" },
		{ configured: true, source: "stored" },
	));
	assert.throws(
		() => assertEmbeddedModelAuth(
			"openai-codex",
			true,
			{ configured: true, source: "oauth" },
			{ configured: false },
		),
		/embedded AgentSession cannot inherit host-only OAuth for openai-codex/,
	);
	assert.throws(
		() => assertEmbeddedModelAuth(
			"openai-codex",
			false,
			{ configured: true, source: "runtime" },
			{ configured: false },
		),
		/embedded AgentSession cannot inherit host-only auth for openai-codex/,
	);
	assert.throws(
		() => assertEmbeddedModelAuth(
			"openai-codex",
			true,
			{ configured: false },
			{ configured: false },
		),
		/embedded AgentSession has no configured auth for openai-codex/,
	);
});
