import { createHash } from "node:crypto";
import { execFile } from "node:child_process";
import { lstat, realpath, stat } from "node:fs/promises";
import { isAbsolute, relative, resolve, sep } from "node:path";

const DEFAULT_TIMEOUT_MS = 15_000;
const DEFAULT_MAX_OUTPUT_BYTES = 1024 * 1024;
const MAX_PATH_BYTES = 4_096;
const MAX_BRANCH_BYTES = 240;
const MAX_SCOPES = 64;
const SHA_PATTERN = /^[0-9a-f]{40}$/;
const IDENTITY_PATTERN = /^[0-9a-f]{64}$/;
const SLUG_PATTERN = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;

export interface GitCommandRequest {
	cwd: string;
	args: readonly string[];
	env: NodeJS.ProcessEnv;
	timeoutMs: number;
	maxOutputBytes: number;
}

export type GitCommandExecutor = (request: GitCommandRequest) => Promise<Buffer>;

export interface GitAdapterOptions {
	execute?: GitCommandExecutor;
	timeoutMs?: number;
	maxOutputBytes?: number;
}

export interface GitBinding {
	cwd: string;
	repositoryIdentity: string;
	worktreeIdentity: string;
	remoteName: "origin";
	remoteIdentity: string;
}

export interface GitStatusEntry {
	code: string;
	path: string;
	originalPath?: string;
}

export interface GitStatusEvidence {
	clean: boolean;
	entries: GitStatusEntry[];
}

export interface GitBranchEvidence {
	branch: string;
	head: string;
}

export interface GitWorktreeEvidence {
	cwd: string;
	head?: string;
	branch?: string;
	detached: boolean;
	bare: boolean;
	locked: boolean;
	prunable: boolean;
}

export interface GitDiffEvidence {
	baseHead: string;
	head: string;
	changedScope: string[];
}

export interface GitCommitRequest {
	issue: number;
	slug: string;
	branch: string;
	expectedHead: string;
	message: string;
	scopes: readonly string[];
}

export interface GitCommitEvidence {
	committed: boolean;
	previousHead: string;
	head: string;
}

export interface GitPushRequest {
	issue: number;
	slug: string;
	branch: string;
	expectedHead: string;
	defaultBranch: string;
}

export interface GitPushEvidence {
	branch: string;
	head: string;
	remoteName: "origin";
}

export interface GitAddWorktreeRequest {
	trustedRoot: string;
	path: string;
	issue: number;
	slug: string;
	branch: string;
	baseHead: string;
}

export class GitAdapterError extends Error {
	constructor(message: string, options?: ErrorOptions) {
		super(message, options);
		this.name = "GitAdapterError";
	}
}

class GitCommandFailure extends Error {
	readonly exitCode: number | undefined;

	constructor(exitCode?: number) {
		super("typed Git command failed");
		this.name = "GitCommandFailure";
		this.exitCode = exitCode;
	}
}

function safeText(value: unknown, maximum: number): value is string {
	return typeof value === "string"
		&& value.length > 0
		&& Buffer.byteLength(value) <= maximum
		&& !/[\u0000-\u001f\u007f-\u009f]/.test(value);
}

function validIssue(issue: unknown): issue is number {
	return Number.isSafeInteger(issue) && (issue as number) > 0 && (issue as number) <= 2_147_483_647;
}

function assertSlug(slug: unknown): asserts slug is string {
	if (!safeText(slug, 100) || !SLUG_PATTERN.test(slug) || ["main", "master", "trunk", "head"].includes(slug)) {
		throw new GitAdapterError("issue slug must be canonical lowercase kebab-case");
	}
}

export function canonicalIssueBranch(issue: number, slug: string): string {
	if (!validIssue(issue)) throw new GitAdapterError("issue must be a positive bounded integer");
	assertSlug(slug);
	const branch = `feat/${issue}-${slug}`;
	if (Buffer.byteLength(branch) > MAX_BRANCH_BYTES) throw new GitAdapterError("canonical issue branch is too long");
	return branch;
}

export function canonicalIssueWorktreeName(issue: number, slug: string): string {
	canonicalIssueBranch(issue, slug);
	return `issue-${issue}-${slug}`;
}

function assertCanonicalIssueBranch(issue: number, slug: string, branch: unknown): asserts branch is string {
	const canonical = canonicalIssueBranch(issue, slug);
	if (branch !== canonical) throw new GitAdapterError(`branch must equal canonical issue branch ${canonical}`);
}

function assertSafeBranch(branch: unknown, description = "branch"): asserts branch is string {
	if (!safeText(branch, MAX_BRANCH_BYTES)
		|| branch.startsWith("-")
		|| branch.endsWith(".")
		|| branch.includes("..")
		|| branch.includes("@{")
		|| /[ ~^:?*\\[\\]\\\\]/.test(branch)
		|| branch.includes("//")) {
		throw new GitAdapterError(`${description} is not a safe branch name`);
	}
}

function assertSha(value: unknown, description: string): asserts value is string {
	if (typeof value !== "string" || !SHA_PATTERN.test(value)) {
		throw new GitAdapterError(`${description} must be an exact lowercase commit SHA`);
	}
}

function assertIdentity(value: unknown, description: string): asserts value is string {
	if (typeof value !== "string" || !IDENTITY_PATTERN.test(value)) {
		throw new GitAdapterError(`${description} must be a canonical identity hash`);
	}
}

function validateScope(scope: unknown): asserts scope is string {
	if (!safeText(scope, MAX_PATH_BYTES)
		|| isAbsolute(scope)
		|| scope === "."
		|| scope.startsWith("-")
		|| scope.split(/[\\/]/).some((part) => part === "" || part === "." || part === "..")
		|| scope === ".git"
		|| scope.startsWith(".git/")
		|| scope.startsWith(".git\\")) {
		throw new GitAdapterError("commit or diff scope must be a bounded repository-relative path");
	}
}

function validateScopes(scopes: readonly string[]): string[] {
	if (!Array.isArray(scopes) || scopes.length === 0 || scopes.length > MAX_SCOPES) {
		throw new GitAdapterError("commit or diff scopes must contain one to 64 paths");
	}
	const unique = new Set<string>();
	for (const scope of scopes) {
		validateScope(scope);
		unique.add(scope.replaceAll("\\", "/"));
	}
	return [...unique].sort();
}

export function canonicalGitScopes(scopes: readonly string[]): string[] {
	return validateScopes(scopes);
}

function pathWithinScope(path: string, scopes: readonly string[]): boolean {
	return scopes.some((scope) => path === scope || path.startsWith(`${scope}/`));
}

function sanitizedGitEnvironment(): NodeJS.ProcessEnv {
	const env = { ...process.env };
	for (const key of Object.keys(env)) {
		if (key === "GIT_DIR"
			|| key === "GIT_WORK_TREE"
			|| key === "GIT_COMMON_DIR"
			|| key === "GIT_INDEX_FILE"
			|| key === "GIT_OBJECT_DIRECTORY"
			|| key === "GIT_ALTERNATE_OBJECT_DIRECTORIES"
			|| key === "GIT_CEILING_DIRECTORIES"
			|| key === "GIT_CONFIG_COUNT"
			|| key.startsWith("GIT_CONFIG_KEY_")
			|| key.startsWith("GIT_CONFIG_VALUE_")) delete env[key];
	}
	env.GIT_CONFIG_NOSYSTEM = "1";
	env.GIT_CONFIG_GLOBAL = process.platform === "win32" ? "NUL" : "/dev/null";
	env.GIT_TERMINAL_PROMPT = "0";
	env.GIT_OPTIONAL_LOCKS = "0";
	env.LC_ALL = "C";
	return env;
}

const defaultExecutor: GitCommandExecutor = (request) => new Promise((resolvePromise, reject) => {
	execFile("git", request.args, {
		cwd: request.cwd,
		encoding: "buffer",
		env: request.env,
		maxBuffer: request.maxOutputBytes,
		timeout: request.timeoutMs,
		killSignal: "SIGTERM",
	}, (error, stdout) => error ? reject(error) : resolvePromise(stdout));
});

function hashIdentity(parts: readonly string[]): string {
	const hash = createHash("sha256");
	for (const part of parts) {
		hash.update(String(Buffer.byteLength(part)));
		hash.update(":");
		hash.update(part);
		hash.update(";");
	}
	return hash.digest("hex");
}

function metadataIdentity(metadata: Awaited<ReturnType<typeof stat>>): string {
	return `${metadata.dev}:${metadata.ino}:${metadata.birthtimeMs}`;
}

function stripLineEnding(value: string): string {
	return value.endsWith("\r\n") ? value.slice(0, -2) : value.endsWith("\n") ? value.slice(0, -1) : value;
}

async function normalizeRemote(rawOutput: string, repositoryRoot: string): Promise<string> {
	const raw = stripLineEnding(rawOutput);
	if (!safeText(raw, MAX_PATH_BYTES)) throw new GitAdapterError("origin remote identity is missing or invalid");
	const scp = /^(?:([^@/:]+)@)?([^/:]+):(.+)$/.exec(raw);
	if (scp && !/^[a-z][a-z0-9+.-]*:\/\//i.test(raw)) {
		const [, user, host, remotePath] = scp;
		if (user !== undefined && user !== "git") throw new GitAdapterError("origin remote must not contain credentials");
		if (!safeText(host, 255) || !safeText(remotePath, MAX_PATH_BYTES) || remotePath.startsWith("/") || remotePath.includes("..")) {
			throw new GitAdapterError("origin remote identity is invalid");
		}
		return `ssh://${host.toLowerCase()}/${remotePath.replace(/\.git$/i, "")}`;
	}

	let url: URL | undefined;
	try {
		url = new URL(raw);
	} catch {
		url = undefined;
	}
	if (url !== undefined) {
		if (url.password !== "" || (url.username !== "" && !(url.protocol === "ssh:" && url.username === "git"))) {
			throw new GitAdapterError("origin remote must not contain credentials");
		}
		if (url.search !== "" || url.hash !== "" || !["https:", "ssh:", "file:"].includes(url.protocol)) {
			throw new GitAdapterError("origin remote uses an unsafe URL shape");
		}
		if (url.protocol === "file:") return `file://${await realpath(url.pathname)}`;
		const path = url.pathname.replace(/^\/+|\/+$/g, "").replace(/\.git$/i, "");
		if (!safeText(url.hostname, 255) || !safeText(path, MAX_PATH_BYTES)) {
			throw new GitAdapterError("origin remote identity is invalid");
		}
		return `${url.protocol}//${url.hostname.toLowerCase()}${url.port ? `:${url.port}` : ""}/${path}`;
	}
	const local = resolve(repositoryRoot, raw);
	return `file://${await realpath(local)}`;
}

function parseStatus(raw: Buffer): GitStatusEvidence {
	if (raw.length === 0) return { clean: true, entries: [] };
	const tokens = raw.toString("utf8").split("\0");
	if (tokens.at(-1) === "") tokens.pop();
	const entries: GitStatusEntry[] = [];
	for (let index = 0; index < tokens.length; index += 1) {
		const record = tokens[index];
		if (record.length < 4 || record[2] !== " ") throw new GitAdapterError("Git returned malformed status evidence");
		const code = record.slice(0, 2);
		const path = record.slice(3).replaceAll("\\", "/");
		validateScope(path);
		if (code.includes("R") || code.includes("C")) {
			const originalPath = tokens[++index]?.replaceAll("\\", "/");
			validateScope(originalPath);
			entries.push({ code, path, originalPath });
		} else {
			entries.push({ code, path });
		}
	}
	entries.sort((left, right) => `${left.path}\0${left.originalPath ?? ""}`.localeCompare(`${right.path}\0${right.originalPath ?? ""}`));
	return { clean: false, entries };
}

function parseWorktrees(raw: string): GitWorktreeEvidence[] {
	const records: GitWorktreeEvidence[] = [];
	let current: GitWorktreeEvidence | undefined;
	for (const token of raw.split("\0")) {
		if (token === "") {
			if (current !== undefined) records.push(current);
			current = undefined;
			continue;
		}
		const separator = token.indexOf(" ");
		const key = separator === -1 ? token : token.slice(0, separator);
		const value = separator === -1 ? "" : token.slice(separator + 1);
		if (key === "worktree") {
			if (current !== undefined) records.push(current);
			if (!safeText(value, MAX_PATH_BYTES) || !isAbsolute(value)) throw new GitAdapterError("Git returned an unsafe worktree path");
			current = { cwd: value, detached: false, bare: false, locked: false, prunable: false };
			continue;
		}
		if (current === undefined) throw new GitAdapterError("Git returned malformed worktree evidence");
		switch (key) {
			case "HEAD":
				assertSha(value, "worktree head");
				current.head = value;
				break;
			case "branch":
				if (!value.startsWith("refs/heads/")) throw new GitAdapterError("Git returned an unsafe worktree branch");
				current.branch = value.slice("refs/heads/".length);
				assertSafeBranch(current.branch, "worktree branch");
				break;
			case "detached": current.detached = true; break;
			case "bare": current.bare = true; break;
			case "locked": current.locked = true; break;
			case "prunable": current.prunable = true; break;
			default: throw new GitAdapterError("Git returned unknown worktree evidence");
		}
	}
	if (current !== undefined) records.push(current);
	return records.sort((left, right) => left.cwd.localeCompare(right.cwd));
}

function errorExitCode(error: unknown): number | undefined {
	if (typeof error !== "object" || error === null || !("code" in error)) return undefined;
	return typeof error.code === "number" ? error.code : undefined;
}

export class GitAdapter {
	readonly #execute: GitCommandExecutor;
	readonly #timeoutMs: number;
	readonly #maxOutputBytes: number;

	constructor(options: GitAdapterOptions = {}) {
		this.#execute = options.execute ?? defaultExecutor;
		this.#timeoutMs = options.timeoutMs ?? DEFAULT_TIMEOUT_MS;
		this.#maxOutputBytes = options.maxOutputBytes ?? DEFAULT_MAX_OUTPUT_BYTES;
		if (!Number.isSafeInteger(this.#timeoutMs) || this.#timeoutMs < 1 || this.#timeoutMs > 60_000) {
			throw new GitAdapterError("Git timeout must be between 1 and 60000 milliseconds");
		}
		if (!Number.isSafeInteger(this.#maxOutputBytes) || this.#maxOutputBytes < 1024 || this.#maxOutputBytes > 8 * 1024 * 1024) {
			throw new GitAdapterError("Git output bound is invalid");
		}
	}

	async #run(cwd: string, args: readonly string[]): Promise<Buffer> {
		try {
			return await this.#execute({
				cwd,
				args: [...args],
				env: sanitizedGitEnvironment(),
				timeoutMs: this.#timeoutMs,
				maxOutputBytes: this.#maxOutputBytes,
			});
		} catch (error) {
			// Raw Git errors may contain remote URLs or host-environment details. Preserve only
			// the bounded exit status needed for typed control flow.
			throw new GitCommandFailure(errorExitCode(error));
		}
	}

	async inspect(cwd: string): Promise<GitBinding> {
		if (!safeText(cwd, MAX_PATH_BYTES) || !isAbsolute(cwd)) throw new GitAdapterError("Git cwd must be an absolute bounded path");
		const canonicalInput = await realpath(cwd);
		const rawPaths = (await this.#run(canonicalInput, [
			"rev-parse", "--path-format=absolute", "--show-toplevel", "--git-common-dir", "--git-dir",
		])).toString("utf8");
		const paths = stripLineEnding(rawPaths).split("\n");
		if (paths.length !== 3 || paths.some((path) => !safeText(path, MAX_PATH_BYTES) || !isAbsolute(path))) {
			throw new GitAdapterError("Git returned invalid canonical repository paths");
		}
		const [repositoryRoot, commonDirectory, worktreeDirectory] = await Promise.all(paths.map((path) => realpath(path)));
		let rawRemote: string;
		try {
			rawRemote = (await this.#run(repositoryRoot, ["config", "--local", "--no-includes", "--get", "remote.origin.url"])).toString("utf8");
		} catch (error) {
			throw new GitAdapterError("origin remote identity is missing or invalid", { cause: error });
		}
		const normalizedRemote = await normalizeRemote(rawRemote, repositoryRoot);
		const [commonMetadata, worktreeMetadata] = await Promise.all([stat(commonDirectory), stat(worktreeDirectory)]);
		const remoteIdentity = hashIdentity(["shepherd-origin-v1", normalizedRemote]);
		const repositoryIdentity = hashIdentity(["shepherd-repository-v2", metadataIdentity(commonMetadata), remoteIdentity]);
		const worktreeIdentity = hashIdentity(["shepherd-worktree-v2", repositoryIdentity, metadataIdentity(worktreeMetadata)]);
		return { cwd: repositoryRoot, repositoryIdentity, worktreeIdentity, remoteName: "origin", remoteIdentity };
	}

	async assertBinding(binding: GitBinding): Promise<GitBinding> {
		if (typeof binding !== "object" || binding === null) throw new GitAdapterError("Git binding is required");
		assertIdentity(binding.repositoryIdentity, "repository identity");
		assertIdentity(binding.worktreeIdentity, "worktree identity");
		assertIdentity(binding.remoteIdentity, "remote identity");
		if (binding.remoteName !== "origin") throw new GitAdapterError("only the origin remote is supported");
		const actual = await this.inspect(binding.cwd);
		if (actual.repositoryIdentity !== binding.repositoryIdentity) throw new GitAdapterError("repository identity mismatch");
		if (actual.worktreeIdentity !== binding.worktreeIdentity) throw new GitAdapterError("worktree identity mismatch");
		if (actual.remoteIdentity !== binding.remoteIdentity) throw new GitAdapterError("origin remote identity mismatch");
		return actual;
	}

	async status(binding: GitBinding): Promise<GitStatusEvidence> {
		const actual = await this.assertBinding(binding);
		return parseStatus(await this.#run(actual.cwd, [
			"-c", "core.fsmonitor=false",
			"-c", "core.untrackedCache=false",
			"status", "--porcelain=v1", "-z", "--untracked-files=all", "--ignore-submodules=none",
		]));
	}

	async currentBranch(binding: GitBinding): Promise<string> {
		const actual = await this.assertBinding(binding);
		const branch = stripLineEnding((await this.#run(actual.cwd, ["branch", "--show-current"])).toString("utf8"));
		assertSafeBranch(branch, "current branch");
		return branch;
	}

	async resolveBranchHead(binding: GitBinding, branch: string): Promise<string> {
		assertSafeBranch(branch);
		const actual = await this.assertBinding(binding);
		let head: string;
		try {
			head = stripLineEnding((await this.#run(actual.cwd, ["rev-parse", "--verify", `refs/heads/${branch}^{commit}`])).toString("utf8"));
		} catch (error) {
			throw new GitAdapterError(`branch ${branch} is not present`, { cause: error });
		}
		assertSha(head, "branch head");
		return head;
	}

	async #assertCommitObject(binding: GitBinding, head: string, description: string): Promise<void> {
		assertSha(head, description);
		const actual = await this.assertBinding(binding);
		try {
			const resolved = stripLineEnding((await this.#run(actual.cwd, ["rev-parse", "--verify", `${head}^{commit}`])).toString("utf8"));
			if (resolved !== head) throw new GitAdapterError(`${description} did not resolve exactly`);
		} catch (error) {
			throw new GitAdapterError(`${description} is not present as an exact commit object`, { cause: error });
		}
	}

	async listLocalBranches(binding: GitBinding): Promise<GitBranchEvidence[]> {
		const actual = await this.assertBinding(binding);
		const raw = (await this.#run(actual.cwd, ["for-each-ref", "--format=%(refname:short)%09%(objectname)", "refs/heads"])).toString("utf8");
		const branches: GitBranchEvidence[] = [];
		for (const line of raw.split("\n")) {
			if (line === "") continue;
			const [branch, head, ...extra] = line.split("\t");
			if (extra.length > 0) throw new GitAdapterError("Git returned malformed branch evidence");
			assertSafeBranch(branch);
			assertSha(head, "branch head");
			branches.push({ branch, head });
		}
		return branches.sort((left, right) => left.branch.localeCompare(right.branch));
	}

	async listWorktrees(binding: GitBinding): Promise<GitWorktreeEvidence[]> {
		const actual = await this.assertBinding(binding);
		return parseWorktrees((await this.#run(actual.cwd, ["worktree", "list", "--porcelain", "-z"])).toString("utf8"));
	}

	async fetchBranch(binding: GitBinding, branch: string): Promise<string> {
		assertSafeBranch(branch);
		const actual = await this.assertBinding(binding);
		await this.#run(actual.cwd, ["fetch", "--no-tags", "origin", branch]);
		const fetched = stripLineEnding((await this.#run(actual.cwd, ["rev-parse", "--verify", "FETCH_HEAD^{commit}"])).toString("utf8"));
		assertSha(fetched, "fetched head");
		return fetched;
	}

	async addIssueWorktree(binding: GitBinding, request: GitAddWorktreeRequest): Promise<GitBinding> {
		assertCanonicalIssueBranch(request.issue, request.slug, request.branch);
		assertSha(request.baseHead, "base head");
		await this.#assertCommitObject(binding, request.baseHead, "base head");
		if (!safeText(request.trustedRoot, MAX_PATH_BYTES) || !isAbsolute(request.trustedRoot)) {
			throw new GitAdapterError("trusted worktree root must be an absolute bounded path");
		}
		const root = await realpath(request.trustedRoot);
		const expectedPath = resolve(root, canonicalIssueWorktreeName(request.issue, request.slug));
		if (request.path !== expectedPath || relative(root, request.path).startsWith(`..${sep}`)) {
			throw new GitAdapterError("worktree path must be the canonical child of the trusted root");
		}
		const branches = await this.listLocalBranches(binding);
		const existing = branches.find((candidate) => candidate.branch === request.branch);
		if (existing !== undefined && !(await this.isAncestor(binding, request.baseHead, existing.head))) {
			throw new GitAdapterError("existing issue branch does not descend from the exact base head");
		}
		const actual = await this.assertBinding(binding);
		try {
			if (existing === undefined) {
				await this.#run(actual.cwd, ["worktree", "add", "-b", request.branch, "--", request.path, request.baseHead]);
			} else {
				await this.#run(actual.cwd, ["worktree", "add", "--", request.path, request.branch]);
			}
		} catch (error) {
			throw new GitAdapterError("typed Git worktree creation failed; existing state was preserved", { cause: error });
		}
		const created = await this.inspect(request.path);
		if (created.repositoryIdentity !== actual.repositoryIdentity || created.remoteIdentity !== actual.remoteIdentity) {
			throw new GitAdapterError("created worktree repository identity mismatch");
		}
		if (await this.currentBranch(created) !== request.branch) throw new GitAdapterError("created worktree branch mismatch");
		const head = await this.resolveBranchHead(created, request.branch);
		if (!(await this.isAncestor(created, request.baseHead, head))) throw new GitAdapterError("created worktree lost its exact base ancestry");
		return created;
	}

	async isAncestor(binding: GitBinding, baseHead: string, head: string): Promise<boolean> {
		assertSha(baseHead, "base head");
		assertSha(head, "head");
		const actual = await this.assertBinding(binding);
		try {
			await this.#run(actual.cwd, ["merge-base", "--is-ancestor", baseHead, head]);
			return true;
		} catch (error) {
			if (error instanceof GitCommandFailure && error.exitCode === 1) return false;
			throw error;
		}
	}

	async diff(binding: GitBinding, request: { baseHead: string; head: string; scopes: readonly string[] }): Promise<GitDiffEvidence> {
		const scopes = validateScopes(request.scopes);
		await this.#assertCommitObject(binding, request.baseHead, "base head");
		await this.#assertCommitObject(binding, request.head, "head");
		const actual = await this.assertBinding(binding);
		const raw = (await this.#run(actual.cwd, ["diff", "--name-only", "-z", request.baseHead, request.head, "--", ...scopes])).toString("utf8");
		const changedScope = raw.split("\0").filter(Boolean).map((path) => {
			validateScope(path);
			if (!pathWithinScope(path, scopes)) throw new GitAdapterError("Git diff escaped declared scopes");
			return path;
		}).sort();
		return { baseHead: request.baseHead, head: request.head, changedScope };
	}

	async commitIssueChanges(binding: GitBinding, request: GitCommitRequest): Promise<GitCommitEvidence> {
		assertCanonicalIssueBranch(request.issue, request.slug, request.branch);
		assertSha(request.expectedHead, "expected head");
		if (!safeText(request.message, 512)) throw new GitAdapterError("commit message must be bounded safe text");
		const scopes = validateScopes(request.scopes);
		const actual = await this.assertBinding(binding);
		if (await this.currentBranch(actual) !== request.branch) throw new GitAdapterError("current branch does not match canonical issue branch");
		const previousHead = await this.resolveBranchHead(actual, request.branch);
		if (previousHead !== request.expectedHead) throw new GitAdapterError("stale expected head; commit was not attempted");
		const status = await this.status(actual);
		const outside = status.entries.flatMap((entry) => [entry.path, ...(entry.originalPath ? [entry.originalPath] : [])])
			.filter((path) => !pathWithinScope(path, scopes));
		if (outside.length > 0) throw new GitAdapterError(`dirty or staged state exists outside declared scopes: ${outside.sort().join(", ")}`);
		if (status.clean) return { committed: false, previousHead, head: previousHead };
		await this.#run(actual.cwd, ["add", "-A", "--", ...scopes]);
		const staged = (await this.#run(actual.cwd, ["diff", "--cached", "--name-only", "-z"])).toString("utf8")
			.split("\0").filter(Boolean);
		if (staged.some((path) => !pathWithinScope(path, scopes))) {
			throw new GitAdapterError("staged state escaped declared scopes; state was preserved for inspection");
		}
		if (staged.length === 0) return { committed: false, previousHead, head: previousHead };
		await this.#run(actual.cwd, [
			"-c", "core.hooksPath=/dev/null",
			"-c", "commit.gpgSign=false",
			"commit", "--no-verify", "-m", request.message,
		]);
		const head = await this.resolveBranchHead(actual, request.branch);
		if (head === previousHead) throw new GitAdapterError("commit did not advance the exact head");
		return { committed: true, previousHead, head };
	}

	async pushIssueBranch(binding: GitBinding, request: GitPushRequest): Promise<GitPushEvidence> {
		assertCanonicalIssueBranch(request.issue, request.slug, request.branch);
		assertSafeBranch(request.defaultBranch, "default branch");
		assertSha(request.expectedHead, "expected head");
		if (["main", "master", "trunk", request.defaultBranch].includes(request.branch)) {
			throw new GitAdapterError("direct default branch push is unavailable");
		}
		const actual = await this.assertBinding(binding);
		if (await this.currentBranch(actual) !== request.branch) throw new GitAdapterError("current branch does not match canonical issue branch");
		const head = await this.resolveBranchHead(actual, request.branch);
		if (head !== request.expectedHead) throw new GitAdapterError("stale expected head; push was not attempted");
		await this.#run(actual.cwd, ["push", "--porcelain", "origin", request.branch]);
		const remote = stripLineEnding((await this.#run(actual.cwd, ["ls-remote", "--heads", "origin", `refs/heads/${request.branch}`])).toString("utf8"));
		const [remoteHead, remoteRef, ...extra] = remote.split("\t");
		if (extra.length > 0 || remoteHead !== head || remoteRef !== `refs/heads/${request.branch}`) {
			throw new GitAdapterError("remote exact-head verification failed after push");
		}
		return { branch: request.branch, head, remoteName: "origin" };
	}
}
