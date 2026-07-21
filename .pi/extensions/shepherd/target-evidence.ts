import { execFile } from "node:child_process";
import { createHash } from "node:crypto";
import { realpath, stat } from "node:fs/promises";
import { isAbsolute } from "node:path";

const MAX_GITHUB_NUMBER = 2_147_483_647;
const EVIDENCE_COMMAND_TIMEOUT_MS = 15_000;
const IDENTITY_COMMAND_TIMEOUT_MS = 15_000;
const SAFE_IDENTITY = /^[0-9a-f]{64}$/;
const CLEAN_STATUS_ARGS = [
	"-c", "core.fsmonitor=false",
	"-c", "core.untrackedCache=false",
	"-c", "status.showUntrackedFiles=all",
	"-c", "diff.ignoreSubmodules=none",
	"status", "--porcelain=v1", "--untracked-files=all", "--ignore-submodules=none",
] as const;

export interface TargetEvidenceRequest {
	cwd: string;
	issue: number;
	pr?: number;
	repositoryIdentity: string;
	worktreeIdentity: string;
}

export interface TargetEvidence {
	cwd: string;
	repositoryIdentity: string;
	worktreeIdentity: string;
	branch: string;
	candidateHead: string;
	clean: true;
	pr?: number;
	prUrl?: string;
	baseBranch?: string;
	draft?: boolean;
	prState?: string;
	mergeStateStatus?: string;
	reviewDecision?: string;
	statusChecks?: StatusCheckEvidence[];
}

export interface CanonicalGitWorktree {
	cwd: string;
	repositoryIdentity: string;
	worktreeIdentity: string;
}

export interface CanonicalGitWorktreeOptions {
	signal?: AbortSignal;
	timeoutMs?: number;
}

export interface StatusCheckEvidence {
	name: string;
	status: string;
	conclusion?: string;
}

export type ArgvExecutor = (file: string, args: string[], cwd: string) => Promise<string>;

interface GitHubPullRequest {
	number: number;
	state: string;
	isDraft: boolean;
	baseRefName: string;
	headRefName: string;
	headRefOid: string;
	url: string;
	mergeStateStatus: string;
	reviewDecision: string;
	statusCheckRollup: unknown[];
}

interface GitHubRepositoryIdentity {
	host: string;
	owner: string;
	name: string;
	selector: string;
}

export class TargetEvidenceError extends Error {
	constructor(message: string, options?: ErrorOptions) {
		super(message, options);
		this.name = "TargetEvidenceError";
	}
}

function validNumber(value: unknown): value is number {
	return Number.isSafeInteger(value) && (value as number) > 0 && (value as number) <= MAX_GITHUB_NUMBER;
}

function safeText(value: unknown, maximum = 512): value is string {
	return typeof value === "string" && value.length > 0 && value.length <= maximum && !/[\u0000-\u001f\u007f]/.test(value);
}

function safeOptionalText(value: unknown, maximum = 512): value is string {
	return typeof value === "string" && value.length <= maximum && !/[\u0000-\u001f\u007f]/.test(value);
}

function safeIdentity(value: unknown): value is string {
	return typeof value === "string" && SAFE_IDENTITY.test(value);
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
			|| key.startsWith("GIT_CONFIG_VALUE_")) {
			delete env[key];
		}
	}
	env.GIT_CONFIG_NOSYSTEM = "1";
	env.GIT_CONFIG_GLOBAL = process.platform === "win32" ? "NUL" : "/dev/null";
	env.GIT_OPTIONAL_LOCKS = "0";
	return env;
}

function parseStatusChecks(value: unknown): StatusCheckEvidence[] {
	if (!Array.isArray(value) || value.length > 256) {
		throw new TargetEvidenceError("GitHub returned invalid status check evidence");
	}
	return value.map((candidate) => {
		if (typeof candidate !== "object" || candidate === null || Array.isArray(candidate)) {
			throw new TargetEvidenceError("GitHub returned invalid status check evidence");
		}
		const check = candidate as Record<string, unknown>;
		const name = check.name ?? check.context;
		const status = check.status ?? check.state;
		if (!safeText(name) || !safeText(status, 64)) {
			throw new TargetEvidenceError("GitHub returned invalid status check evidence");
		}
		if (check.conclusion !== undefined && check.conclusion !== null && !safeText(check.conclusion, 64)) {
			throw new TargetEvidenceError("GitHub returned invalid status check evidence");
		}
		return {
			name,
			status,
			...(typeof check.conclusion === "string" ? { conclusion: check.conclusion } : {}),
		};
	}).sort((left, right) =>
		`${left.name}\u0000${left.status}\u0000${left.conclusion ?? ""}`.localeCompare(
			`${right.name}\u0000${right.status}\u0000${right.conclusion ?? ""}`,
		),
	);
}

function defaultExec(file: string, args: string[], cwd: string): Promise<string> {
	return new Promise((resolve, reject) => {
		execFile(file, args, {
			cwd,
			encoding: "utf8",
			env: file === "git" ? sanitizedGitEnvironment() : process.env,
			maxBuffer: 1024 * 1024,
			timeout: EVIDENCE_COMMAND_TIMEOUT_MS,
			killSignal: "SIGTERM",
		}, (error, stdout) => {
			if (error) {
				reject(new TargetEvidenceError(`${file} command failed while capturing target evidence`, { cause: error }));
				return;
			}
			resolve(stdout);
		});
	});
}

function executeIdentityCommand(
	args: string[],
	cwd: string,
	signal: AbortSignal,
	timeoutMs: number,
): Promise<string> {
	return new Promise((resolve, reject) => {
		execFile("git", args, {
			cwd,
			encoding: "utf8",
			env: sanitizedGitEnvironment(),
			maxBuffer: 64 * 1024,
			timeout: timeoutMs,
			killSignal: "SIGTERM",
			signal,
		}, (error, stdout) => {
			if (error) {
				reject(error);
				return;
			}
			resolve(stdout);
		});
	});
}

function stripOneLineEnding(value: string): string {
	return value.endsWith("\r\n") ? value.slice(0, -2) : value.endsWith("\n") ? value.slice(0, -1) : value;
}

function normalizeRemoteIdentity(raw: string): string {
	const remote = stripOneLineEnding(raw);
	if (remote === "") return "no-remote";
	if (remote.length > 4_096 || /[\u0000-\u001f\u007f]/.test(remote)) {
		throw new TargetEvidenceError("Git remote identity is invalid");
	}
	const scp = /^(?:[^@/:]+@)?([^/:]+):(.+)$/.exec(remote);
	if (scp && !/^[a-z][a-z0-9+.-]*:\/\//i.test(remote)) {
		return `${scp[1].toLowerCase()}/${scp[2].replace(/^\/+|\/+$/g, "").replace(/\.git$/i, "")}`;
	}
	try {
		const url = new URL(remote);
		const path = url.pathname.replace(/^\/+|\/+$/g, "").replace(/\.git$/i, "");
		if (url.protocol === "file:") return `file/${path}`;
		return `${url.hostname.toLowerCase()}${url.port ? `:${url.port}` : ""}/${path}`;
	} catch {
		return `local/${remote.replace(/\\/g, "/").replace(/\/+$/g, "").replace(/\.git$/i, "")}`;
	}
}

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

async function filesystemIdentity(path: string): Promise<string> {
	const metadata = await stat(path, { bigint: true });
	const birthtime = "birthtimeNs" in metadata ? metadata.birthtimeNs : 0n;
	return `${metadata.dev}:${metadata.ino}:${birthtime > 0n ? birthtime : 0n}`;
}

function abortFailure(signal: AbortSignal): Error {
	if (signal.reason instanceof TargetEvidenceError) return signal.reason;
	return new TargetEvidenceError("canonical Git worktree lookup was aborted", {
		...(signal.reason instanceof Error ? { cause: signal.reason } : {}),
	});
}

export async function resolveCanonicalGitWorktree(
	cwd: string,
	options: CanonicalGitWorktreeOptions = {},
): Promise<CanonicalGitWorktree> {
	if (!safeText(cwd, 4_096) || !isAbsolute(cwd)) {
		throw new TargetEvidenceError("target working directory must be an absolute safe path");
	}
	const timeoutMs = options.timeoutMs ?? IDENTITY_COMMAND_TIMEOUT_MS;
	if (!Number.isSafeInteger(timeoutMs) || timeoutMs < 1 || timeoutMs > 60_000) {
		throw new TargetEvidenceError("canonical Git worktree timeout is invalid");
	}
	const controller = new AbortController();
	const abortFromCaller = () => controller.abort(options.signal?.reason);
	if (options.signal?.aborted) abortFromCaller();
	else options.signal?.addEventListener("abort", abortFromCaller, { once: true });
	const timer = setTimeout(
		() => controller.abort(new TargetEvidenceError("canonical Git worktree lookup timed out")),
		timeoutMs,
	);
	const aborted = new Promise<never>((_resolve, reject) => {
		if (controller.signal.aborted) reject(abortFailure(controller.signal));
		else controller.signal.addEventListener("abort", () => reject(abortFailure(controller.signal)), { once: true });
	});
	const lookup = (async () => {
		const rawPaths = await executeIdentityCommand(
			["rev-parse", "--path-format=absolute", "--show-toplevel", "--git-common-dir", "--git-dir"],
			cwd,
			controller.signal,
			timeoutMs,
		);
		const paths = stripOneLineEnding(rawPaths).split("\n");
		if (paths.length !== 3 || paths.some((path) => !safeText(path, 4_096) || !isAbsolute(path))) {
			throw new TargetEvidenceError("Git returned invalid canonical worktree paths");
		}
		const [canonicalCwd, commonDir, gitDir] = await Promise.all(paths.map((path) => realpath(path)));
		if (controller.signal.aborted) throw abortFailure(controller.signal);
		let rawRemote = "";
		try {
			rawRemote = await executeIdentityCommand(
				["config", "--local", "--no-includes", "--get", "remote.origin.url"],
				canonicalCwd,
				controller.signal,
				timeoutMs,
			);
		} catch (error) {
			if (controller.signal.aborted) throw abortFailure(controller.signal);
			if (!(typeof error === "object" && error !== null && "code" in error && error.code === 1)) throw error;
		}
		const [repositoryFilesystemIdentity, worktreeFilesystemIdentity] = await Promise.all([
			filesystemIdentity(commonDir),
			filesystemIdentity(gitDir),
		]);
		if (controller.signal.aborted) throw abortFailure(controller.signal);
		const repositoryIdentity = hashIdentity([
			"shepherd-repository-v1",
			repositoryFilesystemIdentity,
			normalizeRemoteIdentity(rawRemote),
		]);
		const worktreeIdentity = hashIdentity([
			"shepherd-worktree-v1",
			repositoryIdentity,
			worktreeFilesystemIdentity,
		]);
		return { cwd: canonicalCwd, repositoryIdentity, worktreeIdentity };
	})();
	try {
		return await Promise.race([lookup, aborted]);
	} catch (error) {
		if (error instanceof TargetEvidenceError) throw error;
		throw new TargetEvidenceError("failed to resolve canonical Git worktree identity", { cause: error });
	} finally {
		clearTimeout(timer);
		options.signal?.removeEventListener("abort", abortFromCaller);
	}
}

function parseGitHubRepositoryIdentity(rawRemote: string): GitHubRepositoryIdentity {
	const normalized = normalizeRemoteIdentity(rawRemote);
	const match = /^([a-z0-9.-]+(?::[0-9]{1,5})?)\/([a-z0-9][a-z0-9._-]{0,99})\/([a-z0-9][a-z0-9._-]{0,99})$/i.exec(normalized);
	if (!match) {
		throw new TargetEvidenceError("Git origin does not identify a safe GitHub repository");
	}
	const [, host, owner, name] = match;
	return {
		host: host.toLowerCase(),
		owner,
		name,
		selector: `${host.toLowerCase()}/${owner}/${name}`,
	};
}

function safePullRequestUrl(
	value: string,
	expectedPR: number,
	expectedRepository: GitHubRepositoryIdentity,
): string {
	let url: URL;
	try {
		url = new URL(value);
	} catch (error) {
		throw new TargetEvidenceError("GitHub returned an invalid pull request URL", { cause: error });
	}
	if (url.protocol !== "https:"
		|| url.username !== ""
		|| url.password !== ""
		|| url.search !== ""
		|| url.hash !== ""
		|| url.host.toLowerCase() !== expectedRepository.host
		|| url.pathname.toLowerCase() !== `/${expectedRepository.owner}/${expectedRepository.name}/pull/${expectedPR}`.toLowerCase()) {
		if (url.host.toLowerCase() !== expectedRepository.host
			|| !url.pathname.toLowerCase().startsWith(`/${expectedRepository.owner}/${expectedRepository.name}/`.toLowerCase())) {
			throw new TargetEvidenceError("GitHub returned a pull request URL for a different repository");
		}
		throw new TargetEvidenceError("GitHub returned an unsafe pull request URL");
	}
	return url.toString();
}

function parsePullRequest(
	raw: string,
	expectedPR: number,
	expectedRepository: GitHubRepositoryIdentity,
): GitHubPullRequest {
	let value: unknown;
	try {
		value = JSON.parse(raw);
	} catch (error) {
		throw new TargetEvidenceError("GitHub returned invalid pull request evidence", { cause: error });
	}
	if (typeof value !== "object" || value === null || Array.isArray(value)) {
		throw new TargetEvidenceError("GitHub returned invalid pull request evidence");
	}
	const pr = value as Record<string, unknown>;
	if (pr.number !== expectedPR) throw new TargetEvidenceError("pull request identity mismatch");
	if (
		!safeText(pr.state, 32)
		|| typeof pr.isDraft !== "boolean"
		|| !safeText(pr.baseRefName)
		|| !safeText(pr.headRefName)
		|| typeof pr.headRefOid !== "string"
		|| !/^[0-9a-f]{40}$/i.test(pr.headRefOid)
		|| !safeText(pr.url, 2_048)
		|| !safeText(pr.mergeStateStatus, 64)
		|| !safeOptionalText(pr.reviewDecision, 64)
	) {
		throw new TargetEvidenceError("GitHub returned incomplete pull request evidence");
	}
	return {
		...(pr as unknown as GitHubPullRequest),
		url: safePullRequestUrl(pr.url as string, expectedPR, expectedRepository),
	};
}

export async function captureTargetEvidence(
	request: TargetEvidenceRequest,
	execute: ArgvExecutor = defaultExec,
): Promise<TargetEvidence> {
	if (!validNumber(request?.issue)) throw new TargetEvidenceError("issue must be a positive bounded integer");
	if (request?.pr !== undefined && !validNumber(request.pr)) {
		throw new TargetEvidenceError("pull request must be a positive bounded integer");
	}
	if (!safeText(request?.cwd, 4_096) || !isAbsolute(request.cwd)) {
		throw new TargetEvidenceError("target working directory must be an absolute safe path");
	}
	if (!safeIdentity(request.repositoryIdentity) || !safeIdentity(request.worktreeIdentity)) {
		throw new TargetEvidenceError("target repository and worktree identities must be safe hashes");
	}

	const [rawHead, rawBranch, rawStatus] = await Promise.all([
		execute("git", ["rev-parse", "HEAD"], request.cwd),
		execute("git", ["branch", "--show-current"], request.cwd),
		execute("git", [...CLEAN_STATUS_ARGS], request.cwd),
	]);
	let rawPR: string | undefined;
	let expectedRepository: GitHubRepositoryIdentity | undefined;
	if (request.pr !== undefined) {
		const rawRemote = await execute(
			"git",
			["config", "--local", "--no-includes", "--get", "remote.origin.url"],
			request.cwd,
		);
		expectedRepository = parseGitHubRepositoryIdentity(rawRemote);
		rawPR = await execute(
			"gh",
			[
				"pr",
				"view",
				String(request.pr),
				"--repo",
				expectedRepository.selector,
				"--json",
				"number,state,isDraft,baseRefName,headRefName,headRefOid,url,mergeStateStatus,reviewDecision,statusCheckRollup",
			],
			request.cwd,
		);
	}

	const candidateHead = rawHead.trim().toLowerCase();
	const branch = rawBranch.trim();
	if (!/^[0-9a-f]{40}$/.test(candidateHead)) throw new TargetEvidenceError("local target head is invalid");
	if (!safeText(branch)) throw new TargetEvidenceError("local target branch is invalid or detached");
	if (rawStatus.trim() !== "") throw new TargetEvidenceError("target working tree must be clean");

	if (rawPR === undefined || request.pr === undefined || expectedRepository === undefined) {
		return {
			cwd: request.cwd,
			repositoryIdentity: request.repositoryIdentity,
			worktreeIdentity: request.worktreeIdentity,
			branch,
			candidateHead,
			clean: true,
		};
	}

	const pr = parsePullRequest(rawPR, request.pr, expectedRepository);
	if (pr.state.toUpperCase() !== "OPEN") throw new TargetEvidenceError(`pull request #${request.pr} must be open`);
	if (branch !== pr.headRefName) throw new TargetEvidenceError("local branch does not match pull request head branch");
	if (candidateHead !== pr.headRefOid.toLowerCase()) throw new TargetEvidenceError("local head does not match pull request head");

	return {
		cwd: request.cwd,
		repositoryIdentity: request.repositoryIdentity,
		worktreeIdentity: request.worktreeIdentity,
		branch,
		candidateHead,
		clean: true,
		pr: request.pr,
		prUrl: pr.url,
		baseBranch: pr.baseRefName,
		draft: pr.isDraft,
		prState: pr.state.toUpperCase(),
		mergeStateStatus: pr.mergeStateStatus,
		reviewDecision: pr.reviewDecision,
		statusChecks: parseStatusChecks(pr.statusCheckRollup),
	};
}
