import { createHash, randomUUID } from "node:crypto";
import { constants } from "node:fs";
import { link, lstat, mkdir, open, readdir, realpath, rm, stat } from "node:fs/promises";
import { basename, dirname, isAbsolute, join, relative, resolve, sep } from "node:path";

import {
	GitAdapter,
	GitAdapterError,
	canonicalGitScopes,
	canonicalIssueBranch,
	canonicalIssueWorktreeName,
	type GitBinding,
	type GitCommitEvidence,
	type GitCommitRequest,
	type GitMutationLease,
	type GitMutationLeaseRequest,
	type GitPushEvidence,
	type GitPushRequest,
	type GitReconcileIssueHeadEvidence,
	type GitRefreshIssueEvidence,
	type GitStatusEvidence,
} from "./git-adapter.ts";
import type { FileStateStoreOptions } from "./state-store.ts";

const CLAIM_DIRECTORY = ".shepherd-workspace-claims";
const CLAIM_SCHEMA_VERSION = 4;
const BINDING_SCHEMA_VERSION = 1;
const MAX_CLAIM_BYTES = 32_768;
const MAX_PATH_BYTES = 4_096;
const MAX_OWNERSHIP_BYTES = 256;
const SAFE_OWNERSHIP = /^[A-Za-z0-9](?:[A-Za-z0-9._:-]*[A-Za-z0-9])?$/;
const SAFE_PARENT_REF = /^(?!\/|.*(?:\.\.|\s|[~^:?*\\\[\]])|.*\/$)[A-Za-z0-9][A-Za-z0-9._\/-]{0,239}$/;
const SHA_PATTERN = /^[0-9a-f]{40}$/;
const IDENTITY_PATTERN = /^[0-9a-f]{64}$/;

export type GitAdapterMutationLeaseAcquirer = (
	binding: GitBinding,
	request: GitMutationLeaseRequest,
	options?: Omit<FileStateStoreOptions, "trustedRoot">,
) => Promise<GitMutationLease>;

const gitAdapterMutationLeaseAcquirers = new WeakMap<object, GitAdapterMutationLeaseAcquirer>();

/** @internal One-way registration: WorkspaceAdapter can retrieve authority, callers cannot. */
export function registerGitAdapterMutationLeaseAcquirer(
	adapter: object,
	acquire: GitAdapterMutationLeaseAcquirer,
): void {
	if (typeof adapter !== "object" || adapter === null || typeof acquire !== "function") {
		throw new WorkspaceAdapterError("Git adapter mutation lease registration is invalid");
	}
	if (gitAdapterMutationLeaseAcquirers.has(adapter)) {
		throw new WorkspaceAdapterError("Git adapter mutation lease authority is already registered");
	}
	gitAdapterMutationLeaseAcquirers.set(adapter, acquire);
}

export type VerificationState = "pending" | "passed" | "failed";

export interface WorkspaceClaimRequest {
	coordinator: GitBinding;
	trustedWorktreeRoot: string;
	issue: number;
	slug: string;
	parentIssue: number;
	/** Exact non-default parent branch from the immutable production plan. */
	parentBranch: string;
	parentHead: string;
	ownershipId: string;
	allowedScopes: readonly string[];
	leaseMode?: "start" | "resume";
}

export interface ClaimedWorkspace extends GitBinding {
	issue: number;
	slug: string;
	branch: string;
	prBase: string;
	baseHead: string;
	head: string;
	trustedWorktreeRoot: string;
	allowedScopes: readonly string[];
	claimId: string;
	reused: boolean;
	status: GitStatusEvidence;
	changedScope: string[];
	assertOwned(): Promise<void>;
	release(): Promise<void>;
}

export interface WorkspaceHandoffEvidence {
	issue: number;
	branch: string;
	prBase: string;
	baseHead: string;
	head: string;
	changedScope: string[];
	verificationState: VerificationState;
	repositoryIdentity: string;
	worktreeIdentity: string;
	dirty: boolean;
}

export interface WorkspaceParentRefreshRequest {
	previousParentHead: string;
	newParentHead: string;
	effectKey: string;
}

export interface WorkspaceParentRefreshEvidence extends GitRefreshIssueEvidence {
	verificationInvalidated: true;
	reviewInvalidated: true;
}

export interface WorkspaceChildHeadReconciliationRequest {
	previousHead: string;
	effectKey: string;
}

export interface WorkspaceChildHeadReconciliationEvidence extends GitReconcileIssueHeadEvidence {
	verificationInvalidated: true;
	reviewInvalidated: true;
	integrationInvalidated: true;
}

export interface WorkspaceAdapterOptions {
	leaseOptions?: Omit<FileStateStoreOptions, "trustedRoot">;
}

interface WorkspaceClaimRecord {
	schemaVersion: 4;
	issue: number;
	slug: string;
	branch: string;
	path: string;
	trustedWorktreeRoot: string;
	prBase: string;
	baseHead: string;
	allowedScopes: string[];
	repositoryIdentity: string;
	remoteIdentity: string;
	fetchEndpointIdentity: string;
	pushEndpointIdentity: string;
	defaultBranch: string;
	ownerHash: string;
	requestHash: string;
}

interface WorkspaceBindingRecord {
	schemaVersion: 1;
	claimId: string;
	worktreeIdentity: string;
}

interface WorkspaceRefreshRecord {
	schemaVersion: 1;
	claimId: string;
	effectKeyHash: string;
	previousBaseHead: string;
	baseHead: string;
	previousHead: string;
	head: string;
	outcome: "unchanged" | "rebased" | "reclaimed";
}

interface WorkspaceChildHeadRecord {
	schemaVersion: 1;
	claimId: string;
	effectKeyHash: string;
	branch: string;
	baseHead: string;
	previousHead: string;
	head: string;
	changedScope: string[];
}

interface ActiveClaimContext {
	claimPath: string;
	bindingPath: string;
	claim: WorkspaceClaimRecord;
	binding: WorkspaceBindingRecord;
	coordinator: GitBinding;
	lease: GitMutationLease;
	effectiveBaseHead: string;
}

export class WorkspaceAdapterError extends Error {
	constructor(message: string, options?: ErrorOptions) {
		super(message, options);
		this.name = "WorkspaceAdapterError";
	}
}

function safeText(value: unknown, maximum: number): value is string {
	return typeof value === "string"
		&& value.length > 0
		&& Buffer.byteLength(value) <= maximum
		&& !/[\u0000-\u001f\u007f-\u009f]/.test(value);
}

function validIssue(value: unknown): value is number {
	return Number.isSafeInteger(value) && (value as number) > 0 && (value as number) <= 2_147_483_647;
}

function hash(parts: readonly string[]): string {
	const digest = createHash("sha256");
	for (const part of parts) {
		digest.update(String(Buffer.byteLength(part)));
		digest.update(":");
		digest.update(part);
		digest.update(";");
	}
	return digest.digest("hex");
}

function fileErrorCode(error: unknown): string | undefined {
	if (typeof error !== "object" || error === null || !("code" in error)) return undefined;
	return typeof error.code === "string" ? error.code : undefined;
}

function isWithin(parent: string, candidate: string): boolean {
	const child = relative(parent, candidate);
	return child === "" || (!child.startsWith(`..${sep}`) && child !== "..");
}

function changedPaths(status: GitStatusEvidence): string[] {
	return [...new Set(status.entries.flatMap((entry) => [entry.path, ...(entry.originalPath ? [entry.originalPath] : [])]))].sort();
}

function equalStrings(left: readonly string[], right: readonly string[]): boolean {
	return left.length === right.length && left.every((value, index) => value === right[index]);
}

function assertVerificationState(value: unknown): asserts value is VerificationState {
	if (!new Set(["pending", "passed", "failed"]).has(value as string)) {
		throw new WorkspaceAdapterError("verification state is invalid");
	}
}

function assertOnlyFields(record: Record<string, unknown>, allowed: ReadonlySet<string>, description: string): void {
	if (Object.keys(record).some((key) => !allowed.has(key))) {
		throw new WorkspaceAdapterError(`${description} is malformed; existing state was preserved`);
	}
}

function parseJsonRecord(raw: string, description: string): Record<string, unknown> {
	let value: unknown;
	try {
		value = JSON.parse(raw);
	} catch (error) {
		throw new WorkspaceAdapterError(`${description} is malformed; existing state was preserved`, { cause: error });
	}
	if (typeof value !== "object" || value === null || Array.isArray(value)) {
		throw new WorkspaceAdapterError(`${description} is malformed; existing state was preserved`);
	}
	return value as Record<string, unknown>;
}

function parseClaim(raw: string): WorkspaceClaimRecord {
	const record = parseJsonRecord(raw, "workspace ownership claim");
	assertOnlyFields(record, new Set([
		"schemaVersion", "issue", "slug", "branch", "path", "trustedWorktreeRoot", "prBase", "baseHead",
		"allowedScopes", "repositoryIdentity", "remoteIdentity", "fetchEndpointIdentity", "pushEndpointIdentity",
		"defaultBranch", "ownerHash", "requestHash",
	]), "workspace ownership claim");
	let allowedScopes: string[];
	try {
		if (!Array.isArray(record.allowedScopes) || record.allowedScopes.some((scope) => typeof scope !== "string")) throw new Error();
		allowedScopes = canonicalGitScopes(record.allowedScopes as string[]);
		if (!equalStrings(allowedScopes, record.allowedScopes as string[])) throw new Error();
	} catch (error) {
		throw new WorkspaceAdapterError("workspace ownership claim is malformed; existing state was preserved", { cause: error });
	}
	if (record.schemaVersion !== CLAIM_SCHEMA_VERSION
		|| !validIssue(record.issue)
		|| !safeText(record.slug, 100)
		|| !safeText(record.branch, 240)
		|| !safeText(record.path, MAX_PATH_BYTES) || !isAbsolute(record.path)
		|| !safeText(record.trustedWorktreeRoot, MAX_PATH_BYTES) || !isAbsolute(record.trustedWorktreeRoot)
		|| !safeText(record.prBase, 240)
		|| typeof record.baseHead !== "string" || !SHA_PATTERN.test(record.baseHead)
		|| typeof record.repositoryIdentity !== "string" || !IDENTITY_PATTERN.test(record.repositoryIdentity)
		|| typeof record.remoteIdentity !== "string" || !IDENTITY_PATTERN.test(record.remoteIdentity)
		|| typeof record.fetchEndpointIdentity !== "string" || !IDENTITY_PATTERN.test(record.fetchEndpointIdentity)
		|| typeof record.pushEndpointIdentity !== "string" || !IDENTITY_PATTERN.test(record.pushEndpointIdentity)
		|| !safeText(record.defaultBranch, 240)
		|| typeof record.ownerHash !== "string" || !IDENTITY_PATTERN.test(record.ownerHash)
		|| typeof record.requestHash !== "string" || !IDENTITY_PATTERN.test(record.requestHash)) {
		throw new WorkspaceAdapterError("workspace ownership claim is malformed; existing state was preserved");
	}
	return { ...record, allowedScopes } as unknown as WorkspaceClaimRecord;
}

function parseBinding(raw: string): WorkspaceBindingRecord {
	const record = parseJsonRecord(raw, "workspace identity binding");
	assertOnlyFields(record, new Set(["schemaVersion", "claimId", "worktreeIdentity"]), "workspace identity binding");
	if (record.schemaVersion !== BINDING_SCHEMA_VERSION
		|| typeof record.claimId !== "string" || !IDENTITY_PATTERN.test(record.claimId)
		|| typeof record.worktreeIdentity !== "string" || !IDENTITY_PATTERN.test(record.worktreeIdentity)) {
		throw new WorkspaceAdapterError("workspace identity binding is malformed; existing state was preserved");
	}
	return record as unknown as WorkspaceBindingRecord;
}

function serializedClaim(record: WorkspaceClaimRecord): string {
	return JSON.stringify({
		schemaVersion: record.schemaVersion,
		issue: record.issue,
		slug: record.slug,
		branch: record.branch,
		path: record.path,
		trustedWorktreeRoot: record.trustedWorktreeRoot,
		prBase: record.prBase,
		baseHead: record.baseHead,
		allowedScopes: record.allowedScopes,
		repositoryIdentity: record.repositoryIdentity,
		remoteIdentity: record.remoteIdentity,
		fetchEndpointIdentity: record.fetchEndpointIdentity,
		pushEndpointIdentity: record.pushEndpointIdentity,
		defaultBranch: record.defaultBranch,
		ownerHash: record.ownerHash,
		requestHash: record.requestHash,
	});
}

function serializedBinding(record: WorkspaceBindingRecord): string {
	return JSON.stringify({
		schemaVersion: record.schemaVersion,
		claimId: record.claimId,
		worktreeIdentity: record.worktreeIdentity,
	});
}

function parseRefresh(raw: string): WorkspaceRefreshRecord {
	const record = parseJsonRecord(raw, "workspace parent refresh record");
	assertOnlyFields(record, new Set([
		"schemaVersion", "claimId", "effectKeyHash", "previousBaseHead", "baseHead",
		"previousHead", "head", "outcome",
	]), "workspace parent refresh record");
	if (record.schemaVersion !== 1
		|| typeof record.claimId !== "string" || !IDENTITY_PATTERN.test(record.claimId)
		|| typeof record.effectKeyHash !== "string" || !IDENTITY_PATTERN.test(record.effectKeyHash)
		|| typeof record.previousBaseHead !== "string" || !SHA_PATTERN.test(record.previousBaseHead)
		|| typeof record.baseHead !== "string" || !SHA_PATTERN.test(record.baseHead)
		|| typeof record.previousHead !== "string" || !SHA_PATTERN.test(record.previousHead)
		|| typeof record.head !== "string" || !SHA_PATTERN.test(record.head)
		|| !new Set(["unchanged", "rebased", "reclaimed"]).has(record.outcome as string)) {
		throw new WorkspaceAdapterError("workspace parent refresh record is malformed; existing state was preserved");
	}
	return record as unknown as WorkspaceRefreshRecord;
}

function serializedRefresh(record: WorkspaceRefreshRecord): string {
	return JSON.stringify({
		schemaVersion: record.schemaVersion,
		claimId: record.claimId,
		effectKeyHash: record.effectKeyHash,
		previousBaseHead: record.previousBaseHead,
		baseHead: record.baseHead,
		previousHead: record.previousHead,
		head: record.head,
		outcome: record.outcome,
	});
}

function parseChildHead(raw: string): WorkspaceChildHeadRecord {
	const record = parseJsonRecord(raw, "workspace child-head reconciliation record");
	assertOnlyFields(record, new Set([
		"schemaVersion", "claimId", "effectKeyHash", "branch", "baseHead", "previousHead", "head",
		"changedScope",
	]), "workspace child-head reconciliation record");
	let changedScope: string[];
	try {
		if (!Array.isArray(record.changedScope) || record.changedScope.some((path) => typeof path !== "string")) throw new Error();
		changedScope = record.changedScope.length === 0 ? [] : canonicalGitScopes(record.changedScope as string[]);
		if (!equalStrings(changedScope, record.changedScope as string[])) throw new Error();
	} catch (error) {
		throw new WorkspaceAdapterError("workspace child-head reconciliation record is malformed", { cause: error });
	}
	if (record.schemaVersion !== 1
		|| typeof record.claimId !== "string" || !IDENTITY_PATTERN.test(record.claimId)
		|| typeof record.effectKeyHash !== "string" || !IDENTITY_PATTERN.test(record.effectKeyHash)
		|| !safeText(record.branch, 240) || !SAFE_PARENT_REF.test(record.branch)
		|| typeof record.baseHead !== "string" || !SHA_PATTERN.test(record.baseHead)
		|| typeof record.previousHead !== "string" || !SHA_PATTERN.test(record.previousHead)
		|| typeof record.head !== "string" || !SHA_PATTERN.test(record.head)) {
		throw new WorkspaceAdapterError("workspace child-head reconciliation record is malformed");
	}
	return { ...record, changedScope } as unknown as WorkspaceChildHeadRecord;
}

function serializedChildHead(record: WorkspaceChildHeadRecord): string {
	return JSON.stringify({
		schemaVersion: record.schemaVersion,
		claimId: record.claimId,
		effectKeyHash: record.effectKeyHash,
		branch: record.branch,
		baseHead: record.baseHead,
		previousHead: record.previousHead,
		head: record.head,
		changedScope: record.changedScope,
	});
}

function claimIdentity(record: WorkspaceClaimRecord): string {
	return hash(["shepherd-workspace-claim-v4", serializedClaim(record)]);
}

async function readImmutable(path: string, description: string): Promise<string> {
	let handle: Awaited<ReturnType<typeof open>> | undefined;
	try {
		handle = await open(path, constants.O_RDONLY | constants.O_NOFOLLOW);
		const metadata = await handle.stat();
		if (!metadata.isFile() || metadata.size < 1 || metadata.size > MAX_CLAIM_BYTES || (metadata.mode & 0o777) !== 0o600) {
			throw new WorkspaceAdapterError(`${description} is not a bounded mode-0600 regular file`);
		}
		return await handle.readFile("utf8");
	} catch (error) {
		if (error instanceof WorkspaceAdapterError) throw error;
		throw new WorkspaceAdapterError(`failed to read ${description}; existing state was preserved`, { cause: error });
	} finally {
		await handle?.close().catch(() => undefined);
	}
}

async function publishImmutable(path: string, payload: string, description: string): Promise<boolean> {
	if (Buffer.byteLength(payload) > MAX_CLAIM_BYTES) throw new WorkspaceAdapterError(`${description} exceeds its byte limit`);
	const temporary = join(dirname(path), `.${basename(path)}.${process.pid}.${randomUUID()}.tmp`);
	let handle: Awaited<ReturnType<typeof open>> | undefined;
	try {
		handle = await open(
			temporary,
			constants.O_CREAT | constants.O_EXCL | constants.O_WRONLY | constants.O_NOFOLLOW,
			0o600,
		);
		await handle.chmod(0o600);
		await handle.writeFile(payload, "utf8");
		await handle.sync();
		await handle.close();
		handle = undefined;
		try {
			await link(temporary, path);
			return true;
		} catch (error) {
			if (fileErrorCode(error) === "EEXIST") return false;
			throw new WorkspaceAdapterError(`failed to publish ${description} atomically`, { cause: error });
		}
	} finally {
		await handle?.close().catch(() => undefined);
		await rm(temporary, { force: true }).catch(() => undefined);
	}
}

async function readExistingClaim(path: string): Promise<WorkspaceClaimRecord | undefined> {
	try {
		await lstat(path);
	} catch (error) {
		if (fileErrorCode(error) === "ENOENT") return undefined;
		throw new WorkspaceAdapterError("failed to inspect existing workspace ownership claim", { cause: error });
	}
	return parseClaim(await readImmutable(path, "workspace ownership claim"));
}

async function loadEffectiveBase(
	claimDirectory: string,
	issue: number,
	claimId: string,
	initialBaseHead: string,
): Promise<string> {
	const prefix = `issue-${issue}.refresh-`;
	const names = (await readdir(claimDirectory)).filter((name) => name.startsWith(prefix) && name.endsWith(".json"));
	if (names.length > 64) throw new WorkspaceAdapterError("workspace parent refresh history exceeds its bound");
	const remaining: WorkspaceRefreshRecord[] = [];
	for (const name of names.sort()) {
		const suffix = name.slice(prefix.length, -".json".length);
		if (!IDENTITY_PATTERN.test(suffix)) throw new WorkspaceAdapterError("workspace parent refresh filename is malformed");
		const record = parseRefresh(await readImmutable(join(claimDirectory, name), "workspace parent refresh record"));
		if (record.claimId !== claimId || record.effectKeyHash !== suffix) {
			throw new WorkspaceAdapterError("workspace parent refresh record does not match its immutable claim or effect key");
		}
		remaining.push(record);
	}
	let baseHead = initialBaseHead;
	while (remaining.length > 0) {
		const matching = remaining.filter((record) => record.previousBaseHead === baseHead);
		if (matching.length !== 1) {
			throw new WorkspaceAdapterError("workspace parent refresh history is disconnected or ambiguous");
		}
		const next = matching[0];
		if (next.baseHead === baseHead && next.outcome !== "unchanged") {
			throw new WorkspaceAdapterError("workspace parent refresh history contains an invalid no-op transition");
		}
		baseHead = next.baseHead;
		remaining.splice(remaining.indexOf(next), 1);
	}
	return baseHead;
}

async function acquireRefresh(path: string, expected: WorkspaceRefreshRecord): Promise<void> {
	await publishImmutable(path, `${serializedRefresh(expected)}\n`, "workspace parent refresh record");
	const current = parseRefresh(await readImmutable(path, "workspace parent refresh record"));
	if (serializedRefresh(current) !== serializedRefresh(expected)) {
		throw new WorkspaceAdapterError("workspace parent refresh retry does not match its exact prepared effect");
	}
}

async function acquireChildHead(path: string, expected: WorkspaceChildHeadRecord): Promise<void> {
	await publishImmutable(path, `${serializedChildHead(expected)}\n`, "workspace child-head reconciliation record");
	const current = parseChildHead(await readImmutable(path, "workspace child-head reconciliation record"));
	if (serializedChildHead(current) !== serializedChildHead(expected)) {
		throw new WorkspaceAdapterError("workspace child-head reconciliation retry does not match its exact prepared effect");
	}
}

async function acquireClaim(path: string, expected: WorkspaceClaimRecord): Promise<WorkspaceClaimRecord> {
	await publishImmutable(path, `${serializedClaim(expected)}\n`, "workspace ownership claim");
	const current = parseClaim(await readImmutable(path, "workspace ownership claim"));
	if (current.issue !== expected.issue
		|| current.slug !== expected.slug
		|| current.branch !== expected.branch
		|| current.path !== expected.path
		|| current.trustedWorktreeRoot !== expected.trustedWorktreeRoot
		|| current.repositoryIdentity !== expected.repositoryIdentity
		|| current.remoteIdentity !== expected.remoteIdentity
		|| current.fetchEndpointIdentity !== expected.fetchEndpointIdentity
		|| current.pushEndpointIdentity !== expected.pushEndpointIdentity
		|| current.defaultBranch !== expected.defaultBranch) {
		throw new WorkspaceAdapterError("issue has aliased or mismatched workspace ownership; existing state was preserved");
	}
	if (current.ownerHash !== expected.ownerHash) {
		throw new WorkspaceAdapterError("canonical issue workspace is already owned by another mutator");
	}
	if (current.requestHash !== expected.requestHash || serializedClaim(current) !== serializedClaim(expected)) {
		throw new WorkspaceAdapterError("workspace retry does not match its original exact base or scope binding");
	}
	return current;
}

async function acquireBinding(path: string, expected: WorkspaceBindingRecord): Promise<WorkspaceBindingRecord> {
	await publishImmutable(path, `${serializedBinding(expected)}\n`, "workspace identity binding");
	const current = parseBinding(await readImmutable(path, "workspace identity binding"));
	if (serializedBinding(current) !== serializedBinding(expected)) {
		throw new WorkspaceAdapterError("workspace identity no longer matches its immutable original claim");
	}
	return current;
}

export class WorkspaceAdapter {
	readonly #git: GitAdapter;
	readonly #acquireMutationLease: GitAdapterMutationLeaseAcquirer;
	readonly #leaseOptions: Omit<FileStateStoreOptions, "trustedRoot">;
	readonly #activeClaims = new WeakMap<ClaimedWorkspace, ActiveClaimContext>();

	constructor(git: GitAdapter, options: WorkspaceAdapterOptions = {}) {
		this.#git = git;
		const acquireMutationLease = gitAdapterMutationLeaseAcquirers.get(git);
		if (acquireMutationLease === undefined) {
			throw new WorkspaceAdapterError("workspace adapter requires a genuine registered Git adapter");
		}
		this.#acquireMutationLease = acquireMutationLease;
		if (typeof options !== "object" || options === null
			|| (options.leaseOptions !== undefined && (typeof options.leaseOptions !== "object" || options.leaseOptions === null))) {
			throw new WorkspaceAdapterError("workspace adapter options are invalid");
		}
		this.#leaseOptions = { ...(options.leaseOptions ?? {}) };
	}

	async #acquireLease(
		coordinator: GitBinding,
		claimDirectory: string,
		target: string,
		branch: string,
		allowedScopes: readonly string[],
		request: WorkspaceClaimRequest,
	): Promise<GitMutationLease> {
		const mode = request.leaseMode ?? "start";
		if (mode !== "start" && mode !== "resume") throw new WorkspaceAdapterError("workspace lease mode must be start or resume");
		try {
			return await this.#acquireMutationLease(coordinator, {
				issue: request.issue,
				slug: request.slug,
				branch,
				baseHead: request.parentHead,
				targetCwd: target,
				allowedScopes,
				stateRoot: claimDirectory,
				runId: request.ownershipId,
				mode,
			}, this.#leaseOptions);
		} catch (error) {
			const message = error instanceof Error ? error.message : "";
			if (/held by live process/i.test(message)) {
				throw new WorkspaceAdapterError("canonical issue workspace is already owned by an active writable mutator", { cause: error });
			}
			if (/stale.*resume/i.test(message)) {
				throw new WorkspaceAdapterError("workspace mutator lease is stale; use leaseMode resume for explicit crash recovery", { cause: error });
			}
			throw new WorkspaceAdapterError("failed to acquire exclusive workspace mutator lease", { cause: error });
		}
	}

	async claim(request: WorkspaceClaimRequest): Promise<ClaimedWorkspace> {
		if (typeof request !== "object" || request === null) throw new WorkspaceAdapterError("workspace claim request is required");
		const branch = canonicalIssueBranch(request.issue, request.slug);
		if (typeof request.parentBranch !== "string" || !SAFE_PARENT_REF.test(request.parentBranch)) {
			throw new WorkspaceAdapterError("parent branch must be an exact bounded Git ref");
		}
		const prBase = request.parentBranch;
		if (request.issue === request.parentIssue) throw new WorkspaceAdapterError("issue and parent issue must be distinct");
		if (typeof request.parentHead !== "string" || !SHA_PATTERN.test(request.parentHead)) {
			throw new WorkspaceAdapterError("parent head must be an exact lowercase commit SHA");
		}
		if (!safeText(request.ownershipId, MAX_OWNERSHIP_BYTES) || !SAFE_OWNERSHIP.test(request.ownershipId)) {
			throw new WorkspaceAdapterError("workspace ownership ID must be bounded safe text");
		}
		let allowedScopes: string[];
		try {
			allowedScopes = canonicalGitScopes(request.allowedScopes);
		} catch (error) {
			throw new WorkspaceAdapterError("workspace requires bounded safe allowed scopes", { cause: error });
		}
		const coordinator = await this.#git.assertBinding(request.coordinator);
		if (coordinator.defaultBranch === undefined) {
			throw new WorkspaceAdapterError("workspace requires bound origin default branch evidence");
		}
		if (!safeText(request.trustedWorktreeRoot, MAX_PATH_BYTES) || !isAbsolute(request.trustedWorktreeRoot)) {
			throw new WorkspaceAdapterError("trusted worktree root must be an absolute bounded path");
		}
		const rootLinkMetadata = await lstat(request.trustedWorktreeRoot);
		if (!rootLinkMetadata.isDirectory() || rootLinkMetadata.isSymbolicLink()) {
			throw new WorkspaceAdapterError("trusted worktree root must be a real directory, not a symlink");
		}
		const trustedRoot = await realpath(request.trustedWorktreeRoot);
		const rootMetadata = await stat(trustedRoot);
		const target = resolve(trustedRoot, canonicalIssueWorktreeName(request.issue, request.slug));
		if (!isWithin(trustedRoot, target) || isWithin(coordinator.cwd, target) || isWithin(target, coordinator.cwd)) {
			throw new WorkspaceAdapterError("canonical worktree must be contained by its trusted root and isolated from the coordinator checkout");
		}

		let parentHead: string;
		try {
			parentHead = await this.#git.resolveBranchHead(coordinator, prBase);
		} catch (error) {
			throw new WorkspaceAdapterError("canonical parent branch is not present", { cause: error });
		}
		if (parentHead !== request.parentHead) throw new WorkspaceAdapterError("parent head mismatch; workspace creation was not attempted");

		const claimDirectory = join(trustedRoot, CLAIM_DIRECTORY);
		await mkdir(claimDirectory, { recursive: true, mode: 0o700 });
		const claimDirectoryMetadata = await lstat(claimDirectory);
		if (!claimDirectoryMetadata.isDirectory()
			|| claimDirectoryMetadata.isSymbolicLink()
			|| await realpath(claimDirectory) !== claimDirectory
			|| (claimDirectoryMetadata.mode & 0o077) !== 0) {
			throw new WorkspaceAdapterError("workspace claim directory is unsafe");
		}
		const lease = await this.#acquireLease(coordinator, claimDirectory, target, branch, allowedScopes, request);
		try {
			const branches = await this.#git.listLocalBranches(coordinator);
			const issuePrefix = `feat/${request.issue}-`;
			const aliases = branches.filter((candidate) => candidate.branch.startsWith(issuePrefix) && candidate.branch !== branch);
			if (aliases.length > 0) {
				throw new WorkspaceAdapterError(`aliased branch ownership detected: ${aliases.map((candidate) => candidate.branch).sort().join(", ")}`);
			}
			const inventory = await this.#git.listWorktrees(coordinator);
			const branchOwners = inventory.filter((entry) => entry.branch === branch);
			if (branchOwners.length > 1) throw new WorkspaceAdapterError("canonical issue branch has duplicate active worktree owners");
			const targetOwner = inventory.find((entry) => resolve(entry.cwd) === target);
			if (targetOwner !== undefined && targetOwner.branch !== branch) {
				throw new WorkspaceAdapterError("canonical worktree path is owned by another branch; existing state was preserved");
			}
			if (branchOwners.length === 1 && resolve(branchOwners[0].cwd) !== target) {
				throw new WorkspaceAdapterError("canonical issue branch is already active in another worktree");
			}
			if (targetOwner === undefined) {
				try {
					await lstat(target);
					throw new WorkspaceAdapterError("worktree path collision contains unique state; existing state was preserved");
				} catch (error) {
					if (error instanceof WorkspaceAdapterError) throw error;
					if (fileErrorCode(error) !== "ENOENT") {
						throw new WorkspaceAdapterError("failed to inspect canonical worktree path", { cause: error });
					}
				}
			}

			const claimPath = join(claimDirectory, `issue-${request.issue}.json`);
			const existingClaim = await readExistingClaim(claimPath);
			const initialBaseHead = existingClaim?.baseHead ?? request.parentHead;
			const ownerHash = hash(["shepherd-workspace-owner-v1", request.ownershipId]);
			const requestHash = hash([
				"shepherd-workspace-request-v4",
				branch,
				target,
				prBase,
				initialBaseHead,
				allowedScopes.join("\0"),
				coordinator.fetchEndpointIdentity,
				coordinator.pushEndpointIdentity,
				coordinator.defaultBranch,
			]);
			const requestedClaim: WorkspaceClaimRecord = {
				schemaVersion: CLAIM_SCHEMA_VERSION,
				issue: request.issue,
				slug: request.slug,
				branch,
				path: target,
				trustedWorktreeRoot: trustedRoot,
				prBase,
				baseHead: initialBaseHead,
				allowedScopes,
				repositoryIdentity: coordinator.repositoryIdentity,
				remoteIdentity: coordinator.remoteIdentity,
				fetchEndpointIdentity: coordinator.fetchEndpointIdentity,
				pushEndpointIdentity: coordinator.pushEndpointIdentity,
				defaultBranch: coordinator.defaultBranch,
				ownerHash,
				requestHash,
			};
			const persistedClaim = await acquireClaim(claimPath, requestedClaim);
			const claimId = claimIdentity(persistedClaim);
			const effectiveBaseHead = await loadEffectiveBase(
				claimDirectory,
				request.issue,
				claimId,
				persistedClaim.baseHead,
			);
			if (effectiveBaseHead !== request.parentHead) {
				throw new WorkspaceAdapterError("workspace resume parent base does not match its durable typed refresh history");
			}

			const rootAfterClaim = await stat(trustedRoot);
			if (rootAfterClaim.dev !== rootMetadata.dev || rootAfterClaim.ino !== rootMetadata.ino) {
				throw new WorkspaceAdapterError("trusted worktree root identity changed during claim");
			}

			let reused = targetOwner !== undefined;
			let binding: GitBinding;
			if (targetOwner !== undefined) {
				binding = await this.#git.inspect(target);
			} else {
				try {
					binding = await this.#git.addIssueWorktree(lease, coordinator, {
						trustedRoot,
						path: target,
						issue: request.issue,
						slug: request.slug,
						branch,
						baseHead: request.parentHead,
					});
				} catch (error) {
					// A dead process can leave the exact Git operation complete but the binding unpublished.
					// Reconcile that one exact outcome and never remove or rewrite unique state.
					const afterFailure = await this.#git.listWorktrees(coordinator);
					const exact = afterFailure.filter((entry) => entry.branch === branch && resolve(entry.cwd) === target);
					if (exact.length !== 1) {
						if (error instanceof GitAdapterError && /unsafe Git mutation configuration/i.test(error.message)) {
							throw new WorkspaceAdapterError("workspace rejected unsafe Git mutation configuration", { cause: error });
						}
						throw new WorkspaceAdapterError("workspace creation failed and no exact retry state exists", { cause: error });
					}
					binding = await this.#git.inspect(target);
					reused = true;
				}
			}

			if (binding.repositoryIdentity !== coordinator.repositoryIdentity
				|| binding.remoteIdentity !== coordinator.remoteIdentity
				|| binding.fetchEndpointIdentity !== coordinator.fetchEndpointIdentity
				|| binding.pushEndpointIdentity !== coordinator.pushEndpointIdentity
				|| binding.defaultBranch !== coordinator.defaultBranch) {
				throw new WorkspaceAdapterError("workspace repository identity mismatch");
			}
			if (await this.#git.currentBranch(binding) !== branch) throw new WorkspaceAdapterError("workspace branch ownership mismatch");
			const head = await this.#git.resolveBranchHead(binding, branch);
			if (!(await this.#git.isAncestor(binding, request.parentHead, head))) {
				throw new WorkspaceAdapterError("exact parent base is not an ancestor of workspace head");
			}
			const finalInventory = await this.#git.listWorktrees(coordinator);
			if (finalInventory.filter((entry) => entry.branch === branch).length !== 1) {
				throw new WorkspaceAdapterError("two active mutators were detected for the canonical issue branch");
			}
			const requestedBinding: WorkspaceBindingRecord = {
				schemaVersion: BINDING_SCHEMA_VERSION,
				claimId,
				worktreeIdentity: binding.worktreeIdentity,
			};
			const bindingPath = join(claimDirectory, `issue-${request.issue}.binding.json`);
			const persistedBinding = await acquireBinding(bindingPath, requestedBinding);
			const status = await this.#git.status(binding);
			const rootAfterCreation = await stat(trustedRoot);
			if (rootAfterCreation.dev !== rootMetadata.dev || rootAfterCreation.ino !== rootMetadata.ino) {
				throw new WorkspaceAdapterError("trusted worktree root identity changed during creation; created state was preserved");
			}

			let releasePromise: Promise<void> | undefined;
			const workspace: ClaimedWorkspace = {
				...binding,
				issue: request.issue,
				slug: request.slug,
				branch,
				prBase,
				baseHead: effectiveBaseHead,
				head,
				trustedWorktreeRoot: trustedRoot,
				allowedScopes: [...allowedScopes],
				claimId,
				reused,
				status,
				changedScope: changedPaths(status),
				assertOwned: () => lease.assertOwned(),
				release: () => releasePromise ??= lease.release(),
			};
			this.#activeClaims.set(workspace, {
				claimPath,
				bindingPath,
				claim: persistedClaim,
				binding: persistedBinding,
				coordinator,
				lease,
				effectiveBaseHead,
			});
			return workspace;
		} catch (error) {
			await lease.release().catch(() => undefined);
			throw error;
		}
	}

	#mutationContext(workspace: ClaimedWorkspace): ActiveClaimContext {
		const context = this.#activeClaims.get(workspace);
		if (context === undefined) {
			throw new WorkspaceAdapterError("workspace mutation requires an active claim issued by this adapter");
		}
		if (workspace.claimId !== context.binding.claimId
			|| workspace.issue !== context.claim.issue
			|| workspace.slug !== context.claim.slug
			|| workspace.branch !== context.claim.branch
			|| workspace.cwd !== context.claim.path
			|| workspace.trustedWorktreeRoot !== context.claim.trustedWorktreeRoot
			|| workspace.prBase !== context.claim.prBase
			|| workspace.baseHead !== context.effectiveBaseHead
			|| !equalStrings(workspace.allowedScopes, context.claim.allowedScopes)
			|| workspace.repositoryIdentity !== context.claim.repositoryIdentity
			|| workspace.remoteIdentity !== context.claim.remoteIdentity
			|| workspace.fetchEndpointIdentity !== context.claim.fetchEndpointIdentity
			|| workspace.pushEndpointIdentity !== context.claim.pushEndpointIdentity
			|| workspace.defaultBranch !== context.claim.defaultBranch
			|| workspace.worktreeIdentity !== context.binding.worktreeIdentity) {
			throw new WorkspaceAdapterError("workspace mutation does not match its immutable active claim");
		}
		return context;
	}

	#worktreeBinding(context: ActiveClaimContext): GitBinding {
		return {
			cwd: context.claim.path,
			repositoryIdentity: context.claim.repositoryIdentity,
			worktreeIdentity: context.binding.worktreeIdentity,
			remoteName: "origin",
			remoteIdentity: context.claim.remoteIdentity,
			fetchEndpointIdentity: context.claim.fetchEndpointIdentity,
			pushEndpointIdentity: context.claim.pushEndpointIdentity,
			defaultBranch: context.claim.defaultBranch,
		};
	}

	async commitIssueChanges(workspace: ClaimedWorkspace, request: GitCommitRequest): Promise<GitCommitEvidence> {
		const context = this.#mutationContext(workspace);
		return this.#git.commitIssueChanges(context.lease, this.#worktreeBinding(context), request);
	}

	async fetchBranch(workspace: ClaimedWorkspace, branch: string): Promise<string> {
		const context = this.#mutationContext(workspace);
		return this.#git.fetchBranch(context.lease, context.coordinator, branch);
	}

	async pushIssueBranch(workspace: ClaimedWorkspace, request: GitPushRequest): Promise<GitPushEvidence> {
		const context = this.#mutationContext(workspace);
		return this.#git.pushIssueBranch(context.lease, this.#worktreeBinding(context), request);
	}

	async refreshParent(
		workspace: ClaimedWorkspace,
		request: WorkspaceParentRefreshRequest,
	): Promise<WorkspaceParentRefreshEvidence> {
		const context = this.#mutationContext(workspace);
		if (typeof request !== "object" || request === null
			|| typeof request.previousParentHead !== "string" || !SHA_PATTERN.test(request.previousParentHead)
			|| typeof request.newParentHead !== "string" || !SHA_PATTERN.test(request.newParentHead)
			|| !safeText(request.effectKey, MAX_OWNERSHIP_BYTES) || !SAFE_OWNERSHIP.test(request.effectKey)
			|| request.previousParentHead !== context.effectiveBaseHead) {
			throw new WorkspaceAdapterError("parent refresh request does not match the active exact workspace base");
		}
		const parentHead = await this.#git.resolveBranchHead(context.coordinator, context.claim.prBase);
		if (parentHead !== request.newParentHead) {
			throw new WorkspaceAdapterError("new parent head is not the authoritative local parent branch head");
		}
		const evidence = await this.#git.refreshIssueBranch(context.lease, this.#worktreeBinding(context), {
			issue: context.claim.issue,
			slug: context.claim.slug,
			branch: context.claim.branch,
			previousBaseHead: request.previousParentHead,
			newBaseHead: request.newParentHead,
			expectedHead: workspace.head,
		});
		if (evidence.baseHead !== evidence.previousBaseHead) {
			const effectKeyHash = hash(["shepherd-workspace-refresh-effect-v1", request.effectKey]);
			await acquireRefresh(
				join(dirname(context.claimPath), `issue-${context.claim.issue}.refresh-${effectKeyHash}.json`),
				{
					schemaVersion: 1,
					claimId: context.binding.claimId,
					effectKeyHash,
					previousBaseHead: evidence.previousBaseHead,
					baseHead: evidence.baseHead,
					previousHead: evidence.previousHead,
					head: evidence.head,
					outcome: evidence.outcome,
				},
			);
		}
		context.effectiveBaseHead = evidence.baseHead;
		workspace.baseHead = evidence.baseHead;
		workspace.head = evidence.head;
		workspace.status = await this.#git.status(this.#worktreeBinding(context));
		workspace.changedScope = changedPaths(workspace.status);
		return { ...evidence, verificationInvalidated: true, reviewInvalidated: true };
	}

	async reconcileChildHead(
		workspace: ClaimedWorkspace,
		request: WorkspaceChildHeadReconciliationRequest,
	): Promise<WorkspaceChildHeadReconciliationEvidence> {
		const context = this.#mutationContext(workspace);
		if (typeof request !== "object" || request === null
			|| typeof request.previousHead !== "string" || !SHA_PATTERN.test(request.previousHead)
			|| !safeText(request.effectKey, MAX_OWNERSHIP_BYTES) || !SAFE_OWNERSHIP.test(request.effectKey)) {
			throw new WorkspaceAdapterError("child-head reconciliation request is invalid");
		}
		const evidence = await this.#git.reconcileIssueBranchHead(context.lease, this.#worktreeBinding(context), {
			issue: context.claim.issue,
			slug: context.claim.slug,
			branch: context.claim.branch,
			baseHead: context.effectiveBaseHead,
			previousHead: request.previousHead,
		});
		const effectKeyHash = hash(["shepherd-workspace-child-head-effect-v1", request.effectKey]);
		await acquireChildHead(
			join(dirname(context.claimPath), `issue-${context.claim.issue}.child-head-${effectKeyHash}.json`),
			{
				schemaVersion: 1,
				claimId: context.binding.claimId,
				effectKeyHash,
				branch: evidence.branch,
				baseHead: evidence.baseHead,
				previousHead: evidence.previousHead,
				head: evidence.head,
				changedScope: evidence.changedScope,
			},
		);
		workspace.head = evidence.head;
		workspace.status = await this.#git.status(this.#worktreeBinding(context));
		workspace.changedScope = [...evidence.changedScope];
		return {
			...evidence,
			verificationInvalidated: true,
			reviewInvalidated: true,
			integrationInvalidated: true,
		};
	}

	async captureHandoff(workspace: ClaimedWorkspace, verificationState: VerificationState): Promise<WorkspaceHandoffEvidence> {
		assertVerificationState(verificationState);
		const context = this.#activeClaims.get(workspace);
		if (context === undefined) {
			throw new WorkspaceAdapterError("handoff workspace was not issued with an active immutable claim by this adapter");
		}
		await context.lease.assertOwned();
		const persistedClaim = parseClaim(await readImmutable(context.claimPath, "workspace ownership claim"));
		const persistedBinding = parseBinding(await readImmutable(context.bindingPath, "workspace identity binding"));
		if (serializedClaim(persistedClaim) !== serializedClaim(context.claim)
			|| serializedBinding(persistedBinding) !== serializedBinding(context.binding)
			|| claimIdentity(persistedClaim) !== persistedBinding.claimId) {
			throw new WorkspaceAdapterError("immutable persisted workspace claim changed before handoff");
		}
		if (workspace.claimId !== persistedBinding.claimId
			|| workspace.issue !== persistedClaim.issue
			|| workspace.slug !== persistedClaim.slug
			|| workspace.branch !== persistedClaim.branch
			|| workspace.cwd !== persistedClaim.path
			|| workspace.trustedWorktreeRoot !== persistedClaim.trustedWorktreeRoot
			|| workspace.prBase !== persistedClaim.prBase
			|| workspace.baseHead !== context.effectiveBaseHead
			|| !equalStrings(workspace.allowedScopes, persistedClaim.allowedScopes)
			|| workspace.repositoryIdentity !== persistedClaim.repositoryIdentity
			|| workspace.remoteIdentity !== persistedClaim.remoteIdentity
			|| workspace.fetchEndpointIdentity !== persistedClaim.fetchEndpointIdentity
			|| workspace.pushEndpointIdentity !== persistedClaim.pushEndpointIdentity
			|| workspace.defaultBranch !== persistedClaim.defaultBranch
			|| workspace.worktreeIdentity !== persistedBinding.worktreeIdentity) {
			throw new WorkspaceAdapterError("handoff evidence does not match its immutable persisted original claim");
		}
		const canonicalBranch = canonicalIssueBranch(persistedClaim.issue, persistedClaim.slug);
		if (persistedClaim.branch !== canonicalBranch) {
			throw new WorkspaceAdapterError("handoff branch is not the canonical branch for this issue");
		}
		const binding = await this.#git.assertBinding({
			cwd: persistedClaim.path,
			repositoryIdentity: persistedClaim.repositoryIdentity,
			worktreeIdentity: persistedBinding.worktreeIdentity,
			remoteName: "origin",
			remoteIdentity: persistedClaim.remoteIdentity,
			fetchEndpointIdentity: persistedClaim.fetchEndpointIdentity,
			pushEndpointIdentity: persistedClaim.pushEndpointIdentity,
			defaultBranch: persistedClaim.defaultBranch,
		});
		if (await this.#git.currentBranch(binding) !== canonicalBranch) {
			throw new WorkspaceAdapterError("handoff current branch is not canonical");
		}
		const head = await this.#git.resolveBranchHead(binding, canonicalBranch);
		if (!(await this.#git.isAncestor(binding, context.effectiveBaseHead, head))) {
			throw new WorkspaceAdapterError("handoff exact base is not an ancestor of exact head");
		}
		const [status, diff] = await Promise.all([
			this.#git.status(binding),
			this.#git.diff(binding, {
				baseHead: context.effectiveBaseHead,
				head,
				scopes: persistedClaim.allowedScopes,
			}),
		]);
		const dirtyPaths = changedPaths(status);
		const outsideDirty = dirtyPaths.filter((path) => !persistedClaim.allowedScopes.some(
			(scope) => path === scope || path.startsWith(`${scope}/`),
		));
		if (outsideDirty.length > 0) {
			throw new WorkspaceAdapterError(`handoff contains dirty paths outside immutable allowed scopes: ${outsideDirty.join(", ")}`);
		}
		const changedScope = [...new Set([...diff.changedScope, ...dirtyPaths])].sort();
		return {
			issue: persistedClaim.issue,
			branch: canonicalBranch,
			prBase: persistedClaim.prBase,
			baseHead: context.effectiveBaseHead,
			head,
			changedScope,
			verificationState,
			repositoryIdentity: binding.repositoryIdentity,
			worktreeIdentity: binding.worktreeIdentity,
			dirty: !status.clean,
		};
	}
}

export { GitAdapterError };
