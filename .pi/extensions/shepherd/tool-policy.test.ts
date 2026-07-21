import assert from "node:assert/strict";
import test from "node:test";

import {
	ToolPolicyError,
	createToolPolicy,
	redactSensitiveText,
	type HostCapability,
	type ScopedWorkspace,
	type SessionTool,
} from "./tool-policy.ts";

function text(result: Awaited<ReturnType<SessionTool["execute"]>>): string {
	return result.content.map((part) => part.text).join("");
}

function fakeWorkspace(readOutput?: string): ScopedWorkspace & {
	reads: string[];
	edits: string[];
	writes: string[];
} {
	return {
		id: "workspace-475",
		cwd: "/opaque/worktrees/issue-475",
		reads: [],
		edits: [],
		writes: [],
		async readText(path) {
			this.reads.push(path);
			return readOutput ?? `source from ${path}\nTOKEN=top-secret-value`;
		},
		async editText(path, oldText, newText) {
			this.edits.push(`${path}:${oldText}->${newText}`);
			return { changed: true, summary: `edited ${path}` };
		},
		async writeText(path, content) {
			this.writes.push(`${path}:${content}`);
			return { changed: true, summary: `wrote ${path}` };
		},
	};
}

function capability(
	name: string,
	options: { mutates?: boolean; output?: string } = {},
): HostCapability {
	return {
		name,
		description: `Typed ${name} capability`,
		mutates: options.mutates ?? false,
		parameters: {
			type: "object",
			additionalProperties: false,
			properties: { target: { type: "string", maxLength: 128 } },
			required: ["target"],
		},
		async execute() {
			return {
				status: "ok",
				summary: options.output ?? `${name} complete`,
				references: ["ref-1"],
			};
		},
	};
}

function policyInput(readOnly: boolean, readOutput?: string) {
	return {
		readOnly,
		workspace: fakeWorkspace(readOutput),
		authority: {
			workspaceId: "workspace-475",
			readPrefixes: [".pi/extensions/shepherd", ".planning/phases/475-shepherd-agent-session-runtime"],
			writePrefixes: [".pi/extensions/shepherd"],
			capabilityNames: ["host_inspect", "host_verify"],
		},
		capabilities: [capability("host_inspect"), capability("host_verify", { mutates: true })],
	};
}

test("redaction covers single-line and multiline structured secret forms", () => {
	const probes = [
		{ secret: "single-json", value: (secret: string) => `"token": "${secret}"` },
		{ secret: "single-yaml", value: (secret: string) => `password: '${secret}'` },
		{ secret: "quoted-yaml", value: (secret: string) => `client_secret: "${secret}\n  continuation"` },
		{ secret: "block-yaml", value: (secret: string) => `client_secret: |-\n  ${secret}\n  continuation\nsafe: retained` },
		{ secret: "client-assignment", value: (secret: string) => `clientSecret=${secret}` },
		{ secret: "quoted-bearer", value: (secret: string) => `Authorization: "Bearer ${secret}\n  continuation"` },
		{ secret: "flow-yaml", value: (secret: string) => `{ safe: retained, client_secret: ${secret} with spaces, enabled: true }` },
		{ secret: "spaced-yaml", value: (secret: string) => `client_secret: ${secret} with spaces\nsafe: retained` },
		{ secret: "apostrophe-prefix", value: (secret: string) => `Don't skip the structured value below.\nclient_secret: "${secret}"` },
		{ secret: "nested-flow", value: (secret: string) => `token: harmless prose before { client_secret: ${secret} with spaces, safe: retained }` },
	];

	for (const probe of probes) {
		const secret = ["synthetic", probe.secret, "issue-475"].join("-");
		const value = probe.value(secret);
		const redacted = redactSensitiveText(value);
		assert.equal(redacted.includes(secret), false, value);
		assert.match(redacted, /\[REDACTED\]/, value);
	}
});

test("redaction preserves harmless assignment-like prose byte-identically", () => {
	const controls = [
		"The token bucket algorithm uses bearer capabilities, and authorization is described here.",
		"token: describes a lexical unit in parser documentation.",
		"password = number of characters accepted by this form.",
		"secret: a surprising detail in a story.",
		"Authorization: Bearer is the HTTP authentication scheme discussed here.",
		"In prose, client_secret: explains the OAuth field name without assigning a value.",
	];
	for (const control of controls) assert.equal(redactSensitiveText(control), control);
});

test("redaction balances nested flow values before scanning later sensitive siblings", () => {
	const secret = ["synthetic", "nested-sibling", "issue-475"].join("-");
	const value = `{ token: { retained: true }, client_secret: ${secret} with spaces, safe: retained }`;
	const redacted = redactSensitiveText(value);

	assert.equal(redacted.includes(secret), false);
	assert.match(redacted, /\[REDACTED\]/);
});

test("redaction keeps multiline nested values owned before later same-line sensitive siblings", () => {
	const secret = ["synthetic", "multiline-nested-sibling", "issue-475"].join("-");
	const value = `{ token: {\n retained: true }, client_secret: ${secret} with spaces, safe: retained }`;
	const redacted = redactSensitiveText(value);

	assert.equal(redacted.includes(secret), false);
	assert.match(redacted, /\[REDACTED\]/);
});

test("redaction treats punctuation-adjacent apostrophes as unquoted scalar text", () => {
	const secret = ["synthetic", "punctuation-apostrophe", "issue-475"].join("-");
	const value = `{ flavor: rock-'n-roll, client_secret: ${secret} with spaces, safe: retained }`;
	const redacted = redactSensitiveText(value);

	assert.equal(redacted.includes(secret), false);
	assert.match(redacted, /\[REDACTED\]/);
});

test("redaction resets unmatched leading prose apostrophes at the next structured line", () => {
	const secret = ["synthetic", "leading-apostrophe", "issue-475"].join("-");
	const value = `'This leading apostrophe is ordinary prose, not a YAML scalar\nclient_secret: ${secret} with spaces`;
	const redacted = redactSensitiveText(value);

	assert.equal(redacted.includes(secret), false);
	assert.match(redacted, /\[REDACTED\]/);
});

test("redaction leaves ordinary braces and flow-shaped comments byte-identical", () => {
	const controls = [
		"Use { to denote an opening delimiter in parser prose.\ntoken: describes a lexical unit in documentation.",
		"# example only: { client_secret: placeholder }\nsecret: a surprising detail in a story.",
		"{ flavor: rock-'n-roll, safe: retained }",
	];

	assert.deepEqual(controls.map(redactSensitiveText), controls);
});

test("redaction line-boundary discovery remains near-linear for dense single-line flows", () => {
	type ScanMetrics = { lineBoundaryVisits: number };
	type InstrumentedRedactor = (value: string, metrics: ScanMetrics) => string;
	const instrumentedRedactor = redactSensitiveText as InstrumentedRedactor;
	const denseFlow = (minimumBytes: number): string => {
		const assignment = "token: synthetic-linear-secret, ";
		const count = Math.ceil(minimumBytes / assignment.length);
		return `{ ${assignment.repeat(count)}safe: retained }`;
	};
	const samples = [25, 50, 100].map((kibibytes) => denseFlow(kibibytes * 1024));
	const visits = samples.map((value) => {
		const metrics: ScanMetrics = { lineBoundaryVisits: 0 };
		const redacted = instrumentedRedactor(value, metrics);
		assert.equal(redacted.includes("synthetic-linear-secret"), false);
		return metrics.lineBoundaryVisits;
	});

	for (const [index, count] of visits.entries()) {
		assert.ok(count > 0, `missing scan metrics for ${samples[index].length} bytes`);
		assert.ok(count <= samples[index].length * 2, `${count} visits for ${samples[index].length} bytes`);
	}
	for (let index = 1; index < visits.length; index += 1) {
		const inputRatio = samples[index].length / samples[index - 1].length;
		const visitRatio = visits[index] / visits[index - 1];
		assert.ok(visitRatio <= inputRatio * 1.15, `${visitRatio} work ratio for ${inputRatio} input ratio`);
	}
});

test("read-only policy exposes workspace reads and non-mutating typed capabilities only", async () => {
	const input = policyInput(true);
	const policy = createToolPolicy(input);

	assert.deepEqual(policy.names, ["workspace_read", "host_inspect"]);
	assert.equal(policy.tools.some((tool) => tool.name === "workspace_edit"), false);
	assert.equal(policy.tools.some((tool) => tool.name === "workspace_write"), false);
	assert.equal(policy.tools.some((tool) => tool.name === "host_verify"), false);

	const read = policy.tools.find((tool) => tool.name === "workspace_read");
	assert.ok(read);
	const result = await read.execute(
		"call-1",
		{ path: ".pi/extensions/shepherd/controller.ts", offset: 0, limit: 100 },
		undefined,
	);
	assert.match(text(result), /source from \.pi\/extensions\/shepherd\/controller\.ts/);
	assert.doesNotMatch(text(result), /top-secret-value/);
	assert.match(text(result), /\[REDACTED\]/);
	assert.deepEqual(input.workspace.reads, [".pi/extensions/shepherd/controller.ts"]);
});

test("workspace reads redact multiline nested and punctuation-apostrophe sibling secrets", async () => {
	const multilineSecret = ["synthetic", "workspace-multiline-nested", "issue-475"].join("-");
	const apostropheSecret = ["synthetic", "workspace-punctuation-apostrophe", "issue-475"].join("-");
	const input = policyInput(true, [
		`{ token: {\n retained: true }, client_secret: ${multilineSecret} with spaces, safe: retained }`,
		`{ flavor: rock-'n-roll, client_secret: ${apostropheSecret} with spaces, safe: retained }`,
	].join("\n"));
	const policy = createToolPolicy(input);
	const read = policy.tools.find((tool) => tool.name === "workspace_read");
	assert.ok(read);

	const result = text(await read.execute(
		"read-cycle-6",
		{ path: ".pi/extensions/shepherd/controller.ts", offset: 0, limit: 1000 },
		undefined,
	));
	assert.equal(result.includes(multilineSecret), false);
	assert.equal(result.includes(apostropheSecret), false);
	assert.match(result, /\[REDACTED\]/);
});

test("mutating policy exposes only workspace-bound read/edit/write and allowlisted typed capabilities", async () => {
	const input = policyInput(false);
	const policy = createToolPolicy(input);

	assert.deepEqual(policy.names, [
		"workspace_read",
		"workspace_edit",
		"workspace_write",
		"host_inspect",
		"host_verify",
	]);
	assert.equal(policy.names.includes("bash"), false);
	assert.equal(policy.names.includes("subagent"), false);

	const edit = policy.tools.find((tool) => tool.name === "workspace_edit");
	const write = policy.tools.find((tool) => tool.name === "workspace_write");
	assert.ok(edit);
	assert.ok(write);
	await edit.execute("call-edit", {
		path: ".pi/extensions/shepherd/tool-policy.ts",
		oldText: "old",
		newText: "new",
	}, undefined);
	await write.execute("call-write", {
		path: ".pi/extensions/shepherd/role-prompts.ts",
		content: "bounded",
	}, undefined);
	assert.equal(input.workspace.edits.length, 1);
	assert.equal(input.workspace.writes.length, 1);
});

test("workspace tools reject traversal, absolute, control-character, sensitive, and out-of-scope paths", async () => {
	const policy = createToolPolicy(policyInput(false));
	const read = policy.tools.find((tool) => tool.name === "workspace_read");
	const write = policy.tools.find((tool) => tool.name === "workspace_write");
	assert.ok(read);
	assert.ok(write);

	for (const path of [
		"../outside",
		"/etc/passwd",
		".pi/extensions/shepherd/../../.env",
		".pi/extensions/shepherd/file\u0000.ts",
		".env",
		".git/config",
		"credentials.json",
		"internal/not-owned.go",
	]) {
		await assert.rejects(
			() => read.execute("read", { path }, undefined),
			ToolPolicyError,
			path,
		);
	}

	await assert.rejects(
		() => write.execute("write", {
			path: ".planning/phases/475-shepherd-agent-session-runtime/PLAN.md",
			content: "not in write prefixes",
		}, undefined),
		ToolPolicyError,
	);
});

test("capability authority is closed and cannot add generic or recursive tools", () => {
	const forbidden = [
		"bash",
		"shell_exec",
		"host_http_write",
		"host_sql_write",
		"host_subagent",
		"host_spawn_agent",
		"host_orchestrate",
		"host_delegate",
		"host_secret_read",
	];
	for (const name of forbidden) {
		assert.throws(
			() => createToolPolicy({
				...policyInput(false),
				authority: {
					...policyInput(false).authority,
					capabilityNames: [name],
				},
				capabilities: [capability(name)],
			}),
			ToolPolicyError,
			name,
		);
	}
});

test("policy rejects workspace identity mismatch, duplicate capabilities, and undeclared capability injection", () => {
	assert.throws(() => createToolPolicy({
		...policyInput(false),
		authority: { ...policyInput(false).authority, workspaceId: "swapped-workspace" },
	}), /workspace identity/i);

	assert.throws(() => createToolPolicy({
		...policyInput(false),
		capabilities: [capability("host_inspect"), capability("host_inspect")],
	}), /duplicate/i);

	assert.throws(() => createToolPolicy({
		...policyInput(false),
		capabilities: [
			capability("host_inspect"),
			capability("host_verify", { mutates: true }),
			capability("host_publish"),
		],
	}), /undeclared capability/i);
});

test("tool inputs and outputs are bounded and host summaries are redacted", async () => {
	const blockSecret = ["synthetic", "tool-block", "issue-475"].join("-");
	const flowSecret = ["synthetic", "tool-flow", "issue-475"].join("-");
	const spacedSecret = ["synthetic", "tool-spaced", "issue-475"].join("-");
	const input = policyInput(false);
	input.capabilities = [
		capability("host_inspect", { output: [
			`client_secret: |-\n  ${blockSecret}\n  continuation`,
			`{ safe: retained, client_secret: ${flowSecret} with spaces, enabled: true }`,
			`client_secret: ${spacedSecret} with spaces\nsafe: retained`,
		].join("\n") }),
		capability("host_verify", { mutates: true }),
	];
	const policy = createToolPolicy(input, { maxToolOutputBytes: 256 });
	const inspect = policy.tools.find((tool) => tool.name === "host_inspect");
	const write = policy.tools.find((tool) => tool.name === "workspace_write");
	assert.ok(inspect);
	assert.ok(write);

	const result = await inspect.execute("inspect", { target: "owned" }, undefined);
	assert.equal(text(result).includes(blockSecret), false);
	assert.equal(text(result).includes(flowSecret), false);
	assert.equal(text(result).includes(spacedSecret), false);
	assert.match(text(result), /\[REDACTED\]/);

	await assert.rejects(
		() => write.execute("large", {
			path: ".pi/extensions/shepherd/x.ts",
			content: "x".repeat(300_000),
		}, undefined),
		/bounded|large|limit/i,
	);
});

test("typed tool output redacts nested-flow and apostrophe-boundary siblings", async () => {
	const probes = [
		{
			secret: ["synthetic", "tool-nested-sibling", "issue-475"].join("-"),
			value(secret: string) {
				return `{ token: { retained: true }, client_secret: ${secret} with spaces, safe: retained }`;
			},
		},
		{
			secret: ["synthetic", "tool-leading-apostrophe", "issue-475"].join("-"),
			value(secret: string) {
				return `'This leading apostrophe is ordinary prose\nclient_secret: ${secret} with spaces`;
			},
		},
		{
			secret: ["synthetic", "tool-multiline-nested", "issue-475"].join("-"),
			value(secret: string) {
				return `{ token: {\n retained: true }, client_secret: ${secret} with spaces, safe: retained }`;
			},
		},
		{
			secret: ["synthetic", "tool-punctuation-apostrophe", "issue-475"].join("-"),
			value(secret: string) {
				return `{ flavor: rock-'n-roll, client_secret: ${secret} with spaces, safe: retained }`;
			},
		},
	];
	const rendered: string[] = [];
	for (const probe of probes) {
		const input = policyInput(false);
		input.capabilities = [
			capability("host_inspect", { output: probe.value(probe.secret) }),
			capability("host_verify", { mutates: true }),
		];
		const policy = createToolPolicy(input, { maxToolOutputBytes: 512 });
		const inspect = policy.tools.find((tool) => tool.name === "host_inspect");
		assert.ok(inspect);
		rendered.push(text(await inspect.execute("inspect", { target: "owned" }, undefined)));
	}

	assert.deepEqual(rendered.map((value, index) => value.includes(probes[index].secret)), [false, false, false, false]);
	assert.deepEqual(rendered.map((value) => /\[REDACTED\]/.test(value)), [true, true, true, true]);
});
