import { types as nodeTypes } from "node:util";

import {
	createIndependentReviewWork,
	reconcileIndependentReview,
	validateIndependentReviewRecord,
	type AgentSessionAttestation,
	type IndependentReviewRecord,
	type IndependentReviewTarget,
} from "./review-router.ts";

const MAX_GITHUB_NUMBER = 2_147_483_647;
const MAX_COLLECTION = 128;
const MAX_REVIEWS = 64;
const MAX_BODY_BYTES = 65_536;
const SHA = /^[0-9a-f]{40}$/;
const MARKER = /^<!-- shepherd-(?:child|parent)-pr:v1:[A-Za-z0-9:._-]{1,300} -->$/;
const RFC3339_UTC = /^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})(?:\.(\d{1,3}))?Z$/;
const UNSAFE_INLINE = /[\u0000-\u001f\u007f-\u009f\u00ad\u061c\u200b-\u200f\u2028-\u202e\u2060-\u206f\ufeff]/u;
const UNSAFE_BODY = /[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f-\u009f\u00ad\u061c\u200b-\u200f\u2028-\u202e\u2060-\u206f\ufeff]/u;

export interface GitHubCheckEvidence {
	id: string;
	name: string;
	producerId: string;
	status: "queued" | "in_progress" | "completed";
	conclusion: "success" | "failure" | "cancelled" | "timed_out" | "action_required" | "neutral" | "skipped" | null;
	headSha: string;
	completedAt: string;
}

export interface RequiredGitHubCheck {
	name: string;
	producerId: string;
}

export interface GitHubRequestedChangeEvidence {
	id: string;
	actor: string;
	commitSha: string;
	dismissed: boolean;
	submittedAt: string;
}

export interface GitHubReviewThreadEvidence {
	id: string;
	resolved: boolean;
	headSha: string;
}

export interface GitHubFindingDisposition {
	findingId: string;
	kind: "fixed" | "not_actionable";
	rationale: string;
	actor: string;
	headSha: string;
	recordedAt: string;
}

export interface GitHubPullRequestEvidence {
	schemaVersion: 1;
	number: number;
	marker: string;
	title: string;
	body: string;
	state: "open" | "closed" | "merged";
	draft: boolean;
	baseBranch: string;
	headBranch: string;
	baseSha: string;
	headSha: string;
	changedPaths: string[];
	mergeState: "clean" | "blocked" | "behind" | "conflicting" | "unknown";
	checksComplete: boolean;
	checks: GitHubCheckEvidence[];
	requestedChanges: GitHubRequestedChangeEvidence[];
	threads: GitHubReviewThreadEvidence[];
	reviews: IndependentReviewRecord[];
	reviewsComplete: boolean;
	dispositions: GitHubFindingDisposition[];
	observedAt: string;
}

export interface ExpectedPullRequestEvidence {
	number: number;
	marker: string;
	baseBranch: string;
	headBranch: string;
	baseSha: string;
	headSha: string;
	changedPaths: readonly string[];
	requiredChecks: readonly RequiredGitHubCheck[];
	reviewTarget: IndependentReviewTarget;
	attestations: readonly AgentSessionAttestation[];
}

export type GitHubEvidenceBlocker =
	| "resource_mismatch"
	| "marker_collision"
	| "pr_not_open"
	| "draft"
	| "topology_mismatch"
	| "head_moved"
	| "merge_blocked"
	| "ci_not_green"
	| "changes_requested"
	| "unresolved_thread"
	| "undispositioned_finding"
	| "review_missing";

export type GitHubPullRequestDecision =
	| { kind: "eligible"; review: IndependentReviewRecord }
	| { kind: "blocked"; blockers: GitHubEvidenceBlocker[] };

interface PullRequestEvaluationOptions {
	allowDraft?: boolean;
}

type ExactRecord = Record<string, unknown>;

function exactRecord(value: unknown, required: readonly string[]): ExactRecord {
	if (typeof value !== "object" || value === null || Array.isArray(value) || nodeTypes.isProxy(value)) {
		throw new Error("invalid GitHub evidence shape");
	}
	const prototype = Object.getPrototypeOf(value);
	if (prototype !== Object.prototype && prototype !== null) throw new Error("invalid GitHub evidence shape");
	const descriptors = Object.getOwnPropertyDescriptors(value);
	if (Reflect.ownKeys(descriptors).length !== required.length) throw new Error("unknown GitHub evidence field");
	for (const key of required) {
		const descriptor = descriptors[key];
		if (descriptor === undefined || !Object.hasOwn(descriptor, "value") || descriptor.enumerable !== true) {
			throw new Error("invalid GitHub evidence shape");
		}
	}
	for (const key of Reflect.ownKeys(descriptors)) {
		if (typeof key !== "string" || !required.includes(key)) throw new Error("unknown GitHub evidence field");
	}
	return Object.fromEntries(Object.entries(descriptors).map(([key, descriptor]) => [key, descriptor.value]));
}

function inlineText(value: unknown, description: string, maximum = 1_024): string {
	if (typeof value !== "string" || value.length === 0 || value.length > maximum || Buffer.byteLength(value) > maximum
		|| value.trim() !== value || UNSAFE_INLINE.test(value)) {
		throw new Error(`invalid ${description}`);
	}
	return value;
}

function bodyText(value: unknown): string {
	if (typeof value !== "string" || value.length === 0 || value.length > MAX_BODY_BYTES
		|| Buffer.byteLength(value) > MAX_BODY_BYTES || UNSAFE_BODY.test(value)) {
		throw new Error("pull request body must be bounded safe text");
	}
	return value.replace(/\r\n?/gu, "\n");
}

function githubNumber(value: unknown, description: string): number {
	if (!Number.isSafeInteger(value) || (value as number) < 1 || (value as number) > MAX_GITHUB_NUMBER) {
		throw new Error(`invalid ${description}`);
	}
	return value as number;
}

function sha(value: unknown, description: string): string {
	if (typeof value !== "string" || !SHA.test(value)) throw new Error(`invalid ${description}`);
	return value;
}

function timestamp(value: unknown, description: string): string {
	if (typeof value !== "string" || value.length > 64) throw new Error(`invalid ${description}`);
	const match = RFC3339_UTC.exec(value);
	if (match === null) throw new Error(`invalid ${description}`);
	const canonical = `${match[1]}.${(match[2] ?? "").padEnd(3, "0")}Z`;
	const date = new Date(canonical);
	if (!Number.isFinite(date.valueOf()) || date.toISOString() !== canonical) throw new Error(`invalid ${description}`);
	return canonical;
}

export function canonicalGitRef(value: unknown, description = "Git ref"): string {
	const result = inlineText(value, description, 240);
	if (result.startsWith("-") || result.startsWith("/") || result.endsWith("/") || result.includes("\\")
		|| result.includes("..") || result.includes("//") || result.includes("@{") || result === "@"
		|| /[ ~^:?*\[\]{}]/u.test(result)
		|| result.split("/").some((segment) => segment === "" || segment === "." || segment === ".."
			|| segment.startsWith(".") || segment.endsWith(".") || segment.endsWith(".lock"))) {
		throw new Error(`invalid ${description}`);
	}
	return result;
}

function boundedArray(value: unknown, description: string, maximum = MAX_COLLECTION): unknown[] {
	if (!Array.isArray(value) || nodeTypes.isProxy(value) || Object.getPrototypeOf(value) !== Array.prototype) {
		throw new Error(`${description} must be a canonical array`);
	}
	const descriptors = Object.getOwnPropertyDescriptors(value);
	const lengthDescriptor = Object.getOwnPropertyDescriptor(value, "length");
	if (lengthDescriptor === undefined || !Object.hasOwn(lengthDescriptor, "value")
		|| !Number.isSafeInteger(lengthDescriptor.value) || lengthDescriptor.value < 0
		|| lengthDescriptor.value > maximum) {
		throw new Error(`${description} must be a bounded array of at most ${maximum} values`);
	}
	const length = lengthDescriptor.value as number;
	const values: unknown[] = [];
	for (const key of Reflect.ownKeys(descriptors)) {
		if (key === "length") continue;
		if (typeof key !== "string" || !/^(?:0|[1-9]\d*)$/u.test(key)) throw new Error(`${description} has an invalid array field`);
		const index = Number(key);
		const descriptor = descriptors[key];
		if (index >= length || !Object.hasOwn(descriptor, "value") || descriptor.enumerable !== true) {
			throw new Error(`${description} must contain only dense data values`);
		}
		values[index] = descriptor.value;
	}
	for (let index = 0; index < length; index += 1) {
		if (!Object.hasOwn(values, index)) throw new Error(`${description} must be a dense canonical array`);
	}
	return values;
}

function unique<T>(values: readonly T[], key: (value: T) => string, description: string): void {
	const keys = values.map(key);
	if (new Set(keys).size !== keys.length) throw new Error(`duplicate ${description}`);
}

function pathValue(value: unknown, description: string): string {
	const path = inlineText(value, description, 4_096).normalize("NFC");
	if (path.startsWith("/") || path.endsWith("/") || path.includes("\\") || /[*?\[\]{}]/u.test(path)
		|| path.split("/").some((segment) => segment === "" || segment === "." || segment === "..")) {
		throw new Error(`invalid ${description}`);
	}
	return path;
}

function stringArray(value: unknown, description: string, allowEmpty = false): string[] {
	const paths = boundedArray(value, description);
	if (!allowEmpty && paths.length === 0) throw new Error(`${description} must not be empty`);
	const canonical = paths.map((entry) => pathValue(entry, description));
	unique(canonical, (entry) => entry, description);
	return canonical.sort();
}

function validateCheck(value: unknown): GitHubCheckEvidence {
	const candidate = exactRecord(value, ["id", "name", "producerId", "status", "conclusion", "headSha", "completedAt"]);
	if (candidate.status !== "queued" && candidate.status !== "in_progress" && candidate.status !== "completed") {
		throw new Error("invalid check status");
	}
	const conclusions = ["success", "failure", "cancelled", "timed_out", "action_required", "neutral", "skipped", null];
	if (!conclusions.includes(candidate.conclusion as never)) throw new Error("invalid check conclusion");
	if (candidate.status === "completed" ? candidate.conclusion === null : candidate.conclusion !== null) {
		throw new Error("check status and conclusion disagree");
	}
	return {
		id: inlineText(candidate.id, "check ID"),
		name: inlineText(candidate.name, "check name"),
		producerId: inlineText(candidate.producerId, "check producer ID"),
		status: candidate.status,
		conclusion: candidate.conclusion as GitHubCheckEvidence["conclusion"],
		headSha: sha(candidate.headSha, "check head SHA"),
		completedAt: timestamp(candidate.completedAt, "check completion timestamp"),
	};
}

function validateRequiredCheck(value: unknown): RequiredGitHubCheck {
	const candidate = exactRecord(value, ["name", "producerId"]);
	return {
		name: inlineText(candidate.name, "required check name"),
		producerId: inlineText(candidate.producerId, "required check producer ID"),
	};
}

function validateRequestedChange(value: unknown): GitHubRequestedChangeEvidence {
	const candidate = exactRecord(value, ["id", "actor", "commitSha", "dismissed", "submittedAt"]);
	if (typeof candidate.dismissed !== "boolean") throw new Error("invalid requested-change dismissal state");
	return {
		id: inlineText(candidate.id, "requested-change ID"),
		actor: inlineText(candidate.actor, "requested-change actor"),
		commitSha: sha(candidate.commitSha, "requested-change commit SHA"),
		dismissed: candidate.dismissed,
		submittedAt: timestamp(candidate.submittedAt, "requested-change timestamp"),
	};
}

function validateThread(value: unknown): GitHubReviewThreadEvidence {
	const candidate = exactRecord(value, ["id", "resolved", "headSha"]);
	if (typeof candidate.resolved !== "boolean") throw new Error("invalid review-thread resolution state");
	return {
		id: inlineText(candidate.id, "review-thread ID"),
		resolved: candidate.resolved,
		headSha: sha(candidate.headSha, "review-thread head SHA"),
	};
}

function validateDisposition(value: unknown): GitHubFindingDisposition {
	const candidate = exactRecord(value, ["findingId", "kind", "rationale", "actor", "headSha", "recordedAt"]);
	if (candidate.kind !== "fixed" && candidate.kind !== "not_actionable") throw new Error("invalid finding disposition");
	return {
		findingId: inlineText(candidate.findingId, "disposition finding ID"),
		kind: candidate.kind,
		rationale: inlineText(candidate.rationale, "disposition rationale", 2_048),
		actor: inlineText(candidate.actor, "disposition actor"),
		headSha: sha(candidate.headSha, "disposition head SHA"),
		recordedAt: timestamp(candidate.recordedAt, "disposition timestamp"),
	};
}

export function validateGitHubPullRequestEvidence(value: unknown): GitHubPullRequestEvidence {
	const candidate = exactRecord(value, [
		"schemaVersion",
		"number",
		"marker",
		"title",
		"body",
		"state",
		"draft",
		"baseBranch",
		"headBranch",
		"baseSha",
		"headSha",
		"changedPaths",
		"mergeState",
		"checksComplete",
		"checks",
		"requestedChanges",
		"threads",
		"reviews",
		"reviewsComplete",
		"dispositions",
		"observedAt",
	]);
	if (candidate.schemaVersion !== 1) throw new Error("unsupported GitHub evidence schema");
	const marker = inlineText(candidate.marker, "pull request marker", 512);
	if (!MARKER.test(marker)) throw new Error("invalid pull request marker");
	const body = bodyText(candidate.body);
	if (body.split(marker).length !== 2) throw new Error("pull request marker must occur exactly once in its body");
	if (candidate.state !== "open" && candidate.state !== "closed" && candidate.state !== "merged") {
		throw new Error("invalid pull request state");
	}
	if (typeof candidate.draft !== "boolean") throw new Error("invalid pull request draft state");
	if (typeof candidate.checksComplete !== "boolean" || typeof candidate.reviewsComplete !== "boolean") {
		throw new Error("invalid evidence completeness attestation");
	}
	if (!["clean", "blocked", "behind", "conflicting", "unknown"].includes(candidate.mergeState as string)) {
		throw new Error("invalid pull request merge state");
	}
	const checks = boundedArray(candidate.checks, "checks").map(validateCheck);
	const requestedChanges = boundedArray(candidate.requestedChanges, "requested changes").map(validateRequestedChange);
	const threads = boundedArray(candidate.threads, "threads").map(validateThread);
	const reviews = boundedArray(candidate.reviews, "reviews", MAX_REVIEWS).map(validateIndependentReviewRecord);
	const dispositions = boundedArray(candidate.dispositions, "dispositions").map(validateDisposition);
	unique(checks, (check) => check.id, "check ID");
	unique(requestedChanges, (change) => change.id, "requested-change ID");
	unique(threads, (thread) => thread.id, "review-thread ID");
	unique(dispositions, (disposition) => disposition.findingId, "finding disposition");
	const findings = reviews.flatMap((review) => review.findings);
	unique(findings, (finding) => finding.id, "review finding ID");
	const findingIds = new Set(findings.map((finding) => finding.id));
	for (const disposition of dispositions) {
		if (!findingIds.has(disposition.findingId)) throw new Error("finding disposition does not reference authoritative review evidence");
	}
	return {
		schemaVersion: 1,
		number: githubNumber(candidate.number, "pull request number"),
		marker,
		title: inlineText(candidate.title, "pull request title", 256),
		body,
		state: candidate.state,
		draft: candidate.draft,
		baseBranch: canonicalGitRef(candidate.baseBranch, "base branch"),
		headBranch: canonicalGitRef(candidate.headBranch, "head branch"),
		baseSha: sha(candidate.baseSha, "pull request base SHA"),
		headSha: sha(candidate.headSha, "pull request head SHA"),
		changedPaths: stringArray(candidate.changedPaths, "pull request changed paths", true),
		mergeState: candidate.mergeState as GitHubPullRequestEvidence["mergeState"],
		checksComplete: candidate.checksComplete,
		checks,
		requestedChanges,
		threads,
		reviews,
		reviewsComplete: candidate.reviewsComplete,
		dispositions,
		observedAt: timestamp(candidate.observedAt, "pull request observation timestamp"),
	};
}

function expectedEvidence(value: ExpectedPullRequestEvidence): ExpectedPullRequestEvidence {
	const candidate = exactRecord(value, [
		"number", "marker", "baseBranch", "headBranch", "baseSha", "headSha", "changedPaths",
		"requiredChecks", "reviewTarget", "attestations",
	]);
	const marker = inlineText(candidate.marker, "expected pull request marker", 512);
	if (!MARKER.test(marker)) throw new Error("invalid expected pull request marker");
	const requiredChecks = boundedArray(candidate.requiredChecks, "required checks").map(validateRequiredCheck);
	if (requiredChecks.length === 0) throw new Error("required checks must not be empty");
	unique(requiredChecks, (check) => `${check.name}\u0000${check.producerId}`, "required check context");
	const work = createIndependentReviewWork(candidate.reviewTarget as IndependentReviewTarget);
	const changedPaths = stringArray(candidate.changedPaths, "expected changed paths", true);
	if (changedPaths.length !== work.changedPaths.length
		|| changedPaths.some((path, index) => path !== work.changedPaths[index])) {
		throw new Error("expected review target must bind the authoritative changed-path set");
	}
	return {
		number: githubNumber(candidate.number, "expected pull request number"),
		marker,
		baseBranch: canonicalGitRef(candidate.baseBranch, "expected base branch"),
		headBranch: canonicalGitRef(candidate.headBranch, "expected head branch"),
		baseSha: sha(candidate.baseSha, "expected base SHA"),
		headSha: sha(candidate.headSha, "expected head SHA"),
		changedPaths,
		requiredChecks,
		reviewTarget: {
			repository: work.repository,
			workItemId: work.workItemId,
			pullRequest: work.pullRequest,
			generation: work.generation,
			baseSha: work.baseSha,
			headSha: work.headSha,
			changedPaths: work.changedPaths,
			allowedScopes: work.allowedScopes,
		},
		attestations: boundedArray(candidate.attestations, "AgentSession attestations") as unknown as AgentSessionAttestation[],
	};
}

export function evaluateGitHubPullRequestEvidence(
	value: GitHubPullRequestEvidence,
	expectedValue: ExpectedPullRequestEvidence,
	options: PullRequestEvaluationOptions = {},
): GitHubPullRequestDecision {
	const evidence = validateGitHubPullRequestEvidence(value);
	const expected = expectedEvidence(expectedValue);
	const blockers = new Set<GitHubEvidenceBlocker>();
	if (evidence.number !== expected.number) blockers.add("resource_mismatch");
	if (evidence.marker !== expected.marker) blockers.add("marker_collision");
	if (evidence.state !== "open") blockers.add("pr_not_open");
	if (evidence.draft && options.allowDraft !== true) blockers.add("draft");
	if (evidence.baseBranch !== expected.baseBranch || evidence.headBranch !== expected.headBranch
		|| evidence.baseSha !== expected.baseSha) blockers.add("topology_mismatch");
	if (evidence.headSha !== expected.headSha) blockers.add("head_moved");
	if (evidence.changedPaths.length !== expected.changedPaths.length
		|| evidence.changedPaths.some((path, index) => path !== expected.changedPaths[index])) blockers.add("resource_mismatch");
	if (evidence.mergeState !== "clean") blockers.add("merge_blocked");
	let ciGreen = evidence.checksComplete;
	for (const required of expected.requiredChecks) {
		const matches = evidence.checks
			.filter((check) => check.name === required.name && check.producerId === required.producerId
				&& check.headSha === expected.headSha)
			.sort((left, right) => right.completedAt.localeCompare(left.completedAt) || left.id.localeCompare(right.id));
		const latest = matches[0];
		if (latest === undefined || latest.status !== "completed" || latest.conclusion !== "success"
			|| (matches[1] !== undefined && matches[1].completedAt === latest.completedAt)) ciGreen = false;
	}
	if (!ciGreen) {
		blockers.add("ci_not_green");
	}
	if (evidence.requestedChanges.some((change) => !change.dismissed)) blockers.add("changes_requested");
	if (evidence.threads.some((thread) => !thread.resolved)) blockers.add("unresolved_thread");
	const dispositions = new Map(evidence.dispositions.map((disposition) => [disposition.findingId, disposition]));
	const blockingFindings = evidence.reviews.flatMap((review) => review.findings)
		.filter((finding) => finding.severity === "blocking");
	if (blockingFindings.some((finding) => {
		const disposition = dispositions.get(finding.id);
		return disposition?.kind !== "fixed" || disposition.headSha !== expected.headSha;
	})) {
		blockers.add("undispositioned_finding");
	}
	const reviewDecision = reconcileIndependentReview({
		target: expected.reviewTarget,
		reviews: evidence.reviews,
		attestations: expected.attestations,
	});
	if (!evidence.reviewsComplete || reviewDecision.kind !== "satisfied") blockers.add("review_missing");
	if (blockers.size > 0 || reviewDecision.kind !== "satisfied") {
		return { kind: "blocked", blockers: [...blockers] };
	}
	return { kind: "eligible", review: reviewDecision.review };
}
