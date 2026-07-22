import assert from "node:assert/strict";
import { chmod, mkdir, mkdtemp, realpath, symlink, writeFile } from "node:fs/promises";
import { join } from "node:path";
import { tmpdir } from "node:os";
import test from "node:test";

import { BoundedVerificationRunner } from "./bounded-verification.ts";
import type { ProductionVerificationCommand } from "./autonomous-production-contract.ts";

function command(overrides: Partial<ProductionVerificationCommand> = {}): ProductionVerificationCommand {
	return {
		id: "focused",
		executable: "node",
		args: ["-e", "process.stdout.write('ok')"],
		cwd: ".",
		timeoutMs: 2_000,
		maxOutputBytes: 64 * 1024,
		...overrides,
	};
}

async function repositoryFixture(t: test.TestContext): Promise<string> {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-verification-"));
	t.after(() => import("node:fs/promises").then(({ rm }) => rm(root, { recursive: true, force: true })));
	await mkdir(join(root, "nested"));
	return root;
}

test("runs only an allowlisted executable with argv, a canonical repository-relative cwd, and sanitized env", async (t) => {
	const root = await repositoryFixture(t);
	const runner = new BoundedVerificationRunner({
		executables: { node: process.execPath },
		environment: { SHEPHERD_VISIBLE: "yes" },
	});
	process.env.SHEPHERD_SECRET_SHOULD_NOT_LEAK = "secret";
	t.after(() => delete process.env.SHEPHERD_SECRET_SHOULD_NOT_LEAK);
	const script = [
		"const path=require('node:path')",
		"const root=process.argv[1]",
		"process.stdout.write(JSON.stringify({relative:path.relative(root,process.cwd()),visible:process.env.SHEPHERD_VISIBLE,secret:process.env.SHEPHERD_SECRET_SHOULD_NOT_LEAK}))",
	].join(";");
	const result = await runner.run(root, command({ args: ["-e", script, await realpath(root)], cwd: "nested" }));
	assert.equal(result.status, "passed");
	assert.equal(result.exitCode, 0);
	assert.deepEqual(JSON.parse(result.stdout), { relative: "nested", visible: "yes" });
	assert.equal(result.stderr, "");
});

test("rejects unallowlisted executables, absolute/traversing cwd, symlink cwd, and executable symlinks", async (t) => {
	const root = await repositoryFixture(t);
	const outside = await mkdtemp(join(tmpdir(), "pm-shepherd-verification-outside-"));
	t.after(() => import("node:fs/promises").then(({ rm }) => rm(outside, { recursive: true, force: true })));
	await symlink(outside, join(root, "linked"), "dir");
	const executableLink = join(root, "node-link");
	await symlink(process.execPath, executableLink);

	const runner = new BoundedVerificationRunner({ executables: { node: process.execPath } });
	await assert.rejects(runner.run(root, command({ executable: "sh" })), /allowlist/i);
	for (const cwd of [outside, "../outside", "linked"]) {
		await assert.rejects(runner.run(root, command({ cwd })), /cwd|worktree|symlink|relative/i);
	}
	assert.throws(
		() => new BoundedVerificationRunner({ executables: { node: executableLink } }),
		/absolute|symlink|executable/i,
	);
});

test("bounds combined stdout and stderr and returns no unbounded process output", async (t) => {
	const root = await repositoryFixture(t);
	const runner = new BoundedVerificationRunner({ executables: { node: process.execPath } });
	const result = await runner.run(root, command({
		args: ["-e", "process.stdout.write('a'.repeat(900));process.stderr.write('b'.repeat(900))"],
		maxOutputBytes: 1_024,
	}));
	assert.equal(result.status, "failed");
	assert.equal(result.failureKind, "output_limit");
	assert.ok(Buffer.byteLength(result.stdout) + Buffer.byteLength(result.stderr) <= 1_024);
});

test("terminates timed-out and aborted processes and preserves the authoritative failure kind", async (t) => {
	const root = await repositoryFixture(t);
	const runner = new BoundedVerificationRunner({ executables: { node: process.execPath }, terminationGraceMs: 20 });
	const waitScript = "setInterval(()=>{},1000)";
	const timedOut = await runner.run(root, command({ args: ["-e", waitScript], timeoutMs: 30 }));
	assert.equal(timedOut.status, "failed");
	assert.equal(timedOut.failureKind, "timeout");

	const controller = new AbortController();
	const running = runner.run(root, command({ args: ["-e", waitScript] }), controller.signal);
	controller.abort();
	const aborted = await running;
	assert.equal(aborted.status, "failed");
	assert.equal(aborted.failureKind, "aborted");
});

test("runAll is ordered, fail-fast, and never launches work after a failed command", async (t) => {
	const root = await repositoryFixture(t);
	const marker = join(root, "should-not-run");
	const runner = new BoundedVerificationRunner({ executables: { node: process.execPath } });
	const results = await runner.runAll(root, [
		command({ id: "one", args: ["-e", "process.stdout.write('one')"] }),
		command({ id: "two", args: ["-e", "process.exit(7)"] }),
		command({ id: "three", args: ["-e", `require('node:fs').writeFileSync(${JSON.stringify(marker)},'ran')`] }),
	]);
	assert.deepEqual(results.map((result) => [result.id, result.status]), [["one", "passed"], ["two", "failed"]]);
	await assert.rejects(import("node:fs/promises").then(({ access }) => access(marker)));
});

test("rejects a non-executable allowlist target before starting a child", async (t) => {
	const root = await repositoryFixture(t);
	const file = join(root, "not-executable");
	await writeFile(file, "plain text\n");
	await chmod(file, 0o600);
	assert.throws(
		() => new BoundedVerificationRunner({ executables: { node: file } }),
		/executable/i,
	);
});
