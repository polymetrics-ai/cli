import { createHash } from "node:crypto";
import { lstat, mkdir, open, readFile, realpath, stat } from "node:fs/promises";
import { isAbsolute, join, relative, resolve, sep } from "node:path";

import {
	GitAdapter,
	GitAdapterError,
	canonicalGitScopes,
	canonicalIssueBranch,
	canonicalIssueWorktreeName,
	type GitBinding,
	type GitStatusEvidence,
} from "./git-adapter.ts";

const CLAIM_DIRECTORY = ".shepherd-workspace-claims";
const CLAIM_SCHEMA_VERSION = 1;
const MAX_PATH_BYTES = 4_096;
const MAX_OWNERSHIP_BYTES = 256;
const SAFE_OWNERSHIP = /^[A-Za-z0-9](?:[A-Za-z0-9._:-]*[A-Za-z0-9])?$/;
const SHA_PATTERN = /^[0-9a-f]{40}$/;
const IDENTITY_PATTERN = /^[0-9a-f]{64}$/;

export type VerificationState = "pending" | "passed" | "failed";

export interface WorkspaceClaimRequest {
	coordinator: GitBinding;
	trustedWorktreeRoot: string;
	issue: number;
	slug: string;
	parentIssue: number;
	parentSlug: string;
	parentHead: string;
	ownershipId: string;
	allowedScopes: readonly string[];
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
	reused: boolean;
	status: GitStatusEvidence;
	changedScope: string[];
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

interface WorkspaceClaimRecord {
	schemaVersion: 1;
	issue: number;
	branch: string;
	path: string;
	prBase: string;
	baseHead: string;
	repositoryIdentity: string;
	remoteIdentity: string;
	ownerHash: string;
	requestHash: string;
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

function assertVerificationState(value: unknown): asserts value is VerificationState {
	if (!new Set(["pending", "passed", "failed"]).has(value as string)) {
		throw new WorkspaceAdapterError("verification state is invalid");
	}
}

function parseClaim(raw: string): WorkspaceClaimRecord {
	let value: unknown;
	try {
		value = JSON.parse(raw);
	} catch (error) {
		throw new WorkspaceAdapterError("workspace ownership claim is malformed; existing state was preserved", { cause: error });
	}
	if (typeof value !== "object" || value === null || Array.isArray(value)) {
		throw new WorkspaceAdapterError("workspace ownership claim is malformed; existing state was preserved");
	}
	const record = value as Record<string, unknown>;
	const allowed = new Set([
		"schemaVersion", "issue", "branch", "path", "prBase", "baseHead",
		"repositoryIdentity", "remoteIdentity", "ownerHash", "requestHash",
	]);
	if (Object.keys(record).some((key) => !allowed.has(key))
		|| record.schemaVersion !== CLAIM_SCHEMA_VERSION
		|| !Number.isSafeInteger(record.issue)
		|| !safeText(record.branch, 240)
		|| !safeText(record.path, MAX_PATH_BYTES)
		|| !safeText(record.prBase, 240)
		|| typeof record.baseHead !== "string" || !SHA_PATTERN.test(record.baseHead)
		|| typeof record.repositoryIdentity !== "string" || !IDENTITY_PATTERN.test(record.repositoryIdentity)
		|| typeof record.remoteIdentity !== "string" || !IDENTITY_PATTERN.test(record.remoteIdentity)
		|| typeof record.ownerHash !== "string" || !IDENTITY_PATTERN.test(record.ownerHash)
		|| typeof record.requestHash !== "string" || !IDENTITY_PATTERN.test(record.requestHash)) {
		throw new WorkspaceAdapterError("workspace ownership claim is malformed; existing state was preserved");
	}
	return record as unknown as WorkspaceClaimRecord;
}

async function assertRegularClaim(path: string): Promise<void> {
	const metadata = await lstat(path);
	if (!metadata.isFile() || metadata.isSymbolicLink() || metadata.size > 16_384) {
		throw new WorkspaceAdapterError("workspace ownership claim is not a bounded regular file");
	}
}

async function acquireClaim(path: string, expected: WorkspaceClaimRecord): Promise<void> {
	let handle: Awaited<ReturnType<typeof open>> | undefined;
	try {
		handle = await open(path, "wx", 0o600);
		await handle.writeFile(`${JSON.stringify(expected)}\n`, "utf8");
		await handle.sync();
		return;
	} catch (error) {
		if (fileErrorCode(error) !== "EEXIST") throw new WorkspaceAdapterError("failed to acquire workspace ownership", { cause: error });
	} finally {
		await handle?.close();
	}
	await assertRegularClaim(path);
	let current: WorkspaceClaimRecord;
	try {
		current = parseClaim(await readFile(path, "utf8"));
	} catch (error) {
		// A concurrently created claim can be observed between O_EXCL creation and its bounded write.
		await new Promise<void>((resolvePromise) => setTimeout(resolvePromise, 5));
		await assertRegularClaim(path);
		current = parseClaim(await readFile(path, "utf8"));
	}
	if (current.issue !== expected.issue
		|| current.branch !== expected.branch
		|| current.path !== expected.path
		|| current.repositoryIdentity !== expected.repositoryIdentity
		|| current.remoteIdentity !== expected.remoteIdentity) {
		throw new WorkspaceAdapterError("issue has aliased or mismatched workspace ownership; existing state was preserved");
	}
	if (current.ownerHash !== expected.ownerHash) {
		throw new WorkspaceAdapterError("canonical issue workspace is already owned by another mutator");
	}
	if (current.requestHash !== expected.requestHash) {
		throw new WorkspaceAdapterError("workspace retry does not match its original exact base or scope binding");
	}
}

export class WorkspaceAdapter {
	readonly #git: GitAdapter;

	constructor(git: GitAdapter) {
		this.#git = git;
	}

	async claim(request: WorkspaceClaimRequest): Promise<ClaimedWorkspace> {
		if (typeof request !== "object" || request === null) throw new WorkspaceAdapterError("workspace claim request is required");
		const branch = canonicalIssueBranch(request.issue, request.slug);
		const prBase = canonicalIssueBranch(request.parentIssue, request.parentSlug);
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
				if (fileErrorCode(error) !== "ENOENT") throw new WorkspaceAdapterError("failed to inspect canonical worktree path", { cause: error });
			}
		}

		const claimDirectory = join(trustedRoot, CLAIM_DIRECTORY);
		await mkdir(claimDirectory, { recursive: true, mode: 0o700 });
		const claimDirectoryMetadata = await lstat(claimDirectory);
		if (!claimDirectoryMetadata.isDirectory() || claimDirectoryMetadata.isSymbolicLink() || await realpath(claimDirectory) !== claimDirectory) {
			throw new WorkspaceAdapterError("workspace claim directory is unsafe");
		}
		const ownerHash = hash(["shepherd-workspace-owner-v1", request.ownershipId]);
		const scopeBinding = allowedScopes.join("\0");
		const requestHash = hash(["shepherd-workspace-request-v1", branch, target, prBase, request.parentHead, scopeBinding]);
		const claim: WorkspaceClaimRecord = {
			schemaVersion: CLAIM_SCHEMA_VERSION,
			issue: request.issue,
			branch,
			path: target,
			prBase,
			baseHead: request.parentHead,
			repositoryIdentity: coordinator.repositoryIdentity,
			remoteIdentity: coordinator.remoteIdentity,
			ownerHash,
			requestHash,
		};
		await acquireClaim(join(claimDirectory, `issue-${request.issue}.json`), claim);

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
				binding = await this.#git.addIssueWorktree(coordinator, {
					trustedRoot,
					path: target,
					issue: request.issue,
					slug: request.slug,
					branch,
					baseHead: request.parentHead,
				});
			} catch (error) {
				// Reconcile the exact outcome of a crash or same-owner racing retry. Never remove state.
				const afterFailure = await this.#git.listWorktrees(coordinator);
				const exact = afterFailure.filter((entry) => entry.branch === branch && resolve(entry.cwd) === target);
				if (exact.length !== 1) throw new WorkspaceAdapterError("workspace creation failed and no exact retry state exists", { cause: error });
				binding = await this.#git.inspect(target);
				reused = true;
			}
		}

		if (binding.repositoryIdentity !== coordinator.repositoryIdentity || binding.remoteIdentity !== coordinator.remoteIdentity) {
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
		const status = await this.#git.status(binding);
		const rootAfterCreation = await stat(trustedRoot);
		if (rootAfterCreation.dev !== rootMetadata.dev || rootAfterCreation.ino !== rootMetadata.ino) {
			throw new WorkspaceAdapterError("trusted worktree root identity changed during creation; created state was preserved");
		}
		return {
			...binding,
			issue: request.issue,
			slug: request.slug,
			branch,
			prBase,
			baseHead: request.parentHead,
			head,
			trustedWorktreeRoot: trustedRoot,
			allowedScopes,
			reused,
			status,
			changedScope: changedPaths(status),
		};
	}

	async captureHandoff(workspace: ClaimedWorkspace, verificationState: VerificationState): Promise<WorkspaceHandoffEvidence> {
		assertVerificationState(verificationState);
		const canonicalBranch = canonicalIssueBranch(workspace.issue, workspace.slug);
		if (workspace.branch !== canonicalBranch) throw new WorkspaceAdapterError("handoff branch is not the canonical branch for this issue");
		const binding = await this.#git.assertBinding(workspace);
		if (binding.repositoryIdentity !== workspace.repositoryIdentity || binding.worktreeIdentity !== workspace.worktreeIdentity) {
			throw new WorkspaceAdapterError("handoff workspace identity mismatch");
		}
		if (await this.#git.currentBranch(binding) !== canonicalBranch) throw new WorkspaceAdapterError("handoff current branch is not canonical");
		const head = await this.#git.resolveBranchHead(binding, canonicalBranch);
		if (!(await this.#git.isAncestor(binding, workspace.baseHead, head))) {
			throw new WorkspaceAdapterError("handoff exact base is not an ancestor of exact head");
		}
		const [status, diff] = await Promise.all([
			this.#git.status(binding),
			this.#git.diff(binding, { baseHead: workspace.baseHead, head, scopes: workspace.allowedScopes }),
		]);
		const changedScope = [...new Set([...diff.changedScope, ...changedPaths(status)])].sort();
		return {
			issue: workspace.issue,
			branch: canonicalBranch,
			prBase: workspace.prBase,
			baseHead: workspace.baseHead,
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
