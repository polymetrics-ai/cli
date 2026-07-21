import { execFile } from "node:child_process";
import { isAbsolute } from "node:path";

const MAX_GITHUB_NUMBER = 2_147_483_647;

export interface TargetEvidenceRequest {
	cwd: string;
	issue: number;
	pr?: number;
}

export interface TargetEvidence {
	cwd: string;
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
	});
}

function defaultExec(file: string, args: string[], cwd: string): Promise<string> {
	return new Promise((resolve, reject) => {
		execFile(file, args, { cwd, encoding: "utf8", maxBuffer: 1024 * 1024 }, (error, stdout) => {
			if (error) {
				reject(new TargetEvidenceError(`${file} command failed while capturing target evidence`, { cause: error }));
				return;
			}
			resolve(stdout);
		});
	});
}

function parsePullRequest(raw: string, expectedPR: number): GitHubPullRequest {
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
	return pr as unknown as GitHubPullRequest;
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

	const [rawHead, rawBranch, rawStatus] = await Promise.all([
		execute("git", ["rev-parse", "HEAD"], request.cwd),
		execute("git", ["branch", "--show-current"], request.cwd),
		execute("git", ["status", "--porcelain"], request.cwd),
	]);
	const rawPR = request.pr === undefined
		? undefined
		: await execute(
			"gh",
			[
				"pr",
				"view",
				String(request.pr),
				"--json",
				"number,state,isDraft,baseRefName,headRefName,headRefOid,url,mergeStateStatus,reviewDecision,statusCheckRollup",
			],
			request.cwd,
		);

	const candidateHead = rawHead.trim().toLowerCase();
	const branch = rawBranch.trim();
	if (!/^[0-9a-f]{40}$/.test(candidateHead)) throw new TargetEvidenceError("local target head is invalid");
	if (!safeText(branch)) throw new TargetEvidenceError("local target branch is invalid or detached");
	if (rawStatus.trim() !== "") throw new TargetEvidenceError("target working tree must be clean");

	if (rawPR === undefined || request.pr === undefined) {
		return { cwd: request.cwd, branch, candidateHead, clean: true };
	}

	const pr = parsePullRequest(rawPR, request.pr);
	if (pr.state.toUpperCase() !== "OPEN") throw new TargetEvidenceError(`pull request #${request.pr} must be open`);
	if (branch !== pr.headRefName) throw new TargetEvidenceError("local branch does not match pull request head branch");
	if (candidateHead !== pr.headRefOid.toLowerCase()) throw new TargetEvidenceError("local head does not match pull request head");

	return {
		cwd: request.cwd,
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
