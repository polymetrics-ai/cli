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
    slug: 'human-harnesses',
    title: 'Humans Need Harnesses Too',
    description:
      'We built approval gates so agents could not quietly damage your data. The repository taught us that people, pull requests, releases, and deployments need the same hard stops.',
    publishedAt: '2026-08-04',
    updatedAt: '2026-08-04',
    readingTime: '12 min read',
    category: 'Build in public',
    tags: ['human harnesses', 'GitHub Actions', 'approval gates', 'AI agents'],
    summary:
      'A harness is not an agent feature. It is a sequence that makes intent, evidence, approval, and mutation explicit for every operator, including the human at the keyboard.',
    sections: [
      {
        heading: 'The tool changed the shape of the work',
        body: [
          'I will start with a confession. I stopped using the large AI coding platforms as my daily driver. They can move quickly, but I was spending too much time steering the platform, arranging agents, and recovering context instead of staying with the problem.',
          'Then I found Pi. It was small, direct, and easy to shape around the repository. Combined with GSD, it gave me a programming loop I could inspect: define the issue, plan the slice, write the failing test, make it pass, verify the result, and leave evidence for the next person or agent.',
          'That workflow helped make Polymetrics possible. More importantly, it exposed the lesson behind this article: speed without a visible control loop is just a faster way to arrive at an unreviewed decision.',
        ],
      },
      {
        heading: 'The itch',
        body: [
          'I wanted a real CLI for every connector, the way gh is my daily driver for GitHub. Not a configuration file interpreted by a cloud service I cannot see, but a command surface that can inspect a connector, extract data, query it locally, plan a write, and explain every failure.',
          'The connector rewrite now stores each integration as a JSON bundle and interprets it through a shared engine. API operations are classified as streams, direct reads, writes, binary transfers, native protocol work, hooks, or typed exclusions. That classification matters because a list operation, a file download, and a DELETE should not inherit the same runtime policy.',
          'I also built a Shepherd-style supervisor around the agent workflow. One correction to my earlier shorthand is important: Shepherd is not the thing that writes all 547 connectors. Connector workers perform the mapping; Shepherd independently checks whether the orchestration is progressing, whether workers collide, whether verification is real, and whether the run should continue, revert, or stop for a human.',
        ],
      },
      {
        heading: 'What the repository actually contains',
        body: [
          'These are repository inventory numbers, not a claim that every connector is production-certified. At the time of writing there are 547 connector definition directories, each with an API-surface inventory. Those files contain 29,129 classified endpoint entries: 14,780 GET reads, 3 HEAD checks, 14,169 explicit POST, PUT, PATCH, or DELETE mutations, and 177 hook, wildcard, GraphQL, WebSocket, or composite-method rows. The explicit mutations include 2,903 DELETE operations.',
          'The same bundles define 7,088 ETL streams. Some operations are implemented, some are deliberately excluded, some depend on hooks or native code, and some still need live certification. The inventory tells us the size and shape of the work. It does not let us skip the proof that a connector behaves correctly against a real service.',
        ],
        points: [
          '547 connector bundles with explicit API-surface inventories.',
          '29,129 operations classified before they are exposed as product behavior.',
          '14,780 GET reads, 3 HEAD checks, and 14,169 explicit HTTP mutations.',
          '177 mixed or nonstandard method rows remain visible instead of being mislabeled as writes.',
          '2,903 DELETE operations are called out explicitly inside the mutation inventory.',
          '7,088 ETL streams defined for conformance and certification work.',
        ],
      },
      {
        heading: 'The uncomfortable realization',
        body: [
          'From the start, Polymetrics had an agent contract: machine-readable JSON, stable exit behavior, credentials referenced instead of printed, and reverse ETL split into plan, preview, approval, and execution. The point was to let an LLM use the same CLI as a person without giving it a quiet path to mutate a destination.',
          'Then one night I watched myself work. It was late, I was tired, and I was about to aim a write at the wrong environment. The thing that saved me was not better judgment. It was the same stop I had designed for an agent.',
          'That is the uncomfortable part: I am also an unreliable agent. Humans paste a correct command into the wrong terminal. We skim diffs, approve familiar-looking output, reuse stale plans, and reach for force when the system slows us down. A safety model that protects data only when an AI is operating it has misunderstood the operator.',
        ],
      },
      {
        heading: 'The runtime harness',
        body: [
          'A harness turns a mutation from one action into a sequence with named states. First describe the intended change. Then calculate and preview the concrete effect. Bind approval to that plan. Only then execute it. A failure returns structured status and leaves the destination unchanged whenever the operation boundary allows it.',
          'The August 4 target is to make that model explicit for people as well as agents. Destructive and sensitive paths should be unavailable until policy enables them. Approval should be scoped to a specific plan rather than becoming an ambient yes to everything. The same command should behave predictably from a terminal, CI job, cron entry, or agent loop.',
          'This is also where honesty matters. A mapped operation is not automatically an executable command, and an executable command is not automatically certified. The inventory, runtime policy, conformance tests, and live certification are separate gates because each answers a different question.',
        ],
        code: `pm reverse plan candidates_to_github --source-table candidates --destination github:github-local --action create_issue --map title:title
pm reverse preview <plan-id> --json
pm reverse run <plan-id> --approve <approval-token> --json`,
      },
      {
        heading: 'The repository became a harness',
        body: [
          'While building the runtime, we ended up applying the same structure to the repository. A pull request is also a proposed mutation. It changes the code people install, the generated connector catalog, the release artifacts, and sometimes the process running in production. Treating it as a loose diff would contradict the product we are building.',
          'The repository therefore asks for intent before code, evidence before confidence, review before integration, and an immutable artifact before deployment. GitHub Actions is the execution layer, but the harness begins earlier in the issue contract and ends later in rollout verification.',
        ],
        code: `data:    plan -> preview -> approve -> execute -> audit
code:    issue -> plan -> red -> green -> review -> merge
release: change -> release PR -> tag -> versioned build artifacts
deploy:  tested commit -> immutable image -> rollout -> health check`,
      },
      {
        heading: 'Intent before diff',
        body: [
          'Every non-trivial change starts with an issue that states the objective, scope, exclusions, acceptance criteria, verification, safety notes, and review route. The PR Issue Guard checks that the title and PR body use an accepted issue-reference shape. It does not prove that the issue exists or that the diff matches its scope; reviewers still own that judgment. The conventions workflow checks the branch name and Conventional Commit PR title, so the change identifies itself before anyone reads the implementation.',
          'When production Go code under cmd or internal changes, the GSD workflow checks for planning and test evidence. In practice that means a plan, a TDD ledger, and a verification checklist exist before production edits. The ledger does not prove the code is correct, but it makes a useful distinction visible: did a test fail because the capability was absent, and did the same test pass because of the change?',
          'This is the code equivalent of binding an approval to a plan. The issue describes the intended mutation; the branch and PR limit its scope; the test demonstrates the behavioral delta. A reviewer can challenge any of those layers instead of reverse-engineering intent from the final diff.',
        ],
      },
      {
        heading: 'Proof before confidence',
        body: [
          'The general verification workflow installs the pinned Go toolchain and linter, runs make verify, and then fails if verification changed generated files. That last check is easy to overlook. A generator that produces an uncommitted diff is evidence that source and published artifacts disagree, so the harness treats drift as a failure rather than a cleanup task.',
          'Security runs in parallel: govulncheck inspects reachable Go vulnerabilities, CodeQL analyzes Go and JavaScript or TypeScript, and dependency review rejects newly introduced high-severity dependency risk. A weekly OpenSSF Scorecard run adds a slower repository-level view of supply-chain posture.',
          'The website has its own integration harness. It starts PostgreSQL 17, regenerates website data and checks for drift, typechecks, runs unit tests, installs Chromium, exercises the site with Playwright, creates a production build, and then builds the container image. That sequence catches failures that a component test cannot: migrations, generated catalog mismatches, route rendering, browser behavior, and production compilation.',
        ],
        points: [
          'Deterministic checks: formatting, linting, tests, builds, and generated-file drift.',
          'Security checks: govulncheck, CodeQL, dependency review, and scheduled Scorecard analysis.',
          'Website checks: PostgreSQL migrations, generated data, typecheck, unit tests, Chromium e2e, build, and image construction.',
        ],
      },
      {
        heading: 'Reviewing untrusted code without trusting it',
        body: [
          'The Claude review workflow contains one of the clearest examples of a harness changing behavior. A public pull request is untrusted code, and the review credential is valuable. Running that pull request while the credential is available would give hostile code a chance to exfiltrate it.',
          'So the review is static. The workflow checks the author relationship before the secret reaches the job, checks permissions again inside the action, blocks bots by default, checks out the base repository, and asks the reviewer to inspect the PR diff without running the proposed code. Compilation and tests belong to separate jobs that do not receive the Claude credential.',
          'The workflow also avoids reviewing every push. It runs automatically when a trusted author opens, reopens, or marks a PR ready, and maintainers can request a deliberate follow-up review after fixes. That is a quota decision, but it is also a state-machine decision: review a meaningful checkpoint, record dispositions, then request another pass only when the reviewed commit range changed.',
        ],
      },
      {
        heading: 'Release and deployment are mutations too',
        body: [
          'Merging code is not the final write. The release workflow lets release-please assemble the version and changelog on main, and GoReleaser builds artifacts only when a release is actually created. That separates ordinary integration from publication.',
          'The website follows a similar boundary. Pull requests can build the image, but only a main-branch push or an explicit dispatch can publish it. Deployment requires the website deploy variable, uses a self-hosted Tailscale runner, passes the image tagged with the exact commit SHA to the Quadlet deployment script, and verifies rollout health. The deploy consumes an immutable input instead of rebuilding whatever happens to be on the server.',
          'This is the production form of plan and execute: CI proves one commit, the registry stores an image for that commit, and the deploy step rolls out that exact image. If those identities diverge, the harness is no longer describing the mutation it performs.',
        ],
      },
      {
        heading: 'What GitHub really blocks',
        body: [
          'There is a difference between a workflow that runs and a check that GitHub requires before main can move. At the time of writing, main uses strict required status checks for verify, govulncheck, CodeQL, branch-name, and pr-title, and the rule applies to administrators. It also requires linear history and resolved review conversations. A branch that is behind main must be brought current before those checks can authorize the merge.',
          'Other useful workflows also run, including the issue guard, GSD evidence check, dependency review, and website suite when relevant files change. They are part of our delivery contract, but they are not all listed as required branch-protection contexts today. Saying otherwise would turn a process expectation into a false technical guarantee.',
          'The parent PR is still human-gated. Sub-PRs can be integrated into a parent branch after scoped checks and review coverage, but the parent branch does not merge to main automatically. That gate matters because a green collection of local changes can still be incoherent as one product release.',
        ],
      },
      {
        heading: 'What the harness still does not do',
        body: [
          'The repository is active and the harness is not finished. Branch protection currently requires status checks but not a positive review count. The human gate on the parent PR is repository policy, not a GitHub rule that mathematically prevents every maintainer from merging without review.',
          'Automated review is also an input, not an oracle. Claude can be disabled or run out of quota, and a skipped or failed review is not coverage. The fallback is a recorded review of the relevant commit range by another reviewer, not pretending that a workflow name in the checks list means somebody examined the change.',
          'The production environment does not yet require an environment reviewer, so a qualifying main push can deploy when the deployment variable is enabled. Actions are version-tagged rather than pinned to full commit SHAs, and repository settings do not currently enforce SHA pinning. Those are real hardening opportunities, and naming them is part of the harness: an unrecorded gap is impossible to prioritize.',
        ],
        points: [
          'Add an enforced review requirement or preserve an auditable human-gate record for parent merges.',
          'Add a production environment approval when deployment independence becomes more important than immediate rollout.',
          'Pin third-party actions by commit and enable repository-level pinning policy.',
          'Treat unavailable automated review as a visible blocker or explicit human fallback, never as a silent pass.',
        ],
      },
      {
        heading: 'Same contract, both species',
        body: [
          'The useful idea is not that humans and agents are identical. It is that neither should receive a special path around the state machine. A shell script should not get an easier write path than an agent. A maintainer should not get a less reproducible deployment because they know the server. Familiarity is not evidence.',
          'Stable JSON and exit behavior make runtime state legible. Issues, TDD ledgers, checks, and reviewed commit ranges make repository state legible. Immutable image tags make deployment state legible. In each case the harness replaces an assumption with an object we can inspect.',
          'This page is part of that feedback loop. A reader can highlight a passage, attach a marginal note to the exact text, continue the reply thread, or open the page-level GitHub discussion. The annotation does not make the argument correct; it makes disagreement specific, durable, and reviewable.',
          'That is why human harnesses are becoming a first-class part of Polymetrics. The safest operator is not the one who promises never to make a mistake. It is the one whose tools expect mistakes and stop them before intention becomes mutation.',
        ],
      },
      {
        heading: 'Where this goes next',
        body: [
          'The immediate work is certification. Connector inventory is broad; live, repeatable evidence is the next gate. We are also tightening approval scopes, improving bulk-write previews, extending dry-run diffs, and making destructive confirmation policy consistent across connector actions.',
          'Distribution is part of the same story: release binaries and a Homebrew path should make the first useful run short without bypassing provenance. The MCP surface should let agents discover the same typed operations without inventing a second, more permissive API.',
          'The next engineering note will separate two stories that are easy to conflate: how connector operations are classified into product surfaces, and how the Shepherd supervisor judges the agents doing that work. One creates the map; the other checks the journey.',
        ],
        points: [
          'Repository: github.com/polymetrics-ai/cli',
          'Documentation: cli.polymetrics.ai',
          'Inventory snapshot: 547 API-surface files and 29,129 endpoint entries on July 16, 2026.',
        ],
      },
    ],
  },
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
