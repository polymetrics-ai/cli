import assert from "node:assert/strict";
import test from "node:test";

import { parseShepherdCommand } from "./arguments.ts";

test("bare command renders help without enabling a run", () => {
	assert.deepEqual(parseShepherdCommand(""), { action: "help" });
});

test("parses an explicit read-only in-process canary", () => {
	assert.deepEqual(
		parseShepherdCommand(
			"canary --issue 397 --pr 438 --read-only --backend sdk-inproc --experimental --max-concurrency 2 --timeout-seconds 900",
		),
		{
			action: "canary",
			issue: 397,
			pr: 438,
			readOnly: true,
			backend: "sdk-inproc",
			experimental: true,
			maxConcurrency: 2,
			timeoutMs: 900_000,
		},
	);
});

test("status is read-only and does not require experimental acknowledgement", () => {
	assert.deepEqual(parseShepherdCommand("status --issue 471"), {
		action: "status",
		issue: 471,
	});
});

test("resume accepts either an omitted or explicit PR for persisted-target binding", () => {
	const base = "resume --issue 397 --read-only --backend sdk-inproc --experimental";
	assert.equal(parseShepherdCommand(base).pr, undefined);
	assert.equal(parseShepherdCommand(`${base} --pr 438`).pr, 438);
});

test("rejects unsafe or ambiguous command shapes", () => {
	const invalid = [
		"canary --issue 397 --pr 438 --read-only --backend sdk-inproc",
		"canary --issue 397 --pr 438 --backend sdk-inproc --experimental",
		"canary --issue 397 --read-only --backend sdk-inproc --experimental",
		"start --issue 397 --backend subprocess --experimental --read-only",
		"start --issue 0 --backend sdk-inproc --experimental --read-only",
		"start --issue 397 --issue 398 --backend sdk-inproc --experimental --read-only",
		"start --issue 397 --backend sdk-inproc --experimental --read-only --max-concurrency 3",
		"start --issue 397 --backend sdk-inproc --experimental --read-only --timeout-seconds 5",
		"stop --issue ../397",
		"status --issue 397\u0000",
		"dance --issue 397",
		"status --unknown true",
	];

	for (const input of invalid) {
		assert.throws(() => parseShepherdCommand(input), { name: "ShepherdArgumentError" }, input);
	}
});
