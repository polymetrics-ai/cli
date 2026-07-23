import assert from "node:assert/strict";
import { mkdtemp, rm } from "node:fs/promises";
import { tmpdir } from "node:os";
import { resolve } from "node:path";
import { spawn } from "node:child_process";

const pi = process.env.SHEPHERD_PI_BIN ?? "pi";
const root = process.cwd();
const shepherd = resolve(root, ".pi/extensions/shepherd/index.ts");
const workflowEngine = resolve(
	root,
	".pi/npm/node_modules/pi-workflow-engine/.pi/extensions/pi-workflow-engine/index.ts",
);
const workflowCommands = [
	"workflow",
	"workflow:dynamax",
	"workflow:inspector",
	"workflow:models",
	"workflow:results",
	"workflow:runs",
];

async function commandNames(label, extensions) {
	const agentDir = await mkdtemp(resolve(tmpdir(), `pm-shepherd-${label}-`));
	try {
		const args = [
			"--mode", "rpc",
			"--no-session",
			"--approve",
			"--no-extensions",
			"--no-skills",
			"--no-prompt-templates",
			"--no-context-files",
		];
		for (const extension of extensions) args.push("-e", extension);
		const child = spawn(pi, args, {
			cwd: root,
			env: { ...process.env, PI_OFFLINE: "1", PI_CODING_AGENT_DIR: agentDir },
			stdio: ["pipe", "pipe", "pipe"],
		});
		let stdout = "";
		let stderr = "";
		const maximum = 1024 * 1024;
		child.stdout.setEncoding("utf8");
		child.stderr.setEncoding("utf8");
		child.stdout.on("data", (chunk) => {
			stdout += chunk;
			if (stdout.length > maximum) child.kill("SIGKILL");
		});
		child.stderr.on("data", (chunk) => {
			stderr += chunk;
			if (stderr.length > maximum) child.kill("SIGKILL");
		});
		child.stdin.end('{"id":"commands","type":"get_commands"}\n');
		const exit = await new Promise((settle, reject) => {
			const timer = setTimeout(() => {
				child.kill("SIGKILL");
				reject(new Error(`${label} RPC timed out`));
			}, 30_000);
			child.once("error", (error) => {
				clearTimeout(timer);
				reject(error);
			});
			child.once("close", (code, signal) => {
				clearTimeout(timer);
				settle({ code, signal });
			});
		});
		assert.deepEqual(exit, { code: 0, signal: null }, `${label} RPC failed: ${stderr.slice(0, 2_000)}`);
		assert.ok(stdout.length <= maximum, `${label} RPC output exceeded its bound`);
		const messages = stdout.trim().split("\n").filter(Boolean).map((line) => JSON.parse(line));
		const response = messages.find((message) => message.id === "commands");
		assert.equal(response?.success, true, `${label} RPC did not return a successful command response`);
		return response.data.commands.map((command) => command.name);
	} finally {
		await rm(agentDir, { recursive: true, force: true });
	}
}

const isolatedShepherd = await commandNames("isolated-shepherd", [shepherd]);
assert.ok(isolatedShepherd.includes("pm-shepherd"));
assert.ok(!isolatedShepherd.includes("workflow"));

const isolatedWorkflow = await commandNames("isolated-workflow", [workflowEngine]);
for (const command of workflowCommands) assert.ok(isolatedWorkflow.includes(command), `missing isolated ${command}`);
assert.ok(!isolatedWorkflow.includes("pm-shepherd"));

const coLoaded = await commandNames("co-load", [workflowEngine, shepherd]);
assert.equal(coLoaded.filter((name) => name === "pm-shepherd").length, 1, "co-load duplicated pm-shepherd");
for (const command of workflowCommands) {
	assert.equal(coLoaded.filter((name) => name === command).length, 1, `co-load duplicated or omitted ${command}`);
}

console.log("verified offline isolated and co-loaded Shepherd/workflow-engine RPC command surfaces");
