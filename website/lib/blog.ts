export type BlogSection = {
  heading: string;
  body: string[];
  points?: string[];
  code?: string;
};

export type BlogPost = {
  slug: string;
  title: string;
  description: string;
  publishedAt: string;
  updatedAt: string;
  readingTime: string;
  category: string;
  tags: string[];
  summary: string;
  sections: BlogSection[];
};

export const BLOG_POSTS: BlogPost[] = [
  {
    slug: 'one-cli-to-rule-them-all',
    title: 'One CLI To Rule Them All',
    description:
      'Why Polymetrics puts ETL, DuckDB SQL, reverse ETL, scheduling, and agent-safe JSON contracts behind one local binary.',
    publishedAt: '2026-07-02',
    updatedAt: '2026-07-02',
    readingTime: '5 min read',
    category: 'Product essay',
    tags: ['local-first ETL', 'DuckDB', 'reverse ETL', 'AI agents'],
    summary:
      'A single binary is easier to install, easier to audit, easier to automate, and easier for AI agents to operate without hidden infrastructure.',
    sections: [
      {
        heading: 'The data loop should not need a platform team',
        body: [
          'Most operational data work follows the same shape: extract from a source, land it somewhere queryable, decide what should happen, then write the result back to the systems where work happens.',
          'Polymetrics makes that loop a command-line workflow. The same binary owns connector setup, local storage, DuckDB queries, reverse ETL planning, approval, execution, and structured JSON output.',
        ],
      },
      {
        heading: 'Why a CLI is the right control plane',
        body: [
          'A CLI is portable, scriptable, observable, and easy to put under source control. It works on a laptop, in CI, in cron, in a container, and inside an AI agent loop without changing the interface.',
          'That matters because data workflows fail at the edges: credentials, schema drift, rate limits, approvals, retries, and audit trails. Keeping those edges in one binary keeps the mental model small.',
        ],
        points: [
          'No server to deploy before the first sync.',
          'No separate SQL service for local analysis.',
          'No second product just to write data back.',
          'No special agent API when the CLI already emits JSON.',
        ],
      },
      {
        heading: 'The loop in one place',
        body: [
          'Polymetrics treats extract, query, and act as one product surface. That lets a developer test the full workflow locally before promoting it to automation.',
        ],
        code: `pm etl run --connection github --stream issues --json
pm query run --sql "select * from issues where state = 'open'" --json
pm reverse plan sync --source-table stale_issues --destination github:write --json`,
      },
      {
        heading: 'What this unlocks',
        body: [
          'The product goal is simple: make serious data automation feel as lightweight as a Unix tool. The long-term moat is a broad connector catalog, a predictable local runtime, and safe write paths that both humans and agents can understand.',
        ],
      },
    ],
  },
  {
    slug: 'agent-native-data-workflows',
    title: 'Agent-Native Data Workflows Need Boring Contracts',
    description:
      'How JSON envelopes, stable exit codes, local credentials, and approval-gated writes make CLI automation safer for LLM agents.',
    publishedAt: '2026-07-02',
    updatedAt: '2026-07-02',
    readingTime: '6 min read',
    category: 'Engineering note',
    tags: ['AI agents', 'JSON', 'approval gates', 'CLI design'],
    summary:
      'Agentic data tools should be predictable before they are powerful. The contract must make success, failure, and mutation explicit.',
    sections: [
      {
        heading: 'Agents need less magic, not more',
        body: [
          'LLM agents are good at planning across tools, but they are fragile when the tools speak in prose, hide state, or mutate systems without a preview. A data CLI should expose a narrow, explicit contract.',
          'Polymetrics keeps the agent interface boring: JSON on stdout, logs on stderr, stable exit codes, and no destination writes until a plan is approved.',
        ],
      },
      {
        heading: 'The minimum safe contract',
        body: [
          'The agent should never need to parse progress bars or infer whether an operation succeeded from a sentence. Every command should make its status, resource identifiers, warnings, and retryability machine-readable.',
        ],
        points: [
          'Structured output with versioned envelopes.',
          'Exit codes that separate usage, validation, auth, connector, runtime, policy, and internal errors.',
          'Credentials referenced by name, never printed.',
          'Writes split into plan, preview, approve, and execute.',
        ],
      },
      {
        heading: 'Approval is a product feature',
        body: [
          'Reverse ETL is where automation becomes risky. The product should show the exact intended change before it touches the destination. That is useful for humans, and it is essential for agents.',
        ],
        code: `pm reverse plan sync --source-table candidates --destination github:write --json
pm reverse preview <plan-id> --json
pm reverse run <plan-id> --approve <token> --json`,
      },
      {
        heading: 'A better default for automation',
        body: [
          'The CLI becomes the shared contract for humans, CI, cron, and agent pods. There is no separate agent-only surface to secure, document, or debug.',
        ],
      },
    ],
  },
  {
    slug: 'local-first-data-engine',
    title: 'Local-First Data Pipelines Without Warehouse Sprawl',
    description:
      'A practical case for keeping extraction, local storage, analytical SQL, and write-back automation close to the developer.',
    publishedAt: '2026-07-02',
    updatedAt: '2026-07-02',
    readingTime: '5 min read',
    category: 'Architecture',
    tags: ['local-first', 'ETL', 'warehouse', 'open source'],
    summary:
      'Local-first does not mean toy. It means the first useful run happens on your machine, with production paths available when the workflow earns them.',
    sections: [
      {
        heading: 'Start local, promote deliberately',
        body: [
          'A surprising amount of data automation starts as a question: pull a few records, inspect a schema, join it with another source, then decide whether the result should become a scheduled workflow.',
          'If the first step requires a hosted warehouse, a worker deployment, and multiple service accounts, the experiment becomes heavier than the question.',
        ],
      },
      {
        heading: 'DuckDB changes the default',
        body: [
          'Embedded analytical SQL lets a CLI run real joins, aggregations, and window functions without asking the user to provision a database first. That makes local data work fast enough for exploration and structured enough for automation.',
        ],
        points: [
          'Extract connector data into a local warehouse.',
          'Run SQL near the files and credentials.',
          'Promote only the useful workflow to schedule or CI.',
          'Keep sensitive data on the developer machine unless a workflow intentionally writes it elsewhere.',
        ],
      },
      {
        heading: 'Local does not mean isolated',
        body: [
          'The same commands can run in cron, GitHub Actions, a container, or a managed runner later. Local-first is a product sequence, not a deployment ceiling.',
        ],
      },
      {
        heading: 'The open-source opportunity',
        body: [
          'Developers already understand the shape of command-line tools. A repo that explains its value clearly, documents the first run, exposes crawlable examples, and publishes useful technical essays gives search engines and AI answer engines real evidence to cite.',
        ],
      },
    ],
  },
];

export function getBlogPost(slug: string): BlogPost | undefined {
  return BLOG_POSTS.find((post) => post.slug === slug);
}

export function blogUrl(slug: string): string {
  return `/blog/${slug}`;
}
