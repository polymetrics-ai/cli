import { execFile } from "node:child_process";
import { mkdir, mkdtemp, rm, writeFile } from "node:fs/promises";
import { tmpdir } from "node:os";
import { dirname, join } from "node:path";
import { promisify } from "node:util";

const execFileAsync = promisify(execFile);

export interface LocalGitFixture {
	root: string;
	remote: string;
	coordinator: string;
	worktreeRoot: string;
	parentBranch: string;
	parentHead: string;
	cleanup(): Promise<void>;
}

export async function git(cwd: string, ...args: string[]): Promise<string> {
	const { stdout } = await execFileAsync("git", args, {
		cwd,
		encoding: "utf8",
		maxBuffer: 1024 * 1024,
		env: {
			...process.env,
			GIT_CONFIG_NOSYSTEM: "1",
			GIT_CONFIG_GLOBAL: process.platform === "win32" ? "NUL" : "/dev/null",
			GIT_TERMINAL_PROMPT: "0",
		},
	});
	return stdout;
}

export async function write(root: string, relativePath: string, contents: string): Promise<void> {
	const path = join(root, relativePath);
	await mkdir(dirname(path), { recursive: true });
	await writeFile(path, contents);
}

export async function createLocalGitFixture(): Promise<LocalGitFixture> {
	const root = await mkdtemp(join(tmpdir(), "pm-shepherd-476-"));
	const remote = join(root, "origin.git");
	const coordinator = join(root, "coordinator");
	const worktreeRoot = join(root, "workers");
	const parentBranch = "feat/471-pi-agent-session-shepherd";
	await mkdir(worktreeRoot, { recursive: true, mode: 0o700 });
	await git(root, "init", "--bare", "--initial-branch=main", remote);
	await git(root, "clone", remote, coordinator);
	await git(coordinator, "config", "user.name", "Shepherd Test");
	await git(coordinator, "config", "user.email", "shepherd@example.invalid");
	await write(coordinator, "README.md", "seed\n");
	await git(coordinator, "add", "--", "README.md");
	await git(coordinator, "commit", "-m", "test: seed repository");
	await git(coordinator, "push", "origin", "main");
	await git(coordinator, "remote", "set-head", "origin", "main");
	await git(coordinator, "switch", "-c", parentBranch);
	await write(coordinator, "parent.txt", "parent\n");
	await git(coordinator, "add", "--", "parent.txt");
	await git(coordinator, "commit", "-m", "test: seed parent branch");
	await git(coordinator, "push", "-u", "origin", parentBranch);
	const parentHead = (await git(coordinator, "rev-parse", "HEAD")).trim();
	return {
		root,
		remote,
		coordinator,
		worktreeRoot,
		parentBranch,
		parentHead,
		cleanup: () => rm(root, { recursive: true, force: true }),
	};
}
