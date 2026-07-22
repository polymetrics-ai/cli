import { createHash } from "node:crypto";
import { lstat, mkdir, readFile, realpath, writeFile } from "node:fs/promises";
import { dirname, join, relative, resolve, sep } from "node:path";

import type { AgentSessionHandoff, RoleRunRequest } from "./agent-session-runtime.ts";
import type {
	ProductionChildSpec,
	ProductionWorkspaceBinding,
} from "./autonomous-production-contract.ts";
import { ProductionLifecycleError as LifecycleError } from "./autonomous-production-contract.ts";
import type {
	GitBinding,
	GitCommitEvidence,
	GitCommitRequest,
	GitPushEvidence,
	GitPushRequest,
} from "./git-adapter.ts";
import type {
	ClaimedWorkspace,
	WorkspaceClaimRequest,
	WorkspaceHandoffEvidence,
} from "./workspace-adapter.ts";
import type { ProductionVerificationResult } from "./bounded-verification.ts";
import { validateScopedPath, type ScopedWorkspace, type WorkspaceMutationResult } from "./tool-policy.ts";

const SHA_PATTERN = /^[0-9a-f]{40}$/;
const SAFE_IDENTIFIER = /^[A-Za-z0-9](?:[A-Za-z0-9._:-]*[A-Za-z0-9])?$/;

export interface ProductionAgentSessionPort {
	run(request: RoleRunRequest): Promise<AgentSessionHandoff>;
	abort(runId: string): Promise<void>;
	close?(): Promise<void>;
}

export interface ProductionVerificationPort {
	runAll(
		worktreeRoot: string,
		commands: ProductionChildSpec["verification"],
		signal?: AbortSignal,
	): Promise<ProductionVerificationResult[]>;
}

export interface ProductionParentRefreshRequest {
	previousParentHead: string;
	newParentHead: string;
	effectKey: string;
}

export interface ProductionParentRefreshEvidence {
	outcome: "unchanged" | "rebased" | "reclaimed";
	previousBaseHead: string;
	baseHead: string;
	previousHead: string;
	head: string;
	verificationInvalidated: true;
	reviewInvalidated: true;
}

export interface ProductionChildHeadReconciliationRequest {
	previousHead: string;
	effectKey: string;
}

export interface ProductionChildHeadReconciliationEvidence {
	outcome: "reclaimed" | "reused";
	branch: string;
	baseHead: string;
	previousHead: string;
	head: string;
	changedScope: string[];
	verificationInvalidated: true;
	reviewInvalidated: true;
	integrationInvalidated: true;
}

export interface ProductionWorkspaceAdapterPort {
	claim(request: WorkspaceClaimRequest): Promise<ClaimedWorkspace>;
	commitIssueChanges(workspace: ClaimedWorkspace, request: GitCommitRequest): Promise<GitCommitEvidence>;
	pushIssueBranch(workspace: ClaimedWorkspace, request: GitPushRequest): Promise<GitPushEvidence>;
	captureHandoff(workspace: ClaimedWorkspace, verificationState: "pending" | "passed" | "failed"): Promise<WorkspaceHandoffEvidence>;
	refreshParent(workspace: ClaimedWorkspace, request: ProductionParentRefreshRequest): Promise<ProductionParentRefreshEvidence>;
	reconcileChildHead(
		workspace: ClaimedWorkspace,
		request: ProductionChildHeadReconciliationRequest,
	): Promise<ProductionChildHeadReconciliationEvidence>;
}

export interface ProductionWorkspaceLifecycleOptions {
	workspaceAdapter: ProductionWorkspaceAdapterPort;
	verification: ProductionVerificationPort;
	agentSession: ProductionAgentSessionPort;
}

export interface ProductionWorkspaceClaim {
	runId: string;
	generation: number;
	coordinator: GitBinding;
	trustedWorktreeRoot: string;
	parentIssue: number;
	parentBranch: string;
	parentHead: string;
	child: ProductionChildSpec;
	mode: "start" | "resume";
	/** Exact prior owner on resume; it must equal the stable plan/issue-derived owner. */
	ownershipId?: string;
}

export interface ProductionImplementationRequest {
	timeoutMs: number;
	context?: readonly string[];
	signal?: AbortSignal;
}

export interface ProductionWorkspaceSession {
	readonly binding: ProductionWorkspaceBinding;
	implement(request: ProductionImplementationRequest): Promise<AgentSessionHandoff>;
	correct(request: ProductionImplementationRequest & { findings: string[] }): Promise<AgentSessionHandoff>;
	verify(signal?: AbortSignal): Promise<ProductionVerificationResult[]>;
	commit(message: string, signal?: AbortSignal): Promise<GitCommitEvidence>;
	push(signal?: AbortSignal): Promise<GitPushEvidence>;
	captureHandoff(signal?: AbortSignal): Promise<WorkspaceHandoffEvidence>;
	refreshParent(request: ProductionParentRefreshRequest, signal?: AbortSignal): Promise<ProductionParentRefreshEvidence>;
	reconcileChildHead(
		request: ProductionChildHeadReconciliationRequest,
		signal?: AbortSignal,
	): Promise<ProductionChildHeadReconciliationEvidence>;
	join(): Promise<void>;
}

export function productionWorkspaceOwnershipId(parentIssue: number, childIssue: number, childId: string): string {
	if (!Number.isSafeInteger(parentIssue) || parentIssue < 1 || !Number.isSafeInteger(childIssue) || childIssue < 1
		|| typeof childId !== "string" || !/^[a-z0-9][a-z0-9_-]{0,63}$/.test(childId)) {
		throw new Error("production workspace ownership inputs are invalid");
	}
	return `production:${createHash("sha256")
		.update(`${parentIssue}\0${childIssue}\0${childId}`)
		.digest("hex")}`;
}

export class ProductionWorkspaceLifecycle {
	readonly #workspaceAdapter: ProductionWorkspaceAdapterPort;
	readonly #verification: ProductionVerificationPort;
	readonly #agentSession: ProductionAgentSessionPort;
	readonly #sessions = new Set<OwnedProductionWorkspaceSession>();
	#closed = false;
	#closePromise: Promise<void> | undefined;

	constructor(options: ProductionWorkspaceLifecycleOptions) {
		if (typeof options !== "object" || options === null
			|| typeof options.workspaceAdapter?.claim !== "function"
			|| typeof options.workspaceAdapter?.commitIssueChanges !== "function"
			|| typeof options.workspaceAdapter?.pushIssueBranch !== "function"
			|| typeof options.workspaceAdapter?.captureHandoff !== "function"
			|| typeof options.workspaceAdapter?.refreshParent !== "function"
			|| typeof options.workspaceAdapter?.reconcileChildHead !== "function"
			|| typeof options.verification?.runAll !== "function"
			|| typeof options.agentSession?.run !== "function"
			|| typeof options.agentSession?.abort !== "function") {
			throw new Error("production workspace lifecycle ports are invalid");
		}
		this.#workspaceAdapter = options.workspaceAdapter;
		this.#verification = options.verification;
		this.#agentSession = options.agentSession;
	}

	async claim(request: ProductionWorkspaceClaim): Promise<ProductionWorkspaceSession> {
		if (this.#closed) throw new Error("production workspace lifecycle is closed");
		validateClaim(request);
		const ownershipId = productionWorkspaceOwnershipId(request.parentIssue, request.child.issue, request.child.id);
		if (request.ownershipId !== undefined && request.ownershipId !== ownershipId) {
			throw new Error("persisted production workspace ownership changed across resume");
		}
		const workspace = await this.#workspaceAdapter.claim({
			coordinator: request.coordinator,
			trustedWorktreeRoot: request.trustedWorktreeRoot,
			issue: request.child.issue,
			slug: request.child.slug,
			parentIssue: request.parentIssue,
			parentBranch: request.parentBranch,
			parentHead: request.parentHead,
			ownershipId,
			allowedScopes: request.child.writeScopes,
			leaseMode: request.mode,
		});
		const session = new OwnedProductionWorkspaceSession({
			workspace,
			ownershipId,
			request,
			workspaceAdapter: this.#workspaceAdapter,
			verification: this.#verification,
			agentSession: this.#agentSession,
			onJoined: () => this.#sessions.delete(session),
		});
		this.#sessions.add(session);
		return session;
	}

	async abort(runId: string): Promise<void> {
		if (typeof runId !== "string" || !SAFE_IDENTIFIER.test(runId)) throw new Error("production run ID is invalid");
		const sessions = [...this.#sessions].filter((session) => session.runId === runId);
		const failures: unknown[] = [];
		collectFailures(await Promise.allSettled([this.#agentSession.abort(runId)]), failures);
		collectFailures(await Promise.allSettled(sessions.map((session) => session.join())), failures);
		throwFailures(failures, "production run abort/join failed");
	}

	async close(): Promise<void> {
		if (this.#closePromise !== undefined) return this.#closePromise;
		this.#closed = true;
		this.#closePromise = (async () => {
			const runIds = [...new Set([...this.#sessions].map((session) => session.runId))];
			const failures: unknown[] = [];
			collectFailures(await Promise.allSettled(runIds.map((runId) => this.#agentSession.abort(runId))), failures);
			collectFailures(await Promise.allSettled([...this.#sessions].map((session) => session.join())), failures);
			if (this.#agentSession.close) {
				collectFailures(await Promise.allSettled([this.#agentSession.close()]), failures);
			}
			throwFailures(failures, "production workspace lifecycle close failed");
		})();
		return this.#closePromise;
	}
}

interface OwnedSessionOptions {
	workspace: ClaimedWorkspace;
	ownershipId: string;
	request: ProductionWorkspaceClaim;
	workspaceAdapter: ProductionWorkspaceAdapterPort;
	verification: ProductionVerificationPort;
	agentSession: ProductionAgentSessionPort;
	onJoined(): void;
}

class OwnedProductionWorkspaceSession implements ProductionWorkspaceSession {
	readonly runId: string;
	readonly #workspace: ClaimedWorkspace;
	readonly #ownershipId: string;
	readonly #request: ProductionWorkspaceClaim;
	readonly #workspaceAdapter: ProductionWorkspaceAdapterPort;
	readonly #verification: ProductionVerificationPort;
	readonly #agentSession: ProductionAgentSessionPort;
	readonly #onJoined: () => void;
	#verificationState: "pending" | "passed" | "failed" = "pending";
	#reviewValid = false;
	#accepting = true;
	#tail: Promise<void> = Promise.resolve();
	#joinPromise: Promise<void> | undefined;

	constructor(options: OwnedSessionOptions) {
		this.runId = options.request.runId;
		this.#workspace = options.workspace;
		this.#ownershipId = options.ownershipId;
		this.#request = options.request;
		this.#workspaceAdapter = options.workspaceAdapter;
		this.#verification = options.verification;
		this.#agentSession = options.agentSession;
		this.#onJoined = options.onJoined;
	}

	get binding(): ProductionWorkspaceBinding {
		return {
			claimId: this.#workspace.claimId,
			ownershipId: this.#ownershipId,
			repositoryIdentity: this.#workspace.repositoryIdentity,
			worktreeIdentity: this.#workspace.worktreeIdentity,
			cwd: this.#workspace.cwd,
			branch: this.#workspace.branch,
			baseBranch: this.#workspace.prBase,
			baseHead: this.#workspace.baseHead,
			head: this.#workspace.head,
			writeScopes: [...this.#workspace.allowedScopes],
		};
	}

	implement(request: ProductionImplementationRequest): Promise<AgentSessionHandoff> {
		return this.#runMutatingAgent("implementation", request, this.#request.child.task);
	}

	correct(request: ProductionImplementationRequest & { findings: string[] }): Promise<AgentSessionHandoff> {
		if (!Array.isArray(request?.findings) || request.findings.length < 1 || request.findings.length > 64
			|| request.findings.some((finding) => typeof finding !== "string" || finding.length < 1
				|| Buffer.byteLength(finding) > 4_096 || /[\u0000-\u001f\u007f-\u009f]/.test(finding))) {
			return Promise.reject(new Error("production correction findings are invalid"));
		}
		return this.#runMutatingAgent(
			"correction",
			request,
			`${this.#request.child.task}\n\nCorrect these bounded review findings:\n${request.findings.map((finding) => `- ${finding}`).join("\n")}`,
		);
	}

	#runMutatingAgent(
		role: "implementation" | "correction",
		request: ProductionImplementationRequest,
		task: string,
	): Promise<AgentSessionHandoff> {
		if (typeof request !== "object" || request === null || !Number.isSafeInteger(request.timeoutMs)
			|| request.timeoutMs < 1 || request.timeoutMs > 24 * 60 * 60 * 1_000
			|| (request.context !== undefined && (!Array.isArray(request.context)
				|| request.context.some((item) => typeof item !== "string")))) {
			return Promise.reject(new Error(`production ${role} request is invalid`));
		}
		return this.#enqueue(async () => {
			assertNotAborted(request.signal);
			await this.#workspace.assertOwned();
			const binding = this.binding;
			const laneId = `${this.#request.child.id}-${role}`;
			const promptBinding = {
				runId: this.runId,
				generation: this.#request.generation,
				laneId,
				candidateHead: binding.head,
				validationNonce: createHash("sha256")
					.update(`${this.runId}:${this.#request.generation}:${laneId}:${binding.head}`)
					.digest("hex"),
			};
			const context = [
				`Parent issue #${this.#request.parentIssue}; child issue #${this.#request.child.issue}.`,
				`Canonical branch ${binding.branch}; PR base ${binding.baseBranch}; exact base ${binding.baseHead}.`,
				`Declared write scopes: ${binding.writeScopes.join(", ")}.`,
				`Required skills: ${this.#request.child.requiredSkills.join(", ")}.`,
				"Use strict RED GREEN REFACTOR and return only the typed handoff.",
				...(request.context ?? []),
			];
			const handoff = await this.#agentSession.run({
				role,
				task,
				context,
				timeoutMs: request.timeoutMs,
				...(request.signal ? { signal: request.signal } : {}),
				workspace: createLeaseBoundWorkspace(this.#workspace),
				capabilities: [],
				authority: {
					issue: this.#request.child.issue,
					branch: binding.branch,
					readOnly: false,
					workspaceId: `production-${this.#request.child.issue}-${this.#request.child.id}`,
					readPrefixes: ["."],
					writePrefixes: [...binding.writeScopes],
					capabilityNames: [],
				},
				binding: promptBinding,
			});
			assertNotAborted(request.signal);
			if (handoff.status !== "completed" || handoff.role !== role
				|| handoff.runId !== promptBinding.runId || handoff.generation !== promptBinding.generation
				|| handoff.laneId !== promptBinding.laneId || handoff.candidateHead !== promptBinding.candidateHead
				|| handoff.validationNonce !== promptBinding.validationNonce) {
				throw new LifecycleError("correction_required", `${role} AgentSession returned an invalid or incomplete handoff`);
			}
			this.#verificationState = "pending";
			this.#reviewValid = false;
			return handoff;
		});
	}

	verify(signal?: AbortSignal): Promise<ProductionVerificationResult[]> {
		return this.#enqueue(async () => {
			assertNotAborted(signal);
			await this.#workspace.assertOwned();
			const results = await this.#verification.runAll(
				this.#workspace.cwd,
				this.#request.child.verification,
				signal,
			);
			assertNotAborted(signal);
			this.#verificationState = results.length === this.#request.child.verification.length
				&& results.every((result) => result.status === "passed") ? "passed" : "failed";
			this.#reviewValid = false;
			return results;
		});
	}

	commit(message: string, signal?: AbortSignal): Promise<GitCommitEvidence> {
		return this.#enqueue(async () => {
			this.#assertVerified();
			assertNotAborted(signal);
			const evidence = await this.#workspaceAdapter.commitIssueChanges(this.#workspace, {
				issue: this.#request.child.issue,
				slug: this.#request.child.slug,
				branch: this.#workspace.branch,
				expectedHead: this.#workspace.head,
				message,
				scopes: this.#request.child.writeScopes,
			});
			assertNotAborted(signal);
			this.#workspace.head = evidence.head;
			return evidence;
		});
	}

	push(signal?: AbortSignal): Promise<GitPushEvidence> {
		return this.#enqueue(async () => {
			this.#assertVerified();
			assertNotAborted(signal);
			if (this.#workspace.defaultBranch === undefined) throw new Error("workspace has no authoritative default branch");
			const evidence = await this.#workspaceAdapter.pushIssueBranch(this.#workspace, {
				issue: this.#request.child.issue,
				slug: this.#request.child.slug,
				branch: this.#workspace.branch,
				expectedHead: this.#workspace.head,
				defaultBranch: this.#workspace.defaultBranch,
			});
			assertNotAborted(signal);
			return evidence;
		});
	}

	captureHandoff(signal?: AbortSignal): Promise<WorkspaceHandoffEvidence> {
		return this.#enqueue(async () => {
			assertNotAborted(signal);
			const evidence = await this.#workspaceAdapter.captureHandoff(this.#workspace, this.#verificationState);
			assertNotAborted(signal);
			if (this.#verificationState !== "passed" || evidence.verificationState !== "passed"
				|| evidence.head !== this.#workspace.head || evidence.baseHead !== this.#workspace.baseHead
				|| evidence.branch !== this.#workspace.branch || evidence.prBase !== this.#workspace.prBase
				|| evidence.dirty) {
				throw new LifecycleError("correction_required", "workspace handoff is dirty, stale, or unverified");
			}
			return evidence;
		});
	}

	refreshParent(request: ProductionParentRefreshRequest, signal?: AbortSignal): Promise<ProductionParentRefreshEvidence> {
		return this.#enqueue(async () => {
			validateRefresh(request, this.#workspace.baseHead);
			assertNotAborted(signal);
			const evidence = await this.#workspaceAdapter.refreshParent(this.#workspace, request);
			assertNotAborted(signal);
			if (evidence.previousBaseHead !== request.previousParentHead || evidence.baseHead !== request.newParentHead
				|| evidence.verificationInvalidated !== true || evidence.reviewInvalidated !== true
				|| !SHA_PATTERN.test(evidence.head)) {
				throw new LifecycleError("stale_parent", "parent refresh returned mismatched authoritative evidence");
			}
			this.#workspace.baseHead = evidence.baseHead;
			this.#workspace.head = evidence.head;
			this.#verificationState = "pending";
			this.#reviewValid = false;
			return evidence;
		});
	}

	reconcileChildHead(
		request: ProductionChildHeadReconciliationRequest,
		signal?: AbortSignal,
	): Promise<ProductionChildHeadReconciliationEvidence> {
		return this.#enqueue(async () => {
			if (typeof request !== "object" || request === null || !SHA_PATTERN.test(request.previousHead)
				|| typeof request.effectKey !== "string" || !SAFE_IDENTIFIER.test(request.effectKey)
				|| Buffer.byteLength(request.effectKey) > 256) {
				throw new LifecycleError("terminal", "child-head reconciliation request is invalid");
			}
			assertNotAborted(signal);
			await this.#workspace.assertOwned();
			const expectedBranch = this.#workspace.branch;
			const expectedBase = this.#workspace.baseHead;
			const evidence = await this.#workspaceAdapter.reconcileChildHead(this.#workspace, request);
			// Once the adapter may have changed the local head, old evidence is never reusable,
			// including when the returned adapter evidence later fails validation.
			this.#verificationState = "pending";
			this.#reviewValid = false;
			assertNotAborted(signal);
			let changedScope: string[];
			try {
				if (!Array.isArray(evidence.changedScope) || evidence.changedScope.length > 4_096) throw new Error();
				changedScope = evidence.changedScope.map((path) => validateScopedPath(path, this.#request.child.writeScopes));
			} catch {
				throw new LifecycleError("terminal", "child-head reconciliation returned invalid scope evidence", ["scope_mismatch"]);
			}
			const canonicalScope = [...new Set(changedScope)].sort();
			if ((evidence.outcome !== "reclaimed" && evidence.outcome !== "reused")
				|| evidence.branch !== expectedBranch || evidence.baseHead !== expectedBase
				|| evidence.previousHead !== request.previousHead || !SHA_PATTERN.test(evidence.head)
				|| evidence.head === request.previousHead || this.#workspace.head !== evidence.head
				|| JSON.stringify(canonicalScope) !== JSON.stringify(evidence.changedScope)
				|| evidence.verificationInvalidated !== true || evidence.reviewInvalidated !== true
				|| evidence.integrationInvalidated !== true) {
				throw new LifecycleError("terminal", "child-head reconciliation returned mismatched authoritative evidence", ["child_head_reconciliation_mismatch"]);
			}
			this.#workspace.head = evidence.head;
			return evidence;
		});
	}

	join(): Promise<void> {
		if (this.#joinPromise !== undefined) return this.#joinPromise;
		this.#accepting = false;
		const accepted = this.#tail;
		this.#joinPromise = (async () => {
			try {
				await accepted;
			} finally {
				await this.#workspace.release();
				this.#onJoined();
			}
		})();
		return this.#joinPromise;
	}

	#assertVerified(): void {
		if (this.#verificationState !== "passed") {
			throw new LifecycleError("correction_required", "current exact workspace head requires passing verification");
		}
	}

	#enqueue<T>(operation: () => Promise<T>): Promise<T> {
		if (!this.#accepting) return Promise.reject(new Error("production workspace session is joining or closed"));
		const result = this.#tail.then(operation);
		this.#tail = result.then(() => undefined, () => undefined);
		return result;
	}
}

function validateClaim(request: ProductionWorkspaceClaim): void {
	if (typeof request !== "object" || request === null || !SAFE_IDENTIFIER.test(request.runId)
		|| !Number.isSafeInteger(request.generation) || request.generation < 1
		|| !Number.isSafeInteger(request.parentIssue) || request.parentIssue < 1
		|| typeof request.parentBranch !== "string"
		|| !/^(?!\/|.*(?:\.\.|\s|[~^:?*\\\[\]])|.*\/$)[A-Za-z0-9][A-Za-z0-9._\/-]{0,239}$/.test(request.parentBranch)
		|| !SHA_PATTERN.test(request.parentHead) || !isProductionChild(request.child)
		|| (request.mode !== "start" && request.mode !== "resume")
		|| (request.ownershipId !== undefined && (typeof request.ownershipId !== "string"
			|| !SAFE_IDENTIFIER.test(request.ownershipId) || Buffer.byteLength(request.ownershipId) > 256))) {
		throw new Error("production workspace claim is invalid");
	}
}

function isProductionChild(child: ProductionChildSpec): boolean {
	return typeof child === "object" && child !== null && child.access === "mutating"
		&& Number.isSafeInteger(child.issue) && child.issue > 0
		&& /^[a-z0-9][a-z0-9_-]{0,63}$/.test(child.id)
		&& /^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(child.slug)
		&& Array.isArray(child.writeScopes) && child.writeScopes.length > 0;
}

function validateRefresh(request: ProductionParentRefreshRequest, currentBase: string): void {
	if (typeof request !== "object" || request === null || !SHA_PATTERN.test(request.previousParentHead)
		|| !SHA_PATTERN.test(request.newParentHead) || request.previousParentHead !== currentBase
		|| typeof request.effectKey !== "string" || !SAFE_IDENTIFIER.test(request.effectKey)
		|| Buffer.byteLength(request.effectKey) > 256) {
		throw new LifecycleError("stale_parent", "parent refresh does not match the current workspace base");
	}
}

function assertNotAborted(signal?: AbortSignal): void {
	if (signal !== undefined && !(signal instanceof AbortSignal)) throw new Error("operation AbortSignal is invalid");
	if (signal?.aborted) throw new LifecycleError("retryable", "production workspace operation was aborted");
}

function createLeaseBoundWorkspace(workspace: ClaimedWorkspace): ScopedWorkspace {
	return {
		id: `claim-${workspace.claimId}`,
		cwd: workspace.cwd,
		async readText(path, options) {
			assertNotAborted(options.signal);
			await workspace.assertOwned();
			const target = await resolveScopedFile(workspace, path, ["."], false);
			const value = await readFile(target, "utf8");
			assertNotAborted(options.signal);
			const offset = options.offset ?? 0;
			return value.slice(offset, options.limit === undefined ? undefined : offset + options.limit);
		},
		async editText(path, oldText, newText, signal) {
			assertNotAborted(signal);
			await workspace.assertOwned();
			const target = await resolveScopedFile(workspace, path, workspace.allowedScopes, false);
			const current = await readFile(target, "utf8");
			const first = current.indexOf(oldText);
			if (first < 0 || current.indexOf(oldText, first + oldText.length) >= 0) {
				throw new Error("workspace edit requires exactly one oldText match");
			}
			const next = `${current.slice(0, first)}${newText}${current.slice(first + oldText.length)}`;
			await workspace.assertOwned();
			assertNotAborted(signal);
			await writeFile(target, next, "utf8");
			return mutationResult(current, next);
		},
		async writeText(path, content, signal) {
			assertNotAborted(signal);
			await workspace.assertOwned();
			const target = await resolveScopedFile(workspace, path, workspace.allowedScopes, true);
			let current: string | undefined;
			try { current = await readFile(target, "utf8"); } catch { /* New scoped file. */ }
			await workspace.assertOwned();
			assertNotAborted(signal);
			await writeFile(target, content, "utf8");
			return mutationResult(current, content);
		},
	};
}

async function resolveScopedFile(
	workspace: ClaimedWorkspace,
	path: string,
	prefixes: readonly string[],
	createParents: boolean,
): Promise<string> {
	const normalized = validateScopedPath(path, prefixes);
	const root = await realpath(workspace.cwd);
	const target = resolve(root, normalized);
	const back = relative(root, target);
	if (back === ".." || back.startsWith(`..${sep}`)) throw new Error("workspace path escapes its claim");
	let current = root;
	const parents = normalized.split("/").slice(0, -1);
	for (const part of parents) {
		current = join(current, part);
		try {
			const metadata = await lstat(current);
			if (!metadata.isDirectory() || metadata.isSymbolicLink()) throw new Error("workspace path traverses a symlink");
		} catch (error) {
			if (!createParents || !(error instanceof Error && "code" in error && error.code === "ENOENT")) throw error;
			await mkdir(current, { mode: 0o700 });
		}
	}
	try {
		const metadata = await lstat(target);
		if (!metadata.isFile() || metadata.isSymbolicLink()) throw new Error("workspace target must be a regular file, not a symlink");
	} catch (error) {
		if (!createParents || !(error instanceof Error && "code" in error && error.code === "ENOENT")) throw error;
	}
	return target;
}

function mutationResult(previous: string | undefined, next: string): WorkspaceMutationResult {
	return { changed: previous !== next, summary: previous === next ? "unchanged" : "updated one scoped file" };
}

function collectFailures(settlements: readonly PromiseSettledResult<unknown>[], failures: unknown[]): void {
	for (const settlement of settlements) if (settlement.status === "rejected") failures.push(settlement.reason);
}

function throwFailures(failures: readonly unknown[], message: string): void {
	if (failures.length === 1) throw failures[0];
	if (failures.length > 1) throw new AggregateError(failures, message);
}
