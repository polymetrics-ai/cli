const MAX_COMMAND_LENGTH = 4_096;
const MAX_GITHUB_NUMBER = 2_147_483_647;
const DEFAULT_MAX_CONCURRENCY = 2;
const DEFAULT_TIMEOUT_SECONDS = 900;
const MIN_TIMEOUT_SECONDS = 30;
const MAX_TIMEOUT_SECONDS = 3_600;

export class ShepherdArgumentError extends Error {
	constructor(message: string) {
		super(message);
		this.name = "ShepherdArgumentError";
	}
}

export interface ShepherdHelpCommand {
	action: "help";
}

export interface ShepherdStatusCommand {
	action: "status";
	issue: number;
}

export interface ShepherdStopCommand {
	action: "stop";
	issue: number;
}

interface ShepherdAutonomousRunCommandBase {
	issue: number;
	pr?: number;
	backend: "sdk-inproc";
	maxConcurrency: number;
	timeoutMs: number;
}

interface ShepherdCanaryRunCommandBase {
	issue: number;
	pr: number;
	readOnly: true;
	backend: "sdk-inproc";
	experimental: true;
	maxConcurrency: number;
	timeoutMs: number;
}

export interface ShepherdStartCommand extends ShepherdAutonomousRunCommandBase {
	action: "start";
}

export interface ShepherdResumeCommand extends ShepherdAutonomousRunCommandBase {
	action: "resume";
}

export interface ShepherdCanaryCommand extends ShepherdCanaryRunCommandBase {
	action: "canary";
}

export type ShepherdCommand =
	| ShepherdHelpCommand
	| ShepherdStatusCommand
	| ShepherdStopCommand
	| ShepherdStartCommand
	| ShepherdResumeCommand
	| ShepherdCanaryCommand;

type ParsedFlags = Map<string, string | true>;

const valueFlags = new Set([
	"--issue",
	"--pr",
	"--backend",
	"--max-concurrency",
	"--timeout-seconds",
]);
const booleanFlags = new Set(["--read-only", "--experimental"]);

function fail(message: string): never {
	throw new ShepherdArgumentError(message);
}

function parseFlags(tokens: string[]): ParsedFlags {
	const flags: ParsedFlags = new Map();
	for (let index = 0; index < tokens.length; index += 1) {
		const flag = tokens[index];
		if (!flag.startsWith("--")) fail(`unexpected positional argument ${JSON.stringify(flag)}`);
		if (!valueFlags.has(flag) && !booleanFlags.has(flag)) fail(`unknown flag ${flag}`);
		if (flags.has(flag)) fail(`duplicate flag ${flag}`);

		if (booleanFlags.has(flag)) {
			flags.set(flag, true);
			continue;
		}

		const value = tokens[index + 1];
		if (value === undefined || value.startsWith("--")) fail(`flag ${flag} requires a value`);
		flags.set(flag, value);
		index += 1;
	}
	return flags;
}

function rejectUnexpected(flags: ParsedFlags, allowed: ReadonlySet<string>): void {
	for (const flag of flags.keys()) {
		if (!allowed.has(flag)) fail(`flag ${flag} is not valid for this action`);
	}
}

function positiveInteger(raw: string | true | undefined, name: string, maximum: number): number {
	if (typeof raw !== "string" || !/^[1-9][0-9]*$/.test(raw)) {
		fail(`${name} must be a positive integer`);
	}
	const value = Number(raw);
	if (!Number.isSafeInteger(value) || value > maximum) fail(`${name} is out of range`);
	return value;
}

function requiredIssue(flags: ParsedFlags): number {
	return positiveInteger(flags.get("--issue"), "--issue", MAX_GITHUB_NUMBER);
}

function optionalPR(flags: ParsedFlags): number | undefined {
	const raw = flags.get("--pr");
	return raw === undefined ? undefined : positiveInteger(raw, "--pr", MAX_GITHUB_NUMBER);
}

function concurrencyAndTimeout(flags: ParsedFlags): { maxConcurrency: number; timeoutMs: number } {
	const concurrencyRaw = flags.get("--max-concurrency");
	const maxConcurrency = concurrencyRaw === undefined
		? DEFAULT_MAX_CONCURRENCY
		: positiveInteger(concurrencyRaw, "--max-concurrency", DEFAULT_MAX_CONCURRENCY);
	const timeoutRaw = flags.get("--timeout-seconds");
	const timeoutSeconds = timeoutRaw === undefined
		? DEFAULT_TIMEOUT_SECONDS
		: positiveInteger(timeoutRaw, "--timeout-seconds", MAX_TIMEOUT_SECONDS);
	if (timeoutSeconds < MIN_TIMEOUT_SECONDS) {
		fail(`--timeout-seconds must be between ${MIN_TIMEOUT_SECONDS} and ${MAX_TIMEOUT_SECONDS}`);
	}
	return { maxConcurrency, timeoutMs: timeoutSeconds * 1_000 };
}

function parseAutonomousCommand(
	action: "start" | "resume",
	flags: ParsedFlags,
): ShepherdStartCommand | ShepherdResumeCommand {
	rejectUnexpected(flags, new Set(["--issue", "--pr", "--backend", "--max-concurrency", "--timeout-seconds"]));
	const backend = flags.get("--backend");
	if (backend !== undefined && backend !== "sdk-inproc") fail(`${action} supports only --backend sdk-inproc`);
	const issue = requiredIssue(flags);
	const pr = optionalPR(flags);
	return {
		action,
		issue,
		...(pr === undefined ? {} : { pr }),
		backend: "sdk-inproc",
		...concurrencyAndTimeout(flags),
	};
}

function parseCanaryCommand(flags: ParsedFlags): ShepherdCanaryCommand {
	rejectUnexpected(
		flags,
		new Set([
			"--issue",
			"--pr",
			"--read-only",
			"--backend",
			"--experimental",
			"--max-concurrency",
			"--timeout-seconds",
		]),
	);

	const issue = requiredIssue(flags);
	const pr = optionalPR(flags);
	if (pr === undefined) fail("canary requires --pr");
	if (flags.get("--read-only") !== true) fail("canary requires --read-only");
	if (flags.get("--experimental") !== true) fail("canary requires --experimental");
	if (flags.get("--backend") !== "sdk-inproc") fail("canary requires --backend sdk-inproc");
	return {
		action: "canary",
		issue,
		pr,
		readOnly: true as const,
		backend: "sdk-inproc" as const,
		experimental: true as const,
		...concurrencyAndTimeout(flags),
	};
}

export function parseShepherdCommand(input: string): ShepherdCommand {
	if (typeof input !== "string") fail("command arguments must be text");
	if (input.length > MAX_COMMAND_LENGTH) fail("command arguments are too long");
	if (/[\u0000-\u001f\u007f]/.test(input)) fail("command arguments contain control characters");

	const tokens = input.trim().split(/\s+/).filter(Boolean);
	if (tokens.length === 0) return { action: "help" };
	const [rawAction, ...remaining] = tokens;
	const action = rawAction === "--help" ? "help" : rawAction;
	const flags = parseFlags(remaining);

	switch (action) {
		case "help":
			rejectUnexpected(flags, new Set());
			return { action };
		case "status":
			rejectUnexpected(flags, new Set(["--issue"]));
			return { action, issue: requiredIssue(flags) };
		case "stop":
			rejectUnexpected(flags, new Set(["--issue"]));
			return { action, issue: requiredIssue(flags) };
		case "start":
		case "resume":
			return parseAutonomousCommand(action, flags);
		case "canary":
			return parseCanaryCommand(flags);
		default:
			return fail(`unknown action ${JSON.stringify(action)}`);
	}
}
