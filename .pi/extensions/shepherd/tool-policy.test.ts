import assert from "node:assert/strict";
import { dirname, join } from "node:path";
import test from "node:test";
import { pathToFileURL } from "node:url";

import {
	ToolPolicyError,
	createToolPolicy,
	normalizeScopedPrefixes,
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

function cycle7SecretPayload(prefix: string): { value: string; markers: string[] } {
	const markers = {
		outerFlow: `synthetic-${prefix}-outer-flow-475`,
		indented: `synthetic-${prefix}-indented-475`,
		keyOnly: `synthetic-${prefix}-key-only-475`,
		continued: `synthetic-${prefix}-continued-475`,
		numeric: `9475475475${String(prefix.length).padStart(2, "0")}`,
		basic: `synthetic-${prefix}-basic-475`,
		nonBearer: `synthetic-${prefix}-non-bearer-475`,
		awsAlias: `synthetic-${prefix}-aws-alias-475`,
		databaseAlias: `synthetic-${prefix}-database-alias-475`,
		githubAlias: `synthetic-${prefix}-github-alias-475`,
		pkcs8: `synthetic-${prefix}-pkcs8-475`,
		unmatched: `synthetic-${prefix}-unmatched-quote-475`,
		afterUnmatched: `synthetic-${prefix}-after-unmatched-475`,
	};
	return {
		value: [
			"{",
			`  safe: retained, client_secret: ${markers.outerFlow} with spaces, enabled: true`,
			"}",
			`  token: ${markers.indented} with spaces`,
			"client_secret:",
			`  ${markers.keyOnly}`,
			"  continuation",
			"client_secret: first-segment",
			`  ${markers.continued} with spaces`,
			`access_token: ${markers.numeric}`,
			`Authorization: Basic ${markers.basic}`,
			`Authorization: ApiKey ${markers.nonBearer}`,
			`AWS_SECRET_ACCESS_KEY=${markers.awsAlias}`,
			`DATABASE_URL=${markers.databaseAlias}`,
			`GITHUB_TOKEN=${markers.githubAlias}`,
			"-----BEGIN PRIVATE KEY-----",
			markers.pkcs8,
			"-----END PRIVATE KEY-----",
			`Authorization: "Basic ${markers.unmatched}`,
			`client_secret: ${markers.afterUnmatched} with spaces`,
		].join("\n"),
		markers: Object.values(markers),
	};
}

function cycle8SecretPayload(prefix: string): { value: string; markers: string[] } {
	const markers = {
		digest: `synthetic-${prefix}-digest-475`,
		signature: `synthetic-${prefix}-signature-475`,
		awsAuth: `synthetic-${prefix}-aws-auth-475`,
		commaSuffix: `synthetic-${prefix}-comma-suffix-475`,
		flowKeyOnly: `synthetic-${prefix}-flow-key-only-475`,
		flowContinued: `synthetic-${prefix}-flow-continued-475`,
		sequenceKeyOnly: `synthetic-${prefix}-sequence-key-only-475`,
		sequenceContinued: `synthetic-${prefix}-sequence-continued-475`,
		escapedClientSecret: `synthetic-${prefix}-escaped-client-secret-475`,
		escapedToken: `synthetic-${prefix}-escaped-token-475`,
		malformedEscapedSecret: `synthetic-${prefix}-malformed-escaped-secret-475`,
	};
	return {
		value: [
			`Authorization: Digest username="public", realm="example", response="${markers.digest}"`,
			`Authorization: Signature keyId="public", algorithm="rsa-sha256", signature="${markers.signature}"`,
			`Authorization: AWS4-HMAC-SHA256 Credential=public, SignedHeaders=host, Signature=${markers.awsAuth}`,
			`client_secret: prefix,${markers.commaSuffix}`,
			"{",
			"  client_secret:",
			`    ${markers.flowKeyOnly},`,
			"  safe: retained",
			"}",
			"{",
			"  client_secret: prefix",
			`    ${markers.flowContinued},`,
			"  safe: retained",
			"}",
			"[",
			"  { client_secret:",
			`      ${markers.sequenceKeyOnly}, safe: retained },`,
			"  { client_secret: prefix",
			`      ${markers.sequenceContinued}, safe: retained }`,
			"]",
			`{"client\\u005fsecret":"${markers.escapedClientSecret}","safe":true}`,
			`{"to\\u006ben":"${markers.escapedToken}"}`,
			`{"client_secret\\u00ZZ":"${markers.malformedEscapedSecret}"}`,
		].join("\n"),
		markers: Object.values(markers),
	};
}

function cycle9SecretPayload(prefix: string): { value: string; markers: string[] } {
	const markers = {
		tokenEquals: `synthetic-${prefix}-token-equals-475`,
		passwordEquals: `synthetic-${prefix}-password-equals-475`,
		secretEquals: `synthetic-${prefix}-secret-equals-475`,
		opaqueAuthorization: `synthetic-${prefix}-opaque-authorization-475`,
		implicitFlow: `synthetic-${prefix}-implicit-flow-475`,
		urlUserinfo: `synthetic-${prefix}-url-userinfo-475`,
		urlQuery: `synthetic-${prefix}-url-query-475`,
		registryAuth: `synthetic-${prefix}-registry-auth-475`,
		malformedMiddle: `synthetic-${prefix}-malformed-middle-475`,
		escaped63: `synthetic-${prefix}-escaped-63-475`,
		escaped64: `synthetic-${prefix}-escaped-64-475`,
		escaped65: `synthetic-${prefix}-escaped-65-475`,
	};
	const fullyEscapedKey = (length: number): string => {
		const decoded = `${"a".repeat(length - "token".length)}token`;
		return [...decoded].map((character) =>
			`\\u${character.charCodeAt(0).toString(16).padStart(4, "0")}`).join("");
	};
	return {
		value: [
			`token=${markers.tokenEquals} with spaces`,
			`password = ${markers.passwordEquals} with spaces`,
			`secret=${markers.secretEquals} with spaces`,
			`Authorization: ${markers.opaqueAuthorization}`,
			`[client_secret: ${markers.implicitFlow}]`,
			`request failed https://public:${markers.urlUserinfo}@x.invalid/path`,
			`request failed https://x.invalid/path?access_token=${markers.urlQuery}&safe=retained`,
			`//registry.npmjs.org/:_authToken=${markers.registryAuth}`,
			`{"to\\u00ZZken":"${markers.malformedMiddle}"}`,
			`{"${fullyEscapedKey(63)}":"${markers.escaped63}"}`,
			`{"${fullyEscapedKey(64)}":"${markers.escaped64}"}`,
			`{"${fullyEscapedKey(65)}":"${markers.escaped65}"}`,
		].join("\n"),
		markers: Object.values(markers),
	};
}

function leakedMarkers(value: string, markers: readonly string[]): string[] {
	return markers.filter((marker) => value.includes(marker));
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
		"password: number of characters accepted by this form.",
		"secret: a surprising detail in a story.",
		"Authorization: Bearer is the HTTP authentication scheme discussed here.",
		"In prose, client_secret: explains the OAuth field name without assigning a value.",
	];
	for (const control of controls) assert.equal(redactSensitiveText(control), control);
	assert.equal(
		redactSensitiveText("password = number of characters accepted by this form."),
		"password = [REDACTED]",
	);
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

test("cycle 7 direct redaction covers multiline flow, YAML continuation, aliases, authorization, and PKCS8", () => {
	const payload = cycle7SecretPayload("direct");
	const redacted = redactSensitiveText(payload.value);

	assert.deepEqual(leakedMarkers(redacted, payload.markers), []);
	assert.match(redacted, /\[REDACTED\]/);
});

test("cycle 7 preserves a harmless structurally quoted multiline scalar byte-identically", () => {
	const control = [
		"message: \"The following lines document configuration vocabulary:",
		"  client_secret: names an OAuth field without carrying a value.",
		"  Authorization: Basic names an authentication scheme.\"",
		"safe: retained",
	].join("\n");

	assert.equal(redactSensitiveText(control), control);
});

test("cycle 7 padded-flow diagnostics account for all scanner work near-linearly", () => {
	type ScanMetrics = {
		lineBoundaryVisits: number;
		keyStartVisits: number;
		totalCharacterVisits: number;
	};
	type InstrumentedRedactor = (value: string, metrics: ScanMetrics) => string;
	const instrumentedRedactor = redactSensitiveText as unknown as InstrumentedRedactor;
	const paddedFlow = (minimumBytes: number): string => {
		const padding = " ".repeat(Math.floor(minimumBytes / 3));
		const assignment = "token: synthetic-padded-flow-secret, ";
		const assignmentCount = Math.ceil((minimumBytes - padding.length) / assignment.length);
		return `${padding}{ ${assignment.repeat(assignmentCount)}safe: retained }`;
	};
	const samples = [25, 50, 100].map((kibibytes) => paddedFlow(kibibytes * 1024));
	const observations = samples.map((value) => {
		const metrics: ScanMetrics = {
			lineBoundaryVisits: 0,
			keyStartVisits: 0,
			totalCharacterVisits: 0,
		};
		const redacted = instrumentedRedactor(value, metrics);
		return { value, redacted, metrics };
	});
	const work = observations.map(({ metrics }) => metrics.totalCharacterVisits);

	assert.deepEqual({
		redacted: observations.every(({ redacted }) => !redacted.includes("synthetic-padded-flow-secret")),
		metricsPresent: observations.every(({ metrics }) =>
			metrics.lineBoundaryVisits > 0 && metrics.keyStartVisits > 0 && metrics.totalCharacterVisits > 0),
		bounded: observations.every(({ value, metrics }) => metrics.totalCharacterVisits <= value.length * 8),
		nearDoubling: work.slice(1).every((count, index) => {
			const inputRatio = samples[index + 1].length / samples[index].length;
			return count / work[index] <= inputRatio * 1.25;
		}),
	}, {
		redacted: true,
		metricsPresent: true,
		bounded: true,
		nearDoubling: true,
	});
});

test("cycle 8 direct redaction closes line commas, multiline-flow scalars, and escaped quoted keys", () => {
	const payload = cycle8SecretPayload("direct");
	const redacted = redactSensitiveText(payload.value);

	assert.deepEqual(leakedMarkers(redacted, payload.markers), []);
	assert.match(redacted, /\[REDACTED\]/);
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

test("cycle 7 workspace reads apply the complete structured secret vocabulary", async () => {
	const payload = cycle7SecretPayload("workspace");
	const input = policyInput(true, payload.value);
	const policy = createToolPolicy(input);
	const read = policy.tools.find((tool) => tool.name === "workspace_read");
	assert.ok(read);

	const result = text(await read.execute(
		"read-cycle-7",
		{ path: ".pi/extensions/shepherd/controller.ts", offset: 0, limit: 4096 },
		undefined,
	));
	assert.deepEqual(leakedMarkers(result, payload.markers), []);
	assert.match(result, /\[REDACTED\]/);
});

test("cycle 8 workspace reads close line commas, multiline-flow scalars, and escaped quoted keys", async () => {
	const payload = cycle8SecretPayload("workspace");
	const input = policyInput(true, payload.value);
	const policy = createToolPolicy(input);
	const read = policy.tools.find((tool) => tool.name === "workspace_read");
	assert.ok(read);

	const result = text(await read.execute(
		"read-cycle-8",
		{ path: ".pi/extensions/shepherd/controller.ts", offset: 0, limit: 4096 },
		undefined,
	));
	assert.deepEqual(leakedMarkers(result, payload.markers), []);
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

test("cycle 7 typed capability output applies the complete structured secret vocabulary", async () => {
	const payload = cycle7SecretPayload("typed-tool");
	const input = policyInput(false);
	input.capabilities = [
		capability("host_inspect", { output: payload.value }),
		capability("host_verify", { mutates: true }),
	];
	const policy = createToolPolicy(input, { maxToolOutputBytes: 8 * 1024 });
	const inspect = policy.tools.find((tool) => tool.name === "host_inspect");
	assert.ok(inspect);

	const result = text(await inspect.execute("inspect-cycle-7", { target: "owned" }, undefined));
	assert.deepEqual(leakedMarkers(result, payload.markers), []);
	assert.match(result, /\[REDACTED\]/);
});

test("cycle 8 mutation and typed capability outputs share the complete parser closure", async () => {
	const mutationPayload = cycle8SecretPayload("mutation");
	const capabilityPayload = cycle8SecretPayload("typed-tool");
	const referenceMarker = "synthetic-typed-reference-escaped-key-475";
	const input = policyInput(false);
	input.workspace.editText = async () => ({ changed: true, summary: mutationPayload.value });
	input.workspace.writeText = async () => ({ changed: true, summary: mutationPayload.value });
	input.capabilities = [
		{
			...capability("host_inspect"),
			async execute() {
				return {
					status: "ok" as const,
					summary: capabilityPayload.value,
					references: [`{"client\\u005fsecret":"${referenceMarker}"}`],
				};
			},
		},
		capability("host_verify", { mutates: true }),
	];
	const policy = createToolPolicy(input, { maxToolOutputBytes: 16 * 1024 });
	const edit = policy.tools.find((tool) => tool.name === "workspace_edit");
	const write = policy.tools.find((tool) => tool.name === "workspace_write");
	const inspect = policy.tools.find((tool) => tool.name === "host_inspect");
	assert.ok(edit);
	assert.ok(write);
	assert.ok(inspect);

	const rendered = [
		text(await edit.execute("edit-cycle-8", {
			path: ".pi/extensions/shepherd/tool-policy.ts",
			oldText: "old",
			newText: "new",
		}, undefined)),
		text(await write.execute("write-cycle-8", {
			path: ".pi/extensions/shepherd/tool-policy.ts",
			content: "new",
		}, undefined)),
		text(await inspect.execute("inspect-cycle-8", { target: "owned" }, undefined)),
	].join("\n");
	assert.deepEqual(leakedMarkers(rendered, [
		...mutationPayload.markers,
		...capabilityPayload.markers,
		referenceMarker,
	]), []);
	assert.match(rendered, /\[REDACTED\]/);
});

test("cycle 8 tool-policy options reject one above every embedded hard ceiling", () => {
	const cases = [
		["maxToolOutputBytes", { maxToolOutputBytes: 256 * 1024 + 1 }],
		["maxReadCharacters", { maxReadCharacters: 256 * 1024 + 1 }],
		["maxWriteCharacters", { maxWriteCharacters: 1024 * 1024 + 1 }],
	] as const;
	const accepted: string[] = [];
	for (const [name, options] of cases) {
		try {
			createToolPolicy(policyInput(false), options);
			accepted.push(name);
		} catch (error) {
			assert.match(String(error), /bound|maximum|max|limit|exceed/i, name);
		}
	}
	assert.deepEqual(accepted, []);
});

test("cycle 9 capability schemas are bounded accessor-free deep immutable snapshots", () => {
	const enumValues = ["read"];
	const nestedSchema = {
		type: "object",
		additionalProperties: false,
		properties: {
			action: { type: "string", enum: enumValues },
		},
		required: ["action"],
	};
	const input = policyInput(false);
	input.capabilities = [
		{ ...capability("host_inspect"), parameters: nestedSchema },
		capability("host_verify", { mutates: true }),
	];
	const policy = createToolPolicy(input);
	const inspect = policy.tools.find((tool) => tool.name === "host_inspect");
	assert.ok(inspect);
	enumValues[0] = "delete";
	const parameters = inspect.parameters as {
		properties: { action: { enum: string[] } };
	};

	let accessorReads = 0;
	const accessorProperties: Record<string, unknown> = {};
	Object.defineProperty(accessorProperties, "action", {
		enumerable: true,
		get() {
			accessorReads += 1;
			return { type: "string" };
		},
	});
	const cyclic: Record<string, unknown> = { type: "object", additionalProperties: false, properties: {} };
	cyclic.self = cyclic;
	const deep: Record<string, unknown> = { type: "object", additionalProperties: false, properties: {} };
	let cursor = deep.properties as Record<string, unknown>;
	for (let index = 0; index < 80; index += 1) {
		const next = { type: "object", additionalProperties: false, properties: {} as Record<string, unknown> };
		cursor[`level${index}`] = next;
		cursor = next.properties;
	}
	const symbolSchema = { type: "object", additionalProperties: false, properties: {} } as Record<PropertyKey, unknown>;
	symbolSchema[Symbol("hidden")] = "forbidden";
	const sparseEnum = new Array<string>(1_000);
	sparseEnum[999] = "read";
	const sparse = {
		type: "object",
		additionalProperties: false,
		properties: { action: { type: "string", enum: sparseEnum } },
	};
	let proxyOwnKeys = 0;
	const proxy = new Proxy({ type: "object", additionalProperties: false, properties: {} }, {
		ownKeys(target) {
			proxyOwnKeys += 1;
			return Reflect.ownKeys(target);
		},
	});
	const invalidSchemas: Array<[string, Readonly<Record<string, unknown>>]> = [
		["accessor", { type: "object", additionalProperties: false, properties: accessorProperties }],
		["cycle", cyclic],
		["deep", deep],
		["symbol", symbolSchema],
		["sparse", sparse],
		["proxy", proxy],
	];
	const accepted: string[] = [];
	for (const [name, parameters] of invalidSchemas) {
		try {
			createToolPolicy({
				...policyInput(false),
				authority: { ...policyInput(false).authority, capabilityNames: ["host_inspect"] },
				capabilities: [{ ...capability("host_inspect"), parameters }],
			});
			accepted.push(name);
		} catch (error) {
			assert.ok(error instanceof ToolPolicyError, name);
		}
	}

	assert.deepEqual({
		retainedEnum: parameters.properties.action.enum,
		deepFrozen: [
			inspect.parameters,
			(inspect.parameters as { properties: object }).properties,
			parameters.properties.action,
			parameters.properties.action.enum,
		].every(Object.isFrozen),
		accepted,
		accessorReads,
		proxyOwnKeys,
		policyNamesFrozen: Object.isFrozen(policy.names),
		policyToolsFrozen: Object.isFrozen(policy.tools),
	}, {
		retainedEnum: ["read"],
		deepFrozen: true,
		accepted: ["symbol"],
		accessorReads: 0,
		proxyOwnKeys: 0,
		policyNamesFrozen: true,
		policyToolsFrozen: true,
	});
});

test("cycle 9 workspace and capability results are captured once into immutable DTOs", async () => {
	const workspaceInput = policyInput(false);
	let changedReads = 0;
	let mutationSummaryReads = 0;
	workspaceInput.workspace.editText = async () => {
		const result = {} as { changed: boolean; summary: string };
		Object.defineProperties(result, {
			changed: {
				enumerable: true,
				get() { changedReads += 1; return changedReads === 1; },
			},
			summary: {
				enumerable: true,
				get() { mutationSummaryReads += 1; return mutationSummaryReads === 1 ? "original mutation" : "mutated mutation"; },
			},
		});
		return result;
	};
	let statusReads = 0;
	let capabilitySummaryReads = 0;
	let referencesReads = 0;
	workspaceInput.capabilities = [
		{
			...capability("host_inspect"),
			async execute() {
				const result = {} as { status: "ok"; summary: string; references: string[] };
				Object.defineProperties(result, {
					status: { enumerable: true, get() { statusReads += 1; return "ok"; } },
					summary: {
						enumerable: true,
						get() { capabilitySummaryReads += 1; return "original capability"; },
					},
					references: { enumerable: true, get() { referencesReads += 1; return ["original reference"]; } },
				});
				return result;
			},
		},
		capability("host_verify", { mutates: true }),
	];
	const policy = createToolPolicy(workspaceInput);
	const edit = policy.tools.find((tool) => tool.name === "workspace_edit");
	const inspect = policy.tools.find((tool) => tool.name === "host_inspect");
	assert.ok(edit);
	assert.ok(inspect);
	const editResult = await edit.execute("cycle9-edit", {
		path: ".pi/extensions/shepherd/tool-policy.ts",
		oldText: "old",
		newText: "new",
	}, undefined);
	const capabilityResult = await inspect.execute("cycle9-inspect", { target: "owned" }, undefined);

	assert.deepEqual({
		edit: JSON.parse(text(editResult)),
		capability: JSON.parse(text(capabilityResult)),
		reads: { changedReads, mutationSummaryReads, statusReads, capabilitySummaryReads, referencesReads },
		details: [editResult, capabilityResult].map((result) => [Object.hasOwn(result, "details"), result.details]),
		immutableContent: [editResult, capabilityResult].every((result) =>
			Object.isFrozen(result) && Object.isFrozen(result.content) && result.content.every(Object.isFrozen)),
	}, {
		edit: { changed: true, summary: "original mutation" },
		capability: { status: "ok", summary: "original capability", references: ["original reference"] },
		reads: { changedReads: 1, mutationSummaryReads: 1, statusReads: 1, capabilitySummaryReads: 1, referencesReads: 1 },
		details: [[true, null], [true, null]],
		immutableContent: true,
	});
});

test("cycle 9 redaction grammar closes equals, opaque auth, URLs, implicit flow, and escaped-key boundaries", async () => {
	const direct = cycle9SecretPayload("direct");
	const workspacePayload = cycle9SecretPayload("workspace");
	const mutationPayload = cycle9SecretPayload("mutation");
	const capabilitySummary = cycle9SecretPayload("capability-summary");
	const capabilityReference = cycle9SecretPayload("capability-reference");
	const input = policyInput(false, workspacePayload.value);
	input.workspace.editText = async () => ({ changed: true, summary: mutationPayload.value });
	input.workspace.writeText = async () => ({ changed: true, summary: mutationPayload.value });
	input.capabilities = [
		{
			...capability("host_inspect"),
			async execute() {
				return { status: "ok" as const, summary: capabilitySummary.value, references: capabilityReference.value.split("\n") };
			},
		},
		capability("host_verify", { mutates: true }),
	];
	const policy = createToolPolicy(input, { maxToolOutputBytes: 64 * 1024 });
	const read = policy.tools.find((tool) => tool.name === "workspace_read");
	const edit = policy.tools.find((tool) => tool.name === "workspace_edit");
	const inspect = policy.tools.find((tool) => tool.name === "host_inspect");
	assert.ok(read);
	assert.ok(edit);
	assert.ok(inspect);
	const rendered = [
		redactSensitiveText(direct.value),
		text(await read.execute("cycle9-read", { path: ".pi/extensions/shepherd/tool-policy.ts" }, undefined)),
		text(await edit.execute("cycle9-edit", {
			path: ".pi/extensions/shepherd/tool-policy.ts",
			oldText: "old",
			newText: "new",
		}, undefined)),
		text(await inspect.execute("cycle9-inspect", { target: "owned" }, undefined)),
	].join("\n");
	const markers = [
		...direct.markers,
		...workspacePayload.markers,
		...mutationPayload.markers,
		...capabilitySummary.markers,
		...capabilityReference.markers,
	];
	assert.deepEqual(leakedMarkers(rendered, markers), []);
	assert.match(rendered, /\[REDACTED\]/);
});

test("cycle 9 root-scoped reads deny credential stores before invoking the workspace", async () => {
	const input = policyInput(true, "must never be read");
	input.authority.readPrefixes = ["."];
	input.authority.writePrefixes = [];
	const policy = createToolPolicy(input);
	const read = policy.tools.find((tool) => tool.name === "workspace_read");
	assert.ok(read);
	const paths = [
		".npmrc",
		"nested/.NPMRC",
		".yarnrc",
		".yarnrc.yml",
		".pnpmrc",
		".pypirc",
		".netrc",
		"nested/_NETRC",
		".git-credentials",
		".kube/config",
		"nested/.KUBE/config",
		".docker/config.json",
		".config/containers/auth.json",
		".aws/credentials",
		".azure/accessTokens.json",
		".config/gcloud/application_default_credentials.json",
		".config/gh/hosts.yml",
		"pip/pip.conf",
		"nuget.config",
	];
	const accepted: string[] = [];
	for (const path of paths) {
		try {
			await read.execute(`credential-${accepted.length}`, { path }, undefined);
			accepted.push(path);
		} catch (error) {
			assert.ok(error instanceof ToolPolicyError, path);
		}
	}
	assert.deepEqual({ accepted, reads: input.workspace.reads }, { accepted: [], reads: [] });
});

test("cycle 9 capability names deny sensitive noun and acquisition verb permutations for every role", () => {
	const nouns = ["secret", "secrets", "credential", "credentials", "token", "tokens", "password", "passwords", "auth", "api_key"];
	const verbs = ["read", "get", "list", "dump", "export", "acquire", "fetch", "retrieve"];
	const accepted: string[] = [];
	for (const readOnly of [true, false]) {
		for (const noun of nouns) {
			for (const verb of verbs) {
				for (const suffix of [`${noun}_${verb}`, `${verb}_${noun}`]) {
					const name = `host_${suffix}`;
					try {
						createToolPolicy({
							...policyInput(readOnly),
							authority: { ...policyInput(readOnly).authority, capabilityNames: [name] },
							capabilities: [capability(name)],
						});
						accepted.push(`${readOnly ? "read" : "write"}:${name}`);
					} catch (error) {
						assert.ok(error instanceof ToolPolicyError, name);
					}
				}
			}
		}
	}
	assert.deepEqual(accepted, []);
});

test("cycle 9 Pi 0.80.6 validates a real custom-tool call and receives required result details offline", async () => {
	type PiValidationTool = {
		name: string;
		description: string;
		parameters: Readonly<Record<string, unknown>>;
	};
	type PiValidationCall = {
		type: "toolCall";
		id: string;
		name: string;
		arguments: Record<string, unknown>;
	};
	type PiValidationModule = {
		validateToolArguments(tool: PiValidationTool, toolCall: PiValidationCall): Readonly<Record<string, unknown>>;
	};
	const piAiPath = join(
		dirname(process.execPath),
		"..",
		"lib",
		"node_modules",
		"@earendil-works",
		"pi-coding-agent",
		"node_modules",
		"@earendil-works",
		"pi-ai",
		"dist",
		"index.js",
	);
	const piValidation = await import(pathToFileURL(piAiPath).href) as PiValidationModule;
	const input = policyInput(true, "offline result");
	const policy = createToolPolicy(input);
	const read = policy.tools.find((tool) => tool.name === "workspace_read");
	assert.ok(read);
	const validCall: PiValidationCall = {
		type: "toolCall",
		id: "cycle9-offline-valid",
		name: read.name,
		arguments: { path: ".pi/extensions/shepherd/tool-policy.ts", offset: 0, limit: 32 },
	};
	const params = piValidation.validateToolArguments(read, validCall);
	assert.throws(() => piValidation.validateToolArguments(read, {
		...validCall,
		id: "cycle9-offline-invalid",
		arguments: { path: ".pi/extensions/shepherd/tool-policy.ts", forbidden: true },
	}), /invalid|additional|unexpected|properties|argument/i);
	const result = await read.execute(validCall.id, params, undefined);
	assert.deepEqual({
		reads: input.workspace.reads,
		hasDetails: Object.hasOwn(result, "details"),
		details: result.details,
		content: text(result),
	}, {
		reads: [".pi/extensions/shepherd/tool-policy.ts"],
		hasDetails: true,
		details: null,
		content: "offline result",
	});
});

function cycle10SecretPayload(prefix: string): { value: string; markers: string[] } {
	const markers = {
		documentaryEquals: `synthetic-${prefix}-documentary-equals-475`,
		proxyAuthorization: `synthetic-${prefix}-proxy-authorization-475`,
		quotedFlow: `synthetic-${prefix}-quoted-flow-475`,
		oauthFragment: `synthetic-${prefix}-oauth-fragment-475`,
	};
	return {
		value: [
			`token=number of ${markers.documentaryEquals} documentary entries`,
			`Proxy-Authorization: Basic ${markers.proxyAuthorization}`,
			`["client_secret": ${markers.quotedFlow}]`,
			`https://x.invalid/callback#access_token=${markers.oauthFragment}`,
		].join("\n"),
		markers: Object.values(markers),
	};
}

function toolErrorGraphContains(root: unknown, target: unknown): boolean {
	const pending: unknown[] = [root];
	const seen = new Set<unknown>();
	while (pending.length > 0) {
		const current = pending.shift();
		if (current === target) return true;
		if (current === null || current === undefined || seen.has(current)) continue;
		seen.add(current);
		if (current instanceof Error) {
			if (Object.hasOwn(current, "cause")) pending.push(current.cause);
			if (current instanceof AggregateError) pending.push(...current.errors);
		}
	}
	return false;
}

async function toolOutcome(operation: Promise<unknown>): Promise<{ status: "resolved" } | { status: "rejected"; reason: unknown }> {
	return operation.then(
		() => ({ status: "resolved" as const }),
		(reason: unknown) => ({ status: "rejected" as const, reason }),
	);
}

test("cycle 10 schema and result snapshots preserve own prototype-named data fields", async () => {
	const defineData = (target: Record<PropertyKey, unknown>, key: PropertyKey, value: unknown): void => {
		Object.defineProperty(target, key, { configurable: true, enumerable: true, writable: true, value });
	};
	const properties = Object.create(null) as Record<PropertyKey, unknown>;
	defineData(properties, "__proto__", { type: "string" });
	defineData(properties, "prototype", { type: "string" });
	defineData(properties, "constructor", {
		type: "object",
		additionalProperties: false,
		properties: Object.assign(Object.create(null), {
			prototype: { type: "boolean" },
			constructor: { type: "number" },
		}),
	});
	const schema = Object.create(null) as Record<PropertyKey, unknown>;
	defineData(schema, "type", "object");
	defineData(schema, "additionalProperties", false);
	defineData(schema, "properties", properties);
	defineData(schema, "required", ["__proto__", "prototype", "constructor"]);

	const input = policyInput(true);
	input.authority.capabilityNames = ["host_inspect"];
	input.capabilities = [{ ...capability("host_inspect"), parameters: schema as Readonly<Record<string, unknown>> }];
	const policy = createToolPolicy(input);
	const inspect = policy.tools.find((tool) => tool.name === "host_inspect");
	assert.ok(inspect);
	const snapshot = inspect.parameters as Record<string, unknown>;
	const snapshotProperties = snapshot.properties as Record<string, unknown>;

	const inheritedSchema = Object.create(null) as Record<PropertyKey, unknown>;
	defineData(inheritedSchema, "__proto__", {
		type: "object",
		additionalProperties: false,
		properties: {},
	});
	let inheritedAccepted = false;
	try {
		createToolPolicy({
			...policyInput(true),
			authority: { ...policyInput(true).authority, capabilityNames: ["host_inspect"] },
			capabilities: [{ ...capability("host_inspect"), parameters: inheritedSchema as Readonly<Record<string, unknown>> }],
		});
		inheritedAccepted = true;
	} catch (error) {
		assert.ok(error instanceof ToolPolicyError);
	}

	const inheritedWorkspaceResult = Object.create({ changed: true, summary: "inherited mutation" }) as {
		changed: boolean;
		summary: string;
	};
	const inheritedCapabilityResult = Object.create({ status: "ok", summary: "inherited capability", references: [] }) as {
		status: "ok";
		summary: string;
		references: string[];
	};
	const resultInput = policyInput(false);
	resultInput.workspace.editText = async () => inheritedWorkspaceResult;
	resultInput.capabilities = [
		{ ...capability("host_inspect"), async execute() { return inheritedCapabilityResult; } },
		capability("host_verify", { mutates: true }),
	];
	const resultPolicy = createToolPolicy(resultInput);
	const edit = resultPolicy.tools.find((tool) => tool.name === "workspace_edit");
	const resultInspect = resultPolicy.tools.find((tool) => tool.name === "host_inspect");
	assert.ok(edit);
	assert.ok(resultInspect);
	const editOutcome = await toolOutcome(edit.execute("cycle10-prototype-edit", {
		path: ".pi/extensions/shepherd/tool-policy.ts", oldText: "old", newText: "new",
	}, undefined));
	const capabilityOutcome = await toolOutcome(resultInspect.execute("cycle10-prototype-capability", { target: "owned" }, undefined));

	assert.deepEqual({
		ownPrototypeKeys: ["__proto__", "prototype", "constructor"].map((key) => Object.hasOwn(snapshotProperties, key)),
		serializedIdentically: JSON.stringify(snapshot) === JSON.stringify(schema),
		frozen: Object.isFrozen(snapshot) && Object.isFrozen(snapshotProperties),
		inheritedAccepted,
		editStatus: editOutcome.status,
		capabilityStatus: capabilityOutcome.status,
	}, {
		ownPrototypeKeys: [true, true, true],
		serializedIdentically: true,
		frozen: true,
		inheritedAccepted: false,
		editStatus: "rejected",
		capabilityStatus: "rejected",
	});
});

test("cycle 10 schema breadth rejects incrementally before complete own-key materialization", () => {
	const wideProperties: Record<string, unknown> = {};
	for (let index = 0; index < 10_000; index += 1) wideProperties[`field${index}`] = { type: "string" };
	const schema = { type: "object", additionalProperties: false, properties: wideProperties };
	const originalOwnKeys = Reflect.ownKeys;
	let materializedKeys = 0;
	Reflect.ownKeys = ((target: object) => {
		const keys = originalOwnKeys(target);
		if (target === wideProperties) materializedKeys += keys.length;
		return keys;
	}) as typeof Reflect.ownKeys;
	let rejected = false;
	try {
		createToolPolicy({
			...policyInput(true),
			authority: { ...policyInput(true).authority, capabilityNames: ["host_inspect"] },
			capabilities: [{ ...capability("host_inspect"), parameters: schema }],
		});
	} catch (error) {
		rejected = error instanceof ToolPolicyError;
	} finally {
		Reflect.ownKeys = originalOwnKeys;
	}
	assert.deepEqual({ rejected, materializedWithinCeiling: materializedKeys <= 4_097 }, {
		rejected: true,
		materializedWithinCeiling: true,
	});
});

test("cycle 10 workspace and capability failures cross tool boundaries only as sanitized typed errors", async () => {
	const markers = {
		read: "synthetic-cycle10-workspace-read-secret-475",
		edit: "synthetic-cycle10-workspace-edit-secret-475",
		write: "synthetic-cycle10-workspace-write-secret-475",
		capability: "synthetic-cycle10-capability-secret-475",
	};
	const raw = {
		read: new Error(`token=${markers.read}`),
		edit: new Error(`password=${markers.edit}`),
		write: new Error(`client_secret=${markers.write}`),
		capability: new Error(`Proxy-Authorization: Basic ${markers.capability}`),
	};
	const input = policyInput(false);
	input.workspace.readText = async () => { throw raw.read; };
	input.workspace.editText = async () => { throw raw.edit; };
	input.workspace.writeText = async () => { throw raw.write; };
	input.capabilities = [
		{ ...capability("host_inspect"), async execute() { throw raw.capability; } },
		capability("host_verify", { mutates: true }),
	];
	const policy = createToolPolicy(input);
	const tools = new Map(policy.tools.map((tool) => [tool.name, tool]));
	const outcomes = [
		["read", await toolOutcome(tools.get("workspace_read")!.execute("cycle10-error-read", {
			path: ".pi/extensions/shepherd/tool-policy.ts",
		}, undefined))],
		["edit", await toolOutcome(tools.get("workspace_edit")!.execute("cycle10-error-edit", {
			path: ".pi/extensions/shepherd/tool-policy.ts", oldText: "old", newText: "new",
		}, undefined))],
		["write", await toolOutcome(tools.get("workspace_write")!.execute("cycle10-error-write", {
			path: ".pi/extensions/shepherd/tool-policy.ts", content: "new",
		}, undefined))],
		["capability", await toolOutcome(tools.get("host_inspect")!.execute("cycle10-error-capability", {
			target: "owned",
		}, undefined))],
	] as const;
	const problems: string[] = [];
	for (const [name, outcome] of outcomes) {
		const reason = outcome.status === "rejected" ? outcome.reason : undefined;
		const marker = markers[name];
		if (!(reason instanceof ToolPolicyError)) problems.push(`${name}:untyped`);
		if (String(reason).includes(marker)) problems.push(`${name}:marker`);
		if (toolErrorGraphContains(reason, raw[name])) problems.push(`${name}:raw-cause`);
	}
	assert.deepEqual(problems, []);
});

test("cycle 10 redaction closes documentary equals, proxy auth, quoted flow, and OAuth fragments in every tool consumer", async () => {
	const direct = cycle10SecretPayload("direct");
	const workspacePayload = cycle10SecretPayload("workspace");
	const mutationPayload = cycle10SecretPayload("mutation");
	const capabilitySummary = cycle10SecretPayload("capability-summary");
	const capabilityReference = cycle10SecretPayload("capability-reference");
	const input = policyInput(false, workspacePayload.value);
	input.workspace.editText = async () => ({ changed: true, summary: mutationPayload.value });
	input.workspace.writeText = async () => ({ changed: true, summary: mutationPayload.value });
	input.capabilities = [
		{
			...capability("host_inspect"),
			async execute() {
				return { status: "ok" as const, summary: capabilitySummary.value, references: capabilityReference.value.split("\n") };
			},
		},
		capability("host_verify", { mutates: true }),
	];
	const policy = createToolPolicy(input, { maxToolOutputBytes: 64 * 1024 });
	const tools = new Map(policy.tools.map((tool) => [tool.name, tool]));
	const rendered = [
		redactSensitiveText(direct.value),
		text(await tools.get("workspace_read")!.execute("cycle10-redact-read", {
			path: ".pi/extensions/shepherd/tool-policy.ts",
		}, undefined)),
		text(await tools.get("workspace_edit")!.execute("cycle10-redact-edit", {
			path: ".pi/extensions/shepherd/tool-policy.ts", oldText: "old", newText: "new",
		}, undefined)),
		text(await tools.get("workspace_write")!.execute("cycle10-redact-write", {
			path: ".pi/extensions/shepherd/tool-policy.ts", content: "new",
		}, undefined)),
		text(await tools.get("host_inspect")!.execute("cycle10-redact-capability", { target: "owned" }, undefined)),
	].join("\n");
	const markers = [
		...direct.markers,
		...workspacePayload.markers,
		...mutationPayload.markers,
		...capabilitySummary.markers,
		...capabilityReference.markers,
	];
	const harmlessColonProse = "token: number of records processed";
	assert.deepEqual({
		leaks: leakedMarkers(rendered, markers),
		harmlessColonProse: redactSensitiveText(harmlessColonProse),
	}, {
		leaks: [],
		harmlessColonProse,
	});
});

test("cycle 10 root reads deny cloud stores envrc and common private-key names segment-wise", async () => {
	const input = policyInput(true, "opaque credential bytes without assignment syntax");
	input.authority.readPrefixes = ["."];
	input.authority.writePrefixes = [];
	const policy = createToolPolicy(input);
	const read = policy.tools.find((tool) => tool.name === "workspace_read");
	assert.ok(read);
	const paths = [
		".aws/config",
		"nested/.AWS/CONFIG",
		".azure/msal_token_cache.json",
		"nested/.Azure/MSAL_TOKEN_CACHE.JSON",
		".config/gcloud/legacy_credentials/user@example.invalid/adc.json",
		"nested/.CONFIG/GCLOUD/LEGACY_CREDENTIALS/user/ADC.JSON",
		".config/gcloud/access_tokens.db",
		".envrc",
		"nested/.ENVRC",
		".ssh/id_dsa",
		".ssh/id_ecdsa",
		"nested/.SSH/ID_RSA",
		"nested/.ssh/id_ed25519",
	];
	const accepted: string[] = [];
	for (const path of paths) {
		try {
			await read.execute(`cycle10-path-${accepted.length}`, { path }, undefined);
			accepted.push(path);
		} catch (error) {
			assert.ok(error instanceof ToolPolicyError, path);
		}
	}
	assert.deepEqual({ accepted, reads: input.workspace.reads }, { accepted: [], reads: [] });
});

test("cycle 10 capability authority denies sensitive nouns regardless of acquisition synonym or order", () => {
	const names = [
		"host_show_secrets",
		"host_view_credentials",
		"host_download_tokens",
		"host_obtain_password",
		"host_copy_auth",
		"host_reveal_api_key",
		"host_lookup_private_key",
		"host_token_cache_print",
		"host_secret_status",
		"host_credentials_metadata",
		"host_tokens_show",
		"host_password_obtain",
	];
	const accepted: string[] = [];
	for (const readOnly of [true, false]) {
		for (const name of names) {
			try {
				createToolPolicy({
					...policyInput(readOnly),
					authority: { ...policyInput(readOnly).authority, capabilityNames: [name] },
					capabilities: [capability(name)],
				});
				accepted.push(`${readOnly ? "read" : "write"}:${name}`);
			} catch (error) {
				assert.ok(error instanceof ToolPolicyError, name);
			}
		}
	}
	assert.deepEqual(accepted, []);
});

test("cycle 11 capability authority denies concatenated separated plural and mixed forbidden compounds", () => {
	const names = [
		"host_shellcommand",
		"host_shellcommands",
		"host_shell_command",
		"host_execcommand",
		"host_exec_commands",
		"host_spawnagent",
		"host_spawn_agents",
		"host_agentrunner",
		"host_agent_runner",
		"host_recursiveagent",
		"host_recursive_agents",
		"host_httppost",
		"host_http_posts",
		"host_httpwrite",
		"host_http_writes",
		"host_webrequestwrite",
		"host_web_request_writes",
		"host_sqlquery",
		"host_sql_queries",
		"host_sqlwrite",
		"host_sql_writes",
		"host_secretstore_view",
		"host_secretstoresview",
		"host_credentialstore",
		"host_credential_stores",
		"host_accesstoken",
		"host_access_tokens",
		"host_refreshtoken",
		"host_refresh_tokens",
		"host_clientsecret",
		"host_client_secrets",
		"host_apikey",
		"host_api_keys",
		"host_privatekey",
		"host_private_keys",
		"host_tokencache",
		"host_token_caches",
	];
	const accepted: string[] = [];
	for (const readOnly of [true, false]) {
		for (const name of names) {
			try {
				const input = policyInput(readOnly);
				createToolPolicy({
					...input,
					authority: { ...input.authority, capabilityNames: [name] },
					capabilities: [capability(name)],
				});
				accepted.push(`${readOnly ? "read" : "write"}:${name}`);
			} catch (error) {
				assert.ok(error instanceof ToolPolicyError, name);
			}
		}
	}
	assert.deepEqual(accepted, []);
});

test("cycle 11 workspace reads deny AWS SSO and CLI caches before host callbacks", async () => {
	const input = policyInput(true, "opaque credential bytes without assignment syntax");
	input.authority.readPrefixes = ["."];
	input.authority.writePrefixes = [];
	const policy = createToolPolicy(input);
	const read = policy.tools.find((tool) => tool.name === "workspace_read");
	assert.ok(read);
	const paths = [
		".aws/sso/cache/session.json",
		"nested/.aws/sso/cache/session.json",
		"nested/.AWS/SSO/CACHE/SESSION.JSON",
		".aws/cli/cache/session.json",
		"nested/.aws/cli/cache/session.json",
		"nested/.AWS/CLI/CACHE/SESSION.JSON",
	];
	const accepted: string[] = [];
	for (const path of paths) {
		try {
			await read.execute(`cycle11-path-${accepted.length}`, { path }, undefined);
			accepted.push(path);
		} catch (error) {
			assert.ok(error instanceof ToolPolicyError, path);
		}
	}
	assert.deepEqual({ accepted, reads: input.workspace.reads }, { accepted: [], reads: [] });
});

test("cycle 12 every hostile tool input fails as a typed bounded redacted DTO error", async () => {
	const input = policyInput(false);
	const policy = createToolPolicy(input);
	const tools = new Map(policy.tools.map((tool) => [tool.name, tool]));
	const markers = {
		proxy: "synthetic-cycle12-tool-proxy-475",
		accessor: "synthetic-cycle12-tool-accessor-475",
		signal: "synthetic-cycle12-tool-signal-475",
		toJSON: "synthetic-cycle12-tool-tojson-475",
		cycle: "synthetic-cycle12-tool-cycle-475",
	};
	let proxyTraps = 0;
	const proxy = new Proxy({ path: ".pi/extensions/shepherd/tool-policy.ts" }, {
		ownKeys() {
			proxyTraps += 1;
			throw new Error(`token=${markers.proxy}`);
		},
	});
	const accessor: Record<string, unknown> = { oldText: "old", newText: "new" };
	Object.defineProperty(accessor, "path", {
		enumerable: true,
		get() { throw new Error(`client_secret=${markers.accessor}`); },
	});
	const signalController = new AbortController();
	Object.defineProperty(signalController.signal, "aborted", {
		configurable: true,
		get() { throw new Error(`Cookie: session=${markers.signal}`); },
	});
	const toJSON = {
		target: "owned",
		toJSON() { throw new Error(`Set-Cookie: auth=${markers.toJSON}`); },
	};
	const cyclic: Record<string, unknown> = { target: "owned" };
	cyclic.self = cyclic;
	const cases = [
		["proxy", tools.get("workspace_read")!, proxy, undefined],
		["accessor", tools.get("workspace_edit")!, accessor, undefined],
		["signal", tools.get("workspace_write")!, {
			path: ".pi/extensions/shepherd/tool-policy.ts", content: "bounded",
		}, signalController.signal],
		["toJSON", tools.get("host_inspect")!, toJSON, undefined],
		["cycle", tools.get("host_verify")!, cyclic, undefined],
	] as const;
	const problems: string[] = [];
	for (const [name, tool, params, signal] of cases) {
		const outcome = await toolOutcome(tool.execute(`cycle12-tool-input-${name}`, params, signal));
		const reason = outcome.status === "rejected" ? outcome.reason : undefined;
		if (!(reason instanceof ToolPolicyError)) problems.push(`${name}:${outcome.status}:untyped`);
		const messages: string[] = [];
		let current: unknown = reason;
		for (let depth = 0; depth < 4 && current instanceof Error; depth += 1) {
			messages.push(current.message);
			current = Object.hasOwn(current, "cause") ? current.cause : undefined;
		}
		if (messages.some((message) => message.includes(markers[name]))) problems.push(`${name}:marker`);
	}
	if (proxyTraps > 0) problems.push(`proxy:traps-${proxyTraps}`);
	assert.deepEqual(problems, []);
});

test("cycle 13 public tool-policy arrays are intrinsic dense snapshots with no caller behavior", async () => {
	let callerBehaviorCalls = 0;
	const poison = <T>(values: T[]): T[] => {
		Object.defineProperties(values, {
			[Symbol.iterator]: {
				configurable: true,
				value() {
					callerBehaviorCalls += 1;
					return Array.prototype[Symbol.iterator].call(this);
				},
			},
			some: {
				configurable: true,
				value(callback: (...args: unknown[]) => unknown) {
					callerBehaviorCalls += 1;
					return Array.prototype.some.call(this, callback);
				},
			},
			map: {
				configurable: true,
				value(callback: (...args: unknown[]) => unknown) {
					callerBehaviorCalls += 1;
					return Array.prototype.map.call(this, callback);
				},
			},
			join: {
				configurable: true,
				value(separator?: string) {
					callerBehaviorCalls += 1;
					return Array.prototype.join.call(this, separator);
				},
			},
			constructor: {
				configurable: true,
				get() {
					callerBehaviorCalls += 1;
					return Array;
				},
			},
		});
		return values;
	};

	const directPrefixes = poison(["direct/owned"]);
	const normalized = normalizeScopedPrefixes(directPrefixes, "cycle 13 direct");
	const input = policyInput(false);
	const readPrefixes = poison([".pi/extensions/shepherd"]);
	const writePrefixes = poison([".pi/extensions/shepherd"]);
	const capabilityNames = poison(["host_inspect", "host_verify"]);
	const capabilities = poison([capability("host_inspect"), capability("host_verify")]);
	input.authority.readPrefixes = readPrefixes;
	input.authority.writePrefixes = writePrefixes;
	input.authority.capabilityNames = capabilityNames;
	input.capabilities = capabilities;
	const policy = createToolPolicy(input);
	const beforeMutation = [...policy.names];
	readPrefixes[0] = "outside";
	writePrefixes[0] = "outside";
	capabilityNames[0] = "host_process_run";
	capabilities[0] = capability("host_process_run");

	const read = policy.tools.find((tool) => tool.name === "workspace_read");
	assert.ok(read);
	const outside = await Promise.resolve(read.execute("cycle13-policy-outside", {
		path: "outside/file.txt",
	}, undefined)).then(
		() => "resolved",
		(error: unknown) => error instanceof ToolPolicyError ? "typed-rejected" : "untyped-rejected",
	);
	assert.throws(
		() => normalizeScopedPrefixes(Array.from({ length: 65 }, (_value, index) => `owned/${index}`), "cycle 13 max"),
		ToolPolicyError,
	);
	assert.deepEqual({
		callerBehaviorCalls,
		normalized,
		beforeMutation,
		afterMutation: policy.names,
		outside,
	}, {
		callerBehaviorCalls: 0,
		normalized: ["direct/owned"],
		beforeMutation: ["workspace_read", "workspace_edit", "workspace_write", "host_inspect", "host_verify"],
		afterMutation: ["workspace_read", "workspace_edit", "workspace_write", "host_inspect", "host_verify"],
		outside: "typed-rejected",
	});
});

test("cycle 13 capability names structurally deny equivalent generic and protected-data authority", () => {
	const forbidden = [
		"host_write_http",
		"host_query_sql",
		"host_create_agent",
		"host_process_run",
		"host_run_process",
		"host_web_send",
		"host_send_web",
		"host_database_modify",
		"host_modify_database",
		"host_vault_view",
		"host_view_vault",
		"host_environment_export",
		"host_export_environment",
		"host_keychain_dump",
		"host_dump_keychain",
	] as const;
	const accepted: string[] = [];
	for (const readOnly of [true, false]) {
		for (const name of forbidden) {
			try {
				const input = policyInput(readOnly);
				createToolPolicy({
					...input,
					authority: { ...input.authority, capabilityNames: [name] },
					capabilities: [capability(name)],
				});
				accepted.push(`${readOnly ? "read" : "write"}:${name}`);
			} catch (error) {
				assert.ok(error instanceof ToolPolicyError, name);
			}
		}
	}
	const safeInput = policyInput(false);
	const safePolicy = createToolPolicy(safeInput);
	assert.deepEqual({
		accepted,
		safeHostTools: safePolicy.names.filter((name) => name.startsWith("host_")),
	}, {
		accepted: [],
		safeHostTools: ["host_inspect", "host_verify"],
	});
});
