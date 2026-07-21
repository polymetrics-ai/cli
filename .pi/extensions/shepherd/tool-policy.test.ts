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

function fakeWorkspace(): ScopedWorkspace & {
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
			return `source from ${path}\nTOKEN=top-secret-value`;
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

function policyInput(readOnly: boolean) {
	return {
		readOnly,
		workspace: fakeWorkspace(),
		authority: {
			workspaceId: "workspace-475",
			readPrefixes: [".pi/extensions/shepherd", ".planning/phases/475-shepherd-agent-session-runtime"],
			writePrefixes: [".pi/extensions/shepherd"],
			capabilityNames: ["host_inspect", "host_verify"],
		},
		capabilities: [capability("host_inspect"), capability("host_verify", { mutates: true })],
	};
}

test("redaction covers quoted JSON/YAML assignments and quoted Bearer values without changing ordinary prose", () => {
	const secret = ["synthetic", "credential", "issue-475"].join("-");
	const probes = [
		`"token": "${secret}"`,
		`password: '${secret}'`,
		`"Authorization": "Bearer ${secret}"`,
		`Authorization: 'Bearer ${secret}'`,
	];

	for (const probe of probes) {
		const redacted = redactSensitiveText(probe);
		assert.equal(redacted.includes(secret), false, probe);
		assert.match(redacted, /\[REDACTED\]/, probe);
	}

	const ordinary = "The token bucket algorithm uses bearer capabilities, and authorization is described here.";
	assert.equal(redactSensitiveText(ordinary), ordinary);
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
	const secret = ["synthetic", "tool-output", "issue-475"].join("-");
	const input = policyInput(false);
	input.capabilities = [
		capability("host_inspect", { output: `{"Authorization": "Bearer ${secret}"}` }),
		capability("host_verify", { mutates: true }),
	];
	const policy = createToolPolicy(input, { maxToolOutputBytes: 256 });
	const inspect = policy.tools.find((tool) => tool.name === "host_inspect");
	const write = policy.tools.find((tool) => tool.name === "workspace_write");
	assert.ok(inspect);
	assert.ok(write);

	const result = await inspect.execute("inspect", { target: "owned" }, undefined);
	assert.equal(text(result).includes(secret), false);
	assert.match(text(result), /\[REDACTED\]/);

	await assert.rejects(
		() => write.execute("large", {
			path: ".pi/extensions/shepherd/x.ts",
			content: "x".repeat(300_000),
		}, undefined),
		/bounded|large|limit/i,
	);
});
