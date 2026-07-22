import { createHash, randomUUID } from "node:crypto";
import { constants } from "node:fs";
import { link, lstat, mkdir, open, readFile, realpath, unlink } from "node:fs/promises";
import { isAbsolute, join, relative, resolve, sep } from "node:path";

import {
	validateProductionParentPlan,
	type ProductionChildSpec,
	type ProductionParentPlanDocument,
} from "./autonomous-production-contract.ts";
import type { RoleRunRequest } from "./agent-session-runtime.ts";
import type {
	OrchestrationCallContext,
	ParentOrchestrationPlan,
} from "./github-orchestrator.ts";
import { ProductionRepositoryPlanIntake, type ProductionPlanSnapshot } from "./production-intake.ts";
import { productionOrchestrationObjective } from "./production-orchestration-plan.ts";
import type { ProductionAgentSessionPort } from "./production-workspace-lifecycle.ts";
import {
	validateScopedPath,
	type CapabilityResult,
	type HostCapability,
	type ScopedWorkspace,
} from "./tool-policy.ts";

const SHA = /^[0-9a-f]{40}$/u;
const DIGEST = /^[0-9a-f]{64}$/u;
const REPOSITORY = /^[A-Za-z0-9][A-Za-z0-9._-]{0,99}\/[A-Za-z0-9][A-Za-z0-9._-]{0,99}$/u;
const SAFE_REF = /^(?!\/|.*(?:\.\.|\s|[~^:?*\\\[\]])|.*\/$)[A-Za-z0-9][A-Za-z0-9._\/-]{0,239}$/u;
const SAFE_ACTOR = /^[a-z\d](?:[a-z\d-]{0,37}[a-z\d])?$/u;
const MAX_CHILDREN = 64;
const MAX_FACTS_BYTES = 512 * 1024;
const MAX_READ_BYTES = 256 * 1024;
const DEFAULT_TIMEOUT_MS = 15 * 60 * 1_000;

export interface ProductionPlanningIssueFacts {
	schemaVersion: 1;
	repository: string;
	defaultBranch: string;
	viewer: { login: string; permission: "admin" | "maintain" };
	parent: {
		number: number;
		nodeId: string;
		title: string;
		body: string;
		state: "open";
		updatedAt: string;
	};
	subissues: Array<{
		number: number;
		title: string;
		body: string;
		state: "open" | "closed";
		updatedAt: string;
	}>;
	complete: true;
	revisionDigest: string;
	observedAt: string;
}

export interface ProductionPlanAuthority {
	repository: string;
	parentIssue: number;
	parentBranch: string;
	parentBaseBranch: string;
	candidateHead: string;
}

export interface ProductionParentPlanProposal {
	schemaVersion: 1;
	sourceRevisionDigest: string;
	title: string;
	objective: string;
	children: Array<Omit<ProductionChildSpec, "issue">>;
}

export interface ProductionPlanningCallContext {
	signal: AbortSignal;
	deadlineAt: string;
}

export interface ProductionPlanningIssueSource {
	observe(
		query: { repositoryRoot: string; parentIssue: number },
		context: ProductionPlanningCallContext,
	): Promise<ProductionPlanningIssueFacts>;
}

export interface ProductionPlanAuthoritySource {
	observe(
		issue: number,
		facts: ProductionPlanningIssueFacts,
		signal: AbortSignal,
	): Promise<ProductionPlanAuthority>;
}

export interface ProductionPlanSession {
	propose(
		input: { facts: ProductionPlanningIssueFacts; authority: ProductionPlanAuthority },
		context: ProductionPlanningCallContext,
	): Promise<ProductionParentPlanProposal>;
}

export interface ProductionPlanGitHubPort {
	createPlan(value: unknown, context?: OrchestrationCallContext): Promise<ParentOrchestrationPlan>;
	ensureChildIssue(
		plan: ParentOrchestrationPlan,
		childId: string,
		context?: OrchestrationCallContext,
	): Promise<{ number: number }>;
	stop(context?: OrchestrationCallContext): Promise<unknown>;
}

export interface ProductionPlanBootstrapperOptions {
	repositoryRoot: string;
	stateRoot: string;
	intake: ProductionRepositoryPlanIntake;
	issueSource: ProductionPlanningIssueSource;
	authoritySource: ProductionPlanAuthoritySource;
	planSession: ProductionPlanSession;
	github: ProductionPlanGitHubPort;
	now?: () => Date;
	timeoutMs?: number;
}

interface BootstrapJournal {
	schemaVersion: 1;
	parentIssue: number;
	repository: string;
	parentBranch: string;
	parentBaseBranch: string;
	candidateHead: string;
	sourceRevisionDigest: string;
	actor: string;
	decisionExpiresAt: string;
	proposal: ProductionParentPlanProposal;
}

function exactRecord(value: unknown, keys: readonly string[], description: string): Record<string, unknown> {
	if (typeof value !== "object" || value === null || Array.isArray(value)) {
		throw new Error(`${description} must be an exact record`);
	}
	const actual = Object.keys(value).sort();
	const expected = [...keys].sort();
	if (JSON.stringify(actual) !== JSON.stringify(expected)) {
		throw new Error(`${description} contains an unknown or missing field`);
	}
	return value as Record<string, unknown>;
}

function safeText(value: unknown, description: string, maximum: number): string {
	if (typeof value !== "string" || value.length < 1 || Buffer.byteLength(value) > maximum
		|| /[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f-\u009f]/u.test(value)) {
		throw new Error(`${description} must be bounded safe text`);
	}
	return value;
}

function safeTimestamp(value: unknown, description: string): string {
	const text = safeText(value, description, 128);
	const date = new Date(text);
	if (!Number.isFinite(date.valueOf())) throw new Error(`${description} must be an RFC3339 timestamp`);
	return date.toISOString();
}

function positiveInteger(value: unknown, description: string): number {
	if (!Number.isSafeInteger(value) || (value as number) < 1) throw new Error(`${description} must be positive`);
	return value as number;
}

function assertActive(signal: AbortSignal): void {
	if (!(signal instanceof AbortSignal)) throw new Error("production bootstrap AbortSignal is invalid");
	if (signal.aborted) throw signal.reason ?? new Error("production plan bootstrap cancelled");
}

function isMissing(error: unknown): boolean {
	return typeof error === "object" && error !== null && "code" in error && error.code === "ENOENT";
}

function canonicalJson(value: unknown): string {
	if (Array.isArray(value)) return `[${value.map(canonicalJson).join(",")}]`;
	if (value !== null && typeof value === "object") {
		return `{${Object.keys(value as Record<string, unknown>).sort()
			.map((key) => `${JSON.stringify(key)}:${canonicalJson((value as Record<string, unknown>)[key])}`)
			.join(",")}}`;
	}
	return JSON.stringify(value);
}

function canonicalDigest(value: unknown): string {
	return createHash("sha256").update(canonicalJson(value)).digest("hex");
}

export function validateProductionPlanningIssueFacts(
	value: unknown,
	expectedIssue?: number,
): ProductionPlanningIssueFacts {
	const root = exactRecord(value, [
		"schemaVersion", "repository", "defaultBranch", "viewer", "parent", "subissues",
		"complete", "revisionDigest", "observedAt",
	], "planning issue facts");
	if (root.schemaVersion !== 1 || root.complete !== true) throw new Error("planning issue facts are incomplete");
	const repository = safeText(root.repository, "planning repository", 201);
	if (!REPOSITORY.test(repository)) throw new Error("planning repository is invalid");
	const defaultBranch = safeText(root.defaultBranch, "planning default branch", 240);
	if (!SAFE_REF.test(defaultBranch)) throw new Error("planning default branch is invalid");
	const viewer = exactRecord(root.viewer, ["login", "permission"], "planning viewer");
	const login = safeText(viewer.login, "planning viewer login", 39).toLowerCase();
	if (!SAFE_ACTOR.test(login) || (viewer.permission !== "admin" && viewer.permission !== "maintain")) {
		throw new Error("planning requires an authenticated admin or maintainer");
	}
	const parent = exactRecord(root.parent, [
		"number", "nodeId", "title", "body", "state", "updatedAt",
	], "planning parent issue");
	const parentIssue = positiveInteger(parent.number, "planning parent issue");
	if (expectedIssue !== undefined && parentIssue !== expectedIssue) throw new Error("planning parent issue mismatch");
	if (parent.state !== "open") throw new Error("planning parent issue must be open");
	const normalizedParent = {
		number: parentIssue,
		nodeId: safeText(parent.nodeId, "planning parent node ID", 512),
		title: safeText(parent.title, "planning parent title", 256),
		body: typeof parent.body === "string" && Buffer.byteLength(parent.body) <= 64 * 1024
			? parent.body : (() => { throw new Error("planning parent body is invalid"); })(),
		state: "open" as const,
		updatedAt: safeTimestamp(parent.updatedAt, "planning parent update time"),
	};
	if (!Array.isArray(root.subissues) || root.subissues.length > MAX_CHILDREN) {
		throw new Error("planning subissues must be a bounded array");
	}
	const seen = new Set<number>();
	const subissues = root.subissues.map((entry, index) => {
		const issue = exactRecord(entry, ["number", "title", "body", "state", "updatedAt"], `planning subissue ${index}`);
		const number = positiveInteger(issue.number, "planning subissue number");
		if (number === parentIssue || seen.has(number)) throw new Error("planning subissue numbers must be unique");
		seen.add(number);
		const state = issue.state;
		if (state !== "open" && state !== "closed") throw new Error("planning subissue state is invalid");
		if (typeof issue.body !== "string" || Buffer.byteLength(issue.body) > 64 * 1024) {
			throw new Error("planning subissue body is invalid");
		}
		return {
			number,
			title: safeText(issue.title, "planning subissue title", 256),
			body: issue.body,
			state: state as "open" | "closed",
			updatedAt: safeTimestamp(issue.updatedAt, "planning subissue update time"),
		};
	});
	const revisionDigest = safeText(root.revisionDigest, "planning source revision", 64);
	if (!DIGEST.test(revisionDigest)) throw new Error("planning source revision is invalid");
	const result: ProductionPlanningIssueFacts = {
		schemaVersion: 1,
		repository,
		defaultBranch,
		viewer: { login, permission: viewer.permission },
		parent: normalizedParent,
		subissues,
		complete: true,
		revisionDigest,
		observedAt: safeTimestamp(root.observedAt, "planning observation time"),
	};
	if (Buffer.byteLength(JSON.stringify(result)) > MAX_FACTS_BYTES) throw new Error("planning issue facts exceed their bound");
	return result;
}

function validateAuthority(value: unknown, facts: ProductionPlanningIssueFacts): ProductionPlanAuthority {
	const candidate = exactRecord(value, [
		"repository", "parentIssue", "parentBranch", "parentBaseBranch", "candidateHead",
	], "production plan authority");
	const repository = safeText(candidate.repository, "authority repository", 201);
	const parentIssue = positiveInteger(candidate.parentIssue, "authority parent issue");
	const parentBranch = safeText(candidate.parentBranch, "authority parent branch", 240);
	const parentBaseBranch = safeText(candidate.parentBaseBranch, "authority parent base", 240);
	const candidateHead = safeText(candidate.candidateHead, "authority candidate head", 40);
	if (repository !== facts.repository || parentIssue !== facts.parent.number
		|| parentBaseBranch !== facts.defaultBranch || parentBranch === parentBaseBranch
		|| !SAFE_REF.test(parentBranch) || !SAFE_REF.test(parentBaseBranch) || !SHA.test(candidateHead)) {
		throw new Error("production plan authority conflicts with authoritative issue or Git coordinates");
	}
	return { repository, parentIssue, parentBranch, parentBaseBranch, candidateHead };
}

function syntheticPlan(
	proposal: ProductionParentPlanProposal,
	facts: ProductionPlanningIssueFacts,
	authority: ProductionPlanAuthority,
	actor: string,
	decisionExpiresAt: string,
): ProductionParentPlanDocument {
	return validateProductionParentPlan({
		schemaVersion: 2,
		planId: `issue-${authority.parentIssue}-${proposal.sourceRevisionDigest.slice(0, 16)}`,
		parentIssue: authority.parentIssue,
		repository: authority.repository,
		title: proposal.title,
		objective: proposal.objective,
		parentBranch: authority.parentBranch,
		parentBaseBranch: authority.parentBaseBranch,
		actorAllowlist: [actor],
		decisionExpiresAt,
		children: proposal.children.map((child, index) => ({ ...child, issue: index + 1 })),
	}, facts.parent.number);
}

function assertAcyclic(children: readonly Omit<ProductionChildSpec, "issue">[]): void {
	const byId = new Map(children.map((child) => [child.id, child]));
	const visiting = new Set<string>();
	const visited = new Set<string>();
	const visit = (id: string): void => {
		if (visiting.has(id)) throw new Error("production plan proposal dependency cycle");
		if (visited.has(id)) return;
		visiting.add(id);
		for (const dependency of byId.get(id)?.dependsOn ?? []) visit(dependency);
		visiting.delete(id);
		visited.add(id);
	};
	for (const child of children) visit(child.id);
}

export function validateProductionParentPlanProposal(
	value: unknown,
	factsValue: ProductionPlanningIssueFacts,
	authorityValue: ProductionPlanAuthority,
): ProductionParentPlanProposal {
	const facts = validateProductionPlanningIssueFacts(factsValue);
	const authority = validateAuthority(authorityValue, facts);
	const root = exactRecord(value, [
		"schemaVersion", "sourceRevisionDigest", "title", "objective", "children",
	], "production plan proposal");
	if (root.schemaVersion !== 1 || root.sourceRevisionDigest !== facts.revisionDigest) {
		throw new Error("production plan proposal source revision mismatch");
	}
	if (!Array.isArray(root.children) || root.children.length < 1 || root.children.length > MAX_CHILDREN) {
		throw new Error("production plan proposal children are invalid");
	}
	const childKeys = [
		"id", "title", "task", "slug", "dependsOn", "access", "writeScopes", "requiredSkills",
		"verification", "humanGates", "maxAttempts", "maxCorrections",
	] as const;
	const children = root.children.map((child, index) => {
		const exact = exactRecord(child, childKeys, `production plan proposal child ${index}`);
		return { ...exact, issue: index + 1 };
	});
	const candidate = {
		schemaVersion: 1 as const,
		sourceRevisionDigest: facts.revisionDigest,
		title: safeText(root.title, "proposal title", 256),
		objective: safeText(root.objective, "proposal objective", 4_096),
		children: children.map(({ issue: _issue, ...child }) => child) as Array<Omit<ProductionChildSpec, "issue">>,
	};
	const validated = syntheticPlan(candidate, facts, authority, facts.viewer.login, "2099-01-01T00:00:00.000Z");
	const proposal: ProductionParentPlanProposal = {
		schemaVersion: 1,
		sourceRevisionDigest: facts.revisionDigest,
		title: validated.title,
		objective: validated.objective,
		children: validated.children.map(({ issue: _issue, ...child }) => child),
	};
	assertAcyclic(proposal.children);
	return proposal;
}

export class ProductionPlanBootstrapper {
	readonly #options: ProductionPlanBootstrapperOptions;
	readonly #now: () => Date;
	readonly #timeoutMs: number;
	readonly #tails = new Map<number, Promise<ProductionPlanSnapshot>>();

	constructor(options: ProductionPlanBootstrapperOptions) {
		if (typeof options !== "object" || options === null || !isAbsolute(options.repositoryRoot)
			|| !isAbsolute(options.stateRoot) || typeof options.intake?.tryLoad !== "function"
			|| typeof options.intake.publish !== "function" || typeof options.issueSource?.observe !== "function"
			|| typeof options.authoritySource?.observe !== "function" || typeof options.planSession?.propose !== "function"
			|| typeof options.github?.createPlan !== "function" || typeof options.github.ensureChildIssue !== "function") {
			throw new Error("production plan bootstrap options are invalid");
		}
		this.#options = options;
		this.#now = options.now ?? (() => new Date());
		this.#timeoutMs = options.timeoutMs ?? DEFAULT_TIMEOUT_MS;
		if (!Number.isSafeInteger(this.#timeoutMs) || this.#timeoutMs < 1 || this.#timeoutMs > 60 * 60 * 1_000) {
			throw new Error("production plan bootstrap timeout is invalid");
		}
	}

	ensure(issue: number, signal: AbortSignal): Promise<ProductionPlanSnapshot> {
		if (!Number.isSafeInteger(issue) || issue < 1) return Promise.reject(new Error("issue must be positive"));
		const active = this.#tails.get(issue);
		if (active !== undefined) return active;
		const pending = this.#ensure(issue, signal);
		this.#tails.set(issue, pending);
		void pending.finally(() => { if (this.#tails.get(issue) === pending) this.#tails.delete(issue); }).catch(() => undefined);
		return pending;
	}

	async #ensure(issue: number, signal: AbortSignal): Promise<ProductionPlanSnapshot> {
		assertActive(signal);
		const existing = await this.#options.intake.tryLoad(issue, signal);
		if (existing !== undefined) return existing;
		const deadlineAt = new Date(Date.now() + this.#timeoutMs).toISOString();
		const context = { signal, deadlineAt };
		const facts = validateProductionPlanningIssueFacts(
			await this.#options.issueSource.observe({
				repositoryRoot: resolve(this.#options.repositoryRoot),
				parentIssue: issue,
			}, context),
			issue,
		);
		assertActive(signal);
		const authority = validateAuthority(await this.#options.authoritySource.observe(issue, facts, signal), facts);
		const journal = await this.#loadJournal(issue);
		let durable: BootstrapJournal;
		if (journal === undefined) {
			const proposal = validateProductionParentPlanProposal(
				await this.#options.planSession.propose({ facts, authority }, context),
				facts,
				authority,
			);
			const expires = new Date(this.#now().valueOf() + 7 * 24 * 60 * 60 * 1_000).toISOString();
			durable = await this.#publishJournal({
				schemaVersion: 1,
				parentIssue: issue,
				repository: authority.repository,
				parentBranch: authority.parentBranch,
				parentBaseBranch: authority.parentBaseBranch,
				candidateHead: authority.candidateHead,
				sourceRevisionDigest: facts.revisionDigest,
				actor: facts.viewer.login,
				decisionExpiresAt: expires,
				proposal,
			});
		} else {
			durable = journal;
			this.#assertJournal(durable, facts, authority);
		}
		this.#assertJournal(durable, facts, authority);
		const scaffold = syntheticPlan(
			durable.proposal,
			facts,
			authority,
			durable.actor,
			durable.decisionExpiresAt,
		);
		const orchestration = await this.#options.github.createPlan(
			productionOrchestrationObjective(scaffold, 1),
			context,
		);
		const assignments: number[] = [];
		for (const child of durable.proposal.children) {
			assertActive(signal);
			const materialized = await this.#options.github.ensureChildIssue(orchestration, child.id, context);
			const number = positiveInteger(materialized.number, `materialized issue for ${child.id}`);
			if (number === issue || assignments.includes(number)) throw new Error("materialized child issue numbers conflict");
			assignments.push(number);
		}
		const plan = validateProductionParentPlan({
			...scaffold,
			children: scaffold.children.map((child, index) => ({ ...child, issue: assignments[index] })),
		}, issue);
		return this.#options.intake.publish(issue, plan, signal);
	}

	async close(): Promise<void> {
		await this.#options.github.stop({ deadlineAt: new Date(Date.now() + 5_000).toISOString() });
	}

	#journalPath(issue: number): string {
		return join(resolve(this.#options.stateRoot), "plan-bootstrap", `issue-${issue}.json`);
	}

	async #loadJournal(issue: number): Promise<BootstrapJournal | undefined> {
		const path = this.#journalPath(issue);
		try {
			const metadata = await lstat(path);
			if (!metadata.isFile() || metadata.isSymbolicLink() || metadata.nlink !== 1 || metadata.size > MAX_FACTS_BYTES) {
				throw new Error("production plan bootstrap journal is unsafe");
			}
			return JSON.parse(await readFile(path, "utf8")) as BootstrapJournal;
		} catch (error) {
			if (isMissing(error)) return undefined;
			throw error;
		}
	}

	#assertJournal(
		journal: BootstrapJournal,
		facts: ProductionPlanningIssueFacts,
		authority: ProductionPlanAuthority,
	): void {
		const root = exactRecord(journal, [
			"schemaVersion", "parentIssue", "repository", "parentBranch", "parentBaseBranch",
			"candidateHead", "sourceRevisionDigest", "actor", "decisionExpiresAt", "proposal",
		], "production plan bootstrap journal");
		if (root.schemaVersion !== 1 || root.parentIssue !== authority.parentIssue
			|| root.repository !== authority.repository || root.parentBranch !== authority.parentBranch
			|| root.parentBaseBranch !== authority.parentBaseBranch || root.candidateHead !== authority.candidateHead
			|| root.sourceRevisionDigest !== facts.revisionDigest || root.actor !== facts.viewer.login
			|| safeTimestamp(root.decisionExpiresAt, "bootstrap decision expiry") !== root.decisionExpiresAt) {
			throw new Error("production plan bootstrap authority or source changed; explicit replan required");
		}
		validateProductionParentPlanProposal(root.proposal, facts, authority);
	}

	async #publishJournal(document: BootstrapJournal): Promise<BootstrapJournal> {
		const root = resolve(this.#options.stateRoot);
		await mkdir(root, { recursive: true, mode: 0o700 });
		const rootMetadata = await lstat(root);
		if (!rootMetadata.isDirectory() || rootMetadata.isSymbolicLink()) throw new Error("bootstrap state root is unsafe");
		const canonicalRoot = await realpath(root);
		const directory = join(canonicalRoot, "plan-bootstrap");
		await mkdir(directory, { mode: 0o700 }).catch((error) => { if (!isMissing(error) && !("code" in (error as object) && (error as { code?: string }).code === "EEXIST")) throw error; });
		const directoryMetadata = await lstat(directory);
		if (!directoryMetadata.isDirectory() || directoryMetadata.isSymbolicLink()) throw new Error("bootstrap journal directory is unsafe");
		const path = join(directory, `issue-${document.parentIssue}.json`);
		const relativePath = relative(canonicalRoot, path);
		if (relativePath.startsWith("..") || relativePath.includes(`..${sep}`)) throw new Error("bootstrap journal escapes state root");
		const temporary = join(directory, `.${document.parentIssue}.${process.pid}.${randomUUID()}.tmp`);
		const handle = await open(temporary, constants.O_WRONLY | constants.O_CREAT | constants.O_EXCL, 0o600);
		try {
			await handle.writeFile(`${JSON.stringify(document)}\n`, "utf8");
			await handle.sync();
		} finally {
			await handle.close();
		}
		try {
			await link(temporary, path);
			const directoryHandle = await open(directory, constants.O_RDONLY);
			try { await directoryHandle.sync(); } finally { await directoryHandle.close(); }
			return document;
		} catch (error) {
			if (!(typeof error === "object" && error !== null && "code" in error && error.code === "EEXIST")) throw error;
			const raced = await this.#loadJournal(document.parentIssue);
			if (raced === undefined || canonicalDigest(raced) !== canonicalDigest(document)) {
				throw new Error("concurrent production plan proposals conflict");
			}
			return raced;
		} finally {
			await unlink(temporary).catch((error) => { if (!isMissing(error)) throw error; });
		}
	}
}

export interface AgentSessionProductionPlanSessionOptions {
	repositoryRoot: string;
	agentSession: ProductionAgentSessionPort;
}

export class AgentSessionProductionPlanSession implements ProductionPlanSession {
	readonly #repositoryRoot: string;
	readonly #agentSession: ProductionAgentSessionPort;

	constructor(options: AgentSessionProductionPlanSessionOptions) {
		if (typeof options !== "object" || options === null || !isAbsolute(options.repositoryRoot)
			|| typeof options.agentSession?.run !== "function" || typeof options.agentSession.abort !== "function") {
			throw new Error("production planning session options are invalid");
		}
		this.#repositoryRoot = resolve(options.repositoryRoot);
		this.#agentSession = options.agentSession;
	}

	async propose(
		input: { facts: ProductionPlanningIssueFacts; authority: ProductionPlanAuthority },
		context: ProductionPlanningCallContext,
	): Promise<ProductionParentPlanProposal> {
		const facts = validateProductionPlanningIssueFacts(input.facts, input.authority.parentIssue);
		const authority = validateAuthority(input.authority, facts);
		assertActive(context.signal);
		let captured: ProductionParentPlanProposal | undefined;
		let conflict = false;
		const capability: HostCapability = {
			name: "host_inspect",
			description: "Submit exactly one bounded issue-less Shepherd plan proposal for host validation.",
			mutates: false,
			parameters: proposalSchema(),
			async execute(raw): Promise<CapabilityResult> {
				const proposal = validateProductionParentPlanProposal(raw, facts, authority);
				if (captured !== undefined && canonicalDigest(captured) !== canonicalDigest(proposal)) {
					conflict = true;
					return { status: "failed", summary: "conflicting plan proposal rejected" };
				}
				captured = proposal;
				return { status: "ok", summary: "semantic plan proposal validated and captured" };
			},
		};
		const workspaceId = `planning-${authority.parentIssue}-${facts.revisionDigest.slice(0, 12)}`;
		const binding = {
			runId: `plan-${authority.parentIssue}-${facts.revisionDigest.slice(0, 12)}`,
			generation: 1,
			laneId: "planning",
			candidateHead: authority.candidateHead,
			validationNonce: canonicalDigest({ authority, source: facts.revisionDigest }),
		};
		const request: RoleRunRequest = {
			role: "planning",
			task: "Read .shepherd/facts.json and relevant repository files, then submit one complete semantic plan through host_inspect. Do not invent child issue numbers or host authority fields.",
			context: [
				"Use strict dependency ordering and disjoint write scopes where work can run in parallel.",
				"Every child must be mutating, bounded, independently verifiable, and include required skills.",
				"After host_inspect accepts the proposal, return the typed completed handoff.",
			],
			timeoutMs: DEFAULT_TIMEOUT_MS,
			signal: context.signal,
			workspace: planningWorkspace(this.#repositoryRoot, workspaceId, facts),
			capabilities: [capability],
			authority: {
				issue: authority.parentIssue,
				branch: authority.parentBranch,
				readOnly: true,
				workspaceId,
				readPrefixes: ["."],
				writePrefixes: [],
				capabilityNames: ["host_inspect"],
			},
			binding,
		};
		const handoff = await this.#agentSession.run(request);
		assertActive(context.signal);
		if (conflict || captured === undefined || handoff.status !== "completed" || handoff.role !== "planning"
			|| handoff.runId !== binding.runId || handoff.generation !== binding.generation
			|| handoff.laneId !== binding.laneId || handoff.candidateHead !== binding.candidateHead
			|| handoff.validationNonce !== binding.validationNonce || handoff.observedMutation
			|| handoff.changedPaths.length !== 0) {
			throw new Error("planning AgentSession returned no unique validated proposal");
		}
		return captured;
	}
}

function proposalSchema(): Readonly<Record<string, unknown>> {
	const stringArray = { type: "array", maxItems: 64, items: { type: "string", maxLength: 4_096 } };
	const verification = {
		type: "object",
		additionalProperties: false,
		required: ["id", "executable", "args", "cwd", "timeoutMs", "maxOutputBytes"],
		properties: {
			id: { type: "string", minLength: 1, maxLength: 64 },
			executable: { type: "string", maxLength: 128 },
			args: { type: "array", maxItems: 256, items: { type: "string", maxLength: 4_096 } },
			cwd: { type: "string", maxLength: 4_096 },
			timeoutMs: { type: "integer", minimum: 1, maximum: 120_000 },
			maxOutputBytes: { type: "integer", minimum: 1_024, maximum: 4 * 1024 * 1024 },
		},
	};
	const child = {
		type: "object",
		additionalProperties: false,
		required: [
			"id", "title", "task", "slug", "dependsOn", "access", "writeScopes", "requiredSkills",
			"verification", "humanGates", "maxAttempts", "maxCorrections",
		],
		properties: {
			id: { type: "string", maxLength: 64 },
			title: { type: "string", maxLength: 256 },
			task: { type: "string", maxLength: 4_096 },
			slug: { type: "string", maxLength: 100 },
			dependsOn: stringArray,
			access: { type: "string", enum: ["mutating"] },
			writeScopes: stringArray,
			requiredSkills: { ...stringArray, minItems: 1 },
			verification: { type: "array", minItems: 1, maxItems: 64, items: verification },
			humanGates: stringArray,
			maxAttempts: { type: "integer", minimum: 1, maximum: 10 },
			maxCorrections: { type: "integer", minimum: 1, maximum: 5 },
		},
	};
	return {
		type: "object",
		additionalProperties: false,
		required: ["schemaVersion", "sourceRevisionDigest", "title", "objective", "children"],
		properties: {
			schemaVersion: { type: "integer", enum: [1] },
			sourceRevisionDigest: { type: "string", minLength: 64, maxLength: 64 },
			title: { type: "string", maxLength: 256 },
			objective: { type: "string", maxLength: 4_096 },
			children: { type: "array", minItems: 1, maxItems: MAX_CHILDREN, items: child },
		},
	};
}

function planningWorkspace(
	repositoryRoot: string,
	id: string,
	facts: ProductionPlanningIssueFacts,
): ScopedWorkspace {
	const factsJson = `${JSON.stringify(facts, null, 2)}\n`;
	return Object.freeze({
		id,
		cwd: repositoryRoot,
		async readText(
			path: string,
			options: { offset?: number; limit?: number; signal?: AbortSignal },
		) {
			assertActive(options.signal ?? new AbortController().signal);
			const normalized = validateScopedPath(path, ["."]);
			const offset = options.offset ?? 0;
			const limit = options.limit ?? MAX_READ_BYTES;
			if (!Number.isSafeInteger(offset) || offset < 0 || !Number.isSafeInteger(limit)
				|| limit < 1 || limit > MAX_READ_BYTES) throw new Error("planning read range is invalid");
			if (normalized === ".shepherd/facts.json") return factsJson.slice(offset, offset + limit);
			const root = await realpath(repositoryRoot);
			const target = resolve(root, normalized);
			const back = relative(root, target);
			if (back === ".." || back.startsWith(`..${sep}`)) throw new Error("planning read escapes repository");
			let current = root;
			for (const component of back.split(sep)) {
				if (component.length === 0) continue;
				current = resolve(current, component);
				const metadata = await lstat(current);
				if (metadata.isSymbolicLink()) throw new Error("planning read cannot traverse symlinks");
			}
			const metadata = await lstat(target);
			if (!metadata.isFile() || metadata.size > MAX_READ_BYTES) throw new Error("planning read requires a bounded regular file");
			const value = await readFile(target, "utf8");
			return value.slice(offset, offset + limit);
		},
		async editText() { throw new Error("planning workspace is read-only"); },
		async writeText() { throw new Error("planning workspace is read-only"); },
	});
}
