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

test("parses autonomous start and resume without read-only canary acknowledgements", () => {
	assert.deepEqual(parseShepherdCommand("start --issue 479"), {
		action: "start",
		issue: 479,
		backend: "sdk-inproc",
		maxConcurrency: 2,
		timeoutMs: 900_000,
	});
	assert.equal(parseShepherdCommand("resume --issue 479 --pr 472").pr, 472);
	assert.throws(
		() => parseShepherdCommand("start --issue 479 --read-only"),
		/canary|not valid/i,
	);
});

test("rejects unsafe or ambiguous command shapes", () => {
	const invalid = [
		"canary --issue 397 --pr 438 --read-only --backend sdk-inproc",
		"canary --issue 397 --pr 438 --backend sdk-inproc --experimental",
		"canary --issue 397 --read-only --backend sdk-inproc --experimental",
		"start --issue 397 --backend subprocess",
		"start --issue 0",
		"start --issue 397 --issue 398",
		"start --issue 397 --max-concurrency 3",
		"start --issue 397 --timeout-seconds 5",
		"stop --issue ../397",
		"status --issue 397\u0000",
		"dance --issue 397",
		"status --unknown true",
	];

	for (const input of invalid) {
		assert.throws(() => parseShepherdCommand(input), { name: "ShepherdArgumentError" }, input);
	}
});
