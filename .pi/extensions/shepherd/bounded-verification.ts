import { spawn, type ChildProcessWithoutNullStreams } from "node:child_process";
import { lstatSync, realpathSync, statSync } from "node:fs";
import { isAbsolute, relative, resolve, sep } from "node:path";

import type { ProductionVerificationCommand } from "./autonomous-production-contract.ts";

const MAX_EXECUTABLES = 32;
const MAX_ENVIRONMENT_FIELDS = 64;
const MAX_ARGUMENTS = 256;
const MAX_ARGUMENT_BYTES = 16 * 1024;
const MAX_TIMEOUT_MS = 120_000;
const MAX_OUTPUT_BYTES = 4 * 1024 * 1024;
const SAFE_NAME = /^[A-Za-z0-9][A-Za-z0-9._+-]{0,127}$/;
const SAFE_ENVIRONMENT_NAME = /^[A-Za-z_][A-Za-z0-9_]{0,127}$/;
const UNSAFE_TEXT = /[\u0000-\u0008\u000b\u000c\u000e-\u001f\u007f-\u009f]/u;

export interface ProductionVerificationResult {
	id: string;
	status: "passed" | "failed";
	exitCode: number | null;
	signal: NodeJS.Signals | null;
	stdout: string;
	stderr: string;
	durationMs: number;
	failureKind?: "exit" | "timeout" | "aborted" | "output_limit" | "spawn";
}

export interface BoundedVerificationRunnerOptions {
	executables?: Readonly<Record<string, string>>;
	environment?: Readonly<Record<string, string>>;
	terminationGraceMs?: number;
}

export class BoundedVerificationRunner {
	readonly #executables: ReadonlyMap<string, string>;
	readonly #environment: Readonly<Record<string, string>>;
	readonly #terminationGraceMs: number;

	constructor(options: BoundedVerificationRunnerOptions = {}) {
		if (typeof options !== "object" || options === null) throw new Error("verification runner options are invalid");
		const configured = options.executables ?? { node: process.execPath };
		const entries = Object.entries(configured);
		if (entries.length < 1 || entries.length > MAX_EXECUTABLES) {
			throw new Error("verification executable allowlist must contain one to 32 entries");
		}
		const executables = new Map<string, string>();
		for (const [name, path] of entries) {
			if (!SAFE_NAME.test(name) || typeof path !== "string" || !isAbsolute(path) || UNSAFE_TEXT.test(path)) {
				throw new Error("verification executable allowlist contains an invalid absolute executable");
			}
			let metadata: ReturnType<typeof lstatSync>;
			try {
				metadata = lstatSync(path);
			} catch {
				throw new Error(`verification executable ${name} is unavailable`);
			}
			if (!metadata.isFile() || metadata.isSymbolicLink() || realpathSync(path) !== resolve(path)
				|| (process.platform !== "win32" && (metadata.mode & 0o111) === 0)) {
				throw new Error(`verification executable ${name} must be a canonical executable regular file, not a symlink`);
			}
			executables.set(name, path);
		}
		this.#executables = executables;

		const environment = options.environment ?? {};
		const environmentEntries = Object.entries(environment);
		if (environmentEntries.length > MAX_ENVIRONMENT_FIELDS) throw new Error("verification environment is too large");
		const sanitized = Object.create(null) as Record<string, string>;
		for (const [name, value] of environmentEntries) {
			if (!SAFE_ENVIRONMENT_NAME.test(name) || typeof value !== "string"
				|| Buffer.byteLength(value) > 16 * 1024 || UNSAFE_TEXT.test(value)) {
				throw new Error("verification environment contains unsafe data");
			}
			sanitized[name] = value;
		}
		sanitized.LC_ALL = "C";
		sanitized.LANG = "C";
		sanitized.TZ = "UTC";
		this.#environment = Object.freeze(sanitized);

		const terminationGraceMs = options.terminationGraceMs ?? 250;
		if (!Number.isSafeInteger(terminationGraceMs) || terminationGraceMs < 1 || terminationGraceMs > 5_000) {
			throw new Error("verification termination grace must be between 1 and 5000 milliseconds");
		}
		this.#terminationGraceMs = terminationGraceMs;
	}

	async run(
		worktreeRoot: string,
		command: ProductionVerificationCommand,
		signal?: AbortSignal,
	): Promise<ProductionVerificationResult> {
		const executable = this.#executables.get(command.executable);
		if (executable === undefined) throw new Error(`verification executable ${command.executable} is not allowlisted`);
		validateCommand(command);
		const cwd = canonicalCommandCwd(worktreeRoot, command.cwd);
		if (signal !== undefined && !(signal instanceof AbortSignal)) throw new Error("verification AbortSignal is invalid");
		if (signal?.aborted) return failedBeforeSpawn(command.id, "aborted");
		return new Promise<ProductionVerificationResult>((resolveResult) => {
			const startedAt = Date.now();
			let child: ChildProcessWithoutNullStreams;
			try {
				child = spawn(executable, [...command.args], {
					cwd,
					env: this.#environment,
					shell: false,
					stdio: ["ignore", "pipe", "pipe"],
					windowsHide: true,
				});
			} catch {
				resolveResult({ ...failedBeforeSpawn(command.id, "spawn"), durationMs: Date.now() - startedAt });
				return;
			}

			const stdout: Buffer[] = [];
			const stderr: Buffer[] = [];
			let capturedBytes = 0;
			let failureKind: ProductionVerificationResult["failureKind"];
			let finished = false;
			let forceTimer: ReturnType<typeof setTimeout> | undefined;
			const terminate = (kind: NonNullable<ProductionVerificationResult["failureKind"]>): void => {
				failureKind ??= kind;
				if (child.exitCode !== null || child.signalCode !== null) return;
				child.kill("SIGTERM");
				forceTimer ??= setTimeout(() => {
					if (child.exitCode === null && child.signalCode === null) child.kill("SIGKILL");
				}, this.#terminationGraceMs);
				forceTimer.unref?.();
			};
			const capture = (target: Buffer[], value: Buffer): void => {
				const remaining = command.maxOutputBytes - capturedBytes;
				if (remaining > 0) {
					const bounded = value.length <= remaining ? value : value.subarray(0, remaining);
					target.push(Buffer.from(bounded));
					capturedBytes += bounded.length;
				}
				if (value.length > remaining) terminate("output_limit");
			};
			child.stdout.on("data", (value: Buffer) => capture(stdout, value));
			child.stderr.on("data", (value: Buffer) => capture(stderr, value));
			const timeout = setTimeout(() => terminate("timeout"), command.timeoutMs);
			timeout.unref?.();
			const onAbort = (): void => terminate("aborted");
			signal?.addEventListener("abort", onAbort, { once: true });

			const finish = (exitCode: number | null, processSignal: NodeJS.Signals | null): void => {
				if (finished) return;
				finished = true;
				clearTimeout(timeout);
				if (forceTimer) clearTimeout(forceTimer);
				signal?.removeEventListener("abort", onAbort);
				const kind = failureKind ?? (exitCode === 0 ? undefined : "exit");
				resolveResult({
					id: command.id,
					status: kind === undefined ? "passed" : "failed",
					exitCode,
					signal: processSignal,
					stdout: Buffer.concat(stdout).toString("utf8"),
					stderr: Buffer.concat(stderr).toString("utf8"),
					durationMs: Date.now() - startedAt,
					...(kind === undefined ? {} : { failureKind: kind }),
				});
			};
			child.once("error", () => {
				failureKind ??= "spawn";
				finish(child.exitCode, child.signalCode);
			});
			child.once("close", finish);
		});
	}

	async runAll(
		worktreeRoot: string,
		commands: readonly ProductionVerificationCommand[],
		signal?: AbortSignal,
	): Promise<ProductionVerificationResult[]> {
		if (!Array.isArray(commands) || commands.length < 1 || commands.length > 64) {
			throw new Error("verification command list must contain one to 64 commands");
		}
		const results: ProductionVerificationResult[] = [];
		for (const command of commands) {
			const result = await this.run(worktreeRoot, command, signal);
			results.push(result);
			if (result.status !== "passed") break;
		}
		return results;
	}
}

function validateCommand(command: ProductionVerificationCommand): void {
	if (typeof command !== "object" || command === null || !SAFE_NAME.test(command.id)
		|| !SAFE_NAME.test(command.executable)) throw new Error("verification command identity is invalid");
	if (!Array.isArray(command.args) || command.args.length > MAX_ARGUMENTS
		|| command.args.some((argument) => typeof argument !== "string"
			|| Buffer.byteLength(argument) > MAX_ARGUMENT_BYTES || UNSAFE_TEXT.test(argument))) {
		throw new Error("verification argv is invalid or exceeds its bound");
	}
	if (!Number.isSafeInteger(command.timeoutMs) || command.timeoutMs < 1 || command.timeoutMs > MAX_TIMEOUT_MS) {
		throw new Error("verification timeout is invalid");
	}
	if (!Number.isSafeInteger(command.maxOutputBytes) || command.maxOutputBytes < 1_024
		|| command.maxOutputBytes > MAX_OUTPUT_BYTES) throw new Error("verification output limit is invalid");
}

function canonicalCommandCwd(worktreeRoot: string, cwd: string): string {
	if (typeof worktreeRoot !== "string" || !isAbsolute(worktreeRoot) || Buffer.byteLength(worktreeRoot) > 4_096
		|| typeof cwd !== "string" || cwd.length < 1 || isAbsolute(cwd) || cwd.includes("\\")
		|| cwd.endsWith("/") || cwd.split("/").some((part) => part === "" || part === "..")) {
		throw new Error("verification cwd must be repository-relative");
	}
	const suppliedRoot = resolve(worktreeRoot);
	let rootMetadata: ReturnType<typeof lstatSync>;
	try {
		rootMetadata = lstatSync(suppliedRoot);
	} catch {
		throw new Error("verification worktree root is unavailable");
	}
	if (!rootMetadata.isDirectory() || rootMetadata.isSymbolicLink()) {
		throw new Error("verification worktree root must be a canonical directory without symlinks");
	}
	const root = realpathSync(suppliedRoot);
	const target = resolve(root, cwd);
	const back = relative(root, target);
	if (back === ".." || back.startsWith(`..${sep}`)) throw new Error("verification cwd escapes the worktree");
	let current = root;
	for (const part of back === "" ? [] : back.split(sep)) {
		current = resolve(current, part);
		let metadata: ReturnType<typeof lstatSync>;
		try {
			metadata = lstatSync(current);
		} catch {
			throw new Error("verification cwd is unavailable");
		}
		if (!metadata.isDirectory() || metadata.isSymbolicLink()) {
			throw new Error("verification cwd must contain only real directories, not symlinks");
		}
	}
	if (realpathSync(target) !== target || !statSync(target).isDirectory()) {
		throw new Error("verification cwd is not canonical");
	}
	return target;
}

function failedBeforeSpawn(
	id: string,
	failureKind: NonNullable<ProductionVerificationResult["failureKind"]>,
): ProductionVerificationResult {
	return {
		id,
		status: "failed",
		exitCode: null,
		signal: null,
		stdout: "",
		stderr: "",
		durationMs: 0,
		failureKind,
	};
}
