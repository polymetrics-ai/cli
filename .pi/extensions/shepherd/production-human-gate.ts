import {
	assertHumanDecisionBinding,
	validateHumanDecisionRecord,
	type HumanDecisionBinding,
	type HumanDecisionRecord,
} from "./human-decision.ts";
import type {
	ExternalCallContext,
	ParentDecisionBroker,
} from "./github-orchestrator.ts";
import type { GitHubDecisionRequest } from "./github-decision-broker.ts";
import {
	defaultGhOrchestrationExecutor,
	type GhOrchestrationExecutor,
} from "./gh-orchestration-transport.ts";

const REPOSITORY = /^[A-Za-z0-9][A-Za-z0-9._-]{0,99}\/[A-Za-z0-9][A-Za-z0-9._-]{0,99}$/u;
const SHA = /^[0-9a-f]{40}$/u;
const LOGIN = /^[a-z\d](?:[a-z\d-]{0,37}[a-z\d])?$/u;

export interface ProductionParentMergeRequest {
	requestId: string;
	repository: string;
	parentIssue: number;
	pullRequest: number;
	generation: number;
	headSha: string;
	actorAllowlist: string[];
	expiresAt: string;
	question: string;
}

export interface ProductionChildInterventionRequest {
	requestId: string;
	repository: string;
	childIssue: number;
	generation: number;
	reason: "retry_budget_exhausted" | "correction_budget_exhausted";
	actorAllowlist: string[];
	expiresAt: string;
	question: string;
}

export interface AuthoritativeParentMergeState {
	repository: string;
	pullRequest: number;
	headSha: string;
	state: "open" | "closed" | "merged";
	mergedAt: string | null;
	mergeCommitSha: string | null;
	revision: number;
	observedAt: string;
}

export interface ParentPullRequestMergeLookup {
	observeExactPullRequest(
		query: { repository: string; pullRequest: number; headSha: string },
		context: ExternalCallContext,
	): Promise<AuthoritativeParentMergeState>;
}

export type ProductionParentMergeObservation =
	| { status: "pending" }
	| { status: "rejected"; record: HumanDecisionRecord }
	| { status: "approved_waiting_for_merge"; record: HumanDecisionRecord; observation: AuthoritativeParentMergeState }
	| {
		status: "merged";
		record: HumanDecisionRecord;
		repository: string;
		pullRequest: number;
		headSha: string;
		mergedAt: string;
		mergeCommitSha: string;
		revision: number;
		observedAt: string;
	};

function positiveInteger(value: number, description: string): number {
	if (!Number.isSafeInteger(value) || value < 1) throw new Error(`${description} must be a positive integer`);
	return value;
}

function boundedText(value: string, description: string, maximum = 4_096): string {
	if (typeof value !== "string" || value.length === 0 || Buffer.byteLength(value) > maximum
		|| /[\u0000-\u001f\u007f-\u009f]/u.test(value)) {
		throw new Error(`${description} must be bounded safe text`);
	}
	return value;
}

function canonicalTimestamp(value: string, description: string): string {
	const date = new Date(value);
	if (!Number.isFinite(date.valueOf())) throw new Error(`${description} must be RFC3339`);
	return date.toISOString();
}

function parentRequest(value: ProductionParentMergeRequest): ProductionParentMergeRequest {
	if (!REPOSITORY.test(value.repository)) throw new Error("invalid parent merge repository");
	if (!SHA.test(value.headSha)) throw new Error("invalid parent merge exact head");
	const actors = [...value.actorAllowlist].map((actor) => actor.toLowerCase());
	if (actors.length === 0 || new Set(actors).size !== actors.length || actors.some((actor) => !LOGIN.test(actor))) {
		throw new Error("invalid parent merge actor allowlist");
	}
	return {
		requestId: boundedText(value.requestId, "parent merge request ID", 128),
		repository: value.repository,
		parentIssue: positiveInteger(value.parentIssue, "parent merge issue"),
		pullRequest: positiveInteger(value.pullRequest, "parent merge pull request"),
		generation: positiveInteger(value.generation, "parent merge generation"),
		headSha: value.headSha,
		actorAllowlist: actors,
		expiresAt: canonicalTimestamp(value.expiresAt, "parent merge expiry"),
		question: boundedText(value.question, "parent merge question"),
	};
}

function bindingFor(value: ProductionParentMergeRequest): HumanDecisionBinding {
	return {
		repository: value.repository,
		target: { kind: "pull_request", number: value.pullRequest },
		generation: value.generation,
		headSha: value.headSha,
	};
}

function decisionRequest(value: ProductionParentMergeRequest): GitHubDecisionRequest {
	return {
		...value,
		gate: "parent_merge",
		allowedOptions: ["approve-merge", "reject"],
	};
}

export function buildProductionChildInterventionDecisionRequest(
	value: ProductionChildInterventionRequest,
): GitHubDecisionRequest {
	if (!REPOSITORY.test(value.repository)) {
		throw new Error("invalid child intervention repository");
	}
	if (value.reason !== "retry_budget_exhausted" && value.reason !== "correction_budget_exhausted") {
		throw new Error("child intervention is restricted to exhausted durable budgets");
	}
	const actors = [...value.actorAllowlist].map((actor) => actor.toLowerCase());
	if (actors.length === 0 || new Set(actors).size !== actors.length || actors.some((actor) => !LOGIN.test(actor))) {
		throw new Error("invalid child intervention actor allowlist");
	}
	return {
		requestId: boundedText(value.requestId, "child intervention request ID", 128),
		gate: "scope",
		repository: value.repository,
		parentIssue: positiveInteger(value.childIssue, "child intervention issue"),
		// The broker's issue-gate route ignores this legacy coordinate; keep it exact and non-fabricated.
		pullRequest: positiveInteger(value.childIssue, "child intervention issue route"),
		generation: positiveInteger(value.generation, "child intervention generation"),
		allowedOptions: ["authorize-one-retry", "abort-child"],
		actorAllowlist: actors,
		expiresAt: canonicalTimestamp(value.expiresAt, "child intervention expiry"),
		question: `[${value.reason}] ${boundedText(value.question, "child intervention question")}`,
	};
}

function validateBoundRecord(
	recordValue: HumanDecisionRecord,
	request: ProductionParentMergeRequest,
): HumanDecisionRecord {
	const record = validateHumanDecisionRecord(recordValue);
	assertHumanDecisionBinding(record, bindingFor(request));
	if (record.requestId !== request.requestId || record.gate !== "parent_merge"
		|| record.allowedOptions.length !== 2 || record.allowedOptions[0] !== "approve-merge"
		|| record.allowedOptions[1] !== "reject") {
		throw new Error("parent merge decision record does not match the exact request binding");
	}
	return record;
}

export class ProductionHumanParentMergeGate {
	readonly #broker: ParentDecisionBroker;
	readonly #lookup: ParentPullRequestMergeLookup;

	constructor(broker: ParentDecisionBroker, lookup: ParentPullRequestMergeLookup) {
		this.#broker = broker;
		this.#lookup = lookup;
	}

	async request(value: ProductionParentMergeRequest, context: ExternalCallContext): Promise<HumanDecisionRecord> {
		const request = parentRequest(value);
		return validateBoundRecord(await this.#broker.request(decisionRequest(request), context), request);
	}

	async observe(value: ProductionParentMergeRequest, context: ExternalCallContext): Promise<ProductionParentMergeObservation> {
		const request = parentRequest(value);
		const binding = bindingFor(request);
		let record = validateBoundRecord(
			await this.#broker.poll(request.requestId, binding, { signal: context.signal }, context),
			request,
		);
		if (record.status === "pending" || record.status === "expired") return { status: "pending" };
		if (record.status !== "decided" && record.status !== "consumed") {
			throw new Error("parent merge decision is in an invalid state");
		}
		if (record.status !== "consumed") {
			record = validateBoundRecord(await this.#broker.consume(request.requestId, binding, context), request);
		}
		if (record.status !== "consumed" || record.decision === undefined) {
			throw new Error("parent merge decision was not durably consumed");
		}
		if (record.decision.option === "reject") return { status: "rejected", record };
		if (record.decision.option !== "approve-merge") throw new Error("invalid parent merge decision option");
		const observation = await this.#lookup.observeExactPullRequest({
			repository: request.repository,
			pullRequest: request.pullRequest,
			headSha: request.headSha,
		}, context);
		if (observation.repository !== request.repository || observation.pullRequest !== request.pullRequest
			|| observation.headSha !== request.headSha) {
			throw new Error("authoritative parent pull request moved from the approved exact head");
		}
		if (observation.state !== "merged") {
			if (observation.mergedAt !== null || observation.mergeCommitSha !== null) {
				throw new Error("ambiguous authoritative parent merge observation");
			}
			return { status: "approved_waiting_for_merge", record, observation };
		}
		if (observation.mergedAt === null || observation.mergeCommitSha === null || !SHA.test(observation.mergeCommitSha)) {
			throw new Error("merged parent observation lacks authoritative merge evidence");
		}
		return {
			status: "merged",
			record,
			repository: observation.repository,
			pullRequest: observation.pullRequest,
			headSha: observation.headSha,
			mergedAt: observation.mergedAt,
			mergeCommitSha: observation.mergeCommitSha,
			revision: observation.revision,
			observedAt: observation.observedAt,
		};
	}
}

function parseRecord(value: string): Record<string, unknown> {
	if (Buffer.byteLength(value) > 2 * 1024 * 1024) throw new Error("GitHub parent pull request output is oversized");
	let parsed: unknown;
	try { parsed = JSON.parse(value); } catch { throw new Error("GitHub returned malformed parent pull request JSON"); }
	if (typeof parsed !== "object" || parsed === null || Array.isArray(parsed)) {
		throw new Error("GitHub returned malformed parent pull request evidence");
	}
	return parsed as Record<string, unknown>;
}

export class GhParentPullRequestMergeLookup implements ParentPullRequestMergeLookup {
	readonly #execute: GhOrchestrationExecutor;
	readonly #now: () => Date;

	constructor(execute: GhOrchestrationExecutor = defaultGhOrchestrationExecutor, now: () => Date = () => new Date()) {
		this.#execute = execute;
		this.#now = now;
	}

	async observeExactPullRequest(
		query: { repository: string; pullRequest: number; headSha: string },
		context: ExternalCallContext,
	): Promise<AuthoritativeParentMergeState> {
		if (!REPOSITORY.test(query.repository) || !SHA.test(query.headSha)) throw new Error("invalid parent pull request lookup");
		const timeoutMs = Math.max(1, Math.min(15_000, new Date(context.deadlineAt).valueOf() - Date.now()));
		let output: string;
		try {
			output = await this.#execute("gh", ["api", "--method", "GET", `/repos/${query.repository}/pulls/${positiveInteger(query.pullRequest, "pull request")}`], {
				signal: context.signal,
				timeoutMs,
				maxOutputBytes: 2 * 1024 * 1024,
			});
		} finally {
			if (context.signal.aborted) context.acknowledgeAbort();
		}
		const raw = parseRecord(output);
		const head = typeof raw.head === "object" && raw.head !== null ? raw.head as Record<string, unknown> : {};
		const headSha = head.sha;
		if (headSha !== query.headSha) throw new Error("authoritative parent pull request head moved");
		const mergedAt = raw.merged_at === null ? null : canonicalTimestamp(String(raw.merged_at), "parent merge time");
		const merged = raw.merged === true || mergedAt !== null;
		const state = merged ? "merged" : raw.state === "open" ? "open" : "closed";
		const observedAt = this.#now().toISOString();
		return {
			repository: query.repository,
			pullRequest: query.pullRequest,
			headSha: query.headSha,
			state,
			mergedAt,
			mergeCommitSha: merged ? boundedText(String(raw.merge_commit_sha), "merge commit SHA", 40) : null,
			revision: Math.max(1, Math.floor(new Date(canonicalTimestamp(String(raw.updated_at), "parent pull request update time")).valueOf() / 1_000)),
			observedAt,
		};
	}
}
