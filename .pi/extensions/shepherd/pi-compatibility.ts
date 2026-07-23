const STABLE_VERSION = /^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$/u;
const INTRINSIC_NUMBER = Number;
const INTRINSIC_NUMBER_IS_SAFE_INTEGER = Number.isSafeInteger;

export const SHEPHERD_PI_MIN_VERSION = "0.80.10";
export const SHEPHERD_PI_MAX_EXCLUSIVE = "0.80.11";
export const REQUIRED_PI_VERSION = SHEPHERD_PI_MIN_VERSION;

function parseStableVersion(value: unknown): readonly [number, number, number] | undefined {
	if (typeof value !== "string") return undefined;
	const match = STABLE_VERSION.exec(value);
	if (!match) return undefined;
	const parts = [INTRINSIC_NUMBER(match[1]), INTRINSIC_NUMBER(match[2]), INTRINSIC_NUMBER(match[3])] as const;
	return parts.every(INTRINSIC_NUMBER_IS_SAFE_INTEGER) ? parts : undefined;
}

function compare(
	left: readonly [number, number, number],
	right: readonly [number, number, number],
): number {
	for (let index = 0; index < left.length; index += 1) {
		if (left[index] !== right[index]) return left[index]! < right[index]! ? -1 : 1;
	}
	return 0;
}

export function isShepherdPiVersionSupported(value: unknown): value is string {
	const version = parseStableVersion(value);
	const minimum = parseStableVersion(SHEPHERD_PI_MIN_VERSION)!;
	const maximum = parseStableVersion(SHEPHERD_PI_MAX_EXCLUSIVE)!;
	return version !== undefined && compare(version, minimum) >= 0 && compare(version, maximum) < 0;
}

export function assertShepherdPiCompatibility(version: unknown, requiredVersion?: unknown): asserts version is string {
	if (!isShepherdPiVersionSupported(version)
		|| (requiredVersion !== undefined && (!isShepherdPiVersionSupported(requiredVersion) || requiredVersion !== version))) {
		throw new Error(
			`AgentSession Shepherd requires bounded Pi compatibility >=${SHEPHERD_PI_MIN_VERSION} <${SHEPHERD_PI_MAX_EXCLUSIVE}; found ${String(version)}`,
		);
	}
}
