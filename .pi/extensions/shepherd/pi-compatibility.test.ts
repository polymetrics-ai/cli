import assert from "node:assert/strict";
import test from "node:test";

import {
	assertShepherdPiCompatibility,
	isShepherdPiVersionSupported,
	SHEPHERD_PI_MAX_EXCLUSIVE,
	SHEPHERD_PI_MIN_VERSION,
} from "./pi-compatibility.ts";

test("Pi compatibility is the explicit stable 0.80.10 interval", () => {
	assert.equal(SHEPHERD_PI_MIN_VERSION, "0.80.10");
	assert.equal(SHEPHERD_PI_MAX_EXCLUSIVE, "0.80.11");
	assert.equal(isShepherdPiVersionSupported("0.80.10"), true);
	for (const version of [
		"0.80.9",
		"0.80.11",
		"0.81.0",
		"1.0.0",
		"0.80.10-beta.1",
		"v0.80.10",
		"00.80.10",
		"0.080.10",
		"0.80.010",
		"invalid",
		"",
		null,
		undefined,
	]) {
		assert.equal(isShepherdPiVersionSupported(version), false, String(version));
	}
});

test("Pi runtime and required versions must both match the same supported release", () => {
	assert.doesNotThrow(() => assertShepherdPiCompatibility("0.80.10"));
	assert.doesNotThrow(() => assertShepherdPiCompatibility("0.80.10", "0.80.10"));
	assert.throws(() => assertShepherdPiCompatibility("0.80.10", "0.80.9"), /bounded Pi compatibility/);
	assert.throws(() => assertShepherdPiCompatibility("0.80.9", "0.80.10"), /bounded Pi compatibility/);
});
