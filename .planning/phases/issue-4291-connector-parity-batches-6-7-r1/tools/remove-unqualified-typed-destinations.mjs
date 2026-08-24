#!/usr/bin/env node

import { readFile, writeFile } from "node:fs/promises";
import { resolve } from "node:path";

const connectors = [
  "close-com",
  "outreach",
  "zoho-bigin",
  "braze",
  "customer-io",
  "help-scout",
  "gorgias",
  "service-now",
  "chatwoot",
  "chargebee",
];

const root = resolve(import.meta.dirname, "../../../..");

for (const connector of connectors) {
  const file = resolve(root, "internal/connectors/defs", connector, "sync_transport.json");
  const raw = await readFile(file, "utf8");
  const transport = JSON.parse(raw);
  const destination = transport.destination_transport;

  if (destination?.executor?.id !== "declarative_typed_destination") {
    throw new Error(`${connector}: expected declarative_typed_destination`);
  }
  if (!destination.apply_strategies?.some(({ mode }) => mode === "full_overwrite")) {
    throw new Error(`${connector}: expected rejected full_overwrite strategy`);
  }

  const destinationKey = raw.indexOf('"destination_transport"');
  const propertyStart = raw.lastIndexOf(",", destinationKey);
  const valueStart = raw.indexOf("{", destinationKey);
  let depth = 0;
  let quoted = false;
  let escaped = false;
  let valueEnd = -1;

  for (let index = valueStart; index < raw.length; index += 1) {
    const character = raw[index];
    if (quoted) {
      if (escaped) {
        escaped = false;
      } else if (character === "\\") {
        escaped = true;
      } else if (character === '"') {
        quoted = false;
      }
      continue;
    }
    if (character === '"') {
      quoted = true;
    } else if (character === "{") {
      depth += 1;
    } else if (character === "}" && --depth === 0) {
      valueEnd = index;
      break;
    }
  }
  if (propertyStart < 0 || valueStart < 0 || valueEnd < 0) {
    throw new Error(`${connector}: could not remove destination block without reformatting source transport`);
  }
  await writeFile(file, `${raw.slice(0, propertyStart)}${raw.slice(valueEnd + 1)}`);
}
