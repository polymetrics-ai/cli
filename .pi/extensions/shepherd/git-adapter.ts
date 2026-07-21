import { createHash } from "node:crypto";
import { execFile } from "node:child_process";
import { lstat, realpath, stat } from "node:fs/promises";
import { isAbsolute, join, relative, resolve, sep } from "node:path";
import { fileURLToPath } from "node:url";

import { FileStateStore, type FileStateStoreOptions, type RunLease } from "./state-store.ts";
import {
	registerGitAdapterMutationLeaseAcquirer,
	type GitAdapterMutationLeaseAcquirer,
} from "./workspace-adapter.ts";

const DEFAULT_TIMEOUT_MS = 15_000;
const DEFAULT_MAX_OUTPUT_BYTES = 1024 * 1024;
const MAX_PATH_BYTES = 4_096;
const MAX_BRANCH_BYTES = 240;
const MAX_SCOPES = 64;
const SHA_PATTERN = /^[0-9a-f]{40}$/;
const IDENTITY_PATTERN = /^[0-9a-f]{64}$/;
const SLUG_PATTERN = /^[a-z0-9]+(?:-[a-z0-9]+)*$/;
const NULL_DEVICE = process.platform === "win32" ? "NUL" : "/dev/null";
const SAFE_MUTATION_CONFIG = [
	["core.hooksPath", NULL_DEVICE],
	["core.fsmonitor", "false"],
	["core.untrackedCache", "false"],
	["core.attributesFile", NULL_DEVICE],
	["core.askPass", ""],
	["credential.helper", ""],
	["commit.gpgSign", "false"],
	["push.gpgSign", "false"],
	["push.pushOption", ""],
	["tag.gpgSign", "false"],
	["protocol.ext.allow", "never"],
	["submodule.recurse", "false"],
] as const;

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
	fetchEndpointIdentity: string;
	pushEndpointIdentity: string;
	defaultBranch?: string;
}

export interface GitMutationLease {
	assertOwned(): Promise<void>;
	release(): Promise<void>;
}

export interface GitMutationLeaseRequest {
	issue: number;
	slug: string;
	branch: string;
	baseHead: string;
	targetCwd: string;
	allowedScopes: readonly string[];
	stateRoot: string;
	runId: string;
	mode: "start" | "resume";
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

interface EffectiveGitEndpoint {
	value?: string;
	identity: string;
}

interface EffectiveGitRemote {
	fetch: EffectiveGitEndpoint;
	push: EffectiveGitEndpoint;
}

interface MutationLeaseState {
	runLease: RunLease;
	coordinator: GitBinding;
	targetCwd: string;
	targetWorktreeIdentity?: string;
	issue: number;
	slug: string;
	branch: string;
	baseHead: string;
	allowedScopes: string[];
	accepting: boolean;
	tail: Promise<void>;
	releasePromise?: Promise<void>;
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
		|| scope.includes("\\")
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
		unique.add(scope);
	}
	return [...unique].sort();
}

export function canonicalGitScopes(scopes: readonly string[]): string[] {
	return validateScopes(scopes);
}

function pathWithinScope(path: string, scopes: readonly string[]): boolean {
	return scopes.some((scope) => path === scope || path.startsWith(`${scope}/`));
}

function sanitizedGitEnvironment(mutation: boolean): NodeJS.ProcessEnv {
	const env = { ...process.env };
	for (const key of Object.keys(env)) {
		const upper = key.toUpperCase();
		if (upper.startsWith("GIT_")
			|| upper === "SSH_ASKPASS"
			|| upper === "SSH_ASKPASS_REQUIRE"
			|| upper === "GCM_INTERACTIVE"
			|| upper === "PAGER"
			|| upper === "EDITOR"
			|| upper === "VISUAL"
			|| upper.endsWith("_PROXY")
			|| upper === "LD_PRELOAD"
			|| upper.startsWith("DYLD_")) delete env[key];
	}
	env.GIT_CONFIG_NOSYSTEM = "1";
	env.GIT_CONFIG_GLOBAL = NULL_DEVICE;
	env.GIT_TERMINAL_PROMPT = "0";
	env.GIT_OPTIONAL_LOCKS = "0";
	env.GIT_PAGER = "";
	env.LC_ALL = "C";
	if (mutation) {
		env.GIT_CONFIG_COUNT = String(SAFE_MUTATION_CONFIG.length);
		for (const [index, [key, value]] of SAFE_MUTATION_CONFIG.entries()) {
			env[`GIT_CONFIG_KEY_${index}`] = key;
			env[`GIT_CONFIG_VALUE_${index}`] = value;
		}
	}
	return env;
}

function unsafeMutationConfigKey(rawKey: string): boolean {
	const key = rawKey.toLowerCase();
	return key === "core.hookspath"
		|| key === "core.fsmonitor"
		|| key === "core.sshcommand"
		|| key === "core.gitproxy"
		|| key === "core.askpass"
		|| key === "ssh.variant"
		|| key === "credential.helper"
		|| key.startsWith("credential.")
		|| /^filter\..+\.(clean|smudge|process|required)$/.test(key)
		|| /^remote\..+\.(receivepack|uploadpack|vcs|proxy)$/.test(key)
		|| key === "https.proxy"
		|| key.startsWith("http.")
		|| key === "protocol.allow"
		|| /^protocol\..+\.allow$/.test(key)
		|| key.startsWith("push.")
		|| /^submodule\..+\.update$/.test(key)
		|| key === "include.path"
		|| /^includeif\..+\.path$/.test(key);
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

async function filesystemIdentity(path: string): Promise<string> {
	const metadata = await stat(path, { bigint: true });
	const birthtime = "birthtimeNs" in metadata ? metadata.birthtimeNs : 0n;
	return `${metadata.dev}:${metadata.ino}:${birthtime > 0n ? birthtime : 0n}`;
}

function stripLineEnding(value: string): string {
	return value.endsWith("\r\n") ? value.slice(0, -2) : value.endsWith("\n") ? value.slice(0, -1) : value;
}

function normalizeRemote(rawOutput: string): string {
	const raw = stripLineEnding(rawOutput);
	if (raw === "") return "no-remote";
	if (raw.length > MAX_PATH_BYTES || /[\u0000-\u001f\u007f-\u009f]/.test(raw)) {
		throw new GitAdapterError("origin remote identity is missing or invalid");
	}
	const scp = /^(?:([^@/:]+)@)?([^/:]+):(.+)$/.exec(raw);
	if (scp && !/^[a-z][a-z0-9+.-]*:\/\//i.test(raw)) {
		const [, user, host, remotePath] = scp;
		if (user !== undefined && user !== "git") throw new GitAdapterError("origin remote must not contain credentials");
		if (!safeText(host, 255) || !safeText(remotePath, MAX_PATH_BYTES) || remotePath.startsWith("/") || remotePath.includes("..")) {
			throw new GitAdapterError("origin remote identity is invalid");
		}
		return `${host.toLowerCase()}/${remotePath.replace(/^\/+|\/+$/g, "").replace(/\.git$/i, "")}`;
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
		const path = url.pathname.replace(/^\/+|\/+$/g, "").replace(/\.git$/i, "");
		if (url.protocol !== "file:" && (!safeText(url.hostname, 255) || !safeText(path, MAX_PATH_BYTES))) {
			throw new GitAdapterError("origin remote identity is invalid");
		}
		if (url.protocol === "file:") return `file/${path}`;
		return `${url.hostname.toLowerCase()}${url.port ? `:${url.port}` : ""}/${path}`;
	}
	return `local/${raw.replace(/\\/g, "/").replace(/\/+$/g, "").replace(/\.git$/i, "")}`;
}

async function normalizeEffectiveEndpoint(raw: string, cwd: string): Promise<string> {
	if (!safeText(raw, MAX_PATH_BYTES) || raw.startsWith("-")) {
		throw new GitAdapterError("origin effective endpoint is missing or invalid");
	}
	const scp = /^(?:([^@/:]+)@)?([^/:]+):(.+)$/.exec(raw);
	if (scp && !/^[a-z][a-z0-9+.-]*:\/\//i.test(raw)) {
		const [, user, host, remotePath] = scp;
		if (user !== undefined && user !== "git") throw new GitAdapterError("origin effective endpoint must not contain credentials");
		if (!safeText(host, 255) || !safeText(remotePath, MAX_PATH_BYTES)
			|| remotePath.startsWith("/") || remotePath.startsWith("-")
			|| remotePath.split("/").some((part) => part === "" || part === "." || part === "..")) {
			throw new GitAdapterError("origin effective endpoint is invalid");
		}
		return `scp:${user === "git" ? "git@" : ""}${host.toLowerCase()}:${remotePath}`;
	}

	let url: URL | undefined;
	try {
		url = new URL(raw);
	} catch {
		url = undefined;
	}
	if (url !== undefined) {
		if (url.password !== "" || (url.username !== "" && !(url.protocol === "ssh:" && url.username === "git"))) {
			throw new GitAdapterError("origin effective endpoint must not contain credentials");
		}
		if (url.search !== "" || url.hash !== "" || !["https:", "ssh:", "file:"].includes(url.protocol)) {
			throw new GitAdapterError("origin effective endpoint uses an unsafe URL shape");
		}
		if (url.protocol === "file:") {
			if (url.hostname !== "") throw new GitAdapterError("origin file endpoint must be local");
			let endpointPath: string;
			try {
				endpointPath = await realpath(fileURLToPath(url));
			} catch (error) {
				throw new GitAdapterError("origin file endpoint is not a canonical local path", { cause: error });
			}
			return `file:${endpointPath}`;
		}
		if (!safeText(url.hostname, 255) || !safeText(url.pathname, MAX_PATH_BYTES) || url.pathname === "/") {
			throw new GitAdapterError("origin effective endpoint is invalid");
		}
		const user = url.protocol === "ssh:" && url.username === "git" ? "git@" : "";
		return `${url.protocol}//${user}${url.hostname.toLowerCase()}${url.port ? `:${url.port}` : ""}${url.pathname}`;
	}

	let endpointPath: string;
	try {
		endpointPath = await realpath(resolve(cwd, raw));
	} catch (error) {
		throw new GitAdapterError("origin local endpoint is not a canonical path", { cause: error });
	}
	return `local:${endpointPath}`;
}

function endpointIdentity(normalized: string): string {
	return hashIdentity(["shepherd-git-endpoint-v1", normalized]);
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
		const path = record.slice(3);
		validateScope(path);
		if (code.includes("R") || code.includes("C")) {
			const originalPath = tokens[++index];
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
	readonly #mutationLeases = new WeakMap<GitMutationLease, MutationLeaseState>();

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
		const acquireMutationLease: GitAdapterMutationLeaseAcquirer = (binding, request, leaseOptions) => (
			this.#acquireMutationLease(binding, request, leaseOptions)
		);
		registerGitAdapterMutationLeaseAcquirer(this, acquireMutationLease);
	}

	async #run(cwd: string, args: readonly string[], mutation = false): Promise<Buffer> {
		try {
			return await this.#execute({
				cwd,
				args: [...args],
				env: sanitizedGitEnvironment(mutation),
				timeoutMs: this.#timeoutMs,
				maxOutputBytes: this.#maxOutputBytes,
			});
		} catch (error) {
			// Raw Git errors may contain remote URLs or host-environment details. Preserve only
			// the bounded exit status needed for typed control flow.
			throw new GitCommandFailure(errorExitCode(error));
		}
	}

	async #localConfigKeys(cwd: string, scope: "--local" | "--worktree"): Promise<string[]> {
		const raw = (await this.#run(cwd, [
			"config", scope, "--no-includes", "--null", "--name-only", "--list",
		])).toString("utf8");
		return raw.split("\0").filter(Boolean).map((key) => {
			if (!safeText(key, 512)) throw new GitAdapterError("repository Git configuration key is invalid");
			return key;
		});
	}

	async #assertSafeMutationConfiguration(cwd: string): Promise<void> {
		const localKeys = await this.#localConfigKeys(cwd, "--local");
		const worktreeKeys = localKeys.some((key) => key.toLowerCase() === "extensions.worktreeconfig")
			? await this.#localConfigKeys(cwd, "--worktree")
			: [];
		const unsafe = [...new Set([...localKeys, ...worktreeKeys].filter(unsafeMutationConfigKey))].sort();
		if (unsafe.length > 0) {
			throw new GitAdapterError(`repository contains unsafe Git mutation configuration: ${unsafe.join(", ")}`);
		}
	}

	#runMutation(cwd: string, args: readonly string[]): Promise<Buffer> {
		return this.#run(cwd, args, true);
	}

	async #effectiveEndpointValues(cwd: string, push: boolean): Promise<string[] | undefined> {
		try {
			const raw = (await this.#run(cwd, ["remote", "get-url", ...(push ? ["--push"] : []), "--all", "origin"])).toString("utf8");
			const values = stripLineEnding(raw).split(/\r?\n/).filter((value) => value !== "");
			if (values.length === 0) throw new GitAdapterError("origin effective endpoint is missing");
			return values;
		} catch (error) {
			if (error instanceof GitCommandFailure && error.exitCode === 2) return undefined;
			throw error;
		}
	}

	async #effectiveRemote(cwd: string): Promise<EffectiveGitRemote> {
		const [fetchValues, pushValues] = await Promise.all([
			this.#effectiveEndpointValues(cwd, false),
			this.#effectiveEndpointValues(cwd, true),
		]);
		if (fetchValues === undefined && pushValues === undefined) {
			const identity = endpointIdentity("no-remote");
			return { fetch: { identity }, push: { identity } };
		}
		if (fetchValues === undefined || pushValues === undefined || fetchValues.length !== 1 || pushValues.length !== 1) {
			throw new GitAdapterError("origin effective fetch or push endpoint is missing or ambiguous");
		}
		const [fetchNormalized, pushNormalized] = await Promise.all([
			normalizeEffectiveEndpoint(fetchValues[0], cwd),
			normalizeEffectiveEndpoint(pushValues[0], cwd),
		]);
		return {
			fetch: { value: fetchValues[0], identity: endpointIdentity(fetchNormalized) },
			push: { value: pushValues[0], identity: endpointIdentity(pushNormalized) },
		};
	}

	async #localDefaultBranch(cwd: string): Promise<string | undefined> {
		let raw: string;
		try {
			raw = stripLineEnding((await this.#run(cwd, [
				"symbolic-ref", "--quiet", "--short", "refs/remotes/origin/HEAD",
			])).toString("utf8"));
		} catch (error) {
			if (error instanceof GitCommandFailure && (error.exitCode === 1 || error.exitCode === 128)) return undefined;
			throw error;
		}
		if (!raw.startsWith("origin/")) throw new GitAdapterError("origin symbolic HEAD is malformed");
		const branch = raw.slice("origin/".length);
		assertSafeBranch(branch, "origin default branch");
		return branch;
	}

	async #remoteDefaultBranch(cwd: string, endpoint: string): Promise<string> {
		const raw = stripLineEnding((await this.#runMutation(cwd, [
			"ls-remote", "--symref", "--", endpoint, "HEAD",
		])).toString("utf8"));
		const symbolic = raw.split("\n").filter((line) => line.startsWith("ref: "));
		if (symbolic.length !== 1) throw new GitAdapterError("remote symbolic HEAD evidence is missing or ambiguous");
		const [reference, name, ...extra] = symbolic[0].slice("ref: ".length).split("\t");
		if (extra.length > 0 || name !== "HEAD" || !reference.startsWith("refs/heads/")) {
			throw new GitAdapterError("remote symbolic HEAD evidence is malformed");
		}
		const branch = reference.slice("refs/heads/".length);
		assertSafeBranch(branch, "remote default branch");
		return branch;
	}

	async #assertEndpointRewriteStable(cwd: string, endpoint: string, push: boolean): Promise<void> {
		let raw: string;
		try {
			raw = (await this.#run(cwd, [
				"config", "--null", "--get-regexp", "^url\\..*\\.(insteadof|pushinsteadof)$",
			])).toString("utf8");
		} catch (error) {
			if (error instanceof GitCommandFailure && error.exitCode === 1) return;
			throw error;
		}
		for (const record of raw.split("\0").filter(Boolean)) {
			const separator = record.indexOf("\n");
			if (separator < 1) throw new GitAdapterError("Git returned malformed URL rewrite configuration");
			const key = record.slice(0, separator).toLowerCase();
			const prefix = record.slice(separator + 1);
			if (!safeText(prefix, MAX_PATH_BYTES)) throw new GitAdapterError("Git URL rewrite prefix is invalid");
			if (!push && key.endsWith(".pushinsteadof")) continue;
			if (endpoint.startsWith(prefix)) {
				throw new GitAdapterError("origin effective endpoint is not stable under Git URL rewrite rules");
			}
		}
	}

	#leaseState(capability: GitMutationLease): MutationLeaseState {
		if (typeof capability !== "object" || capability === null) {
			throw new GitAdapterError("Git mutation requires an adapter-issued active lease capability");
		}
		const state = this.#mutationLeases.get(capability);
		if (state === undefined) throw new GitAdapterError("Git mutation lease capability was not issued by this adapter");
		return state;
	}

	async #assertCapabilityOwned(capability: GitMutationLease): Promise<void> {
		const state = this.#leaseState(capability);
		if (!state.accepting) throw new GitAdapterError("Git mutation lease was released");
		try {
			await state.runLease.assertOwned();
		} catch (error) {
			throw new GitAdapterError("Git mutation lease ownership was lost", { cause: error });
		}
	}

	#releaseCapability(capability: GitMutationLease): Promise<void> {
		const state = this.#leaseState(capability);
		if (state.releasePromise !== undefined) return state.releasePromise;
		state.accepting = false;
		const acceptedMutations = state.tail;
		state.releasePromise = (async () => {
			await acceptedMutations;
			try {
				await state.runLease.release();
			} catch (error) {
				throw new GitAdapterError("failed to release Git mutation lease", { cause: error });
			}
		})();
		return state.releasePromise;
	}

	#enqueueMutation<T>(capability: GitMutationLease, operation: (state: MutationLeaseState) => Promise<T>): Promise<T> {
		let state: MutationLeaseState;
		try {
			state = this.#leaseState(capability);
			if (!state.accepting) throw new GitAdapterError("Git mutation lease was released");
		} catch (error) {
			return Promise.reject(error);
		}
		const result = state.tail.then(() => operation(state));
		state.tail = result.then(() => undefined, () => undefined);
		return result;
	}

	async #assertLeaseOwnedForMutation(state: MutationLeaseState): Promise<void> {
		try {
			await state.runLease.assertOwned();
		} catch (error) {
			throw new GitAdapterError("Git mutation lease ownership was lost before mutation", { cause: error });
		}
	}

	#assertLeaseRequest(state: MutationLeaseState, issue: number, slug: string, branch: string): void {
		assertCanonicalIssueBranch(issue, slug, branch);
		if (issue !== state.issue || slug !== state.slug || branch !== state.branch) {
			throw new GitAdapterError("Git mutation request does not match its immutable lease claim");
		}
	}

	async #assertMutationBinding(state: MutationLeaseState, binding: GitBinding, target: "coordinator" | "worktree"): Promise<GitBinding> {
		const expectedCwd = target === "coordinator" ? state.coordinator.cwd : state.targetCwd;
		if (binding.cwd !== expectedCwd
			|| binding.repositoryIdentity !== state.coordinator.repositoryIdentity
			|| binding.remoteIdentity !== state.coordinator.remoteIdentity
			|| binding.fetchEndpointIdentity !== state.coordinator.fetchEndpointIdentity
			|| binding.pushEndpointIdentity !== state.coordinator.pushEndpointIdentity
			|| binding.defaultBranch !== state.coordinator.defaultBranch) {
			throw new GitAdapterError(`Git mutation ${target} binding does not match its immutable lease claim`);
		}
		const actual = await this.assertBinding(binding);
		if (target === "coordinator") {
			if (actual.worktreeIdentity !== state.coordinator.worktreeIdentity) {
				throw new GitAdapterError("Git mutation coordinator worktree identity changed");
			}
		} else if (state.targetWorktreeIdentity === undefined) {
			state.targetWorktreeIdentity = actual.worktreeIdentity;
		} else if (actual.worktreeIdentity !== state.targetWorktreeIdentity) {
			throw new GitAdapterError("Git mutation target worktree identity changed");
		}
		return actual;
	}

	async #acquireMutationLease(
		binding: GitBinding,
		request: GitMutationLeaseRequest,
		options: Omit<FileStateStoreOptions, "trustedRoot"> = {},
	): Promise<GitMutationLease> {
		if (typeof request !== "object" || request === null) throw new GitAdapterError("Git mutation lease request is required");
		assertCanonicalIssueBranch(request.issue, request.slug, request.branch);
		assertSha(request.baseHead, "mutation lease base head");
		const allowedScopes = validateScopes(request.allowedScopes);
		if (!safeText(request.targetCwd, MAX_PATH_BYTES) || !isAbsolute(request.targetCwd) || resolve(request.targetCwd) !== request.targetCwd) {
			throw new GitAdapterError("Git mutation target must be a canonical absolute path");
		}
		if (!safeText(request.stateRoot, MAX_PATH_BYTES) || !isAbsolute(request.stateRoot)
			|| await realpath(request.stateRoot) !== request.stateRoot) {
			throw new GitAdapterError("Git mutation lease state root must be a canonical existing path");
		}
		if (!safeText(request.runId, 256)) throw new GitAdapterError("Git mutation lease run ID must be bounded safe text");
		if (request.mode !== "start" && request.mode !== "resume") throw new GitAdapterError("Git mutation lease mode must be start or resume");
		const coordinator = await this.assertBinding(binding);
		const runLease: RunLease = await new FileStateStore(join(request.stateRoot, "leases", `issue-${request.issue}`), {
			...options,
			trustedRoot: request.stateRoot,
		}).acquireLease({ issue: request.issue, runId: request.runId, mode: request.mode });
		const state: MutationLeaseState = {
			runLease,
			coordinator,
			targetCwd: request.targetCwd,
			issue: request.issue,
			slug: request.slug,
			branch: request.branch,
			baseHead: request.baseHead,
			allowedScopes,
			accepting: true,
			tail: Promise.resolve(),
		};
		let capability: GitMutationLease;
		capability = Object.freeze({
			assertOwned: () => this.#assertCapabilityOwned(capability),
			release: () => this.#releaseCapability(capability),
		});
		this.#mutationLeases.set(capability, state);
		return capability;
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
		let rawRemote = "";
		try {
			rawRemote = (await this.#run(repositoryRoot, ["config", "--local", "--no-includes", "--get", "remote.origin.url"])).toString("utf8");
		} catch (error) {
			if (!(error instanceof GitCommandFailure) || error.exitCode !== 1) {
				throw new GitAdapterError("origin remote identity is missing or invalid", { cause: error });
			}
		}
		const normalizedRemote = normalizeRemote(rawRemote);
		const [repositoryFilesystemIdentity, worktreeFilesystemIdentity, effectiveRemote] = await Promise.all([
			filesystemIdentity(commonDirectory),
			filesystemIdentity(worktreeDirectory),
			this.#effectiveRemote(repositoryRoot),
		]);
		const defaultBranch = effectiveRemote.fetch.value === undefined
			? undefined
			: await this.#localDefaultBranch(repositoryRoot);
		const remoteIdentity = hashIdentity(["shepherd-origin-v1", normalizedRemote]);
		const repositoryIdentity = hashIdentity(["shepherd-repository-v1", repositoryFilesystemIdentity, normalizedRemote]);
		const worktreeIdentity = hashIdentity(["shepherd-worktree-v1", repositoryIdentity, worktreeFilesystemIdentity]);
		return {
			cwd: repositoryRoot,
			repositoryIdentity,
			worktreeIdentity,
			remoteName: "origin",
			remoteIdentity,
			fetchEndpointIdentity: effectiveRemote.fetch.identity,
			pushEndpointIdentity: effectiveRemote.push.identity,
			...(defaultBranch === undefined ? {} : { defaultBranch }),
		};
	}

	async assertBinding(binding: GitBinding): Promise<GitBinding> {
		if (typeof binding !== "object" || binding === null) throw new GitAdapterError("Git binding is required");
		assertIdentity(binding.repositoryIdentity, "repository identity");
		assertIdentity(binding.worktreeIdentity, "worktree identity");
		assertIdentity(binding.remoteIdentity, "remote identity");
		assertIdentity(binding.fetchEndpointIdentity, "fetch endpoint identity");
		assertIdentity(binding.pushEndpointIdentity, "push endpoint identity");
		if (binding.defaultBranch !== undefined) assertSafeBranch(binding.defaultBranch, "default branch");
		if (binding.remoteName !== "origin") throw new GitAdapterError("only the origin remote is supported");
		const actual = await this.inspect(binding.cwd);
		if (actual.repositoryIdentity !== binding.repositoryIdentity) throw new GitAdapterError("repository identity mismatch");
		if (actual.worktreeIdentity !== binding.worktreeIdentity) throw new GitAdapterError("worktree identity mismatch");
		if (actual.remoteIdentity !== binding.remoteIdentity) throw new GitAdapterError("origin remote identity mismatch");
		if (actual.fetchEndpointIdentity !== binding.fetchEndpointIdentity) throw new GitAdapterError("origin fetch endpoint mismatch");
		if (actual.pushEndpointIdentity !== binding.pushEndpointIdentity) throw new GitAdapterError("origin push endpoint mismatch");
		if (actual.defaultBranch !== binding.defaultBranch) throw new GitAdapterError("origin default branch mismatch");
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

	async fetchBranch(capability: GitMutationLease, binding: GitBinding, branch: string): Promise<string> {
		return this.#enqueueMutation(capability, async (state) => {
			assertSafeBranch(branch);
			if (branch !== state.branch) throw new GitAdapterError("fetched branch does not match the immutable lease claim");
			const actual = await this.#assertMutationBinding(state, binding, "coordinator");
			await this.#assertSafeMutationConfiguration(actual.cwd);
			const endpoint = (await this.#effectiveRemote(actual.cwd)).fetch;
			if (endpoint.value === undefined || endpoint.identity !== state.coordinator.fetchEndpointIdentity) {
				throw new GitAdapterError("origin fetch endpoint no longer matches the immutable lease claim");
			}
			await this.#assertEndpointRewriteStable(actual.cwd, endpoint.value, false);
			await this.#assertLeaseOwnedForMutation(state);
			await this.#runMutation(actual.cwd, ["fetch", "--no-tags", "--", endpoint.value, branch]);
			const fetched = stripLineEnding((await this.#run(actual.cwd, ["rev-parse", "--verify", "FETCH_HEAD^{commit}"])).toString("utf8"));
			assertSha(fetched, "fetched head");
			return fetched;
		});
	}

	async addIssueWorktree(capability: GitMutationLease, binding: GitBinding, request: GitAddWorktreeRequest): Promise<GitBinding> {
		return this.#enqueueMutation(capability, async (state) => {
			this.#assertLeaseRequest(state, request.issue, request.slug, request.branch);
			assertSha(request.baseHead, "base head");
			if (request.baseHead !== state.baseHead) throw new GitAdapterError("worktree base does not match the immutable lease claim");
			const actual = await this.#assertMutationBinding(state, binding, "coordinator");
			await this.#assertSafeMutationConfiguration(actual.cwd);
			await this.#assertCommitObject(actual, request.baseHead, "base head");
			if (!safeText(request.trustedRoot, MAX_PATH_BYTES) || !isAbsolute(request.trustedRoot)) {
				throw new GitAdapterError("trusted worktree root must be an absolute bounded path");
			}
			const root = await realpath(request.trustedRoot);
			const expectedPath = resolve(root, canonicalIssueWorktreeName(request.issue, request.slug));
			if (request.path !== state.targetCwd || request.path !== expectedPath || relative(root, request.path).startsWith(`..${sep}`)) {
				throw new GitAdapterError("worktree path must be the canonical child bound by the immutable lease claim");
			}
			const branches = await this.listLocalBranches(actual);
			const existing = branches.find((candidate) => candidate.branch === request.branch);
			if (existing !== undefined && !(await this.isAncestor(actual, request.baseHead, existing.head))) {
				throw new GitAdapterError("existing issue branch does not descend from the exact base head");
			}
			try {
				await this.#assertLeaseOwnedForMutation(state);
				if (existing === undefined) {
					await this.#runMutation(actual.cwd, ["worktree", "add", "-b", request.branch, "--", request.path, request.baseHead]);
				} else {
					await this.#runMutation(actual.cwd, ["worktree", "add", "--", request.path, request.branch]);
				}
			} catch (error) {
				throw new GitAdapterError("typed Git worktree creation failed; existing state was preserved", { cause: error });
			}
			const created = await this.inspect(request.path);
			if (created.repositoryIdentity !== actual.repositoryIdentity
				|| created.remoteIdentity !== actual.remoteIdentity
				|| created.fetchEndpointIdentity !== actual.fetchEndpointIdentity
				|| created.pushEndpointIdentity !== actual.pushEndpointIdentity
				|| created.defaultBranch !== actual.defaultBranch) {
				throw new GitAdapterError("created worktree repository identity mismatch");
			}
			state.targetWorktreeIdentity = created.worktreeIdentity;
			if (await this.currentBranch(created) !== request.branch) throw new GitAdapterError("created worktree branch mismatch");
			const head = await this.resolveBranchHead(created, request.branch);
			if (!(await this.isAncestor(created, request.baseHead, head))) throw new GitAdapterError("created worktree lost its exact base ancestry");
			return created;
		});
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

	async #historyPaths(binding: GitBinding, baseHead: string, head: string): Promise<string[]> {
		await this.#assertCommitObject(binding, baseHead, "base head");
		await this.#assertCommitObject(binding, head, "head");
		if (!(await this.isAncestor(binding, baseHead, head))) {
			throw new GitAdapterError("immutable base head is not an ancestor of canonical head");
		}
		const actual = await this.assertBinding(binding);
		const raw = (await this.#run(actual.cwd, [
			"log", "--format=", "--name-only", "-z", "--no-renames", "--no-ext-diff", "--no-textconv",
			"--no-show-signature",
			"--full-history", "-m", `${baseHead}..${head}`, "--",
		])).toString("utf8");
		return [...new Set(raw.split("\0").filter(Boolean).map((path) => {
			validateScope(path);
			return path;
		}))].sort();
	}

	async #assertHistoryWithinScopes(
		binding: GitBinding,
		baseHead: string,
		head: string,
		scopes: readonly string[],
	): Promise<string[]> {
		const changedScope = await this.#historyPaths(binding, baseHead, head);
		const outside = changedScope.filter((path) => !pathWithinScope(path, scopes));
		if (outside.length > 0) {
			throw new GitAdapterError(`Git history contains paths outside immutable allowed scopes: ${outside.join(", ")}`);
		}
		return changedScope;
	}

	async diff(binding: GitBinding, request: { baseHead: string; head: string; scopes: readonly string[] }): Promise<GitDiffEvidence> {
		const scopes = validateScopes(request.scopes);
		const changedScope = await this.#assertHistoryWithinScopes(binding, request.baseHead, request.head, scopes);
		return { baseHead: request.baseHead, head: request.head, changedScope };
	}

	async commitIssueChanges(capability: GitMutationLease, binding: GitBinding, request: GitCommitRequest): Promise<GitCommitEvidence> {
		return this.#enqueueMutation(capability, async (state) => {
			this.#assertLeaseRequest(state, request.issue, request.slug, request.branch);
			assertSha(request.expectedHead, "expected head");
			if (!safeText(request.message, 512)) throw new GitAdapterError("commit message must be bounded safe text");
			const scopes = validateScopes(request.scopes);
			if (scopes.some((scope) => !pathWithinScope(scope, state.allowedScopes))) {
				throw new GitAdapterError("commit scopes exceed the immutable lease claim");
			}
			const actual = await this.#assertMutationBinding(state, binding, "worktree");
			await this.#assertSafeMutationConfiguration(actual.cwd);
			if (await this.currentBranch(actual) !== request.branch) throw new GitAdapterError("current branch does not match canonical issue branch");
			const previousHead = await this.resolveBranchHead(actual, request.branch);
			if (previousHead !== request.expectedHead) throw new GitAdapterError("stale expected head; commit was not attempted");
			await this.#assertHistoryWithinScopes(actual, state.baseHead, previousHead, state.allowedScopes);
			const status = await this.status(actual);
			const outside = status.entries.flatMap((entry) => [entry.path, ...(entry.originalPath ? [entry.originalPath] : [])])
				.filter((path) => !pathWithinScope(path, scopes));
			if (outside.length > 0) throw new GitAdapterError(`dirty or staged state exists outside declared scopes: ${outside.sort().join(", ")}`);
			if (status.clean) return { committed: false, previousHead, head: previousHead };
			await this.#assertLeaseOwnedForMutation(state);
			await this.#runMutation(actual.cwd, ["add", "-A", "--", ...scopes]);
			const staged = (await this.#run(actual.cwd, ["diff", "--cached", "--name-only", "-z"])).toString("utf8")
				.split("\0").filter(Boolean);
			if (staged.some((path) => !pathWithinScope(path, scopes))) {
				throw new GitAdapterError("staged state escaped declared scopes; state was preserved for inspection");
			}
			if (staged.length === 0) return { committed: false, previousHead, head: previousHead };
			if (await this.resolveBranchHead(actual, request.branch) !== previousHead) {
				throw new GitAdapterError("canonical issue head changed while staging; commit was not attempted");
			}
			await this.#assertLeaseOwnedForMutation(state);
			await this.#runMutation(actual.cwd, ["commit", "--no-verify", "-m", request.message]);
			const head = await this.resolveBranchHead(actual, request.branch);
			if (head === previousHead) throw new GitAdapterError("commit did not advance the exact head");
			await this.#assertHistoryWithinScopes(actual, state.baseHead, head, state.allowedScopes);
			return { committed: true, previousHead, head };
		});
	}

	async pushIssueBranch(capability: GitMutationLease, binding: GitBinding, request: GitPushRequest): Promise<GitPushEvidence> {
		return this.#enqueueMutation(capability, async (state) => {
			this.#assertLeaseRequest(state, request.issue, request.slug, request.branch);
			assertSafeBranch(request.defaultBranch, "default branch");
			assertSha(request.expectedHead, "expected head");
			if (state.coordinator.defaultBranch === undefined
				|| request.defaultBranch !== state.coordinator.defaultBranch) {
				throw new GitAdapterError("requested default branch does not match bound origin symbolic HEAD evidence");
			}
			if (["main", "master", "trunk", state.coordinator.defaultBranch].includes(request.branch)) {
				throw new GitAdapterError("direct default branch push is unavailable");
			}
			const actual = await this.#assertMutationBinding(state, binding, "worktree");
			await this.#assertSafeMutationConfiguration(actual.cwd);
			if (await this.currentBranch(actual) !== request.branch) throw new GitAdapterError("current branch does not match canonical issue branch");
			let head = await this.resolveBranchHead(actual, request.branch);
			if (head !== request.expectedHead) throw new GitAdapterError("stale expected head; push was not attempted");
			const endpoints = await this.#effectiveRemote(actual.cwd);
			if (endpoints.fetch.value === undefined || endpoints.push.value === undefined
				|| endpoints.fetch.identity !== state.coordinator.fetchEndpointIdentity
				|| endpoints.push.identity !== state.coordinator.pushEndpointIdentity
				|| endpoints.fetch.identity !== endpoints.push.identity) {
				throw new GitAdapterError("origin push endpoint does not match the inspected and bound fetch endpoint");
			}
			await this.#assertEndpointRewriteStable(actual.cwd, endpoints.push.value, true);
			const remoteDefaultBranch = await this.#remoteDefaultBranch(actual.cwd, endpoints.push.value);
			if (remoteDefaultBranch !== state.coordinator.defaultBranch) {
				throw new GitAdapterError("remote symbolic HEAD no longer matches the bound default branch");
			}
			head = await this.resolveBranchHead(actual, request.branch);
			if (head !== request.expectedHead) throw new GitAdapterError("canonical issue head changed before push");
			await this.#assertHistoryWithinScopes(actual, state.baseHead, head, state.allowedScopes);
			await this.#assertLeaseOwnedForMutation(state);
			await this.#runMutation(actual.cwd, [
				"push", "--porcelain", "--", endpoints.push.value, `${head}:refs/heads/${request.branch}`,
			]);
			const remote = stripLineEnding((await this.#runMutation(actual.cwd, [
				"ls-remote", "--heads", "--", endpoints.push.value, `refs/heads/${request.branch}`,
			])).toString("utf8"));
			const [remoteHead, remoteRef, ...extra] = remote.split("\t");
			if (extra.length > 0 || remoteHead !== head || remoteRef !== `refs/heads/${request.branch}`) {
				throw new GitAdapterError("remote exact-head verification failed after push");
			}
			return { branch: request.branch, head, remoteName: "origin" };
		});
	}
}
