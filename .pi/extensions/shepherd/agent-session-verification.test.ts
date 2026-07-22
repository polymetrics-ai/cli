import assert from "node:assert/strict";
import { constants } from "node:fs";
import { access, mkdir, mkdtemp, realpath, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { delimiter, join } from "node:path";
import test from "node:test";

import type { AgentSessionHandoff, RoleRunRequest } from "./agent-session-runtime.ts";
import {
	AgentSessionVerificationRunner,
} from "./agent-session-verification.ts";
import {
	BoundedVerificationRunner,
	type ProductionVerificationResult,
} from "./bounded-verification.ts";
import type { ProductionVerificationCommand } from "./autonomous-production-contract.ts";

const HEAD = "c".repeat(40);

function verificationBinding() {
	return {
		issue: 479,
		branch: "feat/479-verification",
		runId: "run-479",
		generation: 1,
		laneId: "lane-a-verification",
		candidateHead: HEAD,
	};
}

function command(id: string, executable = "node", args = ["--test"]): ProductionVerificationCommand {
	return {
		id,
		executable,
		args,
		cwd: ".",
		timeoutMs: 30_000,
		maxOutputBytes: 65_536,
	};
}

function result(id: string, status: "passed" | "failed" = "passed"): ProductionVerificationResult {
	return {
		id,
		status,
		exitCode: status === "passed" ? 0 : 1,
		signal: null,
		stdout: "",
		stderr: "",
		durationMs: 1,
		...(status === "passed" ? {} : { failureKind: "exit" as const }),
	};
}

function handoff(request: RoleRunRequest): AgentSessionHandoff {
	return {
		schemaVersion: 1,
		...request.binding,
		role: "verification",
		status: "completed",
		summary: "requested every immutable verification ID",
		observedMutation: true,
		changedPaths: [],
		verification: [],
		findings: [],
	};
}

test("implementation and correction agents can rerun only declared test IDs and receive bounded diagnostics", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-agent-verify-interactive-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const planned = command("focused", "go", ["test", "./..."]);
	let passing = false;
	const executions: ProductionVerificationCommand[] = [];
	const runner = new AgentSessionVerificationRunner({
		agentSession: { async run() { throw new Error("not used"); }, async abort() {} },
		executor: {
			async run(worktreeRoot, value) {
				assert.equal(worktreeRoot, await realpath(root));
				executions.push(structuredClone(value));
				return passing
					? result(value.id)
					: { ...result(value.id, "failed"), stderr: "expected 4, received 5" };
			},
		},
	});
	const capability = await runner.createAgentCapability(root, [planned], new AbortController().signal);
	const red = await capability.execute({ id: "focused" });
	assert.equal(red.status, "failed");
	assert.match(red.summary, /expected 4, received 5/);
	passing = true;
	assert.equal((await capability.execute({ id: "focused" })).status, "ok");
	assert.equal((await capability.execute({ id: "undeclared" })).status, "blocked");
	assert.equal((await capability.execute({ id: "focused", args: ["--invented"] })).status, "blocked");
	assert.deepEqual(executions, [planned, planned]);
});

test("verification AgentSession can select only immutable command IDs and host runs exact tuples in order", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-agent-verify-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const commands = [command("first"), command("second")];
	const executed: ProductionVerificationCommand[] = [];
	const runner = new AgentSessionVerificationRunner({
		agentSession: {
			async run(request) {
				assert.equal(request.role, "verification");
				assert.equal(request.authority.readOnly, false);
				assert.equal(request.workspaceMutation, false);
				assert.deepEqual(request.authority.capabilityNames, ["host_verify"]);
				assert.equal(request.capabilities.length, 1);
				assert.deepEqual(Object.keys(request.capabilities[0]!.parameters), [
					"type", "additionalProperties", "required", "properties",
				]);
				assert.equal((await request.capabilities[0]!.execute({ id: "first" }, request.signal)).status, "ok");
				assert.equal((await request.capabilities[0]!.execute({ id: "second" }, request.signal)).status, "ok");
				return handoff(request);
			},
			async abort() {},
		},
		executor: {
			async run(worktreeRoot, planned) {
				assert.equal(worktreeRoot, await realpath(root));
				executed.push(structuredClone(planned));
				return result(planned.id);
			},
		},
	});
	const results = await runner.runAll(root, commands, new AbortController().signal, verificationBinding());
	assert.deepEqual(executed, commands);
	assert.deepEqual(results.map((entry) => entry.id), ["first", "second"]);
});

test("out-of-order or omitted verification IDs fail closed instead of trusting agent prose", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-agent-verify-omit-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	let executions = 0;
	const runner = new AgentSessionVerificationRunner({
		agentSession: {
			async run(request) {
				const response = await request.capabilities[0]!.execute({ id: "second" }, request.signal);
				assert.equal(response.status, "blocked");
				return handoff(request);
			},
			async abort() {},
		},
		executor: {
			async run(_root, planned) { executions += 1; return result(planned.id); },
		},
	});
	await assert.rejects(
		runner.runAll(root, [command("first"), command("second")], new AbortController().signal, verificationBinding()),
		/did not run every required command|out of order/i,
	);
	assert.equal(executions, 0);
});

test("verification AgentSession returns authoritative host failure for the correction loop", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-agent-verify-fail-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const runner = new AgentSessionVerificationRunner({
		agentSession: {
			async run(request) {
				const response = await request.capabilities[0]!.execute({ id: "tests" }, request.signal);
				assert.equal(response.status, "failed");
				assert.match(response.summary, /compile failed/i);
				return handoff(request);
			},
			async abort() {},
		},
		executor: {
			async run(_root, planned) {
				return { ...result(planned.id, "failed"), stderr: "compile failed at fixture.go:9" };
			},
		},
	});
	const results = await runner.runAll(root, [command("tests")], new AbortController().signal, verificationBinding());
	assert.equal(results[0]?.status, "failed");
	assert.equal(results[0]?.failureKind, "exit");
});

test("a runtime or cleanup failure after a failed test never bypasses exact handoff validation", async (t) => {
	const root = await mkdtemp(join(tmpdir(), "shepherd-agent-verify-runtime-fail-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	const runner = new AgentSessionVerificationRunner({
		agentSession: {
			async run(request) {
				assert.equal((await request.capabilities[0]!.execute({ id: "tests" }, request.signal)).status, "failed");
				throw new Error("AgentSession cleanup failed");
			},
			async abort() {},
		},
		executor: { async run(_root, planned) { return result(planned.id, "failed"); } },
	});
	await assert.rejects(
		runner.runAll(root, [command("tests")], new AbortController().signal, verificationBinding()),
		/AgentSession cleanup failed/,
	);
});

test("real Go fixture flows through verification AgentSession to bounded host_verify execution", async (t) => {
	const go = await executableOnPath("go");
	const root = await mkdtemp(join(tmpdir(), "shepherd-agent-verify-go-"));
	t.after(() => rm(root, { recursive: true, force: true }));
	await mkdir(join(root, "home"));
	await writeFile(join(root, "go.mod"), "module example.com/shepherdfixture\n\ngo 1.25\n");
	await writeFile(join(root, "sum_test.go"), `package fixture

import "testing"

func TestSum(t *testing.T) {
	if 2 + 3 != 5 { t.Fatal("bad sum") }
}
`);
	const runner = new AgentSessionVerificationRunner({
		agentSession: {
			async run(request) {
				assert.equal((await request.capabilities[0]!.execute({ id: "go-tests" }, request.signal)).status, "ok");
				return handoff(request);
			},
			async abort() {},
		},
		executor: new BoundedVerificationRunner({
			executables: { go },
			environment: {
				HOME: join(root, "home"),
				GOCACHE: join(root, "go-cache"),
				GOPATH: join(root, "go-path"),
			},
		}),
	});
	const results = await runner.runAll(
		root,
		[command("go-tests", "go", ["test", "./..."])],
		new AbortController().signal,
		verificationBinding(),
	);
	assert.equal(results[0]?.status, "passed", results[0]?.stderr);
});

async function executableOnPath(name: string): Promise<string> {
	for (const directory of (process.env.PATH ?? "").split(delimiter)) {
		if (directory.length === 0) continue;
		const candidate = join(directory, name);
		try {
			await access(candidate, constants.X_OK);
			return realpath(candidate);
		} catch { /* Continue to the next trusted PATH entry. */ }
	}
	throw new Error(`${name} is unavailable for the real verification fixture`);
}
