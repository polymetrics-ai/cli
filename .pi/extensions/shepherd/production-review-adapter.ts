import { createHash } from "node:crypto";

import {
	ShepherdAgentSessionRuntime,
	type RoleRunRequest,
} from "./agent-session-runtime.ts";
import type { GitHubChangedPathEvidence, GitHubFindingDisposition } from "./github-evidence.ts";
import {
	defaultGhOrchestrationExecutor,
	type GhOrchestrationExecutor,
} from "./gh-orchestration-transport.ts";
import type {
	AgentSessionAttestationSource,
	AuthoritativeLookup,
	ExternalCallContext,
} from "./github-orchestrator.ts";
import {
	createAgentSessionAttestation,
	createIndependentReviewWork,
	independentReviewResultDigest,
	reconcileIndependentReview,
	validateAgentSessionAttestation,
	validateIndependentReviewRecord,
	type AgentSessionAttestation,
	type IndependentReviewFinding,
	type IndependentReviewRecord,
	type IndependentReviewTarget,
	type IndependentReviewWork,
} from "./review-router.ts";

const SHA = /^[0-9a-f]{40}$/u;
const REVIEW_PREFIX = "<!-- shepherd-production-review:v1:";
const REVIEW_PENDING_PREFIX = "<!-- shepherd-production-review-pending:v1:";

export type ProductionReviewDisposition = GitHubFindingDisposition;

export interface ProductionReviewArtifact {
	schemaVersion: 1;
	review: IndependentReviewRecord;
	attestation: AgentSessionAttestation;
	dispositions: ProductionReviewDisposition[];
	revision: number;
	publishedAt: string;
}

export interface ProductionReviewRepository {
	find(target: IndependentReviewTarget, context: ExternalCallContext): Promise<AuthoritativeLookup<ProductionReviewArtifact>>;
	publish(
		artifact: Omit<ProductionReviewArtifact, "revision" | "publishedAt">,
		context: ExternalCallContext,
	): Promise<ProductionReviewArtifact>;
	recordDispositions(
		target: IndependentReviewTarget,
		dispositions: readonly ProductionReviewDisposition[],
		context: ExternalCallContext,
	): Promise<ProductionReviewArtifact>;
}

export interface GhProductionReviewRepositoryOptions {
	execute?: GhOrchestrationExecutor;
	now?: () => Date;
	timeoutMs?: number;
	maxOutputBytes?: number;
	maxPages?: number;
}

export interface ProductionReviewSessionResult {
	sessionId: string;
	runId: string;
	completedAt: string;
	verdict: "clean" | "findings";
	findings: IndependentReviewFinding[];
}

export interface ProductionReviewSession {
	run(work: IndependentReviewWork, context: ExternalCallContext): Promise<ProductionReviewSessionResult>;
}

export interface ProductionChangedPathEvidenceSource {
	findChangedPathEvidence(
		query: Omit<IndependentReviewTarget, "changedPaths" | "allowedScopes">,
		context: ExternalCallContext,
	): Promise<AuthoritativeLookup<GitHubChangedPathEvidence>>;
}

function canonicalTimestamp(value: string, description: string): string {
	const date = new Date(value);
	if (!Number.isFinite(date.valueOf()) || date.toISOString() !== value) throw new Error(`${description} must be canonical RFC3339`);
	return value;
}

function sameTarget(review: IndependentReviewRecord, work: IndependentReviewWork): boolean {
	return review.repository === work.repository && review.workItemId === work.workItemId
		&& review.pullRequest === work.pullRequest && review.generation === work.generation
		&& review.baseBranch === work.baseBranch && review.headBranch === work.headBranch
		&& review.baseSha === work.baseSha && review.headSha === work.headSha
		&& review.idempotencyMarker === work.idempotencyMarker
		&& JSON.stringify(review.changedPaths) === JSON.stringify(work.changedPaths)
		&& JSON.stringify(review.allowedScopes) === JSON.stringify(work.allowedScopes);
}

function canonicalDisposition(value: ProductionReviewDisposition, review: IndependentReviewRecord): ProductionReviewDisposition {
	if (!review.findings.some((finding) => finding.id === value.findingId)) {
		throw new Error("review disposition does not reference an authoritative finding");
	}
	if (value.kind !== "fixed" && value.kind !== "not_actionable") throw new Error("invalid review disposition kind");
	if (typeof value.rationale !== "string" || value.rationale.length === 0 || Buffer.byteLength(value.rationale) > 2_048) {
		throw new Error("invalid review disposition rationale");
	}
	if (value.headSha !== review.headSha || !SHA.test(value.headSha)) {
		throw new Error("review disposition is stale for the exact reviewed head");
	}
	return {
		findingId: value.findingId,
		kind: value.kind,
		rationale: value.rationale,
		actor: value.actor,
		headSha: value.headSha,
		recordedAt: canonicalTimestamp(value.recordedAt, "review disposition time"),
	};
}

function canonicalArtifact(value: ProductionReviewArtifact): ProductionReviewArtifact {
	const review = validateIndependentReviewRecord(value.review);
	const attestation = validateAgentSessionAttestation(value.attestation, review);
	if (!Number.isSafeInteger(value.revision) || value.revision < 1) throw new Error("invalid production review revision");
	const dispositions = value.dispositions.map((entry) => canonicalDisposition(entry, review));
	if (new Set(dispositions.map((entry) => entry.findingId)).size !== dispositions.length) {
		throw new Error("duplicate production review disposition");
	}
	return {
		schemaVersion: 1,
		review,
		attestation,
		dispositions,
		revision: value.revision,
		publishedAt: canonicalTimestamp(value.publishedAt, "production review publication time"),
	};
}

function encodeArtifact(value: ProductionReviewArtifact): string {
	return `${REVIEW_PREFIX}${Buffer.from(JSON.stringify(value)).toString("base64url")} -->`;
}

function decodeArtifact(value: unknown): ProductionReviewArtifact | null {
	if (typeof value !== "string" || !value.startsWith(REVIEW_PREFIX) || !value.endsWith(" -->")) return null;
	const encoded = value.slice(REVIEW_PREFIX.length, -4);
	if (!/^[A-Za-z0-9_-]+$/u.test(encoded) || encoded.length > 1_000_000) throw new Error("invalid durable review marker envelope");
	let parsed: unknown;
	try { parsed = JSON.parse(Buffer.from(encoded, "base64url").toString("utf8")); }
	catch { throw new Error("invalid durable review marker envelope"); }
	return canonicalArtifact(parsed as ProductionReviewArtifact);
}

function ghRecord(value: unknown, description: string): Record<string, unknown> {
	if (typeof value !== "object" || value === null || Array.isArray(value)) throw new Error(`GitHub returned malformed ${description}`);
	return value as Record<string, unknown>;
}

/** Durable GitHub issue-comment repository; exact markers survive process restart and reconcile uncertain publication. */
export class GhProductionReviewRepository implements ProductionReviewRepository {
	readonly #execute: GhOrchestrationExecutor;
	readonly #now: () => Date;
	readonly #timeoutMs: number;
	readonly #maxOutputBytes: number;
	readonly #maxPages: number;

	constructor(options: GhProductionReviewRepositoryOptions = {}) {
		this.#execute = options.execute ?? defaultGhOrchestrationExecutor;
		this.#now = options.now ?? (() => new Date());
		this.#timeoutMs = options.timeoutMs ?? 15_000;
		this.#maxOutputBytes = options.maxOutputBytes ?? 2 * 1024 * 1024;
		this.#maxPages = options.maxPages ?? 10;
		for (const [value, maximum, name] of [
			[this.#timeoutMs, 120_000, "timeout"],
			[this.#maxOutputBytes, 8 * 1024 * 1024, "output limit"],
			[this.#maxPages, 100, "page limit"],
		] as const) {
			if (!Number.isSafeInteger(value) || value < 1 || value > maximum) throw new Error(`invalid production review ${name}`);
		}
	}

	async #api(
		method: "GET" | "POST" | "PATCH",
		endpoint: string,
		context: ExternalCallContext,
		body?: string,
	): Promise<unknown> {
		const deadline = new Date(context.deadlineAt).valueOf();
		try {
			const output = await this.#execute("gh", [
				"api", "--method", method, endpoint,
				...(body === undefined ? [] : ["-f", `body=${body}`]),
			], {
				signal: context.signal,
				timeoutMs: Math.max(1, Math.min(this.#timeoutMs, deadline - Date.now())),
				maxOutputBytes: this.#maxOutputBytes,
			});
			if (Buffer.byteLength(output) > this.#maxOutputBytes) throw new Error("GitHub review repository output is oversized");
			return JSON.parse(output);
		} catch (error) {
			if (context.signal.aborted) {
				context.acknowledgeAbort();
				throw new Error("production review repository call cancelled");
			}
			throw new Error(method === "GET"
				? "production review repository lookup failed"
				: "production review repository mutation publication is uncertain");
		}
	}

	async #comments(target: IndependentReviewTarget, context: ExternalCallContext) {
		const comments: Record<string, unknown>[] = [];
		for (let page = 1; page <= this.#maxPages; page += 1) {
			const payload = await this.#api("GET",
				`/repos/${target.repository}/issues/${target.pullRequest}/comments?per_page=100&page=${page}`, context);
			if (!Array.isArray(payload) || payload.length > 100) throw new Error("GitHub returned malformed review comments");
			comments.push(...payload.map((entry) => ghRecord(entry, "review comment")));
			if (payload.length < 100) return { comments, complete: true };
		}
		return { comments, complete: false };
	}

	async find(target: IndependentReviewTarget, context: ExternalCallContext): Promise<AuthoritativeLookup<ProductionReviewArtifact>> {
		const work = createIndependentReviewWork(target);
		const result = await this.#comments(target, context);
		const items = result.comments.map((comment) => ({ artifact: decodeArtifact(comment.body), comment }))
			.filter((entry): entry is { artifact: ProductionReviewArtifact; comment: Record<string, unknown> } => entry.artifact !== null)
			.map(({ artifact, comment }) => {
				if (artifact.revision !== Number(comment.id)) throw new Error("durable review artifact revision is not bound to its GitHub comment");
				return artifact;
			})
			.filter((artifact) => artifact.review.repository === work.repository
				&& artifact.review.pullRequest === work.pullRequest);
		const identities = new Set<string>();
		for (const artifact of items) {
			const identity = `${artifact.review.idempotencyMarker}\u0000${artifact.review.completedAt}\u0000${independentReviewResultDigest(artifact.review)}`;
			if (identities.has(identity)) throw new Error("duplicate durable production review artifact is ambiguous");
			identities.add(identity);
		}
		return { items, complete: result.complete };
	}

	async publish(
		artifactValue: Omit<ProductionReviewArtifact, "revision" | "publishedAt">,
		context: ExternalCallContext,
	): Promise<ProductionReviewArtifact> {
		const review = validateIndependentReviewRecord(artifactValue.review);
		const target: IndependentReviewTarget = {
			repository: review.repository,
			workItemId: review.workItemId,
			pullRequest: review.pullRequest,
			generation: review.generation,
			baseBranch: review.baseBranch,
			headBranch: review.headBranch,
			baseSha: review.baseSha,
			headSha: review.headSha,
			changedPaths: review.changedPaths,
			allowedScopes: review.allowedScopes,
		};
		const before = await this.find(target, context);
		const digest = independentReviewResultDigest(review);
		const identical = before.items.filter((entry) => independentReviewResultDigest(entry.review) === digest);
		if (identical.length === 1) return identical[0];
		if (identical.length > 1 || before.items.some((entry) => entry.review.idempotencyMarker === review.idempotencyMarker
			&& entry.review.completedAt === review.completedAt)) {
			throw new Error("ambiguous durable production review publication");
		}
		const pendingBody = `${REVIEW_PENDING_PREFIX}${Buffer.from(`${review.idempotencyMarker}\u0000${digest}`).toString("base64url")} -->`;
		let pending = (await this.#comments(target, context)).comments.find((comment) => comment.body === pendingBody);
		if (pending === undefined) {
			try {
				pending = ghRecord(await this.#api("POST", `/repos/${review.repository}/issues/${review.pullRequest}/comments`, context, pendingBody), "review comment");
			} catch (error) {
				pending = (await this.#comments(target, context)).comments.find((comment) => comment.body === pendingBody);
				if (pending === undefined) throw error;
			}
		}
		const revision = Number(pending.id);
		const artifact = canonicalArtifact({
			...artifactValue,
			schemaVersion: 1,
			revision,
			publishedAt: this.#now().toISOString(),
		});
		try {
			await this.#api("PATCH", `/repos/${review.repository}/issues/comments/${revision}`, context, encodeArtifact(artifact));
		} catch (error) {
			const recovered = await this.find(target, context);
			const match = recovered.items.find((entry) => independentReviewResultDigest(entry.review) === digest);
			if (match !== undefined) return match;
			throw error;
		}
		return artifact;
	}

	async recordDispositions(
		target: IndependentReviewTarget,
		dispositions: readonly ProductionReviewDisposition[],
		context: ExternalCallContext,
	): Promise<ProductionReviewArtifact> {
		const work = createIndependentReviewWork(target);
		const found = await this.find(target, context);
		if (!found.complete) throw new Error("durable production review lookup is incomplete");
		const exact = found.items.filter((artifact) => sameTarget(artifact.review, work))
			.sort((left, right) => right.review.completedAt.localeCompare(left.review.completedAt));
		if (exact.length === 0) throw new Error("exact-head durable production review does not exist");
		const current = exact[0];
		const updated = canonicalArtifact({
			...current,
			dispositions: dispositions.map((entry) => canonicalDisposition(entry, current.review)),
			publishedAt: this.#now().toISOString(),
		});
		try {
			await this.#api("PATCH", `/repos/${current.review.repository}/issues/comments/${current.revision}`, context, encodeArtifact(updated));
		} catch (error) {
			const recovered = await this.find(target, context);
			const match = recovered.items.find((artifact) => artifact.revision === current.revision
				&& JSON.stringify(artifact.dispositions) === JSON.stringify(updated.dispositions));
			if (match !== undefined) return match;
			throw error;
		}
		return updated;
	}
}

export class MemoryProductionReviewRepository implements ProductionReviewRepository {
	readonly #artifacts: ProductionReviewArtifact[] = [];
	#revision = 0;

	async find(target: IndependentReviewTarget): Promise<AuthoritativeLookup<ProductionReviewArtifact>> {
		const work = createIndependentReviewWork(target);
		return {
			items: this.#artifacts.filter((artifact) => artifact.review.repository === work.repository
				&& artifact.review.pullRequest === work.pullRequest).map(canonicalArtifact),
			complete: true,
		};
	}

	async publish(artifact: Omit<ProductionReviewArtifact, "revision" | "publishedAt">): Promise<ProductionReviewArtifact> {
		const review = validateIndependentReviewRecord(artifact.review);
		const existing = this.#artifacts.find((entry) => entry.review.idempotencyMarker === review.idempotencyMarker
			&& independentReviewResultDigest(entry.review) === independentReviewResultDigest(review));
		if (existing !== undefined) return canonicalArtifact(existing);
		const conflicting = this.#artifacts.filter((entry) => entry.review.idempotencyMarker === review.idempotencyMarker
			&& entry.review.completedAt === review.completedAt);
		if (conflicting.length > 0) throw new Error("ambiguous production review publication");
		const value = canonicalArtifact({
			...artifact,
			schemaVersion: 1,
			revision: ++this.#revision,
			publishedAt: new Date().toISOString(),
		});
		this.#artifacts.push(value);
		return canonicalArtifact(value);
	}

	async recordDispositions(
		target: IndependentReviewTarget,
		dispositions: readonly ProductionReviewDisposition[],
		_context?: ExternalCallContext,
	): Promise<ProductionReviewArtifact> {
		const work = createIndependentReviewWork(target);
		const matches = this.#artifacts.filter((artifact) => sameTarget(artifact.review, work))
			.sort((left, right) => right.review.completedAt.localeCompare(left.review.completedAt));
		if (matches.length === 0) throw new Error("exact-head production review does not exist");
		const current = matches[0];
		const next = canonicalArtifact({
			...current,
			dispositions: dispositions.map((entry) => canonicalDisposition(entry, current.review)),
			revision: ++this.#revision,
			publishedAt: new Date().toISOString(),
		});
		const index = this.#artifacts.indexOf(current);
		this.#artifacts[index] = next;
		return canonicalArtifact(next);
	}
}

export type ProductionReviewRoleRequestFactory = (
	work: IndependentReviewWork,
	context: ExternalCallContext,
) => RoleRunRequest;

/** Concrete bridge from the exact-range review port to the embedded Pi AgentSession runtime. */
export class EmbeddedAgentSessionProductionReviewSession implements ProductionReviewSession {
	readonly #runtime: ShepherdAgentSessionRuntime;
	readonly #request: ProductionReviewRoleRequestFactory;
	readonly #now: () => Date;

	constructor(
		runtime: ShepherdAgentSessionRuntime,
		request: ProductionReviewRoleRequestFactory,
		now: () => Date = () => new Date(),
	) {
		this.#runtime = runtime;
		this.#request = request;
		this.#now = now;
	}

	async run(work: IndependentReviewWork, context: ExternalCallContext): Promise<ProductionReviewSessionResult> {
		const request = this.#request(work, context);
		if (request.role !== "review" || request.authority.readOnly !== true
			|| request.binding.generation !== work.generation || request.binding.candidateHead !== work.headSha) {
			throw new Error("production review AgentSession request is not exact-range read-only review authority");
		}
		const handoff = await this.#runtime.run({ ...request, signal: context.signal });
		if (handoff.status !== "completed" || handoff.observedMutation) {
			throw new Error("independent review AgentSession did not complete read-only");
		}
		const findings = handoff.findings.map((summary, index): IndependentReviewFinding => ({
			id: `AS-${createHash("sha256").update(`${work.idempotencyMarker}\u0000${index}\u0000${summary}`).digest("hex").slice(0, 16)}`,
			severity: "blocking",
			summary,
		}));
		return {
			sessionId: handoff.runId,
			runId: handoff.runId,
			completedAt: this.#now().toISOString(),
			verdict: findings.length === 0 ? "clean" : "findings",
			findings,
		};
	}
}

export class AgentSessionProductionReviewAdapter implements AgentSessionAttestationSource {
	readonly #repository: ProductionReviewRepository;
	readonly #session: ProductionReviewSession;
	readonly #changedPaths?: ProductionChangedPathEvidenceSource;

	constructor(
		repository: ProductionReviewRepository,
		session: ProductionReviewSession,
		changedPaths?: ProductionChangedPathEvidenceSource,
	) {
		this.#repository = repository;
		this.#session = session;
		this.#changedPaths = changedPaths;
	}

	async review(target: IndependentReviewTarget, context: ExternalCallContext): Promise<ProductionReviewArtifact> {
		const work = createIndependentReviewWork(target);
		const existing = await this.#repository.find(target, context);
		if (!existing.complete) throw new Error("production review repository lookup is incomplete");
		const decision = reconcileIndependentReview({
			target,
			reviews: existing.items.map((artifact) => artifact.review),
			attestations: existing.items.map((artifact) => artifact.attestation),
		});
		if (decision.kind === "satisfied") {
			const exact = existing.items.filter((artifact) => independentReviewResultDigest(artifact.review)
				=== independentReviewResultDigest(decision.review));
			if (exact.length !== 1) throw new Error("durable production review is absent or ambiguous");
			return canonicalArtifact(exact[0]);
		}
		const result = await this.#session.run(work, context);
		const review = validateIndependentReviewRecord({
			...work,
			completedAt: canonicalTimestamp(result.completedAt, "AgentSession review completion time"),
			verdict: result.verdict,
			findings: result.findings,
		});
		const attestation = createAgentSessionAttestation({
			sessionId: result.sessionId,
			runId: result.runId,
			review,
		});
		return this.#repository.publish({ schemaVersion: 1, review, attestation, dispositions: [] }, context);
	}

	async findAttestations(
		target: IndependentReviewTarget,
		context: ExternalCallContext,
	): Promise<AuthoritativeLookup<AgentSessionAttestation>> {
		const result = await this.#repository.find(target, context);
		return {
			items: result.items.filter((artifact) => sameTarget(artifact.review, createIndependentReviewWork(target)))
				.map((artifact) => validateAgentSessionAttestation(artifact.attestation, artifact.review)),
			complete: result.complete,
		};
	}

	async findChangedPathEvidence(
		query: Omit<IndependentReviewTarget, "changedPaths" | "allowedScopes">,
		context: ExternalCallContext,
	): Promise<AuthoritativeLookup<GitHubChangedPathEvidence>> {
		if (this.#changedPaths === undefined) throw new Error("authoritative changed-path source is required for production composition");
		return this.#changedPaths.findChangedPathEvidence(query, context);
	}
}
