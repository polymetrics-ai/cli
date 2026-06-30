export const meta = {
  name: 'enrich-connectors',
  description: 'Research + independently adversarially verify connector Setup/Auth enrichment (idempotent, batched)',
  phases: [{ title: 'Scout' }, { title: 'Research' }, { title: 'Verify' }, { title: 'Rescout' }],
};

const INPUT = '/Users/karthiksivadas/Development/polymetrics-agentic-team/polymetrics-v2/website/.enrich/connectors-input.json';
const SLUGS = '/Users/karthiksivadas/Development/polymetrics-agentic-team/polymetrics-v2/website/.enrich/slugs.json';
const OUTDIR = '/Users/karthiksivadas/Development/polymetrics-agentic-team/polymetrics-v2/website/.enrich/enr';

// Per-invocation budget. Full-catalog multi-stage enrichment can exceed the
// agent lifetime cap, so each invocation handles at most BATCH not-yet-done
// slugs across <=MAX_PASSES.
// Worst case agents = BATCH * 2 * MAX_PASSES + scouts; keep well under 1000.
const BATCH = 240;
const CHUNK = 8;
const MAX_PASSES = 2;

const SCOUT_SCHEMA = {
  type: 'object', additionalProperties: false,
  properties: {
    all: { type: 'array', items: { type: 'string' } },
    done: { type: 'array', items: { type: 'string' } },
  },
  required: ['all', 'done'],
};

const ENRICH_SCHEMA = {
  type: 'object', additionalProperties: false,
  properties: {
    prerequisites: { type: 'array', items: { type: 'string' } },
    authMethods: { type: 'array', items: { type: 'object', additionalProperties: false, properties: { name: { type: 'string' }, summary: { type: 'string' } }, required: ['name', 'summary'] } },
    setupSteps: { type: 'array', items: { type: 'object', additionalProperties: false, properties: { title: { type: 'string' }, body: { type: 'string' } }, required: ['title', 'body'] } },
    sources: { type: 'array', items: { type: 'object', additionalProperties: false, properties: { title: { type: 'string' }, url: { type: 'string' } }, required: ['title', 'url'] } },
  },
  required: ['prerequisites', 'authMethods', 'setupSteps', 'sources'],
};

const VERDICT_SCHEMA = {
  type: 'object', additionalProperties: false,
  properties: {
    slug: { type: 'string' }, ok: { type: 'boolean' }, wrote: { type: 'boolean' },
    prerequisites: { type: 'number' }, authMethods: { type: 'number' },
    setupSteps: { type: 'number' }, sources: { type: 'number' }, notes: { type: 'string' },
  },
  required: ['slug', 'ok', 'wrote'],
};

const scoutPrompt = `Use Bash only; do not research anything.
1. Read the full slug list: cat "${SLUGS}"  (it is a JSON array of connector slugs).
2. List already-written enrichment files: ls "${OUTDIR}" 2>/dev/null
Return { all: <the JSON array from slugs.json, verbatim>, done: <array of slugs that already have a file, i.e. each existing "<slug>.json" filename in ${OUTDIR} with the .json suffix removed> }.`;

function researchPrompt(slug) {
  return `You are writing accurate "Setup & Authentication" enrichment for the data connector with slug "${slug}".

STEP 1 - read its descriptor (name, category, language, secretFields, configFields, appDocUrl, officialDocs):
  python3 -c "import json;print(json.dumps(json.load(open('${INPUT}'))['${slug}']))"

STEP 2 - research authoritative setup/auth using (a) the descriptor's officialDocs + appDocUrl (vendor docs), and (b) the local SearXNG JSON API:
  curl -s "http://localhost:8888/search?q=<URL-ENCODED-QUERY>&format=json&engines=google"
  Good queries: "<name> API authentication", "<name> create API key OR personal access token", "<name> getting started prerequisites". Prefer official vendor domains.

STRICT ACCURACY RULES (this output will be independently adversarially verified - unverifiable claims will be deleted):
- Only include facts confirmable from a resolvable authoritative source you cite in "sources".
- Map authMethods to the connector's REAL secretFields from the descriptor (e.g. a secret field "api_key" => an API-key method; "credentials.client_id/client_secret/access_token" => OAuth). Do NOT invent auth methods.
- Derive prerequisites and config steps from the REAL configFields. Never invent config field names, scopes, or permissions.
- Keep prose concise, factual, imperative. NO em dashes or en dashes.
- This is the "pm" tool, NOT Airbyte. Never mention "Airbyte", never cite airbyte.com as a source, and never describe Airbyte-specific deployment (Docker images, Kubernetes, "Airbyte Cloud"). Describe prerequisites in terms of the data system itself (the database/API), not the ETL runtime.
- If the connector has NO vendor docs (officialDocs empty + appDocUrl empty) and SearXNG returns nothing authoritative: produce a MINIMAL entry derived ONLY from configFields/secretFields - prerequisites = the required config fields, authMethods inferred from the secret field names, setupSteps = one generic "Provide credentials and configure the required fields" step, sources = []. No fabrication.

Return the ConnectorEnrichment object (prerequisites[], authMethods[{name,summary}], setupSteps[{title,body}], sources[{title,url}]).`;
}

function verifyPrompt(slug, draft) {
  return `Independently and adversarially VERIFY the drafted Setup/Auth enrichment for connector "${slug}". Assume it may contain errors; prune everything not provably correct.

Descriptor (ground truth): python3 -c "import json;print(json.dumps(json.load(open('${INPUT}'))['${slug}']))"

DRAFT to verify:
${JSON.stringify(draft)}

VERIFY + PRUNE:
1. Sources: test each URL - curl -s -o /dev/null -w "%{http_code}" -L --max-time 12 "<url>". DROP any source whose status is 000, 404, or any 4xx (a 403 from a real vendor domain may stay, since some block bots).
2. authMethods: DROP any method that does not correspond to a real entry in the descriptor's secretFields (or an obviously supported no-secret method).
3. setupSteps / prerequisites: DROP or fix any step that references a config field NOT present in the descriptor's configFields, or any invented scope/permission/claim not backed by a surviving source.
4. Remove marketing fluff and any em dashes or en dashes.
5. DELETE every mention of "Airbyte" and drop any airbyte.com source (this is the pm tool, not Airbyte); reword any "reachable from Airbyte"/"Airbyte Cloud"/"Airbyte deployment" prose to describe the data system or pm generically.

Then build the PRUNED, verified ConnectorEnrichment and WRITE it as JSON to "${OUTDIR}/${slug}.json" (use the Write tool). The JSON must match exactly: {"prerequisites":[...],"authMethods":[{"name","summary"}],"setupSteps":[{"title","body"}],"sources":[{"title","url"}]}.
- If after pruning there are NO authMethods AND NO setupSteps, do NOT write the file (nothing useful survived).
Return the verdict {slug, ok, wrote, prerequisites, authMethods, setupSteps, sources, notes} where the counts are the FINAL written counts and "wrote" is whether you wrote the file.`;
}

// ---- orchestration --------------------------------------------------------

const scout0 = await agent(scoutPrompt, { label: 'scout', phase: 'Scout', schema: SCOUT_SCHEMA });
const all = Array.isArray(scout0?.all) ? scout0.all : [];
let done = new Set(Array.isArray(scout0?.done) ? scout0.done : []);

// args overrides: array => exactly those slugs; number/numeric string => BATCH cap.
let explicit = null;
let batch = BATCH;
if (Array.isArray(args)) explicit = args;
else if (typeof args === 'number' && args > 0) batch = args;
else if (typeof args === 'string' && /^\d+$/.test(args.trim())) batch = parseInt(args.trim(), 10);

const pool = (explicit || all).filter((s) => !done.has(s));
let remaining = pool.slice(0, batch);
log(`Scout: ${all.length} total, ${done.size} already done. This invocation will attempt ${remaining.length} (batch=${batch}).`);

const verdicts = [];
for (let pass = 1; pass <= MAX_PASSES && remaining.length; pass++) {
  log(`Pass ${pass}/${MAX_PASSES}: ${remaining.length} connectors, chunk=${CHUNK}.`);
  for (let i = 0; i < remaining.length; i += CHUNK) {
    const chunk = remaining.slice(i, i + CHUNK);
    const r = await pipeline(
      chunk,
      (slug) => agent(researchPrompt(slug), { label: `research:${slug}`, phase: 'Research', schema: ENRICH_SCHEMA }),
      (draft, slug) => {
        if (!draft) return { slug, ok: false, wrote: false, notes: 'research failed (rate limit?)' };
        return agent(verifyPrompt(slug, draft), { label: `verify:${slug}`, phase: 'Verify', schema: VERDICT_SCHEMA });
      },
    );
    for (const v of r) if (v) verdicts.push(v);
    log(`  pass ${pass}: ${Math.min(i + CHUNK, remaining.length)}/${remaining.length} processed`);
  }
  // Re-scout from disk (authoritative) to find what still needs work.
  const rs = await agent(scoutPrompt, { label: `rescout:p${pass}`, phase: 'Rescout', schema: SCOUT_SCHEMA });
  done = new Set(Array.isArray(rs?.done) ? rs.done : [...done]);
  remaining = (explicit || all).filter((s) => !done.has(s)).slice(0, batch).filter((s) => remaining.includes(s));
}

const stillMissing = (explicit || all).filter((s) => !done.has(s));
log(`Done. Files on disk: ${done.size}/${all.length}. Still missing overall: ${stillMissing.length}.`);
return { total: all.length, doneOnDisk: done.size, stillMissing: stillMissing.length, attempted: verdicts.length, sampleMissing: stillMissing.slice(0, 20) };
